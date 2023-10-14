package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/payment/constant"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *suite) orderReviewedFlagUpdated(ctx context.Context, updateStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	queryOrder, err := s.getOrder(ctx, stepState.Request.(*pb.UpdateOrderReviewedFlagRequest).OrderId)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("order record not found in db")
	}

	expectStatus := stepState.Request.(*pb.UpdateOrderReviewedFlagRequest).IsReviewed

	isReviewed := queryOrder.IsReviewed
	switch updateStatus {
	case "successfully":
		if expectStatus != isReviewed.Bool {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot update reviewed flag in order id: %v", queryOrder.OrderID)
		}
	case "unsuccessfully":
		if expectStatus == isReviewed.Bool {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error should not update reivewed flag success")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) requestPayloadToUpdateOrderReviewedFlag(ctx context.Context, typeOfRequest string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	orderResp := stepState.Response.(*pb.CreateOrderResponse)

	switch typeOfRequest {
	case "true":
		{
			stepState.Request = &pb.UpdateOrderReviewedFlagRequest{
				OrderId:            orderResp.OrderId,
				IsReviewed:         true,
				OrderVersionNumber: int32(0),
			}
		}

	case "false":
		{
			stepState.Request = &pb.UpdateOrderReviewedFlagRequest{
				OrderId:            orderResp.OrderId,
				IsReviewed:         false,
				OrderVersionNumber: int32(0),
			}
		}

	default:
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anExistingOrderFromOrderRecords(ctx context.Context) (context.Context, error) {
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
	req := stepState.Request.(*pb.CreateOrderRequest)
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

func (s *suite) submittedTheUpdateOrderReviewedFlagRequest(ctx context.Context, role string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, role)
	if err != nil {
		return ctx, err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewOrderServiceClient(s.PaymentConn).UpdateOrderReviewedFlag(contextWithToken(ctx), stepState.Request.(*pb.UpdateOrderReviewedFlagRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateOrderReviewedFlagResponseSuccess(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.Response != nil {
		if !stepState.Response.(*pb.UpdateOrderReviewedFlagResponse).Successful {
			return StepStateToContext(ctx, stepState), fmt.Errorf("update flag is failed")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) requestOutOfVersionPayloadToUpdateOrderReviewedFlag(ctx context.Context, typeOfRequest string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	orderResp := stepState.Response.(*pb.CreateOrderResponse)

	switch typeOfRequest {
	case "true":
		{
			stepState.Request = &pb.UpdateOrderReviewedFlagRequest{
				OrderId:            orderResp.OrderId,
				IsReviewed:         true,
				OrderVersionNumber: int32(5),
			}
		}

	case "false":
		{
			stepState.Request = &pb.UpdateOrderReviewedFlagRequest{
				OrderId:            orderResp.OrderId,
				IsReviewed:         false,
				OrderVersionNumber: int32(5),
			}
		}

	default:
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateOrderReviewedFlagResponseUnsuccess(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr.Error() != status.Error(codes.FailedPrecondition, constant.OptimisticLockingEntityVersionMismatched).Error() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("valid version number")
	}

	return StepStateToContext(ctx, stepState), nil
}
