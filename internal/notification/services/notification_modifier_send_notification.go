package services

import (
	"context"
	"fmt"
	"time"

	bobEntities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	"github.com/manabie-com/backend/internal/notification/services/validation"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *NotificationModifierService) SendNotification(ctx context.Context, req *npb.SendNotificationRequest) (*npb.SendNotificationResponse, error) {
	userInfo := golibs.UserInfoFromCtx(ctx)
	userID := interceptors.UserIDFromContext(ctx)
	if userID == "" {
		// Support context from nats
		userID = userInfo.UserID
	}
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id doesn't exist in request")
	}

	notification, err := svc.findSendableNotification(ctx, req.NotificationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	notificationMsg, err := svc.findNotificationMsg(ctx, notification.NotificationMsgID.String)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("FindNotificationMsg: %v", err))
	}

	if err := validation.ValidateNotification(notificationMsg, notification); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("SendNotification.validateNotification: %v", err))
	}

	schoolID := notification.Owner.Int
	err = svc.sendNotification(ctx, notification, notificationMsg, schoolID, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("sendNotification: %v", err))
	}
	resp := &npb.SendNotificationResponse{}
	return resp, nil
}

func (svc *NotificationModifierService) SendNotificationToTargetWithoutSave(ctx context.Context, notification *cpb.Notification) error {
	if len(notification.GenericReceiverIds) == 0 {
		err := validation.ValidateTargetGroup(notification)
		if err != nil {
			return err
		}
	}

	err := validation.ValidateMessageRequiredField(notification)
	if err != nil {
		return err
	}

	// get infor notification and infor notification msgs
	infoNotificationMsg, err := mappers.PbToInfoNotificationMsgEnt(notification.Message)
	if err != nil {
		return fmt.Errorf("cannot convert toInfoNotificationMsgEnt")
	}
	infoNotification, err := mappers.PbToInfoNotificationEnt(notification)
	if err != nil {
		return fmt.Errorf("cannot convert toInfoNotificationEnt")
	}

	audiences, err := svc.NotificationAudienceRetriever.FindAudiences(ctx, svc.DB, infoNotification)
	if err != nil {
		return fmt.Errorf("failed FindAudiences: %v", err)
	}

	receiverIDs := make([]string, 0)
	for _, audience := range audiences {
		receiverIDs = append(receiverIDs, audience.UserID.String)
	}

	// send notification
	if _, _, err = svc.pushNotificationToUsers(ctx, svc.DB, infoNotification, infoNotificationMsg, receiverIDs); err != nil {
		return err
	}

	return nil
}

func (svc *NotificationModifierService) sendNotification(ctx context.Context, notification *entities.InfoNotification, notificationMsg *entities.InfoNotificationMsg, schoolID int32, userID string) error {
	userIDs := make([]string, 0)
	err := database.ExecInTxWithRetry(ctx, svc.DB, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		userIDs, err = svc.addNotificationForUsers(ctx, tx, notification, notificationMsg, schoolID)
		if err != nil {
			return fmt.Errorf("AddNotificationForUsers: %v", err)
		}

		err = svc.InfoNotificationRepo.UpdateNotification(ctx, tx, notification.NotificationID, map[string]interface{}{
			"status":  cpb.NotificationStatus_NOTIFICATION_STATUS_SENT.String(),
			"sent_at": time.Now(),
		})
		if err != nil {
			return fmt.Errorf("InfoNotificationRepo.UpdateNotification: %v", err)
		}

		return nil
	})
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("ExecInTxWithRetry: %v", err))
	}

	userInfo := golibs.UserInfoFromCtx(ctx)
	// Send FCM
	go func(resourcePathCtx, userCtx string, db database.QueryExecer, logger *zap.Logger, userIDs []string, notification *entities.InfoNotification, notificationMsg *entities.InfoNotificationMsg) {
		// Will send in background, if use outside context, will get context cancel of gRPC context.
		fcmContext := interceptors.ContextWithJWTClaims(context.Background(), &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{
				ResourcePath: resourcePathCtx,
				UserID:       userCtx,
			},
		})

		userIDs = golibs.GetUniqueElementStringArray(userIDs)
		successCount, failureCount, err := svc.pushNotificationToUsers(fcmContext, db, notification, notificationMsg, userIDs)
		if err != nil {
			logger.Error("send notification for users using FCM occurred an error: " + err.Error())
		}

		activityLogEnt := &bobEntities.ActivityLog{
			ID:         database.Text(idutil.ULIDNow()),
			UserID:     database.Text(userID),
			ActionType: database.Text("send_notification"),
			Payload: database.JSONB(struct {
				NotificationID string `json:"notification_id"`
				SuccessCount   int    `json:"success_count"`
				FailureCount   int    `json:"failure_count"`
			}{
				NotificationID: notification.NotificationID.String,
				SuccessCount:   successCount,
				FailureCount:   failureCount,
			}),
		}
		err = svc.ActivityLogRepo.Create(fcmContext, db, activityLogEnt)
		if err != nil {
			logger.Error("ActivityLogRepo.Create: " + err.Error())
		}
	}(userInfo.ResourcePath, userInfo.UserID, svc.DB, ctxzap.Extract(ctx), userIDs, notification, notificationMsg)

	return nil
}

// nolint
func (svc *NotificationModifierService) addNotificationForUsers(ctx context.Context, tx pgx.Tx, notification *entities.InfoNotification, notificationMsg *entities.InfoNotificationMsg, schoolID int32) ([]string, error) {
	// find notification user first in case that already create
	userNotifyFilter := repositories.NewFindUserNotificationFilter()
	userNotifyFilter.NotiIDs = database.TextArray([]string{notification.NotificationID.String})

	userNotifications, err := svc.UserNotificationRepo.Find(ctx, tx, userNotifyFilter)
	if err != nil {
		return nil, fmt.Errorf("svc.UserNotificationRepo.Find %w", err)
	}

	if len(userNotifications) == 0 {
		audiences, err := svc.NotificationAudienceRetriever.FindAudiences(ctx, tx, notification)
		if err != nil {
			return nil, fmt.Errorf("failed FindAudiences: %v", err)
		}

		userNotifications, err = svc.toUserNotifications(notification.NotificationID.String, audiences)
		if err != nil {
			return nil, fmt.Errorf("toUserNotifications: %v", err)
		}

		// Assign user name for user notification
		userNotifications, err = svc.DataRetentionService.AssignRetentionNameForUserNotification(ctx, tx, userNotifications)
		if err != nil {
			return nil, fmt.Errorf("svc.DataRetentionService.AssignRetentionNameForUserNotification: %v", err)
		}

		err = svc.UserNotificationRepo.Upsert(ctx, tx, userNotifications)
		if err != nil {
			return nil, fmt.Errorf("UserNotificationRepo.Upsert: %v", err)
		}
	}

	svc.RecordUserNotificationCreated(float64(len(userNotifications)))

	userIDs := make([]string, len(userNotifications))
	for i, un := range userNotifications {
		userIDs[i] = un.UserID.String
	}

	return userIDs, nil
}

func (svc *NotificationModifierService) toUserNotifications(notificationID string, audiences []*entities.Audience) ([]*entities.UserInfoNotification, error) {
	userNotifications := make([]*entities.UserInfoNotification, 0, len(audiences))
	for _, audience := range audiences {
		un, err := mappers.AudienceToUserNotificationEnt(notificationID, audience)
		if err != nil {
			return nil, err
		}
		userNotifications = append(userNotifications, un)
	}

	return userNotifications, nil
}

func (svc *NotificationModifierService) findNotificationByID(ctx context.Context, db database.Ext, notificationID string) (*entities.InfoNotification, error) {
	filter := repositories.NewFindNotificationFilter()

	_ = filter.Status.Set(nil)
	_ = filter.NotiIDs.Set([]string{notificationID})

	es, err := svc.InfoNotificationRepo.Find(ctx, db, filter)
	if err != nil {
		return nil, fmt.Errorf("InfoNotificationRepo.Find: %v", err)
	}
	if len(es) == 0 {
		return nil, fmt.Errorf("InfoNotificationRepo.Find: can not find notification with id %v", notificationID)
	}
	notification := es[0]

	return notification, nil
}

func (svc *NotificationModifierService) findSendableNotification(ctx context.Context, notificationID string) (*entities.InfoNotification, error) {
	notification, err := svc.findNotificationByID(ctx, svc.DB, notificationID)

	if err != nil {
		isDeleted, err := svc.InfoNotificationRepo.IsNotificationDeleted(ctx, svc.DB, database.Text(notificationID))

		if err != nil {
			return nil, fmt.Errorf("svc.InfoNotificationRepo.IsNotificationDeleted: %w", err)
		}
		if isDeleted {
			return nil, fmt.Errorf("the notification has been deleted, you can no longer send this notification")
		}
	}

	if notification.Status.String == cpb.NotificationStatus_NOTIFICATION_STATUS_SENT.String() {
		return nil, fmt.Errorf("the notification has been sent")
	}

	if notification.Status.String != cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT.String() && notification.Status.String != cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String() {
		return nil, fmt.Errorf("notification %s has status %s and not sendable", notification.NotificationID.String, notification.Status.String)
	}

	return notification, nil
}

func (svc *NotificationModifierService) findNotificationMsg(ctx context.Context, notificationMsgID string) (*entities.InfoNotificationMsg, error) {
	es, err := svc.InfoNotificationMsgRepo.GetByIDs(ctx, svc.DB, database.TextArray([]string{notificationMsgID}))
	if err != nil {
		return nil, fmt.Errorf("InfoNotificationMsgRepo.GetByIDs: %v", err)
	}
	if len(es) == 0 {
		return nil, fmt.Errorf("InfoNotificationMsgRepo.GetByIDs: can not find notification message with id %v", notificationMsgID)
	}
	notificationMsg := es[0]

	return notificationMsg, nil
}
