package resolver

import (
	"context"

	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/pkg/eddy"
	"github.com/theopenlane/core/pkg/eddy/helpers"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/utils/contextx"
)

// configureProviderRules adds the resolver rules that determine which provider to use for a request
func configureProviderRules(
	resolver *providerResolver,
	config storage.ProviderConfig,
	s3Builder eddy.Builder[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions],
	r2Builder eddy.Builder[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions],
	diskBuilder eddy.Builder[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions],
	dbBuilder eddy.Builder[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions],
) {
	coordinator := newRuleCoordinator(resolver, config, s3Builder, r2Builder, diskBuilder, dbBuilder)
	if coordinator.handleDevMode() {
		return
	}

	coordinator.addKnownProviderRule()
	coordinator.addModuleRule(models.CatalogTrustCenterModule, storage.R2Provider)
	coordinator.addModuleRule(models.CatalogComplianceModule, storage.S3Provider)
	coordinator.addPreferredProviderRule()
	coordinator.addFallbackRule([]storage.ProviderType{storage.S3Provider, storage.R2Provider, storage.DiskProvider, storage.DatabaseProvider})
}

// ruleCoordinator groups state required to add resolver rules in a readable way
type ruleCoordinator struct {
	resolver    *providerResolver
	config      storage.ProviderConfig
	s3Builder   eddy.Builder[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]
	r2Builder   eddy.Builder[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]
	diskBuilder eddy.Builder[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]
	dbBuilder   eddy.Builder[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]
}

// newRuleCoordinator returns a helper for building provider rules
func newRuleCoordinator(
	resolver *providerResolver,
	config storage.ProviderConfig,
	s3Builder eddy.Builder[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions],
	r2Builder eddy.Builder[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions],
	diskBuilder eddy.Builder[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions],
	dbBuilder eddy.Builder[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions],
) *ruleCoordinator {
	return &ruleCoordinator{
		resolver:    resolver,
		config:      config,
		s3Builder:   s3Builder,
		r2Builder:   r2Builder,
		diskBuilder: diskBuilder,
		dbBuilder:   dbBuilder,
	}
}

// getBuilder returns the appropriate builder for a provider type
func (rc *ruleCoordinator) getBuilder(provider storage.ProviderType) eddy.Builder[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions] {
	switch provider {
	case storage.S3Provider:
		return rc.s3Builder
	case storage.R2Provider:
		return rc.r2Builder
	case storage.DiskProvider:
		return rc.diskBuilder
	case storage.DatabaseProvider:
		return rc.dbBuilder
	default:
		return nil
	}
}

// handleDevMode adds a dev-only disk rule when configured and returns true if handled
func (rc *ruleCoordinator) handleDevMode() bool {
	if !rc.config.DevMode {
		return false
	}

	options := storage.NewProviderOptions(
		storage.WithBucket(objects.DefaultDevStorageBucket),
		storage.WithBasePath(objects.DefaultDevStorageBucket),
		storage.WithExtra("dev_mode", true),
	)

	devRule := &helpers.ConditionalRule[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		Predicate: func(_ context.Context) bool {
			return true
		},
		Resolver: func(_ context.Context) (*eddy.ResolvedProvider[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions], error) {
			return &eddy.ResolvedProvider[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
				Builder: rc.diskBuilder,
				Output:  storage.ProviderCredentials{},
				Config:  options.Clone(),
			}, nil
		},
	}

	rc.resolver.AddRule(devRule)
	return true
}

// addKnownProviderRule resolves providers when a known provider hint is supplied
func (rc *ruleCoordinator) addKnownProviderRule() {
	rule := &helpers.ConditionalRule[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		Predicate: func(ctx context.Context) bool {
			known, ok := contextx.From[objects.KnownProviderHint](ctx)
			return ok && rc.providerEnabled(storage.ProviderType(known))
		},
		Resolver: func(ctx context.Context) (*eddy.ResolvedProvider[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions], error) {
			known, _ := contextx.From[objects.KnownProviderHint](ctx)
			provider := storage.ProviderType(known)
			return rc.resolveProviderWithBuilder(ctx, provider)
		},
	}

	rc.resolver.AddRule(rule)
}

// addModuleRule routes requests for a specific module to the desired provider
func (rc *ruleCoordinator) addModuleRule(module models.OrgModule, provider storage.ProviderType) {
	if !rc.providerEnabled(provider) {
		return
	}

	moduleProvider := provider
	rule := &helpers.ConditionalRule[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		Predicate: func(ctx context.Context) bool {
			hint, ok := contextx.From[objects.ModuleHint](ctx)
			return ok && models.OrgModule(hint) == module
		},
		Resolver: func(ctx context.Context) (*eddy.ResolvedProvider[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions], error) {
			return rc.resolveProviderWithBuilder(ctx, moduleProvider)
		},
	}

	rc.resolver.AddRule(rule)
}

// addPreferredProviderRule respects preferred provider hints when available
func (rc *ruleCoordinator) addPreferredProviderRule() {
	rule := &helpers.ConditionalRule[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		Predicate: func(ctx context.Context) bool {
			preferred, ok := contextx.From[objects.PreferredProviderHint](ctx)
			return ok && rc.providerEnabled(storage.ProviderType(preferred))
		},
		Resolver: func(ctx context.Context) (*eddy.ResolvedProvider[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions], error) {
			preferred, _ := contextx.From[objects.PreferredProviderHint](ctx)
			provider := storage.ProviderType(preferred)
			return rc.resolveProviderWithBuilder(ctx, provider)
		},
	}

	rc.resolver.AddRule(rule)
}

// addFallbackRule registers the default provider order when no other hint applies
func (rc *ruleCoordinator) addFallbackRule(order []storage.ProviderType) {
	for _, provider := range order {
		if !rc.providerEnabled(provider) {
			continue
		}

		fallbackProvider := provider
		rule := &helpers.ConditionalRule[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
			Predicate: func(_ context.Context) bool {
				return true
			},
			Resolver: func(ctx context.Context) (*eddy.ResolvedProvider[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions], error) {
				return rc.resolveProviderWithBuilder(ctx, fallbackProvider)
			},
		}

		rc.resolver.AddRule(rule)
		break
	}
}

// resolveProviderWithBuilder resolves provider credentials and returns them with the appropriate builder
func (rc *ruleCoordinator) resolveProviderWithBuilder(ctx context.Context, provider storage.ProviderType) (*eddy.ResolvedProvider[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions], error) {
	resolved, err := rc.resolveProvider(ctx, provider)
	if err != nil {
		return nil, err
	}

	builder := rc.getBuilder(provider)
	if builder == nil {
		return nil, errUnsupportedProvider
	}

	return &eddy.ResolvedProvider[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		Builder: builder,
		Output:  resolved.Output,
		Config:  resolved.Config,
	}, nil
}
