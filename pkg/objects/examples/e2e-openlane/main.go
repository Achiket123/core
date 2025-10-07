//go:build examples
// +build examples

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var (
	apiURL       = flag.String("api", "http://localhost:17608", "Openlane API URL")
	authToken    = flag.String("token", "", "Authentication token")
	evidenceName = flag.String("name", "Security Compliance Evidence", "Evidence name")
	description  = flag.String("description", "Security compliance validation evidence", "Evidence description")
	evidenceFile = flag.String("file", "", "Path to evidence file")
	verbose      = flag.Bool("v", false, "Verbose output")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	fmt.Println("=== Openlane End-to-End Integration Example ===")
	fmt.Println()

	// Get auth token from env if not provided
	token := *authToken
	if token == "" {
		token = os.Getenv("OPENLANE_AUTH_TOKEN")
	}
	if token == "" {
		log.Fatal("Authentication token required. Set -token flag or OPENLANE_AUTH_TOKEN env var")
	}

	// Parse API URL
	baseURL, err := url.Parse(*apiURL)
	if err != nil {
		log.Fatalf("invalid API URL: %v", err)
	}

	fmt.Printf("API Endpoint: %s\n", baseURL.String())
	fmt.Printf("Evidence Name: %s\n", *evidenceName)
	fmt.Printf("Description: %s\n", *description)
	fmt.Println()

	// Initialize Openlane client
	client, err := initializeClient(baseURL, token)
	if err != nil {
		log.Fatalf("failed to initialize client: %v", err)
	}

	// Determine evidence file
	var filePath string
	if *evidenceFile != "" {
		filePath = *evidenceFile
	} else {
		// Use a sample file from testdata
		filePath = "../multi-tenant/testdata/sample-files/document.json"
	}

	if *verbose {
		fmt.Printf("Evidence file: %s\n", filePath)
	}

	// Load file for upload
	upload, err := createUpload(filePath)
	if err != nil {
		log.Fatalf("failed to load evidence file: %v", err)
	}

	// Create evidence with file attachment
	fmt.Println("Creating evidence with file attachment...")
	evidence, err := createEvidenceWithFile(ctx, client, *evidenceName, *description, upload)
	if err != nil {
		log.Fatalf("failed to create evidence: %v", err)
	}

	fmt.Printf("  ✓ Evidence created: %s\n", evidence.CreateEvidence.Evidence.ID)
	fmt.Printf("  ✓ Name: %s\n", evidence.CreateEvidence.Evidence.Name)
	if evidence.CreateEvidence.Evidence.Status != nil {
		fmt.Printf("  ✓ Status: %s\n", *evidence.CreateEvidence.Evidence.Status)
	}

	// Display file attachments
	if len(evidence.CreateEvidence.Evidence.Files.Edges) > 0 {
		fmt.Printf("\n  Attachments (%d):\n", len(evidence.CreateEvidence.Evidence.Files.Edges))
		for _, edge := range evidence.CreateEvidence.Evidence.Files.Edges {
			if edge.Node != nil {
				fmt.Printf("    - File ID: %s\n", edge.Node.ID)
				if edge.Node.PresignedURL != nil && *edge.Node.PresignedURL != "" {
					fmt.Printf("      Presigned URL: %s\n", *edge.Node.PresignedURL)
				}
			}
		}
	}

	fmt.Println("\n=== Integration test completed successfully ===")
}

func initializeClient(baseURL *url.URL, token string) (*openlaneclient.OpenlaneClient, error) {
	config := openlaneclient.NewDefaultConfig()

	// Load org ID for PAT header
	orgID, _ := os.ReadFile(".benchmark-org-id")
	if len(orgID) > 0 {
		config.Interceptors = append(config.Interceptors, openlaneclient.WithOrganizationHeader(strings.TrimSpace(string(orgID))))
	}

	client, err := openlaneclient.New(
		config,
		openlaneclient.WithBaseURL(baseURL),
		openlaneclient.WithCredentials(openlaneclient.Authorization{
			BearerToken: token,
		}),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func createUpload(filePath string) (*graphql.Upload, error) {
	uploadFile, err := storage.NewUploadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload file: %w", err)
	}

	return &graphql.Upload{
		File:        uploadFile.RawFile,
		Filename:    uploadFile.OriginalName,
		Size:        uploadFile.Size,
		ContentType: uploadFile.ContentType,
	}, nil
}

func createEvidenceWithFile(ctx context.Context, client *openlaneclient.OpenlaneClient, name, desc string, upload *graphql.Upload) (*openlaneclient.CreateEvidence, error) {
	status := enums.EvidenceSubmitted

	input := openlaneclient.CreateEvidenceInput{
		Name:        name,
		Description: &desc,
		Status:      &status,
	}

	return client.CreateEvidence(ctx, input, []*graphql.Upload{upload})
}
