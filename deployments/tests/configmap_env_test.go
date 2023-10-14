package tests

import (
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"
	skaffoldwrapper "github.com/manabie-com/backend/internal/golibs/execwrapper/skaffold"
	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

type serviceConfig struct {
	Common configs.CommonConfig `yaml:"common"`
}

// TestConfigMapCommonValues ensures that `common.organization`, `common.environment`
// and `common.actual_environment` values in services' configs are correct.
//
// It also checks that these fields are not overriden by per-env config (since it would
// make no sense to re-set them).
func TestConfigMapCommonValues(t *testing.T) {
	// Instead of using vr.IterPES here, we use vr.IterPE to reuse the
	// generated k8s manifests per env/org combination (so as to speed up this test).
	vr.Iter(t).IterPE(testConfigMapCommonValues)
}

func testConfigMapCommonValues(t *testing.T, p vr.P, e vr.E) {
	manifestObjects, err := skaffoldwrapper.New().E(e).P(p).Filename("skaffold2.backend.yaml").V2CachedRender()
	require.NoError(t, err)

	configmaps := make(map[string]*corev1.ConfigMap, 64)
	for _, o := range manifestObjects {
		switch v := o.(type) {
		// TODO: there might be duplicates here (same name but different namespace)
		case *corev1.ConfigMap:
			if _, exists := configmaps[v.ObjectMeta.Name]; exists {
				t.Logf("duplicated configmap found with name %s (are there configmaps with the same name but located in different namespaces?)", v.ObjectMeta.Name)
			}
			configmaps[v.ObjectMeta.Name] = v
		}
	}

	testService := func(s vr.S) {
		svcName := s.String()
		t.Run(svcName, func(t *testing.T) {
			// asserts fields have correct values for interested fields
			commonCMName := commonConfigmapName(svcName)
			conf, err := getConfigFromManifest[serviceConfig](configmaps, nil, svcName, commonCMName, nil, nil)
			require.NoError(t, err)
			require.Equal(t, p.String(), conf.Common.Organization, `wrong "organization" field in %s`, *commonCMName)
			require.Equal(t, e.String(), conf.Common.ActualEnvironment, `wrong "actual_environment" value in %s`, *commonCMName)
			expectedEnv := e.String()
			if e == vr.EnvPreproduction {
				expectedEnv = vr.EnvProduction.String()
			}
			require.Equal(t, expectedEnv, conf.Common.Environment, `wrong "environment" value in %s`, *commonCMName)

			// asserts per-env configs do not reset those fields
			cmName := configmapName(svcName)
			conf, err = getConfigFromManifest[serviceConfig](configmaps, nil, svcName, nil, cmName, nil)
			require.NoError(t, err)
			require.Empty(t, conf.Common.Organization, "setting `common.organization` field in not allowed in %s", *cmName)
			require.Empty(t, conf.Common.ActualEnvironment, "setting `common.actual_environment` field in not allowed in %s", *cmName)
			require.Empty(t, conf.Common.Environment, "setting `common.environment` field in not allowed in %s", *cmName)
		})
	}

	for _, s := range vr.BackendServices() {
		if !vr.IsBackendServiceEnabled(p, e, s) {
			continue
		}
		testService(s)
	}
}
