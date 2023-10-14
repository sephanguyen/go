package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	discountEntities "github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

func (s *suite) prepareDataForRecurringProductForDiscountAutomation(ctx context.Context, product string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch product {
	case "material":
		return s.prepareDataForCreateOrderRecurringMaterial(ctx)
	case "fee":
		return s.prepareDataForCreateOrderRecurringFeeWithValidRequest(ctx, "order with discount and prorating")
	case "frequency-base package":
		return s.prepareDataForCreateOrderFrequencyBasePackage(ctx)
	case "schedule-base package":
		return s.prepareDataForCreateOrderScheduleBasePackage(ctx)
	default:
		err := fmt.Errorf("failed to map product type for discount automation")
		return StepStateToContext(ctx, stepState), err
	}
}

func (s *suite) studentTaggedOrgLevelDiscount(ctx context.Context, discount string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	createOrderReq := stepState.Request.(*pb.CreateOrderRequest)
	var discountType string

	switch discount {
	case "single-parent":
		discountType = pb.DiscountType_DISCOUNT_TYPE_SINGLE_PARENT.String()
	case "family":
		discountType = pb.DiscountType_DISCOUNT_TYPE_FAMILY.String()
	case "employee full-time":
		discountType = pb.DiscountType_DISCOUNT_TYPE_EMPLOYEE_FULL_TIME.String()
	case "employee part-time":
		discountType = pb.DiscountType_DISCOUNT_TYPE_EMPLOYEE_PART_TIME.String()
	}

	discountID, discountTagID, err := mockdata.InsertOrgDiscount(ctx, s.FatimaDBTrace, discountType)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.DiscountID = discountID
	userDiscountTag := &discountEntities.UserDiscountTag{
		UserID:        pgtype.Text{String: createOrderReq.StudentId, Status: pgtype.Present},
		LocationID:    pgtype.Text{String: createOrderReq.LocationId, Status: pgtype.Present},
		DiscountType:  pgtype.Text{String: discountType, Status: pgtype.Present},
		DiscountTagID: pgtype.Text{String: discountTagID, Status: pgtype.Present},
	}

	err = mockdata.UpdateStudentStatus(ctx, s.FatimaDBTrace, createOrderReq.StudentId, createOrderReq.LocationId, "STUDENT_ENROLLMENT_STATUS_ENROLLED", time.Now().AddDate(-1, 0, 0), time.Now().AddDate(1, 0, 0))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	err = mockdata.InsertUserDiscountTag(ctx, s.FatimaDBTrace, userDiscountTag)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) discountServiceSendsDataForDiscountUpdate(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	createOrderReq := stepState.Request.(*pb.CreateOrderRequest)
	discountID := stepState.DiscountID

	discount, err := mockdata.GetDiscountByID(ctx, s.FatimaDBTrace, discountID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	studentProducts, err := mockdata.GetStudentProductsByStudentID(ctx, s.FatimaDBTrace, createOrderReq.StudentId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, studentProduct := range studentProducts {
		var (
			studentProductEndDate   time.Time
			studentProductStartDate time.Time
			effectiveDate           time.Time
		)
		err = multierr.Combine(
			studentProduct.EndDate.AssignTo(&studentProductEndDate),
			studentProduct.StartDate.AssignTo(&studentProductStartDate),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if studentProductStartDate.After(time.Now()) {
			effectiveDate = studentProductStartDate
		} else {
			effectiveDate = time.Now().AddDate(0, 0, 8)
		}

		updateStudentProductDiscount := discountEntities.UpdateProductDiscount{
			StudentID:             createOrderReq.StudentId,
			LocationID:            createOrderReq.LocationId,
			ProductID:             studentProduct.ProductID.String,
			StudentProductID:      studentProduct.StudentProductID.String,
			EffectiveDate:         effectiveDate,
			StudentProductEndDate: studentProductEndDate,
			DiscountID:            discount.DiscountID.String,
			DiscountType:          pb.DiscountType(pb.DiscountType_value[discount.DiscountType.String]),
			DiscountAmountType:    pb.DiscountAmountType(pb.DiscountAmountType_value[discount.DiscountAmountType.String]),
			DiscountAmountValue:   utils.ConvertNumericToFloat32(discount.DiscountAmountValue),
		}

		data, err := json.Marshal(updateStudentProductDiscount)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		msgID, err := s.JSM.PublishAsyncContext(ctx, constants.SubjectUpdateStudentProductCreated, data)
		if err != nil {
			return StepStateToContext(ctx, stepState), nats.HandlePushMsgFail(ctx, fmt.Errorf("SubjectStudentPackageEventNats JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, err))
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) recurringProductDiscountUpdatedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	createOrderReq := stepState.Request.(*pb.CreateOrderRequest)
	_, err := mockdata.GetStudentProductsByStudentID(ctx, s.FatimaDBTrace, createOrderReq.StudentId)
	if err != nil {
		err = fmt.Errorf("student products of student %v not updated", createOrderReq.StudentId)
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), err
}
