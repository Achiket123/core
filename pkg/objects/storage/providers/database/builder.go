package database

import (
	"context"

	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/pkg/cp"
	storage "github.com/theopenlane/core/pkg/objects/storage"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

// Option configures the database provider builder.
type Option func(*Builder)

// WithTokenManager supplies the token manager used for presigned URL generation.
func WithTokenManager(tm *tokens.TokenManager) Option {
	return func(b *Builder) {
		b.tokenManager = tm
	}
}

// WithTokenClaims configures issuer and audience values for presigned tokens.
func WithTokenClaims(issuer, audience string) Option {
	return func(b *Builder) {
		b.tokenIssuer = issuer
		b.tokenAudience = audience
	}
}

// Builder creates database providers for the client pool.
type Builder struct {
	credentials   storage.ProviderCredentials
	options       *storage.ProviderOptions
	tokenManager  *tokens.TokenManager
	tokenAudience string
	tokenIssuer   string
}

// NewBuilder returns a new database provider builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// WithCredentials implements cp.ClientBuilder.
func (b *Builder) WithCredentials(credentials storage.ProviderCredentials) cp.ClientBuilder[storagetypes.Provider, storage.ProviderCredentials, *storage.ProviderOptions] {
	b.credentials = credentials
	return b
}

// WithConfig implements cp.ClientBuilder.
func (b *Builder) WithConfig(config *storage.ProviderOptions) cp.ClientBuilder[storagetypes.Provider, storage.ProviderCredentials, *storage.ProviderOptions] {
	if config == nil {
		b.options = storage.NewProviderOptions()
	} else {
		b.options = config.Clone()
	}

	return b
}

// WithOptions allows applying builder-specific options.
func (b *Builder) WithOptions(opts ...Option) *Builder {
	for _, opt := range opts {
		if opt != nil {
			opt(b)
		}
	}

	return b
}

// Build implements cp.ClientBuilder.
func (b *Builder) Build(_ context.Context) (storagetypes.Provider, error) {
	if b.options == nil {
		b.options = storage.NewProviderOptions()
	}

	provider := &Provider{
		options:       b.options.Clone(),
		tokenManager:  b.tokenManager,
		tokenAudience: b.tokenAudience,
		tokenIssuer:   b.tokenIssuer,
	}

	return provider, nil
}

// ClientType implements cp.ClientBuilder.
func (b *Builder) ClientType() cp.ProviderType {
	return cp.ProviderType(storage.DatabaseProvider)
}
