package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/timesheet/domain/common"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type TimesheetLessonHoursRepoImpl struct{}

func (r *TimesheetLessonHoursRepoImpl) FindTimesheetLessonHoursByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) ([]*entity.TimesheetLessonHours, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetLessonHoursRepoImpl.FindTimesheetLessonHoursByLessonIDs")
	defer span.End()

	timesheetLessonHour := &entity.TimesheetLessonHours{}
	timesheetLessonHours := &entity.ListTimesheetLessonHours{}

	values, _ := timesheetLessonHour.FieldMap()
	lessonIDsToTextArray := &pgtype.TextArray{}
	err := lessonIDsToTextArray.Set(lessonIDs)

	if err != nil {
		return nil, fmt.Errorf("err convert string array to TextArray: %v", err.Error())
	}

	stmt := fmt.Sprintf(
		"SELECT %s FROM %s WHERE lesson_id = ANY($1::_TEXT) AND deleted_at IS NULL",
		strings.Join(values, ", "),
		timesheetLessonHour.TableName())

	if err := database.Select(ctx, db, stmt, lessonIDsToTextArray).ScanAll(timesheetLessonHours); err != nil {
		return nil, err
	}

	return *timesheetLessonHours, nil
}

func (r *TimesheetLessonHoursRepoImpl) FindTimesheetLessonHoursByTimesheetID(ctx context.Context, db database.QueryExecer, timesheetID pgtype.Text) ([]*entity.TimesheetLessonHours, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetLessonHoursRepoImpl.FindLessonRecordsByTimesheetID")
	defer span.End()

	timesheetLessonHour := &entity.TimesheetLessonHours{}
	timesheetLessonHours := &entity.ListTimesheetLessonHours{}

	values, _ := timesheetLessonHour.FieldMap()

	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE timesheet_id = $1 AND deleted_at IS NULL", strings.Join(values, ", "), timesheetLessonHour.TableName())

	if err := database.Select(ctx, db, stmt, &timesheetID).ScanAll(timesheetLessonHours); err != nil {
		return nil, err
	}

	return *timesheetLessonHours, nil
}

func (r *TimesheetLessonHoursRepoImpl) InsertMultiple(ctx context.Context, db database.QueryExecer, listTimesheetLessonHours []*entity.TimesheetLessonHours) ([]*entity.TimesheetLessonHours, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetLessonHoursRepoImpl.InsertMultiple")
	defer span.End()

	batch := &pgx.Batch{}
	now := time.Now()
	for _, timesheetLessonHours := range listTimesheetLessonHours {
		err := multierr.Combine(
			timesheetLessonHours.CreatedAt.Set(now),
			timesheetLessonHours.UpdatedAt.Set(now),
			timesheetLessonHours.DeletedAt.Set(nil),
		)
		if err != nil {
			return nil, err
		}

		fields, values := timesheetLessonHours.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s);",
			timesheetLessonHours.TableName(),
			strings.Join(fields, ","),
			placeHolders,
		)

		batch.Queue(stmt, values...)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < len(listTimesheetLessonHours); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return nil, err
		}
		if cmdTag.RowsAffected() != 1 {
			return nil, fmt.Errorf("err insert TimesheetLessonHours: %d RowsAffected", cmdTag.RowsAffected())
		}
	}

	return listTimesheetLessonHours, nil
}

func (r *TimesheetLessonHoursRepoImpl) FindByTimesheetIDs(ctx context.Context, db database.QueryExecer, timesheetIDs []string) ([]*entity.TimesheetLessonHours, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetLessonHoursRepoImpl.FindByTimesheetIDs")
	defer span.End()

	timesheetLessonHours := &entity.TimesheetLessonHours{}
	listTimesheetLessonHours := &entity.ListTimesheetLessonHours{}

	values, _ := timesheetLessonHours.FieldMap()

	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE deleted_at IS NULL AND timesheet_id IN (%s)",
		strings.Join(values, ", "), timesheetLessonHours.TableName(), common.ConcatQueryValue(timesheetIDs...))

	if err := database.Select(ctx, db, stmt).ScanAll(listTimesheetLessonHours); err != nil {
		return nil, err
	}

	return *listTimesheetLessonHours, nil
}

func (r *TimesheetLessonHoursRepoImpl) SoftDelete(ctx context.Context, db database.QueryExecer, listTimesheetLessonHours []*entity.TimesheetLessonHours) error {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetLessonHoursRepoImpl.SoftDelete")
	defer span.End()
	var (
		timesheetIDs = make([]string, 0, len(listTimesheetLessonHours))
		lessonIDs    = make([]string, 0, len(listTimesheetLessonHours))
	)
	for _, e := range listTimesheetLessonHours {
		timesheetIDs = append(timesheetIDs, e.TimesheetID.String)
		lessonIDs = append(lessonIDs, e.LessonID.String)
	}

	timesheetLessonHours := &entity.TimesheetLessonHours{}
	stmt := fmt.Sprintf(`
		UPDATE %s SET deleted_at = $1
		WHERE timesheet_id = ANY($2::_TEXT)
		AND lesson_id = ANY($3::_TEXT)
		AND deleted_at IS NULL;`, timesheetLessonHours.TableName())

	if _, err := db.Exec(ctx, stmt, time.Now(), timesheetIDs, lessonIDs); err != nil {
		return fmt.Errorf("err delete SoftDelete: %w", err)
	}
	return nil
}

func (r *TimesheetLessonHoursRepoImpl) UpsertMultiple(ctx context.Context, db database.QueryExecer, listTimesheetLessonHours []*entity.TimesheetLessonHours) ([]*entity.TimesheetLessonHours, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetLessonHoursRepoImpl.InsertMultiple")
	defer span.End()

	batch := &pgx.Batch{}
	now := time.Now()
	for _, timesheetLessonHours := range listTimesheetLessonHours {
		err := multierr.Combine(
			timesheetLessonHours.CreatedAt.Set(now),
			timesheetLessonHours.UpdatedAt.Set(now),
			timesheetLessonHours.DeletedAt.Set(nil),
		)
		if err != nil {
			return nil, err
		}

		fields, values := timesheetLessonHours.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s ;",
			timesheetLessonHours.TableName(),
			strings.Join(fields, ","),
			placeHolders,
			timesheetLessonHours.PrimaryKey(),
			timesheetLessonHours.UpdateOnConflictQuery(),
		)

		batch.Queue(stmt, values...)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < len(listTimesheetLessonHours); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return nil, err
		}
		if cmdTag.RowsAffected() != 1 {
			return nil, fmt.Errorf("err insert TimesheetLessonHours: %d RowsAffected", cmdTag.RowsAffected())
		}
	}

	return listTimesheetLessonHours, nil
}

func (r *TimesheetLessonHoursRepoImpl) UpdateAutoCreateFlagStateAfterTime(ctx context.Context, db database.QueryExecer, timesheetIDs []string, updateTime time.Time, flagOn bool) error {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetLessonHoursRepoImpl.UpdateAutoCreateFlagStateAfterTime")
	defer span.End()

	stmt := fmt.Sprintf(`
		UPDATE
			%s t
		SET 
			flag_on = $3, updated_at = NOW()
		FROM
			%s l
		WHERE 
			t.deleted_at IS NULL
			AND t.timesheet_id = ANY($1::_TEXT)
			AND t.lesson_id = l.lesson_id
			AND l.start_time >= $2
			AND t.flag_on <> $3;`, (&entity.TimesheetLessonHours{}).TableName(), (&entity.Lesson{}).TableName())

	_, err := db.Exec(ctx, stmt, timesheetIDs, updateTime, flagOn)
	if err != nil {
		return fmt.Errorf("err update TimesheetLessonHoursRepoImpl: %w", err)
	}

	return nil
}

func (r *TimesheetLessonHoursRepoImpl) UpdateTimesheetLessonAutoCreateFlagByTimesheetIDs(ctx context.Context, db database.QueryExecer, timesheetIDs []string, flagOn bool) error {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetLessonHoursRepoImpl.UpdateTimesheetLessonAutoCreateFlagByTimesheetIDs")
	defer span.End()

	stmt := fmt.Sprintf(`
		UPDATE
			%s t
		SET 
			flag_on = $2, updated_at = NOW()
		FROM
			%s l
		WHERE 
			t.deleted_at IS NULL
			AND t.timesheet_id = ANY($1::_TEXT)
			AND t.lesson_id = l.lesson_id
			AND t.flag_on <> $2;`, (&entity.TimesheetLessonHours{}).TableName(), (&entity.Lesson{}).TableName())

	_, err := db.Exec(ctx, stmt, timesheetIDs, flagOn)
	if err != nil {
		return fmt.Errorf("err update TimesheetLessonHoursRepoImpl::UpdateTimesheetLessonAutoCreateFlagByTimesheetIDs: %w", err)
	}

	return nil
}

// Get Map Existing lesson hours by timesheet ids
// return: map existing ids
func (r *TimesheetLessonHoursRepoImpl) MapExistingLessonHoursByTimesheetIds(ctx context.Context, db database.QueryExecer, ids []string) (map[string]struct{}, error) {
	ctx, span := interceptors.StartSpan(ctx, "MapExistingLessonHoursByTimesheetIds")
	defer span.End()

	stmt := fmt.Sprintf(`
	SELECT timesheet_id	
	FROM %s
	WHERE timesheet_id = ANY($1::_TEXT) AND deleted_at IS NULL AND flag_on = true
	GROUP BY timesheet_id;`,
		(&entity.TimesheetLessonHours{}).TableName(),
	)

	// flag_on = true: mean not count
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
