package usecase

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgx/v4"
)

func (a *AssessmentUsecaseImpl) AllocateMarkerSubmissions(ctx context.Context, submissions []domain.Submission) error {
	err := database.ExecInTx(ctx, a.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		if err := a.SubmissionRepo.UpdateAllocateMarkerSubmissions(ctx, a.DB, submissions); err != nil {
			return errors.New("SubmissionRepo.UpdateAllocateMarkerSubmissions", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
