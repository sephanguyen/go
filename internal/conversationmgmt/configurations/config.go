package configurations

import (
	"github.com/manabie-com/backend/internal/golibs/configs"
)

type Config struct {
	Common     configs.CommonConfig
	PostgresV2 configs.PostgresConfigV2 `yaml:"postgres_v2"`
	Issuers    []configs.TokenIssuerConfig
	NatsJS     configs.NatsJetStreamConfig
	Agora      configs.AgoraConfig `yaml:"agora"`
}
