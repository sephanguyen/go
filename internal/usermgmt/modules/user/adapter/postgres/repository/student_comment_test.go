package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func studentCommentRepoWithMock() (*StudentCommentRepo, *testutil.MockDB) {
	r := &StudentCommentRepo{}
	return r, testutil.NewMockDB()
}

func TestStudentCommentRepo_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entity.StudentComment{}
	_, fieldMap := mockE.FieldMap()

	r, mockDB := studentCommentRepoWithMock()
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.Upsert(ctx, mockDB.DB, mockE)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("err: upsert failed", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), pgx.ErrTxClosed, args...)

		err := r.Upsert(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("%w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})

	t.Run("err: no row affected", func(t *testing.T) {
		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.MockExecArgs(t, cmdTag, nil, args...)

		err := r.Upsert(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, errors.New("cannot insert new student_comments").Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentCommentRepo_DeleteStudentComments(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := studentCommentRepoWithMock()
	ids := []string{"comment-1", "comment-2"}

	t.Run("happy case", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("pgtype.Timestamptz")}, database.TextArray(ids))
		mockDB.MockExecArgs(t, pgconn.CommandTag("2"), nil, args...)

		err := r.DeleteStudentComments(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertUpdatedTable(t, "student_comments")
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at", "updated_at")
	})

	t.Run("err update", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("pgtype.Timestamptz")}, database.TextArray(ids))
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		err := r.DeleteStudentComments(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("no rows affected", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("pgtype.Timestamptz")}, database.TextArray(ids))
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		err := r.DeleteStudentComments(ctx, mockDB.DB, ids)
		assert.EqualError(t, err, fmt.Errorf("unexpected RowsAffected value").Error())
	})
}

func TestStudentCommentRepo_RetrieveByStudentID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	commentIds := pgtype.TextArray{}
	_ = commentIds.Set([]string{uuid.NewString()})

	r, mockDB := studentCommentRepoWithMock()
	studentID := database.Text("id")

	_, studentCommentValues := (&entity.StudentComment{}).FieldMap()
	argsStudentComments := append([]interface{}{}, genSliceMock(len(studentCommentValues))...)

	t.Run("happy case retrieve comment by student id success", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &studentID).Once().Return(mockDB.Rows, nil)
		for range commentIds.Elements {
			mockDB.Rows.On("Next").Once().Return(true)
			mockDB.Rows.On("Scan", argsStudentComments...).Once().Return(nil)
		}
		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		comments, err := r.RetrieveByStudentID(ctx, mockDB.DB, studentID)
		assert.Nil(t, err)
		assert.NotNil(t, comments)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("db Query return error", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &studentID).Once().Return(mockDB.Rows, pgx.ErrTxClosed)

		comments, err := r.RetrieveByStudentID(ctx, mockDB.DB, studentID)
		assert.NotNil(t, t, err)
		assert.Nil(t, comments)
		assert.EqualError(t, err, pgx.ErrTxClosed.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("rows Scan returns error", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &studentID).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsStudentComments...).Once().Return(pgx.ErrTxClosed)
		mockDB.Rows.On("Close").Once().Return(nil)

		comments, err := r.RetrieveByStudentID(ctx, mockDB.DB, studentID)
		assert.NotNil(t, t, err)
		assert.Nil(t, comments)
		assert.EqualError(t, err, pgx.ErrTxClosed.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("rows.Error return error", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &studentID).Once().Return(mockDB.Rows, nil)
		for i := 0; i < len(commentIds.Elements); i++ {
			mockDB.Rows.On("Next").Once().Return(true)
			mockDB.Rows.On("Scan", argsStudentComments...).Once().Return(nil)
		}
		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(errors.New("mock-error"))
		mockDB.Rows.On("Close").Once().Return(nil)

		comments, err := r.RetrieveByStudentID(ctx, mockDB.DB, studentID)
		assert.NotNil(t, t, err)
		assert.Nil(t, comments)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)

	})
}
