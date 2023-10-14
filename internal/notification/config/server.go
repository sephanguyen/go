package config

import "github.com/manabie-com/backend/internal/golibs/configs"

type ScheduledNotificationConfig struct {
	IsRunningForAllTenant bool     `yaml:"is_running_for_all_tenants"`
	TenantIDs             []string `yaml:"tenant_ids"`
}

type Config struct {
	Common                      configs.CommonConfig
	Issuers                     []configs.TokenIssuerConfig
	PostgresV2                  configs.PostgresConfigV2 `yaml:"postgres_v2"`
	NatsJS                      configs.NatsJetStreamConfig
	KafkaCluster                configs.KafkaClusterConfig `yaml:"kafka_cluster"`
	Storage                     configs.StorageConfig
	ScheduledNotificationConfig ScheduledNotificationConfig `yaml:"scheduled_notification"`
}
