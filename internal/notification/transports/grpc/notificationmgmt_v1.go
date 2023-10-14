package grpc

import (
	"context"

	"github.com/manabie-com/backend/internal/notification/services"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
)

func NewNotificationReaderService(notiReaderSvc *services.NotificationReaderService) *NotificationReaderService {
	return &NotificationReaderService{notiReaderSvc: notiReaderSvc}
}

type NotificationReaderService struct {
	notiReaderSvc *services.NotificationReaderService
	npb.UnimplementedNotificationReaderServiceServer
}

func (rcv *NotificationReaderService) RetrieveNotificationDetail(ctx context.Context, rq *npb.RetrieveNotificationDetailRequest) (*npb.RetrieveNotificationDetailResponse, error) {
	return rcv.notiReaderSvc.RetrieveNotificationDetail(ctx, rq)
}

func (rcv *NotificationReaderService) RetrieveNotifications(ctx context.Context, rq *npb.RetrieveNotificationsRequest) (*npb.RetrieveNotificationsResponse, error) {
	return rcv.notiReaderSvc.RetrieveNotifications(ctx, rq)
}

func (rcv *NotificationReaderService) GetAnswersByFilter(ctx context.Context, rq *npb.GetAnswersByFilterRequest) (*npb.GetAnswersByFilterResponse, error) {
	return rcv.notiReaderSvc.GetAnswersByFilter(ctx, rq)
}

func (rcv *NotificationReaderService) CountUserNotification(ctx context.Context, rq *npb.CountUserNotificationRequest) (*npb.CountUserNotificationResponse, error) {
	return rcv.notiReaderSvc.CountUserNotification(ctx, rq)
}

func (rcv *NotificationReaderService) GetNotificationsByFilter(ctx context.Context, rq *npb.GetNotificationsByFilterRequest) (*npb.GetNotificationsByFilterResponse, error) {
	return rcv.notiReaderSvc.GetNotificationsByFilter(ctx, rq)
}

func (rcv *NotificationReaderService) RetrieveGroupAudience(ctx context.Context, rq *npb.RetrieveGroupAudienceRequest) (*npb.RetrieveGroupAudienceResponse, error) {
	return rcv.notiReaderSvc.RetrieveGroupAudience(ctx, rq)
}

func (rcv *NotificationReaderService) GetQuestionnaireAnswersCSV(ctx context.Context, rq *npb.GetQuestionnaireAnswersCSVRequest) (*npb.GetQuestionnaireAnswersCSVResponse, error) {
	return rcv.notiReaderSvc.GetQuestionnaireAnswersCSV(ctx, rq)
}

func (rcv *NotificationReaderService) RetrieveDraftAudience(ctx context.Context, rq *npb.RetrieveDraftAudienceRequest) (*npb.RetrieveDraftAudienceResponse, error) {
	return rcv.notiReaderSvc.RetrieveDraftAudience(ctx, rq)
}

type NotificationModifierService struct {
	notiModifierSvc *services.NotificationModifierService
	npb.NotificationModifierServiceServer
}

func NewNotificationModifierService(notiModifierSvc *services.NotificationModifierService) *NotificationModifierService {
	return &NotificationModifierService{notiModifierSvc: notiModifierSvc}
}

func (rcv *NotificationModifierService) UpsertNotification(ctx context.Context, rq *npb.UpsertNotificationRequest) (*npb.UpsertNotificationResponse, error) {
	return rcv.notiModifierSvc.UpsertNotification(ctx, rq)
}

func (rcv *NotificationModifierService) SendNotification(ctx context.Context, rq *npb.SendNotificationRequest) (*npb.SendNotificationResponse, error) {
	return rcv.notiModifierSvc.SendNotification(ctx, rq)
}

func (rcv *NotificationModifierService) DiscardNotification(ctx context.Context, rq *npb.DiscardNotificationRequest) (*npb.DiscardNotificationResponse, error) {
	return rcv.notiModifierSvc.DiscardNotification(ctx, rq)
}

func (rcv *NotificationModifierService) NotifyUnreadUser(ctx context.Context, rq *npb.NotifyUnreadUserRequest) (*npb.NotifyUnreadUserResponse, error) {
	return rcv.notiModifierSvc.NotifyUnreadUser(ctx, rq)
}

func (rcv *NotificationModifierService) SendScheduledNotification(ctx context.Context, rq *npb.SendScheduledNotificationRequest) (*npb.SendScheduledNotificationResponse, error) {
	return rcv.notiModifierSvc.SendScheduledNotification(ctx, rq)
}

func (rcv *NotificationModifierService) SubmitQuestionnaire(ctx context.Context, rq *npb.SubmitQuestionnaireRequest) (*npb.SubmitQuestionnaireResponse, error) {
	return rcv.notiModifierSvc.SubmitQuestionnaire(ctx, rq)
}

func (rcv *NotificationModifierService) SetStatusForUserNotifications(ctx context.Context, rq *npb.SetStatusForUserNotificationsRequest) (*npb.SetStatusForUserNotificationsResponse, error) {
	return rcv.notiModifierSvc.SetStatusForUserNotifications(ctx, rq)
}

func (rcv *NotificationModifierService) UpdateUserDeviceToken(ctx context.Context, rq *npb.UpdateUserDeviceTokenRequest) (*npb.UpdateUserDeviceTokenResponse, error) {
	return rcv.notiModifierSvc.UpdateUserDeviceToken(ctx, rq)
}

func (rcv *NotificationModifierService) UpsertQuestionnaireTemplate(ctx context.Context, rq *npb.UpsertQuestionnaireTemplateRequest) (*npb.UpsertQuestionnaireTemplateResponse, error) {
	return rcv.notiModifierSvc.UpsertQuestionnaireTemplate(ctx, rq)
}

func (rcv *NotificationModifierService) DeleteNotification(ctx context.Context, rq *npb.DeleteNotificationRequest) (*npb.DeleteNotificationResponse, error) {
	return rcv.notiModifierSvc.DeleteNotification(ctx, rq)
}
