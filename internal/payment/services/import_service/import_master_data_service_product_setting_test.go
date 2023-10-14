package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
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

func TestImportProductSetting(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	db := new(mockDb.Ext)
	mockTx := new(mockDb.Tx)
	mockProductSettingRepo := new(mockRepositories.MockProductSettingRepo)

	s := &ImportMasterDataService{
		DB:                 db,
		ProductSettingRepo: mockProductSettingRepo,
	}

	testcases := []utils.TestCase{
		{
			Name:         constant.HappyCase,
			Ctx:          interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedResp: &pb.ImportProductSettingResponse{Errors: []*pb.ImportProductSettingResponse_ImportProductSettingError{}},
			Req: &pb.ImportProductSettingRequest{
				Payload: []byte(`product_id,is_enrollment_required,is_pausable,is_added_to_enrollment_by_default,is_operation_fee
				1,true,true,true,true
				2,false,false,false,true
				3,true,true,true,true
				4,false,false,false,true`),
			},
			Setup: func(ctx context.Context) {
				mockTx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(mockTx, nil)
				mockProductSettingRepo.On("GetByID", ctx, mockTx, "1").Once().Return(entities.ProductSetting{}, constant.ErrDefault)
				mockProductSettingRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockProductSettingRepo.On("GetByID", ctx, mockTx, "2").Once().Return(entities.ProductSetting{}, constant.ErrDefault)
				mockProductSettingRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockProductSettingRepo.On("GetByID", ctx, mockTx, "3").Once().Return(entities.ProductSetting{}, constant.ErrDefault)
				mockProductSettingRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockProductSettingRepo.On("GetByID", ctx, mockTx, "4").Once().Return(entities.ProductSetting{}, constant.ErrDefault)
				mockProductSettingRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
			},
		},
		{
			Name:        constant.NoDataInCsvFile,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req:         &pb.ImportProductSettingRequest{},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - mismatched number of fields in header and content",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "record on line 2: wrong number of fields"),
			Req: &pb.ImportProductSettingRequest{
				Payload: []byte(`product_id,is_enrollment_required
				1,true,1
				2,false,2
				3,true,3`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - number of column != 3 - missing is_enrollment_required",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - number of column should be 5"),
			Req: &pb.ImportProductSettingRequest{
				Payload: []byte(`product_id,is_pausable
				1,true
				2,false
				3,true`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - first column name (toLowerCase) != product_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - first column (toLowerCase) should be 'product_id'"),
			Req: &pb.ImportProductSettingRequest{
				Payload: []byte(`wrong_product_id,is_enrollment_required,is_pausable,is_added_to_enrollment_by_default,is_operation_fee
				1,true,true,true,false
				2,false,false,false,false
				3,true,true,true,false`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - second column name (toLowerCase) != is_enrollment_required",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - second column (toLowerCase) should be 'is_enrollment_required'"),
			Req: &pb.ImportProductSettingRequest{
				Payload: []byte(`product_id,wrong_is_enrollment_required,is_pausable,is_added_to_enrollment_by_default,is_operation_fee
				1,true,true,true,false
				2,false,true,false,false
				3,true,false,true,false`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - third column name (toLowerCase) != is_pausable",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - third column (toLowerCase) should be 'is_pausable'"),
			Req: &pb.ImportProductSettingRequest{
				Payload: []byte(`product_id,is_enrollment_required,wrong_is_pausable,is_added_to_enrollment_by_default,is_operation_fee
				1,true,true,true,false
				2,false,true,false,false
				3,true,false,true,false`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - fourth column name (toLowerCase) != is_added_to_enrollment_by_default",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - fourth column (toLowerCase) should be 'is_added_to_enrollment_by_default'"),
			Req: &pb.ImportProductSettingRequest{
				Payload: []byte(`product_id,is_enrollment_required,is_pausable,wrong_is_added_to_enrollment_by_default,is_operation_fee
				1,true,true,true,false
				2,false,true,false,false
				3,true,false,true,false`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - fifth column name (toLowerCase) != is_operation_fee",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - fifth column (toLowerCase) should be 'is_operation_fee'"),
			Req: &pb.ImportProductSettingRequest{
				Payload: []byte(`product_id,is_enrollment_required,is_pausable,is_added_to_enrollment_by_default,wrong_is_operation_fee
				1,true,true,true,false
				2,false,true,false,false
				3,true,false,true,false`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "parsing valid file (with error lines in response)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportProductSettingRequest{
				Payload: []byte(`product_id,is_enrollment_required,is_pausable,is_added_to_enrollment_by_default,is_operation_fee
				1,,true,true,false
				,true,true,true,false
				2,false,false,true,false
				3,true,true,true,false
				1,false,false,true,false`),
			},
			ExpectedResp: &pb.ImportProductSettingResponse{
				Errors: []*pb.ImportProductSettingResponse_ImportProductSettingError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("unable to parse product setting item: %s", fmt.Errorf("missing mandatory data: is_enrollment_required")),
					},
					{
						RowNumber: 3,
						Error:     fmt.Sprintf("unable to parse product setting item: %s", fmt.Errorf("missing mandatory data: product_id")),
					},
					{
						RowNumber: 5,
						Error:     fmt.Sprintf("unable to create new product setting item: %s", fmt.Errorf("tx is closed")),
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockTx.On("Rollback", mock.Anything).Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(mockTx, nil)
				mockProductSettingRepo.On("GetByID", ctx, mockTx, "2").Once().Return(entities.ProductSetting{}, constant.ErrDefault)
				mockProductSettingRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockProductSettingRepo.On("GetByID", ctx, mockTx, "3").Once().Return(entities.ProductSetting{}, constant.ErrDefault)
				mockProductSettingRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockProductSettingRepo.On("GetByID", ctx, mockTx, "1").Once().Return(entities.ProductSetting{}, constant.ErrDefault)
				mockProductSettingRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			resp, err := s.ImportProductSetting(testCase.Ctx, testCase.Req.(*pb.ImportProductSettingRequest))
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.ExpectedResp.(*pb.ImportProductSettingResponse)
				for i, err := range resp.Errors {
					assert.Equal(t, expectedResp.Errors[i].RowNumber, err.RowNumber)
					assert.Contains(t, err.Error, expectedResp.Errors[i].Error)
				}
			}

			mock.AssertExpectationsForObjects(t, db, mockProductSettingRepo)
		})
	}
}
