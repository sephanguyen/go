package ordermgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	locationService "github.com/manabie-com/backend/internal/payment/services/domain_service/location"
	productService "github.com/manabie-com/backend/internal/payment/services/domain_service/product"
	"github.com/manabie-com/backend/internal/payment/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

type IProductServiceForProductList interface {
	GetProductStatsByFilter(ctx context.Context, db database.QueryExecer, req *pb.RetrieveListOfProductsRequest) (productStats entities.ProductStats, err error)
	GetListOfProductsByFilter(ctx context.Context, db database.QueryExecer, req *pb.RetrieveListOfProductsRequest, from int64, limit int64) (products []entities.Product, err error)
	GetGradeIDsByProductID(ctx context.Context, db database.QueryExecer, productID string) (gradeIDs []string, err error)
	GetLocationIDsWithProductID(ctx context.Context, db database.QueryExecer, productID string) (locationIDs []string, err error)
	GetProductTypeByProductID(ctx context.Context, db database.QueryExecer, productID string, currentProductType string) (productType pb.ProductSpecificType, err error)
	GetGradeNamesByIDs(ctx context.Context, db database.Ext, gradeIDs []string) (gradeNames []string, err error)
}

type ILocationServiceForProductList interface {
	GetLocationsByIDs(ctx context.Context, db database.Ext, locationIDs []string) (locations []entities.Location, err error)
}

type ProductList struct {
	DB database.Ext

	ProductService  IProductServiceForProductList
	LocationService ILocationServiceForProductList
}

func (s *ProductList) RetrieveListOfProducts(ctx context.Context, req *pb.RetrieveListOfProductsRequest) (res *pb.RetrieveListOfProductsResponse, err error) {
	fromIdx := int64(0)
	limit := int64(req.Paging.Limit)
	switch u := req.Paging.Offset.(type) {
	case *cpb.Paging_OffsetInteger:
		fromIdx = u.OffsetInteger
	case *cpb.Paging_OffsetCombined:
		fromIdx = u.OffsetCombined.OffsetInteger
	default:
	}

	var (
		productStats                 entities.ProductStats
		mapGradeWithProductID        map[string][]string
		products                     []entities.Product
		prevPage                     *cpb.Paging
		nextPage                     *cpb.Paging
		mapProductIDsWithLocationIDs map[string][]string
		locations                    map[string]entities.Location
	)

	productStats, err = s.ProductService.GetProductStatsByFilter(ctx, s.DB, req)
	if err != nil {
		err = fmt.Errorf("error while get product stats: %v", err)
		return
	}
	prevPage, nextPage, err = utils.ConvertCommonPaging(int(productStats.TotalItems.Int), fromIdx, limit)
	if err != nil {
		err = fmt.Errorf("error while pagging with filter: %v", err)
		return nil, err
	}
	res = &pb.RetrieveListOfProductsResponse{
		NextPage:        nextPage,
		PreviousPage:    prevPage,
		TotalItems:      uint32(productStats.TotalItems.Int),
		TotalOfActive:   uint32(productStats.TotalOfActive.Int),
		TotalOfInactive: uint32(productStats.TotalOfInactive.Int),
	}

	products, err = s.ProductService.GetListOfProductsByFilter(ctx, s.DB, req, fromIdx, limit)
	if err != nil {
		err = fmt.Errorf("error while get list of products by filter: %v", err)
		return
	}

	if len(products) == 0 {
		return
	}

	result := make([]*pb.RetrieveListOfProductsResponse_Product, 0, len(products))

	mapProductIDsWithLocationIDs, locations, err = s.getProductsLocations(ctx, s.DB, products)
	if err != nil {
		err = fmt.Errorf("error while get locations of products: %v", err)
		return
	}
	mapGradeWithProductID, err = s.getGradeOfProductsReturningMapGradeWithProductID(ctx, s.DB, products)
	if err != nil {
		err = fmt.Errorf("error while get grades of products: %v", err)
		return
	}
	for _, product := range products {
		var productLocations []*pb.LocationInfo
		if len(mapProductIDsWithLocationIDs[product.ProductID.String]) > 0 {
			for _, location := range mapProductIDsWithLocationIDs[product.ProductID.String] {
				locationInfo := &pb.LocationInfo{
					LocationId:   locations[location].LocationID.String,
					LocationName: locations[location].Name.String,
				}
				productLocations = append(productLocations, locationInfo)
			}
		}
		productStatus := pb.ProductStatus_PRODUCT_STATUS_ACTIVE
		if product.AvailableFrom.Time.After(time.Now()) || product.AvailableUntil.Time.Before(time.Now()) {
			productStatus = pb.ProductStatus_PRODUCT_STATUS_INACTIVE
		}
		var (
			productTypeValue pb.ProductSpecificType
			gradeNames       []string
		)
		gradeNames, err = s.ProductService.GetGradeNamesByIDs(ctx, s.DB, mapGradeWithProductID[product.ProductID.String])
		if err != nil {
			err = fmt.Errorf("error while get grade names with filter: %v", err)
			return nil, err
		}
		productTypeValue, err = s.ProductService.GetProductTypeByProductID(ctx, s.DB, product.ProductID.String, product.ProductType.String)
		if err != nil {
			err = fmt.Errorf("error while get product type with filter: %v", err)
			return nil, err
		}
		result = append(result, &pb.RetrieveListOfProductsResponse_Product{
			ProductName:   product.Name.String,
			ProductType:   &productTypeValue,
			ProductStatus: productStatus,
			Grades:        gradeNames,
			LocationInfo:  productLocations,
			ProductId:     product.ProductID.String,
		})
	}
	res.Items = result
	return
}

func (s *ProductList) getProductsLocations(ctx context.Context, db database.QueryExecer, products []entities.Product) (mapProductIDsWithLocationIDs map[string][]string, locations map[string]entities.Location, err error) {
	var (
		productIDs        []string
		locationsIDs      []string
		productsLocations []entities.Location
	)

	for _, product := range products {
		productIDs = append(productIDs, product.ProductID.String)
	}

	mapProductIDsWithLocationIDs = make(map[string][]string)
	for _, productID := range productIDs {
		mapProductIDsWithLocationIDs[productID], err = s.ProductService.GetLocationIDsWithProductID(ctx, db, productID)
		if err != nil {
			return
		}
	}

	for _, product := range mapProductIDsWithLocationIDs {
		for _, locationID := range product {
			if notContains(locationsIDs, locationID) {
				locationsIDs = append(locationsIDs, locationID)
			}
		}
	}

	productsLocations, err = s.LocationService.GetLocationsByIDs(ctx, s.DB, locationsIDs)
	if err != nil {
		err = fmt.Errorf("error while getting locations: %v", err)
		return
	}
	locations = make(map[string]entities.Location)
	for _, productsLocation := range productsLocations {
		locations[productsLocation.LocationID.String] = productsLocation
	}
	return
}

func notContains(array []string, value string) bool {
	for _, v := range array {
		if v == value {
			return false
		}
	}
	return true
}

func (s *ProductList) getGradeOfProductsReturningMapGradeWithProductID(ctx context.Context, db database.QueryExecer, products []entities.Product) (
	mapGradesWithProductID map[string][]string, err error) {
	var gradeIDs []string
	mapGradesWithProductID = make(map[string][]string, len(products))
	for _, product := range products {
		gradeIDs, err = s.ProductService.GetGradeIDsByProductID(ctx, db, product.ProductID.String)
		if err != nil {
			err = fmt.Errorf("error while get grade of product %v: %v", product.ProductID.String, err)
			return nil, err
		}
		mapGradesWithProductID[product.ProductID.String] = append(mapGradesWithProductID[product.ProductID.String], gradeIDs...)
	}
	return
}

func NewRetrieveListOfProducts(db database.Ext) *ProductList {
	return &ProductList{
		DB:              db,
		ProductService:  productService.NewProductService(),
		LocationService: locationService.NewLocationService(),
	}
}
