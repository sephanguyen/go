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

func StaffTransportationExpenseRepoWithSqlMock() (StaffTransportationExpenseRepoImpl, *testutil.MockDB) {
	mockDB := testutil.NewMockDB()
	repo := StaffTransportationExpenseRepoImpl{}

	return repo, mockDB
}

func TestStaffTransportationExpenseRepoImpl_UpsertMultiple(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := StaffTransportationExpenseRepoWithSqlMock()

	listTransportExpeneses := entity.ListStaffTransportationExpense{
		{
			ID: database.Text(idutil.ULIDNow()),
		},
	}

	testCases := []struct {
		name      string
		req       entity.ListStaffTransportationExpense
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
			expectErr: fmt.Errorf("err upsert staff transportation expense: %d RowsAffected", 0),
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

func TestStaffTransportationExpenseRepoImpl_FindListTransportExpensesByStaffIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := StaffTransportationExpenseRepoWithSqlMock()
	staffID := idutil.ULIDNow()
	locationID := idutil.ULIDNow()
	staffTE := entity.NewStaffTransportationExpense()

	selectFields := database.GetFieldNames(staffTE)

	staffTE.ID.Set(idutil.ULIDNow())
	staffTE.StaffID.Set(staffID)
	staffTE.LocationID.Set(locationID)

	testCases := []struct {
		name         string
		setup        func()
		expectErr    error
		expectedResp []*entity.StaffTransportationExpense
	}{
		{
			name:         "happy case",
			expectErr:    nil,
			expectedResp: []*entity.StaffTransportationExpense{staffTE},
			setup: func() {
				mockDB.MockQueryArgs(t, nil,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)

				mockDB.MockScanArray(nil, selectFields, [][]interface{}{
					database.GetScanFields(staffTE, selectFields),
				})
			},
		},
		{
			name:         "err query",
			expectErr:    fmt.Errorf("rows.Scan: %w", pgx.ErrNoRows),
			expectedResp: nil,
			setup: func() {
				mockDB.MockQueryArgs(t, nil,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)

				mockDB.MockScanArray(pgx.ErrNoRows, selectFields, [][]interface{}{
					database.GetScanFields(staffTE, selectFields),
				})
			},
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			testcase.setup()
			resp, err := repo.FindListTransportExpensesByStaffIDs(ctx, mockDB.DB, database.TextArray([]string{staffID}))

			assert.Equal(t, testcase.expectErr, err)
			assert.Equal(t, testcase.expectedResp, resp)

		})
	}
}

func TestStaffTransportationExpenseRepoImpl_FindListTransportExpensesByStaffIDsAndLocation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := StaffTransportationExpenseRepoWithSqlMock()
	staffID := idutil.ULIDNow()
	locationID := idutil.ULIDNow()
	staffTE := entity.NewStaffTransportationExpense()

	selectFields := database.GetFieldNames(staffTE)

	staffTE.ID.Set(idutil.ULIDNow())
	staffTE.StaffID.Set(staffID)
	staffTE.LocationID.Set(locationID)

	testCases := []struct {
		name         string
		setup        func()
		expectErr    error
		expectedResp map[string][]entity.StaffTransportationExpense
	}{
		{
			name:         "happy case",
			expectErr:    nil,
			expectedResp: map[string][]entity.StaffTransportationExpense{staffID: {*staffTE}},
			setup: func() {
				mockDB.MockQueryArgs(t, nil,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)

				mockDB.MockScanArray(nil, selectFields, [][]interface{}{
					database.GetScanFields(staffTE, selectFields),
				})
			},
		},
		{
			name:         "err query",
			expectErr:    fmt.Errorf("rows.Scan: %w", pgx.ErrNoRows),
			expectedResp: nil,
			setup: func() {
				mockDB.MockQueryArgs(t, nil,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)

				mockDB.MockScanArray(pgx.ErrNoRows, selectFields, [][]interface{}{
					database.GetScanFields(staffTE, selectFields),
				})
			},
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			testcase.setup()
			resp, err := repo.FindListTransportExpensesByStaffIDsAndLocation(ctx, mockDB.DB, []string{staffID}, locationID)

			assert.Equal(t, testcase.expectErr, err)
			assert.Equal(t, testcase.expectedResp, resp)

		})
	}
}
