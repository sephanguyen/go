package utils

import "github.com/manabie-com/backend/internal/golibs/try"

func DoWithMaxRetry(fn func(attempt int) (retry bool, err error), maxRetries int) error {
	var err error
	var cont bool
	attempt := 1
	for {
		cont, err = fn(attempt)
		if !cont || err == nil {
			break
		}
		attempt++
		if attempt > maxRetries {
			return try.ErrMaxRetriesReached
		}
	}
	return err
}
