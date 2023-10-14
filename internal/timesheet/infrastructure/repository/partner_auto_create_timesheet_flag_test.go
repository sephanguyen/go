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
)

func PartnerAutoCreateTimesheetFlagWithSqlMock() (PartnerAutoCreateTimesheetFlagRepoImpl, *testutil.MockDB) {
	mockDB := testutil.NewMockDB()
	repo := PartnerAutoCreateTimesheetFlagRepoImpl{}

	return repo, mockDB
}

func TestPartnerAutoCreateTimesheetFlag_GetPartnerAutoCreateDefaultValue(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := PartnerAutoCreateTimesheetFlagWithSqlMock()

	partnerAutoCreateFlagE := &entity.PartnerAutoCreateTimesheetFlag{}

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
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				fields, values := partnerAutoCreateFlagE.FieldMap()
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
			_, err := repo.GetPartnerAutoCreateDefaultValue(ctx, mockDB.DB)
			assert.Equal(t, testcase.expectErr, err)
		})
	}
}
