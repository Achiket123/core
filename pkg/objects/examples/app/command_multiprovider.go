package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/theopenlane/core/pkg/eddy"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/core/pkg/objects/storage/providers/disk"
	s3provider "github.com/theopenlane/core/pkg/objects/storage/providers/s3"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

func multiProviderCommand() *cli.Command {
	const skipSetupFlag = "skip-setup"
	const skipTeardownFlag = "skip-teardown"

	return &cli.Command{
		Name:  "multi-provider",
		Usage: "Demonstrate provider resolution across disk and S3 backends",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: skipSetupFlag, Usage: "Assume infrastructure is already running and provisioned"},
			&cli.BoolFlag{Name: skipTeardownFlag, Usage: "Leave supporting services running after the example"},
		},
		Commands: []*cli.Command{
			{Name: "setup", Usage: "Start docker services and seed credentials", Action: func(ctx context.Context, cmd *cli.Command) error {
				out := cmd.Writer
				if out == nil {
					out = os.Stdout
				}
				return multiProviderSetup(ctx, out)
			}},
			{Name: "teardown", Usage: "Stop docker services and remove containers", Action: func(ctx context.Context, cmd *cli.Command) error {
				out := cmd.Writer
				if out == nil {
					out = os.Stdout
				}
				return multiProviderTeardown(ctx, out)
			}},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg := multiProviderRunConfig{
				SkipSetup:    cmd.Bool(skipSetupFlag),
				SkipTeardown: cmd.Bool(skipTeardownFlag),
			}
			out := cmd.Writer
			if out == nil {
				out = os.Stdout
			}
			return runMultiProvider(ctx, out, cfg)
		},
	}
}

type multiProviderRunConfig struct {
	SkipSetup    bool
	SkipTeardown bool
}

type multiProviderStorageClient struct {
	Provider storagetypes.Provider
	Type     string
}

type multiProviderCacheKey struct {
	TenantID   string
	ProviderID string
}

func (k multiProviderCacheKey) String() string {
	return fmt.Sprintf("%s:%s", k.TenantID, k.ProviderID)
}

func runMultiProvider(ctx context.Context, out io.Writer, cfg multiProviderRunConfig) error {
	if !cfg.SkipSetup {
		if err := multiProviderSetup(ctx, out); err != nil {
			return err
		}
	}

	fmt.Fprintln(out, "=== Multi-Provider Object Storage Example ===")

	pool := eddy.NewClientPool[*multiProviderStorageClient](10 * time.Minute)
	service := eddy.NewClientService[
		*multiProviderStorageClient,
		storage.ProviderCredentials,
		*storage.ProviderOptions,
	](pool, eddy.WithConfigClone[
		*multiProviderStorageClient,
		storage.ProviderCredentials,
		*storage.ProviderOptions,
	](cloneProviderOptions))

	resolver := createMultiProviderResolver()

	scenarios := []struct {
		heading  string
		provider string
		tenantID string
		userID   string
	}{
		{"1. Testing Disk Provider...", "disk", "tenant1", "provider1"},
		{"\n2. Testing S3 Provider (MinIO - Provider 1)...", "s3-provider1", "tenant1", "provider1"},
		{"\n3. Testing S3 Provider (MinIO - Provider 2)...", "s3-provider2", "tenant2", "provider2"},
		{"\n4. Testing S3 Provider (MinIO - Provider 3)...", "s3-provider3", "tenant3", "provider3"},
	}

	for _, scenario := range scenarios {
		fmt.Fprintln(out, scenario.heading)
		if err := exerciseProvider(ctx, out, service, resolver, scenario.provider, scenario.tenantID, scenario.userID); err != nil {
			return err
		}
	}

	fmt.Fprintln(out, "\n5. Concurrent Operations Across All Providers...")
	if err := multiProviderConcurrent(ctx, out, service, resolver); err != nil {
		return err
	}

	fmt.Fprintln(out, "\n6. Pool Statistics...")
	printMultiProviderPoolStats(out)
	fmt.Fprintln(out, "\n=== Example completed successfully ===")

	if !cfg.SkipTeardown {
		if err := multiProviderTeardown(ctx, out); err != nil {
			return err
		}
	}

	return nil
}

func createMultiProviderResolver() *eddy.Resolver[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions] {
	resolver := eddy.NewResolver[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions]()

	diskBuilder := &multiProviderDiskBuilder{providerType: "disk"}
	resolver.AddRule(eddy.NewRule[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions]().
		WhenFunc(func(ctx context.Context) bool {
			provider, _ := ctx.Value(multiProviderTypeKey{}).(string)
			return provider == "disk"
		}).
		Resolve(func(context.Context) (*eddy.ResolvedProvider[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions], error) {
			return &eddy.ResolvedProvider[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions]{
				Builder: diskBuilder,
				Output:  storage.ProviderCredentials{},
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
		{"s3-provider1", "provider1", "provider1secret", "provider1-bucket", "http://localhost:19000"},
		{"s3-provider2", "provider2", "provider2secret", "provider2-bucket", "http://localhost:19000"},
		{"s3-provider3", "provider3", "provider3secret", "provider3-bucket", "http://localhost:19000"},
	}

	for _, cfg := range s3Configs {
		providerName := cfg.name
		builder := &multiProviderS3Builder{providerType: providerName}
		resolver.AddRule(eddy.NewRule[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions]().
			WhenFunc(func(ctx context.Context) bool {
				provider, _ := ctx.Value(multiProviderTypeKey{}).(string)
				return provider == providerName
			}).
			Resolve(func(context.Context) (*eddy.ResolvedProvider[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions], error) {
				return &eddy.ResolvedProvider[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions]{
					Builder: builder,
					Output: storage.ProviderCredentials{
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

type multiProviderTypeKey struct{}

type multiProviderTenantKey struct{}

func exerciseProvider(ctx context.Context, out io.Writer, service *eddy.ClientService[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions], resolver *eddy.Resolver[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions], providerType, tenantID, userID string) error {
	ctx = context.WithValue(ctx, multiProviderTypeKey{}, providerType)
	ctx = context.WithValue(ctx, multiProviderTenantKey{}, tenantID)

	resolutionOpt := resolver.Resolve(ctx)
	if !resolutionOpt.IsPresent() {
		return fmt.Errorf("provider resolution failed for %s", providerType)
	}
	resolution := resolutionOpt.MustGet()

	cacheKey := multiProviderCacheKey{
		TenantID:   tenantID,
		ProviderID: resolution.Builder.ProviderType(),
	}

	clientOpt := service.GetClient(ctx, cacheKey, resolution.Builder, resolution.Output, resolution.Config)
	if !clientOpt.IsPresent() {
		return fmt.Errorf("client acquisition failed for %s", providerType)
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
		return fmt.Errorf("upload failed for %s: %w", providerType, err)
	}
	fmt.Fprintf(out, "Uploaded to %s: %s\n", providerType, uploaded.Key)

	storageFile := &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: uploaded.Key, Size: uploaded.Size, ContentType: uploaded.ContentType}}

	downloaded, err := objService.Download(ctx, client.Provider, storageFile, &storage.DownloadOptions{})
	if err != nil {
		return fmt.Errorf("download failed for %s: %w", providerType, err)
	}
	fmt.Fprintf(out, "Downloaded from %s: %d bytes\n", providerType, len(downloaded.File))

	if err := objService.Delete(ctx, client.Provider, storageFile, &storagetypes.DeleteFileOptions{}); err != nil {
		return fmt.Errorf("delete failed for %s: %w", providerType, err)
	}
	fmt.Fprintf(out, "Deleted from %s\n", providerType)

	return nil
}

func multiProviderConcurrent(ctx context.Context, out io.Writer, service *eddy.ClientService[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions], resolver *eddy.Resolver[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions]) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	providers := []string{"disk", "s3-provider1", "s3-provider2", "s3-provider3"}
	errCh := make(chan error, len(providers))

	for i, provider := range providers {
		wg.Add(1)
		go func(idx int, prov string) {
			defer wg.Done()

			testCtx := context.WithValue(ctx, multiProviderTypeKey{}, prov)
			testCtx = context.WithValue(testCtx, multiProviderTenantKey{}, fmt.Sprintf("tenant%d", idx))

			resolutionOpt := resolver.Resolve(testCtx)
			if !resolutionOpt.IsPresent() {
				errCh <- fmt.Errorf("resolve failed for %s", prov)
				return
			}
			resolution := resolutionOpt.MustGet()

			cacheKey := multiProviderCacheKey{
				TenantID:   fmt.Sprintf("tenant%d", idx),
				ProviderID: resolution.Builder.ProviderType(),
			}

			clientOpt := service.GetClient(testCtx, cacheKey, resolution.Builder, resolution.Output, resolution.Config)
			if !clientOpt.IsPresent() {
				errCh <- fmt.Errorf("client retrieval failed for %s", prov)
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
				errCh <- fmt.Errorf("concurrent upload failed for %s: %w", prov, err)
				return
			}

			mu.Lock()
			fmt.Fprintf(out, "Concurrent upload to %s completed\n", prov)
			mu.Unlock()
		}(i, provider)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	fmt.Fprintln(out, "All concurrent operations completed")
	return nil
}

func printMultiProviderPoolStats(out io.Writer) {
	fmt.Fprintln(out, "   Pool statistics:")
	fmt.Fprintln(out, "   - Providers configured: 4 (1 disk + 3 S3)")
	fmt.Fprintln(out, "   - Clients cached: Based on tenant+provider combinations")
	fmt.Fprintln(out, "   - TTL: 10 minutes")
}

type multiProviderDiskBuilder struct {
	providerType string
}

func (b *multiProviderDiskBuilder) Build(_ context.Context, _ storage.ProviderCredentials, options *storage.ProviderOptions) (*multiProviderStorageClient, error) {
	opts := storage.NewProviderOptions()
	if options != nil {
		opts = options.Clone()
	} else {
		opts = storage.NewProviderOptions(storage.WithBucket("./tmp/disk-storage"), storage.WithBasePath("./tmp/disk-storage"))
	}
	if err := os.MkdirAll(opts.Bucket, 0o755); err != nil {
		return nil, err
	}

	provider, err := disk.NewDiskProvider(opts)
	if err != nil {
		return nil, err
	}

	return &multiProviderStorageClient{
		Provider: provider,
		Type:     b.providerType,
	}, nil
}

func (b *multiProviderDiskBuilder) ProviderType() string {
	return b.providerType
}

type multiProviderS3Builder struct {
	providerType string
}

func (b *multiProviderS3Builder) Build(_ context.Context, credentials storage.ProviderCredentials, options *storage.ProviderOptions) (*multiProviderStorageClient, error) {
	opts := storage.NewProviderOptions()
	if options != nil {
		opts = options.Clone()
	}
	opts.Apply(storage.WithCredentials(credentials))

	provider, err := s3provider.NewS3Provider(opts, s3provider.WithUsePathStyle(true))
	if err != nil {
		return nil, err
	}

	return &multiProviderStorageClient{
		Provider: provider,
		Type:     b.providerType,
	}, nil
}

func (b *multiProviderS3Builder) ProviderType() string {
	return b.providerType
}

func multiProviderSetup(ctx context.Context, out io.Writer, composeArgs ...string) error {
	fmt.Fprintln(out, "Starting docker services...")
	args := append([]string{"-f", composeFilePath(), "up", "-d"}, composeArgs...)
	if err := runCommand(ctx, out, "docker-compose", args...); err != nil {
		return err
	}

	fmt.Fprintln(out, "Waiting for MinIO...")
	if err := waitForMinIO(ctx, "objects-examples-minio", "admin", "adminsecretpassword"); err != nil {
		return err
	}

	fmt.Fprintln(out, "Waiting for fake GCS...")
	if err := waitForGCS(ctx); err != nil {
		return err
	}

	fmt.Fprintln(out, "Configuring MinIO users and buckets...")
	users := []minioUser{
		{username: "provider1", password: "provider1secret", bucket: "provider1-bucket"},
		{username: "provider2", password: "provider2secret", bucket: "provider2-bucket"},
		{username: "provider3", password: "provider3secret", bucket: "provider3-bucket"},
	}

	if err := setupMinIOUsers(ctx, users); err != nil {
		return err
	}

	fmt.Fprintln(out, "Seeding fake GCS buckets...")
	if err := setupGCSBuckets(ctx); err != nil {
		return err
	}

	fmt.Fprintln(out, "Infrastructure ready.")
	return nil
}

func multiProviderTeardown(ctx context.Context, out io.Writer) error {
	fmt.Fprintln(out, "Stopping docker services...")
	return runCommand(ctx, out, "docker-compose", "-f", composeFilePath(), "down", "--remove-orphans")
}

type minioUser struct {
	username string
	password string
	bucket   string
}

func runCommand(ctx context.Context, out io.Writer, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = outWriter(out)
	cmd.Stderr = outWriter(out)
	return cmd.Run()
}

func outWriter(out io.Writer) io.Writer {
	if out == nil {
		return os.Stdout
	}
	return out
}

func waitForMinIO(ctx context.Context, container, accessKey, secretKey string) error {
	for i := 0; i < 30; i++ {
		cmd := exec.CommandContext(ctx, "docker", "exec", container,
			"mc", "alias", "set", "local", "http://localhost:9000", accessKey, secretKey)
		if err := cmd.Run(); err == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("timeout waiting for MinIO")
}

func waitForGCS(ctx context.Context) error {
	client := &http.Client{Timeout: 5 * time.Second}
	for i := 0; i < 30; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:4443/storage/v1/b", nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("timeout waiting for fake GCS")
}

func setupMinIOUsers(ctx context.Context, users []minioUser) error {
	for _, user := range users {
		cmd := exec.CommandContext(ctx, "docker", "exec", "objects-examples-minio",
			"mc", "admin", "user", "add", "local", user.username, user.password)
		if output, err := cmd.CombinedOutput(); err != nil {
			if !bytes.Contains(output, []byte("already exists")) {
				return fmt.Errorf("failed to create user %s: %w - %s", user.username, err, output)
			}
		}

		cmd = exec.CommandContext(ctx, "docker", "exec", "objects-examples-minio",
			"mc", "mb", fmt.Sprintf("local/%s", user.bucket), "--ignore-existing")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create bucket %s: %w", user.bucket, err)
		}

		cmd = exec.CommandContext(ctx, "docker", "exec", "objects-examples-minio",
			"mc", "admin", "policy", "attach", "local", "readwrite", "--user", user.username)
		if output, err := cmd.CombinedOutput(); err != nil {
			if !bytes.Contains(output, []byte("already mapped")) {
				return fmt.Errorf("failed to attach policy for %s: %w - %s", user.username, err, output)
			}
		}
	}

	return nil
}

func setupGCSBuckets(ctx context.Context) error {
	buckets := []string{"provider1-gcs", "provider2-gcs", "provider3-gcs"}
	for _, bucket := range buckets {
		payload := map[string]string{"name": bucket}
		body, _ := json.Marshal(payload)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost:4443/storage/v1/b", bytes.NewReader(body))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		resp.Body.Close()
	}
	return nil
}
