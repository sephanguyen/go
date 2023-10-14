package utils

import (
	"errors"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/stretchr/testify/assert"
)

func Test_PaymentRequestValidator_ValidatePayment(t *testing.T) {

	type testCase struct {
		name                     string
		payment                  *entities.Payment
		paymentMethod            string
		isExported               bool
		maxPaymentSequenceNumber int
		expectedError            error
	}

	id := "test-id"

	testCases := []testCase{
		{
			name: "Happy case",
			payment: &entities.Payment{
				PaymentID:             database.Text(id),
				PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
				PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
				IsExported:            database.Bool(true),
				PaymentSequenceNumber: database.Int4(1),
			},
			paymentMethod:            invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
			isExported:               true,
			maxPaymentSequenceNumber: 10,
			expectedError:            nil,
		},
		{
			name: "Payment method not equal",
			payment: &entities.Payment{
				PaymentID:             database.Text(id),
				PaymentMethod:         database.Text(invoice_pb.PaymentMethod_DIRECT_DEBIT.String()),
				PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
				IsExported:            database.Bool(true),
				PaymentSequenceNumber: database.Int4(1),
			},
			paymentMethod:            invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
			isExported:               true,
			maxPaymentSequenceNumber: 10,
			expectedError:            errors.New("The payment method is not equal to the given payment method parameter"),
		},
		{
			name: "Payment status is not PENDING",
			payment: &entities.Payment{
				PaymentID:     database.Text(id),
				PaymentMethod: database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
				PaymentStatus: database.Text(invoice_pb.PaymentStatus_PAYMENT_FAILED.String()),
			},
			paymentMethod:            invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
			isExported:               true,
			maxPaymentSequenceNumber: 10,
			expectedError:            errors.New("The payment status should be PENDING"),
		},
		{
			name: "Payment isExported not equal",
			payment: &entities.Payment{
				PaymentID:             database.Text(id),
				PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
				PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
				IsExported:            database.Bool(false),
				PaymentSequenceNumber: database.Int4(1),
			},
			paymentMethod:            invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
			isExported:               true,
			maxPaymentSequenceNumber: 10,
			expectedError:            fmt.Errorf("Payment isExported field should be %v", true),
		},
		{
			name: "Payment sequence nubmer exceeds the limit",
			payment: &entities.Payment{
				PaymentID:             database.Text(id),
				PaymentMethod:         database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
				PaymentStatus:         database.Text(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
				IsExported:            database.Bool(true),
				PaymentSequenceNumber: database.Int4(123456789),
			},
			paymentMethod:            invoice_pb.PaymentMethod_CONVENIENCE_STORE.String(),
			isExported:               true,
			maxPaymentSequenceNumber: 5,
			expectedError:            errors.New("The payment sequence number length exceeds the limit"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := &PaymentRequestValidator{}

			err := validator.ValidatePayment(tc.payment, tc.paymentMethod, tc.isExported, tc.maxPaymentSequenceNumber)
			assert.Equal(t, err, tc.expectedError)
		})
	}
}

func Test_PaymentRequestValidator_ValidateInvoice(t *testing.T) {

	type testCase struct {
		name           string
		invoice        *entities.Invoice
		isExported     bool
		maxTotalLength int
		expectedError  error
	}

	id := "test-id"

	testCases := []testCase{
		{
			name: "Happy case",
			invoice: &entities.Invoice{
				InvoiceID:  database.Text(id),
				Total:      database.Numeric(100),
				IsExported: database.Bool(true),
			},
			isExported:     true,
			maxTotalLength: 10,
			expectedError:  nil,
		},
		{
			name: "Invoice total exceeds the limit",
			invoice: &entities.Invoice{
				InvoiceID:  database.Text(id),
				Total:      database.Numeric(100),
				IsExported: database.Bool(true),
			},
			isExported:     true,
			maxTotalLength: 1,
			expectedError:  errors.New("The invoice total length exceeds the limit"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := &PaymentRequestValidator{}

			err := validator.ValidateInvoice(tc.invoice, tc.isExported, tc.maxTotalLength)
			assert.Equal(t, err, tc.expectedError)
		})
	}
}

func Test_PaymentRequestValidator_ValidateStudentPaymentDetail(t *testing.T) {

	type testCase struct {
		name                 string
		studentPaymentDetail *entities.StudentPaymentDetail
		expectedError        error
	}

	id := "test-id"

	testCases := []testCase{
		{
			name: "Happy case",
			studentPaymentDetail: &entities.StudentPaymentDetail{
				StudentPaymentDetailID: database.Text(id),
				PayerName:              database.Text("test-payer-name"),
				PayerPhoneNumber:       database.Text("123-123-123"),
				PaymentMethod:          database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
			},
			expectedError: nil,
		},
		{
			name: "Happy case without the optional field (PayerPhoneNumber)",
			studentPaymentDetail: &entities.StudentPaymentDetail{
				StudentPaymentDetailID: database.Text(id),
				PayerName:              database.Text("test-payer-name"),
				PaymentMethod:          database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
			},
			expectedError: nil,
		},
		{
			name: "Student Payment Detail ID is empty",
			studentPaymentDetail: &entities.StudentPaymentDetail{
				StudentPaymentDetailID: database.Text(""),
				PayerName:              database.Text("test-payer-name"),
				PayerPhoneNumber:       database.Text("123-123-123"),
				PaymentMethod:          database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
			},
			expectedError: errors.New("There is no student payment detail"),
		},
		{
			name: "Payer name is empty",
			studentPaymentDetail: &entities.StudentPaymentDetail{
				StudentPaymentDetailID: database.Text(id),
				PayerName:              database.Text(""),
				PayerPhoneNumber:       database.Text("123-123-123"),
				PaymentMethod:          database.Text(invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()),
			},
			expectedError: errors.New("The payer name in student payment detail is empty"),
		},
		{
			name: "Payment method is empty",
			studentPaymentDetail: &entities.StudentPaymentDetail{
				StudentID:              database.Text("test-student"),
				StudentPaymentDetailID: database.Text(id),
				PayerName:              database.Text("test"),
				PayerPhoneNumber:       database.Text("123-123-123"),
				PaymentMethod:          database.Text(""),
			},
			expectedError: errors.New("student: test-student payment method in student payment detail is empty"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := &PaymentRequestValidator{}

			err := validator.ValidateStudentPaymentDetail(tc.studentPaymentDetail)
			assert.Equal(t, err, tc.expectedError)
		})
	}
}

func Test_PaymentRequestValidator_ValidateBillingAddress(t *testing.T) {

	type testCase struct {
		name                                     string
		billingAddress                           *entities.BillingAddress
		expectedError                            error
		enableOptionalValidationInPaymentRequest bool
	}

	const (
		id                 = "test-id"
		testPostalCode     = "test-postal-code"
		testPrefectureCode = "test-prefecture-code"
		testCity           = "test-city"
		testStreet1        = "test-street1"
	)

	testCases := []testCase{
		{
			name: "Happy case",
			billingAddress: &entities.BillingAddress{
				BillingAddressID: database.Text(id),
				PostalCode:       database.Text(testPostalCode),
				PrefectureCode:   database.Text(testPrefectureCode),
				City:             database.Text(testCity),
				Street1:          database.Text(testStreet1),
			},
			expectedError: nil,
		},
		{
			name: "Empty billing address id",
			billingAddress: &entities.BillingAddress{
				BillingAddressID: database.Text(""),
				PostalCode:       database.Text(testPostalCode),
				PrefectureCode:   database.Text(testPrefectureCode),
				City:             database.Text(testCity),
				Street1:          database.Text(testStreet1),
			},
			expectedError: errors.New("There is no billing address"),
		},
		{
			name: "Empty postal code",
			billingAddress: &entities.BillingAddress{
				BillingAddressID: database.Text(id),
				PostalCode:       database.Text(""),
				PrefectureCode:   database.Text(testPrefectureCode),
				City:             database.Text(testCity),
				Street1:          database.Text(testStreet1),
			},
			expectedError: errors.New("The student postal code is empty"),
		},
		{
			name: "Empty prefecture code",
			billingAddress: &entities.BillingAddress{
				BillingAddressID: database.Text(id),
				PostalCode:       database.Text(testPostalCode),
				PrefectureCode:   database.Text(""),
				City:             database.Text(testCity),
				Street1:          database.Text(testStreet1),
			},
			expectedError: errors.New("the student prefecture code is empty"),
		},
		{
			name: "Empty city",
			billingAddress: &entities.BillingAddress{
				BillingAddressID: database.Text(id),
				PostalCode:       database.Text(testPostalCode),
				PrefectureCode:   database.Text(testPrefectureCode),
				City:             database.Text(""),
				Street1:          database.Text(testStreet1),
			},
			expectedError: errors.New("The student city is empty"),
		},
		{
			name: "Empty street1",
			billingAddress: &entities.BillingAddress{
				BillingAddressID: database.Text(id),
				PostalCode:       database.Text(testPostalCode),
				PrefectureCode:   database.Text(testPrefectureCode),
				City:             database.Text(testCity),
				Street1:          database.Text(""),
			},
			expectedError: errors.New("The student street1 is empty"),
		},
		{
			name: "Empty street1 with enableOptionalValidationInPaymentRequest = true",
			billingAddress: &entities.BillingAddress{
				BillingAddressID: database.Text(id),
				PostalCode:       database.Text(testPostalCode),
				PrefectureCode:   database.Text(testPrefectureCode),
				City:             database.Text(testCity),
				Street1:          database.Text(""),
			},
			enableOptionalValidationInPaymentRequest: true,
			expectedError:                            nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := &PaymentRequestValidator{}

			err := validator.ValidateBillingAddress(tc.billingAddress, &FeatureFlags{EnableOptionalValidationInPaymentRequest: tc.enableOptionalValidationInPaymentRequest})
			assert.Equal(t, err, tc.expectedError)
		})
	}
}

func Test_PaymentRequestValidator_ValidateBankAccount(t *testing.T) {

	type testCase struct {
		name          string
		bankAccount   *entities.BankAccount
		expectedError error
	}
	const (
		id                       = "test-id"
		partnerBankAccountNumber = "1234567"
	)

	testCases := []testCase{
		{
			name: "Happy case",
			bankAccount: &entities.BankAccount{
				BankAccountID:     database.Text(id),
				BankAccountNumber: database.Text(partnerBankAccountNumber),
				BankAccountHolder: database.Text("test-bank-account-holder"),
				BankAccountType:   database.Text(constant.PartnerBankDepositItems[1]),
				IsVerified:        database.Bool(true),
			},
			expectedError: nil,
		},
		{
			name: "Bank account id is empty",
			bankAccount: &entities.BankAccount{
				BankAccountID:     database.Text(""),
				BankAccountNumber: database.Text("123"),
				BankAccountHolder: database.Text("test-bank-account-holder"),
				BankAccountType:   database.Text(constant.PartnerBankDepositItems[1]),
				IsVerified:        database.Bool(true),
			},
			expectedError: errors.New("There is no bank account"),
		},
		{
			name: "Bank account number is empty",
			bankAccount: &entities.BankAccount{
				BankAccountID:     database.Text(id),
				BankAccountNumber: database.Text(""),
				BankAccountHolder: database.Text("test-bank-account-holder"),
				BankAccountType:   database.Text(constant.PartnerBankDepositItems[1]),
				IsVerified:        database.Bool(true),
			},
			expectedError: errors.New("The student bank account number is empty"),
		},
		{
			name: "Bank account holder is empty",
			bankAccount: &entities.BankAccount{
				BankAccountID:     database.Text(id),
				BankAccountNumber: database.Text("123"),
				BankAccountHolder: database.Text(""),
				BankAccountType:   database.Text(constant.PartnerBankDepositItems[1]),
				IsVerified:        database.Bool(true),
			},
			expectedError: errors.New("The student bank account holder is empty"),
		},
		{
			name: "Bank account type is empty",
			bankAccount: &entities.BankAccount{
				BankAccountID:     database.Text(id),
				BankAccountNumber: database.Text("123"),
				BankAccountHolder: database.Text("test-bank-account-holder"),
				BankAccountType:   database.Text(""),
				IsVerified:        database.Bool(true),
			},
			expectedError: errors.New("The student bank account type is empty"),
		},
		{
			name: "Bank account number not exact 7 digits",
			bankAccount: &entities.BankAccount{
				BankAccountID:     database.Text(id),
				BankAccountNumber: database.Text("123123"),
				BankAccountHolder: database.Text("test-bank-account-holder"),
				BankAccountType:   database.Text(constant.PartnerBankDepositItems[1]),
				IsVerified:        database.Bool(true),
			},
			expectedError: errors.New("The bank account number can only accept 7 digit numbers."),
		},
		{
			name: "Bank account is not verified",
			bankAccount: &entities.BankAccount{
				BankAccountID:     database.Text(id),
				BankAccountNumber: database.Text(partnerBankAccountNumber),
				BankAccountHolder: database.Text("test-bank-account-holder"),
				BankAccountType:   database.Text(constant.PartnerBankDepositItems[1]),
				IsVerified:        database.Bool(false),
			},
			expectedError: errors.New("The bank account is not verified"),
		},
		{
			name: "Bank account type is invalid",
			bankAccount: &entities.BankAccount{
				BankAccountID:     database.Text(id),
				BankAccountNumber: database.Text(partnerBankAccountNumber),
				BankAccountHolder: database.Text("test-bank-account-holder"),
				BankAccountType:   database.Text("invalid"),
				IsVerified:        database.Bool(true),
			},
			expectedError: errors.New("The partner bank deposit item name doesn't have equivalent int value. Please check the default partner bank."),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := &PaymentRequestValidator{}

			err := validator.ValidateBankAccount(tc.bankAccount)
			assert.Equal(t, err, tc.expectedError)
		})
	}
}

func Test_PaymentRequestValidator_ValidateBankBranch(t *testing.T) {

	type testCase struct {
		name          string
		bankBranch    *entities.BankBranch
		expectedError error
	}

	id := "test-id"

	testCases := []testCase{
		{
			name: "Happy case",
			bankBranch: &entities.BankBranch{
				BankBranchID:   database.Text(id),
				BankBranchCode: database.Text("123"),
				BankBranchName: database.Text("test-bank-branch-name"),
			},
			expectedError: nil,
		},
		{
			name: "Bank branch id is empty",
			bankBranch: &entities.BankBranch{
				BankBranchID:   database.Text(""),
				BankBranchCode: database.Text("123"),
				BankBranchName: database.Text("test-bank-branch-name"),
			},
			expectedError: errors.New("The bank branch does not exist"),
		},
		{
			name: "Bank branch code is empty",
			bankBranch: &entities.BankBranch{
				BankBranchID:   database.Text(id),
				BankBranchCode: database.Text(""),
				BankBranchName: database.Text("test-bank-branch-name"),
			},
			expectedError: errors.New("The bank branch code is empty"),
		},
		{
			name: "Bank branch name is empty",
			bankBranch: &entities.BankBranch{
				BankBranchID:   database.Text(id),
				BankBranchCode: database.Text("123"),
				BankBranchName: database.Text(""),
			},
			expectedError: errors.New("The bank branch name is empty"),
		},
		{
			name: "Bank branch code exceeds the limit",
			bankBranch: &entities.BankBranch{
				BankBranchID:   database.Text(id),
				BankBranchCode: database.Text("12345"),
				BankBranchName: database.Text("test-bank-branch-name"),
			},
			expectedError: errors.New("The bank branch code length exceeds the limit. Please check the bank branch."),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := &PaymentRequestValidator{}

			err := validator.ValidateBankBranch(tc.bankBranch)
			assert.Equal(t, err, tc.expectedError)
		})
	}
}

func Test_PaymentRequestValidator_ValidateBank(t *testing.T) {

	type testCase struct {
		name          string
		bank          *entities.Bank
		expectedError error
	}

	id := "test-id"

	testCases := []testCase{
		{
			name: "Happy case",
			bank: &entities.Bank{
				BankID:   database.Text(id),
				BankCode: database.Text("1234"),
				BankName: database.Text("test-bank-name"),
			},
			expectedError: nil,
		},
		{
			name: "Bank id is empty",
			bank: &entities.Bank{
				BankID:   database.Text(""),
				BankCode: database.Text("1234"),
				BankName: database.Text("test-bank-name"),
			},
			expectedError: errors.New("The bank does not exist"),
		},
		{
			name: "Bank code is empty",
			bank: &entities.Bank{
				BankID:   database.Text(id),
				BankCode: database.Text(""),
				BankName: database.Text("test-bank-name"),
			},
			expectedError: errors.New("The bank code is empty"),
		},
		{
			name: "Bank name is empty",
			bank: &entities.Bank{
				BankID:   database.Text(id),
				BankCode: database.Text("1234"),
				BankName: database.Text(""),
			},
			expectedError: errors.New("The bank name is empty"),
		},
		{
			name: "Bank code exceeds the limit",
			bank: &entities.Bank{
				BankID:   database.Text(id),
				BankCode: database.Text("123456"),
				BankName: database.Text("test-bank-name"),
			},
			expectedError: errors.New("The bank code length exceeds the limit. Please check the bank."),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := &PaymentRequestValidator{}

			err := validator.ValidateBank(tc.bank)
			assert.Equal(t, err, tc.expectedError)
		})
	}
}

func Test_PaymentRequestValidator_ValidatePartnerConvenienceStore(t *testing.T) {

	type testCase struct {
		name          string
		partnerCS     *entities.PartnerConvenienceStore
		expectedError error
	}

	id := "test-id"

	testCases := []testCase{
		{
			name: "Happy case",
			partnerCS: &entities.PartnerConvenienceStore{
				PartnerConvenienceStoreID: database.Text(id),
				ShopCode:                  database.Text("1234"),
				CompanyName:               database.Text("test-company-name"),
				CompanyTelNumber:          database.Text("123-4567-890"),
				PostalCode:                database.Text("1234"),
				ManufacturerCode:          database.Int4(123456),
				CompanyCode:               database.Int4(12345),
			},
			expectedError: nil,
		},
		{
			name: "Happy case without the optional fields (ShopCode, PostalCode, CompanyTelNumber)",
			partnerCS: &entities.PartnerConvenienceStore{
				PartnerConvenienceStoreID: database.Text(id),
				CompanyName:               database.Text("test-company-name"),
				ManufacturerCode:          database.Int4(123456),
				CompanyCode:               database.Int4(12345),
			},
			expectedError: nil,
		},
		{
			name: "Company name is empty",
			partnerCS: &entities.PartnerConvenienceStore{
				PartnerConvenienceStoreID: database.Text(id),
				ShopCode:                  database.Text("1234"),
				CompanyName:               database.Text(""),
				CompanyTelNumber:          database.Text("123-4567-890"),
				PostalCode:                database.Text("1234"),
				ManufacturerCode:          database.Int4(123456),
				CompanyCode:               database.Int4(12345),
			},
			expectedError: errors.New("The partner CS company name is empty"),
		},
		{
			name: "Manufacturer code incorrect length",
			partnerCS: &entities.PartnerConvenienceStore{
				PartnerConvenienceStoreID: database.Text(id),
				ShopCode:                  database.Text("1234"),
				CompanyName:               database.Text("test-company-name"),
				CompanyTelNumber:          database.Text("123-4567-890"),
				PostalCode:                database.Text("1234"),
				ManufacturerCode:          database.Int4(123),
				CompanyCode:               database.Int4(12345),
			},
			expectedError: errors.New("The manufacturer_code of the partner CS should be 6 digits"),
		},
		{
			name: "Company code incorrect length",
			partnerCS: &entities.PartnerConvenienceStore{
				PartnerConvenienceStoreID: database.Text(id),
				ShopCode:                  database.Text("1234"),
				CompanyName:               database.Text("test-company-name"),
				CompanyTelNumber:          database.Text("123-4567-890"),
				PostalCode:                database.Text("1234"),
				ManufacturerCode:          database.Int4(123456),
				CompanyCode:               database.Int4(123),
			},
			expectedError: errors.New("The company_code of the partner CS should be 5 digits"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := &PaymentRequestValidator{}

			err := validator.ValidatePartnerConvenienceStore(tc.partnerCS)
			assert.Equal(t, err, tc.expectedError)
		})
	}
}

func Test_PaymentRequestValidator_ValidatePartnerBank(t *testing.T) {

	type testCase struct {
		name          string
		partnerBank   *entities.PartnerBank
		expectedError error
	}
	const (
		id                       = "test-id"
		partnerBankAccountNumber = "1234567"
	)

	testCases := []testCase{
		{
			name: "Happy case",
			partnerBank: &entities.PartnerBank{
				PartnerBankID:    database.Text(id),
				ConsignorCode:    database.Text("1234567890"),
				ConsignorName:    database.Text("test-consignor-name"),
				BankNumber:       database.Text("1234"),
				BankName:         database.Text("test-bank-name"),
				BankBranchNumber: database.Text("123"),
				BankBranchName:   database.Text("bank-branch-name"),
				DepositItems:     database.Text(constant.PartnerBankDepositItems[1]),
				AccountNumber:    database.Text(partnerBankAccountNumber),
			},
			expectedError: nil,
		},
		{
			name: "Consignor code is empty",
			partnerBank: &entities.PartnerBank{
				PartnerBankID:    database.Text(id),
				ConsignorCode:    database.Text(""),
				ConsignorName:    database.Text("test-consignor-name"),
				BankNumber:       database.Text("1234"),
				BankName:         database.Text("test-bank-name"),
				BankBranchNumber: database.Text("123"),
				BankBranchName:   database.Text("bank-branch-name"),
				DepositItems:     database.Text(constant.PartnerBankDepositItems[1]),
				AccountNumber:    database.Text(partnerBankAccountNumber),
			},
			expectedError: errors.New("The partner bank consignor code is empty"),
		},
		{
			name: "Consignor name is empty",
			partnerBank: &entities.PartnerBank{
				PartnerBankID:    database.Text(id),
				ConsignorCode:    database.Text("1234567890"),
				ConsignorName:    database.Text(""),
				BankNumber:       database.Text("1234"),
				BankName:         database.Text("test-bank-name"),
				BankBranchNumber: database.Text("123"),
				BankBranchName:   database.Text("bank-branch-name"),
				DepositItems:     database.Text(constant.PartnerBankDepositItems[1]),
				AccountNumber:    database.Text(partnerBankAccountNumber),
			},
			expectedError: errors.New("The partner bank consignor name is empty"),
		},
		{
			name: "Partner bank number is empty",
			partnerBank: &entities.PartnerBank{
				PartnerBankID:    database.Text(id),
				ConsignorCode:    database.Text("1234567890"),
				ConsignorName:    database.Text("test-consignor-name"),
				BankNumber:       database.Text(""),
				BankName:         database.Text("test-bank-name"),
				BankBranchNumber: database.Text("123"),
				BankBranchName:   database.Text("bank-branch-name"),
				DepositItems:     database.Text(constant.PartnerBankDepositItems[1]),
				AccountNumber:    database.Text(partnerBankAccountNumber),
			},
			expectedError: errors.New("The partner bank number is empty"),
		},
		{
			name: "Partner bank name is empty",
			partnerBank: &entities.PartnerBank{
				PartnerBankID:    database.Text(id),
				ConsignorCode:    database.Text("1234567890"),
				ConsignorName:    database.Text("test-consignor-name"),
				BankNumber:       database.Text("1234"),
				BankName:         database.Text(""),
				BankBranchNumber: database.Text("123"),
				BankBranchName:   database.Text("bank-branch-name"),
				DepositItems:     database.Text(constant.PartnerBankDepositItems[1]),
				AccountNumber:    database.Text(partnerBankAccountNumber),
			},
			expectedError: errors.New("The partner bank name is empty"),
		},
		{
			name: "Partner bank branch number is empty",
			partnerBank: &entities.PartnerBank{
				PartnerBankID:    database.Text(id),
				ConsignorCode:    database.Text("1234567890"),
				ConsignorName:    database.Text("test-consignor-name"),
				BankNumber:       database.Text("1234"),
				BankName:         database.Text("test-bank-name"),
				BankBranchNumber: database.Text(""),
				BankBranchName:   database.Text("bank-branch-name"),
				DepositItems:     database.Text(constant.PartnerBankDepositItems[1]),
				AccountNumber:    database.Text(partnerBankAccountNumber),
			},
			expectedError: errors.New("The partner bank branch number is empty"),
		},
		{
			name: "Partner bank branch name is empty",
			partnerBank: &entities.PartnerBank{
				PartnerBankID:    database.Text(id),
				ConsignorCode:    database.Text("1234567890"),
				ConsignorName:    database.Text("test-consignor-name"),
				BankNumber:       database.Text("1234"),
				BankName:         database.Text("test-bank-name"),
				BankBranchNumber: database.Text("123"),
				BankBranchName:   database.Text(""),
				DepositItems:     database.Text(constant.PartnerBankDepositItems[1]),
				AccountNumber:    database.Text(partnerBankAccountNumber),
			},
			expectedError: errors.New("The partner bank branch name is empty"),
		},
		{
			name: "Partner bank deposit item is empty",
			partnerBank: &entities.PartnerBank{
				PartnerBankID:    database.Text(id),
				ConsignorCode:    database.Text("1234567890"),
				ConsignorName:    database.Text("test-consignor-name"),
				BankNumber:       database.Text("1234"),
				BankName:         database.Text("test-bank-name"),
				BankBranchNumber: database.Text("123"),
				BankBranchName:   database.Text("bank-branch-name"),
				DepositItems:     database.Text(""),
				AccountNumber:    database.Text(partnerBankAccountNumber),
			},
			expectedError: errors.New("The partner bank deposit item is empty"),
		},
		{
			name: "Partner bank account number is empty",
			partnerBank: &entities.PartnerBank{
				PartnerBankID:    database.Text(id),
				ConsignorCode:    database.Text("1234567890"),
				ConsignorName:    database.Text("test-consignor-name"),
				BankNumber:       database.Text("1234"),
				BankName:         database.Text("test-bank-name"),
				BankBranchNumber: database.Text("123"),
				BankBranchName:   database.Text("bank-branch-name"),
				DepositItems:     database.Text(constant.PartnerBankDepositItems[1]),
				AccountNumber:    database.Text(""),
			},
			expectedError: errors.New("The partner bank account number is empty"),
		},
		{
			name: "Partner bank number length exceeds limit",
			partnerBank: &entities.PartnerBank{
				PartnerBankID:    database.Text(id),
				ConsignorCode:    database.Text("1234567890"),
				ConsignorName:    database.Text("test-consignor-name"),
				BankNumber:       database.Text("123456"),
				BankName:         database.Text("test-bank-name"),
				BankBranchNumber: database.Text("123"),
				BankBranchName:   database.Text("bank-branch-name"),
				DepositItems:     database.Text(constant.PartnerBankDepositItems[1]),
				AccountNumber:    database.Text(partnerBankAccountNumber),
			},
			expectedError: errors.New("The partner bank number length exceeds the limit. Please check the default partner bank."),
		},
		{
			name: "Partner bank branch number exceeds the limit",
			partnerBank: &entities.PartnerBank{
				PartnerBankID:    database.Text(id),
				ConsignorCode:    database.Text("1234567890"),
				ConsignorName:    database.Text("test-consignor-name"),
				BankNumber:       database.Text("1234"),
				BankName:         database.Text("test-bank-name"),
				BankBranchNumber: database.Text("12345"),
				BankBranchName:   database.Text("bank-branch-name"),
				DepositItems:     database.Text(constant.PartnerBankDepositItems[1]),
				AccountNumber:    database.Text(partnerBankAccountNumber),
			},
			expectedError: errors.New("The partner bank branch number length exceeds the limit. Please check the default partner bank."),
		},
		{
			name: "Partner bank account number exceeds the limit",
			partnerBank: &entities.PartnerBank{
				PartnerBankID:    database.Text(id),
				ConsignorCode:    database.Text("1234567890"),
				ConsignorName:    database.Text("test-consignor-name"),
				BankNumber:       database.Text("1234"),
				BankName:         database.Text("test-bank-name"),
				BankBranchNumber: database.Text("123"),
				BankBranchName:   database.Text("bank-branch-name"),
				DepositItems:     database.Text(constant.PartnerBankDepositItems[1]),
				AccountNumber:    database.Text("1234567890"),
			},
			expectedError: errors.New("The partner bank account number can only accept 7 digit numbers."),
		},
		{
			name: "Partner bank consignor coded exceeds the limit",
			partnerBank: &entities.PartnerBank{
				PartnerBankID:    database.Text(id),
				ConsignorCode:    database.Text("1234567890123"),
				ConsignorName:    database.Text("test-consignor-name"),
				BankNumber:       database.Text("1234"),
				BankName:         database.Text("test-bank-name"),
				BankBranchNumber: database.Text("123"),
				BankBranchName:   database.Text("bank-branch-name"),
				DepositItems:     database.Text(constant.PartnerBankDepositItems[1]),
				AccountNumber:    database.Text(partnerBankAccountNumber),
			},
			expectedError: errors.New("The partner bank consignor code length exceeds the limit. Please check the default partner bank."),
		},
		{
			name: "Partner bank deposit item is invalid",
			partnerBank: &entities.PartnerBank{
				PartnerBankID:    database.Text(id),
				ConsignorCode:    database.Text("1234567890"),
				ConsignorName:    database.Text("test-consignor-name"),
				BankNumber:       database.Text("1234"),
				BankName:         database.Text("test-bank-name"),
				BankBranchNumber: database.Text("123"),
				BankBranchName:   database.Text("bank-branch-name"),
				DepositItems:     database.Text("invalid"),
				AccountNumber:    database.Text(partnerBankAccountNumber),
			},
			expectedError: errors.New("The partner bank deposit item name doesn't have equivalent int value. Please check the default partner bank."),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := &PaymentRequestValidator{}

			err := validator.ValidatePartnerBank(tc.partnerBank)
			assert.Equal(t, err, tc.expectedError)
		})
	}
}
