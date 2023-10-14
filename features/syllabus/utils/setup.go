package utils

import (
	"context"

	"github.com/manabie-com/backend/features/common"

	"go.uber.org/zap"
)

type Suite[T any] struct {
	StepState *T
	*common.Connections
	ZapLogger  *zap.Logger
	Cfg        *common.Config
	AuthHelper *AuthHelper
}

func NewEntitySuite[T any](stepState *T, connections *common.Connections, zapLogger *zap.Logger, cfg *common.Config, authHelper *AuthHelper) *Suite[T] {
	return &Suite[T]{
		StepState:   stepState,
		Connections: connections,
		ZapLogger:   zapLogger,
		Cfg:         cfg,
		AuthHelper:  authHelper,
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
