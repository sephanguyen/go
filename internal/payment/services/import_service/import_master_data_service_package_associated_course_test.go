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

func TestImportPackageCourse(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	db := new(mockDb.Ext)
	tx := new(mockDb.Tx)
	packageCourseRepo := new(mockRepositories.MockPackageCourseRepo)

	s := &ImportMasterDataService{
		DB:                db,
		PackageCourseRepo: packageCourseRepo,
	}

	testcases := []utils.TestCase{
		{
			Name:        "no data in csv file",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req: &pb.ImportProductAssociatedDataRequest{
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_COURSE,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "miss some column",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - number of column should be 5"),
			Req: &pb.ImportProductAssociatedDataRequest{
				Payload: []byte(`package_id,course_id,mandatory_flag
1,Course-1,0
`),
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_COURSE,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "wrong name column",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - first column (toLowerCase) should be 'package_id'"),
			Req: &pb.ImportProductAssociatedDataRequest{
				Payload: []byte(`package_idas,course_id,mandatory_flag,max_slots_per_course,course_weight
1,Course-1,0,2,2
`),
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_COURSE,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "mismatched number of fields in header and content",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "record on line 2: wrong number of fields"),
			Req: &pb.ImportProductAssociatedDataRequest{
				Payload: []byte(`package_id,course_id,max_slots_per_course,mandatory_flag
1,Course-1,0,2,3
`),
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_COURSE,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		// 		{
		// 			Name:        "Fail when parse product course",
		// 			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
		// 			ExpectedErr: nil,
		// 			Req: &pb.ImportProductAssociatedDataRequest{
		// 				Payload: []byte(`package_id,course_id,mandatory_flag,max_slots_per_course,course_weight
		// 1ass,Course-1,0,2,3
		// `),
		// 				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_COURSE,
		// 			},
		// 			ExpectedResp: &pb.ImportProductAssociatedDataResponse{Errors: []*pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError{
		// 				{
		// 					RowNumber: int32(2),
		// 					Error:     `unable to parse package course item: error parsing package_id: strconv.Atoi: parsing "1ass": invalid syntax`,
		// 				},
		// 			}},
		// 			Setup: func(ctx context.Context) {
		// 				tx.On("Rollback", mock.Anything).Return(nil)
		// 				db.On("Begin", mock.Anything).Return(tx, nil)
		// 			},
		// 		},
		{
			Name:        "Fail when import product course",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &pb.ImportProductAssociatedDataRequest{
				Payload: []byte(`package_id,course_id,mandatory_flag,max_slots_per_course,course_weight
1,Course-1,0,2,3
1,Course-2,0,2,3
`),
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_COURSE,
			},
			ExpectedResp: &pb.ImportProductAssociatedDataResponse{Errors: []*pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError{
				{
					RowNumber: int32(2),
					Error:     `unable to import package course item: error something`,
				},
				{
					RowNumber: int32(3),
					Error:     `unable to import package course item: error something`,
				},
			}},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				packageCourseRepo.On("Upsert", ctx, tx, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error something"))
			},
		},
		{
			Name:        "Happy case import product course success",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &pb.ImportProductAssociatedDataRequest{
				Payload: []byte(`package_id,course_id,mandatory_flag,max_slots_per_course,course_weight
1,Course-1,0,2,3
1,Course-2,0,2,3
`),
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_COURSE,
			},
			ExpectedResp: &pb.ImportProductAssociatedDataResponse{Errors: []*pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError{}},
			Setup: func(ctx context.Context) {
				tx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				packageCourseRepo.On("Upsert", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			resp, err := s.ImportProductAssociatedData(testCase.Ctx, testCase.Req.(*pb.ImportProductAssociatedDataRequest))
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.ExpectedResp.(*pb.ImportProductAssociatedDataResponse)
				for i, err := range resp.Errors {
					assert.Equal(t, expectedResp.Errors[i].RowNumber, err.RowNumber)
					assert.Equal(t, err.Error, expectedResp.Errors[i].Error)
				}
			}

			mock.AssertExpectationsForObjects(t, db, packageCourseRepo)
		})
	}
}
