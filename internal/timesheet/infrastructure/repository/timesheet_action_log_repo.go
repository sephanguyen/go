package repository

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
)

type TimesheetActionLogRepoImpl struct{}

func (r *TimesheetActionLogRepoImpl) Create(ctx context.Context, db database.QueryExecer, log *entity.TimesheetActionLog) error {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetActionLogRepo.Create")
	defer span.End()

	if err := log.PreInsert(); err != nil {
		return fmt.Errorf("err PreInsert: %w", err)
	}

	cmdTag, err := database.Insert(ctx, log, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert TimesheetActionLog: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert TimesheetActionLog: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}
