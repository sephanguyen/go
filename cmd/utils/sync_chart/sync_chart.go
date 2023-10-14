package syncchart

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/execwrapper"

	cp "github.com/otiai10/copy"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func syncChart(_ *cobra.Command, _ []string) {
	syncAllCharts()
}
func syncAllCharts() {
	newPath := execwrapper.Abs("deployments/helm/backend")
	oldPath := execwrapper.Abs("deployments/helm/manabie-all-in-one")
	servicesInNewNamespaces, envOrgValues := GetMovedCharts(newPath)
	for _, serviceName := range servicesInNewNamespaces {
		// Get enabled from current values file in namespace backend
		newValuePath := filepath.Join(newPath, serviceName, "values.yaml")
		newValueYaml, err := ReadYamlFile(newValuePath)
		if err != nil {
			panic(err)
		}
		commonEnabled := newValueYaml["enabled"]
		// cleanup
		err = cleanDir(filepath.Join(newPath, serviceName))
		if err != nil {
			panic(err)
		}
		// Copy all from deployments/helm/manabie-all-in-one/charts/<service_name> into deployments/helm/backend/<service_name>
		err = cp.Copy(filepath.Join(oldPath, "charts", serviceName), filepath.Join(newPath, serviceName))
		if err != nil {
			panic(err)
		}
		// Update value files
		err = updateValues(oldPath, newPath, serviceName, envOrgValues, commonEnabled)
		if err != nil {
			panic(err)
		}
		// Update file in templates folder
		templatePath := filepath.Join(newPath, serviceName, "templates")
		files, err := os.ReadDir(templatePath)
		if err != nil {
			panic(err)
		}
		for _, file := range files {
			fileName := file.Name()
			if strings.Contains(fileName, ".yaml") {
				filePath := filepath.Join(templatePath, fileName)
				bs, err := os.ReadFile(filePath)
				if err != nil {
					panic(err)
				}
				if strings.Contains(string(bs), "util.app") {
					err = os.WriteFile(filePath, []byte("{{- include \"util.appWithToggle\" . -}}\n"), 0600)
					if err != nil {
						panic(err)
					}
				} else {
					var out []byte
					out = append([]byte("{{- if .Values.enabled -}}\n"), bs...)
					out = append(out, []byte("{{- end -}}\n")...)
					err = os.WriteFile(filePath, out, 0600)
					if err != nil {
						panic(err)
					}
				}
			}
		}
	}
}

func syncAllChartsManaE2ELocal(_ *cobra.Command, _ []string) {
	// copy all config from deployments/helm/manabie-all-in-one/charts/<service_name>/configs/manabie/local into deployments/helm/manabie-all-in-one/<service_name>/configs/e2e/local
	// copy all config from deployments/helm/backend/charts/<service_name>/configs/manabie/local into deployments/helm/backend/<service_name>/configs/e2e/local
	backend := execwrapper.Abs("deployments/helm/backend")
	manabieAllInOnce := execwrapper.Abs("deployments/helm/manabie-all-in-one")
	dataWarehouse := execwrapper.Abs("deployments/helm/data-warehouse")
	platforms := execwrapper.Abs("deployments/helm/platforms")
	chartFolders := []string{backend, manabieAllInOnce, dataWarehouse, platforms}
	services, _ := GetMovedCharts(backend)

	for _, chartFolder := range chartFolders {
		isSubChart := false
		if chartFolder == manabieAllInOnce {
			isSubChart = true
		} else {
			services, _ = GetMovedCharts(chartFolder)
		}

		for _, serviceName := range services {
			cpConfig2Secrets(chartFolder, serviceName, isSubChart)
		}
	}
}

func cpConfig2Secrets(folder string, serviceName string, isSubChart bool) {
	subChartFolder := ""
	if isSubChart {
		subChartFolder = "charts"
	}
	e2eConfig := filepath.Join(folder, subChartFolder, serviceName, "configs", "e2e", "local")
	_ = os.RemoveAll(e2eConfig)
	_ = cp.Copy(filepath.Join(folder, subChartFolder, serviceName, "configs", "manabie", "local"), e2eConfig)
	e2eSecrets := filepath.Join(folder, subChartFolder, serviceName, "secrets", "e2e", "local")
	_ = os.RemoveAll(e2eSecrets)
	_ = cp.Copy(filepath.Join(folder, subChartFolder, serviceName, "secrets", "manabie", "local"), e2eSecrets)
}

func updateValues(oldPath, newPath, serviceName string, envOrgValues []string, commonEnabled interface{}) error {
	for _, fileName := range envOrgValues {
		filePath := filepath.Join(newPath, serviceName, fileName)
		mapYaml, err := ReadYamlFile(filePath)
		if err != nil {
			return err
		}
		enabled := mapYaml["enabled"]
		oldFilePath := filepath.Join(oldPath, fileName)
		oldMapYaml, err := ReadYamlFile(oldFilePath)
		if err != nil {
			return err
		}
		serviceValue := oldMapYaml[serviceName]
		if serviceValue == nil {
			if enabled == nil {
				err = os.WriteFile(filePath, []byte(""), 0600)
				if err != nil {
					return err
				}
			} else {
				if enabled.(bool) {
					err = os.WriteFile(filePath, []byte("enabled: true"), 0600)
					if err != nil {
						return err
					}
				} else {
					err = os.WriteFile(filePath, []byte("enabled: false"), 0600)
					if err != nil {
						return err
					}
				}
			}
		} else {
			if enabled != nil {
				serviceValue.(map[string]interface{})["enabled"] = enabled
			}
			out, err := yaml.Marshal(serviceValue)
			if err != nil {
				return err
			}
			err = os.WriteFile(filePath, out, 0600)
			if err != nil {
				return err
			}
		}
	}
	var out map[string]interface{}
	commonPath := filepath.Join(oldPath, "values.yaml")
	commonYaml, err := ReadYamlFile(commonPath)
	if err != nil {
		return err
	}
	serviceValuePath := filepath.Join(oldPath, "charts", serviceName, "values.yaml")
	serviceYaml, err := ReadYamlFile(serviceValuePath)
	if err != nil {
		return err
	}
	serviceValue := commonYaml[serviceName]
	if serviceValue != nil {
		out = MergeMaps(serviceValue.(map[string]interface{}), serviceYaml)
	} else {
		out = serviceYaml
	}
	if commonEnabled != nil {
		out["enabled"] = commonEnabled
	}
	outYaml, err := yaml.Marshal(out)
	if err != nil {
		return err
	}
	newValuePath := filepath.Join(newPath, serviceName, "values.yaml")
	err = os.WriteFile(newValuePath, outYaml, 0600)
	if err != nil {
		return err
	}
	return nil
}

func ReadYamlFile(filePath string) (map[string]interface{}, error) {
	var mapYaml map[string]interface{}
	bs, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(bs, &mapYaml)
	if err != nil {
		return nil, err
	}
	return mapYaml, nil
}

// Copy from https://github.com/helm/helm/blob/main/pkg/cli/values/options.go#L108
func MergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = MergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

func GetMovedCharts(newPath string) ([]string, []string) {
	var servicesInNewNamespaces []string
	var envOrgValues []string

	charts, err := os.ReadDir(newPath)
	if err != nil {
		panic(err)
	}

	for _, chart := range charts {
		chartName := chart.Name()
		if chart.IsDir() {
			if chartName != "common" {
				servicesInNewNamespaces = append(servicesInNewNamespaces, chartName)
			}
		} else {
			if strings.Contains(chartName, "-values.yaml") {
				envOrgValues = append(envOrgValues, chartName)
			}
		}
	}
	return servicesInNewNamespaces, envOrgValues
}

func cleanDir(path string) error {
	elements, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, e := range elements {
		if e.IsDir() {
			err = os.RemoveAll(filepath.Join(path, e.Name()))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
