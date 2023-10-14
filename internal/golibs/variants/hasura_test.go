package vr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsHasuraEnabled(t *testing.T) {
	t.Parallel()

	b, err := IsHasuraEnabled(PartnerTokyo, EnvProduction, ServiceEnigma)
	require.NoError(t, err)
	require.False(t, b)
}
