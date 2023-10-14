package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	internal_auth_user "github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/constants"
	libdatabase "github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/port/grpc"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	mock_usermgmt "github.com/manabie-com/backend/internal/usermgmt/pkg/mock"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	mock_fatima "github.com/manabie-com/backend/mock/fatima/services"
	mock_multitenant "github.com/manabie-com/backend/mock/golibs/auth/multitenant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_clients "github.com/manabie-com/backend/mock/lessonmgmt/zoom/clients"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	mock_service "github.com/manabie-com/backend/mock/usermgmt/service"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func DomainStudentServiceMock() (prepareDomainStudentMock, DomainStudent) {
	m := prepareDomainStudentMock{
		&mock_database.Ext{},
		&mock_database.Tx{},
		&mock_multitenant.TenantClient{},
		&mock_nats.JetStreamManagement{},
		&mock_multitenant.TenantClient{},
		&mock_multitenant.TenantManager{},
		&mock_fatima.SubscriptionModifierServiceClient{},
		&mock_repositories.OrganizationRepo{},
		&mock_repositories.MockDomainStudentRepo{},
		&mock_repositories.MockDomainUserRepo{},
		&mock_repositories.MockDomainUserGroupRepo{},
		&mock_repositories.MockDomainUserAddressRepo{},
		&mock_repositories.MockDomainUserPhoneNumberRepo{},
		&mock_repositories.MockDomainSchoolHistoryRepo{},
		&mock_repositories.MockDomainLocationRepo{},
		&mock_repositories.MockDomainGradeRepo{},
		&mock_repositories.MockDomainSchoolRepo{},
		&mock_repositories.MockDomainSchoolCourseRepo{},
		&mock_repositories.MockDomainPrefectureRepo{},
		&mock_repositories.MockDomainUsrEmailRepo{},
		&mock_repositories.MockDomainEnrollmentStatusHistoryRepo{},
		&mock_repositories.MockDomainUserAccessPathRepo{},
		&mock_repositories.MockDomainTagRepo{},
		&mock_repositories.MockDomainTaggedUserRepo{},
		&mock_repositories.MockDomainStudentPackageRepo{},
		&mock_repositories.MockDomainStudentParentRelationshipRepo{},
		&mock_repositories.MockDomainInternalConfigurationRepo{},
		&mock_service.MockDomainParent{},
		&mock_clients.MockConfigurationClient{},
		nil,
		nil,
		&mock_unleash_client.UnleashClientInstance{},
		&mock_service.MockStudentValidationManager{},
	}

	service := DomainStudent{
		DB:                               m.db,
		JSM:                              m.jsm,
		FirebaseAuthClient:               m.firebaseAuthClient,
		TenantManager:                    m.tenantManager,
		FatimaClient:                     m.fatimaClient,
		OrganizationRepo:                 m.OrganizationRepo,
		StudentRepo:                      m.studentRepo,
		UserRepo:                         m.userRepo,
		UserGroupRepo:                    m.userGroupRepo,
		UserAddressRepo:                  m.userAddressRepo,
		UserPhoneNumberRepo:              m.userPhoneNumberRepo,
		SchoolHistoryRepo:                m.schoolHistoryRepo,
		LocationRepo:                     m.locationRepo,
		GradeRepo:                        m.gradeRepo,
		SchoolRepo:                       m.schoolRepo,
		SchoolCourseRepo:                 m.schoolCourseRepo,
		PrefectureRepo:                   m.prefectureRepo,
		UsrEmailRepo:                     m.usrEmailRepo,
		TagRepo:                          m.tagRepo,
		TaggedUserRepo:                   m.taggedUserRepo,
		EnrollmentStatusHistoryRepo:      m.enrollmentStatusHistoryRepo,
		UserAccessPathRepo:               m.UserAccessPathRepo,
		StudentPackage:                   m.studentPackage,
		StudentParentRepo:                m.StudentParentRepo,
		InternalConfigurationRepo:        m.internalConfigurationRepo,
		DomainParentService:              m.parentService,
		ConfigurationClient:              m.configurationClient,
		StudentParentRelationshipManager: nil,
		UnleashClient:                    m.unleashClient,
		StudentValidationManager:         m.studentValidationManager,
	}
	return m, service
}

type prepareDomainStudentMock struct {
	db                               *mock_database.Ext
	tx                               *mock_database.Tx
	firebaseAuthClient               *mock_multitenant.TenantClient
	jsm                              *mock_nats.JetStreamManagement
	tenantClient                     *mock_multitenant.TenantClient
	tenantManager                    *mock_multitenant.TenantManager
	fatimaClient                     *mock_fatima.SubscriptionModifierServiceClient
	OrganizationRepo                 *mock_repositories.OrganizationRepo
	studentRepo                      *mock_repositories.MockDomainStudentRepo
	userRepo                         *mock_repositories.MockDomainUserRepo
	userGroupRepo                    *mock_repositories.MockDomainUserGroupRepo
	userAddressRepo                  *mock_repositories.MockDomainUserAddressRepo
	userPhoneNumberRepo              *mock_repositories.MockDomainUserPhoneNumberRepo
	schoolHistoryRepo                *mock_repositories.MockDomainSchoolHistoryRepo
	locationRepo                     *mock_repositories.MockDomainLocationRepo
	gradeRepo                        *mock_repositories.MockDomainGradeRepo
	schoolRepo                       *mock_repositories.MockDomainSchoolRepo
	schoolCourseRepo                 *mock_repositories.MockDomainSchoolCourseRepo
	prefectureRepo                   *mock_repositories.MockDomainPrefectureRepo
	usrEmailRepo                     *mock_repositories.MockDomainUsrEmailRepo
	enrollmentStatusHistoryRepo      *mock_repositories.MockDomainEnrollmentStatusHistoryRepo
	UserAccessPathRepo               *mock_repositories.MockDomainUserAccessPathRepo
	tagRepo                          *mock_repositories.MockDomainTagRepo
	taggedUserRepo                   *mock_repositories.MockDomainTaggedUserRepo
	studentPackage                   *mock_repositories.MockDomainStudentPackageRepo
	StudentParentRepo                *mock_repositories.MockDomainStudentParentRelationshipRepo
	internalConfigurationRepo        *mock_repositories.MockDomainInternalConfigurationRepo
	parentService                    *mock_service.MockDomainParent
	configurationClient              *mock_clients.MockConfigurationClient
	studentParentRelationshipManager StudentParentRelationshipManager
	authUserUpserter                 AuthUserUpserter
	unleashClient                    *mock_unleash_client.UnleashClientInstance
	studentValidationManager         *mock_service.MockStudentValidationManager
}

type randomUserAccessPath struct {
	entity.DefaultUserAccessPath
}

func (randomUserAccessPath randomUserAccessPath) LocationID() field.String {
	return field.NewString(idutil.ULIDNow())
}

func TestDomainStudent_UpsertMultiple(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	domainStudent := &mock_usermgmt.Student{
		RandomStudent: mock_usermgmt.RandomStudent{
			GradeID:          field.NewString("grade-id-1"),
			Email:            field.NewString("test@manabie.com"),
			UserName:         field.NewString("username"),
			Gender:           field.NewString(upb.Gender_FEMALE.String()),
			FirstName:        field.NewString("test first name"),
			LastName:         field.NewString("test last name"),
			ExternalUserID:   field.NewString("external-user-id"),
			CurrentGrade:     field.NewInt16(1),
			EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
			LoginEmail:       field.NewString("login-email"),
		},
	}

	domainStudent2 := &mock_usermgmt.Student{
		RandomStudent: mock_usermgmt.RandomStudent{
			GradeID:          field.NewString("grade-id-2"),
			Email:            field.NewString("test2@manabie.com"),
			Gender:           field.NewString(upb.Gender_FEMALE.String()),
			UserName:         field.NewString("testUsername2"),
			FirstName:        field.NewString("test first name"),
			LastName:         field.NewString("test last name"),
			ExternalUserID:   field.NewString("external-user-id2"),
			CurrentGrade:     field.NewInt16(1),
			EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
			LoginEmail:       field.NewString("login-email"),
		},
	}

	existingDomainStudent := &mock_usermgmt.Student{
		RandomStudent: mock_usermgmt.RandomStudent{
			GradeID:          field.NewString("grade-id-1"),
			UserID:           field.NewString("student-id"),
			Email:            field.NewString("test@manabie.com"),
			UserName:         field.NewString("existStudentUsername"),
			Gender:           field.NewString(upb.Gender_FEMALE.String()),
			FirstName:        field.NewString("test first name"),
			LastName:         field.NewString("test last name"),
			ExternalUserID:   field.NewString("external-user-id"),
			CurrentGrade:     field.NewInt16(1),
			EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
			LoginEmail:       field.NewString("login-email"),
		},
	}
	testCases := []TestCase{
		{
			name: "happy case: create students",
			ctx:  ctx,
			req: []aggregate.DomainStudent{
				{
					DomainStudent:    domainStudent,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.usrEmailRepo.On("CreateMultiple", ctx, domainStudentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainStudentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainStudentMock.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, []string{"username"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainStudentMock.db, constant.RoleStudent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainStudentMock.locationRepo.On("RetrieveLowestLevelLocations", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainLocations{mock_usermgmt.NewLocation("", "")}, nil)
				domainStudentMock.studentPackage.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainStudentPackages{}, nil)
				domainStudentMock.db.On("Begin", ctx).Return(domainStudentMock.tx, nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
				domainStudentMock.locationRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainStudentMock.studentRepo.On("UpsertMultiple", ctx, domainStudentMock.tx, mock.Anything, mock.Anything).Return(nil)
				domainStudentMock.OrganizationRepo.On("GetTenantIDByOrgID", ctx, domainStudentMock.tx, mock.Anything).Return("", nil)
				hashConfig := mockScryptHash()
				domainStudentMock.tenantManager.On("TenantClient", ctx, mock.Anything).Return(domainStudentMock.tenantClient, nil)
				domainStudentMock.tenantClient.On("GetHashConfig").Return(hashConfig)
				domainStudentMock.tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Return(&internal_auth_user.ImportUsersResult{}, nil)
				domainStudentMock.authUserUpserter = func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) (entity.LegacyUsers, error) {
					return entity.LegacyUsers{}, nil
				}
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
				domainStudentMock.jsm.On("TracedPublish", mock.Anything, "publishDomainUserEvent", mock.Anything, mock.Anything).Return(nil, nil)
				domainStudentMock.tx.On("Commit", ctx).Return(nil)
				// upsertEnrollmentStatusHistories
				domainStudentMock.unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureUsingMasterReplicatedTable, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainStudentMock.internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(entity.NullDomainConfiguration{}, nil)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("UpdateStudentStatusBasedEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "happy case: update students",
			ctx:  ctx,
			req: []aggregate.DomainStudent{
				{
					DomainStudent:    existingDomainStudent,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.usrEmailRepo.On("CreateMultiple", ctx, domainStudentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainStudentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainStudentMock.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, []string{"existstudentusername"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, []string{"student-id"}).Return(entity.Users{domainStudent}, nil)
				domainStudentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainStudentMock.db, constant.RoleStudent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainStudentMock.locationRepo.On("RetrieveLowestLevelLocations", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainLocations{mock_usermgmt.NewLocation("", "")}, nil)
				domainStudentMock.studentPackage.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainStudentPackages{}, nil)
				domainStudentMock.db.On("Begin", ctx).Return(domainStudentMock.tx, nil)
				domainStudentMock.usrEmailRepo.On("UpdateEmail", ctx, domainStudentMock.tx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
				domainStudentMock.locationRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainStudentMock.studentRepo.On("UpsertMultiple", ctx, domainStudentMock.tx, mock.Anything, mock.Anything).Return(nil)
				domainStudentMock.OrganizationRepo.On("GetTenantIDByOrgID", ctx, domainStudentMock.tx, mock.Anything).Return("", nil)
				domainStudentMock.tenantManager.On("TenantClient", ctx, mock.Anything).Return(domainStudentMock.tenantClient, nil)
				domainStudentMock.tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				domainStudentMock.tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				domainStudentMock.authUserUpserter = func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) (entity.LegacyUsers, error) {
					return entity.LegacyUsers{}, nil
				}
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
				domainStudentMock.jsm.On("TracedPublish", mock.Anything, "publishDomainUserEvent", mock.Anything, mock.Anything).Return(nil, nil)
				domainStudentMock.tx.On("Commit", ctx).Return(nil)
				// upsertEnrollmentStatusHistories
				domainStudentMock.unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureUsingMasterReplicatedTable, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainStudentMock.internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(entity.NullDomainConfiguration{}, nil)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("UpdateStudentStatusBasedEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "service can not get existing enrollment status histories of students to send event",
			ctx:  ctx,
			req: []aggregate.DomainStudent{
				{
					DomainStudent:    existingDomainStudent,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
			},
			expectedErr: errcode.Error{
				Code: errcode.InternalError,
				Err:  errors.Wrap(assert.AnError, "service.EnrollmentStatusHistoryRepo.GetByStudentIDs"),
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.usrEmailRepo.On("CreateMultiple", ctx, domainStudentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainStudentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainStudentMock.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, []string{"existstudentusername"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, []string{"student-id"}).Return(entity.Users{domainStudent}, nil)
				domainStudentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainStudentMock.db, constant.RoleStudent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainStudentMock.locationRepo.On("RetrieveLowestLevelLocations", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainLocations{mock_usermgmt.NewLocation("", "")}, nil)
				domainStudentMock.studentPackage.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainStudentPackages{}, nil)
				domainStudentMock.db.On("Begin", ctx).Return(domainStudentMock.tx, nil)
				domainStudentMock.usrEmailRepo.On("UpdateEmail", ctx, domainStudentMock.tx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
				domainStudentMock.locationRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainStudentMock.studentRepo.On("UpsertMultiple", ctx, domainStudentMock.tx, mock.Anything, mock.Anything).Return(nil)
				domainStudentMock.OrganizationRepo.On("GetTenantIDByOrgID", ctx, domainStudentMock.tx, mock.Anything).Return("", nil)
				domainStudentMock.tenantManager.On("TenantClient", ctx, mock.Anything).Return(domainStudentMock.tenantClient, nil)
				domainStudentMock.tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				domainStudentMock.tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				domainStudentMock.authUserUpserter = func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) (entity.LegacyUsers, error) {
					return entity.LegacyUsers{}, nil
				}

				// upsertEnrollmentStatusHistories
				domainStudentMock.unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureUsingMasterReplicatedTable, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainStudentMock.internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(entity.NullDomainConfiguration{}, nil)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("UpdateStudentStatusBasedEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{},
					errcode.Error{
						Code: errcode.InternalError,
						Err:  errors.Wrap(assert.AnError, "service.EnrollmentStatusHistoryRepo.GetByStudentIDs"),
					},
				)
				domainStudentMock.tx.On("Rollback", ctx).Return(nil)
			},
		},
		{
			name: "bad case: missing username in payload",
			ctx:  ctx,
			expectedErr: entity.MissingMandatoryFieldError{
				Index:      0,
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
			},
			req: []aggregate.DomainStudent{
				{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							GradeID:          field.NewString("grade-id-1"),
							Email:            field.NewString("test@manabie.com"),
							Gender:           field.NewString(upb.Gender_FEMALE.String()),
							FirstName:        field.NewString("test first name"),
							LastName:         field.NewString("test last name"),
							ExternalUserID:   field.NewString("external-user-id"),
							CurrentGrade:     field.NewInt16(1),
							EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
							LoginEmail:       field.NewString("login-email"),
						},
					},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
			},
		},
		{
			name: "bad case: duplicated username in payload",
			ctx:  ctx,
			expectedErr: entity.DuplicatedFieldError{
				EntityName:      entity.UserEntity,
				DuplicatedField: string(entity.UserFieldUserName),
				Index:           1,
			},
			req: []aggregate.DomainStudent{
				{
					DomainStudent: domainStudent,
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
				{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							GradeID:          field.NewString("grade-id-2"),
							Email:            field.NewString("test2@manabie.com"),
							Gender:           field.NewString(upb.Gender_FEMALE.String()),
							UserName:         field.NewString("USERNAME"),
							FirstName:        field.NewString("test first name"),
							LastName:         field.NewString("test last name"),
							ExternalUserID:   field.NewString("external-user-id2"),
							CurrentGrade:     field.NewInt16(1),
							EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
							LoginEmail:       field.NewString("login-email"),
						},
					},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
			},
		},
		{
			name: "bad case: duplicated email in payload",
			ctx:  ctx,
			expectedErr: entity.DuplicatedFieldError{
				EntityName:      entity.UserEntity,
				DuplicatedField: string(entity.UserFieldEmail),
				Index:           1,
			},
			req: []aggregate.DomainStudent{
				{
					DomainStudent: domainStudent,
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
				{
					DomainStudent: existingDomainStudent,
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
			},
		},
		{
			name: "bad case: duplicated email in payload (case-insensitive)",
			ctx:  ctx,
			expectedErr: entity.DuplicatedFieldError{
				EntityName:      entity.UserEntity,
				DuplicatedField: string(entity.UserFieldEmail),
				Index:           1,
			},
			req: []aggregate.DomainStudent{
				{
					DomainStudent: domainStudent,
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
				{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							GradeID:          field.NewString("grade-id-1"),
							UserID:           field.NewString("student-id"),
							Email:            field.NewString("TEST@manabie.com"),
							UserName:         field.NewString("username2"),
							Gender:           field.NewString(upb.Gender_FEMALE.String()),
							FirstName:        field.NewString("test first name"),
							LastName:         field.NewString("test last name"),
							ExternalUserID:   field.NewString("external-user-id"),
							CurrentGrade:     field.NewInt16(1),
							EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
						},
					},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
			},
		},
		{
			name: "bad case: duplicated user_id in payload",
			ctx:  ctx,
			expectedErr: entity.DuplicatedFieldError{
				EntityName:      entity.UserEntity,
				DuplicatedField: string(entity.UserFieldUserID),
				Index:           1,
			},
			req: []aggregate.DomainStudent{
				{
					DomainStudent: existingDomainStudent,
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
				{
					DomainStudent: existingDomainStudent,
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
			},
		},
		{
			name: "bad case: duplicated external_user_id in payload",
			ctx:  ctx,
			expectedErr: entity.DuplicatedFieldError{
				EntityName:      entity.UserEntity,
				DuplicatedField: string(entity.UserFieldExternalUserID),
				Index:           1,
			},
			req: []aggregate.DomainStudent{
				{
					DomainStudent: domainStudent,
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
				{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							GradeID:          field.NewString("grade-id-1"),
							UserID:           field.NewString("student-id"),
							Email:            field.NewString("TEST2@manabie.com"),
							UserName:         field.NewString("username2"),
							Gender:           field.NewString(upb.Gender_FEMALE.String()),
							FirstName:        field.NewString("test first name"),
							LastName:         field.NewString("test last name"),
							ExternalUserID:   field.NewString("external-user-id"),
							CurrentGrade:     field.NewInt16(1),
							EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
						},
					},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
			},
		},
		{
			name: "bad case: invalid tag type",
			ctx:  ctx,
			expectedErr: entity.InvalidFieldError{
				EntityName: entity.StudentEntity,
				FieldName:  entity.StudentTagsField,
				Reason:     entity.InvalidTagType,
				Index:      0,
			},
			req: []aggregate.DomainStudent{
				{
					DomainStudent: domainStudent,
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
					TaggedUsers: entity.DomainTaggedUsers{
						&entity.EmptyDomainTaggedUser{},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainStudentMock.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, []string{"username"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.usrEmailRepo.On("CreateMultiple", ctx, domainStudentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainStudentMock.db, constant.RoleStudent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
				domainStudentMock.locationRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainStudentMock.tagRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainTags{&repository.Tag{
					TagAttribute: repository.TagAttribute{
						TagType: field.NewString(entity.UserTagTypeParent),
					},
				}}, nil)
			},
		},
		{
			name: "bad case: invalid contact_preference type",
			ctx:  ctx,
			expectedErr: entity.InvalidFieldError{
				FieldName:  entity.StudentFieldContactPreference,
				EntityName: entity.UserEntity,
				Index:      0,
				Reason:     entity.NotMatchingEnum,
			},
			req: []aggregate.DomainStudent{
				{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							Email:             field.NewString("test@manabie.com"),
							UserName:          field.NewString("username"),
							Gender:            field.NewString(upb.Gender_FEMALE.String()),
							FirstName:         field.NewString("test first name"),
							LastName:          field.NewString("test last name"),
							CurrentGrade:      field.NewInt16(1),
							EnrollmentStatus:  field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
							ContactPreference: field.NewString("invalid-contact-preference"),
						},
					},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.usrEmailRepo.On("CreateMultiple", ctx, domainStudentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainStudentMock.db, constant.RoleStudent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
				domainStudentMock.locationRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
			},
		},
		{
			name: "bad case: location is archived",
			ctx:  ctx,
			expectedErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: "students[0].locations[0]",
			},
			req: []aggregate.DomainStudent{
				{
					DomainStudent:    domainStudent,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainStudentMock.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, []string{"username"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.usrEmailRepo.On("CreateMultiple", ctx, domainStudentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainStudentMock.db, constant.RoleStudent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainStudentMock.locationRepo.On("RetrieveLowestLevelLocations", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainLocations{mock_usermgmt.NewLocation("", "")}, nil)
				domainStudentMock.studentPackage.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainStudentPackages{}, nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
				domainStudentMock.locationRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainLocations{&repository.Location{
					LocationAttribute: repository.LocationAttribute{
						IsArchived: field.NewBoolean(true),
					},
				}}, nil)
			},
		},
		{
			name: "bad case: username is duplicated in system",
			ctx:  ctx,
			expectedErr: entity.ExistingDataError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      0,
			},
			req: []aggregate.DomainStudent{
				{
					DomainStudent:    domainStudent2,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userRepo.On("GetByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id2"}).Return(entity.Users{}, nil)
				domainStudentMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id2"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByEmailsInsensitiveCase", mock.Anything, domainStudentMock.db, []string{"test2@manabie.com"}).Return(entity.Users{domainStudent2}, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{&mock_usermgmt.Student{
					RandomStudent: mock_usermgmt.RandomStudent{
						UserID:   field.NewString(idutil.ULIDNow()),
						UserName: domainStudent2.UserName(),
					}}}, nil)
			},
		},
		{
			name: "bad case: email is duplicated in system",
			ctx:  ctx,
			expectedErr: entity.ExistingDataError{
				FieldName:  string(entity.UserFieldEmail),
				EntityName: entity.UserEntity,
				Index:      0,
			},
			req: []aggregate.DomainStudent{
				{
					DomainStudent:    domainStudent,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userRepo.On("GetByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainStudentMock.db, []string{"test@manabie.com"}).Return(entity.Users{&mock_usermgmt.Student{
					RandomStudent: mock_usermgmt.RandomStudent{
						UserID: field.NewString("existed-user-id"),
						Email:  field.NewString("test@manabie.com"),
					}}}, nil)
			},
		},
		{
			name: "bad case: location is empty in payload",
			ctx:  ctx,
			expectedErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: "students[0].enrollment_status_histories[0].location",
			},
			req: []aggregate.DomainStudent{
				{
					DomainStudent:    domainStudent,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
							},
						},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainStudentMock.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, []string{"username"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.usrEmailRepo.On("CreateMultiple", ctx, domainStudentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainStudentMock.db, constant.RoleStudent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainStudentMock.locationRepo.On("RetrieveLowestLevelLocations", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainLocations{mock_usermgmt.NewLocation("", "")}, nil)
				domainStudentMock.studentPackage.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainStudentPackages{}, nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
			},
		},
		{
			name: "bad case: location is duplicated in payload",
			ctx:  ctx,
			expectedErr: errcode.Error{
				Code:      errcode.DuplicatedData,
				FieldName: "students[0].enrollment_status_histories[1].location",
			},
			req: []aggregate.DomainStudent{
				{
					DomainStudent:    domainStudent,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								LocationID:       field.NewString("location-id"),
								EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
							},
						},
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								LocationID:       field.NewString("location-id"),
								EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
							},
						},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainStudentMock.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, []string{"username"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.usrEmailRepo.On("CreateMultiple", ctx, domainStudentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainStudentMock.db, constant.RoleStudent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainStudentMock.locationRepo.On("RetrieveLowestLevelLocations", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainLocations{mock_usermgmt.NewLocation("", "")}, nil)
				domainStudentMock.studentPackage.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainStudentPackages{}, nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
			},
		},
		{
			name: "bad case: school is empty in payload",
			ctx:  ctx,
			expectedErr: entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					EntityName: entity.StudentEntity,
					FieldName:  entity.StudentSchoolField,
					Index:      0,
				},
				NestedFieldName: entity.StudentSchoolHistoryField,
				NestedIndex:     0,
			},
			req: []aggregate.DomainStudent{
				{
					DomainStudent:    domainStudent,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
					SchoolHistories: entity.DomainSchoolHistories{
						mock_usermgmt.SchoolHistory{
							RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{},
						},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainStudentMock.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, []string{"username"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
			},
		},
		{
			name: "bad case: school is duplicated in payload",
			ctx:  ctx,
			expectedErr: entity.DuplicatedFieldErrorWithArrayNestedField{
				DuplicatedFieldError: entity.DuplicatedFieldError{
					EntityName:      entity.StudentEntity,
					DuplicatedField: entity.StudentSchoolField,
					Index:           0,
				},
				NestedFieldName: entity.StudentSchoolHistoryField,
				NestedIndex:     1,
			},
			req: []aggregate.DomainStudent{
				{
					DomainStudent:    domainStudent,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
					SchoolHistories: entity.DomainSchoolHistories{
						mock_usermgmt.SchoolHistory{
							RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
								SchoolID: field.NewString("school-id"),
							},
						},
						mock_usermgmt.SchoolHistory{
							RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
								SchoolID: field.NewString("school-id"),
							},
						},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainStudentMock.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, []string{"username"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
			},
		},
		{
			name: "bad case: school course is duplicated in payload",
			ctx:  ctx,
			expectedErr: entity.DuplicatedFieldErrorWithArrayNestedField{
				DuplicatedFieldError: entity.DuplicatedFieldError{
					EntityName:      entity.StudentEntity,
					DuplicatedField: entity.StudentSchoolCourseField,
					Index:           0,
				},
				NestedFieldName: entity.StudentSchoolHistoryField,
				NestedIndex:     1,
			},
			req: []aggregate.DomainStudent{
				{
					DomainStudent:    domainStudent,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
					SchoolHistories: entity.DomainSchoolHistories{
						mock_usermgmt.SchoolHistory{
							RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
								SchoolID:       field.NewString("school-id-1"),
								SchoolCourseID: field.NewString("school-course-id-1"),
							},
						},
						mock_usermgmt.SchoolHistory{
							RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
								SchoolID:       field.NewString("school-id-2"),
								SchoolCourseID: field.NewString("school-course-id-1"),
							},
						},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainStudentMock.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, []string{"username"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
			},
		},
		{
			name: "bad case: deactivate and reactivate students error",
			ctx:  ctx,
			expectedErr: errcode.Error{
				Code: errcode.InternalError},
			req: []aggregate.DomainStudent{
				{
					DomainStudent:    domainStudent,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					// UserAccessPaths: entity.DomainUserAccessPaths{
					// 	entity.DefaultUserAccessPath{},
					// },
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								LocationID:       field.NewString("location-id"),
								EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NON_POTENTIAL.String()),
							},
						},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.usrEmailRepo.On("CreateMultiple", ctx, domainStudentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainStudentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainStudentMock.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, []string{"username"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, []string{"student-id"}).Return(entity.Users{domainStudent}, nil)
				domainStudentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainStudentMock.db, constant.RoleStudent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainStudentMock.locationRepo.On("RetrieveLowestLevelLocations", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainLocations{mock_usermgmt.NewLocation("location-id", "location-partner-id")}, nil)
				domainStudentMock.studentPackage.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainStudentPackages{}, nil)
				domainStudentMock.db.On("Begin", ctx).Return(domainStudentMock.tx, nil)
				domainStudentMock.usrEmailRepo.On("UpdateEmail", ctx, domainStudentMock.tx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
				domainStudentMock.locationRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainLocations{mock_usermgmt.Location{
					LocationIDAttr: field.NewString("location-id"),
				}}, nil)
				domainStudentMock.studentRepo.On("UpsertMultiple", ctx, domainStudentMock.tx, mock.Anything, mock.Anything).Return(nil)
				domainStudentMock.OrganizationRepo.On("GetTenantIDByOrgID", ctx, domainStudentMock.tx, mock.Anything).Return("", nil)
				domainStudentMock.tenantManager.On("TenantClient", ctx, mock.Anything).Return(domainStudentMock.tenantClient, nil)
				domainStudentMock.tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				domainStudentMock.tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				domainStudentMock.authUserUpserter = func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) (entity.LegacyUsers, error) {
					return entity.LegacyUsers{}, nil
				}
				// upsertEnrollmentStatusHistories
				domainStudentMock.unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureUsingMasterReplicatedTable, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainStudentMock.internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(entity.NullDomainConfiguration{}, nil)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					[]entity.DomainEnrollmentStatusHistory{}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					[]entity.DomainEnrollmentStatusHistory{}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("UpdateStudentStatusBasedEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errcode.Error{
					Code: errcode.InternalError})
				domainStudentMock.tx.On("Rollback", ctx).Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log(testCase.name)
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.ManabieSchool),
				},
			}
			testCase.ctx = interceptors.ContextWithJWTClaims(ctx, claim)
			testCase.ctx = interceptors.ContextWithUserID(testCase.ctx, "user-id")

			m, service := DomainStudentServiceMock()
			m.userPhoneNumberRepo.On("SoftDeleteByUserIDs", testCase.ctx, m.tx, mock.Anything).Return(nil)
			m.userAddressRepo.On("SoftDeleteByUserIDs", testCase.ctx, m.tx, mock.Anything).Return(nil)
			m.taggedUserRepo.On("SoftDeleteByUserIDs", testCase.ctx, m.tx, mock.Anything).Return(nil)
			m.schoolHistoryRepo.On("SoftDeleteByStudentIDs", testCase.ctx, m.tx, mock.Anything).Return(nil)
			testCase.setupWithMock(testCase.ctx, &m)
			service.AuthUserUpserter = m.authUserUpserter

			option := unleash.DomainStudentFeatureOption{
				DomainUserFeatureOption: unleash.DomainUserFeatureOption{
					EnableIgnoreUpdateEmail: false,
					EnableUsername:          true,
				},
				EnableAutoDeactivateAndReactivateStudentV2:            true,
				DisableAutoDeactivateAndReactivateStudent:             false,
				EnableExperimentalBulkInsertEnrollmentStatusHistories: false,
			}
			if testCase.option != nil {
				option = testCase.option.(unleash.DomainStudentFeatureOption)
			}
			_, err := service.UpsertMultiple(testCase.ctx, option, testCase.req.([]aggregate.DomainStudent)...)
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestDomainStudent_UpsertMultipleWithErrorCollection(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	domainStudentCreate := &mock_usermgmt.Student{
		RandomStudent: mock_usermgmt.RandomStudent{
			GradeID:          field.NewString("grade-id-1"),
			Email:            field.NewString("test@manabie.com"),
			UserName:         field.NewString("username1"),
			Gender:           field.NewString(upb.Gender_FEMALE.String()),
			FirstName:        field.NewString("test first name"),
			LastName:         field.NewString("test last name"),
			ExternalUserID:   field.NewString("external-user-id"),
			CurrentGrade:     field.NewInt16(1),
			EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusPotential),
		},
	}
	domainStudentUpdate := &mock_usermgmt.Student{
		RandomStudent: mock_usermgmt.RandomStudent{
			UserID:           field.NewString("user-id"),
			GradeID:          field.NewString("grade-id-1"),
			Email:            field.NewString("test@manabie.com"),
			UserName:         field.NewString("username2"),
			Gender:           field.NewString(upb.Gender_FEMALE.String()),
			FirstName:        field.NewString("test first name"),
			LastName:         field.NewString("test last name"),
			ExternalUserID:   field.NewString("external-user-id"),
			CurrentGrade:     field.NewInt16(1),
			EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusPotential),
		},
	}

	type MultipleTestCase struct {
		name               string
		ctx                context.Context
		studentWithIndexes aggregate.DomainStudents
		setup              func(ctx context.Context)
		setupWithMock      func(ctx context.Context, mockInterface interface{})
		wantStudents       aggregate.DomainStudents
		wantErrors         []error
	}

	testCases := []MultipleTestCase{
		{
			name: "happy case: create students with error collection",
			ctx:  ctx,
			studentWithIndexes: aggregate.DomainStudents{
				aggregate.DomainStudent{
					DomainStudent:    domainStudentCreate,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
					IndexAttr: 0,
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.studentValidationManager.On("FullyValidate", ctx, domainStudentMock.db, mock.Anything, mock.Anything).Once().Return(
					aggregate.DomainStudents{
						aggregate.DomainStudent{
							DomainStudent: domainStudentCreate,
						},
					},
					aggregate.DomainStudents{},
					[]error{},
				)
				domainStudentMock.usrEmailRepo.On("CreateMultiple", ctx, domainStudentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainStudentMock.db.On("Begin", ctx).Return(domainStudentMock.tx, nil)
				domainStudentMock.studentRepo.On("UpsertMultiple", ctx, domainStudentMock.tx, mock.Anything, mock.Anything).Return(nil)
				domainStudentMock.authUserUpserter = func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) (entity.LegacyUsers, error) {
					return entity.LegacyUsers{}, nil
				}

				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("BulkInsert", mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
				)
				domainStudentMock.jsm.On("TracedPublish", mock.Anything, "publishDomainUserEvent", mock.Anything, mock.Anything).Return(nil, nil)
				domainStudentMock.tx.On("Commit", ctx).Return(nil)
				domainStudentMock.enrollmentStatusHistoryRepo.On("UpdateStudentStatusBasedEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantStudents: aggregate.DomainStudents{
				{
					DomainStudent:    domainStudentCreate,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
			},
		},
		{
			name: "happy case: update students with error collection",
			ctx:  ctx,
			studentWithIndexes: aggregate.DomainStudents{
				aggregate.DomainStudent{
					DomainStudent:    domainStudentUpdate,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.studentValidationManager.On("FullyValidate", ctx, domainStudentMock.db, mock.Anything, mock.Anything).Once().Return(
					aggregate.DomainStudents{},
					aggregate.DomainStudents{
						aggregate.DomainStudent{
							DomainStudent: domainStudentUpdate,
						},
					},
					[]error{},
				)
				domainStudentMock.usrEmailRepo.On("CreateMultiple", ctx, domainStudentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainStudentMock.db.On("Begin", ctx).Return(domainStudentMock.tx, nil)
				domainStudentMock.studentRepo.On("UpsertMultiple", ctx, domainStudentMock.tx, mock.Anything, mock.Anything).Return(nil)
				domainStudentMock.authUserUpserter = func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) (entity.LegacyUsers, error) {
					return entity.LegacyUsers{}, nil
				}
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("BulkInsert", mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("UpdateStudentStatusBasedEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainStudentMock.jsm.On("TracedPublish", mock.Anything, "publishDomainUserEvent", mock.Anything, mock.Anything).Return(nil, nil)
				domainStudentMock.tx.On("Commit", ctx).Return(nil)
			},
			wantStudents: aggregate.DomainStudents{
				{
					DomainStudent:    domainStudentUpdate,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
			},
			wantErrors: []error{},
		},
		{
			name: "happy case: both create and update students with error collection",
			ctx:  ctx,
			studentWithIndexes: aggregate.DomainStudents{
				aggregate.DomainStudent{
					DomainStudent:    domainStudentUpdate,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.studentValidationManager.On("FullyValidate", ctx, domainStudentMock.db, mock.Anything, mock.Anything).Once().Return(
					aggregate.DomainStudents{
						aggregate.DomainStudent{
							DomainStudent: domainStudentCreate,
						},
					},
					aggregate.DomainStudents{
						aggregate.DomainStudent{
							DomainStudent: domainStudentUpdate,
						},
					},
					[]error{},
				)
				domainStudentMock.usrEmailRepo.On("CreateMultiple", ctx, domainStudentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainStudentMock.db.On("Begin", ctx).Return(domainStudentMock.tx, nil)
				domainStudentMock.studentRepo.On("UpsertMultiple", ctx, domainStudentMock.tx, mock.Anything, mock.Anything).Return(nil)
				domainStudentMock.authUserUpserter = func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) (entity.LegacyUsers, error) {
					return entity.LegacyUsers{}, nil
				}

				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("BulkInsert", mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("UpdateStudentStatusBasedEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainStudentMock.jsm.On("TracedPublish", mock.Anything, "publishDomainUserEvent", mock.Anything, mock.Anything).Return(nil, nil)
				domainStudentMock.tx.On("Commit", ctx).Return(nil)
			},
			wantStudents: aggregate.DomainStudents{
				{
					DomainStudent:    domainStudentCreate,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
				{
					DomainStudent:    domainStudentUpdate,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
			},
			wantErrors: []error{},
		},
		{
			name: "unhappy case: throw list of errors when both create and update",
			ctx:  ctx,
			studentWithIndexes: aggregate.DomainStudents{
				aggregate.DomainStudent{
					DomainStudent:    domainStudentUpdate,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.studentValidationManager.On("FullyValidate", ctx, domainStudentMock.db, mock.Anything, mock.Anything).Once().Return(
					aggregate.DomainStudents{
						aggregate.DomainStudent{
							DomainStudent: domainStudentCreate,
						},
					},
					aggregate.DomainStudents{
						aggregate.DomainStudent{
							DomainStudent: domainStudentUpdate,
						},
					},
					[]error{
						errcode.Error{
							Code:      errcode.DuplicatedData,
							FieldName: string(entity.UserFieldEmail),
						},
						errcode.Error{
							Code:      errcode.DuplicatedData,
							FieldName: string(entity.StudentLocationsField),
						},
					},
				)
				domainStudentMock.usrEmailRepo.On("CreateMultiple", ctx, domainStudentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainStudentMock.db.On("Begin", ctx).Return(domainStudentMock.tx, nil)
				domainStudentMock.studentRepo.On("UpsertMultiple", ctx, domainStudentMock.tx, mock.Anything, mock.Anything).Return(nil)
				domainStudentMock.authUserUpserter = func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) (entity.LegacyUsers, error) {
					return entity.LegacyUsers{}, nil
				}

				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("BulkInsert", mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("UpdateStudentStatusBasedEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainStudentMock.jsm.On("TracedPublish", mock.Anything, "publishDomainUserEvent", mock.Anything, mock.Anything).Return(nil, nil)
				domainStudentMock.tx.On("Commit", ctx).Return(nil)
			},
			wantStudents: aggregate.DomainStudents{
				{
					DomainStudent:    domainStudentCreate,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
				{
					DomainStudent:    domainStudentUpdate,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
			},
			wantErrors: []error{
				errcode.Error{
					Code:      errcode.DuplicatedData,
					FieldName: string(entity.UserFieldEmail),
				},
				errcode.Error{
					Code:      errcode.DuplicatedData,
					FieldName: string(entity.StudentLocationsField),
				},
			},
		},
		{
			name: "unhappy case: throw error when create email user failed",
			ctx:  ctx,
			studentWithIndexes: aggregate.DomainStudents{
				aggregate.DomainStudent{
					DomainStudent:    domainStudentUpdate,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.studentValidationManager.On("FullyValidate", ctx, domainStudentMock.db, mock.Anything, mock.Anything).Once().Return(
					aggregate.DomainStudents{
						aggregate.DomainStudent{
							DomainStudent: domainStudentCreate,
						},
					},
					aggregate.DomainStudents{
						aggregate.DomainStudent{
							DomainStudent: domainStudentUpdate,
						},
					},
					[]error{},
				)
				domainStudentMock.usrEmailRepo.On("CreateMultiple", ctx, domainStudentMock.db, mock.Anything).Return(valueobj.HasUserIDs{}, errors.New("error"))
				domainStudentMock.db.On("Begin", ctx).Return(domainStudentMock.tx, nil)
				domainStudentMock.studentRepo.On("UpsertMultiple", ctx, domainStudentMock.tx, mock.Anything, mock.Anything).Return(nil)
				domainStudentMock.authUserUpserter = func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) (entity.LegacyUsers, error) {
					return entity.LegacyUsers{}, nil
				}

				domainStudentMock.jsm.On("TracedPublish", mock.Anything, "publishDomainUserEvent", mock.Anything, mock.Anything).Return(nil, nil)
				domainStudentMock.tx.On("Commit", ctx).Return(nil)
				domainStudentMock.enrollmentStatusHistoryRepo.On("UpdateStudentStatusBasedEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantStudents: aggregate.DomainStudents{
				{
					DomainStudent:    domainStudentCreate,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
				{
					DomainStudent:    domainStudentUpdate,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserAccessPaths: entity.DomainUserAccessPaths{
						entity.DefaultUserAccessPath{},
					},
				},
			},
			wantErrors: []error{
				entity.InternalError{
					RawErr: errors.Wrap(errors.New("error"), "service.UsrEmailRepo.CreateMultiple"),
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log(testCase.name)
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.ManabieSchool),
				},
			}
			testCase.ctx = interceptors.ContextWithJWTClaims(ctx, claim)
			testCase.ctx = interceptors.ContextWithUserID(testCase.ctx, "user-id")

			m, service := DomainStudentServiceMock()
			m.userPhoneNumberRepo.On("SoftDeleteByUserIDs", testCase.ctx, m.tx, mock.Anything).Return(nil)
			m.userAddressRepo.On("SoftDeleteByUserIDs", testCase.ctx, m.tx, mock.Anything).Return(nil)
			m.taggedUserRepo.On("SoftDeleteByUserIDs", testCase.ctx, m.tx, mock.Anything).Return(nil)
			m.schoolHistoryRepo.On("SoftDeleteByStudentIDs", testCase.ctx, m.tx, mock.Anything).Return(nil)
			testCase.setupWithMock(testCase.ctx, &m)
			service.AuthUserUpserter = m.authUserUpserter
			option := unleash.DomainStudentFeatureOption{
				DomainUserFeatureOption: unleash.DomainUserFeatureOption{
					EnableIgnoreUpdateEmail: true,
					EnableUsername:          true,
				},
				EnableAutoDeactivateAndReactivateStudentV2:            true,
				DisableAutoDeactivateAndReactivateStudent:             false,
				EnableExperimentalBulkInsertEnrollmentStatusHistories: true,
			}
			students, listOfErrors := service.UpsertMultipleWithErrorCollection(testCase.ctx, testCase.studentWithIndexes, option)
			t.Log(listOfErrors)
			if len(testCase.wantStudents) > 0 {
				for i, student := range students {
					assert.Equal(t, testCase.wantStudents[i].DomainStudent, student.DomainStudent)
				}
			}
			if len(testCase.wantErrors) > 0 {
				for i, err := range listOfErrors {
					assert.Equal(t, testCase.wantErrors[i].Error(), err.Error())
				}
			}
		})
	}
}

func TestDomainStudent_UpsertMultipleWithSchoolHistories(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	domainStudent := &mock_usermgmt.Student{
		RandomStudent: mock_usermgmt.RandomStudent{
			GradeID:          field.NewString("grade-id-1"),
			Email:            field.NewString("test@manabie.com"),
			UserName:         field.NewString("username1"),
			Gender:           field.NewString(upb.Gender_FEMALE.String()),
			FirstName:        field.NewString("test first name"),
			LastName:         field.NewString("test last name"),
			CurrentGrade:     field.NewInt16(1),
			EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
		},
	}

	testCases := []TestCase{
		{
			name: "happy case: upsert students with school histories",
			ctx:  ctx,
			req: aggregate.DomainStudent{
				DomainStudent:    domainStudent,
				LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
				UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
				UserAccessPaths: entity.DomainUserAccessPaths{
					entity.DefaultUserAccessPath{},
				},
				SchoolHistories: entity.DomainSchoolHistories{
					&repository.SchoolHistory{
						SchoolHistoryAttribute: repository.SchoolHistoryAttribute{
							SchoolID: field.NewString("school-id-1"),
						},
					},
				},
				EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "location-id",
						upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
						time.Now().Add(-100*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1),
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByExternalUserIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainStudentMock.db, constant.RoleStudent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainStudentMock.locationRepo.On("RetrieveLowestLevelLocations", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainLocations{mock_usermgmt.NewLocation("", "")}, nil)
				domainStudentMock.studentPackage.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainStudentPackages{}, nil)
				domainStudentMock.db.On("Begin", ctx).Return(domainStudentMock.tx, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
				domainStudentMock.locationRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainLocations{&repository.Location{LocationAttribute: repository.LocationAttribute{ID: field.NewString("location-id")}}}, nil)
				domainStudentMock.schoolRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainSchools{
					&repository.School{
						SchoolAttribute: repository.SchoolAttribute{
							ID: field.NewString("school-id-1"),
						},
					},
				}, nil)
				domainStudentMock.schoolCourseRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainSchoolCourses{}, nil)
				domainStudentMock.studentRepo.On("UpsertMultiple", ctx, domainStudentMock.tx, mock.Anything, mock.Anything).Return(nil)
				domainStudentMock.schoolRepo.On("GetByIDsAndGradeID", ctx, domainStudentMock.tx, mock.Anything, mock.Anything).Return(entity.DomainSchools{}, nil)
				domainStudentMock.schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, domainStudentMock.tx, mock.Anything).Return(nil)
				domainStudentMock.schoolHistoryRepo.On("SetCurrentSchoolByStudentIDsAndSchoolIDs", ctx, domainStudentMock.tx, mock.Anything, mock.Anything).Return(nil)
				domainStudentMock.schoolHistoryRepo.On("UpsertMultiple", ctx, domainStudentMock.tx, mock.Anything).Return(nil)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
				domainStudentMock.tx.On("Commit", ctx).Return(nil)
				// upsertEnrollmentStatusHistories
				domainStudentMock.unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureUsingMasterReplicatedTable, mock.Anything, mock.Anything).Once().Return(false, nil)
				domainStudentMock.configurationClient.On("GetConfigurationByKey", ctx, &mpb.GetConfigurationByKeyRequest{Key: constant.KeyEnrollmentStatusHistoryConfig}).Once().Return(&mpb.GetConfigurationByKeyResponse{
					Configuration: &mpb.Configuration{
						ConfigValue: constant.ConfigValueOn,
						ConfigKey:   constant.KeyEnrollmentStatusHistoryConfig,
					},
				}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]entity.DomainEnrollmentStatusHistory{}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
				)
				hashConfig := mockScryptHash()
				domainStudentMock.tenantManager.On("TenantClient", ctx, mock.Anything).Return(domainStudentMock.tenantClient, nil)
				domainStudentMock.tenantClient.On("GetHashConfig").Return(hashConfig)
				domainStudentMock.tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Return(&internal_auth_user.ImportUsersResult{}, nil)
				domainStudentMock.authUserUpserter = func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) (entity.LegacyUsers, error) {
					return entity.LegacyUsers{}, nil
				}

				domainStudentMock.enrollmentStatusHistoryRepo.On("UpdateStudentStatusBasedEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "bad case: upsert students with school histories: school not found",
			ctx:  ctx,
			expectedErr: entity.NotFoundErrorWithArrayNestedField{
				NotFoundError: entity.NotFoundError{
					EntityName: entity.StudentEntity,
					FieldName:  entity.StudentSchoolField,
					FieldValue: "school-id-1",
					Index:      0,
				},
				NestedFieldName: entity.StudentSchoolHistoryField,
				NestedIndex:     0,
			},
			req: aggregate.DomainStudent{
				DomainStudent:    domainStudent,
				LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
				UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
				UserAccessPaths: entity.DomainUserAccessPaths{
					entity.DefaultUserAccessPath{},
				},
				SchoolHistories: entity.DomainSchoolHistories{
					&repository.SchoolHistory{
						SchoolHistoryAttribute: repository.SchoolHistoryAttribute{
							SchoolID: field.NewString("school-id-1"),
						},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByExternalUserIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainStudentMock.db, constant.RoleStudent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
				domainStudentMock.locationRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainStudentMock.schoolRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainSchools{}, nil)
			},
		},
		{
			name: "bad case: upsert students with school histories: school_course not found",
			ctx:  ctx,
			expectedErr: entity.NotFoundErrorWithArrayNestedField{
				NotFoundError: entity.NotFoundError{
					EntityName: entity.StudentEntity,
					FieldName:  entity.StudentSchoolCourseField,
					FieldValue: "school-course-id-1",
					Index:      0,
				},
				NestedFieldName: entity.StudentSchoolHistoryField,
				NestedIndex:     0,
			},
			req: aggregate.DomainStudent{
				DomainStudent:    domainStudent,
				LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
				UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
				UserAccessPaths: entity.DomainUserAccessPaths{
					entity.DefaultUserAccessPath{},
				},
				SchoolHistories: entity.DomainSchoolHistories{
					&repository.SchoolHistory{
						SchoolHistoryAttribute: repository.SchoolHistoryAttribute{
							SchoolID:       field.NewString("school-id-1"),
							SchoolCourseID: field.NewString("school-course-id-1"),
						},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByExternalUserIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainStudentMock.db, constant.RoleStudent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
				domainStudentMock.locationRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainStudentMock.schoolRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainSchools{
					&repository.School{
						SchoolAttribute: repository.SchoolAttribute{
							ID: field.NewString("school-id-1"),
						},
					},
				}, nil)
				domainStudentMock.schoolCourseRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainSchoolCourses{}, nil)
			},
		},
		{
			name: "bad case: upsert students with school histories: duplicate school_partner_id in system",
			ctx:  ctx,
			expectedErr: entity.NotFoundErrorWithArrayNestedField{
				NotFoundError: entity.NotFoundError{
					EntityName: entity.StudentEntity,
					FieldName:  entity.StudentSchoolField,
					FieldValue: "school-id-1",
					Index:      0,
				},
				NestedFieldName: entity.StudentSchoolHistoryField,
				NestedIndex:     0,
			},
			req: aggregate.DomainStudent{
				DomainStudent:    domainStudent,
				LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
				UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
				UserAccessPaths: entity.DomainUserAccessPaths{
					entity.DefaultUserAccessPath{},
				},
				SchoolHistories: entity.DomainSchoolHistories{
					&repository.SchoolHistory{
						SchoolHistoryAttribute: repository.SchoolHistoryAttribute{
							SchoolID:       field.NewString("school-id-1"),
							SchoolCourseID: field.NewString("school-course-id-1"),
						},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByExternalUserIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainStudentMock.db, constant.RoleStudent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
				domainStudentMock.locationRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainStudentMock.schoolRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainSchools{
					entity.DefaultDomainSchool{},
					entity.DefaultDomainSchool{},
				}, nil)
				domainStudentMock.schoolCourseRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainSchoolCourses{}, nil)
			},
		},
		{
			name: "bad case: upsert students with school histories: duplicate school_level_id in system",
			ctx:  ctx,
			expectedErr: entity.DuplicatedFieldErrorWithArrayNestedField{
				DuplicatedFieldError: entity.DuplicatedFieldError{
					EntityName:      entity.StudentEntity,
					DuplicatedField: entity.StudentSchoolHistorySchoolLevel,
					Index:           0,
				},
				NestedFieldName: entity.StudentSchoolHistoryField,
				NestedIndex:     1,
			},
			req: aggregate.DomainStudent{
				DomainStudent:    domainStudent,
				LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
				UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
				UserAccessPaths: entity.DomainUserAccessPaths{
					entity.DefaultUserAccessPath{},
				},
				SchoolHistories: entity.DomainSchoolHistories{
					&repository.SchoolHistory{
						SchoolHistoryAttribute: repository.SchoolHistoryAttribute{
							SchoolID:       field.NewString("school-id-1"),
							SchoolCourseID: field.NewString("school-course-id-1"),
						},
					},
					&repository.SchoolHistory{
						SchoolHistoryAttribute: repository.SchoolHistoryAttribute{
							SchoolID:       field.NewString("school-id-2"),
							SchoolCourseID: field.NewString("school-course-id-2"),
						},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByExternalUserIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainStudentMock.db, constant.RoleStudent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
				domainStudentMock.locationRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainStudentMock.schoolRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainSchools{
					&repository.School{
						SchoolAttribute: repository.SchoolAttribute{
							ID:            field.NewString("school-id-1"),
							SchoolLevelID: field.NewString("level-1"),
						},
					},
					&repository.School{
						SchoolAttribute: repository.SchoolAttribute{
							ID:            field.NewString("school-id-2"),
							SchoolLevelID: field.NewString("level-1"),
						},
					},
				}, nil)
				domainStudentMock.schoolCourseRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainSchoolCourses{}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log(testCase.name)
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.ManabieSchool),
				},
			}
			testCase.ctx = interceptors.ContextWithJWTClaims(ctx, claim)
			testCase.ctx = interceptors.ContextWithUserID(testCase.ctx, "user-id")

			m, service := DomainStudentServiceMock()
			m.userRepo.On("GetByEmailsInsensitiveCase", testCase.ctx, m.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
			m.userRepo.On("GetByExternalUserIDs", testCase.ctx, m.db, []string{""}).Return(entity.Users{}, nil)
			m.usrEmailRepo.On("CreateMultiple", testCase.ctx, m.db, mock.Anything).Return(valueobj.HasUserIDs{
				repository.UsrEmail{
					UsrID: field.NewString("new-id"),
				},
			}, nil)

			m.userPhoneNumberRepo.On("SoftDeleteByUserIDs", testCase.ctx, m.tx, mock.Anything).Return(nil)
			m.userAddressRepo.On("SoftDeleteByUserIDs", testCase.ctx, m.tx, mock.Anything).Return(nil)
			m.taggedUserRepo.On("SoftDeleteByUserIDs", testCase.ctx, m.tx, mock.Anything).Return(nil)
			m.usrEmailRepo.On("CreateMultiple", testCase.ctx, m.db, mock.Anything).Return(valueobj.HasUserIDs{}, nil)
			m.OrganizationRepo.On("GetTenantIDByOrgID", testCase.ctx, m.tx, mock.Anything).Return("", nil)
			/*hashConfig := mockScryptHash()
			m.tenantManager.On("TenantClient", testCase.ctx, mock.Anything).Return(m.tenantClient, nil)
			m.tenantClient.On("GetHashConfig").Return(hashConfig)
			m.tenantClient.On("ImportUsers", testCase.ctx, mock.Anything, mock.Anything).Return(&internal_auth_user.ImportUsersResult{}, nil)*/
			m.jsm.On("TracedPublish", mock.Anything, "publishDomainUserEvent", mock.Anything, mock.Anything).Return(nil, nil)

			testCase.setupWithMock(testCase.ctx, &m)
			service.AuthUserUpserter = m.authUserUpserter
			option := unleash.DomainStudentFeatureOption{
				DomainUserFeatureOption: unleash.DomainUserFeatureOption{
					EnableIgnoreUpdateEmail: false,
					EnableUsername:          true,
				},
				EnableAutoDeactivateAndReactivateStudentV2:            true,
				DisableAutoDeactivateAndReactivateStudent:             false,
				EnableExperimentalBulkInsertEnrollmentStatusHistories: false,
			}
			_, err := service.UpsertMultiple(testCase.ctx, option, testCase.req.(aggregate.DomainStudent))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestDomainStudent_UpsertMultipleWithUserAddress(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	domainStudent := &mock_usermgmt.Student{
		RandomStudent: mock_usermgmt.RandomStudent{
			GradeID:          field.NewString("grade-id-1"),
			Email:            field.NewString("test@manabie.com"),
			UserName:         field.NewString("username1"),
			Gender:           field.NewString(upb.Gender_FEMALE.String()),
			FirstName:        field.NewString("test first name"),
			LastName:         field.NewString("test last name"),
			ExternalUserID:   field.NewString("external-user-id"),
			CurrentGrade:     field.NewInt16(1),
			EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
		},
	}

	testCases := []TestCase{
		{
			name: "happy case: upsert students with user address mandatory only",
			ctx:  ctx,
			req: aggregate.DomainStudent{
				DomainStudent:    domainStudent,
				LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
				UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
				UserAccessPaths: entity.DomainUserAccessPaths{
					entity.DefaultUserAccessPath{},
				},
				UserAddress: entity.DefaultDomainUserAddress{},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByExternalUserIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainStudentMock.db, constant.RoleStudent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainStudentMock.locationRepo.On("RetrieveLowestLevelLocations", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainLocations{mock_usermgmt.NewLocation("", "")}, nil)
				domainStudentMock.studentPackage.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainStudentPackages{}, nil)
				domainStudentMock.db.On("Begin", ctx).Return(domainStudentMock.tx, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
				domainStudentMock.locationRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainStudentMock.prefectureRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainPrefecture{entity.DefaultDomainPrefecture{}}, nil)
				domainStudentMock.studentRepo.On("UpsertMultiple", ctx, domainStudentMock.tx, mock.Anything, mock.Anything).Return(nil)
				domainStudentMock.userAddressRepo.On("SoftDeleteByUserIDs", ctx, domainStudentMock.tx, mock.Anything).Return(nil)
				domainStudentMock.userAddressRepo.On("UpsertMultiple", ctx, domainStudentMock.tx, mock.Anything).Return(nil)
				domainStudentMock.tx.On("Commit", ctx).Return(nil)
				// upsertEnrollmentStatusHistories
				domainStudentMock.unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureUsingMasterReplicatedTable, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainStudentMock.internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(entity.NullDomainConfiguration{}, nil)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
				)
				hashConfig := mockScryptHash()
				domainStudentMock.tenantManager.On("TenantClient", ctx, mock.Anything).Return(domainStudentMock.tenantClient, nil)
				domainStudentMock.tenantClient.On("GetHashConfig").Return(hashConfig)
				domainStudentMock.tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Return(&internal_auth_user.ImportUsersResult{}, nil)
				domainStudentMock.authUserUpserter = func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) (entity.LegacyUsers, error) {
					return entity.LegacyUsers{}, nil
				}

				domainStudentMock.enrollmentStatusHistoryRepo.On("UpdateStudentStatusBasedEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
			},
		},
		{
			name: "happy case: upsert students with user address",
			ctx:  ctx,
			req: aggregate.DomainStudent{
				DomainStudent:    domainStudent,
				LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
				UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
				UserAccessPaths: entity.DomainUserAccessPaths{
					entity.DefaultUserAccessPath{},
				},
				UserAddress: entity.DefaultDomainUserAddress{},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByExternalUserIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainStudentMock.db, constant.RoleStudent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainStudentMock.locationRepo.On("RetrieveLowestLevelLocations", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainLocations{mock_usermgmt.NewLocation("", "")}, nil)
				domainStudentMock.studentPackage.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainStudentPackages{}, nil)
				domainStudentMock.db.On("Begin", ctx).Return(domainStudentMock.tx, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
				domainStudentMock.locationRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainStudentMock.prefectureRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainPrefecture{entity.DefaultDomainPrefecture{}}, nil)
				domainStudentMock.studentRepo.On("UpsertMultiple", ctx, domainStudentMock.tx, mock.Anything, mock.Anything).Return(nil)
				domainStudentMock.userAddressRepo.On("SoftDeleteByUserIDs", ctx, domainStudentMock.tx, mock.Anything).Return(nil)
				domainStudentMock.userAddressRepo.On("UpsertMultiple", ctx, domainStudentMock.tx, mock.Anything).Return(nil)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
				domainStudentMock.tx.On("Commit", ctx).Return(nil)
				// upsertEnrollmentStatusHistories
				domainStudentMock.unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureUsingMasterReplicatedTable, mock.Anything, mock.Anything).Once().Return(false, nil)
				domainStudentMock.configurationClient.On("GetConfigurationByKey", ctx, &mpb.GetConfigurationByKeyRequest{Key: constant.KeyEnrollmentStatusHistoryConfig}).Once().Return(&mpb.GetConfigurationByKeyResponse{
					Configuration: &mpb.Configuration{
						ConfigValue: constant.ConfigValueOn,
						ConfigKey:   constant.KeyEnrollmentStatusHistoryConfig,
					},
				}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
				)
				hashConfig := mockScryptHash()
				domainStudentMock.tenantManager.On("TenantClient", ctx, mock.Anything).Return(domainStudentMock.tenantClient, nil)
				domainStudentMock.tenantClient.On("GetHashConfig").Return(hashConfig)
				domainStudentMock.tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Return(&internal_auth_user.ImportUsersResult{}, nil)
				domainStudentMock.authUserUpserter = func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) (entity.LegacyUsers, error) {
					return entity.LegacyUsers{}, nil
				}

				domainStudentMock.enrollmentStatusHistoryRepo.On("UpdateStudentStatusBasedEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "bad case: invalid prefecture id",
			ctx:  ctx,
			expectedErr: entity.InvalidFieldError{
				EntityName: entity.StudentEntity,
				FieldName:  entity.StudentUserAddressPrefectureField,
				Index:      0,
			},
			req: aggregate.DomainStudent{
				DomainStudent:    domainStudent,
				LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
				UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
				UserAccessPaths: entity.DomainUserAccessPaths{
					entity.DefaultUserAccessPath{},
				},
				UserAddress: &entity.UserAddressWillBeDelegated{
					UserAddressAttribute: entity.DefaultDomainUserAddress{},
					HasPrefectureID: &repository.Prefecture{
						PrefectureAttribute: repository.PrefectureAttribute{
							ID: field.NewString("id-1"),
						},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByExternalUserIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainStudentMock.db, constant.RoleStudent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
				domainStudentMock.locationRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainStudentMock.prefectureRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainPrefecture{}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log(testCase.name)
			m, service := DomainStudentServiceMock()
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.ManabieSchool),
				},
			}
			testCase.ctx = interceptors.ContextWithJWTClaims(ctx, claim)
			testCase.ctx = interceptors.ContextWithUserID(testCase.ctx, "user-id")
			m.userRepo.On("GetByEmailsInsensitiveCase", testCase.ctx, m.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
			m.userRepo.On("GetByExternalUserIDs", testCase.ctx, m.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
			m.usrEmailRepo.On("CreateMultiple", testCase.ctx, m.db, mock.Anything).Return(valueobj.HasUserIDs{
				repository.UsrEmail{
					UsrID: field.NewString("new-id"),
				},
			}, nil)
			m.userPhoneNumberRepo.On("SoftDeleteByUserIDs", testCase.ctx, m.tx, mock.Anything).Return(nil)
			m.taggedUserRepo.On("SoftDeleteByUserIDs", testCase.ctx, m.tx, mock.Anything).Return(nil)
			m.schoolHistoryRepo.On("SoftDeleteByStudentIDs", testCase.ctx, m.tx, mock.Anything).Return(nil)
			m.usrEmailRepo.On("CreateMultiple", testCase.ctx, m.db, mock.Anything).Return(valueobj.HasUserIDs{}, nil)
			m.OrganizationRepo.On("GetTenantIDByOrgID", testCase.ctx, m.tx, mock.Anything).Return("", nil)
			/*hashConfig := mockScryptHash()
			m.tenantManager.On("TenantClient", testCase.ctx, mock.Anything).Return(m.tenantClient, nil)
			m.tenantClient.On("GetHashConfig").Return(hashConfig)
			m.tenantClient.On("ImportUsers", testCase.ctx, mock.Anything, mock.Anything).Return(&internal_auth_user.ImportUsersResult{}, nil)*/
			m.jsm.On("TracedPublish", mock.Anything, "publishDomainUserEvent", mock.Anything, mock.Anything).Return(nil, nil)

			testCase.setupWithMock(testCase.ctx, &m)
			service.AuthUserUpserter = m.authUserUpserter
			option := unleash.DomainStudentFeatureOption{
				DomainUserFeatureOption: unleash.DomainUserFeatureOption{
					EnableIgnoreUpdateEmail: false,
					EnableUsername:          true,
				},
				EnableAutoDeactivateAndReactivateStudentV2:            true,
				DisableAutoDeactivateAndReactivateStudent:             false,
				EnableExperimentalBulkInsertEnrollmentStatusHistories: false,
			}
			_, err := service.UpsertMultiple(testCase.ctx, option, testCase.req.(aggregate.DomainStudent))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestDomainStudent_UpsertMultipleWithUserPhoneNumber(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	domainStudent := &mock_usermgmt.Student{
		RandomStudent: mock_usermgmt.RandomStudent{
			Email:            field.NewString("test@manabie.com"),
			UserName:         field.NewString("username1"),
			GradeID:          field.NewString("grade-id-1"),
			Gender:           field.NewString(upb.Gender_FEMALE.String()),
			FirstName:        field.NewString("test first name"),
			LastName:         field.NewString("test last name"),
			CurrentGrade:     field.NewInt16(1),
			EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
		},
	}

	testCases := []TestCase{
		{
			name: "happy case: upsert students with user phone number",
			ctx:  ctx,
			req: aggregate.DomainStudent{
				DomainStudent:    domainStudent,
				LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
				UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
				UserAccessPaths: entity.DomainUserAccessPaths{
					entity.DefaultUserAccessPath{},
				},
				UserPhoneNumbers: entity.DomainUserPhoneNumbers{
					&repository.UserPhoneNumber{
						UserPhoneNumberAttribute: repository.UserPhoneNumberAttribute{
							PhoneNumber: field.NewString("09876544321"),
						},
					},
					&repository.UserPhoneNumber{
						UserPhoneNumberAttribute: repository.UserPhoneNumberAttribute{
							PhoneNumber: field.NewString("09876544322"),
						},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainStudentMock.db, constant.RoleStudent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainStudentMock.locationRepo.On("RetrieveLowestLevelLocations", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainLocations{mock_usermgmt.NewLocation("", "")}, nil)
				domainStudentMock.studentPackage.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainStudentPackages{}, nil)
				domainStudentMock.db.On("Begin", ctx).Return(domainStudentMock.tx, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
				domainStudentMock.locationRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainStudentMock.studentRepo.On("UpsertMultiple", ctx, domainStudentMock.tx, mock.Anything, mock.Anything).Return(nil)
				domainStudentMock.userPhoneNumberRepo.On("SoftDeleteByUserIDs", ctx, domainStudentMock.tx, mock.Anything).Return(nil)
				domainStudentMock.userPhoneNumberRepo.On("UpsertMultiple", ctx, domainStudentMock.tx, mock.Anything).Return(nil)
				domainStudentMock.tx.On("Commit", ctx).Return(nil)
				// upsertEnrollmentStatusHistories
				domainStudentMock.unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureUsingMasterReplicatedTable, mock.Anything, mock.Anything).Once().Return(false, nil)
				domainStudentMock.configurationClient.On("GetConfigurationByKey", ctx, &mpb.GetConfigurationByKeyRequest{Key: constant.KeyEnrollmentStatusHistoryConfig}).Once().Return(&mpb.GetConfigurationByKeyResponse{
					Configuration: &mpb.Configuration{
						ConfigValue: constant.ConfigValueOn,
						ConfigKey:   constant.KeyEnrollmentStatusHistoryConfig,
					},
				}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
				)
				hashConfig := mockScryptHash()
				domainStudentMock.tenantManager.On("TenantClient", ctx, mock.Anything).Return(domainStudentMock.tenantClient, nil)
				domainStudentMock.tenantClient.On("GetHashConfig").Return(hashConfig)
				domainStudentMock.tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Return(&internal_auth_user.ImportUsersResult{}, nil)
				domainStudentMock.authUserUpserter = func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) (entity.LegacyUsers, error) {
					return entity.LegacyUsers{}, nil
				}

				domainStudentMock.enrollmentStatusHistoryRepo.On("UpdateStudentStatusBasedEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
			},
		},
		{
			name: "bad case: upsert students with invalid phone number",
			ctx:  ctx,
			expectedErr: entity.InvalidFieldError{
				EntityName: entity.UserEntity,
				FieldName:  entity.StudentFieldStudentPhoneNumber,
				Index:      0,
			},
			req: aggregate.DomainStudent{
				DomainStudent:    domainStudent,
				LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
				UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
				UserAccessPaths: entity.DomainUserAccessPaths{
					entity.DefaultUserAccessPath{},
				},
				UserPhoneNumbers: entity.DomainUserPhoneNumbers{
					&repository.UserPhoneNumber{
						UserPhoneNumberAttribute: repository.UserPhoneNumberAttribute{
							PhoneNumber: field.NewString("asd"),
							Type:        field.NewString(constant.StudentPhoneNumber),
						},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainStudentMock.db, constant.RoleStudent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainStudentMock.locationRepo.On("RetrieveLowestLevelLocations", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainLocations{mock_usermgmt.NewLocation("", "")}, nil)
				domainStudentMock.studentPackage.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainStudentPackages{}, nil)
				domainStudentMock.db.On("Begin", ctx).Return(domainStudentMock.tx, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
				domainStudentMock.locationRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainStudentMock.studentRepo.On("UpsertMultiple", ctx, domainStudentMock.tx, mock.Anything, mock.Anything).Return(nil)
				domainStudentMock.userPhoneNumberRepo.On("SoftDeleteByUserIDs", ctx, domainStudentMock.tx, mock.Anything).Return(nil)
				domainStudentMock.userPhoneNumberRepo.On("UpsertMultiple", ctx, domainStudentMock.tx, mock.Anything).Return(nil)
				domainStudentMock.tx.On("Commit", ctx).Return(nil)
			},
		},
		{
			name: "bad case: upsert students with duplicated phone number",
			ctx:  ctx,
			expectedErr: entity.DuplicatedFieldError{
				EntityName:      entity.UserEntity,
				DuplicatedField: entity.StudentFieldHomePhoneNumber,
				Index:           0,
			},
			req: aggregate.DomainStudent{
				DomainStudent:    domainStudent,
				LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
				UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
				UserAccessPaths: entity.DomainUserAccessPaths{
					entity.DefaultUserAccessPath{},
				},
				UserPhoneNumbers: entity.DomainUserPhoneNumbers{
					&repository.UserPhoneNumber{
						UserPhoneNumberAttribute: repository.UserPhoneNumberAttribute{
							PhoneNumber: field.NewString("0987654321"),
							Type:        field.NewString(constant.StudentPhoneNumber),
						},
					},
					&repository.UserPhoneNumber{
						UserPhoneNumberAttribute: repository.UserPhoneNumberAttribute{
							PhoneNumber: field.NewString("0987654321"),
							Type:        field.NewString(constant.StudentPhoneNumber),
						},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainStudentMock.db, constant.RoleStudent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainStudentMock.locationRepo.On("RetrieveLowestLevelLocations", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainLocations{mock_usermgmt.NewLocation("", "")}, nil)
				domainStudentMock.studentPackage.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainStudentPackages{}, nil)
				domainStudentMock.db.On("Begin", ctx).Return(domainStudentMock.tx, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
				domainStudentMock.locationRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainStudentMock.studentRepo.On("UpsertMultiple", ctx, domainStudentMock.tx, mock.Anything, mock.Anything).Return(nil)
				domainStudentMock.userPhoneNumberRepo.On("SoftDeleteByUserIDs", ctx, domainStudentMock.tx, mock.Anything).Return(nil)
				domainStudentMock.userPhoneNumberRepo.On("UpsertMultiple", ctx, domainStudentMock.tx, mock.Anything).Return(nil)
				domainStudentMock.tx.On("Commit", ctx).Return(nil)
				// upsertEnrollmentStatusHistories
				domainStudentMock.unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureUsingMasterReplicatedTable, mock.Anything, mock.Anything).Once().Return(false, nil)
				domainStudentMock.configurationClient.On("GetConfigurationByKey", ctx, &mpb.GetConfigurationByKeyRequest{Key: constant.KeyEnrollmentStatusHistoryConfig}).Once().Return(&mpb.GetConfigurationByKeyResponse{
					Configuration: &mpb.Configuration{
						ConfigValue: constant.ConfigValueOn,
						ConfigKey:   constant.KeyEnrollmentStatusHistoryConfig,
					},
				}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
				)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log(testCase.name)
			m, service := DomainStudentServiceMock()
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.ManabieSchool),
				},
			}
			testCase.ctx = interceptors.ContextWithJWTClaims(ctx, claim)
			testCase.ctx = interceptors.ContextWithUserID(testCase.ctx, "user-id")
			m.userRepo.On("GetByEmailsInsensitiveCase", testCase.ctx, m.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
			m.userRepo.On("GetByExternalUserIDs", testCase.ctx, m.db, []string{""}).Return(entity.Users{}, nil)
			m.studentRepo.On("GetUsersByExternalUserIDs", testCase.ctx, m.db, []string{""}).Return(entity.Users{}, nil)
			m.usrEmailRepo.On("CreateMultiple", testCase.ctx, m.db, mock.Anything).Return(valueobj.HasUserIDs{
				repository.UsrEmail{
					UsrID: field.NewString("new-id"),
				},
			}, nil)
			m.userAddressRepo.On("SoftDeleteByUserIDs", testCase.ctx, m.tx, mock.Anything).Return(nil)
			m.taggedUserRepo.On("SoftDeleteByUserIDs", testCase.ctx, m.tx, mock.Anything).Return(nil)
			m.schoolHistoryRepo.On("SoftDeleteByStudentIDs", testCase.ctx, m.tx, mock.Anything).Return(nil)
			m.usrEmailRepo.On("CreateMultiple", testCase.ctx, m.db, mock.Anything).Return(valueobj.HasUserIDs{}, nil)
			m.OrganizationRepo.On("GetTenantIDByOrgID", testCase.ctx, m.tx, mock.Anything).Return("", nil)
			/*hashConfig := mockScryptHash()
			m.tenantManager.On("TenantClient", testCase.ctx, mock.Anything).Return(m.tenantClient, nil)
			m.tenantClient.On("GetHashConfig").Return(hashConfig)
			m.tenantClient.On("ImportUsers", testCase.ctx, mock.Anything, mock.Anything).Return(&internal_auth_user.ImportUsersResult{}, nil)*/
			m.jsm.On("TracedPublish", mock.Anything, "publishDomainUserEvent", mock.Anything, mock.Anything).Return(nil, nil)

			testCase.setupWithMock(testCase.ctx, &m)
			service.AuthUserUpserter = m.authUserUpserter
			option := unleash.DomainStudentFeatureOption{
				DomainUserFeatureOption: unleash.DomainUserFeatureOption{
					EnableIgnoreUpdateEmail: false,
					EnableUsername:          true,
				},
				EnableAutoDeactivateAndReactivateStudentV2:            true,
				DisableAutoDeactivateAndReactivateStudent:             false,
				EnableExperimentalBulkInsertEnrollmentStatusHistories: false,
			}
			_, err := service.UpsertMultiple(testCase.ctx, option, testCase.req.(aggregate.DomainStudent))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestDomainStudent_MapFunctions(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	externalIDs := []string{"external-id"}

	t.Run("external user id to user id", func(t *testing.T) {
		m, service := DomainStudentServiceMock()
		m.userRepo.On("GetByExternalUserIDs", ctx, m.db, externalIDs).Return(entity.Users{}, nil)
		_, err := service.GetUsersByExternalIDs(ctx, externalIDs)
		if err != nil {
			fmt.Println(err)
		}
		assert.Equal(t, nil, err)
	})
	t.Run("school external id to school id", func(t *testing.T) {
		m, service := DomainStudentServiceMock()
		m.schoolRepo.On("GetByPartnerInternalIDs", ctx, m.db, externalIDs).Return(entity.DomainSchools{}, nil)
		_, err := service.GetSchoolsByExternalIDs(ctx, externalIDs)
		if err != nil {
			fmt.Println(err)
		}
		assert.Equal(t, nil, err)
	})
	t.Run("school course external id to school course id", func(t *testing.T) {
		m, service := DomainStudentServiceMock()
		m.schoolCourseRepo.On("GetByPartnerInternalIDs", ctx, m.db, externalIDs).Return(entity.DomainSchoolCourses{}, nil)
		_, err := service.GetSchoolCoursesByExternalIDs(ctx, externalIDs)
		if err != nil {
			fmt.Println(err)
		}
		assert.Equal(t, nil, err)
	})
	t.Run("grade external id to grade id", func(t *testing.T) {
		m, service := DomainStudentServiceMock()
		m.gradeRepo.On("GetByPartnerInternalIDs", ctx, m.db, externalIDs).Return([]entity.DomainGrade{}, nil)
		_, err := service.GetGradesByExternalIDs(ctx, externalIDs)
		if err != nil {
			fmt.Println(err)
		}
		assert.Equal(t, nil, err)
	})
	t.Run("tag external id to tag id", func(t *testing.T) {
		m, service := DomainStudentServiceMock()
		m.tagRepo.On("GetByPartnerInternalIDs", ctx, m.db, externalIDs).Return(entity.DomainTags{}, nil)
		_, err := service.GetTagsByExternalIDs(ctx, externalIDs)
		if err != nil {
			fmt.Println(err)
		}
		assert.Equal(t, nil, err)
	})
	t.Run("location external id to location id", func(t *testing.T) {
		m, service := DomainStudentServiceMock()
		m.locationRepo.On("GetByPartnerInternalIDs", ctx, m.db, externalIDs).Return(entity.DomainLocations{}, nil)
		_, err := service.GetLocationsByExternalIDs(ctx, externalIDs)
		if err != nil {
			fmt.Println(err)
		}
		assert.Equal(t, nil, err)
	})
	t.Run("prefecture code to prefecture id", func(t *testing.T) {
		m, service := DomainStudentServiceMock()
		m.prefectureRepo.On("GetByPrefectureCodes", ctx, m.db, externalIDs).Return(entity.DomainPrefectures{}, nil)
		_, err := service.GetPrefecturesByCodes(ctx, externalIDs)
		if err != nil {
			fmt.Println(err)
		}
		assert.Equal(t, nil, err)
	})
}

func TestDomainStudent_UpsertMultipleWithAssignedParent(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	randomStudent := mock_usermgmt.RandomStudent{
		GradeID:          field.NewString("grade-id-1"),
		Email:            field.NewString("test@manabie.com"),
		UserName:         field.NewString("username1"),
		Gender:           field.NewString(upb.Gender_FEMALE.String()),
		FirstName:        field.NewString("test first name"),
		LastName:         field.NewString("test last name"),
		ExternalUserID:   field.NewString("external-user-id"),
		CurrentGrade:     field.NewInt16(1),
		EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
	}

	domainStudent := &mock_usermgmt.StudentWithAssignedParent{
		Student: mock_usermgmt.Student{RandomStudent: randomStudent},
	}

	domainStudentWithAssignedParent := &mock_usermgmt.StudentWithAssignedParent{
		Student: mock_usermgmt.Student{RandomStudent: randomStudent},
		Parents: []mock_usermgmt.Parent{
			{
				RandomParent: mock_usermgmt.RandomParent{
					EmailAttr:     field.NewString("parent+test@manabie.com"),
					FirstNameAttr: field.NewString("parent first name"),
					LastNameAttr:  field.NewString("parent last name"),
				},
			},
		},
	}
	studentParent := &mock_usermgmt.StudentParentRelationship{
		RandomStudentParentRelationship: mock_usermgmt.RandomStudentParentRelationship{
			StudentIDAttr:    field.NewString("student_id"),
			ParentIDAttr:     field.NewString("parent_id"),
			RelationshipAttr: field.NewString(string(constant.FamilyRelationshipFather)),
		},
	}

	testCases := []TestCase{
		{
			name: "happy case: upsert students without parent",
			ctx:  ctx,
			req: []aggregate.DomainStudentWithAssignedParent{
				{
					DomainStudent: aggregate.DomainStudent{
						DomainStudent:    &domainStudent.Student,
						LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
						UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
						UserAccessPaths: entity.DomainUserAccessPaths{
							entity.DefaultUserAccessPath{},
						},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.usrEmailRepo.On("CreateMultiple", ctx, domainStudentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainStudentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainStudentMock.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, []string{"username1"}).Return(entity.Users{}, nil)
				domainStudentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainStudentMock.db, constant.RoleStudent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainStudentMock.locationRepo.On("RetrieveLowestLevelLocations", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainLocations{mock_usermgmt.NewLocation("", "")}, nil)
				domainStudentMock.studentPackage.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainStudentPackages{}, nil)
				domainStudentMock.db.On("Begin", ctx).Return(domainStudentMock.tx, nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
				domainStudentMock.locationRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainStudentMock.studentRepo.On("UpsertMultiple", ctx, domainStudentMock.tx, mock.Anything, mock.Anything).Return(nil)
				domainStudentMock.authUserUpserter = func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) (entity.LegacyUsers, error) {
					return entity.LegacyUsers{}, nil
				}
				domainStudentMock.OrganizationRepo.On("GetTenantIDByOrgID", ctx, domainStudentMock.tx, mock.Anything).Return("", nil)
				hashConfig := mockScryptHash()
				domainStudentMock.tenantManager.On("TenantClient", ctx, mock.Anything).Return(domainStudentMock.tenantClient, nil)
				domainStudentMock.tenantClient.On("GetHashConfig").Return(hashConfig)
				domainStudentMock.tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Return(&internal_auth_user.ImportUsersResult{}, nil)

				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
				domainStudentMock.jsm.On("TracedPublish", mock.Anything, "publishDomainUserEvent", mock.Anything, mock.Anything).Return(nil, nil)
				domainStudentMock.tx.On("Commit", ctx).Return(nil)
				// insertEnrollmentStatusHistories
				domainStudentMock.unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureUsingMasterReplicatedTable, mock.Anything, mock.Anything).Once().Return(true, nil)
				domainStudentMock.internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(entity.NullDomainConfiguration{}, nil)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("UpdateStudentStatusBasedEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "happy case: upsert students with parent",
			ctx:  ctx,
			req: []aggregate.DomainStudentWithAssignedParent{
				{
					DomainStudent: aggregate.DomainStudent{
						DomainStudent:    &domainStudentWithAssignedParent.Student,
						LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
						UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
						UserAccessPaths: entity.DomainUserAccessPaths{
							entity.DefaultUserAccessPath{},
						},
					},
					Parents: []aggregate.DomainParent{{
						DomainParent:     &domainStudentWithAssignedParent.Parents[0],
						LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
						UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
						UserAccessPaths: entity.DomainUserAccessPaths{
							entity.DefaultUserAccessPath{},
						},
					}},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainStudentMock.usrEmailRepo.On("CreateMultiple", ctx, domainStudentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainStudentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainStudentMock.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByUserNames", ctx, domainStudentMock.db, []string{"username1"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, domainStudentMock.db, []string{"external-user-id"}).Return(entity.Users{}, nil)
				domainStudentMock.userRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainStudentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainStudentMock.db, constant.RoleStudent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainStudentMock.locationRepo.On("RetrieveLowestLevelLocations", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainLocations{mock_usermgmt.NewLocation("", "")}, nil)
				domainStudentMock.studentPackage.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainStudentPackages{}, nil)
				domainStudentMock.db.On("Begin", ctx).Return(domainStudentMock.tx, nil)
				domainStudentMock.gradeRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return([]entity.DomainGrade{entity.NullDomainGrade{}}, nil)
				domainStudentMock.locationRepo.On("GetByIDs", ctx, domainStudentMock.db, mock.Anything).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainStudentMock.studentRepo.On("UpsertMultiple", ctx, domainStudentMock.tx, mock.Anything, mock.Anything).Return(nil)
				domainStudentMock.OrganizationRepo.On("GetTenantIDByOrgID", ctx, domainStudentMock.tx, mock.Anything).Return("", nil)
				hashConfig := mockScryptHash()
				domainStudentMock.tenantManager.On("TenantClient", ctx, mock.Anything).Return(domainStudentMock.tenantClient, nil)
				domainStudentMock.tenantClient.On("GetHashConfig").Return(hashConfig)
				domainStudentMock.tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Return(&internal_auth_user.ImportUsersResult{}, nil)
				domainStudentMock.authUserUpserter = func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) (entity.LegacyUsers, error) {
					return entity.LegacyUsers{}, nil
				}

				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
				domainStudentMock.jsm.On("TracedPublish", mock.Anything, "publishDomainUserEvent", mock.Anything, mock.Anything).Return(nil, nil)
				domainStudentMock.parentService.On("UpsertMultiple", ctx, mock.Anything).Once().Return([]aggregate.DomainParent{}, nil)
				domainStudentMock.StudentParentRepo.On("GetByStudentIDs", ctx, mock.Anything, mock.Anything).Once().Return(entity.DomainStudentParentRelationships{studentParent}, nil)
				domainStudentMock.StudentParentRepo.On("SoftDeleteByStudentIDs", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				domainStudentMock.studentParentRelationshipManager = func(ctx context.Context, db libdatabase.QueryExecer, org valueobj.HasOrganizationID, relationship field.String, studentIDToBeAssigned valueobj.HasUserID, parentIDsToAssign ...valueobj.HasUserID) error {
					return nil
				}
				domainStudentMock.tx.On("Commit", ctx).Return(nil)
				// insertEnrollmentStatusHistories
				domainStudentMock.unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureUsingMasterReplicatedTable, mock.Anything, mock.Anything).Once().Return(false, nil)
				domainStudentMock.configurationClient.On("GetConfigurationByKey", ctx, &mpb.GetConfigurationByKeyRequest{Key: constant.KeyEnrollmentStatusHistoryConfig}).Once().Return(&mpb.GetConfigurationByKeyResponse{
					Configuration: &mpb.Configuration{
						ConfigValue: constant.ConfigValueOn,
						ConfigKey:   constant.KeyEnrollmentStatusHistoryConfig,
					},
				}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					[]entity.DomainEnrollmentStatusHistory{
						createMockDomainEnrollmentStatusHistory("student-id", "location-id",
							upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
							time.Now().Add(-100*time.Hour),
							time.Now().Add(200*time.Hour),
							"order-id",
							1),
					}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
				domainStudentMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
				)
				domainStudentMock.parentService.On("DomainParentsToUpsert", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().
					Return([]aggregate.DomainParent{}, []aggregate.DomainParent{}, []aggregate.DomainParent{}, nil)
				domainStudentMock.parentService.On("UpsertMultipleParentsInTx", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]aggregate.DomainParent{}, nil)
				domainStudentMock.enrollmentStatusHistoryRepo.On("UpdateStudentStatusBasedEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log(testCase.name)
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.ManabieSchool),
				},
			}
			testCase.ctx = interceptors.ContextWithJWTClaims(ctx, claim)
			testCase.ctx = interceptors.ContextWithUserID(testCase.ctx, "user-id")

			m, service := DomainStudentServiceMock()
			m.userPhoneNumberRepo.On("SoftDeleteByUserIDs", testCase.ctx, m.tx, mock.Anything).Return(nil)
			m.userAddressRepo.On("SoftDeleteByUserIDs", testCase.ctx, m.tx, mock.Anything).Return(nil)
			m.taggedUserRepo.On("SoftDeleteByUserIDs", testCase.ctx, m.tx, mock.Anything).Return(nil)
			m.schoolHistoryRepo.On("SoftDeleteByStudentIDs", testCase.ctx, m.tx, mock.Anything).Return(nil)
			testCase.setupWithMock(testCase.ctx, &m)
			service.StudentParentRelationshipManager = m.studentParentRelationshipManager
			service.AuthUserUpserter = m.authUserUpserter
			option := unleash.DomainStudentFeatureOption{
				DomainUserFeatureOption: unleash.DomainUserFeatureOption{
					EnableIgnoreUpdateEmail: false,
					EnableUsername:          true,
				},
				EnableAutoDeactivateAndReactivateStudentV2:            true,
				DisableAutoDeactivateAndReactivateStudent:             false,
				EnableExperimentalBulkInsertEnrollmentStatusHistories: false,
			}
			_, err := service.UpsertMultipleWithAssignedParent(testCase.ctx, testCase.req.([]aggregate.DomainStudentWithAssignedParent), option)

			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestDomainStudent_validateStudentsLocations(t *testing.T) {
	serviceMock, student := DomainStudentServiceMock()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(3*time.Second))
	defer cancel()

	userID := idutil.ULIDNow()
	locationID1 := idutil.ULIDNow()
	locationID2 := idutil.ULIDNow()

	type args struct {
		ctx              context.Context
		db               libdatabase.QueryExecer
		studentsToUpsert []aggregate.DomainStudent
	}
	tests := []struct {
		name  string
		args  args
		setup func(serviceMock *prepareDomainStudentMock)
		err   error
	}{
		{
			name: "happy case: student locations match with active course locations",
			args: args{
				ctx: ctx,
				studentsToUpsert: []aggregate.DomainStudent{{
					DomainStudent: &grpc.DomainStudentImpl{
						UserIDAttr: field.NewString(userID),
					},
					UserAccessPaths: []entity.DomainUserAccessPath{&grpc.DomainUserAccessPathImpl{
						UserIDAttr:     field.NewString(userID),
						LocationIDAttr: field.NewString(locationID1),
					}},
				}},
			},
			setup: func(sm *prepareDomainStudentMock) {
				startAt := time.Now()
				endAt := time.Now().Add(3 * time.Hour)

				serviceMock.locationRepo.On("RetrieveLowestLevelLocations", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					entity.DomainLocations{
						mock_usermgmt.NewLocation(locationID1, locationID1),
						mock_usermgmt.NewLocation(locationID2, locationID2),
					}, nil)

				serviceMock.studentPackage.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					entity.DomainStudentPackages{
						&repository.StudentPackage{
							StudentIDAttr:   field.NewString(userID),
							StartDateAttr:   field.NewTime(startAt),
							EndDateAttr:     field.NewTime(endAt),
							LocationIDsAttr: []string{locationID1},
						},
					}, nil)
			},
			err: nil,
		},
		{
			name: "happy case: student locations are lowest level locations",
			args: args{
				ctx: ctx,
				studentsToUpsert: []aggregate.DomainStudent{{
					DomainStudent: &grpc.DomainStudentImpl{
						UserIDAttr: field.NewString(userID),
					},
					UserAccessPaths: []entity.DomainUserAccessPath{&grpc.DomainUserAccessPathImpl{
						UserIDAttr:     field.NewString(userID),
						LocationIDAttr: field.NewString(locationID1),
					}},
				}},
			},
			setup: func(studentMock *prepareDomainStudentMock) {
				startAt := time.Now().Add(-2 * 24 * time.Hour) // 2 day ago
				endAt := time.Now().Add(-1 * 24 * time.Hour)   // 1 day ago

				serviceMock.locationRepo.On("RetrieveLowestLevelLocations", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					entity.DomainLocations{
						mock_usermgmt.NewLocation(locationID1, locationID1),
						mock_usermgmt.NewLocation(locationID2, locationID2),
					}, nil)

				serviceMock.studentPackage.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainStudentPackages{
						&repository.StudentPackage{
							StudentIDAttr:   field.NewString(userID),
							StartDateAttr:   field.NewTime(startAt),
							EndDateAttr:     field.NewTime(endAt),
							LocationIDsAttr: []string{locationID2},
						},
					}, nil,
				)
			},
			err: nil,
		},
		{
			name: "bad case: student locations are not lowest level",
			args: args{
				ctx: ctx,
				studentsToUpsert: []aggregate.DomainStudent{{
					DomainStudent: &grpc.DomainStudentImpl{
						UserIDAttr: field.NewString(userID),
					},
					UserAccessPaths: []entity.DomainUserAccessPath{&grpc.DomainUserAccessPathImpl{
						UserIDAttr:     field.NewString(userID),
						LocationIDAttr: field.NewString(locationID1),
					}},
				}},
			},
			setup: func(studentMock *prepareDomainStudentMock) {
				startAt := time.Now()
				endAt := time.Now().Add(3 * time.Hour)

				serviceMock.locationRepo.On("RetrieveLowestLevelLocations", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(
					entity.DomainLocations{
						mock_usermgmt.NewLocation(locationID2, locationID2),
					}, nil)

				serviceMock.studentPackage.On("GetByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainStudentPackages{
						&repository.StudentPackage{
							StudentIDAttr:   field.NewString(userID),
							StartDateAttr:   field.NewTime(startAt),
							EndDateAttr:     field.NewTime(endAt),
							LocationIDsAttr: []string{locationID2},
						},
					}, nil,
				)
			},
			err: entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					EntityName: entity.StudentEntity,
					FieldName:  entity.StudentLocationTypeField,
					Reason:     entity.LocationIsNotLowestLocation,
					Index:      0,
				},
				NestedFieldName: entity.StudentLocationsField,
				NestedIndex:     0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)
			tt.setup(&serviceMock)
			err := student.validateStudentsLocations(tt.args.ctx, serviceMock.db, tt.args.studentsToUpsert)
			if tt.err != nil {
				assert.Equal(t, err.Error(), tt.err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestDomainStudent_setUserPhoneNumbers(t *testing.T) {
	type args struct {
		students []aggregate.DomainStudent
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "duplicated phone numbers",
			args: args{
				students: []aggregate.DomainStudent{{
					UserPhoneNumbers: []entity.DomainUserPhoneNumber{
						mock_usermgmt.NewUserPhoneNumber("0900000000", entity.UserPhoneNumberTypeStudentPhoneNumber),
						mock_usermgmt.NewUserPhoneNumber("0900000000", entity.UserPhoneNumberTypeStudentHomePhoneNumber),
					},
				}},
			},
			wantErr: errcode.Error{
				Code:      errcode.DuplicatedData,
				FieldName: fmt.Sprintf("students[%d].phone_number.home_phone_number", 0),
				Index:     0,
			},
		},
		{
			name: "2 phone numbers for multiple students",
			args: args{
				students: []aggregate.DomainStudent{
					{
						UserPhoneNumbers: []entity.DomainUserPhoneNumber{
							mock_usermgmt.NewUserPhoneNumber("0900000000", entity.UserPhoneNumberTypeStudentPhoneNumber),
							mock_usermgmt.NewUserPhoneNumber("0800000000", entity.UserPhoneNumberTypeStudentHomePhoneNumber),
						},
					},
					{
						UserPhoneNumbers: []entity.DomainUserPhoneNumber{
							mock_usermgmt.NewUserPhoneNumber("0900000000", entity.UserPhoneNumberTypeStudentPhoneNumber),
							mock_usermgmt.NewUserPhoneNumber("0800000000", entity.UserPhoneNumberTypeStudentHomePhoneNumber),
						},
					},
					{
						UserPhoneNumbers: []entity.DomainUserPhoneNumber{
							mock_usermgmt.NewUserPhoneNumber("0900000000", entity.UserPhoneNumberTypeStudentPhoneNumber),
							mock_usermgmt.NewUserPhoneNumber("0800000000", entity.UserPhoneNumberTypeStudentHomePhoneNumber),
						},
					},
					{
						UserPhoneNumbers: []entity.DomainUserPhoneNumber{
							mock_usermgmt.NewUserPhoneNumber("0900000000", entity.UserPhoneNumberTypeStudentPhoneNumber),
							mock_usermgmt.NewUserPhoneNumber("0800000000", entity.UserPhoneNumberTypeStudentHomePhoneNumber),
						},
					},
					{
						UserPhoneNumbers: []entity.DomainUserPhoneNumber{
							mock_usermgmt.NewUserPhoneNumber("0900000000", entity.UserPhoneNumberTypeStudentPhoneNumber),
							mock_usermgmt.NewUserPhoneNumber("0900000000", entity.UserPhoneNumberTypeStudentHomePhoneNumber),
						},
					},
				},
			},
			wantErr: errcode.Error{
				Code:      errcode.DuplicatedData,
				FieldName: fmt.Sprintf("students[%d].phone_number.home_phone_number", 4),
				Index:     4,
			},
		},
		{
			name: "without phone number",
			args: args{
				students: []aggregate.DomainStudent{{
					UserPhoneNumbers: []entity.DomainUserPhoneNumber{
						entity.DefaultDomainUserPhoneNumber{},
						entity.DefaultDomainUserPhoneNumber{},
					},
				}},
			},
			wantErr: nil,
		},
		{
			name: "with 2 empty phone numbers",
			args: args{
				students: []aggregate.DomainStudent{{
					UserPhoneNumbers: []entity.DomainUserPhoneNumber{
						mock_usermgmt.NewUserPhoneNumber("", entity.UserPhoneNumberTypeStudentPhoneNumber),
						mock_usermgmt.NewUserPhoneNumber("", entity.UserPhoneNumberTypeStudentHomePhoneNumber),
					},
				}},
			},
			wantErr: nil,
		},
		{
			name: "with 2 phone numbers",
			args: args{
				students: []aggregate.DomainStudent{{
					UserPhoneNumbers: []entity.DomainUserPhoneNumber{
						mock_usermgmt.NewUserPhoneNumber("0900000000", entity.UserPhoneNumberTypeStudentPhoneNumber),
						mock_usermgmt.NewUserPhoneNumber("0800000000", entity.UserPhoneNumberTypeStudentHomePhoneNumber),
					},
				}},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setUserPhoneNumbers(tt.args.students...)
			if tt.wantErr != nil || err != nil {
				e, _ := err.(errcode.Error)
				wantErr, _ := tt.wantErr.(errcode.Error)

				assert.Equal(t, e.Code, wantErr.Code)
				assert.Equal(t, e.FieldName, wantErr.FieldName)
			}
		})
	}
}

func TestDomainStudent_ValidateUpdateSystemAndExternalUserID(t *testing.T) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	domainMockStudent, service := DomainStudentServiceMock()
	userID1 := idutil.ULIDNow()
	userID2 := idutil.ULIDNow()

	type args struct {
		ctx              context.Context
		studentsToUpdate aggregate.DomainStudents
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
		setup   func()
	}{
		{
			name: "error occurred when finding user by id",
			args: args{
				ctx: ctx,
				studentsToUpdate: []aggregate.DomainStudent{
					{DomainStudent: mock_usermgmt.NewStudent(userID1, "")},
				},
			},
			wantErr: errcode.Error{
				Code: errcode.InternalError,
			},
			setup: func() {
				domainMockStudent.userRepo.
					On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).
					Once().Return(entity.Users{}, fmt.Errorf("error"))
			},
		},
		{
			name: "user id was not found in system",
			args: args{
				ctx: ctx,
				studentsToUpdate: []aggregate.DomainStudent{
					{DomainStudent: mock_usermgmt.NewStudent(userID1, "")},
					{DomainStudent: mock_usermgmt.NewStudent(userID2, "")},
				},
			},
			wantErr: errcode.Error{Code: errcode.InvalidData},
			setup: func() {
				domainMockStudent.userRepo.
					On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).
					Once().Return(entity.Users{mock_usermgmt.NewStudent(userID1, "")}, nil)
			},
		},
		{
			name: "try to update external user id when external user id is already existed",
			args: args{
				ctx: ctx,
				studentsToUpdate: []aggregate.DomainStudent{
					{DomainStudent: mock_usermgmt.NewStudent(userID1, idutil.ULIDNow())},
				},
			},
			wantErr: errcode.Error{Code: errcode.UpdateFieldFail},
			setup: func() {
				domainMockStudent.userRepo.
					On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).
					Once().Return(entity.Users{mock_usermgmt.NewStudent(userID1, idutil.ULIDNow())}, nil)
			},
		},
		{
			name: "happy case",
			args: args{
				ctx: ctx,
				studentsToUpdate: []aggregate.DomainStudent{
					{DomainStudent: mock_usermgmt.NewStudent(userID1, idutil.ULIDNow())},
				},
			},
			wantErr: nil,
			setup: func() {
				domainMockStudent.userRepo.
					On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).
					Once().Return(entity.Users{mock_usermgmt.NewStudent(userID1, "")}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			err := service.ValidateUpdateSystemAndExternalUserID(tt.args.ctx, tt.args.studentsToUpdate)
			if tt.wantErr != nil || err != nil {
				e, _ := err.(errcode.Error)
				wantErr, _ := tt.wantErr.(errcode.Error)
				assert.Equal(t, wantErr.Code, e.Code)
			}
		})
	}
}

func TestDomainStudent_validateSchoolHistories(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(3*time.Second))
	defer cancel()
	domainMockStudent, service := DomainStudentServiceMock()
	mockSchoolInfo_1 := mock_usermgmt.School{
		RandomSchool: mock_usermgmt.RandomSchool{
			SchoolID:      field.NewString("school-id-1"),
			SchoolLevelID: field.NewString("school-level-id-1"),
		},
	}
	mockSchoolInfo_2 := mock_usermgmt.School{
		RandomSchool: mock_usermgmt.RandomSchool{
			SchoolID:      field.NewString("school-id-2"),
			SchoolLevelID: field.NewString("school-level-id-2"),
		},
	}
	mockSchoolCourse_1 := mock_usermgmt.SchoolCourse{
		RandomSchoolCourse: mock_usermgmt.RandomSchoolCourse{
			SchoolCourseID: field.NewString("school-course-id-1"),
			SchoolID:       field.NewString("school-id-1"),
		},
	}
	mockSchoolCourse_2 := mock_usermgmt.SchoolCourse{
		RandomSchoolCourse: mock_usermgmt.RandomSchoolCourse{
			SchoolCourseID: field.NewString("school-course-id-2"),
			SchoolID:       field.NewString("school-id-2"),
		},
	}
	mockStartDate, _ := time.Parse(constant.DateLayout, "2020/10/10")
	mockEndDate, _ := time.Parse(constant.DateLayout, "2020/10/11")
	testCases := []TestCase{
		{
			name: "not return error with valid data",
			ctx:  ctx,
			req: []aggregate.DomainStudent{
				{
					SchoolHistories: []entity.DomainSchoolHistory{
						mock_usermgmt.SchoolHistory{
							RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
								SchoolID:       field.NewString("school-id-1"),
								SchoolCourseID: field.NewString("school-course-id-1"),
								StartDate:      field.NewTime(mockStartDate),
								EndDate:        field.NewTime(mockEndDate),
							},
						},
						mock_usermgmt.SchoolHistory{
							RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
								SchoolID:       field.NewString("school-id-2"),
								SchoolCourseID: field.NewString("school-course-id-2"),
								StartDate:      field.NewTime(mockStartDate),
								EndDate:        field.NewTime(mockEndDate),
							},
						},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				domainMockStudent.schoolRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainSchools{mockSchoolInfo_1, mockSchoolInfo_2}, nil,
				)
				domainMockStudent.schoolCourseRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainSchoolCourses{mockSchoolCourse_1, mockSchoolCourse_2}, nil,
				)
			},
		},
		{
			name: "not return error with valid data and empty optional data",
			ctx:  ctx,
			req: []aggregate.DomainStudent{
				{
					SchoolHistories: []entity.DomainSchoolHistory{
						mock_usermgmt.SchoolHistory{
							RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
								SchoolID:       field.NewString("school-id-1"),
								SchoolCourseID: field.NewString("school-course-id-1"),
								StartDate:      field.NewTime(mockStartDate),
								EndDate:        field.NewTime(mockEndDate),
							},
						},
						mock_usermgmt.SchoolHistory{
							RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
								SchoolID:       field.NewString("school-id-2"),
								SchoolCourseID: field.NewNullString(),
								StartDate:      field.NewNullTime(),
								EndDate:        field.NewNullTime(),
							},
						},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				domainMockStudent.schoolRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainSchools{mockSchoolInfo_1, mockSchoolInfo_2}, nil,
				)
				domainMockStudent.schoolCourseRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainSchoolCourses{mockSchoolCourse_1}, nil,
				)
			},
		},
		{
			name: "return error when start date is greater than end date",
			ctx:  ctx,
			req: []aggregate.DomainStudent{
				{
					SchoolHistories: []entity.DomainSchoolHistory{
						mock_usermgmt.SchoolHistory{
							RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
								SchoolID:       field.NewString("school-id-1"),
								SchoolCourseID: field.NewString("school-course-id-1"),
								StartDate:      field.NewTime(mockEndDate),
								EndDate:        field.NewTime(mockStartDate),
							},
						},
					},
				},
			},
			expectedErr: entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					EntityName: entity.StudentEntity,
					FieldName:  entity.StartDateFieldEnrollmentStatusHistory,
					Index:      0,
					Reason:     entity.StartDateAfterCurrentDate,
				},
				NestedFieldName: entity.StudentSchoolHistoryField,
				NestedIndex:     0,
			},
		},
		{
			name: "return error when schools are duplicated",
			ctx:  ctx,
			req: []aggregate.DomainStudent{
				{
					SchoolHistories: []entity.DomainSchoolHistory{
						mock_usermgmt.SchoolHistory{
							RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
								SchoolID:       field.NewString("school-id-1"),
								SchoolCourseID: field.NewString("school-course-id-1"),
								StartDate:      field.NewTime(mockStartDate),
								EndDate:        field.NewTime(mockEndDate),
							},
						},
						mock_usermgmt.SchoolHistory{
							RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
								SchoolID:       field.NewString("school-id-1"),
								SchoolCourseID: field.NewString("school-course-id-2"),
								StartDate:      field.NewTime(mockStartDate),
								EndDate:        field.NewTime(mockEndDate),
							},
						},
					},
				},
			},
			expectedErr: entity.DuplicatedFieldErrorWithArrayNestedField{
				DuplicatedFieldError: entity.DuplicatedFieldError{
					EntityName:      entity.StudentEntity,
					DuplicatedField: entity.StudentSchoolField,
					Index:           0,
				},
				NestedFieldName: entity.StudentSchoolHistoryField,
				NestedIndex:     0,
			},
			setup: func(ctx context.Context) {},
		},
		{
			name: "return error when schools are duplicate school level",
			ctx:  ctx,
			req: []aggregate.DomainStudent{
				{
					SchoolHistories: []entity.DomainSchoolHistory{
						mock_usermgmt.SchoolHistory{
							RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
								SchoolID:       field.NewString("school-id-1"),
								SchoolCourseID: field.NewString("school-course-id-1"),
								StartDate:      field.NewTime(mockStartDate),
								EndDate:        field.NewTime(mockEndDate),
							},
						},
						mock_usermgmt.SchoolHistory{
							RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
								SchoolID:       field.NewString("school-id-2"),
								SchoolCourseID: field.NewString("school-course-id-2"),
								StartDate:      field.NewTime(mockStartDate),
								EndDate:        field.NewTime(mockEndDate),
							},
						},
					},
				},
			},
			expectedErr: entity.DuplicatedFieldErrorWithArrayNestedField{
				DuplicatedFieldError: entity.DuplicatedFieldError{
					EntityName:      entity.StudentEntity,
					DuplicatedField: entity.StudentSchoolHistorySchoolLevel,
					Index:           0,
				},
				NestedFieldName: entity.StudentSchoolHistoryField,
				NestedIndex:     0,
			},
			setup: func(ctx context.Context) {
				domainMockStudent.schoolRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainSchools{mockSchoolInfo_1,
						mock_usermgmt.School{
							RandomSchool: mock_usermgmt.RandomSchool{
								SchoolID:      field.NewString("school-id-1"),
								SchoolLevelID: mockSchoolInfo_1.SchoolLevelID(),
							},
						},
					}, nil,
				)
			},
		},
		{
			name: "return error when school does not exist in DB",
			ctx:  ctx,
			req: []aggregate.DomainStudent{
				{
					SchoolHistories: []entity.DomainSchoolHistory{
						mock_usermgmt.SchoolHistory{
							RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
								SchoolID:       field.NewString("school-id-1"),
								SchoolCourseID: field.NewString("school-course-id-1"),
								StartDate:      field.NewTime(mockStartDate),
								EndDate:        field.NewTime(mockEndDate),
							},
						},
						mock_usermgmt.SchoolHistory{
							RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
								SchoolID:       field.NewString("school-id-2"),
								SchoolCourseID: field.NewString("school-course-id-2"),
								StartDate:      field.NewTime(mockStartDate),
								EndDate:        field.NewTime(mockEndDate),
							},
						},
					},
				},
			},
			expectedErr: entity.NotFoundErrorWithArrayNestedField{
				NotFoundError: entity.NotFoundError{
					EntityName: entity.StudentEntity,
					FieldName:  entity.StudentSchoolField,
					FieldValue: "school-id-2",
					Index:      0,
				},
				NestedFieldName: entity.StudentSchoolHistoryField,
				NestedIndex:     1,
			},
			setup: func(ctx context.Context) {
				domainMockStudent.schoolRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainSchools{mockSchoolInfo_1}, nil,
				)
			},
		},
		{
			name: "return error when school course does not exist in DB",
			ctx:  ctx,
			req: []aggregate.DomainStudent{
				{
					SchoolHistories: []entity.DomainSchoolHistory{
						mock_usermgmt.SchoolHistory{
							RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
								SchoolID:       field.NewString("school-id-1"),
								SchoolCourseID: field.NewString("school-course-id-1"),
								StartDate:      field.NewTime(mockStartDate),
								EndDate:        field.NewTime(mockEndDate),
							},
						},
						mock_usermgmt.SchoolHistory{
							RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
								SchoolID:       field.NewString("school-id-2"),
								SchoolCourseID: field.NewString("school-course-id-2"),
								StartDate:      field.NewTime(mockStartDate),
								EndDate:        field.NewTime(mockEndDate),
							},
						},
					},
				},
			},
			expectedErr: entity.NotFoundErrorWithArrayNestedField{
				NotFoundError: entity.NotFoundError{
					EntityName: entity.StudentEntity,
					FieldName:  entity.StudentSchoolCourseField,
					FieldValue: "school-id-2",
					Index:      0,
				},
				NestedFieldName: entity.StudentSchoolHistoryField,
				NestedIndex:     1,
			},
			setup: func(ctx context.Context) {
				domainMockStudent.schoolRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainSchools{mockSchoolInfo_1, mockSchoolInfo_2}, nil,
				)
				domainMockStudent.schoolCourseRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainSchoolCourses{mockSchoolCourse_1}, nil,
				)
			},
		},
		{
			name: "return error when school course does not belong to school",
			ctx:  ctx,
			req: []aggregate.DomainStudent{
				{
					SchoolHistories: []entity.DomainSchoolHistory{
						mock_usermgmt.SchoolHistory{
							RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
								SchoolID:       field.NewString("school-id-1"),
								SchoolCourseID: field.NewString("school-course-id-1"),
								StartDate:      field.NewTime(mockStartDate),
								EndDate:        field.NewTime(mockEndDate),
							},
						},
						mock_usermgmt.SchoolHistory{
							RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
								SchoolID:       field.NewString("school-id-2"),
								SchoolCourseID: field.NewString("school-course-id-2"),
								StartDate:      field.NewTime(mockStartDate),
								EndDate:        field.NewTime(mockEndDate),
							},
						},
					},
				},
			},
			expectedErr: entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					EntityName: entity.StudentEntity,
					FieldName:  entity.StudentSchoolCourseField,
					Index:      0,
					Reason:     entity.SchoolCourseDoesNotBelongToSchool,
				},
				NestedFieldName: entity.StudentSchoolHistoryField,
				NestedIndex:     0,
			},
			setup: func(ctx context.Context) {
				domainMockStudent.schoolRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainSchools{mockSchoolInfo_1, mockSchoolInfo_2}, nil,
				)
				domainMockStudent.schoolCourseRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainSchoolCourses{
						mockSchoolCourse_1,
						mock_usermgmt.SchoolCourse{
							RandomSchoolCourse: mock_usermgmt.RandomSchoolCourse{
								SchoolCourseID: field.NewString("school-course-id-2"),
								SchoolID:       field.NewString("school-id-1"),
							},
						},
					}, nil,
				)
			},
		},
		{
			name: "return error when query schools failed",
			ctx:  ctx,
			req: []aggregate.DomainStudent{
				{
					SchoolHistories: []entity.DomainSchoolHistory{
						mock_usermgmt.SchoolHistory{
							RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
								SchoolID:       field.NewString("school-id-1"),
								SchoolCourseID: field.NewString("school-course-id-1"),
								StartDate:      field.NewTime(mockStartDate),
								EndDate:        field.NewTime(mockEndDate),
							},
						},
					},
				},
			},
			expectedErr: repository.InternalError{RawError: errors.Wrap(fmt.Errorf("query school failed"), "db.Query")},
			setup: func(ctx context.Context) {
				domainMockStudent.schoolRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainSchools{}, fmt.Errorf("query school failed"),
				)
			},
		},
		{
			name: "return error when query school courses failed",
			ctx:  ctx,
			req: []aggregate.DomainStudent{
				{
					SchoolHistories: []entity.DomainSchoolHistory{
						mock_usermgmt.SchoolHistory{
							RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
								SchoolID:       field.NewString("school-id-1"),
								SchoolCourseID: field.NewString("school-course-id-1"),
								StartDate:      field.NewTime(mockStartDate),
								EndDate:        field.NewTime(mockEndDate),
							},
						},
					},
				},
			},
			expectedErr: repository.InternalError{RawError: errors.Wrap(fmt.Errorf("query school course failed"), "db.Query")},
			setup: func(ctx context.Context) {
				domainMockStudent.schoolRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainSchools{mockSchoolInfo_1}, fmt.Errorf("query school failed"),
				)
				domainMockStudent.schoolCourseRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainSchoolCourses{}, fmt.Errorf("query school course failed"),
				)
			},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)
			if tt.setup != nil {
				tt.setup(ctx)
			}

			err := service.validateSchoolHistories(tt.ctx, tt.req.([]aggregate.DomainStudent)...)
			if tt.expectedErr != nil {
				e, _ := err.(errcode.Error)
				expectedErr, _ := tt.expectedErr.(errcode.Error)
				assert.Equal(t, expectedErr.Code, e.Code)
				assert.Equal(t, expectedErr.FieldName, e.FieldName)
			}
		})
	}
}

func TestDomainStudent_upsertUserInIdentityPlatform(t *testing.T) {
	service, domainStudent := DomainStudentServiceMock()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ctx = context.WithValue(
		ctx,
		interceptors.JwtClaims(0),
		&interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{
				ResourcePath: fmt.Sprint(constants.ManabieSchool),
			},
		},
	)
	organization, _ := interceptors.OrganizationFromContext(ctx)
	createStudent := aggregate.DomainStudent{
		DomainStudent: &mock_usermgmt.Student{
			RandomStudent: mock_usermgmt.RandomStudent{
				UserID: field.NewString(idutil.ULIDNow()),
				Email:  field.NewString(idutil.ULIDNow()),
			},
		},
	}
	updateStudent := aggregate.DomainStudent{
		DomainStudent: &mock_usermgmt.Student{
			RandomStudent: mock_usermgmt.RandomStudent{
				UserID: field.NewString(idutil.ULIDNow()),
				Email:  field.NewString(idutil.ULIDNow()),
			},
		},
	}

	type args struct {
		ctx              context.Context
		tx               libdatabase.Ext
		organization     *interceptors.Organization
		studentsToCreate aggregate.DomainStudents
		studentsToUpdate aggregate.DomainStudents
		option           unleash.DomainUserFeatureOption
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		wantErr error
	}{
		{
			name: "ignore update email true",
			args: args{
				ctx:              ctx,
				tx:               service.tx,
				organization:     organization,
				studentsToCreate: []aggregate.DomainStudent{createStudent},
				studentsToUpdate: []aggregate.DomainStudent{updateStudent},
				option: unleash.DomainUserFeatureOption{
					EnableIgnoreUpdateEmail: true,
					EnableUsername:          true,
				},
			},
			setup: func() {
				hashConfig := mockScryptHash()
				service.OrganizationRepo.On("GetTenantIDByOrgID", mock.Anything, mock.Anything, mock.Anything).Once().Return(idutil.ULIDNow(), nil)
				service.tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(service.tenantClient, nil)
				service.tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				service.tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				service.unleashClient.On("IsFeatureEnabledOnOrganization", featureToggleIgnoreEmailValidation, mock.Anything, mock.Anything).Once().Return(true, nil)
			},
			wantErr: nil,
		},
		{
			name: "ignore update email false",
			args: args{
				ctx:              ctx,
				tx:               service.tx,
				organization:     organization,
				studentsToCreate: []aggregate.DomainStudent{createStudent},
				studentsToUpdate: []aggregate.DomainStudent{updateStudent},
				option: unleash.DomainUserFeatureOption{
					EnableIgnoreUpdateEmail: false,
					EnableUsername:          true,
				},
			},
			setup: func() {
				hashConfig := mockScryptHash()
				service.OrganizationRepo.On("GetTenantIDByOrgID", mock.Anything, mock.Anything, mock.Anything).Once().Return(idutil.ULIDNow(), nil)
				service.tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(service.tenantClient, nil)
				service.tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				service.tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				service.unleashClient.On("IsFeatureEnabledOnOrganization", featureToggleIgnoreEmailValidation, mock.Anything, mock.Anything).Once().Return(false, nil)
				service.tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(service.tenantClient, nil)
				service.tenantClient.On("GetUser", ctx, mock.Anything).Once().Return(nil, nil)
				service.tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			err := domainStudent.upsertUserInIdentityPlatform(tt.args.ctx, tt.args.tx, tt.args.organization, tt.args.studentsToCreate.Users(), tt.args.studentsToUpdate.Users(), tt.args.option)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestDomainStudent_GetEmailWithStudentID(t *testing.T) {
	service, domainStudent := DomainStudentServiceMock()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	updateStudent1 := aggregate.DomainStudent{
		DomainStudent: &mock_usermgmt.Student{
			RandomStudent: mock_usermgmt.RandomStudent{
				UserID: field.NewString("user_id-1"),
				Email:  field.NewString("email-1"),
			},
		},
	}
	updateStudent2 := aggregate.DomainStudent{
		DomainStudent: &mock_usermgmt.Student{
			RandomStudent: mock_usermgmt.RandomStudent{
				UserID: field.NewString("user_id-2"),
				Email:  field.NewString("email-2"),
			},
		},
	}

	type args struct {
		ctx        context.Context
		studentIDs []string
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		want    map[string]entity.User
		wantErr error
	}{
		{
			name: "find user with only one existed user",
			args: args{
				ctx:        ctx,
				studentIDs: []string{"user_id-1", ""},
			},
			setup: func() {
				service.userRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).
					Once().Return(entity.Users{updateStudent1}, nil)
			},
			want:    map[string]entity.User{"user_id-1": updateStudent1},
			wantErr: nil,
		},
		{
			name: "find user with two existed user",
			args: args{
				ctx:        ctx,
				studentIDs: []string{"user_id-1", "user_id-2"},
			},
			setup: func() {
				service.userRepo.On("GetByIDs", mock.Anything, mock.Anything, []string{"user_id-1", "user_id-2"}).
					Once().Return(entity.Users{updateStudent1, updateStudent2}, nil)
			},
			want: map[string]entity.User{
				"user_id-1": updateStudent1,
				"user_id-2": updateStudent2,
			},
			wantErr: nil,
		},
		{
			name: "error orcured when finding",
			args: args{
				ctx:        ctx,
				studentIDs: []string{"user_id-1", "user_id-2"},
			},
			setup: func() {
				service.userRepo.On("GetByIDs", mock.Anything, mock.Anything, []string{"user_id-1", "user_id-2"}).
					Once().Return(entity.Users{}, assert.AnError)
			},
			want: nil,
			wantErr: errcode.Error{
				Code: errcode.InternalError,
				Err:  fmt.Errorf("service.UserRepo.GetByIDs: %w", assert.AnError),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			got, err := domainStudent.GetEmailWithStudentID(tt.args.ctx, tt.args.studentIDs)
			assert.Equalf(t, tt.want, got, "GetEmailWithStudentID(%v, %v)", tt.args.ctx, tt.args.studentIDs)
			assert.Equalf(t, tt.wantErr, err, "GetEmailWithStudentID(%v, %v)", tt.args.ctx, tt.args.studentIDs)
		})
	}
}

func TestDomainStudent_validateUserAddress(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(3*time.Second))
	defer cancel()
	domainMockStudent, service := DomainStudentServiceMock()
	testCases := []TestCase{
		{
			name: "not return error with valid data",
			ctx:  ctx,
			req: []aggregate.DomainStudent{
				{
					UserAddress: mock_usermgmt.UserAddress{
						RandomUserAddress: mock_usermgmt.RandomUserAddress{
							PostalCode:   field.NewString("700000"),
							PrefectureID: field.NewString("prefecture-id"),
						},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				domainMockStudent.prefectureRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]entity.DomainPrefecture{mock_usermgmt.Prefecture{RandomPrefecture: mock_usermgmt.RandomPrefecture{
						PrefectureID: field.NewString("prefecture-id"),
					}}}, nil,
				)
			},
		},
		{
			name: "return error when prefecture not exist in db",
			ctx:  ctx,
			req: []aggregate.DomainStudent{
				{
					UserAddress: mock_usermgmt.UserAddress{
						RandomUserAddress: mock_usermgmt.RandomUserAddress{
							PostalCode:   field.NewString("700000"),
							PrefectureID: field.NewString("prefecture-id"),
						},
					},
				},
			},
			expectedErr: entity.InvalidFieldError{
				EntityName: entity.StudentEntity,
				FieldName:  entity.StudentUserAddressPrefectureField,
				Index:      0,
			},
			setup: func(ctx context.Context) {
				domainMockStudent.prefectureRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]entity.DomainPrefecture{}, nil,
				)
			},
		},
		{
			name: "return error when query prefecture error",
			ctx:  ctx,
			req: []aggregate.DomainStudent{
				{
					UserAddress: mock_usermgmt.UserAddress{
						RandomUserAddress: mock_usermgmt.RandomUserAddress{
							PostalCode:   field.NewString("700000"),
							PrefectureID: field.NewString("prefecture-id"),
						},
					},
				},
			},
			expectedErr: entity.InvalidFieldError{
				EntityName: entity.StudentEntity,
				FieldName:  entity.StudentUserAddressPrefectureField,
				Index:      0,
			},
			setup: func(ctx context.Context) {
				domainMockStudent.prefectureRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]entity.DomainPrefecture{}, fmt.Errorf("query error"),
				)
			},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)
			if tt.setup != nil {
				tt.setup(ctx)
			}

			err := service.validateUserAddress(tt.ctx, tt.req.([]aggregate.DomainStudent)...)
			if tt.expectedErr != nil {
				e, _ := err.(errcode.Error)
				expectedErr, _ := tt.expectedErr.(errcode.Error)
				assert.Equal(t, expectedErr.Code, e.Code)
				assert.Equal(t, expectedErr.FieldName, e.FieldName)
			}
		})
	}
}

func TestDomainStudent_validateTags(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(3*time.Second))
	defer cancel()

	domainMockStudent, service := DomainStudentServiceMock()

	type args struct {
		ctx      context.Context
		students []aggregate.DomainStudent
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
		setup   func()
	}{
		{
			name: "invalid tags: id is not in db",
			args: args{
				ctx: ctx,
				students: []aggregate.DomainStudent{
					{
						TaggedUsers: entity.DomainTaggedUsers{
							mock_usermgmt.NewTaggedUser("tag-id-1", "user-id-1"),
							mock_usermgmt.NewTaggedUser("tag-id-2", "user-id-1"),
						},
					},
				},
			},
			wantErr: entity.InvalidFieldError{
				EntityName: entity.StudentEntity,
				FieldName:  entity.StudentTagsField,
				Index:      0,
			},
			setup: func() {
				domainMockStudent.tagRepo.On("GetByIDs", mock.Anything, mock.Anything, []string{"tag-id-1", "tag-id-2"}).Once().Return(
					entity.DomainTags{mock_usermgmt.NewTag("tag-id-1", entity.UserTagTypeStudent)}, nil,
				)
			},
		},
		{
			name: "invalid tags: id is not in db for multiple students",
			args: args{
				ctx: ctx,
				students: []aggregate.DomainStudent{
					{
						TaggedUsers: entity.DomainTaggedUsers{
							mock_usermgmt.NewTaggedUser("tag-id-1", "user-id-1"),
						},
					},
					{
						TaggedUsers: entity.DomainTaggedUsers{
							mock_usermgmt.NewTaggedUser("tag-id-1", "user-id-2"),
							mock_usermgmt.NewTaggedUser("tag-id-2", "user-id-2"),
						},
					},
				},
			},
			wantErr: entity.InvalidFieldError{
				EntityName: entity.StudentEntity,
				FieldName:  entity.StudentTagsField,
				Index:      0,
			},
			setup: func() {
				domainMockStudent.tagRepo.On("GetByIDs", mock.Anything, mock.Anything, []string{"tag-id-1"}).
					Once().Return(
					entity.DomainTags{mock_usermgmt.NewTag("tag-id-1", entity.UserTagTypeStudent)}, nil,
				)
				domainMockStudent.tagRepo.On("GetByIDs", mock.Anything, mock.Anything, []string{"tag-id-1", "tag-id-2"}).
					Once().Return(
					entity.DomainTags{mock_usermgmt.NewTag("tag-id-1", entity.UserTagTypeStudent)}, nil,
				)
			},
		},
		{
			name: "invalid tags: tag type is not student",
			args: args{
				ctx: ctx,
				students: []aggregate.DomainStudent{
					{
						TaggedUsers: entity.DomainTaggedUsers{
							mock_usermgmt.NewTaggedUser("tag-id-1", "user-id-1"),
						},
					},
				},
			},
			wantErr: entity.InvalidFieldError{
				EntityName: entity.StudentEntity,
				FieldName:  entity.StudentTagsField,
				Index:      0,
				Reason:     entity.InvalidTagType,
			},
			setup: func() {
				domainMockStudent.tagRepo.On("GetByIDs", mock.Anything, mock.Anything, []string{"tag-id-1"}).Once().Return(
					entity.DomainTags{mock_usermgmt.NewTag("tag-id-1", entity.UserTagTypeParent)}, nil,
				)
			},
		},
		{
			name: "valid tag",
			args: args{
				ctx: ctx,
				students: []aggregate.DomainStudent{
					{
						TaggedUsers: entity.DomainTaggedUsers{
							mock_usermgmt.NewTaggedUser("tag-id-1", "user-id-1"),
						},
					},
				},
			},
			wantErr: nil,
			setup: func() {
				domainMockStudent.tagRepo.On("GetByIDs", mock.Anything, mock.Anything, []string{"tag-id-1"}).Once().Return(
					entity.DomainTags{mock_usermgmt.NewTag("tag-id-1", entity.UserTagTypeStudent)}, nil,
				)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)
			if tt.setup != nil {
				tt.setup()
			}

			err := service.validateTags(tt.args.ctx, tt.args.students...)
			if err != nil || tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			}
		})
	}
}

func TestDomainStudent_UpdateUserActivation(t *testing.T) {
	serviceMock, domainStudent := DomainStudentServiceMock()
	testCases := []struct {
		name          string
		users         entity.Users
		expectedError error
		setup         func()
	}{
		{
			name: "can not get users by ids",
			users: entity.Users{
				mock_usermgmt.User{
					RandomUser: mock_usermgmt.RandomUser{
						UserID:        field.NewString("user-id-1"),
						DeactivatedAt: field.NewTime(time.Now()),
					},
				},
				mock_usermgmt.User{
					RandomUser: mock_usermgmt.RandomUser{
						UserID:        field.NewString("user-id-2"),
						DeactivatedAt: field.NewTime(time.Now()),
					},
				},
			},
			expectedError: errcode.Error{
				Code: errcode.InternalError,
				Err:  errors.Wrap(assert.AnError, "userRepo.GetByIDs"),
			},
			setup: func() {
				serviceMock.userRepo.On("GetByIDs", mock.Anything, mock.Anything, []string{"user-id-1", "user-id-2"}).
					Once().Return(entity.Users{}, assert.AnError)
			},
		},
		{
			name: "user ids are not valid",
			users: entity.Users{
				mock_usermgmt.User{
					RandomUser: mock_usermgmt.RandomUser{
						UserID:        field.NewString("user-id-1"),
						DeactivatedAt: field.NewNullTime(),
					},
				},
				mock_usermgmt.User{
					RandomUser: mock_usermgmt.RandomUser{
						UserID:        field.NewString("user-id-2"),
						DeactivatedAt: field.NewTime(time.Now()),
					},
				},
			},
			expectedError: errcode.Error{
				Code: errcode.InvalidData,
				Err:  fmt.Errorf("user ids are not valid"),
			},
			setup: func() {
				serviceMock.userRepo.On("GetByIDs", mock.Anything, mock.Anything, []string{"user-id-1", "user-id-2"}).
					Once().Return(entity.Users{
					&repository.User{ID: field.NewString("user-id-1")},
				}, nil)
			},
		},
		{
			name: "Activate and Deactivate users",
			users: entity.Users{
				mock_usermgmt.User{
					RandomUser: mock_usermgmt.RandomUser{
						UserID:        field.NewString("user-id-1"),
						DeactivatedAt: field.NewNullTime(),
					},
				},
				mock_usermgmt.User{
					RandomUser: mock_usermgmt.RandomUser{
						UserID:        field.NewString("user-id-2"),
						DeactivatedAt: field.NewTime(time.Now()),
					},
				},
			},
			expectedError: nil,
			setup: func() {
				serviceMock.userRepo.On("GetByIDs", mock.Anything, mock.Anything, []string{"user-id-1", "user-id-2"}).
					Once().Return(entity.Users{
					&repository.User{ID: field.NewString("user-id-1")},
					&repository.User{ID: field.NewString("user-id-2")},
				}, nil)
				serviceMock.db.On("Begin", mock.Anything).Return(serviceMock.tx, nil)
				serviceMock.userRepo.On("UpdateActivation", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				serviceMock.tx.On("Commit", mock.Anything).Return(nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup()
			}

			err := domainStudent.UpdateUserActivation(context.Background(), tc.users)
			if tc.expectedError != nil || err != nil {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			}
		})
	}
}

func TestDomainUser_generateUserIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	studentToCreate := &mock_usermgmt.Student{
		RandomStudent: mock_usermgmt.RandomStudent{
			Email:      field.NewString("test@manabie.com"),
			UserName:   field.NewString("username"),
			LoginEmail: field.NewString("test@manabie.com"),
		},
	}
	studentToUpdate := &mock_usermgmt.Student{
		RandomStudent: mock_usermgmt.RandomStudent{
			UserID:   field.NewString(idutil.ULIDNow()),
			Email:    field.NewString("test2@manabie.com"),
			UserName: field.NewString("username2"),
		},
	}
	type args struct {
		students         []aggregate.DomainStudent
		isEnableUsername bool
	}
	testCases := []struct {
		name                     string
		ctx                      context.Context
		args                     args
		setupWithMock            func(ctx context.Context, genericMock interface{})
		expectedErr              error
		expectedStudentsToCreate []aggregate.DomainStudent
		expectedStudentsToUpdate []aggregate.DomainStudent
	}{
		{
			name: "happy case: only student to create",
			ctx:  ctx,
			args: args{
				students: []aggregate.DomainStudent{
					{
						DomainStudent: studentToCreate,
					},
				},
				isEnableUsername: true,
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}

				domainStudentMock.usrEmailRepo.On("CreateMultiple", ctx, domainStudentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
			},
			expectedErr: nil,
			expectedStudentsToCreate: []aggregate.DomainStudent{
				{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							UserID:     field.NewString("new-id"),
							Email:      field.NewString("test@manabie.com"),
							UserName:   field.NewString("username"),
							LoginEmail: field.NewString("new-id@manabie.com"),
						},
					},
				},
			},
		},
		{
			name: "happy case: only student to update",
			ctx:  ctx,
			args: args{
				students: []aggregate.DomainStudent{
					{
						DomainStudent: studentToUpdate,
					},
				},
				isEnableUsername: true,
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}

				domainStudentMock.usrEmailRepo.On("CreateMultiple", ctx, domainStudentMock.db, mock.Anything).Return(valueobj.HasUserIDs{}, nil)
			},
			expectedErr: nil,
			expectedStudentsToUpdate: []aggregate.DomainStudent{
				{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							UserID:   studentToUpdate.UserID(),
							Email:    field.NewString("test2@manabie.com"),
							UserName: field.NewString("username2"),
						},
					},
				},
			},
		},
		{
			name: "happy case: bth student to create and update",
			ctx:  ctx,
			args: args{
				students: []aggregate.DomainStudent{
					{
						DomainStudent: studentToCreate,
					},
					{
						DomainStudent: studentToUpdate,
					},
				},
				isEnableUsername: true,
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}

				domainStudentMock.usrEmailRepo.On("CreateMultiple", ctx, domainStudentMock.db, mock.Anything).Once().Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
			},
			expectedErr: nil,
			expectedStudentsToCreate: []aggregate.DomainStudent{
				{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							UserID:     field.NewString("new-id"),
							Email:      field.NewString("test@manabie.com"),
							UserName:   field.NewString("username"),
							LoginEmail: field.NewString("new-id@manabie.com"),
						},
					},
				},
			},
			expectedStudentsToUpdate: []aggregate.DomainStudent{
				{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							UserID:   studentToUpdate.UserID(),
							Email:    field.NewString("test2@manabie.com"),
							UserName: field.NewString("username2"),
						},
					},
				},
			},
		},
		{
			name: "should return correct student profile when disable username",
			ctx:  ctx,
			args: args{
				students: []aggregate.DomainStudent{
					{
						DomainStudent: studentToCreate,
					},
				},
				isEnableUsername: false,
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainStudentMock, ok := genericMock.(*prepareDomainStudentMock)
				if !ok {
					t.Error("invalid mock")
				}

				domainStudentMock.usrEmailRepo.On("CreateMultiple", ctx, domainStudentMock.db, mock.Anything).Once().Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
			},
			expectedErr: nil,
			expectedStudentsToCreate: []aggregate.DomainStudent{
				{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							UserID:     field.NewString("new-id"),
							Email:      field.NewString("test@manabie.com"),
							UserName:   field.NewString("username"),
							LoginEmail: field.NewString("test@manabie.com"),
						},
					},
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)

			m, service := DomainStudentServiceMock()
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.ManabieSchool),
				},
			}
			tt.ctx = interceptors.ContextWithJWTClaims(ctx, claim)
			tt.ctx = interceptors.ContextWithUserID(tt.ctx, "user-id")
			tt.setupWithMock(tt.ctx, &m)

			studentsToCreate, studentsToUpdate, err := service.generateUserIDs(tt.ctx, tt.args.isEnableUsername, tt.args.students...)
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, len(tt.expectedStudentsToCreate), len(studentsToCreate))
				assert.Equal(t, len(tt.expectedStudentsToUpdate), len(studentsToUpdate))
				for i, student := range tt.expectedStudentsToCreate {
					assert.Equal(t, student.UserID(), studentsToCreate[i].UserID())
					assert.Equal(t, student.Email(), studentsToCreate[i].Email())
					assert.Equal(t, student.UserName(), studentsToCreate[i].UserName())
					assert.Equal(t, student.LoginEmail(), studentsToCreate[i].LoginEmail())
				}
				for i, student := range tt.expectedStudentsToUpdate {
					assert.Equal(t, student.UserID(), studentsToUpdate[i].UserID())
					assert.Equal(t, student.Email(), studentsToUpdate[i].Email())
					assert.Equal(t, student.UserName(), studentsToUpdate[i].UserName())
				}
			}
		})
	}
}

func TestDomainStudent_IsAuthUsernameConfigEnabled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	testCases := []struct {
		name        string
		ctx         context.Context
		setup       func(ctx context.Context, m *prepareDomainStudentMock)
		expect      bool
		expectedErr error
	}{
		{
			name: "should return true when config is on",
			ctx:  ctx,
			setup: func(ctx context.Context, m *prepareDomainStudentMock) {
				m.internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
			},
			expect: true,
		},
		{
			name: "should return true when config is off",
			ctx:  ctx,
			setup: func(ctx context.Context, m *prepareDomainStudentMock) {
				m.internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("off"),
					},
				}, nil)
			},
			expect: false,
		},
		{
			name: "should return false when config is not on db",
			ctx:  ctx,
			setup: func(ctx context.Context, m *prepareDomainStudentMock) {
				m.internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(entity.NullDomainConfiguration{}, pgx.ErrNoRows)
			},
			expect: false,
		},
		{
			name: "should return error when get config error",
			ctx:  ctx,
			setup: func(ctx context.Context, m *prepareDomainStudentMock) {
				m.internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(entity.NullDomainConfiguration{}, fmt.Errorf("get config error"))
			},
			expectedErr: fmt.Errorf("get config error"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			m, service := DomainStudentServiceMock()
			testCase.setup(testCase.ctx, &m)
			res, err := service.IsAuthUsernameConfigEnabled(testCase.ctx)

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expect, res)
			}
		})
	}
}

func TestDomainStudent_validateExternalUserIDExistedInSystem(t *testing.T) {
	t.Parallel()

	student1 := &mock_usermgmt.Student{
		RandomStudent: mock_usermgmt.RandomStudent{
			UserID:         field.NewString("UserID01"),
			ExternalUserID: field.NewString("ExternalUserID01"),
		},
	}
	student2 := &mock_usermgmt.Student{
		RandomStudent: mock_usermgmt.RandomStudent{
			UserID:         field.NewString("UserID02"),
			ExternalUserID: field.NewString("ExternalUserID02"),
		},
	}

	type args struct {
		ctx   context.Context
		users entity.Users
	}
	tests := []struct {
		name    string
		args    args
		setup   func() *DomainStudent
		wantErr error
	}{
		{
			name: "",
			args: args{
				ctx:   context.Background(),
				users: entity.Users{student1, student2},
			},
			setup: func() *DomainStudent {
				mockRepo, service := DomainStudentServiceMock()
				mockRepo.userRepo.
					On("GetByExternalUserIDs", mock.Anything, mock.Anything, []string{student1.ExternalUserID().String(), student2.ExternalUserID().String()}).
					Return(entity.Users{student1, student2}, nil)

				mockRepo.studentRepo.
					On("GetUsersByExternalUserIDs", mock.Anything, mock.Anything, []string{student1.ExternalUserID().String(), student2.ExternalUserID().String()}).
					Return(entity.Users{student1, student2}, nil)

				return &service
			},
			wantErr: nil,
		},
		{
			name: "",
			args: args{
				ctx:   context.Background(),
				users: entity.Users{student1, student2},
			},
			setup: func() *DomainStudent {
				mockRepo, service := DomainStudentServiceMock()
				mockRepo.userRepo.
					On("GetByExternalUserIDs", mock.Anything, mock.Anything, []string{student1.ExternalUserID().String(), student2.ExternalUserID().String()}).
					Return(entity.Users{student1, student2}, nil)

				mockRepo.studentRepo.
					On("GetUsersByExternalUserIDs", mock.Anything, mock.Anything, []string{student1.ExternalUserID().String(), student2.ExternalUserID().String()}).
					Return(entity.Users{student1}, nil)

				return &service
			},
			wantErr: entity.ExistingDataError{
				FieldName:  string(entity.UserFieldExternalUserID),
				EntityName: entity.UserEntity,
				Index:      1,
			},
		},
		{
			name: "",
			args: args{
				ctx:   context.Background(),
				users: entity.Users{student1, student2},
			},
			setup: func() *DomainStudent {
				mockRepo, service := DomainStudentServiceMock()
				mockRepo.userRepo.
					On("GetByExternalUserIDs", mock.Anything, mock.Anything, []string{student1.ExternalUserID().String(), student2.ExternalUserID().String()}).
					Return(entity.Users{student1, student2}, nil)

				mockRepo.studentRepo.
					On("GetUsersByExternalUserIDs", mock.Anything, mock.Anything, []string{student1.ExternalUserID().String(), student2.ExternalUserID().String()}).
					Return(entity.Users{student2}, nil)

				return &service
			},
			wantErr: entity.ExistingDataError{
				FieldName:  string(entity.UserFieldExternalUserID),
				EntityName: entity.UserEntity,
				Index:      0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setup()
			err := service.validateExternalUserIDExistedInSystem(tt.args.ctx, tt.args.users)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
