package order_detail

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	billItemService "github.com/manabie-com/backend/internal/payment/services/domain_service/billing/bill_item"
	orderService "github.com/manabie-com/backend/internal/payment/services/domain_service/order"
	orderItemService "github.com/manabie-com/backend/internal/payment/services/domain_service/order_item"
	studentProductService "github.com/manabie-com/backend/internal/payment/services/domain_service/student_product"
	"github.com/manabie-com/backend/internal/payment/utils"
)

type IBillItemServiceForOrderDetail interface {
	GetBillItemDescriptionsByOrderIDWithPaging(
		ctx context.Context,
		db database.Ext,
		orderID string,
		from int64,
		limit int64,
	) (
		billingDescriptions []utils.BillItemForRetrieveApi,
		total int,
		err error,
	)
	GetFirstBillItemsByOrderIDAndProductID(
		ctx context.Context,
		db database.Ext,
		orderID string,
		from int64,
		limit int64,
	) (
		billItems []*entities.BillItem,
		total int,
		err error,
	)
	BuildMapBillItemWithProductIDByOrderIDAndProductIDs(
		ctx context.Context, db database.QueryExecer, orderID string, productIDs []string) (
		mapProductIDAndBillItem map[string]entities.BillItem, err error)
}

type IStudentProductForOrderDetail interface {
	GetStudentProductByStudentProductID(ctx context.Context,
		db database.QueryExecer,
		studentProductID string,
	) (
		studentProduct entities.StudentProduct,
		err error,
	)
	GetStudentProductsByStudentProductIDs(ctx context.Context,
		db database.Ext,
		studentProductIDs []string,
	) (
		studentProduct []entities.StudentProduct,
		err error,
	)
}

type IOrderServiceForOrderDetail interface {
	GetOrderTypeByOrderID(ctx context.Context, db database.Ext, orderID string) (orderType string, err error)
}

type IOrderItemServiceForOrderDetail interface {
	CountOrderItemsByOrderID(ctx context.Context, db database.Ext, orderID string) (count int, err error)
	GetOrderItemsByOrderIDWithPaging(ctx context.Context, db database.Ext, orderID string, from int64,
		limit int64) (orderItems []entities.OrderItem, err error)
}

type OrderDetail struct {
	DB                    database.Ext
	BillItemService       IBillItemServiceForOrderDetail
	StudentProductService IStudentProductForOrderDetail
	OrderService          IOrderServiceForOrderDetail
	OrderItemService      IOrderItemServiceForOrderDetail
}

func NewOrderDetail(db database.Ext) *OrderDetail {
	return &OrderDetail{
		DB:                    db,
		BillItemService:       billItemService.NewBillItemService(),
		StudentProductService: studentProductService.NewStudentProductService(),
		OrderService:          orderService.NewOrderService(),
		OrderItemService:      orderItemService.NewOrderItemService(),
	}
}
