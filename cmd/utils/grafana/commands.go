package grafana

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/manabie-com/backend/internal/golibs/grafanabuilder"
	fileio "github.com/manabie-com/backend/internal/golibs/io"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	customServices = map[string]bool{
		"virtualclassroom": true,
		"lessonmgmt":       true,
	}

	exceptions = map[string]bool{
		"aphelios":         true,
		"draft":            true,
		"enigma":           true,
		"invoicemgmt":      true,
		"kafka-connect":    true,
		"j4":               true,
		"nats":             true,
		"notificationmgmt": true,
		"shamir":           true,
		"unleash":          true,
		"zeus":             true,
		"nats-jetstream":   true,
		"elasticsearch":    true,
		"fink":             true,
		"hephaestus":       true,
	}
)

const destinationPath = "deployments/helm/platforms/monitoring/grafana/dashboards/feature-services"
const hasuraDestinationPath = "deployments/helm/platforms/monitoring/grafana/dashboards/hasura"

// RootCmd for mock command
var RootCmd = &cobra.Command{
	Use:   "grafana [command]",
	Short: "regenerate grafana dashboard",
}

func init() {
	RootCmd.AddCommand(
		newDefaultDashboardsCmd(),
		newGenVirtualClassroomCmd(),
		newGenLessonMgmtCmd(),
		newHasuraDashboardCmd(),
	)
}

func newDefaultDashboardsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "default",
		Short: "Generate grafana dashboard for services which not yet custom",
		RunE:  GenDefaultDashboards,
	}
}

func newHasuraDashboardCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hasura",
		Short: "Generate grafana dashboard for hasura services which not yet custom",
		RunE:  GenDefaultHasuraDashboards,
	}
}

func GenDefaultDashboards(cmd *cobra.Command, args []string) error {
	names, err := getFeatureServices()
	if err != nil {
		return fmt.Errorf("getFeatureServices: %w", err)
	}

	// remove services which had custom dashboard
	srv := make([]string, 0, len(names))
	for _, name := range names {
		if _, ok := customServices[name]; !ok {
			srv = append(srv, name)
		}
	}

	for _, name := range srv {
		if err := genBasicDashboard(
			fmt.Sprintf(destinationPath+"/backend-%s-gen.json", name),
			[]string{name},
			fmt.Sprintf("Dashboard is generated for %s service", name),
			nil,
		); err != nil {
			return fmt.Errorf("got error when generate Grafana dashboard for %s service: %w", name, err)
		}
	}

	return nil
}

func GenDefaultHasuraDashboards(cmd *cobra.Command, args []string) error {
	svs, err := getHasuraServices()
	if err != nil {
		return fmt.Errorf("getHasuraServices: %w", err)
	}

	// remove services which had custom dashboard
	for _, srv := range svs {
		serviceName := srv + "-hasura"
		if err := genHasuraDashboard(
			fmt.Sprintf(hasuraDestinationPath+"/%s-gen.json", serviceName),
			serviceName,
		); err != nil {
			return fmt.Errorf("got error when generate Grafana dashboard hasura for %s service: %w", serviceName, err)
		}
	}

	return nil
}

func getFeatureServices() ([]string, error) {
	allServices, err := listServicesFromHclFile()
	if err != nil {
		return nil, fmt.Errorf("listServicesFromHclFile: %w", err)
	}

	names := make([]string, 0, len(allServices))
	for _, srv := range allServices {
		if _, ok := exceptions[srv.Name]; !ok {
			names = append(names, srv.Name)
		}
	}

	return names, nil
}

func getHasuraServices() ([]string, error) {
	svs, err := listServicesFromHclFile()
	if err != nil {
		return nil, fmt.Errorf("listServicesFromHclFile: %w", err)
	}

	// remove services which had custom dashboard
	var res []string
	for _, srv := range svs {
		if srv.Hasura.Enabled {
			res = append(res, srv.Name)
		}
	}

	return res, nil
}

type Services struct {
	Name   string `yaml:"name"`
	Hasura struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"hasura"`
}

func listServicesFromHclFile() ([]Services, error) {
	path, err := fileio.GetAbsolutePathFromRepoRoot("deployments/decl/prod-defs.yaml")
	if err != nil {
		return nil, fmt.Errorf("could not find file: %w", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read: %w", err)
	}

	var svcDefList []Services
	if err := yaml.Unmarshal(content, &svcDefList); err != nil {
		return nil, fmt.Errorf("yaml.Unmarshal: %w", err)
	}

	return svcDefList, nil
}

type Config struct {
	UID string `json:"uid"`
}

func getCurrentUID(path string) (string, error) {
	cfgJSON, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("could not read %s: %w", path, err)
	}

	var uid string
	if cfgJSON != nil {
		data := &Config{}
		if err = json.Unmarshal(cfgJSON, data); err != nil {
			return "", err
		}
		uid = data.UID
	}

	return uid, nil
}

func genBasicDashboard(destinationPath string, services []string, title string, methods []string) error {
	destinationPath, err := fileio.GetAbsolutePathFromRepoRoot(destinationPath)
	if err != nil {
		return err
	}

	uid, err := getCurrentUID(destinationPath)
	if err != nil {
		return err
	}
	builder := grafanabuilder.
		NewGRPCDashboardConfig().
		AddServiceNames(services...).
		AddUIDAndTitle(uid, title)
	if len(methods) != 0 {
		builder = builder.AddGRPCMethods(methods...)
	} else {
		builder = builder.AddExceptionRegexMethods("grpc.health.v1.Health/Check|.+TopicIcon.+")
	}

	main, extensions, err := builder.Build()
	if err != nil {
		return fmt.Errorf("grafanabuilder.NewGRPCDashboardConfig: %w", err)
	}

	if err = (&grafanabuilder.Builder{}).
		AddDashboardConfigFiles(main, extensions).
		AddDestinationFilePath(destinationPath).
		Build(); err != nil {
		return fmt.Errorf("build: %w", err)
	}

	return nil
}

func genHasuraDashboard(relativePath, service string) error {
	path, err := fileio.GetAbsolutePathFromRepoRoot(relativePath)
	if err != nil {
		return err
	}
	uid, err := getCurrentUID(path)
	if err != nil {
		return err
	}

	main, exts, err := grafanabuilder.GetHasuraDashboard(uid, service)
	if err != nil {
		return fmt.Errorf("grafanabuilder.GetHasuraDashboard: %w", err)
	}
	if err = (&grafanabuilder.Builder{}).
		AddDashboardConfigFiles(main, exts).
		AddDestinationFilePath(path).
		Build(); err != nil {
		return fmt.Errorf("build: %w", err)
	}

	return nil
}

// pls use defaultDashBoardConfigName such as import default dashboard config in your custom config
const defaultDashBoardConfigName = "default_dashboard_cfg.jsonnet"

func genCustomDashboard(customDashboard string, extensionFiles map[string]string, destinationPath, title string, services []string, methods []string) error {
	customDashboard, err := fileio.GetAbsolutePathFromRepoRoot(customDashboard)
	if err != nil {
		return err
	}
	f, err := os.ReadFile(customDashboard)
	if err != nil {
		return fmt.Errorf("could not read custom dashboard config")
	}
	main := &bytes.Buffer{}
	_, err = main.Write(f)
	if err != nil {
		return fmt.Errorf("could not read custom dashboard config data")
	}

	destinationPath, err = fileio.GetAbsolutePathFromRepoRoot(destinationPath)
	if err != nil {
		return err
	}
	uid, err := getCurrentUID(destinationPath)
	if err != nil {
		return err
	}

	e, extensions, err := grafanabuilder.
		NewGRPCDashboardConfig().
		AddUIDAndTitle(uid, title).
		AddServiceNames(services...).
		AddGRPCMethods(methods...).
		AddCustomExtensionFiles(extensionFiles).
		Build()
	if err != nil {
		return fmt.Errorf("grafanabuilder.NewGRPCDashboardConfig: %w", err)
	}
	extensions[defaultDashBoardConfigName] = e

	err = (&grafanabuilder.Builder{}).
		AddDashboardConfigFiles(main, extensions).
		AddDestinationFilePath(destinationPath).
		Build()
	if err != nil {
		return err
	}
	return nil
}
