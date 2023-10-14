package grpc

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/notification/infra"
	metrics "github.com/manabie-com/backend/internal/notification/infra/metrics"
	"github.com/manabie-com/backend/internal/notification/services"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/aws/aws-sdk-go/aws/client"
)

func NewYasuoLegacyNotificationModifierService(db database.Ext, c configs.StorageConfig, s3Sess client.ConfigProvider, pushNotificationService infra.PushNotificationService, metrics metrics.NotificationMetrics, jsm nats.JetStreamManagement, env string) *YasuoLegacyNotificationModifier {
	notiModifierSvc := services.NewNotificationModifierService(db, c, s3Sess, pushNotificationService, metrics, jsm, env)
	return &YasuoLegacyNotificationModifier{
		notiModifierSvc: notiModifierSvc,
	}
}

type YasuoLegacyNotificationModifier struct {
	notiModifierSvc *services.NotificationModifierService
	ypb.UnimplementedNotificationModifierServiceServer
}

func (rcv *YasuoLegacyNotificationModifier) UpsertNotification(ctx context.Context, req *ypb.UpsertNotificationRequest) (*ypb.UpsertNotificationResponse, error) {
	res, err := rcv.notiModifierSvc.UpsertNotification(ctx, mappers.YasuoToNotiV1_UpsertNotificationRequest(req))
	if err != nil {
		return nil, err
	}
	return mappers.NotiV1ToYasuo_UpsertNotificationResponse(res), err
}
func (rcv *YasuoLegacyNotificationModifier) SubmitQuestionnaire(ctx context.Context, req *ypb.SubmitQuestionnaireRequest) (*ypb.SubmitQuestionnaireResponse, error) {
	res, err := rcv.notiModifierSvc.SubmitQuestionnaire(ctx, mappers.YasuoToNotiV1_SubmitQuestionnaireRequest(req))
	if err != nil {
		return nil, err
	}
	return mappers.NotiV1ToYasuo_SubmitQuestionnaireResponse(res), err
}
func (rcv *YasuoLegacyNotificationModifier) SendScheduledNotification(ctx context.Context, req *ypb.SendScheduledNotificationRequest) (*ypb.SendScheduledNotificationResponse, error) {
	res, err := rcv.notiModifierSvc.SendScheduledNotification(ctx, mappers.YasuoToNotiV1_SendScheduledNotificationRequest(req))
	if err != nil {
		return nil, err
	}
	return mappers.NotiV1ToYasuo_SendScheduledNotificationResponse(res), err
}
func (rcv *YasuoLegacyNotificationModifier) SendNotification(ctx context.Context, req *ypb.SendNotificationRequest) (*ypb.SendNotificationResponse, error) {
	res, err := rcv.notiModifierSvc.SendNotification(ctx, mappers.YasuoToNotiV1_SendNotificationRequest(req))
	if err != nil {
		return nil, err
	}
	return mappers.NotiV1ToYasuo_SendNotificationResponse(res), err
}
func (rcv *YasuoLegacyNotificationModifier) NotifyUnreadUser(ctx context.Context, req *ypb.NotifyUnreadUserRequest) (*ypb.NotifyUnreadUserResponse, error) {
	res, err := rcv.notiModifierSvc.NotifyUnreadUser(ctx, mappers.YasuoToNotiV1_NotifiUnreadUserRequest(req))
	if err != nil {
		return nil, err
	}
	return mappers.NotiV1ToYasuo_NotifiUnreadUserResponse(res), err
}
func (rcv *YasuoLegacyNotificationModifier) DiscardNotification(ctx context.Context, req *ypb.DiscardNotificationRequest) (*ypb.DiscardNotificationResponse, error) {
	res, err := rcv.notiModifierSvc.DiscardNotification(ctx, mappers.YasuoToNotiV1_DiscardNotificationRequest(req))
	if err != nil {
		return nil, err
	}
	return mappers.NotiV1ToYasuo_DiscardNotificationResponse(res), err
}
