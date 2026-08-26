package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// QrisPayment holds the schema definition for the QrisPayment entity.
//
// zcloud: QRIS payment via gomerch-manager (D4/D5). Status lifecycle:
// pending -> paid | expired | review_required | failed
type QrisPayment struct {
	ent.Schema
}

func (QrisPayment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "qris_payments"},
	}
}

func (QrisPayment) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").SchemaType(map[string]string{dialect.Postgres: "uuid"}).DefaultFunc(uuid.NewString),
		field.Int64("user_id"),
		field.Int64("amount_idr"),
		field.String("payment_ref").MaxLen(100).Unique(),
		field.String("qr_string").SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("image_base64").SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Time("expires_at").SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.String("status").MaxLen(20).Default("pending"),
		field.JSON("gomerch_payload", map[string]any{}).Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.String("idempotency_key").MaxLen(100).Optional().Nillable(),
		field.Time("created_at").Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (QrisPayment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("payment_orders", PaymentOrder.Type),
	}
}

func (QrisPayment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("status"),
		index.Fields("payment_ref").Unique(),
	}
}
