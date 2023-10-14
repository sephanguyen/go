package invoicesvc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"go.uber.org/multierr"
)

func (s *InvoiceModifierService) updateInvoiceStatus(ctx context.Context, db database.QueryExecer, invoiceID string, newStatus string) error {
	e := new(entities.Invoice)
	database.AllNullEntity(e)

	err := multierr.Combine(
		e.InvoiceID.Set(invoiceID),
		e.Status.Set(newStatus),
	)

	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	if err := s.InvoiceRepo.Update(ctx, db, e); err != nil {
		return err
	}

	return nil
}

func (s *InvoiceModifierService) updateInvoiceStatusAndExportedTag(ctx context.Context, db database.QueryExecer, invoiceID string, newStatus string, isExported bool) error {
	e := new(entities.Invoice)
	database.AllNullEntity(e)

	err := multierr.Combine(
		e.InvoiceID.Set(invoiceID),
		e.Status.Set(newStatus),
		e.IsExported.Set(isExported),
	)

	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	if err := s.InvoiceRepo.UpdateWithFields(ctx, db, e, []string{"status", "is_exported", "updated_at"}); err != nil {
		return err
	}

	return nil
}

func (s *InvoiceModifierService) updatePaymentStatus(ctx context.Context, db database.QueryExecer, paymentID string, newStatus string) error {
	e := new(entities.Payment)
	database.AllNullEntity(e)

	err := multierr.Combine(
		e.PaymentID.Set(paymentID),
		e.PaymentStatus.Set(newStatus),
	)

	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	if err := s.PaymentRepo.Update(ctx, db, e); err != nil {
		return err
	}

	return nil
}

func (s *InvoiceModifierService) generateActionLogDetails(invoice *entities.Invoice, payment *entities.Payment, remarks string) *InvoiceActionLogDetails {
	actionDetails := &InvoiceActionLogDetails{}

	// For now, PAID and REFUNDED statuses are the only defined in the switch-statement;
	// The rest of the statuses will be defined in other PRs
	switch invoice.Status.String {
	case invoice_pb.InvoiceStatus_PAID.String():
		actionDetails.Action = invoice_pb.InvoiceAction_INVOICE_PAID
	case invoice_pb.InvoiceStatus_REFUNDED.String():
		actionDetails.Action = invoice_pb.InvoiceAction_INVOICE_REFUNDED
	case invoice_pb.InvoiceStatus_ISSUED.String():
		actionDetails.Action = invoice_pb.InvoiceAction_INVOICE_ISSUED
	}

	actionDetails.InvoiceID = invoice.InvoiceID.String
	actionDetails.ActionComment = remarks
	actionDetails.PaymentSequenceNumber = payment.PaymentSequenceNumber.Int
	actionDetails.PaymentMethod = payment.PaymentMethod.String

	return actionDetails
}

// for void invoice v2 update payment amount to 0
// modify this logic to call new payment service to update status once integrated
func (s *InvoiceModifierService) updatePaymentStatusWithZeroAmount(ctx context.Context, db database.QueryExecer, paymentID string, newStatus string) error {
	e := new(entities.Payment)
	database.AllNullEntity(e)

	err := multierr.Combine(
		e.PaymentID.Set(paymentID),
		e.PaymentStatus.Set(newStatus),
		e.Amount.Set(0),
	)

	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	err = s.PaymentRepo.UpdateWithFields(ctx, db, e, []string{"payment_status", "amount", "updated_at"})
	if err != nil {
		return fmt.Errorf("error Payment UpdateWithFields: %v", err)
	}

	return nil
}

func revertBillItemStatusForVoidInvoice(ctx context.Context, db database.QueryExecer, s *InvoiceModifierService, invoiceBillItems []*entities.InvoiceBillItem) error {
	if len(invoiceBillItems) != 0 {
		// Holds bill items to be sent to payment service
		updateBillItemsReq := make([]*payment_pb.UpdateBillItemStatusRequest_UpdateBillItem, 0, len(invoiceBillItems))

		// Builds the request for each invoice bill item
		for _, invoiceBillItem := range invoiceBillItems {
			// Retrieve bill item
			billItem, err := s.BillItemRepo.FindByID(ctx, db, invoiceBillItem.BillItemSequenceNumber.Int)

			if err != nil {
				return fmt.Errorf("error Bill Item FindByID: %v", err)
			}

			// New bill item status will be the previous bill item status
			// Previous bill item status can only be billed and pending
			billingStatusTo := selectBillingStatusTo(invoiceBillItem, billItem)

			req := &payment_pb.UpdateBillItemStatusRequest_UpdateBillItem{
				BillItemSequenceNumber: billItem.BillItemSequenceNumber.Int,
				BillingStatusTo:        billingStatusTo,
			}

			updateBillItemsReq = append(updateBillItemsReq, req)
		}

		// use order internal service for updating bill item status on voiding an invoice
		userInfo := golibs.UserInfoFromCtx(ctx)
		changeBillItemStatusesRequest := &payment_pb.UpdateBillItemStatusRequest{
			UpdateBillItems: updateBillItemsReq,
			OrganizationId:  userInfo.ResourcePath,
			CurrentUserId:   userInfo.UserID,
		}

		resp, err := s.InternalOrderService.UpdateBillItemStatus(ctx, changeBillItemStatusesRequest)
		if err != nil {
			return fmt.Errorf("error UpdateBillItemStatus: %v", err)
		}

		// Check for any validation errors from the payment service
		if len(resp.Errors) > 0 {
			var errorList []string

			for _, err := range resp.Errors {
				errorStr := fmt.Sprintf("BillItemSequenceNumber %v with error %v", err.BillItemSequenceNumber, err.Error)

				errorList = append(errorList, errorStr)
			}

			return fmt.Errorf("error UpdateBillItemStatus: %v", strings.Join(errorList, ","))
		}
	}

	return nil
}

func selectBillingStatusTo(invoiceBillItem *entities.InvoiceBillItem, billItem *entities.BillItem) payment_pb.BillingStatus {
	billingStatusTo := payment_pb.BillingStatus_BILLING_STATUS_BILLED
	if invoiceBillItem.PastBillingStatus.String == payment_pb.BillingStatus_BILLING_STATUS_PENDING.String() && billItem.BillDate.Time.After(time.Now().UTC()) {
		billingStatusTo = payment_pb.BillingStatus_BILLING_STATUS_PENDING
	}

	return billingStatusTo
}
