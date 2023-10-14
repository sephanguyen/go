package vr

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func loadyaml(fp string, dest interface{}) error {
	data, err := os.ReadFile(fp)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %s", fp, err)
	}
	return yaml.Unmarshal(data, dest)
}
