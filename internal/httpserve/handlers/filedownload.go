package handlers

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/oklog/ulid/v2"
	echo "github.com/theopenlane/echox"
	"github.com/vmihailenco/msgpack/v5"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/pkg/objects/storage"
	dbprovider "github.com/theopenlane/core/pkg/objects/storage/providers/database"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/utils/ulids"
)

const (
	tokenPartsExpected     = 2
	objectURIPartsExpected = 3
)

// DatabaseFileDownloadHandler serves files that are stored in the database backend using a presigned token.
func (h *Handler) DatabaseFileDownloadHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	fileID := strings.TrimSpace(ctx.PathParam("id"))
	if fileID == "" {
		return h.BadRequest(ctx, ErrBadRequest, openapi)
	}

	encodedToken := strings.TrimSpace(ctx.QueryParam("token"))
	if encodedToken == "" {
		return h.BadRequest(ctx, ErrBadRequest, openapi)
	}

	if h.ObjectStore == nil {
		return h.InternalServerError(ctx, ErrObjectStoreUnavailable, openapi)
	}

	downloadToken, err := h.verifyDownloadToken(encodedToken)
	if err != nil {
		return h.Unauthorized(ctx, err, openapi)
	}

	requestCtx := ctx.Request().Context()

	if err := validateTokenAuthorization(requestCtx, downloadToken); err != nil {
		return h.Unauthorized(ctx, err, openapi)
	}

	providerType, bucket, key, err := parseObjectURI(downloadToken.ObjectURI)
	if err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	if h.DBClient != nil {
		requestCtx = ent.NewContext(requestCtx, h.DBClient)
	}

	fileEntity, err := h.DBClient.File.Query().Where(file.IDEQ(fileID)).Only(requestCtx)
	if err != nil {
		if ent.IsNotFound(err) {
			return h.NotFound(ctx, ErrNotFound, openapi)
		}

		return h.InternalServerError(ctx, err, openapi)
	}

	if err := validateFileMetadata(fileEntity, providerType, bucket, key, fileID); err != nil {
		return h.Unauthorized(ctx, err, openapi)
	}

	storFile := buildStorageFile(fileEntity, fileID)

	download, err := h.ObjectStore.Download(requestCtx, nil, storFile, &storage.DownloadOptions{})
	if err != nil {
		if errors.Is(err, dbprovider.ErrFileNotFound) || ent.IsNotFound(err) {
			return h.NotFound(ctx, ErrNotFound, openapi)
		}

		return h.InternalServerError(ctx, err, openapi)
	}

	contentType := download.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	fileName := download.Name
	if fileName == "" {
		fileName = fileEntity.ProvidedFileName
		if fileName == "" {
			fileName = fileID
		}
	}

	headers := ctx.Response().Header()
	headers.Set(echo.HeaderContentType, contentType)
	headers.Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s\"", fileName))

	return ctx.Blob(http.StatusOK, contentType, download.File)
}

func (h *Handler) verifyDownloadToken(encodedToken string) (*tokens.DownloadToken, error) {
	unescapedToken, err := url.QueryUnescape(encodedToken)
	if err != nil {
		return nil, ErrUnauthorized
	}

	parts := strings.SplitN(unescapedToken, ".", tokenPartsExpected)
	if len(parts) != tokenPartsExpected {
		return nil, ErrUnauthorized
	}

	signature := parts[0]
	encodedPayload := parts[1]

	payload, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return nil, ErrUnauthorized
	}

	var downloadToken tokens.DownloadToken
	if err := msgpack.Unmarshal(payload, &downloadToken); err != nil {
		return nil, ErrUnauthorized
	}

	secret, ok := h.ObjectStore.LookupDownloadSecret(downloadToken.TokenID)
	if !ok {
		return nil, ErrUnauthorized
	}

	if err := downloadToken.Verify(signature, secret); err != nil {
		return nil, ErrUnauthorized
	}

	return &downloadToken, nil
}

func validateTokenAuthorization(requestCtx context.Context, downloadToken *tokens.DownloadToken) error {
	if !ulids.IsZero(downloadToken.UserID) {
		user, ok := auth.AuthenticatedUserFromContext(requestCtx)
		if !ok || user == nil {
			return ErrUnauthorized
		}

		userULID, err := ulid.Parse(user.SubjectID)
		if err != nil || userULID != downloadToken.UserID {
			return ErrUnauthorized
		}

		if !ulids.IsZero(downloadToken.OrgID) && !userHasOrgULID(user, downloadToken.OrgID) {
			return ErrUnauthorized
		}
	}

	return nil
}

func validateFileMetadata(fileEntity *ent.File, providerType storagetypes.ProviderType, bucket, key, fileID string) error {
	if string(providerType) != fileEntity.StorageProvider {
		return ErrUnauthorized
	}

	if bucket != fileEntity.StorageVolume {
		return ErrUnauthorized
	}

	expectedKey := fileEntity.StoragePath
	if expectedKey == "" {
		expectedKey = fileID
	}

	if key != expectedKey {
		return ErrUnauthorized
	}

	return nil
}

func buildStorageFile(fileEntity *ent.File, fileID string) *storagetypes.File {
	storFile := &storagetypes.File{
		ID:           fileID,
		OriginalName: fileEntity.ProvidedFileName,
		ProviderType: storagetypes.ProviderType(fileEntity.StorageProvider),
		FileMetadata: storagetypes.FileMetadata{
			Key:         fileEntity.StoragePath,
			Bucket:      fileEntity.StorageVolume,
			ContentType: fileEntity.DetectedContentType,
			ProviderHints: &storagetypes.ProviderHints{
				KnownProvider: storagetypes.ProviderType(fileEntity.StorageProvider),
			},
		},
	}

	if storFile.Key == "" {
		storFile.Key = fileID
	}

	if storFile.FileMetadata.ProviderType == "" {
		storFile.FileMetadata.ProviderType = storagetypes.DatabaseProvider
	}

	if storFile.ProviderType == "" {
		storFile.ProviderType = storFile.FileMetadata.ProviderType
	}

	if storFile.ProviderHints != nil && storFile.ProviderHints.KnownProvider == "" {
		storFile.ProviderHints.KnownProvider = storFile.FileMetadata.ProviderType
	}

	storFile.ProviderHints = ensureProviderHints(storFile.ProviderHints)
	if storFile.ProviderHints.KnownProvider == "" {
		storFile.ProviderHints.KnownProvider = storFile.FileMetadata.ProviderType
	}

	storFile.ProviderHints = storFile.FileMetadata.ProviderHints
	if storFile.Bucket == "" {
		storFile.Bucket = fileEntity.StorageVolume
	}

	return storFile
}

func userHasOrgULID(user *auth.AuthenticatedUser, orgID ulid.ULID) bool {
	if user == nil {
		return false
	}

	orgIDStr := orgID.String()

	if user.OrganizationID == orgIDStr {
		return true
	}

	return slices.Contains(user.OrganizationIDs, orgIDStr)
}

func ensureProviderHints(hints *storagetypes.ProviderHints) *storagetypes.ProviderHints {
	if hints == nil {
		return &storagetypes.ProviderHints{}
	}

	return hints
}

func parseObjectURI(objectURI string) (storagetypes.ProviderType, string, string, error) {
	parts := strings.SplitN(objectURI, ":", objectURIPartsExpected)
	if len(parts) != objectURIPartsExpected {
		return "", "", "", ErrBadRequest
	}

	return storagetypes.ProviderType(parts[0]), parts[1], parts[2], nil
}
