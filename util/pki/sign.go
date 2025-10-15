package pki

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/slackhq/nebula/cert"
	"golang.org/x/crypto/curve25519"
)

type SignCertOptions struct {
	CaCrt    []byte
	CaKey    []byte
	Name     string
	Ip       string
	Duration time.Duration
	Subnets  []string
	Groups   []string
}

type SignCertData struct {
	Cert []byte
	Key  []byte
}

var UnmarshalNebulaCertificateFromPEM = cert.UnmarshalNebulaCertificateFromPEM

type NebulaCertificate = cert.NebulaCertificate

const Curve = cert.Curve_CURVE25519

// TODO: Will need to update certs for new format
// https://github.com/slackhq/nebula/pull/1212/files
func SignCert(opts *SignCertOptions) (*SignCertData, error) {
	// Validate options
	if opts == nil {
		return nil, fmt.Errorf("sign cert options should not be nil")
	}
	if len(opts.CaCrt) == 0 {
		return nil, errors.New("ca crt is required")
	}
	if len(opts.CaKey) == 0 {
		return nil, errors.New("ca key is required")
	}
	if opts.Name == "" {
		return nil, errors.New("cert name is required")
	}
	if opts.Ip == "" {
		return nil, errors.New("cert ip is required")
	}

	// Cert Options
	certDuration := opts.Duration
	certName := opts.Name
	certIp := opts.Ip
	groups := opts.Groups

	// Build certificate
	caKey, _, err := cert.UnmarshalEd25519PrivateKey(opts.CaKey)
	if err != nil {
		return nil, fmt.Errorf("error while parsing ca-key: %s", err)
	}

	caCert, _, err := cert.UnmarshalNebulaCertificateFromPEM(opts.CaCrt)
	if err != nil {
		return nil, fmt.Errorf("error while parsing ca-crt: %s", err)
	}

	if err := caCert.VerifyPrivateKey(Curve, caKey); err != nil {
		return nil, fmt.Errorf("refusing to sign, root certificate does not match private key")
	}

	issuer, err := caCert.Sha256Sum()
	if err != nil {
		return nil, fmt.Errorf("error while getting -ca-crt fingerprint: %s", err)
	}

	if caCert.Expired(time.Now()) {
		return nil, fmt.Errorf("ca certificate is expired")
	}

	// if no duration is given, expire one second before the root expires
	if certDuration <= 0 {
		certDuration = time.Until(caCert.Details.NotAfter) - time.Second*1
	}

	ip, ipNet, err := net.ParseCIDR(certIp)
	if err != nil {
		return nil, fmt.Errorf("invalid ip definition: %s", err)
	}
	if ip.To4() == nil {
		return nil, fmt.Errorf("invalid ip definition: can only be ipv4, have %s", certIp)
	}
	ipNet.IP = ip

	subnets := make([]*net.IPNet, 0, len(opts.Subnets))
	if len(opts.Subnets) > 0 {
		for _, subnet := range opts.Subnets {
			subnet := strings.Trim(subnet, " ")
			if subnet != "" {
				_, s, err := net.ParseCIDR(subnet)
				if err != nil {
					return nil, fmt.Errorf("invalid subnet definition: %s", err)
				}
				if s.IP.To4() == nil {
					return nil, fmt.Errorf("invalid subnet definition: can only be ipv4, have %s", subnet)
				}
				subnets = append(subnets, s)
			}
		}
	}

	pub, rawPriv := x25519Keypair()

	// TODO: Look into using non ca.key to sign new certs
	// https://nebula.defined.net/docs/guides/sign-certificates-with-public-keys/
	//
	// var pub, rawPriv []byte
	// if *sf.inPubPath != "" {
	// 	rawPub, err := ioutil.ReadFile(*sf.inPubPath)
	// 	if err != nil {
	// 		return fmt.Errorf("error while reading in-pub: %s", err)
	// 	}
	// 	pub, _, err = cert.UnmarshalX25519PublicKey(rawPub)
	// 	if err != nil {
	// 		return fmt.Errorf("error while parsing in-pub: %s", err)
	// 	}
	// } else {
	// 	pub, rawPriv = x25519Keypair()
	// }

	nebulaCert := cert.NebulaCertificate{
		Details: cert.NebulaCertificateDetails{
			Name:      certName,
			Ips:       []*net.IPNet{ipNet},
			Groups:    groups,
			Subnets:   subnets,
			NotBefore: time.Now(),
			NotAfter:  time.Now().Add(certDuration),
			PublicKey: pub,
			IsCA:      false,
			Issuer:    issuer,
			Curve:     Curve,
		},
	}

	// Check constraints on parent / root CA
	if err := nebulaCert.CheckRootConstrains(caCert); err != nil {
		return nil, fmt.Errorf("refusing to sign, root certificate constraints violated: %s", err)
	}

	err = nebulaCert.Sign(Curve, caKey)
	if err != nil {
		return nil, fmt.Errorf("error while signing: %s", err)
	}

	signedKey := cert.MarshalX25519PrivateKey(rawPriv)
	signedCert, err := nebulaCert.MarshalToPEM()
	if err != nil {
		return nil, fmt.Errorf("error while marshalling certificate: %s", err)
	}

	return &SignCertData{
		Cert: signedCert,
		Key:  signedKey,
	}, nil
}

// Creates a new private and public key cert pair
func x25519Keypair() ([]byte, []byte) {
	privkey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, privkey); err != nil {
		panic(err)
	}

	pubkey, err := curve25519.X25519(privkey, curve25519.Basepoint)
	if err != nil {
		panic(err)
	}

	return pubkey, privkey
}
