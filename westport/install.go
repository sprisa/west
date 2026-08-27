package westport

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/sprisa/west/util/ipconv"
	"github.com/sprisa/west/util/pki"
	"github.com/sprisa/west/westport/acme"
	"github.com/sprisa/west/westport/db"
	"github.com/sprisa/west/westport/db/ent"
	"github.com/sprisa/west/westport/db/helpers"
	"github.com/sprisa/west/westport/db/migrate"
	"github.com/sprisa/west/westport/localconfig"
	"github.com/sprisa/x/errutil"
	l "github.com/sprisa/x/log"
	"github.com/urfave/cli/v3"
)

var InstallCommand = &cli.Command{
	Name:      "install",
	Usage:     "Install west port",
	UsageText: "west port install --datastore <connection-string>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "datastore",
			Usage:    "Datastore connection: sqlite, sqlite://<path>, postgres://, or mysql://",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "ca-crt",
			Value: "ca.crt",
			Usage: "Path to ca cert",
		},
		&cli.StringFlag{
			Name:  "ca-key",
			Value: "ca.key",
			Usage: "Path to ca key",
		},
		&cli.StringFlag{
			Name:  "cidr",
			Value: "10.10.10.1/24",
			Usage: "Network IP cidr range",
		},
		&cli.StringFlag{
			Name:  "domain-zone",
			Usage: "Domain zone to control",
		},
		&cli.StringFlag{
			Name:  "letsencrypt-email",
			Usage: "Email for letsencrypt registration. Required for automated HTTPS certificates",
		},
		&cli.BoolFlag{
			Name:  "letsencrypt-accept-tos",
			Usage: "Accept the letsencrypt terms of service. Required for automated HTTPS certificates",
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		if err := readEncryptionPassword(); err != nil {
			return err
		}

		exists, err := localconfig.Exists()
		if err != nil {
			return errutil.WrapErr(err, "check local west-port installation")
		}
		if exists {
			cfg, err := localconfig.Load()
			if err != nil {
				return fmt.Errorf("config at %s is unreadable: %w; remove it to reinstall", localconfig.FilePath, err)
			}
			l.Log.Info().Str("datastore", cfg.Datastore).
				Msg("West port is already installed on this node")
			return nil
		}

		dataSource := c.String("datastore")
		client, err := db.OpenDB(ctx, dataSource)
		if err != nil {
			return errutil.WrapErr(err, "error opening db")
		}
		defer client.Close()
		if err := migrate.MigrateClient(ctx, client); err != nil {
			return errutil.WrapErr(err, "error migrating db")
		}

		_, err = client.Settings.Query().First(ctx)
		if ent.IsNotFound(err) == false {
			return errors.New("west port already installed with database present.")
		}

		caPath := c.String("ca-crt")
		caKeyPath := c.String("ca-key")
		ca, err := os.ReadFile(caPath)
		if err != nil {
			return errutil.WrapErr(err, "error reading ca at `%s`", caPath)
		}
		caKey, err := os.ReadFile(caKeyPath)
		if err != nil {
			return errutil.WrapErr(err, "error reading ca-key at `%s`", caKeyPath)
		}
		cidr := c.String("cidr")
		domainZone := strings.ToLower(c.String("domain-zone"))
		letsencryptEmail := c.String("letsencrypt-email")
		if letsencryptEmail != "" && !c.Bool("letsencrypt-accept-tos") {
			return errors.New("required to accept Let's Encrypt terms of service (--letsencrypt-accept-tos)")
		}
		if letsencryptEmail != "" && domainZone == "" {
			return errors.New("domain zone must be specified to use Let's Encrypt certificates (--domain-zone)")
		}

		lhCert, err := pki.SignCert(&pki.SignCertOptions{
			CaCrt: ca,
			CaKey: caKey,
			Name:  "west-port-1",
			Ip:    cidr,
		})
		if err != nil {
			return errutil.WrapErr(err, "error generating west-port cert")
		}

		ipCidr, err := helpers.NewIpCidr(cidr)
		if err != nil {
			return errutil.WrapErr(err, "error parsing cidr")
		}
		overlayIp, err := ipconv.FromIPAddr(ipCidr.Addr())
		if err != nil {
			return err
		}

		var acmeRegistration []byte
		if letsencryptEmail != "" {
			acmeUser, err := acme.NewUserRegistration(letsencryptEmail)
			if err != nil {
				return errutil.WrapErr(err, "error creating new lets encrypt user")
			}
			acmeRegistration, err = acmeUser.ToBytes()
			if err != nil {
				return errutil.WrapErr(err, "error serializing acme registration")
			}
			l.Log.Info().Str("email", letsencryptEmail).Msg("Registered with Let's Encrypt")
		}

		tx, err := client.Tx(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		err = tx.Settings.Create().
			SetCaCrt(ca).
			SetCaKey(caKey).
			SetLighthouseCrt(lhCert.Cert).
			SetLighthouseKey(lhCert.Key).
			SetCidr(ipCidr).
			SetPortOverlayIP(overlayIp).
			SetDomainZone(domainZone).
			SetLetsencryptRegistration(acmeRegistration).
			Exec(ctx)
		if err != nil {
			return errutil.WrapErr(err, "error saving settings")
		}
		err = tx.Host.Create().
			SetIP(overlayIp).
			Exec(ctx)
		if err != nil {
			return errutil.WrapErr(err, "error reserving lighthouse IP")
		}
		err = tx.Commit()
		if err != nil {
			return err
		}

		if err := localconfig.Save(localconfig.Config{
			Datastore:    dataSource,
			LighthouseIP: ipCidr.Addr().String(),
		}); err != nil {
			return errutil.WrapErr(err, "save local west-port config")
		}

		l.Log.Info().Msg("Done! Use `west port start` to run")
		return nil
	},
}
