package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/rs/zerolog"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/objects"
)

func HookTrustCenterWatermarkConfig() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterWatermarkConfigFunc(func(ctx context.Context, m *generated.TrustCenterWatermarkConfigMutation) (generated.Value, error) {
			zerolog.Ctx(ctx).Debug().Msg("trust center watermark config hook")

			fileIDs := objects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = checkTrustCenterWatermarkConfigFiles(ctx, m)
				if err != nil {
					return nil, err
				}

				m.SetFileID(fileIDs[0])
			}
			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

func checkTrustCenterWatermarkConfigFiles(ctx context.Context, m *generated.TrustCenterWatermarkConfigMutation) (context.Context, error) {
	key := "logoFile"

	files, _ := objects.FilesFromContextWithKey(ctx, key)
	if len(files) == 0 {
		return ctx, nil
	}

	if len(files) > 1 {
		return ctx, ErrNotSingularUpload
	}

	adapter := objects.NewGenericMutationAdapter(m,
		func(mut *generated.TrustCenterWatermarkConfigMutation) (string, bool) { return mut.ID() },
		func(mut *generated.TrustCenterWatermarkConfigMutation) string { return mut.Type() },
	)

	return objects.ProcessFilesForMutation(ctx, adapter, key, "trust_center_watermark_config")
}
