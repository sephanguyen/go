package tests

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/execwrapper"
	vr "github.com/manabie-com/backend/internal/golibs/variants"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type KafkaConnectProperties struct {
	ElasticUsername string `properties:"elastic_username"`
	ElasticPassword string `properties:"elastic_pwd"`
}

func (k KafkaConnectProperties) Path(p vr.P, e vr.E, _ vr.S) string {
	return filepath.Join(
		execwrapper.RootDirectory(),
		fmt.Sprintf("deployments/helm/platforms/kafka-connect/secrets/%v/%v/kafka-connect.secrets.encrypted.properties", p, e),
	)
}

func TestKafkaConnectElasticProperties(t *testing.T) {
	testfunc := func(t *testing.T, p vr.P, e vr.E) {
		kcp, err := configs.LoadAndDecryptProperties[KafkaConnectProperties](p, e, vr.ServiceKafkaConnect)
		require.NoError(t, err)

		es, err := configs.LoadAndDecrypt[ElasticKibana](p, e, vr.ServiceKafkaConnect)
		require.NoError(t, err)

		assert.Equal(t, es.Username, kcp.ElasticUsername)
		assert.Equal(t, es.Password, kcp.ElasticPassword)
	}
	vr.Iter(t).SkipE(vr.EnvPreproduction).IterPE(testfunc)
}

type KafkaRootCA struct {
	Data string `yaml:"data"`
}

func (KafkaRootCA) Path(p vr.P, e vr.E, _ vr.S) string {
	return filepath.Join(
		execwrapper.RootDirectory(),
		fmt.Sprintf("deployments/helm/platforms/kafka-connect/secrets/%v/%v/root-ca.pem.encrypted.yaml", p, e),
	)
}

type KafkaRootCAKey struct {
	Data string `yaml:"data"`
}

func (KafkaRootCAKey) Path(p vr.P, e vr.E, _ vr.S) string {
	return filepath.Join(
		execwrapper.RootDirectory(),
		fmt.Sprintf("deployments/helm/platforms/kafka-connect/secrets/%v/%v/root-ca-key.pem.encrypted.yaml", p, e),
	)
}

func TestKafkaConnectRootCA(t *testing.T) {
	testfunc := func(t *testing.T, p vr.P, e vr.E) {
		kc, err := configs.LoadAndDecrypt[KafkaRootCA](p, e, vr.ServiceKafkaConnect)
		require.NoError(t, err)

		es, err := configs.LoadAndDecrypt[ElasticRootCA](p, e, vr.ServiceElasticsearch)
		require.NoError(t, err)

		assert.Equal(t, kc.Data, es.Data)
	}
	vr.Iter(t).SkipE(vr.EnvPreproduction).IterPE(testfunc)
}

func TestKafkaConnectRootCAKey(t *testing.T) {
	testfunc := func(t *testing.T, p vr.P, e vr.E) {
		kc, err := configs.LoadAndDecrypt[KafkaRootCAKey](p, e, vr.ServiceKafkaConnect)
		require.NoError(t, err)

		es, err := configs.LoadAndDecrypt[ElasticRootCAKey](p, e, vr.ServiceElasticsearch)
		require.NoError(t, err)

		assert.Equal(t, kc.Data, es.Data)
	}
	vr.Iter(t).SkipE(vr.EnvPreproduction).IterPE(testfunc)
}
