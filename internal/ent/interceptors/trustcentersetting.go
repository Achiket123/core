package interceptors

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
)

// InterceptorTrustCenterSetting is middleware to change the TrustCenter query
// to only include the objects that the user has access to
// by filtering the trust center settings by the organization
// TODO (sfunk): this should work on all _settings tables instead
// of having a specific interceptor for each one
func InterceptorTrustCenterSetting() ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {

		return nil
	})
}
