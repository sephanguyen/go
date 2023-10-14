package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/puddle"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func CourseBookRepoWithSqlMock() (*CourseBookRepo, *testutil.MockDB) {
	r := &CourseBookRepo{}
	return r, testutil.NewMockDB()
}

func TestCourseBookRepo_FindByCourseIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseBookRepoWithSqlMock()
	ids := []string{"id", "id-1"}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &ids)

		books, err := r.FindByCourseIDs(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, books)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &ids)

		e := &entities_bob.CoursesBooks{}
		fields, values := e.FieldMap()
		_ = e.CourseID.Set("id")
		_ = e.BookID.Set(ksuid.New().String())

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		books, err := r.FindByCourseIDs(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.Equal(t, map[string][]string{
			e.CourseID.String: {e.BookID.String},
		}, books)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"course_id":  {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestCourseBookRepo_FindByBookID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseBookRepoWithSqlMock()
	id := "id"
	pgID := database.Text(id)

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &pgID)

		courseIDs, err := r.FindByBookID(ctx, mockDB.DB, id)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, courseIDs)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &pgID)

		e := &entities_bob.CoursesBooks{}
		fields, values := e.FieldMap()
		_ = e.CourseID.Set("id")
		_ = e.BookID.Set(ksuid.New().String())

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		courseIDs, err := r.FindByBookID(ctx, mockDB.DB, id)
		assert.Nil(t, err)
		assert.Equal(t, []string{e.CourseID.String}, courseIDs)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"book_id":    {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}
