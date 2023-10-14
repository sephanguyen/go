package utils

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
)

// PaymentRequestValidator contains the reusable validations that can be used by create payment and download payment file
type PaymentRequestValidator struct{}

type FeatureFlags struct {
	EnableOptionalValidationInPaymentRequest bool
}

func (v *PaymentRequestValidator) ValidatePayment(payment *entities.Payment, paymentMethod string, isExported bool, maxPaymentSequenceNumber int) error {
	if payment.PaymentMethod.String != paymentMethod {
		return errors.New("The payment method is not equal to the given payment method parameter")
	}

	if payment.PaymentStatus.String != invoice_pb.PaymentStatus_PAYMENT_PENDING.String() {
		return errors.New("The payment status should be PENDING")
	}

	if payment.IsExported.Bool != isExported {
		return fmt.Errorf("Payment isExported field should be %v", isExported)
	}

	// Check if the payment sequence number digit length exceeds the requirement
	paymentSeqNumStr := strconv.Itoa(int(payment.PaymentSequenceNumber.Int))
	if len(paymentSeqNumStr) > maxPaymentSequenceNumber {
		return errors.New("The payment sequence number length exceeds the limit")
	}

	return nil
}

func (v *PaymentRequestValidator) ValidateInvoice(invoice *entities.Invoice, isExported bool, maxTotalAmount int) error {
	// Check if the number of digits in the total amount exceeds the file requirement
	exactTotal, err := GetFloat64ExactValueAndDecimalPlaces(invoice.Total, "2")
	if err != nil {
		return err
	}
	if len(strconv.FormatInt(int64(exactTotal), 10)) > maxTotalAmount {
		return errors.New("The invoice total length exceeds the limit")
	}

	return nil
}

func (v *PaymentRequestValidator) ValidateStudentPaymentDetail(spd *entities.StudentPaymentDetail) error {
	if strings.TrimSpace(spd.StudentPaymentDetailID.String) == "" {
		return errors.New("There is no student payment detail")
	}

	if strings.TrimSpace(spd.PayerName.String) == "" {
		return errors.New("The payer name in student payment detail is empty")
	}

	if strings.TrimSpace(spd.PaymentMethod.String) == "" {
		return fmt.Errorf("student: %v payment method in student payment detail is empty", spd.StudentID.String)
	}

	return nil
}

func (v *PaymentRequestValidator) ValidateBillingAddress(billingAddress *entities.BillingAddress, flags *FeatureFlags) error {
	if strings.TrimSpace(billingAddress.BillingAddressID.String) == "" {
		return errors.New("There is no billing address")
	}

	if strings.TrimSpace(billingAddress.PostalCode.String) == "" {
		return errors.New("The student postal code is empty")
	}

	if strings.TrimSpace(billingAddress.PrefectureCode.String) == "" {
		return errors.New("the student prefecture code is empty")
	}

	if strings.TrimSpace(billingAddress.City.String) == "" {
		return errors.New("The student city is empty")
	}

	if strings.TrimSpace(billingAddress.Street1.String) == "" && !flags.EnableOptionalValidationInPaymentRequest {
		return errors.New("The student street1 is empty")
	}

	return nil
}

func (v *PaymentRequestValidator) ValidateBankAccount(bankAccount *entities.BankAccount) error {
	if strings.TrimSpace(bankAccount.BankAccountID.String) == "" {
		return errors.New("There is no bank account")
	}

	if strings.TrimSpace(bankAccount.BankAccountNumber.String) == "" {
		return errors.New("The student bank account number is empty")
	}

	if strings.TrimSpace(bankAccount.BankAccountHolder.String) == "" {
		return errors.New("The student bank account holder is empty")
	}

	if strings.TrimSpace(bankAccount.BankAccountType.String) == "" {
		return errors.New("The student bank account type is empty")
	}

	// Validate the length
	if len(bankAccount.BankAccountNumber.String) != 7 {
		return errors.New("The bank account number can only accept 7 digit numbers.")
	}

	// Check if verified
	if !bankAccount.IsVerified.Bool {
		return errors.New("The bank account is not verified")
	}

	var depositItems string
	for k, v := range constant.PartnerBankDepositItems {
		if v == bankAccount.BankAccountType.String {
			depositItems = strconv.Itoa(k)
		}
	}

	if depositItems == "" {
		return errors.New("The partner bank deposit item name doesn't have equivalent int value. Please check the default partner bank.")
	}

	return nil
}

func (v *PaymentRequestValidator) ValidateBankBranch(bankBranch *entities.BankBranch) error {
	if strings.TrimSpace(bankBranch.BankBranchID.String) == "" {
		return errors.New("The bank branch does not exist")
	}

	if strings.TrimSpace(bankBranch.BankBranchCode.String) == "" {
		return errors.New("The bank branch code is empty")
	}

	if strings.TrimSpace(bankBranch.BankBranchName.String) == "" {
		return errors.New("The bank branch name is empty")
	}

	// Validate the length
	if len(bankBranch.BankBranchCode.String) > 3 {
		return errors.New("The bank branch code length exceeds the limit. Please check the bank branch.")
	}

	return nil
}

func (v *PaymentRequestValidator) ValidateBank(bank *entities.Bank) error {
	if strings.TrimSpace(bank.BankID.String) == "" {
		return errors.New("The bank does not exist")
	}

	if strings.TrimSpace(bank.BankCode.String) == "" {
		return errors.New("The bank code is empty")
	}

	if strings.TrimSpace(bank.BankName.String) == "" {
		return errors.New("The bank name is empty")
	}

	// Validate the length
	if len(bank.BankCode.String) > 4 {
		return errors.New("The bank code length exceeds the limit. Please check the bank.")
	}

	return nil
}

func (v *PaymentRequestValidator) ValidatePartnerConvenienceStore(partnerCS *entities.PartnerConvenienceStore) error {
	if strings.TrimSpace(partnerCS.CompanyName.String) == "" {
		return errors.New("The partner CS company name is empty")
	}

	// Validate the length
	if len(strconv.Itoa(int(partnerCS.ManufacturerCode.Int))) != 6 {
		return errors.New("The manufacturer_code of the partner CS should be 6 digits")
	}

	if len(strconv.Itoa(int(partnerCS.CompanyCode.Int))) != 5 {
		return errors.New("The company_code of the partner CS should be 5 digits")
	}

	return nil
}

func (v *PaymentRequestValidator) ValidatePartnerBank(partnerBank *entities.PartnerBank) error {
	if strings.TrimSpace(partnerBank.ConsignorCode.String) == "" {
		return errors.New("The partner bank consignor code is empty")
	}

	if strings.TrimSpace(partnerBank.ConsignorName.String) == "" {
		return errors.New("The partner bank consignor name is empty")
	}

	if strings.TrimSpace(partnerBank.BankNumber.String) == "" {
		return errors.New("The partner bank number is empty")
	}

	if strings.TrimSpace(partnerBank.BankName.String) == "" {
		return errors.New("The partner bank name is empty")
	}

	if strings.TrimSpace(partnerBank.BankBranchNumber.String) == "" {
		return errors.New("The partner bank branch number is empty")
	}

	if strings.TrimSpace(partnerBank.BankBranchName.String) == "" {
		return errors.New("The partner bank branch name is empty")
	}

	if strings.TrimSpace(partnerBank.DepositItems.String) == "" {
		return errors.New("The partner bank deposit item is empty")
	}

	if strings.TrimSpace(partnerBank.AccountNumber.String) == "" {
		return errors.New("The partner bank account number is empty")
	}

	// Validate the length
	if len(partnerBank.BankNumber.String) > 4 {
		return errors.New("The partner bank number length exceeds the limit. Please check the default partner bank.")
	}

	if len(partnerBank.BankBranchNumber.String) > 3 {
		return errors.New("The partner bank branch number length exceeds the limit. Please check the default partner bank.")
	}

	if len(partnerBank.AccountNumber.String) != 7 {
		return errors.New("The partner bank account number can only accept 7 digit numbers.")
	}

	if len(partnerBank.ConsignorCode.String) > 10 {
		return errors.New("The partner bank consignor code length exceeds the limit. Please check the default partner bank.")
	}

	// validate the deposit item
	var depositItems string
	for k, v := range constant.PartnerBankDepositItems {
		if v == partnerBank.DepositItems.String {
			depositItems = strconv.Itoa(k)
		}
	}

	if depositItems == "" {
		return errors.New("The partner bank deposit item name doesn't have equivalent int value. Please check the default partner bank.")
	}

	return nil
}
