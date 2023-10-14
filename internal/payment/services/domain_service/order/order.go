package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/repositories"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderService struct {
	orderRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.Order) error
		UpdateIsReviewFlagByOrderID(
			ctx context.Context,
			db database.QueryExecer,
			orderID string,
			isReviewFlag bool,
			orderVersionNumber int32,
		) (err error)
		GetOrderByIDForUpdate(
			ctx context.Context,
			db database.QueryExecer,
			orderID string,
		) (order entities.Order, err error)
		GetOrderTypeByOrderID(
			ctx context.Context,
			db database.QueryExecer,
			orderID string,
		) (orderType string, err error)
		UpdateOrderStatusByOrderID(
			ctx context.Context,
			db database.QueryExecer,
			orderID string,
			orderStatus string,
		) (err error)
		GetOrderByStudentIDAndLocationIDsPaging(
			ctx context.Context,
			db database.QueryExecer,
			studentID string,
			locationIDs []string,
			from int64,
			limit int64,
		) (
			orders []*entities.Order,
			err error,
		)
		CountOrderByStudentIDAndLocationIDs(
			ctx context.Context,
			db database.QueryExecer,
			studentID string,
			locationIDs []string,
		) (
			total int,
			err error,
		)
		GetAll(ctx context.Context, db database.QueryExecer) ([]*entities.Order, error)
		GetOrderStatsByFilter(ctx context.Context, db database.QueryExecer, filter repositories.OrderListFilter) (orderStats entities.OrderStats, err error)
		GetOrdersByFilter(ctx context.Context, db database.QueryExecer, filter repositories.OrderListFilter) (orders []entities.Order, err error)
		UpdateOrderStatusByOrderIDAndVersion(
			ctx context.Context,
			db database.QueryExecer,
			orderID string,
			orderStatus string,
			orderVersionNumber int32,
		) (err error)
		GetOrderByStudentIDAndLocationIDForResume(
			ctx context.Context,
			db database.QueryExecer,
			studentID string,
			locationID string,
		) (orderID string, err error)
		GetLatestOrderByStudentIDAndLocationIDAndOrderType(
			ctx context.Context,
			db database.QueryExecer,
			studentID, locationID, orderType string,
		) (order entities.Order, err error)
	}
	orderItemRepo interface {
		GetStudentProductIDsForVoidOrderByOrderID(ctx context.Context,
			db database.QueryExecer,
			orderID string,
		) (studentProductIDs []string, err error)
		GetAllByOrderID(ctx context.Context, db database.QueryExecer, orderID string) ([]*entities.OrderItem, error)
		GetOrderItemsByProductIDs(ctx context.Context, db database.QueryExecer, orderIDs []string) ([]entities.OrderItem, error)
	}
	orderActionLogRepo interface {
		Create(ctx context.Context, db database.QueryExecer, orderActionLog *entities.OrderActionLog) error
		GetOrderCreatorsByOrderIDs(ctx context.Context, db database.QueryExecer, orderIDs []string) ([]entities.OrderCreator, error)
	}
	productRepo interface {
		GetByIDs(ctx context.Context, db database.QueryExecer, entitiesIDs []string) ([]entities.Product, error)
	}
	orderLeavingReasonRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.OrderLeavingReason) error
	}
}

func (s *OrderService) CreateOrder(
	ctx context.Context,
	db database.QueryExecer,
	req *pb.CreateOrderRequest,
	studentName string,
	orderStatus pb.OrderStatus,
) (
	order entities.Order,
	err error,
) {
	orderID := idutil.ULIDNow()
	err = multierr.Combine(
		order.OrderID.Set(orderID),
		order.StudentID.Set(req.StudentId),
		order.StudentFullName.Set(studentName),
		order.LocationID.Set(req.LocationId),
		order.OrderComment.Set(req.OrderComment),
		order.OrderStatus.Set(orderStatus),
		order.OrderType.Set(req.OrderType),
		order.IsReviewed.Set(false),
		order.WithdrawalEffectiveDate.Set(nil),
		order.LOAStartDate.Set(nil),
		order.LOAEndDate.Set(nil),
		order.VersionNumber.Set(0),
		setFieldsForStudentStatusUpdate(req, &order),
	)
	switch req.OrderType {
	case pb.OrderType_ORDER_TYPE_WITHDRAWAL, pb.OrderType_ORDER_TYPE_GRADUATE:
		if req.EffectiveDate != nil {
			err = order.WithdrawalEffectiveDate.Set(req.EffectiveDate.AsTime().AddDate(0, 0, 1))
			if err != nil {
				err = status.Errorf(codes.Internal, "error setting effective date for update student status from order request")
				return
			}
		} else {
			err = status.Errorf(codes.FailedPrecondition, "missing effective date for update student status")
			return
		}
	case pb.OrderType_ORDER_TYPE_LOA:
		if req.StartDate == nil || req.EndDate == nil {
			err = status.Errorf(codes.FailedPrecondition, "missing start date or end date for update student status")
			return
		}
		if req.StartDate.AsTime().Before(utils.TimeNow(req.Timezone)) {
			err = status.Errorf(codes.FailedPrecondition, "start_date must not be before current time for update student status")
			return
		}
		if req.EndDate.AsTime().Before(req.StartDate.AsTime()) {
			err = status.Errorf(codes.FailedPrecondition, "end_date must not be before start_date for update student status")
			return
		}
		err = multierr.Combine(
			order.LOAStartDate.Set(req.StartDate.AsTime().AddDate(0, 0, 1)),
			order.LOAEndDate.Set(req.EndDate.AsTime()),
		)
		if err != nil {
			err = status.Errorf(codes.Internal, "error setting end date or start date of LOA from order request")
			return
		}
		if len(req.LeavingReasonIds) == 0 {
			err = status.Errorf(codes.FailedPrecondition, "missing leaving reasons for update student status")
			return
		}
	case pb.OrderType_ORDER_TYPE_RESUME:
		if req.StartDate == nil {
			err = status.Errorf(codes.FailedPrecondition, "missing start date for update student status")
			return
		}
		err = multierr.Combine(
			order.LOAStartDate.Set(req.StartDate.AsTime()),
		)
		if err != nil {
			err = status.Errorf(codes.Internal, "error setting start date of Resume from order request")
			return
		}
	}

	err = s.orderRepo.Create(ctx, db, &order)
	if err != nil {
		err = status.Errorf(codes.Internal, "creating order have error %v", err.Error())
		return
	}

	for _, leavingReasonID := range req.LeavingReasonIds {
		var orderLeavingReason entities.OrderLeavingReason
		err = multierr.Combine(
			orderLeavingReason.OrderID.Set(orderID),
			orderLeavingReason.LeavingReasonID.Set(leavingReasonID),
		)
		if err != nil {
			err = status.Errorf(codes.Internal, "multierr.Combine OrderID.Set LeavingReasonID.Set")
			return
		}

		err = s.orderLeavingReasonRepo.Create(ctx, db, &orderLeavingReason)
		if err != nil {
			err = status.Errorf(codes.Internal, "error when create order leaving reason: %v", err.Error())
			return
		}
	}

	userID := interceptors.UserIDFromContext(ctx)
	actionLog := pgtype.Text{}
	if order.OrderStatus.String == pb.OrderStatus_ORDER_STATUS_SUBMITTED.String() {
		err = actionLog.Set(pb.OrderActionStatus_ORDER_ACTION_SUBMITTED.String())
		if err != nil {
			return
		}
	}
	orderActionLog := entities.OrderActionLog{
		OrderID: order.OrderID,
		Action:  actionLog,
		UserID:  pgtype.Text{Status: pgtype.Present, String: userID},
		Comment: order.OrderComment,
	}
	err = s.orderActionLogRepo.Create(ctx, db, &orderActionLog)
	if err != nil {
		err = status.Errorf(codes.Internal, "creating order action log have error %v", err.Error())
	}
	return
}

func (s *OrderService) CreateCustomOrder(
	ctx context.Context,
	db database.QueryExecer,
	req *pb.CreateCustomBillingRequest,
	studentName string,
	orderStatus pb.OrderStatus,
) (
	order entities.Order,
	err error,
) {

	orderID := idutil.ULIDNow()
	err = multierr.Combine(
		order.OrderID.Set(orderID),
		order.StudentID.Set(req.StudentId),
		order.StudentFullName.Set(studentName),
		order.LocationID.Set(req.LocationId),
		order.OrderComment.Set(req.OrderComment),
		order.OrderStatus.Set(orderStatus),
		order.OrderType.Set(req.OrderType),
		order.IsReviewed.Set(false),
		order.WithdrawalEffectiveDate.Set(nil),
		order.LOAStartDate.Set(nil),
		order.LOAEndDate.Set(nil),
		order.Background.Set(nil),
		order.FutureMeasures.Set(nil),
		order.VersionNumber.Set(0),
	)

	err = s.orderRepo.Create(ctx, db, &order)
	if err != nil {
		err = status.Errorf(codes.Internal, "creating order have error %v", err.Error())
		return
	}
	userID := interceptors.UserIDFromContext(ctx)
	actionLog := pgtype.Text{}
	if order.OrderStatus.String == pb.OrderStatus_ORDER_STATUS_SUBMITTED.String() {
		err = actionLog.Set(pb.OrderActionStatus_ORDER_ACTION_SUBMITTED.String())
		if err != nil {
			return
		}
	}
	orderActionLog := entities.OrderActionLog{
		OrderID: order.OrderID,
		Action:  actionLog,
		UserID:  pgtype.Text{Status: pgtype.Present, String: userID},
		Comment: order.OrderComment,
	}
	err = s.orderActionLogRepo.Create(ctx, db, &orderActionLog)
	if err != nil {
		err = status.Errorf(codes.Internal, "creating order action log have error %v", err.Error())
	}
	return
}

func (s *OrderService) UpdateOrderReview(
	ctx context.Context,
	db database.QueryExecer,
	orderId string,
	isReview bool,
	orderVersionNumber int32,
) (err error) {
	var (
		order entities.Order
	)
	order, err = s.orderRepo.GetOrderByIDForUpdate(ctx, db, orderId)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting order by order id %v have error : %v", orderId, err.Error())
		return err
	}

	err = utils.CheckOutVersion(order.VersionNumber.Int, orderVersionNumber)
	if err != nil {
		return
	}

	err = s.orderRepo.UpdateIsReviewFlagByOrderID(ctx, db, orderId, isReview, orderVersionNumber)
	if err != nil {
		err = status.Errorf(codes.Internal, "updating is review flag by order id %v have error: %v", orderId, err.Error())
		return err
	}

	userID := interceptors.UserIDFromContext(ctx)
	actionLog := pgtype.Text{}
	if isReview {
		err = actionLog.Set(pb.ReviewedFlag_REVIEWED.String())
	} else {
		err = actionLog.Set(pb.ReviewedFlag_NOT_REVIEWED.String())

	}
	if err != nil {
		return err
	}

	orderActionLog := entities.OrderActionLog{
		OrderID: order.OrderID,
		Comment: order.OrderComment,
		Action:  actionLog,
		UserID:  pgtype.Text{Status: pgtype.Present, String: userID},
	}
	err = s.orderActionLogRepo.Create(ctx, db, &orderActionLog)
	if err != nil {
		err = status.Errorf(codes.Internal, "creating order action log for update review flag have error : %v", err.Error())
	}
	return
}

func (s *OrderService) UpdateOrderStatus(
	ctx context.Context,
	db database.QueryExecer,
	orderId string,
	orderStatus pb.OrderStatus,
) (err error) {
	err = s.orderRepo.UpdateOrderStatusByOrderID(ctx, db, orderId, orderStatus.String())
	return
}

func (s *OrderService) VoidOrderReturnOrderAndStudentProductIDs(
	ctx context.Context,
	db database.QueryExecer,
	orderID string,
	orderVersionNumber int32,
) (
	order entities.Order,
	studentProductIDs []string,
	err error,
) {
	order, err = s.orderRepo.GetOrderByIDForUpdate(ctx, db, orderID)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting order have error: %v", err)
		return
	}
	err = utils.CheckOutVersion(order.VersionNumber.Int, orderVersionNumber)
	if err != nil {
		return
	}
	if order.OrderStatus.String == pb.OrderStatus_ORDER_STATUS_INVOICED.String() ||
		order.OrderStatus.String == pb.OrderStatus_ORDER_STATUS_VOIDED.String() {
		err = status.Errorf(codes.Internal, "error when void an invoiced/voided order")
		return
	}
	orderType := order.OrderType.String

	orderItems, err := s.orderItemRepo.GetAllByOrderID(ctx, db, order.OrderID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting all of order items by order id have error: %v", err)
		return
	}
	for i, item := range orderItems {
		studentProductIDs = append(studentProductIDs, orderItems[i].StudentProductID.String)
		switch orderType {
		case pb.OrderType_ORDER_TYPE_UPDATE.String(),
			pb.OrderType_ORDER_TYPE_LOA.String(),
			pb.OrderType_ORDER_TYPE_WITHDRAWAL.String():
			err = validateOrderForVoidOrder(*item)
			if err != nil {
				return
			}
		default:
			continue
		}
	}
	err = multierr.Combine(
		order.OrderStatus.Set(pb.OrderStatus_ORDER_STATUS_VOIDED.String()),
	)
	if err != nil {
		err = status.Errorf(codes.Internal, fmt.Errorf("multierr.Combine : %w", err).Error())
		return
	}
	err = s.orderRepo.UpdateOrderStatusByOrderIDAndVersion(ctx, db, order.OrderID.String, order.OrderStatus.String, orderVersionNumber)
	if err != nil {
		err = status.Errorf(codes.Internal, "updating order status have error : %v", err)
		return
	}

	userID := interceptors.UserIDFromContext(ctx)
	actionLog := pgtype.Text{}
	if err = actionLog.Set(pb.OrderActionStatus_ORDER_ACTION_VOIDED.String()); err != nil {
		return
	}
	orderActionLog := entities.OrderActionLog{
		OrderID: order.OrderID,
		Action:  actionLog,
		UserID:  pgtype.Text{Status: pgtype.Present, String: userID},
		Comment: order.OrderComment,
	}
	err = s.orderActionLogRepo.Create(ctx, db, &orderActionLog)
	if err != nil {
		err = status.Errorf(codes.Internal, "creating order action log have error: %v", err)
	}
	return
}

func validateOrderForVoidOrder(orderItem entities.OrderItem) (err error) {
	now := time.Now()
	if orderItem.EffectiveDate.Status == pgtype.Present && !orderItem.EffectiveDate.Time.After(now) {
		err = status.Errorf(codes.Internal, "cannot void an update order when effective_date have passed")
		return
	}
	return
}

func (s *OrderService) GetOrderTypeByOrderID(ctx context.Context, db database.Ext, orderID string) (orderType string, err error) {
	orderType, err = s.orderRepo.GetOrderTypeByOrderID(ctx, db, orderID)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting order type by order id have error: %v", err)
	}
	return
}

func (s *OrderService) GetOrdersByStudentIDAndLocationIDs(
	ctx context.Context,
	db database.Ext,
	studentID string,
	locationIDs []string,
	from int64,
	limit int64,
) (
	orders []*entities.Order,
	total int,
	err error,
) {
	total, err = s.orderRepo.CountOrderByStudentIDAndLocationIDs(ctx, db, studentID, locationIDs)
	if err != nil {
		err = status.Errorf(codes.Internal, "counting order by student id and location ids with error: %v", err)
		return
	}
	orders = make([]*entities.Order, 0, limit)
	orders, err = s.orderRepo.GetOrderByStudentIDAndLocationIDsPaging(ctx, db, studentID, locationIDs, from, limit)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting order by student id and location ids and pagination with error: %v", err)
		return
	}
	return
}

func (s *OrderService) GetAllOrdersFromDB(ctx context.Context, db database.QueryExecer) ([]*entities.OrderSync, error) {
	orders, err := s.orderRepo.GetAll(ctx, db)
	if err != nil {
		return nil, err
	}

	result := make([]*entities.OrderSync, 0, len(orders))
	for _, order := range orders {
		orderItems, err := s.orderItemRepo.GetAllByOrderID(ctx, db, order.OrderID.String)
		if err != nil {
			return nil, err
		}

		// Get information of products
		mapProductIDs := make(map[string]bool)
		for _, orderItem := range orderItems {
			mapProductIDs[orderItem.ProductID.String] = true
		}
		productIDs := make([]string, 0)
		for productID := range mapProductIDs {
			productIDs = append(productIDs, productID)
		}
		products, err := s.productRepo.GetByIDs(ctx, db, productIDs)
		if err != nil {
			return nil, err
		}
		mapProducts := make(map[string]*entities.ProductSync)
		for _, product := range products {
			mapProducts[product.ProductID.String] = &entities.ProductSync{
				ID:                   product.ProductID,
				Name:                 product.Name,
				ProductType:          product.ProductType,
				TaxID:                product.TaxID,
				AvailableFrom:        product.AvailableFrom,
				AvailableUntil:       product.AvailableUntil,
				CustomBillingPeriod:  product.CustomBillingPeriod,
				BillingScheduleID:    product.BillingScheduleID,
				DisableProRatingFlag: product.DisableProRatingFlag,
				Remarks:              product.Remarks,
				IsArchived:           product.IsArchived,
				UpdatedAt:            product.UpdatedAt,
				CreatedAt:            product.CreatedAt,
				ResourcePath:         product.ResourcePath,
			}
		}

		// Build OrderItems information
		orderProducts := make([]*entities.OrderItemSync, 0, len(orderItems))
		for _, orderItem := range orderItems {
			product := mapProducts[orderItem.ProductID.String]
			if product == nil {
				continue
			}
			orderProducts = append(orderProducts, &entities.OrderItemSync{
				DiscountID:   orderItem.DiscountID,
				StartDate:    orderItem.StartDate,
				CreatedAt:    orderItem.CreatedAt,
				Product:      product,
				ResourcePath: orderItem.ResourcePath,
			})
		}
		result = append(result, &entities.OrderSync{
			ID:                  order.OrderID,
			StudentID:           order.StudentID,
			LocationID:          order.LocationID,
			OrderSequenceNumber: order.OrderSequenceNumber,
			OrderComment:        order.OrderComment,
			OrderStatus:         order.OrderStatus,
			OrderType:           order.OrderType,
			UpdatedAt:           order.UpdatedAt,
			CreatedAt:           order.CreatedAt,
			OrderItems:          orderProducts,
			ResourcePath:        order.ResourcePath,
		})
	}

	return result, nil
}

func (s *OrderService) GetOrderStatByFilter(ctx context.Context, db database.QueryExecer, req *pb.RetrieveListOfOrdersRequest) (orderStat entities.OrderStats, err error) {
	var filter repositories.OrderListFilter

	if len(req.Keyword) > 0 {
		filter.StudentName = req.Keyword
	}

	if len(req.LocationIds) > 0 {
		filter.LocationIDs = req.LocationIds
	}

	if req.Filter != nil {
		if len(req.Filter.OrderTypes) > 0 {
			orderTypes := make([]string, 0, len(req.Filter.OrderTypes))
			for _, orderType := range req.Filter.OrderTypes {
				orderTypes = append(orderTypes, orderType.String())
			}
			filter.OrderTypes = orderTypes
		}

		// Filter for product ids
		if len(req.Filter.ProductIds) > 0 {
			orderItems := make([]entities.OrderItem, 0, len(req.Filter.ProductIds))
			orderItems, err = s.orderItemRepo.GetOrderItemsByProductIDs(ctx, db, req.Filter.ProductIds)
			if err != nil {
				err = status.Errorf(codes.Internal, "Error when getting order items by product ids with error: %v", err)
				return
			}
			if len(orderItems) == 0 {
				return
			}
			mapProductIDsWithOrderID := make(map[string][]string)
			for _, item := range orderItems {
				productIDs, ok := mapProductIDsWithOrderID[item.OrderID.String]
				if !ok {
					productIDs = []string{}
				}
				productIDs = append(productIDs, item.ProductID.String)
				mapProductIDsWithOrderID[item.OrderID.String] = productIDs
			}
			orderIDs := make([]string, 0, len(req.Filter.ProductIds))
			for orderID := range mapProductIDsWithOrderID {
				orderIDs = append(orderIDs, orderID)
			}
			filter.OrderIDs = orderIDs
		}

		if req.Filter.CreatedFrom != nil {
			filter.CreatedFrom = req.Filter.CreatedFrom.AsTime()
		}

		if req.Filter.CreatedTo != nil {
			filter.CreatedTo = req.Filter.CreatedTo.AsTime()
		}
		if req.Filter.OnlyNotReviewed {
			onlyReviewed := false
			filter.IsReviewed = &onlyReviewed
			filter.OrderStatus = pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()
		}

		filter.IsStudentNotEnrolled = req.Filter.OnlyStudentNotEnrolled
	}

	orderStat, err = s.orderRepo.GetOrderStatsByFilter(ctx, db, filter)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when getting order stat by filter with error: %v", err)
		return
	}
	return
}

func (s *OrderService) GetListOfOrdersByFilter(ctx context.Context, db database.QueryExecer, req *pb.RetrieveListOfOrdersRequest, from int64, limit int64) (orders []entities.Order, err error) {
	var filter repositories.OrderListFilter

	if len(req.Keyword) > 0 {
		filter.StudentName = req.Keyword
	}
	if req.Paging != nil {
		filter.Limit = &limit
		filter.Offset = &from
	}
	if req.OrderStatus != pb.OrderStatus_ORDER_STATUS_ALL {
		filter.OrderStatus = req.OrderStatus.String()
	}
	if len(req.LocationIds) > 0 {
		filter.LocationIDs = req.LocationIds
	}
	if req.Filter != nil {
		if len(req.Filter.OrderTypes) > 0 {
			var orderTypes []string
			for _, orderType := range req.Filter.OrderTypes {
				orderTypes = append(orderTypes, orderType.String())
			}
			filter.OrderTypes = orderTypes
		}

		// Filter for product ids
		if len(req.Filter.ProductIds) > 0 {
			orderItems := make([]entities.OrderItem, 0, len(req.Filter.ProductIds))
			orderItems, err = s.orderItemRepo.GetOrderItemsByProductIDs(ctx, db, req.Filter.ProductIds)
			if err != nil {
				err = status.Errorf(codes.Internal, "Error when getting order items by product ids with error: %v", err)
				return
			}
			orderIDs := make([]string, 0, len(req.Filter.ProductIds))
			if len(orderItems) == 0 {
				return
			}
			mapProductIDsWithOrderID := make(map[string][]string)
			for _, item := range orderItems {
				productIDs, ok := mapProductIDsWithOrderID[item.OrderID.String]
				if !ok {
					productIDs = []string{}
				}
				productIDs = append(productIDs, item.ProductID.String)
				mapProductIDsWithOrderID[item.OrderID.String] = productIDs
			}
			for orderID := range mapProductIDsWithOrderID {
				orderIDs = append(orderIDs, orderID)
			}
			filter.OrderIDs = orderIDs
		}

		filterCreatedFrom := req.Filter.CreatedFrom
		if filterCreatedFrom != nil {
			filter.CreatedFrom = filterCreatedFrom.AsTime()
		}

		filterCreatedTo := req.Filter.CreatedTo
		if filterCreatedTo != nil {
			filter.CreatedTo = filterCreatedTo.AsTime()
		}

		if req.Filter.OnlyNotReviewed {
			onlyReviewed := false
			filter.IsReviewed = &onlyReviewed
			filter.OrderStatus = pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()
		}

		filter.IsStudentNotEnrolled = req.Filter.OnlyStudentNotEnrolled
	}

	orders, err = s.orderRepo.GetOrdersByFilter(ctx, db, filter)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when getting orders by filter with error: %v", err)
		return
	}
	return
}

func (s *OrderService) GetOrderByID(ctx context.Context, db database.QueryExecer, orderID string) (order entities.Order, err error) {
	if orderID == "" {
		err = status.Errorf(codes.Internal, "Missing order ID when getting order by ID")
		return
	}
	order, err = s.orderRepo.GetOrderByIDForUpdate(ctx, db, orderID)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when getting order by ID: %v", err)
		return
	}
	return
}

func (s *OrderService) GetOrderCreatorsByOrderIDs(ctx context.Context, db database.QueryExecer, orderIDs []string) (orderCreators []entities.OrderCreator, err error) {
	orderCreators, err = s.orderActionLogRepo.GetOrderCreatorsByOrderIDs(ctx, db, orderIDs)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting order creators by order_ids have error: %v", err)
		return
	}
	return
}

func setFieldsForStudentStatusUpdate(req *pb.CreateOrderRequest, order *entities.Order) (err error) {
	var orderType = req.OrderType

	if orderType == pb.OrderType_ORDER_TYPE_WITHDRAWAL ||
		orderType == pb.OrderType_ORDER_TYPE_GRADUATE ||
		orderType == pb.OrderType_ORDER_TYPE_LOA {
		if req.Background == nil {
			err = status.Errorf(codes.FailedPrecondition, "missing background field")
			return
		}
		if req.FutureMeasures == nil {
			err = status.Errorf(codes.FailedPrecondition, "missing future_measures field")
			return
		}
		err = multierr.Combine(
			order.Background.Set(req.Background.Value),
			order.FutureMeasures.Set(req.FutureMeasures.Value),
		)
		if err != nil {
			return
		}
	} else {
		err = multierr.Combine(
			order.Background.Set(nil),
			order.FutureMeasures.Set(nil),
		)
	}

	return
}

func (s *OrderService) GetStudentProductIDsForResume(ctx context.Context, db database.QueryExecer, studentID string, locationID string) (studentProductIDs []string, err error) {
	orderID, err := s.orderRepo.GetOrderByStudentIDAndLocationIDForResume(ctx, db, studentID, locationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return studentProductIDs, nil
		}
		err = status.Errorf(codes.Internal, "getting orderID by studentID and locationID have error: %v", err)
		return
	}

	orderItems, err := s.orderItemRepo.GetAllByOrderID(ctx, db, orderID)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting order item list by order_id have error: %v", err)
		return
	}

	for _, orderItem := range orderItems {
		studentProductIDs = append(studentProductIDs, orderItem.StudentProductID.String)
	}

	return
}

func (s *OrderService) GetLOAOrderForResume(ctx context.Context, db database.QueryExecer, studentID string, locationID string) (order entities.Order, err error) {
	order, err = s.orderRepo.GetLatestOrderByStudentIDAndLocationIDAndOrderType(ctx, db, studentID, locationID, pb.OrderType_ORDER_TYPE_LOA.String())
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get latest LOA order of student with student_id=%s and location=%s: %v", studentID, locationID, err)
		return
	}

	return
}

func NewOrderService() *OrderService {
	return &OrderService{
		orderRepo:              &repositories.OrderRepo{},
		orderActionLogRepo:     &repositories.OrderActionLogRepo{},
		orderItemRepo:          &repositories.OrderItemRepo{},
		productRepo:            &repositories.ProductRepo{},
		orderLeavingReasonRepo: &repositories.OrderLeavingReasonRepo{},
	}
}
