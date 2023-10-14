package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/manabie-com/backend/internal/golibs/constants"
	libdatabase "github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	mock_usermgmt "github.com/manabie-com/backend/internal/usermgmt/pkg/mock"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	mock_multitenant "github.com/manabie-com/backend/mock/golibs/auth/multitenant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func DomainParentServiceMock() (prepareDomainParentMock, DomainParent) {
	m := prepareDomainParentMock{
		db:                            &mock_database.Ext{},
		tx:                            &mock_database.Tx{},
		firebaseAuthClient:            &mock_multitenant.TenantClient{},
		jsm:                           &mock_nats.JetStreamManagement{},
		tenantManager:                 &mock_multitenant.TenantManager{},
		tenantClient:                  &mock_multitenant.TenantClient{},
		organizationRepo:              &mock_repositories.OrganizationRepo{},
		userRepo:                      &mock_repositories.MockDomainUserRepo{},
		usrEmailRepo:                  &mock_repositories.MockDomainUsrEmailRepo{},
		userGroupRepo:                 &mock_repositories.MockDomainUserGroupRepo{},
		parentRepo:                    &mock_repositories.MockDomainParentRepo{},
		taggedUserRepo:                &mock_repositories.MockDomainTaggedUserRepo{},
		tagRepo:                       &mock_repositories.MockDomainTagRepo{},
		studentParentRelationshipRepo: &mock_repositories.MockDomainStudentParentRelationshipRepo{},
		userPhoneNumberRepo:           &mock_repositories.MockDomainUserPhoneNumberRepo{},
		unleashClient:                 &mock_unleash_client.UnleashClientInstance{},
		assignParentToStudentsManager: nil,
		internalConfigurationRepo:     &mock_repositories.MockDomainInternalConfigurationRepo{},
	}
	service := DomainParent{
		DB:                        m.db,
		JSM:                       m.jsm,
		FirebaseAuthClient:        m.firebaseAuthClient,
		TenantManager:             m.tenantManager,
		OrganizationRepo:          m.organizationRepo,
		UserRepo:                  m.userRepo,
		UsrEmailRepo:              m.usrEmailRepo,
		UserGroupRepo:             m.userGroupRepo,
		ParentRepo:                m.parentRepo,
		TagRepo:                   m.tagRepo,
		TaggedUserRepo:            m.taggedUserRepo,
		StudentParentRepo:         m.studentParentRelationshipRepo,
		AuthUserUpserter:          nil,
		UserPhoneNumberRepo:       m.userPhoneNumberRepo,
		UnleashClient:             m.unleashClient,
		InternalConfigurationRepo: m.internalConfigurationRepo,
	}
	return m, service
}

type prepareDomainParentMock struct {
	db                            *mock_database.Ext
	tx                            *mock_database.Tx
	firebaseAuthClient            *mock_multitenant.TenantClient
	jsm                           *mock_nats.JetStreamManagement
	tenantManager                 *mock_multitenant.TenantManager
	tenantClient                  *mock_multitenant.TenantClient
	organizationRepo              *mock_repositories.OrganizationRepo
	userRepo                      *mock_repositories.MockDomainUserRepo
	usrEmailRepo                  *mock_repositories.MockDomainUsrEmailRepo
	userGroupRepo                 *mock_repositories.MockDomainUserGroupRepo
	parentRepo                    *mock_repositories.MockDomainParentRepo
	taggedUserRepo                *mock_repositories.MockDomainTaggedUserRepo
	tagRepo                       *mock_repositories.MockDomainTagRepo
	studentParentRelationshipRepo *mock_repositories.MockDomainStudentParentRelationshipRepo
	authUserUpserter              AuthUserUpserter
	userPhoneNumberRepo           *mock_repositories.MockDomainUserPhoneNumberRepo
	assignParentToStudentsManager AssignParentToStudentsManager
	unleashClient                 *mock_unleash_client.UnleashClientInstance
	internalConfigurationRepo     *mock_repositories.MockDomainInternalConfigurationRepo
}

func TestDomainParent_UpsertMultiple(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	domainParent := &mock_usermgmt.Parent{
		RandomParent: mock_usermgmt.RandomParent{
			EmailAttr:     field.NewString("test@manabie.com"),
			UserNameAttr:  field.NewString("username.test@manabie.com"),
			FirstNameAttr: field.NewString("test first name"),
			LastNameAttr:  field.NewString("test last name"),
		},
	}
	existingDomainParent := &mock_usermgmt.Parent{
		RandomParent: mock_usermgmt.RandomParent{
			EmailAttr:     field.NewString("test@manabie.com"),
			UserNameAttr:  field.NewString("username.test@manabie.com"),
			UserID:        field.NewString("parent_id"),
			FirstNameAttr: field.NewString("test first name"),
			LastNameAttr:  field.NewString("test last name"),
		},
	}

	testCases := []TestCase{
		{
			name: "happy case: create parents",
			ctx:  ctx,
			req: []aggregate.DomainParent{
				{
					DomainParent:     domainParent,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainParentMock, ok := genericMock.(*prepareDomainParentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainParentMock.userRepo.On("GetByExternalUserIDs", ctx, domainParentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainParentMock.parentRepo.On("GetUsersByExternalUserIDs", ctx, domainParentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainParentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainParentMock.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
				domainParentMock.userRepo.On("GetByUserNames", ctx, domainParentMock.db, []string{"username.test@manabie.com"}).Return(entity.Users{}, nil)
				domainParentMock.usrEmailRepo.On("CreateMultiple", ctx, domainParentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainParentMock.userRepo.On("GetByIDs", ctx, domainParentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainParentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainParentMock.db, constant.RoleParent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainParentMock.db.On("Begin", ctx).Return(domainParentMock.tx, nil)
				domainParentMock.parentRepo.On("UpsertMultiple", ctx, domainParentMock.tx, mock.Anything, mock.Anything).Return(nil)
				domainParentMock.organizationRepo.On("GetTenantIDByOrgID", ctx, domainParentMock.tx, mock.Anything).Return("", nil)
				domainParentMock.taggedUserRepo.On("SoftDeleteByUserIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainParentMock.userPhoneNumberRepo.On("SoftDeleteByUserIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainParentMock.authUserUpserter = func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) (entity.LegacyUsers, error) {
					return entity.LegacyUsers{}, nil
				}
				domainParentMock.tx.On("Commit", ctx).Return(nil)
				// domainParentMock.jsm.On("TracedPublish", mock.Anything, "publishDomainUserEvent", mock.Anything, mock.Anything).Return(nil, nil)
				// domainParentMock.tx.On("Commit", ctx).Return(nil)
			},
		},
		{
			name: "happy case: update parents",
			ctx:  ctx,
			req: []aggregate.DomainParent{
				{
					DomainParent:     existingDomainParent,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainParentMock, ok := genericMock.(*prepareDomainParentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainParentMock.userRepo.On("GetByExternalUserIDs", ctx, domainParentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainParentMock.parentRepo.On("GetUsersByExternalUserIDs", ctx, domainParentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainParentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainParentMock.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
				domainParentMock.userRepo.On("GetByUserNames", ctx, domainParentMock.db, []string{"username.test@manabie.com"}).Return(entity.Users{}, nil)
				domainParentMock.usrEmailRepo.On("CreateMultiple", ctx, domainParentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainParentMock.userRepo.On("GetByIDs", ctx, domainParentMock.db, []string{"user-id"}).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainParentMock.userRepo.On("GetByIDs", ctx, domainParentMock.db, []string{"parent_id"}).Return(entity.Users{domainParent}, nil)
				domainParentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainParentMock.db, constant.RoleParent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainParentMock.db.On("Begin", ctx).Return(domainParentMock.tx, nil)
				domainParentMock.usrEmailRepo.On("UpdateEmail", ctx, domainParentMock.tx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				domainParentMock.parentRepo.On("UpsertMultiple", ctx, domainParentMock.tx, mock.Anything, mock.Anything).Return(nil)
				domainParentMock.taggedUserRepo.On("SoftDeleteByUserIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainParentMock.userPhoneNumberRepo.On("SoftDeleteByUserIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainParentMock.authUserUpserter = func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) (entity.LegacyUsers, error) {
					return entity.LegacyUsers{}, nil
				}
				domainParentMock.tx.On("Commit", ctx).Return(nil)
			},
		},
		{
			name: "happy case: create parent success with tags",
			ctx:  ctx,
			req: []aggregate.DomainParent{
				{
					DomainParent:     domainParent,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					TaggedUsers:      entity.DomainTaggedUsers{&entity.EmptyDomainTaggedUser{}},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainParentMock, ok := genericMock.(*prepareDomainParentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainParentMock.userRepo.On("GetByExternalUserIDs", ctx, domainParentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainParentMock.parentRepo.On("GetUsersByExternalUserIDs", ctx, domainParentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainParentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainParentMock.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
				domainParentMock.userRepo.On("GetByUserNames", ctx, domainParentMock.db, []string{"username.test@manabie.com"}).Return(entity.Users{}, nil)
				domainParentMock.usrEmailRepo.On("CreateMultiple", ctx, domainParentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainParentMock.userRepo.On("GetByIDs", ctx, domainParentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainParentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainParentMock.db, constant.RoleParent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainParentMock.db.On("Begin", ctx).Return(domainParentMock.tx, nil)
				domainParentMock.usrEmailRepo.On("UpdateEmail", ctx, domainParentMock.tx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				domainParentMock.parentRepo.On("UpsertMultiple", ctx, domainParentMock.tx, mock.Anything, mock.Anything).Return(nil)
				/*domainParentMock.organizationRepo.On("GetTenantIDByOrgID", ctx, domainParentMock.tx, mock.Anything).Return("", nil)
				domainParentMock.tenantManager.On("TenantClient", ctx, mock.Anything).Return(domainParentMock.tenantClient, nil)
				domainParentMock.tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				domainParentMock.tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)*/
				domainParentMock.tagRepo.On("GetByIDs", ctx, mock.Anything, mock.Anything).Return(entity.DomainTags{&repository.Tag{
					TagAttribute: repository.TagAttribute{
						TagType: field.NewString(entity.UserTagTypeParent),
					},
				}}, nil)
				domainParentMock.taggedUserRepo.On("SoftDeleteByUserIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainParentMock.taggedUserRepo.On("UpsertBatch", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainParentMock.userPhoneNumberRepo.On("SoftDeleteByUserIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainParentMock.authUserUpserter = func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) (entity.LegacyUsers, error) {
					return entity.LegacyUsers{}, nil
				}
				domainParentMock.tx.On("Commit", ctx).Return(nil)
			},
		},
		{
			name: "bad case: invalid tag",
			ctx:  ctx,
			expectedErr: entity.InvalidFieldError{
				EntityName: entity.ParentEntity,
				Index:      0,
				FieldName:  entity.ParentTagsField,
				Reason:     entity.Invalid,
			},
			req: []aggregate.DomainParent{
				{
					DomainParent:     domainParent,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					TaggedUsers:      entity.DomainTaggedUsers{&entity.EmptyDomainTaggedUser{}},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainParentMock, ok := genericMock.(*prepareDomainParentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainParentMock.userRepo.On("GetByExternalUserIDs", ctx, domainParentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainParentMock.parentRepo.On("GetUsersByExternalUserIDs", ctx, domainParentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainParentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainParentMock.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
				domainParentMock.userRepo.On("GetByUserNames", ctx, domainParentMock.db, []string{"username.test@manabie.com"}).Return(entity.Users{}, nil)
				domainParentMock.usrEmailRepo.On("CreateMultiple", ctx, domainParentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainParentMock.tagRepo.On("GetByIDs", ctx, mock.Anything, mock.Anything).Return(entity.DomainTags{&repository.Tag{
					TagAttribute: repository.TagAttribute{
						TagType: field.NewString(entity.UserTagTypeStudent),
					},
				}}, nil)
			},
		},
		{
			name: "bad case: duplicated email",
			ctx:  ctx,
			expectedErr: entity.DuplicatedFieldError{
				EntityName:      entity.UserEntity,
				DuplicatedField: string(entity.UserFieldEmail),
				Index:           1,
			},
			req: []aggregate.DomainParent{
				{
					DomainParent: domainParent,
				},
				{
					DomainParent: existingDomainParent,
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
			},
		},
		{
			name: "bad case: duplicated email (case-insensitive)",
			ctx:  ctx,
			expectedErr: entity.DuplicatedFieldError{
				EntityName:      entity.UserEntity,
				DuplicatedField: string(entity.UserFieldEmail),
				Index:           1,
			},
			req: []aggregate.DomainParent{
				{
					DomainParent: domainParent,
				},
				{
					DomainParent: &mock_usermgmt.Parent{
						RandomParent: mock_usermgmt.RandomParent{
							EmailAttr:     field.NewString("TEST@manabie.com"),
							UserNameAttr:  field.NewString("username"),
							FirstNameAttr: field.NewString("test first name"),
							LastNameAttr:  field.NewString("test last name"),
						},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
			},
		},
		{
			name: "bad case: missing username",
			ctx:  ctx,
			expectedErr: entity.MissingMandatoryFieldError{
				Index:      0,
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
			},
			req: []aggregate.DomainParent{
				{
					DomainParent: &mock_usermgmt.Parent{
						RandomParent: mock_usermgmt.RandomParent{
							FirstNameAttr: field.NewString("test first name"),
							LastNameAttr:  field.NewString("test last name"),
						},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
			},
		},
		{
			name: "bad case: username is duplicated",
			ctx:  ctx,
			expectedErr: entity.DuplicatedFieldError{
				Index:           1,
				DuplicatedField: string(entity.UserFieldUserName),
				EntityName:      entity.UserEntity,
			},
			req: []aggregate.DomainParent{
				{
					DomainParent: domainParent,
				},
				{
					DomainParent: &mock_usermgmt.Parent{
						RandomParent: mock_usermgmt.RandomParent{
							EmailAttr:     field.NewString("test1@manabie.com"),
							UserNameAttr:  field.NewString("username.test@manabie.com"),
							FirstNameAttr: field.NewString("test first name"),
							LastNameAttr:  field.NewString("test last name"),
						},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
			},
		},
		{
			name: "bad case: username is already existed",
			ctx:  ctx,
			expectedErr: entity.ExistingDataError{
				Index:      0,
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
			},
			req: []aggregate.DomainParent{
				{
					DomainParent: domainParent,
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainParentMock, ok := genericMock.(*prepareDomainParentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainParentMock.userRepo.On("GetByExternalUserIDs", ctx, domainParentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainParentMock.parentRepo.On("GetUsersByExternalUserIDs", ctx, domainParentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainParentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainParentMock.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
				domainParentMock.userRepo.On("GetByUserNames", ctx, domainParentMock.db, []string{"username.test@manabie.com"}).Return(entity.Users{
					&MockDomainUser{
						UsernameAttr: field.NewString("username.test@manabie.com"),
						userID:       field.NewString("user-id"),
					},
				}, nil)
			},
		},
		{
			name: "bad case: duplicated user_id",
			ctx:  ctx,
			expectedErr: entity.DuplicatedFieldError{
				EntityName:      entity.UserEntity,
				DuplicatedField: string(entity.UserFieldUserID),
				Index:           1,
			},
			req: []aggregate.DomainParent{
				{
					DomainParent: existingDomainParent,
				},
				{
					DomainParent: existingDomainParent,
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
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
			req: []aggregate.DomainParent{
				{
					DomainParent:     domainParent,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainParentMock, ok := genericMock.(*prepareDomainParentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainParentMock.userRepo.On("GetByExternalUserIDs", ctx, domainParentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainParentMock.parentRepo.On("GetUsersByExternalUserIDs", ctx, domainParentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainParentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainParentMock.db, []string{"test@manabie.com"}).Return(entity.Users{&mock_usermgmt.Student{
					RandomStudent: mock_usermgmt.RandomStudent{
						UserID: field.NewString("existed-user-id"),
						Email:  field.NewString("test@manabie.com"),
					}}}, nil)
			},
		},
		{
			name: "happy case: create parents with user phone numbers",
			ctx:  ctx,
			req: []aggregate.DomainParent{
				{
					DomainParent:     domainParent,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserPhoneNumbers: entity.DomainUserPhoneNumbers{entity.DefaultDomainUserPhoneNumber{}},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainParentMock, ok := genericMock.(*prepareDomainParentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainParentMock.userRepo.On("GetByExternalUserIDs", ctx, domainParentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainParentMock.parentRepo.On("GetUsersByExternalUserIDs", ctx, domainParentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainParentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainParentMock.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
				domainParentMock.userRepo.On("GetByUserNames", ctx, domainParentMock.db, []string{"username.test@manabie.com"}).Return(entity.Users{}, nil)
				domainParentMock.usrEmailRepo.On("CreateMultiple", ctx, domainParentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainParentMock.userRepo.On("GetByIDs", ctx, domainParentMock.db, mock.Anything).Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainParentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainParentMock.db, constant.RoleParent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainParentMock.db.On("Begin", ctx).Return(domainParentMock.tx, nil)
				domainParentMock.parentRepo.On("UpsertMultiple", ctx, domainParentMock.tx, mock.Anything, mock.Anything).Return(nil)
				domainParentMock.organizationRepo.On("GetTenantIDByOrgID", ctx, domainParentMock.tx, mock.Anything).Return("", nil)
				domainParentMock.taggedUserRepo.On("SoftDeleteByUserIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainParentMock.userPhoneNumberRepo.On("SoftDeleteByUserIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainParentMock.userPhoneNumberRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainParentMock.authUserUpserter = func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) (entity.LegacyUsers, error) {
					return entity.LegacyUsers{}, nil
				}
				domainParentMock.tx.On("Commit", ctx).Return(nil)
			},
		},
		{
			name: "bad case: invalid user phone number",
			ctx:  ctx,
			expectedErr: entity.InvalidFieldError{
				EntityName: entity.UserEntity,
				FieldName:  entity.StudentFieldPrimaryPhoneNumber,
				Index:      0,
			},
			req: []aggregate.DomainParent{
				{
					DomainParent:     domainParent,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserPhoneNumbers: entity.DomainUserPhoneNumbers{
						&repository.UserPhoneNumber{
							UserPhoneNumberAttribute: repository.UserPhoneNumberAttribute{
								PhoneNumber: field.NewString("asd"),
								Type:        field.NewString(entity.ParentPrimaryPhoneNumber),
							},
						},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainParentMock, ok := genericMock.(*prepareDomainParentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainParentMock.usrEmailRepo.On("CreateMultiple", ctx, domainParentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainParentMock.userRepo.On("GetByExternalUserIDs", ctx, domainParentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainParentMock.parentRepo.On("GetUsersByExternalUserIDs", ctx, domainParentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainParentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainParentMock.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
			},
		},
		{
			name: "bad case: duplicated user phone number",
			ctx:  ctx,
			expectedErr: entity.DuplicatedFieldError{
				EntityName:      entity.UserEntity,
				DuplicatedField: entity.StudentFieldSecondaryPhoneNumber,
				Index:           0,
			},
			req: []aggregate.DomainParent{
				{
					DomainParent:     domainParent,
					LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
					UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					UserPhoneNumbers: entity.DomainUserPhoneNumbers{
						&repository.UserPhoneNumber{
							UserPhoneNumberAttribute: repository.UserPhoneNumberAttribute{
								PhoneNumber: field.NewString("0987654321"),
								Type:        field.NewString(entity.ParentPrimaryPhoneNumber),
							},
						},
						&repository.UserPhoneNumber{
							UserPhoneNumberAttribute: repository.UserPhoneNumberAttribute{
								PhoneNumber: field.NewString("0987654321"),
								Type:        field.NewString(entity.ParentSecondaryPhoneNumber),
							},
						},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainParentMock, ok := genericMock.(*prepareDomainParentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainParentMock.usrEmailRepo.On("CreateMultiple", ctx, domainParentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainParentMock.userRepo.On("GetByExternalUserIDs", ctx, domainParentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainParentMock.parentRepo.On("GetUsersByExternalUserIDs", ctx, domainParentMock.db, mock.Anything).Return(entity.Users{}, nil)
				domainParentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainParentMock.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
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

			m, service := DomainParentServiceMock()
			testCase.setupWithMock(testCase.ctx, &m)
			service.AuthUserUpserter = m.authUserUpserter
			option := unleash.DomainParentFeatureOption{
				DomainUserFeatureOption: unleash.DomainUserFeatureOption{
					EnableIgnoreUpdateEmail: false,
					EnableUsername:          true,
				},
			}
			if testCase.option != nil {
				option = testCase.option.(unleash.DomainParentFeatureOption)
			}
			_, err := service.UpsertMultiple(testCase.ctx, option, testCase.req.([]aggregate.DomainParent)...)
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

func TestDomainParent_UpsertMultipleWithChildren(t *testing.T) {
	ctx := context.Background()

	domainParent := &mock_usermgmt.Parent{
		RandomParent: mock_usermgmt.RandomParent{
			EmailAttr:     field.NewString("test@manabie.com"),
			UserNameAttr:  field.NewString("username.test@manabie.com"),
			FirstNameAttr: field.NewString("test first name"),
			LastNameAttr:  field.NewString("test last name"),
		},
	}
	existingDomainUser := &mock_usermgmt.Parent{
		RandomParent: mock_usermgmt.RandomParent{
			EmailAttr:     field.NewString("user@manabie.com"),
			UserNameAttr:  field.NewString("username.user@manabie.com"),
			UserID:        field.NewString("user_id"),
			FirstNameAttr: field.NewString("test first name"),
			LastNameAttr:  field.NewString("test last name"),
		},
	}
	existingDomainParent := &mock_usermgmt.Parent{
		RandomParent: mock_usermgmt.RandomParent{
			EmailAttr:     field.NewString("test@manabie.com"),
			UserNameAttr:  field.NewString("username.test@manabie.com"),
			UserID:        field.NewString("parent_id"),
			FirstNameAttr: field.NewString("test first name"),
			LastNameAttr:  field.NewString("test last name"),
		},
	}
	existingDomainParentWithUserName := &mock_usermgmt.Parent{
		RandomParent: mock_usermgmt.RandomParent{
			EmailAttr:     field.NewString("test1@manabie.com"),
			UserNameAttr:  field.NewString("username.test@manabie.com"),
			UserID:        field.NewString("parent_id"),
			FirstNameAttr: field.NewString("test first name"),
			LastNameAttr:  field.NewString("test last name"),
		},
	}

	children := &mock_usermgmt.StudentParentRelationship{
		RandomStudentParentRelationship: mock_usermgmt.RandomStudentParentRelationship{
			StudentIDAttr:    field.NewString("student_id"),
			ParentIDAttr:     field.NewString("parent_id"),
			RelationshipAttr: field.NewString(string(constant.FamilyRelationshipFather)),
		},
	}
	testCases := []TestCase{
		{
			name: "good case: upsert parent with children",
			ctx:  ctx,
			req: []aggregate.DomainParentWithChildren{
				{
					DomainParent: aggregate.DomainParent{
						DomainParent:     domainParent,
						LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
						UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
					},
					Children: entity.DomainStudentParentRelationships{
						children,
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainParentMock, ok := genericMock.(*prepareDomainParentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainParentMock.userRepo.On("GetByExternalUserIDs", ctx, domainParentMock.db, mock.Anything).Once().Return(entity.Users{}, nil)
				domainParentMock.parentRepo.On("GetUsersByExternalUserIDs", ctx, domainParentMock.db, mock.Anything).Once().Return(entity.Users{}, nil)
				domainParentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainParentMock.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
				domainParentMock.userRepo.On("GetByUserNames", ctx, domainParentMock.db, []string{"username.test@manabie.com"}).Return(entity.Users{}, nil)
				domainParentMock.usrEmailRepo.On("CreateMultiple", ctx, domainParentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainParentMock.userRepo.On("GetByIDs", ctx, domainParentMock.db, mock.Anything).Twice().Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainParentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainParentMock.db, constant.RoleParent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainParentMock.db.On("Begin", ctx).Return(domainParentMock.tx, nil)
				domainParentMock.parentRepo.On("UpsertMultiple", ctx, domainParentMock.tx, mock.Anything, mock.Anything).Return(nil)
				domainParentMock.organizationRepo.On("GetTenantIDByOrgID", ctx, domainParentMock.tx, mock.Anything).Return("", nil)

				domainParentMock.taggedUserRepo.On("SoftDeleteByUserIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainParentMock.userRepo.On("GetByIDs", ctx, domainParentMock.tx, mock.Anything).Once().Return(entity.Users{existingDomainUser}, nil)
				domainParentMock.userPhoneNumberRepo.On("SoftDeleteByUserIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainParentMock.studentParentRelationshipRepo.On("GetByParentIDs", ctx, domainParentMock.tx, mock.Anything).Once().Return(entity.DomainStudentParentRelationships{children}, nil)
				domainParentMock.jsm.On("TracedPublish", mock.Anything, "publishDomainUserEvent", constants.SubjectUserUpdated, mock.Anything).Once().Return(nil, nil) // soft delete relationship
				domainParentMock.jsm.On("TracedPublish", mock.Anything, "publishDomainUserEvent", constants.SubjectUserCreated, mock.Anything).Once().Return(nil, nil) // create parent
				domainParentMock.jsm.On("TracedPublish", mock.Anything, "publishDomainUserEvent", constants.SubjectUserUpdated, mock.Anything).Once().Return(nil, nil) // assign parent to student
				domainParentMock.studentParentRelationshipRepo.On("SoftDeleteByParentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainParentMock.authUserUpserter = func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) (entity.LegacyUsers, error) {
					return entity.LegacyUsers{}, nil
				}
				domainParentMock.assignParentToStudentsManager = func(ctx context.Context, db libdatabase.QueryExecer, org valueobj.HasOrganizationID, relationship field.String, parentIDtoBeAssigned valueobj.HasUserID, studentIDsToBeAssigned ...valueobj.HasStudentID) error {
					return nil
				}
				domainParentMock.tx.On("Commit", ctx).Return(nil)
			},
		},
		{
			name: "happy case: update parents",
			ctx:  ctx,
			req: []aggregate.DomainParentWithChildren{
				{
					DomainParent: aggregate.DomainParent{
						DomainParent: existingDomainParent,
					},
					Children: entity.DomainStudentParentRelationships{
						children,
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainParentMock, ok := genericMock.(*prepareDomainParentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainParentMock.userRepo.On("GetByExternalUserIDs", ctx, domainParentMock.db, mock.Anything).Once().Return(entity.Users{}, nil)
				domainParentMock.parentRepo.On("GetUsersByExternalUserIDs", ctx, domainParentMock.db, mock.Anything).Once().Return(entity.Users{}, nil)
				domainParentMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, domainParentMock.db, []string{"test@manabie.com"}).Return(entity.Users{}, nil)
				domainParentMock.userRepo.On("GetByUserNames", ctx, domainParentMock.db, []string{"username.test@manabie.com"}).Return(entity.Users{}, nil)
				domainParentMock.usrEmailRepo.On("CreateMultiple", ctx, domainParentMock.db, mock.Anything).Return(valueobj.HasUserIDs{
					repository.UsrEmail{
						UsrID: field.NewString("new-id"),
					},
				}, nil)
				domainParentMock.userRepo.On("GetByIDs", ctx, domainParentMock.db, []string{"user-id"}).Once().Return(entity.Users{entity.NullDomainSchoolAdmin{}}, nil)
				domainParentMock.userRepo.On("GetByIDs", ctx, domainParentMock.db, []string{"parent_id"}).Twice().Return(entity.Users{domainParent}, nil)
				domainParentMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, domainParentMock.db, constant.RoleParent).Return(entity.UserGroupWillBeDelegated{}, nil)
				domainParentMock.db.On("Begin", ctx).Return(domainParentMock.tx, nil)
				domainParentMock.usrEmailRepo.On("UpdateEmail", ctx, domainParentMock.tx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				domainParentMock.parentRepo.On("UpsertMultiple", ctx, domainParentMock.tx, mock.Anything, mock.Anything).Return(nil)

				domainParentMock.userRepo.On("GetByIDs", ctx, domainParentMock.tx, mock.Anything).Once().Return(entity.Users{existingDomainUser}, nil)
				domainParentMock.studentParentRelationshipRepo.On("GetByParentIDs", ctx, domainParentMock.tx, mock.Anything).Once().Return(entity.DomainStudentParentRelationships{children}, nil)
				domainParentMock.jsm.On("TracedPublish", mock.Anything, "publishDomainUserEvent", constants.SubjectUserUpdated, mock.Anything).Once().Return(nil, nil) // soft delete relationship
				domainParentMock.jsm.On("TracedPublish", mock.Anything, "publishDomainUserEvent", constants.SubjectUserCreated, mock.Anything).Once().Return(nil, nil) // create parent
				domainParentMock.jsm.On("TracedPublish", mock.Anything, "publishDomainUserEvent", constants.SubjectUserUpdated, mock.Anything).Once().Return(nil, nil) // assign parent to student

				domainParentMock.taggedUserRepo.On("SoftDeleteByUserIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainParentMock.userPhoneNumberRepo.On("SoftDeleteByUserIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainParentMock.authUserUpserter = func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) (entity.LegacyUsers, error) {
					return entity.LegacyUsers{}, nil
				}
				domainParentMock.studentParentRelationshipRepo.On("SoftDeleteByParentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainParentMock.assignParentToStudentsManager = func(ctx context.Context, db libdatabase.QueryExecer, org valueobj.HasOrganizationID, relationship field.String, parentIDtoBeAssigned valueobj.HasUserID, studentIDsToBeAssigned ...valueobj.HasStudentID) error {
					return nil
				}
				domainParentMock.tx.On("Commit", ctx).Return(nil)
			},
		},
		{
			name: "bad case: duplicated email",
			ctx:  ctx,
			expectedErr: entity.DuplicatedFieldError{
				EntityName:      entity.UserEntity,
				DuplicatedField: string(entity.UserFieldEmail),
				Index:           1,
			},
			req: []aggregate.DomainParentWithChildren{
				{

					DomainParent: aggregate.DomainParent{
						DomainParent: domainParent,
					},
					Children: entity.DomainStudentParentRelationships{
						children,
					},
				},
				{
					DomainParent: aggregate.DomainParent{
						DomainParent: existingDomainParent,
					},
					Children: entity.DomainStudentParentRelationships{
						children,
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainParentMock, ok := genericMock.(*prepareDomainParentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainParentMock.db.On("Begin", mock.Anything).Return(domainParentMock.tx, nil)
				domainParentMock.tx.On("Rollback", ctx).Return(nil)
			},
		},
		{
			name: "bad case: duplicated username",
			ctx:  ctx,
			expectedErr: entity.DuplicatedFieldError{
				EntityName:      entity.UserEntity,
				DuplicatedField: string(entity.UserFieldUserName),
				Index:           1,
			},
			req: []aggregate.DomainParentWithChildren{
				{

					DomainParent: aggregate.DomainParent{
						DomainParent: domainParent,
					},
					Children: entity.DomainStudentParentRelationships{
						children,
					},
				},
				{
					DomainParent: aggregate.DomainParent{
						DomainParent: existingDomainParentWithUserName,
					},
					Children: entity.DomainStudentParentRelationships{
						children,
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainParentMock, ok := genericMock.(*prepareDomainParentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainParentMock.db.On("Begin", mock.Anything).Return(domainParentMock.tx, nil)
				domainParentMock.tx.On("Rollback", ctx).Return(nil)
			},
		},
		{
			name: "bad case: duplicated email (case-insensitive)",
			ctx:  ctx,
			expectedErr: entity.DuplicatedFieldError{
				EntityName:      entity.UserEntity,
				DuplicatedField: string(entity.UserFieldEmail),
				Index:           1,
			},
			req: []aggregate.DomainParentWithChildren{
				{
					DomainParent: aggregate.DomainParent{
						DomainParent: domainParent,
					},
				},
				{
					DomainParent: aggregate.DomainParent{
						DomainParent: &mock_usermgmt.Parent{
							RandomParent: mock_usermgmt.RandomParent{
								EmailAttr:     field.NewString("Test@manabie.com"),
								UserID:        field.NewString("parent_id"),
								UserNameAttr:  field.NewString("username"),
								FirstNameAttr: field.NewString("test first name"),
								LastNameAttr:  field.NewString("test last name"),
							},
						},
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				domainParentMock, ok := genericMock.(*prepareDomainParentMock)
				if !ok {
					t.Error("invalid mock")
				}
				domainParentMock.db.On("Begin", mock.Anything).Return(domainParentMock.tx, nil)
				domainParentMock.tx.On("Rollback", ctx).Return(nil)
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

			m, service := DomainParentServiceMock()
			testCase.setupWithMock(testCase.ctx, &m)
			service.AuthUserUpserter = m.authUserUpserter
			service.AssignParentToStudentsManager = m.assignParentToStudentsManager
			option := unleash.DomainParentFeatureOption{
				DomainUserFeatureOption: unleash.DomainUserFeatureOption{
					EnableIgnoreUpdateEmail: true,
					EnableUsername:          true,
				},
			}
			if testCase.option != nil {
				option = testCase.option.(unleash.DomainParentFeatureOption)
			}
			_, err := service.UpsertMultipleWithChildren(testCase.ctx, option, testCase.req.([]aggregate.DomainParentWithChildren)...)
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

func TestDomainStudent_isAuthUsernameConfigEnabled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	testCases := []struct {
		name        string
		ctx         context.Context
		setup       func(ctx context.Context, m *prepareDomainParentMock)
		expect      bool
		expectedErr error
	}{
		{
			name: "should return true when config is on",
			ctx:  ctx,
			setup: func(ctx context.Context, m *prepareDomainParentMock) {
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
			setup: func(ctx context.Context, m *prepareDomainParentMock) {
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
			setup: func(ctx context.Context, m *prepareDomainParentMock) {
				m.internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(entity.NullDomainConfiguration{}, pgx.ErrNoRows)
			},
			expect: false,
		},
		{
			name: "should return error when get config error",
			ctx:  ctx,
			setup: func(ctx context.Context, m *prepareDomainParentMock) {
				m.internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(entity.NullDomainConfiguration{}, fmt.Errorf("get config error"))
			},
			expectedErr: fmt.Errorf("get config error"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			m, service := DomainParentServiceMock()
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

func TestDomainParent_validateExternUserIDUsedByOtherRole(t *testing.T) {
	t.Parallel()

	parent1 := &mock_usermgmt.Parent{
		RandomParent: mock_usermgmt.RandomParent{
			UserID:             field.NewString("userID1"),
			ExternalUserIDAttr: field.NewString("ExternalUserIDAttr1"),
			EmailAttr:          field.NewString("test@manabie.com"),
			UserNameAttr:       field.NewString("username.test@manabie.com"),
			FirstNameAttr:      field.NewString("test first name"),
			LastNameAttr:       field.NewString("test last name"),
		},
	}
	parent2 := &mock_usermgmt.Parent{
		RandomParent: mock_usermgmt.RandomParent{
			UserID:             field.NewString("userID1"),
			ExternalUserIDAttr: field.NewString("ExternalUserIDAttr2"),
			EmailAttr:          field.NewString("test@manabie.com"),
			UserNameAttr:       field.NewString("username.test@manabie.com"),
			FirstNameAttr:      field.NewString("test first name"),
			LastNameAttr:       field.NewString("test last name"),
		},
	}

	type args struct {
		ctx     context.Context
		parents aggregate.DomainParents
	}
	tests := []struct {
		name    string
		args    args
		setup   func() *DomainParent
		wantErr error
	}{
		{
			name: "happy case",
			args: args{
				ctx: context.Background(),
				parents: []aggregate.DomainParent{
					{DomainParent: parent1},
					{DomainParent: parent2},
				},
			},
			setup: func() *DomainParent {
				serviceMock, parent := DomainParentServiceMock()

				serviceMock.userRepo.
					On("GetByExternalUserIDs", mock.Anything, mock.Anything, []string{parent1.ExternalUserIDAttr.String(), parent2.ExternalUserIDAttr.String()}).
					Return(entity.Users{parent1, parent2}, nil)

				serviceMock.parentRepo.
					On("GetUsersByExternalUserIDs", mock.Anything, mock.Anything, []string{parent1.ExternalUserIDAttr.String(), parent2.ExternalUserIDAttr.String()}).
					Return(entity.Users{parent1, parent2}, nil)

				return &parent
			},
			wantErr: nil,
		},
		{
			name: "bad case 1: external_user_id found in list users but not found in list parent",
			args: args{
				ctx: context.Background(),
				parents: []aggregate.DomainParent{
					{DomainParent: parent1},
					{DomainParent: parent2},
				},
			},
			setup: func() *DomainParent {
				serviceMock, parent := DomainParentServiceMock()

				serviceMock.userRepo.
					On("GetByExternalUserIDs", mock.Anything, mock.Anything, []string{parent1.ExternalUserIDAttr.String(), parent2.ExternalUserIDAttr.String()}).
					Return(entity.Users{parent1, parent2}, nil)

				serviceMock.parentRepo.
					On("GetUsersByExternalUserIDs", mock.Anything, mock.Anything, []string{parent1.ExternalUserIDAttr.String(), parent2.ExternalUserIDAttr.String()}).
					Return(entity.Users{parent1}, nil)

				return &parent
			},
			wantErr: entity.ExistingDataError{
				FieldName:  string(entity.UserFieldExternalUserID),
				EntityName: entity.ParentEntity,
				Index:      1,
			},
		},
		{
			name: "bad case 2: external_user_id found in list users but not found in list parent",
			args: args{
				ctx: context.Background(),
				parents: []aggregate.DomainParent{
					{DomainParent: parent1},
					{DomainParent: parent2},
				},
			},
			setup: func() *DomainParent {
				serviceMock, parent := DomainParentServiceMock()

				serviceMock.userRepo.
					On("GetByExternalUserIDs", mock.Anything, mock.Anything, []string{parent1.ExternalUserIDAttr.String(), parent2.ExternalUserIDAttr.String()}).
					Return(entity.Users{parent1, parent2}, nil)

				serviceMock.parentRepo.
					On("GetUsersByExternalUserIDs", mock.Anything, mock.Anything, []string{parent1.ExternalUserIDAttr.String(), parent2.ExternalUserIDAttr.String()}).
					Return(entity.Users{parent2}, nil)

				return &parent
			},
			wantErr: entity.ExistingDataError{
				FieldName:  string(entity.UserFieldExternalUserID),
				EntityName: entity.ParentEntity,
				Index:      0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setup()
			err := service.validateExternUserIDUsedByOtherRole(tt.args.ctx, tt.args.parents)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
