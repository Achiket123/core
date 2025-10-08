package serveropts

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/rs/zerolog/log"
	echoprometheus "github.com/theopenlane/echo-prometheus"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/iam/providers/webauthn"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/iam/totp"
	"github.com/theopenlane/riverboat/pkg/riverqueue"
	"github.com/theopenlane/utils/cache"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/echox/middleware/echocontext"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hush/crypto"
	"github.com/theopenlane/core/internal/graphapi"
	"github.com/theopenlane/core/internal/httpserve/config"
	"github.com/theopenlane/core/internal/httpserve/server"
	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/internal/objects/resolver"
	"github.com/theopenlane/core/pkg/entitlements"
	authmw "github.com/theopenlane/core/pkg/middleware/auth"
	"github.com/theopenlane/core/pkg/middleware/cachecontrol"
	"github.com/theopenlane/core/pkg/middleware/cors"
	"github.com/theopenlane/core/pkg/middleware/csrf"
	"github.com/theopenlane/core/pkg/middleware/impersonation"
	"github.com/theopenlane/core/pkg/middleware/ratelimit"
	"github.com/theopenlane/core/pkg/middleware/redirect"
	"github.com/theopenlane/core/pkg/middleware/secure"
	"github.com/theopenlane/core/pkg/objects/storage"
	disk "github.com/theopenlane/core/pkg/objects/storage/providers/disk"
	r2provider "github.com/theopenlane/core/pkg/objects/storage/providers/r2"
	s3provider "github.com/theopenlane/core/pkg/objects/storage/providers/s3"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
	"github.com/theopenlane/core/pkg/summarizer"
	"github.com/theopenlane/core/pkg/windmill"
)

type ServerOption interface {
	apply(*ServerOptions)
}

type applyFunc struct {
	applyInternal func(*ServerOptions)
}

func (fso *applyFunc) apply(s *ServerOptions) {
	fso.applyInternal(s)
}

func newApplyFunc(apply func(option *ServerOptions)) *applyFunc {
	return &applyFunc{
		applyInternal: apply,
	}
}

// WithConfigProvider supplies the config for the server
func WithConfigProvider(cfgProvider config.Provider) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		s.ConfigProvider = cfgProvider
	})
}

// WithHTTPS sets up TLS config settings for the server
func WithHTTPS() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if !s.Config.Settings.Server.TLS.Enabled {
			// this is set to enabled by WithServer
			// if TLS is not enabled, move on
			return
		}

		s.Config.WithTLSDefaults()

		if !s.Config.Settings.Server.TLS.AutoCert {
			s.Config.WithTLSCerts(s.Config.Settings.Server.TLS.CertFile, s.Config.Settings.Server.TLS.CertKey)
		}
	})
}

// WithGeneratedKeys will generate a public/private key pair
// that can be used for jwt signing.
// This should only be used in a development environment
func WithGeneratedKeys() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		privFileName := "private_key.pem"

		// generate a new private key if one doesn't exist
		if _, err := os.Stat(privFileName); err != nil {
			// Generate a new RSA private key with 2048 bits
			privateKey, err := rsa.GenerateKey(rand.Reader, 2048) //nolint:mnd
			if err != nil {
				log.Panic().Err(err).Msg("Error generating RSA private key")
			}

			// Encode the private key to the PEM format
			privateKeyPEM := &pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
			}

			privateKeyFile, err := os.Create(privFileName)
			if err != nil {
				log.Panic().Err(err).Msg("Error creating private key file")
			}

			if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
				log.Panic().Err(err).Msg("unable to encode pem on startup")
			}

			privateKeyFile.Close()
		}

		keys := map[string]string{}

		// check if kid was passed in
		kidPriv := s.Config.Settings.Auth.Token.KID

		// if we didn't get a kid in the settings, assign one
		if kidPriv == "" {
			kidPriv = ulids.New().String()
		}

		keys[kidPriv] = fmt.Sprintf("%v", privFileName)

		s.Config.Settings.Auth.Token.Keys = keys
	})
}

// WithTokenManager sets up the token manager for the server
func WithTokenManager() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		// Setup token manager
		tm, err := tokens.New(s.Config.Settings.Auth.Token)
		if err != nil {
			log.Panic().Err(err).Msg("Error creating token manager")
		}

		keys, err := tm.Keys()
		if err != nil {
			log.Panic().Err(err).Msg("Error getting keys from token manager")
		}

		// pass to the REST handlers
		s.Config.Handler.JWTKeys = keys
		s.Config.Handler.TokenManager = tm
	})
}

// WithAuth supplies the authn and jwt config for the server
func WithAuth() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		// add oauth providers for social login
		s.Config.Handler.OauthProvider = s.Config.Settings.Auth.Providers

		// add oauth providers for integrations (separate config)
		s.Config.Handler.IntegrationOauthProvider = s.Config.Settings.IntegrationOauthProvider

		// add auth middleware
		opts := []authmw.Option{
			authmw.WithAudience(s.Config.Settings.Auth.Token.Audience),
			authmw.WithIssuer(s.Config.Settings.Auth.Token.Issuer),
			authmw.WithJWKSEndpoint(s.Config.Settings.Auth.Token.JWKSEndpoint),
			authmw.WithDBClient(s.Config.Handler.DBClient),
			authmw.WithCookieConfig(s.Config.SessionConfig.CookieConfig),
			authmw.WithAllowAnonymous(true),
		}

		if s.Config.Handler.RedisClient != nil {
			opts = append(opts, authmw.WithRedisClient(s.Config.Handler.RedisClient))
		}

		conf := authmw.NewAuthOptions(opts...)

		s.Config.Handler.WebAuthn = webauthn.NewWithConfig(s.Config.Settings.Auth.Providers.Webauthn)

		s.Config.GraphMiddleware = append(s.Config.GraphMiddleware, authmw.Authenticate(&conf), impersonation.SystemAdminUserContextMiddleware())
		s.Config.Handler.AuthMiddleware = append(s.Config.Handler.AuthMiddleware, authmw.Authenticate(&conf))
	})
}

// WithReadyChecks adds readiness checks to the server
func WithReadyChecks(c *entx.EntClientConfig, f *fgax.Client, r *redis.Client, j riverqueue.JobClient) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		// Always add a check to the primary db connection
		s.Config.Handler.AddReadinessCheck("db_primary", entx.Healthcheck(c.GetPrimaryDB()))

		// Check the connection to the job queue
		jc := j.(*riverqueue.Client)
		s.Config.Handler.AddReadinessCheck(("job_queue"), riverqueue.Healthcheck(jc))

		// Check the secondary db, if enabled
		if s.Config.Settings.DB.MultiWrite {
			s.Config.Handler.AddReadinessCheck("db_secondary", entx.Healthcheck(c.GetSecondaryDB()))
		}

		// Check the connection to openFGA, if enabled
		if s.Config.Settings.Authz.Enabled {
			s.Config.Handler.AddReadinessCheck("fga", fgax.Healthcheck(*f))
		}

		// Check the connection to redis, if enabled
		if s.Config.Settings.Redis.Enabled {
			s.Config.Handler.AddReadinessCheck("redis", cache.Healthcheck(r))
		}
	})
}

// WithGraphRoute adds the graph handler to the server
func WithGraphRoute(srv *server.Server, c *ent.Client) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		// Setup Graph API Handlers
		r := graphapi.NewResolver(c, s.Config.StorageService).
			WithExtensions(s.Config.Settings.Server.EnableGraphExtensions).
			WithDevelopment(s.Config.Settings.Server.Dev).
			WithComplexityLimitConfig(s.Config.Settings.Server.ComplexityLimit).
			WithMaxResultLimit(s.Config.Settings.Server.MaxResultLimit).
			WithTrustCenterCnameTarget(s.Config.Settings.Server.TrustCenterCnameTarget).
			WithTrustCenterDefaultDomain(s.Config.Settings.Server.DefaultTrustCenterDomain)

		// add pool to the resolver to manage the number of goroutines
		r.WithPool(
			s.Config.Settings.Server.GraphPool.MaxWorkers,
			true, // include metrics collectors
		)

		handler := r.Handler(s.Config.Settings.Server.Dev)

		// Add Graph Handler
		srv.AddHandler(handler)
	})
}

// WithMiddleware adds the middleware to the server
func WithMiddleware() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		// Initialize middleware if null
		if s.Config.DefaultMiddleware == nil {
			s.Config.DefaultMiddleware = []echo.MiddlewareFunc{}
		}

		// default middleware
		s.Config.DefaultMiddleware = append(s.Config.DefaultMiddleware,
			echoprometheus.MetricsMiddleware(),           // add prometheus metrics
			echocontext.EchoContextToContextMiddleware(), // adds echo context to parent
		)
	})
}

// WithEmailConfig sets up the email config to be used to send emails to users
// on registration, password reset, etc
func WithEmailConfig() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		s.Config.Handler.Emailer = s.Config.Settings.Email
	})
}

// WithDefaultTrustCenterDomain sets up the default trust center domain for the server
func WithDefaultTrustCenterDomain() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		s.Config.Handler.DefaultTrustCenterDomain = s.Config.Settings.Server.DefaultTrustCenterDomain
	})
}

// WithSessionManager sets up the default session manager with a 10 minute ttl
// with persistence to redis
func WithSessionManager(rc *redis.Client) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		cc := sessions.DefaultCookieConfig

		// In order for things to work in dev mode with localhost
		// we need to se the debug cookie config
		if s.Config.Settings.Server.Dev {
			cc = sessions.DebugOnlyCookieConfig
		} else {
			cc.Name = sessions.DefaultCookieName
		}

		if s.Config.Settings.Sessions.MaxAge > 0 {
			cc.MaxAge = s.Config.Settings.Sessions.MaxAge
		}

		if s.Config.Settings.Sessions.Domain != "" {
			cc.Domain = s.Config.Settings.Sessions.Domain
		}

		cc.HTTPOnly = s.Config.Settings.Sessions.HTTPOnly
		cc.Secure = s.Config.Settings.Sessions.Secure

		if s.Config.Settings.Sessions.SameSite != "" {
			switch strings.ToLower(s.Config.Settings.Sessions.SameSite) {
			case "lax":
				cc.SameSite = http.SameSiteLaxMode
			case "strict":
				cc.SameSite = http.SameSiteStrictMode
			case "none":
				cc.SameSite = http.SameSiteNoneMode
			default:
				cc.SameSite = http.SameSiteDefaultMode
			}
		}

		sm := sessions.NewCookieStore[map[string]any](cc,
			[]byte(s.Config.Settings.Sessions.SigningKey),
			[]byte(s.Config.Settings.Sessions.EncryptionKey),
		)

		// add session middleware, this has to be added after the authMiddleware so we have the user id
		// when we get to the session. this is also added here so its only added to the graph routes
		// REST routes are expected to add the session middleware, as required
		sessionConfig := sessions.NewSessionConfig(
			sm,
			sessions.WithPersistence(rc),
			sessions.WithSkipperFunc(authmw.SessionSkipperFunc),
		)

		// set cookie config to be used
		sessionConfig.CookieConfig = cc

		// Add redis client to Handlers Config
		s.Config.Handler.RedisClient = rc

		// Make the cookie session store available
		// to graph and REST endpoints
		s.Config.Handler.SessionConfig = &sessionConfig
		s.Config.SessionConfig = &sessionConfig
	})
}

// WithSessionMiddleware sets up the session middleware for the server
func WithSessionMiddleware() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		// add session middleware, this has to be added after the authMiddleware
		s.Config.GraphMiddleware = append(s.Config.GraphMiddleware,
			sessions.LoadAndSaveWithConfig(*s.Config.SessionConfig),
		)
	})
}

// WithOTP sets up the OTP provider
func WithOTP() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if s.Config.Settings.TOTP.Enabled {
			if s.Config.Settings.TOTP.Secret == "" {
				s.Config.Settings.TOTP.Secret = ulids.New().String()
			}

			opts := []totp.ConfigOption{
				totp.WithCodeLength(s.Config.Settings.TOTP.CodeLength),
				totp.WithIssuer(s.Config.Settings.TOTP.Issuer),
				totp.WithSecret(totp.Secret{
					Version: 0,
					Key:     s.Config.Settings.TOTP.Secret,
				}),
				totp.WithRecoveryCodeLength(s.Config.Settings.TOTP.RecoveryCodeLength),
				totp.WithRecoveryCodeCount(s.Config.Settings.TOTP.RecoveryCodeCount),
			}

			// append redis client if enabled
			if s.Config.Settings.TOTP.WithRedis {
				opts = append(opts, totp.WithRedis(s.Config.Handler.RedisClient))
			}

			// setup new opt manager
			otpMan := totp.NewOTP(
				opts...,
			)

			s.Config.Handler.OTPManager = &totp.Client{
				Manager: otpMan,
			}
		}
	})
}

// WithRateLimiter sets up the rate limiter for the server
func WithRateLimiter() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if s.Config.Settings.Ratelimit.Enabled {
			s.Config.DefaultMiddleware = append(s.Config.DefaultMiddleware, ratelimit.RateLimiterWithConfig(&s.Config.Settings.Ratelimit))
		}
	})
}

// WithSecureMW sets up the secure middleware for the server
func WithSecureMW() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if s.Config.Settings.Server.Secure.Enabled {
			s.Config.DefaultMiddleware = append(s.Config.DefaultMiddleware, secure.Secure(&s.Config.Settings.Server.Secure))
		}
	})
}

// WithRedirects sets up the redirects for the server
func WithRedirects() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if s.Config.Settings.Server.Redirects.Enabled {
			redirects := s.Config.Settings.Server.Redirects
			s.Config.DefaultMiddleware = append(s.Config.DefaultMiddleware, redirect.NewWithConfig(redirects))
		}
	})
}

// WithCacheHeaders sets up the cache control headers for the server
func WithCacheHeaders() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if s.Config.Settings.Server.CacheControl.Enabled {
			cacheConfig := s.Config.Settings.Server.CacheControl
			s.Config.DefaultMiddleware = append(s.Config.DefaultMiddleware, cachecontrol.NewWithConfig(cacheConfig))
		}
	})
}

// WithCORS sets up the CORS middleware for the server
func WithCORS() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if s.Config.Settings.Server.CORS.Enabled {
			s.Config.DefaultMiddleware = append(s.Config.DefaultMiddleware, cors.MustNew(s.Config.Settings.Server.CORS.AllowOrigins))
		}
	})
}

// WithObjectStorage sets up the object storage for the server
func WithObjectStorage() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		// Create StorageService with resolver and cp using runtime config
		storageService := resolver.NewServiceFromConfig(s.Config.Settings.ObjectStorage)

		// Store in config for access
		s.Config.StorageService = storageService
		s.Config.Handler.ObjectStore = storageService

		validateConfiguredStorageProviders(s.Config.Settings.ObjectStorage)

		log.Info().Msg("Object storage initialized")
	})
}

// WithStorageCredentialSync synchronizes configured storage credentials into the database once dependencies are ready
func WithStorageCredentialSync() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if s.Config.Settings.ObjectStorage.DevMode {
			return
		}

		if s.Config.Handler.DBClient == nil {
			log.Warn().Msg("skipping storage credential sync: database client not ready")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), storageCredentialSyncTimeout)
		defer cancel()

		credentialSync := NewCredentialSyncService(s.Config.Handler.DBClient, nil, &s.Config.Settings.ObjectStorage.Providers)
		if err := credentialSync.SyncConfigCredentials(ctx); err != nil {
			log.Error().Err(err).Msg("failed to sync storage credentials to database")
		}
	})
}

// WithEntitlements sets up the entitlements client for the server which currently only supports stripe
func WithEntitlements() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if s.Config.Settings.Entitlements.Enabled {
			client, err := entitlements.NewStripeClient(
				entitlements.WithAPIKey(s.Config.Settings.Entitlements.PrivateStripeKey),
				entitlements.WithConfig(s.Config.Settings.Entitlements))
			if err != nil {
				log.Panic().Err(err).Msg("Error creating entitlements client")
			}

			s.Config.Handler.Entitlements = client
		}
	})
}

// WithSummarizer sets up the logic for summarizing long blurbs of texts
func WithSummarizer() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		client, err := summarizer.NewSummarizer(s.Config.Settings.EntConfig)
		if err != nil {
			log.Panic().Err(err).Msg("error creating Summarizer client")
		}

		s.Config.Handler.Summarizer = client
	})
}

// WithWindmill sets up the Windmill workflow automation client
func WithWindmill() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		client, err := windmill.NewWindmill(s.Config.Settings.EntConfig)
		if err != nil {
			if errors.Is(err, windmill.ErrWindmillDisabled) {
				log.Info().Msg("Windmill is not enabled, skipping Windmill client setup")
				return
			}

			log.Panic().Err(err).Msg("error creating Windmill client")
		}

		s.Config.Handler.Windmill = client
	})
}

// WithKeyDirOption allows the key directory to be set via server config.
func WithKeyDirOption() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if s.Config.Settings.Keywatcher.Enabled && s.Config.Settings.Keywatcher.KeyDir != "" {
			WithKeyDir(s.Config.Settings.Keywatcher.KeyDir).apply(s)
			WithKeyDirWatcher(s.Config.Settings.Keywatcher.KeyDir).apply(s)
		}
	})
}

// WithSecretManagerKeysOption allows the secret manager secret name to be set via server config.
func WithSecretManagerKeysOption() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if s.Config.Settings.Keywatcher.SecretManagerSecret != "" && s.Config.Settings.Keywatcher.ExternalSecretsIntegration {
			WithSecretManagerKeys(s.Config.Settings.Server.SecretManagerSecret).apply(s)
		}
	})
}

// WithCSRF sets up the CSRF middleware for the server
func WithCSRF() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if s.Config.Settings.Server.CSRFProtection.Enabled {
			cfg := &s.Config.Settings.Server.CSRFProtection

			// add the CSRF middleware for all requests, using the graph middleware and
			// additional middleware for REST requests
			// this ensures it can be applied after any auth middleware
			s.Config.GraphMiddleware = append(s.Config.GraphMiddleware, csrf.Middleware(cfg))
			s.Config.Handler.AdditionalMiddleware = append(s.Config.Handler.AdditionalMiddleware, csrf.Middleware(cfg))
		}
	})
}

// WithCrypto initializes the field level encryption system for the server
func WithCrypto() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if err := crypto.Init(s.Config.Settings.Server.FieldLevelEncryption); err != nil {
			log.Panic().Err(err).Msg("Error initializing crypto")
		}

		log.Info().Bool("enabled", s.Config.Settings.Server.FieldLevelEncryption.Enabled).Msg("Crypto initialized")
	})
}

const (
	storageValidationTimeout     = 10 * time.Second
	storageCredentialSyncTimeout = 10 * time.Second
)

func validateConfiguredStorageProviders(cfg storage.ProviderConfig) {
	if !cfg.Enabled {
		log.Info().Msg("object storage disabled; skipping validation")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), storageValidationTimeout)
	defer cancel()

	if cfg.DevMode {
		if cfg.Providers.Disk.Enabled {
			bucket := cfg.Providers.Disk.Bucket
			if bucket == "" {
				bucket = objects.DefaultDevStorageBucket
			}

			if err := ensureDirectoryExists(bucket); err != nil {
				log.Warn().Err(err).Str("path", bucket).Msg("unable to ensure development storage directory")
			}
		}

		return
	}

	validateDiskProvider(ctx, cfg.Providers.Disk)
	validateS3Provider(ctx, cfg.Providers.S3)
	validateR2Provider(ctx, cfg.Providers.CloudflareR2)

	if cfg.Providers.GCS.Enabled {
		log.Warn().Msg("skipping GCS provider validation (not implemented)")
	}
}

func validateDiskProvider(ctx context.Context, cfg storage.ProviderConfigs) {
	if !cfg.Enabled {
		return
	}

	bucket := cfg.Bucket
	if bucket == "" {
		bucket = objects.DefaultDevStorageBucket
	}

	options := storage.NewProviderOptions(
		storage.WithBucket(bucket),
		storage.WithBasePath(bucket),
	)
	if cfg.Endpoint != "" {
		options.Apply(storage.WithLocalURL(cfg.Endpoint))
	}

	provider, err := disk.NewDiskBuilder().WithCredentials(cfg.Credentials).WithConfig(options).Build(ctx)
	if err != nil {
		log.Warn().Err(err).Str("provider", "disk").Msg("disk storage validation failed")
		return
	}
	defer provider.Close()

	validateBuckets("disk", provider, bucket)
}

func validateS3Provider(ctx context.Context, cfg storage.ProviderConfigs) {
	if !cfg.Enabled {
		return
	}

	options := storage.NewProviderOptions()
	if cfg.Bucket != "" {
		options.Apply(storage.WithBucket(cfg.Bucket))
	}
	region := cfg.Region
	if region == "" {
		region = objects.DefaultS3Region
	}
	options.Apply(storage.WithRegion(region))
	if cfg.Endpoint != "" {
		options.Apply(storage.WithEndpoint(cfg.Endpoint))
	}

	provider, err := s3provider.NewS3Builder().WithCredentials(cfg.Credentials).WithConfig(options).Build(ctx)
	if err != nil {
		log.Warn().Err(err).Str("provider", "s3").Msg("s3 storage validation failed")
		return
	}
	defer provider.Close()

	validateBuckets("s3", provider, cfg.Bucket)
}

func validateR2Provider(ctx context.Context, cfg storage.ProviderConfigs) {
	if !cfg.Enabled {
		return
	}

	options := storage.NewProviderOptions()
	if cfg.Bucket != "" {
		options.Apply(storage.WithBucket(cfg.Bucket))
	}
	if cfg.Endpoint != "" {
		options.Apply(storage.WithEndpoint(cfg.Endpoint))
	}

	provider, err := r2provider.NewR2Builder().WithCredentials(cfg.Credentials).WithConfig(options).Build(ctx)
	if err != nil {
		log.Warn().Err(err).Str("provider", "r2").Msg("r2 storage validation failed")
		return
	}
	defer provider.Close()

	validateBuckets("r2", provider, cfg.Bucket)
}

func validateBuckets(providerName string, provider storagetypes.Provider, expectedBucket string) {
	buckets, err := provider.ListBuckets()
	if err != nil {
		log.Warn().Err(err).Str("provider", providerName).Msg("unable to list buckets during storage validation")
		return
	}

	log.Info().Str("provider", providerName).Strs("available_buckets", buckets).Msg("storage provider connectivity verified")

	if expectedBucket != "" && !slices.Contains(buckets, expectedBucket) {
		log.Warn().Str("provider", providerName).Str("expected_bucket", expectedBucket).Strs("available_buckets", buckets).Msg("configured bucket not present in provider")
	}
}

func ensureDirectoryExists(path string) error {
	if path == "" {
		return nil
	}

	return os.MkdirAll(path, os.ModePerm)
}
