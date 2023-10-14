package grpc

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/services"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
)

func NewBobLegacyNotificationReaderService(db database.Ext, env string) *BobLegacyNotificationReader {
	notiReaderSvc := services.NewNotificationReaderService(db, env)
	return &BobLegacyNotificationReader{
		notiReaderSvc: notiReaderSvc,
	}
}

type BobLegacyNotificationReader struct {
	notiReaderSvc *services.NotificationReaderService
	bpb.UnimplementedNotificationReaderServiceServer
}

func (rcv *BobLegacyNotificationReader) RetrieveNotificationDetail(ctx context.Context, req *bpb.RetrieveNotificationDetailRequest) (*bpb.RetrieveNotificationDetailResponse, error) {
	res, err := rcv.notiReaderSvc.RetrieveNotificationDetail(ctx, mappers.BobToNotiV1_RetrieveNotificationDetailRequest(req))
	if err != nil {
		return nil, err
	}
	return mappers.NotiV1ToBob_RetrieveNotificationDetailResponse(res), err
}

func (rcv *BobLegacyNotificationReader) RetrieveNotifications(ctx context.Context, req *bpb.RetrieveNotificationsRequest) (*bpb.RetrieveNotificationsResponse, error) {
	res, err := rcv.notiReaderSvc.RetrieveNotifications(ctx, mappers.BobToNotiV1_RetrieveNotificationsRequest(req))
	if err != nil {
		return nil, err
	}
	return mappers.NotiV1ToBob_RetrieveNotificationsResponse(res), err
}

func (rcv *BobLegacyNotificationReader) GetAnswersByFilter(ctx context.Context, req *bpb.GetAnswersByFilterRequest) (*bpb.GetAnswersByFilterResponse, error) {
	res, err := rcv.notiReaderSvc.GetAnswersByFilter(ctx, mappers.BobToNotiV1_GetAnswersByFilterRequest(req))
	if err != nil {
		return nil, err
	}
	return mappers.NotiV1ToBob_GetAnswersByFilterResponse(res), err
}

func (rcv *BobLegacyNotificationReader) CountUserNotification(ctx context.Context, req *bpb.CountUserNotificationRequest) (*bpb.CountUserNotificationResponse, error) {
	res, err := rcv.notiReaderSvc.CountUserNotification(ctx, mappers.BobToNotiV1_CountUserNotificationRequest(req))
	if err != nil {
		return nil, err
	}
	return mappers.NotiV1ToBob_CountUserNotificationResponse(res), err
}

// nolint
func NewSimpleBobLegacyNotificationModifierService(db database.Ext) *BobLegacyNotificationModifier {
	notiModifierSvc := services.NewSimpleNotificationModifierService(db)
	return &BobLegacyNotificationModifier{
		notiModifierSvc: notiModifierSvc,
	}
}

type BobLegacyNotificationModifier struct {
	notiModifierSvc *services.NotificationModifierService
	bpb.UnimplementedNotificationModifierServiceServer
}

func (rcv *BobLegacyNotificationModifier) SetUserNotificationStatus(ctx context.Context, req *bpb.SetUserNotificationStatusRequest) (*bpb.SetUserNotificationStatusResponse, error) {
	res, err := rcv.notiModifierSvc.SetUserNotificationStatus(ctx, mappers.BobToNotiV1_SetUserNotificationStatusRequest(req))
	if err != nil {
		return nil, err
	}
	return mappers.NotiV1ToBob_SetUserNotificationStatusResponse(res), err
}

// backward compatible with old api /manabie.bob.UserService/UpdateUserDeviceToken
type BobLegacyUserService struct {
	pb.UnimplementedUserServiceServer
	notiModifierSvc *services.NotificationModifierService
}

func (rvc *BobLegacyUserService) UpdateUserDeviceToken(ctx context.Context, req *pb.UpdateUserDeviceTokenRequest) (*pb.UpdateUserDeviceTokenResponse, error) {
	res, err := rvc.notiModifierSvc.UpdateUserDeviceToken(ctx, mappers.BobToNotiV1_UpdateUserDeviceTokenRequest(req))
	if err != nil {
		return nil, err
	}
	return mappers.NotiV1ToBob_UpdateUserDeviceTokenResponse(res), err
}

func NewBobLegacyUserService(notiModifierSvc *services.NotificationModifierService) *BobLegacyUserService {
	return &BobLegacyUserService{
		notiModifierSvc: notiModifierSvc,
	}
}
