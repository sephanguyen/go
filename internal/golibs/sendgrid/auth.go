package sendgrid

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"fmt"
	"math/big"
	"net/http"
)

const (
	// VerificationHTTPHeader is the signature verification http header name for the signature being sent
	VerificationHTTPHeader = "X-Twilio-Email-Event-Webhook-Signature"
	// TimestampHTTPHeader is the timestamp http header name for timestamp
	TimestampHTTPHeader = "X-Twilio-Email-Event-Webhook-Timestamp"
)

// RS represents the ECDSA signature
type RS struct {
	R *big.Int
	S *big.Int
}

func (sgc *sgClientImpl) AuthenticateHTTPRequest(header http.Header, payload []byte) (bool, error) {
	sgn := header.Get(VerificationHTTPHeader)
	timeStamp := header.Get(TimestampHTTPHeader)

	if sgn == "" || timeStamp == "" {
		return false, fmt.Errorf("invalid request")
	}

	pKey, err := ConvertPublicKeyBase64ToECDSA(sgc.PublicKey)
	if err != nil {
		return false, fmt.Errorf("failed ConvertPublicKeyBase64ToECDSA: %+v", err)
	}
	passed, err := VerifySignature(pKey, payload, sgn, timeStamp)
	if err != nil {
		return false, fmt.Errorf("failed VerifySignature; %+v", err)
	}

	return passed, nil
}

// ConvertPublicKeyBase64ToECDSA takes a base64 ECDSA public key and converts it into the ECDSA Public Key type
func ConvertPublicKeyBase64ToECDSA(base64PublicKey string) (*ecdsa.PublicKey, error) {
	pk, err := base64.StdEncoding.DecodeString(base64PublicKey)
	if err != nil {
		return nil, err
	}

	publicKey, err := x509.ParsePKIXPublicKey(pk)
	if err != nil {
		return nil, err
	}

	return publicKey.(*ecdsa.PublicKey), nil
}

// VerifySignature uses the ECDSA publicKey and verifies received payload and signature
func VerifySignature(publicKey *ecdsa.PublicKey, payload []byte, signature, timestamp string) (bool, error) {
	signatureBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, err
	}

	ecdsaSig := &RS{}
	_, err = asn1.Unmarshal(signatureBytes, ecdsaSig)
	if err != nil {
		return false, err
	}

	hash := sha256.New()
	hash.Write([]byte(timestamp))
	hash.Write(payload)

	return ecdsa.Verify(publicKey, hash.Sum(nil), ecdsaSig.R, ecdsaSig.S), nil
}
