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

type FeeService struct {
	feeRepo interface {
		GetAll(ctx context.Context, db database.QueryExecer) ([]entities.Fee, error)
	}
	productRepo interface {
		GetByIDsForExport(ctx context.Context, db database.QueryExecer, entitiesIDs []string) ([]entities.Product, error)
	}
}

func (s *FeeService) GetAllFeesForExport(ctx context.Context, db database.QueryExecer) (fees []export_entities.ProductFeeExport, err error) {
	feesRepo, err := s.feeRepo.GetAll(ctx, db)
	if err != nil {
		return
	}
	feeIDs := make([]string, 0, len(feesRepo))
	for _, fee := range feesRepo {
		feeIDs = append(feeIDs, fee.FeeID.String)
	}

	products, err := s.productRepo.GetByIDsForExport(ctx, db, feeIDs)
	if err != nil {
		return
	}

	mapProductAndProductIDs := make(map[string]entities.Product, len(products))
	for _, product := range products {
		mapProductAndProductIDs[product.ProductID.String] = product
	}

	for _, fee := range feesRepo {
		product, exist := mapProductAndProductIDs[fee.FeeID.String]
		if !exist {
			err = status.Errorf(codes.Internal, "Missing product info with id: %s", fee.FeeID.String)
			return
		}
		fees = append(fees, export_entities.ProductFeeExport{
			FeeID:                fee.FeeID.String,
			Name:                 product.Name.String,
			FeeType:              fee.FeeType.String,
			TaxID:                product.TaxID.String,
			ProductTag:           product.ProductTag.String,
			ProductPartnerID:     product.ProductPartnerID.String,
			AvailableFrom:        product.AvailableFrom.Time,
			AvailableUntil:       product.AvailableUntil.Time,
			CustomBillingPeriod:  product.CustomBillingPeriod.Time,
			BillingScheduleID:    product.BillingScheduleID.String,
			DisableProRatingFlag: product.DisableProRatingFlag.Bool,
			Remarks:              product.Remarks.String,
			IsUnique:             product.IsUnique.Bool,
			IsArchived:           fee.IsArchived.Bool,
		})
	}

	return
}

func NewFeeService() *FeeService {
	return &FeeService{
		feeRepo:     &repositories.FeeRepo{},
		productRepo: &repositories.ProductRepo{},
	}
}
