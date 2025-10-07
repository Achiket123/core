# End-to-End Openlane Integration Example

This example demonstrates the complete end-to-end capability of the pkg/objects storage system by:

1. Connecting to the Openlane GraphQL API
2. Creating evidence records with file attachments
3. Uploading files through the storage provider system
4. Demonstrating multi-tenant isolation with unique credentials

## Features

- **Real API Integration**: Connects to actual Openlane server
- **Evidence Creation**: Creates compliance evidence with file attachments
- **Multi-Provider Support**: Demonstrates S3, Disk, and GCS providers
- **Tenant Isolation**: Each tenant uses isolated storage credentials
- **Complete Workflow**: Upload → Attach → Retrieve → Verify

## Prerequisites

- Openlane server running (default: http://localhost:17608)
- Valid authentication credentials
- Storage provider configured (S3/Disk/GCS)

## Configuration

The example uses automated setup - no manual configuration required!

### Storage Configuration

**IMPORTANT**: Configure the Openlane server with storage before running benchmarks.

**Option 1: Disk Storage**
```bash
# In the server terminal/config
export OPENLANE_OBJECTSTORAGE_PROVIDER=disk
export OPENLANE_OBJECTSTORAGE_DEFAULTBUCKET=./tmp/file_uploads

# Create storage directory
mkdir -p ./tmp/file_uploads

# Restart server
```

**Option 2: MinIO/S3**
```bash
# Start MinIO
task setup:storage:minio

# In the server terminal/config
export OPENLANE_OBJECTSTORAGE_PROVIDER=s3
export OPENLANE_OBJECTSTORAGE_DEFAULTBUCKET=evidence-uploads
export OPENLANE_OBJECTSTORAGE_ACCESSKEY=admin
export OPENLANE_OBJECTSTORAGE_SECRETKEY=adminsecretpassword
export OPENLANE_OBJECTSTORAGE_ENDPOINT=http://localhost:19000
export OPENLANE_OBJECTSTORAGE_REGION=us-east-1
export OPENLANE_OBJECTSTORAGE_USEPATHSTYLE=true

# Restart server
```

### Quick Start

```bash
# Build binaries, bootstrap auth, and run once
task run
```

To execute the benchmark suite:

```bash
task benchmark
```

## Usage

```bash
# Run with default settings (uses generated token)
task run

# Provide your own evidence file
task run-with-file FILE=./evidence/screenshot.png

# Manual invocation with generated token
go run -tags examples . \
  -token="$(cat .benchmark-token)" \
  -name="Security Evidence" \
  -description="Security validation"

# Verbose output
go run -tags examples . -token="$(cat .benchmark-token)" -v
```

## Workflow

1. **Initialize Storage**: Set up provider based on configuration
2. **Upload File**: Upload evidence file to storage
3. **Create Evidence**: Create evidence record via GraphQL API
4. **Attach File**: Link uploaded file to evidence
5. **Verify**: Retrieve and verify the evidence with attachments

## Example Output

```
=== Openlane End-to-End Integration Example ===

Storage Provider: S3 (evidence-bucket)
API Endpoint: http://localhost:17608/query

Uploading evidence file...
  ✓ Uploaded: evidence/screenshot-20251004.png (2.3 MB)
  ✓ Storage URL: s3://evidence-bucket/evidence/screenshot-20251004.png

Creating evidence record...
  ✓ Evidence ID: ev_1234567890
  ✓ Control: AC-2 (Access Control)
  ✓ Severity: high

Attaching file to evidence...
  ✓ File attached successfully
  ✓ Evidence updated with 1 attachment(s)

Verifying evidence...
  ✓ Retrieved evidence: ev_1234567890
  ✓ Attachments: 1 file(s)
  ✓ File metadata verified

=== Integration test completed successfully ===
```

## API Operations Demonstrated

### GraphQL Mutations

```graphql
mutation CreateEvidence {
  createEvidence(input: {
    controlID: "AC-2"
    finding: "Unauthorized access detected"
    severity: HIGH
    status: UNDER_REVIEW
  }) {
    evidence {
      id
      controlID
      finding
    }
  }
}

mutation AttachFileToEvidence {
  updateEvidence(id: "ev_123", input: {
    attachments: ["file_456"]
  }) {
    evidence {
      id
      attachments {
        id
        fileName
        fileSize
        contentType
      }
    }
  }
}
```

### Storage Operations

- Upload file with multipart form data
- Generate presigned URL for download
- Store file metadata with evidence
- Support multiple storage backends

## Multi-Tenant Isolation

The example demonstrates how different tenants can use the same API but with isolated storage:

```go
// Tenant 1: Uses S3 with unique credentials
tenant1Client := setupTenant("tenant1", "s3", s3Creds)

// Tenant 2: Uses Disk storage
tenant2Client := setupTenant("tenant2", "disk", diskConfig)

// Tenant 3: Uses GCS
tenant3Client := setupTenant("tenant3", "gcs", gcsCreds)
```

Each tenant's files are completely isolated through:
- Unique storage buckets/folders
- Separate credentials
- Isolated client pool entries
- Context-based provider resolution

## Performance Benchmarks

The example includes comprehensive benchmarks for measuring Openlane API performance:

### Running Benchmarks

```bash
# Run all benchmarks
task benchmark

# Quick benchmark (3s each)
task benchmark-quick

# Specific benchmarks
task benchmark-basic              # Basic evidence creation
task benchmark-with-file          # Evidence with file upload
task benchmark-concurrent         # Concurrent operations
task benchmark-multiple-files     # Multiple file uploads
task benchmark-large-file         # Large file uploads
task benchmark-memory             # Memory allocation
task benchmark-metadata           # Full metadata

# Profiling
task profile-cpu                  # CPU profiling
task profile-mem                  # Memory profiling
```

### Benchmark Descriptions

- **BenchmarkEvidenceCreationBasic**: Tests basic evidence creation without files
- **BenchmarkEvidenceCreationWithFile**: Tests evidence creation with single file upload
- **BenchmarkEvidenceCreationConcurrent**: Tests concurrent evidence creation under load
- **BenchmarkEvidenceCreationWithMultipleFiles**: Tests evidence with 3 file uploads
- **BenchmarkEvidenceCreationLargeFile**: Tests large file upload performance
- **BenchmarkClientInitialization**: Measures client initialization overhead
- **BenchmarkUploadFileCreation**: Measures upload file preparation overhead
- **BenchmarkMemoryAllocation**: Analyzes memory allocation patterns
- **BenchmarkEvidenceCreationWithMetadata**: Tests evidence with full metadata

### Requirements

Benchmarks require:
- Openlane server running at `http://localhost:17608`
- Automated setup via `task setup` (runs automatically with benchmark tasks)

No manual environment variables needed - everything is automated!

### Example Output

```
goos: darwin
goarch: arm64
pkg: github.com/theopenlane/core/pkg/objects/examples/e2e-openlane
cpu: Apple M2 Max
BenchmarkEvidenceCreationBasic-12                     290    12,449,421 ns/op    30822 B/op    551 allocs/op
BenchmarkEvidenceCreationConcurrent-12                831     4,171,284 ns/op    31381 B/op    552 allocs/op
BenchmarkClientInitialization-12                  401,043         9,056 ns/op     2544 B/op     36 allocs/op
BenchmarkUploadFileCreation-12                    284,632        12,452 ns/op     4098 B/op     13 allocs/op
BenchmarkMemoryAllocation-12                          290    12,473,586 ns/op    30761 B/op    550 allocs/op
BenchmarkEvidenceCreationWithMetadata-12              288    12,670,274 ns/op    33595 B/op    599 allocs/op
PASS
```

**Note**: File upload benchmarks require server storage configuration (see Storage Configuration section).

## Testing

```bash
# Run example (auto-setup included)
task run

# Run with verbose output
task run -- -v

# Run with custom file
task run-with-file FILE=path/to/file.pdf

# Build example
task build
```

## Cleanup

```bash
# Remove authentication token
task teardown

# Remove all generated files (tokens, profiles)
task clean
```

## Manual Setup (Optional)

If you prefer manual setup:
```bash
# Run setup once
task setup

# Then run benchmarks
go test -tags examples -bench=. -benchmem .

# Or run the example
go run -tags examples . -token="$(cat .benchmark-token)"
```

## Architecture

```
┌─────────────────────────────────────┐
│      Openlane GraphQL Client        │
│   (Evidence Creation & Management)   │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│        Object Storage Service        │
│     (Upload/Download/Delete)         │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│        CP Client Service             │
│  (Pooling + Resolution + Caching)    │
└──────────────┬──────────────────────┘
               │
    ┌──────────▼─────────────┐
    │   Storage Providers     │
    │  ┌────┐ ┌────┐ ┌────┐ │
    │  │ S3 │ │Disk│ │GCS │ │
    │  └────┘ └────┘ └────┘ │
    └─────────────────────────┘
```

## Notes

- Authentication token must have evidence creation permissions
- Storage provider must be configured and accessible
- File size limits depend on provider configuration
- Supports all file types (documents, images, PDFs, etc.)
