package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func StudentsRepoWithSqlMock() (*StudentsRepo, *testutil.MockDB) {
	r := &StudentsRepo{}
	return r, testutil.NewMockDB()
}

func TestStudentsRepo_GetStudentByStudentID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := StudentsRepoWithSqlMock()
	dto := &Student{}
	fields, values := dto.FieldMap()
	studentID := "student-id1"

	t.Run("successful", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), &studentID)
		mockDB.MockRowScanFields(nil, fields, values)

		student, err := repo.GetStudentByStudentID(ctx, mockDB.DB, studentID)
		assert.NoError(t, err)
		assert.NotNil(t, student)
	})

	t.Run("failed with no rows found", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), &studentID)
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)

		student, err := repo.GetStudentByStudentID(ctx, mockDB.DB, studentID)
		assert.True(t, errors.Is(err, domain.ErrStudentNotFound))
		assert.NotNil(t, student)
	})

	t.Run("failed", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), &studentID)
		mockDB.MockRowScanFields(pgx.ErrTxClosed, fields, values)

		student, err := repo.GetStudentByStudentID(ctx, mockDB.DB, studentID)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.NotNil(t, student)
	})
}

func TestStudentsRepo_IsUserIDAStudent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := StudentsRepoWithSqlMock()
	dto := &Student{}
	fields, values := dto.FieldMap()
	studentID := "student-id1"

	t.Run("successful", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), &studentID)
		mockDB.MockRowScanFields(nil, fields, values)

		isStudent, err := repo.IsUserIDAStudent(ctx, mockDB.DB, studentID)
		assert.NoError(t, err)
		assert.True(t, isStudent)
	})

	t.Run("failed with no rows found", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), &studentID)
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)

		isStudent, err := repo.IsUserIDAStudent(ctx, mockDB.DB, studentID)
		assert.Nil(t, err)
		assert.False(t, isStudent)
	})

	t.Run("failed", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), &studentID)
		mockDB.MockRowScanFields(pgx.ErrTxClosed, fields, values)

		isStudent, err := repo.IsUserIDAStudent(ctx, mockDB.DB, studentID)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.False(t, isStudent)
	})
}
