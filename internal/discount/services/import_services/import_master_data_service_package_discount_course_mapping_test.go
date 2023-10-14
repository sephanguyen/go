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

func TestImportPackageDiscountCourseMapping(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	db := new(mockDb.Ext)
	mockTx := new(mockDb.Tx)
	mockPackageDiscountCourseMappingRepo := new(mockRepositories.MockPackageDiscountCourseMappingRepo)

	s := &ImportMasterDataService{
		DB:                               db,
		PackageDiscountCourseMappingRepo: mockPackageDiscountCourseMappingRepo,
	}

	testcases := []utils.TestCase{
		{
			Name:         constant.HappyCase + "upsert single record",
			Ctx:          interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedResp: &pb.ImportPackageDiscountCourseMappingResponse{Errors: []*pb.ImportPackageDiscountCourseMappingResponse_ImportPackageDiscountCourseMappingError{}},
			Req: &pb.ImportPackageDiscountCourseMappingRequest{
				Payload: []byte(`package_id,course_combination_ids,discount_tag_id,is_archived,product_group_id
				package-id-1,course a;course b;course c,discount-tag-id-1,0,product_group_id`),
			},
			Setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				mockPackageDiscountCourseMappingRepo.On("Upsert", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			Name:         constant.HappyCase + "upsert new multiple record",
			Ctx:          interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedResp: &pb.ImportPackageDiscountCourseMappingResponse{Errors: []*pb.ImportPackageDiscountCourseMappingResponse_ImportPackageDiscountCourseMappingError{}},
			Req: &pb.ImportPackageDiscountCourseMappingRequest{
				Payload: []byte(`package_id,course_combination_ids,discount_tag_id,is_archived,product_group_id
				package-id-1,course a;course b;course c,discount-tag-id-1,0,product_group_id
				package-id-2,course c,discount-tag-id-2,0,product_group_id
				package-id-3,course b,discount-tag-id-3,1,product_group_id`),
			},
			Setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(mockTx, nil)
				mockPackageDiscountCourseMappingRepo.On("Upsert", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPackageDiscountCourseMappingRepo.On("Upsert", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockPackageDiscountCourseMappingRepo.On("Upsert", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			Name:        constant.NoDataInCsvFile + " for package discount course mapping",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req:         &pb.ImportPackageDiscountCourseMappingRequest{},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - mismatched number of fields in header and content",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "record on line 2: wrong number of fields"),
			Req: &pb.ImportPackageDiscountCourseMappingRequest{
				Payload: []byte(`package_id,course_combination_ids,discount_tag_id,is_archived,product_group_id
				1
				1,2
				2,3,2`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - number of column != 5 - missing package_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - number of column should be 5"),
			Req: &pb.ImportPackageDiscountCourseMappingRequest{
				Payload: []byte(`course_combination_ids,discount_tag_id,is_archived,product_group_id
				course A,discount-tag-id-1,0,product_group_id
				course B,discount-tag-id-1,0,product_group_id
				course C,discount-tag-id-1,1,product_group_id`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - first column name (toLowerCase) != package_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - first column (toLowerCase) should be 'package_id'"),
			Req: &pb.ImportPackageDiscountCourseMappingRequest{
				Payload: []byte(`incorrect_package_id,course_combination_ids,discount_tag_id,is_archived,product_group_id
				package-id-1,course a;course b;course c,discount-tag-id-1,0,product_group_id`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - second column name (toLowerCase) != course_combination_ids",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - second column (toLowerCase) should be 'course_combination_ids'"),
			Req: &pb.ImportPackageDiscountCourseMappingRequest{
				Payload: []byte(`package_id,incorrect_course_combination_ids,discount_tag_id,is_archived,product_group_id
				package-id-1,course a;course b;course c,discount-tag-id-1,0,product_group_id`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - third column name (toLowerCase) != discount_tag_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - third column (toLowerCase) should be 'discount_tag_id'"),
			Req: &pb.ImportPackageDiscountCourseMappingRequest{
				Payload: []byte(`package_id,course_combination_ids,incorrect_discount_tag_id,is_archived,product_group_id
				package-id-1,course a;course b;course c,discount-tag-id-1,0,product_group_id`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - fourth column name (toLowerCase) != is_archived",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - fourth column (toLowerCase) should be 'is_archived'"),
			Req: &pb.ImportPackageDiscountCourseMappingRequest{
				Payload: []byte(`package_id,course_combination_ids,discount_tag_id,incorrect_is_archived,product_group_id
				package-id-1,course a;course b;course c,discount-tag-id-1,0,product_group_id`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - fifth column name (toLowerCase) != is_archived",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - fifth column (toLowerCase) should be 'product_group_id'"),
			Req: &pb.ImportPackageDiscountCourseMappingRequest{
				Payload: []byte(`package_id,course_combination_ids,discount_tag_id,is_archived,invalid_product_group_id
				package-id-1,course a;course b;course c,discount-tag-id-1,0,product_group_id`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "parsing package discount setting valid file (with error lines in response)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportPackageDiscountCourseMappingRequest{
				Payload: []byte(`package_id,course_combination_ids,discount_tag_id,is_archived,product_group_id
				,course a;course b;course c,discount-tag-id-1,0,product_group_id
				package-id-1,,discount-tag-id-1,0,product_group_id
				package-id-1,course a;course b;course c,,0,product_group_id`),
			},
			ExpectedResp: &pb.ImportPackageDiscountCourseMappingResponse{
				Errors: []*pb.ImportPackageDiscountCourseMappingResponse_ImportPackageDiscountCourseMappingError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("unable to parse package discount course mapping: %s", fmt.Errorf("missing mandatory data: package_id")),
					},
					{
						RowNumber: 3,
						Error:     fmt.Sprintf("unable to parse package discount course mapping: %s", fmt.Errorf("missing mandatory data: course_combination_ids")),
					},
					{
						RowNumber: 4,
						Error:     fmt.Sprintf("unable to parse package discount course mapping: %s", fmt.Errorf("missing mandatory data: discount_tag_id")),
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
			Req: &pb.ImportPackageDiscountCourseMappingRequest{
				Payload: []byte(`package_id,course_combination_ids,discount_tag_id,is_archived,product_group_id
				package-id-1,course a;course b;course c,discount-tag-id-1,0,product_group_id`),
			},
			ExpectedResp: &pb.ImportPackageDiscountCourseMappingResponse{
				Errors: []*pb.ImportPackageDiscountCourseMappingResponse_ImportPackageDiscountCourseMappingError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("unable to import package discount course mapping: error something"),
					},
				},
			},
			Setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Return(mockTx, nil)
				mockPackageDiscountCourseMappingRepo.On("Upsert", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(constant.ErrDefault)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			resp, err := s.ImportPackageDiscountCourseMapping(testCase.Ctx, testCase.Req.(*pb.ImportPackageDiscountCourseMappingRequest))
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.ExpectedResp.(*pb.ImportPackageDiscountCourseMappingResponse)
				for i, err := range resp.Errors {
					assert.Equal(t, expectedResp.Errors[i].RowNumber, err.RowNumber)
					assert.Contains(t, err.Error, expectedResp.Errors[i].Error)
				}
			}

			mock.AssertExpectationsForObjects(t, db, mockPackageDiscountCourseMappingRepo, mockTx)
		})
	}
}
