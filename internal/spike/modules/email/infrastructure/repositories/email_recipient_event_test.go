package repositories

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/model"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEmailRecipientEventRepo_BulkUpsertEmailRecipients(t *testing.T) {
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
			Req: model.EmailRecipientEvents{
				{
					EmailRecipientEventID: database.Text("email-recipient-event-id"),
					EmailRecipientID:      database.Text("email-recipient-id"),
					Type:                  database.Text("type"),
					Event:                 database.Text("event"),
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

	repo := &EmailRecipientEventRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.BulkInsertEmailRecipientEventRepo(ctx, db, testCase.Req.(model.EmailRecipientEvents))
			if testCase.ExpectedErr == nil {
				assert.Nil(t, err)
			}
		})
	}
}
