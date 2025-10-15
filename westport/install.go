package westport

import (
	"context"
	"errors"
	"os"

	"github.com/sprisa/west/util/errutil"
	"github.com/sprisa/west/util/ipconv"
	l "github.com/sprisa/west/util/log"
	"github.com/sprisa/west/util/pki"
	"github.com/sprisa/west/westport/db"
	"github.com/sprisa/west/westport/db/ent"
	"github.com/sprisa/west/westport/db/helpers"
	"github.com/sprisa/west/westport/db/migrate"
	"github.com/urfave/cli/v3"
)

var InstallCommand = &cli.Command{
	Name:      "install",
	Usage:     "Install west port",
	UsageText: "west port install",
	Flags: []cli.Flag{
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
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		caPath := c.String("ca-crt")
		caKeyPath := c.String("ca-key")
		ca, err := os.ReadFile(caPath)
		if err != nil {
			return errutil.WrapError(err, "error reading ca at `%s`", caPath)
		}
		caKey, err := os.ReadFile(caKeyPath)
		if err != nil {
			return errutil.WrapError(err, "error reading ca-key at `%s`", caPath)
		}
		cidr := c.String("cidr")

		client, err := db.OpenDB()
		if err != nil {
			return errutil.WrapError(err, "error opening db")
		}
		defer client.Close()
		err = migrate.MigrateClient(ctx, client)
		if err != nil {
			return errutil.WrapError(err, "error migrating db")
		}

		_, err = client.Settings.Query().First(ctx)
		if ent.IsNotFound(err) == false {
			return errors.New("west port already installed with database present.")
		}

		lhCert, err := pki.SignCert(&pki.SignCertOptions{
			CaCrt: ca,
			CaKey: caKey,
			Name:  "west-port-1",
			Ip:    cidr,
		})
		if err != nil {
			return errutil.WrapError(err, "error generating west-port cert")
		}

		ipCidr, err := helpers.NewIpCidr(cidr)
		if err != nil {
			return errutil.WrapError(err, "error parsing cidr")
		}
		overlayIp, err := ipconv.FromIPAddr(ipCidr.Addr())
		if err != nil {
			return err
		}

		l.Log.Info().Msg("Create a encryption a password")
		err = promptEncryptionPassword()
		if err != nil {
			return err
		}

		err = client.Settings.Create().
			SetCaCrt(ca).
			SetCaKey(caKey).
			SetLighthouseCrt(lhCert.Cert).
			SetLighthouseKey(lhCert.Key).
			SetCidr(ipCidr).
			SetPortOverlayIP(overlayIp).
			Exec(ctx)
		if err != nil {
			return errutil.WrapError(err, "error saving settings")
		}

		l.Log.Info().Msg("Done! Use `west port start` to run")
		return nil
	},
}
