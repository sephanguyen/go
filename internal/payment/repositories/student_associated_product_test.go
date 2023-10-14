package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func StudentAssociatedProductRepoWithSqlMock() (*StudentAssociatedProductRepo, *testutil.MockDB) {
	r := &StudentAssociatedProductRepo{}
	return r, testutil.NewMockDB()
}

func TestStudentAssociatedProductRepo_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentAssociatedProductRepoWithSqlMock, mockDB := StudentAssociatedProductRepoWithSqlMock()
	db := mockDB.DB
	mockEntity := &entities.StudentAssociatedProduct{}
	_, fieldValues := mockEntity.FieldMap()
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldValues))...)
	testCases := []utils.TestCase{
		{
			Name: "Happy case",
			Req:  entities.StudentAssociatedProduct{},
			Setup: func(ctx context.Context) {
				db.On("Exec", args...).Return(constant.SuccessCommandTag, nil)
			},
		},
		{
			Name:        "Failed case: Error when insert",
			Req:         entities.StudentAssociatedProduct{},
			ExpectedErr: fmt.Errorf("err insert StudentAssociatedProduct: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				db.On("Exec", args...).Return(constant.SuccessCommandTag, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			studentAssociatedProductRepoWithSqlMock, mockDB = StudentAssociatedProductRepoWithSqlMock()
			db = mockDB.DB
			testCase.Setup(ctx)
			req := (testCase.Req).(entities.StudentAssociatedProduct)
			err := studentAssociatedProductRepoWithSqlMock.Create(ctx, db, req)
			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, testCase.ExpectedErr, err)
		})
	}
}

func TestStudentAssociatedProductRepo_GetAssociatedProduct(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentAssociatedProductRepoWithSqlMock, mockDB := StudentAssociatedProductRepoWithSqlMock()

	studentProductID := "studentProductID"
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
			studentProductID,
		)
		e := &entities.StudentProduct{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		studentAssociatedProduct, err := studentAssociatedProductRepoWithSqlMock.GetMapAssociatedProducts(ctx, mockDB.DB, studentProductID)
		assert.Nil(t, err)
		assert.NotNil(t, studentAssociatedProduct)

	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
			studentProductID,
		)
		e := &entities.StudentAssociatedProduct{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(pgx.ErrTxClosed, fields, [][]interface{}{
			values,
		})
		studentAssociatedProduct, err := studentAssociatedProductRepoWithSqlMock.GetMapAssociatedProducts(ctx, mockDB.DB, studentProductID)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Nil(t, studentAssociatedProduct)

	})
	t.Run("err case query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, studentProductID)
		studentAssociatedProduct, err := studentAssociatedProductRepoWithSqlMock.GetMapAssociatedProducts(ctx, mockDB.DB, studentProductID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, studentAssociatedProduct)
	})
}

func TestStudentAssociatedProductRepo_GetAssociatedProductIDsByStudentProductID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentAssociatedProductRepoWithSqlMock, mockDB := StudentAssociatedProductRepoWithSqlMock()

	studentProductID := "studentProductID"
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
			studentProductID, mock.Anything, mock.Anything,
		)
		e := &entities.StudentAssociatedProduct{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		studentAssociatedProductIDs, err := studentAssociatedProductRepoWithSqlMock.GetAssociatedProductIDsByStudentProductID(ctx, mockDB.DB, studentProductID, int64(1), int64(1))
		assert.Nil(t, err)
		assert.NotNil(t, studentAssociatedProductIDs)

	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
			studentProductID, mock.Anything, mock.Anything,
		)
		e := &entities.StudentAssociatedProduct{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(pgx.ErrTxClosed, fields, [][]interface{}{
			values,
		})
		studentAssociatedProductIDs, err := studentAssociatedProductRepoWithSqlMock.GetAssociatedProductIDsByStudentProductID(ctx, mockDB.DB, studentProductID, int64(1), int64(1))
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Nil(t, studentAssociatedProductIDs)

	})
	t.Run("err case query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, studentProductID, mock.Anything, mock.Anything)
		studentAssociatedProductIDs, err := studentAssociatedProductRepoWithSqlMock.GetAssociatedProductIDsByStudentProductID(ctx, mockDB.DB, studentProductID, int64(1), int64(1))
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, studentAssociatedProductIDs)
	})
}
