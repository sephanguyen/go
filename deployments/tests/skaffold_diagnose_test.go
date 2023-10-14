package tests

import (
	"testing"

	skaffoldwrapper "github.com/manabie-com/backend/internal/golibs/execwrapper/skaffold"
	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	skaffoldv2schema "github.com/GoogleContainerTools/skaffold/v2/pkg/skaffold/schema/latest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSkaffoldDiagnose does a simple tests on skaffoldwrapper.Command.Diagnose.
func TestSkaffoldDiagnose(t *testing.T) {
	t.Parallel()

	res, err := skaffoldwrapper.New().E(vr.EnvLocal).P(vr.PartnerManabie).Filename("skaffold.local.yaml").Diagnose()
	require.NoError(t, err)

	names := getReleaseNames(res)
	assert.Contains(t, names, "cert-manager")
	assert.Contains(t, names, "local-manabie-gateway")
	assert.Contains(t, names, "kafka-connect")
	assert.Contains(t, names, "kafka-ui")
	assert.Contains(t, names, "manabie-all-in-one")
	assert.Contains(t, names, "nats-jetstream")
	assert.Contains(t, names, "elastic")
	assert.Contains(t, names, "kafka")
	assert.Contains(t, names, "cp-schema-registry")
}

func getReleaseNames(configList []latest.SkaffoldConfig) []string {
	names := []string{}
	for _, c := range configList {
		if c.Deploy.HelmDeploy == nil {
			continue
		}
		for _, r := range c.Deploy.HelmDeploy.Releases {
			names = append(names, r.Name)
		}
	}
	return names
}

func getHelmReleaseNamesV2(configList []skaffoldv2schema.SkaffoldConfig) []string {
	names := []string{}
	for _, c := range configList {
		if c.Deploy.LegacyHelmDeploy == nil {
			continue
		}
		for _, r := range c.Deploy.LegacyHelmDeploy.Releases {
			names = append(names, r.Name)
		}
	}
	return names
}
