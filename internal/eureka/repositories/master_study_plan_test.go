package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestMasterStudyPlanRepo_BulkUpdateTime(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	masterStudyPlanRepo := &MasterStudyPlanRepo{}
	testCases := []TestCase{
		{
			name: "happy Case",
			req: []*entities.MasterStudyPlan{
				&entities.MasterStudyPlan{},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := masterStudyPlanRepo.BulkUpdateTime(ctx, db, testCase.req.([]*entities.MasterStudyPlan))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, nil, err)
		}
	}
}

func TestMasterStudyPlanRepo_BulkUpsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	masterStudyPlanRepo := &MasterStudyPlanRepo{}
	testCases := []TestCase{
		{
			name: "happy Case",
			req: []*entities.MasterStudyPlan{
				&entities.MasterStudyPlan{},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag(`1`)
				db.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
			},
		},
		{
			name: "error exec error",
			req: []*entities.MasterStudyPlan{
				&entities.MasterStudyPlan{},
			},
			expectedErr: fmt.Errorf("database.BulkUpsert error: error exec error"),
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag(`1`)
				db.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, fmt.Errorf("error exec error"))
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := masterStudyPlanRepo.BulkUpsert(ctx, db, testCase.req.([]*entities.MasterStudyPlan))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, nil, err)
		}
	}
}

func TestMasterStudyPlanRepo_FindByIDs(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &MasterStudyPlanRepo{}
	rows := mockDB.Rows

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				// db.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				// rows.On("Scan", mock.Anything, mock.Anything).Once().Return(nil)
				rows.On("Scan", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					arg := args.Get(0).(*pgtype.Text)
					arg.Set("id")
				}).Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
			req: database.Text("id-1"),
			expectedResp: []*entities.MasterStudyPlan{
				&entities.MasterStudyPlan{
					StudyPlanID: database.Text("id"),
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.FindByID(ctx, mockDB.DB, testCase.req.(pgtype.Text))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp.([]*entities.MasterStudyPlan), resp)
			}
		})
	}
}
