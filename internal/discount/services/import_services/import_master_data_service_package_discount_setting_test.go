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

func TestImportPackageDiscountSetting(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	db := new(mockDb.Ext)
	mockTx := new(mockDb.Tx)
	mockPackageDiscountSettingRepo := new(mockRepositories.MockPackageDiscountSettingRepo)

	s := &ImportMasterDataService{
		DB:                         db,
		PackageDiscountSettingRepo: mockPackageDiscountSettingRepo,
	}

	testcases := []utils.TestCase{
		{
			Name:         constant.HappyCase + "upser single record",
			Ctx:          interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedResp: &pb.ImportPackageDiscountSettingResponse{Errors: []*pb.ImportPackageDiscountSettingResponse_ImportPackageDiscountSettingError{}},
			Req: &pb.ImportPackageDiscountSettingRequest{
				Payload: []byte(`package_id,min_slot_trigger,max_slot_trigger,discount_tag_id,is_archived,product_group_id
				package-id-1,1,5,discount-tag-id-1,0,product_group_id`),
			},
			Setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				mockPackageDiscountSettingRepo.On("Upsert", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			Name:         constant.HappyCase + "upsert new multiple record",
			Ctx:          interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedResp: &pb.ImportPackageDiscountSettingResponse{Errors: []*pb.ImportPackageDiscountSettingResponse_ImportPackageDiscountSettingError{}},
			Req: &pb.ImportPackageDiscountSettingRequest{
				Payload: []byte(`package_id,min_slot_trigger,max_slot_trigger,discount_tag_id,is_archived,product_group_id
				package-id-1,1,5,discount-tag-id-1,0,product_group_id
				package-id-2,2,2,discount-tag-id-2,0,product_group_id
				package-id-3,3,4,discount-tag-id-3,1,product_group_id`),
			},
			Setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				mockPackageDiscountSettingRepo.On("Upsert", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPackageDiscountSettingRepo.On("Upsert", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPackageDiscountSettingRepo.On("Upsert", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			Name:        constant.NoDataInCsvFile + " for package discount setting",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req:         &pb.ImportPackageDiscountSettingRequest{},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - mismatched number of fields in header and content",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "record on line 2: wrong number of fields"),
			Req: &pb.ImportPackageDiscountSettingRequest{
				Payload: []byte(`package_id,min_slot_trigger,max_slot_trigger,discount_tag_id,is_archived,product_group_id
				1,2
				2,3,2
				3,3,2,1`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - number of column != 6 - missing package_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - number of column should be 6"),
			Req: &pb.ImportPackageDiscountSettingRequest{
				Payload: []byte(`min_slot_trigger,max_slot_trigger,discount_tag_id,is_archived,product_group_id
				1,5,discount-tag-id-1,0,product_group_id
				2,2,discount-tag-id-2,0,product_group_id
				3,4,discount-tag-id-3,1,product_group_id`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - first column name (toLowerCase) != package_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - first column (toLowerCase) should be 'package_id'"),
			Req: &pb.ImportPackageDiscountSettingRequest{
				Payload: []byte(`incorrect_package_id,min_slot_trigger,max_slot_trigger,discount_tag_id,is_archived,product_group_id
				package-id-1-update,4,4,discount-tag-id-1-update,1,product_group_id`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - second column name (toLowerCase) != min_slot_trigger",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - second column (toLowerCase) should be 'min_slot_trigger'"),
			Req: &pb.ImportPackageDiscountSettingRequest{
				Payload: []byte(`package_id,incorrect_min_slot_trigger,max_slot_trigger,discount_tag_id,is_archived,product_group_id
				package-id-1-update,4,4,discount-tag-id-1-update,1,product_group_id`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - third column name (toLowerCase) != max_slot_trigger",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - third column (toLowerCase) should be 'max_slot_trigger'"),
			Req: &pb.ImportPackageDiscountSettingRequest{
				Payload: []byte(`package_id,min_slot_trigger,incorrect_max_slot_trigger,discount_tag_id,is_archived,product_group_id
				package-id-1-update,4,4,discount-tag-id-1-update,1,product_group_id`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - fourth column name (toLowerCase) != discount_tag_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - fourth column (toLowerCase) should be 'discount_tag_id'"),
			Req: &pb.ImportPackageDiscountSettingRequest{
				Payload: []byte(`package_id,min_slot_trigger,max_slot_trigger,incorrect_discount_tag_id,is_archived,product_group_id
				package-id-1-update,4,4,discount-tag-id-1-update,1,product_group_id`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - fifth column name (toLowerCase) != is_archived",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - fifth column (toLowerCase) should be 'is_archived'"),
			Req: &pb.ImportPackageDiscountSettingRequest{
				Payload: []byte(`package_id,min_slot_trigger,max_slot_trigger,discount_tag_id,incorrect_is_archived,product_group_id
				package-id-1-update,4,4,discount-tag-id-1-update,1,product_group_id`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "parsing package discount setting valid file (with error lines in response)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportPackageDiscountSettingRequest{
				Payload: []byte(`package_id,min_slot_trigger,max_slot_trigger,discount_tag_id,is_archived,product_group_id
				,1,5,discount-tag-id-1,0,product_group_id
				package-id-2,2,2,,0,product_group_id`),
			},
			ExpectedResp: &pb.ImportPackageDiscountSettingResponse{
				Errors: []*pb.ImportPackageDiscountSettingResponse_ImportPackageDiscountSettingError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("unable to parse package discount setting: %s", fmt.Errorf("missing mandatory data: package_id")),
					},
					{
						RowNumber: 3,
						Error:     fmt.Sprintf("unable to parse package discount setting: %s", fmt.Errorf("missing mandatory data: discount_tag_id")),
					},
				},
			},
			Setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			Name: " failed upsert single record",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportPackageDiscountSettingRequest{
				Payload: []byte(`package_id,min_slot_trigger,max_slot_trigger,discount_tag_id,is_archived,product_group_id
				package-id-1,1,5,discount-tag-id-1,0,product_group_id`),
			},
			ExpectedResp: &pb.ImportPackageDiscountSettingResponse{
				Errors: []*pb.ImportPackageDiscountSettingResponse_ImportPackageDiscountSettingError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("unable to import package discount setting: error something"),
					},
				},
			},
			Setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Return(mockTx, nil)
				mockPackageDiscountSettingRepo.On("Upsert", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(constant.ErrDefault)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			resp, err := s.ImportPackageDiscountSetting(testCase.Ctx, testCase.Req.(*pb.ImportPackageDiscountSettingRequest))
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.ExpectedResp.(*pb.ImportPackageDiscountSettingResponse)
				for i, err := range resp.Errors {
					assert.Equal(t, expectedResp.Errors[i].RowNumber, err.RowNumber)
					assert.Contains(t, err.Error, expectedResp.Errors[i].Error)
				}
			}

			mock.AssertExpectationsForObjects(t, db, mockPackageDiscountSettingRepo, mockTx)
		})
	}
}
