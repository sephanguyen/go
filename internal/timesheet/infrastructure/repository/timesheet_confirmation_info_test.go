package repository

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TimesheetConfirmationInfoWithSqlMock() (TimesheetConfirmationInfoRepoImpl, *testutil.MockDB) {
	mockDB := testutil.NewMockDB()
	repo := TimesheetConfirmationInfoRepoImpl{}

	return repo, mockDB
}

func TestTimesheetConfirmationInfo_InsertConfirmInfo(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := TimesheetConfirmationInfoWithSqlMock()

	confirmationInfoE := &entity.TimesheetConfirmationInfo{
		ID:         database.Text("confirmation_info_id"),
		LocationID: database.Text("location_id"),
		PeriodID:   database.Text("period_id"),
	}

	_, confirmationInfoFields := confirmationInfoE.FieldMap()
	argsConfirmationInfo := append(
		[]interface{}{mock.Anything, mock.Anything},
		genSliceMock(len(confirmationInfoFields))...,
	)

	testCases := []struct {
		name         string
		setup        func()
		expectErr    error
		expectedResp *timestamp.Timestamp
	}{
		{
			name:      "happy case",
			expectErr: nil,
			setup: func() {
				cmtTag := pgconn.CommandTag(`1`)
				mockDB.DB.On("Exec", argsConfirmationInfo...).Once().Return(cmtTag, nil)
			},
		},
	}

	for _, testcase := range testCases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			testcase.setup()
			_, err := repo.InsertConfirmationInfo(ctx, mockDB.DB, confirmationInfoE)
			assert.Equal(t, testcase.expectErr, err)
		})
	}
}
