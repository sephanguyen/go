package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/payment/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestImportAssociatedProducts_Material(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	mockDb := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)
	mockPackageCourseMaterialRepo := new(mock_repositories.MockPackageCourseMaterialRepo)

	s := &ImportMasterDataService{
		DB:                        mockDb,
		PackageCourseMaterialRepo: mockPackageCourseMaterialRepo,
	}

	testcases := []utils.TestCase{
		{
			Name:        "associated products type is none",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "invalid associated products type"),
			Req: &pb.ImportAssociatedProductsRequest{
				AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_NONE,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        constant.NoDataInCsvFile,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req: &pb.ImportAssociatedProductsRequest{
				AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "only headers in csv file",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req: &pb.ImportAssociatedProductsRequest{
				AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
				Payload:                []byte(`package_id,course_id,material_id,available_from,available_until`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "wrong number of data in a record",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportAssociatedProductsRequest{
				AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
				Payload: []byte(`package_id,course_id,material_id,available_from,available_until
				1,Course-2,3`),
			},
			ExpectedErr: status.Error(codes.InvalidArgument, "record on line 2: wrong number of fields"),
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - number of columns != 6",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - number of column should be 6"),
			Req: &pb.ImportAssociatedProductsRequest{
				AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
				Payload: []byte(`package_id,course_id,material_id
				1,Course-2,3`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - first column name (toLowerCase) != package_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - first column (toLowerCase) should be 'package_id'"),
			Req: &pb.ImportAssociatedProductsRequest{
				AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
				Payload: []byte(`wrong_header,course_id,material_id,available_from,available_until,is_added_by_default
				1,Course-2,3,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,true`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - second column name (toLowerCase) != course_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - second column (toLowerCase) should be 'course_id'"),
			Req: &pb.ImportAssociatedProductsRequest{
				AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
				Payload: []byte(`package_id,wrong_header,material_id,available_from,available_until,is_added_by_default
				1,Course-2,3,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,true`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - third column name (toLowerCase) != material_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - third column (toLowerCase) should be 'material_id'"),
			Req: &pb.ImportAssociatedProductsRequest{
				AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
				Payload: []byte(`package_id,course_id,wrong_header,available_from,available_until,is_added_by_default
				1,Course-2,3,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,false`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - fourth column name (toLowerCase) != available_from",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - fourth column (toLowerCase) should be 'available_from'"),
			Req: &pb.ImportAssociatedProductsRequest{
				AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
				Payload: []byte(`package_id,course_id,material_id,wrong_header,available_until,is_added_by_default
				1,Course-2,3,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,false`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - fifth column name (toLowerCase) != available_until",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - fifth column (toLowerCase) should be 'available_until'"),
			Req: &pb.ImportAssociatedProductsRequest{
				AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
				Payload: []byte(`package_id,course_id,material_id,available_from,wrong_header,is_added_by_default
				1,Course-2,3,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,false`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "parsing associated products by material with error lines in response",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportAssociatedProductsRequest{
				AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
				Payload: []byte(`package_id,course_id,material_id,available_from,available_until,is_added_by_default
				1,Course-2,3,d,2022-12-07,true
				1,Course-2,3,2021-12-07,e,false
				,Course-2,3,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,true
				1,,3,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,false
				1,Course-2,,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,true
				1,Course-2,3,,2022-12-07T00:00:00-07:00,false
				1,Course-2,3,2022-12-07T00:00:00-07:00,,true`),
			},
			ExpectedResp: &pb.ImportAssociatedProductsResponse{
				Errors: []*pb.ImportAssociatedProductsResponse_ImportAssociatedProductsError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf(constant.UnableToParseAssociatedProductsByMaterial, fmt.Errorf("error parsing available_from: parsing time \"d\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"d\" as \"2006\"")),
					},
					{
						RowNumber: 3,
						Error:     fmt.Sprintf(constant.UnableToParseAssociatedProductsByMaterial, fmt.Errorf("error parsing available_until: parsing time \"e\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"e\" as \"2006\"")),
					},
					{
						RowNumber: 4,
						Error:     fmt.Sprintf(constant.UnableToParseAssociatedProductsByMaterial, fmt.Errorf("missing mandatory data: package_id")),
					},
					{
						RowNumber: 5,
						Error:     fmt.Sprintf(constant.UnableToParseAssociatedProductsByMaterial, fmt.Errorf("missing mandatory data: course_id")),
					},
					{
						RowNumber: 6,
						Error:     fmt.Sprintf(constant.UnableToParseAssociatedProductsByMaterial, fmt.Errorf("missing mandatory data: material_id")),
					},
					{
						RowNumber: 7,
						Error:     fmt.Sprintf(constant.UnableToParseAssociatedProductsByMaterial, fmt.Errorf("missing mandatory data: available_from")),
					},
					{
						RowNumber: 8,
						Error:     fmt.Sprintf(constant.UnableToParseAssociatedProductsByMaterial, fmt.Errorf("missing mandatory data: available_until")),
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockTx.On("Rollback", mock.Anything).Return(nil)
				mockDb.On("Begin", mock.Anything).Return(mockTx, nil)
			},
		},
		{
			Name: "parsing valid csv rows",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportAssociatedProductsRequest{
				AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
				Payload: []byte(`package_id,course_id,material_id,available_from,available_until,is_added_by_default
				1,Course-2,3,2021-12-07,2022-12-07,true
				1,Course-3,4,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,false
				2,Course-4,5,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,true
				3,Course-5,6,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,false`),
			},
			ExpectedResp: &pb.ImportAssociatedProductsResponse{
				Errors: []*pb.ImportAssociatedProductsResponse_ImportAssociatedProductsError{},
			},
			Setup: func(ctx context.Context) {
				mockTx.On("Commit", mock.Anything).Return(nil)
				mockDb.On("Begin", mock.Anything).Return(mockTx, nil)
				mockPackageCourseMaterialRepo.On("Upsert", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPackageCourseMaterialRepo.On("Upsert", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPackageCourseMaterialRepo.On("Upsert", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			Name: "parsing valid csv rows but fail on import",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportAssociatedProductsRequest{
				AssociatedProductsType: pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL,
				Payload: []byte(`package_id,course_id,material_id,available_from,available_until,is_added_by_default
				1,Course-2,3,2021-12-07T00:00:00-07:00,2022-12-07T00:00:00-07:00,true`),
			},
			ExpectedResp: &pb.ImportAssociatedProductsResponse{
				Errors: []*pb.ImportAssociatedProductsResponse_ImportAssociatedProductsError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("unable to create new associated products by material: error something"),
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockTx.On("Rollback", mock.Anything).Return(nil)
				mockDb.On("Begin", mock.Anything).Return(mockTx, nil)
				mockPackageCourseMaterialRepo.On("Upsert", ctx, mockTx, mock.Anything, mock.Anything).Return(fmt.Errorf("error something"))
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			resp, err := s.ImportAssociatedProducts(testCase.Ctx, testCase.Req.(*pb.ImportAssociatedProductsRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)

				for i, expectedErr := range testCase.ExpectedResp.(*pb.ImportAssociatedProductsResponse).Errors {
					assert.Equal(t, expectedErr.RowNumber, resp.Errors[i].RowNumber)
					assert.Contains(t, expectedErr.Error, resp.Errors[i].Error)
				}
			}

			mock.AssertExpectationsForObjects(t, mockDb, mockPackageCourseMaterialRepo)
		})
	}
}
