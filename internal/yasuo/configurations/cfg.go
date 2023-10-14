package configurations

import (
	"github.com/manabie-com/backend/internal/golibs/configs"
)

// Config for Yasuo
type Config struct {
	Common                     configs.CommonConfig
	Issuers                    []configs.TokenIssuerConfig
	Brightcove                 configs.BrightcoveConfig
	PostgresV2                 configs.PostgresConfigV2 `yaml:"postgres_v2"`
	Storage                    configs.StorageConfig
	Upload                     configs.UploadConfig
	Whiteboard                 configs.WhiteboardConfig
	FirebaseAPIKey             string `yaml:"firebase_api_key"`
	ClassCodeLength            int    `yaml:"class_code_length"`
	JWTApplicant               string `yaml:"jwt_applicant"`
	QuestionBucket             string `yaml:"question_bucket"`
	FakeBrightcoveServer       string `yaml:"fake_brightcove_server"`
	QuestionPublishedTopic     string `yaml:"question_published_topic"`
	QuestionRenderedSubscriber string `yaml:"question_rendered_sub"`
	BobHasuraAdminURL          string `yaml:"bob_hasura_admin_url"`
	ElasticSearch              configs.ElasticSearchConfig
	NatsJS                     configs.NatsJetStreamConfig
	KeyCloakAuth               configs.KeyCloakAuthConfig    `yaml:"keycloak_auth"`
	TenantsCLIAuth             []configs.TenantCLIAuthConfig `yaml:"tenants_cli_auth"`
	TraceEnabled               bool                          `yaml:"trace_enabled"`
	UnleashClientConfig        configs.UnleashClientConfig   `yaml:"unleash_client"`
}
