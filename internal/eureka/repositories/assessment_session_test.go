package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAssessmentSessionRepo_GetAssessmentSessionByAssessmentIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	repo := &AssessmentSessionRepo{}

	args := database.Text("test")

	result := &entities.AssessmentSession{}
	result1 := []*entities.AssessmentSession{
		result,
	}
	fields, values := result.FieldMap()

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			req:          args,
			expectedResp: result1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := repo.GetAssessmentSessionByAssessmentIDs(ctx, mockDB.DB, database.TextArray([]string{"test1", "test2"}))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}
