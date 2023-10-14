package user_group

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUserGroupService_ValidateUserLogin(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	validResourcePath := fmt.Sprint(constants.ManabieSchool)
	tx := new(mock_database.Tx)
	userRepo := new(mock_repositories.MockUserRepo)
	userGroupV2Repo := new(mock_repositories.MockUserGroupV2Repo)
	unleashClient := new(mock_unleash_client.UnleashClientInstance)

	ugs := &UserGroupService{
		UserGroupV2Repo: userGroupV2Repo,
		UserRepo:        userRepo,
		UserModifierService: &service.UserModifierService{
			UnleashClient: unleashClient,
		},
		DB: tx,
	}

	type params struct {
		resourcePath string
	}

	existingUser := &entity.LegacyUser{
		ID:          database.Text(idutil.ULIDNow()),
		Email:       database.Text("existing-user-email@example.com"),
		PhoneNumber: database.Text("existing-user-phone-number"),
		Group:       database.Text(constant.UserGroupSchoolAdmin),
	}

	tests := []TestCase{
		{
			name:    "user not found",
			ctx:     ctx,
			req:     &pb.ValidateUserLoginRequest{Platform: cpb.Platform_PLATFORM_BACKOFFICE},
			Options: params{resourcePath: validResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabled", isAllowRolesToLoginTeacherWeb, mock.Anything).Once().Return(true, nil)
				userRepo.On("Get", ctx, tx, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
			expectedErr: status.Errorf(codes.Internal, errors.Wrap(pgx.ErrNoRows, "get user failed").Error()),
		},
		{
			name:    "err when get user",
			ctx:     ctx,
			req:     &pb.ValidateUserLoginRequest{Platform: cpb.Platform_PLATFORM_BACKOFFICE},
			Options: params{resourcePath: validResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabled", isAllowRolesToLoginTeacherWeb, mock.Anything).Once().Return(true, nil)
				userRepo.On("Get", ctx, tx, mock.Anything).Once().Return(nil, puddle.ErrClosedPool)
			},
			expectedErr: status.Errorf(codes.Internal, errors.Wrap(puddle.ErrClosedPool, "get user failed").Error()),
		},
		{
			name:    "err when get user group",
			ctx:     ctx,
			req:     &pb.ValidateUserLoginRequest{Platform: cpb.Platform_PLATFORM_BACKOFFICE},
			Options: params{resourcePath: validResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabled", isAllowRolesToLoginTeacherWeb, mock.Anything).Once().Return(true, nil)
				userRepo.On("Get", ctx, tx, mock.Anything).Once().Return(existingUser, nil)
				userGroupV2Repo.On("FindUserGroupAndRoleByUserID", ctx, tx, mock.Anything).Once().Return(nil, puddle.ErrClosedPool)
			},
			expectedErr: status.Errorf(codes.Internal, errors.Wrap(puddle.ErrClosedPool, "find user group and roles failed").Error()),
		},
		{
			name:    "user have not been assigned user_group",
			ctx:     ctx,
			req:     &pb.ValidateUserLoginRequest{Platform: cpb.Platform_PLATFORM_BACKOFFICE},
			Options: params{resourcePath: validResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabled", isAllowRolesToLoginTeacherWeb, mock.Anything).Once().Return(true, nil)
				userRepo.On("Get", ctx, tx, mock.Anything).Once().Return(&entity.LegacyUser{}, nil)
				userGroupV2Repo.On("FindUserGroupAndRoleByUserID", ctx, tx, mock.Anything).Once().Return(nil, nil)
			},
			expectedResp: false,
			expectedErr:  nil,
		},
		{
			name:    "user don't have permission can't access any platform",
			ctx:     ctx,
			req:     &pb.ValidateUserLoginRequest{Platform: cpb.Platform_PLATFORM_BACKOFFICE},
			Options: params{resourcePath: validResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabled", isAllowRolesToLoginTeacherWeb, mock.Anything).Once().Return(true, nil)
				userRepo.On("Get", ctx, tx, mock.Anything).Once().Return(&entity.LegacyUser{}, nil)
				userGroupV2Repo.On("FindUserGroupAndRoleByUserID", ctx, tx, mock.Anything).Once().Return(map[string][]*entity.Role{constant.UserGroupSchoolAdmin: {&entity.Role{}}}, nil)
			},
			expectedResp: false,
			expectedErr:  nil,
		},
		{
			name:    "new roles can't login teacher web when disable toggle isAllowRolesToLoginTeacherWeb",
			ctx:     ctx,
			req:     &pb.ValidateUserLoginRequest{Platform: cpb.Platform_PLATFORM_TEACHER},
			Options: params{resourcePath: validResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabled", isAllowRolesToLoginTeacherWeb, mock.Anything).Once().Return(false, nil)
				userRepo.On("Get", ctx, tx, mock.Anything).Once().Return(&entity.LegacyUser{}, nil)
				userGroupV2Repo.On("FindUserGroupAndRoleByUserID", ctx, tx, mock.Anything).Once().Return(map[string][]*entity.Role{constant.UserGroupSchoolAdmin: {&entity.Role{RoleName: database.Text(constant.RoleHQStaff)}}}, nil)
			},
			expectedResp: false,
			expectedErr:  nil,
		},
		{
			name:    "happy case: user have permission to access this platform",
			ctx:     ctx,
			req:     &pb.ValidateUserLoginRequest{Platform: cpb.Platform_PLATFORM_BACKOFFICE},
			Options: params{resourcePath: validResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabled", isAllowRolesToLoginTeacherWeb, mock.Anything).Once().Return(false, nil)
				userRepo.On("Get", ctx, tx, mock.Anything).Once().Return(&entity.LegacyUser{}, nil)
				userGroupV2Repo.On("FindUserGroupAndRoleByUserID", ctx, tx, mock.Anything).Once().Return(map[string][]*entity.Role{constant.UserGroupSchoolAdmin: {&entity.Role{RoleName: database.Text(constant.RoleSchoolAdmin)}}}, nil)
			},
			expectedResp: true,
			expectedErr:  nil,
		},
		{
			name:    "happy case: new roles can login teacher web when enable toggle isAllowRolesToLoginTeacherWeb",
			ctx:     ctx,
			req:     &pb.ValidateUserLoginRequest{Platform: cpb.Platform_PLATFORM_TEACHER},
			Options: params{resourcePath: validResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabled", isAllowRolesToLoginTeacherWeb, mock.Anything).Once().Return(true, nil)
				userRepo.On("Get", ctx, tx, mock.Anything).Once().Return(&entity.LegacyUser{}, nil)
				userGroupV2Repo.On("FindUserGroupAndRoleByUserID", ctx, tx, mock.Anything).Once().Return(map[string][]*entity.Role{constant.UserGroupSchoolAdmin: {&entity.Role{RoleName: database.Text(constant.RoleHQStaff)}}}, nil)
			},
			expectedResp: true,
			expectedErr:  nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: testCase.Options.(params).resourcePath,
				},
			}

			testCase.ctx = interceptors.ContextWithUserID(
				interceptors.ContextWithJWTClaims(ctx, claim),
				existingUser.ID.String,
			)
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}

			response, err := ugs.ValidateUserLogin(testCase.ctx, testCase.req.(*pb.ValidateUserLoginRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, response.Allowable, testCase.expectedResp)
			}

			mock.AssertExpectationsForObjects(t, tx, userRepo, userGroupV2Repo)
		})
	}
}
