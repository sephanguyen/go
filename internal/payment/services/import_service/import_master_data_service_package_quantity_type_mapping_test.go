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

func TestImportPackageQuantityTypeMapping(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	db := new(mockDb.Ext)
	mockTx := new(mockDb.Tx)
	mockPackageQuantityTypeMappingRepo := new(mockRepositories.MockPackageQuantityTypeMappingRepo)

	s := &ImportMasterDataService{
		DB:                             db,
		PackageQuantityTypeMappingRepo: mockPackageQuantityTypeMappingRepo,
	}

	testcases := []utils.TestCase{
		{
			Name:         "happy case",
			Ctx:          interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedResp: &pb.ImportPackageQuantityTypeMappingResponse{Errors: []*pb.ImportPackageQuantityTypeMappingResponse_ImportPackageQuantityTypeMappingError{}},
			Req: &pb.ImportPackageQuantityTypeMappingRequest{
				Payload: []byte(`package_type,quantity_type
				1,1
				2,2
				3,3
				4,1`),
			},
			Setup: func(ctx context.Context) {
				mockTx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(mockTx, nil)
				mockPackageQuantityTypeMappingRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPackageQuantityTypeMappingRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPackageQuantityTypeMappingRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPackageQuantityTypeMappingRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
			},
		},
		{
			Name:        constant.NoDataInCsvFile,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req:         &pb.ImportPackageQuantityTypeMappingRequest{},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - mismatched number of fields in header and content",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "record on line 2: wrong number of fields"),
			Req: &pb.ImportPackageQuantityTypeMappingRequest{
				Payload: []byte(`package_type,quantity_type
				1,1,1
				2,2,2
				3,3,3`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - number of column != 2 - missing quantity_type",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - number of column should be 2"),
			Req: &pb.ImportPackageQuantityTypeMappingRequest{
				Payload: []byte(`package_type
				1
				2
				3`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - first column name (toLowerCase) != package_type",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - first column (toLowerCase) should be 'package_type'"),
			Req: &pb.ImportPackageQuantityTypeMappingRequest{
				Payload: []byte(`wrong_package_type,quantity_type
				1,1
				2,2
				3,3`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - second column name (toLowerCase) != quantity_type",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - second column (toLowerCase) should be 'quantity_type'"),
			Req: &pb.ImportPackageQuantityTypeMappingRequest{
				Payload: []byte(`package_type,wrong_quantity_type
				1,1
				2,2
				3,3`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "parsing valid file (with error lines in response)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportPackageQuantityTypeMappingRequest{
				Payload: []byte(`package_type,quantity_type
				1,
				,1
				2,2
				3,3
				1,1`),
			},
			ExpectedResp: &pb.ImportPackageQuantityTypeMappingResponse{
				Errors: []*pb.ImportPackageQuantityTypeMappingResponse_ImportPackageQuantityTypeMappingError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("unable to parse package quantity type mapping item: %s", fmt.Errorf("missing mandatory data: quantity_type")),
					},
					{
						RowNumber: 3,
						Error:     fmt.Sprintf("unable to parse package quantity type mapping item: %s", fmt.Errorf("missing mandatory data: package_type")),
					},
					{
						RowNumber: 5,
						Error:     fmt.Sprintf("unable to import package quantity type mapping item: %s", fmt.Errorf("tx is closed")),
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockTx.On("Rollback", mock.Anything).Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(mockTx, nil)
				mockPackageQuantityTypeMappingRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPackageQuantityTypeMappingRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockPackageQuantityTypeMappingRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			resp, err := s.ImportPackageQuantityTypeMapping(testCase.Ctx, testCase.Req.(*pb.ImportPackageQuantityTypeMappingRequest))
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.ExpectedResp.(*pb.ImportPackageQuantityTypeMappingResponse)
				for i, err := range resp.Errors {
					assert.Equal(t, expectedResp.Errors[i].RowNumber, err.RowNumber)
					assert.Contains(t, err.Error, expectedResp.Errors[i].Error)
				}
			}

			mock.AssertExpectationsForObjects(t, db, mockPackageQuantityTypeMappingRepo)
		})
	}
}
