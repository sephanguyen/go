package services

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/shamir/configurations"

	"github.com/golang-jwt/jwt"
)

const ContentTypeForm = "application/x-www-form-urlencoded"

type SalesforceService struct {
	Config configurations.SalesforceConfigs
}

type ExchangeSalesforceTokenResp struct {
	AccessToken      string `json:"access_token"`
	Scope            string `json:"scope"`
	InstanceURL      string `json:"instance_url"`
	ID               string `json:"id"`
	TokenType        string `json:"token_type"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (ss SalesforceService) GetAccessToken(orgID, userID string) (string, error) {
	resp, err := ss.getAccessToken(orgID, userID)
	if err != nil {
		return "", err
	}

	if resp.Error != "" {
		return "", fmt.Errorf("ss.GetAccessToken %s: %s", resp.Error, resp.ErrorDescription)
	}

	return resp.AccessToken, nil
}

func (ss SalesforceService) getAccessToken(orgID string, userID string) (ExchangeSalesforceTokenResp, error) {
	orgConfig := ss.Config.Configurations[mapOrgIDAndDomainName[orgID]]

	privateKey, err := toRSAPrivateKey(orgConfig.Key)
	if err != nil {
		return ExchangeSalesforceTokenResp{}, err
	}

	token, err := signToken(privateKey, orgConfig.ClientID, ss.Config.Aud, userID)
	if err != nil {
		return ExchangeSalesforceTokenResp{}, err
	}

	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	data.Set("assertion", token)

	// Encode the form data as AccessTokenURL-encoded format
	payload := strings.NewReader(data.Encode())

	resp, err := http.Post(ss.Config.AccessTokenEndpoint, ContentTypeForm, payload)
	if err != nil {
		return ExchangeSalesforceTokenResp{}, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ExchangeSalesforceTokenResp{}, err
	}

	accessToken := ExchangeSalesforceTokenResp{}
	if err := json.Unmarshal(respBody, &accessToken); err != nil {
		return ExchangeSalesforceTokenResp{}, err
	}

	return accessToken, nil
}

func toRSAPrivateKey(key string) (*rsa.PrivateKey, error) {
	// Decode private key
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, fmt.Errorf("invalid private key")
	}
	// Parse the private key
	privateKeyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %v", err)
	}
	// Type assert to *rsa.PrivateKey
	privateKey, ok := privateKeyInterface.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not a RSA private key")
	}
	return privateKey, nil
}

func signToken(privateKey *rsa.PrivateKey, clientID string, aud string, subject string) (string, error) {
	// TODO set expire time by config
	claims := &jwt.StandardClaims{
		Issuer:    clientID,
		Audience:  aud,
		Subject:   subject,
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // Expiration time
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

var mapOrgIDAndDomainName = map[string]string{
	fmt.Sprint(constants.ManabieSchool): "manabie",
	fmt.Sprint(constants.UsermgmtSF):    "usermgmt-sf",
}
