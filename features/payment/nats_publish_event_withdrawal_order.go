package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

func (s *suite) prepareDataValidOrderRequestForWithdrawal(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	defaultOptionPrepareData := mockdata.OptionToPrepareDataForCreateOrderRecurringProduct{
		InsertTax:                     true,
		InsertDiscount:                true,
		InsertStudent:                 true,
		InsertProductPrice:            true,
		InsertProductLocation:         true,
		InsertLocation:                false,
		InsertProductGrade:            true,
		InsertFee:                     false,
		InsertMaterial:                true,
		InsertBillingSchedule:         true,
		InsertBillingScheduleArchived: false,
		IsTaxExclusive:                false,
		InsertDiscountNotAvailable:    false,
		InsertProductOutOfTime:        false,
		BillingScheduleStartDate:      time.Now(),
	}
	var (
		insertOrderReq pb.CreateOrderRequest
		billItems      []*entities.BillItem
	)

	defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
	insertOrderReq, billItems, err := s.createRecurringMaterialDisabledProrating(ctx, defaultOptionPrepareData)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := s.validWithdrawalRequestDisabledProrating(&insertOrderReq, billItems)
	stepState.Request = &req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareDataValidOrderRequestForWithdrawalWithEmptyProduct(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	defaultOptionPrepareData := mockdata.OptionToPrepareDataForCreateOrderRecurringProduct{
		InsertTax:                     true,
		InsertDiscount:                true,
		InsertStudent:                 true,
		InsertProductPrice:            true,
		InsertProductLocation:         true,
		InsertLocation:                false,
		InsertProductGrade:            true,
		InsertFee:                     false,
		InsertMaterial:                true,
		InsertBillingSchedule:         true,
		InsertBillingScheduleArchived: false,
		IsTaxExclusive:                false,
		InsertDiscountNotAvailable:    false,
		InsertProductOutOfTime:        false,
		BillingScheduleStartDate:      time.Now(),
	}
	defaultOptionPrepareData.BillingScheduleStartDate = time.Now()
	req := s.validWithdrawalEmptyProducts(ctx, defaultOptionPrepareData)
	stepState.Request = &req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eventPublishedSignalWithdrawalOrderSubmitted(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	timer := time.NewTimer(time.Minute * 5)
	defer timer.Stop()

	req := stepState.Request.(*pb.CreateOrderRequest)
	for {
		select {
		case data := <-stepState.FoundChanForJetStream:
			orderEventLog, ok := data.(*entities.OrderEventLog)
			if !ok {
				return StepStateToContext(ctx, stepState), fmt.Errorf("fail to parse data to *entities.OrderEventLog")
			}
			if orderEventLog.OrderType != pb.OrderType_ORDER_TYPE_WITHDRAWAL.String() && orderEventLog.StudentID != req.StudentId {
				data, err := json.Marshal(orderEventLog)
				if err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("fail to publish event to NATS")
				}
				_, err = connections.JSM.PublishAsyncContext(ctx, constants.SubjectOrderEventLogCreated, data)
				if err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("fail to publish event to NATS")
				}
				continue
			}
			return StepStateToContext(ctx, stepState), nil
		case <-ctx.Done():
			return StepStateToContext(ctx, stepState), fmt.Errorf("timeout waiting for event to be published")
		case <-timer.C:
			return StepStateToContext(ctx, stepState), errors.New("time out cause of failing")
		}
	}
}

func (s *suite) eventPublishedSignalVoidWithdrawalOrderSubmitted(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	timer := time.NewTimer(time.Minute * 5)
	defer timer.Stop()

	for {
		select {
		case data := <-stepState.FoundChanForJetStream:
			orderEventLog, ok := data.(*entities.OrderEventLog)
			if !ok {
				return StepStateToContext(ctx, stepState), fmt.Errorf("fail to parse data to *entities.OrderEventLog")
			}
			if orderEventLog.OrderType != pb.OrderType_ORDER_TYPE_WITHDRAWAL.String() && orderEventLog.OrderStatus != pb.OrderStatus_ORDER_STATUS_VOIDED.String() {
				data, err := json.Marshal(orderEventLog)
				if err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("fail to publish event to NATS")
				}
				_, err = connections.JSM.PublishAsyncContext(ctx, constants.SubjectOrderEventLogCreated, data)
				if err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("fail to publish event to NATS")
				}
				continue
			}
			return StepStateToContext(ctx, stepState), nil
		case <-ctx.Done():
			return StepStateToContext(ctx, stepState), fmt.Errorf("timeout waiting for event to be published")
		case <-timer.C:
			return StepStateToContext(ctx, stepState), errors.New("time out cause of failing")
		}
	}
}

func (s *suite) voidWithdrawalOrderWithNoProducts(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	createOrderResp := stepState.Response.(*pb.CreateOrderResponse)

	voidOrderReq := &pb.VoidOrderRequest{
		OrderId: createOrderResp.OrderId,
	}

	client := pb.NewOrderServiceClient(s.PaymentConn)
	stepState.RequestSentAt = time.Now()

	voidOrderResp, err := client.VoidOrder(contextWithToken(ctx), voidOrderReq)
	for err != nil && strings.Contains(err.Error(), "(SQLSTATE 23505)") {
		time.Sleep(1000)
		voidOrderResp, err = pb.NewOrderServiceClient(s.PaymentConn).
			VoidOrder(contextWithToken(ctx), voidOrderReq)
	}
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = voidOrderReq
	stepState.Response = voidOrderResp

	return StepStateToContext(ctx, stepState), nil
}
