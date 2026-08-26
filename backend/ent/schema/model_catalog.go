package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type ModelCatalog struct{ ent.Schema }

func (ModelCatalog) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "model_catalog"}}
}

func (ModelCatalog) Mixin() []ent.Mixin { return []ent.Mixin{mixins.TimeMixin{}} }

func (ModelCatalog) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").SchemaType(map[string]string{dialect.Postgres: "uuid"}).DefaultFunc(uuid.NewString),
		field.String("canonical_name").MaxLen(100).NotEmpty().Unique(),
		field.String("public_name").MaxLen(200).NotEmpty(),
		field.Int64("context_window").Optional().Nillable(),
		field.JSON("source_suppliers", []string{}).Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.String("status").MaxLen(20).Default("active"),
	}
}

func (ModelCatalog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("pricing", ModelPricing.Type),
		edge.To("supplier_pricing", SupplierPricing.Type),
		edge.To("balances", ModelBalance.Type),
	}
}

func (ModelCatalog) Indexes() []ent.Index { return []ent.Index{index.Fields("status")} }
