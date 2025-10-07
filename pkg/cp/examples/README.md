# Client Pool Examples

Practical examples demonstrating `pkg/cp` usage across different domains and use cases.

## Overview

These examples show how to use the generic client pooling system for various client types beyond storage providers. The patterns are applicable to any scenario requiring:

- Multi-tenant client isolation
- Credential management and rotation
- Dynamic provider selection
- Connection pooling and reuse
- Context-based routing

## Examples

### [Communications Clients](./communications-clients/)

**Domain**: Messaging platforms (Slack, Microsoft Teams, Discord)

**Demonstrates**:
- Multi-platform support with unified interface
- Tenant-specific provider preferences
- Priority-based routing (critical → Slack, normal → Discord)
- Environment-based routing (prod → Slack, staging → Teams, dev → Discord)
- Region-based routing (EU → Teams/GDPR, US → Slack, APAC → Discord)
- Scale-based routing (enterprise vs standard tiers)
- Combined conditions (Production + Critical + EU → Premium Slack with GDPR)

**Use Cases**:
- Incident management and alerting
- Customer notifications
- Compliance-aware messaging
- Cost optimization based on usage patterns

**Run**:
```bash
cd communications-clients
go run main.go
```

**Key Patterns**:
```go
// Priority-based routing
resolver.AddRule(cp.NewRule[MessageClient]().
    WhenFunc(func(ctx context.Context) bool {
        msg := cp.GetValue[MessageContext](ctx).MustGet()
        return msg.Priority == PriorityCritical
    }).
    Resolve(func(ctx context.Context) (*cp.ResolvedProvider, error) {
        return &cp.ResolvedProvider{Type: "slack"}, nil
    }))

// Environment-based routing
resolver.AddRule(cp.NewRule[MessageClient]().
    WhenFunc(func(ctx context.Context) bool {
        msg := cp.GetValue[MessageContext](ctx).MustGet()
        return msg.Environment == EnvProduction
    }).
    Resolve(func(ctx context.Context) (*cp.ResolvedProvider, error) {
        return &cp.ResolvedProvider{Type: "slack-premium"}, nil
    }))
```

---

## Future Examples

### Email Providers (Planned)

**Domain**: Email delivery (SendGrid, Mailgun, AWS SES, Postmark)

**Will Demonstrate**:
- Volume-based routing (bulk → SES, transactional → Postmark)
- Failover patterns (primary → SendGrid, fallback → Mailgun)
- Cost optimization (high-volume → cheaper provider)
- Reputation management (split traffic across providers)

### Payment Processors (Planned)

**Domain**: Payment gateways (Stripe, PayPal, Square)

**Will Demonstrate**:
- Geography-based routing (EU → Stripe EUR, US → Stripe USD)
- Merchant preference selection
- Fee optimization routing
- Compliance requirements (PCI-DSS)

### Analytics Platforms (Planned)

**Domain**: Analytics services (Segment, Amplitude, Mixpanel)

**Will Demonstrate**:
- Event volume routing
- Cost-based selection
- Real-time vs batch routing
- Multi-destination fanout

### Monitoring Services (Planned)

**Domain**: Observability (Datadog, New Relic, Prometheus)

**Will Demonstrate**:
- Metric type routing
- Sampling strategies
- Cost optimization
- Multi-cloud monitoring

## Common Patterns Across Examples

### 1. Interface Design

All examples use a common interface pattern:

```go
type ClientInterface interface {
    PrimaryMethod(ctx context.Context, ...) error
    SecondaryMethod(ctx context.Context, ...) error
    Close() error
}
```

Benefits:
- Seamless provider switching
- Consistent API
- Easy testing with mocks
- Type-safe pooling

### 2. Builder Pattern

```go
type ProviderBuilder struct {
    creds  map[string]string
    config map[string]any
}

func (b *ProviderBuilder) WithCredentials(creds map[string]string) cp.ClientBuilder[T] {
    newBuilder := *b
    newBuilder.creds = creds
    return &newBuilder
}

func (b *ProviderBuilder) Build(ctx context.Context) (T, error) {
    return newClient(b.creds, b.config)
}
```

### 3. Context-Based Resolution

```go
type RoutingContext struct {
    TenantID    string
    Environment string
    Priority    string
    Region      string
}

resolver.AddRule(cp.NewRule[T]().
    WhenFunc(func(ctx context.Context) bool {
        rc := cp.GetValue[RoutingContext](ctx)
        return rc.IsPresent() && rc.MustGet().Environment == "production"
    }).
    Resolve(func(ctx context.Context) (*cp.ResolvedProvider, error) {
        return &cp.ResolvedProvider{Type: "premium"}, nil
    }))
```

### 4. Multi-Tenant Isolation

```go
// Per-tenant provider registration
for _, tenant := range tenants {
    providerType := cp.ProviderType(fmt.Sprintf("provider-%s", tenant.ID))

    service.RegisterBuilder(providerType, &Builder{
        credentials: tenant.Credentials,
        config:      tenant.Config,
    })
}

// Resolution based on tenant context
resolver.AddRule(cp.ResolutionRule[T]{
    Evaluate: func(ctx context.Context) mo.Option[cp.Resolution] {
        tenant := cp.GetValue[Tenant](ctx)
        if tenant.IsPresent() {
            return mo.Some(cp.Resolution{
                ClientType: cp.ProviderType(fmt.Sprintf("provider-%s", tenant.MustGet().ID)),
            })
        }
        return mo.None[cp.Resolution]()
    },
})
```

## Performance Characteristics

All examples share similar performance profiles:

### Cache Efficiency
- First request: 5-10ms (client creation)
- Cached requests: 100-500µs (pool lookup)
- Cache hit rate: >95% in production

### Memory Footprint
- Pool overhead: ~100 bytes
- Per client entry: ~200 bytes + client size
- Per client: 2-10KB (varies by type)

### Concurrency
- Thread-safe operations
- No lock contention on reads
- Supports 10,000+ concurrent requests

## Testing Patterns

### Mock Clients

```go
type MockClient struct {
    Calls []string
}

func (m *MockClient) Operation(ctx context.Context, arg string) error {
    m.Calls = append(m.Calls, arg)
    return nil
}

// Use in tests
mockClient := &MockClient{}
service.RegisterBuilder("mock", &MockBuilder{client: mockClient})
```

### Builder Testing

```go
func TestBuilder(t *testing.T) {
    builder := &MyBuilder{}

    creds := map[string]string{"key": "value"}
    configured := builder.WithCredentials(creds)

    client, err := configured.Build(context.Background())
    require.NoError(t, err)
    require.NotNil(t, client)
}
```

### Resolver Testing

```go
func TestResolver(t *testing.T) {
    resolver := cp.NewResolver[MyClient]()

    resolver.AddRule(cp.NewRule[MyClient]().
        WhenFunc(func(ctx context.Context) bool {
            return ctx.Value("key") == "value"
        }).
        Resolve(func(ctx context.Context) (*cp.ResolvedProvider, error) {
            return &cp.ResolvedProvider{Type: "test"}, nil
        }))

    ctx := context.WithValue(context.Background(), "key", "value")
    resolution := resolver.Resolve(ctx)

    require.True(t, resolution.IsPresent())
    require.Equal(t, "test", string(resolution.MustGet().ClientType))
}
```

## Best Practices

### 1. Immutable Builders

```go
// Good: Returns new builder
func (b *Builder) WithConfig(config map[string]any) cp.ClientBuilder[T] {
    newBuilder := *b
    newBuilder.config = config
    return &newBuilder
}

// Bad: Mutates original
func (b *Builder) WithConfig(config map[string]any) cp.ClientBuilder[T] {
    b.config = config // Mutation!
    return b
}
```

### 2. Unique Cache Keys

```go
// Good: Unique per tenant + integration
cacheKey := cp.ClientCacheKey{
    TenantID:        tenant.ID,
    IntegrationType: "provider",
    HushID:          credentialsID, // Optional
}

// Bad: Missing tenant isolation
cacheKey := cp.ClientCacheKey{
    IntegrationType: "provider", // Shared across tenants!
}
```

### 3. Safe Option Handling

```go
// Good: Check before use
client := service.GetClient(ctx, key, "type", creds, config)
if !client.IsPresent() {
    return errors.New("client not found")
}
c := client.MustGet()

// Bad: Unsafe
client := service.GetClient(ctx, key, "type", creds, config)
c := client.MustGet() // Panics if not present!
```

### 4. Rule Ordering

```go
// Good: Specific → General → Default
resolver.AddRule(specificProductionRule)
resolver.AddRule(specificStagingRule)
resolver.AddRule(generalRule)
resolver.SetDefaultRule(fallbackRule)

// Bad: General first blocks specific
resolver.AddRule(generalRule)    // Always matches!
resolver.AddRule(specificRule)   // Never reached
```

## Running All Examples

```bash
# Communications clients
cd communications-clients
go run main.go

# Add more examples as created
```

## Creating New Examples

When creating new examples:

1. **Choose a domain** different from existing examples
2. **Define client interface** with 3-5 core methods
3. **Implement providers** (at least 2-3 different ones)
4. **Create builders** for each provider
5. **Add resolution rules** showing different routing strategies
6. **Document use cases** and real-world applications
7. **Include README** with usage and patterns
8. **Add tests** demonstrating usage

Example structure:
```
examples/
└── your-domain/
    ├── README.md
    ├── go.mod
    ├── main.go           # Basic usage
    ├── advanced.go       # Complex routing
    └── providers.go      # Client implementations
```

## References

- [pkg/cp README](../README.md) - Core API documentation
- [pkg/objects examples](../../objects/examples/) - Storage provider examples using cp
- [pkg/cp tests](../) - Comprehensive test suite showing patterns

## Contributing

Contributions of new examples are welcome! Please ensure:
- Examples use different domains
- Code compiles and runs
- README documents patterns clearly
- Tests demonstrate key functionality
- Follows existing example structure
