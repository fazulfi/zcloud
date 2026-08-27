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

type ApiKeyModelScope struct {
	ent.Schema
}

func (ApiKeyModelScope) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "api_key_model_scopes"},
	}
}

func (ApiKeyModelScope) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (ApiKeyModelScope) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			DefaultFunc(uuid.NewString),
		field.Int64("api_key_id"),
		field.String("model_id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Optional().
			Nillable(),
		field.Int("rate_limit_per_min").
			Default(60),
		field.Int("rate_limit_per_hour").
			Default(0),
		field.Bool("enabled").
			Default(true),
	}
}

func (ApiKeyModelScope) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("api_key", APIKey.Type).
			Ref("model_scopes").
			Field("api_key_id").
			Unique().
			Required(),
	}
}

func (ApiKeyModelScope) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("api_key_id", "model_id").Unique(),
	}
}
