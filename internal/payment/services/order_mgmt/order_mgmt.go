package ordermgmt

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/kafka"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/payment/search"
	createOrderService "github.com/manabie-com/backend/internal/payment/services/order_mgmt/create_order"
	orderDetail "github.com/manabie-com/backend/internal/payment/services/order_mgmt/order_detail"
	studentBilling "github.com/manabie-com/backend/internal/payment/services/order_mgmt/student_billing"
	uniqueProduct "github.com/manabie-com/backend/internal/payment/services/order_mgmt/unique_product"
	fatima_pb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

type OrderMgMt struct {
	pb.UnimplementedOrderServiceServer
	createOrderService                           *createOrderService.CreateOrderService
	createCustomOrder                            *CreateCustomOrder
	updateOrderReviewFlag                        *UpdateOrderReviewFlag
	updateOrderStatus                            *UpdateOrderStatus
	voidOrder                                    *VoidOrder
	orderDetail                                  *orderDetail.OrderDetail
	studentBilling                               *studentBilling.StudentBilling
	uniqueProduct                                *uniqueProduct.UniqueProduct
	orderList                                    *OrderList
	getLocationsForCreatingOrder                 *GetLocationsForCreatingOrder
	getProductList                               *ProductList
	retrieveStudentEnrollmentStatusByLocation    *RetrieveStudentEnrollmentStatusByLocation
	retrieveStudentEnrolledLocations             *RetrieveStudentEnrolledLocations
	getOrgLevelStudentStatus                     *GetOrgLevelStudentStatus
	retrieveRecurringProductsOfStudentInLocation *RetrieveRecurringProductsOfStudentInLocation
}

func NewOrderMgMt(db database.Ext, elasticSearch search.Engine, jsm nats.JetStreamManagement, fatimaClient fatima_pb.SubscriptionModifierServiceClient, kafka kafka.KafkaManagement, config configs.CommonConfig) *OrderMgMt {
	return &OrderMgMt{
		createOrderService:           createOrderService.NewCreateOrderService(db, elasticSearch, jsm, fatimaClient, kafka, config),
		createCustomOrder:            NewCreateCustomOrder(db, elasticSearch, jsm, kafka, config),
		updateOrderReviewFlag:        NewUpdateOrderReviewFlag(db),
		updateOrderStatus:            NewUpdateOrderStatus(db),
		voidOrder:                    NewVoidOrder(db, jsm, fatimaClient, kafka, config),
		orderDetail:                  orderDetail.NewOrderDetail(db),
		orderList:                    NewOrderList(db, elasticSearch),
		studentBilling:               studentBilling.NewStudentBilling(db),
		uniqueProduct:                uniqueProduct.NewUniqueProduct(db),
		getLocationsForCreatingOrder: NewGetLocationsForCreatingOrder(db),
		getProductList:               NewRetrieveListOfProducts(db),
		retrieveStudentEnrollmentStatusByLocation:    NewRetrieveStudentEnrollmentStatusByLocation(db),
		retrieveStudentEnrolledLocations:             NewRetrieveStudentEnrolledLocations(db),
		getOrgLevelStudentStatus:                     NewGetOrgLevelStudentStatus(db),
		retrieveRecurringProductsOfStudentInLocation: NewRetrieveRecurringProductsOfStudentInLocation(db),
	}
}

func (s *OrderMgMt) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (res *pb.CreateOrderResponse, err error) {
	return s.createOrderService.CreateOrder(ctx, req)
}

func (s *OrderMgMt) CreateBulkOrder(ctx context.Context, req *pb.CreateBulkOrderRequest) (res *pb.CreateBulkOrderResponse, err error) {
	return s.createOrderService.CreateBulkOrder(ctx, req)
}

func (s *OrderMgMt) CreateCustomBilling(ctx context.Context, req *pb.CreateCustomBillingRequest) (res *pb.CreateCustomBillingResponse, err error) {
	return s.createCustomOrder.CreateCustomBilling(ctx, req)
}

func (s *OrderMgMt) RetrieveListOfBillItems(ctx context.Context, req *pb.RetrieveListOfBillItemsRequest) (res *pb.RetrieveListOfBillItemsResponse, err error) {
	return s.studentBilling.RetrieveListOfBillItems(ctx, req)
}

func (s *OrderMgMt) RetrieveBillingOfOrderDetails(ctx context.Context, req *pb.RetrieveBillingOfOrderDetailsRequest) (res *pb.RetrieveBillingOfOrderDetailsResponse, err error) {

	return s.orderDetail.RetrieveBillItemsOfOrder(ctx, req)
}

func (s *OrderMgMt) RetrieveListOfUniqueProductIDs(ctx context.Context, req *pb.RetrieveListOfUniqueProductIDsRequest) (res *pb.RetrieveListOfUniqueProductIDsResponse, err error) {
	return s.uniqueProduct.RetrieveListOfUniqueProductIDs(ctx, req)
}

func (s *OrderMgMt) RetrieveListOfUniqueProductIDForBulkOrder(ctx context.Context, req *pb.RetrieveListOfUniqueProductIDForBulkOrderRequest) (res *pb.RetrieveListOfUniqueProductIDForBulkOrderResponse, err error) {
	return s.uniqueProduct.RetrieveListOfUniqueProductIDForBulkOrder(ctx, req)
}

func (s *OrderMgMt) RetrieveListOfOrderDetailProducts(ctx context.Context, req *pb.RetrieveListOfOrderDetailProductsRequest) (res *pb.RetrieveListOfOrderDetailProductsResponse, err error) {
	return s.orderDetail.RetrieveProductsOfOrder(ctx, req)
}

func (s *OrderMgMt) RetrieveListOfOrderItems(ctx context.Context, req *pb.RetrieveListOfOrderItemsRequest) (*pb.RetrieveListOfOrderItemsResponse, error) {
	return s.studentBilling.RetrieveListOfOrderItems(ctx, req)
}

func (s *OrderMgMt) RetrieveListOfOrderProducts(ctx context.Context, req *pb.RetrieveListOfOrderProductsRequest) (res *pb.RetrieveListOfOrderProductsResponse, err error) {
	return s.studentBilling.RetrieveListOfOrderProducts(ctx, req)
}

func (s *OrderMgMt) RetrieveListOfOrders(ctx context.Context, req *pb.RetrieveListOfOrdersRequest) (res *pb.RetrieveListOfOrdersResponse, err error) {
	return s.orderList.RetrieveListOfOrders(ctx, req)
}

func (s *OrderMgMt) UpdateOrderReviewedFlag(ctx context.Context, req *pb.UpdateOrderReviewedFlagRequest) (res *pb.UpdateOrderReviewedFlagResponse, err error) {
	return s.updateOrderReviewFlag.UpdateOrderReviewedFlag(ctx, req)
}

func (s *OrderMgMt) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.UpdateOrderStatusResponse, error) {
	return s.updateOrderStatus.UpdateOrderStatus(ctx, req)
}

func (s *OrderMgMt) VoidOrder(ctx context.Context, req *pb.VoidOrderRequest) (res *pb.VoidOrderResponse, err error) {
	return s.voidOrder.VoidOrder(ctx, req)
}

func (s *OrderMgMt) GetLocationsForCreatingOrder(ctx context.Context, req *pb.GetLocationsForCreatingOrderRequest) (res *pb.GetLocationsForCreatingOrderResponse, err error) {
	return s.getLocationsForCreatingOrder.GetLocationsForCreatingOrder(ctx, req)
}

func (s *OrderMgMt) RetrieveListOfOrderAssociatedProductOfPackages(ctx context.Context, req *pb.RetrieveListOfOrderAssociatedProductOfPackagesRequest) (res *pb.RetrieveListOfOrderAssociatedProductOfPackagesResponse, err error) {
	return s.studentBilling.RetrieveListOfOrderAssociatedProductOfPackages(ctx, req)
}

func (s *OrderMgMt) RetrieveListOfProducts(ctx context.Context, req *pb.RetrieveListOfProductsRequest) (res *pb.RetrieveListOfProductsResponse, err error) {
	return s.getProductList.RetrieveListOfProducts(ctx, req)
}

func (s *OrderMgMt) RetrieveStudentEnrollmentStatusByLocation(ctx context.Context, req *pb.RetrieveStudentEnrollmentStatusByLocationRequest) (res *pb.RetrieveStudentEnrollmentStatusByLocationResponse, err error) {
	return s.retrieveStudentEnrollmentStatusByLocation.RetrieveStudentEnrollmentStatusByLocation(ctx, req)
}

func (s *OrderMgMt) RetrieveStudentEnrolledLocations(ctx context.Context, req *pb.RetrieveStudentEnrolledLocationsRequest) (res *pb.RetrieveStudentEnrolledLocationsResponse, err error) {
	return s.retrieveStudentEnrolledLocations.RetrieveStudentEnrolledLocations(ctx, req)
}

func (s *OrderMgMt) GetOrgLevelStudentStatus(ctx context.Context, req *pb.GetOrgLevelStudentStatusRequest) (res *pb.GetOrgLevelStudentStatusResponse, err error) {
	return s.getOrgLevelStudentStatus.GetOrgLevelStudentStatus(ctx, req)
}

func (s *OrderMgMt) RetrieveRecurringProductsOfStudentInLocation(ctx context.Context, req *pb.RetrieveRecurringProductsOfStudentInLocationRequest) (res *pb.RetrieveRecurringProductsOfStudentInLocationResponse, err error) {
	return s.retrieveRecurringProductsOfStudentInLocation.RetrieveRecurringProductsOfStudentInLocation(ctx, req)
}
