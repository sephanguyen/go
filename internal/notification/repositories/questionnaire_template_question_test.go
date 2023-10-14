package repositories

import (
	"context"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_BulkForceUpsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}

	testCases := []struct {
		Name        string
		Req         interface{}
		ExpectedErr error
		SetUp       func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Req: []*entities.QuestionnaireTemplateQuestion{
				{
					QuestionnaireTemplateQuestionID: database.Text("1"),
				},
			},
			ExpectedErr: nil,
			SetUp: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}

	repo := &QuestionnaireTemplateQuestionRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.BulkForceUpsert(ctx, db, testCase.Req.([]*entities.QuestionnaireTemplateQuestion))
			if testCase.ExpectedErr == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_SoftDelete(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}

	testCases := []struct {
		Name        string
		Req         []string
		ExpectedErr error
		SetUp       func(ctx context.Context)
	}{
		{
			Name:        "happy case",
			Req:         []string{"1"},
			ExpectedErr: nil,
			SetUp: func(ctx context.Context) {
				db.On("Exec", ctx, mock.AnythingOfType("string"), mock.Anything).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
		},
	}

	repo := &QuestionnaireTemplateQuestionRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.SoftDelete(ctx, db, testCase.Req)
			if testCase.ExpectedErr == nil {
				assert.Nil(t, err)
			}
		})
	}
}
