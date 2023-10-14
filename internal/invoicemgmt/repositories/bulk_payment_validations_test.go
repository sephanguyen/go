package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func BulkPaymentValidationsWithSqlMock() (*BulkPaymentValidationsRepo, *testutil.MockDB) {
	repo := &BulkPaymentValidationsRepo{}
	return repo, testutil.NewMockDB()
}

func TestBulkPaymentValidationsRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.BulkPaymentValidations{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := BulkPaymentValidationsWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		_, err := repo.Create(ctx, mockDB.DB, mockE)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("insert failed", func(t *testing.T) {
		repo, mockDB := BulkPaymentValidationsWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		_, err := repo.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert BulkPaymentValidationsRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("No rows affected after inserted", func(t *testing.T) {
		repo, mockDB := BulkPaymentValidationsWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		_, err := repo.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert BulkPaymentValidationsRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestBulkPaymentValidationsRepo_FindByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := BulkPaymentValidationsWithSqlMock()
	mockE := &entities.BulkPaymentValidations{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		e, err := repo.FindByID(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)
		assert.Equal(t, mockE, e)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("negative test - tx closed", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		e, err := repo.FindByID(ctx, mockDB.DB, "")
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("err FindByID BulkPaymentValidations: err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, e)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("negative test - no rows", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		e, err := repo.FindByID(ctx, mockDB.DB, "")
		assert.True(t, errors.Is(err, pgx.ErrNoRows))

		assert.Equal(t, fmt.Errorf("err FindByID BulkPaymentValidations: err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, e)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestBulkPaymentValidationsRepo_UpdateWithFields(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := BulkPaymentValidationsWithSqlMock()
	mockE := &entities.BulkPaymentValidations{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		cmdTag := pgconn.CommandTag([]byte(`1`))

		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		err := repo.UpdateWithFields(ctx, mockDB.DB, mockE, []string{})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("update failed", func(t *testing.T) {
		repo, mockDB := BulkPaymentValidationsWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.UpdateWithFields(ctx, mockDB.DB, mockE, []string{})

		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err updateWithFields BulkPaymentValidationsRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("No rows affected after update", func(t *testing.T) {
		repo, mockDB := BulkPaymentValidationsWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.UpdateWithFields(ctx, mockDB.DB, mockE, []string{})
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err updateWithFields BulkPaymentValidationsRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
