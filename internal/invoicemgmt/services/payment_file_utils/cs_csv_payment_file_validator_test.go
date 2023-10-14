package paymentfileutils

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testConvenienceStoreFileContent struct {
	Amount       int
	CodeForUser2 string
	Category     string
	CreatedDate  int
}

func TestConvenienceStoreCSVPaymentFileValidator_Validate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockInvoiceActionLogRepo := new(mock_repositories.MockInvoiceActionLogRepo)
	mockBulkValidationsRepo := new(mock_repositories.MockBulkPaymentValidationsRepo)
	mockBulkValidationsDetailRepo := new(mock_repositories.MockBulkPaymentValidationsDetailRepo)
	mockUserBasicInfoRepo := new(mock_repositories.MockUserBasicInfoRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	baseValidator := &BasePaymentFileValidator{
		DB:                               mockDB,
		InvoiceRepo:                      mockInvoiceRepo,
		PaymentRepo:                      mockPaymentRepo,
		InvoiceActionLogRepo:             mockInvoiceActionLogRepo,
		BulkPaymentValidationsRepo:       mockBulkValidationsRepo,
		BulkPaymentValidationsDetailRepo: mockBulkValidationsDetailRepo,
		UserBasicInfoRepo:                mockUserBasicInfoRepo,
		UnleashClient:                    mockUnleashClient,
	}

	convenienceStoreValidator := &ConvenienceStoreCSVPaymentFileValidator{
		BasePaymentFileValidator: baseValidator,
	}

	invoiceIssued := invoice_pb.InvoiceStatus_ISSUED.String()
	invoiceFailed := invoice_pb.InvoiceStatus_FAILED.String()
	paymentPending := invoice_pb.PaymentStatus_PAYMENT_PENDING.String()
	paymentFailed := invoice_pb.PaymentStatus_PAYMENT_FAILED.String()

	// Data for scenario 3 which have category equal to 02
	scenario3Invoice, scenario3Payment, scenario3User, scenario3FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100},
			InvoiceStatuses:     []string{invoiceIssued},
			PaymentStatuses:     []string{paymentPending},
			ExistingResultCodes: []string{""},
			ResultCodes:         []string{"02"},
			DepositAmount:       []int{100},
		},
	)

	// Data for scenario 4 which have category equal to 01
	scenario4Invoice, scenario4Payment, scenario4User, scenario4FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100},
			InvoiceStatuses:     []string{invoiceIssued},
			PaymentStatuses:     []string{paymentPending},
			ExistingResultCodes: []string{""},
			ResultCodes:         []string{"01"},
			DepositAmount:       []int{100},
		},
	)

	// Data for scenario 5 which have category equal to 03
	scenario5Invoice, scenario5Payment, scenario5User, scenario5FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100},
			InvoiceStatuses:     []string{invoiceIssued},
			PaymentStatuses:     []string{paymentPending},
			ExistingResultCodes: []string{""},
			ResultCodes:         []string{"03"},
			DepositAmount:       []int{100},
		},
	)

	// Data for scenario 6 which have category equal to 03 and payment has existing C-R1 result code
	scenario6Invoice, scenario6Payment, scenario6User, scenario6FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100},
			InvoiceStatuses:     []string{invoiceIssued},
			PaymentStatuses:     []string{paymentPending},
			ExistingResultCodes: []string{"C-R1"},
			ResultCodes:         []string{"03"},
			DepositAmount:       []int{100},
		},
	)

	// Data for scenario 7 which have category equal to 01 and payment has existing C-R2 result code
	scenario7Invoice, scenario7Payment, scenario7User, scenario7FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100},
			InvoiceStatuses:     []string{invoiceIssued},
			PaymentStatuses:     []string{paymentPending},
			ExistingResultCodes: []string{"C-R2"},
			ResultCodes:         []string{"01"},
			DepositAmount:       []int{100},
		},
	)

	// Data for scenario 8 which have category equal to 02 and payment has existing C-R2 or C-R1 result code
	scenario8Invoice, scenario8Payment, scenario8User, scenario8FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100, 100},
			InvoiceStatuses:     []string{invoiceIssued, invoiceIssued},
			PaymentStatuses:     []string{paymentPending, paymentPending},
			ExistingResultCodes: []string{"C-R2", "C-R1"},
			ResultCodes:         []string{"02", "02"},
			DepositAmount:       []int{100, 100},
		},
	)

	// Data for scenario 10 which have category equal to 01 - 03 and amount is not equal to invoice total amount
	scenario10Invoice, scenario10Payment, scenario10User, scenario10FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100, 100, 100},
			InvoiceStatuses:     []string{invoiceIssued, invoiceIssued, invoiceIssued},
			PaymentStatuses:     []string{paymentPending, paymentPending, paymentPending},
			ExistingResultCodes: []string{"", "", ""},
			ResultCodes:         []string{"01", "02", "03"},
			DepositAmount:       []int{50, 50, 50},
		},
	)

	// Data for scenario 14-15 which have category equal to 01 - 03 and invoice is not issued and payment status is failed
	scenario14Invoice, scenario14Payment, scenario14User, scenario14FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100, 100, 100},
			InvoiceStatuses:     []string{invoiceFailed, invoiceFailed, invoiceFailed},
			PaymentStatuses:     []string{paymentFailed, paymentFailed, paymentFailed},
			ExistingResultCodes: []string{"", "", ""},
			ResultCodes:         []string{"01", "02", "03"},
			DepositAmount:       []int{100, 100, 100},
		},
	)

	// Data for scenario 16 which have category equal to 01 - 03 and invoice is not issued and payment status is failed and amount is not equal to invoice total amount
	scenario16Invoice, scenario16Payment, scenario16User, scenario16FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100, 100, 100},
			InvoiceStatuses:     []string{invoiceFailed, invoiceFailed, invoiceFailed},
			PaymentStatuses:     []string{paymentFailed, paymentFailed, paymentFailed},
			ExistingResultCodes: []string{"", "", ""},
			ResultCodes:         []string{"01", "02", "03"},
			DepositAmount:       []int{50, 50, 50},
		},
	)

	type testCase struct {
		name                 string
		convenienceStoreFile *ConvenienceStoreFile
		ctx                  context.Context
		expectedResult       interface{}
		expectedErr          error
		setup                func(ctx context.Context)
	}

	for _, testCase := range []*testCase{
		{
			name:                 "Scenario 3 - Successful convenience store payment",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario3FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R0",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String(),
					},
				},
				SuccessfulPayments: 1,
				PendingPayments:    0,
				FailedPayments:     0,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario3Payment[0], nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario3Invoice[0], nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario3User[0], nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 4 - Convenience store payment with category 01",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario4FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R1",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					},
				},
				SuccessfulPayments: 0,
				PendingPayments:    1,
				FailedPayments:     0,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario4Payment[0], nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario4Invoice[0], nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario4User[0], nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 5 - Convenience store payment with category 03",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario5FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R2",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					},
				},
				SuccessfulPayments: 0,
				PendingPayments:    1,
				FailedPayments:     0,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario5Payment[0], nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario5Invoice[0], nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario5User[0], nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 6 - Convenience store payment with category 03 and payment has existing C-R1 result code",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario6FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R2",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					},
				},
				SuccessfulPayments: 0,
				PendingPayments:    1,
				FailedPayments:     0,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario6Payment[0], nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario6Invoice[0], nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario6User[0], nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 7 - Convenience store payment with category 01 and payment has existing C-R2 result code",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario7FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R1",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					},
				},
				SuccessfulPayments: 0,
				PendingPayments:    1,
				FailedPayments:     0,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario7Payment[0], nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario7Invoice[0], nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario7User[0], nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 8 - Convenience store payment with category 02 and payment has existing C-R2 or C-R1 result code",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario8FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R0",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String(),
					},
					{
						PaymentSequenceNumber: 2,
						InvoiceSequenceNumber: 2,
						ResultCode:            "C-R0",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String(),
					},
				},
				SuccessfulPayments: 2,
				PendingPayments:    0,
				FailedPayments:     0,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)

				for i := 0; i < 2; i++ {
					mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
					mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario8Payment[i], nil)
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario8Invoice[i], nil)
					mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario8User[i], nil)
					mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}

				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 10 - Convenience store payment with category 01 - 03 and amount is not equal to invoice total amount",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario10FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R1-1",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
					{
						PaymentSequenceNumber: 2,
						InvoiceSequenceNumber: 2,
						ResultCode:            "C-R0-1",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
					{
						PaymentSequenceNumber: 3,
						InvoiceSequenceNumber: 3,
						ResultCode:            "C-R2-1",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
				},
				SuccessfulPayments: 0,
				PendingPayments:    0,
				FailedPayments:     3,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)

				for i := 0; i < 3; i++ {
					mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
					mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario10Payment[i], nil)
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario10Invoice[i], nil)
					mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario10User[i], nil)
					mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}

				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 14-15 - Convenience store payment with category 01 - 03 and invoice status is not ISSUED and payment status is FAILED",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario14FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R1-2",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
					{
						PaymentSequenceNumber: 2,
						InvoiceSequenceNumber: 2,
						ResultCode:            "C-R0-2",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
					{
						PaymentSequenceNumber: 3,
						InvoiceSequenceNumber: 3,
						ResultCode:            "C-R2-2",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
				},
				SuccessfulPayments: 0,
				PendingPayments:    0,
				FailedPayments:     3,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)

				for i := 0; i < 3; i++ {
					mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
					mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario14Payment[i], nil)
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario14Invoice[i], nil)
					mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario14User[i], nil)
					mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}

				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 16 - Convenience store payment with category 01 - 03 and invoice status is not ISSUED and payment status is FAILED and amount is not equal to invoice total amount",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario16FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R1-3",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
					{
						PaymentSequenceNumber: 2,
						InvoiceSequenceNumber: 2,
						ResultCode:            "C-R0-3",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
					{
						PaymentSequenceNumber: 3,
						InvoiceSequenceNumber: 3,
						ResultCode:            "C-R2-3",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
				},
				SuccessfulPayments: 0,
				PendingPayments:    0,
				FailedPayments:     3,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)

				for i := 0; i < 3; i++ {
					mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
					mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario16Payment[i], nil)
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario16Invoice[i], nil)
					mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario16User[i], nil)
					mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}

				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			validationResult, err := convenienceStoreValidator.Validate(testCase.ctx, &PaymentFile{ConvenienceStoreFile: testCase.convenienceStoreFile})

			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)

				expectedValidationResult := testCase.expectedResult.(*PaymentValidationResult)

				assert.Equal(t, len(validationResult.ValidatedPayments), len(expectedValidationResult.ValidatedPayments))
				assert.Equal(t, validationResult.SuccessfulPayments, expectedValidationResult.SuccessfulPayments)
				assert.Equal(t, validationResult.PendingPayments, expectedValidationResult.PendingPayments)
				assert.Equal(t, validationResult.FailedPayments, expectedValidationResult.FailedPayments)

				for i, r := range expectedValidationResult.ValidatedPayments {
					assert.Equal(t, r.PaymentSequenceNumber, validationResult.ValidatedPayments[i].PaymentSequenceNumber)
					assert.Equal(t, r.ResultCode, validationResult.ValidatedPayments[i].ResultCode)
					assert.Equal(t, r.Amount, validationResult.ValidatedPayments[i].Amount)
					assert.Equal(t, r.PaymentMethod, validationResult.ValidatedPayments[i].PaymentMethod)
					assert.Equal(t, r.InvoiceSequenceNumber, validationResult.ValidatedPayments[i].InvoiceSequenceNumber)
					assert.Equal(t, r.PaymentStatus, validationResult.ValidatedPayments[i].PaymentStatus)
				}

			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockInvoiceRepo, mockPaymentRepo, mockInvoiceActionLogRepo, mockBulkValidationsRepo, mockBulkValidationsDetailRepo, mockUserBasicInfoRepo, mockTx)
		})
	}
}

func TestConvenienceStoreCSVPaymentFileValidator_EnableBulkValidate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockInvoiceActionLogRepo := new(mock_repositories.MockInvoiceActionLogRepo)
	mockBulkValidationsRepo := new(mock_repositories.MockBulkPaymentValidationsRepo)
	mockBulkValidationsDetailRepo := new(mock_repositories.MockBulkPaymentValidationsDetailRepo)
	mockUserBasicInfoRepo := new(mock_repositories.MockUserBasicInfoRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	baseValidator := &BasePaymentFileValidator{
		DB:                               mockDB,
		InvoiceRepo:                      mockInvoiceRepo,
		PaymentRepo:                      mockPaymentRepo,
		InvoiceActionLogRepo:             mockInvoiceActionLogRepo,
		BulkPaymentValidationsRepo:       mockBulkValidationsRepo,
		BulkPaymentValidationsDetailRepo: mockBulkValidationsDetailRepo,
		UserBasicInfoRepo:                mockUserBasicInfoRepo,
		UnleashClient:                    mockUnleashClient,
	}

	convenienceStoreValidator := &ConvenienceStoreCSVPaymentFileValidator{
		BasePaymentFileValidator: baseValidator,
	}

	invoiceIssued := invoice_pb.InvoiceStatus_ISSUED.String()
	paymentPending := invoice_pb.PaymentStatus_PAYMENT_PENDING.String()
	paymentFailed := invoice_pb.PaymentStatus_PAYMENT_FAILED.String()

	// Data for scenario 3 which have category equal to 02
	scenario3Invoice, scenario3Payment, scenario3User, scenario3FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100},
			InvoiceStatuses:     []string{invoiceIssued},
			PaymentStatuses:     []string{paymentPending},
			ExistingResultCodes: []string{""},
			ResultCodes:         []string{"02"},
			DepositAmount:       []int{100},
		},
	)

	// Data for scenario 4 which have category equal to 01
	scenario4Invoice, scenario4Payment, scenario4User, scenario4FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100},
			InvoiceStatuses:     []string{invoiceIssued},
			PaymentStatuses:     []string{paymentPending},
			ExistingResultCodes: []string{""},
			ResultCodes:         []string{"01"},
			DepositAmount:       []int{100},
		},
	)

	// Data for scenario 5 which have category equal to 03
	scenario5Invoice, scenario5Payment, scenario5User, scenario5FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100},
			InvoiceStatuses:     []string{invoiceIssued},
			PaymentStatuses:     []string{paymentPending},
			ExistingResultCodes: []string{""},
			ResultCodes:         []string{"03"},
			DepositAmount:       []int{100},
		},
	)

	// Data for scenario 6 which have category equal to 03 and payment has existing C-R1 result code
	scenario6Invoice, scenario6Payment, scenario6User, scenario6FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100},
			InvoiceStatuses:     []string{invoiceIssued},
			PaymentStatuses:     []string{paymentPending},
			ExistingResultCodes: []string{"C-R1"},
			ResultCodes:         []string{"03"},
			DepositAmount:       []int{100},
		},
	)

	// Data for scenario 7 which have category equal to 01 and payment has existing C-R2 result code
	scenario7Invoice, scenario7Payment, scenario7User, scenario7FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100},
			InvoiceStatuses:     []string{invoiceIssued},
			PaymentStatuses:     []string{paymentPending},
			ExistingResultCodes: []string{"C-R2"},
			ResultCodes:         []string{"01"},
			DepositAmount:       []int{100},
		},
	)

	// Data for scenario 8 which have category equal to 02 and payment has existing C-R2 or C-R1 result code
	scenario8Invoice, scenario8Payment, scenario8User, scenario8FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100, 100},
			InvoiceStatuses:     []string{invoiceIssued, invoiceIssued},
			PaymentStatuses:     []string{paymentPending, paymentPending},
			ExistingResultCodes: []string{"C-R2", "C-R1"},
			ResultCodes:         []string{"02", "02"},
			DepositAmount:       []int{100, 100},
		},
	)

	// Data for scenario 10 which have category equal to 01 - 03 and amount is not equal to invoice total amount
	scenario10Invoice, scenario10Payment, scenario10User, scenario10FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100, 100, 100},
			InvoiceStatuses:     []string{invoiceIssued, invoiceIssued, invoiceIssued},
			PaymentStatuses:     []string{paymentPending, paymentPending, paymentPending},
			ExistingResultCodes: []string{"", "", ""},
			ResultCodes:         []string{"01", "02", "03"},
			DepositAmount:       []int{50, 50, 50},
		},
	)

	// Data for scenario 14-15 which have category equal to 01 - 03 and invoice is not issued and payment status is failed
	scenario14Invoice, scenario14Payment, scenario14User, scenario14FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100, 100, 100},
			InvoiceStatuses:     []string{invoiceIssued, invoiceIssued, invoiceIssued},
			PaymentStatuses:     []string{paymentFailed, paymentFailed, paymentFailed},
			ExistingResultCodes: []string{"", "", ""},
			ResultCodes:         []string{"01", "02", "03"},
			DepositAmount:       []int{100, 100, 100},
		},
	)

	// Data for scenario 16 which have category equal to 01 - 03 and invoice is not issued and payment status is failed and amount is not equal to invoice total amount
	scenario16Invoice, scenario16Payment, scenario16User, scenario16FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100, 100, 100},
			InvoiceStatuses:     []string{invoiceIssued, invoiceIssued, invoiceIssued},
			PaymentStatuses:     []string{paymentFailed, paymentFailed, paymentFailed},
			ExistingResultCodes: []string{"", "", ""},
			ResultCodes:         []string{"01", "02", "03"},
			DepositAmount:       []int{50, 50, 50},
		},
	)

	type testCase struct {
		name                 string
		convenienceStoreFile *ConvenienceStoreFile
		ctx                  context.Context
		expectedResult       interface{}
		expectedErr          error
		setup                func(ctx context.Context)
	}

	for _, testCase := range []*testCase{
		{
			name:                 "Scenario 3 - Successful convenience store payment",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario3FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R0",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String(),
					},
				},
				SuccessfulPayments: 1,
				PendingPayments:    0,
				FailedPayments:     0,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(true, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario3Payment[0], nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario3Invoice[0], nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario3User[0], nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 4 - Convenience store payment with category 01",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario4FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R1",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					},
				},
				SuccessfulPayments: 0,
				PendingPayments:    1,
				FailedPayments:     0,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(true, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario4Payment[0], nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario4Invoice[0], nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario4User[0], nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 5 - Convenience store payment with category 03",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario5FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R2",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					},
				},
				SuccessfulPayments: 0,
				PendingPayments:    1,
				FailedPayments:     0,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(true, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario5Payment[0], nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario5Invoice[0], nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario5User[0], nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 6 - Convenience store payment with category 03 and payment has existing C-R1 result code",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario6FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R2",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					},
				},
				SuccessfulPayments: 0,
				PendingPayments:    1,
				FailedPayments:     0,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(true, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario6Payment[0], nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario6Invoice[0], nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario6User[0], nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 7 - Convenience store payment with category 01 and payment has existing C-R2 result code",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario7FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R1",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					},
				},
				SuccessfulPayments: 0,
				PendingPayments:    1,
				FailedPayments:     0,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(true, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario7Payment[0], nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario7Invoice[0], nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario7User[0], nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 8 - Convenience store payment with category 02 and payment has existing C-R2 or C-R1 result code",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario8FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R0",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String(),
					},
					{
						PaymentSequenceNumber: 2,
						InvoiceSequenceNumber: 2,
						ResultCode:            "C-R0",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String(),
					},
				},
				SuccessfulPayments: 2,
				PendingPayments:    0,
				FailedPayments:     0,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(true, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)

				for i := 0; i < 2; i++ {
					mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
					mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario8Payment[i], nil)
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario8Invoice[i], nil)
					mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario8User[i], nil)
					mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}

				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 10 - Convenience store payment with category 01 - 03 and amount is not equal to invoice total amount",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario10FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R1-1",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
					{
						PaymentSequenceNumber: 2,
						InvoiceSequenceNumber: 2,
						ResultCode:            "C-R0-1",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
					{
						PaymentSequenceNumber: 3,
						InvoiceSequenceNumber: 3,
						ResultCode:            "C-R2-1",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
				},
				SuccessfulPayments: 0,
				PendingPayments:    0,
				FailedPayments:     3,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(true, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)

				for i := 0; i < 3; i++ {
					mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
					mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario10Payment[i], nil)
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario10Invoice[i], nil)
					mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario10User[i], nil)
					mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}

				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 14-15 - Convenience store payment with category 01 - 03 and invoice status is not ISSUED and payment status is FAILED",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario14FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R1-2",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
					{
						PaymentSequenceNumber: 2,
						InvoiceSequenceNumber: 2,
						ResultCode:            "C-R0-2",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
					{
						PaymentSequenceNumber: 3,
						InvoiceSequenceNumber: 3,
						ResultCode:            "C-R2-2",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
				},
				SuccessfulPayments: 0,
				PendingPayments:    0,
				FailedPayments:     3,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)

				for i := 0; i < 3; i++ {
					mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
					mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario14Payment[i], nil)
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario14Invoice[i], nil)
					mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario14User[i], nil)
					mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}

				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 16 - Convenience store payment with category 01 - 03 and invoice status is not ISSUED and payment status is FAILED and amount is not equal to invoice total amount",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario16FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R1-3",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
					{
						PaymentSequenceNumber: 2,
						InvoiceSequenceNumber: 2,
						ResultCode:            "C-R0-3",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
					{
						PaymentSequenceNumber: 3,
						InvoiceSequenceNumber: 3,
						ResultCode:            "C-R2-3",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
				},
				SuccessfulPayments: 0,
				PendingPayments:    0,
				FailedPayments:     3,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)

				for i := 0; i < 3; i++ {
					mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
					mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario16Payment[i], nil)
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario16Invoice[i], nil)
					mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario16User[i], nil)
					mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}

				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			validationResult, err := convenienceStoreValidator.Validate(testCase.ctx, &PaymentFile{ConvenienceStoreFile: testCase.convenienceStoreFile})

			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)

				expectedValidationResult := testCase.expectedResult.(*PaymentValidationResult)

				assert.Equal(t, len(validationResult.ValidatedPayments), len(expectedValidationResult.ValidatedPayments))
				assert.Equal(t, validationResult.SuccessfulPayments, expectedValidationResult.SuccessfulPayments)
				assert.Equal(t, validationResult.PendingPayments, expectedValidationResult.PendingPayments)
				assert.Equal(t, validationResult.FailedPayments, expectedValidationResult.FailedPayments)

				for i, r := range expectedValidationResult.ValidatedPayments {
					assert.Equal(t, r.PaymentSequenceNumber, validationResult.ValidatedPayments[i].PaymentSequenceNumber)
					assert.Equal(t, r.ResultCode, validationResult.ValidatedPayments[i].ResultCode)
					assert.Equal(t, r.Amount, validationResult.ValidatedPayments[i].Amount)
					assert.Equal(t, r.PaymentMethod, validationResult.ValidatedPayments[i].PaymentMethod)
					assert.Equal(t, r.InvoiceSequenceNumber, validationResult.ValidatedPayments[i].InvoiceSequenceNumber)
					assert.Equal(t, r.PaymentStatus, validationResult.ValidatedPayments[i].PaymentStatus)
				}

			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockInvoiceRepo, mockPaymentRepo, mockInvoiceActionLogRepo, mockBulkValidationsRepo, mockBulkValidationsDetailRepo, mockUserBasicInfoRepo, mockTx)
		})
	}
}

func getPaymentInvoiceUserList(invoices []*entities.Invoice, payments []*entities.Payment, userBasicInfos []*entities.UserBasicInfo) []*entities.PaymentInvoiceUserMap {
	res := []*entities.PaymentInvoiceUserMap{}

	for i := 0; i < len(invoices); i++ {
		res = append(res, &entities.PaymentInvoiceUserMap{
			Invoice:       invoices[i],
			Payment:       payments[i],
			UserBasicInfo: userBasicInfos[i],
		})
	}

	return res
}

func TestConvenienceStoreCSVPaymentFileValidator_EnableImproveBulkPaymentValidation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)
	mockInvoiceRepo := new(mock_repositories.MockInvoiceRepo)
	mockPaymentRepo := new(mock_repositories.MockPaymentRepo)
	mockInvoiceActionLogRepo := new(mock_repositories.MockInvoiceActionLogRepo)
	mockBulkValidationsRepo := new(mock_repositories.MockBulkPaymentValidationsRepo)
	mockBulkValidationsDetailRepo := new(mock_repositories.MockBulkPaymentValidationsDetailRepo)
	mockUserBasicInfoRepo := new(mock_repositories.MockUserBasicInfoRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	baseValidator := &BasePaymentFileValidator{
		DB:                               mockDB,
		InvoiceRepo:                      mockInvoiceRepo,
		PaymentRepo:                      mockPaymentRepo,
		InvoiceActionLogRepo:             mockInvoiceActionLogRepo,
		BulkPaymentValidationsRepo:       mockBulkValidationsRepo,
		BulkPaymentValidationsDetailRepo: mockBulkValidationsDetailRepo,
		UserBasicInfoRepo:                mockUserBasicInfoRepo,
		UnleashClient:                    mockUnleashClient,
	}

	convenienceStoreValidator := &ConvenienceStoreCSVPaymentFileValidator{
		BasePaymentFileValidator: baseValidator,
	}

	invoiceIssued := invoice_pb.InvoiceStatus_ISSUED.String()
	paymentPending := invoice_pb.PaymentStatus_PAYMENT_PENDING.String()
	paymentFailed := invoice_pb.PaymentStatus_PAYMENT_FAILED.String()

	// Data for scenario 3 which have category equal to 02
	scenario3Invoice, scenario3Payment, scenario3User, scenario3FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100},
			InvoiceStatuses:     []string{invoiceIssued},
			PaymentStatuses:     []string{paymentPending},
			ExistingResultCodes: []string{""},
			ResultCodes:         []string{"02"},
			DepositAmount:       []int{100},
		},
	)

	mockScenario3InvoicePaymentUser := getPaymentInvoiceUserList(scenario3Invoice, scenario3Payment, scenario3User)

	// // Data for scenario 4 which have category equal to 01
	scenario4Invoice, scenario4Payment, scenario4User, scenario4FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100},
			InvoiceStatuses:     []string{invoiceIssued},
			PaymentStatuses:     []string{paymentPending},
			ExistingResultCodes: []string{""},
			ResultCodes:         []string{"01"},
			DepositAmount:       []int{100},
		},
	)

	mockScenario4InvoicePaymentUser := getPaymentInvoiceUserList(scenario4Invoice, scenario4Payment, scenario4User)

	// // Data for scenario 5 which have category equal to 03
	scenario5Invoice, scenario5Payment, scenario5User, scenario5FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100},
			InvoiceStatuses:     []string{invoiceIssued},
			PaymentStatuses:     []string{paymentPending},
			ExistingResultCodes: []string{""},
			ResultCodes:         []string{"03"},
			DepositAmount:       []int{100},
		},
	)

	mockScenario5InvoicePaymentUser := getPaymentInvoiceUserList(scenario5Invoice, scenario5Payment, scenario5User)

	// // Data for scenario 6 which have category equal to 03 and payment has existing C-R1 result code
	scenario6Invoice, scenario6Payment, scenario6User, scenario6FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100},
			InvoiceStatuses:     []string{invoiceIssued},
			PaymentStatuses:     []string{paymentPending},
			ExistingResultCodes: []string{"C-R1"},
			ResultCodes:         []string{"03"},
			DepositAmount:       []int{100},
		},
	)

	mockScenario6InvoicePaymentUser := getPaymentInvoiceUserList(scenario6Invoice, scenario6Payment, scenario6User)

	// // Data for scenario 7 which have category equal to 01 and payment has existing C-R2 result code
	scenario7Invoice, scenario7Payment, scenario7User, scenario7FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100},
			InvoiceStatuses:     []string{invoiceIssued},
			PaymentStatuses:     []string{paymentPending},
			ExistingResultCodes: []string{"C-R2"},
			ResultCodes:         []string{"01"},
			DepositAmount:       []int{100},
		},
	)

	mockScenario7InvoicePaymentUser := getPaymentInvoiceUserList(scenario7Invoice, scenario7Payment, scenario7User)

	// Data for scenario 8 which have category equal to 02 and payment has existing C-R2 or C-R1 result code
	scenario8Invoice, scenario8Payment, scenario8User, scenario8FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100, 100},
			InvoiceStatuses:     []string{invoiceIssued, invoiceIssued},
			PaymentStatuses:     []string{paymentPending, paymentPending},
			ExistingResultCodes: []string{"C-R2", "C-R1"},
			ResultCodes:         []string{"02", "02"},
			DepositAmount:       []int{100, 100},
		},
	)

	mockScenario8InvoicePaymentUser := getPaymentInvoiceUserList(scenario8Invoice, scenario8Payment, scenario8User)

	// Data for scenario 10 which have category equal to 01 - 03 and amount is not equal to invoice total amount
	scenario10Invoice, scenario10Payment, scenario10User, scenario10FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100, 100, 100},
			InvoiceStatuses:     []string{invoiceIssued, invoiceIssued, invoiceIssued},
			PaymentStatuses:     []string{paymentPending, paymentPending, paymentPending},
			ExistingResultCodes: []string{"", "", ""},
			ResultCodes:         []string{"01", "02", "03"},
			DepositAmount:       []int{50, 50, 50},
		},
	)

	mockScenario10InvoicePaymentUser := getPaymentInvoiceUserList(scenario10Invoice, scenario10Payment, scenario10User)

	// Data for scenario 14-15 which have category equal to 01 - 03 and invoice is not issued and payment status is failed
	scenario14Invoice, scenario14Payment, scenario14User, scenario14FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100, 100, 100},
			InvoiceStatuses:     []string{invoiceIssued, invoiceIssued, invoiceIssued},
			PaymentStatuses:     []string{paymentFailed, paymentFailed, paymentFailed},
			ExistingResultCodes: []string{"", "", ""},
			ResultCodes:         []string{"01", "02", "03"},
			DepositAmount:       []int{100, 100, 100},
		},
	)

	mockScenario14InvoicePaymentUser := getPaymentInvoiceUserList(scenario14Invoice, scenario14Payment, scenario14User)

	// Data for scenario 16 which have category equal to 01 - 03 and invoice is not issued and payment status is failed and amount is not equal to invoice total amount
	scenario16Invoice, scenario16Payment, scenario16User, scenario16FileContent := genTestConvenienceStoreValidatorData(
		genTestConvenienceStoreDataInput{
			Total:               []int{100, 100, 100},
			InvoiceStatuses:     []string{invoiceIssued, invoiceIssued, invoiceIssued},
			PaymentStatuses:     []string{paymentFailed, paymentFailed, paymentFailed},
			ExistingResultCodes: []string{"", "", ""},
			ResultCodes:         []string{"01", "02", "03"},
			DepositAmount:       []int{50, 50, 50},
		},
	)

	mockScenario16InvoicePaymentUser := getPaymentInvoiceUserList(scenario16Invoice, scenario16Payment, scenario16User)

	type testCase struct {
		name                 string
		convenienceStoreFile *ConvenienceStoreFile
		ctx                  context.Context
		expectedResult       interface{}
		expectedErr          error
		setup                func(ctx context.Context)
	}

	for _, testCase := range []*testCase{
		{
			name:                 "Scenario 3 - Successful convenience store payment",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario3FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R0",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String(),
					},
				},
				SuccessfulPayments: 1,
				PendingPayments:    0,
				FailedPayments:     0,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(true, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockPaymentRepo.On("InsertPaymentNumbersTempTable", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockPaymentRepo.On("FindPaymentInvoiceUserFromTempTable", ctx, mockTx).Times(1).Return(mockScenario3InvoicePaymentUser, nil)
				mockPaymentRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsDetailRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockInvoiceActionLogRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 4 - Convenience store payment with category 01",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario4FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R1",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					},
				},
				SuccessfulPayments: 0,
				PendingPayments:    1,
				FailedPayments:     0,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(true, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockPaymentRepo.On("InsertPaymentNumbersTempTable", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockPaymentRepo.On("FindPaymentInvoiceUserFromTempTable", ctx, mockTx).Times(1).Return(mockScenario4InvoicePaymentUser, nil)
				mockPaymentRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsDetailRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockInvoiceActionLogRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 5 - Convenience store payment with category 03",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario5FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R2",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					},
				},
				SuccessfulPayments: 0,
				PendingPayments:    1,
				FailedPayments:     0,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(true, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockPaymentRepo.On("InsertPaymentNumbersTempTable", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockPaymentRepo.On("FindPaymentInvoiceUserFromTempTable", ctx, mockTx).Times(1).Return(mockScenario5InvoicePaymentUser, nil)
				mockPaymentRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsDetailRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockInvoiceActionLogRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 6 - Convenience store payment with category 03 and payment has existing C-R1 result code",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario6FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R2",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					},
				},
				SuccessfulPayments: 0,
				PendingPayments:    1,
				FailedPayments:     0,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(true, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockPaymentRepo.On("InsertPaymentNumbersTempTable", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockPaymentRepo.On("FindPaymentInvoiceUserFromTempTable", ctx, mockTx).Times(1).Return(mockScenario6InvoicePaymentUser, nil)
				mockPaymentRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsDetailRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockInvoiceActionLogRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 7 - Convenience store payment with category 01 and payment has existing C-R2 result code",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario7FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R1",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_PENDING.String(),
					},
				},
				SuccessfulPayments: 0,
				PendingPayments:    1,
				FailedPayments:     0,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(true, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockPaymentRepo.On("InsertPaymentNumbersTempTable", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockPaymentRepo.On("FindPaymentInvoiceUserFromTempTable", ctx, mockTx).Times(1).Return(mockScenario7InvoicePaymentUser, nil)
				mockPaymentRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsDetailRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockInvoiceActionLogRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 8 - Convenience store payment with category 02 and payment has existing C-R2 or C-R1 result code",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario8FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R0",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String(),
					},
					{
						PaymentSequenceNumber: 2,
						InvoiceSequenceNumber: 2,
						ResultCode:            "C-R0",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String(),
					},
				},
				SuccessfulPayments: 2,
				PendingPayments:    0,
				FailedPayments:     0,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(true, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockPaymentRepo.On("InsertPaymentNumbersTempTable", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockPaymentRepo.On("FindPaymentInvoiceUserFromTempTable", ctx, mockTx).Times(1).Return(mockScenario8InvoicePaymentUser, nil)
				mockPaymentRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsDetailRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockInvoiceActionLogRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 10 - Convenience store payment with category 01 - 03 and amount is not equal to invoice total amount",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario10FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R1-1",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
					{
						PaymentSequenceNumber: 2,
						InvoiceSequenceNumber: 2,
						ResultCode:            "C-R0-1",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
					{
						PaymentSequenceNumber: 3,
						InvoiceSequenceNumber: 3,
						ResultCode:            "C-R2-1",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
				},
				SuccessfulPayments: 0,
				PendingPayments:    0,
				FailedPayments:     3,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(true, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockPaymentRepo.On("InsertPaymentNumbersTempTable", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockPaymentRepo.On("FindPaymentInvoiceUserFromTempTable", ctx, mockTx).Times(1).Return(mockScenario10InvoicePaymentUser, nil)
				mockPaymentRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsDetailRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockInvoiceActionLogRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 14-15 - Convenience store payment with category 01 - 03 and invoice status is not ISSUED and payment status is FAILED",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario14FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R1-2",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
					{
						PaymentSequenceNumber: 2,
						InvoiceSequenceNumber: 2,
						ResultCode:            "C-R0-2",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
					{
						PaymentSequenceNumber: 3,
						InvoiceSequenceNumber: 3,
						ResultCode:            "C-R2-2",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
				},
				SuccessfulPayments: 0,
				PendingPayments:    0,
				FailedPayments:     3,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(true, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockPaymentRepo.On("InsertPaymentNumbersTempTable", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockPaymentRepo.On("FindPaymentInvoiceUserFromTempTable", ctx, mockTx).Times(1).Return(mockScenario14InvoicePaymentUser, nil)
				mockPaymentRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsDetailRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockInvoiceActionLogRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:                 "Scenario 16 - Convenience store payment with category 01 - 03 and invoice status is not ISSUED and payment status is FAILED and amount is not equal to invoice total amount",
			convenienceStoreFile: genTestConvenienceStoreFile(scenario16FileContent),
			ctx:                  ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "C-R1-3",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
					{
						PaymentSequenceNumber: 2,
						InvoiceSequenceNumber: 2,
						ResultCode:            "C-R0-3",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
					{
						PaymentSequenceNumber: 3,
						InvoiceSequenceNumber: 3,
						ResultCode:            "C-R2-3",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_CONVENIENCE_STORE,
						PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
					},
				},
				SuccessfulPayments: 0,
				PendingPayments:    0,
				FailedPayments:     3,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(true, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockPaymentRepo.On("InsertPaymentNumbersTempTable", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockPaymentRepo.On("FindPaymentInvoiceUserFromTempTable", ctx, mockTx).Times(1).Return(mockScenario16InvoicePaymentUser, nil)
				mockPaymentRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsDetailRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockInvoiceActionLogRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			validationResult, err := convenienceStoreValidator.Validate(testCase.ctx, &PaymentFile{ConvenienceStoreFile: testCase.convenienceStoreFile})

			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)

				expectedValidationResult := testCase.expectedResult.(*PaymentValidationResult)

				assert.Equal(t, len(validationResult.ValidatedPayments), len(expectedValidationResult.ValidatedPayments))
				assert.Equal(t, validationResult.SuccessfulPayments, expectedValidationResult.SuccessfulPayments)
				assert.Equal(t, validationResult.PendingPayments, expectedValidationResult.PendingPayments)
				assert.Equal(t, validationResult.FailedPayments, expectedValidationResult.FailedPayments)

				for i, r := range expectedValidationResult.ValidatedPayments {
					assert.Equal(t, r.PaymentSequenceNumber, validationResult.ValidatedPayments[i].PaymentSequenceNumber)
					assert.Equal(t, r.ResultCode, validationResult.ValidatedPayments[i].ResultCode)
					assert.Equal(t, r.Amount, validationResult.ValidatedPayments[i].Amount)
					assert.Equal(t, r.PaymentMethod, validationResult.ValidatedPayments[i].PaymentMethod)
					assert.Equal(t, r.InvoiceSequenceNumber, validationResult.ValidatedPayments[i].InvoiceSequenceNumber)
					assert.Equal(t, r.PaymentStatus, validationResult.ValidatedPayments[i].PaymentStatus)
				}

			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockInvoiceRepo, mockPaymentRepo, mockInvoiceActionLogRepo, mockBulkValidationsRepo, mockBulkValidationsDetailRepo, mockUserBasicInfoRepo, mockTx)
		})
	}
}

type genTestConvenienceStoreDataInput struct {
	Total               []int
	InvoiceStatuses     []string
	PaymentStatuses     []string
	ExistingResultCodes []string
	ResultCodes         []string
	DepositAmount       []int
}

func genTestConvenienceStoreValidatorData(input genTestConvenienceStoreDataInput) (
	invoices []*entities.Invoice,
	payments []*entities.Payment,
	userBasicInfos []*entities.UserBasicInfo,
	fileContent []*testConvenienceStoreFileContent,
) {
	for i := 0; i < len(input.Total); i++ {
		invoices = append(invoices, &entities.Invoice{
			InvoiceID:             database.Text(fmt.Sprintf("%v", i+1)),
			InvoiceSequenceNumber: database.Int4(int32(i + 1)),
			Total:                 database.Numeric(float32(input.Total[i])),
			Status:                database.Text(input.InvoiceStatuses[i]),
			StudentID:             database.Text(fmt.Sprintf("%v", i+1)),
		})

		payments = append(payments, &entities.Payment{
			PaymentID:             database.Text(fmt.Sprintf("%v", i+1)),
			PaymentSequenceNumber: database.Int4(int32(i + 1)),
			PaymentStatus:         database.Text(input.PaymentStatuses[i]),
			ResultCode:            database.Text(input.ExistingResultCodes[i]),
			PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
		})

		userBasicInfos = append(userBasicInfos, &entities.UserBasicInfo{
			UserID: database.Text("1"),
			Name:   database.Text(fmt.Sprintf("Name-%v", i+1)),
		})

		fileContent = append(fileContent, &testConvenienceStoreFileContent{
			Amount:       input.DepositAmount[i],
			CodeForUser2: fmt.Sprintf("%v", i+1),
			Category:     input.ResultCodes[i],
			CreatedDate:  20060102,
		})
	}
	return invoices, payments, userBasicInfos, fileContent
}

func genTestConvenienceStoreFile(contents []*testConvenienceStoreFileContent) *ConvenienceStoreFile {
	data := []*ConvenienceStoreFileDataRecord{}
	for _, content := range contents {
		data = append(data, &ConvenienceStoreFileDataRecord{
			Category:        content.Category,
			CodeForUser2:    content.CodeForUser2,
			TransferredDate: 20060102,
			Amount:          content.Amount,
			CreatedDate:     content.CreatedDate,
			DateOfReceipt:   20060103,
		})
	}

	return &ConvenienceStoreFile{
		DataRecord: data,
	}
}
