package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"

	"go.uber.org/multierr"
)

type OrderLeavingReasonRepo struct {
}

// Create creates leaving reason entity
func (r *OrderLeavingReasonRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.OrderLeavingReason) error {
	ctx, span := interceptors.StartSpan(ctx, "OrderLeavingReasonRepo.Create")
	defer span.End()
	now := time.Now()
	if err := multierr.Combine(
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert OrderLeavingReason: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert OrderLeavingReason: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}
