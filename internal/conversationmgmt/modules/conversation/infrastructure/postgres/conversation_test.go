package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/infrastructure/postgres/dto"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConversationRepo_UpsertConversation(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	convoID := "convo-id-1"
	testCases := []struct {
		Name  string
		Ent   *domain.Conversation
		Err   error
		SetUp func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Ent:  &domain.Conversation{ID: convoID},
			SetUp: func(ctx context.Context) {
				e := &dto.ConversationPgDTO{}
				fields := database.GetFieldNames(e)
				values := database.GetScanFields(e, fields)
				mockValues := make([]interface{}, 0, len(values)+2)
				mockValues = append(mockValues, mock.Anything)
				mockValues = append(mockValues, mock.AnythingOfType("string"))
				for range values {
					mockValues = append(mockValues, mock.Anything)
				}
				db.On("Exec", mockValues...).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
		},
	}

	repo := &ConversationRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.UpsertConversation(ctx, db, testCase.Ent)
			if testCase.Err == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func TestConversationRepo_UpdateLatestMessage(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	convoID := "convo-id-1"
	testCases := []struct {
		Name  string
		Ent   *domain.Message
		Err   error
		SetUp func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Ent:  &domain.Message{ConversationID: convoID},
			SetUp: func(ctx context.Context) {
				e := &dto.ConversationPgDTO{}
				fields := database.GetFieldNames(e)
				values := database.GetScanFields(e, fields)
				mockValues := make([]interface{}, 0, len(values)+2)
				mockValues = append(mockValues, mock.Anything)
				mockValues = append(mockValues, mock.AnythingOfType("string"))
				for range values {
					mockValues = append(mockValues, mock.Anything)
				}
				db.On("Exec", mockValues...).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
		},
	}

	repo := &ConversationRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.UpdateLatestMessage(ctx, db, testCase.Ent)
			if testCase.Err == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func TestConversationRepo_FindByIDsAndUserID(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()

	repo := &ConversationRepo{}
	ctx := context.Background()

	conversation1 := &dto.ConversationPgDTO{}
	conversation2 := &dto.ConversationPgDTO{}
	database.AllRandomEntity(conversation1)
	database.AllRandomEntity(conversation2)
	fields1, values1 := conversation1.FieldMap()
	_, values2 := conversation2.FieldMap()

	conversationIDs := []string{"conversation-id-1", "conversation-id-2"}
	userID := "user-id-1"

	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, ctx, mock.AnythingOfType("string"), database.Text(userID), database.TextArray(conversationIDs))
		mockDB.MockScanArray(nil, fields1, [][]interface{}{values1, values2})
		res, err := repo.FindByIDsAndUserID(ctx, mockDB.DB, userID, conversationIDs)
		assert.Nil(t, err)
		expectedConvo1, _ := conversation1.ToConversationDomain()
		expectedConvo2, _ := conversation2.ToConversationDomain()
		assert.Equal(t, expectedConvo1, res[0])
		assert.Equal(t, expectedConvo2, res[1])
	})

	t.Run("error select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, ctx, mock.AnythingOfType("string"), database.Text(userID), database.TextArray(conversationIDs))
		_, err := repo.FindByIDsAndUserID(ctx, mockDB.DB, userID, conversationIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
}

func TestConversationRepo_FindByIds(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	ctx := context.Background()

	t.Run("should no error", func(t *testing.T) {
		convIDs := []string{"conv-1", "conv-2"}
		repo := &ConversationRepo{}
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.TextArray(convIDs))
		emptyTime := time.Time{}
		e1 := &dto.ConversationPgDTO{
			ID:                    database.Text("conv-1"),
			Name:                  database.Text("name-1"),
			LatestMessage:         database.JSONB(nil),
			LatestMessageSentTime: database.Timestamptz(emptyTime),
			CreatedAt:             database.Timestamptz(emptyTime),
			UpdatedAt:             database.Timestamptz(emptyTime),
			DeletedAt:             database.Timestamptz(emptyTime),
			OptionalConfig:        database.JSONB(nil),
		}
		fields, values := e1.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		domainConversations := []*domain.Conversation{
			{
				ID:                    "conv-1",
				Name:                  "name-1",
				LatestMessage:         nil,
				LatestMessageSentTime: &emptyTime,
				CreatedAt:             emptyTime,
				UpdatedAt:             emptyTime,
				OptionalConfig:        nil,
			},
		}
		res, err := repo.FindByIDs(ctx, mockDB.DB, convIDs)
		assert.Equal(t, nil, err)
		assert.Equal(t, domainConversations, res)
	})
}

func TestConversationRepo_UpdateConversationInfo(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	repo := &ConversationRepo{}

	ctx := context.Background()

	conversation := &domain.Conversation{
		ID:             "conversation-id-1",
		Name:           "conversation-name",
		OptionalConfig: []byte("optional-config"),
	}

	t.Run("happy case", func(t *testing.T) {
		db.On("Exec", ctx, mock.AnythingOfType("string"), database.Text(conversation.Name), database.JSONB(conversation.OptionalConfig), database.Text(conversation.ID)).
			Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
		err := repo.UpdateConversationInfo(ctx, db, conversation)
		assert.Nil(t, err)
	})
}
