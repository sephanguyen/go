package service

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/kafka"
	"github.com/manabie-com/backend/internal/golibs/kafka/payload"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/search"
	billingService "github.com/manabie-com/backend/internal/payment/services/domain_service/billing"
	elasticSearchService "github.com/manabie-com/backend/internal/payment/services/domain_service/elastic_search"
	locationService "github.com/manabie-com/backend/internal/payment/services/domain_service/location"
	orderService "github.com/manabie-com/backend/internal/payment/services/domain_service/order"
	orderItemService "github.com/manabie-com/backend/internal/payment/services/domain_service/order_item"
	packageService "github.com/manabie-com/backend/internal/payment/services/domain_service/package"
	productPriceService "github.com/manabie-com/backend/internal/payment/services/domain_service/price"
	productService "github.com/manabie-com/backend/internal/payment/services/domain_service/product"
	studentService "github.com/manabie-com/backend/internal/payment/services/domain_service/student"
	studentPackageService "github.com/manabie-com/backend/internal/payment/services/domain_service/student_package"
	studentProductService "github.com/manabie-com/backend/internal/payment/services/domain_service/student_product"
	subscriptionService "github.com/manabie-com/backend/internal/payment/services/domain_service/subscription"
	userService "github.com/manabie-com/backend/internal/payment/services/domain_service/user"
	"github.com/manabie-com/backend/internal/payment/utils"
	fatima_pb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IStudentProductServiceForCreateOrder interface {
	CreateStudentProduct(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
	) (
		studentProduct entities.StudentProduct,
		err error,
	)
	CreateAssociatedStudentProduct(
		ctx context.Context,
		db database.QueryExecer,
		associatedProducts []*pb.ProductAssociation,
		mapKeyWithOrderItemData map[string]utils.OrderItemData,
	) (
		err error,
	)

	MutationStudentProductForUpdateOrder(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
	) (
		studentProduct entities.StudentProduct,
		rootStudentProduct entities.StudentProduct,
		err error,
	)

	MutationStudentProductForCancelOrder(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
	) (
		studentProduct entities.StudentProduct,
		rootStudentProduct entities.StudentProduct,
		err error,
	)

	MutationStudentProductForWithdrawalOrder(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
	) (
		studentProduct entities.StudentProduct,
		err error,
	)

	MutationStudentProductForGraduateOrder(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
	) (
		studentProduct entities.StudentProduct,
		err error,
	)

	MutationStudentProductForLOAOrder(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
	) (
		studentProduct entities.StudentProduct,
		err error,
	)

	MutationStudentProductForResumeOrder(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
	) (
		studentProduct entities.StudentProduct,
		err error,
	)

	CreateAssociatedStudentProductByAssociatedStudentProductID(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
	) (
		err error,
	)

	DeleteAssociatedStudentProductByAssociatedStudentProductID(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
	) (err error)

	ValidateProductSettingForCreateOrder(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
	) (err error)

	ValidateProductSettingForLOAOrder(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
	) (err error)
	GetStudentProductByStudentProductID(
		ctx context.Context,
		db database.QueryExecer,
		studentProductID string,
	) (
		studentProduct entities.StudentProduct,
		err error,
	)
}

type IOrderServiceForCreateOrder interface {
	CreateOrder(
		ctx context.Context,
		db database.QueryExecer,
		req *pb.CreateOrderRequest,
		studentName string,
		orderStatus pb.OrderStatus,
	) (
		order entities.Order,
		err error,
	)
	GetLOAOrderForResume(ctx context.Context, db database.QueryExecer, studentID string, locationID string,
	) (order entities.Order, err error)
}

type IOrderItemServiceForCreateOrder interface {
	CreateOrderItem(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
	) (
		order entities.OrderItem,
		err error,
	)
}

type IProductServiceForCreateOrder interface {
	VerifiedProductWithStudentInfoReturnProductInfoAndBillingType(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
	) (
		productInfo entities.Product,
		isOneTimeProduct bool,
		isDisableProRatingFlag bool,
		productType pb.ProductType,
		gradeName string,
		productSetting entities.ProductSetting,
		err error,
	)
	VerifiedProductReturnProductInfoAndBillingType(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
	) (
		productInfo entities.Product,
		isOneTimeProduct bool,
		isDisableProRatingFlag bool,
		productType pb.ProductType,
		productSetting entities.ProductSetting,
		err error,
	)
}

type ILocationServiceForCreateOrder interface {
	GetLocationNameByID(ctx context.Context, db database.QueryExecer, locationID string) (locationName string, err error)
}

type IStudentServiceForCreateOrder interface {
	ValidateStudentStatusForOrderType(ctx context.Context, db database.QueryExecer, orderType pb.OrderType, student entities.Student, locationID string, effectiveDate time.Time) (err error)
	IsEnrolledInLocation(ctx context.Context, db database.QueryExecer, orderItemData utils.OrderItemData) (isEnrolledInLocation bool, err error)
	IsEnrolledInOrg(ctx context.Context, db database.QueryExecer, orderItemData utils.OrderItemData) (bool, error)
	GetStudentAndNameByID(ctx context.Context, db database.QueryExecer, studentID string) (
		student entities.Student,
		studentName string,
		err error,
	)
	CheckIsEnrolledInOrgByStudentIDAndTime(
		ctx context.Context,
		db database.QueryExecer,
		studentID string,
		time time.Time,
	) (isEnrolledInOrg bool, err error)
}

type IPackageServiceForCreateOrder interface {
	VerifyPackageDataAndUpsertRelateData(ctx context.Context, db database.QueryExecer, orderItemData utils.OrderItemData) (packageInfo utils.PackageInfo, err error)
}

type ISubscriptionServiceForCreateOrder interface {
	Publish(ctx context.Context, db database.QueryExecer, message utils.MessageSyncData) (err error)
	ToNotificationMessage(ctx context.Context, tx database.QueryExecer, order entities.Order, student entities.Student,
		upsertNotificationData utils.UpsertSystemNotificationData,
	) (notificationMessage *payload.UpsertSystemNotification, err error)
}

type IElasticSearchServiceForCreateOrder interface {
	InsertOrderData(ctx context.Context, data utils.ElasticSearchData) (err error)
}

type IStudentPackageForCreateOrder interface {
	MutationStudentPackageForCreateOrder(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
	) (
		eventMessages []*npb.EventStudentPackage,
		err error,
	)
	MutationStudentPackageForUpdateOrder(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
	) (
		eventMessages []*npb.EventStudentPackage,
		err error,
	)
	MutationStudentPackageForCancelOrder(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
	) (
		eventMessages []*npb.EventStudentPackage,
		err error,
	)
}

type IProductPriceServiceForCreateOrder interface {
	GetProductPricesByProductIDAndPriceType(
		ctx context.Context,
		db database.QueryExecer,
		productID string,
		priceType string,
	) (productPrices []entities.ProductPrice, err error)
}

type IUserServiceForCreateOrder interface {
	GetUserIDsForLoaNotification(ctx context.Context, db database.QueryExecer, locationID string) (userIDs []string, err error)
}

type CreateOrderService struct {
	DB database.Ext

	OrderService          IOrderServiceForCreateOrder
	ProductService        IProductServiceForCreateOrder
	StudentService        IStudentServiceForCreateOrder
	BillingService        utils.IBillingService
	SubscriptionService   ISubscriptionServiceForCreateOrder
	LocationService       ILocationServiceForCreateOrder
	OrderItemService      IOrderItemServiceForCreateOrder
	ElasticSearchService  IElasticSearchServiceForCreateOrder
	StudentProductService IStudentProductServiceForCreateOrder
	PackageService        IPackageServiceForCreateOrder
	StudentPackageService IStudentPackageForCreateOrder
	ProductPriceService   IProductPriceServiceForCreateOrder
	UserService           IUserServiceForCreateOrder
}

func (s *CreateOrderService) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (res *pb.CreateOrderResponse, err error) {
	var (
		studentName             string
		locationName            string
		orderInfo               entities.Order
		studentInfo             entities.Student
		typeOfOrder             *utils.OrderType
		message                 utils.MessageSyncData
		elasticData             utils.ElasticSearchData
		mapKeyWithOrderItemData map[string]utils.OrderItemData
	)
	err = database.ExecInTxWithContextDeadline(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		studentInfo, studentName, err = s.StudentService.GetStudentAndNameByID(ctx, tx, req.StudentId)
		if err != nil {
			return
		}

		locationName, err = s.LocationService.GetLocationNameByID(ctx, tx, req.LocationId)
		if err != nil {
			return
		}

		orderInfo, err = s.OrderService.CreateOrder(ctx, tx, req, studentName, pb.OrderStatus_ORDER_STATUS_SUBMITTED)
		if err != nil {
			return
		}

		if len(req.BillingItems)+len(req.UpcomingBillingItems) == 0 {
			if req.OrderType == pb.OrderType_ORDER_TYPE_WITHDRAWAL ||
				req.OrderType == pb.OrderType_ORDER_TYPE_GRADUATE ||
				req.OrderType == pb.OrderType_ORDER_TYPE_LOA ||
				req.OrderType == pb.OrderType_ORDER_TYPE_RESUME {

				err = s.StudentService.ValidateStudentStatusForOrderType(ctx, tx, req.OrderType, studentInfo, req.LocationId, req.EffectiveDate.AsTime())
				if err != nil {
					return
				}

				message = utils.MessageSyncData{
					Order:   orderInfo,
					Student: studentInfo,
				}
				elasticData.Order = orderInfo
				res = &pb.CreateOrderResponse{
					OrderId:    orderInfo.OrderID.String,
					Successful: true,
				}
				upsertSystemNotificationData := utils.UpsertSystemNotificationData{
					StudentDetailPath: req.StudentDetailPath.GetValue(),
					LocationName:      locationName,
					EndDate:           req.EndDate.AsTime(),
					StartDate:         req.StartDate.AsTime(),
					Timezone:          req.Timezone,
				}
				message.SystemNotificationMessage, err = s.SubscriptionService.ToNotificationMessage(ctx, tx, orderInfo, studentInfo, upsertSystemNotificationData)
				if err != nil {
					return
				}
				return
			}
			err = utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.UpdateLikeOrdersMissingBillItem,
				nil,
			)
			return
		}

		mapKeyWithOrderItemData, typeOfOrder, err = mapOrderItemData(req, orderInfo, studentInfo, locationName, studentName)
		if err != nil {
			return
		}
		message, elasticData, err = s.CreateOrderItems(ctx, tx, typeOfOrder, mapKeyWithOrderItemData)
		if err != nil {
			return
		}
		message.Order = orderInfo
		message.OrderType = *typeOfOrder
		message.Student = studentInfo
		message.SystemNotificationMessage, err = s.SubscriptionService.ToNotificationMessage(ctx, tx, orderInfo, studentInfo, utils.UpsertSystemNotificationData{
			StudentDetailPath: req.StudentDetailPath.GetValue(),
			LocationName:      locationName,
			EndDate:           req.EndDate.AsTime(),
			StartDate:         req.StartDate.AsTime(),
			Timezone:          req.Timezone,
		})
		if err != nil {
			return
		}
		elasticData.Order = orderInfo
		res = &pb.CreateOrderResponse{
			OrderId:    orderInfo.OrderID.String,
			Successful: true,
		}
		return
	})
	if err != nil {
		return
	}
	_ = s.SubscriptionService.Publish(ctx, s.DB, message)
	return
}

func (s *CreateOrderService) CreateOrderItems(ctx context.Context, tx database.QueryExecer, typeOfOrder *utils.OrderType, mapKeyWithOrderItemData map[string]utils.OrderItemData) (
	message utils.MessageSyncData,
	elasticData utils.ElasticSearchData,
	err error,
) {
	switch *typeOfOrder {
	case utils.OrderCreate:
		return s.OrderItemCreate(ctx, tx, mapKeyWithOrderItemData)
	case utils.OrderUpdate:
		return s.OrderItemUpdate(ctx, tx, mapKeyWithOrderItemData)
	case utils.OrderCancel:
		return s.OrderItemCancel(ctx, tx, mapKeyWithOrderItemData)
	case utils.OrderWithdraw:
		return s.OrderItemWithdrawal(ctx, tx, mapKeyWithOrderItemData)
	case utils.OrderGraduate:
		return s.OrderItemGraduate(ctx, tx, mapKeyWithOrderItemData)
	case utils.OrderLOA:
		return s.OrderItemLoa(ctx, tx, mapKeyWithOrderItemData)
	case utils.OrderEnrollment:
		return s.OrderItemEnrollment(ctx, tx, mapKeyWithOrderItemData)
	case utils.OrderResume:
		return s.OrderItemResume(ctx, tx, mapKeyWithOrderItemData)
	}
	return
}

func (s *CreateOrderService) CreateBulkOrder(ctx context.Context, req *pb.CreateBulkOrderRequest) (res *pb.CreateBulkOrderResponse, err error) {
	newOrderResponses := make([]*pb.CreateBulkOrderResponse_CreateNewOrderResponse, 0, len(req.NewOrderRequests))
	messages := make([]utils.MessageSyncData, 0, len(req.NewOrderRequests))
	err = database.ExecInTxWithContextDeadline(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		for index, v := range req.NewOrderRequests {
			var (
				studentName             string
				locationName            string
				orderInfo               entities.Order
				studentInfo             entities.Student
				typeOfOrder             *utils.OrderType
				message                 utils.MessageSyncData
				mapKeyWithOrderItemData map[string]utils.OrderItemData
			)
			if v.OrderType != pb.OrderType_ORDER_TYPE_NEW {
				err = status.Errorf(codes.FailedPrecondition, "we don't support this order type in bulk order")
				return
			}
			orderRequest := req.NewOrderRequests[index]
			tmpReq := &pb.CreateOrderRequest{
				StudentId:            orderRequest.StudentId,
				LocationId:           orderRequest.LocationId,
				OrderComment:         orderRequest.OrderComment,
				OrderType:            orderRequest.OrderType,
				OrderItems:           orderRequest.OrderItems,
				UpcomingBillingItems: orderRequest.UpcomingBillingItems,
				BillingItems:         orderRequest.BillingItems,
				Timezone:             orderRequest.Timezone,
			}

			studentInfo, studentName, err = s.StudentService.GetStudentAndNameByID(ctx, tx, tmpReq.StudentId)
			if err != nil {
				return
			}

			locationName, err = s.LocationService.GetLocationNameByID(ctx, tx, tmpReq.LocationId)
			if err != nil {
				return
			}

			orderInfo, err = s.OrderService.CreateOrder(ctx, tx, tmpReq, studentName, pb.OrderStatus_ORDER_STATUS_SUBMITTED)
			if err != nil {
				return
			}

			if len(tmpReq.BillingItems)+len(tmpReq.UpcomingBillingItems) == 0 {
				err = utils.StatusErrWithDetail(
					codes.FailedPrecondition,
					constant.UpdateLikeOrdersMissingBillItem,
					nil,
				)
				return
			}

			mapKeyWithOrderItemData, typeOfOrder, err = mapOrderItemData(tmpReq, orderInfo, studentInfo, locationName, studentName)
			if err != nil {
				return
			}
			message, _, err = s.OrderItemCreate(ctx, tx, mapKeyWithOrderItemData)
			if err != nil {
				return
			}
			message.Order = orderInfo
			message.OrderType = *typeOfOrder
			message.Student = studentInfo
			if err != nil {
				return
			}
			newOrderResponses = append(newOrderResponses, &pb.CreateBulkOrderResponse_CreateNewOrderResponse{
				Successful: true,
				OrderId:    orderInfo.OrderID.String,
			})
			messages = append(messages, message)
		}
		return
	})
	if err != nil {
		return
	}
	for i := range messages {
		err = s.SubscriptionService.Publish(ctx, s.DB, messages[i])
		if err != nil {
			return
		}
	}
	res = &pb.CreateBulkOrderResponse{
		NewOrderResponses: newOrderResponses,
	}
	return
}

func NewCreateOrderService(db database.Ext, searchEngine search.Engine, jsm nats.JetStreamManagement, fatimaClient fatima_pb.SubscriptionModifierServiceClient, kafka kafka.KafkaManagement, config configs.CommonConfig) *CreateOrderService {
	return &CreateOrderService{
		DB:                    db,
		SubscriptionService:   subscriptionService.NewSubscriptionService(jsm, db, kafka, config),
		ElasticSearchService:  elasticSearchService.NewElasticSearchService(searchEngine),
		OrderService:          orderService.NewOrderService(),
		OrderItemService:      orderItemService.NewOrderItemService(),
		StudentService:        studentService.NewStudentService(),
		LocationService:       locationService.NewLocationService(),
		ProductService:        productService.NewProductService(),
		BillingService:        billingService.NewBillingService(),
		StudentProductService: studentProductService.NewStudentProductService(),
		PackageService:        packageService.NewPackageService(),
		StudentPackageService: studentPackageService.NewStudentPackage(),
		ProductPriceService:   productPriceService.NewPriceService(),
		UserService:           userService.NewUserService(),
	}
}
