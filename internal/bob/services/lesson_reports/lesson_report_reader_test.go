package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestRetrievePartnerDomain(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// mock structs
	db := &mock_database.Ext{}
	schoolAdminRepo := &mock_repositories.MockSchoolAdminRepo{}
	configRepo := &mock_repositories.MockConfigRepo{}
	userRepo := &mock_repositories.MockUserRepo{}
	teacherRepo := &mock_repositories.MockTeacherRepo{}
	studentRepo := &mock_repositories.MockStudentRepo{}
	var schoolIDs pgtype.Int4Array
	_ = schoolIDs.Set([]int{1})

	tcs := []struct {
		name        string
		reqUserID   string
		req         *bpb.GetPartnerDomainRequest
		setup       func(context.Context)
		expectedRes *bpb.GetPartnerDomainResponse
		hasError    bool
	}{
		{
			name:      "School admin get partner domain bo has in configs",
			reqUserID: "admin-1",
			req: &bpb.GetPartnerDomainRequest{
				Type: bpb.DomainType_DOMAIN_TYPE_BO,
			},
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupSchoolAdmin, nil)
				schoolAdminRepo.On("Get", ctx, mock.Anything, mock.Anything).Once().Return(&entities.SchoolAdmin{SchoolID: database.Int4(1)}, nil)
				configRepo.On("Retrieve",
					mock.Anything, mock.Anything,
					database.Text("COUNTRY_MASTER"),
					database.Text("lesson"),
					database.TextArray([]string{"domain_local_bo_1"})).
					Return([]*entities.Config{
						{
							Key:     database.Text("domain_local_bo_1"),
							Group:   database.Text("lesson"),
							Country: database.Text("COUNTRY_MASTER"),
							Value:   database.Text("https://staging-jprep-school-portal.web.app/"),
						},
					}, nil)
			},
			expectedRes: &bpb.GetPartnerDomainResponse{
				Domain: "https://staging-jprep-school-portal.web.app/",
			},
		},
		{
			name:      "School admin get partner domain teacher has in configs",
			reqUserID: "admin-1",
			req: &bpb.GetPartnerDomainRequest{
				Type: bpb.DomainType_DOMAIN_TYPE_TEACHER,
			},
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupSchoolAdmin, nil)
				schoolAdminRepo.On("Get", ctx, mock.Anything, mock.Anything).Once().Return(&entities.SchoolAdmin{SchoolID: database.Int4(1)}, nil)
				configRepo.On("Retrieve",
					mock.Anything, mock.Anything,
					database.Text("COUNTRY_MASTER"),
					database.Text("lesson"),
					database.TextArray([]string{"domain_local_teacher_1"})).
					Return([]*entities.Config{
						{
							Key:     database.Text("domain_local_teacher_1"),
							Group:   database.Text("lesson"),
							Country: database.Text("COUNTRY_MASTER"),
							Value:   database.Text("https://teacher.com/"),
						},
					}, nil)
			},
			expectedRes: &bpb.GetPartnerDomainResponse{
				Domain: "https://teacher.com/",
			},
		},
		{
			name:      "School admin get partner domain teacher has in configs",
			reqUserID: "admin-1",
			req: &bpb.GetPartnerDomainRequest{
				Type: bpb.DomainType_DOMAIN_TYPE_LEARNER,
			},
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupSchoolAdmin, nil)
				schoolAdminRepo.On("Get", ctx, mock.Anything, mock.Anything).Once().Return(&entities.SchoolAdmin{SchoolID: database.Int4(1)}, nil)
				configRepo.On("Retrieve",
					mock.Anything, mock.Anything,
					database.Text("COUNTRY_MASTER"),
					database.Text("lesson"),
					database.TextArray([]string{"domain_local_learner_1"})).
					Return([]*entities.Config{
						{
							Key:     database.Text("domain_local_learner_1"),
							Group:   database.Text("lesson"),
							Country: database.Text("COUNTRY_MASTER"),
							Value:   database.Text("https://learner.com/"),
						},
					}, nil)
			},
			expectedRes: &bpb.GetPartnerDomainResponse{
				Domain: "https://learner.com/",
			},
		},
		{
			name:      "School admin get partner domain bo without configs",
			reqUserID: "admin-1",
			req: &bpb.GetPartnerDomainRequest{
				Type: bpb.DomainType_DOMAIN_TYPE_BO,
			},
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupSchoolAdmin, nil)
				schoolAdminRepo.On("Get", ctx, mock.Anything, mock.Anything).Once().Return(&entities.SchoolAdmin{SchoolID: database.Int4(2)}, nil)
				configRepo.On("Retrieve",
					mock.Anything, mock.Anything,
					database.Text("COUNTRY_MASTER"),
					database.Text("lesson"),
					database.TextArray([]string{"domain_local_bo_2"})).
					Return([]*entities.Config{}, nil)
			},
			expectedRes: &bpb.GetPartnerDomainResponse{
				Domain: "https://staging-jprep-school-portal.web.app/",
			},
		},
		{
			name:      "School admin get partner domain teacher without configs",
			reqUserID: "admin-1",
			req: &bpb.GetPartnerDomainRequest{
				Type: bpb.DomainType_DOMAIN_TYPE_TEACHER,
			},
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupSchoolAdmin, nil)
				schoolAdminRepo.On("Get", ctx, mock.Anything, mock.Anything).Once().Return(&entities.SchoolAdmin{SchoolID: database.Int4(2)}, nil)
				configRepo.On("Retrieve",
					mock.Anything, mock.Anything,
					database.Text("COUNTRY_MASTER"),
					database.Text("lesson"),
					database.TextArray([]string{"domain_local_teacher_2"})).
					Return([]*entities.Config{}, nil)
			},
			expectedRes: &bpb.GetPartnerDomainResponse{
				Domain: "https://staging-jprep-teacher.web.app/",
			},
		},
		{
			name:      "School admin get partner domain learner without configs",
			reqUserID: "admin-1",
			req: &bpb.GetPartnerDomainRequest{
				Type: bpb.DomainType_DOMAIN_TYPE_LEARNER,
			},
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupSchoolAdmin, nil)
				schoolAdminRepo.On("Get", ctx, mock.Anything, mock.Anything).Once().Return(&entities.SchoolAdmin{SchoolID: database.Int4(2)}, nil)
				configRepo.On("Retrieve",
					mock.Anything, mock.Anything,
					database.Text("COUNTRY_MASTER"),
					database.Text("lesson"),
					database.TextArray([]string{"domain_local_learner_2"})).
					Return([]*entities.Config{}, nil)
			},
			expectedRes: &bpb.GetPartnerDomainResponse{
				Domain: "https://staging-jprep-learner.web.app/",
			},
		},
		{
			name:      "Teacher get partner domain bo has in configs",
			reqUserID: "teacher-1",
			req: &bpb.GetPartnerDomainRequest{
				Type: bpb.DomainType_DOMAIN_TYPE_BO,
			},
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupTeacher, nil)
				teacherRepo.On("FindByID", ctx, mock.Anything, mock.Anything).Once().
					Return(&entities.Teacher{SchoolIDs: schoolIDs}, nil)
				configRepo.On("Retrieve",
					mock.Anything, mock.Anything,
					database.Text("COUNTRY_MASTER"),
					database.Text("lesson"),
					database.TextArray([]string{"domain_local_bo_1"})).
					Return([]*entities.Config{
						{
							Key:     database.Text("domain_local_bo_1"),
							Group:   database.Text("lesson"),
							Country: database.Text("COUNTRY_MASTER"),
							Value:   database.Text("https://staging-jprep-school-portal.web.app/"),
						},
					}, nil)
			},
			expectedRes: &bpb.GetPartnerDomainResponse{
				Domain: "https://staging-jprep-school-portal.web.app/",
			},
		},
		{
			name:      "Student get partner domain bo has in configs",
			reqUserID: "learner-1",
			req: &bpb.GetPartnerDomainRequest{
				Type: bpb.DomainType_DOMAIN_TYPE_BO,
			},
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupStudent, nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().
					Return(&entities.Student{
						ID:       database.Text("1"),
						SchoolID: database.Int4(4),
					}, nil)
				configRepo.On("Retrieve",
					mock.Anything, mock.Anything,
					database.Text("COUNTRY_MASTER"),
					database.Text("lesson"),
					database.TextArray([]string{"domain_local_bo_4"})).
					Return([]*entities.Config{
						{
							Key:     database.Text("domain_local_bo_4"),
							Group:   database.Text("lesson"),
							Country: database.Text("COUNTRY_MASTER"),
							Value:   database.Text("https://staging-jprep-school-portal.web.app/"),
						},
					}, nil)
			},
			expectedRes: &bpb.GetPartnerDomainResponse{
				Domain: "https://staging-jprep-school-portal.web.app/",
			},
		},
		{
			name:      "Admin get partner domain bo",
			reqUserID: "admin-1",
			req: &bpb.GetPartnerDomainRequest{
				Type: bpb.DomainType_DOMAIN_TYPE_BO,
			},
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupAdmin, nil)
			},
			expectedRes: &bpb.GetPartnerDomainResponse{
				Domain: "",
			},
			hasError: true,
		},
		{
			name:      "User get domain error when cannot get UserGroup",
			reqUserID: "admin-1",
			req: &bpb.GetPartnerDomainRequest{
				Type: bpb.DomainType_DOMAIN_TYPE_BO,
			},
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupAdmin, errors.New("error"))
			},
			expectedRes: &bpb.GetPartnerDomainResponse{
				Domain: "",
			},
			hasError: true,
		},
		{
			name:      "User get domain error when cannot get schoolID",
			reqUserID: "admin-1",
			req: &bpb.GetPartnerDomainRequest{
				Type: bpb.DomainType_DOMAIN_TYPE_BO,
			},
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupSchoolAdmin, nil)
				schoolAdminRepo.On("Get", ctx, mock.Anything, mock.Anything).Once().Return(&entities.SchoolAdmin{}, errors.New("error"))
			},
			expectedRes: &bpb.GetPartnerDomainResponse{
				Domain: "",
			},
			hasError: true,
		},
		{
			name:      "User get domain error when get config error",
			reqUserID: "admin-1",
			req: &bpb.GetPartnerDomainRequest{
				Type: bpb.DomainType_DOMAIN_TYPE_BO,
			},
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupSchoolAdmin, nil)
				schoolAdminRepo.On("Get", ctx, mock.Anything, mock.Anything).Once().Return(&entities.SchoolAdmin{SchoolID: database.Int4(3)}, nil)
				configRepo.On("Retrieve",
					mock.Anything, mock.Anything,
					database.Text("COUNTRY_MASTER"),
					database.Text("lesson"),
					database.TextArray([]string{"domain_local_bo_3"})).
					Return([]*entities.Config{}, errors.New("error"))
			},
			expectedRes: &bpb.GetPartnerDomainResponse{
				Domain: "",
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctxT := interceptors.ContextWithUserID(ctx, tc.reqUserID)
			tc.setup(ctxT)
			srv := LessonReportReaderService{
				Cfg: &configurations.Config{
					Common: configs.CommonConfig{
						Environment: "local",
					},
					Partner: configs.PartnerConfig{
						DomainBo:      "https://staging-jprep-school-portal.web.app/",
						DomainTeacher: "https://staging-jprep-teacher.web.app/",
						DomainLearner: "https://staging-jprep-learner.web.app/",
					},
				},
				DB:              db,
				UserRepo:        userRepo,
				SchoolAdminRepo: schoolAdminRepo,
				ConfigRepo:      configRepo,
				TeacherRepo:     teacherRepo,
				StudentRepo:     studentRepo,
			}
			actualRes, err := srv.RetrievePartnerDomain(ctxT, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.EqualValues(t, tc.expectedRes, actualRes)
				mock.AssertExpectationsForObjects(t, db, userRepo, schoolAdminRepo, configRepo, teacherRepo, studentRepo)
			}
		})
	}
}
