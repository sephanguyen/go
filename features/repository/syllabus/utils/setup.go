package utils

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"

	"go.uber.org/zap"
)

type Suite[T any] struct {
	StepState      *T
	DB             database.Ext
	BobDBTrace     *database.DBTrace
	DBTrace        *database.DBTrace
	ZapLogger      *zap.Logger
	HasuraAdminURL string
	HasuraPassword string
}

func NewEntitySuite[T any](stepState *T, db database.Ext, bobDBTrace *database.DBTrace, zapLogger *zap.Logger, hasuraURL, hasuraPwd string) *Suite[T] {
	return &Suite[T]{
		StepState: stepState,
		DB:        db,
		DBTrace: &database.DBTrace{
			DB: db,
		},
		ZapLogger:      zapLogger,
		HasuraAdminURL: hasuraURL,
		HasuraPassword: hasuraPwd,
		BobDBTrace:     bobDBTrace,
	}
}

type StepStateKey struct{}

func AppendSteps(dest, src map[string]interface{}) {
	for k, v := range src {
		dest[k] = v
	}
}

func StepStateFromContext[StepState any](ctx context.Context) *StepState {
	return ctx.Value(StepStateKey{}).(*StepState)
}

func StepStateToContext[StepState any](ctx context.Context, state *StepState) context.Context {
	return context.WithValue(ctx, StepStateKey{}, state)
}
