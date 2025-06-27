package controlmgr

import (
	"context"
	"testing"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	"github.com/stretchr/testify/mock"

	"github.com/theopenlane/core/pkg/compmgr"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

type mockProvider struct{ reports []compmgr.Report }

func (m mockProvider) ListReports(context.Context) ([]compmgr.Report, error) {
	return m.reports, nil
}

// mockClient implements the GraphClient interface for testing.
type mockClient struct{ mock.Mock }

func (m *mockClient) UpdateControl(ctx context.Context, id string, input openlaneclient.UpdateControlInput, i ...clientv2.RequestInterceptor) (*openlaneclient.UpdateControl, error) {
	args := m.Called(ctx, id, input)
	return args.Get(0).(*openlaneclient.UpdateControl), args.Error(1)
}

func TestManagerSync(t *testing.T) {
	client := new(mockClient)
	m := NewManager(client)
	provider := mockProvider{reports: []compmgr.Report{{ID: "r1", Passed: true, ControlIDs: []string{"c1"}}}}

	updateRes := &openlaneclient.UpdateControl{}
	client.On("UpdateControl", mock.Anything, "c1", mock.Anything).Return(updateRes, nil)

	if err := m.Sync(context.Background(), provider); err != nil {
		t.Fatalf("sync failed: %v", err)
	}
	client.AssertExpectations(t)
}

func TestManagerSyncIgnoresFailing(t *testing.T) {
	client := new(mockClient)
	m := NewManager(client)
	provider := mockProvider{reports: []compmgr.Report{{ID: "r1", Passed: false, ControlIDs: []string{"c1"}}}}

	if err := m.Sync(context.Background(), provider); err != nil {
		t.Fatalf("sync failed: %v", err)
	}
	client.AssertExpectations(t)
}
