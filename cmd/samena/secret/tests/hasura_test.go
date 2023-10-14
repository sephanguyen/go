package tests

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/execwrapper"
	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type Hasura struct {
	DatabaseURL string          `yaml:"HASURA_GRAPHQL_DATABASE_URL"`
	JWTSecret   HasuraJWTSecret `yaml:"HASURA_GRAPHQL_JWT_SECRET"`
	AdminSecret string          `yaml:"HASURA_GRAPHQL_ADMIN_SECRET"`
}

type HasuraJWTSecret struct {
	Type     string `yaml:"type"`
	JWKURL   string `yaml:"jwk_url"`
	Audience string `yaml:"audience"`
	Issuer   string `yaml:"issuer"`
}

func (h *HasuraJWTSecret) UnmarshalYAML(value *yaml.Node) error {
	var raw string
	if err := value.Decode(&raw); err != nil {
		return fmt.Errorf("failed to decode raw string: %s", err)
	}
	type rawHasuraJWTSecret HasuraJWTSecret // to avoid infinite recursion
	return yaml.Unmarshal([]byte(raw), (*rawHasuraJWTSecret)(h))
}

// Note that s should the business service where the Hasura belongs to, not Hasura itself.
func (Hasura) Path(p vr.P, e vr.E, s vr.S) string {
	return filepath.Join(
		execwrapper.RootDirectory(),
		fmt.Sprintf("deployments/helm/manabie-all-in-one/charts/%v/secrets/%v/%v/hasura.secrets.encrypted.yaml", s, p, e),
	)
}

// TODO: parse deployments/helm/manabie-all-in-one/values.yaml instead
func hasuraEnabledServices(p vr.P) []vr.S {
	res := []vr.S{
		vr.ServiceBob,
		vr.ServiceEntryExitMgmt,
		vr.ServiceEureka,
		vr.ServiceFatima,
		vr.ServiceInvoiceMgmt,
		vr.ServiceLessonMgmt,
		vr.ServiceTimesheet,
	}
	if p != vr.PartnerJPREP {
		res = append(res, vr.ServiceCalendar)
	}
	return res
}

func TestHasuraDatabaseURL(t *testing.T) {
	testfunc := func(t *testing.T, p vr.P, e vr.E) {
		// We test bob service first, so that we use its password to match against
		s := vr.ServiceBob
		bobHasuraCfg, err := configs.LoadAndDecrypt[Hasura](p, e, s)
		require.NoError(t, err)

		// Due to legacy reasons, some envs use accounts that are
		// different than expected
		hasPrefix := true
		if e == vr.EnvProduction && p == vr.PartnerGA {
			hasPrefix = false
		} else if e == vr.EnvUAT && p == vr.PartnerManabie {
			hasPrefix = false
		}
		bobDBConn, err := database.NewPGConnString(
			p, e, s, bobHasuraCfg.DatabaseURL,
			database.RequireDBUser("hasura", hasPrefix), database.ForcePassword(true),
			database.EnableIAMLogin(isHasuraUsingIAMLogin(p, e, s)),
			database.RequireIAMDBUser(getExpectedHasuraIAMPrefix(p, e, s)),
		)
		require.NoError(t, err, "failed to parse connection string for: %s/%s/%s", p, e, s)
		bobDBConn.AssertAll(t)

		// Next, test for all other services
		commonPassword := bobDBConn.DBPassword()
		for _, s := range hasuraEnabledServices(p) {
			if s == vr.ServiceBob {
				continue
			}
			cfg, err := configs.LoadAndDecrypt[Hasura](p, e, s)
			require.NoError(t, err)
			dbConn, err := database.NewPGConnString(
				p, e, s, cfg.DatabaseURL,
				database.RequireDBUser("hasura", hasPrefix), database.ForcePassword(true),
				database.RequireDBPassword(commonPassword), database.RequireDBName(s.String()),
				database.EnableIAMLogin(isHasuraUsingIAMLogin(p, e, s)),
				database.RequireIAMDBUser(getExpectedHasuraIAMPrefix(p, e, s)),
			)
			require.NoError(t, err, "failed to parse connection string for: %s/%s/%s", p, e, s)
			dbConn.AssertAll(t)
		}
	}
	vr.Iter(t).SkipE(vr.EnvLocal, vr.EnvPreproduction).IterPE(testfunc)
}

func TestHasuraJWTSecret(t *testing.T) {
	testfunc := func(t *testing.T, p vr.P, e vr.E) {
		for _, s := range hasuraEnabledServices(p) {
			cfg, err := configs.LoadAndDecrypt[Hasura](p, e, s)
			require.NoError(t, err)

			// This may feel like secret leakage, but these fields are not actually secret.
			// (check "issuers" config field in each service)
			assert.Equal(t, "RS256", cfg.JWTSecret.Type)
			assert.Contains(t,
				[]string{
					"http://shamir:5680/.well-known/jwks.json",
				}, cfg.JWTSecret.JWKURL,
				`"jwk_url" is not valid`,
			)
			assert.Equal(t, getExpectedHasuraAud(p, e), cfg.JWTSecret.Audience)
			assert.Equal(t, "manabie", cfg.JWTSecret.Issuer)
		}
	}
	vr.Iter(t).SkipE(vr.EnvLocal, vr.EnvPreproduction).IterPE(testfunc)
}

func getExpectedHasuraAud(p vr.P, e vr.E) string {
	switch e {
	case vr.EnvLocal:
		return "manabie-local"
	case vr.EnvStaging:
		switch p {
		case vr.PartnerManabie:
			return "manabie-stag"
		case vr.PartnerJPREP:
			return "jprep-stag"
		}
	case vr.EnvUAT:
		switch p {
		case vr.PartnerManabie:
			return "manabie-stag"
		case vr.PartnerJPREP:
			return "803wsd1dyl3x5jz22t"
		}
	case vr.EnvProduction:
		if p == vr.PartnerJPREP {
			return "b5e72419a81ca9e1a5"
		}
	}
	return fmt.Sprintf("%v-%v", e, p)
}

func getExpectedHasuraIAMPrefix(p vr.P, e vr.E, s vr.S) string {
	if p == vr.PartnerJPREP && e == vr.EnvStaging {
		return fmt.Sprintf("stag-jprep-%v-h", s)
	}
	return fmt.Sprintf("%v-%v-h", e, s)
}

func isHasuraUsingIAMLogin(p vr.P, e vr.E, s vr.S) bool {
	switch e {
	case vr.EnvStaging:
		if s == vr.ServiceEntryExitMgmt && p == vr.PartnerJPREP {
			return false
		}
		return true
	case vr.EnvUAT:
		if s == vr.ServiceEntryExitMgmt && p == vr.PartnerJPREP {
			return false
		}
		return true
	default:
		return false
	}
}
