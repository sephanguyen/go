package repositories

import (
	"context"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestImportStudyPlan_Insert(t *testing.T) {
	t.Parallel()
	importStudyPlanRepo := &ImportStudyPlanTaskRepo{}
	mockDB := testutil.NewMockDB()
	task := &entities.ImportStudyPlanTask{
		TaskID: database.Text("task-id"),
	}
	testCases := []InsertTestCase{
		{
			name:        "happy case",
			req:         task,
			expectedErr: nil,
			setup: func(ctx context.Context, db *mock_database.QueryExecer, row *mock_database.Row) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmdTag, nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx, (*mock_database.QueryExecer)(mockDB.DB), mockDB.Row)
		err := importStudyPlanRepo.Insert(ctx, mockDB.DB, testCase.req.(*entities.ImportStudyPlanTask))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}

	return
}

func TestImportStudyPlan_Update(t *testing.T) {
	t.Parallel()
	importStudyPlanRepo := &ImportStudyPlanTaskRepo{}
	mockDB := testutil.NewMockDB()
	task := &entities.ImportStudyPlanTask{
		TaskID: database.Text("task-id"),
	}
	testCases := []InsertTestCase{
		{
			name:        "happy case",
			req:         task,
			expectedErr: nil,
			setup: func(ctx context.Context, db *mock_database.QueryExecer, row *mock_database.Row) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmdTag, nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx, (*mock_database.QueryExecer)(mockDB.DB), mockDB.Row)
		err := importStudyPlanRepo.Update(ctx, mockDB.DB, testCase.req.(*entities.ImportStudyPlanTask))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}

	return
}
