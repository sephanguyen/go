package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	"github.com/manabie-com/backend/mock/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TimesheetActionLogWithSqlMock() (TimesheetActionLogRepoImpl, *testutil.MockDB) {
	mockDB := testutil.NewMockDB()
	repo := TimesheetActionLogRepoImpl{}

	return repo, mockDB
}

func TestTimesheetActionLog_CreateTimesheetActionLog(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := TimesheetActionLogWithSqlMock()
	timesheetActionLogE := &entity.TimesheetActionLog{}
	_, fieldMap := timesheetActionLogE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	testCases := []struct {
		name      string
		setup     func()
		expectErr error
	}{
		{
			name: "happy case",
			setup: func() {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("Exec", args...).Return(cmdTag, nil).Once()
			},
			expectErr: nil,
		},
		{
			name: "error case",
			setup: func() {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed).Once()
			},
			expectErr: fmt.Errorf("err insert TimesheetActionLog: %w", pgx.ErrTxClosed),
		},
	}

	for _, tc := range testCases {
		tc.setup()
		err := repo.Create(ctx, mockDB.DB, timesheetActionLogE)
		assert.Equal(t, tc.expectErr, err)
	}
}
