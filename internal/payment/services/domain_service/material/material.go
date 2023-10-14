package service

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	export_entities "github.com/manabie-com/backend/internal/payment/export_entities"
	"github.com/manabie-com/backend/internal/payment/repositories"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MaterialService struct {
	materialRepo interface {
		GetByID(ctx context.Context, db database.QueryExecer, materialID string) (entities.Material, error)
		GetAll(ctx context.Context, db database.QueryExecer) ([]entities.Material, error)
	}
	productRepo interface {
		GetByIDsForExport(ctx context.Context, db database.QueryExecer, entitiesIDs []string) ([]entities.Product, error)
	}
}

func (s *MaterialService) GetMaterialByID(ctx context.Context, db database.QueryExecer, materialID string) (material entities.Material, err error) {

	material, err = s.materialRepo.GetByID(ctx, db, materialID)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when checking material id: %v", err.Error())
		return
	}
	return
}

func (s *MaterialService) GetAllMaterialsForExport(ctx context.Context, db database.QueryExecer) (materials []export_entities.ProductMaterialExport, err error) {
	materialsRepo, err := s.materialRepo.GetAll(ctx, db)
	if err != nil {
		return
	}
	materialIDs := make([]string, 0, len(materialsRepo))
	for _, material := range materialsRepo {
		materialIDs = append(materialIDs, material.MaterialID.String)
	}

	products, err := s.productRepo.GetByIDsForExport(ctx, db, materialIDs)
	if err != nil {
		return
	}

	mapProductAndProductIDs := make(map[string]entities.Product, len(products))
	for _, product := range products {
		mapProductAndProductIDs[product.ProductID.String] = product
	}

	for _, material := range materialsRepo {
		product, exist := mapProductAndProductIDs[material.MaterialID.String]
		if !exist {
			err = status.Errorf(codes.Internal, "Missing product info with id: %s", material.MaterialID.String)
			return
		}
		materials = append(materials, export_entities.ProductMaterialExport{
			MaterialID:           material.MaterialID.String,
			Name:                 product.Name.String,
			MaterialType:         material.MaterialType.String,
			TaxID:                product.TaxID.String,
			ProductTag:           product.ProductTag.String,
			ProductPartnerID:     product.ProductPartnerID.String,
			AvailableFrom:        product.AvailableFrom.Time,
			AvailableUntil:       product.AvailableUntil.Time,
			CustomBillingPeriod:  product.CustomBillingPeriod.Time,
			CustomBillingDate:    material.CustomBillingDate.Time,
			DisableProRatingFlag: product.DisableProRatingFlag.Bool,
			BillingScheduleID:    product.BillingScheduleID.String,
			Remarks:              product.Remarks.String,
			IsUnique:             product.IsUnique.Bool,
			IsArchived:           material.IsArchived.Bool,
		})
	}
	return
}

func NewMaterialService() *MaterialService {
	return &MaterialService{
		materialRepo: &repositories.MaterialRepo{},
		productRepo:  &repositories.ProductRepo{},
	}
}
