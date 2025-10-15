package mixin

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Stores created and updated time
type TimeMixin struct {
	mixin.Schema
}

func (t TimeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_time").
			Default(time.Now).
			Immutable().
			Comment("Time ent was created"),
		field.Time("updated_time").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("Time ent was updated"),
	}
}

type CreatedTimeMixin struct {
	mixin.Schema
}

func (CreatedTimeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_time").
			Default(time.Now).
			Immutable().
			Comment("Time ent was created"),
	}
}
