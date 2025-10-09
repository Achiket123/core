//go:build examples
// +build examples

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

const (
	databaseProviderPathMarker = "/v1/files/"
	databaseProviderTokenParam = "token"
)

func TestDatabaseProviderEvidenceRoundTrip(t *testing.T) {
	authToken := loadToken()
	if authToken == "" {
		t.Skip("Token not found. Run 'task setup' to generate authentication token")
	}

	client := newIntegrationClientWithToken(t, authToken)
	ctx := context.Background()

	status := enums.EvidenceSubmitted
	now := time.Now().UTC()
	timestamp := now.UnixNano()

	fileContent := []byte(fmt.Sprintf("database provider e2e content %d", timestamp))
	fileName := fmt.Sprintf("database-provider-%d.txt", timestamp)

	t.Log("=== Database provider round-trip ===")
	t.Logf("API: %s", benchmarkAPIURL)
	t.Log("1. Uploading evidence + file")

	upload := newUploadFromBytes(fileContent, fileName, "text/plain")
	input := openlaneclient.CreateEvidenceInput{
		Name:        fmt.Sprintf("Database provider evidence %d", timestamp),
		Description: stringPtr("Verifies database-backed object storage end-to-end"),
		Status:      &status,
	}

	result, err := client.CreateEvidence(ctx, input, []*graphql.Upload{upload})
	if err != nil {
		t.Fatalf("failed to create evidence: %v", err)
	}

	evidence := result.GetCreateEvidence().GetEvidence()
	if evidence == nil {
		t.Fatalf("create evidence response missing evidence payload")
	}

	evidenceID := evidence.GetID()
	if evidenceID == "" {
		t.Fatalf("expected evidence ID in response")
	}

	t.Cleanup(func() {
		if _, cleanupErr := client.DeleteEvidence(ctx, evidenceID); cleanupErr != nil {
			t.Logf("warning: failed to delete evidence %s: %v", evidenceID, cleanupErr)
		}
	})

	files := evidence.GetFiles()
	if files == nil {
		t.Fatalf("expected file connection in create evidence response")
	}

	edges := files.GetEdges()
	if len(edges) == 0 {
		t.Fatalf("expected at least one file attachment in create evidence response")
	}

	node := edges[0].GetNode()
	if node == nil {
		t.Fatalf("expected file node in response")
	}

	fileID := node.GetID()
	if fileID == "" {
		t.Fatalf("expected file ID in response")
	}

	t.Logf("   Evidence ID: %s", evidenceID)
	t.Logf("   File ID: %s", fileID)
	t.Logf("   Uploaded content type: %s", upload.ContentType)

	t.Log("2. Fetching presigned download URL")

	presignedURL := node.GetPresignedURL()
	if presignedURL == nil || *presignedURL == "" {
		t.Fatalf("expected presigned URL for file %s", fileID)
	}

	if !strings.Contains(*presignedURL, databaseProviderPathMarker) {
		t.Fatalf("expected database handler presigned URL to contain %q, got %q", databaseProviderPathMarker, *presignedURL)
	}

	parsedURL, err := url.Parse(*presignedURL)
	if err != nil {
		t.Fatalf("failed to parse presigned URL %q: %v", *presignedURL, err)
	}

	if parsedURL.Query().Get(databaseProviderTokenParam) == "" {
		t.Fatalf("presigned URL missing %q query parameter: %q", databaseProviderTokenParam, *presignedURL)
	}

	downloadURL := resolveDownloadURL(t, *presignedURL)
	t.Logf("   Download URL: %s", downloadURL)
	t.Log("3. Downloading file via application proxy")
	body, headers := downloadFileWithRetry(t, downloadURL, authToken, 5, 200*time.Millisecond)

	if !bytes.Equal(body, fileContent) {
		t.Fatalf("downloaded file content mismatch: expected %q got %q", string(fileContent), string(body))
	}

	contentType := headers.Get("Content-Type")
	if contentType == "" || !strings.HasPrefix(contentType, "text/plain") {
		t.Fatalf("unexpected content type header: %q", contentType)
	}

	disposition := headers.Get("Content-Disposition")
	if disposition == "" || !strings.Contains(disposition, fileName) {
		t.Fatalf("content-disposition header %q does not include filename %q", disposition, fileName)
	}

	t.Log("4. Validating headers and content succeeded")
}

func newIntegrationClientWithToken(t testing.TB, token string) *openlaneclient.OpenlaneClient {
	t.Helper()

	baseURL, err := url.Parse(benchmarkAPIURL)
	if err != nil {
		t.Fatalf("invalid API URL %q: %v", benchmarkAPIURL, err)
	}

	client, err := initializeClient(baseURL, token)
	if err != nil {
		t.Fatalf("failed to initialize client: %v", err)
	}

	return client
}

func newUploadFromBytes(content []byte, filename, contentType string) *graphql.Upload {
	reader := bytes.NewReader(content)
	return &graphql.Upload{
		File:        reader,
		Filename:    filename,
		Size:        int64(len(content)),
		ContentType: contentType,
	}
}

func resolveDownloadURL(t testing.TB, raw string) string {
	t.Helper()

	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}

	baseURL, err := url.Parse(benchmarkAPIURL)
	if err != nil {
		t.Fatalf("invalid API base URL %q: %v", benchmarkAPIURL, err)
	}

	relative, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("invalid relative presigned URL %q: %v", raw, err)
	}

	return baseURL.ResolveReference(relative).String()
}

func downloadFileWithRetry(t testing.TB, downloadURL string, token string, attempts int, delay time.Duration) ([]byte, http.Header) {
	t.Helper()

	client := &http.Client{Timeout: 5 * time.Second}
	var lastErr error

	for i := 0; i < attempts; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
		if err != nil {
			cancel()
			t.Fatalf("failed to create download request: %v", err)
		}

		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			cancel()
		} else {
			data, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			cancel()

			if readErr != nil {
				lastErr = readErr
			} else if resp.StatusCode == http.StatusOK {
				return data, resp.Header.Clone()
			} else {
				lastErr = fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(data))
			}
		}

		time.Sleep(delay)
	}

	t.Fatalf("failed to download file from %s after %d attempts: %v", downloadURL, attempts, lastErr)
	return nil, nil
}
