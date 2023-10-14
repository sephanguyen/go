package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"

	"go.uber.org/multierr"
)

type StudentPaymentDetailActionLogRepo struct {
}

func (r *StudentPaymentDetailActionLogRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.StudentPaymentDetailActionLog) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentPaymentDetailActionLogRepo.Create")
	defer span.End()

	now := time.Now()

	if err := multierr.Combine(
		e.StudentPaymentDetailActionID.Set(idutil.ULIDNow()),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine StudentPaymentDetailActionID.Set,CreatedAt.Set,UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert StudentPaymentDetailActionLogRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert StudentPaymentDetailActionLogRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}
