package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"entgo.io/contrib/entgql"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/sprisa/west/util/errutil"
	l "github.com/sprisa/west/util/log"
	"github.com/sprisa/west/util/sig"
	"github.com/sprisa/west/westport/db"
	"github.com/sprisa/west/westport/db/ent"
	"github.com/sprisa/west/westport/db/migrate"
	"github.com/sprisa/west/westport/gql"
	"github.com/vektah/gqlparser/v2/ast"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx := sig.ShutdownContext(context.Background())
	client, err := db.OpenDB()
	errutil.InvariantError(err, "error opening db")
	defer client.Close()
	err = migrate.MigrateClient(ctx, client)
	errutil.InvariantError(err, "error migrating db")

	settings, err := client.Settings.Query().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			settings, err = client.Settings.Create().Save(ctx)
			errutil.InvariantError(err, "error initializing settings")
		} else {
			errutil.InvariantError(err, "error fetching settings")
		}
	}

	l.Log.Info().Msgf("settings: %+v", settings)

	group, _ := errgroup.WithContext(ctx)

	group.Go(func() error {
		handler := NewGQLServer(gql.NewSchema(client), client)
		address := ":3003"
		mux := http.NewServeMux()
		server := &http.Server{Addr: address, Handler: mux}
		mux.Handle(
			"/graphql",
			handler,
		)

		l.Log.Info().Msgf("GQL Server up at: http://localhost%s", server.Addr)
		go func() {
			<-ctx.Done()
			l.Log.Info().Msg("Shutting down gql server")
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
			defer cancel()
			err := server.Shutdown(ctx)
			if err != nil && errors.Is(err, http.ErrServerClosed) == false {
				l.Log.Err(err).Msg("gql server shutdown")
			}
		}()
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	})

	err = group.Wait()
	if err != nil {
		l.Log.Err(err).Msg("server error")
	}
	l.Log.Info().Msg("Goodbye")
}

func NewGQLServer(es graphql.ExecutableSchema, client *ent.Client) *handler.Server {
	// From handler.NewDefaultServer
	srv := handler.New(es)
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.POST{})
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})
	srv.Use(entgql.Transactioner{TxOpener: client})

	return srv
}
