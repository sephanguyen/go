package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"

	"go.uber.org/multierr"
)

type LeavingReasonRepo struct {
}

// Create creates leaving reason entity
func (r *LeavingReasonRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.LeavingReason) error {
	ctx, span := interceptors.StartSpan(ctx, "LeavingReasonRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		e.LeavingReasonID.Set(idutil.ULIDNow()),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert LeavingReason: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert LeavingReason: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

// Update updates LeavingReason entity
func (r *LeavingReasonRepo) Update(ctx context.Context, db database.QueryExecer, e *entities.LeavingReason) error {
	ctx, span := interceptors.StartSpan(ctx, "LeavingReasonRepo.Update")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, e, db.Exec, "leaving_reason_id", []string{"name", "leaving_reason_type", "remark", "is_archived", "updated_at"})

	if err != nil {
		return fmt.Errorf("err update LeavingReason: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update LeavingReason: %d RowsAffected", cmdTag.RowsAffected())
	}
	return nil
}
