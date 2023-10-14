package interceptors

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGroupDecider_Check(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	allowedGroups := map[string][]string{
		"/usermgmt.v2/CreateUser": {constant.RoleSchoolAdmin, constant.RoleTeacher},
		"/usermgmt.v2/Login":      nil,
	}

	testCases := []struct {
		name          string
		userID        string
		fullMethod    string
		groupFetcher  func(ctx context.Context, userID string) ([]string, error)
		expected      []string
		expectedError error
	}{
		{
			name: "happy case",
			groupFetcher: func(ctx context.Context, userID string) ([]string, error) {
				return []string{constant.RoleSchoolAdmin}, nil
			},
			userID:     idutil.ULIDNow(),
			fullMethod: "/usermgmt.v2/CreateUser",
			expected:   []string{constant.RoleSchoolAdmin},
		},
		{
			name: "success: method don't require permission",
			groupFetcher: func(ctx context.Context, userID string) ([]string, error) {
				return []string{constant.RoleSchoolAdmin}, nil
			},
			userID:     idutil.ULIDNow(),
			fullMethod: "/usermgmt.v2/Login",
			expected:   []string{constant.RoleSchoolAdmin},
		},
		{
			name: "missing decider",
			groupFetcher: func(ctx context.Context, userID string) ([]string, error) {
				return []string{constant.RoleSchoolAdmin}, nil
			},
			userID:        idutil.ULIDNow(),
			fullMethod:    "/usermgmt.v2/CreateStudent",
			expectedError: sttNoDeciderProvided,
		},
		{
			name: "user had been deactivated",
			groupFetcher: func(ctx context.Context, userID string) ([]string, error) {
				return nil, sttDeactivatedUser
			},
			userID:        idutil.ULIDNow(),
			fullMethod:    "/usermgmt.v2/CreateUser",
			expectedError: sttDeactivatedUser,
		},
		{
			name: "internal error when fetching user's group",
			groupFetcher: func(ctx context.Context, userID string) ([]string, error) {
				return nil, assert.AnError
			},
			userID:        idutil.ULIDNow(),
			fullMethod:    "/usermgmt.v2/CreateUser",
			expectedError: sttDeniedAll,
		},
		{
			name: "user hasn't been granted role",
			groupFetcher: func(ctx context.Context, userID string) ([]string, error) {
				return nil, nil
			},
			userID:        idutil.ULIDNow(),
			fullMethod:    "/usermgmt.v2/CreateUser",
			expectedError: sttDeniedAll,
		},
		{
			name: "not allow",
			groupFetcher: func(ctx context.Context, userID string) ([]string, error) {
				return []string{constant.RoleStudent}, nil
			},
			userID:        idutil.ULIDNow(),
			fullMethod:    "/usermgmt.v2/CreateUser",
			expectedError: sttNotAllowed,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			groupDecider := GroupDecider{
				GroupFetcher:  testCase.groupFetcher,
				AllowedGroups: allowedGroups,
			}
			groups, err := groupDecider.Check(ctx, testCase.userID, testCase.fullMethod)
			assert.Equal(t, testCase.expectedError, err)
			assert.True(t, stringutil.SliceEqual(groups, testCase.expected))
		})
	}
}

func TestRetrieveUserRoles(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	userRepo := new(mock_repositories.MockUserRepo)

	user := &entity.LegacyUser{
		ID:    database.Text(idutil.ULIDNow()),
		Group: database.Text(constant.UserGroupSchoolAdmin),
	}
	roleSchoolAdmin := &entity.Role{
		RoleID:   database.Text(idutil.ULIDNow()),
		RoleName: database.Text(constant.RoleSchoolAdmin),
	}
	roleStudent := &entity.Role{
		RoleID:   database.Text(idutil.ULIDNow()),
		RoleName: database.Text(constant.RoleStudent),
	}

	activeUserGroupMemberAdmin := &entity.UserGroupMember{
		UserID:      user.UserID,
		UserGroupID: database.Text(idutil.ULIDNow()),
		DeletedAt: pgtype.Timestamptz{
			Status: pgtype.Null,
			Time:   time.Now(),
		},
	}
	activeUserGroupMemberStudent := &entity.UserGroupMember{
		UserID:      user.UserID,
		UserGroupID: database.Text(idutil.ULIDNow()),
		DeletedAt: pgtype.Timestamptz{
			Status: pgtype.Null,
			Time:   time.Now(),
		},
	}
	userGroupMembers := []*entity.UserGroupMember{activeUserGroupMemberAdmin}

	testCases := []struct {
		name        string
		ctx         context.Context
		userID      string
		setup       func(ctx context.Context)
		expectedErr error
		expected    []string
	}{
		{
			name:   "happy case",
			userID: user.ID.String,
			setup: func(ctx context.Context) {
				userRepo.On("GetUserGroupMembers", ctx, mock.Anything, mock.Anything).Once().Return(append(userGroupMembers, activeUserGroupMemberStudent), nil)
				userRepo.On("GetUserRoles", ctx, mock.Anything, mock.Anything).Once().Return(entity.Roles{roleSchoolAdmin, roleStudent}, nil)
			},
			expected: []string{roleSchoolAdmin.RoleName.String, roleStudent.RoleName.String},
		},
		{
			name:        "missing user id",
			userID:      "",
			expectedErr: status.Error(codes.Unauthenticated, "missing user id"),
		},
		{
			name:   "cannot get user_group's member",
			userID: user.ID.String,
			setup: func(ctx context.Context) {
				userRepo.On("GetUserGroupMembers", ctx, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("get user group member: %w", pgx.ErrNoRows).Error()),
		},
		{
			name:   "user hasn't been assigned user_group",
			userID: user.ID.String,
			setup: func(ctx context.Context) {
				userRepo.On("GetUserGroupMembers", ctx, mock.Anything, mock.Anything).Once().Return([]*entity.UserGroupMember{}, nil)
			},
			expectedErr: status.Error(codes.Internal, "user haven't been assigned user_group"),
		},
		{
			name:   "cannot get user's roles",
			userID: user.ID.String,
			setup: func(ctx context.Context) {
				userRepo.On("GetUserGroupMembers", ctx, mock.Anything, mock.Anything).Once().Return(userGroupMembers, nil)
				userRepo.On("GetUserRoles", ctx, mock.Anything, mock.Anything).Once().Return(entity.Roles{}, pgx.ErrNoRows)
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("get user's roles: %w", pgx.ErrNoRows).Error()),
		},
		{
			name:   "user's role empty",
			userID: user.ID.String,
			setup: func(ctx context.Context) {
				userRepo.On("GetUserGroupMembers", ctx, mock.Anything, mock.Anything).Once().Return(userGroupMembers, nil)
				userRepo.On("GetUserRoles", ctx, mock.Anything, mock.Anything).Once().Return(entity.Roles{}, nil)
			},
			expectedErr: status.Error(codes.Unauthenticated, "missing grant role for user"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.ManabieSchool),
					UserID:       testCase.userID,
				},
			}

			testCase.ctx = interceptors.ContextWithJWTClaims(interceptors.ContextWithUserID(ctx, testCase.userID), claim)
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}

			userRoles, err := RetrieveUserRoles(testCase.ctx, userRepo, db)
			assert.Equal(t, testCase.expectedErr, err)
			assert.True(t, stringutil.SliceEqual(testCase.expected, userRoles))

			mock.AssertExpectationsForObjects(t, db, userRepo)
		})
	}
}

func TestRetrieveUserRolesV2(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	userRepo := new(mock_repositories.MockUserRepo)

	user := &entity.LegacyUser{
		ID:    database.Text(idutil.ULIDNow()),
		Group: database.Text(constant.UserGroupSchoolAdmin),
	}

	roleSchoolAdmin := &entity.Role{
		RoleID:   database.Text(idutil.ULIDNow()),
		RoleName: database.Text(constant.RoleSchoolAdmin),
	}
	roleStudent := &entity.Role{
		RoleID:   database.Text(idutil.ULIDNow()),
		RoleName: database.Text(constant.RoleStudent),
	}

	activeUserGroupMemberAdmin := &entity.UserGroupMember{
		UserID:      user.UserID,
		UserGroupID: database.Text(idutil.ULIDNow()),
		DeletedAt: pgtype.Timestamptz{
			Status: pgtype.Null,
			Time:   time.Now(),
		},
	}
	activeUserGroupMemberStudent := &entity.UserGroupMember{
		UserID:      user.UserID,
		UserGroupID: database.Text(idutil.ULIDNow()),
		DeletedAt: pgtype.Timestamptz{
			Status: pgtype.Null,
			Time:   time.Now(),
		},
	}
	userGroupMembers := []*entity.UserGroupMember{activeUserGroupMemberAdmin}

	testCases := []struct {
		name        string
		ctx         context.Context
		userID      string
		setup       func(ctx context.Context)
		expectedErr error
		expected    []string
	}{
		{
			name:   "happy case",
			userID: user.ID.String,
			setup: func(ctx context.Context) {
				userRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetUserGroupMembers", ctx, mock.Anything, mock.Anything).Once().Return(append(userGroupMembers, activeUserGroupMemberStudent), nil)
				userRepo.On("GetUserRoles", ctx, mock.Anything, mock.Anything).Once().Return(entity.Roles{roleSchoolAdmin, roleStudent}, nil)
			},
			expected: []string{roleSchoolAdmin.RoleName.String, roleStudent.RoleName.String},
		},
		{
			name:        "missing user id",
			userID:      "",
			expectedErr: status.Error(codes.Unauthenticated, "missing user id"),
		},
		{
			name:   "the user was deactivated",
			userID: user.ID.String,
			setup: func(ctx context.Context) {
				// clone another instance of user
				deactivatedUser := *user
				deactivatedUser.DeactivatedAt = database.Timestamptz(time.Now())

				userRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Once().Return(&deactivatedUser, nil)
			},
			expectedErr: sttDeactivatedUser,
		},
		{
			name:   "cannot get user_group's member",
			userID: user.ID.String,
			setup: func(ctx context.Context) {
				userRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetUserGroupMembers", ctx, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("get user group member: %w", pgx.ErrNoRows).Error()),
		},
		{
			name:   "user have't been assigned user_group",
			userID: user.ID.String,
			setup: func(ctx context.Context) {
				userRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetUserGroupMembers", ctx, mock.Anything, mock.Anything).Once().Return([]*entity.UserGroupMember{}, nil)
			},
			expectedErr: status.Error(codes.Internal, "user haven't been assigned user_group"),
		},
		{
			name:   "user's role empty",
			userID: user.ID.String,
			setup: func(ctx context.Context) {
				userRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetUserGroupMembers", ctx, mock.Anything, mock.Anything).Once().Return(userGroupMembers, nil)
				userRepo.On("GetUserRoles", ctx, mock.Anything, mock.Anything).Once().Return(entity.Roles{}, nil)
			},
			expectedErr: status.Error(codes.Unauthenticated, "missing grant role for user"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.ManabieSchool),
					UserID:       testCase.userID,
				},
			}

			testCase.ctx = interceptors.ContextWithJWTClaims(interceptors.ContextWithUserID(ctx, testCase.userID), claim)
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}

			userRoles, err := RetrieveUserRolesV2(testCase.ctx, userRepo, db)
			assert.Equal(t, testCase.expectedErr, err)
			assert.ElementsMatch(t, testCase.expected, userRoles)

			mock.AssertExpectationsForObjects(t, db, userRepo)
		})
	}
}

func TestRetrieveLegacyUserGroups(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	userRepo := new(mock_repositories.MockUserRepo)

	user := &entity.LegacyUser{
		ID:    database.Text(idutil.ULIDNow()),
		Group: database.Text(constant.UserGroupSchoolAdmin),
	}
	roleSchoolAdmin := &entity.Role{
		RoleID:   database.Text(idutil.ULIDNow()),
		RoleName: database.Text(constant.RoleSchoolAdmin),
	}
	roleStudent := &entity.Role{
		RoleID:   database.Text(idutil.ULIDNow()),
		RoleName: database.Text(constant.RoleStudent),
	}

	activeUserGroupMemberAdmin := &entity.UserGroupMember{
		UserID:      user.UserID,
		UserGroupID: database.Text(idutil.ULIDNow()),
		DeletedAt: pgtype.Timestamptz{
			Status: pgtype.Null,
			Time:   time.Now(),
		},
	}
	activeUserGroupMemberStudent := &entity.UserGroupMember{
		UserID:      user.UserID,
		UserGroupID: database.Text(idutil.ULIDNow()),
		DeletedAt: pgtype.Timestamptz{
			Status: pgtype.Null,
			Time:   time.Now(),
		},
	}
	userGroupMembers := []*entity.UserGroupMember{activeUserGroupMemberAdmin}

	testCases := []struct {
		name        string
		ctx         context.Context
		userID      string
		setup       func(ctx context.Context)
		expectedErr error
		expected    []string
	}{
		{
			name:   "happy case",
			userID: user.ID.String,
			setup: func(ctx context.Context) {
				userRepo.On("GetUserGroupMembers", ctx, mock.Anything, mock.Anything).Once().Return(append(userGroupMembers, activeUserGroupMemberStudent), nil)
				userRepo.On("GetUserRoles", ctx, mock.Anything, mock.Anything).Once().Return(entity.Roles{roleSchoolAdmin, roleStudent}, nil)
			},
			expected: []string{constant.UserGroupTeacher, constant.UserGroupSchoolAdmin},
		},
		{
			name:   "cannot get user_group's member",
			userID: user.ID.String,
			setup: func(ctx context.Context) {
				userRepo.On("GetUserGroupMembers", ctx, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("get user group member: %w", pgx.ErrNoRows).Error()),
		},
		{
			name:   "cannot get user's roles",
			userID: user.ID.String,
			setup: func(ctx context.Context) {
				userRepo.On("GetUserGroupMembers", ctx, mock.Anything, mock.Anything).Once().Return(userGroupMembers, nil)
				userRepo.On("GetUserRoles", ctx, mock.Anything, mock.Anything).Once().Return(entity.Roles{}, pgx.ErrNoRows)
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("get user's roles: %w", pgx.ErrNoRows).Error()),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.ManabieSchool),
					UserID:       testCase.userID,
				},
			}

			testCase.ctx = interceptors.ContextWithJWTClaims(interceptors.ContextWithUserID(ctx, testCase.userID), claim)
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}

			userRoles, err := RetrieveLegacyUserGroups(testCase.ctx, userRepo, db)
			assert.Equal(t, testCase.expectedErr, err)
			assert.True(t, stringutil.SliceEqual(testCase.expected, userRoles))

			mock.AssertExpectationsForObjects(t, db, userRepo)
		})
	}
}
