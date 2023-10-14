package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ProductDiscountRepoWithMock() (*ProductDiscountRepo, *testutil.MockDB, *mock_database.Tx) {
	repo := &ProductDiscountRepo{}
	return repo, testutil.NewMockDB(), &mock_database.Tx{}
}

func TestProductDiscountRepo_GetByProductIDAndDiscountID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	productDiscountRepoWithSqlMock, mockDB, _ := ProductDiscountRepoWithMock()
	var productID string = "1"
	var discountID string = "1"
	t.Run("Success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			productID,
			discountID,
		)
		entities := &entities.ProductDiscount{}
		fields, values := entities.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		productDiscount, err := productDiscountRepoWithSqlMock.GetByProductIDAndDiscountID(ctx, mockDB.DB, productID, discountID)
		assert.Nil(t, err)
		assert.NotNil(t, productDiscount)

	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			productID,
			discountID,
		)
		entities := &entities.ProductDiscount{}
		fields, values := entities.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		productDiscount, err := productDiscountRepoWithSqlMock.GetByProductIDAndDiscountID(ctx, mockDB.DB, productID, discountID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, productDiscount)

	})
}

func TestProductDiscounRepo_Upsert(t *testing.T) {
	t.Parallel()
	db := &mockDb.Ext{}
	ProductDiscountRepo := &ProductDiscountRepo{}
	testCases := []utils.TestCase{
		{
			Name: "happy case",
			Req: []*entities.ProductDiscount{
				{
					ProductID:  pgtype.Text{String: "1", Status: pgtype.Present},
					DiscountID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				batchResults := &mockDb.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(constant.SuccessCommandTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			Name: "happy case: upsert multiple parents",
			Req: []*entities.ProductDiscount{
				{
					ProductID:  pgtype.Text{String: "1", Status: pgtype.Present},
					DiscountID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
				},
				{
					ProductID:  pgtype.Text{String: "1", Status: pgtype.Present},
					DiscountID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				batchResults := &mockDb.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(constant.SuccessCommandTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			Name: "error send batch",
			Req: []*entities.ProductDiscount{
				{
					ProductID:  pgtype.Text{String: "1", Status: pgtype.Present},
					DiscountID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
				},
			},
			ExpectedErr: errors.Wrap(puddle.ErrClosedPool, "batchResults.Exec"),
			Setup: func(ctx context.Context) {
				batchResults := &mockDb.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.Setup(ctx)
		err := ProductDiscountRepo.Upsert(ctx, db, pgtype.Text{String: "1", Status: pgtype.Present}, testCase.Req.([]*entities.ProductDiscount))
		if testCase.ExpectedErr != nil {
			assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.ExpectedErr, err)
		}
	}
}
