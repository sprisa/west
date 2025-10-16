package acme

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/sprisa/west/util/errutil"
	l "github.com/sprisa/west/util/log"
	"github.com/sprisa/west/westport/db/ent"
)

func GetCertificate(
	ctx context.Context,
	settings *ent.Settings,
	httpProvider *HTTPProvider,
	dnsProvider *DNSProvider,
) (*tls.Certificate, error) {
	// Check for existing key
	if settings.TLSCert != nil && len(*settings.TLSCert) > 0 &&
		settings.TLSCertKey != nil && len(*settings.TLSCertKey) > 0 {
		tlsCert, tlsCertKey := *settings.TLSCert, *settings.TLSCertKey

		cert, err := tls.X509KeyPair(tlsCert, tlsCertKey)
		if err != nil {
			return nil, errutil.WrapError(err, "error parsing cert")
		}

		// Parse certificate to check expiry
		x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			return nil, errutil.WrapError(err, "error parsing x509 cert")
		}

		// Check if certificate needs renewal (30 days before expiry)
		if x509Cert.NotAfter.Unix()-time.Now().Unix() < 30*24*3600 {
			l.Log.Warn().Msg("Certificate expiring in < 30 days. Renewing now.")
		} else {
			l.Log.Info().Msg("Using cached TLS certificate")
			return &cert, nil
		}
	}

	domain := settings.DomainZone
	user, err := UserRegistrationFromBytes(settings.LetsencryptRegistration)
	if err != nil {
		return nil, err
	}

	config := lego.NewConfig(user)
	// config.CADirURL = lego.LEDirectoryStaging
	client, err := lego.NewClient(config)
	if err != nil {
		return nil, errutil.WrapError(err, "error creating acme client")
	}

	// Set HTTP-01 challenge provider
	if httpProvider != nil {
		err = client.Challenge.SetHTTP01Provider(httpProvider)
		if err != nil {
			return nil, errutil.WrapError(err, "failed to set HTTP provider")
		}
	}

	// Configure DNS-01 challenge
	if dnsProvider != nil {
		err = client.Challenge.SetDNS01Provider(
			dnsProvider,
			dns01.DisableAuthoritativeNssPropagationRequirement(),
		)
		if err != nil {
			return nil, errutil.WrapError(err, "failed to set DNS provider")
		}
	}

	// Obtain certificate
	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}

	certs, err := client.Certificate.Obtain(request)
	if err != nil {
		return nil, errutil.WrapError(err, "failed to obtain certificate")
	}

	cert, err := tls.X509KeyPair(certs.Certificate, certs.PrivateKey)
	if err != nil {
		return nil, errutil.WrapError(err, "error parsing cert")
	}

	err = settings.Update().
		SetTLSCert(certs.Certificate).
		SetTLSCertKey(certs.PrivateKey).
		Exec(ctx)
	if err != nil {
		return nil, errutil.WrapError(err, "error saving tls cert")
	}

	return &cert, nil
}
