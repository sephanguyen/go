package discount

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/manabie-com/backend/features/discount/mockdata"
	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	paymentPb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

func (s *suite) prepareDataForSiblingDiscountAutomationForCase(ctx context.Context, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	_, _, err := mockdata.InsertOrgDiscount(ctx, s.FatimaDBTrace, paymentPb.DiscountType_DISCOUNT_TYPE_SIBLING.String())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	switch testcase {
	case "without sibling":
		studentID, err := mockdata.InsertOneStudent(ctx, s.FatimaDBTrace, "1")
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		stepState.StudentID = studentID

		_, productIDs, err := mockdata.InsertProductGroupMappingForSpecialDiscount(ctx, s.FatimaDBTrace, paymentPb.DiscountType_DISCOUNT_TYPE_SIBLING.String())
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		stepState.ProductID = productIDs[0]
	case "not valid for tracking":
		studentIDs, err := mockdata.InsertNSiblingsAndReturnIDs(ctx, s.FatimaDBTrace, 3)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		stepState.StudentID = studentIDs[0]

		productIDs, err := mockdata.InsertRecurringProducts(ctx, s.FatimaDBTrace)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		stepState.ProductID = productIDs[0]
	case "valid for tracking but not valid for tagging":
		studentIDs, err := mockdata.InsertNSiblingsAndReturnIDs(ctx, s.FatimaDBTrace, 3)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		stepState.StudentID = studentIDs[0]

		_, productIDs, err := mockdata.InsertProductGroupMappingForSpecialDiscount(ctx, s.FatimaDBTrace, paymentPb.DiscountType_DISCOUNT_TYPE_SIBLING.String())
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		stepState.ProductID = productIDs[0]
	case "valid for tagging":
		studentIDs, err := mockdata.InsertNSiblingsAndReturnIDs(ctx, s.FatimaDBTrace, 3)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		stepState.StudentID = studentIDs[0]
		_, productIDs, err := mockdata.InsertProductGroupMappingForSpecialDiscount(ctx, s.FatimaDBTrace, paymentPb.DiscountType_DISCOUNT_TYPE_SIBLING.String())
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		for idx, studentID := range studentIDs {
			if idx != 0 {
				orderID, locationID, studentProductIDs, err := mockdata.InsertOrderForStudentWithProducts(ctx, s.FatimaDBTrace, studentID, productIDs)
				if err != nil {
					return StepStateToContext(ctx, stepState), err
				}

				orderInfoLog := entities.OrderWithProductInfoLog{
					OrderID:           orderID,
					StudentID:         studentID,
					LocationID:        locationID,
					OrderStatus:       paymentPb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
					OrderType:         paymentPb.OrderType_ORDER_TYPE_NEW.String(),
					StudentProductIDs: studentProductIDs,
				}

				data, err := json.Marshal(orderInfoLog)
				if err != nil {
					return StepStateToContext(ctx, stepState), err
				}

				msgID, err := s.JSM.PublishAsyncContext(ctx, constants.SubjectOrderWithProductInfoLogCreated, data)
				if err != nil {
					return StepStateToContext(ctx, stepState), nats.HandlePushMsgFail(ctx, fmt.Errorf("SubjectOrderWithProductInfoLogCreated JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, err))
				}
			}
		}

		stepState.ProductID = productIDs[0]
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) serviceReceivedOrderInfoStream(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentID := stepState.StudentID
	productID := stepState.ProductID

	orderID, locationID, studentProductIDs, err := mockdata.InsertOrderForStudentWithProducts(ctx, s.FatimaDBTrace, studentID, []string{productID})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	orderInfoLog := entities.OrderWithProductInfoLog{
		OrderID:           orderID,
		StudentID:         studentID,
		LocationID:        locationID,
		OrderStatus:       paymentPb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
		OrderType:         paymentPb.OrderType_ORDER_TYPE_NEW.String(),
		StudentProductIDs: studentProductIDs,
	}

	data, err := json.Marshal(orderInfoLog)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	msgID, err := s.JSM.PublishAsyncContext(ctx, constants.SubjectOrderWithProductInfoLogCreated, data)
	if err != nil {
		return StepStateToContext(ctx, stepState), nats.HandlePushMsgFail(ctx, fmt.Errorf("SubjectOrderWithProductInfoLogCreated JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, err))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentTrackedForSiblingDiscount(ctx context.Context, trackingEvent string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if trackingEvent == "not tracked" {
		isTracked := s.checkStudentDiscountTrackerByStudentID(ctx, stepState.StudentID)
		if isTracked {
			return StepStateToContext(ctx, stepState), fmt.Errorf("incorrect data for student discount tracker")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
