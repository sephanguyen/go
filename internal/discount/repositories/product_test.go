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

func ProductRepoWithSqlMock() (*ProductRepo, *testutil.MockDB) {
	productRepo := &ProductRepo{}
	return productRepo, testutil.NewMockDB()
}
func TestProductRepo_GetByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userRepo, mockDB := ProductRepoWithSqlMock()

	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryRowArgs(
			t,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		entity := &entities.Product{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		product, err := userRepo.GetByID(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)
		assert.NotNil(t, product)
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(
			t,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.Product{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		product, err := userRepo.GetByID(ctx, mockDB.DB, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, product)

	})
}
