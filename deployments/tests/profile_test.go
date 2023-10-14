package tests

import (
	"strings"
	"testing"

	skaffoldwrapper "github.com/manabie-com/backend/internal/golibs/execwrapper/skaffold"
	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestProfileUnleashOnly(t *testing.T) {
	t.Parallel()
	res, err := skaffoldwrapper.New().E(vr.EnvStaging).P(vr.PartnerManabie).
		Filename("skaffold.backbone.yaml").Profile("unleash-only").
		Diagnose()
	require.NoError(t, err)

	releaseNames := getReleaseNames(res)
	assert.Equal(t, []string{"unleash"}, releaseNames)
}

// TestSkaffold2BackendYAML does basic smoke tests with skaffold v2.
func TestSkaffold2BackendYAML(t *testing.T) {
	t.Parallel()

	res, err := skaffoldwrapper.New().E(vr.EnvProduction).P(vr.PartnerTokyo).
		Filename("skaffold2.backend.yaml").V2Diagnose()
	require.NoError(t, err)
	releaseNames := getHelmReleaseNamesV2(res)
	require.Contains(t, releaseNames, "common")
	require.Contains(t, releaseNames, "draft")
	require.Contains(t, releaseNames, "zeus")

	vr.Iter(t).E(vr.EnvLocal, vr.EnvStaging).IterPE(func(t *testing.T, p vr.P, e vr.E) {
		res, err := skaffoldwrapper.New().E(e).P(p).Filename("skaffold2.backend.yaml").V2RenderRaw()
		require.NoError(t, err)
		require.NotNil(t, res)
	})

	// On production, all services have an additional on-demand node deployment template,
	// which will deploy pods on on-demand nodes. That template should be identical with
	// the spot node deployment template, except for the affinity and tolerations.
	vr.Iter(t).E(vr.EnvProduction).IterPE(func(t *testing.T, p vr.P, e vr.E) {
		manifests, err := skaffoldwrapper.New().E(e).P(p).Filename("skaffold.manaverse.yaml").CachedRender()
		require.NoError(t, err)
		{
			manifests2, err := skaffoldwrapper.New().E(e).P(p).Filename("skaffold2.backend.yaml").V2CachedRender()
			require.NoError(t, err)
			manifests = append(manifests, manifests2...)
		}
		var (
			svcSpotNodeDeployment, svcOnDemandNodeDeployment       *v1.Deployment
			hasuraSpotNodeDeployment, hasuraOnDemandNodeDeployment *v1.Deployment
		)
		for _, m := range manifests {
			if v, ok := m.(*v1.Deployment); ok {
				if !strings.HasPrefix(v.ObjectMeta.Name, "bob") {
					// since all deployments share the same template
					// we need to check for one service only
					// TODO: how about tom service?
					continue
				}
				onDemandNodeDeployment := strings.HasSuffix(v.Name, "-on-demand-node")
				if strings.Contains(v.Name, "hasura") {
					if onDemandNodeDeployment {
						hasuraOnDemandNodeDeployment = v
					} else {
						hasuraSpotNodeDeployment = v
					}
				} else {
					if onDemandNodeDeployment {
						svcOnDemandNodeDeployment = v
					} else {
						svcSpotNodeDeployment = v
					}
				}
			}
		}

		deployments := []struct {
			spot, onDemand *v1.Deployment
		}{
			{svcSpotNodeDeployment, svcOnDemandNodeDeployment},
			{hasuraSpotNodeDeployment, hasuraOnDemandNodeDeployment},
		}
		for _, d := range deployments {
			// verify that service account name, volumes, and containers are identical
			require.Equal(
				t,
				d.spot.Spec.Template.Spec.ServiceAccountName,
				d.onDemand.Spec.Template.Spec.ServiceAccountName,
			)
			require.Equal(
				t,
				d.spot.Spec.Template.Spec.Volumes,
				d.onDemand.Spec.Template.Spec.Volumes,
			)
			require.Equal(
				t,
				d.spot.Spec.Template.Spec.Containers,
				d.onDemand.Spec.Template.Spec.Containers,
			)

			// make sure on-demand node deployment has correct tolerations and affinity
			require.Equal(t, d.onDemand.Spec.Template.Spec.Tolerations, []corev1.Toleration{
				{
					Key:      "backend-on-demand-node",
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectNoSchedule,
				},
			})
			require.Equal(t, d.onDemand.Spec.Template.Spec.Affinity.NodeAffinity, &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "backend-on-demand-node",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"true"},
								},
							},
						},
					},
				},
			})

			// make sure spot node deployment has correct tolerations and affinity
			require.Equal(t, d.spot.Spec.Template.Spec.Tolerations, []corev1.Toleration{
				{
					Key:      "cloud.google.com/gke-spot",
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectNoSchedule,
				},
			})
			require.Equal(t, d.spot.Spec.Template.Spec.Affinity.NodeAffinity, &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "cloud.google.com/gke-spot",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"true"},
								},
							},
						},
					},
				},
			})
		}
	})
}

// TestSkaffoldMonitoringYAML runs `skaffold diagnose` and `skaffold render` for skaffold.monitoring.yaml
// as a sanity check. It helps to ensure monitoring deployment is not broken.
func TestSkaffoldMonitoringYAML(t *testing.T) {
	t.Parallel()

	t.Run("prod.manabie", func(t *testing.T) {
		t.Parallel()
		cmd := skaffoldwrapper.New().Filename("skaffold.monitoring.yaml").Profile("monitoring,manabie")
		res, err := cmd.Diagnose()
		require.NoError(t, err)
		releaseNames := getReleaseNames(res)
		require.Contains(t, releaseNames, "grafana")
		require.Contains(t, releaseNames, "jaeger-all-in-one")
		require.Contains(t, releaseNames, "oncall")
		require.Contains(t, releaseNames, "thanos")

		_, err = cmd.RenderRaw()
		require.NoError(t, err)
	})

	t.Run("stag.manabie", func(t *testing.T) {
		t.Parallel()
		cmd := skaffoldwrapper.New().Filename("skaffold.monitoring.yaml").Profile("monitoring,staging-2")
		res, err := cmd.Diagnose()
		require.NoError(t, err)
		releaseNames := getReleaseNames(res)
		require.Contains(t, releaseNames, "jaeger-all-in-one")
		require.Contains(t, releaseNames, "opentelemetry-collector")
		require.Contains(t, releaseNames, "prometheus")

		_, err = cmd.RenderRaw()
		require.NoError(t, err)
	})
}
