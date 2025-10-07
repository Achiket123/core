//go:build examples
// +build examples

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/theopenlane/core/pkg/cp"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/core/pkg/objects/storage/providers/disk"
	"github.com/theopenlane/core/pkg/objects/storage/providers/s3"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

// StorageClient wraps a provider for use with cp
type StorageClient struct {
	Provider storagetypes.Provider
	Type     string
}

type contextKey string

const (
	providerTypeKey contextKey = "provider-type"
	tenantIDKey     contextKey = "tenant-id"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== Multi-Provider Object Storage Example ===")

	pool := cp.NewClientPool[*StorageClient](10 * time.Minute)
	service := cp.NewClientService[
		*StorageClient,
		storage.ProviderCredentials,
		*storage.ProviderOptions,
	](pool, cp.WithConfigClone[
		*StorageClient,
		storage.ProviderCredentials,
		*storage.ProviderOptions,
	](cloneProviderOptions))

	registerProviderBuilders(service)
	resolver := createResolver()

	fmt.Println("1. Testing Disk Provider...")
	testProvider(ctx, service, resolver, "disk", "tenant1", "provider1")

	fmt.Println("\n2. Testing S3 Provider (MinIO - Provider 1)...")
	testProvider(ctx, service, resolver, "s3-provider1", "tenant1", "provider1")

	fmt.Println("\n3. Testing S3 Provider (MinIO - Provider 2)...")
	testProvider(ctx, service, resolver, "s3-provider2", "tenant2", "provider2")

	fmt.Println("\n4. Testing S3 Provider (MinIO - Provider 3)...")
	testProvider(ctx, service, resolver, "s3-provider3", "tenant3", "provider3")

	fmt.Println("\n5. Concurrent Operations Across All Providers...")
	testConcurrent(ctx, service, resolver)

	fmt.Println("\n6. Pool Statistics...")
	printPoolStats(pool)

	fmt.Println("\n=== Example completed successfully ===")
}

func registerProviderBuilders(service *cp.ClientService[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions]) {
	service.RegisterBuilder(cp.ProviderType("disk"), &diskBuilder{clientType: cp.ProviderType("disk")})

	s3Configs := []struct {
		name      string
		accessKey string
		secretKey string
		bucket    string
		endpoint  string
	}{
		{"s3-provider1", "provider1", "provider1secret", "provider1-bucket", "http://localhost:9000"},
		{"s3-provider2", "provider2", "provider2secret", "provider2-bucket", "http://localhost:9000"},
		{"s3-provider3", "provider3", "provider3secret", "provider3-bucket", "http://localhost:9000"},
	}

	for _, cfg := range s3Configs {
		service.RegisterBuilder(cp.ProviderType(cfg.name), &s3Builder{clientType: cp.ProviderType(cfg.name)})
	}
}

func createResolver() *cp.Resolver[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions] {
	resolver := cp.NewResolver[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions]()

	resolver.AddRule(cp.NewRule[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions]().
		WhenFunc(func(ctx context.Context) bool {
			provider, _ := ctx.Value(providerTypeKey).(string)
			return provider == "disk"
		}).
		Resolve(func(context.Context) (*cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions], error) {
			return &cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions]{
				Type:        cp.ProviderType("disk"),
				Credentials: storage.ProviderCredentials{},
				Config: storage.NewProviderOptions(
					storage.WithBucket("./tmp/disk-storage"),
					storage.WithBasePath("./tmp/disk-storage"),
					storage.WithLocalURL("http://localhost:8080/files"),
				),
			}, nil
		}))

	s3Configs := []struct {
		name      string
		accessKey string
		secretKey string
		bucket    string
		endpoint  string
	}{
		{"s3-provider1", "provider1", "provider1secret", "provider1-bucket", "http://localhost:9000"},
		{"s3-provider2", "provider2", "provider2secret", "provider2-bucket", "http://localhost:9000"},
		{"s3-provider3", "provider3", "provider3secret", "provider3-bucket", "http://localhost:9000"},
	}

	for _, cfg := range s3Configs {
		providerName := cfg.name
		resolver.AddRule(cp.NewRule[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions]().
			WhenFunc(func(ctx context.Context) bool {
				provider, _ := ctx.Value(providerTypeKey).(string)
				return provider == providerName
			}).
			Resolve(func(context.Context) (*cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions], error) {
				return &cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions]{
					Type: cp.ProviderType(providerName),
					Credentials: storage.ProviderCredentials{
						AccessKeyID:     cfg.accessKey,
						SecretAccessKey: cfg.secretKey,
						Endpoint:        cfg.endpoint,
					},
					Config: storage.NewProviderOptions(
						storage.WithBucket(cfg.bucket),
						storage.WithRegion("us-east-1"),
						storage.WithEndpoint(cfg.endpoint),
					),
				}, nil
			}))
	}

	return resolver
}

func testProvider(ctx context.Context, service *cp.ClientService[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions], resolver *cp.Resolver[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions], providerType, tenantID, userID string) {
	ctx = context.WithValue(ctx, providerTypeKey, providerType)
	ctx = context.WithValue(ctx, tenantIDKey, tenantID)

	resolutionOpt := resolver.Resolve(ctx)
	if !resolutionOpt.IsPresent() {
		log.Fatalf("failed to resolve provider")
	}
	resolution := resolutionOpt.MustGet()

	cacheKey := cp.ClientCacheKey{
		TenantID:        tenantID,
		IntegrationType: providerType,
	}

	clientOpt := service.GetClient(ctx, cacheKey, resolution.ClientType, resolution.Credentials, resolution.Config)
	if !clientOpt.IsPresent() {
		log.Fatalf("failed to get client")
	}
	client := clientOpt.MustGet()

	objService := storage.NewObjectService()
	content := strings.NewReader(fmt.Sprintf("Test content for %s-%s-%s", providerType, tenantID, userID))

	uploadOpts := &storage.UploadOptions{
		FileName:    fmt.Sprintf("test-%s.txt", userID),
		ContentType: "text/plain",
	}

	uploaded, err := objService.Upload(ctx, client.Provider, content, uploadOpts)
	if err != nil {
		log.Fatalf("upload failed: %v", err)
	}
	fmt.Printf("   ✓ Uploaded to %s: %s\n", providerType, uploaded.Key)

	storageFile := &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: uploaded.Key, Size: uploaded.Size, ContentType: uploaded.ContentType}}

	downloaded, err := objService.Download(ctx, client.Provider, storageFile, &storage.DownloadOptions{})
	if err != nil {
		log.Fatalf("download failed: %v", err)
	}
	fmt.Printf("   ✓ Downloaded from %s: %d bytes\n", providerType, len(downloaded.File))

	if err := objService.Delete(ctx, client.Provider, storageFile, &storagetypes.DeleteFileOptions{}); err != nil {
		log.Fatalf("delete failed: %v", err)
	}
	fmt.Printf("   ✓ Deleted from %s\n", providerType)
}

func testConcurrent(ctx context.Context, service *cp.ClientService[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions], resolver *cp.Resolver[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions]) {
	var wg sync.WaitGroup
	providers := []string{"disk", "s3-provider1", "s3-provider2", "s3-provider3"}

	for i, provider := range providers {
		wg.Add(1)
		go func(idx int, prov string) {
			defer wg.Done()

			testCtx := context.WithValue(ctx, providerTypeKey, prov)
			testCtx = context.WithValue(testCtx, tenantIDKey, fmt.Sprintf("tenant%d", idx))

			resolutionOpt := resolver.Resolve(testCtx)
			if !resolutionOpt.IsPresent() {
				log.Printf("resolve failed")
				return
			}
			resolution := resolutionOpt.MustGet()

			cacheKey := cp.ClientCacheKey{
				TenantID:        fmt.Sprintf("tenant%d", idx),
				IntegrationType: prov,
			}

			clientOpt := service.GetClient(testCtx, cacheKey, resolution.ClientType, resolution.Credentials, resolution.Config)
			if !clientOpt.IsPresent() {
				log.Printf("get client failed")
				return
			}
			client := clientOpt.MustGet()

			objService := storage.NewObjectService()
			content := strings.NewReader(fmt.Sprintf("Concurrent test %d", idx))

			uploadOpts := &storage.UploadOptions{
				FileName:    "concurrent.txt",
				ContentType: "text/plain",
			}

			if _, err := objService.Upload(testCtx, client.Provider, content, uploadOpts); err != nil {
				log.Printf("concurrent upload failed: %v", err)
				return
			}

			fmt.Printf("   ✓ Concurrent upload to %s completed\n", prov)
		}(i, provider)
	}

	wg.Wait()
	fmt.Println("   ✓ All concurrent operations completed")
}

func printPoolStats(_ *cp.ClientPool[*StorageClient]) {
	fmt.Println("   Pool statistics:")
	fmt.Println("   - Providers configured: 4 (1 disk + 3 S3)")
	fmt.Println("   - Clients cached: Based on tenant+provider combinations")
	fmt.Println("   - TTL: 10 minutes")
}

type diskBuilder struct {
	credentials storage.ProviderCredentials
	options     *storage.ProviderOptions
	clientType  cp.ProviderType
}

func (b *diskBuilder) WithCredentials(credentials storage.ProviderCredentials) cp.ClientBuilder[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions] {
	b.credentials = credentials
	return b
}

func (b *diskBuilder) WithConfig(options *storage.ProviderOptions) cp.ClientBuilder[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions] {
	if options != nil {
		b.options = options.Clone()
	}
	return b
}

func (b *diskBuilder) Build(_ context.Context) (*StorageClient, error) {
	opts := storage.NewProviderOptions(storage.WithBucket("./tmp/disk-storage"))
	if b.options != nil {
		opts = b.options.Clone()
	}
	if err := os.MkdirAll(opts.Bucket, 0o755); err != nil {
		return nil, err
	}

	provider, err := disk.NewDiskProvider(opts)
	if err != nil {
		return nil, err
	}

	return &StorageClient{
		Provider: provider,
		Type:     string(b.clientType),
	}, nil
}

func (b *diskBuilder) ClientType() cp.ProviderType {
	return b.clientType
}

type s3Builder struct {
	credentials storage.ProviderCredentials
	options     *storage.ProviderOptions
	clientType  cp.ProviderType
}

func (b *s3Builder) WithCredentials(credentials storage.ProviderCredentials) cp.ClientBuilder[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions] {
	b.credentials = credentials
	return b
}

func (b *s3Builder) WithConfig(options *storage.ProviderOptions) cp.ClientBuilder[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions] {
	if options != nil {
		b.options = options.Clone()
	}
	return b
}

func (b *s3Builder) Build(_ context.Context) (*StorageClient, error) {
	opts := storage.NewProviderOptions(storage.WithBucket("example-bucket"))
	if b.options != nil {
		opts = b.options.Clone()
	}
	opts.Credentials = b.credentials

	provider, err := s3.NewS3Provider(opts, s3.WithUsePathStyle(true))
	if err != nil {
		return nil, err
	}

	return &StorageClient{
		Provider: provider,
		Type:     string(b.clientType),
	}, nil
}

func (b *s3Builder) ClientType() cp.ProviderType {
	return b.clientType
}

func cloneProviderOptions(in *storage.ProviderOptions) *storage.ProviderOptions {
	if in == nil {
		return nil
	}
	return in.Clone()
}
