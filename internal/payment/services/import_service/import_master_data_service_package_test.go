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

func TestImportPackage(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	db := new(mockDb.Ext)
	tx := new(mockDb.Tx)
	packageRepo := new(mockRepositories.MockPackageRepo)
	productSettingRepo := new(mockRepositories.MockProductSettingRepo)

	s := &ImportMasterDataService{
		DB:                 db,
		PackageRepo:        packageRepo,
		ProductSettingRepo: productSettingRepo,
	}

	testcases := []utils.TestCase{
		{
			Name:        "package type is none",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "package type is none"),
			Req: &pb.ImportProductRequest{
				ProductType: pb.ProductType_PRODUCT_TYPE_NONE,
			},
			ExpectedResp: &pb.ImportProductResponse{Errors: nil},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        constant.NoDataInCsvFile,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req: &pb.ImportProductRequest{
				ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "miss some column",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "record on line 2: wrong number of fields"),
			Req: &pb.ImportProductRequest{
				Payload: []byte(`package_id,name,package_type,tax_id,available_from,available_until,max_slot,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,package_start_date,package_end_date,remarks
,Package %s,1,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,2,2022-10-07T00:00:00-07:00,,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0
`),
				ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "wrong name column",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "rpc error: code = InvalidArgument desc = csv file invalid format - first column (toLowerCase) should be 'package_id'"),
			Req: &pb.ImportProductRequest{
				Payload: []byte(`some_id,name,package_type,tax_id,product_tag,product_partner_id,available_from,available_until,max_slot,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,package_start_date,package_end_date,remarks,is_archived,is_unique
,Package %s,1,,,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,2,2022-10-07T00:00:00-07:00,,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,0
`),
				ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "mismatched number of fields in header and content",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "record on line 2: wrong number of fields"),
			Req: &pb.ImportProductRequest{
				Payload: []byte(`package_id,name,package_type,tax_id,available_from,available_until,max_slot,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,package_start_date,package_end_date,remarks,is_archived,is_unique
,Package %s,1,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,2,2022-10-07T00:00:00-07:00,,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0
`),
				ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail when parse product",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &pb.ImportProductRequest{
				Payload: []byte(`package_id,name,package_type,tax_id,product_tag,product_partner_id,available_from,available_until,max_slot,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,package_start_date,package_end_date,remarks,is_archived,is_unique
,Package %s,1,,,,34,2022-10-07T00:00:00-07:00,2,2022-10-07T00:00:00-07:00,,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,0
`),
				ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
			},
			ExpectedResp: &pb.ImportProductResponse{Errors: []*pb.ImportProductResponse_ImportProductError{
				{
					RowNumber: int32(2),
					Error:     `unable to parse package item: error parsing available_from: parsing time "34" as "2006-01-02T15:04:05Z07:00": cannot parse "34" as "2006"`,
				},
			}},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name:        "Fail when parse package",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &pb.ImportProductRequest{
				Payload: []byte(`package_id,name,package_type,tax_id,product_tag,product_partner_id,available_from,available_until,max_slot,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,package_start_date,package_end_date,remarks,is_archived,is_unique
,Package %s,6,,,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,2,2022-10-07T00:00:00-07:00,,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,0
`),
				ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
			},
			ExpectedResp: &pb.ImportProductResponse{Errors: []*pb.ImportProductResponse_ImportProductError{
				{
					RowNumber: int32(2),
					Error:     `unable to parse package item: invalid package_type: 6`,
				},
			}},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name:        "Fail when create product",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &pb.ImportProductRequest{
				Payload: []byte(`package_id,name,package_type,tax_id,product_tag,product_partner_id,available_from,available_until,max_slot,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,package_start_date,package_end_date,remarks,is_archived,is_unique
,Package %s,1,,,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,2,2022-10-07T00:00:00-07:00,,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,0
`),
				ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
			},
			ExpectedResp: &pb.ImportProductResponse{Errors: []*pb.ImportProductResponse_ImportProductError{
				{
					RowNumber: int32(2),
					Error:     `unable to insert package item: error something`,
				},
			}},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				packageRepo.On("Create", ctx, tx, mock.Anything).Once().Return(fmt.Errorf("error something"))
			},
		},
		{
			Name:        "Happy case create package",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &pb.ImportProductRequest{
				Payload: []byte(`package_id,name,package_type,tax_id,product_tag,product_partner_id,available_from,available_until,max_slot,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,package_start_date,package_end_date,remarks,is_archived,is_unique
,Package %s,1,,,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,2,2022-10-07T00:00:00-07:00,,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,0
,Package %s,1,,,,2021-12-08,2022-10-08,2,2022-10-08,,1,2021-12-08,2022-10-08,Remarks,0,0
`),
				ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
			},
			ExpectedResp: &pb.ImportProductResponse{Errors: []*pb.ImportProductResponse_ImportProductError{}},
			Setup: func(ctx context.Context) {
				tx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				productSettingRepo.On("Create", ctx, tx, mock.Anything).Return(nil)
				packageRepo.On("Create", ctx, tx, mock.Anything).Twice().Return(nil)
			},
		},
		{
			Name:        "Fail when update product",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &pb.ImportProductRequest{
				Payload: []byte(`package_id,name,package_type,tax_id,product_tag,product_partner_id,available_from,available_until,max_slot,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,package_start_date,package_end_date,remarks,is_archived,is_unique
1,Package %s,1,,,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,2,2022-10-07T00:00:00-07:00,,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,0
`),
				ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
			},
			ExpectedResp: &pb.ImportProductResponse{Errors: []*pb.ImportProductResponse_ImportProductError{
				{
					RowNumber: int32(2),
					Error:     `unable to update package item: error something`,
				},
			}},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				packageRepo.On("Update", ctx, tx, mock.Anything).Once().Return(fmt.Errorf("error something"))
			},
		},
		{
			Name:        "Fail when missing package_start_date/package_end_date with one_time/base_slot package",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &pb.ImportProductRequest{
				Payload: []byte(`package_id,name,package_type,tax_id,product_tag,product_partner_id,available_from,available_until,max_slot,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,package_start_date,package_end_date,remarks,is_archived,is_unique
,Package %s,1,,,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,2,2022-10-07T00:00:00-07:00,,1,,2022-10-07T00:00:00-07:00,Remarks,0,0
1,Package %s,1,,,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,2,2022-10-07T00:00:00-07:00,,1,2021-12-07T00:00:00-07:00,,Remarks,0,0
1,Package %s,1,,,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,2,2022-10-07T00:00:00-07:00,,1,,,Remarks,0,0
`),
				ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
			},
			ExpectedResp: &pb.ImportProductResponse{Errors: []*pb.ImportProductResponse_ImportProductError{
				{
					RowNumber: int32(2),
					Error:     `package_start_date is missing`,
				},
				{
					RowNumber: int32(3),
					Error:     `package_end_date is missing`,
				},
				{
					RowNumber: int32(4),
					Error:     `package_start_date, package_end_date are missing`,
				},
			}},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name:        "Happy case update package",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &pb.ImportProductRequest{
				Payload: []byte(`package_id,name,package_type,tax_id,product_tag,product_partner_id,available_from,available_until,max_slot,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,package_start_date,package_end_date,remarks,is_archived,is_unique
1,Package %s,1,,,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,2,2022-10-07T00:00:00-07:00,,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,0
1,Package %s,1,,,,2021-12-08,2022-10-08,2,2022-10-08,,1,2021-12-08,2022-10-08,Remarks,0,0
`),
				ProductType: pb.ProductType_PRODUCT_TYPE_PACKAGE,
			},
			ExpectedResp: &pb.ImportProductResponse{Errors: []*pb.ImportProductResponse_ImportProductError{}},
			Setup: func(ctx context.Context) {
				tx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				packageRepo.On("Update", ctx, tx, mock.Anything).Twice().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			resp, err := s.ImportProduct(testCase.Ctx, testCase.Req.(*pb.ImportProductRequest))
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Equal(t, &pb.ImportProductResponse{Errors: nil}, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.ExpectedResp.(*pb.ImportProductResponse)
				for i, err := range resp.Errors {
					assert.Equal(t, expectedResp.Errors[i].RowNumber, err.RowNumber)
					assert.Equal(t, err.Error, expectedResp.Errors[i].Error)
				}
			}

			mock.AssertExpectationsForObjects(t, db, packageRepo, productSettingRepo)
		})
	}
}
