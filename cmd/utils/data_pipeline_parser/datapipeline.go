package dplparser

import (
	"fmt"

	"go.uber.org/multierr"
)

type DataPipeline struct {
	Name   string        `yaml:"name"`
	Table  string        `yaml:"table"`
	Source SourceConfig  `yaml:"source"`
	Sinks  []*SinkConfig `yaml:"sinks"`
	// Sink used for mapping to template file
	Sink      SinkConfig
	DeployEnv string
	DeployOrg string

	PipelineConfigs []PipelineConfig `yaml:"pipelineConfigs"`
}

type PipelineConfig struct {
	Env string `yaml:"env"`
	Org string `yaml:"org"`
}

type DataPipelineDef struct {
	Datapipelines         []*DataPipeline `yaml:"datapipelines"`
	Envs                  []string        `yaml:"envs"`
	Orgs                  []string        `yaml:"orgs"`
	DBUseCustomHeartBeat  []string        `yaml:"dbUseCustomHeartBeat"`
	DefaultHeartBeatQuery string          `yaml:"defaultHeartBeatQuery"`
	CustomHeartbeatQuery  string          `yaml:"customHeartbeatQuery"`
	Database              string          `yaml:"database"`
	Schema                string          `yaml:"schema"`
	PreProductionEnabled  bool            `yaml:"preProductionEnabled"`

	PipelineConfigs []PipelineConfig `yaml:"pipelineConfigs"`
}

func (p *DataPipelineDef) UpdatePipelineConfig(tableSchemaDir string) (err error) {
	// Update name and filename of the pipeline
	for _, pl := range p.Datapipelines {
		for _, s := range pl.Sinks {
			s.UpdateName(p.Database, s.Database, pl.Table)
			s.UpdateFileName()
			s.UpdateSchema()
			s.UpdateDeployEnvsAndOrg(p.Envs, p.Orgs, p.PreProductionEnabled)
		}
		pl.Source.Name = fmt.Sprintf("%s_source", p.Database)
		pl.Source.FileName = fmt.Sprintf("%s.json", pl.Source.Name)
	}

	if tableSchemaDir != "" {
		for _, c := range p.Datapipelines {
			for _, s := range c.Sinks {
				for _, schema := range s.DeploySchemas {
					err = multierr.Combine(
						s.AddPrimaryKeyConfig(tableSchemaDir, c.Table, schema),
						s.AddColumnConfig(tableSchemaDir, c.Table, schema),
					)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return err
}
