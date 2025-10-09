package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/theopenlane/core/pkg/eddy"
	"github.com/theopenlane/core/pkg/objects/storage"
	s3provider "github.com/theopenlane/core/pkg/objects/storage/providers/s3"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

func multiTenantCommand() *cli.Command {
	return &cli.Command{
		Name:  "multi-tenant",
		Usage: "Run the high-throughput multi-tenant example",
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "ops", Usage: "Number of operations per tenant", Value: 100, Sources: cli.EnvVars("OBJECTS_EXAMPLE_TENANT_OPS")},
			&cli.IntFlag{Name: "concurrent", Usage: "Number of concurrent workers", Value: 10, Sources: cli.EnvVars("OBJECTS_EXAMPLE_TENANT_CONCURRENCY")},
			&cli.StringFlag{Name: "tenants", Usage: "Path to tenant configuration file", Value: "tenants.json", Sources: cli.EnvVars("OBJECTS_EXAMPLE_TENANT_FILE")},
		},
		Commands: []*cli.Command{
			{Name: "setup", Usage: "Provision tenants and supporting resources", Action: func(ctx context.Context, cmd *cli.Command) error {
				out := cmd.Writer
				if out == nil {
					out = os.Stdout
				}
				return multiTenantSetup(ctx, out, multiTenantSetupConfig{
					TenantCount: cmd.Int("tenants"),
				})
			}},
			{Name: "setup-1000", Usage: "Provision 1000 tenants", Action: func(ctx context.Context, cmd *cli.Command) error {
				out := cmd.Writer
				if out == nil {
					out = os.Stdout
				}
				return multiTenantSetup(ctx, out, multiTenantSetupConfig{TenantCount: 1000, Parallel: 20})
			}},
			{Name: "teardown", Usage: "Remove generated tenants and docker services", Action: func(ctx context.Context, cmd *cli.Command) error {
				out := cmd.Writer
				if out == nil {
					out = os.Stdout
				}
				return multiTenantTeardown(ctx, out)
			}},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			runCfg := multiTenantRunConfig{
				OpsPerTenant:  cmd.Int("ops"),
				Concurrent:    cmd.Int("concurrent"),
				TenantCfgPath: cmd.String("tenants"),
			}
			out := cmd.Writer
			if out == nil {
				out = os.Stdout
			}
			return runMultiTenant(ctx, out, runCfg)
		},
	}
}

// --- Setup ---

type multiTenantSetupConfig struct {
	TenantCount int
	Parallel    int
}

func multiTenantSetup(ctx context.Context, out io.Writer, cfg multiTenantSetupConfig) error {
	if cfg.TenantCount <= 0 {
		cfg.TenantCount = 10
	}
	if cfg.Parallel <= 0 {
		cfg.Parallel = 5
	}

	fmt.Fprintln(out, "Starting docker services for multi-tenant example...")
	if err := runCommand(ctx, out, "docker-compose", "-f", composeFilePath(), "up", "-d", "minio"); err != nil {
		return err
	}

	fmt.Fprintln(out, "Waiting for MinIO...")
	if err := waitForMinIO(ctx, "objects-examples-minio", "admin", "adminsecretpassword"); err != nil {
		return err
	}

	tenants, err := provisionTenants(ctx, cfg)
	if err != nil {
		return err
	}

	path := resolvePath("tenants.json")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(f).Encode(tenants); err != nil {
		f.Close()
		return err
	}
	f.Close()

	fmt.Fprintf(out, "Provisioned %d tenants and wrote %s\n", len(tenants), path)
	return nil
}

func provisionTenants(ctx context.Context, cfg multiTenantSetupConfig) ([]tenant, error) {
	type result struct {
		tenant tenant
		err    error
	}

	results := make(chan result, cfg.TenantCount)
	sem := make(chan struct{}, cfg.Parallel)

	for i := 0; i < cfg.TenantCount; i++ {
		sem <- struct{}{}
		go func(idx int) {
			defer func() { <-sem }()
			r := result{tenant: tenant{ID: idx}}

			bucket := fmt.Sprintf("tenant-%04d", idx)
			username := fmt.Sprintf("tenant-user-%04d", idx)
			password := fmt.Sprintf("tenant-secret-%04d", idx)

			if err := createMinIOUser(ctx, username, password); err != nil {
				r.err = err
			} else if err := createMinIOBucket(ctx, bucket); err != nil {
				r.err = err
			}

			r.tenant.Username = username
			r.tenant.Password = password
			r.tenant.Bucket = bucket
			r.tenant.AccessKey = username
			r.tenant.SecretKey = password
			results <- r
		}(i)
	}

	for i := 0; i < cfg.Parallel; i++ {
		sem <- struct{}{}
	}
	close(results)

	var tenants []tenant
	for r := range results {
		if r.err != nil {
			return nil, r.err
		}
		tenants = append(tenants, r.tenant)
	}

	return tenants, nil
}

func createMinIOUser(ctx context.Context, username, password string) error {
	cmd := exec.CommandContext(ctx, "docker", "exec", "objects-examples-minio",
		"mc", "admin", "user", "add", "local", username, password)
	if output, err := cmd.CombinedOutput(); err != nil {
		if !strings.Contains(string(output), "already exists") {
			return fmt.Errorf("create user %s: %w - %s", username, err, output)
		}
	}

	cmd = exec.CommandContext(ctx, "docker", "exec", "objects-examples-minio",
		"mc", "admin", "policy", "attach", "local", "readwrite", "--user", username)
	if output, err := cmd.CombinedOutput(); err != nil {
		if !strings.Contains(string(output), "already mapped") {
			return fmt.Errorf("attach policy %s: %w - %s", username, err, output)
		}
	}

	return nil
}

func createMinIOBucket(ctx context.Context, bucket string) error {
	cmd := exec.CommandContext(ctx, "docker", "exec", "objects-examples-minio",
		"mc", "mb", fmt.Sprintf("local/%s", bucket), "--ignore-existing")
	return cmd.Run()
}

func multiTenantTeardown(ctx context.Context, out io.Writer) error {
	fmt.Fprintln(out, "Stopping services and cleaning files...")
	_ = runCommand(ctx, out, "docker-compose", "-f", composeFilePath(), "down", "--remove-orphans")
	os.Remove(resolvePath("tenants.json"))
	return nil
}

// --- Run Example ---

type tenant struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Bucket    string `json:"bucket"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

type multiTenantRunConfig struct {
	OpsPerTenant  int
	Concurrent    int
	TenantCfgPath string
}

type multiTenantClient struct {
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

type tenantKey struct{}

type tenantCacheKey struct {
	TenantID   string
	ProviderID string
}

func (k tenantCacheKey) String() string {
	return fmt.Sprintf("%s:%s", k.TenantID, k.ProviderID)
}

type contextKey string

func runMultiTenant(ctx context.Context, out io.Writer, cfg multiTenantRunConfig) error {
	if cfg.OpsPerTenant <= 0 {
		cfg.OpsPerTenant = 100
	}
	if cfg.Concurrent <= 0 {
		cfg.Concurrent = 10
	}

	fmt.Fprintln(out, "=== Multi-Tenant High-Throughput Example ===")

	tenants, err := loadTenants(resolvePath(cfg.TenantCfgPath))
	if err != nil {
		return fmt.Errorf("load tenants: %w", err)
	}
	fmt.Fprintf(out, "Loaded %d tenants\n", len(tenants))

	pool := eddy.NewClientPool[*multiTenantClient](30 * time.Minute)
	service := eddy.NewClientService[
		*multiTenantClient,
		storage.ProviderCredentials,
		*storage.ProviderOptions,
	](pool, eddy.WithConfigClone[
		*multiTenantClient,
		storage.ProviderCredentials,
		*storage.ProviderOptions,
	](cloneProviderOptions))

	resolver := createTenantResolver(tenants)

	var s stats
	totalOps := cfg.OpsPerTenant * len(tenants)
	fmt.Fprintf(out, "\nRunning %d operations across %d tenants with %d workers...\n", totalOps, len(tenants), cfg.Concurrent)

	startMem := getMemStats()
	start := time.Now()

	if err := performTenantOperations(ctx, out, service, resolver, tenants, cfg, &s); err != nil {
		return err
	}

	elapsed := time.Since(start)
	endMem := getMemStats()
	printTenantResults(out, tenants, &s, elapsed, startMem, endMem)
	return nil
}

func loadTenants(path string) ([]tenant, error) {
	f, err := os.Open(path)
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

func createTenantResolver(tenants []tenant) *eddy.Resolver[*multiTenantClient, storage.ProviderCredentials, *storage.ProviderOptions] {
	resolver := eddy.NewResolver[*multiTenantClient, storage.ProviderCredentials, *storage.ProviderOptions]()

	for _, t := range tenants {
		tenantRule := t
		providerType := fmt.Sprintf("s3-tenant-%d", tenantRule.ID)
		builder := &tenantS3Builder{providerType: providerType}

		resolver.AddRule(eddy.NewRule[*multiTenantClient, storage.ProviderCredentials, *storage.ProviderOptions]().
			WhenFunc(func(ctx context.Context) bool {
				id, _ := ctx.Value(tenantKey{}).(int)
				return id == tenantRule.ID
			}).
			Resolve(func(context.Context) (*eddy.ResolvedProvider[*multiTenantClient, storage.ProviderCredentials, *storage.ProviderOptions], error) {
				return &eddy.ResolvedProvider[*multiTenantClient, storage.ProviderCredentials, *storage.ProviderOptions]{
					Builder: builder,
					Output: storage.ProviderCredentials{
						AccessKeyID:     tenantRule.AccessKey,
						SecretAccessKey: tenantRule.SecretKey,
						Endpoint:        "http://localhost:19000",
					},
					Config: storage.NewProviderOptions(
						storage.WithBucket(tenantRule.Bucket),
						storage.WithRegion("us-east-1"),
						storage.WithEndpoint("http://localhost:19000"),
					),
				}, nil
			}))

	}

	return resolver
}

func performTenantOperations(ctx context.Context, out io.Writer, service *eddy.ClientService[*multiTenantClient, storage.ProviderCredentials, *storage.ProviderOptions], resolver *eddy.Resolver[*multiTenantClient, storage.ProviderCredentials, *storage.ProviderOptions], tenants []tenant, cfg multiTenantRunConfig, s *stats) error {
	var wg sync.WaitGroup
	work := make(chan tenant, len(tenants))
	errCh := make(chan error, cfg.Concurrent)

	for i := 0; i < cfg.Concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for tenant := range work {
				if err := tenantWorkflow(ctx, service, resolver, tenant, cfg.OpsPerTenant, s); err != nil {
					errCh <- err
				}
			}
		}()
	}

	for _, t := range tenants {
		work <- t
	}
	close(work)
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func tenantWorkflow(ctx context.Context, service *eddy.ClientService[*multiTenantClient, storage.ProviderCredentials, *storage.ProviderOptions], resolver *eddy.Resolver[*multiTenantClient, storage.ProviderCredentials, *storage.ProviderOptions], t tenant, ops int, s *stats) error {
	ctx = context.WithValue(ctx, tenantKey{}, t.ID)

	resolutionOpt := resolver.Resolve(ctx)
	if !resolutionOpt.IsPresent() {
		s.errors.Add(1)
		return fmt.Errorf("resolve failed for tenant %d", t.ID)
	}
	resolution := resolutionOpt.MustGet()

	cacheKey := tenantCacheKey{
		TenantID:   fmt.Sprintf("tenant-%d", t.ID),
		ProviderID: resolution.Builder.ProviderType(),
	}

	clientOpt := service.GetClient(ctx, cacheKey, resolution.Builder, resolution.Output, resolution.Config)
	if !clientOpt.IsPresent() {
		s.errors.Add(1)
		s.cacheMisses.Add(1)
		return fmt.Errorf("client acquisition failed for tenant %d", t.ID)
	}
	client := clientOpt.MustGet()
	s.cacheHits.Add(1)

	objService := storage.NewObjectService()

	for i := 0; i < ops; i++ {
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

	return nil
}

func printTenantResults(out io.Writer, tenants []tenant, s *stats, elapsed time.Duration, startMem, endMem runtime.MemStats) {
	fmt.Fprintln(out, "\n=== Results ===")
	fmt.Fprintf(out, "\nTenants: %d\n", len(tenants))
	fmt.Fprintf(out, "Total Operations: %d\n", s.uploads.Load()+s.downloads.Load())
	fmt.Fprintf(out, "  Uploads: %d\n", s.uploads.Load())
	fmt.Fprintf(out, "  Downloads: %d\n", s.downloads.Load())
	fmt.Fprintf(out, "  Errors: %d\n", s.errors.Load())

	fmt.Fprintln(out, "\nCache Statistics:")
	totalCache := s.cacheHits.Load() + s.cacheMisses.Load()
	hitRate := 0.0
	if totalCache > 0 {
		hitRate = float64(s.cacheHits.Load()) / float64(totalCache) * 100
	}
	fmt.Fprintf(out, "  Hits: %d\n", s.cacheHits.Load())
	fmt.Fprintf(out, "  Misses: %d\n", s.cacheMisses.Load())
	fmt.Fprintf(out, "  Hit Rate: %.2f%%\n", hitRate)

	fmt.Fprintln(out, "\nPerformance:")
	fmt.Fprintf(out, "  Total Time: %v\n", elapsed)
	totalOps := s.uploads.Load() + s.downloads.Load()
	if totalOps > 0 {
		fmt.Fprintf(out, "  Operations/sec: %.2f\n", float64(totalOps)/elapsed.Seconds())
		fmt.Fprintf(out, "  Avg Time/op: %v\n", elapsed/time.Duration(totalOps))
	}

	fmt.Fprintln(out, "\nMemory:")
	const bytesToMB = 1024 * 1024
	fmt.Fprintf(out, "  Start Alloc: %.2f MB\n", float64(startMem.Alloc)/bytesToMB)
	fmt.Fprintf(out, "  End Alloc: %.2f MB\n", float64(endMem.Alloc)/bytesToMB)
	fmt.Fprintf(out, "  Delta: %.2f MB\n", float64(endMem.Alloc-startMem.Alloc)/bytesToMB)
	fmt.Fprintf(out, "  Sys: %.2f MB\n", float64(endMem.Sys)/bytesToMB)
	fmt.Fprintf(out, "  NumGC: %d\n", endMem.NumGC-startMem.NumGC)

	fmt.Fprintln(out, "\n=== Example completed successfully ===")
}

func getMemStats() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}

type tenantS3Builder struct {
	providerType string
}

func (b *tenantS3Builder) Build(_ context.Context, credentials storage.ProviderCredentials, options *storage.ProviderOptions) (*multiTenantClient, error) {
	opts := storage.NewProviderOptions()
	if options != nil {
		opts = options.Clone()
	}
	opts.Apply(storage.WithCredentials(credentials))

	provider, err := s3provider.NewS3Provider(opts, s3provider.WithUsePathStyle(true))
	if err != nil {
		return nil, err
	}

	return &multiTenantClient{
		Provider: provider,
		TenantID: opts.Bucket,
	}, nil
}

func (b *tenantS3Builder) ProviderType() string {
	return b.providerType
}
