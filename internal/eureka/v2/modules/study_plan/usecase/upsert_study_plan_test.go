package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	study_plan_mock_postgres "github.com/manabie-com/backend/mock/eureka/v2/modules/study_plan/repository/postgres"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStudyPlanUsecaseImpl_UpsertStudyPlan(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	sp := domain.StudyPlan{
		Name:         "name",
		CourseID:     "course_id",
		AcademicYear: "study_plan_id",
		Status:       domain.StudyPlanStatusActive,
	}

	t.Run("return error when StudyPlanRepo.Upsert returns error", func(t *testing.T) {
		studyPlanRepo := &study_plan_mock_postgres.MockStudyPlanRepo{}
		mockDB := testutil.NewMockDB()
		sut := &StudyPlanUsecaseImpl{
			DB:            mockDB.DB,
			StudyPlanRepo: studyPlanRepo,
		}
		repoErr := errors.NewDBError("db err", nil)
		expectedErr := errors.New("StudyPlanUsecase.StudyPlanRepo.Upsert", repoErr)
		studyPlanRepo.On("Upsert", ctx, mockDB.DB, mock.Anything, mock.Anything).Once().Return("", repoErr)

		// act
		actual, err := sut.UpsertStudyPlan(ctx, sp)

		// assert
		assert.Empty(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, studyPlanRepo)
	})

	t.Run("happy case", func(t *testing.T) {
		studyPlanRepo := &study_plan_mock_postgres.MockStudyPlanRepo{}
		mockDB := testutil.NewMockDB()
		sut := &StudyPlanUsecaseImpl{
			DB:            mockDB.DB,
			StudyPlanRepo: studyPlanRepo,
		}
		expectedId := "study_plan_id"
		studyPlanRepo.On("Upsert", ctx, mockDB.DB, mock.Anything, mock.Anything).Once().Return(expectedId, nil)

		// act
		actual, err := sut.UpsertStudyPlan(ctx, sp)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, expectedId, actual)
		mock.AssertExpectationsForObjects(t, studyPlanRepo)
	})
}
