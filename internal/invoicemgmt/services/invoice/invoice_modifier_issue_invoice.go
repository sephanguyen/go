package invoicesvc

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Set maxRetry to 150 since there are some instance that it exceeded 10 retries
const issueInvoiceMaxRetry = 150

func (s *InvoiceModifierService) IssueInvoice(ctx context.Context, req *invoice_pb.IssueInvoiceRequest) (*invoice_pb.IssueInvoiceResponse, error) {
	failedResp := &invoice_pb.IssueInvoiceResponse{
		Successful: false,
	}

	successfulResp := &invoice_pb.IssueInvoiceResponse{
		Successful: true,
	}

	// input data validation
	if err := utils.ValidateDueDateAndExpiryDate(req.DueDate, req.ExpiryDate); err != nil {
		return failedResp, status.Error(codes.InvalidArgument, err.Error())
	}

	invoice, err := s.InvoiceRepo.RetrieveInvoiceByInvoiceID(ctx, s.DB, req.InvoiceIdString)
	if err != nil {
		return failedResp, status.Error(codes.Internal, err.Error())
	}

	// create entity for storing to the DB
	paymentEntity, err := generatePaymentData(req, invoice)
	if err != nil {
		return failedResp, status.Error(codes.InvalidArgument, err.Error())
	}

	// validate payment method
	if err := validateIssueInvoicePaymentMethod(req.PaymentMethod, invoice.Type.String); err != nil {
		return failedResp, status.Error(codes.InvalidArgument, err.Error())
	}
	// Retry when there is a duplicate error. Most likely from payment sequence number
	err = utils.DoWithMaxRetry(func(attempt int) (bool, error) {
		err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
			// store entity to the DB
			if err := s.PaymentRepo.Create(ctx, tx, paymentEntity); err != nil {
				return err
			}

			// update invoice status to issued
			if err := s.updateInvoiceStatusAndExportedTag(ctx, tx, req.InvoiceIdString, invoice_pb.InvoiceStatus_ISSUED.String(), false); err != nil {
				return err
			}

			payment, err := s.PaymentRepo.GetLatestPaymentDueDateByInvoiceID(ctx, tx, req.InvoiceIdString)
			if err != nil {
				return err
			}

			invoice.Status = database.Text(invoice_pb.InvoiceStatus_ISSUED.String())

			// generate action logs details
			actionDetails := s.generateActionLogDetails(invoice, payment, req.Remarks)
			if err := s.createActionLog(ctx, tx, actionDetails); err != nil {
				return err
			}

			return nil
		})

		if err == nil {
			return false, nil
		}

		if err != nil && !strings.Contains(err.Error(), "(SQLSTATE 23505)") {
			return false, err
		}

		log.Printf("Retrying issuing single invoice. Attempt: %d \n", attempt)
		return attempt < issueInvoiceMaxRetry, fmt.Errorf("cannot issue invoice: %v", err)
	}, issueInvoiceMaxRetry)

	if err != nil {
		return failedResp, status.Error(codes.Internal, err.Error())
	}

	return successfulResp, nil
}

func validateIssueInvoicePaymentMethod(paymentMethod invoice_pb.PaymentMethod, invoiceType string) error {
	if !constant.SingleInvoicePaymentMethods[paymentMethod.String()] {
		return fmt.Errorf("invalid PaymentMethod value: %s", paymentMethod)
	}

	return nil
}

func generatePaymentData(data *invoice_pb.IssueInvoiceRequest, invoice *entities.Invoice) (*entities.Payment, error) {
	e := new(entities.Payment)
	database.AllNullEntity(e)

	invoiceTotal, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.Total, "2")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	now := time.Now()
	if err := multierr.Combine(
		e.InvoiceID.Set(data.InvoiceIdString),
		e.PaymentMethod.Set(data.PaymentMethod.String()),
		e.PaymentDueDate.Set(database.TimestamptzFromPb(data.DueDate)),
		e.PaymentExpiryDate.Set(database.TimestamptzFromPb(data.ExpiryDate)),
		e.PaymentStatus.Set(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
		e.StudentID.Set(invoice.StudentID.String),
		e.Amount.Set(invoiceTotal),
		e.IsExported.Set(false),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	); err != nil {
		return nil, fmt.Errorf("multierr.Combine: %w", err)
	}

	return e, nil
}
