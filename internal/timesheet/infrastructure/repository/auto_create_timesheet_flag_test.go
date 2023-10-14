package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func AutoCreateTimesheetFlagRepoWithSqlMock() (AutoCreateFlagRepoImpl, *testutil.MockDB) {
	mockDB := testutil.NewMockDB()
	repo := AutoCreateFlagRepoImpl{}

	return repo, mockDB
}

func TestAutoCreateTimesheetFlag_Retrieve(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := AutoCreateTimesheetFlagRepoWithSqlMock()

	autoCreateFlag := &entity.AutoCreateTimesheetFlag{StaffID: database.Text(idutil.ULIDNow())}

	respValues := [][]byte{}
	respValues = append(respValues, []byte(fmt.Sprintf("%v", autoCreateFlag)))
	testCases := []struct {
		name         string
		setup        func()
		expectErr    error
		expectedResp []*entity.AutoCreateTimesheetFlag
	}{
		{
			name:         "happy case",
			expectErr:    nil,
			expectedResp: []*entity.AutoCreateTimesheetFlag{autoCreateFlag},
			setup: func() {
				mockDB.MockQueryArgs(t, nil, mock.Anything,
					mock.Anything,
					database.TextArray([]string{autoCreateFlag.StaffID.String}),
				)

				fields, values := autoCreateFlag.FieldMap()

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
					database.TextArray([]string{autoCreateFlag.StaffID.String}),
				)

				fields, values := autoCreateFlag.FieldMap()

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
			resp, err := repo.Retrieve(ctx, mockDB.DB, database.TextArray([]string{autoCreateFlag.StaffID.String}))

			assert.Equal(t, testcase.expectErr, err)
			assert.Equal(t, testcase.expectedResp, resp)

		})
	}
}

func TestAutoCreateTimesheetFlag_FindAutoCreatedFlagByStaffID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := AutoCreateTimesheetFlagRepoWithSqlMock()

	autoFlag := &entity.AutoCreateTimesheetFlag{StaffID: database.Text(idutil.ULIDNow())}

	testCases := []struct {
		name         string
		setup        func()
		expectErr    error
		expectedResp interface{}
	}{
		{
			name:         "happy case",
			expectErr:    nil,
			expectedResp: autoFlag,
			setup: func() {
				mockDB.MockQueryArgs(t, nil, mock.Anything,
					mock.Anything,
					database.TextArray([]string{autoFlag.StaffID.String}),
				)

				fields, values := autoFlag.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
		{
			name:         "error query",
			expectErr:    fmt.Errorf("err db.Query: %w", pgx.ErrNoRows),
			expectedResp: (*entity.AutoCreateTimesheetFlag)(nil),
			setup: func() {
				mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything,
					mock.Anything,
					database.TextArray([]string{autoFlag.StaffID.String}),
				)
				fields, values := autoFlag.FieldMap()
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
			autoFlag, err := repo.FindAutoCreatedFlagByStaffID(ctx, mockDB.DB, autoFlag.StaffID)

			assert.Equal(t, testcase.expectErr, err)
			assert.Equal(t, testcase.expectedResp, autoFlag)
		})
	}
}

func TestAutoCreateTimesheetFlag_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := AutoCreateTimesheetFlagRepoWithSqlMock()

	autoCreateFlag := &entity.AutoCreateTimesheetFlag{}
	_, autoCreateFlagValues := autoCreateFlag.FieldMap()
	argsTimesheet := append(
		[]interface{}{mock.Anything, mock.Anything},
		genSliceMock(len(autoCreateFlagValues))...,
	)
	internalErr := errors.New(" internal server error")
	testCases := []struct {
		name         string
		setup        func()
		expectErr    error
		expectedResp *entity.AutoCreateTimesheetFlag
	}{
		{
			name:         "happy case",
			expectErr:    nil,
			expectedResp: autoCreateFlag,
			setup: func() {
				cmtTag := pgconn.CommandTag(`1`)
				mockDB.DB.On("Exec", argsTimesheet...).Return(cmtTag, nil).Once()
			},
		},
		{
			name:      "error case fail to insert timesheet internal server error",
			expectErr: fmt.Errorf("upsert auto create timesheet flag: %w", internalErr),
			setup: func() {
				cmtTag := pgconn.CommandTag(`0`)
				mockDB.DB.On("Exec", argsTimesheet...).Once().Return(cmtTag, internalErr)
			},
		},
		{
			name:      "error case row affected different one",
			expectErr: fmt.Errorf("upsert auto create timesheet flag: %d RowsAffected", 0),
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
			err := repo.Upsert(ctx, mockDB.DB, autoCreateFlag)
			assert.Equal(t, testcase.expectErr, err)
			// assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}
