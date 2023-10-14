package payment

import (
	"context"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) prepareDataForResumeProducts(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var (
		loaOrderReq    pb.CreateOrderRequest
		resumeOrderReq pb.CreateOrderRequest
		err            error
	)

	loaOrderReq, _, err = s.createLOAForResumeProductDisabledProrating(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	resumeOrderReq = s.validResumeRequestDisabledProrating(&loaOrderReq)
	leavingReasonIDs, err := s.insertLeavingReasonsAndReturnID(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	resumeOrderReq.LeavingReasonIds = []string{leavingReasonIDs[0]}

	stepState.Request = &resumeOrderReq

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) resumeProductsSuccess(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	err := s.checkCreatedOrderDetailsAndActionLogs(ctx, pb.OrderType_ORDER_TYPE_NEW)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	err = s.validateCreatedOrderItemsAndBillItemsForRecurringProducts(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createLOAForResumeProductDisabledProrating(
	ctx context.Context,
) (
	loaOrderReq pb.CreateOrderRequest,
	billItems []*entities.BillItem,
	err error,
) {
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
		InsertProductDiscount:         true,
		BillingScheduleStartDate:      time.Now(),
	}
	var (
		insertOrderReq pb.CreateOrderRequest
		LOAOrderResp   *pb.CreateOrderResponse
	)

	defaultOptionPrepareData.BillingScheduleStartDate = time.Now().AddDate(0, -2, 0)
	insertOrderReq, billItems, err = s.createRecurringMaterialDisabledProrating(ctx, defaultOptionPrepareData)
	if err != nil {
		return
	}
	loaOrderReq = s.validLOARequestDisabledProrating(&insertOrderReq, billItems)

	ctx, err = s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return
	}

	stepState.RequestSentAt = time.Now()

	LOAOrderResp, err = pb.NewOrderServiceClient(s.PaymentConn).
		CreateOrder(contextWithToken(ctx), &loaOrderReq)
	for err != nil && strings.Contains(err.Error(), "(SQLSTATE 23505)") {
		time.Sleep(1000)
		LOAOrderResp, err = pb.NewOrderServiceClient(s.PaymentConn).CreateOrder(contextWithToken(ctx), &loaOrderReq)
	}
	if err != nil {
		stepState.ResponseErr = err
		return
	}

	billItems, err = s.getBillItemsByOrderIDAndProductID(ctx, LOAOrderResp.OrderId, loaOrderReq.OrderItems[0].ProductId)
	if err != nil {
		return
	}

	return
}

func (s *suite) validResumeRequestDisabledProrating(loaRequest *pb.CreateOrderRequest) (req pb.CreateOrderRequest) {
	orderItems := []*pb.OrderItem{}
	billedAtOrderItems := []*pb.BillingItem{}
	upcomingBillingItems := []*pb.BillingItem{}

	orderItems = append(orderItems,
		&pb.OrderItem{
			ProductId:        loaRequest.OrderItems[0].ProductId,
			StartDate:        &timestamppb.Timestamp{Seconds: time.Now().Unix()},
			StudentProductId: loaRequest.OrderItems[0].StudentProductId,
		},
	)
	// prorating disabled
	billedAtOrderItems = append(billedAtOrderItems,
		&pb.BillingItem{
			ProductId:               loaRequest.OrderItems[0].ProductId,
			StudentProductId:        loaRequest.OrderItems[0].StudentProductId,
			BillingSchedulePeriodId: loaRequest.BillingItems[0].BillingSchedulePeriodId,
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         loaRequest.BillingItems[0].TaxItem.TaxId,
				TaxPercentage: loaRequest.BillingItems[0].TaxItem.TaxPercentage,
				TaxCategory:   loaRequest.BillingItems[0].TaxItem.TaxCategory,
				TaxAmount:     getInclusivePercentTax(PriceOrder, loaRequest.BillingItems[0].TaxItem.TaxPercentage),
			},
			FinalPrice: PriceOrder,
		},
	)
	upcomingBillingItems = append(upcomingBillingItems,
		&pb.BillingItem{
			ProductId:               loaRequest.OrderItems[0].ProductId,
			StudentProductId:        loaRequest.OrderItems[0].StudentProductId,
			BillingSchedulePeriodId: loaRequest.UpcomingBillingItems[0].BillingSchedulePeriodId,
			Price:                   PriceOrder,
			Quantity:                &wrapperspb.Int32Value{Value: 1},
			TaxItem: &pb.TaxBillItem{
				TaxId:         loaRequest.UpcomingBillingItems[0].TaxItem.TaxId,
				TaxPercentage: loaRequest.UpcomingBillingItems[0].TaxItem.TaxPercentage,
				TaxCategory:   loaRequest.UpcomingBillingItems[0].TaxItem.TaxCategory,
				TaxAmount:     getInclusivePercentTax(PriceOrder, loaRequest.UpcomingBillingItems[0].TaxItem.TaxPercentage),
			},
			FinalPrice: PriceOrder,
		},
	)

	req.StudentId = loaRequest.StudentId
	req.LocationId = loaRequest.LocationId
	req.OrderComment = "test resume request"
	req.OrderItems = orderItems
	req.BillingItems = billedAtOrderItems
	req.UpcomingBillingItems = upcomingBillingItems
	req.OrderType = pb.OrderType_ORDER_TYPE_RESUME
	req.Background = &wrapperspb.StringValue{Value: "Sample background"}
	req.FutureMeasures = &wrapperspb.StringValue{Value: "Sample future measures"}

	return
}

func (s *suite) selectVersionStudentProduct(ctx context.Context, studentProductID string) (int32, error) {
	var versionNumber int32
	stmt :=
		`
		SELECT version_number
		FROM 
			student_product
		WHERE 
			student_product_id = $1`
	row := s.FatimaDBTrace.QueryRow(ctx, stmt, studentProductID)
	err := row.Scan(&versionNumber)
	if err != nil {
		return versionNumber, err
	}
	return versionNumber, nil
}
