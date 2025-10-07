package cp

import (
	"context"

	"github.com/samber/mo"
)

// Resolution is a struct that represents the result of rule evaluation
type Resolution[Creds any, Conf any] struct {
	ClientType  ProviderType
	Credentials Creds
	Config      Conf
	CacheKey    ClientCacheKey
}

// ResolutionRule is a generic struct that evaluates context and returns resolution
type ResolutionRule[T any, Creds any, Conf any] struct {
	Evaluate func(ctx context.Context) mo.Option[Resolution[Creds, Conf]]
}

// Resolver is a generic struct that handles rule-based client resolution
type Resolver[T any, Creds any, Conf any] struct {
	rules       []ResolutionRule[T, Creds, Conf]
	defaultRule mo.Option[ResolutionRule[T, Creds, Conf]]
}

// NewResolver is a constructor function that creates a resolver
func NewResolver[T any, Creds any, Conf any]() *Resolver[T, Creds, Conf] {
	return &Resolver[T, Creds, Conf]{
		rules: make([]ResolutionRule[T, Creds, Conf], 0),
	}
}

// AddRule is a method that adds a resolution rule to the resolver
func (r *Resolver[T, Creds, Conf]) AddRule(rule ResolutionRule[T, Creds, Conf]) *Resolver[T, Creds, Conf] {
	r.rules = append(r.rules, rule)
	return r
}

// SetDefaultRule sets a fallback rule that always matches
func (r *Resolver[T, Creds, Conf]) SetDefaultRule(rule ResolutionRule[T, Creds, Conf]) *Resolver[T, Creds, Conf] {
	r.defaultRule = mo.Some(rule)
	return r
}

// Resolve evaluates rules and returns the first matching resolution
func (r *Resolver[T, Creds, Conf]) Resolve(ctx context.Context) mo.Option[Resolution[Creds, Conf]] {
	for _, rule := range r.rules {
		if resolution := rule.Evaluate(ctx); resolution.IsPresent() {
			return resolution
		}
	}

	if r.defaultRule.IsPresent() {
		defaultRule := r.defaultRule.MustGet()
		return defaultRule.Evaluate(ctx)
	}

	return mo.None[Resolution[Creds, Conf]]()
}
