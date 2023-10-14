package communication

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/communication/common"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
)

type SetStatusForUserNotificationsSuite struct {
	*common.NotificationSuite
}

func (c *SuiteConstructor) InitSetStatusForUserNotifications(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &SetStatusForUserNotificationsSuite{
		NotificationSuite: dep.notiCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^school admin sends some notificationss to a student$`: s.schoolAdminSendsSomeNotificationssToAStudent,
		`^user set "([^"]*)" the notification$`:                 s.userSetStatusTheNotification,
		`^mark the user notification as status "([^"]*)"$`:      s.markTheUserNotificationAsStatus,
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" students$`: s.CreatesNumberOfStudents,
		`^student "([^"]*)" logins to Learner App$`: s.StudentLoginsToLearnerApp,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *SetStatusForUserNotificationsSuite) schoolAdminSendsSomeNotificationssToAStudent(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	randNum := common.RandRangeIn(1, 5)
	opts := &common.NotificationWithOpts{
		UserGroups:       "student",
		CourseFilter:     "random",
		GradeFilter:      "random",
		LocationFilter:   "none",
		ClassFilter:      "none",
		IndividualFilter: "none",
		ScheduledStatus:  "none",
		Status:           "NOTIFICATION_STATUS_DRAFT",
		IsImportant:      false,
		ReceiverIds:      []string{commonState.Students[0].ID},
	}
	for i := 0; i < randNum; i++ {
		_, notification, err := s.GetNotificationWithOptions(ctx, opts)
		if err != nil {
			return ctx, fmt.Errorf("failed GetNotificationWithOptions %v", err)
		}

		res, err := npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(
			ctx,
			&npb.UpsertNotificationRequest{
				Notification: notification,
			},
		)
		if err != nil {
			return ctx, fmt.Errorf("failed UpsertNotification %v", err)
		}

		_, err = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).SendNotification(
			ctx,
			&npb.SendNotificationRequest{
				NotificationId: res.NotificationId,
			},
		)
		if err != nil {
			return ctx, fmt.Errorf("failed SendNotification %v", err)
		}
	}

	return ctx, nil
}

func (s *SetStatusForUserNotificationsSuite) userSetStatusTheNotification(ctx context.Context, status string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	notificationsList, err := s.GetNotificationByUser(commonState.Students[0].Token, false)
	if err != nil {
		return ctx, fmt.Errorf("GetUserNotification: %v", err)
	}

	userNotificationIds := make([]string, 0)
	for _, notification := range notificationsList {
		userNotificationIds = append(userNotificationIds, notification.UserNotification.UserNotificationId)
	}

	req := &npb.SetStatusForUserNotificationsRequest{
		UserNotificationIds: userNotificationIds,
		Status:              cpb.UserNotificationStatus(cpb.UserNotificationStatus_value[status]),
	}

	ctx, cancel := common.ContextWithTokenAndTimeOut(ctx, commonState.Students[0].Token)
	defer cancel()
	_, err = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).SetStatusForUserNotifications(ctx, req)

	if err != nil {
		return ctx, fmt.Errorf("SetStatusForUserNotifications: %v", err)
	}

	return ctx, nil
}

func (s *SetStatusForUserNotificationsSuite) markTheUserNotificationAsStatus(ctx context.Context, status string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	notifications, err := s.GetNotificationByUser(commonState.Students[0].Token, false)
	if err != nil {
		return ctx, fmt.Errorf("GetUserNotification %s", err)
	}

	for _, resNoti := range notifications {
		if resNoti.UserNotification.Status.String() != status {
			return ctx, fmt.Errorf("expected status %s, got %s", status, resNoti.UserNotification.Status.String())
		}
	}

	return ctx, nil
}
