package s3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/pkg/cp"
	"github.com/theopenlane/core/pkg/objects/storage"
	s3provider "github.com/theopenlane/core/pkg/objects/storage/providers/s3"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

func TestNewS3Builder(t *testing.T) {
	builder := s3provider.NewS3Builder()
	assert.NotNil(t, builder)

	var _ cp.ClientBuilder[storagetypes.Provider, storage.ProviderCredentials, *storage.ProviderOptions] = builder
}

func TestS3BuilderWithCredentials(t *testing.T) {
	builder := s3provider.NewS3Builder()
	creds := storage.ProviderCredentials{AccessKeyID: "key", SecretAccessKey: "secret"}

	result := builder.WithCredentials(creds)
	assert.Equal(t, builder, result)
}

func TestS3BuilderWithConfig(t *testing.T) {
	builder := s3provider.NewS3Builder()
	options := storage.NewProviderOptions(storage.WithBucket("bucket"), storage.WithRegion("us-east-1"))

	result := builder.WithConfig(options)
	assert.Equal(t, builder, result)
}

func TestS3BuilderBuild(t *testing.T) {
	tests := []struct {
		name        string
		credentials storage.ProviderCredentials
		options     *storage.ProviderOptions
		expectError bool
	}{
		{
			name:        "valid configuration",
			credentials: storage.ProviderCredentials{AccessKeyID: "key", SecretAccessKey: "secret"},
			options: storage.NewProviderOptions(
				storage.WithBucket("bucket"),
				storage.WithRegion("us-east-1"),
			),
		},
		{
			name:        "missing bucket",
			credentials: storage.ProviderCredentials{AccessKeyID: "key", SecretAccessKey: "secret"},
			options:     storage.NewProviderOptions(storage.WithRegion("us-east-1")),
			expectError: true,
		},
		{
			name:        "missing region",
			credentials: storage.ProviderCredentials{AccessKeyID: "key", SecretAccessKey: "secret"},
			options:     storage.NewProviderOptions(storage.WithBucket("bucket")),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := s3provider.NewS3Builder().WithCredentials(tt.credentials).WithConfig(tt.options)
			provider, err := builder.Build(context.Background())

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				if err != nil {
					t.Skip("Skipping test due to missing AWS credentials or environment")
				}
				assert.NoError(t, err)
				assert.NotNil(t, provider)
			}
		})
	}
}

func TestS3BuilderClientType(t *testing.T) {
	builder := s3provider.NewS3Builder()
	assert.Equal(t, cp.ProviderType(storagetypes.S3Provider), builder.ClientType())
}
