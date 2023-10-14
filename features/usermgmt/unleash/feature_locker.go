package unleash

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

var ErrFeatureLockerShutdown = errors.New(`ErrFeatureLockerShutdown`)

type FeatureLocker interface {
	Lock(ctx context.Context, featureName string) error
	Unlock(ctx context.Context, featureName string) error
	Shutdown()
}

type featureLocker struct {
	featureLock map[string]bool

	featureLockRequests   chan *featureLockRequest
	featureUnlockRequests chan *featureUnlockRequest

	pendingLockRequestsMap map[string][]*featureLockRequest

	shutdown chan struct{}
}

func NewDefaultFeatureLocker() FeatureLocker {
	featureLocker := &featureLocker{
		featureLock:            make(map[string]bool),
		featureLockRequests:    make(chan *featureLockRequest),
		featureUnlockRequests:  make(chan *featureUnlockRequest),
		pendingLockRequestsMap: make(map[string][]*featureLockRequest),
		shutdown:               make(chan struct{}),
	}
	go featureLocker.run()
	return featureLocker
}

func (featureLocker *featureLocker) run() {
	for {
		select {
		case lockRequest := <-featureLocker.featureLockRequests:
			if featureLocked := featureLocker.featureLock[lockRequest.featureName]; featureLocked {
				// other process is using, temporarily queue this request to process later
				featureLocker.pendingLockRequestsMap[lockRequest.featureName] = append(featureLocker.pendingLockRequestsMap[lockRequest.featureName], lockRequest)
				continue
			} else {
				// there is no request is using, lock by feature name
				select {
				case lockRequest.callback <- nil:
					featureLocker.featureLock[lockRequest.featureName] = true
				case <-time.After(5 * time.Second):
					continue
				case <-featureLocker.shutdown:
					return
				}
			}
		case unlockRequest := <-featureLocker.featureUnlockRequests:
			// Get pending requests by feature name
			waitingLockRequests, exist := featureLocker.pendingLockRequestsMap[unlockRequest.featureName]
			if exist && len(waitingLockRequests) > 0 {
				waitingLockRequest := waitingLockRequests[0]
				featureLocker.pendingLockRequestsMap[unlockRequest.featureName] = featureLocker.pendingLockRequestsMap[unlockRequest.featureName][1:]

				go func() {
					select {
					case <-time.After(time.Minute):
						unlockRequest.callback <- errors.New("")
					case featureLocker.featureLockRequests <- waitingLockRequest:
						return
					}
				}()
			}
			select {
			case unlockRequest.callback <- nil:
				featureLocker.featureLock[unlockRequest.featureName] = false
			case <-time.After(3 * time.Second):
				continue
			}
		case <-featureLocker.shutdown:
			return
		}
	}
}

type featureUnlockRequest struct {
	featureName string
	callback    chan error
}
type featureLockRequest struct {
	featureName string
	callback    chan error
}

func (featureLocker *featureLocker) Lock(ctx context.Context, featureName string) error {
	lockRequest := &featureLockRequest{
		featureName: featureName,
		callback:    make(chan error, 1),
	}

	// send request that attached callback
	select {
	case featureLocker.featureLockRequests <- lockRequest:
		break
	case <-ctx.Done():
		lockRequest.callback <- ctx.Err()
	case <-featureLocker.shutdown:
		return ErrFeatureLockerShutdown
	}

	// waiting for result will be returned to callback
	select {
	case err := <-lockRequest.callback:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-featureLocker.shutdown:
		return ErrFeatureLockerShutdown
	}
}

func (featureLocker *featureLocker) Unlock(ctx context.Context, featureName string) error {
	unlockRequest := &featureUnlockRequest{
		featureName: featureName,
		callback:    make(chan error, 1),
	}

	// send request that attached callback
	select {
	case featureLocker.featureUnlockRequests <- unlockRequest:
		break
	case <-ctx.Done():
		unlockRequest.callback <- ctx.Err()
	case <-featureLocker.shutdown:
		return ErrFeatureLockerShutdown
	}

	// waiting for result will be returned to callback
	select {
	case err := <-unlockRequest.callback:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-featureLocker.shutdown:
		return ErrFeatureLockerShutdown
	}
}

func (featureLocker *featureLocker) Shutdown() {
	select {
	case <-featureLocker.shutdown:
		// this prevents panic when Shutdown func is called multiple times
		return
	default:
		close(featureLocker.shutdown)
	}
}
