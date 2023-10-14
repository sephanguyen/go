package tests

import (
	"testing"

	skaffoldwrapper "github.com/manabie-com/backend/internal/golibs/execwrapper/skaffold"
	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNetworkPolicy(t *testing.T) {
	t.Parallel()

	allowIngressFromSameNamespace := func(from []networkingv1.NetworkPolicyPeer) bool {
		for _, v := range from {
			if assert.ObjectsAreEqual(networkingv1.NetworkPolicyPeer{PodSelector: &metav1.LabelSelector{}}, v) {
				return true
			}
		}
		return false
	}

	allowIngressFromNamespace := func(from []networkingv1.NetworkPolicyPeer, ns string) bool {
		for _, v := range from {
			if assert.ObjectsAreEqual(networkingv1.NetworkPolicyPeer{NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"kubernetes.io/metadata.name": ns},
			}}, v) {
				return true
			}
		}
		return false
	}

	testfunc := func(t *testing.T, p vr.P, e vr.E) {
		manifest, err := skaffoldwrapper.New().E(e).P(p).Filename("skaffold2.backend.yaml").V2CachedRender()
		require.NoError(t, err)

		npList := filterKind[networkingv1.NetworkPolicy](manifest)
		if e != vr.EnvLocal && e != vr.EnvStaging {
			require.Empty(t, npList, "Network Policy should not be enabled on UAT or PROD yet")
			return
		}

		for _, p := range npList {
			p := p
			t.Run(p.Name, func(t *testing.T) {
				require.Len(t, p.Spec.Ingress, 1, "there should only be 1 ingress rule")
				require.True(t, allowIngressFromSameNamespace(p.Spec.Ingress[0].From), "rule is missing allowing ingress from pods in the same namespace")
				require.True(t, allowIngressFromNamespace(p.Spec.Ingress[0].From, "monitoring"), "rule is missing allowing ingress from monitoring namespace")
			})
		}
	}

	vr.Iter(t).IterPE(testfunc)
}

func filterKind[T any](manifest []interface{}) []*T {
	res := make([]*T, 0)
	for _, val := range manifest {
		switch v := val.(type) {
		case T:
			res = append(res, &v)
		case *T:
			res = append(res, v)
		}
	}
	return res
}
