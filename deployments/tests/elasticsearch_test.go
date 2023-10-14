package tests

import (
	"fmt"
	"testing"

	skaffoldwrapper "github.com/manabie-com/backend/internal/golibs/execwrapper/skaffold"
	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
)

// TestElasticSearchShamirAddress ensures that the `openid_connect_url` points
// to our shamir service correctly.
//
// See: https://manabie.atlassian.net/browse/LT-46866
//
// Note: we can also implement this using integration test inside k8s, but that would
// not test for non-local environments. This test, on the other hand, is more of a
// unit test and might not cover all edge-cases.
func TestElasticSearchShamirAddress(t *testing.T) {
	t.Parallel()

	type elasticConfig struct {
		Config struct {
			Dynamic struct {
				AuthC struct {
					OpenIDAuthDomain struct {
						HTTPAuthenticator struct {
							Config struct {
								OpenIDConnectURL string `yaml:"openid_connect_url"`
							} `yaml:"config"`
						} `yaml:"http_authenticator"`
					} `yaml:"openid_auth_domain"`
				} `yaml:"authc"`
			} `yaml:"dynamic"`
		} `yaml:"config"`
	}

	// lookupESConfigMap retrieves the configmap of elasticsearch from the rendered manifest
	lookupESConfigMap := func(manifests []interface{}) (*corev1.ConfigMap, error) {
		const elasticCMName = "elasticsearch-elastic"
		for _, v := range manifests {
			switch o := v.(type) {
			case *corev1.ConfigMap:
				if o.ObjectMeta.Name == elasticCMName {
					return o, nil
				}
			}
		}
		return nil, fmt.Errorf("failed to find any configmap name %q in generated manifest", elasticCMName)
	}

	getOpenIDConnectURL := func(cm *corev1.ConfigMap) (string, error) {
		wantedKey := "config.yml"
		data, ok := cm.Data[wantedKey]
		if !ok {
			return "", fmt.Errorf("elasticsearch configmap does not contain key %q", wantedKey)
		}
		c := elasticConfig{}
		if err := yaml.Unmarshal([]byte(data), &c); err != nil {
			return "", fmt.Errorf("failed to unmarshal: %s", err)
		}
		res := c.Config.Dynamic.AuthC.OpenIDAuthDomain.HTTPAuthenticator.Config.OpenIDConnectURL
		if res == "" {
			return "", fmt.Errorf("`openid_connect_url` is missing in %s", wantedKey)
		}
		return res, nil
	}

	getShamirNamespace := func(p vr.P, e vr.E) (string, error) {
		configs, err := skaffoldwrapper.New().E(e).P(p).Filename("skaffold2.backend.yaml").V2Diagnose()
		if err != nil {
			return "", fmt.Errorf("failed to run skaffold diagnose: %s", err)
		}
		if len(configs) != 1 {
			return "", fmt.Errorf("expected 1 SkaffoldConfig from skaffold diagnose, got %d", len(configs))
		}
		config := configs[0]
		const expectedChartName = "shamir"
		for _, v := range config.Deploy.LegacyHelmDeploy.Releases {
			if v.Name != expectedChartName {
				continue
			}
			res := v.Namespace
			if v.Namespace == "" {
				return "", fmt.Errorf("namespace of release %q is missing", expectedChartName)
			}
			return res, nil
		}

		return "", fmt.Errorf("failed to find any release name %q from skaffold diagnose", expectedChartName)
	}

	// main test function
	testfunc := func(t *testing.T, p vr.P, e vr.E) {
		// get elasticsearch config
		esManifests, err := skaffoldwrapper.New().E(e).P(p).Filename("skaffold2.backbone.yaml").V2CachedRender()
		require.NoError(t, err)
		esConfigmap, err := lookupESConfigMap(esManifests)
		require.NoError(t, err)
		openIDConnectURL, err := getOpenIDConnectURL(esConfigmap)
		require.NoError(t, err)

		// get shamir namespace
		shamirNamespace, err := getShamirNamespace(p, e)
		require.NoError(t, err)

		// Note that the following namespaces are correct:
		//	- "shamir.<shamir-namespace>.svc"
		//	- "shamir.<shamir-namespace>.svc.cluster.local"
		// So if you use either namespace, update this test accordingly.
		expectedOpenIDConnectURL := fmt.Sprintf("http://shamir.%s.svc.cluster.local:5680/oidc/.well-known/openid-configuration", shamirNamespace)
		require.Equal(t, expectedOpenIDConnectURL, openIDConnectURL)
	}

	vr.Iter(t).IterPE(testfunc)
}
