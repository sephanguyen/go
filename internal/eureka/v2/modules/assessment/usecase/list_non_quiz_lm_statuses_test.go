package usecase

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_postgres "github.com/manabie-com/backend/mock/eureka/v2/modules/assessment/repository/postgres"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAssessmentUsecaseImpl_ListNonQuizLearningMaterialStatuses(t *testing.T) {
	t.Parallel()

	courseID := idutil.ULIDNow()
	userID := idutil.ULIDNow()
	eventTypes := []string{"study_guide_finished", "video_finished"}
	lmIDs := []string{"LM1", "LM2"}

	t.Run("Returns all completed status when all event types are exist for each LM", func(t *testing.T) {
		t.Parallel()
		// arrange
		mockDB := testutil.NewMockDB()
		ctx := context.Background()
		repo := &mock_postgres.MockStudentEventLogRepo{}
		sut := &AssessmentUsecaseImpl{StudentEventLogRepo: repo, DB: mockDB.DB}
		events := []domain.StudentEventLog{
			{
				EventType:          "study_guide_finished",
				LearningMaterialID: "LM1",
			},
			{
				EventType:          "video_finished",
				LearningMaterialID: "LM1",
			},
			{
				EventType:          "study_guide_finished",
				LearningMaterialID: "LM2",
			},
			{
				EventType:          "video_finished",
				LearningMaterialID: "LM2",
			},
		}
		repo.On("GetManyByEventTypesAndLMs", mock.Anything, mockDB.DB, courseID, userID, eventTypes, lmIDs).
			Once().Return(events, nil)
		expectedStatus := map[string]bool{
			"LM1": true,
			"LM2": true,
		}

		// act
		actual, err := sut.ListNonQuizLearningMaterialStatuses(ctx, courseID, userID, lmIDs)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, expectedStatus, actual)
		mock.AssertExpectationsForObjects(t, repo)
	})

	t.Run("Returns mixed completed and not completed status when not all LM have enough event types", func(t *testing.T) {
		t.Parallel()
		// arrange
		mockDB := testutil.NewMockDB()
		ctx := context.Background()
		repo := &mock_postgres.MockStudentEventLogRepo{}
		sut := &AssessmentUsecaseImpl{StudentEventLogRepo: repo, DB: mockDB.DB}
		events := []domain.StudentEventLog{
			{
				EventID:            idutil.ULIDNow(),
				EventType:          "study_guide_finished",
				LearningMaterialID: "LM1",
			},
			{
				EventID:            idutil.ULIDNow(),
				EventType:          "study_guide_finished",
				LearningMaterialID: "LM1",
			},
			{
				EventID:            idutil.ULIDNow(),
				EventType:          "video_finished",
				LearningMaterialID: "LM1",
			},
			{
				EventID:            idutil.ULIDNow(),
				EventType:          "video_finished",
				LearningMaterialID: "LM2",
			},
		}
		repo.On("GetManyByEventTypesAndLMs", mock.Anything, mockDB.DB, courseID, userID, eventTypes, lmIDs).
			Once().Return(events, nil)
		expectedStatus := map[string]bool{
			"LM1": true,
			"LM2": false,
		}

		// act
		actual, err := sut.ListNonQuizLearningMaterialStatuses(ctx, courseID, userID, lmIDs)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, expectedStatus, actual)
		mock.AssertExpectationsForObjects(t, repo)
	})

	t.Run("Returns all uncompleted status when not any LM have enough event types", func(t *testing.T) {
		t.Parallel()
		// arrange
		mockDB := testutil.NewMockDB()
		ctx := context.Background()
		repo := &mock_postgres.MockStudentEventLogRepo{}
		sut := &AssessmentUsecaseImpl{StudentEventLogRepo: repo, DB: mockDB.DB}
		events := []domain.StudentEventLog{
			{
				EventID:            idutil.ULIDNow(),
				EventType:          "study_guide_finished",
				LearningMaterialID: "LM1",
			},
			{
				EventID:            idutil.ULIDNow(),
				EventType:          "video_finished",
				LearningMaterialID: "LM2",
			},
			{
				EventID:            idutil.ULIDNow(),
				EventType:          "video_finished",
				LearningMaterialID: "LM2",
			},
		}
		repo.On("GetManyByEventTypesAndLMs", mock.Anything, mockDB.DB, courseID, userID, eventTypes, lmIDs).
			Once().Return(events, nil)
		expectedStatus := map[string]bool{
			"LM1": false,
			"LM2": false,
		}

		// act
		actual, err := sut.ListNonQuizLearningMaterialStatuses(ctx, courseID, userID, lmIDs)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, expectedStatus, actual)
		mock.AssertExpectationsForObjects(t, repo)
	})

	t.Run("Returns nil status and error when repo occurred an error", func(t *testing.T) {
		t.Parallel()
		// arrange
		mockDB := testutil.NewMockDB()
		ctx := context.Background()
		repo := &mock_postgres.MockStudentEventLogRepo{}
		sut := &AssessmentUsecaseImpl{StudentEventLogRepo: repo, DB: mockDB.DB}
		repoErr := errors.NewDBError("DB err", nil)
		repo.On("GetManyByEventTypesAndLMs", mock.Anything, mockDB.DB, courseID, userID, eventTypes, lmIDs).
			Once().Return(nil, repoErr)
		expectedErr := errors.New("AssessmentUsecase.ListNonQuizLearningMaterialStatuses", repoErr)

		// act
		actual, err := sut.ListNonQuizLearningMaterialStatuses(ctx, courseID, userID, lmIDs)

		// assert
		assert.Nil(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, repo)
	})
}
