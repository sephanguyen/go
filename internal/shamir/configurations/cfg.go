package configurations

import (
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/configs"

	"go.mozilla.org/sops/v3/decrypt"
)

var (
	errNotEnoughKeys = fmt.Errorf("verifier: not enough key")
	errInvalidPem    = fmt.Errorf("verifier: invalid pem file")
)

// Config belong to Shamir
type Config struct {
	Common              configs.CommonConfig
	Vendor              string
	KeysGlob            string `yaml:"keys_glob"`
	Issuers             []configs.TokenIssuerConfig
	GRPC                configs.GRPCConfig
	NatsJS              configs.NatsJetStreamConfig
	PostgresV2          configs.PostgresConfigV2    `yaml:"postgres_v2"`
	UnleashClientConfig configs.UnleashClientConfig `yaml:"unleash_client"`
	PrimaryKeyFile      string                      `yaml:"primary_key_file"`
	OpenAPI             OpenAPI                     `yaml:"open_api"`
	SalesforceConfigs   SalesforceConfigs           `yaml:"salesforce"`
}

type OpenAPI struct {
	AESKey string `yaml:"aes_key"`
	AESIV  string `yaml:"aes_iv"`
}

type SopsPrivateKeyPayLoad struct {
	Data string `yaml:"data"`
}

type SalesforceOrgConfig struct {
	Key      string `yaml:"key"`
	ClientID string `yaml:"client_id"`
}

type SalesforceConfigs struct {
	Aud                 string                         `yaml:"aud"`
	AccessTokenEndpoint string                         `yaml:"access_token_endpoint"`
	Configurations      map[string]SalesforceOrgConfig `yaml:",inline"`
}

func checkFileName(filePath string, fileName string) bool {
	result := strings.Split(filePath, "/")

	if len(result) > 0 {
		return result[len(result)-1] == fileName
	}
	return false
}

func LoadPrivateKeysWithSopsFormat(keysGlob string, primaryKeyFile string) (map[string]*rsa.PrivateKey, string, error) {
	return loadPrivateKeysWithSopsFormat(keysGlob, primaryKeyFile, decrypt.File)
}

func loadPrivateKeysWithSopsFormat(keysGlob string, primaryKeyFile string, f configs.DecryptFunc) (map[string]*rsa.PrivateKey, string, error) {
	matches, err := filepath.Glob(keysGlob)
	if err != nil {
		return nil, "", err
	}

	keys := make(map[string]*rsa.PrivateKey, len(matches))
	var privateKeyPayload SopsPrivateKeyPayLoad
	var primaryKeyID string
	for _, name := range matches {
		err = configs.DecryptFile(name, &privateKeyPayload, f)
		if err != nil {
			return nil, "", fmt.Errorf("configs.DecryptSopsFile error: %v", err)
		}

		block, _ := pem.Decode([]byte(privateKeyPayload.Data))
		if block == nil {
			return nil, "", errInvalidPem
		}

		privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, "", err
		}

		h := sha1.New() //nolint:gosec
		h.Write(block.Bytes)
		keyID := fmt.Sprintf("%x", h.Sum(nil))
		keys[keyID] = privateKey

		if checkFileName(name, primaryKeyFile) {
			primaryKeyID = keyID
		}
	}

	if n := len(keys); n < 2 {
		return nil, "", fmt.Errorf("%w: expecting at least 2 pem files, got %d", errNotEnoughKeys, n)
	}

	return keys, primaryKeyID, nil
}
