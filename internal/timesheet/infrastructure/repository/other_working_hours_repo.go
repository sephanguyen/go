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
	"go.uber.org/multierr"
)

type OtherWorkingHoursRepoImpl struct{}

func (r *OtherWorkingHoursRepoImpl) UpsertMultiple(ctx context.Context, db database.QueryExecer, listOWHs []*entity.OtherWorkingHours) error {
	ctx, span := interceptors.StartSpan(ctx, "OtherWorkingHoursRepoImpl.UpsertMultiple")
	defer span.End()

	batch := &pgx.Batch{}
	now := time.Now()

	for _, otherWorkingHoursInfo := range listOWHs {
		err := multierr.Combine(
			otherWorkingHoursInfo.UpdatedAt.Set(now),
			otherWorkingHoursInfo.CreatedAt.Set(now),
		)
		if err != nil {
			return err
		}

		fields, values := otherWorkingHoursInfo.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s ;",
			otherWorkingHoursInfo.TableName(),
			strings.Join(fields, ","),
			placeHolders,
			otherWorkingHoursInfo.UpsertConflictField(),
			otherWorkingHoursInfo.UpdateOnConflictQuery(),
		)

		batch.Queue(stmt, values...)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < len(listOWHs); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return err
		}
		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("err upsert Other Working Hours: %d RowsAffected", cmdTag.RowsAffected())
		}
	}

	return nil
}

func (r *OtherWorkingHoursRepoImpl) FindListOtherWorkingHoursByTimesheetIDs(ctx context.Context, db database.QueryExecer, timesheetIDs pgtype.TextArray) ([]*entity.OtherWorkingHours, error) {
	ctx, span := interceptors.StartSpan(ctx, "OtherWorkingHoursRepoImpl.Retrieve")
	defer span.End()

	owhsE := &entity.OtherWorkingHours{}
	listOWHsE := &entity.ListOtherWorkingHours{}

	values, _ := owhsE.FieldMap()
	stmt := fmt.Sprintf(`
	SELECT %s
	FROM %s
	WHERE deleted_at IS NULL
	AND timesheet_id = ANY($1::_TEXT);`, strings.Join(values, ", "), owhsE.TableName())

	if err := database.Select(ctx, db, stmt, timesheetIDs).ScanAll(listOWHsE); err != nil {
		return nil, err
	}

	return *listOWHsE, nil
}

func (r *OtherWorkingHoursRepoImpl) SoftDeleteByTimesheetID(ctx context.Context, db database.QueryExecer, timesheetID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "OtherWorkingHoursRepoImpl.SoftDeleteByTimesheetID")
	defer span.End()

	e := &entity.OtherWorkingHours{}

	stmt := fmt.Sprintf(`
		UPDATE %s SET deleted_at = NOW()
		WHERE timesheet_id = $1 AND deleted_at IS NULL;
	`, e.TableName())

	_, err := db.Exec(ctx, stmt, &timesheetID)
	if err != nil {
		return fmt.Errorf("err delete SoftDeleteByTimesheetID: %w", err)
	}

	return nil
}

// Get Map Existing other working hours by timesheet ids
// return: map existing ids
func (r *OtherWorkingHoursRepoImpl) MapExistingOWHsByTimesheetIds(ctx context.Context, db database.QueryExecer, ids []string) (map[string]struct{}, error) {
	ctx, span := interceptors.StartSpan(ctx, "OtherWorkingHoursRepoImpl.MapExistingOWHsByTimesheetIds")
	defer span.End()

	stmt := fmt.Sprintf(`
	SELECT timesheet_id 
	FROM %s 
	WHERE timesheet_id = ANY($1::_TEXT) AND deleted_at IS NULL 
	GROUP BY timesheet_id;`,
		(&entity.OtherWorkingHours{}).TableName())

	// Query
	rows, err := db.Query(ctx, stmt, ids)

	if err != nil {
		return nil, err
	}

	result := map[string]struct{}{}
	for rows.Next() {
		var timesheetID string
		if err = rows.Scan(&timesheetID); err != nil {
			return nil, err
		}
		result[timesheetID] = struct{}{}
	}

	return result, nil
}
