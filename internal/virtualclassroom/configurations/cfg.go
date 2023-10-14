package configurations

import (
	"github.com/manabie-com/backend/internal/golibs/configs"
)

type Config struct {
	Common              configs.CommonConfig
	Issuers             []configs.TokenIssuerConfig
	PostgresV2          configs.PostgresConfigV2 `yaml:"postgres_v2"`
	NatsJS              configs.NatsJetStreamConfig
	Storage             configs.StorageConfig
	Agora               AgoraConfig `yaml:"agora"`
	Whiteboard          configs.WhiteboardConfig
	UnleashClientConfig configs.UnleashClientConfig `yaml:"unleash_client"`
	ZegoCloudConfig     ZegoCloudConfig             `yaml:"zegocloud"`
}

// AgoraConfig for Agora
type AgoraConfig struct {
	AppID                    string `yaml:"app_id"`
	Cert                     string `yaml:"cert"`
	VideoTokenSuffix         string `yaml:"video_token_suffix"`
	MaximumLearnerStreamings int    `yaml:"maximum_learner_streamings"`
	CustomerID               string `yaml:"customer_id"`
	CustomerSecret           string `yaml:"customer_secret"`
	Endpoint                 string `yaml:"endpoint"`
	BucketName               string `yaml:"bucket"`
	BucketAccessKey          string `yaml:"bucket_access_key"`
	BucketSecretKey          string `yaml:"bucket_secret_key"`
	CallbackSignature        string `yaml:"callback_signature"`
	MaxIdleTime              int    `yaml:"max_idle_time"`
}

type ZegoCloudConfig struct {
	AppID         int    `yaml:"app_id"`
	ServerSecret  string `yaml:"server_secret"`
	AppSign       string `yaml:"app_sign"`
	TokenValidity int    `yaml:"token_validity"` // in seconds
}
