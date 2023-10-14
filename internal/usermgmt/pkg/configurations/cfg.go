package configurations

import (
	"github.com/manabie-com/backend/internal/golibs/configs"
)

// Config for usermgmt
type Config struct {
	Common              configs.CommonConfig
	Issuers             []configs.TokenIssuerConfig
	PostgresV2          configs.PostgresConfigV2 `yaml:"postgres_v2"`
	Brightcove          configs.BrightcoveConfig
	NatsJS              configs.NatsJetStreamConfig
	UnleashClientConfig configs.UnleashClientConfig `yaml:"unleash_client"`
	IdentityPlatform    IdentityPlatform            `yaml:"identity_platform"`
	FirebaseAPIKey      string                      `yaml:"firebase_api_key"`
	JWTApplicant        string                      `yaml:"jwt_applicant"`
	OpenAPI             OpenAPI                     `yaml:"open_api"`
	JobAccounts         []JobAccount                `yaml:"job_accounts"`
	WithUsConfig        WithUsConfig                `yaml:"with_us"`
	SlackWebhook        string                      `yaml:"slack_webhook"`
}

type OpenAPI struct {
	AESKey string `yaml:"aes_key"`
	AESIV  string `yaml:"aes_iv"`
}

type IdentityPlatform struct {
	TokenSignerAccountID string `yaml:"token_signer_account_id"`
	ConfigAESKey         string `yaml:"config_aes_key"`
	ConfigAESIv          string `yaml:"config_aes_iv"`
}

type JobAccount struct {
	OrganizationID string `yaml:"organization_id"`
	DomainName     string `yaml:"domain_name"`
	Email          string `yaml:"email"`
	Password       string `yaml:"password"`
}

type WithUsConfig struct {
	BucketName       string `yaml:"bucket_name"`
	SlackChannel     string `yaml:"slack_channel"`
	WithusChannel    string `yaml:"withus_channel"`
	WithusWebhookURL string `yaml:"withus_webhook_url"`
}

func (config Config) GetMultiTenantProjectID() string {
	identityPlatformProjectID := config.Common.IdentityPlatformProject

	// Because JPREP doesn't have firebase project nor identity platform project,
	// so just return google cloud project to avoid crash at initialization
	if identityPlatformProjectID == "" {
		identityPlatformProjectID = config.Common.GoogleCloudProject
	}

	return identityPlatformProjectID
}

func (config Config) MultiTenantConfig() MultiTenantConfig {
	// No service account id
	identityManagerConfig := MultiTenantConfig{
		projectID: config.GetMultiTenantProjectID(),
	}
	return identityManagerConfig
}

func (config Config) MultiTenantTokenSignerConfig() MultiTenantConfig {
	// Use custom service account id
	identityManagerConfig := MultiTenantConfig{
		projectID:        config.GetMultiTenantProjectID(),
		serviceAccountID: config.IdentityPlatform.TokenSignerAccountID,
	}
	return identityManagerConfig
}
