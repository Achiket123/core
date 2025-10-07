//go:build examples
// +build examples

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/core/pkg/objects/storage/providers/disk"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

func main() {
	ctx := context.Background()

	// Create a temporary directory for storage
	const dirPermissions = 0o755
	storageDir := "./tmp/storage"
	if err := os.MkdirAll(storageDir, dirPermissions); err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll("./tmp")

	// Initialize disk provider
	providerOptions := storage.NewProviderOptions(
		storage.WithBucket(storageDir),
		storage.WithLocalURL("http://localhost:8080/files"),
	)

	provider, err := disk.NewDiskProvider(providerOptions)
	if err != nil {
		log.Fatalf("failed to create provider: %v", err)
	}
	defer provider.Close()

	fmt.Println("=== Simple Object Storage Example ===")

	// Create object service
	service := storage.NewObjectService()

	// Example 1: Upload a file
	fmt.Println("1. Uploading file...")
	content := strings.NewReader("Hello, World! This is a test file.")

	uploadOpts := &storage.UploadOptions{
		FileName:    "hello.txt",
		ContentType: "text/plain",
	}

	uploadedFile, err := service.Upload(ctx, provider, content, uploadOpts)
	if err != nil {
		log.Fatalf("upload failed: %v", err)
	}
	fmt.Printf("   ✓ Uploaded: %s (size: %d bytes)\n", uploadedFile.Key, uploadedFile.Size)

	// Example 2: Download the file
	fmt.Println("\n2. Downloading file...")
	storageFile := &storagetypes.File{
		FileMetadata: storagetypes.FileMetadata{
			Key:         uploadedFile.Key,
			Size:        uploadedFile.Size,
			ContentType: uploadedFile.ContentType,
		},
	}

	downloaded, err := service.Download(ctx, provider, storageFile, &storage.DownloadOptions{})
	if err != nil {
		log.Fatalf("download failed: %v", err)
	}
	fmt.Printf("   ✓ Downloaded: %s (%d bytes)\n", uploadedFile.Key, len(downloaded.File))
	fmt.Printf("   Content: %s\n", string(downloaded.File))

	// Example 3: Get presigned URL
	fmt.Println("\n3. Getting presigned URL...")
	const presignedURLDuration = 15
	url, err := service.GetPresignedURL(ctx, provider, storageFile, &storagetypes.PresignedURLOptions{
		Duration: presignedURLDuration * time.Minute,
	})
	if err != nil {
		log.Fatalf("failed to get URL: %v", err)
	}
	fmt.Printf("   ✓ URL: %s\n", url)

	// Example 4: Check if file exists
	fmt.Println("\n4. Checking file existence...")
	exists, err := provider.Exists(ctx, storageFile)
	if err != nil {
		log.Fatalf("exists check failed: %v", err)
	}
	fmt.Printf("   ✓ File exists: %v\n", exists)

	// Example 5: Delete the file
	fmt.Println("\n5. Deleting file...")
	if err := service.Delete(ctx, provider, storageFile, &storagetypes.DeleteFileOptions{}); err != nil {
		log.Fatalf("delete failed: %v", err)
	}
	fmt.Println("   ✓ File deleted successfully")

	// Verify deletion
	exists, _ = provider.Exists(ctx, storageFile)
	fmt.Printf("   ✓ File exists after deletion: %v\n", exists)

	fmt.Println("\n=== Example completed successfully ===")
}
