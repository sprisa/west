package acme

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/gob"

	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/sprisa/west/util/errutil"
)

type UserRegistration struct {
	Email        string
	Key          UserRegistrationKey
	Registration registration.Resource
}

func (s *UserRegistration) ToBytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(s)
	return buf.Bytes(), err
}

func (s *UserRegistration) GetEmail() string {
	return s.Email
}

func (s *UserRegistration) GetPrivateKey() crypto.PrivateKey {
	return s.Key.key
}

func (s *UserRegistration) GetRegistration() *registration.Resource {
	return &s.Registration
}

type UserRegistrationKey struct {
	key *ecdsa.PrivateKey
}

func (s *UserRegistrationKey) GobEncode() ([]byte, error) {
	return x509.MarshalECPrivateKey(s.key)
}
func (s *UserRegistrationKey) GobDecode(data []byte) error {
	key, err := x509.ParseECPrivateKey(data)
	if err != nil {
		return err
	}
	s.key = key
	return nil
}

func NewUserRegistration(email string) (*UserRegistration, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, errutil.WrapError(err, "failed to generate private key")
	}

	user := &UserRegistration{
		Email: email,
		Key: UserRegistrationKey{
			key: key,
		},
	}

	config := lego.NewConfig(user)
	client, err := lego.NewClient(config)
	if err != nil {
		return nil, errutil.WrapError(err, "error creating acme client")
	}

	reg, err := client.Registration.Register(registration.RegisterOptions{
		TermsOfServiceAgreed: true,
	})
	if err != nil {
		return nil, errutil.WrapError(err, "error registering acme")
	}
	user.Registration = *reg

	return user, nil
}

func UserRegistrationFromBytes(b []byte) (*UserRegistration, error) {
	dec := gob.NewDecoder(bytes.NewReader(b))
	u := &UserRegistration{}
	err := dec.Decode(u)
	return u, err
}
