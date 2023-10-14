package configurations

import (
	"github.com/manabie-com/backend/internal/golibs/configs"
)

// Config for tom
type Config struct {
	Common              configs.CommonConfig
	Issuers             []configs.TokenIssuerConfig
	PostgresV2          configs.PostgresConfigV2 `yaml:"postgres_v2"`
	BobDBConnection     string                   `yaml:"bob_db_connection"`
	FirebaseKey         string                   `yaml:"firebase_key"`
	MaxCacheEntry       int                      `yaml:"max_cache_entry"`
	ElasticSearch       configs.ElasticSearchConfig
	NatsJS              configs.NatsJetStreamConfig
	KeyCloakAuth        configs.KeyCloakAuthConfig    `yaml:"keycloak_auth"`
	TenantsCLIAuth      []configs.TenantCLIAuthConfig `yaml:"tenants_cli_auth"`
	GrpcWebAddress      string                        `yaml:"grpc_web_addr"`
	UnleashClientConfig configs.UnleashClientConfig   `yaml:"unleash_client"`
}
