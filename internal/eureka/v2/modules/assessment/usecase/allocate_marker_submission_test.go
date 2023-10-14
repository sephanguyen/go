package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	assessment_mock_postgres "github.com/manabie-com/backend/mock/eureka/v2/modules/assessment/repository/postgres"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAssessmentUsecaseImpl_AllocateMarkerSubmissions(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	submissions := []domain.Submission{{
		ID:                "Submission_1",
		AllocatedMarkerID: "AllocatedMarkerID_1",
	},
		{
			ID:                "Submission_1",
			AllocatedMarkerID: "",
		},
	}

	t.Run("happy case: repo.UpdateAllocateMarkerSubmissions success", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		mockTx := &mock_database.Tx{}
		mockRepo := &assessment_mock_postgres.MockSubmissionRepo{}
		usecase := &AssessmentUsecaseImpl{
			DB:             mockDB.DB,
			SubmissionRepo: mockRepo,
		}

		mockDB.DB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
		mockTx.On("Commit", mock.Anything).Return(nil)

		mockRepo.On("UpdateAllocateMarkerSubmissions", mock.Anything, mockDB.DB, submissions).Once().Return(nil)

		// actual
		err := usecase.AllocateMarkerSubmissions(ctx, submissions)

		// assert
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("error on SubmissionRepo.UpdateAllocateMarkerSubmissions", func(t *testing.T) {
		// arrange
		mockTx := &mock_database.Tx{}
		mockDB := testutil.NewMockDB()
		mockRepo := &assessment_mock_postgres.MockSubmissionRepo{}
		usecase := &AssessmentUsecaseImpl{
			DB:             mockDB.DB,
			SubmissionRepo: mockRepo,
		}

		repoErr := fmt.Errorf("UpdateAllocateMarkerSubmissions error")
		expectedErr := errors.New("SubmissionRepo.UpdateAllocateMarkerSubmissions", repoErr)

		mockDB.DB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
		mockTx.On("Rollback", mock.Anything).Return(nil)
		mockRepo.On("UpdateAllocateMarkerSubmissions", mock.Anything, mockDB.DB, submissions).
			Once().Return(repoErr)

		// actual
		err := usecase.AllocateMarkerSubmissions(ctx, submissions)

		// assert
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}
