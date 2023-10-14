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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func OtherWorkingHoursRepoWithSqlMock() (OtherWorkingHoursRepoImpl, *testutil.MockDB) {
	mockDB := testutil.NewMockDB()
	repo := OtherWorkingHoursRepoImpl{}

	return repo, mockDB
}

func TestOtherWorkingHoursRepoImpl_UpsertMultiple(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := OtherWorkingHoursRepoWithSqlMock()

	owhs := entity.ListOtherWorkingHours{
		{
			ID: database.Text(idutil.ULIDNow()),
		},
	}

	testCases := []struct {
		name      string
		req       entity.ListOtherWorkingHours
		expectErr error
		setup     func()
	}{
		{
			name:      "happy case",
			req:       owhs,
			expectErr: nil,
			setup: func() {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name:      "error send batch",
			req:       owhs,
			expectErr: puddle.ErrClosedPool,
			setup: func() {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name:      "error row affected different one",
			req:       owhs,
			expectErr: fmt.Errorf("err upsert Other Working Hours: %d RowsAffected", 0),
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

			err := repo.UpsertMultiple(ctx, mockDB.DB, testcase.req)
			assert.Equal(t, testcase.expectErr, err)
		})
	}
}

func TestOtherWorkingHoursRepoImpl_FindListOtherWorkingHoursByTimesheetIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := OtherWorkingHoursRepoWithSqlMock()
	timesheetID := database.Text(idutil.ULIDNow())
	owhsE := &entity.OtherWorkingHours{
		ID:          database.Text(idutil.ULIDNow()),
		TimesheetID: timesheetID,
	}
	owhsE2 := &entity.OtherWorkingHours{
		ID:          database.Text(idutil.ULIDNow()),
		TimesheetID: timesheetID,
	}

	testCases := []struct {
		name         string
		setup        func()
		expectErr    error
		expectedResp []*entity.OtherWorkingHours
	}{
		{
			name:         "happy case",
			expectErr:    nil,
			expectedResp: []*entity.OtherWorkingHours{owhsE, owhsE2},
			setup: func() {
				mockDB.MockQueryArgs(t, nil, mock.Anything,
					mock.Anything,
					database.TextArray([]string{timesheetID.String}),
				)

				fields, values := owhsE.FieldMap()
				_, values2 := owhsE2.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{values, values2})
			},
		},
		{
			name:         "err query",
			expectErr:    fmt.Errorf("err db.Query: %w", pgx.ErrNoRows),
			expectedResp: nil,
			setup: func() {
				mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything,
					mock.Anything,
					database.TextArray([]string{timesheetID.String}),
				)

				fields, values := owhsE.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			testcase.setup()
			resp, err := repo.FindListOtherWorkingHoursByTimesheetIDs(ctx, mockDB.DB, database.TextArray([]string{timesheetID.String}))

			assert.Equal(t, testcase.expectErr, err)
			assert.Equal(t, testcase.expectedResp, resp)

		})
	}
}

func TestOtherWorkingHoursRepoImpl_SoftDeleteByTimesheetID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	timesheetID := database.Text("test-id")

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &timesheetID)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := OtherWorkingHoursRepoWithSqlMock()

		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := repo.SoftDeleteByTimesheetID(ctx, mockDB.DB, timesheetID)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertUpdatedTable(t, "other_working_hours")
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"timesheet_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"deleted_at":   {HasNullTest: true},
		})
	})
	t.Run("soft delete timesheet record fail", func(t *testing.T) {
		repo, mockDB := OtherWorkingHoursRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.SoftDeleteByTimesheetID(ctx, mockDB.DB, timesheetID)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err delete SoftDeleteByTimesheetID: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestOtherWorkingHoursRepoImpl_MapExistingOWHsByTimesheetIds(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := OtherWorkingHoursRepoWithSqlMock()
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
			resp, err := repo.MapExistingOWHsByTimesheetIds(ctx, mockDB.DB, []string{timesheetID})

			assert.Equal(t, testcase.expectErr, err)
			assert.Equal(t, testcase.expectedResp, resp)

		})
	}
}
