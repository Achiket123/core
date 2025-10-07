package disk

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	storage "github.com/theopenlane/core/pkg/objects/storage"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

const (
	// DefaultDirPermissions defines the default permissions for created directories
	DefaultDirPermissions = 0755
	// DefaultFilePermissions defines the default permissions for created files
	DefaultFilePermissions = 0644
)

// Provider implements the storagetypes.Provider interface for local filesystem storage
type Provider struct {
	options           *storage.ProviderOptions
	Scheme            string
	destinationFolder string
}

// NewDiskProvider creates a new disk provider instance
func NewDiskProvider(options *storage.ProviderOptions) (*Provider, error) {
	if options == nil || lo.IsEmpty(options.Bucket) {
		return nil, ErrInvalidFolderPath
	}

	disk := &Provider{
		options: options.Clone(),
		Scheme:  "file://",
	}

	if _, err := disk.ListBuckets(); os.IsNotExist(err) {
		log.Info().Str("folder", options.Bucket).Msg("directory does not exist, creating directory")

		if err := os.MkdirAll(options.Bucket, os.ModePerm); err != nil {
			return nil, fmt.Errorf("%w: failed to create directory", ErrInvalidFolderPath)
		}
	}

	disk.destinationFolder = options.Bucket

	return disk, nil
}

// Upload implements storagetypes.Provider
func (p *Provider) Upload(_ context.Context, reader io.Reader, opts *storagetypes.UploadFileOptions) (*storagetypes.UploadedFileMetadata, error) {
	f, err := os.Create(filepath.Join(p.options.Bucket, opts.FileName))
	if err != nil {
		return nil, err
	}

	defer f.Close()

	n, err := io.Copy(f, reader)

	return &storagetypes.UploadedFileMetadata{
		FileMetadata: storagetypes.FileMetadata{
			Key:    opts.FileName,
			Size:   n,
			Folder: opts.FolderDestination,
		},
	}, err
}

// Download implements storagetypes.Provider
func (p *Provider) Download(_ context.Context, file *storagetypes.File, _ *storagetypes.DownloadFileOptions) (*storagetypes.DownloadedFileMetadata, error) {
	filePath := filepath.Join(p.options.Bucket, file.Key)
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return &storagetypes.DownloadedFileMetadata{
		File: fileData,
		Size: int64(len(fileData)),
	}, nil
}

// Delete implements storagetypes.Provider
func (p *Provider) Delete(_ context.Context, file *storagetypes.File, _ *storagetypes.DeleteFileOptions) error {
	err := os.Remove(filepath.Join(p.options.Bucket, file.Key))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// GetPresignedURL implements storagetypes.Provider
func (p *Provider) GetPresignedURL(_ context.Context, file *storagetypes.File, _ *storagetypes.PresignedURLOptions) (string, error) {
	if p.options.LocalURL == "" {
		return "", ErrMissingLocalURL
	}

	base := strings.TrimRight(p.options.LocalURL, "/")

	return fmt.Sprintf("%s/%s", base, file.Key), nil
}

// Exists checks if a file exists on disk
func (p *Provider) Exists(_ context.Context, file *storagetypes.File) (bool, error) {
	fullPath := filepath.Join(p.options.Bucket, file.Key)

	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("%w %s: %w", ErrDiskCheckExists, fullPath, err)
	}

	return true, nil
}

// GetScheme returns the URI scheme for disk
func (p *Provider) GetScheme() *string {
	scheme := "file://"
	return &scheme
}

// Close cleans up resources
func (p *Provider) Close() error {
	return nil
}

// ListBuckets lists the local bucket if it exists
func (p *Provider) ListBuckets() ([]string, error) {
	if _, err := os.Stat(p.options.Bucket); err != nil {
		return nil, err
	}

	return []string{p.options.Bucket}, nil
}

func (p *Provider) ProviderType() storagetypes.ProviderType {
	return storagetypes.DiskProvider
}
