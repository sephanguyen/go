package godogutil

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cucumber/godog"
)

func SetupStepDeadlockDebug(ctx *godog.ScenarioContext) {
	mlock := &sync.Mutex{}
	globmap := map[string]chan struct{}{}
	ctx.StepContext().Before(func(ctx context.Context, st *godog.Step) (context.Context, error) {
		mlock.Lock()
		thisChan := make(chan struct{})
		identifier := fmt.Sprintf("%s: %s", st.Id, st.Text)
		globmap[identifier] = thisChan
		mlock.Unlock()
		go func() {
			timer := time.NewTimer(45 * time.Second)
			select {
			case <-thisChan:
				return
			case <-timer.C:
				fmt.Printf("%s has been running for too long\n", identifier)
				return
			}
		}()
		return ctx, nil
	})

	ctx.StepContext().After(func(ctx context.Context, st *godog.Step, ret godog.StepResultStatus, err error) (context.Context, error) {
		mlock.Lock()
		escChan := globmap[fmt.Sprintf("%s: %s", st.Id, st.Text)]
		mlock.Unlock()
		escChan <- struct{}{}
		return ctx, err
	})
}
func SetupScenarioDeadlockDebug(ctx *godog.ScenarioContext) {
	mlock := &sync.Mutex{}
	globmap := map[string]chan struct{}{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		mlock.Lock()
		thisChan := make(chan struct{})
		identifier := fmt.Sprintf("%s: %s", sc.Id, sc.Name)
		globmap[identifier] = thisChan
		mlock.Unlock()
		go func() {
			timer := time.NewTimer(45 * time.Second)
			select {
			case <-thisChan:
				return
			case <-timer.C:
				fmt.Printf("%s has been running for too long\n", identifier)
				return
			}
		}()
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		mlock.Lock()
		escChan := globmap[fmt.Sprintf("%s: %s", sc.Id, sc.Name)]
		mlock.Unlock()
		escChan <- struct{}{}
		return ctx, err
	})
}
