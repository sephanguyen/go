package service

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Group func for cancel order and change student package and student course

func (s *StudentPackageService) MutationStudentPackageForCancelOrder(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (
	eventMessages []*npb.EventStudentPackage,
	err error,
) {
	var (
		previousStudentProduct entities.StudentProduct
		studentProductID       = orderItemData.OrderItem.StudentProductId.Value
		studentID              = orderItemData.Order.StudentID.String
	)

	mapStudentCourseKeyWithStudentPackageAccessPath, err := s.StudentPackageAccessPathRepo.
		GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs(ctx, db, []string{studentID})
	if err != nil {
		return
	}
	if orderItemData.RootStudentProduct.StudentProductID.String == studentProductID {
		previousStudentProduct = orderItemData.RootStudentProduct
	} else {
		previousStudentProduct, err = s.StudentProductRepo.GetStudentProductForUpdateByStudentProductID(ctx, db, studentProductID)
		if err != nil {
			err = fmt.Errorf("error when getting student product by student product id %v", err.Error())
			return
		}
	}

	isOneTimeProductOrCancelCompleteOrder := orderItemData.IsOneTimeProduct ||
		previousStudentProduct.StartDate.Time.Equal(utils.StartOfDate(
			orderItemData.OrderItem.EffectiveDate.AsTime(),
			orderItemData.Timezone,
		))

	if isOneTimeProductOrCancelCompleteOrder {
		for _, item := range orderItemData.OrderItem.CourseItems {
			var eventMessage *npb.EventStudentPackage
			eventMessage, err = s.cancelStudentPackageDataForCompleteCancelOrder(ctx, db, orderItemData, item.CourseId, mapStudentCourseKeyWithStudentPackageAccessPath)
			if err != nil {
				return
			}
			if eventMessage != nil {
				eventMessages = append(eventMessages, eventMessage)
			}
		}
		return
	}
	for _, item := range orderItemData.OrderItem.CourseItems {
		var eventMessage *npb.EventStudentPackage
		eventMessage, err = s.cancelStudentPackageDataForNonCompleteUpdateOrder(ctx, db, orderItemData, item.CourseId, mapStudentCourseKeyWithStudentPackageAccessPath)
		if err != nil {
			return
		}
		if eventMessage != nil {
			eventMessages = append(eventMessages, eventMessage)
		}
	}
	return
}

func (s *StudentPackageService) cancelStudentPackageDataForCompleteCancelOrder(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
	courseID string,
	mapStudentCourseWithStudentPackageAccessPath map[string]entities.StudentPackageAccessPath,
) (eventMessage *npb.EventStudentPackage, err error) {
	var (
		studentID                                = orderItemData.StudentInfo.StudentID.String
		key                                      = fmt.Sprintf("%v_%v", studentID, courseID)
		currentStudentPackageAccessPath          = mapStudentCourseWithStudentPackageAccessPath[key]
		action                                   = pb.StudentPackageActions_STUDENT_PACKAGE_ACTION_CANCELED.String()
		flow                                     = "Cancel student package"
		newCurrentStudentPackageOrder            *entities.StudentPackageOrder
		studentPackageID                         = currentStudentPackageAccessPath.StudentPackageID.String
		hasCurrentStudentPackageOrderAfterCancel bool
	)
	startTime := getStartTimeFromOrder(orderItemData.OrderItem)
	studentPackageOrderForCreateOrder, err := s.StudentPackageOrderService.GetStudentPackageOrderByStudentPackageIDAndTime(ctx, db, studentPackageID,
		utils.ConvertToLocalTime(startTime, orderItemData.Timezone))
	if err != nil {
		return
	}

	// This check is used for invalid data (exists before student_package_order) => missing student_package_order for this student_course
	if studentPackageOrderForCreateOrder == nil {
		err = status.Errorf(codes.FailedPrecondition, "missing student package order with student package id = %s and effective date = %v", studentPackageID, startTime)
		return
	}
	err = s.StudentPackageOrderService.DeleteStudentPackageOrderByID(ctx, db, studentPackageOrderForCreateOrder.ID.String)
	if err != nil {
		return
	}

	studentPackage, err := s.StudentPackageRepo.GetByID(ctx, db, studentPackageOrderForCreateOrder.StudentPackageID.String)
	if err != nil {
		return
	}

	studentPackageOrderPosition, err := s.getPositionOfStudentPackageOrder(ctx, db, *studentPackageOrderForCreateOrder)
	if err != nil {
		return
	}

	studentPackageOrderAfterCancel := *studentPackageOrderForCreateOrder
	err = utils.GroupErrorFunc(
		studentPackageOrderAfterCancel.ID.Set(uuid.NewString()),
		studentPackageOrderAfterCancel.OrderID.Set(orderItemData.Order.OrderID.String),
		studentPackageOrderAfterCancel.StartAt.Set(nil),
		studentPackageOrderAfterCancel.EndAt.Set(nil),
		studentPackageOrderAfterCancel.FromStudentPackageOrderID.Set(studentPackageOrderForCreateOrder.ID),
	)
	if err != nil {
		return
	}

	switch studentPackageOrderPosition {
	case entities.PastStudentPackage:
		err = status.Errorf(codes.Internal, "error when cancel student package in past time with student_package_id = %v and student_package_order_id = %v", studentPackageOrderForCreateOrder.StudentPackageID.String, studentPackageOrderForCreateOrder.ID.String)
		return
	case entities.CurrentStudentPackage:
		newCurrentStudentPackageOrder, err = s.StudentPackageOrderService.SetCurrentStudentPackageOrderByTimeAndStudentPackageID(ctx, db, studentPackage.ID.String)
		if err != nil {
			return
		}

		if err := utils.GroupErrorFunc(
			s.StudentPackageOrderService.InsertStudentPackageOrder(ctx, db, studentPackageOrderAfterCancel, entities.CurrentStudentPackage)); err != nil {
			return nil, err
		}
		hasCurrentStudentPackageOrderAfterCancel = newCurrentStudentPackageOrder != nil
		if hasCurrentStudentPackageOrderAfterCancel {
			var (
				studentCourseAfterCancel  *entities.StudentCourse
				studentPackageAfterCancel entities.StudentPackages
			)
			studentCourseAfterCancel, studentPackageAfterCancel, eventMessage, err = s.convertStudentPackageDataForCancelOrder(
				orderItemData.PackageInfo.Package.PackageType.String,
				orderItemData.PackageInfo.QuantityType.String(),
				*newCurrentStudentPackageOrder)
			if err != nil {
				return
			}

			err = utils.GroupErrorFunc(
				s.StudentPackageRepo.Upsert(ctx, db, &studentPackageAfterCancel),
				s.StudentCourseRepo.UpsertStudentCourse(ctx, db, *studentCourseAfterCancel),
				s.writeStudentPackageLog(ctx, db, &studentPackage, currentStudentPackageAccessPath.CourseID.String, action, flow),
			)
			if err != nil {
				return
			}
		} else {
			eventMessage = &npb.EventStudentPackage{
				StudentPackage: &npb.EventStudentPackage_StudentPackage{
					StudentId: studentID,
					IsActive:  false,
					Package: &npb.EventStudentPackage_Package{
						StudentPackageId: currentStudentPackageAccessPath.StudentPackageID.String,
						LocationIds: []string{
							orderItemData.Order.LocationID.String,
						},
						CourseIds: []string{
							courseID,
						},
						StartDate: timestamppb.New(time.Now().AddDate(0, 0, -1)),
						EndDate:   timestamppb.New(time.Now().AddDate(0, 0, -1)),
					},
				},
				LocationIds: []string{
					orderItemData.Order.LocationID.String,
				},
			}
			err = utils.GroupErrorFunc(
				s.StudentPackageRepo.CancelByID(ctx, db, currentStudentPackageAccessPath.StudentPackageID.String),
				s.StudentCourseRepo.CancelByStudentPackageIDAndCourseID(ctx, db,
					currentStudentPackageAccessPath.StudentPackageID.String, courseID),
				s.StudentPackageAccessPathRepo.DeleteMulti(ctx, db, []entities.StudentPackageAccessPath{currentStudentPackageAccessPath}),
				s.writeStudentPackageLog(ctx, db, &studentPackage, currentStudentPackageAccessPath.CourseID.String, action, flow),
			)
			if err != nil {
				return
			}
		}
	case entities.FutureStudentPackage:
		err = s.StudentPackageOrderService.InsertStudentPackageOrder(ctx, db, studentPackageOrderAfterCancel, entities.FutureStudentPackage)
		if err != nil {
			return
		}
	}
	return
}

func (s *StudentPackageService) cancelStudentPackageDataForNonCompleteUpdateOrder(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
	courseID string,
	mapStudentCourseWithStudentPackageAccessPath map[string]entities.StudentPackageAccessPath,
) (
	eventMessage *npb.EventStudentPackage,
	err error,
) {
	var (
		action                          = pb.StudentPackageActions_STUDENT_PACKAGE_ACTION_CANCELED.String()
		flow                            = "Update time"
		studentID                       = orderItemData.StudentInfo.StudentID.String
		studentCourseKey                = fmt.Sprintf("%v_%v", studentID, courseID)
		currentStudentPackageAccessPath = mapStudentCourseWithStudentPackageAccessPath[studentCourseKey]
		effectiveDate                   = utils.EndOfDate(orderItemData.OrderItem.EffectiveDate.AsTime(), orderItemData.Timezone)
		studentPackageID                = currentStudentPackageAccessPath.StudentPackageID.String
		isCurrentStudentPackageOrder    entities.StudentPackagePosition
	)
	studentPackage, err := s.StudentPackageRepo.GetByID(ctx, db, studentPackageID)
	if err != nil {
		return
	}
	startTime := getStartTimeFromOrder(orderItemData.OrderItem)
	studentPackageOrderForCreateOrder, err := s.StudentPackageOrderService.GetStudentPackageOrderByStudentPackageIDAndTime(
		ctx, db,
		studentPackageID,
		utils.ConvertToLocalTime(startTime, orderItemData.Timezone))
	if err != nil {
		return
	}

	// This check is used for invalid data (exists before student_package_order) => missing student_package_order for this student_course
	if studentPackageOrderForCreateOrder == nil {
		err = status.Errorf(codes.FailedPrecondition, "missing student package order with student package id = %s and effective date = %v", studentPackageID, startTime)
		return
	}
	studentPackageOrderAfterCancel := *studentPackageOrderForCreateOrder
	cancelledStudentPackage, err := studentPackageOrderForCreateOrder.GetStudentPackageObject()
	if err != nil {
		return
	}

	err = utils.GroupErrorFunc(
		cancelledStudentPackage.EndAt.Set(effectiveDate),

		studentPackageOrderAfterCancel.ID.Set(uuid.NewString()),
		studentPackageOrderAfterCancel.OrderID.Set(orderItemData.Order.OrderID),
		studentPackageOrderAfterCancel.EndAt.Set(effectiveDate),
		studentPackageOrderAfterCancel.FromStudentPackageOrderID.Set(studentPackageOrderForCreateOrder.ID),
		studentPackageOrderAfterCancel.StudentPackageObject.Set(cancelledStudentPackage),
	)
	if err != nil {
		return
	}

	isCurrentStudentPackageOrder, err = s.getPositionOfStudentPackageOrder(ctx, db, *studentPackageOrderForCreateOrder)
	if err != nil {
		return
	}
	switch isCurrentStudentPackageOrder {
	case entities.PastStudentPackage:
		err = status.Errorf(codes.Internal, "error when cancel student package in past time with student_package_id = %v and student_package_order_id = %v", studentPackageOrderForCreateOrder.StudentPackageID.String, studentPackageOrderForCreateOrder.ID.String)
		return
	case entities.CurrentStudentPackage:
		err = utils.GroupErrorFunc(
			s.StudentPackageRepo.UpdateTimeByID(ctx, db, studentPackage.ID.String, effectiveDate),
			s.StudentCourseRepo.UpdateTimeByID(ctx, db, studentPackage.ID.String, courseID, effectiveDate),
			s.StudentPackageOrderService.InsertStudentPackageOrder(ctx, db, studentPackageOrderAfterCancel, entities.CurrentStudentPackage),
			s.writeStudentPackageLog(ctx, db, &studentPackage, currentStudentPackageAccessPath.CourseID.String, action, flow),
			s.StudentPackageOrderService.DeleteStudentPackageOrderByID(ctx, db, studentPackageOrderForCreateOrder.ID.String),
		)
		if err != nil {
			return
		}
		eventMessage = &npb.EventStudentPackage{
			StudentPackage: &npb.EventStudentPackage_StudentPackage{
				StudentId: studentID,
				IsActive:  true,
				Package: &npb.EventStudentPackage_Package{
					StudentPackageId: currentStudentPackageAccessPath.StudentPackageID.String,
					LocationIds: []string{
						orderItemData.Order.LocationID.String,
					},
					CourseIds: []string{
						courseID,
					},
					StartDate: timestamppb.New(studentPackage.StartAt.Time),
					EndDate:   timestamppb.New(effectiveDate),
				},
			},
			LocationIds: []string{
				orderItemData.Order.LocationID.String,
			},
		}
	case entities.FutureStudentPackage:
		err = utils.GroupErrorFunc(
			s.StudentPackageOrderService.InsertStudentPackageOrder(ctx, db, studentPackageOrderAfterCancel, entities.FutureStudentPackage),
			s.StudentPackageOrderService.DeleteStudentPackageOrderByID(ctx, db, studentPackageOrderForCreateOrder.ID.String),
		)
		if err != nil {
			return
		}
	}
	return
}

func (s *StudentPackageService) convertStudentPackageDataForCancelOrder(
	packageType,
	quantityType string,
	studentPackageOrder entities.StudentPackageOrder,
) (
	studentCourse *entities.StudentCourse,
	studentPackage entities.StudentPackages,
	eventMessage *npb.EventStudentPackage,
	err error,
) {
	var (
		property *entities.PackageProperties
	)
	studentPackage, err = studentPackageOrder.GetStudentPackageObject()
	if err != nil {
		return
	}
	property, err = studentPackage.GetProperties()
	if err != nil {
		return
	}
	courseInfo := property.AllCourseInfo[0]
	studentCourse = &entities.StudentCourse{
		StudentPackageID:  pgtype.Text{Status: pgtype.Present, String: studentPackage.ID.String},
		StudentID:         studentPackage.StudentID,
		CourseID:          pgtype.Text{Status: pgtype.Present, String: courseInfo.CourseID},
		LocationID:        studentPackage.LocationIDs.Elements[0],
		StudentStartDate:  studentPackage.StartAt,
		StudentEndDate:    studentPackage.EndAt,
		CourseSlot:        pgtype.Int4{Status: pgtype.Null},
		CourseSlotPerWeek: pgtype.Int4{Status: pgtype.Null},
		Weight:            pgtype.Int4{Status: pgtype.Null},
		CreatedAt:         pgtype.Timestamptz{Time: studentPackageOrder.CreatedAt.Time, Status: pgtype.Present},
		UpdatedAt:         pgtype.Timestamptz{Time: studentPackageOrder.UpdatedAt.Time, Status: pgtype.Present},
		DeletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
		PackageType:       pgtype.Text{Status: pgtype.Present, String: packageType},
		ResourcePath:      pgtype.Text{Status: pgtype.Null},
	}
	switch pb.QuantityType(pb.QuantityType_value[quantityType]) {
	case pb.QuantityType_QUANTITY_TYPE_SLOT:
		_ = studentCourse.CourseSlot.Set(courseInfo.NumberOfSlots)
	case pb.QuantityType_QUANTITY_TYPE_SLOT_PER_WEEK:
		_ = studentCourse.CourseSlotPerWeek.Set(courseInfo.NumberOfSlots)
	case pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT:
		_ = studentCourse.Weight.Set(courseInfo.Weight)
	}

	eventMessage = &npb.EventStudentPackage{
		StudentPackage: &npb.EventStudentPackage_StudentPackage{
			StudentId: studentPackage.StudentID.String,
			Package: &npb.EventStudentPackage_Package{
				CourseIds:        []string{studentCourse.CourseID.String},
				StartDate:        timestamppb.New(studentPackage.StartAt.Time),
				EndDate:          timestamppb.New(studentPackage.EndAt.Time),
				LocationIds:      database.FromTextArray(studentPackage.LocationIDs),
				StudentPackageId: studentPackage.ID.String,
			},
			IsActive: true,
		},
		LocationIds: database.FromTextArray(studentPackage.LocationIDs),
	}

	return
}

func (s *StudentPackageService) voidStudentPackageForCancelOrder(
	ctx context.Context,
	db database.QueryExecer,
	args utils.VoidStudentPackageArgs,
	mapStudentCourseKeyWithStudentPackageAccessPath map[string]entities.StudentPackageAccessPath,
) (
	studentPackageEvents []*npb.EventStudentPackage,
	err error,
) {
	var (
		studentID = args.Order.StudentID.String
	)

	mapCourseIDWithOrderItemCourse, err := s.OrderItemCourseRepo.GetMapOrderItemCourseByOrderIDAndPackageID(ctx, db, args.Order.OrderID.String, args.Product.ProductID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get map order item course for void update order with order_id = %s and package_id = %s and error = %v", args.Order.OrderID.String, args.Product.ProductID.String, err)
		return
	}
	for courseID, course := range mapCourseIDWithOrderItemCourse {
		var event *npb.EventStudentPackage
		studentCourseKey := fmt.Sprintf("%v_%v", studentID, courseID)
		studentPackageAccessPath := mapStudentCourseKeyWithStudentPackageAccessPath[studentCourseKey]
		event, err = s.voidStudentPackageDataForCancelOrder(ctx, db, args, course, studentPackageAccessPath)
		if err != nil {
			err = status.Errorf(codes.Internal, "error when void student package data for cancel order with error = %v", err)
			return
		}
		if event != nil {
			studentPackageEvents = append(studentPackageEvents, event)
		}
	}
	return
}

func (s *StudentPackageService) voidStudentPackageDataForCancelOrder(
	ctx context.Context,
	db database.QueryExecer,
	args utils.VoidStudentPackageArgs,
	orderItemCourse entities.OrderItemCourse,
	studentPackageAccessPath entities.StudentPackageAccessPath,
) (
	eventMessage *npb.EventStudentPackage,
	err error,
) {
	var (
		currentStudentPackageOrder,
		studentPackageOrderAfterCancel *entities.StudentPackageOrder
		studentPackageID = studentPackageAccessPath.StudentPackageID.String
		studentPackage   entities.StudentPackages
		studentCourse    *entities.StudentCourse
		action           = pb.StudentPackageActions_STUDENT_PACKAGE_ACTION_UPSERT.String()
		flow             = "Void cancel order flow"
	)
	// Used for case Void Update Order
	if studentPackageAccessPath.StudentPackageID.String == "" || studentPackageAccessPath.StudentPackageID.Status != pgtype.Present {
		err = s.StudentPackageAccessPathRepo.RevertByStudentIDAndCourseID(ctx, db, args.StudentProduct.StudentID.String, orderItemCourse.CourseID.String)
		if err != nil {
			err = status.Errorf(codes.Internal, "error when revert student package access path with student_id = %s and course_id = %s and error = %v", args.StudentProduct.StudentID.String, orderItemCourse.CourseID.String, err)
			return
		}
		studentPackageAccessPath, err = s.StudentPackageAccessPathRepo.GetByStudentIDAndCourseID(ctx, db, args.StudentProduct.StudentID.String, orderItemCourse.CourseID.String)
		if err != nil {
			err = status.Errorf(codes.Internal, "error when get student package access path with student_id = %s and course_id = %s and error = %v", args.StudentProduct.StudentID.String, orderItemCourse.CourseID.String, err)
			return
		}
		studentPackageID = studentPackageAccessPath.StudentPackageID.String
	}

	packageData, err := s.PackageRepo.GetByID(ctx, db, orderItemCourse.PackageID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get package with package_id = %s and error = %v", orderItemCourse.PackageID.String, err)
		return
	}
	quantityType, err := s.PackageQuantityTypeMappingRepo.GetByPackageTypeForUpdate(ctx, db, packageData.PackageType.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get quantity type with package_type = %s and error = %v", packageData.PackageType.String, err)
		return
	}

	studentPackageOrderAfterCancel, err = s.StudentPackageOrderService.GetStudentPackageOrderByStudentPackageIDAndOrderID(ctx, db, studentPackageID, args.Order.OrderID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get student package order with student_package_id = %s and order_id = %s and error = %v", studentPackageID, args.Order.OrderID.String, err)
		return
	}

	if studentPackageOrderAfterCancel.FromStudentPackageOrderID.Status == pgtype.Present {
		err = s.StudentPackageOrderService.RevertStudentPackageOrderByStudentPackageOrderID(ctx, db, studentPackageOrderAfterCancel.FromStudentPackageOrderID.String)
		if err != nil {
			err = status.Errorf(codes.Internal, "error when revert student package order by id with student_package_order_id = %s and error = %v", studentPackageOrderAfterCancel.FromStudentPackageOrderID.String, err)
			return
		}
	}

	studentPackageOrderPosition, err := s.getPositionOfStudentPackageOrder(ctx, db, *studentPackageOrderAfterCancel)
	if err != nil {
		return
	}

	err = s.StudentPackageOrderService.DeleteStudentPackageOrderByID(ctx, db, studentPackageOrderAfterCancel.ID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when delete student package order after cancel by id with student_package_order_id = %s and error = %v", studentPackageOrderAfterCancel.ID.String, err)
		return
	}
	switch studentPackageOrderPosition {
	case entities.PastStudentPackage:
		err = status.Errorf(codes.Internal, "error when void student package in past time with student_package_id = %v and student_package_order_id = %v", studentPackageOrderAfterCancel.StudentPackageID.String, studentPackageOrderAfterCancel.ID.String)
		return
	case entities.CurrentStudentPackage:
		currentStudentPackageOrder, err = s.StudentPackageOrderService.SetCurrentStudentPackageOrderByTimeAndStudentPackageID(ctx, db, studentPackageID)
		if err != nil {
			return
		}

		studentCourse, studentPackage, eventMessage, err = s.convertStudentPackageDataForCancelOrder(
			packageData.PackageType.String,
			quantityType.String(),
			*currentStudentPackageOrder)
		if err != nil {
			return
		}
		err = utils.GroupErrorFunc(
			s.StudentPackageRepo.Upsert(ctx, db, &studentPackage),
			s.StudentCourseRepo.UpsertStudentCourse(ctx, db, *studentCourse),
			s.writeStudentPackageLog(ctx, db, &studentPackage, studentCourse.CourseID.String, action, flow),
		)
		if err != nil {
			return
		}
	case entities.FutureStudentPackage:
	}
	return
}
