package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBillItemAccountCategoryRepo_MultiCreate(t *testing.T) {
	t.Parallel()
	db := &mockDb.Ext{}
	billItemAccountCategoryRepo := &BillItemAccountCategoryRepo{}
	testCases := []utils.TestCase{
		{
			Name: "happy case",
			Req: []*entities.BillItemAccountCategory{
				{
					BillItemSequenceNumber: pgtype.Int4{Int: 1, Status: pgtype.Present},
					AccountCategoryID:      pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
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
			Req: []*entities.BillItemAccountCategory{
				{
					BillItemSequenceNumber: pgtype.Int4{Int: 1, Status: pgtype.Present},
					AccountCategoryID:      pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
				},
				{
					BillItemSequenceNumber: pgtype.Int4{Int: 1, Status: pgtype.Present},
					AccountCategoryID:      pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
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
			Req: []*entities.BillItemAccountCategory{
				{
					BillItemSequenceNumber: pgtype.Int4{Int: 1, Status: pgtype.Present},
					AccountCategoryID:      pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
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
		{
			Name: "error : order item course not inserted",
			Req: []*entities.BillItemAccountCategory{
				{
					BillItemSequenceNumber: pgtype.Int4{Int: 1, Status: pgtype.Present},
					AccountCategoryID:      pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
				},
			},
			ExpectedErr: fmt.Errorf("bill item account categories not inserted"),
			Setup: func(ctx context.Context) {
				batchResults := &mockDb.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(constant.FailCommandTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.Setup(ctx)
		err := billItemAccountCategoryRepo.CreateMultiple(ctx, db, testCase.Req.([]*entities.BillItemAccountCategory))
		if testCase.ExpectedErr != nil {
			assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.ExpectedErr, err)
		}
	}
}
