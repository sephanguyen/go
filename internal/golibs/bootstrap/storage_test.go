package bootstrap

import (
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/stretchr/testify/assert"
)

func TestInitStorage(t *testing.T) {
	t.Parallel()
	t.Run("with storage config", func(t *testing.T) {
		type testConfig struct {
			Storage configs.StorageConfig
		}

		c := testConfig{
			Storage: configs.StorageConfig{},
		}

		rsc := NewResources(WithStorage(&configs.StorageConfig{}))
		err := initStorage(c, rsc)
		assert.NoError(t, err)
	})

	t.Run("without storage config", func(t *testing.T) {
		type testConfig struct{}
		c := testConfig{}
		rsc := NewResources(WithStorage(&configs.StorageConfig{}))
		err := initStorage(c, rsc)
		assert.NoError(t, err)
	})
}
