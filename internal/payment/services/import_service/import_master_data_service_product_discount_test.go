package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestImportProductDiscount(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	db := new(mockDb.Ext)
	tx := new(mockDb.Tx)
	productDiscountRepo := new(mockRepositories.MockProductDiscountRepo)

	s := &ImportMasterDataService{
		DB:                  db,
		ProductDiscountRepo: productDiscountRepo,
	}

	testcases := []utils.TestCase{
		{
			Name:        constant.NoDataInCsvFile,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req: &pb.ImportProductAssociatedDataRequest{
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_DISCOUNT,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - mismatched number of fields in header and content",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "record on line 2: wrong number of fields"),
			Req: &pb.ImportProductAssociatedDataRequest{
				Payload: []byte(`product_id,discount_id
				,1,`),
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_DISCOUNT,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - number of column != 2 - missing product_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "record on line 2: wrong number of fields"),
			Req: &pb.ImportProductAssociatedDataRequest{
				Payload: []byte(`discount_id
				1,1`),
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_DISCOUNT,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "wrong name column",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - second column (toLowerCase) should be 'discount_id'"),
			Req: &pb.ImportProductAssociatedDataRequest{
				Payload: []byte(`product_id,Number
				1,1
				2,2
				3,3`),
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_DISCOUNT,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - second column name (toLowerCase) != product_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - first column (toLowerCase) should be 'product_id'"),
			Req: &pb.ImportProductAssociatedDataRequest{
				Payload: []byte(`ProductID,discount_id
				1,1
				2,1
				3,2`),
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_DISCOUNT,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Happy case import product discount course success",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportProductAssociatedDataRequest{
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_DISCOUNT,
				Payload: []byte(`product_id,discount_id
				1,1
				1,2
				2,2
				3,3`),
			},
			ExpectedResp: &pb.ImportProductAssociatedDataResponse{
				Errors: []*pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError{},
			},
			Setup: func(ctx context.Context) {
				tx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				productDiscountRepo.On("Upsert", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				productDiscountRepo.On("Upsert", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				productDiscountRepo.On("Upsert", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			Name: "parsing valid csv rows but fail on import",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportProductAssociatedDataRequest{
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_DISCOUNT,
				Payload: []byte(`product_id,discount_id
				1,1`),
			},
			ExpectedResp: &pb.ImportProductAssociatedDataResponse{
				Errors: []*pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("unable to import product discount item: error something"),
					},
				},
			},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				productDiscountRepo.On("Upsert", ctx, tx, mock.Anything, mock.Anything).Return(fmt.Errorf("error something"))
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			resp, err := s.ImportProductAssociatedData(testCase.Ctx, testCase.Req.(*pb.ImportProductAssociatedDataRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
				for i, expectedErr := range testCase.ExpectedResp.(*pb.ImportProductAssociatedDataResponse).Errors {
					assert.Equal(t, expectedErr.RowNumber, resp.Errors[i].RowNumber)
					assert.Contains(t, expectedErr.Error, resp.Errors[i].Error)
				}
			}

			mock.AssertExpectationsForObjects(t, db, productDiscountRepo)
		})
	}
}
