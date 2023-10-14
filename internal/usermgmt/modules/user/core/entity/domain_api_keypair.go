package entity

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/crypt"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type APIKeyPair interface {
	PublicKey() field.String
	PrivateKey() field.String
}

type DomainAPIKeypair interface {
	APIKeyPair
	valueobj.HasUserID
	valueobj.HasOrganizationID
}

type APIKeyPairToDelegate struct {
	APIKeyPair
	valueobj.HasUserID
	valueobj.HasOrganizationID
}

type randomDomainAPIKeypair struct {
	publicKey  field.String
	privateKey field.String
}

func (e randomDomainAPIKeypair) PublicKey() field.String {
	return e.publicKey
}
func (e randomDomainAPIKeypair) PrivateKey() field.String {
	return e.privateKey
}

func NewRandomDomainAPIKeypair() (APIKeyPair, error) {
	keypair, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("rsa.GenerateKey err: %v", err)
	}

	privateKey := x509.MarshalPKCS1PrivateKey(keypair)
	publicKey := x509.MarshalPKCS1PublicKey(&keypair.PublicKey)

	return &randomDomainAPIKeypair{
		publicKey:  field.NewString(crypt.EncodeBase64(publicKey)),
		privateKey: field.NewString(crypt.EncodeBase64(privateKey)),
	}, nil
}
