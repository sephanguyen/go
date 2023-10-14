package configurations

import "github.com/manabie-com/backend/internal/golibs/configs"

type EmailWebhook struct {
	ReceiveFromAllTenant  bool     `yaml:"receive_from_all_tenant"`
	ReceiveOnlyFronTenant []string `yaml:"receive_only_from_tenant"`
}

type Config struct {
	Common             configs.CommonConfig
	PostgresV2         configs.PostgresConfigV2 `yaml:"postgres_v2"`
	Issuers            []configs.TokenIssuerConfig
	NatsJS             configs.NatsJetStreamConfig
	KafkaCluster       configs.KafkaClusterConfig `yaml:"kafka_cluster"`
	SendGrid           configs.SendGridConfig     `yaml:"sendgrid"`
	EmailWebhookConfig EmailWebhook               `yaml:"email_webhook"`
}
