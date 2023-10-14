package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/infrastructure"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/infrastructure/repo"
	systemNotification "github.com/manabie-com/backend/internal/notification/modules/system_notification/util/mapper/systemnotification"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"go.uber.org/multierr"
)

type SystemNotificationQueryHandler struct {
	DB database.Ext

	SystemNotificationRepo        infrastructure.SystemNotificationRepo
	SystemNotificationContentRepo infrastructure.SystemNotificationContentRepo
}

func (query *SystemNotificationQueryHandler) RetrieveSystemNotifications(ctx context.Context, payload *RetrieveSystemNotificationPayload) *RetrieveSystemNotificationResponse {
	response := new(RetrieveSystemNotificationResponse)

	systemNotificationFilter := repo.NewFindSystemNotificationFilter()
	now := time.Now()
	err := multierr.Combine(
		systemNotificationFilter.UserID.Set(payload.UserID),
		systemNotificationFilter.ValidFrom.Set(now), // get only events that are enabled from NOW
		systemNotificationFilter.Limit.Set(payload.Limit),
		systemNotificationFilter.Offset.Set(payload.Offset),
		systemNotificationFilter.Keyword.Set(payload.Keyword),
	)
	if err != nil {
		response.Error = fmt.Errorf("multierr.Combine args params: %v", err)
		return response
	}

	if payload.Language != "" {
		_ = systemNotificationFilter.Language.Set(payload.Language)
	}

	if payload.Status == npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_NONE.String() {
		_ = systemNotificationFilter.Status.Set([]string{npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_DONE.String(), npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_NEW.String()})
	} else {
		_ = systemNotificationFilter.Status.Set([]string{payload.Status})
	}

	systemNotifications, err := query.SystemNotificationRepo.FindSystemNotifications(ctx, query.DB, &systemNotificationFilter)
	if err != nil {
		response.Error = fmt.Errorf("query.SystemNotificationRepo.FindSystemNotifications: %v", err)
		return response
	}

	snIDs := []string{}
	for _, sn := range systemNotifications {
		snIDs = append(snIDs, sn.SystemNotificationID.String)
	}

	systemNotificationContents, err := query.SystemNotificationContentRepo.FindBySystemNotificationIDs(ctx, query.DB, snIDs)
	if err != nil {
		response.Error = fmt.Errorf("query.SystemNotificationContentRepo.FindBySystemNotificationIDs: %+v", err)
	}

	totalForStatus, err := query.SystemNotificationRepo.CountSystemNotifications(ctx, query.DB, &systemNotificationFilter)
	if err != nil {
		response.Error = fmt.Errorf("query.SystemNotificationRepo.CountUserSystemNotifications: %v", err)
		return response
	}

	systemNotificationDTO, err := systemNotification.EntitiesToDTO(systemNotifications, systemNotificationContents)
	if err != nil {
		response.Error = err
		return response
	}
	response.TotalCount = totalForStatus[npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_NONE.String()]
	response.TotalForStatus = totalForStatus
	response.SystemNotifications = systemNotificationDTO

	return response
}
