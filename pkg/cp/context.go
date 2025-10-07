package cp

import (
	"context"

	"github.com/samber/mo"
	"github.com/theopenlane/utils/contextx"
)

// WithValue is a simplified helper that adds any typed value directly to context
func WithValue[T any](ctx context.Context, value T) context.Context {
	return contextx.With(ctx, value)
}

// GetValue is a simplified helper that retrieves any typed value directly from context
func GetValue[T any](ctx context.Context) mo.Option[T] {
	if value, ok := contextx.From[T](ctx); ok {
		return mo.Some(value)
	}

	return mo.None[T]()
}

// GetValueEquals checks if a context value equals the expected value
func GetValueEquals[T comparable](ctx context.Context, expected T) bool {
	return GetValue[T](ctx).OrElse(*new(T)) == expected
}
