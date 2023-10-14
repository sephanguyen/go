package communication

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/godogutil"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
)

func (s *suite) ofLoginsLearnerApp(ctx context.Context, parentArg, studentArg string) (context.Context, error) {
	return s.loginsLearnerApp(ctx, parentArg)
}

func (s *suite) counteUserNotification(ctx context.Context, role string) (int, error) {
	role = strings.TrimSpace(role)
	token := s.getToken(role)
	req := &bpb.CountUserNotificationRequest{
		Status: cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW,
	}
	resp, err := bpb.NewNotificationReaderServiceClient(s.bobConn).CountUserNotification(contextWithToken(ctx, token), req)
	if err != nil {
		return 0, err
	}

	return int(resp.NumByStatus), nil
}

func (s *suite) receivesNotificationWithBadgeNumberOfNotificationBellDisplaysOnLearnerApp(ctx context.Context, userAccountArg, numArg string) (context.Context, error) {
	num, err := strconv.Atoi(numArg)
	if err != nil {
		return ctx, err
	}
	cur, err := s.counteUserNotification(ctx, s.currentUserAccount)
	if err != nil {
		return ctx, err
	}
	if num != cur {
		return ctx, fmt.Errorf("expect number of unread user notification %d but got %d", num, cur)
	}
	return ctx, nil
}

func (s *suite) scheduledNotificationHasSentTo(ctx context.Context, userAccountArg string) (context.Context, error) {
	st := StepStateFromContext(ctx)
	userAccount := ""
	if choices := parseOneOf(userAccountArg); choices != nil {
		userAccount = selectOneOf(userAccountArg)
	}
	st.currentUserAccount = userAccount
	userAccount = strings.TrimSpace(userAccount)
	st.notification.ReceiverIds = append(st.notification.ReceiverIds, st.profile.defaultStudent.id)
	switch userAccount {
	case student:
		st.notification.TargetGroup.UserGroupFilter = &cpb.NotificationTargetGroup_UserGroupFilter{
			UserGroups: []cpb.UserGroup{cpb.UserGroup_USER_GROUP_STUDENT},
		}
	case parent:
		st.notification.TargetGroup.UserGroupFilter = &cpb.NotificationTargetGroup_UserGroupFilter{
			UserGroups: []cpb.UserGroup{cpb.UserGroup_USER_GROUP_PARENT},
		}
	default:
		return ctx, fmt.Errorf("not support %s type", userAccount)
	}
	var err error
	ctx, err = godogutil.MultiErrChain(ctx,
		s.upsertNotification, st.notification,
		s.sendNotification, st.notification,
	)

	if err != nil {
		return ctx, err
	}

	return StepStateToContext(ctx, st), nil
}

func (s *suite) withTheScheduledNotification(ctx context.Context, userAccountArg, actionArg string) (context.Context, error) {
	st := StepStateFromContext(ctx)
	if actionArg == "read" {
		err := s.userReadNotification(ctx, st.currentUserAccount, st.notification.NotificationId)
		if err != nil {
			return ctx, err
		}
	}
	// else unread do nothing
	return ctx, nil
}
