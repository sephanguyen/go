package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	internal_auth_user "github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/port/grpc"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	entity_mock "github.com/manabie-com/backend/internal/usermgmt/pkg/mock"
	mock_usermgmt "github.com/manabie-com/backend/internal/usermgmt/pkg/mock"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	mock_multitenant "github.com/manabie-com/backend/mock/golibs/auth/multitenant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_firebase "github.com/manabie-com/backend/mock/golibs/firebase"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestImportParent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	tx := new(mock_database.Tx)
	orgRepo := new(mock_repositories.OrganizationRepo)
	usrEmailRepo := new(mock_repositories.MockUsrEmailRepo)
	userRepo := new(mock_repositories.MockUserRepo)
	studentRepo := new(mock_repositories.MockStudentRepo)
	userGroupRepo := new(mock_repositories.MockUserGroupRepo)
	userAccessPathRepo := new(mock_repositories.MockUserAccessPathRepo)
	userPhoneNumberRepo := new(mock_repositories.MockUserPhoneNumberRepo)
	studentParentRepo := new(mock_repositories.MockStudentParentRepo)
	jsm := new(mock_nats.JetStreamManagement)
	importUserEventRepo := new(mock_repositories.MockImportUserEventRepo)
	firebaseAuth := new(mock_firebase.AuthClient)
	tenantManager := new(mock_multitenant.TenantManager)
	firebaseAuthClient := new(mock_multitenant.TenantClient)
	userGroupsMemberRepo := new(mock_repositories.MockUserGroupsMemberRepo)
	userGroupV2Repo := new(mock_repositories.MockUserGroupV2Repo)
	tenantClient := &mock_multitenant.TenantClient{}
	parentRepo := new(mock_repositories.MockParentRepo)
	taskQueue := &mockTaskQueue{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	domainTagRepo := new(mock_repositories.MockDomainTagRepo)
	domainTaggedUserRepo := new(mock_repositories.MockDomainTaggedUserRepo)
	domainUserRepo := new(mock_repositories.MockDomainUserRepo)
	internalConfigurationRepo := new(mock_repositories.MockDomainInternalConfigurationRepo)
	domainStudentRepo := new(mock_repositories.MockDomainStudentRepo)
	// sampleID := uuid.NewString()

	userModifierService := UserModifierService{
		DB:                        db,
		OrganizationRepo:          orgRepo,
		UsrEmailRepo:              usrEmailRepo,
		UserRepo:                  userRepo,
		StudentRepo:               studentRepo,
		ParentRepo:                parentRepo,
		UserGroupRepo:             userGroupRepo,
		UserAccessPathRepo:        userAccessPathRepo,
		UserPhoneNumberRepo:       userPhoneNumberRepo,
		FirebaseClient:            firebaseAuth,
		StudentParentRepo:         studentParentRepo,
		UserGroupsMemberRepo:      userGroupsMemberRepo,
		UserGroupV2Repo:           userGroupV2Repo,
		FirebaseAuthClient:        firebaseAuthClient,
		TenantManager:             tenantManager,
		JSM:                       jsm,
		ImportUserEventRepo:       importUserEventRepo,
		DomainTagRepo:             domainTagRepo,
		DomainTaggedUserRepo:      domainTaggedUserRepo,
		TaskQueue:                 taskQueue,
		UnleashClient:             mockUnleashClient,
		DomainUserRepo:            domainUserRepo,
		InternalConfigurationRepo: internalConfigurationRepo,
		DomainStudentRepo:         domainStudentRepo,
	}

	// userIds := []string{}
	student1 := &entity.LegacyStudent{}
	student1.ID.Set("student-01")

	student2 := &entity.LegacyStudent{}
	student2.ID.Set("student-02")

	// students := []*entities.Student{student1, student2}

	usrEmail := []*entity.UsrEmail{
		{
			UsrID: database.Text("example-id"),
		},
	}

	hashConfig := mockScryptHash()

	payload1001Rows := "last_name,first_name,last_name_phonetic,first_name_phonetic,email,phone_number,student_email,relationship"
	for i := 0; i < 1001; i++ {
		payload1001Rows += "\nparent last name,parent first name,phonetic name,phonetic name,parent-01@example.com,0981143301,student-01@email.com;student-02@email.com,1;2"
	}

	var builder strings.Builder
	sizeInMB := 1024 * 1024 * 10
	builder.Grow(sizeInMB)
	for i := 0; i < sizeInMB; i++ {
		builder.WriteByte(0)
	}
	payload10MB := []byte(builder.String())

	testCases := []TestCase{
		{
			name: "happy case: no row",
			ctx:  ctx,
			req: &pb.ImportParentsAndAssignToStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,phone_number,student_email,relationship`),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:         "happy case: with username",
			ctx:          ctx,
			expectedResp: nil,
			req: &pb.ImportParentsAndAssignToStudentRequest{
				Payload: []byte(`last_name,first_name,username,last_name_phonetic,first_name_phoneticame,email,phone_number,student_email,relationship
				Parent 02,parent first name,username2,phonetic name,phonetic name,parent-02@example.com,0981143311,student-02@email.com,1
				Parent 03,parent first name,username3,phonetic name,phonetic name,parent-03@example.com,0981143321,student-02@email.com,1
				Parent 04,parent first name,username4,phonetic name,phonetic name,parent-04@example.com,0981143331,student-02@email.com,1
				Parent 05,parent first name,username5,phonetic name,phonetic name,parent-05@example.com,0981143341,student-02@email.com,1
				Parent 06,parent first name,username6,phonetic name,phonetic name,parent-06@example.com,0981143351,student-02@email.com,1
				Parent 07,parent first name,username7,phonetic name,phonetic name,parent-07@example.com,0981143361,student-02@email.com,1
				Parent 08,parent first name,username8,phonetic name,phonetic name,parent-08@example.com,0981143371,student-02@email.com,1`),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				domainStudentRepo.On("GetByEmails", ctx, db, mock.Anything).Times(7).Return([]entity.DomainStudent{&repository.Student{}}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainUserRepo.On("GetByExternalUserIDs", ctx, db, mock.Anything).Once().Return(entity.Users{}, nil)
				domainUserRepo.On("GetByUserNames", ctx, db, []string{"username2", "username3", "username4", "username5", "username6", "username7", "username8"}).Once().Return(entity.Users{}, nil)
				usrEmailRepo.On("CreateMultiple", ctx, db, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleParent).Once().Return(&entity.UserGroupV2{}, nil)
				parentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Times(7).Return(nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Times(7).Return(nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				importUserEventRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return([]*entity.ImportUserEvent{}, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "bad case: duplicated username",
			ctx:  ctx,
			expectedResp: &pb.ImportParentsAndAssignToStudentResponse{
				Errors: []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError{
					{
						RowNumber: 3,
						Error:     "duplicationRow",
						FieldName: userNameParentCSVHeader,
					},
				},
			},
			req: &pb.ImportParentsAndAssignToStudentRequest{
				Payload: []byte(`last_name,first_name,username,last_name_phonetic,first_name_phoneticame,email,phone_number,student_email,relationship
				Parent 02,parent first name,username2,phonetic name,phonetic name,parent-02@example.com,0981143311,student-02@email.com,1
				Parent 03,parent first name,username2,phonetic name,phonetic name,parent-03@example.com,0981143321,student-02@email.com,1
				Parent 04,parent first name,username4,phonetic name,phonetic name,parent-04@example.com,0981143331,student-02@email.com,1
				Parent 05,parent first name,username5,phonetic name,phonetic name,parent-05@example.com,0981143341,student-02@email.com,1
				Parent 06,parent first name,username6,phonetic name,phonetic name,parent-06@example.com,0981143351,student-02@email.com,1
				Parent 07,parent first name,username7,phonetic name,phonetic name,parent-07@example.com,0981143361,student-02@email.com,1
				Parent 08,parent first name,username8,phonetic name,phonetic name,parent-08@example.com,0981143371,student-02@email.com,1`),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				domainStudentRepo.On("GetByEmails", ctx, db, mock.Anything).Times(7).Return([]entity.DomainStudent{&repository.Student{}}, nil)
			},
		},
		{
			name: "bad case: username is already existed",
			ctx:  ctx,
			expectedResp: &pb.ImportParentsAndAssignToStudentResponse{
				Errors: []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError{
					{
						RowNumber: 2,
						Error:     "alreadyRegisteredRow",
						FieldName: userNameParentCSVHeader,
					},
				},
			},
			req: &pb.ImportParentsAndAssignToStudentRequest{
				Payload: []byte(`last_name,first_name,username,last_name_phonetic,first_name_phoneticame,email,phone_number,student_email,relationship
				Parent 01,parent first name,USERNAME1,phonetic name,phonetic name,parent-01@example.com,0981143311,student-01@email.com,1`),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainStudentRepo.On("GetByEmails", ctx, db, mock.Anything).Once().Return([]entity.DomainStudent{&repository.Student{}}, nil)
				domainUserRepo.On("GetByUserNames", ctx, db, []string{"username1"}).Once().Return(entity.Users{&MockDomainUser{UsernameAttr: field.NewString("username1")}}, nil)
				userRepo.On("GetByPhone", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainUserRepo.On("GetByExternalUserIDs", ctx, db, mock.Anything).Once().Return(entity.Users{}, nil)
			},
		},
		{
			name: "happy case: 1 row with last&first name",
			ctx:  ctx,
			req: &pb.ImportParentsAndAssignToStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,username,email,phone_number,student_email,relationship
				parent last name,parent first name,phonetic name,phonetic name,username1,parent1@example.com,0981143301,student-02@email.com,1`),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				domainStudentRepo.On("GetByEmails", ctx, db, mock.Anything).Once().Return([]entity.DomainStudent{&repository.Student{}}, nil)
				domainUserRepo.On("GetByUserNames", ctx, db, []string{"username1"}).Once().Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, db, mock.Anything).Once().Return(nil, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleParent).Once().Return(&entity.UserGroupV2{}, nil)
				// studentRepo.On("FindStudentProfilesByIDs", ctx, mock.Anything, mock.Anything).Return(students, nil)
				usrEmailRepo.On("CreateMultiple", ctx, db, mock.Anything).Once().Return(usrEmail, nil)
				parentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Twice().Return(hashConfig)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				domainUserRepo.On("GetByExternalUserIDs", ctx, db, mock.Anything).Once().Return(entity.Users{}, nil)
				// jsm.On("TracedPublish", mock.Anything, mock.Anything, constants.SubjectImportStudentEvent, mock.Anything).Once().Return(nil, nil)
				importUserEventRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return([]*entity.ImportUserEvent{}, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "happy case: 1 row with parent tag",
			ctx:  ctx,
			req: &pb.ImportParentsAndAssignToStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,username,email,phone_number,student_email,relationship,parent_tag
				parent last name,parent first name,phonetic name,phonetic name,username1,parent1@example.com,0981143301,student-02@email.com,1,partner-id-tag_id1`),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				domainTagRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags(
					[]entity.DomainTag{
						createMockDomainTagWithType("tag_id1", pb.UserTagType_USER_TAG_TYPE_PARENT),
					},
				), nil)
				domainUserRepo.On("GetByUserNames", ctx, db, []string{"username1"}).Once().Return(entity.Users{}, nil)
				domainStudentRepo.On("GetByEmails", ctx, db, mock.Anything).Once().Return([]entity.DomainStudent{&repository.Student{}}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, db, mock.Anything).Once().Return(nil, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleParent).Once().Return(&entity.UserGroupV2{}, nil)
				// studentRepo.On("FindStudentProfilesByIDs", ctx, mock.Anything, mock.Anything).Return(students, nil)
				usrEmailRepo.On("CreateMultiple", ctx, db, mock.Anything).Once().Return(usrEmail, nil)
				parentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Twice().Return(hashConfig)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				domainTaggedUserRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				domainUserRepo.On("GetByExternalUserIDs", ctx, db, mock.Anything).Once().Return(entity.Users{}, nil)
				// jsm.On("TracedPublish", mock.Anything, mock.Anything, constants.SubjectImportStudentEvent, mock.Anything).Once().Return(nil, nil)
				importUserEventRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return([]*entity.ImportUserEvent{}, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "internal err: can not find tag by partner-id",
			ctx:  ctx,
			req: &pb.ImportParentsAndAssignToStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,username,email,phone_number,student_email,relationship,parent_tag
				parent last name,parent first name,phonetic name,phonetic name,username1,parent1@example.com,0981143301,student-02@email.com,1,partner-id-tag_id1`),
			},
			expectedResp: &pb.ImportParentsAndAssignToStudentResponse{
				Errors: []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError{
					{
						RowNumber: 2,
						Error:     "notFollowParentTemplate",
						FieldName: tagParentCSVHeader,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				domainUserRepo.On("GetByUserNames", ctx, db, []string{"username1"}).Once().Return(entity.Users{}, nil)
				domainStudentRepo.On("GetByEmails", ctx, db, mock.Anything).Once().Return([]entity.DomainStudent{&repository.Student{}}, nil)
				domainTagRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags{}, nil)
			},
		},
		{
			name: "internal: upsert tag with error",
			ctx:  ctx,
			req: &pb.ImportParentsAndAssignToStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,username,email,phone_number,student_email,relationship,parent_tag
				parent last name,parent first name,phonetic name,phonetic name,username1,parent1@example.com,0981143301,student-02@email.com,1,partner-id-tag_id1`),
			},
			expectedErr: status.Errorf(
				codes.Internal,
				"otherErrorImport database.ExecInTx: %v",
				fmt.Errorf("UpsertTaggedUsers: DomainTaggedUserRepo.UpsertBatch: %v", fmt.Errorf("error")),
			),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				domainTagRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags(
					[]entity.DomainTag{
						createMockDomainTagWithType("tag_id1", pb.UserTagType_USER_TAG_TYPE_PARENT),
					},
				), nil)

				domainStudentRepo.On("GetByEmails", ctx, db, mock.Anything).Once().Return([]entity.DomainStudent{&repository.Student{}}, nil)
				domainUserRepo.On("GetByUserNames", ctx, db, []string{"username1"}).Once().Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, db, mock.Anything).Once().Return(nil, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleParent).Once().Return(&entity.UserGroupV2{}, nil)
				// studentRepo.On("FindStudentProfilesByIDs", ctx, mock.Anything, mock.Anything).Return(students, nil)
				usrEmailRepo.On("CreateMultiple", ctx, db, mock.Anything).Once().Return(usrEmail, nil)
				parentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				domainTaggedUserRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(fmt.Errorf("error"))
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				domainUserRepo.On("GetByExternalUserIDs", ctx, db, mock.Anything).Once().Return(entity.Users{}, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "alreadyRegisteredRow if parent email exist",
			ctx:  ctx,
			expectedResp: &pb.ImportParentsAndAssignToStudentResponse{
				Errors: []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError{
					{
						RowNumber: 2,
						Error:     "alreadyRegisteredRow",
						FieldName: emailParentCSVHeader,
					},
					{
						RowNumber: 3,
						Error:     "alreadyRegisteredRow",
						FieldName: emailParentCSVHeader,
					},
				},
			},
			req: &pb.ImportParentsAndAssignToStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,username,email,phone_number,student_email,relationship
				parent 01 last name,parent first name,phonetic name,phonetic name,username1,parent1@example.com,0981143301,student-02@email.com,1
		        parent 02 last name,parent first name,phonetic name,phonetic name,username2,parent2@example.com,0981141311,student-02@email.com,1`),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				userRepo.On("GetByPhone", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainUserRepo.On("GetByUserNames", ctx, db, []string{"username1", "username2"}).Once().Return(entity.Users{}, nil)
				domainStudentRepo.On("GetByEmails", ctx, db, mock.Anything).Twice().Return([]entity.DomainStudent{&repository.Student{}}, nil)
				domainUserRepo.On("GetByExternalUserIDs", ctx, db, mock.Anything).Once().Return(entity.Users{}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, db, mock.Anything).Once().Return([]*entity.LegacyUser{
					{
						ID:          database.Text("Parent 01"),
						Email:       database.Text("parent1@example.com"),
						PhoneNumber: database.Text("0981143311"),
					},
					{
						ID:          database.Text("Parent 02"),
						Email:       database.Text("parent2@example.com"),
						PhoneNumber: database.Text("0981141311"),
					},
				}, nil)
			},
		},
		{
			name: "happy case: multiple rows",
			ctx:  ctx,
			req: &pb.ImportParentsAndAssignToStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phoneticame,username,email,phone_number,student_email,relationship
				Parent 02,parent first name,phonetic name,phonetic name,username1,parent-02@example.com,0981143311,student-02@email.com,1
				Parent 03,parent first name,phonetic name,phonetic name,username2,parent-03@example.com,0981143321,student-02@email.com,1
				Parent 04,parent first name,phonetic name,phonetic name,username3,parent-04@example.com,0981143331,student-02@email.com,1
				Parent 05,parent first name,phonetic name,phonetic name,username4,parent-05@example.com,0981143341,student-02@email.com,1
				Parent 06,parent first name,phonetic name,phonetic name,username5,parent-06@example.com,0981143351,student-02@email.com,1
				Parent 07,parent first name,phonetic name,phonetic name,username6,parent-07@example.com,0981143361,student-02@email.com,1
				Parent 08,parent first name,phonetic name,phonetic name,username7,parent-08@example.com,0981143371,student-02@email.com,1`),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				domainUserRepo.On("GetByUserNames", ctx, db, []string{"username1", "username2", "username3", "username4", "username5", "username6", "username7"}).Once().Return(entity.Users{}, nil)
				domainStudentRepo.On("GetByEmails", ctx, db, mock.Anything).Times(7).Return([]entity.DomainStudent{&repository.Student{}}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, db, mock.Anything).Once().Return(nil, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleParent).Once().Return(&entity.UserGroupV2{}, nil)
				usrEmailRepo.On("CreateMultiple", ctx, db, mock.Anything).Once().Return(usrEmail, nil)
				parentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Times(7).Return(nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Times(7).Return(nil)
				importUserEventRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return([]*entity.ImportUserEvent{}, nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Return(hashConfig)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserRepo.On("GetByExternalUserIDs", ctx, db, mock.Anything).Once().Return(entity.Users{}, nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				// jsm.On("TracedPublish", mock.Anything, mock.Anything, constants.SubjectImportStudentEvent, mock.Anything).Once().Return(nil, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "cannot import parent if duplicationRow",
			ctx:  ctx,
			expectedResp: &pb.ImportParentsAndAssignToStudentResponse{
				Errors: []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError{
					{
						RowNumber: 3,
						Error:     "duplicationRow",
					},
					{
						RowNumber: 4,
						Error:     "duplicationRow",
					},
				},
			},
			req: &pb.ImportParentsAndAssignToStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,phone_number,student_email,relationship
				Parent 02,parent first name,phonetic name,phonetic name,parent-02@example.com,0981143311,student-02@email.com,1
				Parent 02,parent first name,phonetic name,phonetic name,parent-02@example.com,0981143311,student-02@email.com,1
				Parent 02,parent first name,phonetic name,phonetic name,parent-02@example.com,0981143311,student-02@email.com,1`),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				domainStudentRepo.On("GetByEmails", ctx, db, mock.Anything).Times(3).Return([]entity.DomainStudent{&repository.Student{}}, nil)
			},
		},
		{
			name: "cannot import parent if duplicationRow (case-insensitive)",
			ctx:  ctx,
			expectedResp: &pb.ImportParentsAndAssignToStudentResponse{
				Errors: []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError{
					{
						RowNumber: 3,
						Error:     "duplicationRow",
					},
					{
						RowNumber: 4,
						Error:     "duplicationRow",
					},
				},
			},
			req: &pb.ImportParentsAndAssignToStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,phone_number,student_email,relationship
				Parent 02,parent first name,phonetic name,phonetic name,parent-02@example.com,0981143311,student-02@email.com,1
				Parent 02,parent first name,phonetic name,phonetic name,parent-02@example.com,0981143311,student-02@email.com,1
				Parent 02,parent first name,phonetic name,phonetic name,PARENT-02@example.com,0981143311,student-02@email.com,1`),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				domainStudentRepo.On("GetByEmails", ctx, db, mock.Anything).Times(3).Return([]entity.DomainStudent{&repository.Student{}}, nil)
			},
		},
		{
			name: "internal error: error when assigning student user group to users",
			ctx:  ctx,
			req: &pb.ImportParentsAndAssignToStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,username,email,phone_number,student_email,relationship
				Parent 02,parent first name,phonetic name,phonetic name,username1,parent-02@example.com,0981143311,student-02@email.com,1
				Parent 03,parent first name,phonetic name,phonetic name,username2,parent-03@example.com,0981143321,student-02@email.com,1
				Parent 04,parent first name,phonetic name,phonetic name,username3,arent-04@example.com,0981143331,student-02@email.com,1
				Parent 05,parent first name,phonetic name,phonetic name,username4,parent-05@example.com,0981143341,student-02@email.com,1
				Parent 06,parent first name,phonetic name,phonetic name,username5,parent-06@example.com,0981143351,student-02@email.com,1
				Parent 07,parent first name,phonetic name,phonetic name,username6,parent-07@example.com,0981143361,student-02@email.com,1
				Parent 08,parent first name,phonetic name,phonetic name,username7,parent-08@example.com,0981143371,student-02@email.com,1`),
			},
			expectedErr: status.Errorf(
				codes.Internal,
				"otherErrorImport database.ExecInTx: %v",
				fmt.Errorf("s.UserGroupsMemberRepo.AssignWithUserGroup: %v", fmt.Errorf("error")),
			),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				domainUserRepo.On("GetByUserNames", ctx, db, []string{"username1", "username2", "username3", "username4", "username5", "username6", "username7"}).Once().Return(entity.Users{}, nil)
				domainStudentRepo.On("GetByEmails", ctx, db, mock.Anything).Times(7).Return([]entity.DomainStudent{&repository.Student{}}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, db, mock.Anything).Once().Return(nil, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleParent).Once().Return(&entity.UserGroupV2{}, nil)
				usrEmailRepo.On("CreateMultiple", ctx, db, mock.Anything).Once().Return(usrEmail, nil)
				parentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Times(7).Return(nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Times(7).Return(nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
				domainUserRepo.On("GetByExternalUserIDs", ctx, db, mock.Anything).Once().Return(entity.Users{}, nil)

				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "invalidMaxSizeFile",
			ctx:         ctx,
			expectedErr: status.Error(codes.InvalidArgument, "invalidMaxSizeFile"),
			req: &pb.ImportParentsAndAssignToStudentRequest{
				Payload: payload10MB,
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "emptyFile",
			ctx:         ctx,
			expectedErr: status.Error(codes.InvalidArgument, "emptyFile"),
			req: &pb.ImportParentsAndAssignToStudentRequest{
				Payload: []byte(``),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalidNumberRow",
			ctx:         ctx,
			expectedErr: status.Error(codes.InvalidArgument, "invalidNumberRow"),
			req: &pb.ImportParentsAndAssignToStudentRequest{
				Payload: []byte(payload1001Rows),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "missingMandatory",
			ctx:  ctx,
			expectedResp: &pb.ImportParentsAndAssignToStudentResponse{
				Errors: []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError{
					{
						RowNumber: 2,
						Error:     "missingMandatory",
						FieldName: nameParentCSVHeader,
					},
					{
						RowNumber: 3,
						Error:     "missingMandatory",
						FieldName: emailParentCSVHeader,
					},
				},
			},
			req: &pb.ImportParentsAndAssignToStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,phone_number,student_email,relationship
				,,,,,0981143301,,
				parent-01,,,,,0981143311,,`),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
			},
		},
		{
			name: "todo: missing username",
			ctx:  ctx,
			expectedResp: &pb.ImportParentsAndAssignToStudentResponse{
				Errors: []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError{
					{
						RowNumber: 2,
						Error:     "missingMandatory",
						FieldName: userNameParentCSVHeader,
					},
				},
			},
			req: &pb.ImportParentsAndAssignToStudentRequest{
				Payload: []byte(`last_name,first_name,username,last_name_phonetic,first_name_phoneticame,email,phone_number,student_email,relationship
				Parent 02,parent first name,,phonetic name,phonetic name,parent-02@example.com,0981143311,student-02@email.com,1
				Parent 03,parent first name,username,phonetic name,phonetic name,parent-03@example.com,0981143321,student-02@email.com,1
				Parent 04,parent first name,username,phonetic name,phonetic name,parent-04@example.com,0981143331,student-02@email.com,1
				Parent 05,parent first name,username,phonetic name,phonetic name,parent-05@example.com,0981143341,student-02@email.com,1
				Parent 06,parent first name,username,phonetic name,phonetic name,parent-06@example.com,0981143351,student-02@email.com,1
				Parent 07,parent first name,username,phonetic name,phonetic name,parent-07@example.com,0981143361,student-02@email.com,1
				Parent 08,parent first name,username,phonetic name,phonetic name,parent-08@example.com,0981143371,student-02@email.com,1`),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				domainStudentRepo.On("GetByEmails", ctx, db, mock.Anything).Times(7).Return([]entity.DomainStudent{&repository.Student{}}, nil)
			},
		},
		{
			name: "missingMandatory last&first name",
			ctx:  ctx,
			expectedResp: &pb.ImportParentsAndAssignToStudentResponse{
				Errors: []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError{
					{
						RowNumber: 2,
						Error:     "missingMandatory",
						FieldName: firstNameParentCSVHeader,
					},
					{
						RowNumber: 3,
						Error:     "missingMandatory",
						FieldName: lastNameParentCSVHeader,
					},
				},
			},
			req: &pb.ImportParentsAndAssignToStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,phone_number,student_email,relationship
				parent last name,,,,parent-01@example.com,0981143311,,
				,parent first name,,,parent-02@example.com,0981143311,,`),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
			},
		},
		{
			name: "notMatchRelationshipAndEmailStudent unleash on",
			ctx:  ctx,
			expectedResp: &pb.ImportParentsAndAssignToStudentResponse{
				Errors: []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError{
					{
						RowNumber: 2,
						Error:     "notMatchRelationshipAndEmailStudent",
						FieldName: studentEmailsParentCSVHeader,
					},
					{
						RowNumber: 3,
						Error:     "notMatchRelationshipAndEmailStudent",
						FieldName: studentEmailsParentCSVHeader,
					},
				},
			},
			req: &pb.ImportParentsAndAssignToStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,phone_number,student_email,relationship
				Parent 02,parent first name,phonetic name,phonetic name,parent-05@example.com,0981143311,student-05@email.com,1;2
				Parent 03,parent first name,phonetic name,phonetic name,parent-06@example.com,0981143311,student-06@email.com;student-04@email.com,1;1;1`),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
			},
		},
		{
			name: "notFollowParentTemplate",
			ctx:  ctx,
			expectedResp: &pb.ImportParentsAndAssignToStudentResponse{
				Errors: []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError{
					{
						RowNumber: 2,
						Error:     "notFollowParentTemplate",
						FieldName: emailParentCSVHeader,
					},
					{
						RowNumber: 3,
						Error:     "notMatchRelationshipAndEmailStudent",
						FieldName: relationshipsParentCSVHeader,
					},
					{
						RowNumber: 4,
						Error:     "notFollowParentTemplate",
						FieldName: relationshipsParentCSVHeader,
					},
					{
						RowNumber: 5,
						Error:     "notFollowParentTemplate",
						FieldName: relationshipsParentCSVHeader,
					},
					{
						RowNumber: 6,
						Error:     "missingMandatory",
						FieldName: emailParentCSVHeader,
					},
					{
						RowNumber: 7,
						Error:     "notMatchRelationshipAndEmailStudent",
						FieldName: relationshipsParentCSVHeader,
					},
					{
						RowNumber: 8,
						Error:     "notFollowParentTemplate",
						FieldName: emailParentCSVHeader,
					},
					{
						RowNumber: 9,
						Error:     "notFollowParentTemplate",
						FieldName: studentEmailsParentCSVHeader,
					},
					{
						RowNumber: 10,
						Error:     "notFollowParentTemplate",
						FieldName: relationshipsParentCSVHeader,
					},
					{
						RowNumber: 11,
						Error:     "notFollowParentTemplate",
						FieldName: relationshipsParentCSVHeader,
					},
					{
						RowNumber: 12,
						Error:     "notFollowParentTemplate",
						FieldName: phoneNumberParentCSVHeader,
					},
					{
						RowNumber: 13,
						Error:     "notFollowParentTemplate",
						FieldName: studentEmailsParentCSVHeader,
					},
					{
						RowNumber: 14,
						Error:     "notFollowParentTemplate",
						FieldName: emailParentCSVHeader,
					},
					{
						RowNumber: 15,
						Error:     "notFollowParentTemplate",
						FieldName: relationshipsParentCSVHeader,
					},
				},
			},
			req: &pb.ImportParentsAndAssignToStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,phone_number,student_email,relationship
				Parent 01,parent first name,phonetic name,phonetic name,parent-01@example..com,0981143311,student-01@email.com;student-02@email.com,
				Parent 02,parent first name,phonetic name,phonetic name,parent-02@example.com,invalid-phone,student-02@email.com;student-01@email.com,father
				Parent 03,parent first name,phonetic name,phonetic name,parent-03example.com,0981143311,student-03@email.com;,1;-2
				Parent 04,parent first name,phonetic name,phonetic name,parent-04example.com,0981143311,student-04@email.com;student-07@email.com,1;mother
				Parent 05,parent first name,phonetic name,phonetic name,,-1,student-05@email.com;student-02@email.com,1;2
				Parent 06,parent first name,phonetic name,phonetic name,parent-06@example.com,0981143311,student-06@email.com;student-04@email.com,1;1;1
				Parent 07,parent first name,phonetic name,phonetic name,parent-07@example..com,0981143311,student-07@email.com;student-05@email.com,81;2
				Parent 08,parent first name,phonetic name,phonetic name,parent-08@example.com,0981143311,student-08@email.com;student-07@email..com,1;1
				Parent 09,parent first name,phonetic name,phonetic name,parent-09@example.com,0981143311,student-09@email.com;student-01@email.com,-1;2
				Parent 10,parent first name,phonetic name,phonetic name,parent-10@example.com,0981143311,student-10@email.com;student-02@email.com,1;-2
				Parent 11,parent first name,phonetic name,phonetic name,parent-11@example.com,invalid-phoneNumber,student-10@email.com;student-02@email.com,1;2
				Parent 13,parent first name,phonetic name,phonetic name,parent-13@example.com,0981143311,invalid-studentEmail;student-01@email.com,1;-12
				Parent 15,parent first name,phonetic name,phonetic name,invalid-parentEmail,0981143311,student-03@email.com;student-04@email.com,1;-2
				Parent 16,parent first name,phonetic name,phonetic name,parent-16@example.com,0981143311,student-03@email.com;student-04@email.com,1;0`),
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				domainStudentRepo.On("GetByEmails", ctx, db, mock.Anything).Return(nil, nil)
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

			testCase.setup(testCase.ctx)

			resp, err := userModifierService.ImportParentsAndAssignToStudent(testCase.ctx, testCase.req.(*pb.ImportParentsAndAssignToStudentRequest))
			if err != nil {
				fmt.Println(err)
			}
			if resp != nil && len(resp.Errors) > 0 {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.expectedResp.(*pb.ImportParentsAndAssignToStudentResponse)
				assert.Equal(t, len(expectedResp.Errors), len(resp.Errors))
				for i, err := range resp.Errors {
					assert.Equal(t, expectedResp.Errors[i].RowNumber, err.RowNumber)
					assert.Contains(t, expectedResp.Errors[i].Error, err.Error)
					// assert.Equal(t, expectedResp.Errors[i].FieldName, err.FieldName)
				}
			}

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})

		mock.AssertExpectationsForObjects(t, db, tx, orgRepo, usrEmailRepo, userRepo, studentRepo, userGroupRepo, userAccessPathRepo, studentParentRepo, jsm, importUserEventRepo,
			firebaseAuth, tenantManager, userGroupsMemberRepo, userGroupV2Repo, parentRepo, userPhoneNumberRepo, domainTagRepo, domainTaggedUserRepo, mockUnleashClient)
	}
}

func TestUserModifierService_getAndValidUserPhoneNumberCSV(t *testing.T) {
	t.Parallel()
	_, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	type newParams struct {
		*ImportParentCSVField
		ParentCSV
	}

	testCases := []TestCase{
		{
			name: "empty params",
			req: newParams{
				&ImportParentCSVField{Name: &CsvField{Text: "testing name", Exist: true}},
				ParentCSV{Parent: entity.Parent{ID: database.Text("")}},
			},
			expectedResp: ParentCSV{},
		},
		{
			name: "happy case with primary phone number",
			req: newParams{
				&ImportParentCSVField{
					Name:               &CsvField{Text: "testing name", Exist: true},
					PrimaryPhoneNumber: &CsvField{Text: "0123456789", Exist: true},
				},
				ParentCSV{
					Parent: entity.Parent{
						ID:         pgtype.Text{String: "123", Status: pgtype.Present},
						LegacyUser: entity.LegacyUser{ResourcePath: pgtype.Text{String: "resourcePath - 123", Status: pgtype.Present}},
					},
				},
			},
			expectedResp: ParentCSV{
				UserPhoneNumbers: entity.UserPhoneNumbers{
					{
						UserID:          pgtype.Text{String: "123", Status: pgtype.Present},
						PhoneNumber:     pgtype.Text{String: "0123456789", Status: pgtype.Present},
						PhoneNumberType: pgtype.Text{String: pb.ParentPhoneNumber_PARENT_PRIMARY_PHONE_NUMBER.String(), Status: pgtype.Present},
						ResourcePath:    pgtype.Text{String: "resourcePath - 123", Status: pgtype.Present},
					},
				},
			},
		},
		{
			name: "happy case with secondary phone number",
			req: newParams{
				&ImportParentCSVField{
					Name:                 &CsvField{Text: "testing name", Exist: true},
					SecondaryPhoneNumber: &CsvField{Text: "0123456789", Exist: true},
				},
				ParentCSV{
					Parent: entity.Parent{
						ID:         pgtype.Text{String: "123", Status: pgtype.Present},
						LegacyUser: entity.LegacyUser{ResourcePath: pgtype.Text{String: "resourcePath - 123", Status: pgtype.Present}},
					},
				},
			},
			expectedResp: ParentCSV{
				UserPhoneNumbers: entity.UserPhoneNumbers{
					{
						UserID:          pgtype.Text{String: "123", Status: pgtype.Present},
						PhoneNumber:     pgtype.Text{String: "0123456789", Status: pgtype.Present},
						PhoneNumberType: pgtype.Text{String: pb.ParentPhoneNumber_PARENT_SECONDARY_PHONE_NUMBER.String(), Status: pgtype.Present},
						ResourcePath:    pgtype.Text{String: "resourcePath - 123", Status: pgtype.Present},
					},
				},
			},
		},
		{
			name: "fail case with primary phone number with invalid number",
			req: newParams{
				&ImportParentCSVField{
					Name:               &CsvField{Text: "testing name", Exist: true},
					PrimaryPhoneNumber: &CsvField{Text: "123", Exist: true},
				},
				ParentCSV{
					Parent: entity.Parent{
						ID:         pgtype.Text{String: "123", Status: pgtype.Present},
						LegacyUser: entity.LegacyUser{ResourcePath: pgtype.Text{String: "resourcePath - 123", Status: pgtype.Present}},
					},
				},
			},
			expectedErr: nil,
			expectedResp: &pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError{
				Error:     "notFollowParentTemplate",
				FieldName: primaryPhoneNumberParentCSVHeader,
				RowNumber: 3,
			},
		},
		{
			name: "fail case with secondary phone number with invalid number",
			req: newParams{
				&ImportParentCSVField{
					Name:                 &CsvField{Text: "testing name", Exist: true},
					SecondaryPhoneNumber: &CsvField{Text: "123", Exist: true},
				},
				ParentCSV{
					Parent: entity.Parent{
						ID:         pgtype.Text{String: "123", Status: pgtype.Present},
						LegacyUser: entity.LegacyUser{ResourcePath: pgtype.Text{String: "resourcePath - 123", Status: pgtype.Present}},
					},
				},
			},
			expectedResp: &pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError{
				Error:     "notFollowParentTemplate",
				FieldName: secondaryPhoneNumberParentCSVHeader,
				RowNumber: 3,
			},
		},
		{
			name: "fail case with primary phone number and secondary phone number same",
			req: newParams{
				&ImportParentCSVField{
					Name:                 &CsvField{Text: "testing name", Exist: true},
					PrimaryPhoneNumber:   &CsvField{Text: "123456789", Exist: true},
					SecondaryPhoneNumber: &CsvField{Text: "123456789", Exist: true},
				},
				ParentCSV{
					Parent: entity.Parent{
						ID:         pgtype.Text{String: "123456789", Status: pgtype.Present},
						LegacyUser: entity.LegacyUser{ResourcePath: pgtype.Text{String: "resourcePath - 123", Status: pgtype.Present}},
					},
				},
			},
			expectedResp: &pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError{
				Error:     "duplicationRow",
				FieldName: primaryPhoneNumberParentCSVHeader,
				RowNumber: 3,
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			result, errCSV, err := getAndValidUserPhoneNumberCSV(testCase.req.(newParams).ImportParentCSVField, testCase.req.(newParams).ParentCSV, 1)

			switch {
			case err != nil:
				assert.Equal(t, testCase.expectedErr, err.Error())
			case errCSV != nil:
				expectedErr := testCase.expectedResp.(*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError)
				assert.Equal(t, expectedErr.Error, errCSV.Error)
				assert.Equal(t, expectedErr.FieldName, errCSV.FieldName)
				assert.Equal(t, expectedErr.RowNumber, errCSV.RowNumber)
			default:
				assert.NotNil(t, result)
				expectedResp := testCase.expectedResp.(ParentCSV)
				for i, resultPhone := range result.UserPhoneNumbers {
					assert.Equal(t, expectedResp.UserPhoneNumbers[i].UserID, resultPhone.UserID)
					assert.Equal(t, expectedResp.UserPhoneNumbers[i].PhoneNumber, resultPhone.PhoneNumber)
					assert.Equal(t, expectedResp.UserPhoneNumbers[i].PhoneNumberType, resultPhone.PhoneNumberType)
					assert.Equal(t, expectedResp.UserPhoneNumbers[i].ResourcePath, resultPhone.ResourcePath)
				}
			}

		})

	}
}

func TestUserModifierService_checkDuplicateData(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// db := new(mock_database.Ext)
	// tx := new(mock_database.Tx)

	userRepo := new(mock_repositories.MockUserRepo)
	domainUserRepo := new(mock_repositories.MockDomainUserRepo)

	s := &UserModifierService{
		UserRepo:       userRepo,
		DomainUserRepo: domainUserRepo,
	}

	testCases := []TestCase{
		{
			name:         "empty params",
			ctx:          ctx,
			req:          entity.Parents{},
			expectedResp: []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError(nil),
			setup: func(ctx context.Context) {
				userRepo.On("GetByEmailInsensitiveCase", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				domainUserRepo.On("GetByExternalUserIDs", ctx, mock.Anything, mock.Anything).Once().Return(entity.Users{}, nil)
			},
		},
		{
			name: "happy case",
			ctx:  ctx,
			req: entity.Parents{
				{
					LegacyUser: entity.LegacyUser{
						Email: pgtype.Text{String: "abc@gmail.com", Status: pgtype.Present},
					},
				},
			},
			expectedResp: []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError(nil),
			setup: func(ctx context.Context) {
				userRepo.On("GetByEmailInsensitiveCase", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				domainUserRepo.On("GetByExternalUserIDs", ctx, mock.Anything, mock.Anything).Once().Return(entity.Users{}, nil)
			},
		},
		{
			name: "fail case: email already have in database ",
			ctx:  ctx,
			req: entity.Parents{
				{
					LegacyUser: entity.LegacyUser{
						Email: pgtype.Text{String: "abc@gmail.com", Status: pgtype.Present},
					},
				},
			},

			expectedResp: []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError{
				{
					Error:     "alreadyRegisteredRow",
					FieldName: emailParentCSVHeader,
					RowNumber: 2,
				},
			},
			setup: func(ctx context.Context) {
				userRepo.On("GetByEmailInsensitiveCase", ctx, mock.Anything, mock.Anything).Once().Return([]*entity.LegacyUser{
					{
						Email: database.Text("abc@gmail.com"),
					},
				}, nil)
				userRepo.On("GetByPhone", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				domainUserRepo.On("GetByExternalUserIDs", ctx, mock.Anything, mock.Anything).Once().Return(entity.Users{}, nil)
			},
		},
		{
			name: "fail case: email already have in database (case-insensitive)",
			ctx:  ctx,
			req: entity.Parents{
				{
					LegacyUser: entity.LegacyUser{
						Email: pgtype.Text{String: "abc@gmail.com", Status: pgtype.Present},
					},
				},
			},

			expectedResp: []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError{
				{
					Error:     "alreadyRegisteredRow",
					FieldName: emailParentCSVHeader,
					RowNumber: 2,
				},
			},
			setup: func(ctx context.Context) {
				userRepo.On("GetByEmailInsensitiveCase", ctx, mock.Anything, mock.Anything).Once().Return([]*entity.LegacyUser{
					{
						Email: database.Text("ABC@gmail.com"),
					},
				}, nil)
				userRepo.On("GetByPhone", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				domainUserRepo.On("GetByExternalUserIDs", ctx, mock.Anything, mock.Anything).Once().Return(entity.Users{}, nil)
			},
		},
		{
			name: "fail case: s.UserRepo.GetByEmail fail",
			ctx:  ctx,
			req: entity.Parents{
				{
					LegacyUser: entity.LegacyUser{
						Email: pgtype.Text{String: "abc@gmail.com", Status: pgtype.Present},
					},
				},
			},

			expectedResp: "rpc error: code = Internal desc = s.UserRepo.GetByEmailInsensitiveCase: err this case",
			setup: func(ctx context.Context) {
				userRepo.On("GetByEmailInsensitiveCase", ctx, mock.Anything, mock.Anything).Once().Return(nil, errors.New("err this case"))
			},
		},
		{
			name: "fail case: s.UserRepo.GetByPhone fail",
			ctx:  ctx,
			req: entity.Parents{
				{
					LegacyUser: entity.LegacyUser{
						Email: pgtype.Text{String: "abc@gmail.com", Status: pgtype.Present},
					},
				},
			},

			expectedResp: "rpc error: code = Internal desc = s.UserRepo.GetByPhone: err this case",
			setup: func(ctx context.Context) {
				userRepo.On("GetByEmailInsensitiveCase", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, mock.Anything, mock.Anything).Once().Return(nil, errors.New("err this case"))
			},
		},
		{
			name: "fail case: external user id alreadyRegisteredRow",
			ctx:  ctx,
			req: entity.Parents{
				{
					LegacyUser: entity.LegacyUser{
						Email:          pgtype.Text{String: "abc@gmail.com", Status: pgtype.Present},
						ExternalUserID: pgtype.Text{String: "external_user_id", Status: pgtype.Present},
					},
				},
			},

			expectedResp: []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError{
				{
					Error:     "alreadyRegisteredRow",
					FieldName: externalUserIDParentCSVHeader,
					RowNumber: 2,
				},
			},
			setup: func(ctx context.Context) {
				userRepo.On("GetByEmailInsensitiveCase", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				domainUserRepo.On("GetByExternalUserIDs", ctx, mock.Anything, mock.Anything).Once().Return(entity.Users{entity_mock.User{
					RandomUser: entity_mock.RandomUser{
						ExternalUserID: field.NewString("external_user_id"),
					},
				}}, nil)
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

			testCase.setup(testCase.ctx)

			errCSV, err := s.checkDuplicateData(testCase.ctx, testCase.req.(entity.Parents))
			if err != nil {
				assert.Equal(t, testCase.expectedResp, err.Error())
				return
			}

			if errCSV != nil {
				for i, actualErr := range errCSV {

					expectedErr := testCase.expectedResp.([]*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError)[i]
					assert.Equal(t, expectedErr.Error, actualErr.Error)
					assert.Equal(t, expectedErr.FieldName, actualErr.FieldName)
					assert.Equal(t, expectedErr.RowNumber, actualErr.RowNumber)
				}

			} else {
				assert.Equal(t, testCase.expectedResp.([]*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError), errCSV)
			}

		})

	}
}

func TestUserModifierService_checkDuplicateRow(t *testing.T) {
	t.Parallel()
	_, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	s := &UserModifierService{}

	testCases := []TestCase{
		{
			name:         "empty params",
			req:          entity.Parents{},
			expectedResp: []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError(nil),
		},
		{
			name: "happy case",
			req: entity.Parents{
				{LegacyUser: entity.LegacyUser{Email: pgtype.Text{String: "abc@gmail.com", Status: pgtype.Present}}},
			},
			expectedResp: []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError(nil),
		},
		{
			name: "fail case: duplicate email in the array Parents",
			req: entity.Parents{
				{LegacyUser: entity.LegacyUser{Email: pgtype.Text{String: "abc@gmail.com", Status: pgtype.Present}}},
				{LegacyUser: entity.LegacyUser{Email: pgtype.Text{String: "abc@gmail.com", Status: pgtype.Present}}},
			},

			expectedResp: []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError{
				{
					Error:     "duplicationRow",
					FieldName: emailParentCSVHeader,
					RowNumber: 3,
				},
			},
		},
		{
			name: "fail case: duplicate external user id in the array Parents",
			req: entity.Parents{
				{LegacyUser: entity.LegacyUser{
					Email:          pgtype.Text{String: "abc@gmail.com", Status: pgtype.Present},
					ExternalUserID: pgtype.Text{String: "external-user-id", Status: pgtype.Present}}},
				{LegacyUser: entity.LegacyUser{
					Email:          pgtype.Text{String: "123@gmail.com", Status: pgtype.Present},
					ExternalUserID: pgtype.Text{String: "external-user-id", Status: pgtype.Present}}},
			},

			expectedResp: []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError{
				{
					Error:     "duplicationRow",
					FieldName: externalUserIDParentCSVHeader,
					RowNumber: 3,
				},
			},
		},
		{
			name: "fail case: duplicate email in the array Parents (case-insensitive)",
			req: entity.Parents{
				{LegacyUser: entity.LegacyUser{Email: pgtype.Text{String: "abc@gmail.com", Status: pgtype.Present}}},
				{LegacyUser: entity.LegacyUser{Email: pgtype.Text{String: "ABC@gmail.com", Status: pgtype.Present}}},
			},

			expectedResp: []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError{
				{
					Error:     "duplicationRow",
					FieldName: emailParentCSVHeader,
					RowNumber: 3,
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			errCSV := s.checkDuplicateRow(testCase.req.(entity.Parents))
			if errCSV != nil {
				for i, actualErr := range errCSV {
					expectedErr := testCase.expectedResp.([]*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError)[i]
					assert.Equal(t, expectedErr.Error, actualErr.Error)
					assert.Equal(t, expectedErr.FieldName, actualErr.FieldName)
					assert.Equal(t, expectedErr.RowNumber, actualErr.RowNumber)
				}
			} else {
				assert.Equal(t, testCase.expectedResp.([]*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError), errCSV)
			}

		})

	}
}

func Test_convertToImportError(t *testing.T) {
	tests := []struct {
		name     string
		inputErr error
		expected error
	}{
		{
			name:     "MissingMandatoryFieldError",
			inputErr: entity.MissingMandatoryFieldError{},
			expected: errMissingMandatory,
		},
		{
			name:     "InvalidFieldError",
			inputErr: entity.InvalidFieldError{},
			expected: errNotFollowParentTemplate,
		},
		{
			name:     "DuplicatedFieldError",
			inputErr: entity.DuplicatedFieldError{},
			expected: errDuplicationRow,
		},
		{
			name:     "ExistingDataError",
			inputErr: entity.ExistingDataError{},
			expected: errAlreadyRegisteredRow,
		},
		{
			name:     "OtherError",
			inputErr: errors.New("some error"),
			expected: errors.New("some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToImportError(tt.inputErr)
			assert.Equal(t, result, tt.expected)
		})
	}
}

func Test_generatedParents(t *testing.T) {
	type args struct {
		parentCSVs                     []*ParentCSV
		isUserNameStudentParentEnabled bool
	}
	tests := []struct {
		name             string
		args             args
		parents          entity.Parents
		users            entity.Users
		userPhoneNumbers entity.UserPhoneNumbers
		tags             map[entity.User][]entity.DomainTag
	}{
		{
			name: "map parent csv to parent entity with enable username",
			args: args{
				isUserNameStudentParentEnabled: true,
				parentCSVs: []*ParentCSV{
					{
						Parent: entity.Parent{
							LegacyUser: entity.LegacyUser{
								ID:             database.Text("ID"),
								ExternalUserID: database.Text("ExternalUserID"),
								Email:          database.Text("Email"),
								UserName:       database.Text("UserName"),
								FirstName:      database.Text("FirstName"),
								LastName:       database.Text("LastName"),
							},
						},
					},
				},
			},
			parents: entity.Parents{&entity.Parent{
				LegacyUser: entity.LegacyUser{
					ID:             database.Text("ID"),
					ExternalUserID: database.Text("ExternalUserID"),
					Email:          database.Text("Email"),
					UserName:       database.Text("UserName"),
					FirstName:      database.Text("FirstName"),
					LastName:       database.Text("LastName"),
				},
			}},
			users: entity.Users{&grpc.UserProfile{
				Profile: &pb.UserProfile{
					UserId:         "ID",
					ExternalUserId: "ExternalUserID",
					Email:          "Email",
					Username:       "UserName",
					FirstName:      "FirstName",
					LastName:       "LastName",
				},
			}},
		},
		{
			name: "map parent csv to parent entity with disable username",
			args: args{
				isUserNameStudentParentEnabled: false,
				parentCSVs: []*ParentCSV{
					{
						Parent: entity.Parent{
							LegacyUser: entity.LegacyUser{
								ID:             database.Text("ID"),
								ExternalUserID: database.Text("ExternalUserID"),
								Email:          database.Text("Email"),
								FirstName:      database.Text("FirstName"),
								LastName:       database.Text("LastName"),
							},
						},
					},
				},
			},
			parents: entity.Parents{&entity.Parent{
				LegacyUser: entity.LegacyUser{
					ID:             database.Text("ID"),
					ExternalUserID: database.Text("ExternalUserID"),
					Email:          database.Text("Email"),
					UserName:       database.Text("Email"),
					FirstName:      database.Text("FirstName"),
					LastName:       database.Text("LastName"),
				},
			}},
			users: entity.Users{&grpc.UserProfile{
				Profile: &pb.UserProfile{
					UserId:         "ID",
					ExternalUserId: "ExternalUserID",
					Email:          "Email",
					Username:       "Email",
					FirstName:      "FirstName",
					LastName:       "LastName",
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parents, users, _, _ := generatedParents(tt.args.parentCSVs, tt.args.isUserNameStudentParentEnabled)
			assert.Equalf(t, tt.parents, parents, "generatedParents(%v, %v)", tt.args.parentCSVs, tt.args.isUserNameStudentParentEnabled)
			assert.Equalf(t, tt.users, users, "generatedParents(%v, %v)", tt.args.parentCSVs, tt.args.isUserNameStudentParentEnabled)
		})
	}
}
