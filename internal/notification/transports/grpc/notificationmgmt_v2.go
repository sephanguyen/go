package grpc

import (
	"context"

	"github.com/manabie-com/backend/internal/notification/services"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
	npbv2 "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v2"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewNotificationReaderV2Service(notiReaderSvc *services.NotificationReaderService) *NotificationReaderV2Service {
	return &NotificationReaderV2Service{notiReaderSvc: notiReaderSvc}
}

type NotificationReaderV2Service struct {
	notiReaderSvc *services.NotificationReaderService
	npbv2.UnimplementedNotificationReaderServiceServer
}

func (rcv *NotificationReaderV2Service) RetrieveNotificationDetail(ctx context.Context, rq *npbv2.RetrieveNotificationDetailRequest) (*npbv2.RetrieveNotificationDetailResponse, error) {
	if rq.GetUserNotificationId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_notification_id doesn't exist in request")
	}
	compatibleReq := &npb.RetrieveNotificationDetailRequest{
		UserNotificationId: rq.GetUserNotificationId(),
	}
	res, err := rcv.notiReaderSvc.RetrieveNotificationDetail(ctx, compatibleReq)
	if err != nil {
		return &npbv2.RetrieveNotificationDetailResponse{}, err
	}
	return mappers.NotiV1ToV2_RetrieveNotificationDetailResponse(res), nil
}
