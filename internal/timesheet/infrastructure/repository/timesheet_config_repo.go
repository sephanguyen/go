package repository

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
)

type TimesheetConfigRepoImpl struct{}

func (r *TimesheetConfigRepoImpl) Create(ctx context.Context, db database.QueryExecer, config *entity.TimesheetConfig) error {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetConfigRepoImpl.Create")
	defer span.End()

	if err := config.PreInsert(); err != nil {
		return fmt.Errorf("err PreInsert: %w", err)
	}

	cmdTag, err := database.Insert(ctx, config, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert TimesheetConfig: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert TimesheetConfig: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *TimesheetConfigRepoImpl) Update(ctx context.Context, db database.QueryExecer, config *entity.TimesheetConfig) error {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetConfigRepoImpl.Update")
	defer span.End()

	if err := config.PreUpdate(); err != nil {
		return fmt.Errorf("err PreUpdate: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, config, db.Exec, config.PrimaryField(), []string{"config_type", "config_value", "is_archived", "updated_at"})
	if err != nil {
		return fmt.Errorf("err update TimesheetConfig: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update TimesheetConfig: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}
