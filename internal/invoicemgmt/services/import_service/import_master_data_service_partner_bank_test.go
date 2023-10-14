package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestInvoiceModifierService_ImportPartnerBank(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID = "user-id"
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockTx := &mock_database.Tx{}
	mockDB := new(mock_database.Ext)
	mockPartnerBankRepo := new(mock_repositories.MockPartnerBankRepo)

	s := &ImportMasterDataService{
		DB:              mockDB,
		PartnerBankRepo: mockPartnerBankRepo,
	}

	dt := time.Now()
	dt = dt.AddDate(0, 1, 0)

	// for invalid header count
	csvInvalidHeaderCount := `test_dummy, consignor_code
	%v,%v`

	// for invalid header value consignor_test
	csvInvalidHeader := `partner_bank_id,consignor_code,consignor_test,bank_number,bank_name,bank_branch_number,bank_branch_name,deposit_items,account_number,is_archived,remarks,is_default,record_limit
	,,,,,,,,,,,,`

	// if no values it will throw error
	csvJustheader := `partner_bank_id,consignor_code,consignor_name,bank_number,bank_name,bank_branch_number,bank_branch_name,deposit_items,account_number,is_archived,remarks,is_default,record_limit`

	// if no value it will throw error missing first mandatory field consignor code
	csvPartnerBankConsignorCodeEmpty := `partner_bank_id,consignor_code,consignor_name,bank_number,bank_name,bank_branch_number,bank_branch_name,deposit_items,account_number,is_archived,remarks,is_default,record_limit
	,,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,152,ｲｺﾏ,1,12345,,,,`

	csvPartnerBankMultipleValueEmpty := `partner_bank_id,consignor_code,consignor_name,bank_number,bank_name,bank_branch_number,bank_branch_name,deposit_items,account_number,is_archived,remarks,is_default,record_limit
	,,,,,,,,,,,,
	,0000004819,,,,,,,,,,,
	,0000004819,ｶ)ｹ-ｲ-ｼ-,,,,,,,,,,
	,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,,,,,,,,,
	,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,,,,,,,,
	,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,150,,,,,,,
	,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,150,ｲｺﾏ,,,,,,
	,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,150,ｲｺﾏ,1,,,,,`

	csvPartnerBankConsignorCodeLimit := `partner_bank_id,consignor_code,consignor_name,bank_number,bank_name,bank_branch_number,bank_branch_name,deposit_items,account_number,is_archived,remarks,is_default,record_limit
	,000000481923,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,150,ｲｺﾏ,1,123456,,,,`

	csvPartnerBankMultipleLimitValues := `partner_bank_id,consignor_code,consignor_name,bank_number,bank_name,bank_branch_number,bank_branch_name,deposit_items,account_number,is_archived,remarks,is_default,record_limit
	,000000481923,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,150,ｲｺﾏ,1,123456,,,,
	,0000004819,ｶ)ｹ-ｲ-ｼ-ｶ)ｹ-ｲ-ｼ-ｶ)ｹ-ｲ-ｼ-ｶ)ｹ-ｲ-ｼ-ｶ)ｹ-ｲ-ｼ-ｶ)ｹ-ｲ-ｼ-ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,150,ｲｺﾏ,1,123456,,,,
	,0000004819,ｶ)ｹ-ｲ-ｼ-,032674,ﾅﾝﾄ,150,ｲｺﾏ,1,123456,,,,
	,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄﾅﾝﾄ,150,ｲｺﾏ,1,123456,,,,
	,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,1520,ｲｺﾏ,1,123456,,,,
	,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,152,ｲｺﾏｲｺﾏｲｺﾏｲｺﾏｲｺﾏｲｺﾏｲｺﾏｲｺﾏｲｺﾏ,1,123456,,,,
	,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,152,ｲｺﾏ,"12",123456,,,,
	,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,152,ｲｺﾏ,1,"12345678",,,,`

	csvPartnerBankMultipleInvalidValues := `partner_bank_id,consignor_code,consignor_name,bank_number,bank_name,bank_branch_number,bank_branch_name,deposit_items,account_number,is_archived,remarks,is_default,record_limit
	,aswasad,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,150,ｲｺﾏ,1,1234567,,,,
	,0000004819,ｶ)ｹ-ｲ-ｼ-,wasss,ﾅﾝﾄ,150,ｲｺﾏ,1,1234567,,,,
	,0000004819,ｶ)ｹ-ｲ-ｼ-,521,ﾅﾝﾄ,testInvalid,ｲｺﾏ,1,1234567,,,,
	,0000004819,ｶ)ｹ-ｲ-ｼ-,521,ﾅﾝﾄ,150,ｲｺﾏ,"asd",1234567,,,,
	,0000004819,ｶ)ｹ-ｲ-ｼ-,521,ﾅﾝﾄ,150,ｲｺﾏ,"8",1234567,,,,
	,0000004819,ｶ)ｹ-ｲ-ｼ-,521,ﾅﾝﾄ,150,ｲｺﾏ,1,1234567,asd,,,
	,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,150,ｲｺﾏ,1,1234567,,,asd,`

	csvSingleRowValid := `partner_bank_id,consignor_code,consignor_name,bank_number,bank_name,bank_branch_number,bank_branch_name,deposit_items,account_number,is_archived,remarks,is_default,record_limit
	,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,152,ｲｺﾏ,1,1234567,,,,`

	csvSingleRowValidWithKatakana := `partner_bank_id,consignor_code,consignor_name,bank_number,bank_name,bank_branch_number,bank_branch_name,deposit_items,account_number,is_archived,remarks,is_default,record_limit
	,0000004819,この世界で幸せな荷主様,0326,ﾅﾝﾄ	こんにちは,152,ｲｺﾏ,1,1234567,,,,`

	csvSingleRowValidWithLimit := `partner_bank_id,consignor_code,consignor_name,bank_number,bank_name,bank_branch_number,bank_branch_name,deposit_items,account_number,is_archived,remarks,is_default,record_limit
	,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,152,ｲｺﾏ,1,1234567,,,,50`

	csvSingleRowInvalidWithLimit := `partner_bank_id,consignor_code,consignor_name,bank_number,bank_name,bank_branch_number,bank_branch_name,deposit_items,account_number,is_archived,remarks,is_default,record_limit
	,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,152,ｲｺﾏ,1,1234567,,,,invalid`

	csvSingleRowInvalidWithNegativeLimit := `partner_bank_id,consignor_code,consignor_name,bank_number,bank_name,bank_branch_number,bank_branch_name,deposit_items,account_number,is_archived,remarks,is_default,record_limit
	,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,152,ｲｺﾏ,1,1234567,,,,-1`

	csvMultipleRowValid := `partner_bank_id,consignor_code,consignor_name,bank_number,bank_name,bank_branch_number,bank_branch_name,deposit_items,account_number,is_archived,remarks,is_default,record_limit
	,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,152,ｲｺﾏ,1,1234567,,,,
	,0000004819,ｶ)ｹ-ｲ-ｼ-,2134,ﾅﾝﾄ,122,ｲｺﾏ,1,1234567,,,,
	,0000002331,ｶ)ｹ-ｲ-ｼ-,0612,ﾅﾝﾄ,333,ｲｺﾏ,4,1234567,,,,`

	csvMultipleRowValidWithLimit := `partner_bank_id,consignor_code,consignor_name,bank_number,bank_name,bank_branch_number,bank_branch_name,deposit_items,account_number,is_archived,remarks,is_default,record_limit
	,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,152,ｲｺﾏ,1,1234567,,,,
	,0000004819,ｶ)ｹ-ｲ-ｼ-,2134,ﾅﾝﾄ,122,ｲｺﾏ,1,1234567,,,,5000
	,0000002331,ｶ)ｹ-ｲ-ｼ-,0612,ﾅﾝﾄ,333,ｲｺﾏ,4,1234567,,,,100`

	// single record exist to archive
	partnerBankExistToArchive := &entities.PartnerBank{
		PartnerBankID: database.Text("123"),
	}
	csvSingleRowValidArchive := `partner_bank_id,consignor_code,consignor_name,bank_number,bank_name,bank_branch_number,bank_branch_name,deposit_items,account_number,is_archived,remarks,is_default,record_limit
	123,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,152,ｲｺﾏ,1,1234567,1,,,`

	// multiple record exist to archive
	partnerBankExistToArchiveTwo := &entities.PartnerBank{
		PartnerBankID: database.Text("1235"),
	}
	csvMultipleRowValidArchive := `partner_bank_id,consignor_code,consignor_name,bank_number,bank_name,bank_branch_number,bank_branch_name,deposit_items,account_number,is_archived,remarks,is_default,record_limit
	,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,152,ｲｺﾏ,1,1234567,,,,
	1235,0000004819,ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,152,ｲｺﾏ,1,1234567,1,,,`

	// Existing partner bank id but archive false
	csvPartnerBankExistingIDNotArchive := `partner_bank_id,consignor_code,consignor_name,bank_number,bank_name,bank_branch_number,bank_branch_name,deposit_items,account_number,is_archived,remarks,is_default,record_limit
	555,"000000481923",ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,150,ｲｺﾏ,1,123456,,,,`

	// Not Existing partner bank id but archive true
	csvPartnerBankNotExistIDArchive := `partner_bank_id,consignor_code,consignor_name,bank_number,bank_name,bank_branch_number,bank_branch_name,deposit_items,account_number,is_archived,remarks,is_default,record_limit
	,"000000481923",ｶ)ｹ-ｲ-ｼ-,0326,ﾅﾝﾄ,150,ｲｺﾏ,1,123456,1,,,`

	testcases := []TestCase{
		{
			name: "happy test - single row",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportPartnerBankRequest{
				Payload: []byte(csvSingleRowValid),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPartnerBankRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy test - single row with limit",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportPartnerBankRequest{
				Payload: []byte(csvSingleRowValidWithLimit),
			},
			expectedResp: &invoice_pb.ImportPartnerBankResponse{
				Errors: []*invoice_pb.ImportPartnerBankResponse_ImportPartnerBankError{},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPartnerBankRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy test - multiple row",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportPartnerBankRequest{
				Payload: []byte(csvMultipleRowValid),
			},
			expectedResp: &invoice_pb.ImportPartnerBankResponse{
				Errors: []*invoice_pb.ImportPartnerBankResponse_ImportPartnerBankError{},
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPartnerBankRepo.On("Upsert", ctx, mockTx, mock.Anything).Times(3).Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy test - multiple row with Limit",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportPartnerBankRequest{
				Payload: []byte(csvMultipleRowValidWithLimit),
			},
			expectedResp: &invoice_pb.ImportPartnerBankResponse{
				Errors: []*invoice_pb.ImportPartnerBankResponse_ImportPartnerBankError{},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPartnerBankRepo.On("Upsert", ctx, mockTx, mock.Anything).Times(3).Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - single row with katakana",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportPartnerBankRequest{
				Payload: []byte(csvSingleRowValidWithKatakana),
			},
			expectedResp: &invoice_pb.ImportPartnerBankResponse{
				Errors: []*invoice_pb.ImportPartnerBankResponse_ImportPartnerBankError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("unable to parse partner bank detail: %s", status.Error(codes.InvalidArgument, "consignor_name field has invalid half width character").Error()),
					},
				},
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - single row with invalid record limit",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportPartnerBankRequest{
				Payload: []byte(csvSingleRowInvalidWithLimit),
			},
			expectedResp: &invoice_pb.ImportPartnerBankResponse{
				Errors: []*invoice_pb.ImportPartnerBankResponse_ImportPartnerBankError{
					{
						RowNumber: 2,
						Error:     "unable to parse partner bank detail: invalid record_limit format",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - single row with invalid record negative limit",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportPartnerBankRequest{
				Payload: []byte(csvSingleRowInvalidWithNegativeLimit),
			},
			expectedResp: &invoice_pb.ImportPartnerBankResponse{
				Errors: []*invoice_pb.ImportPartnerBankResponse_ImportPartnerBankError{
					{
						RowNumber: 2,
						Error:     "unable to parse partner bank detail: invalid record limit: should be greater than or equal to 0",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy test - archive single row",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportPartnerBankRequest{
				Payload: []byte(csvSingleRowValidArchive),
			},
			expectedResp: &invoice_pb.ImportPartnerBankResponse{
				Errors: []*invoice_pb.ImportPartnerBankResponse_ImportPartnerBankError{},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPartnerBankRepo.On("RetrievePartnerBankByID", ctx, mockTx, mock.Anything).Once().Return(partnerBankExistToArchive, nil)
				mockPartnerBankRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy test - with archive multiple row",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportPartnerBankRequest{
				Payload: []byte(csvMultipleRowValidArchive),
			},
			expectedResp: &invoice_pb.ImportPartnerBankResponse{
				Errors: []*invoice_pb.ImportPartnerBankResponse_ImportPartnerBankError{},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPartnerBankRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPartnerBankRepo.On("RetrievePartnerBankByID", ctx, mockTx, mock.Anything).Once().Return(partnerBankExistToArchiveTwo, nil)
				mockPartnerBankRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - header no values csv",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportPartnerBankRequest{
				Payload: []byte(fmt.Sprintf(csvJustheader)),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "No data in CSV file"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - empty CSV file",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportPartnerBankRequest{
				Payload: nil,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "No data in CSV file"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - Invalid header count",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportPartnerBankRequest{
				Payload: []byte(fmt.Sprintf(csvInvalidHeaderCount, "test-1", "test-consignor-code")),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "Invalid CSV format: number of column should be 13"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - Invalid header",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportPartnerBankRequest{
				Payload: []byte(fmt.Sprintf(csvInvalidHeader)),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "Invalid CSV format: third column should be 'consignor_name'"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - multiple csv values required",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportPartnerBankRequest{
				Payload: []byte(fmt.Sprintf(csvPartnerBankMultipleValueEmpty)),
			},
			expectedResp: &invoice_pb.ImportPartnerBankResponse{
				Errors: []*invoice_pb.ImportPartnerBankResponse_ImportPartnerBankError{
					{
						RowNumber: 2,
						Error:     "unable to parse partner bank detail: field consignor_code is required",
					},
					{
						RowNumber: 3,
						Error:     "unable to parse partner bank detail: field consignor_name is required",
					},
					{
						RowNumber: 4,
						Error:     "unable to parse partner bank detail: field bank_number is required",
					},
					{
						RowNumber: 5,
						Error:     "unable to parse partner bank detail: field bank_name is required",
					},
					{
						RowNumber: 6,
						Error:     "unable to parse partner bank detail: field bank_branch_number is required",
					},
					{
						RowNumber: 7,
						Error:     "unable to parse partner bank detail: field bank_branch_name is required",
					},
					{
						RowNumber: 8,
						Error:     "unable to parse partner bank detail: field deposit_items is required",
					},
					{
						RowNumber: 9,
						Error:     "unable to parse partner bank detail: field account_number is required",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - single required field consignor code is empty",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportPartnerBankRequest{
				Payload: []byte(fmt.Sprintf(csvPartnerBankConsignorCodeEmpty)),
			},
			expectedResp: &invoice_pb.ImportPartnerBankResponse{
				Errors: []*invoice_pb.ImportPartnerBankResponse_ImportPartnerBankError{
					{
						RowNumber: 2,
						Error:     "unable to parse partner bank detail: field consignor_code is required",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - single limit consignor code",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportPartnerBankRequest{
				Payload: []byte(fmt.Sprintf(csvPartnerBankConsignorCodeLimit)),
			},
			expectedResp: &invoice_pb.ImportPartnerBankResponse{
				Errors: []*invoice_pb.ImportPartnerBankResponse_ImportPartnerBankError{
					{
						RowNumber: 2,
						Error:     "unable to parse partner bank detail: invalid consignor code digit limit",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - multiple csv values limit",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportPartnerBankRequest{
				Payload: []byte(fmt.Sprintf(csvPartnerBankMultipleLimitValues)),
			},
			expectedResp: &invoice_pb.ImportPartnerBankResponse{
				Errors: []*invoice_pb.ImportPartnerBankResponse_ImportPartnerBankError{
					{
						RowNumber: 2,
						Error:     "unable to parse partner bank detail: invalid consignor code digit limit",
					},
					{
						RowNumber: 3,
						Error:     "unable to parse partner bank detail: invalid consignor name character limit",
					},
					{
						RowNumber: 4,
						Error:     "unable to parse partner bank detail: invalid bank number digit limit",
					},
					{
						RowNumber: 5,
						Error:     "unable to parse partner bank detail: invalid bank name character limit",
					},
					{
						RowNumber: 6,
						Error:     "unable to parse partner bank detail: invalid bank branch number digit limit",
					},
					{
						RowNumber: 7,
						Error:     "unable to parse partner bank detail: invalid bank branch name character limit",
					},
					{
						RowNumber: 8,
						Error:     "unable to parse partner bank detail: invalid deposit items digit limit",
					},
					{
						RowNumber: 9,
						Error:     "unable to parse partner bank detail: the account number can only accept 7 digit numbers",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - multiple csv values invalid format",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportPartnerBankRequest{
				Payload: []byte(fmt.Sprintf(csvPartnerBankMultipleInvalidValues)),
			},
			expectedResp: &invoice_pb.ImportPartnerBankResponse{
				Errors: []*invoice_pb.ImportPartnerBankResponse_ImportPartnerBankError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("unable to parse partner bank detail: %s", status.Error(codes.InvalidArgument, "consignor_code field has invalid half width number").Error()),
					},
					{
						RowNumber: 3,
						Error:     fmt.Sprintf("unable to parse partner bank detail: %s", status.Error(codes.InvalidArgument, "bank_number field has invalid half width number").Error()),
					},
					{
						RowNumber: 4,
						Error:     fmt.Sprintf("unable to parse partner bank detail: %s", status.Error(codes.InvalidArgument, "bank_branch_number field has invalid half width number").Error()),
					},
					{
						RowNumber: 5,
						Error:     fmt.Sprintf("unable to parse partner bank detail: %s", status.Error(codes.InvalidArgument, "deposit_items field has invalid half width number").Error()),
					},
					{
						RowNumber: 6,
						Error:     "unable to parse partner bank detail: invalid deposit items account",
					},
					{
						RowNumber: 7,
						Error:     "unable to parse partner bank detail: invalid archive value",
					},
					{
						RowNumber: 8,
						Error:     "unable to parse partner bank detail: invalid default value",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - existing partner bank id but archive false",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportPartnerBankRequest{
				Payload: []byte(fmt.Sprintf(csvPartnerBankExistingIDNotArchive)),
			},
			expectedResp: &invoice_pb.ImportPartnerBankResponse{
				Errors: []*invoice_pb.ImportPartnerBankResponse_ImportPartnerBankError{
					{
						RowNumber: 2,
						Error:     "unable to parse partner bank detail: partner_bank_id and is_archived can only be both present or absent",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - not exist partner bank id but archive true",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportPartnerBankRequest{
				Payload: []byte(fmt.Sprintf(csvPartnerBankNotExistIDArchive)),
			},
			expectedResp: &invoice_pb.ImportPartnerBankResponse{
				Errors: []*invoice_pb.ImportPartnerBankResponse_ImportPartnerBankError{
					{
						RowNumber: 2,
						Error:     "unable to parse partner bank detail: partner_bank_id and is_archived can only be both present or absent",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - single row tx closed",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportPartnerBankRequest{
				Payload: []byte(csvSingleRowValid),
			},
			expectedResp: &invoice_pb.ImportPartnerBankResponse{
				Errors: []*invoice_pb.ImportPartnerBankResponse_ImportPartnerBankError{
					{
						RowNumber: 2,
						Error:     "unable to upsert partner bank tx is closed",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPartnerBankRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - multiple row with tx closed",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportPartnerBankRequest{
				Payload: []byte(csvMultipleRowValid),
			},
			expectedResp: &invoice_pb.ImportPartnerBankResponse{
				Errors: []*invoice_pb.ImportPartnerBankResponse_ImportPartnerBankError{
					{
						RowNumber: 4,
						Error:     "unable to upsert partner bank tx is closed",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPartnerBankRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPartnerBankRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPartnerBankRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - multiple row with with archive no rows",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportPartnerBankRequest{
				Payload: []byte(csvMultipleRowValidArchive),
			},
			expectedResp: &invoice_pb.ImportPartnerBankResponse{
				Errors: []*invoice_pb.ImportPartnerBankResponse_ImportPartnerBankError{
					{
						RowNumber: 3,
						Error:     "cannot find partner bank with error 'no rows in result set'",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockPartnerBankRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPartnerBankRepo.On("RetrievePartnerBankByID", ctx, mockTx, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			response, err := s.ImportPartnerBank(testCase.ctx, testCase.req.(*invoice_pb.ImportPartnerBankRequest))

			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)

				if response == nil {
					fmt.Println(err)
				}

				if testCase.expectedResp != nil {
					assert.Equal(t, compareImportPartnerBankResponseErr(testCase.expectedResp.(*invoice_pb.ImportPartnerBankResponse), response), true)
				}
			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}

		})
	}

}

func compareImportPartnerBankResponseErr(expectedResp *invoice_pb.ImportPartnerBankResponse, actualResp *invoice_pb.ImportPartnerBankResponse) bool {
	if len(expectedResp.Errors) != len(actualResp.Errors) {
		fmt.Printf("Errors length: expected %v but got %v\n", len(expectedResp.Errors), len(actualResp.Errors))
		fmt.Println(actualResp)
		return false
	}

	for i := 0; i < len(expectedResp.Errors); i++ {
		if expectedResp.Errors[i].RowNumber != actualResp.Errors[i].RowNumber {
			fmt.Printf("RowNumber: expected %v but got %v at line %v\n", expectedResp.Errors[i].RowNumber, actualResp.Errors[i].RowNumber, i+1)
			return false
		}

		if expectedResp.Errors[i].Error != actualResp.Errors[i].Error {
			fmt.Printf("Error: expected %v but got %v at line %v\n", expectedResp.Errors[i].Error, actualResp.Errors[i].Error, i+1)
			return false
		}
	}

	return true
}
