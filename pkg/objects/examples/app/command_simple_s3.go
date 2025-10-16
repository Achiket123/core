package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/urfave/cli/v3"

	"github.com/theopenlane/core/pkg/objects/storage"
	s3local "github.com/theopenlane/core/pkg/objects/storage/providers/s3"
)

func simpleS3Command() *cli.Command {
	return &cli.Command{
		Name:  "simple-s3",
		Usage: "Run the S3/MinIO object storage example",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "endpoint", Usage: "S3 or MinIO endpoint URL", Value: envOrDefault("OBJECTS_EXAMPLE_S3_ENDPOINT", envOrDefault("MINIO_ENDPOINT", "http://127.0.0.1:9000"))},
			&cli.StringFlag{Name: "access-key", Usage: "Access key ID", Value: envOrDefault("OBJECTS_EXAMPLE_S3_ACCESS_KEY", envOrDefault("MINIO_ACCESS_KEY", "minioadmin"))},
			&cli.StringFlag{Name: "secret-key", Usage: "Secret access key", Value: envOrDefault("OBJECTS_EXAMPLE_S3_SECRET_KEY", envOrDefault("MINIO_SECRET_KEY", "minioadmin"))},
			&cli.StringFlag{Name: "region", Usage: "AWS region", Value: envOrDefault("OBJECTS_EXAMPLE_S3_REGION", envOrDefault("MINIO_REGION", "us-east-1"))},
			&cli.StringFlag{Name: "bucket", Usage: "Bucket to read/write", Value: envOrDefault("OBJECTS_EXAMPLE_S3_BUCKET", envOrDefault("MINIO_BUCKET", "core-simple-s3"))},
			&cli.StringFlag{Name: "source", Usage: "Path to the file that will be uploaded", Value: envOrDefault("OBJECTS_EXAMPLE_S3_SOURCE", "assets/sample-data.txt")},
			&cli.StringFlag{Name: "object", Usage: "Object key inside the bucket", Value: envOrDefault("OBJECTS_EXAMPLE_S3_OBJECT", "examples/simple-s3/sample-data.txt")},
			&cli.StringFlag{Name: "download", Usage: "Destination path for the downloaded file", Value: envOrDefault("OBJECTS_EXAMPLE_S3_DOWNLOAD", "output/downloaded-sample.txt")},
			&cli.BoolFlag{Name: "path-style", Usage: "Use path-style addressing when talking to the endpoint", Value: envOrDefault("OBJECTS_EXAMPLE_S3_PATH_STYLE", "true") == "true"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg := simpleS3Config{
				Endpoint:   cmd.String("endpoint"),
				AccessKey:  cmd.String("access-key"),
				SecretKey:  cmd.String("secret-key"),
				Region:     cmd.String("region"),
				Bucket:     cmd.String("bucket"),
				SourcePath: cmd.String("source"),
				ObjectKey:  cmd.String("object"),
				Download:   cmd.String("download"),
				PathStyle:  cmd.Bool("path-style"),
			}
			out := cmd.Writer
			if out == nil {
				out = os.Stdout
			}
			return runSimpleS3(ctx, out, cfg)
		},
	}
}

type simpleS3Config struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	Region     string
	Bucket     string
	SourcePath string
	ObjectKey  string
	Download   string
	PathStyle  bool
}

func runSimpleS3(ctx context.Context, out io.Writer, cfg simpleS3Config) error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("endpoint cannot be empty")
	}
	if cfg.Bucket == "" {
		return fmt.Errorf("bucket cannot be empty")
	}

	awsClient, err := newS3Client(ctx, cfg.Endpoint, cfg.AccessKey, cfg.SecretKey, cfg.Region)
	if err != nil {
		return fmt.Errorf("create aws client: %w", err)
	}

	if err := ensureBucket(ctx, awsClient, cfg.Bucket); err != nil {
		return fmt.Errorf("ensure bucket %q: %w", cfg.Bucket, err)
	}

	providerOptions := storage.NewProviderOptions(
		storage.WithBucket(cfg.Bucket),
		storage.WithRegion(cfg.Region),
		storage.WithEndpoint(cfg.Endpoint),
		storage.WithCredentials(storage.ProviderCredentials{
			AccessKeyID:     cfg.AccessKey,
			SecretAccessKey: cfg.SecretKey,
			Endpoint:        cfg.Endpoint,
		}),
	)

	provider, err := s3local.NewS3Provider(providerOptions, s3local.WithUsePathStyle(cfg.PathStyle))
	if err != nil {
		return fmt.Errorf("create s3 provider: %w", err)
	}
	defer provider.Close()

	fmt.Fprintln(out, "=== S3 Object Storage Example ===")

	srcFile, err := os.Open(resolvePath(cfg.SourcePath))
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}

	fileInfo, err := srcFile.Stat()
	if err != nil {
		srcFile.Close()
		return fmt.Errorf("stat source file: %w", err)
	}

	fmt.Fprintf(out, "Uploading %q (%d bytes) to bucket %s...\n", cfg.SourcePath, fileInfo.Size(), cfg.Bucket)

	service := storage.NewObjectService()

	_, err = runLifecycle(ctx, out, service, provider, lifecycleConfig{
		FileName:      cfg.ObjectKey,
		ContentType:   "text/plain",
		Bucket:        cfg.Bucket,
		Reader:        srcFile,
		ProviderLabel: "s3",
		AfterDownload: func(downloaded *storage.DownloadedMetadata) error {
			abs := resolvePath(cfg.Download)
			if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
				return fmt.Errorf("create download directory: %w", err)
			}
			if err := os.WriteFile(abs, downloaded.File, 0o644); err != nil {
				return fmt.Errorf("write downloaded file: %w", err)
			}
			fmt.Fprintf(out, "Wrote %d bytes to %s\n", len(downloaded.File), abs)
			return nil
		},
	})
	if err != nil {
		return err
	}

	fmt.Fprintln(out, "=== Example completed successfully ===")
	return nil
}

func newS3Client(ctx context.Context, endpoint, accessKey, secretKey, region string) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
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
		switch {
		case errors.As(err, &owned), errors.As(err, &exists):
			return nil
		case strings.Contains(err.Error(), "BucketAlreadyExists"), strings.Contains(err.Error(), "BucketAlreadyOwnedByYou"):
			return nil
		default:
			return fmt.Errorf("create bucket: %w", err)
		}
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
