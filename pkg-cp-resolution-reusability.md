# Resolver/Rules Pattern Reusability Analysis

## Executive Summary

Setting aside the client pooling aspect of `pkg/cp`, the **resolver/rules construct is MORE broadly applicable** than the pooling mechanism. This pattern solves a common problem: context-aware decision making with configurable, ordered evaluation rules.

**Key Finding**: The resolver pattern addresses decision-making scenarios that appear throughout codebases but are typically implemented as ad-hoc if/else chains or switch statements.

## What the Resolver Pattern Provides

```go
type Resolver[Request, Result, Config] struct {
    rules []ResolutionRule[Request, Result, Config]
}
```

The pattern offers:
- **First-match semantics**: Rules evaluated in order, first matching rule wins
- **Type safety**: Full generics support without reflection
- **Fluent DSL**: Chainable rule construction
- **Context integration**: Built-in context.Context support
- **Hint system**: Structured metadata passing via HintKey[T]
- **Testability**: Rules can be unit tested independently

## Generic Use Cases Beyond Storage

### 1. Authorization Decision Engine

```go
type AuthRequest struct {
    UserID     string
    ResourceID string
    Action     string
}

type AuthDecision struct {
    Allowed bool
    Reason  string
}

resolver := cp.NewResolver[AuthRequest, AuthDecision, EmptyConfig]()
resolver.AddRule(
    cp.NewRule[AuthRequest, AuthDecision, EmptyConfig]().
        WhenFunc(func(ctx context.Context, req AuthRequest, _ EmptyConfig) bool {
            return isSystemAdmin(ctx, req.UserID)
        }).
        Resolve(func(ctx context.Context, req AuthRequest, _ EmptyConfig) (AuthDecision, error) {
            return AuthDecision{Allowed: true, Reason: "system admin bypass"}, nil
        }),
)
```

### 2. Workflow State Machine

```go
type WorkflowTransition struct {
    FromState string
    ToState   string
    Actor     string
}

type TransitionResult struct {
    Allowed      bool
    NextState    string
    RequiredApprovals []string
}

// Rules determine valid state transitions based on actor role
resolver.AddRule(
    cp.NewRule[WorkflowTransition, TransitionResult, WorkflowConfig]().
        WhenFunc(func(ctx context.Context, t WorkflowTransition, _ WorkflowConfig) bool {
            return t.FromState == "draft" && t.ToState == "review"
        }).
        Resolve(func(ctx context.Context, t WorkflowTransition, cfg WorkflowConfig) (TransitionResult, error) {
            return TransitionResult{
                Allowed:   true,
                NextState: "review",
                RequiredApprovals: cfg.ReviewApprovers,
            }, nil
        }),
)
```

### 3. Notification Routing

```go
type NotificationEvent struct {
    EventType string
    Severity  string
    UserID    string
}

type RoutingDecision struct {
    Channels []string // email, sms, slack, pagerduty
    Priority string
}

// Route critical events to multiple channels
resolver.AddRule(
    cp.NewRule[NotificationEvent, RoutingDecision, NotificationConfig]().
        WhenFunc(func(ctx context.Context, evt NotificationEvent, _ NotificationConfig) bool {
            return evt.Severity == "critical"
        }).
        Resolve(func(ctx context.Context, evt NotificationEvent, cfg NotificationConfig) (RoutingDecision, error) {
            return RoutingDecision{
                Channels: []string{"email", "sms", "pagerduty"},
                Priority: "immediate",
            }, nil
        }),
)
```

### 4. Rate Limiting Strategy

```go
type RateLimitRequest struct {
    UserID   string
    Endpoint string
    UserTier string
}

type RateLimitDecision struct {
    RequestsPerMinute int
    BurstSize         int
}

// Different limits for different user tiers
resolver.AddRule(
    cp.NewRule[RateLimitRequest, RateLimitDecision, RateLimitConfig]().
        WhenFunc(func(ctx context.Context, req RateLimitRequest, _ RateLimitConfig) bool {
            return req.UserTier == "enterprise"
        }).
        Resolve(func(ctx context.Context, req RateLimitRequest, _ RateLimitConfig) (RateLimitDecision, error) {
            return RateLimitDecision{
                RequestsPerMinute: 10000,
                BurstSize:         1000,
            }, nil
        }),
)
```

### 5. Validation Pipeline

```go
type ValidationRequest struct {
    Entity     interface{}
    EntityType string
    Operation  string
}

type ValidationResult struct {
    Valid  bool
    Errors []string
}

// Different validation rules for create vs update
resolver.AddRule(
    cp.NewRule[ValidationRequest, ValidationResult, ValidationConfig]().
        WhenFunc(func(ctx context.Context, req ValidationRequest, _ ValidationConfig) bool {
            return req.Operation == "create" && req.EntityType == "user"
        }).
        Resolve(func(ctx context.Context, req ValidationRequest, _ ValidationConfig) (ValidationResult, error) {
            // User creation validation
            return validateUserCreation(req.Entity)
        }),
)
```

### 6. Retry Policy Selection

```go
type RetryRequest struct {
    Operation string
    Error     error
    Attempt   int
}

type RetryDecision struct {
    ShouldRetry bool
    Backoff     time.Duration
    MaxAttempts int
}

// Exponential backoff for transient errors
resolver.AddRule(
    cp.NewRule[RetryRequest, RetryDecision, RetryConfig]().
        WhenFunc(func(ctx context.Context, req RetryRequest, _ RetryConfig) bool {
            return isTransientError(req.Error) && req.Attempt < 5
        }).
        Resolve(func(ctx context.Context, req RetryRequest, _ RetryConfig) (RetryDecision, error) {
            return RetryDecision{
                ShouldRetry: true,
                Backoff:     time.Duration(math.Pow(2, float64(req.Attempt))) * time.Second,
                MaxAttempts: 5,
            }, nil
        }),
)
```

### 7. Log Level Determination

```go
type LogRequest struct {
    Message  string
    Module   string
    UserID   string
}

type LogDecision struct {
    Level      string
    ShouldLog  bool
    Redactions []string
}

// Verbose logging for specific users during debugging
resolver.AddRule(
    cp.NewRule[LogRequest, LogDecision, LogConfig]().
        WhenFunc(func(ctx context.Context, req LogRequest, cfg LogConfig) bool {
            return contains(cfg.DebugUsers, req.UserID)
        }).
        Resolve(func(ctx context.Context, req LogRequest, _ LogConfig) (LogDecision, error) {
            return LogDecision{
                Level:     "debug",
                ShouldLog: true,
            }, nil
        }),
)
```

### 8. Data Export Format Selection

```go
type ExportRequest struct {
    DataType    string
    RecordCount int
    UserAgent   string
}

type ExportDecision struct {
    Format      string // csv, json, parquet
    Compression bool
    Streaming   bool
}

// Large exports use streaming parquet
resolver.AddRule(
    cp.NewRule[ExportRequest, ExportDecision, ExportConfig]().
        WhenFunc(func(ctx context.Context, req ExportRequest, _ ExportConfig) bool {
            return req.RecordCount > 100000
        }).
        Resolve(func(ctx context.Context, req ExportRequest, _ ExportConfig) (ExportDecision, error) {
            return ExportDecision{
                Format:      "parquet",
                Compression: true,
                Streaming:   true,
            }, nil
        }),
)
```

### 9. Cache Strategy Selection

```go
type CacheRequest struct {
    Key        string
    DataType   string
    AccessFreq int
}

type CacheDecision struct {
    Backend string // redis, memcached, local
    TTL     time.Duration
}

// Hot data goes to local cache
resolver.AddRule(
    cp.NewRule[CacheRequest, CacheDecision, CacheConfig]().
        WhenFunc(func(ctx context.Context, req CacheRequest, _ CacheConfig) bool {
            return req.AccessFreq > 1000
        }).
        Resolve(func(ctx context.Context, req CacheRequest, _ CacheConfig) (CacheDecision, error) {
            return CacheDecision{
                Backend: "local",
                TTL:     5 * time.Minute,
            }, nil
        }),
)
```

### 10. Database Shard Selection

```go
type ShardRequest struct {
    TenantID   string
    RecordType string
}

type ShardDecision struct {
    ShardID   string
    ReadOnly  bool
    Replicas  []string
}

// Route specific tenants to dedicated shards
resolver.AddRule(
    cp.NewRule[ShardRequest, ShardDecision, ShardConfig]().
        WhenFunc(func(ctx context.Context, req ShardRequest, cfg ShardConfig) bool {
            return contains(cfg.DedicatedShardTenants, req.TenantID)
        }).
        Resolve(func(ctx context.Context, req ShardRequest, cfg ShardConfig) (ShardDecision, error) {
            return ShardDecision{
                ShardID:  "dedicated-" + req.TenantID,
                ReadOnly: false,
                Replicas: cfg.DedicatedReplicas[req.TenantID],
            }, nil
        }),
)
```

## Use Cases in the Core Repository

### 1. Document Approval Workflows

**Location**: `internal/ent/hooks/` or workflow management

**Current Problem**: Approval logic scattered across multiple functions with nested conditionals.

**Resolver Application**:
```go
type ApprovalRequest struct {
    DocumentID string
    UserID     string
    UserRoles  []string
    CurrentState string
}

type ApprovalDecision struct {
    RequiresApproval bool
    Approvers        []string
    AutoApprove      bool
}

// System admins auto-approve
resolver.AddRule(
    cp.NewRule[ApprovalRequest, ApprovalDecision, ApprovalConfig]().
        WhenFunc(func(ctx context.Context, req ApprovalRequest, _ ApprovalConfig) bool {
            return contains(req.UserRoles, "system-admin")
        }).
        Resolve(func(ctx context.Context, req ApprovalRequest, _ ApprovalConfig) (ApprovalDecision, error) {
            return ApprovalDecision{AutoApprove: true}, nil
        }),
)
```

### 2. Risk Assessment Routing

**Location**: Risk scoring or compliance modules

**Current Problem**: Complex routing logic for risk assessments based on score, industry, region.

**Resolver Application**:
```go
type RiskRequest struct {
    Score      float64
    Industry   string
    Region     string
}

type RiskDecision struct {
    ReviewLevel string
    Assignees   []string
    SLA         time.Duration
}

// High-risk regulated industries
resolver.AddRule(
    cp.NewRule[RiskRequest, RiskDecision, RiskConfig]().
        WhenFunc(func(ctx context.Context, req RiskRequest, _ RiskConfig) bool {
            return req.Score > 0.8 && contains(regulatedIndustries, req.Industry)
        }).
        Resolve(func(ctx context.Context, req RiskRequest, _ RiskConfig) (RiskDecision, error) {
            return RiskDecision{
                ReviewLevel: "senior-compliance",
                SLA:         4 * time.Hour,
            }, nil
        }),
)
```

### 3. Module-Specific Access Control

**Location**: `internal/ent/privacy/rule/modules.go`

**Current Problem**: Feature access control with 12+ bypass conditions (system admin, privacy allow, org creation, trust center anon, etc.)

**Resolver Application**:
```go
type FeatureAccessRequest struct {
    UserID       string
    OrgID        string
    FeatureName  string
    IsSystemOp   bool
    IsOrgCreation bool
}

type FeatureAccessDecision struct {
    Allowed         bool
    Reason          string
    MissingFeature  *models.OrgModule
}

// System operations bypass
resolver.AddRule(
    cp.NewRule[FeatureAccessRequest, FeatureAccessDecision, EmptyConfig]().
        WhenFunc(func(ctx context.Context, req FeatureAccessRequest, _ EmptyConfig) bool {
            return req.IsSystemOp
        }).
        Resolve(func(ctx context.Context, req FeatureAccessRequest, _ EmptyConfig) (FeatureAccessDecision, error) {
            return FeatureAccessDecision{Allowed: true, Reason: "system operation"}, nil
        }),
)
```

### 4. Tag Suggestion Engine

**Location**: Tag management or metadata modules

**Current Problem**: Tag suggestion logic based on entity type, existing tags, user history.

**Resolver Application**:
```go
type TagSuggestionRequest struct {
    EntityType   string
    ExistingTags []string
    UserID       string
}

type TagSuggestionDecision struct {
    SuggestedTags []string
    Confidence    float64
}

// Entity-specific tag templates
resolver.AddRule(
    cp.NewRule[TagSuggestionRequest, TagSuggestionDecision, TagConfig]().
        WhenFunc(func(ctx context.Context, req TagSuggestionRequest, cfg TagConfig) bool {
            return req.EntityType == "compliance-document"
        }).
        Resolve(func(ctx context.Context, req TagSuggestionRequest, cfg TagConfig) (TagSuggestionDecision, error) {
            return TagSuggestionDecision{
                SuggestedTags: cfg.ComplianceTemplateTags,
                Confidence:    0.9,
            }, nil
        }),
)
```

### 5. Audit Log Detail Level

**Location**: Audit logging infrastructure

**Current Problem**: Determining what level of detail to log based on entity type, operation, compliance requirements.

**Resolver Application**:
```go
type AuditRequest struct {
    EntityType string
    Operation  string
    UserID     string
    OrgID      string
}

type AuditDecision struct {
    LogLevel      string
    IncludeDiff   bool
    IncludeParams bool
    Retention     time.Duration
}

// Compliance-sensitive operations
resolver.AddRule(
    cp.NewRule[AuditRequest, AuditDecision, AuditConfig]().
        WhenFunc(func(ctx context.Context, req AuditRequest, _ AuditConfig) bool {
            return contains(complianceEntities, req.EntityType)
        }).
        Resolve(func(ctx context.Context, req AuditRequest, _ AuditConfig) (AuditDecision, error) {
            return AuditDecision{
                LogLevel:      "verbose",
                IncludeDiff:   true,
                IncludeParams: true,
                Retention:     7 * 365 * 24 * time.Hour, // 7 years
            }, nil
        }),
)
```

### 6. Scheduled Job Priority

**Location**: Job queue or background processing

**Current Problem**: Determining job priority and resource allocation based on job type, tenant, time of day.

**Resolver Application**:
```go
type JobRequest struct {
    JobType    string
    TenantID   string
    SubmitTime time.Time
}

type JobDecision struct {
    Priority      int
    MaxConcurrent int
    Timeout       time.Duration
}

// Enterprise tenants get priority
resolver.AddRule(
    cp.NewRule[JobRequest, JobDecision, JobConfig]().
        WhenFunc(func(ctx context.Context, req JobRequest, cfg JobConfig) bool {
            return contains(cfg.EnterpriseTenants, req.TenantID)
        }).
        Resolve(func(ctx context.Context, req JobRequest, _ JobConfig) (JobDecision, error) {
            return JobDecision{
                Priority:      10,
                MaxConcurrent: 5,
                Timeout:       30 * time.Minute,
            }, nil
        }),
)
```

## Comparison to Alternatives

### vs. Simple If/Else Chains
**Resolver Advantages**:
- Testable rules in isolation
- Clear precedence via ordering
- Easier to add/remove/reorder rules
- Context and hints built-in

**If/Else Disadvantages**:
- Difficult to test individual conditions
- Order dependencies hidden
- Hard to maintain as complexity grows

### vs. Strategy Pattern
**Resolver Advantages**:
- Multiple strategies can apply (first-match)
- Built-in context passing
- Fluent rule construction
- Generic without interface{} casting

**Strategy Pattern Disadvantages**:
- Typically single strategy selection
- Manual context threading
- More boilerplate per strategy

### vs. Chain of Responsibility
**Resolver Advantages**:
- Type-safe with generics
- First-match semantics built-in
- Hint system for metadata
- Less boilerplate

**Chain of Responsibility Disadvantages**:
- Manual chain management
- Often uses interface{} for flexibility
- Each handler needs next pointer

### vs. Rules Engines (drools, govaluate)
**Resolver Advantages**:
- Type-safe at compile time
- Native Go code (no DSL parsing)
- Better IDE support
- Simpler for Go developers

**Rules Engines Disadvantages**:
- Runtime overhead
- Learning curve for DSL
- Type safety lost with string expressions

## Recommendations

### Extraction as Standalone Package

**Suggested Name**: `pkg/resolver` or standalone `github.com/theopenlane/resolver`

**What to Extract**:
1. Core resolver types and logic
2. RuleBuilder with fluent API
3. Context helpers (WithValue, GetValue, HintKey)
4. HintSet for metadata passing

**What to Leave Behind**:
- Client pooling (pkg/cp specific use case)
- Storage-specific types (ClientBuilder, ClientService)

**Package Structure**:
```
pkg/resolver/
├── resolver.go      # Core Resolver[Request, Result, Config]
├── rules.go         # RuleBuilder, Rule types
├── context.go       # Context helpers
├── hints.go         # HintKey, HintSet
├── doc.go           # Package documentation
└── examples_test.go # Example use cases
```

### Documentation Emphasis

Focus documentation on:
1. Context-aware decision making
2. First-match rule evaluation
3. Type-safe with Go 1.18+ generics
4. Common patterns (authorization, workflows, routing, validation)
5. Migration from if/else chains

### Conclusion

The resolver/rules construct is **highly reusable** for any scenario requiring context-aware decision making with ordered evaluation. It addresses a genuine gap in the Go ecosystem: a type-safe, generic, testable rules engine without the overhead of string-based DSLs.

**Reusability Score: 9/10** (even higher than the full pkg/cp package)

The pattern is applicable to:
- Any codebase with complex conditional logic
- Systems requiring configurable decision making
- Multi-tenant applications with varying rules per tenant
- Workflow engines
- Policy enforcement systems
- Routing and selection logic
