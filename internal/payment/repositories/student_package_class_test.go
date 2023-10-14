package repositories

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func StudentPackageClassRepoWithSqlMock() (*StudentPackageClassRepo, *testutil.MockDB) {
	studentPackageClassRepo := &StudentPackageClassRepo{}
	return studentPackageClassRepo, testutil.NewMockDB()
}
func TestStudentPackageClass_Upsert(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := entities.StudentPackageClass{}
	_, fieldMap := mockEntities.FieldMap()
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)
	t.Run(constant.HappyCase, func(t *testing.T) {
		studentProductRepoWithSqlMock, mockDB := StudentPackageClassRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)
		err := studentProductRepoWithSqlMock.Upsert(ctx, mockDB.DB, &entities.StudentPackageClass{})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("Upsert student class fail", func(t *testing.T) {
		studentProductRepoWithSqlMock, mockDB := StudentPackageClassRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := studentProductRepoWithSqlMock.Upsert(ctx, mockDB.DB, &mockEntities)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), pgx.ErrTxClosed.Error()))

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("Upsert student class fail with none record", func(t *testing.T) {
		studentProductRepoWithSqlMock, mockDB := StudentPackageClassRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := studentProductRepoWithSqlMock.Upsert(ctx, mockDB.DB, &mockEntities)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), "error when upsert student package class in payment"))

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentPackageClass_Delete(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := entities.StudentPackageClass{}
	_, fieldMap := mockEntities.FieldMap()
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)
	t.Run(constant.HappyCase, func(t *testing.T) {
		studentProductRepoWithSqlMock, mockDB := StudentPackageClassRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)
		err := studentProductRepoWithSqlMock.Delete(ctx, mockDB.DB, &entities.StudentPackageClass{})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("Delete student class fail", func(t *testing.T) {
		studentProductRepoWithSqlMock, mockDB := StudentPackageClassRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := studentProductRepoWithSqlMock.Delete(ctx, mockDB.DB, &mockEntities)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), pgx.ErrTxClosed.Error()))

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("Delete student class fail with none record", func(t *testing.T) {
		studentProductRepoWithSqlMock, mockDB := StudentPackageClassRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := studentProductRepoWithSqlMock.Delete(ctx, mockDB.DB, &mockEntities)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), "delete student package class have no row affected"))

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentPackageClass_GetByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var studentPackageID string = "1"
	studentPackageRepoWithSqlMock, mockDB := StudentPackageClassRepoWithSqlMock()
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryRowArgs(t,
			mock.Anything,
			mock.Anything,
			studentPackageID,
		)
		entity := &entities.StudentPackageClass{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		period, err := studentPackageRepoWithSqlMock.GetByStudentPackageID(ctx, mockDB.DB, studentPackageID)
		assert.Nil(t, err)
		assert.NotNil(t, period)
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			studentPackageID,
		)
		e := &entities.StudentPackageClass{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrTxClosed, fields, values)
		discount, err := studentPackageRepoWithSqlMock.GetByStudentPackageID(ctx, mockDB.DB, studentPackageID)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.NotNil(t, discount)
	})
}
