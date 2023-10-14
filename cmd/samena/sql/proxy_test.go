package sql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloudSQLConfig(t *testing.T) {
	t.Parallel()
	t.Run("clone database should have \"clone-\" prefix", func(t *testing.T) {
		t.Parallel()
		e := "dorp"
		for _, p := range []string{
			"jprep",
			"synersia",
			"renseikai",
			"ga", "aic",
			"tokyo",
		} {
			for _, i := range []string{"common", "lms"} {
				c, err := getCloudSQLConfig(p, e, i)
				require.NoError(t, err)
				assert.Contains(t, c.connectionString, "clone-")
			}
		}
	})

	t.Run("non-prod databases should not be in prod projects", func(t *testing.T) {
		t.Parallel()
		for _, e := range []string{"stag", "uat"} {
			for _, p := range []string{"manabie", "jprep"} {
				for _, i := range []string{"common", "lms"} {
					c, err := getCloudSQLConfig(p, e, i)
					require.NoError(t, err)
					assert.NotContains(t, c.connectionString, "student-coach-e1e95:")
					assert.NotContains(t, c.connectionString, "live-manabie:")
					assert.NotContains(t, c.connectionString, "synersia:")
					assert.NotContains(t, c.connectionString, "production-renseikai:")
				}
			}
		}
	})
}
