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

func TestImportProductLocation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	db := new(mockDb.Ext)
	tx := new(mockDb.Tx)
	productLocationRepo := new(mockRepositories.MockProductLocationRepo)

	s := &ImportMasterDataService{
		DB:                  db,
		ProductLocationRepo: productLocationRepo,
	}

	testcases := []utils.TestCase{
		{
			Name:        constant.NoDataInCsvFile,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req: &pb.ImportProductAssociatedDataRequest{
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_LOCATION,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - mismatched number of fields in header and content",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "record on line 2: wrong number of fields"),
			Req: &pb.ImportProductAssociatedDataRequest{
				Payload: []byte(
					`product_id,location_id
					1`,
				),
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_LOCATION,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - number of column != 2 - missing location_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "record on line 2: wrong number of fields"),
			Req: &pb.ImportProductAssociatedDataRequest{
				Payload: []byte(
					`product_id
					1,Location-1`,
				),
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_LOCATION,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "wrong name column",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - first column (toLowerCase) should be 'product_id'"),
			Req: &pb.ImportProductAssociatedDataRequest{
				Payload: []byte(
					`product_ids,location_id
					1,Location-1`,
				),
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_LOCATION,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "update location with missing mandatory column",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportProductAssociatedDataRequest{
				Payload: []byte(`product_id,location_id
				4,`),
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_LOCATION,
			},
			ExpectedResp: &pb.ImportProductAssociatedDataResponse{
				Errors: []*pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("unable to parse product location item: %s", fmt.Errorf("missing mandatory data: location_id")),
					},
				},
			},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		// {
		// 	Name:        "Fail when parse product location",
		// 	Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
		// 	ExpectedErr: nil,
		// 	Req: &pb.ImportProductAssociatedDataRequest{
		// 		Payload: []byte(
		// 			`product_id,location_id
		// 			1ass,Location-1`,
		// 		),
		// 		ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_LOCATION,
		// 	},
		// 	ExpectedResp: &pb.ImportProductAssociatedDataResponse{Errors: []*pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError{
		// 		{
		// 			RowNumber: int32(2),
		// 			Error:     `unable to parse product location item: error parsing product_id: strconv.Atoi: parsing "1ass": invalid syntax`,
		// 		},
		// 	}},
		// 	Setup: func(ctx context.Context) {
		// 		tx.On("Rollback", mock.Anything).Return(nil)
		// 		db.On("Begin", mock.Anything).Return(tx, nil)
		// 	},
		// },
		{
			Name:        "Fail when import product location",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &pb.ImportProductAssociatedDataRequest{
				Payload: []byte(
					`product_id,location_id
					1,Location-1
					1,Location-2`,
				),
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_LOCATION,
			},
			ExpectedResp: &pb.ImportProductAssociatedDataResponse{Errors: []*pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError{
				{
					RowNumber: int32(2),
					Error:     `unable to import product location item: error something`,
				},
				{
					RowNumber: int32(3),
					Error:     `unable to import product location item: error something`,
				},
			}},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				productLocationRepo.On("Replace", ctx, tx, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error something"))
				productLocationRepo.On("Replace", ctx, tx, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error something"))
			},
		},
		{
			Name:        "happy case import product location success",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &pb.ImportProductAssociatedDataRequest{
				Payload: []byte(
					`product_id,location_id
					1,Location-1
					1,Location-2`,
				),
				ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_LOCATION,
			},
			ExpectedResp: &pb.ImportProductAssociatedDataResponse{Errors: []*pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError{}},
			Setup: func(ctx context.Context) {
				tx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				productLocationRepo.On("Replace", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				productLocationRepo.On("Replace", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
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

			mock.AssertExpectationsForObjects(t, db, productLocationRepo)
		})
	}
}
