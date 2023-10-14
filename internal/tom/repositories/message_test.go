package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	domain "github.com/manabie-com/backend/internal/tom/domain/core"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_FindLessonMessages(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &MessageRepo{}
	conversationID := database.Text(idutil.ULIDNow())
	t.Run("err select", func(t *testing.T) {
		args := domain.FindMessagesArgs{}

		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything,
			&conversationID, &args.EndAt, &args.Limit)

		conversation, err := r.FindLessonMessages(ctx, db, conversationID, &args)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, conversation)
	})
	t.Run("success without system msg", func(t *testing.T) {
		args := domain.FindMessagesArgs{}
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
			&conversationID, &args.EndAt, &args.Limit)

		fields := database.GetFieldNames(&entities.Message{})
		scannedValues := [][]interface{}{}
		expecteds := []*entities.Message{}
		for i := 0; i < 10; i++ {
			expecteds = append(expecteds, &entities.Message{
				ID:             randomText(),
				ConversationID: randomText(),
				UserID:         randomText(),
				Message:        randomText(),
				UrlMedia:       randomText(),
				Type:           randomText(),
				DeletedAt:      randomTime(),
				UpdatedAt:      randomTime(),
				CreatedAt:      randomTime(),
				TargetUser:     randomText(),
			})
		}
		for idx := range expecteds {
			val := expecteds[idx]
			scannedValues = append(scannedValues, database.GetScanFields(val, fields))
		}
		mockDB.MockScanArray(nil, fields, scannedValues)

		msges, err := r.FindLessonMessages(ctx, db, conversationID, &args)
		assert.NoError(t, err)
		assert.Equal(t, expecteds, msges)
	})
	t.Run("success include system msg", func(t *testing.T) {
		args := domain.FindMessagesArgs{
			IncludeSystemMsg: true,
		}
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
			&conversationID, &args.EndAt, &args.IncludeMessageTypes, &args.ExcludeMessagesTypes, &args.Limit)

		fields := database.GetFieldNames(&entities.Message{})
		scannedValues := [][]interface{}{}
		expecteds := []*entities.Message{}
		for i := 0; i < 10; i++ {
			expecteds = append(expecteds, &entities.Message{
				ID:             randomText(),
				ConversationID: randomText(),
				UserID:         randomText(),
				Message:        randomText(),
				UrlMedia:       randomText(),
				Type:           randomText(),
				DeletedAt:      randomTime(),
				UpdatedAt:      randomTime(),
				CreatedAt:      randomTime(),
				TargetUser:     randomText(),
			})
		}
		for idx := range expecteds {
			val := expecteds[idx]
			scannedValues = append(scannedValues, database.GetScanFields(val, fields))
		}
		mockDB.MockScanArray(nil, fields, scannedValues)

		msges, err := r.FindLessonMessages(ctx, db, conversationID, &args)
		assert.NoError(t, err)
		assert.Equal(t, expecteds, msges)
	})
}

func Test_MessageRepo_BulkUpsertMessage(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	repo := &MessageRepo{}
	batchResults := &mock_database.BatchResults{}
	messages := []*domain.Message{
		{},
	}
	cmdTag := pgconn.CommandTag([]byte(`1`))
	t.Run("error closing batch", func(t *testing.T) {
		db.On("SendBatch", mock.Anything, mock.MatchedBy(func(b *pgx.Batch) bool {
			return b.Len() == len(messages)
		})).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, pgx.ErrNoRows)
		batchResults.On("Close").Once().Return(nil)
		err := repo.BulkUpsert(context.Background(), db, messages)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
	t.Run("success", func(t *testing.T) {
		db.On("SendBatch", mock.Anything, mock.MatchedBy(func(b *pgx.Batch) bool {
			return b.Len() == len(messages)
		})).Once().Return(batchResults)
		batchResults.On("Exec").Times(len(messages)).Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)
		err := repo.BulkUpsert(context.Background(), db, messages)
		assert.NoError(t, err)
	})
}
