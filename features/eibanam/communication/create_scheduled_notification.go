package communication

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/godogutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) clicksButton(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	noti := stepState.notification
	return s.upsertNotification(ctx, noti)
}

func (s *suite) fillsScheduledNotificationInformation(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	noti, err := s.newNotification(stepState.schoolID, stepState.profile.schoolAdmin.id)
	if err != nil {
		return ctx, err
	}
	scheduledAt := time.Now().Truncate(time.Minute).Add(time.Minute)
	noti = s.notificationWithScheduledAt(scheduledAt, noti)
	noti = s.notificationWithReceiver([]string{stepState.getID("student")}, noti)
	stepState.notification = noti
	noti.TargetGroup.CourseFilter.CourseIds = []string{"fake_course_id"}
	noti.TargetGroup.CourseFilter.Type = cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) hasCreatedAStudentWithGradeCourseAndParentInfo(ctx context.Context) (context.Context, error) {
	var err error
	ctx, err = godogutil.MultiErrChain(ctx,
		s.hasCreatedCourse, "school admin", 1,
		s.hasCreatedAStudentWithGradeAndParentInfo, "school admin",
		s.hasAddedCreatedCourseForStudent, "school admin",
	)
	return ctx, err
}

func (s *suite) hasOpenedComposeNewNotificationFullscreenDialog(ctx context.Context, role string) (context.Context, error) {
	return ctx, nil
}

func (s *suite) seesNewScheduledNotificationOnCMS(ctx context.Context) (context.Context, error) {
	return s.storeNotificationSuccessfully(ctx)
}

func (s *suite) hasCreatedADraftNotification(ctx context.Context, arg1 string) (context.Context, error) {
	return s.schoolAdminHasSavedADraftNotificationWithRequiredFields(ctx)
}

func (s *suite) opensEditorFullscreenDialogOfDraftNotification(ctx context.Context, arg1 string) (context.Context, error) {
	return ctx, nil
}

func (s *suite) seesDraftNotificationHasBeenSavedToScheduledNotification(ctx context.Context, arg1 string) (context.Context, error) {
	return s.storeNotificationSuccessfully(ctx)
}

func (s *suite) selectsDateTimeOfScheduleNotification(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.notification.ScheduledAt = timestamppb.New(time.Now().Truncate(time.Minute).Add(time.Minute))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectsNotificationStatus(ctx context.Context, arg1, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.notification.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED
	return StepStateToContext(ctx, stepState), nil
}
