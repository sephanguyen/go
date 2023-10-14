package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/payment/constant"

	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

func (s *suite) getProductListOfOrder(ctx context.Context, account string, orderTypeTestcase string, filterTestcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	var (
		req     *pb.RetrieveListOfOrderDetailProductsRequest
		orderID string
	)
	switch orderTypeTestcase {
	case "custom billing":
		createOrderResp := stepState.Response.(*pb.CreateCustomBillingResponse)
		orderID = createOrderResp.OrderId
	default:
		createOrderResp := stepState.Response.(*pb.CreateOrderResponse)
		orderID = createOrderResp.OrderId
	}

	switch filterTestcase {
	case "non-exists order":
		req = &pb.RetrieveListOfOrderDetailProductsRequest{
			OrderId: constant.InvalidOrderID,
			Paging: &cpb.Paging{
				Limit: 2,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 0,
				},
			},
		}
	case "empty":
		req = &pb.RetrieveListOfOrderDetailProductsRequest{
			OrderId: "",
			Paging: &cpb.Paging{
				Limit: 2,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 0,
				},
			},
		}
	case "valid":
		req = &pb.RetrieveListOfOrderDetailProductsRequest{
			OrderId: orderID,
			Paging: &cpb.Paging{
				Limit: 2,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 0,
				},
			},
		}
	}

	client := pb.NewOrderServiceClient(s.PaymentConn)
	stepState.RequestSentAt = time.Now()
	resp, err := client.RetrieveListOfOrderDetailProducts(contextWithToken(ctx), req)
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Response = resp

	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("an unexpected error returned")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkProductListOfOrderResponse(ctx context.Context, typeResponseTestcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp := stepState.Response.(*pb.RetrieveListOfOrderDetailProductsResponse)
	switch typeResponseTestcase {
	case "empty":
		if len(resp.Items) != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong count item response")
		}
	case "non-empty":
		expectedResponse := &pb.RetrieveListOfOrderDetailProductsResponse{
			Items: []*pb.RetrieveListOfOrderDetailProductsResponse_OrderProduct{},
			NextPage: &cpb.Paging{
				Limit: 2,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 2,
				},
			},
			PreviousPage: nil,
			TotalItems:   6,
		}
		for _, item := range resp.Items {
			if item.ProductId == "" {
				return StepStateToContext(ctx, stepState), fmt.Errorf("wrong productID")
			}
			if item.ProductType == 0 {
				return StepStateToContext(ctx, stepState), fmt.Errorf("wrong productType")
			}
		}

		if resp.NextPage != nil && expectedResponse.NextPage != nil {
			if resp.NextPage.GetOffsetInteger() != expectedResponse.NextPage.GetOffsetInteger() {
				return StepStateToContext(ctx, stepState), fmt.Errorf("wrong NextPage.Offset")
			}

			if resp.NextPage.Limit != expectedResponse.NextPage.Limit {
				return StepStateToContext(ctx, stepState), fmt.Errorf("wrong NextPage.Limit")
			}
		}

		if resp.PreviousPage != nil && expectedResponse.PreviousPage != nil {
			if resp.PreviousPage.GetOffsetInteger() != expectedResponse.PreviousPage.GetOffsetInteger() {
				return StepStateToContext(ctx, stepState), fmt.Errorf("wrong PreviousPage.Offset, %v", resp.PreviousPage.GetOffsetInteger())
			}

			if resp.PreviousPage.Limit != expectedResponse.PreviousPage.Limit {
				return StepStateToContext(ctx, stepState), fmt.Errorf("wrong PreviousPage.Limit")
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
