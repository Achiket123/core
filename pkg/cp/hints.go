package cp

import "context"

// HintSet accumulates typed hints that can later be applied to a context.
type HintSet struct {
	hints []func(context.Context) context.Context
}

// NewHintSet creates an empty HintSet.
func NewHintSet() *HintSet {
	return &HintSet{hints: make([]func(context.Context) context.Context, 0)}
}

// AddHint appends a typed hint to the set.
func AddHint[T any](hs *HintSet, key HintKey[T], value T) {
	if hs == nil {
		return
	}

	hs.hints = append(hs.hints, func(ctx context.Context) context.Context {
		return WithHint(ctx, key, value)
	})
}

// Apply injects all accumulated hints into the provided context.
func (hs *HintSet) Apply(ctx context.Context) context.Context {
	if hs == nil {
		return ctx
	}

	for _, hint := range hs.hints {
		ctx = hint(ctx)
	}

	return ctx
}
