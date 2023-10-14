package service

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	billItemService "github.com/manabie-com/backend/internal/payment/services/domain_service/billing/bill_item"
	taxService "github.com/manabie-com/backend/internal/payment/services/domain_service/tax"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

type ITaxServiceForCustomOrder interface {
	IsValidTaxForCustomOrder(
		ctx context.Context,
		db database.QueryExecer,
		customBillingItem *pb.CustomBillingItem,
	) (err error)
}

type IBillItemServiceForCustomOrder interface {
	CreateCustomBillItem(
		ctx context.Context,
		db database.QueryExecer,
		customBillingItem *pb.CustomBillingItem,
		order entities.Order,
		locationName string,
	) (err error)
}

type BillingServiceForCustomOrder struct {
	TaxService      ITaxServiceForCustomOrder
	BillItemService IBillItemServiceForCustomOrder
}

func (s *BillingServiceForCustomOrder) CreateBillItemForCustomOrder(
	ctx context.Context,
	db database.QueryExecer,
	req *pb.CreateCustomBillingRequest,
	order entities.Order,
	locationName string,
) (
	err error,
) {
	for _, item := range req.CustomBillingItems {
		err = utils.GroupErrorFunc(
			s.TaxService.IsValidTaxForCustomOrder(ctx, db, item),
			s.BillItemService.CreateCustomBillItem(ctx, db, item, order, locationName),
		)
		if err != nil {
			return
		}
	}
	return
}

func NewBillingServiceForCustomOrder() *BillingServiceForCustomOrder {
	return &BillingServiceForCustomOrder{
		BillItemService: billItemService.NewBillItemService(),
		TaxService:      taxService.NewTaxService(),
	}
}
