package configs

import (
	"fmt"
	"log"
	"os"
	"testing"

	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testSecret testConfig

func (b testSecret) Path(_ vr.P, _ vr.E, _ vr.S) string {
	return sopsSecretFilePath
}

func setSOPSSecrets() (err error, cleanup func()) {
	err = func() error {
		err := os.MkdirAll(tmpSubDir, 0o777)
		if err != nil {
			return fmt.Errorf("os.MkdirAll: %s", err)
		}
		if err := setServiceCredential(); err != nil {
			return fmt.Errorf("setServiceCredential: %s", err)
		}
		if err := setSOPSSecret(); err != nil {
			return fmt.Errorf("setSopsSecret: %s", err)
		}
		return nil
	}()

	cleanup = func() {
		err := os.RemoveAll(tmpSubDir)
		if err != nil {
			log.Printf("could not clean up temp directory: %s", err)
		}
	}

	if err != nil {
		cleanup()
		return err, nil
	}
	return nil, cleanup
}

// This test should never be run in parallel.
func TestLoadAndDecrypt(t *testing.T) {
	err, cleanup := setSOPSSecrets()
	defer cleanup()
	require.NoError(t, err)

	c, err := LoadAndDecrypt[testSecret](vr.PartnerManabie, vr.EnvLocal, vr.ServiceBob)
	require.NoError(t, err)
	assert.Equal(t, "secret.secret", c.Secret)
	assert.Equal(t, "secret.secret", c.SubConfig.Secret)
}
