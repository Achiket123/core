package cp

import (
	"context"

	"github.com/samber/mo"
)

// ServiceOption configures a ClientService
type ServiceOption[T any, Creds any, Conf any] func(*ClientService[T, Creds, Conf])

// WithCredentialClone sets the credential cloning function for defensive copying
func WithCredentialClone[T any, Creds any, Conf any](cloneFn func(Creds) Creds) ServiceOption[T, Creds, Conf] {
	return func(s *ClientService[T, Creds, Conf]) {
		s.credentialCopy = cloneFn
	}
}

// WithConfigClone sets the config cloning function for defensive copying
func WithConfigClone[T any, Creds any, Conf any](cloneFn func(Conf) Conf) ServiceOption[T, Creds, Conf] {
	return func(s *ClientService[T, Creds, Conf]) {
		s.configCopy = cloneFn
	}
}

// NewClientService creates a new client service with the specified pool
func NewClientService[T any, Creds any, Conf any](pool *ClientPool[T], opts ...ServiceOption[T, Creds, Conf]) *ClientService[T, Creds, Conf] {
	s := &ClientService[T, Creds, Conf]{
		pool:     pool,
		builders: make(map[ProviderType]ClientBuilder[T, Creds, Conf]),
		credentialCopy: func(c Creds) Creds {
			return c
		},
		configCopy: func(c Conf) Conf {
			return c
		},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// RegisterBuilder registers a client builder for a specific client type
func (s *ClientService[T, Creds, Conf]) RegisterBuilder(clientType ProviderType, builder ClientBuilder[T, Creds, Conf]) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.builders[clientType] = builder
}

// GetClient retrieves a client from cache or builds a new one
func (s *ClientService[T, Creds, Conf]) GetClient(ctx context.Context, key ClientCacheKey, clientType ProviderType, credentials Creds, config Conf) mo.Option[T] {
	// Try cache first
	if cached := s.pool.GetClient(key); cached.IsPresent() {
		return cached
	}

	// Build new client
	s.mu.RLock()
	builderPtr, exists := s.builders[clientType]
	s.mu.RUnlock()

	if !exists {
		return mo.None[T]()
	}

	// creates a defensive copy of the credentials and config
	// to avoid potential side effects from external modifications
	// after the client has been built
	client, err := builderPtr.
		WithCredentials(s.credentialCopy(credentials)).
		WithConfig(s.configCopy(config)).
		Build(ctx)
	if err != nil {
		return mo.None[T]()
	}

	// Cache the new client
	s.pool.SetClient(key, client)

	return mo.Some(client)
}

// Pool returns the underlying client pool
func (s *ClientService[T, Creds, Conf]) Pool() *ClientPool[T] {
	return s.pool
}
