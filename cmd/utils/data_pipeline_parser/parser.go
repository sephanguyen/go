package dplparser

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/manabie-com/backend/internal/golibs"

	"gopkg.in/yaml.v2"
)

type Option func(*DataPipelineParser)

type ExcludeConfig struct {
	Env      string
	Org      string
	SinkDB   string
	SourceDB string
}

type DataPipelineParser struct {
	FilePath        string
	DataPipelineDef *DataPipelineDef
	Tpl             string
	TableSchemaDir  string
	ReadFileCustom  func(fileName string) ([]byte, error)
	Excludes        []ExcludeConfig
	DeleteConnector bool
}

type ConnectorConfig struct {
	fileName string
	filePath string
}

func WithTpl(tpl string) Option {
	return func(d *DataPipelineParser) {
		d.Tpl = tpl
	}
}

func WithTableSchemaDir(dir string) Option {
	return func(d *DataPipelineParser) {
		d.TableSchemaDir = dir
	}
}

func WithCustomReader(rf func(fileName string) ([]byte, error)) Option {
	return func(d *DataPipelineParser) {
		d.ReadFileCustom = rf
	}
}

func parseExcludes(excludes []string) []ExcludeConfig {
	result := make([]ExcludeConfig, 0)
	for _, e := range excludes {
		s := strings.Split(e, ":")
		if len(s) == 4 {
			result = append(result, ExcludeConfig{
				Env:      s[0],
				Org:      s[1],
				SinkDB:   s[2],
				SourceDB: s[3],
			})
		}
	}
	return result
}

func WithExcluded(excludes []string) Option {
	return func(d *DataPipelineParser) {
		d.Excludes = parseExcludes(excludes)
	}
}

func NewDataPipelineParser(filePath string, opts ...Option) (*DataPipelineParser, error) {
	p := &DataPipelineParser{FilePath: filePath}
	for _, opt := range opts {
		opt(p)
	}
	err := p.ParseDataPipelineDef()
	if err != nil {
		return nil, err
	}
	err = p.DataPipelineDef.UpdatePipelineConfig(p.TableSchemaDir)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *DataPipelineParser) ParseDataPipelineDef() error {
	var (
		b   []byte
		err error
	)
	if p.ReadFileCustom != nil {
		b, err = p.ReadFileCustom(p.FilePath)
	} else {
		b, err = os.ReadFile(p.FilePath)
	}
	if err != nil {
		return err
	}
	dpl := &DataPipelineDef{}
	err = yaml.Unmarshal(b, dpl)
	if err != nil {
		return fmt.Errorf("failed to parse yaml: %v", err)
	}
	p.DataPipelineDef = dpl
	return nil
}

func compareConfig(excludeConfigStr string, currentConfigStr string) bool {
	if excludeConfigStr == "" || excludeConfigStr == currentConfigStr {
		return true
	}
	return false
}

func isExcludedSinkDB(excludes []ExcludeConfig, env, org, sinkDB, sourceDB string) bool {
	for _, e := range excludes {
		if compareConfig(e.Env, env) && compareConfig(e.Org, org) && compareConfig(e.SinkDB, sinkDB) && compareConfig(e.SourceDB, sourceDB) {
			return true
		}
	}
	return false
}

func isIgnoreFile(str, filePath string) bool {
	if str == "" || strings.Contains(filePath, str) {
		return true
	}

	return false
}
func ignoreWhenDeleteFile(excludes []ExcludeConfig, filePath string) bool {
	for _, e := range excludes {
		if isIgnoreFile(e.Env, filePath) && isIgnoreFile(e.Org, filePath) && isIgnoreFile(e.SinkDB, filePath) && isIgnoreFile(e.SourceDB, filePath) {
			return true
		}
	}
	return false
}

func (p *DataPipelineParser) Parse() (map[ConnectorConfig]string, error) {
	if p.DataPipelineDef == nil {
		return nil, fmt.Errorf("data pipeline definition is not defined")
	}
	result := make(map[ConnectorConfig]string)
	tpl := template.Must(template.New("").Delims("[[", "]]").Parse(p.Tpl))
	for _, dpl := range p.DataPipelineDef.Datapipelines {
		dpl.Source.Database = p.DataPipelineDef.Database
		dpl.Source.Schema = p.DataPipelineDef.Schema
		for _, sink := range dpl.Sinks {
			sink.Table = dpl.Table
			if sink.PipelineConfigs == nil {
				for _, env := range sink.DeployEnvs {
					for _, org := range sink.DeployOrgs {
						if validateEnvAndOrg(env, org) && !isExcludedSinkDB(p.Excludes, env, org, sink.Database, dpl.Source.Database) {
							for _, schema := range sink.DeploySchemas {
								// Keep the old consumer name
								if schema == "public" {
									sink.Name = strings.ReplaceAll(sink.Name, "SCHEMA.", "")
									sink.Table = dpl.Table
								} else {
									sink.Name = strings.ReplaceAll(sink.Name, "SCHEMA", schema)
									sink.Table = fmt.Sprintf("%s.%s", schema, dpl.Table)
								}

								dpl.Sink = *sink
								dpl.DeployEnv = env
								dpl.DeployOrg = org
								dpl.Sink.CaptureDeleteEnabled = dpl.Sink.CaptureDeleteAll || dpl.Sink.isCaptureDeleteEventEnvs(env)
								var configRaw bytes.Buffer
								connectorConfig := buildSinkConnectorConfig(org, env, sink, schema)
								err := tpl.Execute(&configRaw, dpl)
								if err != nil {
									return nil, err
								}

								result[connectorConfig] = strings.TrimSpace(configRaw.String())
							}
						}
					}
				}
			} else {
				for _, config := range *sink.PipelineConfigs {
					if !isExcludedSinkDB(p.Excludes, config.Env, config.Org, sink.Database, dpl.Source.Database) {
						for _, schema := range sink.DeploySchemas {
							// Keep the old consumer name
							if schema == "public" {
								sink.Name = strings.ReplaceAll(sink.Name, "SCHEMA.", "")
								sink.Table = dpl.Table
							} else {
								sink.Name = strings.ReplaceAll(sink.Name, "SCHEMA", schema)
								sink.Table = fmt.Sprintf("%s.%s", schema, dpl.Table)
							}

							dpl.Sink = *sink
							dpl.DeployEnv = config.Env
							dpl.DeployOrg = config.Org
							dpl.Sink.CaptureDeleteEnabled = dpl.Sink.CaptureDeleteAll || dpl.Sink.isCaptureDeleteEventEnvs(config.Env)
							var configRaw bytes.Buffer
							connectorConfig := buildSinkConnectorConfig(config.Org, config.Env, sink, schema)
							err := tpl.Execute(&configRaw, dpl)
							if err != nil {
								return nil, err
							}

							result[connectorConfig] = strings.TrimSpace(configRaw.String())
						}
					}
				}
			}
		}
	}
	return result, nil
}

func buildSinkConnectorConfig(org string, env string, sink *SinkConfig, schema string) ConnectorConfig {
	// KEEP old filename
	if schema == "public" {
		return ConnectorConfig{strings.ReplaceAll(sink.FileName, "SCHEMA.", ""), filepath.Join(org, env)}
	}
	return ConnectorConfig{strings.ReplaceAll(sink.FileName, "SCHEMA", schema), filepath.Join(org, env)}
}

func (p *DataPipelineParser) Export(dpls map[ConnectorConfig]string, outDir string) error {
	result := make([]string, 0)
	for connectorConfig, config := range dpls {
		_, _ = create(filepath.Join(outDir, connectorConfig.filePath, connectorConfig.fileName))
		err := os.WriteFile(filepath.Join(outDir, connectorConfig.filePath, connectorConfig.fileName), []byte(config), 0600)
		if err != nil {
			return err
		}
		result = append(result, path.Join(connectorConfig.filePath, connectorConfig.fileName))
	}
	sort.Strings(result)
	for _, res := range result {
		fmt.Println("write file ", res)
	}
	return nil
}

func (p *DataPipelineParser) mapToDeletedConnectorConfig(dpls map[ConnectorConfig]string, outDir string, savedMapFiles map[string]bool) {
	for connectorConfig := range dpls {
		savedMapFiles[filepath.Join(outDir, connectorConfig.filePath, connectorConfig.fileName)] = true
	}
}

type InteractFile interface {
	Remove(name string) error
	ReadDir(dirname string) ([]os.DirEntry, error)
}

type RealInteractFile struct{}

func (r *RealInteractFile) Remove(name string) error {
	return os.Remove(name)
}
func (r *RealInteractFile) ReadDir(dirname string) ([]os.DirEntry, error) {
	return os.ReadDir(dirname)
}

func DeleteConnectorNotExisted(r InteractFile, dpls map[string]bool, outDir string, excludes []ExcludeConfig) error {
	entries, err := r.ReadDir(outDir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			err := DeleteConnectorNotExisted(r, dpls, path.Join(outDir, e.Name()), excludes)
			if err != nil {
				return err
			}
		} else {
			currentFilePath := path.Join(outDir, e.Name())
			_, ok := dpls[currentFilePath]
			isIgnoreFile := ignoreWhenDeleteFile(excludes, currentFilePath)
			if !ok && !isIgnoreFile {
				fmt.Println("delete file ", currentFilePath, ok, isIgnoreFile)
				err := r.Remove(currentFilePath)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func create(p string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(p), 0770); err != nil {
		return nil, err
	}
	return os.Create(p)
}

// only use for env in local, stag, uat
var envAllowOrg = map[string][]string{
	"local": {"manabie", "e2e"},
	"stag":  {"manabie", "jprep"},
	"uat":   {"manabie", "jprep"},
}

func validateEnvAndOrg(env, org string) bool {
	if env == "prod" && org != "e2e" {
		return true
	}
	if env == "dorp" && org == "tokyo" {
		return true
	}
	// else env = [local, stag, uat]

	return golibs.InArrayString(org, envAllowOrg[env])
}
