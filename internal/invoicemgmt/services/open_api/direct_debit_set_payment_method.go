package openapisvc

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	http_util "github.com/manabie-com/backend/internal/invoicemgmt/services/http"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/http/errcode"
	utils "github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_common "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/common"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func (s *OpenAPIModifierService) UpsertStudentBankAccountInfo(c *gin.Context) {
	isSetDirectDebitOpenAPI, err := s.UnleashClient.IsFeatureEnabled(constant.EnableSetDirectDebitFeatureFlag, s.Env)
	if err != nil {
		http_util.ResponseError(c, errcode.Error{
			Code: errcode.InternalError,
			Err:  fmt.Errorf("%v unleashClient.IsFeatureEnabled err: %v", constant.EnableSetDirectDebitFeatureFlag, err),
		})
		return
	}
	if !isSetDirectDebitOpenAPI {
		http_util.ResponseError(c, errcode.Error{
			Code: errcode.PermissionDenied,
			Err:  errors.New("permission denied"),
		})
	}
	var req UpsertStudentBankRequestInfo
	if err := http_util.ParseJSONPayload(c.Request, &req); err != nil {
		http_util.ResponseError(c, err)
		return
	}
	// validate payload request fields if existing
	err = s.checkRequestPayloadFields(req.StudentBankRequestInfo)
	if err != nil {
		http_util.ResponseError(c, err)
		return
	}
	// validate bank account format
	err = s.validateBankAccountFormat(req.StudentBankRequestInfo)
	if err != nil {
		http_util.ResponseError(c, err)
		return
	}
	// validate bank account data on system
	validBankAccountInfo, err := s.validateBankAccountData(c.Request.Context(), req.StudentBankRequestInfo)
	if err != nil {
		http_util.ResponseError(c, err)
		return
	}
	// billing address should be existing
	err = s.validateBillingAddress(c.Request.Context(), validBankAccountInfo.StudentID)
	if err != nil {
		http_util.ResponseError(c, err)
		return
	}

	// get the student payment detail record
	studentPaymentDetailToUpsert, err := s.getStudentPaymentDetailRecordToUpsert(c.Request.Context(), validBankAccountInfo.StudentID)
	if err != nil {
		http_util.ResponseError(c, err)
		return
	}

	validBankAccountInfo.StudentPaymentDetailID = studentPaymentDetailToUpsert.StudentPaymentDetailID.String

	studentPaymentActionLogBankAccount := genActionLogInfoWithBankAccount()
	bankAccountInfoToUpsert, studentPaymentActionLogBankAccount, err := s.getBankAccountRecordToUpsert(c.Request.Context(), validBankAccountInfo, studentPaymentDetailToUpsert, studentPaymentActionLogBankAccount)
	if err != nil {
		http_util.ResponseError(c, err)
		return
	}

	err = database.ExecInTx(c.Request.Context(), s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		// generate student payment detail to upsert from bank info
		studentPaymentDetailToUpsert, err := s.generateStudentPaymentDetailFromBankInfo(studentPaymentDetailToUpsert, bankAccountInfoToUpsert.IsVerified.Bool)
		if err != nil {
			return errcode.Error{
				Code: errcode.InternalError,
				Err:  err,
			}
		}

		if err := s.StudentPaymentDetailRepo.Upsert(ctx, tx, studentPaymentDetailToUpsert); err != nil {
			return errcode.Error{
				Code: errcode.InternalError,
				Err:  errors.Wrap(err, "s.StudentPaymentDetailRepo.Upsert"),
			}
		}

		if err := s.BankAccountRepo.Upsert(ctx, tx, bankAccountInfoToUpsert); err != nil {
			return errcode.Error{
				Code: errcode.InternalError,
				Err:  errors.Wrap(err, "s.BankAccountRepo.Upsert"),
			}
		}

		if reflect.DeepEqual(studentPaymentActionLogBankAccount, genActionLogInfoWithBankAccount()) {
			return nil
		}

		// Create action log
		err = s.createPaymentDetailActionLog(ctx, tx, studentPaymentDetailToUpsert.StudentPaymentDetailID.String, invoice_common.StudentPaymentDetailAction_UPDATED_BANK_DETAILS.String(), studentPaymentActionLogBankAccount)
		if err != nil {
			return errcode.Error{
				Code: errcode.InternalError,
				Err:  err,
			}
		}

		return nil
	})
	if err != nil {
		http_util.ResponseError(c, err)
		return
	}

	data := map[string]interface{}{
		"student_id":       bankAccountInfoToUpsert.StudentID.String,
		"external_user_id": req.StudentBankRequestInfo.ExternalUserID,
	}

	c.JSON(http.StatusOK, http_util.ResponseSuccess(data))
}

func (s *OpenAPIModifierService) validateBankAccountFormat(studentBankRequestInfo StudentBankRequestInfoProfile) error {
	// return error right away if there's a empty mandatory field
	if strings.TrimSpace(studentBankRequestInfo.ExternalUserID.String) == "" {
		return errcode.Error{
			FieldName: "external_user_id",
			Code:      errcode.MissingMandatory,
		}
	}
	// validate all fields if verified is set to true
	switch studentBankRequestInfo.IsVerified.Bool {
	case true:
		// validate first if fields have empty values and return missing mandatory error code
		err := s.validateEmptyFields(studentBankRequestInfo)
		if err != nil {
			return err
		}
		// return invalid data format error code
		err = s.validateDataFormat(studentBankRequestInfo)
		if err != nil {
			return err
		}
	default:
		// catch when not verified to validate bank branch code when value is existing and bank code value is not
		if strings.TrimSpace(studentBankRequestInfo.BankBranchCode.String) != "" && strings.TrimSpace(studentBankRequestInfo.BankCode.String) == "" {
			return errcode.Error{
				FieldName: "bank_code",
				Code:      errcode.MissingMandatory,
			}
		}
	}

	// Validate account number
	if strings.TrimSpace(studentBankRequestInfo.BankAccountNumber.String) != "" {
		numericRegex := regexp.MustCompile(constant.NumericRegex)
		if !numericRegex.MatchString(studentBankRequestInfo.BankAccountNumber.String) {
			return errcode.Error{
				FieldName: "bank_account_number",
				Code:      errcode.InvalidData,
			}
		}
	}

	// Validate bank account type
	if invoice_pb.BankAccountType_name[studentBankRequestInfo.BankAccountType.Int] == "" {
		return errcode.Error{
			FieldName: "bank_account_type",
			Code:      errcode.InvalidData,
		}
	}

	return nil
}

func (s *OpenAPIModifierService) checkRequestPayloadFields(studentBankRequestInfo StudentBankRequestInfoProfile) error {
	reflectVal := reflect.ValueOf(studentBankRequestInfo)
	for i := 0; i < reflectVal.NumField(); i++ {
		varValue := reflectVal.Field(i).Interface()
		jsonField := reflectVal.Type().Field(i).Tag.Get("json")
		fieldStatus := castFieldStatusFromInterface(varValue)
		if fieldStatus != pgtype.Present {
			return errcode.Error{
				FieldName: jsonField,
				Code:      errcode.MissingMandatory,
			}
		}
	}

	return nil
}

func (s *OpenAPIModifierService) validateEmptyFields(studentBankRequestInfo StudentBankRequestInfoProfile) error {
	if strings.TrimSpace(studentBankRequestInfo.BankCode.String) == "" {
		return errcode.Error{
			FieldName: "bank_code",
			Code:      errcode.MissingMandatory,
		}
	}

	if strings.TrimSpace(studentBankRequestInfo.BankBranchCode.String) == "" {
		return errcode.Error{
			FieldName: "bank_branch_code",
			Code:      errcode.MissingMandatory,
		}
	}

	if strings.TrimSpace(studentBankRequestInfo.BankAccountNumber.String) == "" {
		return errcode.Error{
			FieldName: "bank_account_number",
			Code:      errcode.MissingMandatory,
		}
	}

	if strings.TrimSpace(studentBankRequestInfo.BankAccountHolder.String) == "" {
		return errcode.Error{
			FieldName: "bank_account_holder",
			Code:      errcode.MissingMandatory,
		}
	}
	// bank account type and is verified are already handle when not present and invalid data type
	return nil
}

func (s *OpenAPIModifierService) validateDataFormat(studentBankRequestInfo StudentBankRequestInfoProfile) error {
	if len(strings.TrimSpace(studentBankRequestInfo.BankAccountNumber.String)) != 7 {
		return errcode.Error{
			FieldName: "bank_account_number",
			Code:      errcode.InvalidData,
		}
	}

	if err := utils.ValidateBankHolder(studentBankRequestInfo.BankAccountHolder.String); err != nil {
		return errcode.Error{
			FieldName: "bank_account_holder",
			Code:      errcode.InvalidData,
		}
	}

	return nil
}

func (s *OpenAPIModifierService) validateBankAccountData(ctx context.Context, studentBankRequestInfo StudentBankRequestInfoProfile) (*BankAccountInfo, error) {
	var (
		bank       *entities.Bank
		bankBranch *entities.BankBranch
	)
	// check if external user is existing on db
	user, err := s.UserRepo.FindByUserExternalID(ctx, s.DB, studentBankRequestInfo.ExternalUserID.String)
	if err != nil {
		return nil, errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "s.UserRepo.FindByUserExternalID"),
		}
	}

	// validate bank if existing whether verified or unverified if there's value
	if strings.TrimSpace(studentBankRequestInfo.BankCode.String) != "" {
		bank, err = s.BankRepo.FindByBankCode(ctx, s.DB, studentBankRequestInfo.BankCode.String)
		if err != nil {
			return nil, errcode.Error{
				Code: errcode.InternalError,
				Err:  errors.Wrap(err, "s.BankRepo.FindByBankCode"),
			}
		}
	}
	// validate bank branch code if existing whether verified or unverified if there's value
	if strings.TrimSpace(studentBankRequestInfo.BankBranchCode.String) != "" && bank != nil {
		bankBranch, err = s.BankBranchRepo.FindByBankBranchCodeAndBank(ctx, s.DB, studentBankRequestInfo.BankBranchCode.String, bank.BankID.String)
		if err != nil {
			return nil, errcode.Error{
				Code: errcode.InternalError,
				Err:  errors.Wrap(err, "s.BankBranchRepo.FindByBankBranchCodeAndBank"),
			}
		}
	}

	bankAccountInfo := &BankAccountInfo{
		StudentID:         user.UserID.String,
		IsVerified:        studentBankRequestInfo.IsVerified.Bool,
		BankAccountNumber: studentBankRequestInfo.BankAccountNumber.String,
		BankAccountHolder: studentBankRequestInfo.BankAccountHolder.String,
		BankAccountType:   invoice_pb.BankAccountType_name[studentBankRequestInfo.BankAccountType.Int],
		BankID:            "",
		BankBranchID:      "",
	}
	if bank != nil {
		bankAccountInfo.BankID = bank.BankID.String
	}
	if bankBranch != nil {
		bankAccountInfo.BankBranchID = bankBranch.BankBranchID.String
	}

	return bankAccountInfo, nil
}

func (s *OpenAPIModifierService) validateBillingAddress(ctx context.Context, studentID string) error {
	_, err := s.BillingAddressRepo.FindByUserID(ctx, s.DB, studentID)
	if err != nil {
		return errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "s.BillingAddressRepo.FindByUserID"),
		}
	}

	return nil
}

func (s *OpenAPIModifierService) getStudentPaymentDetailRecordToUpsert(ctx context.Context, studentID string) (*entities.StudentPaymentDetail, error) {
	studentPaymentDetail, err := s.StudentPaymentDetailRepo.FindByStudentID(ctx, s.DB, studentID)
	if err != nil {
		return nil, errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "s.BillingAddressRepo.FindByUserID"),
		}
	}

	return studentPaymentDetail, nil
}

func (s *OpenAPIModifierService) getBankAccountRecordToUpsert(
	ctx context.Context,
	bankAccountInfo *BankAccountInfo,
	studentPaymentDetail *entities.StudentPaymentDetail,
	studentPaymentActionLogBankAccount *StudentPaymentActionDetailLogType,
) (*entities.BankAccount, *StudentPaymentActionDetailLogType, error) {
	bankAccount, err := s.BankAccountRepo.FindByStudentID(ctx, s.DB, bankAccountInfo.StudentID)
	now := time.Now()
	switch err {
	case nil:
		if bankAccount.StudentPaymentDetailID.String != bankAccountInfo.StudentPaymentDetailID {
			return nil, nil, errcode.Error{
				Code: errcode.InternalError,
				Err:  errors.Wrap(err, fmt.Sprintf("existing bank account student payment detail got: %v expected: %v", bankAccount.StudentPaymentDetailID.String, bankAccountInfo.StudentPaymentDetailID)),
			}
		}

		studentPaymentActionLogBankAccount = setActionLogFromBankAccountAndEvent(bankAccount, bankAccountInfo, studentPaymentDetail.PaymentMethod.String, studentPaymentActionLogBankAccount)

		// update the existing bank account fields
		if err := multierr.Combine(
			bankAccount.IsVerified.Set(bankAccountInfo.IsVerified),
			bankAccount.BankID.Set(bankAccountInfo.BankID),
			bankAccount.BankBranchID.Set(bankAccountInfo.BankBranchID),
			bankAccount.BankAccountNumber.Set(bankAccountInfo.BankAccountNumber),
			bankAccount.BankAccountHolder.Set(bankAccountInfo.BankAccountHolder),
			bankAccount.BankAccountType.Set(bankAccountInfo.BankAccountType),
			bankAccount.UpdatedAt.Set(now),
		); err != nil {
			return nil, nil, errcode.Error{
				Code: errcode.InternalError,
				Err:  errors.Wrap(err, fmt.Sprintf("update existing bank account fields multierr.Combine: %v", err)),
			}
		}

	case pgx.ErrNoRows:
		// create new bank account entity
		bankAccount = new(entities.BankAccount)
		database.AllNullEntity(bankAccount)

		if err := multierr.Combine(
			bankAccount.BankAccountID.Set(database.Text(idutil.ULIDNow())),
			bankAccount.StudentPaymentDetailID.Set(bankAccountInfo.StudentPaymentDetailID),
			bankAccount.StudentID.Set(bankAccountInfo.StudentID),
			bankAccount.IsVerified.Set(bankAccountInfo.IsVerified),
			bankAccount.BankID.Set(bankAccountInfo.BankID),
			bankAccount.BankBranchID.Set(bankAccountInfo.BankBranchID),
			bankAccount.BankAccountNumber.Set(bankAccountInfo.BankAccountNumber),
			bankAccount.BankAccountHolder.Set(bankAccountInfo.BankAccountHolder),
			bankAccount.BankAccountType.Set(bankAccountInfo.BankAccountType),
			bankAccount.CreatedAt.Set(now),
			bankAccount.UpdatedAt.Set(now),
		); err != nil {
			return nil, nil, errcode.Error{
				Code: errcode.InternalError,
				Err:  errors.Wrap(err, fmt.Sprintf("create new bank account entity multierr.Combine: %v", err)),
			}
		}

	default:
		return nil, nil, errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "s.BankAccountRepo.FindByStudentID"),
		}
	}

	return bankAccount, studentPaymentActionLogBankAccount, nil
}

func (s *OpenAPIModifierService) generateStudentPaymentDetailFromBankInfo(studentPaymentDetail *entities.StudentPaymentDetail, isVerified bool) (*entities.StudentPaymentDetail, error) {
	now := time.Now()
	var paymentMethod string

	switch isVerified {
	case true:
		paymentMethod = invoice_pb.PaymentMethod_DIRECT_DEBIT.String()
	default:
		paymentMethod = invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()
	}

	if err := multierr.Combine(
		studentPaymentDetail.PaymentMethod.Set(paymentMethod),
		studentPaymentDetail.UpdatedAt.Set(now),
	); err != nil {
		return nil, errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, fmt.Sprintf("generate student payment detail from bank info multierr.Combine: %v", err)),
		}
	}

	return studentPaymentDetail, nil
}

func genActionLogInfoWithBankAccount() *StudentPaymentActionDetailLogType {
	return &StudentPaymentActionDetailLogType{
		Previous: &PreviousDataStudentActionDetailLog{
			BankAccount: &BankAccountHistoryInfo{},
		},
		New: &NewDataStudentActionDetailLog{
			BankAccount: &BankAccountHistoryInfo{},
		},
	}
}

func setActionLogFromBankAccountAndEvent(
	existingBankAccount *entities.BankAccount,
	bankAccountInfo *BankAccountInfo,
	existingPaymentMethod string,
	studentPaymentActionLogBankAccount *StudentPaymentActionDetailLogType,
) *StudentPaymentActionDetailLogType {
	if existingBankAccount.IsVerified.Bool != bankAccountInfo.IsVerified {
		studentPaymentActionLogBankAccount.Previous.BankAccount.IsVerified = existingBankAccount.IsVerified.Bool
		studentPaymentActionLogBankAccount.New.BankAccount.IsVerified = bankAccountInfo.IsVerified
		studentPaymentActionLogBankAccount.Previous.PaymentMethod = existingPaymentMethod
		newPaymentMethod := invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()

		if bankAccountInfo.IsVerified {
			newPaymentMethod = invoice_pb.PaymentMethod_DIRECT_DEBIT.String()
		}

		studentPaymentActionLogBankAccount.New.PaymentMethod = newPaymentMethod
	}

	if existingBankAccount.BankID.String != bankAccountInfo.BankID {
		studentPaymentActionLogBankAccount.Previous.BankAccount.BankID = existingBankAccount.BankID.String
		studentPaymentActionLogBankAccount.New.BankAccount.BankID = bankAccountInfo.BankID
	}

	if existingBankAccount.BankBranchID.String != bankAccountInfo.BankBranchID {
		studentPaymentActionLogBankAccount.Previous.BankAccount.BankBranchID = existingBankAccount.BankBranchID.String
		studentPaymentActionLogBankAccount.New.BankAccount.BankBranchID = bankAccountInfo.BankBranchID
	}

	if existingBankAccount.BankAccountNumber.String != bankAccountInfo.BankAccountNumber {
		studentPaymentActionLogBankAccount.Previous.BankAccount.BankAccountNumber = existingBankAccount.BankAccountNumber.String
		studentPaymentActionLogBankAccount.New.BankAccount.BankAccountNumber = bankAccountInfo.BankAccountNumber
	}

	if existingBankAccount.BankAccountHolder.String != bankAccountInfo.BankAccountHolder {
		studentPaymentActionLogBankAccount.Previous.BankAccount.BankAccountHolder = existingBankAccount.BankAccountHolder.String
		studentPaymentActionLogBankAccount.New.BankAccount.BankAccountHolder = bankAccountInfo.BankAccountHolder
	}

	if existingBankAccount.BankAccountType.String != bankAccountInfo.BankAccountType {
		studentPaymentActionLogBankAccount.Previous.BankAccount.BankAccountType = existingBankAccount.BankAccountType.String
		studentPaymentActionLogBankAccount.New.BankAccount.BankAccountType = bankAccountInfo.BankAccountType
	}

	return studentPaymentActionLogBankAccount
}
