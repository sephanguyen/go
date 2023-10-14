package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

func (s *suite) prepareDataValidOrderRequestForGraduation(ctx context.Context) (context.Context, error) {
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
	req := s.validGraduateRequestDisabledProrating(&insertOrderReq, billItems)
	stepState.Request = &req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eventPublishedSignalGraduationOrderSubmitted(ctx context.Context) (context.Context, error) {
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
			if orderEventLog.OrderType != pb.OrderType_ORDER_TYPE_GRADUATE.String() && orderEventLog.StudentID != req.StudentId {
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
