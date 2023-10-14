package infras

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/j4/serviceutil/syllabus"
	virtualclassroom "github.com/manabie-com/backend/j4/serviceutil/virtualclassroom"

	j4 "github.com/manabie-com/j4/pkg/runner"
)

type ManabieJ4Config struct {
	Env             string                   `yaml:"env"`
	ClusterGrpcAddr string                   `yaml:"cluster_grpc_addr"`
	ShamirAddr      string                   `yaml:"shamir_addr"`
	PostgresV2      configs.PostgresConfigV2 `yaml:"postgres_v2"`
	FirebaseAPIKey  string                   `yaml:"firebase_api_key"`
	AdminID         string                   `yaml:"admin_id"` // for local only
	SchoolID        string                   `yaml:"school_id"`

	ScenarioConfigs []ScenarioConfig `yaml:"scenario_configs"`
	HasuraConfigs   []HasuraConfig   `yaml:"hasura_configs"`
	KafkaConfig     KafkaCluster     `yaml:"kafka_cluster"`

	VirtualClassroomConfig virtualclassroom.Config `yaml:"virtualclassroom_config"`
	SyllabusConfig         syllabus.Config         `yaml:"syllabus_config"`
}
type HasuraConfig struct {
	Name      string `yaml:"name"`
	FilePath  string `yaml:"file_path"`
	AdminAddr string `yaml:"admin_addr"`
}

type KafkaCluster struct {
	Address          string `yaml:"address"`
	ObjectNamePrefix string `yaml:"object_name_prefix"`
	IsLocal          bool   `yaml:"is_local"`
}

func (c *ManabieJ4Config) GetScenarioConfig(scname string) (ScenarioConfig, error) {
	for _, item := range c.ScenarioConfigs {
		if item.Name == scname {
			return item, nil
		}
	}
	return ScenarioConfig{}, fmt.Errorf("not found config for scenario %s", scname)
}

type ScenarioConfig struct {
	Name                     string `yaml:"name"`
	Interval                 int64  `yaml:"interval"`
	TargetCount              int64  `yaml:"target_count"`
	RampUpCycles             int64  `yaml:"ramp_up_cycles"`
	HoldCycles               int64  `yaml:"hold_cycles"`
	RampDownCycles           int64  `yaml:"ramp_down_cycles"`
	IntervalBetweenExecution int64  `yaml:"interval_between_execution"`
}

func MustOptionFromConfig(cfg *ScenarioConfig) *j4.Option {
	if cfg.Interval == 0 || cfg.RampUpCycles == 0 ||
		cfg.RampDownCycles == 0 || cfg.HoldCycles == 0 || cfg.TargetCount == 0 {
		panic("invalid j4 run option")
	}
	return &j4.Option{
		TargetCount:              cfg.TargetCount,
		Interval:                 cfg.Interval,
		RampUpCycles:             cfg.RampDownCycles,
		RampDownCycles:           cfg.RampDownCycles,
		HoldCycles:               cfg.HoldCycles,
		IntervalBetweenExecution: cfg.IntervalBetweenExecution,
	}
}
