package invoicemgmt

import (
	"context"
	"fmt"
	"time"

	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
)

func (s *suite) invoiceHasStatusWithBillItemsCount(ctx context.Context, invoiceStatus string, expectedBillItemCount int32) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var err error

	// Update invoice status
	if ctx, err = s.updateInvoiceStatus(ctx, invoiceStatus); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// Create payment record
	if ctx, err = s.createPayment(ctx, invoice_pb.PaymentMethod_DIRECT_DEBIT, invoice_pb.PaymentStatus_PAYMENT_PENDING.String(), "", false); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// Remove existing bill items
	if expectedBillItemCount == 0 {
		if err := s.deleteBillItemsByInvoiceID(ctx, stepState.InvoiceID); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	// Deduct 1 as there's already a bill item created before this
	for i := 0; i < int(expectedBillItemCount)-1; i++ {

		if ctx, err = s.createBillItemBasedOnStatusAndType(ctx, "BILLED", payment_pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER.String()); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if ctx, err = s.createInvoiceBillItem(ctx, "PENDING"); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) loggedinUserViewsAnInvoice(ctx context.Context) (context.Context, error) {
	time.Sleep(invoiceConst.KafkaSyncSleepDuration)

	stepState := StepStateFromContext(ctx)

	req := &invoice_pb.RetrieveInvoiceInfoRequest{
		InvoiceIdString: s.InvoiceID,
	}

	s.StepState.Response, s.StepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).RetrieveInvoiceInfo(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) receivesBillItemsCount(ctx context.Context, expectedBillItemCount int32) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp := s.StepState.Response.(*invoice_pb.RetrieveInvoiceInfoResponse)

	if len(resp.BillItems) != int(expectedBillItemCount) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %v bill item count but got %v for invoice_id %v", expectedBillItemCount, len(resp.BillItems), s.StepState.InvoiceID)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisParentHasAnExistingStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.createStudent(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.createStudentParentRelationship(StepStateToContext(ctx, stepState), upb.FamilyRelationship_name[int32(upb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER)])
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
