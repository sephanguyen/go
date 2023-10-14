package staff

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
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	ums "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	usvc "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	ugs "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service/user_group"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	entity_mock "github.com/manabie-com/backend/internal/usermgmt/pkg/mock"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	mock_multitenant "github.com/manabie-com/backend/mock/golibs/auth/multitenant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_firebase "github.com/manabie-com/backend/mock/golibs/firebase"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_location "github.com/manabie-com/backend/mock/mastermgmt/modules/location/infrastructure/repo"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestStaffService_UpdateStaff(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tx := new(mock_database.Tx)
	userRepo := new(mock_repositories.MockUserRepo)
	schoolAdminRepo := new(mock_repositories.MockSchoolAdminRepo)
	firebaseClient := new(mock_firebase.AuthClient)
	firebaseAuthClient := new(mock_multitenant.TenantClient)
	tenantManager := new(mock_multitenant.TenantManager)
	organizationRepo := new(mock_repositories.OrganizationRepo)
	teacherRepo := new(mock_repositories.MockTeacherRepo)
	staffRepo := new(mock_repositories.MockStaffRepo)
	userGroupRepo := new(mock_repositories.MockUserGroupRepo)
	userGroupV2Repo := new(mock_repositories.MockUserGroupV2Repo)
	userGroupsMemberRepo := new(mock_repositories.MockUserGroupsMemberRepo)
	jsm := new(mock_nats.JetStreamManagement)
	defaultResourcePath := fmt.Sprint(constants.ManabieSchool)
	usrEmailRepo := new(mock_repositories.MockUsrEmailRepo)
	userPhoneNumberRepo := new(mock_repositories.MockUserPhoneNumberRepo)
	roleRepo := new(mock_repositories.MockRoleRepo)
	userAccessPathRepo := new(mock_repositories.MockUserAccessPathRepo)
	locationRepo := new(mock_location.MockLocationRepo)
	domainUserRepo := new(mock_repositories.MockDomainUserRepo)
	domainRoleRepo := new(mock_repositories.MockDomainRoleRepo)
	unleashClient := new(mock_unleash_client.UnleashClientInstance)
	domainTagRepo := new(mock_repositories.MockDomainTagRepo)
	domainTaggedUserRepo := new(mock_repositories.MockDomainTaggedUserRepo)

	existingUser := &entity.LegacyUser{
		ID:          database.Text(idutil.ULIDNow()),
		FullName:    database.Text("John"),
		Group:       database.Text(constant.UserGroupTeacher),
		Email:       database.Text("existing-staff-email@example.com"),
		PhoneNumber: database.Text("existing-staff-phone-number"),
	}

	existingUserExternal := entity.Users{
		entity_mock.User{
			RandomUser: entity_mock.RandomUser{
				Email:          field.NewString("existing-staff-email@example.com"),
				ExternalUserID: field.NewString("existing-external_user_id"),
			},
		},
	}

	roleSchoolAdmin := repository.NewNullRole()
	roleSchoolAdmin.RoleAttribute.RoleName = field.NewString(constant.RoleSchoolAdmin)

	roleTeacher := repository.NewNullRole()
	roleTeacher.RoleAttribute.RoleName = field.NewString(constant.RoleTeacher)

	newStaffUpdateProfile := &entity.Staff{
		ID:           database.Text(idutil.ULIDNow()),
		ResourcePath: database.Text(fmt.Sprint(constants.ManabieSchool)),
		LegacyUser: entity.LegacyUser{
			Email:        database.Text("update-staff-email@example.com"),
			ResourcePath: database.Text(fmt.Sprint(constants.ManabieSchool)),
		},
	}
	umsvc := &usvc.UserModifierService{
		DB:                   tx,
		UserRepo:             userRepo,
		UserPhoneNumberRepo:  userPhoneNumberRepo,
		UsrEmailRepo:         usrEmailRepo,
		SchoolAdminRepo:      schoolAdminRepo,
		TeacherRepo:          teacherRepo,
		FirebaseClient:       firebaseClient,
		TenantManager:        tenantManager,
		OrganizationRepo:     organizationRepo,
		LocationRepo:         locationRepo,
		JSM:                  jsm,
		DomainTagRepo:        domainTagRepo,
		DomainTaggedUserRepo: domainTaggedUserRepo,
	}

	staffService := &StaffService{
		DB:                  umsvc.DB,
		UnleashClient:       unleashClient,
		JSM:                 umsvc.JSM,
		FirebaseClient:      umsvc.FirebaseClient,
		FirebaseAuthClient:  firebaseAuthClient,
		TenantManager:       tenantManager,
		UserModifierService: umsvc,
		UserGroupV2Service: &ugs.UserGroupService{
			UserGroupV2Repo:      userGroupV2Repo,
			UserGroupsMemberRepo: userGroupsMemberRepo,
			RoleRepo:             roleRepo,
		},
		SchoolAdminRepo:     schoolAdminRepo,
		TeacherRepo:         teacherRepo,
		StaffRepo:           staffRepo,
		UserGroupRepo:       userGroupRepo,
		UserAccessPathRepo:  userAccessPathRepo,
		UserPhoneNumberRepo: userPhoneNumberRepo,
		UserRepo:            domainUserRepo,
		RoleRepo:            domainRoleRepo,
		DomainUser: &ums.DomainUser{
			DB:       umsvc.DB,
			UserRepo: domainUserRepo,
		},
	}
	type params struct {
		resourcePath string
	}
	testCases := []TestCase{
		{
			name:    "invalid profile: missing name",
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:  "",
					Email: existingUser.Email.String,
				},
			},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
			},
			expectedErr: status.Error(codes.InvalidArgument, errcode.ErrUserFullNameIsEmpty.Error()),
		},
		{
			name:    "invalid profile: end date before start date",
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:      "Tuan",
					Email:     existingUser.Email.String,
					StartDate: timestamppb.New(time.Now()),
					EndDate:   timestamppb.New(time.Now().Add(-10 * time.Hour)),
					Username:  "username",
				},
			},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"username"}).Return(entity.Users{}, nil)
			},
			expectedErr: status.Error(codes.InvalidArgument, errcode.ErrStaffStartDateIsLessThanEndDate.Error()),
		},
		{
			name:    "invalid profile: missing email",
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:     existingUser.GetName(),
					Email:    "",
					Username: "username",
				},
			},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"username"}).Return(entity.Users{}, nil)
			},
			expectedErr: status.Error(codes.InvalidArgument, errcode.ErrUserEmailIsEmpty.Error()),
		},
		{
			name:    "invalid profile: missing location ids",
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:     existingUser.GetName(),
					Email:    existingUser.Email.String,
					Username: "username",
				},
			},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"username"}).Return(entity.Users{}, nil)
			},
			expectedErr: status.Error(codes.InvalidArgument, errcode.ErrUserLocationIsEmpty.Error()),
		},
		{
			name:    "err getting staff",
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:        "Name",
					Email:       "existing-staff-email@example.com",
					StaffId:     existingUser.GetUID() + "-",
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    "username",
				},
			},
			expectedErr: status.Error(codes.Internal, errors.Wrap(fmt.Errorf("err"), "s.StaffRepo.FindByID").Error()),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"username"}).Return(entity.Users{}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				domainUserRepo.On("GetUserRoles", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				staffRepo.On("FindByID", ctx, tx, mock.Anything).Once().Return(nil, fmt.Errorf("err"))
				domainUserRepo.On("GetByIDs", ctx, tx, mock.Anything).Once().Return(existingUserExternal, nil)
			},
		},
		{
			name:    "staff need to be update not found in db",
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:        "Name",
					Email:       "existing-staff-email@example.com",
					StaffId:     existingUser.GetUID(),
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    "username",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.Errorf("staff with id %s not found", existingUser.GetUID()).Error()),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"username"}).Return(entity.Users{}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				domainUserRepo.On("GetUserRoles", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				staffRepo.On("FindByID", ctx, tx, mock.Anything).Once().Return(nil, nil)
				domainUserRepo.On("GetByIDs", ctx, tx, mock.Anything).Once().Return(existingUserExternal, nil)
			},
		},
		{
			name:    "check permission when update staff",
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:        "Name",
					Email:       "existing-staff-email@example.com",
					StaffId:     existingUser.GetUID() + "-",
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    "username",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.Wrap(fmt.Errorf("get user role failed"), "checkPermissionToAssignUserGroup").Error()),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"username"}).Return(entity.Users{}, nil)
				domainUserRepo.On("GetByIDs", ctx, tx, mock.Anything).Once().Return(existingUserExternal, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, fmt.Errorf("get user role failed"))
			},
		},
		{
			name:    "s.UserModifierService.UserRepo.UpdateProfileV1: user update email error",
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:        "Name",
					Email:       "existing-staff-email@example.com",
					StaffId:     existingUser.GetUID() + "-",
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    "username",
				},
			},
			expectedErr: status.Error(
				codes.Internal,
				errors.Wrap(
					fmt.Errorf("error"),
					"s.UserModifierService.UsrEmailRepo.UpdateEmail",
				).Error(),
			),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"username"}).Return(entity.Users{}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				domainUserRepo.On("GetUserRoles", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				staffRepo.On("FindByID", ctx, tx, mock.Anything).Once().Return(newStaffUpdateProfile, nil)
				domainUserRepo.On("GetByIDs", ctx, tx, mock.Anything).Once().Return(existingUserExternal, nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{
					createMockDomainTagWithTypeAndIsArchived("tag_id_1", pb.UserTagType_USER_TAG_TYPE_STAFF, false),
					createMockDomainTagWithTypeAndIsArchived("tag_id_2", pb.UserTagType_USER_TAG_TYPE_PARENT, false),
				}, nil)
				userGroupsMemberRepo.On("GetByUserID", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupMember{
					{
						UserID:      database.Text(existingUser.GetUID() + "-"),
						UserGroupID: database.Text(idutil.ULIDNow()),
					},
				}, nil)

				tx.On("Begin", mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]*domain.Location{
						{LocationID: idutil.ULIDNow(), ResourcePath: defaultResourcePath},
						{LocationID: idutil.ULIDNow(), ResourcePath: defaultResourcePath},
					},
					nil,
				)
				tx.On("Commit", mock.Anything).Once().Return(nil)

				tx.On("Begin", mock.Anything).Once().Return(tx, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:    "err when getting user group ids",
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:         existingUser.GetName(),
					Email:        existingUser.Email.String,
					StaffId:      existingUser.GetUID() + "-",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					LocationIds:  []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:     "username",
				},
			},
			expectedErr: status.Error(
				codes.Internal,
				errors.Wrap(
					fmt.Errorf("err when getting user group ids"),
					"UserGroupV2Repo.FindByIDs",
				).Error(),
			),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"username"}).Return(entity.Users{}, nil)
				staffRepo.On("FindByID", ctx, tx, mock.Anything).Once().Return(newStaffUpdateProfile, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return(nil, fmt.Errorf("err when getting user group ids"))
			},
		},
		{
			name:    "err when upserting user group member",
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:         existingUser.GetName(),
					Email:        existingUser.Email.String,
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					LocationIds:  []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:     "username",
				},
			},
			expectedErr: status.Error(
				codes.Internal,
				errors.Wrapf(
					fmt.Errorf("error"),
					"s.UserGroupService.UserGroupsMemberRepo.UpsertBatch",
				).Error(),
			),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"username"}).Return(entity.Users{}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				staffRepo.On("FindByID", ctx, tx, mock.Anything).Once().Return(newStaffUpdateProfile, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetByIDs", ctx, tx, mock.Anything).Once().Return(existingUserExternal, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				userGroupsMemberRepo.On("GetByUserID", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupMember{
					{
						UserID:      database.Text(existingUser.GetUID() + "-"),
						UserGroupID: database.Text(idutil.ULIDNow()),
					},
				}, nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{
					createMockDomainTagWithTypeAndIsArchived("tag_id_1", pb.UserTagType_USER_TAG_TYPE_STAFF, false),
					createMockDomainTagWithTypeAndIsArchived("tag_id_2", pb.UserTagType_USER_TAG_TYPE_PARENT, false),
				}, nil)
				tx.On("Begin", mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]*domain.Location{
						{LocationID: idutil.ULIDNow(), ResourcePath: defaultResourcePath},
						{LocationID: idutil.ULIDNow(), ResourcePath: defaultResourcePath},
					},
					nil,
				)
				tx.On("Commit", mock.Anything).Once().Return(nil)

				tx.On("Begin", mock.Anything).Once().Return(tx, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, mock.Anything, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(firebaseAuthClient, nil)
				staffRepo.On("Update", ctx, tx, mock.Anything, mock.Anything).Once().Return(&entity.Staff{}, nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(fmt.Errorf("error"))
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:    "Err when UserPhoneNumberRepo.Upsert fail",
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:         existingUser.GetName(),
					Email:        existingUser.Email.String,
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					LocationIds:  []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:     "username",
				},
			},
			expectedErr: status.Error(codes.Unknown, "error from Upsert"),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"username"}).Return(entity.Users{}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				staffRepo.On("FindByID", ctx, tx, mock.Anything).Once().Return(newStaffUpdateProfile, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetByIDs", ctx, tx, mock.Anything).Once().Return(existingUserExternal, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				userGroupsMemberRepo.On("GetByUserID", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupMember{
					{
						UserID:      database.Text(existingUser.GetUID() + "-"),
						UserGroupID: database.Text(idutil.ULIDNow()),
					},
				}, nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{
					createMockDomainTagWithTypeAndIsArchived("tag_id_1", pb.UserTagType_USER_TAG_TYPE_STAFF, false),
					createMockDomainTagWithTypeAndIsArchived("tag_id_2", pb.UserTagType_USER_TAG_TYPE_PARENT, false),
				}, nil)
				tx.On("Begin", mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]*domain.Location{
						{LocationID: idutil.ULIDNow(), ResourcePath: defaultResourcePath},
						{LocationID: idutil.ULIDNow(), ResourcePath: defaultResourcePath},
					},
					nil,
				)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Begin", mock.Anything).Once().Return(tx, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, mock.Anything, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(firebaseAuthClient, nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("UpdateOrigin", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupRepo.On("UpdateStatus", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				staffRepo.On("Update", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(&entity.Staff{}, nil)
				teacherRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(fmt.Errorf("error from Upsert"))
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:    "Err when fail validation",
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:         existingUser.GetName(),
					Email:        existingUser.Email.String,
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StaffPhoneNumber: []*pb.StaffPhoneNumber{
						{PhoneNumber: "12345678", PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER},
						{PhoneNumber: "12345679", PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER},
					},
					Username: "username",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, errcode.ErrUserPrimaryPhoneNumberIsRedundant.Error()),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"username"}).Return(entity.Users{}, nil)
			},
		},
		{
			name:    happyCase,
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:          "Name",
					Email:         existingUser.Email.String,
					StaffId:       existingUser.GetUID() + "-",
					UserGroupIds:  []string{idutil.ULIDNow(), idutil.ULIDNow()},
					LocationIds:   []string{idutil.ULIDNow(), idutil.ULIDNow()},
					WorkingStatus: pb.StaffWorkingStatus_AVAILABLE,
					StartDate:     timestamppb.New(time.Now()),
					EndDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
					Remarks:       "hello",
					TagIds:        []string{"tag_id_1"},
					Username:      "username",
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"username"}).Return(entity.Users{}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				staffRepo.On("FindByID", ctx, tx, mock.Anything).Once().Return(newStaffUpdateProfile, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetByIDs", ctx, tx, mock.Anything).Once().Return(existingUserExternal, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				userGroupsMemberRepo.On("GetByUserID", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupMember{
					{
						UserID:      database.Text(existingUser.GetUID() + "-"),
						UserGroupID: database.Text(idutil.ULIDNow()),
					},
				}, nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{
					createMockDomainTagWithTypeAndIsArchived("tag_id_1", pb.UserTagType_USER_TAG_TYPE_STAFF, false),
				}, nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{
					createMockDomainTagWithTypeAndIsArchived("tag_id_1", pb.UserTagType_USER_TAG_TYPE_STAFF, false),
				}, nil)

				tx.On("Begin", mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]*domain.Location{
						{LocationID: idutil.ULIDNow(), ResourcePath: defaultResourcePath},
						{LocationID: idutil.ULIDNow(), ResourcePath: defaultResourcePath},
					},
					nil,
				)
				tx.On("Commit", mock.Anything).Once().Return(nil)

				tx.On("Begin", mock.Anything).Once().Return(tx, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, mock.Anything, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(firebaseAuthClient, nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
				domainTaggedUserRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("UpdateOrigin", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupRepo.On("UpdateStatus", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				staffRepo.On("Update", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(&entity.Staff{}, nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectUpsertStaff, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name: "error when passing unexisted tag ids",
			ctx:  ctx,
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:             "Staff Name",
					Email:            "sample-staff-email@example.com",
					UserGroupIds:     []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StaffPhoneNumber: []*pb.StaffPhoneNumber{},
					LocationIds:      []string{idutil.ULIDNow(), idutil.ULIDNow()},
					TagIds:           []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:         "username",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, ErrTagIDsMustBeExisted.Error()),
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"username"}).Return(entity.Users{}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
			},
		},
		{
			name: "error when passing wrong type tag ids",
			ctx:  ctx,
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:             "Staff Name",
					Email:            "sample-staff-email@example.com",
					UserGroupIds:     []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StaffPhoneNumber: []*pb.StaffPhoneNumber{},
					LocationIds:      []string{idutil.ULIDNow(), idutil.ULIDNow()},
					TagIds:           []string{"tag_id_1", "tag_id_2"},
					Username:         "username",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, ErrTagIsNotForStaff.Error()),
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"username"}).Return(entity.Users{}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{
					createMockDomainTagWithTypeAndIsArchived("tag_id_1", pb.UserTagType_USER_TAG_TYPE_STAFF, false),
					createMockDomainTagWithTypeAndIsArchived("tag_id_2", pb.UserTagType_USER_TAG_TYPE_PARENT, false),
				}, nil)
			},
		},
		{
			name:    "error publish event",
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:          "Name",
					Email:         existingUser.Email.String,
					StaffId:       existingUser.GetUID() + "-",
					UserGroupIds:  []string{idutil.ULIDNow(), idutil.ULIDNow()},
					LocationIds:   []string{idutil.ULIDNow(), idutil.ULIDNow()},
					WorkingStatus: pb.StaffWorkingStatus_AVAILABLE,
					StartDate:     timestamppb.New(time.Now()),
					EndDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
					Remarks:       "hello",
					Username:      "username",
				},
			},
			expectedErr: status.Error(codes.Unknown, "publishUpsertStaffEvent error: error publish event"),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"username"}).Return(entity.Users{}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				staffRepo.On("FindByID", ctx, tx, mock.Anything).Once().Return(newStaffUpdateProfile, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleTeacher}, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				userGroupsMemberRepo.On("GetByUserID", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupMember{
					{
						UserID:      database.Text(existingUser.GetUID() + "-"),
						UserGroupID: database.Text(idutil.ULIDNow()),
					},
				}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				tx.On("Begin", mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]*domain.Location{
						{LocationID: idutil.ULIDNow(), ResourcePath: defaultResourcePath},
						{LocationID: idutil.ULIDNow(), ResourcePath: defaultResourcePath},
					},
					nil,
				)
				tx.On("Commit", mock.Anything).Once().Return(nil)

				tx.On("Begin", mock.Anything).Once().Return(tx, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, mock.Anything, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(firebaseAuthClient, nil)
				domainUserRepo.On("GetByIDs", ctx, tx, mock.Anything).Once().Return(existingUserExternal, nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{
					createMockDomainTagWithTypeAndIsArchived("tag_id_1", pb.UserTagType_USER_TAG_TYPE_STAFF, false),
					createMockDomainTagWithTypeAndIsArchived("tag_id_2", pb.UserTagType_USER_TAG_TYPE_STAFF, false),
				}, nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				domainTaggedUserRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("UpdateOrigin", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupRepo.On("UpdateStatus", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				staffRepo.On("Update", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(&entity.Staff{}, nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectUpsertStaff, mock.Anything).Once().Return(&nats.PubAck{}, errors.New("error publish event"))
			},
		},
		{
			name:    "happyCase: update to non existed external_user_id",
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:           "Name",
					Email:          existingUser.Email.String,
					StaffId:        existingUser.GetUID() + "-",
					UserGroupIds:   []string{idutil.ULIDNow(), idutil.ULIDNow()},
					LocationIds:    []string{idutil.ULIDNow(), idutil.ULIDNow()},
					WorkingStatus:  pb.StaffWorkingStatus_AVAILABLE,
					StartDate:      timestamppb.New(time.Now()),
					EndDate:        timestamppb.New(time.Now().Add(1 * time.Hour)),
					Remarks:        "hello",
					TagIds:         []string{"tag_id_1"},
					ExternalUserId: "external-user-id",
					Username:       "username",
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"username"}).Return(entity.Users{}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				staffRepo.On("FindByID", ctx, tx, mock.Anything).Once().Return(newStaffUpdateProfile, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleTeacher}, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				userGroupsMemberRepo.On("GetByUserID", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupMember{
					{
						UserID:      database.Text(existingUser.GetUID() + "-"),
						UserGroupID: database.Text(idutil.ULIDNow()),
					},
				}, nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{
					createMockDomainTagWithTypeAndIsArchived("tag_id_1", pb.UserTagType_USER_TAG_TYPE_STAFF, false),
				}, nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{
					createMockDomainTagWithTypeAndIsArchived("tag_id_1", pb.UserTagType_USER_TAG_TYPE_STAFF, false),
				}, nil)

				domainUserRepo.On("GetByExternalUserIDs", ctx, tx, mock.Anything).Once().Return(entity.Users{}, nil)
				domainUserRepo.On("GetByIDs", ctx, tx, mock.Anything).Once().Return(existingUserExternal, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(firebaseAuthClient, nil)
				firebaseAuthClient.On("GetUser", ctx, mock.Anything).Once().Return(nil, nil)
				firebaseAuthClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				tx.On("Begin", mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]*domain.Location{
						{LocationID: idutil.ULIDNow(), ResourcePath: defaultResourcePath},
						{LocationID: idutil.ULIDNow(), ResourcePath: defaultResourcePath},
					},
					nil,
				)
				tx.On("Commit", mock.Anything).Once().Return(nil)

				tx.On("Begin", mock.Anything).Once().Return(tx, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, mock.Anything, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(firebaseAuthClient, nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
				domainTaggedUserRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("UpdateOrigin", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupRepo.On("UpdateStatus", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				staffRepo.On("Update", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(&entity.Staff{}, nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectUpsertStaff, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name:    "happyCase: update to non existed external_user_id and space",
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:           "Name",
					Email:          existingUser.Email.String,
					StaffId:        existingUser.GetUID() + "-",
					UserGroupIds:   []string{idutil.ULIDNow(), idutil.ULIDNow()},
					LocationIds:    []string{idutil.ULIDNow(), idutil.ULIDNow()},
					WorkingStatus:  pb.StaffWorkingStatus_AVAILABLE,
					StartDate:      timestamppb.New(time.Now()),
					EndDate:        timestamppb.New(time.Now().Add(1 * time.Hour)),
					Remarks:        "hello",
					TagIds:         []string{"tag_id_1"},
					ExternalUserId: " external-user-id ",
					Username:       "username",
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"username"}).Return(entity.Users{}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				staffRepo.On("FindByID", ctx, tx, mock.Anything).Once().Return(newStaffUpdateProfile, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleTeacher}, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				userGroupsMemberRepo.On("GetByUserID", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupMember{
					{
						UserID:      database.Text(existingUser.GetUID() + "-"),
						UserGroupID: database.Text(idutil.ULIDNow()),
					},
				}, nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{
					createMockDomainTagWithTypeAndIsArchived("tag_id_1", pb.UserTagType_USER_TAG_TYPE_STAFF, false),
				}, nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{
					createMockDomainTagWithTypeAndIsArchived("tag_id_1", pb.UserTagType_USER_TAG_TYPE_STAFF, false),
				}, nil)

				domainUserRepo.On("GetByExternalUserIDs", ctx, tx, mock.Anything).Once().Return(entity.Users{}, nil)
				domainUserRepo.On("GetByIDs", ctx, tx, mock.Anything).Once().Return(existingUserExternal, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(firebaseAuthClient, nil)
				firebaseAuthClient.On("GetUser", ctx, mock.Anything).Once().Return(nil, nil)
				firebaseAuthClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				tx.On("Begin", mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]*domain.Location{
						{LocationID: idutil.ULIDNow(), ResourcePath: defaultResourcePath},
						{LocationID: idutil.ULIDNow(), ResourcePath: defaultResourcePath},
					},
					nil,
				)
				tx.On("Commit", mock.Anything).Once().Return(nil)

				tx.On("Begin", mock.Anything).Once().Return(tx, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, mock.Anything, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(firebaseAuthClient, nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
				domainTaggedUserRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("UpdateOrigin", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupRepo.On("UpdateStatus", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				staffRepo.On("Update", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(&entity.Staff{}, nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectUpsertStaff, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name:    "update to existed external_user_id",
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:           "Name",
					Email:          existingUser.Email.String,
					StaffId:        existingUser.GetUID() + "-",
					UserGroupIds:   []string{idutil.ULIDNow(), idutil.ULIDNow()},
					LocationIds:    []string{idutil.ULIDNow(), idutil.ULIDNow()},
					WorkingStatus:  pb.StaffWorkingStatus_AVAILABLE,
					StartDate:      timestamppb.New(time.Now()),
					EndDate:        timestamppb.New(time.Now().Add(1 * time.Hour)),
					Remarks:        "hello",
					TagIds:         []string{"tag_id_1"},
					ExternalUserId: existingUserExternal[0].ExternalUserID().String(),
					Username:       "username",
				},
			},
			expectedErr: status.Error(
				codes.AlreadyExists,
				"s.DomainUser.validateExternalUserIDExistedInSystem: existing data in field 'users[0].external_user_id' in entity 'staff' at index 0",
			),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"username"}).Return(entity.Users{}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				staffRepo.On("FindByID", ctx, tx, mock.Anything).Once().Return(newStaffUpdateProfile, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetByExternalUserIDs", ctx, tx, mock.Anything).Once().Return(existingUserExternal, nil)
				domainUserRepo.On("GetByIDs", ctx, tx, mock.Anything).Once().Return(entity.Users{}, nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{
					createMockDomainTagWithTypeAndIsArchived("tag_id_1", pb.UserTagType_USER_TAG_TYPE_STAFF, false),
				}, nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{
					createMockDomainTagWithTypeAndIsArchived("tag_id_1", pb.UserTagType_USER_TAG_TYPE_STAFF, false),
				}, nil)
			},
		},
		{
			name:    "update to existed external_user_id and space",
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:           "Name",
					Email:          existingUser.Email.String,
					StaffId:        existingUser.GetUID() + "-",
					UserGroupIds:   []string{idutil.ULIDNow(), idutil.ULIDNow()},
					LocationIds:    []string{idutil.ULIDNow(), idutil.ULIDNow()},
					WorkingStatus:  pb.StaffWorkingStatus_AVAILABLE,
					StartDate:      timestamppb.New(time.Now()),
					EndDate:        timestamppb.New(time.Now().Add(1 * time.Hour)),
					Remarks:        "hello",
					TagIds:         []string{"tag_id_1"},
					ExternalUserId: " " + existingUserExternal[0].ExternalUserID().String() + " ",
					Username:       "username",
				},
			},
			expectedErr: status.Error(
				codes.AlreadyExists,
				"s.DomainUser.validateExternalUserIDExistedInSystem: existing data in field 'users[0].external_user_id' in entity 'staff' at index 0",
			),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"username"}).Return(entity.Users{}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				staffRepo.On("FindByID", ctx, tx, mock.Anything).Once().Return(newStaffUpdateProfile, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetByExternalUserIDs", ctx, tx, mock.Anything).Once().Return(existingUserExternal, nil)
				domainUserRepo.On("GetByIDs", ctx, tx, mock.Anything).Once().Return(entity.Users{}, nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{
					createMockDomainTagWithTypeAndIsArchived("tag_id_1", pb.UserTagType_USER_TAG_TYPE_STAFF, false),
				}, nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{
					createMockDomainTagWithTypeAndIsArchived("tag_id_1", pb.UserTagType_USER_TAG_TYPE_STAFF, false),
				}, nil)
			},
		},
		{
			name:    "happy case: update staff with valid username with email format",
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:          "Name",
					Email:         existingUser.Email.String,
					StaffId:       existingUser.GetUID() + "-",
					UserGroupIds:  []string{idutil.ULIDNow(), idutil.ULIDNow()},
					LocationIds:   []string{idutil.ULIDNow(), idutil.ULIDNow()},
					WorkingStatus: pb.StaffWorkingStatus_AVAILABLE,
					StartDate:     timestamppb.New(time.Now()),
					EndDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
					Remarks:       "hello",
					TagIds:        []string{"tag_id_1"},
					Username:      "Valid_Username@gmail.com",
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"valid_username@gmail.com"}).Return(entity.Users{}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				staffRepo.On("FindByID", ctx, tx, mock.Anything).Once().Return(newStaffUpdateProfile, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, tx, mock.Anything).Once().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetByIDs", ctx, tx, mock.Anything).Once().Return(existingUserExternal, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				userGroupsMemberRepo.On("GetByUserID", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupMember{
					{
						UserID:      database.Text(existingUser.GetUID() + "-"),
						UserGroupID: database.Text(idutil.ULIDNow()),
					},
				}, nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{
					createMockDomainTagWithTypeAndIsArchived("tag_id_1", pb.UserTagType_USER_TAG_TYPE_STAFF, false),
				}, nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{
					createMockDomainTagWithTypeAndIsArchived("tag_id_1", pb.UserTagType_USER_TAG_TYPE_STAFF, false),
				}, nil)

				tx.On("Begin", mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]*domain.Location{
						{LocationID: idutil.ULIDNow(), ResourcePath: defaultResourcePath},
						{LocationID: idutil.ULIDNow(), ResourcePath: defaultResourcePath},
					},
					nil,
				)
				tx.On("Commit", mock.Anything).Once().Return(nil)

				tx.On("Begin", mock.Anything).Once().Return(tx, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, mock.Anything, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(firebaseAuthClient, nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("SoftDelete", ctx, tx, mock.Anything).Once().Return(nil)
				domainTaggedUserRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("UpdateOrigin", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupRepo.On("UpdateStatus", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				staffRepo.On("Update", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(&entity.Staff{}, nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectUpsertStaff, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name:    "invalid profile: missing username",
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:     existingUser.GetName(),
					Email:    existingUser.Email.String,
					Username: "",
				},
			},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
			},
			expectedErr: status.Error(codes.InvalidArgument, entity.MissingMandatoryFieldError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      -1,
			}.Error()),
		},
		{
			name:    "invalid profile: existed username",
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:     existingUser.GetName(),
					Email:    existingUser.Email.String,
					Username: "ExistedUsername",
				},
			},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"existedusername"}).Return(entity.Users{
					entity_mock.User{
						RandomUser: entity_mock.RandomUser{
							UserID:   field.NewString("user-id"),
							UserName: field.NewString("existedusername"),
						},
					},
				}, nil)
			},
			expectedErr: status.Error(codes.AlreadyExists, entity.ExistingDataError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      0,
			}.Error()),
		},
		{
			name:    "invalid profile: invalid username",
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:     existingUser.GetName(),
					Email:    existingUser.Email.String,
					Username: "invalid_username",
				},
			},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
			},
			expectedErr: status.Error(codes.InvalidArgument, entity.InvalidFieldError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      -1,
				Reason:     entity.NotMatchingPattern,
			}.Error()),
		},
		{
			name:    "invalid profile: invalid username with spaces",
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:     existingUser.GetName(),
					Email:    existingUser.Email.String,
					Username: "invalid username",
				},
			},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
			},
			expectedErr: status.Error(codes.InvalidArgument, entity.InvalidFieldError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      -1,
				Reason:     entity.NotMatchingPattern,
			}.Error()),
		},
		{
			name:    "invalid profile: invalid username with wrong email format",
			ctx:     ctx,
			Options: params{resourcePath: defaultResourcePath},
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					Name:     existingUser.GetName(),
					Email:    existingUser.Email.String,
					Username: "invalid_username@manabie",
				},
			},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
			},
			expectedErr: status.Error(codes.InvalidArgument, entity.InvalidFieldError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      -1,
				Reason:     entity.NotMatchingPattern,
			}.Error()),
		},
	}

	for _, testCase := range testCases {
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

			t.Log("Test case: " + testCase.name)
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}

			_, err := staffService.UpdateStaff(testCase.ctx, testCase.req.(*pb.UpdateStaffRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
	mock.AssertExpectationsForObjects(t, userRepo, userPhoneNumberRepo, schoolAdminRepo, teacherRepo, userGroupsMemberRepo, userGroupV2Repo, jsm, tx, roleRepo)
}

func TestStaffService_UpdateStaffEmail(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tx := new(mock_database.Tx)

	userRepo := new(mock_repositories.MockUserRepo)
	schoolAdminRepo := new(mock_repositories.MockSchoolAdminRepo)
	organizationRepo := new(mock_repositories.OrganizationRepo)
	firebaseClient := new(mock_firebase.AuthClient)
	firebaseAuthClient := new(mock_multitenant.TenantClient)
	tenantManager := new(mock_multitenant.TenantManager)
	teacherRepo := new(mock_repositories.MockTeacherRepo)
	jsm := new(mock_nats.JetStreamManagement)
	usrEmailRepo := new(mock_repositories.MockUsrEmailRepo)

	umsvc := &usvc.UserModifierService{
		DB:                 tx,
		UserRepo:           userRepo,
		SchoolAdminRepo:    schoolAdminRepo,
		TeacherRepo:        teacherRepo,
		UsrEmailRepo:       usrEmailRepo,
		FirebaseClient:     firebaseClient,
		FirebaseAuthClient: firebaseAuthClient,
		OrganizationRepo:   organizationRepo,
		TenantManager:      tenantManager,
		JSM:                jsm,
	}

	s := &StaffService{
		DB:                  umsvc.DB,
		JSM:                 umsvc.JSM,
		FirebaseClient:      umsvc.FirebaseClient,
		FirebaseAuthClient:  umsvc.FirebaseAuthClient,
		UserModifierService: umsvc,
	}
	type params struct {
		profile      *entity.LegacyUser
		resourcePath string
	}

	existingUser := &entity.LegacyUser{
		ID:          database.Text(idutil.ULIDNow()),
		Email:       database.Text("existing-student-email@example.com"),
		PhoneNumber: database.Text("existing-student-phone-number"),
	}

	testCases := []TestCase{
		{
			name: "err when find by email",
			ctx:  ctx,
			req:  nil,
			Options: params{
				profile:      &entity.LegacyUser{},
				resourcePath: "1",
			},
			setup: func(ctx context.Context) {
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return(nil, fmt.Errorf("err when find by email"))
			},
			expectedErr: status.Error(codes.Internal, errors.Wrap(fmt.Errorf("err when find by email"), "s.UserModifierService.UserRepo.GetByEmailInsensitiveCase").Error()),
		},
		{
			name: "email is already exists",
			ctx:  ctx,
			req:  nil,
			Options: params{
				profile: &entity.LegacyUser{
					Email: database.Text(existingUser.Email.String),
				},
				resourcePath: "1",
			},
			setup: func(ctx context.Context) {
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{existingUser}, nil)
			},
			expectedErr: status.Error(
				codes.AlreadyExists,
				errcode.ErrUserEmailExists.Error(),
			),
		},
		{
			name: "err when update email",
			ctx:  ctx,
			req:  nil,
			Options: params{
				profile: &entity.LegacyUser{
					Email: database.Text(existingUser.Email.String),
				},
				resourcePath: "1",
			},
			setup: func(ctx context.Context) {
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("err when update email"))

			},
			expectedErr: status.Error(codes.Internal, errors.Wrap(fmt.Errorf("err when update email"), "s.UserModifierService.UsrEmailRepo.UpdateEmail").Error()),
		},
		{
			name: happyCase,
			ctx:  ctx,
			req:  nil,
			Options: params{
				profile: &entity.LegacyUser{
					Email: database.Text(existingUser.Email.String),
				},
				resourcePath: "1",
			},
			setup: func(ctx context.Context) {
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, mock.Anything, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(firebaseAuthClient, nil)
				firebaseAuthClient.On("GetUser", ctx, mock.Anything).Once().Return(nil, nil)
				firebaseAuthClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			expectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			t.Log("Test case: " + testCase.name)

			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: testCase.Options.(params).resourcePath,
				},
			}

			testCase.ctx = interceptors.ContextWithUserID(
				interceptors.ContextWithJWTClaims(testCase.ctx, claim),
				existingUser.ID.String,
			)

			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}

			err := s.updateStaffEmail(testCase.ctx, tx, testCase.Options.(params).profile, existingUser.Email.String)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			}

			mock.AssertExpectationsForObjects(t, userRepo, firebaseAuthClient, organizationRepo, tenantManager)
		})
	}
}

func TestStaffService_CheckPermissionUpdateUser(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tx := new(mock_database.Tx)
	userRepo := new(mock_repositories.MockUserRepo)
	schoolAdminRepo := new(mock_repositories.MockSchoolAdminRepo)
	teacherRepo := new(mock_repositories.MockTeacherRepo)
	firebaseClient := new(mock_firebase.AuthClient)

	umsvc := &usvc.UserModifierService{
		DB:              tx,
		UserRepo:        userRepo,
		SchoolAdminRepo: schoolAdminRepo,
		TeacherRepo:     teacherRepo,
		FirebaseClient:  firebaseClient,
	}

	s := &StaffService{
		DB:                  umsvc.DB,
		FirebaseClient:      umsvc.FirebaseClient,
		UserModifierService: umsvc,
	}
	type params struct {
		currentUserID string
		currentUGroup string
	}

	uid := idutil.ULIDNow()
	existingUser := &entity.LegacyUser{
		ID:          database.Text(uid),
		Email:       database.Text(uid + "-email@example.com"),
		PhoneNumber: database.Text(uid + "-existing-phone-number"),
	}

	testCases := []TestCase{
		{
			name:        happyCase,
			ctx:         ctx,
			req:         idutil.ULIDNow(),
			expectedErr: nil,
			Options: params{
				currentUserID: existingUser.GetUID(),
				currentUGroup: entity.UserGroupSchoolAdmin,
			},
		},
		{
			name:        "check permission update staff, staff was not your staff",
			ctx:         ctx,
			req:         existingUser.GetUID(),
			expectedErr: status.Error(codes.PermissionDenied, "school admin can only update their staff profile"),
			Options: params{
				currentUserID: existingUser.GetUID(),
				currentUGroup: entity.UserGroupSchoolAdmin,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{ResourcePath: fmt.Sprint(constants.ManabieSchool)},
			}

			testCase.ctx = interceptors.ContextWithUserID(
				interceptors.ContextWithJWTClaims(ctx, claim),
				existingUser.GetUID(),
			)

			t.Log(testCaseLog + testCase.name)

			err := s.checkPermissionUpdateStaff(
				testCase.req.(string),
				testCase.Options.(params).currentUserID,
			)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			}
		})
	}
}

func TestStaffService_validationsUpdateStaff(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tx := new(mock_database.Tx)

	userGroupV2Repo := new(mock_repositories.MockUserGroupV2Repo)
	unleashClient := new(mock_unleash_client.UnleashClientInstance)
	userRepo := new(mock_repositories.MockUserRepo)
	domainUserRepo := new(mock_repositories.MockDomainUserRepo)

	umsvc := &usvc.UserModifierService{
		DB:             tx,
		DomainUserRepo: domainUserRepo,
		UserRepo:       userRepo,
	}

	staffService := StaffService{
		DB: tx,
		UserGroupV2Service: &ugs.UserGroupService{
			UserGroupV2Repo: userGroupV2Repo,
		},
		UserRepo:            domainUserRepo,
		UserModifierService: umsvc,
		UnleashClient:       unleashClient,
	}
	type params struct {
		userProfile                   *pb.UpdateStaffRequest_StaffProfile
		isFeatureStaffUsernameEnabled bool
	}

	testCases := []TestCase{
		{
			name:        happyCase,
			ctx:         ctx,
			req:         nil,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"username"}).Return(entity.Users{}, nil)
			},
			Options: params{
				userProfile: &pb.UpdateStaffRequest_StaffProfile{
					StaffId:     idutil.ULIDNow(),
					Name:        "Tuan",
					Email:       "tuan@email.com",
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StartDate:   timestamppb.New(time.Now()),
					EndDate:     timestamppb.New(time.Now().Add(80 * time.Hour)),
					Username:    "username",
				},
				isFeatureStaffUsernameEnabled: true,
			},
		},
		{
			name: "name cannot be empty",
			ctx:  ctx,
			req:  nil,
			expectedErr: status.Error(
				codes.InvalidArgument,
				errcode.ErrUserFullNameIsEmpty.Error(),
			),
			Options: params{
				userProfile: &pb.UpdateStaffRequest_StaffProfile{
					StaffId:  "123",
					Username: "username",
				},
				isFeatureStaffUsernameEnabled: false,
			},
		},
		{
			name: "first name cannot be empty",
			ctx:  ctx,
			req:  nil,
			expectedErr: status.Error(
				codes.InvalidArgument,
				errcode.ErrUserFirstNameIsEmpty.Error(),
			),
			Options: params{
				userProfile: &pb.UpdateStaffRequest_StaffProfile{
					StaffId: "123",
					UserNameFields: &pb.UserNameFields{
						LastName: "John",
					},
					Username: "username",
				},
				isFeatureStaffUsernameEnabled: false,
			},
		},
		{
			name: "or last name cannot be empty",
			ctx:  ctx,
			req:  nil,
			expectedErr: status.Error(
				codes.InvalidArgument,
				errcode.ErrUserLastNameIsEmpty.Error(),
			),
			Options: params{
				userProfile: &pb.UpdateStaffRequest_StaffProfile{
					StaffId: "123",
					UserNameFields: &pb.UserNameFields{
						FirstName: "Doe",
					},
					Username: "username",
				},
				isFeatureStaffUsernameEnabled: false,
			},
		},
		{
			name: "email cannot be empty",
			ctx:  ctx,
			req:  nil,
			expectedErr: status.Error(
				codes.InvalidArgument,
				errcode.ErrUserEmailIsEmpty.Error(),
			),
			Options: params{
				userProfile: &pb.UpdateStaffRequest_StaffProfile{
					StaffId:  idutil.ULIDNow(),
					Name:     "Tuan",
					Username: "username",
				},
				isFeatureStaffUsernameEnabled: false,
			},
		},
		{
			name: "end date cannot before start date",
			ctx:  ctx,
			req:  nil,
			expectedErr: status.Error(
				codes.InvalidArgument,
				errcode.ErrStaffStartDateIsLessThanEndDate.Error(),
			),
			Options: params{
				userProfile: &pb.UpdateStaffRequest_StaffProfile{
					StaffId:   idutil.ULIDNow(),
					Name:      "Tuan",
					Email:     "tuan@email.com",
					StartDate: timestamppb.New(time.Now()),
					EndDate:   timestamppb.New(time.Now().Add(-1 * time.Hour)),
					Username:  "username",
				},
				isFeatureStaffUsernameEnabled: false,
			},
		},
		{
			name: "username cannot be empty",
			ctx:  ctx,
			req:  nil,
			expectedErr: status.Error(codes.InvalidArgument, entity.MissingMandatoryFieldError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      -1,
			}.Error()),
			Options: params{
				userProfile: &pb.UpdateStaffRequest_StaffProfile{
					StaffId:     idutil.ULIDNow(),
					Name:        "Tuan",
					Email:       "tuan@email.com",
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StartDate:   timestamppb.New(time.Now()),
					EndDate:     timestamppb.New(time.Now().Add(80 * time.Hour)),
					Username:    "",
				},
				isFeatureStaffUsernameEnabled: true,
			},
		},
		{
			name: "validation create user username is exist in DB",
			ctx:  ctx,
			req:  nil,
			expectedErr: status.Error(codes.AlreadyExists, entity.ExistingDataError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      0,
			}.Error()),
			setup: func(ctx context.Context) {
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"existedusername"}).Return(entity.Users{
					entity_mock.User{
						RandomUser: entity_mock.RandomUser{
							UserID:   field.NewString("user-id"),
							UserName: field.NewString("existedusername"),
						},
					},
				}, nil)
			},
			Options: params{
				userProfile: &pb.UpdateStaffRequest_StaffProfile{
					StaffId:     idutil.ULIDNow(),
					Name:        "Tuan",
					Email:       "tuan@email.com",
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StartDate:   timestamppb.New(time.Now()),
					EndDate:     timestamppb.New(time.Now().Add(80 * time.Hour)),
					Username:    "existedusername",
				},
				isFeatureStaffUsernameEnabled: true,
			},
		},
		{
			name: "invalid username",
			ctx:  ctx,
			req:  nil,
			expectedErr: status.Error(codes.InvalidArgument, entity.InvalidFieldError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      -1,
				Reason:     entity.NotMatchingPattern,
			}.Error()),
			Options: params{
				userProfile: &pb.UpdateStaffRequest_StaffProfile{
					StaffId:     idutil.ULIDNow(),
					Name:        "Tuan",
					Email:       "tuan@email.com",
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StartDate:   timestamppb.New(time.Now()),
					EndDate:     timestamppb.New(time.Now().Add(80 * time.Hour)),
					Username:    "invalid_username",
				},
				isFeatureStaffUsernameEnabled: true,
			},
		},
		{
			name: "invalid username with spaces",
			ctx:  ctx,
			req:  nil,
			expectedErr: status.Error(codes.InvalidArgument, entity.InvalidFieldError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      -1,
				Reason:     entity.NotMatchingPattern,
			}.Error()),
			Options: params{
				userProfile: &pb.UpdateStaffRequest_StaffProfile{
					StaffId:     idutil.ULIDNow(),
					Name:        "Tuan",
					Email:       "tuan@email.com",
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StartDate:   timestamppb.New(time.Now()),
					EndDate:     timestamppb.New(time.Now().Add(80 * time.Hour)),
					Username:    "invalid username",
				},
				isFeatureStaffUsernameEnabled: true,
			},
		},
		{
			name: "invalid username with wrong email format",
			ctx:  ctx,
			req:  nil,
			expectedErr: status.Error(codes.InvalidArgument, entity.InvalidFieldError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      -1,
				Reason:     entity.NotMatchingPattern,
			}.Error()),
			Options: params{
				userProfile: &pb.UpdateStaffRequest_StaffProfile{
					StaffId:     idutil.ULIDNow(),
					Name:        "Tuan",
					Email:       "tuan@email.com",
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StartDate:   timestamppb.New(time.Now()),
					EndDate:     timestamppb.New(time.Now().Add(80 * time.Hour)),
					Username:    "invalid_username@manabie",
				},
				isFeatureStaffUsernameEnabled: true,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.ManabieSchool),
				},
			}

			testCase.ctx = interceptors.ContextWithJWTClaims(ctx, claim)

			t.Log(testCaseLog + testCase.name)
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}

			err := staffService.validationsUpdateStaff(
				testCase.ctx,
				testCase.Options.(params).userProfile,
				testCase.Options.(params).isFeatureStaffUsernameEnabled,
			)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func Test_updateStaffPbToStaffEnt(t *testing.T) {
	t.Parallel()
	_, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	defaultStaff := &entity.Staff{}
	database.AllNullEntity(defaultStaff)

	resourcePath := "-2147483648"
	userGroup := constant.UserGroupSchoolAdmin
	birthday := timestamppb.New(time.Now().Add(-87600 * 10 * time.Hour))
	gender := pb.Gender_MALE
	startDate := timestamppb.New(time.Now())
	endDate := timestamppb.New(time.Now().Add(87600 * time.Hour))
	userId := "staff-id"
	externalUserId := "external_user_id"
	username := "username"

	testCases := []TestCase{
		{
			name: happyCase,
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					StaffId:        userId,
					Name:           "John Doe",
					Email:          "staff_manabie@manabie.com",
					Gender:         gender,
					Birthday:       birthday,
					WorkingStatus:  pb.StaffWorkingStatus_AVAILABLE,
					StartDate:      startDate,
					EndDate:        endDate,
					Remarks:        "Staff Remarks",
					TagIds:         []string{"tag_id_1"},
					ExternalUserId: externalUserId,
					Username:       username,
				},
			},
			Options: func() {
				//staff
				defaultStaff.ResourcePath = database.Text(resourcePath)
				defaultStaff.WorkingStatus = database.Text(pb.StaffWorkingStatus_AVAILABLE.String())
				defaultStaff.StartDate = database.DateFromPb(startDate)
				defaultStaff.EndDate = database.DateFromPb(endDate)
				defaultStaff.ID = database.Text(userId)
				//user
				defaultUser := entity.LegacyUser{}
				database.AllNullEntity(&defaultUser)
				defaultUser.FullName = database.Text("John Doe")
				defaultUser.Email = database.Text("staff_manabie@manabie.com")
				defaultUser.Birthday = database.DateFromPb(birthday)
				defaultUser.Gender = database.Text(gender.String())
				defaultUser.Group = database.Text(constant.UserGroupSchoolAdmin)
				defaultUser.LastName = database.Text("John")
				defaultUser.FirstName = database.Text("Doe")
				defaultUser.ResourcePath = database.Text(resourcePath)
				defaultUser.Remarks = database.Text("Staff Remarks")
				defaultUser.ExternalUserID = database.Text(externalUserId)
				defaultUser.UserName = database.Text(username)
				// Temporarily set loginEmail equal Email
				defaultUser.LoginEmail = database.Text("staff_manabie@manabie.com")
				defaultStaff.LegacyUser = defaultUser
			},
			expectedRes: defaultStaff,
		},
		{
			name: "return correct data when providing UserNameFields",
			req: &pb.UpdateStaffRequest{
				Staff: &pb.UpdateStaffRequest_StaffProfile{
					StaffId:        userId,
					Email:          "staff_manabie@manabie.com",
					Gender:         gender,
					Birthday:       birthday,
					WorkingStatus:  pb.StaffWorkingStatus_AVAILABLE,
					StartDate:      startDate,
					EndDate:        endDate,
					Remarks:        "Staff Remarks",
					ExternalUserId: externalUserId,
					UserNameFields: &pb.UserNameFields{
						LastName:          "John",
						FirstName:         "Doe",
						LastNamePhonetic:  "LastNamePhonetic",
						FirstNamePhonetic: "FirstNamePhonetic",
					},
					Username: username,
				},
			},
			Options: func() {
				//staff
				defaultStaff.ResourcePath = database.Text(resourcePath)
				defaultStaff.WorkingStatus = database.Text(pb.StaffWorkingStatus_AVAILABLE.String())
				defaultStaff.StartDate = database.DateFromPb(startDate)
				defaultStaff.EndDate = database.DateFromPb(endDate)
				defaultStaff.ID = database.Text(userId)
				//user
				defaultUser := entity.LegacyUser{}
				database.AllNullEntity(&defaultUser)
				defaultUser.FullName = database.Text("John Doe")
				defaultUser.Email = database.Text("staff_manabie@manabie.com")
				defaultUser.Birthday = database.DateFromPb(birthday)
				defaultUser.Gender = database.Text(gender.String())
				defaultUser.Group = database.Text(constant.UserGroupSchoolAdmin)
				defaultUser.LastName = database.Text("John")
				defaultUser.FirstName = database.Text("Doe")
				defaultUser.LastNamePhonetic = database.Text("LastNamePhonetic")
				defaultUser.FirstNamePhonetic = database.Text("FirstNamePhonetic")
				defaultUser.FullNamePhonetic = database.Text("LastNamePhonetic FirstNamePhonetic")
				defaultUser.ExternalUserID = database.Text(externalUserId)
				defaultUser.UserName = database.Text(username)
				defaultUser.ResourcePath = database.Text(resourcePath)
				defaultUser.Remarks = database.Text("Staff Remarks")
				// Temporarily set loginEmail equal Email
				defaultUser.LoginEmail = database.Text("staff_manabie@manabie.com")
				defaultStaff.LegacyUser = defaultUser
			},
			expectedRes: defaultStaff,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			t.Log(testCaseLog + testCase.name)
			testCase.Options.(func())()

			staff, err := updateStaffPbToStaffEnt(testCase.req.(*pb.UpdateStaffRequest), userGroup, resourcePath)

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
			if testCase.expectedRes != nil {
				testCase.expectedRes.(*entity.Staff).UpdatedAt = staff.UpdatedAt
				assert.Equal(t, *testCase.expectedRes.(*entity.Staff), *staff)
			}
		})
	}
}
