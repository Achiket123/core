package cp

import (
	"context"
	"sync"
	"time"
)

// ProviderType represents a provider type identifier
type ProviderType string

// ClientBuilder builds client instances with credentials and configuration
type ClientBuilder[T any, Creds any, Conf any] interface {
	WithCredentials(credentials Creds) ClientBuilder[T, Creds, Conf]
	WithConfig(config Conf) ClientBuilder[T, Creds, Conf]
	Build(ctx context.Context) (T, error)
	ClientType() ProviderType
}

// ClientCacheKey uniquely identifies a cached client
type ClientCacheKey struct {
	TenantID        string
	IntegrationType string
	HushID          string
	IntegrationID   string
}

// ClientEntry wraps a client instance with expiration metadata
type ClientEntry[T any] struct {
	Client     T
	Expiration time.Time
}

// ClientPool holds cached client instances with TTL expiration
type ClientPool[T any] struct {
	mu      sync.RWMutex
	clients map[ClientCacheKey]*ClientEntry[T]
	ttl     time.Duration
}

// ClientService manages client builders and provides cached client instances
type ClientService[T any, Creds any, Conf any] struct {
	pool           *ClientPool[T]
	builders       map[ProviderType]ClientBuilder[T, Creds, Conf]
	mu             sync.RWMutex
	credentialCopy func(Creds) Creds
	configCopy     func(Conf) Conf
}

type ClientOption[T any, Creds any, Conf any] func(*ClientService[T, Creds, Conf])
