package services

import (
	"context"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/discount/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type StudentProductService struct {
	DB                 database.Ext
	StudentProductRepo interface {
		GetActiveStudentProductsByStudentIDAndLocationID(ctx context.Context, db database.QueryExecer, userID string, locationID string) ([]*entities.StudentProduct, error)
		GetByID(ctx context.Context, db database.QueryExecer, studentProductID string) (entities.StudentProduct, error)
		GetByIDs(ctx context.Context, db database.QueryExecer, studentProductIDs []string) ([]entities.StudentProduct, error)
	}
	BillItemRepo interface {
		GetLastBillItemOfStudentProduct(ctx context.Context, db database.QueryExecer, studentProductID string) (entities.BillItem, error)
	}
	DiscountRepo interface {
		GetByID(ctx context.Context, db database.QueryExecer, discountID string) (entities.Discount, error)
	}
	OrderItemRepo interface {
		GetStudentProductIDsByOrderID(ctx context.Context, db database.QueryExecer, orderID string) ([]string, error)
	}
}

func (s *StudentProductService) RetrieveActiveStudentProductsOfStudentInLocation(
	ctx context.Context,
	db database.QueryExecer,
	userID string,
	locationID string,
) (
	studentProducts []*entities.StudentProduct,
	err error,
) {
	return s.StudentProductRepo.GetActiveStudentProductsByStudentIDAndLocationID(ctx, db, userID, locationID)
}

func (s *StudentProductService) RetrieveStudentProductByID(
	ctx context.Context,
	db database.QueryExecer,
	studentProductID string,
) (
	studentProduct entities.StudentProduct,
	err error,
) {
	return s.StudentProductRepo.GetByID(ctx, db, studentProductID)
}

func (s *StudentProductService) RetrieveStudentProductsByIDs(
	ctx context.Context,
	db database.QueryExecer,
	studentProductIDs []string,
) (
	studentProduct []entities.StudentProduct,
	err error,
) {
	return s.StudentProductRepo.GetByIDs(ctx, db, studentProductIDs)
}

func (s *StudentProductService) RetrieveDiscountOfStudentProduct(
	ctx context.Context,
	studentProductID string,
) (
	discount entities.Discount,
	err error,
) {
	latestBillItemOfStudentProduct, err := s.BillItemRepo.GetLastBillItemOfStudentProduct(ctx, s.DB, studentProductID)
	if err != nil {
		return
	}

	if latestBillItemOfStudentProduct.DiscountID.Status == pgtype.Present {
		discount, err = s.DiscountRepo.GetByID(ctx, s.DB, latestBillItemOfStudentProduct.DiscountID.String)
		if err != nil {
			return
		}
	}

	return
}

func (s *StudentProductService) RetrieveStudentProductsByOrderID(
	ctx context.Context,
	db database.QueryExecer,
	orderID string,
) (
	studentProduct []entities.StudentProduct,
	err error,
) {
	studentProductIDs, err := s.OrderItemRepo.GetStudentProductIDsByOrderID(ctx, db, orderID)
	if err != nil {
		return
	}

	return s.StudentProductRepo.GetByIDs(ctx, db, studentProductIDs)
}

func NewStudentProductService(db database.Ext) *StudentProductService {
	return &StudentProductService{
		DB:                 db,
		StudentProductRepo: &repositories.StudentProductRepo{},
		BillItemRepo:       &repositories.BillItemRepo{},
		DiscountRepo:       &repositories.DiscountRepo{},
		OrderItemRepo:      &repositories.OrderItemRepo{},
	}
}
