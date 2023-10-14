package deployer

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/ci/env"
	skaffoldwrapper "github.com/manabie-com/backend/internal/golibs/execwrapper/skaffold"
	"github.com/manabie-com/backend/internal/golibs/logger"
)

var defaultEnvVar = env.Default().ToEnv()

func DoDeploy(args []string) error {
	if len(args) > 0 {
		command := args[0]
		switch command { //nolint
		case "render":
			if err := SkaffoldRender(); err != nil {
				return fmt.Errorf("skaffoldRender: %s", err)
			}
			return nil
		}
	}
	if err := SkaffoldBuild(); err != nil {
		return fmt.Errorf("skaffoldBuild: %s", err)
	}
	if err := SkaffoldDeploy(); err != nil {
		return fmt.Errorf("skaffoldRun: %s", err)
	}
	return nil
}

var defaultBuildArgs = []string{
	"--default-repo=asia.gcr.io/student-coach-e1e95",
}

var defaultDeployArgs = []string{
	"--status-check=false",
	"--default-repo=asia.gcr.io/student-coach-e1e95",
}

func SkaffoldRender(args ...string) error {
	logger.Infof("start rendering")
	if err := skaffoldwrapper.Render(defaultEnvVar, args...); err != nil {
		return err
	}

	logger.Infof("done rendering")
	return nil
}

func SkaffoldBuild() error {
	logger.Infof("start building")
	buildArgs := []string{"-f", "skaffold.local.yaml"}
	buildArgs = append(buildArgs, defaultBuildArgs...)
	err := skaffoldwrapper.Build(defaultEnvVar, buildArgs...)
	if err != nil {
		return err
	}
	logger.Infof("done building")
	return nil
}

func SkaffoldDeploy() error {
	logger.Infof("start deploying")
	deployArgs := []string{"-f", "skaffold.local.yaml"}
	deployArgs = append(deployArgs, defaultDeployArgs...)
	err := skaffoldwrapper.Deploy(defaultEnvVar, deployArgs...)
	if err != nil {
		return err
	}
	logger.Infof("done deploying")
	return nil
}

func SkaffoldRun() error {
	logger.Infof("starting skaffold run...")
	e, err := giveDefaultExecutor()
	if err != nil {
		return err
	}
	return e.Do()
}
