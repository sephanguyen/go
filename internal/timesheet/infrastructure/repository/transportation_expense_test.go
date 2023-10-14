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

func TransportationExpenseRepoWithSqlMock() (TransportationExpenseRepoImpl, *testutil.MockDB) {
	mockDB := testutil.NewMockDB()
	repo := TransportationExpenseRepoImpl{}

	return repo, mockDB
}

func TestTransportationExpenseRepoImpl_UpsertMultiple(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := TransportationExpenseRepoWithSqlMock()

	listTransportExpeneses := entity.ListTransportationExpenses{
		{
			TransportationExpenseID: database.Text(idutil.ULIDNow()),
		},
	}

	testCases := []struct {
		name      string
		req       entity.ListTransportationExpenses
		expectErr error
		setup     func()
	}{
		{
			name:      "happy case",
			req:       listTransportExpeneses,
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
			req:       listTransportExpeneses,
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
			name:      "error row upsert affected different one",
			req:       listTransportExpeneses,
			expectErr: fmt.Errorf("err upsert transportation expense: %d RowsAffected", 0),
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

func TestTransportationExpenseRepoImpl_FindListTransportationExpenseByTimesheetIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := TransportationExpenseRepoWithSqlMock()
	timesheetID := database.Text(idutil.ULIDNow())
	transportExpenseE := &entity.TransportationExpense{
		TransportationExpenseID: database.Text(idutil.ULIDNow()),
		TimesheetID:             timesheetID,
	}
	transportExpenseE2 := &entity.TransportationExpense{
		TransportationExpenseID: database.Text(idutil.ULIDNow()),
		TimesheetID:             timesheetID,
	}

	testCases := []struct {
		name         string
		setup        func()
		expectErr    error
		expectedResp []*entity.TransportationExpense
	}{
		{
			name:         "happy case",
			expectErr:    nil,
			expectedResp: []*entity.TransportationExpense{transportExpenseE, transportExpenseE2},
			setup: func() {
				mockDB.MockQueryArgs(t, nil, mock.Anything,
					mock.Anything,
					database.TextArray([]string{timesheetID.String}),
				)

				fields, values := transportExpenseE.FieldMap()
				_, values2 := transportExpenseE2.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{values, values2})
			},
		},
		{
			name:         "err exec query",
			expectErr:    fmt.Errorf("err db.Query: %w", pgx.ErrNoRows),
			expectedResp: nil,
			setup: func() {
				mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything,
					mock.Anything,
					database.TextArray([]string{timesheetID.String}),
				)

				fields, values := transportExpenseE.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			testcase.setup()
			resp, err := repo.FindListTransportExpensesByTimesheetIDs(ctx, mockDB.DB, database.TextArray([]string{timesheetID.String}))

			assert.Equal(t, testcase.expectErr, err)
			assert.Equal(t, testcase.expectedResp, resp)

		})
	}
}

func TestTransportationExpenseRepoImpl_SoftDeleteByTimesheetID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	timesheetID := database.Text("test-timesheet-id")

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &timesheetID)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := TransportationExpenseRepoWithSqlMock()

		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := repo.SoftDeleteByTimesheetID(ctx, mockDB.DB, timesheetID)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertUpdatedTable(t, "transportation_expense")
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"timesheet_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"deleted_at":   {HasNullTest: true},
		})
	})
	t.Run("soft delete by timesheet record fail", func(t *testing.T) {
		repo, mockDB := TransportationExpenseRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.SoftDeleteByTimesheetID(ctx, mockDB.DB, timesheetID)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err delete SoftDeleteByTimesheetID: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestTransportationExpenseRepoImpl_SoftDeleteMultipleByTimesheetIDs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	timesheetIDs := []string{"test-timesheet-id"}

	mockE := &entity.TransportationExpense{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := TransportationExpenseRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.SoftDeleteMultipleByTimesheetIDs(ctx, mockDB.DB, timesheetIDs)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("soft delete by timesheet ids record fail", func(t *testing.T) {
		repo, mockDB := TransportationExpenseRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.SoftDeleteMultipleByTimesheetIDs(ctx, mockDB.DB, timesheetIDs)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err delete SoftDeleteMultipleByTimesheetIDs: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
