package objects

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/pkg/cp"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/objects/storage"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
)

// mockProvider implements storagetypes.Provider for testing
type mockProvider struct {
	name string
}

// ProviderType implements storagetypes.Provider.
func (m *mockProvider) ProviderType() storagetypes.ProviderType {
	return storagetypes.ProviderType(m.name)
}

func (m *mockProvider) Upload(ctx context.Context, reader io.Reader, opts *storagetypes.UploadFileOptions) (*storagetypes.UploadedFileMetadata, error) {
	return &storagetypes.UploadedFileMetadata{
		FileMetadata: storagetypes.FileMetadata{
			Key:          "test-file-key",
			Size:         100,
			ContentType:  opts.ContentType,
			ProviderType: storagetypes.ProviderType(m.name),
			ProviderHints: &storagetypes.ProviderHints{
				IntegrationID:  "hint-integration",
				HushID:         "hint-hush",
				OrganizationID: "hint-org",
			},
		},
	}, nil
}

func (m *mockProvider) Download(ctx context.Context, file *storagetypes.File, opts *storagetypes.DownloadFileOptions) (*storagetypes.DownloadedFileMetadata, error) {
	// Return a fixed size download to match test expectations
	content := make([]byte, 100) // 100 bytes to match test
	for i := range content {
		content[i] = byte('A' + (i % 26)) // Fill with letters
	}
	return &storagetypes.DownloadedFileMetadata{
		File: content,
		Size: 100,
	}, nil
}

func (m *mockProvider) Delete(ctx context.Context, file *storagetypes.File, opts *storagetypes.DeleteFileOptions) error {
	return nil
}

func (m *mockProvider) GetPresignedURL(ctx context.Context, file *storagetypes.File, opts *storagetypes.PresignedURLOptions) (string, error) {
	key := ""
	expires := time.Hour
	if file != nil {
		key = file.ID
	}
	if opts != nil && opts.Duration > 0 {
		expires = opts.Duration
	}
	return fmt.Sprintf("https://%s.example.com/presigned/%s?expires=%d", m.name, key, int64(expires.Seconds())), nil
}

func (m *mockProvider) Exists(ctx context.Context, file *storagetypes.File) (bool, error) {
	return true, nil
}

func (m *mockProvider) GetScheme() *string {
	scheme := fmt.Sprintf("%s://", m.name)
	return &scheme
}

func (m *mockProvider) ListBuckets() ([]string, error) {
	// Return a mock list of buckets for testing
	return []string{"bucket1", "bucket2"}, nil
}

func (m *mockProvider) Close() error {
	return nil
}

// mockBuilder implements cp.ClientBuilder for testing
type mockBuilder struct {
	provider *mockProvider
}

func (m *mockBuilder) WithCredentials(storage.ProviderCredentials) cp.ClientBuilder[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions] {
	return m
}

func (m *mockBuilder) WithConfig(*storage.ProviderOptions) cp.ClientBuilder[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions] {
	return m
}

func (m *mockBuilder) Build(ctx context.Context) (storage.Provider, error) {
	return m.provider, nil
}

func (m *mockBuilder) ClientType() cp.ProviderType {
	return cp.ProviderType(m.provider.name)
}

func cloneProviderOptions(in *storage.ProviderOptions) *storage.ProviderOptions {
	if in == nil {
		return nil
	}
	return in.Clone()
}

// createTestService creates a test service with typed context resolution rules
func createTestService() *Service {
	// Create resolver with typed context rules
	resolver := cp.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()

	// Trust Center Module → R2 Provider
	trustCenterRule := cp.NewRule[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]().
		WhenFunc(func(ctx context.Context) bool {
			return cp.GetValueEquals(ctx, models.CatalogTrustCenterModule)
		}).
		Resolve(func(context.Context) (*cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions], error) {
			return &cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions]{
				Type:        "r2",
				Credentials: storage.ProviderCredentials{},
				Config:      storage.NewProviderOptions(storage.WithBucket("trust-center"), storage.WithRegion("us-east-1")),
			}, nil
		})

	// Compliance Module → S3 Provider
	complianceRule := cp.NewRule[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]().
		WhenFunc(func(ctx context.Context) bool {
			return cp.GetValueEquals(ctx, models.CatalogComplianceModule)
		}).
		Resolve(func(context.Context) (*cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions], error) {
			return &cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions]{
				Type:        "s3",
				Credentials: storage.ProviderCredentials{},
				Config:      storage.NewProviderOptions(storage.WithBucket("compliance"), storage.WithRegion("us-west-2")),
			}, nil
		})

	// Default rule for fallback
	defaultRule := cp.DefaultRule[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions](cp.Resolution[storage.ProviderCredentials, *storage.ProviderOptions]{
		ClientType:  "disk",
		Credentials: storage.ProviderCredentials{},
		Config:      storage.NewProviderOptions(storage.WithBasePath("/tmp"), storage.WithBucket("/tmp")),
	})

	resolver.AddRule(trustCenterRule)
	resolver.AddRule(complianceRule)
	resolver.AddRule(defaultRule)

	// Create client service with mock builders
	pool := cp.NewClientPool[storage.Provider](5 * time.Minute)
	clientService := cp.NewClientService[
		storage.Provider,
		storage.ProviderCredentials,
		*storage.ProviderOptions,
	](pool, cp.WithConfigClone[
		storage.Provider,
		storage.ProviderCredentials,
		*storage.ProviderOptions,
	](cloneProviderOptions))

	// Register mock builders
	clientService.RegisterBuilder("r2", &mockBuilder{provider: &mockProvider{name: "r2"}})
	clientService.RegisterBuilder("s3", &mockBuilder{provider: &mockProvider{name: "s3"}})
	clientService.RegisterBuilder("disk", &mockBuilder{provider: &mockProvider{name: "disk"}})

	return NewService(Config{
		Resolver:      resolver,
		ClientService: clientService,
	})
}

func TestService_Upload_TypedContextResolution(t *testing.T) {
	service := createTestService()

	tests := []struct {
		name             string
		setupContext     func() context.Context
		expectedProvider string
	}{
		{
			name: "TrustCenter module resolves to R2",
			setupContext: func() context.Context {
				// Create authenticated user context with organization ID
				ctx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())
				// Add module information for provider resolution
				ctx = cp.WithValue(ctx, models.CatalogTrustCenterModule)
				return ctx
			},
			expectedProvider: "r2",
		},
		{
			name: "Compliance module resolves to S3",
			setupContext: func() context.Context {
				ctx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())
				ctx = cp.WithValue(ctx, models.CatalogComplianceModule)
				return ctx
			},
			expectedProvider: "s3",
		},
		{
			name: "Unknown module falls back to disk",
			setupContext: func() context.Context {
				ctx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())
				ctx = cp.WithValue(ctx, "unknown-module")
				return ctx
			},
			expectedProvider: "disk",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			reader := strings.NewReader("test content")
			opts := &storage.UploadOptions{
				FileName:    "test-file.txt",
				ContentType: "text/plain",
				FileMetadata: storagetypes.FileMetadata{
					ProviderHints: &storagetypes.ProviderHints{},
				},
			}

			file, err := service.Upload(ctx, reader, opts)
			require.NoError(t, err)
			assert.Equal(t, "test-file.txt", file.OriginalName)
			assert.Equal(t, tt.expectedProvider, string(file.ProviderType))
		})
	}
}

func TestService_Download_TypedContextResolution(t *testing.T) {
	service := createTestService()

	// Create a test file with metadata
	testFile := &storagetypes.File{
		ID:           "test-file-id",
		OriginalName: "test-file.txt",
		FileMetadata: storagetypes.FileMetadata{
			Key:         "test-file-id", // Use same value as ID for testing
			Size:        100,
			ContentType: "text/plain",
		},
	}

	ctx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())

	downloadOpts := &storage.DownloadOptions{}
	result, err := service.Download(ctx, nil, testFile, downloadOpts)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify the download result
	assert.Equal(t, int64(100), result.Size)
	assert.NotNil(t, result.File)
}

func TestService_GetPresignedURL_TypedContextResolution(t *testing.T) {
	service := createTestService()

	testFile := &storagetypes.File{
		ID:           "test-file-id",
		OriginalName: "test-file.txt",
		FileMetadata: storagetypes.FileMetadata{
			Key:         "test-file-id", // Use same value as ID for testing
			Size:        100,
			ContentType: "text/plain",
			ProviderHints: &storagetypes.ProviderHints{
				IntegrationID:  "test-integration",
				HushID:         "test-hush",
				OrganizationID: "test-org",
			},
		},
	}

	ctx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())

	url, err := service.GetPresignedURL(ctx, testFile, time.Hour)
	require.NoError(t, err)
	assert.Contains(t, url, "presigned")
	assert.Contains(t, url, testFile.ID)
}

func TestService_Delete_TypedContextResolution(t *testing.T) {
	service := createTestService()

	testFile := &storagetypes.File{
		ID:           "test-file-id",
		OriginalName: "test-file.txt",
		FileMetadata: storagetypes.FileMetadata{
			Key:         "test-file-id", // Use same value as ID for testing
			Size:        100,
			ContentType: "text/plain",
			ProviderHints: &storagetypes.ProviderHints{
				IntegrationID:  "test-integration",
				HushID:         "test-hush",
				OrganizationID: "test-org",
			},
		},
	}

	ctx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())

	err := service.Delete(ctx, testFile, &storagetypes.DeleteFileOptions{})
	require.NoError(t, err)
}

func TestService_Exists_TypedContextResolution(t *testing.T) {
	service := createTestService()

	testFile := &storagetypes.File{
		ID:           "test-file-id",
		OriginalName: "test-file.txt",
		FileMetadata: storagetypes.FileMetadata{
			Key:         "test-file-id", // Use same value as ID for testing
			Size:        100,
			ContentType: "text/plain",
			ProviderHints: &storagetypes.ProviderHints{
				IntegrationID:  "test-integration",
				HushID:         "test-hush",
				OrganizationID: "test-org",
			},
		},
	}

	ctx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())

	exists, err := service.Exists(ctx, testFile)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestService_BuildResolutionContext(t *testing.T) {
	service := createTestService()

	ctx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())

	opts := &storage.UploadOptions{
		FileName:    "test-file.txt",
		ContentType: "text/plain",
		FileMetadata: storagetypes.FileMetadata{
			ProviderHints: &storagetypes.ProviderHints{
				IntegrationID:  "hint-integration",
				HushID:         "hint-hush",
				OrganizationID: "hint-org",
				Module:         models.CatalogComplianceModule,
				Metadata: map[string]string{
					"size_bytes": "256",
				},
			},
		},
	}

	enrichedCtx := service.buildResolutionContext(ctx, opts)

	// Verify context contains expected values
	uploadOpts := cp.GetValue[*storage.UploadOptions](enrichedCtx)
	assert.True(t, uploadOpts.IsPresent())
	assert.Equal(t, opts, uploadOpts.MustGet())

	hints := cp.GetValue[*storagetypes.ProviderHints](enrichedCtx)
	assert.True(t, hints.IsPresent())
	assert.Equal(t, "hint-integration", hints.MustGet().IntegrationID)

	moduleHint := cp.GetHint(enrichedCtx, ModuleHintKey())
	assert.True(t, moduleHint.IsPresent())
	assert.Equal(t, models.CatalogComplianceModule, moduleHint.MustGet())
}

func TestService_BuildResolutionContextForFile(t *testing.T) {
	service := createTestService()

	ctx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())

	file := &storagetypes.File{
		ID:           "test-file-id",
		OriginalName: "test-file.txt",
		FileMetadata: storagetypes.FileMetadata{
			ProviderHints: &storagetypes.ProviderHints{
				IntegrationID:  "file-integration",
				HushID:         "file-hush",
				OrganizationID: "file-org",
				Module:         models.CatalogTrustCenterModule,
			},
		},
	}

	enrichedCtx := service.buildResolutionContextForFile(ctx, file)

	// Verify context contains expected values
	contextFile := cp.GetValue[*storagetypes.File](enrichedCtx)
	assert.True(t, contextFile.IsPresent())
	assert.Equal(t, file, contextFile.MustGet())

	moduleHint := cp.GetHint(enrichedCtx, ModuleHintKey())
	assert.True(t, moduleHint.IsPresent())
	assert.Equal(t, models.CatalogTrustCenterModule, moduleHint.MustGet())
}

func TestService_ResolveProvider_ErrorHandling(t *testing.T) {
	// Create resolver with no rules (should fail resolution)
	resolver := cp.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	pool := cp.NewClientPool[storage.Provider](5 * time.Minute)
	clientService := cp.NewClientService[
		storage.Provider,
		storage.ProviderCredentials,
		*storage.ProviderOptions,
	](pool, cp.WithConfigClone[
		storage.Provider,
		storage.ProviderCredentials,
		*storage.ProviderOptions,
	](cloneProviderOptions))
	service := NewService(Config{
		Resolver:      resolver,
		ClientService: clientService,
	})

	ctx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())

	opts := &storage.UploadOptions{
		FileName:    "test-file.txt",
		ContentType: "text/plain",
		FileMetadata: storagetypes.FileMetadata{
			ProviderHints: &storagetypes.ProviderHints{},
		},
	}

	reader := strings.NewReader("test content")
	_, err := service.Upload(ctx, reader, opts)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrProviderResolutionFailed)
}

func TestService_ResolveProviderForFile_ErrorHandling(t *testing.T) {
	// Create resolver with no rules
	resolver := cp.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	pool := cp.NewClientPool[storage.Provider](5 * time.Minute)
	clientService := cp.NewClientService[
		storage.Provider,
		storage.ProviderCredentials,
		*storage.ProviderOptions,
	](pool, cp.WithConfigClone[
		storage.Provider,
		storage.ProviderCredentials,
		*storage.ProviderOptions,
	](cloneProviderOptions))
	service := NewService(Config{
		Resolver:      resolver,
		ClientService: clientService,
	})

	ctx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())

	testFile := &storagetypes.File{
		ID: "test-file-id",
		FileMetadata: storagetypes.FileMetadata{
			Key: "test-file-id", // Use same value as ID for testing
			ProviderHints: &storagetypes.ProviderHints{
				IntegrationID:  "test-integration",
				HushID:         "test-hush",
				OrganizationID: "test-org",
			},
		},
	}

	_, err := service.Download(ctx, nil, testFile, &storage.DownloadOptions{})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrProviderResolutionFailed)
}

func TestAuthContext_Verification(t *testing.T) {
	// Test that auth context functions work as expected - use generated ULIDs
	testUserID := ulids.New().String()
	testOrgID := ulids.New().String()

	ctx := auth.NewTestContextWithOrgID(testUserID, testOrgID)

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	require.NoError(t, err, "Should be able to get organization ID from test context")
	assert.Equal(t, testOrgID, orgID)

	subjectID, err := auth.GetSubjectIDFromContext(ctx)
	require.NoError(t, err, "Should be able to get subject ID from test context")
	assert.Equal(t, testUserID, subjectID)
}
