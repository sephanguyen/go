package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

func (s *suite) prepareDataWithdrawalRecurringPackage(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	defaultOptionPrepareData := mockdata.OptionToPrepareDataForCreateOrderRecurringProduct{
		InsertTax:                        true,
		InsertDiscount:                   true,
		InsertStudent:                    true,
		InsertProductPrice:               false,
		InsertProductLocation:            true,
		InsertLocation:                   false,
		InsertProductGrade:               true,
		InsertFee:                        false,
		InsertMaterial:                   false,
		InsertBillingSchedule:            true,
		InsertBillingScheduleArchived:    false,
		IsTaxExclusive:                   false,
		InsertDiscountNotAvailable:       false,
		InsertProductOutOfTime:           false,
		InsertPackageCourses:             true,
		InsertPackageCourseScheduleBased: false,
		InsertProductDiscount:            true,
		BillingScheduleStartDate:         time.Now(),
		InsertProductSetting:             false,
	}

	var (
		insertOrderReq   pb.CreateOrderRequest
		withdrawOrderReq pb.CreateOrderRequest
		billItems        []*entities.BillItem
		data             mockdata.DataForRecurringProduct
		err              error
	)
	ctx, err = s.subscribeStudentCourseEventSync(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
	insertOrderReq, billItems, data, err = s.createFrequencyBasePackageForWithdrawalDisabledProrating(ctx, defaultOptionPrepareData)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	withdrawOrderReq = s.validWithdrawalRequestFrequencyBasePackageDisabledProrating(&insertOrderReq, billItems, data)

	stepState.Request = &withdrawOrderReq
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareDataGraduationRecurringPackage(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	defaultOptionPrepareData := mockdata.OptionToPrepareDataForCreateOrderRecurringProduct{
		InsertTax:                        true,
		InsertDiscount:                   true,
		InsertStudent:                    true,
		InsertProductPrice:               false,
		InsertProductLocation:            true,
		InsertLocation:                   false,
		InsertProductGrade:               true,
		InsertFee:                        false,
		InsertMaterial:                   false,
		InsertBillingSchedule:            true,
		InsertBillingScheduleArchived:    false,
		IsTaxExclusive:                   false,
		InsertDiscountNotAvailable:       false,
		InsertProductOutOfTime:           false,
		InsertPackageCourses:             true,
		InsertPackageCourseScheduleBased: false,
		InsertProductDiscount:            true,
		BillingScheduleStartDate:         time.Now(),
		InsertProductSetting:             false,
	}

	var (
		insertOrderReq   pb.CreateOrderRequest
		graduateOrderReq pb.CreateOrderRequest
		billItems        []*entities.BillItem
		data             mockdata.DataForRecurringProduct
		err              error
	)
	ctx, err = s.subscribeStudentCourseEventSync(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
	insertOrderReq, billItems, data, err = s.createFrequencyBasePackageForWithdrawal(ctx, defaultOptionPrepareData)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	graduateOrderReq = s.validGraduationRequestFrequencyBasePackage(&insertOrderReq, billItems, data)

	stepState.Request = &graduateOrderReq
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) packageUpsertedToStudentPackageTable(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.CreateOrderRequest)
	studentID := req.StudentId

	// student_package_by_order managed by payment
	studentPackageByOrder, err := s.getStudentPackageByOrderByStudentID(ctx, studentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	studentProduct, err := s.getListStudentProductBaseOnProductID(ctx, studentPackageByOrder.PackageID.String)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if studentPackageByOrder == nil {
		err = fmt.Errorf("fail to save package data to student_packages_by_order table")
		return StepStateToContext(ctx, stepState), err
	}
	if studentPackageByOrder.StartAt != studentProduct[0].StartDate {
		err = fmt.Errorf("fail when studentPackageByOrder.StartAt = %s not equal studentProduct.StartDate = %s, isDelete = %v, student_package_id = %v", studentPackageByOrder.StartAt.Time.String(), studentProduct[0].StartDate.Time.String(), studentPackageByOrder.DeletedAt.Time.String(), studentPackageByOrder.ID.String)
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
