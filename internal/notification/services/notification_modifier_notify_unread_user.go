package services

import (
	"context"
	"fmt"

	bobEntities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *NotificationModifierService) NotifyUnreadUser(ctx context.Context, req *npb.NotifyUnreadUserRequest) (*npb.NotifyUnreadUserResponse, error) {
	// Find all unread user of sent notification
	// Use FCM to send notification to user by device token
	userID := interceptors.UserIDFromContext(ctx)
	filter := repositories.NewFindNotificationFilter()

	_ = filter.NotiIDs.Set([]string{req.NotificationId})
	_ = filter.Status.Set([]string{cpb.NotificationStatus_NOTIFICATION_STATUS_SENT.String()})

	es, err := svc.InfoNotificationRepo.Find(ctx, svc.DB, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("InfoNotificationRepo.Find: %v", err))
	}
	if len(es) == 0 {
		return nil, status.Error(codes.Internal, fmt.Sprintf("InfoNotificationRepo.Find: can not find notification with id %v", req.NotificationId))
	}
	noti := es[0]

	notiMsg, err := svc.findNotificationMsg(ctx, noti.NotificationMsgID.String)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("FindNotificationMsg: %v", err))
	}

	var numberOfUser int64 = 1000
	findUserNotiFilter := mappers.UnreadUserNotificationFilter(req.NotificationId, numberOfUser)

	userIDs := make([]string, 0)

	totalSuccess := 0
	totalFailure := 0
	for {
		userNotificationsMap, err := svc.UserNotificationRepo.FindUserIDs(
			ctx,
			svc.DB,
			findUserNotiFilter,
		)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("UserNotificationRepo.Find %v", err))
		}
		userNotifications := userNotificationsMap[req.NotificationId]
		if len(userNotifications) == 0 {
			break
		}
		for _, un := range userNotifications {
			userIDs = append(userIDs, un.UserID.String)
		}

		err = findUserNotiFilter.OffsetText.Set(userNotifications[len(userNotifications)-1].UserID.String)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("can not set offset user id %v", err))
		}
	}

	userIDs = golibs.GetUniqueElementStringArray(userIDs)
	err = svc.UserNotificationRepo.UpdateUnreadUser(ctx, svc.DB, database.Text(req.NotificationId), database.TextArray(userIDs))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("UserNotificationRepo.UpdateUnreadUser %v", err))
	}
	totalSuccess, totalFailure, err = svc.pushNotificationToUsers(ctx, svc.DB, noti, notiMsg, userIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, "error sendNotificationForUsers: "+err.Error())
	}

	activityLogEnt := &bobEntities.ActivityLog{
		ID:         database.Text(idutil.ULIDNow()),
		UserID:     database.Text(userID),
		ActionType: database.Text("notify_user_unread_notification"),
		Payload: database.JSONB(struct {
			NotificationID string `json:"notification_id"`
			SuccessCount   int    `json:"success_count"`
			FailureCount   int    `json:"failure_count"`
		}{
			NotificationID: noti.NotificationID.String,
			SuccessCount:   totalSuccess,
			FailureCount:   totalFailure,
		}),
	}

	err = svc.ActivityLogRepo.Create(ctx, svc.DB, activityLogEnt)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("ActivityLogRepo.Create: %v", err))
	}
	return &npb.NotifyUnreadUserResponse{}, nil
}
