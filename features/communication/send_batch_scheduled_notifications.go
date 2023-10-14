package communication

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SendBatchScheduledNotificationsSuite struct {
	*common.NotificationSuite
	NotificationNeedToSent *cpb.Notification
}

func (c *SuiteConstructor) InitSendBatchScheduledNotifications(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &SendBatchScheduledNotificationsSuite{
		NotificationSuite: dep.notiCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" students with "([^"]*)" parents info for each student$`:                                    s.CreatesNumberOfStudentsWithParentsInfo,
		`^school admin creates "([^"]*)" courses$`:                                                                                  s.CreatesNumberOfCourses,
		`^school admin add packages data of those courses for each student$`:                                                        s.AddPackagesDataOfThoseCoursesForEachStudent,
		`^call send scheduled notification within (\d+) hour$`:                                                                      s.callSendScheduledNotificationsWithinHour,
		`^notification scheduled to be sent after (\d+) hour are not sent$`:                                                         s.notificationsScheduledToBeSentAfterHourAreNotSent,
		`^current staff schedules a notification to be sent after (\d+) hour$`:                                                      s.userScheduledSomeNotificationToBeSentAfterHour,
		`^notification scheduled to be sent within (\d+) hour are sent$`:                                                            s.notificationsScheduledToBeSentWithinHourAreSent,
		`^current staff schedules a notification to be sent within (\d+) hour$`:                                                     s.userScheduledSomeNotificationToBeSentWithinHour,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *SendBatchScheduledNotificationsSuite) userScheduledSomeNotificationToBeSentWithinHour(ctx context.Context, numHour int) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	opts := &common.NotificationWithOpts{
		UserGroups:       "student, parent",
		CourseFilter:     "random",
		GradeFilter:      "random",
		LocationFilter:   "all",
		ClassFilter:      "none",
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

	// schedule notification to be sent in with in [num] hour
	// random from minute [1, numHour*60]
	max := int((time.Duration(numHour) * time.Hour).Minutes())
	duration := time.Duration(rand.Intn(max-1) + 1)
	s.NotificationNeedToSent.ScheduledAt = timestamppb.New(time.Now().Add(duration * time.Minute))

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
	return s.checkCreatedNotification(common.StepStateToContext(ctx, commonState), resp.NotificationId)
}

func (s *SendBatchScheduledNotificationsSuite) checkCreatedNotification(ctx context.Context, notificationID string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	infoNotification := &entities.InfoNotification{}
	fields := database.GetFieldNames(infoNotification)
	queryGetNotication := fmt.Sprintf(`SELECT %s FROM %s WHERE notification_id = $1 AND deleted_at IS NULL;`, strings.Join(fields, ","), infoNotification.TableName())

	err := database.Select(ctx, s.BobDBConn, queryGetNotication, database.Text(notificationID)).ScanOne(infoNotification)
	if err != nil {
		return ctx, err
	}

	req, ok := commonState.Request.(*npb.UpsertNotificationRequest)
	if !ok {
		return ctx, fmt.Errorf("expect npb.UpsertNotificationRequest but got %v", commonState.Request)
	}

	ctx, err = s.CheckInfoNotificationResponse(ctx, req.Notification, infoNotification)
	if err != nil {
		return ctx, err
	}

	infoNotificationMsg := &entities.InfoNotificationMsg{}
	fields = database.GetFieldNames(infoNotificationMsg)
	queryGetInfoNoticationMsg := fmt.Sprintf(`SELECT %s FROM %s WHERE notification_msg_id = $1 AND deleted_at IS NULL;`, strings.Join(fields, ","), infoNotificationMsg.TableName())

	err = database.Select(ctx, s.BobDBConn, queryGetInfoNoticationMsg, infoNotification.NotificationMsgID).ScanOne(infoNotificationMsg)
	if err != nil {
		return ctx, err
	}

	msgEnt, err := mappers.PbToInfoNotificationMsgEnt(req.Notification.Message)
	if err != nil {
		return common.StepStateToContext(ctx, commonState), fmt.Errorf("mappers.PbToInfoNotificationMsgEnt: %v", err)
	}

	err = s.CheckInfoNotificationMsgResponse(msgEnt, infoNotificationMsg)
	if err != nil {
		return ctx, err
	}

	return common.StepStateToContext(ctx, commonState), nil
}

func (s *SendBatchScheduledNotificationsSuite) userScheduledSomeNotificationToBeSentAfterHour(ctx context.Context, numHour int) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	opts := &common.NotificationWithOpts{
		UserGroups:       "student, parent",
		CourseFilter:     "random",
		GradeFilter:      "random",
		LocationFilter:   "none",
		ClassFilter:      "none",
		IndividualFilter: "random",
		ScheduledStatus:  "random",
		Status:           "NOTIFICATION_STATUS_SCHEDULED",
		IsImportant:      false,
	}
	var err error
	ctx, commonState.Notification, err = s.GetNotificationWithOptions(ctx, opts)
	if err != nil {
		return ctx, err
	}

	// schedule notification to be sent after [num] hour
	h := time.Now().Add(time.Duration(numHour) * time.Hour)
	commonState.Notification.ScheduledAt = timestamppb.New(h.Add(10 * time.Minute))

	commonState.Request = &npb.UpsertNotificationRequest{
		Notification: commonState.Notification,
	}
	commonState.Response, commonState.ResponseErr = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(s.ContextWithToken(ctx, commonState.CurrentStaff.Token), commonState.Request.(*npb.UpsertNotificationRequest))
	if commonState.ResponseErr == nil {
		resp := commonState.Response.(*npb.UpsertNotificationResponse)
		commonState.Notification.NotificationId = resp.NotificationId
	} else {
		return ctx, fmt.Errorf("s.UpsertNotification: %v", commonState.ResponseErr)
	}

	ctx, err = s.CheckReturnStatusCode(ctx, "OK")
	if err != nil {
		return ctx, err
	}
	resp := commonState.Response.(*npb.UpsertNotificationResponse)
	commonState.Notification.NotificationId = resp.NotificationId

	ctx, err = s.checkCreatedNotification(ctx, resp.NotificationId)
	return common.StepStateToContext(ctx, commonState), err
}

func (s *SendBatchScheduledNotificationsSuite) notificationsScheduledToBeSentAfterHourAreNotSent(ctx context.Context, arg1 int) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	id := commonState.Notification.NotificationId
	noti, err := s.queryNotification(ctx, id)
	if err != nil {
		return ctx, err
	}
	if noti.Status.String == cpb.NotificationStatus_NOTIFICATION_STATUS_SENT.String() {
		return ctx, fmt.Errorf("expect notification id %s to be sent, got: %s", id, noti.Status.String)
	}
	return common.StepStateToContext(ctx, commonState), nil
}

func (s *SendBatchScheduledNotificationsSuite) notificationsScheduledToBeSentWithinHourAreSent(ctx context.Context, arg1 int) (context.Context, error) {
	id := s.NotificationNeedToSent.NotificationId
	noti, err := s.queryNotification(ctx, id)
	if err != nil {
		return ctx, err
	}
	if noti.Status.String != cpb.NotificationStatus_NOTIFICATION_STATUS_SENT.String() {
		return ctx, fmt.Errorf("expect notification id %s to be sent, got: %s", id, noti.Status.String)
	}
	return ctx, nil
}

func (s *SendBatchScheduledNotificationsSuite) callSendScheduledNotificationsWithinHour(ctx context.Context, arg1 int) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	commonState.Request = &npb.SendScheduledNotificationRequest{
		From:                   timestamppb.New(time.Now().Add(-1 * time.Hour)),
		To:                     timestamppb.New(time.Now().Add(time.Hour)),
		IsRunningForAllTenants: true,
	}
	var err error
	if err != nil {
		return ctx, err
	}
	commonState.Response, commonState.ResponseErr = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).SendScheduledNotification(s.ContextWithToken(ctx, commonState.CurrentStaff.Token), commonState.Request.(*npb.SendScheduledNotificationRequest))
	if commonState.ResponseErr != nil {
		return ctx, commonState.ResponseErr
	}
	return common.StepStateToContext(ctx, commonState), nil
}

func (s *SendBatchScheduledNotificationsSuite) queryNotification(ctx context.Context, notificationID string) (*entities.InfoNotification, error) {
	e := &entities.InfoNotification{}
	query := fmt.Sprintf(`SELECT %s FROM info_notifications WHERE notification_id = $1 AND deleted_at IS NULL;`, strings.Join(database.GetFieldNames(e), ","))

	err := s.BobDBConn.QueryRow(ctx, query, database.Text(notificationID)).Scan(database.GetScanFields(e, database.GetFieldNames(e))...)
	if err != nil {
		return nil, fmt.Errorf("cannot find notification id %s error: %w", notificationID, err)
	}
	return e, nil
}
