package schema

import (
	"crypto/rand"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	l "github.com/sprisa/west/util/log"
	"github.com/sprisa/west/westport/db/mixin"
)

type Settings struct {
	ent.Schema
}

func (Settings) Fields() []ent.Field {

	return []ent.Field{
		field.String("cipher").
			Default("aes").
			Comment("Nebula cipher. aes or chachapoly"),
		field.Bytes("hmac").
			Sensitive().
			DefaultFunc(func() []byte {
				key, err := GenerateHMAC(64)
				if err != nil {
					l.Log.Fatal().Err(err).Msg("error generated hmac key")
				}
				return key
			}).
			Comment("HS512"),
	}
}

func (Settings) Edges() []ent.Edge {
	return []ent.Edge{}
}

func (Settings) Indexes() []ent.Index {
	return []ent.Index{}
}

func (Settings) Annotations() []schema.Annotation {
	return []schema.Annotation{}
}

func (Settings) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeMixin{},
	}
}

// func (Settings) Policy() ent.Policy {
// 	return policy.DefaultPolicy(privacy.Policy{
// 		Mutation: privacy.MutationPolicy{
// 		},
// 		Query: privacy.QueryPolicy{
// 		},
// 	})
// }

func GenerateHMAC(length int) ([]byte, error) {
	key := make([]byte, length)
	_, err := rand.Read(key)
	return key, err
}
