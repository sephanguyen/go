package invoicesvc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Set maxRetry to 50 since there are some instance that it exceeded 10 retries
const createInvoiceFromOrderMaxRetry = 50

type validateOrderParam struct {
	orderID                   string
	enableReviewOrderChecking bool
}

func (s *InvoiceModifierService) CreateInvoiceFromOrder(ctx context.Context, req *invoice_pb.CreateInvoiceFromOrderRequest) (*invoice_pb.CreateInvoiceFromOrderResponse, error) {
	if len(req.OrderDetails) == 0 {
		return nil, status.Error(codes.InvalidArgument, "the OrderDetails should not be empty")
	}

	// Map the list of orders by student ID
	studentOrdersMap, err := s.getStudentOrderMap(ctx, req)
	if err != nil {
		return nil, err
	}

	// Get the earliest invoice schedule
	invoiceSchedule, err := s.InvoiceScheduleRepo.GetCurrentEarliestInvoiceSchedule(ctx, s.DB, invoice_pb.InvoiceScheduleStatus_INVOICE_SCHEDULE_SCHEDULED.String())
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Error(codes.Internal, fmt.Sprintf("s.InvoiceScheduleRepo.GetCurrentEarliestInvoiceSchedule err: %v", err))
	}

	response := &invoice_pb.CreateInvoiceFromOrderResponse{
		Successful: true,
	}

	err = utils.DoWithMaxRetry(func(attempt int) (bool, error) {
		err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
			for studentID, orders := range studentOrdersMap {
				// Get the overall bill items of list of order of the student
				billItemList, err := s.getOverallValidBillItemOfOrders(ctx, tx, orders, invoiceSchedule)
				if err != nil {
					return err
				}

				// if there are no valid billing items left to a student, just continue
				if len(billItemList) == 0 {
					continue
				}

				studentBillItemDetails, err := getStudentBillItemDetails(billItemList)
				if err != nil {
					return err
				}

				// Create the invoice of overall bill items
				invoiceDetail := &invoice_pb.GenerateInvoiceDetail{
					StudentId:   studentID,
					BillItemIds: studentBillItemDetails.IDs,
					SubTotal:    studentBillItemDetails.SubTotal,
					Total:       studentBillItemDetails.Total,
					InvoiceType: req.InvoiceType,
				}
				invoiceID, err := s.createInvoice(ctx, tx, invoiceDetail, billItemList)
				if err != nil {
					return status.Error(codes.Internal, err.Error())
				}

				// Create the response per order ID
				for _, order := range orders {
					response.OrderInvoiceData = append(response.OrderInvoiceData, &invoice_pb.OrderInvoiceData{
						InvoiceId: invoiceID,
						OrderId:   order.OrderID.String,
						StudentId: order.StudentID.String,
					})
				}
			}

			return nil
		})

		if err == nil {
			return false, nil
		}

		if err != nil && !strings.Contains(err.Error(), "(SQLSTATE 23505)") {
			return false, err
		}

		log.Printf("Retrying creating invoice from order. Attempt: %d \n", attempt)
		return attempt < createInvoiceFromOrderMaxRetry, fmt.Errorf("cannot generate invoice data, err %v", err)
	}, createInvoiceFromOrderMaxRetry)

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *InvoiceModifierService) getStudentOrderMap(ctx context.Context, req *invoice_pb.CreateInvoiceFromOrderRequest) (map[string][]*entities.Order, error) {
	studentOrdersMap := make(map[string][]*entities.Order)

	enableReviewOrderChecking, err := s.UnleashClient.IsFeatureEnabled(constant.EnableReviewOrderChecking, s.Env)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("s.UnleashClient.IsFeatureEnabled err: %v", err))
	}

	for _, orderDetail := range req.OrderDetails {
		err := validateOrderDetail(orderDetail)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		order, err := s.validateOrder(ctx, &validateOrderParam{orderID: orderDetail.OrderId, enableReviewOrderChecking: enableReviewOrderChecking})
		if err != nil {
			return nil, err
		}

		if orders, exist := studentOrdersMap[order.StudentID.String]; exist {
			studentOrdersMap[order.StudentID.String] = append(orders, order)
		} else {
			studentOrdersMap[order.StudentID.String] = []*entities.Order{order}
		}
	}

	return studentOrdersMap, nil
}

func (s *InvoiceModifierService) getOverallValidBillItemOfOrders(ctx context.Context, db database.Ext, orders []*entities.Order, invoiceSchedule *entities.InvoiceSchedule) ([]*entities.BillItem, error) {
	billItemList := []*entities.BillItem{}
	for _, order := range orders {
		billItems, err := s.validateAndGetOrderBillItem(ctx, db, order.OrderID.String, invoiceSchedule)
		if err != nil {
			return nil, err
		}
		billItemList = append(billItemList, billItems...)
	}

	return billItemList, nil
}

func validateOrderDetail(orderDetail *invoice_pb.OrderDetail) error {
	if orderDetail.OrderId == "" {
		return errors.New("the order ID cannot be empty")
	}

	return nil
}

func (s *InvoiceModifierService) validateAndGetOrderBillItem(ctx context.Context, db database.Ext, orderID string, invoiceSchedule *entities.InvoiceSchedule) ([]*entities.BillItem, error) {
	billItems, err := s.BillItemRepo.FindByOrderID(ctx, db, orderID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("s.BillItemRepo.FindByOrderID err: %v", err))
	}

	if len(billItems) == 0 {
		return nil, status.Error(codes.Internal, fmt.Sprintf("order with ID %s has no associated billing item", orderID))
	}

	validBillItems := []*entities.BillItem{}
	for _, billItem := range billItems {
		// Check if the bill item has correct status
		if billItem.BillStatus.String != payment_pb.BillingStatus_BILLING_STATUS_BILLED.String() &&
			billItem.BillStatus.String != payment_pb.BillingStatus_BILLING_STATUS_PENDING.String() {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("the bill item %v of order with ID %s has invalid status %s",
				billItem.BillItemSequenceNumber.Int, orderID, billItem.BillStatus.String))
		}

		// Filter out the bill items that have billing_date > upcoming invoice_date.
		if invoiceSchedule != nil {
			if billItem.BillDate.Time.After(invoiceSchedule.InvoiceDate.Time) {
				continue
			}
		}

		validBillItems = append(validBillItems, billItem)
	}

	return validBillItems, nil
}

func (s *InvoiceModifierService) validateOrder(ctx context.Context, param *validateOrderParam) (*entities.Order, error) {
	order, err := s.OrderRepo.FindByOrderID(ctx, s.DB, param.orderID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("s.OrderRepo.FindByOrderID err: %v", err))
	}

	if order.OrderStatus.String != payment_pb.OrderStatus_ORDER_STATUS_SUBMITTED.String() {
		return nil, status.Error(codes.InvalidArgument, "order status should be SUBMITTED")
	}

	if !order.IsReviewed.Bool && param.enableReviewOrderChecking {
		return nil, status.Error(codes.InvalidArgument, "order should not contain review required tag")
	}

	return order, nil
}

func getStudentBillItemDetails(billItemList []*entities.BillItem) (*studentBillItemDetails, error) {
	studentBillItemDetails := &studentBillItemDetails{}

	for _, billItem := range billItemList {
		exactTotal, err := getBillItemPrice(billItem)
		if err != nil {
			return nil, err
		}

		studentBillItemDetails.Total += int32(exactTotal)
		studentBillItemDetails.SubTotal += float32(exactTotal)
		studentBillItemDetails.IDs = append(studentBillItemDetails.IDs, billItem.BillItemSequenceNumber.Int)
	}

	return studentBillItemDetails, nil
}

func getBillItemPrice(billItem *entities.BillItem) (float64, error) {
	// Check if the ADJUSTMENT_BILLING type billing item has adjustment price
	if billItem.BillType.String == payment_pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String() && billItem.AdjustmentPrice.Status != pgtype.Present {
		return 0, status.Error(codes.Internal, fmt.Sprintf("The bill item %d with type BILLING_TYPE_ADJUSTMENT_BILLING has no present adjustment price", billItem.BillItemSequenceNumber.Int))
	}

	// Check if the billing item with adjustment price has a ADJUSTMENT_BILLING type
	if billItem.BillType.String != payment_pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String() && billItem.AdjustmentPrice.Status == pgtype.Present {
		return 0, status.Error(codes.Internal, fmt.Sprintf("The bill item %d has present adjustment price but has no BILLING_TYPE_ADJUSTMENT_BILLING type", billItem.BillItemSequenceNumber.Int))
	}

	// Use the adjustment price as amount if bill type is ADJUSTMENT_BILLING and if adjustment price present
	amount := billItem.FinalPrice
	if billItem.AdjustmentPrice.Status == pgtype.Present && billItem.BillType.String == payment_pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String() {
		amount = billItem.AdjustmentPrice
	}

	exactTotal, err := utils.GetFloat64ExactValueAndDecimalPlaces(amount, "2")
	if err != nil {
		return 0, status.Error(codes.Internal, err.Error())
	}

	return exactTotal, nil
}
