package tests

import (
	"fmt"
	"path/filepath"

	"github.com/manabie-com/backend/internal/golibs/execwrapper"
	vr "github.com/manabie-com/backend/internal/golibs/variants"
)

type ElasticKibana struct {
	Username string `yaml:"elasticsearch.username"`
	Password string `yaml:"elasticsearch.password"`
}

func (ElasticKibana) Path(p vr.P, e vr.E, _ vr.S) string {
	return filepath.Join(
		execwrapper.RootDirectory(),
		fmt.Sprintf("deployments/helm/platforms/elastic/secrets/%v/%v/kibana.encrypted.yaml", p, e),
	)
}

type ElasticRootCA struct {
	Data string `yaml:"data"`
}

func (ElasticRootCA) Path(p vr.P, e vr.E, _ vr.S) string {
	return filepath.Join(
		execwrapper.RootDirectory(),
		fmt.Sprintf("deployments/helm/platforms/elastic/secrets/%v/%v/root-ca.pem.encrypted.yaml", p, e),
	)
}

type ElasticRootCAKey struct {
	Data string `yaml:"data"`
}

func (ElasticRootCAKey) Path(p vr.P, e vr.E, _ vr.S) string {
	return filepath.Join(
		execwrapper.RootDirectory(),
		fmt.Sprintf("deployments/helm/platforms/elastic/secrets/%v/%v/root-ca-key.pem.encrypted.yaml", p, e),
	)
}
