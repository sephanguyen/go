package vr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsServiceEnabled(t *testing.T) {
	t.Parallel()

	// local.manabie should have service draft disabled by default
	status, err := IsServiceEnabled(PartnerManabie, EnvLocal, ServiceDraft)
	require.NoError(t, err)
	require.False(t, status)
}

func TestGetHelmSubchart(t *testing.T) {
	t.Parallel()

	eureka, err := GetHelmSubchart(PartnerGA, EnvProduction, ServiceEureka)
	require.NoError(t, err)
	require.Equal(t, HelmSubchart{MigrationEnabled: true}, *eureka)

	yasuo, err := GetHelmSubchart(PartnerGA, EnvProduction, ServiceYasuo)
	require.NoError(t, err)
	require.Equal(t, HelmSubchart{MigrationEnabled: false}, *yasuo)
}

func TestIsBackendServiceEnabled(t *testing.T) {
	t.Parallel()

	require.NotPanics(t, func() {
		require.True(t, IsBackendServiceEnabled(PartnerManabie, EnvLocal, ServiceZeus))
		require.False(t, IsBackendServiceEnabled(PartnerManabie, EnvLocal, ServiceDraft))
	})
	require.Panics(t, func() {
		_ = IsBackendServiceEnabled(PartnerManabie, EnvLocal, ServiceNotDefined)
	})
}
