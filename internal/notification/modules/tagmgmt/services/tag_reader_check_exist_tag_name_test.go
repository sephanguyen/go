package services

import (
	"context"
	"errors"
	"testing"

	mock_repositories "github.com/manabie-com/backend/mock/notification/modules/tagmgmt/repositories"
	"github.com/manabie-com/backend/mock/testutil"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_CheckExistTagName(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	tagRepo := &mock_repositories.MockTagRepo{}
	svc := &TagMgmtReaderService{
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
			Name:         "tag name exist",
			Request:      &npb.CheckExistTagNameRequest{TagName: "manabie-tag-1"},
			ExpcResponse: &npb.CheckExistTagNameResponse{IsExist: true},
			ExpcErr:      nil,
			Setup: func(ctx context.Context) {
				tagRepo.On("DoesTagNameExist", ctx, mockDB.DB, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			Name:         "tag name not exist",
			Request:      &npb.CheckExistTagNameRequest{TagName: "manabie-tag-1"},
			ExpcResponse: &npb.CheckExistTagNameResponse{IsExist: false},
			ExpcErr:      nil,
			Setup: func(ctx context.Context) {
				tagRepo.On("DoesTagNameExist", ctx, mockDB.DB, mock.Anything).Once().Return(false, nil)
			},
		},
		{
			Name:         "err DoesTagNameExist",
			Request:      &npb.CheckExistTagNameRequest{TagName: "manabie-tag-1"},
			ExpcResponse: &npb.CheckExistTagNameResponse{IsExist: true},
			ExpcErr:      status.Error(codes.Internal, "CheckExistTagName: some error"),
			Setup: func(ctx context.Context) {
				tagRepo.On("DoesTagNameExist", ctx, mockDB.DB, mock.Anything).Once().Return(false, errors.New("some error"))
			},
		},
	}

	ctx := context.Background()

	for _, testCase := range testCases {
		testCase.Setup(ctx)
		t.Run(testCase.Name, func(t *testing.T) {
			res, err := svc.CheckExistTagName(ctx, testCase.Request.(*npb.CheckExistTagNameRequest))
			if testCase.ExpcErr == nil {
				assert.Equal(t, testCase.ExpcResponse.(*npb.CheckExistTagNameResponse).IsExist, res.IsExist)
			} else {
				assert.Equal(t, testCase.ExpcErr.Error(), err.Error())
			}
		})
	}
}
