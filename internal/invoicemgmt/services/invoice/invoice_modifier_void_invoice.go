package invoicesvc

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *InvoiceModifierService) VoidInvoice(ctx context.Context, req *invoice_pb.VoidInvoiceRequest) (*invoice_pb.VoidInvoiceResponse, error) {
	failedResp := &invoice_pb.VoidInvoiceResponse{
		Successful: false,
	}

	successfulResp := &invoice_pb.VoidInvoiceResponse{
		Successful: true,
	}

	// Validate request data and invoice existence
	if err := validateVoidInvoiceRequest(ctx, s.DB, s, req.InvoiceId); err != nil {
		return failedResp, err
	}

	// Roll back DB tx if there's error
	err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		// Step 1: DB update transactions should be successful first

		// Step 1a: Update payment record (if there's any) to failed status
		payment, err := s.PaymentRepo.GetLatestPaymentDueDateByInvoiceID(ctx, tx, req.InvoiceId)

		// Allow even if there's no payment record
		if err != nil && err != pgx.ErrNoRows {
			return fmt.Errorf("error Payment GetLatestPaymentDueDateByInvoiceID: %v", err)
		}

		if payment != nil {
			if err := s.updatePaymentStatus(ctx, tx, payment.PaymentID.String, invoice_pb.PaymentStatus_PAYMENT_FAILED.String()); err != nil {
				return fmt.Errorf("error Payment Update: %v", err)
			}
		}

		// Step 1b: Update invoice status to void status
		if err := s.updateInvoiceStatus(ctx, tx, req.InvoiceId, invoice_pb.InvoiceStatus_VOID.String()); err != nil {
			return fmt.Errorf("error Invoice Update: %v", err)
		}

		// Step 1c: Create action log
		actionDetails := &InvoiceActionLogDetails{
			InvoiceID:     req.InvoiceId,
			Action:        invoice_pb.InvoiceAction_INVOICE_VOIDED,
			ActionComment: req.Remarks,
		}
		if err := s.createActionLog(ctx, tx, actionDetails); err != nil {
			return err
		}

		// Step 2: Call payment service to update bill items

		// Step 2a: Retrieve invoice bill items
		invoiceBillItems, err := s.InvoiceBillItemRepo.FindAllByInvoiceID(ctx, tx, req.InvoiceId)

		// Allow even if there's no invoice bill item records
		if err != nil && err != pgx.ErrNoRows {
			return fmt.Errorf("error Invoice Bill Item FindAllByInvoiceID: %v", err)
		}

		// Step 2b: Update bill items via payment service
		// Service call at the last step ensures no roll back from the other service is needed
		if err := revertBillItemStatusForVoidInvoice(ctx, tx, s, invoiceBillItems.ToArray()); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return failedResp, status.Error(codes.Internal, err.Error())
	}

	return successfulResp, nil
}

func validateVoidInvoiceRequest(ctx context.Context, db database.QueryExecer, s *InvoiceModifierService, invoiceID string) error {
	if len(invoiceID) == 0 {
		return status.Error(codes.InvalidArgument, "invoiceID is required")
	}

	invoice, err := s.InvoiceRepo.RetrieveInvoiceByInvoiceID(ctx, db, invoiceID)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("error Invoice RetrieveInvoiceByInvoiceID: %v", err))
	}

	// Don't allow an invoice with invalid status to be voided
	switch invoice.Status.String {
	case invoice_pb.InvoiceStatus_VOID.String(), invoice_pb.InvoiceStatus_PAID.String(), invoice_pb.InvoiceStatus_REFUNDED.String():
		return status.Error(codes.InvalidArgument, "Invoice should be in DRAFT, ISSUED, or FAILED status")
	}

	return nil
}
