package skaffoldwrapper

import (
	"os"

	"github.com/manabie-com/backend/internal/golibs/execwrapper"
)

// type BuildArgs struct {
// 	Filename string
// }

func Build(envs []string, args ...string) error {
	cmd, err := command("build", args...)
	if err != nil {
		return err
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = envs
	cmd.Dir = execwrapper.RootDirectory()
	return cmd.Run()
}

func Render(envs []string, args ...string) error {
	cmd, err := command("render", args...)
	if err != nil {
		return err
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = envs
	cmd.Dir = execwrapper.RootDirectory()
	return cmd.Run()
}

func Deploy(envs []string, args ...string) error {
	cmd, err := command("deploy", args...)
	if err != nil {
		return err
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = envs
	cmd.Dir = execwrapper.RootDirectory()
	return cmd.Run()
}
