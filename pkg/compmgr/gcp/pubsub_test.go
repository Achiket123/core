package gcp

import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/pubsub"
	securitycenterpb "cloud.google.com/go/securitycenter/apiv2/securitycenterpb"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestNewPubSub(t *testing.T) {
	proj := os.Getenv("GCP_PROJECT")
	sub := os.Getenv("PUBSUB_SUBSCRIPTION")
	if proj == "" || sub == "" {
		t.Skip("GCP_PROJECT or PUBSUB_SUBSCRIPTION not set")
	}
	if _, err := NewPubSub(context.Background(), WithPubSubProjectID(proj), WithSubscriptionID(sub)); err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}
}

func TestPubSubListReports(t *testing.T) {
	proj := os.Getenv("GCP_PROJECT")
	sub := os.Getenv("PUBSUB_SUBSCRIPTION")
	if proj == "" || sub == "" {
		t.Skip("GCP_PROJECT or PUBSUB_SUBSCRIPTION not set")
	}
	ctx := context.Background()
	p, err := NewPubSub(ctx, WithPubSubProjectID(proj), WithSubscriptionID(sub))
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}
	if _, err := p.ListReports(ctx); err != nil {
		t.Fatalf("failed to list reports: %v", err)
	}
}

type fakeSub struct{ msgs []*pubsub.Message }

func (f *fakeSub) Receive(ctx context.Context, fn func(context.Context, *pubsub.Message)) error {
	for _, m := range f.msgs {
		fn(ctx, m)
	}
	return nil
}

func TestPubSubFilter(t *testing.T) {
	finding1 := &securitycenterpb.Finding{Name: "f1"}
	note1 := &securitycenterpb.NotificationMessage{Event: &securitycenterpb.NotificationMessage_Finding{Finding: finding1}}
	data1, _ := protojson.Marshal(note1)

	finding2 := &securitycenterpb.Finding{Name: "f2", SecurityMarks: &securitycenterpb.SecurityMarks{Marks: map[string]string{"env": "nonprod"}}}
	note2 := &securitycenterpb.NotificationMessage{Event: &securitycenterpb.NotificationMessage_Finding{Finding: finding2}}
	data2, _ := protojson.Marshal(note2)

	sub := &fakeSub{msgs: []*pubsub.Message{{Data: data1}, {Data: data2}}}
	filter := func(f *securitycenterpb.Finding) bool {
		if m := f.GetSecurityMarks(); m != nil {
			if env, ok := m.Marks["env"]; ok && env == "nonprod" {
				return false
			}
		}
		return true
	}
	p, err := NewPubSub(context.Background(), WithSubscription(sub), WithPubSubFindingFilter(filter))
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}
	reports, err := p.ListReports(context.Background())
	if err != nil {
		t.Fatalf("failed to list reports: %v", err)
	}
	if len(reports) != 1 || reports[0].ID != "f1" {
		t.Fatalf("unexpected reports: %+v", reports)
	}
}
