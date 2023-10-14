package vr

import (
	"fmt"
	"sync"

	"github.com/manabie-com/backend/internal/golibs/execwrapper"
)

// TFService represents a service whose resources will be
// created by terraform.
type TFService struct {
	Name string `yaml:"name"`
}

// TFServiceDefs contains a list of TerraformServiceDef.
type TFServiceDefs []TFService

// TFServiceDefinitions returns the TFServiceDefs for environment e.
// The result is cached so it is cheap to call multiple times.
// It is safe for concurrent use.
func TFServiceDefinitions(e E) (TFServiceDefs, error) {
	return tfServiceDefinitions(e)
}

var tfServiceDefinitions = func() func(e E) (TFServiceDefs, error) {
	loadTFDef := func(e E) (TFServiceDefs, error) {
		fp := execwrapper.Absf("deployments/decl/%v-defs.yaml", e)
		out := TFServiceDefs{}
		if err := loadyaml(fp, &out); err != nil {
			return nil, err
		}
		return out, nil
	}

	stagDef := TFServiceDefs{}
	uatDef := TFServiceDefs{}
	prodDef := TFServiceDefs{}
	var stagOnce, uatOnce, prodOnce sync.Once
	return func(e E) (TFServiceDefs, error) {
		var err error
		switch e {
		case EnvStaging:
			stagOnce.Do(func() { stagDef, err = loadTFDef(e) })
			return stagDef, err
		case EnvUAT:
			uatOnce.Do(func() { uatDef, err = loadTFDef(e) })
			return uatDef, err
		case EnvProduction:
			prodOnce.Do(func() { prodDef, err = loadTFDef(e) })
			return prodDef, err
		default:
			return nil, fmt.Errorf("invalid environment: %v", e)
		}
	}
}()
