package services

import (
	"context"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/discount/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type ProductGroupService struct {
	DB               database.Ext
	ProductGroupRepo interface {
		GetByID(ctx context.Context, db database.QueryExecer, productGroupID string) (entities.ProductGroup, error)
	}
	ProductGroupMappingRepo interface {
		GetByProductID(ctx context.Context, db database.QueryExecer, productID string) ([]*entities.ProductGroupMapping, error)
	}
}

func (s *ProductGroupService) RetrieveProductGroupsOfProductIDByDiscountType(ctx context.Context, productID string, discountType string) (productGroups []entities.ProductGroup, err error) {
	productGroupMapping, err := s.ProductGroupMappingRepo.GetByProductID(ctx, s.DB, productID)
	if err != nil {
		return
	}

	productGroups = []entities.ProductGroup{}
	for _, mapping := range productGroupMapping {
		group, err := s.ProductGroupRepo.GetByID(ctx, s.DB, mapping.ProductGroupID.String)
		if err != nil {
			return productGroups, err
		}

		if group.DiscountType.String == discountType || discountType == "" {
			productGroups = append(productGroups, group)
		}
	}

	return
}

func (s *ProductGroupService) RetrieveEligibleProductGroupsOfStudentProductsByDiscountType(ctx context.Context, studentProducts []entities.StudentProduct, discountType string) (productDiscountGroups []entities.ProductDiscountGroup) {
	productDiscountGroups = []entities.ProductDiscountGroup{}
	for _, studentProduct := range studentProducts {
		productGroups, err := s.RetrieveProductGroupsOfProductIDByDiscountType(ctx, studentProduct.ProductID.String, discountType)
		if err != nil || len(productGroups) == 0 {
			continue
		}

		productDiscountGroup := entities.ProductDiscountGroup{
			StudentProduct: studentProduct,
			ProductGroups:  productGroups,
			DiscountType:   discountType,
		}

		productDiscountGroups = append(productDiscountGroups, productDiscountGroup)
	}

	return
}

func NewProductGroupService(db database.Ext) *ProductGroupService {
	return &ProductGroupService{
		DB:                      db,
		ProductGroupRepo:        &repositories.ProductGroupRepo{},
		ProductGroupMappingRepo: &repositories.ProductGroupMappingRepo{},
	}
}
