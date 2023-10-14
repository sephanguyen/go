package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	internal_auth_user "github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	mock_fatima "github.com/manabie-com/backend/mock/fatima/services"
	mock_multitenant "github.com/manabie-com/backend/mock/golibs/auth/multitenant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_firebase "github.com/manabie-com/backend/mock/golibs/firebase"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_locationRepo "github.com/manabie-com/backend/mock/mastermgmt/modules/location/infrastructure/repo"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestUpdateStudent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	tx := new(mock_database.Tx)
	tx2 := new(mock_database.Tx)
	userRepo := new(mock_repositories.MockUserRepo)
	usrEmailRepo := new(mock_repositories.MockUsrEmailRepo)
	studentRepo := new(mock_repositories.MockStudentRepo)
	userGroupRepo := new(mock_repositories.MockUserGroupRepo)
	userAccessPathRepo := new(mock_repositories.MockUserAccessPathRepo)
	locationRepo := new(mock_locationRepo.MockLocationRepo)
	orgRepo := new(mock_repositories.OrganizationRepo)
	firebaseAuth := new(mock_firebase.AuthClient)
	tenantManager := new(mock_multitenant.TenantManager)
	firebaseAuthClient := new(mock_multitenant.TenantClient)
	jsm := new(mock_nats.JetStreamManagement)
	fatimaClient := new(mock_fatima.SubscriptionModifierServiceClient)
	schoolHistoryRepo := new(mock_repositories.MockSchoolHistoryRepo)
	schoolInfoRepo := new(mock_repositories.MockSchoolInfoRepo)
	schoolCourseRepo := new(mock_repositories.MockSchoolCourseRepo)
	userAddressRepo := new(mock_repositories.MockUserAddressRepo)
	prefectureRepo := new(mock_repositories.MockPrefectureRepo)
	userPhoneNumberRepo := new(mock_repositories.MockUserPhoneNumberRepo)
	domainGradeRepo := new(mock_repositories.MockDomainGradeRepo)
	gradeOrganizationRepo := new(mock_repositories.MockGradeOrganizationRepo)
	domainTaggedUserRepo := new(mock_repositories.MockDomainTaggedUserRepo)
	domainTagRepo := new(mock_repositories.MockDomainTagRepo)
	studentParentRepo := new(mock_repositories.MockStudentParentRepo)

	service := UserModifierService{
		DB:                    db,
		UserRepo:              userRepo,
		UsrEmailRepo:          usrEmailRepo,
		StudentRepo:           studentRepo,
		UserGroupRepo:         userGroupRepo,
		UserAccessPathRepo:    userAccessPathRepo,
		LocationRepo:          locationRepo,
		OrganizationRepo:      orgRepo,
		FirebaseClient:        firebaseAuth,
		TenantManager:         tenantManager,
		FirebaseAuthClient:    firebaseAuthClient,
		JSM:                   jsm,
		FatimaClient:          fatimaClient,
		SchoolHistoryRepo:     schoolHistoryRepo,
		SchoolInfoRepo:        schoolInfoRepo,
		SchoolCourseRepo:      schoolCourseRepo,
		UserAddressRepo:       userAddressRepo,
		PrefectureRepo:        prefectureRepo,
		UserPhoneNumberRepo:   userPhoneNumberRepo,
		DomainGradeRepo:       domainGradeRepo,
		GradeOrganizationRepo: gradeOrganizationRepo,
		DomainTaggedUserRepo:  domainTaggedUserRepo,
		DomainTagRepo:         domainTagRepo,
		StudentParentRepo:     studentParentRepo,
	}

	user := &entity.LegacyUser{
		FullName: database.Text("Albert Einstein JR"),
		Email:    database.Text("student-id-1@example.com"),
	}
	student := &entity.LegacyStudent{
		ID:           database.Text("student-id-1"),
		CurrentGrade: database.Int2(int16(3)),
		SchoolID:     database.Int4(1),
		LegacyUser:   *user,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Studnet 1",
					Grade:             2,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1-edited@example.com",
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
					TagIds:            []string{"tag_id1"},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags([]entity.DomainTag{createMockDomainTag("tag_id1")}), nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				studentRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				domainTaggedUserRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
			expectedErr: nil,
		},
		{
			name: "happy case with isStudentLocationFlagEnabled had error",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Studnet 1",
					Grade:             2,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1-edited@example.com",
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				studentRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
			expectedErr: nil,
		},
		{
			name: "student non-exists",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "-1",
					Name:              "student 2",
					Grade:             2,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1-edited@example.com",
					LocationIds:       []string{"location_1"},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(&entity.LegacyStudent{}, puddle.ErrClosedPool)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},

			expectedErr: fmt.Errorf("rpc error: code = InvalidArgument desc = student id is not exists"),
		},
		{
			name: "user non-exists",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "1",
					Name:              "student name",
					Grade:             2,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1-edited@example.com",
					LocationIds:       []string{"location_1"},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(&entity.LegacyStudent{}, nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{}, puddle.ErrClosedPool)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			expectedErr: fmt.Errorf("rpc error: code = InvalidArgument desc = user id is not exists"),
		},
		{
			name: "student id empty",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id: "",
				},
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("rpc error: code = InvalidArgument desc = student id cannot be empty"),
		},
		{
			name: "student name empty",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:    "1111",
					Email: "email@manabie.com",
					Name:  "",
				},
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("rpc error: code = InvalidArgument desc = student name cannot be empty"),
		},
		{
			name: "student first name empty",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:       "1111",
					Email:    "email@manabie.com",
					Name:     "",
					LastName: "Last",
				},
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("rpc error: code = InvalidArgument desc = student first name cannot be empty"),
		},
		{
			name: "student last name empty",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:        "1111",
					Email:     "email@manabie.com",
					Name:      "",
					FirstName: "First",
				},
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("rpc error: code = InvalidArgument desc = student last name cannot be empty"),
		},
		{
			name: "student email empty",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:   "1111",
					Name: "student name",
				},
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("rpc error: code = InvalidArgument desc = student email cannot be empty"),
		},
		{
			name: "student enroll status unknown",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:               "1111",
					Name:             "student name",
					Email:            "student-id-1-edited@example.com",
					EnrollmentStatus: pb.StudentEnrollmentStatus(-1),
					LocationIds:      []string{"location_1"},
				},
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("rpc error: code = InvalidArgument desc = %v", ErrStudentEnrollmentStatusUnknown),
		},
		{
			name: "student enroll status nil",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:          "1111",
					Name:        "student name",
					Email:       "student-id-1-edited@example.com",
					LocationIds: []string{"location_1"},
				},
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("rpc error: code = InvalidArgument desc = %v", ErrStudentEnrollmentStatusNotAllowedTobeNone),
		},
		{
			name: "error when finding tag by ids",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Studnet 1",
					Grade:             2,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1-edited@example.com",
					LocationIds:       []string{"location_1"},
					TagIds:            []string{idutil.ULIDNow()},
				},
			},
			setup: func(ctx context.Context) {
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
			},
			expectedErr: status.Error(codes.InvalidArgument, ErrTagIDsMustBeExisted.Error()),
		},
		{
			name: "can't get tag",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Studnet 1",
					Grade:             2,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1-edited@example.com",
					LocationIds:       []string{"location_1"},
					TagIds:            []string{idutil.ULIDNow()},
				},
			},
			setup: func(ctx context.Context) {
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, fmt.Errorf("error"))
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("DomainTagRepo.GetByIDs: %v", fmt.Errorf("error")).Error()),
		},
		{
			name: "locationIds has empty element",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "1111",
					Name:              "student name",
					Email:             "student-id-1-edited@example.com",
					Grade:             2,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{""},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
			},
			expectedErr: fmt.Errorf("rpc error: code = InvalidArgument desc = getLocations invalid params: location_id empty"),
		},
		{
			name: "locationIds has invalid resource_path",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "1111",
					Name:              "student name",
					Email:             "student-id-1-edited@example.com",
					Grade:             2,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_invalid"},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID: "location-id_2",
						Name:       "center",
					},
				}, nil).Once()
			},
			expectedErr: fmt.Errorf("rpc error: code = InvalidArgument desc = getLocations fail: resource path invalid, expect %d, but actual ", constants.ManabieSchool),
		},
		{
			name: "update student who already have student package: cannot get student package",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:               "1111",
					Name:             "student name",
					Email:            "student-id-1-edited@example.com",
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					LocationIds:      []string{"location_1"},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(nil, grpc.ErrServerStopped)
			},
			expectedErr: fmt.Errorf("rpc error: code = InvalidArgument desc = %v", status.Error(codes.Internal, fmt.Errorf("validateLocationsForUpdateStudent: %w", grpc.ErrServerStopped).Error())),
		},
		{
			name: "update student who already had student package student package active: fail",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:               "1111",
					Name:             "student name",
					Email:            "student-id-1-edited@example.com",
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					LocationIds:      []string{"location-1"},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{
					StudentPackages: []*fpb.StudentPackage{
						{
							StudentId:   "1111",
							StartAt:     timestamppb.Now(),
							EndAt:       timestamppb.New(time.Now().Add(24 * time.Hour)),
							LocationIds: []string{"location-2"},
						},
					},
				}, nil)
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf(constant.InvalidLocations).Error()),
		},
		{
			name: "update student who already had student package student package inactive: success",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:               "1111",
					Name:             "student name",
					Email:            "student-id-1-edited@example.com",
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					LocationIds:      []string{"location-1"},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{
					StudentPackages: []*fpb.StudentPackage{
						{
							StudentId:   "1111",
							StartAt:     timestamppb.New(time.Now().Add(-30 * 24 * time.Hour)),
							EndAt:       timestamppb.New(time.Now().Add(-24 * time.Hour)),
							LocationIds: []string{"location-2"},
						},
					},
				}, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				studentRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
			expectedErr: nil,
		},
		{
			name: "update student who already had student package student package active with valid location: success",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:               "1111",
					Name:             "student name",
					Email:            "student-id-1-edited@example.com",
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					LocationIds:      []string{"location-1"},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{
					StudentPackages: []*fpb.StudentPackage{
						{
							StudentId:   "1111",
							StartAt:     timestamppb.Now(),
							EndAt:       timestamppb.New(time.Now().Add(24 * time.Hour)),
							LocationIds: []string{"location-1"},
						},
					},
				}, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				studentRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
			expectedErr: nil,
		},
		{
			name: "update student with first name and last name success",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "1111",
					FirstName:         "First",
					LastName:          "Last",
					FirstNamePhonetic: "First Phonetic",
					LastNamePhonetic:  "Last Phonetic",
					Email:             "student-id-1-edited@example.com",
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					LocationIds:       []string{"location-1"},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{
					StudentPackages: []*fpb.StudentPackage{
						{
							StudentId:   "1111",
							StartAt:     timestamppb.Now(),
							EndAt:       timestamppb.New(time.Now().Add(24 * time.Hour)),
							LocationIds: []string{"location-1"},
						},
					},
				}, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				studentRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
			expectedErr: nil,
		},
		{
			name: "update student with first name and last name and last name phonetic",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:               "1111",
					FirstName:        "First",
					LastName:         "Last",
					LastNamePhonetic: "Last name phonetic",
					Email:            "student-id-1-edited@example.com",
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					LocationIds:      []string{"location-1"},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{
					StudentPackages: []*fpb.StudentPackage{
						{
							StudentId:   "1111",
							StartAt:     timestamppb.Now(),
							EndAt:       timestamppb.New(time.Now().Add(24 * time.Hour)),
							LocationIds: []string{"location-1"},
						},
					},
				}, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				studentRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
			expectedErr: nil,
		},
		{
			name: "update student with first name and last name and both first name phonetic and last name phonetic",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "1111",
					FirstName:         "First",
					LastName:          "Last",
					FirstNamePhonetic: "First name phonetic",
					LastNamePhonetic:  "Last name phonetic",
					Email:             "student-id-1-edited@example.com",
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					LocationIds:       []string{"location-1"},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{
					StudentPackages: []*fpb.StudentPackage{
						{
							StudentId:   "1111",
							StartAt:     timestamppb.Now(),
							EndAt:       timestamppb.New(time.Now().Add(24 * time.Hour)),
							LocationIds: []string{"location-1"},
						},
					},
				}, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				studentRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
			expectedErr: nil,
		},
		{
			name: "update student failed because can't get tenant id by organization ID",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "1111",
					FirstName:         "First",
					LastName:          "Last",
					FirstNamePhonetic: "First name phonetic",
					LastNamePhonetic:  "Last name phonetic",
					Email:             "student-id-1-edited@example.com",
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					LocationIds:       []string{"location-1"},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{
					StudentPackages: []*fpb.StudentPackage{
						{
							StudentId:   "1111",
							StartAt:     timestamppb.Now(),
							EndAt:       timestamppb.New(time.Now().Add(24 * time.Hour)),
							LocationIds: []string{"location-1"},
						},
					},
				}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				tx.On("Commit", mock.Anything).Once().Return(nil)
				db.On("Begin", mock.Anything).Once().Return(tx2, nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetByEmail", ctx, tx2, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx2, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", pgx.ErrNoRows)
				tx2.On("Rollback", mock.Anything).Once().Return(nil)
			},
			expectedErr: status.Error(codes.FailedPrecondition, errcode.TenantDoesNotExistErr{OrganizationID: fmt.Sprint(constants.ManabieSchool)}.Error()),
		},
		{
			name: "update student failed because can't get tenant client by tenant ID",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "1111",
					FirstName:         "First",
					LastName:          "Last",
					FirstNamePhonetic: "First name phonetic",
					LastNamePhonetic:  "Last name phonetic",
					Email:             "student-id-1-edited@example.com",
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					LocationIds:       []string{"location-1"},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{
					StudentPackages: []*fpb.StudentPackage{
						{
							StudentId:   "1111",
							StartAt:     timestamppb.Now(),
							EndAt:       timestamppb.New(time.Now().Add(24 * time.Hour)),
							LocationIds: []string{"location-1"},
						},
					},
				}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				tx.On("Commit", mock.Anything).Once().Return(nil)
				db.On("Begin", mock.Anything).Once().Return(tx2, nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetByEmail", ctx, tx2, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx2, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return(aValidTenantID, nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(nil, internal_auth_user.ErrTenantNotFound)
				tx2.On("Rollback", mock.Anything).Once().Return(nil)
			},
			expectedErr: status.Error(codes.Internal, errors.Wrap(internal_auth_user.ErrTenantNotFound, "TenantClient").Error()),
		},
		{
			name: "update student with school history successfully",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Student 1",
					Grade:             2,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1-edited@example.com",
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
				},
				SchoolHistories: []*pb.SchoolHistory{
					{
						SchoolId:       uuid.NewString(),
						SchoolCourseId: uuid.NewString(),
						StartDate:      timestamppb.Now(),
						EndDate:        timestamppb.Now(),
					},
					{
						SchoolId:       uuid.NewString(),
						SchoolCourseId: uuid.NewString(),
						StartDate:      timestamppb.Now(),
						EndDate:        timestamppb.Now(),
					},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				studentRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{
					{
						ID:      database.Text("school_info-id-1"),
						LevelID: database.Text("school_level-id-1"),
					},
					{
						ID:      database.Text("school_info-id-2"),
						LevelID: database.Text("school_level-id-2"),
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
				schoolHistoryRepo.On("GetCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
			expectedErr: nil,
		},
		{
			name: "update student with school history: missing school_id",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Student 1",
					Grade:             2,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1-edited@example.com",
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
				},
				SchoolHistories: []*pb.SchoolHistory{
					{
						SchoolCourseId: uuid.NewString(),
						StartDate:      timestamppb.Now(),
						EndDate:        timestamppb.Now(),
					},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				studentRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil, nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("validateSchoolHistoriesReq: %v", fmt.Errorf("school_id cannot be empty at row: %v", 1)).Error()),
		},
		{
			name: "update student with home address successfully",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Student 1",
					Grade:             2,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1-edited@example.com",
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
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
						Prefecture:  "ID-test-03",
						City:        "city-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				studentRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
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
				prefectureRepo.On("GetByPrefectureID", ctx, mock.Anything, database.Text("ID-test-03")).Return(&entity.Prefecture{
					ID:             database.Text("ID-test-03"),
					PrefectureCode: database.Text("03"),
					Name:           database.Text("name-03"),
				}, nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
			expectedErr: nil,
		},
		{
			name: "update student with home address: missing address type",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Student 1",
					Grade:             2,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1-edited@example.com",
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
				},
				UserAddresses: []*pb.UserAddress{
					{
						AddressId:   uuid.NewString(),
						AddressType: pb.AddressType_BILLING_ADDRESS,
						Prefecture:  "ID-test-01",
					},
				},
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("validateUserAddressesReq: %v", fmt.Errorf("address_type cannot be other type, must be HOME_ADDRESS: %v", 1)).Error()),
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				studentRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				prefectureRepo.On("GetByPrefectureID", ctx, mock.Anything, database.Text("ID-test-01")).Return(&entity.Prefecture{
					ID:             database.Text("ID-test-01"),
					PrefectureCode: database.Text("01"),
					Name:           database.Text("name-01"),
				}, nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "update student with student phone number Upsert error",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Student 1",
					Grade:             2,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1-edited@example.com",
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
					StudentPhoneNumber: &pb.StudentPhoneNumber{
						PhoneNumber:       validPhoneNumber,
						HomePhoneNumber:   validHomePhoneNumber,
						ContactPreference: pb.StudentContactPreference_STUDENT_PHONE_NUMBER,
					},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				studentRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil, nil)
				userPhoneNumberRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(errors.New("error Upsert"))
			},
			expectedErr: status.Error(codes.Unknown, fmt.Errorf("error Upsert").Error()),
		},
		{
			name: "update student with update student phone number successfully",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Student 1",
					Grade:             2,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1-edited@example.com",
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
					StudentPhoneNumbers: &pb.UpdateStudentPhoneNumber{
						StudentPhoneNumber: []*pb.StudentPhoneNumberWithID{
							{
								StudentPhoneNumberId: "1",
								PhoneNumber:          validPhoneNumber,
								PhoneNumberType:      pb.StudentPhoneNumberType_PHONE_NUMBER,
							},
							{
								StudentPhoneNumberId: "2",
								PhoneNumber:          validHomePhoneNumber,
								PhoneNumberType:      pb.StudentPhoneNumberType_HOME_PHONE_NUMBER,
							},
						},
						ContactPreference: pb.StudentContactPreference_STUDENT_PHONE_NUMBER,
					},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				studentRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeID", ctx, db, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
			expectedErr: nil,
		},
		{
			name: "update student with update student phone number error SoftDeleteByUserID",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Student 1",
					Grade:             2,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1-edited@example.com",
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
					StudentPhoneNumbers: &pb.UpdateStudentPhoneNumber{
						StudentPhoneNumber: []*pb.StudentPhoneNumberWithID{
							{
								StudentPhoneNumberId: "1",
								PhoneNumber:          validPhoneNumber,
								PhoneNumberType:      pb.StudentPhoneNumberType_PHONE_NUMBER,
							},
							{
								StudentPhoneNumberId: "2",
								PhoneNumber:          validHomePhoneNumber,
								PhoneNumberType:      pb.StudentPhoneNumberType_HOME_PHONE_NUMBER,
							},
						},
						ContactPreference: pb.StudentContactPreference_STUDENT_PHONE_NUMBER,
					},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				studentRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeID", ctx, db, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil, nil)
				userPhoneNumberRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(errors.New("error SoftDeleteByUserIDs"))
			},
			expectedErr: status.Error(codes.Unknown, fmt.Errorf("error SoftDeleteByUserIDs").Error()),
		},
		{
			name: "update student with update student phone number error Upsert",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Student 1",
					Grade:             2,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1-edited@example.com",
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
					StudentPhoneNumbers: &pb.UpdateStudentPhoneNumber{
						StudentPhoneNumber: []*pb.StudentPhoneNumberWithID{
							{
								StudentPhoneNumberId: "1",
								PhoneNumber:          validPhoneNumber,
								PhoneNumberType:      pb.StudentPhoneNumberType_PHONE_NUMBER,
							},
							{
								StudentPhoneNumberId: "2",
								PhoneNumber:          validHomePhoneNumber,
								PhoneNumberType:      pb.StudentPhoneNumberType_HOME_PHONE_NUMBER,
							},
						},
						ContactPreference: pb.StudentContactPreference_STUDENT_PHONE_NUMBER,
					},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				studentRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeID", ctx, db, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil, nil)
				userPhoneNumberRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(errors.New("error Upsert"))
			},
			expectedErr: status.Error(codes.Unknown, fmt.Errorf("error Upsert").Error()),
		},
		{
			name: "update student with student phone number successfully",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Student 1",
					Grade:             2,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1-edited@example.com",
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
					StudentPhoneNumber: &pb.StudentPhoneNumber{
						PhoneNumber:       validPhoneNumber,
						HomePhoneNumber:   validHomePhoneNumber,
						ContactPreference: pb.StudentContactPreference_STUDENT_PHONE_NUMBER,
					},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				studentRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
			expectedErr: nil,
		},
		{
			name: "update student with student phone number SoftDeleteUserIDs error",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Student 1",
					Grade:             2,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1-edited@example.com",
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
					StudentPhoneNumber: &pb.StudentPhoneNumber{
						PhoneNumber:       validPhoneNumber,
						HomePhoneNumber:   validHomePhoneNumber,
						ContactPreference: pb.StudentContactPreference_STUDENT_PHONE_NUMBER,
					},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				studentRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil, nil)
				userPhoneNumberRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(errors.New("error SoftDeleteByUserIDs"))
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
			},
			expectedErr: status.Error(codes.Unknown, fmt.Errorf("error SoftDeleteByUserIDs").Error()),
		},
		{
			name: "update student with student phone number invalid student phone number",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Student 1",
					Grade:             2,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1-edited@example.com",
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
					StudentPhoneNumber: &pb.StudentPhoneNumber{
						PhoneNumber:       invalidPhoneNumber,
						HomePhoneNumber:   validHomePhoneNumber,
						ContactPreference: pb.StudentContactPreference_STUDENT_PHONE_NUMBER,
					},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				studentRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
			expectedErr: status.Error(codes.InvalidArgument, "validateStudentPhoneNumber: error regexp.MatchString: doesn't match"),
		},
		{
			name: "update student with student phone number invalid student home phone number",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Student 1",
					Grade:             2,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1-edited@example.com",
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
					StudentPhoneNumber: &pb.StudentPhoneNumber{
						PhoneNumber:       validPhoneNumber,
						HomePhoneNumber:   invalidPhoneNumber,
						ContactPreference: pb.StudentContactPreference_STUDENT_PHONE_NUMBER,
					},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				studentRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
			expectedErr: status.Error(codes.InvalidArgument, "validateStudentPhoneNumber: error regexp.MatchString: doesn't match"),
		},
		{
			name: "update student with student phone number same with student home phone number",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Student 1",
					Grade:             2,
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1-edited@example.com",
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
					StudentPhoneNumber: &pb.StudentPhoneNumber{
						PhoneNumber:       validPhoneNumber,
						HomePhoneNumber:   validPhoneNumber,
						ContactPreference: pb.StudentContactPreference_STUDENT_PHONE_NUMBER,
					},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				studentRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				// schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
			expectedErr: status.Error(codes.InvalidArgument, "validateStudentPhoneNumber: phone number and home phone number must not be the same"),
		},
		{
			name: "update student success with grade master",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Studnet 1",
					GradeId:           "grade-id-1",
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1-edited@example.com",
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
				},
			},
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				grades := []entity.DomainGrade{repository.NewGrade(entity.NullDomainGrade{})}
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(grades, nil)
				gradeOrganizationRepo.On("GetByGradeIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				userRepo.On("GetByEmail", ctx, tx, mock.Anything).Once().Return([]*entity.LegacyUser{}, nil)
				usrEmailRepo.On("UpdateEmail", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, db, mock.Anything).Once().Return("", nil)
				tenantClient := &mock_multitenant.TenantClient{}
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetUser", ctx, mock.Anything).Return(nil, nil)
				tenantClient.On("LegacyUpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				schoolInfoRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return([]*entity.SchoolInfo{}, nil)
				schoolCourseRepo.On("GetByIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{}, nil)
				schoolHistoryRepo.On("SoftDeleteByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				schoolHistoryRepo.On("SetCurrentSchoolByStudentIDAndSchoolID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("RemoveCurrentSchoolByStudentID", ctx, db, mock.Anything).Once().Return(nil)
				userAddressRepo.On("SoftDeleteByUserIDs", ctx, tx, mock.Anything).Once().Return(nil)
				userAddressRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				domainTaggedUserRepo.On("GetByUserIDs", ctx, tx, mock.Anything).Once().Return([]entity.DomainTaggedUser{}, nil)
				studentParentRepo.On("UpsertParentAccessPathByStudentIDs", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil, nil)
			},
			expectedErr: nil,
		},
		{
			name: "update student failed with grade master",
			ctx:  ctx,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:                "student-id-1",
					Name:              "Studnet 1",
					GradeId:           "grade-id-1",
					EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					StudentExternalId: "some student external ID",
					StudentNote:       "some student note",
					Email:             "student-id-1-edited@example.com",
					Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
					Gender:            pb.Gender_MALE,
					LocationIds:       []string{"location_1"},
				},
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "grade_id does not exist"),
			setup: func(ctx context.Context) {
				domainTagRepo.On("GetByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainTags{}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID:   "location-id",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
			},
		},

		// {
		// 	name: "update student failed with grade master archived",
		// 	ctx:  ctx,
		// 	req: &pb.UpdateStudentRequest{
		// 		StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
		// 			Id:                "student-id-1",
		// 			Name:              "Studnet 1",
		// 			GradeId:           "grade-id-1",
		// 			EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
		// 			StudentExternalId: "some student external ID",
		// 			StudentNote:       "some student note",
		// 			Email:             "student-id-1-edited@example.com",
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
		// 		tx.On("Commit", mock.Anything).Once().Return(nil)
		// 		studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(student, nil)
		// 		domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(grades, nil)
		// 		userRepo.On("Get", ctx, db, mock.Anything).Once().Return(user, nil)
		// 		locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
		// 			{
		// 				LocationID:   "location-id",
		// 				Name:         "center",
		// 				ResourcePath: fmt.Sprint(constants.ManabieSchool),
		// 			},
		// 		}, nil).Once()
		// 		fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
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

			_, err := service.UpdateStudent(testCase.ctx, testCase.req.(*pb.UpdateStudentRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, studentRepo, userRepo, userGroupRepo, userAccessPathRepo, locationRepo, firebaseAuth, fatimaClient)
		})
	}
}

func TestUserModifierService_ValidUpdateStudentRequest(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	// tx := new(mock_database.Tx)
	fatimaClient := new(mock_fatima.SubscriptionModifierServiceClient)

	s := UserModifierService{
		DB:           db,
		FatimaClient: fatimaClient,
	}

	testCases := []struct {
		name        string
		ctx         context.Context
		req         interface{}
		expectedErr error
		setup       func(ctx context.Context)
	}{
		{
			name:        "err when input create student miss locations ID",
			ctx:         ctx,
			expectedErr: errors.New("student location length must be at least 1"),
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:               "testing id",
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
			setup: func(ctx context.Context) {
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
			},
		},
		{
			name:        "happy case",
			ctx:         ctx,
			expectedErr: nil,
			req: &pb.UpdateStudentRequest{
				StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
					Id:               "testing id",
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
			setup: func(ctx context.Context) {
				fatimaClient.On("ListStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.ListStudentPackageResponse{}, nil)
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

			err := s.validUpdateStudentRequest(testCase.ctx,
				testCase.req.(*pb.UpdateStudentRequest),
			)

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}

}
