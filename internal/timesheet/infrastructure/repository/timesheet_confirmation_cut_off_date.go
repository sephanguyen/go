package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
)

type TimesheetConfirmationCutOffDateRepoImpl struct {
}

func (r *TimesheetConfirmationCutOffDateRepoImpl) GetCutOffDateByDate(ctx context.Context, db database.QueryExecer, date time.Time) (*entity.TimesheetConfirmationCutOffDate, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetConfirmationCutOffDateRepoImpl.GetCutOffDateByDate")
	defer span.End()

	tsConfirmationCutOffDate := &entity.TimesheetConfirmationCutOffDate{}
	fields, _ := tsConfirmationCutOffDate.FieldMap()

	stmt := fmt.Sprintf(`SELECT %s FROM %s
	WHERE deleted_at IS NULL
	AND start_date <= $1
	AND end_date >= $1 limit 1`, strings.Join(fields, ", "), tsConfirmationCutOffDate.TableName())

	if err := database.Select(ctx, db, stmt, date).ScanOne(tsConfirmationCutOffDate); err != nil {
		return nil, err
	}
	return tsConfirmationCutOffDate, nil
}
