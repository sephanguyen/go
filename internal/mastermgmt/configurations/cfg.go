package configurations

import (
	"github.com/manabie-com/backend/internal/golibs/configs"
)

// Config for mastermgmt service
type Config struct {
	Common               configs.CommonConfig
	Issuers              []configs.TokenIssuerConfig
	PostgresV2           configs.PostgresConfigV2 `yaml:"postgres_v2"`
	NatsJS               configs.NatsJetStreamConfig
	UnleashClientConfig  configs.UnleashClientConfig `yaml:"unleash_client"`
	CheckClientVersions  []string                    `yaml:"check_client_versions"`
	AppsmithMongoDB      configs.MongoConfig         `yaml:"appsmith_mongodb"` // appsmith mongodb
	AppsmithAPI          configs.AppsmithAPI         `yaml:"appsmith_api"`     // appsmith_api
	ElasticSearch        configs.ElasticSearchConfig
	Zoom                 configs.ZoomConfig
	AppsmithSlackWebHook string `yaml:"appsmith_slack_webhook"`
}
