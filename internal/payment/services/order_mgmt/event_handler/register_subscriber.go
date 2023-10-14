package eventhandler

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/payment/repositories"

	"go.uber.org/zap"
)

func RegisterDiscountEventHandler(
	jsm nats.JetStreamManagement,
	logger *zap.Logger,
	db database.Ext,
	orderService IOrderServiceServiceForDiscountEventSubscription,
) error {
	s := &DiscountEventSubscription{
		JSM:    jsm,
		DB:     db,
		Logger: logger,

		BillingRatioRepo:                   &repositories.BillingRatioRepo{},
		BillingSchedulePeriodRepo:          &repositories.BillingSchedulePeriodRepo{},
		BillItemRepo:                       &repositories.BillItemRepo{},
		OrderItemRepo:                      &repositories.OrderItemRepo{},
		OrderItemCourseRepo:                &repositories.OrderItemCourseRepo{},
		OrderService:                       orderService,
		PackageRepo:                        &repositories.PackageRepo{},
		PackageCourseRepo:                  &repositories.PackageCourseRepo{},
		PackageQuantityTypeRepo:            &repositories.PackageQuantityTypeMappingRepo{},
		ProductRepo:                        &repositories.ProductRepo{},
		ProductPriceRepo:                   &repositories.ProductPriceRepo{},
		TaxRepo:                            &repositories.TaxRepo{},
		StudentEnrollmentStatusHistoryRepo: &repositories.StudentEnrollmentStatusHistoryRepo{},
	}

	return s.Subscribe()
}
