package service

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExportService_ExportMasterData(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db *mockDb.Ext
	)
	mockDB := testutil.NewMockDB()
	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when export master data",
			Ctx:  ctx,
			Req: &pb.ExportMasterDataRequest{
				ExportDataType: pb.ExportMasterDataType_EXPORT_DISCOUNT,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				rows := mockDB.Rows
				db.On("Query", mock.Anything, mock.Anything).Return(rows, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case (data_type = 'discount')",
			Ctx:  ctx,
			Req: &pb.ExportMasterDataRequest{
				ExportDataType: pb.ExportMasterDataType_EXPORT_DISCOUNT,
			},
			Setup: func(ctx context.Context) {
				rows := mockDB.Rows
				db.On("Query", mock.Anything, mock.Anything).Return(rows, nil)

				rows.On("Next").Times(1).Return(true)

				discountMasterData := &entities.Discount{}
				fields, _ := discountMasterData.FieldMap()
				scanFields := database.GetScanFields(discountMasterData, fields)
				rows.On("Scan", scanFields...).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			testCase.Setup(testCase.Ctx)
			s := &ExportService{
				DB: db,
			}

			response, err := s.ExportMasterData(testCase.Ctx, testCase.Req.(*pb.ExportMasterDataRequest))

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, response)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, response)
			}

			mock.AssertExpectationsForObjects(t, db)

		})
	}

}
