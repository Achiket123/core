package credsync

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hush"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
)

// ProviderRecord represents resolved provider credentials and options from the database.
type ProviderRecord struct {
	Credentials storage.ProviderCredentials
	Options     *storage.ProviderOptions
}

// QuerySystemProvider retrieves the most recent system-owned provider configuration for the given provider type.
func QuerySystemProvider(ctx context.Context, providerType storage.ProviderType) (*ProviderRecord, error) {
	entClient := ent.FromContext(ctx)
	if entClient == nil {
		return nil, objects.ErrNoSystemIntegration
	}

	ctx = systemOwnedQueryContext(ctx, entClient)

	integrations, err := entClient.Integration.Query().
		Where(
			integration.KindEQ(string(providerType)),
			integration.SystemOwnedEQ(true),
		).
		WithSecrets(func(q *ent.HushQuery) {
			q.Where(hush.SystemOwnedEQ(true))
		}).
		All(ctx)
	if err != nil || len(integrations) == 0 {
		return nil, fmt.Errorf("%w for provider %s", objects.ErrNoSystemIntegration, providerType)
	}

	var activeInteg *ent.Integration
	for _, integ := range integrations {
		if len(integ.Edges.Secrets) == 0 {
			continue
		}

		if activeInteg == nil {
			activeInteg = integ
			continue
		}

		current, ok := integ.Metadata["synchronized_at"].(string)
		if !ok {
			continue
		}

		best, ok := activeInteg.Metadata["synchronized_at"].(string)
		if !ok {
			activeInteg = integ
			continue
		}

		currentTime, errCurrent := time.Parse(time.RFC3339, current)
		bestTime, errBest := time.Parse(time.RFC3339, best)

		if errCurrent == nil && (errBest != nil || currentTime.After(bestTime)) {
			activeInteg = integ
		}
	}

	if activeInteg == nil {
		return nil, fmt.Errorf("%w for provider %s", objects.ErrNoIntegrationWithSecrets, providerType)
	}

	secret := activeInteg.Edges.Secrets[0]
	credentials := storage.ProviderCredentials{
		AccessKeyID:     secret.CredentialSet.AccessKeyID,
		SecretAccessKey: secret.CredentialSet.SecretAccessKey,
		Endpoint:        secret.CredentialSet.Endpoint,
		ProjectID:       secret.CredentialSet.ProjectID,
		AccountID:       secret.CredentialSet.AccountID,
		APIToken:        secret.CredentialSet.APIToken,
	}

	options := storage.NewProviderOptions(storage.WithCredentials(credentials))

	if activeInteg.Metadata != nil {
		for key, value := range activeInteg.Metadata {
			switch strings.ToLower(key) {
			case "bucket":
				if strVal, ok := stringValue(value); ok {
					options.Apply(storage.WithBucket(strVal))
				}
			case "region":
				if strVal, ok := stringValue(value); ok {
					options.Apply(storage.WithRegion(strVal))
				}
			case "endpoint":
				if strVal, ok := stringValue(value); ok {
					options.Apply(storage.WithEndpoint(strVal))
				}
			case "base_path":
				if strVal, ok := stringValue(value); ok {
					options.Apply(storage.WithBasePath(strVal))
				}
			case "local_url":
				if strVal, ok := stringValue(value); ok {
					options.Apply(storage.WithLocalURL(strVal))
				}
			default:
				options.Apply(storage.WithExtra(key, value))
			}
		}
	}

	return &ProviderRecord{
		Credentials: credentials,
		Options:     options,
	}, nil
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
