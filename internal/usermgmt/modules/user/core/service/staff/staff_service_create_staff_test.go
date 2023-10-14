package staff

import (
	"context"
	"fmt"
	"testing"
	"time"

	internal_user "github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
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
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MockDomainTag struct {
	tagID             field.String
	partnerInternalID field.String
	tagType           field.String
	isArchived        field.Boolean

	entity.EmptyDomainTag
}

func createMockDomainTagWithTypeAndIsArchived(tagID string, tagType pb.UserTagType, isArchived bool) entity.DomainTag {
	return &MockDomainTag{
		tagID:             field.NewString(tagID),
		partnerInternalID: field.NewString(fmt.Sprintf("partner-id-%s", tagID)),
		tagType:           field.NewString(tagType.String()),
		isArchived:        field.NewBoolean(isArchived),
	}
}

func (m *MockDomainTag) TagID() field.String {
	return m.tagID
}

func (m *MockDomainTag) PartnerInternalID() field.String {
	return m.partnerInternalID
}

func (m *MockDomainTag) TagType() field.String {
	return m.tagType
}

func (m *MockDomainTag) IsArchived() field.Boolean {
	return m.isArchived
}

func TestStaffService_CreateStaff(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tx := new(mock_database.Tx)
	defaultResourcePath := "1"
	userRepo := new(mock_repositories.MockUserRepo)
	userGroupRepo := new(mock_repositories.MockUserGroupRepo)
	userPhoneNumberRepo := new(mock_repositories.MockUserPhoneNumberRepo)
	userGroupV2Repo := new(mock_repositories.MockUserGroupV2Repo)
	schoolAdminRepo := new(mock_repositories.MockSchoolAdminRepo)
	userGroupsMemberRepo := new(mock_repositories.MockUserGroupsMemberRepo)
	firebaseClient := new(mock_firebase.AuthClient)
	firebaseAuthClient := new(mock_multitenant.TenantClient)
	tenantClient := new(mock_multitenant.TenantClient)
	staffRepo := new(mock_repositories.MockStaffRepo)
	teacherRepo := new(mock_repositories.MockTeacherRepo)
	organizationRepo := new(mock_repositories.OrganizationRepo)
	tenantManager := new(mock_multitenant.TenantManager)
	usrEmailRepo := new(mock_repositories.MockUsrEmailRepo)
	roleRepo := new(mock_repositories.MockRoleRepo)
	locationRepo := new(mock_location.MockLocationRepo)
	userAccessPathRepo := new(mock_repositories.MockUserAccessPathRepo)
	jsm := new(mock_nats.JetStreamManagement)
	unleashClient := new(mock_unleash_client.UnleashClientInstance)
	domainTagRepo := new(mock_repositories.MockDomainTagRepo)
	domainTaggedUserRepo := new(mock_repositories.MockDomainTaggedUserRepo)
	domainRoleRepo := new(mock_repositories.MockDomainRoleRepo)
	domainUserRepo := new(mock_repositories.MockDomainUserRepo)

	umsvc := &usvc.UserModifierService{
		DB:                   tx,
		UserRepo:             userRepo,
		UsrEmailRepo:         usrEmailRepo,
		UserGroupRepo:        userGroupRepo,
		UserPhoneNumberRepo:  userPhoneNumberRepo,
		SchoolAdminRepo:      schoolAdminRepo,
		TenantManager:        tenantManager,
		TeacherRepo:          teacherRepo,
		FirebaseClient:       firebaseClient,
		OrganizationRepo:     organizationRepo,
		FirebaseAuthClient:   firebaseAuthClient,
		LocationRepo:         locationRepo,
		DomainTagRepo:        domainTagRepo,
		DomainTaggedUserRepo: domainTaggedUserRepo,
	}
	s := &StaffService{
		DB:                  umsvc.DB,
		UnleashClient:       unleashClient,
		FirebaseClient:      umsvc.FirebaseClient,
		FirebaseAuthClient:  umsvc.FirebaseAuthClient,
		UserModifierService: umsvc,
		UserGroupV2Service: &ugs.UserGroupService{
			UserGroupV2Repo:      userGroupV2Repo,
			UserGroupsMemberRepo: userGroupsMemberRepo,
			RoleRepo:             roleRepo,
		},
		TeacherRepo:         teacherRepo,
		UserGroupRepo:       userGroupRepo,
		UserPhoneNumberRepo: userPhoneNumberRepo,
		StaffRepo:           staffRepo,
		UserAccessPathRepo:  userAccessPathRepo,
		JSM:                 jsm,
		UserRepo:            domainUserRepo,
		RoleRepo:            domainRoleRepo,
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

	roleSchoolAdmin := repository.NewNullRole()
	roleSchoolAdmin.RoleAttribute.RoleName = field.NewString(constant.RoleSchoolAdmin)

	roleTeacher := repository.NewNullRole()
	roleTeacher.RoleAttribute.RoleName = field.NewString(constant.RoleTeacher)

	hashConfig := mockScryptHash()

	testCases := []TestCase{
		{
			name: "err when getting user group ids",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:         "Staff Name",
					UserGroup:    pb.UserGroup_USER_GROUP_TEACHER,
					Country:      cpb.Country_COUNTRY_JP,
					Email:        "sample-staff-email@example.com",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					LocationIds:  []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:     "validusername",
				},
			},
			expectedErr: status.Error(
				codes.Internal,
				errors.Wrap(
					fmt.Errorf("err when getting user group ids"),
					"UserGroupV2Repo.FindByIDs",
				).Error(),
			),
			Options: params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return(nil, fmt.Errorf("err when getting user group ids"))
			},
		},
		{
			name: "err in creating staff when creating user email",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:        "Staff Name",
					UserGroup:   pb.UserGroup_USER_GROUP_TEACHER,
					Country:     cpb.Country_COUNTRY_JP,
					Email:       "sample-staff-email@example.com",
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    "validusername",
				},
			},
			expectedErr: status.Error(
				codes.Internal,
				errors.Wrap(
					fmt.Errorf("err when creating user email"),
					"s.UserModifierService.UsrEmailRepo.Create",
				).Error(),
			),
			Options: params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, s.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				domainUserRepo.On("GetUserRoles", ctx, s.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				usrEmailRepo.On("Create", ctx, s.DB, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("err when creating user email"))
			},
		},
		{
			name: "get locations fail",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:         "Staff Name",
					UserGroup:    pb.UserGroup_USER_GROUP_TEACHER,
					Country:      cpb.Country_COUNTRY_JP,
					Email:        "sample-staff-email@example.com",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					LocationIds:  []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:     "validusername",
				},
			},
			expectedErr: status.Error(
				codes.InvalidArgument,
				errors.Wrap(
					fmt.Errorf("getLocations fail: error"),
					"UserModifierService.GetLocations",
				).Error(),
			),
			Options: params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, s.DB, mock.Anything).Twice().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				domainUserRepo.On("GetUserRoles", ctx, s.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)

				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				usrEmailRepo.On("Create", ctx, s.DB, mock.Anything, mock.Anything).Once().Return(&entity.UsrEmail{UsrID: database.Text(idutil.ULIDNow())}, nil)
				tx.On("Begin", mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					nil,
					fmt.Errorf("error"),
				)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "err in creating staff when creating staff",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:        "Staff Name",
					UserGroup:   pb.UserGroup_USER_GROUP_TEACHER,
					Country:     cpb.Country_COUNTRY_JP,
					Email:       "sample-staff-email@example.com",
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    "validusername",
				},
			},
			expectedErr: status.Error(codes.Internal, errors.Wrapf(fmt.Errorf("err when creating staff"), "s.StaffRepo.Create").Error()),
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, s.DB, mock.Anything).Twice().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				domainUserRepo.On("GetUserRoles", ctx, s.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				usrEmailRepo.On("Create", ctx, s.DB, mock.Anything, mock.Anything).Once().Return(&entity.UsrEmail{UsrID: database.Text(idutil.ULIDNow())}, nil)

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
				staffRepo.On("Create", ctx, tx, mock.Anything).Once().Return(fmt.Errorf("err when creating staff"))
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "err in creating staff when creating teacher",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:        "Staff Name",
					UserGroup:   pb.UserGroup_USER_GROUP_TEACHER,
					Country:     cpb.Country_COUNTRY_JP,
					Email:       "sample-staff-email@example.com",
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    "validusername",
				},
			},
			expectedErr: status.Error(codes.Internal, errors.Wrapf(fmt.Errorf("err when creating teacher"), "s.createTeacher: %s", pb.UserGroup_USER_GROUP_TEACHER).Error()),
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, s.DB, mock.Anything).Twice().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				domainUserRepo.On("GetUserRoles", ctx, s.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				usrEmailRepo.On("Create", ctx, s.DB, mock.Anything, mock.Anything).Once().Return(&entity.UsrEmail{UsrID: database.Text(idutil.ULIDNow())}, nil)

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
				staffRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_user.ImportUsersResult{}, nil)
				teacherRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(fmt.Errorf("err when creating teacher"))
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "err when upserting user group member",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:         "Staff Name",
					UserGroup:    pb.UserGroup_USER_GROUP_TEACHER,
					Country:      cpb.Country_COUNTRY_JP,
					Email:        "sample-staff-email@example.com",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					LocationIds:  []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:     "validusername",
				},
			},
			expectedErr: errors.Wrapf(
				fmt.Errorf("err when upserting user group member"),
				"s.UserGroupService.UserGroupsMemberRepo.UpsertBatch",
			),
			Options: params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, s.DB, mock.Anything).Twice().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				domainUserRepo.On("GetUserRoles", ctx, s.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				usrEmailRepo.On("Create", ctx, s.DB, mock.Anything, mock.Anything).Once().Return(&entity.UsrEmail{UsrID: database.Text(idutil.ULIDNow())}, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)

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
				staffRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(fmt.Errorf("err when upserting user group member"))
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "user access path upsert error",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:         "Staff Name",
					UserGroup:    pb.UserGroup_USER_GROUP_TEACHER,
					Country:      cpb.Country_COUNTRY_JP,
					Email:        "sample-staff-email@example.com",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					LocationIds:  []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:     "validusername",
				},
			},
			expectedErr: status.Error(
				codes.Internal,
				errors.Wrap(
					errors.Wrap(fmt.Errorf("error"), "userAccessPathRepo.Upsert"),
					"usvc.UpsertUserAccessPath",
				).Error(),
			),
			Options: params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, s.DB, mock.Anything).Twice().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				domainUserRepo.On("GetUserRoles", ctx, s.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				usrEmailRepo.On("Create", ctx, s.DB, mock.Anything, mock.Anything).Once().Return(&entity.UsrEmail{UsrID: database.Text(idutil.ULIDNow())}, nil)

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
				staffRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(fmt.Errorf("error"))
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "cannot get tenant id",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:         "Staff Name",
					UserGroup:    pb.UserGroup_USER_GROUP_TEACHER,
					Country:      cpb.Country_COUNTRY_JP,
					Email:        "sample-staff-email@example.com",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					LocationIds:  []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:     "validusername",
				},
			},
			expectedErr: status.Error(codes.Internal, errcode.ErrCannotGetTenant.Error()),
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, s.DB, mock.Anything).Twice().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, s.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(false, nil)
				usrEmailRepo.On("Create", ctx, s.DB, mock.Anything, mock.Anything).Once().Return(&entity.UsrEmail{UsrID: database.Text(idutil.ULIDNow())}, nil)

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
				staffRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", fmt.Errorf("err when getting tenant id"))
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "err when getting tenant client",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:         "Staff Name",
					UserGroup:    pb.UserGroup_USER_GROUP_TEACHER,
					Country:      cpb.Country_COUNTRY_JP,
					Email:        "sample-staff-email@example.com",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					LocationIds:  []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:     "validusername",
				},
			},
			expectedErr: status.Error(codes.Internal, errors.Wrap(errors.Wrap(internal_user.ErrTenantNotFound, "TenantClient"), "cannot create user").Error()),
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, s.DB, mock.Anything).Twice().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, s.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(false, nil)
				usrEmailRepo.On("Create", ctx, s.DB, mock.Anything, mock.Anything).Once().Return(&entity.UsrEmail{UsrID: database.Text(idutil.ULIDNow())}, nil)

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
				staffRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(nil, internal_user.ErrTenantNotFound)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "happy case with flag staff location turn off",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:         "Staff Name",
					UserGroup:    pb.UserGroup_USER_GROUP_TEACHER,
					Country:      cpb.Country_COUNTRY_JP,
					Email:        "sample-staff-email@example.com",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StaffPhoneNumber: []*pb.StaffPhoneNumber{
						{
							PhoneNumber:     "123456789",
							PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER,
						},
						{
							PhoneNumber:     "023456888",
							PhoneNumberType: pb.StaffPhoneNumberType_STAFF_SECONDARY_PHONE_NUMBER,
						},
					},
					Gender:      pb.Gender_MALE,
					Birthday:    timestamppb.Now(),
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    "validusername",
				},
			},
			expectedErr: nil,
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, s.DB, mock.Anything).Twice().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, s.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				usrEmailRepo.On("Create", ctx, s.DB, mock.Anything, mock.Anything).Once().Return(&entity.UsrEmail{UsrID: database.Text(idutil.ULIDNow())}, nil)

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
				staffRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				teacherRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_user.ImportUsersResult{}, nil)
				tenantClient.On("GetHashConfig").Return(hashConfig)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectStaffUpsertTimesheetConfig, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectUpsertStaff, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name: happyCase,
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:         "Staff Name",
					UserGroup:    pb.UserGroup_USER_GROUP_TEACHER,
					Country:      cpb.Country_COUNTRY_JP,
					Email:        "sample-staff-email@example.com",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StaffPhoneNumber: []*pb.StaffPhoneNumber{
						{
							PhoneNumber:     "123456789",
							PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER,
						},
						{
							PhoneNumber:     "023456888",
							PhoneNumberType: pb.StaffPhoneNumberType_STAFF_SECONDARY_PHONE_NUMBER,
						},
					},
					Gender:        pb.Gender_MALE,
					Birthday:      timestamppb.Now(),
					LocationIds:   []string{idutil.ULIDNow(), idutil.ULIDNow()},
					WorkingStatus: pb.StaffWorkingStatus_RESIGNED,
					StartDate:     timestamppb.Now(),
					EndDate:       timestamppb.Now(),
					Remarks:       "Hello World",
					Username:      "validusername",
				},
			},
			expectedErr: nil,
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, s.DB, mock.Anything).Twice().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, s.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				usrEmailRepo.On("Create", ctx, s.DB, mock.Anything, mock.Anything).Once().Return(&entity.UsrEmail{UsrID: database.Text(idutil.ULIDNow())}, nil)

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
				staffRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				teacherRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_user.ImportUsersResult{}, nil)
				tenantClient.On("GetHashConfig").Return(hashConfig)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectStaffUpsertTimesheetConfig, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectUpsertStaff, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name: "error publish event upsert timesheet config",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:         "Staff Name",
					UserGroup:    pb.UserGroup_USER_GROUP_TEACHER,
					Country:      cpb.Country_COUNTRY_JP,
					Email:        "sample-staff-email@example.com",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					LocationIds:  []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:     "validusername",
				},
			},
			expectedErr: status.Error(codes.Unknown, "publishStaffSettingEvent error: error publish event"),
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, s.DB, mock.Anything).Twice().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, s.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				usrEmailRepo.On("Create", ctx, s.DB, mock.Anything, mock.Anything).Once().Return(&entity.UsrEmail{UsrID: database.Text(idutil.ULIDNow())}, nil)

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
				staffRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				teacherRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_user.ImportUsersResult{}, nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectStaffUpsertTimesheetConfig, mock.Anything).Once().Return(&nats.PubAck{}, errors.New("error publish event"))
			},
		},
		{
			name: "error publish event upsert staff",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:         "Staff Name",
					UserGroup:    pb.UserGroup_USER_GROUP_TEACHER,
					Country:      cpb.Country_COUNTRY_JP,
					Email:        "sample-staff-email@example.com",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					LocationIds:  []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:     "validusername",
				},
			},
			expectedErr: status.Error(codes.Unknown, "publishUpsertStaffEvent error: error publish event"),
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, s.DB, mock.Anything).Twice().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, s.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				usrEmailRepo.On("Create", ctx, s.DB, mock.Anything, mock.Anything).Once().Return(&entity.UsrEmail{UsrID: database.Text(idutil.ULIDNow())}, nil)

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
				staffRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				teacherRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_user.ImportUsersResult{}, nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(false, nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectStaffUpsertTimesheetConfig, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectUpsertStaff, mock.Anything).Once().Return(&nats.PubAck{}, errors.New("error publish event"))
			},
		},
		{
			name: "create new staff with primary phone number",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:         "Staff Name",
					UserGroup:    pb.UserGroup_USER_GROUP_TEACHER,
					Country:      cpb.Country_COUNTRY_JP,
					Email:        "sample-staff-email@example.com",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StaffPhoneNumber: []*pb.StaffPhoneNumber{
						{
							PhoneNumber:     "123456789",
							PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER,
						},
					},
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    "validusername",
				},
			},
			expectedErr: nil,
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, s.DB, mock.Anything).Twice().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, s.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				usrEmailRepo.On("Create", ctx, s.DB, mock.Anything, mock.Anything).Once().Return(&entity.UsrEmail{UsrID: database.Text(idutil.ULIDNow())}, nil)

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
				staffRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				teacherRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_user.ImportUsersResult{}, nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectStaffUpsertTimesheetConfig, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectUpsertStaff, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name: "create new staff with primary phone number with phone number empty",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:         "Staff Name",
					UserGroup:    pb.UserGroup_USER_GROUP_TEACHER,
					Country:      cpb.Country_COUNTRY_JP,
					Email:        "sample-staff-email@example.com",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StaffPhoneNumber: []*pb.StaffPhoneNumber{
						{
							PhoneNumber:     "",
							PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER,
						},
					},
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    "validusername",
				},
			},
			expectedErr: nil,
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, s.DB, mock.Anything).Twice().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, s.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				usrEmailRepo.On("Create", ctx, s.DB, mock.Anything, mock.Anything).Once().Return(&entity.UsrEmail{UsrID: database.Text(idutil.ULIDNow())}, nil)

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
				staffRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				teacherRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_user.ImportUsersResult{}, nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectStaffUpsertTimesheetConfig, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectUpsertStaff, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name: "err when create new staff with primary and secondary phone number is the same",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:         "Staff Name",
					UserGroup:    pb.UserGroup_USER_GROUP_TEACHER,
					Country:      cpb.Country_COUNTRY_JP,
					Email:        "sample-staff-email@example.com",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StaffPhoneNumber: []*pb.StaffPhoneNumber{
						{
							PhoneNumber:     "123456789",
							PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER,
						},
						{
							PhoneNumber:     "123456789",
							PhoneNumberType: pb.StaffPhoneNumberType_STAFF_SECONDARY_PHONE_NUMBER,
						},
					},
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    "validusername",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, errcode.ErrUserPhoneNumberIsDuplicate.Error()),
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
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
				staffRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectStaffUpsertTimesheetConfig, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectUpsertStaff, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name: "err when create new staff with primary phone number is not a number",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:         "Staff Name",
					UserGroup:    pb.UserGroup_USER_GROUP_TEACHER,
					Country:      cpb.Country_COUNTRY_JP,
					Email:        "sample-staff-email@example.com",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StaffPhoneNumber: []*pb.StaffPhoneNumber{
						{
							PhoneNumber:     "not a number",
							PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER,
						},
					},
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    "validusername",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "error regexp.MatchString: doesn't match"),
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				tx.On("Begin", mock.Anything).Once().Return(tx, nil)
				staffRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)

				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectStaffUpsertTimesheetConfig, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectUpsertStaff, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name: "err when create new staff with wrong type of phone number",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:         "Staff Name",
					UserGroup:    pb.UserGroup_USER_GROUP_TEACHER,
					Country:      cpb.Country_COUNTRY_JP,
					Email:        "sample-staff-email@example.com",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StaffPhoneNumber: []*pb.StaffPhoneNumber{
						{
							PhoneNumber:     "not a number",
							PhoneNumberType: 123,
						},
					},
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    "validusername",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, errcode.ErrUserPhoneNumberIsWrongType.Error()),
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
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
				staffRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectStaffUpsertTimesheetConfig, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectUpsertStaff, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name: "Err when UserPhoneNumberRepo.Upsert fail",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:         "Staff Name",
					UserGroup:    pb.UserGroup_USER_GROUP_TEACHER,
					Country:      cpb.Country_COUNTRY_JP,
					Email:        "sample-staff-email@example.com",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StaffPhoneNumber: []*pb.StaffPhoneNumber{
						{
							PhoneNumber:     "123456789",
							PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER,
						},
					},
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    "validusername",
				},
			},
			expectedErr: status.Error(codes.Unknown, "error from Upsert"),
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, s.DB, mock.Anything).Twice().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, s.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				usrEmailRepo.On("Create", ctx, s.DB, mock.Anything, mock.Anything).Once().Return(&entity.UsrEmail{UsrID: database.Text(idutil.ULIDNow())}, nil)

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
				staffRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_user.ImportUsersResult{}, nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(fmt.Errorf("error from Upsert"))
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectStaffUpsertTimesheetConfig, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectUpsertStaff, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "err when endate before start date",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:         "Staff Name",
					UserGroup:    pb.UserGroup_USER_GROUP_TEACHER,
					Country:      cpb.Country_COUNTRY_JP,
					Email:        "sample-staff-email@example.com",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					LocationIds:  []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StartDate:    timestamppb.New(time.Now()),
					EndDate:      timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Username:     "validusername",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, errcode.ErrStaffStartDateIsLessThanEndDate.Error()),
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
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
				staffRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectStaffUpsertTimesheetConfig, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectUpsertStaff, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name: "error when passing unexisted tag ids",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:             "Staff Name",
					UserGroup:        pb.UserGroup_USER_GROUP_TEACHER,
					Country:          cpb.Country_COUNTRY_JP,
					Email:            "sample-staff-email@example.com",
					UserGroupIds:     []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StaffPhoneNumber: []*pb.StaffPhoneNumber{},
					LocationIds:      []string{idutil.ULIDNow(), idutil.ULIDNow()},
					TagIds:           []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:         "validusername",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, ErrTagIDsMustBeExisted.Error()),
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
			},
		},
		{
			name: "error when passing wrong type tag ids",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:             "Staff Name",
					UserGroup:        pb.UserGroup_USER_GROUP_TEACHER,
					Country:          cpb.Country_COUNTRY_JP,
					Email:            "sample-staff-email@example.com",
					UserGroupIds:     []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StaffPhoneNumber: []*pb.StaffPhoneNumber{},
					LocationIds:      []string{idutil.ULIDNow(), idutil.ULIDNow()},
					TagIds:           []string{"tag_id_1", "tag_id_2"},
					Username:         "validusername",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, ErrTagIsNotForStaff.Error()),
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{
					createMockDomainTagWithTypeAndIsArchived("tag_id_1", pb.UserTagType_USER_TAG_TYPE_STAFF, false),
					createMockDomainTagWithTypeAndIsArchived("tag_id_2", pb.UserTagType_USER_TAG_TYPE_PARENT, false),
				}, nil)
			},
		},
		{
			name: "error when passing archived tag",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:             "Staff Name",
					UserGroup:        pb.UserGroup_USER_GROUP_TEACHER,
					Country:          cpb.Country_COUNTRY_JP,
					Email:            "sample-staff-email@example.com",
					UserGroupIds:     []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StaffPhoneNumber: []*pb.StaffPhoneNumber{},
					LocationIds:      []string{idutil.ULIDNow(), idutil.ULIDNow()},
					TagIds:           []string{"tag_id_1", "tag_id_2"},
					Username:         "validusername",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, ErrTagIsArchived.Error()),
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{
					createMockDomainTagWithTypeAndIsArchived("tag_id_1", pb.UserTagType_USER_TAG_TYPE_STAFF, false),
					createMockDomainTagWithTypeAndIsArchived("tag_id_2", pb.UserTagType_USER_TAG_TYPE_STAFF, true),
				}, nil)
			},
		},
		{
			name: "create new staff with staff tags",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:         "Staff Name",
					UserGroup:    pb.UserGroup_USER_GROUP_TEACHER,
					Country:      cpb.Country_COUNTRY_JP,
					Email:        "sample-staff-email@example.com",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StaffPhoneNumber: []*pb.StaffPhoneNumber{
						{
							PhoneNumber:     "123456789",
							PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER,
						},
					},
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					TagIds:      []string{"tag_id_1", "tag_id_2"},
					Username:    "validusername",
				},
			},
			expectedErr: nil,
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, s.DB, mock.Anything).Twice().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, s.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				usrEmailRepo.On("Create", ctx, s.DB, mock.Anything, mock.Anything).Once().Return(&entity.UsrEmail{UsrID: database.Text(idutil.ULIDNow())}, nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Twice().Return(entity.DomainTags{
					createMockDomainTagWithTypeAndIsArchived("tag_id_1", pb.UserTagType_USER_TAG_TYPE_STAFF, false),
					createMockDomainTagWithTypeAndIsArchived("tag_id_2", pb.UserTagType_USER_TAG_TYPE_STAFF, false),
				}, nil)
				domainTaggedUserRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)

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
				staffRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				teacherRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_user.ImportUsersResult{}, nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectStaffUpsertTimesheetConfig, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectUpsertStaff, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name: "happy case: create staff with external_user_id",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:         "Staff Name",
					UserGroup:    pb.UserGroup_USER_GROUP_TEACHER,
					Country:      cpb.Country_COUNTRY_JP,
					Email:        "sample-staff-email@example.com",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StaffPhoneNumber: []*pb.StaffPhoneNumber{
						{
							PhoneNumber:     "123456789",
							PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER,
						},
						{
							PhoneNumber:     "023456888",
							PhoneNumberType: pb.StaffPhoneNumberType_STAFF_SECONDARY_PHONE_NUMBER,
						},
					},
					Gender:         pb.Gender_MALE,
					Birthday:       timestamppb.Now(),
					LocationIds:    []string{idutil.ULIDNow(), idutil.ULIDNow()},
					WorkingStatus:  pb.StaffWorkingStatus_RESIGNED,
					StartDate:      timestamppb.Now(),
					EndDate:        timestamppb.Now(),
					Remarks:        "Hello World",
					ExternalUserId: "external_user_id",
					Username:       "validusername",
				},
			},
			expectedErr: nil,
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				domainUserRepo.On("GetByExternalUserIDs", ctx, s.DB, mock.Anything).Once().Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, s.DB, mock.Anything).Twice().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, s.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				usrEmailRepo.On("Create", ctx, s.DB, mock.Anything, mock.Anything).Once().Return(&entity.UsrEmail{UsrID: database.Text(idutil.ULIDNow())}, nil)

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
				staffRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				teacherRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_user.ImportUsersResult{}, nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectStaffUpsertTimesheetConfig, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectUpsertStaff, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name: "create staff with external_user_id with spaces",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:         "Staff Name",
					UserGroup:    pb.UserGroup_USER_GROUP_TEACHER,
					Country:      cpb.Country_COUNTRY_JP,
					Email:        "sample-staff-email@example.com",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StaffPhoneNumber: []*pb.StaffPhoneNumber{
						{
							PhoneNumber:     "123456789",
							PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER,
						},
						{
							PhoneNumber:     "023456888",
							PhoneNumberType: pb.StaffPhoneNumberType_STAFF_SECONDARY_PHONE_NUMBER,
						},
					},
					Gender:         pb.Gender_MALE,
					Birthday:       timestamppb.Now(),
					LocationIds:    []string{idutil.ULIDNow(), idutil.ULIDNow()},
					WorkingStatus:  pb.StaffWorkingStatus_RESIGNED,
					StartDate:      timestamppb.Now(),
					EndDate:        timestamppb.Now(),
					Remarks:        "Hello World",
					ExternalUserId: " trim external_user_id ",
					Username:       "validusername",
				},
			},
			expectedErr: nil,
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				domainUserRepo.On("GetByExternalUserIDs", ctx, s.DB, mock.Anything).Once().Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, s.DB, mock.Anything).Twice().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, s.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				usrEmailRepo.On("Create", ctx, s.DB, mock.Anything, mock.Anything).Once().Return(&entity.UsrEmail{UsrID: database.Text(idutil.ULIDNow())}, nil)

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
				staffRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				teacherRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_user.ImportUsersResult{}, nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectStaffUpsertTimesheetConfig, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectUpsertStaff, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name: "create staff with empty external_user_id with spaces",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:         "Staff Name",
					UserGroup:    pb.UserGroup_USER_GROUP_TEACHER,
					Country:      cpb.Country_COUNTRY_JP,
					Email:        "sample-staff-email@example.com",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StaffPhoneNumber: []*pb.StaffPhoneNumber{
						{
							PhoneNumber:     "123456789",
							PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER,
						},
						{
							PhoneNumber:     "023456888",
							PhoneNumberType: pb.StaffPhoneNumberType_STAFF_SECONDARY_PHONE_NUMBER,
						},
					},
					Gender:         pb.Gender_MALE,
					Birthday:       timestamppb.Now(),
					LocationIds:    []string{idutil.ULIDNow(), idutil.ULIDNow()},
					WorkingStatus:  pb.StaffWorkingStatus_RESIGNED,
					StartDate:      timestamppb.Now(),
					EndDate:        timestamppb.Now(),
					Remarks:        "Hello World",
					ExternalUserId: "  ",
					Username:       "validusername",
				},
			},
			expectedErr: nil,
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, s.DB, mock.Anything).Twice().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, s.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				usrEmailRepo.On("Create", ctx, s.DB, mock.Anything, mock.Anything).Once().Return(&entity.UsrEmail{UsrID: database.Text(idutil.ULIDNow())}, nil)

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
				staffRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				teacherRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_user.ImportUsersResult{}, nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectStaffUpsertTimesheetConfig, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectUpsertStaff, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name: "happy case: create staff with username",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:         "Staff Name",
					UserGroup:    pb.UserGroup_USER_GROUP_TEACHER,
					Country:      cpb.Country_COUNTRY_JP,
					Email:        "sample-staff-email@example.com",
					Username:     "ValidUsername",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StaffPhoneNumber: []*pb.StaffPhoneNumber{
						{
							PhoneNumber:     "123456789",
							PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER,
						},
						{
							PhoneNumber:     "023456888",
							PhoneNumberType: pb.StaffPhoneNumberType_STAFF_SECONDARY_PHONE_NUMBER,
						},
					},
					Gender:        pb.Gender_MALE,
					Birthday:      timestamppb.Now(),
					LocationIds:   []string{idutil.ULIDNow(), idutil.ULIDNow()},
					WorkingStatus: pb.StaffWorkingStatus_RESIGNED,
					StartDate:     timestamppb.Now(),
					EndDate:       timestamppb.Now(),
					Remarks:       "Hello World",
				},
			},
			expectedErr: nil,
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
				domainUserRepo.On("GetByExternalUserIDs", ctx, s.DB, mock.Anything).Once().Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, s.DB, mock.Anything).Twice().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, s.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				usrEmailRepo.On("Create", ctx, s.DB, mock.Anything, mock.Anything).Once().Return(&entity.UsrEmail{UsrID: database.Text(idutil.ULIDNow())}, nil)

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
				staffRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				teacherRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_user.ImportUsersResult{}, nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectStaffUpsertTimesheetConfig, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectUpsertStaff, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name: "happy case: create staff with username with email format",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:         "Staff Name",
					UserGroup:    pb.UserGroup_USER_GROUP_TEACHER,
					Country:      cpb.Country_COUNTRY_JP,
					Email:        "sample-staff-email@example.com",
					Username:     "sample-staff-username@example.com",
					UserGroupIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					StaffPhoneNumber: []*pb.StaffPhoneNumber{
						{
							PhoneNumber:     "123456789",
							PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER,
						},
						{
							PhoneNumber:     "023456888",
							PhoneNumberType: pb.StaffPhoneNumberType_STAFF_SECONDARY_PHONE_NUMBER,
						},
					},
					Gender:        pb.Gender_MALE,
					Birthday:      timestamppb.Now(),
					LocationIds:   []string{idutil.ULIDNow(), idutil.ULIDNow()},
					WorkingStatus: pb.StaffWorkingStatus_RESIGNED,
					StartDate:     timestamppb.Now(),
					EndDate:       timestamppb.Now(),
					Remarks:       "Hello World",
				},
			},
			expectedErr: nil,
			Options:     params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"sample-staff-username@example.com"}).Return(entity.Users{}, nil)
				domainUserRepo.On("GetByExternalUserIDs", ctx, s.DB, mock.Anything).Once().Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				userGroupV2Repo.On("FindByIDs", ctx, tx, mock.Anything).Once().Return([]*entity.UserGroupV2{{ResourcePath: database.Text(defaultResourcePath)}, {ResourcePath: database.Text(defaultResourcePath)}}, nil)
				domainRoleRepo.On("GetByUserGroupIDs", ctx, s.DB, mock.Anything).Twice().Return(entity.DomainRoles{roleTeacher}, nil)
				domainUserRepo.On("GetUserRoles", ctx, s.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				unleashClient.On("IsFeatureEnabled", constant.FeatureToggleAllowCombinationMultipleRoles, mock.Anything).Once().Return(true, nil)
				usrEmailRepo.On("Create", ctx, s.DB, mock.Anything, mock.Anything).Once().Return(&entity.UsrEmail{UsrID: database.Text(idutil.ULIDNow())}, nil)

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
				staffRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				teacherRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				organizationRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_user.ImportUsersResult{}, nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectStaffUpsertTimesheetConfig, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectUpsertStaff, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name: "err: missing username",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:        "Staff Name",
					UserGroup:   pb.UserGroup_USER_GROUP_TEACHER,
					Country:     cpb.Country_COUNTRY_JP,
					Email:       "sample-staff-email@example.com",
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    "",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, entity.MissingMandatoryFieldError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      -1,
			}.Error()),
			Options: params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "err: existed username",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:        "Staff Name",
					UserGroup:   pb.UserGroup_USER_GROUP_TEACHER,
					Country:     cpb.Country_COUNTRY_JP,
					Email:       "sample-staff-email@example.com",
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    "existedusername",
				},
			},
			expectedErr: status.Error(codes.AlreadyExists, entity.ExistingDataError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      0,
			}.Error()),
			Options: params{resourcePath: defaultResourcePath},
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
		},
		{
			name: "err: invalid username",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:        "Staff Name",
					UserGroup:   pb.UserGroup_USER_GROUP_TEACHER,
					Country:     cpb.Country_COUNTRY_JP,
					Email:       "sample-staff-email@example.com",
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    "invalid_username",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, entity.InvalidFieldError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      -1,
				Reason:     entity.NotMatchingPattern,
			}.Error()),
			Options: params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "err: invalid username with spaces",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:        "Staff Name",
					UserGroup:   pb.UserGroup_USER_GROUP_TEACHER,
					Country:     cpb.Country_COUNTRY_JP,
					Email:       "sample-staff-email@example.com",
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    "invalid username",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, entity.InvalidFieldError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      -1,
				Reason:     entity.NotMatchingPattern,
			}.Error()),
			Options: params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
			},
		},
		{
			name: "err: invalid username with wrong email format",
			ctx:  ctx,
			req: &pb.CreateStaffRequest{
				Staff: &pb.CreateStaffRequest_StaffProfile{
					Name:        "Staff Name",
					UserGroup:   pb.UserGroup_USER_GROUP_TEACHER,
					Country:     cpb.Country_COUNTRY_JP,
					Email:       "sample-staff-email@example.com",
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    "invalid_username@",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, entity.InvalidFieldError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      -1,
				Reason:     entity.NotMatchingPattern,
			}.Error()),
			Options: params{resourcePath: defaultResourcePath},
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleStaffUsername, mock.Anything, mock.Anything).Once().Return(true, nil)
			},
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

			t.Log(testCaseLog + testCase.name)
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}

			_, err := s.CreateStaff(testCase.ctx, testCase.req.(*pb.CreateStaffRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, userRepo, schoolAdminRepo, firebaseClient, usrEmailRepo, userGroupRepo, teacherRepo, userGroupsMemberRepo, firebaseAuthClient, organizationRepo, tenantManager, tenantClient, userPhoneNumberRepo, userGroupV2Repo, domainTagRepo, roleRepo)
		})
	}
}

func TestStaffService_ValidateStaffPhoneNumber(t *testing.T) {
	t.Parallel()
	_, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	type params struct {
		resourcePath string
	}

	testCases := []TestCase{
		{
			name: "happy case ",
			req: []*pb.StaffPhoneNumber{
				{
					PhoneNumber:     "123214125",
					PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER,
				},
				{
					PhoneNumber:     "123214126",
					PhoneNumberType: pb.StaffPhoneNumberType_STAFF_SECONDARY_PHONE_NUMBER,
				},
				{
					PhoneNumber:     "123214127",
					PhoneNumberType: pb.StaffPhoneNumberType_STAFF_SECONDARY_PHONE_NUMBER,
				},
			},
			expectedErr: nil,
		},
		{
			name: "create new staff with primary phone number",
			req: []*pb.StaffPhoneNumber{
				{
					PhoneNumber:     "123214125",
					PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER,
				},
			},
			expectedErr: nil,
		},
		{
			name: "create new staff with empty primary phone number ",
			req: []*pb.StaffPhoneNumber{
				{
					PhoneNumber:     "",
					PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER,
				},
			},
			expectedErr: nil,
		},
		{
			name:        "create new staff with nil phone number",
			req:         []*pb.StaffPhoneNumber{},
			expectedErr: nil,
		},
		{
			name: "err when create new staff with wrong type phone number",
			req: []*pb.StaffPhoneNumber{
				{
					PhoneNumber:     "123214125",
					PhoneNumberType: 4,
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, errcode.ErrUserPhoneNumberIsWrongType.Error()),
		},
		{
			name: "err when create new staff with duplicate phone number",
			req: []*pb.StaffPhoneNumber{
				{
					PhoneNumber:     "123214125",
					PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER,
				},
				{
					PhoneNumber:     "32141256",
					PhoneNumberType: pb.StaffPhoneNumberType_STAFF_SECONDARY_PHONE_NUMBER,
				},
				{
					PhoneNumber:     "123214125",
					PhoneNumberType: pb.StaffPhoneNumberType_STAFF_SECONDARY_PHONE_NUMBER,
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, errcode.ErrUserPhoneNumberIsDuplicate.Error()),
		},
		{
			name: "err when create new staff with duplicate type primary phone number",
			req: []*pb.StaffPhoneNumber{
				{
					PhoneNumber:     "123214125",
					PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER,
				},
				{
					PhoneNumber:     "32141256",
					PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER,
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, errcode.ErrUserPrimaryPhoneNumberIsRedundant.Error()),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			t.Log(testCaseLog + testCase.name)

			err := validateStaffPhoneNumber(testCase.req.([]*pb.StaffPhoneNumber))

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}

		})
	}
}

func TestStaffService_ValidationsCreateStaff(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// db := new(mock_database.Ext)
	tx := new(mock_database.Tx)

	userRepo := new(mock_repositories.MockUserRepo)
	schoolAdminRepo := new(mock_repositories.MockSchoolAdminRepo)
	firebaseClient := new(mock_firebase.AuthClient)
	firebaseUtils := new(mock_firebase.AuthUtils)
	domainUserRepo := new(mock_repositories.MockDomainUserRepo)
	unleashClient := new(mock_unleash_client.UnleashClientInstance)

	umsvc := &usvc.UserModifierService{
		DB:              tx,
		UserRepo:        userRepo,
		SchoolAdminRepo: schoolAdminRepo,
		FirebaseClient:  firebaseClient,
		DomainUserRepo:  domainUserRepo,
	}
	s := &StaffService{
		DB:                  umsvc.DB,
		FirebaseClient:      umsvc.FirebaseClient,
		FirebaseUtils:       firebaseUtils,
		FatimaClient:        umsvc.FatimaClient,
		UserModifierService: umsvc,
		UserRepo:            domainUserRepo,
		UnleashClient:       unleashClient,
	}

	user := &entity.LegacyUser{
		ID:       database.Text(idutil.ULIDNow()),
		Email:    database.Text("user-email@manabie.com"),
		FullName: database.Text("FullName"),
		Avatar:   database.Text("http://example.com/avatar.png"),
		UserName: database.Text("ValidUsername"),
	}

	existingUser := &entity.LegacyUser{
		ID:       database.Text(idutil.ULIDNow()),
		Email:    database.Text("existing-student-email@example.com"),
		FullName: database.Text("FullName"),
		Avatar:   database.Text("http://example.com/avatar.png"),
		UserName: database.Text("existingusername"),
	}

	existingUserExternal := entity.Users{
		entity_mock.User{
			RandomUser: entity_mock.RandomUser{
				Email:          field.NewString("staff@manabie.com"),
				ExternalUserID: field.NewString("existing-external_user_id"),
			},
		},
	}

	existingSchoolAdmin := &entity.SchoolAdmin{
		SchoolAdminID: existingUser.ID,
		LegacyUser:    *existingUser,
		SchoolID:      database.Int4(constants.ManabieSchool),
	}

	type params struct {
		userProfile                   *pb.CreateStaffRequest_StaffProfile
		userGroup                     string
		organization                  string
		schoolID                      int64
		isFeatureStaffUsernameEnabled bool
	}

	testCases := []TestCase{
		{
			name:        happyCase,
			ctx:         ctx,
			req:         nil,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				existingUser.Group.Set(constant.UserGroupSchoolAdmin)
				userRepo.On("Get", ctx, tx, mock.Anything).Once().Return(existingUser, nil)
				schoolAdminRepo.On("Get", ctx, tx, mock.Anything).Return(existingSchoolAdmin, nil).Once()
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				firebaseUtils.On("IsUserNotFound", mock.Anything).Once().Return(true)
			},
			Options: params{
				userProfile: &pb.CreateStaffRequest_StaffProfile{
					Name:        user.GetName(),
					Country:     cpb.Country_COUNTRY_VN,
					Email:       existingUser.Email.String,
					Avatar:      user.GetPhotoURL(),
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    user.UserName.String,
				},
				userGroup:                     constant.UserGroupTeacher,
				organization:                  "",
				schoolID:                      constants.ManabieSchool,
				isFeatureStaffUsernameEnabled: false,
			},
		},
		{
			name: "validation create user email is exist in DB",
			ctx:  ctx,
			req:  nil,
			expectedErr: status.Error(
				codes.AlreadyExists,
				errcode.ErrUserEmailExists.Error(),
			),
			setup: func(ctx context.Context) {
				existingUser.Group.Set(constant.UserGroupSchoolAdmin)
				userRepo.On("Get", ctx, tx, mock.Anything).Once().Return(existingUser, nil)
				schoolAdminRepo.On("Get", ctx, tx, mock.Anything).Return(existingSchoolAdmin, nil).Once()
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{existingUser}, nil)
				firebaseUtils.On("IsUserNotFound", mock.Anything).Once().Return(true)
			},
			Options: params{
				userProfile: &pb.CreateStaffRequest_StaffProfile{
					Name:        user.GetName(),
					Country:     cpb.Country_COUNTRY_VN,
					Email:       existingUser.Email.String,
					Avatar:      user.GetPhotoURL(),
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    user.UserName.String,
				},
				userGroup:                     constant.UserGroupTeacher,
				organization:                  "",
				schoolID:                      constants.ManabieSchool,
				isFeatureStaffUsernameEnabled: false,
			},
		},
		{
			name: "country cannot be empty",
			ctx:  ctx,
			req:  nil,
			expectedErr: status.Error(
				codes.InvalidArgument,
				errcode.ErrUserCountryIsEmpty.Error(),
			),
			Options: params{
				userProfile: &pb.CreateStaffRequest_StaffProfile{
					Name:     user.GetName(),
					Email:    existingUser.Email.String,
					Avatar:   user.GetPhotoURL(),
					Username: user.UserName.String,
				},
				userGroup:                     constant.UserGroupTeacher,
				organization:                  "",
				schoolID:                      constants.ManabieSchool,
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
				userProfile: &pb.CreateStaffRequest_StaffProfile{
					Name:        user.GetName(),
					Country:     cpb.Country_COUNTRY_VN,
					Avatar:      user.GetPhotoURL(),
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    user.UserName.String,
				},
				userGroup:                     constant.UserGroupTeacher,
				organization:                  "",
				schoolID:                      constants.ManabieSchool,
				isFeatureStaffUsernameEnabled: false,
			},
		},
		{
			name:        "err when UserModifierService.UserRepo.GetByEmailInsensitiveCase fail",
			ctx:         ctx,
			req:         nil,
			expectedErr: status.Error(codes.Internal, errors.New("s.UserRepo.GetByEmailInsensitiveCase: error from get by email").Error()),
			setup: func(ctx context.Context) {
				existingUser.Group.Set(constant.UserGroupSchoolAdmin)
				userRepo.On("Get", ctx, tx, mock.Anything).Once().Return(existingUser, nil)
				schoolAdminRepo.On("Get", ctx, tx, mock.Anything).Return(existingSchoolAdmin, nil).Once()
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, errors.New("error from get by email"))
			},
			Options: params{
				userProfile: &pb.CreateStaffRequest_StaffProfile{
					Name:        user.GetName(),
					Country:     cpb.Country_COUNTRY_VN,
					Email:       existingUser.Email.String,
					Avatar:      user.GetPhotoURL(),
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    user.UserName.String,
				},
				userGroup:                     constant.UserGroupTeacher,
				organization:                  "",
				schoolID:                      constants.ManabieSchool,
				isFeatureStaffUsernameEnabled: false,
			},
		},
		{
			name:        "error when end date before start date",
			ctx:         ctx,
			req:         nil,
			expectedErr: status.Errorf(codes.InvalidArgument, errcode.ErrStaffStartDateIsLessThanEndDate.Error()),
			Options: params{
				userProfile: &pb.CreateStaffRequest_StaffProfile{
					Name:          user.GetName(),
					Country:       cpb.Country_COUNTRY_VN,
					Avatar:        user.GetPhotoURL(),
					Email:         existingUser.Email.String,
					LocationIds:   []string{idutil.ULIDNow(), idutil.ULIDNow()},
					WorkingStatus: pb.StaffWorkingStatus_AVAILABLE,
					StartDate:     timestamppb.New(time.Now()),
					EndDate:       timestamppb.New(time.Now().Add(-24 * time.Hour)),
					Username:      user.UserName.String,
				},
				userGroup:                     constant.UserGroupTeacher,
				organization:                  "",
				schoolID:                      constants.ManabieSchool,
				isFeatureStaffUsernameEnabled: false,
			},
		},
		{
			name:        "error when empty name",
			ctx:         ctx,
			req:         nil,
			expectedErr: status.Errorf(codes.InvalidArgument, errcode.ErrUserFullNameIsEmpty.Error()),
			Options: params{
				userProfile: &pb.CreateStaffRequest_StaffProfile{
					Username: user.UserName.String,
				},
				isFeatureStaffUsernameEnabled: false,
			},
		},
		{
			name:        "error when empty name and empty first name",
			ctx:         ctx,
			req:         nil,
			expectedErr: status.Errorf(codes.InvalidArgument, errcode.ErrUserFirstNameIsEmpty.Error()),
			Options: params{
				userProfile: &pb.CreateStaffRequest_StaffProfile{
					UserNameFields: &pb.UserNameFields{
						LastName: "John",
					},
					Username: user.UserName.String,
				},
				isFeatureStaffUsernameEnabled: false,
			},
		},
		{
			name:        "error when empty name and empty last name",
			ctx:         ctx,
			req:         nil,
			expectedErr: status.Errorf(codes.InvalidArgument, errcode.ErrUserLastNameIsEmpty.Error()),
			Options: params{
				userProfile: &pb.CreateStaffRequest_StaffProfile{
					UserNameFields: &pb.UserNameFields{
						FirstName: "Doe",
					},
					Username: user.UserName.String,
				},
				isFeatureStaffUsernameEnabled: false,
			},
		},
		{
			name: "validation create user external_user_id is exist in DB",
			ctx:  ctx,
			req:  nil,
			expectedErr: status.Error(
				codes.AlreadyExists,
				errcode.ErrUserExternalUserIDExists.Error(),
			),
			setup: func(ctx context.Context) {
				existingUser.Group.Set(constant.UserGroupSchoolAdmin)
				userRepo.On("Get", ctx, tx, mock.Anything).Once().Return(existingUser, nil)
				schoolAdminRepo.On("Get", ctx, tx, mock.Anything).Return(existingSchoolAdmin, nil).Once()
				domainUserRepo.On("GetByExternalUserIDs", ctx, s.DB, mock.Anything).Once().Return(existingUserExternal, nil)
			},
			Options: params{
				userProfile: &pb.CreateStaffRequest_StaffProfile{
					Name:           user.GetName(),
					Country:        cpb.Country_COUNTRY_VN,
					Email:          "sample-staff-email@example.com",
					Avatar:         user.GetPhotoURL(),
					LocationIds:    []string{idutil.ULIDNow(), idutil.ULIDNow()},
					ExternalUserId: existingUserExternal[0].ExternalUserID().String(),
					Username:       user.UserName.String,
				},
				userGroup:                     constant.UserGroupTeacher,
				organization:                  "",
				schoolID:                      constants.ManabieSchool,
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
				userProfile: &pb.CreateStaffRequest_StaffProfile{
					Name:        user.GetName(),
					Country:     cpb.Country_COUNTRY_VN,
					Avatar:      user.GetPhotoURL(),
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    "",
				},
				userGroup:                     constant.UserGroupTeacher,
				organization:                  "",
				schoolID:                      constants.ManabieSchool,
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
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{existingUser.UserName.String}).Return(entity.Users{
					entity_mock.User{
						RandomUser: entity_mock.RandomUser{
							UserID:   field.NewString("user-id"),
							UserName: field.NewString(existingUser.UserName.String),
						},
					},
				}, nil)
			},
			Options: params{
				userProfile: &pb.CreateStaffRequest_StaffProfile{
					Name:        user.GetName(),
					Country:     cpb.Country_COUNTRY_VN,
					Email:       user.GetEmail(),
					Avatar:      user.GetPhotoURL(),
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    existingUser.UserName.String,
				},
				userGroup:                     constant.UserGroupTeacher,
				organization:                  "",
				schoolID:                      constants.ManabieSchool,
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
				userProfile: &pb.CreateStaffRequest_StaffProfile{
					Name:        user.GetName(),
					Country:     cpb.Country_COUNTRY_VN,
					Avatar:      user.GetPhotoURL(),
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    "Invalid_Username",
				},
				userGroup:                     constant.UserGroupTeacher,
				organization:                  "",
				schoolID:                      constants.ManabieSchool,
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
				userProfile: &pb.CreateStaffRequest_StaffProfile{
					Name:        user.GetName(),
					Country:     cpb.Country_COUNTRY_VN,
					Avatar:      user.GetPhotoURL(),
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    "Invalid_Username@",
				},
				userGroup:                     constant.UserGroupTeacher,
				organization:                  "",
				schoolID:                      constants.ManabieSchool,
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
				userProfile: &pb.CreateStaffRequest_StaffProfile{
					Name:        user.GetName(),
					Country:     cpb.Country_COUNTRY_VN,
					Avatar:      user.GetPhotoURL(),
					LocationIds: []string{idutil.ULIDNow(), idutil.ULIDNow()},
					Username:    "Invalid Username",
				},
				userGroup:                     constant.UserGroupTeacher,
				organization:                  "",
				schoolID:                      constants.ManabieSchool,
				isFeatureStaffUsernameEnabled: true,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: "1",
				},
			}

			testCase.ctx = interceptors.ContextWithJWTClaims(ctx, claim)

			t.Log(testCaseLog + testCase.name)
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}
			err := s.validationsCreateStaff(
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

func TestStaffService_PBStaffPhoneNumberToUserPhoneNumber(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		staffPhoneNumbers []*pb.StaffPhoneNumber
		userID            string
		resourcePath      string
		expectedErr       error
	}{
		{
			name: "happy case",
			staffPhoneNumbers: []*pb.StaffPhoneNumber{
				{
					PhoneNumber:     "4567812312",
					PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER,
				},
			},
			userID:       "UserId",
			resourcePath: "Resource path",
			expectedErr:  nil,
		},
		{
			name: "happy case when have phone number id",
			staffPhoneNumbers: []*pb.StaffPhoneNumber{
				{
					PhoneNumberId:   "21321341",
					PhoneNumber:     "4567812312",
					PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER,
				},
			},
			userID:       "UserId",
			resourcePath: "Resource path",
			expectedErr:  nil,
		},
	}

	for _, tcase := range testCases {
		t.Run(tcase.name, func(t *testing.T) {
			userPhoneNumbers, err := pbStaffPhoneNumberToUserPhoneNumber(
				tcase.staffPhoneNumbers,
				tcase.userID,
				tcase.resourcePath,
			)

			assert.NoError(t, err)

			assert.Equal(t, tcase.expectedErr, err)
			assert.Equal(t, len(tcase.staffPhoneNumbers), len(userPhoneNumbers))
			for index, value := range userPhoneNumbers {
				assert.Equal(t, tcase.staffPhoneNumbers[index].PhoneNumber, value.PhoneNumber.Get())
				assert.Equal(t, tcase.staffPhoneNumbers[index].PhoneNumberType.String(), value.PhoneNumberType.Get())
				assert.Equal(t, tcase.userID, value.UserID.Get())
				assert.Equal(t, tcase.resourcePath, value.ResourcePath.Get())
			}
		})
	}
}

func TestStaffService_checkPermissionToAssignUserGroup(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	userRepo := new(mock_repositories.MockDomainUserRepo)
	roleRepo := new(mock_repositories.MockDomainRoleRepo)
	db := new(mock_database.Ext)

	service := StaffService{
		DB:       db,
		UserRepo: userRepo,
		RoleRepo: roleRepo,
	}

	roleSchoolAdmin := repository.NewNullRole()
	roleSchoolAdmin.RoleAttribute.RoleName = field.NewString(constant.RoleSchoolAdmin)

	roleHQStaff := repository.NewNullRole()
	roleHQStaff.RoleAttribute.RoleName = field.NewString(constant.RoleHQStaff)

	roleTeacher := repository.NewNullRole()
	roleTeacher.RoleAttribute.RoleName = field.NewString(constant.RoleTeacher)

	roleCentreManager := repository.NewNullRole()
	roleCentreManager.RoleAttribute.RoleName = field.NewString(constant.RoleCentreManager)

	testCases := []TestCase{
		{
			ctx:  ctx,
			name: "get roles failed",
			req:  []string{idutil.ULIDNow()},
			setup: func(ctx context.Context) {
				roleRepo.On("GetByUserGroupIDs", ctx, service.DB, mock.Anything).Once().Return(entity.DomainRoles{}, pgx.ErrTxClosed)
			},
			expectedErr: pgx.ErrTxClosed,
		},
		{
			ctx:  ctx,
			name: "roles are not include school admin",
			req:  []string{idutil.ULIDNow()},
			setup: func(ctx context.Context) {
				roleRepo.On("GetByUserGroupIDs", ctx, service.DB, mock.Anything).Once().Return(entity.DomainRoles{roleHQStaff}, nil)
			},
			expectedErr: nil,
		},
		{
			ctx:  ctx,
			name: "get user roles failed",
			req:  []string{idutil.ULIDNow()},
			setup: func(ctx context.Context) {
				roleRepo.On("GetByUserGroupIDs", ctx, service.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				userRepo.On("GetUserRoles", ctx, service.DB, mock.Anything).Once().Return(entity.DomainRoles{}, pgx.ErrTxClosed)
			},
			expectedErr: pgx.ErrTxClosed,
		},
		{
			ctx:  ctx,
			name: "user roles empty",
			req:  []string{idutil.ULIDNow()},
			setup: func(ctx context.Context) {
				roleRepo.On("GetByUserGroupIDs", ctx, service.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				userRepo.On("GetUserRoles", ctx, service.DB, mock.Anything).Once().Return(entity.DomainRoles{}, nil)
			},
			expectedErr: fmt.Errorf("current user don't have permission to assign user_group"),
		},
		{
			ctx:  ctx,
			name: "HQ Staff don't have permission to assign user group was school admin role",
			req:  []string{idutil.ULIDNow()},
			setup: func(ctx context.Context) {
				roleRepo.On("GetByUserGroupIDs", ctx, service.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				userRepo.On("GetUserRoles", ctx, service.DB, mock.Anything).Once().Return(entity.DomainRoles{roleHQStaff}, nil)
			},
			expectedErr: fmt.Errorf("%s can't assign this user group", entity.DomainRoles{roleHQStaff}.RoleNames()),
		},
		{
			ctx:  ctx,
			name: "CM don't have permission to assign user group was school admin role",
			req:  []string{idutil.ULIDNow()},
			setup: func(ctx context.Context) {
				roleRepo.On("GetByUserGroupIDs", ctx, service.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				userRepo.On("GetUserRoles", ctx, service.DB, mock.Anything).Once().Return(entity.DomainRoles{roleCentreManager}, nil)
			},
			expectedErr: fmt.Errorf("%s can't assign this user group", entity.DomainRoles{roleCentreManager}.RoleNames()),
		},
		{
			ctx:  ctx,
			name: "happy case: school admin can assign user_group was granted role school admin",
			req:  []string{idutil.ULIDNow()},
			setup: func(ctx context.Context) {
				roleRepo.On("GetByUserGroupIDs", ctx, service.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
				userRepo.On("GetUserRoles", ctx, service.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)
			},
			expectedErr: nil,
		},
		{
			ctx:  ctx,
			name: "happy case: hq staff/centre can assign user_group was granted role teacher",
			req:  []string{idutil.ULIDNow()},
			setup: func(ctx context.Context) {
				roleRepo.On("GetByUserGroupIDs", ctx, service.DB, mock.Anything).Once().Return(entity.DomainRoles{roleTeacher}, nil)

			},
			expectedErr: nil,
		},
		{
			ctx:  ctx,
			name: "happy case: staff role school admin and teacher can assign user_group was granted role school admin",
			req:  []string{idutil.ULIDNow()},
			setup: func(ctx context.Context) {
				roleRepo.On("GetByUserGroupIDs", ctx, service.DB, mock.Anything).Once().Return(entity.DomainRoles{roleTeacher, roleSchoolAdmin}, nil)
				userRepo.On("GetUserRoles", ctx, service.DB, mock.Anything).Once().Return(entity.DomainRoles{roleSchoolAdmin}, nil)

			},
			expectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}

			err := service.checkPermissionToAssignUserGroup(testCase.ctx, service.DB, testCase.req.([]string))

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
	mock.AssertExpectationsForObjects(t, db, userRepo, roleRepo)
}

func Test_createStaffPbToStaffEnt(t *testing.T) {
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
	userId := "random-id"
	externalUserID := "random-external_user_id"
	username := "username"

	testCases := []TestCase{
		{
			name: happyCase,
			req: &pb.CreateStaffRequest_StaffProfile{
				Name:           "John Doe",
				Email:          "staff_manabie@manabie.com",
				Country:        cpb.Country_COUNTRY_VN,
				Gender:         gender,
				Birthday:       birthday,
				WorkingStatus:  pb.StaffWorkingStatus_AVAILABLE,
				StartDate:      startDate,
				EndDate:        endDate,
				Remarks:        "Staff Remarks",
				ExternalUserId: "random-external_user_id",
				Username:       username,
			},
			Options: func() {
				//staff
				defaultStaff.ResourcePath = database.Text(resourcePath)
				defaultStaff.DeletedAt = pgtype.Timestamptz{Status: pgtype.Null}
				defaultStaff.WorkingStatus = database.Text(pb.StaffWorkingStatus_AVAILABLE.String())
				defaultStaff.AutoCreateTimesheet = database.Bool(false)
				defaultStaff.StartDate = database.DateFromPb(startDate)
				defaultStaff.EndDate = database.DateFromPb(endDate)
				defaultStaff.ID = database.Text(userId)
				defaultStaff.UserRole = database.Text(string(constant.UserRoleStaff))
				//user
				defaultUser := entity.LegacyUser{}
				database.AllNullEntity(&defaultUser)
				defaultUser.FullName = database.Text("John Doe")
				defaultUser.Email = database.Text("staff_manabie@manabie.com")
				defaultUser.Birthday = pgtype.Date{Time: birthday.AsTime(), Status: pgtype.Present}
				defaultUser.Gender = database.Text(gender.String())
				defaultUser.Country = database.Text(cpb.Country_COUNTRY_VN.String())
				defaultUser.Group = database.Text(constant.UserGroupSchoolAdmin)
				defaultUser.LastName = database.Text("John")
				defaultUser.FirstName = database.Text("Doe")
				defaultUser.ResourcePath = database.Text(resourcePath)
				defaultUser.Remarks = database.Text("Staff Remarks")
				defaultUser.ID = database.Text(userId)
				defaultUser.ExternalUserID = database.Text(externalUserID)
				defaultUser.UserRole = database.Text(string(constant.UserRoleStaff))
				// Temporarily set loginEmail equal Email
				defaultUser.LoginEmail = database.Text("staff_manabie@manabie.com")
				defaultUser.UserName = database.Text(username)
				defaultStaff.LegacyUser = defaultUser
			},
			expectedRes: defaultStaff,
		},
		{
			name: "return correct data when providing UserNameFields",
			req: &pb.CreateStaffRequest_StaffProfile{
				Email:         "staff_manabie@manabie.com",
				Country:       cpb.Country_COUNTRY_VN,
				Gender:        gender,
				Birthday:      birthday,
				WorkingStatus: pb.StaffWorkingStatus_AVAILABLE,
				StartDate:     startDate,
				EndDate:       endDate,
				Remarks:       "Staff Remarks",
				UserNameFields: &pb.UserNameFields{
					LastName:          "John",
					FirstName:         "Doe",
					LastNamePhonetic:  "LastNamePhonetic",
					FirstNamePhonetic: "FirstNamePhonetic",
				},
				ExternalUserId: "random-external_user_id",
				Username:       username,
			},
			Options: func() {
				//staff
				defaultStaff.ResourcePath = database.Text(resourcePath)
				defaultStaff.DeletedAt = pgtype.Timestamptz{Status: pgtype.Null}
				defaultStaff.WorkingStatus = database.Text(pb.StaffWorkingStatus_AVAILABLE.String())
				defaultStaff.AutoCreateTimesheet = database.Bool(false)
				defaultStaff.StartDate = database.DateFromPb(startDate)
				defaultStaff.EndDate = database.DateFromPb(endDate)
				defaultStaff.ID = database.Text(userId)

				//user
				defaultUser := entity.LegacyUser{}
				database.AllNullEntity(&defaultUser)
				defaultUser.FullName = database.Text("John Doe")
				defaultUser.Email = database.Text("staff_manabie@manabie.com")
				defaultUser.Birthday = pgtype.Date{Time: birthday.AsTime(), Status: pgtype.Present}
				defaultUser.Gender = database.Text(gender.String())
				defaultUser.Country = database.Text(cpb.Country_COUNTRY_VN.String())
				defaultUser.Group = database.Text(constant.UserGroupSchoolAdmin)
				defaultUser.LastName = database.Text("John")
				defaultUser.FirstName = database.Text("Doe")
				defaultUser.LastNamePhonetic = database.Text("LastNamePhonetic")
				defaultUser.FirstNamePhonetic = database.Text("FirstNamePhonetic")
				defaultUser.FullNamePhonetic = database.Text("LastNamePhonetic FirstNamePhonetic")
				defaultUser.ResourcePath = database.Text(resourcePath)
				defaultUser.Remarks = database.Text("Staff Remarks")
				defaultUser.ExternalUserID = database.Text(externalUserID)
				defaultUser.ID = database.Text(userId)
				defaultUser.UserRole = database.Text(string(constant.UserRoleStaff))
				// Temporarily set loginEmail equal Email
				defaultUser.LoginEmail = database.Text("staff_manabie@manabie.com")
				defaultUser.UserName = database.Text(username)
				defaultStaff.LegacyUser = defaultUser
			},
			expectedRes: defaultStaff,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			t.Log(testCaseLog + testCase.name)
			testCase.Options.(func())()

			staff, err := createStaffPbToStaffEnt(testCase.req.(*pb.CreateStaffRequest_StaffProfile), userGroup, resourcePath)

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
			if testCase.expectedRes != nil {
				testCase.expectedRes.(*entity.Staff).ID = staff.ID
				testCase.expectedRes.(*entity.Staff).LegacyUser.ID = staff.LegacyUser.ID
				assert.Equal(t, *testCase.expectedRes.(*entity.Staff), *staff)
			}
		})
	}
}

func TestStaffService_getLegacyUserGroup(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)

	roleSchoolAdmin := repository.NewNullRole()
	roleSchoolAdmin.RoleAttribute.RoleName = field.NewString(constant.RoleSchoolAdmin)

	roleHQStaff := repository.NewNullRole()
	roleHQStaff.RoleAttribute.RoleName = field.NewString(constant.RoleHQStaff)

	roleTeacher := repository.NewNullRole()
	roleTeacher.RoleAttribute.RoleName = field.NewString(constant.RoleTeacher)

	roleTeacherLead := repository.NewNullRole()
	roleTeacherLead.RoleAttribute.RoleName = field.NewString(constant.RoleTeacherLead)

	roleRepo := new(mock_repositories.MockDomainRoleRepo)
	service := &StaffService{
		DB:       db,
		RoleRepo: roleRepo,
	}

	testCases := []TestCase{
		{
			name:        "user group ids in request empty",
			ctx:         ctx,
			req:         []string{},
			expectedRes: constant.UserGroupTeacher,
			expectedErr: nil,
		},
		{
			name: "can not query roles",
			ctx:  ctx,
			req:  []string{"id_1"},
			setup: func(ctx context.Context) {
				roleRepo.On("GetByUserGroupIDs", ctx, service.DB, mock.Anything).Once().Return(entity.DomainRoles{}, pgx.ErrNoRows)
			},
			expectedRes: "",
			expectedErr: fmt.Errorf("GetByUserGroupID failed: %w", pgx.ErrNoRows),
		},
		{
			name: "roles empty",
			ctx:  ctx,
			req:  []string{"id_1"},
			setup: func(ctx context.Context) {
				roleRepo.On("GetByUserGroupIDs", ctx, service.DB, mock.Anything).Once().Return(entity.DomainRoles{}, nil)

			},
			expectedRes: constant.UserGroupTeacher,
			expectedErr: nil,
		},
		{
			name: "only 1 role",
			ctx:  ctx,
			req:  []string{"id_1"},
			setup: func(ctx context.Context) {
				roleRepo.On("GetByUserGroupIDs", ctx, service.DB, mock.Anything).Once().Return(entity.DomainRoles{roleHQStaff}, nil)
			},
			expectedRes: constant.UserGroupSchoolAdmin,
			expectedErr: nil,
		},
		{
			name: "multiple roles (including mapping legacy user_group school_admin)",
			ctx:  ctx,
			req:  []string{"id_1"},
			setup: func(ctx context.Context) {
				roleRepo.On("GetByUserGroupIDs", ctx, service.DB, mock.Anything).Once().Return(entity.DomainRoles{roleHQStaff, roleSchoolAdmin, roleTeacher}, nil)
			},
			expectedRes: constant.UserGroupSchoolAdmin,
			expectedErr: nil,
		},
		{
			name: "multiple roles (only mapping legacy user_group teacher",
			ctx:  ctx,
			req:  []string{"id_1"},
			setup: func(ctx context.Context) {
				roleRepo.On("GetByUserGroupIDs", ctx, service.DB, mock.Anything).Once().Return(entity.DomainRoles{roleTeacherLead, roleTeacher}, nil)
			},
			expectedRes: constant.UserGroupTeacher,
			expectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log(testCaseLog + testCase.name)

			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}

			legacyUserGroup, err := service.getLegacyUserGroup(ctx, service.DB, testCase.req.([]string))
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedRes, legacyUserGroup)
		})
	}
}

func TestStaffService_ValidateStaffTags(t *testing.T) {
	id1 := idutil.ULIDNow()
	id2 := idutil.ULIDNow()

	type args struct {
		tagIDs       []string
		existingTags entity.DomainTags
	}
	tests := []struct {
		name string
		args func(t *testing.T) args

		wantErr    bool
		inspectErr func(err error, t *testing.T)
	}{
		{
			name: "happy case: valid staff tags",
			args: func(t *testing.T) args {
				return args{
					tagIDs: []string{id1, id2},
					existingTags: entity.DomainTags([]entity.DomainTag{
						createMockDomainTagWithTypeAndIsArchived(id1, pb.UserTagType_USER_TAG_TYPE_STAFF, false),
						createMockDomainTagWithTypeAndIsArchived(id2, pb.UserTagType_USER_TAG_TYPE_STAFF, false),
					}),
				}
			},
			wantErr: false,
		},
		{
			name: "happy case: valid with empty staff tags",
			args: func(t *testing.T) args {
				return args{
					tagIDs:       []string{},
					existingTags: entity.DomainTags([]entity.DomainTag{}),
				}
			},
			wantErr: false,
		},
		{
			name: "error: tag is not for staff",
			args: func(t *testing.T) args {
				return args{
					tagIDs: []string{id1, id2},
					existingTags: entity.DomainTags([]entity.DomainTag{
						createMockDomainTagWithTypeAndIsArchived(id1, pb.UserTagType_USER_TAG_TYPE_STAFF, false),
						createMockDomainTagWithTypeAndIsArchived(id2, pb.UserTagType_USER_TAG_TYPE_PARENT, false),
					}),
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				assert.Equal(t, err, ErrTagIsNotForStaff)
			},
		},
		{
			name: "error: tag is not existed",
			args: func(t *testing.T) args {
				return args{
					tagIDs: []string{id1, id2},
					existingTags: entity.DomainTags([]entity.DomainTag{
						createMockDomainTagWithTypeAndIsArchived(id2, pb.UserTagType_USER_TAG_TYPE_STAFF, false),
					}),
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				assert.Equal(t, err, ErrTagIDsMustBeExisted)
			},
		},
		{
			name: "error: tag is archived",
			args: func(t *testing.T) args {
				return args{
					tagIDs: []string{id1},
					existingTags: entity.DomainTags([]entity.DomainTag{
						createMockDomainTagWithTypeAndIsArchived(id1, pb.UserTagType_USER_TAG_TYPE_STAFF, true),
					}),
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				assert.Equal(t, err, ErrTagIsArchived)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			err := validateStaffTags(tArgs.tagIDs, tArgs.existingTags)

			if (err != nil) != tt.wantErr {
				t.Fatalf("validateStaffTags error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}

func TestStaffService_validateStaffUsername(t *testing.T) {
	type args struct {
		username string
	}

	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tx := new(mock_database.Tx)

	userRepo := new(mock_repositories.MockUserRepo)
	domainUserRepo := new(mock_repositories.MockDomainUserRepo)

	umsvc := &usvc.UserModifierService{
		DB:             tx,
		DomainUserRepo: domainUserRepo,
		UserRepo:       userRepo,
	}

	s := &StaffService{
		DB:                  tx,
		UserModifierService: umsvc,
		UserRepo:            domainUserRepo,
	}

	testCases := []TestCase{
		{
			name: "valid username",
			ctx:  ctx,
			req: entity_mock.User{
				RandomUser: entity_mock.RandomUser{
					UserName: field.NewString("validusername"),
				},
			},
			setup: func(ctx context.Context) {
				domainUserRepo.On("GetByUserNames", ctx, tx, []string{"validusername"}).Return(entity.Users{}, nil)
			},
			expectedErr: nil,
		},
		{
			name: "empty username",
			ctx:  ctx,
			req: entity_mock.User{
				RandomUser: entity_mock.RandomUser{
					UserName: field.NewString("  "),
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, entity.MissingMandatoryFieldError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      -1,
			}.Error()),
		},
		{
			name: "invalid username",
			ctx:  ctx,
			req: entity_mock.User{
				RandomUser: entity_mock.RandomUser{
					UserName: field.NewString("invalid username"),
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, entity.InvalidFieldError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      -1,
				Reason:     entity.NotMatchingPattern,
			}.Error()),
		},
		{
			name: "username existed in DB",
			ctx:  ctx,
			req: entity_mock.User{
				RandomUser: entity_mock.RandomUser{
					UserName: field.NewString("existedusername"),
				},
			},
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
			expectedErr: status.Error(codes.AlreadyExists, entity.ExistingDataError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      0,
			}.Error()),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log(testCaseLog + testCase.name)

			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}

			err := s.validateStaffUsername(ctx, s.DB, testCase.req.(entity.User))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}
