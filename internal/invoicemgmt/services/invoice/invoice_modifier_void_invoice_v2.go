package invoicesvc

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *InvoiceModifierService) VoidInvoiceV2(ctx context.Context, req *invoice_pb.VoidInvoiceRequestV2) (*invoice_pb.VoidInvoiceResponseV2, error) {
	failedResp := &invoice_pb.VoidInvoiceResponseV2{
		Successful: false,
	}

	successfulResp := &invoice_pb.VoidInvoiceResponseV2{
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
			if err := s.updatePaymentStatusWithZeroAmount(ctx, tx, payment.PaymentID.String, invoice_pb.PaymentStatus_PAYMENT_FAILED.String()); err != nil {
				return fmt.Errorf("error Payment Update: %v", err)
			}

			// Check if payment belongs to bulk and cancel if all payments are FAILED
			err = s.validateAndCancelBulkPaymentRecord(ctx, tx, payment)
			if err != nil {
				return status.Error(codes.Internal, err.Error())
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

		// Step 2: Call order(payment) service to update bill items

		// Step 2a: Retrieve invoice bill items
		invoiceBillItems, err := s.InvoiceBillItemRepo.FindAllByInvoiceID(ctx, tx, req.InvoiceId)
		if err != nil {
			return fmt.Errorf("error Invoice Bill Item FindAllByInvoiceID: %v", err)
		}

		if len(invoiceBillItems.ToArray()) == 0 {
			return fmt.Errorf("error no Invoice Bill Item records found")
		}

		// Step 2b: Update bill items via order(payment) service
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

func (s *InvoiceModifierService) validateAndCancelBulkPaymentRecord(ctx context.Context, db database.QueryExecer, payment *entities.Payment) error {
	if payment.BulkPaymentID.Status == pgtype.Present && strings.TrimSpace(payment.BulkPaymentID.String) != "" {
		// check if there's other payments that are not failed
		paymentCount, err := s.PaymentRepo.CountOtherPaymentsByBulkPaymentIDNotInStatus(ctx, db, payment.BulkPaymentID.String, payment.PaymentID.String, invoice_pb.PaymentStatus_PAYMENT_FAILED.String())
		if err != nil {
			return fmt.Errorf("PaymentRepo.CountOtherPaymentsByBulkPaymentIDNotInStatus err: %v", err)
		}

		// if there's no other payments that are not failed in status aside from the current payment; cancel the bulk payment record
		if paymentCount == 0 {
			if err := s.BulkPaymentRepo.UpdateBulkPaymentStatusByIDs(ctx, db, invoice_pb.BulkPaymentStatus_BULK_PAYMENT_CANCELLED.String(), []string{payment.BulkPaymentID.String}); err != nil {
				return fmt.Errorf("error BulkPaymentRepo UpdateBulkPaymentStatusByIDs: %v", err)
			}
		}
	}
	return nil
}
