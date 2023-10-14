package deployer

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/manabie-com/backend/internal/golibs/execwrapper"
	skaffoldwrapper "github.com/manabie-com/backend/internal/golibs/execwrapper/skaffold"
	"github.com/manabie-com/backend/internal/golibs/logger"
)

type PipelineStatus int

const (
	NotStarted PipelineStatus = iota
	InProgress
	Done
)

type PipelineI interface {
	Do() error
	CanStart() bool
}

type BuildPipeline struct {
	Filename string

	buildOutput string

	pipelineStatus
}

func (p *BuildPipeline) Do() error {
	p.buildOutput = filepath.Join(os.TempDir(), fmt.Sprintf("%s-skaffold-build", p.Filename))
	args := []string{
		"-f", p.Filename,
	}
	args = append(args, defaultBuildArgs...)
	log.Printf("building with args: %+v", args)
	if err := skaffoldwrapper.Build(defaultEnvVar, args...); err != nil {
		return err
	}
	p.MarkDone()
	return nil
}

func (p *BuildPipeline) CanStart() bool {
	return true
}

func (p *BuildPipeline) BuildOutput() string {
	if p.Status() != Done {
		panic("BuildOutput: cannot call when not done")
	}
	return p.buildOutput
}

type DeployPipeline struct {
	Filename         string
	BuildRequired    bool
	DependencieNames []string

	dependencies []*DeployPipeline

	pipelineStatus
}

func (p *DeployPipeline) Do() error {
	args := []string{
		"-f", p.Filename,
	}
	args = append(args, defaultDeployArgs...)

	// hack for `skaffold deploy skaffold.manaverse.yaml`
	if p.Filename == "skaffold.manaverse.yaml" {
		args = append(args, "--images", "asia.gcr.io/student-coach-e1e95/backend:locally")
	}

	log.Printf("deploying with args: %+v", args)
	if err := skaffoldwrapper.Deploy(defaultEnvVar, args...); err != nil {
		return err
	}
	p.MarkDone()
	return nil
}

func (p *DeployPipeline) CanStart() bool {
	for _, v := range p.dependencies {
		if v.Status() != Done {
			return false
		}
	}
	return true
}

type PullImagePipeline struct {
	Name                   string
	LocalRegistryImages    []string
	ArtifactRegistryImages []string

	pipelineStatus
}

func (p *PullImagePipeline) Do() error {
	dockerPull := func(imageName string) error {
		dockerBinary, err := execwrapper.LookPath("docker")
		if err != nil {
			return err
		}
		dockerCmd := exec.Command(dockerBinary,
			"exec", "kind-control-plane", "crictl", "pull", "--creds", fmt.Sprintf("oauth2accesstoken:%s", os.Getenv("AR_ACCESS_TOKEN")),
			imageName,
		)
		dockerCmd.Stderr = os.Stderr
		dockerCmd.Stdout = os.Stdout
		dockerCmd.Env = defaultEnvVar
		dockerCmd.Dir = execwrapper.RootDirectory()
		return dockerCmd.Run()
	}

	// const localRegistryRepo := "localhost:5001"
	localRegistryRepo, exists := os.LookupEnv("LOCAL_REGISTRY_DOMAIN")
	if !exists {
		panic("env var LOCAL_REGISTRY_DOMAIN must be set")
	}

	arRepo, ok := os.LookupEnv("ARTIFACT_REGISTRY_DOMAIN")
	if !ok {
		panic("env var ARTIFACT_REGISTRY_DOMAIN must be set")
	}

	for _, imageName := range p.LocalRegistryImages {
		fullImageName := localRegistryRepo + "/" + imageName
		logger.Infof("pulling image %s", fullImageName)
		if err := dockerPull(fullImageName); err != nil {
			return fmt.Errorf("failed to pull %s: %s", fullImageName, err)
		}
	}

	for _, imageName := range p.ArtifactRegistryImages {
		fullImageName := arRepo + "/" + imageName
		logger.Infof("pulling image %s", fullImageName)
		if err := dockerPull(fullImageName); err != nil {
			return fmt.Errorf("failed to pull %s: %s", fullImageName, err)
		}
	}
	return nil
}

func (p *PullImagePipeline) CanStart() bool {
	return true
}

type pipelineStatus struct {
	v  PipelineStatus
	mu sync.Mutex
}

func (s *pipelineStatus) Status() PipelineStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.v
}

func (s *pipelineStatus) MarkInProgress() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.v = InProgress
}

func (s *pipelineStatus) MarkDone() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.v = Done
}
