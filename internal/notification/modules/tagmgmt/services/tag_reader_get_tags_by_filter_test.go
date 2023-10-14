package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/modules/tagmgmt/entities"
	mock_repositories "github.com/manabie-com/backend/mock/notification/modules/tagmgmt/repositories"
	"github.com/manabie-com/backend/mock/testutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_GetTagsByFilter(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	tagRepo := &mock_repositories.MockTagRepo{}
	svc := &TagMgmtReaderService{
		DB:      mockDB.DB,
		TagRepo: tagRepo,
	}
	tag1 := &entities.Tag{}
	tag2 := &entities.Tag{}
	database.AllRandomEntity(tag1)
	database.AllRandomEntity(tag2)
	tagResult := []*npb.GetTagsByFilterResponse_Tag{
		{
			TagId: tag1.TagID.String,
			Name:  tag1.TagName.String,
		},
		{
			TagId: tag2.TagID.String,
			Name:  tag2.TagName.String,
		},
	}
	limit := 10
	offset := 0
	testCases := []struct {
		Name         string
		Request      interface{}
		ExpcResponse interface{}
		ExpcErr      error
		Setup        func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Request: &npb.GetTagsByFilterRequest{
				Keyword: "manabie-tag-1",
				Paging: &cpb.Paging{
					Limit:  uint32(limit),
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: int64(offset)},
				},
			},
			ExpcResponse: &npb.GetTagsByFilterResponse{
				Tags:       tagResult,
				TotalItems: uint32(len(tagResult)),
				NextPage: &cpb.Paging{
					Limit:  uint32(limit),
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: int64(offset) + int64(len(tagResult))},
				},
				PreviousPage: &cpb.Paging{
					Limit:  uint32(limit),
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: int64(0)},
				},
			},
			ExpcErr: nil,
			Setup: func(ctx context.Context) {
				tags := entities.Tags{}
				tags = append(tags, tag1)
				tags = append(tags, tag2)
				tagRepo.On("FindByFilter", ctx, mockDB.DB, mock.Anything).Once().Return(
					tags,
					uint32(len(tagResult)),
					nil,
				)
			},
		},
		{
			Name: "err find",
			Request: &npb.GetTagsByFilterRequest{
				Keyword: "manabie-tag-1",
				Paging: &cpb.Paging{
					Limit:  uint32(limit),
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: int64(offset)},
				},
			},
			ExpcResponse: &npb.GetTagsByFilterResponse{},
			ExpcErr:      status.Error(codes.Internal, fmt.Sprintf("GetTagsByFilter: %v", pgx.ErrNoRows)),
			Setup: func(ctx context.Context) {
				tags := entities.Tags{}
				tagRepo.On("FindByFilter", ctx, mockDB.DB, mock.Anything).Once().Return(
					tags,
					uint32(0),
					pgx.ErrNoRows,
				)
			},
		},
		{
			Name: "paging missing",
			Request: &npb.GetTagsByFilterRequest{
				Keyword: "manabie-tag-1",
				Paging:  nil,
			},
			ExpcResponse: &npb.GetTagsByFilterResponse{
				Tags:       tagResult,
				TotalItems: uint32(len(tagResult)),
				NextPage: &cpb.Paging{
					Limit:  uint32(100),
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: int64(offset) + int64(len(tagResult))},
				},
				PreviousPage: &cpb.Paging{
					Limit:  uint32(100),
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: int64(0)},
				},
			},
			ExpcErr: nil,
			Setup: func(ctx context.Context) {
				tags := entities.Tags{}
				tags = append(tags, tag1)
				tags = append(tags, tag2)
				tagRepo.On("FindByFilter", ctx, mockDB.DB, mock.Anything).Once().Return(
					tags,
					uint32(len(tagResult)),
					nil,
				)
			},
		},
	}

	ctx := context.Background()

	for _, testCase := range testCases {
		testCase.Setup(ctx)
		t.Run(testCase.Name, func(t *testing.T) {
			res, err := svc.GetTagsByFilter(ctx, testCase.Request.(*npb.GetTagsByFilterRequest))
			if testCase.ExpcErr == nil {
				assert.Equal(t, testCase.ExpcResponse.(*npb.GetTagsByFilterResponse), res)
			} else {
				assert.Equal(t, testCase.ExpcErr.Error(), err.Error())
			}
		})
	}
}
