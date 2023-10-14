package lessonmgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

func (s *Suite) exportCourseLocationSchedule(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &lpb.ExportCourseLocationScheduleRequest{}
	stepState.Response, stepState.ResponseErr = lpb.NewCourseLocationScheduleServiceClient(s.LessonMgmtConn).
		ExportCourseLocationSchedule(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnsCourseLocationScheduleInCsv(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("can not export zoom account: %s", stepState.ResponseErr.Error())
	}
	return StepStateToContext(ctx, stepState), nil
}
