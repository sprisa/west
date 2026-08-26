package schema

import (
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/sprisa/west/util/ipconv"
	"github.com/sprisa/west/westport/db/mixin"
)

type Host struct {
	ent.Schema
}

func (Host) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("ip").
			Immutable().
			GoType(ipconv.IP(0)).
			Validate(func(v uint32) error {
				if ipconv.IP(v).ToIPV4() == nil {
					return fmt.Errorf("invalid ipv4 address `%d`", v)
				}
				return nil
			}),
	}
}

func (Host) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("ip").Unique(),
	}
}

func (Host) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeMixin{},
	}
}
