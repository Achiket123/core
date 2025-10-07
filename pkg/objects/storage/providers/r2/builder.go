package r2

import (
	"context"

	"github.com/samber/mo"
	"github.com/theopenlane/core/pkg/cp"
	storage "github.com/theopenlane/core/pkg/objects/storage"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

// Builder creates R2 providers for the client pool
type Builder struct {
	credentials storage.ProviderCredentials
	options     *storage.ProviderOptions
}

// NewR2Builder creates a new R2Builder
func NewR2Builder() *Builder {
	return &Builder{}
}

// WithCredentials implements cp.ClientBuilder
func (b *Builder) WithCredentials(credentials storage.ProviderCredentials) cp.ClientBuilder[storagetypes.Provider, storage.ProviderCredentials, *storage.ProviderOptions] {
	b.credentials = credentials
	return b
}

// WithConfig implements cp.ClientBuilder
func (b *Builder) WithConfig(config *storage.ProviderOptions) cp.ClientBuilder[storagetypes.Provider, storage.ProviderCredentials, *storage.ProviderOptions] {
	if config == nil {
		b.options = storage.NewProviderOptions()
	} else {
		b.options = config.Clone()
	}
	return b
}

// Build implements cp.ClientBuilder
func (b *Builder) Build(_ context.Context) (storagetypes.Provider, error) {
	if b.options == nil {
		b.options = storage.NewProviderOptions()
	}

	cfg := b.options.Clone()
	cfg.Credentials = b.credentials

	if cfg.Bucket == "" || cfg.Credentials.AccountID == "" || cfg.Credentials.AccessKeyID == "" || cfg.Credentials.SecretAccessKey == "" {
		return nil, ErrR2CredentialsRequired
	}

	return NewR2Provider(cfg)
}

// ClientType implements cp.ClientBuilder
func (b *Builder) ClientType() cp.ProviderType {
	return cp.ProviderType(storagetypes.R2Provider)
}

// NewR2ProviderFromCredentials creates an R2 provider using the supplied credentials and options
func NewR2ProviderFromCredentials(credentials storage.ProviderCredentials, options *storage.ProviderOptions) mo.Result[storagetypes.Provider] {
	cfg := storage.NewProviderOptions()
	if options != nil {
		cfg = options.Clone()
	}
	cfg.Credentials = credentials

	provider, err := NewR2Provider(cfg)
	if err != nil {
		return mo.Err[storagetypes.Provider](err)
	}

	return mo.Ok[storagetypes.Provider](provider)
}
