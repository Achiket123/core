package objects

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/cp"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/objects/storage"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

func TestPopulateProviderHints(t *testing.T) {
	file := storage.File{
		OriginalName:         "evidence.json",
		FieldName:            "uploadFile",
		CorrelatedObjectType: "evidence",
		FileMetadata: storage.FileMetadata{
			Size:        1024,
			ContentType: "application/json",
		},
	}

	PopulateProviderHints(&file, "org-123")

	require.NotNil(t, file.ProviderHints)
	assert.Equal(t, "org-123", file.ProviderHints.OrganizationID)
	assert.Equal(t, models.CatalogComplianceModule, file.ProviderHints.Module)
	assert.Equal(t, "uploadFile", file.ProviderHints.Metadata["key"])
	assert.Equal(t, "evidence", file.ProviderHints.Metadata["object_type"])
	assert.Equal(t, "1024", file.ProviderHints.Metadata["size_bytes"])
	assert.Equal(t, string(models.CatalogComplianceModule), file.ProviderHints.Metadata["module"])
}

func TestApplyProviderHints(t *testing.T) {
	hints := &storagetypes.ProviderHints{
		PreferredProvider: storagetypes.ProviderType("s3"),
		KnownProvider:     storagetypes.ProviderType("disk"),
		Metadata: map[string]string{
			"size_bytes": "2048",
		},
	}

	module := models.CatalogComplianceModule
	hints.Module = module

	ctx := ApplyProviderHints(context.Background(), hints)

	pref := cp.GetHint(ctx, PreferredProviderHintKey())
	require.True(t, pref.IsPresent())
	assert.Equal(t, storagetypes.ProviderType("s3"), pref.MustGet())

	known := cp.GetHint(ctx, KnownProviderHintKey())
	require.True(t, known.IsPresent())
	assert.Equal(t, storagetypes.ProviderType("disk"), known.MustGet())

	resModule := cp.GetHint(ctx, ModuleHintKey())
	require.True(t, resModule.IsPresent())
	assert.Equal(t, module, resModule.MustGet())

	size := cp.GetHint(ctx, SizeBytesHintKey())
	require.True(t, size.IsPresent())
	assert.Equal(t, int64(2048), size.MustGet())
}
