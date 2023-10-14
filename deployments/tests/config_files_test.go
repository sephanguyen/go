package tests

import (
	"os"
	"testing"

	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/stretchr/testify/require"
)

// TestConfigFiles ensures that configs and secrets are properly set up
// for a service in all environments (if enabled).
//
// To run, a service needs:
//   - common config (e.g. configs/bob.common.config.yaml)
//   - per-env config (e.g. configs/manabie/local/bob.config.yaml)
//   - per-env secret (e.g. secrets/manabie/local/bob.secrets.encrypted.yaml)
//
// TODO(@anhpngt) we can check for hasura configs here as well.
func TestConfigFiles(t *testing.T) {
	t.Parallel()

	testfunc := func(t *testing.T, p vr.P, e vr.E, s vr.S) {
		commonConfigFp := vr.CommonConfigFilepath(s)
		require.NoError(t, checkFileExists(commonConfigFp), "invalid common config, please make sure that it exists")

		configFp := vr.ConfigFilepath(p, e, s)
		require.NoError(t, checkFileExists(configFp), "invalid config, please make sure that it exists")

		secretFp := vr.SecretFilePath(p, e, s)
		require.NoError(t, checkFileExists(secretFp), "invalid secret, please make sure that it exists")

		isMigrationEnabled := vr.GetHelmValues(p, e, s).MigrationEnabled
		if isMigrationEnabled {
			migrateSecretFp := vr.MigrationSecretFilePath(p, e, s)
			require.NoError(t, checkFileExists(migrateSecretFp), "invalid migration secret, please make sure that it exists")
		}
	}
	vr.Iter(t).SkipE(vr.EnvPreproduction).SkipDisabledServices().IterPES(testfunc)
}

func checkFileExists(fp string) error {
	_, err := os.Stat(fp)
	return err
}
