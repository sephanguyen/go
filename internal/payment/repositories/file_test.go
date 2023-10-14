package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func FileRepoWithSqlMock() (*FileRepo, *testutil.MockDB) {
	fileRepo := &FileRepo{}
	return fileRepo, testutil.NewMockDB()
}

func TestFileRepoGetByFileName(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	fileRepo, mockDB := FileRepoWithSqlMock()

	const fileID string = "1"
	t.Run("Success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			fileID,
		)
		entity := &entities.File{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		discount, err := fileRepo.GetByFileName(ctx, mockDB.DB, fileID)
		assert.Nil(t, err)
		assert.NotNil(t, discount)

	})
	t.Run("Fail when scan data", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			fileID,
		)
		entity := &entities.File{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(constant.ErrDefault, fields, values)
		_, err := fileRepo.GetByFileName(ctx, mockDB.DB, fileID)
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), constant.ErrDefault.Error())

	})
}

func TestFileRepoCreate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.File{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("insert file succeeds", func(t *testing.T) {
		fileRepo, mockDB := FileRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := fileRepo.Create(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("insert file fails", func(t *testing.T) {
		fileRepo, mockDB := FileRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := fileRepo.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert File: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("no rows affected after insert file", func(t *testing.T) {
		fileRepo, mockDB := FileRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := fileRepo.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert FileRepo: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
