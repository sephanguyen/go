package openapisvc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BillingAddressInfo struct {
	StudentID    string
	PayerName    string
	PostalCode   string
	PrefectureID string
	City         string
	Street1      string
	Street2      string
}

type BankAccountInfo struct {
	StudentPaymentDetailID string
	StudentID              string
	IsVerified             bool
	BankID                 string
	BankBranchID           string
	BankAccountNumber      string
	BankAccountHolder      string
	BankAccountType        string
}

// PayerName and PayerPhoneNumber are updated to StudentPayment Detail table
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

type studentPaymentDetailActionLogData struct {
	actionDetailInfo       pgtype.JSONB
	action                 string
	StudentPaymentDetailID string
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

const (
	AutoSetConvenienceStoreConfigKey = "invoice.invoicemgmt.enable_auto_default_convenience_store"
	numberOfBankInfoFieldsPayload    = 7
)

type UpsertStudentBankRequestInfo struct {
	StudentBankRequestInfo StudentBankRequestInfoProfile `json:"student_bank_info"`
}

type StudentBankRequestInfoProfile struct {
	ExternalUserID    pgtype.Text `json:"external_user_id"`
	BankCode          pgtype.Text `json:"bank_code"`
	BankBranchCode    pgtype.Text `json:"bank_branch_code"`
	BankAccountNumber pgtype.Text `json:"bank_account_number"`
	BankAccountHolder pgtype.Text `json:"bank_account_holder"`
	BankAccountType   pgtype.Int4 `json:"bank_account_type"`
	IsVerified        pgtype.Bool `json:"is_verified"`
}

var PartnerBankDepositItems = map[int]string{
	1: "ORDINARY_BANK_ACCOUNT",
	2: "CHECKING_ACCOUNT",
}

func checkPostgresSQLErrorCode(err error) (bool, error) {
	pgerr, ok := errors.Unwrap(err).(*pgconn.PgError)

	if ok {
		switch pgerr.Code {
		case "23503":
			return true, errors.New(constant.PgConnForeignKeyError)
		case "42501":
			return true, errors.New(constant.TableRLSError)
		}
	}
	return false, nil
}

func castFieldStatusFromInterface(i interface{}) pgtype.Status {
	var fieldStatus pgtype.Status
	switch field := i.(type) {
	case pgtype.Text:
		fieldStatus = field.Status
	case pgtype.Int4:
		fieldStatus = field.Status
	case pgtype.Bool:
		fieldStatus = field.Status
	}

	return fieldStatus
}

func hasEmptyField(values []string) bool {
	for _, v := range values {
		if strings.TrimSpace(v) == "" {
			return true
		}
	}

	return false
}

func validateUpdateBillingAddressEventInfo(info *UpdateBillingAddressEventInfo) error {
	switch {
	case strings.TrimSpace(info.StudentID) == "":
		return errors.New("student id cannot be empty")
	case strings.TrimSpace(info.PayerName) == "":
		return errors.New("payer name cannot be empty")
	}

	return nil
}

func generateStudentPaymentDetailActionLog(ctx context.Context, studentPaymentDetailActionLogDataInfo *studentPaymentDetailActionLogData) (*entities.StudentPaymentDetailActionLog, error) {
	studentPaymentDetailActionLog := new(entities.StudentPaymentDetailActionLog)
	database.AllNullEntity(studentPaymentDetailActionLog)

	// get user id from ctx
	userInfo := golibs.UserInfoFromCtx(ctx)

	if err := multierr.Combine(
		studentPaymentDetailActionLog.StudentPaymentDetailID.Set(studentPaymentDetailActionLogDataInfo.StudentPaymentDetailID),
		studentPaymentDetailActionLog.UserID.Set(userInfo.UserID),
		studentPaymentDetailActionLog.Action.Set(studentPaymentDetailActionLogDataInfo.action),
		studentPaymentDetailActionLog.ActionDetail.Set(studentPaymentDetailActionLogDataInfo.actionDetailInfo),
	); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("err cannot generate student payment detail action log entity: %v", err))
	}

	return studentPaymentDetailActionLog, nil
}
