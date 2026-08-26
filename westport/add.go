package westport

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sprisa/west/util/auth"
	"github.com/sprisa/west/util/ipconv"
	"github.com/sprisa/west/westport/db"
	"github.com/sprisa/west/westport/db/ent"
	"github.com/sprisa/west/westport/db/ent/lighthouse"
	"github.com/sprisa/west/westport/db/helpers"
	"github.com/sprisa/west/westport/db/migrate"
	"github.com/sprisa/west/westport/localconfig"
	"github.com/sprisa/x/errutil"
	"github.com/urfave/cli/v3"
)

var AddCommand = &cli.Command{
	Name:      "add",
	Usage:     "Register a new west device",
	UsageText: "west port add",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "name",
			Required: true,
			Usage:    "Device name. Must be unique.",
		},
		&cli.StringFlag{
			Name:     "ip",
			Usage:    "IP for device. Must be unique within existing cidr.",
			Required: true,
			Validator: func(s string) error {
				_, err := ipconv.ParseToIP(s)
				return err
			},
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		name := c.String("name")
		ipStr := c.String("ip")
		ip, err := netip.ParseAddr(ipStr)
		if err != nil {
			return errutil.WrapErr(err, "error parsing ip `%s`", ipStr)
		}

		err = readEncryptionPassword()
		if err != nil {
			return err
		}
		localCfg, err := localconfig.Load()
		if err != nil {
			return errutil.WrapErr(err, "load local west-port config")
		}
		client, err := db.OpenDB(ctx, localCfg.Datastore)
		if err != nil {
			return errutil.WrapErr(err, "error opening db")
		}
		defer client.Close()
		err = migrate.MigrateClient(ctx, client)
		if err != nil {
			return errutil.WrapErr(err, "error migrating db")
		}

		settings, err := client.Settings.Query().Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return errors.New("error finding settings. Trying installing first.")
			}
			return errutil.WrapErr(err, "error initializing settings")
		}

		if settings.Cidr.Contains(ip) == false {
			return fmt.Errorf("ip `%s` must be within network cidr `%s`", ip, settings.Cidr)
		}

		nebulaIp := netip.PrefixFrom(ip, settings.Cidr.Bits())

		ports, err := client.Lighthouse.Query().Order(ent.Asc(lighthouse.FieldIP)).All(ctx)
		if err != nil {
			return errutil.WrapErr(err, "error loading west ports")
		}
		if len(ports) == 0 {
			return errors.New("no west ports are installed")
		}

		var endpoint url.URL
		if settings.DomainZone != "" && len(settings.LetsencryptRegistration) > 0 {
			endpoint = url.URL{
				Scheme: "https",
				Host:   settings.DomainZone,
				Path:   "api",
			}
		} else {
			host := settings.DomainZone
			if host == "" {
				host = ports[0].APIEndpoint
			}
			endpoint = url.URL{
				Scheme: "http",
				Host:   host,
				Path:   "api",
			}
		}
		endpointAddresses := make([]string, 0, len(ports))
		for _, westPort := range ports {
			endpointAddresses = append(endpointAddresses, westPort.APIEndpoint)
		}

		claims := &auth.TokenClaims{
			Endpoint:          endpoint.String(),
			EndpointAddresses: endpointAddresses,
			IP:                nebulaIp.String(),
			Ca:                string(settings.CaCrt),
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(
					// 1 year
					time.Now().Add(time.Hour * 8760),
				),
			},
		}

		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
			SignedString(helpers.EncryptionKey[:])
		if err != nil {
			return errutil.WrapErr(err, "error creating token")
		}

		ipInt, err := ipconv.FromIPAddr(nebulaIp.Addr())
		if err != nil {
			return errutil.WrapErr(err, "error converting ip")
		}
		tx, err := client.Tx(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		host, err := tx.Host.Create().SetIP(ipInt).Save(ctx)
		if err != nil {
			if ent.IsConstraintError(err) {
				return fmt.Errorf("overlay IP %s is already in use", ip)
			}
			return errutil.WrapErr(err, "reserve overlay IP")
		}
		if _, err := tx.Device.Create().
			SetName(name).
			SetIP(ipInt).
			SetToken(helpers.EncryptedBytes(token)).
			SetHostID(host.ID).
			Save(ctx); err != nil {
			return errutil.WrapErr(err, "error saving device")
		}
		if err := tx.Commit(); err != nil {
			return err
		}

		println(token)
		return nil
	},
}
