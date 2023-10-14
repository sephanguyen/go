package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/notification/services/mappers"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *NotificationReaderService) RetrieveGroupAudience(ctx context.Context, req *npb.RetrieveGroupAudienceRequest) (*npb.RetrieveGroupAudienceResponse, error) {
	if req.Paging == nil {
		req.Paging = &cpb.Paging{
			Limit:  100,
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 0},
		}
	}

	if req.Paging.Limit == 0 {
		req.Paging.Limit = 100
	}

	audiences, total, err := svc.NotificationAudienceRetriever.FindGroupAudiencesWithPaging(ctx, svc.DB, "",
		mappers.PbToNotificationTargetEnt(req.TargetGroup), req.Keyword, req.GetUserIds(), int(req.Paging.GetLimit()), int(req.Paging.GetOffsetInteger()))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("NotificationAudienceRetriever.FindUserIDWithPaging %v", err))
	}

	offsetPre := req.Paging.GetOffsetInteger() - int64(req.Paging.Limit)

	if offsetPre < 0 {
		offsetPre = 0
	}

	response := &npb.RetrieveGroupAudienceResponse{
		Audiences:  mappers.NotificationGroupAudiencesToPb(audiences),
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
