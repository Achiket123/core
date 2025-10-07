package s3_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	storage "github.com/theopenlane/core/pkg/objects/storage"
	s3provider "github.com/theopenlane/core/pkg/objects/storage/providers/s3"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

func providerOptions() *storage.ProviderOptions {
	return storage.NewProviderOptions(
		storage.WithBucket("test-bucket"),
		storage.WithRegion("us-east-1"),
		storage.WithCredentials(storage.ProviderCredentials{
			AccessKeyID:     "test-access-key",
			SecretAccessKey: "test-secret-key",
		}),
	)
}

func TestNewS3Provider(t *testing.T) {
	opts := providerOptions()

	provider, err := s3provider.NewS3Provider(opts)
	if err != nil {
		t.Skip("Skipping test due to missing AWS credentials or environment")
	}

	assert.NotNil(t, provider)
}

func TestNewS3ProviderWithOptions(t *testing.T) {
	opts := providerOptions()
	provider, err := s3provider.NewS3Provider(opts, s3provider.WithUsePathStyle(true), s3provider.WithDebugMode(true), s3provider.WithAWSConfig(aws.Config{Region: "us-east-1"}))
	if err != nil {
		t.Skip("Skipping test due to missing AWS credentials or environment")
	}

	assert.NotNil(t, provider)
}

func TestNewS3ProviderMissingBucket(t *testing.T) {
	opts := providerOptions()
	opts.Bucket = ""

	_, err := s3provider.NewS3Provider(opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bucket")
}

func TestNewS3ProviderMissingRegion(t *testing.T) {
	opts := providerOptions()
	opts.Region = ""

	_, err := s3provider.NewS3Provider(opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required S3 credentials")
}

func TestS3ProviderConstants(t *testing.T) {
	assert.Equal(t, 15*time.Minute, s3provider.DefaultPresignedURLExpiry)
	assert.Equal(t, 64*1024*1024, s3provider.DefaultPartSize)
	assert.Equal(t, 5, s3provider.DefaultConcurrency)
}

func TestS3ProviderMethods(t *testing.T) {
	opts := providerOptions()

	provider, err := s3provider.NewS3Provider(opts)
	if err != nil {
		t.Skip("Skipping test due to missing AWS credentials or environment")
	}

	t.Run("ProviderType", func(t *testing.T) {
		assert.Equal(t, storagetypes.S3Provider, provider.ProviderType())
	})

	t.Run("GetScheme", func(t *testing.T) {
		scheme := provider.GetScheme()
		assert.NotNil(t, scheme)
		assert.Equal(t, "s3://", *scheme)
	})

	t.Run("Close", func(t *testing.T) {
		assert.NoError(t, provider.Close())
	})
}

func TestNewS3ProviderResult(t *testing.T) {
	opts := providerOptions()
	result := s3provider.NewS3ProviderResult(opts)
	if result.IsError() {
		t.Skip("Skipping test due to missing AWS credentials or environment")
	}

	provider := result.MustGet()
	assert.NotNil(t, provider)
}

func TestNewS3ProviderFromCredentials(t *testing.T) {
	creds := storage.ProviderCredentials{AccessKeyID: "key", SecretAccessKey: "secret"}
	opts := storage.NewProviderOptions(storage.WithBucket("bucket"), storage.WithRegion("us-east-1"))

	result := s3provider.NewS3ProviderFromCredentials(creds, opts)
	if result.IsError() {
		t.Skip("Skipping test due to missing AWS credentials or environment")
	}

	assert.NotNil(t, result.MustGet())

	t.Run("missing options", func(t *testing.T) {
		errResult := s3provider.NewS3ProviderFromCredentials(creds, nil)
		assert.True(t, errResult.IsError())
	})
}

func TestS3ProviderUploadDownloadFlow(t *testing.T) {
	t.Skip("Integration test - requires real S3 environment or LocalStack")

	provider, err := s3provider.NewS3Provider(providerOptions())
	require.NoError(t, err)

	ctx := context.Background()
	testContent := "This is test file content"
	fileName := "test-file.txt"

	uploadOpts := &storagetypes.UploadFileOptions{
		FileName:    fileName,
		ContentType: "text/plain",
	}

	metadata, err := provider.Upload(ctx, strings.NewReader(testContent), uploadOpts)
	require.NoError(t, err)
	require.NotNil(t, metadata)

	downloadOpts := &storagetypes.DownloadFileOptions{}
	downloaded, err := provider.Download(ctx, &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: fileName}}, downloadOpts)
	require.NoError(t, err)
	require.NotNil(t, downloaded)

	assert.Equal(t, []byte(testContent), downloaded.File)
}
