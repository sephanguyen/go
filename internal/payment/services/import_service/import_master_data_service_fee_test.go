package service

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestImportFee(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	db := new(mockDb.Ext)
	tx := new(mockDb.Tx)

	feeRepo := new(mockRepositories.MockFeeRepo)
	productSettingRepo := new(mockRepositories.MockProductSettingRepo)

	s := &ImportMasterDataService{
		DB:                 db,
		FeeRepo:            feeRepo,
		ProductSettingRepo: productSettingRepo,
	}

	columnNames := []string{
		"fee_id",
		"name",
		"fee_type",
		"tax_id",
		"product_tag",
		"product_partner_id",
		"available_from",
		"available_until",
		"custom_billing_period",
		"billing_schedule_id",
		"disable_pro_rating_flag",
		"remarks",
		"is_archived",
		"is_unique",
	}

	testcases := []utils.TestCase{
		// {
		// 	Name:        "no data in csv file",
		// 	Ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
		// 	ExpectedErr: status.Error(codes.InvalidArgument, "no data in csv file"),
		// 	Req: &pb.ImportProductRequest{
		// 		ProductType: pb.ProductType_PRODUCT_TYPE_FEE,
		// 	},
		// 	Setup: func(ctx context.Context) {
		// 	},
		// },
		// {
		// 	Name:        "invalid file - mismatched number of fields in header and content",
		// 	Ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
		// 	ExpectedErr: status.Error(codes.InvalidArgument, ""),
		// 	Req: &pb.ImportProductRequest{
		// 		ProductType: pb.ProductType_PRODUCT_TYPE_FEE,
		// 		Payload: []byte(`fee_id,name,fee_type,tax_id,available_from,available_until,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived
		// 		1,Cat 1,1,0,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 1
		// 		2,Cat 2,1,0,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 2
		// 		3,Cat 3,1,0,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 3`),
		// 	},
		// 	Setup: func(ctx context.Context) {
		// 	},
		// },
		// {
		// 	Name:        "invalid file - number of column != 11 - missing is_archived",
		// 	Ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
		// 	ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - number of column should be 11"),
		// 	Req: &pb.ImportProductRequest{
		// 		ProductType: pb.ProductType_PRODUCT_TYPE_FEE,
		// 		Payload: []byte(`fee_id,name,fee_type,tax_id,available_from,available_until,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,remarks
		// 		1,Cat 1,1,0,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 1
		// 		2,Cat 2,1,0,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 2
		// 		3,Cat 3,1,0,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 3`),
		// 	},
		// 	Setup: func(ctx context.Context) {
		// 	},
		// },
		{
			Name: "parsing valid file - create a new fee success",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.ImportProductRequest{
				ProductType: pb.ProductType_PRODUCT_TYPE_FEE,
				Payload: []byte(`fee_id,name,fee_type,tax_id,product_tag,product_partner_id,available_from,available_until,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived,is_unique
				,Cat 1,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 1,0,0
				,Cat 2,1,0,,,2021-12-07,2021-12-08,2021-12-09,5,1,Remarks 1,0,0`),
			},
			ExpectedResp: &pb.ImportProductResponse{},
			Setup: func(ctx context.Context) {
				feeRepo.On("Create", ctx, tx, mock.Anything).Return(nil)
				productSettingRepo.On("Create", ctx, tx, mock.Anything).Return(nil)
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)

			},
		},
		// {
		// 	Name: "parsing valid file - create a new fee success 2",
		// 	Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
		// 	Req: &pb.ImportProductRequest{
		// 		ProductType: pb.ProductType_PRODUCT_TYPE_FEE,
		// 		Payload: []byte(`fee_id,name,fee_type,tax_id,available_from,available_until,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived
		// 		,Cat 1,1,0,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 2,0`),
		// 	},
		// 	ExpectedResp: &pb.ImportProductResponse{},
		// 	Setup: func(ctx context.Context) {
		// 		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		// 		tx.On("Commit", mock.Anything).Return(nil)
		// 		feeRepo.On("Create", ctx, tx, mock.Anything).Return(nil)
		// 	},
		// },
		// {
		// 	Name: "parsing valid file - create a new fee with invalid `archived` value",
		// 	Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
		// 	Req: &pb.ImportProductRequest{
		// 		ProductType: pb.ProductType_PRODUCT_TYPE_FEE,
		// 		Payload: []byte(`fee_id,name,fee_type,tax_id,available_from,available_until,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived
		// 		,Cat 2,1,0,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 2,Archived`),
		// 	},
		// 	ExpectedResp: &pb.ImportProductResponse{
		// 		Errors: []*pb.ImportProductResponse_ImportProductError{
		// 			{
		// 				RowNumber: 2,
		// 				Error:     fmt.Sprintf("unable to parse fee item: %s", fmt.Errorf("error parsing is_archived")),
		// 			},
		// 		},
		// 	},
		// 	Setup: func(ctx context.Context) {
		// 		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		// 		tx.On("Rollback", mock.Anything).Return(nil)
		// 		feeRepo.On("Create", ctx, tx, mock.Anything).Return(nil)
		// 	},
		// },
		// {
		// 	Name: "parsing valid file - update fee with missing mandatory column",
		// 	Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
		// 	Req: &pb.ImportProductRequest{
		// 		ProductType: pb.ProductType_PRODUCT_TYPE_FEE,
		// 		Payload: []byte(`fee_id,name,fee_type,tax_id,available_from,available_until,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived
		// 		3,Cat 2,,0,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 2,0`),
		// 	},
		// 	ExpectedResp: &pb.ImportProductResponse{
		// 		Errors: []*pb.ImportProductResponse_ImportProductError{
		// 			{
		// 				RowNumber: 2,
		// 				Error:     fmt.Sprintf("unable to parse fee item: %s", fmt.Errorf("missing mandatory data: fee_type")),
		// 			},
		// 		},
		// 	},
		// 	Setup: func(ctx context.Context) {
		// 		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		// 		tx.On("Rollback", mock.Anything).Return(nil)
		// 	},
		// },
		// {
		// 	Name: "parsing valid file - create fee with tx error",
		// 	Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
		// 	Req: &pb.ImportProductRequest{
		// 		ProductType: pb.ProductType_PRODUCT_TYPE_FEE,
		// 		Payload: []byte(`fee_id,name,fee_type,tax_id,available_from,available_until,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived
		// 		,Cat 2,1,0,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 4,0`),
		// 	},
		// 	ExpectedResp: &pb.ImportProductResponse{
		// 		Errors: []*pb.ImportProductResponse_ImportProductError{
		// 			{
		// 				RowNumber: 2,
		// 				Error:     fmt.Sprintf("unable to create fee item: %s", pgx.ErrTxClosed),
		// 			},
		// 		},
		// 	},
		// 	Setup: func(ctx context.Context) {
		// 		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		// 		tx.On("Rollback", mock.Anything).Return(nil)
		// 		feeRepo.On("Create", ctx, tx, mock.Anything).Return(pgx.ErrTxClosed)
		// 	},
		// },
		// {
		// 	Name: "parsing valid file - update fee with tx error",
		// 	Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
		// 	Req: &pb.ImportProductRequest{
		// 		ProductType: pb.ProductType_PRODUCT_TYPE_FEE,
		// 		Payload: []byte(`fee_id,name,fee_type,tax_id,available_from,available_until,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived
		// 		4,Cat 2,1,0,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 5,1`),
		// 	},
		// 	ExpectedResp: &pb.ImportProductResponse{
		// 		Errors: []*pb.ImportProductResponse_ImportProductError{
		// 			{
		// 				RowNumber: 2,
		// 				Error:     fmt.Sprintf("unable to update fee item: %s", pgx.ErrTxClosed),
		// 			},
		// 		},
		// 	},
		// 	Setup: func(ctx context.Context) {
		// 		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		// 		tx.On("Rollback", mock.Anything).Return(nil)
		// 		feeRepo.On("Update", ctx, tx, mock.Anything).Return(pgx.ErrTxClosed)
		// 	},
		// },
	}
	testcases = append(testcases, GenerateProductWrongColumNameTestCases(
		ctx,
		pb.ProductType_PRODUCT_TYPE_FEE,
		columnNames,
		`1,Cat 1,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 1,1,0
		2,Cat 2,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 2,1,0
		3,Cat 3,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 3,1,0`,
	)...)

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			resp, err := s.ImportProduct(testCase.Ctx, testCase.Req.(*pb.ImportProductRequest))
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.ExpectedResp.(*pb.ImportProductResponse)
				for i, err := range resp.Errors {
					assert.Equal(t, expectedResp.Errors[i].RowNumber, err.RowNumber)
					assert.Contains(t, err.Error, expectedResp.Errors[i].Error)
				}
			}

			mock.AssertExpectationsForObjects(t, db, feeRepo, productSettingRepo)
		})
	}
}
