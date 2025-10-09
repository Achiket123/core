package resolver

import (
	"context"

	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/pkg/cp"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/objects/storage"
)

// configureProviderRules adds the resolver rules that determine which provider to use for a request.
func configureProviderRules(resolver *providerResolver, config storage.ProviderConfig) {
	coordinator := newRuleCoordinator(resolver, config)
	if coordinator.handleDevMode() {
		return
	}

	coordinator.addKnownProviderRule()
	coordinator.addModuleRule(models.CatalogTrustCenterModule, storage.R2Provider)
	coordinator.addModuleRule(models.CatalogComplianceModule, storage.S3Provider)
	coordinator.addPreferredProviderRule()
	coordinator.addFallbackRule([]storage.ProviderType{storage.S3Provider, storage.R2Provider, storage.DiskProvider, storage.DatabaseProvider})
}

// ruleCoordinator groups state required to add resolver rules in a readable way.
type ruleCoordinator struct {
	resolver          *providerResolver
	config            storage.ProviderConfig
	moduleHint        cp.HintKey[models.OrgModule]
	preferredProvider cp.HintKey[storage.ProviderType]
	knownProvider     cp.HintKey[storage.ProviderType]
}

// newRuleCoordinator returns a helper for building provider rules.
func newRuleCoordinator(resolver *providerResolver, config storage.ProviderConfig) *ruleCoordinator {
	return &ruleCoordinator{
		resolver:          resolver,
		config:            config,
		moduleHint:        objects.ModuleHintKey(),
		preferredProvider: objects.PreferredProviderHintKey(),
		knownProvider:     objects.KnownProviderHintKey(),
	}
}

// handleDevMode adds a dev-only disk rule when configured and returns true if handled.
func (rc *ruleCoordinator) handleDevMode() bool {
	if !rc.config.DevMode {
		return false
	}

	options := storage.NewProviderOptions(
		storage.WithBucket(objects.DefaultDevStorageBucket),
		storage.WithBasePath(objects.DefaultDevStorageBucket),
		storage.WithExtra("dev_mode", true),
	)

	devRule := cp.NewRule[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]().
		Resolve(func(_ context.Context) (*cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions], error) {
			return &cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions]{
				Type:        cp.ProviderType(storage.DiskProvider),
				Credentials: storage.ProviderCredentials{},
				Config:      options.Clone(),
			}, nil
		})

	rc.resolver.AddRule(devRule)
	return true
}

// addKnownProviderRule resolves providers when a known provider hint is supplied.
func (rc *ruleCoordinator) addKnownProviderRule() {
	rule := cp.NewRule[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]().
		WhenFunc(func(ctx context.Context) bool {
			known := cp.GetHint(ctx, rc.knownProvider)
			return known.IsPresent() && rc.providerEnabled(known.MustGet())
		}).
		Resolve(func(ctx context.Context) (*cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions], error) {
			provider := cp.GetHint(ctx, rc.knownProvider).MustGet()
			return rc.resolveProvider(ctx, provider)
		})

	rc.resolver.AddRule(rule)
}

// addModuleRule routes requests for a specific module to the desired provider.
func (rc *ruleCoordinator) addModuleRule(module models.OrgModule, provider storage.ProviderType) {
	if !rc.providerEnabled(provider) {
		return
	}

	moduleProvider := provider
	rule := cp.NewRule[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]().
		WhenFunc(func(ctx context.Context) bool {
			hint := cp.GetHint(ctx, rc.moduleHint)
			return hint.IsPresent() && hint.MustGet() == module
		}).
		Resolve(func(ctx context.Context) (*cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions], error) {
			return rc.resolveProvider(ctx, moduleProvider)
		})

	rc.resolver.AddRule(rule)
}

// addPreferredProviderRule respects preferred provider hints when available.
func (rc *ruleCoordinator) addPreferredProviderRule() {
	rule := cp.NewRule[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]().
		WhenFunc(func(ctx context.Context) bool {
			preferred := cp.GetHint(ctx, rc.preferredProvider)
			return preferred.IsPresent() && rc.providerEnabled(preferred.MustGet())
		}).
		Resolve(func(ctx context.Context) (*cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions], error) {
			provider := cp.GetHint(ctx, rc.preferredProvider).MustGet()
			return rc.resolveProvider(ctx, provider)
		})

	rc.resolver.AddRule(rule)
}

// addFallbackRule registers the default provider order when no other hint applies.
func (rc *ruleCoordinator) addFallbackRule(order []storage.ProviderType) {
	for _, provider := range order {
		if !rc.providerEnabled(provider) {
			continue
		}

		fallbackProvider := provider
		rule := cp.NewRule[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]().
			WhenFunc(func(_ context.Context) bool {
				return true
			}).
			Resolve(func(ctx context.Context) (*cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions], error) {
				return rc.resolveProvider(ctx, fallbackProvider)
			})

		rc.resolver.AddRule(rule)
		break
	}
}
