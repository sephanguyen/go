package dplparser

type SourceConfig struct {
	DeployEnvs []string `yaml:"deployEnv"`
	DeployOrgs []string `yaml:"deployOrg"`

	Name           string
	FileName       string
	Table          string
	Database       string
	Schema         string
	HeartbeatQuery string
}
