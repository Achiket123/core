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

	fmt.Println("=== Multi-Provider Example Teardown ===")

	fmt.Println("Stopping Docker services...")
	cmd := exec.CommandContext(ctx, "docker-compose", "down", "-v")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("failed to stop services: %v", err)
	}

	fmt.Println("\n=== Teardown Complete ===")
}
