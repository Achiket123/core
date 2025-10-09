package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"github.com/stretchr/testify/require"
	echo "github.com/theopenlane/echox"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	handlerpkg "github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/internal/objects/resolver"
	"github.com/theopenlane/core/internal/objects/upload"
	"github.com/theopenlane/core/pkg/objects/storage"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
)

const testTokenIssuer = "http://localhost:17608"

func (suite *HandlerTestSuite) TestDatabaseFileDownloadHandler_Success() {
	t := suite.T()

	restore := suite.swapObjectStoreToDatabase()
	t.Cleanup(restore)

	user := suite.userBuilder(context.Background())

	uploadCtx := auth.NewTestContextWithOrgID(user.ID, user.PersonalOrgID)
	uploadCtx = privacy.DecisionContext(uploadCtx, privacy.Allow)
	uploadCtx = ent.NewContext(uploadCtx, suite.db)

	fileContent := []byte("database provider payload")

	uploadCtx, uploadedFiles, err := upload.HandleUploads(uploadCtx, suite.objectStore, []storage.File{
		{
			RawFile:              bytes.NewReader(fileContent),
			OriginalName:         "example.txt",
			FieldName:            "uploadFile",
			CorrelatedObjectID:   user.PersonalOrgID,
			CorrelatedObjectType: "organization",
			FileMetadata: storage.FileMetadata{
				ContentType: "text/plain",
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, uploadedFiles, 1)

	uploaded := uploadedFiles[0]
	require.NotEmpty(t, uploaded.ID)

	fileRecord, err := suite.db.File.Get(uploadCtx, uploaded.ID)
	require.NoError(t, err)

	presignedURL, err := suite.objectStore.GetPresignedURL(uploadCtx, &storagetypes.File{
		ID: uploaded.ID,
		FileMetadata: storagetypes.FileMetadata{
			Key:          fileRecord.StoragePath,
			Bucket:       fileRecord.StorageVolume,
			ContentType:  fileRecord.DetectedContentType,
			ProviderType: storagetypes.DatabaseProvider,
		},
	}, time.Minute)
	require.NoError(t, err)

	parsedURL, err := url.Parse(presignedURL)
	require.NoError(t, err)
	require.NotEmpty(t, parsedURL.Query().Get("token"))

	requestURL := presignedURL
	if !strings.HasPrefix(requestURL, "http") {
		requestURL = "http://localhost" + requestURL
	}

	req := httptest.NewRequest(http.MethodGet, requestURL, nil)
	req = req.WithContext(uploadCtx)
	rec := httptest.NewRecorder()

	ecCtx := suite.e.NewContext(req, rec)
	ecCtx.SetPathParams(echo.PathParams{{Name: "id", Value: uploaded.ID}})

	err = suite.h.DatabaseFileDownloadHandler(ecCtx, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, fileContent, rec.Body.Bytes())
	require.Contains(t, rec.Header().Get(echo.HeaderContentDisposition), "example.txt")
	require.Equal(t, "text/plain", rec.Header().Get(echo.HeaderContentType))
}

func (suite *HandlerTestSuite) TestDatabaseFileDownloadHandler_InvalidToken() {
	t := suite.T()

	restore := suite.swapObjectStoreToDatabase()
	t.Cleanup(restore)

	user := suite.userBuilder(context.Background())

	uploadCtx := auth.NewTestContextWithOrgID(user.ID, user.PersonalOrgID)
	uploadCtx = privacy.DecisionContext(uploadCtx, privacy.Allow)
	uploadCtx = ent.NewContext(uploadCtx, suite.db)

	fileContent := []byte("invalid token test")

	uploadCtx, uploadedFiles, err := upload.HandleUploads(uploadCtx, suite.objectStore, []storage.File{
		{
			RawFile:              bytes.NewReader(fileContent),
			OriginalName:         "invalid.txt",
			FieldName:            "uploadFile",
			CorrelatedObjectID:   user.PersonalOrgID,
			CorrelatedObjectType: "organization",
			FileMetadata: storage.FileMetadata{
				ContentType: "text/plain",
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, uploadedFiles, 1)

	uploaded := uploadedFiles[0]
	require.NotEmpty(t, uploaded.ID)

	_, err = suite.db.File.Get(uploadCtx, uploaded.ID)
	require.NoError(t, err)

	requestURL := "http://localhost/v1/files/" + uploaded.ID + "/download?token=invalid"
	req := httptest.NewRequest(http.MethodGet, requestURL, nil)
	req = req.WithContext(uploadCtx)
	rec := httptest.NewRecorder()

	ecCtx := suite.e.NewContext(req, rec)
	ecCtx.SetPathParams(echo.PathParams{{Name: "id", Value: uploaded.ID}})

	err = suite.h.DatabaseFileDownloadHandler(ecCtx, nil)
	require.ErrorIs(t, err, handlerpkg.ErrUnauthorized)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func (suite *HandlerTestSuite) swapObjectStoreToDatabase() func() {
	cfg := storage.ProviderConfig{
		Enabled: true,
		Providers: storage.Providers{
			Database: storage.ProviderConfigs{Enabled: true},
		},
	}

	dbStore := resolver.NewServiceFromConfig(cfg, resolver.WithPresignConfig(func() *tokens.TokenManager {
		return suite.sharedTokenManager
	}, testTokenIssuer, testTokenIssuer))

	originalStore := suite.objectStore
	originalHandlerStore := suite.h.ObjectStore

	suite.objectStore = dbStore
	suite.h.ObjectStore = dbStore

	var originalRouterHandler *handlerpkg.Handler
	if suite.router != nil {
		originalRouterHandler = suite.router.Handler
		if suite.router.Handler == nil {
			suite.router.Handler = suite.h
		}
		suite.router.Handler.ObjectStore = dbStore
	}

	return func() {
		suite.objectStore = originalStore
		suite.h.ObjectStore = originalHandlerStore
		if suite.router != nil {
			if suite.router.Handler != nil {
				suite.router.Handler.ObjectStore = originalHandlerStore
			}
			suite.router.Handler = originalRouterHandler
		}
	}
}
