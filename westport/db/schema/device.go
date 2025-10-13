package schema

import (
	"fmt"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/anandvarma/namegen"
	"github.com/sprisa/west/util/ipconv"
	"github.com/sprisa/west/westport/db/helpers"
	"github.com/sprisa/west/westport/db/mixin"
)

type Device struct {
	ent.Schema
}

func (Device) Fields() []ent.Field {
	nameSchema := []namegen.DictType{
		namegen.Colors,
		namegen.Animals,
	}
	ngen := namegen.NewWithDicts(nameSchema)
	ngen.SetDelimiter("")

	return []ent.Field{
		field.String("name").
			DefaultFunc(func() string {
				return ngen.Get()
			}).
			Comment("Device name. Unique within the Network"),
		field.Uint32("ip").
			Immutable().
			GoType(ipconv.IP(0)).
			Validate(func(v uint32) error {
				ip := ipconv.IP(v)
				ipv4 := ip.ToIPV4()
				if ipv4 == nil {
					return fmt.Errorf("invalid ipv4 address `%d`", ip)
				}
				return nil
			}).
			Comment("Overlay IPv4 of host"),
		field.String("leased_access_token").
			Sensitive().
			Optional().
			Nillable().
			Comment("Access Token leased to a provisioned device. Can only issue 1 at a time, similar to a lock. Used to verify only 1 instance of the Device is running."),
		field.String("cert_fingerprint").
			Sensitive().
			Comment("Cert fingerprint"),
	}
}

func (Device) Edges() []ent.Edge {
	return []ent.Edge{}
}

func (Device) Indexes() []ent.Index {
	return []ent.Index{
		// IP must be unique
		index.Fields("ip").
			Unique(),
		index.Fields("name").
			Unique(),
	}
}

func (Device) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.Skip(helpers.EntGQLSkipNone),
		entgql.RelayConnection(),
	}
}

func (Device) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeMixin{},
	}
}

// func (Device) Policy() ent.Policy {
// 	return policy.DefaultPolicy(privacy.Policy{
// 		Mutation: privacy.MutationPolicy{
// 		},
// 		Query: privacy.QueryPolicy{
// 		},
// 	})
// }
