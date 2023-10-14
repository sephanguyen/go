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

func TestImportTax(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	tx := new(mockDb.Tx)
	db := new(mockDb.Ext)
	mockTaxRepo := new(mockRepositories.MockTaxRepo)

	s := &ImportMasterDataService{
		DB:      db,
		TaxRepo: mockTaxRepo,
	}

	testcases := []utils.TestCase{
		{
			Name:        constant.NoDataInCsvFile,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req:         &pb.ImportTaxRequest{},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "only headers in csv file",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req: &pb.ImportTaxRequest{
				Payload: []byte(`tax_id,name,tax_percentage,tax_category,default_flag,is_archived`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - number of columns != 6",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - number of column should be 6"),
			Req: &pb.ImportTaxRequest{
				Payload: []byte(`tax_id,name,tax_percentage
				1,Tax 1,10
				2,Tax 2,20
				3,Tax 3,30`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - first column name (toLowerCase) != tax_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - first column (toLowerCase) should be 'tax_id'"),
			Req: &pb.ImportTaxRequest{
				Payload: []byte(`wrong_header,name,tax_percentage,tax_category,default_flag,is_archived
				1,Tax 1,10,1,0,0
				2,Tax 2,20,2,1,1`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - second column name (toLowerCase) != name",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - second column (toLowerCase) should be 'name'"),
			Req: &pb.ImportTaxRequest{
				Payload: []byte(`tax_id,wrong_header,tax_percentage,tax_category,default_flag,is_archived
				1,Tax 1,10,1,0,0
				2,Tax 2,20,2,1,1`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - third column name (toLowerCase) != tax_percentage",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - third column (toLowerCase) should be 'tax_percentage'"),
			Req: &pb.ImportTaxRequest{
				Payload: []byte(`tax_id,name,wrong_header,tax_category,default_flag,is_archived
				1,Tax 1,10,1,0,0
				2,Tax 2,20,2,1,1`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - fourth column name (toLowerCase) != tax_category",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - fourth column (toLowerCase) should be 'tax_category'"),
			Req: &pb.ImportTaxRequest{
				Payload: []byte(`tax_id,name,tax_percentage,wrong_header,default_flag,is_archived
				1,Tax 1,10,1,0,0
				2,Tax 2,20,2,1,1`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - fifth column name (toLowerCase) != default_flag",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - fifth column (toLowerCase) should be 'default_flag'"),
			Req: &pb.ImportTaxRequest{
				Payload: []byte(`tax_id,name,tax_percentage,tax_category,wrong_header,is_archived
				1,Tax 1,10,1,0,0
				2,Tax 2,20,2,1,1`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - sixth column name (toLowerCase) != is_archived",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - sixth column (toLowerCase) should be 'is_archived'"),
			Req: &pb.ImportTaxRequest{
				Payload: []byte(`tax_id,name,tax_percentage,tax_category,default_flag,wrong_header
				1,Tax 1,10,1,0,0
				2,Tax 2,20,2,1,1`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "parsing valid file (capitalized header still valid) with no error lines in response",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportTaxRequest{
				Payload: []byte(`TAX_ID,NAME,TAX_PERCENTAGE,TAX_CATEGORY,DEFAULT_FLAG,IS_ARCHIVED
				,Tax 1,10,1,0,0`),
			},
			ExpectedResp: &pb.ImportTaxResponse{
				Errors: []*pb.ImportTaxResponse_ImportTaxError{},
			},
			Setup: func(ctx context.Context) {
				mockTaxRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
			},
		},
		{
			Name: "update tax fields with no error lines in response",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportTaxRequest{
				Payload: []byte(`tax_id,name,tax_percentage,tax_category,default_flag,is_archived
				1,Tax 1,10,1,0,0`),
			},
			ExpectedResp: &pb.ImportTaxResponse{
				Errors: []*pb.ImportTaxResponse_ImportTaxError{},
			},
			Setup: func(ctx context.Context) {
				mockTaxRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
			},
		},
		{
			Name: "missing mandatory data (except ID), error lines in response",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportTaxRequest{
				Payload: []byte(`tax_id,name,tax_percentage,tax_category,default_flag,is_archived
				1,,10,1,0,0
				2,Tax 2,,1,0,0
				3,Tax 3,10,,0,0
				4,Tax 4,10,1,,0
				5,Tax 5,10,1,0,`),
			},
			ExpectedResp: &pb.ImportTaxResponse{
				Errors: []*pb.ImportTaxResponse_ImportTaxError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf(constant.UnableToParseTaxItem, fmt.Errorf("missing mandatory data: name")),
					},
					{
						RowNumber: 3,
						Error:     fmt.Sprintf(constant.UnableToParseTaxItem, fmt.Errorf("missing mandatory data: tax_percentage")),
					},
					{
						RowNumber: 4,
						Error:     fmt.Sprintf(constant.UnableToParseTaxItem, fmt.Errorf("missing mandatory data: tax_category")),
					},
					{
						RowNumber: 5,
						Error:     fmt.Sprintf(constant.UnableToParseTaxItem, fmt.Errorf("missing mandatory data: default_flag")),
					},
					{
						RowNumber: 6,
						Error:     fmt.Sprintf(constant.UnableToParseTaxItem, fmt.Errorf("missing mandatory data: is_archived")),
					},
				},
			},
			Setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				tx.On("Rollback", mock.Anything).Return(nil)
			},
		},
		{
			Name: "wrong number of data in a record",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportTaxRequest{
				Payload: []byte(`tax_id,name,tax_percentage,tax_category,default_flag,is_archived
				1,Tax 1,10,1,0,0
				3,Tax 3,1,0,0`),
			},
			ExpectedErr: status.Error(codes.InvalidArgument, "record on line 3: wrong number of fields"),
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "parsing default_flag and is_archived with error lines in response",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportTaxRequest{
				Payload: []byte(`tax_id,name,tax_percentage,tax_category,default_flag,is_archived
				1,Tax 1,11,1,3,0
				2,Tax 2,12,1,0,0
				3,Tax 3,13,1,1,0
				4,Tax 4,14,1,0,3
				5,Tax 5,15,1,0,0
				6,Tax 6,16,1,0,1`),
			},
			ExpectedResp: &pb.ImportTaxResponse{
				Errors: []*pb.ImportTaxResponse_ImportTaxError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf(constant.UnableToParseTaxItem, fmt.Errorf("error parsing default_flag: strconv.ParseBool: parsing \"3\": invalid syntax")),
					},
					{
						RowNumber: 5,
						Error:     fmt.Sprintf(constant.UnableToParseTaxItem, fmt.Errorf("error parsing is_archived: strconv.ParseBool: parsing \"3\": invalid syntax")),
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockTaxRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				mockTaxRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				mockTaxRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				mockTaxRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
			},
		},
		{
			Name: "parsing tax_percentage with error lines in response",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportTaxRequest{
				Payload: []byte(`tax_id,name,tax_percentage,tax_category,default_flag,is_archived
				,Tax 1,11,1,0,0
				,Tax 2,ABC,1,0,0`),
			},
			ExpectedResp: &pb.ImportTaxResponse{
				Errors: []*pb.ImportTaxResponse_ImportTaxError{
					{
						RowNumber: 3,
						Error:     fmt.Sprintf(constant.UnableToParseTaxItem, fmt.Errorf("error parsing tax_percentage: strconv.Atoi: parsing \"ABC\": invalid syntax")),
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockTaxRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
			},
		},
		{
			Name: "parsing tax_category with error lines in response",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportTaxRequest{
				Payload: []byte(`tax_id,name,tax_percentage,tax_category,default_flag,is_archived
				,Tax 1,11,0,0,0
				,Tax 2,12,1,0,0
				,Tax 3,13,2,0,0
				,Tax 4,14,3,0,0
				,Tax 5,15,ABC,0,0`),
			},
			ExpectedResp: &pb.ImportTaxResponse{
				Errors: []*pb.ImportTaxResponse_ImportTaxError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf(constant.UnableToParseTaxItem, fmt.Errorf("invalid tax_category: 0")),
					},
					{
						RowNumber: 5,
						Error:     fmt.Sprintf(constant.UnableToParseTaxItem, fmt.Errorf("invalid tax_category: 3")),
					},
					{
						RowNumber: 6,
						Error:     fmt.Sprintf(constant.UnableToParseTaxItem, fmt.Errorf("invalid tax_category: ABC")),
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockTaxRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				mockTaxRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
			},
		},
		{
			Name: "create/update tax category with error lines in response",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportTaxRequest{
				Payload: []byte(`tax_id,name,tax_percentage,tax_category,default_flag,is_archived
				,Tax 1,11,1,0,0
				,Tax 2,12,2,1,1
				3,Tax 3,13,2,0,0
				4,Tax 4,14,1,1,1`),
			},
			ExpectedResp: &pb.ImportTaxResponse{
				Errors: []*pb.ImportTaxResponse_ImportTaxError{
					{
						RowNumber: 3,
						Error:     fmt.Sprintf("unable to create new tax item: %s", pgx.ErrTxClosed),
					},
					{
						RowNumber: 5,
						Error:     fmt.Sprintf("unable to update tax item: %s", pgx.ErrTxClosed),
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockTaxRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				mockTaxRepo.On("Create", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTaxRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				mockTaxRepo.On("Update", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			resp, err := s.ImportTax(testCase.Ctx, testCase.Req.(*pb.ImportTaxRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)

				for i, expectedErr := range testCase.ExpectedResp.(*pb.ImportTaxResponse).Errors {
					assert.Equal(t, expectedErr.RowNumber, resp.Errors[i].RowNumber)
					assert.Contains(t, expectedErr.Error, resp.Errors[i].Error)
				}
			}

			mock.AssertExpectationsForObjects(t, db, mockTaxRepo)
		})
	}
}
