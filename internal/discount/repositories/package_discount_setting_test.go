package repositories

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/discount/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPackageDiscountSettingRepo_Upsert(t *testing.T) {
	t.Parallel()
	db := &mockDb.Ext{}
	packageDiscountSettingRepo := &PackageDiscountSettingRepo{}
	testCases := []utils.TestCase{
		{
			Name: "happy case",
			Req: []*entities.PackageDiscountSetting{
				{
					PackageID:      pgtype.Text{String: "1", Status: pgtype.Present},
					DiscountTagID:  pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					MinSlotTrigger: pgtype.Int4{Int: 2, Status: pgtype.Present},
					MaxSlotTrigger: pgtype.Int4{Int: 3, Status: pgtype.Present},
					IsArchived:     pgtype.Bool{Bool: false, Status: pgtype.Present},
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
			Name: "happy case: upsert multiple package discount setting",
			Req: []*entities.PackageDiscountSetting{
				{
					PackageID:      pgtype.Text{String: "1", Status: pgtype.Present},
					DiscountTagID:  pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					MinSlotTrigger: pgtype.Int4{Int: 2, Status: pgtype.Present},
					MaxSlotTrigger: pgtype.Int4{Int: 3, Status: pgtype.Present},
					IsArchived:     pgtype.Bool{Bool: false, Status: pgtype.Present},
				},
				{
					PackageID:      pgtype.Text{String: "1", Status: pgtype.Present},
					DiscountTagID:  pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					MinSlotTrigger: pgtype.Int4{Int: 1, Status: pgtype.Present},
					MaxSlotTrigger: pgtype.Int4{Int: 5, Status: pgtype.Present},
					IsArchived:     pgtype.Bool{Bool: false, Status: pgtype.Present},
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
			Req: []*entities.PackageDiscountSetting{
				{
					PackageID:      pgtype.Text{String: "1", Status: pgtype.Present},
					DiscountTagID:  pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					MinSlotTrigger: pgtype.Int4{Int: 1, Status: pgtype.Present},
					MaxSlotTrigger: pgtype.Int4{Int: 5, Status: pgtype.Present},
					IsArchived:     pgtype.Bool{Bool: false, Status: pgtype.Present},
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
		err := packageDiscountSettingRepo.Upsert(ctx, db, pgtype.Text{String: "1", Status: pgtype.Present}, testCase.Req.([]*entities.PackageDiscountSetting))
		if testCase.ExpectedErr != nil {
			assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.ExpectedErr, err)
		}
	}
}
