package gcp

import (
	"context"
	"errors"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	securitycenterpb "cloud.google.com/go/securitycenter/apiv2/securitycenterpb"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/theopenlane/core/pkg/compmgr"
)

// PubSubProvider implements the compmgr.Provider interface by reading
// Security Command Center notifications from a Pub/Sub subscription. The
// provider is read-only and does not push updates back to Security Command
// Center or any other GCP service.
type PubSubProvider struct {
	sub    subscription
	filter FindingFilter
}

// subscription abstracts the Receive method of pubsub.Subscription.
type subscription interface {
	Receive(context.Context, func(context.Context, *pubsub.Message)) error
}

// PubSubOptions configure the PubSubProvider.
type PubSubOptions struct {
	ProjectID      string
	SubscriptionID string
	client         *pubsub.Client
	subscription   subscription
	Filter         FindingFilter
}

// PubSubOption configures PubSubOptions.
type PubSubOption func(*PubSubOptions)

// WithPubSubClient sets an existing Pub/Sub client.
func WithPubSubClient(c *pubsub.Client) PubSubOption {
	return func(o *PubSubOptions) { o.client = c }
}

// WithProjectID sets the GCP project ID hosting the subscription.
func WithPubSubProjectID(id string) PubSubOption {
	return func(o *PubSubOptions) { o.ProjectID = id }
}

// WithSubscriptionID sets the subscription ID to pull messages from.
func WithSubscriptionID(id string) PubSubOption {
	return func(o *PubSubOptions) { o.SubscriptionID = id }
}

// WithSubscription sets an existing Pub/Sub subscription.
func WithSubscription(s subscription) PubSubOption {
	return func(o *PubSubOptions) { o.subscription = s }
}

// WithFindingFilter sets a filter to exclude findings from notifications.
func WithPubSubFindingFilter(f FindingFilter) PubSubOption {
	return func(o *PubSubOptions) { o.Filter = f }
}

// ErrSubscriptionRequired is returned when no subscription is configured.
var ErrSubscriptionRequired = errors.New("pubsub subscription must be provided")

// NewPubSub creates a PubSubProvider from the given options.
func NewPubSub(ctx context.Context, opts ...PubSubOption) (*PubSubProvider, error) {
	conf := PubSubOptions{}
	for _, opt := range opts {
		opt(&conf)
	}
	if conf.subscription == nil {
		if conf.client == nil {
			if conf.ProjectID == "" {
				return nil, ErrSubscriptionRequired
			}
			c, err := pubsub.NewClient(ctx, conf.ProjectID)
			if err != nil {
				return nil, err
			}
			conf.client = c
		}
		if conf.SubscriptionID == "" {
			return nil, ErrSubscriptionRequired
		}
		conf.subscription = conf.client.Subscription(conf.SubscriptionID)
	}
	return &PubSubProvider{sub: conf.subscription, filter: conf.Filter}, nil
}

// ListReports reads available Pub/Sub messages and converts them to Reports.
func (p *PubSubProvider) ListReports(ctx context.Context) ([]compmgr.Report, error) {
	if p.sub == nil {
		return nil, ErrSubscriptionRequired
	}

	const timeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var mu sync.Mutex
	var reports []compmgr.Report
	err := p.sub.Receive(ctx, func(_ context.Context, m *pubsub.Message) {
		defer m.Ack()
		var note securitycenterpb.NotificationMessage
		if err := protojson.Unmarshal(m.Data, &note); err != nil {
			// skip unparsable messages
			return
		}
		f := note.GetFinding()
		if f == nil {
			return
		}
		if p.filter != nil && !p.filter(f) {
			return
		}

		passed := f.GetState() == securitycenterpb.Finding_INACTIVE
		var controls []compmgr.Control
		var ids []string
		for _, c := range f.GetCompliances() {
			controls = append(controls, compmgr.Control{
				Standard: c.GetStandard(),
				Version:  c.GetVersion(),
				IDs:      c.GetIds(),
			})
			ids = append(ids, c.GetIds()...)
		}
		r := compmgr.Report{
			ID:          f.GetName(),
			Name:        f.GetCategory(),
			Description: f.GetDescription(),
			Link:        f.GetExternalUri(),
			Passed:      passed,
			Controls:    controls,
			ControlIDs:  ids,
		}
		mu.Lock()
		reports = append(reports, r)
		mu.Unlock()
	})
	if err != nil && !errors.Is(err, context.Canceled) {
		return nil, err
	}
	return reports, nil
}

var _ compmgr.Provider = (*PubSubProvider)(nil)
