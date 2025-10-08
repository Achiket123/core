package resolver

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hush"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/pkg/cp"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/core/pkg/objects/storage/providers/disk"
	r2provider "github.com/theopenlane/core/pkg/objects/storage/providers/r2"
	s3provider "github.com/theopenlane/core/pkg/objects/storage/providers/s3"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"
)

// providerResolver simplifies references to the cp resolver used for object providers.
type providerResolver = cp.Resolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]

// providerClientService simplifies references to the cp client service used for object providers.
type providerClientService = cp.ClientService[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]

var (
	errUnsupportedProvider = errors.New("unsupported storage provider")
	errProviderDisabled    = errors.New("storage provider disabled")
)

// NewServiceFromConfig constructs a storage service complete with resolver rules derived from runtime configuration.
func NewServiceFromConfig(config storage.ProviderConfig) *objects.Service {
	clientService, resolver := Build(config)

	service := objects.NewService(objects.Config{
		Resolver:       resolver,
		ClientService:  clientService,
		ValidationFunc: objects.MimeTypeValidator,
	})

	return service
}

// Build constructs the cp client service and provider resolver from runtime configuration.
func Build(config storage.ProviderConfig) (*providerClientService, *providerResolver) {
	pool := cp.NewClientPool[storage.Provider](objects.DefaultClientPoolTTL)
	clientService := cp.NewClientService(pool, cp.WithConfigClone[
		storage.Provider,
		storage.ProviderCredentials](cloneProviderOptions))

	clientService.RegisterBuilder(cp.ProviderType(storage.S3Provider), s3provider.NewS3Builder())
	clientService.RegisterBuilder(cp.ProviderType(storage.R2Provider), r2provider.NewR2Builder())
	clientService.RegisterBuilder(cp.ProviderType(storage.DiskProvider), disk.NewDiskBuilder())

	resolver := cp.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	configureProviderRules(resolver, config)

	return clientService, resolver
}

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
	coordinator.addFallbackRule([]storage.ProviderType{storage.S3Provider, storage.R2Provider, storage.DiskProvider})
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

// providerEnabled returns whether a provider can be used based on configuration.
func (rc *ruleCoordinator) providerEnabled(provider storage.ProviderType) bool {
	switch provider {
	case storage.R2Provider:
		return rc.config.Providers.CloudflareR2.Enabled
	case storage.S3Provider:
		return rc.config.Providers.S3.Enabled
	case storage.DiskProvider:
		return rc.config.Providers.Disk.Enabled
	default:
		return false
	}
}

// resolveProvider returns provider credentials from system integrations or config fallback.
func (rc *ruleCoordinator) resolveProvider(ctx context.Context, provider storage.ProviderType) (*cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions], error) {
	if resolved, err := querySystemProvider(ctx, provider); err == nil {
		return resolved, nil
	} else if err != nil && !errors.Is(err, objects.ErrNoSystemIntegration) && !errors.Is(err, objects.ErrNoIntegrationWithSecrets) {
		log.Ctx(ctx).Warn().Err(err).Str("provider", string(provider)).Msg("system provider lookup failed, falling back to config")
	}

	return resolveProviderFromConfig(provider, rc.config)
}

func querySystemProvider(ctx context.Context, providerType storage.ProviderType) (*cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions], error) {
	entClient := ent.FromContext(ctx)
	if entClient == nil {
		return nil, objects.ErrNoSystemIntegration
	}

	ctx = systemOwnedQueryContext(ctx, entClient)

	integrations, err := entClient.Integration.Query().
		Where(
			integration.KindEQ(string(providerType)),
			integration.SystemOwnedEQ(true),
		).
		WithSecrets(func(q *ent.HushQuery) {
			q.Where(hush.SystemOwnedEQ(true))
		}).
		All(ctx)
	if err != nil || len(integrations) == 0 {
		return nil, fmt.Errorf("%w for provider %s", objects.ErrNoSystemIntegration, providerType)
	}

	var activeInteg *ent.Integration
	for _, integ := range integrations {
		if len(integ.Edges.Secrets) == 0 {
			continue
		}

		if activeInteg == nil {
			activeInteg = integ
			continue
		}

		current, ok := integ.Metadata["synchronized_at"].(string)
		if !ok {
			continue
		}

		best, ok := activeInteg.Metadata["synchronized_at"].(string)
		if !ok {
			activeInteg = integ
			continue
		}

		currentTime, errCurrent := time.Parse(time.RFC3339, current)
		bestTime, errBest := time.Parse(time.RFC3339, best)

		if errCurrent == nil && (errBest != nil || currentTime.After(bestTime)) {
			activeInteg = integ
		}
	}

	if activeInteg == nil {
		return nil, fmt.Errorf("%w for provider %s", objects.ErrNoIntegrationWithSecrets, providerType)
	}

	secret := activeInteg.Edges.Secrets[0]
	credentials := storage.ProviderCredentials{
		AccessKeyID:     secret.CredentialSet.AccessKeyID,
		SecretAccessKey: secret.CredentialSet.SecretAccessKey,
		Endpoint:        secret.CredentialSet.Endpoint,
		ProjectID:       secret.CredentialSet.ProjectID,
		AccountID:       secret.CredentialSet.AccountID,
		APIToken:        secret.CredentialSet.APIToken,
	}

	options := storage.NewProviderOptions(storage.WithCredentials(credentials))

	if activeInteg.Metadata != nil {
		for key, value := range activeInteg.Metadata {
			switch strings.ToLower(key) {
			case "bucket":
				if strVal, ok := stringValue(value); ok {
					options.Apply(storage.WithBucket(strVal))
				}
			case "region":
				if strVal, ok := stringValue(value); ok {
					options.Apply(storage.WithRegion(strVal))
				}
			case "endpoint":
				if strVal, ok := stringValue(value); ok {
					options.Apply(storage.WithEndpoint(strVal))
				}
			case "base_path":
				if strVal, ok := stringValue(value); ok {
					options.Apply(storage.WithBasePath(strVal))
				}
			case "local_url":
				if strVal, ok := stringValue(value); ok {
					options.Apply(storage.WithLocalURL(strVal))
				}
			default:
				options.Apply(storage.WithExtra(key, value))
			}
		}
	}

	return &cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions]{
		Type:        cp.ProviderType(providerType),
		Credentials: credentials,
		Config:      options,
	}, nil
}

func resolveProviderFromConfig(provider storage.ProviderType, config storage.ProviderConfig) (*cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions], error) {
	options, creds, err := providerOptionsFromConfig(provider, config)
	if err != nil {
		return nil, err
	}

	return &cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions]{
		Type:        cp.ProviderType(provider),
		Credentials: creds,
		Config:      options,
	}, nil
}

func providerOptionsFromConfig(provider storage.ProviderType, config storage.ProviderConfig) (*storage.ProviderOptions, storage.ProviderCredentials, error) {
	var providerCfg storage.ProviderConfigs

	switch provider {
	case storage.S3Provider:
		providerCfg = config.Providers.S3
	case storage.R2Provider:
		providerCfg = config.Providers.CloudflareR2
	case storage.GCSProvider:
		providerCfg = config.Providers.GCS
	case storage.DiskProvider:
		providerCfg = config.Providers.Disk
	default:
		return nil, storage.ProviderCredentials{}, fmt.Errorf("%w: %s", errUnsupportedProvider, provider)
	}

	if !providerCfg.Enabled {
		return nil, storage.ProviderCredentials{}, fmt.Errorf("%w: %s", errProviderDisabled, provider)
	}

	options := storage.NewProviderOptions(storage.WithCredentials(providerCfg.Credentials))

	switch provider {
	case storage.S3Provider:
		if providerCfg.Bucket != "" {
			options.Apply(storage.WithBucket(providerCfg.Bucket))
		}
		region := providerCfg.Region
		if region == "" {
			region = objects.DefaultS3Region
		}
		options.Apply(storage.WithRegion(region))
		if providerCfg.Endpoint != "" {
			options.Apply(storage.WithEndpoint(providerCfg.Endpoint))
		}
	case storage.R2Provider, storage.GCSProvider:
		if providerCfg.Bucket != "" {
			options.Apply(storage.WithBucket(providerCfg.Bucket))
		}
		if providerCfg.Endpoint != "" {
			options.Apply(storage.WithEndpoint(providerCfg.Endpoint))
		}
	case storage.DiskProvider:
		bucket := providerCfg.Bucket
		if bucket == "" {
			bucket = objects.DefaultDevStorageBucket
		}
		options.Apply(storage.WithBucket(bucket), storage.WithBasePath(bucket))
		if providerCfg.Endpoint != "" {
			options.Apply(storage.WithLocalURL(providerCfg.Endpoint))
		}
	}

	return options, providerCfg.Credentials, nil
}

func systemOwnedQueryContext(ctx context.Context, entClient *ent.Client) context.Context {
	user := &auth.AuthenticatedUser{
		SubjectID:          "system-storage-resolver",
		SubjectName:        "System Storage Resolver",
		AuthenticationType: auth.APITokenAuthentication,
		IsSystemAdmin:      true,
	}

	ctx = auth.WithAuthenticatedUser(ctx, user)
	ctx = auth.WithSystemAdminContext(ctx, user)
	ctx = contextx.With(ctx, auth.OrganizationCreationContextKey{})
	ctx = privacy.DecisionContext(ctx, privacy.Allow)
	return ent.NewContext(ctx, entClient)
}

func stringValue(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case fmt.Stringer:
		return v.String(), true
	case []byte:
		return string(v), true
	default:
		if v == nil {
			return "", false
		}
		return fmt.Sprintf("%v", v), true
	}
}

func cloneProviderOptions(in *storage.ProviderOptions) *storage.ProviderOptions {
	if in == nil {
		return nil
	}
	return in.Clone()
}
