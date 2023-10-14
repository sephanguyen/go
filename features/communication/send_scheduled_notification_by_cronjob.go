package communication

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SendScheduledNotificationByCronjobSuite struct {
	*common.NotificationSuite
	NotificationNeedToSent *cpb.Notification
	minWaiting             int
	truncateMin            time.Duration
}

func (c *SuiteConstructor) InitSendScheduledNotificationByCronjob(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &SendScheduledNotificationByCronjobSuite{
		NotificationSuite: dep.notiCommonSuite,
		truncateMin:       time.Minute,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" students with "([^"]*)" parents info for each student$`:                                    s.CreatesNumberOfStudentsWithParentsInfo,
		`^school admin creates "([^"]*)" courses$`:                                                                                  s.CreatesNumberOfCourses,
		`^school admin add packages data of those courses for each student$`:                                                        s.AddPackagesDataOfThoseCoursesForEachStudent,
		`^notificationmgmt services must send notification to user$`:                                                                s.NotificationMgmtMustSendNotificationToUser,
		`^notificationmgmt services must store the notification with correctly info$`:                                               s.NotificationMgmtMustStoreTheNotification,
		`^current staff schedules a notification to be sent after (\d+) minutes$`:                                                   s.userScheduledSomeNotificationToBeSentAfterSomeMinutes,
		`^waiting for scheduled notification to be sent$`:                                                                           s.waitToNotificationSend,
		`^sent time is valid with scheduled time$`:                                                                                  s.checkValidSentTimeAndScheduledTime,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *SendScheduledNotificationByCronjobSuite) userScheduledSomeNotificationToBeSentAfterSomeMinutes(ctx context.Context, minutes int) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	s.minWaiting = minutes + 1
	opts := &common.NotificationWithOpts{
		UserGroups:       "student, parent",
		CourseFilter:     "all",
		GradeFilter:      "all",
		LocationFilter:   "all",
		ClassFilter:      "random",
		IndividualFilter: "random",
		ScheduledStatus:  "random",
		Status:           "NOTIFICATION_STATUS_SCHEDULED",
		IsImportant:      false,
	}
	var err error
	ctx, s.NotificationNeedToSent, err = s.GetNotificationWithOptions(ctx, opts)
	commonState.Notification = s.NotificationNeedToSent
	if err != nil {
		return ctx, err
	}

	// only set minutes to the same with FE -> need to truncate
	scheduledTime := time.Now().Add(time.Minute * time.Duration(minutes)).Truncate(s.truncateMin)
	s.NotificationNeedToSent.ScheduledAt = timestamppb.New(scheduledTime)

	commonState.Request = &npb.UpsertNotificationRequest{
		Notification: s.NotificationNeedToSent,
	}
	commonState.Response, commonState.ResponseErr = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(s.ContextWithToken(ctx, commonState.CurrentStaff.Token), commonState.Request.(*npb.UpsertNotificationRequest))
	if commonState.ResponseErr == nil {
		resp := commonState.Response.(*npb.UpsertNotificationResponse)
		s.NotificationNeedToSent.NotificationId = resp.NotificationId
		commonState.Notification.NotificationId = resp.NotificationId
	} else {
		return ctx, fmt.Errorf("s.UpsertNotification: %v", commonState.ResponseErr)
	}

	ctx, err = s.CheckReturnStatusCode(ctx, "OK")
	if err != nil {
		return ctx, err
	}
	resp := commonState.Response.(*npb.UpsertNotificationResponse)
	s.NotificationNeedToSent.NotificationId = resp.NotificationId
	commonState.Notification.NotificationId = resp.NotificationId
	return common.StepStateToContext(ctx, commonState), nil
}

func (s *SendScheduledNotificationByCronjobSuite) waitToNotificationSend(ctx context.Context) (context.Context, error) {
	fmt.Printf("\nWaiting for scheduled notification to be sent by cronjob...\n")
	waitTime := time.Minute * time.Duration(s.minWaiting)
	time.Sleep(waitTime)
	return ctx, nil
}

func (s *SendScheduledNotificationByCronjobSuite) checkValidSentTimeAndScheduledTime(ctx context.Context) (context.Context, error) {
	infoNotification := &entities.InfoNotification{}
	fields := database.GetFieldNames(infoNotification)
	queryGetNotication := fmt.Sprintf(`SELECT %s FROM %s WHERE notification_id = $1 AND deleted_at IS NULL;`, strings.Join(fields, ","), infoNotification.TableName())

	err := database.Select(ctx, s.BobDBConn, queryGetNotication, database.Text(s.NotificationNeedToSent.NotificationId)).ScanOne(infoNotification)
	if err != nil {
		return ctx, err
	}

	sentTime := infoNotification.SentAt.Time.Truncate(s.truncateMin).String()
	scheduledTime := infoNotification.ScheduledAt.Time.String()

	if sentTime != scheduledTime {
		return ctx, fmt.Errorf("sent time is not valid: expect [%s], got [%s]", scheduledTime, sentTime)
	}

	return ctx, nil
}
