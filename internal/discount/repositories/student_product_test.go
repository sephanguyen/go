package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func StudentProductRepoWithSqlMock() (*StudentProductRepo, *testutil.MockDB) {
	studentProductRepo := &StudentProductRepo{}
	return studentProductRepo, testutil.NewMockDB()
}

func TestStudentProductRepo_GetActiveStudentProductsByStudentIDAndLocationID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentProductRepo, mockDB := StudentProductRepoWithSqlMock()

	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			nil,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.StudentProduct{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		studentProducts, err := studentProductRepo.GetActiveStudentProductsByStudentIDAndLocationID(ctx, mockDB.DB, mock.Anything, mock.Anything)
		assert.Nil(t, err)
		assert.Equal(t, e, studentProducts[0])
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			pgx.ErrNoRows,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.StudentProduct{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		studentProducts, err := studentProductRepo.GetActiveStudentProductsByStudentIDAndLocationID(ctx, mockDB.DB, mock.Anything, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, studentProducts)
	})
}

func TestStudentProductRepo_GetByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentProductRepo, mockDB := StudentProductRepoWithSqlMock()

	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryRowArgs(
			t,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		entity := &entities.StudentProduct{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		studentProduct, err := studentProductRepo.GetByID(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)
		assert.NotNil(t, studentProduct)
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(
			t,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.StudentProduct{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		studentProduct, err := studentProductRepo.GetByID(ctx, mockDB.DB, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, studentProduct)

	})
}

func TestStudentProductRepo_GetByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentProductRepo, mockDB := StudentProductRepoWithSqlMock()

	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			nil,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := entities.StudentProduct{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		studentProducts, err := studentProductRepo.GetByIDs(ctx, mockDB.DB, []string{mock.Anything})
		assert.Nil(t, err)
		assert.Equal(t, e, studentProducts[0])
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			pgx.ErrNoRows,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := entities.StudentProduct{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		studentProducts, err := studentProductRepo.GetByIDs(ctx, mockDB.DB, []string{mock.Anything})
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, studentProducts)
	})
}
