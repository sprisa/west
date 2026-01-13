package westport

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/sprisa/west/util/ipconv"
	"github.com/sprisa/west/util/pki"
	"github.com/sprisa/west/westport/acme"
	"github.com/sprisa/west/westport/db"
	"github.com/sprisa/west/westport/db/ent"
	"github.com/sprisa/west/westport/db/helpers"
	"github.com/sprisa/west/westport/db/migrate"
	"github.com/sprisa/x/errutil"
	l "github.com/sprisa/x/log"
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
		caPath := c.String("ca-crt")
		caKeyPath := c.String("ca-key")
		ca, err := os.ReadFile(caPath)
		if err != nil {
			return errutil.WrapErr(err, "error reading ca at `%s`", caPath)
		}
		caKey, err := os.ReadFile(caKeyPath)
		if err != nil {
			return errutil.WrapErr(err, "error reading ca-key at `%s`", caPath)
		}
		cidr := c.String("cidr")
		domainZone := strings.ToLower(c.String("domain-zone"))
		letsencryptEmail := c.String("letsencrypt-email")
		letsencryptTOSAccepted := c.Bool("letsencrypt-accept-tos")
		if letsencryptEmail != "" && letsencryptTOSAccepted == false {
			return errors.New("Required to accept Let's Encrypt terms of service (--letsencrypt-accept-tos)")
		}
		if letsencryptEmail != "" && domainZone == "" {
			return errors.New("Domain zone must be specified in order to use Let's Encrypt certificates (--domain-zone)")
		}

		client, err := db.OpenDB()
		if err != nil {
			return errutil.WrapErr(err, "error opening db")
		}
		defer client.Close()
		err = migrate.MigrateClient(ctx, client)
		if err != nil {
			return errutil.WrapErr(err, "error migrating db")
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

			l.Log.Info().
				Str("email", letsencryptEmail).
				Msg("Registered with Let's Encrypt")
		}

		l.Log.Info().Msg("Create a encryption a password")
		err = readEncryptionPassword()
		if err != nil {
			return err
		}

		err = client.Settings.Create().
			SetCaCrt(ca).
			SetCaKey(caKey).
			// TODO: Store info in a device so it get's all the
			// DNS and uniqueness built in.
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

		l.Log.Info().Msg("Done! Use `west port start` to run")
		// TODO: Show extra steps on snap mode
		// sudo snap connect west:network-control
		return nil
	},
}
