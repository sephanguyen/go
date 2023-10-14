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

func TestImportLeavingReason(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	db := new(mockDb.Ext)
	tx := new(mockDb.Tx)
	leavingReasonRepo := new(mockRepositories.MockLeavingReasonRepo)

	s := &ImportMasterDataService{
		DB:                db,
		LeavingReasonRepo: leavingReasonRepo,
	}

	testcases := []utils.TestCase{
		{
			Name:        constant.NoDataInCsvFile,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req:         &pb.ImportLeavingReasonRequest{},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - mismatched number of fields in header and content",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, ""),
			Req: &pb.ImportLeavingReasonRequest{
				Payload: []byte(`leaving_reason_id,name,leaving_reason_type,remark
				,1,Cat 1,1,Remarks 1
				,2,Cat 2,2,Remarks 2
				,3,Cat 3,1,Remarks 3`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - number of column != 5",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - number of column should be 5"),
			Req: &pb.ImportLeavingReasonRequest{
				Payload: []byte(`leaving_reason_id,name,remark,
				1,Cat 1,1,Remarks 1
				2,Cat 2,2,Remarks 2
				3,Cat 3,1,Remarks 3`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - first column name (toLowerCase) != leaving_reason_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - first column (toLowerCase) should be 'leaving_reason_id'"),
			Req: &pb.ImportLeavingReasonRequest{
				Payload: []byte(`Number,leaving_reason_type,name,remark,is_archived
				1,Cat 1,1,Remarks 1,0
				2,Cat 2,2,Remarks 2,0
				3,Cat 3,1,Remarks 3,0`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - second column name (toLowerCase) != name",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - second column (toLowerCase) should be 'name'"),
			Req: &pb.ImportLeavingReasonRequest{
				Payload: []byte(`leaving_reason_id,Naming,leaving_reason_type,remark,is_archived
				1,Cat 1,1,Remarks 1,0
				2,Cat 2,1,Remarks 2,0
				3,Cat 3,2,Remarks 3,0`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "parsing valid file (with error lines in response)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportLeavingReasonRequest{
				Payload: []byte(`leaving_reason_id,name,leaving_reason_type,remark,is_archived
				,Cat 1,1,Remarks 1,1
				,Cat 1,1,Remarks 1,2
				,Cat 2,2,Remarks 2,
				3,Cat 3,1,,0
				,Cat 4,2,Remarks 4,1
				4,Cat 5,1,Remarks 5,0`),
			},
			ExpectedResp: &pb.ImportLeavingReasonResponse{
				Errors: []*pb.ImportLeavingReasonResponse_ImportLeavingReasonError{
					{
						RowNumber: 3,
						Error:     fmt.Sprintf(constant.UnableToParseLeavingReasonItem, fmt.Errorf("error parsing is_archived: strconv.ParseBool: parsing \"2\": invalid syntax")),
					},
					{
						RowNumber: 4,
						Error:     fmt.Sprintf(constant.UnableToParseLeavingReasonItem, fmt.Errorf("missing mandatory data: is_archived")),
					},
					{
						RowNumber: 6,
						Error:     fmt.Sprintf("unable to create leaving reason item: %s", pgx.ErrTxClosed),
					},
					{
						RowNumber: 7,
						Error:     fmt.Sprintf("unable to update leaving reason item: %s", pgx.ErrTxClosed),
					},
				},
			},
			Setup: func(ctx context.Context) {
				leavingReasonRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				leavingReasonRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				leavingReasonRepo.On("Create", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				leavingReasonRepo.On("Update", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				tx.On("Commit", mock.Anything).Return(nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)

			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			resp, err := s.ImportLeavingReason(testCase.Ctx, testCase.Req.(*pb.ImportLeavingReasonRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.ExpectedResp.(*pb.ImportLeavingReasonResponse)
				for i, err := range resp.Errors {
					assert.Equal(t, expectedResp.Errors[i].RowNumber, err.RowNumber)
					assert.Contains(t, err.Error, expectedResp.Errors[i].Error)
				}
			}

			mock.AssertExpectationsForObjects(t, db, leavingReasonRepo)
		})
	}
}
