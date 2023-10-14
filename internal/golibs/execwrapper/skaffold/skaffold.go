package skaffoldwrapper

import (
	"os/exec"

	"github.com/manabie-com/backend/internal/golibs/execwrapper"
)

const (
	skaffoldBinary   string = "skaffold"
	skaffoldv2Binary string = "skaffoldv2"
)

func command(verb string, args ...string) (*exec.Cmd, error) {
	skaffoldExe, err := execwrapper.LookPath(skaffoldBinary)
	if err != nil {
		return nil, err
	}
	return exec.Command(skaffoldExe, appendargs(verb, args...)...), nil
}

func command2(verb string, args ...string) (*exec.Cmd, error) {
	skaffoldExe, err := execwrapper.LookPath(skaffoldv2Binary)
	if err != nil {
		return nil, err
	}
	return exec.Command(skaffoldExe, appendargs(verb, args...)...), nil
}

func appendargs(first string, others ...string) []string {
	all := make([]string, 0, len(others)+1)
	all = append(all, first)
	all = append(all, others...)
	return all
}
