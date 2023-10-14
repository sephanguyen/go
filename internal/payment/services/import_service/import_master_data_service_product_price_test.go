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

func TestImportProductPrice(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	db := new(mockDb.Ext)
	tx := new(mockDb.Tx)
	productPriceRepo := new(mockRepositories.MockProductPriceRepo)

	s := &ImportMasterDataService{
		DB:               db,
		ProductPriceRepo: productPriceRepo,
	}

	testcases := []utils.TestCase{
		{
			Name:        constant.NoDataInCsvFile,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req:         &pb.ImportProductPriceRequest{},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - number of column != 5",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - number of column should be 5"),
			Req: &pb.ImportProductPriceRequest{
				Payload: []byte(`product_id,quantity,price,price_type
				1,3,12.25,DEFAULT_PRICE`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - wrong name column",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - first column (toLowerCase) should be 'product_id'"),
			Req: &pb.ImportProductPriceRequest{
				Payload: []byte(`some_id,billing_schedule_period_id,quantity,price,price_type
				1,1,3,12.25,DEFAULT_PRICE`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - mismatched number of fields in header and content",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "record on line 2: wrong number of fields"),
			Req: &pb.ImportProductPriceRequest{
				Payload: []byte(`product_id,billing_schedule_period_id,quantity,price,price_type
				1,1,12.25,DEFAULT_PRICE`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail when parse product price column",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &pb.ImportProductPriceRequest{
				Payload: []byte(`product_id,billing_schedule_period_id,quantity,price,price_type
				1,1,3,12..25,DEFAULT_PRICE`),
			},
			ExpectedResp: &pb.ImportProductPriceResponse{Errors: []*pb.ImportProductPriceResponse_ImportProductPriceError{
				{
					RowNumber: 2,
					Error:     fmt.Sprintf("unable to parse product_price item: %s", fmt.Errorf("error parsing price")),
				},
			}},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				productPriceRepo.On("DeleteByProductID", ctx, tx, mock.Anything).Once().Return(nil)
			},
		},
		{
			Name:        "Fail when import product price due to db Create error",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &pb.ImportProductPriceRequest{
				Payload: []byte(`product_id,billing_schedule_period_id,quantity,price,price_type
				1,1,3,12.25,DEFAULT_PRICE`),
			},
			ExpectedResp: &pb.ImportProductPriceResponse{Errors: []*pb.ImportProductPriceResponse_ImportProductPriceError{
				{
					RowNumber: 2,
					Error:     `unable to create new product_price item: mock error`,
				},
			}},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				productPriceRepo.On("DeleteByProductID", ctx, tx, mock.Anything).Once().Return(nil)
				productPriceRepo.On("Create", ctx, tx, mock.Anything).Once().Return(fmt.Errorf("mock error"))
			},
		},
		{
			Name:        "Fail when import product price due to db DeleteByProductID error",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &pb.ImportProductPriceRequest{
				Payload: []byte(`product_id,billing_schedule_period_id,quantity,price,price_type
				1,1,3,12.25,DEFAULT_PRICE`),
			},
			ExpectedResp: &pb.ImportProductPriceResponse{Errors: []*pb.ImportProductPriceResponse_ImportProductPriceError{
				{
					RowNumber: 2,
					Error:     "something wrong when delete product_price",
				},
			}},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				productPriceRepo.On("DeleteByProductID", ctx, tx, mock.Anything).Once().Return(fmt.Errorf("mock error"))
			},
		},
		{
			Name:        "import product price - happy cases",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &pb.ImportProductPriceRequest{
				Payload: []byte(`product_id,billing_schedule_period_id,quantity,price,price_type
				1,1,2,7,DEFAULT_PRICE
				1,1,3,12.25,DEFAULT_PRICE
				2,,2,7,DEFAULT_PRICE
				3,,,7,DEFAULT_PRICE
				4,,,7,DEFAULT_PRICE
				5,1,2,7,DEFAULT_PRICE
				5,1,3,12.25,DEFAULT_PRICE`),
			},
			ExpectedResp: &pb.ImportProductPriceResponse{Errors: []*pb.ImportProductPriceResponse_ImportProductPriceError{}},
			Setup: func(ctx context.Context) {
				tx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				productPriceRepo.On("DeleteByProductID", ctx, tx, mock.Anything).Times(5).Return(nil)
				productPriceRepo.On("Create", ctx, tx, mock.Anything).Times(7).Return(nil)
			},
		},
		{
			Name:        "product_id is empty - fail case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &pb.ImportProductPriceRequest{
				Payload: []byte(`product_id,billing_schedule_period_id,quantity,price,price_type
				,1,3,12.25,DEFAULT_PRICE`),
			},
			ExpectedResp: &pb.ImportProductPriceResponse{Errors: []*pb.ImportProductPriceResponse_ImportProductPriceError{
				{
					RowNumber: 2,
					Error:     "product_id is empty",
				},
			}},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		// {
		// 	Name:        "product_id is not an int",
		// 	Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
		// 	ExpectedErr: nil,
		// 	Req: &pb.ImportProductPriceRequest{
		// 		Payload: []byte(`product_id,billing_schedule_period_id,quantity,price
		// 		a,1,3,12.25`),
		// 	},
		// 	ExpectedResp: &pb.ImportProductPriceResponse{Errors: []*pb.ImportProductPriceResponse_ImportProductPriceError{
		// 		{
		// 			RowNumber: 2,
		// 			Error:     "wrong product_id format",
		// 		},
		// 	}},
		// 	Setup: func(ctx context.Context) {
		// 		tx.On("Rollback", mock.Anything).Return(nil)
		// 		db.On("Begin", mock.Anything).Return(tx, nil)
		// 	},
		// },
		{
			Name:        "invalid data and valid data",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &pb.ImportProductPriceRequest{
				Payload: []byte(`product_id,billing_schedule_period_id,quantity,price,price_type
				1,1,2,7,DEFAULT_PRICE
				,1,3,12.25,DEFAULT_PRICE
				3,,,7,DEFAULT_PRICE`),
			},
			ExpectedResp: &pb.ImportProductPriceResponse{Errors: []*pb.ImportProductPriceResponse_ImportProductPriceError{
				// {
				// 	RowNumber: 2,
				// 	Error:     "wrong product_id format",
				// },
				{
					RowNumber: 3,
					Error:     "product_id is empty",
				},
			}},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				productPriceRepo.On("DeleteByProductID", ctx, tx, mock.Anything).Return(nil)
				productPriceRepo.On("Create", ctx, tx, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "invalid data and valid data",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, fmt.Sprintf("error when import product prices without DEFAULT_PRICE value with product_ids")),
			Req: &pb.ImportProductPriceRequest{
				Payload: []byte(`product_id,billing_schedule_period_id,quantity,price,price_type
				1,1,2,7,ENROLLED_PRICE
				2,1,2,7,ENROLLED_PRICE
				2,1,2,7,DEFAULT_PRICE
				3,1,2,7,ENROLLED_PRICE`),
			},
			ExpectedResp: &pb.ImportProductPriceResponse{Errors: []*pb.ImportProductPriceResponse_ImportProductPriceError{}},
			Setup: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			resp, err := s.ImportProductPrice(testCase.Ctx, testCase.Req.(*pb.ImportProductPriceRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.ExpectedResp.(*pb.ImportProductPriceResponse)
				for i, err := range resp.Errors {
					assert.Equal(t, err.RowNumber, expectedResp.Errors[i].RowNumber)
					assert.Contains(t, err.Error, expectedResp.Errors[i].Error)
				}
			}

			mock.AssertExpectationsForObjects(t, db, productPriceRepo)
		})
	}
}
