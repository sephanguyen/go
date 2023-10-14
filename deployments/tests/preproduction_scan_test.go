package tests

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	skaffoldwrapper "github.com/manabie-com/backend/internal/golibs/execwrapper/skaffold"
	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/stretchr/testify/require"
)

// TestPreproductionDatabaseConnection checks that the preproduction manifest do not
// contain any production's database URLs or k8s services.
func TestPreproductionDatabaseConnection(t *testing.T) {
	t.Parallel()

	// sanity test to ensure this method works
	t.Run("sanity test", func(t *testing.T) {
		t.Parallel()

		re1 := regexp.MustCompile(`student-coach-e1e95:asia-northeast1:clone-prod-tokyo`)
		re2 := regexp.MustCompile(`student-coach-e1e95:asia-northeast1:clone-prod-tokyo-lms-b2dc4508`)
		out, err := skaffoldwrapper.New().E(vr.EnvPreproduction).P(vr.PartnerTokyo).
			Filename("skaffold.manaverse.yaml").RenderRaw()
		{
			out2, err := skaffoldwrapper.New().E(vr.EnvPreproduction).P(vr.PartnerTokyo).
				Filename("skaffold2.backend.yaml").V2RenderRaw()
			require.NoError(t, err)
			out = append(out, []byte("\n---\n")...)
			out = append(out, out2...)
		}
		require.NoError(t, err)
		require.True(t, re1.Match(out), "failed to find preproduction common database settings in manifest")
		require.True(t, re2.Match(out), "failed to find preproduction lms database settings in manifest")

		nsRe := regexp.MustCompile(`\.dorp-\w+-(?:appsmith|elastic|frontend|kafka|nats-jetstream|services|unleash)`)
		require.True(t, nsRe.Match(out), "ailed to find preproduction namespace configs in manifest")
	})

	buildRegexpStr := func(in []string) string {
		processed := make([]string, 0, len(in))
		for _, v := range in {
			processed = append(processed, fmt.Sprintf(`(?:%s)`, v))
		}
		return strings.Join(processed, `|`)
	}

	// list of actual production databases
	bannedConns := []string{
		"student-coach-e1e95:asia-northeast1:jp-partners-b04fbb69",
		"student-coach-e1e95:asia-northeast1:prod-jprep-d995522c",
		"student-coach-e1e95:asia-northeast1:prod-tokyo",
		"student-coach-e1e95:asia-northeast1:prod-tokyo-auth-42c5a298",
		"student-coach-e1e95:asia-northeast1:prod-tokyo-data-warehouse-251f01f8",
		"student-coach-e1e95:asia-northeast1:prod-tokyo-lms-b2dc4508",
		"production-renseikai:asia-northeast1:renseikai-83fc",
	}
	rdRs := buildRegexpStr(bannedConns)
	dbRe := regexp.MustCompile(rdRs)

	// regexp to find production namespaces
	nsRe := regexp.MustCompile(`\.prod-\w+-(?:appsmith|elastic|frontend|kafka|nats-jetstream|services|unleash)`)

	// main testing function
	testfunc := func(t *testing.T, p vr.P, e vr.E) {
		out, err := skaffoldwrapper.New().E(e).P(p).Filename("skaffold.manaverse.yaml").RenderRaw()
		require.NoError(t, err)

		{
			out2, err := skaffoldwrapper.New().E(vr.EnvPreproduction).P(vr.PartnerTokyo).
				Filename("skaffold2.backend.yaml").V2RenderRaw()
			require.NoError(t, err)
			out = append(out, []byte("\n---\n")...)
			out = append(out, out2...)
		}

		require.False(t, dbRe.Match(out), "production database connection found in preproduction manifest")
		require.False(t, nsRe.Match(out), "production namespace config found in preproduction manifest")
	}
	vr.Iter(t).E(vr.EnvPreproduction).IterPE(testfunc)
}
