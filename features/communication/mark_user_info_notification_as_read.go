package communication

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/communication/common"

	"github.com/cucumber/godog"
)

type MarkUserInfoNotificationAsReadSuite struct {
	*common.NotificationSuite
	studentToken string
}

func (c *SuiteConstructor) InitMarkUserInfoNotificationAsRead(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &MarkUserInfoNotificationAsReadSuite{
		NotificationSuite: dep.notiCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^mark the user notification as status "([^"]*)"$`:                                                                          s.markTheUserNotificationAsStatus,
		`^school admin creates "([^"]*)" students$`:                                                                                 s.CreatesNumberOfStudents,
		`^school admin sends notifications to student$`:                                                                             s.CurrentStaffSendNotification,
		`^school admin upsert notification to student$`:                                                                             s.schoolAdminUpsertNotification,
		`^student logins to Learner App$`:                                                                                           s.studentLoginsToLearnerApp,
		`^user set "([^"]*)" the notification$`:                                                                                     s.userSetTheNotification,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *MarkUserInfoNotificationAsReadSuite) studentLoginsToLearnerApp(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	student := commonState.Students[0]
	studentToken, err := s.GenerateExchangeTokenCtx(ctx, student.ID, "student")
	if err != nil {
		return ctx, fmt.Errorf("failed login learner app: %v", err)
	}
	s.studentToken = studentToken
	return ctx, nil
}

func (s *MarkUserInfoNotificationAsReadSuite) schoolAdminUpsertNotification(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	opts := &common.NotificationWithOpts{
		UserGroups:         "student",
		CourseFilter:       "random",
		GradeFilter:        "random",
		LocationFilter:     "none",
		ClassFilter:        "none",
		IndividualFilter:   "random",
		ScheduledStatus:    "random",
		Status:             "NOTIFICATION_STATUS_DRAFT",
		IsImportant:        false,
		GenericReceiverIds: []string{commonState.Students[0].ID},
	}
	return s.CurrentStaffUpsertNotificationWithOpts(ctx, opts)
}

func (s *MarkUserInfoNotificationAsReadSuite) userSetTheNotification(ctx context.Context, status string) (context.Context, error) {
	return s.UserSetStatusToNotification(ctx, s.studentToken, status)
}

func (s *MarkUserInfoNotificationAsReadSuite) markTheUserNotificationAsStatus(ctx context.Context, status string) (context.Context, error) {
	notifications, err := s.GetNotificationByUser(s.studentToken, false)
	if err != nil {
		return ctx, fmt.Errorf("failed get notifications: %v", err)
	}

	for _, noti := range notifications {
		if noti.UserNotification.Status.String() != status {
			return ctx, fmt.Errorf("expected status %s, got %s", status, noti.UserNotification.Status.String())
		}
	}
	return ctx, nil
}
