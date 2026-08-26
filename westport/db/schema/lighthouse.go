package schema

import (
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/sprisa/west/util/ipconv"
	"github.com/sprisa/west/westport/db/helpers"
	"github.com/sprisa/west/westport/db/mixin"
)

type Lighthouse struct {
	ent.Schema
}

func (Lighthouse) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("ip").
			Immutable().
			GoType(ipconv.IP(0)).
			Validate(func(v uint32) error {
				if ipconv.IP(v).ToIPV4() == nil {
					return fmt.Errorf("invalid ipv4 address `%d`", v)
				}
				return nil
			}).
			Comment("Nebula overlay IPv4 of the lighthouse"),
		field.String("endpoint").
			Comment("Public Nebula host and UDP port"),
		field.Bytes("certificate").
			Sensitive().
			GoType(helpers.EncryptedBytes{}),
		field.Bytes("key").
			Sensitive().
			GoType(helpers.EncryptedBytes{}),
		field.String("api_endpoint").
			Comment("Physical API host:port for this node"),
	}
}

func (Lighthouse) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("host", Host.Type).Required().Unique(),
	}
}

func (Lighthouse) Indexes() []ent.Index {
	return []ent.Index{index.Fields("ip").Unique()}
}

func (Lighthouse) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.TimeMixin{}}
}
