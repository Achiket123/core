package mixin

import (
	"context"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/rs/zerolog"

	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

// SystemOwnedMixin implements the revision pattern for schemas.
type SystemOwnedMixin struct {
	mixin.Schema
}

// Fields of the SystemOwnedMixin.
func (SystemOwnedMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("system_owned").
			Optional().
			Default(false).
			Annotations(
				// the field is automatically set to true if the user is a system admin
				// do not allow this field to be set in the mutation manually
				entgql.Skip(entgql.SkipMutationUpdateInput, entgql.SkipMutationCreateInput),
			).
			Immutable(). // don't allow this to be changed after creation, a new record must be created
			Comment("indicates if the record is owned by the the openlane system and not by an organization"),
		field.String("internal_notes").
			Optional().
			Comment("internal notes about the object creation, this field is only available to system admins").
			Nillable(),
		field.String("system_internal_id").
			Optional().
			Comment("an internal identifier for the mapping, this field is only available to system admins").
			Nillable(),
	}
}

// Hooks of the SystemOwnedMixin.
func (d SystemOwnedMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		HookSystemOwnedCreate(),
	}
}

func (SystemOwnedMixin) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Policy of the JobTemplate
func (SystemOwnedMixin) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			rule.AllowMutationIfSystemAdmin(),
			rule.SystemOwnedSchema(),
		),
	)
}

// SystemOwnedMutation is an interface for interacting with the system_owned field in mutations
// it will add the system_owned_field and will automatically set the field to true if the user is a system admin
type SystemOwnedMutation interface {
	utils.GenericMutation

	SetSystemOwned(bool)
	OldSystemOwned(context.Context) (bool, error)
}

// HookSystemOwnedCreate will automatically set the system_owned field to true if the user is a system admin
func HookSystemOwnedCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			admin, err := rule.CheckIsSystemAdminWithContext(ctx)
			if err != nil {
				zerolog.Ctx(ctx).Error().Err(err).Msg("unable to check if user is system admin, skipping setting system owned")

				return next.Mutate(ctx, m)
			}

			if admin {
				mut, ok := m.(SystemOwnedMutation)
				if ok && mut != nil {
					mut.SetSystemOwned(true)
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}

// InterceptorSystemFields handles returning internal only fields for system owned schemas
func InterceptorSystemFields() ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		admin, err := rule.CheckIsSystemAdminWithContext(ctx)
		if err != nil {
			return err
		}

		if admin {
			return nil
		}

		// if not a system admin, do not return system owned fields

		return nil
	})
}
