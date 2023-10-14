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

func AutoCreateFlagActivityLogRepoWithSqlMock() (AutoCreateFlagActivityLogRepoImpl, *testutil.MockDB) {
	mockDB := testutil.NewMockDB()
	repo := AutoCreateFlagActivityLogRepoImpl{}

	return repo, mockDB
}

func TestAutoCreateFlagActivityLog_InsertFlagLog(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := AutoCreateFlagActivityLogRepoWithSqlMock()

	flagLogData := &entity.AutoCreateFlagActivityLog{
		StaffID:    database.Text(idutil.ULIDNow()),
		FlagOn:     database.Bool(true),
		ChangeTime: database.Timestamptz(time.Now()),
	}

	_, flagLogValues := flagLogData.FieldMap()

	argsFlagLog := append(
		[]interface{}{mock.Anything, mock.Anything},
		genSliceMock(len(flagLogValues))...,
	)

	internalErr := errors.New(" internal server error")
	testCases := []struct {
		name         string
		setup        func()
		expectErr    error
		expectedResp *entity.AutoCreateFlagActivityLog
	}{
		{
			name:         "happy case",
			expectErr:    nil,
			expectedResp: flagLogData,
			setup: func() {
				cmtTag := pgconn.CommandTag(`1`)
				mockDB.DB.On("Exec", argsFlagLog...).Once().Return(cmtTag, nil)
			},
		},
		{
			name:         "error case fail to insert auto create log data internal server error",
			expectErr:    fmt.Errorf("err insert auto create log data: %w", internalErr),
			expectedResp: nil,
			setup: func() {
				cmtTag := pgconn.CommandTag(`0`)
				mockDB.DB.On("Exec", argsFlagLog...).Once().Return(cmtTag, internalErr)
			},
		},
		{
			name:         "error case row affected different one",
			expectErr:    fmt.Errorf("err insert auto create log data: %d RowsAffected", 0),
			expectedResp: nil,
			setup: func() {
				cmtTag := pgconn.CommandTag(`0`)
				mockDB.DB.On("Exec", argsFlagLog...).Once().Return(cmtTag, nil)
			},
		},
	}

	for _, testCase := range testCases {
		testName := fmt.Sprintf("TestCase: %s", testCase.name)
		t.Run(testName, func(t *testing.T) {
			testCase.setup()
			resp, err := repo.InsertFlagLog(ctx, mockDB.DB, flagLogData)
			assert.Equal(t, testCase.expectErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}

func TestAutoCreateFlagActivityLog_GetAutoCreateFlagActivityLogByStaffIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := AutoCreateFlagActivityLogRepoWithSqlMock()
	now := time.Now()
	teacherIDs := []string{"teacherID_1", "teacherID_2"}
	flagLogData := []*entity.AutoCreateFlagActivityLog{
		{
			StaffID:    database.Text(idutil.ULIDNow()),
			FlagOn:     database.Bool(true),
			ChangeTime: database.Timestamptz(now.Add(-2)),
		},
	}

	flagLogFields := database.GetFieldNames(flagLogData[0])

	testCases := []struct {
		name         string
		startTime    time.Time
		staffIDs     []string
		expectErr    error
		expectedResp []*entity.AutoCreateFlagActivityLog
		setup        func()
	}{
		{
			name:         "happy case",
			startTime:    now,
			staffIDs:     teacherIDs,
			expectErr:    nil,
			expectedResp: flagLogData,
			setup: func() {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				mockDB.MockScanArray(nil, flagLogFields, [][]interface{}{
					database.GetScanFields(flagLogData[0], flagLogFields),
				})
			},
		},
	}

	for _, testCase := range testCases {
		testName := fmt.Sprintf("TestCase: %s", testCase.name)
		t.Run(testName, func(t *testing.T) {
			testCase.setup()
			resp, err := repo.GetAutoCreateFlagActivityLogByStaffIDs(ctx, mockDB.DB, testCase.startTime, testCase.staffIDs)
			assert.Equal(t, testCase.expectErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}

func TestAutoCreateFlagActivityLog_SoftDeleteFlagLogsWithinTimeRange(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := AutoCreateFlagActivityLogRepoWithSqlMock()
	staffId := "staff_id-0"
	timeNow := time.Now()
	dateNow := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), 0, 0, 0, 0, timeNow.Location())

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &staffId, &dateNow)

	t.Run("happy case", func(t *testing.T) {

		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := repo.SoftDeleteFlagLogsAfterTime(ctx, mockDB.DB, staffId, dateNow)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertUpdatedTable(t, "auto_create_flag_activity_log")
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at")
	})
	t.Run("soft delete timesheet record fail", func(t *testing.T) {

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.SoftDeleteFlagLogsAfterTime(ctx, mockDB.DB, staffId, dateNow)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err delete SoftDeleteFlagLogsAfterTime: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
