package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	mock_multitenant "github.com/manabie-com/backend/mock/golibs/auth/multitenant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestReissueUserPassword(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	orgRepo := new(mock_repositories.OrganizationRepo)
	userRepo := new(mock_repositories.MockUserRepo)
	tenantManager := new(mock_multitenant.TenantManager)
	firebaseAuthClient := new(mock_multitenant.TenantClient)

	service := UserModifierService{
		DB:                 db,
		UserRepo:           userRepo,
		OrganizationRepo:   orgRepo,
		TenantManager:      tenantManager,
		FirebaseAuthClient: firebaseAuthClient,
	}
	roleSchoolAdmin := &entity.Role{
		RoleName: database.Text(constant.RoleSchoolAdmin),
	}
	roleStudent := &entity.Role{
		RoleName: database.Text(constant.RoleStudent),
	}
	roleStaffTeacher := &entity.Role{
		RoleName: database.Text(constant.RoleTeacher),
	}

	ownerID := idutil.ULIDNow()
	testCases := []TestCase{
		{
			name: "happy case: userGroupSchoolAdmin update user password",
			ctx:  interceptors.ContextWithUserID(ctx, idutil.ULIDNow()),
			req: &pb.ReissueUserPasswordRequest{
				UserId:      "123",
				NewPassword: "Passw0rd",
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{ID: database.Text(idutil.ULIDNow()), Group: database.Text(constant.UserGroupStudent)}, nil)
				userRepo.On("GetUserRoles", ctx, db, mock.Anything).Once().Return(entity.Roles{roleSchoolAdmin}, nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			expectedErr: nil,
		},
		{
			name: "happy case: owner can reissue password",
			ctx:  interceptors.ContextWithUserID(ctx, ownerID),
			req: &pb.ReissueUserPasswordRequest{
				UserId:      "123",
				NewPassword: "Passw0rd",
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{ID: database.Text(ownerID), Group: database.Text(constant.UserGroupStudent)}, nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			expectedErr: nil,
		},
		{
			name: "invalid params: user id empty",
			ctx:  ctx,
			req:  &pb.ReissueUserPasswordRequest{},
			setup: func(ctx context.Context) {
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid params"),
		},
		{
			name: "invalid params: password empty",
			ctx:  ctx,
			req: &pb.ReissueUserPasswordRequest{
				UserId: "123",
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid params"),
		},
		{
			name: "invalid params: password invalid",
			ctx:  ctx,
			req: &pb.ReissueUserPasswordRequest{
				UserId:      "123",
				NewPassword: "pass",
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: status.Error(codes.InvalidArgument, "password length must be larger than 6"),
		},
		{
			name: "cannot find user",
			ctx:  ctx,
			req: &pb.ReissueUserPasswordRequest{
				UserId:      "123",
				NewPassword: "PWssw0rd",
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("failed to get user: %s", pgx.ErrNoRows.Error())),
		},
		{
			name: "can't find caller's roles",
			ctx:  interceptors.ContextWithUserID(ctx, idutil.ULIDNow()),
			req: &pb.ReissueUserPasswordRequest{
				UserId:      "123",
				NewPassword: "Passw0rd",
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{ID: database.Text(idutil.ULIDNow()), Group: database.Text(constant.UserGroupStudent)}, nil)
				userRepo.On("GetUserRoles", ctx, db, mock.Anything).Once().Return(entity.Roles{}, pgx.ErrNoRows)
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("failed to get caller's roles: %s", pgx.ErrNoRows.Error())),
		},
		{
			name: "student can't reissue another user password",
			ctx:  interceptors.ContextWithUserID(ctx, idutil.ULIDNow()),
			req: &pb.ReissueUserPasswordRequest{
				UserId:      "123",
				NewPassword: "Passw0rd",
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{ID: database.Text(idutil.ULIDNow()), Group: database.Text(constant.UserGroupStudent)}, nil)
				userRepo.On("GetUserRoles", ctx, db, mock.Anything).Once().Return(entity.Roles{roleStudent}, nil)
			},
			expectedErr: status.Error(codes.PermissionDenied, "user don't have permission to reissue password"),
		},
		{
			name: "staff role teacher can't reissue another user password",
			ctx:  interceptors.ContextWithUserID(ctx, idutil.ULIDNow()),
			req: &pb.ReissueUserPasswordRequest{
				UserId:      "123",
				NewPassword: "Passw0rd",
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{ID: database.Text(idutil.ULIDNow()), Group: database.Text(constant.UserGroupStudent)}, nil)
				userRepo.On("GetUserRoles", ctx, db, mock.Anything).Once().Return(entity.Roles{roleStaffTeacher}, nil)
			},
			expectedErr: status.Error(codes.PermissionDenied, "user don't have permission to reissue password"),
		},
		{
			name: "only allow to reissue password of student, parent and staff role teacher",
			ctx:  interceptors.ContextWithUserID(ctx, idutil.ULIDNow()),
			req: &pb.ReissueUserPasswordRequest{
				UserId:      "123",
				NewPassword: "Passw0rd",
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{ID: database.Text(idutil.ULIDNow()), Group: database.Text(constant.UserGroupSchoolAdmin)}, nil)
				userRepo.On("GetUserRoles", ctx, db, mock.Anything).Once().Return(entity.Roles{roleSchoolAdmin}, nil)
			},
			expectedErr: status.Error(codes.PermissionDenied, "school staff and school admin don't have permission to reissue this user password"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.ManabieSchool),
				},
			}
			testCase.ctx = interceptors.ContextWithJWTClaims(testCase.ctx, claim)
			testCase.setup(testCase.ctx)

			_, err := service.ReissueUserPassword(testCase.ctx, testCase.req.(*pb.ReissueUserPasswordRequest))
			assert.Equal(t, testCase.expectedErr, err)

			mock.AssertExpectationsForObjects(t, db, userRepo, orgRepo, firebaseAuthClient, tenantManager)
		})
	}
}
