package bob

import (
	"context"

	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

const (
	RetrieveNotificationRequest = "RetrieveNotificationRequest"
	RetrieveNotificationStats   = "RetrieveNotificationStats"
)

func (s *suite) aValidRequest(ctx context.Context, reqName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch reqName {
	case RetrieveNotificationRequest:
		stepState.Request = &pb.RetrieveNotificationRequest{}
	case RetrieveNotificationStats:
		stepState.Request = &pb.NotificationStatsRequest{}
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) anInvalidRequest(ctx context.Context, reqName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch reqName {
	case RetrieveNotificationRequest:
		stepState.Request = &pb.RetrieveNotificationRequest{Type: pb.NOTIFICATION_TYPE_PROMO_CODE, Page: 1, Limit: 1}
	case RetrieveNotificationStats:
		stepState.Request = &pb.NotificationStatsRequest{}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userTryToMakeRequest(ctx context.Context, reqName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch reqName {
	case RetrieveNotificationRequest:
		return s.makeRetrieveNotifications(ctx)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) requestHasPageAndLimitAndType(ctx context.Context, reqName string, page, limit int32, notificationType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = &pb.RetrieveNotificationRequest{Type: pb.NotificationType(pb.NotificationType_value[notificationType]), Page: page, Limit: limit}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) makeRetrieveNotifications(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewNotificationServiceClient(s.Conn).RetrieveNotifications(s.signedCtx(ctx), stepState.Request.(*pb.RetrieveNotificationRequest))

	return StepStateToContext(ctx, stepState), nil
}
