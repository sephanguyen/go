package ordermgmt

import (
	"context"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/search"
	locationService "github.com/manabie-com/backend/internal/payment/services/domain_service/location"
	orderService "github.com/manabie-com/backend/internal/payment/services/domain_service/order"
	orderItemService "github.com/manabie-com/backend/internal/payment/services/domain_service/order_item"
	productService "github.com/manabie-com/backend/internal/payment/services/domain_service/product"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type IOrderServiceForOrderList interface {
	GetOrderStatByFilter(ctx context.Context, db database.QueryExecer, req *pb.RetrieveListOfOrdersRequest) (orderStat entities.OrderStats, err error)
	GetListOfOrdersByFilter(ctx context.Context, db database.QueryExecer, req *pb.RetrieveListOfOrdersRequest, from int64, limit int64) (orders []entities.Order, err error)
	GetOrderCreatorsByOrderIDs(ctx context.Context, db database.QueryExecer, orderIDs []string) (orderCreators []entities.OrderCreator, err error)
}

type IOrderItemServiceForOrderList interface {
	GetOrderItemsByOrderIDs(ctx context.Context, db database.QueryExecer, orderIDs []string) (orderItems []entities.OrderItem, err error)
}

type IProductServiceForOrderList interface {
	GetProductsByIDs(ctx context.Context, db database.Ext, productIDs []string) (products []entities.Product, err error)
}

type ILocationServiceForOrderList interface {
	GetLocationsByIDs(ctx context.Context, db database.Ext, locationIDs []string) (locations []entities.Location, err error)
}

type OrderList struct {
	DB database.Ext

	OrderService     IOrderServiceForOrderList
	OrderItemService IOrderItemServiceForOrderList
	ProductService   IProductServiceForOrderList
	LocationService  ILocationServiceForOrderList
}

func (s *OrderList) RetrieveListOfOrders(ctx context.Context, req *pb.RetrieveListOfOrdersRequest) (res *pb.RetrieveListOfOrdersResponse, err error) {
	fromIdx := int64(0)
	limit := int64(req.Paging.Limit)
	switch u := req.Paging.Offset.(type) {
	case *cpb.Paging_OffsetInteger:
		fromIdx = u.OffsetInteger
	case *cpb.Paging_OffsetCombined:
		fromIdx = u.OffsetCombined.OffsetInteger
	default:
	}

	var (
		orderStats entities.OrderStats
	)

	orderStats, err = s.OrderService.GetOrderStatByFilter(ctx, s.DB, req)
	if err != nil {
		return nil, err
	}

	orders, err := s.OrderService.GetListOfOrdersByFilter(ctx, s.DB, req, fromIdx, limit)
	if err != nil {
		return nil, err
	}
	result := make([]*pb.RetrieveListOfOrdersResponse_Order, 0, len(orders))

	mapLocationWithOrderID, err := s.getLocationsOfOrdersReturningMapLocationWithOrderID(ctx, orders)
	if err != nil {
		return nil, err
	}

	mapOrderProductsWithOrderID, err := s.buildMapProductsWithOrderID(ctx, orders)
	if err != nil {
		return nil, err
	}

	mapOrderIDWithCreatorInfo, err := s.getMapOrderIDWithCreatorInfo(ctx, orders)
	if err != nil {
		return nil, err
	}

	for _, order := range orders {
		var locationID, locationName, productName, creatorID, creatorName string
		products, ok := mapOrderProductsWithOrderID[order.OrderID.String]
		if ok {
			productNames := make([]string, 0, len(products))
			for _, product := range products {
				productNames = append(productNames, product.Name.String)
			}
			productName = strings.Join(productNames, ", ")
		}

		location, ok := mapLocationWithOrderID[order.OrderID.String]
		if ok {
			locationID = location.LocationID.String
			locationName = location.Name.String
		}

		creator, ok := mapOrderIDWithCreatorInfo[order.OrderID.String]
		if ok {
			creatorID = creator.UserID.String
			creatorName = creator.Name.String
		}

		result = append(result, &pb.RetrieveListOfOrdersResponse_Order{
			OrderSequenceNumber: order.OrderSequenceNumber.Int,
			OrderId:             order.OrderID.String,
			StudentId:           order.StudentID.String,
			StudentName:         order.StudentFullName.String,
			OrderStatus:         pb.OrderStatus(pb.OrderStatus_value[order.OrderStatus.String]),
			OrderType:           pb.OrderType(pb.OrderType_value[order.OrderType.String]),
			ProductDetails:      productName,
			CreateDate:          timestamppb.New(order.CreatedAt.Time),
			IsReviewed:          order.IsReviewed.Bool,
			LocationId:          locationID,
			LocationName:        locationName,
			CreatorInfo: &pb.RetrieveListOfOrdersResponse_Order_CreatorInfo{
				UserId:   creatorID,
				UserName: creatorName,
			},
		})
	}
	var prevPage *cpb.Paging = nil
	var nextPage *cpb.Paging = nil

	if len(result) != 0 {
		totalItems := getTotalItemsByOrderStatus(req.OrderStatus, orderStats)
		prevPage, nextPage = s.buildPaging(req, totalItems, fromIdx, limit)
	}

	return &pb.RetrieveListOfOrdersResponse{
		Items:                    result,
		NextPage:                 nextPage,
		PreviousPage:             prevPage,
		TotalItems:               uint32(orderStats.TotalItems.Int),
		TotalOfSubmitted:         uint32(orderStats.TotalOfSubmitted.Int),
		TotalOfPending:           uint32(orderStats.TotalOfPending.Int),
		TotalOfRejected:          uint32(orderStats.TotalOfRejected.Int),
		TotalOfVoided:            uint32(orderStats.TotalOfVoided.Int),
		TotalOfInvoiced:          uint32(orderStats.TotalOfInvoiced.Int),
		TotalOfOrderNeedToReview: uint32(orderStats.TotalOfNeedToReview.Int),
	}, nil
}

func (s *OrderList) buildPaging(req *pb.RetrieveListOfOrdersRequest, totalItems uint32, fromIdx, limit int64) (prevPage, nextPage *cpb.Paging) {
	paging := req.Paging
	if paging == nil {
		return
	}

	// Build info for next and previous page
	prevOffset := fromIdx - limit
	if prevOffset >= 0 {
		prevPage = &cpb.Paging{
			Limit: uint32(limit),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: prevOffset,
			},
		}
	}
	nextOffset := fromIdx + limit
	if uint32(nextOffset) < totalItems {
		nextPage = &cpb.Paging{
			Limit: uint32(limit),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: nextOffset,
			},
		}
	}
	return
}

func (s *OrderList) buildMapProductsWithOrderID(ctx context.Context, orders []entities.Order) (mapProductsWithOrderID map[string][]entities.Product, err error) {
	orderIDs := make([]string, 0, len(orders))
	for _, order := range orders {
		orderIDs = append(orderIDs, order.OrderID.String)
	}
	orderItems, err := s.OrderItemService.GetOrderItemsByOrderIDs(ctx, s.DB, orderIDs)
	if err != nil {
		return
	}
	totalProductIDs := make([]string, 0, len(orderItems))
	mapProductWithProductID := make(map[string]entities.Product)
	mapProductsWithOrderID = make(map[string][]entities.Product)
	for _, orderItem := range orderItems {
		totalProductIDs = append(totalProductIDs, orderItem.ProductID.String)
	}
	totalProducts, err := s.ProductService.GetProductsByIDs(ctx, s.DB, totalProductIDs)
	if err != nil {
		return
	}
	for _, product := range totalProducts {
		mapProductWithProductID[product.ProductID.String] = product
	}

	for _, orderItem := range orderItems {
		product, ok := mapProductWithProductID[orderItem.ProductID.String]
		if !ok {
			continue
		}
		productsByOrderID, ok := mapProductsWithOrderID[orderItem.OrderID.String]
		if !ok {
			productsByOrderID = []entities.Product{}
		}
		productsByOrderID = append(productsByOrderID, product)
		mapProductsWithOrderID[orderItem.OrderID.String] = productsByOrderID
	}
	return mapProductsWithOrderID, nil
}

func getTotalItemsByOrderStatus(orderStatus pb.OrderStatus, stats entities.OrderStats) uint32 {
	switch orderStatus {
	case pb.OrderStatus_ORDER_STATUS_ALL:
		return uint32(stats.TotalItems.Int)
	case pb.OrderStatus_ORDER_STATUS_SUBMITTED:
		return uint32(stats.TotalOfSubmitted.Int)
	case pb.OrderStatus_ORDER_STATUS_PENDING:
		return uint32(stats.TotalOfPending.Int)
	case pb.OrderStatus_ORDER_STATUS_REJECTED:
		return uint32(stats.TotalOfRejected.Int)
	case pb.OrderStatus_ORDER_STATUS_VOIDED:
		return uint32(stats.TotalOfVoided.Int)
	case pb.OrderStatus_ORDER_STATUS_INVOICED:
		return uint32(stats.TotalOfInvoiced.Int)
	default:
		return 0
	}
}

func (s *OrderList) getLocationsOfOrdersReturningMapLocationWithOrderID(ctx context.Context, orders []entities.Order) (
	mapLocationWithOrderID map[string]entities.Location, err error) {
	locationIDs := make([]string, 0, len(orders))
	for _, order := range orders {
		locationIDs = append(locationIDs, order.LocationID.String)
	}
	locations, err := s.LocationService.GetLocationsByIDs(ctx, s.DB, locationIDs)
	if err != nil {
		return
	}
	mapLocationWithLocationID := make(map[string]entities.Location, len(locations))
	mapLocationWithOrderID = make(map[string]entities.Location, len(orders))
	for _, location := range locations {
		mapLocationWithLocationID[location.LocationID.String] = location
	}
	for _, order := range orders {
		orderLocation, ok := mapLocationWithLocationID[order.LocationID.String]
		if ok {
			mapLocationWithOrderID[order.OrderID.String] = orderLocation
			continue
		}
		mapLocationWithOrderID[order.OrderID.String] = entities.Location{}
	}
	return
}

func (s *OrderList) getMapOrderIDWithCreatorInfo(ctx context.Context, orders []entities.Order) (
	mapOrderIDWithCreatorInfo map[string]entities.OrderCreator,
	err error,
) {
	mapOrderIDWithCreatorInfo = make(map[string]entities.OrderCreator)
	orderIDs := sliceutils.Map(orders, func(order entities.Order) string {
		return order.OrderID.String
	})
	orderCreators, err := s.OrderService.GetOrderCreatorsByOrderIDs(ctx, s.DB, orderIDs)
	if err != nil {
		return
	}
	for _, orderCreator := range orderCreators {
		mapOrderIDWithCreatorInfo[orderCreator.OrderID.String] = orderCreator
	}
	return
}

func NewOrderList(db database.Ext, elasticSearch search.Engine) *OrderList {
	return &OrderList{
		DB:               db,
		OrderService:     orderService.NewOrderService(),
		OrderItemService: orderItemService.NewOrderItemService(),
		ProductService:   productService.NewProductService(),
		LocationService:  locationService.NewLocationService(),
	}
}

//func (s *OrderList) buildStatsOrderStatusCondition(filterCondition op.Condition, orderStatus pb.OrderStatus) op.Condition {
//	condition := op.Equal("order_status", orderStatus.String())
//	if filterCondition != nil {
//		condition = op.And(
//			filterCondition,
//			condition,
//		)
//	}
//	return condition
//}

//func (s *OrderList) buildStatsNeedToReviewCondition(filterCondition op.Condition, isReviewed bool) op.Condition {
//	condition := op.And(
//		op.Equal("is_reviewed", isReviewed),
//		op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
//	)
//	if filterCondition != nil {
//		condition = op.And(
//			filterCondition,
//			condition,
//		)
//	}
//	return condition
//}

//func (s *OrderList) calculateOrderStats(ctx context.Context, filterCondition op.Condition) (entities.OrderStats, error) {
//	// Calculate stats
//	totalItems, err := s.SearchEngine.CountValue(ctx, constant.ElasticOrderTableName, "order_sequence_number", filterCondition)
//	if err != nil {
//		return entities.OrderStats{}, err
//	}
//
//	condition := s.buildStatsOrderStatusCondition(filterCondition, pb.OrderStatus_ORDER_STATUS_SUBMITTED)
//	totalOfSubmitted, err := s.SearchEngine.CountValue(ctx, constant.ElasticOrderTableName, "order_sequence_number", condition)
//	if err != nil {
//		return entities.OrderStats{}, err
//	}
//
//	condition = s.buildStatsOrderStatusCondition(filterCondition, pb.OrderStatus_ORDER_STATUS_PENDING)
//	totalOfPending, err := s.SearchEngine.CountValue(ctx, constant.ElasticOrderTableName, "order_sequence_number", condition)
//	if err != nil {
//		return entities.OrderStats{}, err
//	}
//
//	condition = s.buildStatsOrderStatusCondition(filterCondition, pb.OrderStatus_ORDER_STATUS_REJECTED)
//	totalOfRejected, err := s.SearchEngine.CountValue(ctx, constant.ElasticOrderTableName, "order_sequence_number", condition)
//	if err != nil {
//		return entities.OrderStats{}, err
//	}
//
//	condition = s.buildStatsOrderStatusCondition(filterCondition, pb.OrderStatus_ORDER_STATUS_VOIDED)
//	totalOfVoided, err := s.SearchEngine.CountValue(ctx, constant.ElasticOrderTableName, "order_sequence_number", condition)
//	if err != nil {
//		return entities.OrderStats{}, err
//	}
//
//	condition = s.buildStatsOrderStatusCondition(filterCondition, pb.OrderStatus_ORDER_STATUS_INVOICED)
//	totalOfInvoiced, err := s.SearchEngine.CountValue(ctx, constant.ElasticOrderTableName, "order_sequence_number", condition)
//	if err != nil {
//		return entities.OrderStats{}, err
//	}
//
//	condition = s.buildStatsNeedToReviewCondition(filterCondition, false)
//	totalOfNeedToReview, err := s.SearchEngine.CountValue(ctx, constant.ElasticOrderTableName, "order_sequence_number", condition)
//	if err != nil {
//		return entities.OrderStats{}, err
//	}
//
//	return entities.OrderStats{
//		TotalItems:          totalItems,
//		TotalOfSubmitted:    totalOfSubmitted,
//		TotalOfPending:      totalOfPending,
//		TotalOfRejected:     totalOfRejected,
//		TotalOfVoided:       totalOfVoided,
//		TotalOfInvoiced:     totalOfInvoiced,
//		TotalOfNeedToReview: totalOfNeedToReview,
//	}, nil
//}

//func (s *OrderList) buildFilterConditions(ctx context.Context, req *pb.RetrieveListOfOrdersRequest) (op.Condition, error) {
//	keywordConditions := make([]op.Condition, 0)
//	if len(req.Keyword) > 0 {
//		keywordConditions = append(keywordConditions, op.Equal("student_full_name", req.Keyword))
//		keywordConditions = append(keywordConditions, op.NewRegexpQuery("student_full_name", req.Keyword))
//	}
//
//	// Build searching condition
//	filter := req.Filter
//	orderTypeConditions := make([]op.Condition, 0)
//	orderProductConditions := make([]op.Condition, 0)
//	orderDateConditions := make([]op.Condition, 0)
//
//	// Only not reviewed condition
//	onlyNotReviewed := false
//
//	if filter != nil {
//		orderTypeConditions = make([]op.Condition, 0, len(filter.OrderTypes))
//		for _, orderType := range filter.OrderTypes {
//			orderTypeConditions = append(orderTypeConditions, op.Equal("order_type", orderType.String()))
//		}
//		if len(filter.ProductIds) > 0 {
//			mapOrderProductsWithOrderID, err := s.searchProductsByIDs(ctx, filter.ProductIds)
//			if err != nil {
//				return nil, err
//			}
//			orderProductConditions = make([]op.Condition, 0, len(filter.ProductIds))
//			for orderID := range mapOrderProductsWithOrderID {
//				orderProductConditions = append(orderProductConditions, op.Equal("order_id", orderID))
//			}
//			if len(orderProductConditions) == 0 {
//				return nil, nil
//			}
//		}
//		filterCreatedFrom := req.Filter.CreatedFrom
//		if filterCreatedFrom != nil {
//			orderDateConditions = append(orderDateConditions, op.GreaterThanOrEqual("created_at", filterCreatedFrom.AsTime()))
//		}
//		filterCreatedTo := req.Filter.CreatedTo
//		if filterCreatedTo != nil {
//			orderDateConditions = append(orderDateConditions, op.LessThanOrEqual("created_at", filterCreatedTo.AsTime()))
//		}
//
//		filterOnlyNotReviewed := req.Filter.OnlyNotReviewed
//		if filterOnlyNotReviewed {
//			onlyNotReviewed = true
//		}
//	}
//	if onlyNotReviewed {
//		return op.And(
//			op.Or(orderTypeConditions...),
//			op.Or(orderProductConditions...),
//			op.And(orderDateConditions...),
//			op.Or(keywordConditions...),
//			op.Equal("is_reviewed", false),
//			op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
//		), nil
//	}
//	return op.And(
//		op.Or(orderTypeConditions...),
//		op.Or(orderProductConditions...),
//		op.And(orderDateConditions...),
//		op.Or(keywordConditions...),
//	), nil
//
//}

//func (s *OrderList) buildMapProducts(ctx context.Context, orders []entities.Order) (mapProductsWithOrderID map[string][]entities.Product, err error) {
//	orderIDs := make([]string, 0, len(orders))
//	for _, order := range orders {
//		orderIDs = append(orderIDs, order.OrderID.String)
//	}
//	mapProductsWithOrderID, err = s.searchProductsByOrderIDs(ctx, orderIDs)
//	if err != nil {
//		return nil, err
//	}
//	return mapProductsWithOrderID, nil
//}

//func (s *OrderList) searchProductsByOrderIDs(ctx context.Context, orderIDs []string) (mapProductsWithOrderID map[string][]entities.ElasticProduct, err error) {
//	if len(orderIDs) == 0 {
//		return mapProductsWithOrderID, nil
//	}
//	orderIDConditions := make([]op.Condition, 0, len(orderIDs))
//	for _, orderID := range orderIDs {
//		orderIDConditions = append(orderIDConditions, op.Equal("order_id", orderID))
//	}
//	condition := op.Or(orderIDConditions...)
//	mapOrderProductIDs, err := s.searchProducts(ctx, condition)
//	if err != nil {
//		return nil, err
//	}
//	return s.mapProductsToOrderIDs(ctx, mapOrderProductIDs)
//}

//func (s *OrderList) searchProductsByIDs(ctx context.Context, productIDs []string) (map[string][]*entities.ElasticProduct, error) {
//	if len(productIDs) == 0 {
//		return make(map[string][]*entities.ElasticProduct), nil
//	}
//	productIDConditions := make([]op.Condition, 0, len(productIDs))
//	for _, productID := range productIDs {
//		productIDConditions = append(productIDConditions, op.Equal("product_id", productID))
//	}
//
//	condition := op.Or(productIDConditions...)
//	mapOrderProductIDs, err := s.searchProducts(ctx, condition)
//	if err != nil {
//		return nil, err
//	}
//	return s.mapProductsToOrderIDs(ctx, mapOrderProductIDs)
//}

//func (s *OrderList) searchProducts(ctx context.Context, condition op.Condition) (map[string][]string, error) {
//	items, err := s.SearchEngine.SearchWithoutPaging(ctx, constant.ElasticOrderItemTableName, condition, func(data []byte) (interface{}, error) {
//		orderProduct := new(entities.ElasticOrderItem)
//		err := json.Unmarshal(data, &orderProduct)
//		return orderProduct, err
//	})
//	if err != nil {
//		return nil, err
//	}
//	mapOrderProductIDs := make(map[string][]string)
//	for _, item := range items {
//		orderProduct := item.(*entities.ElasticOrderItem)
//		productIDs, ok := mapOrderProductIDs[orderProduct.OrderID]
//		if !ok {
//			productIDs = []string{}
//		}
//		productIDs = append(productIDs, orderProduct.ProductID)
//		mapOrderProductIDs[orderProduct.OrderID] = productIDs
//	}
//	return mapOrderProductIDs, err
//}

//func (s *OrderList) mapProductsToOrderIDs(ctx context.Context, mapOrderProductIDs map[string][]string) (map[string][]*entities.ElasticProduct, error) {
//	result := make(map[string][]*entities.ElasticProduct)
//	for orderID, productIDs := range mapOrderProductIDs {
//		productIDConditions := make([]op.Condition, 0, len(productIDs))
//		for _, productID := range productIDs {
//			productIDConditions = append(productIDConditions, op.Equal("product_id", productID))
//		}
//		condition := op.Or(productIDConditions...)
//		items, err := s.SearchEngine.SearchWithoutPaging(ctx, constant.ElasticProductTableName, condition, func(data []byte) (interface{}, error) {
//			product := new(entities.ElasticProduct)
//			err := json.Unmarshal(data, &product)
//			return product, err
//		})
//		if err != nil {
//			return nil, err
//		}
//		products := make([]*entities.ElasticProduct, 0, len(items))
//		for _, item := range items {
//			products = append(products, item.(*entities.ElasticProduct))
//		}
//		result[orderID] = products
//	}
//	return result, nil
//}

//func (s *OrderList) RetrieveListOfOrders(ctx context.Context, req *pb.RetrieveListOfOrdersRequest) (res *pb.RetrieveListOfOrdersResponse, err error) {
//	fromIdx := int64(0)
//	limit := int64(req.Paging.Limit)
//	switch u := req.Paging.Offset.(type) {
//	case *cpb.Paging_OffsetInteger:
//		fromIdx = u.OffsetInteger
//	case *cpb.Paging_OffsetCombined:
//		fromIdx = u.OffsetCombined.OffsetInteger
//	default:
//	}
//
//	var (
//		filterCondition op.Condition
//		orderStats      entities.OrderStats
//	)
//
//	filterCondition, err = s.buildFilterConditions(ctx, req)
//	if err != nil {
//		return nil, err
//	}
//
//	orderStats, err = s.calculateOrderStats(ctx, filterCondition)
//	if err != nil {
//		return nil, err
//	}
//
//	var items []interface{}
//	if filterCondition != nil {
//		if req.OrderStatus != pb.OrderStatus_ORDER_STATUS_ALL {
//			filterCondition = op.And(
//				filterCondition,
//				op.Equal("order_status", req.OrderStatus.String()),
//			)
//		}
//		if req.Paging != nil {
//			items, err = s.SearchEngine.Search(ctx, constant.ElasticOrderTableName, filterCondition, func(data []byte) (interface{}, error) {
//				order := new(entities.ElasticOrder)
//				err := json.Unmarshal(data, &order)
//				return order, err
//			}, search.PagingParam{
//				FromIdx:    fromIdx,
//				NumberRows: uint32(limit),
//			}, search.SortParam{
//				ColumnName: "created_at",
//				Ascending:  false,
//			})
//		} else {
//			items, err = s.SearchEngine.SearchWithoutPaging(ctx, constant.ElasticOrderTableName, filterCondition, func(data []byte) (interface{}, error) {
//				order := new(entities.ElasticOrder)
//				err := json.Unmarshal(data, &order)
//				return order, err
//			}, search.SortParam{
//				ColumnName: "created_at",
//				Ascending:  false,
//			})
//		}
//		if err != nil {
//			return nil, err
//		}
//	} else {
//		orderStats = entities.OrderStats{}
//	}
//
//	orders := make([]*entities.ElasticOrder, 0, len(items))
//	for _, item := range items {
//		orders = append(orders, item.(*entities.ElasticOrder))
//	}
//
//	mapOrderProducts, err := s.buildMapProducts(ctx, orders)
//	if err != nil {
//		return nil, err
//	}
//
//	// Convert orders to response
//	result := make([]*pb.RetrieveListOfOrdersResponse_Order, 0)
//	for _, order := range orders {
//		productName := ""
//		products, ok := mapOrderProducts[order.OrderID]
//		if ok {
//			productNames := make([]string, 0, len(products))
//			for _, product := range products {
//				productNames = append(productNames, product.Name)
//			}
//			productName = strings.Join(productNames, ", ")
//		}
//		result = append(result, &pb.RetrieveListOfOrdersResponse_Order{
//			OrderSequenceNumber: order.OrderSequenceNumber,
//			OrderId:             order.OrderID,
//			StudentId:           order.StudentID,
//			StudentName:         order.StudentName,
//			OrderStatus:         pb.OrderStatus(pb.OrderStatus_value[order.OrderStatus]),
//			OrderType:           pb.OrderType(pb.OrderType_value[order.OrderType]),
//			ProductDetails:      productName,
//			CreateDate:          timestamppb.New(order.CreatedAt),
//			IsReviewed:          order.IsReviewed,
//		})
//	}
//
//	var prevPage *cpb.Paging = nil
//	var nextPage *cpb.Paging = nil
//
//	if len(result) != 0 {
//		totalItems := getTotalItemsByOrderStatus(req.OrderStatus, orderStats)
//		prevPage, nextPage = s.buildPaging(req, totalItems, fromIdx, limit)
//	}
//
//	return &pb.RetrieveListOfOrdersResponse{
//		Items:                    result,
//		NextPage:                 nextPage,
//		PreviousPage:             prevPage,
//		TotalItems:               orderStats.TotalItems,
//		TotalOfSubmitted:         orderStats.TotalOfSubmitted,
//		TotalOfPending:           orderStats.TotalOfPending,
//		TotalOfRejected:          orderStats.TotalOfRejected,
//		TotalOfVoided:            orderStats.TotalOfVoided,
//		TotalOfInvoiced:          orderStats.TotalOfInvoiced,
//		TotalOfOrderNeedToReview: orderStats.TotalOfNeedToReview,
//	}, nil
//}

//type OrderList struct {
//	DB           database.Ext
//	SearchEngine search.Engine
//}
//
//func (s *OrderList) RetrieveListOfOrders(ctx context.Context, req *pb.RetrieveListOfOrdersRequest) (res *pb.RetrieveListOfOrdersResponse, err error) {
//	fromIdx := int64(0)
//	limit := int64(req.Paging.Limit)
//	switch u := req.Paging.Offset.(type) {
//	case *cpb.Paging_OffsetInteger:
//		fromIdx = u.OffsetInteger
//	case *cpb.Paging_OffsetCombined:
//		fromIdx = u.OffsetCombined.OffsetInteger
//	default:
//	}
//
//	var (
//		filterCondition op.Condition
//		orderStats      entities.OrderStats
//	)
//
//	filterCondition, err = s.buildFilterConditions(ctx, req)
//	if err != nil {
//		return nil, err
//	}
//
//	orderStats, err = s.calculateOrderStats(ctx, filterCondition)
//	if err != nil {
//		return nil, err
//	}
//
//	if req.OrderStatus != pb.OrderStatus_ORDER_STATUS_ALL && filterCondition != nil {
//		filterCondition = op.And(
//			filterCondition,
//			op.Equal("order_status", req.OrderStatus.String()),
//		)
//	}
//
//	var items []interface{}
//	if filterCondition != nil {
//		if req.Paging != nil {
//			items, err = s.SearchEngine.Search(ctx, constant.ElasticOrderTableName, filterCondition, func(data []byte) (interface{}, error) {
//				order := new(entities.ElasticOrder)
//				err := json.Unmarshal(data, &order)
//				return order, err
//			}, search.PagingParam{
//				FromIdx:    fromIdx,
//				NumberRows: uint32(limit),
//			}, search.SortParam{
//				ColumnName: "created_at",
//				Ascending:  false,
//			})
//		} else {
//			items, err = s.SearchEngine.SearchWithoutPaging(ctx, constant.ElasticOrderTableName, filterCondition, func(data []byte) (interface{}, error) {
//				order := new(entities.ElasticOrder)
//				err := json.Unmarshal(data, &order)
//				return order, err
//			}, search.SortParam{
//				ColumnName: "created_at",
//				Ascending:  false,
//			})
//		}
//		if err != nil {
//			return nil, err
//		}
//	} else {
//		orderStats = entities.OrderStats{}
//	}
//
//	orders := make([]*entities.ElasticOrder, 0, len(items))
//	for _, item := range items {
//		orders = append(orders, item.(*entities.ElasticOrder))
//	}
//
//	mapOrderProducts, err := s.buildMapProducts(ctx, orders)
//	if err != nil {
//		return nil, err
//	}
//
//	// Convert orders to response
//	result := make([]*pb.RetrieveListOfOrdersResponse_Order, 0)
//	for _, order := range orders {
//		productName := ""
//		products, ok := mapOrderProducts[order.OrderID]
//		if ok {
//			productNames := make([]string, 0, len(products))
//			for _, product := range products {
//				productNames = append(productNames, product.Name)
//			}
//			productName = strings.Join(productNames, ", ")
//		}
//		result = append(result, &pb.RetrieveListOfOrdersResponse_Order{
//			OrderSequenceNumber: order.OrderSequenceNumber,
//			OrderId:             order.OrderID,
//			StudentId:           order.StudentID,
//			StudentName:         order.StudentName,
//			OrderStatus:         pb.OrderStatus(pb.OrderStatus_value[order.OrderStatus]),
//			OrderType:           pb.OrderType(pb.OrderType_value[order.OrderType]),
//			ProductDetails:      productName,
//			CreateDate:          timestamppb.New(order.CreatedAt),
//			IsReviewed:          order.IsReviewed,
//		})
//	}
//
//	var prevPage *cpb.Paging = nil
//	var nextPage *cpb.Paging = nil
//
//	if len(result) != 0 {
//		totalItems := getTotalItemsByOrderStatus(req.OrderStatus, orderStats)
//		prevPage, nextPage = s.buildPaging(req, totalItems, fromIdx, limit)
//	}
//
//	return &pb.RetrieveListOfOrdersResponse{
//		Items:                    result,
//		NextPage:                 nextPage,
//		PreviousPage:             prevPage,
//		TotalItems:               orderStats.TotalItems,
//		TotalOfSubmitted:         orderStats.TotalOfSubmitted,
//		TotalOfPending:           orderStats.TotalOfPending,
//		TotalOfRejected:          orderStats.TotalOfRejected,
//		TotalOfVoided:            orderStats.TotalOfVoided,
//		TotalOfInvoiced:          orderStats.TotalOfInvoiced,
//		TotalOfOrderNeedToReview: orderStats.TotalOfNeedToReview,
//	}, nil
//}
//
//func (s *OrderList) buildPaging(req *pb.RetrieveListOfOrdersRequest, totalItems uint32, fromIdx, limit int64) (prevPage, nextPage *cpb.Paging) {
//	paging := req.Paging
//	if paging == nil {
//		return
//	}
//
//	// Build info for next and previous page
//	prevOffset := fromIdx - limit
//	if prevOffset >= 0 {
//		prevPage = &cpb.Paging{
//			Limit: uint32(limit),
//			Offset: &cpb.Paging_OffsetInteger{
//				OffsetInteger: prevOffset,
//			},
//		}
//	}
//	nextOffset := fromIdx + limit
//	if uint32(nextOffset) < totalItems {
//		nextPage = &cpb.Paging{
//			Limit: uint32(limit),
//			Offset: &cpb.Paging_OffsetInteger{
//				OffsetInteger: nextOffset,
//			},
//		}
//	}
//	return
//}
//
//func (s *OrderList) buildStatsOrderStatusCondition(filterCondition op.Condition, orderStatus pb.OrderStatus) op.Condition {
//	condition := op.Equal("order_status", orderStatus.String())
//	if filterCondition != nil {
//		condition = op.And(
//			filterCondition,
//			condition,
//		)
//	}
//	return condition
//}
//
//func (s *OrderList) buildStatsNeedToReviewCondition(filterCondition op.Condition, isReviewed bool) op.Condition {
//	condition := op.And(
//		op.Equal("is_reviewed", isReviewed),
//		op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
//	)
//	if filterCondition != nil {
//		condition = op.And(
//			filterCondition,
//			condition,
//		)
//	}
//	return condition
//}
//
//func (s *OrderList) calculateOrderStats(ctx context.Context, filterCondition op.Condition) (entities.OrderStats, error) {
//	// Calculate stats
//	totalItems, err := s.SearchEngine.CountValue(ctx, constant.ElasticOrderTableName, "order_sequence_number", filterCondition)
//	if err != nil {
//		return entities.OrderStats{}, err
//	}
//
//	condition := s.buildStatsOrderStatusCondition(filterCondition, pb.OrderStatus_ORDER_STATUS_SUBMITTED)
//	totalOfSubmitted, err := s.SearchEngine.CountValue(ctx, constant.ElasticOrderTableName, "order_sequence_number", condition)
//	if err != nil {
//		return entities.OrderStats{}, err
//	}
//
//	condition = s.buildStatsOrderStatusCondition(filterCondition, pb.OrderStatus_ORDER_STATUS_PENDING)
//	totalOfPending, err := s.SearchEngine.CountValue(ctx, constant.ElasticOrderTableName, "order_sequence_number", condition)
//	if err != nil {
//		return entities.OrderStats{}, err
//	}
//
//	condition = s.buildStatsOrderStatusCondition(filterCondition, pb.OrderStatus_ORDER_STATUS_REJECTED)
//	totalOfRejected, err := s.SearchEngine.CountValue(ctx, constant.ElasticOrderTableName, "order_sequence_number", condition)
//	if err != nil {
//		return entities.OrderStats{}, err
//	}
//
//	condition = s.buildStatsOrderStatusCondition(filterCondition, pb.OrderStatus_ORDER_STATUS_VOIDED)
//	totalOfVoided, err := s.SearchEngine.CountValue(ctx, constant.ElasticOrderTableName, "order_sequence_number", condition)
//	if err != nil {
//		return entities.OrderStats{}, err
//	}
//
//	condition = s.buildStatsOrderStatusCondition(filterCondition, pb.OrderStatus_ORDER_STATUS_INVOICED)
//	totalOfInvoiced, err := s.SearchEngine.CountValue(ctx, constant.ElasticOrderTableName, "order_sequence_number", condition)
//	if err != nil {
//		return entities.OrderStats{}, err
//	}
//
//	condition = s.buildStatsNeedToReviewCondition(filterCondition, false)
//	totalOfNeedToReview, err := s.SearchEngine.CountValue(ctx, constant.ElasticOrderTableName, "order_sequence_number", condition)
//	if err != nil {
//		return entities.OrderStats{}, err
//	}
//
//	return entities.OrderStats{
//		TotalItems:          totalItems,
//		TotalOfSubmitted:    totalOfSubmitted,
//		TotalOfPending:      totalOfPending,
//		TotalOfRejected:     totalOfRejected,
//		TotalOfVoided:       totalOfVoided,
//		TotalOfInvoiced:     totalOfInvoiced,
//		TotalOfNeedToReview: totalOfNeedToReview,
//	}, nil
//}
//
//func (s *OrderList) buildFilterConditions(ctx context.Context, req *pb.RetrieveListOfOrdersRequest) (op.Condition, error) {
//	keywordConditions := make([]op.Condition, 0)
//	if len(req.Keyword) > 0 {
//		keywordConditions = append(keywordConditions, op.Equal("student_full_name", req.Keyword))
//		keywordConditions = append(keywordConditions, op.NewRegexpQuery("student_full_name", req.Keyword))
//	}
//
//	// Build searching condition
//	filter := req.Filter
//	orderTypeConditions := make([]op.Condition, 0)
//	orderProductConditions := make([]op.Condition, 0)
//	orderDateConditions := make([]op.Condition, 0)
//
//	// Only not reviewed condition
//	onlyNotReviewed := false
//
//	if filter != nil {
//		orderTypeConditions = make([]op.Condition, 0, len(filter.OrderTypes))
//		for _, orderType := range filter.OrderTypes {
//			orderTypeConditions = append(orderTypeConditions, op.Equal("order_type", orderType.String()))
//		}
//		if len(filter.ProductIds) > 0 {
//			mapOrderProducts, err := s.searchProductsByIDs(ctx, filter.ProductIds)
//			if err != nil {
//				return nil, err
//			}
//			orderProductConditions = make([]op.Condition, 0, len(filter.ProductIds))
//			for orderID := range mapOrderProducts {
//				orderProductConditions = append(orderProductConditions, op.Equal("order_id", orderID))
//			}
//			if len(orderProductConditions) == 0 {
//				return nil, nil
//			}
//		}
//		filterCreatedFrom := req.Filter.CreatedFrom
//		if filterCreatedFrom != nil {
//			orderDateConditions = append(orderDateConditions, op.GreaterThanOrEqual("created_at", filterCreatedFrom.AsTime()))
//		}
//		filterCreatedTo := req.Filter.CreatedTo
//		if filterCreatedTo != nil {
//			orderDateConditions = append(orderDateConditions, op.LessThanOrEqual("created_at", filterCreatedTo.AsTime()))
//		}
//
//		filterOnlyNotReviewed := req.Filter.OnlyNotReviewed
//		if filterOnlyNotReviewed {
//			onlyNotReviewed = true
//		}
//	}
//	// Combine filter conditions
//	resourcePath := golibs.ResourcePathFromCtx(ctx)
//	if len(resourcePath) == 0 {
//		if onlyNotReviewed {
//			return op.And(
//				op.Or(orderTypeConditions...),
//				op.Or(orderProductConditions...),
//				op.And(orderDateConditions...),
//				op.Or(keywordConditions...),
//				op.Equal("is_reviewed", false),
//				op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
//			), nil
//		}
//		return op.And(
//			op.Or(orderTypeConditions...),
//			op.Or(orderProductConditions...),
//			op.And(orderDateConditions...),
//			op.Or(keywordConditions...),
//		), nil
//
//	}
//	if onlyNotReviewed {
//		return op.And(
//			op.Or(orderTypeConditions...),
//			op.Or(orderProductConditions...),
//			op.And(orderDateConditions...),
//			op.Equal("resource_path", resourcePath),
//			op.Or(keywordConditions...),
//			op.Equal("is_reviewed", false),
//			op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
//		), nil
//	}
//	return op.And(
//		op.Or(orderTypeConditions...),
//		op.Or(orderProductConditions...),
//		op.And(orderDateConditions...),
//		op.Equal("resource_path", resourcePath),
//		op.Or(keywordConditions...),
//	), nil
//
//}
//
//func (s *OrderList) buildMapProducts(ctx context.Context, orders []*entities.ElasticOrder) (map[string][]*entities.ElasticProduct, error) {
//	orderIDs := make([]string, 0, len(orders))
//	for _, order := range orders {
//		orderIDs = append(orderIDs, order.OrderID)
//	}
//	mapOrderProducts, err := s.searchProductsByOrderIDs(ctx, orderIDs)
//	if err != nil {
//		return nil, err
//	}
//	return mapOrderProducts, nil
//}
//
//func (s *OrderList) searchProductsByOrderIDs(ctx context.Context, orderIDs []string) (map[string][]*entities.ElasticProduct, error) {
//	if len(orderIDs) == 0 {
//		return make(map[string][]*entities.ElasticProduct), nil
//	}
//	orderIDConditions := make([]op.Condition, 0, len(orderIDs))
//	for _, orderID := range orderIDs {
//		orderIDConditions = append(orderIDConditions, op.Equal("order_id", orderID))
//	}
//	condition := op.Or(orderIDConditions...)
//	mapOrderProductIDs, err := s.searchProducts(ctx, condition)
//	if err != nil {
//		return nil, err
//	}
//	return s.mapProductsToOrderIDs(ctx, mapOrderProductIDs)
//}
//
//func (s *OrderList) searchProductsByIDs(ctx context.Context, productIDs []string) (map[string][]*entities.ElasticProduct, error) {
//	if len(productIDs) == 0 {
//		return make(map[string][]*entities.ElasticProduct), nil
//	}
//	productIDConditions := make([]op.Condition, 0, len(productIDs))
//	for _, productID := range productIDs {
//		productIDConditions = append(productIDConditions, op.Equal("product_id", productID))
//	}
//
//	condition := op.Or(productIDConditions...)
//	mapOrderProductIDs, err := s.searchProducts(ctx, condition)
//	if err != nil {
//		return nil, err
//	}
//	return s.mapProductsToOrderIDs(ctx, mapOrderProductIDs)
//}
//
//func (s *OrderList) searchProducts(ctx context.Context, condition op.Condition) (map[string][]string, error) {
//	items, err := s.SearchEngine.SearchWithoutPaging(ctx, constant.ElasticOrderItemTableName, condition, func(data []byte) (interface{}, error) {
//		orderProduct := new(entities.ElasticOrderItem)
//		err := json.Unmarshal(data, &orderProduct)
//		return orderProduct, err
//	})
//	if err != nil {
//		return nil, err
//	}
//	mapOrderProductIDs := make(map[string][]string)
//	for _, item := range items {
//		orderProduct := item.(*entities.ElasticOrderItem)
//		productIDs, ok := mapOrderProductIDs[orderProduct.OrderID]
//		if !ok {
//			productIDs = []string{}
//		}
//		productIDs = append(productIDs, orderProduct.ProductID)
//		mapOrderProductIDs[orderProduct.OrderID] = productIDs
//	}
//	return mapOrderProductIDs, err
//}
//
//func (s *OrderList) mapProductsToOrderIDs(ctx context.Context, mapOrderProductIDs map[string][]string) (map[string][]*entities.ElasticProduct, error) {
//	result := make(map[string][]*entities.ElasticProduct)
//	for orderID, productIDs := range mapOrderProductIDs {
//		productIDConditions := make([]op.Condition, 0, len(productIDs))
//		for _, productID := range productIDs {
//			productIDConditions = append(productIDConditions, op.Equal("product_id", productID))
//		}
//		condition := op.Or(productIDConditions...)
//		items, err := s.SearchEngine.SearchWithoutPaging(ctx, constant.ElasticProductTableName, condition, func(data []byte) (interface{}, error) {
//			product := new(entities.ElasticProduct)
//			err := json.Unmarshal(data, &product)
//			return product, err
//		})
//		if err != nil {
//			return nil, err
//		}
//		products := make([]*entities.ElasticProduct, 0, len(items))
//		for _, item := range items {
//			products = append(products, item.(*entities.ElasticProduct))
//		}
//		result[orderID] = products
//	}
//	return result, nil
//}
//
//func getTotalItemsByOrderStatus(orderStatus pb.OrderStatus, stats entities.OrderStats) uint32 {
//	switch orderStatus {
//	case pb.OrderStatus_ORDER_STATUS_ALL:
//		return stats.TotalItems
//	case pb.OrderStatus_ORDER_STATUS_SUBMITTED:
//		return stats.TotalOfSubmitted
//	case pb.OrderStatus_ORDER_STATUS_PENDING:
//		return stats.TotalOfPending
//	case pb.OrderStatus_ORDER_STATUS_REJECTED:
//		return stats.TotalOfRejected
//	case pb.OrderStatus_ORDER_STATUS_VOIDED:
//		return stats.TotalOfVoided
//	case pb.OrderStatus_ORDER_STATUS_INVOICED:
//		return stats.TotalOfInvoiced
//	default:
//		return 0
//	}
//}
