package invoicemgmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	invoice_common "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/common"

	"github.com/pkg/errors"
)

type BillingAddressHistoryInfo struct {
	PayerName        string `json:"payer_name"`
	PayerPhoneNumber string `json:"payer_phone_number"`
	PostalCode       string `json:"postal_code"`
	PrefectureCode   string `json:"prefecture_code"`
	City             string `json:"city"`
	Street1          string `json:"street1"`
	Street2          string `json:"street2"`
}

// IsVerified is updated to StudentPayment Detail table
type BankAccountHistoryInfo struct {
	IsVerified        bool   `json:"is_verified"`
	BankID            string `json:"bank_id"`
	BankBranchID      string `json:"bank_branch_id"`
	BankAccountNumber string `json:"bank_account_number"`
	BankAccountHolder string `json:"bank_account_holder"`
	BankAccountType   string `json:"bank_account_type"`
}
type PreviousDataStudentActionDetailLog struct {
	BillingAddress *BillingAddressHistoryInfo `json:"billing_address"`
	BankAccount    *BankAccountHistoryInfo    `json:"bank_account"`
	PaymentMethod  string                     `json:"payment_method"`
}

type NewDataStudentActionDetailLog struct {
	BillingAddress *BillingAddressHistoryInfo `json:"billing_address"`
	BankAccount    *BankAccountHistoryInfo    `json:"bank_account"`
	PaymentMethod  string                     `json:"payment_method"`
}

type StudentPaymentActionDetailLogType struct {
	Previous *PreviousDataStudentActionDetailLog `json:"previous"`
	New      *NewDataStudentActionDetailLog      `json:"new"`
}

func (s *suite) validateStudentPaymentDetailActionLogDetail(ctx context.Context, result *StudentPaymentActionDetailLogType, updateAction string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentPaymentDetail, err := (&repositories.StudentPaymentDetailRepo{}).FindByStudentID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.StudentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "StudentPaymentDetailRepo.FindByID")
	}

	switch updateAction {
	case invoice_common.StudentPaymentDetailAction_UPDATED_BILLING_AND_BANK_DETAILS.String():
		if result.New.BillingAddress == nil || result.New.BankAccount == nil {
			return StepStateToContext(ctx, stepState), errors.New("error action detail should contain both billing address and bank account information")
		}

		ctx, err = s.validateBillingAddressActionDetail(ctx, studentPaymentDetail, result)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		ctx, err = s.validateBankAccountActionDetail(ctx, result, studentPaymentDetail.PaymentMethod.String)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

	case invoice_common.StudentPaymentDetailAction_UPDATED_BILLING_DETAILS.String():
		if result.New.BankAccount != nil {
			return StepStateToContext(ctx, stepState), errors.New("error action detail should contain billing address information only but has bank account")
		}

		ctx, err = s.validateBillingAddressActionDetail(ctx, studentPaymentDetail, result)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case invoice_common.StudentPaymentDetailAction_UPDATED_BANK_DETAILS.String():
		if result.New.BillingAddress != nil {
			return StepStateToContext(ctx, stepState), errors.New("error action detail should contain bank account information only but has billing address")
		}

		ctx, err = s.validateBankAccountActionDetail(ctx, result, studentPaymentDetail.PaymentMethod.String)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case invoice_common.StudentPaymentDetailAction_UPDATED_PAYMENT_METHOD.String():
		if result.New.BillingAddress != nil && result.New.BankAccount != nil {
			return StepStateToContext(ctx, stepState), errors.New("error action detail should contain payment method only but has billing address and bank account")
		}

		ctx, err = s.validateStudentPaymentMethodActionLog(ctx, studentPaymentDetail, result)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	default:
		return StepStateToContext(ctx, stepState), errors.New("invalid update info on student payment")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validateBillingAddressActionDetail(ctx context.Context, studentPaymentDetail *entities.StudentPaymentDetail, result *StudentPaymentActionDetailLogType) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	billingAddress, err := (&repositories.BillingAddressRepo{}).FindByUserID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.StudentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "BillingAddressRepo.FindByID")
	}

	// validate billing address for student payment
	if strings.TrimSpace(result.New.BillingAddress.PayerName) != "" && result.New.BillingAddress.PayerName != studentPaymentDetail.PayerName.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error expected billing address payer name: %v but got: %v on action detail", studentPaymentDetail.PayerName.String, result.New.BillingAddress.PayerName)
	}

	if strings.TrimSpace(result.New.BillingAddress.PayerPhoneNumber) != "" && result.New.BillingAddress.PayerPhoneNumber != studentPaymentDetail.PayerPhoneNumber.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error expected billing address payer phone number: %v but got: %v on action detail", studentPaymentDetail.PayerPhoneNumber.String, result.New.BillingAddress.PayerPhoneNumber)
	}

	if strings.TrimSpace(result.New.BillingAddress.PostalCode) != "" && result.New.BillingAddress.PostalCode != billingAddress.PostalCode.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error expected billing address postal code: %v but got: %v on action detail", billingAddress.PostalCode.String, result.New.BillingAddress.PostalCode)
	}

	if strings.TrimSpace(result.New.BillingAddress.PrefectureCode) != "" && result.New.BillingAddress.PrefectureCode != billingAddress.PrefectureCode.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error expected billing address prefecture code: %v but got: %v on action detail", billingAddress.PrefectureCode.String, result.New.BillingAddress.PrefectureCode)
	}

	if strings.TrimSpace(result.New.BillingAddress.City) != "" && result.New.BillingAddress.City != billingAddress.City.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error expected billing address city: %v but got: %v on action detail", billingAddress.City.String, result.New.BillingAddress.City)
	}

	if strings.TrimSpace(result.New.BillingAddress.Street1) != "" && result.New.BillingAddress.Street1 != billingAddress.Street1.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error expected billing address street1: %v but got: %v on action detail", billingAddress.Street1.String, result.New.BillingAddress.Street1)
	}

	if strings.TrimSpace(result.New.BillingAddress.Street2) != "" && result.New.BillingAddress.Street2 != billingAddress.Street2.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error expected billing address street2: %v but got: %v on action detail", billingAddress.Street2.String, result.New.BillingAddress.Street2)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validateStudentPaymentMethodActionLog(ctx context.Context, studentPaymentDetail *entities.StudentPaymentDetail, result *StudentPaymentActionDetailLogType) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if studentPaymentDetail.PaymentMethod.String != result.New.PaymentMethod {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error expected payment method: %v but got: %v on action detail", studentPaymentDetail.PaymentMethod.String, result.New.PaymentMethod)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validateBankAccountActionDetail(ctx context.Context, result *StudentPaymentActionDetailLogType, paymentMethod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	bankAccount, err := (&repositories.BankAccountRepo{}).FindByStudentID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.StudentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "BankAccountRepo.FindByStudentID")
	}

	if result.New.BankAccount.IsVerified != bankAccount.IsVerified.Bool {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error expected bank account is verified: %v but got: %v on action detail", bankAccount.IsVerified.Bool, result.New.BankAccount.IsVerified)
	} else if result.New.PaymentMethod != result.Previous.PaymentMethod && result.New.PaymentMethod != paymentMethod {
		// there is an update on is verified on bank account and the expected payment method is not the same with action detail payment method
		return StepStateToContext(ctx, stepState), fmt.Errorf("error expected payment method: %v but got: %v on action detail", paymentMethod, result.New.PaymentMethod)
	}

	if strings.TrimSpace(result.New.BankAccount.BankID) != "" && result.New.BankAccount.BankID != bankAccount.BankID.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error expected bank account bank id: %v but got: %v on action detail", bankAccount.BankID.String, result.New.BankAccount.BankID)
	}

	if strings.TrimSpace(result.New.BankAccount.BankBranchID) != "" && result.New.BankAccount.BankBranchID != bankAccount.BankBranchID.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error expected bank account bank branch id: %v but got: %v on action detail", bankAccount.BankBranchID.String, result.New.BankAccount.BankBranchID)
	}

	if strings.TrimSpace(result.New.BankAccount.BankAccountNumber) != "" && result.New.BankAccount.BankAccountNumber != bankAccount.BankAccountNumber.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error expected bank account bank account number: %v but got: %v on action detail", bankAccount.BankAccountNumber.String, result.New.BankAccount.BankAccountNumber)
	}

	if strings.TrimSpace(result.New.BankAccount.BankAccountHolder) != "" && result.New.BankAccount.BankAccountHolder != bankAccount.BankAccountHolder.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error expected bank account bank holder: %v but got: %v on action detail", bankAccount.BankAccountHolder.String, result.New.BankAccount.BankAccountHolder)
	}

	if strings.TrimSpace(result.New.BankAccount.BankAccountType) != "" && result.New.BankAccount.BankAccountType != bankAccount.BankAccountType.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error expected bank account type: %v but got: %v on action detail", bankAccount.BankAccountType.String, result.New.BankAccount.BankAccountType)
	}

	return StepStateToContext(ctx, stepState), nil
}
