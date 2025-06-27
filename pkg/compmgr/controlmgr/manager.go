package controlmgr

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"

	"github.com/samber/lo"

	"github.com/theopenlane/core/pkg/compmgr"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// GraphClient defines the subset of the Openlane GraphQL client used by the Manager.
type GraphClient interface {
	UpdateControl(ctx context.Context, id string, input openlaneclient.UpdateControlInput, interceptors ...clientv2.RequestInterceptor) (*openlaneclient.UpdateControl, error)
}

// Manager updates controls based on provider reports.
type Manager struct {
	client GraphClient
}

// NewManager returns a Manager using the provided client.
func NewManager(client GraphClient) *Manager {
	return &Manager{client: client}
}

// Sync updates controls referenced by passing provider reports. Each control ID
// listed in the report is marked as approved via UpdateControl.
func (m *Manager) Sync(ctx context.Context, provider compmgr.Provider) error {
	reports, err := provider.ListReports(ctx)
	if err != nil {
		return err
	}
	for _, r := range reports {
		if !r.Passed {
			continue
		}
		for _, id := range r.ControlIDs {
			input := openlaneclient.UpdateControlInput{
				Status: lo.ToPtr(enums.ControlStatusApproved),
			}
			if _, err := m.client.UpdateControl(ctx, id, input); err != nil {
				return err
			}
		}
	}
	return nil
}
