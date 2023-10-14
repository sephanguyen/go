package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

type mockUserMgmtStudentService struct {
	getStudentProfile       func(ctx context.Context, in *upb.GetStudentProfileRequest, opts ...grpc.CallOption) (*upb.GetStudentProfileResponse, error)
	upsertStudentComment    func(ctx context.Context, in *upb.UpsertStudentCommentRequest, opts ...grpc.CallOption) (*upb.UpsertStudentCommentResponse, error)
	deleteStudentComments   func(ctx context.Context, in *upb.DeleteStudentCommentsRequest, opts ...grpc.CallOption) (*upb.DeleteStudentCommentsResponse, error)
	retrieveStudentComments func(ctx context.Context, in *upb.RetrieveStudentCommentRequest, opts ...grpc.CallOption) (*upb.RetrieveStudentCommentResponse, error)
}

func (m *mockUserMgmtStudentService) GetStudentProfile(ctx context.Context, in *upb.GetStudentProfileRequest, opts ...grpc.CallOption) (*upb.GetStudentProfileResponse, error) {
	return m.getStudentProfile(ctx, in, opts...)
}

func (m *mockUserMgmtStudentService) UpsertStudentComment(ctx context.Context, in *upb.UpsertStudentCommentRequest, opts ...grpc.CallOption) (*upb.UpsertStudentCommentResponse, error) {
	return m.upsertStudentComment(ctx, in, opts...)
}

func (m *mockUserMgmtStudentService) DeleteStudentComments(ctx context.Context, in *upb.DeleteStudentCommentsRequest, opts ...grpc.CallOption) (*upb.DeleteStudentCommentsResponse, error) {
	return m.deleteStudentComments(ctx, in, opts...)
}

func (m *mockUserMgmtStudentService) RetrieveStudentComment(ctx context.Context, in *upb.RetrieveStudentCommentRequest, opts ...grpc.CallOption) (*upb.RetrieveStudentCommentResponse, error) {
	return m.retrieveStudentComments(ctx, in, opts...)
}

func TestStudentModifier_DeleteStudentComments(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	s := &StudentModifierServices{
		DB: db,
	}

	testCases := []TestCase{
		{
			name:        "happy case",
			ctx:         interceptors.ContextWithUserID(ctx, "id"),
			req:         &bpb.DeleteStudentCommentsRequest{CommentIds: []string{"cmt-1", "cmt-2"}},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				s.UserMgmtStudentService = &mockUserMgmtStudentService{
					deleteStudentComments: func(ctx context.Context, in *upb.DeleteStudentCommentsRequest, opts ...grpc.CallOption) (*upb.DeleteStudentCommentsResponse, error) {
						return &upb.DeleteStudentCommentsResponse{
							Successful: true,
						}, nil
					},
				}
			},
		},
		{
			name:        "error from usermgmt service(something like internal error,...)",
			ctx:         interceptors.ContextWithUserID(ctx, "id"),
			req:         &bpb.DeleteStudentCommentsRequest{CommentIds: []string{"cmt-1", "cmt-2"}},
			expectedErr: fmt.Errorf("rpc error: code = INTERNAL"),
			setup: func(ctx context.Context) {
				s.UserMgmtStudentService = &mockUserMgmtStudentService{
					deleteStudentComments: func(ctx context.Context, in *upb.DeleteStudentCommentsRequest, opts ...grpc.CallOption) (*upb.DeleteStudentCommentsResponse, error) {
						return nil, fmt.Errorf("rpc error: code = INTERNAL")
					},
				}
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*bpb.DeleteStudentCommentsRequest)
			_, err := s.DeleteStudentComments(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}
