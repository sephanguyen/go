package entitiescreator

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/common"
	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/try"
	invoicePackageConst "github.com/manabie-com/backend/internal/invoicemgmt/constant"
	invoiceEntities "github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoiceRepo "github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	seqnumberservice "github.com/manabie-com/backend/internal/invoicemgmt/services/sequence_number"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateInvoice used repo.Create to insert invoice.
// stepState dependency:
//   - stepState.StudentID
//
// stepState assigned:
//   - stepState.InvoiceID
func (c *EntitiesCreator) CreateInvoice(ctx context.Context, db database.QueryExecer, status string) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		if err := utils.DoWithMaxRetry(func(attempt int) (bool, error) {
			repo := invoiceRepo.InvoiceRepo{}
			invoice := &invoiceEntities.Invoice{}
			database.AllNullEntity(invoice)
			now := time.Now()

			invoiceTotal := 1000
			var (
				outstandingBalance = invoiceTotal
				amountPaid         = 0
				amountRefunded     = 0
			)

			switch status {
			case invoice_pb.InvoiceStatus_PAID.String():
				amountPaid = invoiceTotal
				outstandingBalance = 0
			case invoice_pb.InvoiceStatus_REFUNDED.String():
				amountRefunded = invoiceTotal
				outstandingBalance = 0
			}

			err := multierr.Combine(
				invoice.Type.Set(fmt.Sprintf("TYPE-%v", idutil.ULIDNow())),
				invoice.Status.Set(status),
				invoice.StudentID.Set(stepState.StudentID),
				// invoice total is set that is copied on payment amount
				invoice.Total.Set(1000),
				invoice.SubTotal.Set(2000),
				invoice.CreatedAt.Set(now),
				invoice.UpdatedAt.Set(now),
				invoice.IsExported.Set(false),
				invoice.OutstandingBalance.Set(outstandingBalance),
				invoice.AmountPaid.Set(amountPaid),
				invoice.AmountRefunded.Set(amountRefunded),
			)
			if err != nil {
				return false, fmt.Errorf("invoice set: %w", err)
			}

			invoiceID, err := repo.Create(ctx, db, invoice)
			if err == nil {
				stepState.InvoiceID = invoiceID.String
				stepState.InvoiceTotal = invoice.Total.Int.Int64()
				return false, nil
			}

			if err != nil && !strings.Contains(err.Error(), "(SQLSTATE 23505)") {
				return false, fmt.Errorf("repo.Create: %w", err)
			}

			time.Sleep(invoiceConst.DuplicateSleepDuration)
			return attempt < MaxRetry, fmt.Errorf("cannot create invoice, err %v", err)
		}, MaxRetry); err != nil {
			return err
		}

		return nil
	}
}

// CreateInvoiceBillItem used InsertInvoiceBillItemStmt to insert invoice_bill_item.
// stepState dependency:
//   - stepState.InvoiceID
//   - stepState.BillItemSequenceNumber
func (c *EntitiesCreator) CreateInvoiceBillItem(ctx context.Context, db database.QueryExecer, pastBillingStatus string) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.WithInvoiceBillItem")
		defer span.End()

		row := db.QueryRow(ctx, InsertInvoiceBillItemStmt,
			stepState.InvoiceID,
			stepState.BillItemSequenceNumber,
			pastBillingStatus,
		)

		var invoiceBillItemID string
		err := row.Scan(&invoiceBillItemID)
		if err != nil {
			return fmt.Errorf("WithInvoiceBillItem err %v", err)
		}

		return nil
	}
}

// CreatePayment used paymentRepo.Create to create payment.
// stepState dependency:
//   - stepState.InvoiceID
func (c *EntitiesCreator) CreatePayment(ctx context.Context, db database.QueryExecer, paymentMethod, paymentStatus, resultCode string, isExported bool) InsertEntityFunction {
	return func(stepState *common.StepState) error {

		payment := &invoiceEntities.Payment{}
		database.AllNullEntity(payment)
		now := time.Now()
		paymentID := idutil.ULIDNow()
		if err := multierr.Combine(
			payment.PaymentID.Set(paymentID),
			payment.InvoiceID.Set(stepState.InvoiceID),
			payment.PaymentMethod.Set(paymentMethod),
			payment.PaymentDueDate.Set(database.TimestamptzFromPb(timestamppb.Now())),
			payment.PaymentExpiryDate.Set(database.TimestamptzFromPb(timestamppb.Now())),
			payment.PaymentStatus.Set(paymentStatus),
			payment.StudentID.Set(stepState.StudentID),
			payment.IsExported.Set(isExported),
			payment.Amount.Set(stepState.InvoiceTotalFloat),
			payment.CreatedAt.Set(now),
			payment.UpdatedAt.Set(now),
		); err != nil {
			return fmt.Errorf("payment set: %w", err)
		}

		if err := utils.DoWithMaxRetry(func(attempt int) (bool, error) {
			err := func() error {
				paymentRepo := &invoiceRepo.PaymentRepo{}
				latestSeqNumber, err := paymentRepo.GetLatestPaymentSequenceNumber(ctx, db)
				if err != nil {
					return err
				}
				latestSeqNumber++

				err = payment.PaymentSequenceNumber.Set(latestSeqNumber)
				if err != nil {
					return err
				}

				cmdTag, err := database.InsertExcept(ctx, payment, []string{"resource_path"}, db.Exec)
				if err != nil {
					return err
				}

				if cmdTag.RowsAffected() != 1 {
					return fmt.Errorf("err insert Payment: %d RowsAffected", cmdTag.RowsAffected())
				}

				return nil
			}()
			if err == nil {
				return false, nil
			}

			if err != nil && !strings.Contains(err.Error(), "(SQLSTATE 23505)") && !strings.Contains(err.Error(), seqnumberservice.PaymentSeqNumberLockAcquiredErr) {
				return false, fmt.Errorf("paymentRepo.Create: %w", err)
			}

			time.Sleep(invoiceConst.DuplicateSleepDuration)
			return attempt < MaxRetry, fmt.Errorf("cannot create payment, err %v", err)
		}, MaxRetry); err != nil {
			return err
		}

		stepState.PaymentID = paymentID

		return nil
	}
}

func (c *EntitiesCreator) CreateInvoiceSchedule(ctx context.Context, db database.QueryExecer, invoiceDate time.Time, scheduledDate time.Time, status string) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.CreateInvoiceSchedule")
		defer span.End()

		invoiceScheduleID := idutil.ULIDNow()

		now := time.Now()
		e := &invoiceEntities.InvoiceSchedule{}
		database.AllNullEntity(e)

		err := multierr.Combine(
			e.InvoiceScheduleID.Set(invoiceScheduleID),
			e.InvoiceDate.Set(invoiceDate),
			e.ScheduledDate.Set(scheduledDate),
			e.Status.Set(status),
			e.UserID.Set(stepState.CurrentUserID),
			e.Remarks.Set(fmt.Sprintf("remarks-%s", invoiceScheduleID)),
			e.IsArchived.Set(false),
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("multierr.Err: %w", err)
		}

		cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
		if err != nil {
			return fmt.Errorf("err insert InvoiceSchedule: %w", err)
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("err insert InvoiceSchedule: %d RowsAffected", cmdTag.RowsAffected())
		}

		stepState.InvoiceScheduleID = invoiceScheduleID

		return nil
	}
}

// CreateBillItemOnInvoicemgmt inserts directly to invoicemgmt database
// Use only you intend to insert directly to invoicemgmt. If not, use CreateBillItem method instead.
func (c *EntitiesCreator) CreateBillItemOnInvoicemgmt(ctx context.Context, db database.QueryExecer, status string) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		randomStr := idutil.ULIDNow()
		billingStartDate := time.Now()
		billingEndDate := billingStartDate.AddDate(0, 0, 28)
		billingDate := billingStartDate.AddDate(0, -1, 15)

		// Set the bill item entity based on its bounded context
		billItem := &invoiceEntities.BillItem{}
		database.AllNullEntity(billItem)

		claims := interceptors.JWTClaimsFromContext(ctx)
		resourcePath := claims.Manabie.ResourcePath

		now := time.Now().UTC()
		err := multierr.Combine(
			billItem.ProductDescription.Set(fmt.Sprintf("PRODUCT-DESCRIPTION-%s", randomStr)),
			billItem.ProductPricing.Set(10),
			billItem.DiscountAmountType.Set(fmt.Sprintf("DISCOUNT-AMOUNT-%s", randomStr)),
			billItem.DiscountAmountValue.Set(int64(10)),
			billItem.TaxID.Set(idutil.ULIDNow()),
			billItem.TaxCategory.Set(fmt.Sprintf("TAX-CATEGORY-%s", randomStr)),
			billItem.TaxPercentage.Set(10),
			billItem.OrderID.Set(idutil.ULIDNow()),
			billItem.BillType.Set(fmt.Sprintf("BILL-TYPE-%s", randomStr)),
			billItem.BillStatus.Set(status),
			billItem.BillDate.Set(billingDate),
			billItem.BillFrom.Set(billingStartDate),
			billItem.BillTo.Set(billingEndDate),
			billItem.BillSchedulePeriodID.Set(idutil.ULIDNow()),
			billItem.DiscountAmount.Set(int64(10)),
			billItem.TaxAmount.Set(int64(10)),
			billItem.FinalPrice.Set(int64(10)),
			billItem.StudentID.Set(stepState.StudentID),
			billItem.BillApprovalStatus.Set(fmt.Sprintf("BILL-APPROVAL-STATUS-%s", randomStr)),
			billItem.LocationID.Set(stepState.LocationID),
			billItem.CreatedAt.Set(now),
			billItem.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("billItem set %v ", err)
		}

		if err := try.Do(func(attempt int) (bool, error) {
			// Fetch the latest bill item sequence number
			stmt := "SELECT bill_item_sequence_number FROM bill_item WHERE resource_path = $1 ORDER BY bill_item DESC LIMIT 1"
			var prevBillItemSequenceNumber int32
			err := db.QueryRow(ctx, stmt, resourcePath).Scan(&prevBillItemSequenceNumber)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return false, fmt.Errorf("error on fetching latest bill item sequence number: %w", err)
			}

			billItemSequenceNumber := prevBillItemSequenceNumber + 1
			_ = billItem.BillItemSequenceNumber.Set(billItemSequenceNumber)

			_, err = database.InsertExcept(ctx, billItem, []string{"resource_path"}, db.Exec)
			if err == nil {
				stepState.BillItemSequenceNumber = billItemSequenceNumber
				stepState.BillItemSequenceNumbers = append(stepState.BillItemSequenceNumbers, billItemSequenceNumber)
				return false, nil
			}

			if err != nil && !strings.Contains(err.Error(), "(SQLSTATE 23505)") {
				return false, fmt.Errorf("billItemRepo.Create: %w", err)
			}

			time.Sleep(invoiceConst.DuplicateSleepDuration)
			return attempt < 20, fmt.Errorf("cannot create bill item, err %v", err)
		}); err != nil {
			return err
		}

		return nil
	}
}

func (c *EntitiesCreator) CreatePartnerConvenienceStore(ctx context.Context, db database.QueryExecer) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.CreateConvenienceStore")
		defer span.End()

		id := idutil.ULIDNow()

		now := time.Now()
		e := &invoiceEntities.PartnerConvenienceStore{}
		database.AllNullEntity(e)

		err := multierr.Combine(
			e.PartnerConvenienceStoreID.Set(id),
			e.ManufacturerCode.Set(976483),
			e.CompanyCode.Set(43267),
			e.ShopCode.Set(fmt.Sprintf("shop-code-%s", id)),
			e.CompanyName.Set(fmt.Sprintf("company-name-%s", id)),
			e.CompanyTelNumber.Set("1234-5678-90"),
			e.PostalCode.Set(fmt.Sprintf("postal-code-%s", id)),
			e.Address1.Set(fmt.Sprintf("address-1-%s", id)),
			e.Address2.Set(fmt.Sprintf("address-2-%s", id)),
			e.Message1.Set(fmt.Sprintf("message-1-%s", id)),
			e.Message2.Set(fmt.Sprintf("message-2-%s", id)),
			e.Message3.Set(fmt.Sprintf("message-3-%s", id)),
			e.Message4.Set(fmt.Sprintf("message-4-%s", id)),
			e.Message5.Set(fmt.Sprintf("message-5-%s", id)),
			e.Message6.Set(fmt.Sprintf("message-6-%s", id)),
			e.Message7.Set(fmt.Sprintf("message-7-%s", id)),
			e.Message8.Set(fmt.Sprintf("message-8-%s", id)),
			e.Message9.Set(fmt.Sprintf("message-9-%s", id)),
			e.Message10.Set(fmt.Sprintf("message-10-%s", id)),
			e.Message11.Set(fmt.Sprintf("message-11-%s", id)),
			e.Message12.Set(fmt.Sprintf("message-12-%s", id)),
			e.Message13.Set(fmt.Sprintf("message-13-%s", id)),
			e.Message14.Set(fmt.Sprintf("message-14-%s", id)),
			e.Message15.Set(fmt.Sprintf("message-15-%s", id)),
			e.Message16.Set(fmt.Sprintf("message-16-%s", id)),
			e.Message17.Set(fmt.Sprintf("message-17-%s", id)),
			e.Message18.Set(fmt.Sprintf("message-18-%s", id)),
			e.Message19.Set(fmt.Sprintf("message-19-%s", id)),
			e.Message20.Set(fmt.Sprintf("message-20-%s", id)),
			e.Message21.Set(fmt.Sprintf("message-21-%s", id)),
			e.Message22.Set(fmt.Sprintf("message-22-%s", id)),
			e.Message23.Set(fmt.Sprintf("message-23-%s", id)),
			e.Message24.Set(fmt.Sprintf("message-24-%s", id)),
			e.IsArchived.Set(false),
			e.Remarks.Set(fmt.Sprintf("remarks-%s", id)),
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("multierr.Err: %w", err)
		}

		cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
		if err != nil {
			return fmt.Errorf("err insert PartnerConvenienceStore: %w", err)
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("err insert PartnerConvenienceStore: %d RowsAffected", cmdTag.RowsAffected())
		}

		stepState.PartnerConvenienceStoreID = id

		return nil
	}
}

func (c *EntitiesCreator) CreatePartnerBank(ctx context.Context, db database.QueryExecer, isDefault bool) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.CreatePartnerBank")
		defer span.End()

		id := idutil.ULIDNow()

		now := time.Now()
		e := &invoiceEntities.PartnerBank{}
		database.AllNullEntity(e)

		err := multierr.Combine(
			e.PartnerBankID.Set(id),
			e.ConsignorCode.Set("74632"),
			e.ConsignorName.Set(fmt.Sprintf("consignor-name-test-%s", id)),
			e.BankNumber.Set("3892"),
			e.BankName.Set(fmt.Sprintf("素晴らしい bank-%s", id)),
			e.BankBranchNumber.Set("352"),
			e.BankBranchName.Set(fmt.Sprintf("ﾊｯﾋﾟｰ bank-branch-%s", id)),
			e.DepositItems.Set(invoicePackageConst.PartnerBankDepositItems[1]),
			e.AccountNumber.Set("1442322"),
			e.IsArchived.Set(false),
			e.IsDefault.Set(isDefault),
			e.Remarks.Set(fmt.Sprintf("remarks-%s", id)),
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
			e.RecordLimit.Set(0),
		)
		if err != nil {
			return fmt.Errorf("multierr.Err: %w", err)
		}

		cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
		if err != nil {
			return fmt.Errorf("err insert PartnerBank: %w", err)
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("err insert PartnerBank: %d RowsAffected", cmdTag.RowsAffected())
		}

		stepState.PartnerBankID = id

		return nil
	}
}

func (c *EntitiesCreator) CreateBulkPaymentRequest(ctx context.Context, db database.QueryExecer, paymentMethod string) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.CreateBulkPaymentRequest")
		defer span.End()

		now := time.Now()
		e := &invoiceEntities.BulkPaymentRequest{}
		database.AllNullEntity(e)

		err := multierr.Combine(
			e.PaymentMethod.Set(paymentMethod),
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("multierr.Err: %w", err)
		}

		repo := &invoiceRepo.BulkPaymentRequestRepo{}
		id, err := repo.Create(ctx, db, e)
		if err != nil {
			return err
		}

		stepState.BulkPaymentRequestID = id

		return nil
	}
}

func (c *EntitiesCreator) CreateBulkPaymentRequestFile(ctx context.Context, db database.QueryExecer, fileExtension string) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.CreateBulkPaymentRequest")
		defer span.End()

		now := time.Now()
		e := &invoiceEntities.BulkPaymentRequestFile{}
		database.AllNullEntity(e)

		err := multierr.Combine(
			e.BulkPaymentRequestID.Set(stepState.BulkPaymentRequestID),
			e.FileName.Set(fmt.Sprintf("test-file-name.%v", fileExtension)),
			e.FileSequenceNumber.Set(1),
			e.TotalFileCount.Set(1),
			e.IsDownloaded.Set(false),
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("multierr.Err: %w", err)
		}

		repo := &invoiceRepo.BulkPaymentRequestFileRepo{}
		id, err := repo.Create(ctx, db, e)
		if err != nil {
			return err
		}

		stepState.BulkPaymentRequestFileID = id

		return nil
	}
}

func (c *EntitiesCreator) CreateBulkPaymentRequestFilePayment(ctx context.Context, db database.QueryExecer, paymentID string) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.CreateBulkPaymentRequestFilePayment")
		defer span.End()

		now := time.Now()
		e := &invoiceEntities.BulkPaymentRequestFilePayment{}
		database.AllNullEntity(e)

		err := multierr.Combine(
			e.BulkPaymentRequestFileID.Set(stepState.BulkPaymentRequestFileID),
			e.PaymentID.Set(paymentID),
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("multierr.Err: %w", err)
		}

		repo := &invoiceRepo.BulkPaymentRequestFilePaymentRepo{}
		_, err = repo.Create(ctx, db, e)
		if err != nil {
			return err
		}

		return nil
	}
}

func (c *EntitiesCreator) CreateBulkPaymentValidations(ctx context.Context, db database.QueryExecer, paymentMethod string) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.CreateBulkPaymentValidations")
		defer span.End()

		id := idutil.ULIDNow()

		now := time.Now()
		e := &invoiceEntities.BulkPaymentValidations{}
		database.AllNullEntity(e)

		switch paymentMethod {
		case "CONVENIENCE STORE":
			paymentMethod = invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()
		case "DIRECT DEBIT":
			paymentMethod = invoice_pb.PaymentMethod_DIRECT_DEBIT.String()
		default:
			return fmt.Errorf("payment method %s is not supported", paymentMethod)
		}

		err := multierr.Combine(
			e.BulkPaymentValidationsID.Set(id),
			e.PaymentMethod.Set(paymentMethod),
			// set default 0 first to update
			e.SuccessfulPayments.Set(0),
			e.FailedPayments.Set(0),
			e.PendingPayments.Set(0),
			e.ValidationDate.Set(now),
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("multierr.Err: %w", err)
		}

		cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
		if err != nil {
			return fmt.Errorf("err insert Bulk Payment Validations: %w", err)
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("err insert Bulk Payment Validations: %d RowsAffected", cmdTag.RowsAffected())
		}

		stepState.BulkPaymentValidationsID = id

		return nil
	}
}

func (c *EntitiesCreator) CreateStudentPaymentDetail(ctx context.Context, db database.QueryExecer, paymentMethod, studentID string) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		studentPaymentDetail := &invoiceEntities.StudentPaymentDetail{}
		database.AllNullEntity(studentPaymentDetail)

		now := time.Now()
		studentPaymentDetailID := idutil.ULIDNow()
		payerName := "payer-name-" + studentPaymentDetailID

		if err := multierr.Combine(
			studentPaymentDetail.StudentPaymentDetailID.Set(studentPaymentDetailID),
			studentPaymentDetail.StudentID.Set(studentID),
			studentPaymentDetail.PaymentMethod.Set(paymentMethod),
			studentPaymentDetail.PayerPhoneNumber.Set("123-4567-890"),
			studentPaymentDetail.PayerName.Set(payerName),
			studentPaymentDetail.CreatedAt.Set(now),
			studentPaymentDetail.UpdatedAt.Set(now),
		); err != nil {
			return fmt.Errorf("student payment detail set: %w", err)
		}

		if err := utils.DoWithMaxRetry(func(attempt int) (bool, error) {
			cmdTag, err := database.InsertExcept(ctx, studentPaymentDetail, []string{"resource_path"}, db.Exec)
			if err == nil {
				if cmdTag.RowsAffected() != 1 {
					return false, fmt.Errorf("err insert Student Payment Detail: %d RowsAffected", cmdTag.RowsAffected())
				}

				return false, nil
			}

			if err != nil && !strings.Contains(err.Error(), "(SQLSTATE 23505)") {
				return false, fmt.Errorf("studentPaymentDetailrepo.Create: %w", err)
			}

			time.Sleep(invoiceConst.DuplicateSleepDuration)
			return attempt < MaxRetry, fmt.Errorf("cannot create student payment detail, err %v", err)
		}, MaxRetry); err != nil {
			return err
		}

		stepState.StudentPaymentDetailID = studentPaymentDetailID
		stepState.CurrentPayerName = payerName

		return nil
	}
}

func (c *EntitiesCreator) CreateBillingAddress(ctx context.Context, db database.QueryExecer) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.CreateBillingAddress")
		defer span.End()

		id := idutil.ULIDNow()

		now := time.Now()
		e := &invoiceEntities.BillingAddress{}
		database.AllNullEntity(e)

		err := multierr.Combine(
			e.BillingAddressID.Set(id),
			e.StudentPaymentDetailID.Set(stepState.StudentPaymentDetailID),
			e.UserID.Set(stepState.StudentID),
			e.PostalCode.Set(fmt.Sprintf("postal-code-%s", id)),
			e.PrefectureCode.Set(stepState.PrefectureCode),
			e.City.Set(fmt.Sprintf("city-%s", id)),
			e.Street1.Set(fmt.Sprintf("street1-%s", id)),
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("multierr.Err: %w", err)
		}

		cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
		if err != nil {
			return fmt.Errorf("err insert BillingAddress: %w", err)
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("err insert BillingAddress: %d RowsAffected", cmdTag.RowsAffected())
		}

		return nil
	}
}

// CreateBank used insert bank statement to insert bank.
// stepState assigned:
//   - stepState.BankID
func (c *EntitiesCreator) CreateBank(ctx context.Context, db database.QueryExecer, isArchived bool) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.CreateBank")
		defer span.End()

		bankID := idutil.ULIDNow()
		// removes the nil in the random generator
		bankCode, err := rand.Int(rand.Reader, big.NewInt(9999))
		if err != nil {
			return err
		}

		bankName := fmt.Sprintf("大きい-%v", bankID)
		bankNamePhonetic := fmt.Sprintf("大きい-phonetic-%v", bankID)

		stmt := InsertBankStmt
		args := []interface{}{bankID, fmt.Sprint(bankCode.Int64()), bankName, bankNamePhonetic, isArchived}

		if _, err := db.Exec(ctx, stmt, args...); err != nil {
			return fmt.Errorf("error insert new bank record: %v", err)
		}

		stepState.BankID = bankID
		stepState.BankCode = fmt.Sprint(bankCode.Int64())

		return nil
	}
}

func (c *EntitiesCreator) CreateBankMapping(ctx context.Context, db database.QueryExecer) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.CreateBankMapping")
		defer span.End()

		id := idutil.ULIDNow()

		now := time.Now()
		e := &invoiceEntities.BankMapping{}
		database.AllNullEntity(e)

		err := multierr.Combine(
			e.BankMappingID.Set(id),
			e.PartnerBankID.Set(stepState.PartnerBankID),
			e.BankID.Set(stepState.BankID),
			e.IsArchived.Set(false),
			e.Remarks.Set(fmt.Sprintf("remarks-%s", id)),
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("multierr.Err: %w", err)
		}

		cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
		if err != nil {
			return fmt.Errorf("err insert BankMapping: %w", err)
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("err insert BankMapping: %d RowsAffected", cmdTag.RowsAffected())
		}

		return nil
	}
}

func (c *EntitiesCreator) CreateBankBranch(ctx context.Context, db database.QueryExecer, isArchived bool) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.CreateBankBranch")
		defer span.End()

		id := idutil.ULIDNow()

		now := time.Now()
		e := &invoiceEntities.BankBranch{}
		database.AllNullEntity(e)
		// removes the nil in the random generator
		bankBranchCode, err := rand.Int(rand.Reader, big.NewInt(999))
		if err != nil {
			return err
		}

		err = multierr.Combine(
			e.BankBranchID.Set(id),
			e.BankID.Set(stepState.BankID),
			e.BankBranchCode.Set(fmt.Sprint(bankBranchCode.Int64())),
			e.BankBranchName.Set(fmt.Sprintf("bank-branch-name-%s", id)),
			e.BankBranchPhoneticName.Set(fmt.Sprintf("bank-branch-phonetic-name-%s", id)),
			e.IsArchived.Set(isArchived),
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("multierr.Err: %w", err)
		}

		cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
		if err != nil {
			return fmt.Errorf("err insert BankBranch: %w", err)
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("err insert BankBranch: %d RowsAffected", cmdTag.RowsAffected())
		}

		stepState.BankBranchID = id
		stepState.BankBranchCode = fmt.Sprint(bankBranchCode.Int64())

		return nil
	}
}

func (c *EntitiesCreator) CreateBankAccount(ctx context.Context, db database.QueryExecer, isVerified bool) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.CreateBankAccount")
		defer span.End()

		id := idutil.ULIDNow()

		now := time.Now()
		e := &invoiceEntities.BankAccount{}
		database.AllNullEntity(e)

		err := multierr.Combine(
			e.BankAccountID.Set(id),
			e.StudentPaymentDetailID.Set(stepState.StudentPaymentDetailID),
			e.StudentID.Set(stepState.StudentID),
			e.IsVerified.Set(isVerified),
			e.BankBranchID.Set(stepState.BankBranchID),
			e.BankAccountNumber.Set("1234567"),
			e.BankAccountHolder.Set(fmt.Sprintf("ﾊﾝｻﾑ-%s", id)),
			e.BankAccountType.Set(invoicePackageConst.PartnerBankDepositItems[1]),
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("multierr.Err: %w", err)
		}

		cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
		if err != nil {
			return fmt.Errorf("err insert BankAccount: %w", err)
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("err insert BankAccount: %d RowsAffected", cmdTag.RowsAffected())
		}

		return nil
	}
}

func (c *EntitiesCreator) UpsertStudentPaymentDetail(ctx context.Context, db database.QueryExecer, paymentMethod string) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		studentPaymentDetailRepo := &invoiceRepo.StudentPaymentDetailRepo{}
		e := &invoiceEntities.StudentPaymentDetail{}
		database.AllNullEntity(e)
		now := time.Now()
		err := multierr.Combine(
			e.StudentPaymentDetailID.Set(idutil.ULIDNow()),
			e.StudentID.Set(stepState.StudentID),
			e.PayerName.Set(fmt.Sprintf("payer-name-%s", stepState.StudentID)),
			e.PayerPhoneNumber.Set(fmt.Sprintf("payer-phone-num-%s", stepState.StudentID)),
			e.PaymentMethod.Set(paymentMethod),
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("multierr.Err: %w", err)
		}

		err = studentPaymentDetailRepo.Upsert(ctx, db)
		if err != nil {
			return fmt.Errorf("studentPaymentDetailRepo.Upsert err: %v", err)
		}

		stepState.StudentPaymentDetailID = e.StudentPaymentDetailID.String

		return nil
	}
}

func (c *EntitiesCreator) CreateInvoiceAdjustment(ctx context.Context, db database.QueryExecer, invoiceID, studentID string, amount float64) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.CreateInvoiceAdjustment")
		defer span.End()

		id := idutil.ULIDNow()

		now := time.Now()
		e := &invoiceEntities.InvoiceAdjustment{}
		database.AllNullEntity(e)

		err := multierr.Combine(
			e.InvoiceAdjustmentID.Set(id),
			e.InvoiceID.Set(invoiceID),
			e.StudentID.Set(studentID),
			e.Description.Set(fmt.Sprintf("description-%s", id)),
			e.Amount.Set(amount),
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("multierr.Err: %w", err)
		}

		if err := utils.DoWithMaxRetry(func(attempt int) (bool, error) {
			cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path", "invoice_adjustment_sequence_number"}, db.Exec)
			if err == nil {
				if cmdTag.RowsAffected() != 1 {
					return false, fmt.Errorf("err insert InvoiceAdjustment: %d RowsAffected", cmdTag.RowsAffected())
				}

				return false, nil
			}

			if err != nil && !strings.Contains(err.Error(), "(SQLSTATE 23505)") {
				return false, fmt.Errorf("err insert InvoiceAdjustment: %w", err)
			}

			time.Sleep(invoiceConst.DuplicateSleepDuration)
			return attempt < MaxRetry, fmt.Errorf("cannot create payment, err %v", err)
		}, MaxRetry); err != nil {
			return err
		}

		stepState.InvoiceAdjustmentIDs = append(stepState.InvoiceAdjustmentIDs, id)

		return nil
	}
}

func (c *EntitiesCreator) CreateBulkPayment(ctx context.Context, db database.QueryExecer, bulkPaymentStatus, paymentMethod string) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.WithBulkPayment")
		defer span.End()

		id := idutil.ULIDNow()

		now := time.Now()
		e := &invoiceEntities.BulkPayment{}
		database.AllNullEntity(e)

		err := multierr.Combine(
			e.BulkPaymentID.Set(id),
			e.BulkPaymentStatus.Set(bulkPaymentStatus),
			e.PaymentMethod.Set(paymentMethod),
			e.InvoiceStatus.Set(invoice_pb.InvoiceStatus_ISSUED.String()),
			e.InvoiceType.Set([]string{invoice_pb.InvoiceType_MANUAL.String()}),
			e.PaymentStatus.Set([]string{invoice_pb.PaymentStatus_PAYMENT_FAILED.String()}),
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("multierr.Err: %w", err)
		}

		cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
		if err != nil {
			return fmt.Errorf("EntitiesCreator.WithBulkPayment err %v", err)
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("err insert BulkPayment: %d RowsAffected", cmdTag.RowsAffected())
		}

		stepState.BulkPaymentID = id

		return nil
	}
}

func (c *EntitiesCreator) CreateMigratedInvoice(ctx context.Context, db database.QueryExecer, status string) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		if err := utils.DoWithMaxRetry(func(attempt int) (bool, error) {
			repo := invoiceRepo.InvoiceRepo{}
			invoice := &invoiceEntities.Invoice{}
			database.AllNullEntity(invoice)
			now := time.Now()
			referenceID := idutil.ULIDNow()
			referenceID2 := fmt.Sprintf("reference-2-%v", referenceID)

			invoiceTotal := 1000
			var (
				outstandingBalance = invoiceTotal
				amountPaid         = 0
				amountRefunded     = 0
			)

			switch status {
			case invoice_pb.InvoiceStatus_PAID.String():
				amountPaid = invoiceTotal
				outstandingBalance = 0
			case invoice_pb.InvoiceStatus_REFUNDED.String():
				amountRefunded = invoiceTotal
				outstandingBalance = 0
			}

			err := multierr.Combine(
				invoice.Type.Set(fmt.Sprintf("TYPE-%v", idutil.ULIDNow())),
				invoice.Status.Set(status),
				invoice.StudentID.Set(stepState.StudentID),
				// invoice total is set that is copied on payment amount
				invoice.Total.Set(1000),
				invoice.SubTotal.Set(2000),
				invoice.CreatedAt.Set(now),
				invoice.UpdatedAt.Set(now),
				invoice.IsExported.Set(false),
				invoice.OutstandingBalance.Set(outstandingBalance),
				invoice.AmountPaid.Set(amountPaid),
				invoice.AmountRefunded.Set(amountRefunded),
				invoice.MigratedAt.Set(now),
				invoice.InvoiceReferenceID.Set(referenceID),
				invoice.InvoiceReferenceID2.Set(referenceID2),
			)
			if err != nil {
				return false, fmt.Errorf("invoice set: %w", err)
			}

			invoiceID, err := repo.Create(ctx, db, invoice)
			if err == nil {
				stepState.InvoiceID = invoiceID.String
				exactInvoiceTotal, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.Total, "2")
				if err != nil {
					return false, fmt.Errorf("err invoice total set: %w", err)
				}
				stepState.InvoiceTotalFloat = exactInvoiceTotal
				stepState.InvoiceReferenceID = referenceID
				stepState.InvoiceReferenceID2 = referenceID2

				return false, nil
			}

			if err != nil && !strings.Contains(err.Error(), "(SQLSTATE 23505)") {
				return false, fmt.Errorf("repo.Create: %w", err)
			}

			time.Sleep(invoiceConst.DuplicateSleepDuration)
			return attempt < MaxRetry, fmt.Errorf("cannot create invoice, err %v", err)
		}, MaxRetry); err != nil {
			return err
		}

		return nil
	}
}

func (c *EntitiesCreator) CreateInvoiceV2(ctx context.Context, db database.QueryExecer, status string, invoiceTotal float64) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		if err := utils.DoWithMaxRetry(func(attempt int) (bool, error) {
			repo := invoiceRepo.InvoiceRepo{}
			invoice := &invoiceEntities.Invoice{}
			database.AllNullEntity(invoice)
			now := time.Now()

			var (
				outstandingBalance = invoiceTotal
				amountPaid         = 0.00
				amountRefunded     = 0.00
			)

			switch status {
			case invoice_pb.InvoiceStatus_PAID.String():
				amountPaid = invoiceTotal
				outstandingBalance = 0
			case invoice_pb.InvoiceStatus_REFUNDED.String():
				amountRefunded = invoiceTotal
				outstandingBalance = 0
			}

			err := multierr.Combine(
				invoice.Type.Set(fmt.Sprintf("TYPE-%v", idutil.ULIDNow())),
				invoice.Status.Set(status),
				invoice.StudentID.Set(stepState.StudentID),
				// invoice total is set that is copied on payment amount
				invoice.Total.Set(invoiceTotal),
				invoice.SubTotal.Set(invoiceTotal),
				invoice.CreatedAt.Set(now),
				invoice.UpdatedAt.Set(now),
				invoice.IsExported.Set(false),
				invoice.OutstandingBalance.Set(outstandingBalance),
				invoice.AmountPaid.Set(amountPaid),
				invoice.AmountRefunded.Set(amountRefunded),
			)
			if err != nil {
				return false, fmt.Errorf("invoice set: %w", err)
			}

			invoiceID, err := repo.Create(ctx, db, invoice)
			if err == nil {
				stepState.InvoiceID = invoiceID.String
				stepState.InvoiceTotal = invoice.Total.Int.Int64()
				return false, nil
			}

			if err != nil && !strings.Contains(err.Error(), "(SQLSTATE 23505)") {
				return false, fmt.Errorf("repo.Create: %w", err)
			}

			time.Sleep(invoiceConst.DuplicateSleepDuration)
			return attempt < MaxRetry, fmt.Errorf("cannot create invoice, err %v", err)
		}, MaxRetry); err != nil {
			return err
		}

		return nil
	}
}
