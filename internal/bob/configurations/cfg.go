package configurations

import (
	"github.com/manabie-com/backend/internal/golibs/configs"
)

// Config for Bob
type Config struct {
	Common                    configs.CommonConfig
	Issuers                   []configs.TokenIssuerConfig
	PostgresV2                configs.PostgresConfigV2 `yaml:"postgres_v2"`
	Upload                    configs.UploadConfig
	Storage                   configs.StorageConfig
	Brightcove                configs.BrightcoveConfig
	Whiteboard                configs.WhiteboardConfig
	Agora                     AgoraConfig        `yaml:"agora"`
	AsiaPay                   AsiaPayConfig      `yaml:"asiapay"`
	ClassCodeLength           int                `yaml:"class_code_length"`
	JWTApplicant              string             `yaml:"jwt_applicant"`
	CheckClientVersions       []string           `yaml:"check_client_versions"`
	NotAnsweredQuestionLimit  int                `yaml:"not_answered_question_limit"`
	PromotionCodeLength       int                `yaml:"promo_code_len"`
	PromotionCodePrefixes     []string           `yaml:"promo_code_prefix"`
	GHNProvinceDataFile       string             `yaml:"ghn_province_file"`
	PaymentProcessingDuration *configs.Duration  `yaml:"payment_processing_duration"`
	FakeBrightcoveServer      string             `yaml:"fake_brightcove_server"`
	FakeAppleServer           string             `yaml:"fake_apple_server"`
	CloudConvert              CloudConvertConfig `yaml:"cloud_convert"`
	NatsJS                    configs.NatsJetStreamConfig
	Partner                   configs.PartnerConfig
	FirebaseAPIKey            string                        `yaml:"firebase_api_key"` // used local only
	KeyCloakAuth              configs.KeyCloakAuthConfig    `yaml:"keycloak_auth"`
	TenantsCLIAuth            []configs.TenantCLIAuthConfig `yaml:"tenants_cli_auth"`
	GetPostgresUserKey        string                        `yaml:"get_postgres_user_key"`
	GetPostgresUserPrivateKey string                        `yaml:"get_postgres_user_private_key"`
	UnleashClientConfig       configs.UnleashClientConfig   `yaml:"unleash_client"`
	UnleashLocalAdminAPIKey   string                        `yaml:"unleash_local_admin_api_key"`
}

// CloudConvertConfig for Cloud Convert
type CloudConvertConfig struct {
	Host                string
	Token               string
	ServiceAccountEmail string `yaml:"sa_email"`
	ServiceAccountPK    string `yaml:"sa_pk"`
}

// AgoraConfig for Agora
type AgoraConfig struct {
	AppID                    string `yaml:"app_id"`
	Cert                     string `yaml:"cert"`
	VideTokenSuffix          string `yaml:"video_token_suffix"`
	MaximumLearnerStreamings int    `yaml:"maximum_learner_streamings"`
	CustomerID               string `yaml:"customer_id"`
	CustomerSecret           string `yaml:"customer_secret"`
	Endpoint                 string `yaml:"endpoint"`
	BucketName               string `yaml:"bucket"`
	BucketAccessKey          string `yaml:"bucket_access_key"`
	BucketSecretKey          string `yaml:"bucket_secret_key"`
	CallbackSignature        string `yaml:"callback_signature"`
}

// AsiaPayConfig for AsiaPay
type AsiaPayConfig struct {
	Secret    string
	MerchanID string `yaml:"merchant_id"`
	Currency  string
	Endpoint  string
}
