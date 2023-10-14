package lessonmgmt

import (
	"context"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"
	commonpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
)

func (s *Suite) UserCreateALessonClassDoWithAllRequiredFields(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.ClassDoLink = "https://app.class.com/join?token=ABCDEFGHIJKLMNOP"
	stepState.ClassDoRoomID = idutil.ULIDNow()

	req := s.CommonSuite.UserCreateALessonRequestWithMissingFieldsInLessonmgmt(StepStateToContext(ctx, stepState), commonpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_CLASS_DO)

	return s.CommonSuite.UserCreateALessonWithRequestInLessonmgmt(ctx, req)
}

func (s *Suite) UserGetsALessonClassDo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// wait for sync process done
	time.Sleep(1 * time.Second)

	req := &cpb.GetLessonDetailOnCalendarRequest{
		LessonId: stepState.CurrentLessonID,
	}
	ctx = helper.GRPCContext(ctx, "token", stepState.AuthToken)
	stepState.Response, stepState.ResponseErr = cpb.NewLessonReaderServiceClient(s.CalendarConn).
		GetLessonDetailOnCalendar(ctx, req)

	return StepStateToContext(ctx, stepState), nil
}
