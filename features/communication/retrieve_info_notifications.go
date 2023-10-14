package communication

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/communication/common"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
)

type RetrieveInfoNotificationsSuite struct {
	// *NotificationSuite
	*common.NotificationSuite
	receivedNotifications []*npb.RetrieveNotificationsResponse_NotificationInfo
	studentToken          string
	studentID             string
	notifications         []*cpb.Notification
}

func (c *SuiteConstructor) InitRetrieveInfoNotifications(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &RetrieveInfoNotificationsSuite{
		NotificationSuite: dep.notiCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^school admin sends "(\d+)" of notifications with "(\d+)" of important notifications to a student$`:                        s.schoolAdminSendsNotificationsWithImportantNotificationsToAStudent,
		`^student retrieves list of notifications with important only filter is "([^"]*)"$`:                                         s.studentRetrievesListOfNotificationsWithImportantOnlyFilter,
		`^returns correct list of notifications with counting is "(\d+)"$`:                                                          s.returnsCorrectListOfNotificationsWithCountingIs,
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" students$`:                                                                                 s.CreatesNumberOfStudents,
		`^student logins to Learner App$`:                                                                                           s.studentLoginsToLearnerApp,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *RetrieveInfoNotificationsSuite) studentLoginsToLearnerApp(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	student := commonState.Students[0]
	studentToken, err := s.GenerateExchangeTokenCtx(ctx, student.ID, "student")
	if err != nil {
		return ctx, fmt.Errorf("failed login learner app: %v", err)
	}
	s.studentToken = studentToken
	s.studentID = student.ID
	return ctx, nil
}

func (s *RetrieveInfoNotificationsSuite) schoolAdminSendsNotificationsWithImportantNotificationsToAStudent(ctx context.Context, numNoti int, numImportantNoti int) (context.Context, error) {
	countImportantNoti := 0
	for i := 0; i < numNoti; i++ {
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
			ReceiverIds:      []string{s.studentID},
		}
		if countImportantNoti < numImportantNoti {
			opts.IsImportant = true
			countImportantNoti++
		}
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
		notification.NotificationId = res.NotificationId

		_, err = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).SendNotification(
			ctx,
			&npb.SendNotificationRequest{
				NotificationId: res.NotificationId,
			},
		)
		if err != nil {
			return ctx, fmt.Errorf("failed SendNotification %v", err)
		}
		s.notifications = append(s.notifications, notification)
	}

	return ctx, nil
}

func (s *RetrieveInfoNotificationsSuite) studentRetrievesListOfNotificationsWithImportantOnlyFilter(ctx context.Context, isImportantFilterStr string) (context.Context, error) {
	isImportantFilter := false
	if isImportantFilterStr == "true" {
		isImportantFilter = true
	}

	notifications, err := s.GetNotificationByUser(s.studentToken, isImportantFilter)
	if err != nil {
		return ctx, fmt.Errorf("GetUserNotification %s", err)
	}

	s.receivedNotifications = notifications

	return ctx, nil
}

func (s *RetrieveInfoNotificationsSuite) returnsCorrectListOfNotificationsWithCountingIs(ctx context.Context, countingNoti int) (context.Context, error) {
	if countingNoti != len(s.receivedNotifications) {
		return ctx, fmt.Errorf("expected list of %d notificaions, got %d notificaions", countingNoti, len(s.receivedNotifications))
	}
	return ctx, nil
}
