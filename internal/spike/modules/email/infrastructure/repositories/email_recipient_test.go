package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/model"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEmailRecipientRepo_BulkUpsertEmailRecipients(t *testing.T) {
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
			Req: model.EmailRecipients{
				{
					EmailID:          database.Text("email-id"),
					EmailRecipientID: database.Text("email-recipient-d"),
					RecipientAddress: database.Text("recipient-address"),
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

	repo := &EmailRecipientRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.BulkUpsertEmailRecipients(ctx, db, testCase.Req.(model.EmailRecipients))
			if testCase.ExpectedErr == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_GetEmailRecipientsByEmailID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &EmailRecipientRepo{}

	ent1 := &model.EmailRecipient{}
	ent2 := &model.EmailRecipient{}
	database.AllRandomEntity(ent1)
	database.AllRandomEntity(ent2)
	emailID := "email-id"
	t.Run("success", func(t *testing.T) {
		fields, vals1 := ent1.FieldMap()
		_, vals2 := ent2.FieldMap()
		// scan twice using values from 2 entities
		mockDB.MockScanArray(nil, fields, [][]interface{}{vals1, vals2})

		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, emailID)
		questions, err := r.GetEmailRecipientsByEmailID(ctx, db, emailID)
		assert.Nil(t, err)
		assert.Equal(t, ent1, questions[0])
		assert.Equal(t, ent2, questions[1])
	})
	t.Run("error query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, emailID)
		questions, err := r.GetEmailRecipientsByEmailID(ctx, db, emailID)
		assert.Nil(t, questions)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
	t.Run("error scan", func(t *testing.T) {
		fields, vals := ent1.FieldMap()
		mockDB.MockScanFields(pgx.ErrNoRows, fields, vals)
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, emailID)
		questions, err := r.GetEmailRecipientsByEmailID(ctx, db, emailID)
		assert.Nil(t, questions)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
}
