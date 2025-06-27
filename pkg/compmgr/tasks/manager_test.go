package tasks

import (
	"context"
	"testing"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	"github.com/stretchr/testify/mock"

	"github.com/theopenlane/core/pkg/compmgr"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

type mockProvider struct {
	reports []compmgr.Report
	err     error
}

func (m mockProvider) ListReports(context.Context) ([]compmgr.Report, error) {
	return m.reports, m.err
}

// mockClient satisfies the OpenlaneGraphClient interface for testing.
type mockClient struct{ mock.Mock }

func (m *mockClient) CreateTask(ctx context.Context, input openlaneclient.CreateTaskInput, i ...clientv2.RequestInterceptor) (*openlaneclient.CreateTask, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*openlaneclient.CreateTask), args.Error(1)
}

func (m *mockClient) UpdateTask(ctx context.Context, id string, input openlaneclient.UpdateTaskInput, i ...clientv2.RequestInterceptor) (*openlaneclient.UpdateTask, error) {
	args := m.Called(ctx, id, input)
	return args.Get(0).(*openlaneclient.UpdateTask), args.Error(1)
}

func TestManagerSync(t *testing.T) {
	client := new(mockClient)
	manager := NewManager(client, "org1", "user1")

	// simulate create success
	createRes := &openlaneclient.CreateTask{
		CreateTask: openlaneclient.CreateTask_CreateTask{
			Task: openlaneclient.CreateTask_CreateTask_Task{ID: "1"},
		},
	}
	client.On("CreateTask", mock.Anything, mock.MatchedBy(func(input openlaneclient.CreateTaskInput) bool {
		return len(input.Tags) == 2 &&
			input.Tags[0] == "provider_id:f1" &&
			input.Tags[1] == "provider_url:https://link"
	})).Return(createRes, nil)
	updateRes := &openlaneclient.UpdateTask{}
	client.On("UpdateTask", mock.Anything, mock.Anything, mock.Anything).Return(updateRes, nil)

	provider := mockProvider{reports: []compmgr.Report{{ID: "f1", Name: "issue", Passed: false, Link: "https://link"}, {ID: "p1", Name: "pass", Passed: true}}}
	if err := manager.Sync(context.Background(), provider); err != nil {
		t.Fatalf("sync failed: %v", err)
	}
	client.AssertNumberOfCalls(t, "CreateTask", 1)

	// second sync with no reports should mark task complete
	provider.reports = nil
	if err := manager.Sync(context.Background(), provider); err != nil {
		t.Fatalf("sync failed: %v", err)
	}
}

func TestManagerPlan(t *testing.T) {
	manager := NewManager(nil, "org1", "user1")
	provider := mockProvider{reports: []compmgr.Report{{ID: "f1", Name: "issue", Passed: false, Link: "https://link"}}}

	inputs, err := manager.Plan(context.Background(), provider)
	if err != nil {
		t.Fatalf("plan failed: %v", err)
	}
	if len(inputs) != 1 {
		t.Fatalf("expected 1 task, got %d", len(inputs))
	}
	if inputs[0].Title != "issue" {
		t.Fatalf("unexpected task title: %s", inputs[0].Title)
	}
	if len(inputs[0].Tags) != 2 || inputs[0].Tags[0] != "provider_id:f1" || inputs[0].Tags[1] != "provider_url:https://link" {
		t.Fatalf("unexpected tags: %v", inputs[0].Tags)
	}
}

func TestManagerPlanIgnoresPassing(t *testing.T) {
	manager := NewManager(nil, "org1", "user1")
	provider := mockProvider{reports: []compmgr.Report{{ID: "p1", Name: "pass", Passed: true}}}

	inputs, err := manager.Plan(context.Background(), provider)
	if err != nil {
		t.Fatalf("plan failed: %v", err)
	}
	if len(inputs) != 0 {
		t.Fatalf("expected 0 tasks, got %d", len(inputs))
	}
}
