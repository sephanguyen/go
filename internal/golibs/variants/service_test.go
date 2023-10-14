package vr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestBackendServices runs a simple test for BackendServices function.
// If this test fails, that means:
//   - A new service is added to deployments/helm/backend but not yet added to this package.
//     In that case, you can add that new service to the serviceToString map.
//   - Services in this tests are disabled/removed unexpectedly. In that case, you can update
//     this test to reflect that change.
func TestBackendServices(t *testing.T) {
	t.Parallel()
	res := BackendServices()
	require.Contains(t, res, ServiceDraft)
	require.Contains(t, res, ServiceZeus)
	require.Contains(t, res, ServiceYasuo)
	require.Contains(t, res, ServiceBob)
}
