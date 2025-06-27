package tasks

import (
	"context"
	"time"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"

	"github.com/theopenlane/core/pkg/compmgr"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// Manager coordinates task creation and reconciliation using compliance provider findings.
type GraphClient interface {
	CreateTask(ctx context.Context, input openlaneclient.CreateTaskInput, interceptors ...clientv2.RequestInterceptor) (*openlaneclient.CreateTask, error)
	UpdateTask(ctx context.Context, updateTaskID string, input openlaneclient.UpdateTaskInput, interceptors ...clientv2.RequestInterceptor) (*openlaneclient.UpdateTask, error)
}

// Manager coordinates task creation and reconciliation using compliance provider findings.
// Manager keeps track of provider findings and mirrors them as tasks in
// Openlane. The ownerID should be the organization that owns the new task and
// assigneeID may specify a default user to assign. Tasks are keyed by provider
// finding ID so they can be reconciled when findings are resolved.
type Manager struct {
	client     GraphClient
	ownerID    string
	assigneeID string
	tasks      map[string]string // provider finding ID -> task ID
}

// NewManager returns a Manager that creates tasks owned by the given organization and assigned to the given user.
func NewManager(client GraphClient, ownerID, assigneeID string) *Manager {
	return &Manager{
		client:     client,
		ownerID:    ownerID,
		assigneeID: assigneeID,
		tasks:      make(map[string]string),
	}
}

// Plan returns the tasks that would be created for new provider findings
// without actually creating them. Each planned task uses the provider report
// name as the title and the report description for details.
func (m *Manager) Plan(ctx context.Context, provider compmgr.Provider) ([]openlaneclient.CreateTaskInput, error) {
	reports, err := provider.ListReports(ctx)
	if err != nil {
		return nil, err
	}

	var inputs []openlaneclient.CreateTaskInput
	for _, r := range reports {
		if r.Passed {
			continue
		}
		if _, ok := m.tasks[r.ID]; ok {
			continue
		}

		tags := []string{"provider_id:" + r.ID}
		if r.Link != "" {
			tags = append(tags, "provider_url:"+r.Link)
		}
		input := openlaneclient.CreateTaskInput{
			Title:   r.Name,
			Details: &r.Description,
			OwnerID: &m.ownerID,
			Tags:    tags,
		}
		if m.assigneeID != "" {
			input.AssigneeID = &m.assigneeID
		}
		inputs = append(inputs, input)
	}

	return inputs, nil
}

// Sync pulls findings from the provider and ensures tasks are created for new
// findings and closed when findings are resolved. New tasks use the report name
// and description and are tagged with provider metadata for traceability.
func (m *Manager) Sync(ctx context.Context, provider compmgr.Provider) error {
	reports, err := provider.ListReports(ctx)
	if err != nil {
		return err
	}

	active := make(map[string]compmgr.Report, len(reports))
	for _, r := range reports {
		if r.Passed {
			continue
		}
		active[r.ID] = r
		if _, ok := m.tasks[r.ID]; !ok {
			// create new task
			tags := []string{"provider_id:" + r.ID}
			if r.Link != "" {
				tags = append(tags, "provider_url:"+r.Link)
			}
			input := openlaneclient.CreateTaskInput{
				Title:   r.Name,
				Details: &r.Description,
				OwnerID: &m.ownerID,
				Tags:    tags,
			}
			if m.assigneeID != "" {
				input.AssigneeID = &m.assigneeID
			}
			resp, err := m.client.CreateTask(ctx, input)
			if err != nil {
				return err
			}
			if t := resp.GetCreateTask(); t != nil {
				m.tasks[r.ID] = t.GetTask().ID
			}
		}
	}

	// close tasks whose findings have been resolved
	for findingID, taskID := range m.tasks {
		if _, ok := active[findingID]; !ok {
			now := models.DateTime(time.Now())
			status := enums.TaskStatusCompleted
			note := "Finding resolved by provider"
			input := openlaneclient.UpdateTaskInput{
				Status:    &status,
				Completed: &now,
				Details:   &note,
			}
			if _, err := m.client.UpdateTask(ctx, taskID, input); err != nil {
				return err
			}
			delete(m.tasks, findingID)
		}
	}

	return nil
}
