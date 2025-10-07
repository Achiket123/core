# cp – Context-Aware Client Pooling

`pkg/cp` is a lightweight, generic client-pooling and provider-resolution
library. It helps long-running services create, reuse, and expire expensive
clients (SDKs, database connections, storage clients, etc.) based on runtime
context. The package is completely self-contained and can be imported into
another module as-is.

---

## Why `cp`?

- **Context-aware resolution** – rules decide which provider/credentials to use
  for a request (per tenant, per feature, per region…).
- **Safe client reuse** – clients are cached with TTL-based expiry and can be
  reused across goroutines.
- **Generic API** – implemented with Go generics so you can pool any client
  type, credential struct, or configuration struct without adapters.
- **Extensible builders** – plug in new providers by registering a
  `ClientBuilder` that knows how to construct a concrete client.

The library powers the object storage service, but it is provider-agnostic and
usable for any type of resource (email APIs, payment SDKs, etc.).

---

## Core Building Blocks

```
pkg/cp
├── builder.go     # ClientBuilder interface, ClientPool, cache keys
├── service.go     # ClientService: builder registry + pool orchestration
├── resolution.go  # Resolver: ordered rule evaluation
├── rules.go       # Rule builder DSL and helper types
├── context.go     # Generic context helpers built on contextx
└── examples/      # Usage samples (standalone module)
```

### ClientPool

```go
pool := cp.NewClientPool[*StorageClient](10 * time.Minute)
```

- Stores clients keyed by `cp.ClientCacheKey` (tenant, integration type, hush ID,
  integration ID).
- Automatically expires entries after the configured TTL.
- Thread-safe get/set operations.

### ClientService

Wraps the pool and the registered builders:

```go
service := cp.NewClientService[
    *StorageClient,
    storage.ProviderCredentials,
    *storage.ProviderOptions,
](pool,
  cp.WithConfigClone[*StorageClient](cloneProviderOptions),
)

service.RegisterBuilder(cp.ProviderType("s3"), &s3Builder{})
service.RegisterBuilder(cp.ProviderType("disk"), &diskBuilder{})
```

`WithConfigClone` / `WithCredentialClone` let you supply defensive copy
functions so builders never mutate caller-owned state.

`GetClient` uses the cache first, otherwise invokes the registered builder,
stores the result, and returns it as an `mo.Option[T]`:

```go
client := service.GetClient(
    ctx,
    cp.ClientCacheKey{TenantID: "tenant-1", IntegrationType: "s3"},
    cp.ProviderType("s3"),
    creds,
    opts,
)
if client.IsPresent() {
    use(client.MustGet())
}
```

### Resolver & Rules

The resolver encapsulates the decision tree that picks credentials/config per
request. Several similarly named types collaborate to make that happen:

- **`Resolution[Creds, Conf]`** – the result a rule returns. It records the
  provider type plus the credential/config structs your builders expect
  (`pkg/cp/resolution.go:10`).
- **`ResolutionRule[T, Creds, Conf]`** – a function wrapper that inspects the
  context and either returns `Some(Resolution)` or `None` to skip the rule
  (`pkg/cp/resolution.go:18`).
- **`Resolver[T, Creds, Conf]`** – an ordered list of rules (with an optional
  default) evaluated until one produces a resolution (`pkg/cp/resolution.go:47`).
- **`RuleBuilder[T, Creds, Conf]`** – a fluent DSL for constructing a
  `ResolutionRule`. Chain `WhenFunc` predicates to guard execution, then call
  `Resolve` with the function that produces a provider
  (`pkg/cp/rules.go:14`).
- **`ResolvedProvider[Creds, Conf]`** – the value returned from that resolver
  function. The builder converts it into a `Resolution` for the outer
  machinery (`pkg/cp/rules.go:35`).

Putting those pieces together looks like this:

```go
resolver := cp.NewResolver[
    *StorageClient,
    storage.ProviderCredentials,
    *storage.ProviderOptions,
]()

resolver.AddRule(
    cp.NewRule[*StorageClient]().
        WhenFunc(func(ctx context.Context) bool {
            provider, _ := ctx.Value(providerKey).(string)
            return provider == "disk"
        }).
        Resolve(func(context.Context) (*cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions], error) {
            return &cp.ResolvedProvider[storage.ProviderCredentials, *storage.ProviderOptions]{
                Type:        cp.ProviderType("disk"),
                Credentials: storage.ProviderCredentials{},
                Config:      storage.NewProviderOptions(storage.WithBucket("./tmp")),
            }, nil
        }),
)
```

Resolvers evaluate rules in the order they were added. The first rule that
returns a value wins; you can also configure a fallback rule with
`SetDefaultRule`. Rules are ordinary functions, so they can perform database
lookups, call external services, or compose other resolvers.

### Context Helpers

`cp.WithValue` / `cp.GetValue` wrap `context.Context` using
`github.com/theopenlane/utils/contextx`, letting you stash arbitrary typed
values that your rules can inspect without bespoke keys. For richer metadata,
`cp.NewHintKey`, `cp.WithHint`, and `cp.GetHint` provide structured, typed hint
storage—handy for sharing things like "known provider", "preferred provider", or
"feature module" across multiple rules.

---

## Typical Workflow

1. **Create a pool** – choose an eviction TTL appropriate for your client type.
2. **Register builders** – implement `ClientBuilder` for each provider; builders
   usually read credentials/options produced by the resolver.
3. **Describe routing rules** – create a `Resolver` and add rules that inspect
   request context (tenant ID, provider hints, feature flags, …).
4. **Resolve + get client** – at request time, resolve a provider to obtain the
   provider type + credentials/config, then call `ClientService.GetClient`.
5. **Use the client** – hand the cached client to your business logic. Since
   `GetClient` returns an `mo.Option[T]`, always check `.IsPresent()`.

The [`pkg/objects/examples`](../objects/examples) directories demonstrate this
workflow in depth (`multi-provider` and `multi-tenant` use cp extensively).

---

## Extending & Testing

- Builders can cache internal state; they are recreated only when the resolver
  chooses their provider type.
- Implement `ClientBuilder.ClientType()` to keep `ClientService` honest – it can
  assert that the registered type matches what the builder reports.
- For unit tests, construct in-memory resolvers/pools with short TTLs and stub
  builders. Because the API is generic, you can reuse the same helpers for any
  client type (see `pkg/cp/examples`).
- Production code often wraps `ClientService` with a thin service (see
  [`internal/objects/service.go`](../../internal/objects/service.go)) that
  combines resolver + service + domain-specific helpers.

---

## Moving to Another Repository

`pkg/cp` has zero dependencies on the rest of this module outside of
`github.com/theopenlane/utils/contextx` (for type-safe context helpers) and
`github.com/samber/mo` (for the `Option` type). To reuse it elsewhere:

1. Copy the `pkg/cp` directory into your project.
2. Replace module paths in imports if needed.
3. Provide concrete builders/resolvers specific to your domain.

The package is intentionally small and stable so it can sit at the heart of
other services that require dynamic client reuse.

---

## Additional Resources

- `pkg/cp/examples` – a standalone module with runnable examples.
- `pkg/objects/examples/multi-tenant` – realistic cp usage with per-tenant S3
  credentials.
- `pkg/objects/examples/multi-provider` – cp orchestrating multiple providers in
  parallel.
