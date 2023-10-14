package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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

const DateFormat = "2006-01-02"

func (s *DataMigrationModifierService) InsertInvoiceDataMigration(ctx context.Context, db database.QueryExecer, lines [][]string) ([]*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError, error) {
	validInvoices, errLines := s.retrieveValidatedInvoiceData(ctx, db, lines)
	var err error
	if len(validInvoices) > 0 {
		err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
			for _, invoice := range validInvoices {
				if _, err := s.InvoiceRepo.Create(ctx, tx, invoice); err != nil {
					// if there's error on creating a valid invoice, roll back
					return status.Error(codes.Internal, fmt.Sprintf("Data Migration error: %v when creating a valid invoice with reference: %v", err.Error(), invoice.InvoiceReferenceID.String))
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

func (s *DataMigrationModifierService) retrieveValidatedInvoiceData(ctx context.Context, db database.QueryExecer, lines [][]string) ([]*entities.Invoice, []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError) {
	errLines := []*invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{}
	validInvoices := []*entities.Invoice{}

	mandatoryColumns := []int{
		InvoiceStudentIDReference,
		InvoiceType,
		InvoiceStatus,
		InvoiceTotal,
		InvoiceSubTotal,
		InvoiceCreatedAt,
		InvoiceIsExported,
		InvoiceReference1,
		InvoiceReference2,
	}

	for i, line := range lines {
		// Validate the mandatory fields
		err := checkMandatoryColumnAndGetIndex(line, mandatoryColumns, invoice_pb.DataMigrationEntityName_INVOICE_ENTITY.String())
		if err != nil {
			errLines = append(errLines, &invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				RowNumber: int32(i) + 2,
				Error:     status.Error(codes.InvalidArgument, err.Error()).Error(),
			})
			continue
		}
		// Validate invoice data fields
		invoice, err := setAndValidateInvoiceDataTypeFields(line)
		if err != nil {
			errLines = append(errLines, &invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				RowNumber: int32(i) + 2,
				Error:     status.Error(codes.InvalidArgument, err.Error()).Error(),
			})
			continue
		}
		// Set the outstanding_balance, amount_paid and amount_refunded based on invoice status
		err = setInvoiceOutstandingAmountValues(invoice, line)
		if err != nil {
			errLines = append(errLines, &invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				RowNumber: int32(i) + 2,
				Error:     status.Error(codes.InvalidArgument, err.Error()).Error(),
			})
			continue
		}
		// Validate invoice fields
		err = validateInvoiceFields(invoice)
		if err != nil {
			errLines = append(errLines, &invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				RowNumber: int32(i) + 2,
				Error:     status.Error(codes.InvalidArgument, err.Error()).Error(),
			})
			continue
		}
		studentIDTrim := strings.TrimSpace(line[InvoiceStudentIDReference])

		// Validate bill items of invoice
		err = s.validateInvoiceBillItems(ctx, db, studentIDTrim, invoice.InvoiceReferenceID2.String, invoice.Total)
		if err != nil {
			errLines = append(errLines, &invoice_pb.ImportDataMigrationResponse_ImportMigrationDataError{
				RowNumber: int32(i) + 2,
				Error:     status.Error(codes.Internal, err.Error()).Error(),
			})
			continue
		}

		validInvoices = append(validInvoices, invoice)
	}

	return validInvoices, errLines
}

func setAndValidateInvoiceDataTypeFields(line []string) (*entities.Invoice, error) {
	invoice := &entities.Invoice{}
	database.AllNullEntity(invoice)

	now := time.Now()
	if err := multierr.Combine(
		StringToFormatString("type", line[InvoiceType], false, invoice.Type.Set),
		StringToFormatString("status", line[InvoiceStatus], false, invoice.Status.Set),
		StringToFloat64("sub_total", line[InvoiceSubTotal], false, invoice.SubTotal.Set),
		StringToFloat64("total", line[InvoiceTotal], false, invoice.Total.Set),
		StringToDate("created_at", line[InvoiceCreatedAt], utils.CountryJp, false, invoice.CreatedAt.Set),
		StringToBool("is_exported", line[InvoiceIsExported], false, invoice.IsExported.Set),
		StringToFormatString("student_id", line[InvoiceStudentIDReference], false, invoice.StudentID.Set),
		StringToFormatString("invoice_reference_id", line[InvoiceReference1], false, invoice.InvoiceReferenceID.Set),
		StringToFormatString("invoice_reference_id2", line[InvoiceReference2], false, invoice.InvoiceReferenceID2.Set),
		invoice.MigratedAt.Set(now),
	); err != nil {
		return nil, err
	}

	return invoice, nil
}

func setInvoiceOutstandingAmountValues(invoice *entities.Invoice, line []string) error {
	var (
		outstandingBalance float64
		amountPaid         float64
		amountRefunded     float64
	)

	total, err := strconv.ParseFloat(line[InvoiceTotal], 64)
	if err != nil {
		return fmt.Errorf("error parsing string to float64 %v: %w", "total", err)
	}

	switch line[InvoiceStatus] {
	case "PAID":
		amountPaid = total
	case "REFUNDED":
		amountRefunded = total
	default:
		outstandingBalance = total
	}

	return multierr.Combine(
		invoice.OutstandingBalance.Set(outstandingBalance),
		invoice.AmountPaid.Set(amountPaid),
		invoice.AmountRefunded.Set(amountRefunded),
	)
}

func (s *DataMigrationModifierService) validateInvoiceBillItems(ctx context.Context, db database.QueryExecer, studentID, invoiceReference string, invoiceTotal pgtype.Numeric) error {
	totalFinalPrice, err := s.BillItemRepo.GetBillItemTotalByStudentAndReference(ctx, db, studentID, invoiceReference)
	if err != nil || totalFinalPrice.Status == pgtype.Null {
		return fmt.Errorf("cannot retrieve bill items with student id: %v and invoice reference: %v", studentID, invoiceReference)
	}

	exactTotalFinalPrice, err := utils.GetFloat64ExactValueAndDecimalPlaces(totalFinalPrice, "2")
	if err != nil {
		return fmt.Errorf("cannot assign bill item total final price: %v to float data type", totalFinalPrice)
	}

	exactInvoiceTotal, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoiceTotal, "2")
	if err != nil {
		return fmt.Errorf("cannot assign invoice total: %v to float data type", exactInvoiceTotal)
	}

	if exactTotalFinalPrice != exactInvoiceTotal {
		return fmt.Errorf("invoice and bill item has different total amount and total final price %v - %v", exactInvoiceTotal, exactTotalFinalPrice)
	}

	return nil
}

func validateInvoiceFields(invoice *entities.Invoice) error {
	// Validate invoice status
	if _, ok := InvoiceStatusStructMap[invoice.Status.String]; !ok {
		return fmt.Errorf("invoice status %v is invalid", invoice.Status.String)
	}

	// Validate invoice type
	if _, ok := InvoiceTypeStructMap[invoice.Type.String]; !ok {
		return fmt.Errorf("invoice type %v is invalid", invoice.Type.String)
	}

	exactTotal, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.Total, "2")
	if err != nil {
		return err
	}

	exactSubTotal, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.SubTotal, "2")
	if err != nil {
		return err
	}

	// Validate sub total and total based on invoice status
	switch invoice.Status.String {
	case invoice_pb.InvoiceStatus_PAID.String():
		if exactTotal < 0 || exactSubTotal < 0 {
			return errors.New("total or sub_total should not be negative")
		}
	case invoice_pb.InvoiceStatus_REFUNDED.String():
		if exactTotal > 0 || exactSubTotal > 0 {
			return errors.New("total or sub_total should not be positive")
		}
	}

	return nil
}
