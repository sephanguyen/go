package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/constants"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	assessment_mock_postgres "github.com/manabie-com/backend/mock/eureka/v2/modules/assessment/repository/postgres"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAssessmentUsecaseImpl_ListAssessmentAttemptHistory(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	userID, courseID, lmID := idutil.ULIDNow(), idutil.ULIDNow(), idutil.ULIDNow()

	t.Run("return error when AssessmentRepo.GetOneByLMAndCourseID returns error", func(t *testing.T) {
		// arrange
		assessmentRepo := &assessment_mock_postgres.MockAssessmentRepo{}
		submissionRepo := &assessment_mock_postgres.MockSubmissionRepo{}
		mockDB := testutil.NewMockDB()
		sut := &AssessmentUsecaseImpl{
			AssessmentRepo: assessmentRepo,
			SubmissionRepo: submissionRepo,
			DB:             mockDB.DB,
		}
		repoErr := errors.NewDBError("db err", nil)
		expectedErr := errors.New("AssessmentUsecase.ListAssessmentAttemptHistory", repoErr)
		assessmentRepo.On("GetOneByLMAndCourseID", ctx, mockDB.DB, courseID, lmID).Once().Return(nil, repoErr)

		// act
		actual, err := sut.ListAssessmentAttemptHistory(ctx, userID, courseID, lmID)

		// assert
		assert.Nil(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, assessmentRepo)
	})

	t.Run("return error when AssessmentSessionRepo.GetManyByAssessments returns error", func(t *testing.T) {
		// arrange
		assessmentRepo := &assessment_mock_postgres.MockAssessmentRepo{}
		sessionRepo := &assessment_mock_postgres.MockAssessmentSessionRepo{}
		feedbackRepo := &assessment_mock_postgres.MockFeedbackSessionRepo{}
		submissionRepo := &assessment_mock_postgres.MockSubmissionRepo{}
		mockDB := testutil.NewMockDB()
		sut := &AssessmentUsecaseImpl{
			AssessmentRepo:        assessmentRepo,
			FeedbackSessionRepo:   feedbackRepo,
			SubmissionRepo:        submissionRepo,
			AssessmentSessionRepo: sessionRepo,
			DB:                    mockDB.DB,
		}
		assessment := &domain.Assessment{
			ID:                   idutil.ULIDNow(),
			CourseID:             courseID,
			LearningMaterialID:   lmID,
			LearningMaterialType: constants.LearningObjective,
		}
		assessmentRepo.On("GetOneByLMAndCourseID", ctx, mockDB.DB, courseID, lmID).Once().Return(assessment, nil)

		repoErr := errors.NewDBError("db err 2", nil)
		expectedErr := errors.New("AssessmentUsecase.ListAssessmentAttemptHistory", repoErr)
		sessionRepo.On("GetManyByAssessments", ctx, mockDB.DB, assessment.ID, userID).
			Once().
			Return(nil, repoErr)

		// act
		actual, err := sut.ListAssessmentAttemptHistory(ctx, userID, courseID, lmID)

		// assert
		assert.Nil(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, assessmentRepo, sessionRepo)
	})

	t.Run("return empty slice when does not find any sessions by AssessmentSessionRepo.GetManyByAssessments", func(t *testing.T) {
		// arrange
		assessmentRepo := &assessment_mock_postgres.MockAssessmentRepo{}
		sessionRepo := &assessment_mock_postgres.MockAssessmentSessionRepo{}
		mockDB := testutil.NewMockDB()
		sut := &AssessmentUsecaseImpl{
			AssessmentRepo:        assessmentRepo,
			AssessmentSessionRepo: sessionRepo,
			DB:                    mockDB.DB,
		}
		assessment := &domain.Assessment{
			ID:                   idutil.ULIDNow(),
			CourseID:             courseID,
			LearningMaterialID:   lmID,
			LearningMaterialType: constants.LearningObjective,
		}
		assessmentRepo.On("GetOneByLMAndCourseID", ctx, mockDB.DB, courseID, lmID).Once().Return(assessment, nil)
		sessionRepo.On("GetManyByAssessments", ctx, mockDB.DB, assessment.ID, userID).
			Once().
			Return(nil, nil)

		// act
		actual, err := sut.ListAssessmentAttemptHistory(ctx, userID, courseID, lmID)

		// assert
		assert.Equal(t, []domain.Session{}, actual)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, assessmentRepo, sessionRepo)
	})

	t.Run("return error when SubmissionRepo.GetManyByAssessments returns error", func(t *testing.T) {
		// arrange
		assessmentRepo := &assessment_mock_postgres.MockAssessmentRepo{}
		feedbackRepo := &assessment_mock_postgres.MockFeedbackSessionRepo{}
		sessionRepo := &assessment_mock_postgres.MockAssessmentSessionRepo{}
		submissionRepo := &assessment_mock_postgres.MockSubmissionRepo{}
		mockDB := testutil.NewMockDB()
		sut := &AssessmentUsecaseImpl{
			AssessmentRepo:        assessmentRepo,
			FeedbackSessionRepo:   feedbackRepo,
			AssessmentSessionRepo: sessionRepo,
			SubmissionRepo:        submissionRepo,
			DB:                    mockDB.DB,
		}
		assessment := &domain.Assessment{
			ID:                   idutil.ULIDNow(),
			CourseID:             courseID,
			LearningMaterialID:   lmID,
			LearningMaterialType: constants.LearningObjective,
		}
		assessmentRepo.On("GetOneByLMAndCourseID", ctx, mockDB.DB, courseID, lmID).Once().Return(assessment, nil)

		repoErr := errors.NewDBError("db err 2", nil)
		expectedErr := errors.New("AssessmentUsecase.ListAssessmentAttemptHistory", repoErr)
		sessionRepo.On("GetManyByAssessments", ctx, mockDB.DB, assessment.ID, userID).
			Once().
			Return([]domain.Session{
				{ID: idutil.ULIDNow()},
			}, nil)
		submissionRepo.On("GetManyByAssessments", ctx, mockDB.DB, userID, assessment.ID).
			Once().
			Return(nil, repoErr)

		// act
		actual, err := sut.ListAssessmentAttemptHistory(ctx, userID, courseID, lmID)

		// assert
		assert.Nil(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, assessmentRepo, submissionRepo)
	})

	t.Run("return error when FeedbackSessionRepo.GetManyBySubmissionIDs returns error", func(t *testing.T) {
		// arrange
		assessmentRepo := &assessment_mock_postgres.MockAssessmentRepo{}
		feedbackRepo := &assessment_mock_postgres.MockFeedbackSessionRepo{}
		sessionRepo := &assessment_mock_postgres.MockAssessmentSessionRepo{}
		submissionRepo := &assessment_mock_postgres.MockSubmissionRepo{}
		mockDB := testutil.NewMockDB()
		sut := &AssessmentUsecaseImpl{
			AssessmentRepo:        assessmentRepo,
			FeedbackSessionRepo:   feedbackRepo,
			AssessmentSessionRepo: sessionRepo,
			SubmissionRepo:        submissionRepo,
			DB:                    mockDB.DB,
		}
		assessment := &domain.Assessment{
			ID:                   idutil.ULIDNow(),
			CourseID:             courseID,
			LearningMaterialID:   lmID,
			LearningMaterialType: constants.LearningObjective,
		}
		now := time.Now().Add(10 * time.Minute)
		submissions := []domain.Submission{
			{
				ID:                "sub 1",
				SessionID:         "ses 1",
				AssessmentID:      assessment.ID,
				StudentID:         userID,
				AllocatedMarkerID: idutil.ULIDNow(),
				GradingStatus:     domain.GradingStatusMarked,
				MaxScore:          10,
				GradedScore:       9,
				MarkedBy:          idutil.ULIDNow(),
				MarkedAt:          &now,
				FeedBackSessionID: idutil.ULIDNow(),
				FeedBackBy:        idutil.ULIDNow(),
				CreatedAt:         time.Now(),
				CompletedAt:       now,
			},
			{
				ID:                "sub 2",
				SessionID:         "ses 2",
				AssessmentID:      assessment.ID,
				StudentID:         userID,
				AllocatedMarkerID: idutil.ULIDNow(),
				GradingStatus:     domain.GradingStatusMarked,
				MaxScore:          10,
				GradedScore:       8,
				MarkedBy:          idutil.ULIDNow(),
				MarkedAt:          &now,
				FeedBackSessionID: idutil.ULIDNow(),
				FeedBackBy:        idutil.ULIDNow(),
				CreatedAt:         time.Now(),
				CompletedAt:       now,
			},
		}
		subIDs := []string{"sub 1", "sub 2"}

		assessmentRepo.On("GetOneByLMAndCourseID", ctx, mockDB.DB, courseID, lmID).Once().Return(assessment, nil)
		sessionRepo.On("GetManyByAssessments", ctx, mockDB.DB, assessment.ID, userID).
			Once().
			Return([]domain.Session{
				{ID: idutil.ULIDNow()},
			}, nil)
		submissionRepo.On("GetManyByAssessments", ctx, mockDB.DB, userID, assessment.ID).
			Once().
			Return(submissions, nil)
		repoErr := errors.NewDBError("db err 3", nil)
		expectedErr := errors.New("AssessmentUsecase.ListAssessmentAttemptHistory", repoErr)
		feedbackRepo.On("GetManyBySubmissionIDs", ctx, mockDB.DB, subIDs).Once().Return(nil, repoErr)

		// act
		actual, err := sut.ListAssessmentAttemptHistory(ctx, userID, courseID, lmID)

		// assert
		assert.Nil(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, assessmentRepo, sessionRepo, submissionRepo, feedbackRepo)
	})

	t.Run("return attempt histories sorted in completed descending with no feedback", func(t *testing.T) {
		// arrange
		assessmentRepo := &assessment_mock_postgres.MockAssessmentRepo{}
		feedbackRepo := &assessment_mock_postgres.MockFeedbackSessionRepo{}
		sessionRepo := &assessment_mock_postgres.MockAssessmentSessionRepo{}
		submissionRepo := &assessment_mock_postgres.MockSubmissionRepo{}
		mockDB := testutil.NewMockDB()
		sut := &AssessmentUsecaseImpl{
			AssessmentRepo:        assessmentRepo,
			FeedbackSessionRepo:   feedbackRepo,
			AssessmentSessionRepo: sessionRepo,
			SubmissionRepo:        submissionRepo,
			DB:                    mockDB.DB,
		}
		assessment := &domain.Assessment{
			ID:                   idutil.ULIDNow(),
			CourseID:             courseID,
			LearningMaterialID:   lmID,
			LearningMaterialType: constants.LearningObjective,
		}
		now1 := time.Now()
		now2 := now1.Add(5 * time.Minute)
		now3 := now1.Add(10 * time.Minute)
		sessions := []domain.Session{
			{
				ID:                 "ses 1",
				AssessmentID:       assessment.ID,
				UserID:             userID,
				CourseID:           courseID,
				LearningMaterialID: lmID,
				Submission:         nil,
				MaxScore:           20,
				GradedScore:        19,
				Status:             "NONE",
				CreatedAt:          now1,
				CompletedAt:        nil,
			},
			{
				ID:                 "ses 2",
				AssessmentID:       assessment.ID,
				UserID:             userID,
				CourseID:           courseID,
				LearningMaterialID: lmID,
				Submission:         nil,
				MaxScore:           20,
				GradedScore:        19,
				Status:             "COMPLETED",
				CreatedAt:          now2,
				CompletedAt:        &now3,
			},
		}
		submissions := []domain.Submission{
			{
				ID:                "sub 2",
				SessionID:         "ses 2",
				AssessmentID:      assessment.ID,
				StudentID:         userID,
				AllocatedMarkerID: idutil.ULIDNow(),
				GradingStatus:     domain.GradingStatusMarked,
				MaxScore:          10,
				GradedScore:       8,
				MarkedBy:          idutil.ULIDNow(),
				MarkedAt:          &now1,
				FeedBackSessionID: idutil.ULIDNow(),
				FeedBackBy:        idutil.ULIDNow(),
				CreatedAt:         now3,
				CompletedAt:       now3,
			},
		}
		expected := []domain.Session{
			{
				ID:                 "ses 2",
				AssessmentID:       assessment.ID,
				UserID:             userID,
				CourseID:           courseID,
				LearningMaterialID: lmID,
				Submission:         &submissions[0],
				MaxScore:           20,
				GradedScore:        19,
				Status:             "COMPLETED",
				CreatedAt:          now2,
				CompletedAt:        &now3,
			},
			{
				ID:                 "ses 1",
				AssessmentID:       assessment.ID,
				UserID:             userID,
				CourseID:           courseID,
				LearningMaterialID: lmID,
				Submission:         nil,
				MaxScore:           20,
				GradedScore:        19,
				Status:             "NONE",
				CreatedAt:          now1,
				CompletedAt:        nil,
			},
		}
		subIDs := []string{"sub 2"}
		feedbacks := make([]domain.FeedbackSession, 0)
		assessmentRepo.On("GetOneByLMAndCourseID", ctx, mockDB.DB, courseID, lmID).Once().Return(assessment, nil)
		sessionRepo.On("GetManyByAssessments", ctx, mockDB.DB, assessment.ID, userID).
			Once().
			Return(sessions, nil)
		submissionRepo.On("GetManyByAssessments", ctx, mockDB.DB, userID, assessment.ID).
			Once().
			Return(submissions, nil)
		feedbackRepo.On("GetManyBySubmissionIDs", ctx, mockDB.DB, subIDs).Once().Return(feedbacks, nil)

		// act
		actual, err := sut.ListAssessmentAttemptHistory(ctx, userID, courseID, lmID)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, expected, actual)
		mock.AssertExpectationsForObjects(t, assessmentRepo, submissionRepo, sessionRepo, feedbackRepo)
	})

	t.Run("return attempt histories sorted in completed descending with feedbacks", func(t *testing.T) {
		// arrange
		assessmentRepo := &assessment_mock_postgres.MockAssessmentRepo{}
		feedbackRepo := &assessment_mock_postgres.MockFeedbackSessionRepo{}
		sessionRepo := &assessment_mock_postgres.MockAssessmentSessionRepo{}
		submissionRepo := &assessment_mock_postgres.MockSubmissionRepo{}
		mockDB := testutil.NewMockDB()
		sut := &AssessmentUsecaseImpl{
			AssessmentRepo:        assessmentRepo,
			FeedbackSessionRepo:   feedbackRepo,
			AssessmentSessionRepo: sessionRepo,
			SubmissionRepo:        submissionRepo,
			DB:                    mockDB.DB,
		}
		assessment := &domain.Assessment{
			ID:                   idutil.ULIDNow(),
			CourseID:             courseID,
			LearningMaterialID:   lmID,
			LearningMaterialType: constants.LearningObjective,
		}
		now1 := time.Now()
		now2 := now1.Add(5 * time.Minute)
		now3 := now1.Add(10 * time.Minute)
		sessions := []domain.Session{
			{
				ID:                 "ses 2",
				AssessmentID:       assessment.ID,
				UserID:             userID,
				CourseID:           courseID,
				LearningMaterialID: lmID,
				Submission:         nil,
				MaxScore:           20,
				GradedScore:        19,
				Status:             "COMPLETED",
				CreatedAt:          now2,
				CompletedAt:        &now3,
			},
			{
				ID:                 "ses 1",
				AssessmentID:       assessment.ID,
				UserID:             userID,
				CourseID:           courseID,
				LearningMaterialID: lmID,
				Submission:         nil,
				MaxScore:           20,
				GradedScore:        19,
				Status:             "NONE",
				CreatedAt:          now1,
				CompletedAt:        nil,
			},
		}
		submissions := []domain.Submission{
			{
				ID:                "sub 2",
				SessionID:         "ses 2",
				AssessmentID:      assessment.ID,
				StudentID:         userID,
				AllocatedMarkerID: idutil.ULIDNow(),
				GradingStatus:     domain.GradingStatusMarked,
				MaxScore:          10,
				GradedScore:       8,
				MarkedBy:          idutil.ULIDNow(),
				MarkedAt:          &now1,
				FeedBackSessionID: idutil.ULIDNow(),
				FeedBackBy:        idutil.ULIDNow(),
				CreatedAt:         now3,
				CompletedAt:       now3,
			},
		}
		feedbacks := []domain.FeedbackSession{
			{
				ID:           "FEED ID 1",
				SubmissionID: "sub 2",
				CreatedBy:    "KIENN",
				CreatedAt:    time.Now(),
			},
		}
		expectedSubmission := submissions[0]
		expectedSubmission.FeedBackBy = feedbacks[0].CreatedBy
		expectedSubmission.FeedBackSessionID = feedbacks[0].ID
		expected := []domain.Session{
			{
				ID:                 "ses 2",
				AssessmentID:       assessment.ID,
				UserID:             userID,
				CourseID:           courseID,
				LearningMaterialID: lmID,
				Submission:         &expectedSubmission,
				MaxScore:           20,
				GradedScore:        19,
				Status:             "COMPLETED",
				CreatedAt:          now2,
				CompletedAt:        &now3,
			},
			{
				ID:                 "ses 1",
				AssessmentID:       assessment.ID,
				UserID:             userID,
				CourseID:           courseID,
				LearningMaterialID: lmID,
				Submission:         nil,
				MaxScore:           20,
				GradedScore:        19,
				Status:             "NONE",
				CreatedAt:          now1,
				CompletedAt:        nil,
			},
		}
		subIDs := []string{"sub 2"}
		assessmentRepo.On("GetOneByLMAndCourseID", ctx, mockDB.DB, courseID, lmID).Once().Return(assessment, nil)
		sessionRepo.On("GetManyByAssessments", ctx, mockDB.DB, assessment.ID, userID).
			Once().
			Return(sessions, nil)
		submissionRepo.On("GetManyByAssessments", ctx, mockDB.DB, userID, assessment.ID).
			Once().
			Return(submissions, nil)
		feedbackRepo.On("GetManyBySubmissionIDs", ctx, mockDB.DB, subIDs).Once().Return(feedbacks, nil)

		// act
		actual, err := sut.ListAssessmentAttemptHistory(ctx, userID, courseID, lmID)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, expected, actual)
		mock.AssertExpectationsForObjects(t, assessmentRepo, submissionRepo, sessionRepo, feedbackRepo)
	})
}
