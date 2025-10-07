# Multi-Tenant High-Throughput Object Storage Example

This example demonstrates high-throughput, multi-tenant object storage with comprehensive benchmarking and performance profiling.

## Features

- **Scalable Multi-Tenancy**: Support for 1000+ isolated tenants
- **Unique Credentials**: Each tenant has isolated credentials and buckets
- **Client Pooling**: Efficient connection pooling with `pkg/cp`
- **High Throughput**: Concurrent operations across all tenants
- **Performance Metrics**: Cache hit rates, memory footprint, operations/second
- **Comprehensive Benchmarks**: Multiple benchmark scenarios
- **CPU & Memory Profiling**: Built-in profiling support

## Prerequisites

- Docker and Docker Compose
- Go 1.21+
- Task (optional, but recommended)

## Quick Start

### 1. Setup Infrastructure

**Small setup (10 tenants):**
```bash
task setup
```

**Large setup (1000 tenants):**
```bash
task setup-1000
```

This creates:
- 1 MinIO instance
- N unique users with individual credentials
- N unique buckets (one per tenant)
- IAM policies for each user
- Saves configuration to `tenants.json`

### 2. Run Examples

**Small load test:**
```bash
task run-small
# 100 operations per tenant, 10 concurrent workers
```

**Medium load test:**
```bash
task run-medium
# 500 operations per tenant, 50 concurrent workers
```

**Large load test:**
```bash
task run-large
# 1000 operations per tenant, 100 concurrent workers
```

**Custom configuration:**
```bash
task run -- -ops=250 -concurrent=25
```

### 3. Run Benchmarks

**All benchmarks:**
```bash
task benchmark
```

**Specific benchmarks:**
```bash
task benchmark-pooling      # Client pool performance
task benchmark-multitenant  # Multi-tenant concurrency
task benchmark-memory       # Memory allocation patterns
task benchmark-resolver     # Resolver with 1000 tenants
```

### 4. Profiling

**CPU profiling:**
```bash
task profile-cpu
# Opens pprof web interface at http://localhost:8080
```

**Memory profiling:**
```bash
task profile-mem
# Opens pprof web interface at http://localhost:8080
```

### 5. Cleanup

```bash
task clean
```

## Architecture

```
┌─────────────────────────────────────────────┐
│           Application Layer                  │
│  (Multi-tenant operations manager)           │
└────────────────┬────────────────────────────┘
                 │
┌────────────────▼────────────────────────────┐
│          CP Service Layer                    │
│  • Client Pool (30min TTL)                   │
│  • Dynamic Builder Registration              │
│  • Context-based Resolution                  │
└────────────────┬────────────────────────────┘
                 │
        ┌────────▼──────────┐
        │   Resolver        │
        │  (1000 rules)     │
        └────────┬──────────┘
                 │
    ┌────────────▼─────────────────┐
    │   Per-Tenant Providers        │
    │  1000 S3 Provider Instances   │
    │  Each with:                   │
    │  • Unique credentials         │
    │  • Isolated bucket            │
    │  • Dedicated IAM policy       │
    └───────────────────────────────┘
```

## Performance Characteristics

### With 1000 Tenants

Based on typical benchmark runs:

**Client Pool Performance:**
- Cache Hit Rate: ~99.9% after warm-up
- Client Resolution: <100ns per lookup
- Memory Overhead: ~2-5MB per 1000 clients
- Concurrent Access: Fully thread-safe

**Operation Throughput:**
- Upload/Download: 500-2000 ops/sec (depends on file size and concurrency)
- Resolver Performance: 10M+ resolutions/sec
- Pool Access: 50M+ gets/sec (cached)

**Memory Footprint:**
- Base: ~50MB
- Per 100 Tenants: ~10-20MB
- 1000 Tenants: ~150-200MB total

## Example Output

```
=== Multi-Tenant High-Throughput Example ===

Loaded 1000 tenants
Registering tenant providers...
  Registered 100/1000 providers
  Registered 200/1000 providers
  ...
  Registered 1000 providers

Running 100000 operations across 1000 tenants with 100 workers...

=== Results ===

Tenants: 1000
Total Operations: 200000
  Uploads: 100000
  Downloads: 100000
  Errors: 0

Cache Statistics:
  Hits: 1000
  Misses: 0
  Hit Rate: 100.00%

Performance:
  Total Time: 2m15s
  Operations/sec: 1481.48
  Avg Time/op: 675µs

Memory:
  Start Alloc: 45.23 MB
  End Alloc: 187.45 MB
  Delta: 142.22 MB
  Sys: 256.78 MB
  NumGC: 42

=== Example completed successfully ===
```

## Benchmark Examples

```bash
$ task benchmark-pooling

BenchmarkClientPooling-12    	50000000	        23.4 ns/op	       0 B/op	       0 allocs/op
```

```bash
$ task benchmark-multitenant

BenchmarkMultiTenantConcurrency-12    	10000000	       187 ns/op	      48 B/op	       2 allocs/op
```

```bash
$ task benchmark-resolver

BenchmarkResolverPerformance-12    	15000000	        89.2 ns/op	       0 B/op	       0 allocs/op
```

## Configuration

### Setup Options

```bash
go run ./setup/main.go -h

Flags:
  -tenants int
        Number of tenants to create (default: 10, use 1000 for benchmarks)
  -parallel int
        Number of parallel workers for tenant creation (default: 10)
```

### Runtime Options

```bash
go run main.go -h

Flags:
  -ops int
        Number of operations per tenant (default: 100)
  -concurrent int
        Number of concurrent workers (default: 10)
  -tenants string
        Tenant configuration file (default: "tenants.json")
```

## Key Concepts

### Client Pooling

Clients are cached based on:
- **TenantID**: Isolates clients per tenant
- **IntegrationType**: Provider type (s3, disk, gcs)
- **TTL**: 30 minutes default

### Dynamic Builder Registration

Each tenant gets a dedicated builder:
```go
providerType := cp.ProviderType(fmt.Sprintf("s3-tenant-%d", tenantID))
service.RegisterBuilder(providerType, &s3Builder{
    accessKey: tenant.AccessKey,
    secretKey: tenant.SecretKey,
    bucket:    tenant.Bucket,
})
```

### Context-Based Resolution

Resolution uses context values to route to the correct tenant:
```go
ctx = context.WithValue(ctx, tenantIDKey, tenantID)
resolution, err := resolver.Resolve(ctx)
```

## Testing Scenarios

### Scenario 1: Cache Efficiency

Test with many tenants, few operations each:
```bash
task setup-1000
go run main.go -ops=10 -concurrent=100
```
Expected: High cache hit rate, fast resolution

### Scenario 2: High Throughput

Test with few tenants, many operations:
```bash
task setup  # 10 tenants
go run main.go -ops=10000 -concurrent=50
```
Expected: High ops/sec, efficient memory usage

### Scenario 3: Memory Pressure

Test memory management with many concurrent operations:
```bash
task setup-1000
go run main.go -ops=1000 -concurrent=200
```
Expected: Efficient GC, stable memory growth

## Troubleshooting

**Setup timeout:**
- Increase timeout in `setup/main.go` if creating many tenants
- Reduce `-parallel` flag if hitting rate limits

**Benchmark failures:**
- Ensure MinIO is running: `task status`
- Check logs: `task logs`

**Out of memory:**
- Reduce `-concurrent` flag
- Reduce number of tenants
- Reduce `-ops` per tenant

## Services

- **MinIO Console**: http://localhost:19001
  - Username: admin
  - Password: adminsecretpassword

- **MinIO API**: http://localhost:19000

## Files Generated

- `tenants.json`: Tenant configuration
- `cpu.prof`: CPU profile (from `task profile-cpu`)
- `mem.prof`: Memory profile (from `task profile-mem`)

## Notes

- Each tenant is completely isolated
- No credential sharing between tenants
- Client connections are reused via pooling
- All operations are thread-safe
- Benchmarks require running setup first
