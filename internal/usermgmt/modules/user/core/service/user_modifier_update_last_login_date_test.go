package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_repos "github.com/manabie-com/backend/mock/usermgmt/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestUpdateUserLastLoginDate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userRepo := new(mock_repos.MockUserRepo)
	userModifierService := &UserModifierService{
		UserRepo: userRepo,
	}

	now := time.Now().UTC()
	student := &entity.LegacyUser{}
	err := multierr.Combine(
		student.ID.Set("id"),
		student.LastLoginDate.Set(now),
	)
	assert.NoError(t, err)

	testCases := []TestCase{
		{
			name:        "happy case",
			ctx:         interceptors.ContextWithUserID(ctx, student.ID.String),
			req:         &pb.UpdateUserLastLoginDateRequest{LastLoginDate: timestamppb.New(now)},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				userRepo.On("UpdateLastLoginDate", ctx, mock.Anything, student).Once().Return(nil)
			},
		},
		{
			name:        "invalid request",
			ctx:         interceptors.ContextWithUserID(ctx, student.ID.String),
			req:         nil,
			expectedErr: status.Error(codes.InvalidArgument, "invalid last login date request"),
			setup:       func(ctx context.Context) {},
		},
		{
			name:        "invalid request last login date value",
			ctx:         interceptors.ContextWithUserID(ctx, student.ID.String),
			req:         &pb.UpdateUserLastLoginDateRequest{LastLoginDate: timestamppb.New(time.Time{})},
			expectedErr: status.Error(codes.InvalidArgument, "invalid last login date value"),
			setup:       func(ctx context.Context) {},
		},
		{
			name:        "invalid request last login date value",
			ctx:         interceptors.ContextWithUserID(ctx, student.ID.String),
			req:         &pb.UpdateUserLastLoginDateRequest{},
			expectedErr: status.Error(codes.InvalidArgument, "invalid last login date value"),
			setup:       func(ctx context.Context) {},
		},
		{
			name:        "failed due to UpdateLastLoginDate return error",
			ctx:         interceptors.ContextWithUserID(ctx, student.ID.String),
			req:         &pb.UpdateUserLastLoginDateRequest{LastLoginDate: timestamppb.New(now)},
			expectedErr: fmt.Errorf("failed to update user last login date: mock error"),
			setup: func(ctx context.Context) {
				userRepo.On("UpdateLastLoginDate", ctx, mock.Anything, student).Once().Return(fmt.Errorf("mock error"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req, ok := testCase.req.(*pb.UpdateUserLastLoginDateRequest)
			if !ok {
				assert.Nil(t, nil, req)
			}

			resp, err := userModifierService.UpdateUserLastLoginDate(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
				assert.Nil(t, nil, resp)
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}
