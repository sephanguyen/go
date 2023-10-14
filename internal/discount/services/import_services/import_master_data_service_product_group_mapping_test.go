package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/internal/discount/utils"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mockRepositories "github.com/manabie-com/backend/mock/discount/repositories"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	pb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestImportProductGroupMapping(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	db := new(mockDb.Ext)
	mockTx := new(mockDb.Tx)
	mockProductGroupMappingRepo := new(mockRepositories.MockProductGroupMappingRepo)

	s := &ImportMasterDataService{
		DB:                      db,
		ProductGroupMappingRepo: mockProductGroupMappingRepo,
	}

	testcases := []utils.TestCase{
		{
			Name:         constant.HappyCase,
			Ctx:          interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedResp: &pb.ImportProductGroupMappingResponse{Errors: []*pb.ImportProductGroupMappingResponse_ImportProductGroupMappingError{}},
			Req: &pb.ImportProductGroupMappingRequest{
				Payload: []byte(`product_group_id,product_id
				1,product-1
				1,product-2
				2,product-3
				2,product-4`),
			},
			Setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				mockProductGroupMappingRepo.On("Upsert", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockProductGroupMappingRepo.On("Upsert", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			Name:        constant.NoDataInCsvFile,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req:         &pb.ImportProductGroupMappingRequest{},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - mismatched number of fields in header and content",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "record on line 2: wrong number of fields"),
			Req: &pb.ImportProductGroupMappingRequest{
				Payload: []byte(`product_group_id,product_id
				1
				2`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - number of column != 2 - missing product_group_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - number of column should be 2"),
			Req: &pb.ImportProductGroupMappingRequest{
				Payload: []byte(`product_id
				product-1
				product-2
				product-3`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - first column name (toLowerCase) != product_group_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - first column (toLowerCase) should be 'product_group_id'"),
			Req: &pb.ImportProductGroupMappingRequest{
				Payload: []byte(`incorrect_product_group_id,product_id
				1,product-1
				1,product-2
				2,product-3
				2,product-4`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - second column name (toLowerCase) != prodict_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - second column (toLowerCase) should be 'product_id'"),
			Req: &pb.ImportProductGroupMappingRequest{
				Payload: []byte(`product_group_id,incorrect_product_id
				1,product-1
				1,product-2
				2,product-3
				2,product-4`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "parsing valid file (with error lines in response)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportProductGroupMappingRequest{
				Payload: []byte(`product_group_id,product_id
				1,
				,product-2
				2,product-3
				2,product-4`),
			},
			ExpectedResp: &pb.ImportProductGroupMappingResponse{
				Errors: []*pb.ImportProductGroupMappingResponse_ImportProductGroupMappingError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("unable to import product group mapping item: %s", fmt.Errorf("missing mandatory data: product_id")),
					},
					{
						RowNumber: 3,
						Error:     fmt.Sprintf("unable to import product group mapping item: %s", fmt.Errorf("missing mandatory data: product_group_id")),
					},
					{
						RowNumber: 5,
						Error:     fmt.Sprintf("unable to import product group mapping item: %s", fmt.Errorf("tx is closed")),
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockTx.On("Rollback", mock.Anything).Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(mockTx, nil)
				mockProductGroupMappingRepo.On("Upsert", ctx, mockTx, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			resp, err := s.ImportProductGroupMapping(testCase.Ctx, testCase.Req.(*pb.ImportProductGroupMappingRequest))
			fmt.Println(err)
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.ExpectedResp.(*pb.ImportProductGroupMappingResponse)
				for i, err := range resp.Errors {
					assert.Equal(t, expectedResp.Errors[i].RowNumber, err.RowNumber)
					assert.Contains(t, err.Error, expectedResp.Errors[i].Error)
				}
			}

			mock.AssertExpectationsForObjects(t, db, mockProductGroupMappingRepo)
		})
	}
}
