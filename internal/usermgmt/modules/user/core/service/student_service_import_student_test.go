package service

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	internal_auth_user "github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	mock_usermgmt "github.com/manabie-com/backend/internal/usermgmt/pkg/mock"
	mock_multitenant "github.com/manabie-com/backend/mock/golibs/auth/multitenant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_firebase "github.com/manabie-com/backend/mock/golibs/firebase"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_clients "github.com/manabie-com/backend/mock/lessonmgmt/zoom/clients"
	mock_locationRepo "github.com/manabie-com/backend/mock/mastermgmt/modules/location/infrastructure/repo"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vmihailenco/taskq/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockTaskQueue struct {
}

func (m *mockTaskQueue) Add(msg *taskq.Message) error {
	return nil
}

func TestImportStudent(t *testing.T) {
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
	importUserEventRepo := new(mock_repositories.MockImportUserEventRepo)
	locationRepo := new(mock_locationRepo.MockLocationRepo)
	jsm := new(mock_nats.JetStreamManagement)
	firebaseAuth := new(mock_firebase.AuthClient)
	tenantManager := new(mock_multitenant.TenantManager)
	firebaseAuthClient := new(mock_multitenant.TenantClient)
	tenantClient := &mock_multitenant.TenantClient{}
	taskQueue := &mockTaskQueue{}
	userAddressRepo := new(mock_repositories.MockUserAddressRepo)
	prefectureRepo := new(mock_repositories.MockPrefectureRepo)
	domainGradeRepo := new(mock_repositories.MockDomainGradeRepo)
	gradeOrganizationRepo := new(mock_repositories.MockGradeOrganizationRepo)
	userPhoneNumberRepo := new(mock_repositories.MockUserPhoneNumberRepo)
	schoolHistoryRepo := new(mock_repositories.MockSchoolHistoryRepo)
	schoolInfoRepo := new(mock_repositories.MockSchoolInfoRepo)
	schoolCourseRepo := new(mock_repositories.MockSchoolCourseRepo)
	domainTagRepo := new(mock_repositories.MockDomainTagRepo)
	domainTaggedUserRepo := new(mock_repositories.MockDomainTaggedUserRepo)
	domainEnrollmentStatusHistoryRepo := new(mock_repositories.MockDomainEnrollmentStatusHistoryRepo)
	domainUserAccessPathRepo := new(mock_repositories.MockDomainUserAccessPathRepo)
	domainLocationRepo := new(mock_repositories.MockDomainLocationRepo)
	mockConfigurationClient := new(mock_clients.MockConfigurationClient)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	userModifierService := UserModifierService{
		DB:                    db,
		OrganizationRepo:      orgRepo,
		UsrEmailRepo:          usrEmailRepo,
		UserRepo:              userRepo,
		StudentRepo:           studentRepo,
		UserGroupRepo:         userGroupRepo,
		UserAccessPathRepo:    userAccessPathRepo,
		LocationRepo:          locationRepo,
		FirebaseClient:        firebaseAuth,
		FirebaseAuthClient:    firebaseAuthClient,
		TenantManager:         tenantManager,
		JSM:                   jsm,
		DomainGradeRepo:       domainGradeRepo,
		GradeOrganizationRepo: gradeOrganizationRepo,
		DomainTaggedUserRepo:  domainTaggedUserRepo,
		UnleashClient:         mockUnleashClient,
	}

	s := StudentService{
		DB:                          db,
		FirebaseAuthClient:          firebaseAuthClient,
		OrganizationRepo:            orgRepo,
		StudentRepo:                 studentRepo,
		UserRepo:                    userRepo,
		UsrEmailRepo:                usrEmailRepo,
		UserGroupRepo:               userGroupRepo,
		UserGroupV2Repo:             userGroupV2Repo,
		UserGroupsMemberRepo:        userGroupsMemberRepo,
		UserAccessPathRepo:          userAccessPathRepo,
		UserModifierService:         &userModifierService,
		JSM:                         jsm,
		ImportUserEventRepo:         importUserEventRepo,
		TaskQueue:                   taskQueue,
		GradeOrganizationRepo:       gradeOrganizationRepo,
		UserAddressRepo:             userAddressRepo,
		PrefectureRepo:              prefectureRepo,
		SchoolHistoryRepo:           schoolHistoryRepo,
		SchoolInfoRepo:              schoolInfoRepo,
		SchoolCourseRepo:            schoolCourseRepo,
		UserPhoneNumberRepo:         userPhoneNumberRepo,
		DomainTagRepo:               domainTagRepo,
		EnrollmentStatusHistoryRepo: domainEnrollmentStatusHistoryRepo,
		DomainUserAccessPathRepo:    domainUserAccessPathRepo,
		DomainLocationRepo:          domainLocationRepo,
		ConfigurationClient:         mockConfigurationClient,
		UnleashClient:               mockUnleashClient,
	}

	getConfigReq := &mpb.GetConfigurationByKeyRequest{Key: constant.KeyEnrollmentStatusHistoryConfig}
	organizationID := "id"
	configurationsDataLMS := &mpb.Configuration{
		Id:          organizationID,
		ConfigValue: "on",
	}
	configurationsDataERP := &mpb.Configuration{
		Id:          organizationID,
		ConfigValue: "off",
	}

	usrEmail := []*entity.UsrEmail{
		{
			UsrID: database.Text("example-id"),
		},
	}
	hashConfig := mockScryptHash()

	payload1001Rows := "name,email,enrollment_status,grade,phone_number,birthday,gender,location"
	for i := 0; i < 1001; i++ {
		payload1001Rows += "\nStudent 01,student-01@example.com,1,1,0981143301,1999/01/12,1,location-01;location-02"
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
			name:        "invalidMaxSizeFile",
			ctx:         ctx,
			expectedErr: status.Error(codes.InvalidArgument, "invalidMaxSizeFile"),
			req: &pb.ImportStudentRequest{
				Payload: payload10MB,
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "emptyFile",
			ctx:         ctx,
			expectedErr: status.Error(codes.InvalidArgument, "emptyFile"),
			req: &pb.ImportStudentRequest{
				Payload: []byte(``),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalidNumberRow",
			ctx:         ctx,
			expectedErr: status.Error(codes.InvalidArgument, "invalidNumberRow"),
			req: &pb.ImportStudentRequest{
				Payload: []byte(payload1001Rows),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "happy case no row",
			ctx:  ctx,
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location`),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "error: with student phone number",
			ctx:  ctx,
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,email,enrollment_status,grade,student_phone_number,home_phone_number,contact_preference,location
				Student 01 Last Name,Student 01 First Name,student-01@example.com,3,partner_id_1,0993133ff231,0312731737,1,1
				Student 01 Last Name,Student 01 First Name,student-01@example.com,3,partner_id_1,0993133231,031273ff1737,1,1
				Student 01 Last Name,Student 01 First Name,student-01@example.com,3,partner_id_1,0993133231,0993133231,1,1
				Student 01 Last Name,Student 01 First Name,student-01@example.com,3,partner_id_1,0993133231,0983133231,5,1`),
			},
			expectedResp: &pb.ImportStudentResponse{
				Errors: []*pb.ImportStudentResponse_ImportStudentError{
					{
						RowNumber: 2,
						Error:     "notFollowTemplate",
						FieldName: string(studentPhoneNumberCSVHeader),
					},
					{
						RowNumber: 3,
						Error:     "notFollowTemplate",
						FieldName: string(studentHomePhoneNumberCSVHeader),
					},
					{
						RowNumber: 4,
						Error:     "duplicationRow",
						FieldName: string(studentPhoneNumberCSVHeader),
					},
					{
						RowNumber: 5,
						Error:     "notFollowTemplate",
						FieldName: string(studentContactPreferenceCSVHeader),
					},
				},
			},
			setup: func(ctx context.Context) {
				grades := []entity.DomainGrade{&mock_usermgmt.Grade{}}
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)

				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Times(4).Return(grades, nil)
				gradeOrganizationRepo.On("GetByGradeIDs", ctx, db, mock.Anything).Times(4).Return(nil, nil)
				db.On("Begin", mock.Anything).Times(4).Return(tx, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Times(4)
				domainLocationRepo.On("GetByPartnerInternalIDs", ctx, db, []string{"1"}).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Times(4).Return(nil)
			},
		},
		{
			name: "happy case: 1 row with student phone number",
			ctx:  ctx,
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,email,enrollment_status,grade,student_phone_number,home_phone_number,contact_preference,location
				Student 01 Last Name,Student 01 First Name,student-01@example.com,1,partner_id_1,0993133231,0312731737,1,1`),
			},
			setup: func(ctx context.Context) {
				grades := []entity.DomainGrade{&mock_usermgmt.Grade{}}
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(grades, nil)
				gradeOrganizationRepo.On("GetByGradeIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, db, mock.Anything).Once().Return(nil, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{}, nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("CreateMultiple", ctx, db, mock.Anything).Once().Return(usrEmail, nil)
				studentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				importUserEventRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return([]*entity.ImportUserEvent{}, nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				domainLocationRepo.On("GetByPartnerInternalIDs", ctx, db, []string{"1"}).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				mockConfigurationClient.On("GetConfigurationByKey", ctx, getConfigReq).Return(&mpb.GetConfigurationByKeyResponse{Configuration: configurationsDataLMS}, nil).Once()
			},
		},
		{
			name: "error: duplicated values in tag",
			ctx:  ctx,
			expectedResp: &pb.ImportStudentResponse{
				Errors: []*pb.ImportStudentResponse_ImportStudentError{
					{
						RowNumber: 2,
						Error:     "notFollowTemplate",
						FieldName: string(tagStudentCSVHeader),
					},
				},
			},
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location,student_tag
				Student 01 Last Name,Student 01 Last Name,,,student-01@example.com,3,partner_id_1,,,,1,tag_id1;tag_id1`),
			},
			setup: func(ctx context.Context) {
				grades := []entity.DomainGrade{&mock_usermgmt.Grade{}}
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(grades, nil)
				gradeOrganizationRepo.On("GetByGradeIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainTagRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags([]entity.DomainTag{createMockDomainTag("tag_id1")}), nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				domainLocationRepo.On("GetByPartnerInternalIDs", ctx, db, []string{"1"}).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "happy case: 1 row with student tag",
			ctx:  ctx,
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,email,enrollment_status,grade,student_phone_number,home_phone_number,contact_preference,student_tag,location
				Student 01 Last Name,Student 01 First Name,student-01@example.com,1,partner_id_1,0993133231,0312731737,1,partner-id-tag_id1,1`),
			},
			setup: func(ctx context.Context) {
				grades := []entity.DomainGrade{&mock_usermgmt.Grade{}}
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(grades, nil)
				gradeOrganizationRepo.On("GetByGradeIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainTagRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(entity.DomainTags(
					[]entity.DomainTag{
						createMockDomainTagWithType("tag_id1", pb.UserTagType_USER_TAG_TYPE_STUDENT),
					},
				), nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, db, mock.Anything).Once().Return(nil, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{}, nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("CreateMultiple", ctx, db, mock.Anything).Once().Return(usrEmail, nil)
				studentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				importUserEventRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return([]*entity.ImportUserEvent{}, nil)
				userPhoneNumberRepo.On("Upsert", ctx, tx, mock.Anything).Return(nil)
				domainTaggedUserRepo.On("UpsertBatch", ctx, tx, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				domainLocationRepo.On("GetByPartnerInternalIDs", ctx, db, []string{"1"}).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				mockConfigurationClient.On("GetConfigurationByKey", ctx, getConfigReq).Return(&mpb.GetConfigurationByKeyResponse{Configuration: configurationsDataLMS}, nil).Once()
			},
		},
		{
			name: "happy case: 1 row with grade master",
			ctx:  ctx,
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location
				Student 01 Last Name,Student 01 Last Name,,,student-01@example.com,1,partner_id_1,,,,1`),
			},
			setup: func(ctx context.Context) {
				grades := []entity.DomainGrade{&mock_usermgmt.Grade{}}
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(grades, nil)
				gradeOrganizationRepo.On("GetByGradeIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, db, mock.Anything).Once().Return(nil, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{}, nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("CreateMultiple", ctx, db, mock.Anything).Once().Return(usrEmail, nil)
				studentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				importUserEventRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return([]*entity.ImportUserEvent{}, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				domainLocationRepo.On("GetByPartnerInternalIDs", ctx, db, []string{"1"}).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				mockConfigurationClient.On("GetConfigurationByKey", ctx, getConfigReq).Return(&mpb.GetConfigurationByKeyResponse{Configuration: configurationsDataLMS}, nil).Once()
			},
		},
		{
			name: "error AssignWithUserGroup",
			ctx:  ctx,
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location
				Student 01 Last Name,Student 01 Last Name,,,student-01@example.com,1,partner_id_1,,,,1`),
			},
			expectedErr: status.Error(codes.Internal, "otherErrorImport database.ExecInTx: error when assigning student user group to users: error assign with user group"),
			setup: func(ctx context.Context) {
				grades := []entity.DomainGrade{&mock_usermgmt.Grade{}}
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(grades, nil)
				gradeOrganizationRepo.On("GetByGradeIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, db, mock.Anything).Once().Return(nil, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{}, nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(errors.New("error assign with user group"))
				usrEmailRepo.On("CreateMultiple", ctx, db, mock.Anything).Once().Return(usrEmail, nil)
				studentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				importUserEventRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return([]*entity.ImportUserEvent{}, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				domainLocationRepo.On("GetByPartnerInternalIDs", ctx, db, []string{"1"}).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				mockConfigurationClient.On("GetConfigurationByKey", ctx, getConfigReq).Return(&mpb.GetConfigurationByKeyResponse{Configuration: configurationsDataLMS}, nil).Once()
			},
		},
		{
			name: "happy case with phonetic name 1 row",
			ctx:  ctx,
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location
Student 01 last name,Student 01 first name,Student 01 last name phonetic,Student 01 first name phonetic,student-01@example.com,1,0,,,,1`),
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				gradeOrganizationRepo.On("GetByGradeValues", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, db, mock.Anything).Once().Return(nil, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{}, nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("CreateMultiple", ctx, db, mock.Anything).Once().Return(usrEmail, nil)
				studentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				importUserEventRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return([]*entity.ImportUserEvent{}, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				domainLocationRepo.On("GetByPartnerInternalIDs", ctx, db, []string{"1"}).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				mockConfigurationClient.On("GetConfigurationByKey", ctx, getConfigReq).Return(&mpb.GetConfigurationByKeyResponse{Configuration: configurationsDataLMS}, nil).Once()
			},
		},
		{
			name: "error user repo create multiple",
			ctx:  ctx,
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location
Student 01 last name,Student 01 first name,Student 01 last name phonetic,Student 01 first name phonetic,student-01@example.com,1,0,,,,1`),
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				gradeOrganizationRepo.On("GetByGradeValues", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, db, mock.Anything).Once().Return(nil, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{}, nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("CreateMultiple", ctx, db, mock.Anything).Once().Return(usrEmail, nil)
				studentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				importUserEventRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return([]*entity.ImportUserEvent{}, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				domainLocationRepo.On("GetByPartnerInternalIDs", ctx, db, []string{"1"}).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				mockConfigurationClient.On("GetConfigurationByKey", ctx, getConfigReq).Return(&mpb.GetConfigurationByKeyResponse{Configuration: configurationsDataLMS}, nil).Once()
			},
		},
		{
			name: "missingMandatory with first name, last name and phonetic name",
			ctx:  ctx,
			expectedResp: &pb.ImportStudentResponse{
				Errors: []*pb.ImportStudentResponse_ImportStudentError{
					{
						RowNumber: 2,
						Error:     "missingMandatory",
						FieldName: string(lastNameStudentCSVHeader),
					},
					{
						RowNumber: 3,
						Error:     "missingMandatory",
						FieldName: string(firstNameStudentCSVHeader),
					},
					{
						RowNumber: 4,
						Error:     "missingMandatory",
						FieldName: string(emailStudentCSVHeader),
					},
					{
						RowNumber: 5,
						Error:     "missingMandatory",
						FieldName: string(gradeStudentCSVHeader),
					},
				},
			},
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name_phonetic,last_name,first_name,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location
										,,student first name 01,,student-01@example.com,1,1,,,,
										,Student last name 02,,,student-02@example.com,1,1,,,,
										,Student last name 03,student first name 03,,,1,1,,,,
										,Student last name 05,student last name 05,,student-02@example.com,1,,,,,`),
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
			},
		},
		{
			name: "notFollowTemplate with firstname lastname",
			ctx:  ctx,
			expectedResp: &pb.ImportStudentResponse{
				Errors: []*pb.ImportStudentResponse_ImportStudentError{
					{
						RowNumber: 2,
						Error:     "notFollowTemplate",
						FieldName: string(locationStudentCSVHeader),
					},
					{
						RowNumber: 3,
						Error:     "notFollowTemplate",
						FieldName: string(emailStudentCSVHeader),
					},
					{
						RowNumber: 4,
						Error:     "notFollowTemplate",
						FieldName: string(enrollmentStatusStudentCSVHeader),
					},
					{
						RowNumber: 5,
						Error:     "notFollowTemplate",
						FieldName: string(enrollmentStatusStudentCSVHeader),
					},
					{
						RowNumber: 6,
						Error:     "notFollowTemplate",
						FieldName: string(gradeStudentCSVHeader),
					},
					{
						RowNumber: 7,
						Error:     "notFollowTemplate",
						FieldName: string(gradeStudentCSVHeader),
					},
					{
						RowNumber: 8,
						Error:     "notFollowTemplate",
						FieldName: string(genderStudentCSVHeader),
					},
					{
						RowNumber: 9,
						Error:     "notFollowTemplate",
						FieldName: string(genderStudentCSVHeader),
					},
					{
						RowNumber: 10,
						Error:     "notFollowTemplate",
						FieldName: string(birthdayStudentCSVHeader),
					},
					{
						RowNumber: 11,
						Error:     "notFollowTemplate",
						FieldName: string(locationStudentCSVHeader),
					},
					{
						RowNumber: 12,
						Error:     "notFollowTemplate",
						FieldName: string(locationStudentCSVHeader),
					},
					{
						RowNumber: 13,
						Error:     "notFollowTemplate",
						FieldName: string(enrollmentStatusStudentCSVHeader),
					},
					{
						RowNumber: 14,
						Error:     "notFollowTemplate",
						FieldName: string(gradeStudentCSVHeader),
					},
					{
						RowNumber: 15,
						Error:     "notFollowTemplate",
						FieldName: string(genderStudentCSVHeader),
					},
					{
						RowNumber: 16,
						Error:     "notFollowTemplate",
						FieldName: string(birthdayStudentCSVHeader),
					},
					{
						RowNumber: 17,
						Error:     "notFollowTemplate",
						FieldName: string(phoneNumberStudentCSVHeader),
					},
					{
						RowNumber: 18,
						Error:     "notFollowTemplate",
						FieldName: string(emailStudentCSVHeader),
					},
				},
			},
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location
				Student 00 last name,Student 00 first Name,,,student-00@example.com,1,1,0981143300,1999/01/12,1,location-01
				Student 01 last name,Student 01 first name,,,student-01@example..com,1,0,0981143301,1999/01/12,1,location-01;location-02
				Student 02 last name,Student 02 first name,,,Student-02@example.com,0,16,0981143302,1999/01/12,2,location-01;location-02
				Student 03 last name,Student 03 first name,,,student-03@example.com,8,0,0981143303,1999/01/12,1,location-01;location-02
				Student 04 last name,Student 04 first name,,,student-04@example.com,5,-1,0981143304,1999/01/12,2,location-01;location-02
				Student 05 last name,Student 05 first name,,,student-05@example.com,1,17,0981143305,1999/01/12,1,location-01;location-02
				Student 06 last name,Student 06 first name,,,student-06@example.com,5,16,0981143306,1999/01/12,0,location-01;location-02
				Student 07 last name,Student 07 first name,,,student-07@example.com,1,0,0981143307,1999/01/12,3,location-01;location-02
				Student 08 last name,Student 08 first name,,,student-08@example.com,5,16,0981143308,1999-01-12,2,location-01;location-02
				Student 09 last name,Student 09 first name,,,student-09@example.com,5,16,0981143309,1999/01/12,2,invalid-location
				Student 10 last name,Student 10 first name,,,student-10@example.com,5,16,0981143310,1999/01/12,2,archived-location
				Student 11 last name,Student 11 first name,,,student-11@example.com,invalid-status,16,0981143311,1999/01/12,2,location-01;location-02
				Student 12 last name,Student 12 first name,,,student-12@example.com,5,invalid-grade,0981143312,1999/01/12,2,location-01;location-02
				Student 13 last name,Student 13 first name,,,student-13@example.com,5,16,0981143313,1999/01/12,invalid-gender,location-01;location-02
				Student 14 last name,Student 14 first name,,,student-14@example.com,5,16,0981143314,invalid-birthday,1,location-01;location-02
				Student 15 last name,Student 15 first name,,,student-15@example.com,5,16,invalid-phone,1999/01/12,1,location-01;location-02
				Student 16 last name,Student 16 first name,,,invalid-email,5,16,0981143316,1999/01/12,1,location-01;location-02`),
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				db.On("Begin", mock.Anything).Times(3).Return(tx, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Times(13).Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Times(13).Return(nil, nil)
				gradeOrganizationRepo.On("GetByGradeValues", ctx, db, mock.Anything).Times(13).Return(nil, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Twice().Return(nil, pgx.ErrNoRows)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Once().Return([]*domain.Location{
					{
						IsArchived:   true,
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil)
				tx.On("Rollback", mock.Anything).Twice().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "duplicationRow with email",
			ctx:  ctx,
			expectedResp: &pb.ImportStudentResponse{
				Errors: []*pb.ImportStudentResponse_ImportStudentError{
					{
						RowNumber: 4,
						Error:     "duplicationRow",
						FieldName: string(emailStudentCSVHeader),
					},
				},
			},
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location
				Student 01 last name,Student 01 first name,,,student-01@example.com,1,1,0981143301,1999/01/12,1,1
				Student 02 last name,Student 02 first name,,,student-02@example.com,1,1,0981143302,1999/01/12,1,1
				Student 03 last name,Student 03 first name,,,student-02@example.com,1,1,0981143303,1999/01/12,1,1`),
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Times(3)
				db.On("Begin", mock.Anything).Times(3).Return(tx, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Times(3).Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Times(3).Return(nil, nil)
				domainLocationRepo.On("GetByPartnerInternalIDs", ctx, db, []string{"1"}).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				gradeOrganizationRepo.On("GetByGradeValues", ctx, db, mock.Anything).Times(3).Return(nil, nil)
				tx.On("Commit", mock.Anything).Times(3).Return(nil)
				mockConfigurationClient.On("GetConfigurationByKey", ctx, getConfigReq).Return(&mpb.GetConfigurationByKeyResponse{Configuration: configurationsDataLMS}, nil).Times(3)
			},
		},
		{
			name: "duplicationRow with email (case-insensitive)",
			ctx:  ctx,
			expectedResp: &pb.ImportStudentResponse{
				Errors: []*pb.ImportStudentResponse_ImportStudentError{
					{
						RowNumber: 4,
						Error:     "duplicationRow",
						FieldName: string(emailStudentCSVHeader),
					},
				},
			},
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location
				Student 01 last name,Student 01 first name,,,student-01@example.com,1,1,0981143301,1999/01/12,1,1
				Student 02 last name,Student 02 first name,,,student-02@example.com,1,1,0981143302,1999/01/12,1,1
				Student 03 last name,Student 03 first name,,,STUDENT-02@example.com,1,1,0981143303,1999/01/12,1,1`),
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Times(3)
				db.On("Begin", mock.Anything).Times(3).Return(tx, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Times(3).Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Times(3).Return(nil, nil)
				domainLocationRepo.On("GetByPartnerInternalIDs", ctx, db, []string{"1"}).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				gradeOrganizationRepo.On("GetByGradeValues", ctx, db, mock.Anything).Times(3).Return(nil, nil)
				tx.On("Commit", mock.Anything).Times(3).Return(nil)
				mockConfigurationClient.On("GetConfigurationByKey", ctx, getConfigReq).Return(&mpb.GetConfigurationByKeyResponse{Configuration: configurationsDataLMS}, nil).Times(3)
			},
		},
		{
			name: "duplicationRow with first name last name: phone_number",
			ctx:  ctx,
			expectedResp: &pb.ImportStudentResponse{
				Errors: []*pb.ImportStudentResponse_ImportStudentError{
					{
						RowNumber: 4,
						Error:     "duplicationRow",
						FieldName: string(phoneNumberStudentCSVHeader),
					},
				},
			},
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location
				Student 01 last name,Student 01 first name,,,student-01@example.com,1,1,0981143301,1999/01/12,1,1
				Student 02 last name,Student 02 first name,,,student-02@example.com,1,1,0981143302,1999/01/12,1,1
				Student 03 last name,Student 03 first name,,,student-03@example.com,1,1,0981143302,1999/01/12,1,1`),
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Times(3)
				db.On("Begin", mock.Anything).Times(3).Return(tx, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Times(3).Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Times(3).Return(nil, nil)
				domainLocationRepo.On("GetByPartnerInternalIDs", ctx, db, []string{"1"}).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				gradeOrganizationRepo.On("GetByGradeValues", ctx, db, mock.Anything).Times(3).Return(nil, nil)
				tx.On("Commit", mock.Anything).Times(3).Return(nil)
				mockConfigurationClient.On("GetConfigurationByKey", ctx, getConfigReq).Return(&mpb.GetConfigurationByKeyResponse{Configuration: configurationsDataLMS}, nil).Times(3)
			},
		},
		{
			name: "alreadyRegisteredRow",
			ctx:  ctx,
			expectedResp: &pb.ImportStudentResponse{
				Errors: []*pb.ImportStudentResponse_ImportStudentError{
					{
						RowNumber: 3,
						Error:     "alreadyRegisteredRow",
						FieldName: string(emailStudentCSVHeader),
					},
					{
						RowNumber: 4,
						Error:     "alreadyRegisteredRow",
						FieldName: string(phoneNumberStudentCSVHeader),
					},
				},
			},
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location
				Student 01 last name,Student 01 first name,,,student-01@example.com,1,1,0981143301,1999/01/12,1,1
				Student 02 last name,Student 02 first name,,,student-02@example.com,1,1,0981143302,1999/01/12,1,1
				Student 03 last name,Student 03 first name,,,student-03@example.com,1,1,0981143303,1999/01/12,1,1`),
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, db, mock.Anything).Once().Return([]*entity.LegacyUser{
					{
						ID:          database.Text("student-02"),
						Email:       database.Text("student-02@example.com"),
						PhoneNumber: database.Text("0981143302"),
					},
				}, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Times(3).Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Times(3).Return(nil, nil)
				gradeOrganizationRepo.On("GetByGradeValues", ctx, db, mock.Anything).Times(3).Return(nil, nil)
				userRepo.On("GetByPhone", ctx, db, mock.Anything).Once().Return([]*entity.LegacyUser{
					{
						ID:          database.Text("student-03"),
						Email:       database.Text("student-03@example.com"),
						PhoneNumber: database.Text("0981143303"),
					},
				}, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Times(3)
				db.On("Begin", mock.Anything).Times(3).Return(tx, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Times(3).Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Times(3).Return(nil, nil)
				domainLocationRepo.On("GetByPartnerInternalIDs", ctx, db, []string{"1"}).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				gradeOrganizationRepo.On("GetByGradeValues", ctx, db, mock.Anything).Times(3).Return(nil, nil)
				tx.On("Commit", mock.Anything).Times(3).Return(nil)
				mockConfigurationClient.On("GetConfigurationByKey", ctx, getConfigReq).Return(&mpb.GetConfigurationByKeyResponse{Configuration: configurationsDataLMS}, nil).Times(3)
			},
		},
		{
			name: "happy case with home address: 1 row",
			ctx:  ctx,
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location,postal_code,prefecture,city,first_street,second_street
				Student 01 last name,Student 01 first name,Student 01 last name phonetic,Student 01 first name phonetic,student-01@example.com,1,0,,,,1,9000,02,,,`),
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, db, mock.Anything).Once().Return(nil, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				gradeOrganizationRepo.On("GetByGradeValues", ctx, db, mock.Anything).Once().Return(nil, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{}, nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("CreateMultiple", ctx, db, mock.Anything).Once().Return(usrEmail, nil)
				studentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				importUserEventRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return([]*entity.ImportUserEvent{}, nil)
				prefectureRepo.On("GetByPrefectureCode", ctx, mock.Anything, database.Text("02")).Return(&entity.Prefecture{ID: database.Text("ID-01")}, nil)
				userAddressRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				domainLocationRepo.On("GetByPartnerInternalIDs", ctx, db, []string{"1"}).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				mockConfigurationClient.On("GetConfigurationByKey", ctx, getConfigReq).Return(&mpb.GetConfigurationByKeyResponse{Configuration: configurationsDataLMS}, nil).Once()
			},
		},
		{
			name: "happy case with home address: 1 row with error isStudentLocationFlagEnabled",
			ctx:  ctx,
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location,postal_code,prefecture,city,first_street,second_street
				Student 01 last name,Student 01 first name,Student 01 last name phonetic,Student 01 first name phonetic,student-01@example.com,1,0,,,,1,,,,,`),
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, db, mock.Anything).Once().Return(nil, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				gradeOrganizationRepo.On("GetByGradeValues", ctx, db, mock.Anything).Once().Return(nil, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{}, nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("CreateMultiple", ctx, db, mock.Anything).Once().Return(usrEmail, nil)
				studentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				importUserEventRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return([]*entity.ImportUserEvent{}, nil)
				prefectureRepo.On("GetByPrefectureCode", ctx, mock.Anything, database.Text("")).Return(&entity.Prefecture{}, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				domainLocationRepo.On("GetByPartnerInternalIDs", ctx, db, []string{"1"}).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				mockConfigurationClient.On("GetConfigurationByKey", ctx, getConfigReq).Return(&mpb.GetConfigurationByKeyResponse{Configuration: configurationsDataLMS}, nil).Once()
			},
		},
		{
			name: "happy case with school history: 1 row",
			ctx:  ctx,
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location,postal_code,prefecture,city,first_street,second_street,school,school_course,start_date,end_date
				Student 01 last name,Student 01 first name,Student 01 last name phonetic,Student 01 first name phonetic,student-01@example.com,1,0,,,,1,,,,,,school-partner-id-1;school-partner-id-2,school_course-id-1;school_course-id-2,2022/01/02;2022/11/02,2023/01/02;2023/01/02`),
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, db, mock.Anything).Once().Return(nil, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				gradeOrganizationRepo.On("GetByGradeValues", ctx, db, mock.Anything).Once().Return(nil, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{}, nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("CreateMultiple", ctx, db, mock.Anything).Once().Return(usrEmail, nil)
				studentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				schoolInfoRepo.On("GetBySchoolPartnerIDs", ctx, db, database.TextArray([]string{"school-partner-id-1", "school-partner-id-2"})).Once().Return([]*entity.SchoolInfo{
					{
						ID:      database.Text("school_info-id-1"),
						LevelID: database.Text("school_level-id-1"),
					},
					{
						ID:      database.Text("school_info-id-2"),
						LevelID: database.Text("school_level-id-2"),
					},
				}, nil)
				schoolCourseRepo.On("GetBySchoolCoursePartnerIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{
					{
						ID: database.Text("school_course-id-1"),
					},
					{
						ID: database.Text("school_course-id-2"),
					},
				}, nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				importUserEventRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return([]*entity.ImportUserEvent{}, nil)
				prefectureRepo.On("GetByPrefectureCode", ctx, mock.Anything, database.Text("")).Return(&entity.Prefecture{}, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				domainLocationRepo.On("GetByPartnerInternalIDs", ctx, db, []string{"1"}).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				mockConfigurationClient.On("GetConfigurationByKey", ctx, getConfigReq).Return(&mpb.GetConfigurationByKeyResponse{Configuration: configurationsDataLMS}, nil).Once()
			},
		},
		{
			name: "happy case with school history: 1 row without end date, start date",
			ctx:  ctx,
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location,postal_code,prefecture,city,first_street,second_street,school,school_course,start_date,end_date
				Student 01 last name,Student 01 first name,Student 01 last name phonetic,Student 01 first name phonetic,student-01@example.com,1,0,,,,1,,,,,,school-partner-id-1;school-partner-id-2,school_course-id-1;school_course-id-2,,`),
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, db, mock.Anything).Once().Return(nil, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				gradeOrganizationRepo.On("GetByGradeValues", ctx, db, mock.Anything).Once().Return(nil, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{}, nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("CreateMultiple", ctx, db, mock.Anything).Once().Return(usrEmail, nil)
				studentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				schoolInfoRepo.On("GetBySchoolPartnerIDs", ctx, db, database.TextArray([]string{"school-partner-id-1", "school-partner-id-2"})).Once().Return([]*entity.SchoolInfo{
					{
						ID:      database.Text("school_info-id-1"),
						LevelID: database.Text("school_level-id-1"),
					},
					{
						ID:      database.Text("school_info-id-2"),
						LevelID: database.Text("school_level-id-2"),
					},
				}, nil)
				schoolCourseRepo.On("GetBySchoolCoursePartnerIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{
					{
						ID: database.Text("school_course-id-1"),
					},
					{
						ID: database.Text("school_course-id-2"),
					},
				}, nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				importUserEventRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return([]*entity.ImportUserEvent{}, nil)
				prefectureRepo.On("GetByPrefectureCode", ctx, mock.Anything, database.Text("")).Return(&entity.Prefecture{}, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				domainLocationRepo.On("GetByPartnerInternalIDs", ctx, db, []string{"1"}).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				mockConfigurationClient.On("GetConfigurationByKey", ctx, getConfigReq).Return(&mpb.GetConfigurationByKeyResponse{Configuration: configurationsDataLMS}, nil).Once()
			},
		},
		{
			name: "worst case with school history: duplicate school_level",
			ctx:  ctx,
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location,postal_code,prefecture,city,first_street,second_street,school,school_course,start_date,end_date
				Student 01 last name,Student 01 first name,Student 01 last name phonetic,Student 01 first name phonetic,student-01@example.com,1,0,,,,1,,,,,,school-partner-id-1;school-partner-id-2,school_course-id-1;school_course-id-2,2022/01/02;2022/11/02,2023/01/02;2023/01/02`),
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("otherErrorImport %v", fmt.Errorf("s.generatedStudentCSVs: s.convertLineCSVToStudentCSV: rpc error: code = Internal desc = otherErrorImport s.convertLineCSVToStudentCSV: duplicate school_level_id in school_info school_info-id-2")).Error()),
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				gradeOrganizationRepo.On("GetByGradeValues", ctx, db, mock.Anything).Once().Return(nil, nil)
				schoolInfoRepo.On("GetBySchoolPartnerIDs", ctx, db, database.TextArray([]string{"school-partner-id-1", "school-partner-id-2"})).Once().Return([]*entity.SchoolInfo{
					{
						ID: database.Text("school_info-id-1"),
					},
					{
						ID: database.Text("school_info-id-2"),
					},
				}, nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				prefectureRepo.On("GetByPrefectureCode", ctx, mock.Anything, database.Text("")).Return(&entity.Prefecture{}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				domainLocationRepo.On("GetByPartnerInternalIDs", ctx, db, []string{"1"}).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				mockConfigurationClient.On("GetConfigurationByKey", ctx, getConfigReq).Return(&mpb.GetConfigurationByKeyResponse{Configuration: configurationsDataLMS}, nil).Once()
			},
		},
		{
			name: "worst case with school history: invalid end date",
			ctx:  ctx,
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location,postal_code,prefecture,city,first_street,second_street,school,school_course,start_date,end_date
				Student 01 last name,Student 01 first name,Student 01 last name phonetic,Student 01 first name phonetic,student-01@example.com,1,0,,,,1,,,,,,school-partner-id-1;school-partner-id-2,school_course-id-1;school_course-id-2,2022/01/02;2022/11/02,2023/01/02;2016/01/02`),
			},
			expectedResp: &pb.ImportStudentResponse{
				Errors: []*pb.ImportStudentResponse_ImportStudentError{
					{
						RowNumber: 2,
						Error:     "notFollowTemplate",
						FieldName: string(endDateStudentCSVHeader),
					},
				},
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				gradeOrganizationRepo.On("GetByGradeValues", ctx, db, mock.Anything).Once().Return(nil, nil)
				schoolInfoRepo.On("GetBySchoolPartnerIDs", ctx, db, database.TextArray([]string{"school-partner-id-1", "school-partner-id-2"})).Once().Return([]*entity.SchoolInfo{
					{
						ID:      database.Text("school_info-id-1"),
						LevelID: database.Text("school_level-id-1"),
					},
					{
						ID:      database.Text("school_info-id-2"),
						LevelID: database.Text("school_level-id-2"),
					},
				}, nil)
				schoolCourseRepo.On("GetBySchoolCoursePartnerIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{
					{
						ID: database.Text("school_course-id-1"),
					},
					{
						ID: database.Text("school_course-id-2"),
					},
				}, nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				prefectureRepo.On("GetByPrefectureCode", ctx, mock.Anything, database.Text("")).Return(&entity.Prefecture{}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				domainLocationRepo.On("GetByPartnerInternalIDs", ctx, db, []string{"1"}).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "worst case with school history: school_course does not match with req.SchoolHistories",
			ctx:  ctx,
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location,postal_code,prefecture,city,first_street,second_street,school,school_course,start_date,end_date
				Student 01 last name,Student 01 first name,Student 01 last name phonetic,Student 01 first name phonetic,student-01@example.com,1,0,,,,1,,,,,,school-partner-id-1;school-partner-id-2,school_course-id-1;school_course-id-2,2022/01/02;2022/11/02,2023/01/02;2024/01/02`),
			},
			expectedResp: &pb.ImportStudentResponse{
				Errors: []*pb.ImportStudentResponse_ImportStudentError{
					{
						RowNumber: 2,
						Error:     "notFollowTemplate",
						FieldName: string(schoolCourseStudentCSVHeader),
					},
				},
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				gradeOrganizationRepo.On("GetByGradeValues", ctx, db, mock.Anything).Once().Return(nil, nil)
				schoolInfoRepo.On("GetBySchoolPartnerIDs", ctx, db, database.TextArray([]string{"school-partner-id-1", "school-partner-id-2"})).Once().Return([]*entity.SchoolInfo{
					{
						ID:      database.Text("school_info-id-1"),
						LevelID: database.Text("school_level-id-1"),
					},
					{
						ID:      database.Text("school_info-id-2"),
						LevelID: database.Text("school_level-id-2"),
					},
				}, nil)
				schoolCourseRepo.On("GetBySchoolCoursePartnerIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{
					{
						ID: database.Text("school_course-id-1"),
					},
				}, nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				prefectureRepo.On("GetByPrefectureCode", ctx, mock.Anything, database.Text("")).Return(&entity.Prefecture{}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				domainLocationRepo.On("GetByPartnerInternalIDs", ctx, db, []string{"1"}).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "worst case with school history: not Match Data Record SchoolHistory",
			ctx:  ctx,
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location,postal_code,prefecture,city,first_street,second_street,school,school_course,start_date,end_date
				Student 01 last name,Student 01 first name,Student 01 last name phonetic,Student 01 first name phonetic,student-01@example.com,1,0,,,,1,,,,,,,school_course-id-1;school_course-id-2,2022/01/02;2022/11/02,2023/01/02;2024/01/02`),
			},
			expectedResp: &pb.ImportStudentResponse{
				Errors: []*pb.ImportStudentResponse_ImportStudentError{
					{
						RowNumber: 2,
						Error:     "notFollowTemplate",
						FieldName: string(schoolStudentCSVHeader),
					},
				},
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				gradeOrganizationRepo.On("GetByGradeValues", ctx, db, mock.Anything).Once().Return(nil, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				domainLocationRepo.On("GetByPartnerInternalIDs", ctx, db, []string{"1"}).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "happy case with enrollment status history: 1 row success",
			ctx:  ctx,
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location,postal_code,prefecture,city,first_street,second_street,school,school_course,start_date,end_date,status_start_date
				Student 01 last name,Student 01 first name,Student 01 last name phonetic,Student 01 first name phonetic,student-01@example.com,1,0,,,,1,,,,,,school-partner-id-1;school-partner-id-2,school_course-id-1;school_course-id-2,,,2022/01/02`),
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, db, mock.Anything).Once().Return(nil, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainLocationRepo.On("GetByPartnerInternalIDs", ctx, db, []string{"1"}).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				gradeOrganizationRepo.On("GetByGradeValues", ctx, db, mock.Anything).Once().Return(nil, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{}, nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("CreateMultiple", ctx, db, mock.Anything).Once().Return(usrEmail, nil)
				studentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				schoolInfoRepo.On("GetBySchoolPartnerIDs", ctx, db, database.TextArray([]string{"school-partner-id-1", "school-partner-id-2"})).Once().Return([]*entity.SchoolInfo{
					{
						ID:      database.Text("school_info-id-1"),
						LevelID: database.Text("school_level-id-1"),
					},
					{
						ID:      database.Text("school_info-id-2"),
						LevelID: database.Text("school_level-id-2"),
					},
				}, nil)
				schoolCourseRepo.On("GetBySchoolCoursePartnerIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{
					{
						ID: database.Text("school_course-id-1"),
					},
					{
						ID: database.Text("school_course-id-2"),
					},
				}, nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				importUserEventRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return([]*entity.ImportUserEvent{}, nil)
				prefectureRepo.On("GetByPrefectureCode", ctx, mock.Anything, database.Text("")).Return(&entity.Prefecture{}, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
			},
		},
		{
			name: "happy case with enrollment status history: multiple row success",
			ctx:  ctx,
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location,postal_code,prefecture,city,first_street,second_street,school,school_course,start_date,end_date,status_start_date
				Student 01 last name,Student 01 first name,Student 01 last name phonetic,Student 01 first name phonetic,student-01@example.com,1;1,0,,,,1;2,,,,,,school-partner-id-1;school-partner-id-2,school_course-id-1;school_course-id-2,,,2022/01/02;2022/01/04`),
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				userRepo.On("GetByEmailInsensitiveCase", ctx, db, mock.Anything).Once().Return(nil, nil)
				userRepo.On("GetByPhone", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainLocationRepo.On("GetByPartnerInternalIDs", ctx, db, []string{"1", "2"}).Return(entity.DomainLocations{entity.NullDomainLocation{}, entity.NullDomainLocation{}}, nil)
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				gradeOrganizationRepo.On("GetByGradeValues", ctx, db, mock.Anything).Once().Return(nil, nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", ctx, tx, mock.Anything).Once().Return(&entity.UserGroupV2{}, nil)
				userAccessPathRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupsMemberRepo.On("AssignWithUserGroup", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				usrEmailRepo.On("CreateMultiple", ctx, db, mock.Anything).Once().Return(usrEmail, nil)
				studentRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				userGroupRepo.On("CreateMultiple", ctx, tx, mock.Anything).Once().Return(nil)
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return("", nil)
				schoolHistoryRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				schoolInfoRepo.On("GetBySchoolPartnerIDs", ctx, db, database.TextArray([]string{"school-partner-id-1", "school-partner-id-2"})).Once().Return([]*entity.SchoolInfo{
					{
						ID:      database.Text("school_info-id-1"),
						LevelID: database.Text("school_level-id-1"),
					},
					{
						ID:      database.Text("school_info-id-2"),
						LevelID: database.Text("school_level-id-2"),
					},
				}, nil)
				schoolCourseRepo.On("GetBySchoolCoursePartnerIDsAndSchoolIDs", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolCourse{
					{
						ID: database.Text("school_course-id-1"),
					},
					{
						ID: database.Text("school_course-id-2"),
					},
				}, nil)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				importUserEventRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return([]*entity.ImportUserEvent{}, nil)
				prefectureRepo.On("GetByPrefectureCode", ctx, mock.Anything, database.Text("")).Return(&entity.Prefecture{}, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
					{
						LocationID:   "2",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Times(4).Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil)
				schoolHistoryRepo.On("GetSchoolHistoriesByGradeIDAndStudentID", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entity.SchoolHistory{}, nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				mockConfigurationClient.On("GetConfigurationByKey", ctx, getConfigReq).Return(&mpb.GetConfigurationByKeyResponse{Configuration: configurationsDataLMS}, nil).Once()
			},
		},
		{
			name: "worst case with enrollment status history: data not sync",
			ctx:  ctx,
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location,postal_code,prefecture,city,first_street,second_street,school,school_course,start_date,end_date,status_start_date
				Student 01 last name,Student 01 first name,Student 01 last name phonetic,Student 01 first name phonetic,student-01@example.com,1;1,0,,,,1,,,,,,,,,,2022/01/02;2022/01/04`),
			}, expectedResp: &pb.ImportStudentResponse{
				Errors: []*pb.ImportStudentResponse_ImportStudentError{
					{
						RowNumber: 2,
						Error:     "notFollowTemplate",
						FieldName: string(locationStudentCSVHeader),
					},
				},
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				gradeOrganizationRepo.On("GetByGradeValues", ctx, db, mock.Anything).Once().Return(nil, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				mockConfigurationClient.On("GetConfigurationByKey", ctx, getConfigReq).Return(&mpb.GetConfigurationByKeyResponse{Configuration: configurationsDataLMS}, nil).Once()
			},
		},
		{
			name: "worst case with enrollment status history: invalid enrollment status",
			ctx:  ctx,
			req: &pb.ImportStudentRequest{
				Payload: []byte(`last_name,first_name,last_name_phonetic,first_name_phonetic,email,enrollment_status,grade,phone_number,birthday,gender,location,postal_code,prefecture,city,first_street,second_street,school,school_course,start_date,end_date,status_start_date
				Student 01 last name,Student 01 first name,Student 01 last name phonetic,Student 01 first name phonetic,student-01@example.com,1;2,0,,,,1;2,,,,,,,,,,2022/01/02;2022/01/04`),
			}, expectedResp: &pb.ImportStudentResponse{
				Errors: []*pb.ImportStudentResponse_ImportStudentError{
					{
						RowNumber: 2,
						Error:     "notFollowTemplate",
						FieldName: string(enrollmentStatusStudentCSVHeader),
					},
				},
			},
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entity.LegacyUser{
					Country: database.Text(cpb.Country_COUNTRY_VN.String()),
				}, nil)
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "1",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
					{
						LocationID:   "2",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				gradeOrganizationRepo.On("GetByGradeValues", ctx, db, mock.Anything).Once().Return(nil, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				mockConfigurationClient.On("GetConfigurationByKey", ctx, getConfigReq).Return(&mpb.GetConfigurationByKeyResponse{Configuration: configurationsDataERP}, nil).Once()
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log("====", testCase.name)
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.ManabieSchool),
				},
			}
			testCase.ctx = interceptors.ContextWithJWTClaims(ctx, claim)

			testCase.setup(testCase.ctx)

			resp, err := s.ImportStudent(testCase.ctx, testCase.req.(*pb.ImportStudentRequest))
			if err != nil {
				fmt.Println(err)
			}
			if resp != nil && len(resp.Errors) > 0 {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.expectedResp.(*pb.ImportStudentResponse)
				assert.Equal(t, len(expectedResp.Errors), len(resp.Errors))
				t.Log("Expect ", expectedResp.Errors)
				t.Log("resp ", resp.Errors)
				t.Log("========================================")
				for i, err := range resp.Errors {
					t.Logf("\nRow: %v , field name : %v , Err: %v ", err.RowNumber, err.FieldName, err.Error)
					assert.Equal(t, err.FieldName, expectedResp.Errors[i].FieldName)
					assert.Equal(t, err.RowNumber, expectedResp.Errors[i].RowNumber)
					assert.Contains(t, err.Error, expectedResp.Errors[i].Error)
				}
			}

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})

		mock.AssertExpectationsForObjects(t, db, tx, userRepo, studentRepo, userGroupRepo, userAccessPathRepo, locationRepo, firebaseAuth, userGroupV2Repo, userGroupsMemberRepo, orgRepo, tenantManager, jsm, schoolHistoryRepo, schoolCourseRepo, schoolInfoRepo, mockUnleashClient)
	}
}

func TestConvertLineCSVToStudentCSV(t *testing.T) {
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
	importUserEventRepo := new(mock_repositories.MockImportUserEventRepo)
	locationRepo := new(mock_locationRepo.MockLocationRepo)
	jsm := new(mock_nats.JetStreamManagement)
	firebaseAuth := new(mock_firebase.AuthClient)
	tenantManager := new(mock_multitenant.TenantManager)
	firebaseAuthClient := new(mock_multitenant.TenantClient)
	tenantClient := &mock_multitenant.TenantClient{}
	taskQueue := &mockTaskQueue{}
	userAddressRepo := new(mock_repositories.MockUserAddressRepo)
	prefectureRepo := new(mock_repositories.MockPrefectureRepo)
	domainGradeRepo := new(mock_repositories.MockDomainGradeRepo)
	gradeOrganizationRepo := new(mock_repositories.MockGradeOrganizationRepo)
	userPhoneNumberRepo := new(mock_repositories.MockUserPhoneNumberRepo)

	unleashClient := new(mock_unleash_client.UnleashClientInstance)
	schoolHistoryRepo := new(mock_repositories.MockSchoolHistoryRepo)
	schoolInfoRepo := new(mock_repositories.MockSchoolInfoRepo)
	schoolCourseRepo := new(mock_repositories.MockSchoolCourseRepo)

	domainEnrollmentStatusHistoryRepo := new(mock_repositories.MockDomainEnrollmentStatusHistoryRepo)
	domainUserAccessPathRepo := new(mock_repositories.MockDomainUserAccessPathRepo)
	domainLocationRepo := new(mock_repositories.MockDomainLocationRepo)

	mockConfigurationClient := new(mock_clients.MockConfigurationClient)

	userModifierService := UserModifierService{
		DB:                    db,
		OrganizationRepo:      orgRepo,
		UsrEmailRepo:          usrEmailRepo,
		UserRepo:              userRepo,
		StudentRepo:           studentRepo,
		UserGroupRepo:         userGroupRepo,
		UserAccessPathRepo:    userAccessPathRepo,
		LocationRepo:          locationRepo,
		FirebaseClient:        firebaseAuth,
		FirebaseAuthClient:    firebaseAuthClient,
		TenantManager:         tenantManager,
		JSM:                   jsm,
		DomainGradeRepo:       domainGradeRepo,
		GradeOrganizationRepo: gradeOrganizationRepo,
	}

	s := StudentService{
		DB:                          db,
		FirebaseAuthClient:          firebaseAuthClient,
		OrganizationRepo:            orgRepo,
		StudentRepo:                 studentRepo,
		UserRepo:                    userRepo,
		UsrEmailRepo:                usrEmailRepo,
		UserGroupRepo:               userGroupRepo,
		UserGroupV2Repo:             userGroupV2Repo,
		UserGroupsMemberRepo:        userGroupsMemberRepo,
		UserAccessPathRepo:          userAccessPathRepo,
		UserModifierService:         &userModifierService,
		JSM:                         jsm,
		ImportUserEventRepo:         importUserEventRepo,
		TaskQueue:                   taskQueue,
		UnleashClient:               unleashClient,
		GradeOrganizationRepo:       gradeOrganizationRepo,
		UserAddressRepo:             userAddressRepo,
		PrefectureRepo:              prefectureRepo,
		SchoolHistoryRepo:           schoolHistoryRepo,
		SchoolInfoRepo:              schoolInfoRepo,
		SchoolCourseRepo:            schoolCourseRepo,
		UserPhoneNumberRepo:         userPhoneNumberRepo,
		EnrollmentStatusHistoryRepo: domainEnrollmentStatusHistoryRepo,
		DomainUserAccessPathRepo:    domainUserAccessPathRepo,
		DomainLocationRepo:          domainLocationRepo,
		ConfigurationClient:         mockConfigurationClient,
	}

	getConfigReq := &mpb.GetConfigurationByKeyRequest{Key: constant.KeyEnrollmentStatusHistoryConfig}
	organizationID := "id"

	configurationsDataERP := &mpb.Configuration{
		Id:          organizationID,
		ConfigValue: "off",
	}

	usrEmail := []*entity.UsrEmail{
		{
			UsrID: database.Text("example-id"),
		},
	}
	hashConfig := mockScryptHash()

	payload1001Rows := "name,email,enrollment_status,grade,phone_number,birthday,gender,location"
	for i := 0; i < 1001; i++ {
		payload1001Rows += "\nStudent 01,student-01@example.com,1,1,0981143301,1999/01/12,1,location-01;location-02"
	}

	var builder strings.Builder
	sizeInMB := 1024 * 1024 * 10
	builder.Grow(sizeInMB)
	for i := 0; i < sizeInMB; i++ {
		builder.WriteByte(0)
	}

	testCases := []struct {
		name         string
		ctx          context.Context
		resourcePath string
		req          interface{}
		expectedErr  error
		setup        func(ctx context.Context)
		expectedResp interface{}
	}{
		{
			name: "err case when resourcePath value out of range",
			ctx:  ctx,
			req: &ImportStudentCSVField{
				LastName:          field.NewString("Student 01 last name"),
				FirstName:         field.NewString("Student 01 first name"),
				LastNamePhonetic:  field.NewString("Student 01 last name phonetic"),
				FirstNamePhonetic: field.NewString("Student 01 first name phonetic"),
				Email:             field.NewString("student-01@example.com"),
				EnrollmentStatus:  field.NewString("3"),
				Grade:             field.NewString("0"),
				PhoneNumber:       field.NewNullString(),
				Birthday:          field.NewNullString(),
				Gender:            field.NewNullString(),
				Location:          field.NewString("location-1"),
				PostalCode:        field.NewNullString(),
				Prefecture:        field.NewNullString(),
				City:              field.NewNullString(),
				FirstStreet:       field.NewNullString(),
				SecondStreet:      field.NewNullString(),
				School:            field.NewString("school-partner-id-1;school-partner-id-2"),
				SchoolCourse:      field.NewString(";"),
				StartDate:         field.NewString("2022/01/02;2022/11/02"),
				EndDate:           field.NewString("2023/01/02;2023/01/02"),
			},
			resourcePath: "-21474836481111111111111111111111111111",
			expectedErr:  errors.New("strconv.Atoi: strconv.Atoi: parsing \"-21474836481111111111111111111111111111\": value out of range, row: 1"),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				gradeOrganizationRepo.On("GetByGradeValues", ctx, db, mock.Anything).Once().Return(nil, nil)
				usrEmailRepo.On("CreateMultiple", ctx, db, mock.Anything).Once().Return(usrEmail, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "location-01",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				importUserEventRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return([]*entity.ImportUserEvent{}, nil)
				prefectureRepo.On("GetByPrefectureCode", ctx, mock.Anything, database.Text("")).Return(&entity.Prefecture{}, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "err when missing location",
			ctx:  ctx,
			req: &ImportStudentCSVField{
				LastName:          field.NewString("Student 01 last name"),
				FirstName:         field.NewString("Student 01 first name"),
				LastNamePhonetic:  field.NewString("Student 01 last name phonetic"),
				FirstNamePhonetic: field.NewString("Student 01 first name phonetic"),
				Email:             field.NewString("student-01@example.com"),
				EnrollmentStatus:  field.NewString("3"),
				Grade:             field.NewString("0"),
				PhoneNumber:       field.NewNullString(),
				Birthday:          field.NewNullString(),
				Gender:            field.NewNullString(),
				PostalCode:        field.NewNullString(),
				Prefecture:        field.NewNullString(),
				City:              field.NewNullString(),
				FirstStreet:       field.NewNullString(),
				SecondStreet:      field.NewNullString(),
				School:            field.NewString("school-partner-id-1;school-partner-id-2"),
				SchoolCourse:      field.NewString(";"),
				StartDate:         field.NewString("2022/01/02;2022/11/02"),
				EndDate:           field.NewString("2023/01/02;2023/01/02"),
			},
			resourcePath: "-2147483648",
			expectedResp: &pb.ImportStudentResponse_ImportStudentError{
				RowNumber: 3,
				Error:     "missingMandatory",
				FieldName: string(locationStudentCSVHeader),
			},
			setup: func(ctx context.Context) {
				// db.On("Begin", mock.Anything).Once().Return(tx, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				gradeOrganizationRepo.On("GetByGradeValues", ctx, db, mock.Anything).Once().Return(nil, nil)
				usrEmailRepo.On("CreateMultiple", ctx, db, mock.Anything).Once().Return(usrEmail, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				importUserEventRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return([]*entity.ImportUserEvent{}, nil)

				prefectureRepo.On("GetByPrefectureCode", ctx, mock.Anything, database.Text("")).Return(&entity.Prefecture{}, nil)
			},
		},
		{
			name: "happy case normal",
			ctx:  ctx,
			req: &ImportStudentCSVField{
				LastName:          field.NewString("Student 01 last name"),
				FirstName:         field.NewString("Student 01 first name"),
				LastNamePhonetic:  field.NewString("Student 01 last name phonetic"),
				FirstNamePhonetic: field.NewString("Student 01 first name phonetic"),
				Email:             field.NewString("student-01@example.com"),
				EnrollmentStatus:  field.NewString("1"),
				Grade:             field.NewString("0"),
				PhoneNumber:       field.NewNullString(),
				Birthday:          field.NewNullString(),
				Gender:            field.NewNullString(),
				Location:          field.NewString("location-01"),
				PostalCode:        field.NewNullString(),
				Prefecture:        field.NewNullString(),
				City:              field.NewNullString(),
				FirstStreet:       field.NewNullString(),
				SecondStreet:      field.NewNullString(),
				School:            field.NewString("school-partner-id-1;school-partner-id-2"),
				SchoolCourse:      field.NewString(";"),
				StartDate:         field.NewString("2022/01/02;2022/11/02"),
				EndDate:           field.NewString("2023/01/02;2023/01/02"),
				StatusStartDate:   field.NewString("2023/01/02"),
			},
			resourcePath: "-2147483648",
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				domainGradeRepo.On("GetByIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				domainGradeRepo.On("GetByPartnerInternalIDs", ctx, db, mock.Anything).Once().Return(nil, nil)
				gradeOrganizationRepo.On("GetByGradeValues", ctx, db, mock.Anything).Once().Return(nil, nil)
				usrEmailRepo.On("CreateMultiple", ctx, db, mock.Anything).Once().Return(usrEmail, nil)
				tenantClient.On("GetHashConfig").Once().Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(&internal_auth_user.ImportUsersResult{}, nil)
				locationRepo.On("GetLocationsByPartnerInternalIDs", ctx, tx, mock.Anything).Return([]*domain.Location{
					{
						LocationID:   "location-01",
						Name:         "center",
						ResourcePath: fmt.Sprint(constants.ManabieSchool),
					},
				}, nil).Once()
				schoolInfoRepo.On("GetBySchoolPartnerIDs", ctx, db, database.TextArray([]string{"school-partner-id-1", "school-partner-id-2"})).Once().Return([]*entity.SchoolInfo{
					{
						ID:      database.Text("school_info-id-1"),
						LevelID: database.Text("school_level-id-1"),
					},
					{
						ID:      database.Text("school_info-id-2"),
						LevelID: database.Text("school_level-id-2"),
					},
				}, nil)
				importUserEventRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return([]*entity.ImportUserEvent{}, nil)
				prefectureRepo.On("GetByPrefectureCode", ctx, mock.Anything, database.Text("")).Return(&entity.Prefecture{}, nil)
				domainLocationRepo.On("GetByPartnerInternalIDs", ctx, db, []string{"location-01"}).Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				mockConfigurationClient.On("GetConfigurationByKey", ctx, getConfigReq).Return(&mpb.GetConfigurationByKeyResponse{Configuration: configurationsDataERP}, nil).Once()
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log("====", testCase.name)
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.ManabieSchool),
				},
			}
			testCase.ctx = interceptors.ContextWithJWTClaims(ctx, claim)

			testCase.setup(testCase.ctx)

			_, resp, err := s.convertLineCSVToStudentCSV(testCase.ctx, testCase.req.(*ImportStudentCSVField), 1, "COUNTRY_VN", testCase.resourcePath)
			if err != nil {
				fmt.Println(err)
			}
			if resp != nil && len(resp.Error) > 0 {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.expectedResp.(*pb.ImportStudentResponse_ImportStudentError)
				assert.Equal(t, len(expectedResp.Error), len(resp.Error))

				assert.Equal(t, expectedResp.RowNumber, resp.RowNumber)
				assert.Contains(t, expectedResp.FieldName, resp.FieldName)
				assert.Equal(t, expectedResp.Error, resp.Error)
			}

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})

		mock.AssertExpectationsForObjects(t, db, tx, userRepo, studentRepo, userGroupRepo, userAccessPathRepo, locationRepo, firebaseAuth, userGroupV2Repo, userGroupsMemberRepo, orgRepo, tenantManager, jsm, schoolHistoryRepo, unleashClient)
	}
}

func TestCheckInvalidSchoolHistoryDataImport(t *testing.T) {
	t.Parallel()
	_, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	type reqTestCase struct {
		schools, schoolCourses, startDates, endDates int
	}
	testCases := []struct {
		name         string
		req          reqTestCase
		expectedResp interface{}
	}{
		{
			name: "case - endDateStudentCSVHeader",
			req: reqTestCase{
				schools:       1,
				schoolCourses: 1,
				startDates:    1,
				endDates:      2,
			},
			expectedResp: studentCSVHeader(endDateStudentCSVHeader),
		},
		{
			name: "case - startDateStudentCSVHeader",
			req: reqTestCase{
				schools:       1,
				schoolCourses: 1,
				startDates:    2,
				endDates:      1,
			},
			expectedResp: studentCSVHeader(startDateStudentCSVHeader),
		},
		{
			name: "case - schoolCourseStudentCSVHeader",
			req: reqTestCase{
				schools:       1,
				schoolCourses: 2,
				startDates:    1,
				endDates:      1,
			},
			expectedResp: studentCSVHeader(schoolCourseStudentCSVHeader),
		},
		{
			name: "case - schoolStudentCSVHeader",
			req: reqTestCase{
				schools:       2,
				schoolCourses: 1,
				startDates:    1,
				endDates:      1,
			},
			expectedResp: studentCSVHeader(schoolStudentCSVHeader),
		},
		{
			name: "case return empty string",
			req: reqTestCase{
				schools:       1,
				schoolCourses: 1,
				startDates:    1,
				endDates:      1,
			},
			expectedResp: studentCSVHeader(""),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log("====", testCase.name)

			resp := checkInvalidSchoolHistoryDataImport(testCase.req.schools, testCase.req.schoolCourses, testCase.req.startDates, testCase.req.endDates)

			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}

func TestCheckInvalidEnrollmentStatusHistoriesDataImport(t *testing.T) {
	t.Parallel()
	_, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	type reqTestCase struct {
		locations, enrollmentStatuses, statusStartDates int
	}
	testCases := []struct {
		name         string
		req          reqTestCase
		expectedResp interface{}
	}{
		{
			name: "case - endDateStudentCSVHeader",
			req: reqTestCase{
				locations:          1,
				enrollmentStatuses: 1,
				statusStartDates:   2,
			},
			expectedResp: studentCSVHeader(statusStartDateStudentCSVHeader),
		},
		{
			name: "case - startDateStudentCSVHeader",
			req: reqTestCase{
				locations:          1,
				enrollmentStatuses: 2,
				statusStartDates:   1,
			},
			expectedResp: studentCSVHeader(enrollmentStatusStudentCSVHeader),
		},
		{
			name: "case - schoolStudentCSVHeader",
			req: reqTestCase{
				locations:          2,
				enrollmentStatuses: 1,
				statusStartDates:   1,
			},
			expectedResp: studentCSVHeader(locationStudentCSVHeader),
		},
		{
			name: "case return empty string",
			req: reqTestCase{
				locations:          1,
				enrollmentStatuses: 1,
				statusStartDates:   1,
			},
			expectedResp: studentCSVHeader(""),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log("====", testCase.name)

			resp := checkInvalidEnrollmentStatusHistoriesDataImport(testCase.req.locations, testCase.req.enrollmentStatuses, testCase.req.statusStartDates)

			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}

func Test_importCsvValidateTag(t *testing.T) {
	id1 := idutil.ULIDNow()
	id2 := idutil.ULIDNow()
	partnerID1 := field.NewString(fmt.Sprintf("partner-id-%s", id1))
	partnerID2 := field.NewString(fmt.Sprintf("partner-id-%s", id2))

	type args struct {
		role               string
		partnerInternalIDs []string
		tags               entity.DomainTags
	}
	tests := []struct {
		name string
		args func(t *testing.T) args

		wantErr    bool
		inspectErr func(err error, t *testing.T)
	}{
		{
			name: "happy case: valid for student",
			args: func(t *testing.T) args {
				return args{
					role:               constant.RoleStudent,
					partnerInternalIDs: []string{partnerID1.RawValue(), partnerID2.RawValue()},
					tags: entity.DomainTags([]entity.DomainTag{
						createMockDomainTagWithType(id1, pb.UserTagType_USER_TAG_TYPE_STUDENT),
						createMockDomainTagWithType(id2, pb.UserTagType_USER_TAG_TYPE_STUDENT_DISCOUNT),
					}),
				}
			},
			wantErr: false,
		},
		{
			name: "happy case: valid for parent",
			args: func(t *testing.T) args {
				return args{
					role:               constant.RoleParent,
					partnerInternalIDs: []string{partnerID1.RawValue(), partnerID2.RawValue()},
					tags: entity.DomainTags([]entity.DomainTag{
						createMockDomainTagWithType(id1, pb.UserTagType_USER_TAG_TYPE_PARENT),
						createMockDomainTagWithType(id2, pb.UserTagType_USER_TAG_TYPE_PARENT_DISCOUNT),
					}),
				}
			},
			wantErr: false,
		},
		{
			name: "tag is not for student",
			args: func(t *testing.T) args {
				return args{
					role:               constant.RoleStudent,
					partnerInternalIDs: []string{partnerID1.RawValue(), partnerID2.RawValue()},
					tags: entity.DomainTags([]entity.DomainTag{
						createMockDomainTagWithType(id1, pb.UserTagType_USER_TAG_TYPE_PARENT),
						createMockDomainTagWithType(id2, pb.UserTagType_USER_TAG_TYPE_STUDENT_DISCOUNT),
					}),
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				assert.Equal(t, err, errNotFollowTemplate)
			},
		},
		{
			name: "tag is not for parent",
			args: func(t *testing.T) args {
				return args{
					role:               constant.RoleParent,
					partnerInternalIDs: []string{partnerID1.RawValue(), partnerID2.RawValue()},
					tags: entity.DomainTags([]entity.DomainTag{
						createMockDomainTagWithType(id1, pb.UserTagType_USER_TAG_TYPE_PARENT),
						createMockDomainTagWithType(id2, pb.UserTagType_USER_TAG_TYPE_STUDENT_DISCOUNT),
					}),
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				assert.Equal(t, err, errNotFollowParentTemplate)
			},
		},
		{
			name: "tag is not existed",
			args: func(t *testing.T) args {
				return args{
					role:               constant.RoleStudent,
					partnerInternalIDs: []string{partnerID1.RawValue(), partnerID2.RawValue()},
					tags: entity.DomainTags([]entity.DomainTag{
						createMockDomainTagWithType(id1, pb.UserTagType_USER_TAG_TYPE_STUDENT),
					}),
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				assert.Equal(t, err, errNotFollowTemplate)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			err := importCsvValidateTag(tArgs.role, tArgs.partnerInternalIDs, tArgs.tags)

			if (err != nil) != tt.wantErr {
				t.Fatalf("importCsvValidateTag error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}
