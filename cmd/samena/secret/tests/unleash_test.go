package tests

import (
	"fmt"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/execwrapper"
	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var dbConnWithIAMRe = regexp.MustCompile(`^postgres://([\w\d\%\-\@\.]+)@(?:127\.0\.0\.1):5432/([\w_]+)\?sslmode=disable$`)

// Only run for Tokyo and JPREP
// see https://manabie.atlassian.net/browse/LT-13324
func TestUnleashConfig(t *testing.T) {
	rawPasswordMap, err := configs.LoadAndDecrypt[unleashRawPasswords](vr.PartnerManabie, vr.EnvLocal, vr.ServiceUnleash)
	require.NoError(t, err)
	testfunc := func(t *testing.T, p vr.P, e vr.E) {
		uconf, err := configs.LoadAndDecrypt[unleashConfig](p, e, vr.ServiceUnleash)
		require.NoError(t, err, "failed to load unleash secret")

		testUnleashDBConnection(t, p, e, uconf.DBConnection, true)
		rawPasswordMap.CheckHashedPassword(t, p, e, uconf.AdminPassword)
	}
	vr.Iter(t).SkipE(vr.EnvPreproduction).P(vr.PartnerTokyo, vr.PartnerJPREP).IterPE(testfunc)
}

func testUnleashDBConnection(t *testing.T, p vr.P, e vr.E, connStr string, isV2 bool) {
	m := dbConnWithIAMRe.FindStringSubmatch(connStr)
	require.Len(t, m, 3,
		"failed to regexp parse connection string %q: unexpected number of output elements: %d (should have 3)",
		connStr, len(m),
	)

	user := m[1]
	assert.Equal(t, expectedUnleashUser(p, e), user, "unexpected unleash user")

	dbname := m[2]
	assert.Equal(t, expectedDBName(p, e, isV2), dbname, "unexpected unleash dbname")
}

func expectedUnleashUser(p vr.P, e vr.E) string {
	userPrefix := vr.ServiceAccountPrefix(p, e)
	projectID := vr.GCPProjectID(p, e)
	// unleash will urldecode this name before using it
	// thus we need to urlencode it (i.e. @ -> %40)
	// example: stag-unleash@staging-manabie-online.iam -> stag-unleash%40staging-manabie-online.iam
	return fmt.Sprintf("%sunleash%%40%s.iam", userPrefix, projectID)
}

func expectedDBName(p vr.P, e vr.E, isV2 bool) string {
	dbNamePrefix := vr.DatabaseNamePrefix(p, e)
	if isV2 {
		return dbNamePrefix + "unleashv2"
	}
	return dbNamePrefix + "unleash"
}

// unleashConfig represents the secret YAMLs of Unleash.
type unleashConfig struct {
	DBConnection  string `yaml:"db_connection"`
	AdminPassword string `yaml:"admin_password"`
}

func (unleashConfig) Path(p vr.P, e vr.E, _ vr.S) string {
	return filepath.Join(
		execwrapper.RootDirectory(),
		fmt.Sprintf("deployments/helm/platforms/unleash/secrets/%v/%v/unleash.secrets.encrypted.yaml", p, e),
	)
}

type unleashRawPasswords map[string]map[string]string

func (unleashRawPasswords) Path(_ vr.P, _ vr.E, _ vr.S) string {
	return filepath.Join(
		execwrapper.RootDirectory(),
		fmt.Sprintf("deployments/helm/platforms/unleash/secrets/unleash_raw_passwords.secrets.encrypted.yaml"),
	)
}

// GenerateHashedPassword generates a new hashed password from the raw password from
// unleash_raw_passwords.secrets.encrypted.yaml
func (u *unleashRawPasswords) GenerateHashedPassword(p vr.P, e vr.E) ([]byte, error) {
	rawPasswordByEnv, ok := (*u)[e.String()]
	if !ok {
		return nil, fmt.Errorf("password for environment %v was not found", e)
	}
	rawPassword, ok := rawPasswordByEnv[p.String()]
	if !ok {
		return nil, fmt.Errorf("password for partner %v was not found in environment %v", p, e)
	}
	return bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
}

// CheckPassword tests that the input hashed password (stored in Unleash's secrets)
// matches with the documented raw password.
func (u *unleashRawPasswords) CheckHashedPassword(t *testing.T, p vr.P, e vr.E, hashedPassword string) {
	rawPasswordByEnv, ok := (*u)[e.String()]
	require.True(t, ok, "password for environment %v was not found", e)
	rawPassword, ok := rawPasswordByEnv[p.String()]
	require.True(t, ok, "password for partner %v was not found in environment %v", p, e)

	var newHasedPassword []byte
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(rawPassword))
	if err != nil {
		var err2 error
		newHasedPassword, err2 = u.GenerateHashedPassword(p, e)
		require.NoError(t, err2, "failed to generate a suggested hash password")
	}
	require.NoError(t, err, "invalid hash password, try to use this newly generated hash instead: %s", string(newHasedPassword))
}
