package configurations

import (
	bobConfigs "github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
)

type postgresMigrateConfig struct {
	Database configs.PostgresDatabaseConfig `yaml:"database"`
}

// Config to parse Gandalf's yaml config file.
type Config struct {
	Common                      configs.CommonConfig `yaml:"common"`
	Issuers                     []configs.TokenIssuerConfig
	PostgresV2                  configs.PostgresConfigV2 `yaml:"postgres_v2"`
	PostgresMigrate             postgresMigrateConfig    `yaml:"postgres_migrate"`
	VirtualClassroomHTTPSrvAddr string                   `yaml:"virtualclassroom_http_srv_addr"`
	MasterMgmtHTTPSrvAddr       string                   `yaml:"mastermgmt_http_srv_addr"`
	UserMgmtRestAddr            string                   `yaml:"usermgmt_rest_addr"`
	EnigmaSrvAddr               string                   `yaml:"enigma_srv_addr"`
	JPREPSignatureSecret        string                   `yaml:"jprep_signature_secret"`
	AgoraSignatureSecret        string                   `yaml:"agora_signature_secret"`
	FirebaseAPIKey              string                   `yaml:"firebase_api_key"`
	UnleashAPIKey               string                   `yaml:"unleash_api_key"`
	UnleashLocalAdminAPIKey     string                   `yaml:"unleash_local_admin_api_key"`
	Storage                     configs.StorageConfig    `yaml:"storage"`
	Brightcove                  configs.BrightcoveConfig
	KafkaCluster                configs.KafkaClusterConfig `yaml:"kafka_cluster"`
	NatsJS                      configs.NatsJetStreamConfig
	ElasticSearch               configs.ElasticSearchConfig
	BobHasuraAdminURL           string `yaml:"bob_hasura_admin_url"`
	EurekaHasuraAdminURL        string `yaml:"eureka_hasura_admin_url"`
	FatimaHasuraAdminURL        string `yaml:"fatima_hasura_admin_url"`
	TimesheetHasuraAdminURL     string `yaml:"timesheet_hasura_admin_url"`
	JWTApplicant                string `yaml:"jwt_applicant"`
	IdentityToolkitAPI          string

	Kafka               *configs.KafkaConfig
	KafkaConnectConfig  *configs.KafkaConnectConfig      `yaml:"kafka_connect"`
	UnleashSrvAddr      string                           `yaml:"unleash_srv_addr"`
	IdentityPlatform    *configurations.IdentityPlatform `yaml:"identity_platform"`
	UnleashClientConfig configs.UnleashClientConfig      `yaml:"unleash_client"`
	TraceEnabled        bool

	Upload     configs.UploadConfig     `yaml:"upload"`
	Whiteboard configs.WhiteboardConfig `yaml:"whiteboard"`

	// Bob configs
	Agora                       *bobConfigs.AgoraConfig `yaml:"agora"`
	EntryexitmgmtHasuraAdminURL string                  `yaml:"entryexitmgmt_hasura_admin_url"`
	InvoicemgmtHasuraAdminURL   string                  `yaml:"invoicemgmt_hasura_admin_url"`

	// mastermgmt configs
	MastermgmtHasuraAdminURL string `yaml:"mastermgmt_hasura_admin_url"`

	GitHubWebhookSecret string `yaml:"github_webhook_secret"`

	// usermgmt configs
	JobAccounts  []configurations.JobAccount `yaml:"job_accounts"`
	WithUsConfig configurations.WithUsConfig `yaml:"with_us"`
}
