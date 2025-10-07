package r2_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/pkg/cp"
	"github.com/theopenlane/core/pkg/objects/storage"
	r2provider "github.com/theopenlane/core/pkg/objects/storage/providers/r2"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

func TestNewR2Builder(t *testing.T) {
	builder := r2provider.NewR2Builder()
	assert.NotNil(t, builder)

	var _ cp.ClientBuilder[storagetypes.Provider, storage.ProviderCredentials, *storage.ProviderOptions] = builder
}

func TestR2BuilderWithCredentials(t *testing.T) {
	builder := r2provider.NewR2Builder()
	creds := storage.ProviderCredentials{AccountID: "account", AccessKeyID: "access", SecretAccessKey: "secret"}

	result := builder.WithCredentials(creds)
	assert.Equal(t, builder, result)
}

func TestR2BuilderWithConfig(t *testing.T) {
	builder := r2provider.NewR2Builder()
	options := r2Options()

	result := builder.WithConfig(options)
	assert.Equal(t, builder, result)
}

func TestR2BuilderBuild(t *testing.T) {
	tests := []struct {
		name        string
		credentials storage.ProviderCredentials
		options     *storage.ProviderOptions
		expectError bool
	}{
		{
			name:        "valid configuration",
			credentials: storage.ProviderCredentials{AccountID: "account", AccessKeyID: "access", SecretAccessKey: "secret"},
			options: storage.NewProviderOptions(
				storage.WithBucket("bucket"),
				storage.WithEndpoint("https://account.r2.cloudflarestorage.com"),
			),
		},
		{
			name:        "missing bucket",
			credentials: storage.ProviderCredentials{AccountID: "account", AccessKeyID: "access", SecretAccessKey: "secret"},
			options:     storage.NewProviderOptions(),
			expectError: true,
		},
		{
			name:        "missing account ID",
			credentials: storage.ProviderCredentials{AccessKeyID: "access", SecretAccessKey: "secret"},
			options: storage.NewProviderOptions(
				storage.WithBucket("bucket"),
			),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := r2provider.NewR2Builder().WithCredentials(tt.credentials).WithConfig(tt.options)
			provider, err := builder.Build(context.Background())

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				if err != nil {
					t.Skip("Skipping test due to missing R2-compatible environment")
				}
				assert.NoError(t, err)
				assert.NotNil(t, provider)
			}
		})
	}
}

func TestR2BuilderClientType(t *testing.T) {
	builder := r2provider.NewR2Builder()
	assert.Equal(t, cp.ProviderType(storagetypes.R2Provider), builder.ClientType())
}
