package serveropts

import (
	"context"
	"crypto/md5" //nolint:gosec // MD5 is used only for checksum comparison, not security
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hush"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/cp"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/objects/storage"
)

// ErrNoActiveIntegration is returned when no active system integration is found for a provider
var ErrNoActiveIntegration = errors.New("no active system integration found")

// CredentialSyncService manages synchronization between config file credentials and database records
type CredentialSyncService struct {
	entClient     *generated.Client
	clientService *cp.ClientService[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]
	config        *storage.Providers
}

func (css *CredentialSyncService) systemOperationContext(ctx context.Context) context.Context {
	user := &auth.AuthenticatedUser{
		SubjectID:          "system-credential-sync",
		SubjectName:        "System Credential Sync",
		AuthenticationType: auth.APITokenAuthentication,
		IsSystemAdmin:      true,
		OrganizationID:     "",
		OrganizationIDs:    nil,
		ActiveSubscription: true,
	}

	ctx = auth.WithAuthenticatedUser(ctx, user)
	ctx = auth.WithSystemAdminContext(ctx, user)
	ctx = contextx.With(ctx, auth.OrganizationCreationContextKey{})
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	return generated.NewContext(ctx, css.entClient)
}

// NewCredentialSyncService creates a new credential synchronization service
func NewCredentialSyncService(entClient *generated.Client, clientService *cp.ClientService[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions], config *storage.Providers) *CredentialSyncService {
	return &CredentialSyncService{
		entClient:     entClient,
		clientService: clientService,
		config:        config,
	}
}

// SyncConfigCredentials synchronizes config file credentials with database records on startup
func (css *CredentialSyncService) SyncConfigCredentials(ctx context.Context) error {
	ctx = css.systemOperationContext(ctx)

	providerMap := map[storage.ProviderType]storage.ProviderConfigs{
		storage.S3Provider:   css.config.S3,
		storage.R2Provider:   css.config.CloudflareR2,
		storage.GCSProvider:  css.config.GCS,
		storage.DiskProvider: css.config.Disk,
	}

	for providerType, providerCfg := range providerMap {
		if !providerCfg.Enabled {
			continue
		}
		if err := css.syncProvider(ctx, providerType, providerCfg); err != nil {
			return err
		}
	}
	return nil
}

// syncProvider synchronizes a single provider's credentials
func (css *CredentialSyncService) syncProvider(ctx context.Context, providerType storage.ProviderType, providerCfg storage.ProviderConfigs) error {
	ctx = css.systemOperationContext(ctx)

	// Get current active system integration for this provider using ent client
	integrations, err := css.entClient.Integration.Query().
		Where(
			integration.KindEQ(string(providerType)),
			integration.SystemOwnedEQ(true),
		).
		WithSecrets(func(q *generated.HushQuery) {
			q.Where(hush.SystemOwnedEQ(true))
		}).
		All(ctx)
	if err != nil {
		return err
	}

	// Find active integration with non-expired credentials
	var activeInteg *generated.Integration
	for _, integ := range integrations {
		if len(integ.Edges.Secrets) == 0 {
			continue
		}

		secret := integ.Edges.Secrets[0]
		// Check if credentials match config
		if css.CredentialsMatch(secret, providerCfg.Credentials) {
			zerolog.Ctx(ctx).Debug().Msgf("credentials already up to date for provider %s integration %s", providerType, integ.ID)
			return nil
		}
		// If not matched, keep track of the first integration (for rotation)
		if activeInteg == nil {
			activeInteg = integ
		}
	}

	// Create new integration + hush for updated config credentials
	newInteg, err := css.createSystemIntegration(ctx, providerType, providerCfg)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msgf("failed to create system integration for provider %s", providerType)
		return err
	}

	// If we had an existing integration, mark it as superseded (but don't expire yet)
	if activeInteg != nil {
		zerolog.Ctx(ctx).Info().Msgf("rotated system credentials for provider %s from integration %s to %s", providerType, activeInteg.ID, newInteg.ID)
	}

	return nil
}

// CredentialsMatch checks if the stored credentials match the config credentials
func (css *CredentialSyncService) CredentialsMatch(secret *generated.Hush, configCreds storage.ProviderCredentials) bool {
	if secret == nil {
		return false
	}

	// Generate hash of config credentials for comparison
	configHash := css.GenerateCredentialHash(configCreds)

	// Check if stored secret has the same hash
	storedHash := css.GenerateCredentialHashFromSet(secret.CredentialSet)
	return configHash == storedHash
}

// GenerateCredentialHash creates a hash of credential data for comparison
func (css *CredentialSyncService) GenerateCredentialHash(creds storage.ProviderCredentials) string {
	data, _ := json.Marshal(creds)
	hash := md5.Sum(data) //nolint:gosec // MD5 is used only for checksum comparison, not security
	return hex.EncodeToString(hash[:])
}

// GenerateCredentialHashFromSet creates a hash from a CredentialSet
func (css *CredentialSyncService) GenerateCredentialHashFromSet(credSet models.CredentialSet) string {
	data, _ := json.Marshal(credSet)
	hash := md5.Sum(data) //nolint:gosec // MD5 is used only for checksum comparison, not security
	return hex.EncodeToString(hash[:])
}

// createSystemIntegration creates a new system integration and hush record
func (css *CredentialSyncService) createSystemIntegration(ctx context.Context, providerType storage.ProviderType, providerCfg storage.ProviderConfigs) (*generated.Integration, error) {
	systemCtx := css.systemOperationContext(ctx)

	credSet := models.CredentialSet{
		AccessKeyID:     providerCfg.Credentials.AccessKeyID,
		SecretAccessKey: providerCfg.Credentials.SecretAccessKey,
		Endpoint:        providerCfg.Credentials.Endpoint,
		ProjectID:       providerCfg.Credentials.ProjectID,
		AccountID:       providerCfg.Credentials.AccountID,
	}

	metadata := map[string]any{
		"region":          providerCfg.Region,
		"bucket":          providerCfg.Bucket,
		"source":          "system_config",
		"synchronized_at": time.Now(),
	}

	if providerCfg.Endpoint != "" {
		metadata["endpoint"] = providerCfg.Endpoint
	}

	if providerType == storage.DiskProvider {
		metadata["base_path"] = providerCfg.Bucket
		if providerCfg.Endpoint != "" {
			metadata["local_url"] = providerCfg.Endpoint
		}
	}

	hushRecord, err := css.entClient.Hush.Create().
		SetName(fmt.Sprintf("%s_system_credentials", providerType)).
		SetDescription(fmt.Sprintf("System configuration credentials for %s storage", providerType)).
		SetKind(string(providerType)).
		SetMetadata(metadata).
		SetCredentialSet(credSet).
		SetSystemOwned(true).
		SetSystemInternalID(fmt.Sprintf("storage:%s:config", providerType)).
		SetInternalNotes("Managed via server configuration").
		Save(systemCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to create hush record: %w", err)
	}

	integrationRecord, err := css.entClient.Integration.Create().
		SetName(fmt.Sprintf("System %s Storage", providerType)).
		SetDescription(fmt.Sprintf("System-level %s storage integration", providerType)).
		SetKind(string(providerType)).
		SetIntegrationType("storage").
		SetMetadata(metadata).
		SetSystemOwned(true).
		SetSystemInternalID(fmt.Sprintf("storage:%s:config", providerType)).
		SetInternalNotes("Managed via server configuration").
		AddSecrets(hushRecord).
		Save(systemCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to create integration record: %w", err)
	}

	zerolog.Ctx(ctx).Info().Msgf("created system integration %s for config credentials for provider %s", integrationRecord.ID, providerType)

	return integrationRecord, nil
}

// GetActiveSystemProvider returns the active system provider for a given type
func (css *CredentialSyncService) GetActiveSystemProvider(ctx context.Context, providerType storage.ProviderType) (*generated.Integration, error) {
	ctx = css.systemOperationContext(ctx)

	integrations, err := css.entClient.Integration.Query().
		Where(
			integration.KindEQ(string(providerType)),
			integration.SystemOwnedEQ(true),
		).
		WithSecrets(func(q *generated.HushQuery) {
			q.Where(hush.SystemOwnedEQ(true))
		}).
		All(ctx)
	if err != nil {
		log.Error().Err(err).Msgf("failed to query integrations for provider %s", providerType)

		return nil, err
	}

	var newest *generated.Integration
	var newestTime time.Time

	for _, integ := range integrations {
		if syncTimeStr, ok := integ.Metadata["synchronized_at"].(string); ok {
			if syncTime, err := time.Parse(time.RFC3339, syncTimeStr); err == nil {
				if newest == nil || syncTime.After(newestTime) {
					newest = integ
					newestTime = syncTime
				}
			}
		}
	}

	if newest == nil {
		return nil, fmt.Errorf("%w for provider: %s", ErrNoActiveIntegration, providerType)
	}

	return newest, nil
}
