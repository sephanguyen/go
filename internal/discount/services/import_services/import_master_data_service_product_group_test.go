package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/internal/discount/entities"
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

func TestImportProductGroup(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	db := new(mockDb.Ext)
	mockTx := new(mockDb.Tx)
	mockProductGroupRepo := new(mockRepositories.MockProductGroupRepo)

	s := &ImportMasterDataService{
		DB:               db,
		ProductGroupRepo: mockProductGroupRepo,
	}

	testcases := []utils.TestCase{
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedResp: &pb.ImportProductGroupResponse{Errors: []*pb.ImportProductGroupResponse_ImportProductGroupError{
				{
					RowNumber: 2,
					Error:     "unable to create new product group item: error something",
				},
				{
					RowNumber: 3,
					Error:     "unable to create new product group item: error something",
				},
				{
					RowNumber: 4,
					Error:     "unable to create new product group item: error something",
				},
				{
					RowNumber: 5,
					Error:     "unable to create new product group item: error something",
				},
			}},
			Req: &pb.ImportProductGroupRequest{
				Payload: []byte(`product_group_id,group_name,group_tag,discount_type,is_archived
				1,group-name-1,group-tag-a,discount-type-1,0
				2,group-name-2,group-tag-a,discount-type-1,0
				3,group-name-3,group-tag-b,discount-type-1,0
				4,group-name-4,group-tag-b,discount-type-1,0`),
			},
			Setup: func(ctx context.Context) {
				mockTx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(mockTx, nil)
				mockProductGroupRepo.On("GetByID", ctx, mockTx, mock.Anything).Return(entities.ProductGroup{}, constant.ErrDefault)
				mockProductGroupRepo.On("Create", ctx, mockTx, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase + " with discount type and archived",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedResp: &pb.ImportProductGroupResponse{Errors: []*pb.ImportProductGroupResponse_ImportProductGroupError{
				{
					RowNumber: 2,
					Error:     "unable to create new product group item: error something",
				},
				{
					RowNumber: 3,
					Error:     "unable to create new product group item: error something",
				},
				{
					RowNumber: 4,
					Error:     "unable to create new product group item: error something",
				},
				{
					RowNumber: 5,
					Error:     "unable to create new product group item: error something",
				},
			}},
			Req: &pb.ImportProductGroupRequest{
				Payload: []byte(`product_group_id,group_name,group_tag,discount_type,is_archived
				1,group-name-1,group-tag-a,discount-type-1,1
				2,group-name-2,group-tag-a,discount-type-2,1
				3,group-name-3,group-tag-b,discount-type-3,1
				4,group-name-4,group-tag-b,discount-type-4,1`),
			},
			Setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Return(mockTx, nil)
				mockTx.On("Rollback", mock.Anything).Return(nil)
				mockProductGroupRepo.On("GetByID", ctx, mockTx, mock.Anything).Return(entities.ProductGroup{}, constant.ErrDefault)
				mockProductGroupRepo.On("Create", ctx, mockTx, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        constant.NoDataInCsvFile,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req:         &pb.ImportProductGroupRequest{},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - mismatched number of fields in header and content",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "record on line 2: wrong number of fields"),
			Req: &pb.ImportProductGroupRequest{
				Payload: []byte(`product_group_id,group_name,group_tag
				1,group-name-1
				2,group-name-2
				3,group-name-3`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - number of column != 5 - missing product_group_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - number of column should be 5"),
			Req: &pb.ImportProductGroupRequest{
				Payload: []byte(`group_name,group_tag,discount_type,is_archived
				group-name-1, group-tag-a,,
				group-name-2, group-tag-a,,
				group-name-3, group-tag-a,,`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - first column name (toLowerCase) != product_group_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - first column (toLowerCase) should be 'product_group_id'"),
			Req: &pb.ImportProductGroupRequest{
				Payload: []byte(`incorrect_product_group_id,group_name,group_tag,discount_type,is_archived
				,group-name-1,group-tag-a,,
				,group-name-2,group-tag-a,,
				,group-name-3,group-tag-b,,
				,group-name-4,group-tag-b,,`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - second column name (toLowerCase) != group_name",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - second column (toLowerCase) should be 'group_name'"),
			Req: &pb.ImportProductGroupRequest{
				Payload: []byte(`product_group_id,incorrect_group_name,group_tag,discount_type,is_archived
				,group-name-1,group-tag-a,,
				,group-name-2,group-tag-a,,
				,group-name-3,group-tag-b,,
				,group-name-4,group-tag-b,,`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - third column name (toLowerCase) != group_tag",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - third column (toLowerCase) should be 'group_tag'"),
			Req: &pb.ImportProductGroupRequest{
				Payload: []byte(`product_group_id,group_name,incorrect_group_tag,discount_type,is_archived
				,group-name-1,group-tag-a,,
				,group-name-2,group-tag-a,,
				,group-name-3,group-tag-b,,
				,group-name-4,group-tag-b,,`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - fourth column name (toLowerCase) != discount_type",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - fourth column (toLowerCase) should be 'discount_type'"),
			Req: &pb.ImportProductGroupRequest{
				Payload: []byte(`product_group_id,group_name,group_tag,incorrect_discount_type,is_archived
				,group-name-1,group-tag-a,,
				,group-name-2,group-tag-a,,
				,group-name-3,group-tag-b,,
				,group-name-4,group-tag-b,,`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - fifth column name (toLowerCase) != is_archived",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - fifth column (toLowerCase) should be 'is_archived'"),
			Req: &pb.ImportProductGroupRequest{
				Payload: []byte(`product_group_id,group_name,group_tag,discount_type,invalid_is_archived
				,group-name-1,group-tag-a,,
				,group-name-2,group-tag-a,,
				,group-name-3,group-tag-b,,
				,group-name-4,group-tag-b,,`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "parsing valid file (with error lines in response)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportProductGroupRequest{
				Payload: []byte(`product_group_id,group_name,group_tag,discount_type,is_archived
				2,,group-tag-b,discount-type-1,0
				3,group-name-4,group-tag-b,discount-type-1,`),
			},
			ExpectedResp: &pb.ImportProductGroupResponse{
				Errors: []*pb.ImportProductGroupResponse_ImportProductGroupError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("unable to parse product group item: %s", fmt.Errorf("missing mandatory data: group_name")),
					},
					{
						RowNumber: 3,
						Error:     fmt.Sprintf("unable to parse product group item: missing mandatory data: is_archived"),
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockTx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(mockTx, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			resp, err := s.ImportProductGroup(testCase.Ctx, testCase.Req.(*pb.ImportProductGroupRequest))
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.ExpectedResp.(*pb.ImportProductGroupResponse)
				for i := range resp.Errors {
					assert.Equal(t, expectedResp.Errors[i].RowNumber, resp.Errors[i].RowNumber)
					assert.Contains(t, resp.Errors[i].Error, expectedResp.Errors[i].Error)
				}
			}

			mock.AssertExpectationsForObjects(t, db, mockProductGroupRepo)
		})
	}
}
