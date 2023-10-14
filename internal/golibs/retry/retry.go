package retry

import (
	"crypto/rand"
	"math/big"
	"time"
)

func Retry[T any](attempts int, sleep time.Duration, f func() (T, error)) (T, error) {
	data, err := f()
	if err != nil {
		if s, ok := err.(Stop); ok {
			// Return the original error for later checking
			return Zero[T](), s.error
		}

		if attempts--; attempts > 0 {
			n, _ := rand.Int(rand.Reader, big.NewInt(int64(sleep)))

			// Add some randomness to prevent creating a Thundering Herd
			jitter := time.Duration(n.Int64())
			sleep += jitter / 2

			time.Sleep(sleep)
			return Retry(attempts, 2*sleep, f)
		}
		return Zero[T](), err
	}

	return data, nil
}

func Zero[T any]() T {
	var zero T
	return zero
}

type Stop struct {
	error
}

func NewStop(err error) Stop {
	return Stop{err}
}
