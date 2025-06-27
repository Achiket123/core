package gcp

import (
	"context"
	"os"
	"testing"
)

func TestValidateSetup(t *testing.T) {
	proj := os.Getenv("GCP_PROJECT")
	sub := os.Getenv("PUBSUB_SUBSCRIPTION")
	if proj == "" || sub == "" {
		t.Skip("GCP_PROJECT or PUBSUB_SUBSCRIPTION not set")
	}
	if err := ValidateSetup(context.Background(), []string{proj}, proj, sub); err != nil {
		t.Fatalf("setup validation failed: %v", err)
	}
}
