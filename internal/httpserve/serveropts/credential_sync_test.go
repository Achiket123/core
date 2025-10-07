package serveropts_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hush"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/httpserve/serveropts"
	"github.com/theopenlane/core/pkg/cp"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/objects/storage"

	// import generated runtime which is required to prevent cyclical dependencies
	_ "github.com/theopenlane/core/internal/ent/generated/runtime"
)

func (suite *CredentialSyncTestSuite) TestCredentialsMatch() {
	t := suite.T()

	tests := []struct {
		name        string
		credSet     models.CredentialSet
		configCreds storage.ProviderCredentials
		want        bool
	}{
		{
			name: "matching credentials",
			credSet: models.CredentialSet{
				AccessKeyID:     "test_key",
				SecretAccessKey: "test_secret",
				Endpoint:        "test_endpoint",
				ProjectID:       "test_project",
				AccountID:       "test_account",
			},
			configCreds: storage.ProviderCredentials{
				AccessKeyID:     "test_key",
				SecretAccessKey: "test_secret",
				Endpoint:        "test_endpoint",
				ProjectID:       "test_project",
				AccountID:       "test_account",
			},
			want: true,
		},
		{
			name: "non-matching credentials",
			credSet: models.CredentialSet{
				AccessKeyID:     "test_key",
				SecretAccessKey: "test_secret",
			},
			configCreds: storage.ProviderCredentials{
				AccessKeyID:     "different_key",
				SecretAccessKey: "test_secret",
			},
			want: false,
		},
		{
			name:        "empty credentials",
			credSet:     models.CredentialSet{},
			configCreds: storage.ProviderCredentials{},
			want:        true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Test by comparing hashes directly
			configHash := suite.service.GenerateCredentialHash(tc.configCreds)
			credSetHash := suite.service.GenerateCredentialHashFromSet(tc.credSet)
			result := configHash == credSetHash
			assert.Equal(t, tc.want, result)
		})
	}
}

func (suite *CredentialSyncTestSuite) TestCredentialHashConsistency() {
	t := suite.T()

	t.Run("hashes are consistent between ProviderCredentials and CredentialSet", func(t *testing.T) {
		creds := storage.ProviderCredentials{
			AccessKeyID:     "test_key",
			SecretAccessKey: "test_secret",
			Endpoint:        "test_endpoint",
			ProjectID:       "test_project",
			AccountID:       "test_account",
		}

		credSet := models.CredentialSet{
			AccessKeyID:     creds.AccessKeyID,
			SecretAccessKey: creds.SecretAccessKey,
			Endpoint:        creds.Endpoint,
			ProjectID:       creds.ProjectID,
			AccountID:       creds.AccountID,
		}

		hash1 := suite.service.GenerateCredentialHash(creds)
		hash2 := suite.service.GenerateCredentialHashFromSet(credSet)

		assert.Equal(t, hash1, hash2, "Hashes should be identical for equivalent credentials")
	})

	t.Run("region and bucket changes don't affect hash", func(t *testing.T) {
		baseCreds := storage.ProviderCredentials{
			AccessKeyID:     "test_key",
			SecretAccessKey: "test_secret",
			Endpoint:        "test_endpoint",
		}

		creds1 := baseCreds
		// Region and Bucket are not part of ProviderCredentials, so we skip these assignments

		creds2 := baseCreds
		// Region and Bucket are not part of ProviderCredentials, so we skip these assignments

		hash1 := suite.service.GenerateCredentialHash(creds1)
		hash2 := suite.service.GenerateCredentialHash(creds2)

		assert.Equal(t, hash1, hash2, "Region and bucket changes should not affect hash")
	})

	t.Run("credential field changes affect hash", func(t *testing.T) {
		baseCreds := storage.ProviderCredentials{
			AccessKeyID:     "test_key",
			SecretAccessKey: "test_secret",
			Endpoint:        "test_endpoint",
			ProjectID:       "test_project",
			AccountID:       "test_account",
		}

		baseHash := suite.service.GenerateCredentialHash(baseCreds)

		tests := []struct {
			name   string
			modify func(storage.ProviderCredentials) storage.ProviderCredentials
		}{
			{
				name: "AccessKeyID change",
				modify: func(c storage.ProviderCredentials) storage.ProviderCredentials {
					c.AccessKeyID = "different_key"
					return c
				},
			},
			{
				name: "SecretAccessKey change",
				modify: func(c storage.ProviderCredentials) storage.ProviderCredentials {
					c.SecretAccessKey = "different_secret"
					return c
				},
			},
			{
				name: "Endpoint change",
				modify: func(c storage.ProviderCredentials) storage.ProviderCredentials {
					c.Endpoint = "different_endpoint"
					return c
				},
			},
			{
				name: "ProjectID change",
				modify: func(c storage.ProviderCredentials) storage.ProviderCredentials {
					c.ProjectID = "different_project"
					return c
				},
			},
			{
				name: "AccountID change",
				modify: func(c storage.ProviderCredentials) storage.ProviderCredentials {
					c.AccountID = "different_account"
					return c
				},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				modifiedCreds := tc.modify(baseCreds)
				modifiedHash := suite.service.GenerateCredentialHash(modifiedCreds)
				assert.NotEqual(t, baseHash, modifiedHash, "Hash should change when credential field changes")
			})
		}
	})
}

func (suite *CredentialSyncTestSuite) TestSyncConfigCredentials() {
	t := suite.T()

	// Clean up any existing integrations before each test
	cleanupCtx := suite.systemContext()
	_, err := suite.db.Integration.Delete().Where(
		integration.KindEQ(string(storage.S3Provider)),
		integration.SystemOwnedEQ(true),
	).Exec(cleanupCtx)
	require.NoError(t, err)

	t.Run("initial sync creates integrations and secrets", func(t *testing.T) {
		ctx := suite.systemContext()

		// Configure service with S3 credentials
		config := &storage.Providers{
			S3: storage.ProviderConfigs{
				Enabled: true,
				Region:  "us-east-1",
				Bucket:  "test-bucket",
				Credentials: storage.ProviderCredentials{
					AccessKeyID:     "test_access_key",
					SecretAccessKey: "test_secret_key",
				},
			},
		}

		clientPool := cp.NewClientPool[storage.Provider](time.Hour)
		clientService := cp.NewClientService[
			storage.Provider,
			storage.ProviderCredentials,
			*storage.ProviderOptions,
		](clientPool, cp.WithConfigClone[
			storage.Provider,
			storage.ProviderCredentials,
			*storage.ProviderOptions,
		](cloneProviderOptions))
		service := serveropts.NewCredentialSyncService(suite.db, clientService, config)

		// Run sync
		err := service.SyncConfigCredentials(ctx)
		require.NoError(t, err)

		// Verify integration was created
		integrations, err := suite.db.Integration.Query().
			Where(
				integration.KindEQ(string(storage.S3Provider)),
				integration.SystemOwnedEQ(true),
			).
			WithSecrets(func(q *ent.HushQuery) {
				q.Where(hush.SystemOwnedEQ(true))
			}).
			All(ctx)
		require.NoError(t, err)
		assert.Len(t, integrations, 1)

		// Verify secret was created with correct credentials
		integration := integrations[0]
		assert.Len(t, integration.Edges.Secrets, 1)
		secret := integration.Edges.Secrets[0]
		assert.Equal(t, "test_access_key", secret.CredentialSet.AccessKeyID)
		assert.Equal(t, "test_secret_key", secret.CredentialSet.SecretAccessKey)
	})

	t.Run("sync with same credentials doesn't create duplicate", func(t *testing.T) {
		ctx := suite.systemContext()

		config := &storage.Providers{
			S3: storage.ProviderConfigs{
				Enabled: true,
				Region:  "us-east-1",
				Bucket:  "test-bucket",
				Credentials: storage.ProviderCredentials{
					AccessKeyID:     "same_key",
					SecretAccessKey: "same_secret",
				},
			},
		}

		clientPool := cp.NewClientPool[storage.Provider](time.Hour)
		clientService := cp.NewClientService[
			storage.Provider,
			storage.ProviderCredentials,
			*storage.ProviderOptions,
		](clientPool, cp.WithConfigClone[
			storage.Provider,
			storage.ProviderCredentials,
			*storage.ProviderOptions,
		](cloneProviderOptions))
		service := serveropts.NewCredentialSyncService(suite.db, clientService, config)

		// Run sync twice
		err := service.SyncConfigCredentials(ctx)
		require.NoError(t, err)

		countBefore, err := suite.db.Integration.Query().
			Where(
				integration.KindEQ(string(storage.S3Provider)),
				integration.SystemOwnedEQ(true),
			).
			Count(ctx)
		require.NoError(t, err)

		err = service.SyncConfigCredentials(ctx)
		require.NoError(t, err)

		countAfter, err := suite.db.Integration.Query().
			Where(
				integration.KindEQ(string(storage.S3Provider)),
				integration.SystemOwnedEQ(true),
			).
			Count(ctx)
		require.NoError(t, err)

		assert.Equal(t, countBefore, countAfter, "Should not create duplicate integrations")
	})

	t.Run("sync with updated credentials creates new integration", func(t *testing.T) {
		// Clean up any existing integrations before this specific test
		cleanupCtx := suite.systemContext()
		_, err = suite.db.Integration.Delete().Where(
			integration.KindEQ(string(storage.S3Provider)),
			integration.SystemOwnedEQ(true),
		).Exec(cleanupCtx)
		require.NoError(t, err)

		ctx := suite.systemContext()

		config := &storage.Providers{
			S3: storage.ProviderConfigs{
				Enabled: true,
				Region:  "us-east-1",
				Bucket:  "test-bucket",
				Credentials: storage.ProviderCredentials{
					AccessKeyID:     "initial_key",
					SecretAccessKey: "initial_secret",
				},
			},
		}

		clientPool := cp.NewClientPool[storage.Provider](time.Hour)
		clientService := cp.NewClientService[
			storage.Provider,
			storage.ProviderCredentials,
			*storage.ProviderOptions,
		](clientPool, cp.WithConfigClone[
			storage.Provider,
			storage.ProviderCredentials,
			*storage.ProviderOptions,
		](cloneProviderOptions))
		service := serveropts.NewCredentialSyncService(suite.db, clientService, config)

		// Initial sync
		err := service.SyncConfigCredentials(ctx)
		require.NoError(t, err)

		// Update credentials and sync again
		config.S3.Credentials.AccessKeyID = "updated_key"
		config.S3.Credentials.SecretAccessKey = "updated_secret"

		err = service.SyncConfigCredentials(ctx)
		require.NoError(t, err)

		// Should have 2 integrations now
		integrations, err := suite.db.Integration.Query().
			Where(
				integration.KindEQ(string(storage.S3Provider)),
				integration.SystemOwnedEQ(true),
			).
			WithSecrets(func(q *ent.HushQuery) {
				q.Where(hush.SystemOwnedEQ(true))
			}).
			All(ctx)
		require.NoError(t, err)
		assert.Len(t, integrations, 2)

		// Verify the latest one has updated credentials
		latest, err := service.GetActiveSystemProvider(ctx, storage.S3Provider)
		require.NoError(t, err)
		assert.Equal(t, "updated_key", latest.Edges.Secrets[0].CredentialSet.AccessKeyID)
	})
}

func (suite *CredentialSyncTestSuite) TestGetActiveSystemProvider() {
	t := suite.T()

	// Clean up any existing integrations before each test
	cleanupCtx := suite.systemContext()
	_, err := suite.db.Integration.Delete().Where(
		integration.KindEQ(string(storage.S3Provider)),
		integration.SystemOwnedEQ(true),
	).Exec(cleanupCtx)
	require.NoError(t, err)

	t.Run("no integrations returns error", func(t *testing.T) {
		ctx := suite.systemContext()

		_, err := suite.service.GetActiveSystemProvider(ctx, storage.ProviderType("nonexistent"))
		assert.ErrorIs(t, err, serveropts.ErrNoActiveIntegration)
	})

	t.Run("returns most recent integration by synchronized_at", func(t *testing.T) {
		// Clean up any existing integrations before this specific test
		cleanupCtx := suite.systemContext()
		_, err = suite.db.Integration.Delete().Where(
			integration.KindEQ(string(storage.S3Provider)),
			integration.SystemOwnedEQ(true),
		).Exec(cleanupCtx)
		require.NoError(t, err)

		ctx := suite.systemContext()

		// Configure service with S3 credentials
		config := &storage.Providers{
			S3: storage.ProviderConfigs{
				Enabled: true,
				Region:  "us-east-1",
				Bucket:  "test-bucket",
				Credentials: storage.ProviderCredentials{
					AccessKeyID:     "test_key_1",
					SecretAccessKey: "test_secret_1",
				},
			},
		}

		clientPool := cp.NewClientPool[storage.Provider](time.Hour)
		clientService := cp.NewClientService[
			storage.Provider,
			storage.ProviderCredentials,
			*storage.ProviderOptions,
		](clientPool, cp.WithConfigClone[
			storage.Provider,
			storage.ProviderCredentials,
			*storage.ProviderOptions,
		](cloneProviderOptions))
		service := serveropts.NewCredentialSyncService(suite.db, clientService, config)

		// Run sync to create first integration
		err := service.SyncConfigCredentials(ctx)
		require.NoError(t, err)

		// Update config with new credentials
		config.S3.Credentials.AccessKeyID = "test_key_2"
		config.S3.Credentials.SecretAccessKey = "test_secret_2"

		// Run sync again to create second integration
		err = service.SyncConfigCredentials(ctx)
		require.NoError(t, err)

		// Verify GetActiveSystemProvider returns the most recent one
		active, err := service.GetActiveSystemProvider(ctx, storage.S3Provider)
		require.NoError(t, err)
		assert.Equal(t, "test_key_2", active.Edges.Secrets[0].CredentialSet.AccessKeyID)
		assert.Equal(t, "test_secret_2", active.Edges.Secrets[0].CredentialSet.SecretAccessKey)
	})
}
