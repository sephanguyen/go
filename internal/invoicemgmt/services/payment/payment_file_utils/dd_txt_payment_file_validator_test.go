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

type testDirectDebitFileContent struct {
	DepositAmount  int
	CustomerNumber string
	ResultCode     string
}

func TestDirectDebitTextPaymentFileValidator_Validate(t *testing.T) {
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

	directDebitValidator := &DirectDebitTextPaymentFileValidator{
		BasePaymentFileValidator: baseValidator,
		PaymentDate:              time.Now().UTC(),
	}

	invoiceIssued := invoice_pb.InvoiceStatus_ISSUED.String()
	invoiceVoid := invoice_pb.InvoiceStatus_VOID.String()
	paymentPending := invoice_pb.PaymentStatus_PAYMENT_PENDING.String()
	paymentFailed := invoice_pb.PaymentStatus_PAYMENT_FAILED.String()

	// Data for scenario 1 which have result code equal to 0
	scenario1Invoice, scenario1Payment, scenario1User, scenario1FileContent := genTestDirectDebitValidatorData(
		genTestDirectDebitDataInput{
			Total:           []int{100},
			InvoiceStatuses: []string{invoiceIssued},
			PaymentStatuses: []string{paymentPending},
			ResultCodes:     []string{"0"},
			DepositAmount:   []int{100},
		},
	)

	// Data for scenario 2 which the result code is not 0
	scenario2Invoice, scenario2Payment, scenario2User, scenario2FileContent := genTestDirectDebitValidatorData(
		genTestDirectDebitDataInput{
			Total:           []int{100, 100, 100, 100, 100, 100},
			InvoiceStatuses: []string{invoiceIssued, invoiceIssued, invoiceIssued, invoiceIssued, invoiceIssued, invoiceIssued},
			PaymentStatuses: []string{paymentPending, paymentPending, paymentPending, paymentPending, paymentPending, paymentPending},
			ResultCodes:     []string{"1", "2", "3", "4", "8", "9"},
			DepositAmount:   []int{100, 100, 100, 100, 100, 100},
		},
	)
	scenario2ExpectedValidatedPayment := []*ValidatedPayment{}
	for i := 0; i < 6; i++ {
		scenario2ExpectedValidatedPayment = append(scenario2ExpectedValidatedPayment, &ValidatedPayment{
			PaymentSequenceNumber: int32(i + 1),
			InvoiceSequenceNumber: int32(i + 1),
			ResultCode:            fmt.Sprintf("D-R%v", scenario2FileContent[i].ResultCode),
			Amount:                100,
			PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT,
			PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
		})
	}

	// Data for scenario 9 which the deposit amount is not equal to invoice total amount
	scenario9Invoice, scenario9Payment, scenario9User, scenario9FileContent := genTestDirectDebitValidatorData(
		genTestDirectDebitDataInput{
			Total:           []int{100, 100, 100, 100, 100, 100, 100},
			InvoiceStatuses: []string{invoiceIssued, invoiceIssued, invoiceIssued, invoiceIssued, invoiceIssued, invoiceIssued, invoiceIssued},
			PaymentStatuses: []string{paymentPending, paymentPending, paymentPending, paymentPending, paymentPending, paymentPending, paymentPending},
			ResultCodes:     []string{"0", "1", "2", "3", "4", "8", "9"},
			DepositAmount:   []int{50, 50, 50, 50, 50, 50, 50},
		},
	)
	scenario9ExpectedValidatedPayment := []*ValidatedPayment{}
	for i := 0; i < 7; i++ {
		scenario9ExpectedValidatedPayment = append(scenario9ExpectedValidatedPayment, &ValidatedPayment{
			PaymentSequenceNumber: int32(i + 1),
			InvoiceSequenceNumber: int32(i + 1),
			ResultCode:            fmt.Sprintf("D-R%v-1", scenario9FileContent[i].ResultCode),
			Amount:                100,
			PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT,
			PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
		})
	}

	// Data for scenario 11-12 where the invoice status is VOID and payment status is FAILED
	scenario11Invoice, scenario11Payment, scenario11User, scenario11FileContent := genTestDirectDebitValidatorData(
		genTestDirectDebitDataInput{
			Total:           []int{100, 100, 100, 100, 100, 100, 100},
			InvoiceStatuses: []string{invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid},
			PaymentStatuses: []string{paymentFailed, paymentFailed, paymentFailed, paymentFailed, paymentFailed, paymentFailed, paymentFailed},
			ResultCodes:     []string{"0", "1", "2", "3", "4", "8", "9"},
			DepositAmount:   []int{100, 100, 100, 100, 100, 100, 100},
		},
	)
	scenario11ExpectedValidatedPayment := []*ValidatedPayment{}
	for i := 0; i < 7; i++ {
		scenario11ExpectedValidatedPayment = append(scenario11ExpectedValidatedPayment, &ValidatedPayment{
			PaymentSequenceNumber: int32(i + 1),
			InvoiceSequenceNumber: int32(i + 1),
			ResultCode:            fmt.Sprintf("D-R%v-2", scenario9FileContent[i].ResultCode),
			Amount:                100,
			PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT,
			PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
		})
	}

	// Data for scenario 13 where the invoice status is VOID and payment status is FAILED and the deposit amount is not the same with invoice total amount
	scenario13Invoice, scenario13Payment, scenario13User, scenario13FileContent := genTestDirectDebitValidatorData(
		genTestDirectDebitDataInput{
			Total:           []int{100, 100, 100, 100, 100, 100, 100},
			InvoiceStatuses: []string{invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid},
			PaymentStatuses: []string{paymentFailed, paymentFailed, paymentFailed, paymentFailed, paymentFailed, paymentFailed, paymentFailed},
			ResultCodes:     []string{"0", "1", "2", "3", "4", "8", "9"},
			DepositAmount:   []int{50, 50, 50, 50, 50, 50, 50},
		},
	)
	scenario13ExpectedValidatedPayment := []*ValidatedPayment{}
	for i := 0; i < 7; i++ {
		scenario13ExpectedValidatedPayment = append(scenario13ExpectedValidatedPayment, &ValidatedPayment{
			PaymentSequenceNumber: int32(i + 1),
			InvoiceSequenceNumber: int32(i + 1),
			ResultCode:            fmt.Sprintf("D-R%v-3", scenario9FileContent[i].ResultCode),
			Amount:                100,
			PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT,
			PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
		})
	}

	type testCase struct {
		name            string
		directDebitFile *DirectDebitFile
		ctx             context.Context
		expectedResult  interface{}
		expectedErr     error
		setup           func(ctx context.Context)
	}

	for _, testCase := range []*testCase{
		{
			name:            "Scenario 1 - Successful direct debit payment",
			directDebitFile: genTestDirectDebitFile(scenario1FileContent),
			ctx:             ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "D-R0",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT,
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
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario1Payment[0], nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario1Invoice[0], nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario1User[0], nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:            "Scenario 2 - Failed direct debit result code",
			directDebitFile: genTestDirectDebitFile(scenario2FileContent),
			ctx:             ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments:  scenario2ExpectedValidatedPayment,
				SuccessfulPayments: 0,
				PendingPayments:    0,
				FailedPayments:     6,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)

				for i := 0; i < 6; i++ {
					mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
					mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario2Payment[i], nil)
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario2Invoice[i], nil)
					mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario2User[i], nil)
					mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}

				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)

			},
		},
		{
			name:            "Scenario 9 - Invoice total amount is not equal to deposit amount",
			directDebitFile: genTestDirectDebitFile(scenario9FileContent),
			ctx:             ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments:  scenario9ExpectedValidatedPayment,
				SuccessfulPayments: 0,
				PendingPayments:    0,
				FailedPayments:     7,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)

				for i := 0; i < 7; i++ {
					mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
					mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario9Payment[i], nil)
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario9Invoice[i], nil)
					mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario9User[i], nil)
					mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}

				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)

			},
		},
		{
			name:            "Scenario 11-12 - Invoice status is VOID and payment status is FAILED",
			directDebitFile: genTestDirectDebitFile(scenario11FileContent),
			ctx:             ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments:  scenario11ExpectedValidatedPayment,
				SuccessfulPayments: 0,
				PendingPayments:    0,
				FailedPayments:     7,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)

				for i := 0; i < 7; i++ {
					mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
					mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario11Payment[i], nil)
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario11Invoice[i], nil)
					mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario11User[i], nil)
					mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}

				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)

			},
		},
		{
			name:            "Scenario 13 - Invoice status is VOID and payment status is FAILED and deposit amount not equal to invoice total amount",
			directDebitFile: genTestDirectDebitFile(scenario13FileContent),
			ctx:             ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments:  scenario13ExpectedValidatedPayment,
				SuccessfulPayments: 0,
				PendingPayments:    0,
				FailedPayments:     7,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)

				for i := 0; i < 7; i++ {
					mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
					mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario13Payment[i], nil)
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario13Invoice[i], nil)
					mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario13User[i], nil)
					mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}

				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)

			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			validationResult, err := directDebitValidator.Validate(testCase.ctx, &PaymentFile{DirectDebitFile: testCase.directDebitFile})

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

type genTestDirectDebitDataInput struct {
	Total           []int
	InvoiceStatuses []string
	PaymentStatuses []string
	ResultCodes     []string
	DepositAmount   []int
}

func genTestDirectDebitValidatorData(input genTestDirectDebitDataInput) (
	invoices []*entities.Invoice,
	payments []*entities.Payment,
	userBasicInfos []*entities.UserBasicInfo,
	fileContent []*testDirectDebitFileContent,
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
			PaymentMethod:         database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
		})

		userBasicInfos = append(userBasicInfos, &entities.UserBasicInfo{
			UserID: database.Text("1"),
			Name:   database.Text(fmt.Sprintf("Name-%v", i+1)),
		})

		fileContent = append(fileContent, &testDirectDebitFileContent{
			DepositAmount:  input.DepositAmount[i],
			CustomerNumber: fmt.Sprintf("%v", i+1),
			ResultCode:     input.ResultCodes[i],
		})
	}
	return invoices, payments, userBasicInfos, fileContent
}

func genTestDirectDebitFile(contents []*testDirectDebitFileContent) *DirectDebitFile {
	data := []*DirectDebitFileDataRecord{}
	for _, content := range contents {
		data = append(data, &DirectDebitFileDataRecord{
			DataCategory:            2,
			DepositBankNumber:       12345,
			DepositBankName:         "test-deposit-bank-name",
			DepositBankBranchNumber: 12345,
			DepositBankBranchName:   "test-deposit-bank-branch-name",
			DepositItems:            1,
			AccountNumber:           "1234567",
			AccountOwnerName:        "test-account-owner-name",
			DepositAmount:           content.DepositAmount,
			NewCustomerCode:         "1",
			CustomerNumber:          content.CustomerNumber,
			ResultCode:              content.ResultCode,
		})
	}

	return &DirectDebitFile{
		Header: &DirectDebitFileHeaderRecord{
			DataCategory:     1,
			TypeCode:         91,
			CodeCategory:     0,
			ConsignorCode:    4819,
			ConsignorName:    "test-consignor-name",
			DepositDate:      20060102,
			BankNumber:       12345,
			BankName:         "test-bank-name",
			BankBranchNumber: 12345,
			BankBranchName:   "test-bank-branch-name",
			DepositItems:     1,
			AccountNumber:    "1234567",
		},
		Data: data,
		Trailer: &DirectDebitFileTrailerRecord{
			DataCategory:      8,
			TotalTransactions: len(data),
		},
		End: &DirectDebitFileEndRecord{
			DataCategory: 9,
		},
	}
}

func TestDirectDebitTextPaymentFileValidator_EnableBulkValidate(t *testing.T) {
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

	directDebitValidator := &DirectDebitTextPaymentFileValidator{
		BasePaymentFileValidator: baseValidator,
		PaymentDate:              time.Now().UTC(),
	}

	invoiceIssued := invoice_pb.InvoiceStatus_ISSUED.String()
	invoiceVoid := invoice_pb.InvoiceStatus_VOID.String()
	paymentPending := invoice_pb.PaymentStatus_PAYMENT_PENDING.String()
	paymentFailed := invoice_pb.PaymentStatus_PAYMENT_FAILED.String()

	// Data for scenario 1 which have result code equal to 0
	scenario1Invoice, scenario1Payment, scenario1User, scenario1FileContent := genTestDirectDebitValidatorData(
		genTestDirectDebitDataInput{
			Total:           []int{100},
			InvoiceStatuses: []string{invoiceIssued},
			PaymentStatuses: []string{paymentPending},
			ResultCodes:     []string{"0"},
			DepositAmount:   []int{100},
		},
	)

	// Data for scenario 2 which the result code is not 0
	scenario2Invoice, scenario2Payment, scenario2User, scenario2FileContent := genTestDirectDebitValidatorData(
		genTestDirectDebitDataInput{
			Total:           []int{100, 100, 100, 100, 100, 100},
			InvoiceStatuses: []string{invoiceIssued, invoiceIssued, invoiceIssued, invoiceIssued, invoiceIssued, invoiceIssued},
			PaymentStatuses: []string{paymentPending, paymentPending, paymentPending, paymentPending, paymentPending, paymentPending},
			ResultCodes:     []string{"1", "2", "3", "4", "8", "9"},
			DepositAmount:   []int{100, 100, 100, 100, 100, 100},
		},
	)
	scenario2ExpectedValidatedPayment := []*ValidatedPayment{}
	for i := 0; i < 6; i++ {
		scenario2ExpectedValidatedPayment = append(scenario2ExpectedValidatedPayment, &ValidatedPayment{
			PaymentSequenceNumber: int32(i + 1),
			InvoiceSequenceNumber: int32(i + 1),
			ResultCode:            fmt.Sprintf("D-R%v", scenario2FileContent[i].ResultCode),
			Amount:                100,
			PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT,
			PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
		})
	}

	// Data for scenario 9 which the deposit amount is not equal to invoice total amount
	scenario9Invoice, scenario9Payment, scenario9User, scenario9FileContent := genTestDirectDebitValidatorData(
		genTestDirectDebitDataInput{
			Total:           []int{100, 100, 100, 100, 100, 100, 100},
			InvoiceStatuses: []string{invoiceIssued, invoiceIssued, invoiceIssued, invoiceIssued, invoiceIssued, invoiceIssued, invoiceIssued},
			PaymentStatuses: []string{paymentPending, paymentPending, paymentPending, paymentPending, paymentPending, paymentPending, paymentPending},
			ResultCodes:     []string{"0", "1", "2", "3", "4", "8", "9"},
			DepositAmount:   []int{50, 50, 50, 50, 50, 50, 50},
		},
	)
	scenario9ExpectedValidatedPayment := []*ValidatedPayment{}
	for i := 0; i < 7; i++ {
		scenario9ExpectedValidatedPayment = append(scenario9ExpectedValidatedPayment, &ValidatedPayment{
			PaymentSequenceNumber: int32(i + 1),
			InvoiceSequenceNumber: int32(i + 1),
			ResultCode:            fmt.Sprintf("D-R%v-1", scenario9FileContent[i].ResultCode),
			Amount:                100,
			PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT,
			PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
		})
	}

	// Data for scenario 11-12 where the invoice status is VOID and payment status is FAILED
	scenario11Invoice, scenario11Payment, scenario11User, scenario11FileContent := genTestDirectDebitValidatorData(
		genTestDirectDebitDataInput{
			Total:           []int{100, 100, 100, 100, 100, 100, 100},
			InvoiceStatuses: []string{invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid},
			PaymentStatuses: []string{paymentFailed, paymentFailed, paymentFailed, paymentFailed, paymentFailed, paymentFailed, paymentFailed},
			ResultCodes:     []string{"0", "1", "2", "3", "4", "8", "9"},
			DepositAmount:   []int{100, 100, 100, 100, 100, 100, 100},
		},
	)
	scenario11ExpectedValidatedPayment := []*ValidatedPayment{}
	for i := 0; i < 7; i++ {
		scenario11ExpectedValidatedPayment = append(scenario11ExpectedValidatedPayment, &ValidatedPayment{
			PaymentSequenceNumber: int32(i + 1),
			InvoiceSequenceNumber: int32(i + 1),
			ResultCode:            fmt.Sprintf("D-R%v-2", scenario9FileContent[i].ResultCode),
			Amount:                100,
			PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT,
			PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
		})
	}

	// Data for scenario 13 where the invoice status is VOID and payment status is FAILED and the deposit amount is not the same with invoice total amount
	scenario13Invoice, scenario13Payment, scenario13User, scenario13FileContent := genTestDirectDebitValidatorData(
		genTestDirectDebitDataInput{
			Total:           []int{100, 100, 100, 100, 100, 100, 100},
			InvoiceStatuses: []string{invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid},
			PaymentStatuses: []string{paymentFailed, paymentFailed, paymentFailed, paymentFailed, paymentFailed, paymentFailed, paymentFailed},
			ResultCodes:     []string{"0", "1", "2", "3", "4", "8", "9"},
			DepositAmount:   []int{50, 50, 50, 50, 50, 50, 50},
		},
	)
	scenario13ExpectedValidatedPayment := []*ValidatedPayment{}
	for i := 0; i < 7; i++ {
		scenario13ExpectedValidatedPayment = append(scenario13ExpectedValidatedPayment, &ValidatedPayment{
			PaymentSequenceNumber: int32(i + 1),
			InvoiceSequenceNumber: int32(i + 1),
			ResultCode:            fmt.Sprintf("D-R%v-3", scenario9FileContent[i].ResultCode),
			Amount:                100,
			PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT,
			PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
		})
	}

	type testCase struct {
		name            string
		directDebitFile *DirectDebitFile
		ctx             context.Context
		expectedResult  interface{}
		expectedErr     error
		setup           func(ctx context.Context)
	}

	for _, testCase := range []*testCase{
		{
			name:            "Scenario 1 - Successful direct debit payment",
			directDebitFile: genTestDirectDebitFile(scenario1FileContent),
			ctx:             ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "D-R0",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT,
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
				mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario1Payment[0], nil)
				mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario1Invoice[0], nil)
				mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario1User[0], nil)
				mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:            "Scenario 2 - Failed direct debit result code",
			directDebitFile: genTestDirectDebitFile(scenario2FileContent),
			ctx:             ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments:  scenario2ExpectedValidatedPayment,
				SuccessfulPayments: 0,
				PendingPayments:    0,
				FailedPayments:     6,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(true, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)

				for i := 0; i < 6; i++ {
					mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
					mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario2Payment[i], nil)
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario2Invoice[i], nil)
					mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario2User[i], nil)
					mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}

				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)

			},
		},
		{
			name:            "Scenario 9 - Invoice total amount is not equal to deposit amount",
			directDebitFile: genTestDirectDebitFile(scenario9FileContent),
			ctx:             ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments:  scenario9ExpectedValidatedPayment,
				SuccessfulPayments: 0,
				PendingPayments:    0,
				FailedPayments:     7,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(true, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)

				for i := 0; i < 7; i++ {
					mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
					mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario9Payment[i], nil)
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario9Invoice[i], nil)
					mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario9User[i], nil)
					mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}

				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)

			},
		},
		{
			name:            "Scenario 11-12 - Invoice status is VOID and payment status is FAILED",
			directDebitFile: genTestDirectDebitFile(scenario11FileContent),
			ctx:             ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments:  scenario11ExpectedValidatedPayment,
				SuccessfulPayments: 0,
				PendingPayments:    0,
				FailedPayments:     7,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(true, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)

				for i := 0; i < 7; i++ {
					mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
					mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario11Payment[i], nil)
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario11Invoice[i], nil)
					mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario11User[i], nil)
					mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}

				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)

			},
		},
		{
			name:            "Scenario 13 - Invoice status is VOID and payment status is FAILED and deposit amount not equal to invoice total amount",
			directDebitFile: genTestDirectDebitFile(scenario13FileContent),
			ctx:             ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments:  scenario13ExpectedValidatedPayment,
				SuccessfulPayments: 0,
				PendingPayments:    0,
				FailedPayments:     7,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(false, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(false, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)

				for i := 0; i < 7; i++ {
					mockBulkValidationsDetailRepo.On("Create", ctx, mockTx, mock.Anything).Times(1).Return(mock.Anything, nil)
					mockPaymentRepo.On("FindByPaymentSequenceNumber", ctx, mockDB, mock.Anything).Times(1).Return(scenario13Payment[i], nil)
					mockInvoiceRepo.On("RetrieveInvoiceByInvoiceID", ctx, mockDB, mock.Anything).Times(1).Return(scenario13Invoice[i], nil)
					mockPaymentRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockInvoiceRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
					mockUserBasicInfoRepo.On("FindByID", ctx, mockDB, mock.Anything).Times(1).Return(scenario13User[i], nil)
					mockInvoiceActionLogRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(nil)
				}

				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)

			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			validationResult, err := directDebitValidator.Validate(testCase.ctx, &PaymentFile{DirectDebitFile: testCase.directDebitFile})

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

func TestDirectDebitTextPaymentFileValidator_EnableImproveBulkPaymentValidation(t *testing.T) {
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

	directDebitValidator := &DirectDebitTextPaymentFileValidator{
		BasePaymentFileValidator: baseValidator,
		PaymentDate:              time.Now().UTC(),
	}

	invoiceIssued := invoice_pb.InvoiceStatus_ISSUED.String()
	invoiceVoid := invoice_pb.InvoiceStatus_VOID.String()
	paymentPending := invoice_pb.PaymentStatus_PAYMENT_PENDING.String()
	paymentFailed := invoice_pb.PaymentStatus_PAYMENT_FAILED.String()

	// Data for scenario 1 which have result code equal to 0
	scenario1Invoice, scenario1Payment, scenario1User, scenario1FileContent := genTestDirectDebitValidatorData(
		genTestDirectDebitDataInput{
			Total:           []int{100},
			InvoiceStatuses: []string{invoiceIssued},
			PaymentStatuses: []string{paymentPending},
			ResultCodes:     []string{"0"},
			DepositAmount:   []int{100},
		},
	)

	mockScenario1InvoicePaymentUser := getPaymentInvoiceUserList(scenario1Invoice, scenario1Payment, scenario1User)

	// Data for scenario 2 which the result code is not 0
	scenario2Invoice, scenario2Payment, scenario2User, scenario2FileContent := genTestDirectDebitValidatorData(
		genTestDirectDebitDataInput{
			Total:           []int{100, 100, 100, 100, 100, 100},
			InvoiceStatuses: []string{invoiceIssued, invoiceIssued, invoiceIssued, invoiceIssued, invoiceIssued, invoiceIssued},
			PaymentStatuses: []string{paymentPending, paymentPending, paymentPending, paymentPending, paymentPending, paymentPending},
			ResultCodes:     []string{"1", "2", "3", "4", "8", "9"},
			DepositAmount:   []int{100, 100, 100, 100, 100, 100},
		},
	)
	scenario2ExpectedValidatedPayment := []*ValidatedPayment{}
	for i := 0; i < 6; i++ {
		scenario2ExpectedValidatedPayment = append(scenario2ExpectedValidatedPayment, &ValidatedPayment{
			PaymentSequenceNumber: int32(i + 1),
			InvoiceSequenceNumber: int32(i + 1),
			ResultCode:            fmt.Sprintf("D-R%v", scenario2FileContent[i].ResultCode),
			Amount:                100,
			PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT,
			PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
		})
	}

	mockScenario2InvoicePaymentUser := getPaymentInvoiceUserList(scenario2Invoice, scenario2Payment, scenario2User)

	// Data for scenario 9 which the deposit amount is not equal to invoice total amount
	scenario9Invoice, scenario9Payment, scenario9User, scenario9FileContent := genTestDirectDebitValidatorData(
		genTestDirectDebitDataInput{
			Total:           []int{100, 100, 100, 100, 100, 100, 100},
			InvoiceStatuses: []string{invoiceIssued, invoiceIssued, invoiceIssued, invoiceIssued, invoiceIssued, invoiceIssued, invoiceIssued},
			PaymentStatuses: []string{paymentPending, paymentPending, paymentPending, paymentPending, paymentPending, paymentPending, paymentPending},
			ResultCodes:     []string{"0", "1", "2", "3", "4", "8", "9"},
			DepositAmount:   []int{50, 50, 50, 50, 50, 50, 50},
		},
	)
	scenario9ExpectedValidatedPayment := []*ValidatedPayment{}
	for i := 0; i < 7; i++ {
		scenario9ExpectedValidatedPayment = append(scenario9ExpectedValidatedPayment, &ValidatedPayment{
			PaymentSequenceNumber: int32(i + 1),
			InvoiceSequenceNumber: int32(i + 1),
			ResultCode:            fmt.Sprintf("D-R%v-1", scenario9FileContent[i].ResultCode),
			Amount:                100,
			PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT,
			PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
		})
	}

	mockScenario9InvoicePaymentUser := getPaymentInvoiceUserList(scenario9Invoice, scenario9Payment, scenario9User)

	// Data for scenario 11-12 where the invoice status is VOID and payment status is FAILED
	scenario11Invoice, scenario11Payment, scenario11User, scenario11FileContent := genTestDirectDebitValidatorData(
		genTestDirectDebitDataInput{
			Total:           []int{100, 100, 100, 100, 100, 100, 100},
			InvoiceStatuses: []string{invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid},
			PaymentStatuses: []string{paymentFailed, paymentFailed, paymentFailed, paymentFailed, paymentFailed, paymentFailed, paymentFailed},
			ResultCodes:     []string{"0", "1", "2", "3", "4", "8", "9"},
			DepositAmount:   []int{100, 100, 100, 100, 100, 100, 100},
		},
	)
	scenario11ExpectedValidatedPayment := []*ValidatedPayment{}
	for i := 0; i < 7; i++ {
		scenario11ExpectedValidatedPayment = append(scenario11ExpectedValidatedPayment, &ValidatedPayment{
			PaymentSequenceNumber: int32(i + 1),
			InvoiceSequenceNumber: int32(i + 1),
			ResultCode:            fmt.Sprintf("D-R%v-2", scenario9FileContent[i].ResultCode),
			Amount:                100,
			PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT,
			PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
		})
	}

	mockScenario11InvoicePaymentUser := getPaymentInvoiceUserList(scenario11Invoice, scenario11Payment, scenario11User)

	// Data for scenario 13 where the invoice status is VOID and payment status is FAILED and the deposit amount is not the same with invoice total amount
	scenario13Invoice, scenario13Payment, scenario13User, scenario13FileContent := genTestDirectDebitValidatorData(
		genTestDirectDebitDataInput{
			Total:           []int{100, 100, 100, 100, 100, 100, 100},
			InvoiceStatuses: []string{invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid, invoiceVoid},
			PaymentStatuses: []string{paymentFailed, paymentFailed, paymentFailed, paymentFailed, paymentFailed, paymentFailed, paymentFailed},
			ResultCodes:     []string{"0", "1", "2", "3", "4", "8", "9"},
			DepositAmount:   []int{50, 50, 50, 50, 50, 50, 50},
		},
	)
	scenario13ExpectedValidatedPayment := []*ValidatedPayment{}
	for i := 0; i < 7; i++ {
		scenario13ExpectedValidatedPayment = append(scenario13ExpectedValidatedPayment, &ValidatedPayment{
			PaymentSequenceNumber: int32(i + 1),
			InvoiceSequenceNumber: int32(i + 1),
			ResultCode:            fmt.Sprintf("D-R%v-3", scenario9FileContent[i].ResultCode),
			Amount:                100,
			PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT,
			PaymentStatus:         invoice_pb.PaymentStatus_PAYMENT_FAILED.String(),
		})
	}

	mockScenario13InvoicePaymentUser := getPaymentInvoiceUserList(scenario13Invoice, scenario13Payment, scenario13User)

	type testCase struct {
		name            string
		directDebitFile *DirectDebitFile
		ctx             context.Context
		expectedResult  interface{}
		expectedErr     error
		setup           func(ctx context.Context)
	}

	for _, testCase := range []*testCase{
		{
			name:            "Scenario 1 - Successful direct debit payment",
			directDebitFile: genTestDirectDebitFile(scenario1FileContent),
			ctx:             ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments: []*ValidatedPayment{
					{
						PaymentSequenceNumber: 1,
						InvoiceSequenceNumber: 1,
						ResultCode:            "D-R0",
						Amount:                100,
						PaymentMethod:         invoice_pb.PaymentMethod_DIRECT_DEBIT,
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
				mockPaymentRepo.On("FindPaymentInvoiceUserFromTempTable", ctx, mockTx).Times(1).Return(mockScenario1InvoicePaymentUser, nil)
				mockPaymentRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsDetailRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockInvoiceActionLogRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:            "Scenario 2 - Failed direct debit result code",
			directDebitFile: genTestDirectDebitFile(scenario2FileContent),
			ctx:             ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments:  scenario2ExpectedValidatedPayment,
				SuccessfulPayments: 0,
				PendingPayments:    0,
				FailedPayments:     6,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(true, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockPaymentRepo.On("InsertPaymentNumbersTempTable", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockPaymentRepo.On("FindPaymentInvoiceUserFromTempTable", ctx, mockTx).Times(1).Return(mockScenario2InvoicePaymentUser, nil)
				mockPaymentRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsDetailRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockInvoiceActionLogRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)

			},
		},
		{
			name:            "Scenario 9 - Invoice total amount is not equal to deposit amount",
			directDebitFile: genTestDirectDebitFile(scenario9FileContent),
			ctx:             ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments:  scenario9ExpectedValidatedPayment,
				SuccessfulPayments: 0,
				PendingPayments:    0,
				FailedPayments:     7,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(true, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockPaymentRepo.On("InsertPaymentNumbersTempTable", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockPaymentRepo.On("FindPaymentInvoiceUserFromTempTable", ctx, mockTx).Times(1).Return(mockScenario9InvoicePaymentUser, nil)
				mockPaymentRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsDetailRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockInvoiceActionLogRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:            "Scenario 11-12 - Invoice status is VOID and payment status is FAILED",
			directDebitFile: genTestDirectDebitFile(scenario11FileContent),
			ctx:             ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments:  scenario11ExpectedValidatedPayment,
				SuccessfulPayments: 0,
				PendingPayments:    0,
				FailedPayments:     7,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(true, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockPaymentRepo.On("InsertPaymentNumbersTempTable", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockPaymentRepo.On("FindPaymentInvoiceUserFromTempTable", ctx, mockTx).Times(1).Return(mockScenario11InvoicePaymentUser, nil)
				mockPaymentRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockInvoiceRepo.On("UpdateMultipleWithFields", ctx, mockTx, mock.Anything, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsDetailRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockInvoiceActionLogRepo.On("CreateMultiple", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockBulkValidationsRepo.On("UpdateWithFields", ctx, mockTx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name:            "Scenario 13 - Invoice status is VOID and payment status is FAILED and deposit amount not equal to invoice total amount",
			directDebitFile: genTestDirectDebitFile(scenario13FileContent),
			ctx:             ctx,
			expectedResult: &PaymentValidationResult{
				ValidatedPayments:  scenario13ExpectedValidatedPayment,
				SuccessfulPayments: 0,
				PendingPayments:    0,
				FailedPayments:     7,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableImproveBulkPaymentValidation, mock.Anything).Once().Return(true, nil)
				mockUnleashClient.On("IsFeatureEnabled", constant.EnableBulkAddValidatePh2, mock.Anything).Once().Return(true, nil)

				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockBulkValidationsRepo.On("Create", ctx, mockTx, mock.Anything).Once().Return(mock.Anything, nil)
				mockPaymentRepo.On("InsertPaymentNumbersTempTable", ctx, mockTx, mock.Anything).Times(1).Return(nil)
				mockPaymentRepo.On("FindPaymentInvoiceUserFromTempTable", ctx, mockTx).Times(1).Return(mockScenario13InvoicePaymentUser, nil)
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
			validationResult, err := directDebitValidator.Validate(testCase.ctx, &PaymentFile{DirectDebitFile: testCase.directDebitFile})

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
