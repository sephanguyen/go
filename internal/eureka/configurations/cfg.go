package configurations

import (
	"github.com/manabie-com/backend/internal/golibs/configs"
)

// Config for Eureka
type Config struct {
	Common               configs.CommonConfig
	Issuers              []configs.TokenIssuerConfig
	Storage              configs.StorageConfig
	PostgresV2           configs.PostgresConfigV2 `yaml:"postgres_v2"`
	NatsJS               configs.NatsJetStreamConfig
	JWTApplicant         string                      `yaml:"jwt_applicant"`
	KeyCloakAuth         configs.KeyCloakAuthConfig  `yaml:"keycloak_auth"`
	SyllabusTimeMonitor  SyllabusTimeMonitorConfig   `yaml:"syllabus_time_monitor"`
	SchoolInformation    SchoolInfoConfig            `yaml:"school_information"`
	SyllabusSlackWebHook string                      `yaml:"syllabus_slack_webhook"`
	Mathpix              MathpixConfig               `yaml:"mathpix"`
	WithusRelayServer    WithusRelayServer           `yaml:"withus_relay_server"`
	LearnosityConfig     LearnosityConfig            `yaml:"learnosity"`
	UnleashClientConfig  configs.UnleashClientConfig `yaml:"unleash_client"`
}

type SyllabusTimeMonitorConfig struct {
	CourseStudentUpserted int `yaml:"course_student_upserted" json:"course_student_upserted,omitempty"` // minutes
	LearningItemUpserted  int `yaml:"learning_item_upserted" json:"learning_item_upserted,omitempty"`
}

type SchoolInfoConfig struct {
	SchoolID   string `yaml:"school_id"`
	SchoolName string `yaml:"school_name"`
}

type MathpixConfig struct {
	AppID  string `yaml:"mathpix_app_id"`
	AppKey string `yaml:"mathpix_app_key"`
}

type WithusRelayServer struct {
	LMSFilePath string `yaml:"lms_file_path"`
	LMSFileName string `yaml:"lms_file_name"`
}

type LearnosityConfig struct {
	ConsumerKey    string `yaml:"consumer_key"`
	ConsumerSecret string `yaml:"consumer_secret"`
}
