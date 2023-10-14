package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/repositories"
	service "github.com/manabie-com/backend/internal/payment/services/domain_service/student_package/student_package_order"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StudentPackageOrderService interface {
	InsertStudentPackageOrder(
		ctx context.Context,
		db database.QueryExecer,
		studentPackageOrder entities.StudentPackageOrder,
		positionStudentPackageOrder entities.StudentPackagePosition,
	) (err error)
	GetPositionForStudentPackageByTime(
		ctx context.Context,
		db database.QueryExecer,
		studentPackageID string,
		startTime time.Time,
		endTime time.Time,
	) (
		studentPackagePosition entities.StudentPackagePosition,
		err error,
	)
	SetCurrentStudentPackageOrderByTimeAndStudentPackageID(
		ctx context.Context,
		db database.QueryExecer,
		studentPackageID string,
	) (
		studentPackageOrder *entities.StudentPackageOrder,
		err error,
	)
	GetStudentPackageOrderByStudentPackageIDAndTime(
		ctx context.Context,
		db database.QueryExecer,
		studentPackageID string,
		startTime time.Time,
	) (
		studentPackageOrder *entities.StudentPackageOrder,
		err error,
	)
	DeleteStudentPackageOrderByID(ctx context.Context, db database.QueryExecer, studentPackageOrderID string) (
		err error,
	)
	UpdateStudentPackageOrder(ctx context.Context, db database.QueryExecer, studentPackageOrder entities.StudentPackageOrder,
	) (err error)
	GetStudentPackageOrderByStudentPackageIDAndOrderID(
		ctx context.Context,
		db database.QueryExecer,
		studentPackageID, orderID string,
	) (
		studentPackageOrder *entities.StudentPackageOrder,
		err error,
	)
	RevertStudentPackageOrderByStudentPackageOrderID(
		ctx context.Context,
		db database.QueryExecer,
		studentPackageID string,
	) (
		err error,
	)
	GetStudentPackageOrderByStudentPackageOrderID(
		ctx context.Context,
		db database.QueryExecer,
		studentPackageOrderID string,
	) (
		studentPackageOrder *entities.StudentPackageOrder,
		err error,
	)
	UpdateExecuteError(
		ctx context.Context,
		db database.QueryExecer,
		studentPackageOrder entities.StudentPackageOrder,
	) (err error)
	UpdateExecuteStatus(
		ctx context.Context,
		db database.QueryExecer,
		studentPackageOrder entities.StudentPackageOrder,
	) (err error)
	GetCurrentStudentPackageOrderByStudentPackageID(
		ctx context.Context,
		db database.QueryExecer,
		studentPackageID string,
	) (
		studentPackageOrder *entities.StudentPackageOrder,
		err error,
	)
}

type StudentPackageService struct {
	StudentPackageRepo interface {
		Insert(ctx context.Context, db database.QueryExecer, studentPackage *entities.StudentPackages) (err error)
		Update(ctx context.Context, db database.QueryExecer, studentPackage *entities.StudentPackages) (err error)
		GetByID(ctx context.Context, db database.QueryExecer, studentPackageID string) (studentPackage entities.StudentPackages, err error)
		Upsert(ctx context.Context, db database.QueryExecer, e *entities.StudentPackages) (err error)
		SoftDeleteByIDs(ctx context.Context, db database.QueryExecer, ids []string, deletedAt time.Time) error
		UpdateTimeByID(ctx context.Context, db database.QueryExecer, id string, endTime time.Time) (err error)
		CancelByID(ctx context.Context, db database.QueryExecer, id string) (err error)
		GetStudentPackagesForCronjobByDay(ctx context.Context, db database.QueryExecer, day int) (studentPackages []entities.StudentPackages, err error)
	}
	StudentPackageAccessPathRepo interface {
		Insert(ctx context.Context, db database.QueryExecer, studentPackageAccessPath *entities.StudentPackageAccessPath) (err error)
		InsertMulti(ctx context.Context, db database.QueryExecer, studentPackageAccessPaths []entities.StudentPackageAccessPath) (err error)
		DeleteMulti(ctx context.Context, db database.QueryExecer, studentPackageAccessPaths []entities.StudentPackageAccessPath) (err error)
		Update(ctx context.Context, db database.QueryExecer, studentPackageAccessPath *entities.StudentPackageAccessPath) (err error)
		GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs(
			ctx context.Context,
			db database.QueryExecer,
			studentIDs []string,
		) (
			mapStudentCourseWithStudentPackageAccessPath map[string]entities.StudentPackageAccessPath,
			err error,
		)
		SoftDeleteByStudentPackageIDs(ctx context.Context, db database.QueryExecer, ids []string, deletedAt time.Time) error
		CheckExistStudentPackageAccessPath(ctx context.Context, db database.QueryExecer,
			studentID, courseID string,
		) (
			err error,
		)
		RevertByStudentIDAndCourseID(ctx context.Context, db database.QueryExecer, studentID, courseID string) error
		GetByStudentIDAndCourseID(ctx context.Context, db database.QueryExecer, studentID, courseID string) (entities.StudentPackageAccessPath, error)
	}
	StudentPackageClassRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, studentPackageClass *entities.StudentPackageClass) (err error)
		Delete(ctx context.Context, db database.QueryExecer, studentPackageClass *entities.StudentPackageClass) (err error)
	}
	StudentCourseRepo interface {
		UpsertStudentCourseData(
			ctx context.Context,
			tx database.QueryExecer,
			studentCourseEntities []entities.StudentCourse,
		) (
			err error,
		)
		GetStudentCoursesByStudentPackageIDForUpdate(
			ctx context.Context,
			tx database.QueryExecer,
			studentPackageID string,
		) (
			studentCourseEntities []entities.StudentCourse,
			err error,
		)
		SoftDeleteByStudentPackageIDs(ctx context.Context, db database.QueryExecer, ids []string, deletedAt time.Time) error
		VoidStudentCoursesByStudentPackageID(
			ctx context.Context,
			tx database.QueryExecer,
			studentEndDate time.Time,
			studentPackageID string,
		) (
			err error,
		)
		GetStudentCoursesByStudentPackageIDsForUpdate(ctx context.Context, db database.QueryExecer, studentPackageIDs []string) (studentCourses []entities.StudentCourse, err error)
		UpsertStudentCourse(
			ctx context.Context,
			tx database.QueryExecer,
			studentCourse entities.StudentCourse,
		) (
			err error,
		)
		UpdateTimeByID(
			ctx context.Context,
			db database.QueryExecer,
			studentPackageID string,
			courseID string,
			endTime time.Time,
		) (
			err error,
		)
		CancelByStudentPackageIDAndCourseID(
			ctx context.Context,
			db database.QueryExecer,
			studentPackageID string,
			courseID string,
		) (
			err error,
		)
		GetByStudentIDAndCourseIDAndLocationID(ctx context.Context, db database.QueryExecer,
			studentID, courseID, locationID string) (studentCourse entities.StudentCourse, err error)
	}
	OrderItemCourseRepo interface {
		GetMapOrderItemCourseByOrderIDAndPackageID(ctx context.Context, db database.QueryExecer, orderID, packageID string) (mapOrderItemCourse map[string]entities.OrderItemCourse, err error)
		GetMapOrderItemCourseByOrderID(ctx context.Context, db database.QueryExecer, orderID string) (mapOrderItemCourse map[string]entities.OrderItemCourse, err error)
	}
	OrderItemRepo interface {
		GetOrderItemByStudentProductID(ctx context.Context, db database.QueryExecer, studentProductID string) (orderItem entities.OrderItem, err error)
	}
	StudentProductRepo interface {
		GetStudentProductForUpdateByStudentProductID(
			ctx context.Context,
			db database.QueryExecer,
			studentProductID string,
		) (
			studentProduct entities.StudentProduct,
			err error,
		)
	}
	StudentPackageLogRepo interface {
		Create(ctx context.Context, db database.QueryExecer, studentPackageLog *entities.StudentPackageLog) (err error)
	}
	StudentPackageOrderRepo interface {
		Create(ctx context.Context, db database.QueryExecer, studentPackageOrder entities.StudentPackageOrder) (err error)
	}

	PackageRepo interface {
		GetByID(ctx context.Context, db database.QueryExecer, packageID string) (entities.Package, error)
	}
	PackageQuantityTypeMappingRepo interface {
		GetByPackageTypeForUpdate(ctx context.Context, db database.QueryExecer, packageType string) (pb.QuantityType, error)
	}
	StudentPackageOrderService StudentPackageOrderService
	ProductRepo                interface {
		GetByIDForUpdate(ctx context.Context, db database.QueryExecer, entitiesID string) (entities.Product, error)
	}
	OrderRepo interface {
		GetOrderByIDForUpdate(ctx context.Context, db database.QueryExecer, orderID string) (order entities.Order, err error)
	}
}

func NewStudentPackage() *StudentPackageService {
	return &StudentPackageService{
		PackageRepo:                    &repositories.PackageRepo{},
		OrderItemRepo:                  &repositories.OrderItemRepo{},
		StudentCourseRepo:              &repositories.StudentCourseRepo{},
		StudentPackageRepo:             &repositories.StudentPackageRepo{},
		StudentProductRepo:             &repositories.StudentProductRepo{},
		OrderItemCourseRepo:            &repositories.OrderItemCourseRepo{},
		StudentPackageClassRepo:        &repositories.StudentPackageClassRepo{},
		StudentPackageAccessPathRepo:   &repositories.StudentPackageAccessPathRepo{},
		PackageQuantityTypeMappingRepo: &repositories.PackageQuantityTypeMappingRepo{},
		StudentPackageLogRepo:          &repositories.StudentPackageLogRepo{},
		StudentPackageOrderRepo:        &repositories.StudentPackageOrderRepo{},
		StudentPackageOrderService:     service.NewStudentPackageOrder(),
		ProductRepo:                    &repositories.ProductRepo{},
		OrderRepo:                      &repositories.OrderRepo{},
	}
}

func (s *StudentPackageService) writeStudentPackageLog(
	ctx context.Context,
	db database.QueryExecer,
	studentPackage *entities.StudentPackages,
	courseID string,
	action string,
	flow string,
) (err error) {
	userID := interceptors.UserIDFromContext(ctx)
	claims := interceptors.JWTClaimsFromContext(ctx)
	if claims != nil && claims.Manabie != nil && claims.Manabie.UserID != "" {
		userID = claims.Manabie.UserID
	}

	studentPackageBytes, err := json.Marshal(studentPackage)
	if err != nil {
		err = status.Errorf(codes.Internal, "error while convert student package to json %v", err.Error())
		return
	}
	studentPackageActionLog := entities.StudentPackageLog{
		StudentPackageID: studentPackage.ID,
		UserID: pgtype.Text{
			String: userID,
			Status: pgtype.Present,
		},
		Action: pgtype.Text{
			String: action,
			Status: pgtype.Present,
		},
		Flow: pgtype.Text{
			String: flow,
			Status: pgtype.Present,
		},
		StudentPackageObject: pgtype.JSONB{
			Bytes:  studentPackageBytes,
			Status: pgtype.Present,
		},
		CourseID: pgtype.Text{
			String: courseID,
			Status: pgtype.Present,
		},
		StudentID: studentPackage.StudentID,
		CreatedAt: pgtype.Timestamptz{
			Time:   time.Now(),
			Status: pgtype.Present,
		},
	}

	if err = s.StudentPackageLogRepo.Create(ctx, db, &studentPackageActionLog); err != nil {
		err = status.Errorf(codes.Internal, "creating student package action log have error %v", err.Error())
		return
	}
	return
}

func getStartTimeFromOrder(orderItem *pb.OrderItem) time.Time {
	if orderItem.StartDate != nil {
		return orderItem.StartDate.AsTime()
	}
	return orderItem.EffectiveDate.AsTime()
}
