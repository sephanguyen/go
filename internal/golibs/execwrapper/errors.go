package execwrapper

import (
	"errors"
	"fmt"
	"os/exec"
)

func binaryNotFound(name string) error {
	return fmt.Errorf("unable to find %s executable in PATH; please install %s before retrying", name, name)
}

// LookPath is similar to exec.LookPath, but have a slightly different error handlings.
func LookPath(file string) (string, error) {
	exe, err := exec.LookPath(file)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", binaryNotFound(file)
		}
		return "", err
	}
	return exe, nil
}
