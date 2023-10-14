package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *NotificationReaderService) RetrieveDraftAudience(ctx context.Context, req *npb.RetrieveDraftAudienceRequest) (*npb.RetrieveDraftAudienceResponse, error) {
	if req.Paging == nil {
		req.Paging = &cpb.Paging{
			Limit:  100,
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 0},
		}
	}

	if req.Paging.Limit == 0 {
		req.Paging.Limit = 100
	}

	notification, err := svc.findDraftOrScheduledNotification(ctx, req.NotificationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("error RetrieveDraftAudience.findDraftOrScheduledNotification %v", err))
	}

	org, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error RetrieveDraftAudience - interceptors.OrganizationFromContext %v", err))
	}
	editorContext := interceptors.ContextWithJWTClaims(context.Background(), &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: org.OrganizationID().String(),
			UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
			UserID:       notification.EditorID.String,
		},
	})

	targetGroup := &entities.InfoNotificationTarget{}
	err = notification.TargetGroups.AssignTo(targetGroup)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("cannot set target group entity to process: %v", err))
	}
	audiences, total, err := svc.NotificationAudienceRetriever.FindDraftAudiencesWithPaging(editorContext, svc.DB, notification.NotificationID.String, targetGroup, database.FromTextArray(notification.GenericReceiverIDs), database.FromTextArray(notification.ExcludedGenericReceiverIDs), int(req.Paging.GetLimit()), int(req.Paging.GetOffsetInteger()))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error NotificationAudienceRetriever.FindDraftAudiencesWithPaging %v", err))
	}

	offsetPre := req.Paging.GetOffsetInteger() - int64(req.Paging.Limit)
	if offsetPre < 0 {
		offsetPre = 0
	}

	response := &npb.RetrieveDraftAudienceResponse{
		Audiences:  mappers.NotificationDraftAudiencesToPb(audiences),
		TotalItems: total,
		NextPage: &cpb.Paging{
			Limit:  req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: req.Paging.GetOffsetInteger() + int64(len(audiences))},
		},
		PreviousPage: &cpb.Paging{
			Limit:  req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: offsetPre},
		},
	}
	return response, nil
}

func (svc *NotificationReaderService) findDraftOrScheduledNotification(ctx context.Context, notificationID string) (*entities.InfoNotification, error) {
	filter := repositories.NewFindNotificationFilter()
	err := filter.NotiIDs.Set([]string{notificationID})
	if err != nil {
		return nil, fmt.Errorf("cannot set notification id for FindNotificationFilter")
	}
	err = filter.Status.Set([]string{cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT.String(), cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String()})
	if err != nil {
		return nil, fmt.Errorf("cannot set notification status for FindNotificationFilter")
	}

	es, err := svc.InfoNotificationRepo.Find(ctx, svc.DB, filter)
	if err != nil {
		return nil, fmt.Errorf("InfoNotificationRepo.Find: %v", err)
	}
	if len(es) == 0 {
		return nil, fmt.Errorf("InfoNotificationRepo.Find: can not find notification with id %v, or your notification is sent", notificationID)
	}
	noti := es[0]

	return noti, nil
}
