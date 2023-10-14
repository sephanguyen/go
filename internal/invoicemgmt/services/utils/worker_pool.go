package utils

type WorkerPool struct {
	maxWorker int
	funcChan  chan func()
}

func NewWorkerPool(maxWorker int) *WorkerPool {
	return &WorkerPool{
		maxWorker: maxWorker,
		funcChan:  make(chan func()),
	}
}

func (wp *WorkerPool) Run() {
	for i := 0; i < wp.maxWorker; i++ {
		go func() {
			for f := range wp.funcChan {
				f()
			}
		}()
	}
}

func (wp *WorkerPool) AddTask(f func()) {
	wp.funcChan <- f
}

func (wp *WorkerPool) Close() {
	close(wp.funcChan)
}
