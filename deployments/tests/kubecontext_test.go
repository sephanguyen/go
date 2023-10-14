package tests

import (
	"path/filepath"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/execwrapper"
	skaffoldwrapper "github.com/manabie-com/backend/internal/golibs/execwrapper/skaffold"
	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLocalKubeContext checks that when ENV=local, the config's
// deploy.kubeContext field must be "kind-kind".
// This is to prevent developers accidentially deploying local k8s onto production clusters.
//
// See: https://manabie.atlassian.net/browse/LT-24825
// Reference: https://skaffold.dev/docs/references/yaml/#deploy-kubeContext
func TestLocalKubeContext(t *testing.T) {
	t.Parallel()

	configs, err := skaffoldwrapper.New().E(vr.EnvLocal).P(vr.PartnerManabie).Filename("skaffold.local.yaml").Diagnose()
	require.NoError(t, err)

	ctxname := "kind-kind"
	for _, c := range configs {
		assert.NotEqualf(t, CheckFailed, isKubeContextLocal(t, c, ctxname),
			"c.Deploy.KubeContext field of config %q should be %q but is %q",
			c.Metadata.Name, ctxname, c.Deploy.KubeContext,
		)
	}
}

func newBool(b bool) *bool {
	return &b
}

// TestLocalKubeContext_Skaffold2 is similar to TestLocalKubeContext, but for skaffoldv2 files.
func TestLocalKubeContext_Skaffold2(t *testing.T) {
	t.Parallel()

	type testcase struct {
		fs skaffoldwrapper.FlagSet
		es skaffoldwrapper.EnvSet
	}
	testcases := []testcase{
		{fs: skaffoldwrapper.FlagSet{F: "skaffold2.appsmith.yaml"}, es: skaffoldwrapper.EnvSet{APPSMITH_DEPLOYMENT_ENABLED: "true"}},
		{fs: skaffoldwrapper.FlagSet{F: "skaffold2.backbone.yaml"}},
		{fs: skaffoldwrapper.FlagSet{F: "skaffold2.backend.yaml"}},
		{fs: skaffoldwrapper.FlagSet{F: "skaffold2.camel-k.yaml"}, es: skaffoldwrapper.EnvSet{CAMEL_K_ENABLED: "true"}},
		{fs: skaffoldwrapper.FlagSet{F: "skaffold2.cp-ksql-server.yaml"}},
		{fs: skaffoldwrapper.FlagSet{F: "skaffold2.emulator.yaml"}},
	}

	for _, tc := range testcases {
		tc := tc
		tcname := tc.fs.F // use filename as the test case name
		t.Run(tcname, func(t *testing.T) {
			configs, err := skaffoldwrapper.New().EnvSet(tc.es).FlagSet(tc.fs).
				E(vr.EnvLocal).P(vr.PartnerManabie).V2Diagnose()
			for _, c := range configs {
				require.NoError(t, err)
				require.Equal(t, "kind-kind", c.Deploy.KubeContext)
			}
		})
	}
}

type KubeContextCheckResult int

const (
	CheckSuccess KubeContextCheckResult = iota
	CheckFailed
	CheckSkipped
)

// assertKubeContext asserts that the kubeContext field in c matches ctxname.
// When c has no helm deployment/hooks, it is ignored from assertions.
func isKubeContextLocal(t *testing.T, c latest.SkaffoldConfig, ctxname string) KubeContextCheckResult {
	if c.Deploy.HelmDeploy == nil {
		// skip when no helm deployment
		return CheckSkipped
	}

	if len(c.Deploy.HelmDeploy.Releases) == 0 &&
		len(c.Deploy.HelmDeploy.LifecycleHooks.PreHooks) == 0 &&
		len(c.Deploy.HelmDeploy.LifecycleHooks.PostHooks) == 0 {
		// skip when helm deployment is defined but is a no-op (nothing to do)
		return CheckSkipped
	}

	if c.Deploy.KubeContext == ctxname {
		return CheckSuccess
	}
	return CheckFailed
}

func configAbsPath(p string) string {
	return filepath.Join(execwrapper.RootDirectory(), p)
}

// This test tests isKubeContextLocal function.
func TestAssertKubeContext(t *testing.T) {
	t.Parallel()

	c := latest.SkaffoldConfig{
		Pipeline: latest.Pipeline{
			Deploy: latest.DeployConfig{
				DeployType:  latest.DeployType{HelmDeploy: &latest.HelmDeploy{Releases: []latest.HelmRelease{{Name: "demo-deploy"}}}},
				KubeContext: "kind-kind",
			},
		},
	}
	assert.Equal(t, CheckSuccess, isKubeContextLocal(t, c, "kind-kind"))

	c = latest.SkaffoldConfig{
		Pipeline: latest.Pipeline{
			Deploy: latest.DeployConfig{
				DeployType:  latest.DeployType{HelmDeploy: &latest.HelmDeploy{Releases: []latest.HelmRelease{{Name: "demo-deploy"}}}},
				KubeContext: "not-kind-kind",
			},
		},
	}
	assert.Equal(t, CheckFailed, isKubeContextLocal(t, c, "kind-kind"))

	c = latest.SkaffoldConfig{
		Pipeline: latest.Pipeline{
			Deploy: latest.DeployConfig{
				KubeContext: "not-kind-kind",
			},
		},
	}
	assert.Equal(t, CheckSkipped, isKubeContextLocal(t, c, "kind-kind"))
}
