package configs

import (
	"fmt"

	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/magiconair/properties"
	"go.mozilla.org/sops/v3/decrypt"
	"gopkg.in/yaml.v3"
)

type Secret interface {
	Path(partner vr.P, environment vr.E, service vr.S) string
}

// LoadAndDecrypt reads, decrypts, then unmarshals the content to output.
// The path to the file is determined at T.Path().
// The content of the file must be a sops-encrypted yaml structure.
func LoadAndDecrypt[T Secret](p vr.P, e vr.E, s vr.S) (*T, error) {
	var secret T
	fp := secret.Path(p, e, s)
	cleartext, err := decrypt.File(fp, "") // leave the format empty to let the library detect from path
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt SOPS secret from %q: %w", fp, err)
	}
	err = yaml.Unmarshal(cleartext, &secret)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret from %q: %w", fp, err)
	}
	return &secret, nil
}

func LoadAndDecryptProperties[T Secret](p vr.P, e vr.E, s vr.S) (*T, error) {
	var secret T
	fp := secret.Path(p, e, s)
	cleartext, err := decrypt.File(fp, "") // leave the format empty to let the library detect from path
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt SOPS secret from %q: %w", fp, err)
	}
	prop, err := properties.Load(cleartext, properties.UTF8)
	if err != nil {
		return nil, fmt.Errorf("properties.Load failed for %q: %w", fp, err)
	}
	err = prop.Decode(&secret)
	if err != nil {
		return nil, fmt.Errorf("properties.Decode failed for %q: %w", fp, err)
	}
	return &secret, nil
}
