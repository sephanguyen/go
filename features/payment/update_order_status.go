package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

func (s *suite) orderStatusUpdated(ctx context.Context, updateStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, order := range stepState.Request.(*pb.UpdateOrderStatusRequest).UpdateOrdersStatuses {
		queryOrder, err := s.getOrder(ctx, order.OrderId)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("order record not found in fatima db")
		}

		queryOrderStatus := queryOrder.OrderStatus.String

		switch updateStatus {
		case "successfully":
			if queryOrderStatus != order.OrderStatus.String() {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot update status in order id: %v", order.OrderId)
			}
		case "unsuccessfully":
			if queryOrderStatus == order.OrderStatus.String() {
				return StepStateToContext(ctx, stepState), fmt.Errorf("error should not update status in order id: %v", order.OrderId)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) requestPayloadToUpdateOrderStatus(ctx context.Context, typeOfRequest string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	orderResp := stepState.Response.(*pb.CreateOrderResponse)
	var ordersToBeUpdated []*pb.UpdateOrderStatusRequest_UpdateOrderStatus
	switch typeOfRequest {
	case "invoiced":
		{
			ordersToBeUpdated = []*pb.UpdateOrderStatusRequest_UpdateOrderStatus{
				{
					OrderId:     orderResp.OrderId,
					OrderStatus: pb.OrderStatus_ORDER_STATUS_INVOICED,
				},
			}
		}
	case "submitted":
		{
			ordersToBeUpdated = []*pb.UpdateOrderStatusRequest_UpdateOrderStatus{
				{
					OrderId:     orderResp.OrderId,
					OrderStatus: pb.OrderStatus_ORDER_STATUS_SUBMITTED,
				},
			}
		}
	case "invalid":
		{
			ordersToBeUpdated = []*pb.UpdateOrderStatusRequest_UpdateOrderStatus{
				{
					OrderId:     orderResp.OrderId,
					OrderStatus: pb.OrderStatus_ORDER_STATUS_ALL,
				},
			}
		}
	default:
	}

	stepState.Request = &pb.UpdateOrderStatusRequest{
		UpdateOrdersStatuses: ordersToBeUpdated,
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsAnExistingOrderFromOrderRecords(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.prepareDataForGetOderItemsList(ctx, "new", "material type one time")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.RequestSentAt = time.Now()
	stepState.RequestSentAt = time.Now()
	req := stepState.Request.(*pb.CreateOrderRequest)
	// by default, order have status ORDER_STATUS_SUBMITTED
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

func (s *suite) submittedTheUpdateOrderRequest(ctx context.Context, role string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, role)
	if err != nil {
		return ctx, err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewOrderServiceClient(s.PaymentConn).UpdateOrderStatus(contextWithToken(ctx), stepState.Request.(*pb.UpdateOrderStatusRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateOrderStatusResponseHasNoErrors(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.Response != nil {
		if len(stepState.Response.(*pb.UpdateOrderStatusResponse).Errors) != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected error in response")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
