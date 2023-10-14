package dplparser

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/spf13/cobra"
)

var (
	dataPipelinePath             string
	connectorConfigTplPath       string
	sourceConnectorConfigTplPath string
	tableSchemaDir               string

	outputDirs      []string
	excluded        []string
	deleteConnector bool
)

var RootCmd = &cobra.Command{
	Use:   "plparser",
	Short: "generate sync data pipeline",
	Run: func(cmd *cobra.Command, args []string) {
		err := runDataPipelineParser(dataPipelinePath, connectorConfigTplPath, sourceConnectorConfigTplPath, tableSchemaDir, outputDirs, excluded, deleteConnector)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.PersistentFlags().StringVarP(
		&dataPipelinePath,
		"data-pipeline-path",
		"f",
		"",
		"data pipeline file path",
	)

	RootCmd.PersistentFlags().StringVarP(
		&connectorConfigTplPath,
		"kafka-connect-config-template-path",
		"t",
		"./pipeline_template.txt",
		"Template for connector config in Kakfa Connect",
	)

	RootCmd.PersistentFlags().StringVarP(
		&sourceConnectorConfigTplPath,
		"kafka-connect-config-source-template-path",
		"k",
		"./pipeline_source_template.txt",
		"Template for source connector config in Kakfa Connect",
	)

	RootCmd.PersistentFlags().StringVarP(
		&tableSchemaDir,
		"table-schema",
		"s",
		"../../../mock/testing/testdata",
		"table schema definitions for current database",
	)

	RootCmd.PersistentFlags().StringArrayVarP(
		&outputDirs,
		"output",
		"o",
		[]string{"output-1", "output-2"},
		"export output generated kafka connect connector config to directory",
	)

	RootCmd.PersistentFlags().StringArrayVarP(
		&excluded,
		"excluded",
		"e",
		[]string{"local:org:source:sink", "local:org:source:sink"},
		"excluded gen connector config for specific env:org:source:sink",
	)

	RootCmd.PersistentFlags().BoolVarP(
		&deleteConnector,
		"delete-connector",
		"d",
		false,
		"delete file connector if not existed in yaml file",
	)
}

func runDataPipelineParser(dataPipelinePath, kafkaConnectConfigTplPath, kafkaConnectSourceConfigTplPath, tableSchemaDir string, outputDirs, excluded []string, deleteConnector bool) error {
	// read template file
	b, err := os.ReadFile(kafkaConnectConfigTplPath)
	if err != nil {
		return err
	}
	tpl := string(b)
	sourceCfg, err := os.ReadFile(kafkaConnectSourceConfigTplPath)
	if err != nil {
		return err
	}
	stpl := string(sourceCfg)

	err = walkFiles(dataPipelinePath, tpl, stpl, tableSchemaDir, outputDirs, excluded, deleteConnector)
	if err != nil {
		return err
	}

	return nil
}

func walkFiles(dir, tpl, stpl, tableSchemaDir string, outputDirs []string, excluded []string, deleteConnector bool) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	connectorFiles := make(map[string]bool)
	for _, e := range entries {
		currentPath := path.Join(dir, e.Name())
		fmt.Printf("generated connector's config base on file: %s \n", currentPath)
		if e.IsDir() {
			err := walkFiles(currentPath, tpl, stpl, tableSchemaDir, outputDirs, excluded, deleteConnector)
			if err != nil {
				return err
			}
		}

		// init Data Pipeline parser
		parser, err := NewDataPipelineParser(
			currentPath,
			WithTpl(tpl),
			WithTableSchemaDir(tableSchemaDir),
			WithExcluded(excluded),
		)
		if err != nil {
			return err
		}

		result, err := parser.Parse()
		if err != nil {
			return err
		}
		// fmt.Println(result)

		for _, outputDir := range outputDirs {
			parseSink := parser.Export(result, outputDir+"sink")
			if parseSink != nil {
				return parseSink
			}
		}

		// save files for check delete later
		for _, outputDir := range outputDirs {
			parser.mapToDeletedConnectorConfig(result, outputDir+"sink", connectorFiles)
		}

		// init Data Pipeline parser
		soureParser, err := NewDataPipelineParser(
			currentPath,
			WithTpl(stpl),
			WithTableSchemaDir(tableSchemaDir),
			WithExcluded(excluded),
		)
		if err != nil {
			return err
		}

		sourceResult, err := soureParser.ParseSource()
		if err != nil {
			return err
		}
		// fmt.Println(result)

		for _, outputDir := range outputDirs {
			err = soureParser.Export(sourceResult, outputDir+"source")
			if err != nil {
				return err
			}
		}
	}

	if deleteConnector {
		interactFile := &RealInteractFile{}
		for _, outputDir := range outputDirs {
			deletedSinkErr := DeleteConnectorNotExisted(interactFile, connectorFiles, path.Join(outputDir, "sink"), parseExcludes(excluded))
			if deletedSinkErr != nil {
				return deletedSinkErr
			}
		}
	}
	return nil
}
