# Object Storage Examples

Comprehensive examples demonstrating the `pkg/objects` storage system with various providers, multi-tenancy, high-throughput scenarios, and real-world API integration.

Each example ships with a local `Taskfile.yml` so you can run them with a single
command. From this directory you can list or forward to individual Taskfiles:

```bash
task -d pkg/objects/examples list
task -d pkg/objects/examples simple:run
```

## Examples

### 1. [Simple](./simple/)

Basic usage of the object storage system with the disk provider.

**What it demonstrates:**
- Creating a storage provider
- Uploading files
- Downloading files
- Getting presigned URLs
- Checking file existence
- Deleting files

**Complexity:** ⭐ Beginner
**Runtime:** <1 second
**Requirements:** None (uses local filesystem)

```bash
cd simple
go run main.go
```

---

### 2. [Simple S3](./simple-s3/)

Minimal S3/MinIO usage with the object storage service. Demonstrates uploading,
downloading, and deleting an object against a live S3-compatible endpoint.

**What it demonstrates:**
- Configuring an S3 provider with `storage.ProviderOptions`
- Creating buckets on demand with the AWS SDK
- Uploading a file and persisting the downloaded copy locally
- Cleaning up remote objects after verification

**Complexity:** ⭐ Beginner
**Runtime:** <1 second
**Requirements:** MinIO or any S3-compatible endpoint

```bash
cd simple-s3
go run .
```

---

### 3. [Multi-Provider](./multi-provider/)

Using multiple storage providers concurrently with client pooling and resolution.

**What it demonstrates:**
- Multiple S3 providers (MinIO) with unique credentials
- Disk provider for local storage
- GCS provider (fake-gcs-server emulation)
- Client pooling with `pkg/cp`
- Context-based provider resolution
- Concurrent operations across providers
- Proper credential isolation

**Complexity:** ⭐⭐ Intermediate
**Runtime:** ~30 seconds
**Requirements:** Docker, Docker Compose

```bash
cd multi-provider
task setup    # Start infrastructure
go run main.go
task teardown
```

**Key Features:**
- 3 isolated S3 providers (MinIO)
- 1 disk provider
- 3 GCS providers (fake-gcs-server)
- Context-based routing
- Client pooling
- Concurrent operations

---

### 4. [Multi-Tenant](./multi-tenant/)

High-throughput multi-tenant storage with comprehensive benchmarking and profiling.

**What it demonstrates:**
- Scalable multi-tenancy (up to 1000+ tenants)
- Unique credentials per tenant
- High-throughput concurrent operations
- Client pool cache hit rates
- Memory footprint analysis
- CPU and memory profiling
- Comprehensive benchmarks

**Complexity:** ⭐⭐⭐ Advanced
**Runtime:** 2-5 minutes (1000 tenants)
**Requirements:** Docker, Docker Compose, Task

```bash
cd multi-tenant

# Small setup (10 tenants)
task setup
task run:small

# Large setup (1000 tenants)
task setup:1000
task run:large

# Benchmarks
task benchmark

# Profiling
task profile:cpu
task profile:mem

task clean
```

**Key Features:**
- 1000+ isolated tenants
- Dynamic provider registration
- Context-based resolution with 1000 rules
- Cache performance metrics
- Memory profiling
- CPU profiling
- Multiple benchmark scenarios

---

### 5. [E2E Openlane](./e2e-openlane/)

Real-world integration with Openlane GraphQL API for evidence management with file attachments.

**What it demonstrates:**
- Openlane GraphQL API client usage
- Evidence creation with file attachments
- Automated authentication flow (register, verify, login, org creation, PAT generation)
- Personal Access Token (PAT) with organization context
- GraphQL file upload mutations
- Presigned URL retrieval
- Comprehensive API performance benchmarks

**Complexity:** ⭐⭐⭐ Advanced
**Runtime:** ~30 seconds (benchmarks: 10-30s per test)
**Requirements:** Openlane server running, optional MinIO for S3 storage

```bash
cd e2e-openlane

# Automatic setup (creates user, org, PAT)
task setup

# Setup storage providers
task setup:storage:disk     # Local disk
task setup:storage:minio    # MinIO S3

# Run example
task run

# Benchmarks
task benchmark               # All benchmarks
task benchmark-with-file     # File upload performance
task benchmark-concurrent    # Concurrent operations

# Profiling
task profile-cpu
task profile-mem

# Cleanup
task teardown
```

**Key Features:**
- Automated authentication workflow
- Evidence creation GraphQL mutations
- File upload through GraphQL
- Storage provider integration
- 9 comprehensive benchmarks
- CPU and memory profiling
- No manual environment variable configuration

**Performance Metrics:**
- Evidence Creation (Basic): ~12-15ms/op
- Evidence Creation (Concurrent): ~4ms/op
- Client Initialization: ~9µs/op

---

## Comparison Matrix

| Feature | Simple | Multi-Provider | Multi-Tenant | E2E Openlane |
|---------|--------|----------------|--------------|--------------|
| **Providers** | 1 (Disk) | 7 (3 S3 + 3 GCS + Disk) | 1000+ (S3) | 1 (Configurable) |
| **Client Pooling** | ❌ | ✅ | ✅ | ✅ |
| **Resolution** | ❌ | ✅ | ✅ | ✅ |
| **Concurrency** | ❌ | ✅ | ✅✅✅ | ✅ |
| **Benchmarks** | ❌ | ❌ | ✅ | ✅ |
| **Profiling** | ❌ | ❌ | ✅ | ✅ |
| **API Integration** | ❌ | ❌ | ❌ | ✅ |
| **GraphQL** | ❌ | ❌ | ❌ | ✅ |
| **Docker Required** | ❌ | ✅ | ✅ | Optional |
| **Setup Time** | None | ~30s | ~2-5min | ~10s |
| **Complexity** | Low | Medium | High | High |

## Learning Path

**New to pkg/objects?**
1. Start with **Simple** to understand basic operations
2. Move to **Multi-Provider** to learn client pooling and resolution
3. Explore **Multi-Tenant** for production-scale scenarios
4. Study **E2E Openlane** for real-world API integration

**Key Concepts by Example:**

### Simple
- `storage.Provider` interface
- Provider initialization
- Basic CRUD operations (Upload, Download, Delete)
- Presigned URL generation
- File existence checks

### Multi-Provider
- `cp.Service` for client pooling
- `cp.Builder` for provider-specific clients
- `cp.Resolver` for routing logic
- Context-based provider selection
- Credential isolation per provider
- Concurrent operations across providers

### Multi-Tenant
- High-throughput operations
- Cache hit rate optimization
- Memory management at scale
- Dynamic provider registration (1000+ providers)
- Performance profiling (CPU and memory)
- Benchmark methodologies
- Context-based tenant isolation

### E2E Openlane
- Openlane GraphQL client usage
- Authentication flows (register, verify, login, PAT)
- Organization context for PAT requests
- GraphQL file upload mutations
- Integration with storage providers
- API performance benchmarking
- Real-world evidence management

## Common Tasks

### Running All Examples

```bash
# Simple
cd simple && go run main.go

# Multi-Provider
cd multi-provider && task setup && task run && task teardown

# Multi-Tenant (small)
cd multi-tenant && task setup && task run:small && task clean

# Multi-Tenant (full)
cd multi-tenant && task setup:1000 && task run:large && task clean
```

### Running All Benchmarks

```bash
cd multi-tenant
task setup:1000
task benchmark
task clean
```

### Profiling

```bash
cd multi-tenant
task setup:1000

# CPU profiling
task profile:cpu

# Memory profiling
task profile:mem

task clean
```

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                    Application Layer                     │
│            (Your code using pkg/objects)                 │
└────────────────────┬────────────────────────────────────┘
                     │
        ┌────────────▼──────────────┐
        │    ObjectService          │
        │  (Upload/Download/etc)    │
        └────────────┬──────────────┘
                     │
        ┌────────────▼──────────────┐
        │    CP Client Service       │
        │  • Pool Management         │
        │  • Builder Registry        │
        │  • Cache (TTL-based)       │
        └────────────┬──────────────┘
                     │
        ┌────────────▼──────────────┐
        │      CP Resolver           │
        │  • Rule Evaluation         │
        │  • Context-based Routing   │
        └────────────┬──────────────┘
                     │
    ┌────────────────▼─────────────────────┐
    │          Provider Layer               │
    │  ┌──────┐  ┌──────┐  ┌──────┐       │
    │  │  S3  │  │ Disk │  │ GCS  │  ...  │
    │  └──────┘  └──────┘  └──────┘       │
    └──────────────────────────────────────┘
```

## Performance Characteristics

**Simple:**
- Latency: <1ms (local disk)
- Throughput: Limited by disk I/O
- Memory: <10MB

**Multi-Provider:**
- Latency: 5-50ms (network + MinIO)
- Throughput: ~100-500 ops/sec
- Memory: ~50-100MB

**Multi-Tenant (1000 tenants):**
- Latency: 50-200ms per operation
- Throughput: 1000-2000 ops/sec total
- Memory: ~150-250MB
- Cache Hit Rate: >99%
- Client Resolution: <100ns

## Requirements

### All Examples
- Go 1.21+

### Multi-Provider & Multi-Tenant
- Docker
- Docker Compose
- Task (recommended: `brew install go-task/tap/go-task`)

## Environment Variables

None required for examples. All configuration is embedded or generated.

## Troubleshooting

### Docker Issues

```bash
# Check if Docker is running
docker info

# Check if containers are running
docker ps

# View logs
docker-compose logs -f
```

### Port Conflicts

**Multi-Provider:**
- MinIO: 9000, 9001
- GCS: 4443

**Multi-Tenant:**
- MinIO: 19000, 19001

Change ports in `docker-compose.yml` if needed.

### Performance Issues

**Low throughput:**
- Increase `-concurrent` flag
- Check Docker resource limits
- Monitor with `docker stats`

**High memory usage:**
- Reduce number of tenants
- Reduce `-ops` per operation
- Check for memory leaks with profiling

## Package Usage Reference

### Basic Storage Operations

```go
import "github.com/theopenlane/core/pkg/objects/storage"

// Create disk provider
provider, err := disk.NewProvider(ctx, &disk.Config{
    Bucket: "./tmp/uploads",
})

// Upload file
err = provider.Upload(ctx, "file.txt", bytes.NewReader(data))

// Download file
reader, err := provider.Download(ctx, "file.txt")
defer reader.Close()

// Presigned URL
url, err := provider.PresignedURL(ctx, "file.txt", time.Hour)

// Check existence
exists, err := provider.Exists(ctx, "file.txt")

// Delete file
err = provider.Delete(ctx, "file.txt")
```

### Client Pooling (pkg/cp)

```go
import "github.com/theopenlane/core/pkg/cp"

// Create service
cpService := cp.NewService()

// Register builder for provider type
cpService.RegisterBuilder("s3-provider", &s3Builder{
    accessKey: "key",
    secretKey: "secret",
    bucket:    "my-bucket",
})

// Get or create client (automatically pooled)
client, err := cpService.GetOrCreateClient(ctx, "tenant-123", "s3-provider")

// Use client for operations
err = client.Upload(ctx, "file.txt", data)
```

### Context-Based Resolution

```go
import "github.com/theopenlane/core/pkg/cp"

// Create resolver
resolver := cp.NewResolver()

// Add resolution rules
resolver.AddRule(func(ctx context.Context) (cp.ProviderType, bool) {
    if tenantID, ok := ctx.Value("tenant-id").(string); ok {
        return cp.ProviderType(fmt.Sprintf("s3-tenant-%s", tenantID)), true
    }
    return "", false
})

// Resolve provider from context
ctx = context.WithValue(ctx, "tenant-id", "tenant-123")
resolution, err := resolver.Resolve(ctx)

// Get client using resolution
client, err := cpService.GetOrCreateClient(ctx, "tenant-123", resolution.ProviderType)
```

### Multi-Tenant Pattern

```go
// Register providers for each tenant
for _, tenant := range tenants {
    providerType := cp.ProviderType(fmt.Sprintf("s3-tenant-%d", tenant.ID))

    cpService.RegisterBuilder(providerType, &s3Builder{
        accessKey: tenant.AccessKey,
        secretKey: tenant.SecretKey,
        bucket:    tenant.Bucket,
    })
}

// Add resolution rule
resolver.AddRule(func(ctx context.Context) (cp.ProviderType, bool) {
    if tenantID, ok := ctx.Value("tenant-id").(int); ok {
        return cp.ProviderType(fmt.Sprintf("s3-tenant-%d", tenantID)), true
    }
    return "", false
})

// Use with tenant context
ctx = context.WithValue(ctx, "tenant-id", 123)
client, err := cpService.GetOrCreateClient(ctx, "tenant-123", resolver)
err = client.Upload(ctx, "file.txt", data)
```

### Openlane GraphQL Integration

```go
import (
    "github.com/theopenlane/core/pkg/openlaneclient"
    "github.com/theopenlane/core/pkg/objects/storage"
    "github.com/99designs/gqlgen/graphql"
)

// Initialize client with organization header
config := openlaneclient.NewDefaultConfig()
config.Interceptors = append(config.Interceptors,
    openlaneclient.WithOrganizationHeader(orgID))

client, err := openlaneclient.New(
    config,
    openlaneclient.WithBaseURL(apiURL),
    openlaneclient.WithCredentials(openlaneclient.Authorization{
        BearerToken: token,
    }),
)

// Create upload file
uploadFile, err := storage.NewUploadFile("path/to/file.pdf")

upload := &graphql.Upload{
    File:        uploadFile.RawFile,
    Filename:    uploadFile.OriginalName,
    Size:        uploadFile.Size,
    ContentType: uploadFile.ContentType,
}

// Create evidence with file
input := openlaneclient.CreateEvidenceInput{
    Name:        "Compliance Evidence",
    Description: &description,
    Status:      &status,
}

evidence, err := client.CreateEvidence(ctx, input, []*graphql.Upload{upload})
```

## End-to-End Testing

Complete integration testing with all storage providers:

```bash
# From core root directory

# Setup e2e environment (MinIO + fake-gcs-server + disk)
task e2e:setup

# Run server with e2e configuration
task e2e:run

# In another terminal, run benchmarks
task e2e:benchmark

# Cleanup
task e2e:teardown
```

This generates `config/config.e2e.yaml` with:
- S3 (MinIO): localhost:9000
- Disk: ./tmp/file_uploads
- GCS (fake-gcs-server): localhost:4443

## Additional Resources

- [pkg/objects Documentation](../../README.md)
- [pkg/cp Documentation](../../../cp/README.md)
- [Openlane Client Documentation](../../../openlaneclient/README.md)
- [Provider Documentation](../../storage/providers/)

## Contributing

When adding new examples:
1. Use Task for orchestration
2. Write setup/teardown in Go
3. Include comprehensive README
4. Add benchmarks for performance examples
5. Document expected output
6. Update this README with comparison matrix entry
