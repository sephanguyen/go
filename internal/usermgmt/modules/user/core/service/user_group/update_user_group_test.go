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

	"github.com/jackc/pgx/v4"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUserGroupService_UpdateUserGroup(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	grantedRoleRepo := new(mock_repositories.MockGrantedRoleRepo)
	userGroupV2Repo := new(mock_repositories.MockUserGroupV2Repo)
	roleRepo := new(mock_repositories.MockRoleRepo)
	locationRepo := new(mock_locationRepo.MockLocationRepo)
	grantedRoleAccessPathRepo := new(mock_repositories.MockGrantedRoleAccessPathRepo)
	userRepo := new(mock_repositories.MockUserRepo)
	teacherRepo := new(mock_repositories.MockTeacherRepo)
	schoolAdminRepo := new(mock_repositories.MockSchoolAdminRepo)
	userGroupRepo := new(mock_repositories.MockUserGroupRepo)
	unleashClient := new(mock_unleash_client.UnleashClientInstance)
	jsm := new(mock_nats.JetStreamManagement)

	tx := new(mock_database.Tx)

	umsvc := &usvc.UserModifierService{
		LocationRepo:    locationRepo,
		DB:              tx,
		TeacherRepo:     teacherRepo,
		SchoolAdminRepo: schoolAdminRepo,
	}

	service := &UserGroupService{
		DB:                        tx,
		UnleashClient:             unleashClient,
		RoleRepo:                  roleRepo,
		UserGroupRepo:             userGroupRepo,
		UserGroupV2Repo:           userGroupV2Repo,
		GrantedRoleRepo:           grantedRoleRepo,
		UserRepo:                  userRepo,
		GrantedRoleAccessPathRepo: grantedRoleAccessPathRepo,
		UserModifierService:       umsvc,
		JSM:                       jsm,
	}

	existedUserGroupID := idutil.ULIDNow()
	existedGrantedRoleID := idutil.ULIDNow()
	validResourcePath := fmt.Sprint(constants.ManabieSchool)
	tests := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &pb.UpdateUserGroupRequest{
				UserGroupId:   existedUserGroupID,
				UserGroupName: fmt.Sprintf("updated-user_group: %s", existedUserGroupID),
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx, mock.Anything).Twice().Return(tx, nil)
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Twice().Return([]*entity.Role{{RoleID: database.Text(idutil.ULIDNow()), RoleName: database.Text(constant.RoleSchoolAdmin), ResourcePath: database.Text(validResourcePath)}}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(false, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, tx, mock.Anything, false).Once().Return([]*domain.Location{{LocationID: idutil.ULIDNow(), ResourcePath: validResourcePath}}, nil)
				userGroupV2Repo.On("Find", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{UserGroupID: database.Text(existedUserGroupID)}, nil)
				userGroupV2Repo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("GetByUserGroup", ctx, tx, mock.Anything).Once().Return([]*entity.GrantedRole{{GrantedRoleID: database.Text(existedGrantedRoleID)}}, nil)
				grantedRoleRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("GetUsersByUserGroupID", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{{ID: database.Text(idutil.ULIDNow()), ResourcePath: database.Text(validResourcePath)}}, nil)
				schoolAdminRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				teacherRepo.On("SoftDeleteMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("UpdateManyUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupRepo.On("DeactivateMultiple", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUpsertUserGroupEvent", constants.SubjectUpsertUserGroup, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name: "happy case: update user group success without roleWithLocations",
			ctx:  ctx,
			req: &pb.UpdateUserGroupRequest{
				UserGroupId:       existedUserGroupID,
				UserGroupName:     fmt.Sprintf("updated-user_group: %s", existedUserGroupID),
				RoleWithLocations: []*pb.RoleWithLocations{},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx, mock.Anything).Once().Return(tx, nil)
				userGroupV2Repo.On("Find", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{UserGroupID: database.Text(existedUserGroupID)}, nil)
				grantedRoleRepo.On("GetByUserGroup", ctx, tx, mock.Anything).Once().Return([]*entity.GrantedRole{{GrantedRoleID: database.Text(existedGrantedRoleID)}}, nil)
				userGroupV2Repo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Once().Return([]*entity.Role{{RoleID: database.Text(idutil.ULIDNow()), ResourcePath: database.Text(validResourcePath)}}, nil)
				userRepo.On("GetUsersByUserGroupID", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{{ID: database.Text(idutil.ULIDNow()), ResourcePath: database.Text(validResourcePath)}}, nil)
				teacherRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				schoolAdminRepo.On("SoftDeleteMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("UpdateManyUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupRepo.On("DeactivateMultiple", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUpsertUserGroupEvent", constants.SubjectUpsertUserGroup, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name: "happy case: remove all role with locations",
			ctx:  ctx,
			req: &pb.UpdateUserGroupRequest{
				UserGroupId:   existedUserGroupID,
				UserGroupName: fmt.Sprintf("updated-user_group: %s", existedUserGroupID),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx, mock.Anything).Once().Return(tx, nil)
				userGroupV2Repo.On("Find", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{UserGroupID: database.Text(existedUserGroupID)}, nil)
				grantedRoleRepo.On("GetByUserGroup", ctx, tx, mock.Anything).Once().Return([]*entity.GrantedRole{{GrantedRoleID: database.Text(existedGrantedRoleID), RoleID: database.Text(existedGrantedRoleID)}}, nil)
				userGroupV2Repo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Once().Return([]*entity.Role{{RoleID: database.Text(idutil.ULIDNow()), ResourcePath: database.Text(validResourcePath)}}, nil)
				userRepo.On("GetUsersByUserGroupID", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{{ID: database.Text(idutil.ULIDNow()), ResourcePath: database.Text(validResourcePath)}}, nil)
				teacherRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				schoolAdminRepo.On("SoftDeleteMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("UpdateManyUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupRepo.On("DeactivateMultiple", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUpsertUserGroupEvent", constants.SubjectUpsertUserGroup, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name: "happy case: not have admin roles",
			ctx:  ctx,
			req: &pb.UpdateUserGroupRequest{
				UserGroupId:   existedUserGroupID,
				UserGroupName: fmt.Sprintf("updated-user_group: %s", existedUserGroupID),
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
			expectedErr: nil,
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx, mock.Anything).Twice().Return(tx, nil)
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Twice().Return([]*entity.Role{{RoleID: database.Text(idutil.ULIDNow()), RoleName: database.Text(constant.RoleTeacher), ResourcePath: database.Text(validResourcePath)}, {RoleID: database.Text(idutil.ULIDNow()), RoleName: database.Text(constant.RoleTeacherLead), ResourcePath: database.Text(validResourcePath)}}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, tx, mock.Anything, false).Once().Return([]*domain.Location{{LocationID: idutil.ULIDNow(), ResourcePath: validResourcePath}, {LocationID: idutil.ULIDNow(), ResourcePath: validResourcePath}}, nil)
				userGroupV2Repo.On("Find", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{UserGroupID: database.Text(existedUserGroupID)}, nil)
				userGroupV2Repo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("GetByUserGroup", ctx, tx, mock.Anything).Once().Return([]*entity.GrantedRole{{GrantedRoleID: database.Text(existedGrantedRoleID)}}, nil)
				grantedRoleRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("GetUsersByUserGroupID", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{{ID: database.Text(idutil.ULIDNow()), ResourcePath: database.Text(validResourcePath)}}, nil)
				teacherRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				schoolAdminRepo.On("SoftDeleteMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("UpdateManyUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupRepo.On("DeactivateMultiple", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUpsertUserGroupEvent", constants.SubjectUpsertUserGroup, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name: "happy case: have admin roles",
			ctx:  ctx,
			req: &pb.UpdateUserGroupRequest{
				UserGroupId:   existedUserGroupID,
				UserGroupName: fmt.Sprintf("updated-user_group: %s", existedUserGroupID),
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
			expectedErr: nil,
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx, mock.Anything).Twice().Return(tx, nil)
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Twice().Return([]*entity.Role{{RoleID: database.Text(idutil.ULIDNow()), RoleName: database.Text(constant.RoleTeacher), ResourcePath: database.Text(validResourcePath)}, {RoleID: database.Text(idutil.ULIDNow()), RoleName: database.Text(constant.RoleHQStaff), ResourcePath: database.Text(validResourcePath)}}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, tx, mock.Anything, false).Once().Return([]*domain.Location{{LocationID: idutil.ULIDNow(), ResourcePath: validResourcePath}, {LocationID: idutil.ULIDNow(), ResourcePath: validResourcePath}}, nil)
				userGroupV2Repo.On("Find", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{UserGroupID: database.Text(existedUserGroupID)}, nil)
				userGroupV2Repo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("GetByUserGroup", ctx, tx, mock.Anything).Once().Return([]*entity.GrantedRole{{GrantedRoleID: database.Text(existedGrantedRoleID)}}, nil)
				grantedRoleRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("GetUsersByUserGroupID", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{{ID: database.Text(idutil.ULIDNow()), ResourcePath: database.Text(validResourcePath)}}, nil)
				schoolAdminRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				teacherRepo.On("SoftDeleteMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("UpdateManyUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupRepo.On("DeactivateMultiple", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUpsertUserGroupEvent", constants.SubjectUpsertUserGroup, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name: "invalid argument: userGroupID empty",
			ctx:  ctx,
			req: &pb.UpdateUserGroupRequest{
				UserGroupId:   "",
				UserGroupName: fmt.Sprintf("updated-user_group: %s", existedUserGroupID),
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.Wrap(fmt.Errorf("userGroupID empty"), "validateUpdateUserGroupParams").Error()),
		},
		{
			name: "invalid argument: roleID empty",
			ctx:  ctx,
			req: &pb.UpdateUserGroupRequest{
				UserGroupId:   existedUserGroupID,
				UserGroupName: fmt.Sprintf("updated-user_group: %s", existedUserGroupID),
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      "",
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.Wrap(fmt.Errorf("roleID empty"), "validateUpdateUserGroupParams").Error()),
		},
		{
			name: "invalid argument: role missing location",
			ctx:  ctx,
			req: &pb.UpdateUserGroupRequest{
				UserGroupId:   existedUserGroupID,
				UserGroupName: fmt.Sprintf("updated-user_group: %s", existedUserGroupID),
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: nil,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.Wrap(fmt.Errorf("granted role missing location"), "validateUpdateUserGroupParams").Error()),
		},
		{
			name: "invalid argument: locationID empty",
			ctx:  ctx,
			req: &pb.UpdateUserGroupRequest{
				UserGroupId:   existedUserGroupID,
				UserGroupName: fmt.Sprintf("updated-user_group: %s", existedUserGroupID),
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{""},
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.Wrap(fmt.Errorf("locationID empty"), "validateUpdateUserGroupParams").Error()),
		},
		{
			name: "invalid argument: userGroupName empty",
			ctx:  ctx,
			req: &pb.UpdateUserGroupRequest{
				UserGroupId:   existedUserGroupID,
				UserGroupName: "",
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.Wrap(fmt.Errorf("userGroupName empty"), "validateUpdateUserGroupParams").Error()),
		},
		{
			name: "invalid argument: cannot get role",
			ctx:  ctx,
			req: &pb.UpdateUserGroupRequest{
				UserGroupId:   existedUserGroupID,
				UserGroupName: fmt.Sprintf("updated-user_group: %s", existedUserGroupID),
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			setup: func(ctx context.Context) {
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Once().Return(nil, errors.New("cannot find role"))
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.Wrap(fmt.Errorf("error when get roles by role ids"), "validRoleWithLocations").Error()),
		},
		{
			name: "invalid argument: combination role invalid",
			ctx:  ctx,
			req: &pb.UpdateUserGroupRequest{
				UserGroupId:   existedUserGroupID,
				UserGroupName: fmt.Sprintf("updated-user_group: %s", existedUserGroupID),
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
			setup: func(ctx context.Context) {
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Once().
					Return([]*entity.Role{{RoleName: database.Text(constant.RoleSchoolAdmin)}, {RoleName: database.Text(constant.RoleTeacher)}}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(false, nil)
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.Wrap(errNotAllowedCombinationRole, "validRoleWithLocations").Error()),
		},
		{
			name: "invalid argument: error when get locations",
			ctx:  ctx,
			req: &pb.UpdateUserGroupRequest{
				UserGroupId:   existedUserGroupID,
				UserGroupName: fmt.Sprintf("updated-user_group: %s", existedUserGroupID),
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx, mock.Anything).Once().Return(tx, nil)
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Once().Return([]*entity.Role{{RoleID: database.Text(idutil.ULIDNow()), ResourcePath: database.Text(validResourcePath)}}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(false, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, tx, mock.Anything, false).Once().Return(nil, errors.New("cannot get locations"))
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.Wrap(fmt.Errorf("location ids are invalid"), "validRoleWithLocations").Error()),
		},
		{
			name: "internal: cannot find userGroup",
			ctx:  ctx,
			req: &pb.UpdateUserGroupRequest{
				UserGroupId:   existedUserGroupID,
				UserGroupName: fmt.Sprintf("updated-user_group: %s", existedUserGroupID),
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx, mock.Anything).Once().Return(tx, nil)
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Once().Return([]*entity.Role{{RoleID: database.Text(idutil.ULIDNow()), ResourcePath: database.Text(validResourcePath)}}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, tx, mock.Anything, false).Once().Return([]*domain.Location{{LocationID: idutil.ULIDNow(), ResourcePath: validResourcePath}}, nil)
				userGroupV2Repo.On("Find", ctx, tx, mock.Anything).Once().Return(nil, errors.New("user group is not existed"))
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("user group is not existed").Error()),
		},
		{
			name: "internal: cannot get grantedRole by userGroup",
			ctx:  ctx,
			req: &pb.UpdateUserGroupRequest{
				UserGroupId:   existedUserGroupID,
				UserGroupName: fmt.Sprintf("updated-user_group: %s", existedUserGroupID),
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx, mock.Anything).Once().Return(tx, nil)
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Once().Return([]*entity.Role{{RoleID: database.Text(idutil.ULIDNow()), ResourcePath: database.Text(validResourcePath)}}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, tx, mock.Anything, false).Once().Return([]*domain.Location{{LocationID: idutil.ULIDNow(), ResourcePath: validResourcePath}}, nil)
				userGroupV2Repo.On("Find", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{UserGroupID: database.Text(existedUserGroupID)}, nil)
				grantedRoleRepo.On("GetByUserGroup", ctx, tx, mock.Anything).Once().Return(nil, errors.New("cannot find grantedRole by userGroup"))
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("cannot find grantedRole by userGroup").Error()),
		},
		{
			name: "internal: update user group failed",
			ctx:  ctx,
			req: &pb.UpdateUserGroupRequest{
				UserGroupId:   existedUserGroupID,
				UserGroupName: fmt.Sprintf("updated-user_group: %s", existedUserGroupID),
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx, mock.Anything).Twice().Return(tx, nil)
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Once().Return([]*entity.Role{{RoleID: database.Text(idutil.ULIDNow()), ResourcePath: database.Text(validResourcePath)}}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(false, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, tx, mock.Anything, false).Once().Return([]*domain.Location{{LocationID: idutil.ULIDNow(), ResourcePath: validResourcePath}}, nil)
				userGroupV2Repo.On("Find", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{UserGroupID: database.Text(existedUserGroupID)}, nil)
				grantedRoleRepo.On("GetByUserGroup", ctx, tx, mock.Anything).Once().Return([]*entity.GrantedRole{{GrantedRoleID: database.Text(existedGrantedRoleID)}}, nil)
				userGroupV2Repo.On("Update", ctx, tx, mock.Anything).Once().Return(errors.New("update user group failed"))
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("rpc error: code = Internal desc = UserGroupV2Repo.Update: update user group failed").Error()),
		},
		{
			name: "internal: soft delete granted role fail",
			ctx:  ctx,
			req: &pb.UpdateUserGroupRequest{
				UserGroupId:   existedUserGroupID,
				UserGroupName: fmt.Sprintf("updated-user_group: %s", existedUserGroupID),
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx, mock.Anything).Twice().Return(tx, nil)
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Once().Return([]*entity.Role{{RoleID: database.Text(idutil.ULIDNow()), ResourcePath: database.Text(validResourcePath)}}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(false, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, tx, mock.Anything, false).Once().Return([]*domain.Location{{LocationID: idutil.ULIDNow(), ResourcePath: validResourcePath}}, nil)
				userGroupV2Repo.On("Find", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{UserGroupID: database.Text(existedUserGroupID)}, nil)
				grantedRoleRepo.On("GetByUserGroup", ctx, tx, mock.Anything).Once().Return([]*entity.GrantedRole{{GrantedRoleID: database.Text(existedGrantedRoleID)}}, nil)
				userGroupV2Repo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(errors.New(pgx.ErrTxClosed.Error()))
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("rpc error: code = Internal desc = UserGroupV2Repo.Update: %w", pgx.ErrTxClosed).Error()),
		},
		{
			name: "internal: upsert granted role fail",
			ctx:  ctx,
			req: &pb.UpdateUserGroupRequest{
				UserGroupId:   existedUserGroupID,
				UserGroupName: fmt.Sprintf("updated-user_group: %s", existedUserGroupID),
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx, mock.Anything).Twice().Return(tx, nil)
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Once().Return([]*entity.Role{{RoleID: database.Text(idutil.ULIDNow()), ResourcePath: database.Text(validResourcePath)}}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, tx, mock.Anything, false).Once().Return([]*domain.Location{{LocationID: idutil.ULIDNow(), ResourcePath: validResourcePath}}, nil)
				userGroupV2Repo.On("Find", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{UserGroupID: database.Text(existedUserGroupID)}, nil)
				grantedRoleRepo.On("GetByUserGroup", ctx, tx, mock.Anything).Once().Return([]*entity.GrantedRole{{GrantedRoleID: database.Text(existedGrantedRoleID)}}, nil)
				userGroupV2Repo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(errors.New("upsert granted role failed"))
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("rpc error: code = Internal desc = GrantedRoleRepo.Upsert: upsert granted role failed").Error()),
		},
		{
			name: "internal: upsert granted role access path failed",
			ctx:  ctx,
			req: &pb.UpdateUserGroupRequest{
				UserGroupId:   existedUserGroupID,
				UserGroupName: fmt.Sprintf("updated-user_group: %s", existedUserGroupID),
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx, mock.Anything).Twice().Return(tx, nil)
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Once().Return([]*entity.Role{{RoleID: database.Text(idutil.ULIDNow()), ResourcePath: database.Text(validResourcePath)}}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(false, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, tx, mock.Anything, false).Once().Return([]*domain.Location{{LocationID: idutil.ULIDNow(), ResourcePath: validResourcePath}}, nil)
				userGroupV2Repo.On("Find", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{UserGroupID: database.Text(existedUserGroupID)}, nil)
				grantedRoleRepo.On("GetByUserGroup", ctx, tx, mock.Anything).Once().Return([]*entity.GrantedRole{{GrantedRoleID: database.Text(existedGrantedRoleID)}}, nil)
				userGroupV2Repo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(errors.New("upsert granted role access path failed"))
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("rpc error: code = Internal desc = GrantedRoleAccessPathRepo.Upsert: upsert granted role access path failed").Error()),
		},
		{
			name: "internal: get role by role ids",
			ctx:  ctx,
			req: &pb.UpdateUserGroupRequest{
				UserGroupId:   existedUserGroupID,
				UserGroupName: fmt.Sprintf("updated-user_group: %s", existedUserGroupID),
			},
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx, mock.Anything).Once().Return(tx, nil)
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
				userGroupV2Repo.On("Find", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{UserGroupID: database.Text(existedUserGroupID)}, nil)
				grantedRoleRepo.On("GetByUserGroup", ctx, tx, mock.Anything).Once().Return([]*entity.GrantedRole{{GrantedRoleID: database.Text(existedGrantedRoleID), RoleID: database.Text(existedGrantedRoleID)}}, nil)
				userGroupV2Repo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("s.RoleRepo.GetRolesByRoleIDs: error").Error()),
		},
		{
			name: "internal: get users by user group id",
			ctx:  ctx,
			req: &pb.UpdateUserGroupRequest{
				UserGroupId:   existedUserGroupID,
				UserGroupName: fmt.Sprintf("updated-user_group: %s", existedUserGroupID),
			},
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx, mock.Anything).Once().Return(tx, nil)
				userGroupV2Repo.On("Find", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{UserGroupID: database.Text(existedUserGroupID)}, nil)
				grantedRoleRepo.On("GetByUserGroup", ctx, tx, mock.Anything).Once().Return([]*entity.GrantedRole{{GrantedRoleID: database.Text(existedGrantedRoleID), RoleID: database.Text(existedGrantedRoleID)}}, nil)
				userGroupV2Repo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Once().Return([]*entity.Role{{RoleID: database.Text(idutil.ULIDNow()), RoleName: database.Text(constant.RoleSchoolAdmin), ResourcePath: database.Text(validResourcePath)}}, nil)
				userRepo.On("GetUsersByUserGroupID", ctx, tx, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("s.updateLegacyUserGroupOfUsers: s.UserRepo.GetUsersByUserGroupID: error").Error()),
		},
		{
			name: "error publish upsertUserGroup event",
			ctx:  ctx,
			req: &pb.UpdateUserGroupRequest{
				UserGroupId:   existedUserGroupID,
				UserGroupName: fmt.Sprintf("updated-user_group: %s", existedUserGroupID),
				RoleWithLocations: []*pb.RoleWithLocations{
					{
						RoleId:      idutil.ULIDNow(),
						LocationIds: []string{idutil.ULIDNow()},
					},
				},
			},
			expectedErr: errors.New("s.publishUpsertUserGroupEvent: publishUpsertUserGroupEvent with UserGroup.Upserted: s.JSM.Publish failed: error publish event"),
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx, mock.Anything).Twice().Return(tx, nil)
				roleRepo.On("GetRolesByRoleIDs", ctx, tx, mock.Anything).Twice().Return([]*entity.Role{{RoleID: database.Text(idutil.ULIDNow()), RoleName: database.Text(constant.RoleSchoolAdmin), ResourcePath: database.Text(validResourcePath)}}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(false, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, tx, mock.Anything, false).Once().Return([]*domain.Location{{LocationID: idutil.ULIDNow(), ResourcePath: validResourcePath}}, nil)
				userGroupV2Repo.On("Find", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{UserGroupID: database.Text(existedUserGroupID)}, nil)
				userGroupV2Repo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("GetByUserGroup", ctx, tx, mock.Anything).Once().Return([]*entity.GrantedRole{{GrantedRoleID: database.Text(existedGrantedRoleID)}}, nil)
				grantedRoleRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				grantedRoleAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("GetUsersByUserGroupID", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{{ID: database.Text(idutil.ULIDNow()), ResourcePath: database.Text(validResourcePath)}}, nil)
				schoolAdminRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				teacherRepo.On("SoftDeleteMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("UpdateManyUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupRepo.On("DeactivateMultiple", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
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
					ResourcePath: validResourcePath,
				},
			}

			testCase.ctx = interceptors.ContextWithJWTClaims(ctx, claim)
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}

			res, err := service.UpdateUserGroup(testCase.ctx, testCase.req.(*pb.UpdateUserGroupRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.True(t, res.Successful)
			}

			mock.AssertExpectationsForObjects(t, tx, grantedRoleRepo, userGroupV2Repo, roleRepo, locationRepo, grantedRoleAccessPathRepo, userRepo, unleashClient)
		})
	}
}
