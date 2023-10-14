package user_group

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	usvc "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_locationRepo "github.com/manabie-com/backend/mock/mastermgmt/modules/location/infrastructure/repo"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
	"github.com/nats-io/nats.go"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedErr  error
	expectedResp interface{}
	setup        func(ctx context.Context)
	Options      interface{}
}

var orgLocationManabie = &domain.Location{
	LocationID: constants.ManabieOrgLocation,
}

func TestUserGroupService_CreateUserGroup(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tx := new(mock_database.Tx)
	grantedRoleRepo := new(mock_repositories.MockGrantedRoleRepo)
	userGroupV2Repo := new(mock_repositories.MockUserGroupV2Repo)
	roleRepo := new(mock_repositories.MockRoleRepo)
	userRepo := new(mock_repositories.MockUserRepo)
	locationRepo := new(mock_locationRepo.MockLocationRepo)
	unleashClient := new(mock_unleash_client.UnleashClientInstance)
	jsm := new(mock_nats.JetStreamManagement)

	umsvc := &usvc.UserModifierService{
		DB:           tx,
		LocationRepo: locationRepo,
		UserRepo:     userRepo,
	}

	type params struct {
		resourcePath string
	}

	s := &UserGroupService{
		DB:                  tx,
		UnleashClient:       unleashClient,
		UserModifierService: umsvc,
		RoleRepo:            roleRepo,
		GrantedRoleRepo:     grantedRoleRepo,
		UserGroupV2Repo:     userGroupV2Repo,
		LocationRepo:        locationRepo,
		JSM:                 jsm,
	}
	validResourcePath := "1"

	tests := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &pb.CreateUserGroupRequest{
				UserGroupName: "UserGroupName",
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			Options: params{
				resourcePath: validResourcePath,
			},
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx, mock.Anything).Twice().Return(tx, nil)
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Once().Return([]*entity.Role{{RoleID: database.Text(idutil.ULIDNow()), ResourcePath: database.Text(validResourcePath)}}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, tx, mock.Anything, false).Once().Return([]*domain.Location{{LocationID: idutil.ULIDNow(), ResourcePath: validResourcePath}}, nil)
				locationRepo.On("GetLocationOrg", ctx, tx, mock.Anything).Once().Return(orgLocationManabie, nil)
				userGroupV2Repo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("LinkGrantedRoleToAccessPath", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUpsertUserGroupEvent", constants.SubjectUpsertUserGroup, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name: "resource path is invalid",
			ctx:  ctx,
			req:  &pb.CreateUserGroupRequest{},
			Options: params{
				resourcePath: "",
			},
			expectedErr: status.Error(codes.InvalidArgument, "resource path is invalid"),
		},
		{
			name: "error when get roles by role ids",
			ctx:  ctx,
			req: &pb.CreateUserGroupRequest{
				UserGroupName: "UserGroupName",
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			Options: params{
				resourcePath: validResourcePath,
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.Wrap(fmt.Errorf("error when get roles by role ids"), "validRoleWithLocations").Error()),
			setup: func(ctx context.Context) {
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Once().Return(nil, fmt.Errorf("error when getting roles"))
			},
		},
		{
			name: "role ids are invalid",
			ctx:  ctx,
			req: &pb.CreateUserGroupRequest{
				UserGroupName: "UserGroupName",
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			Options: params{
				resourcePath: validResourcePath,
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.Wrap(fmt.Errorf("role ids are invalid"), "validRoleWithLocations").Error()),
			setup: func(ctx context.Context) {
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Once().Return([]*entity.Role{{}, {}}, nil)
			},
		},
		{
			name: "error when get locations by role ids",
			ctx:  ctx,
			req: &pb.CreateUserGroupRequest{
				UserGroupName: "UserGroupName",
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			Options: params{
				resourcePath: validResourcePath,
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.Wrap(fmt.Errorf("location ids are invalid"), "validRoleWithLocations").Error()),
			setup: func(ctx context.Context) {
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Once().Return([]*entity.Role{{ResourcePath: database.Text(validResourcePath)}}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				tx.On("Begin", ctx, mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, tx, mock.Anything, false).Once().Return(nil, fmt.Errorf("error when getting locations"))
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "combination role invalid",
			ctx:  ctx,
			req: &pb.CreateUserGroupRequest{
				UserGroupName: "UserGroupName",
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			Options: params{
				resourcePath: validResourcePath,
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.Wrap(errNotAllowedCombinationRole, "validRoleWithLocations").Error()),
			setup: func(ctx context.Context) {
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Once().
					Return([]*entity.Role{{RoleName: database.Text(constant.RoleSchoolAdmin)}, {RoleName: database.Text(constant.RoleTeacher)}}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(false, nil)
			},
		},
		{
			name: "error when get org location",
			ctx:  ctx,
			req: &pb.CreateUserGroupRequest{
				UserGroupName: "UserGroupName",
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			Options: params{
				resourcePath: validResourcePath,
			},
			expectedErr: fmt.Errorf("s.HandleCreateUserGroup: database.ExecInTx: %w", fmt.Errorf("s.LocationRepo.GetLocationOrg: %w", pgx.ErrNoRows)),
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx, mock.Anything).Twice().Return(tx, nil)
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Once().Return([]*entity.Role{{ResourcePath: database.Text(validResourcePath)}}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, tx, mock.Anything, false).Once().Return([]*domain.Location{{LocationID: idutil.ULIDNow(), ResourcePath: validResourcePath}}, nil)
				locationRepo.On("GetLocationOrg", ctx, tx, mock.Anything).Once().Return(nil, errors.New(pgx.ErrNoRows.Error()))
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "error when linking granted role to access path",
			ctx:  ctx,
			req: &pb.CreateUserGroupRequest{
				UserGroupName: "UserGroupName",
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			Options: params{
				resourcePath: validResourcePath,
			},
			expectedErr: fmt.Errorf(
				"s.HandleCreateUserGroup: %w",
				fmt.Errorf(
					"database.ExecInTx: %w",
					fmt.Errorf(
						"s.GrantedRoleRepo.LinkGrantedRoleToAccessPath: %w",
						fmt.Errorf("error when linking granted role to access path"),
					),
				),
			),
			setup: func(ctx context.Context) {
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Once().Return([]*entity.Role{{RoleID: database.Text(idutil.ULIDNow()), ResourcePath: database.Text(validResourcePath)}}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				tx.On("Begin", ctx, mock.Anything).Twice().Return(tx, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, tx, mock.Anything, false).Once().Return([]*domain.Location{{LocationID: idutil.ULIDNow(), ResourcePath: validResourcePath}}, nil)
				locationRepo.On("GetLocationOrg", ctx, tx, mock.Anything).Once().Return(orgLocationManabie, nil)
				userGroupV2Repo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("LinkGrantedRoleToAccessPath", ctx, tx, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error when linking granted role to access path"))
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "error publish upsertUserGroup event",
			ctx:  ctx,
			req: &pb.CreateUserGroupRequest{
				UserGroupName: "UserGroupName",
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			Options: params{
				resourcePath: validResourcePath,
			},
			expectedErr: errors.New("s.publishUpsertUserGroupEvent: publishUpsertUserGroupEvent with UserGroup.Upserted: s.JSM.Publish failed: error publish event"),
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx, mock.Anything).Twice().Return(tx, nil)
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Once().Return([]*entity.Role{{RoleID: database.Text(idutil.ULIDNow()), ResourcePath: database.Text(validResourcePath)}}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, tx, mock.Anything, false).Once().Return([]*domain.Location{{LocationID: idutil.ULIDNow(), ResourcePath: validResourcePath}}, nil)
				locationRepo.On("GetLocationOrg", ctx, tx, mock.Anything).Once().Return(orgLocationManabie, nil)
				userGroupV2Repo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("LinkGrantedRoleToAccessPath", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUpsertUserGroupEvent", constants.SubjectUpsertUserGroup, mock.Anything).Once().Return(nil, errors.New("error publish event"))
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log(testCase.name)

			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: testCase.Options.(params).resourcePath,
				},
			}
			testCase.ctx = interceptors.ContextWithJWTClaims(ctx, claim)

			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}

			resp, err := s.CreateUserGroup(testCase.ctx, testCase.req.(*pb.CreateUserGroupRequest))
			if err != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NotNil(t, resp)
			}

			mock.AssertExpectationsForObjects(t, tx, grantedRoleRepo, userGroupV2Repo, roleRepo, userRepo, locationRepo, unleashClient)
		})
	}
}

func TestUserGroupService_HandleCreateUserGroup(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tx := new(mock_database.Tx)
	roleRepo := new(mock_repositories.MockRoleRepo)
	grantedRoleRepo := new(mock_repositories.MockGrantedRoleRepo)
	userGroupV2Repo := new(mock_repositories.MockUserGroupV2Repo)
	locationRepo := new(mock_locationRepo.MockLocationRepo)

	umsvc := &usvc.UserModifierService{
		DB:           tx,
		LocationRepo: locationRepo,
	}

	type params struct {
		resourcePath int
	}

	s := &UserGroupService{
		DB:                  tx,
		UserModifierService: umsvc,
		RoleRepo:            roleRepo,
		UserGroupV2Repo:     userGroupV2Repo,
		GrantedRoleRepo:     grantedRoleRepo,
		LocationRepo:        locationRepo,
	}
	tests := []TestCase{
		{
			name: "error when creating user group",
			ctx:  ctx,
			req:  &pb.CreateUserGroupRequest{},
			Options: params{
				resourcePath: 0,
			},
			expectedErr: fmt.Errorf("database.ExecInTx: %w", fmt.Errorf("s.UserGroupV2Repo.Create: %w", fmt.Errorf("error when creating user group"))),
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx, mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationOrg", ctx, tx, mock.Anything).Once().Return(orgLocationManabie, nil)
				userGroupV2Repo.On("Create", ctx, tx, mock.Anything).Once().Return(fmt.Errorf("error when creating user group"))
				tx.On("Rollback", mock.Anything).Return(nil)
			},
		},
		{
			name: "error when creating granted role",
			ctx:  ctx,
			req: &pb.CreateUserGroupRequest{
				UserGroupName: "",
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			Options: params{
				resourcePath: 0,
			},
			expectedErr: fmt.Errorf("database.ExecInTx: %w", fmt.Errorf("s.UserGroupV2Repo.Create: %w", fmt.Errorf("error when creating granted role"))),
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx, mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationOrg", ctx, tx, mock.Anything).Once().Return(orgLocationManabie, nil)
				userGroupV2Repo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("Create", ctx, tx, mock.Anything).Once().Return(fmt.Errorf("error when creating granted role"))
				tx.On("Rollback", mock.Anything).Return(nil)
			},
		},
		{
			name: "error when creating user group v2",
			ctx:  ctx,
			req: &pb.CreateUserGroupRequest{
				UserGroupName: "",
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			Options: params{
				resourcePath: 0,
			},
			expectedErr: fmt.Errorf("database.ExecInTx: %w", fmt.Errorf("s.UserGroupV2Repo.Create: %w", fmt.Errorf("error when creating user group"))),
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx, mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationOrg", ctx, tx, mock.Anything).Once().Return(orgLocationManabie, nil)
				userGroupV2Repo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("LinkGrantedRoleToAccessPath", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Return(nil)
			},
		},
		{
			name: "error when linking granted role to access path",
			ctx:  ctx,
			req: &pb.CreateUserGroupRequest{
				UserGroupName: "UserGroupName",
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			Options: params{
				resourcePath: 0,
			},
			expectedErr: fmt.Errorf("database.ExecInTx: %w", fmt.Errorf("s.GrantedRoleRepo.LinkGrantedRoleToAccessPath: %w", fmt.Errorf("error when linking granted role to access path"))),
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx, mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationOrg", ctx, tx, mock.Anything).Once().Return(orgLocationManabie, nil)
				userGroupV2Repo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("LinkGrantedRoleToAccessPath", ctx, tx, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error when linking granted role to access path"))
				tx.On("Commit", mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(testCase.Options.(params).resourcePath),
				},
			}
			testCase.ctx = interceptors.ContextWithJWTClaims(ctx, claim)

			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}

			res, err := s.HandleCreateUserGroup(testCase.ctx, testCase.req.(*pb.CreateUserGroupRequest), testCase.Options.(params).resourcePath)
			if err != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NotNil(t, res)
			}

			mock.AssertExpectationsForObjects(t, tx, userGroupV2Repo, grantedRoleRepo, roleRepo, locationRepo)
		})
	}
}

func TestUserGroupService_ValidationCreateUserGroup(t *testing.T) {
	t.Parallel()

	tests := []TestCase{
		{
			name:        "user group name is empty",
			expectedErr: status.Error(codes.InvalidArgument, "user group name is empty"),
			req: &pb.CreateUserGroupRequest{
				UserGroupName: "",
			},
		},
		{
			name:        "role id is empty",
			expectedErr: status.Error(codes.InvalidArgument, "roleID empty"),
			req: &pb.CreateUserGroupRequest{
				UserGroupName: "UserGroupName",
				RoleWithLocations: []*pb.RoleWithLocations{
					{},
				},
			},
		},
		{
			name:        "location id is empty",
			expectedErr: status.Error(codes.InvalidArgument, "locationID empty"),
			req: &pb.CreateUserGroupRequest{
				UserGroupName: "UserGroupName",
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{""},
					},
				},
			},
		},
		{
			name:        "happy case",
			expectedErr: nil,
			req: &pb.CreateUserGroupRequest{
				UserGroupName: "UserGroupName",
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}

			err := validationCreateUserGroup(testCase.req.(*pb.CreateUserGroupRequest))
			if err != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			}
		})
	}
}

func TestUserGroupService_validRoleWithLocations(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	tx := new(mock_database.Tx)

	roleRepo := new(mock_repositories.MockRoleRepo)
	locationRepo := new(mock_locationRepo.MockLocationRepo)
	unleashClient := new(mock_unleash_client.UnleashClientInstance)

	umsvc := &usvc.UserModifierService{
		DB:           db,
		LocationRepo: locationRepo,
	}

	service := &UserGroupService{
		DB:                  db,
		UnleashClient:       unleashClient,
		UserModifierService: umsvc,
		RoleRepo:            roleRepo,
	}

	testCases := []TestCase{
		{
			name:        "role with location empty",
			ctx:         ctx,
			req:         []*pb.RoleWithLocations{},
			expectedErr: nil,
		},
		{
			name: "cannot query role by ids",
			ctx:  ctx,
			req:  []*pb.RoleWithLocations{{RoleId: idutil.ULIDNow()}},
			setup: func(ctx context.Context) {
				roleRepo.On("GetRolesByRoleIDs", ctx, db, mock.Anything).Once().Return(nil, fmt.Errorf("error when getting roles"))
			},
			expectedErr: fmt.Errorf("error when get roles by role ids"),
		},
		{
			name: "list roles return are different from request",
			ctx:  ctx,
			req:  []*pb.RoleWithLocations{{RoleId: idutil.ULIDNow()}, {RoleId: idutil.ULIDNow()}},
			setup: func(ctx context.Context) {
				roleRepo.On("GetRolesByRoleIDs", ctx, db, mock.Anything).Once().
					Return([]*entity.Role{{RoleName: database.Text(constant.RoleSchoolAdmin)}}, nil)
			},
			expectedErr: fmt.Errorf("role ids are invalid"),
		},
		{
			name: "disable toggle FeatureToggleAllowCombinationMultipleRoles: combine role failed",
			ctx:  ctx,
			req:  []*pb.RoleWithLocations{{RoleId: idutil.ULIDNow()}, {RoleId: idutil.ULIDNow()}},
			setup: func(ctx context.Context) {
				roleRepo.On("GetRolesByRoleIDs", ctx, db, mock.Anything).Once().
					Return([]*entity.Role{{RoleName: database.Text(constant.RoleSchoolAdmin)}, {RoleName: database.Text(constant.RoleTeacher)}}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(false, nil)
			},
			expectedErr: errNotAllowedCombinationRole,
		},
		{
			name: "disable toggle FeatureToggleAllowCombinationMultipleRoles: valid case",
			ctx:  ctx,
			req:  []*pb.RoleWithLocations{{RoleId: idutil.ULIDNow()}, {RoleId: idutil.ULIDNow()}},
			setup: func(ctx context.Context) {
				roleRepo.On("GetRolesByRoleIDs", ctx, db, mock.Anything).Once().
					Return([]*entity.Role{{RoleName: database.Text(constant.RoleTeacherLead)}, {RoleName: database.Text(constant.RoleTeacher)}}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(false, nil)
			},
			expectedErr: nil,
		},
		{
			name: "enable toggle FeatureToggleAllowCombinationMultipleRoles: cannot query locations",
			ctx:  ctx,
			req:  []*pb.RoleWithLocations{{RoleId: idutil.ULIDNow(), LocationIds: []string{idutil.ULIDNow()}}},
			setup: func(ctx context.Context) {
				roleRepo.On("GetRolesByRoleIDs", ctx, db, mock.Anything).Once().
					Return([]*entity.Role{{RoleName: database.Text(constant.RoleTeacherLead)}}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				db.On("Begin", ctx, mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Once().Return(nil, fmt.Errorf("error when getting locations"))
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
			expectedErr: fmt.Errorf("location ids are invalid"),
		},
		{
			name: "enable toggle FeatureToggleAllowCombinationMultipleRoles: valid case",
			ctx:  ctx,
			req:  []*pb.RoleWithLocations{{RoleId: idutil.ULIDNow(), LocationIds: []string{idutil.ULIDNow()}}},
			setup: func(ctx context.Context) {
				roleRepo.On("GetRolesByRoleIDs", ctx, db, mock.Anything).Once().
					Return([]*entity.Role{{RoleName: database.Text(constant.RoleTeacherLead)}}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				db.On("Begin", ctx, mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Once().Return([]*domain.Location{{LocationID: idutil.ULIDNow()}}, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			expectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log(testCase.name)

			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}

			err := service.validRoleWithLocations(ctx, testCase.req.([]*pb.RoleWithLocations))
			assert.Equal(t, testCase.expectedErr, err)
		})
		mock.AssertExpectationsForObjects(t, tx, db, roleRepo, unleashClient, locationRepo)
	}
}
