package invoicemgmt

import (
	"context"
	"fmt"
	"time"

	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

func (s *suite) paymentEndpointIsCalledToUpdateTheseBillItemsStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var queryBillItemStatus string
	stmt := `
		SELECT
			billing_status
		FROM
			bill_item
		WHERE
			bill_item_sequence_number = $1
		`
	billItemRow := s.FatimaDBTrace.QueryRow(ctx, stmt, s.StepState.BillItemSequenceNumbers[0])
	err := billItemRow.Scan(&queryBillItemStatus)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("bill item record not found in fatima db")
	}

	ctx, err = s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return ctx, err
	}
	req := &payment_pb.UpdateBillItemStatusRequest{
		UpdateBillItems: []*payment_pb.UpdateBillItemStatusRequest_UpdateBillItem{
			{
				BillItemSequenceNumber: int32(s.StepState.BillItemSequenceNumbers[0]),
				BillingStatusTo:        payment_pb.BillingStatus_BILLING_STATUS_INVOICED,
			},
		},
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = payment_pb.NewInternalServiceClient(s.PaymentConn).UpdateBillItemStatus(contextWithToken(ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsAnExistingBillItemsCreatedOnPayment(ctx context.Context) (context.Context, error) {
	return s.createStudentWithBillItem(ctx, payment_pb.BillingStatus_BILLING_STATUS_BILLED.String(), payment_pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER.String())
}

func (s *suite) receivesStatusCode(ctx context.Context, expectedCode string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.CommonSuite.ReturnsStatusCode(StepStateToContext(ctx, stepState), expectedCode)
	if err != nil {
		return ctx, err
	}
	return StepStateToContext(ctx, stepState), nil
}
