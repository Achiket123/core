//go:build examples
// +build examples

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	models "github.com/theopenlane/core/pkg/openapi"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

const (
	defaultAPIURL = "http://localhost:17608"
	defaultEmail  = "benchmark@theopenlane.io"
	defaultPass   = "SuperSecureBenchmarkPass123!"
	tokenFile     = ".benchmark-token"
	orgIDFile     = ".benchmark-org-id"
)

type registerResponse struct {
	Token string `json:"token"`
}

func main() {
	ctx := context.Background()

	apiURL := getEnvOrDefault("OPENLANE_API_URL", defaultAPIURL)
	email := getEnvOrDefault("BENCHMARK_USER_EMAIL", defaultEmail)
	password := getEnvOrDefault("BENCHMARK_USER_PASSWORD", defaultPass)

	fmt.Println("=== E2E Openlane Benchmark Setup ===")
	fmt.Printf("API: %s\n", apiURL)
	fmt.Printf("User: %s\n", email)

	baseURL, err := url.Parse(apiURL)
	if err != nil {
		log.Fatalf("invalid API URL: %v", err)
	}

	// Step 1: Register and verify user
	fmt.Println("\nStep 1: Registering user...")
	verifyToken, err := registerUser(ctx, baseURL, email, password)
	if err != nil {
		if strings.Contains(err.Error(), "409") {
			fmt.Println("  User already exists, attempting login...")
		} else {
			log.Fatalf("failed to register user: %v", err)
		}
	} else {
		fmt.Println("  ✓ User registered")

		// Verify the user
		fmt.Println("\nStep 2: Verifying user...")
		if err := verifyUser(ctx, baseURL, verifyToken); err != nil {
			log.Fatalf("failed to verify user: %v", err)
		}
		fmt.Println("  ✓ User verified")
	}

	// Step 3: Login to get session
	fmt.Println("\nStep 3: Logging in...")
	client, err := loginUser(ctx, baseURL, email, password)
	if err != nil {
		log.Fatalf("failed to login: %v", err)
	}
	fmt.Println("  ✓ Login successful")

	// Step 4: Create organization
	fmt.Println("\nStep 4: Creating organization...")
	orgID, err := createOrganization(ctx, client)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Println("  Organization already exists")
			orgID, err = getOrganizationID(ctx, client)
			if err != nil {
				log.Fatalf("failed to get organization ID: %v", err)
			}
		} else {
			log.Fatalf("failed to create organization: %v", err)
		}
	}
	fmt.Printf("  ✓ Organization ID: %s\n", orgID)

	// Step 5: Create PAT
	fmt.Println("\nStep 5: Creating Personal Access Token...")
	token, err := createPAT(ctx, client, orgID)
	if err != nil {
		log.Fatalf("failed to create PAT: %v", err)
	}
	fmt.Println("  ✓ PAT created")

	// Step 6: Store credentials
	fmt.Println("\nStep 6: Storing credentials...")
	const filePermissions = 0o600
	if err := os.WriteFile(tokenFile, []byte(token), filePermissions); err != nil {
		log.Fatalf("failed to store token: %v", err)
	}
	if err := os.WriteFile(orgIDFile, []byte(orgID), filePermissions); err != nil {
		log.Fatalf("failed to store org ID: %v", err)
	}
	fmt.Printf("  ✓ Credentials stored\n")

	fmt.Println("\n=== Setup complete ===")
	fmt.Println("You can now run benchmarks with: task benchmark")
}

func registerUser(ctx context.Context, baseURL *url.URL, email, password string) (string, error) {
	registerURL := baseURL.String() + "/v1/register"

	payload := map[string]string{
		"email":     email,
		"password":  password,
		"firstName": "Benchmark",
		"lastName":  "User",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", registerURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("registration failed with status %d", resp.StatusCode)
	}

	var result registerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Token, nil
}

func verifyUser(ctx context.Context, baseURL *url.URL, token string) error {
	verifyURL := fmt.Sprintf("%s/v1/verify?token=%s", baseURL.String(), token)

	req, err := http.NewRequestWithContext(ctx, "GET", verifyURL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("verification failed with status %d", resp.StatusCode)
	}

	return nil
}

func loginUser(ctx context.Context, baseURL *url.URL, email, password string) (*openlaneclient.OpenlaneClient, error) {
	config := openlaneclient.NewDefaultConfig()

	client, err := openlaneclient.New(config, openlaneclient.WithBaseURL(baseURL))
	if err != nil {
		return nil, err
	}

	loginInput := models.LoginRequest{
		Username: email,
		Password: password,
	}

	resp, err := client.Login(ctx, &loginInput)
	if err != nil {
		return nil, err
	}

	// Create authenticated client with session
	session, err := client.GetSessionFromCookieJar()
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	authClient, err := openlaneclient.New(
		config,
		openlaneclient.WithBaseURL(baseURL),
		openlaneclient.WithCredentials(openlaneclient.Authorization{
			BearerToken: resp.AccessToken,
			Session:     session,
		}),
	)
	if err != nil {
		return nil, err
	}

	return authClient, nil
}

func createOrganization(ctx context.Context, client *openlaneclient.OpenlaneClient) (string, error) {
	name := "Benchmark Organization"
	description := "Organization for benchmark testing"

	input := openlaneclient.CreateOrganizationInput{
		Name:        name,
		Description: &description,
	}

	org, err := client.CreateOrganization(ctx, input, nil)
	if err != nil {
		return "", err
	}

	return org.CreateOrganization.Organization.ID, nil
}

func getOrganizationID(ctx context.Context, client *openlaneclient.OpenlaneClient) (string, error) {
	orgs, err := client.GetOrganizations(ctx, nil, nil, nil)
	if err != nil {
		return "", err
	}

	if len(orgs.Organizations.Edges) == 0 {
		return "", fmt.Errorf("no organizations found")
	}

	return orgs.Organizations.Edges[0].Node.ID, nil
}

func createPAT(ctx context.Context, client *openlaneclient.OpenlaneClient, orgID string) (string, error) {
	name := "Benchmark PAT"
	description := "Personal Access Token for benchmark testing"

	input := openlaneclient.CreatePersonalAccessTokenInput{
		Name:            name,
		Description:     &description,
		OrganizationIDs: []string{orgID},
	}

	pat, err := client.CreatePersonalAccessToken(ctx, input)
	if err != nil {
		return "", err
	}

	return pat.CreatePersonalAccessToken.PersonalAccessToken.Token, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
