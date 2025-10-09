package resolver

import (
	"context"
	"fmt"
	"strings"
	"time"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hush"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
)

// providerEnabled returns whether a provider can be used based on configuration.
func (rc *ruleCoordinator) providerEnabled(provider storage.ProviderType) bool {
	switch provider {
	case storage.R2Provider:
		return rc.config.Providers.CloudflareR2.Enabled
	case storage.S3Provider:
		return rc.config.Providers.S3.Enabled
	case storage.DiskProvider:
		return rc.config.Providers.Disk.Enabled
	case storage.DatabaseProvider:
		return rc.config.Providers.Database.Enabled
	default:
		return false
	}
}

// providerResolution is an internal type for credential resolution before adding builder
type providerResolution struct {
	Output storage.ProviderCredentials
	Config *storage.ProviderOptions
}

// resolveProvider returns provider credentials from system integrations or config fallback
func (rc *ruleCoordinator) resolveProvider(ctx context.Context, provider storage.ProviderType) (*providerResolution, error) {
	if rc.config.CredentialSync.Enabled {
		if resolved, err := querySystemProvider(ctx, provider); err == nil {
			return resolved, nil
		}
	}

	return resolveProviderFromConfig(provider, rc.config)
}

func querySystemProvider(ctx context.Context, providerType storage.ProviderType) (*providerResolution, error) {
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

	return &providerResolution{
		Output: credentials,
		Config: options,
	}, nil
}

func resolveProviderFromConfig(provider storage.ProviderType, config storage.ProviderConfig) (*providerResolution, error) {
	options, creds, err := providerOptionsFromConfig(provider, config)
	if err != nil {
		return nil, err
	}

	return &providerResolution{
		Output: creds,
		Config: options,
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
	case storage.DatabaseProvider:
		providerCfg = config.Providers.Database
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
	case storage.DatabaseProvider:
		if providerCfg.Bucket != "" {
			options.Apply(storage.WithBucket(providerCfg.Bucket))
		}
		if providerCfg.Endpoint != "" {
			options.Apply(storage.WithEndpoint(providerCfg.Endpoint))
		}
	}

	return options, providerCfg.Credentials, nil
}
