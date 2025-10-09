package serveropts_test

import (
	"context"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	fgatest "github.com/theopenlane/iam/fgax/testutils"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/iam/totp"
	"github.com/theopenlane/riverboat/pkg/riverqueue"
	"github.com/theopenlane/utils/contextx"
	"github.com/theopenlane/utils/testutils"

	"github.com/theopenlane/core/internal/credsync"
	"github.com/theopenlane/core/internal/ent/entconfig"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/entdb"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/internal/httpserve/route"
	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/pkg/eddy"
	"github.com/theopenlane/core/pkg/entitlements/mocks"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/objects/storage"
	coreutils "github.com/theopenlane/core/pkg/testutils"
	"github.com/theopenlane/utils/ulids"

	// import generated runtime which is required to prevent cyclical dependencies
	_ "github.com/theopenlane/core/internal/ent/generated/runtime"
)

func cloneProviderOptions(in *storage.ProviderOptions) *storage.ProviderOptions {
	if in == nil {
		return nil
	}
	return in.Clone()
}

const (
	fgaModelFile = "../../../fga/model/model.fga"
)

// CredentialSyncTestSuite handles the setup and teardown between tests
type CredentialSyncTestSuite struct {
	suite.Suite
	e                    *echo.Echo
	db                   *ent.Client
	api                  *testclient.TestClient
	h                    *handlers.Handler
	router               *route.Router
	tf                   *testutils.TestFixture
	ofgaTF               *fgatest.OpenFGATestFixture
	stripeMockBackend    *mocks.MockStripeBackend
	objectStore          *objects.Service
	sharedTokenManager   *tokens.TokenManager
	sharedRedisClient    *redis.Client
	sharedSessionManager sessions.Store[map[string]any]
	sharedFGAClient      *fgax.Client
	sharedOTPManager     *totp.Client
	sharedPondPool       *soiree.PondPool
	service              *credsync.CredentialSyncService
	testUserID           string
	testOrgID            string
}

// TestCredentialSyncTestSuite runs all the tests in the CredentialSyncTestSuite
func TestCredentialSyncTestSuite(t *testing.T) {
	suite.Run(t, new(CredentialSyncTestSuite))
}

func (suite *CredentialSyncTestSuite) SetupSuite() {
	if testing.Verbose() {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}

	// setup db container
	suite.tf = entdb.NewTestFixture()

	// setup openFGA container
	suite.ofgaTF = fgatest.NewFGATestcontainer(context.Background(),
		fgatest.WithModelFile(fgaModelFile),
		fgatest.WithEnvVars(map[string]string{
			"OPENFGA_MAX_CHECKS_PER_BATCH_CHECK":          "100",
			"OPENFGA_CHECK_ITERATOR_CACHE_ENABLED":        "false",
			"OPENFGA_LIST_OBJECTS_ITERATOR_CACHE_ENABLED": "false",
		}),
	)

	// create shared instances once to avoid expensive recreation in each test
	var err error

	// shared token manager to avoid RSA key generation
	suite.sharedTokenManager, err = coreutils.CreateTokenManager(15 * time.Minute) //nolint:mnd
	require.NoError(suite.T(), err)

	// shared redis client to avoid miniredis server startup
	suite.sharedRedisClient = coreutils.NewRedisClient()

	// shared session manager to avoid random key generation
	suite.sharedSessionManager = coreutils.CreateSessionManager()

	// shared FGA client to avoid repeated container connections
	suite.sharedFGAClient, err = suite.ofgaTF.NewFgaClient(context.Background())
	require.NoError(suite.T(), err)

	// shared pond pool to avoid worker pool creation
	suite.sharedPondPool = soiree.NewPondPool(
		soiree.WithMaxWorkers(100), //nolint:mnd
		soiree.WithName("ent_client_pool"),
	)
}

func (suite *CredentialSyncTestSuite) SetupTest() {
	t := suite.T()

	ctx := context.Background()

	// use all shared instances to avoid expensive recreation
	sessionConfig := sessions.NewSessionConfig(
		suite.sharedSessionManager,
		sessions.WithPersistence(suite.sharedRedisClient),
	)

	sessionConfig.CookieConfig = sessions.DebugOnlyCookieConfig

	opts := []ent.Option{
		ent.Authz(*suite.sharedFGAClient),
		ent.Emailer(&emailtemplates.Config{
			CompanyName: "Meow Inc.",
		}),
		ent.TokenManager(suite.sharedTokenManager),
		ent.SessionConfig(&sessionConfig),
		ent.EntConfig(&entconfig.Config{
			Modules: entconfig.Modules{
				Enabled: false,
			},
		}),
		ent.TOTP(suite.sharedOTPManager),
		ent.PondPool(suite.sharedPondPool),
	}

	// create database connection
	jobOpts := []riverqueue.Option{riverqueue.WithConnectionURI(suite.tf.URI)}

	db, err := entdb.NewTestClient(ctx, suite.tf, jobOpts, opts)
	require.NoError(t, err, "failed opening connection to database")

	suite.objectStore, _, err = coreutils.MockStorageServiceWithValidationAndProvider(t, nil, nil)
	require.NoError(t, err)

	// truncate river tables
	err = db.Job.TruncateRiverTables(ctx)
	require.NoError(t, err)

	// add db to test client
	suite.db = db

	// add the client
	suite.api, err = coreutils.TestClient(suite.db, suite.objectStore)
	require.NoError(t, err)

	// Create a test user and use their personal org for system integrations
	ctx = privacy.DecisionContext(context.Background(), privacy.Allow)
	ctx = ent.NewContext(ctx, suite.db)

	// Create user setting first
	userSetting, err := suite.db.UserSetting.Create().
		SetEmailConfirmed(true).
		Save(ctx)
	require.NoError(t, err)

	// Create the user with proper setup
	suite.testUserID = ulids.New().String()
	userInfo, err := suite.db.User.Create().
		SetID(suite.testUserID).
		SetEmail(suite.testUserID + "@example.com").
		SetFirstName("Test").
		SetLastName("User").
		SetLastLoginProvider(enums.AuthProviderCredentials).
		SetSetting(userSetting).
		Save(ctx)
	require.NoError(t, err)

	// Get the personal org that was auto-created
	personalOrg, err := userInfo.Edges.Setting.DefaultOrg(ctx)
	require.NoError(t, err)

	suite.testOrgID = personalOrg.ID

	// Create credential sync service after database is ready
	clientPool := eddy.NewClientPool[storage.Provider](time.Hour)
	clientService := eddy.NewClientService[
		storage.Provider,
		storage.ProviderCredentials,
		*storage.ProviderOptions,
	](clientPool, eddy.WithConfigClone[
		storage.Provider,
		storage.ProviderCredentials,
		*storage.ProviderOptions,
	](cloneProviderOptions))

	suite.service = credsync.NewCredentialSyncService(
		suite.db,
		clientService,
		&storage.Providers{},
	)
}

func (suite *CredentialSyncTestSuite) TearDownTest() {
	if suite.db != nil {
		err := suite.db.CloseAll()
		require.NoError(suite.T(), err)
	}
}

func (suite *CredentialSyncTestSuite) TearDownSuite() {
	testutils.TeardownFixture(suite.tf)

	// terminate all fga containers
	err := suite.ofgaTF.TeardownFixture()
	require.NoError(suite.T(), err)
}

func (suite *CredentialSyncTestSuite) systemContext() context.Context {
	ctx := context.Background()
	user := &auth.AuthenticatedUser{
		SubjectID:          suite.testUserID,
		SubjectName:        "Credential Sync Test User",
		OrganizationID:     suite.testOrgID,
		OrganizationIDs:    []string{suite.testOrgID},
		AuthenticationType: auth.APITokenAuthentication,
		IsSystemAdmin:      true,
		ActiveSubscription: true,
	}

	ctx = auth.WithAuthenticatedUser(ctx, user)
	ctx = auth.WithSystemAdminContext(ctx, user)
	ctx = contextx.With(ctx, auth.OrganizationCreationContextKey{})
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	return ent.NewContext(ctx, suite.db)
}
