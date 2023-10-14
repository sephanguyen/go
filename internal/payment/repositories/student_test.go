package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func StudentRepoWithSqlMock() (*StudentRepo, *testutil.MockDB) {
	studentRepo := &StudentRepo{}
	return studentRepo, testutil.NewMockDB()
}

func TestStudentRepo_GetByIDForUpdate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentRepoWithSqlMock, mockDB := StudentRepoWithSqlMock()

	userID := "1"
	t.Run("Success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			userID,
		)
		e := &entities.Student{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		student, err := studentRepoWithSqlMock.GetByIDForUpdate(ctx, mockDB.DB, userID)
		assert.Nil(t, err)
		assert.NotNil(t, student)

	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			userID,
		)
		e := &entities.Student{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		user, err := studentRepoWithSqlMock.GetByIDForUpdate(ctx, mockDB.DB, userID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, user)

	})
}

func TestStudentRepo_GetByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	studentRepoWithSqlMock, mockDB := StudentRepoWithSqlMock()
	studentIDs := []string{"student_1", "student_2", "student_3"}
	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, studentIDs)
		e := &entities.Student{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		student, err := studentRepoWithSqlMock.GetByIDs(ctx, mockDB.DB, studentIDs)
		assert.Nil(t, err)
		assert.NotNil(t, student)
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, studentIDs)
		e := &entities.Student{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(pgx.ErrTxClosed, fields, [][]interface{}{
			values,
		})
		student, err := studentRepoWithSqlMock.GetByIDs(ctx, mockDB.DB, studentIDs)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Nil(t, student)
	})
	t.Run("err case query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, studentIDs)
		student, err := studentRepoWithSqlMock.GetByIDs(ctx, mockDB.DB, studentIDs)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, student)
	})

}
