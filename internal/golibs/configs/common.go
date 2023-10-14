package configs

import (
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Duration custom duration type to work with yaml
type Duration time.Duration

// UnmarshalYAML implements https://pkg.go.dev/gopkg.in/yaml.v3#Unmarshaler
func (d *Duration) UnmarshalYAML(v *yaml.Node) error {
	duration, err := time.ParseDuration(v.Value)
	*d = Duration(duration)
	return err
}

// CommonConfig with common configuration for all services
type CommonConfig struct {
	// Name is the name of the service.
	Name string `yaml:"name"`

	// Environment can be local, stag, uat, prod.
	// Note that for preproduction environment, the value here will be "prod".
	//
	// To distinguish between production vs preproduction, use ActualEnvironment instead.
	Environment string `yaml:"environment"`

	// ActualEnvironment can be local, stag, uat, dorp, prod.
	//
	// This should only be used by Platform squad. From a backend perspective,
	// there should be no differences between "prod" vs "dorp", and thus you should
	// use Environment instead, unless there is a strong reason not to.
	ActualEnvironment string `yaml:"actual_environment"`

	// Organization can be manabie, jprep, aic, ga, synersia, renseikai, tokyo.
	Organization string `yaml:"organization"`

	// ServiceAccountEmail is the email of the IAM service account running this server.
	// For example: prod-bob@student-coach-e1e95.iam.gserviceaccount.com.
	ServiceAccountEmail string `yaml:"sa_email"`

	// ImageTag is the version string of the currently deployed server.
	ImageTag string `yaml:"image_tag"`

	GoogleCloudProject      string            `yaml:"google_cloud_project"`
	FirebaseProject         string            `yaml:"firebase_project"`
	IdentityPlatformProject string            `yaml:"identity_platform_project"`
	StatsEnabled            bool              `yaml:"stats_enabled"`
	RemoteTrace             RemoteTraceConfig `yaml:"remote_trace"`
	GRPC                    GRPCConfig        `yaml:"grpc"`
	Log                     LogConfig
}

// Hostname return empty if cannot get hostname
func (cc *CommonConfig) Hostname() string {
	h, err := os.Hostname()
	if err != nil {
		return ""
	}

	h = strings.ReplaceAll(h, ".", "")

	return h
}

type GRPCClientsConfig struct {
	BobSrvAddr              string       `yaml:"bob_srv_addr"`
	TomSrvAddr              string       `yaml:"tom_srv_addr"`
	YasuoSrvAddr            string       `yaml:"yasuo_srv_addr"`
	EurekaSrvAddr           string       `yaml:"eureka_srv_addr"`
	FatimaSrvAddr           string       `yaml:"fatima_srv_addr"`
	ShamirSrvAddr           string       `yaml:"shamir_srv_addr"`
	UserMgmtSrvAddr         string       `yaml:"userMgmt_srv_addr"`
	NotificationMgmtSrvAddr string       `yaml:"notificationMgmt_srv_addr"`
	PaymentSrvAddr          string       `yaml:"payment_srv_addr"`
	EntryExitMgmtSrvAddr    string       `yaml:"entryExitMgmt_srv_addr"`
	MasterMgmtSrvAddr       string       `yaml:"masterMgmt_srv_addr"`
	InvoiceMgmtSrvAddr      string       `yaml:"invoiceMgmt_srv_addr"`
	LessonMgmtSrvAddr       string       `yaml:"lessonMgmt_srv_addr"`
	EnigmaSrvAddr           string       `yaml:"enigma_srv_addr"`
	VirtualClassroomSrvAddr string       `yaml:"virtualClassroom_srv_addr"`
	CalendarSrvAddr         string       `yaml:"calendar_srv_addr"`
	TimesheetSrvAddr        string       `yaml:"timesheet_srv_addr"`
	DiscountSrvAddr         string       `yaml:"discount_srv_addr"`
	RetryOptions            RetryOptions `yaml:"retry_config"`
}

type RetryOptions struct {
	MaxCall      int `yaml:"max_retry"`
	RetryTimeout int `yaml:"retry_timeout"`
}

// LogConfig for log
type LogConfig struct {
	ApplicationLevel string `yaml:"app_level"`
	LogPayload       bool   `yaml:"log_payload"`
}

// RemoteTraceConfig for common tracing configuration
type RemoteTraceConfig struct {
	Enabled               bool
	OtelCollectorReceiver string `yaml:"otel_collector_receiver"`
}

// GRPCConfig for common config gRPC services
type GRPCConfig struct {
	TraceEnabled     bool                 `yaml:"trace_enabled"`
	HandlerTimeout   time.Duration        `yaml:"handler_timeout"`
	HandlerTimeoutV2 GRPCHandlerTimeoutV2 `yaml:"handler_timeout_v2"`

	// HandlerTimeoutV2Enabled is a feature flag to turn on/off handler timeout v2.
	HandlerTimeoutV2Enabled bool `yaml:"handler_timeout_v2_enabled"`
}

// GRPCHandlerTimeoutV2 contains the timeout (implemented using GRPC unary interceptor) for
// each specified GRPC API. The key must be the full method name of the API.
// An example of a full method name is: "/bob.v1.InternalReaderService/VerifyAppVersion".
//
// The "default" key-value, when specified, contains the default timeout for all APIs.
// A negative value disables the timeout for that specific API (or disables timeout by default
// when set to the "default" key).
type GRPCHandlerTimeoutV2 map[string]time.Duration

// TokenIssuerConfig contains information to work with OIDC provider
type TokenIssuerConfig struct {
	Issuer       string `yaml:"issuer"`
	Audience     string `yaml:"audience"`
	JWKSEndpoint string `yaml:"jwks_endpoint"`
}

// BrightcoveConfig contains Brightcove config
type BrightcoveConfig struct {
	AccountID           string `yaml:"account_id"`
	ClientID            string `yaml:"client_id"`
	Secret              string
	Profile             string
	PolicyKey           string `yaml:"policy_key"`
	PolicyKeyWithSearch string `yaml:"policy_key_with_search"`
}

// StorageConfig contains s3 config
type StorageConfig struct {
	Endpoint                 string
	Region                   string
	Bucket                   string
	AccessKey                string        `yaml:"access_key"`
	SecretKey                string        `yaml:"secret_key"`
	MaximumURLExpiryDuration time.Duration `yaml:"maximum_url_expiry_duration"`
	MinimumURLExpiryDuration time.Duration `yaml:"minimum_url_expiry_duration"`
	DefaultURLExpiryDuration time.Duration `yaml:"default_url_expiry_duration"`
	Secure                   bool          `yaml:"secure"`
	FileUploadFolderPath     string        `yaml:"file_upload_folder_path"`
	InsecureSkipVerify       bool          `yaml:"insecure_skip_verify"`
}

// UploadConfig contains config for uploading file
type UploadConfig struct {
	MaxChunkSize int64 `yaml:"max_chunk_size"`
	MaxFileSize  int64 `yaml:"max_file_size"`
}

// ListenerConfig contains config for listener
type ListenerConfig struct {
	GRPC                 string
	HTTP                 string
	MigratedEnvironments []string
}

// WhiteboardConfig for netless
type WhiteboardConfig struct {
	AppID              string `yaml:"app_id"`
	Endpoint           string
	Token              string
	TokenLifeSpan      time.Duration `yaml:"token_life_span"`
	HttpTracingEnabled bool          `yaml:"http_tracing_enabled"`
}

type ElasticSearchConfig struct {
	Username  string
	Password  string
	Addresses []string `yaml:"addresses"`
}

type NatsJetStreamConfig struct {
	Address        string        `yaml:"address"`
	DefaultAckWait time.Duration `yaml:"default_ack_wait"`
	MaxRedelivery  int           `yaml:"max_redelivery"`
	JprepAckWait   *Duration     `yaml:"jprep_ack_wait"`
	MaxReconnects  int           `yaml:"max_reconnect"`
	ReconnectWait  time.Duration `yaml:"reconnect_wait"`
	User           string        `yaml:"user"`
	Password       string        `yaml:"password"`

	// IsLocal indicates whether the current environment is local.
	// When in local:
	// 	 - stream replica count is 1 (down from 3)
	//   - ack wait is reduced
	IsLocal bool `yaml:"is_local"`
}

type PartnerConfig struct {
	DomainBo      string `yaml:"domain_bo"`
	DomainTeacher string `yaml:"domain_teacher"`
	DomainLearner string `yaml:"domain_learner"`
}

type KeyCloakAuthConfig struct {
	Path     string `yaml:"path"`
	Realm    string `yaml:"realm"`
	ClientID string `yaml:"client_id"`
}

// Used by migration scripts to authenticate BatchJob
type TenantCLIAuthConfig struct {
	Name     string `yaml:"name"`
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
}

type KafkaClusterConfig struct {
	Address          string `yaml:"address"`
	ObjectNamePrefix string `yaml:"object_name_prefix"`
	IsLocal          bool   `yaml:"is_local"`
}

type KafkaConfig struct {
	Addr     []string `yaml:"address"`
	KsqlAddr string   `yaml:"ksql_addr"`
	Connect  KafkaConnectConfig
	Username *string `yaml:"username"`
	Password *string `yaml:"password"`
	EnableAC bool    `yaml:"enable_ac"`
}

type KafkaConnectConfig struct {
	Addr               string `yaml:"addr"`
	SourceConfigDir    string `yaml:"source_config_dir"`
	SinkConfigDir      string `yaml:"sink_config_dir"`
	GenSourceConfigDir string `yaml:"gen_source_config_dir"`
	GenSinkConfigDir   string `yaml:"gen_sink_config_dir"`
}

type UnleashClientConfig struct {
	URL      string `yaml:"url"`
	AppName  string `yaml:"app_name"`
	APIToken string `yaml:"api_token"`
}

type UnleashAdminConfig struct {
	APIToken string `yaml:"api_token"`
}

type JiraConfig struct {
	Email         string `yaml:"email"`
	Token         string `yaml:"token"`
	APIBaseURL    string `yaml:"api_base_url"`
	APITimeFormat string `yaml:"api_time_format"`
}

type GithubConfig struct {
	AppID          int64 `yaml:"app_id"`
	InstallationID int64 `yaml:"installation_id"`
}

type MongoConfig struct {
	Connection string `yaml:"connection"`
	Database   string `yaml:"database"`
}

type ZoomConfig struct {
	SecretKey    string `yaml:"secret_key"`
	EndpointAuth string `yaml:"endpoint_oauth"`
	Endpoint     string `yaml:"endpoint"`
}

type ClassDoConfig struct {
	SecretKey string `yaml:"secret_key"`
	Endpoint  string `yaml:"endpoint"`
}

type AppsmithAPI struct {
	ENDPOINT      string `yaml:"endpoint"`
	ApplicationID string `yaml:"application_id"`
	Authorization string `yaml:"authorization"`
}

type ZegoCloudConfig struct {
	AppID        int    `yaml:"app_id"`
	ServerSecret string `yaml:"server_secret"`
	AppSign      string `yaml:"app_sign"`
}

type SendGridConfig struct {
	APIKey    string `yaml:"api_key"`
	PublicKey string `yaml:"public_key"`
}

type AgoraConfig struct {
	AppID              string `yaml:"app_id"`
	PrimaryCertificate string `yaml:"primary_certificate"`
	AppName            string `yaml:"app_name"`
	OrgName            string `yaml:"org_name"`
	RestAPI            string `yaml:"rest_api"`
	WebhookSecret      string `yaml:"webhook_secret"`
}
