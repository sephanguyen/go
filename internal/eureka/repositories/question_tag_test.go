package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestQuestionTagRepo_BulkUpsert(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &QuestionTagRepo{}
	validQuestionTag1 := &entities.QuestionTag{
		QuestionTagID:     database.Text("question-tag-id-1"),
		Name:              database.Text("question-tag-name-1"),
		QuestionTagTypeID: database.Text("question-tag-type-id-1"),
	}
	validQuestionTag2 := &entities.QuestionTag{
		QuestionTagID:     database.Text("question-tag-id-2"),
		Name:              database.Text("question-tag-name-2"),
		QuestionTagTypeID: database.Text("question-tag-type-id-2"),
	}
	validQuestionTags := []*entities.QuestionTag{
		validQuestionTag1, validQuestionTag2,
	}
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				_, fields1 := validQuestionTag1.FieldMap()
				_, fields2 := validQuestionTag2.FieldMap()
				var fields = make([]interface{}, 0)
				fields = append(fields, fields1...)
				fields = append(fields, fields2...)
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
			req: validQuestionTags,
		},
		{
			name: "error execute query",
			setup: func(ctx context.Context) {
				_, fields1 := validQuestionTag1.FieldMap()
				_, fields2 := validQuestionTag2.FieldMap()
				var fields = make([]interface{}, 0)
				fields = append(fields, fields1...)
				fields = append(fields, fields2...)
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), fmt.Errorf("error execute query"), args...)
			},
			req:         validQuestionTags,
			expectedErr: fmt.Errorf("QuestionTagRepo.BulkUpsert error: error execute query"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.BulkUpsert(ctx, mockDB.DB, testCase.req.([]*entities.QuestionTag))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}
