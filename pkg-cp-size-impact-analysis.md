# pkg/cp Size Impact Analysis

## Executive Summary

Analyzing how much the proposed improvements would add to `pkg/cp` package size:

**Current Size**: 445 lines (excluding tests)
**With All Improvements**: 635 lines (core) to 735 lines (with testing utils)
**Size Increase**: +43% to +65%

**Trade-off**: +190 lines in library saves ~388 lines of glue code per user

## Current Package Size

### Breakdown by File

```
pkg/cp/
├── builder.go      51 lines   (ClientBuilder, CacheKey, Pool, Service interfaces)
├── context.go      80 lines   (WithValue, GetValue, HintKey, WithHint, GetHint)
├── hints.go        38 lines   (HintSet, AddHint, Apply)
├── pool.go         63 lines   (ClientPool implementation)
├── resolution.go   62 lines   (Resolution, ResolutionRule, Resolver)
├── rules.go        66 lines   (RuleBuilder, ResolvedProvider, DefaultRule)
├── service.go      90 lines   (ClientService with RegisterBuilder, GetClient)
├── doc.go           2 lines   (Package documentation)
└── errors.go        3 lines   (Error definitions)
───────────────────────────────
Total:             445 lines
```

### Test Coverage

```
pkg/cp/
├── context_test.go       ~120 lines
├── hints_test.go         ~80 lines
├── pool_test.go          ~150 lines
├── resolution_test.go    ~100 lines
├── rules_test.go         ~95 lines
├── service_test.go       ~180 lines
└── examples_test.go      ~200 lines
─────────────────────────────────
Total tests:             ~925 lines
```

**Total Current**: 445 + 925 = 1,370 lines

## Proposed Changes - Size Impact

### 1. Type Renames (±0 lines)

**Changes**:
- `Resolution` → `Result`
- `Credentials` → `Output`
- `ClientBuilder` → `Builder`

**Impact**: Name changes only, no size impact

```diff
- type Resolution[Creds, Conf] struct {
-     ClientType  ProviderType
-     Credentials Creds
-     Config      Conf
- }

+ type Result[Output, Config] struct {
+     ProviderType ProviderType
+     Output       Output
+     Config       Config
+ }
```

**Size**: ±0 lines (rename only)

### 2. Simplified Builder Interface (-15 lines)

**Current** (4 methods):
```go
type ClientBuilder[T, Creds, Conf] interface {
    WithCredentials(Creds) ClientBuilder[T, Creds, Conf]  // ~3 lines per impl
    WithConfig(Conf) ClientBuilder[T, Creds, Conf]        // ~3 lines per impl
    Build(context.Context) (T, error)                      // ~10 lines per impl
    ClientType() ProviderType                              // ~2 lines per impl
}

// Interface definition: ~10 lines
// Example implementation: ~30 lines
```

**Proposed** (2 methods):
```go
type Builder[T, Output, Config] interface {
    Build(ctx context.Context, output Output, config Config) (T, error)  // ~10 lines per impl
    ProviderType() ProviderType                                          // ~2 lines per impl
}

// Interface definition: ~5 lines
// Example implementation: ~15 lines
```

**Size**: -15 lines (interface + example)

### 3. Generic CacheKey (+20 lines)

**Current**:
```go
type ClientCacheKey struct {
    TenantID        string
    IntegrationType string
    HushID          string  // Domain-specific
    IntegrationID   string
}
// ~10 lines
```

**Proposed**:
```go
// CacheKey is the interface that cache keys must implement
type CacheKey interface {
    comparable
    String() string
}

// DefaultCacheKey provides a sensible default implementation
type DefaultCacheKey struct {
    TenantID   string
    ProviderID string
}

func (k DefaultCacheKey) String() string {
    return fmt.Sprintf("%s:%s", k.TenantID, k.ProviderID)
}

// ExtendableCacheKey allows embedding for custom fields
type ExtendableCacheKey[T any] struct {
    DefaultCacheKey
    Custom T
}

func (k ExtendableCacheKey[T]) String() string {
    return fmt.Sprintf("%s:%v", k.DefaultCacheKey.String(), k.Custom)
}

// ~30 lines total
```

**Size**: +20 lines

### 4. Rule Helpers (+120 lines)

**New File**: `helpers.go`

```go
// MatchHint creates a rule that matches a hint value
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
// ~15 lines

// MatchHintAny creates a rule that matches any of the provided hint values
func MatchHintAny[Req, Res, Cfg any](
    hintKey HintKey[string],
    expectedValues []string,
    resolveFunc func(context.Context, Req, Cfg) (Res, error),
) *Rule[Req, Res, Cfg] {
    return NewRule[Req, Res, Cfg]().
        WhenFunc(func(ctx context.Context, _ Req, _ Cfg) bool {
            hint, ok := GetHint[string](ctx, hintKey)
            if !ok {
                return false
            }
            for _, expected := range expectedValues {
                if hint == expected {
                    return true
                }
            }
            return false
        }).
        Resolve(resolveFunc)
}
// ~20 lines

// FallbackChain creates a rule that tries resolvers in order
func FallbackChain[Req, Res, Cfg any](
    resolvers ...func(context.Context, Req, Cfg) (Res, error),
) *Rule[Req, Res, Cfg] {
    return NewRule[Req, Res, Cfg]().
        WhenFunc(func(ctx context.Context, _ Req, _ Cfg) bool {
            return true
        }).
        Resolve(func(ctx context.Context, req Req, cfg Cfg) (Res, error) {
            var lastErr error
            for i, resolver := range resolvers {
                res, err := resolver(ctx, req, cfg)
                if err == nil {
                    return res, nil
                }
                lastErr = fmt.Errorf("resolver %d failed: %w", i, err)
            }
            return *new(Res), fmt.Errorf("all resolvers failed: %w", lastErr)
        })
}
// ~25 lines

// Conditional creates a rule with a custom predicate
func Conditional[Req, Res, Cfg any](
    predicate func(context.Context, Req, Cfg) bool,
    resolveFunc func(context.Context, Req, Cfg) (Res, error),
) *Rule[Req, Res, Cfg] {
    return NewRule[Req, Res, Cfg]().
        WhenFunc(predicate).
        Resolve(resolveFunc)
}
// ~10 lines

// MapRequest creates a rule that transforms the request
func MapRequest[Req1, Req2, Res, Cfg any](
    mapper func(Req1) Req2,
    rule *Rule[Req2, Res, Cfg],
) *Rule[Req1, Res, Cfg] {
    return NewRule[Req1, Res, Cfg]().
        WhenFunc(func(ctx context.Context, req Req1, cfg Cfg) bool {
            req2 := mapper(req)
            return rule.when(ctx, req2, cfg)
        }).
        Resolve(func(ctx context.Context, req Req1, cfg Cfg) (Res, error) {
            req2 := mapper(req)
            return rule.resolve(ctx, req2, cfg)
        })
}
// ~20 lines

// CacheRule wraps a rule with caching
func CacheRule[Req comparable, Res, Cfg any](
    rule *Rule[Req, Res, Cfg],
    ttl time.Duration,
) *Rule[Req, Res, Cfg] {
    cache := sync.Map{}

    type cacheEntry struct {
        result Res
        expiry time.Time
    }

    return NewRule[Req, Res, Cfg]().
        WhenFunc(rule.when).
        Resolve(func(ctx context.Context, req Req, cfg Cfg) (Res, error) {
            // Check cache
            if val, ok := cache.Load(req); ok {
                entry := val.(cacheEntry)
                if time.Now().Before(entry.expiry) {
                    return entry.result, nil
                }
                cache.Delete(req)
            }

            // Resolve and cache
            res, err := rule.resolve(ctx, req, cfg)
            if err == nil {
                cache.Store(req, cacheEntry{
                    result: res,
                    expiry: time.Now().Add(ttl),
                })
            }
            return res, err
        })
}
// ~30 lines

// Package documentation and examples
// ~10 lines
```

**Size**: +120 lines

### 5. Credential Provider Interface (+85 lines)

**New File**: `providers.go`

```go
// CredentialProvider fetches credentials/output based on a cache key
type CredentialProvider[Output any] interface {
    FetchCredentials(ctx context.Context, key CacheKey) (Output, error)
}
// ~5 lines

// DatabaseCredentialProvider fetches credentials from a database
type DatabaseCredentialProvider[Output any] struct {
    queryFunc   func(ctx context.Context, key CacheKey) (interface{}, error)
    extractFunc func(interface{}) (Output, error)
}

func NewDatabaseCredentialProvider[Output any](
    queryFunc func(ctx context.Context, key CacheKey) (interface{}, error),
    extractFunc func(interface{}) (Output, error),
) *DatabaseCredentialProvider[Output] {
    return &DatabaseCredentialProvider[Output]{
        queryFunc:   queryFunc,
        extractFunc: extractFunc,
    }
}

func (p *DatabaseCredentialProvider[Output]) FetchCredentials(ctx context.Context, key CacheKey) (Output, error) {
    data, err := p.queryFunc(ctx, key)
    if err != nil {
        return *new(Output), err
    }
    return p.extractFunc(data)
}
// ~30 lines

// EnvironmentCredentialProvider fetches credentials from environment variables
type EnvironmentCredentialProvider[Output any] struct {
    extractFunc func() (Output, error)
}

func NewEnvironmentCredentialProvider[Output any](
    extractFunc func() (Output, error),
) *EnvironmentCredentialProvider[Output] {
    return &EnvironmentCredentialProvider[Output]{
        extractFunc: extractFunc,
    }
}

func (p *EnvironmentCredentialProvider[Output]) FetchCredentials(ctx context.Context, key CacheKey) (Output, error) {
    return p.extractFunc()
}
// ~20 lines

// ChainedCredentialProvider tries providers in order until one succeeds
type ChainedCredentialProvider[Output any] struct {
    providers []CredentialProvider[Output]
}

func NewChainedCredentialProvider[Output any](
    providers ...CredentialProvider[Output],
) *ChainedCredentialProvider[Output] {
    return &ChainedCredentialProvider[Output]{
        providers: providers,
    }
}

func (p *ChainedCredentialProvider[Output]) FetchCredentials(ctx context.Context, key CacheKey) (Output, error) {
    var lastErr error
    for i, provider := range p.providers {
        output, err := provider.FetchCredentials(ctx, key)
        if err == nil {
            return output, nil
        }
        lastErr = fmt.Errorf("provider %d failed: %w", i, err)
    }
    return *new(Output), fmt.Errorf("all providers failed: %w", lastErr)
}
// ~30 lines
```

**Size**: +85 lines

### 6. Testing Utilities (+100 lines, optional)

**New File**: `testing/testing.go`

```go
package testing

// MockResolver for testing
type MockResolver[Req, Res, Cfg any] struct {
    ResolveFunc func(context.Context, Req, Cfg) (Res, error)
    calls       []Req
    mu          sync.Mutex
}

func NewMockResolver[Req, Res, Cfg any](
    resolveFunc func(context.Context, Req, Cfg) (Res, error),
) *MockResolver[Req, Res, Cfg] {
    return &MockResolver[Req, Res, Cfg]{
        ResolveFunc: resolveFunc,
    }
}

func (m *MockResolver[Req, Res, Cfg]) Resolve(ctx context.Context, req Req, cfg Cfg) (Res, error) {
    m.mu.Lock()
    m.calls = append(m.calls, req)
    m.mu.Unlock()
    return m.ResolveFunc(ctx, req, cfg)
}

func (m *MockResolver[Req, Res, Cfg]) Calls() []Req {
    m.mu.Lock()
    defer m.mu.Unlock()
    return append([]Req{}, m.calls...)
}

func (m *MockResolver[Req, Res, Cfg]) CallCount() int {
    m.mu.Lock()
    defer m.mu.Unlock()
    return len(m.calls)
}
// ~40 lines

// MockBuilder for testing
type MockBuilder[T, Output, Config any] struct {
    BuildFunc        func(context.Context, Output, Config) (T, error)
    ProviderTypeFunc func() ProviderType
    calls            []mockBuildCall[Output, Config]
    mu               sync.Mutex
}

type mockBuildCall[Output, Config any] struct {
    output Output
    config Config
}

func NewMockBuilder[T, Output, Config any](
    buildFunc func(context.Context, Output, Config) (T, error),
    providerType ProviderType,
) *MockBuilder[T, Output, Config] {
    return &MockBuilder[T, Output, Config]{
        BuildFunc:        buildFunc,
        ProviderTypeFunc: func() ProviderType { return providerType },
    }
}

func (m *MockBuilder[T, Output, Config]) Build(ctx context.Context, output Output, config Config) (T, error) {
    m.mu.Lock()
    m.calls = append(m.calls, mockBuildCall[Output, Config]{output: output, config: config})
    m.mu.Unlock()
    return m.BuildFunc(ctx, output, config)
}

func (m *MockBuilder[T, Output, Config]) ProviderType() ProviderType {
    return m.ProviderTypeFunc()
}

func (m *MockBuilder[T, Output, Config]) Calls() []mockBuildCall[Output, Config] {
    m.mu.Lock()
    defer m.mu.Unlock()
    return append([]mockBuildCall[Output, Config]{}, m.calls...)
}
// ~60 lines
```

**Size**: +100 lines (optional, separate package)

## Total Size Impact

### Core Package

| Component | Current | With Changes | Delta |
|-----------|---------|--------------|-------|
| Type renames | 445 | 445 | ±0 |
| Simplified builder | 445 | 430 | -15 |
| Generic cache key | 430 | 450 | +20 |
| Rule helpers | 450 | 570 | +120 |
| Credential providers | 570 | 655 | +85 |
| **Total Core** | **445** | **655** | **+210 (+47%)** |

### With Optional Testing Package

| Component | Lines |
|-----------|-------|
| Core package | 655 |
| Testing utilities | 100 |
| **Total** | **755 (+70%)** |

### Test Coverage

| Component | Current | Additions | Total |
|-----------|---------|-----------|-------|
| Existing tests | 925 | - | 925 |
| Helpers tests | - | +150 | 150 |
| Providers tests | - | +120 | 120 |
| **Total Tests** | **925** | **+270** | **1,195** |

## Comparison: Library Size vs Glue Code Savings

### Per-User Analysis

**One User**:
- Library cost: +190 lines (core additions, excluding testing)
- User savings: -388 lines of glue code
- **Net savings: 198 lines**

**Two Users**:
- Library cost: +190 lines (one-time)
- User savings: 2 × -388 = -776 lines
- **Net savings: 586 lines**

**Five Users**:
- Library cost: +190 lines (one-time)
- User savings: 5 × -388 = -1,940 lines
- **Net savings: 1,750 lines**

### Break-Even Point

**Break-even**: 1 user

With just ONE user of the library, the glue code savings (-388 lines) exceed the library additions (+190 lines).

## Alternative Approaches

### Minimal Essentials Only (+85 lines)

Only add the most impactful helpers:

```go
// helpers.go - Minimal version
MatchHint          // ~15 lines
FallbackChain      // ~20 lines
ChainedCredentialProvider  // ~50 lines
```

**Size**: +85 lines
**Glue code savings**: ~300 lines per user
**Net savings**: 215 lines for one user

### Modular Approach (±0 core bloat)

Separate optional helpers into sub-packages:

```
github.com/theopenlane/resolver/
├── resolver/          # Core (445 lines, unchanged)
│   ├── resolver.go
│   ├── rules.go
│   ├── context.go
│   └── hints.go
├── helpers/           # Optional helpers (205 lines)
│   ├── helpers.go     # Rule helpers (120 lines)
│   └── providers.go   # Credential providers (85 lines)
└── testing/           # Test utilities (100 lines)
    └── testing.go
```

**Benefits**:
- Core stays minimal (445 lines)
- Users opt into helpers via `import "resolver/helpers"`
- Can add more helpers without bloating core
- Clear separation of concerns

**Usage**:
```go
import (
    "github.com/theopenlane/resolver/resolver"
    "github.com/theopenlane/resolver/helpers"
)

// Core resolver
r := resolver.NewResolver[Req, Res, Cfg]()

// With helpers
r.AddRule(helpers.MatchHint(key, "value", resolveFunc))
```

## Recommendations

### Approach: Modular Package Structure

**Rationale**:
- Keeps core minimal for basic use cases
- Provides opt-in helpers for advanced users
- Allows library growth without core bloat
- Clear separation of essential vs convenience

### Implementation Plan

**Phase 1: Core Package** (445 lines)
- Type renames (Result, Output, Builder)
- Simplified builder interface (-15 lines)
- Generic CacheKey (+20 lines)
- Comprehensive tests
- **Total**: 450 lines

**Phase 2: Helpers Package** (205 lines)
- Rule helpers (+120 lines)
- Credential providers (+85 lines)
- Separate import path
- **Total**: 205 lines

**Phase 3: Testing Package** (100 lines)
- Mock implementations
- Test utilities
- Separate import path
- **Total**: 100 lines

### Documentation Priority

1. **Core Package**: Essential docs with examples
2. **Helpers Package**: Cookbook-style examples showing before/after
3. **Migration Guide**: How to reduce glue code
4. **Best Practices**: When to use core vs helpers

## Conclusion

**Recommended Approach**: Modular structure

**Core Package**:
- Size: 450 lines (+1%)
- Essential functionality only
- Zero bloat

**Helpers Package**:
- Size: 205 lines (optional)
- Saves ~388 lines of glue code per user
- Break-even: 1 user
- Net savings: 183 lines per user

**Testing Package**:
- Size: 100 lines (optional)
- Improves developer experience
- Not counted in "bloat"

**Total Impact**:
- Library: 450 (core) + 205 (helpers) = 655 lines
- User glue code: 548 → 160 lines (-71%)
- Net benefit: +388 lines saved per user (after -207 library cost)

The modular approach provides the best balance: minimal core for basic users, powerful helpers for advanced users, zero compromise on either end.
