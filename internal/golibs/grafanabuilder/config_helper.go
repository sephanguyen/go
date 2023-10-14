package grafanabuilder

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"text/template"

	"github.com/manabie-com/backend/internal/golibs/grpc"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	fileio "github.com/manabie-com/backend/internal/golibs/io"
)

var (
	defaultMainFile      []byte
	onceConfig           sync.Once
	onceExtFiles         sync.Once
	defaultExtensionFile map[string]string

	// Grafana dashboard for Hasura services
	defaultHasuraConfig              []byte
	onceHasuraConfig                 sync.Once
	defaultHasuraConfigExtensionFile []string
	onceHasuraExtFiles               sync.Once
)

// GRPCDashboardConfig will return data of a dashboard config jsonnet and relate files
// see example grafana dashboard at "example-basic-dashboard.png"
type GRPCDashboardConfig struct {
	helper               *ExtensionFilesHelper
	haveExceptionMethods bool
	grpcMethods          []string
	protoFiles           []string
	customExtensionFiles map[string]string
	yPos                 int
	panels               []string // option panels
}

func NewGRPCDashboardConfig() *GRPCDashboardConfig {
	b := &GRPCDashboardConfig{
		helper:               &ExtensionFilesHelper{},
		customExtensionFiles: map[string]string{},
	}

	// add custom function map to fill into extension files
	b.helper.AddFuncMap(func(data *ExtensionFilesHelperData) template.FuncMap {
		var resMethods string
		switch {
		case len(data.ExceptionRegexMethods) != 0:
			resMethods = fmt.Sprintf("grpc_server_method!~\"%s\",", data.ExceptionRegexMethods)
		case len(data.ExceptionMethods) != 0:
			resMethods = fmt.Sprintf("grpc_server_method!~\"%s\",", strings.Join(data.ExceptionMethods.GRPCMethods(), "|"))
		case len(data.GRPCMethods) != 0:
			resMethods = fmt.Sprintf("grpc_server_method=~\"%s\",", strings.Join(data.GRPCMethods.GRPCMethods(), "|"))
		}
		return template.FuncMap{
			"AppKubernetesIOName": func() string {
				return strings.Join(data.SrvNames, "|")
			},
			"GRPCServerMethod": func() string {
				return resMethods
			},
			"GoGoroutinesPod": func() string {
				return fmt.Sprintf("^(%s).+", strings.Join(data.SrvNames, "|"))
			},
			"Title": func() string {
				if data.Pros != nil && len(data.Pros.Title) != 0 {
					return data.Pros.Title
				}
				return fmt.Sprintf("Dashboard is generated for %v service", strings.Join(data.SrvNames, ", "))
			},
			"UID": func() string {
				if data.Pros != nil && len(data.Pros.UID) != 0 {
					return data.Pros.UID
				}
				return "UID_" + idutil.ULIDNow()
			},
			"calculateYByHeight": func(height int) int {
				res := b.yPos
				b.yPos += height
				return res
			},
		}
	})
	return b
}

// AddUIDAndTitle receive uid of dashboard
// If uid is empty, it will be auto generated
func (b *GRPCDashboardConfig) AddUIDAndTitle(uid, title string) *GRPCDashboardConfig {
	b.helper.AddDashBoardProperties(&ExtensionFilesHelperProperties{
		UID:   uid,
		Title: title,
	})
	return b
}

func (b *GRPCDashboardConfig) AddServiceNames(serviceNames ...string) *GRPCDashboardConfig {
	b.helper.AddServiceNames(serviceNames)
	return b
}

// AddExceptionMethods receive a list exception Methods
// which using in queries of panels.
// if this exception Methods is set, it will replace list grpc method
func (b *GRPCDashboardConfig) AddExceptionMethods(exceptionMethods ...string) *GRPCDashboardConfig {
	b.helper.AddExceptionGRPCMethods(exceptionMethods)
	b.haveExceptionMethods = true
	return b
}

// AddExceptionRegexMethods receive an exception Regex Method
// which using in queries of panels.
// if this exception Regex Methods is set, it will replace exception Methods
func (b *GRPCDashboardConfig) AddExceptionRegexMethods(exceptionRegexMethods string) *GRPCDashboardConfig {
	b.helper.AddExceptionRegexMethods(exceptionRegexMethods)
	b.haveExceptionMethods = true
	return b
}

// AddGRPCMethods receive a list grpc method
// which using in queries of panels.
func (b *GRPCDashboardConfig) AddGRPCMethods(methods ...string) *GRPCDashboardConfig {
	b.grpcMethods = methods
	return b
}

func (b *GRPCDashboardConfig) AddProtoFile(protoFiles ...string) *GRPCDashboardConfig {
	b.protoFiles = protoFiles
	return b
}

// AddCustomExtensionFiles receive a map with key file name and value is path
func (b *GRPCDashboardConfig) AddCustomExtensionFiles(files map[string]string) *GRPCDashboardConfig {
	b.customExtensionFiles = files
	return b
}

func (b *GRPCDashboardConfig) AddRequestsPerSecondsByMethodsPanel() *GRPCDashboardConfig {
	panel := `.addPanel(
  graphPanel.new(
    title='Requests per seconds by methods',
    datasource=properties.datasource,
    legend_alignAsTable=true,
    legend_rightSide=true,
    legend_max=true,
    legend_min=true,
    legend_current=true,
    legend_avg=true,
    legend_sortDesc=true,
    legend_sort='max',
    legend_values=true,
    legend_sideWidth=350,
    pointradius=2,
    unit='short',
  ).resetYaxes().
  addYaxis(
    format='short',
  ).addYaxis(
    format='short',
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.RequestsPerSecondsByMethods.expr,
        legendFormat='{{ grpc_server_method }}',
    )
  ), gridPos=properties.gridPos[9]
)
`
	b.panels = append(b.panels, panel)
	return b
}

func (b *GRPCDashboardConfig) Build() (mainFileData io.Reader, extensionFiles map[string]io.Reader, err error) {
	if !b.haveExceptionMethods {
		if err = b.helper.AddGRPCMethods(b.grpcMethods); err != nil {
			return nil, nil, err
		}
		b.helper.AddGRPCMethodsFromProtoFiles(b.protoFiles)
	}

	// execute default extension files
	extFiles, err := b.getExtensionFiles()
	if err != nil {
		return nil, nil, fmt.Errorf("getExtensionFiles: %w", err)
	}
	if extensionFiles, err = b.executeExtensionFiles(extFiles); err != nil {
		return nil, nil, fmt.Errorf("executeExtensionFiles: %w", err)
	}

	// execute custom extension files after default extension files
	// because calculateYByHeight function in default extension files need run before
	if extFiles, err = b.getCustomExtensionFiles(); err != nil {
		return nil, nil, fmt.Errorf("getCustomExtensionFiles: %w", err)
	}
	custom, err := b.executeExtensionFiles(extFiles)
	if err != nil {
		return nil, nil, fmt.Errorf("could not custom execute extension files: %w", err)
	}
	for name := range custom {
		extensionFiles[name] = custom[name]
	}

	if mainFileData, err = b.getMainFileData(); err != nil {
		return nil, nil, fmt.Errorf("getMainFileData: %w", err)
	}

	return mainFileData, extensionFiles, nil

}

func (b *GRPCDashboardConfig) executeExtensionFiles(files map[string]io.Reader) (map[string]io.Reader, error) {
	out, err := b.helper.Execute(files)
	if err != nil {
		return nil, fmt.Errorf("could not execute extension files: %w", err)
	}
	for _, file := range files {
		file.(*os.File).Close()
	}

	return out, nil
}

func (b *GRPCDashboardConfig) getMainFileData() (*bytes.Buffer, error) {
	onceConfig.Do(func() {
		var err error
		defaultMainFile, err = fileio.ReadFileByRepoPath("internal/golibs/grafanabuilder/config/default_dashboard_cfg.jsonnet")
		if err != nil {
			panic(fmt.Errorf("could not read default_dashboard_cfg.jsonnet: %v", err))
		}
	})

	mainFile := defaultMainFile
	customPanels := make([]byte, 0)
	for _, p := range b.panels {
		customPanels = append(customPanels, p...)
	}
	if len(customPanels) != 0 {
		mainFile = append(mainFile, customPanels...)
	}
	buff := &bytes.Buffer{}
	if _, err := buff.Write(mainFile); err != nil {
		return nil, fmt.Errorf("could not load main file data: %w", err)
	}
	return buff, nil
}

func (b *GRPCDashboardConfig) getExtensionFiles() (map[string]io.Reader, error) {
	onceExtFiles.Do(func() {
		defaultExtensionFile = make(map[string]string)
		ext1, err := fileio.GetAbsolutePathFromRepoRoot("internal/golibs/grafanabuilder/config/panel_target.jsonnet")
		if err != nil {
			panic(fmt.Errorf("could not get absolute path of panel_target.jsonnet: %v", err))
		}
		defaultExtensionFile["panel_target.jsonnet"] = ext1

		ext2, err := fileio.GetAbsolutePathFromRepoRoot("internal/golibs/grafanabuilder/config/default_dashboard_properties.jsonnet")
		if err != nil {
			panic(fmt.Errorf("could not get absolute path of default_dashboard_properties.jsonnet: %v", err))
		}
		defaultExtensionFile["default_dashboard_properties.jsonnet"] = ext2
	})

	extFiles := make(map[string]io.Reader)
	for fileName, path := range defaultExtensionFile {
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("could not open extension file %w", err)
		}
		extFiles[fileName] = f
	}
	return extFiles, nil
}

func (b *GRPCDashboardConfig) getCustomExtensionFiles() (map[string]io.Reader, error) {
	extFiles := make(map[string]io.Reader)
	for fileName, path := range b.customExtensionFiles {
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("could not open custom extension file %w", err)
		}
		extFiles[fileName] = f
	}
	return extFiles, nil
}

// TODO: refactor later
func GetHasuraDashboard(uid, serviceName string) (io.Reader, map[string]io.Reader, error) {
	onceHasuraConfig.Do(func() {
		var err error
		defaultHasuraConfig, err = fileio.ReadFileByRepoPath("internal/golibs/grafanabuilder/config/hasura_dashboard_cfg.jsonnet")
		if err != nil {
			panic(fmt.Errorf("could not read hasura_dashboard_cfg.jsonnet: %v", err))
		}
	})
	if defaultHasuraConfig == nil {
		return nil, nil, fmt.Errorf("could not init default grafana dashboard config for hasura service")
	}

	onceHasuraExtFiles.Do(func() {
		ext1, err := fileio.GetAbsolutePathFromRepoRoot("internal/golibs/grafanabuilder/config/hasura_panel_target.jsonnet")
		if err != nil {
			panic(fmt.Errorf("could not get absolute path of hasura_panel_target.jsonnet: %v", err))
		}
		defaultHasuraConfigExtensionFile = append(defaultHasuraConfigExtensionFile, ext1)

		ext2, err := fileio.GetAbsolutePathFromRepoRoot("internal/golibs/grafanabuilder/config/hasura_dashboard_properties.jsonnet")
		if err != nil {
			panic(fmt.Errorf("could not get absolute path of hasura_dashboard_properties.jsonnet: %v", err))
		}
		defaultHasuraConfigExtensionFile = append(defaultHasuraConfigExtensionFile, ext2)
	})

	helper := &ExtensionFilesHelper{}
	helper.AddServiceNames([]string{serviceName})
	helper.AddFuncMap(func(data *ExtensionFilesHelperData) template.FuncMap {
		return template.FuncMap{
			"Service": func() string {
				return data.SrvNames[0]
			},
			"Title": func() string {
				if data.Pros != nil {
					return data.Pros.Title
				}
				return fmt.Sprintf("Dashboard is generated for %v service", strings.Join(data.SrvNames, ", "))
			},
			"UID": func() string {
				if len(uid) == 0 {
					return "UID_" + idutil.ULIDNow()
				}
				return uid
			},
		}
	})

	extendFiles := make([]io.Reader, 0, len(defaultHasuraConfigExtensionFile))
	for _, ext := range defaultHasuraConfigExtensionFile {
		f, err := os.Open(ext)
		if err != nil {
			return nil, nil, fmt.Errorf("could not open extension file %w", err)
		}
		extendFiles = append(extendFiles, f)
	}
	res, err := helper.Execute(map[string]io.Reader{"hasura_panel_target.jsonnet": extendFiles[0], "hasura_dashboard_properties.jsonnet": extendFiles[1]})
	if err != nil {
		return nil, nil, fmt.Errorf("could not execute file: %w", err)
	}
	buff := &bytes.Buffer{}
	if _, err = buff.Write(defaultHasuraConfig); err != nil {
		return nil, nil, err
	}
	for _, file := range extendFiles {
		file.(*os.File).Close()
	}

	return buff, res, nil
}

type ExtensionFilesHelperProperties struct {
	Title string
	UID   string
}

type ExtensionFilesHelperData struct {
	GRPCMethods           grpc.Services
	ExceptionMethods      grpc.Services
	ExceptionRegexMethods string
	SrvNames              []string
	MetricsName           []string // TODO: implement later

	Pros *ExtensionFilesHelperProperties
}

func (c *ExtensionFilesHelper) AddFuncMap(fn func(*ExtensionFilesHelperData) template.FuncMap) {
	c.userDefinedFuncMap = fn
}

func (c *ExtensionFilesHelper) getFuncMapFromUserDefined() error {
	helper := &grpc.Helper{}
	if len(c.gRPCMethods) != 0 {
		helper.AddItems(c.gRPCMethods)
		if len(c.protoFiles) != 0 {
			if err := helper.ValidateByProtoFiles(c.protoFiles); err != nil {
				return fmt.Errorf("ValidateByProtoFiles: %w", err)
			}
		}
	} else if len(c.protoFiles) != 0 {
		if err := helper.ParseFromProtoFiles(c.protoFiles); err != nil {
			return fmt.Errorf("could not parse proto files: %w", err)
		}
	}

	helper2 := &grpc.Helper{}
	if err := helper2.AddFullMethods(c.exceptionMethods); err != nil {
		return fmt.Errorf("invalid exception grpc methods: %w", err)
	}

	ud := c.userDefinedFuncMap(&ExtensionFilesHelperData{
		GRPCMethods:           helper.GRPCMethods(),
		ExceptionMethods:      helper2.GRPCMethods(),
		ExceptionRegexMethods: c.exceptionRegexMethods,
		SrvNames:              c.srvNames,
		Pros:                  c.pros,
	})
	if c.funcMap == nil {
		c.funcMap = make(template.FuncMap)
	}
	for k, v := range ud {
		c.funcMap[k] = v
	}
	return nil
}

type ExtensionFilesHelper struct {
	gRPCMethods           grpc.Services
	protoFiles            []string
	exceptionMethods      []string
	exceptionRegexMethods string
	srvNames              []string
	pros                  *ExtensionFilesHelperProperties

	// funcMap is used Execute config file,
	// with go template syntax
	funcMap            template.FuncMap
	userDefinedFuncMap func(*ExtensionFilesHelperData) template.FuncMap
}

func (c *ExtensionFilesHelper) AddDashBoardProperties(pros *ExtensionFilesHelperProperties) {
	c.pros = pros
}

func (c *ExtensionFilesHelper) AddServiceNames(services []string) {
	c.srvNames = services
}

// AddGRPCMethodsFromProtoFiles will receive list proto files,
// if input list grpc method was provided, these methods in these files is used to validate,
// otherwise if there are not any grpc method,them will be input by these proto files.
func (c *ExtensionFilesHelper) AddGRPCMethodsFromProtoFiles(protoFiles []string) {
	c.protoFiles = protoFiles
}

func (c *ExtensionFilesHelper) AddGRPCMethods(methods []string) error {
	if len(methods) == 0 {
		return nil
	}
	helper := &grpc.Helper{}
	if err := helper.AddFullMethods(methods); err != nil {
		return fmt.Errorf("invalid grpc method: %w", err)
	}
	c.gRPCMethods = helper.GRPCMethods()

	return nil
}

func (c *ExtensionFilesHelper) AddExceptionGRPCMethods(exceptionMethods []string) {
	c.exceptionMethods = exceptionMethods
}

func (c *ExtensionFilesHelper) AddExceptionRegexMethods(exceptionRegexMethods string) {
	c.exceptionRegexMethods = exceptionRegexMethods
}

func (c *ExtensionFilesHelper) Execute(templateCfgFile map[string]io.Reader) (map[string]io.Reader, error) {
	if err := c.getFuncMapFromUserDefined(); err != nil {
		return nil, fmt.Errorf("getFuncMapFromUserDefined: %w", err)
	}
	res := make(map[string]io.Reader)
	for name, cfg := range templateCfgFile {
		buf := &bytes.Buffer{}
		_, err := buf.ReadFrom(cfg)
		if err != nil {
			return nil, fmt.Errorf("could not read template config data: %w", err)
		}

		tmp, err := template.New("todos").Funcs(c.funcMap).
			Parse(buf.String())
		if err != nil {
			return nil, fmt.Errorf("could not parse template config data: %w", err)
		}

		out := &bytes.Buffer{}
		if err = tmp.Execute(out, nil); err != nil {
			return nil, fmt.Errorf("could not execute template config data %w", err)
		}
		res[name] = out
	}

	return res, nil
}

func (c *ExtensionFilesHelper) UseDefaultFuncMap() {
	c.funcMap["AppKubernetesIOName"] = func() string {
		return strings.Join(c.srvNames, "|")
	}
	c.funcMap["GoGoroutinesPod"] = func() string {
		return fmt.Sprintf("^(%s).+", strings.Join(c.srvNames, "|"))
	}
}
