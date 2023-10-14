package unleash

import (
	"context"
	"runtime"

	"github.com/pkg/errors"
)

var ErrManagerShutdown = errors.New(`ErrFeatureLockerShutdown`)

type ToggleChoice string

const (
	ToggleChoiceEnable  ToggleChoice = "enable"
	ToggleChoiceDisable              = "disable"
)

var (
	ErrInvalidToggleOption = errors.New("invalid toggle option")
)

type toggleRequest struct {
	ctx          context.Context
	featureName  string
	toggleChoice ToggleChoice
	result       chan error
}

type Manager interface {
	Toggle(ctx context.Context, featureName string, toggleChoice ToggleChoice) error
	Unlock(ctx context.Context, featureName string) error
	Shutdown()
}

type manager struct {
	unleashClient interface {
		ToggleUnleashFeatureWithName(ctx context.Context, featureName string, toggleChoice ToggleChoice) error
		IsFeatureToggleCorrect(ctx context.Context, featureName string, toggleSelect ToggleChoice) (bool, error)
	}
	featureLocker FeatureLocker

	toggleRequests chan *toggleRequest
	shutdown       chan struct{}

	numberOfWorker int
}

func NewManager(unleashSrvAddr string, unleashAPIKey string, unleashLocalAdminAPIKey string, managerOptions ...ManagerOption) Manager {
	client := NewDefaultClient(unleashSrvAddr, unleashAPIKey, unleashLocalAdminAPIKey)
	manager := &manager{
		unleashClient:  client,
		featureLocker:  NewDefaultFeatureLocker(),
		toggleRequests: make(chan *toggleRequest),
		shutdown:       make(chan struct{}),
	}

	for _, managerOption := range managerOptions {
		managerOption(manager)
	}

	if manager.numberOfWorker < 1 {
		manager.numberOfWorker = runtime.NumCPU()
	}

	for i := 0; i < manager.numberOfWorker; i++ {
		go manager.run()
	}

	return manager
}

func validToggleOption(toggleChoice ToggleChoice) error {
	switch toggleChoice {
	case ToggleChoiceEnable, ToggleChoiceDisable:
		return nil
	default:
		return ErrInvalidToggleOption
	}
}

func (manager *manager) Toggle(ctx context.Context, featureName string, toggleChoice ToggleChoice) error {
	if err := validToggleOption(toggleChoice); err != nil {
		return err
	}

	request := &toggleRequest{
		ctx:          ctx,
		featureName:  featureName,
		toggleChoice: toggleChoice,
		result:       make(chan error),
	}

	select {
	case manager.toggleRequests <- request:
		break
	case <-ctx.Done():
		return ctx.Err()
	case <-manager.shutdown:
		return ErrManagerShutdown
	}

	select {
	case err := <-request.result:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-manager.shutdown:
		return ErrManagerShutdown
	}
}

func (manager *manager) run() {
	for {
		select {
		case toggleRequest := <-manager.toggleRequests:
			select {
			case toggleRequest.result <- manager.processToggleRequest(toggleRequest.ctx, toggleRequest.featureName, toggleRequest.toggleChoice):
				continue
			case <-manager.shutdown:
				return
			}
		case <-manager.shutdown:
			return
		}
	}
}

func (manager *manager) processToggleRequest(ctx context.Context, featureName string, toggleChoice ToggleChoice) error {
	err := manager.featureLocker.Lock(ctx, featureName)
	if err != nil {
		return err
	}

	isToggleChoiceCorrect := false
	for !isToggleChoiceCorrect {
		err = manager.unleashClient.ToggleUnleashFeatureWithName(ctx, featureName, toggleChoice)
		if err != nil {
			_ = manager.featureLocker.Unlock(ctx, featureName)
			return err
		}
		isToggleChoiceCorrect, err = manager.unleashClient.IsFeatureToggleCorrect(ctx, featureName, toggleChoice)
		if err != nil {
			_ = manager.featureLocker.Unlock(ctx, featureName)
			return err
		}
	}

	return nil
}

func (manager *manager) Unlock(ctx context.Context, featureName string) error {
	return manager.featureLocker.Unlock(ctx, featureName)
}

func (manager *manager) Shutdown() {
	select {
	case <-manager.shutdown:
		return
	default:
		close(manager.shutdown)
		manager.featureLocker.Shutdown()
	}
}
