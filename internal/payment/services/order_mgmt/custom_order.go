package ordermgmt

import (
	"context"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/kafka"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/search"
	billingService "github.com/manabie-com/backend/internal/payment/services/domain_service/billing"
	elasticSearchService "github.com/manabie-com/backend/internal/payment/services/domain_service/elastic_search"
	locationService "github.com/manabie-com/backend/internal/payment/services/domain_service/location"
	orderService "github.com/manabie-com/backend/internal/payment/services/domain_service/order"
	orderItemService "github.com/manabie-com/backend/internal/payment/services/domain_service/order_item"
	studentService "github.com/manabie-com/backend/internal/payment/services/domain_service/student"
	subscriptionService "github.com/manabie-com/backend/internal/payment/services/domain_service/subscription"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IOrderServiceForCreateCustomOrder interface {
	CreateCustomOrder(
		ctx context.Context,
		db database.QueryExecer,
		req *pb.CreateCustomBillingRequest,
		studentName string,
		orderStatus pb.OrderStatus,
	) (
		order entities.Order,
		err error,
	)
}

type IStudentServiceForCreateCustomOrder interface {
	GetStudentAndNameByID(ctx context.Context, db database.QueryExecer, studentID string) (
		student entities.Student,
		studentName string,
		err error,
	)
}

type ILocationServiceForCreateCustomOrder interface {
	GetLocationNameByID(ctx context.Context, db database.QueryExecer, locationID string) (locationName string, err error)
}

type IOrderItemServiceForCreateCustomOrder interface {
	CreateMultiCustomOrderItem(
		ctx context.Context,
		db database.QueryExecer,
		req *pb.CreateCustomBillingRequest,
		order entities.Order,
	) (
		orderItem []entities.OrderItem,
		err error,
	)
}

type IBillingServiceForCreateCustomOrder interface {
	CreateBillItemForCustomOrder(
		ctx context.Context,
		db database.QueryExecer,
		req *pb.CreateCustomBillingRequest,
		order entities.Order,
		locationName string,
	) (
		err error,
	)
}

type IElasticSearchServiceForCreateCustomOrder interface {
	InsertOrderData(ctx context.Context, data utils.ElasticSearchData) (err error)
}

type ISubscriptionServiceForCreateCustomOrder interface {
	Publish(ctx context.Context, db database.QueryExecer, message utils.MessageSyncData) (err error)
}

type CreateCustomOrder struct {
	DB database.Ext

	OrderService         IOrderServiceForCreateCustomOrder
	OrderItemService     IOrderItemServiceForCreateCustomOrder
	BillingService       IBillingServiceForCreateCustomOrder
	LocationService      ILocationServiceForCreateCustomOrder
	StudentService       IStudentServiceForCreateCustomOrder
	ElasticSearchService IElasticSearchServiceForCreateCustomOrder
	SubscriptionService  ISubscriptionServiceForCreateCustomOrder
}

func (s *CreateCustomOrder) CreateCustomBilling(ctx context.Context, req *pb.CreateCustomBillingRequest) (res *pb.CreateCustomBillingResponse, err error) {
	var (
		studentName  string
		studentInfo  entities.Student
		locationName string
		order        entities.Order
		orderItems   []entities.OrderItem
		elasticData  utils.ElasticSearchData
		message      utils.MessageSyncData
	)

	err = database.ExecInTxWithContextDeadline(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		studentInfo, studentName, err = s.StudentService.GetStudentAndNameByID(ctx, tx, req.StudentId)
		if err != nil {
			return
		}

		if len(strings.TrimSpace(req.LocationId)) == 0 {
			return status.Errorf(codes.FailedPrecondition, "Missing mandatory data: location")
		}

		locationName, err = s.LocationService.GetLocationNameByID(ctx, tx, req.LocationId)
		if err != nil {
			return
		}

		order, err = s.OrderService.CreateCustomOrder(ctx, tx, req, studentName, pb.OrderStatus_ORDER_STATUS_SUBMITTED)
		if err != nil {
			return
		}

		orderItems, err = s.OrderItemService.CreateMultiCustomOrderItem(ctx, tx, req, order)
		if err != nil {
			return
		}

		err = s.BillingService.CreateBillItemForCustomOrder(ctx, tx, req, order, locationName)
		if err != nil {
			return
		}

		message.Order = order
		message.OrderType = utils.OrderCustom
		message.Student = studentInfo
		elasticData.Order = order
		elasticData.OrderItems = orderItems
		res = &pb.CreateCustomBillingResponse{Successful: true, OrderId: order.OrderID.String}
		return
	})

	if err != nil {
		return
	}
	_ = s.SubscriptionService.Publish(ctx, s.DB, message)
	// err = s.ElasticSearchService.InsertOrderData(ctx, elasticData)

	return
}

func NewCreateCustomOrder(db database.Ext, searchEngine search.Engine, jsm nats.JetStreamManagement, kafka kafka.KafkaManagement, config configs.CommonConfig) *CreateCustomOrder {
	return &CreateCustomOrder{
		DB:                   db,
		OrderService:         orderService.NewOrderService(),
		OrderItemService:     orderItemService.NewOrderItemService(),
		BillingService:       billingService.NewBillingServiceForCustomOrder(),
		LocationService:      locationService.NewLocationService(),
		StudentService:       studentService.NewStudentService(),
		ElasticSearchService: elasticSearchService.NewElasticSearchService(searchEngine),
		SubscriptionService:  subscriptionService.NewSubscriptionService(jsm, db, kafka, config),
	}
}
