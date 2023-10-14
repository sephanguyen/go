package tests

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/execwrapper"
	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"go.mozilla.org/sops/v3/decrypt"
)

type ServiceNATJSPassword struct {
	NATSJS struct {
		Password string `yaml:"password"`
	} `yaml:"natsjs"`
}

func (ServiceNATJSPassword) Path(p vr.P, e vr.E, s vr.S) string {
	return filepath.Join(
		execwrapper.RootDirectory(),
		fmt.Sprintf("deployments/helm/manabie-all-in-one/charts/%v/secrets/%v/%v/%v.secrets.encrypted.yaml", s, p, e, s),
	)
}

// TestNATSJetstreamAccountPasswords ensures that, if a service use NATS,
// the password in that service's secret match the password in NATS server.
func TestNATSJetstreamAccountPasswords(t *testing.T) {
	t.Parallel()

	checkValueQuoted := func(s vr.S, rawData []byte) error {
		rs := fmt.Sprintf(`^%v_password=\s*\".+\"(?:\r?\n)?$`, s)
		re := regexp.MustCompile(rs)
		if !re.Match(rawData) {
			return fmt.Errorf("decrypted data %q has invalid format (does not match regexp \"%s\")", string(rawData), rs)
		}
		return nil
	}

	getExpectedPassword := func(p vr.P, e vr.E, s vr.S) (string, error) {
		fp := execwrapper.Absf("deployments/helm/platforms/nats-jetstream/secrets/%v/%v/%v_nats.secrets.encrypted.env", p, e, s)
		d, err := decrypt.File(fp, "")
		if err != nil {
			return "", fmt.Errorf("decrypt.File: %w", err)
		}
		if err := checkValueQuoted(s, d); err != nil {
			return "", err
		}
		v, err := godotenv.UnmarshalBytes(d)
		if err != nil {
			return "", fmt.Errorf("godotenv.UnmarshalBytes: %w", err)
		}
		out, ok := v[s.String()+"_password"]
		if !ok {
			return "", fmt.Errorf(`field "%v_password" not found in %s`, s, fp)
		}
		return out, nil
	}

	testfunc := func(t *testing.T, p vr.P, e vr.E, s vr.S) {
		serviceConf, err := configs.LoadAndDecrypt[ServiceNATJSPassword](p, e, s)
		require.NoError(t, err, "failed to load natsjs password from service's secret")
		actualPassword := serviceConf.NATSJS.Password

		expectedPassword, err := getExpectedPassword(p, e, s)
		if errors.Is(err, os.ErrNotExist) {
			// this error is allowed only when service does not use natsjs
			require.Empty(t, actualPassword, "missing natsjs password in server's config (%v_nats.secrets.encrypted.env does not exist)", s)
			return
		}
		require.NoError(t, err, "failed to load natsjs password from nat server's secret")

		require.Equal(t, expectedPassword, actualPassword)
	}

	vr.Iter(t).SkipE(vr.EnvPreproduction).SkipDisabledServices().IterPES(testfunc)
}
