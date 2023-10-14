package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TimesheetRepoWithSqlMock() (TimesheetRepoImpl, *testutil.MockDB) {
	mockDB := testutil.NewMockDB()
	repo := TimesheetRepoImpl{}

	return repo, mockDB
}

func TestTimesheet_InsertTimeSheet(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := TimesheetRepoWithSqlMock()

	timesheet := &entity.Timesheet{TimesheetID: database.Text(idutil.ULIDNow())}
	_, timesheetValues := timesheet.FieldMap()
	argsTimesheet := append(
		[]interface{}{mock.Anything, mock.Anything},
		genSliceMock(len(timesheetValues))...,
	)
	internalErr := errors.New(" internal server error")
	testCases := []struct {
		name         string
		setup        func()
		expectErr    error
		expectedResp *entity.Timesheet
	}{
		{
			name:         "happy case",
			expectErr:    nil,
			expectedResp: timesheet,
			setup: func() {
				cmtTag := pgconn.CommandTag(`1`)
				mockDB.DB.On("Exec", argsTimesheet...).Once().Return(cmtTag, nil)
			},
		},
		{
			name:         "error case fail to insert timesheet internal server error",
			expectErr:    fmt.Errorf("err insert Timesheet: %w", internalErr),
			expectedResp: nil,
			setup: func() {
				cmtTag := pgconn.CommandTag(`0`)
				mockDB.DB.On("Exec", argsTimesheet...).Once().Return(cmtTag, internalErr)
			},
		},
		{
			name:         "error case row affected different one",
			expectErr:    fmt.Errorf("err insert Timesheet: %d RowsAffected", 0),
			expectedResp: nil,
			setup: func() {
				cmtTag := pgconn.CommandTag(`0`)
				mockDB.DB.On("Exec", argsTimesheet...).Once().Return(cmtTag, nil)
			},
		},
	}

	for _, testcase := range testCases {
		testName := fmt.Sprintf("TestCase: %s", testcase.name)
		t.Run(testName, func(t *testing.T) {
			testcase.setup()
			resp, err := repo.InsertTimeSheet(ctx, mockDB.DB, timesheet)
			assert.Equal(t, testcase.expectErr, err)
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestTimesheet_UpdateTimeSheet(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := TimesheetRepoWithSqlMock()

	timesheet := &entity.Timesheet{}
	_, timesheetValues := timesheet.FieldMap()
	argsTimesheet := append(
		[]interface{}{mock.Anything, mock.Anything},
		genSliceMock(len(timesheetValues))...,
	)
	internalErr := errors.New(" internal server error")
	testCases := []struct {
		name         string
		setup        func()
		expectErr    error
		expectedResp *entity.Timesheet
	}{
		{
			name:         "happy case",
			expectErr:    nil,
			expectedResp: timesheet,
			setup: func() {
				cmtTag := pgconn.CommandTag(`1`)
				mockDB.DB.On("Exec", argsTimesheet...).Return(cmtTag, nil).Once()
			},
		},
		{
			name:      "error case fail to insert timesheet internal server error",
			expectErr: fmt.Errorf("err update Timesheet: %w", internalErr),
			setup: func() {
				cmtTag := pgconn.CommandTag(`0`)
				mockDB.DB.On("Exec", argsTimesheet...).Once().Return(cmtTag, internalErr)
			},
		},
		{
			name:      "error case row affected different one",
			expectErr: fmt.Errorf("err update Timesheet: %d RowsAffected", 0),
			setup: func() {
				cmtTag := pgconn.CommandTag(`0`)
				mockDB.DB.On("Exec", argsTimesheet...).Once().Return(cmtTag, nil)
			},
		},
	}

	for _, testcase := range testCases {
		testName := fmt.Sprintf("TestCase: %s", testcase.name)
		t.Run(testName, func(t *testing.T) {
			testcase.setup()
			resp, err := repo.UpdateTimeSheet(ctx, mockDB.DB, timesheet)
			assert.Equal(t, testcase.expectErr, err)
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestTimesheet_Retrieve(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := TimesheetRepoWithSqlMock()

	timesheet := &entity.Timesheet{TimesheetID: database.Text(idutil.ULIDNow())}

	respValues := [][]byte{}
	respValues = append(respValues, []byte(fmt.Sprintf("%v", timesheet)))
	testCases := []struct {
		name         string
		setup        func()
		expectErr    error
		expectedResp []*entity.Timesheet
	}{
		{
			name:         "happy case",
			expectErr:    nil,
			expectedResp: []*entity.Timesheet{timesheet},
			setup: func() {
				mockDB.MockQueryArgs(t, nil, mock.Anything,
					mock.Anything,
					database.TextArray([]string{timesheet.TimesheetID.String}),
				)

				fields, values := timesheet.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
		{
			name:         "err query",
			expectErr:    fmt.Errorf("err db.Query: %w", pgx.ErrNoRows),
			expectedResp: nil,
			setup: func() {
				mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything,
					mock.Anything,
					database.TextArray([]string{timesheet.TimesheetID.String}),
				)

				fields, values := timesheet.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
	}

	for _, testcase := range testCases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			testcase.setup()
			resp, err := repo.Retrieve(ctx, mockDB.DB, database.TextArray([]string{timesheet.TimesheetID.String}))

			assert.Equal(t, testcase.expectErr, err)
			assert.Equal(t, testcase.expectedResp, resp)

		})
	}
}

func TestTimesheet_FindTimesheetByTimesheetID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := TimesheetRepoWithSqlMock()

	timesheet := &entity.Timesheet{TimesheetID: database.Text(idutil.ULIDNow())}

	respValues := [][]byte{}
	respValues = append(respValues, []byte(fmt.Sprintf("%v", timesheet)))
	testCases := []struct {
		name         string
		setup        func()
		expectErr    error
		expectedResp interface{}
	}{
		{
			name:         "happy case",
			expectErr:    nil,
			expectedResp: timesheet,
			setup: func() {
				mockDB.MockQueryArgs(t, nil, mock.Anything,
					mock.Anything,
					database.TextArray([]string{timesheet.TimesheetID.String}),
				)

				fields, values := timesheet.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
		{
			name:         "error query",
			expectErr:    fmt.Errorf("err db.Query: %s, timesheet_id: %s", pgx.ErrNoRows.Error(), timesheet.TimesheetID.String),
			expectedResp: (*entity.Timesheet)(nil),
			setup: func() {
				mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything,
					mock.Anything,
					database.TextArray([]string{timesheet.TimesheetID.String}),
				)
				fields, values := timesheet.FieldMap()
				mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{
					values,
				})
			},
		},
	}

	for _, testcase := range testCases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			testcase.setup()
			timesheet, err := repo.FindTimesheetByTimesheetID(ctx, mockDB.DB, timesheet.TimesheetID)

			assert.Equal(t, testcase.expectErr, err)
			assert.Equal(t, testcase.expectedResp, timesheet)
		})
	}
}

func TestTimesheet_FindTimesheetByTimesheetArgs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := TimesheetRepoWithSqlMock()

	timesheet := &entity.Timesheet{TimesheetID: database.Text(idutil.ULIDNow())}

	var respValues [][]byte
	respValues = append(respValues, []byte(fmt.Sprintf("%v", timesheet)))
	testCases := []struct {
		name         string
		setup        func()
		expectErr    error
		expectedResp []*entity.Timesheet
	}{
		{
			name:         "happy case",
			expectErr:    nil,
			expectedResp: []*entity.Timesheet{timesheet},
			setup: func() {
				mockDB.MockQueryArgs(t, nil, mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)

				fields, values := timesheet.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
		{
			name:      "err query",
			expectErr: fmt.Errorf("err db.Query: %w", pgx.ErrNoRows),
			setup: func() {
				mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)

				fields, values := timesheet.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
	}

	for _, testcase := range testCases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			testcase.setup()
			resp, err := repo.FindTimesheetByTimesheetArgs(ctx, mockDB.DB, &dto.TimesheetQueryArgs{})

			assert.Equal(t, testcase.expectErr, err)
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestTimesheet_GetStaffFutureTimesheetIDsWithLocations(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := TimesheetRepoWithSqlMock()

	timesheetID := idutil.ULIDNow()
	locationID := idutil.ULIDNow()
	values := []interface{}{&timesheetID, &locationID}

	testCases := []struct {
		name         string
		setup        func()
		expectErr    error
		expectedResp []dto.TimesheetLocationDto
	}{
		{
			name:         "happy case",
			expectErr:    nil,
			expectedResp: []dto.TimesheetLocationDto{{TimesheetID: timesheetID, LocationID: locationID}},
			setup: func() {
				mockDB.MockQueryArgs(t, nil,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)

				mockDB.MockScanArray(nil, []string{"timesheet_id", "location_id"}, [][]interface{}{
					values,
				})
			},
		},
		{
			name:      "err query",
			expectErr: fmt.Errorf("err db.Query: %w", pgx.ErrNoRows),
			setup: func() {
				mockDB.MockQueryArgs(t, pgx.ErrNoRows,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)

				mockDB.MockScanArray(nil, []string{"timesheet_id", "location_id"}, [][]interface{}{
					values,
				})
			},
		},
	}

	for _, testcase := range testCases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			testcase.setup()
			staffID := "test-id"
			dateNow := time.Now()
			locationIDs := []string{"location-id"}

			resp, err := repo.GetStaffFutureTimesheetIDsWithLocations(ctx, mockDB.DB, staffID, dateNow, locationIDs)

			assert.Equal(t, testcase.expectErr, err)
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestTimesheet_GetStaffTimesheetIDsAfterDateCanChange(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := TimesheetRepoWithSqlMock()

	timesheetID := idutil.ULIDNow()
	values := []interface{}{&timesheetID}
	testCases := []struct {
		name         string
		setup        func()
		expectErr    error
		expectedResp []string
	}{
		{
			name:         "happy case",
			expectErr:    nil,
			expectedResp: []string{timesheetID},
			setup: func() {
				mockDB.MockQueryArgs(t, nil,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)

				mockDB.MockScanArray(nil, []string{"timesheet_id"}, [][]interface{}{
					values,
				})
			},
		},
		{
			name:      "err query",
			expectErr: fmt.Errorf("err db.Query: %w", pgx.ErrNoRows),
			setup: func() {
				mockDB.MockQueryArgs(t, pgx.ErrNoRows,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)

				mockDB.MockScanArray(nil, []string{"timesheet_id"}, [][]interface{}{
					values,
				})
			},
		},
	}

	for _, testcase := range testCases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			testcase.setup()
			staffID := "test-id"
			dateNow := time.Now()

			resp, err := repo.GetStaffTimesheetIDsAfterDateCanChange(ctx, mockDB.DB, staffID, dateNow)

			assert.Equal(t, testcase.expectErr, err)
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func genSliceMock(n int) []interface{} {
	var result []interface{}
	for i := 0; i < n; i++ {
		result = append(result, mock.Anything)
	}
	return result
}

func TestTimesheet_SoftDeleteByIDs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	timesheetID := database.TextArray([]string{"test-id"})
	mockE := &entity.Timesheet{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := TimesheetRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.SoftDeleteByIDs(ctx, mockDB.DB, timesheetID)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("soft delete timesheet record fail", func(t *testing.T) {
		repo, mockDB := TimesheetRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.SoftDeleteByIDs(ctx, mockDB.DB, timesheetID)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err delete TimesheetRepoImpl: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after soft deleting timesheet record", func(t *testing.T) {
		repo, mockDB := TimesheetRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.SoftDeleteByIDs(ctx, mockDB.DB, timesheetID)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err delete TimesheetRepoImpl: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestTimesheet_RemoveTimesheetRemarkByTimesheetIDs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	timesheetIDs := []string{"test-id"}
	mockE := &entity.Timesheet{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := TimesheetRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.RemoveTimesheetRemarkByTimesheetIDs(ctx, mockDB.DB, timesheetIDs)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("remove timesheet remark fail", func(t *testing.T) {
		repo, mockDB := TimesheetRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.RemoveTimesheetRemarkByTimesheetIDs(ctx, mockDB.DB, timesheetIDs)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err RemoveTimesheetRemarkByTimesheetIDs: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after remove timesheet remarks record", func(t *testing.T) {
		repo, mockDB := TimesheetRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.RemoveTimesheetRemarkByTimesheetIDs(ctx, mockDB.DB, timesheetIDs)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err RemoveTimesheetRemarkByTimesheetIDs: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestTimesheetRepo_UpsertMultiple(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := TimesheetRepoWithSqlMock()
	timesheetIDs := []string{"timesheetID_1", "timesheetID_2"}
	staffIDs := []string{"staffID_1", "staffID_2"}
	timesheetEntities := []*entity.Timesheet{
		{
			TimesheetID: database.Text(timesheetIDs[0]),
			StaffID:     database.Text(staffIDs[0]),
		},
		{
			StaffID: database.Text(staffIDs[1]),
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
			name:         "err upsert multiple timesheet success",
			request:      timesheetEntities,
			expectErr:    nil,
			expectedResp: timesheetEntities,
			setup: func() {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name:         "err upsert multiple timesheet failed send batch error",
			request:      timesheetEntities,
			expectErr:    puddle.ErrClosedPool,
			expectedResp: []*entity.Timesheet(nil),
			setup: func() {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name:         "err upsert multiple timesheet failed row affected different error",
			request:      timesheetEntities,
			expectErr:    fmt.Errorf("err upsert multiple timesheet: %d RowsAffected", 0),
			expectedResp: []*entity.Timesheet(nil),
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
			resp, err := repo.UpsertMultiple(ctx, mockDB.DB, testcase.request.([]*entity.Timesheet))
			assert.Equal(t, testcase.expectErr, err)
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestTimesheetRepo_FindTimesheetByTimesheetIDsAndStatus(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := TimesheetRepoWithSqlMock()
	timesheetIDs := []string{"timesheetID_1", "timesheetID_2"}
	timesheetSubmitEntities := []*entity.Timesheet{
		{
			TimesheetID:     database.Text(timesheetIDs[0]),
			TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String()),
		},
		{
			TimesheetID:     database.Text(timesheetIDs[0]),
			TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String()),
		},
	}
	timesheet := &entity.Timesheet{}
	selectFields, _ := timesheet.FieldMap()
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(selectFields))...)
	t.Run("failed to select invoice records", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		timesheetRecords, err := repo.FindTimesheetByTimesheetIDsAndStatus(ctx, mockDB.DB, []string{timesheetSubmitEntities[0].TimesheetID.String, timesheetSubmitEntities[1].TimesheetID.String}, pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String())
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, timesheetRecords)
		mock.AssertExpectationsForObjects(t, mockDB.DB)

	})
	t.Run("No rows affected", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		timesheetRecords, err := repo.FindTimesheetByTimesheetIDsAndStatus(ctx, mockDB.DB, []string{timesheetSubmitEntities[0].TimesheetID.String, timesheetSubmitEntities[1].TimesheetID.String}, pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String())
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, timesheetRecords)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("success retrieving single timesheet record", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything)

		value := database.GetScanFields(timesheetSubmitEntities[0], selectFields)

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			value,
		})

		timesheetRecords, err := repo.FindTimesheetByTimesheetIDsAndStatus(ctx, mockDB.DB, []string{timesheetSubmitEntities[0].TimesheetID.String}, pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String())
		assert.Nil(t, err)
		assert.Equal(t, []*entity.Timesheet{
			{
				TimesheetID:     database.Text(timesheetIDs[0]),
				TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String()),
			},
		}, timesheetRecords)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
	})

	t.Run("success retrieving multiple timesheet records submit", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything)

		valueOne := database.GetScanFields(timesheetSubmitEntities[0], selectFields)

		value := database.GetScanFields(timesheetSubmitEntities[1], selectFields)

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			valueOne,
			value,
		})

		timesheetRecords, err := repo.FindTimesheetByTimesheetIDsAndStatus(ctx, mockDB.DB, []string{timesheetSubmitEntities[0].TimesheetID.String, timesheetSubmitEntities[1].TimesheetID.String}, pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String())
		assert.Nil(t, err)
		assert.Equal(t, timesheetSubmitEntities, timesheetRecords)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
	})

	t.Run("success retrieving multiple timesheet records approve", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything)

		timesheetApproveEntities := []*entity.Timesheet{
			{
				TimesheetID:     database.Text("test-12"),
				TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String()),
			},
			{
				TimesheetID:     database.Text("test-15"),
				TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String()),
			},
		}

		valueOne := database.GetScanFields(timesheetApproveEntities[0], selectFields)

		value := database.GetScanFields(timesheetApproveEntities[1], selectFields)

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			valueOne,
			value,
		})

		timesheetRecords, err := repo.FindTimesheetByTimesheetIDsAndStatus(ctx, mockDB.DB, []string{timesheetApproveEntities[0].TimesheetID.String, timesheetApproveEntities[1].TimesheetID.String}, pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String())
		assert.Nil(t, err)
		assert.Equal(t, timesheetApproveEntities, timesheetRecords)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
	})
}

func TestTimesheetRepo_UpsertTimesheetStatusMultiple(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := TimesheetRepoWithSqlMock()
	timesheetIDs := []string{"timesheetID_1", "timesheetID_2"}
	staffIDs := []string{"staffID_1", "staffID_2"}
	timesheetEntities := []*entity.Timesheet{
		{
			TimesheetID: database.Text(timesheetIDs[0]),
			StaffID:     database.Text(staffIDs[0]),
		},
		{
			StaffID: database.Text(staffIDs[1]),
		},
	}
	testCases := []struct {
		name            string
		request         interface{}
		expectErr       error
		expectedResp    interface{}
		setup           func()
		timesheetStatus string
	}{
		{
			name:            "err update timesheet status multiple success",
			request:         timesheetEntities,
			timesheetStatus: pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String(),
			expectErr:       nil,
			expectedResp:    timesheetEntities,
			setup: func() {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name:         "err update timesheet status multiple timesheet failed send batch error",
			request:      timesheetEntities,
			expectErr:    fmt.Errorf("results.Exec: %w", puddle.ErrClosedPool),
			expectedResp: []*entity.Timesheet(nil),
			setup: func() {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name:         "err update timesheet status multiple timesheet failed row affected different error",
			request:      timesheetEntities,
			expectErr:    fmt.Errorf("err update timesheet status multiple: %d RowsAffected", 0),
			expectedResp: []*entity.Timesheet(nil),
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
			err := repo.UpdateTimesheetStatusMultiple(ctx, mockDB.DB, testcase.request.([]*entity.Timesheet), testcase.timesheetStatus)
			assert.Equal(t, testcase.expectErr, err)
		})
	}
}

func TestTimesheetRepoImpl_FindTimesheetByTimesheetIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := TimesheetRepoWithSqlMock()
	timesheetIDs := []string{"timesheetID_1", "timesheetID_2"}
	now := time.Now()
	timesheetEntities := []*entity.Timesheet{
		{
			TimesheetID: database.Text(timesheetIDs[0]),
			CreatedAt:   database.Timestamptz(now),
			UpdatedAt:   database.Timestamptz(now),
		},
		{
			TimesheetID: database.Text(timesheetIDs[1]),
			CreatedAt:   database.Timestamptz(now),
			UpdatedAt:   database.Timestamptz(now),
		},
	}
	selectFields := database.GetFieldNames(timesheetEntities[0])

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
			expectedResp: timesheetEntities,
			setup: func() {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything)

				mockDB.MockScanArray(nil, selectFields, [][]interface{}{
					database.GetScanFields(timesheetEntities[0], selectFields),
					database.GetScanFields(timesheetEntities[1], selectFields),
				})
			},
		},
	}
	for _, testcase := range testCases {
		testName := fmt.Sprintf("TestCase: %s", testcase.name)
		t.Run(testName, func(t *testing.T) {
			testcase.setup()
			resp, err := repo.FindTimesheetByTimesheetIDs(ctx, mockDB.DB, testcase.request)
			assert.Equal(t, testcase.expectErr, err)
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestTimesheetRepoImpl_FindTimesheetByLessonIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := TimesheetRepoWithSqlMock()
	timesheetIDs := []string{"timesheetID_1", "timesheetID_2"}
	lessonIDs := []string{"lessonID_1", "lessonID_2"}
	now := time.Now()
	timesheetEntities := []*entity.Timesheet{
		{
			TimesheetID: database.Text(timesheetIDs[0]),
			CreatedAt:   database.Timestamptz(now),
			UpdatedAt:   database.Timestamptz(now),
		},
		{
			TimesheetID: database.Text(timesheetIDs[1]),
			CreatedAt:   database.Timestamptz(now),
			UpdatedAt:   database.Timestamptz(now),
		},
	}
	selectFields := database.GetFieldNames(timesheetEntities[0])

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
			expectedResp: timesheetEntities,
			setup: func() {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
				mockDB.MockScanArray(nil, selectFields, [][]interface{}{
					database.GetScanFields(timesheetEntities[0], selectFields),
					database.GetScanFields(timesheetEntities[1], selectFields),
				})
			},
		},
	}
	for _, testcase := range testCases {
		testName := fmt.Sprintf("TestCase: %s", testcase.name)
		t.Run(testName, func(t *testing.T) {
			testcase.setup()
			resp, err := repo.FindTimesheetByLessonIDs(ctx, mockDB.DB, testcase.request)
			assert.Equal(t, testcase.expectErr, err)
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}
