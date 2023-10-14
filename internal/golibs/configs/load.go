package configs

import (
	"fmt"
	"os"
	"strings"

	"go.mozilla.org/sops/v3/decrypt"
	"gopkg.in/yaml.v3"
)

// LoadAll loads and merges values from input files.
// Values from secretPath take the highest precedence, then configPath, and finally commonConfigPath.
func LoadAll[T any](commonConfigPath, configPath, secretPath string) (*T, error) {
	return loadAll[T](commonConfigPath, configPath, secretPath, decrypt.File)
}

// MustLoadAll is similar to LoadAll, but panics instead of errors.
func MustLoadAll[T any](commonConfigPath, configPath, secretPath string) *T {
	out, err := LoadAll[T](commonConfigPath, configPath, secretPath)
	if err != nil {
		panic(err.Error())
	}
	return out
}

// DecryptFunc is implemented by SOPS' decrypt.File.
type DecryptFunc func(string, string) ([]byte, error)

// loadAll is similar to LoadAll, but also takes in a function which will be used to decrypt secrets.
func loadAll[T any](commonConfigPath, configPath, secretPath string, f DecryptFunc) (*T, error) {
	out := new(T)
	if err := loadFile(commonConfigPath, out); err != nil {
		return nil, fmt.Errorf("cannot load common configuration from %q: %w", commonConfigPath, err)
	}
	if err := loadFile(configPath, out); err != nil {
		return nil, fmt.Errorf("cannot load configuration from %q: %w", configPath, err)
	}
	if err := DecryptFile(secretPath, out, f); err != nil {
		return nil, fmt.Errorf("cannot load encrypted configuration from %q: %w", secretPath, err)
	}
	return out, nil
}

func loadFile(path string, out interface{}) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(raw, out)
}

func DecryptFile(path string, out interface{}, f DecryptFunc) error {
	if len(strings.TrimSpace(path)) == 0 {
		return nil
	}
	raw, err := f(path, "yaml")
	if err != nil {
		return err
	}
	return yaml.Unmarshal(raw, out)
}
