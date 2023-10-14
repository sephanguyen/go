package service

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/manabie-com/backend/internal/bob/constants"
	internal_auth_user "github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
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

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateParentsAndAssignToStudent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	tx := new(mock_database.Tx)
	orgRepo := new(mock_repositories.OrganizationRepo)
	usrEmailRepo := new(mock_repositories.MockUsrEmailRepo)
	userRepo := new(mock_repositories.MockUserRepo)
	studentRepo := new(mock_repositories.MockStudentRepo)
	parentRepo := new(mock_repositories.MockParentRepo)
	studentParentRepo := new(mock_repositories.MockStudentParentRepo)
	userGroupRepo := new(mock_repositories.MockUserGroupRepo)
	userGroupV2Repo := new(mock_repositories.MockUserGroupV2Repo)
	userGroupsMemberRepo := new(mock_repositories.MockUserGroupsMemberRepo)
	userPhoneNumberRepo := new(mock_repositories.MockUserPhoneNumberRepo)
	userAccessPathRepo := new(mock_repositories.MockUserAccessPathRepo)
	firebaseAuthClient := new(mock_multitenant.TenantClient)
	tenantManager := new(mock_multitenant.TenantManager)
	jsm := new(mock_nats.JetStreamManagement)
	sampleID := uuid.NewString()
	domainTagRepo := new(mock_repositories.MockDomainTagRepo)
	domainTaggedUserRepo := new(mock_repositories.MockDomainTaggedUserRepo)
	domainUserRepo := new(mock_repositories.MockDomainUserRepo)
	unleashClient := new(mock_unleash_client.UnleashClientInstance)
	firebaseAuth := new(mock_firebase.AuthClient)
	internalConfigurationRepo := new(mock_repositories.MockDomainInternalConfigurationRepo)

	s := UserModifierService{
		DB:                        db,
		OrganizationRepo:          orgRepo,
		UsrEmailRepo:              usrEmailRepo,
		UserRepo:                  userRepo,
		StudentRepo:               studentRepo,
		ParentRepo:                parentRepo,
		StudentParentRepo:         studentParentRepo,
		UserGroupRepo:             userGroupRepo,
		UserGroupV2Repo:           userGroupV2Repo,
		UserGroupsMemberRepo:      userGroupsMemberRepo,
		UserPhoneNumberRepo:       userPhoneNumberRepo,
		UserAccessPathRepo:        userAccessPathRepo,
		FirebaseClient:            firebaseAuth,
		FirebaseAuthClient:        firebaseAuthClient,
		TenantManager:             tenantManager,
		UnleashClient:             unleashClient,
		DomainTagRepo:             domainTagRepo,
		DomainTaggedUserRepo:      domainTaggedUserRepo,
		JSM:                       jsm,
		DomainUserRepo:            domainUserRepo,
		InternalConfigurationRepo: internalConfigurationRepo,
	}

	existingParentUser := &entity.LegacyUser{
		ID:                 database.Text(idutil.ULIDNow()),
		GivenName:          database.Text("existing-parent-name"),
		Email:              database.Text("existing-parent-email@example.com"),
		PhoneNumber:        database.Text("existing-parent-phone-number"),
		UserAdditionalInfo: entity.UserAdditionalInfo{Password: "password"},
		FirstName:          database.Text("first-name"),
		LastName:           database.Text("last-name"),
		FirstNamePhonetic:  database.Text("first-name-phonetic"),
		LastNamePhonetic:   database.Text("last-name-phonetic"),
		FullName:           database.Text(CombineFirstNameAndLastNameToFullName("first-name", "last-name")),
		FullNamePhonetic:   database.Text(CombineFirstNamePhoneticAndLastNamePhoneticToFullName("first-name-phonetic", "last-name-phonetic")),
		LoginEmail:         database.Text("existing-parent-login-email@example.com"),
	}
	existingParent := &entity.Parent{
		ID:         existingParentUser.ID,
		SchoolID:   database.Int4(1),
		LegacyUser: *existingParentUser,
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
			name: "cannot create if student ID to assign is empty",
			ctx:  ctx,
			req: &pb.CreateParentsAndAssignToStudentRequest{
				SchoolId:  constants.ManabieSchool,
				StudentId: "",
				ParentProfiles: []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					{
						Name:         existingParentUser.GetDisplayName(),
						CountryCode:  cpb.Country_COUNTRY_VN,
						PhoneNumber:  existingParentUser.GetPhoneNumber(),
						Email:        existingParent.Email.String,
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     existingParentUser.GetRawPassword(),
						UserNameFields: &pb.UserNameFields{
							FirstName:         existingParentUser.FirstName.String,
							LastName:          existingParentUser.LastName.String,
							FirstNamePhonetic: existingParentUser.FirstNamePhonetic.String,
							LastNamePhonetic:  existingParentUser.LastNamePhonetic.String,
						},
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
			name: "cannot create if parent data has empty email",
			ctx:  ctx,
			req: &pb.CreateParentsAndAssignToStudentRequest{
				SchoolId:  constants.ManabieSchool,
				StudentId: sampleID,
				ParentProfiles: []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					{
						Name:         existingParentUser.GetDisplayName(),
						CountryCode:  cpb.Country_COUNTRY_VN,
						PhoneNumber:  existingParentUser.GetPhoneNumber(),
						Email:        "",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     existingParentUser.GetRawPassword(),
						UserNameFields: &pb.UserNameFields{
							FirstName:         existingParentUser.FirstName.String,
							LastName:          existingParentUser.LastName.String,
							FirstNamePhonetic: existingParentUser.FirstNamePhonetic.String,
							LastNamePhonetic:  existingParentUser.LastNamePhonetic.String,
						},
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
			name: "cannot create if parent data has empty name",
			ctx:  ctx,
			req: &pb.CreateParentsAndAssignToStudentRequest{
				SchoolId:  constants.ManabieSchool,
				StudentId: sampleID,
				ParentProfiles: []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					{
						Name:         "",
						CountryCode:  cpb.Country_COUNTRY_VN,
						PhoneNumber:  existingParentUser.GetPhoneNumber(),
						Email:        existingParent.Email.String,
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     existingParentUser.GetRawPassword(),
						UserNameFields: &pb.UserNameFields{
							FirstName:         "",
							LastName:          "",
							LastNamePhonetic:  "",
							FirstNamePhonetic: "",
						},
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "parent first_name cannot be empty"),
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
			name: "cannot create if parent data has empty country code",
			ctx:  ctx,
			req: &pb.CreateParentsAndAssignToStudentRequest{
				SchoolId:  constants.ManabieSchool,
				StudentId: sampleID,
				ParentProfiles: []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					{
						Name:         existingParentUser.GetDisplayName(),
						CountryCode:  cpb.Country(999999),
						PhoneNumber:  existingParentUser.GetPhoneNumber(),
						Email:        existingParent.Email.String,
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     existingParentUser.GetRawPassword(),
						UserNameFields: &pb.UserNameFields{
							FirstName:         existingParentUser.FirstName.String,
							LastName:          existingParentUser.LastName.String,
							FirstNamePhonetic: existingParentUser.FirstNamePhonetic.String,
							LastNamePhonetic:  existingParentUser.LastNamePhonetic.String,
						},
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "parent country code is not valid"),
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
			name: "cannot create if parent data has password too short ( < 6 chars )",
			ctx:  ctx,
			req: &pb.CreateParentsAndAssignToStudentRequest{
				SchoolId:  constants.ManabieSchool,
				StudentId: sampleID,
				ParentProfiles: []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					{
						Name:         existingParentUser.GetDisplayName(),
						CountryCode:  cpb.Country_COUNTRY_VN,
						PhoneNumber:  existingParentUser.GetPhoneNumber(),
						Email:        existingParent.Email.String,
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     existingParentUser.GetRawPassword()[:2],
						UserNameFields: &pb.UserNameFields{
							FirstName:         existingParentUser.FirstName.String,
							LastName:          existingParentUser.LastName.String,
							FirstNamePhonetic: existingParentUser.FirstNamePhonetic.String,
							LastNamePhonetic:  existingParentUser.LastNamePhonetic.String,
						},
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "parent password length should be at least 6"),
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
			name: "cannot create if username is empty",
			ctx:  ctx,
			req: &pb.CreateParentsAndAssignToStudentRequest{
				SchoolId:  constants.ManabieSchool,
				StudentId: idutil.ULIDNow(),
				ParentProfiles: []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					{
						Name:         existingParentUser.GetDisplayName(),
						CountryCode:  cpb.Country_COUNTRY_VN,
						PhoneNumber:  existingParentUser.GetPhoneNumber(),
						Email:        existingParent.Email.String,
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     existingParentUser.GetRawPassword(),
						UserNameFields: &pb.UserNameFields{
							FirstName:         existingParentUser.FirstName.String,
							LastName:          existingParentUser.LastName.String,
							FirstNamePhonetic: existingParentUser.FirstNamePhonetic.String,
							LastNamePhonetic:  existingParentUser.LastNamePhonetic.String,
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
			name: "cannot create if username was used by another user",
			ctx:  ctx,
			req: &pb.CreateParentsAndAssignToStudentRequest{
				SchoolId:  constants.ManabieSchool,
				StudentId: sampleID,
				ParentProfiles: []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					{
						Name:         existingParentUser.GetDisplayName(),
						CountryCode:  cpb.Country_COUNTRY_VN,
						PhoneNumber:  existingParentUser.GetPhoneNumber(),
						Email:        existingParent.Email.String,
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Password:     existingParentUser.GetRawPassword(),
						TagIds:       []string{"tag_id1", "tag_id2", "tag_id3"},
						Username:     "username",
						UserNameFields: &pb.UserNameFields{
							FirstName:         existingParentUser.FirstName.String,
							LastName:          existingParentUser.LastName.String,
							FirstNamePhonetic: existingParentUser.FirstNamePhonetic.String,
							LastNamePhonetic:  existingParentUser.LastNamePhonetic.String,
						},
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
				domainUserRepo.On("GetByUserNames", ctx, db, mock.Anything).Once().Return(entity.Users{&repository.User{ID: field.NewString("user_id"), UserNameAttr: field.NewString("username")}}, nil)
			},
		},
		{
			name: "cannot create if tag id was not existed in system",
			ctx:  ctx,
			req: &pb.CreateParentsAndAssignToStudentRequest{
				SchoolId:  constants.ManabieSchool,
				StudentId: sampleID,
				ParentProfiles: []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					{
						Name:         existingParentUser.GetDisplayName(),
						CountryCode:  cpb.Country_COUNTRY_VN,
						PhoneNumber:  existingParentUser.GetPhoneNumber(),
						Email:        existingParent.Email.String,
						Username:     "username",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Password:     existingParentUser.GetRawPassword(),
						TagIds:       []string{"tag_id1", "tag_id2", "tag_id3"},
						UserNameFields: &pb.UserNameFields{
							FirstName:         existingParentUser.FirstName.String,
							LastName:          existingParentUser.LastName.String,
							FirstNamePhonetic: existingParentUser.FirstNamePhonetic.String,
							LastNamePhonetic:  existingParentUser.LastNamePhonetic.String,
						},
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, ErrTagIDsMustBeExisted.Error()),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				domainTagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags(
					[]entity.DomainTag{
						createMockDomainTag("tag_id1"),
						createMockDomainTag("tag_id2"),
					},
				), nil)
				domainUserRepo.On("GetByUserNames", ctx, db, mock.Anything).Once().Return(entity.Users{}, nil)
			},
		},
		{
			name: "cannot create if student to assign does not exist in db",
			ctx:  ctx,
			req: &pb.CreateParentsAndAssignToStudentRequest{
				SchoolId:  constants.ManabieSchool,
				StudentId: sampleID,
				ParentProfiles: []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					{
						Name:         existingParentUser.GetDisplayName(),
						CountryCode:  cpb.Country_COUNTRY_VN,
						PhoneNumber:  existingParentUser.GetPhoneNumber(),
						Email:        existingParentUser.Email.String,
						Username:     "username",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     existingParentUser.GetRawPassword(),
						UserNameFields: &pb.UserNameFields{
							FirstName:         existingParentUser.FirstName.String,
							LastName:          existingParentUser.LastName.String,
							FirstNamePhonetic: existingParentUser.FirstNamePhonetic.String,
							LastNamePhonetic:  existingParentUser.LastNamePhonetic.String,
						},
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "cannot assign parent with un-existing student in system: "+sampleID),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				domainTagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{sampleID})).Once().Return([]*entity.LegacyStudent{}, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				domainUserRepo.On("GetByUserNames", ctx, db, mock.Anything).Once().Return(entity.Users{}, nil)
			},
		},
		{
			name: "cannot create if parent data has email already exist in db",
			ctx:  ctx,
			req: &pb.CreateParentsAndAssignToStudentRequest{
				SchoolId:  constants.ManabieSchool,
				StudentId: sampleID,
				ParentProfiles: []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					{
						Name:         existingParentUser.GetDisplayName(),
						CountryCode:  cpb.Country_COUNTRY_VN,
						PhoneNumber:  existingParentUser.GetPhoneNumber(),
						Email:        existingParentUser.Email.String,
						Username:     "username",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     existingParentUser.GetRawPassword(),
						UserNameFields: &pb.UserNameFields{
							FirstName:         existingParentUser.FirstName.String,
							LastName:          existingParentUser.LastName.String,
							FirstNamePhonetic: existingParentUser.FirstNamePhonetic.String,
							LastNamePhonetic:  existingParentUser.LastNamePhonetic.String,
						},
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
				domainTagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				domainUserRepo.On("GetByUserNames", ctx, db, mock.Anything).Once().Return(entity.Users{}, nil)
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{sampleID})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{existingParentUser}, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "cannot create if cannot get tenant id by organization id",
			ctx:  ctx,
			req: &pb.CreateParentsAndAssignToStudentRequest{
				SchoolId:  constants.ManabieSchool,
				StudentId: sampleID,
				ParentProfiles: []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					{
						Name:         existingParentUser.GetDisplayName(),
						CountryCode:  cpb.Country_COUNTRY_VN,
						PhoneNumber:  existingParentUser.PhoneNumber.String,
						Email:        existingParent.Email.String,
						Username:     "username",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Password:     existingParentUser.GetRawPassword(),
						UserNameFields: &pb.UserNameFields{
							FirstName:         existingParentUser.FirstName.String,
							LastName:          existingParentUser.LastName.String,
							FirstNamePhonetic: existingParentUser.FirstNamePhonetic.String,
							LastNamePhonetic:  existingParentUser.LastNamePhonetic.String,
						},
					},
				},
			},
			expectedErr: status.Error(codes.FailedPrecondition, errcode.TenantDoesNotExistErr{OrganizationID: "1"}.Error()),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				domainUserRepo.On("GetByUserNames", ctx, db, mock.Anything).Once().Return(entity.Users{}, nil)
				domainTagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{sampleID})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				userRepo.On("GetByPhone", ctx, tx, mock.Anything).Once().Return(nil, nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				parentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				studentParentRepo.On("InsertParentAccessPathByStudentID", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleParent).Once().Return(&entity.UserGroupV2{}, nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", pgx.ErrNoRows)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "cannot create if tenant manager cannot get tenant client",
			ctx:  ctx,
			req: &pb.CreateParentsAndAssignToStudentRequest{
				SchoolId:  constants.ManabieSchool,
				StudentId: sampleID,
				ParentProfiles: []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					{
						Name:         existingParentUser.GetDisplayName(),
						CountryCode:  cpb.Country_COUNTRY_VN,
						PhoneNumber:  existingParentUser.PhoneNumber.String,
						Email:        existingParent.Email.String,
						Username:     "username",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Password:     existingParentUser.GetRawPassword(),
						UserNameFields: &pb.UserNameFields{
							FirstName:         existingParentUser.FirstName.String,
							LastName:          existingParentUser.LastName.String,
							FirstNamePhonetic: existingParentUser.FirstNamePhonetic.String,
							LastNamePhonetic:  existingParentUser.LastNamePhonetic.String,
						},
					},
				},
			},
			expectedErr: status.Error(codes.Internal, errors.Wrap(internal_auth_user.ErrTenantNotFound, "TenantClient").Error()),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				domainUserRepo.On("GetByUserNames", ctx, db, mock.Anything).Once().Return(entity.Users{}, nil)
				domainTagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{sampleID})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				userRepo.On("GetByPhone", ctx, tx, mock.Anything).Once().Return(nil, nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				parentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				studentParentRepo.On("InsertParentAccessPathByStudentID", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleParent).Once().Return(&entity.UserGroupV2{}, nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return(aValidTenantID, nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(nil, internal_auth_user.ErrTenantNotFound)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "error when assigning user with user group",
			ctx:  ctx,
			req: &pb.CreateParentsAndAssignToStudentRequest{
				SchoolId:  constants.ManabieSchool,
				StudentId: sampleID,
				ParentProfiles: []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					{
						Name:         existingParentUser.GetDisplayName(),
						CountryCode:  cpb.Country_COUNTRY_VN,
						PhoneNumber:  existingParentUser.GetPhoneNumber(),
						Email:        existingParent.Email.String,
						Username:     "username",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Password:     existingParentUser.GetRawPassword(),
						UserNameFields: &pb.UserNameFields{
							FirstName:         existingParentUser.FirstName.String,
							LastName:          existingParentUser.LastName.String,
							FirstNamePhonetic: existingParentUser.FirstNamePhonetic.String,
							LastNamePhonetic:  existingParentUser.LastNamePhonetic.String,
						},
					},
				},
			},
			expectedErr: status.Error(codes.Internal, errors.Wrapf(fmt.Errorf("error"), "can not assign parent user group to user %s", existingStudent.GetUID()).Error()),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				domainUserRepo.On("GetByUserNames", ctx, db, mock.Anything).Once().Return(entity.Users{}, nil)
				domainTagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{sampleID})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				userRepo.On("GetByPhone", ctx, tx, mock.Anything).Once().Return(nil, nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				parentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				studentParentRepo.On("InsertParentAccessPathByStudentID", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleParent).Once().Return(&entity.UserGroupV2{}, nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "create parent successfully without parent user group",
			ctx:  ctx,
			req: &pb.CreateParentsAndAssignToStudentRequest{
				SchoolId:  constants.ManabieSchool,
				StudentId: sampleID,
				ParentProfiles: []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					{
						Name:         existingParentUser.GetDisplayName(),
						CountryCode:  cpb.Country_COUNTRY_VN,
						PhoneNumber:  existingParentUser.GetPhoneNumber(),
						Email:        existingParent.Email.String,
						Username:     "username",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Password:     existingParentUser.GetRawPassword(),
						UserNameFields: &pb.UserNameFields{
							FirstName:         existingParentUser.FirstName.String,
							LastName:          existingParentUser.LastName.String,
							FirstNamePhonetic: existingParentUser.FirstNamePhonetic.String,
							LastNamePhonetic:  existingParentUser.LastNamePhonetic.String,
						},
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
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				domainUserRepo.On("GetByUserNames", ctx, db, mock.Anything).Once().Return(entity.Users{}, nil)
				domainTagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleParent).Once().Return(nil, fmt.Errorf("error"))
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{sampleID})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				userRepo.On("GetByPhone", ctx, tx, mock.Anything).Once().Return(nil, nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				parentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				studentParentRepo.On("InsertParentAccessPathByStudentID", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
				hashConfig := mockScryptHash()
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("TracedPublishAsync", mock.Anything, "nats.TracedPublishAsync", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name: "error when upserting user tag for parent",
			ctx:  ctx,
			req: &pb.CreateParentsAndAssignToStudentRequest{
				SchoolId:  constants.ManabieSchool,
				StudentId: sampleID,
				ParentProfiles: []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					{
						Name:         existingParentUser.GetDisplayName(),
						CountryCode:  cpb.Country_COUNTRY_VN,
						PhoneNumber:  existingParentUser.GetPhoneNumber(),
						Email:        existingParent.Email.String,
						Username:     "username",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Password:     existingParentUser.GetRawPassword(),
						TagIds:       []string{"tag_id1", "tag_id2", "tag_id3"},
						UserNameFields: &pb.UserNameFields{
							FirstName:         existingParentUser.FirstName.String,
							LastName:          existingParentUser.LastName.String,
							FirstNamePhonetic: existingParentUser.FirstNamePhonetic.String,
							LastNamePhonetic:  existingParentUser.LastNamePhonetic.String,
						},
					},
				},
			},
			expectedErr: status.Error(codes.Internal, errors.Wrap(errors.Wrap(fmt.Errorf("error"), "DomainTaggedUserRepo.UpsertBatch"), "UpsertTaggedUsers").Error()),
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
				internalConfigurationRepo.On("GetByKey", ctx, mock.Anything, mock.Anything).Once().Return(mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigValue: field.NewString("on"),
					},
				}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				domainUserRepo.On("GetByUserNames", ctx, db, mock.Anything).Once().Return(entity.Users{}, nil)
				domainTagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags(
					[]entity.DomainTag{
						createMockDomainTagWithType("tag_id1", pb.UserTagType_USER_TAG_TYPE_PARENT),
						createMockDomainTagWithType("tag_id2", pb.UserTagType_USER_TAG_TYPE_PARENT),
						createMockDomainTagWithType("tag_id3", pb.UserTagType_USER_TAG_TYPE_PARENT),
					},
				), nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{sampleID})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				userRepo.On("GetByPhone", ctx, tx, mock.Anything).Once().Return(nil, nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				parentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				studentParentRepo.On("InsertParentAccessPathByStudentID", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleParent).Once().Return(&entity.UserGroupV2{}, nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				hashConfig := mockScryptHash()
				tenantClient := &mock_multitenant.TenantClient{}
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				domainTaggedUserRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(fmt.Errorf("error"))
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "create parent successfully with existed parent",
			ctx:  ctx,
			req: &pb.CreateParentsAndAssignToStudentRequest{
				SchoolId:  constants.ManabieSchool,
				StudentId: sampleID,
				ParentProfiles: []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					{
						Name:         existingParentUser.GetDisplayName(),
						CountryCode:  cpb.Country_COUNTRY_VN,
						PhoneNumber:  existingParentUser.GetPhoneNumber(),
						Email:        existingParent.Email.String,
						Username:     "username",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Password:     existingParentUser.GetRawPassword(),
						TagIds:       []string{"tag_id1", "tag_id2", "tag_id3"},
						UserNameFields: &pb.UserNameFields{
							FirstName:         existingParentUser.FirstName.String,
							LastName:          existingParentUser.LastName.String,
							FirstNamePhonetic: existingParentUser.FirstNamePhonetic.String,
							LastNamePhonetic:  existingParentUser.LastNamePhonetic.String,
						},
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
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				domainUserRepo.On("GetByUserNames", ctx, db, mock.Anything).Once().Return(entity.Users{}, nil)
				domainTagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags(
					[]entity.DomainTag{
						createMockDomainTagWithType("tag_id1", pb.UserTagType_USER_TAG_TYPE_PARENT),
						createMockDomainTagWithType("tag_id2", pb.UserTagType_USER_TAG_TYPE_PARENT),
						createMockDomainTagWithType("tag_id3", pb.UserTagType_USER_TAG_TYPE_PARENT),
					},
				), nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{sampleID})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				userRepo.On("GetByPhone", ctx, tx, mock.Anything).Once().Return(nil, nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				parentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
				studentParentRepo.On("InsertParentAccessPathByStudentID", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleParent).Once().Return(&entity.UserGroupV2{}, nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				hashConfig := mockScryptHash()
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{
					createMockDomainTaggedUser(idutil.ULIDNow(), idutil.ULIDNow()),
					createMockDomainTaggedUser(idutil.ULIDNow(), idutil.ULIDNow()),
				}, nil)
				domainTaggedUserRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				domainTaggedUserRepo.On("SoftDelete", ctx, tx, mock.Anything).Return(nil)
				/*firebaseAuth.On("ImportUsers", ctx, mock.Anything).Once().Return(&auth.UserImportResult{}, nil)
				firebaseAuth.On("GetUser", ctx, mock.Anything).Once().Return(nil, nil)
				firebaseAuth.On("UpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)*/
				tx.On("Commit", mock.Anything).Once().Return(nil)
				// bus.On("PublishAsync", mock.Anything, mock.Anything, mock.Anything).Once().Return("", nil)
				jsm.On("TracedPublishAsync", mock.Anything, "nats.TracedPublishAsync", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name: "create parent successfully",
			ctx:  ctx,
			req: &pb.CreateParentsAndAssignToStudentRequest{
				SchoolId:  constants.ManabieSchool,
				StudentId: sampleID,
				ParentProfiles: []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					{
						Name:         existingParentUser.GetDisplayName(),
						CountryCode:  cpb.Country_COUNTRY_VN,
						PhoneNumber:  existingParentUser.GetPhoneNumber(),
						Email:        existingParent.Email.String,
						Username:     "username",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Password:     existingParentUser.GetRawPassword(),
						TagIds:       []string{"tag_id1", "tag_id2", "tag_id3"},
						UserNameFields: &pb.UserNameFields{
							FirstName:         existingParentUser.FirstName.String,
							LastName:          existingParentUser.LastName.String,
							FirstNamePhonetic: existingParentUser.FirstNamePhonetic.String,
							LastNamePhonetic:  existingParentUser.LastNamePhonetic.String,
						},
						ExternalUserId: "random-external-user-id",
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
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				domainUserRepo.On("GetByUserNames", ctx, db, mock.Anything).Once().Return(entity.Users{}, nil)
				domainTagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags(
					[]entity.DomainTag{
						createMockDomainTagWithType("tag_id1", pb.UserTagType_USER_TAG_TYPE_PARENT),
						createMockDomainTagWithType("tag_id2", pb.UserTagType_USER_TAG_TYPE_PARENT),
						createMockDomainTagWithType("tag_id3", pb.UserTagType_USER_TAG_TYPE_PARENT),
					},
				), nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{sampleID})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				userRepo.On("GetByPhone", ctx, tx, mock.Anything).Once().Return(nil, nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				parentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				studentParentRepo.On("InsertParentAccessPathByStudentID", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleParent).Once().Return(&entity.UserGroupV2{}, nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				hashConfig := mockScryptHash()
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				domainTaggedUserRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				domainUserRepo.On("GetByExternalUserIDs", ctx, tx, mock.Anything).Once().Return(entity.Users{}, nil)
				jsm.On("TracedPublishAsync", mock.Anything, "nats.TracedPublishAsync", mock.Anything, mock.Anything).Once().Return("", nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "create parent successfully when UserNameFields is nil and name has value",
			ctx:  ctx,
			req: &pb.CreateParentsAndAssignToStudentRequest{
				SchoolId:  constants.ManabieSchool,
				StudentId: sampleID,
				ParentProfiles: []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					{
						Name:           existingParentUser.GetDisplayName(),
						CountryCode:    cpb.Country_COUNTRY_VN,
						PhoneNumber:    existingParentUser.GetPhoneNumber(),
						Email:          existingParent.Email.String,
						Username:       "username",
						Relationship:   pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Password:       existingParentUser.GetRawPassword(),
						TagIds:         []string{"tag_id1", "tag_id2", "tag_id3"},
						UserNameFields: nil,
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
				userRepo.On("GetByEmailInsensitiveCase", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				domainUserRepo.On("GetByUserNames", ctx, db, mock.Anything).Once().Return(entity.Users{}, nil)
				domainTagRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags(
					[]entity.DomainTag{
						createMockDomainTagWithType("tag_id1", pb.UserTagType_USER_TAG_TYPE_PARENT),
						createMockDomainTagWithType("tag_id2", pb.UserTagType_USER_TAG_TYPE_PARENT),
						createMockDomainTagWithType("tag_id3", pb.UserTagType_USER_TAG_TYPE_PARENT),
					},
				), nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{sampleID})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				userRepo.On("GetByPhone", ctx, tx, mock.Anything).Once().Return(nil, nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				parentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				studentParentRepo.On("InsertParentAccessPathByStudentID", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleParent).Once().Return(&entity.UserGroupV2{}, nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				hashConfig := mockScryptHash()
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				domainTaggedUserRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublishAsync", mock.Anything, "nats.TracedPublishAsync", mock.Anything, mock.Anything).Once().Return("", nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
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

			t.Log("Test case: " + testCase.name)
			testCase.setup(testCase.ctx)

			_, err := s.CreateParentsAndAssignToStudent(testCase.ctx, testCase.req.(*pb.CreateParentsAndAssignToStudentRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}

			mock.AssertExpectationsForObjects(
				t,
				db, tx,
				orgRepo,
				usrEmailRepo,
				userRepo,
				studentRepo,
				parentRepo,
				studentParentRepo,
				userGroupRepo,
				userGroupV2Repo,
				userGroupsMemberRepo,
				userAccessPathRepo,
				firebaseAuthClient,
				tenantManager,
				domainTaggedUserRepo,
				domainTagRepo,
				domainUserRepo,
				unleashClient,
				jsm,
			)
		})
	}
}

func TestUserModifierService_parentPbPhoneNumberToUserPhoneNumber(t *testing.T) {
	t.Parallel()
	_, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	testCases := []TestCase{
		{
			name:         "happy case - empty case",
			req:          &pb.CreateParentsAndAssignToStudentRequest_ParentProfile{},
			expectedErr:  nil,
			expectedResp: entity.UserPhoneNumbers{},
		},
		{
			name: "happy case",
			req: &pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
				ParentPhoneNumbers: []*pb.ParentPhoneNumber{
					{PhoneNumber: "12345678", PhoneNumberType: pb.ParentPhoneNumber_PARENT_PRIMARY_PHONE_NUMBER, PhoneNumberId: "123"},
				},
			},
			expectedErr: nil,
			expectedResp: entity.UserPhoneNumbers{
				{
					ID:              pgtype.Text{String: "123", Status: pgtype.Present},
					UserID:          pgtype.Text{String: "test", Status: pgtype.Present},
					PhoneNumber:     pgtype.Text{String: "12345678", Status: pgtype.Present},
					PhoneNumberType: pgtype.Text{String: "PARENT_PRIMARY_PHONE_NUMBER", Status: pgtype.Present},
					ResourcePath:    pgtype.Text{String: "1", Status: pgtype.Present}},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log(testCase.name)

			results, err := parentPbPhoneNumberToUserPhoneNumber(1, testCase.req.(*pb.CreateParentsAndAssignToStudentRequest_ParentProfile), "test")
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				for index, result := range results {
					expectResult := testCase.expectedResp.(entity.UserPhoneNumbers)[index]
					assert.Equal(t, expectResult.ID, result.ID)
					assert.Equal(t, expectResult.UserID, result.UserID)
					assert.Equal(t, expectResult.PhoneNumber, result.PhoneNumber)
					assert.Equal(t, expectResult.PhoneNumberType, result.PhoneNumberType)
					assert.Equal(t, expectResult.ResourcePath, result.ResourcePath)
				}
			}

		})
	}
}

func TestUserModifierService_ValidateParentPhoneNumber(t *testing.T) {
	t.Parallel()
	_, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	type params struct {
		resourcePath string
	}

	testCases := []TestCase{
		{
			name: "happy case ",
			req: []*pb.ParentPhoneNumber{
				{
					PhoneNumber:     "123214125",
					PhoneNumberType: pb.ParentPhoneNumber_PARENT_PRIMARY_PHONE_NUMBER,
				},
				{
					PhoneNumber:     "123214126",
					PhoneNumberType: pb.ParentPhoneNumber_PARENT_SECONDARY_PHONE_NUMBER,
				},
				{
					PhoneNumber:     "123214127",
					PhoneNumberType: pb.ParentPhoneNumber_PARENT_SECONDARY_PHONE_NUMBER,
				},
			},
			expectedErr: nil,
		},
		{
			name: "create new parent with primary phone number",
			req: []*pb.ParentPhoneNumber{
				{
					PhoneNumber:     "123214125",
					PhoneNumberType: pb.ParentPhoneNumber_PARENT_PRIMARY_PHONE_NUMBER,
				},
			},
			expectedErr: nil,
		},
		{
			name: "create new parent with empty primary phone number ",
			req: []*pb.ParentPhoneNumber{
				{
					PhoneNumber:     "",
					PhoneNumberType: pb.ParentPhoneNumber_PARENT_PRIMARY_PHONE_NUMBER,
				},
			},
			expectedErr: nil,
		},
		{
			name:        "create new parent with nil phone number",
			req:         []*pb.ParentPhoneNumber{},
			expectedErr: nil,
		},
		{
			name: "err when create new parent with wrong type phone number",
			req: []*pb.ParentPhoneNumber{
				{
					PhoneNumber:     "123214125",
					PhoneNumberType: 4,
				},
			},
			expectedErr: errors.New("parent's phone number is wrong type"),
		},
		{
			name: "err when create new parent with duplicate phone number",
			req: []*pb.ParentPhoneNumber{
				{
					PhoneNumber:     "123214125",
					PhoneNumberType: pb.ParentPhoneNumber_PARENT_PRIMARY_PHONE_NUMBER,
				},
				{
					PhoneNumber:     "32141256",
					PhoneNumberType: pb.ParentPhoneNumber_PARENT_SECONDARY_PHONE_NUMBER,
				},
				{
					PhoneNumber:     "123214125",
					PhoneNumberType: pb.ParentPhoneNumber_PARENT_SECONDARY_PHONE_NUMBER,
				},
			},
			expectedErr: errors.New("parent's phone number can not be duplicate"),
		},
		{
			name: "err when create new parent with duplicate type primary phone number",
			req: []*pb.ParentPhoneNumber{
				{
					PhoneNumber:     "123214125",
					PhoneNumberType: pb.ParentPhoneNumber_PARENT_PRIMARY_PHONE_NUMBER,
				},
				{
					PhoneNumber:     "32141256",
					PhoneNumberType: pb.ParentPhoneNumber_PARENT_PRIMARY_PHONE_NUMBER,
				},
			},
			expectedErr: errors.New("parent only need one primary phone number"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			t.Log(testCase.name)

			err := validateParentPhoneNumber(testCase.req.([]*pb.ParentPhoneNumber))

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}

		})
	}
}

func Test_getTagIDsFromParentProfiles(t *testing.T) {
	tests := []struct {
		name string
		args func(t *testing.T) interface{}

		expected []string
	}{
		{
			name: "happy case: create parent",
			args: func(t *testing.T) interface{} {
				return &pb.CreateParentsAndAssignToStudentRequest{
					ParentProfiles: []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
						{TagIds: []string{"id1", "id2", "id3"}},
					},
				}
			},
			expected: []string{"id1", "id2", "id3"},
		},
		{
			name: "happy case: create parent",
			args: func(t *testing.T) interface{} {
				return &pb.CreateParentsAndAssignToStudentRequest{
					ParentProfiles: []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
						{TagIds: []string{"id1", "id2", "id3"}},
						{TagIds: []string{"id4", "id5", "id6"}},
					},
				}
			},
			expected: []string{"id1", "id2", "id3", "id4", "id5", "id6"},
		},
		{
			name: "happy case: update parent",
			args: func(t *testing.T) interface{} {
				return &pb.UpdateParentsAndFamilyRelationshipRequest{
					ParentProfiles: []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
						{TagIds: []string{"id1", "id2", "id3"}},
						{TagIds: []string{"id4", "id5", "id6"}},
					},
				}
			},
			expected: []string{"id1", "id2", "id3", "id4", "id5", "id6"},
		},
		{
			name: "happy case: uknown request",
			args: func(t *testing.T) interface{} {
				return nil
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)
			got := getTagIDsFromParentProfiles(tArgs)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("getTagIDsFromParentProfiles got = %v, expected: %v", got, tt.expected)
			}
		})
	}
}

func Test_validUserNameFieldsCreateParentRequest(t *testing.T) {
	tests := []struct {
		name string
		args func(t *testing.T) *pb.CreateParentsAndAssignToStudentRequest_ParentProfile

		expected error
	}{
		{
			name: "name is empty, UserNameFields is nil",
			args: func(t *testing.T) *pb.CreateParentsAndAssignToStudentRequest_ParentProfile {
				return &pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					UserNameFields: nil,
					Name:           "",
				}
			},
			expected: errors.New("parent name cannot be empty"),
		},
		{
			name: "name is empty, the both first_name is empty and last_name are empty",
			args: func(t *testing.T) *pb.CreateParentsAndAssignToStudentRequest_ParentProfile {
				return &pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					UserNameFields: &pb.UserNameFields{
						FirstName: "",
						LastName:  "",
					},
					Name: "",
				}
			},
			expected: errors.New("parent first_name cannot be empty"),
		},
		{
			name: "name is empty, first_name is empty",
			args: func(t *testing.T) *pb.CreateParentsAndAssignToStudentRequest_ParentProfile {
				return &pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					UserNameFields: &pb.UserNameFields{
						FirstName: "",
						LastName:  "LastName",
					},
					Name: "",
				}
			},
			expected: errors.New("parent first_name cannot be empty"),
		},
		{
			name: "name is empty, last_name is empty",
			args: func(t *testing.T) *pb.CreateParentsAndAssignToStudentRequest_ParentProfile {
				return &pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					UserNameFields: &pb.UserNameFields{
						FirstName: "FirstName",
						LastName:  "",
					},
					Name: "",
				}
			},
			expected: errors.New("parent last_name cannot be empty"),
		},
		{
			name: "name has value, last_name is empty",
			args: func(t *testing.T) *pb.CreateParentsAndAssignToStudentRequest_ParentProfile {
				return &pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					UserNameFields: &pb.UserNameFields{
						FirstName: "FirstName",
						LastName:  "",
					},
					Name: "Name",
				}
			},
			expected: errors.New("parent last_name cannot be empty"),
		},

		{
			name: "name has value, first_name is empty",
			args: func(t *testing.T) *pb.CreateParentsAndAssignToStudentRequest_ParentProfile {
				return &pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					UserNameFields: &pb.UserNameFields{
						FirstName: "",
						LastName:  "LastName",
					},
					Name: "Name",
				}
			},
			expected: errors.New("parent first_name cannot be empty"),
		},
		{
			name: "name has value, first_name and last_name are empty",
			args: func(t *testing.T) *pb.CreateParentsAndAssignToStudentRequest_ParentProfile {
				return &pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					UserNameFields: &pb.UserNameFields{
						FirstName: "",
						LastName:  "",
					},
					Name: "Name",
				}
			},
			expected: errors.New("parent first_name cannot be empty"),
		},
		{
			name: "happy case: name has value, first_name and last_name have value",
			args: func(t *testing.T) *pb.CreateParentsAndAssignToStudentRequest_ParentProfile {
				return &pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					UserNameFields: &pb.UserNameFields{
						FirstName: "FirstName",
						LastName:  "LastName",
					},
					Name: "Name",
				}
			},
			expected: nil,
		},
		{
			name: "happy case: name is empty, first_name and last_name have value",
			args: func(t *testing.T) *pb.CreateParentsAndAssignToStudentRequest_ParentProfile {
				return &pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					UserNameFields: &pb.UserNameFields{
						FirstName: "FirstName",
						LastName:  "LastName",
					},
					Name: "",
				}
			},
			expected: nil,
		},
		{
			name: "happy case: name has value, UserNameFields is nil",
			args: func(t *testing.T) *pb.CreateParentsAndAssignToStudentRequest_ParentProfile {
				return &pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					UserNameFields: nil,
					Name:           "Name",
				}
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)
			got := validUserNameFieldsCreateParentRequest(tArgs)

			if tt.expected != nil && got != nil && got.Error() != tt.expected.Error() {
				t.Errorf("validUserNameFieldsCreateParentRequest got = %v, expected: %v", got, tt.expected)
			}
			if tt.expected != nil && got == nil {
				t.Errorf("validUserNameFieldsCreateParentRequest got = %v, expected: %v", got, tt.expected)
			}
			if tt.expected == nil && got != nil {
				t.Errorf("validUserNameFieldsCreateParentRequest got = %v, expected: %v", got, tt.expected)
			}

		})
	}
}

func Test_createParentProfileToParentDomain(t *testing.T) {
	type args struct {
		profile *pb.CreateParentsAndAssignToStudentRequest_ParentProfile
	}
	tests := []struct {
		name string
		args args
		want entity.User
	}{
		{
			name: "full profile",
			args: args{
				profile: &pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					Email:    "Email",
					Username: "Username",
					UserNameFields: &pb.UserNameFields{
						FirstName: "FirstName",
						LastName:  "LastName",
					},
				},
			},
			want: entity.ParentWillBeDelegated{
				DomainParentProfile: &repository.User{
					UserNameAttr:  field.NewString("Username"),
					EmailAttr:     field.NewString("Email"),
					FirstNameAttr: field.NewString("FirstName"),
					LastNameAttr:  field.NewString("LastName"),
				},
				HasUserID: &repository.User{
					ID: field.NewNullString(),
				},
			},
		},
		{
			name: "without first_name and last_name",
			args: args{
				profile: &pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
					Name:     "LastName FirtName",
					Email:    "Email",
					Username: "Username",
				},
			},
			want: entity.ParentWillBeDelegated{
				DomainParentProfile: &repository.User{
					UserNameAttr:  field.NewString("Username"),
					EmailAttr:     field.NewString("Email"),
					FirstNameAttr: field.NewString("FirtName"),
					LastNameAttr:  field.NewString("LastName"),
				},
				HasUserID: &repository.User{
					ID: field.NewNullString(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := createParentProfileToParentDomain(tt.args.profile)
			assert.Equal(t, tt.want.FirstName(), u.FirstName())
			assert.Equal(t, tt.want.LastName(), u.LastName())
			assert.Equal(t, tt.want.Email(), u.Email())
			assert.Equal(t, tt.want.UserName(), u.UserName())
		})
	}
}

func Test_updateParentProfileToParentDomain(t *testing.T) {
	type args struct {
		profile *pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile
	}
	tests := []struct {
		name string
		args args
		want entity.User
	}{
		{
			name: "full profile",
			args: args{
				profile: &pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					Email:    "Email",
					Username: "Username",
					UserNameFields: &pb.UserNameFields{
						FirstName: "FirstName",
						LastName:  "LastName",
					},
				},
			},
			want: entity.ParentWillBeDelegated{
				DomainParentProfile: &repository.User{
					UserNameAttr:  field.NewString("Username"),
					EmailAttr:     field.NewString("Email"),
					FirstNameAttr: field.NewString("FirstName"),
					LastNameAttr:  field.NewString("LastName"),
				},
				HasUserID: &repository.User{
					ID: field.NewNullString(),
				},
			},
		},
		{
			name: "without first_name and last_name",
			args: args{
				profile: &pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
					Email:    "Email",
					Username: "Username",
				},
			},
			want: entity.ParentWillBeDelegated{
				DomainParentProfile: &repository.User{
					UserNameAttr:  field.NewString("Username"),
					EmailAttr:     field.NewString("Email"),
					FirstNameAttr: field.NewNullString(),
					LastNameAttr:  field.NewNullString(),
				},
				HasUserID: &repository.User{
					ID: field.NewNullString(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := updateParentProfileToParentDomain(tt.args.profile)
			assert.Equal(t, tt.want.FirstName(), u.FirstName())
			assert.Equal(t, tt.want.LastName(), u.LastName())
			assert.Equal(t, tt.want.Email(), u.Email())
			assert.Equal(t, tt.want.UserName(), u.UserName())
		})
	}
}
