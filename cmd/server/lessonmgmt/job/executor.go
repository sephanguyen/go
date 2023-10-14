package job

import "context"

type Executor interface {
	ExecuteJob(ctx context.Context, limit int, offSet int) error
	GetTotal(ctx context.Context) (int, error)
	PreExecute(ctx context.Context) error
}
