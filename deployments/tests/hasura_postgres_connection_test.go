package tests

import (
	"fmt"
	"net/url"
	"regexp"
	"testing"

	skaffoldwrapper "github.com/manabie-com/backend/internal/golibs/execwrapper/skaffold"
	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestHasuraPostgresConnection(t *testing.T) {
	t.Parallel()

	extractDraftHasuraDeployment := func(objects []interface{}, targetName string) (*appsv1.Deployment, error) {
		for _, o := range objects {
			switch v := o.(type) {
			case *appsv1.Deployment:
				if v.ObjectMeta.Name == targetName {
					return v, nil
				}
			}
		}
		return nil, fmt.Errorf("failed to find %s deployment in manifest", targetName)
	}

	extractHasuraContainer := func(in *appsv1.Deployment) (*corev1.Container, error) {
		cs := in.Spec.Template.Spec.Containers
		for _, c := range cs {
			if c.Name == "hasura" {
				return &c, nil
			}
		}
		return nil, fmt.Errorf("failed to find container hasura in deployment")
	}

	extractHasuraPGConnStr := func(in []corev1.EnvVar) (string, error) {
		for _, ev := range in {
			if ev.Name == "HASURA_GRAPHQL_DATABASE_URL" {
				return ev.Value, nil
			}
		}
		return "", fmt.Errorf("failed to find HASURA_GRAPHQL_DATABASE_URL env var in list (full list: %+v)", in)
	}

	reStr := `^postgres://([\w\-\%\.]+)@([\w\d\.]+):(\d+)/(\w+)\?sslmode=disable&application_name=([\w\-\%\.]+)$`
	re := regexp.MustCompile(reStr)

	vr.Iter(t).E(vr.EnvStaging).SkipDisabledServices().IterPES(func(t *testing.T, p vr.P, e vr.E, s vr.S) {
		hasuraEnabled, err := vr.IsHasuraEnabled(p, e, s)
		require.NoError(t, err)
		if !hasuraEnabled {
			return
		}

		manifestObjects, err := skaffoldwrapper.New().E(e).P(p).Filename("skaffold.manaverse.yaml").CachedRender()
		require.NoError(t, err)

		hasuraDeploy, err := extractDraftHasuraDeployment(manifestObjects, s.String()+"-hasura")
		require.NoError(t, err)

		hasuraContainer, err := extractHasuraContainer(hasuraDeploy)
		require.NoError(t, err)

		pgconn, err := extractHasuraPGConnStr(hasuraContainer.Env)
		require.NoError(t, err)

		m := re.FindStringSubmatch(pgconn)
		require.Len(t, m, 6)

		expectedHasuraDBUser := url.QueryEscape(vr.ServiceAccountHasuraDBUser(p, e, s))
		require.Equal(t, expectedHasuraDBUser, m[1])
		require.Equal(t, "127.0.0.1", m[2])
		require.Equal(t, "5432", m[3])
		require.Equal(t, vr.DatabaseName(p, e, s), m[4])
		require.Equal(t, expectedHasuraDBUser, m[5])
	})

	// We moved draft to the backend chart group
	// TODO(@anhpngt) re-implement this for backend group
	// // This block is the same as above, but for draft hasura v2
	// vr.Iter(t).P(vr.PartnerManabie).E(vr.EnvStaging).IterPE(func(t *testing.T, p vr.P, e vr.E) {
	// 	manifestObjects, err := skaffoldwrapper.NewCommand().Env(e.String()).Org(p.String()).Filename("skaffold.manaverse.yaml").CachedRender()
	// 	require.NoError(t, err)

	// 	draftDeploy, err := extractDraftHasuraDeployment(manifestObjects, "draft-hasurav2")
	// 	require.NoError(t, err)

	// 	hasuraContainer, err := extractHasuraContainer(draftDeploy)
	// 	require.NoError(t, err)

	// 	pgconn, err := extractHasuraPGConnStr(hasuraContainer.Env)
	// 	require.NoError(t, err)

	// 	m := re.FindStringSubmatch(pgconn)
	// 	require.Len(t, m, 6)

	// 	expectedHasuraDBUser := url.QueryEscape(vr.ServiceAccountHasuraDBUser(p, e, vr.ServiceDraft))
	// 	require.Equal(t, expectedHasuraDBUser, m[1])
	// 	require.Equal(t, "127.0.0.1", m[2])
	// 	require.Equal(t, "5432", m[3])
	// 	require.Equal(t, vr.DatabaseName(p, e, vr.ServiceDraft), m[4])
	// 	require.Equal(t, expectedHasuraDBUser, m[5])
	// })
}
