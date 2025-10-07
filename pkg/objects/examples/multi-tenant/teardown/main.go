//go:build examples
// +build examples

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== Multi-Tenant Teardown ===")

	fmt.Println("Stopping Docker services...")
	cmd := exec.CommandContext(ctx, "docker-compose", "down", "-v")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("failed to stop services: %v", err)
	}

	// Clean up tenant config
	if err := os.Remove("tenants.json"); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: failed to remove tenants.json: %v", err)
	}

	fmt.Println("\n=== Teardown Complete ===")
}
