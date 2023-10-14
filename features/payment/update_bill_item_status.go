package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

func (s *suite) thereIsAnExistingBillItemsFromOrderRecords(ctx context.Context, product string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch product {
	case "material type one time":
		_, err := s.prepareDataForGetOderItemsList(ctx, "new", product)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "order with discount and prorating":
		_, err := s.prepareDataForCreateOrderRecurringFeeWithValidRequest(ctx, product)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	ctx, err := s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.RequestSentAt = time.Now()
	stepState.RequestSentAt = time.Now()
	req := stepState.Request.(*pb.CreateOrderRequest)
	// by default bill items have status Billed/Pending
	stepState.Response, stepState.ResponseErr = pb.NewOrderServiceClient(s.PaymentConn).CreateOrder(contextWithToken(ctx), req)

	for stepState.ResponseErr != nil && strings.Contains(stepState.ResponseErr.Error(), "(SQLSTATE 23505)") {
		time.Sleep(1000)
		stepState.Response, stepState.ResponseErr = pb.NewOrderServiceClient(s.PaymentConn).
			CreateOrder(contextWithToken(ctx), stepState.Request.(*pb.CreateOrderRequest))
	}
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) requestPayloadToUpdateBillItemsStatus(ctx context.Context, typeOfRequest string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	orderResp := stepState.Response.(*pb.CreateOrderResponse)
	billItemResp, err := s.getBillItems(ctx, orderResp.OrderId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	var reqBillItems []*pb.UpdateBillItemStatusRequest_UpdateBillItem
	switch typeOfRequest {
	case "invoiced":
		for _, billItem := range billItemResp {
			reqBillItems = append(reqBillItems, &pb.UpdateBillItemStatusRequest_UpdateBillItem{
				BillItemSequenceNumber: billItem.BillItemSequenceNumber.Int,
				BillingStatusTo:        pb.BillingStatus_BILLING_STATUS_INVOICED,
			})
		}
	case "billed":
		for _, billItem := range billItemResp {
			reqBillItems = append(reqBillItems, &pb.UpdateBillItemStatusRequest_UpdateBillItem{
				BillItemSequenceNumber: billItem.BillItemSequenceNumber.Int,
				BillingStatusTo:        pb.BillingStatus_BILLING_STATUS_BILLED,
			})
		}
	case "pending":
		for _, billItem := range billItemResp {
			reqBillItems = append(reqBillItems, &pb.UpdateBillItemStatusRequest_UpdateBillItem{
				BillItemSequenceNumber: billItem.BillItemSequenceNumber.Int,
				BillingStatusTo:        pb.BillingStatus_BILLING_STATUS_PENDING,
			})
		}
	case "invalid":
		reqBillItems = []*pb.UpdateBillItemStatusRequest_UpdateBillItem{
			{
				BillItemSequenceNumber: billItemResp[0].BillItemSequenceNumber.Int,
				BillingStatusTo:        66,
			},
			{
				BillItemSequenceNumber: billItemResp[1].BillItemSequenceNumber.Int,
				BillingStatusTo:        100,
			},
		}
	}

	stepState.Request = &pb.UpdateBillItemStatusRequest{
		UpdateBillItems: reqBillItems,
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) submittedTheRequest(ctx context.Context, role string, service string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return ctx, err
	}

	stepState.RequestSentAt = time.Now()

	switch service {
	case "order":
		stepState.Response, stepState.ResponseErr = pb.NewOrderServiceClient(s.PaymentConn).UpdateBillItemStatus(contextWithToken(ctx), stepState.Request.(*pb.UpdateBillItemStatusRequest))
	case "internal":
		resourcePath, err := interceptors.ResourcePathFromContext(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		// Assign the resource path to OrganizationID and the current user ID to the request
		req := stepState.Request.(*pb.UpdateBillItemStatusRequest)
		req.OrganizationId = resourcePath
		req.CurrentUserId = stepState.CurrentUserID
		stepState.Response, stepState.ResponseErr = pb.NewInternalServiceClient(s.PaymentConn).UpdateBillItemStatus(contextWithToken(ctx), req)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) billItemsStatusAreUpdated(ctx context.Context, updateStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, billItem := range stepState.Request.(*pb.UpdateBillItemStatusRequest).UpdateBillItems {
		var queryBillItemStatus string

		stmt := `
		SELECT
			billing_status
		FROM
			bill_item
		WHERE
			bill_item_sequence_number = $1
		`
		billItemRow := s.FatimaDBTrace.QueryRow(ctx, stmt, billItem.BillItemSequenceNumber)
		err := billItemRow.Scan(&queryBillItemStatus)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("bill item record not found in fatima db")
		}
		switch updateStatus {
		case "successfully":
			if queryBillItemStatus != billItem.BillingStatusTo.String() {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot update status in bill item sequence number: %v", billItem.BillItemSequenceNumber)
			}
		case "unsuccessfully":
			if queryBillItemStatus == billItem.BillingStatusTo.String() {
				return StepStateToContext(ctx, stepState), fmt.Errorf("error should not update status in bill item sequence number: %v", billItem.BillItemSequenceNumber)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) responseHasNoErrors(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.Response != nil {
		if len(stepState.Response.(*pb.UpdateBillItemStatusResponse).Errors) != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected error in response")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) invoicedBillItemsWillHaveInvoicedOrder(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, billItem := range stepState.Request.(*pb.UpdateBillItemStatusRequest).UpdateBillItems {
		if billItem.BillingStatusTo != pb.BillingStatus_BILLING_STATUS_INVOICED {
			continue
		}

		var queryOrderId string

		stmt := `
		SELECT
			order_id
		FROM
			bill_item
		WHERE
			bill_item_sequence_number = $1
		`
		billItemRow := s.FatimaDBTrace.QueryRow(ctx, stmt, billItem.BillItemSequenceNumber)
		err := billItemRow.Scan(&queryOrderId)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("bill item record not found in fatima db")
		}

		order, err := s.getOrder(ctx, queryOrderId)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if order.OrderStatus.String != pb.OrderStatus_ORDER_STATUS_INVOICED.String() {
			return StepStateToContext(ctx, stepState), fmt.Errorf("fail to update order status to invoiced")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
