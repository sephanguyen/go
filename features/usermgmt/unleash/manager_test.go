package unleash

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type MockFeatureLocker struct {
	lockFn   func(ctx context.Context, featureName string) error
	unlockFn func(ctx context.Context, featureName string) error
}

func (m *MockFeatureLocker) Lock(ctx context.Context, featureName string) error {
	return m.lockFn(ctx, featureName)
}

func (m *MockFeatureLocker) Unlock(ctx context.Context, featureName string) error {
	return m.unlockFn(ctx, featureName)
}

func (m *MockFeatureLocker) Shutdown() {
}

type MockUnleashClient struct {
	toggleUnleashFeatureWithNameFunction func(ctx context.Context, featureName string, toggleChoice ToggleChoice) error
	isFeatureToggleCorrectFunctions      []func(ctx context.Context, featureName string, toggleSelect ToggleChoice) (bool, error)
}

func (m *MockUnleashClient) ToggleUnleashFeatureWithName(ctx context.Context, featureName string, toggleChoice ToggleChoice) error {
	return m.toggleUnleashFeatureWithNameFunction(ctx, featureName, toggleChoice)
}

func (m *MockUnleashClient) IsFeatureToggleCorrect(ctx context.Context, featureName string, toggleSelect ToggleChoice) (bool, error) {
	defer func() {
		m.isFeatureToggleCorrectFunctions = m.isFeatureToggleCorrectFunctions[1:]
	}()
	return m.isFeatureToggleCorrectFunctions[0](ctx, featureName, toggleSelect)
}

func TestManager_ToggleWithMockFeatureLockerAndMockUnleashClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	testcases := []struct {
		name              string
		initManager       func() *manager
		inputFeatureName  string
		inputToggleChoice ToggleChoice
		expectedResult    error
	}{
		{
			name: "lock feature, toggle feature, unlock feature successfully",
			initManager: func() *manager {
				manager := NewManager("", "", "").(*manager)
				manager.featureLocker = MockFeatureLockerAlwaysCanExecuteFnSuccessfully()
				manager.unleashClient = &MockUnleashClient{
					toggleUnleashFeatureWithNameFunction: func(ctx context.Context, featureName string, toggleChoice ToggleChoice) error {
						return nil
					},
					isFeatureToggleCorrectFunctions: []func(ctx context.Context, featureName string, toggleSelect ToggleChoice) (bool, error){
						func(ctx context.Context, featureName string, toggleSelect ToggleChoice) (bool, error) {
							return false, nil
						},
						func(ctx context.Context, featureName string, toggleSelect ToggleChoice) (bool, error) {
							return false, nil
						},
						func(ctx context.Context, featureName string, toggleSelect ToggleChoice) (bool, error) {
							return true, nil
						},
					},
				}
				return manager
			},
			inputFeatureName:  TestFeatureName,
			inputToggleChoice: ToggleChoiceEnable,
			expectedResult:    nil,
		},
		{
			name: "failed to lock feature",
			initManager: func() *manager {
				manager := NewManager("", "", "").(*manager)
				manager.featureLocker = &MockFeatureLocker{
					lockFn: func(ctx context.Context, featureName string) error {
						return assert.AnError
					},
				}
				return manager
			},
			inputFeatureName:  TestFeatureName,
			inputToggleChoice: ToggleChoiceEnable,
			expectedResult:    assert.AnError,
		},
		{
			name: "lock feature successfully but failed to toggle feature",
			initManager: func() *manager {
				manager := NewManager("", "", "").(*manager)
				manager.featureLocker = MockFeatureLockerAlwaysCanExecuteFnSuccessfully()
				manager.unleashClient = &MockUnleashClient{
					toggleUnleashFeatureWithNameFunction: func(ctx context.Context, featureName string, toggleChoice ToggleChoice) error {
						return assert.AnError
					},
				}
				return manager
			},
			inputFeatureName:  TestFeatureName,
			inputToggleChoice: ToggleChoiceEnable,
			expectedResult:    assert.AnError,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			manager := testcase.initManager()
			err := manager.Toggle(ctx, TestFeatureName, ToggleChoiceEnable)

			if testcase.expectedResult == nil {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, testcase.expectedResult.Error(), err.Error())
			}
		})
	}
}

func MockFeatureLockerAlwaysCanExecuteFnSuccessfully() *MockFeatureLocker {
	mockFeatureLocker := &MockFeatureLocker{
		lockFn: func(ctx context.Context, featureName string) error {
			return nil
		},
		unlockFn: func(ctx context.Context, featureName string) error {
			return nil
		},
	}
	return mockFeatureLocker
}

func MockUnleashAlwaysCanExecuteFnSuccessfully() *MockUnleashClient {
	mockUnleashClient := &MockUnleashClient{
		toggleUnleashFeatureWithNameFunction: func(ctx context.Context, featureName string, toggleChoice ToggleChoice) error {
			return nil
		},
		isFeatureToggleCorrectFunctions: []func(ctx context.Context, featureName string, toggleSelect ToggleChoice) (bool, error){
			func(ctx context.Context, featureName string, toggleSelect ToggleChoice) (bool, error) {
				return true, nil
			},
			func(ctx context.Context, featureName string, toggleSelect ToggleChoice) (bool, error) {
				return true, nil
			},
			func(ctx context.Context, featureName string, toggleSelect ToggleChoice) (bool, error) {
				return true, nil
			},
			func(ctx context.Context, featureName string, toggleSelect ToggleChoice) (bool, error) {
				return true, nil
			},
		},
	}
	return mockUnleashClient
}

func MockManagerAlwaysCanToggleSuccessfully(options ...ManagerOption) Manager {
	manager := NewManager("", "", "", options...).(*manager)
	manager.featureLocker = MockFeatureLockerAlwaysCanExecuteFnSuccessfully()
	manager.unleashClient = MockUnleashAlwaysCanExecuteFnSuccessfully()
	return manager
}

func TestManager_ToggleWhenManagerHasOnlyOneWorker(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	manager := MockManagerAlwaysCanToggleSuccessfully(NumberOfWorker(1))

	result := make(chan error, 1)
	go func() {
		result <- manager.Toggle(ctx, TestFeatureName, ToggleChoiceEnable)
	}()

	select {
	case err := <-result:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("manager should process this toggle request immediately")
	}
}

func TestManager_ToggleWithMockUnleashClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	manager := NewManager("", "", "", NumberOfWorker(2)).(*manager)
	manager.unleashClient = MockUnleashAlwaysCanExecuteFnSuccessfully()

	result := make(chan error, 1)
	go func() {
		result <- manager.Toggle(ctx, TestFeatureName, ToggleChoiceEnable)
	}()

	select {
	case err := <-result:
		assert.NoError(t, err)
	case <-ctx.Done():
		t.Fatal("expected manager process toggle request successfully")
	}
}

func TestManager_ToggleSameFeatureWithMockUnleashClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	manager := NewManager("", "", "", NumberOfWorker(2)).(*manager)
	manager.unleashClient = MockUnleashAlwaysCanExecuteFnSuccessfully()

	result1 := make(chan error, 1)
	unleashInProcess1Executing := make(chan struct{})
	unleashInProcess1Blocker := make(chan struct{})
	unleashInProcess1Unlocking := make(chan struct{})
	unleashInProcess1Unlocked := make(chan error)
	go func() {
		manager.unleashClient.(*MockUnleashClient).toggleUnleashFeatureWithNameFunction = func(ctx context.Context, featureName string, toggleChoice ToggleChoice) error {
			close(unleashInProcess1Executing)
			/*fmt.Println("close(unleashInProcess1Executing)")
			defer fmt.Println("close(unleashInProcess1Executed)")*/
			select {
			case <-unleashInProcess1Blocker:

			}
			return nil
		}

		result1 <- manager.Toggle(ctx, TestFeatureName, ToggleChoiceEnable)

		select {
		case <-unleashInProcess1Unlocking:
			unleashInProcess1Unlocked <- manager.Unlock(ctx, TestFeatureName)
		}
	}()

	result2 := make(chan error, 1)
	unleashInProcess2Executing := make(chan struct{})
	go func() {
		<-unleashInProcess1Executing
		manager.unleashClient.(*MockUnleashClient).toggleUnleashFeatureWithNameFunction = func(ctx context.Context, featureName string, toggleChoice ToggleChoice) error {
			close(unleashInProcess2Executing)
			/*fmt.Println("close(unleashInProcess2Executing)")
			defer fmt.Println("close(unleashInProcess2Executed)")*/
			return nil
		}

		// fmt.Println("unleashInProcess2 beforeExecuting")
		result2 <- manager.Toggle(ctx, TestFeatureName, ToggleChoiceEnable)
	}()

	select {
	case <-unleashInProcess1Executing:
	case <-ctx.Done():
		t.Fatal("expected toggling and locking progress in process 1 is executing")
	}

	select {
	case _ = <-result1:
		// expect process 1 doesn't receive result yet since the progress of toggling and locking doesn't complete
		t.Fatal("expected process 1 doesn't finish yet")
	case _ = <-result2:
		// expect process 2 doesn't receive result until the process 1 finished toggling and unlocking the feature
		t.Fatal("expected manager blocks toggle request of process 2")
	case <-ctx.Done():
		t.Fatal("test context is time out")
	case <-time.After(300 * time.Millisecond):
		break
	}

	// alert mock blocking progress in process 1 to finish so the
	// feature will be toggled and locked
	close(unleashInProcess1Blocker)
	// wait until process 1 toggled and locked the feature
	select {
	case err := <-result1:
		// process 1 should receive result now
		assert.NoError(t, err)
	case <-ctx.Done():
		t.Fatal("expected process 1 finish and get success result")
	}

	select {
	case _ = <-result2:
		// expect process 2 to be blocked until process 1 unlock the feature
		t.Fatal("expected manager blocks toggle request of process 2")
	case <-ctx.Done():
		t.Fatal("test context is time out")
	case <-time.After(500 * time.Millisecond):

	}

	// alert process 1 to unlock
	close(unleashInProcess1Unlocking)
	// wait until feature flag was completely unlocked by process 1
	select {
	case err := <-unleashInProcess1Unlocked:
		// expect process 1 unlock feature successfully
		assert.NoError(t, err)
	case <-ctx.Done():
		t.Fatal("expect process 1 unlock feature successfully")
	}

	select {
	case err := <-result2:
		// expect process 2 can receive result after process 1 unlocked the feature flag
		assert.NoError(t, err)
	case <-ctx.Done():
		t.Fatal("expect process 2 receive result after process 1 unlocked the feature")
	}
}
