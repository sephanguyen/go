package postgres

import (
	"context"
	"errors"
	"testing"

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

func TestConversationMemberRepo_BulkUpsert(t *testing.T) {
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
			Req: []domain.ConversationMember{
				{
					ID:             "id",
					ConversationID: "convo-id",
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

	repo := &ConversationMemberRepo{}
	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.BulkUpsert(ctx, db, testCase.Req.([]domain.ConversationMember))
			if testCase.ExpectedErr == nil {
				assert.Nil(t, err)
			}
		})
	}
}

func TestConversationMemberRepo_GetConversationMembersByUserID(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()

	repo := &ConversationMemberRepo{}
	ctx := context.Background()

	conversationMember1 := &dto.ConversationMemberPgDTO{}
	conversationMember2 := &dto.ConversationMemberPgDTO{}
	database.AllRandomEntity(conversationMember1)
	database.AllRandomEntity(conversationMember2)
	fields1, values1 := conversationMember1.FieldMap()
	_, values2 := conversationMember2.FieldMap()

	conversationIDs := []string{"conversation-id-1", "conversation-id-2"}
	userID := "user-id-1"

	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, ctx, mock.AnythingOfType("string"), database.Text(userID), database.TextArray(conversationIDs))
		mockDB.MockScanArray(nil, fields1, [][]interface{}{values1, values2})
		res, err := repo.GetConversationMembersByUserID(ctx, mockDB.DB, userID, conversationIDs)
		assert.Nil(t, err)
		assert.Equal(t, conversationMember1.ToConversationMemberDomain(), res[0])
		assert.Equal(t, conversationMember2.ToConversationMemberDomain(), res[1])
	})

	t.Run("error select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, ctx, mock.AnythingOfType("string"), database.Text(userID), database.TextArray(conversationIDs))
		_, err := repo.GetConversationMembersByUserID(ctx, mockDB.DB, userID, conversationIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
}

func TestConversationMemberRepo_CheckMembersExistInConversation(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	ctx := context.Background()

	t.Run("should no error", func(t *testing.T) {
		conversationID := "conv-id"
		conversationMemberIDs := []string{"member-1", "member-2"}

		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.Text(conversationID), database.TextArray(conversationMemberIDs))

		id1 := conversationMemberIDs[0]
		id2 := conversationMemberIDs[1]
		mockDB.MockScanArray(nil, []string{"user_id"}, [][]interface{}{
			{&id1},
			{&id2},
		})

		repo := &ConversationMemberRepo{}
		res, err := repo.CheckMembersExistInConversation(ctx, mockDB.DB, conversationID, conversationMemberIDs)
		assert.Equal(t, nil, err)
		assert.Equal(t, conversationMemberIDs, res)
	})
}
