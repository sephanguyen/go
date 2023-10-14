package repositories

import (
	"context"
	"testing"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ClassMemberRepoWithSqlMock() (*ClassMemberRepo, *testutil.MockDB) {
	r := &ClassMemberRepo{}
	return r, testutil.NewMockDB()
}

func TestClassMemberRepo_Insert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ClassMemberRepoWithSqlMock()

	t.Run("err insert", func(t *testing.T) {
		e := &entities_bob.ClassMember{}
		_, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("no rows affected", func(t *testing.T) {
		e := &entities_bob.ClassMember{}
		_, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.EqualError(t, err, "cannot insert new ClassMember")
	})

	t.Run("success", func(t *testing.T) {
		e := &entities_bob.ClassMember{}
		fields, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})
}

func TestLessonMember_Find(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonMemberRepoWithSqlMock()

	pgStudentID := database.Text("id")

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &pgStudentID)

		result, err := r.Find(ctx, mockDB.DB, pgStudentID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, result)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &pgStudentID)

		e := &entities_bob.LessonMember{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		result, err := r.Find(ctx, mockDB.DB, pgStudentID)
		assert.Nil(t, err)
		assert.Equal(t, []*entities_bob.LessonMember{e}, result)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, "lesson_members", "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"user_id":    {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestClassMemberRepo_FindByUserIDsAndCourseIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ClassMemberRepoWithSqlMock()
	ids := []string{"id", "id-1"}

	pgIDs := database.TextArray([]string{"id", "id-1"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, pgIDs, pgIDs)

		schoolIDs, err := r.FindByUserIDsAndCourseIDs(ctx, mockDB.DB, database.TextArray(ids), database.TextArray(ids))
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, schoolIDs)
	})
}

func TestChapterRepo_FindByClassIDsAndUserIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ClassMemberRepoWithSqlMock()
	ids := []string{"id", "id-1"}

	pgIDs := database.TextArray([]string{"id", "id-1"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, mock.Anything, pgIDs)

		schoolIDs, err := r.FindByClassIDsAndUserIDs(ctx, mockDB.DB, database.TextArray(ids), database.TextArray(ids))
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, schoolIDs)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, pgIDs)

		e := &EnSchoolID{}
		fields, values := e.FieldMap()
		e.SchoolID = 1

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.FindByClassIDsAndUserIDs(ctx, mockDB.DB, database.TextArray(ids), database.TextArray(ids))
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedTable(t, "class_member", "")
	})
}
