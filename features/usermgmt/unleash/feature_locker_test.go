package unleash

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const TestFeatureName = "example-feature-name-1"
const TestFeatureName2 = "example-feature-name-2"

func MustSendLockRequest(t *testing.T, ctx context.Context, featureLocker *featureLocker) {
	switch err := featureLocker.Lock(ctx, TestFeatureName); err {
	case nil:
		assert.Equal(t, true, featureLocker.featureLock[TestFeatureName])
	default:
		t.Fatal("expected fail")
	}
}

func MustSendUnlockRequest(t *testing.T, ctx context.Context, featureLocker *featureLocker) {
	err := featureLocker.Unlock(ctx, TestFeatureName)
	assert.NoError(t, err)
	assert.Equal(t, false, featureLocker.featureLock[TestFeatureName])
}

// When feature lock is running
// Then process 1 can send lock feature lock request successfully and that feature must be locked
func TestFeatureLocker_SendLockRequest(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	featureLocker := NewDefaultFeatureLocker().(*featureLocker)
	defer featureLocker.Shutdown()
	go featureLocker.run()

	MustSendLockRequest(t, ctx, featureLocker)
}

// After the feature locker shutdown, all send operations to it must be failed
func TestFeatureLocker_SendLockRequestAfterFeatureLockerShutdown(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	featureLocker := NewDefaultFeatureLocker().(*featureLocker)
	defer featureLocker.Shutdown()
	go featureLocker.run()

	// Shutdown feature locker
	featureLocker.Shutdown()
	time.Sleep(100 * time.Millisecond)

	err := featureLocker.Lock(ctx, TestFeatureName)
	assert.NotNil(t, err)
	assert.Equal(t, false, featureLocker.featureLock[TestFeatureName])

	err = featureLocker.Unlock(ctx, TestFeatureName)
	assert.NotNil(t, err)
	assert.Equal(t, false, featureLocker.featureLock[TestFeatureName])
}

// When feature lock is running
// Then process 1 can send lock feature lock request successfully and that feature must be locked
// And process 1 can send unlock feature lock request successfully after that and that feature must be released
func TestFeatureLocker_SendLockRequestThenSendUnlockRequest(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	featureLocker := NewDefaultFeatureLocker().(*featureLocker)
	defer featureLocker.Shutdown()
	go featureLocker.run()

	MustSendLockRequest(t, ctx, featureLocker)

	MustSendUnlockRequest(t, ctx, featureLocker)
}

// When feature lock is running
// Then process 1 can send lock feature lock request successfully and that feature must be locked
// And process 2 is trying to lock that feature too but failed
func TestFeatureLocker_BothProcess1AndProcess2SendLockRequests(t *testing.T) {
	ctx := context.Background()

	featureLocker := NewDefaultFeatureLocker().(*featureLocker)
	defer featureLocker.Shutdown()
	go featureLocker.run()

	// process 1
	firstProcessDone := make(chan struct{}, 1)
	go func() {
		MustSendLockRequest(t, ctx, featureLocker)

		firstProcessDone <- struct{}{}
	}()
	select {
	case <-firstProcessDone:
		break
	case <-time.After(time.Second):
		t.Fatal("process 1 shouldn't be blocked when trying to lock a feature")
	}

	secondProcessDone := make(chan struct{}, 1)
	// simulate other processes are trying to send lock request after that feature is locked
	go func() {
		// this will block until other process unlock the feature
		_ = featureLocker.Lock(ctx, TestFeatureName)

		secondProcessDone <- struct{}{}
	}()

	select {
	case <-secondProcessDone:
		t.Fatal("process 2 should be blocked when trying to lock a feature that is already locked")
	case <-time.After(time.Second):
		break
	}

	assert.Equal(t, true, featureLocker.featureLock[TestFeatureName])
	assert.Equal(t, 1, len(featureLocker.pendingLockRequestsMap[TestFeatureName]))
}

// When feature lock is running
// Then process 1 can send lock feature lock request successfully and that feature must be locked
// And process 2 is trying to lock that feature too but blocked until process 1 unlock the feature
func TestFeatureLocker_SendMultipleLockRequests2(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	featureLocker := NewDefaultFeatureLocker().(*featureLocker)
	defer featureLocker.Shutdown()
	go featureLocker.run()

	// process 1
	firstProcessDone := make(chan struct{}, 1)
	go func() {
		MustSendLockRequest(t, ctx, featureLocker)

		firstProcessDone <- struct{}{}
	}()
	select {
	case <-firstProcessDone:
		break
	case <-time.After(time.Second):
		t.Fatal("process 1 shouldn't be blocked when trying to lock a feature")
	}

	// process 2
	secondProcessDone := make(chan struct{}, 1)
	go func() {
		err := featureLocker.Lock(ctx, TestFeatureName)
		assert.NoError(t, err)
		assert.Equal(t, true, featureLocker.featureLock[TestFeatureName])
		assert.Equal(t, 0, len(featureLocker.pendingLockRequestsMap[TestFeatureName]))

		secondProcessDone <- struct{}{}
	}()

	select {
	case <-secondProcessDone:
		t.Fatal("process 2 should be blocked when trying to lock a feature that is already locked")
	case <-time.After(time.Second):
		break
	}

	// process 1
	firstProcessDone = make(chan struct{}, 1)
	go func() {
		MustSendUnlockRequest(t, ctx, featureLocker)

		firstProcessDone <- struct{}{}
	}()
	select {
	case <-firstProcessDone:
		break
	case <-time.After(time.Second):
		t.Fatal("process 1 shouldn't be blocked when trying to lock a feature")
	}

	select {
	case <-secondProcessDone:
		MustSendUnlockRequest(t, ctx, featureLocker)
		break
	case <-time.After(time.Second):
		t.Fatal("process 2 shouldn't be blocked when trying to lock a feature")
	}
}
