package handlers

import (
	"errors"

	echo "github.com/theopenlane/echox"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/internal/objects/store"
	"github.com/theopenlane/core/internal/objects/upload"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
	models "github.com/theopenlane/core/pkg/openapi"
)

// FileUploadHandler is responsible for uploading files
func (h *Handler) FileUploadHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	if h.ObjectStore == nil {
		return h.InternalServerError(ctx, errors.New("object storage not configured"), openapi)
	}

	req := ctx.Request()

	if err := req.ParseMultipartForm(h.ObjectStore.MaxSize()); err != nil {
		log.Error().Err(err).Msg("failed to parse multipart form")
		return h.BadRequest(ctx, err, openapi)
	}
	if req.MultipartForm != nil {
		defer func() {
			if removeErr := req.MultipartForm.RemoveAll(); removeErr != nil {
				log.Warn().Err(removeErr).Msg("failed to cleanup multipart form")
			}
		}()
	}

	filesMap, err := pkgobjects.ParseFilesFromSource(req)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse files from request")
		return h.BadRequest(ctx, err, openapi)
	}

	uploads := flattenFileMap(filesMap)
	if len(uploads) == 0 {
		return h.BadRequest(ctx, errors.New("no files uploaded"), openapi)
	}

	newCtx, uploadedFiles, err := upload.HandleUploads(req.Context(), h.ObjectStore, uploads)
	if err != nil {
		log.Error().Err(err).Msg("failed to upload files")
		return h.InternalServerError(ctx, err, openapi)
	}

	newCtx, err = store.AddFilePermissions(newCtx)
	if err != nil {
		log.Error().Err(err).Msg("failed to add file permissions")
		return h.InternalServerError(ctx, err, openapi)
	}

	ctx.SetRequest(req.WithContext(newCtx))

	out := models.UploadFilesReply{Message: "file(s) uploaded successfully"}
	for _, f := range uploadedFiles {
		out.Files = append(out.Files, models.File{
			ID:           f.ID,
			Name:         f.Name,
			PresignedURL: f.PresignedURL,
			MimeType:     f.ContentType,
			ContentType:  f.ContentType,
			MD5:          f.MD5,
			Size:         f.Size,
			CreatedAt:    f.CreatedAt,
			UpdatedAt:    f.UpdatedAt,
		})
	}
	out.FileCount = int64(len(out.Files))

	return h.SuccessBlob(ctx, out)
}

// flattenFileMap combines multiple field-specific file slices into a single collection.
func flattenFileMap(fileMap map[string][]storage.File) []storage.File {
	var uploads []storage.File
	for _, files := range fileMap {
		uploads = append(uploads, files...)
	}

	return uploads
}
