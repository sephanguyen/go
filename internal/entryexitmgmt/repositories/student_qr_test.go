package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func StudentQRRepoWithSqlMock() (*StudentQRRepo, *testutil.MockDB) {
	repo := &StudentQRRepo{}
	return repo, testutil.NewMockDB()
}

func TestStudentQRRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.StudentQR{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := StudentQRRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.Create(ctx, mockDB.DB, mockE)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("insert student_qr failed", func(t *testing.T) {
		repo, mockDB := StudentQRRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert StudentQR: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("No rows affected after student_qr insert", func(t *testing.T) {
		repo, mockDB := StudentQRRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert StudentQR: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentQRRepo_Upsert(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.StudentQR{}
	fields := []string{"student_id", "qr_url", "version", "created_at", "updated_at"}

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fields))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := StudentQRRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.Upsert(ctx, mockDB.DB, mockE)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("upsert student_qr failed", func(t *testing.T) {
		repo, mockDB := StudentQRRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.Upsert(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err upsert StudentQR: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("No rows affected after student_qr insert", func(t *testing.T) {
		repo, mockDB := StudentQRRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.Upsert(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err upsert StudentQR: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentQRRepo_FindByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := StudentQRRepoWithSqlMock()
	mockE := &entities.StudentQR{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("select student_qr failed", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		latestRecord, err := repo.FindByID(ctx, mockDB.DB, string(mock.AnythingOfType("string")))
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("err FindByID StudentQR: err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, latestRecord)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		latestRecord, err := repo.FindByID(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)
		assert.Equal(t, mockE, latestRecord)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("No rows after student_qr select", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		latestRecord, err := repo.FindByID(ctx, mockDB.DB, string(mock.AnythingOfType("string")))

		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("err FindByID StudentQR: err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, latestRecord)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestStudentQRRepo_DeleteByStudentID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.StudentQR{}
	mockE.StudentID.Set(uuid.NewString())
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := StudentQRRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.DeleteByStudentID(ctx, mockDB.DB, mockE.StudentID.String)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("delete student_qr failed", func(t *testing.T) {
		repo, mockDB := StudentQRRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.DeleteByStudentID(ctx, mockDB.DB, mockE.StudentID.String)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err DeleteByStudentID StudentQR: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func genSliceMock(n int) []interface{} {
	result := []interface{}{}
	for i := 0; i < n; i++ {
		result = append(result, mock.Anything)
	}
	return result
}
