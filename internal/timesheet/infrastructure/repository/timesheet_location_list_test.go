package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TimesheetLocationListRepoWithSqlMock() (TimesheetLocationListRepoImpl, *testutil.MockDB) {
	mockDB := testutil.NewMockDB()
	repo := TimesheetLocationListRepoImpl{}

	return repo, mockDB
}

type timesheetLocation struct {
	LocationID       pgtype.Text
	Name             pgtype.Text
	IsConfirmed      pgtype.Bool
	DraftCount       pgtype.Int8
	SubmittedCount   pgtype.Int8
	ApprovedCount    pgtype.Int8
	ConfirmedCount   pgtype.Int8
	UnconfirmedCount pgtype.Int8
}

type nonConfirmedLocationCount struct {
	Count pgtype.Int4
}

func (t *timesheetLocation) FieldMap() ([]string, []interface{}) {
	return []string{
			"location_id", "name", "is_confirmed", "draft_count", "submitted_count", "approved_count", "confirmed_count", "unconfirmed_count",
		}, []interface{}{
			&t.LocationID, &t.Name, &t.IsConfirmed, &t.DraftCount, &t.SubmittedCount, &t.ApprovedCount, &t.ConfirmedCount, &t.UnconfirmedCount,
		}
}

func (t *nonConfirmedLocationCount) TableName() string {
	return "get_non_confirmed_locations"
}

func (t *nonConfirmedLocationCount) FieldMap() ([]string, []interface{}) {
	return []string{
			"count",
		}, []interface{}{
			&t.Count,
		}
}

func (t *timesheetLocation) TableName() string {
	return "location_timesheets_non_confirmed_count_v3"
}

func TestTimesheetLocationListRepoImpl_GetTimesheetLocationList(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	selectFields := []string{"location_id", "name", "draft_count", "submitted_count", "approved_count", "confirmed_count", "unconfirmed_count", "is_confirmed"}
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(selectFields))...)
	repo, mockDB := TimesheetLocationListRepoWithSqlMock()

	mockGetTimesheetLocationListReq := &dto.GetTimesheetLocationListReq{
		FromDate: time.Now(),
		ToDate:   time.Now().Add(time.Hour),
		Keyword:  "keyword",
		Limit:    10,
		Offset:   0,
	}

	timesheetLocationMockOne := &timesheetLocation{
		LocationID:       database.Text("location_0"),
		Name:             database.Text("name_0"),
		IsConfirmed:      database.Bool(false),
		DraftCount:       database.Int8(20),
		SubmittedCount:   database.Int8(21),
		ApprovedCount:    database.Int8(22),
		ConfirmedCount:   database.Int8(23),
		UnconfirmedCount: database.Int8(24),
	}

	t.Run("failed to select timesheet location record tx closed", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		retrieveLocations, err := repo.GetTimesheetLocationList(ctx, mockDB.DB, mockGetTimesheetLocationListReq)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Equal(t, fmt.Errorf("%w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, retrieveLocations)
		mock.AssertExpectationsForObjects(t, mockDB.DB)

	})
	t.Run("failed to select timesheet location record tx closed no rows affected", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		retrieveLocations, err := repo.GetTimesheetLocationList(ctx, mockDB.DB, mockGetTimesheetLocationListReq)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("%w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, retrieveLocations)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("success retrieving single timesheet location record", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything)

		value := database.GetScanFields(timesheetLocationMockOne, selectFields)

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			value,
		})

		retrieveLocations, err := repo.GetTimesheetLocationList(ctx, mockDB.DB, mockGetTimesheetLocationListReq)
		assert.Nil(t, err)
		assert.Equal(t, []*dto.TimesheetLocation{
			{
				LocationID:       "location_0",
				Name:             "name_0",
				IsConfirmed:      false,
				DraftCount:       20,
				SubmittedCount:   21,
				ApprovedCount:    22,
				ConfirmedCount:   23,
				UnconfirmedCount: 24,
			},
		}, retrieveLocations)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
	})
}

func TestTimesheetLocationListRepoImpl_GetNonConfirmedLocationCount(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	selectFields := []string{"count"}
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(selectFields))...)

	mockPeriodDate := &time.Time{}

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := TimesheetLocationListRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Row, nil)
		mockDB.Row.On("Scan", args...).Once().Return(nil)
		retrieveCount, err := repo.GetNonConfirmedLocationCount(ctx, mockDB.DB, *mockPeriodDate)
		assert.Nil(t, err)
		assert.NotNil(t, retrieveCount)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("query row error", func(t *testing.T) {
		repo, mockDB := TimesheetLocationListRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Row)
		mockDB.Row.On("Scan", args...).Once().Return(puddle.ErrClosedPool)
		prefecture, err := repo.GetNonConfirmedLocationCount(ctx, mockDB.DB, *mockPeriodDate)
		assert.Nil(t, prefecture)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
