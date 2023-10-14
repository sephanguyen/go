package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func LessonMemberRepoWithSqlMock() (*LessonMemberRepo, *testutil.MockDB) {
	r := &LessonMemberRepo{}
	return r, testutil.NewMockDB()
}

func TestLessonMemberRepo_SoftDelete(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonMemberRepoWithSqlMock()

	studentID := database.Text("studentID")
	lessonIDs := database.TextArray([]string{"lesson-1", "lesson-2"})

	t.Run("err update", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &studentID, &lessonIDs)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := r.SoftDelete(ctx, mockDB.DB, studentID, lessonIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &studentID, &lessonIDs)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.SoftDelete(ctx, mockDB.DB, studentID, lessonIDs)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertUpdatedTable(t, "lesson_members")
		// move primaryField to the last
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"user_id":    {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"lesson_id":  {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 2}},
			"deleted_at": {HasNullTest: true},
		})
	})
}
