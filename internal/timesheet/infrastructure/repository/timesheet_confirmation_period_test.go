package repository

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TimesheetConfirmationPeriodWithSqlMock() (TimesheetConfirmationPeriodRepoImpl, *testutil.MockDB) {
	mockDB := testutil.NewMockDB()
	repo := TimesheetConfirmationPeriodRepoImpl{}

	return repo, mockDB
}

func TestTimesheetConfirmationPeriod_GetPeriodByDate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := TimesheetConfirmationPeriodWithSqlMock()

	var startDateExpect *timestamp.Timestamp = timestamppb.Now()
	periodE := &entity.TimesheetConfirmationPeriod{}

	testCases := []struct {
		name         string
		setup        func()
		expectErr    error
		expectedResp *timestamp.Timestamp
	}{
		{
			name:         "happy case",
			expectErr:    nil,
			expectedResp: startDateExpect,
			setup: func() {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				fields, values := periodE.FieldMap()
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
			_, err := repo.GetPeriodByDate(ctx, mockDB.DB, time.Now())
			assert.Equal(t, testcase.expectErr, err)
		})
	}
}
