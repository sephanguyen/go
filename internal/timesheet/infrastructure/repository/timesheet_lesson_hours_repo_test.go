package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TimesheetLessonHoursRepoWithSqlMock() (TimesheetLessonHoursRepoImpl, *testutil.MockDB) {
	mockDB := testutil.NewMockDB()
	repo := TimesheetLessonHoursRepoImpl{}

	return repo, mockDB
}

func TestTimesheetLessonHourRepoImpl_FindTimesheetLessonHoursByTimesheetID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	timesheetMockNotExistRecord := &entity.Timesheet{
		TimesheetID: database.Text("not-exist"),
	}

	timesheetMockRecord := &entity.Timesheet{
		TimesheetID: database.Text("10"),
	}

	timesheetLessonHourMockOne := &entity.TimesheetLessonHours{
		TimesheetID: database.Text("10"),
		LessonID:    database.Text("1"),
	}

	selectFields := []string{"timesheet_id", "lesson_id", "flag_on", "created_at", "updated_at", "deleted_at"}
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(selectFields))...)
	repo, mockDB := TimesheetLessonHoursRepoWithSqlMock()

	t.Run("failed to select timesheet lesson hour record tx closed", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		retrieveLessons, err := repo.FindTimesheetLessonHoursByTimesheetID(ctx, mockDB.DB, timesheetMockNotExistRecord.TimesheetID)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, retrieveLessons)
		mock.AssertExpectationsForObjects(t, mockDB.DB)

	})
	t.Run("failed to select timesheet lesson hour record tx closed no rows affected", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		retrieveLessons, err := repo.FindTimesheetLessonHoursByTimesheetID(ctx, mockDB.DB, timesheetMockNotExistRecord.TimesheetID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, retrieveLessons)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("success retrieving single timesheet lesson hour record", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything)

		value := database.GetScanFields(timesheetLessonHourMockOne, selectFields)

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			value,
		})

		timesheetLessonRecords, err := repo.FindTimesheetLessonHoursByTimesheetID(ctx, mockDB.DB, timesheetMockRecord.TimesheetID)
		assert.Nil(t, err)
		assert.Equal(t, []*entity.TimesheetLessonHours{
			{
				TimesheetID: database.Text("10"),
				LessonID:    database.Text("1"),
			},
		}, timesheetLessonRecords)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
	})

	t.Run("success retrieving multiple timesheet lesson hour records", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything)
		var timesheetLessonRecords []*entity.TimesheetLessonHours

		timesheetLessonRecords = append(timesheetLessonRecords, timesheetLessonHourMockOne)
		valueOne := database.GetScanFields(timesheetLessonHourMockOne, selectFields)

		timesheetLessonHourMockTwo := &entity.TimesheetLessonHours{
			TimesheetID: database.Text("10"),
			LessonID:    database.Text("2"),
		}

		timesheetLessonRecords = append(timesheetLessonRecords, timesheetLessonHourMockTwo)

		valueTwo := database.GetScanFields(timesheetLessonHourMockTwo, selectFields)

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			valueOne,
			valueTwo,
		})

		retrieveLessons, err := repo.FindTimesheetLessonHoursByTimesheetID(ctx, mockDB.DB, timesheetMockRecord.TimesheetID)
		assert.Nil(t, err)
		assert.Equal(t, retrieveLessons, timesheetLessonRecords)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
	})
}

func TestTimesheetLessonHourRepoImpl_FindByTimesheetIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := TimesheetLessonHoursRepoWithSqlMock()
	timesheetIDs := []string{"timesheetID_1", "timesheetID_2"}
	lessonIDs := []string{"lessonID_1", "lessonID_2"}
	now := time.Now()
	timesheetLessonHoursEntities := []*entity.TimesheetLessonHours{
		{
			TimesheetID: database.Text(timesheetIDs[0]),
			LessonID:    database.Text(lessonIDs[0]),
			CreatedAt:   database.Timestamptz(now),
			UpdatedAt:   database.Timestamptz(now),
		},
		{
			TimesheetID: database.Text(timesheetIDs[1]),
			LessonID:    database.Text(lessonIDs[1]),
			CreatedAt:   database.Timestamptz(now),
			UpdatedAt:   database.Timestamptz(now),
		},
	}
	selectFields := database.GetFieldNames(timesheetLessonHoursEntities[0])
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(selectFields))...)

	testCases := []struct {
		name         string
		request      []string
		expectErr    error
		expectedResp interface{}
		setup        func()
	}{
		{
			name:         "find by timesheetIDs success",
			request:      timesheetIDs,
			expectErr:    nil,
			expectedResp: timesheetLessonHoursEntities,
			setup: func() {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything)

				mockDB.MockScanArray(nil, selectFields, [][]interface{}{
					database.GetScanFields(timesheetLessonHoursEntities[0], selectFields),
					database.GetScanFields(timesheetLessonHoursEntities[1], selectFields),
				})
			},
		},
		{
			name:         "find by timesheetIDs failed",
			request:      timesheetIDs,
			expectErr:    fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed),
			expectedResp: []*entity.TimesheetLessonHours(nil),
			setup: func() {
				mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
	}
	for _, testcase := range testCases {
		testName := fmt.Sprintf("TestCase: %s", testcase.name)
		t.Run(testName, func(t *testing.T) {
			testcase.setup()
			resp, err := repo.FindByTimesheetIDs(ctx, mockDB.DB, testcase.request)
			assert.Equal(t, testcase.expectErr, err)
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestTimesheetLessonHourRepoImpl_InsertMultiple(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := TimesheetLessonHoursRepoWithSqlMock()
	timesheetIDs := []string{"timesheetID_1", "timesheetID_2"}
	lessonIDs := []string{"lessonID_1", "lessonID_2"}
	timesheetLessonHoursEntities := []*entity.TimesheetLessonHours{
		{
			TimesheetID: database.Text(timesheetIDs[0]),
			LessonID:    database.Text(lessonIDs[0]),
		},
		{
			TimesheetID: database.Text(timesheetIDs[1]),
			LessonID:    database.Text(lessonIDs[1]),
		},
	}
	testCases := []struct {
		name         string
		request      interface{}
		expectErr    error
		expectedResp interface{}
		setup        func()
	}{
		{
			name:         "insert timesheet lesson hours success",
			request:      timesheetLessonHoursEntities,
			expectErr:    nil,
			expectedResp: timesheetLessonHoursEntities,
			setup: func() {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name:         "insert timesheet lesson hours failed send batch error",
			request:      timesheetLessonHoursEntities,
			expectErr:    puddle.ErrClosedPool,
			expectedResp: []*entity.TimesheetLessonHours(nil),
			setup: func() {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name:         "insert timesheet lesson hours failed row affected different error",
			request:      timesheetLessonHoursEntities,
			expectErr:    fmt.Errorf("err insert TimesheetLessonHours: %d RowsAffected", 0),
			expectedResp: []*entity.TimesheetLessonHours(nil),
			setup: func() {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`0`))
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testcase := range testCases {
		testName := fmt.Sprintf("TestCase: %s", testcase.name)
		t.Run(testName, func(t *testing.T) {
			testcase.setup()
			resp, err := repo.InsertMultiple(ctx, mockDB.DB, testcase.request.([]*entity.TimesheetLessonHours))
			assert.Equal(t, testcase.expectErr, err)
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestTimesheetLessonHourRepoImpl_UpdateAutoCreateFlagStateAfterTime(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	timesheetIDs := []string{"test-id"}
	timeNow := time.Now()
	flagOn := true

	mockE := &entity.TimesheetLessonHours{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := TimesheetLessonHoursRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.UpdateAutoCreateFlagStateAfterTime(ctx, mockDB.DB, timesheetIDs, timeNow, flagOn)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("update timesheet lesson hours record fail", func(t *testing.T) {
		repo, mockDB := TimesheetLessonHoursRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.UpdateAutoCreateFlagStateAfterTime(ctx, mockDB.DB, timesheetIDs, timeNow, flagOn)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update TimesheetLessonHoursRepoImpl: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestTimesheetLessonHourRepoImpl_UpdateTimesheetLessonAutoCreateFlagByTimesheetIDs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	timesheetIDs := []string{"test-id"}
	flagOn := true

	mockE := &entity.TimesheetLessonHours{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := TimesheetLessonHoursRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.UpdateTimesheetLessonAutoCreateFlagByTimesheetIDs(ctx, mockDB.DB, timesheetIDs, flagOn)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("update timesheet lesson hours record fail", func(t *testing.T) {
		repo, mockDB := TimesheetLessonHoursRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.UpdateTimesheetLessonAutoCreateFlagByTimesheetIDs(ctx, mockDB.DB, timesheetIDs, flagOn)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update TimesheetLessonHoursRepoImpl::UpdateTimesheetLessonAutoCreateFlagByTimesheetIDs: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestTimesheetLessonHoursRepoImpl_FindTimesheetLessonHoursByLessonIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := TimesheetLessonHoursRepoWithSqlMock()
	timesheetIDs := []string{"timesheetID_1", "timesheetID_2"}
	lessonIDs := []string{"lessonID_1", "lessonID_2"}
	now := time.Now()
	timesheetLessonHoursEntities := []*entity.TimesheetLessonHours{
		{
			TimesheetID: database.Text(timesheetIDs[0]),
			LessonID:    database.Text(lessonIDs[0]),
			CreatedAt:   database.Timestamptz(now),
			UpdatedAt:   database.Timestamptz(now),
		},
		{
			TimesheetID: database.Text(timesheetIDs[1]),
			LessonID:    database.Text(lessonIDs[1]),
			CreatedAt:   database.Timestamptz(now),
			UpdatedAt:   database.Timestamptz(now),
		},
	}
	selectFields := database.GetFieldNames(timesheetLessonHoursEntities[0])

	testCases := []struct {
		name         string
		request      []string
		expectErr    error
		expectedResp interface{}
		setup        func()
	}{
		{
			name:         "find by lessonIDs success",
			request:      lessonIDs,
			expectErr:    nil,
			expectedResp: timesheetLessonHoursEntities,
			setup: func() {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything)

				mockDB.MockScanArray(nil, selectFields, [][]interface{}{
					database.GetScanFields(timesheetLessonHoursEntities[0], selectFields),
					database.GetScanFields(timesheetLessonHoursEntities[1], selectFields),
				})
			},
		},
	}
	for _, testcase := range testCases {
		testName := fmt.Sprintf("TestCase: %s", testcase.name)
		t.Run(testName, func(t *testing.T) {
			testcase.setup()
			resp, err := repo.FindTimesheetLessonHoursByLessonIDs(ctx, mockDB.DB, testcase.request)
			assert.Equal(t, testcase.expectErr, err)
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestOtherWorkingHoursRepoImpl_MapExistingLessonHoursByTimesheetIds(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := TimesheetLessonHoursRepoWithSqlMock()
	timesheetID := idutil.ULIDNow()

	testCases := []struct {
		name         string
		setup        func()
		expectErr    error
		expectedResp map[string]struct{}
	}{
		{
			name:         "happy case",
			expectErr:    nil,
			expectedResp: map[string]struct{}{timesheetID: {}},
			setup: func() {
				mockDB.MockQueryArgs(t, nil,
					mock.Anything,
					mock.Anything,
					[]string{timesheetID},
				)

				mockDB.MockScanArray(nil, []string{"timesheet_id"}, [][]interface{}{{&timesheetID}})
			},
		},
		{
			name:         "err query",
			expectErr:    pgx.ErrNoRows,
			expectedResp: nil,
			setup: func() {
				mockDB.MockQueryArgs(t, pgx.ErrNoRows,
					mock.Anything,
					mock.Anything,
					[]string{timesheetID},
				)

				mockDB.MockScanArray(nil, []string{"timesheet_id"}, [][]interface{}{{&timesheetID}})
			},
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			testcase.setup()
			resp, err := repo.MapExistingLessonHoursByTimesheetIds(ctx, mockDB.DB, []string{timesheetID})

			assert.Equal(t, testcase.expectErr, err)
			assert.Equal(t, testcase.expectedResp, resp)

		})
	}
}
