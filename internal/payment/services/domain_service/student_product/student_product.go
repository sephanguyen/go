package service

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/repositories"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StudentProductService struct {
	StudentProductRepo interface {
		GetLatestEndDateStudentProductWithProductIDAndStudentID(
			ctx context.Context,
			db database.QueryExecer,
			studentID, productID string,
		) (
			studentProducts []*entities.StudentProduct,
			err error,
		)
		Create(
			ctx context.Context,
			db database.QueryExecer,
			studentProduct entities.StudentProduct,
		) (
			err error,
		)
		GetStudentProductForUpdateByStudentProductID(
			ctx context.Context,
			db database.QueryExecer,
			studentProductID string,
		) (studentProduct entities.StudentProduct, err error)
		Update(
			ctx context.Context,
			db database.QueryExecer,
			studentProduct entities.StudentProduct,
		) (err error)
		UpdateStatusStudentProductAndResetStudentProductLabel(
			ctx context.Context,
			db database.QueryExecer,
			studentProductID string,
			studentProductStatus string,
		) (err error)
		GetStudentProductsByStudentProductLabelForUpdate(
			ctx context.Context,
			db database.QueryExecer,
			studentProductLabels []string,
		) (
			studentProducts []*entities.StudentProduct,
			err error,
		)
		GetUniqueProductsByStudentID(ctx context.Context, db database.QueryExecer, studentID string) ([]*entities.StudentProduct, error)
		GetUniqueProductsByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) ([]*entities.StudentProduct, error)

		GetStudentProductByStudentProductID(
			ctx context.Context,
			db database.QueryExecer,
			studentProductID string,
		) (studentProduct entities.StudentProduct, err error)
		CountStudentProductIDsByStudentIDAndLocationIDs(ctx context.Context, db database.QueryExecer, studentID string, locationIDs []string) (int, error)
		GetStudentProductIDsByRootStudentProductID(ctx context.Context, db database.QueryExecer, rootStudentProductID string) ([]*entities.StudentProduct, error)
		GetByStudentIDAndLocationIDsWithPaging(ctx context.Context, db database.QueryExecer, studentID string, locationIDs []string, offset int64, limit int64) ([]*entities.StudentProduct, error)
		GetByIDs(ctx context.Context, db database.QueryExecer, entitiesIDs []string) ([]entities.StudentProduct, error)
		GetStudentProductAssociatedByStudentProductID(ctx context.Context, db database.QueryExecer, studentProductIDs []string) (result []*entities.StudentProduct, err error)
		UpdateWithVersionNumber(ctx context.Context, db database.QueryExecer, e entities.StudentProduct, versionNumber int32) error
		GetActiveRecurringProductsOfStudentInLocation(ctx context.Context, db database.QueryExecer, studentID string, locationID string, ignoreStudentProductID []string) (studentProducts []entities.StudentProduct, err error)
		GetIgnoreStudentProductIDOfRecurringProductsOfStudentInLocation(ctx context.Context, db database.QueryExecer, studentID string, locationID string) (result []string, err error)
		GetActiveOperationFeeOfStudent(ctx context.Context, db database.QueryExecer, studentID string) (result []entities.StudentProduct, err error)
	}

	BillingSchedulePeriodRepo interface {
		GetLatestPeriodByScheduleIDForUpdate(ctx context.Context, db database.QueryExecer, billingScheduleID string) (entities.BillingSchedulePeriod, error)
		GetPeriodByScheduleIDAndEndTime(ctx context.Context, db database.QueryExecer, billingScheduleID string, endTime time.Time) (entities.BillingSchedulePeriod, error)
		GetAllBillingPeriodsByBillingScheduleID(ctx context.Context, db database.QueryExecer, billingScheduleID string) ([]entities.BillingSchedulePeriod, error)
	}

	StudentAssociatedProductRepo interface {
		Create(ctx context.Context, db database.QueryExecer, studentAssociatedProduct entities.StudentAssociatedProduct) (err error)
		Delete(ctx context.Context, db database.QueryExecer, studentAssociatedProduct entities.StudentAssociatedProduct) (err error)
		GetMapAssociatedProducts(ctx context.Context, db database.QueryExecer, associatedStudentProductID string) (mapProductIDWithStudentProductIDs map[string]string, err error)
		CountAssociatedProductIDsByStudentProductID(ctx context.Context, db database.QueryExecer, entitiesID string) (total int, err error)
		GetAssociatedProductIDsByStudentProductID(ctx context.Context, db database.QueryExecer, entitiesID string, offset int64, limit int64) (associatedProductIDs []string, err error)
	}

	ProductRepo interface {
		GetByID(ctx context.Context, db database.QueryExecer, entitiesID string) (entities.Product, error)
	}

	PackageRepo interface {
		GetByIDForUpdate(ctx context.Context, db database.QueryExecer, productID string) (packageData entities.Package, err error)
	}

	ProductSettingRepo interface {
		GetByID(ctx context.Context, db database.QueryExecer, productID string) (entities.ProductSetting, error)
	}

	ProductLocationRepo interface {
		GetLocationIDsWithProductID(ctx context.Context, db database.QueryExecer, productID string) (locationIDs []string, err error)
	}

	StudentEnrollmentStatusHistoryRepo interface {
		GetLatestStatusEnrollmentByStudentIDAndLocationIDs(
			ctx context.Context,
			db database.QueryExecer,
			studentID string,
			locationIDs []string,
		) (
			[]*entities.StudentEnrollmentStatusHistory,
			error,
		)
	}
}

func (s *StudentProductService) GetStudentProductsByStudentProductIDs(ctx context.Context, db database.Ext, studentProductIDs []string) (studentProducts []entities.StudentProduct, err error) {
	studentProducts, err = s.StudentProductRepo.GetByIDs(ctx, db, studentProductIDs)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when get student product by student product ids: %v", err.Error())
	}
	return
}

func (s *StudentProductService) CreateStudentProduct(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (
	studentProduct entities.StudentProduct,
	err error,
) {
	isAssociated := false
	if orderItemData.OrderItem.PackageAssociatedId != nil || orderItemData.OrderItem.AssociatedStudentProductId != nil {
		isAssociated = true
	}

	now := time.Now()
	err = utils.GroupErrorFunc(
		multierr.Combine(
			studentProduct.ProductID.Set(orderItemData.ProductInfo.ProductID.String),
			studentProduct.StudentID.Set(orderItemData.StudentInfo.StudentID.String),
			studentProduct.ProductStatus.Set(pb.StudentProductStatus_ORDERED.String()),
			studentProduct.ApprovalStatus.Set(nil),
			studentProduct.LocationID.Set(orderItemData.Order.LocationID.String),
			studentProduct.UpcomingBillingDate.Set(nil),
			studentProduct.UpdatedToStudentProductID.Set(nil),
			studentProduct.UpdatedFromStudentProductID.Set(nil),
			studentProduct.StartDate.Set(nil),
			studentProduct.EndDate.Set(nil),
			studentProduct.DeletedAt.Set(nil),
			studentProduct.UpdatedAt.Set(now),
			studentProduct.CreatedAt.Set(now),
			studentProduct.StudentProductID.Set(idutil.ULIDNow()),
			studentProduct.StudentProductLabel.Set(pb.StudentProductLabel_CREATED.String()),
			studentProduct.IsUnique.Set(orderItemData.ProductInfo.IsUnique.Bool),
			studentProduct.RootStudentProductID.Set(nil),
			studentProduct.IsAssociated.Set(isAssociated),
			studentProduct.VersionNumber.Set(0),
		),
		s.checkStartTimeAndEndTimeForStudentProduct(ctx, db, orderItemData, &studentProduct),
		s.checkUniqueStudentProduct(ctx, db, orderItemData, &studentProduct),
		s.StudentProductRepo.Create(ctx, db, studentProduct),
	)
	return
}

func (s *StudentProductService) checkStartTimeAndEndTimeForStudentProduct(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
	studentProduct *entities.StudentProduct,
) (err error) {
	var latestPeriod entities.BillingSchedulePeriod
	var packageInfo entities.Package
	if orderItemData.IsOneTimeProduct {
		if orderItemData.ProductType != pb.ProductType_PRODUCT_TYPE_PACKAGE {
			return
		}
		packageInfo, err = s.PackageRepo.GetByIDForUpdate(ctx, db, orderItemData.OrderItem.ProductId)
		if err != nil {
			err = status.Errorf(codes.Internal, "getting package have error %v", err.Error())
			return
		}
		_ = studentProduct.StartDate.Set(utils.StartOfDate(packageInfo.PackageStartDate.Time, orderItemData.Timezone))
		_ = studentProduct.EndDate.Set(utils.EndOfDate(packageInfo.PackageEndDate.Time, orderItemData.Timezone))
		return
	}
	_ = studentProduct.StartDate.Set(utils.StartOfDate(orderItemData.OrderItem.StartDate.AsTime(), orderItemData.Timezone))

	latestPeriod, err = s.BillingSchedulePeriodRepo.GetLatestPeriodByScheduleIDForUpdate(ctx, db, orderItemData.ProductInfo.BillingScheduleID.String)
	if err != nil {
		return
	}
	_ = studentProduct.EndDate.Set(utils.EndOfDate(latestPeriod.EndDate.Time, orderItemData.Timezone))

	return
}

func (s *StudentProductService) checkUniqueStudentProduct(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
	studentProduct *entities.StudentProduct,
) (err error) {
	var (
		studentProducts      []*entities.StudentProduct
		latestStudentProduct *entities.StudentProduct
	)
	if !orderItemData.ProductInfo.IsUnique.Bool {
		return
	}
	studentProducts, err = s.StudentProductRepo.GetLatestEndDateStudentProductWithProductIDAndStudentID(ctx, db, orderItemData.Order.StudentID.String, orderItemData.ProductInfo.ProductID.String)
	if err != nil {
		err = status.Errorf(codes.InvalidArgument, fmt.Sprintf("error when get latest end date student product with product id and student id: %v", err))
		return
	}
	if len(studentProducts) == 0 {
		return
	}

	latestStudentProduct = studentProducts[0]
	if orderItemData.IsOneTimeProduct {
		if orderItemData.ProductInfo.ProductType.String == pb.ProductType_PRODUCT_TYPE_PACKAGE.String() && latestStudentProduct.ProductStatus.String == pb.StudentProductStatus_CANCELLED.String() {
			return
		}
		err = status.Errorf(codes.InvalidArgument, "creating one time student product have error because it is unique product and it already have active student product")
		return
	}

	if !studentProduct.StartDate.Time.After(latestStudentProduct.EndDate.Time) {
		err = status.Errorf(codes.InvalidArgument, "creating return student product have error because it is unique product and it have conflict time range with previous student product")
	}

	return
}

func (s *StudentProductService) CreateAssociatedStudentProduct(
	ctx context.Context,
	db database.QueryExecer,
	associatedProducts []*pb.ProductAssociation,
	mapKeyWithOrderItemData map[string]utils.OrderItemData,
) (
	err error,
) {
	if len(associatedProducts) == 0 {
		return
	}
	for _, associatedProduct := range associatedProducts {
		var studentAssociatedProduct entities.StudentAssociatedProduct
		associatedKey, packageKey := utils.GetKeyFromStudentAssociatedProduct(associatedProduct)
		associatedOrderItemData, ok := mapKeyWithOrderItemData[associatedKey]
		if !ok {
			err = status.Errorf(codes.FailedPrecondition, "getting order item data from associated product")
			return
		}

		packageOrderItemData, ok := mapKeyWithOrderItemData[packageKey]
		if !ok {
			err = status.Errorf(codes.FailedPrecondition, "getting order item data from package")
			return
		}
		var mapProductIDWithAssociatedProductIDs map[string]string
		mapProductIDWithAssociatedProductIDs, err = s.StudentAssociatedProductRepo.GetMapAssociatedProducts(ctx, db, packageOrderItemData.StudentProduct.StudentProductID.String)
		if err != nil {
			return
		}
		if _, ok := mapProductIDWithAssociatedProductIDs[associatedOrderItemData.StudentProduct.ProductID.String]; ok {
			studentProductCheck, errGetStudentProduct := s.StudentProductRepo.GetStudentProductForUpdateByStudentProductID(ctx, db, packageOrderItemData.StudentProduct.StudentProductID.String)
			if errGetStudentProduct != nil {
				err = status.Errorf(codes.Internal, "fail to get student product: %v with err: %v", packageOrderItemData.StudentProduct.StudentProductID.String, errGetStudentProduct.Error())
				return
			}
			if !s.associatedProductIsCanceled(studentProductCheck) {
				err = utils.StatusErrWithDetail(
					codes.Internal,
					constant.DuplicatedAssociate,
					&errdetails.DebugInfo{Detail: fmt.Sprintf("product %v is already associated with %v", associatedOrderItemData.StudentProduct.StudentProductID.String, packageOrderItemData.StudentProduct.StudentProductID.String)},
				)
				return
			}
		}
		err = multierr.Combine(
			studentAssociatedProduct.AssociatedProductID.Set(associatedOrderItemData.StudentProduct.StudentProductID.String),
			studentAssociatedProduct.StudentProductID.Set(packageOrderItemData.StudentProduct.StudentProductID.String),
			studentAssociatedProduct.CreatedAt.Set(time.Now()),
			studentAssociatedProduct.UpdatedAt.Set(time.Now()),
			studentAssociatedProduct.DeletedAt.Set(nil),
		)
		if err != nil {
			return
		}

		err = s.StudentAssociatedProductRepo.Create(ctx, db, studentAssociatedProduct)
		if err != nil {
			err = status.Errorf(codes.Internal, "creating student associated product have error %v", err.Error())
			return
		}
	}
	return
}

func (s *StudentProductService) MutationStudentProductForUpdateOrder(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (
	studentProduct entities.StudentProduct,
	rootStudentProduct entities.StudentProduct,
	err error,
) {
	isAssociated := false
	if orderItemData.OrderItem.PackageAssociatedId != nil {
		isAssociated = true
	}

	if orderItemData.OrderItem.StudentProductId == nil {
		err = status.Errorf(codes.FailedPrecondition, "updating student product without student product id")
		return
	}

	studentProduct, err = s.StudentProductRepo.GetStudentProductForUpdateByStudentProductID(ctx, db, orderItemData.OrderItem.StudentProductId.Value)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting student product for update have error %v", err.Error())
		return
	}

	err = utils.CheckOutVersion(studentProduct.VersionNumber.Int, orderItemData.OrderItem.StudentProductVersionNumber)
	if err != nil {
		return
	}

	if orderItemData.IsOneTimeProduct {
		_ = studentProduct.StudentProductLabel.Set(pb.StudentProductLabel_UPDATED.String())
		err = s.StudentProductRepo.UpdateWithVersionNumber(
			ctx,
			db,
			studentProduct,
			orderItemData.OrderItem.StudentProductVersionNumber,
		)
		if err != nil {
			err = status.Errorf(codes.Internal, "updating student product label and status have error %v", err.Error())
		}
		return
	}

	err = checkEffectiveDateForUpdateOrder(orderItemData.OrderItem, studentProduct, orderItemData.Timezone)
	if err != nil {
		return
	}

	switch studentProduct.StudentProductLabel.String {
	case pb.StudentProductLabel_WITHDRAWAL_SCHEDULED.String(),
		pb.StudentProductLabel_GRADUATION_SCHEDULED.String(),
		pb.StudentProductLabel_PAUSED.String():
		err = status.Errorf(codes.FailedPrecondition, constant.UnableToUpdateProductDueToPendingOrder)
	case pb.StudentProductLabel_UPDATE_SCHEDULED.String():
		err = s.RevertUpdateStudentProduct(ctx, db, &studentProduct, orderItemData.ProductInfo.BillingScheduleID.String)
	}
	if err != nil {
		return
	}
	newStudentProductID := idutil.ULIDNow()
	newStudentProduct := studentProduct
	rootStudentProductID := studentProduct.StudentProductID.String
	endDateOfStudentProduct := utils.EndOfDate(studentProduct.EndDate.Time, orderItemData.Timezone)
	now := time.Now()
	endDateOfOldStudentProduct := orderItemData.OrderItem.EffectiveDate.AsTime()
	if endDateOfOldStudentProduct.Truncate(24 * time.Hour).After(studentProduct.StartDate.Time.Truncate(24 * time.Hour)) {
		endDateOfOldStudentProduct = utils.EndOfDate(endDateOfOldStudentProduct.AddDate(0, 0, -1), orderItemData.Timezone)
	}
	if studentProduct.RootStudentProductID.Status == pgtype.Present {
		rootStudentProductID = studentProduct.RootStudentProductID.String
	}

	if orderItemData.OrderItem.CancellationDate != nil {
		err = studentProduct.ProductStatus.Set(pb.StudentProductStatus_CANCELLED.String())
		if err != nil {
			return
		}
	} else {
		err = studentProduct.ProductStatus.Set(pb.StudentProductStatus_ORDERED.String())
		if err != nil {
			return
		}
	}

	err = multierr.Combine(
		studentProduct.StudentProductLabel.Set(pb.StudentProductLabel_UPDATE_SCHEDULED.String()),
		studentProduct.EndDate.Set(endDateOfOldStudentProduct),
		studentProduct.UpdatedToStudentProductID.Set(newStudentProductID),

		newStudentProduct.UpdatedFromStudentProductID.Set(studentProduct.StudentProductID.String),
		newStudentProduct.StartDate.Set(utils.StartOfDate(orderItemData.OrderItem.EffectiveDate.AsTime(), orderItemData.Timezone)),
		newStudentProduct.EndDate.Set(endDateOfStudentProduct),
		newStudentProduct.UpdatedAt.Set(now),
		newStudentProduct.CreatedAt.Set(now),
		newStudentProduct.StudentProductID.Set(newStudentProductID),
		newStudentProduct.StudentProductLabel.Set(pb.StudentProductLabel_CREATED.String()),
		newStudentProduct.RootStudentProductID.Set(rootStudentProductID),
		newStudentProduct.IsAssociated.Set(isAssociated),
		newStudentProduct.VersionNumber.Set(0),
	)

	if err != nil {
		err = status.Errorf(codes.FailedPrecondition, "converting student product with error: %v", err.Error())
		return
	}

	err = utils.GroupErrorFunc(
		s.StudentProductRepo.Create(ctx, db, newStudentProduct),
		s.StudentProductRepo.UpdateWithVersionNumber(ctx, db, studentProduct, orderItemData.OrderItem.StudentProductVersionNumber),
	)

	if err != nil {
		err = status.Errorf(codes.FailedPrecondition, "updating student product with error: %v", err.Error())
	}

	return newStudentProduct, studentProduct, err
}

func (s *StudentProductService) MutationStudentProductForCancelOrder(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (
	studentProduct entities.StudentProduct,
	rootStudentProduct entities.StudentProduct,
	err error,
) {
	isAssociated := false
	if orderItemData.OrderItem.PackageAssociatedId != nil || orderItemData.OrderItem.AssociatedStudentProductId != nil {
		isAssociated = true
	}

	if orderItemData.OrderItem.StudentProductId == nil {
		err = status.Errorf(codes.FailedPrecondition, "cancelling student product without student product id")
		return
	}
	studentProduct, err = s.StudentProductRepo.GetStudentProductForUpdateByStudentProductID(ctx, db, orderItemData.OrderItem.StudentProductId.Value)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting student product for update have error %v", err.Error())
		return
	}

	err = utils.CheckOutVersion(studentProduct.VersionNumber.Int, orderItemData.OrderItem.StudentProductVersionNumber)
	if err != nil {
		return
	}

	if orderItemData.IsOneTimeProduct {
		_ = studentProduct.StudentProductLabel.Set(pb.StudentProductLabel_UPDATED.String())
		_ = studentProduct.ProductStatus.Set(pb.StudentProductStatus_CANCELLED.String())
		if orderItemData.ProductType == pb.ProductType_PRODUCT_TYPE_PACKAGE {
			_ = studentProduct.EndDate.Set(utils.EndOfDate(time.Now(), orderItemData.Timezone))
		}
		err = s.StudentProductRepo.UpdateWithVersionNumber(
			ctx,
			db,
			studentProduct,
			orderItemData.OrderItem.StudentProductVersionNumber,
		)
		if err != nil {
			err = status.Errorf(codes.Internal, "updating student product label and status have error %v", err.Error())
		}

		return
	}

	err = checkEffectiveDateForUpdateOrder(orderItemData.OrderItem, studentProduct, orderItemData.Timezone)
	if err != nil {
		return
	}

	switch studentProduct.StudentProductLabel.String {
	case pb.StudentProductLabel_WITHDRAWAL_SCHEDULED.String():
		err = status.Errorf(codes.FailedPrecondition, constant.UnableToUpdateProductDueToPendingOrder)
	case pb.StudentProductLabel_GRADUATION_SCHEDULED.String():
		err = status.Errorf(codes.FailedPrecondition, constant.UnableToUpdateProductDueToPendingOrder)
	case pb.StudentProductLabel_UPDATE_SCHEDULED.String():
		err = s.RevertUpdateStudentProduct(ctx, db, &studentProduct, orderItemData.ProductInfo.BillingScheduleID.String)
	}
	if err != nil {
		return
	}

	if orderItemData.OrderItem.EffectiveDate.AsTime().Truncate(24 * time.Hour).Equal(studentProduct.StartDate.Time.Truncate(24 * time.Hour)) {
		err = multierr.Combine(
			studentProduct.ProductStatus.Set(pb.StudentProductStatus_CANCELLED.String()),
			studentProduct.EndDate.Set(utils.EndOfDate(orderItemData.OrderItem.EffectiveDate.AsTime(), orderItemData.Timezone)),
			studentProduct.StudentProductLabel.Set(pb.StudentProductLabel_UPDATED.String()),
		)
		if err != nil {
			err = status.Errorf(codes.FailedPrecondition, "Fail to convert student product: %v", err.Error())
			return
		}
		err = s.StudentProductRepo.UpdateWithVersionNumber(ctx, db, studentProduct, orderItemData.OrderItem.StudentProductVersionNumber)
		if err != nil {
			return
		}
	} else {
		newStudentProductID := idutil.ULIDNow()
		now := time.Now()
		endDateOfOldStudentProduct := utils.EndOfDate(orderItemData.OrderItem.EffectiveDate.AsTime().AddDate(0, 0, -1), orderItemData.Timezone)
		studentProductNeedCancel := studentProduct
		rootStudentProductID := studentProduct.StudentProductID.String
		if studentProduct.RootStudentProductID.Status == pgtype.Present {
			rootStudentProductID = studentProduct.RootStudentProductID.String
		}
		err = multierr.Combine(
			studentProduct.StudentProductLabel.Set(pb.StudentProductLabel_UPDATE_SCHEDULED.String()),
			studentProduct.EndDate.Set(endDateOfOldStudentProduct),
			studentProduct.UpdatedToStudentProductID.Set(newStudentProductID),

			studentProductNeedCancel.UpdatedFromStudentProductID.Set(studentProduct.StudentProductID.String),
			studentProductNeedCancel.StartDate.Set(utils.StartOfDate(orderItemData.OrderItem.EffectiveDate.AsTime(), orderItemData.Timezone)),
			studentProductNeedCancel.EndDate.Set(utils.EndOfDate(orderItemData.OrderItem.EffectiveDate.AsTime(), orderItemData.Timezone)),
			studentProductNeedCancel.UpdatedAt.Set(now),
			studentProductNeedCancel.CreatedAt.Set(now),
			studentProductNeedCancel.StudentProductID.Set(newStudentProductID),
			studentProductNeedCancel.ProductStatus.Set(pb.StudentProductStatus_CANCELLED.String()),
			studentProductNeedCancel.StudentProductLabel.Set(pb.StudentProductLabel_CREATED.String()),
			studentProductNeedCancel.RootStudentProductID.Set(rootStudentProductID),
			studentProductNeedCancel.IsAssociated.Set(isAssociated),
			studentProductNeedCancel.VersionNumber.Set(0),
		)

		if err != nil {
			err = status.Errorf(codes.FailedPrecondition, "converting student product with error: %v", err.Error())
			return
		}

		err = utils.GroupErrorFunc(
			s.StudentProductRepo.Create(ctx, db, studentProductNeedCancel),
			s.StudentProductRepo.UpdateWithVersionNumber(ctx, db, studentProduct, orderItemData.OrderItem.StudentProductVersionNumber),
		)

		if err != nil {
			err = status.Errorf(codes.FailedPrecondition, "updating student product with error: %v", err.Error())
		}
		return studentProductNeedCancel, studentProduct, err
	}

	return
}

func (s *StudentProductService) MutationStudentProductForWithdrawalOrder(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (
	studentProduct entities.StudentProduct,
	err error,
) {
	if orderItemData.IsOneTimeProduct {
		err = status.Errorf(codes.Internal, "updating student product label and status for withdraw order is unimplemented")
		return
	}

	studentProduct, err = s.StudentProductRepo.GetStudentProductForUpdateByStudentProductID(ctx, db, orderItemData.OrderItem.StudentProductId.Value)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting student product for withdrawal student product with err %v", err.Error())
		return
	}

	err = utils.CheckOutVersion(studentProduct.VersionNumber.Int, orderItemData.OrderItem.StudentProductVersionNumber)
	if err != nil {
		return
	}

	err = checkEffectiveDateForUpdateOrder(orderItemData.OrderItem, studentProduct, orderItemData.Timezone)
	if err != nil {
		return
	}

	_ = studentProduct.StudentProductLabel.Set(pb.StudentProductLabel_WITHDRAWAL_SCHEDULED.String())
	_ = studentProduct.EndDate.Set(utils.EndOfDate(orderItemData.OrderItem.EffectiveDate.AsTime().UTC(), orderItemData.Timezone).AddDate(0, 0, 1))
	_ = studentProduct.UpdatedAt.Set(time.Now())

	err = s.StudentProductRepo.UpdateWithVersionNumber(ctx, db, studentProduct, orderItemData.OrderItem.StudentProductVersionNumber)
	if err != nil {
		err = status.Errorf(codes.Internal, "updating student product for withdrawal have error %v", err.Error())
	}
	return
}

func (s *StudentProductService) MutationStudentProductForGraduateOrder(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (
	studentProduct entities.StudentProduct,
	err error,
) {
	if orderItemData.IsOneTimeProduct {
		err = status.Errorf(codes.Internal, "updating student product label and status for graduate order is unimplemented")
		return
	}

	studentProduct, err = s.StudentProductRepo.GetStudentProductForUpdateByStudentProductID(ctx, db, orderItemData.OrderItem.StudentProductId.Value)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting student product for graduate student product with err %v", err.Error())
		return
	}

	err = utils.CheckOutVersion(studentProduct.VersionNumber.Int, orderItemData.OrderItem.StudentProductVersionNumber)
	if err != nil {
		return
	}

	err = checkEffectiveDateForUpdateOrder(orderItemData.OrderItem, studentProduct, orderItemData.Timezone)
	if err != nil {
		return
	}

	_ = studentProduct.StudentProductLabel.Set(pb.StudentProductLabel_GRADUATION_SCHEDULED.String())
	_ = studentProduct.EndDate.Set(utils.EndOfDate(orderItemData.OrderItem.EffectiveDate.AsTime().UTC(), orderItemData.Timezone).AddDate(0, 0, 1))
	_ = studentProduct.UpdatedAt.Set(time.Now())

	err = s.StudentProductRepo.UpdateWithVersionNumber(ctx, db, studentProduct, orderItemData.OrderItem.StudentProductVersionNumber)
	if err != nil {
		err = status.Errorf(codes.Internal, "updating student product for graduate have error %v", err.Error())
	}
	return
}

func (s *StudentProductService) MutationStudentProductForLOAOrder(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (
	studentProduct entities.StudentProduct,
	err error,
) {
	if orderItemData.IsOneTimeProduct {
		err = status.Errorf(codes.Internal, "updating student product label and status for LOA order is unimplemented")
		return
	}

	studentProduct, err = s.StudentProductRepo.GetStudentProductForUpdateByStudentProductID(ctx, db, orderItemData.OrderItem.StudentProductId.Value)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting student product for LOA student product with err %v", err.Error())
		return
	}
	if studentProduct.StudentProductLabel.String == pb.StudentProductLabel_UPDATE_SCHEDULED.String() {
		err = status.Errorf(codes.FailedPrecondition, constant.UnableToUpdateProductDueToPendingOrder)
		return
	}
	err = utils.CheckOutVersion(studentProduct.VersionNumber.Int, orderItemData.OrderItem.StudentProductVersionNumber)
	if err != nil {
		return
	}

	err = checkLOADuration(orderItemData.OrderItem, studentProduct, orderItemData.Timezone)
	if err != nil {
		return
	}

	_ = studentProduct.StudentProductLabel.Set(pb.StudentProductLabel_PAUSE_SCHEDULED.String())
	_ = studentProduct.EndDate.Set(utils.EndOfDate(orderItemData.OrderItem.EffectiveDate.AsTime().UTC(), orderItemData.Timezone).AddDate(0, 0, 1))
	_ = studentProduct.UpdatedAt.Set(time.Now())

	err = s.StudentProductRepo.UpdateWithVersionNumber(ctx, db, studentProduct, orderItemData.OrderItem.StudentProductVersionNumber)
	if err != nil {
		err = status.Errorf(codes.Internal, "updating student product for LOA have error %v", err.Error())
	}
	return
}

func (s *StudentProductService) MutationStudentProductForResumeOrder(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (
	studentProduct entities.StudentProduct,
	err error,
) {
	if orderItemData.IsOneTimeProduct {
		err = status.Errorf(codes.Internal, "updating student product label and status for Resume order is unimplemented")
		return
	}
	isAssociated := false
	if orderItemData.OrderItem.PackageAssociatedId != nil {
		isAssociated = true
	}

	pausedStudentProduct, err := s.StudentProductRepo.GetStudentProductForUpdateByStudentProductID(ctx, db, orderItemData.OrderItem.StudentProductId.Value)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting student product for LOA student product with err %v", err.Error())
		return
	}

	err = utils.CheckOutVersion(pausedStudentProduct.VersionNumber.Int, orderItemData.OrderItem.StudentProductVersionNumber)
	if err != nil {
		return
	}

	now := time.Now()
	newStudentProductID := idutil.ULIDNow()

	rootStudentProductID := pausedStudentProduct.StudentProductID.String
	if pausedStudentProduct.RootStudentProductID.Status == pgtype.Present {
		rootStudentProductID = pausedStudentProduct.RootStudentProductID.String
	}

	err = utils.GroupErrorFunc(
		multierr.Combine(
			pausedStudentProduct.UpdatedAt.Set(now),
			pausedStudentProduct.UpdatedToStudentProductID.Set(newStudentProductID),

			studentProduct.StudentProductID.Set(newStudentProductID),
			studentProduct.ProductID.Set(orderItemData.ProductInfo.ProductID.String),
			studentProduct.StudentID.Set(orderItemData.StudentInfo.StudentID.String),
			studentProduct.ProductStatus.Set(pb.StudentProductStatus_ORDERED.String()),
			studentProduct.ApprovalStatus.Set(nil),
			studentProduct.LocationID.Set(orderItemData.Order.LocationID.String),
			studentProduct.UpcomingBillingDate.Set(nil),
			studentProduct.UpdatedToStudentProductID.Set(nil),
			studentProduct.UpdatedFromStudentProductID.Set(pausedStudentProduct.StudentProductID),
			studentProduct.StartDate.Set(nil),
			studentProduct.EndDate.Set(nil),
			studentProduct.DeletedAt.Set(nil),
			studentProduct.UpdatedAt.Set(now),
			studentProduct.CreatedAt.Set(now),
			studentProduct.StudentProductLabel.Set(pb.StudentProductLabel_CREATED.String()),
			studentProduct.IsUnique.Set(orderItemData.ProductInfo.IsUnique.Bool),
			studentProduct.RootStudentProductID.Set(rootStudentProductID),
			studentProduct.IsAssociated.Set(isAssociated),
			studentProduct.VersionNumber.Set(0),
		),

		s.checkStartTimeAndEndTimeForStudentProduct(ctx, db, orderItemData, &studentProduct),
		s.StudentProductRepo.Create(ctx, db, studentProduct),
		s.StudentProductRepo.UpdateWithVersionNumber(ctx, db, pausedStudentProduct, orderItemData.OrderItem.StudentProductVersionNumber),
	)
	return
}

func (s *StudentProductService) VoidStudentProduct(
	ctx context.Context,
	db database.QueryExecer,
	studentProductID string,
	orderType string,
) (
	studentProduct entities.StudentProduct,
	product entities.Product,
	isCancel bool,
	err error,
) {
	var isOneTimeProduct bool
	studentProduct, err = s.StudentProductRepo.GetStudentProductForUpdateByStudentProductID(ctx, db, studentProductID)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting student product for update have error: %v", err)
		return
	}

	product, err = s.ProductRepo.GetByID(ctx, db, studentProduct.ProductID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting product for update have error: %v", err)
		return
	}

	if product.BillingScheduleID.Status == pgtype.Null {
		isOneTimeProduct = true
	}

	err = checkStudentProductLabelByOrderTypeForVoidOrder(orderType, studentProduct.StudentProductLabel.String)
	if err != nil {
		return
	}

	switch orderType {
	case pb.OrderType_ORDER_TYPE_WITHDRAWAL.String(),
		pb.OrderType_ORDER_TYPE_GRADUATE.String(),
		pb.OrderType_ORDER_TYPE_LOA.String():
		if isOneTimeProduct {
			err = status.Errorf(codes.Internal, "can't void withdraw,graduate,loa order with one time product")
			return
		}
		studentProduct, err = s.voidStudentProductForWithdrawal(ctx, db, studentProduct, product)
		if err != nil {
			return
		}
	case pb.OrderType_ORDER_TYPE_UPDATE.String():
		if isOneTimeProduct {
			err = status.Errorf(codes.Internal, "can't void update order with one time product")
			return
		}
		if studentProduct.ProductStatus.String == pb.StudentProductStatus_CANCELLED.String() {
			isCancel = true
			studentProduct, err = s.voidStudentProductForCancel(ctx, db, product, studentProduct)
			if err != nil {
				return
			}
			return
		}
		studentProduct, err = s.voidStudentProductForUpdate(ctx, db, studentProduct)
		if err != nil {
			return
		}
	case pb.OrderType_ORDER_TYPE_NEW.String(),
		pb.OrderType_ORDER_TYPE_RESUME.String(),
		pb.OrderType_ORDER_TYPE_ENROLLMENT.String():
		studentProduct, err = s.voidStudentProductForCreate(ctx, db, studentProduct)
		if err != nil {
			return
		}
	default:
		err = status.Errorf(codes.Internal, fmt.Sprintf("voiding %s order is unimplemented", orderType))
	}
	return
}

func (s *StudentProductService) voidStudentProductForWithdrawal(
	ctx context.Context,
	db database.QueryExecer,
	studentProduct entities.StudentProduct,
	product entities.Product,
) (studentProductReturn entities.StudentProduct, err error) {
	var period entities.BillingSchedulePeriod
	period, err = s.BillingSchedulePeriodRepo.GetLatestPeriodByScheduleIDForUpdate(ctx, db, product.BillingScheduleID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "can't get latest period from product id: %v", product.ProductID.String)
		return
	}
	_ = multierr.Combine(
		studentProduct.ProductStatus.Set(pb.StudentProductStatus_ORDERED.String()),
		studentProduct.StudentProductLabel.Set(pb.StudentProductLabel_CREATED.String()),
		studentProduct.EndDate.Set(period.EndDate.Time),
	)
	err = s.StudentProductRepo.Update(ctx, db, studentProduct)
	if err != nil {
		err = status.Errorf(codes.Internal, "updating student product have error: %v", err)
		return
	}
	studentProductReturn = studentProduct
	return
}

func (s *StudentProductService) voidStudentProductForCreate(
	ctx context.Context,
	db database.QueryExecer,
	studentProduct entities.StudentProduct,
) (studentProductReturn entities.StudentProduct, err error) {
	_ = multierr.Combine(
		studentProduct.ProductStatus.Set(pb.StudentProductStatus_CANCELLED.String()),
		studentProduct.StartDate.Set(nil),
		studentProduct.EndDate.Set(nil),
	)
	err = s.StudentProductRepo.Update(ctx, db, studentProduct)
	if err != nil {
		err = status.Errorf(codes.Internal, "updating student product have error: %v", err)
		return
	}
	studentProductReturn = studentProduct
	return
}

func (s *StudentProductService) voidStudentProductForUpdate(
	ctx context.Context,
	db database.QueryExecer,
	studentProduct entities.StudentProduct,
) (studentProductReturn entities.StudentProduct, err error) {
	endDateOriginal := studentProduct.EndDate.Time
	prevStudentProductID := studentProduct.UpdatedFromStudentProductID.String
	_ = multierr.Combine(
		studentProduct.ProductStatus.Set(pb.StudentProductStatus_CANCELLED.String()),
		studentProduct.StartDate.Set(nil),
		studentProduct.EndDate.Set(nil),
	)
	err = s.StudentProductRepo.Update(ctx, db, studentProduct)
	if err != nil {
		err = status.Errorf(codes.Internal, "updating student product have error: %v", err)
		return
	}
	prevStudentProduct, err := s.StudentProductRepo.GetStudentProductByStudentProductID(ctx, db, prevStudentProductID)
	if err != nil {
		err = status.Errorf(codes.Internal, "can't get prev student product with error: %v", err)
		return
	}
	_ = prevStudentProduct.EndDate.Set(endDateOriginal)
	_ = prevStudentProduct.ProductStatus.Set(pb.StudentProductStatus_ORDERED.String())
	_ = prevStudentProduct.StudentProductLabel.Set(pb.StudentProductLabel_CREATED.String())
	err = s.StudentProductRepo.Update(ctx, db, prevStudentProduct)
	if err != nil {
		err = status.Errorf(codes.Internal, "updating prev student product have error: %v", err)
		return
	}
	studentProductReturn = prevStudentProduct
	return
}

func (s *StudentProductService) voidStudentProductForCancel(
	ctx context.Context,
	db database.QueryExecer,
	product entities.Product,
	studentProduct entities.StudentProduct,
) (studentProductReturn entities.StudentProduct, err error) {
	var (
		lastPeriod           entities.BillingSchedulePeriod
		prevStudentProductID string
	)
	lastPeriod, err = s.BillingSchedulePeriodRepo.GetLatestPeriodByScheduleIDForUpdate(ctx, db, product.BillingScheduleID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "can't get latest period from product id: %v", product.ProductID.String)
		return
	}
	if studentProduct.UpdatedFromStudentProductID.Status == pgtype.Null {
		_ = multierr.Combine(
			studentProduct.ProductStatus.Set(pb.StudentProductStatus_ORDERED.String()),
			studentProduct.StudentProductLabel.Set(pb.StudentProductLabel_CREATED.String()),
			studentProduct.EndDate.Set(lastPeriod.EndDate.Time),
		)
	} else {
		prevStudentProductID = studentProduct.UpdatedFromStudentProductID.String
		_ = multierr.Combine(
			studentProduct.ProductStatus.Set(pb.StudentProductStatus_CANCELLED.String()),
			studentProduct.StartDate.Set(nil),
			studentProduct.EndDate.Set(nil),
		)
	}
	err = s.StudentProductRepo.Update(ctx, db, studentProduct)
	if err != nil {
		err = status.Errorf(codes.Internal, "updating student product have error: %v", err)
		return
	}
	if prevStudentProductID == "" {
		studentProductReturn = studentProduct
		return
	}

	prevStudentProduct, err := s.StudentProductRepo.GetStudentProductByStudentProductID(ctx, db, prevStudentProductID)
	if err != nil {
		err = status.Errorf(codes.Internal, "can't get prev student product with error: %v", err)
		return
	}
	_ = prevStudentProduct.EndDate.Set(lastPeriod.EndDate.Time)
	_ = prevStudentProduct.ProductStatus.Set(pb.StudentProductStatus_ORDERED.String())
	_ = prevStudentProduct.StudentProductLabel.Set(pb.StudentProductLabel_CREATED.String())
	err = s.StudentProductRepo.Update(ctx, db, prevStudentProduct)
	if err != nil {
		err = status.Errorf(codes.Internal, "updating prev student product have error: %v", err)
		return
	}
	studentProductReturn = prevStudentProduct
	return
}

func checkStudentProductLabelByOrderTypeForVoidOrder(orderType string, studentProductLabel string) (err error) {
	switch orderType {
	case pb.OrderType_ORDER_TYPE_WITHDRAWAL.String(), pb.OrderType_ORDER_TYPE_GRADUATE.String():
		if studentProductLabel == pb.StudentProductLabel_UPDATED.String() ||
			studentProductLabel == pb.StudentProductLabel_UPDATE_SCHEDULED.String() {
			err = status.Errorf(codes.Internal, "error when cannot void withdrawal/graduate order if any of the products have UPDATED/UPDATE_SCHEDULED label")
			return
		}
	default:
		if studentProductLabel == pb.StudentProductLabel_UPDATED.String() ||
			studentProductLabel == pb.StudentProductLabel_UPDATE_SCHEDULED.String() ||
			studentProductLabel == pb.StudentProductLabel_WITHDRAWAL_SCHEDULED.String() ||
			studentProductLabel == pb.StudentProductLabel_GRADUATION_SCHEDULED.String() {
			err = status.Errorf(codes.Internal, "error when cannot void if any of the products have UPDATED/UPDATE_SCHEDULED/WITHDRAWAL_SCHEDULED/GRADUATION_SCHEDULED label")
			return
		}
	}
	return
}

func checkEffectiveDateForUpdateOrder(
	orderItem *pb.OrderItem,
	studentProduct entities.StudentProduct,
	timeZone int32,
) (err error) {
	var (
		timeNow          time.Time
		effectiveDate    time.Time
		productStartDate time.Time
		productEndDate   time.Time
	)
	timeNow = utils.TimeNow(timeZone)
	if orderItem.EffectiveDate == nil {
		err = status.Errorf(codes.FailedPrecondition, "Missing effective date of order")
		return
	}
	effectiveDate = utils.StartOfDate(orderItem.EffectiveDate.AsTime().UTC(), timeZone)
	if studentProduct.StartDate.Status != pgtype.Present || studentProduct.EndDate.Status != pgtype.Present {
		err = status.Errorf(codes.FailedPrecondition, "Start date or end date of student product is empty")
		return
	}

	productStartDate = utils.StartOfDate(studentProduct.StartDate.Time, timeZone)
	productEndDate = utils.EndOfDate(studentProduct.EndDate.Time, timeZone)

	if effectiveDate.Before(productStartDate) ||
		effectiveDate.Before(timeNow) ||
		effectiveDate.Equal(productEndDate) ||
		effectiveDate.After(productEndDate) {
		err = utils.StatusErrWithDetail(
			codes.FailedPrecondition,
			constant.UpdateLikeOrdersInvalidEffectiveDate,
			nil,
		)
	}
	return
}

func checkLOADuration(
	orderItem *pb.OrderItem,
	studentProduct entities.StudentProduct,
	timeZone int32,
) (err error) {
	var (
		timeNow   time.Time
		startDate time.Time
		endDate   time.Time
	)
	timeNow = utils.TimeNow(timeZone)
	if orderItem.StartDate == nil {
		err = status.Errorf(codes.FailedPrecondition, "Missing start date of LOA")
		return
	}
	startDate = utils.StartOfDate(orderItem.StartDate.AsTime(), timeZone)

	if orderItem.EndDate == nil {
		err = status.Errorf(codes.FailedPrecondition, "Missing end date of LOA")
		return
	}
	endDate = utils.StartOfDate(orderItem.EndDate.AsTime(), timeZone)

	if studentProduct.StartDate.Status != pgtype.Present || studentProduct.EndDate.Status != pgtype.Present {
		err = status.Errorf(codes.FailedPrecondition, "Start date and end date of student product is empty")
		return
	}

	if startDate.Before(studentProduct.StartDate.Time) ||
		startDate.Before(timeNow) ||
		startDate.Equal(studentProduct.EndDate.Time) ||
		startDate.After(studentProduct.EndDate.Time) {
		err = status.Errorf(codes.FailedPrecondition, "Invalid start date")
	}

	if endDate.Before(startDate) ||
		endDate.Equal(startDate) {
		err = status.Errorf(codes.FailedPrecondition, "Invalid end date")
	}

	return
}

func (s *StudentProductService) RevertUpdateStudentProduct(
	ctx context.Context,
	db database.QueryExecer,
	studentProduct *entities.StudentProduct,
	billingScheduleID string,
) (err error) {
	var updatedToStudentProduct entities.StudentProduct
	if studentProduct.UpdatedToStudentProductID.Status != pgtype.Present {
		err = status.Errorf(codes.FailedPrecondition, "Cannot revert student product without updatedToStudentProductID")
		return
	}
	updatedToStudentProduct, err = s.StudentProductRepo.GetStudentProductForUpdateByStudentProductID(ctx, db, studentProduct.UpdatedToStudentProductID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "Fail when get updated to student product with err: %v", err.Error())
		return
	}

	oldEndDate := updatedToStudentProduct.EndDate.Time
	if updatedToStudentProduct.ProductStatus.String == pb.StudentProductStatus_CANCELLED.String() {
		var billingPeriod entities.BillingSchedulePeriod
		billingPeriod, err = s.BillingSchedulePeriodRepo.GetLatestPeriodByScheduleIDForUpdate(ctx, db, billingScheduleID)
		if err != nil {
			return status.Errorf(codes.Internal, "Can't get latest period with err %v", err.Error())
		}
		oldEndDate = billingPeriod.EndDate.Time
	}
	err = utils.GroupErrorFunc(
		multierr.Combine(
			studentProduct.StudentProductLabel.Set(pb.StudentProductLabel_CREATED.String()),
			studentProduct.EndDate.Set(oldEndDate),
			studentProduct.UpdatedToStudentProductID.Set(nil),
			updatedToStudentProduct.UpdatedFromStudentProductID.Set(nil),
			updatedToStudentProduct.ProductStatus.Set(pb.StudentProductStatus_CANCELLED.String()),
		),
		s.StudentProductRepo.Update(ctx, db, *studentProduct),
		s.StudentProductRepo.Update(ctx, db, updatedToStudentProduct),
	)
	if err != nil {
		err = status.Errorf(codes.Internal, "Fail when revert student product: %v", err.Error())
	}
	return
}

func (s *StudentProductService) GetStudentProductsByStudentProductLabel(
	ctx context.Context,
	db database.QueryExecer,
	studentProductLabels []string,
) (studentProducts []*entities.StudentProduct, err error) {
	for _, label := range studentProductLabels {
		_, ok := pb.StudentProductLabel_value[label]
		if !ok {
			err = status.Errorf(codes.FailedPrecondition, "invalid student product label %v", label)
			return
		}
	}
	studentProducts, err = s.StudentProductRepo.GetStudentProductsByStudentProductLabelForUpdate(ctx, db, studentProductLabels)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting student products by student product label have error: %v", err)
	}
	return
}

func (s *StudentProductService) GetStudentProductByStudentProductIDForUpdate(ctx context.Context, db database.QueryExecer, studentProductID string) (studentProduct entities.StudentProduct, err error) {
	studentProduct, err = s.StudentProductRepo.GetStudentProductForUpdateByStudentProductID(ctx, db, studentProductID)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting student product by student product id with error: %v", err)
	}
	return
}

func (s *StudentProductService) CancelStudentProduct(
	ctx context.Context,
	db database.QueryExecer,
	studentProductID string,
) (err error) {
	err = s.StudentProductRepo.UpdateStatusStudentProductAndResetStudentProductLabel(ctx, db, studentProductID, pb.StudentProductStatus_CANCELLED.String())
	if err != nil {
		err = status.Errorf(codes.Internal, "updating student product status to cancelled and reset student product label")
	}
	return
}

func (s *StudentProductService) PauseStudentProduct(
	ctx context.Context,
	db database.QueryExecer,
	studentProduct entities.StudentProduct,
) (err error) {
	err = studentProduct.StudentProductLabel.Set(pb.StudentProductLabel_PAUSED)
	if err != nil {
		return status.Errorf(codes.Internal, "Error setting student product label for LOA: %v", err.Error())
	}

	err = s.StudentProductRepo.Update(ctx, db, studentProduct)
	if err != nil {
		return status.Errorf(codes.Internal, "Error updating student product label for LOA: %v", err.Error())
	}

	return
}

func (s *StudentProductService) GetUniqueProductsByStudentID(
	ctx context.Context,
	db database.QueryExecer,
	studentID string,
) (studentProductOfUniqueProducts []*entities.StudentProduct, err error) {
	studentProducts, err := s.StudentProductRepo.GetUniqueProductsByStudentID(ctx, db, studentID)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when get student product of unique product by StudentID: %v", err.Error())
		return
	}
	mapProductIDAndStudentProduct := make(map[string]*entities.StudentProduct, len(studentProducts))
	for _, studentProduct := range studentProducts {
		if _, ok := mapProductIDAndStudentProduct[studentProduct.ProductID.String]; ok {
			continue
		}

		if studentProduct.EndDate.Status != pgtype.Present && studentProduct.ProductStatus.String == pb.StudentProductStatus_CANCELLED.String() {
			continue
		}
		mapProductIDAndStudentProduct[studentProduct.ProductID.String] = studentProduct
		studentProductOfUniqueProducts = append(studentProductOfUniqueProducts, studentProduct)
	}
	return
}

func (s *StudentProductService) GetUniqueProductsByStudentIDs(
	ctx context.Context,
	db database.QueryExecer,
	studentIDs []string,
) (mapStudentIDAndStudentProducts map[string][]*entities.StudentProduct, err error) {
	studentProducts, err := s.StudentProductRepo.GetUniqueProductsByStudentIDs(ctx, db, studentIDs)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when get student product of unique product by StudentIDs: %v", err.Error())
		return
	}
	mapStudentIDAndStudentProducts = make(map[string][]*entities.StudentProduct, len(studentIDs))
	mapKeyAndStudentProduct := make(map[string]*entities.StudentProduct, len(studentIDs))

	for _, studentProduct := range studentProducts {
		key := fmt.Sprintf("%s-%s", studentProduct.ProductID.String, studentProduct.StudentID.String)
		if _, ok := mapKeyAndStudentProduct[key]; ok {
			continue
		}
		if studentProduct.EndDate.Status != pgtype.Present && studentProduct.ProductStatus.String == pb.StudentProductStatus_CANCELLED.String() {
			continue
		}
		mapKeyAndStudentProduct[key] = studentProduct
		mapStudentIDAndStudentProducts[studentProduct.StudentID.String] = append(mapStudentIDAndStudentProducts[studentProduct.StudentID.String], studentProduct)
	}
	return
}

func (s *StudentProductService) EndDateOfUniqueRecurringProduct(
	ctx context.Context,
	db database.QueryExecer,
	productID string,
	endDateOfStudentProduct time.Time,
) (endTimeOfUniqueProduct time.Time, err error) {
	product, err := s.ProductRepo.GetByID(ctx, db, productID)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when get product of unique product by productID: %v", err.Error())
		return
	}
	if !product.DisableProRatingFlag.Bool {
		endTimeOfUniqueProduct = endDateOfStudentProduct
		return
	}

	billingSchedulePeriod, err := s.BillingSchedulePeriodRepo.GetPeriodByScheduleIDAndEndTime(ctx, db, product.BillingScheduleID.String, endDateOfStudentProduct)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when get billing_schedule_period of unique product by BillingScheduleID and endTime: %v", err.Error())
		return
	}
	endTimeOfUniqueProduct = billingSchedulePeriod.EndDate.Time
	return
}

func (s *StudentProductService) GetStudentProductByStudentProductID(
	ctx context.Context,
	db database.QueryExecer,
	studentProductID string,
) (
	studentProduct entities.StudentProduct,
	err error,
) {
	studentProduct, err = s.StudentProductRepo.GetStudentProductByStudentProductID(ctx, db, studentProductID)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting student product by student product id have error: %v", err)
	}
	return
}

func (s *StudentProductService) GetStudentProductByStudentIDAndLocationIDs(
	ctx context.Context,
	db database.Ext,
	studentID string,
	locationIDs []string,
	from int64,
	limit int64,
) (
	studentProductIDs []string,
	studentProducts []*entities.StudentProduct,
	total int,
	err error,
) {
	total, err = s.StudentProductRepo.CountStudentProductIDsByStudentIDAndLocationIDs(ctx, db, studentID, locationIDs)
	if err != nil {
		err = status.Errorf(codes.Internal, "counting student product by student id have error: %v", err)
		return
	}
	studentProductIDs, studentProducts, err = s.GetStudentProductIDsAndStudentProductByStudentIDAndLocation(ctx, db, studentID, locationIDs, from, limit)
	if err != nil {
		return nil, nil, 0, err
	}

	return
}

func (s *StudentProductService) GetStudentProductIDsAndStudentProductByStudentIDAndLocation(ctx context.Context, db database.QueryExecer, studentID string, locationIDs []string, offset int64, limit int64) (studentProductIDs []string, studentProductsFiltered []*entities.StudentProduct, err error) {
	studentProducts, err := s.StudentProductRepo.GetByStudentIDAndLocationIDsWithPaging(ctx, db, studentID, locationIDs, offset, limit)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when get student product by student ID with paging: %v", err.Error())
		return
	}

	for _, studentProduct := range studentProducts {
		studentProductFiltered := studentProduct
		if studentProduct.StartDate.Status == pgtype.Present && studentProduct.EndDate.Status == pgtype.Present {
			studentProductFiltered, err = s.GetStudentProductWithRootStudentProductID(ctx, db, studentProduct.StudentProductID.String)
			if err != nil {
				return nil, nil, err
			}
		}

		studentProductsFiltered = append(studentProductsFiltered, studentProductFiltered)
		studentProductIDs = append(studentProductIDs, studentProductFiltered.StudentProductID.String)
	}
	return
}

func (s *StudentProductService) GetStudentProductWithRootStudentProductID(
	ctx context.Context,
	db database.QueryExecer,
	rootStudentProductID string,
) (studentProductFiltered *entities.StudentProduct, err error) {
	studentProducts, err := s.StudentProductRepo.GetStudentProductIDsByRootStudentProductID(ctx, db, rootStudentProductID)
	lenOfStudentProducts := len(studentProducts)
	if err != nil || lenOfStudentProducts == 0 {
		err = status.Errorf(codes.Internal, "Error when get student product by root student product id: %v", err.Error())
		return
	}

	if studentProducts[0].StartDate.Time.After(time.Now()) && !(studentProducts[0].StartDate.Time.After(studentProducts[0].EndDate.Time)) {
		studentProductFiltered = studentProducts[0]
		return
	}

	for _, studentProduct := range studentProducts {
		if (studentProduct.StudentProductLabel.String != pb.StudentProductLabel_PAUSED.String() && studentProduct.StartDate.Time.Before(time.Now()) && studentProduct.EndDate.Time.After(time.Now())) ||
			(studentProduct.StudentProductLabel.String == pb.StudentProductLabel_PAUSED.String() && studentProduct.UpdatedToStudentProductID.Status != pgtype.Present) {
			studentProductFiltered = studentProduct
			return
		}
	}
	studentProductFiltered = studentProducts[lenOfStudentProducts-1]
	return
}

func (s *StudentProductService) GetStudentAssociatedProductByStudentProductID(
	ctx context.Context,
	db database.Ext,
	studentProductID string,
	from int64,
	limit int64,
) (
	studentProductIDs []string,
	studentProducts []*entities.StudentProduct,
	total int,
	err error,
) {
	total, err = s.StudentAssociatedProductRepo.CountAssociatedProductIDsByStudentProductID(ctx, db, studentProductID)
	if err != nil {
		err = status.Errorf(codes.Internal, "counting student associated product by student product id have error: %v", err)
		return
	}
	studentProductIDs, studentProducts, err = s.GetStudentProductIDsAndStudentProductOfProductAssociatedByStudentProductID(ctx, db, studentProductID, from, limit)
	if err != nil {
		return nil, nil, 0, err
	}

	return
}

func (s *StudentProductService) GetStudentProductIDsAndStudentProductOfProductAssociatedByStudentProductID(ctx context.Context, db database.QueryExecer, studentProductID string, offset int64, limit int64) (studentProductIDs []string, studentProductsFiltered []*entities.StudentProduct, err error) {
	studentProductIDs, err = s.StudentAssociatedProductRepo.GetAssociatedProductIDsByStudentProductID(ctx, db, studentProductID, offset, limit)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when get student associated product by student product ID with paging: %v", err.Error())
		return
	}
	studentProductsAssociated, err := s.StudentProductRepo.GetStudentProductAssociatedByStudentProductID(ctx, db, studentProductIDs)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when get student product of product associated by student product ID: %v", err.Error())
		return
	}

	for _, studentProduct := range studentProductsAssociated {
		studentProductFiltered := studentProduct
		if studentProduct.StartDate.Status == pgtype.Present && studentProduct.EndDate.Status == pgtype.Present {
			studentProductFiltered, err = s.GetStudentProductWithRootStudentProductID(ctx, db, studentProduct.StudentProductID.String)
			if err != nil {
				return nil, nil, err
			}
		}

		studentProductsFiltered = append(studentProductsFiltered, studentProductFiltered)
		studentProductIDs = append(studentProductIDs, studentProductFiltered.StudentProductID.String)
	}

	return
}

func (s *StudentProductService) getOriginalEndDateOfStudentProduct(ctx context.Context, db database.QueryExecer, studentProduct entities.StudentProduct) (endDate pgtype.Timestamp, err error) {
	var (
		product          entities.Product
		billingPeriods   []entities.BillingSchedulePeriod
		endBillingPeriod entities.BillingSchedulePeriod
	)

	product, err = s.ProductRepo.GetByID(ctx, db, studentProduct.ProductID.String)
	if err != nil {
		return
	}

	billingPeriods, err = s.BillingSchedulePeriodRepo.GetAllBillingPeriodsByBillingScheduleID(ctx, db, product.BillingScheduleID.String)
	if err != nil {
		return
	}

	if len(billingPeriods) > 0 {
		endBillingPeriod = billingPeriods[len(billingPeriods)-1]
		err = endDate.Set(endBillingPeriod.EndDate)
		if err != nil {
			return
		}
	}

	return
}

func (s *StudentProductService) CreateAssociatedStudentProductByAssociatedStudentProductID(ctx context.Context, db database.QueryExecer, orderItemData utils.OrderItemData) (err error) {
	if orderItemData.OrderItem.AssociatedStudentProductId == nil {
		return
	}
	studentAssociatedProduct := entities.StudentAssociatedProduct{}
	_ = multierr.Combine(
		// Root student product
		studentAssociatedProduct.StudentProductID.Set(orderItemData.OrderItem.AssociatedStudentProductId.Value),
		// Associate student product
		studentAssociatedProduct.AssociatedProductID.Set(orderItemData.StudentProduct.StudentProductID.String),
		studentAssociatedProduct.DeletedAt.Set(nil),
	)
	orderItemType := utils.ConvertOrderItemType(pb.OrderType(pb.OrderType_value[orderItemData.Order.OrderType.String]), orderItemData.BillItems[0].BillingItem)
	if orderItemType == utils.OrderCreate || orderItemType == utils.OrderResume || orderItemType == utils.OrderEnrollment {
		err = s.verifyDuplicated(ctx, db, orderItemData)
		if err != nil {
			return
		}
	}
	err = s.StudentAssociatedProductRepo.Create(ctx, db, studentAssociatedProduct)
	if err != nil {
		err = status.Errorf(codes.Internal, "create associated student product have error:  %v", err.Error())
	}
	return
}

func (s *StudentProductService) verifyDuplicated(ctx context.Context, db database.QueryExecer, orderItemData utils.OrderItemData) (err error) {
	mapProductIDWithAssociatedProductIDs, err := s.StudentAssociatedProductRepo.GetMapAssociatedProducts(ctx, db, orderItemData.OrderItem.AssociatedStudentProductId.Value)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when getting associate product's %v", orderItemData.ProductInfo.ProductID.String)
		return
	}
	if _, ok := mapProductIDWithAssociatedProductIDs[orderItemData.OrderItem.ProductId]; ok {
		studentProductCheck, errGetStudentProduct := s.StudentProductRepo.GetStudentProductForUpdateByStudentProductID(ctx, db, orderItemData.OrderItem.AssociatedStudentProductId.Value)
		if errGetStudentProduct != nil {
			err = status.Errorf(codes.Internal, "fail to get student product: %v with err: %v", orderItemData.OrderItem.AssociatedStudentProductId.Value, errGetStudentProduct.Error())
			return
		}
		if s.associatedProductIsCanceled(studentProductCheck) {
			return
		}
		err = utils.StatusErrWithDetail(
			codes.Internal,
			constant.DuplicatedAssociate,
			&errdetails.DebugInfo{Detail: fmt.Sprintf("product %v is already associated with %v", orderItemData.OrderItem.AssociatedStudentProductId, orderItemData.ProductInfo.ProductID)},
		)
		return
	}
	return
}

func (s *StudentProductService) associatedProductIsCanceled(studentProduct entities.StudentProduct) bool {
	// if student product is canceled, this is also allow to add again
	// one-time and won't have start date
	// recurring product passed due date of cancelled
	if studentProduct.ProductStatus.String == pb.StudentProductStatus_CANCELLED.String() &&
		((studentProduct.StartDate.Status != pgtype.Present) || (studentProduct.StartDate.Status == pgtype.Present && studentProduct.StartDate.Time.Before(time.Now()))) {
		return true
	}
	return false
}

func (s *StudentProductService) DeleteAssociatedStudentProductByAssociatedStudentProductID(ctx context.Context, db database.QueryExecer, orderItemData utils.OrderItemData) (err error) {
	if orderItemData.OrderItem.AssociatedStudentProductId == nil {
		return
	}
	studentAssociatedProduct := entities.StudentAssociatedProduct{}
	_ = multierr.Combine(
		studentAssociatedProduct.AssociatedProductID.Set(orderItemData.StudentProduct.StudentProductID.String),
		studentAssociatedProduct.StudentProductID.Set(orderItemData.OrderItem.AssociatedStudentProductId.Value),
	)
	if !orderItemData.IsOneTimeProduct && orderItemData.RootStudentProduct.StudentProductID.Status == pgtype.Present {
		_ = studentAssociatedProduct.AssociatedProductID.Set(orderItemData.RootStudentProduct.StudentProductID.String)
	}
	err = s.StudentAssociatedProductRepo.Delete(ctx, db, studentAssociatedProduct)
	if err != nil {
		err = status.Errorf(codes.Internal, "delete associated student product have error:  %v", err.Error())
	}
	return
}

func (s *StudentProductService) ValidateProductSettingForCreateOrder(ctx context.Context, db database.QueryExecer, orderItemData utils.OrderItemData) error {
	productID := orderItemData.ProductInfo.ProductID.String
	studentID := orderItemData.StudentInfo.StudentID.String
	locationID := orderItemData.Order.LocationID.String

	if orderItemData.ProductSetting.IsEnrollmentRequired.Status == pgtype.Present {
		if orderItemData.ProductSetting.IsEnrollmentRequired.Bool && !orderItemData.IsEnrolledInLocation {
			return status.Errorf(codes.Internal, "product %v has enrollment required tag but student %v not enrolled in location %v", productID, studentID, locationID)
		}
	}

	return nil
}

func (s *StudentProductService) ValidateProductSettingForLOAOrder(_ context.Context, _ database.QueryExecer, orderItemData utils.OrderItemData) error {
	if orderItemData.ProductSetting.IsPausable.Status == pgtype.Present {
		if !orderItemData.ProductSetting.IsPausable.Bool {
			return status.Errorf(codes.Internal, "LOA order created for product %v but product is not pausable", orderItemData.ProductInfo.ProductID.String)
		}
	}

	return nil
}

func (s *StudentProductService) GetActiveRecurringProductsOfStudentInLocation(ctx context.Context, db database.QueryExecer, studentID string, locationID string) (studentProducts []entities.StudentProduct, err error) {
	ignoreStudentProduct, err := s.StudentProductRepo.GetIgnoreStudentProductIDOfRecurringProductsOfStudentInLocation(ctx, db, studentID, locationID)
	if err != nil {
		err = status.Errorf(codes.Internal, "get ignore student product have error: %v", err.Error())
		return
	}

	studentProducts, err = s.StudentProductRepo.GetActiveRecurringProductsOfStudentInLocation(ctx, db, studentID, locationID, ignoreStudentProduct)
	if err != nil {
		err = status.Errorf(codes.Internal, "get active student product have error: %v", err.Error())
		return
	}

	studentProductsOfOperationFee, err := s.StudentProductRepo.GetActiveOperationFeeOfStudent(ctx, db, studentID)
	if err != nil {
		err = status.Errorf(codes.Internal, "get active student product of operation fee have error: %v", err.Error())
		return
	}

	for _, studentProductOfOperationFee := range studentProductsOfOperationFee {
		locationIDs, err := s.ProductLocationRepo.GetLocationIDsWithProductID(ctx, db, studentProductOfOperationFee.ProductID.String)
		if err != nil {
			err = status.Errorf(codes.Internal, "get locationIDs of product of operation fee have error: %v", err.Error())
			return nil, err
		}
		locationIDs = append(locationIDs, locationID)
		latestListOfStudentEnrollmentStatusHistory, err := s.StudentEnrollmentStatusHistoryRepo.GetLatestStatusEnrollmentByStudentIDAndLocationIDs(ctx, db, studentID, locationIDs)
		if err != nil {
			err = status.Errorf(codes.Internal, "get locationIDs of product of operation fee have error: %v", err.Error())
			return nil, err
		}

		// count location with operation fee product had associated with
		countLocationStillEnrollment := 0
		for _, latestStudentEnrollmentStatusHistory := range latestListOfStudentEnrollmentStatusHistory {
			if latestStudentEnrollmentStatusHistory.EnrollmentStatus.String == upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String() {
				countLocationStillEnrollment++
			}
		}
		// if only have one location enroll with operation fee, will add student_product of operation to cancel
		if countLocationStillEnrollment == 1 {
			studentProducts = append(studentProducts, studentProductOfOperationFee)
		}
	}

	return
}

func (s *StudentProductService) GetRecurringProductsOfStudentInLocationForLOA(ctx context.Context, db database.QueryExecer, studentID string, locationID string) (studentProducts []entities.StudentProduct, err error) {
	ignoreStudentProduct, err := s.StudentProductRepo.GetIgnoreStudentProductIDOfRecurringProductsOfStudentInLocation(ctx, db, studentID, locationID)
	if err != nil {
		err = status.Errorf(codes.Internal, "get ignore student product have error: %v", err.Error())
		return
	}

	studentProductsRecurring, err := s.StudentProductRepo.GetActiveRecurringProductsOfStudentInLocation(ctx, db, studentID, locationID, ignoreStudentProduct)
	if err != nil {
		err = status.Errorf(codes.Internal, "get active student product have error: %v", err.Error())
		return
	}

	for _, studentProduct := range studentProductsRecurring {
		productSetting, err := s.ProductSettingRepo.GetByID(ctx, db, studentProduct.ProductID.String)
		if err != nil {
			return nil, err
		}

		if productSetting.IsPausable.Bool {
			studentProducts = append(studentProducts, studentProduct)
		}
	}

	return
}

func NewStudentProductService() *StudentProductService {
	return &StudentProductService{
		StudentProductRepo:                 &repositories.StudentProductRepo{},
		BillingSchedulePeriodRepo:          &repositories.BillingSchedulePeriodRepo{},
		StudentAssociatedProductRepo:       &repositories.StudentAssociatedProductRepo{},
		ProductRepo:                        &repositories.ProductRepo{},
		PackageRepo:                        &repositories.PackageRepo{},
		ProductSettingRepo:                 &repositories.ProductSettingRepo{},
		ProductLocationRepo:                &repositories.ProductLocationRepo{},
		StudentEnrollmentStatusHistoryRepo: &repositories.StudentEnrollmentStatusHistoryRepo{},
	}
}
