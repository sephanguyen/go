package service

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/kafka"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/payment/entities"
	billingItemService "github.com/manabie-com/backend/internal/payment/services/domain_service/billing/bill_item"
	billingScheduleService "github.com/manabie-com/backend/internal/payment/services/domain_service/billing/billing_schedule"
	upcomingBillItemService "github.com/manabie-com/backend/internal/payment/services/domain_service/billing/upcoming_bill_item"
	discount "github.com/manabie-com/backend/internal/payment/services/domain_service/discount"
	orderService "github.com/manabie-com/backend/internal/payment/services/domain_service/order"
	packageService "github.com/manabie-com/backend/internal/payment/services/domain_service/package"
	price "github.com/manabie-com/backend/internal/payment/services/domain_service/price"
	productService "github.com/manabie-com/backend/internal/payment/services/domain_service/product"
	studentService "github.com/manabie-com/backend/internal/payment/services/domain_service/student"
	studentPackageService "github.com/manabie-com/backend/internal/payment/services/domain_service/student_package"
	studentPackageOrderService "github.com/manabie-com/backend/internal/payment/services/domain_service/student_package/student_package_order"
	studentProductService "github.com/manabie-com/backend/internal/payment/services/domain_service/student_product"
	subscriptionService "github.com/manabie-com/backend/internal/payment/services/domain_service/subscription"
	taxService "github.com/manabie-com/backend/internal/payment/services/domain_service/tax"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

type IOrderServiceForInternalService interface {
	UpdateOrderStatus(
		ctx context.Context,
		db database.QueryExecer,
		orderId string,
		orderStatus pb.OrderStatus,
	) (err error)
}

type IBillItemServiceForInternalService interface {
	UpdateBillItemStatusAndReturnOrderID(
		ctx context.Context,
		db database.QueryExecer,
		billItemSequenceNumber int32,
		billItemStatus string,
	) (orderID string, err error)
	CreateUpcomingBillItems(
		ctx context.Context,
		db database.QueryExecer,
		billItem *entities.BillItem,
	) (err error)
	GetRecurringBillItemsByOrderIDAndProductID(
		ctx context.Context,
		db database.QueryExecer,
		orderID string,
		productID string,
	) (billItems []entities.BillItem, err error)
}

type IUpcomingBillItemServiceForInternalService interface {
	GetUpcomingBillItemsForGenerate(
		ctx context.Context,
		db database.QueryExecer,
	) (billItems []entities.UpcomingBillItem, err error)
	CreateUpcomingBillItem(
		ctx context.Context,
		db database.QueryExecer,
		billItem entities.BillItem,
	) (err error)
	AddExecuteNoteForCurrentUpcomingBillItem(
		ctx context.Context,
		db database.QueryExecer,
		upcomingBillItem entities.UpcomingBillItem,
		err error,
	) (importErr error)
	UpdateCurrentUpcomingBillItemStatus(
		ctx context.Context,
		db database.QueryExecer,
		upcomingBillItem entities.UpcomingBillItem,
	) (err error)
	SetLastUpcomingBillItem(
		ctx context.Context,
		db database.QueryExecer,
		upcomingBillItem entities.UpcomingBillItem,
	) (err error)
	GetUpcomingBillItemByOrderIDProductIDBillingSchedulePeriodID(
		ctx context.Context,
		db database.QueryExecer,
		orderID string,
		productID string,
		billingSchedulePeriodID string,
	) ([]entities.UpcomingBillItem, error)
}

type ITaxServiceForInternalService interface {
	GetTaxByID(
		ctx context.Context,
		db database.QueryExecer,
		taxID string,
	) (taxEntity entities.Tax, err error)
}

type IDiscountServiceForInternalService interface {
	VerifyDiscountForGenerateUpcomingBillItem(ctx context.Context, db database.QueryExecer, billItems []entities.BillItem) (discount entities.Discount, err error)
}

type IBillingScheduleForInternalService interface {
	GetBillingSchedulePeriodByID(
		ctx context.Context,
		db database.QueryExecer,
		billingSchedulePeriodID string,
	) (
		billingSchedulePeriod entities.BillingSchedulePeriod,
		err error,
	)
	GetLatestBillingSchedulePeriod(
		ctx context.Context,
		db database.QueryExecer,
		billingScheduleID string,
	) (
		latestBillingSchedulePeriod entities.BillingSchedulePeriod,
		err error,
	)

	GetAllBillingPeriodsByBillingScheduleID(
		ctx context.Context,
		db database.QueryExecer,
		billingScheduleID string,
	) (
		billingSchedulePeriods []entities.BillingSchedulePeriod,
		err error,
	)

	GetNextBillingSchedulePeriod(
		ctx context.Context,
		db database.QueryExecer,
		billingScheduleID string,
		endTime time.Time,
	) (nextPeriod entities.BillingSchedulePeriod, err error)
}

type IStudentProductServiceForInternalService interface {
	GetStudentProductsByStudentProductLabel(
		ctx context.Context,
		db database.QueryExecer,
		studentProductLabels []string,
	) (studentProducts []*entities.StudentProduct, err error)
	CancelStudentProduct(
		ctx context.Context,
		db database.QueryExecer,
		studentProductID string,
	) (err error)
	PauseStudentProduct(
		ctx context.Context,
		db database.QueryExecer,
		studentProduct entities.StudentProduct,
	) (err error)
	GetStudentProductByStudentProductIDForUpdate(
		ctx context.Context,
		db database.QueryExecer,
		studentProductID string,
	) (
		studentProduct entities.StudentProduct,
		err error,
	)
}

type IProductServiceForInternalService interface {
	GetProductByID(ctx context.Context, db database.QueryExecer, productID string) (product entities.Product, err error)
}

type IPriceServiceForInternalService interface {
	CalculatorBillItemPrice(
		ctx context.Context,
		db database.QueryExecer,
		billItem *entities.BillItem,
		upcomingBillItem entities.UpcomingBillItem,
		tax entities.Tax,
		discount entities.Discount,
		priceType string,
		billItemDescription *entities.BillingItemDescription,
		billingSchedulePeriod entities.BillingSchedulePeriod,
	) (err error)
	GetProductPricesByProductIDAndPriceType(
		ctx context.Context,
		db database.QueryExecer,
		productID string,
		priceType string,
	) (productPrices []entities.ProductPrice, err error)
}

type IStudentPackageForInternalService interface {
	GetStudentPackagesForCronJob(ctx context.Context, db database.QueryExecer) (
		studentPackages []entities.StudentPackages,
		err error,
	)
	UpsertStudentPackageDataForCronjob(ctx context.Context,
		db database.QueryExecer,
		studentPackage entities.StudentPackages,
	) (eventMessage *npb.EventStudentPackage,
		currentStudentPackageOrder *entities.StudentPackageOrder,
		err error)
}

type IPackageForInternalService interface {
	GetQuantityTypeByID(ctx context.Context, db database.Ext, packageID string) (quantityType pb.QuantityType, err error)
}

type IStudentForInternalService interface {
	CheckIsEnrolledInOrgByStudentIDAndTime(
		ctx context.Context,
		db database.QueryExecer,
		studentID string,
		time time.Time,
	) (isEnrolledInOrg bool, err error)
}

type ISubscriptionServiceForInternalService interface {
	PublishStudentPackageForCreateOrder(ctx context.Context, eventMessages []*npb.EventStudentPackage) (err error)
}

type IStudentPackageOrderForInternalService interface {
	UpdateExecuteError(ctx context.Context, db database.QueryExecer, studentPackageOrder entities.StudentPackageOrder) (err error)
}

type InternalService struct {
	pb.UnimplementedInternalServiceServer
	DB                         database.Ext
	JSM                        nats.JetStreamManagement
	orderService               IOrderServiceForInternalService
	billItemService            IBillItemServiceForInternalService
	studentProductService      IStudentProductServiceForInternalService
	billingScheduleService     IBillingScheduleForInternalService
	upcomingBillItemService    IUpcomingBillItemServiceForInternalService
	taxService                 ITaxServiceForInternalService
	discountService            IDiscountServiceForInternalService
	productService             IProductServiceForInternalService
	priceService               IPriceServiceForInternalService
	studentPackageService      IStudentPackageForInternalService
	packageService             IPackageForInternalService
	studentService             IStudentForInternalService
	subscriptionService        ISubscriptionServiceForInternalService
	studentPackageOrderService IStudentPackageOrderForInternalService
}

// NewInternalService Todo: Need to breakdown cronjob services
func NewInternalService(db database.Ext, jsm nats.JetStreamManagement, kafka kafka.KafkaManagement, config configs.CommonConfig) *InternalService {
	return &InternalService{
		DB:                         db,
		JSM:                        jsm,
		studentProductService:      studentProductService.NewStudentProductService(),
		orderService:               orderService.NewOrderService(),
		billItemService:            billingItemService.NewBillItemService(),
		billingScheduleService:     billingScheduleService.NewBillingScheduleService(),
		upcomingBillItemService:    upcomingBillItemService.NewUpcomingBillItemService(),
		taxService:                 taxService.NewTaxService(),
		discountService:            discount.NewDiscountService(),
		productService:             productService.NewProductService(),
		priceService:               price.NewPriceService(),
		studentPackageService:      studentPackageService.NewStudentPackage(),
		packageService:             packageService.NewPackageService(),
		studentService:             studentService.NewStudentService(),
		subscriptionService:        subscriptionService.NewSubscriptionService(jsm, db, kafka, config),
		studentPackageOrderService: studentPackageOrderService.NewStudentPackageOrder(),
	}
}
