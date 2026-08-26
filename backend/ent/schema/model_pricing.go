package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type ModelPricing struct{ ent.Schema }

func (ModelPricing) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "model_pricing"}}
}

func (ModelPricing) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").SchemaType(map[string]string{dialect.Postgres: "uuid"}).DefaultFunc(uuid.NewString),
		field.String("model_id").SchemaType(map[string]string{dialect.Postgres: "uuid"}).Optional().Nillable(),
		field.Int("version"),
		field.Float("input_rate").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.Float("output_rate").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.Float("cached_read_rate").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.Float("cached_write_rate").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.String("context_tier").MaxLen(20).Optional().Nillable(),
		field.Int64("tokens_per_dollar").Optional().Nillable(),
		field.Float("pct_per_1m_tokens").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "numeric(10,6)"}),
		field.Time("effective_from").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("effective_to").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.String("source_ref").MaxLen(100).Optional().Nillable(),
	}
}

func (ModelPricing) Edges() []ent.Edge {
	return []ent.Edge{edge.From("model", ModelCatalog.Type).Ref("pricing").Field("model_id").Unique()}
}

func (ModelPricing) Indexes() []ent.Index {
	return []ent.Index{index.Fields("model_id", "version").Unique()}
}
