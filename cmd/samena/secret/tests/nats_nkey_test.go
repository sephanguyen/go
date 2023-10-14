package tests

import (
	"fmt"
	"path"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/execwrapper"
	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/joho/godotenv"
	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/require"
	"go.mozilla.org/sops/v3/decrypt"
)

type ControllerSeed struct {
	Data string `yaml:"data"`
}

func (ControllerSeed) Path(p vr.P, e vr.E, _ vr.S) string {
	return path.Join(
		execwrapper.RootDirectory(),
		fmt.Sprintf("deployments/helm/platforms/nats-jetstream/secrets/%v/%v/controller.seed.encrypted.yaml", p, e),
	)
}

func TestNATSJetstreamNKeys(t *testing.T) {
	loadControllerNKey := func(fp string) (string, error) {
		d, err := decrypt.File(fp, "")
		if err != nil {
			return "", fmt.Errorf("decrypt.File: %w", err)
		}
		v, err := godotenv.UnmarshalBytes(d)
		if err != nil {
			return "", fmt.Errorf("godotenv.UnmarshalBytes: %w", err)
		}
		out, ok := v["controller_nkey"]
		if !ok {
			return "", fmt.Errorf(`field "controller_nkey" not found in %s`, fp)
		}
		return out, nil
	}

	testfunc := func(t *testing.T, p vr.P, e vr.E) {
		seedData, err := configs.LoadAndDecrypt[ControllerSeed](p, e, vr.ServiceNATSJetstream)
		require.NoError(t, err, "failed to load nkey seed")

		natsNKey, err := loadControllerNKey(execwrapper.Absf("deployments/helm/platforms/nats-jetstream/secrets/%v/%v/nats.secrets.encrypted.env", p, e))
		require.NoError(t, err, "failed to load nats nkey")

		// Try encrypting and signing with the provided seed
		randomData := []byte("lEr+5YR0qCe0f1JWU07nVWzNnFdCQGc16v47IJI7uYY=") // random string generate with `openssl rand -base64 32`
		user, err := nkeys.FromSeed([]byte(seedData.Data))
		require.NoError(t, err, "failed to create user from private seed %q", seedData.Data)
		signedData, err := user.Sign(randomData)
		require.NoError(t, err, "failed to sign random data using private seed")

		// Try verifying the signed data with the original seed
		err = user.Verify(randomData, signedData)
		require.NoError(t, err, "failed to verify signed data using private seed")

		// Try verifying the signed data with the public key in the config
		user, err = nkeys.FromPublicKey(natsNKey)
		require.NoError(t, err, "failed to create user from public key %q", natsNKey)
		err = user.Verify(randomData, signedData)
		require.NoError(t, err, "failed to verify signed data using public key")
	}
	vr.Iter(t).SkipE(vr.EnvPreproduction).IterPE(testfunc)
}
