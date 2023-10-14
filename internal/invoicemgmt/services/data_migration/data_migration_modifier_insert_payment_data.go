package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *DataMigrationModifierService) InsertPaymentDataMigration(ctx context.Context, db database.QueryExecer, lines [][]string) ([]*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError, error) {
	validPayments, errLines := s.retrieveValidatedPaymentData(ctx, db, lines)
	var err error
	if len(validPayments) > 0 {
		err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
			for _, payment := range validPayments {
				if err := s.PaymentRepo.Create(ctx, tx, payment); err != nil {
					// if there's error on creating a valid payment, roll back
					return status.Error(codes.Internal, fmt.Sprintf("Data Migration error: %v when creating a valid payment invoice with reference: %v", err.Error(), payment.PaymentReferenceID.String))
				}
			}

			return nil
		})
	}

	if err != nil {
		return errLines, err
	}

	return errLines, nil
}

func (s *DataMigrationModifierService) retrieveValidatedPaymentData(ctx context.Context, db database.QueryExecer, lines [][]string) ([]*entities.Payment, []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError) {
	errLines := []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{}
	validPayments := []*entities.Payment{}

	// validate mandatory payment data column values
	mandatoryColumns := []int{PaymentMethod, PaymentStatus, PaymentDueDate, PaymentExpiryDate, PaymentStudentID, PaymentIsExported, PaymentCreatedAt, PaymentInvoiceReference}
	for i, line := range lines {
		err := checkMandatoryColumnAndGetIndex(line, mandatoryColumns, invoice_pb.DataMigrationEntityName_PAYMENT_ENTITY.String())
		if err != nil {
			errLines = append(errLines, &invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				RowNumber: int32(i) + 2,
				Error:     status.Error(codes.InvalidArgument, err.Error()).Error(),
			})
			continue
		}
		// validate payment data type fields
		payment, err := setAndValidatePaymentDataTypeFields(line)
		if err != nil {
			errLines = append(errLines, &invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				RowNumber: int32(i) + 2,
				Error:     status.Error(codes.InvalidArgument, err.Error()).Error(),
			})
			continue
		}

		// validate payment data conditions for status and dates
		err = validatePaymentStatusAndDates(payment)
		if err != nil {
			errLines = append(errLines, &invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				RowNumber: int32(i) + 2,
				Error:     status.Error(codes.InvalidArgument, err.Error()).Error(),
			})
			continue
		}

		invoice, err := s.getPaymentRelatedInvoice(ctx, db, payment.PaymentReferenceID.String)
		if err != nil {
			errLines = append(errLines, &invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				RowNumber: int32(i) + 2,
				Error:     status.Error(codes.Internal, err.Error()).Error(),
			})
			continue
		}
		err = validatePaymentStatusWithInvoiceStatus(payment.PaymentStatus.String, invoice.Status.String)
		if err != nil {
			errLines = append(errLines, &invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				RowNumber: int32(i) + 2,
				Error:     status.Error(codes.Internal, err.Error()).Error(),
			})
			continue
		}
		paymentStudentTrim := strings.TrimSpace(line[PaymentStudentID])
		if paymentStudentTrim != invoice.StudentID.String {
			err = status.Error(codes.Internal, fmt.Sprintf("error student id: %v mismatch on invoice student id: %v", paymentStudentTrim, invoice.StudentID.String))
			errLines = append(errLines, &invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				RowNumber: int32(i) + 2,
				Error:     err.Error(),
			})
			continue
		}

		payment.StudentID = database.Text(invoice.StudentID.String)
		payment.Amount = invoice.Total
		payment.InvoiceID = database.Text(invoice.InvoiceID.String)
		validPayments = append(validPayments, payment)
	}

	return validPayments, errLines
}

func (s *DataMigrationModifierService) getPaymentRelatedInvoice(ctx context.Context, db database.QueryExecer, invoiceReference string) (*entities.Invoice, error) {
	invoice, err := s.InvoiceRepo.RetrieveInvoiceByInvoiceReferenceID(ctx, db, invoiceReference)
	if err != nil {
		return nil, fmt.Errorf("error retrieving invoice with reference: %v", invoiceReference)
	}

	return invoice, nil
}

func setAndValidatePaymentDataTypeFields(line []string) (*entities.Payment, error) {
	payment := &entities.Payment{}
	database.AllNullEntity(payment)

	if err := multierr.Combine(
		StringToFormatString("payment_method", line[PaymentMethod], false, payment.PaymentMethod.Set),
		StringToFormatString("payment_status", line[PaymentStatus], false, payment.PaymentStatus.Set),
		StringToDate("payment_due_date", line[PaymentDueDate], utils.CountryJp, false, payment.PaymentDueDate.Set),
		StringToDate("payment_expiry_date", line[PaymentExpiryDate], utils.CountryJp, false, payment.PaymentExpiryDate.Set),
		StringToDate("payment_date", line[PaymentDate], utils.CountryJp, true, payment.PaymentDate.Set),
		StringToBool("is_exported", line[PaymentIsExported], false, payment.IsExported.Set),
		StringToDate("created_at", line[PaymentCreatedAt], utils.CountryJp, false, payment.CreatedAt.Set),
		StringToFormatString("reference", line[PaymentInvoiceReference], false, payment.PaymentReferenceID.Set),
		payment.MigratedAt.Set(time.Now().UTC()),
	); err != nil {
		return nil, err
	}

	return payment, nil
}

func validatePaymentStatusAndDates(payment *entities.Payment) error {
	if payment.PaymentStatus.String == invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String() && payment.PaymentDate.Status == pgtype.Null {
		return fmt.Errorf("payment invoice reference: %v with successful status should have a payment date", payment.PaymentReferenceID.String)
	}

	if payment.PaymentDueDate.Time.After(payment.PaymentExpiryDate.Time) {
		return fmt.Errorf("invalid payment due date: %v must be before expiry date: %v on invoice reference: %v", payment.PaymentDueDate.Time.String(), payment.PaymentExpiryDate.Time.String(), payment.PaymentReferenceID.String)
	}

	return nil
}

func validatePaymentStatusWithInvoiceStatus(paymentStatus, invoiceStatus string) error {
	switch invoiceStatus {
	case invoice_pb.InvoiceStatus_ISSUED.String():
		if paymentStatus != invoice_pb.PaymentStatus_PAYMENT_PENDING.String() {
			return fmt.Errorf("%v invoice should have payment status %v but got: %v", invoice_pb.InvoiceStatus_ISSUED.String(), invoice_pb.PaymentStatus_PAYMENT_PENDING, paymentStatus)
		}
	case invoice_pb.InvoiceStatus_PAID.String():
		if paymentStatus != invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String() {
			return fmt.Errorf("%v invoice should have payment status %v but got: %v", invoice_pb.InvoiceStatus_PAID.String(), invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL, paymentStatus)
		}
	case invoice_pb.InvoiceStatus_REFUNDED.String():
		if paymentStatus != invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String() {
			return fmt.Errorf("%v invoice should have payment status %v but got: %v", invoice_pb.InvoiceStatus_REFUNDED.String(), invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL, paymentStatus)
		}
	case invoice_pb.InvoiceStatus_FAILED.String():
		if paymentStatus != invoice_pb.PaymentStatus_PAYMENT_FAILED.String() {
			return fmt.Errorf("%v invoice should have payment status %v but got: %v", invoice_pb.InvoiceStatus_FAILED.String(), invoice_pb.PaymentStatus_PAYMENT_FAILED, paymentStatus)
		}
	default:
		return fmt.Errorf("invalid invoice status: %v for payment status: %v", invoiceStatus, paymentStatus)
	}

	return nil
}
