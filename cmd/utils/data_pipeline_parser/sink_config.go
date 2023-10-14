package dplparser

import (
	"fmt"
	"sort"
	"strings"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type SinkConfig struct {
	Database string `yaml:"database"`
	Name     string `yaml:"name"`

	CaptureDeleteAll   bool     `yaml:"captureDeleteAll"`
	CaptureDeleteEnvs  []string `yaml:"captureDeleteEnvs"`
	ExcludeColumns     []string `yaml:"excludeColumns"`
	DeployEnvs         []string `yaml:"deployEnv"`
	DeployOrgs         []string `yaml:"deployOrg"`
	DeploySchemas      []string `yaml:"deploySchema"`
	FilterResourcePath string   `yaml:"filterResourcePath"`

	CaptureDeleteEnabled bool
	// need to update these config
	// Name is the connector name in kafka connect
	// $DbSource_to_$DbSink_$Table_sink_connector
	// FileName is the file name when generate gen file with json file extension
	// it will be in format $Name.json.
	FileName string

	// auto generated primary key from table schema
	PrimaryKeys []string
	// auto generated columns from table schema
	Columns []string
	Table   string

	PipelineConfigs *[]PipelineConfig `yaml:"pipelineConfigs"`
}

func (s *SinkConfig) AddPrimaryKeyConfig(tableSchemaDir, tableName, schemaName string) error {
	tableSchema, err := database.LoadTableSchema(tableSchemaDir, s.Database, tableName, schemaName)
	if err != nil {
		return err
	}

	for _, c := range tableSchema.Constraint {
		if c.ConstraintType == "PRIMARY KEY" {
			s.PrimaryKeys = append(s.PrimaryKeys, c.ColumName)
		}
	}

	return nil
}

func (s *SinkConfig) AddColumnConfig(tableSchemaDir, tableName, schemaName string) error {
	tableSchema, err := database.LoadTableSchema(tableSchemaDir, s.Database, tableName, schemaName)
	if err != nil {
		return err
	}

	for _, column := range tableSchema.Schema {
		s.Columns = append(s.Columns, column.ColumnName)
	}

	s.RemoveColumnsInExludeList()
	if len(s.Columns) == 0 {
		return fmt.Errorf("empty column config")
	}

	sort.Strings(s.Columns)
	return nil
}

func (s *SinkConfig) UpdateName(srcDB, sinkDB, table string) {
	if s.Name == "" {
		s.Name = fmt.Sprintf("%s_to_%s_SCHEMA.%s_sink_connector", srcDB, sinkDB, table)
	}
}

func (s *SinkConfig) UpdateFileName() {
	switch {
	case strings.HasSuffix(s.Name, "_sink_connector"):
		s.FileName = fmt.Sprintf("%s.json", strings.TrimSuffix(s.Name, "_sink_connector"))
	case strings.HasSuffix(s.Name, "_connector"):
		s.FileName = fmt.Sprintf("%s.json", strings.TrimSuffix(s.Name, "_connector"))
	default:
		s.FileName = fmt.Sprintf("%s.json", s.Name)
	}
}

func (s *SinkConfig) UpdateDeployEnvsAndOrg(envList, orgList []string, preProductionEnabled bool) {
	if s.DeployEnvs == nil || len(s.DeployEnvs) == 0 {
		s.DeployEnvs = envList
	}
	if preProductionEnabled {
		s.AddPreProductionEnv()
	}
	if len(s.DeployOrgs) == 0 {
		for _, env := range s.DeployEnvs {
			switch env {
			case "local":
				s.DeployOrgs = append(s.DeployOrgs, "manabie")
			case "stag":
				s.DeployOrgs = append(s.DeployOrgs, []string{"manabie", "jprep"}...)
			case "prod":
				s.DeployOrgs = append(s.DeployOrgs, orgList...)
			}
		}
	}
	s.DeployOrgs = golibs.Uniq(s.DeployOrgs)
}

func (s *SinkConfig) AddPreProductionEnv() {
	s.DeployEnvs = append(s.DeployEnvs, "dorp")
}

func (s *SinkConfig) RemoveColumnsInExludeList() {
	newColumnList := make([]string, 0, len(s.Columns))
	mp := make(map[string]bool)
	for _, column := range s.ExcludeColumns {
		mp[column] = true
	}

	for _, column := range s.Columns {
		if _, ok := mp[column]; !ok {
			newColumnList = append(newColumnList, column)
		}
	}

	s.Columns = newColumnList
}

func (s *SinkConfig) isCaptureDeleteEventEnvs(env string) bool {
	return golibs.InArrayString(env, s.CaptureDeleteEnvs)
}

func (s *SinkConfig) UpdateSchema() {
	if s.DeploySchemas == nil || len(s.DeploySchemas) == 0 {
		s.DeploySchemas = []string{"public"}
	}
}
