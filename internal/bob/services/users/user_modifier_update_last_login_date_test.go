package users

import (
	"context"
	"fmt"
	"testing"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestUpdateUserLastLoginDate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userRepo := new(mock_repositories.MockUserRepo)

	s := &UserModifierService{
		UserRepo: userRepo,
	}

	now := time.Now().UTC()
	student := &entities_bob.User{}
	_ = student.ID.Set("id")
	_ = student.LastLoginDate.Set(now)

	testCases := []testcase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, student.ID.String),
			req:  &pb.UpdateUserLastLoginDateRequest{LastLoginDate: timestamppb.New(now)},
			expectedResp: &pb.UpdateUserLastLoginDateResponse{
				Successful: true,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				userRepo.On("UpdateLastLoginDate", ctx, mock.Anything, student).Once().Return(nil)
			},
		},
		{
			name:         "invalid request",
			ctx:          interceptors.ContextWithUserID(ctx, student.ID.String),
			req:          nil,
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid last login date request"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:         "invalid request last login date value",
			ctx:          interceptors.ContextWithUserID(ctx, student.ID.String),
			req:          &pb.UpdateUserLastLoginDateRequest{LastLoginDate: timestamppb.New(time.Time{})},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid last login date value"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:         "invalid request last login date value",
			ctx:          interceptors.ContextWithUserID(ctx, student.ID.String),
			req:          &pb.UpdateUserLastLoginDateRequest{},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid last login date value"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:         "failed due to UpdateLastLoginDate return error",
			ctx:          interceptors.ContextWithUserID(ctx, student.ID.String),
			req:          &pb.UpdateUserLastLoginDateRequest{LastLoginDate: timestamppb.New(now)},
			expectedResp: nil,
			expectedErr:  fmt.Errorf("mock error"),
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
			resp, err := s.UpdateUserLastLoginDate(testCase.ctx, req)

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
				assert.Nil(t, nil, resp)
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}
