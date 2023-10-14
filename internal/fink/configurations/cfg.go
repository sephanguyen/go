package configurations

import "github.com/manabie-com/backend/internal/golibs/configs"

type Config struct {
	Common       configs.CommonConfig
	NatsJS       configs.NatsJetStreamConfig
	KafkaCluster configs.KafkaClusterConfig `yaml:"kafka_cluster"`
}
