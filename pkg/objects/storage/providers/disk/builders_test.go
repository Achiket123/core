package disk_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/pkg/cp"
	"github.com/theopenlane/core/pkg/objects/storage"
	diskprovider "github.com/theopenlane/core/pkg/objects/storage/providers/disk"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

func TestNewDiskBuilder(t *testing.T) {
	builder := diskprovider.NewDiskBuilder()
	assert.NotNil(t, builder)

	var _ cp.ClientBuilder[storagetypes.Provider, storage.ProviderCredentials, *storage.ProviderOptions] = builder
}

func TestDiskBuilderWithCredentials(t *testing.T) {
	builder := diskprovider.NewDiskBuilder()
	creds := storage.ProviderCredentials{Endpoint: "http://localhost:8080/files"}

	result := builder.WithCredentials(creds)
	assert.Equal(t, builder, result)
}

func TestDiskBuilderWithConfig(t *testing.T) {
	builder := diskprovider.NewDiskBuilder()
	options := storage.NewProviderOptions(storage.WithBucket("/tmp/test-storage"))

	result := builder.WithConfig(options)
	assert.Equal(t, builder, result)
}

func TestDiskBuilderBuild(t *testing.T) {
	tests := []struct {
		name        string
		credentials storage.ProviderCredentials
		options     *storage.ProviderOptions
		expectError bool
	}{
		{
			name:        "valid configuration",
			credentials: storage.ProviderCredentials{Endpoint: "http://localhost:8080/files"},
			options:     storage.NewProviderOptions(storage.WithBucket("/tmp/test-storage")),
		},
		{
			name:        "missing bucket uses default",
			credentials: storage.ProviderCredentials{},
			options:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := diskprovider.NewDiskBuilder().WithCredentials(tt.credentials).WithConfig(tt.options)
			provider, err := builder.Build(context.Background())

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
			}
		})
	}
}

func TestDiskBuilderClientType(t *testing.T) {
	builder := diskprovider.NewDiskBuilder()
	assert.Equal(t, cp.ProviderType(storagetypes.DiskProvider), builder.ClientType())
}
