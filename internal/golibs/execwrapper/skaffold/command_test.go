package skaffoldwrapper

import (
	"testing"

	"github.com/manabie-com/backend/internal/golibs/execwrapper"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
)

// TestCommand_Render tests Command.Render function.
// It should NOT be run in parallel with other tests.
func TestCommand_Render(t *testing.T) {
	oldRenderScriptPath := renderScriptPath
	renderScriptPath = execwrapper.Abs("internal/golibs/execwrapper/skaffold/testdata/gen_skaffold_render.sh")
	t.Cleanup(func() { renderScriptPath = oldRenderScriptPath })

	c := Command{}
	objects, err := c.Render()
	require.NoError(t, err)
	assert.NotNil(t, objects)

	for _, o := range objects {
		switch v := o.(type) {
		case *corev1.LimitRange:
			assert.Equal(t, "cpu-limit-range", v.ObjectMeta.Name)
		case *policyv1beta1.PodDisruptionBudget:
			assert.Equal(t, "draft", v.ObjectMeta.Name)
			assert.Equal(t, 1, v.Spec.MaxUnavailable.IntValue())
		case *corev1.ServiceAccount:
			assert.Equal(t, "local-draft", v.ObjectMeta.Name)
		case *corev1.Secret:
			assert.Equal(t, corev1.SecretType("Opaque"), v.Type)
		case *corev1.ConfigMap:
			assert.Contains(t, v.Data, "draft.common.config.yaml")
		case *rbacv1.ClusterRole:
			assert.Contains(t, v.Name, "local-manabie-tester-cluster-role")
		case *rbacv1.ClusterRoleBinding:
			assert.Contains(t, v.Name, "local-manabie-tester-cluster-role-binding")
		case *corev1.Service:
			assert.Equal(t, corev1.ServiceType("ClusterIP"), v.Spec.Type)
		case *appsv1.Deployment:
			assert.Equal(t, "local-draft", v.Spec.Template.Spec.ServiceAccountName)
		case *istionetworkingv1beta1.VirtualService:
			assert.Equal(t, "draft-api", v.ObjectMeta.Name)
		default:
			t.Fatalf("invalid type %T for object %v", v, v)
		}
	}
}

// TestCommand_CacheRender tests Command.CacheRender function.
// It should NOT be run in parallel with other tests.
func TestCommand_CacheRender(t *testing.T) {
	key1 := Command{es: EnvSet{Env: "local"}, fs: FlagSet{F: "skaffold.manaverse.yaml"}}
	key2 := Command{es: EnvSet{Env: "local"}, fs: FlagSet{F: "skaffold.manaverse.yaml"}}
	key3 := Command{es: EnvSet{Env: "stag"}, fs: FlagSet{F: "skaffold.manaverse.yaml"}}

	// reset the cache, in case some other tests have filled it up
	// therefore, this test should NEVER be run with t.Parallel()
	savedRenderCache := globalRenderCache.cache
	globalRenderCache.cache = make(map[Command]renderResult)
	defer func() { globalRenderCache.cache = savedRenderCache }() // reset the cache to the original state

	// pre-cache key1
	t.Run("precache key1", func(t *testing.T) {
		require.Empty(t, globalRenderCache.cache, "globalRenderCache.cache should have no elements, since we have not used it")
		_, err, exists := key1.cachedRender()
		require.NoError(t, err)
		require.False(t, exists, "unexpected cache hit")
		require.Len(t, globalRenderCache.cache, 1, "globalRenderCache.cache map should have 1 element (from key1)")
	})

	t.Run("key2 == key1, thus key2 should have cache", func(t *testing.T) {
		require.Equal(t, key1, key2)
		_, err, exists := key2.cachedRender()
		require.NoError(t, err)
		require.True(t, exists, "unexpected cache miss")
		require.Len(t, globalRenderCache.cache, 1, "globalRenderCache.cache map should have 1 element (from key1)")
	})

	t.Run("key3 != key1, thus key3 should NOT have cache", func(t *testing.T) {
		require.NotEqual(t, key1, key3)
		_, err, exists := key3.cachedRender()
		require.NoError(t, err)
		require.False(t, exists, "unexpected cache hit")
		require.Len(t, globalRenderCache.cache, 2, "globalRenderCache.cache map should have 2 elements (from key1 and key3)")
	})
}
