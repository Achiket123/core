package evidence

import (
	"context"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	"github.com/stretchr/testify/mock"

	"github.com/theopenlane/core/pkg/compmgr"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

type mockProvider struct{ reports []compmgr.Report }

func (m mockProvider) ListReports(context.Context) ([]compmgr.Report, error) {
	return m.reports, nil
}

type mockClient struct{ mock.Mock }

func (m *mockClient) CreateEvidence(ctx context.Context, input openlaneclient.CreateEvidenceInput, files []*graphql.Upload, i ...clientv2.RequestInterceptor) (*openlaneclient.CreateEvidence, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*openlaneclient.CreateEvidence), args.Error(1)
}

func TestManagerSync(t *testing.T) {
	client := new(mockClient)
	manager := NewManager(client, "org1")

	provider := mockProvider{reports: []compmgr.Report{{ID: "p1", Name: "pass", Passed: true, Link: "https://provider"}}}

	res := &openlaneclient.CreateEvidence{}
	client.On("CreateEvidence", mock.Anything, mock.Anything).Return(res, nil)

	if err := manager.Sync(context.Background(), provider); err != nil {
		t.Fatalf("sync failed: %v", err)
	}
}
