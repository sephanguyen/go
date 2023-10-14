package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type AutoCreateFlagRepoImpl struct {
}

func (r *AutoCreateFlagRepoImpl) Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entity.AutoCreateTimesheetFlag, error) {
	ctx, span := interceptors.StartSpan(ctx, "AutoCreateFlagRepoImpl.Retrieve")
	defer span.End()

	autoCreateFlag := &entity.AutoCreateTimesheetFlag{}
	autoCreateFlags := &entity.AutoCreateTimesheetFlags{}

	fields, _ := autoCreateFlag.FieldMap()
	stmt := fmt.Sprintf(`SELECT %s
	FROM %s
	WHERE deleted_at IS NULL
	AND staff_id = ANY($1::_TEXT);`, strings.Join(fields, ", "), autoCreateFlag.TableName())

	if err := database.Select(ctx, db, stmt, ids).ScanAll(autoCreateFlags); err != nil {
		return nil, err
	}

	return *autoCreateFlags, nil
}

func (r *AutoCreateFlagRepoImpl) FindAutoCreatedFlagByStaffID(ctx context.Context, db database.QueryExecer, staffID pgtype.Text) (*entity.AutoCreateTimesheetFlag, error) {
	ctx, span := interceptors.StartSpan(ctx, "AutoCreateFlagRepoImpl.FindTimesheetByStaffID")
	defer span.End()
	autoCreateFlags, err := r.Retrieve(ctx, db, database.TextArray([]string{staffID.String}))
	if err != nil {
		return nil, err
	}

	if len(autoCreateFlags) == 0 {
		return nil, pgx.ErrNoRows
	}

	if len(autoCreateFlags) > 1 {
		return nil, errors.New("Too many auto create flag")
	}
	return autoCreateFlags[0], nil
}

func (r *AutoCreateFlagRepoImpl) Upsert(ctx context.Context, db database.QueryExecer, e *entity.AutoCreateTimesheetFlag) error {
	ctx, span := interceptors.StartSpan(ctx, "AutoCreateFlagRepoImpl.FindTimesheetByStaffID")
	defer span.End()

	now := time.Now()

	err := multierr.Combine(
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	)
	if err != nil {
		return fmt.Errorf("err set userGroup: %w", err)
	}

	fields, values := e.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fields))

	stmt := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s ;",
		e.TableName(),
		strings.Join(fields, ","),
		placeHolders,
		e.PrimaryKey(),
		e.UpdateOnConflictQuery(),
	)

	cmd, err := db.Exec(ctx, stmt, values...)
	if err != nil {
		return fmt.Errorf("upsert auto create timesheet flag: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("upsert auto create timesheet flag: %d RowsAffected", 0)
	}

	return nil
}
