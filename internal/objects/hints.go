package objects

import (
	"context"
	"strconv"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/entitlements/features"
	"github.com/theopenlane/core/pkg/cp"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/objects/storage"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
)

var (
	hintModuleKey            = cp.NewHintKey[models.OrgModule]("storage.module")
	hintPreferredProviderKey = cp.NewHintKey[storagetypes.ProviderType]("storage.preferred_provider")
	hintKnownProviderKey     = cp.NewHintKey[storagetypes.ProviderType]("storage.known_provider")
	hintSizeBytesKey         = cp.NewHintKey[int64]("storage.size_bytes")
)

// ModuleHintKey returns the typed hint key used to store feature module metadata in the context.
func ModuleHintKey() cp.HintKey[models.OrgModule] {
	return hintModuleKey
}

// PreferredProviderHintKey returns the hint key for explicitly preferred provider selections.
func PreferredProviderHintKey() cp.HintKey[storagetypes.ProviderType] {
	return hintPreferredProviderKey
}

// KnownProviderHintKey returns the hint key for known provider selections.
func KnownProviderHintKey() cp.HintKey[storagetypes.ProviderType] {
	return hintKnownProviderKey
}

// SizeBytesHintKey returns the hint key representing the payload size in bytes.
func SizeBytesHintKey() cp.HintKey[int64] {
	return hintSizeBytesKey
}

// PopulateProviderHints ensures standard metadata is present on the file's provider hints.
func PopulateProviderHints(file *storage.File, orgID string) {
	if file == nil {
		return
	}

	hints := file.ProviderHints
	if hints == nil {
		hints = &storage.ProviderHints{}
		file.ProviderHints = hints
	}

	if hints.Metadata == nil {
		hints.Metadata = map[string]string{}
	}

	if orgID != "" && hints.OrganizationID == "" {
		hints.OrganizationID = orgID
	}

	if file.FieldName != "" {
		hints.Metadata["key"] = file.FieldName
	}

	if file.CorrelatedObjectType != "" {
		hints.Metadata["object_type"] = file.CorrelatedObjectType
	}

	if size := file.FileMetadata.Size; size > 0 {
		hints.Metadata["size_bytes"] = strconv.FormatInt(size, 10)
	}

	if module, ok := ResolveModuleFromFile(*file); ok {
		hints.Module = module
		hints.Metadata["module"] = string(module)
	}
}

// ResolveModuleFromFile attempts to determine the module associated with the upload.
func ResolveModuleFromFile(f storage.File) (models.OrgModule, bool) {
	if module, ok := moduleFromHints(f.ProviderHints); ok {
		return module, true
	}

	if f.CorrelatedObjectType != "" {
		featureKey := lo.PascalCase(f.CorrelatedObjectType)
		if modules, ok := features.FeatureOfType[featureKey]; ok && len(modules) > 0 {
			return modules[0], true
		}
	}

	return "", false
}

// ApplyProviderHints injects hint values into the resolution context.
func ApplyProviderHints(ctx context.Context, hints *storagetypes.ProviderHints) context.Context {
	if hints == nil {
		return ctx
	}

	hintSet := cp.NewHintSet()

	if module, ok := moduleFromHints(hints); ok {
		cp.AddHint(hintSet, hintModuleKey, module)
	}

	if hints.PreferredProvider != "" {
		cp.AddHint(hintSet, hintPreferredProviderKey, hints.PreferredProvider)
	}

	if hints.KnownProvider != "" {
		cp.AddHint(hintSet, hintKnownProviderKey, hints.KnownProvider)
	}

	if sizeStr, ok := hints.Metadata["size_bytes"]; ok {
		if size, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
			cp.AddHint(hintSet, hintSizeBytesKey, size)
		}
	}

	return hintSet.Apply(ctx)
}

func moduleFromHints(hints *storagetypes.ProviderHints) (models.OrgModule, bool) {
	if hints == nil {
		return "", false
	}

	if hints.Module != nil {
		switch v := hints.Module.(type) {
		case models.OrgModule:
			if v != "" {
				return v, true
			}
		case string:
			if v != "" {
				return models.OrgModule(v), true
			}
		}
	}

	if hints.Metadata != nil {
		if module, ok := hints.Metadata["module"]; ok && module != "" {
			return models.OrgModule(module), true
		}
	}

	return "", false
}
