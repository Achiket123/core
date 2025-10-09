package resolver

import (
	"context"
	"fmt"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/pkg/eddy"
	"github.com/theopenlane/core/pkg/objects/storage"
	dbprovider "github.com/theopenlane/core/pkg/objects/storage/providers/database"
	"github.com/theopenlane/core/pkg/objects/storage/providers/disk"
	r2provider "github.com/theopenlane/core/pkg/objects/storage/providers/r2"
	s3provider "github.com/theopenlane/core/pkg/objects/storage/providers/s3"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/utils/contextx"
)

type Option func(*serviceOptions)

type serviceOptions struct {
	tokenManagerFunc func() *tokens.TokenManager
	tokenAudience    string
	tokenIssuer      string
}

// providerResolver simplifies references to the eddy resolver used for object providers
type providerResolver = eddy.Resolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]

// providerClientService simplifies references to the eddy client service used for object providers
type providerClientService = eddy.ClientService[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]

// WithPresignConfig configures presigned URL token generation for providers that support it.
func WithPresignConfig(tokenManager func() *tokens.TokenManager, issuer, audience string) Option {
	return func(opts *serviceOptions) {
		opts.tokenManagerFunc = tokenManager
		opts.tokenIssuer = issuer
		opts.tokenAudience = audience
	}
}

// NewServiceFromConfig constructs a storage service complete with resolver rules derived from runtime configuration.
func NewServiceFromConfig(config storage.ProviderConfig, opts ...Option) *objects.Service {
	runtime := serviceOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&runtime)
		}
	}

	clientService, resolver := buildWithRuntime(config, runtime)

	service := objects.NewService(objects.Config{
		Resolver:       resolver,
		ClientService:  clientService,
		ValidationFunc: objects.MimeTypeValidator,
		TokenManager:   runtime.tokenManagerFunc,
		TokenIssuer:    runtime.tokenIssuer,
		TokenAudience:  runtime.tokenAudience,
	})

	return service
}

// Build constructs the cp client service and provider resolver from runtime configuration.
func Build(config storage.ProviderConfig) (*providerClientService, *providerResolver) {
	return buildWithRuntime(config, serviceOptions{})
}

func buildWithRuntime(config storage.ProviderConfig, runtime serviceOptions) (*providerClientService, *providerResolver) {
	pool := eddy.NewClientPool[storage.Provider](objects.DefaultClientPoolTTL)
	clientService := eddy.NewClientService(pool, eddy.WithConfigClone[
		storage.Provider,
		storage.ProviderCredentials](cloneProviderOptions))

	// Create builder instances
	s3Builder := s3provider.NewS3Builder()
	r2Builder := r2provider.NewR2Builder()
	diskBuilder := disk.NewDiskBuilder()
	dbBuilder := dbprovider.NewBuilder()
	if runtime.tokenManagerFunc != nil {
		if tm := runtime.tokenManagerFunc(); tm != nil {
			dbBuilder = dbBuilder.WithOptions(
				dbprovider.WithTokenManager(tm),
				dbprovider.WithTokenClaims(runtime.tokenIssuer, runtime.tokenAudience),
			)
		}
	}

	// Create resolver and configure rules with builders
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	configureProviderRules(resolver, config, s3Builder, r2Builder, diskBuilder, dbBuilder)

	return clientService, resolver
}

func systemOwnedQueryContext(ctx context.Context, entClient *ent.Client) context.Context {
	user := &auth.AuthenticatedUser{
		SubjectID:          "system-storage-resolver",
		SubjectName:        "System Storage Resolver",
		AuthenticationType: auth.APITokenAuthentication,
		IsSystemAdmin:      true,
	}

	ctx = auth.WithAuthenticatedUser(ctx, user)
	ctx = auth.WithSystemAdminContext(ctx, user)
	ctx = contextx.With(ctx, auth.OrganizationCreationContextKey{})
	ctx = privacy.DecisionContext(ctx, privacy.Allow)
	return ent.NewContext(ctx, entClient)
}

func stringValue(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case fmt.Stringer:
		return v.String(), true
	case []byte:
		return string(v), true
	default:
		if v == nil {
			return "", false
		}
		return fmt.Sprintf("%v", v), true
	}
}

func cloneProviderOptions(in *storage.ProviderOptions) *storage.ProviderOptions {
	if in == nil {
		return nil
	}
	return in.Clone()
}
