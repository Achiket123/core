# pkg/cp: Existing vs Proposed API Comparison

This document provides side-by-side comparisons of the existing `pkg/cp` API and the proposed improvements in `pkg/cp/proposed/`.

## Quick Reference

| Aspect | Existing | Proposed | Files to Compare |
|--------|----------|----------|------------------|
| Context values | Custom helpers | contextx + HintKey | `context.go` vs `proposed/context.go` |
| Type naming | Resolution, Credentials | Result, Output | `resolution.go` vs `proposed/result.go` |
| Builder interface | 4 methods + state | 2 methods, no state | `builder.go` vs `proposed/builder.go` |
| Cache key | Domain-specific struct | Generic interface | `builder.go` vs `proposed/cachekey.go` |
| Rule helpers | Manual creation | Helper functions | _(none)_ vs `proposed/helpers.go` |

## 1. Context Values: Custom Helpers → contextx + HintKey

### Existing (`pkg/cp/context.go`)

```go
// Custom generic wrappers around context.WithValue
func WithValue[T any](ctx context.Context, key interface{}, value T) context.Context {
    return context.WithValue(ctx, key, value)
}

func GetValue[T any](ctx context.Context, key interface{}) (T, bool) {
    val := ctx.Value(key)
    if val == nil {
        return *new(T), false
    }
    typedVal, ok := val.(T)
    return typedVal, ok
}

// HintKey for named hints
type HintKey[T any] struct {
    name string
}

func WithHint[T any](ctx context.Context, key HintKey[T], value T) context.Context {
    // Custom storage implementation
}
```

**Problems:**
- `WithValue/GetValue` reinvents what contextx already provides
- Maintenance burden for functionality that exists elsewhere
- Users need to learn pkg/cp-specific context helpers

### Proposed (`pkg/cp/proposed/context.go`)

```go
import "github.com/theopenlane/utils/contextx"

// For singleton request-scoped data, use contextx directly:
//   ctx = contextx.With(ctx, dbClient)
//   client, ok := contextx.From[*ent.Client](ctx)

// HintKey remains for named metadata (multiple values of same type)
type HintKey[T any] struct {
    name string
}

func WithHint[T any](ctx context.Context, key HintKey[T], value T) context.Context {
    // Same implementation
}
```

**Why both?**

**contextx** (type-based keys):
- ONE value per type in the context
- Perfect for singletons: logger, DB client, user info
- Example: `contextx.With(ctx, dbClient)` - only ONE *ent.Client

**HintKey** (named keys):
- MULTIPLE values of the same type
- Necessary for resolver hints: provider, tenant, region all strings
- Example:
  ```go
  ctx = WithHint(ctx, ProviderHintKey, "s3")      // string
  ctx = WithHint(ctx, TenantHintKey, "tenant1")   // another string
  ctx = WithHint(ctx, RegionHintKey, "us-east-1") // yet another string
  ```

**Improvements:**
- Remove redundant `WithValue/GetValue` wrappers
- Leverage maintained contextx package
- Keep HintKey for resolver-specific named metadata
- Clear guidance on when to use each

**Impact:** -20 lines, better code reuse

---

## 2. Type Naming: Resolution → Result

### Existing (`pkg/cp/resolution.go`)

```go
// Resolution is a struct that represents the result of rule evaluation
type Resolution[Creds any, Conf any] struct {
    ClientType  ProviderType
    Credentials Creds  // ← Semantically backwards: actually the OUTPUT
    Config      Conf
    CacheKey    ClientCacheKey
}

// Usage reads awkwardly
resolution := resolver.Resolve(ctx)
myOutput := resolution.MustGet().Credentials  // ← "Credentials" confusing
```

**Problems:**
- "Resolution" suggests a process, not a result
- "Credentials" field is actually the output you want to produce
- Generic parameter "Creds" doesn't convey it's the desired output

### Proposed (`pkg/cp/proposed/result.go`)

```go
// Result represents the output of rule evaluation
type Result[Output any, Config any] struct {
    ProviderType ProviderType
    Output       Output  // ← Clear: this is what you get
    Config       Config
    CacheKey     CacheKey
}

// Usage reads naturally
result := resolver.Resolve(ctx)
myOutput := result.MustGet().Output  // ← Clear intent
```

**Improvements:**
- "Result" clearly indicates outcome
- "Output" clearly indicates what you're producing
- Generic parameter "Output" conveys purpose

**Impact:** ±0 lines, massive clarity improvement

---

## 3. Builder Interface: Simplified from 4 to 2 Methods

### Existing (`pkg/cp/builder.go`)

```go
// ClientBuilder builds client instances with credentials and configuration
type ClientBuilder[T any, Creds any, Conf any] interface {
    WithCredentials(credentials Creds) ClientBuilder[T, Creds, Conf]
    WithConfig(config Conf) ClientBuilder[T, Creds, Conf]
    Build(ctx context.Context) (T, error)
    ClientType() ProviderType
}

// Example implementation requires state management
type S3Builder struct {
    credentials S3Credentials  // State field
    config      S3Config       // State field
}

func (b *S3Builder) WithCredentials(creds S3Credentials) ClientBuilder[*s3.Client, S3Credentials, S3Config] {
    b.credentials = creds  // Store state
    return b
}

func (b *S3Builder) WithConfig(cfg S3Config) ClientBuilder[*s3.Client, S3Credentials, S3Config] {
    b.config = cfg  // Store state
    return b
}

func (b *S3Builder) Build(ctx context.Context) (*s3.Client, error) {
    // Use stored state
    return buildS3Client(ctx, b.credentials, b.config)
}

func (b *S3Builder) ClientType() ProviderType {
    return S3Provider
}

// Total: ~30 lines per builder
```

### Proposed (`pkg/cp/proposed/builder.go`)

```go
// Builder builds client instances with output and configuration
type Builder[T any, Output any, Config any] interface {
    Build(ctx context.Context, output Output, config Config) (T, error)
    ProviderType() ProviderType
}

// Example implementation - no state needed
type S3Builder struct{}

func (b *S3Builder) Build(ctx context.Context, output S3Credentials, config S3Config) (*s3.Client, error) {
    // Direct build with parameters
    return buildS3Client(ctx, output, config)
}

func (b *S3Builder) ProviderType() ProviderType {
    return S3Provider
}

// Total: ~15 lines per builder
```

**Improvements:**
- 4 methods → 2 methods (50% reduction)
- No state management needed
- ~30 lines → ~15 lines per implementation (50% reduction)
- Direct parameter passing vs intermediate state

**Impact:** -15 lines per builder

---

## 4. Cache Key: Domain-Specific → Generic

### Existing (`pkg/cp/builder.go`)

```go
// ClientCacheKey uniquely identifies a cached client
type ClientCacheKey struct {
    TenantID        string
    IntegrationType string
    HushID          string  // ← Openlane-specific concept
    IntegrationID   string
}

// Fixed structure, can't extend
```

**Problems:**
- Contains domain-specific "HushID" field
- Not extensible for other use cases
- Forces all users to include unused fields

### Proposed (`pkg/cp/proposed/cachekey.go`)

```go
// CacheKey is the interface that cache keys must implement
type CacheKey interface {
    String() string
}

// Users define whatever structure they need
type MyCacheKey struct {
    TenantID   string
    ProviderID string
    HushID     string  // Add whatever fields you need
}

func (k MyCacheKey) String() string {
    return fmt.Sprintf("%s:%s:%s", k.TenantID, k.ProviderID, k.HushID)
}

// Usage:
key := MyCacheKey{
    TenantID:   "tenant1",
    ProviderID: "s3",
    HushID:     "hush123",
}
```

**Improvements:**
- Clean interface, no opinions
- Define whatever structure you need
- No library-provided "defaults" or "extendables"
- Forces nothing on users

**Impact:** -30 lines, removes domain coupling

---

## 5. Rule Helpers: Manual → Helper Functions

### Existing (`pkg/cp/rules.go` + user code)

Every user must manually write:

```go
// In user code: 15-20 lines PER RULE
resolver.AddRule(
    cp.NewRule[StorageClient, StorageCredentials, StorageConfig]().
        WhenFunc(func(ctx context.Context) bool {
            hint, ok := cp.GetHint[string](ctx, ProviderHintKey)
            return ok && hint == "s3"
        }).
        Resolve(func(ctx context.Context) (*cp.ResolvedProvider[StorageCredentials, StorageConfig], error) {
            // Query database
            integration, err := client.Integration.Query().
                Where(integration.IntegrationTypeEQ("s3")).
                Only(ctx)
            if err != nil {
                return nil, err
            }

            // Parse metadata
            metadata := integration.Metadata

            // Extract credentials
            creds := StorageCredentials{
                AccessKeyID:     metadata["access_key_id"],
                SecretAccessKey: metadata["secret_access_key"],
            }

            // Build config
            config := buildS3Config(creds)

            return &cp.ResolvedProvider[StorageCredentials, StorageConfig]{
                Type:        S3Provider,
                Credentials: creds,
                Config:      config,
            }, nil
        }),
)

// Repeat this pattern for EVERY provider type
// S3: 20 lines
// R2: 20 lines
// Database: 20 lines
// etc.
```

### Proposed (`pkg/cp/proposed/helpers.go`)

```go
// MatchHint helper eliminates boilerplate
resolver.AddRule(
    proposed.MatchHint(ProviderHintKey, "s3", resolveS3Provider),
)

// Additional helpers for common patterns:

// Match any of multiple values
resolver.AddRule(
    proposed.MatchHintAny(
        ProviderHintKey,
        []string{"s3", "r2"},
        resolveObjectStorageProvider,
    ),
)

// Try multiple resolvers in order (fallback chain)
resolver.AddRule(
    proposed.FallbackChain(
        resolveDatabaseCredentials,  // Try database first
        resolveEnvCredentials,       // Fall back to environment
        resolveDefaultCredentials,   // Finally use defaults
    ),
)

// Conditional with custom predicate
resolver.AddRule(
    proposed.Conditional(
        func(ctx context.Context) bool {
            return auth.IsSystemAdmin(ctx)
        },
        resolveAdminProvider,
    ),
)

// Cache rule results for performance
resolver.AddRule(
    proposed.CachedRule(expensiveRule, 5*time.Minute),
)
```

**Available Helpers:**
- `MatchHint` - match single hint value
- `MatchHintAny` - match any of multiple hint values
- `FallbackChain` - try resolvers in order
- `Conditional` - custom predicate
- `CachedRule` - cache rule results
- `FirstMatch` - evaluate multiple rules

**Improvements:**
- 15-20 lines → 1-2 lines per rule
- Eliminates repetitive boilerplate
- Common patterns become one-liners
- Still allows custom rules when needed

**Impact:** +120 lines in library, saves ~100 lines per user

---

## Side-by-Side File Comparison

### Core Resolution Types

| Existing | Proposed | Key Changes |
|----------|----------|-------------|
| `resolution.go` (62 lines) | `result.go` (62 lines) | Renamed: Resolution→Result, Credentials→Output |
| `rules.go` (66 lines) | `rulebuilder.go` (66 lines) | Updated to use Result type |

### Builder and Caching

| Existing | Proposed | Key Changes |
|----------|----------|-------------|
| `builder.go` (51 lines) | `builder.go` (45 lines) | 4-method → 2-method interface |
| `builder.go` (struct in file) | `cachekey.go` (60 lines) | Domain-specific struct → generic interface |

### Helpers and Utilities

| Existing | Proposed | Key Changes |
|----------|----------|-------------|
| _(none)_ | `helpers.go` (120 lines) | NEW: Rule helper functions |

### Context and Hints (unchanged)

| Existing | Proposed | Key Changes |
|----------|----------|-------------|
| `context.go` (80 lines) | `context.go` (80 lines) | No changes needed |
| `hints.go` (38 lines) | `hints.go` (38 lines) | No changes needed |

---

## Size Impact Summary

### Library Size

| Component | Current | Proposed | Delta |
|-----------|---------|----------|-------|
| Core types | 445 lines | 400 lines | -45 (-10%) |
| Helpers | 0 lines | 120 lines | +120 (new) |
| **Total** | **445 lines** | **520 lines** | **+75 (+17%)** |

### User Glue Code (internal/objects/resolver)

| File | Current | With Proposed | Delta |
|------|---------|---------------|-------|
| rules.go | 139 lines | ~40 lines | -99 (-71%) |
| factory.go | 135 lines | ~60 lines | -75 (-56%) |
| **Total** | **274 lines** | **100 lines** | **-174 (-63%)** |

### Net Impact

**Per User:**
- Library cost: +75 lines (one-time)
- User savings: -174 lines (glue code for rules/factory)
- **Net savings: +99 lines**

**Break-even: 1 user**

With just ONE user, the glue code savings exceed the library additions.

---

## Migration Path

### Option 1: Big Bang (v2.0)

1. Replace existing API entirely
2. Update all internal usage
3. Provide migration guide
4. Release as v2.0.0

**Pros:** Clean break, no legacy burden
**Cons:** Breaking change for existing users

### Option 2: Gradual (v1.x → v2.0)

1. Add new types alongside existing (v1.9)
2. Mark old types as deprecated
3. Add helpers and providers (v1.10)
4. Migrate internal usage (v1.11)
5. Remove deprecated types (v2.0)

**Pros:** Smooth transition, backward compatible
**Cons:** Temporary code duplication

### Option 3: Modular (separate packages)

1. Keep existing `pkg/cp` as-is
2. Create `pkg/resolver` with new API
3. Create `pkg/resolver/helpers` for helpers
4. Create `pkg/resolver/providers` for providers
5. Users opt in by importing new package

**Pros:** No breaking changes, users choose when to migrate
**Cons:** Two packages to maintain

---

## Recommendations

1. **Review the proposed code** in `pkg/cp/proposed/`
2. **Run the examples**: `go test -v ./pkg/cp/proposed/...`
3. **Compare implementations** side-by-side
4. **Decide on migration strategy**:
   - Modular approach (recommended) - lowest risk
   - Gradual migration - balanced approach
   - Big bang v2.0 - cleanest API

## Try It Out

```bash
# View the proposed code
ls -la pkg/cp/proposed/

# See the examples
cat pkg/cp/proposed/examples_test.go

# Run the examples (after fixing imports)
go test -v ./pkg/cp/proposed/...

# Compare files side-by-side
diff -u pkg/cp/resolution.go pkg/cp/proposed/result.go
diff -u pkg/cp/builder.go pkg/cp/proposed/builder.go
```

---

## Questions to Consider

1. **contextx integration**: Does combining contextx (for singletons) + HintKey (for named hints) provide the right abstraction?

2. **Type naming**: Do "Result" and "Output" read more clearly than "Resolution" and "Credentials"?

3. **Builder interface**: Is 2-method interface simpler than 4-method? Is losing the fluent API (WithCredentials/WithConfig) acceptable?

4. **Cache key**: Does generic interface provide enough flexibility? Is default implementation sufficient for most cases?

5. **Helpers**: Are the provided helpers useful? Are there other common patterns that should have helpers?

6. **Migration**: Which migration path is most appropriate for your needs?

## Conclusion

The proposed improvements aim to:
- **Improve clarity** through better naming
- **Reduce boilerplate** with helper functions
- **Increase flexibility** with generic interfaces
- **Maintain simplicity** in the core API

Total impact: +75 lines in library, -174 lines per user, **net +99 lines saved** for first user.
