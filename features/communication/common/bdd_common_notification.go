package common

import (
	"context"
	"fmt"

	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
)

func (s *NotificationSuite) UserSetStatusToNotification(ctx context.Context, token, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &bpb.SetUserNotificationStatusRequest{
		NotificationIds: []string{stepState.Notification.NotificationId},
		Status:          cpb.UserNotificationStatus(cpb.UserNotificationStatus_value[status]),
	}

	ctx, cancel := ContextWithTokenAndTimeOut(ctx, token)
	defer cancel()

	_, err := bpb.NewNotificationModifierServiceClient(s.BobGRPCConn).SetUserNotificationStatus(ctx, req)
	if err != nil {
		return ctx, fmt.Errorf("failed SetUserNotificationStatus: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *NotificationSuite) GetNotificationByUser(userToken string, importantOnly bool) ([]*npb.RetrieveNotificationsResponse_NotificationInfo, error) {
	ctx, cancel := ContextWithTokenAndTimeOut(context.Background(), userToken)
	defer cancel()

	retrieveRequest := &npb.RetrieveNotificationsRequest{
		ImportantOnly: importantOnly,
	}
	res, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveNotifications(ctx, retrieveRequest)
	if err != nil {
		return nil, fmt.Errorf("failed retrieve user notifications: %v", err)
	}

	return res.Items, nil
}

func (s *NotificationSuite) CountUserNotificationByStatus(userToken string, status cpb.UserNotificationStatus) (*npb.CountUserNotificationResponse, error) {
	ctx, cancel := ContextWithTokenAndTimeOut(context.Background(), userToken)
	defer cancel()

	req := &npb.CountUserNotificationRequest{
		Status: cpb.UserNotificationStatus(status),
	}
	res, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).CountUserNotification(ctx, req)

	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *NotificationSuite) GetDetailNotification(userToken, notiID string) (*npb.RetrieveNotificationDetailResponse, error) {
	ctx, cancel := ContextWithTokenAndTimeOut(context.Background(), userToken)
	defer cancel()

	req := &npb.RetrieveNotificationDetailRequest{
		NotificationId: notiID,
	}
	res, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveNotificationDetail(ctx, req)

	if err != nil {
		return nil, err
	}

	return res, nil
}
