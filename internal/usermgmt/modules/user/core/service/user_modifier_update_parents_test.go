package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	internal_auth_user "github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	mock_usermgmt "github.com/manabie-com/backend/internal/usermgmt/pkg/mock"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	mock_multitenant "github.com/manabie-com/backend/mock/golibs/auth/multitenant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_firebase "github.com/manabie-com/backend/mock/golibs/firebase"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUpdateParentsAndFamilyRelationship(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	tx := new(mock_database.Tx)
	userRepo := new(mock_repositories.MockUserRepo)
	usrEmailRepo := new(mock_repositories.MockUsrEmailRepo)
	studentRepo := new(mock_repositories.MockStudentRepo)
	parentRepo := new(mock_repositories.MockParentRepo)
	studentParentRepo := new(mock_repositories.MockStudentParentRepo)
	userGroupRepo := new(mock_repositories.MockUserGroupRepo)
	userPhoneNumberRepo := new(mock_repositories.MockUserPhoneNumberRepo)
	userAccessPathRepo := new(mock_repositories.MockUserAccessPathRepo)
	orgRepo := new(mock_repositories.OrganizationRepo)
	jsm := new(mock_nats.JetStreamManagement)
	firebaseAuth := new(mock_firebase.AuthClient)
	tenantManager := new(mock_multitenant.TenantManager)
	firebaseAuthClient := new(mock_multitenant.TenantClient)
	taggedUserRepo := new(mock_repositories.MockDomainTaggedUserRepo)
	tagRepo := new(mock_repositories.MockDomainTagRepo)
	domainUserRepo := new(mock_repositories.MockDomainUserRepo)
	unleashClient := new(mock_unleash_client.UnleashClientInstance)
	internalConfigurationRepo := new(mock_repositories.MockDomainInternalConfigurationRepo)
	id := idutil.ULIDNow()

	s := UserModifierService{
		DB:                        db,
		UserAccessPathRepo:        userAccessPathRepo,
		UserRepo:                  userRepo,
		UsrEmailRepo:              usrEmailRepo,
		UserPhoneNumberRepo:       userPhoneNumberRepo,
		StudentRepo:               studentRepo,
		ParentRepo:                parentRepo,
		StudentParentRepo:         studentParentRepo,
		UserGroupRepo:             userGroupRepo,
		OrganizationRepo:          orgRepo,
		FirebaseClient:            firebaseAuth,
		TenantManager:             tenantManager,
		FirebaseAuthClient:        firebaseAuthClient,
		DomainTaggedUserRepo:      taggedUserRepo,
		DomainTagRepo:             tagRepo,
		JSM:                       jsm,
		UnleashClient:             unleashClient,
		DomainUserRepo:            domainUserRepo,
		InternalConfigurationRepo: internalConfigurationRepo,
	}

	existingParentUser := &entity.LegacyUser{
		ID:             database.Text("id"),
		ExternalUserID: database.Text("external-user-id"),
		Email:          database.Text("existing-parent-email@example.com"),
		PhoneNumber:    database.Text("existing-parent-phone-number"),
	}
	existingParent := &entity.Parent{
		ID:         existingParentUser.ID,
		SchoolID:   database.Int4(1),
		LegacyUser: *existingParentUser,
		ParentAdditionalInfo: &entity.ParentAdditionalInfo{
			Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER.String(),
		},
	}
	existingParentOtherSchoolID := &entity.Parent{
		ID:         existingParentUser.ID,
		SchoolID:   database.Int4(2),
		LegacyUser: *existingParentUser,
		ParentAdditionalInfo: &entity.ParentAdditionalInfo{
			Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER.String(),
		},
	}
	existingParentUser2 := &entity.LegacyUser{
		ID:          database.Text(idutil.ULIDNow()),
		Email:       database.Text("existing-parent-email-2@example.com"),
		PhoneNumber: database.Text("existing-parent-phone-number-2"),
	}
	existingStudentUser := &entity.LegacyUser{
		ID:          database.Text(idutil.ULIDNow()),
		Email:       database.Text("existing-student-email@example.com"),
		PhoneNumber: database.Text("existing-student-phone-number"),
	}
	existingStudent := &entity.LegacyStudent{
		ID:         existingStudentUser.ID,
		SchoolID:   database.Int4(1),
		LegacyUser: *existingStudentUser,
	}

	testCases := []TestCase{
		{
			name: "cannot re-update external user id",
			ctx:  ctx,
			req: &pb.UpdateParentsAndFamilyRelationshipRequest{
				SchoolId:  1,
				StudentId: "some-student-id",
				ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					{
						Id:             "id",
						ExternalUserId: "some-external-user-id",
						Username:       "username",
						UserNameFields: &pb.UserNameFields{
							FirstName: "first-name",
							LastName:  "last-name",
						},
						Email:        "parent-email@example.com",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "cannot re-update external_user_id: id"),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				tagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{"some-student-id"})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				parentRepo.On("GetByIds", ctx, tx, database.TextArray([]string{"id"})).Once().Return(entity.Parents{existingParent}, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				domainUserRepo.On("GetByUserNames", ctx, db, []string{"username"}).Once().Return(entity.Users{}, nil)
			},
		},
		{
			name: "cannot update if external user id is existed",
			ctx:  ctx,
			req: &pb.UpdateParentsAndFamilyRelationshipRequest{
				SchoolId:  1,
				StudentId: "some-student-id",
				ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					{
						Id:             "id",
						ExternalUserId: "some-external-user-id",
						Email:          "parent-email@example.com",
						Username:       "username",
						UserNameFields: &pb.UserNameFields{
							FirstName: "first-name",
							LastName:  "last-name",
						},
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "cannot re-update external_user_id: id"),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				tagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{"some-student-id"})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				parentRepo.On("GetByIds", ctx, tx, database.TextArray([]string{"id"})).Once().Return(entity.Parents{existingParent}, nil)
				domainUserRepo.On("GetByExternalUserIDs", ctx, tx, mock.Anything).Once().Return(entity.Users{
					&repository.User{
						ExternalUserIDAttr: field.NewString("random-external"),
					},
				}, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				domainUserRepo.On("GetByUserNames", ctx, db, []string{"username"}).Once().Return(entity.Users{}, nil)
			},
		},
		{
			name: "cannot update if student ID to assign is empty",
			ctx:  ctx,
			req: &pb.UpdateParentsAndFamilyRelationshipRequest{
				SchoolId:  1,
				StudentId: "",
				ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					{
						Id:       "id",
						Email:    "parent-email@example.com",
						Username: "username",
						UserNameFields: &pb.UserNameFields{
							FirstName: "first-name",
							LastName:  "last-name",
						},
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "student ID cannot be empty"),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
			},
		},
		{
			name: "cannot update if parent data has empty email",
			ctx:  ctx,
			req: &pb.UpdateParentsAndFamilyRelationshipRequest{
				SchoolId:  1,
				StudentId: "some-student-id",
				ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					{
						Id:       "",
						Email:    "parent-email@example.com",
						Username: "username",
						UserNameFields: &pb.UserNameFields{
							FirstName: "first-name",
							LastName:  "last-name",
						},
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "parent id cannot be empty"),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
			},
		},
		{
			name: "cannot update if parent data has empty email (username is disabled)",
			ctx:  ctx,
			req: &pb.UpdateParentsAndFamilyRelationshipRequest{
				SchoolId:  1,
				StudentId: "some-student-id",
				ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					{
						Id:       "id",
						Email:    "",
						Username: "username",
						UserNameFields: &pb.UserNameFields{
							FirstName: "first-name",
							LastName:  "last-name",
						},
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "parent email cannot be empty"),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
			},
		},
		{
			name: "cannot update if parent data has invalid relationship",
			ctx:  ctx,
			req: &pb.UpdateParentsAndFamilyRelationshipRequest{
				SchoolId:  1,
				StudentId: "some-student-id",
				ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					{
						Id:       "id",
						Email:    "parent-email@example.com",
						Username: "username",
						UserNameFields: &pb.UserNameFields{
							FirstName: "first-name",
							LastName:  "last-name",
						},
						Relationship: pb.FamilyRelationship(999999),
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "parent relationship is not valid"),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
			},
		},
		{
			name: "cannot update if parent data has invalid username",
			ctx:  ctx,
			req: &pb.UpdateParentsAndFamilyRelationshipRequest{
				SchoolId:  1,
				StudentId: "some-student-id",
				ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					{
						Id:           "id",
						Email:        "parent-email@example.com",
						Username:     "username-",
						Relationship: pb.FamilyRelationship(1),
						UserNameFields: &pb.UserNameFields{
							FirstName: "first_name",
							LastName:  "last_name",
						},
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "The 'username' of 'user' at -1 is not matching pattern"),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
			},
		},
		{
			name: "cannot update if parent data has missing username",
			ctx:  ctx,
			req: &pb.UpdateParentsAndFamilyRelationshipRequest{
				SchoolId:  1,
				StudentId: "some-student-id",
				ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					{
						Id:           "id",
						Email:        "parent-email@example.com",
						Relationship: pb.FamilyRelationship(1),
						UserNameFields: &pb.UserNameFields{
							FirstName: "first_name",
							LastName:  "last_name",
						},
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "The field 'username' is required in entity 'user' at index -1"),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
			},
		},
		{
			name: "cannot update if parent data has invalid last_name",
			ctx:  ctx,
			req: &pb.UpdateParentsAndFamilyRelationshipRequest{
				SchoolId:  1,
				StudentId: "some-student-id",
				ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					{
						Id:           "id",
						Email:        "parent-email@example.com",
						Username:     "username",
						Relationship: pb.FamilyRelationship(999999),
						UserNameFields: &pb.UserNameFields{
							FirstName: "first_name",
							LastName:  "",
						},
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, errcode.ErrUserLastNameIsEmpty.Error()),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
			},
		},
		{
			name: "cannot update if parent data has invalid first_name",
			ctx:  ctx,
			req: &pb.UpdateParentsAndFamilyRelationshipRequest{
				SchoolId:  1,
				StudentId: "some-student-id",
				ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					{
						Id:           "id",
						Email:        "parent-email@example.com",
						Relationship: pb.FamilyRelationship(999999),
						UserNameFields: &pb.UserNameFields{
							FirstName: "",
							LastName:  "last_name",
						},
						Username: "username",
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, errcode.ErrUserFirstNameIsEmpty.Error()),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
			},
		},
		{
			name: "cannot update if student to assign does not exist in db",
			ctx:  ctx,
			req: &pb.UpdateParentsAndFamilyRelationshipRequest{
				SchoolId:  1,
				StudentId: "some-student-id",
				ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					{
						Id:       "id",
						Email:    "parent-email@example.com",
						Username: "username",
						UserNameFields: &pb.UserNameFields{
							FirstName: "first-name",
							LastName:  "last-name",
						},
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "cannot update parents associated with un-existing student in system: some-student-id"),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				tagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{"some-student-id"})).Once().Return([]*entity.LegacyStudent{}, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				domainUserRepo.On("GetByUserNames", ctx, db, []string{"username"}).Once().Return(entity.Users{}, nil)
			},
		},
		{
			name: "cannot update if parent data has id not exist in db",
			ctx:  ctx,
			req: &pb.UpdateParentsAndFamilyRelationshipRequest{
				SchoolId:  1,
				StudentId: "some-student-id",
				ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					{
						Id:       "id",
						Email:    "parent-email@example.com",
						Username: "username",
						UserNameFields: &pb.UserNameFields{
							FirstName: "first-name",
							LastName:  "last-name",
						},
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "a parent ID in request does not exist in db"),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				tagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{"some-student-id"})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				parentRepo.On("GetByIds", ctx, tx, database.TextArray([]string{"id"})).Once().Return(entity.Parents{}, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				domainUserRepo.On("GetByUserNames", ctx, db, []string{"username"}).Once().Return(entity.Users{}, nil)
			},
		},
		{
			name: "cannot update if parent data has email already exist in db",
			ctx:  ctx,
			req: &pb.UpdateParentsAndFamilyRelationshipRequest{
				SchoolId:  1,
				StudentId: "some-student-id",
				ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					{
						Id:       "id",
						Email:    "parent-email@example.com",
						Username: "username",
						UserNameFields: &pb.UserNameFields{
							FirstName: "first-name",
							LastName:  "last-name",
						},
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
					},
				},
			},
			expectedErr: status.Error(codes.AlreadyExists, errcode.ErrUserEmailExists.Error()),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				tagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{"some-student-id"})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				parentRepo.On("GetByIds", ctx, tx, database.TextArray([]string{"id"})).Once().Return(entity.Parents{existingParent}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, []string{"parent-email@example.com"}).Once().Return([]*entity.LegacyUser{existingParentUser2}, nil)
				domainUserRepo.On("GetByUserNames", ctx, db, []string{"username"}).Once().Return(entity.Users{}, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "cannot update if parent data has school id different from student school id",
			ctx:  ctx,
			req: &pb.UpdateParentsAndFamilyRelationshipRequest{
				SchoolId:  1,
				StudentId: "some-student-id",
				ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					{
						Id:       "id",
						Email:    "parent-email@example.com",
						Username: "username",
						UserNameFields: &pb.UserNameFields{
							FirstName: "first-name",
							LastName:  "last-name",
						},
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "parent id not same school with student"),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				tagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{"some-student-id"})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				parentRepo.On("GetByIds", ctx, tx, database.TextArray([]string{"id"})).Once().Return(entity.Parents{existingParentOtherSchoolID}, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				domainUserRepo.On("GetByUserNames", ctx, db, []string{"username"}).Once().Return(entity.Users{}, nil)
			},
		},
		{
			name: "failed to update parent and family relationship because cannot get tenant id by organization id",
			ctx:  ctx,
			req: &pb.UpdateParentsAndFamilyRelationshipRequest{
				SchoolId:  1,
				StudentId: "some-student-id",
				ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					{
						Id:       "id",
						Email:    "parent-email@example.com",
						Username: "username",
						UserNameFields: &pb.UserNameFields{
							FirstName: "first-name",
							LastName:  "last-name",
						},
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Remarks:      "parent-remarks",
					},
				},
			},
			expectedErr: status.Error(codes.FailedPrecondition, errcode.TenantDoesNotExistErr{OrganizationID: fmt.Sprint(constants.ManabieSchool)}.Error()),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(false, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				tagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{"some-student-id"})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				parentRepo.On("GetByIds", ctx, tx, database.TextArray([]string{"id"})).Once().Return(entity.Parents{existingParent}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, []string{"parent-email@example.com"}).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				userRepo.On("UpdateEmail", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("UpdateProfileV1", ctx, tx, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("FindLocationIDsFromUserID", ctx, tx, mock.Anything).Once().Return([]string{"testing"}, nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
				taggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentParentRepo.On("FindParentIDsFromStudentID", ctx, tx, mock.Anything).Return(nil, nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", pgx.ErrNoRows)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				domainUserRepo.On("GetByUserNames", ctx, db, []string{"username"}).Once().Return(entity.Users{}, nil)
			},
		},
		{
			name: "failed to update parent and family relationship because tenant manager cannot get tenant client",
			ctx:  ctx,
			req: &pb.UpdateParentsAndFamilyRelationshipRequest{
				SchoolId:  1,
				StudentId: "some-student-id",
				ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					{
						Id:       "id",
						Email:    "parent-email@example.com",
						Username: "username",
						UserNameFields: &pb.UserNameFields{
							FirstName: "first-name",
							LastName:  "last-name",
						},
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Remarks:      "parent-remarks",
					},
				},
			},
			expectedErr: status.Error(codes.Internal, errors.Wrap(internal_auth_user.ErrTenantNotFound, "TenantClient").Error()),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(false, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				tagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{"some-student-id"})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				parentRepo.On("GetByIds", ctx, tx, database.TextArray([]string{"id"})).Once().Return(entity.Parents{existingParent}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, []string{"parent-email@example.com"}).Once().Return([]*entity.LegacyUser{}, nil)
				userRepo.On("UpdateEmail", ctx, tx, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				userRepo.On("UpdateProfileV1", ctx, tx, mock.Anything).Once().Return(nil)
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				taggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentParentRepo.On("FindParentIDsFromStudentID", ctx, tx, mock.Anything).Return(nil, nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return(aValidTenantID, nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(nil, internal_auth_user.ErrTenantNotFound)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				domainUserRepo.On("GetByUserNames", ctx, db, []string{"username"}).Once().Return(entity.Users{}, nil)
			},
		},
		{
			name: "failed to update parent and family relationship because username is existed in system",
			ctx:  ctx,
			req: &pb.UpdateParentsAndFamilyRelationshipRequest{
				SchoolId:  1,
				StudentId: "some-student-id",
				ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					{
						Id:             "id",
						ExternalUserId: "external-user-id",
						Email:          "parent-email@example.com",
						Username:       "parent-email@example.com",
						UserNameFields: &pb.UserNameFields{
							FirstName: "first-name",
							LastName:  "last-name",
						},
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Remarks:      "parent-remarks",
						TagIds:       []string{id},
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, entity.ExistingDataError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      0,
			}.Error()),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				domainUserRepo.On("GetByUserNames", ctx, db, []string{"parent-email@example.com"}).Once().Return(entity.Users{
					&repository.User{
						ID:           field.NewString(idutil.ULIDNow()),
						UserNameAttr: field.NewString("parent-email@example.com"),
					},
				}, nil)
			},
		},
		{
			name: "update parent and family relationship successfully with username has email format",
			ctx:  ctx,
			req: &pb.UpdateParentsAndFamilyRelationshipRequest{
				SchoolId:  1,
				StudentId: "some-student-id",
				ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					{
						Id:             "id",
						ExternalUserId: "external-user-id",
						Email:          "parent-email@example.com",
						Username:       "parent-email@example.com",
						UserNameFields: &pb.UserNameFields{
							FirstName: "first-name",
							LastName:  "last-name",
						},
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Remarks:      "parent-remarks",
						TagIds:       []string{id},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				domainUserRepo.On("GetByUserNames", ctx, db, []string{"parent-email@example.com"}).Once().Return(entity.Users{}, nil)
				tagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags(
					[]entity.DomainTag{
						createMockDomainTagWithType(id, pb.UserTagType_USER_TAG_TYPE_PARENT),
					},
				), nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{"some-student-id"})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				parentRepo.On("GetByIds", ctx, tx, database.TextArray([]string{"id"})).Once().Return(entity.Parents{existingParent}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, []string{"parent-email@example.com"}).Once().Return([]*entity.LegacyUser{}, nil)
				userRepo.On("UpdateEmail", ctx, tx, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				userRepo.On("UpdateProfileV1", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				taggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return(nil, nil)
				taggedUserRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				jsm.On("TracedPublish", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "update parent and family relationship successfully with username has username format",
			ctx:  ctx,
			req: &pb.UpdateParentsAndFamilyRelationshipRequest{
				SchoolId:  1,
				StudentId: "some-student-id",
				ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					{
						Id:             "id",
						ExternalUserId: "external-user-id",
						Email:          "parent-email@example.com",
						Username:       "username",
						UserNameFields: &pb.UserNameFields{
							FirstName: "first-name",
							LastName:  "last-name",
						},
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Remarks:      "parent-remarks",
						TagIds:       []string{id},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				domainUserRepo.On("GetByUserNames", ctx, db, []string{"username"}).Once().Return(entity.Users{}, nil)
				tagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags(
					[]entity.DomainTag{
						createMockDomainTagWithType(id, pb.UserTagType_USER_TAG_TYPE_PARENT),
					},
				), nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{"some-student-id"})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				parentRepo.On("GetByIds", ctx, tx, database.TextArray([]string{"id"})).Once().Return(entity.Parents{existingParent}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, []string{"parent-email@example.com"}).Once().Return([]*entity.LegacyUser{}, nil)
				userRepo.On("UpdateEmail", ctx, tx, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				userRepo.On("UpdateProfileV1", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				taggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return(nil, nil)
				taggedUserRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				jsm.On("TracedPublish", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "update parent and family relationship successfully",
			ctx:  ctx,
			req: &pb.UpdateParentsAndFamilyRelationshipRequest{
				SchoolId:  1,
				StudentId: "some-student-id",
				ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					{
						Id:             "id",
						ExternalUserId: "external-user-id",
						Email:          "parent-email@example.com",
						Username:       "username",
						UserNameFields: &pb.UserNameFields{
							FirstName: "first-name",
							LastName:  "last-name",
						},
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Remarks:      "parent-remarks",
						TagIds:       []string{id},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				tagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags(
					[]entity.DomainTag{
						createMockDomainTagWithType(id, pb.UserTagType_USER_TAG_TYPE_PARENT),
					},
				), nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{"some-student-id"})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				parentRepo.On("GetByIds", ctx, tx, database.TextArray([]string{"id"})).Once().Return(entity.Parents{existingParent}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, []string{"parent-email@example.com"}).Once().Return([]*entity.LegacyUser{}, nil)
				userRepo.On("UpdateEmail", ctx, tx, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				userRepo.On("UpdateProfileV1", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				taggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return(nil, nil)
				taggedUserRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				jsm.On("TracedPublish", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				domainUserRepo.On("GetByUserNames", ctx, db, []string{"username"}).Once().Return(entity.Users{}, nil)
			},
		},
		{
			name: "fail when UserRepo.UpdateProfileV1 err",
			ctx:  ctx,
			req: &pb.UpdateParentsAndFamilyRelationshipRequest{
				SchoolId:  1,
				StudentId: "some-student-id",
				ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					{
						Id:       "id",
						Email:    "parent-email@example.com",
						Username: "username",
						UserNameFields: &pb.UserNameFields{
							FirstName: "first-name",
							LastName:  "last-name",
						},
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Remarks:      "parent-remarks",
					},
				},
			},
			expectedErr: errors.New("rpc error: code = Internal desc = s.UserRepo.UpdateProfileV1: update error"),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				tagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{"some-student-id"})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				parentRepo.On("GetByIds", ctx, tx, database.TextArray([]string{"id"})).Once().Return(entity.Parents{existingParent}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, []string{"parent-email@example.com"}).Once().Return([]*entity.LegacyUser{}, nil)
				userRepo.On("UpdateEmail", ctx, tx, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				userRepo.On("UpdateProfileV1", ctx, tx, mock.Anything).Once().Return(errors.New("update error"))
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Once().Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "fail when UserPhoneNumberRepo.Upsert err",
			ctx:  ctx,
			req: &pb.UpdateParentsAndFamilyRelationshipRequest{
				SchoolId:  1,
				StudentId: "some-student-id",
				ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					{
						Id:       "id",
						Email:    "parent-email@example.com",
						Username: "username",
						UserNameFields: &pb.UserNameFields{
							FirstName: "first-name",
							LastName:  "last-name",
						},
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Remarks:      "parent-remarks",
					},
				},
			},
			expectedErr: errors.New("rpc error: code = Unknown desc = upsert error"),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				tagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{"some-student-id"})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				parentRepo.On("GetByIds", ctx, tx, database.TextArray([]string{"id"})).Once().Return(entity.Parents{existingParent}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, []string{"parent-email@example.com"}).Once().Return([]*entity.LegacyUser{}, nil)
				userRepo.On("UpdateEmail", ctx, tx, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				userRepo.On("UpdateProfileV1", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(errors.New("upsert error"))
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				taggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return(nil, nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Once().Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "cannot publish event",
			ctx:  ctx,
			req: &pb.UpdateParentsAndFamilyRelationshipRequest{
				SchoolId:  1,
				StudentId: "some-student-id",
				ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					{
						Id:       "id",
						Email:    "parent-email@example.com",
						Username: "username",
						UserNameFields: &pb.UserNameFields{
							FirstName: "first-name",
							LastName:  "last-name",
						},
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Remarks:      "parent-remarks",
					},
				},
			},
			expectedErr: errors.New("publishUserEvent with User.Updated: s.JSM.Publish failed: publish error"),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				tagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{"some-student-id"})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				parentRepo.On("GetByIds", ctx, tx, database.TextArray([]string{"id"})).Once().Return(entity.Parents{existingParent}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, []string{"parent-email@example.com"}).Once().Return([]*entity.LegacyUser{}, nil)
				userRepo.On("UpdateEmail", ctx, tx, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				userRepo.On("UpdateProfileV1", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				taggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return(nil, nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Once().Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				jsm.On("TracedPublish", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, errors.New("publish error"))
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				domainUserRepo.On("GetByUserNames", ctx, db, []string{"username"}).Once().Return(entity.Users{}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log("Test case: " + testCase.name)
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.ManabieSchool),
				},
			}
			testCase.ctx = interceptors.ContextWithJWTClaims(ctx, claim)
			testCase.setup(testCase.ctx)

			_, err := s.UpdateParentsAndFamilyRelationship(testCase.ctx, testCase.req.(*pb.UpdateParentsAndFamilyRelationshipRequest))
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

func TestUserModifierService_PBParentPhoneNumberToUserPhoneNumbers(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		req          interface{}
		userID       string
		resourcePath string
		expectedErr  error
	}{
		{
			name: "happy case: update parent with ParentPhoneNumbers",
			req: &pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
				Id: "testing id",
				ParentPhoneNumbers: []*pb.ParentPhoneNumber{
					{PhoneNumber: "4567812312", PhoneNumberType: pb.ParentPhoneNumber_PARENT_PRIMARY_PHONE_NUMBER},
					{PhoneNumber: "123456789", PhoneNumberType: pb.ParentPhoneNumber_PARENT_SECONDARY_PHONE_NUMBER},
				},
			},
			userID:       "UserId",
			resourcePath: fmt.Sprint(constants.ManabieSchool),
			expectedErr:  nil,
		},
		{
			name: "happy case when have phone number id in ParentPhoneNumbers",

			req: &pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
				Id: "testing id",
				ParentPhoneNumbers: []*pb.ParentPhoneNumber{
					{PhoneNumber: "4567812312", PhoneNumberType: pb.ParentPhoneNumber_PARENT_PRIMARY_PHONE_NUMBER, PhoneNumberId: "phone_number_id_1"},
				},
			},
			userID:       "UserId",
			resourcePath: fmt.Sprint(constants.ManabieSchool),
			expectedErr:  nil,
		},
		{
			name: "happy case: update parent when having full field UserNameFields",
			req: &pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
				Id: "testing id",
				UserNameFields: &pb.UserNameFields{
					FirstName:         "first_name",
					LastName:          "last_name",
					FirstNamePhonetic: "first_name_phonetic",
					LastNamePhonetic:  "last_name_phonetic",
				},
			},
			userID:       "UserId",
			resourcePath: fmt.Sprint(constants.ManabieSchool),
			expectedErr:  nil,
		},
		{
			name: "happy case: update parent when having nill UserNameFields",
			req: &pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
				Id:             "testing id",
				UserNameFields: nil,
			},
			userID:       "UserId",
			resourcePath: fmt.Sprint(constants.ManabieSchool),
			expectedErr:  nil,
		},
		{
			name: "happy case: update parent without first_name_phonetic && last_name_phonetic",
			req: &pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
				Id: "testing id",
				UserNameFields: &pb.UserNameFields{
					FirstName: "first_name",
					LastName:  "last_name",
				},
			},
			userID:       "UserId",
			resourcePath: fmt.Sprint(constants.ManabieSchool),
			expectedErr:  nil,
		},
		{
			name: "fail case when PhoneNumberType don't have in parentPhoneNumber",

			req: &pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
				Id: "testing id",
				ParentPhoneNumbers: []*pb.ParentPhoneNumber{
					{PhoneNumber: "4567812312", PhoneNumberType: 3, PhoneNumberId: "phone_number_id_1"},
				},
			},
			userID:       "UserId",
			resourcePath: fmt.Sprint(constants.ManabieSchool),
			expectedErr:  errors.New("don't have that type of parent phone"),
		},
	}

	for _, tcase := range testCases {
		t.Run(tcase.name, func(t *testing.T) {
			request := tcase.req.(*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile)
			userPhoneNumbers, err := pbParentPhoneNumberToUserPhoneNumbers(
				request,
				tcase.resourcePath,
			)
			if err != nil {
				assert.Equal(t, tcase.expectedErr.Error(), err.Error())
				return
			}
			assert.NoError(t, err)

			assert.Equal(t, len(request.ParentPhoneNumbers), len(userPhoneNumbers))

			for index, value := range userPhoneNumbers {
				assert.Equal(t, request.ParentPhoneNumbers[index].PhoneNumber, value.PhoneNumber.Get())
				assert.Equal(t, request.ParentPhoneNumbers[index].PhoneNumberType.String(), value.PhoneNumberType.Get())
				assert.Equal(t, request.Id, value.UserID.Get())
				assert.Equal(t, tcase.resourcePath, value.ResourcePath.Get())
			}
		})
	}
}
