package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"go.uber.org/multierr"
)

type EntryExitQueueRepo struct {
}

// Create entryexit_queue entity
func (r *EntryExitQueueRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.EntryExitQueue) error {
	ctx, span := interceptors.StartSpan(ctx, "EntryExitQueueRepo.Create")
	defer span.End()

	now := time.Now()
	err := multierr.Combine(
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	)
	if err != nil {
		return fmt.Errorf("err set EntryExitQueueRepo: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert EntryExitQueueRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert EntryExitQueueRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}
