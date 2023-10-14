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

func TestImportProduct_AccountingCategory(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	db := new(mockDb.Ext)
	mockTx := new(mockDb.Tx)
	mockProductAccountingCategoryRepo := new(mockRepositories.MockProductAccountingCategoryRepo)

	s := &ImportMasterDataService{
		DB:                            db,
		ProductAccountingCategoryRepo: mockProductAccountingCategoryRepo,
	}

	testcases := []utils.TestCase{
		{
			Name:        "product associated data type is none",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "invalid product associated data type"),
			Req: &pb.ImportProductAssociatedDataRequest{
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_NONE,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        constant.NoDataInCsvFile,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req: &pb.ImportProductAssociatedDataRequest{
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_ACCOUNTING_CATEGORY,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "only headers in csv file",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req: &pb.ImportProductAssociatedDataRequest{
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_ACCOUNTING_CATEGORY,
				Payload:                   []byte(`product_id,accounting_category_id`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "wrong number of data in a record",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportProductAssociatedDataRequest{
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_ACCOUNTING_CATEGORY,
				Payload: []byte(`product_id,accounting_category_id
				1`),
			},
			ExpectedErr: status.Error(codes.InvalidArgument, "record on line 2: wrong number of fields"),
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - number of columns != 2",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - number of columns should be 2"),
			Req: &pb.ImportProductAssociatedDataRequest{
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_ACCOUNTING_CATEGORY,
				Payload: []byte(`product_id
				1
				2
				3`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - first column name (toLowerCase) != product_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - first column (toLowerCase) should be 'product_id'"),
			Req: &pb.ImportProductAssociatedDataRequest{
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_ACCOUNTING_CATEGORY,
				Payload: []byte(`wrong_header,accounting_category_id
				1,1
				2,1`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - second column name (toLowerCase) != accounting_category_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - second column (toLowerCase) should be 'accounting_category_id'"),
			Req: &pb.ImportProductAssociatedDataRequest{
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_ACCOUNTING_CATEGORY,
				Payload: []byte(`product_id,wrong_header
				1,1
				2,1`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "parsing product associated data accounting category with error lines in response",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportProductAssociatedDataRequest{
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_ACCOUNTING_CATEGORY,
				Payload: []byte(`product_id,accounting_category_id
				1,
				,1`),
			},
			ExpectedResp: &pb.ImportProductAssociatedDataResponse{
				Errors: []*pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError{
					// {
					// 	RowNumber: 2,
					// 	Error:     fmt.Sprintf("unable to parse product associated data item: %s", fmt.Errorf("error parsing product_id: strconv.Atoi: parsing \"a\": invalid syntax")),
					// },
					// {
					// 	RowNumber: 3,
					// 	Error:     fmt.Sprintf("unable to parse product associated data item: %s", fmt.Errorf("error parsing accounting_category_id: strconv.Atoi: parsing \"b\": invalid syntax")),
					// },
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("unable to parse product associated data item: %s", fmt.Errorf("missing mandatory data: accounting_category_id")),
					},
					{
						RowNumber: 3,
						Error:     fmt.Sprintf("unable to parse product associated data item: %s", fmt.Errorf("missing mandatory data: product_id column")),
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockTx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(mockTx, nil)
			},
		},
		{
			Name: "parsing valid csv rows",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportProductAssociatedDataRequest{
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_ACCOUNTING_CATEGORY,
				Payload: []byte(`product_id,accounting_category_id
				1,1
				1,2
				2,2
				3,3`),
			},
			ExpectedResp: &pb.ImportProductAssociatedDataResponse{
				Errors: []*pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError{},
			},
			Setup: func(ctx context.Context) {
				mockTx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(mockTx, nil)
				mockProductAccountingCategoryRepo.On("Upsert", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockProductAccountingCategoryRepo.On("Upsert", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockProductAccountingCategoryRepo.On("Upsert", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			Name: "parsing valid csv rows but fail on import",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportProductAssociatedDataRequest{
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_ACCOUNTING_CATEGORY,
				Payload: []byte(`product_id,accounting_category_id
				1,1`),
			},
			ExpectedResp: &pb.ImportProductAssociatedDataResponse{
				Errors: []*pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("unable to create new product accounting category item: error something"),
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockTx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(mockTx, nil)
				mockProductAccountingCategoryRepo.On("Upsert", ctx, mockTx, mock.Anything, mock.Anything).Return(fmt.Errorf("error something"))
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

			mock.AssertExpectationsForObjects(t, db, mockProductAccountingCategoryRepo)
		})
	}
}
