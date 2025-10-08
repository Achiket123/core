package upload

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/internal/objects/store"
	"github.com/theopenlane/core/pkg/metrics"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
)

// HandleUploads persists metadata, uploads files to storage, and enriches the request context with uploaded file details.
func HandleUploads(ctx context.Context, svc *objects.Service, files []storage.File) (context.Context, []storage.File, error) {
	if len(files) == 0 {
		return ctx, nil, nil
	}

	var uploadedFiles []storage.File

	for _, file := range files {
		pkgobjects.AddUpload()
		metrics.StartFileUpload()
		startTime := time.Now()

		finish := func(status string) {
			metrics.FinishFileUpload(status, time.Since(startTime).Seconds())
			pkgobjects.DoneUpload()
		}

		orgID, _ := auth.GetOrganizationIDFromContext(ctx)
		if orgID != "" && file.Parent.ID == "" && file.CorrelatedObjectID == "" {
			file.CorrelatedObjectID = orgID
			file.CorrelatedObjectType = "organization"
		}

		entFile, err := store.CreateFileRecord(ctx, file)
		if err != nil {
			log.Error().Err(err).Str("file", file.OriginalName).Msg("failed to create file record")
			finish("error")
			return ctx, nil, err
		}

		uploadOpts := BuildUploadOptions(ctx, file)

		uploadedFile, err := svc.Upload(ctx, file.RawFile, uploadOpts)
		if err != nil {
			log.Error().Err(err).Str("file", file.OriginalName).Msg("failed to upload file")
			finish("error")
			return ctx, nil, err
		}

		if closer, ok := file.RawFile.(io.Closer); ok {
			_ = closer.Close()
		}

		mergeUploadedFileMetadata(uploadedFile, entFile.ID, file)
		if err := store.UpdateFileWithStorageMetadata(ctx, entFile, *uploadedFile); err != nil {
			log.Error().Err(err).Msg("failed to update file metadata")
			finish("error")
			return ctx, nil, err
		}

		uploadedFiles = append(uploadedFiles, *uploadedFile)
		finish("success")
	}

	if len(uploadedFiles) == 0 {
		return ctx, nil, nil
	}

	contextFilesMap := make(storage.Files)
	for _, file := range uploadedFiles {
		fieldName := file.FieldName
		if fieldName == "" {
			fieldName = "uploads"
		}
		contextFilesMap[fieldName] = append(contextFilesMap[fieldName], file)
	}

	ctx = pkgobjects.WriteFilesToContext(ctx, contextFilesMap)
	return ctx, uploadedFiles, nil
}

// BuildUploadOptions prepares upload options enriched with provider hints.
func BuildUploadOptions(ctx context.Context, f storage.File) *storage.UploadOptions {
	if f.ProviderHints == nil {
		f.ProviderHints = &storage.ProviderHints{}
	}

	orgID, _ := auth.GetOrganizationIDFromContext(ctx)
	objects.PopulateProviderHints(&f, orgID)

	contentType := f.ContentType
	if contentType == "" || strings.EqualFold(contentType, "application/octet-stream") {
		if f.RawFile != nil {
			if detected, err := storage.DetectContentType(f.RawFile); err == nil && detected != "" {
				contentType = detected
				f.ContentType = detected
			}
		}
	}

	return &storage.UploadOptions{
		FileName:          f.OriginalName,
		ContentType:       contentType,
		Bucket:            f.Bucket,
		FolderDestination: f.Folder,
		FileMetadata: storage.FileMetadata{
			Key:           f.FieldName,
			ProviderHints: f.ProviderHints,
		},
	}
}

func mergeUploadedFileMetadata(dest *storage.File, entFileID string, src storage.File) {
	dest.ID = entFileID
	dest.FieldName = src.FieldName
	dest.Parent = src.Parent
	dest.CorrelatedObjectID = src.CorrelatedObjectID
	dest.CorrelatedObjectType = src.CorrelatedObjectType
	if len(dest.Metadata) == 0 && len(src.Metadata) > 0 {
		dest.Metadata = src.Metadata
	}
}
