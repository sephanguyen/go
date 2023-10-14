package gandalf

import "time"

type Option func(*options)

type options struct {
	retryTime            int
	waitTimeBetweenRetry time.Duration
}

func WithRetryTime(retryTime int) Option {
	return func(o *options) {
		o.retryTime = retryTime
	}
}

func WithWaitTimeBetweenRetry(waitTimeBetweenRetry time.Duration) Option {
	return func(o *options) {
		o.waitTimeBetweenRetry = waitTimeBetweenRetry
	}
}

var DefaultOption = []Option{
	WithRetryTime(10),
	WithWaitTimeBetweenRetry(2 * time.Second),
}

func Execute(process func() error, opts ...Option) error {
	o := options{}
	for _, opt := range opts {
		opt(&o)
	}

	var count int
	var err error
	for count <= o.retryTime {
		err = process()
		if err == nil {
			return err
		}
		count++
		if count <= o.retryTime {
			time.Sleep(o.waitTimeBetweenRetry)
		}
	}
	return err
}
