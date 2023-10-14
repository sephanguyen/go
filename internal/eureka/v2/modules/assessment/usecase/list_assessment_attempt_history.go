package usecase

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
)

// ListAssessmentAttemptHistory Merging uncompleted sessions and submissions
// If contains submission then include it
func (a *AssessmentUsecaseImpl) ListAssessmentAttemptHistory(ctx context.Context, userID, courseID, lmID string) ([]domain.Session, error) {
	assessment, err := a.AssessmentRepo.GetOneByLMAndCourseID(ctx, a.DB, courseID, lmID)
	if err != nil && !errors.CheckErrType(errors.ErrNoRowsExisted, err) {
		return nil, errors.New("AssessmentUsecase.ListAssessmentAttemptHistory", err)
	}
	sessions, err := a.AssessmentSessionRepo.GetManyByAssessments(ctx, a.DB, assessment.ID, userID)
	if err != nil {
		return nil, errors.New("AssessmentUsecase.ListAssessmentAttemptHistory", err)
	}
	if len(sessions) == 0 {
		return []domain.Session{}, nil
	}
	sessionMap := make(map[string]domain.Session, len(sessions))
	for _, v := range sessions {
		sessionMap[v.ID] = v
	}

	submissions, err := a.SubmissionRepo.GetManyByAssessments(ctx, a.DB, userID, assessment.ID)
	if err != nil {
		return nil, errors.New("AssessmentUsecase.ListAssessmentAttemptHistory", err)
	}
	submissionMap := make(map[string]domain.Submission, len(submissions))
	for _, v := range submissions {
		submissionMap[v.ID] = v
	}
	subIDs := sliceutils.Map(submissions, func(s domain.Submission) string {
		return s.ID
	})

	feedbacks, err := a.FeedbackSessionRepo.GetManyBySubmissionIDs(ctx, a.DB, subIDs)
	if err != nil {
		return nil, errors.New("AssessmentUsecase.ListAssessmentAttemptHistory", err)
	}

	for _, v := range feedbacks {
		sub, ok := submissionMap[v.SubmissionID]
		if ok {
			sub.FeedBackBy = v.CreatedBy
			sub.FeedBackSessionID = v.ID
			submissionMap[v.SubmissionID] = sub
		}
	}
	feedbackSubs := sliceutils.MapValuesToSlice(submissionMap)

	for _, v := range feedbackSubs {
		s, ok := sessionMap[v.SessionID]
		if ok {
			v := v
			s.Submission = &v
			sessionMap[v.SessionID] = s
		}
	}
	var attemptHistories domain.Sessions = sliceutils.MapValuesToSlice(sessionMap)

	attemptHistories.SortDescByCompletedAt()

	return attemptHistories, nil
}
