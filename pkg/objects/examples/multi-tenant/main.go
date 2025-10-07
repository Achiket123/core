//go:build examples
// +build examples

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/theopenlane/core/pkg/cp"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/core/pkg/objects/storage/providers/s3"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

type tenant struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Bucket    string `json:"bucket"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

type StorageClient struct {
	Provider storagetypes.Provider
	TenantID string
}

type stats struct {
	uploads     atomic.Int64
	downloads   atomic.Int64
	cacheHits   atomic.Int64
	cacheMisses atomic.Int64
	errors      atomic.Int64
}

const (
	defaultNumOps     = 100
	defaultConcurrent = 10
	defaultRegion     = "us-east-1"
	defaultEndpoint   = "http://localhost:19000"
)

var (
	numOps     = flag.Int("ops", defaultNumOps, "Number of operations per tenant")
	concurrent = flag.Int("concurrent", defaultConcurrent, "Number of concurrent workers")
	tenantFile = flag.String("tenants", "tenants.json", "Tenant configuration file")
)

type contextKey string

const tenantIDKey contextKey = "tenant-id"

func main() {
	flag.Parse()
	ctx := context.Background()

	fmt.Println("=== Multi-Tenant High-Throughput Example ===")

	tenants, err := loadTenants(*tenantFile)
	if err != nil {
		log.Fatalf("failed to load tenants: %v", err)
	}
	fmt.Printf("Loaded %d tenants\n", len(tenants))

	pool := cp.NewClientPool[*StorageClient](30 * time.Minute)
	service := cp.NewClientService[
		*StorageClient,
		storage.ProviderCredentials,
		*storage.ProviderOptions,
	](pool, cp.WithConfigClone[
		*StorageClient,
		storage.ProviderCredentials,
		*storage.ProviderOptions,
	](cloneProviderOptions))

	fmt.Println("Registering tenant providers...")
	registerTenantBuilders(service, tenants)
	fmt.Printf("  Registered %d providers\n", len(tenants))

	resolver := createResolver(tenants)

	var s stats
	fmt.Printf("\nRunning %d operations across %d tenants with %d workers...\n",
		*numOps*len(tenants), len(tenants), *concurrent)

	startMem := getMemStats()
	start := time.Now()

	runOperations(ctx, service, resolver, tenants, &s)

	elapsed := time.Since(start)
	endMem := getMemStats()

	printResults(tenants, &s, elapsed, startMem, endMem)
}

func loadTenants(filename string) ([]tenant, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var tenants []tenant
	if err := json.NewDecoder(f).Decode(&tenants); err != nil {
		return nil, err
	}

	return tenants, nil
}

func registerTenantBuilders(service *cp.ClientService[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions], tenants []tenant) {
	for _, t := range tenants {
		providerType := cp.ProviderType(fmt.Sprintf("s3-tenant-%d", t.ID))
		service.RegisterBuilder(providerType, &tenantS3Builder{clientType: providerType})
	}
}

func createResolver(tenants []tenant) *cp.Resolver[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions] {
	resolver := cp.NewResolver[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions]()

	for _, t := range tenants {
		tenantID := t.ID
		providerType := cp.ProviderType(fmt.Sprintf("s3-tenant-%d", t.ID))

		resolver.AddRule(cp.NewRule[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions]().
			WhenFunc(func(ctx context.Context) bool {
				id, _ := ctx.Value(tenantIDKey).(int)
				return id == tenantID
			}).
			Resolve(func(context.Context) (*cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions], error) {
				return &cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions]{
					Type: providerType,
					Credentials: storage.ProviderCredentials{
						AccessKeyID:     t.AccessKey,
						SecretAccessKey: t.SecretKey,
						Endpoint:        defaultEndpoint,
					},
					Config: storage.NewProviderOptions(
						storage.WithBucket(t.Bucket),
						storage.WithRegion(defaultRegion),
						storage.WithEndpoint(defaultEndpoint),
					),
				}, nil
			}))
	}

	return resolver
}

func runOperations(ctx context.Context, service *cp.ClientService[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions], resolver *cp.Resolver[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions], tenants []tenant, s *stats) {
	var wg sync.WaitGroup
	workChan := make(chan tenant, len(tenants))

	for i := 0; i < *concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range workChan {
				performOperations(ctx, service, resolver, t, s)
			}
		}()
	}

	for _, t := range tenants {
		workChan <- t
	}
	close(workChan)

	wg.Wait()
}

func performOperations(ctx context.Context, service *cp.ClientService[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions], resolver *cp.Resolver[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions], t tenant, s *stats) {
	ctx = context.WithValue(ctx, tenantIDKey, t.ID)

	resolutionOpt := resolver.Resolve(ctx)
	if !resolutionOpt.IsPresent() {
		s.errors.Add(1)
		return
	}
	resolution := resolutionOpt.MustGet()

	cacheKey := cp.ClientCacheKey{
		TenantID:        fmt.Sprintf("tenant-%d", t.ID),
		IntegrationType: string(resolution.ClientType),
	}

	clientOpt := service.GetClient(ctx, cacheKey, resolution.ClientType, resolution.Credentials, resolution.Config)
	if !clientOpt.IsPresent() {
		s.errors.Add(1)
		s.cacheMisses.Add(1)
		return
	}
	client := clientOpt.MustGet()
	s.cacheHits.Add(1)

	objService := storage.NewObjectService()

	for i := 0; i < *numOps; i++ {
		content := strings.NewReader(fmt.Sprintf("Test data for tenant %d operation %d", t.ID, i))

		uploadOpts := &storage.UploadOptions{
			FileName:    fmt.Sprintf("file-%d.txt", i),
			ContentType: "text/plain",
		}

		uploaded, err := objService.Upload(ctx, client.Provider, content, uploadOpts)
		if err != nil {
			s.errors.Add(1)
			continue
		}
		s.uploads.Add(1)

		storageFile := &storagetypes.File{
			FileMetadata: storagetypes.FileMetadata{
				Key:         uploaded.Key,
				Size:        uploaded.Size,
				ContentType: uploaded.ContentType,
			},
		}

		if _, err := objService.Download(ctx, client.Provider, storageFile, &storage.DownloadOptions{}); err != nil {
			s.errors.Add(1)
			continue
		}
		s.downloads.Add(1)

		if err := objService.Delete(ctx, client.Provider, storageFile, &storagetypes.DeleteFileOptions{}); err != nil {
			s.errors.Add(1)
		}
	}
}

func getMemStats() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}

func printResults(tenants []tenant, s *stats, elapsed time.Duration, startMem, endMem runtime.MemStats) {
	fmt.Println("\n=== Results ===")
	fmt.Printf("\nTenants: %d\n", len(tenants))
	fmt.Printf("Total Operations: %d\n", s.uploads.Load()+s.downloads.Load())
	fmt.Printf("  Uploads: %d\n", s.uploads.Load())
	fmt.Printf("  Downloads: %d\n", s.downloads.Load())
	fmt.Printf("  Errors: %d\n", s.errors.Load())

	fmt.Printf("\nCache Statistics:\n")
	totalCache := s.cacheHits.Load() + s.cacheMisses.Load()
	hitRate := 0.0
	if totalCache > 0 {
		hitRate = float64(s.cacheHits.Load()) / float64(totalCache) * 100
	}
	fmt.Printf("  Hits: %d\n", s.cacheHits.Load())
	fmt.Printf("  Misses: %d\n", s.cacheMisses.Load())
	fmt.Printf("  Hit Rate: %.2f%%\n", hitRate)

	fmt.Printf("\nPerformance:\n")
	fmt.Printf("  Total Time: %v\n", elapsed)
	totalOps := s.uploads.Load() + s.downloads.Load()
	if totalOps > 0 {
		fmt.Printf("  Operations/sec: %.2f\n", float64(totalOps)/elapsed.Seconds())
		fmt.Printf("  Avg Time/op: %v\n", elapsed/time.Duration(totalOps))
	}

	fmt.Printf("\nMemory:\n")
	const bytesToMB = 1024 * 1024
	fmt.Printf("  Start Alloc: %.2f MB\n", float64(startMem.Alloc)/bytesToMB)
	fmt.Printf("  End Alloc: %.2f MB\n", float64(endMem.Alloc)/bytesToMB)
	fmt.Printf("  Delta: %.2f MB\n", float64(endMem.Alloc-startMem.Alloc)/bytesToMB)
	fmt.Printf("  Sys: %.2f MB\n", float64(endMem.Sys)/bytesToMB)
	fmt.Printf("  NumGC: %d\n", endMem.NumGC-startMem.NumGC)

	fmt.Println("\n=== Example completed successfully ===")
}

type tenantS3Builder struct {
	credentials storage.ProviderCredentials
	options     *storage.ProviderOptions
	clientType  cp.ProviderType
}

func (b *tenantS3Builder) WithCredentials(credentials storage.ProviderCredentials) cp.ClientBuilder[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions] {
	b.credentials = credentials
	return b
}

func (b *tenantS3Builder) WithConfig(options *storage.ProviderOptions) cp.ClientBuilder[*StorageClient, storage.ProviderCredentials, *storage.ProviderOptions] {
	if options != nil {
		b.options = options.Clone()
	}
	return b
}

func (b *tenantS3Builder) Build(context.Context) (*StorageClient, error) {
	opts := storage.NewProviderOptions()
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
		TenantID: opts.Bucket,
	}, nil
}

func (b *tenantS3Builder) ClientType() cp.ProviderType {
	return b.clientType
}

func cloneProviderOptions(in *storage.ProviderOptions) *storage.ProviderOptions {
	if in == nil {
		return nil
	}
	return in.Clone()
}
