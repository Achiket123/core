# Multi-Provider Object Storage Example

This example demonstrates using multiple storage providers concurrently with client pooling and resolution.

## Features Demonstrated

- Multiple S3 providers (MinIO) with unique users and buckets
- Disk provider for local storage
- Client pooling with `pkg/cp`
- Context-based provider resolution
- Concurrent operations across providers
- Proper credential isolation per provider

## Prerequisites

- Docker and Docker Compose
- Go 1.21+

## Setup & Execution

Use the Taskfile to orchestrate both infrastructure and workflow:

```bash
cd pkg/objects/examples/multi-provider
task run
```

The `run` task will:
- Build the setup, teardown, and main binaries with the `examples` build tag
- Start MinIO and fake-gcs-server via Docker Compose
- Provision provider credentials and buckets
- Execute the example workflow

To perform steps individually:

```bash
task setup     # starts services and provisions credentials
task run       # runs the workflow (implicitly depends on setup)
task teardown  # stops services and removes containers
```

Manual alternative:

```bash
go run -tags examples ./setup
go run -tags examples .
```

## What It Does

1. **Tests Disk Provider** - Local filesystem storage
2. **Tests S3 Provider 1** - MinIO with provider1 credentials
3. **Tests S3 Provider 2** - MinIO with provider2 credentials
4. **Tests S3 Provider 3** - MinIO with provider3 credentials
5. **Concurrent Operations** - Simultaneous uploads across all providers
6. **Pool Statistics** - Shows client pooling metrics

## Architecture

```
┌─────────────────┐
│  Application    │
└────────┬────────┘
         │
    ┌────▼────┐
    │   CP    │  Client Pool & Resolution
    │ Service │
    └────┬────┘
         │
    ┌────▼──────────────────────┐
    │  Provider Resolution       │
    │  (Context-based routing)   │
    └────┬──────────────────────┘
         │
    ┌────▼────┬────────┬────────┐
    │  Disk   │  S3-1  │  S3-2  │ S3-3
    │Provider │Provider│Provider│Provider
    └─────────┴────────┴────────┴────────┘
```

## Key Concepts

### Client Pooling
Clients are cached based on:
- Tenant ID
- Integration Type (provider)
- Optional: Hush ID (secrets management)

### Provider Resolution
Uses context values to determine the correct provider:
- `provider-type`: Which storage backend to use
- `tenant-id`: Multi-tenant isolation

### Credential Isolation
Each S3 provider has:
- Unique access credentials
- Dedicated bucket
- Isolated IAM policies

## Services

- **MinIO Console**: http://localhost:9001
  - Username: admin
  - Password: adminpassword

- **MinIO API**: http://localhost:9000
- **GCS API**: http://localhost:4443

## Created Resources

### MinIO Users
- `provider1` / `provider1secret` → bucket: `provider1-bucket`
- `provider2` / `provider2secret` → bucket: `provider2-bucket`
- `provider3` / `provider3secret` → bucket: `provider3-bucket`

### GCS Buckets
- `gcs-provider1-bucket`
- `gcs-provider2-bucket`
- `gcs-provider3-bucket`

## Cleanup

```bash
task teardown
task clean
```

## Notes

- Each provider operates independently
- Credentials are never shared between providers
- Client pooling reduces connection overhead
- Resolution rules can be extended for complex routing logic
