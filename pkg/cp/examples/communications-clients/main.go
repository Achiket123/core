package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/samber/mo"
	"github.com/theopenlane/core/pkg/cp"
)

// MessageClient is the interface all communication providers must implement
type MessageClient interface {
	SendMessage(ctx context.Context, channel, message string) error
	SendDirectMessage(ctx context.Context, userID, message string) error
	GetChannels(ctx context.Context) ([]string, error)
	Close() error
}

// SlackBuilder builds Slack clients
type SlackBuilder struct {
	creds  map[string]string
	config map[string]any
}

func (b *SlackBuilder) WithCredentials(creds map[string]string) cp.ClientBuilder[MessageClient] {
	newBuilder := *b
	newBuilder.creds = creds
	return &newBuilder
}

func (b *SlackBuilder) WithConfig(config map[string]any) cp.ClientBuilder[MessageClient] {
	newBuilder := *b
	newBuilder.config = config
	return &newBuilder
}

func (b *SlackBuilder) Build(ctx context.Context) (MessageClient, error) {
	baseURL := b.creds["base_url"]
	if baseURL == "" {
		baseURL = os.Getenv("SLACK_BASE_URL")
	}
	return newSlackClient(b.creds["token"], b.creds["team_id"], baseURL), nil
}

func (b *SlackBuilder) ClientType() cp.ProviderType {
	return "slack"
}

// TeamsBuilder builds Microsoft Teams clients
type TeamsBuilder struct {
	creds  map[string]string
	config map[string]any
}

func (b *TeamsBuilder) WithCredentials(creds map[string]string) cp.ClientBuilder[MessageClient] {
	newBuilder := *b
	newBuilder.creds = creds
	return &newBuilder
}

func (b *TeamsBuilder) WithConfig(config map[string]any) cp.ClientBuilder[MessageClient] {
	newBuilder := *b
	newBuilder.config = config
	return &newBuilder
}

func (b *TeamsBuilder) Build(ctx context.Context) (MessageClient, error) {
	webhookURL := b.creds["webhook_url"]
	if webhookURL == "" {
		webhookURL = os.Getenv("TEAMS_WEBHOOK_URL")
	}
	if webhookURL == "" {
		webhookURL = "http://localhost:8081/teams/webhook"
	}
	return newTeamsClient(webhookURL), nil
}

func (b *TeamsBuilder) ClientType() cp.ProviderType {
	return "teams"
}

// DiscordBuilder builds Discord clients
type DiscordBuilder struct {
	creds  map[string]string
	config map[string]any
}

func (b *DiscordBuilder) WithCredentials(creds map[string]string) cp.ClientBuilder[MessageClient] {
	newBuilder := *b
	newBuilder.creds = creds
	return &newBuilder
}

func (b *DiscordBuilder) WithConfig(config map[string]any) cp.ClientBuilder[MessageClient] {
	newBuilder := *b
	newBuilder.config = config
	return &newBuilder
}

func (b *DiscordBuilder) Build(ctx context.Context) (MessageClient, error) {
	webhookURL := b.creds["webhook_url"]
	if webhookURL == "" {
		webhookURL = os.Getenv("DISCORD_WEBHOOK_URL")
	}
	if webhookURL == "" {
		webhookURL = "http://localhost:8082/discord/webhook"
	}
	return newDiscordClient(webhookURL), nil
}

func (b *DiscordBuilder) ClientType() cp.ProviderType {
	return "discord"
}

// TenantConfig represents tenant communication preferences
type TenantConfig struct {
	ID                  string
	PreferredPlatform   string
	SlackToken          string
	SlackTeamID         string
	SlackBaseURL        string
	TeamsWebhookURL     string
	DiscordWebhookURL   string
	NotificationChannel string
}

func main() {
	ctx := context.Background()

	fmt.Println("=== Communications Client Pooling Example ===\n")

	// Start mock server in background
	mockServer := NewMockServer("8080")
	go func() {
		if err := mockServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("Mock server error: %v", err)
		}
	}()
	defer mockServer.Stop()

	// Give mock server time to start
	time.Sleep(500 * time.Millisecond)

	pool := cp.NewClientPool[MessageClient](30 * time.Minute)
	service := cp.NewClientService(pool)

	service.RegisterBuilder("slack", &SlackBuilder{})
	service.RegisterBuilder("teams", &TeamsBuilder{})
	service.RegisterBuilder("discord", &DiscordBuilder{})

	tenants := []TenantConfig{
		{
			ID:                  "tenant-1",
			PreferredPlatform:   "slack",
			SlackToken:          "xoxb-tenant1-slack-token",
			SlackTeamID:         "T01234567",
			SlackBaseURL:        "http://localhost:8080/slack",
			NotificationChannel: "general",
		},
		{
			ID:                  "tenant-2",
			PreferredPlatform:   "teams",
			TeamsWebhookURL:     "http://localhost:8080/teams/webhook",
			NotificationChannel: "General",
		},
		{
			ID:                  "tenant-3",
			PreferredPlatform:   "discord",
			DiscordWebhookURL:   "http://localhost:8080/discord/webhook",
			NotificationChannel: "general",
		},
		{
			ID:                  "tenant-4",
			PreferredPlatform:   "slack",
			SlackToken:          "xoxb-tenant4-slack-token",
			SlackTeamID:         "T98765432",
			SlackBaseURL:        "http://localhost:8080/slack",
			NotificationChannel: "announcements",
		},
	}

	resolver := cp.NewResolver[MessageClient]()
	resolver.AddRule(cp.ResolutionRule[MessageClient]{
		Evaluate: func(ctx context.Context) mo.Option[cp.Resolution] {
			tenantCfg := cp.GetValue[TenantConfig](ctx)
			if !tenantCfg.IsPresent() {
				return mo.None[cp.Resolution]()
			}

			tenant := tenantCfg.MustGet()

			var resolution cp.Resolution
			switch tenant.PreferredPlatform {
			case "slack":
				resolution = cp.Resolution{
					ClientType: "slack",
					Credentials: map[string]string{
						"token":    tenant.SlackToken,
						"team_id":  tenant.SlackTeamID,
						"base_url": tenant.SlackBaseURL,
					},
				}
			case "teams":
				resolution = cp.Resolution{
					ClientType: "teams",
					Credentials: map[string]string{
						"webhook_url": tenant.TeamsWebhookURL,
					},
				}
			case "discord":
				resolution = cp.Resolution{
					ClientType: "discord",
					Credentials: map[string]string{
						"webhook_url": tenant.DiscordWebhookURL,
					},
				}
			default:
				return mo.None[cp.Resolution]()
			}

			return mo.Some(resolution)
		},
	})

	fmt.Println("Demonstrating multi-tenant communications routing:\n")

	for _, tenant := range tenants {
		fmt.Printf("--- Tenant: %s (Platform: %s) ---\n", tenant.ID, tenant.PreferredPlatform)

		tenantCtx := cp.WithValue(ctx, tenant)

		resolution := resolver.Resolve(tenantCtx)
		if !resolution.IsPresent() {
			fmt.Printf("  ❌ No resolution found for tenant %s\n\n", tenant.ID)
			continue
		}

		res := resolution.MustGet()

		cacheKey := cp.ClientCacheKey{
			TenantID:        tenant.ID,
			IntegrationType: string(res.ClientType),
		}

		client := service.GetClient(tenantCtx, cacheKey, res.ClientType, res.Credentials, res.Config)
		if !client.IsPresent() {
			fmt.Printf("  ❌ Failed to get client for tenant %s\n\n", tenant.ID)
			continue
		}

		c := client.MustGet()

		if err := c.SendMessage(tenantCtx, tenant.NotificationChannel, "Hello from multi-tenant system!"); err != nil {
			log.Printf("  Error: %v\n", err)
		}

		if err := c.SendDirectMessage(tenantCtx, "user123", "Direct notification"); err != nil {
			log.Printf("  Error: %v\n", err)
		}

		channels, err := c.GetChannels(tenantCtx)
		if err != nil {
			log.Printf("  Error getting channels: %v\n", err)
		} else {
			fmt.Printf("  Available channels: %v\n", channels)
		}

		fmt.Println()
	}

	fmt.Println("Demonstrating cache efficiency:")
	fmt.Println("Sending second message to tenant-1 (should use cached client):\n")

	tenant1Ctx := cp.WithValue(ctx, tenants[0])
	resolution := resolver.Resolve(tenant1Ctx)
	res := resolution.MustGet()

	cacheKey := cp.ClientCacheKey{
		TenantID:        tenants[0].ID,
		IntegrationType: string(res.ClientType),
	}

	cachedClient := service.GetClient(tenant1Ctx, cacheKey, res.ClientType, res.Credentials, res.Config)
	if cachedClient.IsPresent() {
		c := cachedClient.MustGet()
		c.SendMessage(tenant1Ctx, "general", "This message uses the cached client!")
	}

	fmt.Println("\n=== Example completed successfully ===")
}
