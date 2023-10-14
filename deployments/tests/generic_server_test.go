package tests

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	skaffoldwrapper "github.com/manabie-com/backend/internal/golibs/execwrapper/skaffold"
	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestAllDeploymentsMustUseGenericServer(t *testing.T) {
	testAllDeploymentsMustUseGenericServer := func(t *testing.T, p vr.P, e vr.E) {
		manifestObjects, err := skaffoldwrapper.New().E(e).P(p).Filename("skaffold2.backend.yaml").V2CachedRender()
		require.NoError(t, err)

		for _, v := range manifestObjects {
			name, c, err := extractServiceContainer(v)
			if errors.Is(err, ignoredObjectKindErr) {
				continue
			} else {
				require.NoError(t, err)
			}

			// if local, gserver is specified in "command"; else, in "args"
			t.Run(name, func(t *testing.T) {
				if e == vr.EnvLocal {
					require.GreaterOrEqual(t, len(c.Command), 3, "\"command\" must contain at least 3 elements")
					bashCommand := c.Command[2]
					assert.True(t, gserverRe.MatchString(bashCommand), "bash command does not contain \"gserver\"")
				} else {
					require.GreaterOrEqual(t, len(c.Args), 1, "\"args\" must contain at least 1 element")
					assert.Equal(t, "gserver", c.Args[0], "args[0] should be \"gserver\"")
				}
			})
		}
	}

	vr.Iter(t).IterPE(testAllDeploymentsMustUseGenericServer)
}

var gserverRe = regexp.MustCompile(`\bgserver\b`)

// ignoredGserverServices contains strings that are matched using strings.Contain()
// to all deployments. If matched, that deployment is ignored from this test.
var ignoredGserverServices = []string{
	"hasura", "eureka-", "gandalf-", "yasuo",
	"teacher-web", "learner-web", "backoffice", "unleash",
	"-caching-redis",
}

func extractServiceContainer(o interface{}) (string, *corev1.Container, error) {
	var name string
	var containers []corev1.Container
	switch v := o.(type) {
	case *appsv1.Deployment:
		name = v.Name
		containers = v.Spec.Template.Spec.Containers
	case *appsv1.StatefulSet:
		name = v.Name
		containers = v.Spec.Template.Spec.Containers
	default:
		return "", nil, ignoredObjectKindErr
	}

	// skip unwanted
	for _, ignored := range ignoredGserverServices {
		if strings.Contains(name, ignored) {
			return "", nil, ignoredObjectKindErr
		}
	}

	for _, c := range containers {
		// Usually that container name is the same as deployment/sts name
		// In on-demand pod cases, they are equal when appended "-on-demand-node".
		if c.Name != name && c.Name+"-on-demand-node" != name {
			continue
		}
		return name, &c, nil
	}

	panic(fmt.Errorf("unhandled object: %+v", o))
}
