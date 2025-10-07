//go:build examples
// +build examples

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type minioUser struct {
	username string
	password string
	bucket   string
}

const newline = "\n"

func main() {
	ctx := context.Background()

	fmt.Printf("=== Multi-Provider Example Setup ===%s", newline)

	// Start Docker services
	if err := startDockerServices(ctx); err != nil {
		log.Fatalf("failed to start services: %v", err)
	}

	// Wait for services
	if err := waitForServices(ctx); err != nil {
		log.Fatalf("services failed to start: %v", err)
	}

	// Setup MinIO users
	users := []minioUser{
		{username: "provider1", password: "provider1secret", bucket: "provider1-bucket"},
		{username: "provider2", password: "provider2secret", bucket: "provider2-bucket"},
		{username: "provider3", password: "provider3secret", bucket: "provider3-bucket"},
	}

	if err := setupMinIOUsers(ctx, users); err != nil {
		log.Fatalf("failed to setup MinIO: %v", err)
	}

	// Setup GCS buckets
	if err := setupGCSBuckets(ctx); err != nil {
		log.Fatalf("failed to setup GCS: %v", err)
	}

	fmt.Printf("%s=== Setup Complete ===%s", newline, newline)
	fmt.Println("MinIO Console: http://localhost:9001 (admin/adminpassword)")
	fmt.Println("MinIO API: http://localhost:9000")
	fmt.Println("GCS API: http://localhost:4443")
	fmt.Printf("%sCreated users:%s", newline, newline)
	for _, u := range users {
		fmt.Printf("  - %s / %s (bucket: %s)\n", u.username, u.password, u.bucket)
	}
}

func startDockerServices(ctx context.Context) error {
	fmt.Println("Starting Docker services...")
	cmd := exec.CommandContext(ctx, "docker-compose", "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func waitForServices(ctx context.Context) error {
	fmt.Println("Waiting for MinIO...")
	if err := waitForMinIO(ctx); err != nil {
		return err
	}

	fmt.Println("Waiting for GCS...")
	if err := waitForGCS(ctx); err != nil {
		return err
	}

	return nil
}

func waitForMinIO(ctx context.Context) error {
	for i := 0; i < 30; i++ {
		cmd := exec.CommandContext(ctx, "docker", "exec", "objects-example-minio",
			"mc", "alias", "set", "local", "http://localhost:9000", "admin", "adminpassword")
		if err := cmd.Run(); err == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("timeout waiting for MinIO")
}

func waitForGCS(_ context.Context) error {
	const (
		httpTimeout  = 5
		maxRetries   = 30
		gcsHealthURL = "http://localhost:4443/storage/v1/b"
	)
	client := &http.Client{Timeout: httpTimeout * time.Second}
	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, gcsHealthURL, nil)
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
	return fmt.Errorf("timeout waiting for GCS")
}

func setupMinIOUsers(ctx context.Context, users []minioUser) error {
	fmt.Println("\nCreating MinIO users and buckets...")

	for _, user := range users {
		fmt.Printf("  Setting up user: %s\n", user.username)

		// Create user
		cmd := exec.CommandContext(ctx, "docker", "exec", "objects-example-minio",
			"mc", "admin", "user", "add", "local", user.username, user.password)
		if output, err := cmd.CombinedOutput(); err != nil {
			if !bytes.Contains(output, []byte("already exists")) {
				return fmt.Errorf("failed to create user %s: %w - %s", user.username, err, output)
			}
		}

		// Create bucket
		cmd = exec.CommandContext(ctx, "docker", "exec", "objects-example-minio",
			"mc", "mb", fmt.Sprintf("local/%s", user.bucket), "--ignore-existing")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create bucket %s: %w", user.bucket, err)
		}

		// Attach built-in readwrite policy to user
		if err := attachMinIOPolicy(ctx, user.username); err != nil {
			return err
		}
	}

	return nil
}

func attachMinIOPolicy(ctx context.Context, username string) error {
	cmd := exec.CommandContext(ctx, "docker", "exec", "objects-example-minio",
		"mc", "admin", "policy", "attach", "local", "readwrite", "--user", username)
	if output, err := cmd.CombinedOutput(); err != nil {
		if !bytes.Contains(output, []byte("already mapped")) {
			return fmt.Errorf("failed to attach policy: %w - %s", err, output)
		}
	}

	return nil
}

func setupGCSBuckets(ctx context.Context) error {
	fmt.Println("\nCreating GCS buckets...")

	const httpTimeout = 10
	client := &http.Client{Timeout: httpTimeout * time.Second}

	for i := 1; i <= 3; i++ {
		bucketName := fmt.Sprintf("gcs-provider%d-bucket", i)

		bucketData := map[string]string{
			"name": bucketName,
		}
		jsonData, _ := json.Marshal(bucketData)

		req, err := http.NewRequestWithContext(ctx, "POST",
			"http://localhost:4443/storage/v1/b?project=test-project",
			bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to create GCS bucket %s: %w", bucketName, err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
			return fmt.Errorf("unexpected status creating bucket %s: %d", bucketName, resp.StatusCode)
		}

		fmt.Printf("  Created bucket: %s\n", bucketName)
	}

	return nil
}
