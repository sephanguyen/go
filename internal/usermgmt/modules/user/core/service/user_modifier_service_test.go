package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/auth"
	internal_auth_user "github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/gcp"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
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
	mock_locationRepo "github.com/manabie-com/backend/mock/mastermgmt/modules/location/infrastructure/repo"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	aValidTenantID       = "valid-tenant-id"
	validPhoneNumber     = "0987678765"
	validHomePhoneNumber = "0912345632"
	invalidPhoneNumber   = "invalid-0987632"
	stubLocationID       = "location-id"
)

type mockDomainTaggedUser struct {
	userID field.String
	tagID  field.String
	entity.EmptyDomainTaggedUser
}

func createMockDomainTaggedUser(userID string, tagID string) entity.DomainTaggedUser {
	return &mockDomainTaggedUser{
		userID: field.NewString(userID),
		tagID:  field.NewString(tagID),
	}
}

func (m *mockDomainTaggedUser) UserID() field.String {
	return m.userID
}

func (m *mockDomainTaggedUser) TagID() field.String {
	return m.tagID
}

type MockDomainTag struct {
	tagID             field.String
	partnerInternalID field.String
	tagType           field.String

	entity.EmptyDomainTag
}

func createMockDomainTagWithType(tagID string, tagType pb.UserTagType) entity.DomainTag {
	return &MockDomainTag{
		tagID:             field.NewString(tagID),
		partnerInternalID: field.NewString(fmt.Sprintf("partner-id-%s", tagID)),
		tagType:           field.NewString(tagType.String()),
	}
}

func createMockDomainTag(tagID string) entity.DomainTag {
	return createMockDomainTagWithType(tagID, pb.UserTagType_USER_TAG_TYPE_STUDENT)
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

type MockDomainUser struct {
	userID       field.String
	UsernameAttr field.String
	entity.EmptyUser
}

func createMockDomainUser(userID string) entity.User {
	return &MockDomainUser{
		userID: field.NewString(userID),
	}
}

func (m *MockDomainUser) UserID() field.String {
	return m.userID
}
func (m *MockDomainUser) UserName() field.String {
	return m.UsernameAttr
}

func mockScryptHash() *gcp.HashConfig {
	return &gcp.HashConfig{
		HashAlgorithm:  "SCRYPT",
		HashRounds:     8,
		HashMemoryCost: 8,
		HashSaltSeparator: gcp.Base64EncodedStr{
			Value:        "salt",
			DecodedBytes: []byte("salt"),
		},
		HashSignerKey: gcp.Base64EncodedStr{
			Value:        "key",
			DecodedBytes: []byte("key"),
		},
	}
}

func TestCreateStudent(t *testing.T) {
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
	userGroupV2Repo := new(mock_repositories.MockUserGroupV2Repo)
	userGroupsMemberRepo := new(mock_repositories.MockUserGroupsMemberRepo)
	userAccessPathRepo := new(mock_repositories.MockUserAccessPathRepo)
	locationRepo := new(mock_locationRepo.MockLocationRepo)
	jsm := new(mock_nats.JetStreamManagement)
	firebaseAuth := new(mock_firebase.AuthClient)
	tenantManager := new(mock_multitenant.TenantManager)
	firebaseAuthClient := new(mock_multitenant.TenantClient)
	schoolHistoryRepo := new(mock_repositories.MockSchoolHistoryRepo)
	schoolInfoRepo := new(mock_repositories.MockSchoolInfoRepo)
	schoolCourseRepo := new(mock_repositories.MockSchoolCourseRepo)
	userAddressRepo := new(mock_repositories.MockUserAddressRepo)
	prefectureRepo := new(mock_repositories.MockPrefectureRepo)
	userPhoneNumberRepo := new(mock_repositories.MockUserPhoneNumberRepo)
	domainGradeRepo := new(mock_repositories.MockDomainGradeRepo)
	gradeOrganizationRepo := new(mock_repositories.MockGradeOrganizationRepo)
	domainTagRepo := new(mock_repositories.MockDomainTagRepo)
	domainTaggedUserRepo := new(mock_repositories.MockDomainTaggedUserRepo)

	s := UserModifierService{
		DB:                    db,
		OrganizationRepo:      orgRepo,
		UsrEmailRepo:          usrEmailRepo,
		UserRepo:              userRepo,
		StudentRepo:           studentRepo,
		UserGroupRepo:         userGroupRepo,
		UserGroupV2Repo:       userGroupV2Repo,
		UserGroupsMemberRepo:  userGroupsMemberRepo,
		UserAccessPathRepo:    userAccessPathRepo,
		LocationRepo:          locationRepo,
		FirebaseClient:        firebaseAuth,
		FirebaseAuthClient:    firebaseAuthClient,
		TenantManager:         tenantManager,
		JSM:                   jsm,
		SchoolHistoryRepo:     schoolHistoryRepo,
		SchoolInfoRepo:        schoolInfoRepo,
		SchoolCourseRepo:      schoolCourseRepo,
		UserAddressRepo:       userAddressRepo,
		PrefectureRepo:        prefectureRepo,
		UserPhoneNumberRepo:   userPhoneNumberRepo,
		DomainGradeRepo:       domainGradeRepo,
		GradeOrganizationRepo: gradeOrganizationRepo,
		DomainTagRepo:         domainTagRepo,
		DomainTaggedUserRepo:  domainTaggedUserRepo,
	}

	testCases := []TestCase{
		{
			name: "cannot create if student info have invalid params",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Grade: 1,
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "student email cannot be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "error when finding tag by ids",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:          "user's password",
					FirstName:         "First",
					LastName:          "Last",
					FirstNamePhonetic: "First Phonetic",
					LastNamePhonetic:  "Last Phonetic",
					Email:             "email@example.com",
					Grade:             1,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
					TagIds:            []string{"tag_id1", "tag_id2"},
				},
			},
			expectedErr: status.Error(codes.Internal, errors.Wrap(fmt.Errorf("error"), "DomainTagRepo.GetByIDs").Error()),
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, fmt.Errorf("error"))
			},
		},
		{
			name: "error when passing unexisted tag ids",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:          "user's password",
					FirstName:         "First",
					LastName:          "Last",
					FirstNamePhonetic: "First Phonetic",
					LastNamePhonetic:  "Last Phonetic",
					Email:             "email@example.com",
					Grade:             1,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
					TagIds:            []string{idutil.ULIDNow(), idutil.ULIDNow()},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, ErrTagIDsMustBeExisted.Error()),
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
			},
		},
		{
			name: "cannot create if student emails already exist in db",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					Name:             "user's name",
					Email:            "existing-email@example",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					LocationIds:      []string{stubLocationID},
				},
			},
			expectedErr: status.Error(codes.AlreadyExists, fmt.Sprintf("cannot create student with emails existing in system: %s", "existing-email@example")),
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, database.TextArray([]string{"existing-email@example"})).Once().Return([]*entity.LegacyUser{{}}, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "cannot create if student phone number already exist in db",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Email:             "existing-email@example",
					Password:          "user's password",
					Name:              "user's name",
					PhoneNumber:       "existing-phone-number",
					Grade:             1,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					LocationIds:       []string{stubLocationID},
				},
			},
			expectedErr: status.Error(codes.AlreadyExists, fmt.Sprintf("cannot create student with phone number existing in system: %s", "existing-phone-number")),
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, database.TextArray([]string{"existing-email@example"})).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, tx, database.TextArray([]string{"existing-phone-number"})).Once().Return([]*entity.LegacyUser{{}}, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "cannot create student if UserRepo.GetByEmail cannot find user by email ",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:          "user's password",
					Name:              "user's name",
					Email:             "email@example.com",
					Grade:             1,
					CountryCode:       cpb.Country_COUNTRY_VN,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					PhoneNumber:       "12312",
					LocationIds:       []string{stubLocationID},
				},
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("s.UserRepo.GetByEmail: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "cannot create student if UserRepo.GetByPhone cannot find user by phone ",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:          "user's password",
					Name:              "user's name",
					Email:             "email@example.com",
					Grade:             1,
					CountryCode:       cpb.Country_COUNTRY_VN,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					PhoneNumber:       "12312",
					LocationIds:       []string{"location_1"},
				},
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("s.UserRepo.GetByPhone: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, tx, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()

				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "cannot create student if StudentRepo.Create fail",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:          "user's password",
					Name:              "user's name",
					Email:             "email@example.com",
					Grade:             1,
					CountryCode:       cpb.Country_COUNTRY_VN,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					PhoneNumber:       "12312",
					LocationIds:       []string{"location_1"},
				},
			},
			expectedErr: status.Error(codes.Unknown, fmt.Errorf("create student fail").Error()),
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(errors.New("create student fail"))
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()

				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "cannot create student if userGroupsMemberRepo.AssignWithUserGroup fail",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:          "user's password",
					Name:              "user's name",
					Email:             "email@example.com",
					Grade:             1,
					CountryCode:       cpb.Country_COUNTRY_VN,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					PhoneNumber:       "12312",
					LocationIds:       []string{"location_1"},
				},
			},
			expectedErr: status.Error(codes.Internal, errors.Wrapf(fmt.Errorf("error"), "can not assign student user group to user %s", "example-id").Error()),
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()

				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "cannot create student if email empty",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:          "user's password",
					Name:              "user's name",
					Grade:             1,
					CountryCode:       cpb.Country_COUNTRY_VN,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					PhoneNumber:       "12312",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("student email cannot be empty").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "cannot create student if password empty",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Name:              "user's name",
					Email:             "example@gmail.com",
					Grade:             1,
					CountryCode:       cpb.Country_COUNTRY_VN,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					PhoneNumber:       "12312",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("student password cannot be empty").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "cannot create student if country is invalid",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:          "P@ssw0rd",
					Name:              "user's name",
					Email:             "example@gmail.com",
					Grade:             1,
					CountryCode:       cpb.Country(-1),
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					PhoneNumber:       "12312",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("student country code is not valid").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "cannot create student if password is too short",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:          "Ps",
					Name:              "user's name",
					Email:             "example@gmail.com",
					Grade:             1,
					CountryCode:       cpb.Country_COUNTRY_VN,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					PhoneNumber:       "12312",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("student password length should be at least 6").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "cannot create student if student name empty",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:          "P@ssw0rd",
					Email:             "example@gmail.com",
					Grade:             1,
					CountryCode:       cpb.Country_COUNTRY_VN,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					PhoneNumber:       "12312",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("student name cannot be empty").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "cannot create student if student name and first name empty",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:          "P@ssw0rd",
					Email:             "example@gmail.com",
					Grade:             1,
					CountryCode:       cpb.Country_COUNTRY_VN,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					PhoneNumber:       "12312",
					LastName:          "Last name",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("student first name cannot be empty").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "cannot create student if student name and last name empty",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:          "P@ssw0rd",
					Email:             "example@gmail.com",
					Grade:             1,
					CountryCode:       cpb.Country_COUNTRY_VN,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					PhoneNumber:       "12312",
					FirstName:         "First",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("student last name cannot be empty").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "create student success with first name and last name",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:          "user's password",
					FirstName:         "First",
					LastName:          "Last",
					FirstNamePhonetic: "First Phonetic",
					LastNamePhonetic:  "Last Phonetic",
					Email:             "email@example.com",
					Grade:             1,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				hashConfig := mockScryptHash()
				tenantClient := &mock_multitenant.TenantClient{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUserEvent", mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
		},
		{
			name: "create student success with first name and last name and first name phonetic",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:          "user's password",
					FirstName:         "First",
					LastName:          "Last",
					FirstNamePhonetic: "First Phonetic",
					Email:             "email@example.com",
					Grade:             1,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				hashConfig := mockScryptHash()
				tenantClient := &mock_multitenant.TenantClient{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUserEvent", mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
		},
		{
			name: "create student success with first name and last name and last name phonetic",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					FirstName:        "First",
					LastName:         "Last",
					LastNamePhonetic: "Last Phonetic",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:         timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:           pb.Gender_MALE,
					LocationIds:      []string{"location_1"},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				hashConfig := mockScryptHash()
				tenantClient := &mock_multitenant.TenantClient{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUserEvent", mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
		},
		{
			name: "failed to create student because cannot get tenant id by organization id",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					FirstName:        "First",
					LastName:         "Last",
					LastNamePhonetic: "Last Phonetic",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:         timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:           pb.Gender_MALE,
					LocationIds:      []string{"location_1"},
				},
			},
			expectedErr: status.Error(codes.FailedPrecondition, errcode.TenantDoesNotExistErr{OrganizationID: fmt.Sprint(constants.ManabieSchool)}.Error()),
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", pgx.ErrNoRows)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "failed to create student because tenant manager cannot get tenant client",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					FirstName:        "First",
					LastName:         "Last",
					LastNamePhonetic: "Last Phonetic",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:         timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:           pb.Gender_MALE,
					LocationIds:      []string{"location_1"},
				},
			},
			expectedErr: status.Error(codes.Internal, errors.Wrap(internal_auth_user.ErrTenantNotFound, "TenantClient").Error()),
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return(aValidTenantID, nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(nil, internal_auth_user.ErrTenantNotFound)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "cannot create if student status data unknown",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					Name:             "user's name",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus(999999), // some unknown status
					LocationIds:      []string{stubLocationID},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, ErrStudentEnrollmentStatusUnknown.Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "cannot create if student status data is STUDENT_STATUS_NONE",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					Name:             "user's name",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NONE,
					LocationIds:      []string{stubLocationID},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, ErrStudentEnrollmentStatusNotAllowedTobeNone.Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "cannot create student with locationIds has empty element",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					Name:             "user's name",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:         timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:           pb.Gender_MALE,
					LocationIds:      []string{""},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "getLocations invalid params: location_id empty"),
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "cannot create student with locationIds has invalid resource_path",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					Name:             "user's name",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:         timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:           pb.Gender_MALE,
					LocationIds:      []string{"location_invalid"},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("getLocations fail: resource path invalid, expect %d, but actual ", constants.ManabieSchool).Error()),
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Once().Return([]*domain.Location{
					{
						LocationID: stubLocationID,
						Name:       "center",
					},
				}, nil).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "create student success without student user group",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					Name:             "user's name",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:         timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:           pb.Gender_MALE,
					LocationIds:      []string{stubLocationID},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleStudent).Once().Return(nil, fmt.Errorf("error"))
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				hashConfig := mockScryptHash()
				tenantClient := &mock_multitenant.TenantClient{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Once().Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUserEvent", mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name: "create student success with locationIds nil",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					Name:             "user's name",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:         timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:           pb.Gender_MALE,
					LocationIds:      []string{stubLocationID},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				hashConfig := mockScryptHash()
				tenantClient := &mock_multitenant.TenantClient{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Once().Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUserEvent", mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name: "create student with only student data successfully",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					Name:             "user's name",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:         timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:           pb.Gender_MALE,
					LocationIds:      []string{"location_1"},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				hashConfig := mockScryptHash()
				tenantClient := &mock_multitenant.TenantClient{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUserEvent", mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
		},
		{
			name: "create student with one school history successfully",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					Name:             "user's name",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:         timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:           pb.Gender_MALE,
					LocationIds:      []string{"location_1"},
				},
				SchoolHistories: []*pb.SchoolHistory{
					{
						SchoolId:       uuid.NewString(),
						SchoolCourseId: uuid.NewString(),
						StartDate:      timestamppb.Now(),
						EndDate:        timestamppb.Now(),
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				hashConfig := mockScryptHash()
				tenantClient := &mock_multitenant.TenantClient{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{
					{
						ID: database.Text("school_info-id-1"),
					},
				}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{
					{
						ID: database.Text("school_course-id-1"),
					},
				}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUserEvent", mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
		},
		{
			name: "create student with many school histories successfully",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					Name:             "user's name",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:         timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:           pb.Gender_MALE,
					LocationIds:      []string{"location_1"},
				},
				SchoolHistories: []*pb.SchoolHistory{
					{
						SchoolId:  uuid.NewString(),
						StartDate: timestamppb.Now(),
						EndDate:   timestamppb.Now(),
					},
					{
						SchoolId:       uuid.NewString(),
						SchoolCourseId: uuid.NewString(),
						EndDate:        timestamppb.Now(),
					},
					{
						SchoolId:       uuid.NewString(),
						SchoolCourseId: uuid.NewString(),
						StartDate:      timestamppb.Now(),
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				hashConfig := mockScryptHash()
				tenantClient := &mock_multitenant.TenantClient{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{
					{
						ID:      database.Text("school_info-id-1"),
						LevelID: database.Text("school_level-id-1"),
					},
					{
						ID:      database.Text("school_info-id-2"),
						LevelID: database.Text("school_level-id-2"),
					},
					{
						ID:      database.Text("school_info-id-3"),
						LevelID: database.Text("school_level-id-3"),
					},
				}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{
					{
						ID: database.Text("school_course-id-1"),
					},
					{
						ID: database.Text("school_course-id-2"),
					},
				}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUserEvent", mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
		},
		{
			name: "create student with school history and mandatory only",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					Name:             "user's name",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:         timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:           pb.Gender_MALE,
					LocationIds:      []string{"location_1"},
				},
				SchoolHistories: []*pb.SchoolHistory{
					{
						SchoolId: uuid.NewString(),
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				hashConfig := mockScryptHash()
				tenantClient := &mock_multitenant.TenantClient{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{
					{
						ID: database.Text("school_info-id-1"),
					},
				}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUserEvent", mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
		},
		{
			name: "create student failed with school history: missing mandatory school_id",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					Name:             "user's name",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:         timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:           pb.Gender_MALE,
					LocationIds:      []string{"location_1"},
				},
				SchoolHistories: []*pb.SchoolHistory{
					{
						SchoolCourseId: uuid.NewString(),
						StartDate:      timestamppb.Now(),
						EndDate:        timestamppb.Now(),
					},
				},
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("validateSchoolHistoriesReq: %v", fmt.Errorf("school_id cannot be empty at row: %v", 1)).Error()),
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				hashConfig := mockScryptHash()
				tenantClient := &mock_multitenant.TenantClient{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUserEvent", mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
		},
		{
			name: "create student with one home address successfully",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					Name:             "user's name",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:         timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:           pb.Gender_MALE,
					LocationIds:      []string{"location_1"},
				},
				UserAddresses: []*pb.UserAddress{
					{
						AddressId:   uuid.NewString(),
						AddressType: pb.AddressType_HOME_ADDRESS,
						PostalCode:  "postal-code-1",
						Prefecture:  "ID-test",
						City:        "city-1",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				hashConfig := mockScryptHash()
				tenantClient := &mock_multitenant.TenantClient{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				prefectureRepo.On("GetByPrefectureID", ctx, mock.Anything, database.Text("ID-test")).Return(&entity.Prefecture{
					ID:             database.Text("ID-test"),
					PrefectureCode: database.Text("01"),
					Name:           database.Text("name-01"),
				}, nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUserEvent", mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
		},
		{
			name: "create student with many home addresses successfully",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					Name:             "user's name",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:         timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:           pb.Gender_MALE,
					LocationIds:      []string{"location_1"},
				},
				UserAddresses: []*pb.UserAddress{
					{
						AddressId:   uuid.NewString(),
						AddressType: pb.AddressType_HOME_ADDRESS,
						PostalCode:  "postal-code-1",
						Prefecture:  "ID-test-01",
						City:        "city-1",
					},
					{
						AddressId:   uuid.NewString(),
						AddressType: pb.AddressType_HOME_ADDRESS,
						PostalCode:  "postal-code-2",
						Prefecture:  "ID-test-02",
						City:        "city-2",
					},
					{
						AddressId:   uuid.NewString(),
						AddressType: pb.AddressType_HOME_ADDRESS,
						PostalCode:  "postal-code-2",
						City:        "city-2",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				hashConfig := mockScryptHash()
				tenantClient := &mock_multitenant.TenantClient{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				prefectureRepo.On("GetByPrefectureID", ctx, mock.Anything, database.Text("ID-test-01")).Return(&entity.Prefecture{
					ID:             database.Text("ID-test-01"),
					PrefectureCode: database.Text("01"),
					Name:           database.Text("name-01"),
				}, nil)
				prefectureRepo.On("GetByPrefectureID", ctx, mock.Anything, database.Text("ID-test-02")).Return(&entity.Prefecture{
					ID:             database.Text("ID-test-02"),
					PrefectureCode: database.Text("02"),
					Name:           database.Text("name-02"),
				}, nil)
				prefectureRepo.On("GetByPrefectureID", ctx, mock.Anything, database.Text("")).Return(&entity.Prefecture{}, nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUserEvent", mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
		},
		{
			name: "create student with home address and mandatory only",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					Name:             "user's name",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:         timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:           pb.Gender_MALE,
					LocationIds:      []string{"location_1"},
				},
				UserAddresses: []*pb.UserAddress{
					{
						AddressId:   uuid.NewString(),
						AddressType: pb.AddressType_HOME_ADDRESS,
						Prefecture:  "ID-test-01",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				hashConfig := mockScryptHash()
				tenantClient := &mock_multitenant.TenantClient{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				prefectureRepo.On("GetByPrefectureID", ctx, mock.Anything, database.Text("ID-test-01")).Return(&entity.Prefecture{
					ID:             database.Text("ID-test-01"),
					PrefectureCode: database.Text("03"),
					Name:           database.Text("name-03"),
				}, nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUserEvent", mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
		},
		{
			name: "create student failed with home address: invalid mandatory address_type",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					Name:             "user's name",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:         timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:           pb.Gender_MALE,
					LocationIds:      []string{"location_1"},
				},
				UserAddresses: []*pb.UserAddress{
					{
						AddressId:   uuid.NewString(),
						AddressType: pb.AddressType_BILLING_ADDRESS,
					},
				},
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("validateUserAddressesReq: %v", fmt.Errorf("address_type cannot be other type, must be HOME_ADDRESS: %v", 1)).Error()),
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				hashConfig := mockScryptHash()
				tenantClient := &mock_multitenant.TenantClient{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUserEvent", mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
		},
		{
			name: "create student success with block student phone number",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					Name:             "user's name",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:         timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:           pb.Gender_MALE,
					LocationIds:      []string{"location_1"},
					StudentPhoneNumber: &pb.StudentPhoneNumber{
						PhoneNumber:       validPhoneNumber,
						HomePhoneNumber:   validHomePhoneNumber,
						ContactPreference: pb.StudentContactPreference_STUDENT_PHONE_NUMBER,
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				hashConfig := mockScryptHash()
				tenantClient := &mock_multitenant.TenantClient{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUserEvent", mock.Anything, mock.Anything).Twice().Return(nil, nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
			},
		},
		{
			name: "create student error invalid phone number with block student phone number",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					Name:             "user's name",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:         timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:           pb.Gender_MALE,
					LocationIds:      []string{"location_1"},
					StudentPhoneNumber: &pb.StudentPhoneNumber{
						PhoneNumber:       invalidPhoneNumber,
						ContactPreference: pb.StudentContactPreference_STUDENT_PHONE_NUMBER,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "validateStudentPhoneNumber: error regexp.MatchString: doesn't match"),
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				hashConfig := mockScryptHash()
				tenantClient := &mock_multitenant.TenantClient{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUserEvent", mock.Anything, mock.Anything).Twice().Return(nil, nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
			},
		},
		{
			name: "create student error invalid home phone number with block student phone number",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					Name:             "user's name",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:         timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:           pb.Gender_MALE,
					LocationIds:      []string{"location_1"},
					StudentPhoneNumber: &pb.StudentPhoneNumber{
						HomePhoneNumber:   invalidPhoneNumber,
						ContactPreference: pb.StudentContactPreference_STUDENT_PHONE_NUMBER,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "validateStudentPhoneNumber: error regexp.MatchString: doesn't match"),
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				hashConfig := mockScryptHash()
				tenantClient := &mock_multitenant.TenantClient{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUserEvent", mock.Anything, mock.Anything).Twice().Return(nil, nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
			},
		},
		{
			name: "create student success with grade master",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:          "user's password",
					FirstName:         "First",
					LastName:          "Last",
					FirstNamePhonetic: "First Phonetic",
					LastNamePhonetic:  "Last Phonetic",
					Email:             "email@example.com",
					GradeId:           "grade-id-1",
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				grades := []entity.DomainGrade{&mock_usermgmt.Grade{}}
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(grades, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				gradeOrganizationRepo.On("GetByGradeIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				hashConfig := mockScryptHash()
				tenantClient := &mock_multitenant.TenantClient{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUserEvent", mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
		},
		{
			name: "create student failed with grade master",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:          "user's password",
					FirstName:         "First",
					LastName:          "Last",
					FirstNamePhonetic: "First Phonetic",
					LastNamePhonetic:  "Last Phonetic",
					Email:             "email@example.com",
					GradeId:           "grade-id-1",
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
				},
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "grade_id does not exist"),
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUserEvent", mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
		},
		{
			name: "happy case with isStudentLocationFlagEnabled error",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:          "user's password",
					FirstName:         "First",
					LastName:          "Last",
					FirstNamePhonetic: "First Phonetic",
					LastNamePhonetic:  "Last Phonetic",
					Email:             "email@example.com",
					Grade:             1,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
					TagIds:            []string{"tag_id1", "tag_id2"},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags([]entity.DomainTag{createMockDomainTag("tag_id1"), createMockDomainTag("tag_id2")}), nil)
				usrEmail := &entity.UsrEmail{UsrID: database.Text("example-id")}
				usrEmailRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Once().Return(usrEmail, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   stubLocationID,
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				hashConfig := mockScryptHash()
				tenantClient := &mock_multitenant.TenantClient{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				domainTaggedUserRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				jsm.On("TracedPublish", mock.Anything, "publishUserEvent", mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
		},
		// {
		// 	name: "create student failed with grade master archived",
		// 	ctx:  ctx,
		// 	req: &pb.CreateStudentRequest{
		// 		StudentProfile: &pb.CreateStudentRequest_StudentProfile{
		// 			Password:          "user's password",
		// 			FirstName:         "First",
		// 			LastName:          "Last",
		// 			FirstNamePhonetic: "First Phonetic",
		// 			LastNamePhonetic:  "Last Phonetic",
		// 			Email:             "email@example.com",
		// 			GradeId:           "grade-id-1",
		// 			EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
		// 			Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
		// 			Gender:            pb.Gender_MALE,
		// 			LocationIds:       []string{"location_1"},
		// 		},
		// 	},
		// 	expectedErr: status.Errorf(codes.InvalidArgument, "grade is archived"),
		// 	setup: func(ctx context.Context) {
		// 		grade := &mock_usermgmt.Grade{}
		// 		grade.RandomGrade.IsArchived = field.NewBoolean(true)
		// 		grades := []entity.DomainGrade{grade}
		// 		db.On("Begin", mock.Anything).Once().Return(tx, nil)
		// 		domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(grades, nil)
		// 		domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
		// 		locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
		// 			{
		// 				LocationID:   stubLocationID,
		// 				Name:         "center",
		// 				ResourcePath: fmt.Sprint(constants.ManabieSchool),
		// 			},
		// 		}, nil).Once()
		// 		tx.On("Commit", mock.Anything).Once().Return(nil)
		// 		jsm.On("TracedPublish", mock.Anything, "publishUserEvent", mock.Anything, mock.Anything).Twice().Return(nil, nil)
		// 	},
		// },
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

			_, err := s.CreateStudent(testCase.ctx, testCase.req.(*pb.CreateStudentRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, studentRepo, userRepo, userGroupRepo, userAccessPathRepo, locationRepo, firebaseAuth, userGroupV2Repo, userGroupsMemberRepo, domainTagRepo)
		})
	}
}

func TestUserModifierService_GenerateImportParentsAndAssignToStudentTemplate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	unleashClient := new(mock_unleash_client.UnleashClientInstance)

	templateImportParentStableHeaders := "external_user_id,last_name,first_name,last_name_phonetic,first_name_phonetic,email,student_email,relationship,parent_tag,primary_phone_number,secondary_phone_number,remarks"
	templateImportParentStableValues := "externaluserid,parent last name,parent first name,phonetic name,phonetic name,parent@email.com,student1@email.com;student2@email.com,1;2,tag_partner_id_1;tag_partner_id_2,parent_primary_phone_number,parent_secondary_phone_number,parent-remarks"

	templateImportParentWithUserNameHeaders := "external_user_id,username,last_name,first_name,last_name_phonetic,first_name_phonetic,email,student_email,relationship,parent_tag,primary_phone_number,secondary_phone_number,remarks"
	templateImportParentWithUserNameValues := "externaluserid,username,parent last name,parent first name,phonetic name,phonetic name,parent@email.com,student1@email.com;student2@email.com,1;2,tag_partner_id_1;tag_partner_id_2,parent_primary_phone_number,parent_secondary_phone_number,parent-remarks"

	s := &UserModifierService{UnleashClient: unleashClient}

	testCases := []TestCase{
		{
			name:         "happy case: stable template",
			ctx:          ctx,
			expectedErr:  nil,
			req:          &pb.GenerateImportParentsAndAssignToStudentTemplateRequest{},
			expectedResp: templateImportParentStableHeaders + "\n" + templateImportParentStableValues,
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(false, nil)
			},
		},
		{
			name:         "happy case: template with username",
			ctx:          ctx,
			expectedErr:  nil,
			req:          &pb.GenerateImportParentsAndAssignToStudentTemplateRequest{},
			expectedResp: templateImportParentWithUserNameHeaders + "\n" + templateImportParentWithUserNameValues,
			setup: func(ctx context.Context) {
				unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleUserNameStudentParent, mock.Anything, mock.Anything).Once().Return(true, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{Manabie: &interceptors.ManabieClaims{ResourcePath: fmt.Sprint(constants.ManabieSchool)}}
			testCase.ctx = interceptors.ContextWithJWTClaims(ctx, claim)

			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}

			resp, err := s.GenerateImportParentsAndAssignToStudentTemplate(testCase.ctx, testCase.req.(*pb.GenerateImportParentsAndAssignToStudentTemplateRequest))
			assert.Equal(t, testCase.expectedErr, err)

			data, err := base64.StdEncoding.DecodeString(string(resp.Data))
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, string(data))
		})
	}
}

func TestUserModifierService_ValidCreateRequest(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		unleash     bool
		req         interface{}
		expectedErr error
	}{
		{
			name:        "err when input create student with nil profile",
			expectedErr: errors.New("student profile is null"),
			req:         &pb.CreateStudentRequest{StudentProfile: nil},
			unleash:     true,
		},
		{
			name:        "err when input create student miss locations ID",
			expectedErr: errors.New("student location length must be at least 1"),
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					FirstName:        "First",
					LastName:         "Last",
					LastNamePhonetic: "Last Phonetic",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:         timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:           pb.Gender_MALE,
				},
			},
			unleash: true,
		},
		{
			name:        "happy case",
			expectedErr: nil,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					FirstName:        "First",
					LastName:         "Last",
					LastNamePhonetic: "Last Phonetic",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:         timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:           pb.Gender_MALE,
					LocationIds:      []string{"location_1"},
				},
			},
			unleash: true,
		},
		{
			name:        "happy case with no location - unleash location false",
			expectedErr: nil,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					FirstName:        "First",
					LastName:         "Last",
					LastNamePhonetic: "Last Phonetic",
					Email:            "email@example.com",
					Grade:            1,
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Birthday:         timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:           pb.Gender_MALE,
					LocationIds:      []string{"location_1"},
				},
			},
			unleash: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validCreateRequest(testCase.req.(*pb.CreateStudentRequest))

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestUserModifierService_validateSchoolHistoriesReq(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db := new(mock_database.Ext)
	schoolInfoRepo := new(mock_repositories.MockSchoolInfoRepo)
	schoolCourseRepo := new(mock_repositories.MockSchoolCourseRepo)

	s := &UserModifierService{
		DB:               db,
		SchoolInfoRepo:   schoolInfoRepo,
		SchoolCourseRepo: schoolCourseRepo,
	}

	testCases := []TestCase{
		{
			name:        "missing mandatory field",
			ctx:         ctx,
			expectedErr: fmt.Errorf("school_id cannot be empty at row: %v", 1),
			req: []*pb.SchoolHistory{
				{
					SchoolCourseId: uuid.NewString(),
					StartDate:      timestamppb.Now(),
					EndDate:        timestamppb.Now(),
				},
			},
		},
		{
			name:        "invalid school_info",
			ctx:         ctx,
			expectedErr: fmt.Errorf("school_info does not match with req.SchoolHistories"),
			req: []*pb.SchoolHistory{
				{
					SchoolId:       uuid.NewString(),
					SchoolCourseId: uuid.NewString(),
					StartDate:      timestamppb.Now(),
					EndDate:        timestamppb.Now(),
				},
			},
			setup: func(ctx context.Context) {
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
			},
		},
		{
			name:        "invalid school_info: duplicate level id",
			ctx:         ctx,
			expectedErr: fmt.Errorf("duplicate school_level_id in school_info school_info-id-1"),
			req: []*pb.SchoolHistory{
				{
					SchoolId: uuid.NewString(),
				},
				{
					SchoolId: uuid.NewString(),
				},
			},
			setup: func(ctx context.Context) {
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{
					{
						ID:      database.Text("school_info-id-1"),
						LevelID: database.Text("school_level-id-1"),
					},
					{
						ID:      database.Text("school_info-id-1"),
						LevelID: database.Text("school_level-id-1"),
					},
				}, nil)
			},
		},
		{
			name:        "invalid start_date and end_date",
			ctx:         ctx,
			expectedErr: fmt.Errorf("start_date must be before end_date at row: 1"),
			req: []*pb.SchoolHistory{
				{
					SchoolId:       uuid.NewString(),
					SchoolCourseId: uuid.NewString(),
					StartDate:      timestamppb.Now(),
					EndDate:        timestamppb.New(time.Now().Add(-time.Hour)),
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid school_course",
			ctx:         ctx,
			expectedErr: fmt.Errorf("school_course does not match with req.SchoolHistories"),
			req: []*pb.SchoolHistory{
				{
					SchoolId:       uuid.NewString(),
					SchoolCourseId: uuid.NewString(),
				},
			},
			setup: func(ctx context.Context) {
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{
					{
						ID: database.Text("invalid school_course school_info-id-1"),
					},
				}, nil).Once()
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
			},
		},
		{
			name:        "school_info is archived",
			ctx:         ctx,
			expectedErr: fmt.Errorf("school_info school_info-id-1 is archived"),
			req: []*pb.SchoolHistory{
				{
					SchoolId: idutil.ULIDNow(),
				},
			},
			setup: func(ctx context.Context) {
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{
					{
						ID:         database.Text("school_info-id-1"),
						IsArchived: database.Bool(true),
					},
				}, nil)
			},
		},
		{
			name:        "school_course is archived",
			ctx:         ctx,
			expectedErr: fmt.Errorf("school_course school_course-id-1 is archived"),
			req: []*pb.SchoolHistory{
				{
					SchoolId:       idutil.ULIDNow(),
					SchoolCourseId: idutil.ULIDNow(),
				},
			},
			setup: func(ctx context.Context) {
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{
					{
						ID: database.Text("school_info-id-1"),
					},
				}, nil).Once()
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{
					{
						ID:         database.Text("school_course-id-1"),
						IsArchived: database.Bool(true),
					},
				}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}
			err := s.validateSchoolHistoriesReq(testCase.ctx, testCase.req.([]*pb.SchoolHistory))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestUserModifierService_UpsertUserAccessPath(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db := new(mock_database.Ext)
	// tx := new(mock_database.Tx)
	studentParentRepo := new(mock_repositories.MockStudentParentRepo)
	userAccessPathRepo := new(mock_repositories.MockUserAccessPathRepo)

	s := &UserModifierService{
		DB:                 db,
		StudentParentRepo:  studentParentRepo,
		UserAccessPathRepo: userAccessPathRepo,
	}

	type MultiTypeReq struct {
		locations []*domain.Location
		studentID string
	}

	testCases := []TestCase{
		{
			name:        "happy case",
			ctx:         ctx,
			expectedErr: nil,
			req: MultiTypeReq{
				locations: []*domain.Location{
					{LocationID: "testing"},
				},
				studentID: "test",
			},
			setup: func(ctx context.Context) {
				userAccessPathRepo.On("Upsert", ctx, db, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "fail case: user_access_path Upsert fail",
			ctx:         ctx,
			expectedErr: errors.New("userAccessPathRepo.Upsert: fail in user access Upsert"),
			req: MultiTypeReq{
				locations: []*domain.Location{
					{LocationID: "testing"},
				},
				studentID: "test",
			},
			setup: func(ctx context.Context) {
				userAccessPathRepo.On("Upsert", ctx, db, mock.Anything).Once().Return(errors.New("fail in user access Upsert"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			request := testCase.req.(MultiTypeReq)

			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}

			err := UpsertUserAccessPath(testCase.ctx, s.UserAccessPathRepo, s.DB, request.locations, request.studentID)

			if err != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
				return
			}

			assert.Nil(t, err)

		})
	}
}

func TestUserModifierService_UpsertUserAccessPathForStudentParents(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db := new(mock_database.Ext)
	// tx := new(mock_database.Tx)
	studentParentRepo := new(mock_repositories.MockStudentParentRepo)
	userAccessPathRepo := new(mock_repositories.MockUserAccessPathRepo)

	s := &UserModifierService{
		DB:                 db,
		StudentParentRepo:  studentParentRepo,
		UserAccessPathRepo: userAccessPathRepo,
	}

	type MultiTypeReq struct {
		locations []*domain.Location
		studentID string
	}

	testCases := []TestCase{
		{
			name:        "happy case",
			ctx:         ctx,
			expectedErr: nil,
			req: MultiTypeReq{
				locations: []*domain.Location{
					{LocationID: "testing"},
				},
				studentID: "test",
			},
			setup: func(ctx context.Context) {
				studentParentRepo.On("FindParentIDsFromStudentID", ctx, db, mock.Anything).Once().Return([]string{"student_id_1"}, nil)
				userAccessPathRepo.On("Upsert", ctx, db, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "fail case: studentParentRepo.FindParentIDsFromStudentID fail",
			ctx:         ctx,
			expectedErr: errors.New("studentParent.FindParentIDsFromStudentID: fail in FindParentIDsFromStudentID"),
			req: MultiTypeReq{
				locations: []*domain.Location{
					{LocationID: "testing"},
				},
				studentID: "test",
			},
			setup: func(ctx context.Context) {
				studentParentRepo.On("FindParentIDsFromStudentID", ctx, db, mock.Anything).Once().Return([]string{"student_id_1"}, errors.New("fail in FindParentIDsFromStudentID"))
			},
		},
		{
			name:        "fail case: user_access_path Upsert fail",
			ctx:         ctx,
			expectedErr: errors.New("userAccessPathRepo.Upsert: fail in user access Upsert"),
			req: MultiTypeReq{
				locations: []*domain.Location{
					{LocationID: "testing"},
				},
				studentID: "test",
			},
			setup: func(ctx context.Context) {
				studentParentRepo.On("FindParentIDsFromStudentID", ctx, db, mock.Anything).Once().Return([]string{"student_id_1"}, nil)
				userAccessPathRepo.On("Upsert", ctx, db, mock.Anything).Once().Return(errors.New("fail in user access Upsert"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			request := testCase.req.(MultiTypeReq)

			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}

			err := UpsertUserAccessPathForStudentParents(testCase.ctx, s.UserAccessPathRepo, s.StudentParentRepo, s.DB, request.locations, request.studentID)

			if err != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
				return
			}

			assert.Nil(t, err)

		})
	}
}

func TestClassifyTaggedUserParams(t *testing.T) {
	ctx := auth.InjectFakeJwtToken(context.Background(), fmt.Sprint(constants.ManabieSchool))

	user := createMockDomainUser(idutil.ULIDNow())
	tag1 := createMockDomainTag(idutil.ULIDNow())
	tag2 := createMockDomainTag(idutil.ULIDNow())
	tag3 := createMockDomainTag(idutil.ULIDNow())
	tag4 := createMockDomainTag(idutil.ULIDNow())

	type args struct {
		ctx               context.Context
		userWithTags      map[entity.User][]entity.DomainTag
		existedTaggedUser []entity.DomainTaggedUser
	}
	tests := []struct {
		name          string
		args          args
		inspectResult func(*testing.T, []entity.DomainTaggedUser, []entity.DomainTaggedUser, error)
	}{
		{
			name: "user tags is kept, not changes",
			args: args{
				ctx: ctx,
				userWithTags: map[entity.User][]entity.DomainTag{
					user: {tag1, tag2},
				},
				existedTaggedUser: []entity.DomainTaggedUser{
					createMockDomainTaggedUser(user.UserID().String(), tag1.TagID().String()),
					createMockDomainTaggedUser(user.UserID().String(), tag2.TagID().String()),
				},
			},
			inspectResult: func(t *testing.T, createTaggedUsers []entity.DomainTaggedUser, deleteTaggedUsers []entity.DomainTaggedUser, err error) {
				assert.True(t, len(createTaggedUsers) == 0)
				assert.True(t, len(deleteTaggedUsers) == 0)
				assert.NoError(t, err)
			},
		},
		{
			name: "passed user tags will be created & not passed user tags will be deleted",
			args: args{
				ctx: ctx,
				userWithTags: map[entity.User][]entity.DomainTag{
					user: {tag1, tag2},
				},
				existedTaggedUser: []entity.DomainTaggedUser{
					createMockDomainTaggedUser(user.UserID().String(), tag3.TagID().String()),
					createMockDomainTaggedUser(user.UserID().String(), tag4.TagID().String()),
				},
			},
			inspectResult: func(t *testing.T, createTaggedUsers []entity.DomainTaggedUser, deleteTaggedUsers []entity.DomainTaggedUser, err error) {
				assert.Equal(t, len(createTaggedUsers), 2)
				assert.Equal(t, len(deleteTaggedUsers), 2)
				assert.NoError(t, err)

				//  new user tags must be existed in createTaggedUsers
				assert.Equal(t, createTaggedUsers[0].TagID().String(), tag1.TagID().String())
				assert.Equal(t, createTaggedUsers[0].TagID().String(), tag1.TagID().String())

				//  outdated user tags must be existed in deleteTaggedUsers
				assert.Equal(t, deleteTaggedUsers[0].TagID().String(), tag3.TagID().String())
				assert.Equal(t, deleteTaggedUsers[1].TagID().String(), tag4.TagID().String())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.args.ctx
			ctx = auth.InjectFakeJwtToken(ctx, fmt.Sprint(constants.ManabieSchool))
			createTaggedUsers, deletedTaggedUsers, err := classifyTaggedUserParams(ctx, tt.args.userWithTags, tt.args.existedTaggedUser)
			tt.inspectResult(t, createTaggedUsers, deletedTaggedUsers, err)
		})
	}
}

func Test_validUserTags(t *testing.T) {
	id1 := idutil.ULIDNow()
	id2 := idutil.ULIDNow()

	type args struct {
		role         string
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
			name: "happy case: valid for parent",
			args: func(t *testing.T) args {
				return args{
					role:   constant.RoleStudent,
					tagIDs: []string{id1, id2},
					existingTags: entity.DomainTags([]entity.DomainTag{
						createMockDomainTagWithType(id1, upb.UserTagType_USER_TAG_TYPE_STUDENT),
						createMockDomainTagWithType(id2, upb.UserTagType_USER_TAG_TYPE_STUDENT),
					}),
				}
			},
			wantErr: false,
		},
		{
			name: "happy case: valid for student",
			args: func(t *testing.T) args {
				return args{
					role:   constant.RoleStudent,
					tagIDs: []string{id1, id2},
					existingTags: entity.DomainTags([]entity.DomainTag{
						createMockDomainTagWithType(id1, upb.UserTagType_USER_TAG_TYPE_STUDENT),
						createMockDomainTagWithType(id2, upb.UserTagType_USER_TAG_TYPE_STUDENT),
					}),
				}
			},
			wantErr: false,
		},
		{
			name: "tag is not for student",
			args: func(t *testing.T) args {
				return args{
					role:   constant.RoleStudent,
					tagIDs: []string{id1, id2},
					existingTags: entity.DomainTags([]entity.DomainTag{
						createMockDomainTagWithType(id1, upb.UserTagType_USER_TAG_TYPE_STUDENT),
						createMockDomainTagWithType(id2, upb.UserTagType_USER_TAG_TYPE_PARENT),
					}),
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				assert.Equal(t, err, ErrTagIsNotForStudent)
			},
		},
		{
			name: "tag is not for parent",
			args: func(t *testing.T) args {
				return args{
					role:   constant.RoleParent,
					tagIDs: []string{id1, id2},
					existingTags: entity.DomainTags([]entity.DomainTag{
						createMockDomainTagWithType(id1, upb.UserTagType_USER_TAG_TYPE_STUDENT),
						createMockDomainTagWithType(id2, upb.UserTagType_USER_TAG_TYPE_PARENT),
					}),
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				assert.Equal(t, err, ErrTagIsNotForParent)
			},
		},
		{
			name: "tag is not existed",
			args: func(t *testing.T) args {
				return args{
					role:   constant.RoleParent,
					tagIDs: []string{id1, id2},
					existingTags: entity.DomainTags([]entity.DomainTag{
						createMockDomainTagWithType(id2, upb.UserTagType_USER_TAG_TYPE_PARENT),
					}),
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				assert.Equal(t, err, ErrTagIDsMustBeExisted)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			err := validUserTags(tArgs.role, tArgs.tagIDs, tArgs.existingTags)

			if (err != nil) != tt.wantErr {
				t.Fatalf("validUserTags error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}
