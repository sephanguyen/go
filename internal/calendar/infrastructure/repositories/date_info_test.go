package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func DateInfoRepoWithSqlMock() (*DateInfoRepo, *testutil.MockDB) {
	mockDB := testutil.NewMockDB()

	mockRepo := &DateInfoRepo{}
	return mockRepo, mockDB
}

func genSliceMock(n int) []interface{} {
	result := []interface{}{}
	for i := 0; i < n; i++ {
		result = append(result, mock.Anything)
	}
	return result
}

func TestDateInfoRepo_GetDateInfoByDateAndLocationID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	dateInfoRepo, mockDB := DateInfoRepoWithSqlMock()

	mockDateInfo := &DateInfo{}
	fields, values := mockDateInfo.FieldMap()
	now := time.Now()
	locationID := "1"

	t.Run("Fetch date info by date and location ID successful", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &now, &locationID)
		mockDB.MockRowScanFields(nil, fields, values)
		dateInfos, err := dateInfoRepo.GetDateInfoByDateAndLocationID(ctx, mockDB.DB, now, locationID)
		assert.NoError(t, err)
		assert.NotNil(t, dateInfos)
	})

	t.Run("Fetch date info by date and location ID failed", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &now, &locationID)
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		dateInfos, err := dateInfoRepo.GetDateInfoByDateAndLocationID(ctx, mockDB.DB, now, locationID)

		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, dateInfos)
	})
}

func TestDateInfoRepo_GetDateInfoByDateRangeAndLocationID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	dateInfoRepo, mockDB := DateInfoRepoWithSqlMock()

	mockDateInfo := &DateInfo{}
	fields, values := mockDateInfo.FieldMap()
	startDate := time.Now().Add(-48 * time.Hour)
	endDate := time.Now().Add(48 * time.Hour)
	locationID := "1"

	t.Run("Fetch date info by date range and location ID successful", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &startDate, &endDate, &locationID)
		mockDB.MockScanFields(nil, fields, values)

		dateInfos, err := dateInfoRepo.GetDateInfoByDateRangeAndLocationID(ctx, mockDB.DB, startDate, endDate, locationID)
		assert.NoError(t, err)
		assert.NotNil(t, dateInfos)
	})

	t.Run("Fetch date info by date range and location ID failed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &startDate, &endDate, &locationID)
		dateInfos, err := dateInfoRepo.GetDateInfoByDateRangeAndLocationID(ctx, mockDB.DB, startDate, endDate, locationID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, dateInfos)
	})
}

func TestDateInfoRepo_GetDateInfoDetailedByDateRangeAndLocationID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	dateInfoRepo, mockDB := DateInfoRepoWithSqlMock()

	var displayName pgtype.Text
	mockDateInfo := &DateInfo{}
	fields, values := mockDateInfo.ExportFieldMap()
	fields = append(fields, "name")
	values = append(values, &displayName)

	startDate := time.Now().Add(-48 * time.Hour)
	endDate := time.Now().Add(48 * time.Hour)
	locationID := "1"
	timezone := "sample-timezone"

	t.Run("Fetch date info by date range detailed and location ID successful", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &startDate, &endDate, &locationID, &timezone)
		mockDB.MockScanFields(nil, fields, values)

		dateInfos, err := dateInfoRepo.GetDateInfoDetailedByDateRangeAndLocationID(ctx, mockDB.DB, startDate, endDate, locationID, timezone)
		assert.NoError(t, err)
		assert.NotNil(t, dateInfos)
	})

	t.Run("Fetch date info by date range detailed and location ID failed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &startDate, &endDate, &locationID, &timezone)
		dateInfos, err := dateInfoRepo.GetDateInfoDetailedByDateRangeAndLocationID(ctx, mockDB.DB, startDate, endDate, locationID, timezone)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, dateInfos)
	})
}

func TestDateInfoRepo_UpsertDateInfo(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDateInfo := &DateInfo{}
	fields := database.GetFieldNames(mockDateInfo)
	now := time.Now()

	args := append([]interface{}{
		mock.Anything,
		mock.AnythingOfType("string")},
		genSliceMock(len(fields))...)

	params := &dto.UpsertDateInfoParams{
		DateInfo: &dto.DateInfo{
			Date:        now,
			LocationID:  "sample-loc-1",
			DateTypeID:  "regular",
			OpeningTime: "09:00",
			Status:      "draft",
			TimeZone:    "sample-timezone",
		},
	}

	t.Run("upsert failed", func(t *testing.T) {
		mockDateInfoRepo, mockDB := DateInfoRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, pgx.ErrTxClosed)

		err := mockDateInfoRepo.UpsertDateInfo(ctx, mockDB.DB, params)

		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("no rows affected after upsert", func(t *testing.T) {
		mockDateInfoRepo, mockDB := DateInfoRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := mockDateInfoRepo.UpsertDateInfo(ctx, mockDB.DB, params)

		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("error on upsert date info, rows affected is not 1").Error(), err.Error())
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("upsert successful", func(t *testing.T) {
		mockDateInfoRepo, mockDB := DateInfoRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := mockDateInfoRepo.UpsertDateInfo(ctx, mockDB.DB, params)

		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestDateInfoRepo_DuplicateDateInfo(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	now := time.Now()

	params := &dto.DuplicateDateInfoParams{
		DateInfo: &dto.DateInfo{
			Date:        now,
			LocationID:  "sample-loc-1",
			DateTypeID:  "regular",
			OpeningTime: "09:00",
			Status:      "draft",
			TimeZone:    "sample-timezone",
		},
		Dates: []time.Time{
			now,
			now,
		},
	}

	t.Run("bulk upsert failed", func(t *testing.T) {
		mockDateInfoRepo, mockDB := DateInfoRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		batchResults.On("Exec").Return(cmdTag, errors.New("error")).Once()
		batchResults.On("Close").Once().Return(nil)

		err := mockDateInfoRepo.DuplicateDateInfo(ctx, mockDB.DB, params)

		assert.NotNil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, batchResults)
	})

	t.Run("no rows affected after upsert", func(t *testing.T) {
		mockDateInfoRepo, mockDB := DateInfoRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`0`))

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		batchResults.On("Exec").Return(cmdTag, nil).Once()
		batchResults.On("Close").Once().Return(nil)

		err := mockDateInfoRepo.DuplicateDateInfo(ctx, mockDB.DB, params)

		assert.NotNil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, batchResults)
	})

	t.Run("bulk upsert successful", func(t *testing.T) {
		mockDateInfoRepo, mockDB := DateInfoRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		for i := 0; i < len(params.Dates); i++ {
			batchResults.On("Exec").Return(cmdTag, nil).Once()
		}
		batchResults.On("Close").Once().Return(nil)

		err := mockDateInfoRepo.DuplicateDateInfo(ctx, mockDB.DB, params)

		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, batchResults)
	})
}
func TestDateInfoRepo_GetAllToExport(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	dateInfoRepo, mockDB := DateInfoRepoWithSqlMock()

	mockDateInfo := &DateInfo{}
	fields, values := mockDateInfo.ExportFieldMap()

	t.Run("Get all date info successful", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)
		mockDB.MockScanFields(nil, fields, values)
		dateInfos, err := dateInfoRepo.GetAllToExport(ctx, mockDB.DB)
		assert.NoError(t, err)
		assert.NotNil(t, dateInfos)
	})

	t.Run("Fetch date info by date range and location ID failed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything)
		dateInfos, err := dateInfoRepo.GetAllToExport(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, dateInfos)
	})
}
