package communication

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
)

type AdminCreateAttachmentNotificationSuite struct {
	*common.NotificationSuite
	receivedNotify *npb.RetrieveNotificationsResponse_NotificationInfo
}
type PromoteObj struct {
	Code      string `json:"code"`
	Type      string `json:"type"`
	Amount    string `json:"amount"`
	ExpiredAt string `json:"expired_at"`
}
type CustomDataNoti struct {
	Promote  []PromoteObj `json:"promote"`
	ImageURL string       `json:"image_url"`
}

func (c *SuiteConstructor) InitSendNotificationWithAttachment(dep *DependencyV2, ctx *godog.ScenarioContext) {
	s := &AdminCreateAttachmentNotificationSuite{
		NotificationSuite: dep.notiCommonSuite,
	}
	stepsMapping := map[string]interface{}{
		`^admin create media attachment notification and sent$`:                                                                     s.adminCreateMediaAttachmentNotification,
		`^student receive media attachment notification$`:                                                                           s.studentReceiveMediaAttachmentNotification,
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" students with "([^"]*)" parents info for each student$`:                                    s.CreatesNumberOfStudentsWithParentsInfo,
		`^school admin creates "([^"]*)" courses$`:                                                                                  s.CreatesNumberOfCourses,
		`^school admin add packages data of those courses for each student$`:                                                        s.AddPackagesDataOfThoseCoursesForEachStudent,
		`^student "([^"]*)" logins to Learner App$`:                                                                                 s.StudentLoginsToLearnerApp,
		`^update user device token to an "([^"]*)" device token$`:                                                                   s.UpdateDeviceTokenForLeanerUser,
	}
	c.InitScenarioStepMapping(ctx, stepsMapping)
}

func (s *AdminCreateAttachmentNotificationSuite) adminCreateMediaAttachmentNotification(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	mediaIDs, err := s.CreateMediaViaGRPC(commonState.CurrentStaff.Token, 1)
	if err != nil {
		return ctx, fmt.Errorf("s.util.CreateMediaViaGRPC: %v", err)
	}

	opts := &common.NotificationWithOpts{
		UserGroups:       "student, parent",
		CourseFilter:     "random",
		GradeFilter:      "random",
		LocationFilter:   "none",
		ClassFilter:      "none",
		IndividualFilter: "none",
		ScheduledStatus:  "1 min",
		Status:           "NOTIFICATION_STATUS_DRAFT",
		IsImportant:      false,
		MediaIds:         mediaIDs,
		ReceiverIds:      []string{commonState.Students[0].ID},
	}
	ctx, commonState.Notification, err = s.GetNotificationWithOptions(ctx, opts)
	if err != nil {
		return common.StepStateToContext(ctx, commonState), fmt.Errorf("GetNotificationWithOptions: %v", err)
	}

	commonState.Request = &npb.UpsertNotificationRequest{
		Notification: commonState.Notification,
	}

	commonState.Response, commonState.ResponseErr = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpsertNotification(ctx, commonState.Request.(*npb.UpsertNotificationRequest))

	if commonState.ResponseErr == nil {
		resp := commonState.Response.(*npb.UpsertNotificationResponse)
		commonState.Notification.NotificationId = resp.NotificationId
	}

	_, err = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).SendNotification(
		ctx,
		&npb.SendNotificationRequest{NotificationId: commonState.Notification.NotificationId},
	)
	if err != nil {
		return common.StepStateToContext(ctx, commonState), err
	}
	return common.StepStateToContext(ctx, commonState), nil
}

func (s *AdminCreateAttachmentNotificationSuite) studentReceiveMediaAttachmentNotification(ctx context.Context) (context.Context, error) {
	// state := util.StateFromContext(ctx)
	commonState := common.StepStateFromContext(ctx)
	token := commonState.Students[0].Token

	notifications, err := s.GetNotificationByUser(token, false)

	if err != nil {
		return ctx, err
	}
	notificationMsg := commonState.Notification.Message
	// compare by title
	for _, notify := range notifications {
		notificationDetail, err := s.GetDetailNotification(token, notify.UserNotification.NotificationId)
		if err != nil {
			return ctx, fmt.Errorf("does not get detail notification")
		}

		if notificationDetail.Item.Message == nil {
			return ctx, fmt.Errorf("can not find message in notification")
		}

		if len(notificationDetail.Item.Message.MediaIds) != len(notificationMsg.MediaIds) {
			return ctx, fmt.Errorf("media does not equal")
		}

		if !stringutil.SliceElementsMatch(notificationDetail.Item.Message.MediaIds, notificationMsg.MediaIds) {
			return ctx, fmt.Errorf("media does not equal")
		}

		ctx, err := s.checkSystemExistMedia(ctx, notificationDetail.Item.Message.MediaIds)
		if err != nil {
			return ctx, err
		}

		if notify.Title == commonState.Notification.Message.Title {
			var dataBeforeSend CustomDataNoti
			var dataAfterSend CustomDataNoti

			err = json.Unmarshal([]byte(commonState.Notification.Data), &dataBeforeSend)

			if err != nil {
				return ctx, fmt.Errorf("cannot parse data to custom struct1: %v", err)
			}

			err = json.Unmarshal([]byte(notify.UserNotification.Data), &dataAfterSend)

			if err != nil {
				return ctx, fmt.Errorf("cannot parse data to custom struct2: %v", err)
			}

			if dataAfterSend.ImageURL != dataBeforeSend.ImageURL {
				return ctx, fmt.Errorf("notify data is not match: ImageURL")
			}

			if len(dataAfterSend.Promote) != len(dataBeforeSend.Promote) {
				return ctx, fmt.Errorf("notify data is not match: Promote")
			}

			if notify.UserNotification.Type.String() != cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String() {
				return ctx, fmt.Errorf("notify type is not match")
			}

			s.receivedNotify = notify
			return common.StepStateToContext(ctx, commonState), nil
		}
	}

	return ctx, fmt.Errorf("target was not received notification")
}

func (s *AdminCreateAttachmentNotificationSuite) waitToNotificationSend(ctx context.Context) (context.Context, error) {
	time.Sleep(5 * time.Second)
	return ctx, nil
}

func (s *AdminCreateAttachmentNotificationSuite) checkSystemExistMedia(ctx context.Context, mediaIDs []string) (context.Context, error) {
	query := `
		SELECT count(*) AS total_media
		FROM media
		WHERE media_id = ANY($1)
			AND deleted_at IS NULL;
	`
	var totalMedia uint32
	row := s.BobDBConn.QueryRow(ctx, query, mediaIDs)
	err := row.Scan(&totalMedia)

	if err != nil {
		return ctx, err
	}

	if int(totalMedia) < len(mediaIDs) {
		return ctx, fmt.Errorf("media doesn't exist in system, expected: %d, got: %d", len(mediaIDs), totalMedia)
	}

	return ctx, nil
}
