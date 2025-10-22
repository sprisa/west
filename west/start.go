package west

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sprisa/west"
	"github.com/sprisa/west/config"
	"github.com/sprisa/west/util/auth"
	"github.com/sprisa/west/util/errutil"
	"github.com/sprisa/west/util/ioutil"
	l "github.com/sprisa/west/util/log"
	"github.com/sprisa/west/west/gql"
	"github.com/urfave/cli/v3"
)

var StartCommand = &cli.Command{
	Name:      "start",
	Usage:     "Start west device",
	UsageText: "west start [jwt_token]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			// TODO: Would be better if this was set in a different command
			// then stored in a secure enclave for start.
			Name:    "token",
			Aliases: []string{"t"},
			Usage:   "API token. Can be passed via flag or stdin.",
		},
		&cli.BoolFlag{
			Name:  "disable-tun",
			Usage: "Disabled TUN network binding. Useful for rootless testing",
		},
		&cli.IntFlag{
			Name:  "port",
			Usage: "Port to use for Nebula. Defaults to a random free port.",
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		port := c.Int("port")
		disableTun := c.Bool("disable-tun")
		token := c.String("token")
		// Read via stdin if available
		if token == "" && ioutil.StdinAvailable() {
			tokenBytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
			token = string(bytes.TrimSpace(tokenBytes))
		}
		if token == "" {
			return errors.New("No token supplied. Pass via flag or stdin.")
		}

		endpoint := os.Getenv("WEST_ENDPOINT")
		url, err := url.Parse(endpoint)
		if err != nil {
			return errutil.WrapError(err, "error parsing endpoint")
		}

		parser := jwt.NewParser()
		claims := &auth.TokenClaims{}
		_, _, err = parser.ParseUnverified(token, claims)
		if err != nil {
			return fmt.Errorf("error parsing token: %w", err)
		}
		if claims.ExpiresAt.Before(time.Now()) {
			return errors.New("token expired")
		}
		if endpoint == "" {
			endpoint = claims.Endpoint
		}

		l.Log.Debug().Msgf("claims: %+v", claims)

		client := graphql.NewClient(endpoint, http.DefaultClient)
		data, err := gql.ProvisionDevice(ctx, client, gql.ProvisionDeviceInput{
			Token: token,
		})
		if err != nil {
			return errutil.WrapError(err, "error provisioning device")
		}
		dvc := data.GetProvision_device()
		l.Log.Info().
			Str("name", dvc.Name).
			Str("ip", claims.IP).
			Msg("Received provisioning")

		if port == 0 {
			port, err = getFreePort()
			if err != nil {
				return errutil.WrapError(err, "erroring find free port")
			}
		}

		srv, err := west.NewServer(&west.ServerOpts{
			Config: &config.Config{
				Pki: config.Pki{
					Ca:   dvc.Ca,
					Cert: dvc.Cert,
					Key:  dvc.Key,
				},
				StaticHostMap: config.StaticHostMap{
					claims.PortIP: []string{
						net.JoinHostPort(url.Hostname(), "4242"),
					},
				},
				Lighthouse: config.Lighthouse{
					Hosts: []string{
						claims.PortIP,
					},
				},
				Tun: config.Tun{
					Disabled: disableTun,
				},
				Listen: config.Listen{
					Host: "::",
					Port: port,
				},
				PreferredRanges: config.DefaultPreferredRanges,
				Cipher:          config.Cipher(dvc.NetworkCipher),
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
		})
		if err != nil {
			return err
		}

		return srv.Listen(ctx)
	},
}

func getFreePort() (port int, err error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
