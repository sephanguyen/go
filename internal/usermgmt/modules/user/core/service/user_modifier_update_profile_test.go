package service

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type testcase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestUpdateUserProfile(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userRepo := new(mock_repositories.MockUserRepo)
	schoolAdminRepo := new(mock_repositories.MockSchoolAdminRepo)
	teacherRepo := new(mock_repositories.MockTeacherRepo)
	jsm := new(mock_nats.JetStreamManagement)

	s := &UserModifierService{
		UserRepo:        userRepo,
		SchoolAdminRepo: schoolAdminRepo,
		TeacherRepo:     teacherRepo,
		JSM:             jsm,
	}

	pStudent := generateUserProfile(entity.UserGroupStudent)
	pStudentWithoutName := generateUserProfileWithoutField(entity.UserGroupStudent, "Name")
	pStudentWithoutPhoneNum := generateUserProfileWithoutField(entity.UserGroupStudent, "PhoneNumber")
	pParent := generateUserProfile(entity.UserGroupParent)
	pParentWithoutPhoneNum := generateUserProfileWithoutField(entity.UserGroupParent, "PhoneNumber")
	pSchoolAdmin := generateUserProfile(entity.UserGroupSchoolAdmin)
	pTeacher := generateUserProfile(entity.UserGroupTeacher)

	testCases := []testcase{
		{
			name: "happy case student",
			ctx:  interceptors.ContextWithUserID(ctx, pStudent.Id),
			req:  &pb.UpdateUserProfileRequest{Profile: pStudent},
			expectedResp: &pb.UpdateUserProfileResponse{
				Successful: true,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, mock.Anything, mock.Anything).Once().Return(&entity.LegacyUser{Group: database.Text(constant.UserGroupStudent)}, nil)
				userRepo.On("UpdateProfileV1", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectUserDeviceTokenUpdated, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name: "happy case student (profile to update has no phone number)",
			ctx:  interceptors.ContextWithUserID(ctx, pStudentWithoutPhoneNum.Id),
			req:  &pb.UpdateUserProfileRequest{Profile: pStudentWithoutPhoneNum},
			expectedResp: &pb.UpdateUserProfileResponse{
				Successful: true,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, mock.Anything, mock.Anything).Once().Return(&entity.LegacyUser{Group: database.Text(constant.UserGroupStudent)}, nil)
				userRepo.On("UpdateProfileV1", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectUserDeviceTokenUpdated, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name: "happy case parent",
			ctx:  interceptors.ContextWithUserID(ctx, pParent.Id),
			req:  &pb.UpdateUserProfileRequest{Profile: pParent},
			expectedResp: &pb.UpdateUserProfileResponse{
				Successful: true,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, mock.Anything, mock.Anything).Once().Return(&entity.LegacyUser{Group: database.Text(constant.UserGroupParent)}, nil)
				userRepo.On("UpdateProfileV1", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectUserDeviceTokenUpdated, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name: "happy case parent (profile to update has no phone number)",
			ctx:  interceptors.ContextWithUserID(ctx, pParentWithoutPhoneNum.Id),
			req:  &pb.UpdateUserProfileRequest{Profile: pParentWithoutPhoneNum},
			expectedResp: &pb.UpdateUserProfileResponse{
				Successful: true,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, mock.Anything, mock.Anything).Once().Return(&entity.LegacyUser{Group: database.Text(constant.UserGroupParent)}, nil)
				userRepo.On("UpdateProfileV1", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectUserDeviceTokenUpdated, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:        "student update other profile",
			ctx:         interceptors.ContextWithUserID(ctx, "studentId"),
			req:         &pb.UpdateUserProfileRequest{Profile: generateUserProfile(entity.UserGroupStudent)},
			expectedErr: status.Error(codes.PermissionDenied, "user can only update own profile"),
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, mock.Anything, mock.Anything).Once().Return(&entity.LegacyUser{Group: database.Text(constant.UserGroupStudent)}, nil)
				userRepo.On("UpdateProfileV1", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectUserDeviceTokenUpdated, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:         "update profile which does not have user name",
			ctx:          interceptors.ContextWithUserID(ctx, pStudentWithoutName.Id),
			req:          &pb.UpdateUserProfileRequest{Profile: pStudentWithoutName},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid profile"),
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, mock.Anything, mock.Anything).Once().Return(&entity.LegacyUser{Group: database.Text(constant.UserGroupParent)}, nil)
				userRepo.On("UpdateProfileV1", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectUserDeviceTokenUpdated, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:         "update profile fail due to db userRepo UpdateProfile tx fail",
			ctx:          interceptors.ContextWithUserID(ctx, pStudent.Id),
			req:          &pb.UpdateUserProfileRequest{Profile: pStudent},
			expectedResp: nil,
			expectedErr:  errors.New("rpc error: code = Unknown desc = cannot update profile"),
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, mock.Anything, mock.Anything).Once().Return(&entity.LegacyUser{Group: database.Text(constant.UserGroupStudent)}, nil)
				userRepo.On("UpdateProfileV1", ctx, mock.Anything, mock.Anything).Once().Return(errors.New("rpc error: code = Unknown desc = cannot update profile"))
			},
		},
		{
			name: "staff update another update staff",
			ctx:  interceptors.ContextWithUserID(ctx, pSchoolAdmin.Id),
			req:  &pb.UpdateUserProfileRequest{Profile: pTeacher},
			expectedResp: &pb.UpdateUserProfileResponse{
				Successful: true,
			},
			expectedErr: status.Error(codes.PermissionDenied, "user can only update own profile"),
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, mock.Anything, mock.Anything).Once().Return(&entity.LegacyUser{Group: database.Text(constant.UserGroupTeacher)}, nil)
				userRepo.On("UpdateProfileV1", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectUserDeviceTokenUpdated, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name: "cannot find user",
			ctx:  interceptors.ContextWithUserID(ctx, pTeacher.Id),
			req: &pb.UpdateUserProfileRequest{
				Profile: pTeacher,
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("failed to get user: %s", pgx.ErrNoRows)),
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log("test case:", testCase.name)
			testCase.setup(testCase.ctx)

			req := testCase.req.(*pb.UpdateUserProfileRequest)
			resp, err := s.UpdateUserProfile(testCase.ctx, req)
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

func TestToUserEntity(t *testing.T) {
	t.Parallel()
	profile := generateUserProfile(entity.UserGroupSchoolAdmin)
	req := &pb.UpdateUserProfileRequest{Profile: profile}
	e, _ := toUserEntity(req)

	assert.True(t, isEqualUserEnAndPb(profile, e))

}
func isEqualUserEnAndPb(p *pb.UpdateUserProfileRequest_UserProfile, e *entity.LegacyUser) bool {
	firstName, lastName := SplitNameToFirstNameAndLastName(p.Name)

	return (e.GetName() == p.Name) &&
		// (pb.Country(pb.Country_value[e.Country.String]) == p.Country) &&
		(e.Email.String == p.Email) &&
		(e.Avatar.String == p.Avatar) &&
		(e.DeviceToken.String == p.DeviceToken) &&
		(e.Group.String == p.Group) &&
		(e.LastName.String == lastName) &&
		(e.FirstName.String == firstName)
}

func generateUserProfile(userGroup string) *pb.UpdateUserProfileRequest_UserProfile {
	rand.Seed(time.Now().UnixNano())
	return &pb.UpdateUserProfileRequest_UserProfile{
		Id:          fmt.Sprintf("%d", rand.Int()),
		Name:        fmt.Sprintf("user %d", rand.Int()),
		Country:     cpb.Country_COUNTRY_VN,
		PhoneNumber: fmt.Sprintf("+849%d", rand.Int()),
		Email:       fmt.Sprintf("valid-%d@email.com", rand.Int()),
		Avatar:      fmt.Sprintf("http://avatar-%d", rand.Int()),
		DeviceToken: fmt.Sprintf("random device %d", rand.Int()),
		Group:       userGroup,
	}
}

func generateUserProfileWithoutField(userGroup, missingFieldName string) *pb.UpdateUserProfileRequest_UserProfile {
	rand.Seed(time.Now().UnixNano())
	p := &pb.UpdateUserProfileRequest_UserProfile{
		Id:          fmt.Sprintf("%d", rand.Int()),
		Name:        fmt.Sprintf("user %d", rand.Int()),
		Country:     cpb.Country_COUNTRY_VN,
		PhoneNumber: fmt.Sprintf("+849%d", rand.Int()),
		Email:       fmt.Sprintf("valid-%d@email.com", rand.Int()),
		Avatar:      fmt.Sprintf("http://avatar-%d", rand.Int()),
		DeviceToken: fmt.Sprintf("random device %d", rand.Int()),
		Group:       userGroup,
	}
	switch missingFieldName {
	case "Id":
		p.Id = ""
	case "PhoneNumber":
		p.PhoneNumber = ""
	case "Name":
		p.Name = ""
	case "UserGroup":
		p.Group = ""
	}
	return p
}

func generateUserCheckUserGroup() *pb.UpdateUserProfileRequest_UserProfile {
	return &pb.UpdateUserProfileRequest_UserProfile{
		Id:    "id",
		Name:  fmt.Sprintf("user %d", rand.Int()),
		Group: entity.UserGroupAdmin,
	}
}
