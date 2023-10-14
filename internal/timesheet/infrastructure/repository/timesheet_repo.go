package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/timesheet/domain/common"
	"github.com/manabie-com/backend/internal/timesheet/domain/constant"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type TimesheetRepoImpl struct {
}

func (t *TimesheetRepoImpl) Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entity.Timesheet, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetRepoImpl.Retrieve")
	defer span.End()

	timesheet := &entity.Timesheet{}
	timesheets := &entity.Timesheets{}

	values, _ := timesheet.FieldMap()
	stmt := fmt.Sprintf(`SELECT %s
	FROM %s
	WHERE deleted_at IS NULL
	AND timesheet_id = ANY($1::_TEXT);`, strings.Join(values, ", "), timesheet.TableName())

	if err := database.Select(ctx, db, stmt, ids).ScanAll(timesheets); err != nil {
		return nil, err
	}

	return *timesheets, nil
}

func (t *TimesheetRepoImpl) CountTimesheets(ctx context.Context, db database.QueryExecer, req *dto.TimesheetCountReq) (*dto.TimesheetCountOut, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetRepoImpl.CountTimesheets")
	defer span.End()

	staffName := pgtype.Text{Status: pgtype.Null}
	locationID := pgtype.Text{Status: pgtype.Null}
	staffID := pgtype.Text{Status: pgtype.Null}
	fromDate := database.Timestamptz(req.FromDate)
	toDate := database.Timestamptz(req.ToDate)

	if req.StaffName != "" {
		staffName = database.Text(req.StaffName)
	}
	if req.LocationID != "" {
		locationID = database.Text(req.LocationID)
	}
	if req.StaffID != "" {
		staffID = database.Text(req.StaffID)
	}

	res := &dto.TimesheetCountOut{}
	// Call the function
	// get_timesheet_count(
	// 	keyword text,
	// 	from_date timestamp with time zone,
	// 	to_date timestamp with time zone,
	// 	location_id_arg text,
	// 	staff_id_arg text
	// )
	values, _ := res.FieldMap()
	stmt := fmt.Sprintf("SELECT %s FROM %s($1,$2,$3,$4,$5)", strings.Join(values, ", "), res.SQLFunctionName())

	err := db.
		QueryRow(ctx, stmt, staffName, fromDate, toDate, locationID, staffID).
		Scan(&res.AllCount, &res.DraftCount, &res.SubmittedCount, &res.ApprovedCount, &res.ConfirmedCount)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	return res, nil
}

func (t *TimesheetRepoImpl) CountTimesheetsV2(ctx context.Context, db database.QueryExecer, req *dto.TimesheetCountV2Req) (*dto.TimesheetCountV2Out, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetRepoImpl.CountTimesheetsV2")
	defer span.End()

	staffName := pgtype.Text{Status: pgtype.Null}
	locationIds := pgtype.TextArray{Status: pgtype.Null}
	staffID := pgtype.Text{Status: pgtype.Null}
	fromDate := database.Timestamptz(req.FromDate)
	toDate := database.Timestamptz(req.ToDate)

	if req.StaffName != "" {
		staffName = database.Text(req.StaffName)
	}
	if len(req.LocationIds) > 0 {
		locationIds = database.TextArray(req.LocationIds)
	}
	if req.StaffID != "" {
		staffID = database.Text(req.StaffID)
	}

	res := &dto.TimesheetCountV2Out{}

	values, _ := res.FieldMap()
	stmt := fmt.Sprintf("SELECT %s FROM %s($1,$2,$3,$4,$5)", strings.Join(values, ", "), res.SQLFunctionName())

	err := db.
		QueryRow(ctx, stmt, staffName, fromDate, toDate, locationIds, staffID).
		Scan(&res.AllCount, &res.DraftCount, &res.SubmittedCount, &res.ApprovedCount, &res.ConfirmedCount)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	return res, nil
}

func (t *TimesheetRepoImpl) GetTimesheetCountByStatusAndLocationIds(ctx context.Context, db database.QueryExecer, req *dto.TimesheetCountByStatusAndLocationIdsReq) (*dto.TimesheetCountByStatusAndLocationIdsResp, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetRepoImpl.GetTimesheetCountByStatusAndLocationIds")
	defer span.End()

	timesheet := &entity.Timesheet{}

	status := database.Text(req.Status)
	locationIds := database.TextArray([]string{})
	fromDate := database.Timestamptz(req.FromDate)
	toDate := database.Timestamptz(req.ToDate)

	if len(req.LocationIds) > 0 {
		locationIds = database.TextArray(req.LocationIds)
	}

	res := &dto.TimesheetCountByStatusAndLocationIdsResp{}
	stmt := fmt.Sprintf(`SELECT
	count(timesheet_id)
FROM %s t
WHERE (t.deleted_at IS NULL)
	AND (
		timesheet_date BETWEEN $2 AND $3
	)
	AND t.location_id = ANY($4)
	AND ( (
			t.timesheet_status = 'TIMESHEET_STATUS_CONFIRMED' :: text
			AND $1 = 'TIMESHEET_STATUS_CONFIRMED' :: text
		)
		OR (
			t.timesheet_status = $1
			AND ( (
					SELECT
						EXISTS(
							SELECT
								1
							FROM
								timesheet_lesson_hours tlh
							WHERE
								tlh.timesheet_id = t.timesheet_id
								AND tlh.flag_on = TRUE
								AND tlh.deleted_at IS NULL
						)
				)
				OR (
					SELECT
						EXISTS(
							SELECT
								1
							FROM
								other_working_hours owh
							WHERE
								owh.timesheet_id = t.timesheet_id
								AND owh.deleted_at IS NULL
						)
				)
				OR (
					SELECT
						EXISTS(
							SELECT
								1
							FROM
								transportation_expense te
							WHERE
								te.timesheet_id = t.timesheet_id
								AND te.deleted_at IS NULL
						)
				)
			)
		)
	)`, timesheet.TableName())

	err := db.
		QueryRow(ctx, stmt, status, fromDate, toDate, locationIds).
		Scan(&res.Count)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (t *TimesheetRepoImpl) FindTimesheetByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) ([]*entity.Timesheet, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetRepoImpl.FindTimesheetByLessonIDs")
	defer span.End()

	timesheet := &entity.Timesheet{}
	timesheetLessonHour := &entity.TimesheetLessonHours{}
	timesheets := &entity.Timesheets{}
	values, _ := timesheet.FieldMap()

	stmt := fmt.Sprintf(
		`SELECT ts.%s FROM %s AS ts INNER JOIN %s AS tlh ON ts.timesheet_id = tlh.timesheet_id
        WHERE lesson_id = ANY($1::_TEXT) AND tlh.deleted_at IS NULL`,
		strings.Join(values, ", ts."), timesheet.TableName(), timesheetLessonHour.TableName())

	if err := database.Select(ctx, db, stmt, lessonIDs).ScanAll(timesheets); err != nil {
		return nil, err
	}

	return *timesheets, nil
}

func (t *TimesheetRepoImpl) FindTimesheetByTimesheetIDs(ctx context.Context, db database.QueryExecer, timesheetIDs []string) ([]*entity.Timesheet, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetRepoImpl.FindTimesheetByTimesheetID")
	defer span.End()
	timesheets, err := t.Retrieve(ctx, db, database.TextArray(timesheetIDs))
	if err != nil {
		return nil, err
	}

	return timesheets, nil
}

func (t *TimesheetRepoImpl) FindTimesheetByTimesheetID(ctx context.Context, db database.QueryExecer, timesheetID pgtype.Text) (*entity.Timesheet, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetRepoImpl.FindTimesheetByTimesheetID")
	defer span.End()
	timesheets, err := t.Retrieve(ctx, db, database.TextArray([]string{timesheetID.String}))
	if err != nil {
		return nil, fmt.Errorf("%s, timesheet_id: %s", err.Error(), timesheetID.String)
	}

	if len(timesheets) == 0 {
		return nil, fmt.Errorf("%s, timesheet_id: %s", pgx.ErrNoRows.Error(), timesheetID.String)
	}

	if len(timesheets) > 1 {
		return nil, fmt.Errorf("too many timesheet, timesheet_id: %s", timesheetID.String)
	}

	return timesheets[0], nil
}

func (t *TimesheetRepoImpl) FindTimesheetByTimesheetArgs(ctx context.Context, db database.QueryExecer, timesheetArgs *dto.TimesheetQueryArgs) ([]*entity.Timesheet, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetRepoImpl.FindTimesheetByTimesheetArgs")
	defer span.End()

	timesheet := &entity.Timesheet{}
	timesheets := &entity.Timesheets{}

	values, _ := timesheet.FieldMap()
	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE deleted_at IS NULL 
	AND staff_id IN (%s) 
	AND timesheet_date = $1 
	AND location_id = $2;`, strings.Join(values, constant.SeparatorComma), timesheet.TableName(), common.ConcatQueryValue(timesheetArgs.StaffIDs...))
	if err := database.Select(ctx, db, stmt, &timesheetArgs.TimesheetDate, &timesheetArgs.LocationID).ScanAll(timesheets); err != nil {
		return nil, err
	}

	return *timesheets, nil
}

func (t *TimesheetRepoImpl) GetStaffTimesheetIDsAfterDateCanChange(ctx context.Context, db database.QueryExecer, staffID string, date time.Time) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetRepoImpl.GetStaffTimesheetIDsAfterDateCanChange")
	defer span.End()

	stmt := fmt.Sprintf(`
		SELECT timesheet_id 
		FROM %s 
		WHERE deleted_at IS NULL 
			AND staff_id = $1 
			AND timesheet_date >= $2
			AND timesheet_status <> $3;`,
		(&entity.Timesheet{}).TableName(),
	)

	rows, err := db.Query(ctx, stmt, &staffID, &date, pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String())
	if err != nil {
		return nil, fmt.Errorf("err db.Query: %w", err)
	}
	defer rows.Close()

	timesheetIDs := make([]string, 0)

	for rows.Next() {
		var timesheetID string
		if err := rows.Scan(&timesheetID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		timesheetIDs = append(timesheetIDs, timesheetID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return timesheetIDs, nil
}

func (t *TimesheetRepoImpl) GetStaffFutureTimesheetIDsWithLocations(ctx context.Context, db database.QueryExecer, staffID string, date time.Time, locationIDs []string) ([]dto.TimesheetLocationDto, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetRepoImpl.GetStaffFutureTimesheetIDsWithLocations")
	defer span.End()

	stmt := fmt.Sprintf(`
		SELECT timesheet_id, location_id  
		FROM %s 
		WHERE deleted_at IS NULL 
			AND staff_id = $1 
			AND timesheet_date > $2
			AND location_id = ANY($3::_TEXT)
			AND timesheet_status <> $4;`,
		(&entity.Timesheet{}).TableName(),
	)

	rows, err := db.Query(ctx, stmt, &staffID, &date, &locationIDs, pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String())
	if err != nil {
		return nil, fmt.Errorf("err db.Query: %w", err)
	}
	defer rows.Close()

	timesheetIDs := make([]dto.TimesheetLocationDto, 0)

	for rows.Next() {
		tsDto := dto.TimesheetLocationDto{}
		if err := rows.Scan(&tsDto.TimesheetID, &tsDto.LocationID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		timesheetIDs = append(timesheetIDs, tsDto)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return timesheetIDs, nil
}

func (t *TimesheetRepoImpl) InsertTimeSheet(ctx context.Context, db database.QueryExecer, timesheet *entity.Timesheet) (*entity.Timesheet, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetRepoImpl.InsertTimeSheet")
	defer span.End()

	if err := timesheet.PreInsert(); err != nil {
		return nil, fmt.Errorf("PreInsert Timesheet failed, err: %w", err)
	}

	cmdTag, err := database.Insert(ctx, timesheet, db.Exec)
	if err != nil {
		return nil, fmt.Errorf("err insert Timesheet: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return nil, fmt.Errorf("err insert Timesheet: %d RowsAffected", cmdTag.RowsAffected())
	}

	return timesheet, nil
}

func (t *TimesheetRepoImpl) UpdateTimeSheet(ctx context.Context, db database.QueryExecer, timesheet *entity.Timesheet) (*entity.Timesheet, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetRepoImpl.Update")
	defer span.End()

	if err := timesheet.PreUpdate(); err != nil {
		return nil, fmt.Errorf("PreUpdate Timesheet failed, err: %w", err)
	}

	var fields []string
	for _, field := range database.GetFieldNames(timesheet) {
		if field != "timesheet_id" && field != "created_at" {
			fields = append(fields, field)
		}
	}
	cmdTag, err := database.UpdateFields(ctx, timesheet, db.Exec, "timesheet_id", fields)

	if err != nil {
		return nil, fmt.Errorf("err update Timesheet: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return nil, fmt.Errorf("err update Timesheet: %d RowsAffected", cmdTag.RowsAffected())
	}
	return timesheet, nil
}

func (t *TimesheetRepoImpl) SoftDeleteByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetRepoImpl.SoftDeleteByIDs")
	defer span.End()

	e := &entity.Timesheet{}

	stmt := fmt.Sprintf(`
		UPDATE %s SET deleted_at = $1
		WHERE timesheet_id = ANY($2::_TEXT)
		AND deleted_at IS NULL;`, e.TableName())

	cmdTag, err := db.Exec(ctx, stmt, time.Now(), ids)
	if err != nil {
		return fmt.Errorf("err delete TimesheetRepoImpl: %w", err)
	}
	if cmdTag.RowsAffected() != int64(len(ids.Elements)) {
		return fmt.Errorf("err delete TimesheetRepoImpl: %d RowsAffected", cmdTag.RowsAffected())
	}
	return nil
}

func (t *TimesheetRepoImpl) RemoveTimesheetRemarkByTimesheetIDs(ctx context.Context, db database.QueryExecer, ids []string) error {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetRepoImpl.RemoveTimesheetRemarkByTimesheetIDs")
	defer span.End()

	e := &entity.Timesheet{}

	stmt := fmt.Sprintf(`
		UPDATE %s SET remark = NULL
		WHERE timesheet_id = ANY($1::_TEXT)
		AND deleted_at IS NULL;`, e.TableName())

	cmdTag, err := db.Exec(ctx, stmt, ids)
	if err != nil {
		return fmt.Errorf("err RemoveTimesheetRemarkByTimesheetIDs: %w", err)
	}
	if cmdTag.RowsAffected() != int64(len(ids)) {
		return fmt.Errorf("err RemoveTimesheetRemarkByTimesheetIDs: %d RowsAffected", cmdTag.RowsAffected())
	}
	return nil
}

func (t *TimesheetRepoImpl) UpsertMultiple(ctx context.Context, db database.QueryExecer, timesheets []*entity.Timesheet) ([]*entity.Timesheet, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetRepoImpl.InsertMultiple")
	defer span.End()

	batch := &pgx.Batch{}
	now := time.Now()

	for _, timesheet := range timesheets {
		if timesheet.CreatedAt.Time.IsZero() {
			err := timesheet.CreatedAt.Set(now)
			if err != nil {
				return nil, err
			}
		}

		err := timesheet.UpdatedAt.Set(now)
		if err != nil {
			return nil, err
		}

		fields, values := timesheet.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s ;",
			timesheet.TableName(),
			strings.Join(fields, ","),
			placeHolders,
			timesheet.PrimaryKey(),
			timesheet.UpdateOnConflictQuery(),
		)

		batch.Queue(stmt, values...)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < len(timesheets); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return nil, err
		}
		if cmdTag.RowsAffected() != 1 {
			return nil, fmt.Errorf("err upsert multiple timesheet: %d RowsAffected", cmdTag.RowsAffected())
		}
	}

	return timesheets, nil
}

func (t *TimesheetRepoImpl) FindTimesheetByTimesheetIDsAndStatus(ctx context.Context, db database.QueryExecer, ids []string, timesheetStatus string) ([]*entity.Timesheet, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetRepoImpl.FindSubmittedTimesheetByTimesheetIDs")
	defer span.End()

	timesheet := &entity.Timesheet{}
	timesheets := &entity.Timesheets{}
	values, _ := timesheet.FieldMap()

	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE deleted_at IS NULL 
	AND timesheet_id IN (%s) 
	AND timesheet_status = $1;`, strings.Join(values, constant.SeparatorComma), timesheet.TableName(), common.ConcatQueryValue(ids...))
	if err := database.Select(ctx, db, stmt, timesheetStatus).ScanAll(timesheets); err != nil {
		return nil, err
	}

	if len(*timesheets) == 0 {
		return nil, fmt.Errorf("no rows affected")
	}

	return *timesheets, nil
}

func (t *TimesheetRepoImpl) UpdateTimesheetStatusMultiple(ctx context.Context, db database.QueryExecer, timesheets []*entity.Timesheet, timesheetStatus string) error {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetRepoImpl.UpdateTimesheetStatusMultiple")
	defer span.End()

	batchQueue := &pgx.Batch{}
	timesheetEntity := &entity.Timesheet{}

	stmt := fmt.Sprintf(`UPDATE %s SET timesheet_status = $1, updated_at = NOW() WHERE timesheet_id = $2;`, timesheetEntity.TableName())

	for _, timesheet := range timesheets {
		batchQueue.Queue(stmt, &timesheetStatus, &timesheet.TimesheetID)
	}

	results := db.SendBatch(ctx, batchQueue)
	defer results.Close()

	for i := 0; i < batchQueue.Len(); i++ {
		cmdTag, err := results.Exec()
		if err != nil {
			return fmt.Errorf("results.Exec: %w", err)
		}
		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("err update timesheet status multiple: %d RowsAffected", cmdTag.RowsAffected())
		}
	}

	return nil
}

func (t *TimesheetRepoImpl) CountNotApprovedAndNotConfirmedTimesheet(ctx context.Context, db database.QueryExecer, locationID string, startDate, endDate time.Time) (int, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetRepoImpl.CountNotApprovedAndNotConfirmedTimesheet")
	defer span.End()
	var count int
	timesheet := &entity.Timesheet{}

	stmt := fmt.Sprintf(`
	SELECT count(*)
	FROM %s t
	WHERE (t.deleted_at IS NULL)
	  	AND ( timesheet_date BETWEEN $1 AND $2 )
		AND t.location_id = $3
		AND t.timesheet_status <> $4
		AND t.timesheet_status <> $5
		AND (
				(
					(
						SELECT COUNT(*)
						FROM timesheet_lesson_hours tlh
						WHERE tlh.timesheet_id = t.timesheet_id
						AND tlh.flag_on = TRUE
						AND tlh.deleted_at IS NULL
					) > 0
				)
				OR (
						(
							SELECT COUNT(*)
							FROM other_working_hours owh
							WHERE owh.timesheet_id = t.timesheet_id
							AND owh.deleted_at IS NULL
						) > 0
					)
				OR (
						(
							SELECT COUNT(*)
							FROM transportation_expense te
							WHERE te.timesheet_id = t.timesheet_id
							AND te.deleted_at IS NULL
						) > 0
					)
	)`, timesheet.TableName())

	err := db.QueryRow(ctx, stmt, startDate, endDate, locationID,
		pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String(),
		pb.TimesheetStatus_TIMESHEET_STATUS_CONFIRMED.String()).Scan(&count)
	if err != nil {
		return -1, err
	}

	return count, nil
}

func (t *TimesheetRepoImpl) FindTimesheetInLocationByDateAndStatus(ctx context.Context, db database.QueryExecer, locationID, timesheetStatus string, startDate, endDate time.Time) ([]*entity.Timesheet, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetRepoImpl.FindTimesheetInLocationByDateAndStatus")
	defer span.End()

	timesheet := &entity.Timesheet{}
	timesheets := &entity.Timesheets{}
	fields, _ := timesheet.FieldMap()

	stmt := fmt.Sprintf(
		`SELECT %s FROM %s
        WHERE timesheet_date <= $1
		AND timesheet_date >= $2 
		AND timesheet_status = $3 
		AND location_id = $4 
		AND deleted_at IS NULL`,
		strings.Join(fields, ", "), timesheet.TableName())

	if err := database.Select(ctx, db, stmt, endDate, startDate, timesheetStatus, locationID).ScanAll(timesheets); err != nil {
		return nil, err
	}

	return *timesheets, nil
}

func (t *TimesheetRepoImpl) UpdateTimesheetStatusToConfirmByDateAndLocation(ctx context.Context, db database.QueryExecer, startDate, endDate time.Time, timesheetStatus, locationID string) error {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetRepoImpl.UpdateTimesheetStatusToConfirmByDateAndLocation")
	defer span.End()

	timesheetEntity := &entity.Timesheet{}

	stmt := fmt.Sprintf(`UPDATE %s SET timesheet_status = $1 
		WHERE timesheet_date >= $2 
		AND timesheet_date <= $3
		AND timesheet_status = $4
		AND location_id = $5 
		AND deleted_at IS NULL;`, timesheetEntity.TableName())

	_, err := db.Exec(ctx, stmt, pb.TimesheetStatus_TIMESHEET_STATUS_CONFIRMED.String(), startDate, endDate, timesheetStatus, locationID)

	return err
}
