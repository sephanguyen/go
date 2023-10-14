package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestQuestionTagTypeRepo_BulkUpsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	questionTagTypeRepo := &QuestionTagTypeRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.QuestionTagType{
				{
					QuestionTagTypeID: database.Text("ID-1"),
					Name:              database.Text("name-1"),
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockFields := []interface{}{mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything}
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mockFields...)
				db.On("Exec", args...).Once().Return(cmdTag, nil)
			},
		},
		{
			name: "error exec error",
			req: []*entities.QuestionTagType{
				{
					QuestionTagTypeID: database.Text("ID-2"),
					Name:              database.Text("name-2"),
				},
			},
			expectedErr: fmt.Errorf("QuestionTagTypeRepo.BulkUpsert error: error exec error"),
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockFields := []interface{}{mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything}
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mockFields...)
				db.On("Exec", args...).Once().Return(cmdTag, fmt.Errorf("error exec error"))
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := questionTagTypeRepo.BulkUpsert(ctx, db, testCase.req.([]*entities.QuestionTagType))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, nil, err)
		}
	}
}
