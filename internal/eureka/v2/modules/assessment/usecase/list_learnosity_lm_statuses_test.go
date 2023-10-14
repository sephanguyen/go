package usecase

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/learnosity"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	mock_asm_learnosity_repo "github.com/manabie-com/backend/mock/eureka/v2/modules/assessment/repository/learnosity"
	mock_postgres "github.com/manabie-com/backend/mock/eureka/v2/modules/assessment/repository/postgres"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAssessmentUsecaseImpl_ListLearnositySessionStatuses(t *testing.T) {
	t.Parallel()

	courseID := idutil.ULIDNow()
	userID := idutil.ULIDNow()
	lmIDs := []string{"LM1", "LM2", "LM3", "LM4"}
	dbAssessments := sliceutils.Map(lmIDs, func(l string) domain.Assessment {
		return domain.Assessment{ID: idutil.ULIDNow(), CourseID: courseID, LearningMaterialID: l}
	})
	asmIDs := sliceutils.Map(dbAssessments, func(a domain.Assessment) string {
		return a.ID
	})
	dataRequest := learnosity.Request{
		"activity_id": asmIDs,
		"user_id":     []string{userID},
		"status":      []string{string(learnosity.SessionStatusCompleted)},
	}

	t.Run("returns all completed statuses when learnosity all session are completed", func(t *testing.T) {
		// Arrange
		mockDB := testutil.NewMockDB()
		ctx := context.Background()
		asmRepo := &mock_postgres.MockAssessmentRepo{}
		mockSessionRepo := &mock_asm_learnosity_repo.MockSessionRepo{}
		handler := &AssessmentUsecaseImpl{
			DB:                    mockDB.DB,
			AssessmentRepo:        asmRepo,
			LearnositySessionRepo: mockSessionRepo,
		}
		asmRepo.On("GetManyByLMAndCourseIDs", ctx, mockDB.DB, mock.Anything).
			Once().
			Return(dbAssessments, nil)
		learnositySessions := []domain.Session{
			{
				ID:           "Session1",
				AssessmentID: dbAssessments[0].ID,
				UserID:       userID,
			},
			{
				ID:           "Session2",
				AssessmentID: dbAssessments[1].ID,
				UserID:       userID,
			},
			{
				ID:           "Session3",
				AssessmentID: dbAssessments[2].ID,
				UserID:       userID,
			},
			{
				ID:           "Session4",
				AssessmentID: dbAssessments[3].ID,
				UserID:       userID,
			},
		}
		expectedStatus := map[string]bool{
			"LM1": true, "LM2": true, "LM3": true, "LM4": true,
		}
		mockSessionRepo.On("GetSessionStatuses", ctx, mock.Anything, dataRequest).
			Once().
			Return(learnositySessions, nil)

		// Act
		actual, err := handler.ListLearnositySessionStatuses(ctx, courseID, userID, lmIDs)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, expectedStatus, actual)
		mock.AssertExpectationsForObjects(t, mockSessionRepo)
	})

	t.Run("return completed when learnosity return more than one completed sessions per each assessment", func(t *testing.T) {
		// Arrange
		mockDB := testutil.NewMockDB()
		ctx := context.Background()
		asmRepo := &mock_postgres.MockAssessmentRepo{}
		mockSessionRepo := &mock_asm_learnosity_repo.MockSessionRepo{}
		handler := &AssessmentUsecaseImpl{
			DB:                    mockDB.DB,
			AssessmentRepo:        asmRepo,
			LearnositySessionRepo: mockSessionRepo,
		}
		asmRepo.On("GetManyByLMAndCourseIDs", ctx, mockDB.DB, mock.Anything).
			Once().
			Return(dbAssessments, nil)
		learnositySessions := []domain.Session{
			{
				ID:           "Session1",
				AssessmentID: dbAssessments[0].ID,
				UserID:       userID,
			},
			{
				ID:           "Session1-1",
				AssessmentID: dbAssessments[0].ID,
				UserID:       userID,
			},
			{
				ID:           "Session2",
				AssessmentID: dbAssessments[1].ID,
				UserID:       userID,
			},
			{
				ID:           "Session3",
				AssessmentID: dbAssessments[2].ID,
				UserID:       userID,
			},
			{
				ID:           "Session3-2",
				AssessmentID: dbAssessments[2].ID,
				UserID:       userID,
			},
			{
				ID:           "Session3-3",
				AssessmentID: dbAssessments[2].ID,
				UserID:       userID,
			},
			{
				ID:           "Session4",
				AssessmentID: dbAssessments[3].ID,
				UserID:       userID,
			},
		}
		expectedStatus := map[string]bool{
			"LM1": true, "LM2": true, "LM3": true, "LM4": true,
		}
		mockSessionRepo.On("GetSessionStatuses", ctx, mock.Anything, dataRequest).
			Once().
			Return(learnositySessions, nil)

		// Act
		actual, err := handler.ListLearnositySessionStatuses(ctx, courseID, userID, lmIDs)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, expectedStatus, actual)
		mock.AssertExpectationsForObjects(t, mockSessionRepo)
	})

	t.Run("return uncompleted status when learnosity response contains uncompleted sessions", func(t *testing.T) {
		// Arrange
		mockDB := testutil.NewMockDB()
		ctx := context.Background()
		asmRepo := &mock_postgres.MockAssessmentRepo{}
		mockSessionRepo := &mock_asm_learnosity_repo.MockSessionRepo{}
		handler := &AssessmentUsecaseImpl{
			DB:                    mockDB.DB,
			AssessmentRepo:        asmRepo,
			LearnositySessionRepo: mockSessionRepo,
		}
		asmRepo.On("GetManyByLMAndCourseIDs", ctx, mockDB.DB, mock.Anything).
			Once().
			Return(dbAssessments, nil)
		learnositySessions := []domain.Session{
			{
				ID:           "Session1",
				AssessmentID: dbAssessments[0].ID,
				UserID:       userID,
			},
			{
				ID:           "Session1-1",
				AssessmentID: dbAssessments[0].ID,
				UserID:       userID,
			},
			{
				ID:           "Session2",
				AssessmentID: dbAssessments[1].ID,
				UserID:       userID,
			},
		}
		expectedStatus := map[string]bool{
			"LM1": true, "LM2": true, "LM3": false, "LM4": false,
		}
		mockSessionRepo.On("GetSessionStatuses", ctx, mock.Anything, dataRequest).
			Once().
			Return(learnositySessions, nil)

		// Act
		actual, err := handler.ListLearnositySessionStatuses(ctx, courseID, userID, lmIDs)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, expectedStatus, actual)
		mock.AssertExpectationsForObjects(t, mockSessionRepo)
	})

	t.Run("return error when learnosity repo return error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockDB := testutil.NewMockDB()
		mockSessionRepo := &mock_asm_learnosity_repo.MockSessionRepo{}
		asmRepo := &mock_postgres.MockAssessmentRepo{}
		handler := &AssessmentUsecaseImpl{
			DB:                    mockDB.DB,
			AssessmentRepo:        asmRepo,
			LearnositySessionRepo: mockSessionRepo,
		}
		asmRepo.On("GetManyByLMAndCourseIDs", ctx, mockDB.DB, mock.Anything).
			Once().
			Return(dbAssessments, nil)
		repoErr := errors.New("Test", fmt.Errorf("%s", "some thing"))
		expectedErr := errors.New("AssessmentUsecase.ListLearnositySessionStatuses", repoErr)
		mockSessionRepo.On("GetSessionStatuses", ctx, mock.Anything, dataRequest).
			Once().
			Return(nil, repoErr)

		// Act
		actual, err := handler.ListLearnositySessionStatuses(ctx, courseID, userID, lmIDs)

		// Assert
		assert.Nil(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockSessionRepo)
	})

	t.Run("return error when assessment repo return error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockDB := testutil.NewMockDB()
		mockSessionRepo := &mock_asm_learnosity_repo.MockSessionRepo{}
		asmRepo := &mock_postgres.MockAssessmentRepo{}
		handler := &AssessmentUsecaseImpl{
			DB:                    mockDB.DB,
			AssessmentRepo:        asmRepo,
			LearnositySessionRepo: mockSessionRepo,
		}
		repoErr := errors.New("Test", fmt.Errorf("%s", "some thing"))
		expectedErr := errors.New("AssessmentUsecase.ListLearnositySessionStatuses", repoErr)
		asmRepo.On("GetManyByLMAndCourseIDs", ctx, mockDB.DB, mock.Anything).
			Once().
			Return(nil, repoErr)

		// Act
		actual, err := handler.ListLearnositySessionStatuses(ctx, courseID, userID, lmIDs)

		// Assert
		assert.Nil(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockSessionRepo)
		mock.AssertExpectationsForObjects(t, asmRepo)
	})
}
