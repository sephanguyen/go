package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

func (s *suite) createOrderSuccessfully(
	ctx context.Context, account string, orderType string, productType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.prepareDataForCreateOrderWithProductType(ctx, orderType, productType)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.createOrderWithOrderType(ctx, account, orderType, productType)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareDataForCreateOrderWithProductType(ctx context.Context, orderType string, productType string) (context.Context, error) {
	switch productType {
	case "one time package":
		return s.prepareDataForCreateOrderOneTimePackage(ctx)
	case "one time material":
		return s.prepareDataForCreateOrderOneTimeMaterial(ctx, "product discount")
	case "one time fee":
		return s.prepareDataForCreateOrderOneTimeFee(ctx)
	case "recurring fee":
		return s.prepareDataForCreateOrderRecurringFeeWithValidRequest(ctx, "order with multiple billed at order item")
	case "recurring package":
		return s.prepareDataForCreateOrderRecurringPackage(ctx)
	case "recurring material":
		if orderType == "withdrawal" {
			return s.prepareDataForCreateOrderWithdrawRecurringMaterial(ctx, "valid withdrawal request with disabled prorating")
		} else if orderType == "graduate" {
			return s.prepareDataForCreateOrderWithdrawRecurringMaterial(ctx, "valid graduate request with disabled prorating")
		} else {
			return s.prepareDataForCreateOrderRecurringMaterial(ctx)
		}
	case "custom billing":
		return s.prepareDataForCreatingCustomBilling(ctx)
	}
	return ctx, fmt.Errorf("invalid product type: %v", productType)
}

func (s *suite) createOrderWithOrderType(ctx context.Context, account string, orderType string, productType string) (context.Context, error) {
	client := pb.NewOrderServiceClient(s.PaymentConn)

	stepState := StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()
	var err error

	switch orderType {
	case "new":
		ctx, err = s.userSubmitOrder(ctx, account)
		stepState = StepStateFromContext(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), stepState.ResponseErr
		}
	case "update":
		ctx, err = s.userSubmitOrder(ctx, account)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if productType == "one time package" {
			ctx, err = s.prepareDataForUpdateOrderOneTimePackage(ctx)
		} else if productType == "recurring package" {
			ctx, err = s.prepareDataForUpdateOrderRecurringPackage(ctx)
		} else if productType == "recurring material" {
			ctx, err = s.prepareDataForUpdateOrderRecurringMaterial(ctx)
		}
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		ctx, err = s.userSubmitOrder(ctx, account)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "custom billing":
		req := stepState.Request.(*pb.CreateCustomBillingRequest)
		createCustomBillingRes, err := client.CreateCustomBilling(contextWithToken(ctx), req)
		for err != nil && strings.Contains(err.Error(), "(SQLSTATE 23505)") {
			time.Sleep(1000)
			createCustomBillingRes, err = pb.NewOrderServiceClient(s.PaymentConn).
				CreateCustomBilling(contextWithToken(ctx), stepState.Request.(*pb.CreateCustomBillingRequest))
		}
		if err != nil {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), err
		}
		stepState.Response = createCustomBillingRes
	case "withdrawal", "graduate":
		ctx, err = s.userSubmitOrder(ctx, account)
		stepState = StepStateFromContext(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), stepState.ResponseErr
		}
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("invalid order type: %v", orderType)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getBillItemsOfOrderDetails(ctx context.Context, account string, orderTypeTestcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("an unexpected error returned in getBillItemsOfOrderDetails()")
	}
	var (
		resp *pb.RetrieveBillingOfOrderDetailsResponse
		req  *pb.RetrieveBillingOfOrderDetailsRequest
	)
	switch orderTypeTestcase {
	case "custom billing":
		createOrderResp := stepState.Response.(*pb.CreateCustomBillingResponse)
		req = &pb.RetrieveBillingOfOrderDetailsRequest{
			OrderId: createOrderResp.OrderId,
			Paging: &cpb.Paging{
				Limit: 2,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 0,
				},
			},
		}

	default:
		createOrderResp := stepState.Response.(*pb.CreateOrderResponse)
		req = &pb.RetrieveBillingOfOrderDetailsRequest{
			OrderId: createOrderResp.OrderId,
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
	resp, err := client.RetrieveBillingOfOrderDetails(contextWithToken(ctx), req)
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Response = resp
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getBillItemsOfOrderDetailsSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("an unexpected error returned")
	}

	resp := stepState.Response.(*pb.RetrieveBillingOfOrderDetailsResponse)

	expectedResponse := &pb.RetrieveBillingOfOrderDetailsResponse{
		Items: []*pb.RetrieveBillingOfOrderDetailsResponse_OrderDetails{},
		NextPage: &cpb.Paging{
			Limit: 2,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 2,
			},
		},
		TotalItems: 6,
	}
	err := s.checkResponseOfBillItem(expectedResponse, resp)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkResponseOfBillItem(expectedResponse *pb.RetrieveBillingOfOrderDetailsResponse, resp *pb.RetrieveBillingOfOrderDetailsResponse) error {
	if len(resp.Items) != 2 {
		return fmt.Errorf("expect 2 items returned")
	}
	for _, item := range resp.Items {
		if item.BillItemSequenceNumber == 0 {
			return fmt.Errorf("wrong BillSequenceNumber")
		}
		if len(item.OrderId) == 0 {
			return fmt.Errorf("wrong OrderID")
		}

		if item.BillingStatus != pb.BillingStatus_BILLING_STATUS_BILLED {
			return fmt.Errorf("wrong BillingStatus: %v", item.BillingStatus)
		}
	}

	if resp.NextPage != nil && expectedResponse.NextPage != nil {
		if resp.NextPage.GetOffsetInteger() != expectedResponse.NextPage.GetOffsetInteger() {
			return fmt.Errorf("wrong NextPage.Offset, %v", resp.NextPage.GetOffsetInteger())
		}

		if resp.NextPage.Limit != expectedResponse.NextPage.Limit {
			return fmt.Errorf("wrong NextPage.Limit")
		}
	}

	if resp.PreviousPage != nil && expectedResponse.PreviousPage != nil {
		if resp.PreviousPage.GetOffsetInteger() != expectedResponse.PreviousPage.GetOffsetInteger() {
			return fmt.Errorf("wrong PreviousPage.Offset, %v", resp.PreviousPage.GetOffsetInteger())
		}

		if resp.PreviousPage.Limit != expectedResponse.PreviousPage.Limit {
			return fmt.Errorf("wrong PreviousPage.Limit")
		}
	}

	return nil
}
