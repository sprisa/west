package dns

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	l "github.com/sprisa/west/util/log"
)

type User struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *User) GetEmail() string {
	return u.Email
}

func (u *User) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

type CertManager struct {
	domain       string
	email        string
	certDir      string
	httpProvider *HTTPProvider
	dnsProvider  *ACMEProvider
}

func NewCertManager(domain, email, certDir string, httpProvider *HTTPProvider, dnsProvider *ACMEProvider) *CertManager {
	return &CertManager{
		domain:       domain,
		email:        email,
		certDir:      certDir,
		httpProvider: httpProvider,
		dnsProvider:  dnsProvider,
	}
}

func (cm *CertManager) GetOrObtainCertificate() (*tls.Certificate, error) {
	// Try to load existing certificate
	cert, err := cm.loadCertificate()
	if err == nil {
		l.Log.Info().Str("domain", cm.domain).Msg("Loaded existing certificate")
		return cert, nil
	}

	l.Log.Info().Str("domain", cm.domain).Msg("Obtaining new certificate")

	// Obtain new certificate
	if err := cm.obtainCertificate(); err != nil {
		return nil, fmt.Errorf("failed to obtain certificate: %w", err)
	}

	// Load the newly obtained certificate
	cert, err = cm.loadCertificate()
	if err != nil {
		return nil, fmt.Errorf("failed to load new certificate: %w", err)
	}

	return cert, nil
}

func (cm *CertManager) obtainCertificate() error {
	// Create or load user account
	// TODO: Create user on install
	user, err := cm.getOrCreateUser()
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Create lego config
	config := lego.NewConfig(user)
	config.CADirURL = lego.LEDirectoryProduction // Use lego.LEDirectoryProduction for production
	// config.CADirURL = lego.LEDirectoryStaging // Use lego.LEDirectoryProduction for production
	config.Certificate.KeyType = certcrypto.RSA2048

	// Create lego client
	client, err := lego.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create lego client: %w", err)
	}

	// Set HTTP-01 challenge provider
	if cm.httpProvider != nil {
		err = client.Challenge.SetHTTP01Provider(cm.httpProvider)
		if err != nil {
			return fmt.Errorf("failed to set HTTP provider: %w", err)
		}
	}

	// Configure DNS-01 challenge
	if cm.dnsProvider != nil {
		err = client.Challenge.SetDNS01Provider(
			cm.dnsProvider,
			dns01.DisableAuthoritativeNssPropagationRequirement(),
		)
		if err != nil {
			return fmt.Errorf("failed to set DNS provider: %w", err)
		}
	}

	// Register account if needed
	if user.Registration == nil {
		reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return fmt.Errorf("failed to register: %w", err)
		}
		user.Registration = reg
		if err := cm.saveUser(user); err != nil {
			return fmt.Errorf("failed to save user: %w", err)
		}
	}

	// Obtain certificate
	request := certificate.ObtainRequest{
		Domains: []string{cm.domain},
		Bundle:  true,
	}

	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return fmt.Errorf("failed to obtain certificate: %w", err)
	}

	// Save certificate and key
	if err := os.MkdirAll(cm.certDir, 0700); err != nil {
		return fmt.Errorf("failed to create cert directory: %w", err)
	}

	certPath := filepath.Join(cm.certDir, cm.domain+".crt")
	keyPath := filepath.Join(cm.certDir, cm.domain+".key")

	if err := os.WriteFile(certPath, certificates.Certificate, 0600); err != nil {
		return fmt.Errorf("failed to save certificate: %w", err)
	}

	if err := os.WriteFile(keyPath, certificates.PrivateKey, 0600); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	l.Log.Info().Str("domain", cm.domain).Msg("Certificate obtained successfully")
	return nil
}

func (cm *CertManager) loadCertificate() (*tls.Certificate, error) {
	certPath := filepath.Join(cm.certDir, cm.domain+".crt")
	keyPath := filepath.Join(cm.certDir, cm.domain+".key")

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	// Parse certificate to check expiry
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, err
	}

	// Check if certificate needs renewal (30 days before expiry)
	if x509Cert.NotAfter.Unix()-time.Now().Unix() < 30*24*3600 {
		return nil, errors.New("certificate expiring soon")
	}

	return &cert, nil
}

func (cm *CertManager) getOrCreateUser() (*User, error) {
	userPath := filepath.Join(cm.certDir, "user.json")
	// keyPath := filepath.Join(cm.certDir, "user.key")

	// Try to load existing user
	if _, err := os.Stat(userPath); err == nil {
		return cm.loadUser()
	}

	// Create new user
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	user := &User{
		Email: cm.email,
		key:   privateKey,
	}

	if err := cm.saveUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (cm *CertManager) saveUser(user *User) error {
	userPath := filepath.Join(cm.certDir, "user.json")
	keyPath := filepath.Join(cm.certDir, "user.key")

	if err := os.MkdirAll(cm.certDir, 0700); err != nil {
		return err
	}

	// Save user data
	userData, err := json.Marshal(user)
	if err != nil {
		return err
	}
	if err := os.WriteFile(userPath, userData, 0600); err != nil {
		return err
	}

	// Save private key
	keyBytes, err := x509.MarshalECPrivateKey(user.key.(*ecdsa.PrivateKey))
	if err != nil {
		return err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return err
	}

	return nil
}

func (cm *CertManager) loadUser() (*User, error) {
	userPath := filepath.Join(cm.certDir, "user.json")
	keyPath := filepath.Join(cm.certDir, "user.key")

	// Load user data
	userData, err := os.ReadFile(userPath)
	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal(userData, &user); err != nil {
		return nil, err
	}

	// Load private key
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	user.key = privateKey
	return &user, nil
}
