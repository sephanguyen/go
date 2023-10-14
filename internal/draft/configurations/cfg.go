package configurations

import (
	"github.com/manabie-com/backend/internal/golibs/configs"
)

type Config struct {
	Common              configs.CommonConfig
	PostgresV2          configs.PostgresConfigV2 `yaml:"postgres_v2"`
	GitHubWebhookSecret string                   `yaml:"github_webhook_secret"`
	GithubPrivateKey    string                   `yaml:"github_private_secret"`
	Github              configs.GithubConfig
	NatsJS              configs.NatsJetStreamConfig
	DataPruneConfig     DataPruneConfig `yaml:"data_prune"`
	DraftAPISecret      string          `yaml:"draft_api_secret"`
}

type DataPruneConfig struct {
	PostgresCommonInstance configs.PostgresDatabaseConfig `yaml:"postgres_common_instance"`
	PostgresLMSInstance    configs.PostgresDatabaseConfig `yaml:"postgres_lms_instance"`
	ServiceConfigs         map[string]CleanTableConfig    `yaml:"clean_data"`
}

type CleanTableConfig struct {
	CreatedAtColName    string            `yaml:"created_at_col_name"`
	ExtraCond           string            `yaml:"extra_cond"`
	IgnoreFks           []string          `yaml:"ignore_fks"`
	SelfRefFKs          []SelfRefFKs      `yaml:"self_ref_fks"`
	SetNullOnCircularFk map[string]string `yaml:"set_null_on_circular_fk"`
}

type SelfRefFKs struct {
	Referencing string `yaml:"referencing"`
	Referenced  string `yaml:"referenced"`
}
