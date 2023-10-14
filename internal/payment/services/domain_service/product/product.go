package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/repositories"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductService struct {
	productRepo interface {
		GetByIDForUpdate(ctx context.Context, db database.QueryExecer, entitiesID string) (entities.Product, error)
		GetByIDs(ctx context.Context, db database.QueryExecer, entitiesIDs []string) ([]entities.Product, error)
		GetProductStatsByFilter(ctx context.Context, db database.QueryExecer, filter repositories.ProductListFilter) (productStats entities.ProductStats, err error)
		GetProductsByFilter(ctx context.Context, db database.QueryExecer, filter repositories.ProductListFilter) (products []entities.Product, err error)
		GetProductIDsByProductTypeAndOrderID(ctx context.Context, db database.QueryExecer, productType, orderID string) (productIDs []string, err error)
	}
	productGradeRepo interface {
		GetByGradeAndProductIDForUpdate(ctx context.Context, db database.QueryExecer, grade string, productID string) (entities.ProductGrade, error)
		GetGradeIDsByProductID(ctx context.Context, db database.QueryExecer, productID string) ([]string, error)
	}
	productLocationRepo interface {
		GetByLocationIDAndProductIDForUpdate(ctx context.Context, db database.QueryExecer, locationID string, productID string) (entities.ProductLocation, error)
		GetLocationIDsWithProductID(ctx context.Context, db database.QueryExecer, productID string) (locationIDs []string, err error)
	}
	gradeRepo interface {
		GetByID(ctx context.Context, db database.QueryExecer, gradeID string) (entities.Grade, error)
		GetGradeNamesByGradeIDs(ctx context.Context, db database.QueryExecer, gradeIDs []string) (gradeNames []string, err error)
	}
	packageRepo interface {
		GetByIDForUpdate(ctx context.Context, db database.QueryExecer, packageID string) (entities.Package, error)
	}
	materialRepo interface {
		GetByIDForUpdate(ctx context.Context, db database.QueryExecer, materialID string) (entities.Material, error)
	}
	feeRepo interface {
		GetFeeByID(ctx context.Context, db database.QueryExecer, feeID string) (entities.Fee, error)
	}
	productSettingRepo interface {
		GetByID(ctx context.Context, db database.QueryExecer, productID string) (entities.ProductSetting, error)
	}
}

func (s *ProductService) VerifiedProductWithStudentInfoReturnProductInfoAndBillingType(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (
	productInfo entities.Product,
	isOnetimeProduct bool,
	isDisableProRatingFlag bool,
	productType pb.ProductType,
	gradeName string,
	productSetting entities.ProductSetting,
	err error,
) {
	productInfo,
		isOnetimeProduct,
		isDisableProRatingFlag,
		productType,
		productSetting,
		err = s.VerifiedProductReturnProductInfoAndBillingType(ctx, db, orderItemData)
	if err != nil {
		return
	}

	gradeEntities, err := s.gradeRepo.GetByID(ctx, db, orderItemData.StudentInfo.GradeID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting grade with id %v has error %v ", orderItemData.StudentInfo.GradeID.String, err.Error())
		return
	}
	gradeName = gradeEntities.Name.String

	_, err = s.productGradeRepo.GetByGradeAndProductIDForUpdate(ctx, db, orderItemData.StudentInfo.GradeID.String, orderItemData.OrderItem.ProductId)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting product id %v and grade id %v has error %v ", orderItemData.OrderItem.ProductId, orderItemData.StudentInfo.GradeID.String, err.Error())
		return
	}

	_, err = s.productLocationRepo.GetByLocationIDAndProductIDForUpdate(ctx, db, orderItemData.Order.LocationID.String, orderItemData.OrderItem.ProductId)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting product location with id %v has error %v ", orderItemData.OrderItem.ProductId, err.Error())
		return
	}

	return
}

func (s *ProductService) VerifiedProductReturnProductInfoAndBillingType(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (
	productInfo entities.Product,
	isOnetimeProduct bool,
	isDisableProRatingFlag bool,
	productType pb.ProductType,
	productSetting entities.ProductSetting,
	err error,
) {
	productInfo, err = s.productRepo.GetByIDForUpdate(ctx, db, orderItemData.OrderItem.ProductId)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting product info with id %v has error %v ", orderItemData.OrderItem.ProductId, err.Error())
		return
	}

	productTypeEnum, ok := pb.ProductType_value[productInfo.ProductType.String]
	if !ok {
		err = status.Errorf(codes.Internal, "product type of product id %v is invalid ", orderItemData.OrderItem.ProductId)
		return
	}

	productType = pb.ProductType(productTypeEnum)

	productSetting, err = s.productSettingRepo.GetByID(ctx, db, orderItemData.OrderItem.ProductId)
	if err != nil {
		productSetting = entities.ProductSetting{
			ProductID:            pgtype.Text{String: orderItemData.OrderItem.ProductId, Status: pgtype.Present},
			IsPausable:           pgtype.Bool{Bool: true, Status: pgtype.Present},
			IsEnrollmentRequired: pgtype.Bool{Bool: false, Status: pgtype.Present},
		}
		err = nil
	}

	if productInfo.BillingScheduleID.Status != pgtype.Present {
		isOnetimeProduct = true
		return
	}

	if productInfo.DisableProRatingFlag.Status == pgtype.Present {
		isDisableProRatingFlag = productInfo.DisableProRatingFlag.Bool
	}

	return
}

func (s *ProductService) GetProductsByIDs(ctx context.Context, db database.Ext, productIDs []string) (products []entities.Product, err error) {
	products, err = s.productRepo.GetByIDs(ctx, db, productIDs)
	if err != nil {
		err = status.Errorf(
			codes.Internal, "Error when get products by ids: %v", err.Error())
		return
	}
	return
}

func (s *ProductService) GetProductStatsByFilter(ctx context.Context, db database.QueryExecer, req *pb.RetrieveListOfProductsRequest) (productStats entities.ProductStats, err error) {
	var filter repositories.ProductListFilter

	if len(req.Keyword) > 0 {
		filter.ProductName = req.Keyword
	}

	if req.Filter != nil {
		if len(req.Filter.ProductTypes) > 0 {
			var productTypes []*pb.ProductSpecificType
			for _, productType := range req.Filter.ProductTypes {
				productTypes = append(productTypes, productType)
			}
			filter.ProductTypes = productTypes
		}
		if len(req.Filter.StudentGrades) > 0 {
			var studentGrades []string
			studentGrades = append(studentGrades, req.Filter.StudentGrades...)
			filter.StudentGrades = studentGrades
		}
	}
	productStats, err = s.productRepo.GetProductStatsByFilter(ctx, db, filter)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when getting product stats by filter with error: %v", err)
		return
	}
	return
}

func (s *ProductService) GetListOfProductsByFilter(ctx context.Context, db database.QueryExecer, req *pb.RetrieveListOfProductsRequest, from int64, limit int64) (products []entities.Product, err error) {
	var filter repositories.ProductListFilter

	if len(req.Keyword) > 0 {
		filter.ProductName = req.Keyword
	}
	if req.Paging != nil {
		filter.Limit = &limit
		filter.Offset = &from
	}
	if req.ProductStatus != pb.ProductStatus_PRODUCT_STATUS_ALL {
		filter.ProductStatus = req.ProductStatus.String()
	}
	if req.Filter != nil {
		if len(req.Filter.ProductTypes) > 0 {
			var productTypes []*pb.ProductSpecificType
			for _, productType := range req.Filter.ProductTypes {
				productTypes = append(productTypes, productType)
			}
			filter.ProductTypes = productTypes
		}
		if len(req.Filter.StudentGrades) > 0 {
			var studentGrades []string
			studentGrades = append(studentGrades, req.Filter.StudentGrades...)
			filter.StudentGrades = studentGrades
		}
		if len(req.Keyword) > 0 {
			filter.ProductName = req.Keyword
		}
	}
	products, err = s.productRepo.GetProductsByFilter(ctx, db, filter)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when getting products by filter with error: %v", err)
		return
	}
	return
}

func (s *ProductService) GetGradeIDsByProductID(ctx context.Context, db database.QueryExecer, productID string) (gradeIDs []string, err error) {
	gradeIDs, err = s.productGradeRepo.GetGradeIDsByProductID(ctx, db, productID)
	if err != nil {
		err = fmt.Errorf("err while getting grade ids by product id with err: %v", err)
		return
	}
	return
}

func (s *ProductService) GetLocationIDsWithProductID(ctx context.Context, db database.QueryExecer, productID string) (locationIDs []string, err error) {
	locationIDs, err = s.productLocationRepo.GetLocationIDsWithProductID(ctx, db, productID)
	if err != nil {
		err = fmt.Errorf("err while getting location ids by product id with err: %v", err)
	}
	return
}

func (s *ProductService) GetProductTypeByProductID(ctx context.Context, db database.QueryExecer, productID string, productType string) (showingType pb.ProductSpecificType, err error) {
	pkg := entities.Package{}
	material := entities.Material{}
	fee := entities.Fee{}
	switch productType {
	case pb.ProductType_PRODUCT_TYPE_PACKAGE.String():
		pkg, err = s.packageRepo.GetByIDForUpdate(ctx, db, productID)
		if err != nil {
			return
		}
		showingType.ProductType = pb.ProductType_PRODUCT_TYPE_PACKAGE
		showingType.PackageType = pb.PackageType(pb.PackageType_value[pkg.PackageType.String])
		showingType.MaterialType = pb.MaterialType_MATERIAL_TYPE_NONE
		showingType.FeeType = pb.FeeType_FEE_TYPE_NONE
		return
	case pb.ProductType_PRODUCT_TYPE_MATERIAL.String():
		material, err = s.materialRepo.GetByIDForUpdate(ctx, db, productID)
		if err != nil {
			return
		}
		showingType.ProductType = pb.ProductType_PRODUCT_TYPE_MATERIAL
		showingType.MaterialType = pb.MaterialType(pb.MaterialType_value[material.MaterialType.String])
		showingType.PackageType = pb.PackageType_PACKAGE_TYPE_NONE
		showingType.FeeType = pb.FeeType_FEE_TYPE_NONE
		return
	case pb.ProductType_PRODUCT_TYPE_FEE.String():
		fee, err = s.feeRepo.GetFeeByID(ctx, db, productID)
		if err != nil {
			return
		}
		showingType.ProductType = pb.ProductType_PRODUCT_TYPE_FEE
		showingType.FeeType = pb.FeeType(pb.FeeType_value[fee.FeeType.String])
		showingType.PackageType = pb.PackageType_PACKAGE_TYPE_NONE
		showingType.MaterialType = pb.MaterialType_MATERIAL_TYPE_NONE
		return
	}
	return
}

func (s *ProductService) GetGradeNamesByIDs(ctx context.Context, db database.Ext, gradeIDs []string) (gradeNames []string, err error) {
	gradeNames, err = s.gradeRepo.GetGradeNamesByGradeIDs(ctx, db, gradeIDs)
	if err != nil {
		err = fmt.Errorf("failed while get grade names by gradeID with this err: %v", err)
		return
	}
	return
}

func (s *ProductService) GetProductIDsByProductTypeAndOrderID(
	ctx context.Context,
	db database.QueryExecer,
	productType,
	orderID string,
) (productIDs []string, err error) {
	productIDs, err = s.productRepo.GetProductIDsByProductTypeAndOrderID(ctx, db, productType, orderID)
	if err != nil {
		err = fmt.Errorf("failed while get product ids by product type and order id with this err: %v", err)
		return
	}
	return
}

func (s *ProductService) GetProductByID(ctx context.Context, db database.QueryExecer, productID string) (product entities.Product, err error) {
	product, err = s.productRepo.GetByIDForUpdate(ctx, db, productID)
	if err != nil {
		err = status.Errorf(codes.Internal, "error while get product by id: %v", err.Error())
		return
	}
	return
}

func (s *ProductService) GetProductSettingByProductID(ctx context.Context, db database.QueryExecer, productID string) (product entities.ProductSetting, err error) {
	product, err = s.productSettingRepo.GetByID(ctx, db, productID)
	if err != nil {
		err = status.Errorf(codes.Internal, "error while get product setting by product id %v: %v", productID, err.Error())
		return
	}
	return
}

func NewProductService() *ProductService {
	return &ProductService{
		productRepo:         &repositories.ProductRepo{},
		productGradeRepo:    &repositories.ProductGradeRepo{},
		productLocationRepo: &repositories.ProductLocationRepo{},
		gradeRepo:           &repositories.GradeRepo{},
		packageRepo:         &repositories.PackageRepo{},
		materialRepo:        &repositories.MaterialRepo{},
		feeRepo:             &repositories.FeeRepo{},
		productSettingRepo:  &repositories.ProductSettingRepo{},
	}
}
