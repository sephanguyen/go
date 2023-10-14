package vr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTFServiceDefinitions(t *testing.T) {
	res, err := TFServiceDefinitions(EnvStaging)
	require.NoError(t, err)
	require.Contains(t, res, TFService{Name: "bob"})

	_, err = TFServiceDefinitions(EnvLocal)
	require.EqualError(t, err, "invalid environment: local")
	_, err = TFServiceDefinitions(EnvUAT)
	require.NoError(t, err)
	_, err = TFServiceDefinitions(EnvPreproduction)
	require.EqualError(t, err, "invalid environment: dorp")
	_, err = TFServiceDefinitions(EnvProduction)
	require.NoError(t, err)
}
