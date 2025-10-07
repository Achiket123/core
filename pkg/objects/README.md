# Object Storage Toolkit

`pkg/objects` provides a modular object storage toolkit built around three core
ideas:

1. **Unified file pipeline** – helpers for parsing uploads, validating files,
   and mapping them onto your domain models.
2. **Provider-agnostic storage service** – a thin runtime (`storage.ObjectService`)
   that wraps individual providers (disk, S3, R2, …) behind a consistent API.
3. **Dynamic provider resolution** – optional integration with
   [`pkg/cp`](../cp) so each request can resolve the correct provider/credentials
   at runtime (per tenant, per user, per feature, etc.).

The package is designed so it can be lifted into a standalone repository without
changes. All cross-package usage flows through public APIs that do not rely on
private project state.

---

## Package Layout

```
pkg/objects
├── adapter.go             # Mutation adapter helpers for ent hooks
├── upload.go              # Multipart / GraphQL upload parsing utilities
├── validators.go          # Reusable validation helpers and combinators
├── storage/
│   ├── service.go         # Provider-neutral object service (upload/download…)
│   ├── types.go           # Aliases & convenience wrappers around storagetypes
│   ├── utils.go           # MIME detection, document parsing, local helpers
│   ├── providers/
│   │   ├── disk/          # Local filesystem implementation
│   │   ├── s3/            # AWS S3 / MinIO provider
│   │   └── r2/            # Cloudflare R2 provider
│   └── types/             # Base interfaces and shared structs
├── examples/              # End-to-end runnable scenarios (Task-based)
└── mocks/                 # `mockery`-generated interfaces for testing
```

Key exported packages:

- `objects` – request/context helpers, validation, ent mutation adapters.
- `objects/storage` – high-level service plus convenience aliases.
- `objects/storage/types` – canonical provider interfaces and DTOs.
- `objects/storage/providers/*` – concrete provider implementations.

---

## High-Level Architecture

```
┌─────────────────────────────────────────────┐
│        Application / GraphQL / REST         │
└───────────────┬────────────────────────────┘
                │ Parse & validate uploads
┌───────────────▼────────────────────────────┐
│                pkg/objects                  │
│  • ParseFilesFromSource / ProcessFiles…     │
│  • ValidationFunc / NameGeneratorFunc       │
│  • Mutation adapters & context helpers      │
└───────────────┬────────────────────────────┘
                │ Resolved provider (optional)
┌───────────────▼────────────────────────────┐
│            storage.ObjectService            │
│      Upload / Download / Delete / URLs      │
└───────────────┬────────────────────────────┘
                │ storagetypes.Provider
┌───────────────▼────────────┬───────────────┐
│  disk.Provider   s3.Provider   r2.Provider │ … custom
└────────────────────────────┴───────────────┘
```

When combined with `pkg/cp`, provider resolution happens before the
`ObjectService` call. See [`internal/objects/service.go`](../../internal/objects/service.go)
for a full orchestration example that powers the production API.

---

## Working with Uploads

### Parse Incoming Files

`ParseFilesFromSource` normalises uploads regardless of the transport you are
using (GraphQL multipart requests, `*http.Request`, or pre-parsed
`map[string]any` payloads):

```go
files, err := objects.ParseFilesFromSource(r, "avatarFile", "attachment")
if err != nil {
    return err
}
for field, uploads := range files {
    for _, file := range uploads {
        // file is storage.File (ReadSeeker + metadata)
    }
}
```

Common tasks:

- `objects.ProcessFilesForMutation` – decorate parsed files with parent IDs in
  ent mutations.
- `storage.DetectContentType` – robust MIME detection for `io.ReadSeeker`s.
- `objects.NewUploadFile` – load local files (CLI, tests) into
  `storage.File` instances.

### Validation & Naming

`storage.ObjectService` lets you plug in custom validators or filename
strategies:

```go
svc := storage.NewObjectService().
    WithValidation(objects.ChainValidators(
        objects.MimeTypeValidator("image/png", "image/jpeg"),
        func(f storage.File) error {
            if f.Size > 4<<20 {
                return fmt.Errorf("max 4 MB")
            }
            return nil
        },
    ))
```

Validator helpers live in `objects/validators.go`. You can supply your own
`ValidationFunc`, `NameGeneratorFunc`, custom upload handlers, or error
formatters via the fluent `With…` helpers on `ObjectService`.

---

## Storage Service API

`storage.ObjectService` focuses on four core operations that work uniformly
across providers:

```go
file, err := svc.Upload(ctx, provider, content, &storage.UploadOptions{
    FileName:    "avatar.png",
    ContentType: "image/png",
    Bucket:      "avatars",
})

metadata, err := svc.Download(ctx, provider, fileRef, &storage.DownloadOptions{})
url, err := svc.GetPresignedURL(ctx, provider, fileRef, &storagetypes.PresignedURLOptions{Duration: 15 * time.Minute})
err = svc.Delete(ctx, provider, fileRef, &storagetypes.DeleteFileOptions{})
```

Returned models embed `storagetypes.FileMetadata`, so you retain object size,
content type, provider info, and generated keys.

### Provider Options

Providers consume `storage.ProviderOptions`, built with functional options such
as `storage.WithBucket`, `storage.WithCredentials`, `storage.WithRegion`,
`storage.WithEndpoint`, and `storage.WithExtra`. Options are immutable – call
`Clone()` when you need to duplicate a configuration.

---

## Provider Catalogue

| Provider | Package | Notes |
|----------|---------|-------|
| Disk | `storage/providers/disk` | Local filesystem with optional `LocalURL` for presigned links |
| S3 / MinIO | `storage/providers/s3` | Full AWS SDK v2 integration, multipart uploads, presigned URLs |
| Cloudflare R2 | `storage/providers/r2` | R2-specific credential and endpoint handling |

Implementing a new provider requires satisfying `storagetypes.Provider`. Start
from the `disk` provider for a concise example. Provider builders typically use
`storage.ProviderOptions` for runtime configuration and must report their
`ProviderType`.

---

## Resolving Providers Dynamically

The package stays agnostic of how you fetch credentials. The recommended
approach is to pair it with [`pkg/cp`](../cp):

1. Define a `cp.Resolver` that looks at context (tenant ID, feature flags, etc.)
   and returns `storage.ProviderCredentials` + `storage.ProviderOptions`.
2. Register builders with `cp.ClientService` for each supported provider type.
3. Use `internal/objects.Service` (or your own orchestrator) to connect the
   resolver + client pool with `storage.ObjectService`.

See `pkg/objects/examples/multi-tenant` and
`pkg/objects/examples/multi-provider` for full examples. Each directory ships
with a `Taskfile.yml`; run them via:

```bash
# from repository root
cd pkg/objects/examples
task multi-provider:run
```

---

## Utilities & Helpers

- `storage.ParseDocument` – JSON/YAML/plain-text parsing into native Go values.
- `storage.NewUploadFile` – turn a filesystem path into an uploadable file.
- `objects.FileContextKey` – store/retrieve uploads from request contexts.
- `objects.GenericMutationAdapter` – adapt generated ent mutations to the
  simplified `objects.Mutation` interface used by the helper functions.

Mocks for provider interfaces live in `pkg/objects/mocks`, generated via
`mockery` to make integration tests straightforward.

---

## Testing & Examples

- `pkg/objects/storage/service_test.go` covers the service API basics.
- Provider packages include focused unit tests exercising edge cases.
- Five runnable examples (`simple`, `simple-s3`, `multi-provider`,
  `multi-tenant`, `e2e-openlane`) demonstrate everything from a local disk flow
  to full GraphQL integration with Openlane. Each example has a detailed README
  and Taskfile to handle setup/build/run/cleanup.

---

## Migrating / Embedding in Other Projects

The package has no hard dependency on the rest of this repository beyond the
`pkg/cp` integration and optional ent helpers. To embed it elsewhere:

1. Copy `pkg/objects` (and optionally `pkg/cp` if you need runtime resolution).
2. Adjust import paths to your module.
3. Wire a resolver + client service (or pass providers directly) when calling
   `storage.ObjectService`.

Because the providers only interact through `storagetypes.Provider`, you can add
custom implementations without modifying any other package files.

---

## Further Reading

- [`internal/objects/service.go`](../../internal/objects/service.go) – example
  of orchestrating resolver + client pool + object service.
- [`pkg/objects/examples`](./examples) – curated tasks demonstrating common
  deployment models.
- [`pkg/cp`](../cp) – resolver and client pooling mechanics used throughout the
  system.
