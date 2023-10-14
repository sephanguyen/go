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

type AccountingCategoryRepo struct {
}

// Create creates AccountingCategory entity
func (r *AccountingCategoryRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.AccountingCategory) error {
	ctx, span := interceptors.StartSpan(ctx, "AccountingCategoryRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		e.AccountingCategoryID.Set(idutil.ULIDNow()),
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert AccountingCategory: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert AccountingCategory: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

// Update updates AccountingCategory entity
func (r *AccountingCategoryRepo) Update(ctx context.Context, db database.QueryExecer, e *entities.AccountingCategory) error {
	ctx, span := interceptors.StartSpan(ctx, "AccountingCategoryRepo.Update")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, e, db.Exec, "accounting_category_id", []string{"name", "remarks", "is_archived", "updated_at"})
	if err != nil {
		return fmt.Errorf("err update AccountingCategory: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update AccountingCategory: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}
