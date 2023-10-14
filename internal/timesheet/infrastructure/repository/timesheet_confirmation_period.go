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
)

type TimesheetConfirmationPeriodRepoImpl struct {
}

func (r *TimesheetConfirmationPeriodRepoImpl) GetPeriodByDate(ctx context.Context, db database.QueryExecer, date time.Time) (*entity.TimesheetConfirmationPeriod, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetConfirmationPeriodRepoImpl.GetPeriodByDate")
	defer span.End()

	tsConfirmationPeriod := &entity.TimesheetConfirmationPeriod{}
	fields, _ := tsConfirmationPeriod.FieldMap()

	stmt := fmt.Sprintf(`SELECT %s FROM %s
	WHERE deleted_at IS NULL
	AND start_date <= $1
	AND end_date >= $1 limit 1`, strings.Join(fields, ", "), tsConfirmationPeriod.TableName())

	if err := database.Select(ctx, db, stmt, date).ScanOne(tsConfirmationPeriod); err != nil {
		return nil, err
	}
	return tsConfirmationPeriod, nil
}

func (r *TimesheetConfirmationPeriodRepoImpl) InsertPeriod(ctx context.Context, db database.QueryExecer, period *entity.TimesheetConfirmationPeriod) (*entity.TimesheetConfirmationPeriod, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetConfirmationPeriodRepoImpl.InsertPeriod")
	defer span.End()

	if err := period.PreInsert(); err != nil {
		return nil, fmt.Errorf("PreInsert Period failed, err: %w", err)
	}

	cmdTag, err := database.Insert(ctx, period, db.Exec)
	if err != nil {
		return nil, fmt.Errorf("err insert Period: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return nil, fmt.Errorf("err insert Period: %d RowsAffected", cmdTag.RowsAffected())
	}

	return period, nil
}

func (r *TimesheetConfirmationPeriodRepoImpl) GetPeriodByID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entity.TimesheetConfirmationPeriod, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetConfirmationPeriodRepoImpl.GetPeriodByID")
	defer span.End()

	tsConfirmationPeriod := &entity.TimesheetConfirmationPeriod{}
	fields, _ := tsConfirmationPeriod.FieldMap()

	stmt := fmt.Sprintf(`SELECT %s FROM %s
	WHERE deleted_at IS NULL
	AND id = $1`, strings.Join(fields, ", "), tsConfirmationPeriod.TableName())

	if err := database.Select(ctx, db, stmt, id).ScanOne(tsConfirmationPeriod); err != nil {
		return nil, err
	}

	return tsConfirmationPeriod, nil
}
