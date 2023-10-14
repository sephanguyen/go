package usecase

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/helper"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/learnosity"

	"github.com/jackc/pgx/v4"
)

func (a *AssessmentUsecaseImpl) CompleteAssessmentSession(ctx context.Context, sessionID string) error {
	now := time.Now()
	dataSecurity := helper.NewLearnositySecurity(ctx, a.LearnosityConfig, "localhost", now)

	dataRequest := learnosity.Request{
		"session_id": []string{sessionID},
	}

	sessionsLRN, err := a.LearnositySessionRepo.GetSessionResponses(ctx, dataSecurity, dataRequest)
	if err != nil {
		return errors.New("LearnositySessionRepo.GetSessionResponses", err)
	}

	if len(sessionsLRN) > 0 && sessionsLRN[0].Status != domain.SessionStatusCompleted {
		return errors.New("the session status is not completed", err)
	}

	if err = database.ExecInTx(ctx, a.DB, func(ctx context.Context, tx pgx.Tx) error {
		if err = a.AssessmentSessionRepo.UpdateStatus(ctx, tx, time.Now(), domain.Session{
			ID:     sessionID,
			Status: domain.SessionStatusCompleted,
		}); err != nil {
			return errors.New("AssessmentSessionRepo.UpdateStatus", err)
		}

		session, err := a.AssessmentSessionRepo.GetByID(ctx, tx, sessionID)
		if err != nil {
			return errors.New("AssessmentSessionRepo.GetByID", err)
		}

		assessment, err := a.AssessmentRepo.GetVirtualByID(ctx, tx, session.AssessmentID)
		if err != nil {
			return errors.New("AssessmentRepo.GetVirtualByID", err)
		}

		gradingStatus := domain.GradingStatusReturned
		if assessment.ManualGrading {
			gradingStatus = domain.GradingStatusNotMarked
		}
		submission := domain.Submission{
			ID:            idutil.ULIDNow(),
			SessionID:     session.ID,
			AssessmentID:  session.AssessmentID,
			StudentID:     session.UserID,
			GradingStatus: gradingStatus,
			MaxScore:      sessionsLRN[0].MaxScore,
			GradedScore:   sessionsLRN[0].GradedScore,
			CompletedAt:   *sessionsLRN[0].CompletedAt,
		}
		err = submission.Validate()
		if err != nil {
			return errors.New("domain.ValidateSubmission", err)
		}

		if err = a.SubmissionRepo.Insert(ctx, tx, now, submission); err != nil {
			return errors.New("SubmissionRepo.Insert", err)
		}

		return nil
	}); err != nil {
		return errors.New("database.ExecInTx", err)
	}

	return nil
}
