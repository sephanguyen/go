package configurations

import "github.com/manabie-com/backend/internal/golibs/configs"

type Config struct {
	Common     configs.CommonConfig
	Kafka      configs.KafkaConfig
	NatsJS     configs.NatsJetStreamConfig
	PostgresV2 configs.PostgresConfigV2 `yaml:"postgres_v2"`

	SlackWebhook *string `yaml:"slack_webhook"`
	SlackUser    string  `yaml:"slack_user"`
	SlackChannel string  `yaml:"slack_channel"`
}

type MigrateConfig struct {
	Common         configs.CommonConfig
	DataLake       configs.PostgresConfigV2 `yaml:"datalake"`
	DataWarehouses configs.PostgresConfigV2 `yaml:"datawarehouses"`
}

func (c Config) GetDBNames() []string {
	res := make([]string, 0)
	for k := range c.PostgresV2.Databases {
		res = append(res, k)
	}
	return res
}
