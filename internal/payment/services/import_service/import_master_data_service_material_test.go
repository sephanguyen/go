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

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestImportMaterial(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	db := new(mockDb.Ext)
	tx := new(mockDb.Tx)

	materialRepo := new(mockRepositories.MockMaterialRepo)
	productSettingRepo := new(mockRepositories.MockProductSettingRepo)

	s := &ImportMasterDataService{
		DB:                 db,
		MaterialRepo:       materialRepo,
		ProductSettingRepo: productSettingRepo,
	}

	columnNames := []string{
		"material_id",
		"name",
		"material_type",
		"tax_id",
		"product_tag",
		"product_partner_id",
		"available_from",
		"available_until",
		"custom_billing_period",
		"custom_billing_date",
		"billing_schedule_id",
		"disable_pro_rating_flag",
		"remarks",
		"is_archived",
		"is_unique",
	}

	testcases := []utils.TestCase{
		{
			Name:        constant.NoDataInCsvFile,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req: &pb.ImportProductRequest{
				ProductType: pb.ProductType_PRODUCT_TYPE_MATERIAL,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - mismatched number of fields in header and content",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, ""),
			Req: &pb.ImportProductRequest{
				ProductType: pb.ProductType_PRODUCT_TYPE_MATERIAL,
				Payload: []byte(`material_id,name,material_type,tax_id,product_tag,product_partner_id,available_from,available_until,custom_billing_period,custom_billing_date,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived,is_unique
				1,Cat 1,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 1
				2,Cat 2,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 2
				3,Cat 3,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 3`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - number of column != 15 - missing is_archived",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - number of column should be 15"),
			Req: &pb.ImportProductRequest{
				ProductType: pb.ProductType_PRODUCT_TYPE_MATERIAL,
				Payload: []byte(`material_id,name,material_type,tax_id,product_tag,product_partner_id,available_from,available_until,custom_billing_period,custom_billing_date,billing_schedule_id,disable_pro_rating_flag,remarks,is_unique
				1,Cat 1,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 1,0
				2,Cat 2,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 2,0
				3,Cat 3,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 3,0`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "parsing valid file - create a new material success",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportProductRequest{
				ProductType: pb.ProductType_PRODUCT_TYPE_MATERIAL,
				Payload: []byte(`material_id,name,material_type,tax_id,product_tag,product_partner_id,available_from,available_until,custom_billing_period,custom_billing_date,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived,is_unique
				,Cat 1,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 1,0,0
				,Cat 2,1,0,,,2021-12-07,2021-12-08,2021-12-09,2021-12-09,5,1,Remarks 1,0,0`),
			},
			ExpectedResp: &pb.ImportProductResponse{},
			Setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				materialRepo.On("Create", ctx, tx, mock.Anything).Return(nil)
				productSettingRepo.On("Create", ctx, tx, mock.Anything).Return(nil)
			},
		},
		{
			Name: "parsing valid file - create a new material success",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportProductRequest{
				ProductType: pb.ProductType_PRODUCT_TYPE_MATERIAL,
				Payload: []byte(`material_id,name,material_type,tax_id,product_tag,product_partner_id,available_from,available_until,custom_billing_period,custom_billing_date,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived,is_unique
				,Cat 1,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 2,0,0`),
			},
			ExpectedResp: &pb.ImportProductResponse{},
			Setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				materialRepo.On("Create", ctx, tx, mock.Anything).Return(nil)
				productSettingRepo.On("Create", ctx, tx, mock.Anything).Return(nil)
			},
		},
		{
			Name: "parsing valid file - create a new material with invalid `archived` value",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportProductRequest{
				ProductType: pb.ProductType_PRODUCT_TYPE_MATERIAL,
				Payload: []byte(`material_id,name,material_type,tax_id,product_tag,product_partner_id,available_from,available_until,custom_billing_period,custom_billing_date,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived,is_unique
				,Cat 2,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 2,Archived,0`),
			},
			ExpectedResp: &pb.ImportProductResponse{
				Errors: []*pb.ImportProductResponse_ImportProductError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("unable to parse material item: %s", fmt.Errorf("error parsing is_archived")),
					},
				},
			},
			Setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				materialRepo.On("Create", ctx, tx, mock.Anything).Return(nil)
			},
		},
		{
			Name: "parsing valid file - update material with missing mandatory column",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportProductRequest{
				ProductType: pb.ProductType_PRODUCT_TYPE_MATERIAL,
				Payload: []byte(`material_id,name,material_type,tax_id,product_tag,product_partner_id,available_from,available_until,custom_billing_period,custom_billing_date,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived,is_unique
				3,Cat 2,,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 2,0,0`),
			},
			ExpectedResp: &pb.ImportProductResponse{
				Errors: []*pb.ImportProductResponse_ImportProductError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("unable to parse material item: %s", fmt.Errorf("missing mandatory data: material_type")),
					},
				},
			},
			Setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
			},
		},
		{
			Name: "parsing valid file - create material with tx error",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportProductRequest{
				ProductType: pb.ProductType_PRODUCT_TYPE_MATERIAL,
				Payload: []byte(`material_id,name,material_type,tax_id,product_tag,product_partner_id,available_from,available_until,custom_billing_period,custom_billing_date,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived,is_unique
				,Cat 2,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 4,0,0`),
			},
			ExpectedResp: &pb.ImportProductResponse{
				Errors: []*pb.ImportProductResponse_ImportProductError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("unable to create material item: %s", pgx.ErrTxClosed),
					},
				},
			},
			Setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				materialRepo.On("Create", ctx, tx, mock.Anything).Return(pgx.ErrTxClosed)
			},
		},
		{
			Name: "parsing valid file - update material with tx error",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportProductRequest{
				ProductType: pb.ProductType_PRODUCT_TYPE_MATERIAL,
				Payload: []byte(`material_id,name,material_type,tax_id,product_tag,product_partner_id,available_from,available_until,custom_billing_period,custom_billing_date,billing_schedule_id,disable_pro_rating_flag,remarks,is_archived,is_unique
				4,Cat 2,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 5,1,0`),
			},
			ExpectedResp: &pb.ImportProductResponse{
				Errors: []*pb.ImportProductResponse_ImportProductError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("unable to update material item: %s", pgx.ErrTxClosed),
					},
				},
			},
			Setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				materialRepo.On("Update", ctx, tx, mock.Anything).Return(pgx.ErrTxClosed)
			},
		},
	}
	testcases = append(testcases, GenerateProductWrongColumNameTestCases(
		ctx,
		pb.ProductType_PRODUCT_TYPE_MATERIAL,
		columnNames,
		`1,Cat 1,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 1,1,0
		2,Cat 2,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 2,1,0
		3,Cat 3,1,0,,,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,2021-12-09T00:00:00-07:00,5,1,Remarks 3,1,0`,
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

			mock.AssertExpectationsForObjects(t, db, materialRepo, productSettingRepo)
		})
	}
}
