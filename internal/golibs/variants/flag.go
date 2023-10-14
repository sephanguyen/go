package vr

import (
	"fmt"
	"sync"

	"github.com/manabie-com/backend/internal/golibs/execwrapper"

	"gopkg.in/yaml.v3"
)

type helmSubchartEnableFlags struct {
	Global map[string]struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"global"`
}

// UnmarshalYAML customizes yaml.Unmarshal for HelmSubchartEnableFlags.
func (f *helmSubchartEnableFlags) UnmarshalYAML(v *yaml.Node) error {
	tmpRes := struct {
		Global map[string]interface{} `yaml:"global"`
	}{}
	if err := v.Decode(&tmpRes); err != nil {
		return err
	}

	if f.Global == nil {
		f.Global = make(map[string]struct {
			Enabled bool "yaml:\"enabled\""
		})
	}
	for k, v := range tmpRes.Global {
		if !IsService(k) {
			continue
		}
		mapVal, ok := v.(map[string]interface{})
		if !ok {
			continue // not a map, skipping
		}
		statusVal, ok := mapVal["enabled"]
		if !ok {
			continue // key "enabled" not exist, skipping
		}
		status, ok := statusVal.(bool)
		if !ok {
			return fmt.Errorf(`"global.%s.enabled" field is not a valid bool (%s)`, k, statusVal)
		}
		f.Global[k] = struct {
			Enabled bool "yaml:\"enabled\""
		}{Enabled: status}
	}
	return nil
}

// isServiceEnabled returns true if service s is enabled in environment e of partner P.
func IsServiceEnabled(p P, e E, s S) (bool, error) {
	return isServiceEnabled(p, e, s)
}

// isServiceEnabled is the implementation of IsServiceEnabled.
// It is written in this form to add closure and protect serviceEnablementMap variable.
var isServiceEnabled = func() func(p P, e E, s S) (bool, error) {
	serviceEnablementMap := make(map[E]map[P]*helmSubchartEnableFlags)
	var serviceEnablementOnce sync.Once
	f := func(p P, e E, s S) (bool, error) {
		var err error
		serviceEnablementOnce.Do(func() { serviceEnablementMap, err = initServiceEnablementMap() })
		if err != nil {
			return false, fmt.Errorf("failed to initialize service enablement map: %s", err)
		}
		emap, ok := serviceEnablementMap[e]
		if !ok {
			return false, fmt.Errorf("environment %v does not exist", e)
		}
		epmap, ok := emap[p]
		if !ok {
			return false, fmt.Errorf("partner %v does not exist for environment %v", p, e)
		}
		epsmap, ok := epmap.Global[s.String()]
		if !ok {
			return false, fmt.Errorf("service %v does not exist in %v.%v", s, e, p)
		}
		return epsmap.Enabled, nil
	}
	return f
}()

func initServiceEnablementMap() (map[E]map[P]*helmSubchartEnableFlags, error) {
	res := make(map[E]map[P]*helmSubchartEnableFlags)
	for e, plist := range PartnerListByEnv() {
		for _, p := range plist {
			f, err := loadHelmValue[helmSubchartEnableFlags](p, e)
			if err != nil {
				return nil, err
			}
			if res[e] == nil {
				res[e] = make(map[P]*helmSubchartEnableFlags)
			}
			res[e][p] = f
		}
	}
	return res, nil
}

// HelmSubchart contains the configuration of the helm charts of a standard service.
type HelmSubchart struct {
	Enabled          bool `yaml:"enabled"`
	MigrationEnabled bool `yaml:"migrationEnabled"`
}

// NewHelmSubchart creates a new instance of helmSubchart from an input data.
func NewHelmSubchart(in map[string]interface{}) (*HelmSubchart, error) {
	res := &HelmSubchart{}
	if v, ok := in["migrationEnabled"]; ok {
		res.MigrationEnabled, ok = v.(bool)
		if !ok {
			return nil, fmt.Errorf(`"migrationEnabled" field is not a valid bool (%s)`, in["migrationEnabled"])
		}
	}
	return res, nil
}

func (h *HelmSubchart) Merge(in map[string]interface{}) error {
	if vi, ok := in["migrationEnabled"]; ok {
		v, ok := vi.(bool)
		if !ok {
			return fmt.Errorf(`"migrationEnabled" field is not a valid bool (%s)`, in["migrationEnabled"])
		}
		h.MigrationEnabled = v
	}
	return nil
}

// GetHelmSubchart returns the config for service s in environment e of partner p.
// There is cache so it is cheap to call this function multiple times.
func GetHelmSubchart(p P, e E, s S) (*HelmSubchart, error) {
	return getHelmSubchartConfig(p, e, s)
}

type helmSubchartMap map[S]*HelmSubchart

// UnmarshalYAML customizes yaml.Unmarshal for helmSubchartMap.
func (h *helmSubchartMap) UnmarshalYAML(v *yaml.Node) error {
	tmpRes := map[string]interface{}{}
	if err := v.Decode(tmpRes); err != nil {
		return err
	}

	if *h == nil {
		*h = make(map[S]*HelmSubchart)
	}

	for k, v := range tmpRes {
		s, err := ToServiceErr(k)
		if err != nil {
			continue // not a service, we skip this
		}
		if (*h)[s] == nil {
			(*h)[s] = &HelmSubchart{}
		}

		mapVal, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		if err := (*h)[s].Merge(mapVal); err != nil {
			return err
		}
	}

	return nil
}

var getHelmSubchartConfig = func() func(p P, e E, s S) (*HelmSubchart, error) {
	var m map[E]map[P]map[S]*HelmSubchart
	var once sync.Once
	return func(p P, e E, s S) (*HelmSubchart, error) {
		var err error
		once.Do(func() { m, err = initHelmSubchartConfigurationMap() })
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Helm subchart configs: %s", err)
		}

		emap, ok := m[e]
		if !ok {
			return nil, fmt.Errorf("environment %v does not exist", e)
		}
		epmap, ok := emap[p]
		if !ok {
			return nil, fmt.Errorf("partner %v does not exist for environment %v", p, e)
		}
		epsmap, ok := epmap[s]
		if !ok {
			return nil, fmt.Errorf("service %v does not exist in %v.%v", s, e, p)
		}
		return epsmap, nil
	}
}()

func initHelmSubchartConfigurationMap() (map[E]map[P]map[S]*HelmSubchart, error) {
	m := make(map[E]map[P]map[S]*HelmSubchart)
	for e, plist := range PartnerListByEnv() {
		for _, p := range plist {
			c, err := loadHelmValue[helmSubchartMap](p, e)
			if err != nil {
				return nil, err
			}
			if m[e] == nil {
				m[e] = make(map[P]map[S]*HelmSubchart)
			}
			m[e][p] = *c
		}
	}
	return m, nil
}

// loadHelmValue is convenience function that simulate how
// value files are loaded and overridden when running Helm.
//
// Note that this lacks implementation the config inside the subchart itself.
// Proceed with caution.
func loadHelmValue[T any](p P, e E) (*T, error) {
	var out T

	// TODO: reuse this global data
	globalValueFp := execwrapper.Abs("deployments/helm/manabie-all-in-one/values.yaml")
	if err := loadyaml(globalValueFp, &out); err != nil {
		return nil, err
	}
	valueFp := execwrapper.Absf("deployments/helm/manabie-all-in-one/%v-%v-values.yaml", e, p)
	if err := loadyaml(valueFp, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetHelmValues returns the Helm value for service S on environment E for partner P.
func GetHelmValues(p P, e E, s S) HelmValues {
	v, err := getHelmValues(p, e, s)
	if err != nil {
		panic(err)
	}
	return *v
}

// IsBackendServiceEnabled checks if service S is enabled on environment E for partner P.
func IsBackendServiceEnabled(p P, e E, s S) bool {
	return GetHelmValues(p, e, s).Enabled
}

var getHelmValues = func() func(p P, e E, s S) (*HelmValues, error) {
	res := make(map[E]map[P]map[S]*HelmValues)
	var once sync.Once
	return func(p P, e E, s S) (*HelmValues, error) {
		var err error
		once.Do(func() { res, err = initHelmValues() })
		if err != nil {
			return nil, err
		}
		resE, ok := res[e]
		if !ok {
			return nil, fmt.Errorf("environment %v does not exist", e)
		}
		resEP, ok := resE[p]
		if !ok {
			return nil, fmt.Errorf("partner %v does not exist for environment %v", p, e)
		}
		resEPS, ok := resEP[s]
		if !ok {
			return nil, fmt.Errorf("service %v does not exist in %v.%v", s, e, p)
		}
		return resEPS, nil
	}
}()

// HelmValues represents the YAML values (think `values.yaml`) for a specific Helm chart.
type HelmValues HelmSubchart

func initHelmValues() (map[E]map[P]map[S]*HelmValues, error) {
	res := make(map[E]map[P]map[S]*HelmValues)
	for _, eps := range AllEPS() {
		e := eps.E
		p := eps.P
		s := eps.S
		v, err := loadHelmValue2[HelmValues](p, e, s)
		if err != nil {
			return nil, fmt.Errorf("failed to load value for %v/%v/%v: %s", p, e, s, err)
		}
		if res[e] == nil {
			res[e] = make(map[P]map[S]*HelmValues)
		}
		if res[e][p] == nil {
			res[e][p] = make(map[S]*HelmValues)
		}
		res[e][p][s] = v
	}
	return res, nil
}

func loadHelmValue2[T any](p P, e E, s S) (*T, error) {
	var out T

	// default values, lowest precedence
	if err := loadyaml(execwrapper.Absf("deployments/helm/backend/%v/values.yaml", s), &out); err != nil {
		return nil, err
	}
	// global values
	if err := loadyaml(execwrapper.Abs("deployments/helm/backend/values.yaml"), &out); err != nil {
		return nil, err
	}
	// global, environment-specific values
	if err := loadyaml(execwrapper.Absf("deployments/helm/backend/%v-%v-values.yaml", e, p), &out); err != nil {
		return nil, err
	}
	// service-specific, environment-specific values, highest precedence
	if err := loadyaml(execwrapper.Absf("deployments/helm/backend/%v/%v-%v-values.yaml", s, e, p), &out); err != nil {
		return nil, err
	}
	return &out, nil
}
