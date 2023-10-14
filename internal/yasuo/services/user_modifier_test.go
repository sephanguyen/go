package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"

	"cloud.google.com/go/storage"
	"github.com/manabie-com/backend/internal/bob/entities"
	golibs_auth "github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_services "github.com/manabie-com/backend/mock/fatima/services"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_firebase "github.com/manabie-com/backend/mock/golibs/firebase"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	ppb_v1 "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"firebase.google.com/go/v4/auth"
	"github.com/golang/protobuf/proto"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestIsSchoolAdmin(t *testing.T) {
	t.Parallel()

	t.Run("should is school admin", func(t *testing.T) {
		t.Parallel()
		result := IsSchoolAdmin("USER_GROUP_SCHOOL_ADMIN")
		assert.Exactly(t, true, result)
	})

	t.Run("should is not school admin", func(t *testing.T) {
		t.Parallel()
		result := IsSchoolAdmin("INVALID_USER_GROUP")
		assert.Exactly(t, false, result)
	})
	t.Run("teacher is not school admin", func(t *testing.T) {
		t.Parallel()
		result := IsSchoolAdmin("USER_GROUP_TEACHER")
		assert.Exactly(t, false, result)
	})
}

func TestUpdateUser(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userId := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userId)
	ctx = interceptors.ContextWithUserGroup(ctx, entities.UserGroupSchoolAdmin)

	useRepo := new(mock_repositories.MockUserRepo)
	schoolAdminRepo := new(mock_repositories.MockSchoolAdminRepo)
	teacherRepo := new(mock_repositories.MockTeacherRepo)
	studentRepo := new(mock_repositories.MockStudentRepo)

	db := new(mock_database.Ext)
	tx := &mock_database.Tx{}

	s := UserModifierService{
		DB:              db,
		UserRepo:        useRepo,
		SchoolAdminRepo: schoolAdminRepo,
		TeacherRepo:     teacherRepo,
		StudentRepo:     studentRepo,
	}

	testCases := []TestCase{
		{
			name: "invalid params",
			ctx:  ctx,
			req: &pb.UpdateUserProfileRequest{
				Id:    "id",
				Name:  "",
				Grade: 1,
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid params"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "user not found",
			ctx:  ctx,
			req: &pb.UpdateUserProfileRequest{
				Id:    "id",
				Name:  "name",
				Grade: 1,
			},
			expectedErr: fmt.Errorf("s.UserRepo.Get: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				useRepo.On("Get", ctx, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "update teacher profile",
			ctx:  ctx,
			req: &pb.UpdateUserProfileRequest{
				Id:    "id",
				Name:  "name",
				Grade: 1,
			},
			expectedErr: fmt.Errorf("s.UserRepo.UpdateProfile: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				useRepo.On("Get", ctx, mock.Anything, database.Text("id")).Once().Return(&entities.User{
					ID:       database.Text("id"),
					LastName: database.Text("name"),
				}, nil)
				useRepo.On("UserGroup", ctx, mock.Anything, database.Text("id")).Once().Return(entities.UserGroupTeacher, nil)

				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				useRepo.On("UpdateProfile", ctx, mock.Anything, mock.AnythingOfType("*entities.User")).Once().Return(pgx.ErrTxClosed)
			},
		},
		{
			name: "student not found",
			ctx:  ctx,
			req: &pb.UpdateUserProfileRequest{
				Id:    "id",
				Name:  "name",
				Grade: 1,
			},
			expectedErr: fmt.Errorf("s.StudentRepo.Find: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				useRepo.On("Get", ctx, mock.Anything, database.Text("id")).Once().Return(&entities.User{
					ID:       database.Text("id"),
					LastName: database.Text("name"),
				}, nil)
				useRepo.On("UserGroup", ctx, mock.Anything, database.Text("id")).Once().Return(entities.UserGroupStudent, nil)

				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				studentRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "update student profile",
			ctx:  ctx,
			req: &pb.UpdateUserProfileRequest{
				Id:    "id",
				Name:  "name",
				Grade: 1,
			},
			expectedErr: fmt.Errorf("s.StudentRepo.Update: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				useRepo.On("Get", ctx, mock.Anything, database.Text("id")).Once().Return(&entities.User{
					ID:       database.Text("id"),
					LastName: database.Text("name"),
				}, nil)
				useRepo.On("UserGroup", ctx, mock.Anything, database.Text("id")).Once().Return(entities.UserGroupStudent, nil)

				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				studentRepo.On("Find", ctx, mock.Anything, database.Text("id")).Once().Return(&entities.Student{
					ID:       database.Text("id"),
					SchoolID: database.Int4(1),
				}, nil)

				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				studentRepo.On("Update", ctx, mock.Anything, mock.AnythingOfType("*entities.Student")).Once().Return(pgx.ErrTxClosed)
			},
		},
	}

	for _, testCase := range testCases {
		t.Log("Test case: " + testCase.name)
		testCase.setup(testCase.ctx)

		_, err := s.UpdateUserProfile(testCase.ctx, testCase.req.(*pb.UpdateUserProfileRequest))

		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestCreateStudent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ctx = golibs_auth.InjectFakeJwtToken(ctx, "1")

	db := new(mock_database.Ext)
	tx := new(mock_database.Tx)
	userRepo := new(mock_repositories.MockUserRepo)
	studentRepo := new(mock_repositories.MockStudentRepo)
	parentRepo := new(mock_repositories.MockParentRepo)
	userGroupRepo := new(mock_repositories.MockUserGroupRepo)
	studentParentRepo := new(mock_repositories.MockStudentParentRepo)
	fatimaClient := new(mock_services.SubscriptionModifierServiceClient)
	jsm := new(mock_nats.JetStreamManagement)

	firebaseAuth := new(mock_firebase.AuthClient)

	s := UserModifierService{
		DB:                db,
		UserRepo:          userRepo,
		StudentRepo:       studentRepo,
		ParentRepo:        parentRepo,
		UserGroupRepo:     userGroupRepo,
		StudentParentRepo: studentParentRepo,
		FirebaseClient:    firebaseAuth,
		FatimaClient:      fatimaClient,
		JSM:               jsm,
	}

	existingParentUser := &entities.User{
		ID:          database.Text(idutil.ULIDNow()),
		Email:       database.Text("existing-email@example.com"),
		PhoneNumber: database.Text("existing-phone-number"),
	}
	existingParent := &entities.Parent{
		ID:       existingParentUser.ID,
		SchoolID: database.Int4(1),
		User:     *existingParentUser,
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
			name: "cannot create if student emails already exist in db",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					Name:             "user's name",
					Email:            "existing-email@example",
					Grade:            1,
					EnrollmentStatus: cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
				},
			},
			expectedErr: status.Error(codes.AlreadyExists, fmt.Sprintf("cannot create student with emails existing in system: %s", "existing-email@example")),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, database.TextArray([]string{"existing-email@example"})).Once().Return([]*entities.User{{}}, nil)
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
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
				},
			},
			expectedErr: status.Error(codes.AlreadyExists, fmt.Sprintf("cannot create student with phone number existing in system: %s", "existing-phone-number")),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, database.TextArray([]string{"existing-email@example"})).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, tx, database.TextArray([]string{"existing-phone-number"})).Once().Return([]*entities.User{{}}, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "cannot create if parent data to create has invalid params",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Email:             "existing-email@example",
					Password:          "user's password",
					Name:              "user's name",
					PhoneNumber:       "existing-phone-number",
					Grade:             1,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
				},
				ParentProfiles: []*pb.CreateStudentRequest_ParentProfile{
					{
						Name:         "",
						CountryCode:  cpb.Country_COUNTRY_VN,
						PhoneNumber:  "parent-number",
						Email:        "parent-email@example.com",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Password:     "parent-password",
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "parent name cannot be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "cannot create if parent data to create has email already exist in db",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Email:             "email@example.com",
					Password:          "user's password",
					Name:              "user's name",
					PhoneNumber:       "existing-phone-number",
					Grade:             1,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
				},
				ParentProfiles: []*pb.CreateStudentRequest_ParentProfile{
					{
						Name:         "parent-name",
						CountryCode:  cpb.Country_COUNTRY_VN,
						PhoneNumber:  "parent-number",
						Email:        existingParentUser.Email.String,
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Password:     "parent-password",
					},
				},
			},
			expectedErr: status.Error(codes.AlreadyExists, fmt.Sprintf("cannot create parent with emails existing in system: %s", strings.Join(entities.Users{existingParentUser}.Emails(), ", "))),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return([]*entities.User{existingParentUser}, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "cannot create if parent data to create has phone number already exist in db",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Email:             "existing-email@example",
					Password:          "user's password",
					Name:              "user's name",
					PhoneNumber:       "existing-phone-number",
					Grade:             1,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
				},
				ParentProfiles: []*pb.CreateStudentRequest_ParentProfile{
					{
						Name:         "parent-name",
						CountryCode:  cpb.Country_COUNTRY_VN,
						PhoneNumber:  existingParentUser.PhoneNumber.String,
						Email:        "parent-email@example.com",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Password:     "parent-password",
					},
				},
			},
			expectedErr: status.Error(codes.AlreadyExists, fmt.Sprintf("cannot create parent with phone number existing in system: %s", strings.Join(entities.Users{existingParentUser}.PhoneNumbers(), ", "))),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, database.TextArray([]string{"existing-email@example"})).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, tx, database.TextArray([]string{"existing-phone-number"})).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, tx, mock.Anything).Once().Return([]*entities.User{existingParentUser}, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "cannot create if parent data to assign has parent id is not exist in db",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Email:             "existing-email@example",
					Password:          "user's password",
					Name:              "user's name",
					PhoneNumber:       "existing-phone-number",
					Grade:             1,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
				},
				ParentProfiles: []*pb.CreateStudentRequest_ParentProfile{
					{
						Id: existingParentUser.ID.String,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Sprintf("cannot assign non-existing parent to student: %s", existingParentUser.ID.String)),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, database.TextArray([]string{"existing-email@example"})).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, tx, database.TextArray([]string{"existing-phone-number"})).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, tx, mock.Anything).Once().Return(nil, nil)
				parentRepo.On("GetByIds", ctx, tx, mock.Anything).Once().Return(entities.Parents{}, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "cannot create if missing student status data",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password: "user's password",
					Name:     "user's name",
					Email:    "email@example.com",
					Grade:    1,
					// Status:   cpb.StudentStatus_STUDENT_STATUS_ENROLLED,
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, ErrStudentEnrollmentStatusNotAllowedTobeNone.Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				firebaseAuth.On("ImportUsers", ctx, mock.Anything).Once().Return(&auth.UserImportResult{}, nil)
				firebaseAuth.On("GetUser", ctx, mock.Anything).Once().Return(nil, nil)
				firebaseAuth.On("UpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("TracedPublishAsync", mock.Anything, "nats.TracedPublishAsync", mock.Anything, mock.Anything).Once().Return("", nil)
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
					EnrollmentStatus: cpb.StudentEnrollmentStatus(999999), // some unknown status
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, ErrStudentEnrollmentStatusUnknown.Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				firebaseAuth.On("ImportUsers", ctx, mock.Anything).Once().Return(&auth.UserImportResult{}, nil)
				firebaseAuth.On("GetUser", ctx, mock.Anything).Once().Return(nil, nil)
				firebaseAuth.On("UpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("TracedPublishAsync", mock.Anything, "nats.TracedPublishAsync", mock.Anything, mock.Anything).Once().Return("", nil)
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
					EnrollmentStatus: cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NONE,
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, ErrStudentEnrollmentStatusNotAllowedTobeNone.Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				firebaseAuth.On("ImportUsers", ctx, mock.Anything).Once().Return(&auth.UserImportResult{}, nil)
				firebaseAuth.On("GetUser", ctx, mock.Anything).Once().Return(nil, nil)
				firebaseAuth.On("UpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("TracedPublishAsync", mock.Anything, "nats.TracedPublishAsync", mock.Anything, mock.Anything).Once().Return("", nil)
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
					EnrollmentStatus: cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				firebaseAuth.On("ImportUsers", ctx, mock.Anything).Once().Return(&auth.UserImportResult{}, nil)
				firebaseAuth.On("GetUser", ctx, mock.Anything).Once().Return(nil, nil)
				firebaseAuth.On("UpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("TracedPublishAsync", mock.Anything, "nats.TracedPublishAsync", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name: "create student with student data and parent data successfully",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:          "user's password",
					Name:              "user's name",
					Email:             "email@example.com",
					Grade:             1,
					CountryCode:       cpb.Country_COUNTRY_VN,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
				},
				ParentProfiles: []*pb.CreateStudentRequest_ParentProfile{
					{
						Name:         "parent-name",
						CountryCode:  cpb.Country_COUNTRY_VN,
						PhoneNumber:  "parent-number",
						Email:        "parent-email@example.com",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Password:     "parent-password",
					},
					{
						Id:           existingParent.ID.String,
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, tx, mock.Anything).Once().Return(nil, nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				parentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				parentRepo.On("GetByIds", ctx, tx, mock.Anything).Once().Return(entities.Parents{existingParent}, nil)
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
				firebaseAuth.On("ImportUsers", ctx, mock.Anything).Once().Return(&auth.UserImportResult{}, nil)
				firebaseAuth.On("GetUser", ctx, mock.Anything).Once().Return(nil, nil)
				firebaseAuth.On("UpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				firebaseAuth.On("GetUser", ctx, mock.Anything).Once().Return(nil, nil)
				firebaseAuth.On("UpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("TracedPublishAsync", mock.Anything, "nats.TracedPublishAsync", mock.Anything, mock.Anything).Once().Return("", nil)
				jsm.On("TracedPublishAsync", mock.Anything, "nats.TracedPublishAsync", mock.Anything, mock.Anything).Once().Return("", nil)
				jsm.On("TracedPublishAsync", mock.Anything, "nats.TracedPublishAsync", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name: "Cannot create student if course start date before end date",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:          "user's password",
					Name:              "user's name",
					Email:             "email@example.com",
					Grade:             1,
					CountryCode:       cpb.Country_COUNTRY_VN,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
				},
				ParentProfiles: []*pb.CreateStudentRequest_ParentProfile{
					{
						Name:         "parent-name",
						CountryCode:  cpb.Country_COUNTRY_VN,
						PhoneNumber:  "parent-number",
						Email:        "parent-email@example.com",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Password:     "parent-password",
					},
					{
						Id:           existingParent.ID.String,
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
					},
				},
				StudentPackageProfiles: []*pb.CreateStudentRequest_StudentPackageProfile{
					{
						CourseId: "course-1",
						Start:    timestamppb.Now(),
						End:      timestamppb.New(time.Now().Add(-5 * time.Hour)),
					},
				},
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: status.Error(codes.InvalidArgument, "UserModifier.validCreateRequest: package profile start date must before end date"),
		},
		{
			name: "create student with full data successfully",
			ctx:  ctx,
			req: &pb.CreateStudentRequest{
				StudentProfile: &pb.CreateStudentRequest_StudentProfile{
					Password:         "user's password",
					Name:             "user's name",
					Email:            "email@example.com",
					Grade:            1,
					CountryCode:      cpb.Country_COUNTRY_VN,
					EnrollmentStatus: cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
				},
				ParentProfiles: []*pb.CreateStudentRequest_ParentProfile{
					{
						Name:         "parent-name",
						CountryCode:  cpb.Country_COUNTRY_VN,
						PhoneNumber:  "parent-number",
						Email:        "parent-email@example.com",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Password:     "parent-password",
					},
					{
						Id:           existingParent.ID.String,
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
					},
				},
				StudentPackageProfiles: []*pb.CreateStudentRequest_StudentPackageProfile{
					{
						CourseId: "1",
						Start:    timestamppb.New(time.Now()),
						End:      timestamppb.New(time.Now().Add(24 * time.Hour)),
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				userRepo.On("GetByEmail", ctx, tx, database.TextArray([]string{"email@example.com"})).Once().Return(nil, nil)
				studentRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, tx, mock.Anything).Once().Return(nil, nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				parentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				parentRepo.On("GetByIds", ctx, tx, mock.Anything).Once().Return(entities.Parents{existingParent}, nil)
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
				fatimaClient.On("AddStudentPackageCourse", signCtx(ctx), mock.Anything).Once().Return(&fpb.AddStudentPackageCourseResponse{}, nil)
				firebaseAuth.On("ImportUsers", ctx, mock.Anything).Once().Return(&auth.UserImportResult{}, nil)
				firebaseAuth.On("GetUser", ctx, mock.Anything).Once().Return(nil, nil)
				firebaseAuth.On("UpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				firebaseAuth.On("GetUser", ctx, mock.Anything).Once().Return(nil, nil)
				firebaseAuth.On("UpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("TracedPublishAsync", mock.Anything, "nats.TracedPublishAsync", mock.Anything, mock.Anything).Once().Return("", nil)
				jsm.On("TracedPublishAsync", mock.Anything, "nats.TracedPublishAsync", mock.Anything, mock.Anything).Once().Return("", nil)
				jsm.On("TracedPublishAsync", mock.Anything, "nats.TracedPublishAsync", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Log("Test case: " + testCase.name)
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
	}
}

func TestCreateParents(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("successful", func(tt *testing.T) {
		userId := idutil.ULIDNow()
		ctx = interceptors.ContextWithUserID(ctx, userId)
		ctx = interceptors.ContextWithUserGroup(ctx, entities.UserGroupSchoolAdmin)

		schoolID := 10
		parentRepo := &mock_repositories.MockParentRepo{}

		db := new(mock_database.Ext)
		tx := &mock_database.Tx{}

		s := UserModifierService{
			DB:         db,
			ParentRepo: parentRepo,
		}

		user1 := &entities.User{}
		database.AllNullEntity(user1)
		multierr.Combine(
			user1.ID.Set(idutil.ULIDNow()),
		)
		user2 := &entities.User{}
		database.AllNullEntity(user2)
		multierr.Combine(
			user2.ID.Set(idutil.ULIDNow()),
		)
		var users []*entities.User
		users = append(users, user1, user2)
		parentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
		err := s.CreateParents(ctx, tx, int64(schoolID), users)
		assert.Nil(t, err)
	})
	t.Run("error create parent", func(tt *testing.T) {
		userId := idutil.ULIDNow()
		ctx = interceptors.ContextWithUserID(ctx, userId)
		ctx = interceptors.ContextWithUserGroup(ctx, entities.UserGroupSchoolAdmin)

		schoolID := 10
		parentRepo := &mock_repositories.MockParentRepo{}

		db := new(mock_database.Ext)
		tx := &mock_database.Tx{}

		s := UserModifierService{
			DB:         db,
			ParentRepo: parentRepo,
		}

		user1 := &entities.User{}
		database.AllNullEntity(user1)
		multierr.Combine(
			user1.ID.Set(idutil.ULIDNow()),
		)
		user2 := &entities.User{}
		database.AllNullEntity(user2)
		multierr.Combine(
			user2.ID.Set(idutil.ULIDNow()),
		)
		var users []*entities.User
		users = append(users, user1, user2)
		parentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
		err := s.CreateParents(ctx, tx, int64(schoolID), users)
		assert.Errorf(t, err, "s.ParentRepo.CreateMultiple: %w", pgx.ErrTxClosed)
	})
}

func TestAssignToParent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userId := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userId)
	ctx = interceptors.ContextWithUserGroup(ctx, entities.UserGroupSchoolAdmin)

	studentRepo := new(mock_repositories.MockStudentRepo)
	studentParentRepo := new(mock_repositories.MockStudentParentRepo)
	jsm := &mock_nats.JetStreamManagement{}
	db := new(mock_database.Ext)
	tx := &mock_database.Tx{}

	s := UserModifierService{
		DB:                db,
		StudentRepo:       studentRepo,
		StudentParentRepo: studentParentRepo,
		JSM:               jsm,
	}

	student1 := &entities.Student{}
	student1.ID.Set("student-1")

	student2 := &entities.Student{}
	student2.ID.Set("student-2")

	students := []*entities.Student{student1, student2}

	testCases := []TestCase{
		{
			name: "fail test",
			ctx:  ctx,
			req: &pb.AssignToParentRequest{
				AssignParents: []*pb.AssignToParentRequest_AssignParent{
					{
						StudentId: "student-1",
						ParentId:  "parent-1",
					},
					{
						StudentId: "student-1",
						ParentId:  "parent-2",
					},
				},
			},
			expectedErr: pgx.ErrInvalidLogLevel,
			setup: func(ctx context.Context) {
				studentRepo.On("FindStudentProfilesByIDs", ctx, db, mock.Anything).Once().Return(nil, pgx.ErrInvalidLogLevel)
			},
		},
		{
			name: "successful",
			ctx:  ctx,
			req: &pb.AssignToParentRequest{
				AssignParents: []*pb.AssignToParentRequest_AssignParent{
					{
						StudentId: "student-1",
						ParentId:  "parent-1",
					},
					{
						StudentId: "student-1",
						ParentId:  "parent-2",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				studentRepo.On("FindStudentProfilesByIDs", ctx, db, mock.Anything).Return(students, nil)
				studentParentRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				jsm.On("TracedPublishAsync", mock.Anything, mock.Anything, constants.SubjectUserCreated, mock.Anything).Return("", nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Log("Test case: " + testCase.name)
		testCase.setup(testCase.ctx)

		_, err := s.AssignToParent(testCase.ctx, testCase.req.(*pb.AssignToParentRequest))

		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestUpdateStudent(t *testing.T) {
	t.Parallel()

	db := new(mock_database.Ext)
	tx := new(mock_database.Tx)
	userRepo := new(mock_repositories.MockUserRepo)
	studentRepo := new(mock_repositories.MockStudentRepo)
	parentRepo := new(mock_repositories.MockParentRepo)
	userGroupRepo := new(mock_repositories.MockUserGroupRepo)
	studentParentRepo := new(mock_repositories.MockStudentParentRepo)
	fatimaClient := new(mock_services.SubscriptionModifierServiceClient)
	firebaseAuth := new(mock_firebase.AuthClient)
	jsm := new(mock_nats.JetStreamManagement)

	parent1 := &entities.Parent{
		ID:       database.Text("parent-id-1"),
		SchoolID: database.Int4(1),
	}
	parent2 := &entities.Parent{
		ID:       database.Text("parent-id-2"),
		SchoolID: database.Int4(1),
	}
	student_parent := &entities.StudentParent{
		StudentID: database.Text("student-id-1"),
		ParentID:  database.Text("existed-parent-1"),
	}

	parent1.Email.Set("parent-id-1-email@example.com")
	parent2.Email.Set("parent-id-2-email@example.com")

	tcs := []struct {
		name     string
		req      *pb.UpdateStudentRequest
		setup    func(context.Context)
		hasError bool
	}{
		// happy case
		{
			name: "update student with some unchanged, new, edited (email), and delete parents",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Albert Einstein",
					Grade:             2,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1-edited@example.com",
				},
				ParentProfiles: []*pb.UpdateStudentRequest_ParentProfile{
					// current parents of this student
					{
						Id:           "parent-id-1",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Email:        "parent-id-1-email@example.com",
					},
					// parent with edited email
					{
						Id:           "parent-id-2",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Email:        "parent-id-2-email-edited@example.com",
					},
					// new parents of this student
					{
						Name:         "Pauline Einstein",
						Email:        "pauline@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "313-876-3458",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     "password-example-1",
					},
					{
						Name:         "Abraham Einstein",
						Email:        "abraham@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "870-847-9833",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER,
						Password:     "password-example-1",
					},
				},
				StudentPackageProfiles: []*pb.UpdateStudentRequest_StudentPackageProfile{
					// update current StudentPackageProfiles
					{
						Id: &pb.UpdateStudentRequest_StudentPackageProfile_StudentPackageId{
							StudentPackageId: "studentPackage-id-1",
						},
						StartTime: timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
						EndTime:   timestamppb.New(time.Date(2021, 3, 3, 4, 5, 6, 0, time.UTC)),
					},
					{
						Id: &pb.UpdateStudentRequest_StudentPackageProfile_StudentPackageId{
							StudentPackageId: "studentPackage-id-2",
						},
						StartTime: timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
						EndTime:   timestamppb.New(time.Date(2021, 2, 4, 4, 5, 6, 0, time.UTC)),
					},
					// add new StudentPackageProfiles
					{
						Id: &pb.UpdateStudentRequest_StudentPackageProfile_CourseId{
							CourseId: "course-id-1",
						},
						StartTime: timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
						EndTime:   timestamppb.New(time.Date(2021, 3, 3, 4, 5, 6, 0, time.UTC)),
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				studentRepo.
					On("Find", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.Student{
						ID:           database.Text("student-id-1"),
						CurrentGrade: database.Int2(int16(3)),
						SchoolID:     database.Int4(1),
					}, nil)
				studentRepo.
					On("UpdateV2", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						student := args.Get(2).(*entities.Student)
						assert.Equal(t, "student-id-1", student.ID.String)
						assert.Equal(t, "Albert Einstein", student.LastName.String)
						assert.Equal(t, int16(2), student.CurrentGrade.Int)
					}).
					Once().
					Return(nil)
				studentParentRepo.
					On("Upsert", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						studentParents := args.Get(2).([]*entities.StudentParent)
						assert.Len(t, studentParents, 4)
						expectedUpdate := map[string]pb.FamilyRelationship{
							"parent-id-1": pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
							"parent-id-2": pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						}

						numberUpdate := 0
						for _, s := range studentParents {
							assert.Equal(t, "student-id-1", s.StudentID.String)
							if v, ok := expectedUpdate[s.ParentID.String]; ok {
								assert.Equal(t, v.String(), s.Relationship.String)
								numberUpdate++
							} else {
								// assign new parents
								assert.Contains(t,
									[]string{
										pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER.String(),
										pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER.String(),
									},
									s.Relationship.String,
								)
							}
						}
						assert.Equal(t, len(expectedUpdate), numberUpdate)
					}).
					Once().
					Return(nil)
				studentParentRepo.On("GetStudentParents", ctx, tx, mock.Anything).Once().Return([]*entities.StudentParent{student_parent}, nil)
				parentRepo.
					On("GetByIds", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						parentIDs := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"parent-id-1", "parent-id-2"}, database.FromTextArray(parentIDs))
					}).
					Once().
					Return(entities.Parents{
						parent1,
						&entities.Parent{
							ID:       database.Text("parent-id-2"),
							SchoolID: database.Int4(1),
							User: entities.User{
								Email: pgtype.Text{
									String: "parent-id-2-email@example.com",
									Status: pgtype.Present,
								},
							},
						},
					}, nil)
				parentRepo.
					On("CreateMultiple", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						parents := args.Get(2).([]*entities.Parent)
						assert.Len(t, parents, 2)
						for _, p := range parents {
							assert.Equal(t, int32(1), p.SchoolID.Int)
							assert.NotEmpty(t, p.ID.String)
						}
					}).
					Once().
					Return(nil)
				userGroupRepo.
					On("CreateMultiple", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						userGr := args.Get(2).([]*entities.UserGroup)
						assert.Len(t, userGr, 2)
						for _, p := range userGr {
							assert.NotEmpty(t, p.UserID.String)
							assert.Equal(t, entities.UserGroupParent, p.GroupID.String)
							assert.True(t, p.IsOrigin.Bool)
							assert.Equal(t, entities.UserGroupStatusActive, p.Status.String)
						}
					}).
					Once().
					Return(nil)
				userRepo.
					On("Get", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.User{
						LastName: database.Text("Albert Einstein JR"),
						Email:    database.Text("student-id-1@example.com"),
					}, nil)
				userRepo.
					On("GetByEmail", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						emails := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t,
							[]string{"student-id-1-edited@example.com"},
							database.FromTextArray(emails))
					}).
					Once().
					Return([]*entities.User{}, nil)
				userRepo.
					On("GetByEmail", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						emails := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t,
							[]string{"pauline@example.com", "abraham@example.com"},
							database.FromTextArray(emails))
					}).
					Once().
					Return(nil, nil)
				userRepo.
					On("GetByEmail", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						emails := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t,
							[]string{"parent-id-2-email-edited@example.com"},
							database.FromTextArray(emails))
					}).
					Once().
					Return([]*entities.User{}, nil)
				userRepo.
					On("UpdateEmail", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						editedParent := args.Get(2).(*entities.User)
						assert.Equal(t, "parent-id-2", editedParent.ID.String)
						assert.Equal(t, "parent-id-2-email-edited@example.com", editedParent.Email.String)
					}).
					Once().
					Return(nil)
				userRepo.
					On("GetByPhone", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						phoneNumbers := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"313-876-3458", "870-847-9833"}, database.FromTextArray(phoneNumbers))
					}).
					Once().
					Return(nil, nil)
				userRepo.
					On("CreateMultiple", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						parents := args.Get(2).([]*entities.User)
						assert.Len(t, parents, 2)
						expectedByName := map[string]*pb.UpdateStudentRequest_ParentProfile{
							"Pauline Einstein": {
								Name:        "Pauline Einstein",
								Email:       "pauline@example.com",
								CountryCode: cpb.Country_COUNTRY_JP,
								PhoneNumber: "313-876-3458",
							},
							"Abraham Einstein": {
								Name:        "Abraham Einstein",
								Email:       "abraham@example.com",
								CountryCode: cpb.Country_COUNTRY_JP,
								PhoneNumber: "870-847-9833",
							},
						}
						for _, p := range parents {
							name := p.LastName.String
							assert.NotNil(t, expectedByName[name])
							assert.Equal(t, expectedByName[name].Email, p.Email.String)
							assert.Equal(t, expectedByName[name].CountryCode.String(), p.Country.String)
							assert.Equal(t, expectedByName[name].PhoneNumber, p.PhoneNumber.String)
						}
					}).
					Once().
					Return(nil)
				firebaseAuth.
					On("ImportUsers", ctx, mock.Anything).
					Run(func(args mock.Arguments) {
						users := args.Get(1).([]*auth.UserToImport)
						assert.Len(t, users, 2)
						for _, p := range users {
							// cheat to access private field params of auth.UserToImport
							rs := reflect.ValueOf(p).Elem()
							rf := rs.Field(0)
							rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
							params := rf.Interface().(map[string]interface{})
							assert.NotEmpty(t, params["localId"].(string))
							assert.NotNil(t, params["customClaims"].(map[string]interface{}))
							assert.Contains(t, []string{"pauline@example.com", "abraham@example.com"}, params["email"].(string))
						}
					}).
					Once().
					Return(&auth.UserImportResult{}, nil)
				firebaseAuth.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				firebaseAuth.
					On("UpdateUser", ctx, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						userId := args.Get(1).(string)
						assert.NotEmpty(t, userId)
						assert.Equal(t, "student-id-1", userId)
						user := args.Get(2).(*auth.UserToUpdate)

						// cheat to access private field params of auth.UserToUpdate
						rs := reflect.ValueOf(user).Elem()
						rf := rs.Field(0)
						rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
						params := rf.Interface().(map[string]interface{})
						assert.Equal(t, "student-id-1-edited@example.com", params["email"].(string))
					}).
					Once().
					Return(nil, nil)
				firebaseAuth.
					On("UpdateUser", ctx, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						userId := args.Get(1).(string)
						assert.NotEmpty(t, userId)
						assert.NotContains(t, []string{"student-id-1", "parent-id-1", "parent-id-2"}, userId)
						user := args.Get(2).(*auth.UserToUpdate)

						// cheat to access private field params of auth.UserToUpdate
						rs := reflect.ValueOf(user).Elem()
						rf := rs.Field(0)
						rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
						params := rf.Interface().(map[string]interface{})
						assert.Contains(t, []string{"password-example-1", "password-example-2"}, params["password"].(string))
					}).
					Twice().
					Return(nil, nil)
				firebaseAuth.
					On("UpdateUser", ctx, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						userId := args.Get(1).(string)
						assert.NotEmpty(t, userId)
						assert.Equal(t, "parent-id-2", userId)
						user := args.Get(2).(*auth.UserToUpdate)

						// cheat to access private field params of auth.UserToUpdate
						rs := reflect.ValueOf(user).Elem()
						rf := rs.Field(0)
						rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
						params := rf.Interface().(map[string]interface{})
						assert.Equal(t, "parent-id-2-email-edited@example.com", params["email"].(string))
					}).
					Once().
					Return(nil, nil)
				jsm.
					On("TracedPublishAsync", mock.Anything, "nats.TracedPublishAsync", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						topic := args[2].(string)
						assert.Equal(t, constants.SubjectUserCreated, topic)

						data := args[3].([]byte)
						msg := &ppb_v1.EvtUser{}
						err := proto.Unmarshal(data, msg)
						require.NoError(t, err)
						switch {
						case msg.GetCreateParent() != nil:
							assert.Equal(t, "student-id-1", msg.GetCreateParent().StudentId)
							assert.NotEmpty(t, msg.GetCreateParent().ParentId)
							assert.NotContains(t, []string{"student-id-1"}, msg.GetCreateParent().ParentId)
							assert.Equal(t, "Albert Einstein", msg.GetCreateParent().StudentName)
							assert.Equal(t, "1", msg.GetCreateParent().SchoolId)
						case msg.GetParentRemovedFromStudent() != nil:
							assert.Equal(t, student_parent.ParentID.String, msg.GetParentRemovedFromStudent().ParentId)
							assert.Equal(t, student_parent.StudentID.String, msg.GetParentRemovedFromStudent().StudentId)
						}
					}).
					Times(5).
					Return("", nil)
				jsm.
					On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						topic := args[1].(string)
						assert.Equal(t, constants.SubjectUserDeviceTokenUpdated, topic)

						data := args[2].([]byte)
						msg := &ppb_v1.EvtUserInfo{}
						err := proto.Unmarshal(data, msg)
						require.NoError(t, err)
						assert.Equal(t, "student-id-1", msg.GetUserId())
						assert.Equal(t, "Albert Einstein", msg.GetName())
					}).
					Once().
					Return("", nil)
				fatimaClient.
					On("EditTimeStudentPackage", signCtx(ctx), mock.Anything).
					Run(func(args mock.Arguments) {
						editReq := args[1].(*fpb.EditTimeStudentPackageRequest)
						expected := map[string]*fpb.EditTimeStudentPackageRequest{
							"studentPackage-id-1": {
								StudentPackageId: "studentPackage-id-1",
								StartAt:          timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
								EndAt:            timestamppb.New(time.Date(2021, 3, 3, 4, 5, 6, 0, time.UTC)),
							},
							"studentPackage-id-2": {
								StudentPackageId: "studentPackage-id-2",
								StartAt:          timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
								EndAt:            timestamppb.New(time.Date(2021, 2, 4, 4, 5, 6, 0, time.UTC)),
							},
						}
						assert.Equal(t, expected[editReq.StudentPackageId].StudentPackageId, editReq.StudentPackageId)
						assert.Equal(t, expected[editReq.StudentPackageId].StartAt.AsTime(), editReq.StartAt.AsTime())
						assert.Equal(t, expected[editReq.StudentPackageId].EndAt.AsTime(), editReq.EndAt.AsTime())
					}).
					Twice().
					Return(&fpb.EditTimeStudentPackageResponse{}, nil)
				fatimaClient.
					On("AddStudentPackageCourse", signCtx(ctx), mock.Anything).
					Run(func(args mock.Arguments) {
						editReq := args[1].(*fpb.AddStudentPackageCourseRequest)
						expected := fpb.AddStudentPackageCourseRequest{
							CourseIds: []string{"course-id-1"},
							StartAt:   timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
							EndAt:     timestamppb.New(time.Date(2021, 3, 3, 4, 5, 6, 0, time.UTC)),
						}
						assert.Equal(t, expected.CourseIds, editReq.CourseIds)
						assert.Equal(t, expected.StartAt.AsTime(), editReq.StartAt.AsTime())
						assert.Equal(t, expected.EndAt.AsTime(), editReq.EndAt.AsTime())
					}).
					Once().
					Return(&fpb.AddStudentPackageCourseResponse{}, nil)
			},
		},
		// specific cases
		{
			name: "update student which has no any parent",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Albert Einstein",
					Grade:             2,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1@example.com",
				},
				StudentPackageProfiles: []*pb.UpdateStudentRequest_StudentPackageProfile{
					{
						Id: &pb.UpdateStudentRequest_StudentPackageProfile_StudentPackageId{
							StudentPackageId: "studentPackage-id-1",
						},
						StartTime: timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
						EndTime:   timestamppb.New(time.Date(2021, 3, 3, 4, 5, 6, 0, time.UTC)),
					},
					{
						Id: &pb.UpdateStudentRequest_StudentPackageProfile_StudentPackageId{
							StudentPackageId: "studentPackage-id-2",
						},
						StartTime: timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
						EndTime:   timestamppb.New(time.Date(2021, 2, 4, 4, 5, 6, 0, time.UTC)),
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				studentRepo.
					On("Find", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.Student{
						ID:           database.Text("student-id-1"),
						CurrentGrade: database.Int2(int16(3)),
						SchoolID:     database.Int4(1),
					}, nil)
				studentRepo.
					On("UpdateV2", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						student := args.Get(2).(*entities.Student)
						assert.Equal(t, "student-id-1", student.ID.String)
						assert.Equal(t, "Albert Einstein", student.LastName.String)
						assert.Equal(t, int16(2), student.CurrentGrade.Int)
					}).
					Once().
					Return(nil)
				userRepo.
					On("Get", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.User{
						LastName: database.Text("Albert Einstein JR"),
						Email:    database.Text("student-id-1@example.com"),
					}, nil)
				studentParentRepo.
					On("Upsert", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						studentParents := args.Get(2).([]*entities.StudentParent)
						assert.Len(t, studentParents, 1)
						assert.Equal(t, "student-id-1", studentParents[0].StudentID.String)
						assert.NotEqual(t, pgtype.Present, studentParents[0].ParentID.Status)
					}).
					Once().
					Return(nil)
				studentParentRepo.On("GetStudentParents", ctx, tx, mock.Anything).Once().Return([]*entities.StudentParent{}, nil)
				fatimaClient.
					On("EditTimeStudentPackage", signCtx(ctx), mock.Anything).
					Run(func(args mock.Arguments) {
						editReq := args[1].(*fpb.EditTimeStudentPackageRequest)
						expected := map[string]*fpb.EditTimeStudentPackageRequest{
							"studentPackage-id-1": {
								StudentPackageId: "studentPackage-id-1",
								StartAt:          timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
								EndAt:            timestamppb.New(time.Date(2021, 3, 3, 4, 5, 6, 0, time.UTC)),
							},
							"studentPackage-id-2": {
								StudentPackageId: "studentPackage-id-2",
								StartAt:          timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
								EndAt:            timestamppb.New(time.Date(2021, 2, 4, 4, 5, 6, 0, time.UTC)),
							},
						}
						assert.Equal(t, expected[editReq.StudentPackageId].StudentPackageId, editReq.StudentPackageId)
						assert.Equal(t, expected[editReq.StudentPackageId].StartAt.AsTime(), editReq.StartAt.AsTime())
						assert.Equal(t, expected[editReq.StudentPackageId].EndAt.AsTime(), editReq.EndAt.AsTime())
					}).
					Twice().
					Return(&fpb.EditTimeStudentPackageResponse{}, nil)
				jsm.
					On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						topic := args[1].(string)
						assert.Equal(t, constants.SubjectUserDeviceTokenUpdated, topic)

						data := args[2].([]byte)
						msg := &ppb_v1.EvtUserInfo{}
						err := proto.Unmarshal(data, msg)
						require.NoError(t, err)
						assert.Equal(t, "student-id-1", msg.GetUserId())
						assert.Equal(t, "Albert Einstein", msg.GetName())
					}).
					Once().
					Return("", nil)
			},
		},
		{
			name: "update student without any current parents",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:               "student-id-1",
					Name:             "Albert Einstein",
					Grade:            2,
					EnrollmentStatus: cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Email:            "student-id-1@example.com",
				},
				StudentPackageProfiles: []*pb.UpdateStudentRequest_StudentPackageProfile{
					{
						Id: &pb.UpdateStudentRequest_StudentPackageProfile_StudentPackageId{
							StudentPackageId: "studentPackage-id-1",
						},
						StartTime: timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
						EndTime:   timestamppb.New(time.Date(2021, 3, 3, 4, 5, 6, 0, time.UTC)),
					},
					{
						Id: &pb.UpdateStudentRequest_StudentPackageProfile_StudentPackageId{
							StudentPackageId: "studentPackage-id-2",
						},
						StartTime: timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
						EndTime:   timestamppb.New(time.Date(2021, 2, 4, 4, 5, 6, 0, time.UTC)),
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				studentRepo.
					On("Find", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.Student{
						ID:           database.Text("student-id-1"),
						CurrentGrade: database.Int2(int16(3)),
						SchoolID:     database.Int4(1),
					}, nil)
				studentRepo.
					On("UpdateV2", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						student := args.Get(2).(*entities.Student)
						assert.Equal(t, "student-id-1", student.ID.String)
						assert.Equal(t, "Albert Einstein", student.LastName.String)
						assert.Equal(t, int16(2), student.CurrentGrade.Int)
					}).
					Once().
					Return(nil)
				studentParentRepo.On("GetStudentParents", ctx, tx, mock.Anything).Once().Return([]*entities.StudentParent{}, nil)
				userRepo.
					On("Get", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.User{
						LastName: database.Text("Albert Einstein JR"),
						Email:    database.Text("student-id-1@example.com"),
					}, nil)
				studentParentRepo.
					On("Upsert", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						studentParents := args.Get(2).([]*entities.StudentParent)
						assert.Len(t, studentParents, 1)
						assert.Equal(t, "student-id-1", studentParents[0].StudentID.String)
						assert.NotEqual(t, pgtype.Present, studentParents[0].ParentID.Status)
					}).
					Once().
					Return(nil)
				fatimaClient.
					On("EditTimeStudentPackage", signCtx(ctx), mock.Anything).
					Run(func(args mock.Arguments) {
						editReq := args[1].(*fpb.EditTimeStudentPackageRequest)
						expected := map[string]*fpb.EditTimeStudentPackageRequest{
							"studentPackage-id-1": {
								StudentPackageId: "studentPackage-id-1",
								StartAt:          timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
								EndAt:            timestamppb.New(time.Date(2021, 3, 3, 4, 5, 6, 0, time.UTC)),
							},
							"studentPackage-id-2": {
								StudentPackageId: "studentPackage-id-2",
								StartAt:          timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
								EndAt:            timestamppb.New(time.Date(2021, 2, 4, 4, 5, 6, 0, time.UTC)),
							},
						}
						assert.Equal(t, expected[editReq.StudentPackageId].StudentPackageId, editReq.StudentPackageId)
						assert.Equal(t, expected[editReq.StudentPackageId].StartAt.AsTime(), editReq.StartAt.AsTime())
						assert.Equal(t, expected[editReq.StudentPackageId].EndAt.AsTime(), editReq.EndAt.AsTime())
					}).
					Twice().
					Return(&fpb.EditTimeStudentPackageResponse{}, nil)
				jsm.
					On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						topic := args[1].(string)
						assert.Equal(t, constants.SubjectUserDeviceTokenUpdated, topic)

						data := args[2].([]byte)
						msg := &ppb_v1.EvtUserInfo{}
						err := proto.Unmarshal(data, msg)
						require.NoError(t, err)
						assert.Equal(t, "student-id-1", msg.GetUserId())
						assert.Equal(t, "Albert Einstein", msg.GetName())
					}).
					Once().
					Return("", nil)
			},
		},
		{
			name: "update student with new parents and remove all current parents",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Albert Einstein",
					Grade:             2,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1@example.com",
				},
				ParentProfiles: []*pb.UpdateStudentRequest_ParentProfile{
					// new parents of this student
					{
						Name:         "Pauline Einstein",
						Email:        "pauline@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "313-876-3458",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     "password-example-1",
					},
					{
						Name:         "Abraham Einstein",
						Email:        "abraham@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "870-847-9833",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER,
						Password:     "password-example-2",
					},
				},
				StudentPackageProfiles: []*pb.UpdateStudentRequest_StudentPackageProfile{
					{
						Id: &pb.UpdateStudentRequest_StudentPackageProfile_StudentPackageId{
							StudentPackageId: "studentPackage-id-1",
						},
						StartTime: timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
						EndTime:   timestamppb.New(time.Date(2021, 3, 3, 4, 5, 6, 0, time.UTC)),
					},
					{
						Id: &pb.UpdateStudentRequest_StudentPackageProfile_StudentPackageId{
							StudentPackageId: "studentPackage-id-2",
						},
						StartTime: timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
						EndTime:   timestamppb.New(time.Date(2021, 2, 4, 4, 5, 6, 0, time.UTC)),
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				studentRepo.
					On("Find", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.Student{
						ID:           database.Text("student-id-1"),
						CurrentGrade: database.Int2(int16(3)),
						SchoolID:     database.Int4(1),
					}, nil)
				studentRepo.
					On("UpdateV2", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						student := args.Get(2).(*entities.Student)
						assert.Equal(t, "student-id-1", student.ID.String)
						assert.Equal(t, "Albert Einstein", student.LastName.String)
						assert.Equal(t, int16(2), student.CurrentGrade.Int)
					}).
					Once().
					Return(nil)
				studentParentRepo.
					On("Upsert", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						studentParents := args.Get(2).([]*entities.StudentParent)
						assert.Len(t, studentParents, 2)

						for _, s := range studentParents {
							assert.Equal(t, "student-id-1", s.StudentID.String)
							// assign new parents
							assert.Contains(t,
								[]string{
									pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER.String(),
									pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER.String(),
								},
								s.Relationship.String,
							)
						}
					}).
					Once().
					Return(nil)
				studentParentRepo.On("GetStudentParents", ctx, tx, mock.Anything).Once().Return([]*entities.StudentParent{student_parent}, nil)
				parentRepo.
					On("CreateMultiple", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						parents := args.Get(2).([]*entities.Parent)
						assert.Len(t, parents, 2)
						for _, p := range parents {
							assert.Equal(t, int32(1), p.SchoolID.Int)
							assert.NotEmpty(t, p.ID.String)
						}
					}).
					Once().
					Return(nil)
				userGroupRepo.
					On("CreateMultiple", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						userGr := args.Get(2).([]*entities.UserGroup)
						assert.Len(t, userGr, 2)
						for _, p := range userGr {
							assert.NotEmpty(t, p.UserID.String)
							assert.Equal(t, entities.UserGroupParent, p.GroupID.String)
							assert.True(t, p.IsOrigin.Bool)
							assert.Equal(t, entities.UserGroupStatusActive, p.Status.String)
						}
					}).
					Once().
					Return(nil)
				userRepo.
					On("Get", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.User{
						LastName: database.Text("Albert Einstein JR"),
						Email:    database.Text("student-id-1@example.com"),
					}, nil)
				userRepo.
					On("GetByEmail", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						emails := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"pauline@example.com", "abraham@example.com"}, database.FromTextArray(emails))
					}).
					Once().
					Return(nil, nil)
				userRepo.
					On("GetByPhone", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						phoneNumbers := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"313-876-3458", "870-847-9833"}, database.FromTextArray(phoneNumbers))
					}).
					Once().
					Return(nil, nil)
				userRepo.
					On("CreateMultiple", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						parents := args.Get(2).([]*entities.User)
						assert.Len(t, parents, 2)
						expectedByName := map[string]*pb.UpdateStudentRequest_ParentProfile{
							"Pauline Einstein": {
								Name:        "Pauline Einstein",
								Email:       "pauline@example.com",
								CountryCode: cpb.Country_COUNTRY_JP,
								PhoneNumber: "313-876-3458",
							},
							"Abraham Einstein": {
								Name:        "Abraham Einstein",
								Email:       "abraham@example.com",
								CountryCode: cpb.Country_COUNTRY_JP,
								PhoneNumber: "870-847-9833",
							},
						}
						for _, p := range parents {
							name := p.LastName.String
							assert.NotNil(t, expectedByName[name])
							assert.Equal(t, expectedByName[name].Email, p.Email.String)
							assert.Equal(t, expectedByName[name].CountryCode.String(), p.Country.String)
							assert.Equal(t, expectedByName[name].PhoneNumber, p.PhoneNumber.String)
						}
					}).
					Once().
					Return(nil)
				firebaseAuth.
					On("ImportUsers", ctx, mock.Anything).
					Run(func(args mock.Arguments) {
						users := args.Get(1).([]*auth.UserToImport)
						assert.Len(t, users, 2)
						for _, p := range users {
							// cheat to access private field params of auth.UserToImport
							rs := reflect.ValueOf(p).Elem()
							rf := rs.Field(0)
							rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
							params := rf.Interface().(map[string]interface{})
							assert.NotEmpty(t, params["localId"].(string))
							assert.NotNil(t, params["customClaims"].(map[string]interface{}))
							assert.Contains(t, []string{"pauline@example.com", "abraham@example.com"}, params["email"].(string))
						}
					}).
					Once().
					Return(&auth.UserImportResult{}, nil)
				firebaseAuth.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				firebaseAuth.
					On("UpdateUser", ctx, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						userId := args.Get(1).(string)
						assert.NotEmpty(t, userId)
						assert.NotContains(t, []string{"student-id-1", "parent-id-1", "parent-id-2"}, userId)
						user := args.Get(2).(*auth.UserToUpdate)

						// cheat to access private field params of auth.UserToUpdate
						rs := reflect.ValueOf(user).Elem()
						rf := rs.Field(0)
						rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
						params := rf.Interface().(map[string]interface{})
						assert.Contains(t, []string{"password-example-1", "password-example-2"}, params["password"].(string))
					}).
					Twice().
					Return(nil, nil)
				jsm.
					On("TracedPublishAsync", mock.Anything, "nats.TracedPublishAsync", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						topic := args[2].(string)
						assert.Equal(t, constants.SubjectUserCreated, topic)

						data := args[3].([]byte)
						msg := &ppb_v1.EvtUser{}
						err := proto.Unmarshal(data, msg)
						require.NoError(t, err)
						switch {
						case msg.GetCreateParent() != nil:
							assert.Equal(t, "student-id-1", msg.GetCreateParent().StudentId)
							assert.NotEmpty(t, msg.GetCreateParent().ParentId)
							assert.NotContains(t, []string{"student-id-1", "parent-id-1", "parent-id-2"}, msg.GetCreateParent().ParentId)
							assert.Equal(t, "Albert Einstein", msg.GetCreateParent().StudentName)
							assert.Equal(t, "1", msg.GetCreateParent().SchoolId)
						case msg.GetParentRemovedFromStudent() != nil:
							assert.Equal(t, student_parent.ParentID.String, msg.GetParentRemovedFromStudent().ParentId)
							assert.Equal(t, student_parent.StudentID.String, msg.GetParentRemovedFromStudent().StudentId)
						}
					}).
					Times(3).
					Return("", nil)
				jsm.
					On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						topic := args[1].(string)
						assert.Equal(t, constants.SubjectUserDeviceTokenUpdated, topic)

						data := args[2].([]byte)
						msg := &ppb_v1.EvtUserInfo{}
						err := proto.Unmarshal(data, msg)
						require.NoError(t, err)
						assert.Equal(t, "student-id-1", msg.GetUserId())
						assert.Equal(t, "Albert Einstein", msg.GetName())
					}).
					Once().
					Return("", nil)
				fatimaClient.
					On("EditTimeStudentPackage", signCtx(ctx), mock.Anything).
					Run(func(args mock.Arguments) {
						editReq := args[1].(*fpb.EditTimeStudentPackageRequest)
						expected := map[string]*fpb.EditTimeStudentPackageRequest{
							"studentPackage-id-1": {
								StudentPackageId: "studentPackage-id-1",
								StartAt:          timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
								EndAt:            timestamppb.New(time.Date(2021, 3, 3, 4, 5, 6, 0, time.UTC)),
							},
							"studentPackage-id-2": {
								StudentPackageId: "studentPackage-id-2",
								StartAt:          timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
								EndAt:            timestamppb.New(time.Date(2021, 2, 4, 4, 5, 6, 0, time.UTC)),
							},
						}
						assert.Equal(t, expected[editReq.StudentPackageId].StudentPackageId, editReq.StudentPackageId)
						assert.Equal(t, expected[editReq.StudentPackageId].StartAt.AsTime(), editReq.StartAt.AsTime())
						assert.Equal(t, expected[editReq.StudentPackageId].EndAt.AsTime(), editReq.EndAt.AsTime())
					}).
					Twice().
					Return(&fpb.EditTimeStudentPackageResponse{}, nil)
			},
		},
		{
			name: "update student with has no any student package profiles",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:               "student-id-1",
					Name:             "Albert Einstein",
					Grade:            2,
					EnrollmentStatus: cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Email:            "student-id-1@example.com",
				},
				ParentProfiles: []*pb.UpdateStudentRequest_ParentProfile{
					// current parents of this student
					{
						Id:           "parent-id-1",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Email:        "parent-id-1-email@example.com",
					},
					{
						Id:           "parent-id-2",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Email:        "parent-id-2-email@example.com",
					},
					// new parents of this student
					{
						Name:         "Pauline Einstein",
						Email:        "pauline@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "313-876-3458",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     "password-example-1",
					},
					{
						Name:         "Abraham Einstein",
						Email:        "abraham@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "870-847-9833",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER,
						Password:     "password-example-1",
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				studentRepo.
					On("Find", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.Student{
						ID:           database.Text("student-id-1"),
						CurrentGrade: database.Int2(int16(3)),
						SchoolID:     database.Int4(1),
					}, nil)
				studentRepo.
					On("UpdateV2", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						student := args.Get(2).(*entities.Student)
						assert.Equal(t, "student-id-1", student.ID.String)
						assert.Equal(t, "Albert Einstein", student.LastName.String)
						assert.Equal(t, int16(2), student.CurrentGrade.Int)
					}).
					Once().
					Return(nil)
				studentParentRepo.
					On("Upsert", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						studentParents := args.Get(2).([]*entities.StudentParent)
						assert.Len(t, studentParents, 4)
						expectedUpdate := map[string]pb.FamilyRelationship{
							"parent-id-1": pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
							"parent-id-2": pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						}

						numberUpdate := 0
						for _, s := range studentParents {
							assert.Equal(t, "student-id-1", s.StudentID.String)
							if v, ok := expectedUpdate[s.ParentID.String]; ok {
								assert.Equal(t, v.String(), s.Relationship.String)
								numberUpdate++
							} else {
								// assign new parents
								assert.Contains(t,
									[]string{
										pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER.String(),
										pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER.String(),
									},
									s.Relationship.String,
								)
							}
						}
						assert.Equal(t, len(expectedUpdate), numberUpdate)
					}).
					Once().
					Return(nil)
				studentParentRepo.On("GetStudentParents", ctx, tx, mock.Anything).Once().Return([]*entities.StudentParent{}, nil)
				parentRepo.
					On("GetByIds", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						parentIDs := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"parent-id-1", "parent-id-2"}, database.FromTextArray(parentIDs))
					}).
					Once().
					Return(entities.Parents{
						parent1,
						parent2,
					}, nil)
				parentRepo.
					On("CreateMultiple", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						parents := args.Get(2).([]*entities.Parent)
						assert.Len(t, parents, 2)
						for _, p := range parents {
							assert.Equal(t, int32(1), p.SchoolID.Int)
							assert.NotEmpty(t, p.ID.String)
						}
					}).
					Once().
					Return(nil)
				userGroupRepo.
					On("CreateMultiple", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						userGr := args.Get(2).([]*entities.UserGroup)
						assert.Len(t, userGr, 2)
						for _, p := range userGr {
							assert.NotEmpty(t, p.UserID.String)
							assert.Equal(t, entities.UserGroupParent, p.GroupID.String)
							assert.True(t, p.IsOrigin.Bool)
							assert.Equal(t, entities.UserGroupStatusActive, p.Status.String)
						}
					}).
					Once().
					Return(nil)
				userRepo.
					On("Get", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.User{
						LastName: database.Text("Albert Einstein JR"),
						Email:    database.Text("student-id-1@example.com"),
					}, nil)
				userRepo.
					On("GetByEmail", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						emails := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"pauline@example.com", "abraham@example.com"}, database.FromTextArray(emails))
					}).
					Once().
					Return(nil, nil)
				userRepo.
					On("GetByPhone", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						phoneNumbers := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"313-876-3458", "870-847-9833"}, database.FromTextArray(phoneNumbers))
					}).
					Once().
					Return(nil, nil)
				userRepo.
					On("CreateMultiple", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						parents := args.Get(2).([]*entities.User)
						assert.Len(t, parents, 2)
						expectedByName := map[string]*pb.UpdateStudentRequest_ParentProfile{
							"Pauline Einstein": {
								Name:        "Pauline Einstein",
								Email:       "pauline@example.com",
								CountryCode: cpb.Country_COUNTRY_JP,
								PhoneNumber: "313-876-3458",
							},
							"Abraham Einstein": {
								Name:        "Abraham Einstein",
								Email:       "abraham@example.com",
								CountryCode: cpb.Country_COUNTRY_JP,
								PhoneNumber: "870-847-9833",
							},
						}
						for _, p := range parents {
							name := p.LastName.String
							assert.NotNil(t, expectedByName[name])
							assert.Equal(t, expectedByName[name].Email, p.Email.String)
							assert.Equal(t, expectedByName[name].CountryCode.String(), p.Country.String)
							assert.Equal(t, expectedByName[name].PhoneNumber, p.PhoneNumber.String)
						}
					}).
					Once().
					Return(nil)
				firebaseAuth.
					On("ImportUsers", ctx, mock.Anything).
					Run(func(args mock.Arguments) {
						users := args.Get(1).([]*auth.UserToImport)
						assert.Len(t, users, 2)
						for _, p := range users {
							// cheat to access private field params of auth.UserToImport
							rs := reflect.ValueOf(p).Elem()
							rf := rs.Field(0)
							rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
							params := rf.Interface().(map[string]interface{})
							assert.NotEmpty(t, params["localId"].(string))
							assert.NotNil(t, params["customClaims"].(map[string]interface{}))
							assert.Contains(t, []string{"pauline@example.com", "abraham@example.com"}, params["email"].(string))
						}
					}).
					Once().
					Return(&auth.UserImportResult{}, nil)
				firebaseAuth.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				firebaseAuth.
					On("UpdateUser", ctx, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						userId := args.Get(1).(string)
						assert.NotEmpty(t, userId)
						assert.NotContains(t, []string{"student-id-1", "parent-id-1", "parent-id-2"}, userId)
						user := args.Get(2).(*auth.UserToUpdate)

						// cheat to access private field params of auth.UserToUpdate
						rs := reflect.ValueOf(user).Elem()
						rf := rs.Field(0)
						rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
						params := rf.Interface().(map[string]interface{})
						assert.Contains(t, []string{"password-example-1", "password-example-2"}, params["password"].(string))
					}).
					Twice().
					Return(nil, nil)
				jsm.
					On("TracedPublishAsync", mock.Anything, "nats.TracedPublishAsync", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						topic := args[2].(string)
						assert.Equal(t, constants.SubjectUserCreated, topic)

						data := args[3].([]byte)
						msg := &ppb_v1.EvtUser{}
						err := nats.UnmarshalIgnoreMetadata(data, msg)
						require.NoError(t, err)
						assert.Equal(t, "student-id-1", msg.GetCreateParent().StudentId)
						assert.NotEmpty(t, msg.GetCreateParent().ParentId)
						assert.NotContains(t, []string{"student-id-1"}, msg.GetCreateParent().ParentId)
						assert.Equal(t, "Albert Einstein", msg.GetCreateParent().StudentName)
						assert.Equal(t, "1", msg.GetCreateParent().SchoolId)
					}).
					Times(4).
					Return("", nil)
				jsm.
					On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						topic := args[1].(string)
						assert.Equal(t, constants.SubjectUserDeviceTokenUpdated, topic)

						data := args[2].([]byte)
						msg := &ppb_v1.EvtUserInfo{}
						err := proto.Unmarshal(data, msg)
						require.NoError(t, err)
						assert.Equal(t, "student-id-1", msg.GetUserId())
						assert.Equal(t, "Albert Einstein", msg.GetName())
					}).
					Once().
					Return("", nil)
			},
		},
		{
			name: "update user but not change any data",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Albert Einstein",
					Grade:             2,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1@example.com",
				},
				ParentProfiles: []*pb.UpdateStudentRequest_ParentProfile{
					// current parents of this student
					{
						Id:           "parent-id-1",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Email:        "parent-id-1-email@example.com",
					},
					{
						Id:           "parent-id-2",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Email:        "parent-id-2-email@example.com",
					},
				},
				StudentPackageProfiles: []*pb.UpdateStudentRequest_StudentPackageProfile{
					{
						Id: &pb.UpdateStudentRequest_StudentPackageProfile_StudentPackageId{
							StudentPackageId: "studentPackage-id-1",
						},
						StartTime: timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
						EndTime:   timestamppb.New(time.Date(2021, 3, 3, 4, 5, 6, 0, time.UTC)),
					},
					{
						Id: &pb.UpdateStudentRequest_StudentPackageProfile_StudentPackageId{
							StudentPackageId: "studentPackage-id-2",
						},
						StartTime: timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
						EndTime:   timestamppb.New(time.Date(2021, 2, 4, 4, 5, 6, 0, time.UTC)),
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				studentRepo.
					On("Find", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.Student{
						ID:           database.Text("student-id-1"),
						CurrentGrade: database.Int2(int16(2)),
						SchoolID:     database.Int4(1),
					}, nil)
				studentRepo.
					On("UpdateV2", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						student := args.Get(2).(*entities.Student)
						assert.Equal(t, "student-id-1", student.ID.String)
						assert.Equal(t, "Albert Einstein", student.LastName.String)
						assert.Equal(t, int16(2), student.CurrentGrade.Int)
					}).
					Once().
					Return(nil)
				studentParentRepo.
					On("Upsert", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						studentParents := args.Get(2).([]*entities.StudentParent)
						assert.Len(t, studentParents, 2)
						expectedUpdate := map[string]pb.FamilyRelationship{
							"parent-id-1": pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
							"parent-id-2": pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						}

						for _, s := range studentParents {
							assert.Equal(t, "student-id-1", s.StudentID.String)
							if v, ok := expectedUpdate[s.ParentID.String]; ok {
								assert.Equal(t, v.String(), s.Relationship.String)
							} else {
								assert.Fail(t, fmt.Sprintf("not expected parent id %s", s.ParentID.String))
							}
						}
					}).
					Once().
					Return(nil)
				studentParentRepo.On("GetStudentParents", ctx, tx, mock.Anything).Once().Return([]*entities.StudentParent{}, nil)
				parentRepo.
					On("GetByIds", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						parentIDs := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"parent-id-1", "parent-id-2"}, database.FromTextArray(parentIDs))
					}).
					Once().
					Return(entities.Parents{
						parent1,
						parent2,
					}, nil)
				userRepo.
					On("Get", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.User{
						LastName: database.Text("Albert Einstein JR"),
						Email:    database.Text("student-id-1@example.com"),
					}, nil)
				fatimaClient.
					On("EditTimeStudentPackage", signCtx(ctx), mock.Anything).
					Run(func(args mock.Arguments) {
						editReq := args[1].(*fpb.EditTimeStudentPackageRequest)
						expected := map[string]*fpb.EditTimeStudentPackageRequest{
							"studentPackage-id-1": {
								StudentPackageId: "studentPackage-id-1",
								StartAt:          timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
								EndAt:            timestamppb.New(time.Date(2021, 3, 3, 4, 5, 6, 0, time.UTC)),
							},
							"studentPackage-id-2": {
								StudentPackageId: "studentPackage-id-2",
								StartAt:          timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
								EndAt:            timestamppb.New(time.Date(2021, 2, 4, 4, 5, 6, 0, time.UTC)),
							},
						}
						assert.Equal(t, expected[editReq.StudentPackageId].StudentPackageId, editReq.StudentPackageId)
						assert.Equal(t, expected[editReq.StudentPackageId].StartAt.AsTime(), editReq.StartAt.AsTime())
						assert.Equal(t, expected[editReq.StudentPackageId].EndAt.AsTime(), editReq.EndAt.AsTime())
					}).
					Twice().
					Return(&fpb.EditTimeStudentPackageResponse{}, nil)
				jsm.
					On("TracedPublishAsync", mock.Anything, "nats.TracedPublishAsync", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						topic := args[2].(string)
						assert.Equal(t, constants.SubjectUserCreated, topic)

						data := args[3].([]byte)
						msg := &ppb_v1.EvtUser{}
						err := nats.UnmarshalIgnoreMetadata(data, msg)
						require.NoError(t, err)
						assert.Equal(t, "student-id-1", msg.GetCreateParent().StudentId)
						assert.NotEmpty(t, msg.GetCreateParent().ParentId)
						assert.NotContains(t, []string{"student-id-1"}, msg.GetCreateParent().ParentId)
						assert.Equal(t, "Albert Einstein", msg.GetCreateParent().StudentName)
						assert.Equal(t, "1", msg.GetCreateParent().SchoolId)
					}).
					Twice().
					Return("", nil)

				jsm.
					On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						topic := args[1].(string)
						assert.Equal(t, constants.SubjectUserDeviceTokenUpdated, topic)

						data := args[2].([]byte)
						msg := &ppb_v1.EvtUserInfo{}
						err := proto.Unmarshal(data, msg)
						require.NoError(t, err)
						assert.Equal(t, "student-id-1", msg.GetUserId())
						assert.Equal(t, "Albert Einstein", msg.GetName())
					}).
					Once().
					Return("", nil)
			},
		},
		{
			name: "update student with parent which will be created be missed country code",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:               "student-id-1",
					Name:             "Albert Einstein",
					Grade:            2,
					EnrollmentStatus: cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Email:            "student-id-1@example.com",
				},
				ParentProfiles: []*pb.UpdateStudentRequest_ParentProfile{
					// current parents of this student
					{
						Id:           "parent-id-1",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Email:        "parent-id-1-email@example.com",
					},
					{
						Id:           "parent-id-2",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Email:        "parent-id-2-email@example.com",
					},
					// new parents of this student
					{
						Name:         "Pauline Einstein",
						Email:        "pauline@example.com",
						PhoneNumber:  "313-876-3458",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     "password-example-1",
					},
					{
						Name:         "Abraham Einstein",
						Email:        "abraham@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "870-847-9833",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER,
						Password:     "password-example-2",
					},
				},
				StudentPackageProfiles: []*pb.UpdateStudentRequest_StudentPackageProfile{
					{
						Id: &pb.UpdateStudentRequest_StudentPackageProfile_StudentPackageId{
							StudentPackageId: "studentPackage-id-1",
						},
						StartTime: timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
						EndTime:   timestamppb.New(time.Date(2021, 3, 3, 4, 5, 6, 0, time.UTC)),
					},
					{
						Id: &pb.UpdateStudentRequest_StudentPackageProfile_StudentPackageId{
							StudentPackageId: "studentPackage-id-2",
						},
						StartTime: timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
						EndTime:   timestamppb.New(time.Date(2021, 2, 4, 4, 5, 6, 0, time.UTC)),
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				studentRepo.
					On("Find", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.Student{
						ID:           database.Text("student-id-1"),
						CurrentGrade: database.Int2(int16(3)),
						SchoolID:     database.Int4(1),
					}, nil)
				studentRepo.
					On("UpdateV2", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						student := args.Get(2).(*entities.Student)
						assert.Equal(t, "student-id-1", student.ID.String)
						assert.Equal(t, "Albert Einstein", student.LastName.String)
						assert.Equal(t, int16(2), student.CurrentGrade.Int)
					}).
					Once().
					Return(nil)
				studentParentRepo.
					On("Upsert", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						studentParents := args.Get(2).([]*entities.StudentParent)
						assert.Len(t, studentParents, 4)
						expectedUpdate := map[string]pb.FamilyRelationship{
							"parent-id-1": pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
							"parent-id-2": pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						}

						numberUpdate := 0
						for _, s := range studentParents {
							assert.Equal(t, "student-id-1", s.StudentID.String)
							if v, ok := expectedUpdate[s.ParentID.String]; ok {
								assert.Equal(t, v.String(), s.Relationship.String)
								numberUpdate++
							} else {
								// assign new parents
								assert.Contains(t,
									[]string{
										pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER.String(),
										pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER.String(),
									},
									s.Relationship.String,
								)
							}
						}
						assert.Equal(t, len(expectedUpdate), numberUpdate)
					}).
					Once().
					Return(nil)
				studentParentRepo.On("GetStudentParents", ctx, tx, mock.Anything).Once().Return([]*entities.StudentParent{}, nil)
				parentRepo.
					On("GetByIds", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						parentIDs := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"parent-id-1", "parent-id-2"}, database.FromTextArray(parentIDs))
					}).
					Once().
					Return(entities.Parents{
						parent1,
						parent2,
					}, nil)
				parentRepo.
					On("CreateMultiple", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						parents := args.Get(2).([]*entities.Parent)
						assert.Len(t, parents, 2)
						for _, p := range parents {
							assert.Equal(t, int32(1), p.SchoolID.Int)
							assert.NotEmpty(t, p.ID.String)
						}
					}).
					Once().
					Return(nil)
				userGroupRepo.
					On("CreateMultiple", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						userGr := args.Get(2).([]*entities.UserGroup)
						assert.Len(t, userGr, 2)
						for _, p := range userGr {
							assert.NotEmpty(t, p.UserID.String)
							assert.Equal(t, entities.UserGroupParent, p.GroupID.String)
							assert.True(t, p.IsOrigin.Bool)
							assert.Equal(t, entities.UserGroupStatusActive, p.Status.String)
						}
					}).
					Once().
					Return(nil)
				userRepo.
					On("Get", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.User{
						LastName: database.Text("Albert Einstein JR"),
						Email:    database.Text("student-id-1@example.com"),
					}, nil)
				userRepo.
					On("GetByEmail", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						emails := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"pauline@example.com", "abraham@example.com"}, database.FromTextArray(emails))
					}).
					Once().
					Return(nil, nil)
				userRepo.
					On("GetByPhone", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						phoneNumbers := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"313-876-3458", "870-847-9833"}, database.FromTextArray(phoneNumbers))
					}).
					Once().
					Return(nil, nil)
				userRepo.
					On("CreateMultiple", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						parents := args.Get(2).([]*entities.User)
						assert.Len(t, parents, 2)
						expectedByName := map[string]*pb.UpdateStudentRequest_ParentProfile{
							"Pauline Einstein": {
								Name:        "Pauline Einstein",
								Email:       "pauline@example.com",
								CountryCode: cpb.Country_COUNTRY_NONE,
								PhoneNumber: "313-876-3458",
							},
							"Abraham Einstein": {
								Name:        "Abraham Einstein",
								Email:       "abraham@example.com",
								CountryCode: cpb.Country_COUNTRY_JP,
								PhoneNumber: "870-847-9833",
							},
						}
						for _, p := range parents {
							name := p.LastName.String
							assert.NotNil(t, expectedByName[name])
							assert.Equal(t, expectedByName[name].Email, p.Email.String)
							assert.Equal(t, expectedByName[name].CountryCode.String(), p.Country.String)
							assert.Equal(t, expectedByName[name].PhoneNumber, p.PhoneNumber.String)
						}
					}).
					Once().
					Return(nil)
				firebaseAuth.
					On("ImportUsers", ctx, mock.Anything).
					Run(func(args mock.Arguments) {
						users := args.Get(1).([]*auth.UserToImport)
						assert.Len(t, users, 2)
						for _, p := range users {
							// cheat to access private field params of auth.UserToImport
							rs := reflect.ValueOf(p).Elem()
							rf := rs.Field(0)
							rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
							params := rf.Interface().(map[string]interface{})
							assert.NotEmpty(t, params["localId"].(string))
							assert.NotNil(t, params["customClaims"].(map[string]interface{}))
							assert.Contains(t, []string{"pauline@example.com", "abraham@example.com"}, params["email"].(string))
						}
					}).
					Once().
					Return(&auth.UserImportResult{}, nil)
				firebaseAuth.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				firebaseAuth.
					On("UpdateUser", ctx, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						userId := args.Get(1).(string)
						assert.NotEmpty(t, userId)
						assert.NotContains(t, []string{"student-id-1", "parent-id-1", "parent-id-2"}, userId)
						user := args.Get(2).(*auth.UserToUpdate)

						// cheat to access private field params of auth.UserToUpdate
						rs := reflect.ValueOf(user).Elem()
						rf := rs.Field(0)
						rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
						params := rf.Interface().(map[string]interface{})
						assert.Contains(t, []string{"password-example-1", "password-example-2"}, params["password"].(string))
					}).
					Twice().
					Return(nil, nil)
				jsm.
					On("TracedPublishAsync", mock.Anything, "nats.TracedPublishAsync", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						topic := args[2].(string)
						assert.Equal(t, constants.SubjectUserCreated, topic)

						data := args[3].([]byte)
						msg := &ppb_v1.EvtUser{}
						err := nats.UnmarshalIgnoreMetadata(data, msg)
						require.NoError(t, err)
						assert.Equal(t, "student-id-1", msg.GetCreateParent().StudentId)
						assert.NotEmpty(t, msg.GetCreateParent().ParentId)
						assert.NotContains(t, []string{"student-id-1"}, msg.GetCreateParent().ParentId)
						assert.Equal(t, "Albert Einstein", msg.GetCreateParent().StudentName)
						assert.Equal(t, "1", msg.GetCreateParent().SchoolId)
					}).
					Times(4).
					Return("", nil)
				jsm.
					On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						topic := args[1].(string)
						assert.Equal(t, constants.SubjectUserDeviceTokenUpdated, topic)

						data := args[2].([]byte)
						msg := &ppb_v1.EvtUserInfo{}
						err := proto.Unmarshal(data, msg)
						require.NoError(t, err)
						assert.Equal(t, "student-id-1", msg.GetUserId())
						assert.Equal(t, "Albert Einstein", msg.GetName())
					}).
					Once().
					Return("", nil)
				fatimaClient.
					On("EditTimeStudentPackage", signCtx(ctx), mock.Anything).
					Run(func(args mock.Arguments) {
						editReq := args[1].(*fpb.EditTimeStudentPackageRequest)
						expected := map[string]*fpb.EditTimeStudentPackageRequest{
							"studentPackage-id-1": {
								StudentPackageId: "studentPackage-id-1",
								StartAt:          timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
								EndAt:            timestamppb.New(time.Date(2021, 3, 3, 4, 5, 6, 0, time.UTC)),
							},
							"studentPackage-id-2": {
								StudentPackageId: "studentPackage-id-2",
								StartAt:          timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
								EndAt:            timestamppb.New(time.Date(2021, 2, 4, 4, 5, 6, 0, time.UTC)),
							},
						}
						assert.Equal(t, expected[editReq.StudentPackageId].StudentPackageId, editReq.StudentPackageId)
						assert.Equal(t, expected[editReq.StudentPackageId].StartAt.AsTime(), editReq.StartAt.AsTime())
						assert.Equal(t, expected[editReq.StudentPackageId].EndAt.AsTime(), editReq.EndAt.AsTime())
					}).
					Twice().
					Return(&fpb.EditTimeStudentPackageResponse{}, nil)
			},
		},
		// invalid param cases
		// invalid user's info
		{
			name: "could not update with student not found",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Albert Einstein",
					Grade:             2,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1@example.com",
				},
				ParentProfiles: []*pb.UpdateStudentRequest_ParentProfile{
					// current parents of this student
					{
						Id:           "parent-id-1",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Email:        "parent-id-1-email@example.com",
					},
					{
						Id:           "parent-id-2",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Email:        "parent-id-2-email@example.com",
					},
					// new parents of this student
					{
						Name:         "Pauline Einstein",
						Email:        "pauline@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "313-876-3458",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     "password-example-1",
					},
					{
						Name:         "Abraham Einstein",
						Email:        "abraham@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "870-847-9833",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER,
						Password:     "password-example-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				studentRepo.
					On("Find", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(nil, pgx.ErrNoRows)
			},
			hasError: true,
		},
		{
			name: "could not update student missing name",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "",
					Grade:             2,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1@example.com",
				},
				ParentProfiles: []*pb.UpdateStudentRequest_ParentProfile{
					// current parents of this student
					{
						Id:           "parent-id-1",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Email:        "parent-id-1-email@example.com",
					},
					{
						Id:           "parent-id-2",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Email:        "parent-id-2-email@example.com",
					},
					// new parents of this student
					{
						Name:         "Pauline Einstein",
						Email:        "pauline@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "313-876-3458",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     "password-example-1",
					},
					{
						Name:         "Abraham Einstein",
						Email:        "abraham@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "870-847-9833",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER,
						Password:     "password-example-2",
					},
				},
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
		{
			name: "could not update student missing enrollment status",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:    "student-id-1",
					Name:  "Albert Einstein",
					Grade: 2,
					// EnrollmentStatus: cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1@example.com",
				},
				ParentProfiles: []*pb.UpdateStudentRequest_ParentProfile{
					// current parents of this student
					{
						Id:           "parent-id-1",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Email:        "parent-id-1-email@example.com",
					},
					{
						Id:           "parent-id-2",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Email:        "parent-id-2-email@example.com",
					},
					// new parents of this student
					{
						Name:         "Pauline Einstein",
						Email:        "pauline@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "313-876-3458",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     "password-example-1",
					},
					{
						Name:         "Abraham Einstein",
						Email:        "abraham@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "870-847-9833",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER,
						Password:     "password-example-2",
					},
				},
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
		{
			name: "could not update student invalid grade",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "",
					Grade:             -1,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1@example.com",
				},
				ParentProfiles: []*pb.UpdateStudentRequest_ParentProfile{
					// current parents of this student
					{
						Id:           "parent-id-1",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Email:        "parent-id-1-email@example.com",
					},
					{
						Id:           "parent-id-2",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Email:        "parent-id-2-email@example.com",
					},
					// new parents of this student
					{
						Name:         "Pauline Einstein",
						Email:        "pauline@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "313-876-3458",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     "password-example-1",
					},
					{
						Name:         "Abraham Einstein",
						Email:        "abraham@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "870-847-9833",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER,
						Password:     "password-example-2",
					},
				},
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
		{
			name: "could not update student with invalid enrollment status",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "",
					Grade:             -1,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus(999999), // 999999 - unknown enrollment status
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1@example.com",
				},
				ParentProfiles: []*pb.UpdateStudentRequest_ParentProfile{
					// current parents of this student
					{
						Id:           "parent-id-1",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Email:        "parent-id-1-email@example.com",
					},
					{
						Id:           "parent-id-2",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Email:        "parent-id-2-email@example.com",
					},
					// new parents of this student
					{
						Name:         "Pauline Einstein",
						Email:        "pauline@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "313-876-3458",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     "password-example-1",
					},
					{
						Name:         "Abraham Einstein",
						Email:        "abraham@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "870-847-9833",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER,
						Password:     "password-example-2",
					},
				},
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
		{
			name: "could not update student missing student email",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Albert Einstein",
					Grade:             2,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					// Email:             "student-id-1@example.com",
				},
				ParentProfiles: []*pb.UpdateStudentRequest_ParentProfile{
					// current parents of this student
					{
						Id:           "parent-id-1",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Email:        "parent-id-1-email@example.com",
					},
					{
						Id:           "parent-id-2",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Email:        "parent-id-2-email@example.com",
					},
					// new parents of this student
					{
						Name:         "Pauline Einstein",
						Email:        "pauline@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "313-876-3458",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     "password-example-1",
					},
					{
						Name:         "Abraham Einstein",
						Email:        "abraham@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "870-847-9833",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER,
						Password:     "password-example-2",
					},
				},
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
		{
			name: "could not update student with student email the same as parent email",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Albert Einstein",
					Grade:             2,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "parent-id-1-email@example.com",
				},
				ParentProfiles: []*pb.UpdateStudentRequest_ParentProfile{
					// current parents of this student
					{
						Id:           "parent-id-1",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Email:        "parent-id-1-email@example.com",
					},
					{
						Id:           "parent-id-2",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Email:        "parent-id-2-email@example.com",
					},
					// new parents of this student
					{
						Name:         "Pauline Einstein",
						Email:        "pauline@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "313-876-3458",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     "password-example-1",
					},
					{
						Name:         "Abraham Einstein",
						Email:        "abraham@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "870-847-9833",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER,
						Password:     "password-example-2",
					},
				},
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
		// invalid student's info which will be updated
		{
			name: "could not update student whose email will be edited to an existing email",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Albert Einstein",
					Grade:             2,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1-existing@example.com",
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				studentRepo.
					On("Find", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.Student{
						ID:           database.Text("student-id-1"),
						CurrentGrade: database.Int2(int16(3)),
						SchoolID:     database.Int4(1),
					}, nil)
				userRepo.
					On("Get", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.User{
						LastName: database.Text("Albert Einstein JR"),
						Email:    database.Text("student-id-1@example.com"),
					}, nil)
				userRepo.
					On("GetByEmail", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						emails := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t,
							[]string{"student-id-1-existing@example.com"},
							database.FromTextArray(emails))
					}).
					Once().
					Return([]*entities.User{
						{
							Email: pgtype.Text{
								String: "student-id-1-existing@example.com",
								Status: pgtype.Present,
							},
						},
					}, nil)
			},
			hasError: true,
		},
		// invalid parent's info which will be created
		{
			name: "could not update student with parent which will be created be existing email",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Albert Einstein",
					Grade:             2,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1@example.com",
				},
				ParentProfiles: []*pb.UpdateStudentRequest_ParentProfile{
					// current parents of this student
					{
						Id:           "parent-id-1",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Email:        "parent-id-1-email@example.com",
					},
					{
						Id:           "parent-id-2",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Email:        "parent-id-2-email@example.com",
					},
					// new parents of this student
					{
						Name:         "Pauline Einstein",
						Email:        "pauline@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "313-876-3458",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     "password-example-1",
					},
					{
						Name:         "Abraham Einstein",
						Email:        "abraham@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "870-847-9833",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER,
						Password:     "password-example-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				studentRepo.
					On("Find", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.Student{
						ID:           database.Text("student-id-1"),
						CurrentGrade: database.Int2(int16(3)),
						SchoolID:     database.Int4(1),
					}, nil)
				studentRepo.
					On("UpdateV2", ctx, tx, mock.Anything).
					Once().
					Return(nil)
				userRepo.
					On("Get", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.User{
						LastName: database.Text("Albert Einstein JR"),
						Email:    database.Text("student-id-1@example.com"),
					}, nil)
				userRepo.
					On("GetByEmail", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						emails := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"pauline@example.com", "abraham@example.com"}, database.FromTextArray(emails))
					}).
					Once().
					Return([]*entities.User{
						{
							ID: database.Text("existing-id"),
						},
					}, nil)
			},
			hasError: true,
		},
		{
			name: "could not update student with parent whose email will be edited to an existing email",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Albert Einstein",
					Grade:             2,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1@example.com",
				},
				ParentProfiles: []*pb.UpdateStudentRequest_ParentProfile{
					// current parents of this student
					{
						Id:           "parent-id-1",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Email:        "parent-id-1-email@example.com",
					},
					// parent with edited email
					{
						Id:           "parent-id-2",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Email:        "parent-id-2-email-existing@example.com", // an existed email
					},
					// new parents of this student
					{
						Name:         "Pauline Einstein",
						Email:        "pauline@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "313-876-3458",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     "password-example-1",
					},
					{
						Name:         "Abraham Einstein",
						Email:        "abraham@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "870-847-9833",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER,
						Password:     "password-example-1",
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				studentRepo.
					On("Find", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.Student{
						ID:           database.Text("student-id-1"),
						CurrentGrade: database.Int2(int16(3)),
						SchoolID:     database.Int4(1),
					}, nil)
				studentRepo.
					On("UpdateV2", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						student := args.Get(2).(*entities.Student)
						assert.Equal(t, "student-id-1", student.ID.String)
						assert.Equal(t, "Albert Einstein", student.LastName.String)
						assert.Equal(t, int16(2), student.CurrentGrade.Int)
					}).
					Once().
					Return(nil)
				parentRepo.
					On("GetByIds", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						parentIDs := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"parent-id-1", "parent-id-2"}, database.FromTextArray(parentIDs))
					}).
					Once().
					Return(entities.Parents{
						parent1,
						&entities.Parent{
							ID:       database.Text("parent-id-2"),
							SchoolID: database.Int4(1),
							User: entities.User{
								Email: pgtype.Text{
									String: "parent-id-2-email@example.com",
									Status: pgtype.Present,
								},
							},
						},
					}, nil)
				parentRepo.
					On("CreateMultiple", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						parents := args.Get(2).([]*entities.Parent)
						assert.Len(t, parents, 2)
						for _, p := range parents {
							assert.Equal(t, int32(1), p.SchoolID.Int)
							assert.NotEmpty(t, p.ID.String)
						}
					}).
					Once().
					Return(nil)
				userGroupRepo.
					On("CreateMultiple", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						userGr := args.Get(2).([]*entities.UserGroup)
						assert.Len(t, userGr, 2)
						for _, p := range userGr {
							assert.NotEmpty(t, p.UserID.String)
							assert.Equal(t, entities.UserGroupParent, p.GroupID.String)
							assert.True(t, p.IsOrigin.Bool)
							assert.Equal(t, entities.UserGroupStatusActive, p.Status.String)
						}
					}).
					Once().
					Return(nil)
				userRepo.
					On("Get", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.User{
						LastName: database.Text("Albert Einstein JR"),
						Email:    database.Text("student-id-1@example.com"),
					}, nil)
				userRepo.
					On("GetByEmail", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						emails := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t,
							[]string{"pauline@example.com", "abraham@example.com"},
							database.FromTextArray(emails))
					}).
					Once().
					Return(nil, nil)
				userRepo.
					On("GetByEmail", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						emails := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t,
							[]string{"parent-id-2-email-existing@example.com"},
							database.FromTextArray(emails))
					}).
					Once().
					Return([]*entities.User{
						{
							Email: pgtype.Text{
								String: "parent-id-2-email-existing@example.com",
								Status: pgtype.Present,
							},
						},
					}, nil)
				userRepo.
					On("GetByPhone", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						phoneNumbers := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"313-876-3458", "870-847-9833"}, database.FromTextArray(phoneNumbers))
					}).
					Once().
					Return(nil, nil)
				userRepo.
					On("CreateMultiple", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						parents := args.Get(2).([]*entities.User)
						assert.Len(t, parents, 2)
						expectedByName := map[string]*pb.UpdateStudentRequest_ParentProfile{
							"Pauline Einstein": {
								Name:        "Pauline Einstein",
								Email:       "pauline@example.com",
								CountryCode: cpb.Country_COUNTRY_JP,
								PhoneNumber: "313-876-3458",
							},
							"Abraham Einstein": {
								Name:        "Abraham Einstein",
								Email:       "abraham@example.com",
								CountryCode: cpb.Country_COUNTRY_JP,
								PhoneNumber: "870-847-9833",
							},
						}
						for _, p := range parents {
							name := p.LastName.String
							assert.NotNil(t, expectedByName[name])
							assert.Equal(t, expectedByName[name].Email, p.Email.String)
							assert.Equal(t, expectedByName[name].CountryCode.String(), p.Country.String)
							assert.Equal(t, expectedByName[name].PhoneNumber, p.PhoneNumber.String)
						}
					}).
					Once().
					Return(nil)
				firebaseAuth.
					On("ImportUsers", ctx, mock.Anything).
					Run(func(args mock.Arguments) {
						users := args.Get(1).([]*auth.UserToImport)
						assert.Len(t, users, 2)
						for _, p := range users {
							// cheat to access private field params of auth.UserToImport
							rs := reflect.ValueOf(p).Elem()
							rf := rs.Field(0)
							rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
							params := rf.Interface().(map[string]interface{})
							assert.NotEmpty(t, params["localId"].(string))
							assert.NotNil(t, params["customClaims"].(map[string]interface{}))
							assert.Contains(t, []string{"pauline@example.com", "abraham@example.com"}, params["email"].(string))
						}
					}).
					Once().
					Return(&auth.UserImportResult{}, nil)
				firebaseAuth.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				firebaseAuth.
					On("UpdateUser", ctx, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						userId := args.Get(1).(string)
						assert.NotEmpty(t, userId)
						assert.NotContains(t, []string{"student-id-1", "parent-id-1", "parent-id-2"}, userId)
						user := args.Get(2).(*auth.UserToUpdate)

						// cheat to access private field params of auth.UserToUpdate
						rs := reflect.ValueOf(user).Elem()
						rf := rs.Field(0)
						rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
						params := rf.Interface().(map[string]interface{})
						assert.Contains(t, []string{"password-example-1", "password-example-2"}, params["password"].(string))
					}).
					Twice().
					Return(nil, nil)
			},
			hasError: true,
		},
		{
			name: "could not update student with parent which will be created be missed email",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Albert Einstein",
					Grade:             2,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1@example.com",
				},
				ParentProfiles: []*pb.UpdateStudentRequest_ParentProfile{
					// current parents of this student
					{
						Id:           "parent-id-1",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Email:        "parent-id-1-email@example.com",
					},
					{
						Id:           "parent-id-2",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Email:        "parent-id-2-email@example.com",
					},
					// new parents of this student
					{
						Name:         "Pauline Einstein",
						Email:        "pauline@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "313-876-3458",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
					},
					{
						Name:         "Abraham Einstein",
						Email:        "",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "870-847-9833",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER,
					},
				},
			},
			setup: func(ctx context.Context) {
				studentRepo.
					On("Find", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.Student{
						ID:           database.Text("student-id-1"),
						CurrentGrade: database.Int2(int16(3)),
						SchoolID:     database.Int4(1),
					}, nil)
				userRepo.
					On("Get", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.User{
						LastName: database.Text("Albert Einstein JR"),
						Email:    database.Text("student-id-1@example.com"),
					}, nil)
			},
			hasError: true,
		},
		{
			name: "could not update student with parent which will be edited missing email",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Albert Einstein",
					Grade:             2,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1@example.com",
				},
				ParentProfiles: []*pb.UpdateStudentRequest_ParentProfile{
					// current parents of this student
					{
						Id:           "parent-id-1",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Email:        "parent-id-1-email@example.com",
					},
					{
						Id:           "parent-id-2",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Email:        "",
					},
				},
			},
			setup: func(ctx context.Context) {
				studentRepo.
					On("Find", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.Student{
						ID:           database.Text("student-id-1"),
						CurrentGrade: database.Int2(int16(3)),
						SchoolID:     database.Int4(1),
					}, nil)
				userRepo.
					On("Get", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.User{
						LastName: database.Text("Albert Einstein JR"),
						Email:    database.Text("student-id-1@example.com"),
					}, nil)
			},
			hasError: true,
		},
		{
			name: "could not update student with parent which will be created be missed name",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Albert Einstein",
					Grade:             2,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1@example.com",
				},
				ParentProfiles: []*pb.UpdateStudentRequest_ParentProfile{
					// current parents of this student
					{
						Id:           "parent-id-1",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Email:        "parent-id-1-email@example.com",
					},
					{
						Id:           "parent-id-2",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Email:        "parent-id-2-email@example.com",
					},
					// new parents of this student
					{
						Name:         "Pauline Einstein",
						Email:        "pauline@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "313-876-3458",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     "password-example-1",
					},
					{
						Name:         "",
						Email:        "abraham@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "870-847-9833",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER,
						Password:     "password-example-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				studentRepo.
					On("Find", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.Student{
						ID:           database.Text("student-id-1"),
						CurrentGrade: database.Int2(int16(3)),
						SchoolID:     database.Int4(1),
					}, nil)
				userRepo.
					On("Get", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.User{
						LastName: database.Text("Albert Einstein JR"),
						Email:    database.Text("student-id-1@example.com"),
					}, nil)
			},
			hasError: true,
		},
		{
			name: "could not update student with parent which will be created have existing phone number",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Albert Einstein",
					Grade:             2,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1@example.com",
				},
				ParentProfiles: []*pb.UpdateStudentRequest_ParentProfile{
					// current parents of this student
					{
						Id:           "parent-id-1",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Email:        "parent-id-1-email@example.com",
					},
					{
						Id:           "parent-id-2",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Email:        "parent-id-2-email@example.com",
					},
					// new parents of this student
					{
						Name:         "Pauline Einstein",
						Email:        "pauline@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "313-876-3458",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     "password-example-1",
					},
					{
						Name:         "Abraham Einstein",
						Email:        "abraham@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "870-847-9833",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER,
						Password:     "password-example-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				studentRepo.
					On("Find", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.Student{
						ID:           database.Text("student-id-1"),
						CurrentGrade: database.Int2(int16(3)),
						SchoolID:     database.Int4(1),
					}, nil)
				studentRepo.
					On("UpdateV2", ctx, tx, mock.Anything).
					Once().
					Return(nil)
				userRepo.
					On("Get", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.User{
						LastName: database.Text("Albert Einstein JR"),
						Email:    database.Text("student-id-1@example.com"),
					}, nil)
				userRepo.
					On("GetByEmail", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						emails := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"pauline@example.com", "abraham@example.com"}, database.FromTextArray(emails))
					}).
					Once().
					Return(nil, nil)
				userRepo.
					On("GetByPhone", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						phoneNumbers := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"313-876-3458", "870-847-9833"}, database.FromTextArray(phoneNumbers))
					}).
					Once().
					Return([]*entities.User{
						{
							ID: database.Text("existing-id"),
						},
					}, nil)
			},
			hasError: true,
		},
		{
			name: "could not update student with parents which will be created have same email",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Albert Einstein",
					Grade:             2,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1@example.com",
				},
				ParentProfiles: []*pb.UpdateStudentRequest_ParentProfile{
					// current parents of this student
					{
						Id:           "parent-id-1",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Email:        "parent-id-1-email@example.com",
					},
					{
						Id:           "parent-id-2",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Email:        "parent-id-2-email@example.com",
					},
					// new parents of this student
					{
						Name:         "Pauline Einstein",
						Email:        "pauline@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "313-876-3458",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     "password-example-1",
					},
					{
						Name:         "Abraham Einstein",
						Email:        "pauline@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "870-847-9833",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER,
						Password:     "password-example-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				studentRepo.
					On("Find", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.Student{
						ID:           database.Text("student-id-1"),
						CurrentGrade: database.Int2(int16(3)),
						SchoolID:     database.Int4(1),
					}, nil)
				studentRepo.
					On("UpdateV2", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						student := args.Get(2).(*entities.Student)
						assert.Equal(t, "student-id-1", student.ID.String)
						assert.Equal(t, "Albert Einstein", student.LastName.String)
						assert.Equal(t, int16(2), student.CurrentGrade.Int)
					}).
					Once().
					Return(nil)
				userRepo.
					On("Get", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.User{
						LastName: database.Text("Albert Einstein JR"),
						Email:    database.Text("student-id-1@example.com"),
					}, nil)
				userRepo.
					On("GetByEmail", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						emails := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"pauline@example.com", "pauline@example.com"}, database.FromTextArray(emails))
					}).
					Once().
					Return(nil, nil)
				userRepo.
					On("GetByPhone", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						phoneNumbers := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"313-876-3458", "870-847-9833"}, database.FromTextArray(phoneNumbers))
					}).
					Once().
					Return(nil, nil)
				userRepo.
					On("CreateMultiple", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						parents := args.Get(2).([]*entities.User)
						assert.Len(t, parents, 2)
						expectedByName := map[string]*pb.UpdateStudentRequest_ParentProfile{
							"Pauline Einstein": {
								Name:        "Pauline Einstein",
								Email:       "pauline@example.com",
								CountryCode: cpb.Country_COUNTRY_JP,
								PhoneNumber: "313-876-3458",
							},
							"Abraham Einstein": {
								Name:        "Abraham Einstein",
								Email:       "pauline@example.com",
								CountryCode: cpb.Country_COUNTRY_JP,
								PhoneNumber: "870-847-9833",
							},
						}
						for _, p := range parents {
							name := p.LastName.String
							assert.NotNil(t, expectedByName[name])
							assert.Equal(t, expectedByName[name].Email, p.Email.String)
							assert.Equal(t, expectedByName[name].CountryCode.String(), p.Country.String)
							assert.Equal(t, expectedByName[name].PhoneNumber, p.PhoneNumber.String)
						}
					}).
					Once().
					Return(fmt.Errorf("duplicated unique key"))
			},
			hasError: true,
		},
		// invalid existing parents
		{
			name: "could not update student with non-existing parent",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Albert Einstein",
					Grade:             2,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1@example.com",
				},
				ParentProfiles: []*pb.UpdateStudentRequest_ParentProfile{
					// current parent of this student
					{
						Id:           "parent-id-1",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Email:        "parent-id-1-email@example.com",
					},
					// non-existing parent
					{
						Id:           "parent-id-2",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Email:        "parent-id-2-email@example.com",
					},
					// new parents of this student
					{
						Name:         "Pauline Einstein",
						Email:        "pauline@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "313-876-3458",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     "password-example-1",
					},
					{
						Name:         "Abraham Einstein",
						Email:        "abraham@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "870-847-9833",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER,
						Password:     "password-example-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				studentRepo.
					On("Find", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.Student{
						ID:           database.Text("student-id-1"),
						CurrentGrade: database.Int2(int16(3)),
						SchoolID:     database.Int4(1),
					}, nil)
				studentRepo.
					On("UpdateV2", ctx, tx, mock.Anything).
					Once().
					Return(nil)
				parentRepo.
					On("GetByIds", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						parentIDs := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"parent-id-1", "parent-id-2"}, database.FromTextArray(parentIDs))
					}).
					Once().
					Return(entities.Parents{
						parent1,
					}, nil)
				parentRepo.
					On("CreateMultiple", ctx, tx, mock.Anything).
					Once().
					Return(nil)
				userGroupRepo.
					On("CreateMultiple", ctx, tx, mock.Anything).
					Once().
					Return(nil)
				userRepo.
					On("GetByEmail", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						emails := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"pauline@example.com", "abraham@example.com"}, database.FromTextArray(emails))
					}).
					Once().
					Return(nil, nil)
				userRepo.
					On("GetByPhone", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						phoneNumbers := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"313-876-3458", "870-847-9833"}, database.FromTextArray(phoneNumbers))
					}).
					Once().
					Return(nil, nil)
				userRepo.
					On("Get", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.User{
						LastName: database.Text("Albert Einstein JR"),
						Email:    database.Text("student-id-1@example.com"),
					}, nil)
				userRepo.
					On("CreateMultiple", ctx, tx, mock.Anything).
					Once().
					Return(nil)
				firebaseAuth.
					On("ImportUsers", ctx, mock.Anything).
					Once().
					Return(&auth.UserImportResult{}, nil)
				firebaseAuth.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				firebaseAuth.
					On("UpdateUser", ctx, mock.Anything, mock.Anything).
					Twice().
					Return(nil, nil)
			},
			hasError: true,
		},
		{
			name: "could not update student with existing parents which have different school id",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Albert Einstein",
					Grade:             2,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1@example.com",
				},
				ParentProfiles: []*pb.UpdateStudentRequest_ParentProfile{
					// current parents of this student
					{
						Id:           "parent-id-1",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Email:        "parent-id-1-email@example.com",
					},
					{
						Id:           "parent-id-2",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Email:        "parent-id-2-email@example.com",
					},
					// new parents of this student
					{
						Name:         "Pauline Einstein",
						Email:        "pauline@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "313-876-3458",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     "password-example-1",
					},
					{
						Name:         "Abraham Einstein",
						Email:        "abraham@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "870-847-9833",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER,
						Password:     "password-example-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				studentRepo.
					On("Find", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.Student{
						ID:           database.Text("student-id-1"),
						CurrentGrade: database.Int2(int16(3)),
						SchoolID:     database.Int4(1),
					}, nil)
				studentRepo.
					On("UpdateV2", ctx, tx, mock.Anything).
					Once().
					Return(nil)
				parentRepo.
					On("GetByIds", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						parentIDs := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"parent-id-1", "parent-id-2"}, database.FromTextArray(parentIDs))
					}).
					Once().
					Return(entities.Parents{
						parent1,
						&entities.Parent{
							ID:       database.Text("parent-id-2"),
							SchoolID: database.Int4(2),
							User: entities.User{
								Email: pgtype.Text{
									String: "parent-id-2-email@example.com",
									Status: pgtype.Present,
								},
							},
						},
					}, nil)
				parentRepo.
					On("CreateMultiple", ctx, tx, mock.Anything).
					Once().
					Return(nil)
				userGroupRepo.
					On("CreateMultiple", ctx, tx, mock.Anything).
					Once().
					Return(nil)
				userRepo.
					On("GetByEmail", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						emails := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"pauline@example.com", "abraham@example.com"}, database.FromTextArray(emails))
					}).
					Once().
					Return(nil, nil)
				userRepo.
					On("GetByPhone", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						phoneNumbers := args.Get(2).(pgtype.TextArray)
						assert.ElementsMatch(t, []string{"313-876-3458", "870-847-9833"}, database.FromTextArray(phoneNumbers))
					}).
					Once().
					Return(nil, nil)
				userRepo.
					On("Get", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						id := args.Get(2).(pgtype.Text)
						assert.Equal(t, "student-id-1", id.String)
					}).
					Once().
					Return(&entities.User{
						LastName: database.Text("Albert Einstein JR"),
						Email:    database.Text("student-id-1@example.com"),
					}, nil)
				userRepo.
					On("CreateMultiple", ctx, tx, mock.Anything).
					Once().
					Return(nil)
				firebaseAuth.
					On("ImportUsers", ctx, mock.Anything).
					Once().
					Return(&auth.UserImportResult{}, nil)
				firebaseAuth.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				firebaseAuth.
					On("UpdateUser", ctx, mock.Anything, mock.Anything).
					Twice().
					Return(nil, nil)
			},
			hasError: true,
		},
		// invalid student package profile
		{
			name: "update student with start time of StudentPackage after end time",
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Albert Einstein",
					Grade:             2,
					EnrollmentStatus:  cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1@example.com",
				},
				ParentProfiles: []*pb.UpdateStudentRequest_ParentProfile{
					// current parents of this student
					{
						Id:           "parent-id-1",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GUARDIAN,
						Email:        "parent-id-1-email@example.com",
					},
					{
						Id:           "parent-id-2",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
						Email:        "parent-id-2-email@example.com",
					},
					// new parents of this student
					{
						Name:         "Pauline Einstein",
						Email:        "pauline@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "313-876-3458",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_MOTHER,
						Password:     "password-example-1",
					},
					{
						Name:         "Abraham Einstein",
						Email:        "abraham@example.com",
						CountryCode:  cpb.Country_COUNTRY_JP,
						PhoneNumber:  "870-847-9833",
						Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_GRANDFATHER,
						Password:     "password-example-1",
					},
				},
				StudentPackageProfiles: []*pb.UpdateStudentRequest_StudentPackageProfile{
					{
						Id: &pb.UpdateStudentRequest_StudentPackageProfile_StudentPackageId{
							StudentPackageId: "studentPackage-id-1",
						},
						StartTime: timestamppb.New(time.Date(2021, 2, 3, 4, 5, 6, 0, time.UTC)),
						EndTime:   timestamppb.New(time.Date(2021, 3, 3, 4, 5, 6, 0, time.UTC)),
					},
					{
						Id: &pb.UpdateStudentRequest_StudentPackageProfile_StudentPackageId{
							StudentPackageId: "studentPackage-id-2",
						},
						StartTime: timestamppb.New(time.Date(2021, 4, 3, 4, 5, 6, 0, time.UTC)),
						EndTime:   timestamppb.New(time.Date(2021, 2, 4, 4, 5, 6, 0, time.UTC)),
					},
				},
			},
			setup: func(ctx context.Context) {
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Log("Test case: " + tc.name)
			tc.setup(context.Background())

			s := UserModifierService{
				DB:                db,
				JSM:               jsm,
				UserRepo:          userRepo,
				StudentRepo:       studentRepo,
				ParentRepo:        parentRepo,
				UserGroupRepo:     userGroupRepo,
				StudentParentRepo: studentParentRepo,
				FatimaClient:      fatimaClient,
				FirebaseClient:    firebaseAuth,
			}
			_, err := s.UpdateStudent(context.Background(), tc.req)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(
				t,
				db,
				tx,
				studentRepo,
				userRepo,
				parentRepo,
				userGroupRepo,
				studentParentRepo,
				fatimaClient,
				firebaseAuth,
				jsm,
			)
		})
	}
}

type transportResult struct {
	res *http.Response
	err error
}

type mockTransport struct {
	gotReq  *http.Request
	gotBody []byte
	results []transportResult
}

func (t *mockTransport) addResult(res *http.Response, err error) {
	t.results = append(t.results, transportResult{res, err})
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.gotReq = req
	t.gotBody = nil
	if req.Body != nil {
		bytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		t.gotBody = bytes
	}
	if len(t.results) == 0 {
		return nil, fmt.Errorf("error handling request")
	}
	result := t.results[0]
	t.results = t.results[1:]
	return result.res, result.err
}

func (t *mockTransport) gotJSONBody() map[string]interface{} {
	m := map[string]interface{}{}
	if err := json.Unmarshal(t.gotBody, &m); err != nil {
		panic(err)
	}
	return m
}

func mockClient(t *testing.T, m *mockTransport) *storage.Client {
	client, err := storage.NewClient(context.Background(), option.WithHTTPClient(&http.Client{Transport: m}))
	if err != nil {
		t.Fatal(err)
	}
	return client
}
func bodyReader(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}

func Test_uploadToCloudStorage(t *testing.T) {
	bucket := mock.Anything
	path := mock.Anything
	data := mock.Anything
	contentType := mock.Anything
	ctx := context.Background()
	doWrite := func(mt *mockTransport) *storage.Writer {
		client := mockClient(t, mt)
		wc := client.Bucket(bucket).Object(path).If(storage.Conditions{DoesNotExist: true}).NewWriter(ctx)
		wc.ContentType = contentType

		// We can't check that the Write fails, since it depends on the write to the
		// underling mockTransport failing which is racy.
		wc.Write([]byte(data))
		return wc
	}
	t.Run("happy case", func(t *testing.T) {
		mt := &mockTransport{}
		mt.addResult(&http.Response{StatusCode: 200, Body: bodyReader("{}")}, nil)
		wc := doWrite(mt)
		err := uploadToCloudStorage(wc, data, contentType)
		assert.Nil(t, err)
	})

	t.Run("error case", func(t *testing.T) {
		wc := doWrite(&mockTransport{})
		err := uploadToCloudStorage(wc, data, contentType)
		assert.NotNil(t, err)
	})
}
