package ordermgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/kafka"
	"github.com/manabie-com/backend/internal/golibs/kafka/payload"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/payment/entities"
	billItemService "github.com/manabie-com/backend/internal/payment/services/domain_service/billing/bill_item"
	upcomingBillItemService "github.com/manabie-com/backend/internal/payment/services/domain_service/billing/upcoming_bill_item"
	orderService "github.com/manabie-com/backend/internal/payment/services/domain_service/order"
	orderItemService "github.com/manabie-com/backend/internal/payment/services/domain_service/order_item"
	productService "github.com/manabie-com/backend/internal/payment/services/domain_service/product"
	studentService "github.com/manabie-com/backend/internal/payment/services/domain_service/student"
	studentPackageService "github.com/manabie-com/backend/internal/payment/services/domain_service/student_package"
	studentProductService "github.com/manabie-com/backend/internal/payment/services/domain_service/student_product"
	subscriptionService "github.com/manabie-com/backend/internal/payment/services/domain_service/subscription"
	"github.com/manabie-com/backend/internal/payment/utils"
	fatima_pb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgx/v4"
)

type IStudentProductServiceForVoidOrder interface {
	VoidStudentProduct(
		ctx context.Context,
		db database.QueryExecer,
		studentProductID string,
		orderType string,
	) (
		studentProduct entities.StudentProduct,
		product entities.Product,
		isCancel bool,
		err error,
	)
	GetStudentProductsByStudentProductIDs(ctx context.Context, db database.Ext, studentProductIDs []string) (studentProducts []entities.StudentProduct, err error)
}

type IOrderServiceForVoidOrder interface {
	VoidOrderReturnOrderAndStudentProductIDs(
		ctx context.Context,
		db database.QueryExecer,
		orderID string,
		orderVersionNumber int32,
	) (
		order entities.Order,
		studentProductIDs []string,
		err error,
	)
	GetOrderByID(
		ctx context.Context,
		db database.QueryExecer,
		orderID string,
	) (order entities.Order, err error)
}

type IBillItemServiceForVoidOrder interface {
	VoidBillItemByOrderID(
		ctx context.Context,
		db database.QueryExecer,
		orderID string,
	) (
		err error,
	)
}

type IStudentPackageServiceForVoidOrder interface {
	VoidStudentPackageAndStudentCourse(
		ctx context.Context,
		db database.QueryExecer,
		voidStudentPackageArgs utils.VoidStudentPackageArgs,
	) (
		studentPackageIDs []*npb.EventStudentPackage,
		err error,
	)
}

type IOrderItemServiceForVoidOrder interface {
	GetOrderItemsByOrderIDs(
		ctx context.Context,
		db database.QueryExecer,
		orderIDs []string,
	) (
		orderItems []entities.OrderItem,
		err error,
	)
}

type IProductServiceForVoidOrder interface {
	GetProductIDsByProductTypeAndOrderID(
		ctx context.Context,
		db database.QueryExecer,
		productType,
		orderID string,
	) (
		productIDs []string,
		err error)
}

type IUpcomingBillItemServiceForVoidOrder interface {
	VoidUpcomingBillItemsByOrder(
		ctx context.Context,
		db database.QueryExecer,
		order entities.Order,
	) (err error)
}

type IStudentServiceForVoidOrder interface {
	GetStudentAndNameByID(ctx context.Context, db database.QueryExecer, studentID string) (
		student entities.Student,
		studentName string,
		err error,
	)
}

type ISubscriptionServiceForVoidOrder interface {
	Publish(ctx context.Context, db database.QueryExecer, message utils.MessageSyncData) (err error)
	ToNotificationMessage(ctx context.Context, tx database.QueryExecer, order entities.Order, student entities.Student, upsertNotificationData utils.UpsertSystemNotificationData) (notificationMessage *payload.UpsertSystemNotification, err error)
}

type VoidOrder struct {
	DB database.Ext

	OrderService            IOrderServiceForVoidOrder
	StudentProductService   IStudentProductServiceForVoidOrder
	BillItemService         IBillItemServiceForVoidOrder
	SubscriptionService     ISubscriptionServiceForVoidOrder
	StudentService          IStudentServiceForVoidOrder
	StudentPackageService   IStudentPackageServiceForVoidOrder
	OrderItemService        IOrderItemServiceForVoidOrder
	ProductService          IProductServiceForVoidOrder
	UpcomingBillItemService IUpcomingBillItemServiceForVoidOrder
}

func (s *VoidOrder) VoidOrder(ctx context.Context, req *pb.VoidOrderRequest) (res *pb.VoidOrderResponse, err error) {
	var (
		orderType                 string
		studentProductIDs         []string
		order                     entities.Order
		studentPackagesEvents     []*npb.EventStudentPackage
		student                   entities.Student
		systemNotificationMessage *payload.UpsertSystemNotification
	)
	err = database.ExecInTxWithContextDeadline(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		order, studentProductIDs, err = s.OrderService.VoidOrderReturnOrderAndStudentProductIDs(ctx, tx, req.OrderId, req.OrderVersionNumber)
		if err != nil {
			return
		}

		student, _, err = s.StudentService.GetStudentAndNameByID(ctx, tx, order.StudentID.String)
		if err != nil {
			return fmt.Errorf("VoidOrder.publishForSubscriptionUpdate: Failed to get student with err: %v", err)
		}

		err = s.BillItemService.VoidBillItemByOrderID(ctx, tx, req.OrderId)
		if err != nil {
			return
		}

		err = s.UpcomingBillItemService.VoidUpcomingBillItemsByOrder(ctx, tx, order)
		if err != nil {
			return
		}

		orderType = order.OrderType.String
		if orderType == pb.OrderType_ORDER_TYPE_CUSTOM_BILLING.String() {
			return
		}

		for _, studentProductID := range studentProductIDs {
			var (
				args                     utils.VoidStudentPackageArgs
				tmpStudentPackagesEvents []*npb.EventStudentPackage
			)
			args.StudentProduct, args.Product, args.IsCancel, err = s.StudentProductService.VoidStudentProduct(ctx, tx, studentProductID, orderType)
			if err != nil {
				return
			}
			if args.Product.ProductType.String != pb.ProductType_PRODUCT_TYPE_PACKAGE.String() {
				continue
			}
			args.Order = order
			tmpStudentPackagesEvents, err = s.StudentPackageService.VoidStudentPackageAndStudentCourse(ctx, tx, args)
			if err != nil {
				return
			}
			studentPackagesEvents = append(studentPackagesEvents, tmpStudentPackagesEvents...)
		}
		systemNotificationMessage, err = s.SubscriptionService.ToNotificationMessage(ctx, tx, order, student, utils.UpsertSystemNotificationData{})
		if err != nil {
			return
		}
		return
	})
	if err != nil {
		return
	}
	err = s.publishForSubscriptionUpdate(ctx, student, order, studentPackagesEvents, systemNotificationMessage)
	if err != nil {
		return
	}
	res = &pb.VoidOrderResponse{Successful: true, OrderId: req.OrderId}
	return
}

func (s *VoidOrder) publishForSubscriptionUpdate(
	ctx context.Context,
	student entities.Student,
	order entities.Order,
	studentPackageEvents []*npb.EventStudentPackage,
	notificationMessage *payload.UpsertSystemNotification,
) (
	err error,
) {
	message := utils.MessageSyncData{
		Order: order,
	}

	orderType := order.OrderType.String
	switch orderType {
	case pb.OrderType_ORDER_TYPE_NEW.String():
		message.OrderType = utils.OrderCreate
	case pb.OrderType_ORDER_TYPE_UPDATE.String():
		message.OrderType = utils.OrderUpdate
	case pb.OrderType_ORDER_TYPE_ENROLLMENT.String():
		message.OrderType = utils.OrderEnrollment
	case pb.OrderType_ORDER_TYPE_WITHDRAWAL.String():
		message.OrderType = utils.OrderWithdraw
	case pb.OrderType_ORDER_TYPE_GRADUATE.String():
		message.OrderType = utils.OrderGraduate
	case pb.OrderType_ORDER_TYPE_LOA.String():
		message.OrderType = utils.OrderLOA
	case pb.OrderType_ORDER_TYPE_RESUME.String():
		message.OrderType = utils.OrderResume
	default:
		return
	}
	message.Student = student
	message.StudentPackages = studentPackageEvents
	message.SystemNotificationMessage = notificationMessage

	err = s.SubscriptionService.Publish(ctx, s.DB, message)
	if err != nil {
		return fmt.Errorf("VoidOrder.publishForSubscriptionUpdate: Error publishing order event log: %v", err)
	}
	return
}

func NewVoidOrder(db database.Ext, jsm nats.JetStreamManagement, _ fatima_pb.SubscriptionModifierServiceClient, kafka kafka.KafkaManagement, config configs.CommonConfig) *VoidOrder {
	return &VoidOrder{
		DB:                      db,
		SubscriptionService:     subscriptionService.NewSubscriptionService(jsm, db, kafka, config),
		StudentProductService:   studentProductService.NewStudentProductService(),
		OrderService:            orderService.NewOrderService(),
		BillItemService:         billItemService.NewBillItemService(),
		StudentService:          studentService.NewStudentService(),
		StudentPackageService:   studentPackageService.NewStudentPackage(),
		OrderItemService:        orderItemService.NewOrderItemService(),
		ProductService:          productService.NewProductService(),
		UpcomingBillItemService: upcomingBillItemService.NewUpcomingBillItemService(),
	}
}
