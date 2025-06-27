package gcp

import (
	"context"
	"errors"
	"fmt"

	securitycenter "cloud.google.com/go/securitycenter/apiv2"
	securitycenterpb "cloud.google.com/go/securitycenter/apiv2/securitycenterpb"
	"golang.org/x/oauth2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/pkg/compmgr"
)

// Provider implements the compmgr.Provider interface using
// Google Cloud Security Command Center.
type Provider struct {
	projects []string
	client   *securitycenter.Client
	filter   FindingFilter
}

// FindingFilter allows excluding Security Command Center findings from
// processing. Returning false skips the finding.
type FindingFilter func(*securitycenterpb.Finding) bool

// ErrProjectIDRequired is returned when a project ID is not provided.
var ErrProjectIDRequired = errors.New("project ID must be provided")

// ErrClientNotInitialized is returned when the provider has no GCP client configured.
var ErrClientNotInitialized = errors.New("gcp client not initialized")

// Options configure the GCP provider.
type Options struct {
	Projects   []string
	client     *securitycenter.Client
	clientOpts []option.ClientOption
	Filter     FindingFilter
}

// Option configures the GCP provider options.
type Option func(*Options)

// WithProjectID sets the GCP project ID.
func WithProjectID(id string) Option {
	return func(o *Options) { o.Projects = append(o.Projects, id) }
}

// WithProjects sets multiple GCP project IDs to fetch findings from.
func WithProjects(ids ...string) Option {
	return func(o *Options) { o.Projects = append(o.Projects, ids...) }
}

// WithCredentialsFile provides a service account credentials file.
func WithCredentialsFile(path string) Option {
	return func(o *Options) { o.clientOpts = append(o.clientOpts, option.WithCredentialsFile(path)) }
}

// WithCredentialsJSON provides service account credentials from JSON bytes.
func WithCredentialsJSON(data []byte) Option {
	return func(o *Options) { o.clientOpts = append(o.clientOpts, option.WithCredentialsJSON(data)) }
}

// WithTokenSource uses the supplied OAuth2 token source for authentication.
func WithTokenSource(ts oauth2.TokenSource) Option {
	return func(o *Options) { o.clientOpts = append(o.clientOpts, option.WithTokenSource(ts)) }
}

// WithClient allows supplying an existing Security Command Center client.
func WithClient(c *securitycenter.Client) Option {
	return func(o *Options) { o.client = c }
}

// WithFindingFilter sets a filter function to exclude findings from the
// provider. If the function returns false, the finding is ignored.
func WithFindingFilter(f FindingFilter) Option {
	return func(o *Options) { o.Filter = f }
}

// New creates a new GCP compliance provider using the Security Command Center.
func New(ctx context.Context, opts ...Option) (*Provider, error) {
	conf := Options{}
	for _, opt := range opts {
		opt(&conf)
	}

	if len(conf.Projects) == 0 {
		return nil, ErrProjectIDRequired
	}

	if conf.client == nil {
		var err error
		conf.client, err = securitycenter.NewClient(ctx, conf.clientOpts...)
		if err != nil {
			return nil, err
		}
	}

	return &Provider{projects: conf.Projects, client: conf.client, filter: conf.Filter}, nil
}

// ListReports retrieves findings from Security Command Center.
func (p *Provider) ListReports(ctx context.Context) ([]compmgr.Report, error) {
	if p.client == nil {
		return nil, ErrClientNotInitialized
	}

	var reports []compmgr.Report

	for _, project := range p.projects {
		req := &securitycenterpb.ListFindingsRequest{
			Parent: fmt.Sprintf("projects/%s/sources/-", project),
		}

		it := p.client.ListFindings(ctx, req)

		for {
			resp, err := it.Next()
			if errors.Is(err, iterator.Done) {
				break
			}
			if err != nil {
				return nil, err
			}
			f := resp.GetFinding()
			if p.filter != nil && !p.filter(f) {
				continue
			}

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

			passed := f.GetState() == securitycenterpb.Finding_INACTIVE

			reports = append(reports, compmgr.Report{
				ID:          f.GetName(),
				Name:        f.GetCategory(),
				Description: f.GetDescription(),
				Link:        f.GetExternalUri(),
				Passed:      passed,
				Controls:    controls,
				ControlIDs:  ids,
			})
		}
	}

	return reports, nil
}

var _ compmgr.Provider = (*Provider)(nil)
