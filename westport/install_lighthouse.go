package westport

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"

	"github.com/sprisa/west/util/info"
	"github.com/sprisa/west/util/ipconv"
	"github.com/sprisa/west/util/pki"
	"github.com/sprisa/west/westport/db"
	"github.com/sprisa/west/westport/db/ent"
	"github.com/sprisa/west/westport/localconfig"
	"github.com/sprisa/x/errutil"
	"github.com/urfave/cli/v3"
)

func validateInstallRequest(initialized bool, dataSource, lighthouseIP string) (bool, error) {
	if !initialized {
		if lighthouseIP != "" {
			return false, errors.New("--add-lighthouse-ip cannot be used before the datastore is initialized")
		}
		return false, nil
	}
	if lighthouseIP == "" {
		return false, errors.New("datastore is already initialized; --add-lighthouse-ip is required on a new west-port node")
	}
	typeName, err := db.DatastoreType(dataSource)
	if err != nil {
		return false, err
	}
	if typeName == db.DatastoreSQLite {
		return false, errors.New("--add-lighthouse-ip requires a PostgreSQL or MySQL datastore")
	}
	return true, nil
}

func installAdditionalLighthouse(
	ctx context.Context,
	c *cli.Command,
	client *ent.Client,
	settings *ent.Settings,
	dataSource string,
) error {
	lighthouseIP, err := netip.ParseAddr(c.String("add-lighthouse-ip"))
	if err != nil || !lighthouseIP.Is4() {
		return fmt.Errorf("invalid --add-lighthouse-ip %q", c.String("add-lighthouse-ip"))
	}
	if !settings.Cidr.Contains(lighthouseIP) {
		return fmt.Errorf("lighthouse IP `%s` must be within network cidr `%s`", lighthouseIP, settings.Cidr)
	}
	lighthouseIPValue, err := ipconv.FromIPAddr(lighthouseIP)
	if err != nil {
		return err
	}
	endpoint, err := resolvePortEndpoint(c.String("port-endpoint"))
	if err != nil {
		return err
	}
	hasHTTPS := settings.DomainZone != "" && len(settings.LetsencryptRegistration) > 0
	apiEndpoint, err := resolveApiEndpoint(endpoint, hasHTTPS)
	if err != nil {
		return err
	}
	certificate, err := signLighthouseCertificate(
		settings.CaCrt,
		settings.CaKey,
		lighthouseIP,
		settings.Cidr.Bits(),
	)
	if err != nil {
		return err
	}

	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	host, err := tx.Host.Create().SetIP(lighthouseIPValue).Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return fmt.Errorf("overlay IP %s is already in use", lighthouseIP)
		}
		return errutil.WrapErr(err, "reserve overlay IP")
	}
	if err := tx.Lighthouse.Create().
		SetIP(lighthouseIPValue).
		SetEndpoint(endpoint).
		SetAPIEndpoint(apiEndpoint).
		SetCertificate(certificate.Cert).
		SetKey(certificate.Key).
		SetHostID(host.ID).
		Exec(ctx); err != nil {
		return errutil.WrapErr(err, "error saving lighthouse")
	}
	return commitLocalInstallation(tx, localconfig.Config{
		Datastore:    dataSource,
		LighthouseIP: lighthouseIP.String(),
	})
}

func commitLocalInstallation(tx *ent.Tx, cfg localconfig.Config) error {
	if err := tx.Commit(); err != nil {
		return err
	}
	return localconfig.Save(cfg)
}

func resolvePortEndpoint(endpoint string) (string, error) {
	if endpoint == "" {
		publicIP, err := info.GetPublicIP()
		if err != nil {
			return "", errutil.WrapErr(err, "detect public IP")
		}
		if publicIP == nil {
			return "", errors.New("public IP service returned an invalid address")
		}
		return net.JoinHostPort(publicIP.String(), "4242"), nil
	}
	host, portValue, err := net.SplitHostPort(endpoint)
	if err != nil || host == "" {
		return "", fmt.Errorf("invalid --port-endpoint %q; expected host:port", endpoint)
	}
	portNumber, err := strconv.ParseUint(portValue, 10, 16)
	if err != nil || portNumber == 0 {
		return "", fmt.Errorf("invalid --port-endpoint port %q", portValue)
	}
	return net.JoinHostPort(host, portValue), nil
}

func resolveApiEndpoint(nebulaEndpoint string, hasHTTPS bool) (string, error) {
	host, _, err := net.SplitHostPort(nebulaEndpoint)
	if err != nil {
		return "", fmt.Errorf("parse nebula endpoint for API address: %w", err)
	}
	port := "80"
	if hasHTTPS {
		port = "443"
	}
	return net.JoinHostPort(host, port), nil
}

func signLighthouseCertificate(ca, caKey []byte, ip netip.Addr, bits int) (*pki.SignCertData, error) {
	name := "west-port-" + strings.ReplaceAll(ip.String(), ".", "-")
	certificate, err := pki.SignCert(&pki.SignCertOptions{
		CaCrt: ca,
		CaKey: caKey,
		Name:  name,
		Ip:    netip.PrefixFrom(ip, bits).String(),
	})
	if err != nil {
		return nil, errutil.WrapErr(err, "error generating west-port certificate")
	}
	return certificate, nil
}
