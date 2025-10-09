package app

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/99designs/gqlgen/graphql"
	"github.com/urfave/cli/v3"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func openlaneCommand() *cli.Command {
	return &cli.Command{
		Name:  "openlane",
		Usage: "Run the Openlane end-to-end integration example",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "api", Usage: "Openlane API base URL", Value: "http://localhost:17608", Sources: cli.EnvVars("OBJECTS_EXAMPLE_OPENLANE_API")},
			&cli.StringFlag{Name: "token", Usage: "Authentication token (fallback to OPENLANE_AUTH_TOKEN)", Sources: cli.EnvVars("OBJECTS_EXAMPLE_OPENLANE_TOKEN")},
			&cli.StringFlag{Name: "name", Usage: "Evidence name", Value: "Security Compliance Evidence", Sources: cli.EnvVars("OBJECTS_EXAMPLE_OPENLANE_NAME")},
			&cli.StringFlag{Name: "description", Usage: "Evidence description", Value: "Security compliance validation evidence", Sources: cli.EnvVars("OBJECTS_EXAMPLE_OPENLANE_DESCRIPTION")},
			&cli.StringFlag{Name: "file", Usage: "Path to evidence file", Sources: cli.EnvVars("OBJECTS_EXAMPLE_OPENLANE_FILE")},
			&cli.BoolFlag{Name: "verbose", Usage: "Enable verbose logging", Sources: cli.EnvVars("OBJECTS_EXAMPLE_OPENLANE_VERBOSE")},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			runCfg := openlaneConfig{
				API:         cmd.String("api"),
				Token:       cmd.String("token"),
				Name:        cmd.String("name"),
				Description: cmd.String("description"),
				FilePath:    cmd.String("file"),
				Verbose:     cmd.Bool("verbose"),
			}
			out := cmd.Writer
			if out == nil {
				out = os.Stdout
			}
			return runOpenlane(ctx, out, runCfg)
		},
	}
}

type openlaneConfig struct {
	API         string
	Token       string
	Name        string
	Description string
	FilePath    string
	Verbose     bool
}

func runOpenlane(ctx context.Context, out io.Writer, cfg openlaneConfig) error {
	fmt.Fprintln(out, "=== Openlane End-to-End Integration Example ===")
	fmt.Fprintln(out)

	token := cfg.Token
	if token == "" {
		token = os.Getenv("OPENLANE_AUTH_TOKEN")
	}
	if token == "" {
		return fmt.Errorf("authentication token required via --token or OPENLANE_AUTH_TOKEN")
	}

	baseURL, err := url.Parse(cfg.API)
	if err != nil {
		return fmt.Errorf("invalid API URL: %w", err)
	}

	fmt.Fprintf(out, "API Endpoint: %s\n", baseURL.String())
	fmt.Fprintf(out, "Evidence Name: %s\n", cfg.Name)
	fmt.Fprintf(out, "Description: %s\n", cfg.Description)
	fmt.Fprintln(out)

	client, err := initializeOpenlaneClient(baseURL, token)
	if err != nil {
		return fmt.Errorf("initialize client: %w", err)
	}

	evidenceFile := cfg.FilePath
	if evidenceFile == "" {
		evidenceFile = defaultEvidenceFile()
	}

	if cfg.Verbose {
		fmt.Fprintf(out, "Evidence file: %s\n", evidenceFile)
	}

	upload, err := createUpload(evidenceFile)
	if err != nil {
		return fmt.Errorf("load evidence file: %w", err)
	}

	fmt.Fprintln(out, "Creating evidence with file attachment...")
	evidence, err := createEvidenceWithFile(ctx, client, cfg.Name, cfg.Description, upload)
	if err != nil {
		return fmt.Errorf("create evidence: %w", err)
	}

	fmt.Fprintf(out, "  ✓ Evidence created: %s\n", evidence.CreateEvidence.Evidence.ID)
	fmt.Fprintf(out, "  ✓ Name: %s\n", evidence.CreateEvidence.Evidence.Name)
	if evidence.CreateEvidence.Evidence.Status != nil {
		fmt.Fprintf(out, "  ✓ Status: %s\n", *evidence.CreateEvidence.Evidence.Status)
	}

	if len(evidence.CreateEvidence.Evidence.Files.Edges) > 0 {
		fmt.Fprintf(out, "\n  Attachments (%d):\n", len(evidence.CreateEvidence.Evidence.Files.Edges))
		for _, edge := range evidence.CreateEvidence.Evidence.Files.Edges {
			if edge.Node != nil {
				fmt.Fprintf(out, "    - File ID: %s\n", edge.Node.ID)
				if edge.Node.PresignedURL != nil && *edge.Node.PresignedURL != "" {
					fmt.Fprintf(out, "      Presigned URL: %s\n", *edge.Node.PresignedURL)
				}
			}
		}
	}

	fmt.Fprintln(out, "\n=== Integration test completed successfully ===")
	return nil
}

func initializeOpenlaneClient(baseURL *url.URL, token string) (*openlaneclient.OpenlaneClient, error) {
	config := openlaneclient.NewDefaultConfig()

	client, err := openlaneclient.New(
		config,
		openlaneclient.WithBaseURL(baseURL),
		openlaneclient.WithCredentials(openlaneclient.Authorization{BearerToken: token}),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func defaultEvidenceFile() string {
	return resolvePath("testdata/sample-files/document.json")
}

func createUpload(filePath string) (*graphql.Upload, error) {
	uploadFile, err := storage.NewUploadFile(resolvePath(filePath))
	if err != nil {
		return nil, fmt.Errorf("create upload file: %w", err)
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
