# Object Storage Examples

All examples are now exposed through a single CLI (`go run -tags examples ./pkg/objects/examples`).
Use the provided `Taskfile` in this directory for shortcuts, or invoke the CLI directly with the
subcommands listed below.

```bash
# List subcommands and global flags
go run -tags examples ./pkg/objects/examples --help
```

## Quick Start

```bash
# Simple disk-backed example
go run -tags examples ./pkg/objects/examples -- simple

# S3 / MinIO example (override defaults with flags or env vars)
go run -tags examples ./pkg/objects/examples -- simple-s3 --endpoint http://127.0.0.1:9000

# Multi-provider example (Docker + MinIO + fake GCS)
go run -tags examples ./pkg/objects/examples -- multi-provider

# Multi-tenant load scenario (requires tenants.json from setup command)
go run -tags examples ./pkg/objects/examples -- multi-tenant --ops=250 --concurrent=25

# Openlane integration (requires running Openlane API + PAT token)
go run -tags examples ./pkg/objects/examples -- openlane --token "$OPENLANE_AUTH_TOKEN"
```

If you prefer `task`, the bundled `Taskfile.yaml` exposes thin wrappers:

```bash
task -d pkg/objects/examples simple
# or
cd pkg/objects/examples
 task simple-s3 -- CLI_ARGS="--endpoint http://127.0.0.1:9000"
```

## Docker Infrastructure

A single `docker-compose.yml` (MinIO + fake-gcs-server) lives in `pkg/objects/examples`. The
`multi-provider` and `multi-tenant` commands start and stop services automatically, or you can use
`task multi-provider:setup` / `multi-tenant:setup` and their corresponding `:teardown` tasks.

## Scenario Overview

| Command | Highlights | Requirements |
|---------|------------|--------------|
| `simple` | Local disk provider, upload/download/delete, presigned URLs | none |
| `simple-s3` | S3-compatible provider (MinIO), local download verification | MinIO or AWS-compatible endpoint |
| `multi-provider` | Resolver + client pool across disk/S3/fake GCS, concurrent ops | Docker, docker-compose |
| `multi-tenant` | 1000+ tenant provisioning, throughput benchmarks, cache stats | Docker, docker-compose |
| `openlane` | Evidence creation against Openlane GraphQL API with file upload | Running Openlane instance, PAT token |

Assets used by the examples live under `pkg/objects/examples/assets` and `pkg/objects/examples/testdata`.
Feel free to replace them with your own files when testing workflows.

## Environment Overrides

Most flags accept environment variable defaults. A few commonly used ones:

- `OBJECTS_EXAMPLE_S3_ENDPOINT`, `OBJECTS_EXAMPLE_S3_ACCESS_KEY`, `OBJECTS_EXAMPLE_S3_SECRET_KEY`
- `OBJECTS_EXAMPLE_TENANT_OPS`, `OBJECTS_EXAMPLE_TENANT_CONCURRENCY`, `OBJECTS_EXAMPLE_TENANT_FILE`
- `OBJECTS_EXAMPLE_OPENLANE_API`, `OBJECTS_EXAMPLE_OPENLANE_TOKEN`

## Clean Up

Use the `:teardown` tasks or the CLI subcommands to stop docker services and remove temporary files:

```bash
# Multi-provider cleanup
go run -tags examples ./pkg/objects/examples -- multi-provider teardown

# Multi-tenant cleanup
go run -tags examples ./pkg/objects/examples -- multi-tenant teardown
```

You can also call `docker-compose -f pkg/objects/examples/docker-compose.yml down --remove-orphans`
if you have been experimenting manually.
