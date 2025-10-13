package schema

import (
	"crypto/rand"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/sprisa/west/westport/db/helpers"
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
		field.Bytes("ca_crt").
			GoType(helpers.EncryptedBytes{}),
		field.Bytes("ca_key").
			Sensitive().
			GoType(helpers.EncryptedBytes{}),
		field.Bytes("lighthouse_crt").
			Sensitive().
			GoType(helpers.EncryptedBytes{}),
		field.Bytes("lighthouse_key").
			Sensitive().
			GoType(helpers.EncryptedBytes{}),
		field.String("cidr").
			GoType(helpers.IpCidr{}).
			Comment("Network cidr range"),
		// field.Bytes("hmac").
		// 	Sensitive().
		// 	GoType(helpers.EncryptedBytes{}).
		// 	DefaultFunc(func() helpers.EncryptedBytes {
		// 		key, err := GenerateHMAC(64)
		// 		if err != nil {
		// 			l.Log.Panic().Err(err).Msg("error generating hmac key")
		// 		}
		// 		return helpers.EncryptedBytes(key)
		// 	}).
		// 	Comment("HS512"),
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
