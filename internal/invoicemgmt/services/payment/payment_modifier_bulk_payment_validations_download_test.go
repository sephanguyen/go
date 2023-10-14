package paymentsvc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
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

func TestPaymentModifierService_DownloadBulkPaymentValidationsDetail(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockUserBasicInfoRepo := new(mock_repositories.MockUserBasicInfoRepo)
	mockBulkPaymentValidationsRepo := new(mock_repositories.MockBulkPaymentValidationsRepo)
	mockBulkPaymentValidationsDetailRepo := new(mock_repositories.MockBulkPaymentValidationsDetailRepo)

	s := &PaymentModifierService{
		DB:                               mockDB,
		InvoiceRepo:                      mockInvoiceRepo,
		PaymentRepo:                      mockPaymentRepo,
		UserBasicInfoRepo:                mockUserBasicInfoRepo,
		BulkPaymentValidationsRepo:       mockBulkPaymentValidationsRepo,
		BulkPaymentValidationsDetailRepo: mockBulkPaymentValidationsDetailRepo,
	}

	mockBulkPaymentValidations := &entities.BulkPaymentValidations{
		BulkPaymentValidationsID: database.Text("bulk-payment-validations-multi"),
		PaymentMethod:            database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
		SuccessfulPayments:       database.Int4(1),
		FailedPayments:           database.Int4(1),
	}

	mockBulkPaymentValidationsDetailOne := &entities.BulkPaymentValidationsDetail{
		BulkPaymentValidationsID:       mockBulkPaymentValidations.BulkPaymentValidationsID,
		BulkPaymentValidationsDetailID: database.Text("bulk-payment-validations-detail-success-multiOne"),
		InvoiceID:                      database.Text("invoice-success-multiOne"),
		PaymentID:                      database.Text("payment-success-multiOne"),
		ValidatedResultCode:            database.Text(""),
	}

	mockBulkPaymentValidationsDetailTwo := &entities.BulkPaymentValidationsDetail{
		BulkPaymentValidationsID:       mockBulkPaymentValidations.BulkPaymentValidationsID,
		BulkPaymentValidationsDetailID: database.Text("bulk-payment-validations-detail-success-multiTwo"),
		InvoiceID:                      database.Text("invoice-success-multiTwo"),
		PaymentID:                      database.Text("payment-success-multiTwo"),
		ValidatedResultCode:            database.Text(""),
	}

	mockInvoiceOne := &entities.Invoice{
		InvoiceID:             database.Text("invoice-success-multiOne"),
		InvoiceSequenceNumber: database.Int4(1),
		StudentID:             database.Text("student-success-multiOne"),
	}

	mockInvoiceTwo := &entities.Invoice{
		InvoiceID:             database.Text("invoice-success-multiTwo"),
		InvoiceSequenceNumber: database.Int4(2),
		StudentID:             database.Text("student-success-multiTwo"),
	}

	mockPaymentOne := &entities.Payment{
		PaymentID:             database.Text("payment-success-multiOne"),
		InvoiceID:             mockInvoiceOne.InvoiceID,
		PaymentSequenceNumber: database.Int4(1),
		PaymentMethod:         database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
	}
	mockPaymentOne.Amount.Set(1223.87)

	mockPaymentTwo := &entities.Payment{
		PaymentID:             database.Text("payment-success-multiTwo"),
		InvoiceID:             mockInvoiceTwo.InvoiceID,
		PaymentSequenceNumber: database.Int4(2),
		PaymentMethod:         database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
	}
	mockPaymentTwo.Amount.Set(555.87)

	mockUserOne := &entities.UserBasicInfo{
		UserID: mockInvoiceOne.StudentID,
		Name:   database.Text("Albert Einstein"),
	}

	mockUserTwo := &entities.UserBasicInfo{
		UserID: mockInvoiceTwo.StudentID,
		Name:   database.Text("Test Another Einstein"),
	}

	// handling single success validation detail only
	singleSuccessBulkPaymentValidations := &entities.BulkPaymentValidations{
		BulkPaymentValidationsID: database.Text("bulk-payment-validations-success-single"),
		PaymentMethod:            database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
		SuccessfulPayments:       database.Int4(1),
		FailedPayments:           database.Int4(0),
	}

	singleSuccessBulkPaymentValidationsDetail := &entities.BulkPaymentValidationsDetail{
		BulkPaymentValidationsID:       singleSuccessBulkPaymentValidations.BulkPaymentValidationsID,
		BulkPaymentValidationsDetailID: database.Text("bulk-payment-validations-detail-success-single"),
		InvoiceID:                      database.Text("invoice-success-single"),
		PaymentID:                      database.Text("payment-success-single"),
		ValidatedResultCode:            database.Text(""),
	}

	singleSuccessInvoice := &entities.Invoice{
		InvoiceID:             database.Text("invoice-success-single"),
		InvoiceSequenceNumber: database.Int4(1),
		StudentID:             database.Text("student-success-single"),
	}

	singleSuccessPayment := &entities.Payment{
		PaymentID:             database.Text("payment-success-single"),
		InvoiceID:             database.Text("invoice-success-single"),
		PaymentSequenceNumber: database.Int4(1),
		PaymentMethod:         database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
	}
	singleSuccessPayment.Amount.Set(12000036.87)

	singleSuccessUser := &entities.UserBasicInfo{
		UserID: singleSuccessInvoice.StudentID,
		Name:   database.Text("Albert Einstein"),
	}

	testcases := []TestCase{
		{
			name: "happy case payment validation detail",
			ctx:  ctx,
			req: &invoice_pb.DownloadBulkPaymentValidationsDetailRequest{
				BulkPaymentValidationsId: "test",
			},
			expectedResp: &invoice_pb.DownloadBulkPaymentValidationsDetailResponse{
				PaymentValidationDetail: []*invoice_pb.ImportPaymentValidationDetail{
					{
						PaymentSequenceNumber: mockPaymentOne.PaymentSequenceNumber.Int,
						Result:                mockBulkPaymentValidationsDetailOne.ValidatedResultCode.String,
						Amount:                1223.87,
						StudentId:             mockUserOne.UserID.String,
						StudentName:           mockUserOne.Name.String,
						PaymentMethod:         constant.PaymentMethodsConvertToEnums[mockPaymentOne.PaymentMethod.String],
						InvoiceSequenceNumber: mockInvoiceOne.InvoiceSequenceNumber.Int,
						InvoiceId:             mockInvoiceOne.InvoiceID.String,
					},
					{
						PaymentSequenceNumber: mockPaymentTwo.PaymentSequenceNumber.Int,
						Result:                mockBulkPaymentValidationsDetailTwo.ValidatedResultCode.String,
						Amount:                555.87,
						StudentId:             mockUserTwo.UserID.String,
						StudentName:           mockUserTwo.Name.String,
						PaymentMethod:         constant.PaymentMethodsConvertToEnums[mockPaymentTwo.PaymentMethod.String],
						InvoiceSequenceNumber: mockInvoiceTwo.InvoiceSequenceNumber.Int,
						InvoiceId:             mockInvoiceTwo.InvoiceID.String,
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBulkPaymentValidationsRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentValidations, nil)
				mockBulkPaymentValidationsDetailRepo.On("RetrieveRecordsByBulkPaymentValidationsID", ctx, mockDB, mock.Anything).Once().Return([]*entities.BulkPaymentValidationsDetail{mockBulkPaymentValidationsDetailOne, mockBulkPaymentValidationsDetailTwo}, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(mockInvoiceOne, nil)
				mockPaymentRepo.On("FindByPaymentID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentOne, nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(mockUserOne, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(mockInvoiceTwo, nil)
				mockPaymentRepo.On("FindByPaymentID", ctx, mockDB, mock.Anything).Once().Return(mockPaymentTwo, nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(mockUserTwo, nil)
			},
		},
		{
			name: "happy case single validation detail",
			ctx:  ctx,
			req: &invoice_pb.DownloadBulkPaymentValidationsDetailRequest{
				BulkPaymentValidationsId: "test",
			},
			expectedResp: &invoice_pb.DownloadBulkPaymentValidationsDetailResponse{
				PaymentValidationDetail: []*invoice_pb.ImportPaymentValidationDetail{
					{
						PaymentSequenceNumber: singleSuccessPayment.PaymentSequenceNumber.Int,
						Result:                singleSuccessBulkPaymentValidationsDetail.ValidatedResultCode.String,
						Amount:                12000036.87,
						StudentId:             singleSuccessUser.UserID.String,
						StudentName:           singleSuccessUser.Name.String,
						PaymentMethod:         constant.PaymentMethodsConvertToEnums[singleSuccessBulkPaymentValidations.PaymentMethod.String],
						InvoiceSequenceNumber: singleSuccessInvoice.InvoiceSequenceNumber.Int,
						InvoiceId:             singleSuccessInvoice.InvoiceID.String,
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockBulkPaymentValidationsRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(singleSuccessBulkPaymentValidations, nil)
				mockBulkPaymentValidationsDetailRepo.On("RetrieveRecordsByBulkPaymentValidationsID", ctx, mockDB, mock.Anything).Once().Return([]*entities.BulkPaymentValidationsDetail{singleSuccessBulkPaymentValidationsDetail}, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(singleSuccessInvoice, nil)
				mockPaymentRepo.On("FindByPaymentID", ctx, mockDB, mock.Anything).Once().Return(singleSuccessPayment, nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(singleSuccessUser, nil)
			},
		},
		{
			name: "failed invalid validations id empty",
			ctx:  ctx,
			req: &invoice_pb.DownloadBulkPaymentValidationsDetailRequest{
				BulkPaymentValidationsId: "",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.FailedPrecondition, "invalid empty bulk payment validations id"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "failed no associated payment validation detail records",
			ctx:  ctx,
			req: &invoice_pb.DownloadBulkPaymentValidationsDetailRequest{
				BulkPaymentValidationsId: "test",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("error no associated bulk payment validation detail records")),
			setup: func(ctx context.Context) {
				mockBulkPaymentValidationsRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentValidations, nil)
				mockBulkPaymentValidationsDetailRepo.On("RetrieveRecordsByBulkPaymentValidationsID", ctx, mockDB, mock.Anything).Once().Return([]*entities.BulkPaymentValidationsDetail{}, nil)
			},
		},
		{
			name: "failed total validation records not match",
			ctx:  ctx,
			req: &invoice_pb.DownloadBulkPaymentValidationsDetailRequest{
				BulkPaymentValidationsId: "test",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "bulk payment validations detail records count not match expected 2 got 1 on bulk payment validations id test"),
			setup: func(ctx context.Context) {
				mockBulkPaymentValidationsRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentValidations, nil)
				mockBulkPaymentValidationsDetailRepo.On("RetrieveRecordsByBulkPaymentValidationsID", ctx, mockDB, mock.Anything).Once().Return([]*entities.BulkPaymentValidationsDetail{mockBulkPaymentValidationsDetailOne}, nil)
			},
		},
		{
			name: "negative test - finding bulk payment validations tx closed",
			ctx:  ctx,
			req: &invoice_pb.DownloadBulkPaymentValidationsDetailRequest{
				BulkPaymentValidationsId: "test",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "tx is closed"),
			setup: func(ctx context.Context) {
				mockBulkPaymentValidationsRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name: "negative test - finding bulk payment validations no rows in result set",
			ctx:  ctx,
			req: &invoice_pb.DownloadBulkPaymentValidationsDetailRequest{
				BulkPaymentValidationsId: "test",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "no rows in result set"),
			setup: func(ctx context.Context) {
				mockBulkPaymentValidationsRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "negative test - retrieve bulk payment validations detail tx closed",
			ctx:  ctx,
			req: &invoice_pb.DownloadBulkPaymentValidationsDetailRequest{
				BulkPaymentValidationsId: "test",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "tx is closed"),
			setup: func(ctx context.Context) {
				mockBulkPaymentValidationsRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentValidations, nil)

				mockBulkPaymentValidationsDetailRepo.On("RetrieveRecordsByBulkPaymentValidationsID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name: "negative test - retrieve bulk payment validations detail no rows in result set",
			ctx:  ctx,
			req: &invoice_pb.DownloadBulkPaymentValidationsDetailRequest{
				BulkPaymentValidationsId: "test",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "no rows in result set"),
			setup: func(ctx context.Context) {
				mockBulkPaymentValidationsRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(mockBulkPaymentValidations, nil)
				mockBulkPaymentValidationsDetailRepo.On("RetrieveRecordsByBulkPaymentValidationsID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "negative test - find invoice tx is closed",
			ctx:  ctx,
			req: &invoice_pb.DownloadBulkPaymentValidationsDetailRequest{
				BulkPaymentValidationsId: "test",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "tx is closed"),
			setup: func(ctx context.Context) {
				mockBulkPaymentValidationsRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(singleSuccessBulkPaymentValidations, nil)
				mockBulkPaymentValidationsDetailRepo.On("RetrieveRecordsByBulkPaymentValidationsID", ctx, mockDB, mock.Anything).Once().Return([]*entities.BulkPaymentValidationsDetail{singleSuccessBulkPaymentValidationsDetail}, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name: "negative test - find invoice no rows in result set",
			ctx:  ctx,
			req: &invoice_pb.DownloadBulkPaymentValidationsDetailRequest{
				BulkPaymentValidationsId: "test",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "no rows in result set"),
			setup: func(ctx context.Context) {
				mockBulkPaymentValidationsRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(singleSuccessBulkPaymentValidations, nil)
				mockBulkPaymentValidationsDetailRepo.On("RetrieveRecordsByBulkPaymentValidationsID", ctx, mockDB, mock.Anything).Once().Return([]*entities.BulkPaymentValidationsDetail{singleSuccessBulkPaymentValidationsDetail}, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "negative test - find payment tx closed",
			ctx:  ctx,
			req: &invoice_pb.DownloadBulkPaymentValidationsDetailRequest{
				BulkPaymentValidationsId: "test",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "tx is closed"),
			setup: func(ctx context.Context) {
				mockBulkPaymentValidationsRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(singleSuccessBulkPaymentValidations, nil)
				mockBulkPaymentValidationsDetailRepo.On("RetrieveRecordsByBulkPaymentValidationsID", ctx, mockDB, mock.Anything).Once().Return([]*entities.BulkPaymentValidationsDetail{singleSuccessBulkPaymentValidationsDetail}, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(singleSuccessInvoice, nil)
				mockPaymentRepo.On("FindByPaymentID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name: "negative test - find payment no rows in result set",
			ctx:  ctx,
			req: &invoice_pb.DownloadBulkPaymentValidationsDetailRequest{
				BulkPaymentValidationsId: "test",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "no rows in result set"),
			setup: func(ctx context.Context) {
				mockBulkPaymentValidationsRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(singleSuccessBulkPaymentValidations, nil)
				mockBulkPaymentValidationsDetailRepo.On("RetrieveRecordsByBulkPaymentValidationsID", ctx, mockDB, mock.Anything).Once().Return([]*entities.BulkPaymentValidationsDetail{singleSuccessBulkPaymentValidationsDetail}, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(singleSuccessInvoice, nil)
				mockPaymentRepo.On("FindByPaymentID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "negative test - find user no rows in result set",
			ctx:  ctx,
			req: &invoice_pb.DownloadBulkPaymentValidationsDetailRequest{
				BulkPaymentValidationsId: "test",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "no rows in result set"),
			setup: func(ctx context.Context) {
				mockBulkPaymentValidationsRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(singleSuccessBulkPaymentValidations, nil)
				mockBulkPaymentValidationsDetailRepo.On("RetrieveRecordsByBulkPaymentValidationsID", ctx, mockDB, mock.Anything).Once().Return([]*entities.BulkPaymentValidationsDetail{singleSuccessBulkPaymentValidationsDetail}, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(singleSuccessInvoice, nil)
				mockPaymentRepo.On("FindByPaymentID", ctx, mockDB, mock.Anything).Once().Return(singleSuccessPayment, nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "negative test - find user tx closed",
			ctx:  ctx,
			req: &invoice_pb.DownloadBulkPaymentValidationsDetailRequest{
				BulkPaymentValidationsId: "test",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "tx is closed"),
			setup: func(ctx context.Context) {
				mockBulkPaymentValidationsRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(singleSuccessBulkPaymentValidations, nil)
				mockBulkPaymentValidationsDetailRepo.On("RetrieveRecordsByBulkPaymentValidationsID", ctx, mockDB, mock.Anything).Once().Return([]*entities.BulkPaymentValidationsDetail{singleSuccessBulkPaymentValidationsDetail}, nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Once().Return(singleSuccessInvoice, nil)
				mockPaymentRepo.On("FindByPaymentID", ctx, mockDB, mock.Anything).Once().Return(singleSuccessPayment, nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := s.DownloadBulkPaymentValidationsDetail(testCase.ctx, testCase.req.(*invoice_pb.DownloadBulkPaymentValidationsDetailRequest))
			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)

				if response == nil {
					fmt.Println(err)
				}

				if testCase.expectedResp != nil {
					expectedResp := testCase.expectedResp.(*invoice_pb.DownloadBulkPaymentValidationsDetailResponse)

					assert.Equal(t, len(expectedResp.PaymentValidationDetail), len(response.PaymentValidationDetail))
					for i, r := range expectedResp.PaymentValidationDetail {
						assert.Equal(t, r.InvoiceSequenceNumber, response.PaymentValidationDetail[i].InvoiceSequenceNumber)
						assert.Equal(t, r.Amount, response.PaymentValidationDetail[i].Amount)
						assert.Equal(t, r.PaymentMethod, response.PaymentValidationDetail[i].PaymentMethod)
						assert.Equal(t, r.Result, response.PaymentValidationDetail[i].Result)
						assert.Equal(t, r.StudentId, response.PaymentValidationDetail[i].StudentId)
						assert.Equal(t, r.StudentName, response.PaymentValidationDetail[i].StudentName)
						assert.Equal(t, r.PaymentSequenceNumber, response.PaymentValidationDetail[i].PaymentSequenceNumber)
						assert.Equal(t, r.InvoiceId, response.PaymentValidationDetail[i].InvoiceId)
					}
				}

			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockInvoiceRepo, mockPaymentRepo, mockBulkPaymentValidationsDetailRepo, mockBulkPaymentValidationsRepo, mockUserBasicInfoRepo)
		})
	}

}
