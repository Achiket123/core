//go:build examples
// +build examples

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/theopenlane/core/pkg/objects/storage"
	s3local "github.com/theopenlane/core/pkg/objects/storage/providers/s3"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

func main() {
	var (
		endpoint     = flag.String("endpoint", envOrDefault("MINIO_ENDPOINT", "http://127.0.0.1:9000"), "S3 or MinIO endpoint URL")
		accessKey    = flag.String("access-key", envOrDefault("MINIO_ACCESS_KEY", "minioadmin"), "Access key ID")
		secretKey    = flag.String("secret-key", envOrDefault("MINIO_SECRET_KEY", "minioadmin"), "Secret access key")
		region       = flag.String("region", envOrDefault("MINIO_REGION", "us-east-1"), "Region to use")
		bucket       = flag.String("bucket", envOrDefault("MINIO_BUCKET", "core-simple-s3"), "Bucket to read/write")
		sourcePath   = flag.String("source", "assets/sample-data.txt", "File to upload")
		objectKey    = flag.String("object", "examples/simple-s3/sample-data.txt", "Object key inside the bucket")
		downloadPath = flag.String("download", "output/downloaded-sample.txt", "Destination path for downloaded file")
	)
	flag.Parse()

	ctx := context.Background()

	awsClient, err := newS3Client(ctx, *endpoint, *accessKey, *secretKey, *region)
	if err != nil {
		panicf("failed to create AWS client: %v", err)
	}

	if err := ensureBucket(ctx, awsClient, *bucket); err != nil {
		panicf("failed to ensure bucket %q: %v", *bucket, err)
	}

	providerOptions := storage.NewProviderOptions(
		storage.WithBucket(*bucket),
		storage.WithRegion(*region),
		storage.WithEndpoint(*endpoint),
		storage.WithCredentials(storage.ProviderCredentials{
			AccessKeyID:     *accessKey,
			SecretAccessKey: *secretKey,
			Endpoint:        *endpoint,
		}),
	)

	provider, err := s3local.NewS3Provider(providerOptions, s3local.WithUsePathStyle(true))
	if err != nil {
		panicf("failed to create S3 provider: %v", err)
	}
	defer provider.Close()

	objectService := storage.NewObjectService()

	srcFile, err := os.Open(*sourcePath)
	if err != nil {
		panicf("failed to open source file %q: %v", *sourcePath, err)
	}
	defer srcFile.Close()

	stat, err := srcFile.Stat()
	if err != nil {
		panicf("failed to stat source file: %v", err)
	}

	uploadOpts := &storage.UploadOptions{
		FileName:    *objectKey,
		ContentType: "text/plain",
		Bucket:      *bucket,
	}

	fmt.Printf("Uploading %q (%d bytes) to %s...\n", *sourcePath, stat.Size(), *bucket)
	uploaded, err := objectService.Upload(ctx, provider, srcFile, uploadOpts)
	if err != nil {
		panicf("upload failed: %v", err)
	}

	fmt.Printf("  ✓ Uploaded object key: %s\n", uploaded.Key)

	storageFile := &storagetypes.File{
		FileMetadata: storagetypes.FileMetadata{
			Key:         uploaded.Key,
			Bucket:      *bucket,
			ContentType: uploaded.ContentType,
			Size:        uploaded.Size,
		},
	}

	fmt.Printf("Downloading object to %q...\n", *downloadPath)
	downloaded, err := objectService.Download(ctx, provider, storageFile, &storage.DownloadOptions{})
	if err != nil {
		panicf("download failed: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(*downloadPath), 0o755); err != nil {
		panicf("failed to create destination directory: %v", err)
	}

	if err := os.WriteFile(*downloadPath, downloaded.File, 0o644); err != nil {
		panicf("failed to write downloaded file: %v", err)
	}

	fmt.Printf("  ✓ Wrote %d bytes to %s\n", len(downloaded.File), *downloadPath)

	fmt.Println("Cleaning up remote object...")
	if err := objectService.Delete(ctx, provider, storageFile, &storagetypes.DeleteFileOptions{}); err != nil {
		panicf("delete failed: %v", err)
	}
	fmt.Println("  ✓ Object deleted")

	fmt.Println("Example completed successfully")
}

func newS3Client(ctx context.Context, endpoint, accessKey, secretKey, region string) (*s3.Client, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint cannot be empty")
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse endpoint: %w", err)
	}

	options := []func(*s3.Options){
		func(o *s3.Options) {
			o.UsePathStyle = true
			o.BaseEndpoint = aws.String(endpoint)
		},
	}

	if parsed.Scheme == "http" {
		options = append(options, func(o *s3.Options) {
			o.EndpointOptions.DisableHTTPS = true
		})
	}

	return s3.NewFromConfig(cfg, options...), nil
}

func ensureBucket(ctx context.Context, client *s3.Client, bucket string) error {
	if bucket == "" {
		return fmt.Errorf("bucket name cannot be empty")
	}

	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)})
	if err == nil {
		return nil
	}

	var nfe *types.NotFound
	if !errors.As(err, &nfe) && !isNotFoundError(err) {
		return fmt.Errorf("head bucket: %w", err)
	}

	_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})
	if err != nil {
		var owned *types.BucketAlreadyOwnedByYou
		var exists *types.BucketAlreadyExists
		if errors.As(err, &owned) || errors.As(err, &exists) {
			return nil
		}

		if strings.Contains(err.Error(), "BucketAlreadyExists") || strings.Contains(err.Error(), "BucketAlreadyOwnedByYou") {
			return nil
		}

		return fmt.Errorf("create bucket: %w", err)
	}

	return nil
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "NotFound") || strings.Contains(msg, "404")
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func panicf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
