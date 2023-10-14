package services

import (
	"context"
	"testing"

	"github.com/jackc/puddle"
	mock_repositories "github.com/manabie-com/backend/mock/notification/modules/tagmgmt/repositories"
	"github.com/manabie-com/backend/mock/testutil"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_UpsertTag(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	tagRepo := &mock_repositories.MockTagRepo{}
	svc := &TagMgmtModifierService{
		DB:      mockDB.DB,
		TagRepo: tagRepo,
	}

	testCases := []struct {
		Name         string
		Request      interface{}
		ExpcResponse interface{}
		ExpcErr      error
		Setup        func(ctx context.Context)
	}{
		{
			Name:         "error upsert",
			Request:      &npb.UpsertTagRequest{TagId: "tag-id-1", Name: "manabie-tag-1"},
			ExpcResponse: nil,
			ExpcErr:      status.Error(codes.Internal, puddle.ErrClosedPool.Error()),
			Setup: func(ctx context.Context) {
				tagRepo.On("DoesTagNameExist", ctx, mockDB.DB, mock.Anything).Once().Return(false, nil)
				tagRepo.On("Upsert", ctx, mockDB.DB, mock.Anything).Once().Return(puddle.ErrClosedPool)
			},
		},
		{
			Name:         "error DoesTagNameExist throw error",
			Request:      &npb.UpsertTagRequest{TagId: "tag-id-1", Name: "manabie-tag-1"},
			ExpcResponse: nil,
			ExpcErr:      status.Error(codes.Internal, puddle.ErrClosedPool.Error()),
			Setup: func(ctx context.Context) {
				err := puddle.ErrClosedPool
				tagRepo.On("DoesTagNameExist", ctx, mockDB.DB, mock.Anything).Once().Return(false, err)
				tagRepo.On("Upsert", ctx, mockDB.DB, mock.Anything).Once().Return(nil)
			},
		},
		{
			Name:         "happy case insert/upsert",
			Request:      &npb.UpsertTagRequest{TagId: "tag-id-1", Name: "manabie-tag-1"},
			ExpcResponse: &npb.UpsertTagResponse{TagId: "tag-id-1"},
			ExpcErr:      nil,
			Setup: func(ctx context.Context) {
				tagRepo.On("DoesTagNameExist", ctx, mockDB.DB, mock.Anything).Once().Return(false, nil)
				tagRepo.On("Upsert", ctx, mockDB.DB, mock.Anything).Once().Return(nil)
			},
		},
		{
			Name:         "error tag name exist",
			Request:      &npb.UpsertTagRequest{TagId: "tag-id-1", Name: "manabie-tag-1"},
			ExpcResponse: nil,
			ExpcErr:      status.Error(codes.InvalidArgument, "TagName is exist"),
			Setup: func(ctx context.Context) {
				tagRepo.On("DoesTagNameExist", ctx, mockDB.DB, mock.Anything).Once().Return(true, nil)
				tagRepo.On("Upsert", ctx, mockDB.DB, mock.Anything).Once().Return(nil)
			},
		},
		{
			Name:         "error empty tag name",
			Request:      &npb.UpsertTagRequest{TagId: "tag-id-1"},
			ExpcResponse: nil,
			ExpcErr:      status.Error(codes.InvalidArgument, "TagName is empty"),
			Setup: func(ctx context.Context) {
				tagRepo.On("DoesTagNameExist", ctx, mockDB.DB, mock.Anything).Once().Return(false, nil)
				tagRepo.On("Upsert", ctx, mockDB.DB, mock.Anything).Once().Return(nil)
			},
		},
	}

	ctx := context.Background()
	for _, testCase := range testCases {
		testCase.Setup(ctx)
		t.Run(testCase.Name, func(t *testing.T) {
			res, err := svc.UpsertTag(ctx, testCase.Request.(*npb.UpsertTagRequest))
			if testCase.ExpcErr == nil {
				assert.Equal(t, testCase.ExpcResponse.(*npb.UpsertTagResponse).TagId, res.TagId)
			} else {
				assert.Equal(t, testCase.ExpcErr.Error(), err.Error())
			}
		})
	}
}
