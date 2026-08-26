package westport

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/sprisa/west/util/ipconv"
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
			Name:  "add-lighthouse-ip",
			Usage: "Add this lighthouse overlay IP to an existing external datastore",
		},
		&cli.StringFlag{
			Name:  "port-endpoint",
			Usage: "Public Nebula endpoint for this lighthouse (defaults to the detected public IP on port 4242)",
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
			Usage: "Network IP cidr range; its address is the first lighthouse IP",
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
			l.Log.Info().Str("datastore", cfg.Datastore).Str("lighthouse", cfg.LighthouseIP).
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

		settings, err := client.Settings.Query().Only(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return errutil.WrapErr(err, "error reading west-port settings")
		}
		addLighthouse, err := validateInstallRequest(
			err == nil,
			dataSource,
			c.String("add-lighthouse-ip"),
		)
		if err != nil {
			return err
		}
		if addLighthouse {
			return installAdditionalLighthouse(ctx, c, client, settings, dataSource)
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

		ipCidr, err := helpers.NewIpCidr(cidr)
		if err != nil {
			return errutil.WrapErr(err, "error parsing cidr")
		}
		if !ipCidr.Addr().Is4() {
			return errors.New("west currently requires an IPv4 network cidr")
		}
		endpoint, err := resolvePortEndpoint(c.String("port-endpoint"))
		if err != nil {
			return err
		}
		lighthouseCert, err := signLighthouseCertificate(ca, caKey, ipCidr.Addr(), ipCidr.Bits())
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

		hasHTTPS := len(acmeRegistration) > 0
		apiEndpoint, err := resolveApiEndpoint(endpoint, hasHTTPS)
		if err != nil {
			return err
		}
		lighthouseIP, err := ipconv.FromIPAddr(ipCidr.Addr())
		if err != nil {
			return err
		}
		tx, err := client.Tx(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		if err := tx.Settings.Create().
			SetCaCrt(ca).
			SetCaKey(caKey).
			SetCidr(ipCidr).
			SetDomainZone(domainZone).
			SetLetsencryptRegistration(acmeRegistration).
			Exec(ctx); err != nil {
			if ent.IsConstraintError(err) {
				return errors.New("datastore is already initialized; use --add-lighthouse-ip on a new node")
			}
			return errutil.WrapErr(err, "error saving settings")
		}
		host, err := tx.Host.Create().SetIP(lighthouseIP).Save(ctx)
		if err != nil {
			return errutil.WrapErr(err, "reserve overlay IP")
		}
		if err := tx.Lighthouse.Create().
			SetIP(lighthouseIP).
			SetEndpoint(endpoint).
			SetAPIEndpoint(apiEndpoint).
			SetCertificate(lighthouseCert.Cert).
			SetKey(lighthouseCert.Key).
			SetHostID(host.ID).
			Exec(ctx); err != nil {
			return errutil.WrapErr(err, "error saving first lighthouse")
		}
		if err := commitLocalInstallation(tx, localconfig.Config{
			Datastore:    dataSource,
			LighthouseIP: ipCidr.Addr().String(),
		}); err != nil {
			return err
		}

		l.Log.Info().Msg("Done! Use `west port start` to run")
		return nil
	},
}
