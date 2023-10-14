package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	syncchart "github.com/manabie-com/backend/cmd/utils/sync_chart"
	"github.com/manabie-com/backend/internal/golibs/execwrapper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/sumdb/dirhash"
	"gopkg.in/yaml.v3"
)

func mergeOldValues(basePath, serviceName string, envOrgValues []string) (map[string]interface{}, error) {
	baseValuesPath := filepath.Join(basePath, "charts", serviceName, "values.yaml")
	commonValuesPath := filepath.Join(basePath, "values.yaml")
	baseValues, err := syncchart.ReadYamlFile(baseValuesPath)
	if err != nil {
		return nil, err
	}
	commonValues, err := syncchart.ReadYamlFile(commonValuesPath)
	if err != nil {
		return nil, err
	}
	out := make(map[string]interface{}, len(envOrgValues))
	for _, v := range envOrgValues {
		envOrgValuePath := filepath.Join(basePath, v)
		envOrgValue, err := syncchart.ReadYamlFile(envOrgValuePath)
		if err != nil {
			return nil, err
		}
		if commonValues[serviceName] != nil {
			out[v] = syncchart.MergeMaps(commonValues[serviceName].(map[string]interface{}), baseValues)
		} else {
			out[v] = baseValues
		}
		if envOrgValue[serviceName] != nil {
			out[v] = syncchart.MergeMaps(out[v].(map[string]interface{}), envOrgValue[serviceName].(map[string]interface{}))
		}
	}
	return out, nil
}

func mergeNewValues(basePath, serviceName string, envOrgValues []string) (map[string]interface{}, error) {
	baseValuesPath := filepath.Join(basePath, serviceName, "values.yaml")
	baseValues, err := syncchart.ReadYamlFile(baseValuesPath)
	if err != nil {
		return nil, err
	}
	out := make(map[string]interface{}, len(envOrgValues))
	for _, v := range envOrgValues {
		envOrgValuePath := filepath.Join(basePath, serviceName, v)
		envOrgValue, err := syncchart.ReadYamlFile(envOrgValuePath)
		if err != nil {
			return nil, err
		}
		out[v] = syncchart.MergeMaps(baseValues, envOrgValue)
		_, ok := out[v].(map[string]interface{})["enabled"]
		if ok {
			delete(out[v].(map[string]interface{}), "enabled")
		}
	}
	return out, nil
}

func TestSyncCharts(t *testing.T) {
	t.Parallel()
	// Test value files
	newPath := execwrapper.Abs("deployments/helm/backend")
	oldPath := execwrapper.Abs("deployments/helm/manabie-all-in-one")
	servicesInNewNamespaces, envOrgValues := syncchart.GetMovedCharts(newPath)
	for _, serviceName := range servicesInNewNamespaces {
		newValue, err := mergeNewValues(newPath, serviceName, envOrgValues)
		if err != nil {
			panic(err)
		}
		oldValue, err := mergeOldValues(oldPath, serviceName, envOrgValues)
		if err != nil {
			panic(err)
		}
		newOut, _ := yaml.Marshal(newValue)
		oldOut, _ := yaml.Marshal(oldValue)
		assert.Equal(t, string(oldOut), string(newOut), "the values between 2 namespaces do not match, have you run `make sync-chart` ?")
	}
	// Test match configs and secrets files
	getDirHash := func(dirpath string) (string, error) {
		return dirhash.HashDir(dirpath, "", dirhash.DefaultHash)
	}
	for _, serviceName := range servicesInNewNamespaces {
		baseNewPath := filepath.Join(newPath, serviceName)
		baseOldPath := filepath.Join(oldPath, "charts", serviceName)
		items, err := os.ReadDir(baseNewPath)
		assert.NoError(t, err)
		for _, item := range items {
			itemName := item.Name()
			if item.IsDir() && itemName != "templates" {
				sourceHash, err := getDirHash(filepath.Join(baseOldPath, itemName))
				require.NoError(t, err)
				destHash, err := getDirHash(filepath.Join(baseNewPath, itemName))
				require.NoError(t, err)
				require.Equal(t, sourceHash, destHash, fmt.Sprintf("dir %s in %s hash mismatched, have you run `make sync-chart` ?", itemName, serviceName))
				require.NoError(t, err)
			}
		}
	}

}
