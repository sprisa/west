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
	"github.com/cqroot/prompt"
	"github.com/cqroot/prompt/input"
	"github.com/sprisa/west"
	"github.com/sprisa/west/config"
	"github.com/sprisa/west/util/errutil"
	l "github.com/sprisa/west/util/log"
	"github.com/sprisa/west/westport/db"
	"github.com/sprisa/west/westport/db/ent"
	"github.com/sprisa/west/westport/db/helpers"
	"github.com/sprisa/west/westport/db/migrate"
	"github.com/sprisa/west/westport/gql"
	"github.com/urfave/cli/v3"
	"github.com/vektah/gqlparser/v2/ast"
	"golang.org/x/sync/errgroup"
)

var StartCommand = &cli.Command{
	Name:      "start",
	Usage:     "Start west port",
	UsageText: "west-port start",
	Flags:     []cli.Flag{},
	Action: func(ctx context.Context, c *cli.Command) error {
		return startWestPort(ctx)
	},
}

func startWestPort(ctx context.Context) error {
	pswd, err := prompt.New().Ask("password:").
		Input("", input.WithEchoMode(input.EchoPassword), input.WithHelp(true))
	if err != nil {
		return err
	}
	copy(helpers.EncryptionKey[:], pswd)
	// l.Log.Info().Msg(pswd)
	// l.Log.Info().Msgf("key: %s", string(helpers.EncryptionKey[:]))

	client, err := db.OpenDB()
	if err != nil {
		return errutil.WrapError(err, "error opening db")
	}
	defer client.Close()
	err = migrate.MigrateClient(ctx, client)
	if err != nil {
		return errutil.WrapError(err, "error migrating db")
	}

	settings, err := client.Settings.Query().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			settings, err = client.Settings.Create().
				SetCa([]byte("todo")).
				SetCaKey([]byte("todo")).
				Save(ctx)
			if err != nil {
				return errutil.WrapError(err, "error initializing settings")
			}
		} else {
			return errutil.WrapError(err, "error initializing settings")
		}
	}

	l.Log.Info().Msgf("settings: %+v", settings)
	l.Log.Info().Msgf("hmac: %+v", settings.Hmac.String())

	group, ctx := errgroup.WithContext(ctx)

	// Start Graphql API Server
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

	// Start Nebula
	group.Go(func() error {
		return nil
		cipher := config.CipherAes
		if settings.Cipher == "chachapoly" {
			cipher = config.CipherChaChaPoly
		}

		opts := &west.ServerArgs{
			Config: &config.Config{
				Lighthouse: config.Lighthouse{
					Am_lighthouse: true,
				},
				Tun: config.Tun{
					Disabled: true,
				},
				Listen: config.Listen{
					Host: "::",
					Port: 4242,
				},
				Preferred_ranges: config.DefaultPreferredRanges,
				Cipher:           cipher,
			},
		}
		srv, err := west.NewServer(opts)
		if err != nil {
			return errutil.WrapError(err, "error creating nebula server")
		}

		return srv.Listen(ctx)
	})

	err = group.Wait()
	if err != nil {
		return errutil.WrapError(err, "server error")
	}
	l.Log.Info().Msg("Goodbye")
	return nil
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
