package timesheet

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/constants"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"google.golang.org/protobuf/proto"
)

func (s *Suite) timesheetSendEventLockLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	msg := &pb.TimesheetLessonLockEvt{
		LessonIds: stepState.TimesheetLessonIDs,
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	_, stepState.ResponseErr = s.JSM.PublishAsyncContext(ctx, constants.SubjectTimesheetLesson, data)

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) timesheetEventLockLessonPublishedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	return StepStateToContext(ctx, stepState), nil
}
