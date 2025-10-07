package main

import (
	"context"
	"fmt"
	"time"

	"github.com/theopenlane/core/pkg/cp"
)

// Priority represents message priority levels
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityNormal   Priority = "normal"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// Environment represents deployment environment
type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvStaging     Environment = "staging"
	EnvProduction  Environment = "production"
)

// MessageContext contains metadata for message routing
type MessageContext struct {
	TenantID    string
	Environment Environment
	Priority    Priority
	Region      string
	UserCount   int
}

// AdvancedExample demonstrates complex routing scenarios
func AdvancedExample() {
	ctx := context.Background()

	pool := cp.NewClientPool[MessageClient](30 * time.Minute)
	service := cp.NewClientService(pool)

	service.RegisterBuilder("slack", &SlackBuilder{})
	service.RegisterBuilder("teams", &TeamsBuilder{})
	service.RegisterBuilder("discord", &DiscordBuilder{})
	service.RegisterBuilder("slack-premium", &SlackBuilder{})
	service.RegisterBuilder("teams-premium", &TeamsBuilder{})

	priorityResolver := cp.NewResolver[MessageClient, map[string]string, map[string]any]()

	priorityResolver.AddRule(cp.NewRule[MessageClient, map[string]string, map[string]any]().
		WhenFunc(func(ctx context.Context) bool {
			// pass in your type or any type
			msgCtx := cp.GetValue[MessageContext](ctx)
			// what conditions do you want to evaluate
			return msgCtx.IsPresent() && msgCtx.MustGet().Priority == PriorityCritical
		}).
		Resolve(func(ctx context.Context) (*cp.ResolvedProvider, error) {
			return &cp.ResolvedProvider[map[string]string, map[string]any]{
				Type: "slack",
				Credentials: map[string]string{
					"token":   "xoxb-critical-alerts",
					"team_id": "T_ONCALL",
				},
			}, nil
		}))

	priorityResolver.AddRule(cp.NewRule[MessageClient, map[string]string, map[string]any]().
		WhenFunc(func(ctx context.Context) bool {
			msgCtx := cp.GetValue[MessageContext](ctx)
			return msgCtx.IsPresent() && Priority(msgCtx.MustGet().TenantID) == PriorityNormal
		}).
		Resolve(func(ctx context.Context) (*cp.ResolvedProvider, error) {
			return &cp.ResolvedProvider[map[string]string, map[string]any]{
				Type: "discord",
				Credentials: map[string]string{
					"token":    "discord-normal-bot",
					"guild_id": "123456789",
				},
			}, nil
		}))

	priorityResolver.SetDefaultRule(cp.DefaultRule[MessageClient, map[string]string, map[string]any](cp.Resolution[map[string]string, map[string]any]{
		ClientType: "teams",
		Credentials: map[string]string{
			"token":     "teams-low-priority",
			"tenant_id": "default-tenant",
		},
	}))

	testPriorities := []Priority{PriorityCritical, PriorityHigh, PriorityNormal, PriorityLow}
	for _, priority := range testPriorities {
		msgCtx := MessageContext{
			TenantID: "tenant-1",
			Priority: priority,
		}
		ctx := cp.WithValue(ctx, msgCtx)

		resolution := priorityResolver.Resolve(ctx)
		if resolution.IsPresent() {
			res := resolution.MustGet()

			cacheKey := cp.ClientCacheKey{
				TenantID:        msgCtx.TenantID,
				IntegrationType: string(res.ClientType),
			}
			client := service.GetClient(ctx, cacheKey, res.ClientType, res.Credentials, res.Config)
			if client.IsPresent() {
				c := client.MustGet()
				c.SendMessage(ctx, "alerts", fmt.Sprintf("[%s] Test message", priority))
			}
		}
	}
}
