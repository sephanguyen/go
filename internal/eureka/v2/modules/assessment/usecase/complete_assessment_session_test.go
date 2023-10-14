package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/constants"
	mock_asm_learnosity_repo "github.com/manabie-com/backend/mock/eureka/v2/modules/assessment/repository/learnosity"
	assessment_mock_postgres "github.com/manabie-com/backend/mock/eureka/v2/modules/assessment/repository/postgres"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
)

func TestAssessmentUsecaseImpl_CompleteAssessmentSession(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	t.Run("happy case", func(t *testing.T) {
		// arrange
		mockLearnositySessionRepo := &mock_asm_learnosity_repo.MockSessionRepo{}

		mockDB := testutil.NewMockDB()
		mockTx := &mock_database.Tx{}
		mockSessionRepo := &assessment_mock_postgres.MockAssessmentSessionRepo{}
		mockAssessmentRepo := &assessment_mock_postgres.MockAssessmentRepo{}
		mockSubmissionRepo := &assessment_mock_postgres.MockSubmissionRepo{}
		usecase := &AssessmentUsecaseImpl{
			DB:                    mockDB.DB,
			LearnositySessionRepo: mockLearnositySessionRepo,
			AssessmentSessionRepo: mockSessionRepo,
			AssessmentRepo:        mockAssessmentRepo,
			SubmissionRepo:        mockSubmissionRepo,
		}

		completedAt := time.Date(2023, 01, 01, 0, 0, 0, 0, time.UTC)
		mockLearnositySessionRepo.On("GetSessionResponses", ctx, mock.Anything, mock.Anything).Once().Return(domain.Sessions{
			{
				ID:          "session_id",
				MaxScore:    8,
				GradedScore: 4,
				Status:      domain.SessionStatusCompleted,
				CreatedAt:   time.Date(2023, 01, 01, 0, 0, 0, 0, time.UTC),
				CompletedAt: &completedAt,
			},
		}, nil)

		mockDB.DB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
		mockSessionRepo.On("UpdateStatus", mock.Anything, mockTx, mock.Anything, domain.Session{
			ID:     "session_id",
			Status: domain.SessionStatusCompleted,
		}).Once().Return(nil)
		mockSessionRepo.On("GetByID", mock.Anything, mockTx, "session_id").Once().Return(domain.Session{
			ID:           "session_id",
			AssessmentID: "assessment_id",
			UserID:       "user_id",
			Status:       domain.SessionStatusCompleted,
		}, nil)
		mockAssessmentRepo.On("GetVirtualByID", mock.Anything, mockTx, "assessment_id").Once().Return(domain.Assessment{
			ID:                   "assessment_id",
			CourseID:             "course_id",
			LearningMaterialID:   "learning_material_id",
			LearningMaterialType: constants.LearningObjective,
			ManualGrading:        true,
		}, nil)
		mockSubmissionRepo.On("Insert", mock.Anything, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
		mockTx.On("Commit", mock.Anything).Return(nil)

		// actual
		err := usecase.CompleteAssessmentSession(ctx, "session_id")

		// assert
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}
