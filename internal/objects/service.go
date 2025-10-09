package objects

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
	"github.com/theopenlane/core/pkg/cp"
	"github.com/theopenlane/core/pkg/objects/storage"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/utils/ulids"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	defaultPresignedURLDuration = 10 * time.Minute
	defaultDatabaseBucket       = "default"
)

// Service orchestrates storage operations using cp provider resolution
type Service struct {
	resolver        *cp.Resolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]
	clientService   *cp.ClientService[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]
	objectService   *storage.ObjectService
	tokenProvider   func() *tokens.TokenManager
	tokenIssuer     string
	tokenAudience   string
	downloadSecrets sync.Map
}

// Config holds configuration for creating a new Service
type Config struct {
	Resolver       *cp.Resolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]
	ClientService  *cp.ClientService[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]
	ValidationFunc storage.ValidationFunc
	TokenManager   func() *tokens.TokenManager
	TokenIssuer    string
	TokenAudience  string
}

// NewService creates a new storage orchestration service
func NewService(cfg Config) *Service {
	objectService := storage.NewObjectService()

	// Configure validation if provided
	if cfg.ValidationFunc != nil {
		objectService = objectService.WithValidation(cfg.ValidationFunc)
	}

	return &Service{
		resolver:      cfg.Resolver,
		clientService: cfg.ClientService,
		objectService: objectService,
		tokenProvider: cfg.TokenManager,
		tokenIssuer:   cfg.TokenIssuer,
		tokenAudience: cfg.TokenAudience,
	}
}

// SetObjectService allows overriding the internal object service (useful for dev mode)
func (s *Service) SetObjectService(objectService *storage.ObjectService) {
	s.objectService = objectService
}

// Upload uploads a file using provider resolution
func (s *Service) Upload(ctx context.Context, reader io.Reader, opts *storage.UploadOptions) (*storage.File, error) {
	provider, err := s.resolveUploadProvider(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Upload the file
	file, err := s.objectService.Upload(ctx, provider, reader, opts)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// Download downloads a file using provider resolution
func (s *Service) Download(ctx context.Context, provider storage.Provider, file *storagetypes.File, opts *storage.DownloadOptions) (*storage.DownloadedMetadata, error) {
	if provider == nil {
		resolvedprovider, err := s.resolveDownloadProvider(ctx, file)
		if err != nil {
			return nil, err
		}
		provider = resolvedprovider
	}

	return s.objectService.Download(ctx, provider, file, opts)
}

// GetPresignedURL gets a presigned URL for a file using provider resolution
func (s *Service) GetPresignedURL(ctx context.Context, file *storagetypes.File, duration time.Duration) (string, error) {
	if s.tokenProvider == nil {
		provider, err := s.resolveDownloadProvider(ctx, file)
		if err != nil {
			return "", err
		}

		opts := &storagetypes.PresignedURLOptions{Duration: duration}
		return s.objectService.GetPresignedURL(ctx, provider, file, opts)
	}

	if file == nil || file.ID == "" {
		return "", ErrMissingFileID
	}

	// ensure provider resolution succeeds before issuing token
	if _, err := s.resolveDownloadProvider(ctx, file); err != nil {
		return "", err
	}

	if duration <= 0 {
		duration = defaultPresignedURLDuration
	}

	objectURI := buildDownloadObjectURI(file.FileMetadata.ProviderType, file.FileMetadata.Bucket, file.FileMetadata.Key)
	options := []tokens.DownloadTokenOption{
		tokens.WithDownloadTokenExpiresIn(duration),
		tokens.WithDownloadTokenContentType(file.FileMetadata.ContentType),
	}

	if file.OriginalName != "" {
		options = append(options, tokens.WithDownloadTokenFileName(file.OriginalName))
	}

	if authUser, ok := auth.AuthenticatedUserFromContext(ctx); ok && authUser != nil {
		if userID, err := ulid.Parse(authUser.SubjectID); err == nil {
			options = append(options, tokens.WithDownloadTokenUserID(userID))
		}
		if authUser.OrganizationID != "" {
			if orgID, err := ulid.Parse(authUser.OrganizationID); err == nil {
				options = append(options, tokens.WithDownloadTokenOrgID(orgID))
			}
		}
	}

	downloadToken, err := tokens.NewDownloadToken(objectURI, options...)
	if err != nil {
		return "", err
	}

	signature, secret, err := downloadToken.Sign()
	if err != nil {
		return "", err
	}

	s.storeDownloadSecret(downloadToken.TokenID, secret, downloadToken.ExpiresAt)

	payload, err := msgpack.Marshal(downloadToken)
	if err != nil {
		return "", err
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	combined := fmt.Sprintf("%s.%s", signature, encodedPayload)
	escaped := url.QueryEscape(combined)

	return fmt.Sprintf("/v1/files/%s/download?token=%s", url.PathEscape(file.ID), escaped), nil
}

// Delete deletes a file using provider resolution
func (s *Service) Delete(ctx context.Context, file *storagetypes.File, opts *storagetypes.DeleteFileOptions) error {
	provider, err := s.resolveDownloadProvider(ctx, file)
	if err != nil {
		return err
	}

	return s.objectService.Delete(ctx, provider, file, opts)
}

// Exists checks if a file exists using provider resolution
func (s *Service) Exists(ctx context.Context, file *storagetypes.File) (bool, error) {
	provider, err := s.resolveDownloadProvider(ctx, file)
	if err != nil {
		return false, err
	}

	return provider.Exists(ctx, file)
}

type downloadSecret struct {
	secret    []byte
	expiresAt time.Time
}

func (s *Service) storeDownloadSecret(tokenID ulid.ULID, secret []byte, expiresAt time.Time) {
	if ulids.IsZero(tokenID) || len(secret) == 0 {
		return
	}

	copySecret := make([]byte, len(secret))
	copy(copySecret, secret)

	key := tokenID.String()
	s.downloadSecrets.Store(key, downloadSecret{secret: copySecret, expiresAt: expiresAt})

	if ttl := time.Until(expiresAt); ttl > 0 {
		time.AfterFunc(ttl, func() {
			s.downloadSecrets.Delete(key)
		})
	}
}

func (s *Service) LookupDownloadSecret(tokenID ulid.ULID) ([]byte, bool) {
	if ulids.IsZero(tokenID) {
		return nil, false
	}

	value, ok := s.downloadSecrets.Load(tokenID.String())
	if !ok {
		return nil, false
	}

	ds := value.(downloadSecret)
	if time.Now().After(ds.expiresAt) {
		s.downloadSecrets.Delete(tokenID.String())
		return nil, false
	}

	return ds.secret, true
}

func buildDownloadObjectURI(provider storagetypes.ProviderType, bucket, key string) string {
	return fmt.Sprintf("%s:%s:%s", string(provider), bucket, key)
}

// Skipper returns the configured skipper function
func (s *Service) Skipper() storage.SkipperFunc {
	return s.objectService.Skipper()
}

// ErrorResponseHandler returns the configured error response handler
func (s *Service) ErrorResponseHandler() storage.ErrResponseHandler {
	return s.objectService.ErrorResponseHandler()
}

// MaxSize returns the configured maximum file size
func (s *Service) MaxSize() int64 {
	return s.objectService.MaxSize()
}

// Keys returns the configured form keys
func (s *Service) Keys() []string {
	return s.objectService.Keys()
}

// IgnoreNonExistentKeys returns whether to ignore non-existent form keys
func (s *Service) IgnoreNonExistentKeys() bool {
	return s.objectService.IgnoreNonExistentKeys()
}

// resolveProvider resolves a storage provider for upload operations
func (s *Service) resolveUploadProvider(ctx context.Context, opts *storage.UploadOptions) (storage.Provider, error) {
	enrichedCtx := s.buildResolutionContext(ctx, opts)
	resolution := s.resolver.Resolve(enrichedCtx)

	if !resolution.IsPresent() {
		logStorageResolutionFailure(ctx, opts)
		return nil, ErrProviderResolutionFailed
	}

	res := resolution.MustGet()
	if res.ClientType == "" {
		logStorageResolutionFailure(ctx, opts)
		return nil, ErrProviderResolutionFailed
	}

	// Get organization ID from auth context
	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil || orgID == "" {
		return nil, ErrNoOrganizationID
	}

	cacheKey := cp.ClientCacheKey{
		TenantID:        orgID,
		IntegrationType: string(res.ClientType),
		//		IntegrationID:   integrationID,
		//		HushID:          hushID,
	}

	client := s.clientService.GetClient(ctx, cacheKey, res.ClientType, res.Credentials, res.Config)
	if !client.IsPresent() {
		logStorageResolutionFailure(ctx, opts)
		return nil, ErrProviderResolutionFailed
	}

	return client.MustGet(), nil
}

// resolveProviderForFile resolves a storage provider for file operations (download, delete, presigned URL)
func (s *Service) resolveDownloadProvider(ctx context.Context, file *storagetypes.File) (storage.Provider, error) {
	enrichedCtx := s.buildResolutionContextForFile(ctx, file)
	resolution := s.resolver.Resolve(enrichedCtx)

	providerType, hasResolution := resolution.Get()
	if !hasResolution {
		zerolog.Ctx(ctx).Error().Msgf("storage provider resolution failed for file %s", file.ID)
		return nil, ErrProviderResolutionFailed
	}

	// Build ClientCacheKey using file metadata with auth context as backup
	orgID, _ := auth.GetOrganizationIDFromContext(ctx)

	cacheKey := cp.ClientCacheKey{
		TenantID:        orgID,
		IntegrationType: string(providerType.ClientType),
		//		IntegrationID:   file.IntegrationID,
		//		HushID:          file.HushID,
	}

	return s.clientService.GetClient(ctx, cacheKey, providerType.ClientType, providerType.Credentials, providerType.Config).
		OrElse(nil), nil
}

// buildResolutionContext builds context for provider resolution from upload options
func (s *Service) buildResolutionContext(ctx context.Context, opts *storage.UploadOptions) context.Context {
	// Add organization and user information from auth context
	orgID, _ := auth.GetOrganizationIDFromContext(ctx)
	ctx = cp.WithValue(ctx, orgID)

	subjectID, _ := auth.GetSubjectIDFromContext(ctx)
	ctx = cp.WithValue(ctx, subjectID)

	// Add upload options
	ctx = cp.WithValue(ctx, opts)

	// Add provider hints if present
	if opts.ProviderHints != nil {
		ctx = cp.WithValue(ctx, opts.ProviderHints)
		ctx = ApplyProviderHints(ctx, opts.ProviderHints)
	}

	return ctx
}

func logStorageResolutionFailure(ctx context.Context, opts *storage.UploadOptions) {
	event := zerolog.Ctx(ctx).Error()

	if orgID, err := auth.GetOrganizationIDFromContext(ctx); err == nil && orgID != "" {
		event = event.Str("org_id", orgID)
	}

	if opts != nil {
		event = event.Str("bucket_hint", opts.Bucket).Str("upload_key", opts.Key)
		if hints := opts.ProviderHints; hints != nil {
			event = event.
				Str("known_provider", string(hints.KnownProvider)).
				Str("preferred_provider", string(hints.PreferredProvider)).
				Str("hint_org_id", hints.OrganizationID).
				Str("hint_integration_id", hints.IntegrationID).
				Str("hint_hush_id", hints.HushID)

			if len(hints.Metadata) > 0 {
				event = event.Interface("hint_metadata", hints.Metadata)
			}
		}
	}

	event.Msg("storage provider resolution failed")
}

// buildResolutionContextForFile builds context for provider resolution from file metadata
func (s *Service) buildResolutionContextForFile(ctx context.Context, file *storagetypes.File) context.Context {
	// Add organization and user information from auth context
	orgID, _ := auth.GetOrganizationIDFromContext(ctx)
	ctx = cp.WithValue(ctx, orgID)

	subjectID, _ := auth.GetSubjectIDFromContext(ctx)
	ctx = cp.WithValue(ctx, subjectID)

	// Add the entire file
	ctx = cp.WithValue(ctx, file)
	ctx = ApplyProviderHints(ctx, file.ProviderHints)

	return ctx
}
