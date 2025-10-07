package cp

import (
	"context"
	"testing"

	"github.com/samber/mo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/pkg/models"
)

// Mock client type for testing
type MockClient struct {
	ID string
}

func TestNewResolver(t *testing.T) {
	resolver := NewResolver[MockClient, models.CredentialSet, map[string]any]()

	assert.NotNil(t, resolver)
	assert.Empty(t, resolver.rules)
	assert.False(t, resolver.defaultRule.IsPresent())
}

func TestResolver_AddRule(t *testing.T) {
	resolver := NewResolver[MockClient, models.CredentialSet, map[string]any]()

	rule := ResolutionRule[MockClient, models.CredentialSet, map[string]any]{
		Evaluate: func(ctx context.Context) mo.Option[Resolution[models.CredentialSet, map[string]any]] {
			return mo.Some(Resolution[models.CredentialSet, map[string]any]{ClientType: "test"})
		},
	}

	result := resolver.AddRule(rule)

	assert.Same(t, resolver, result) // Should return self for chaining
	assert.Len(t, resolver.rules, 1)
	// Rule was added successfully
	assert.NotNil(t, resolver.rules[0].Evaluate)
}

func TestResolver_SetDefaultRule(t *testing.T) {
	resolver := NewResolver[MockClient, models.CredentialSet, map[string]any]()

	defaultRule := ResolutionRule[MockClient, models.CredentialSet, map[string]any]{
		Evaluate: func(ctx context.Context) mo.Option[Resolution[models.CredentialSet, map[string]any]] {
			return mo.Some(Resolution[models.CredentialSet, map[string]any]{ClientType: "default"})
		},
	}

	result := resolver.SetDefaultRule(defaultRule)

	assert.Same(t, resolver, result)
	assert.True(t, resolver.defaultRule.IsPresent())
	// Default rule was set successfully
	assert.NotNil(t, resolver.defaultRule.MustGet().Evaluate)
}

func TestResolver_Resolve_FallsBackToDefault(t *testing.T) {
	resolver := NewResolver[MockClient, models.CredentialSet, map[string]any]()

	// Add rule that won't match
	rule := ResolutionRule[MockClient, models.CredentialSet, map[string]any]{
		Evaluate: func(ctx context.Context) mo.Option[Resolution[models.CredentialSet, map[string]any]] {
			if testType := GetValue[string](ctx); testType.IsPresent() && testType.MustGet() == "specific" {
				return mo.Some(Resolution[models.CredentialSet, map[string]any]{ClientType: "specific"})
			}
			return mo.None[Resolution[models.CredentialSet, map[string]any]]()
		},
	}

	defaultRule := ResolutionRule[MockClient, models.CredentialSet, map[string]any]{
		Evaluate: func(ctx context.Context) mo.Option[Resolution[models.CredentialSet, map[string]any]] {
			return mo.Some(Resolution[models.CredentialSet, map[string]any]{ClientType: "default"})
		},
	}

	resolver.AddRule(rule)
	resolver.SetDefaultRule(defaultRule)

	ctx := WithValue(context.Background(), "other")
	result := resolver.Resolve(ctx)

	require.True(t, result.IsPresent())
	resolution := result.MustGet()
	assert.Equal(t, ProviderType("default"), resolution.ClientType)
}

func TestResolver_Resolve_NoMatch(t *testing.T) {
	resolver := NewResolver[MockClient, models.CredentialSet, map[string]any]()

	// Add rule that won't match
	rule := ResolutionRule[MockClient, models.CredentialSet, map[string]any]{
		Evaluate: func(ctx context.Context) mo.Option[Resolution[models.CredentialSet, map[string]any]] {
			return mo.None[Resolution[models.CredentialSet, map[string]any]]()
		},
	}

	resolver.AddRule(rule)

	ctx := WithValue(context.Background(), "test")
	result := resolver.Resolve(ctx)

	assert.False(t, result.IsPresent())
}

func TestResolver_Resolve_WithCacheKey(t *testing.T) {
	resolver := NewResolver[MockClient, models.CredentialSet, map[string]any]()

	rule := ResolutionRule[MockClient, models.CredentialSet, map[string]any]{
		Evaluate: func(ctx context.Context) mo.Option[Resolution[models.CredentialSet, map[string]any]] {
			return mo.Some(Resolution[models.CredentialSet, map[string]any]{
				ClientType: "test",
				CacheKey:   ClientCacheKey{TenantID: "tenant", IntegrationType: "test", IntegrationID: "integration-test"},
			})
		},
	}

	resolver.AddRule(rule)

	ctx := WithValue(context.Background(), "123")
	result := resolver.Resolve(ctx)

	require.True(t, result.IsPresent())
	resolution := result.MustGet()
	assert.Equal(t, ClientCacheKey{TenantID: "tenant", IntegrationType: "test", IntegrationID: "integration-test"}, resolution.CacheKey)
}

func TestResolution_Structure(t *testing.T) {
	resolution := Resolution[models.CredentialSet, map[string]any]{
		ClientType: "test-client",
		Credentials: models.CredentialSet{
			AccessKeyID:     "user",
			SecretAccessKey: "pass",
		},
		Config: map[string]any{
			"timeout": 30,
			"retries": 3,
		},
		CacheKey: ClientCacheKey{TenantID: "tenant", IntegrationType: "test", IntegrationID: "integration-test"},
	}

	assert.Equal(t, ProviderType("test-client"), resolution.ClientType)
	assert.Equal(t, "user", resolution.Credentials.AccessKeyID)
	assert.Equal(t, "pass", resolution.Credentials.SecretAccessKey)
	assert.Equal(t, 30, resolution.Config["timeout"])
	assert.Equal(t, 3, resolution.Config["retries"])
	assert.Equal(t, "tenant", resolution.CacheKey.TenantID)
}
