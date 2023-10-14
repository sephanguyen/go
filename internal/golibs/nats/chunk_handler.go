package nats

import (
	"sync"

	"go.uber.org/multierr"
)

type ChunkOpts struct {
	Start, End int
}

const (
	totalWorkers = 3
)

func ChunkHandler(total, batchSize int, handler func(start, end int) error) error {
	var (
		errChan = make(chan error)
		tasks   = make(chan *ChunkOpts, totalWorkers)
		wg      sync.WaitGroup
	)

	wg.Add(totalWorkers)

	worker := func() {
		defer wg.Done()

		for chunk := range tasks {
			err := handler(chunk.Start, chunk.End)
			if err != nil {
				errChan <- err
			}
		}
	}

	for i := 0; i < totalWorkers; i++ {
		go worker()
	}

	for i := 0; i < total; i += batchSize {
		max := i + batchSize
		if max > total {
			max = total
		}

		tasks <- &ChunkOpts{
			Start: i,
			End:   max,
		}
	}

	close(tasks)

	go func() {
		wg.Wait()
		close(errChan)
	}()

	var errC error
	for err := range errChan {
		if err == nil {
			continue
		}

		errC = multierr.Append(errC, err)
	}

	return errC
}
