package configurations

import (
	"github.com/manabie-com/backend/internal/golibs/configs"
)

// Config for Fatima
type Config struct {
	Common       configs.CommonConfig
	Issuers      []configs.TokenIssuerConfig
	PostgresV2   configs.PostgresConfigV2 `yaml:"postgres_v2"`
	NatsJS       configs.NatsJetStreamConfig
	JWTApplicant string `yaml:"jwt_applicant"`
}
