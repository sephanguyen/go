package studentbilling

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	billItemService "github.com/manabie-com/backend/internal/payment/services/domain_service/billing/bill_item"
	locationService "github.com/manabie-com/backend/internal/payment/services/domain_service/location"
	materialService "github.com/manabie-com/backend/internal/payment/services/domain_service/material"
	orderService "github.com/manabie-com/backend/internal/payment/services/domain_service/order"
	packageService "github.com/manabie-com/backend/internal/payment/services/domain_service/package"
	productService "github.com/manabie-com/backend/internal/payment/services/domain_service/product"
	studentProductService "github.com/manabie-com/backend/internal/payment/services/domain_service/student_product"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

type IBillItemServiceForStudentBilling interface {
	GetBillItemDescriptionByStudentIDAndLocationIDs(
		ctx context.Context,
		db database.Ext,
		studentID string,
		locationIDs []string,
		from int64,
		limit int64,
	) (
		billingDescriptions []utils.BillItemForRetrieveApi,
		total int,
		err error,
	)
	GetBillItemInfoByOrderIDAndUniqueByProductID(
		ctx context.Context,
		db database.Ext,
		orderID string,
	) (
		billItem []*entities.BillItem,
		err error,
	)
	GetLatestBillItemByStudentProductIDForStudentBilling(
		ctx context.Context,
		db database.Ext,
		studentProductID string,
	) (
		billItem entities.BillItem,
		err error,
	)
	GetMapPresentAndFutureBillItemInfo(
		ctx context.Context,
		db database.QueryExecer,
		studentProductIDs []string,
		studentID string) (
		mapStudentProductIDAndBillItem map[string]*entities.BillItem,
		err error,
	)
	GetMapPastBillItemInfo(ctx context.Context, db database.QueryExecer, studentProductIDs []string, studentID string) (mapStudentProductIDAndBillItem map[string]*entities.BillItem, err error)
	GetUpcomingBilling(ctx context.Context, db database.QueryExecer, studentProductID string, studentID string) (upcomingBillingItem *entities.BillItem, err error)
}

type IOrderServiceForStudentBilling interface {
	GetOrderTypeByOrderID(
		ctx context.Context,
		db database.Ext,
		orderID string,
	) (
		orderType string,
		err error,
	)
	GetOrdersByStudentIDAndLocationIDs(
		ctx context.Context,
		db database.Ext,
		studentID string,
		locationIDs []string,
		from int64,
		limit int64,
	) (
		orderInfo []*entities.Order,
		total int,
		err error,
	)
}

type IStudentProductForStudentBilling interface {
	GetStudentProductByStudentIDAndLocationIDs(
		ctx context.Context,
		db database.Ext,
		studentID string,
		locations []string,
		from int64,
		limit int64,
	) (
		studentIDs []string,
		studentProducts []*entities.StudentProduct,
		total int,
		err error,
	)
	GetStudentAssociatedProductByStudentProductID(
		ctx context.Context,
		db database.Ext,
		studentProductID string,
		from int64,
		limit int64,
	) (
		studentProductIDs []string,
		studentProducts []*entities.StudentProduct,
		total int,
		err error,
	)
}

type IMaterialServiceForStudentBilling interface {
	GetMaterialByID(ctx context.Context, db database.QueryExecer, materialID string) (material entities.Material, err error)
}

type ILocationServiceForStudentBilling interface {
	GetLocationInfoByID(ctx context.Context, db database.Ext, locationID string) (locationInfo *pb.LocationInfo, err error)
}
type IPackageServiceForStudentBilling interface {
	GetTotalAssociatedPackageWithCourseIDAndPackageID(ctx context.Context, db database.Ext, packageID string, courseIDs []string) (total int32, err error)
}

type IProductServiceForStudentBilling interface {
	GetProductSettingByProductID(ctx context.Context, db database.QueryExecer, productID string) (product entities.ProductSetting, err error)
}

type StudentBilling struct {
	DB                    database.Ext
	BillItemService       IBillItemServiceForStudentBilling
	OrderService          IOrderServiceForStudentBilling
	StudentProductService IStudentProductForStudentBilling
	MaterialService       IMaterialServiceForStudentBilling
	LocationService       ILocationServiceForStudentBilling
	PackageService        IPackageServiceForStudentBilling
	ProductService        IProductServiceForStudentBilling
}

func NewStudentBilling(db database.Ext) *StudentBilling {
	return &StudentBilling{
		DB:                    db,
		BillItemService:       billItemService.NewBillItemService(),
		OrderService:          orderService.NewOrderService(),
		StudentProductService: studentProductService.NewStudentProductService(),
		MaterialService:       materialService.NewMaterialService(),
		LocationService:       locationService.NewLocationService(),
		PackageService:        packageService.NewPackageService(),
		ProductService:        productService.NewProductService(),
	}
}
