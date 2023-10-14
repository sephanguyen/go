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

func TestImportDiscount(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	db := new(mockDb.Ext)
	tx := new(mockDb.Tx)
	discountRepo := new(mockRepositories.MockDiscountRepo)

	s := &ImportMasterDataService{
		DB:           db,
		DiscountRepo: discountRepo,
	}

	testcases := []utils.TestCase{
		{
			Name:        constant.NoDataInCsvFile,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req:         &pb.ImportDiscountRequest{},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - mismatched number of fields in header and content",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, ""),
			Req: &pb.ImportDiscountRequest{
				Payload: []byte(`discount_id,name,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,available_until,remarks
				,Discount 1,1,1,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - number of column != 13",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - number of column should be 13"),
			Req: &pb.ImportDiscountRequest{
				Payload: []byte(`
				discount_id,name,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,available_until,remarks,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				,Discount 1,1,1,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,,,`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - first column name (toLowerCase) != discount_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - first column (toLowerCase) should be 'discount_id'"),
			Req: &pb.ImportDiscountRequest{
				Payload: []byte(`Number,name,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,available_until,remarks,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				1,Discount 1,1,1,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - second column name (toLowerCase) != name",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - second column (toLowerCase) should be 'name'"),
			Req: &pb.ImportDiscountRequest{
				Payload: []byte(`discount_id,Naming,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,available_until,remarks,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				1,Discount 1,1,1,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - third column name (toLowerCase) != discount_type",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - third column (toLowerCase) should be 'discount_type'"),
			Req: &pb.ImportDiscountRequest{
				Payload: []byte(`discount_id,name,Discount Type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,available_until,remarks,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				1,Discount 1,1,1,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - fourth column name (toLowerCase) != discount_amount_type",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - fourth column (toLowerCase) should be 'discount_amount_type'"),
			Req: &pb.ImportDiscountRequest{
				Payload: []byte(`discount_id,name,discount_type,Discount Amount Type,discount_amount_value,recurring_valid_duration,available_from,available_until,remarks,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				1,Discount 1,1,1,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - fifth column name (toLowerCase) != discount_amount_value",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - fifth column (toLowerCase) should be 'discount_amount_value'"),
			Req: &pb.ImportDiscountRequest{
				Payload: []byte(`discount_id,name,discount_type,discount_amount_type,Discount Amount Value,recurring_valid_duration,available_from,available_until,remarks,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				1,Discount 1,1,1,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - sixth column name (toLowerCase) != recurring_valid_duration",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - sixth column (toLowerCase) should be 'recurring_valid_duration'"),
			Req: &pb.ImportDiscountRequest{
				Payload: []byte(`discount_id,name,discount_type,discount_amount_type,discount_amount_value,Recurring Valid Duration,available_from,available_until,remarks,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				1,Discount 1,1,1,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - seventh column name (toLowerCase) != available_from",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - seventh column (toLowerCase) should be 'available_from'"),
			Req: &pb.ImportDiscountRequest{
				Payload: []byte(`discount_id,name,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,Available From,available_until,remarks,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				1,Discount 1,1,1,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - eighth column name (toLowerCase) != available_until",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - eighth column (toLowerCase) should be 'available_until'"),
			Req: &pb.ImportDiscountRequest{
				Payload: []byte(`discount_id,name,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,Available Until,remarks,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				1,Discount 1,1,1,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - ninth column name (toLowerCase) != remarks",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - ninth column (toLowerCase) should be 'remarks'"),
			Req: &pb.ImportDiscountRequest{
				Payload: []byte(`discount_id,name,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,available_until,Description,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				1,Discount 1,1,1,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - tenth column name (toLowerCase) != is_archived",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - tenth column (toLowerCase) should be 'is_archived'"),
			Req: &pb.ImportDiscountRequest{
				Payload: []byte(`discount_id,name,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,available_until,remarks,Is Archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				1,Discount 1,1,1,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "parsing valid file (with error lines in response)",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportDiscountRequest{
				Payload: []byte(`discount_id,name,discount_type,discount_amount_type,discount_amount_value,recurring_valid_duration,available_from,available_until,remarks,is_archived,student_tag_id_validation,parent_tag_id_validation,discount_tag_id
				,Discount 1,1,1,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 1,0,,,
				,Discount 2,2,1,12.25,2,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 2,0,,,
				1,Discount 3,1,2,12000,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 3,0,,,
				2,Discount 4,2,2,12000,2,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 4,0,,,
				,Discount 5,1,2,12000,2,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 5,2,,,
				,Discount 6,2,1,Two hundreds,2,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 6,0,,,
				,Discount 7,1,2,12000,,2021-12 23,2022-10-07T00:00:00-07:00,Remarks 7,0,,,
				,Discount 8,2,1,12.25,2,2021-12-07T00:00:00-07:00,2022-10--07,Remarks 8,0,,,
				,Discount 9,3,2,12000,2,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 9,0,,,
				,Discount 10,2,3,12000,2,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 10,0,,,
				,Discount 11,1,2,12000,NaN,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 11,0,,,
				,Discount 12,,1,12.25,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 12,0,,,
				,Discount 13,1,2,12000,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 13,0,,,
				1,Discount 14,2,1,12.25,2,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks 14,0,,,
				,Discount 15,1,1,12.25,,2021-12-07,2022-10-07,Remarks 1,0,,,
				1,Discount 3,1,2,12000,,2021-12-08,2022-10-09,Remarks 3,0,,,`),
			},
			ExpectedResp: &pb.ImportDiscountResponse{
				Errors: []*pb.ImportDiscountResponse_ImportDiscountError{
					{
						RowNumber: 6,
						Error:     fmt.Sprintf(constant.UnableToParseDiscountItem, fmt.Errorf("error parsing is_archived")),
					},
					{
						RowNumber: 7,
						Error:     fmt.Sprintf(constant.UnableToParseDiscountItem, fmt.Errorf("error parsing discount_amount_value")),
					},
					{
						RowNumber: 8,
						Error:     fmt.Sprintf(constant.UnableToParseDiscountItem, fmt.Errorf("error parsing available_from: parsing time \"2021-12 23\" as \"2006-01-02\": cannot parse \" 23\" as \"-\"")),
					},
					{
						RowNumber: 9,
						Error:     fmt.Sprintf(constant.UnableToParseDiscountItem, fmt.Errorf("error parsing available_until: parsing time \"2022-10--07\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"-07\" as \"02\"")),
					},
					{
						RowNumber: 10,
						Error:     fmt.Sprintf(constant.UnableToParseDiscountItem, fmt.Errorf("invalid discount_type: 3")),
					},
					{
						RowNumber: 11,
						Error:     fmt.Sprintf(constant.UnableToParseDiscountItem, fmt.Errorf("invalid discount_amount_type: 3")),
					},
					{
						RowNumber: 12,
						Error:     fmt.Sprintf(constant.UnableToParseDiscountItem, fmt.Errorf("error parsing recurring_valid_duration")),
					},
					{
						RowNumber: 13,
						Error:     fmt.Sprintf(constant.UnableToParseDiscountItem, fmt.Errorf("missing mandatory data: discount_type")),
					},
					{
						RowNumber: 14,
						Error:     fmt.Sprintf("unable to create new discount item: %s", pgx.ErrTxClosed),
					},
					{
						RowNumber: 15,
						Error:     fmt.Sprintf("unable to update discount item: %s", pgx.ErrTxClosed),
					},
				},
			},
			Setup: func(ctx context.Context) {
				discountRepo.On("Create", ctx, tx, mock.Anything).Twice().Return(nil)
				discountRepo.On("Update", ctx, tx, mock.Anything).Twice().Return(nil)
				discountRepo.On("Create", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				discountRepo.On("Update", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				discountRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				discountRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			resp, err := s.ImportDiscount(testCase.Ctx, testCase.Req.(*pb.ImportDiscountRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.ExpectedResp.(*pb.ImportDiscountResponse)
				for i, err := range resp.Errors {
					assert.Equal(t, err.RowNumber, expectedResp.Errors[i].RowNumber)
					assert.Contains(t, err.Error, expectedResp.Errors[i].Error)
				}
			}

			mock.AssertExpectationsForObjects(t, db, discountRepo)
		})
	}
}
