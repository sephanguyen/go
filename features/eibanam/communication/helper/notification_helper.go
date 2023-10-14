package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/eibanam"
	"github.com/manabie-com/backend/features/eibanam/communication/entity"
	"github.com/manabie-com/backend/features/eibanam/communication/util"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/gogo/protobuf/types"
	"github.com/hasura/go-graphql-client"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *CommunicationHelper) CreateNotification(schoolAdmin *entity.Admin, notify *entity.Notification) error {
	cpbNotification := h.NewNotification(schoolAdmin, notify)

	ctx, cancel := util.ContextWithTokenAndTimeOut(context.Background(), schoolAdmin.Token)
	defer cancel()

	req := &ypb.UpsertNotificationRequest{
		Notification: cpbNotification,
	}
	resp, err := ypb.NewNotificationModifierServiceClient(h.yasuoGRPCConn).UpsertNotification(ctx, req)
	if err != nil {
		return err
	}

	notify.ID = resp.NotificationId
	return nil
}

func (h *CommunicationHelper) CreateNotificationWithMediaUrl(schoolAdmin *entity.Admin, notify *entity.Notification, mediaIds []string) error {
	cpbNotification := h.NewNotification(schoolAdmin, notify)
	cpbNotification.Message.MediaIds = mediaIds

	ctx, cancel := util.ContextWithTokenAndTimeOut(context.Background(), schoolAdmin.Token)
	defer cancel()

	req := &ypb.UpsertNotificationRequest{
		Notification: cpbNotification,
	}
	resp, err := ypb.NewNotificationModifierServiceClient(h.yasuoGRPCConn).UpsertNotification(ctx, req)
	if err != nil {
		return err
	}

	notify.ID = resp.NotificationId
	return nil
}

func (h *CommunicationHelper) SendNotification(schoolAdmin *entity.Admin, notify *entity.Notification) error {
	ctx, cancel := util.ContextWithTokenAndTimeOut(context.Background(), schoolAdmin.Token)
	defer cancel()

	switch notify.Status {
	case cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED, cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT:
		req := &ypb.SendNotificationRequest{
			NotificationId: notify.ID,
		}
		_, err := ypb.NewNotificationModifierServiceClient(h.yasuoGRPCConn).SendNotification(ctx, req)
		if err != nil {
			return err
		}
	default:
		return errors.New("notification status unsupported to send")
	}
	notify.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_SENT
	return nil
}

func (h *CommunicationHelper) UpdateNotificationWithProto(schoolAdmin *entity.Admin, notify *cpb.Notification) error {
	ctx, cancel := util.ContextWithTokenAndTimeOut(context.Background(), schoolAdmin.Token)
	defer cancel()

	req := &ypb.UpsertNotificationRequest{
		Notification: notify,
	}

	_, err := ypb.NewNotificationModifierServiceClient(h.yasuoGRPCConn).UpsertNotification(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func (h *CommunicationHelper) UpdateNotification(schoolAdmin *entity.Admin, notify *entity.Notification) error {
	cpbNotification := h.NewNotification(schoolAdmin, notify)
	cpbNotification.NotificationId = notify.ID

	ctx, cancel := util.ContextWithTokenAndTimeOut(context.Background(), schoolAdmin.Token)
	defer cancel()

	req := &ypb.UpsertNotificationRequest{
		Notification: cpbNotification,
	}

	_, err := ypb.NewNotificationModifierServiceClient(h.yasuoGRPCConn).UpsertNotification(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func (h *CommunicationHelper) DiscardNotification(schoolAdmin *entity.Admin, notify *entity.Notification) error {

	ctx, cancel := util.ContextWithTokenAndTimeOut(context.Background(), schoolAdmin.Token)
	defer cancel()

	req := &ypb.DiscardNotificationRequest{
		NotificationId: notify.ID,
	}

	_, err := ypb.NewNotificationModifierServiceClient(h.yasuoGRPCConn).DiscardNotification(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func (h *CommunicationHelper) GetCmsNotification(schoolAdmin *entity.Admin, notificationId string) (*entity.GraphqlNotification, error) {
	if err := eibanam.TrackTableForHasuraQuery(
		h.hasuraAdminUrl,
		"info_notifications",
	); err != nil {
		return nil, errors.Wrap(err, "trackTableForHasuraQuery()")
	}

	if err := eibanam.CreateSelectPermissionForHasuraQuery(
		h.hasuraAdminUrl,
		constant.UserGroupSchoolAdmin,
		"info_notifications",
	); err != nil {
		return nil, errors.Wrap(err, "createSelectPermissionForHasuraQuery()")
	}

	rawQuery := `query ($notification_id:String!){
					info_notifications(where: {notification_id: {_eq: $notification_id}}){
					notification_id,
					notification_msg_id,
					sent_at,
					receiver_ids,
					status,
					type,
					target_groups,
					created_at,
					updated_at,
					editor_id,
					event,
					scheduled_at
				}
			}`

	if err := eibanam.AddQueryToAllowListForHasuraQuery(h.hasuraAdminUrl, rawQuery); err != nil {
		return nil, fmt.Errorf("addQueryToAllowListForHasuraQuery(): %v", err)
	}

	if err := h.GrantPermissionToQueryGraphql(schoolAdmin, "info_notifications"); err != nil {
		return nil, err
	}

	variables := map[string]interface{}{
		"notification_id": graphql.String(notificationId),
	}
	var query entity.GraphqlNotificationQuery
	res, err := h.QueryHasura(schoolAdmin, &query, variables)
	if err != nil {
		return nil, err
	}

	var result entity.GraphqlNotification
	if err := json.Unmarshal(res, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (h *CommunicationHelper) GetCmsNotificationMsg(schoolAdmin *entity.Admin, msgId string) (*entity.GraphqlNotificationMsg, error) {
	if err := eibanam.TrackTableForHasuraQuery(
		h.hasuraAdminUrl,
		"info_notification_msgs",
	); err != nil {
		return nil, errors.Wrap(err, "trackTableForHasuraQuery()")
	}

	if err := eibanam.CreateSelectPermissionForHasuraQuery(
		h.hasuraAdminUrl,
		constant.UserGroupSchoolAdmin,
		"info_notification_msgs",
	); err != nil {
		return nil, errors.Wrap(err, "createSelectPermissionForHasuraQuery()")
	}

	rawQuery := `query ($notification_msg_id:String!){
						info_notification_msgs(where: {notification_msg_id: {_eq: $notification_msg_id}}){
							content,
							created_at,
							media_ids,
							notification_msg_id,
							title,
							updated_at
						}
					}`

	if err := eibanam.AddQueryToAllowListForHasuraQuery(h.hasuraAdminUrl, rawQuery); err != nil {
		return nil, fmt.Errorf("addQueryToAllowListForHasuraQuery(): %v", err)
	}

	if err := h.GrantPermissionToQueryGraphql(schoolAdmin, "info_notification_msgs"); err != nil {
		return nil, err
	}

	variables := map[string]interface{}{
		"notification_msg_id": graphql.String(msgId),
	}
	var query entity.GraphqlNotificationMsgQuery
	res, err := h.QueryHasura(schoolAdmin, &query, variables)
	if err != nil {
		return nil, err
	}

	var result entity.GraphqlNotificationMsg
	if err := json.Unmarshal(res, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (h *CommunicationHelper) NewNotification(schoolAdmin *entity.Admin, notify *entity.Notification) *cpb.Notification {
	jsonData, _ := json.Marshal(notify.Data)

	infoNotification := &cpb.Notification{
		Data:     string(jsonData),
		EditorId: schoolAdmin.ID,
		Message: &cpb.NotificationMessage{
			Title: notify.Title,
			Content: &cpb.RichText{
				Raw:      notify.Content,
				Rendered: notify.HTMLContent,
			},
		},
		Status:      notify.Status,
		SchoolId:    notify.SchoolID,
		Type:        cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED,
		ReceiverIds: notify.IndividualReceivers,
		TargetGroup: &cpb.NotificationTargetGroup{
			CourseFilter: &cpb.NotificationTargetGroup_CourseFilter{
				Type:      notify.FilterByCourse.Type,
				CourseIds: notify.FilterByCourse.Courses,
			},
			GradeFilter: &cpb.NotificationTargetGroup_GradeFilter{
				Type:   notify.FilterByGrade.Type,
			},
			UserGroupFilter: &cpb.NotificationTargetGroup_UserGroupFilter{
				UserGroups: notify.ReceiverGroup,
			},
			LocationFilter: &cpb.NotificationTargetGroup_LocationFilter{
				Type:        notify.FilterByLocation.Type,
				LocationIds: notify.FilterByLocation.Locations,
			},
			ClassFilter: &cpb.NotificationTargetGroup_ClassFilter{
				Type:     notify.FilterByClass.Type,
				ClassIds: notify.FilterByClass.Classes,
			},
		},
		ScheduledAt: timestamppb.New(notify.ScheduledAt),
	}
	return infoNotification
}
func (h *CommunicationHelper) CountUserUnreadNotification(userToken string) (int32, error) {
	ctx, cancel := util.ContextWithTokenAndTimeOut(context.Background(), userToken)
	defer cancel()

	req := &bpb.CountUserNotificationRequest{
		Status: cpb.UserNotificationStatus(bpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_READ),
	}
	res, err := bpb.NewNotificationReaderServiceClient(h.bobGRPCConn).CountUserNotification(ctx, req)

	if err != nil {
		return 0, err
	}
	return res.GetTotal() - res.GetNumByStatus(), nil
}

func (h *CommunicationHelper) ReadNotification(ctx context.Context, userToken string, notification string) error {
	ctx, cancel := util.ContextWithTokenAndTimeOut(ctx, userToken)
	defer cancel()
	_, err := bpb.NewNotificationModifierServiceClient(h.bobGRPCConn).SetUserNotificationStatus(ctx, &bpb.SetUserNotificationStatusRequest{
		NotificationIds: []string{notification},
		Status:          cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_READ,
	})
	return err
}

func (h *CommunicationHelper) GetUserNotification(userToken string, importantOnly bool) ([]*bpb.RetrieveNotificationsResponse_NotificationInfo, error) {
	ctx, cancel := util.ContextWithTokenAndTimeOut(context.Background(), userToken)
	defer cancel()
	req := &bpb.RetrieveNotificationsRequest{
		ImportantOnly: importantOnly,
	}
	res, err := bpb.NewNotificationReaderServiceClient(h.bobGRPCConn).RetrieveNotifications(ctx, req)

	if err != nil {
		return nil, err
	}

	return res.Items, nil
}

func (h *CommunicationHelper) CountingUserNotification(userToken string) (*bpb.CountUserNotificationResponse, error) {
	ctx, cancel := util.ContextWithTokenAndTimeOut(context.Background(), userToken)
	defer cancel()

	req := &bpb.CountUserNotificationRequest{
		Status: cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_READ,
	}
	res, err := bpb.NewNotificationReaderServiceClient(h.bobGRPCConn).CountUserNotification(ctx, req)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (h *CommunicationHelper) generateMedia() *bob_pb.Media {
	return &bob_pb.Media{MediaId: "", Name: fmt.Sprintf("random-name-%s", idutil.ULIDNow()), Resource: idutil.ULIDNow(), CreatedAt: types.TimestampNow(), UpdatedAt: types.TimestampNow(), Comments: []*bob_pb.Comment{{Comment: "Comment-1", Duration: types.DurationProto(10 * time.Second)}, {Comment: "Comment-2", Duration: types.DurationProto(20 * time.Second)}}, Type: bob_pb.MEDIA_TYPE_VIDEO}
}

func (h *CommunicationHelper) CreateMediaViaGRPC(authToken string, numberMedia int) ([]string, error) {
	mediaList := []*bob_pb.Media{}
	for i := 0; i < numberMedia; i++ {
		mediaList = append(mediaList, h.generateMedia())
	}
	req := &bob_pb.UpsertMediaRequest{
		Media: mediaList,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = util.ContextWithToken(ctx, authToken)
	res, err := bob_pb.NewClassClient(h.bobGRPCConn).UpsertMedia(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("bob.UpsertMedia: %v", err)
	}

	return res.MediaIds, nil
}

func (h *CommunicationHelper) GetDetailNotification(authToken string, notificationId string) (*bpb.RetrieveNotificationDetailResponse, error) {
	ctx, cancel := util.ContextWithTokenAndTimeOut(context.Background(), authToken)
	defer cancel()

	req := &bpb.RetrieveNotificationDetailRequest{
		NotificationId: notificationId,
	}
	res, err := bpb.NewNotificationReaderServiceClient(h.bobGRPCConn).RetrieveNotificationDetail(ctx, req)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (h *CommunicationHelper) UpdateDeviceToken(token string, deviceToken string, userId string) error {
	ctx, cancel := util.ContextWithTokenAndTimeOut(context.Background(), token)
	defer cancel()

	_, err := bob_pb.NewUserServiceClient(h.bobGRPCConn).UpdateUserDeviceToken(ctx, &bob_pb.UpdateUserDeviceTokenRequest{
		DeviceToken:       deviceToken,
		UserId:            userId,
		AllowNotification: true,
	})

	if err != nil {
		return fmt.Errorf("err bpb.UpdateUserDeviceToken: %v", err)
	}

	return nil
}
