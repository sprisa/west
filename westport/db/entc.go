//go:build ignore
// +build ignore

package main

import (
	"log"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	l "github.com/sprisa/x/log"
	// _ "github.com/hedwigz/entviz"
)

// Changes GraphQL Generation to opt-in instead of opt-out.
// Schemas must provide an EntGQL annotation to be opted-in.
func DefaultSkipEntGQL() gen.Hook {
	return func(next gen.Generator) gen.Generator {
		return gen.GenerateFunc(func(g *gen.Graph) error {
			for _, node := range g.Nodes {
				_, hasEntGQL := node.Annotations["EntGQL"]
				if !hasEntGQL {
					node.Annotations.Set("EntGQL", entgql.Skip(entgql.SkipAll))

				}
			}
			return next.Generate(g)
		})
	}
}

func main() {
	entgqlExtension, err := entgql.NewExtension(
		entgql.WithConfigPath("../gql/gqlgen.yml"),
		entgql.WithSchemaGenerator(),
		entgql.WithSchemaPath("../gql/ent.graphql"),
		// entgql.WithWhereInputs(true),
		entgql.WithRelaySpec(true),
	)
	if err != nil {
		log.Fatalf("creating entgql extension: %v", err)
	}
	opts := []entc.Option{
		entc.Extensions(
			entgqlExtension,
			// entviz.Extension{}
		),
		entc.FeatureNames(
			"privacy",
			"entql",
			"schema/snapshot",
		),
	}

	err = entc.Generate(
		"./schema",
		&gen.Config{
			Target:  "./ent",
			Package: "github.com/sprisa/west/westport/db/ent",
			Hooks:   []gen.Hook{DefaultSkipEntGQL()},
		},
		opts...,
	)
	if err != nil {
		l.Log.Fatal().Msgf("running ent codegen: %v", err)
	}
}
