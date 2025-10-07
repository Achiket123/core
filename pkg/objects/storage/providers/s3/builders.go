package s3

import (
	"context"

	"github.com/samber/mo"
	"github.com/theopenlane/core/pkg/cp"
	storage "github.com/theopenlane/core/pkg/objects/storage"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

// Builder creates S3 providers for the client pool
type Builder struct {
	credentials storage.ProviderCredentials
	options     *storage.ProviderOptions
	opts        []Option
}

// NewS3Builder creates a new S3Builder
func NewS3Builder() *Builder {
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

// WithOptions allows configuring provider-specific options
func (b *Builder) WithOptions(opts ...Option) *Builder {
	b.opts = append(b.opts, opts...)
	return b
}

// Build implements cp.ClientBuilder
func (b *Builder) Build(_ context.Context) (storagetypes.Provider, error) {
	if b.options == nil {
		b.options = storage.NewProviderOptions()
	}

	cfg := b.options.Clone()
	cfg.Credentials = b.credentials

	if cfg.Bucket == "" || cfg.Region == "" {
		return nil, ErrS3CredentialsRequired
	}

	provider, err := NewS3Provider(cfg, b.opts...)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

// ClientType implements cp.ClientBuilder
func (b *Builder) ClientType() cp.ProviderType {
	return cp.ProviderType(storagetypes.S3Provider)
}

// NewS3ProviderFromCredentials creates an S3 provider from provider credentials and optional configuration
func NewS3ProviderFromCredentials(credentials storage.ProviderCredentials, options *storage.ProviderOptions, opts ...Option) mo.Result[storagetypes.Provider] {
	cfg := storage.NewProviderOptions()
	if options != nil {
		cfg = options.Clone()
	}

	cfg.Credentials = credentials

	provider, err := NewS3Provider(cfg, opts...)
	if err != nil {
		return mo.Err[storagetypes.Provider](err)
	}

	return mo.Ok[storagetypes.Provider](provider)
}
