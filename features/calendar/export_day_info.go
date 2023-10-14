package calendar

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	cal_repo "github.com/manabie-com/backend/internal/calendar/infrastructure/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"

	"github.com/pkg/errors"
)

func (s *suite) exportDayInfo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	time.Sleep(5 * time.Second)

	req := &cpb.ExportDayInfoRequest{}
	stepState.Response, stepState.ResponseErr = cpb.NewDateInfoReaderServiceClient(s.CalendarConn).
		ExportDayInfo(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsDayInfosInCsv(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("can not export date info: %s", stepState.ResponseErr.Error())
	}
	resp := stepState.Response.(*cpb.ExportDayInfoResponse)

	expectedDateInfo, err := s.getAllDayInfoExport(ctx)
	if err != nil {
		return ctx, fmt.Errorf("can not get expected date info: %s", err)
	}
	if string(expectedDateInfo) != string(resp.GetData()) {
		return ctx, fmt.Errorf("date info csv is not valid:\ngot:\n%s \nexpected: \n%s", string(resp.Data), expectedDateInfo)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getAllDayInfoExport(ctx context.Context) ([]byte, error) {
	stepState := StepStateFromContext(ctx)

	query := `SELECT %s 
		FROM %s 
		WHERE resource_path = $1
		AND deleted_at is NULL`
	dateInfo := cal_repo.DateInfo{}
	fieldNames, _ := dateInfo.ExportFieldMap()
	query = fmt.Sprintf(
		query,
		strings.Join(fieldNames, ","),
		dateInfo.TableName(),
	)

	rows, err := s.CalendarDB.Query(ctx, query, database.Text(strconv.Itoa(int(stepState.CurrentSchoolID))))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all date info: db.Query")
	}
	defer rows.Close()

	allDayInfos := []*cal_repo.DateInfo{}
	for rows.Next() {
		item := &cal_repo.DateInfo{}
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
	exportableDayInfos := sliceutils.Map(allDayInfos, func(d *cal_repo.DateInfo) database.Entity {
		return d
	})
	str, err := exporter.ExportBatch(exportableDayInfos, exportCols)
	if err != nil {
		return nil, err
	}
	return exporter.ToCSV(str), nil
}
