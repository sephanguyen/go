package users

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	pb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type testcase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}
type mockUserMgmtModifierService struct {
	updateUserProfile func(ctx context.Context, in *upb.UpdateUserProfileRequest, opts ...grpc.CallOption) (*upb.UpdateUserProfileResponse, error)
}

func (m *mockUserMgmtModifierService) UpdateUserProfile(ctx context.Context, in *upb.UpdateUserProfileRequest, opts ...grpc.CallOption) (*upb.UpdateUserProfileResponse, error) {
	return m.updateUserProfile(ctx, in, opts...)
}

func TestUpdateUserProfile(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userRepo := new(mock_repositories.MockUserRepo)
	schoolAdminRepo := new(mock_repositories.MockSchoolAdminRepo)
	teacherRepo := new(mock_repositories.MockTeacherRepo)
	jsm := new(mock_nats.JetStreamManagement)
	pAdminWithoutName := generateUserProfileWithoutField(entities.UserGroupAdmin, "FullName")
	pAdminWithoutUserGroup := generateUserProfileWithoutField(entities.UserGroupAdmin, "UserGroup")
	pSchoolAdmin := generateUserProfile(entities.UserGroupSchoolAdmin)
	pTeacher := generateUserProfile(entities.UserGroupTeacher)
	s := &UserModifierService{
		UserRepo:        userRepo,
		SchoolAdminRepo: schoolAdminRepo,
		TeacherRepo:     teacherRepo,
		JSM:             jsm,
	}

	userId := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userId)

	pAdmin := generateUserProfile(entities_bob.UserGroupAdmin)

	testCases := []testcase{
		{
			name:         "err: usermgmt conn failed",
			ctx:          interceptors.ContextWithUserID(ctx, pAdmin.Id),
			req:          &pb.UpdateUserProfileRequest{Profile: pAdmin},
			expectedResp: nil,
			expectedErr:  errors.New("s.UserMgmtModifierSvc.UpdateUserProfile: usermgmt conn failed"),
			setup: func(ctx context.Context) {
				s.UserMgmtModifierSvc = &mockUserMgmtModifierService{
					updateUserProfile: func(ctx context.Context, in *upb.UpdateUserProfileRequest, opts ...grpc.CallOption) (*upb.UpdateUserProfileResponse, error) {
						return nil, fmt.Errorf("s.UserMgmtModifierSvc.UpdateUserProfile: usermgmt conn failed")
					},
				}
			},
		},
		{
			name: "happy case success",
			ctx:  interceptors.ContextWithUserID(ctx, pAdmin.Id),
			req:  &pb.UpdateUserProfileRequest{Profile: pAdmin},
			expectedResp: &pb.UpdateUserProfileResponse{
				Successful: true,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				s.UserMgmtModifierSvc = &mockUserMgmtModifierService{
					updateUserProfile: func(ctx context.Context, in *upb.UpdateUserProfileRequest, opts ...grpc.CallOption) (*upb.UpdateUserProfileResponse, error) {
						return &upb.UpdateUserProfileResponse{Successful: true}, nil
					},
				}
			},
		},
		{
			name:         "update profile which does not have user name",
			ctx:          interceptors.ContextWithUserID(ctx, pAdmin.Id),
			req:          &pb.UpdateUserProfileRequest{Profile: pAdminWithoutName},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid profile"),
			setup: func(ctx context.Context) {
				s.UserMgmtModifierSvc = &mockUserMgmtModifierService{
					updateUserProfile: func(ctx context.Context, in *upb.UpdateUserProfileRequest, opts ...grpc.CallOption) (*upb.UpdateUserProfileResponse, error) {
						return nil, status.Error(codes.InvalidArgument, "invalid profile")
					},
				}
			},
		},
		{
			name:         "update profile which does not have user group",
			ctx:          interceptors.ContextWithUserID(ctx, pAdminWithoutUserGroup.Id),
			req:          &pb.UpdateUserProfileRequest{Profile: pAdminWithoutUserGroup},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid profile"),
			setup: func(ctx context.Context) {
				s.UserMgmtModifierSvc = &mockUserMgmtModifierService{
					updateUserProfile: func(ctx context.Context, in *upb.UpdateUserProfileRequest, opts ...grpc.CallOption) (*upb.UpdateUserProfileResponse, error) {
						return nil, status.Error(codes.InvalidArgument, "invalid profile")
					},
				}
			},
		},
		{
			name:         "update profile fail due to db userRepo UpdateProfile tx fail",
			ctx:          interceptors.ContextWithUserID(ctx, pAdmin.Id),
			req:          &pb.UpdateUserProfileRequest{Profile: pAdmin},
			expectedResp: nil,
			expectedErr:  errors.New("rpc error: code = Unknown desc = cannot update profile"),
			setup: func(ctx context.Context) {
				s.UserMgmtModifierSvc = &mockUserMgmtModifierService{
					updateUserProfile: func(ctx context.Context, in *upb.UpdateUserProfileRequest, opts ...grpc.CallOption) (*upb.UpdateUserProfileResponse, error) {
						return nil, errors.New("rpc error: code = Unknown desc = cannot update profile")
					},
				}
			},
		},
		{
			name:         "school admin update teacher profile fail due to db schoolAdminRepo Get tx fail",
			ctx:          interceptors.ContextWithUserID(ctx, pSchoolAdmin.Id),
			req:          &pb.UpdateUserProfileRequest{Profile: pTeacher},
			expectedResp: nil,
			expectedErr:  errors.Wrapf(pgx.ErrNoRows, "s.SchoolAdminRepo.Get: userID: %q", pSchoolAdmin.Id),
			setup: func(ctx context.Context) {
				s.UserMgmtModifierSvc = &mockUserMgmtModifierService{
					updateUserProfile: func(ctx context.Context, in *upb.UpdateUserProfileRequest, opts ...grpc.CallOption) (*upb.UpdateUserProfileResponse, error) {
						return nil, errors.Wrapf(pgx.ErrNoRows, "s.SchoolAdminRepo.Get: userID: %q", pSchoolAdmin.Id)
					},
				}
			},
		},
		{
			name:         "school admin update teacher profile fail due to db schoolAdminRepo Get tx return nil",
			ctx:          interceptors.ContextWithUserID(ctx, pSchoolAdmin.Id),
			req:          &pb.UpdateUserProfileRequest{Profile: pTeacher},
			expectedResp: nil,
			expectedErr:  status.Error(codes.PermissionDenied, "only school admin can update their teacher profile"),
			setup: func(ctx context.Context) {
				s.UserMgmtModifierSvc = &mockUserMgmtModifierService{
					updateUserProfile: func(ctx context.Context, in *upb.UpdateUserProfileRequest, opts ...grpc.CallOption) (*upb.UpdateUserProfileResponse, error) {
						return nil, status.Error(codes.PermissionDenied, "only school admin can update their teacher profile")
					},
				}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.UpdateUserProfileRequest)
			resp, err := s.UpdateUserProfile(testCase.ctx, req)
			if testCase.expectedErr == nil {
				assert.NoError(t, err, "expecting no error")
			} else {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error(), "unexpected error message")
			}
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

func generateUserProfile(userGroup string) *pb.UserProfile {
	rand.Seed(time.Now().UnixNano())
	return &pb.UserProfile{
		Id:          fmt.Sprintf("%d", rand.Int()),
		Name:        fmt.Sprintf("user %d", rand.Int()),
		Country:     cpb.Country_COUNTRY_VN,
		PhoneNumber: fmt.Sprintf("+849%d", rand.Int()),
		Email:       fmt.Sprintf("valid-%d@email.com", rand.Int()),
		Avatar:      fmt.Sprintf("http://avatar-%d", rand.Int()),
		DeviceToken: fmt.Sprintf("random device %d", rand.Int()),
		UserGroup:   userGroup,
		CreatedAt:   &timestamppb.Timestamp{Seconds: time.Now().Unix()},
		UpdatedAt:   &timestamppb.Timestamp{Seconds: time.Now().Unix()},
	}
}

func generateUserProfileWithoutField(userGroup, missingFieldName string) *pb.UserProfile {
	rand.Seed(time.Now().UnixNano())
	p := &pb.UserProfile{
		Id:          fmt.Sprintf("%d", rand.Int()),
		Name:        fmt.Sprintf("user %d", rand.Int()),
		Country:     cpb.Country_COUNTRY_VN,
		PhoneNumber: fmt.Sprintf("+849%d", rand.Int()),
		Email:       fmt.Sprintf("valid-%d@email.com", rand.Int()),
		Avatar:      fmt.Sprintf("http://avatar-%d", rand.Int()),
		DeviceToken: fmt.Sprintf("random device %d", rand.Int()),
		UserGroup:   userGroup,
		CreatedAt:   &timestamppb.Timestamp{Seconds: time.Now().Unix()},
		UpdatedAt:   &timestamppb.Timestamp{Seconds: time.Now().Unix()},
	}
	switch missingFieldName {
	case "Id":
		p.Id = ""
	case "PhoneNumber":
		p.PhoneNumber = ""
	case "FullName":
		p.Name = ""
	case "UserGroup":
		p.UserGroup = ""
	}
	return p
}
