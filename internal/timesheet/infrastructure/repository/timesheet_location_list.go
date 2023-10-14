package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"

	"github.com/jackc/pgtype"
)

type TimesheetLocationListRepoImpl struct {
}

func (r *TimesheetLocationListRepoImpl) GetTimesheetLocationList(ctx context.Context, db database.QueryExecer, req *dto.GetTimesheetLocationListReq) ([]*dto.TimesheetLocation, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetLocationListRepoImpl.GetTimesheetLocationList")
	defer span.End()

	fromDate := database.Timestamptz(req.FromDate)
	toDate := database.Timestamptz(req.ToDate)
	keyword := database.Text(req.Keyword)
	limit := req.Limit
	offset := req.Offset

	result := []*dto.TimesheetLocation{}

	stmt := fmt.Sprintf(`SELECT 
			location_id,
			name,
			draft_count,
			submitted_count,
			approved_count,
			confirmed_count,
			unconfirmed_count,
			is_confirmed 
		FROM location_timesheets_non_confirmed_count_v3($1,$2,$3)
		ORDER BY is_confirmed, unconfirmed_count DESC, location_id
		LIMIT %d 
		OFFSET %d
		`, limit, offset)

	rows, err := db.Query(ctx, stmt, keyword, fromDate, toDate)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var (
			locationID, name                                                            pgtype.Text
			draftCount, submittedCount, approvedCount, confirmedCount, unconfirmedCount pgtype.Int8
			isConfirmed                                                                 pgtype.Bool
		)
		if err := rows.Scan(
			&locationID,
			&name,
			&draftCount,
			&submittedCount,
			&approvedCount,
			&confirmedCount,
			&unconfirmedCount,
			&isConfirmed); err != nil {
			return nil, err
		}
		result = append(result, &dto.TimesheetLocation{
			LocationID:       locationID.String,
			Name:             name.String,
			DraftCount:       int32(draftCount.Int),
			SubmittedCount:   int32(submittedCount.Int),
			ApprovedCount:    int32(approvedCount.Int),
			ConfirmedCount:   int32(confirmedCount.Int),
			UnconfirmedCount: int32(unconfirmedCount.Int),
			IsConfirmed:      isConfirmed.Bool,
		})
	}

	return result, nil
}

func (r *TimesheetLocationListRepoImpl) GetTimesheetLocationCount(ctx context.Context, db database.QueryExecer, keyword string) (*dto.TimesheetLocationAggregate, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetLocationListRepoImpl.GetTimesheetLocationCount")
	defer span.End()

	result := &dto.TimesheetLocationAggregate{}
	keywordArg := database.Text(keyword)
	stmt := `SELECT COUNT(*) FROM locations WHERE deleted_at IS NULL AND name ILIKE $1 AND
        location_type NOT IN (
        SELECT distinct parent_location_type_id
        FROM location_types
        WHERE
            parent_location_type_id IS NOT NULL AND
            deleted_at IS NULL AND is_archived = false)`

	err := db.QueryRow(ctx, stmt, keywordArg).Scan(&result.Count)
	if err != nil {
		return nil, err
	}

	return result, err
}

func (r *TimesheetLocationListRepoImpl) GetNonConfirmedLocationCount(ctx context.Context, db database.QueryExecer, periodDate time.Time) (*dto.GetNonConfirmedLocationCountOut, error) {
	ctx, span := interceptors.StartSpan(ctx, "TimesheetLocationListRepoImpl.GetNonConfirmedLocationCount")
	defer span.End()

	result := &dto.GetNonConfirmedLocationCountOut{}

	periodDateArg := database.Timestamptz(periodDate)
	stmt := "SELECT COUNT(*) FROM get_non_confirmed_locations($1)"

	err := db.QueryRow(ctx, stmt, periodDateArg).Scan(&result.NonconfirmedCount)
	if err != nil {
		return nil, err
	}

	return result, err
}
