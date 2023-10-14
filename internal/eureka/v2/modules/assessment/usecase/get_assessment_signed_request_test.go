package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	book_domain "github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/constants"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_asm_learnosity_repo "github.com/manabie-com/backend/mock/eureka/v2/modules/assessment/repository/learnosity"
	assessment_mock_postgres "github.com/manabie-com/backend/mock/eureka/v2/modules/assessment/repository/postgres"
	book_mock_postgres "github.com/manabie-com/backend/mock/eureka/v2/modules/book/repository/postgres"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_learnosity "github.com/manabie-com/backend/mock/golibs/learnosity"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAssessmentUsecaseImpl_GetAssessmentSignedRequest(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}
	learnosityConfig := configurations.LearnosityConfig{
		ConsumerKey:    "consumer_key",
		ConsumerSecret: "consumer_secret",
	}
	mockLMRepo := &book_mock_postgres.MockLearningMaterialRepo{}
	mockAssessmentRepo := &assessment_mock_postgres.MockAssessmentRepo{}
	mockAssessmentSessionRepo := &assessment_mock_postgres.MockAssessmentSessionRepo{}
	mockHTTP := &mock_learnosity.HTTP{}
	mockDataAPI := &mock_learnosity.DataAPI{}
	mockSessionRepo := &mock_asm_learnosity_repo.MockSessionRepo{}

	assessmentUsecase := &AssessmentUsecaseImpl{
		DB:                    mockDB,
		LearnosityConfig:      learnosityConfig,
		HTTP:                  mockHTTP,
		DataAPI:               mockDataAPI,
		AssessmentRepo:        mockAssessmentRepo,
		AssessmentSessionRepo: mockAssessmentSessionRepo,
		LearningMaterialRepo:  mockLMRepo,
		LearnositySessionRepo: mockSessionRepo,
	}

	type Request struct {
		Session    domain.Session
		HostDomain string
		Config     string
	}

	testCases := []struct {
		Name             string
		Ctx              context.Context
		Request          any
		Setup            func(ctx context.Context)
		ExpectedResponse any
		ExpectedError    error
	}{
		{
			Name: "learning material not found",
			Ctx:  ctx,
			Request: Request{
				Session: domain.Session{
					CourseID:           "course_id",
					LearningMaterialID: "learning_material_id",
					UserID:             "user_id",
				},
				HostDomain: "localhost",
				Config:     "",
			},
			Setup: func(ctx context.Context) {
				mockLMRepo.On("GetByID", ctx, mock.Anything).Once().Return(book_domain.LearningMaterial{}, pgx.ErrNoRows)
			},
			ExpectedResponse: "",
			ExpectedError:    errors.New("LearningMaterialRepo.GetByID", pgx.ErrNoRows),
		},
		{
			Name: "happy case: no existed assessment",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Request: Request{
				Session: domain.Session{
					CourseID:           "course_id",
					LearningMaterialID: "learning_material_id",
					UserID:             "user_id",
				},
				HostDomain: "localhost",
				Config:     "",
			},
			Setup: func(ctx context.Context) {
				mockLMRepo.On("GetByID", ctx, mock.Anything).Once().Return(book_domain.LearningMaterial{
					Name: "lm_name",
					Type: constants.LearningObjective,
				}, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockAssessmentRepo.On("GetOneByLMAndCourseID", ctx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(nil, errors.NewNoRowsExistedError("database.Select", nil))
				mockAssessmentRepo.On("Upsert", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return("id", nil)
				mockAssessmentSessionRepo.On("GetLatestByIdentity", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().
					Return(domain.Session{}, errors.NewNoRowsExistedError("database.Select", nil))
				mockAssessmentSessionRepo.On("Insert", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
			ExpectedResponse: "{\"request\":\"{\\\"activity_id\\\":\\\"id\\\",\\\"activity_template_id\\\":\\\"learning_material_id\\\",\\\"name\\\":\\\"lm_name\\\",\\\"rendering_type\\\":\\\"assess\\\",\\\"session_id\\\":\\\"",
			ExpectedError:    nil,
		},
		{
			Name: "happy case: existed assessment session without status completed",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Request: Request{
				Session: domain.Session{
					CourseID:           "course_id",
					LearningMaterialID: "learning_material_id",
					UserID:             "user_id",
				},
				HostDomain: "localhost",
				Config:     "",
			},
			Setup: func(ctx context.Context) {
				mockLMRepo.On("GetByID", ctx, mock.Anything).Once().Return(book_domain.LearningMaterial{
					Name: "lm_name",
					Type: constants.LearningObjective,
				}, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockAssessmentRepo.On("GetOneByLMAndCourseID", ctx, mock.Anything, mock.Anything, mock.Anything).Once().
					Return(nil, errors.NewNoRowsExistedError("database.Select", nil))
				mockAssessmentRepo.On("Upsert", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return("id", nil)
				mockAssessmentSessionRepo.On("GetLatestByIdentity", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(domain.Session{
					ID: "session_id",
				}, nil)

				mockSessionRepo.On("GetSessionStatuses", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
			ExpectedResponse: "{\"request\":\"{\\\"activity_id\\\":\\\"\\\",\\\"activity_template_id\\\":\\\"learning_material_id\\\",\\\"name\\\":\\\"lm_name\\\",\\\"rendering_type\\\":\\\"assess\\\",\\\"session_id\\\":\\\"session_id\\\",\\\"user_id\\\":\\\"user_id\\\"}\",\"security\":{\"consumer_key\":\"consumer_key\",\"domain\":\"localhost\",\"signature\":\"",
			ExpectedError:    nil,
		},
		{
			Name: "happy case: existed assessment session with status completed",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Request: Request{
				Session: domain.Session{
					CourseID:           "course_id",
					LearningMaterialID: "learning_material_id",
					UserID:             "user_id",
				},
				HostDomain: "localhost",
				Config:     "",
			},
			Setup: func(ctx context.Context) {
				mockLMRepo.On("GetByID", ctx, mock.Anything).Once().Return(book_domain.LearningMaterial{
					Name: "lm_name",
					Type: constants.LearningObjective,
				}, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockAssessmentRepo.On("GetOneByLMAndCourseID", ctx, mock.Anything, mock.Anything, mock.Anything).Once().
					Return(nil, errors.NewNoRowsExistedError("database.Select", nil))
				mockAssessmentRepo.On("Upsert", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return("id", nil)
				mockAssessmentSessionRepo.On("GetLatestByIdentity", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(domain.Session{
					ID: "session_id",
				}, nil)

				mockSessionRepo.On("GetSessionStatuses", ctx, mock.Anything, mock.Anything).Once().Return([]domain.Session{
					{
						Status: domain.SessionStatusCompleted,
					},
				}, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
				mockAssessmentSessionRepo.On("Insert", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			ExpectedResponse: "{\"request\":\"{\\\"activity_id\\\":\\\"id\\\",\\\"activity_template_id\\\":\\\"learning_material_id\\\",\\\"name\\\":\\\"lm_name\\\",\\\"rendering_type\\\":\\\"assess\\\",\\\"session_id\\\":\\\"",
			ExpectedError:    nil,
		},
		{
			Name: "happy case: assessment session with custom config object from request",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Request: Request{
				Session: domain.Session{
					CourseID:           "course_id",
					LearningMaterialID: "learning_material_id",
					UserID:             "user_id",
				},
				HostDomain: "localhost",
				Config:     "{\"regions\": \"main\"}",
			},
			Setup: func(ctx context.Context) {
				mockLMRepo.On("GetByID", ctx, mock.Anything).Once().Return(book_domain.LearningMaterial{
					Name: "lm_name",
					Type: constants.LearningObjective,
				}, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockAssessmentRepo.On("GetOneByLMAndCourseID", ctx, mock.Anything, mock.Anything, mock.Anything).Once().
					Return(nil, errors.NewNoRowsExistedError("database.Select", nil))
				mockAssessmentRepo.On("Upsert", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return("id", nil)
				mockAssessmentSessionRepo.On("GetLatestByIdentity", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().
					Return(domain.Session{}, errors.NewNoRowsExistedError("database.Select", nil))
				mockAssessmentSessionRepo.On("Insert", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
			ExpectedResponse: "{\"request\":\"{\\\"activity_id\\\":\\\"id\\\",\\\"activity_template_id\\\":\\\"learning_material_id\\\",\\\"config\\\":{\\\"regions\\\":\\\"main\\\"},\\\"name\\\":\\\"lm_name\\\",\\\"rendering_type\\\":\\\"assess\\\",\\\"session_id\\\":\\\"",
			ExpectedError:    nil,
		},
		{
			Name: "happy case: no existed assessment session - FlashCard Type",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Request: Request{
				Session: domain.Session{
					CourseID:           "course_id",
					LearningMaterialID: "learning_material_id",
					UserID:             "user_id",
				},
				HostDomain: "localhost",
				Config:     "",
			},
			Setup: func(ctx context.Context) {
				mockLMRepo.On("GetByID", ctx, mock.Anything).Once().Return(book_domain.LearningMaterial{
					Name: "lm_name",
					Type: constants.FlashCard,
				}, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockAssessmentRepo.On("GetOneByLMAndCourseID", ctx, mock.Anything, mock.Anything, mock.Anything).Once().
					Return(nil, errors.NewNoRowsExistedError("database.Select", nil))
				mockAssessmentRepo.On("Upsert", ctx, mock.Anything, mock.Anything,
					mock.MatchedBy(func(assesment domain.Assessment) bool {
						assert.Equal(t, "learning_material_id", assesment.LearningMaterialID)
						assert.Equal(t, "course_id", assesment.CourseID)
						assert.Equal(t, constants.FlashCard, assesment.LearningMaterialType)
						return true
					}),
				).Once().Return("id", nil)
				mockAssessmentSessionRepo.On("GetLatestByIdentity", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().
					Return(domain.Session{}, errors.NewNoRowsExistedError("database.Select", nil))
				mockAssessmentSessionRepo.On("Insert", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
			ExpectedResponse: "{\"request\":\"{\\\"activity_id\\\":\\\"id\\\",\\\"activity_template_id\\\":\\\"learning_material_id\\\",\\\"name\\\":\\\"lm_name\\\",\\\"rendering_type\\\":\\\"assess\\\",\\\"session_id\\\":\\\"",
			ExpectedError:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Setup(tc.Ctx)
			res, err := assessmentUsecase.GetAssessmentSignedRequest(tc.Ctx, tc.Request.(Request).Session, tc.Request.(Request).HostDomain, tc.Request.(Request).Config)
			if err != nil {
				assert.Equal(t, tc.ExpectedError.Error(), err.Error())
			} else {
				assert.Contains(t, res, tc.ExpectedResponse.(string))
			}
		})
	}
}
