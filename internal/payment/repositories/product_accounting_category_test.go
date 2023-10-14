package repositories

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProductAccountingCategoryRepo_Upsert(t *testing.T) {
	t.Parallel()
	db := &mockDb.Ext{}
	productAccountingCategoryRepo := &ProductAccountingCategoryRepo{}
	testCases := []utils.TestCase{
		{
			Name: "happy case",
			Req: []*entities.ProductAccountingCategory{
				{
					ProductID:            pgtype.Text{String: "1", Status: pgtype.Present},
					AccountingCategoryID: pgtype.Text{String: "1", Status: pgtype.Present},
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
			Req: []*entities.ProductAccountingCategory{
				{
					ProductID:            pgtype.Text{String: "1", Status: pgtype.Present},
					AccountingCategoryID: pgtype.Text{String: "1", Status: pgtype.Present},
				},
				{
					ProductID:            pgtype.Text{String: "2", Status: pgtype.Present},
					AccountingCategoryID: pgtype.Text{String: "2", Status: pgtype.Present},
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
			Req: []*entities.ProductAccountingCategory{
				{
					ProductID:            pgtype.Text{String: "1", Status: pgtype.Present},
					AccountingCategoryID: pgtype.Text{String: "1", Status: pgtype.Present},
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
		err := productAccountingCategoryRepo.Upsert(ctx, db, pgtype.Text{String: "1", Status: pgtype.Present}, testCase.Req.([]*entities.ProductAccountingCategory))
		if testCase.ExpectedErr != nil {
			assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.ExpectedErr, err)
		}
	}
}
