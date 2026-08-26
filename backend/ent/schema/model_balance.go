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

type ModelBalance struct {
	ent.Schema
}

func (ModelBalance) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "model_balances"},
	}
}

func (ModelBalance) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (ModelBalance) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			DefaultFunc(uuid.NewString),
		field.Int64("user_id"),
		field.String("model_id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}),
		field.Int64("tokens_purchased").
			Default(0),
		field.Int64("tokens_consumed").
			Default(0),
		field.Int64("balance").
			Default(0),
		field.Float("usage_percent").
			SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}).
			Default(0),
		field.String("status").
			MaxLen(20).
			Default("active"),
		field.Int64("version").
			Default(1),
	}
}

func (ModelBalance) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("model", ModelCatalog.Type).
			Ref("balances").
			Field("model_id").
			Unique().
			Required(),
	}
}

func (ModelBalance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("status"),
		index.Fields("user_id", "model_id").Unique(),
	}
}
