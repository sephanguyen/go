package dplparser

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type SourceConnectorConfig struct {
	FileName       string
	Database       string
	Tables         string
	FilePath       string
	DeployEnv      string
	DeployOrg      string
	Name           string
	HeartbeatQuery string
}

func (p *DataPipelineParser) ParseSource() (map[ConnectorConfig]string, error) {
	err := p.DataPipelineDef.UpdatePipelineConfig(p.TableSchemaDir)
	if err != nil {
		return nil, err
	}
	tpl := template.Must(template.New("").Delims("[[", "]]").Parse(p.Tpl))
	sourceRaw := make(map[string][]SourceConfig)

	for _, dpl := range p.DataPipelineDef.Datapipelines {
		dpl.Source.Database = p.DataPipelineDef.Database
		dpl.Source.Schema = p.DataPipelineDef.Schema
		dpl.Source.Table = dpl.Table
		if contains(p.DataPipelineDef.DBUseCustomHeartBeat, dpl.Source.Database) {
			dpl.Source.HeartbeatQuery = p.DataPipelineDef.CustomHeartbeatQuery
		} else {
			dpl.Source.HeartbeatQuery = p.DataPipelineDef.DefaultHeartBeatQuery
		}

		if dpl.Source.DeployEnvs == nil || len(dpl.Source.DeployEnvs) == 0 {
			dpl.Source.DeployEnvs = p.DataPipelineDef.Envs
		}
		if dpl.PipelineConfigs == nil {
			for _, env := range dpl.Source.DeployEnvs {
				if dpl.Source.DeployOrgs == nil || len(dpl.Source.DeployOrgs) == 0 {
					dpl.Source.DeployOrgs = p.DataPipelineDef.Orgs
				}
				for _, org := range dpl.Source.DeployOrgs {
					if env != "prod" && org != "manabie" {
						continue
					}
					key := buildKeyMap(dpl.Source.Database, env, org)
					sourceRaw[key] = append(sourceRaw[key], dpl.Source)
				}
			}
		} else {
			for _, config := range dpl.PipelineConfigs {
				if !isExcludedSinkDB(p.Excludes, config.Env, config.Org, dpl.Source.Database, dpl.Source.Database) {
					key := buildKeyMap(dpl.Source.Database, config.Env, config.Org)
					sourceRaw[key] = append(sourceRaw[key], dpl.Source)
				}
			}
		}
	}

	return buildSourceConnector(sourceRaw, tpl)
}
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
func buildKeyMap(database string, env string, org string) string {
	return database + "/" + env + "/" + org
}
func buildSourceConnector(raw map[string][]SourceConfig, tpl *template.Template) (map[ConnectorConfig]string, error) {
	source := make(map[ConnectorConfig]string)
	for key, v := range raw {
		env, org, error := splitKey(key)
		if error != nil {
			return nil, error
		}
		tableLists := make([]string, 0)
		tableLists = append(tableLists, "public.dbz_signals")
		heartBeatQuery := ""
		for _, cfg := range v {
			tableLists = append(tableLists, (cfg.Schema + "." + cfg.Table))
			heartBeatQuery = cfg.HeartbeatQuery
		}
		sourceConfig := buildSourceConnectorConfig(tableLists, org, env, v[0], heartBeatQuery)
		var configRaw bytes.Buffer
		err := tpl.Execute(&configRaw, sourceConfig)
		if err != nil {
			return nil, err
		}
		connectorConfig := buildConnectorConfig(org, env, sourceConfig)
		source[connectorConfig] = strings.TrimSpace(configRaw.String())
	}
	return source, nil
}

func splitKey(key string) (string, string, error) {
	strings := strings.Split(key, "/")
	if len(strings) < 3 {
		return "", "", fmt.Errorf("wrong format key ")
	}
	return strings[1], strings[2], nil
}

func buildConnectorConfig(org string, env string, source SourceConnectorConfig) ConnectorConfig {
	return ConnectorConfig{source.FileName, filepath.Join(org, env)}
}

func buildSourceConnectorConfig(tables []string, org string, env string, source SourceConfig, heartBeatQuery string) SourceConnectorConfig {
	name := strings.ReplaceAll(source.FileName, ".json", "")
	return SourceConnectorConfig{source.FileName, source.Database, buildTableList(tables), filepath.Join(org, env), env, org, name, heartBeatQuery}
}

func buildTableList(tables []string) string {
	reduceList := removeDuplicateStr(tables)
	sort.Slice(reduceList, func(i, j int) bool {
		return reduceList[i] < reduceList[j]
	})
	return strings.Join(reduceList, ",")
}

func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func (p *DataPipelineParser) ExportSource(dpls map[ConnectorConfig]string, outDir string) error {
	for connectorConfig, config := range dpls {
		_, _ = create(filepath.Join(outDir, connectorConfig.filePath, connectorConfig.fileName))
		err := os.WriteFile(filepath.Join(outDir, connectorConfig.filePath, connectorConfig.fileName), []byte(config), 0600)
		if err != nil {
			return err
		}
	}
	return nil
}
