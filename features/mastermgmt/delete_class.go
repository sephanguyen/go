package mastermgmt

import (
	"context"
	"fmt"
	"time"

	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
)

func (s *suite) deleteClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = &mpb.DeleteClassRequest{
		ClassId: stepState.CurrentClassId,
	}
	ctx, err := s.subscribeEventClass(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.subscribe: %w", err)
	}
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = mpb.NewClassServiceClient(s.MasterMgmtConn).DeleteClass(contextWithToken(s, ctx), stepState.Request.(*mpb.DeleteClassRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) deletedClassProperly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var deletedAt time.Time
	query := "SELECT deleted_at FROM class WHERE class_id = $1"
	err := s.BobDBTrace.QueryRow(ctx, query, stepState.CurrentClassId).Scan(&deletedAt)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}
