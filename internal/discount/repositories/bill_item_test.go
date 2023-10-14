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

func BillItemRepoWithSqlMock() (*BillItemRepo, *testutil.MockDB) {
	billItemRepo := &BillItemRepo{}
	return billItemRepo, testutil.NewMockDB()
}

func TestBillItemRepo_GetLastBillItemOfStudentProduct(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockBillItemRepo, mockDB := BillItemRepoWithSqlMock()

	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryRowArgs(
			t,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		entity := &entities.BillItem{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		billItem, err := mockBillItemRepo.GetLastBillItemOfStudentProduct(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)
		assert.NotNil(t, billItem)
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(
			t,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.BillItem{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		billItem, err := mockBillItemRepo.GetLastBillItemOfStudentProduct(ctx, mockDB.DB, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, billItem)
	})
}
