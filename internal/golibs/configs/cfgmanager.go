package configs

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"go.mozilla.org/sops/v3/decrypt"
	"gopkg.in/yaml.v3"
)

// MustLoadConfig will panic if it cannot create client or parse the config/secret.
func MustLoadConfig(ctx context.Context, commonConfigPath, configPath, secretsPath string, dst interface{}) {
	if err := Load(ctx, commonConfigPath, true, dst); err != nil {
		log.Panicf("cannot load common configuration from %s: %s", commonConfigPath, err.Error())
	}
	if err := Load(ctx, configPath, true, dst); err != nil {
		log.Panicf("cannot load configuration from %s: %s", configPath, err.Error())
	}
	if err := DecryptFile(secretsPath, dst, decrypt.File); err != nil {
		log.Panicf("cannot load encrypted configuration from %s: %s", secretsPath, err.Error())
	}
}

// Load loads plain yaml configuration into dest
func Load(ctx context.Context, filePath string, isPlaintext bool, dest interface{}) error {
	rawContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("cannot read file %s, err: %w", filePath, err)
	}

	return yaml.Unmarshal(rawContent, dest)
}

func IsSecretV2(fp string) bool {
	// For the prod's migration files, encrypt them using the old keys still,
	// so that tech leads cannot decrypt them.
	if strings.Contains(fp, "deployments/helm/") &&
		strings.Contains(fp, "secrets") &&
		strings.Contains(fp, "prod") &&
		strings.Contains(fp, "_migrate") {
		return false
	}

	// For other files, encrypt them using the new keys.
	if strings.Contains(fp, "manabie-all-in-one/charts") && strings.Contains(fp, "secrets") {
		return true
	}
	return strings.Contains(fp, "v2")
}
