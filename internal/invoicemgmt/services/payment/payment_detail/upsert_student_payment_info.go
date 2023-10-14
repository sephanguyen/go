package paymentdetail

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_common "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/common"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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

// nolint
func (s *EditPaymentDetailService) UpsertStudentPaymentInfo(ctx context.Context, request *invoice_pb.UpsertStudentPaymentInfoRequest) (*invoice_pb.UpsertStudentPaymentInfoResponse, error) {
	switch {
	case strings.TrimSpace(request.StudentId) == "":
		return nil, status.Error(codes.InvalidArgument, "student id to upsert payment info can not be empty")
	case request.BillingInfo == nil:
		if request.BankAccountInfo == nil {
			return nil, status.Error(codes.InvalidArgument, "both billing information and bank account can not be empty")
		}
	case request.BillingInfo.BillingAddress == nil && request.BankAccountInfo == nil:
		return nil, status.Error(codes.InvalidArgument, "both billing information and bank account can not be empty")
	}

	err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		// When billing information exists in request
		if request.BillingInfo != nil && request.BillingInfo.BillingAddress != nil {
			// Validate billing information in request
			studentPaymentDetail, billingAddress, studentPaymentActionLogBillingInfo, err := s.ValidateBillingInformationRequest(ctx, request)
			if err != nil {
				return err
			}

			// If studentPaymentDetail and billingAddress are valid, and this is the first time create
			// billing info then payment method should be updated to "CONVENIENCE_STORE"
			if studentPaymentDetail.StudentPaymentDetailID.String == "" {
				studentPaymentDetail.StudentPaymentDetailID = database.Text(idutil.ULIDNow())
				studentPaymentDetail.PaymentMethod = database.Text(entities.PaymentMethodConvenienceStore)
			}

			// Upsert student payment detail
			if err := s.StudentPaymentDetailRepo.Upsert(ctx, tx, studentPaymentDetail); err != nil {
				return err
			}
			// Upsert billing address
			billingAddress.StudentPaymentDetailID = studentPaymentDetail.StudentPaymentDetailID
			if err := s.BillingAddressRepo.Upsert(ctx, tx, billingAddress); err != nil {
				return err
			}
			// create action log if billing address has changes
			if studentPaymentActionLogBillingInfo != nil && request.BankAccountInfo == nil {
				// create action log for billing address only
				studentPaymentDetailActionLog, err := generateStudentPaymentDetailActionLog(ctx, &studentPaymentDetailActionLogData{
					actionDetailInfo:       database.JSONB(studentPaymentActionLogBillingInfo),
					action:                 invoice_common.StudentPaymentDetailAction_UPDATED_BILLING_DETAILS.String(),
					StudentPaymentDetailID: studentPaymentDetail.StudentPaymentDetailID.String,
				})
				if err != nil {
					return err
				}

				if err := s.StudentPaymentDetailActionLogRepo.Create(ctx, tx, studentPaymentDetailActionLog); err != nil {
					return status.Error(codes.Internal, fmt.Sprintf("err cannot create student payment detail action log: %v", err))
				}
			}
			// When bank account exists together billing information in request
			if request.BankAccountInfo != nil {
				// Valid bank account in request
				bankAccountToUpsert, existingBankAccount, err := s.ValidateBankAccountRequest(ctx, request)
				if err != nil {
					return err
				}

				// Upsert bank account
				bankAccountToUpsert.StudentPaymentDetailID = studentPaymentDetail.StudentPaymentDetailID
				if err := s.BankAccountRepo.Upsert(ctx, tx, bankAccountToUpsert); err != nil {
					return err
				}

				if existingBankAccount != nil {
					// create action log for both billing address and bank account
					if studentPaymentActionBankInfo := retrieveBankAccountFieldsToUpdate(existingBankAccount, bankAccountToUpsert, studentPaymentDetail.PaymentMethod.String); studentPaymentActionBankInfo != nil {
						studentPaymentAction := invoice_common.StudentPaymentDetailAction_UPDATED_BANK_DETAILS.String()
						studentPaymentActionLogDetailType := StudentPaymentActionDetailLogType{
							&PreviousDataStudentActionDetailLog{
								BankAccount:   studentPaymentActionBankInfo.Previous.BankAccount,
								PaymentMethod: studentPaymentActionBankInfo.Previous.PaymentMethod,
							},
							&NewDataStudentActionDetailLog{
								BankAccount:   studentPaymentActionBankInfo.New.BankAccount,
								PaymentMethod: studentPaymentActionBankInfo.New.PaymentMethod,
							},
						}
						if studentPaymentActionLogBillingInfo != nil {
							studentPaymentAction = invoice_common.StudentPaymentDetailAction_UPDATED_BILLING_AND_BANK_DETAILS.String()
							// add billing address action log when both billing and bank account exist
							studentPaymentActionLogDetailType.New.BillingAddress = studentPaymentActionLogBillingInfo.New.BillingAddress
							studentPaymentActionLogDetailType.Previous.BillingAddress = studentPaymentActionLogBillingInfo.Previous.BillingAddress
						}

						if strings.TrimSpace(studentPaymentActionLogDetailType.New.PaymentMethod) != "" {
							studentPaymentAction = invoice_common.StudentPaymentDetailAction_UPDATED_PAYMENT_METHOD.String()
						}

						studentPaymentDetailActionLog, err := generateStudentPaymentDetailActionLog(ctx, &studentPaymentDetailActionLogData{
							actionDetailInfo:       database.JSONB(studentPaymentActionLogDetailType),
							action:                 studentPaymentAction,
							StudentPaymentDetailID: bankAccountToUpsert.StudentPaymentDetailID.String,
						})
						if err != nil {
							return err
						}

						if err := s.StudentPaymentDetailActionLogRepo.Create(ctx, tx, studentPaymentDetailActionLog); err != nil {
							return status.Error(codes.Internal, fmt.Sprintf("err cannot create student payment detail action log: %v", err))
						}
					}
				}

				// no need to update payment method if existing bank account is already verified and the IsVerified in request is true
				if existingBankAccount != nil && existingBankAccount.IsVerified.Bool && request.BankAccountInfo.IsVerified {
					return nil
				}

				paymentMethod := database.Text(entities.PaymentMethodConvenienceStore)
				if bankAccountToUpsert.IsVerified.Bool {
					paymentMethod = database.Text(entities.PaymentMethodDirectDebit)
				}

				// After upsert bank account successfully, payment method of payment detail should be updated to "DIRECT_DEBIT"
				studentPaymentDetail.PaymentMethod = paymentMethod
				if err := s.StudentPaymentDetailRepo.Upsert(ctx, tx, studentPaymentDetail); err != nil {
					return err
				}
			}
		} else if request.BankAccountInfo != nil {
			// When billing information does not exist, but bank account exists in request
			// Then check this student already has billing information or not
			bankAccountToUpsert, existingBankAccount, err := s.ValidateBankAccountRequest(ctx, request)
			if err != nil {
				return err
			}

			existingStudentPaymentDetail, _, err := FindExistingBillingInfoByStudentID(ctx, s.DB, s.StudentPaymentDetailRepo, s.BillingAddressRepo, request.StudentId)
			if err != nil {
				return err
			}
			bankAccountToUpsert.StudentPaymentDetailID = existingStudentPaymentDetail.StudentPaymentDetailID

			if err := s.BankAccountRepo.Upsert(ctx, tx, bankAccountToUpsert); err != nil {
				return err
			}

			if existingBankAccount != nil {
				// create action log for bank account only
				if studentPaymentActionBankInfo := retrieveBankAccountFieldsToUpdate(existingBankAccount, bankAccountToUpsert, existingStudentPaymentDetail.PaymentMethod.String); studentPaymentActionBankInfo != nil {
					studentPaymentAction := invoice_common.StudentPaymentDetailAction_UPDATED_BANK_DETAILS.String()
					if strings.TrimSpace(studentPaymentActionBankInfo.New.PaymentMethod) != "" {
						studentPaymentAction = invoice_common.StudentPaymentDetailAction_UPDATED_PAYMENT_METHOD.String()
					}
					studentPaymentDetailActionLog, err := generateStudentPaymentDetailActionLog(ctx, &studentPaymentDetailActionLogData{
						actionDetailInfo:       database.JSONB(studentPaymentActionBankInfo),
						action:                 studentPaymentAction,
						StudentPaymentDetailID: bankAccountToUpsert.StudentPaymentDetailID.String,
					})
					if err != nil {
						return err
					}

					if err := s.StudentPaymentDetailActionLogRepo.Create(ctx, tx, studentPaymentDetailActionLog); err != nil {
						return status.Error(codes.Internal, fmt.Sprintf("err cannot create student payment detail action log: %v", err))
					}
				}
			}

			// no need to update payment method if existing bank account is already verified and the IsVerified in request is true
			if existingBankAccount != nil && existingBankAccount.IsVerified.Bool && request.BankAccountInfo.IsVerified {
				return nil
			}

			paymentMethod := database.Text(entities.PaymentMethodConvenienceStore)
			if bankAccountToUpsert.IsVerified.Bool {
				paymentMethod = database.Text(entities.PaymentMethodDirectDebit)
			}

			if existingStudentPaymentDetail.PaymentMethod == paymentMethod {
				return nil
			}

			existingStudentPaymentDetail.PaymentMethod = paymentMethod
			//Upsert student payment detail
			if err := s.StudentPaymentDetailRepo.Upsert(ctx, tx, existingStudentPaymentDetail); err != nil {
				return err
			}

		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &invoice_pb.UpsertStudentPaymentInfoResponse{Successful: true}, nil
}

func UpsertStudentPaymentInfoRequestToStudentPaymentDetailEnt(request *invoice_pb.UpsertStudentPaymentInfoRequest) (*entities.StudentPaymentDetail, error) {
	studentPaymentDetail := new(entities.StudentPaymentDetail)
	database.AllNullEntity(studentPaymentDetail)

	now := time.Now()
	if err := multierr.Combine(
		studentPaymentDetail.StudentID.Set(request.StudentId),
		studentPaymentDetail.StudentPaymentDetailID.Set(request.BillingInfo.StudentPaymentDetailId),
		studentPaymentDetail.PayerName.Set(request.BillingInfo.PayerName),
		studentPaymentDetail.PayerPhoneNumber.Set(request.BillingInfo.PayerPhoneNumber),
		studentPaymentDetail.UpdatedAt.Set(now),
		studentPaymentDetail.DeletedAt.Set(nil),
	); err != nil {
		return nil, fmt.Errorf("multierr.Combine: %w", err)
	}

	if request.BillingInfo.StudentPaymentDetailId == "" {
		if err := studentPaymentDetail.CreatedAt.Set(now); err != nil {
			return nil, errors.Wrap(err, "failed to set value to create_at field for student payment detail")
		}
	}

	return studentPaymentDetail, nil
}

func UpsertStudentPaymentInfoRequestToBillingAddressEnt(request *invoice_pb.UpsertStudentPaymentInfoRequest) (*entities.BillingAddress, error) {
	billingAddress := new(entities.BillingAddress)
	database.AllNullEntity(billingAddress)

	now := time.Now()
	if err := multierr.Combine(
		billingAddress.BillingAddressID.Set(request.BillingInfo.BillingAddress.BillingAddressId),
		billingAddress.StudentPaymentDetailID.Set(request.BillingInfo.StudentPaymentDetailId),
		billingAddress.UserID.Set(request.StudentId),
		billingAddress.PostalCode.Set(request.BillingInfo.BillingAddress.PostalCode),
		billingAddress.PrefectureCode.Set(request.BillingInfo.BillingAddress.PrefectureCode),
		billingAddress.City.Set(request.BillingInfo.BillingAddress.City),
		billingAddress.Street1.Set(request.BillingInfo.BillingAddress.Street1),
		billingAddress.Street2.Set(request.BillingInfo.BillingAddress.Street2),
		billingAddress.UpdatedAt.Set(now),
		billingAddress.DeletedAt.Set(nil),
	); err != nil {
		return nil, fmt.Errorf("multierr.Combine: %w", err)
	}

	if request.BillingInfo.BillingAddress.BillingAddressId == "" {
		if err := billingAddress.CreatedAt.Set(now); err != nil {
			return nil, errors.Wrap(err, "failed to set value to create_at field for billing address")
		}
	}

	return billingAddress, nil
}

func UpsertStudentPaymentInfoRequestToBankAccountEnt(request *invoice_pb.UpsertStudentPaymentInfoRequest) (*entities.BankAccount, error) {
	bankAccount := new(entities.BankAccount)
	database.AllNullEntity(bankAccount)

	now := time.Now()
	if err := multierr.Combine(
		bankAccount.StudentID.Set(request.StudentId),
		bankAccount.BankAccountID.Set(request.BankAccountInfo.BankAccountId),
		bankAccount.BankID.Set(request.BankAccountInfo.BankId),
		bankAccount.BankBranchID.Set(request.BankAccountInfo.BankBranchId),
		bankAccount.BankAccountHolder.Set(request.BankAccountInfo.BankAccountHolder),
		bankAccount.BankAccountNumber.Set(request.BankAccountInfo.BankAccountNumber),
		bankAccount.BankAccountType.Set(invoice_pb.BankAccountType_name[int32(request.BankAccountInfo.BankAccountType)]),
		bankAccount.IsVerified.Set(request.BankAccountInfo.IsVerified),
		bankAccount.UpdatedAt.Set(now),
		bankAccount.DeletedAt.Set(nil),
	); err != nil {
		return nil, fmt.Errorf("multierr.Combine: %w", err)
	}

	if request.BankAccountInfo.BankAccountId == "" {
		if err := bankAccount.CreatedAt.Set(now); err != nil {
			return nil, errors.Wrap(err, "failed to set value to create_at field for billing address")
		}
	}

	return bankAccount, nil
}

func (s *EditPaymentDetailService) ValidateBillingInformationRequest(ctx context.Context, request *invoice_pb.UpsertStudentPaymentInfoRequest) (*entities.StudentPaymentDetail, *entities.BillingAddress, *StudentPaymentActionDetailLogType, error) {
	// Validate StudentPaymentDetail
	studentPaymentDetail, err := UpsertStudentPaymentInfoRequestToStudentPaymentDetailEnt(request)
	if err != nil {
		return nil, nil, nil, status.Error(codes.Internal, fmt.Sprintf("UpsertStudentPaymentInfoRequestToStudentPaymentDetailEnt: %v", err))
	}

	studentPaymentDetailActionLogType, err := ValidateStudentPaymentDetail(ctx, s.DB, s.StudentPaymentDetailRepo, studentPaymentDetail)
	if err != nil {
		return nil, nil, nil, err
	}
	// Validate BillingAddress
	billingAddress, err := UpsertStudentPaymentInfoRequestToBillingAddressEnt(request)
	if err != nil {
		return nil, nil, nil, status.Error(codes.Internal, fmt.Sprintf("UpsertStudentPaymentInfoRequestToBillingAddressEnt: %v", err))
	}

	newStudentPaymentDetailActionLogType, err := ValidateBillingAddress(ctx, s.DB, s.BillingAddressRepo, s.PrefectureRepo, billingAddress, studentPaymentDetailActionLogType)
	if err != nil {
		return nil, nil, nil, err
	}

	return studentPaymentDetail, billingAddress, newStudentPaymentDetailActionLogType, nil
}

func (s *EditPaymentDetailService) ValidateBankAccountRequest(ctx context.Context, request *invoice_pb.UpsertStudentPaymentInfoRequest) (*entities.BankAccount, *entities.BankAccount, error) {
	// Validate Bank Account
	bankAccountToUpsert, err := UpsertStudentPaymentInfoRequestToBankAccountEnt(request)
	if err != nil {
		return nil, nil, status.Error(codes.Internal, fmt.Sprintf("UpsertStudentPaymentInfoRequestToBankAccountEnt: %v", err))
	}

	existingBankAccount, err := ValidateBankAccount(ctx, s.DB, s.BankAccountRepo, s.BankRepo, s.BankBranchRepo, bankAccountToUpsert)
	if err != nil {
		return nil, nil, err
	}

	return bankAccountToUpsert, existingBankAccount, nil
}

func ValidateStudentPaymentDetail(ctx context.Context, db database.Ext, studentPaymentDetailRepo StudentPaymentDetailRepo, studentPaymentDetailToUpsert *entities.StudentPaymentDetail) (*StudentPaymentActionDetailLogType, error) {
	var studentPaymentActionDetailLogType *StudentPaymentActionDetailLogType
	// Validate fields in student payment detail
	if studentPaymentDetailToUpsert.PayerName.String == "" {
		return nil, status.Error(codes.InvalidArgument, "payer name can not be empty")
	}

	var (
		existingStudentPaymentDetail *entities.StudentPaymentDetail
		err                          error
	)

	if studentPaymentDetailToUpsert.StudentPaymentDetailID.String == "" {
		// Create mode
		// We don't manage multiple student payment detail at this time, remove this if-case logic when we do
		existingStudentPaymentDetail, err = studentPaymentDetailRepo.FindByStudentID(ctx, db, studentPaymentDetailToUpsert.StudentID.String)
		switch err {
		case nil, pgx.ErrNoRows:
			break
		default:
			return nil, status.Error(codes.Internal, fmt.Sprintf("studentPaymentDetailRepo.FindByStudentID: %v", err))
		}

		if existingStudentPaymentDetail != nil {
			return nil, status.Error(codes.FailedPrecondition, "student payment detail already exists, can not create more")
		}
		return nil, nil
	}

	// Update mode
	// Check precondition if updating student payment detail
	existingStudentPaymentDetail, err = studentPaymentDetailRepo.FindByID(ctx, db, studentPaymentDetailToUpsert.StudentPaymentDetailID.String)
	switch err {
	case nil:
		break
	case pgx.ErrNoRows:
		return nil, status.Error(codes.FailedPrecondition, "student payment detail does not exist, can not update")
	default:
		return nil, status.Error(codes.Internal, fmt.Sprintf("studentPaymentDetailRepo.FindByID: %v", err))
	}

	if existingStudentPaymentDetail == nil {
		return nil, status.Error(codes.FailedPrecondition, "student payment detail does not exist, can not update")
	}

	studentPaymentActionDetailLogType = retrieveStudentPaymentDetailFieldsToUpdate(existingStudentPaymentDetail, studentPaymentDetailToUpsert)

	studentPaymentDetailToUpsert.CreatedAt = existingStudentPaymentDetail.CreatedAt
	studentPaymentDetailToUpsert.PaymentMethod = existingStudentPaymentDetail.PaymentMethod

	return studentPaymentActionDetailLogType, nil
}

func ValidateBillingAddress(ctx context.Context, db database.Ext, billingAddressRepo BillingAddressRepo, prefectureRepo PrefectureRepo, billingAddressToUpsert *entities.BillingAddress, studentPaymentActionDetailLogType *StudentPaymentActionDetailLogType) (*StudentPaymentActionDetailLogType, error) {
	// Validate fields in billing address
	switch {
	case billingAddressToUpsert.PostalCode.String == "":
		return nil, status.Error(codes.InvalidArgument, "postal code can not be empty")
	case billingAddressToUpsert.PrefectureCode.String == "":
		return nil, status.Error(codes.InvalidArgument, "prefecture can not be empty")
	case billingAddressToUpsert.City.String == "":
		return nil, status.Error(codes.InvalidArgument, "city can not be empty")
	}

	switch existingPrefecture, err := prefectureRepo.FindByPrefectureCode(ctx, db, billingAddressToUpsert.PrefectureCode.String); err {
	case nil:
		if existingPrefecture == nil {
			return nil, status.Error(codes.FailedPrecondition, "prefecture code does not exist")
		}
	case pgx.ErrNoRows:
		return nil, status.Error(codes.FailedPrecondition, "prefecture code does not exist")
	default:
		return nil, status.Error(codes.Internal, fmt.Sprintf("prefectureRepo.FindByPrefectureCode: %v", err))
	}

	var (
		existingBillingAddress *entities.BillingAddress
		err                    error
	)

	// Check precondition if updating billing address
	if billingAddressToUpsert.BillingAddressID.String == "" {
		// Create mode
		// We don't manage multiple student payment detail at this time, remove this if-case logic when we do
		existingBillingAddress, err = billingAddressRepo.FindByUserID(ctx, db, billingAddressToUpsert.UserID.String)
		switch err {
		case nil, pgx.ErrNoRows:
			break
		default:
			return nil, status.Error(codes.Internal, fmt.Sprintf("billingAddressRepo.FindByUserID: %v", err))
		}

		if existingBillingAddress != nil {
			return nil, status.Error(codes.FailedPrecondition, "billing address already exists, can not create more")
		}

		billingAddressToUpsert.BillingAddressID = database.Text(idutil.ULIDNow())

		return nil, nil
	}
	// Update mode
	// Check precondition if updating billing address
	existingBillingAddress, err = billingAddressRepo.FindByID(ctx, db, billingAddressToUpsert.BillingAddressID.String)
	switch err {
	case nil:
		break
	case pgx.ErrNoRows:
		return nil, status.Error(codes.FailedPrecondition, "billing address does not exist, can not update")
	default:
		return nil, status.Error(codes.Internal, fmt.Sprintf("billingAddressRepo.FindByID: %v", err))
	}

	if existingBillingAddress == nil {
		return nil, status.Error(codes.FailedPrecondition, `billing address does not exist, can not update`)
	}

	// compare fields to update and add on action log
	studentPaymentActionDetailLogType = retrieveBillingAddressFieldsToUpdate(existingBillingAddress, billingAddressToUpsert, studentPaymentActionDetailLogType)

	billingAddressToUpsert.CreatedAt = existingBillingAddress.CreatedAt

	return studentPaymentActionDetailLogType, nil
}

func ValidateBankAccount(ctx context.Context, db database.Ext, bankAccountRepo BankAccountRepo, bankRepo BankRepo, bankBranchRepo BankBranchRepo, bankAccountToUpsert *entities.BankAccount) (*entities.BankAccount, error) {
	if bankAccountToUpsert.IsVerified.Bool {
		switch {
		case bankAccountToUpsert.BankID.String == "":
			return nil, status.Error(codes.InvalidArgument, "bank id can not be empty")
		case bankAccountToUpsert.BankBranchID.String == "":
			return nil, status.Error(codes.InvalidArgument, "bank branch id can not be empty")
		case bankAccountToUpsert.BankAccountHolder.String == "":
			return nil, status.Error(codes.InvalidArgument, "bank account holder can not be empty")
		case bankAccountToUpsert.BankAccountNumber.String == "":
			return nil, status.Error(codes.InvalidArgument, "bank account number can not be empty")
		case len(bankAccountToUpsert.BankAccountNumber.String) != 7:
			return nil, status.Error(codes.InvalidArgument, "bank account number only can accept 7 digit numbers")
		case bankAccountToUpsert.BankAccountType.String == "":
			return nil, status.Error(codes.InvalidArgument, "bank account type can not be empty")
		}

		if err := utils.ValidateBankHolder(bankAccountToUpsert.BankAccountHolder.String); err != nil {
			return nil, err
		}

		bank, err := bankRepo.FindByID(ctx, db, bankAccountToUpsert.BankID.String)
		switch err {
		case nil:
			if bank == nil {
				return nil, status.Error(codes.FailedPrecondition, "bank id does not exist")
			}
		case pgx.ErrNoRows:
			return nil, status.Error(codes.FailedPrecondition, "bank id does not exist")
		default:
			return nil, status.Error(codes.Internal, fmt.Sprintf("bankRepo.FindByID: %v", err))
		}

		bankBranch, err := bankBranchRepo.FindByID(ctx, db, bankAccountToUpsert.BankBranchID.String)
		switch err {
		case nil:
			if bankBranch == nil {
				return nil, status.Error(codes.FailedPrecondition, "bank branch id does not exist")
			}
		case pgx.ErrNoRows:
			return nil, status.Error(codes.FailedPrecondition, "bank branch id does not exist")
		default:
			return nil, status.Error(codes.Internal, fmt.Sprintf("bankBranchRepo.FindByID: %v", err))
		}
	}

	var (
		existingBankAccount *entities.BankAccount
		err                 error
	)

	// Check precondition when creating bank account
	if bankAccountToUpsert.BankAccountID.String == "" {
		// Create mode
		// We don't manage multiple bank account at this time, remove this if-case logic when we do
		existingBankAccount, err = bankAccountRepo.FindByStudentID(ctx, db, bankAccountToUpsert.StudentID.String)
		switch err {
		case nil, pgx.ErrNoRows:
			break
		default:
			return nil, status.Error(codes.Internal, fmt.Sprintf("bankAccountRepo.FindByStudentID: %v", err))
		}

		if existingBankAccount != nil {
			return nil, status.Error(codes.FailedPrecondition, "bank account already exists, can not create more")
		}
		bankAccountToUpsert.BankAccountID = database.Text(idutil.ULIDNow())
	} else {
		// Update mode
		// Check precondition if updating bank account
		existingBankAccount, err = bankAccountRepo.FindByID(ctx, db, bankAccountToUpsert.BankAccountID.String)
		switch err {
		case nil:
			break
		case pgx.ErrNoRows:
			return nil, status.Error(codes.FailedPrecondition, "bank account does not exist, can not update")
		default:
			return nil, status.Error(codes.Internal, fmt.Sprintf("bankAccountRepo.FindByID: %v", err))
		}

		if existingBankAccount == nil {
			return nil, status.Error(codes.FailedPrecondition, `bank account does not exist, can not update`)
		}
		bankAccountToUpsert.CreatedAt = existingBankAccount.CreatedAt
	}

	return existingBankAccount, nil
}

func FindExistingBillingInfoByStudentID(ctx context.Context, db database.QueryExecer, studentPaymentDetailRepo StudentPaymentDetailRepo, billingAddressRepo BillingAddressRepo, studentID string) (*entities.StudentPaymentDetail, *entities.BillingAddress, error) {
	existingStudentPaymentDetail, err := FindExistingStudentPaymentDetailByStudentID(ctx, db, studentPaymentDetailRepo, studentID)
	if err != nil {
		return nil, nil, err
	}

	existingBillingAddress, err := FindExistingBillingAddressByStudentID(ctx, db, billingAddressRepo, studentID)
	if err != nil {
		return nil, nil, err
	}

	return existingStudentPaymentDetail, existingBillingAddress, nil
}

func FindExistingStudentPaymentDetailByStudentID(ctx context.Context, db database.QueryExecer, studentPaymentDetailRepo StudentPaymentDetailRepo, studentID string) (*entities.StudentPaymentDetail, error) {
	existingStudentPaymentDetail, err := studentPaymentDetailRepo.FindByStudentID(ctx, db, studentID)
	switch err {
	case nil:
		return existingStudentPaymentDetail, nil
	case pgx.ErrNoRows:
		return nil, status.Error(codes.FailedPrecondition, "student payment detail does not exist, can not update")
	default:
		return nil, status.Error(codes.Internal, fmt.Sprintf("studentPaymentDetailRepo.FindByID: %v", err))
	}
}

func FindExistingBillingAddressByStudentID(ctx context.Context, db database.QueryExecer, billingAddressRepo BillingAddressRepo, studentID string) (*entities.BillingAddress, error) {
	existingBillingAddress, err := billingAddressRepo.FindByUserID(ctx, db, studentID)
	switch err {
	case nil:
		return existingBillingAddress, nil
	case pgx.ErrNoRows:
		return nil, status.Error(codes.FailedPrecondition, "billing address does not exist, can not update")
	default:
		return nil, status.Error(codes.Internal, fmt.Sprintf("billingAddressRepo.FindByID: %v", err))
	}
}

func retrieveStudentPaymentDetailFieldsToUpdate(existingStudentPaymentDetail *entities.StudentPaymentDetail, studentPaymentDetailToUpdate *entities.StudentPaymentDetail) *StudentPaymentActionDetailLogType {
	// initializing struct as using StudentPaymentActionDetailLogType field directly causing null reference error
	previousBillingAddressHistoryInfo := &BillingAddressHistoryInfo{}
	newBillingAddressHistoryInfo := &BillingAddressHistoryInfo{}

	if existingStudentPaymentDetail.PayerName.String != studentPaymentDetailToUpdate.PayerName.String {
		previousBillingAddressHistoryInfo.PayerName = existingStudentPaymentDetail.PayerName.String
		newBillingAddressHistoryInfo.PayerName = studentPaymentDetailToUpdate.PayerName.String
	}

	if existingStudentPaymentDetail.PayerPhoneNumber.String != studentPaymentDetailToUpdate.PayerPhoneNumber.String {
		previousBillingAddressHistoryInfo.PayerPhoneNumber = existingStudentPaymentDetail.PayerPhoneNumber.String
		newBillingAddressHistoryInfo.PayerPhoneNumber = studentPaymentDetailToUpdate.PayerPhoneNumber.String
	}
	// check if billing address info has no changes
	if *newBillingAddressHistoryInfo == (BillingAddressHistoryInfo{}) {
		return nil
	}

	return &StudentPaymentActionDetailLogType{
		&PreviousDataStudentActionDetailLog{
			BillingAddress: previousBillingAddressHistoryInfo,
		},
		&NewDataStudentActionDetailLog{
			BillingAddress: newBillingAddressHistoryInfo,
		},
	}
}

func retrieveBillingAddressFieldsToUpdate(existingBillingAddress *entities.BillingAddress, billingAddressToUpdate *entities.BillingAddress, studentPaymentActionDetailLogType *StudentPaymentActionDetailLogType) *StudentPaymentActionDetailLogType {
	previousBillingAddressHistoryInfo := &BillingAddressHistoryInfo{}
	newBillingAddressHistoryInfo := &BillingAddressHistoryInfo{}
	// initialize struct for student payment detail action log if student payment detail has no update
	if studentPaymentActionDetailLogType != nil {
		// assigned the existing value for student payment detail changes (payer name, payer phone number)
		previousBillingAddressHistoryInfo = studentPaymentActionDetailLogType.Previous.BillingAddress
		newBillingAddressHistoryInfo = studentPaymentActionDetailLogType.New.BillingAddress
	}

	if existingBillingAddress.PostalCode.String != billingAddressToUpdate.PostalCode.String {
		previousBillingAddressHistoryInfo.PostalCode = existingBillingAddress.PostalCode.String
		newBillingAddressHistoryInfo.PostalCode = billingAddressToUpdate.PostalCode.String
	}

	if existingBillingAddress.City.String != billingAddressToUpdate.City.String {
		previousBillingAddressHistoryInfo.City = existingBillingAddress.City.String
		newBillingAddressHistoryInfo.City = billingAddressToUpdate.City.String
	}

	if existingBillingAddress.Street1.String != billingAddressToUpdate.Street1.String {
		previousBillingAddressHistoryInfo.Street1 = existingBillingAddress.Street1.String
		newBillingAddressHistoryInfo.Street1 = billingAddressToUpdate.Street1.String
	}

	if existingBillingAddress.Street2.String != billingAddressToUpdate.Street2.String {
		previousBillingAddressHistoryInfo.Street2 = existingBillingAddress.Street2.String
		newBillingAddressHistoryInfo.Street2 = billingAddressToUpdate.Street2.String
	}

	if existingBillingAddress.PrefectureCode.String != billingAddressToUpdate.PrefectureCode.String {
		previousBillingAddressHistoryInfo.PrefectureCode = existingBillingAddress.PrefectureCode.String
		newBillingAddressHistoryInfo.PrefectureCode = billingAddressToUpdate.PrefectureCode.String
	}
	// check if billing address info has no changes

	if *newBillingAddressHistoryInfo == (BillingAddressHistoryInfo{}) {
		return nil
	}

	return &StudentPaymentActionDetailLogType{
		&PreviousDataStudentActionDetailLog{
			BillingAddress: previousBillingAddressHistoryInfo,
		},
		&NewDataStudentActionDetailLog{
			BillingAddress: newBillingAddressHistoryInfo,
		},
	}
}

func retrieveBankAccountFieldsToUpdate(existingBankAccount *entities.BankAccount, bankAccountToUpdate *entities.BankAccount, existingPaymentMethod string) *StudentPaymentActionDetailLogType {
	previousBankAccountHistoryInfo := &BankAccountHistoryInfo{}
	newBankAccountHistoryInfo := &BankAccountHistoryInfo{}
	var newPaymentMethod, oldPaymentMethod string

	if existingBankAccount.IsVerified.Bool != bankAccountToUpdate.IsVerified.Bool {
		previousBankAccountHistoryInfo.IsVerified = existingBankAccount.IsVerified.Bool
		newBankAccountHistoryInfo.IsVerified = bankAccountToUpdate.IsVerified.Bool
		oldPaymentMethod = existingPaymentMethod
		newPaymentMethod = invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()

		if bankAccountToUpdate.IsVerified.Bool {
			newPaymentMethod = invoice_pb.PaymentMethod_DIRECT_DEBIT.String()
		}
	}

	if existingBankAccount.BankID.String != bankAccountToUpdate.BankID.String {
		previousBankAccountHistoryInfo.BankID = existingBankAccount.BankID.String
		newBankAccountHistoryInfo.BankID = bankAccountToUpdate.BankID.String
	}

	if existingBankAccount.BankBranchID.String != bankAccountToUpdate.BankBranchID.String {
		previousBankAccountHistoryInfo.BankBranchID = existingBankAccount.BankBranchID.String
		newBankAccountHistoryInfo.BankBranchID = bankAccountToUpdate.BankBranchID.String
	}

	if existingBankAccount.BankAccountNumber.String != bankAccountToUpdate.BankAccountNumber.String {
		previousBankAccountHistoryInfo.BankAccountNumber = existingBankAccount.BankAccountNumber.String
		newBankAccountHistoryInfo.BankAccountNumber = bankAccountToUpdate.BankAccountNumber.String
	}

	if existingBankAccount.BankAccountHolder.String != bankAccountToUpdate.BankAccountHolder.String {
		previousBankAccountHistoryInfo.BankAccountHolder = existingBankAccount.BankAccountHolder.String
		newBankAccountHistoryInfo.BankAccountHolder = bankAccountToUpdate.BankAccountHolder.String
	}

	if existingBankAccount.BankAccountType.String != bankAccountToUpdate.BankAccountType.String {
		previousBankAccountHistoryInfo.BankAccountType = existingBankAccount.BankAccountType.String
		newBankAccountHistoryInfo.BankAccountType = bankAccountToUpdate.BankAccountType.String
	}
	// check if bank history info has no changes
	if *newBankAccountHistoryInfo == (BankAccountHistoryInfo{}) {
		return nil
	}

	return &StudentPaymentActionDetailLogType{
		&PreviousDataStudentActionDetailLog{
			BankAccount:   previousBankAccountHistoryInfo,
			PaymentMethod: oldPaymentMethod,
		},
		&NewDataStudentActionDetailLog{
			BankAccount:   newBankAccountHistoryInfo,
			PaymentMethod: newPaymentMethod,
		},
	}
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
