package communication

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
)

type CronJobResendScheduledNotificationSuite struct {
	*common.NotificationSuite
	scheduledAt  time.Time
	studentToken string
}

func (c *SuiteConstructor) InitCronJobResendScheduledNotificationAfterSendFailed(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &CronJobResendScheduledNotificationSuite{
		NotificationSuite: dep.notiCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" students$`:                                s.CreatesNumberOfStudents,
		`^student logins to Learner App$`:                                          s.studentLoginsToLearnerApp,
		`^admin check that "(\d+)" notification are sent within a minute$`:         s.adminCheckThatAllNotificationAreSentWithinMinute,
		`^admin create (\d+) group of scheduled notification with different time$`: s.adminCreateGroupOfScheduledNotificationWithDifferentTime,
		`^admin create "(\d+)" scheduled notification to student$`:                 s.adminCreateScheduledNotificationToStudent,
		`^group (\d+) are also sent$`:                                              s.groupAreAlsoSent,
		`^group (\d+) was sent failed$`:                                            s.groupWasSentFailed,
		`^waiting for all notification are sent$`:                                  s.waitingForAllNotificationAreSent,
		`^waiting to group (\d+) are sent$`:                                        s.waitingToGroupAreSent,
	}
	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *CronJobResendScheduledNotificationSuite) studentLoginsToLearnerApp(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	student := commonState.Students[0]
	studentToken, err := s.GenerateExchangeTokenCtx(ctx, student.ID, "student")
	if err != nil {
		return ctx, fmt.Errorf("failed login learner app: %v", err)
	}
	s.studentToken = studentToken
	return ctx, nil
}

func (s *CronJobResendScheduledNotificationSuite) adminCheckThatAllNotificationAreSentWithinMinute(ctx context.Context, numOfNotification int) (context.Context, error) {
	notifications, err := s.GetNotificationByUser(s.studentToken, false)
	if err != nil {
		return ctx, err
	}

	if len(notifications) < numOfNotification {
		err = fmt.Errorf("has a notification still not send. want %v has %v", numOfNotification, len(notifications))
		return ctx, err
	}
	return ctx, nil
}

func (s *CronJobResendScheduledNotificationSuite) adminCreateGroupOfScheduledNotificationWithDifferentTime(ctx context.Context, numOfTime int) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	if commonState.Organization == nil || len(commonState.Organization.Staffs) == 0 {
		return ctx, fmt.Errorf("no school admin found")
	}

	s.scheduledAt = time.Now().Truncate(time.Minute).Add(time.Minute)
	for i := 0; i < numOfTime; i++ {
		scheduledStatus := fmt.Sprintf("%d min", 1)
		s.scheduledAt.Add(time.Duration(i) * time.Minute)
		opts := &common.NotificationWithOpts{
			UserGroups:       "student",
			CourseFilter:     "all",
			GradeFilter:      "all",
			LocationFilter:   "none",
			ClassFilter:      "none",
			IndividualFilter: "none",
			ScheduledStatus:  scheduledStatus,
			Status:           "NOTIFICATION_STATUS_SCHEDULED",
			IsImportant:      false,
			ReceiverIds:      []string{commonState.Students[0].ID},
		}
		_, notification, err := s.GetNotificationWithOptions(ctx, opts)
		if err != nil {
			return ctx, fmt.Errorf("failed GetNotificationWithOptions %v", err)
		}
		res, err := npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(
			common.ContextWithToken(ctx, commonState.CurrentStaff.Token),
			&npb.UpsertNotificationRequest{
				Notification: notification,
			},
		)
		if err != nil {
			return ctx, fmt.Errorf("failed UpsertNotification %v", err)
		} else if i == 0 {
			commonState.Notification = notification
			commonState.Notification.NotificationId = res.NotificationId
		}
	}
	return common.StepStateToContext(ctx, commonState), nil
}

func (s *CronJobResendScheduledNotificationSuite) adminCreateScheduledNotificationToStudent(ctx context.Context, numOfNotification int) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	if commonState.Organization == nil || len(commonState.Organization.Staffs) == 0 {
		return ctx, fmt.Errorf("no school admin found")
	}

	s.scheduledAt = time.Now().Truncate(time.Minute).Add(2 * time.Minute)
	for i := 0; i < numOfNotification; i++ {
		opts := &common.NotificationWithOpts{
			UserGroups:       "student",
			CourseFilter:     "all",
			GradeFilter:      "all",
			LocationFilter:   "none",
			ClassFilter:      "none",
			IndividualFilter: "none",
			ScheduledStatus:  "2 min",
			Status:           "NOTIFICATION_STATUS_SCHEDULED",
			IsImportant:      false,
			ReceiverIds:      []string{commonState.Students[0].ID},
		}
		_, notification, err := s.GetNotificationWithOptions(ctx, opts)
		if err != nil {
			return ctx, fmt.Errorf("failed GetNotificationWithOptions %v", err)
		}
		_, err = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(
			common.ContextWithToken(ctx, commonState.CurrentStaff.Token),
			&npb.UpsertNotificationRequest{
				Notification: notification,
			},
		)
		if err != nil {
			return ctx, fmt.Errorf("failed UpsertNotification %v", err)
		}
	}
	return common.StepStateToContext(ctx, commonState), nil
}

func (s *CronJobResendScheduledNotificationSuite) groupAreAlsoSent(ctx context.Context, group int) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	notifications, err := s.GetNotificationByUser(s.studentToken, false)
	if err != nil {
		return ctx, err
	}

	if len(notifications) == 0 {
		return ctx, fmt.Errorf("no notification are send")
	}

	found := false
	for _, notify := range notifications {
		if notify.UserNotification.NotificationId == commonState.Notification.NotificationId {
			found = true
		}
	}
	if !found {
		return ctx, fmt.Errorf("first notification are not send")
	}

	return ctx, nil
}

func (s *CronJobResendScheduledNotificationSuite) groupWasSentFailed(ctx context.Context, group int) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	if s.scheduledAt.After(time.Now()) {
		// waiting for first scheduled sent
		waitTime := s.scheduledAt.Unix() - time.Now().Unix()
		time.Sleep(time.Duration(waitTime+10) * time.Second)

		// clear database and update status of it back to schedule
		db := s.BobDBConn
		fakeContext := context.Background()
		interceptors.ContextWithJWTClaims(fakeContext, &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{
				ResourcePath: fmt.Sprintf("%v", commonState.Organization.ID),
				UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
			},
		})

		sql := "UPDATE info_notifications SET STATUS =  $1, sent_at = null WHERE notification_id = $2"
		_, err := db.Exec(fakeContext, sql, cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED, commonState.Notification.NotificationId)
		if err != nil {
			return ctx, fmt.Errorf("cannot update notification back to status scheduled")
		}

		sql2 := "DELETE FROM users_info_notifications WHERE notification_id = $1"
		_, err2 := db.Exec(fakeContext, sql2, commonState.Notification.NotificationId)
		if err2 != nil {
			return ctx, fmt.Errorf("cannot remove notifycation user")
		}
	}
	return ctx, nil
}

func (s *CronJobResendScheduledNotificationSuite) waitingForAllNotificationAreSent(ctx context.Context) (context.Context, error) {
	if s.scheduledAt.After(time.Now()) {
		waitTime := s.scheduledAt.Unix() - time.Now().Unix()
		fmt.Printf("\nWaiting for scheduled notification to be sent...\n")
		time.Sleep(time.Duration(waitTime+90) * time.Second)
	}
	return ctx, nil
}

func (s *CronJobResendScheduledNotificationSuite) waitingToGroupAreSent(ctx context.Context, group int) (context.Context, error) {
	if s.scheduledAt.Add(time.Minute).After(time.Now()) {
		waitTime := s.scheduledAt.Add(time.Minute).Unix() - time.Now().Unix()
		fmt.Printf("\nWaiting for scheduled notification to be sent...\n")
		time.Sleep(time.Duration(waitTime+30) * time.Second)
	}
	return ctx, nil
}
