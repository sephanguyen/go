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

func TestImportBillingRatio(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	tx := new(mockDb.Tx)
	db := new(mockDb.Ext)
	mockBillingRatioRepo := new(mockRepositories.MockBillingRatioRepo)

	s := &ImportMasterDataService{
		DB:               db,
		BillingRatioRepo: mockBillingRatioRepo,
	}

	testcases := []utils.TestCase{
		{
			Name:        constant.NoDataInCsvFile,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req:         &pb.ImportBillingRatioRequest{},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "only headers in csv file",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req: &pb.ImportBillingRatioRequest{
				Payload: []byte(`billing_ratio_id,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,is_archived`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - number of columns != 7",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - number of columns should be 7"),
			Req: &pb.ImportBillingRatioRequest{
				Payload: []byte(`billing_ratio_id,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator
				1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - first column name (toLowerCase) != billing_ratio_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - first column (toLowerCase) should be 'billing_ratio_id'"),
			Req: &pb.ImportBillingRatioRequest{
				Payload: []byte(`wrong_header,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,is_archived
				1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,0`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - second column name (toLowerCase) != start_date",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - second column (toLowerCase) should be 'start_date'"),
			Req: &pb.ImportBillingRatioRequest{
				Payload: []byte(`billing_ratio_id,wrong_header,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,is_archived
				1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,0`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - third column name (toLowerCase) != end_date",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - third column (toLowerCase) should be 'end_date'"),
			Req: &pb.ImportBillingRatioRequest{
				Payload: []byte(`billing_ratio_id,start_date,wrong_header,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,is_archived
				1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,0`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - fourth column name (toLowerCase) != billing_schedule_period_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - fourth column (toLowerCase) should be 'billing_schedule_period_id'"),
			Req: &pb.ImportBillingRatioRequest{
				Payload: []byte(`billing_ratio_id,start_date,end_date,wrong_header,billing_ratio_numerator,billing_ratio_denominator,is_archived
				1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,0`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - fifth column name (toLowerCase) != billing_ratio_numerator",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - fifth column (toLowerCase) should be 'billing_ratio_numerator'"),
			Req: &pb.ImportBillingRatioRequest{
				Payload: []byte(`billing_ratio_id,start_date,end_date,billing_schedule_period_id,wrong_header,billing_ratio_denominator,is_archived
				1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,0`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - sixth column name (toLowerCase) != billing_ratio_denominator",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - sixth column (toLowerCase) should be 'billing_ratio_denominator'"),
			Req: &pb.ImportBillingRatioRequest{
				Payload: []byte(`billing_ratio_id,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,wrong_header,is_archived
				1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,0`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - seventh column name (toLowerCase) != is_archived",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - seventh column (toLowerCase) should be 'is_archived'"),
			Req: &pb.ImportBillingRatioRequest{
				Payload: []byte(`billing_ratio_id,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,wrong_header
				1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,0`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "parsing valid file (capitalized header still valid) with no error lines in response",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportBillingRatioRequest{
				Payload: []byte(`Billing_ratio_id,Start_date,End_date,Billing_schedule_period_id,Billing_ratio_numerator,Billing_ratio_denominator,Is_archived
				,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,0`),
			},
			ExpectedResp: &pb.ImportBillingRatioResponse{
				Errors: []*pb.ImportBillingRatioResponse_ImportBillingRatioError{},
			},
			Setup: func(ctx context.Context) {
				mockBillingRatioRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				tx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)

			},
		},
		{
			Name: "update tax fields with no error lines in response",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportBillingRatioRequest{
				Payload: []byte(`billing_ratio_id,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,is_archived
				1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,0`),
			},
			ExpectedResp: &pb.ImportBillingRatioResponse{
				Errors: []*pb.ImportBillingRatioResponse_ImportBillingRatioError{},
			},
			Setup: func(ctx context.Context) {
				mockBillingRatioRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				tx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "missing mandatory data (except ID), error lines in response",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportBillingRatioRequest{
				Payload: []byte(`billing_ratio_id,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,is_archived
				1,,2021-12-08T00:00:00-07:00,3,1,2,0
				1,2021-12-07T00:00:00-07:00,,3,1,2,0
				1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,,1,2,0
				1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,,2,0
				1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,,0
				1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,`),
			},
			ExpectedResp: &pb.ImportBillingRatioResponse{
				Errors: []*pb.ImportBillingRatioResponse_ImportBillingRatioError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf(constant.UnableToParseBillingRatioItem, fmt.Errorf("missing mandatory data: start_date")),
					},
					{
						RowNumber: 3,
						Error:     fmt.Sprintf(constant.UnableToParseBillingRatioItem, fmt.Errorf("missing mandatory data: end_date")),
					},
					{
						RowNumber: 4,
						Error:     fmt.Sprintf(constant.UnableToParseBillingRatioItem, fmt.Errorf("missing mandatory data: billing_schedule_period_id")),
					},
					{
						RowNumber: 5,
						Error:     fmt.Sprintf(constant.UnableToParseBillingRatioItem, fmt.Errorf("missing mandatory data: billing_ratio_numerator")),
					},
					{
						RowNumber: 6,
						Error:     fmt.Sprintf(constant.UnableToParseBillingRatioItem, fmt.Errorf("missing mandatory data: billing_ratio_denominator")),
					},
					{
						RowNumber: 7,
						Error:     fmt.Sprintf(constant.UnableToParseBillingRatioItem, fmt.Errorf("missing mandatory data: is_archived")),
					},
				},
			},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "wrong number of data in a record",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportBillingRatioRequest{
				Payload: []byte(`billing_ratio_id,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,is_archived
				1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,0
				3,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2`),
			},
			ExpectedErr: status.Error(codes.InvalidArgument, "record on line 3: wrong number of fields"),
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "wrong values, error lines in response",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportBillingRatioRequest{
				Payload: []byte(`billing_ratio_id,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,is_archived
				1,c,2021-12-08T00:00:00-07:00,3,1,2,0
				1,2021-12-07T00:00:00-07:00,d,3,1,2,0
				1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,f,2,0
				1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,g,0
				1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,h`),
			},
			ExpectedResp: &pb.ImportBillingRatioResponse{
				Errors: []*pb.ImportBillingRatioResponse_ImportBillingRatioError{
					// {
					// 	RowNumber: 2,
					// 	Error:     fmt.Sprintf(constant.UnableToParseBillingRatioItem, fmt.Errorf("error parsing billing_ratio_id: strconv.Atoi: parsing \"a\": invalid syntax")),
					// },
					{
						RowNumber: 2,
						Error:     fmt.Sprintf(constant.UnableToParseBillingRatioItem, fmt.Errorf("error parsing start_date: parsing time \"c\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"c\" as \"2006\"")),
					},
					{
						RowNumber: 3,
						Error:     fmt.Sprintf(constant.UnableToParseBillingRatioItem, fmt.Errorf("error parsing end_date: parsing time \"d\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"d\" as \"2006\"")),
					},
					// {
					// 	RowNumber: 5,
					// 	Error:     fmt.Sprintf(constant.UnableToParseBillingRatioItem, fmt.Errorf("error parsing billing_schedule_period_id: strconv.Atoi: parsing \"e\": invalid syntax")),
					// },
					{
						RowNumber: 4,
						Error:     fmt.Sprintf(constant.UnableToParseBillingRatioItem, fmt.Errorf("error parsing billing_ratio_numerator: strconv.Atoi: parsing \"f\": invalid syntax")),
					},
					{
						RowNumber: 5,
						Error:     fmt.Sprintf(constant.UnableToParseBillingRatioItem, fmt.Errorf("error parsing billing_ratio_denominator: strconv.Atoi: parsing \"g\": invalid syntax")),
					},
					{
						RowNumber: 6,
						Error:     fmt.Sprintf(constant.UnableToParseBillingRatioItem, fmt.Errorf("error parsing is_archived: strconv.ParseBool: parsing \"h\": invalid syntax")),
					},
				},
			},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "wrong constraints, error lines in response",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportBillingRatioRequest{
				Payload: []byte(`billing_ratio_id,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,is_archived
				1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,-1,2,0
				1,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,0,0
				1,2021-12-08T00:00:00-07:00,2021-12-07T00:00:00-07:00,3,1,2,0`),
			},
			ExpectedResp: &pb.ImportBillingRatioResponse{
				Errors: []*pb.ImportBillingRatioResponse_ImportBillingRatioError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf(constant.UnableToParseBillingRatioItem, fmt.Errorf("billing_ratio_numerator should be >= 0")),
					},
					{
						RowNumber: 3,
						Error:     fmt.Sprintf(constant.UnableToParseBillingRatioItem, fmt.Errorf("billing_ratio_denominator should be >= 1")),
					},
					{
						RowNumber: 4,
						Error:     fmt.Sprintf(constant.UnableToParseBillingRatioItem, fmt.Errorf("start_date should be before end_date")),
					},
				},
			},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "create/update billing ratio with error lines in response",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportBillingRatioRequest{
				Payload: []byte(`billing_ratio_id,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,is_archived
				,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,0
				,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,1
				3,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,0
				4,2021-12-07T00:00:00-07:00,2021-12-08T00:00:00-07:00,3,1,2,1
				,2021-12-08,2021-12-09,3,1,2,1
				5,2021-12-07,2021-12-08,3,1,2,1`),
			},
			ExpectedResp: &pb.ImportBillingRatioResponse{
				Errors: []*pb.ImportBillingRatioResponse_ImportBillingRatioError{
					{
						RowNumber: 3,
						Error:     fmt.Sprintf("unable to create new billing ratio item: %s", pgx.ErrTxClosed),
					},
					{
						RowNumber: 5,
						Error:     fmt.Sprintf("unable to update billing ratio item: %s", pgx.ErrTxClosed),
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockBillingRatioRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				mockBillingRatioRepo.On("Create", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockBillingRatioRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				mockBillingRatioRepo.On("Update", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockBillingRatioRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				mockBillingRatioRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			resp, err := s.ImportBillingRatio(testCase.Ctx, testCase.Req.(*pb.ImportBillingRatioRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)

				for i, expectedErr := range testCase.ExpectedResp.(*pb.ImportBillingRatioResponse).Errors {
					assert.Equal(t, expectedErr.RowNumber, resp.Errors[i].RowNumber)
					assert.Contains(t, expectedErr.Error, resp.Errors[i].Error)
				}
			}

			mock.AssertExpectationsForObjects(t, db, mockBillingRatioRepo)
		})
	}
}
