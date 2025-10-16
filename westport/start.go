package westport

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"entgo.io/contrib/entgql"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/sprisa/west"
	"github.com/sprisa/west/config"
	"github.com/sprisa/west/util/errutil"
	l "github.com/sprisa/west/util/log"
	"github.com/sprisa/west/westport/db"
	"github.com/sprisa/west/westport/db/ent"
	"github.com/sprisa/west/westport/db/migrate"
	"github.com/sprisa/west/westport/dns"
	"github.com/sprisa/west/westport/gql"
	"github.com/urfave/cli/v3"
	"github.com/vektah/gqlparser/v2/ast"
	"golang.org/x/sync/errgroup"
)

var StartCommand = &cli.Command{
	Name:      "start",
	Usage:     "Start west port",
	UsageText: "west port start",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "private-dns",
			Usage: "DNS server will only be accessible within the network. Requires DNS configuration on each client.",
		},
		&cli.BoolFlag{
			Name:  "disable-tun",
			Usage: "Disabled TUN network binding",
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		return startWestPort(ctx, c)
	},
}

func startWestPort(ctx context.Context, c *cli.Command) error {
	privateDns := c.Bool("private-dns")
	disableTun := c.Bool("disable-tun")

	err := promptEncryptionPassword()
	if err != nil {
		return err
	}

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
			return errors.New("error finding settings. Trying installing first.")
		}
		return errutil.WrapError(err, "error initializing settings")
	}

	l.Log.Debug().Msgf("settings: %+v", settings)

	group, ctx := errgroup.WithContext(ctx)

	// Start Graphql API Server
	group.Go(func() error {
		handler := NewGQLServer(gql.NewSchema(client), client)
		address := ":3003"
		mux := http.NewServeMux()
		server := &http.Server{Addr: address, Handler: mux}
		mux.Handle(
			"/",
			handler,
		)

		l.Log.Info().
			Str("addr", server.Addr).
			Msg("Starting Graphql API Server")
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

	// Depends on Nebula interface
	var onNebulaStart = func(ctrl *west.Control) {
		// Start Compass DNS
		group.Go(func() error {
			addr := "0.0.0.0:53"
			if privateDns {
				if disableTun {
					return errors.New("private dns cannot be used with tun disabled")
				}
				addr = net.JoinHostPort(settings.PortOverlayIP.ToIpAddr().String(), "53")
			}
			return dns.StartCompassDNSServer(ctx, addr, client, settings)
		})
	}

	// Start Nebula
	group.Go(func() error {
		cipher := config.CipherAes
		if settings.Cipher == "chachapoly" {
			cipher = config.CipherChaChaPoly
		}

		opts := &west.ServerOpts{
			OnStart: onNebulaStart,
			Config: &config.Config{
				Pki: config.Pki{
					Ca:   string(settings.CaCrt),
					Cert: string(settings.LighthouseCrt),
					Key:  string(settings.LighthouseKey),
				},
				Lighthouse: config.Lighthouse{
					Am_lighthouse: true,
				},
				Tun: config.Tun{
					Disabled: disableTun,
				},
				Listen: config.Listen{
					Host: "::",
					Port: 4242,
				},
				Preferred_ranges: config.DefaultPreferredRanges,
				Cipher:           cipher,
				Firewall: config.Firewall{
					Inbound: []config.FirewallRule{
						{
							Port:  config.PortAny,
							Proto: config.ProtoAny,
							Host:  config.HostAny,
						},
					},
					Outbound: []config.FirewallRule{
						{
							Port:  config.PortAny,
							Proto: config.ProtoAny,
							Host:  config.HostAny,
						},
					},
				},
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
