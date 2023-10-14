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

func ProductGroupMappingRepoWithSqlMock() (*ProductGroupMappingRepo, *testutil.MockDB) {
	productGroupMappingRepo := &ProductGroupMappingRepo{}
	return productGroupMappingRepo, testutil.NewMockDB()
}

func TestProductGroupMappingRepo_GetProductGroupMappingByStudentIDAndLocationID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	productGroupMappingRepo, mockDB := ProductGroupMappingRepoWithSqlMock()

	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			nil,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.ProductGroupMapping{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		productGroupMappings, err := productGroupMappingRepo.GetByProductID(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)
		assert.Equal(t, e, productGroupMappings[0])
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			pgx.ErrNoRows,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.ProductGroupMapping{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		productGroupMappings, err := productGroupMappingRepo.GetByProductID(ctx, mockDB.DB, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, productGroupMappings)
	})
}
