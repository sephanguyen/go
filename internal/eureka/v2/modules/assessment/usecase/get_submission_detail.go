package usecase

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/google/uuid"
)

func (a *AssessmentUsecaseImpl) GetAssessmentSubmissionDetail(ctx context.Context, id string) (sub *domain.Submission, err error) {
	sub, err = a.SubmissionRepo.GetOneBySubmissionID(ctx, a.DB, id)
	if err != nil {
		if errors.CheckErrType(errors.ErrNoRowsExisted, err) {
			return sub, errors.NewEntityNotFoundError("AssessmentUsecase.GetAssessmentSubmissionDetail", err)
		}
		return sub, errors.New("AssessmentUsecase.GetAssessmentSubmissionDetail", err)
	}

	feedback, err := a.FeedbackSessionRepo.GetOneBySubmissionID(ctx, a.DB, id)
	if err != nil && !errors.CheckErrType(errors.ErrNoRowsExisted, err) {
		return nil, errors.New("AssessmentUsecase.GetAssessmentSubmissionDetail", err)
	}
	if errors.CheckErrType(errors.ErrNoRowsExisted, err) {
		requesterID := interceptors.UserIDFromContext(ctx)
		feedID, err := uuid.NewRandom()
		if err != nil {
			return nil, errors.New("AssessmentUsecase.GetAssessmentSubmissionDetail: Cant not generate feedback id", err)
		}

		feedback = &domain.FeedbackSession{
			ID:           feedID.String(),
			SubmissionID: id,
			CreatedBy:    requesterID,
			CreatedAt:    time.Now(),
		}
		err = a.FeedbackSessionRepo.Insert(ctx, a.DB, *feedback)
		if err != nil {
			return nil, errors.New("AssessmentUsecase.GetAssessmentSubmissionDetail: Failed to insert new feedback", err)
		}
	}

	if feedback != nil {
		sub.FeedBackBy = feedback.CreatedBy
		sub.FeedBackSessionID = feedback.ID
	}

	return sub, nil
}
