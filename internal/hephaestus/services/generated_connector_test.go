package services

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	dplparser "github.com/manabie-com/backend/cmd/utils/data_pipeline_parser"
	"github.com/manabie-com/backend/internal/golibs"
)

func TestCheckMissingTableDefineInGeneratedConnector(t *testing.T) {
	generatedConnectordir := "../../../deployments/helm/platforms/kafka-connect/postgresql2postgresql"
	generatedConnectorNames, err := loadGeneratedConnectorFile(generatedConnectordir)
	if err != nil {
		t.Error(err)
		return
	}

	connectorDir := "../../../deployments/helm/manabie-all-in-one/charts/hephaestus/connectors/sink"
	manuallyDefinedConnectorNames, err := loadManuallyDefinedConnector(connectorDir)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println("manuallyDefinedConnectorNames - generatedConnectorNames ")
	for i := range manuallyDefinedConnectorNames {
		if !golibs.InArrayString(manuallyDefinedConnectorNames[i], generatedConnectorNames) {
			fmt.Println(manuallyDefinedConnectorNames[i])
		}
	}

	fmt.Println("generatedConnectorNames - manuallyDefinedConnectorNames")
	for i := range generatedConnectorNames {
		if i >= len(manuallyDefinedConnectorNames) {
			break
		}
		if !golibs.InArrayString(generatedConnectorNames[i], manuallyDefinedConnectorNames) {
			fmt.Println(manuallyDefinedConnectorNames[i])
		}
	}
	// uncomment this line to show the output
	// t.Error("test")
}

func loadGeneratedConnectorFile(dir string) ([]string, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	filesNames := make([]string, 0)
	for _, e := range dirEntries {
		if strings.HasSuffix(e.Name(), "yaml") {
			currentPath := path.Join(dir, e.Name())
			parser, err := dplparser.NewDataPipelineParser(currentPath)
			if err != nil {
				return nil, err
			}
			for _, dpl := range parser.DataPipelineDef.Datapipelines {
				for _, sink := range dpl.Sinks {
					filesNames = append(filesNames, sink.FileName)
				}
			}
		}
	}
	return filesNames, nil
}

func loadManuallyDefinedConnector(dir string) ([]string, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	filesNames := make([]string, 0)
	for _, e := range dirEntries {
		if strings.HasSuffix(e.Name(), "json") {
			filesNames = append(filesNames, e.Name())
		}
	}

	return filesNames, nil
}
