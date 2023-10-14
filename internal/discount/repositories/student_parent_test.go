package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func StudentParentWithSqlMock() (*StudentParentRepo, *testutil.MockDB) {
	repo := &StudentParentRepo{}
	return repo, testutil.NewMockDB()
}

func TestStudentParentRepo_GetSiblingIDsByStudentID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := StudentParentWithSqlMock()

	rows := mockDB.Rows

	studentID := "test-student-1"

	t.Run(constant.HappyCase, func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", mock.Anything).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)

		record, err := repo.GetSiblingIDsByStudentID(ctx, mockDB.DB, studentID)
		assert.Nil(t, err)
		assert.NotEmpty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("db.Query returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)

		record, err := repo.GetSiblingIDsByStudentID(ctx, mockDB.DB, studentID)

		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("Row scan returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", mock.Anything).Once().Return(errors.New("test-error"))

		record, err := repo.GetSiblingIDsByStudentID(ctx, mockDB.DB, studentID)

		assert.Equal(t, "row.Scan: test-error", err.Error())
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}
