package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/repositories"

	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UpcomingBillItemService struct {
	UpcomingBillItemRepo interface {
		GetUpcomingBillItemsForGenerate(
			ctx context.Context,
			db database.QueryExecer,
		) (billItems []entities.UpcomingBillItem, err error)
		Create(
			ctx context.Context,
			db database.QueryExecer,
			e *entities.UpcomingBillItem,
		) (err error)
		AddUpcomingExecuteNote(
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
		VoidUpcomingBillItemsByOrderID(
			ctx context.Context,
			db database.QueryExecer,
			orderID string,
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
		) (upcomingBillItems []entities.UpcomingBillItem, err error)
	}
}

func (s *UpcomingBillItemService) GetUpcomingBillItemByOrderIDProductIDBillingSchedulePeriodID(ctx context.Context, db database.QueryExecer, orderID string, productID string, billingSchedulePeriodID string) (oldUpcomingBillItem []entities.UpcomingBillItem, err error) {
	oldUpcomingBillItem, err = s.UpcomingBillItemRepo.GetUpcomingBillItemByOrderIDProductIDBillingSchedulePeriodID(ctx, db, orderID, productID, billingSchedulePeriodID)
	return
}

func (s *UpcomingBillItemService) GetUpcomingBillItemsForGenerate(
	ctx context.Context,
	db database.QueryExecer,
) (billItems []entities.UpcomingBillItem, err error) {
	billItems, err = s.UpcomingBillItemRepo.GetUpcomingBillItemsForGenerate(ctx, db)
	return
}

func (s *UpcomingBillItemService) CreateUpcomingBillItem(
	ctx context.Context,
	db database.QueryExecer,
	billItem entities.BillItem,
) (err error) {
	upComingBillItem := new(entities.UpcomingBillItem)
	err = multierr.Combine(
		upComingBillItem.BillItemSequenceNumber.Set(billItem.BillItemSequenceNumber.Int),
		upComingBillItem.OrderID.Set(billItem.OrderID.String),
		upComingBillItem.ProductID.Set(billItem.ProductID.String),
		upComingBillItem.StudentProductID.Set(billItem.StudentProductID.String),
		upComingBillItem.DiscountID.Set(billItem.DiscountID.String),
		upComingBillItem.TaxID.Set(billItem.TaxID.String),
		upComingBillItem.BillingSchedulePeriodID.Set(billItem.BillSchedulePeriodID.String),
		upComingBillItem.BillingDate.Set(billItem.BillDate.Time),
		upComingBillItem.IsGenerated.Set(false),
		upComingBillItem.ProductDescription.Set(billItem.ProductDescription.String),
		upComingBillItem.ExecuteNote.Set(nil),
	)
	if err != nil {
		err = fmt.Errorf("err while assigning upcoming bill item: %v", err.Error())
		return
	}

	err = s.UpcomingBillItemRepo.Create(ctx, db, upComingBillItem)
	return
}

func (s *UpcomingBillItemService) AddExecuteNoteForCurrentUpcomingBillItem(ctx context.Context, db database.QueryExecer, upcomingBillItem entities.UpcomingBillItem, err error) (importErr error) {
	importErr = s.UpcomingBillItemRepo.AddUpcomingExecuteNote(ctx, db, upcomingBillItem, err)
	return
}

func (s *UpcomingBillItemService) UpdateCurrentUpcomingBillItemStatus(
	ctx context.Context,
	db database.QueryExecer,
	upcomingBillItem entities.UpcomingBillItem,
) (err error) {
	err = s.UpcomingBillItemRepo.UpdateCurrentUpcomingBillItemStatus(ctx, db, upcomingBillItem)
	return
}

func (s *UpcomingBillItemService) SetLastUpcomingBillItem(
	ctx context.Context,
	db database.QueryExecer,
	upcomingBillItem entities.UpcomingBillItem,
) (err error) {
	err = s.UpcomingBillItemRepo.SetLastUpcomingBillItem(ctx, db, upcomingBillItem)
	return
}

func (s *UpcomingBillItemService) VoidUpcomingBillItemsByOrder(
	ctx context.Context,
	db database.QueryExecer,
	order entities.Order,
) (err error) {
	err = s.UpcomingBillItemRepo.VoidUpcomingBillItemsByOrderID(ctx, db, order.OrderID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "voiding upcoming bill items by order id have error: %v", err)
		return
	}
	return
}

func NewUpcomingBillItemService() *UpcomingBillItemService {
	return &UpcomingBillItemService{
		UpcomingBillItemRepo: &repositories.UpcomingBillItemRepo{},
	}
}
