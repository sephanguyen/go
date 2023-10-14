package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_postgres "github.com/manabie-com/backend/mock/eureka/v2/modules/assessment/repository/postgres"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAssessmentUsecaseImpl_GetAssessmentSubmissionDetail(t *testing.T) {
	t.Parallel()
	id := idutil.ULIDNow()
	parentCtx := context.Background()

	t.Run("Return Not found entity when repo doesnt find any submission", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		ctx := context.Background()
		repo := &mock_postgres.MockSubmissionRepo{}
		sut := &AssessmentUsecaseImpl{SubmissionRepo: repo, DB: mockDB.DB}
		notFoundErr := errors.NewNoRowsExistedError("SubmissionRepo.GetOneBySubmissionID", nil)
		repo.On("GetOneBySubmissionID", mock.Anything, mockDB.DB, id).
			Once().Return(nil, notFoundErr)
		expectedErr := errors.NewEntityNotFoundError("AssessmentUsecase.GetAssessmentSubmissionDetail", notFoundErr)

		// act
		actual, err := sut.GetAssessmentSubmissionDetail(ctx, id)

		// assert
		assert.Nil(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, repo)
	})

	t.Run("Return general error when submission repo occurred error", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		ctx := context.Background()
		repo := &mock_postgres.MockSubmissionRepo{}
		sut := &AssessmentUsecaseImpl{SubmissionRepo: repo, DB: mockDB.DB}
		dbError := errors.NewDBError("SubmissionRepo.GetOneBySubmissionID", nil)
		repo.On("GetOneBySubmissionID", mock.Anything, mockDB.DB, id).
			Once().Return(nil, dbError)
		expectedErr := errors.New("AssessmentUsecase.GetAssessmentSubmissionDetail", dbError)

		// act
		actual, err := sut.GetAssessmentSubmissionDetail(ctx, id)

		// assert
		assert.Nil(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, repo)
	})

	t.Run("Return general error when feedback repo return error different with no rows", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		ctx := context.Background()
		repo := &mock_postgres.MockSubmissionRepo{}
		feedbackRepo := &mock_postgres.MockFeedbackSessionRepo{}
		sut := &AssessmentUsecaseImpl{SubmissionRepo: repo, FeedbackSessionRepo: feedbackRepo, DB: mockDB.DB}
		sub := &domain.Submission{
			ID:                "Some ID",
			SessionID:         idutil.ULIDNow(),
			AssessmentID:      idutil.ULIDNow(),
			StudentID:         idutil.ULIDNow(),
			AllocatedMarkerID: idutil.ULIDNow(),
			GradingStatus:     domain.GradingStatusInProgress,
			MaxScore:          20,
			GradedScore:       10,
			MarkedBy:          idutil.ULIDNow(),
			MarkedAt:          nil,
			FeedBackSessionID: "",
			FeedBackBy:        "",
			CreatedAt:         time.Time{},
			CompletedAt:       time.Time{},
		}
		repo.On("GetOneBySubmissionID", mock.Anything, mockDB.DB, id).
			Once().Return(sub, nil)
		repoErr := errors.NewDBError("Repooo", nil)
		feedbackRepo.On("GetOneBySubmissionID", mock.Anything, mockDB.DB, id).
			Once().Return(nil, repoErr)
		expectedErr := errors.New("AssessmentUsecase.GetAssessmentSubmissionDetail", repoErr)

		// act
		actual, err := sut.GetAssessmentSubmissionDetail(ctx, id)

		// assert
		assert.Nil(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, repo, feedbackRepo)
	})

	t.Run("Return error when can not create feedback", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		ctx := context.WithValue(parentCtx, interceptors.UserIDKey(0), "USER_ID")

		repo := &mock_postgres.MockSubmissionRepo{}
		feedbackRepo := &mock_postgres.MockFeedbackSessionRepo{}
		sut := &AssessmentUsecaseImpl{SubmissionRepo: repo, FeedbackSessionRepo: feedbackRepo, DB: mockDB.DB}
		sub := &domain.Submission{
			ID:                id,
			SessionID:         idutil.ULIDNow(),
			AssessmentID:      idutil.ULIDNow(),
			StudentID:         idutil.ULIDNow(),
			AllocatedMarkerID: idutil.ULIDNow(),
			GradingStatus:     domain.GradingStatusInProgress,
			MaxScore:          20,
			GradedScore:       10,
			MarkedBy:          idutil.ULIDNow(),
			MarkedAt:          nil,
			FeedBackSessionID: "",
			FeedBackBy:        "",
			CreatedAt:         time.Time{},
			CompletedAt:       time.Time{},
		}
		repo.On("GetOneBySubmissionID", mock.Anything, mockDB.DB, id).
			Once().Return(sub, nil)
		feedbackRepo.On("GetOneBySubmissionID", mock.Anything, mockDB.DB, id).
			Once().Return(nil, errors.NewNoRowsExistedError("", nil))
		rootErr := errors.NewDBError("Some thing", nil)
		feedbackRepo.On("Insert", mock.Anything, mockDB.DB, mock.Anything).
			Once().
			Run(func(args mock.Arguments) {
				// assert
				feedback := args[2].(domain.FeedbackSession)
				assert.Equal(t, "USER_ID", feedback.CreatedBy)
				assert.Equal(t, sub.ID, feedback.SubmissionID)
				assert.NotEmpty(t, feedback.CreatedAt, feedback.ID)
				_, err := uuid.Parse(feedback.ID)
				assert.Nil(t, err)
			}).
			Return(rootErr)
		expectedErr := errors.New("AssessmentUsecase.GetAssessmentSubmissionDetail: Failed to insert new feedback", rootErr)

		// act
		actual, err := sut.GetAssessmentSubmissionDetail(ctx, id)

		// assert
		assert.Nil(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, repo, feedbackRepo)
	})

	t.Run("Return submission with newly created feedback when no feedback existed", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		ctx := context.WithValue(parentCtx, interceptors.UserIDKey(0), "USER_ID")

		repo := &mock_postgres.MockSubmissionRepo{}
		feedbackRepo := &mock_postgres.MockFeedbackSessionRepo{}
		sut := &AssessmentUsecaseImpl{SubmissionRepo: repo, FeedbackSessionRepo: feedbackRepo, DB: mockDB.DB}
		sub := &domain.Submission{
			ID:                id,
			SessionID:         idutil.ULIDNow(),
			AssessmentID:      idutil.ULIDNow(),
			StudentID:         idutil.ULIDNow(),
			AllocatedMarkerID: idutil.ULIDNow(),
			GradingStatus:     domain.GradingStatusInProgress,
			MaxScore:          20,
			GradedScore:       10,
			MarkedBy:          idutil.ULIDNow(),
			MarkedAt:          nil,
			FeedBackSessionID: "",
			FeedBackBy:        "",
			CreatedAt:         time.Time{},
			CompletedAt:       time.Time{},
		}
		repo.On("GetOneBySubmissionID", mock.Anything, mockDB.DB, id).
			Once().Return(sub, nil)
		feedbackRepo.On("GetOneBySubmissionID", mock.Anything, mockDB.DB, id).
			Once().Return(nil, errors.NewNoRowsExistedError("", nil))
		feedbackRepo.On("Insert", mock.Anything, mockDB.DB, mock.Anything).
			Once().
			Return(nil).
			Run(func(args mock.Arguments) {
				// assert
				feedback := args[2].(domain.FeedbackSession)
				assert.Equal(t, "USER_ID", feedback.CreatedBy)
				assert.Equal(t, sub.ID, feedback.SubmissionID)
				assert.NotEmpty(t, feedback.CreatedAt, feedback.ID)
				_, err := uuid.Parse(feedback.ID)
				assert.Nil(t, err)
			})

		// act
		actual, err := sut.GetAssessmentSubmissionDetail(ctx, id)

		// assert
		assert.Nil(t, err)
		actual.FeedBackSessionID = ""
		actual.FeedBackBy = ""
		assert.Equal(t, sub, actual)
		mock.AssertExpectationsForObjects(t, repo, feedbackRepo)
	})

	t.Run("Return submission from repo when there is a feedback", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		ctx := context.Background()
		repo := &mock_postgres.MockSubmissionRepo{}
		feedbackRepo := &mock_postgres.MockFeedbackSessionRepo{}
		sut := &AssessmentUsecaseImpl{SubmissionRepo: repo, FeedbackSessionRepo: feedbackRepo, DB: mockDB.DB}
		sub := &domain.Submission{
			ID:                "Some ID",
			SessionID:         idutil.ULIDNow(),
			AssessmentID:      idutil.ULIDNow(),
			StudentID:         idutil.ULIDNow(),
			AllocatedMarkerID: idutil.ULIDNow(),
			GradingStatus:     domain.GradingStatusInProgress,
			MaxScore:          20,
			GradedScore:       10,
			MarkedBy:          idutil.ULIDNow(),
			MarkedAt:          nil,
			FeedBackSessionID: "",
			FeedBackBy:        "",
			CreatedAt:         time.Time{},
			CompletedAt:       time.Time{},
		}
		fb := &domain.FeedbackSession{
			ID:           "F1",
			SubmissionID: "Some ID",
			CreatedBy:    "KIEN",
			CreatedAt:    time.Now(),
		}
		cloned := *sub
		cloned.FeedBackBy = "KIEN"
		cloned.FeedBackSessionID = "F1"
		expectedSub := &cloned
		repo.On("GetOneBySubmissionID", mock.Anything, mockDB.DB, id).
			Once().Return(sub, nil)
		feedbackRepo.On("GetOneBySubmissionID", mock.Anything, mockDB.DB, id).
			Once().
			Return(fb, nil)

		// act
		actual, err := sut.GetAssessmentSubmissionDetail(ctx, id)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, expectedSub, actual)
		mock.AssertExpectationsForObjects(t, repo, feedbackRepo)
	})
}
