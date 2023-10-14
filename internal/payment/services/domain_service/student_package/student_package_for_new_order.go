package service

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Group func for create order and change student package and student course

func (s *StudentPackageService) MutationStudentPackageForCreateOrder(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (
	eventMessages []*npb.EventStudentPackage,
	err error,
) {
	studentID := orderItemData.StudentInfo.StudentID.String
	mapStudentCourseKeyWithStudentPackageAccessPath, err := s.StudentPackageAccessPathRepo.GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs(
		ctx,
		db,
		[]string{studentID},
	)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get map student course key with student package access path by student ids with student_id = %s, err = %v", studentID, err)
		return
	}

	for _, courseItem := range orderItemData.OrderItem.CourseItems {
		var eventMessage *npb.EventStudentPackage
		eventMessage, err = s.upsertStudentPackageDataForNewOrder(ctx, db, orderItemData, courseItem, mapStudentCourseKeyWithStudentPackageAccessPath)
		if err != nil {
			return
		}
		if eventMessage != nil {
			eventMessages = append(eventMessages, eventMessage)
		}
	}
	return
}

func (s *StudentPackageService) upsertStudentPackageDataForNewOrder(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
	courseInfo *pb.CourseItem,
	mapStudentCourseWithStudentPackageAccessPath map[string]entities.StudentPackageAccessPath,
) (
	eventMessage *npb.EventStudentPackage,
	err error,
) {
	var (
		startAt, endAt         time.Time
		action                 = pb.StudentPackageActions_STUDENT_PACKAGE_ACTION_UPSERT.String()
		flow                   = "Upsert student package and student course"
		studentID              = orderItemData.StudentInfo.StudentID.String
		courseID               = courseInfo.CourseId
		studentPackageID       = uuid.NewString()
		studentCourseKey       = fmt.Sprintf("%v_%v", studentID, courseID)
		isUpdateStudentPackage = false
	)

	studentPackageAccessPath, isExistStudentPackage := mapStudentCourseWithStudentPackageAccessPath[studentCourseKey]
	if isExistStudentPackage {
		isUpdateStudentPackage = true
		studentPackageID = studentPackageAccessPath.StudentPackageID.String
	}

	if orderItemData.IsOneTimeProduct {
		startAt = orderItemData.PackageInfo.Package.PackageStartDate.Time
		endAt = orderItemData.PackageInfo.Package.PackageEndDate.Time
	} else {
		startAt = orderItemData.StudentProduct.StartDate.Time
		endAt = orderItemData.StudentProduct.EndDate.Time
	}
	studentPackageOrderPosition, err := s.StudentPackageOrderService.GetPositionForStudentPackageByTime(ctx, db, studentPackageID, startAt, endAt)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get position for student package by time with student_package_id = %s, error = %v", studentPackageID, err)
		return
	}

	studentPackage,
		studentPackageAccessPath,
		studentCourse,
		studentPackageOrder,
		eventMessage := s.convertStudentPackageDataForCreateOrder(ctx, orderItemData, studentPackageID, courseID,
		studentPackageOrderPosition, isUpdateStudentPackage)

	switch studentPackageOrderPosition {
	case entities.PastStudentPackage:
		// Insert a new student package in past
		if !isUpdateStudentPackage {
			err = utils.GroupErrorFunc(
				s.StudentPackageRepo.Upsert(ctx, db, &studentPackage),
				s.StudentCourseRepo.UpsertStudentCourse(ctx, db, studentCourse),
				s.StudentPackageOrderService.InsertStudentPackageOrder(ctx, db, studentPackageOrder, studentPackageOrderPosition),
				s.writeStudentPackageLog(ctx, db, &studentPackage, studentCourse.CourseID.String, action, flow),
			)
			if err != nil {
				return
			}
		}
	case entities.CurrentStudentPackage:
		err = utils.GroupErrorFunc(
			s.StudentPackageRepo.Upsert(ctx, db, &studentPackage),
			s.StudentCourseRepo.UpsertStudentCourse(ctx, db, studentCourse),
			s.StudentPackageOrderService.InsertStudentPackageOrder(ctx, db, studentPackageOrder, studentPackageOrderPosition),
			s.writeStudentPackageLog(ctx, db, &studentPackage, studentCourse.CourseID.String, action, flow),
		)
		if err != nil {
			return
		}
	case entities.FutureStudentPackage:
		err = utils.GroupErrorFunc(
			s.StudentPackageOrderService.InsertStudentPackageOrder(ctx, db, studentPackageOrder, studentPackageOrderPosition),
		)
		if err != nil {
			return
		}
	default:
		err = status.Error(codes.Internal, "error when position package time is invalid")
		return
	}

	if !isUpdateStudentPackage {
		err = s.StudentPackageAccessPathRepo.Insert(ctx, db, &studentPackageAccessPath)
		if err != nil {
			err = status.Errorf(codes.Internal, "error when insert student package access path with err = %v", err)
			return
		}
	}
	return
}

func (s *StudentPackageService) convertStudentPackageDataByStudentPackageOrder(
	ctx context.Context,
	db database.QueryExecer,
	studentPackageOrder entities.StudentPackageOrder,
) (
	studentPackage entities.StudentPackages,
	studentCourse entities.StudentCourse,
	eventMessage *npb.EventStudentPackage,
	err error) {
	var (
		packageProperties = &entities.PackageProperties{}
		packageEntity     entities.Package
		quantityType      pb.QuantityType
	)

	studentPackage, err = studentPackageOrder.GetStudentPackageObject()
	if err != nil {
		return
	}

	packageProperties, err = studentPackage.GetProperties()
	if err != nil {
		return
	}
	courseInfo := packageProperties.AllCourseInfo[0]
	packageEntity, err = s.PackageRepo.GetByID(ctx, db, studentPackage.PackageID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, fmt.Sprintf("Error when get package by package id: %v", err))
		return
	}
	quantityType, err = s.PackageQuantityTypeMappingRepo.GetByPackageTypeForUpdate(ctx, db, packageEntity.PackageType.String)
	if err != nil {
		err = status.Errorf(codes.Internal, fmt.Sprintf("Error when get package quantity type mapping: %v", err))
		return
	}

	studentCourse = entities.StudentCourse{
		StudentPackageID:  pgtype.Text{Status: pgtype.Present, String: studentPackageOrder.StudentPackageID.String},
		StudentID:         studentPackage.StudentID,
		CourseID:          pgtype.Text{Status: pgtype.Present, String: courseInfo.CourseID},
		LocationID:        studentPackage.LocationIDs.Elements[0],
		StudentStartDate:  studentPackageOrder.StartAt,
		StudentEndDate:    studentPackageOrder.EndAt,
		CourseSlot:        pgtype.Int4{Status: pgtype.Null},
		CourseSlotPerWeek: pgtype.Int4{Status: pgtype.Null},
		Weight:            pgtype.Int4{Status: pgtype.Null},
		CreatedAt:         pgtype.Timestamptz{Time: studentPackage.CreatedAt.Time, Status: pgtype.Present},
		UpdatedAt:         pgtype.Timestamptz{Time: studentPackage.UpdatedAt.Time, Status: pgtype.Present},
		DeletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
		PackageType:       packageEntity.PackageType,
		ResourcePath:      pgtype.Text{Status: pgtype.Null},
	}
	switch quantityType {
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

func (s *StudentPackageService) convertStudentPackageDataForCreateOrder(
	ctx context.Context,
	orderItemData utils.OrderItemData,
	studentPackageID, courseID string,
	studentPackageOrderPosition entities.StudentPackagePosition,
	isUpdate bool,
) (
	studentPackage entities.StudentPackages,
	studentPackageAccessPath entities.StudentPackageAccessPath,
	studentCourse entities.StudentCourse,
	studentPackageOrder entities.StudentPackageOrder,
	eventMessage *npb.EventStudentPackage,
) {
	var (
		packageProperties     entities.PackageProperties
		startAt, endAt        pgtype.Timestamptz
		userID                = interceptors.UserIDFromContext(ctx)
		isActive              = false
		isPublishEventMessage = true
	)

	if studentPackageOrderPosition == entities.CurrentStudentPackage {
		isActive = true
	}
	if orderItemData.IsOneTimeProduct {
		startAt = orderItemData.PackageInfo.Package.PackageStartDate
		endAt = orderItemData.PackageInfo.Package.PackageEndDate
	} else {
		startAt = orderItemData.StudentProduct.StartDate
		endAt = orderItemData.StudentProduct.EndDate
	}

	courseInfoProto := orderItemData.PackageInfo.MapCourseInfo[courseID]
	courseInfo := entities.CourseInfo{
		CourseID: courseInfoProto.CourseId,
		Name:     courseInfoProto.CourseName,
	}
	if courseInfoProto.Slot != nil {
		courseInfo.NumberOfSlots = int(courseInfoProto.Slot.Value)
	}
	if courseInfoProto.Weight != nil {
		courseInfo.Weight = int(courseInfoProto.Weight.Value)
	}
	packageProperties = entities.PackageProperties{
		AllCourseInfo:     []entities.CourseInfo{courseInfo},
		CanWatchVideo:     []string{courseID},
		CanDoQuiz:         []string{courseID},
		CanViewStudyGuide: []string{courseID},
	}
	_ = multierr.Combine(
		studentPackage.PackageID.Set(orderItemData.ProductInfo.ProductID),
		studentPackage.StudentID.Set(orderItemData.Order.StudentID),
		studentPackage.LocationIDs.Set([]string{orderItemData.Order.LocationID.String}),
		studentPackage.IsActive.Set(isActive),
		studentPackage.ID.Set(studentPackageID),
		studentPackage.DeletedAt.Set(nil),
		studentPackage.CreatedAt.Set(nil),
		studentPackage.UpdatedAt.Set(nil),
		studentPackage.StartAt.Set(startAt),
		studentPackage.EndAt.Set(endAt),
		studentPackage.Properties.Set(packageProperties),

		studentPackageAccessPath.StudentPackageID.Set(studentPackageID),
		studentPackageAccessPath.StudentID.Set(orderItemData.Order.StudentID),
		studentPackageAccessPath.CourseID.Set(courseID),
		studentPackageAccessPath.LocationID.Set(orderItemData.Order.LocationID),

		studentCourse.StudentPackageID.Set(studentPackageID),
		studentCourse.StudentID.Set(orderItemData.Order.StudentID),
		studentCourse.CourseID.Set(courseID),
		studentCourse.LocationID.Set(orderItemData.Order.LocationID),
		studentCourse.StudentStartDate.Set(startAt),
		studentCourse.StudentEndDate.Set(endAt),
		studentCourse.PackageType.Set(orderItemData.PackageInfo.Package.PackageType),
		studentCourse.CourseSlot.Set(nil),
		studentCourse.CourseSlotPerWeek.Set(nil),
		studentCourse.Weight.Set(nil),
		studentCourse.CreatedAt.Set(nil),
		studentCourse.UpdatedAt.Set(nil),
		studentCourse.DeletedAt.Set(nil),
		studentCourse.ResourcePath.Set(nil),

		studentPackageOrder.ID.Set(uuid.NewString()),
		studentPackageOrder.UserID.Set(userID),
		studentPackageOrder.OrderID.Set(orderItemData.Order.OrderID),
		studentPackageOrder.CourseID.Set(courseID),
		studentPackageOrder.StartAt.Set(startAt),
		studentPackageOrder.EndAt.Set(endAt),
		studentPackageOrder.StudentPackageObject.Set(studentPackage),
		studentPackageOrder.StudentPackageID.Set(studentPackageID),
		studentPackageOrder.IsCurrentStudentPackage.Set(false),
		studentPackageOrder.CreatedAt.Set(time.Now()),
		studentPackageOrder.UpdatedAt.Set(time.Now()),
		studentPackageOrder.DeletedAt.Set(nil),
		studentPackageOrder.FromStudentPackageOrderID.Set(nil),
		studentPackageOrder.IsExecutedByCronJob.Set(false),
		studentPackageOrder.ExecutedError.Set(nil),
	)

	switch orderItemData.PackageInfo.QuantityType {
	case pb.QuantityType_QUANTITY_TYPE_SLOT:
		_ = studentCourse.CourseSlot.Set(courseInfoProto.Slot.Value)
	case pb.QuantityType_QUANTITY_TYPE_SLOT_PER_WEEK:
		_ = studentCourse.CourseSlotPerWeek.Set(courseInfoProto.Slot.Value)
	case pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT:
		_ = studentCourse.Weight.Set(courseInfoProto.Weight.Value)
	}

	switch studentPackageOrderPosition {
	case entities.PastStudentPackage:
		if isUpdate {
			isPublishEventMessage = false
		}
	case entities.CurrentStudentPackage:
		_ = studentPackageOrder.IsCurrentStudentPackage.Set(true)
	case entities.FutureStudentPackage:
		isPublishEventMessage = false
	}

	if !isPublishEventMessage {
		eventMessage = nil
	} else {
		eventMessage = &npb.EventStudentPackage{
			StudentPackage: &npb.EventStudentPackage_StudentPackage{
				StudentId: studentPackage.StudentID.String,
				Package: &npb.EventStudentPackage_Package{
					CourseIds:        []string{courseID},
					StartDate:        timestamppb.New(studentPackage.StartAt.Time),
					EndDate:          timestamppb.New(studentPackage.EndAt.Time),
					LocationIds:      []string{orderItemData.Order.LocationID.String},
					StudentPackageId: studentPackage.ID.String,
				},
				IsActive: true,
			},
			LocationIds: []string{orderItemData.Order.LocationID.String},
		}
	}
	return
}

func (s *StudentPackageService) voidStudentPackageForCreateOrder(
	ctx context.Context,
	db database.QueryExecer,
	args utils.VoidStudentPackageArgs,
	mapStudentCourseKeyWithStudentPackageAccessPath map[string]entities.StudentPackageAccessPath,
) (
	studentPackageEvents []*npb.EventStudentPackage,
	err error,
) {
	mapCourseIDOrderItemCourse, err := s.OrderItemCourseRepo.
		GetMapOrderItemCourseByOrderIDAndPackageID(ctx, db, args.Order.OrderID.String, args.Product.ProductID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get map order item course for void create order with order_id = %s and package_id = %s and error = %v", args.Order.OrderID.String, args.Product.ProductID.String, err)
		return
	}
	studentID := args.Order.StudentID.String
	for courseID := range mapCourseIDOrderItemCourse {
		var event *npb.EventStudentPackage
		studentCourseKey := fmt.Sprintf("%v_%v", studentID, courseID)
		studentPackageAccessPath := mapStudentCourseKeyWithStudentPackageAccessPath[studentCourseKey]
		event, err = s.voidStudentPackageDataForCreateOrder(ctx, db, args, courseID, studentPackageAccessPath)
		if err != nil {
			return
		}
		studentPackageEvents = append(studentPackageEvents, event)
	}
	return
}

func (s *StudentPackageService) voidStudentPackageDataForCreateOrder(
	ctx context.Context,
	db database.QueryExecer,
	args utils.VoidStudentPackageArgs,
	courseID string,
	studentPackageAccessPath entities.StudentPackageAccessPath,
) (
	eventMessage *npb.EventStudentPackage,
	err error,
) {
	var (
		studentPackageOrder           *entities.StudentPackageOrder
		studentPackage                entities.StudentPackages
		newCurrentStudentPackageOrder *entities.StudentPackageOrder
		newStudentPackage             entities.StudentPackages
		studentCourse                 entities.StudentCourse
		flow                          = "Void order for new order flow"
		action                        = pb.StudentProductStatus_CANCELLED.String()
		studentID                     = args.Order.StudentID.String
	)

	studentPackageOrder, err = s.StudentPackageOrderService.GetStudentPackageOrderByStudentPackageIDAndOrderID(ctx, db, studentPackageAccessPath.StudentPackageID.String, args.Order.OrderID.String)
	if err != nil {
		return
	}

	studentPackageOrderPosition, err := s.getPositionOfStudentPackageOrder(ctx, db, *studentPackageOrder)
	if err != nil {
		return
	}

	err = utils.GroupErrorFunc(
		s.StudentPackageOrderService.DeleteStudentPackageOrderByID(ctx, db, studentPackageOrder.ID.String),
	)
	if err != nil {
		return
	}

	switch studentPackageOrderPosition {
	case entities.PastStudentPackage:
		err = status.Errorf(codes.Internal, "error when void student package in past time with student_package_id = %v and student_package_order_id = %v", studentPackageOrder.StudentPackageID.String, studentPackageOrder.ID.String)
		return
	case entities.CurrentStudentPackage:
		studentPackage, err = s.StudentPackageRepo.GetByID(ctx, db, studentPackageAccessPath.StudentPackageID.String)
		if err != nil {
			return
		}

		newCurrentStudentPackageOrder, err = s.StudentPackageOrderService.
			SetCurrentStudentPackageOrderByTimeAndStudentPackageID(ctx, db, studentPackage.ID.String)
		if err != nil {
			return
		}

		// Incase: there is one or more student package orders
		if newCurrentStudentPackageOrder != nil {
			newStudentPackage, studentCourse, eventMessage, err = s.convertStudentPackageDataByStudentPackageOrder(ctx, db, *newCurrentStudentPackageOrder)
			if err != nil {
				return
			}

			err = utils.GroupErrorFunc(
				s.StudentPackageRepo.Upsert(ctx, db, &newStudentPackage),
				s.StudentCourseRepo.UpsertStudentCourse(ctx, db, studentCourse),
				s.writeStudentPackageLog(ctx, db, &studentPackage, courseID, action, flow),
			)
			if err != nil {
				err = status.Errorf(codes.Internal, "error when upsert student package, upscrt student course and write student package log with student_package_id = %v and student_package_order_id = %v and error = %v", studentPackageOrder.StudentPackageID.String, studentPackageOrder.ID.String, err)
				return
			}
		} else {
			err = utils.GroupErrorFunc(
				s.StudentPackageRepo.CancelByID(ctx, db, studentPackageAccessPath.StudentPackageID.String),
				s.StudentCourseRepo.CancelByStudentPackageIDAndCourseID(ctx, db, studentPackageAccessPath.StudentPackageID.String, courseID),
				s.StudentPackageAccessPathRepo.SoftDeleteByStudentPackageIDs(ctx, db, []string{studentPackageAccessPath.StudentPackageID.String}, time.Now()),
				s.writeStudentPackageLog(ctx, db, &studentPackage, courseID, action, flow),
			)
			if err != nil {
				return
			}
			eventMessage = &npb.EventStudentPackage{
				StudentPackage: &npb.EventStudentPackage_StudentPackage{
					StudentId: studentID,
					IsActive:  false,
					Package: &npb.EventStudentPackage_Package{
						StudentPackageId: studentPackageAccessPath.StudentPackageID.String,
						LocationIds:      []string{args.Order.LocationID.String},
						CourseIds: []string{
							courseID,
						},
					},
				},
				LocationIds: []string{args.Order.LocationID.String},
			}
		}
	case entities.FutureStudentPackage:
	}
	return
}

// getPositionOfStudentPackageOrder
// Used to determined position of specific student package order in student package order list
func (s *StudentPackageService) getPositionOfStudentPackageOrder(ctx context.Context,
	db database.QueryExecer,
	studentPackageOrder entities.StudentPackageOrder,
) (position entities.StudentPackagePosition, err error) {
	if studentPackageOrder.IsCurrentStudentPackage.Status == pgtype.Present &&
		studentPackageOrder.IsCurrentStudentPackage.Bool {
		position = entities.CurrentStudentPackage
		return
	}

	if studentPackageOrder.StudentPackageID.String == "" {
		err = status.Errorf(codes.Internal, "error when missing student package id while getting position of student package order")
		return
	}
	currentStudentPackageOrder, err := s.StudentPackageOrderService.GetCurrentStudentPackageOrderByStudentPackageID(ctx, db, studentPackageOrder.StudentPackageID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get current student package order by student package id with student_package_id = %v: %v", studentPackageOrder.StudentPackageID.String, err)
		return
	}

	if studentPackageOrder.EndAt.Status == pgtype.Present &&
		currentStudentPackageOrder.StartAt.Status == pgtype.Present &&
		studentPackageOrder.EndAt.Time.Before(currentStudentPackageOrder.StartAt.Time) {
		position = entities.PastStudentPackage
		return
	}

	if studentPackageOrder.StartAt.Status == pgtype.Present &&
		currentStudentPackageOrder.EndAt.Status == pgtype.Present &&
		studentPackageOrder.StartAt.Time.After(currentStudentPackageOrder.EndAt.Time) {
		position = entities.FutureStudentPackage
		return
	}
	return entities.CurrentStudentPackage, nil
}
