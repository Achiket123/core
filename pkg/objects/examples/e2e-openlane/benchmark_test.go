//go:build examples
// +build examples

package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

const (
	benchmarkAPIURL = "http://localhost:17608"
	tokenFile       = ".benchmark-token"
)

func loadToken() string {
	token, err := os.ReadFile(tokenFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(token))
}

func setupClient(b *testing.B) *openlaneclient.OpenlaneClient {
	b.Helper()

	token := loadToken()
	if token == "" {
		b.Skip("Token not found. Run 'task setup' to generate authentication token")
	}

	baseURL, err := url.Parse(benchmarkAPIURL)
	if err != nil {
		b.Fatalf("invalid API URL: %v", err)
	}

	client, err := initializeClient(baseURL, token)
	if err != nil {
		b.Fatalf("failed to initialize client: %v", err)
	}

	return client
}

func createTestUpload(b *testing.B, content string, filename string) *graphql.Upload {
	b.Helper()

	reader := strings.NewReader(content)

	return &graphql.Upload{
		File:        reader,
		Filename:    filename,
		Size:        int64(len(content)),
		ContentType: "text/plain",
	}
}

// BenchmarkEvidenceCreationBasic benchmarks basic evidence creation without files
func BenchmarkEvidenceCreationBasic(b *testing.B) {
	client := setupClient(b)
	ctx := context.Background()

	status := enums.EvidenceSubmitted

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		input := openlaneclient.CreateEvidenceInput{
			Name:        fmt.Sprintf("Benchmark Evidence %d", i),
			Description: stringPtr("Performance benchmark evidence"),
			Status:      &status,
		}

		_, err := client.CreateEvidence(ctx, input, nil)
		if err != nil {
			b.Fatalf("failed to create evidence: %v", err)
		}
	}
}

// BenchmarkEvidenceCreationWithFile benchmarks evidence creation with file upload
func BenchmarkEvidenceCreationWithFile(b *testing.B) {
	client := setupClient(b)
	ctx := context.Background()

	status := enums.EvidenceSubmitted
	testContent := "This is benchmark test evidence content for performance testing"

	for i := 0; b.Loop(); i++ {
		upload := createTestUpload(b, testContent, fmt.Sprintf("bench-%d.txt", i))

		input := openlaneclient.CreateEvidenceInput{
			Name:        fmt.Sprintf("Benchmark Evidence %d", i),
			Description: stringPtr("Performance benchmark evidence with file"),
			Status:      &status,
		}

		_, err := client.CreateEvidence(ctx, input, []*graphql.Upload{upload})
		if err != nil {
			b.Fatalf("failed to create evidence: %v", err)
		}
	}
}

// BenchmarkEvidenceCreationConcurrent benchmarks concurrent evidence creation
func BenchmarkEvidenceCreationConcurrent(b *testing.B) {
	client := setupClient(b)
	status := enums.EvidenceSubmitted

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		counter := atomic.Int64{}

		for pb.Next() {
			i := counter.Add(1)
			input := openlaneclient.CreateEvidenceInput{
				Name:        fmt.Sprintf("Concurrent Benchmark Evidence %d", i),
				Description: stringPtr("Concurrent performance benchmark evidence"),
				Status:      &status,
			}

			_, err := client.CreateEvidence(ctx, input, nil)
			if err != nil {
				b.Fatalf("failed to create evidence: %v", err)
			}
		}
	})
}

// BenchmarkEvidenceCreationWithMultipleFiles benchmarks evidence creation with multiple file uploads
func BenchmarkEvidenceCreationWithMultipleFiles(b *testing.B) {
	client := setupClient(b)
	ctx := context.Background()

	status := enums.EvidenceSubmitted

	for i := 0; b.Loop(); i++ {
		uploads := []*graphql.Upload{
			createTestUpload(b, "File 1 content for benchmark", fmt.Sprintf("bench-%d-file1.txt", i)),
			createTestUpload(b, "File 2 content for benchmark", fmt.Sprintf("bench-%d-file2.txt", i)),
			createTestUpload(b, "File 3 content for benchmark", fmt.Sprintf("bench-%d-file3.txt", i)),
		}

		input := openlaneclient.CreateEvidenceInput{
			Name:        fmt.Sprintf("Multi-File Benchmark Evidence %d", i),
			Description: stringPtr("Performance benchmark evidence with multiple files"),
			Status:      &status,
		}

		_, err := client.CreateEvidence(ctx, input, uploads)
		if err != nil {
			b.Fatalf("failed to create evidence: %v", err)
		}
	}
}

// BenchmarkEvidenceCreationLargeFile benchmarks evidence creation with a large file
func BenchmarkEvidenceCreationLargeFile(b *testing.B) {
	client := setupClient(b)
	ctx := context.Background()

	status := enums.EvidenceSubmitted
	largeContent := strings.Repeat("Large file content for benchmark testing. ", 1000)

	for i := 0; b.Loop(); i++ {
		upload := createTestUpload(b, largeContent, fmt.Sprintf("large-bench-%d.txt", i))

		input := openlaneclient.CreateEvidenceInput{
			Name:        fmt.Sprintf("Large File Benchmark Evidence %d", i),
			Description: stringPtr("Performance benchmark evidence with large file"),
			Status:      &status,
		}

		_, err := client.CreateEvidence(ctx, input, []*graphql.Upload{upload})
		if err != nil {
			b.Fatalf("failed to create evidence: %v", err)
		}
	}
}

// BenchmarkClientInitialization benchmarks client initialization overhead
func BenchmarkClientInitialization(b *testing.B) {
	token := loadToken()
	if token == "" {
		b.Skip("Token not found. Run 'task setup' to generate authentication token")
	}

	baseURL, _ := url.Parse(benchmarkAPIURL)

	for b.Loop() {
		_, err := initializeClient(baseURL, token)
		if err != nil {
			b.Fatalf("failed to initialize client: %v", err)
		}
	}
}

// BenchmarkUploadFileCreation benchmarks the upload file creation overhead
func BenchmarkUploadFileCreation(b *testing.B) {
	testFiles := []string{
		"../multi-tenant/testdata/sample-files/small.txt",
		"../multi-tenant/testdata/sample-files/document.json",
		"../multi-tenant/testdata/sample-files/report.csv",
	}

	for _, filePath := range testFiles {
		if _, err := os.Stat(filePath); err != nil {
			b.Skipf("test file not found: %s", filePath)
		}
	}

	for i := 0; b.Loop(); i++ {
		fileIdx := i % len(testFiles)
		_, err := storage.NewUploadFile(testFiles[fileIdx])
		if err != nil {
			b.Fatalf("failed to create upload file: %v", err)
		}
	}
}

// BenchmarkMemoryAllocation measures memory allocation patterns for evidence creation
func BenchmarkMemoryAllocation(b *testing.B) {
	client := setupClient(b)
	ctx := context.Background()

	status := enums.EvidenceSubmitted

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		input := openlaneclient.CreateEvidenceInput{
			Name:        fmt.Sprintf("Memory Benchmark Evidence %d", i),
			Description: stringPtr("Memory allocation benchmark evidence"),
			Status:      &status,
		}

		_, err := client.CreateEvidence(ctx, input, nil)
		if err != nil {
			b.Fatalf("failed to create evidence: %v", err)
		}
	}
}

// BenchmarkEvidenceCreationWithMetadata benchmarks evidence creation with full metadata
func BenchmarkEvidenceCreationWithMetadata(b *testing.B) {
	client := setupClient(b)
	ctx := context.Background()

	status := enums.EvidenceSubmitted

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		input := openlaneclient.CreateEvidenceInput{
			Name:                fmt.Sprintf("Metadata Benchmark Evidence %d", i),
			Description:         stringPtr("Evidence with full metadata for performance testing"),
			Status:              &status,
			CollectionProcedure: stringPtr("Automated collection via benchmark"),
			Source:              stringPtr("Benchmark System"),
			IsAutomated:         boolPtr(true),
			Tags:                []string{"benchmark", "performance", "test"},
		}

		_, err := client.CreateEvidence(ctx, input, nil)
		if err != nil {
			b.Fatalf("failed to create evidence: %v", err)
		}
	}
}

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
