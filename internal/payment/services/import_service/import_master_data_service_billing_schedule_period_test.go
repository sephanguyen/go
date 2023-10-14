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

func TestImportBillingSchedulePeriod(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	tx := new(mockDb.Tx)
	db := new(mockDb.Ext)
	billingSchedulePeriodRepo := new(mockRepositories.MockBillingSchedulePeriodRepo)

	s := &ImportMasterDataService{
		DB:                        db,
		BillingSchedulePeriodRepo: billingSchedulePeriodRepo,
	}

	testcases := []utils.TestCase{
		{
			Name:        constant.NoDataInCsvFile,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req:         &pb.ImportBillingSchedulePeriodRequest{},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - mismatched number of fields in header and content",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, ""),
			Req: &pb.ImportBillingSchedulePeriodRequest{
				Payload: []byte(`billing_schedule_period_id,name,billing_schedule_id,start_date,end_date,billing_date,remarks
				,1,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 1
				,2,Cat 2,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 2
				,3,Cat 3,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 3`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - number of column != 8 - missing is_archived",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - number of column should be 8"),
			Req: &pb.ImportBillingSchedulePeriodRequest{
				Payload: []byte(`billing_schedule_period_id,name,billing_schedule_id,start_date,end_date,billing_date,remarks
				1,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 1
				2,Cat 2,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 2
				3,Cat 3,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 3`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - first column name (toLowerCase) != billing_schedule_period_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - first column (toLowerCase) should be 'billing_schedule_period_id'"),
			Req: &pb.ImportBillingSchedulePeriodRequest{
				Payload: []byte(`Number,name,billing_schedule_id,start_date,end_date,billing_date,remarks,is_archived
				1,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 1,0
				2,Cat 2,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 2,0
				3,Cat 3,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 3,0`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - second column name (toLowerCase) != name",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - second column (toLowerCase) should be 'name'"),
			Req: &pb.ImportBillingSchedulePeriodRequest{
				Payload: []byte(`billing_schedule_period_id,Naming,billing_schedule_id,start_date,end_date,billing_date,remarks,is_archived
				1,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 1,0
				2,Cat 2,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 2,0
				3,Cat 3,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 3,0`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - third column name (toLowerCase) != billing_schedule_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - third column (toLowerCase) should be 'billing_schedule_id'"),
			Req: &pb.ImportBillingSchedulePeriodRequest{
				Payload: []byte(`billing_schedule_period_id,name,BillingID,start_date,end_date,billing_date,remarks,is_archived
				1,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 1,0
				2,Cat 2,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 2,0
				3,Cat 3,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 3,0`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - fourth column name (toLowerCase) != start_date",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - fourth column (toLowerCase) should be 'start_date'"),
			Req: &pb.ImportBillingSchedulePeriodRequest{
				Payload: []byte(`billing_schedule_period_id,name,billing_schedule_id,StartDate,end_date,billing_date,remarks,is_archived
				1,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 1,0
				2,Cat 2,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 2,0
				3,Cat 3,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 3,0`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - fifth column name (toLowerCase) != end_date",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - fifth column (toLowerCase) should be 'end_date'"),
			Req: &pb.ImportBillingSchedulePeriodRequest{
				Payload: []byte(`billing_schedule_period_id,name,billing_schedule_id,start_date,EndDate,billing_date,remarks,is_archived
				1,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 1,0
				2,Cat 2,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 2,0
				3,Cat 3,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 3,0`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - sixth column name (toLowerCase) != billing_date",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - sixth column (toLowerCase) should be 'billing_date'"),
			Req: &pb.ImportBillingSchedulePeriodRequest{
				Payload: []byte(`billing_schedule_period_id,name,billing_schedule_id,start_date,end_date,BillingDate,remarks,is_archived
				1,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 1,0
				2,Cat 2,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 2,0
				3,Cat 3,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 3,0`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - seventh column name (toLowerCase) != remarks",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - seventh column (toLowerCase) should be 'remarks'"),
			Req: &pb.ImportBillingSchedulePeriodRequest{
				Payload: []byte(`billing_schedule_period_id,name,billing_schedule_id,start_date,end_date,billing_date,Description,is_archived
				1,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 1,0
				2,Cat 2,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 2,0
				3,Cat 3,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 3,0`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - eighth column name (toLowerCase) != is_archived",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - eighth column (toLowerCase) should be 'is_archived'"),
			Req: &pb.ImportBillingSchedulePeriodRequest{
				Payload: []byte(`billing_schedule_period_id,name,billing_schedule_id,start_date,end_date,billing_date,remarks,IsArchived
				1,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 1,0
				2,Cat 2,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 2,0
				3,Cat 3,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 3,0`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "parsing valid file (with error lines in response)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportBillingSchedulePeriodRequest{
				Payload: []byte(`billing_schedule_period_id,name,billing_schedule_id,start_date,end_date,billing_date,remarks,is_archived
				,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 1,0
				,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 2,0
				,Cat 2,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 2,Archived
				3,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,,Remarks 2,0
				,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 3,0
				4,Cat 1,1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,2021-12-09T00:00:00-07:00,Remarks 5,1
				,Cat 1,1,2021-12-07,2021-12-08,2021-12-09,Remarks 1,0
				4,Cat 4,1,2021-12-07,2021-12-08,2021-12-09,Remarks 5,1`),
			},
			ExpectedResp: &pb.ImportBillingSchedulePeriodResponse{
				Errors: []*pb.ImportBillingSchedulePeriodResponse_ImportBillingSchedulePeriodError{
					{
						RowNumber: 4,
						Error:     fmt.Sprintf(constant.UnableToParseBillingSchedulePeriodItem, fmt.Errorf("error parsing is_archived")),
					},
					{
						RowNumber: 5,
						Error:     fmt.Sprintf(constant.UnableToParseBillingSchedulePeriodItem, fmt.Errorf("missing mandatory data: billing_date")),
					},
					{
						RowNumber: 6,
						Error:     fmt.Sprintf("unable to create new billing schedule period item: %s", pgx.ErrTxClosed),
					},
					{
						RowNumber: 7,
						Error:     fmt.Sprintf("unable to update billing schedule period item: %s", pgx.ErrTxClosed),
					},
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("Create", ctx, tx, mock.Anything).Twice().Return(nil)
				billingSchedulePeriodRepo.On("Create", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				billingSchedulePeriodRepo.On("Update", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				billingSchedulePeriodRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				billingSchedulePeriodRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			resp, err := s.ImportBillingSchedulePeriod(testCase.Ctx, testCase.Req.(*pb.ImportBillingSchedulePeriodRequest))
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.ExpectedResp.(*pb.ImportBillingSchedulePeriodResponse)
				for i, err := range resp.Errors {
					assert.Equal(t, expectedResp.Errors[i].RowNumber, err.RowNumber)
					assert.Contains(t, err.Error, expectedResp.Errors[i].Error)
				}
			}

			mock.AssertExpectationsForObjects(t, db, billingSchedulePeriodRepo)
		})
	}
}
