package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func BookRepoWithSqlMock() (*BookRepo, *testutil.MockDB) {
	r := &BookRepo{}
	return r, testutil.NewMockDB()
}

func TestBookRepo_FindByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := BookRepoWithSqlMock()
	ids := []string{"id", "id-1"}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &ids)

		books, err := r.FindByIDs(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, books)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &ids)

		e := &entities.Book{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(ksuid.New().String())
		_ = e.Name.Set(ksuid.New().String())

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		books, err := r.FindByIDs(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.Equal(t, map[string]*entities.Book{
			e.ID.String: e,
		}, books)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"book_id":    {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestBookRepo_FindByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := BookRepoWithSqlMock()
	bookID := database.Text("mock-book-id")
	e := &entities.Book{}
	_, values := e.FieldMap()
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &bookID)
		mockDB.DB.On("QueryRow").Once().Return(mockDB.Row, nil)
		mockDB.Row.On("Scan", values...).Once().Return(pgx.ErrNoRows)
		book, err := r.FindByID(ctx, mockDB.DB, bookID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, book)
	})

	t.Run("success query", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &bookID)
		mockDB.DB.On("QueryRow").Once().Return(mockDB.Row, nil)
		mockDB.Row.On("Scan", values...).Once().Return(nil)
		book, err := r.FindByID(ctx, mockDB.DB, bookID)
		assert.True(t, errors.Is(err, nil))
		assert.NotNil(t, book)
		mockDB.RawStmt.AssertSelectedFields(t, "book_id", "name", "country", "subject", "grade", "school_id", "updated_at", "created_at", "deleted_at", "copied_from", "current_chapter_display_order")
		mockDB.RawStmt.AssertSelectedTable(t, "books", "")
	})
}

func TestBookRepo_UpdateCurrentChapterDisplayOrder(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := BookRepoWithSqlMock()

	bookID := database.Text("mock-book-id")
	totalGeneratedChapterDisplayOrder := database.Int4(4)
	t.Run("err update", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, totalGeneratedChapterDisplayOrder, bookID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := r.UpdateCurrentChapterDisplayOrder(ctx, mockDB.DB, totalGeneratedChapterDisplayOrder, bookID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, totalGeneratedChapterDisplayOrder, bookID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
		err := r.UpdateCurrentChapterDisplayOrder(ctx, mockDB.DB, totalGeneratedChapterDisplayOrder, bookID)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertUpdatedTable(t, "books")
		mockDB.RawStmt.AssertUpdatedFields(t, "current_chapter_display_order")
	})
}
