package tests

import (
	"fmt"
	"testing"

	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/stretchr/testify/require"
)

// TestTFServiceDefsYAML ensures that when a service is enabled in an
// environment (stag, uat, ...) then it should have proper corresponding
// terraform service definitions.
func TestTFServiceDefsYAML(t *testing.T) {
	t.Parallel()

	checkstatus := func(p vr.P, e vr.E, s vr.S) error {
		serviceDefs, err := vr.TFServiceDefinitions(e)
		if err != nil {
			return fmt.Errorf("failed to get service def: %s", err)
		}

		if !vr.IsBackendServiceEnabled(p, e, s) {
			// if not enabled, we ignore even if the service is defined in terraform
			return nil
		}

		svcName := s.String()
		for _, v := range serviceDefs {
			if v.Name == svcName {
				return nil
			}
		}
		return fmt.Errorf("service %v is enabled in %v.%v, but not defined in %s-defs.yaml", s, e, p, e)
	}

	vr.Iter(t).SkipE(vr.EnvLocal, vr.EnvPreproduction).
		IterPES(func(t *testing.T, p vr.P, e vr.E, s vr.S) { require.NoError(t, checkstatus(p, e, s)) })
}
