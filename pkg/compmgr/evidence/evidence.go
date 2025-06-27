package evidence

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	clientv2 "github.com/Yamashou/gqlgenc/clientv2"

	"github.com/theopenlane/core/pkg/compmgr"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// GraphClient defines the subset of the Openlane GraphQL client used by the Manager.
type GraphClient interface {
	CreateEvidence(ctx context.Context, input openlaneclient.CreateEvidenceInput, evidenceFiles []*graphql.Upload, interceptors ...clientv2.RequestInterceptor) (*openlaneclient.CreateEvidence, error)
}

// Manager uploads evidence for passing provider reports.
type Manager struct {
	client  GraphClient
	ownerID string
}

// NewManager returns a Manager storing evidence under the given organization.
func NewManager(client GraphClient, ownerID string) *Manager {
	return &Manager{client: client, ownerID: ownerID}
}

// Sync uploads evidence for provider reports marked as passed.
func (m *Manager) Sync(ctx context.Context, provider compmgr.Provider) error {
	reports, err := provider.ListReports(ctx)
	if err != nil {
		return err
	}
	for _, r := range reports {
		if !r.Passed {
			continue
		}

		ids := r.ControlIDs
		if len(ids) == 0 {
			for _, c := range r.Controls {
				ids = append(ids, c.IDs...)
			}
		}

		input := openlaneclient.CreateEvidenceInput{
			Name:        r.Name,
			Description: &r.Description,
			URL:         &r.Link,
			Source:      &r.Link,
			ControlIDs:  ids,
			OwnerID:     &m.ownerID,
		}
		if _, err := m.client.CreateEvidence(ctx, input, nil); err != nil {
			return err
		}
	}
	return nil
}
