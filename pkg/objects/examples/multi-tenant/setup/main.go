//go:build examples
// +build examples

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

type tenant struct {
	id        int
	username  string
	password  string
	bucket    string
	accessKey string
	secretKey string
}

const (
	defaultNumTenants = 10
	defaultParallel   = 10
)

var (
	numTenants = flag.Int("tenants", defaultNumTenants, "Number of tenants to create (default: 10, use 1000 for benchmarks)")
	parallel   = flag.Int("parallel", defaultParallel, "Number of parallel workers for tenant creation")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	fmt.Printf("=== Multi-Tenant Setup (%d tenants) ===\n\n", *numTenants)

	// Start Docker services
	if err := startDockerServices(ctx); err != nil {
		log.Fatalf("failed to start services: %v", err)
	}

	// Wait for MinIO
	if err := waitForMinIO(ctx); err != nil {
		log.Fatalf("failed to wait for MinIO: %v", err)
	}

	// Generate tenants
	tenants := generateTenants(*numTenants)

	// Create tenants in parallel
	start := time.Now()
	if err := createTenants(ctx, tenants, *parallel); err != nil {
		log.Fatalf("failed to create tenants: %v", err)
	}
	elapsed := time.Since(start)

	// Save tenant configuration
	if err := saveTenantConfig(tenants); err != nil {
		log.Fatalf("failed to save config: %v", err)
	}

	fmt.Printf("\n=== Setup Complete ===\n")
	fmt.Printf("Created %d tenants in %v\n", *numTenants, elapsed)
	fmt.Printf("Average: %v per tenant\n", elapsed/time.Duration(*numTenants))
	fmt.Printf("\nMinIO Console: http://localhost:19001 (admin/adminsecretpassword)\n")
	fmt.Printf("MinIO API: http://localhost:19000\n")
	fmt.Printf("\nTenant config saved to: tenants.json\n")
}

func startDockerServices(ctx context.Context) error {
	fmt.Println("Starting Docker services...")
	cmd := exec.CommandContext(ctx, "docker-compose", "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func waitForMinIO(ctx context.Context) error {
	fmt.Println("Waiting for MinIO to be ready...")
	for i := 0; i < 60; i++ {
		cmd := exec.CommandContext(ctx, "docker", "exec", "multitenant-minio",
			"mc", "alias", "set", "local", "http://localhost:9000", "admin", "adminsecretpassword")
		cmd.Stderr = nil
		if err := cmd.Run(); err == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("timeout waiting for MinIO")
}

func generateTenants(count int) []tenant {
	tenants := make([]tenant, count)
	for i := 0; i < count; i++ {
		tenants[i] = tenant{
			id:        i,
			username:  fmt.Sprintf("tenant%04d", i),
			password:  fmt.Sprintf("tenant%04dsecret", i),
			bucket:    fmt.Sprintf("tenant%04d-bucket", i),
			accessKey: fmt.Sprintf("tenant%04d", i),
			secretKey: fmt.Sprintf("tenant%04dsecret", i),
		}
	}
	return tenants
}

func createTenants(ctx context.Context, tenants []tenant, parallel int) error {
	fmt.Printf("Creating %d tenants with %d parallel workers...\n", len(tenants), parallel)

	var wg sync.WaitGroup
	tenantChan := make(chan tenant, len(tenants))
	errChan := make(chan error, len(tenants))

	// Start workers
	for i := 0; i < parallel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range tenantChan {
				if err := createTenant(ctx, t); err != nil {
					errChan <- fmt.Errorf("tenant %d: %w", t.id, err)
					return
				}
			}
		}()
	}

	// Send tenants to workers
	for _, t := range tenants {
		tenantChan <- t
	}
	close(tenantChan)

	// Wait for completion
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		return err
	}

	return nil
}

func createTenant(ctx context.Context, t tenant) error {
	// Create user
	cmd := exec.CommandContext(ctx, "docker", "exec", "multitenant-minio",
		"mc", "admin", "user", "add", "local", t.username, t.password)
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Create bucket
	cmd = exec.CommandContext(ctx, "docker", "exec", "multitenant-minio",
		"mc", "mb", fmt.Sprintf("local/%s", t.bucket), "--ignore-existing")
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	// Create policy
	policy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect": "Allow",
				"Action": []string{"s3:*"},
				"Resource": []string{
					fmt.Sprintf("arn:aws:s3:::%s", t.bucket),
					fmt.Sprintf("arn:aws:s3:::%s/*", t.bucket),
				},
			},
		},
	}

	policyJSON, _ := json.Marshal(policy)
	policyName := fmt.Sprintf("%s-policy", t.username)

	cmd = exec.CommandContext(ctx, "docker", "exec", "-i", "multitenant-minio",
		"mc", "admin", "policy", "create", "local", policyName)
	cmd.Stdin = bytes.NewReader(policyJSON)
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create policy: %w", err)
	}

	// Attach policy
	cmd = exec.CommandContext(ctx, "docker", "exec", "multitenant-minio",
		"mc", "admin", "policy", "attach", "local", policyName, "--user", t.username)
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to attach policy: %w", err)
	}

	if (t.id+1)%100 == 0 {
		fmt.Printf("  Progress: %d/%d tenants created\n", t.id+1, *numTenants)
	}

	return nil
}

func saveTenantConfig(tenants []tenant) error {
	f, err := os.Create("tenants.json")
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(tenants)
}
