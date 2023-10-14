package tests

import (
	"testing"

	skaffoldwrapper "github.com/manabie-com/backend/internal/golibs/execwrapper/skaffold"

	"github.com/stretchr/testify/require"
)

// TestSkaffoldDiagnoseForMonitoring ensures that the following clusters:
//   - local cluster (kind-kind)
//   - staging/UAT cluster (staging-2)
//   - tokyo
//   - manabie
//   - jp-partners
//
// has proper helm charts.
func TestSkaffoldDiagnoseForMonitoring(t *testing.T) {
	t.Parallel()

	profilesToActivate := map[string]string{
		"kind-kind":   "monitoring",
		"staging-2":   "monitoring,staging-2",
		"manabie":     "monitoring,manabie",
		"jp-partners": "monitoring,jp-partners",
		"tokyo":       "monitoring,tokyo",
	}
	expectedHelmCharts := map[string][]string{
		"kind-kind":   {"prometheus", "opentelemetry-collector", "jaeger-all-in-one", "thanos", "grafana", "kiali", "oncall"},
		"staging-2":   {"prometheus", "opentelemetry-collector", "jaeger-all-in-one", "kiali"},
		"manabie":     {"opentelemetry-collector", "jaeger-all-in-one", "grafana", "oncall", "thanos"},
		"jp-partners": {"prometheus", "opentelemetry-collector", "jaeger-all-in-one", "kiali"},
		"tokyo":       {"prometheus", "opentelemetry-collector", "jaeger-all-in-one", "kiali"},
	}

	for name, profiles := range profilesToActivate {
		t.Run(name, func(t *testing.T) {
			configs, err := skaffoldwrapper.New().Filename("skaffold.monitoring.yaml").Profile(profiles).Diagnose()
			require.NoError(t, err)

			actualHelmCharts := getReleaseNames(configs)
			require.ElementsMatch(t, expectedHelmCharts[name], actualHelmCharts)
		})
	}

	// sanity check with `skaffold render`
	for name, profiles := range profilesToActivate {
		t.Run(name, func(t *testing.T) {
			_, err := skaffoldwrapper.New().Filename("skaffold.monitoring.yaml").Profile(profiles).CachedRender()
			require.NoError(t, err)
		})
	}
}
