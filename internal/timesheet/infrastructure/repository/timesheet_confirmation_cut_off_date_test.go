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

func TimesheetConfirmationCutOffDateWithSqlMock() (TimesheetConfirmationCutOffDateRepoImpl, *testutil.MockDB) {
	mockDB := testutil.NewMockDB()
	repo := TimesheetConfirmationCutOffDateRepoImpl{}

	return repo, mockDB
}

func TestTimesheetConfirmationCutOffDate_GetCutOffDateByDate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := TimesheetConfirmationCutOffDateWithSqlMock()

	var startDateExpect *timestamp.Timestamp = timestamppb.Now()
	cutOffDateE := &entity.TimesheetConfirmationCutOffDate{}

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
				fields, values := cutOffDateE.FieldMap()
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
			_, err := repo.GetCutOffDateByDate(ctx, mockDB.DB, time.Now())
			assert.Equal(t, testcase.expectErr, err)
		})
	}
}
