package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	sentities "github.com/manabie-com/backend/internal/tom/domain/support"
	"github.com/manabie-com/backend/mock/testutil"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConversationStudentRepo_FindSearchIndexTime(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &ConversationStudentRepo{}

	convIDs := database.TextArray([]string{"conv-1", "conv-2"})
	f := &sentities.ConversationStudent{}
	fields, _ := f.FieldMap()
	convID1 := database.Text("conv-1")
	convID2 := database.Text("conv-2")
	checkTime := database.Timestamptz(time.Now())
	t.Run("lack of result", func(t *testing.T) {
		mockDB.MockScanFields(nil, fields, []interface{}{
			&convID1, &checkTime,
		})

		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, convIDs)

		_, err := r.FindSearchIndexTime(ctx, db, convIDs)

		assert.Equal(t, err, fmt.Errorf("want %d items after searching for search index time, has %d", 2, 1))
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			{
				&convID1, &checkTime,
			},
			{
				&convID2, &checkTime,
			},
		})
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, convIDs)

		idtimeMap, err := r.FindSearchIndexTime(ctx, db, convIDs)

		assert.NoError(t, err)
		assert.Equal(t, idtimeMap["conv-1"].Time, checkTime.Time)
		assert.Equal(t, idtimeMap["conv-2"].Time, checkTime.Time)
	})
}

func TestConversationStudentRepo_UpdateSearchIndexTime(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &ConversationStudentRepo{}

	convIDs := database.TextArray([]string{"conv-1", "conv-2"})
	updateTime := database.Timestamptz(time.Now())

	t.Run("err select", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag{}, nil, mock.Anything, mock.Anything, updateTime, convIDs)

		err := r.UpdateSearchIndexTime(ctx, db, convIDs, updateTime)
		assert.NoError(t, err)
		mockDB.RawStmt.AssertUpdatedFields(t, "search_index_time")
		mockDB.RawStmt.AssertUpdatedTable(t, "conversation_students")
	})
}

func TestConversationStudentRepo_FindByConversationIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &ConversationStudentRepo{}

	ids := []string{"id", "id-1"}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, database.TextArray(ids))

		conversationStudentsMap, err := r.FindByConversationIDs(ctx, db, database.TextArray(ids))
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, conversationStudentsMap)
	})

	t.Run("success with conversation students map", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.TextArray(ids))

		// var conversationID string
		expected := &sentities.ConversationStudent{}
		expected.ID.Set("id-1")
		expected.ConversationID.Set("conversation_id-1")
		fields := database.GetFieldNames(expected)

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			database.GetScanFields(expected, fields),
		})
		e := &sentities.ConversationStudent{}

		conversationStudentsMap, err := r.FindByConversationIDs(ctx, db, database.TextArray(ids))
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, database.GetFieldNames(e)...)

		mockDB.RawStmt.AssertSelectedTable(t, "conversation_students", "")

		convStudent, ok := conversationStudentsMap[expected.ConversationID]
		assert.True(t, ok)
		assert.Equal(t, expected, convStudent)
	})
}

func TestConversationStudentRepo_FindSchoolIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &ConversationStudentRepo{}

	ids := []string{"id", "id-1"}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, database.TextArray(ids), database.Text(tpb.ConversationType_CONVERSATION_STUDENT.String()))

		schoolIDs, err := r.FindByStudentIDs(ctx, db, database.TextArray(ids), database.Text(tpb.ConversationType_CONVERSATION_STUDENT.String()))
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, schoolIDs)
	})

	t.Run("success with conversation id", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.TextArray(ids), database.Text(tpb.ConversationType_CONVERSATION_STUDENT.String()))

		var conversationID string
		mockDB.MockScanArray(nil, []string{"conversation-id"}, [][]interface{}{
			{&conversationID},
		})

		_, err := r.FindByStudentIDs(ctx, db, database.TextArray(ids), database.Text(tpb.ConversationType_CONVERSATION_STUDENT.String()))
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, "conversation_id")

		mockDB.RawStmt.AssertSelectedTable(t, "conversation_students", "")
	})
}
