package tests

import (
	"testing"

	skaffoldwrapper "github.com/manabie-com/backend/internal/golibs/execwrapper/skaffold"
	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileNATSOnly(t *testing.T) {
	t.Parallel()
	res, err := skaffoldwrapper.New().E(vr.EnvStaging).P(vr.PartnerManabie).
		Filename("skaffold.backbone.yaml").Profile("nats-only").
		Diagnose()
	require.NoError(t, err)

	releaseNames := getReleaseNames(res)
	assert.Equal(t, []string{"nats-jetstream"}, releaseNames)
}
