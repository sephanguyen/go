package deployer

import (
	"fmt"
	"log"
	"sync"

	"github.com/manabie-com/backend/internal/golibs/logger"
)

type Executor struct {
	builds  map[string]*BuildPipeline
	deploys map[string]*DeployPipeline
	pulls   map[string]*PullImagePipeline

	firstToDeploy []*DeployPipeline
}

func NewExecutor(p ...PipelineI) (*Executor, error) {
	e := &Executor{
		builds:  make(map[string]*BuildPipeline),
		deploys: make(map[string]*DeployPipeline),
		pulls:   make(map[string]*PullImagePipeline),
	}
	for _, v := range p {
		if err := e.addPipeline(v); err != nil {
			return nil, err
		}
	}
	return e, nil
}

func (e *Executor) Do() error {
	todo := e.workStream()
	errChan := make(chan error, 1)
	defer close(errChan)

	var wg sync.WaitGroup
	for {
		select {
		case err := <-errChan:
			log.Printf("got error: %s", err)
			return err
		case p, ok := <-todo:
			if !ok {
				log.Printf("done work")
				wg.Wait()
				return nil
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				logger.Infof("executing")
				if err := p.Do(); err != nil {
					errChan <- err
				}
			}()
		}
	}
}

// AddPipeline basically builds up a graph of execution.
func (e *Executor) addPipeline(p PipelineI) error {
	switch o := p.(type) {
	case *PullImagePipeline:
		if _, exists := e.pulls[o.Name]; exists {
			return fmt.Errorf("image pull pipeline %s already exists", o.Name)
		}
		e.pulls[o.Name] = o
	case *BuildPipeline:
		if _, exists := e.builds[o.Filename]; exists {
			return fmt.Errorf("build pipeline %s already exists", o.Filename)
		}
		e.builds[o.Filename] = o
	case *DeployPipeline:
		if _, exists := e.deploys[o.Filename]; exists {
			return fmt.Errorf("deploy pipeline %s already exists", o.Filename)
		}
		e.deploys[o.Filename] = o

		if len(o.DependencieNames) == 0 {
			e.firstToDeploy = append(e.firstToDeploy, o)
		}

		for _, v := range o.DependencieNames {
			dep, ok := e.deploys[v]
			if !ok {
				return fmt.Errorf("cannot find any dependency name %s", v)
			}
			o.dependencies = append(o.dependencies, dep)
		}
	default:
		panic(fmt.Errorf("invalid pipeline type: %T", p))
	}
	return nil
}

func giveDefaultExecutor() (*Executor, error) {
	return NewExecutor(
		&BuildPipeline{Filename: "skaffold.manaverse.yaml"},
		&DeployPipeline{Filename: "skaffold.emulator.yaml"},
		&DeployPipeline{Filename: "skaffold.backbone.yaml", DependencieNames: []string{"skaffold.emulator.yaml"}},
		&DeployPipeline{Filename: "skaffold.cp-ksql-server.yaml", DependencieNames: []string{"skaffold.backbone.yaml"}},
		&DeployPipeline{Filename: "skaffold.data-warehouse.yaml", DependencieNames: []string{"skaffold.emulator.yaml"}},
		&DeployPipeline{Filename: "skaffold.manaverse.yaml", DependencieNames: []string{"skaffold.backbone.yaml"}},
	)
}

func (e *Executor) workStream() <-chan PipelineI {
	todo := make(chan PipelineI)
	go func() {
		defer close(todo)

		// first, stream all work that can be started right away
		for _, v := range e.pulls {
			logger.Infof("sending image pull %s to stream", v.Name)
			v.MarkInProgress()
			todo <- v
		}
		for _, v := range e.builds {
			logger.Infof("sending build %s to stream", v.Filename)
			v.MarkInProgress()
			todo <- v
		}
		for _, v := range e.deploys {
			if len(v.DependencieNames) == 0 {
				logger.Infof("sending deploy %s to stream", v.Filename)
				v.MarkInProgress()
				todo <- v
			}
		}

		// loop and check the rest
		for {
			hasWorkLeft := false
			for _, v := range e.deploys {
				if v.Status() != NotStarted {
					continue
				}

				hasWorkLeft = true
				if v.CanStart() {
					logger.Infof("sending deploy %s to stream", v.Filename)
					v.MarkInProgress()
					todo <- v
				}
			}

			if !hasWorkLeft {
				return
			}
		}
	}()
	return todo
}

// func (e *Executor) Execute() error {
// 	var wg sync.WaitGroup
// 	errChan := make(chan error, 2)
// 	defer close(errChan)

// 	wg.Add(2)
// 	go func() {
// 		defer wg.Done()
// 		if err := e.runAllBuilds(); err != nil {
// 			errChan <- err
// 		}
// 	}()
// 	go func() {
// 		defer wg.Done()
// 		if err := e.runAllDeploys(); err != nil {
// 			errChan <- err
// 		}
// 	}()
// 	wg.Wait()

// 	select {
// 	case err := <-errChan:
// 		return err
// 	default:
// 		return nil
// 	}
// }

// func (e *Executor) runAllBuilds() error {
// 	for _, v := range e.builds {
// 		if err := v.Do(); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (e *Executor) runAllDeploys() error {
// }
