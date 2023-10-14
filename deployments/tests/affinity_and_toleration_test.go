package tests

import (
	"errors"
	"fmt"
	"testing"

	skaffoldwrapper "github.com/manabie-com/backend/internal/golibs/execwrapper/skaffold"
	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type scheduledK8sObject struct {
	name        string
	affinity    *corev1.Affinity
	tolerations []corev1.Toleration
}

var ignoredObjectKindErr = errors.New("objects of this kind are ignored")

func extractAffinityAndToleration(o interface{}) (*scheduledK8sObject, error) {
	switch v := o.(type) {
	case *appsv1.Deployment:
		return &scheduledK8sObject{
			name:        v.ObjectMeta.Name,
			affinity:    v.Spec.Template.Spec.Affinity,
			tolerations: v.Spec.Template.Spec.Tolerations,
		}, nil
	case *appsv1.StatefulSet:
		return &scheduledK8sObject{
			name:        v.ObjectMeta.Name,
			affinity:    v.Spec.Template.Spec.Affinity,
			tolerations: v.Spec.Template.Spec.Tolerations,
		}, nil
	default:
		return nil, ignoredObjectKindErr
	}
}

func TestPreproductionAffinityAndToleration(t *testing.T) {
	e := vr.EnvPreproduction
	pList := []vr.P{vr.PartnerAIC, vr.PartnerGA, vr.PartnerJPREP, vr.PartnerRenseikai, vr.PartnerSynersia, vr.PartnerTokyo}
	fList := []string{"skaffold.backbone.yaml", "skaffold.manaverse.yaml", "skaffold.cp-ksql-server.yaml"}
	for _, p := range pList {
		for _, skfile := range fList {
			skfile := skfile
			t.Run(fmt.Sprintf("%v.%v.%v", e, p, skfile), func(t *testing.T) {
				t.Parallel()

				manifestObjects, err := skaffoldwrapper.New().E(e).P(p).Filename(skfile).CachedRender()
				require.NoError(t, err)
				for _, v := range manifestObjects {
					actual, err := extractAffinityAndToleration(v)
					if errors.Is(err, ignoredObjectKindErr) {
						continue
					}
					assertPreproductionAffinityAndToleration(t, p, e, actual)
				}
			})
		}
	}
}

// assertPreproductionAffinityAndToleration asserts that both affinity and toleration should
// make the resource's pods go to preproduction node.
func assertPreproductionAffinityAndToleration(t *testing.T, p vr.P, e vr.E, d *scheduledK8sObject) {
	expectedNodeAffinity := &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{{
				MatchExpressions: []corev1.NodeSelectorRequirement{
					{Key: "cloud.google.com/gke-spot", Operator: "In", Values: []string{"true"}},
					{Key: "environment", Operator: "In", Values: []string{"preproduction"}},
				},
			}},
		},
	}
	require.NotNil(t, d.affinity,
		"deployment %q should not have nil affinity in %v.%v", d.name, e, p)
	assert.Equal(t, expectedNodeAffinity, d.affinity.NodeAffinity,
		"deployment %q has unexpected nodeAffinity in %v.%v", d.name, e, p)

	expectedToleration := []corev1.Toleration{
		{Key: "cloud.google.com/gke-spot", Operator: "Exists", Effect: "NoSchedule"},
		{Key: "environment", Operator: "Equal", Value: "preproduction", Effect: "NoSchedule"},
	}
	assert.Equal(t, expectedToleration, d.tolerations,
		"deployment %q has unexpected tolerations in %v.%v", d.name, e, p)
}

func TestAffinityAndToleration(t *testing.T) {
	for _, v := range []struct {
		e vr.E
		p vr.P
	}{
		{vr.EnvLocal, vr.PartnerManabie},
		// {vr.EnvStaging, vr.PartnerManabie},
		// {vr.EnvStaging, vr.PartnerJPREP},
		// {vr.EnvUAT, vr.PartnerManabie},
		// {vr.EnvUAT, vr.PartnerJPREP},
		// {vr.EnvProduction, vr.PartnerAIC},
		// {vr.EnvProduction, vr.PartnerGA},
		// {vr.EnvProduction, vr.PartnerJPREP},
		// {vr.EnvProduction, vr.PartnerRenseikai},
		// {vr.EnvProduction, vr.PartnerSynersia},
		// {vr.EnvProduction, vr.PartnerTokyo},
	} {
		t.Run(fmt.Sprintf("%v.%v.manaverse", v.e, v.p), func(t *testing.T) { testAffinityAndTolerationManaverse(t, v.p, v.e) })
		t.Run(fmt.Sprintf("%v.%v.backbone", v.e, v.p), func(t *testing.T) { testAffinityAndTolerationBackbone(t, v.p, v.e) })
	}
}

func expectedPodAntiAffinity(k8sObject *scheduledK8sObject, e vr.E) *corev1.PodAntiAffinity {
	switch e {
	default:
		return &corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight: 100,
					PodAffinityTerm: corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app.kubernetes.io/name": k8sObject.name,
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		}
	case vr.EnvProduction:
		return &corev1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
				{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app.kubernetes.io/name": k8sObject.name,
						},
					},
					TopologyKey: "kubernetes.io/hostname",
				},
			},
		}
	}
}

func testAffinityAndTolerationManaverse(t *testing.T, p vr.P, e vr.E) {
	t.Parallel()

	manifestObjects, err := skaffoldwrapper.New().E(e).P(p).Filename("skaffold.manaverse.yaml").CachedRender()
	require.NoError(t, err)

	for _, v := range manifestObjects {
		actual, err := extractAffinityAndToleration(v)
		if errors.Is(err, ignoredObjectKindErr) {
			continue
		}

		switch actual.name {
		// Jerry is a caching services, currently in testing
		case "jerry":
			expectedAffinity := &corev1.Affinity{
				NodeAffinity:    &corev1.NodeAffinity{},
				PodAntiAffinity: &corev1.PodAntiAffinity{},
			}
			assert.Equal(t, expectedAffinity, actual.affinity,
				"deployment %q has unexpected affinity in %v.%v", actual.name, e, p)
			assert.Empty(t, actual.tolerations,
				"deployment %q has unexpected tolerations in %v.%v", actual.name, e, p)

		// Mission-critical service, should not tolerate spot nodes, but still have podAntiAffinity.
		case "eureka-jprep-sync-course-student", "eureka-all-consumers", "unleash", "tom":
			expectedAffinity := &corev1.Affinity{
				NodeAffinity:    &corev1.NodeAffinity{},
				PodAntiAffinity: expectedPodAntiAffinity(actual, e),
			}
			assert.Equal(t, expectedAffinity, actual.affinity,
				"deployment %q has unexpected affinity in %v.%v", actual.name, e, p)
			assert.Empty(t, actual.tolerations,
				"deployment %q has unexpected tolerations in %v.%v", actual.name, e, p)

		// Any other deployment/statefulset should go to spot nodes.
		default:
			// affinity check
			// note that in local, we set "preferred", not "required"
			// though we can even set it to nil, though I don't want to change
			// things to drastically for now.
			expectedNodeAffinity := &corev1.NodeAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
					{
						Weight:     10,
						Preference: corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{{Key: "cloud.google.com/gke-spot", Operator: "In", Values: []string{"true"}}}},
					},
				},
			}

			switch e {
			case vr.EnvLocal:
			case vr.EnvUAT:
				expectedNodeAffinity = &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{NodeSelectorTerms: []corev1.NodeSelectorTerm{{MatchExpressions: []corev1.NodeSelectorRequirement{{Key: "cloud.google.com/gke-spot", Operator: "In", Values: []string{"true"}}, {Key: "n2d-highmem-2-uat-spot", Operator: "In", Values: []string{"true"}}}}}},
				}
			default:
				expectedNodeAffinity = &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{NodeSelectorTerms: []corev1.NodeSelectorTerm{{MatchExpressions: []corev1.NodeSelectorRequirement{{Key: "cloud.google.com/gke-spot", Operator: "In", Values: []string{"true"}}}}}},
				}
			}
			expectedAffinity := &corev1.Affinity{
				NodeAffinity:    expectedNodeAffinity,
				PodAffinity:     nil,
				PodAntiAffinity: expectedPodAntiAffinity(actual, e),
			}
			assert.Equal(t, expectedAffinity, actual.affinity,
				"deployment %q has unexpected affinity in %v.%v", actual.name, e, p)

			// toleration check
			expectedTolerations := []corev1.Toleration{
				{Key: "cloud.google.com/gke-spot", Operator: "Exists", Effect: "NoSchedule"},
			}
			switch e {
			case vr.EnvUAT:
				expectedTolerations = []corev1.Toleration{
					{Key: "cloud.google.com/gke-spot", Operator: "Exists", Effect: "NoSchedule"},
					{Key: "n2d-highmem-2-uat-spot", Operator: "Exists", Effect: "NoSchedule"},
				}
			default:
			}
			assert.Equal(t, expectedTolerations, actual.tolerations,
				"deployment %q has unexpected tolerations in %v.%v", actual.name, e, p)
		}
	}
}

func testAffinityAndTolerationBackbone(t *testing.T, p vr.P, e vr.E) {
	t.Parallel()

	manifestObjects, err := skaffoldwrapper.New().E(e).P(p).Filename("skaffold.backbone.yaml").CachedRender()
	require.NoError(t, err)

	manifestObjects2, err := skaffoldwrapper.New().E(e).P(p).Filename("skaffold.cp-ksql-server.yaml").CachedRender()
	require.NoError(t, err)
	manifestObjects = append(manifestObjects, manifestObjects2...)

	for _, v := range manifestObjects {
		actual, err := extractAffinityAndToleration(v)
		if errors.Is(err, ignoredObjectKindErr) {
			continue
		}

		switch actual.name {
		// Mission-critical service, should not tolerate spot nodes, but still have podAntiAffinity.
		case "nats-jetstream", "elasticsearch-elastic", "kafka",
			"kibana-elastic", "es-exporter-elastic", "kafka-exporter-kafka",
			"cp-schema-registry", "unleash", "cp-ksql-server":
			expectedAffinity := &corev1.Affinity{
				PodAntiAffinity: &corev1.PodAntiAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
						{
							Weight: 1,
							PodAffinityTerm: corev1.PodAffinityTerm{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{
										{
											Key:      "app.kubernetes.io/name",
											Operator: "In",
											Values:   []string{actual.name},
										},
									},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
				},
			}

			assert.Equal(t, expectedAffinity, actual.affinity,
				"deployment %q has unexpected affinity in %v.%v", actual.name, e, p)
			assert.Empty(t, actual.tolerations, "deployment %q has unexpected tolerations in %v.%v", actual.name, e, p)

		case "kafka-connect":
			if (p == vr.PartnerManabie && e == vr.EnvStaging) || (p == vr.PartnerManabie && e == vr.EnvUAT) {
				expectedAffinity := &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
							{
								Weight: 10,
								Preference: corev1.NodeSelectorTerm{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      "cloud.google.com/gke-nodepool",
											Operator: "In",
											Values:   []string{"n2d-highmem-2-on-demand"},
										},
									},
								},
							},
						},
					},
					PodAntiAffinity: &corev1.PodAntiAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
							{
								Weight: 100,
								PodAffinityTerm: corev1.PodAffinityTerm{
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app.kubernetes.io/name": "kafka-connect",
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
				}
				// toleration check
				expectedTolerations := []corev1.Toleration{
					{Key: "cloud.google.com/gke-spot", Operator: "Exists", Effect: "NoSchedule"},
				}
				assert.Equal(t, expectedAffinity, actual.affinity,
					"deployment %q has unexpected affinity in %v.%v", actual.name, e, p)
				assert.Equal(t, expectedTolerations, actual.tolerations, "deployment %q has unexpected tolerations in %v.%v", actual.name, e, p)
			} else {
				expectedAffinity := &corev1.Affinity{
					PodAntiAffinity: &corev1.PodAntiAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
							{
								Weight: 1,
								PodAffinityTerm: corev1.PodAffinityTerm{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app.kubernetes.io/name",
												Operator: "In",
												Values:   []string{actual.name},
											},
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
				}
				assert.Equal(t, expectedAffinity, actual.affinity,
					"deployment %q has unexpected affinity in %v.%v", actual.name, e, p)
				assert.Empty(t, actual.tolerations, "deployment %q has unexpected tolerations in %v.%v", actual.name, e, p)
			}

		// The rest can be on spot instances.
		// Note that in local, we "prefer", not "require", scheduling on spot instances.
		default:
			// affinity check
			expectedNodeAffinity := &corev1.NodeAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
					{
						Weight:     10,
						Preference: corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{{Key: "cloud.google.com/gke-spot", Operator: "In", Values: []string{"true"}}}},
					},
				},
			}
			if e != vr.EnvLocal {
				expectedNodeAffinity = &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{NodeSelectorTerms: []corev1.NodeSelectorTerm{{MatchExpressions: []corev1.NodeSelectorRequirement{{Key: "cloud.google.com/gke-spot", Operator: "In", Values: []string{"true"}}}}}},
				}
			}
			expectedAffinity := &corev1.Affinity{
				NodeAffinity: expectedNodeAffinity,
				PodAntiAffinity: &corev1.PodAntiAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
						{
							Weight: 100,
							PodAffinityTerm: corev1.PodAffinityTerm{
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"app.kubernetes.io/name": actual.name,
									},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
				},
			}
			assert.Equal(t, expectedAffinity, actual.affinity,
				"deployment %q has unexpected affinity in %v.%v", actual.name, e, p)

			// toleration check
			expectedTolerations := []corev1.Toleration{
				{Key: "cloud.google.com/gke-spot", Operator: "Exists", Effect: "NoSchedule"},
			}
			assert.Equal(t, expectedTolerations, actual.tolerations,
				"deployment %q has unexpected tolerations in %v.%v", actual.name, e, p)
		}
	}
}
