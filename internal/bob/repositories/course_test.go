package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func CourseRepoWithSqlMock() (*CourseRepo, *testutil.MockDB) {
	r := &CourseRepo{}
	return r, testutil.NewMockDB()
}

func TestCourseRepo_RetrieveCourses(t *testing.T) {
	// TODO: update course test later
}

func TestCourseRepo_RetrieveByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseRepoWithSqlMock()
	ids := database.TextArray([]string{"id", "id-1"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &ids)

		books, err := r.RetrieveByIDs(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, books)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &ids)

		e := &entities_bob.Course{}
		fields, values := e.FieldMap()
		_ = e.ID.Set("id")

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		books, err := r.RetrieveByIDs(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.Equal(t, entities_bob.Courses{e}, books)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"course_id":  {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestCourseRepo_FindByLessonID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseRepoWithSqlMock()
	lessonID := database.Text("lesson-id")

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &lessonID)

		books, err := r.FindByLessonID(ctx, mockDB.DB, lessonID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, books)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &lessonID)

		e := &entities_bob.Course{}
		fields, values := e.FieldMap()
		_ = e.ID.Set("id")

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		books, err := r.FindByLessonID(ctx, mockDB.DB, lessonID)
		assert.Nil(t, err)
		assert.Equal(t, entities_bob.Courses{e}, books)
	})
}

func TestCourseRepo_FindSchoolIDsOnCourses(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseRepoWithSqlMock()
	ids := []string{"id", "id-1"}

	pgIDs := database.TextArray(ids)

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &pgIDs)

		schoolIDs, err := r.FindSchoolIDsOnCourses(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, schoolIDs)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &pgIDs)

		e := &EnSchoolID{}
		fields, values := e.FieldMap()
		e.SchoolID = 1

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		schoolIDs, err := r.FindSchoolIDsOnCourses(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.Equal(t, []int32{e.SchoolID}, schoolIDs)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, "courses", "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"course_id":  {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestCourseRepo_FindByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseRepoWithSqlMock()

	pgIDs := database.Text("id")

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &pgIDs)

		result, err := r.FindByID(ctx, mockDB.DB, pgIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, result)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &pgIDs)

		e := &entities_bob.Course{}
		fields, values := e.FieldMap()
		e.ID.Set("id")

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		result, err := r.FindByID(ctx, mockDB.DB, pgIDs)
		assert.Nil(t, err)
		assert.Equal(t, e, result)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, "courses", "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"course_id":  {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestCourseRepo_FindByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseRepoWithSqlMock()

	pgIDs := database.TextArray([]string{"id"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &pgIDs)

		result, err := r.FindByIDs(ctx, mockDB.DB, pgIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, result)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &pgIDs)

		e := &entities_bob.Course{}
		fields, values := e.FieldMap()
		e.ID.Set("id")

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		result, err := r.FindByIDs(ctx, mockDB.DB, pgIDs)
		assert.Nil(t, err)
		assert.Equal(t, map[pgtype.Text]*entities_bob.Course{e.ID: e}, result)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, "courses", "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"course_id":  {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestCourseRepo_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseRepoWithSqlMock()

	t.Run("err create", func(t *testing.T) {
		e := &entities_bob.Course{}
		_, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrNotAvailable, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.True(t, errors.Is(err, puddle.ErrNotAvailable))
	})

	t.Run("success", func(t *testing.T) {
		e := &entities_bob.Course{}
		fields, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})
}
