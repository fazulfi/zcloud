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

type SupplierPricing struct{ ent.Schema }

func (SupplierPricing) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "supplier_pricing"}}
}

func (SupplierPricing) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").SchemaType(map[string]string{dialect.Postgres: "uuid"}).DefaultFunc(uuid.NewString),
		field.String("model_id").SchemaType(map[string]string{dialect.Postgres: "uuid"}).Optional().Nillable(),
		field.String("supplier_code").MaxLen(20).NotEmpty(),
		field.Int("version"),
		field.String("tier_label").MaxLen(50).Optional().Nillable(),
		field.String("availability").MaxLen(20).Optional().Nillable(),
		field.Float("input_rate").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.Float("output_rate").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.Float("cached_read_rate").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.Float("cached_write_rate").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "numeric(20,8)"}),
		field.JSON("cache_capabilities", map[string]bool{}).Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Time("effective_from").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("effective_to").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (SupplierPricing) Edges() []ent.Edge {
	return []ent.Edge{edge.From("model", ModelCatalog.Type).Ref("supplier_pricing").Field("model_id").Unique()}
}

func (SupplierPricing) Indexes() []ent.Index {
	return []ent.Index{index.Fields("model_id", "supplier_code", "version", "tier_label").Unique()}
}
