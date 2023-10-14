package configurations

import (
	"github.com/manabie-com/backend/internal/golibs/configs"
)

type Config struct {
	Common              configs.CommonConfig
	Issuers             []configs.TokenIssuerConfig
	PostgresV2          configs.PostgresConfigV2    `yaml:"postgres_v2"`
	UnleashClientConfig configs.UnleashClientConfig `yaml:"unleash_client"`
}
