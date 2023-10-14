package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestInvoiceModifierService_ImportInvoiceSchedule(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID = "user-id"
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockTx := &mock_database.Tx{}
	mockDB := new(mock_database.Ext)
	mockInvoiceScheduleRepo := new(mock_repositories.MockInvoiceScheduleRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	invoiceScheduleEntity := &entities.InvoiceSchedule{
		InvoiceDate: database.Timestamptz(time.Now()),
	}

	s := &ImportMasterDataService{
		DB:                  mockDB,
		InvoiceScheduleRepo: mockInvoiceScheduleRepo,
		UnleashClient:       mockUnleashClient,
	}

	loc, _ := utils.GetTimeLocationByCountry(utils.CountryJp)
	now := utils.ResetTimeComponent(time.Now().In(loc))

	dt := now
	presentDate := fmt.Sprintf("%v/%02d/%02d", dt.Year(), int(dt.Month()), dt.Day())
	dt = dt.AddDate(0, 1, 0)
	futureDate := fmt.Sprintf("%v/%02d/%02d", dt.Year(), int(dt.Month()), dt.Day())
	dt = dt.AddDate(0, 1, 0)
	futureDateNxtMonth1 := fmt.Sprintf("%v/%02d/%02d", dt.Year(), int(dt.Month()), dt.Day())
	dt = dt.AddDate(0, 1, 0)
	futureDateNxtMonth2 := fmt.Sprintf("%v/%02d/%02d", dt.Year(), int(dt.Month()), dt.Day())
	dt = time.Now().AddDate(0, -1, 0)
	pastDate := fmt.Sprintf("%v/%02d/%02d", dt.Year(), int(dt.Month()), dt.Day())

	now = now.AddDate(0, 0, 1)
	futureDateNxtDay := fmt.Sprintf("%v/%02d/%02d", now.Year(), int(now.Month()), now.Day())

	csvOneDate := `invoice_schedule_id,invoice_date,is_archived,remarks
	,%v,,`
	csvMultDate := `invoice_schedule_id,invoice_date,is_archived,remarks
	,%v,,
	,%v,,
	,%v,,`
	csvOneIDToArchive := `invoice_schedule_id,invoice_date,is_archived,remarks
	1,,1,`
	csvMultDateMultError := `invoice_schedule_id,invoice_date,is_archived,remarks
	,%v,,
	,%v,,
	,%v,,
	,2023-12-15,,
	1,,,
	,,1,
	1,,yes,
	,,,remarks`
	csvInvalidHeaderCount := `schedule date, invoice date
	%v,%v`
	csvInvalidHeader := `invoice_schedule,invoice_date,is_archived,remarks
	%v,,,`
	csvBlankFirstLine := `invoice_schedule_id,invoice_date,is_archived,remarks
	`

	testcases := []TestCase{
		{
			name: "negative test - empty CSV file",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: nil,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "No data in CSV file"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - headers only",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(`invoice date`),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "No data in CSV file"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - Invalid header count",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(fmt.Sprintf(csvInvalidHeaderCount, futureDate, futureDate)),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "Invalid CSV format: number of column should be 4"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - Invalid header",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(fmt.Sprintf(csvInvalidHeader, futureDate)),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "Invalid CSV format: first column should be 'invoice_schedule_id'"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - InvoiceScheduleRepo.RetrieveInvoiceScheduleByID DB error",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(csvOneIDToArchive),
			},
			expectedResp: &invoice_pb.ImportInvoiceScheduleResponse{
				Errors: []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{
					{
						RowNumber: 2,
						Error:     "unable to parse invoice schedule: cannot find invoice_schedule_id with error 'tx is closed'",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockInvoiceScheduleRepo.On("RetrieveInvoiceScheduleByID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - InvoiceScheduleRepo.CancelAllSchedule DB error",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(csvOneIDToArchive),
			},
			expectedResp: &invoice_pb.ImportInvoiceScheduleResponse{
				Errors: []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{
					{
						RowNumber: 2,
						Error:     "unable to cancel schedule: tx is closed",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockInvoiceScheduleRepo.On("RetrieveInvoiceScheduleByID", ctx, mockDB, mock.Anything).Once().Return(invoiceScheduleEntity, nil)
				mockInvoiceScheduleRepo.On("CancelScheduleIfExists", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - InvoiceScheduleRepo.Update DB error",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(csvOneIDToArchive),
			},
			expectedResp: &invoice_pb.ImportInvoiceScheduleResponse{
				Errors: []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{
					{
						RowNumber: 2,
						Error:     "unable to update invoice schedule: tx is closed",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockInvoiceScheduleRepo.On("RetrieveInvoiceScheduleByID", ctx, mockDB, mock.Anything).Once().Return(invoiceScheduleEntity, nil)
				mockInvoiceScheduleRepo.On("CancelScheduleIfExists", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceScheduleRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - InvoiceScheduleRepo.Create DB error",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(fmt.Sprintf(csvOneDate, futureDate)),
			},
			expectedResp: &invoice_pb.ImportInvoiceScheduleResponse{
				Errors: []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{
					{
						RowNumber: 2,
						Error:     "unable to create invoice schedule: tx is closed",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockInvoiceScheduleRepo.On("CancelScheduleIfExists", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceScheduleRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - invalid date format",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(fmt.Sprintf(csvOneDate, "9999-99-99")),
			},
			expectedResp: &invoice_pb.ImportInvoiceScheduleResponse{
				Errors: []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{
					{
						RowNumber: 2,
						Error:     "unable to parse invoice schedule: invalid date format",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - invalid date value (present date)",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(fmt.Sprintf(csvOneDate, presentDate)),
			},
			expectedResp: &invoice_pb.ImportInvoiceScheduleResponse{
				Errors: []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{
					{
						RowNumber: 2,
						Error:     "unable to parse invoice schedule: invoice schedule should be a future date",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - invalid date value (past date)",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(fmt.Sprintf(csvOneDate, pastDate)),
			},
			expectedResp: &invoice_pb.ImportInvoiceScheduleResponse{
				Errors: []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{
					{
						RowNumber: 2,
						Error:     "unable to parse invoice schedule: invoice schedule should be a future date",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - empty first line",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(csvBlankFirstLine),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "record on line 2: wrong number of fields"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "negative test - mult errors",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(fmt.Sprintf(csvMultDateMultError, pastDate, presentDate, futureDate)),
			},
			expectedResp: &invoice_pb.ImportInvoiceScheduleResponse{
				Errors: []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{
					{
						RowNumber: 2,
						Error:     "unable to parse invoice schedule: invoice schedule should be a future date",
					},
					{
						RowNumber: 3,
						Error:     "unable to parse invoice schedule: invoice schedule should be a future date",
					},
					{
						RowNumber: 5,
						Error:     "unable to parse invoice schedule: invalid date format",
					},
					{
						RowNumber: 6,
						Error:     "unable to parse invoice schedule: invoice_schedule_id and is_archived can only be both present or absent",
					},
					{
						RowNumber: 7,
						Error:     "unable to parse invoice schedule: invoice_schedule_id and is_archived can only be both present or absent",
					},
					{
						RowNumber: 8,
						Error:     "unable to parse invoice schedule: invalid IsArchived value",
					},
					{
						RowNumber: 9,
						Error:     "unable to parse invoice schedule: invoice date is required",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockInvoiceScheduleRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceScheduleRepo.On("CancelScheduleIfExists", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy test - blank country",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(fmt.Sprintf(csvOneDate, futureDateNxtDay)),
			},
			expectedResp: &invoice_pb.ImportInvoiceScheduleResponse{
				Errors: []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockInvoiceScheduleRepo.On("CancelScheduleIfExists", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceScheduleRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy test - one invoice schedule",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(fmt.Sprintf(csvOneDate, futureDateNxtDay)),
			},
			expectedResp: &invoice_pb.ImportInvoiceScheduleResponse{
				Errors: []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockInvoiceScheduleRepo.On("CancelScheduleIfExists", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceScheduleRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy test - multiple invoice schedule",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(fmt.Sprintf(csvMultDate, futureDate, futureDateNxtMonth1, futureDateNxtMonth2)),
			},
			expectedResp: &invoice_pb.ImportInvoiceScheduleResponse{
				Errors: []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(false, nil)
				mockInvoiceScheduleRepo.On("CancelScheduleIfExists", ctx, mockTx, mock.Anything).Times(3).Return(nil)
				mockInvoiceScheduleRepo.On("Create", ctx, mockTx, mock.Anything).Times(3).Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			response, err := s.ImportInvoiceSchedule(testCase.ctx, testCase.req.(*invoice_pb.ImportInvoiceScheduleRequest))

			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)

				if response == nil {
					fmt.Println(err)
				}

				if testCase.expectedResp != nil {
					assert.Equal(t, compareImportInvoiceScheduleResponseErr(testCase.expectedResp.(*invoice_pb.ImportInvoiceScheduleResponse), response), true)
				}
			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockInvoiceScheduleRepo, mockTx)

		})
	}

}

func TestInvoiceModifierService_ImportInvoiceSchedule_WithEnableKECFeatureFlag(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID = "user-id"
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockTx := &mock_database.Tx{}
	mockDB := new(mock_database.Ext)
	mockInvoiceScheduleRepo := new(mock_repositories.MockInvoiceScheduleRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	invoiceScheduleEntity := &entities.InvoiceSchedule{
		InvoiceDate: database.Timestamptz(time.Now()),
	}

	s := &ImportMasterDataService{
		DB:                  mockDB,
		InvoiceScheduleRepo: mockInvoiceScheduleRepo,
		UnleashClient:       mockUnleashClient,
	}

	loc, _ := utils.GetTimeLocationByCountry(utils.CountryJp)
	now := utils.ResetTimeComponent(time.Now().In(loc))

	dt := now
	presentDate := fmt.Sprintf("%v/%02d/%02d", dt.Year(), int(dt.Month()), dt.Day())
	dt = dt.AddDate(0, 1, 0)
	futureDate := fmt.Sprintf("%v/%02d/%02d", dt.Year(), int(dt.Month()), dt.Day())
	dt = dt.AddDate(0, 1, 0)
	futureDateNxtMonth1 := fmt.Sprintf("%v/%02d/%02d", dt.Year(), int(dt.Month()), dt.Day())
	dt = dt.AddDate(0, 1, 0)
	futureDateNxtMonth2 := fmt.Sprintf("%v/%02d/%02d", dt.Year(), int(dt.Month()), dt.Day())
	dt = time.Now().AddDate(0, -1, 0)
	pastDate := fmt.Sprintf("%v/%02d/%02d", dt.Year(), int(dt.Month()), dt.Day())

	now = now.AddDate(0, 0, 1)
	futureDateNxtDay := fmt.Sprintf("%v/%02d/%02d", now.Year(), int(now.Month()), now.Day())

	csvOneDate := `invoice_schedule_id,invoice_date,is_archived,remarks
	,%v,,`
	csvMultDate := `invoice_schedule_id,invoice_date,is_archived,remarks
	,%v,,
	,%v,,
	,%v,,
	,%v,,`
	csvOneIDToArchive := `invoice_schedule_id,invoice_date,is_archived,remarks
	1,,1,`
	csvMultDateMultError := `invoice_schedule_id,invoice_date,is_archived,remarks
	,%v,,
	,%v,,
	,2023-12-15,,
	1,,,
	,,1,
	1,,yes,
	,,,remarks`
	csvInvalidHeaderCount := `schedule date, invoice date
	%v,%v`
	csvInvalidHeader := `invoice_schedule,invoice_date,is_archived,remarks
	%v,,,`
	csvBlankFirstLine := `invoice_schedule_id,invoice_date,is_archived,remarks
	`

	testcases := []TestCase{
		{
			name: "negative test - empty CSV file",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: nil,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "No data in CSV file"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - headers only",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(`invoice date`),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "No data in CSV file"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - Invalid header count",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(fmt.Sprintf(csvInvalidHeaderCount, futureDate, futureDate)),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "Invalid CSV format: number of column should be 4"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - Invalid header",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(fmt.Sprintf(csvInvalidHeader, futureDate)),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "Invalid CSV format: first column should be 'invoice_schedule_id'"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "negative test - InvoiceScheduleRepo.RetrieveInvoiceScheduleByID DB error",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(csvOneIDToArchive),
			},
			expectedResp: &invoice_pb.ImportInvoiceScheduleResponse{
				Errors: []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{
					{
						RowNumber: 2,
						Error:     "unable to parse invoice schedule: cannot find invoice_schedule_id with error 'tx is closed'",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
				mockInvoiceScheduleRepo.On("RetrieveInvoiceScheduleByID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - InvoiceScheduleRepo.CancelAllSchedule DB error",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(csvOneIDToArchive),
			},
			expectedResp: &invoice_pb.ImportInvoiceScheduleResponse{
				Errors: []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{
					{
						RowNumber: 2,
						Error:     "unable to cancel schedule: tx is closed",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
				mockInvoiceScheduleRepo.On("RetrieveInvoiceScheduleByID", ctx, mockDB, mock.Anything).Once().Return(invoiceScheduleEntity, nil)
				mockInvoiceScheduleRepo.On("CancelScheduleIfExists", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - InvoiceScheduleRepo.Update DB error",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(csvOneIDToArchive),
			},
			expectedResp: &invoice_pb.ImportInvoiceScheduleResponse{
				Errors: []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{
					{
						RowNumber: 2,
						Error:     "unable to update invoice schedule: tx is closed",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
				mockInvoiceScheduleRepo.On("RetrieveInvoiceScheduleByID", ctx, mockDB, mock.Anything).Once().Return(invoiceScheduleEntity, nil)
				mockInvoiceScheduleRepo.On("CancelScheduleIfExists", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceScheduleRepo.On("Update", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - InvoiceScheduleRepo.Create DB error",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(fmt.Sprintf(csvOneDate, futureDate)),
			},
			expectedResp: &invoice_pb.ImportInvoiceScheduleResponse{
				Errors: []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{
					{
						RowNumber: 2,
						Error:     "unable to create invoice schedule: tx is closed",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
				mockInvoiceScheduleRepo.On("CancelScheduleIfExists", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceScheduleRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - invalid date format",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(fmt.Sprintf(csvOneDate, "9999-99-99")),
			},
			expectedResp: &invoice_pb.ImportInvoiceScheduleResponse{
				Errors: []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{
					{
						RowNumber: 2,
						Error:     "unable to parse invoice schedule: invalid date format",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - invalid date value (past date)",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(fmt.Sprintf(csvOneDate, pastDate)),
			},
			expectedResp: &invoice_pb.ImportInvoiceScheduleResponse{
				Errors: []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{
					{
						RowNumber: 2,
						Error:     "unable to parse invoice schedule: invoice schedule should be a present date or future date",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - empty first line",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(csvBlankFirstLine),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "record on line 2: wrong number of fields"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "negative test - mult errors",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(fmt.Sprintf(csvMultDateMultError, pastDate, futureDate)),
			},
			expectedResp: &invoice_pb.ImportInvoiceScheduleResponse{
				Errors: []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{
					{
						RowNumber: 2,
						Error:     "unable to parse invoice schedule: invoice schedule should be a present date or future date",
					},
					{
						RowNumber: 4,
						Error:     "unable to parse invoice schedule: invalid date format",
					},
					{
						RowNumber: 5,
						Error:     "unable to parse invoice schedule: invoice_schedule_id and is_archived can only be both present or absent",
					},
					{
						RowNumber: 6,
						Error:     "unable to parse invoice schedule: invoice_schedule_id and is_archived can only be both present or absent",
					},
					{
						RowNumber: 7,
						Error:     "unable to parse invoice schedule: invalid IsArchived value",
					},
					{
						RowNumber: 8,
						Error:     "unable to parse invoice schedule: invoice date is required",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
				mockInvoiceScheduleRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceScheduleRepo.On("CancelScheduleIfExists", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy test - blank country",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(fmt.Sprintf(csvOneDate, futureDateNxtDay)),
			},
			expectedResp: &invoice_pb.ImportInvoiceScheduleResponse{
				Errors: []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
				mockInvoiceScheduleRepo.On("CancelScheduleIfExists", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceScheduleRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy test - one invoice schedule",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(fmt.Sprintf(csvOneDate, futureDateNxtDay)),
			},
			expectedResp: &invoice_pb.ImportInvoiceScheduleResponse{
				Errors: []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
				mockInvoiceScheduleRepo.On("CancelScheduleIfExists", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockInvoiceScheduleRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy test - multiple invoice schedule",
			ctx:  interceptors.ContextWithUserID(ctx, ctxUserID),
			req: &invoice_pb.ImportInvoiceScheduleRequest{
				Payload: []byte(fmt.Sprintf(csvMultDate, futureDate, presentDate, futureDateNxtMonth1, futureDateNxtMonth2)),
			},
			expectedResp: &invoice_pb.ImportInvoiceScheduleResponse{
				Errors: []*invoice_pb.ImportInvoiceScheduleResponse_ImportInvoiceScheduleError{},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableKECFeedbackPh1, mock.Anything).Once().Return(true, nil)
				mockInvoiceScheduleRepo.On("CancelScheduleIfExists", ctx, mockTx, mock.Anything).Times(4).Return(nil)
				mockInvoiceScheduleRepo.On("Create", ctx, mockTx, mock.Anything).Times(4).Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			response, err := s.ImportInvoiceSchedule(testCase.ctx, testCase.req.(*invoice_pb.ImportInvoiceScheduleRequest))

			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)

				if response == nil {
					fmt.Println(err)
				}

				if testCase.expectedResp != nil {
					assert.Equal(t, compareImportInvoiceScheduleResponseErr(testCase.expectedResp.(*invoice_pb.ImportInvoiceScheduleResponse), response), true)
				}
			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockInvoiceScheduleRepo, mockTx)

		})
	}

}

func compareImportInvoiceScheduleResponseErr(expectedResp *invoice_pb.ImportInvoiceScheduleResponse, actualResp *invoice_pb.ImportInvoiceScheduleResponse) bool {
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
