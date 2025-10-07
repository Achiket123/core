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

// HintKey defines a typed key for storing hints in context
type HintKey[T any] struct {
	name string
}

// NewHintKey creates a new typed hint key
func NewHintKey[T any](name string) HintKey[T] {
	return HintKey[T]{name: name}
}

// WithHint stores a specific hint value in the context
func WithHint[T any](ctx context.Context, key HintKey[T], value T) context.Context {
	storage := cloneHintStorage(contextx.FromOr(ctx, hintStorage{}))
	storage[key.name] = value
	return contextx.With(ctx, storage)
}

// GetHint retrieves a specific hint value from the context
func GetHint[T any](ctx context.Context, key HintKey[T]) mo.Option[T] {
	if storage, ok := contextx.From[hintStorage](ctx); ok {
		if value, exists := storage[key.name]; exists {
			if typed, ok := value.(T); ok {
				return mo.Some(typed)
			}
		}
	}

	return mo.None[T]()
}

// Name returns the human-readable identifier for the hint key. Primarily useful for debugging/logging.
func (k HintKey[T]) Name() string {
	return k.name
}

type (
	hintStorage map[string]any
)

func cloneHintStorage(in hintStorage) hintStorage {
	if len(in) == 0 {
		return make(hintStorage)
	}

	out := make(hintStorage, len(in)+1)
	for key, value := range in {
		out[key] = value
	}

	return out
}
