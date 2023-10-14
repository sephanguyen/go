package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/manabie-com/backend/internal/golibs/database"
	lentities "github.com/manabie-com/backend/internal/tom/domain/lesson"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_UpdateLatestStartTime(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB

	r := &ConversationLessonRepo{}
	lessonID := randomText()
	t.Run("err select", func(t *testing.T) {
		now := database.Timestamptz(time.Now())
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, mock.Anything, mock.Anything,
			&lessonID, &now)

		err := r.UpdateLatestStartTime(ctx, db, lessonID, now)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
	t.Run("success updated new session", func(t *testing.T) {
		now := database.Timestamptz(time.Now())
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), nil, mock.Anything, mock.Anything,
			&lessonID, &now)

		err := r.UpdateLatestStartTime(ctx, db, lessonID, now)
		mockDB.RawStmt.AssertUpdatedTable(t, "conversation_lesson")
		mockDB.RawStmt.AssertUpdatedFields(t, "latest_start_time")

		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"lesson_id":  {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"deleted_at": {HasNullTest: true},
		})
		assert.NoError(t, err)
	})
}

func Test_FindByLessonID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &ConversationLessonRepo{}
	scannedlEntities := &lentities.ConversationLesson{
		ConversationID: randomText(),
		LessonID:       randomText(),
	}

	lessonID := database.Text("lesson-id")
	_, scannedlVal := scannedlEntities.FieldMap()

	e := &lentities.ConversationLesson{}
	fields, _ := e.FieldMap()

	t.Run("err query", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &lessonID)
		mockDB.MockRowScanFields(puddle.ErrClosedPool, fields, scannedlVal)
		conversationLesson, err := r.FindByLessonID(ctx, db, lessonID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, conversationLesson)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &lessonID)
		mockDB.MockRowScanFields(nil, fields, scannedlVal)

		conversationLesson, err := r.FindByLessonID(ctx, db, lessonID)

		assert.Equal(t, err, nil)
		assert.Equal(t, scannedlEntities, conversationLesson)
		mockDB.RawStmt.AssertSelectedTable(t, "conversation_lesson", "")
	})
}

func Test_FindByLessonIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &ConversationLessonRepo{}
	scannedlEntities := []*lentities.ConversationLesson{
		{
			ConversationID: randomText(),
			LessonID:       randomText(),
		},
		{
			ConversationID: randomText(),
			LessonID:       randomText(),
		},
	}

	lessonIDs := database.TextArray([]string{"lesson-1", "lesson-2"})
	scannedVal := [][]interface{}{}
	for _, e := range scannedlEntities {
		_, vals := e.FieldMap()
		scannedVal = append(scannedVal, vals)
	}

	e := &lentities.ConversationLesson{}
	fields, _ := e.FieldMap()

	t.Run("err query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &lessonIDs)
		conversationLessons, err := r.FindByLessonIDs(ctx, db, lessonIDs, false)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, conversationLessons)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &lessonIDs)
		mockDB.MockScanArray(nil, fields, scannedVal)

		conversationLessons, err := r.FindByLessonIDs(ctx, db, lessonIDs, true)

		assert.Equal(t, err, nil)
		assert.Equal(t, scannedlEntities, conversationLessons)
		mockDB.RawStmt.AssertSelectedTable(t, "conversation_lesson", "")
	})
}

func TestConversationLessonRepo_BulkUpdateResourcePath(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &ConversationLessonRepo{}

	offsetID := pgtype.Text{}
	offsetID.Set(nil)
	lessonIDs := []string{"lesson-1", "lesson-2"}
	resourcePath := "manabie"

	t.Run("success", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag{}, nil, mock.Anything, mock.MatchedBy(func(execString string) bool {
			stmt := testutil.ParseSQL(t, execString)
			return stmt.MustGetUpdatedTable() == "conversation_lesson" && cmp.Equal(stmt.MustGetUpdatedFields(), []string{"resource_path"})
		}), database.Text(resourcePath), database.TextArray(lessonIDs))
		err := r.BulkUpdateResourcePath(ctx, db, lessonIDs, resourcePath)
		assert.NoError(t, err)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag{}, pgx.ErrTxClosed, mock.Anything, mock.MatchedBy(func(execString string) bool {
			stmt := testutil.ParseSQL(t, execString)
			return stmt.MustGetUpdatedTable() == "conversation_lesson" && cmp.Equal(stmt.MustGetUpdatedFields(), []string{"resource_path"})
		}), database.Text(resourcePath), database.TextArray(lessonIDs))
		err := r.BulkUpdateResourcePath(ctx, db, lessonIDs, resourcePath)
		assert.ErrorIs(t, err, pgx.ErrTxClosed)
	})
}
