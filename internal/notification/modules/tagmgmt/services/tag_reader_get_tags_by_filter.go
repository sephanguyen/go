package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/notification/modules/tagmgmt/repositories"
	"github.com/manabie-com/backend/internal/notification/modules/tagmgmt/services/mappers"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Deprecated: We don't use this API, but still keep because product requirements might change in the future
func (rcv *TagMgmtReaderService) GetTagsByFilter(ctx context.Context, req *npb.GetTagsByFilterRequest) (*npb.GetTagsByFilterResponse, error) {
	if req.Paging == nil {
		req.Paging = &cpb.Paging{
			Limit:  100,
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 0},
		}
	}

	if req.Paging.Limit == 0 {
		req.Paging.Limit = 100
	}

	tagFilter := repositories.NewFindTagFilter()
	_ = tagFilter.Keyword.Set(req.Keyword)
	_ = tagFilter.Offset.Set(req.Paging.GetOffsetInteger())
	_ = tagFilter.Limit.Set(req.Paging.GetLimit())
	_ = tagFilter.WithCount.Set(true)

	tags, totalResult, err := rcv.TagRepo.FindByFilter(ctx, rcv.DB, tagFilter)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("GetTagsByFilter: %v", err))
	}

	result := mappers.TagEntitiesToTagsByFilterResponse(tags)

	offsetPre := req.Paging.GetOffsetInteger() - int64(req.Paging.Limit)
	if offsetPre < 0 {
		offsetPre = 0
	}

	return &npb.GetTagsByFilterResponse{
		Tags:       result,
		TotalItems: totalResult,
		NextPage: &cpb.Paging{
			Limit:  req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: req.Paging.GetOffsetInteger() + int64(len(result))},
		},
		PreviousPage: &cpb.Paging{
			Limit:  req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: offsetPre},
		},
	}, nil
}
