package invoicesvc

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	seqnumberservice "github.com/manabie-com/backend/internal/invoicemgmt/services/sequence_number"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	mock_sequence_number "github.com/manabie-com/backend/mock/invoicemgmt/services/sequence_number"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestInvoiceModifierService_IssueInvoiceV2(t *testing.T) {

	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDb := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockInvoiceActionLogRepo := new(mock_repositories.MockInvoiceActionLogRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	s := &InvoiceModifierService{
		DB:                   mockDb,
		InvoiceRepo:          mockInvoiceRepo,
		InvoiceActionLogRepo: mockInvoiceActionLogRepo,
		UnleashClient:        mockUnleashClient,
	}

	draftInvoice := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Type:      database.Text(invoice_pb.InvoiceType_MANUAL.String()),
		Total:     database.Numeric(100),
	}

	zeroTotalDraftInvoice := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Type:      database.Text(invoice_pb.InvoiceType_MANUAL.String()),
		Total:     database.Numeric(0),
	}

	failedInvoice := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_FAILED.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Type:      database.Text(invoice_pb.InvoiceType_MANUAL.String()),
		Total:     database.Numeric(100),
	}

	voidInvoice := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_VOID.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Type:      database.Text(invoice_pb.InvoiceType_MANUAL.String()),
		Total:     database.Numeric(100),
	}

	issuedInvoice := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_ISSUED.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Type:      database.Text(invoice_pb.InvoiceType_MANUAL.String()),
		Total:     database.Numeric(100),
	}

	paidInvoice := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_PAID.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Type:      database.Text(invoice_pb.InvoiceType_MANUAL.String()),
		Total:     database.Numeric(100),
	}

	refundedInvoice := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_REFUNDED.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Type:      database.Text(invoice_pb.InvoiceType_MANUAL.String()),
		Total:     database.Numeric(100),
	}

	testError := errors.New("test error")

	successfulResp := &invoice_pb.IssueInvoiceResponseV2{
		Successful: true,
	}

	testcases := []TestCase{
		{
			name: "happy case - DRAFT invoice",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId: "1",
				Remarks:   "test-remarks",
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(false, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(draftInvoice, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - FAILED invoice",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId: "1",
				Remarks:   "test-remarks",
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(false, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(failedInvoice, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - DRAFT invoice with zero total amount",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId: "1",
				Remarks:   "test-remarks",
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(false, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(zeroTotalDraftInvoice, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - empty invoice ID",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId: "   ",
				Remarks:   "test-remarks",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invoice ID cannot be empty"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(false, nil)

			},
		},
		{
			name: "negative test - invalid VOID status",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId: "1",
				Remarks:   "test-remarks",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid invoice status"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(false, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(voidInvoice, nil)
			},
		},
		{
			name: "negative test - invalid PAID status",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId: "1",
				Remarks:   "test-remarks",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid invoice status"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(false, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(paidInvoice, nil)
			},
		},
		{
			name: "negative test - invalid ISSUED status",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId: "1",
				Remarks:   "test-remarks",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid invoice status"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(false, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(issuedInvoice, nil)
			},
		},
		{
			name: "negative test - invalid REFUNDED status",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId: "1",
				Remarks:   "test-remarks",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid invoice status"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(false, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(refundedInvoice, nil)
			},
		},
		{
			name: "negative test - InvoiceRepo.RetrieveInvoiceByInvoiceID failed",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId: "1",
				Remarks:   "test-remarks",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, testError.Error()),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(false, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "negative test - InvoiceRepo.Update failed",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId: "1",
				Remarks:   "test-remarks",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, testError.Error()),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(false, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(draftInvoice, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - mockInvoiceActionLogRepo.Create failed",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId: "1",
				Remarks:   "test-remarks",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, testError.Error()),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(false, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(draftInvoice, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := s.IssueInvoiceV2(testCase.ctx, testCase.req.(*invoice_pb.IssueInvoiceRequestV2))

			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.Equal(t, testCase.expectedResp, response)
			}

			mock.AssertExpectationsForObjects(t, mockDb, mockInvoiceRepo, mockInvoiceActionLogRepo)
		})
	}

}

func TestInvoiceModifierService_IssueInvoiceV2WithPayment(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDb := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockInvoiceActionLogRepo := new(mock_repositories.MockInvoiceActionLogRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockBankAccountRepo := new(mock_repositories.MockBankAccountRepo)

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	mockSeqNumberService := new(mock_sequence_number.ISequenceNumberService)

	s := &InvoiceModifierService{
		DB:                    mockDb,
		InvoiceRepo:           mockInvoiceRepo,
		InvoiceActionLogRepo:  mockInvoiceActionLogRepo,
		UnleashClient:         mockUnleashClient,
		PaymentRepo:           mockPaymentRepo,
		BankAccountRepo:       mockBankAccountRepo,
		SequenceNumberService: mockSeqNumberService,
	}

	concretePaymentSeqNumberService := &seqnumberservice.PaymentSequenceNumberService{
		PaymentRepo: mockPaymentRepo,
	}

	verifiedBankAccount := &entities.BankAccount{
		IsVerified: database.Bool(true),
	}

	unVerifiedBankAccount := &entities.BankAccount{
		IsVerified: database.Bool(false),
	}

	draftInvoice := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Type:      database.Text(invoice_pb.InvoiceType_MANUAL.String()),
		Total:     database.Numeric(100),
	}

	zeroTotalDraftInvoice := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Type:      database.Text(invoice_pb.InvoiceType_MANUAL.String()),
		Total:     database.Numeric(0),
	}

	negativeTotalDraftInvoice := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_DRAFT.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Type:      database.Text(invoice_pb.InvoiceType_MANUAL.String()),
		Total:     database.Numeric(-100),
	}

	voidInvoice := &entities.Invoice{
		InvoiceID: database.Text("123"),
		Status:    database.Text(invoice_pb.InvoiceStatus_VOID.String()),
		CreatedAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour)},
		Type:      database.Text(invoice_pb.InvoiceType_MANUAL.String()),
		Total:     database.Numeric(100),
	}

	testError := errors.New("test error")

	successfulResp := &invoice_pb.IssueInvoiceResponseV2{
		Successful: true,
	}

	testcases := []TestCase{
		{
			name: "happy case auto set of payment seq number - DRAFT invoice - Convenience store payment method",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId:     "1",
				Remarks:       "test-remarks",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        100,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(true, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(draftInvoice, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(false, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)

				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(&entities.Payment{
					PaymentSequenceNumber: database.Int4(1),
					PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
				}, nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - DRAFT invoice - Convenience store payment method",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId:     "1",
				Remarks:       "test-remarks",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        100,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(true, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(draftInvoice, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(&entities.Payment{
					PaymentSequenceNumber: database.Int4(1),
					PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
				}, nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - DRAFT invoice - CASH payment method",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId:     "1",
				Remarks:       "test-remarks",
				PaymentMethod: invoice_pb.PaymentMethod_CASH,
				Amount:        100,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(true, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(draftInvoice, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(&entities.Payment{
					PaymentSequenceNumber: database.Int4(1),
					PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CASH.String()),
				}, nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - DRAFT invoice - BANK TRANSFER payment method",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId:     "1",
				Remarks:       "test-remarks",
				PaymentMethod: invoice_pb.PaymentMethod_BANK_TRANSFER,
				Amount:        100,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(true, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(draftInvoice, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(&entities.Payment{
					PaymentSequenceNumber: database.Int4(1),
					PaymentMethod:         database.Text(invoice_pb.PaymentMethod_BANK_TRANSFER.String()),
				}, nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - DRAFT invoice - DIRECT DEBIT payment method",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId:     "1",
				Remarks:       "test-remarks",
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				Amount:        100,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(true, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(draftInvoice, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDb, mock.Anything).Once().Return(verifiedBankAccount, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(&entities.Payment{
					PaymentSequenceNumber: database.Int4(1),
					PaymentMethod:         database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
				}, nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - DRAFT invoice with zero total amount",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId: "1",
				Remarks:   "test-remarks",
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(true, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(zeroTotalDraftInvoice, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "happy case - DRAFT invoice with negative total amount",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId: "1",
				Remarks:   "test-remarks",
			},
			expectedResp: successfulResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(true, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(negativeTotalDraftInvoice, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - unverified bank account issuing DIRECT DEBIT payment method",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId:     "1",
				Remarks:       "test-remarks",
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				Amount:        100,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
			},
			expectedErr: status.Error(codes.InvalidArgument, "student bank account is not verified"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(true, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(draftInvoice, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDb, mock.Anything).Once().Return(unVerifiedBankAccount, nil)
			},
		},
		{
			name: "negative test - no bank account issuing DIRECT DEBIT payment method",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId:     "1",
				Remarks:       "test-remarks",
				PaymentMethod: invoice_pb.PaymentMethod_DIRECT_DEBIT,
				Amount:        100,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
			},
			expectedErr: status.Error(codes.InvalidArgument, "student has no bank account registered"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(true, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(draftInvoice, nil)
				mockBankAccountRepo.On("FindByStudentID", ctx, mockDb, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "negative test - empty invoice ID",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId:     "   ",
				Remarks:       "test-remarks",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        100,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invoice ID cannot be empty"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(true, nil)

			},
		},
		{
			name: "negative test - due date is past date",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId:     "1",
				Remarks:       "test-remarks",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        100,
				DueDate:       timestamppb.New(time.Now().Add(-1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid date: DueDate must be today or after"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(true, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(draftInvoice, nil)
			},
		},
		{
			name: "negative test - expiry date is past date",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId:     "1",
				Remarks:       "test-remarks",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        100,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(-1 * time.Hour)),
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid date: ExpiryDate must be today or after"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(true, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(draftInvoice, nil)
			},
		},
		{
			name: "negative test - expiry date is less than due date",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId:     "1",
				Remarks:       "test-remarks",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        100,
				DueDate:       timestamppb.New(time.Now().Add(2 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid date: DueDate must be before ExpiryDate"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(true, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(draftInvoice, nil)
			},
		},
		{
			name: "negative test - invalid VOID status",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId:     "1",
				Remarks:       "test-remarks",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        100,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid invoice status"),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(true, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(voidInvoice, nil)
			},
		},
		{
			name: "negative test - InvoiceRepo.RetrieveInvoiceByInvoiceID failed",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId:     "1",
				Remarks:       "test-remarks",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        100,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, testError.Error()),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(true, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(nil, testError)
			},
		},
		{
			name: "negative test - InvoiceRepo.Update failed",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId:     "1",
				Remarks:       "test-remarks",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        100,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, testError.Error()),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(true, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(draftInvoice, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - mockInvoiceActionLogRepo.Create failed",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId:     "1",
				Remarks:       "test-remarks",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        100,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, testError.Error()),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(true, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(draftInvoice, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - mockPaymentRepo.Create failed",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId:     "1",
				Remarks:       "test-remarks",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        100,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, testError.Error()),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(true, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(draftInvoice, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - mockPaymentRepo.GetLatestPaymentDueDateByInvoiceID failed",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId:     "1",
				Remarks:       "test-remarks",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        100,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, testError.Error()),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(true, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(draftInvoice, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(nil, testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "negative test - mockInvoiceActionLogRepo.Create failed when creating second action log",
			ctx:  ctx,
			req: &invoice_pb.IssueInvoiceRequestV2{
				InvoiceId:     "1",
				Remarks:       "test-remarks",
				PaymentMethod: invoice_pb.PaymentMethod_CONVENIENCE_STORE,
				Amount:        100,
				DueDate:       timestamppb.New(time.Now().Add(1 * time.Hour)),
				ExpiryDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, testError.Error()),
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableSingleIssueInvoiceWithPayment, mock.Anything).Once().Return(true, nil)

				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDb, mock.Anything).Once().Return(draftInvoice, nil)
				mockDb.On("Begin", ctx).Once().Return(mockTx, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnablePaymentSequenceNumberManualSetting, mock.Anything).Once().Return(true, nil)
				mockSeqNumberService.On("GetPaymentSequenceNumberService").Once().Return(concretePaymentSeqNumberService)
				mockPaymentRepo.On("GetLatestPaymentSequenceNumber", ctx, mockTx).Once().Return(int32(1), nil)

				mockInvoiceRepo.On("Update", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockPaymentRepo.On("GetLatestPaymentDueDateByInvoiceID", ctx, mockTx, mock.Anything).Once().Return(&entities.Payment{
					PaymentSequenceNumber: database.Int4(1),
					PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
				}, nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(testError)
				mockTx.On("Rollback", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			response, err := s.IssueInvoiceV2(testCase.ctx, testCase.req.(*invoice_pb.IssueInvoiceRequestV2))

			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.Equal(t, testCase.expectedResp, response)
			}

			mock.AssertExpectationsForObjects(t, mockDb, mockInvoiceRepo, mockInvoiceActionLogRepo, mockUnleashClient, mockPaymentRepo, mockBankAccountRepo)
		})
	}

}
