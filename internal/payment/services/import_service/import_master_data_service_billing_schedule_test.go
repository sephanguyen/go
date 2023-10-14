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

func TestImportBillingSchedule(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	db := new(mockDb.Ext)
	tx := new(mockDb.Tx)
	billingScheduleRepo := new(mockRepositories.MockBillingScheduleRepo)

	s := &ImportMasterDataService{
		DB:                  db,
		BillingScheduleRepo: billingScheduleRepo,
	}

	testcases := []utils.TestCase{
		{
			Name:        constant.NoDataInCsvFile,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req:         &pb.ImportBillingScheduleRequest{},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - mismatched number of fields in header and content",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, ""),
			Req: &pb.ImportBillingScheduleRequest{
				Payload: []byte(`billing_schedule_id,name,remarks
				,1,Cat 1,Remarks 1
				,2,Cat 2,Remarks 2
				,3,Cat 3,Remarks 3`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - number of column != 4",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - number of column should be 4"),
			Req: &pb.ImportBillingScheduleRequest{
				Payload: []byte(`billing_schedule_id,name,remarks
				1,Cat 1,Remarks 1
				2,Cat 2,Remarks 2
				3,Cat 3,Remarks 3`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - first column name (toLowerCase) != billing_schedule_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - first column (toLowerCase) should be 'billing_schedule_id'"),
			Req: &pb.ImportBillingScheduleRequest{
				Payload: []byte(`Number,name,remarks,is_archived
				1,Cat 1,Remarks 1,0
				2,Cat 2,Remarks 2,0
				3,Cat 3,Remarks 3,0`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - second column name (toLowerCase) != name",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - second column (toLowerCase) should be 'name'"),
			Req: &pb.ImportBillingScheduleRequest{
				Payload: []byte(`billing_schedule_id,Naming,remarks,is_archived
				1,Cat 1,Remarks 1,0
				2,Cat 2,Remarks 2,0
				3,Cat 3,Remarks 3,0`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - thrid column name (toLowerCase) != remarks",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - third column (toLowerCase) should be 'remarks'"),
			Req: &pb.ImportBillingScheduleRequest{
				Payload: []byte(`billing_schedule_id,name,Description,is_archived
				1,Cat 1,Remarks 1,0
				2,Cat 2,Remarks 2,0
				3,Cat 3,Remarks 3,0`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - fourth column name (toLowerCase) != is_archived",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - fourth column (toLowerCase) should be 'is_archived'"),
			Req: &pb.ImportBillingScheduleRequest{
				Payload: []byte(`billing_schedule_id,name,remarks,is_archiving
				1,Cat 1,Remarks 1,0
				2,Cat 2,Remarks 2,0
				3,Cat 3,Remarks 3,0`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "parsing valid file (with error lines in response)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportBillingScheduleRequest{
				Payload: []byte(`billing_schedule_id,name,remarks,is_archived
				,Cat 1,Remarks 1,1
				,Cat 1,Remarks 1,2
				,Cat 2,Remarks 2,
				3,Cat 3,,0
				,Cat 4,Remarks 4,1
				4,Cat 5,Remarks 5,0`),
			},
			ExpectedResp: &pb.ImportBillingScheduleResponse{
				Errors: []*pb.ImportBillingScheduleResponse_ImportBillingScheduleError{
					{
						RowNumber: 3,
						Error:     fmt.Sprintf(constant.UnableToParseBillingScheduleItem, fmt.Errorf("error parsing is_archived")),
					},
					{
						RowNumber: 4,
						Error:     fmt.Sprintf(constant.UnableToParseBillingScheduleItem, fmt.Errorf("missing mandatory data: is_archived")),
					},
					{
						RowNumber: 6,
						Error:     fmt.Sprintf("unable to create new billing schedule item: %s", pgx.ErrTxClosed),
					},
					{
						RowNumber: 7,
						Error:     fmt.Sprintf("unable to update billing schedule item: %s", pgx.ErrTxClosed),
					},
				},
			},
			Setup: func(ctx context.Context) {
				billingScheduleRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				billingScheduleRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				billingScheduleRepo.On("Create", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				billingScheduleRepo.On("Update", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			resp, err := s.ImportBillingSchedule(testCase.Ctx, testCase.Req.(*pb.ImportBillingScheduleRequest))
			if err != nil {
				// fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.ExpectedResp.(*pb.ImportBillingScheduleResponse)
				for i, err := range resp.Errors {
					assert.Equal(t, expectedResp.Errors[i].RowNumber, err.RowNumber)
					assert.Contains(t, err.Error, expectedResp.Errors[i].Error)
				}
			}

			mock.AssertExpectationsForObjects(t, db, billingScheduleRepo)
		})
	}
}
