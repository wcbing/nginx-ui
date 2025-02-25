package model

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"math/big"
	"net/http"
)

type PrivateKey struct {
	X, Y *big.Int
	D    *big.Int
}

type AcmeUser struct {
	Model
	Name         string                `json:"name"`
	Email        string                `json:"email"`
	CADir        string                `json:"ca_dir"`
	Registration registration.Resource `json:"registration" gorm:"serializer:json"`
	Key          PrivateKey            `json:"-" gorm:"serializer:json"`
}

func (u *AcmeUser) GetEmail() string {
	return u.Email
}

func (u *AcmeUser) GetRegistration() *registration.Resource {
	return &u.Registration
}

func (u *AcmeUser) GetPrivateKey() crypto.PrivateKey {
	return &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     u.Key.X,
			Y:     u.Key.Y,
		},
		D: u.Key.D,
	}
}
func (u *AcmeUser) Register() error {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	u.Key = PrivateKey{
		X: privateKey.PublicKey.X,
		Y: privateKey.PublicKey.Y,
		D: privateKey.D,
	}

	config := lego.NewConfig(u)
	config.CADirURL = u.CADir
	u.Registration = registration.Resource{}

	// Skip TLS check
	if config.HTTPClient != nil {
		config.HTTPClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	client, err := lego.NewClient(config)
	if err != nil {
		return err
	}

	// New users will need to register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return err
	}

	u.Registration = *reg

	return nil
}
