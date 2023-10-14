package communication

import (
	"context"
	"strconv"
	"strings"

	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"
)

func (s *suite) receivesNotification(ctx context.Context, arg1 string) (context.Context, error) {
	return s.receivesTheNotificationInTheirDevice(ctx, arg1)
}

func (s *suite) doesNotReceiveAnyNotification(ctx context.Context, arg1 string) (context.Context, error) {
	return ctx, nil
}

func (s *suite) hasNotReadNotification(ctx context.Context, role string) (context.Context, error) {
	// if we do not read notification so we do nothing
	return ctx, nil
}

func (s *suite) userReadNotification(ctx context.Context, role string, notificationID string) error {
	role = strings.TrimSpace(role)
	token := s.getToken(role)
	req := &bpb.SetUserNotificationStatusRequest{
		NotificationIds: []string{notificationID},
		Status:          cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_READ,
	}
	_, err := bpb.NewNotificationModifierServiceClient(s.bobConn).SetUserNotificationStatus(contextWithToken(ctx, token), req)
	if err != nil {
		return err
	}

	return nil
}

func (s *suite) hasReadTheNotification(ctx context.Context, userArgs string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	usersRaw := ""
	if choices := parseOneOf(userArgs); choices != nil {
		usersRaw = selectOneOf(userArgs)
	} else {
		usersRaw = userArgs
	}
	// userRaw example: student & parent
	us := strings.Split(usersRaw, "&")

	for _, role := range us {
		err := s.userReadNotification(ctx, role, stepState.notification.NotificationId)
		if err != nil {
			return ctx, err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) schoolAdminHasCreatedAStudentWith(ctx context.Context, _, numParentsArg, numCoursesArg string) (context.Context, error) {
	var err error
	numCourses, _ := strconv.Atoi(strings.Split(numCoursesArg, " ")[0])

	ctx, err = s.hasCreatedCourse(ctx, schoolAdmin, numCourses)
	if err != nil {
		return ctx, err
	}

	return s.schoolAdminHasCreatedStudentWithParentInfo(ctx)
}

func (s *suite) schoolAdminHasCreatedNotificationAndSentForCreatedStudentAndParent(ctx context.Context) (context.Context, error) {
	return s.sendsNotificationWithRequiredFieldsToStudentAndParent(ctx, schoolAdmin)
}

func (s *suite) schoolAdminIsAtNotificationPageOnCMS(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (s *suite) schoolAdminResendsNotificationForUnreadRecipients(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	token := s.getToken(schoolAdmin)
	req := &ypb.NotifyUnreadUserRequest{
		NotificationId: stepState.notification.NotificationId,
	}
	_, err := ypb.NewNotificationModifierServiceClient(s.yasuoConn).NotifyUnreadUser(contextWithToken(ctx, token), req)
	if err != nil {
		return ctx, err
	}
	return StepStateToContext(ctx, stepState), nil
}
