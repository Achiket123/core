package resolver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/pkg/eddy"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/utils/contextx"
)

type stubBuilder struct {
	providerType string
	lastConfig   *storage.ProviderOptions
	lastOutput   storage.ProviderCredentials
}

func (b *stubBuilder) Build(_ context.Context, output storage.ProviderCredentials, cfg *storage.ProviderOptions) (storage.Provider, error) {
	b.lastOutput = output
	if cfg != nil {
		copy := cfg.Clone()
		b.lastConfig = copy
	} else {
		b.lastConfig = nil
	}
	return nil, nil
}

func (b *stubBuilder) ProviderType() string {
	return b.providerType
}

func TestConfigureProviderRulesDevMode(t *testing.T) {
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()

	diskBuilder := &stubBuilder{providerType: "disk"}
	config := storage.ProviderConfig{
		DevMode: true,
		Providers: storage.Providers{
			Disk: storage.ProviderConfigs{Enabled: true},
		},
	}

	configureProviderRules(
		resolver,
		config,
		&stubBuilder{providerType: "s3"},
		&stubBuilder{providerType: "r2"},
		diskBuilder,
		&stubBuilder{providerType: "db"},
	)

	option := resolver.Resolve(context.Background())
	require.True(t, option.IsPresent(), "expected dev mode rule to resolve")

	result := option.MustGet()
	require.Equal(t, diskBuilder, result.Builder, "expected disk builder for dev mode")
	require.NotNil(t, result.Config)
	require.Equal(t, objects.DefaultDevStorageBucket, result.Config.Bucket)
	require.Equal(t, objects.DefaultDevStorageBucket, result.Config.BasePath)
	extra, ok := result.Config.Extra("dev_mode")
	require.True(t, ok)
	require.Equal(t, true, extra)
}

func TestKnownProviderRule(t *testing.T) {
	ctx := contextx.With(context.Background(), objects.KnownProviderHint(storage.DiskProvider))
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()

	diskBuilder := &stubBuilder{providerType: "disk"}
	config := storage.ProviderConfig{
		Providers: storage.Providers{
			Disk: storage.ProviderConfigs{
				Enabled:  true,
				Bucket:   "/mnt/storage",
				Endpoint: "http://local",
			},
		},
	}

	configureProviderRules(
		resolver,
		config,
		&stubBuilder{providerType: "s3"},
		&stubBuilder{providerType: "r2"},
		diskBuilder,
		&stubBuilder{providerType: "db"},
	)

	option := resolver.Resolve(ctx)
	require.True(t, option.IsPresent(), "expected known provider rule to resolve")

	result := option.MustGet()
	require.Equal(t, diskBuilder, result.Builder)
	require.Equal(t, "/mnt/storage", result.Config.Bucket)
	require.Equal(t, "/mnt/storage", result.Config.BasePath)
	require.Equal(t, "http://local", result.Config.LocalURL)
}

func TestModuleRules(t *testing.T) {
	ctx := contextx.With(context.Background(), objects.ModuleHint(models.CatalogTrustCenterModule))
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()

	r2Builder := &stubBuilder{providerType: "r2"}
	config := storage.ProviderConfig{
		Providers: storage.Providers{
			CloudflareR2: storage.ProviderConfigs{
				Enabled: true,
				Bucket:  "tc-bucket",
			},
		},
	}

	configureProviderRules(
		resolver,
		config,
		&stubBuilder{providerType: "s3"},
		r2Builder,
		&stubBuilder{providerType: "disk"},
		&stubBuilder{providerType: "db"},
	)

	option := resolver.Resolve(ctx)
	require.True(t, option.IsPresent(), "expected module rule to resolve")

	result := option.MustGet()
	require.Equal(t, r2Builder, result.Builder)
	require.Equal(t, "tc-bucket", result.Config.Bucket)
}

func TestPreferredProviderRule(t *testing.T) {
	ctx := contextx.With(context.Background(), objects.PreferredProviderHint(storage.S3Provider))
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()

	s3Builder := &stubBuilder{providerType: "s3"}
	config := storage.ProviderConfig{
		Providers: storage.Providers{
			S3: storage.ProviderConfigs{
				Enabled: true,
				Bucket:  "preferred-bucket",
				Region:  "us-west-1",
			},
		},
	}

	configureProviderRules(
		resolver,
		config,
		s3Builder,
		&stubBuilder{providerType: "r2"},
		&stubBuilder{providerType: "disk"},
		&stubBuilder{providerType: "db"},
	)

	option := resolver.Resolve(ctx)
	require.True(t, option.IsPresent(), "expected preferred provider rule to resolve")

	result := option.MustGet()
	require.Equal(t, s3Builder, result.Builder)
	require.Equal(t, "preferred-bucket", result.Config.Bucket)
	require.Equal(t, "us-west-1", result.Config.Region)
}

func TestFallbackRuleSelectsFirstEnabledProvider(t *testing.T) {
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()

	s3Builder := &stubBuilder{providerType: "s3"}
	r2Builder := &stubBuilder{providerType: "r2"}

	config := storage.ProviderConfig{
		Providers: storage.Providers{
			S3: storage.ProviderConfigs{
				Enabled: false,
			},
			CloudflareR2: storage.ProviderConfigs{
				Enabled: true,
				Bucket:  "r2-bucket",
			},
		},
	}

	configureProviderRules(
		resolver,
		config,
		s3Builder,
		r2Builder,
		&stubBuilder{providerType: "disk"},
		&stubBuilder{providerType: "db"},
	)

	option := resolver.Resolve(context.Background())
	require.True(t, option.IsPresent(), "expected fallback rule to resolve")

	result := option.MustGet()
	require.Equal(t, r2Builder, result.Builder, "expected first enabled provider to be used")
	require.Equal(t, "r2-bucket", result.Config.Bucket)
}

func TestProviderEnabledChecksConfig(t *testing.T) {
	rc := &ruleCoordinator{
		config: storage.ProviderConfig{
			Providers: storage.Providers{
				S3:   storage.ProviderConfigs{Enabled: true},
				Disk: storage.ProviderConfigs{Enabled: false},
			},
		},
	}

	require.True(t, rc.providerEnabled(storage.S3Provider))
	require.False(t, rc.providerEnabled(storage.DiskProvider))
}

func TestResolveProviderWithUnsupportedBuilder(t *testing.T) {
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	rc := newRuleCoordinator(
		resolver,
		storage.ProviderConfig{},
		nil,
		nil,
		nil,
		nil,
	)

	_, err := rc.resolveProviderWithBuilder(context.Background(), storage.ProviderType("unsupported"))
	require.Error(t, err)
	require.ErrorIs(t, err, errUnsupportedProvider)
}

func TestResolveProviderFromConfigCopiesOptions(t *testing.T) {
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	rc := newRuleCoordinator(
		resolver,
		storage.ProviderConfig{
			Providers: storage.Providers{
				S3: storage.ProviderConfigs{
					Enabled: true,
					Bucket:  "bucket",
					Region:  "us-east-1",
				},
			},
		},
		&stubBuilder{providerType: "s3"},
		nil,
		nil,
		nil,
	)

	resolved, err := rc.resolveProvider(context.Background(), storage.S3Provider)
	require.NoError(t, err)
	require.Equal(t, "bucket", resolved.Config.Bucket)
	require.Equal(t, "us-east-1", resolved.Config.Region)
}

func TestProviderResolveFromConfigDisabled(t *testing.T) {
	_, err := resolveProviderFromConfig(storage.S3Provider, storage.ProviderConfig{
		Providers: storage.Providers{
			S3: storage.ProviderConfigs{Enabled: false},
		},
	})
	require.Error(t, err)
}

func TestHandleDevModeOptionClone(t *testing.T) {
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	diskBuilder := &stubBuilder{providerType: "disk"}
	rc := newRuleCoordinator(
		resolver,
		storage.ProviderConfig{
			DevMode: true,
			Providers: storage.Providers{
				Disk: storage.ProviderConfigs{Enabled: true},
			},
		},
		nil, nil,
		diskBuilder,
		nil,
	)

	require.True(t, rc.handleDevMode())

	option := resolver.Resolve(context.Background())
	require.True(t, option.IsPresent())

	result := option.MustGet()
	require.NotNil(t, result.Config)
	// ensure options cloned on each invocation
	result.Config.Apply(storage.WithExtra("mutated", true))

	option = resolver.Resolve(context.Background())
	require.True(t, option.IsPresent())
	_, ok := option.MustGet().Config.Extra("mutated")
	require.False(t, ok)
}

func TestAddFallbackRuleSkipsDisabledProviders(t *testing.T) {
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	rc := newRuleCoordinator(
		resolver,
		storage.ProviderConfig{
			Providers: storage.Providers{
				S3:           storage.ProviderConfigs{Enabled: false},
				CloudflareR2: storage.ProviderConfigs{Enabled: true},
			},
		},
		&stubBuilder{providerType: "s3"},
		&stubBuilder{providerType: "r2"},
		nil,
		nil,
	)

	rc.addFallbackRule([]storage.ProviderType{storage.S3Provider, storage.R2Provider})

	option := resolver.Resolve(context.Background())
	require.True(t, option.IsPresent())
	require.Equal(t, "r2", option.MustGet().Builder.ProviderType())
}
