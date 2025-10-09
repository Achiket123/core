# pkg/cp Extraction Recommendations

## Executive Summary

Moving `pkg/cp` into its own repository as a standalone package is viable, but several API improvements and refactorings would significantly reduce glue code requirements for users.

**Current Glue Code**: 548 lines in `internal/objects/resolver`
**With Improvements**: ~160 lines (71% reduction)

## Critical Issues to Address

### 🔴 Issue 1: Confusing Naming

**Problem**: `Resolution.Credentials` is semantically backwards - it's actually the OUTPUT you want to produce, not credentials you're providing.

```go
// Current - confusing
type Resolution[Creds, Conf] struct {
    ClientType  ProviderType
    Credentials Creds  // ← Actually the OUTPUT
    Config      Conf   // ← Actually the CONFIG
}

// Example usage that reads backwards
resolution := resolver.Resolve(ctx, request, config)
myOutput := resolution.Credentials  // ← "Credentials" doesn't match intent
```

**Recommendation**: Rename to semantic clarity

```go
// Proposed - clear intent
type Result[Output, Config] struct {
    ProviderType ProviderType
    Output       Output  // Clear: this is what you get
    Config       Config  // Clear: configuration used
}

// Usage reads naturally
result := resolver.Resolve(ctx, request, config)
myOutput := result.Output  // ← Clear what this is
```

**Impact**: No size change, massive clarity improvement

### 🔴 Issue 2: Verbose Builder Interface

**Problem**: 4-method interface with mandatory boilerplate

```go
// Current: verbose 4-method interface
type ClientBuilder[T, Creds, Conf] interface {
    WithCredentials(Creds) ClientBuilder[T, Creds, Conf]
    WithConfig(Conf) ClientBuilder[T, Creds, Conf]
    Build(context.Context) (T, error)
    ClientType() ProviderType
}

// Implementation requires 4 methods + state management
type S3Builder struct {
    credentials S3Credentials
    config      S3Config
}

func (b *S3Builder) WithCredentials(creds S3Credentials) ClientBuilder[*s3.Client, S3Credentials, S3Config] {
    b.credentials = creds
    return b
}

func (b *S3Builder) WithConfig(cfg S3Config) ClientBuilder[*s3.Client, S3Credentials, S3Config] {
    b.config = cfg
    return b
}

func (b *S3Builder) Build(ctx context.Context) (*s3.Client, error) {
    // Build logic
}

func (b *S3Builder) ClientType() ProviderType {
    return S3Provider
}
```

**Recommendation**: Simplify to 2-method interface

```go
// Proposed: minimal 2-method interface
type Builder[T, Output, Config] interface {
    Build(ctx context.Context, output Output, config Config) (T, error)
    ProviderType() ProviderType
}

// Implementation is just Build + ProviderType
type S3Builder struct{}

func (b *S3Builder) Build(ctx context.Context, output S3Credentials, config S3Config) (*s3.Client, error) {
    // Build logic directly with output and config
    return s3.NewClient(ctx, output, config)
}

func (b *S3Builder) ProviderType() ProviderType {
    return S3Provider
}
```

**Impact**: -15 lines per builder, removes state management boilerplate

### 🔴 Issue 3: Domain-Specific CacheKey

**Problem**: `ClientCacheKey` contains Openlane-specific field

```go
type ClientCacheKey struct {
    TenantID        string
    IntegrationType string
    HushID          string  // ← Openlane-specific "hush" concept
    IntegrationID   string
}
```

**Recommendation**: Make generic or provide default

**Option A**: Fully generic key

```go
type CacheKey interface {
    comparable
    String() string
}

// Users provide their own key type
type MyAppCacheKey struct {
    TenantID   string
    ProviderID string
    CustomField string
}
```

**Option B**: Provide sensible default with extension

```go
type DefaultCacheKey struct {
    TenantID   string
    ProviderID string
}

// Users can extend with their own fields
type ExtendedCacheKey struct {
    DefaultCacheKey
    HushID string  // Openlane-specific
}
```

**Impact**: +20 lines, removes domain coupling

### 🟡 Issue 4: No Helpers for Common Patterns

**Problem**: Every user must reinvent common patterns

```go
// Current: Users manually create rules for every provider
func configureProviderRules(resolver *cp.Resolver[...]) {
    // Rule 1: Match S3 provider
    resolver.AddRule(
        cp.NewRule[...]().
            WhenFunc(func(ctx context.Context, req Request, _ Config) bool {
                hint, ok := cp.GetHint[string](ctx, ProviderHintKey)
                return ok && hint == "s3"
            }).
            Resolve(func(ctx context.Context, req Request, _ Config) (cp.Resolution[...], error) {
                // Fetch credentials from DB
                creds, err := queryCredentials(ctx, req.TenantID, "s3")
                if err != nil {
                    return cp.Resolution[...]{}, err
                }

                // Build config
                config := buildS3Config(creds)

                return cp.Resolution[...]{
                    ClientType:  cp.S3Provider,
                    Credentials: creds,
                    Config:      config,
                }, nil
            }),
    )

    // Rule 2: Match R2 provider... (repeat pattern)
    // Rule 3: Match database provider... (repeat pattern)
    // ... (139 lines of this)
}
```

**Recommendation**: Provide rule helpers

```go
// Proposed: Helper for hint-based matching
func MatchHint[Req, Res, Cfg any](
    hintKey HintKey[string],
    expectedValue string,
    resolveFunc func(context.Context, Req, Cfg) (Res, error),
) *Rule[Req, Res, Cfg] {
    return NewRule[Req, Res, Cfg]().
        WhenFunc(func(ctx context.Context, _ Req, _ Cfg) bool {
            hint, ok := GetHint[string](ctx, hintKey)
            return ok && hint == expectedValue
        }).
        Resolve(resolveFunc)
}

// Usage: Much cleaner
resolver.AddRule(
    cp.MatchHint(ProviderHintKey, "s3", resolveS3Provider),
)
resolver.AddRule(
    cp.MatchHint(ProviderHintKey, "r2", resolveR2Provider),
)
```

**Additional Helpers**:

```go
// Fallback chain helper
func FallbackChain[Req, Res, Cfg any](
    resolvers ...func(context.Context, Req, Cfg) (Res, error),
) *Rule[Req, Res, Cfg] {
    return NewRule[Req, Res, Cfg]().
        WhenFunc(func(ctx context.Context, _ Req, _ Cfg) bool {
            return true  // Always matches
        }).
        Resolve(func(ctx context.Context, req Req, cfg Cfg) (Res, error) {
            var lastErr error
            for _, resolver := range resolvers {
                res, err := resolver(ctx, req, cfg)
                if err == nil {
                    return res, nil
                }
                lastErr = err
            }
            return *new(Res), fmt.Errorf("all resolvers failed: %w", lastErr)
        })
}

// Conditional helper
func Conditional[Req, Res, Cfg any](
    predicate func(context.Context, Req, Cfg) bool,
    resolveFunc func(context.Context, Req, Cfg) (Res, error),
) *Rule[Req, Res, Cfg] {
    return NewRule[Req, Res, Cfg]().
        WhenFunc(predicate).
        Resolve(resolveFunc)
}
```

**Impact**: +120 lines in library, saves ~100 lines per user

### 🟡 Issue 5: No Credential Provider Abstraction

**Problem**: Users must manually query databases, parse metadata, convert types

```go
// Current: Manual credential fetching (223 lines in providers.go)
func querySystemProvider(ctx context.Context, client *ent.Client, tenantID string, integrationType models.IntegrationType) (*ent.Integration, error) {
    systemIntegration, err := client.Integration.Query().
        Where(
            integration.HasOwnerWith(organization.ID(tenantID)),
            integration.IntegrationTypeEQ(integrationType),
        ).
        Only(ctx)
    if err != nil {
        if ent.IsNotFound(err) {
            return nil, ErrNoProviderFound
        }
        return nil, err
    }

    return systemIntegration, nil
}

func resolveProviderFromConfig(ctx context.Context, integration *ent.Integration, providerType models.IntegrationType) (cp.Resolution[StorageCredentials, StorageConfig], error) {
    // Parse metadata
    metadata := integration.Metadata

    // Extract credentials based on provider type
    var creds StorageCredentials
    switch providerType {
    case models.IntegrationTypeS3:
        creds = StorageCredentials{
            AccessKeyID:     metadata["access_key_id"],
            SecretAccessKey: metadata["secret_access_key"],
        }
    case models.IntegrationTypeR2:
        // ... repeat pattern
    // ... 150+ more lines
    }
}
```

**Recommendation**: Provide credential provider interface

```go
// Proposed: Credential provider abstraction
type CredentialProvider[Output any] interface {
    FetchCredentials(ctx context.Context, key CacheKey) (Output, error)
}

// Database-backed provider implementation
type DatabaseCredentialProvider[Output any] struct {
    client       *ent.Client
    extractFunc  func(*ent.Integration) (Output, error)
}

func (p *DatabaseCredentialProvider[Output]) FetchCredentials(ctx context.Context, key CacheKey) (Output, error) {
    integration, err := p.client.Integration.Query().
        Where(
            integration.HasOwnerWith(organization.ID(key.TenantID)),
            integration.IntegrationTypeEQ(key.IntegrationType),
        ).
        Only(ctx)
    if err != nil {
        return *new(Output), err
    }

    return p.extractFunc(integration)
}

// Chained provider for fallbacks
type ChainedCredentialProvider[Output any] struct {
    providers []CredentialProvider[Output]
}

func (p *ChainedCredentialProvider[Output]) FetchCredentials(ctx context.Context, key CacheKey) (Output, error) {
    var lastErr error
    for _, provider := range p.providers {
        output, err := provider.FetchCredentials(ctx, key)
        if err == nil {
            return output, nil
        }
        lastErr = err
    }
    return *new(Output), fmt.Errorf("all providers failed: %w", lastErr)
}

// Usage: Clean and composable
dbProvider := &DatabaseCredentialProvider[StorageCredentials]{
    client: entClient,
    extractFunc: extractS3Credentials,
}

envProvider := &EnvironmentCredentialProvider[StorageCredentials]{
    envVars: []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"},
}

credProvider := &ChainedCredentialProvider[StorageCredentials]{
    providers: []CredentialProvider[StorageCredentials]{
        dbProvider,
        envProvider,
    },
}

// Integrate with resolver
resolver.AddRule(
    cp.MatchHint(ProviderHintKey, "s3", func(ctx context.Context, req Request, cfg Config) (Result, error) {
        creds, err := credProvider.FetchCredentials(ctx, buildCacheKey(req))
        if err != nil {
            return Result{}, err
        }

        return Result{
            ProviderType: S3Provider,
            Output:       creds,
            Config:       buildS3Config(creds),
        }, nil
    }),
)
```

**Impact**: +85 lines in library, saves ~180 lines per user

### 🟢 Issue 6: Pool-Specific Types Mixed with Core Resolution

**Problem**: Pool and resolution concerns entangled

**Recommendation**: Separate packages

```go
// Proposed structure
pkg/resolver/           # Core resolution (no pooling)
├── resolver.go         # Resolver, Rule
├── rules.go            # RuleBuilder
├── context.go          # Context helpers
└── hints.go            # HintKey, HintSet

pkg/pool/               # Client pooling (uses resolver)
├── pool.go             # ClientPool
├── service.go          # ClientService
└── builder.go          # ClientBuilder
```

**Impact**: ±0 lines, better separation of concerns

## Glue Code Analysis

### Current Implementation: 548 lines

**providers.go** (223 lines):
- `querySystemProvider`: Database queries (45 lines)
- `resolveProviderFromConfig`: Credential parsing (80 lines)
- `providerOptionsFromConfig`: Config building (98 lines)

**rules.go** (139 lines):
- `configureProviderRules`: Rule setup coordinator (30 lines)
- `ruleCoordinator.addKnownProviderRule`: Hint matching (35 lines)
- `ruleCoordinator.addModuleRule`: Module-based resolution (40 lines)
- `ruleCoordinator.addPreferredProviderRule`: User preference (34 lines)

**factory.go** (135 lines):
- `NewServiceFromConfig`: Wiring resolver + service (50 lines)
- `Build`: Integration resolution (45 lines)
- `buildWithRuntime`: Client building (40 lines)

**errors.go** (8 lines):
- Error definitions

### With Proposed Improvements: ~160 lines

**providers.go** (~80 lines, 65% reduction):
```go
// Use DatabaseCredentialProvider
s3Provider := &DatabaseCredentialProvider[StorageCredentials]{
    client: entClient,
    extractFunc: extractS3Credentials,
}

r2Provider := &DatabaseCredentialProvider[StorageCredentials]{
    client: entClient,
    extractFunc: extractR2Credentials,
}

// Chain with environment fallback
credProvider := &ChainedCredentialProvider[StorageCredentials]{
    providers: []CredentialProvider[StorageCredentials]{
        s3Provider,
        r2Provider,
        envProvider,
    },
}
```

**rules.go** (~40 lines, 71% reduction):
```go
// Use MatchHint helper
resolver.AddRule(cp.MatchHint(ProviderHintKey, "s3", resolveS3))
resolver.AddRule(cp.MatchHint(ProviderHintKey, "r2", resolveR2))
resolver.AddRule(cp.MatchHint(ProviderHintKey, "database", resolveDatabase))

// Use FallbackChain helper
resolver.AddRule(cp.FallbackChain(
    resolvePreferredProvider,
    resolveModuleProvider,
    resolveDefaultProvider,
))
```

**factory.go** (~40 lines, 70% reduction):
```go
// Simplified wiring
func NewServiceFromConfig(cfg Config) (*objects.Service, error) {
    credProvider := buildCredentialProvider(cfg)
    resolver := buildResolver(credProvider)
    clientService := cp.NewClientService[StorageClient, StorageCredentials, StorageConfig](resolver)

    return &objects.Service{
        clientService: clientService,
    }, nil
}
```

## Recommended API Changes

### Summary of Changes

| Change | Size Impact | Glue Code Reduction | Priority |
|--------|-------------|---------------------|----------|
| Rename Resolution → Result | ±0 | Clarity only | 🔴 Critical |
| Simplify Builder interface | -15 lines | -60 lines | 🔴 Critical |
| Generic CacheKey | +20 lines | Removes coupling | 🔴 Critical |
| Rule helpers | +120 lines | -100 lines | 🟡 High |
| Credential providers | +85 lines | -180 lines | 🟡 High |
| Separate pool/resolver | ±0 lines | Better organization | 🟢 Medium |

### Total Impact

**Library Size**:
- Current: 445 lines
- With changes: 445 - 15 + 20 + 120 + 85 = 655 lines (+47%)

**User Glue Code**:
- Current: 548 lines
- With changes: 548 - 60 - 100 - 180 = 208 lines (-62%)
- With optimized structure: ~160 lines (-71%)

**Net Benefit**: +210 lines in library, -388 lines for each user

## Migration Path

### Phase 1: Non-Breaking Improvements
1. Add new `Result` type (keep `Resolution` as alias)
2. Add helper functions (backward compatible)
3. Add `CredentialProvider` interface (optional)
4. Add documentation and examples

### Phase 2: Breaking Changes (v2.0)
1. Remove `Resolution` alias
2. Simplify `Builder` interface
3. Make `CacheKey` generic
4. Separate `pkg/resolver` and `pkg/pool`

### Phase 3: Advanced Features
1. Middleware support for rules
2. Rule composition utilities
3. Observability hooks (metrics, tracing)
4. Testing utilities

## Repository Structure

### Recommended Layout

```
github.com/theopenlane/resolver/
├── README.md
├── LICENSE
├── go.mod
├── resolver/
│   ├── resolver.go      # Core Resolver type
│   ├── rules.go         # Rule, RuleBuilder
│   ├── context.go       # Context helpers
│   ├── hints.go         # HintKey, HintSet
│   ├── helpers.go       # MatchHint, FallbackChain, etc.
│   ├── providers.go     # CredentialProvider interface
│   └── doc.go
├── pool/
│   ├── pool.go          # ClientPool
│   ├── service.go       # ClientService
│   ├── builder.go       # Builder interface
│   └── doc.go
├── examples/
│   ├── basic/
│   ├── authorization/
│   ├── storage/
│   └── workflow/
└── testing/
    └── testing.go       # Test utilities
```

### Documentation Focus

1. **Quick Start**: 5-minute example showing value
2. **Common Patterns**: Authorization, routing, validation
3. **Migration Guide**: From if/else chains to resolver
4. **API Reference**: Complete godoc
5. **Performance**: Benchmarks vs alternatives
6. **Best Practices**: When to use, when not to use

## Conclusion

Extracting `pkg/cp` as a standalone package is viable with these improvements:

**Benefits**:
- 🎯 71% glue code reduction for users
- 🧩 Generic and reusable across domains
- 📚 Clear, well-documented API
- 🔒 Type-safe with Go generics
- 🧪 Highly testable

**Costs**:
- 📈 47% size increase in library itself
- 🔄 Breaking changes for existing users
- 📝 Documentation effort

**Recommendation**: Proceed with extraction, implement all critical improvements before v1.0 release. The 47% library size increase is well worth the 71% glue code reduction for every user.
