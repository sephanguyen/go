package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) prepareDataForOrderWithValidEffectiveDate(ctx context.Context, testcase string) (context.Context, error) {
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
		insertOrderReq                       pb.CreateOrderRequest
		createOrderForUpdateStudentStatusReq pb.CreateOrderRequest
		billItems                            []*entities.BillItem
		err                                  error
	)

	switch testcase {
	case "withdrawal":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		insertOrderReq, billItems, err = s.createRecurringMaterialDisabledProrating(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		createOrderForUpdateStudentStatusReq = s.validWithdrawalRequestDisabledProrating(&insertOrderReq, billItems)
	case "graduation":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		insertOrderReq, billItems, err = s.createRecurringMaterialDisabledProrating(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		createOrderForUpdateStudentStatusReq = s.validGraduateRequestDisabledProrating(&insertOrderReq, billItems)
	case "LOA":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		insertOrderReq, billItems, err = s.createRecurringMaterialDisabledProrating(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		createOrderForUpdateStudentStatusReq = s.validLOARequestDisabledProrating(&insertOrderReq, billItems)
	case "cronjob":
		defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
		insertOrderReq, billItems, err = s.createRecurringMaterialDisabledProrating(ctx, defaultOptionPrepareData)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		createOrderForUpdateStudentStatusReq = s.validLOARequestDisabledProrating(&insertOrderReq, billItems)
	}

	stepState.Request = &createOrderForUpdateStudentStatusReq
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theScheduledJobRunsOnTheEffectiveDateOfOrder(ctx context.Context, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return ctx, err
	}

	resourcePath, err := interceptors.ResourcePathFromContext(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	studentProductLabels := []string{}

	switch testcase {
	case "withdrawal":
		studentProductLabels = append(studentProductLabels, pb.StudentProductLabel_WITHDRAWAL_SCHEDULED.String())
	case "graduation":
		studentProductLabels = append(studentProductLabels, pb.StudentProductLabel_GRADUATION_SCHEDULED.String())
	case "LOA":
		studentProductLabels = append(studentProductLabels, pb.StudentProductLabel_PAUSE_SCHEDULED.String())
	case "cronjob":
		studentProductLabels = append(studentProductLabels, pb.StudentProductLabel_GRADUATION_SCHEDULED.String())
		studentProductLabels = append(studentProductLabels, pb.StudentProductLabel_WITHDRAWAL_SCHEDULED.String())
		studentProductLabels = append(studentProductLabels, pb.StudentProductLabel_PAUSE_SCHEDULED.String())
	}

	req := &pb.UpdateStudentProductStatusRequest{
		OrganizationId:      resourcePath,
		EffectiveDate:       &timestamppb.Timestamp{Seconds: time.Now().Unix()},
		StudentProductLabel: studentProductLabels,
	}

	stepState.Response, stepState.ResponseErr = pb.NewInternalServiceClient(s.PaymentConn).UpdateStudentProductStatus(contextWithToken(ctx), req)
	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentProductStatusChangedToNewStatus(ctx context.Context, testcase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	updatedStudentProductIDs := stepState.Response.(*pb.UpdateStudentProductStatusResponse).StudentProductIds
	studentProducts, err := s.getStudentProductsByIDs(ctx, updatedStudentProductIDs)

	for _, studentProduct := range studentProducts {
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to retrieve student product after update of status")
		}

		if !strings.Contains(testcase, studentProduct.ProductStatus.String) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to update student products status")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
