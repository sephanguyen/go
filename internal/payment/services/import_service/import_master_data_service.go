package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ImportMasterDataService struct {
	pb.UnimplementedImportMasterDataServiceServer
	DB                     database.Ext
	AccountingCategoryRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.AccountingCategory) error
		Update(ctx context.Context, db database.QueryExecer, e *entities.AccountingCategory) error
	}
	TaxRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.Tax) error
		Update(ctx context.Context, db database.QueryExecer, e *entities.Tax) error
	}
	BillingScheduleRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.BillingSchedule) error
		Update(ctx context.Context, db database.QueryExecer, e *entities.BillingSchedule) error
	}
	BillingSchedulePeriodRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.BillingSchedulePeriod) error
		Update(ctx context.Context, db database.QueryExecer, e *entities.BillingSchedulePeriod) error
	}
	DiscountRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.Discount) error
		Update(ctx context.Context, db database.QueryExecer, e *entities.Discount) error
	}
	ProductAccountingCategoryRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, productID pgtype.Text, productAccountingCategories []*entities.ProductAccountingCategory) error
	}
	ProductGradeRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, productID pgtype.Text, productGrades []*entities.ProductGrade) error
	}
	FeeRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.Fee) error
		Update(ctx context.Context, db database.QueryExecer, e *entities.Fee) error
	}
	MaterialRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.Material) error
		Update(ctx context.Context, db database.QueryExecer, e *entities.Material) error
	}
	PackageRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.Package) error
		Update(ctx context.Context, db database.QueryExecer, e *entities.Package) error
	}
	ProductPriceRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.ProductPrice) error
		DeleteByProductID(ctx context.Context, db database.QueryExecer, productID pgtype.Text) error
	}
	BillingRatioRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.BillingRatio) error
		Update(ctx context.Context, db database.QueryExecer, e *entities.BillingRatio) error
	}
	PackageCourseRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, productID pgtype.Text, productCourses []*entities.PackageCourse) error
	}
	ProductLocationRepo interface {
		Replace(ctx context.Context, db database.QueryExecer, productID pgtype.Text, productLocations []*entities.ProductLocation) error
	}
	LeavingReasonRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.LeavingReason) error
		Update(ctx context.Context, db database.QueryExecer, e *entities.LeavingReason) error
	}
	PackageQuantityTypeMappingRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, e *entities.PackageQuantityTypeMapping) error
	}
	ProductSettingRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.ProductSetting) error
		Update(ctx context.Context, db database.QueryExecer, e *entities.ProductSetting) error
		GetByID(ctx context.Context, db database.QueryExecer, productID string) (entities.ProductSetting, error)
	}
	PackageCourseMaterialRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, packageID pgtype.Text, associatedProductsByMaterial []*entities.PackageCourseMaterial) error
	}
	PackageCourseFeeRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, packageID pgtype.Text, associatedProductsByFee []*entities.PackageCourseFee) error
	}
	ProductDiscountRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, productID pgtype.Text, e []*entities.ProductDiscount) error
	}
	NotificationDateRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, notificationDate *entities.NotificationDate) error
	}
}

func (s *ImportMasterDataService) ImportProduct(ctx context.Context, req *pb.ImportProductRequest) (*pb.ImportProductResponse, error) {
	switch req.ProductType {
	case pb.ProductType_PRODUCT_TYPE_PACKAGE:
		errors, err := s.packageModifier(ctx, req.Payload)
		return &pb.ImportProductResponse{
			Errors: errors,
		}, err
	case pb.ProductType_PRODUCT_TYPE_MATERIAL:
		return s.importMaterial(ctx, req.Payload)
	case pb.ProductType_PRODUCT_TYPE_FEE:
		return s.importFee(ctx, req.Payload)
	}
	return &pb.ImportProductResponse{Errors: nil}, status.Error(codes.InvalidArgument, "package type is none")
}

func (s *ImportMasterDataService) ImportProductAssociatedData(ctx context.Context, req *pb.ImportProductAssociatedDataRequest) (*pb.ImportProductAssociatedDataResponse, error) {
	switch req.ProductAssociatedDataType {
	case pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_COURSE:
		errors, err := s.packageCourseModifier(ctx, req.Payload)
		if err != nil {
			return nil, err
		}
		return &pb.ImportProductAssociatedDataResponse{
			Errors: errors,
		}, err
	case pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_LOCATION:
		errors, err := s.importProductLocationModifier(ctx, req.Payload)
		if err != nil {
			return nil, err
		}
		return &pb.ImportProductAssociatedDataResponse{
			Errors: errors,
		}, err
	case pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_GRADE:
		errors, err := s.importProductAssociatedDataGrade(ctx, req.Payload)
		if err != nil {
			return nil, err
		}
		return &pb.ImportProductAssociatedDataResponse{
			Errors: errors,
		}, err
	case pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_ACCOUNTING_CATEGORY:
		errors, err := s.importProductAssociatedDataAccountingCategory(ctx, req.Payload)
		if err != nil {
			return nil, err
		}
		return &pb.ImportProductAssociatedDataResponse{
			Errors: errors,
		}, err
	case pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_DISCOUNT:
		errors, err := s.importProductAssociatedDataDiscount(ctx, req.Payload)
		if err != nil {
			return nil, err
		}
		return &pb.ImportProductAssociatedDataResponse{
			Errors: errors,
		}, err
	}
	return nil, status.Error(codes.InvalidArgument, "invalid product associated data type")
}

func (s *ImportMasterDataService) ImportAssociatedProducts(ctx context.Context, req *pb.ImportAssociatedProductsRequest) (*pb.ImportAssociatedProductsResponse, error) {
	switch req.AssociatedProductsType {
	case pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_MATERIAL:
		errors, err := s.importAssociatedProductsMaterial(ctx, req.Payload)
		if err != nil {
			return nil, err
		}
		return &pb.ImportAssociatedProductsResponse{
			Errors: errors,
		}, err
	case pb.AssociatedProductsType_ASSOCIATED_PRODUCTS_FEE:
		errors, err := s.importAssociatedProductsFee(ctx, req.Payload)
		if err != nil {
			return nil, err
		}
		return &pb.ImportAssociatedProductsResponse{
			Errors: errors,
		}, err
	}
	return nil, status.Error(codes.InvalidArgument, "invalid associated products type")
}

func checkMandatoryColumnAndGetIndex(column []string, positions []int) (bool, int) {
	for _, position := range positions {
		if strings.TrimSpace(column[position]) == "" {
			return false, position
		}
	}
	return true, 0
}

func checkConflictDiscountTag(line []string, posittions []int) (bool, error) {
	if strings.TrimSpace(line[posittions[0]]) != "" && strings.TrimSpace(line[posittions[1]]) != "" {
		return true, fmt.Errorf("only student_tag or parent_tag at the same time")
	}
	return false, nil
}
