//go:build examples
// +build examples

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/theopenlane/core/pkg/cp"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/core/pkg/objects/storage/providers/disk"
	"github.com/theopenlane/core/pkg/objects/storage/providers/s3"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

func BenchmarkClientPooling(b *testing.B) {
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

	service.RegisterBuilder(cp.ProviderType("s3"), &tenantS3Builder{clientType: cp.ProviderType("s3")})

	creds := storage.ProviderCredentials{AccessKeyID: "provider1", SecretAccessKey: "provider1secret", Endpoint: defaultEndpoint}
	opts := storage.NewProviderOptions(storage.WithBucket("provider1-bucket"), storage.WithRegion(defaultRegion), storage.WithEndpoint(defaultEndpoint))
	cacheKey := cp.ClientCacheKey{TenantID: "tenant1", IntegrationType: "s3"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			clientOpt := service.GetClient(ctx, cacheKey, cp.ProviderType("s3"), creds, opts)
			if !clientOpt.IsPresent() {
				b.Fatal("failed to get client")
			}
		}
	})
}

func BenchmarkMultiTenantConcurrency(b *testing.B) {
	const numTenants = 100

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

	tenants := make([]tenant, numTenants)
	for i := 0; i < numTenants; i++ {
		tenants[i] = tenant{
			ID:        i,
			Username:  fmt.Sprintf("tenant%04d", i),
			Bucket:    fmt.Sprintf("tenant%04d-bucket", i),
			AccessKey: fmt.Sprintf("tenant%04d", i),
			SecretKey: fmt.Sprintf("tenant%04dsecret", i),
		}
	}

	registerTenantBuilders(service, tenants)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		tenantCounter := atomic.Int64{}
		for pb.Next() {
			id := int(tenantCounter.Add(1) % numTenants)
			providerType := cp.ProviderType(fmt.Sprintf("s3-tenant-%d", id))
			cacheKey := cp.ClientCacheKey{TenantID: fmt.Sprintf("tenant-%d", id), IntegrationType: string(providerType)}

			creds := storage.ProviderCredentials{AccessKeyID: tenants[id].AccessKey, SecretAccessKey: tenants[id].SecretKey, Endpoint: defaultEndpoint}
			opts := storage.NewProviderOptions(storage.WithBucket(tenants[id].Bucket), storage.WithRegion(defaultRegion), storage.WithEndpoint(defaultEndpoint))

			clientOpt := service.GetClient(ctx, cacheKey, providerType, creds, opts)
			if !clientOpt.IsPresent() {
				b.Fatal("failed to get client")
			}
		}
	})
}

func BenchmarkUploadDownload(b *testing.B) {
	opts := storage.NewProviderOptions(
		storage.WithBucket("provider1-bucket"),
		storage.WithRegion(defaultRegion),
		storage.WithEndpoint(defaultEndpoint),
		storage.WithCredentials(storage.ProviderCredentials{
			AccessKeyID:     "provider1",
			SecretAccessKey: "provider1secret",
			Endpoint:        defaultEndpoint,
		}),
	)

	provider, err := s3.NewS3Provider(opts, s3.WithUsePathStyle(true))
	if err != nil {
		b.Skip("S3-compatible endpoint not available, skipping benchmark")
	}

	objService := storage.NewObjectService()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		content := strings.NewReader(fmt.Sprintf("Benchmark data %d", i))
		uploadOpts := &storage.UploadOptions{FileName: fmt.Sprintf("bench-%d.txt", i), ContentType: "text/plain"}

		uploaded, err := objService.Upload(context.Background(), provider, content, uploadOpts)
		if err != nil {
			b.Fatal(err)
		}

		storageFile := &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: uploaded.Key, Size: uploaded.Size, ContentType: uploaded.ContentType}}
		if _, err := objService.Download(context.Background(), provider, storageFile, &storage.DownloadOptions{}); err != nil {
			b.Fatal(err)
		}
		_ = objService.Delete(context.Background(), provider, storageFile, &storagetypes.DeleteFileOptions{})
	}
}

func BenchmarkResolverPerformance(b *testing.B) {
	const numTenants = 1000

	tenants := make([]tenant, numTenants)
	for i := 0; i < numTenants; i++ {
		tenants[i] = tenant{
			ID:        i,
			Username:  fmt.Sprintf("tenant%04d", i),
			Bucket:    fmt.Sprintf("tenant%04d-bucket", i),
			AccessKey: fmt.Sprintf("tenant%04d", i),
			SecretKey: fmt.Sprintf("tenant%04dsecret", i),
		}
	}

	resolver := createResolver(tenants)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		tenantCounter := atomic.Int64{}
		for pb.Next() {
			id := int(tenantCounter.Add(1) % numTenants)
			ctx := context.WithValue(context.Background(), tenantIDKey, id)
			if !resolver.Resolve(ctx).IsPresent() {
				b.Fatal("failed to resolve tenant")
			}
		}
	})
}

func BenchmarkCacheHitRate(b *testing.B) {
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

	providerType := cp.ProviderType("s3")
	service.RegisterBuilder(providerType, &tenantS3Builder{clientType: providerType})

	creds := storage.ProviderCredentials{AccessKeyID: "provider1", SecretAccessKey: "provider1secret", Endpoint: defaultEndpoint}
	opts := storage.NewProviderOptions(storage.WithBucket("provider1-bucket"), storage.WithRegion(defaultRegion), storage.WithEndpoint(defaultEndpoint))

	ctx := context.Background()
	cacheKey := cp.ClientCacheKey{TenantID: "tenant1", IntegrationType: string(providerType)}
	_ = service.GetClient(ctx, cacheKey, providerType, creds, opts)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clientOpt := service.GetClient(ctx, cacheKey, providerType, creds, opts)
		if !clientOpt.IsPresent() {
			b.Fatal("failed to retrieve client from cache")
		}
	}
}

func BenchmarkMemoryAllocation(b *testing.B) {
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

	for i := 0; i < 100; i++ {
		providerType := cp.ProviderType(fmt.Sprintf("s3-tenant-%d", i))
		service.RegisterBuilder(providerType, &tenantS3Builder{clientType: providerType})
	}

	creds := storage.ProviderCredentials{AccessKeyID: "provider", SecretAccessKey: "secret", Endpoint: defaultEndpoint}
	opts := storage.NewProviderOptions(storage.WithBucket("benchmark-bucket"), storage.WithRegion(defaultRegion), storage.WithEndpoint(defaultEndpoint))
	cacheKey := cp.ClientCacheKey{TenantID: "tenant", IntegrationType: "s3"}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		clientOpt := service.GetClient(context.Background(), cacheKey, cp.ProviderType(fmt.Sprintf("s3-tenant-%d", i%100)), creds, opts)
		if !clientOpt.IsPresent() {
			b.Fatal("client retrieval failed")
		}
	}
}

func BenchmarkMultiProvider3000Clients(b *testing.B) {
	const tenantsPerProvider = 1000

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

	for i := 0; i < tenantsPerProvider; i++ {
		service.RegisterBuilder(cp.ProviderType(fmt.Sprintf("s3-tenant-%d", i)), &diskBenchmarkBuilder{
			clientType: cp.ProviderType(fmt.Sprintf("s3-tenant-%d", i)),
			folder:     fmt.Sprintf("./tmp/s3-sim-tenant-%04d", i),
		})
	}

	for i := 0; i < tenantsPerProvider; i++ {
		service.RegisterBuilder(cp.ProviderType(fmt.Sprintf("disk-tenant-%d", i)), &diskBenchmarkBuilder{
			clientType: cp.ProviderType(fmt.Sprintf("disk-tenant-%d", i)),
			folder:     fmt.Sprintf("./tmp/disk-tenant-%04d", i),
		})
	}

	for i := 0; i < tenantsPerProvider; i++ {
		service.RegisterBuilder(cp.ProviderType(fmt.Sprintf("gcs-tenant-%d", i)), &diskBenchmarkBuilder{
			clientType: cp.ProviderType(fmt.Sprintf("gcs-tenant-%d", i)),
			folder:     fmt.Sprintf("./tmp/gcs-tenant-%04d", i),
		})
	}

	tenants := make([]multiProviderTenant, tenantsPerProvider*3)
	idx := 0
	for i := 0; i < tenantsPerProvider; i++ {
		tenants[idx] = multiProviderTenant{ID: idx, ProviderType: "s3", TenantIndex: i}
		idx++
	}
	for i := 0; i < tenantsPerProvider; i++ {
		tenants[idx] = multiProviderTenant{ID: idx, ProviderType: "disk", TenantIndex: i}
		idx++
	}
	for i := 0; i < tenantsPerProvider; i++ {
		tenants[idx] = multiProviderTenant{ID: idx, ProviderType: "gcs", TenantIndex: i}
		idx++
	}

	resolver := createMultiProviderResolver(tenants)
	testFiles := loadTestFiles()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		tenantCounter := atomic.Int64{}
		objService := storage.NewObjectService()

		for pb.Next() {
			id := int(tenantCounter.Add(1) % int64(len(tenants)))
			ctx := context.WithValue(context.Background(), tenantIDKey, id)

			resolutionOpt := resolver.Resolve(ctx)
			if !resolutionOpt.IsPresent() {
				b.Fatal("failed to resolve provider")
			}
			resolution := resolutionOpt.MustGet()
			tenant := tenants[id]

			cacheKey := cp.ClientCacheKey{
				TenantID:        fmt.Sprintf("%s-tenant-%d", tenant.ProviderType, tenant.TenantIndex),
				IntegrationType: tenant.ProviderType,
			}

			clientOpt := service.GetClient(ctx, cacheKey, resolution.ClientType, resolution.Credentials, resolution.Config)
			if !clientOpt.IsPresent() {
				b.Fatal("failed to obtain client")
			}
			client := clientOpt.MustGet()

			fileIdx := int(tenantCounter.Load() % int64(len(testFiles)))
			testFile := testFiles[fileIdx]

			uploadOpts := &storage.UploadOptions{FileName: testFile.Name, ContentType: testFile.ContentType}
			uploaded, err := objService.Upload(ctx, client.Provider, strings.NewReader(string(testFile.Content)), uploadOpts)
			if err != nil {
				b.Fatalf("upload failed: %v", err)
			}

			storageFile := &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: uploaded.Key, Size: uploaded.Size, ContentType: uploaded.ContentType}}
			if _, err := objService.Download(ctx, client.Provider, storageFile, &storage.DownloadOptions{}); err != nil {
				b.Fatalf("download failed: %v", err)
			}
			_ = objService.Delete(ctx, client.Provider, storageFile, &storagetypes.DeleteFileOptions{})
		}
	})
}

type testFile struct {
	Name        string
	ContentType string
	Content     []byte
}

func loadTestFiles() []testFile {
	return []testFile{
		{Name: "small.txt", ContentType: "text/plain", Content: []byte("small file contents")},
		{Name: "document.json", ContentType: "application/json", Content: []byte(`{"hello":"world"}`)},
		{Name: "report.csv", ContentType: "text/csv", Content: []byte("col1,col2\n1,2")},
		{Name: "document.pdf", ContentType: "application/pdf", Content: []byte("%PDF-1.4")},
		{Name: "image.png", ContentType: "image/png", Content: []byte("PNGDATA")},
		{Name: "medium.bin", ContentType: "application/octet-stream", Content: make([]byte, 1024)},
	}
}

type multiProviderTenant struct {
	ID           int
	ProviderType string
	TenantIndex  int
}

func createMultiProviderResolver(tenants []multiProviderTenant) *cp.Resolver[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions] {
	resolver := cp.NewResolver[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions]()

	for _, tenant := range tenants {
		tenantRule := tenant
		resolver.AddRule(cp.NewRule[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions]().
			WhenFunc(func(ctx context.Context) bool {
				id, _ := ctx.Value(tenantIDKey).(int)
				return id == tenantRule.ID
			}).
			Resolve(func(context.Context) (*cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions], error) {
				providerType := cp.ProviderType(fmt.Sprintf("%s-tenant-%d", tenantRule.ProviderType, tenantRule.TenantIndex))
				options := storage.NewProviderOptions(storage.WithBucket(fmt.Sprintf("./tmp/%s-tenant-%04d", tenantRule.ProviderType, tenantRule.TenantIndex)))
				return &cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions]{
					Type:        providerType,
					Credentials: storage.ProviderCredentials{},
					Config:      options,
				}, nil
			}))
	}

	return resolver
}

type diskBenchmarkBuilder struct {
	credentials storage.ProviderCredentials
	options     *storage.ProviderOptions
	clientType  cp.ProviderType
	folder      string
}

func (b *diskBenchmarkBuilder) WithCredentials(creds storage.ProviderCredentials) cp.ClientBuilder[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions] {
	b.credentials = creds
	return b
}

func (b *diskBenchmarkBuilder) WithConfig(options *storage.ProviderOptions) cp.ClientBuilder[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions] {
	if options != nil {
		b.options = options.Clone()
	}
	return b
}

func (b *diskBenchmarkBuilder) Build(context.Context) (*StorageClient, error) {
	opts := storage.NewProviderOptions(storage.WithBucket(b.folder), storage.WithBasePath(b.folder))
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

	return &StorageClient{Provider: provider, TenantID: opts.Bucket}, nil
}

func (b *diskBenchmarkBuilder) ClientType() cp.ProviderType {
	return b.clientType
}
