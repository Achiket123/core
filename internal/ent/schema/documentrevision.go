package schema

import (
	"fmt"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/entx"
)

// DocumentRevision holds the schema definition for the DocumentRevision entity
type DocumentRevision struct {
	SchemaFuncs

	ent.Schema
}

// SchemaDocumentRevisions the name of the schema in snake case
const SchemaDocumentRevision = "document_revision"

func (DocumentRevision) Name() string {
	return SchemaDocumentRevision
}

func (DocumentRevision) GetType() any {
	return DocumentRevision.Type
}

func (DocumentRevision) PluralName() string {
	return pluralize.NewClient().Plural(SchemaDocumentRevision)
}

// Fields of the DocumentRevision
func (DocumentRevision) Fields() []ent.Field {
	return []ent.Field{
		field.Text("details").
			Optional().
			Annotations(
				entx.FieldSearchable(),
			).
			Comment(fmt.Sprintf("details of the document")),
		field.Enum("status").
			GoType(enums.ApprovalStatus("")).
			Default(enums.ApprovalPending.String()).
			Annotations(
				entgql.OrderField("STATUS"),
			).
			Optional().
			Comment("status of the approval, e.g. pending, approved, rejected"),
		field.Time("approval_date").
			Optional().
			Nillable(),
		field.String("submitted_by_id").
			Optional().
			Unique().
			NotEmpty().
			Comment("the user that submitted the document for approval"),
		field.String("approved_by_id").
			Optional().
			Unique().
			NotEmpty().
			Comment("the user that approved the document"),
		field.String("internal_policy_id").
			Optional().
			Unique().
			Comment("the internal policy the document is related to"),
		field.String("procedure_id").
			Optional().
			Unique().
			Comment("the procedure the document is related to"),
		field.String("action_plan_id").
			Optional().
			Unique().
			Comment("the action plan the document is related to"),
	}
}

// Mixin of the DocumentRevision
func (d DocumentRevision) Mixin() []ent.Mixin {
	return mixinConfig{
		includeRevision: true,
	}.getMixins()
}

// Edges of the DocumentRevision
func (d DocumentRevision) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: d,
			name:       "submitted_by",
			t:          User.Type,
			field:      "submitted_by_id",
			comment:    "the user that submitted the document for approval",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: d,
			name:       "approved_by",
			t:          User.Type,
			field:      "approved_by_id",
			comment:    "the user that approved the document",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: d,
			edgeSchema: InternalPolicy{},
			field:      "internal_policy_id",
			comment:    "the internal policy the document is related to",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: d,
			edgeSchema: Procedure{},
			field:      "procedure_id",
			comment:    "the procedure the document is related to",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: d,
			edgeSchema: ActionPlan{},
			field:      "action_plan_id",
			comment:    "the action plan the document is related to",
		}),
	}
}

// Indexes of the DocumentRevision
func (DocumentRevision) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the DocumentRevision
func (DocumentRevision) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// the AnnotationMixin provides the common annotations for
		// to create all the graphQL goodness; if you need the schema only and not the endpoints, use the below annotation instead and set the mixinConfig to `excludeAnnotations: true

		// if you do not need the graphql bits
		// entgql.Skip(entgql.SkipAll),
		// entx.SchemaGenSkip(true),
		// entx.QueryGenSkip(true)

		// the below annotation adds the entfga policy that will check access to the entity
		// remove this annotation (or replace with another policy) if you want checks to be defined
		// by another object
		// uncomment after the first run
		// entfga.SelfAccessChecks(),
	}
}

// Hooks of the DocumentRevision
func (DocumentRevision) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Interceptors of the DocumentRevision
func (DocumentRevision) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Policy of the DocumentRevision
func (DocumentRevision) Policy() ent.Policy {
	// add the new policy here, the default post-policy is to deny all
	// so you need to ensure there are rules in place to allow the actions you want
	return policy.NewPolicy(
		policy.WithQueryRules(
			// add query rules here, the below is the recommended default
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results

			// // add mutation rules here, the below is the recommended default
			// policy.CheckCreateAccess(),
			// // this needs to be commented out for the first run that had the entfga annotation
			// // the first run will generate the functions required based on the entfa annotation
			// // entfga.CheckEditAccess[*generated.DocumentRevisionMutation](),
		),
	)
}
