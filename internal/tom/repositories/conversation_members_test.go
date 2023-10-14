package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConversationMemberRepo_FindByConversationIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	rows := mockDB.Rows
	r := &ConversationMemberRepo{}
	e := &entities.ConversationMembers{}

	conversationIDs := database.TextArray([]string{"conversation-1", "conversation-2", "conversation-3"})

	t.Run("err query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &conversationIDs)
		conversationMember, err := r.FindByConversationIDs(ctx, db, conversationIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, conversationMember)
	})
	t.Run("success", func(t *testing.T) {
		expectedMap := make(map[pgtype.Text][]*entities.ConversationMembers)
		expectedMap[pgtype.Text{}] = []*entities.ConversationMembers{{}}
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &conversationIDs)

		_, values := e.FieldMap()
		mockDB.DB.On("Query").Once().Return(rows, nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", values...).Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)
		rows.On("Close").Once().Return(nil)

		conversation, err := r.FindByConversationIDs(ctx, db, conversationIDs)

		assert.Equal(t, err, nil)
		assert.Equal(t, expectedMap, conversation)
		assert.Equal(t, len(expectedMap), len(conversation), "the length of map conversation members not equal")
		mockDB.RawStmt.AssertSelectedFields(t, "conversation_statuses_id", "user_id", "conversation_id", "role", "status", "seen_at", "last_notify_at", "created_at", "updated_at")

		mockDB.RawStmt.AssertSelectedTable(t, "conversation_members", "")
	})
}

func TestConversationMemberRepo_Find(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	row := mockDB.Row

	r := &ConversationMemberRepo{}
	e := &entities.ConversationMembers{}
	_, values := e.FieldMap()

	conversationID := database.Text("conversation-1")
	role := database.Text(entities.ConversationRoleTeacher)
	status := database.Text(entities.ConversationStatusActive)

	t.Run("err query", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &conversationID, &role, &status)
		db.On("QueryRow").Once().Return(row, nil)
		row.On("Scan", values...).Once().Return(puddle.ErrClosedPool)
		conversation, err := r.Find(ctx, db, conversationID, role, status)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, conversation)
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &conversationID, &role, &status)
		db.On("QueryRow").Once().Return(row, nil)
		row.On("Scan", values...).Once().Return(nil)
		conversation, err := r.Find(ctx, db, conversationID, role, status)
		assert.True(t, errors.Is(err, nil))
		assert.NotNil(t, conversation)
		mockDB.RawStmt.AssertSelectedFields(t, "conversation_statuses_id", "user_id", "conversation_id", "role", "status", "seen_at", "last_notify_at", "created_at", "updated_at")
		mockDB.RawStmt.AssertSelectedTable(t, "conversation_members", "")
	})
}

func TestConversationMemberRepo_FindByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	row := mockDB.Row

	r := &ConversationMemberRepo{}
	e := &entities.ConversationMembers{}
	_, values := e.FieldMap()

	conversationID := database.Text("conversation-1")

	t.Run("err query", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &conversationID)
		db.On("QueryRow").Once().Return(row, nil)
		row.On("Scan", values...).Once().Return(puddle.ErrClosedPool)
		conversation, err := r.FindByID(ctx, db, conversationID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, conversation)
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &conversationID)
		db.On("QueryRow").Once().Return(row, nil)
		row.On("Scan", values...).Once().Return(nil)
		conversation, err := r.FindByID(ctx, db, conversationID)
		assert.True(t, errors.Is(err, nil))
		assert.NotNil(t, conversation)
		mockDB.RawStmt.AssertSelectedFields(t, "conversation_statuses_id", "user_id", "conversation_id", "role", "status", "seen_at", "last_notify_at", "created_at", "updated_at")
		mockDB.RawStmt.AssertSelectedTable(t, "conversation_members", "")
	})
}

func TestConversationMemberRepo_FindByCIDsAndUserID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	rows := mockDB.Rows

	r := &ConversationMemberRepo{}

	conversationIDs := database.TextArray([]string{"conversation-1", "conversation-2"})
	userID := database.Text("user-id-1")

	t.Run("err query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &userID, &conversationIDs)
		conversation, err := r.FindByCIDsAndUserID(ctx, db, conversationIDs, userID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Len(t, conversation, 0)
	})

	cmember1 := entities.ConversationMembers{
		ID:             database.Text(idutil.ULIDNow()),
		UserID:         userID,
		ConversationID: database.Text("conversation-1"),
	}
	cmember2 := entities.ConversationMembers{
		ID:             database.Text(idutil.ULIDNow()),
		UserID:         userID,
		ConversationID: database.Text("conversation-2"),
	}

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &userID, &conversationIDs)
		db.On("Query").Once().Return(rows, nil)
		fields, val1 := cmember1.FieldMap()
		_, val2 := cmember2.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			val1,
			val2,
		})
		conversationMembers, err := r.FindByCIDsAndUserID(ctx, db, conversationIDs, userID)
		assert.True(t, errors.Is(err, nil))
		assert.NotNil(t, conversationMembers)
		assert.Equal(t, conversationMembers, []*entities.ConversationMembers{
			&cmember1,
			&cmember2,
		})
		mockDB.RawStmt.AssertSelectedFields(t, "conversation_statuses_id", "user_id", "conversation_id", "role", "status", "seen_at", "last_notify_at", "created_at", "updated_at")
		mockDB.RawStmt.AssertSelectedTable(t, "conversation_members", "")
	})
}

func TestConversationMemberRepo_FindByCIDAndUserID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	row := mockDB.Row

	r := &ConversationMemberRepo{}
	e := &entities.ConversationMembers{}
	_, values := e.FieldMap()

	conversationID := database.Text("conversation-1")
	userID := database.Text("user-id-1")

	t.Run("err query", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &userID, &conversationID)
		db.On("QueryRow").Once().Return(row, nil)
		row.On("Scan", values...).Once().Return(puddle.ErrClosedPool)
		conversation, err := r.FindByCIDAndUserID(ctx, db, conversationID, userID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, conversation)
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &userID, &conversationID)
		db.On("QueryRow").Once().Return(row, nil)
		row.On("Scan", values...).Once().Return(nil)
		conversation, err := r.FindByCIDAndUserID(ctx, db, conversationID, userID)
		assert.True(t, errors.Is(err, nil))
		assert.NotNil(t, conversation)
		mockDB.RawStmt.AssertSelectedFields(t, "conversation_statuses_id", "user_id", "conversation_id", "role", "status", "seen_at", "last_notify_at", "created_at", "updated_at")
		mockDB.RawStmt.AssertSelectedTable(t, "conversation_members", "")
	})
}

func TestConversationMemberRepo_FindUnseenSince(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	rows := mockDB.Rows

	r := &ConversationMemberRepo{}
	e := &entities.ConversationMembers{}
	_, values := e.FieldMap()

	since := database.Timestamptz(time.Now())

	t.Run("err query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &since)
		db.On("Query").Once().Return(rows, nil)
		rows.On("Scan", values...).Once().Return(nil)
		conversation, err := r.FindUnseenSince(ctx, db, since)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, conversation)
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &since)
		_, values := e.FieldMap()
		mockDB.DB.On("Query").Once().Return(rows, nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", values...).Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)
		rows.On("Close").Once().Return(nil)

		conversation, err := r.FindUnseenSince(ctx, db, since)

		assert.True(t, errors.Is(err, nil))
		assert.NotNil(t, conversation)
		mockDB.RawStmt.AssertSelectedFields(t, "conversation_statuses_id", "user_id", "conversation_id", "role", "status", "seen_at", "last_notify_at", "created_at", "updated_at")
		mockDB.RawStmt.AssertSelectedTable(t, "conversation_members", "")
	})
}

func TestConversationMemberRepo_SetStatusByConversationAndUserIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB

	r := &ConversationMemberRepo{}
	// e := &entities.ConversationMembers{}
	// _, values := e.FieldMap()

	conversationIDs := database.TextArray([]string{"conversation-1"})
	userIDs := database.TextArray([]string{"user-id-1"})
	status := database.Text(entities.ConversationStatusInActive)

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &conversationIDs, &userIDs, &status)

	t.Run("err query", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrNotAvailable, args...)
		err := r.SetStatusByConversationAndUserIDs(ctx, db, conversationIDs, userIDs, status)
		assert.True(t, errors.Is(err, puddle.ErrNotAvailable))
	})

	t.Run("err no rows", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		err := r.SetStatusByConversationAndUserIDs(ctx, db, conversationIDs, userIDs, status)
		assert.EqualError(t, err, "can not update conversation_members")
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.SetStatusByConversationAndUserIDs(ctx, db, conversationIDs, userIDs, status)
		assert.True(t, errors.Is(err, nil))
	})
}
