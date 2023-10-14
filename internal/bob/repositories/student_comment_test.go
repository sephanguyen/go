package repositories

import (
	"context"
	"fmt"
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

func studentCommentRepoWithMock() (*StudentCommentRepo, *testutil.MockDB) {
	r := &StudentCommentRepo{}
	return r, testutil.NewMockDB()
}

func Test_DeleteStudentComments(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := studentCommentRepoWithMock()
	ids := database.TextArray([]string{"cmt-1", "cmt-2", "cmt-3"})
	t.Run("err update", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("pgtype.Timestamptz")}, ids)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		err := r.DeleteStudentComments(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("no rows affected", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("pgtype.Timestamptz")}, ids)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		err := r.DeleteStudentComments(ctx, mockDB.DB, ids)
		assert.EqualError(t, err, fmt.Errorf("unexpected RowsAffected value").Error())
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("pgtype.Timestamptz")}, ids)
		mockDB.MockExecArgs(t, pgconn.CommandTag("3"), nil, args...)

		err := r.DeleteStudentComments(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertUpdatedTable(t, "student_comments")
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at", "updated_at")
	})
}
