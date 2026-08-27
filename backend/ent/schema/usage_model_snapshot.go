package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// UsageModelSnapshot holds per-user per-model dual metering rollups.
type UsageModelSnapshot struct {
	ent.Schema
}

func (UsageModelSnapshot) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.String("model").MaxLen(100).NotEmpty(),
		field.Int("pricing_version").Default(0),
		field.Float("display_input_cost").SchemaType(map[string]string{"postgres": "numeric(20,8)"}).Default(0),
		field.Float("display_output_cost").SchemaType(map[string]string{"postgres": "numeric(20,8)"}).Default(0),
		field.Float("display_cache_read_cost").SchemaType(map[string]string{"postgres": "numeric(20,8)"}).Default(0),
		field.Float("display_cache_write_cost").SchemaType(map[string]string{"postgres": "numeric(20,8)"}).Default(0),
		field.Float("display_total_cost").SchemaType(map[string]string{"postgres": "numeric(20,8)"}).Default(0),
		field.Float("display_blend_cost").SchemaType(map[string]string{"postgres": "numeric(20,8)"}).Default(0),
		field.Float("cost_input").SchemaType(map[string]string{"postgres": "numeric(20,8)"}).Default(0),
		field.Float("cost_output").SchemaType(map[string]string{"postgres": "numeric(20,8)"}).Default(0),
		field.Float("cost_cache_read").SchemaType(map[string]string{"postgres": "numeric(20,8)"}).Default(0),
		field.Float("cost_cache_write").SchemaType(map[string]string{"postgres": "numeric(20,8)"}).Default(0),
		field.Float("cost_total").SchemaType(map[string]string{"postgres": "numeric(20,8)"}).Default(0),
		field.String("cost_supplier_code").MaxLen(20).Default(""),
		field.Int64("input_tokens").Default(0),
		field.Int64("output_tokens").Default(0),
		field.Int64("cache_read_tokens").Default(0),
		field.Int64("cache_write_tokens").Default(0),
		field.Float("usage_model_pct").SchemaType(map[string]string{"postgres": "numeric(10,6)"}).Default(0),
	}
}

func (UsageModelSnapshot) Edges() []ent.Edge {
	return nil
}

func (UsageModelSnapshot) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

func (UsageModelSnapshot) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "model", "pricing_version").Unique(),
	}
}
