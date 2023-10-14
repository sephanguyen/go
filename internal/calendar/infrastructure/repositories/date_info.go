package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type DateInfoRepo struct{}

func (d *DateInfoRepo) GetDateInfoByDateAndLocationID(ctx context.Context, db database.QueryExecer, date time.Time, locationID string) (*dto.DateInfo, error) {
	ctx, span := interceptors.StartSpan(ctx, "DateInfoRepo.GetDateInfoByDateAndLocationID")
	defer span.End()

	dateInfo := &DateInfo{}
	fields, values := dateInfo.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s 
		FROM %s 
		WHERE date = $1::Date and location_id = $2
		AND deleted_at IS NULL`,
		strings.Join(fields, ","),
		dateInfo.TableName(),
	)

	if err := db.QueryRow(ctx, query, &date, &locationID).Scan(values...); err != nil {
		return nil, fmt.Errorf("failed to query date info: %w", err)
	}

	return dateInfo.ConvertToDTO(), nil
}

func (d *DateInfoRepo) GetDateInfoByDateRangeAndLocationID(ctx context.Context, db database.QueryExecer, startDate, endDate time.Time, locationID string) ([]*dto.DateInfo, error) {
	ctx, span := interceptors.StartSpan(ctx, "DateInfoRepo.GetDateInfoByDateRangeAndLocationID")
	defer span.End()

	dateInfo := &DateInfo{}
	fields := database.GetFieldNames(dateInfo)
	query := fmt.Sprintf(`SELECT %s
		FROM %s
		WHERE (date BETWEEN ($1 at time zone time_zone)::date AND ($2 at time zone time_zone)::date )
			AND location_id = $3
			AND deleted_at IS NULL `,
		strings.Join(fields, ","),
		dateInfo.TableName(),
	)

	rows, err := db.Query(ctx, query, &startDate, &endDate, &locationID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query date info by date range: db.Query")
	}
	defer rows.Close()

	dateInfos := []*DateInfo{}
	for rows.Next() {
		dateInfo := &DateInfo{}
		if err := rows.Scan(database.GetScanFields(dateInfo, fields)...); err != nil {
			return nil, errors.Wrap(err, "failed to query date info by date range: rows.Scan")
		}

		dateInfos = append(dateInfos, dateInfo)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to query date info by date range: rows.Err")
	}

	dateInfoList := []*dto.DateInfo{}
	for _, dateInfo := range dateInfos {
		dateInfoList = append(dateInfoList, dateInfo.ConvertToDTO())
	}

	return dateInfoList, nil
}

func (d *DateInfoRepo) GetDateInfoDetailedByDateRangeAndLocationID(ctx context.Context, db database.QueryExecer, startDate, endDate time.Time, locationID, timezone string) ([]*dto.DateInfo, error) {
	ctx, span := interceptors.StartSpan(ctx, "DateInfoRepo.GetDateInfoDetailedByDateRangeAndLocationID")
	defer span.End()

	query := `SELECT di.date, di.location_id, di.day_type_id, di.opening_time, 
					 di.time_zone, di.status, dt.display_name
		FROM day_info di
		LEFT JOIN day_type dt ON di.day_type_id = dt.day_type_id 
			AND di.resource_path = dt.resource_path
		WHERE (di.date BETWEEN ($1 at time zone $4)::date AND ($2 at time zone $4)::date)
			AND di.location_id = $3
			AND di.deleted_at IS NULL 
			AND dt.deleted_at IS NULL
		ORDER BY di.date ASC `

	rows, err := db.Query(ctx, query, &startDate, &endDate, &locationID, &timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to query date info detailed by date range: db.Query: %w", err)
	}
	defer rows.Close()

	dateInfo := &DateInfo{}
	_, values := dateInfo.ExportFieldMap()
	var displayName pgtype.Text
	values = append(values, &displayName)

	dateInfoList := []*dto.DateInfo{}
	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("failed to query date info detailed by date range: rows.Scan: %w", err)
		}
		dateInfoDTO := dateInfo.ConvertToDTO()
		dateInfoDTO.DateTypeDisplayName = displayName.String

		dateInfoList = append(dateInfoList, dateInfoDTO)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to query date info detailed by date range: rows.Err: %w", err)
	}

	return dateInfoList, nil
}

func (d *DateInfoRepo) UpsertDateInfo(ctx context.Context, db database.QueryExecer, params *dto.UpsertDateInfoParams) error {
	ctx, span := interceptors.StartSpan(ctx, "DateInfoRepo.UpsertDateInfo")
	defer span.End()

	dateInfo, err := NewDateInfo(map[string]interface{}{
		"date":         params.DateInfo.Date,
		"location_id":  params.DateInfo.LocationID,
		"day_type_id":  params.DateInfo.DateTypeID,
		"opening_time": params.DateInfo.OpeningTime,
		"status":       params.DateInfo.Status,
		"time_zone":    params.DateInfo.TimeZone,
	})
	if err != nil {
		return fmt.Errorf("got error on NewDateInfo: %w", err)
	}

	if err := dateInfo.PreUpsert(); err != nil {
		return fmt.Errorf("got error on PreUpsert date info: %w", err)
	}

	fields := database.GetFieldNamesExcepts(dateInfo, []string{"deleted_at"})
	args := database.GetScanFields(dateInfo, fields)
	placeHolders := database.GeneratePlaceholders(len(fields))

	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT day_info_pk
		DO UPDATE SET 
				date = $1::Date, 
				location_id = $2, 
				day_type_id = $3, 
				opening_time = $4, 
				status = $5, 
				updated_at = $7, 
				time_zone = $8`,
		dateInfo.TableName(),
		strings.Join(fields, ","),
		placeHolders,
	)

	commandTag, err := db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to upsert date info: %w", err)
	}

	if commandTag.RowsAffected() != 1 {
		return fmt.Errorf("error on upsert date info, rows affected is not 1")
	}

	return nil
}

func (d *DateInfoRepo) DuplicateDateInfo(ctx context.Context, db database.QueryExecer, params *dto.DuplicateDateInfoParams) error {
	ctx, span := interceptors.StartSpan(ctx, "DateInfoRepo.DuplicateDateInfo")
	defer span.End()

	dateInfos := make([]*DateInfo, 0, len(params.Dates))

	for _, date := range params.Dates {
		dateInfo, err := NewDateInfo(map[string]interface{}{
			"date":         date,
			"location_id":  params.DateInfo.LocationID,
			"day_type_id":  params.DateInfo.DateTypeID,
			"opening_time": params.DateInfo.OpeningTime,
			"status":       params.DateInfo.Status,
			"time_zone":    params.DateInfo.TimeZone,
		})
		if err != nil {
			return fmt.Errorf("got error on NewDateInfo: %w", err)
		}

		dateInfos = append(dateInfos, dateInfo)
	}

	if err := d.bulkUpsertDateInfo(ctx, db, dateInfos); err != nil {
		return fmt.Errorf("got error on bulkUpsertDateInfo: %w", err)
	}

	return nil
}

func (d *DateInfoRepo) bulkUpsertDateInfo(ctx context.Context, db database.QueryExecer, dateInfos []*DateInfo) error {
	queueFn := func(b *pgx.Batch, dateInfo *DateInfo) {
		fieldNames := database.GetFieldNamesExcepts(dateInfo, []string{"deleted_at"})
		args := database.GetScanFields(dateInfo, fieldNames)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)
			ON CONFLICT ON CONSTRAINT day_info_pk
			DO UPDATE SET 
				date = $1::Date, 
				location_id = $2, 
				day_type_id = $3, 
				opening_time = $4, 
				status = $5, 
				updated_at = $7, 
				time_zone = $8`,
			dateInfo.TableName(),
			strings.Join(fieldNames, ","),
			placeHolders,
		)
		b.Queue(query, args...)
	}

	b := &pgx.Batch{}
	for _, dateInfo := range dateInfos {
		if err := dateInfo.PreUpsert(); err != nil {
			return fmt.Errorf("got error on PreUpsert date info: %w", err)
		}

		queueFn(b, dateInfo)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(dateInfos); i++ {
		commandTag, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("failed to bulk upsert date info batchResults.Exec: %w", err)
		}
		if commandTag.RowsAffected() != 1 {
			return fmt.Errorf("date info setting not inserted")
		}
	}

	return nil
}

func (d *DateInfoRepo) GetAllToExport(ctx context.Context, db database.QueryExecer) ([]byte, error) {
	ctx, span := interceptors.StartSpan(ctx, "DateInfoRepo.GetAllToExport")
	defer span.End()

	stmt := `SELECT %s FROM %s WHERE deleted_at is NULL`
	dateInfo := DateInfo{}
	fieldNames, _ := dateInfo.ExportFieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		dateInfo.TableName(),
	)

	rows, err := db.Query(ctx, stmt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all date info: db.Query")
	}
	defer rows.Close()

	allDayInfos := []*DateInfo{}
	for rows.Next() {
		item := &DateInfo{}
		if err := rows.Scan(database.GetScanFields(item, fieldNames)...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		allDayInfos = append(allDayInfos, item)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to query date info by date range: rows.Err")
	}

	exportCols := []exporter.ExportColumnMap{
		{
			DBColumn: "date",
		},
		{
			DBColumn: "location_id",
		},
		{
			DBColumn: "day_type_id",
		},
		{
			DBColumn: "opening_time",
		},
		{
			DBColumn: "time_zone",
		},
		{
			DBColumn: "status",
		},
	}
	exportableDayInfos := sliceutils.Map(allDayInfos, func(d *DateInfo) database.Entity {
		return d
	})
	str, err := exporter.ExportBatch(exportableDayInfos, exportCols)
	if err != nil {
		return nil, err
	}
	return exporter.ToCSV(str), nil
}
