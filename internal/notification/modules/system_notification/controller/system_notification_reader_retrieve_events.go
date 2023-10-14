package controller

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/application/queries"
	systemNotification "github.com/manabie-com/backend/internal/notification/modules/system_notification/util/mapper/systemnotification"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	statusCountings = [...]string{
		npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_NONE.String(),
		npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_NEW.String(),
		npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_DONE.String(),
	}
)

func (svc *SystemNotificationReaderService) RetrieveSystemNotifications(ctx context.Context, req *npb.RetrieveSystemNotificationsRequest) (*npb.RetrieveSystemNotificationsResponse, error) {
	if req.Paging == nil {
		req.Paging = &cpb.Paging{
			Limit:  100,
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 0},
		}
	}

	if req.Paging.Limit == 0 {
		req.Paging.Limit = 100
	}

	userID := interceptors.UserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "svc.SystemNotificationQueryHandler.RetrieveSystemNotifications: user ID is missing")
	}

	payload := &queries.RetrieveSystemNotificationPayload{
		UserID:   userID,
		Limit:    req.Paging.GetLimit(),
		Offset:   req.Paging.GetOffsetInteger(),
		Language: req.Language,
		Status:   req.Status.String(),
		Keyword:  req.Keyword,
	}

	response := svc.SystemNotificationQueryHandler.RetrieveSystemNotifications(ctx, payload)
	if response.Error != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("svc.SystemNotificationQueryHandler.RetrieveSystemNotifications: %v", response.Error))
	}

	systemNotificationsPb := systemNotification.ToSystemNotificationPb(response.SystemNotifications)

	totalForStatusPb := []*npb.RetrieveSystemNotificationsResponse_TotalSystemNotificationForStatus{}
	for _, statusCounting := range statusCountings {
		totalForStatus := &npb.RetrieveSystemNotificationsResponse_TotalSystemNotificationForStatus{
			Status:     npb.SystemNotificationStatus(npb.SystemNotificationStatus_value[statusCounting]),
			TotalItems: 0,
		}
		totalCount, ok := response.TotalForStatus[statusCounting]
		if ok {
			totalForStatus.TotalItems = totalCount
		}
		totalForStatusPb = append(totalForStatusPb, totalForStatus)
	}

	offsetPre := req.Paging.GetOffsetInteger() - int64(req.Paging.Limit)
	if offsetPre < 0 {
		offsetPre = 0
	}

	return &npb.RetrieveSystemNotificationsResponse{
		SystemNotifications: systemNotificationsPb,
		TotalItems:          response.TotalCount,
		TotalItemsForStatus: totalForStatusPb,
		NextPage: &cpb.Paging{
			Limit:  req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: req.Paging.GetOffsetInteger() + int64(len(response.SystemNotifications))},
		},
		PreviousPage: &cpb.Paging{
			Limit:  req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: offsetPre},
		},
	}, nil
}
