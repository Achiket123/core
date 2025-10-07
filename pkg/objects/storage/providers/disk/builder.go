package disk

import (
	"context"

	"github.com/samber/mo"
	"github.com/theopenlane/core/pkg/cp"
	storage "github.com/theopenlane/core/pkg/objects/storage"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

// Builder creates disk providers for the client pool
type Builder struct {
	credentials storage.ProviderCredentials
	options     *storage.ProviderOptions
}

// NewDiskBuilder creates a new Builder
func NewDiskBuilder() *Builder {
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
func (b *Builder) Build(context.Context) (storagetypes.Provider, error) {
	if b.options == nil {
		b.options = storage.NewProviderOptions()
	}

	cfg := b.options.Clone()
	cfg.Credentials = b.credentials

	if cfg.Bucket == "" {
		cfg.Bucket = "./storage"
	}

	if cfg.LocalURL == "" {
		cfg.LocalURL = b.credentials.Endpoint
	}

	provider, err := NewDiskProvider(cfg)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

// ClientType implements cp.ClientBuilder
func (b *Builder) ClientType() cp.ProviderType {
	return cp.ProviderType(storagetypes.DiskProvider)
}

// NewDiskProviderFromCredentials creates a disk provider from credential struct
func NewDiskProviderFromCredentials(credentials storage.ProviderCredentials) mo.Result[storagetypes.Provider] {
	options := storage.NewProviderOptions(
		storage.WithCredentials(credentials),
		storage.WithBucket("./storage"),
		storage.WithLocalURL(credentials.Endpoint),
	)
	provider, err := NewDiskProvider(options)
	if err != nil {
		return mo.Err[storagetypes.Provider](err)
	}

	return mo.Ok[storagetypes.Provider](provider)
}
