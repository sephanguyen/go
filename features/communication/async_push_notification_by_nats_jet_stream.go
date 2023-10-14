package communication

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

type PushByNatsSuite struct {
	*common.NotificationSuite
	receivedNotify *npb.RetrieveNotificationsResponse_NotificationInfo
	studentToken   string
	parentToken    string
}

type CustomData struct {
	CustomDataType string `json:"custom_data_type"`
}

func (c *SuiteConstructor) InitAsyncPushNotificationByNatsJetStream(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &PushByNatsSuite{
		NotificationSuite: dep.notiCommonSuite,
	}
	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		// `^a new "([^"]*)" of Manabie organization with default location logged in Back Office$`:  s.StaffGrantedRoleOfManabieOrganizationWithDefaultLocationLoggedInBackOffice,
		`^client push "([^"]*)" notification to "([^"]*)" with config permanent save is "([^"]*)"$`:                  s.clientPushNotificationToWithConfigPermanentSaveIs,
		`^client push "([^"]*)" notification using generic id of "([^"]*)" with config permanent save is "([^"]*)"$`: s.clientPushNotificationUsingGenericIDOfWithConfigPermanentSaveIs,
		`^"([^"]*)" must be receive notification$`:                                                                   s.mustBeReceiveNotification,
		`^notification bell display "([^"]*)" new notification of "([^"]*)"$`:                                        s.notificationBellDisplayNewNotificationOf,
		`^notification list of "([^"]*)" must be show right data$`:                                                   s.notificationListOfMustBeShowRightData,
		`^school admin add packages data of those courses for each student$`:                                         s.AddPackagesDataOfThoseCoursesForEachStudent,
		`^school admin creates "([^"]*)" course$`:                                                                    s.CreatesNumberOfCourses,
		`^school admin creates "([^"]*)" students with "([^"]*)" parents$`:                                           s.CreatesNumberOfStudentsWithParentsInfo,
		`^student and parent login to learner app$`:                                                                  s.studentAndParentLoginToLearnerApp,
		`^wait to "([^"]*)" notification send$`:                                                                      s.waitToNotificationSend,
		`^update user device token to an "([^"]*)" device token$`:                                                    s.UpdateDeviceTokenForLeanerUser,
		`^wait for FCM is sent to target user$`:                                                                      s.WaitingForFCMIsSent,
	}
	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *PushByNatsSuite) studentAndParentLoginToLearnerApp(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	student := commonState.Students[0]
	studentToken, err := s.GenerateExchangeTokenCtx(ctx, student.ID, constant.UserGroupStudent)
	if err != nil {
		return ctx, fmt.Errorf("failed GenerateExchangeTokenCtx student: %v", err)
	}
	s.studentToken = studentToken

	parentToken, err := s.GenerateExchangeTokenCtx(ctx, student.Parents[0].ID, constant.UserGroupParent)
	if err != nil {
		return ctx, fmt.Errorf("failed GenerateExchangeTokenCtx parent: %v", err)
	}
	s.parentToken = parentToken

	return ctx, nil
}

func (s *PushByNatsSuite) notificationListOfMustBeShowRightData(ctx context.Context, target string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	tokens := make([]string, 0)
	switch target {
	case "student":
		tokens = append(tokens, s.studentToken)
	case "parent":
		tokens = append(tokens, s.parentToken)
	default:
		tokens = append(tokens, s.studentToken, s.parentToken)
	}

	for _, token := range tokens {
		notifications, err := s.GetNotificationByUser(token, false)
		if err != nil {
			return ctx, err
		}

		if !commonState.NatsNotification.NotificationConfig.PermanentStorage {
			if len(notifications) != 0 {
				return ctx, fmt.Errorf("target will not receive notification with permanentStorage is false")
			}
		} else {
			if len(notifications) == 0 {
				return ctx, fmt.Errorf("target was not received notification")
			}
		}

		// compare by title
		for _, notify := range notifications {
			if notify.Title != commonState.NatsNotification.NotificationConfig.Notification.Title {
				return ctx, fmt.Errorf("target was not received notification")
			}
			s.receivedNotify = notify
		}
	}

	natsNotify := commonState.NatsNotification
	if !natsNotify.NotificationConfig.PermanentStorage {
		return common.StepStateToContext(ctx, commonState), nil
	}

	if s.receivedNotify.Title != natsNotify.NotificationConfig.Notification.Title {
		return ctx, fmt.Errorf("notify title is not match, expected: '%s', receive: '%s'", natsNotify.NotificationConfig.Notification.Title, s.receivedNotify.Title)
	}

	if s.receivedNotify.Description != natsNotify.NotificationConfig.Notification.Message {
		return ctx, fmt.Errorf("notify message is not match, expected: '%s', receive: '%s'", natsNotify.NotificationConfig.Notification.Message, s.receivedNotify.Description)
	}

	dataNotification, err := json.Marshal(natsNotify.NotificationConfig.Data)
	if err != nil {
		return ctx, fmt.Errorf("can not parse data notification")
	}
	var dataBeforeSend CustomData
	var dataAfterSend CustomData

	err = json.Unmarshal(dataNotification, &dataBeforeSend)

	if err != nil {
		return ctx, fmt.Errorf("cannot parse data to custom struct")
	}

	err = json.Unmarshal([]byte(s.receivedNotify.UserNotification.Data), &dataAfterSend)

	if err != nil {
		return ctx, fmt.Errorf("cannot parse data to custom struct")
	}

	if dataAfterSend.CustomDataType != dataBeforeSend.CustomDataType {
		return ctx, fmt.Errorf("notify data is not match")
	}

	if s.receivedNotify.UserNotification.Type.String() != cpb.NotificationType_NOTIFICATION_TYPE_NATS_ASYNC.String() {
		return ctx, fmt.Errorf("notify type is not match")
	}

	return common.StepStateToContext(ctx, commonState), nil
}

func (s *PushByNatsSuite) clientPushNotificationUsingGenericIDOfWithConfigPermanentSaveIs(ctx context.Context, notificationType, target, isSave string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	commonState.ClientID = "bdd_testing_client_id"

	tracingID := uuid.New().String()
	notification := s.MakeSampleNatsNotification(commonState.ClientID, tracingID, commonState.Organization.ID, notificationType)

	ctxInjectResourcePathAndUserID := contextWithResourcePathAndUserID(ctx, commonState.CurrentResourcePath, commonState.CurrentUserID)
	notification.Target = &ypb.NatsNotificationTarget{}
	switch target {
	case "student":
		notification.Target.GenericUserIds = []string{commonState.Students[0].ID}
	case "parent":
		notification.Target.GenericUserIds = []string{commonState.Students[0].Parents[0].ID}
	}

	commonState.NatsNotification = notification

	data, _ := proto.Marshal(notification)
	err := s.PublishToNats(ctxInjectResourcePathAndUserID, "Notification.Created", data)
	if err != nil {
		return ctxInjectResourcePathAndUserID, err
	}

	return common.StepStateToContext(ctxInjectResourcePathAndUserID, commonState), nil
}

func (s *PushByNatsSuite) clientPushNotificationToWithConfigPermanentSaveIs(ctx context.Context, notificationType, target, isSave string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	commonState.ClientID = "bdd_testing_client_id"
	tracingID := uuid.New().String()

	natsNotification := s.MakeSampleNatsNotification(commonState.ClientID, tracingID, commonState.Organization.ID, notificationType)

	ctxInjectResourcePathAndUserID := contextWithResourcePathAndUserID(ctx, commonState.CurrentResourcePath, commonState.CurrentUserID)

	if notificationType == consts.NotificationTypeScheduled {
		natsNotification.SendTime.Time = time.Now().Add(1 * time.Minute).Format(LauoutTimeFormat)
	}

	if isSave == "false" {
		natsNotification.NotificationConfig.PermanentStorage = false
	} else {
		natsNotification.NotificationConfig.PermanentStorage = true
	}

	natsNotification.Target.ReceivedUserIds = []string{commonState.Students[0].ID}

	natsNotification.TargetGroup = &ypb.NatsNotificationTargetGroup{
		UserGroupFilter: &ypb.NatsNotificationTargetGroup_UserGroupFilter{
			UserGroups: []string{},
		},
	}

	switch target {
	case "student":
		natsNotification.TargetGroup.UserGroupFilter = &ypb.NatsNotificationTargetGroup_UserGroupFilter{
			UserGroups: []string{consts.TargetUserGroupStudent},
		}
	case "parent":
		natsNotification.TargetGroup.UserGroupFilter = &ypb.NatsNotificationTargetGroup_UserGroupFilter{
			UserGroups: []string{consts.TargetUserGroupParent},
		}
	default:
		natsNotification.TargetGroup.UserGroupFilter = &ypb.NatsNotificationTargetGroup_UserGroupFilter{
			UserGroups: []string{consts.TargetUserGroupStudent, consts.TargetUserGroupParent},
		}
	}

	commonState.NatsNotification = natsNotification

	data, _ := proto.Marshal(natsNotification)
	err := s.PublishToNats(ctxInjectResourcePathAndUserID, "Notification.Created", data)
	if err != nil {
		return ctxInjectResourcePathAndUserID, err
	}

	return common.StepStateToContext(ctx, commonState), nil
}

func (s *PushByNatsSuite) mustBeReceiveNotification(ctx context.Context, target string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	userIds := make([]string, 0)

	switch target {
	case "student":
		userIds = append(userIds, commonState.Students[0].User.ID)
	case "parent":
		userIds = append(userIds, commonState.Students[0].Parents[0].ID)
	default:
		userIds = append(userIds, []string{commonState.Students[0].User.ID, commonState.Students[0].Parents[0].ID}...)
	}

	for _, receiveID := range userIds {
		var deviceToken string
		row := s.BobDBConn.QueryRow(ctx, "SELECT device_token FROM public.user_device_tokens WHERE user_id = $1", receiveID)
		if err := row.Scan(&deviceToken); err != nil {
			return ctx, fmt.Errorf("error finding user device token with userid: %s: %w", receiveID, err)
		}

		resp, err := retrievePushedNotification(ctx, s.NotificationMgmtGRPCConn, deviceToken)
		if err != nil {
			return ctx, fmt.Errorf("error when call NotificationModifierService.RetrievePushedNotificationMessages: %w", err)
		}
		if len(resp.Messages) == 0 {
			err = fmt.Errorf("wrong node: user receive id: " + receiveID + ", device_token: " + deviceToken)
			return ctx, err
		}

		gotNoti := resp.Messages[len(resp.Messages)-1]
		gotTile := gotNoti.Title
		if gotTile != commonState.NatsNotification.NotificationConfig.Notification.Title {
			return ctx, fmt.Errorf("want notification title to be: %s, got %s", commonState.NatsNotification.NotificationConfig.Notification.Title, gotTile)
		}
	}
	return ctx, nil
}

func (s *PushByNatsSuite) waitToNotificationSend(ctx context.Context, sendType string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	if sendType == consts.NotificationTypeScheduled {
		fmt.Printf("\nWaiting for %s notification to be sent...\n", sendType)
		notify := commonState.NatsNotification
		sendTime, _ := time.Parse(LauoutTimeFormat, notify.SendTime.Time)
		if sendTime.After(time.Now()) {
			waitTime := time.Duration(sendTime.Unix()-time.Now().Unix()+90) * time.Second
			time.Sleep(waitTime)
		}
	} else {
		time.Sleep(10 * time.Second)
	}
	return ctx, nil
}

func (s *PushByNatsSuite) notificationBellDisplayNewNotificationOf(ctx context.Context, newNoti int, target string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	tokens := make([]string, 0)
	switch target {
	case "student":
		tokens = append(tokens, s.studentToken)
	case "parent":
		tokens = append(tokens, s.parentToken)
	default:
		tokens = append(tokens, s.studentToken, s.parentToken)
	}

	for _, token := range tokens {
		countNotificationByUser, err := s.CountUserNotificationByStatus(token, cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_READ)
		if err != nil {
			return ctx, err
		}

		if countNotificationByUser.Total-countNotificationByUser.NumByStatus != int32(newNoti) {
			return ctx, fmt.Errorf("couting notification not equal")
		}
	}

	return common.StepStateToContext(ctx, commonState), nil
}
