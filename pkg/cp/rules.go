package cp

import (
	"context"

	"github.com/samber/mo"
)

// NewRule creates a rule builder for static resolution
func NewRule[T any, Creds any, Conf any]() *RuleBuilder[T, Creds, Conf] {
	return &RuleBuilder[T, Creds, Conf]{}
}

// RuleBuilder provides an interface for creating static resolution rules
type RuleBuilder[T any, Creds any, Conf any] struct {
	conditions []func(context.Context) bool
}

// DefaultRule creates a rule that always matches (for fallbacks)
func DefaultRule[T any, Creds any, Conf any](resolution Resolution[Creds, Conf]) ResolutionRule[T, Creds, Conf] {
	return ResolutionRule[T, Creds, Conf]{
		Evaluate: func(_ context.Context) mo.Option[Resolution[Creds, Conf]] {
			return mo.Some(resolution)
		},
	}
}

// WhenFunc adds a custom condition function
func (b *RuleBuilder[T, Creds, Conf]) WhenFunc(condition func(context.Context) bool) *RuleBuilder[T, Creds, Conf] {
	b.conditions = append(b.conditions, condition)
	return b
}

// ResolvedProvider represents a resolved provider configuration
type ResolvedProvider[Creds any, Conf any] struct {
	Type        ProviderType
	Credentials Creds
	Config      Conf
}

// Resolve creates a rule that uses a function to resolve the provider
func (b *RuleBuilder[T, Creds, Conf]) Resolve(resolver func(context.Context) (*ResolvedProvider[Creds, Conf], error)) ResolutionRule[T, Creds, Conf] {
	conditions := b.conditions
	return ResolutionRule[T, Creds, Conf]{
		Evaluate: func(ctx context.Context) mo.Option[Resolution[Creds, Conf]] {
			for _, condition := range conditions {
				if !condition(ctx) {
					return mo.None[Resolution[Creds, Conf]]()
				}
			}

			provider, err := resolver(ctx)
			if err != nil || provider == nil {
				return mo.None[Resolution[Creds, Conf]]()
			}

			return mo.Some(Resolution[Creds, Conf]{
				ClientType:  provider.Type,
				Credentials: provider.Credentials,
				Config:      provider.Config,
			})
		},
	}
}
