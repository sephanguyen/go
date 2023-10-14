package service

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	export_entities "github.com/manabie-com/backend/internal/payment/export_entities"
	"github.com/manabie-com/backend/internal/payment/repositories"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PackageService struct {
	PackageRepo interface {
		GetByIDForUpdate(ctx context.Context, db database.QueryExecer, productID string) (packageData entities.Package, err error)
		GetPackagesForExport(ctx context.Context, db database.QueryExecer) (packages []*entities.Package, err error)
		GetByIDForUniqueProduct(ctx context.Context, db database.QueryExecer, packageID string) (entities.Package, error)
	}
	PackageQuantityTypeMappingRepo interface {
		GetByPackageTypeForUpdate(ctx context.Context, db database.QueryExecer, packageType string) (pb.QuantityType, error)
	}
	PackageCourseRepo interface {
		GetByPackageIDForUpdate(ctx context.Context, db database.QueryExecer, packageID string) ([]entities.PackageCourse, error)
	}
	OrderItemCourseRepo interface {
		MultiCreate(ctx context.Context, db database.QueryExecer, course []entities.OrderItemCourse) error
	}
	ProductRepo interface {
		GetByIDsForExport(ctx context.Context, db database.QueryExecer, entitiesIDs []string) ([]entities.Product, error)
	}

	PackageCourseFeeRepo interface {
		GetToTalAssociatedByCourseIDAndPackageID(ctx context.Context, db database.QueryExecer, packageID string, courseIDs []string) (total int32, err error)
	}

	PackageCourseMaterialRepo interface {
		GetToTalAssociatedByCourseIDAndPackageID(ctx context.Context, db database.QueryExecer, packageID string, courseIDs []string) (total int32, err error)
	}
}

func (s *PackageService) VerifyPackageDataAndUpsertRelateData(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (packageInfo utils.PackageInfo, err error) {
	var orderItemCourse []entities.OrderItemCourse
	packageInfo.Package, err = s.PackageRepo.GetByIDForUpdate(ctx, db, orderItemData.ProductInfo.ProductID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting package have err %v", err.Error())
		return
	}

	packageInfo.QuantityType, err = s.PackageQuantityTypeMappingRepo.GetByPackageTypeForUpdate(ctx, db, packageInfo.Package.PackageType.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting quantity type have err %v", err.Error())
		return
	}
	packageInfo.MapCourseInfo,
		orderItemCourse,
		packageInfo.Quantity,
		err = convertMapCourseAndStudentCourse(orderItemData, packageInfo.QuantityType.String())
	if err != nil {
		return
	}
	err = s.verifyPackageCourse(ctx, db, packageInfo)
	if err != nil {
		return
	}

	err = s.OrderItemCourseRepo.MultiCreate(ctx, db, orderItemCourse)
	if err != nil {
		err = status.Errorf(codes.Internal, "creating order item course with error %v", err.Error())
		return
	}
	return
}

func (s *PackageService) verifyPackageCourse(
	ctx context.Context,
	db database.QueryExecer,
	packageInfo utils.PackageInfo,
) (err error) {
	var packageCourses []entities.PackageCourse
	packageCourses, err = s.PackageCourseRepo.GetByPackageIDForUpdate(ctx, db, packageInfo.Package.PackageID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting package course have err %v", err.Error())
		return
	}
	countPackage := 0
	for _, packageCourse := range packageCourses {
		tmpCourseItem, ok := packageInfo.MapCourseInfo[packageCourse.CourseID.String]
		if !ok {
			if packageCourse.MandatoryFlag.Bool {
				err = status.Errorf(codes.FailedPrecondition, "Missing mandatory course with id %s in bill item", packageCourse.CourseID.String)
				return
			}
			continue
		}

		if packageInfo.QuantityType == pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT {
			if packageCourse.CourseWeight.Status == pgtype.Present && tmpCourseItem.Weight.Value == packageCourse.CourseWeight.Int {
				countPackage++
				continue
			}
			err = status.Errorf(codes.FailedPrecondition,
				"This course with id %v doesn't equal course-weight between database and orderItem %v %v %v",
				packageCourse.CourseID.String,
				packageCourse.CourseWeight.Status,
				packageCourse.CourseWeight.Int,
				tmpCourseItem.Weight.Value)
			return
		}
		if packageCourse.MaxSlotsPerCourse.Status == pgtype.Present && tmpCourseItem.Slot.Value <= packageCourse.MaxSlotsPerCourse.Int {
			countPackage++
			continue
		}
		err = status.Errorf(codes.FailedPrecondition,
			constant.CourseHasSlotGreaterThanMaxSlot,
			packageCourse.CourseID.String)
		return
	}
	if countPackage != len(packageInfo.MapCourseInfo) {
		err = status.Errorf(codes.FailedPrecondition, "Some courses is order item not available in package courses")
		return
	}
	if packageInfo.QuantityType != pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT {
		if packageInfo.Package.MaxSlot.Status != pgtype.Present {
			err = status.Errorf(codes.FailedPrecondition, "package is missing max slot")
			return
		}
		if packageInfo.Package.MaxSlot.Int < packageInfo.Quantity {
			err = status.Errorf(codes.FailedPrecondition, "Package with id %s has slot greater than max slot allowed for the package", packageInfo.Package.PackageID.String)
			return
		}
	}
	return
}

func convertMapCourseAndStudentCourse(
	orderItemData utils.OrderItemData,
	quantityType string,
) (
	mapCourse map[string]*pb.CourseItem,
	orderItemCourse []entities.OrderItemCourse,
	quantity int32,
	err error,
) {
	mapCourse = make(map[string]*pb.CourseItem, len(orderItemData.OrderItem.CourseItems))
	for i, item := range orderItemData.OrderItem.CourseItems {
		orderItemCourseID := uuid.NewString()
		mapCourse[item.CourseId] = orderItemData.OrderItem.CourseItems[i]
		tmpOrderItemCourse := entities.OrderItemCourse{}
		_ = multierr.Combine(
			tmpOrderItemCourse.CourseID.Set(item.CourseId),
			tmpOrderItemCourse.OrderID.Set(orderItemData.Order.OrderID.String),
			tmpOrderItemCourse.OrderItemCourseID.Set(orderItemCourseID),
			tmpOrderItemCourse.CourseName.Set(item.CourseName),
			tmpOrderItemCourse.CourseSlot.Set(nil),
			tmpOrderItemCourse.PackageID.Set(orderItemData.ProductInfo.ProductID.String),
		)
		switch quantityType {
		case pb.QuantityType_QUANTITY_TYPE_SLOT.String():
			if item.Slot == nil {
				err = status.Errorf(codes.FailedPrecondition, constant.CourseItemMissingSlotField)
				return
			}
			_ = tmpOrderItemCourse.CourseSlot.Set(item.Slot.Value)
			quantity += item.Slot.Value
		case pb.QuantityType_QUANTITY_TYPE_SLOT_PER_WEEK.String():
			if item.Slot == nil {
				err = status.Errorf(codes.FailedPrecondition, constant.CourseItemMissingSlotField)
				return
			}
			_ = tmpOrderItemCourse.CourseSlot.Set(item.Slot.Value)
			quantity += item.Slot.Value
		case pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT.String():
			if item.Weight == nil {
				err = status.Errorf(codes.FailedPrecondition, "course item for weight is missing weight field")
				return
			}
			quantity += item.Weight.Value
		}
		orderItemCourse = append(orderItemCourse, tmpOrderItemCourse)
	}
	return
}

func (s *PackageService) GetAllPackagesForExport(ctx context.Context, db database.Ext) (packages []*export_entities.ProductPackageExport, err error) {
	packagesRepo, err := s.PackageRepo.GetPackagesForExport(ctx, db)
	if err != nil {
		err = status.Errorf(
			codes.Internal, "Error when get packages for export: %v", err.Error())
		return
	}
	packageIDs := make([]string, 0, len(packagesRepo))
	for _, packageData := range packagesRepo {
		packageIDs = append(packageIDs, packageData.PackageID.String)
	}

	products, err := s.ProductRepo.GetByIDsForExport(ctx, db, packageIDs)
	if err != nil {
		err = status.Errorf(
			codes.Internal, "Error when get product for export package data: %v", err.Error())
		return
	}
	mapProductAndProductIDs := make(map[string]entities.Product, len(products))
	for _, product := range products {
		mapProductAndProductIDs[product.ProductID.String] = product
	}

	for _, packageData := range packagesRepo {
		product, exist := mapProductAndProductIDs[packageData.PackageID.String]
		if !exist {
			err = status.Errorf(codes.Internal, "Missing product info with id: %s", packageData.PackageID.String)
			return
		}
		packages = append(packages, &export_entities.ProductPackageExport{
			PackageID:            packageData.PackageID.String,
			Name:                 product.Name.String,
			PackageType:          packageData.PackageType.String,
			TaxID:                product.TaxID.String,
			ProductTag:           product.ProductTag.String,
			ProductPartnerID:     product.ProductPartnerID.String,
			AvailableFrom:        product.AvailableFrom.Time,
			AvailableUntil:       product.AvailableUntil.Time,
			MaxSlot:              packageData.MaxSlot.Int,
			CustomBillingPeriod:  product.CustomBillingPeriod.Time,
			BillingScheduleID:    product.BillingScheduleID.String,
			DisableProRatingFlag: product.DisableProRatingFlag.Bool,
			PackageStartDate:     packageData.PackageStartDate.Time,
			PackageEndDate:       packageData.PackageEndDate.Time,
			Remarks:              product.Remarks.String,
			IsArchived:           packageData.IsArchived.Bool,
			IsUnique:             product.IsUnique.Bool,
		})
	}
	return
}

func (s *PackageService) GetTotalAssociatedPackageWithCourseIDAndPackageID(ctx context.Context, db database.Ext, packageID string, courseIDs []string) (total int32, err error) {
	totalPackageCourseFee, err := s.PackageCourseFeeRepo.GetToTalAssociatedByCourseIDAndPackageID(ctx, db, packageID, courseIDs)
	if err != nil {
		err = status.Errorf(
			codes.Internal, "Error when get package course fee: %v", err.Error())
		return
	}
	totalPackageCourseMaterial, err := s.PackageCourseMaterialRepo.GetToTalAssociatedByCourseIDAndPackageID(ctx, db, packageID, courseIDs)
	if err != nil {
		err = status.Errorf(
			codes.Internal, "Error when get package course material: %v", err.Error())
		return
	}
	total = totalPackageCourseFee + totalPackageCourseMaterial
	return
}

func (s *PackageService) GetByIDForUniqueProduct(ctx context.Context, db database.Ext, packageID string) (packageEntities entities.Package, err error) {
	packageEntities, err = s.PackageRepo.GetByIDForUniqueProduct(ctx, db, packageID)
	if err != nil {
		return
	}
	return
}

func (s *PackageService) GetQuantityTypeByID(ctx context.Context, db database.Ext, packageID string) (quantityType pb.QuantityType, err error) {
	var packageEntity entities.Package
	packageEntity, err = s.PackageRepo.GetByIDForUpdate(ctx, db, packageID)
	if err != nil {
		return
	}
	quantityType, err = s.PackageQuantityTypeMappingRepo.GetByPackageTypeForUpdate(ctx, db, packageEntity.PackageType.String)
	return
}

func NewPackageService() *PackageService {
	return &PackageService{
		PackageRepo:                    &repositories.PackageRepo{},
		PackageQuantityTypeMappingRepo: &repositories.PackageQuantityTypeMappingRepo{},
		PackageCourseRepo:              &repositories.PackageCourseRepo{},
		OrderItemCourseRepo:            &repositories.OrderItemCourseRepo{},
		ProductRepo:                    &repositories.ProductRepo{},
		PackageCourseFeeRepo:           &repositories.PackageCourseFeeRepo{},
		PackageCourseMaterialRepo:      &repositories.PackageCourseMaterialRepo{},
	}
}
