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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *StudentPackageService) MutationStudentPackageForUpdateOrder(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (
	eventMessages []*npb.EventStudentPackage,
	err error,
) {
	var (
		studentID        = orderItemData.StudentInfo.StudentID.String
		studentProductID = orderItemData.OrderItem.StudentProductId.Value
	)

	orderItemOfCreateOrder, err := s.OrderItemRepo.GetOrderItemByStudentProductID(ctx, db, studentProductID)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get order item by student product id with student_product_id = %s and err = %v", studentProductID, err)
		return
	}

	mapCourseIDWithOrderItemCourseOfCreateOrder, err := s.OrderItemCourseRepo.GetMapOrderItemCourseByOrderIDAndPackageID(ctx, db, orderItemOfCreateOrder.OrderID.String, orderItemOfCreateOrder.ProductID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get map order item course by order id and package id with order_id = %s, package_id = %s and err = %v", orderItemOfCreateOrder.OrderID.String, orderItemOfCreateOrder.ProductID.String, err)
		return
	}

	studentProduct, err := s.StudentProductRepo.GetStudentProductForUpdateByStudentProductID(ctx, db, studentProductID)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get student product by student product id with student_product_id = %s and err = %v", studentProductID, err)
		return
	}

	mapStudentCourseKeyWithStudentPackageAccessPath, err := s.StudentPackageAccessPathRepo.
		GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs(ctx, db, []string{studentID})
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get map student course key with student package access path by student ids  with student_id = %s and err = %v", studentID, err)
		return
	}

	isOneTimeProductOrUpdateCompleteOrder := orderItemData.IsOneTimeProduct ||
		studentProduct.StartDate.Time.Equal(utils.StartOfDate(
			orderItemData.OrderItem.EffectiveDate.AsTime(),
			orderItemData.Timezone,
		))

	for i, reqCourseItem := range orderItemData.OrderItem.CourseItems {
		studentCourseKey := fmt.Sprintf("%v_%v", studentID, reqCourseItem.CourseId)
		studentPackageAccessPath, isExistStudentPackageAccessPath := mapStudentCourseKeyWithStudentPackageAccessPath[studentCourseKey]

		var eventMessage *npb.EventStudentPackage
		orderItemCourseOfCreateOrder, isExists := mapCourseIDWithOrderItemCourseOfCreateOrder[reqCourseItem.CourseId]
		// Update existing student package data, that existed in Create Order, updated by Update Order
		if isExists && isExistStudentPackageAccessPath {
			delete(mapCourseIDWithOrderItemCourseOfCreateOrder, reqCourseItem.CourseId)
			if reqCourseItem.Slot == nil || reqCourseItem.Slot.Value == orderItemCourseOfCreateOrder.CourseSlot.Int {
				continue
			}

			if isOneTimeProductOrUpdateCompleteOrder {
				eventMessage, err = s.updateStudentPackageDataForCompleteUpdateOrder(ctx, db, orderItemData,
					orderItemData.OrderItem.CourseItems[i],
					studentPackageAccessPath,
				)
				if err != nil {
					err = status.Errorf(codes.Internal, "error when upsert student package and student course for complete update order: %s", err.Error())
					return
				}
				if eventMessage != nil {
					eventMessages = append(eventMessages, eventMessage)
				}
			} else {
				eventMessage, err = s.updateStudentPackageDataForNonCompleteUpdateOrder(ctx, db, orderItemData,
					reqCourseItem.CourseId,
					mapStudentCourseKeyWithStudentPackageAccessPath,
				)
				if err != nil {
					return
				}
				if eventMessage != nil {
					eventMessages = append(eventMessages, eventMessage)
				}
			}
		} else {
			// Insert new student package data, that have not exists in Create Order, added in Update Order
			eventMessage, err = s.upsertStudentPackageDataForNewOrder(ctx, db, orderItemData,
				orderItemData.OrderItem.CourseItems[i],
				mapStudentCourseKeyWithStudentPackageAccessPath,
			)
			if err != nil {
				return
			}
			if eventMessage != nil {
				eventMessages = append(eventMessages, eventMessage)
			}
		}
	}

	if len(mapCourseIDWithOrderItemCourseOfCreateOrder) == 0 {
		return
	}

	// Cancel/Delete existing student package data, that existed in Create Order, deleted by Update Order
	if isOneTimeProductOrUpdateCompleteOrder {
		for courseID := range mapCourseIDWithOrderItemCourseOfCreateOrder {
			var eventMessage *npb.EventStudentPackage
			eventMessage, err = s.cancelStudentPackageDataForCompleteCancelOrder(ctx, db, orderItemData, courseID, mapStudentCourseKeyWithStudentPackageAccessPath)
			if err != nil {
				return
			}
			if eventMessage != nil {
				eventMessages = append(eventMessages, eventMessage)
			}
		}
	} else {
		for courseID := range mapCourseIDWithOrderItemCourseOfCreateOrder {
			var eventMessage *npb.EventStudentPackage
			eventMessage, err = s.cancelStudentPackageDataForNonCompleteUpdateOrder(ctx, db, orderItemData, courseID, mapStudentCourseKeyWithStudentPackageAccessPath)
			if err != nil {
				return
			}
			if eventMessage != nil {
				eventMessages = append(eventMessages, eventMessage)
			}
		}
	}

	return
}

func (s *StudentPackageService) updateStudentPackageDataForCompleteUpdateOrder(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
	reqCourseItem *pb.CourseItem,
	studentPackageAccessPath entities.StudentPackageAccessPath,
) (
	eventMessage *npb.EventStudentPackage,
	err error,
) {
	var (
		studentPackageOrderForCreateOrder *entities.StudentPackageOrder
		courseID                          = reqCourseItem.CourseId
		effectiveDate                     = getStartTimeFromOrder(orderItemData.OrderItem)
		studentPackageID                  = studentPackageAccessPath.StudentPackageID.String
		action                            = pb.StudentPackageActions_STUDENT_PACKAGE_ACTION_UPSERT.String()
		flow                              = "upsert student package and student course"
		studentPackageOrderPosition       entities.StudentPackagePosition
	)
	studentPackageOrderForCreateOrder, err = s.StudentPackageOrderService.GetStudentPackageOrderByStudentPackageIDAndTime(ctx, db,
		studentPackageID, utils.ConvertToLocalTime(effectiveDate, orderItemData.Timezone))
	if err != nil {
		err = status.Errorf(codes.Internal, "error when getting student package order by student package id and time with student_package_id = %s and err = %v", studentPackageID, err.Error())
		return
	}

	// This check is used for invalid data (exists before student_package_order) => missing student_package_order for this student_course
	if studentPackageOrderForCreateOrder == nil {
		err = status.Errorf(codes.FailedPrecondition, "missing student package order with student package id = %s and effective date = %v", studentPackageID, effectiveDate)
		return
	}
	studentCourseAfterUpdate,
		studentPackageAfterUpdate,
		studentPackageAccessPath,
		studentPackageOrderForUpdateOrder,
		eventMessage,
		err := convertStudentPackageDataForCompleteUpdateOrder(ctx, orderItemData, studentPackageID,
		courseID, studentPackageOrderForCreateOrder)
	if err != nil {
		return
	}

	studentPackageOrderPosition, err = s.getPositionOfStudentPackageOrder(ctx, db, *studentPackageOrderForCreateOrder)
	if err != nil {
		return
	}

	switch studentPackageOrderPosition {
	case entities.PastStudentPackage:
	case entities.CurrentStudentPackage:
		err = utils.GroupErrorFunc(
			s.StudentPackageRepo.Upsert(ctx, db, &studentPackageAfterUpdate),
			s.StudentCourseRepo.UpsertStudentCourse(ctx, db, studentCourseAfterUpdate),
			s.writeStudentPackageLog(ctx, db, &studentPackageAfterUpdate, studentCourseAfterUpdate.CourseID.String, action, flow),

			s.StudentPackageOrderService.DeleteStudentPackageOrderByID(ctx, db, studentPackageOrderForCreateOrder.ID.String),
			s.StudentPackageOrderService.InsertStudentPackageOrder(ctx, db, studentPackageOrderForUpdateOrder, entities.CurrentStudentPackage),
		)
		if err != nil {
			return
		}
	case entities.FutureStudentPackage:
		err = utils.GroupErrorFunc(
			s.StudentPackageOrderService.DeleteStudentPackageOrderByID(ctx, db, studentPackageOrderForCreateOrder.ID.String),
			s.StudentPackageOrderService.InsertStudentPackageOrder(ctx, db, studentPackageOrderForUpdateOrder, entities.FutureStudentPackage),
		)
		if err != nil {
			return
		}
		eventMessage = nil
	}
	return
}

func convertStudentPackageDataForCompleteUpdateOrder(
	ctx context.Context,
	orderItemData utils.OrderItemData,
	studentPackageID, courseID string,
	studentPackageOrderOfCreateOrder *entities.StudentPackageOrder,
) (
	studentCourseAfterUpdate entities.StudentCourse,
	studentPackageAfterUpdate entities.StudentPackages,
	studentPackageAccessPath entities.StudentPackageAccessPath,
	studentPackageOrderForUpdateOrder entities.StudentPackageOrder,
	eventMessage *npb.EventStudentPackage,
	err error,
) {
	var (
		packageProperty entities.PackageProperties
	)
	err = utils.GroupErrorFunc(
		studentPackageAfterUpdate.PackageID.Set(orderItemData.ProductInfo.ProductID),
		studentPackageAfterUpdate.StudentID.Set(orderItemData.Order.StudentID),
		studentPackageAfterUpdate.LocationIDs.Set([]string{orderItemData.Order.LocationID.String}),
		studentPackageAfterUpdate.IsActive.Set(true),
		studentPackageAfterUpdate.ID.Set(studentPackageID),
		studentPackageAfterUpdate.DeletedAt.Set(nil),
		studentPackageAfterUpdate.CreatedAt.Set(nil),
		studentPackageAfterUpdate.UpdatedAt.Set(nil),

		studentPackageAccessPath.StudentPackageID.Set(studentPackageID),
		studentPackageAccessPath.StudentID.Set(orderItemData.Order.StudentID),
		studentPackageAccessPath.CourseID.Set(courseID),
		studentPackageAccessPath.LocationID.Set(orderItemData.Order.LocationID),
	)
	if err != nil {
		return
	}
	if orderItemData.IsOneTimeProduct {
		_ = studentPackageAfterUpdate.StartAt.Set(orderItemData.PackageInfo.Package.PackageStartDate.Time)
		_ = studentPackageAfterUpdate.EndAt.Set(orderItemData.PackageInfo.Package.PackageEndDate.Time)
	} else {
		_ = studentPackageAfterUpdate.StartAt.Set(orderItemData.StudentProduct.StartDate.Time)
		_ = studentPackageAfterUpdate.EndAt.Set(orderItemData.StudentProduct.EndDate.Time)
	}

	courseInfoProto := orderItemData.PackageInfo.MapCourseInfo[courseID]
	courseInfo := entities.CourseInfo{
		CourseID: courseInfoProto.CourseId,
		Name:     courseInfoProto.CourseName,
	}
	studentCourseAfterUpdate = entities.StudentCourse{
		StudentPackageID:  pgtype.Text{Status: pgtype.Present, String: studentPackageID},
		StudentID:         orderItemData.StudentInfo.StudentID,
		CourseID:          pgtype.Text{Status: pgtype.Present, String: courseID},
		LocationID:        orderItemData.Order.LocationID,
		StudentStartDate:  studentPackageAfterUpdate.StartAt,
		StudentEndDate:    studentPackageAfterUpdate.EndAt,
		CourseSlot:        pgtype.Int4{Status: pgtype.Null},
		CourseSlotPerWeek: pgtype.Int4{Status: pgtype.Null},
		Weight:            pgtype.Int4{Status: pgtype.Null},
		CreatedAt:         pgtype.Timestamptz{Status: pgtype.Null},
		UpdatedAt:         pgtype.Timestamptz{Status: pgtype.Null},
		DeletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
		PackageType:       orderItemData.PackageInfo.Package.PackageType,
		ResourcePath:      pgtype.Text{Status: pgtype.Null},
	}
	if courseInfoProto.Slot != nil {
		courseInfo.NumberOfSlots = int(courseInfoProto.Slot.Value)
	}
	if courseInfoProto.Weight != nil {
		courseInfo.Weight = int(courseInfoProto.Weight.Value)
	}
	switch orderItemData.PackageInfo.QuantityType {
	case pb.QuantityType_QUANTITY_TYPE_SLOT:
		_ = studentCourseAfterUpdate.CourseSlot.Set(courseInfoProto.Slot.Value)
	case pb.QuantityType_QUANTITY_TYPE_SLOT_PER_WEEK:
		_ = studentCourseAfterUpdate.CourseSlotPerWeek.Set(courseInfoProto.Slot.Value)
	case pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT:
		_ = studentCourseAfterUpdate.Weight.Set(courseInfoProto.Weight.Value)
	}
	packageProperty = entities.PackageProperties{
		AllCourseInfo:     []entities.CourseInfo{courseInfo},
		CanWatchVideo:     []string{courseID},
		CanDoQuiz:         []string{courseID},
		CanViewStudyGuide: []string{courseID},
	}
	_ = studentPackageAfterUpdate.Properties.Set(packageProperty)

	// Convert for student package order
	userID := interceptors.UserIDFromContext(ctx)
	studentPackageOrderForUpdateOrder = entities.StudentPackageOrder{
		ID:                        pgtype.Text{Status: pgtype.Present, String: uuid.NewString()},
		UserID:                    pgtype.Text{Status: pgtype.Present, String: userID},
		OrderID:                   orderItemData.Order.OrderID,
		CourseID:                  studentCourseAfterUpdate.CourseID,
		StartAt:                   studentPackageAfterUpdate.StartAt,
		EndAt:                     studentPackageAfterUpdate.EndAt,
		StudentPackageObject:      pgtype.JSONB{Status: pgtype.Null},
		StudentPackageID:          studentPackageAfterUpdate.ID,
		IsCurrentStudentPackage:   pgtype.Bool{Status: pgtype.Present, Bool: true},
		FromStudentPackageOrderID: pgtype.Text{Status: pgtype.Null},
		ExecutedError:             pgtype.Text{Status: pgtype.Null},
		IsExecutedByCronJob:       pgtype.Bool{Status: pgtype.Present, Bool: false},
		CreatedAt:                 pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now()},
		UpdatedAt:                 pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now()},
		DeletedAt:                 pgtype.Timestamptz{Status: pgtype.Null},
	}
	_ = studentPackageOrderForUpdateOrder.StudentPackageObject.Set(studentPackageAfterUpdate)
	if studentPackageOrderOfCreateOrder != nil && studentPackageOrderOfCreateOrder.ID.Status == pgtype.Present {
		_ = studentPackageOrderForUpdateOrder.FromStudentPackageOrderID.Set(studentPackageOrderOfCreateOrder.ID)
	}
	eventMessage = &npb.EventStudentPackage{
		StudentPackage: &npb.EventStudentPackage_StudentPackage{
			StudentId: studentPackageAfterUpdate.StudentID.String,
			Package: &npb.EventStudentPackage_Package{
				CourseIds:        []string{courseID},
				StartDate:        timestamppb.New(studentPackageAfterUpdate.StartAt.Time),
				EndDate:          timestamppb.New(studentPackageAfterUpdate.EndAt.Time),
				LocationIds:      []string{orderItemData.Order.LocationID.String},
				StudentPackageId: studentPackageAfterUpdate.ID.String,
			},
			IsActive: true,
		},
		LocationIds: []string{orderItemData.Order.LocationID.String},
	}

	isCurrentStudentPackageOrder := studentPackageOrderOfCreateOrder != nil && studentPackageOrderOfCreateOrder.IsCurrentStudentPackage.Status == pgtype.Present && studentPackageOrderOfCreateOrder.IsCurrentStudentPackage.Bool
	if !isCurrentStudentPackageOrder {
		err = utils.GroupErrorFunc(
			studentPackageOrderForUpdateOrder.IsCurrentStudentPackage.Set(false),
		)
		if err != nil {
			return
		}
	}
	return
}

func (s *StudentPackageService) convertStudentPackageDataForNonCompleteUpdateOrder(
	orderItemData utils.OrderItemData,
	courseID string,
	studentPackageOrderBeforeUpdate entities.StudentPackageOrder,
) (
	studentPackageAfterUpdate entities.StudentPackages,
	studentCourseAfterUpdate entities.StudentCourse,
	oldStudentPackageOrderAfterUpdate entities.StudentPackageOrder,
	newStudentPackageOrderAfterUpdate entities.StudentPackageOrder,
	eventMessage *npb.EventStudentPackage,
	err error,
) {
	var (
		packageProperty              entities.PackageProperties
		endDate                      = utils.EndOfDate(getStartTimeFromOrder(orderItemData.OrderItem).AddDate(0, 0, -1), orderItemData.Timezone)
		newStudentPackageAfterUpdate entities.StudentPackages
	)

	studentCourseBeforeUpdate, studentPackageBeforeUpdate, _, err := s.convertStudentPackageDataForCancelOrder(
		orderItemData.PackageInfo.Package.PackageType.String,
		orderItemData.PackageInfo.QuantityType.String(),
		studentPackageOrderBeforeUpdate)
	if err != nil {
		return
	}
	studentCourseAfterUpdate = *studentCourseBeforeUpdate
	_ = studentCourseAfterUpdate.StudentEndDate.Set(endDate)
	studentPackageAfterUpdate = studentPackageBeforeUpdate
	_ = studentPackageAfterUpdate.EndAt.Set(endDate)

	oldStudentPackageOrderAfterUpdate = studentPackageOrderBeforeUpdate
	_ = oldStudentPackageOrderAfterUpdate.StudentPackageObject.Set(studentPackageAfterUpdate)
	_ = oldStudentPackageOrderAfterUpdate.EndAt.Set(endDate)

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
	newStudentCourseAfterUpdate := entities.StudentCourse{
		StudentPackageID:  pgtype.Text{Status: pgtype.Present, String: studentPackageBeforeUpdate.ID.String},
		StudentID:         orderItemData.StudentInfo.StudentID,
		CourseID:          pgtype.Text{Status: pgtype.Present, String: courseID},
		LocationID:        orderItemData.Order.LocationID,
		StudentStartDate:  pgtype.Timestamptz{Time: utils.StartOfDate(orderItemData.OrderItem.EffectiveDate.AsTime(), orderItemData.Timezone), Status: pgtype.Present},
		StudentEndDate:    studentPackageBeforeUpdate.EndAt,
		CourseSlot:        pgtype.Int4{Status: pgtype.Null},
		CourseSlotPerWeek: pgtype.Int4{Status: pgtype.Null},
		Weight:            pgtype.Int4{Status: pgtype.Null},
		CreatedAt:         pgtype.Timestamptz{Status: pgtype.Null},
		UpdatedAt:         pgtype.Timestamptz{Status: pgtype.Null},
		DeletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
		PackageType:       orderItemData.PackageInfo.Package.PackageType,
		ResourcePath:      pgtype.Text{Status: pgtype.Null},
	}
	switch orderItemData.PackageInfo.QuantityType {
	case pb.QuantityType_QUANTITY_TYPE_SLOT:
		_ = newStudentCourseAfterUpdate.CourseSlot.Set(courseInfoProto.Slot.Value)
	case pb.QuantityType_QUANTITY_TYPE_SLOT_PER_WEEK:
		_ = newStudentCourseAfterUpdate.CourseSlotPerWeek.Set(courseInfoProto.Slot.Value)
	case pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT:
		_ = newStudentCourseAfterUpdate.Weight.Set(courseInfoProto.Weight.Value)
	}
	packageProperty = entities.PackageProperties{
		AllCourseInfo:     []entities.CourseInfo{courseInfo},
		CanWatchVideo:     []string{courseID},
		CanDoQuiz:         []string{courseID},
		CanViewStudyGuide: []string{courseID},
	}
	err = utils.GroupErrorFunc(
		newStudentPackageAfterUpdate.ID.Set(studentPackageBeforeUpdate.ID),
		newStudentPackageAfterUpdate.PackageID.Set(orderItemData.ProductInfo.ProductID),
		newStudentPackageAfterUpdate.StudentID.Set(orderItemData.Order.StudentID),
		newStudentPackageAfterUpdate.LocationIDs.Set([]string{orderItemData.Order.LocationID.String}),
		newStudentPackageAfterUpdate.IsActive.Set(false),
		newStudentPackageAfterUpdate.CreatedAt.Set(pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present}),
		newStudentPackageAfterUpdate.UpdatedAt.Set(pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present}),
		newStudentPackageAfterUpdate.DeletedAt.Set(nil),
		newStudentPackageAfterUpdate.StartAt.Set(pgtype.Timestamptz{Time: utils.StartOfDate(orderItemData.OrderItem.EffectiveDate.AsTime(), orderItemData.Timezone), Status: pgtype.Present}),
		newStudentPackageAfterUpdate.EndAt.Set(studentPackageBeforeUpdate.EndAt),
		newStudentPackageAfterUpdate.Properties.Set(packageProperty),

		newStudentPackageOrderAfterUpdate.ID.Set(pgtype.Text{String: uuid.NewString(), Status: pgtype.Present}),
		newStudentPackageOrderAfterUpdate.UserID.Set(studentPackageOrderBeforeUpdate.UserID),
		newStudentPackageOrderAfterUpdate.OrderID.Set(orderItemData.Order.OrderID),
		newStudentPackageOrderAfterUpdate.CourseID.Set(pgtype.Text{String: courseID, Status: pgtype.Present}),
		newStudentPackageOrderAfterUpdate.StartAt.Set(pgtype.Timestamptz{Time: utils.StartOfDate(orderItemData.OrderItem.EffectiveDate.AsTime(), orderItemData.Timezone), Status: pgtype.Present}),
		newStudentPackageOrderAfterUpdate.EndAt.Set(studentPackageOrderBeforeUpdate.EndAt),
		newStudentPackageOrderAfterUpdate.StudentPackageObject.Set(newStudentPackageAfterUpdate),
		newStudentPackageOrderAfterUpdate.StudentPackageID.Set(studentPackageBeforeUpdate.ID),
		newStudentPackageOrderAfterUpdate.IsCurrentStudentPackage.Set(pgtype.Bool{Bool: false, Status: pgtype.Present}),
		newStudentPackageOrderAfterUpdate.FromStudentPackageOrderID.Set(studentPackageOrderBeforeUpdate.ID),
		newStudentPackageOrderAfterUpdate.IsExecutedByCronJob.Set(false),
		newStudentPackageOrderAfterUpdate.ExecutedError.Set(nil),
		newStudentPackageOrderAfterUpdate.CreatedAt.Set(pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present}),
		newStudentPackageOrderAfterUpdate.UpdatedAt.Set(pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present}),
		newStudentPackageOrderAfterUpdate.DeletedAt.Set(nil),

		newStudentPackageAfterUpdate.CreatedAt.Set(pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present}),
		newStudentPackageAfterUpdate.UpdatedAt.Set(pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present}),
		newStudentPackageAfterUpdate.DeletedAt.Set(nil),
	)
	if err != nil {
		return
	}
	eventMessage = &npb.EventStudentPackage{
		StudentPackage: &npb.EventStudentPackage_StudentPackage{
			StudentId: studentPackageBeforeUpdate.StudentID.String,
			Package: &npb.EventStudentPackage_Package{
				CourseIds:        []string{courseID},
				StartDate:        timestamppb.New(oldStudentPackageOrderAfterUpdate.StartAt.Time),
				EndDate:          timestamppb.New(oldStudentPackageOrderAfterUpdate.EndAt.Time),
				LocationIds:      []string{orderItemData.Order.LocationID.String},
				StudentPackageId: studentPackageBeforeUpdate.ID.String,
			},
			IsActive: true,
		},
		LocationIds: []string{orderItemData.Order.LocationID.String},
	}
	return
}

func (s *StudentPackageService) updateStudentPackageDataForNonCompleteUpdateOrder(
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
		action                            = pb.StudentPackageActions_STUDENT_PACKAGE_ACTION_UPDATE.String()
		flow                              = "Update time"
		studentID                         = orderItemData.StudentInfo.StudentID.String
		studentCourseKey                  = fmt.Sprintf("%v_%v", studentID, courseID)
		currentStudentPackageAccessPath   = mapStudentCourseWithStudentPackageAccessPath[studentCourseKey]
		studentPackageOrderPosition       entities.StudentPackagePosition
		studentPackageOrderForCreateOrder *entities.StudentPackageOrder
	)
	studentPackageOrderForCreateOrder, err = s.StudentPackageOrderService.GetStudentPackageOrderByStudentPackageIDAndTime(
		ctx, db, currentStudentPackageAccessPath.StudentPackageID.String,
		utils.ConvertToLocalTime(getStartTimeFromOrder(orderItemData.OrderItem), orderItemData.Timezone))
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get student package order by id and time with student_package_id = %s and error = %s", currentStudentPackageAccessPath.StudentPackageID.String, err)
		return
	}

	// This check is used for invalid data (exists before student_package_order) => missing student_package_order for this student_course
	if studentPackageOrderForCreateOrder == nil {
		err = status.Errorf(codes.FailedPrecondition, "missing student package order with student package id = %s and effective date = %v", currentStudentPackageAccessPath.StudentPackageID.String, getStartTimeFromOrder(orderItemData.OrderItem))
		return
	}
	studentPackageOrderPosition, err = s.getPositionOfStudentPackageOrder(ctx, db, *studentPackageOrderForCreateOrder)
	if err != nil {
		return
	}

	studentPackage, err := s.StudentPackageRepo.GetByID(ctx, db, studentPackageOrderForCreateOrder.StudentPackageID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get student package by id with student_package_id = %s and error = %s", currentStudentPackageAccessPath.StudentPackageID.String, err)
		return
	}

	studentPackageAfterUpdate,
		studentCourseAfterUpdate,
		oldStudentPackageOrderAfterUpdate,
		newStudentPackageOrderAfterUpdate,
		eventMessage, err := s.convertStudentPackageDataForNonCompleteUpdateOrder(orderItemData, courseID, *studentPackageOrderForCreateOrder)

	switch studentPackageOrderPosition {
	case entities.PastStudentPackage:
	case entities.CurrentStudentPackage:
		err = utils.GroupErrorFunc(
			s.StudentPackageRepo.Update(ctx, db, &studentPackageAfterUpdate),
			s.StudentCourseRepo.UpsertStudentCourse(ctx, db, studentCourseAfterUpdate),
			s.StudentPackageOrderService.UpdateStudentPackageOrder(ctx, db, oldStudentPackageOrderAfterUpdate),
			s.StudentPackageOrderService.InsertStudentPackageOrder(ctx, db, newStudentPackageOrderAfterUpdate, entities.CurrentStudentPackage),
			s.writeStudentPackageLog(ctx, db, &studentPackage, currentStudentPackageAccessPath.CourseID.String, action, flow),
		)
		if err != nil {
			return
		}
	case entities.FutureStudentPackage:
		eventMessage = nil
		err = utils.GroupErrorFunc(
			s.StudentPackageOrderService.UpdateStudentPackageOrder(ctx, db, oldStudentPackageOrderAfterUpdate),
			s.StudentPackageOrderService.InsertStudentPackageOrder(ctx, db, newStudentPackageOrderAfterUpdate, entities.FutureStudentPackage),
			s.writeStudentPackageLog(ctx, db, &studentPackage, currentStudentPackageAccessPath.CourseID.String, action, flow),
		)
		if err != nil {
			return
		}
	}
	return
}

func (s *StudentPackageService) voidStudentPackageForUpdateOrder(
	ctx context.Context,
	db database.QueryExecer,
	args utils.VoidStudentPackageArgs,
	mapStudentCourseKeyWithStudentPackageAccessPath map[string]entities.StudentPackageAccessPath,
) (
	studentPackageEvents []*npb.EventStudentPackage,
	err error,
) {
	studentPackageEvents = []*npb.EventStudentPackage{}
	var (
		mapCourseIDWithOrderItemCourseOfUpdateOrder map[string]entities.OrderItemCourse
		mapCourseIDWithOrderItemCourseOfCreateOrder map[string]entities.OrderItemCourse
		orderItemOfCreateOrder                      entities.OrderItem
		studentID                                   = args.Order.StudentID.String
	)
	mapCourseIDWithOrderItemCourseOfUpdateOrder, err = s.OrderItemCourseRepo.GetMapOrderItemCourseByOrderIDAndPackageID(ctx, db, args.Order.OrderID.String, args.Product.ProductID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get map order item course by order id and package id with order_id = %s, product_id = %s and err = %v", args.Order.OrderID.String, args.Product.ProductID.String, err)
		return
	}

	orderItemOfCreateOrder, err = s.OrderItemRepo.GetOrderItemByStudentProductID(ctx, db, args.StudentProduct.StudentProductID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get order item by student product id with student_product_id = %s and err = %v", args.StudentProduct.StudentProductID.String, err)
		return
	}

	mapCourseIDWithOrderItemCourseOfCreateOrder, err = s.OrderItemCourseRepo.GetMapOrderItemCourseByOrderIDAndPackageID(ctx, db, orderItemOfCreateOrder.OrderID.String, args.Product.ProductID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get map order item course by order id and package id with order_id = %s, package_id = %s and err = %v", orderItemOfCreateOrder.OrderID.String, args.Product.ProductID.String, err)
		return
	}

	for courseIDOfUpdateOrder, orderItemCourseOfUpdateOrder := range mapCourseIDWithOrderItemCourseOfUpdateOrder {
		var (
			tmpEvent                 *npb.EventStudentPackage
			studentCourseKey         = fmt.Sprintf("%v_%v", studentID, courseIDOfUpdateOrder)
			studentPackageAccessPath = mapStudentCourseKeyWithStudentPackageAccessPath[studentCourseKey]
		)
		orderItemCourseOfCreateOrder, exists := mapCourseIDWithOrderItemCourseOfCreateOrder[courseIDOfUpdateOrder]

		// To remove student package data, that inserted in Update Order
		if !exists {
			tmpEvent, err = s.voidStudentPackageDataForCreateOrder(ctx, db, args, courseIDOfUpdateOrder, studentPackageAccessPath)
			if err != nil {
				return
			}
			if tmpEvent != nil {
				studentPackageEvents = append(studentPackageEvents, tmpEvent)
			}
		} else {
			// To revert student package data, that updated in Update Order
			delete(mapCourseIDWithOrderItemCourseOfCreateOrder, courseIDOfUpdateOrder)
			if orderItemCourseOfUpdateOrder.CourseSlot.Status == pgtype.Null ||
				orderItemCourseOfUpdateOrder.CourseSlot.Int == orderItemCourseOfCreateOrder.CourseSlot.Int {
				continue
			}
			tmpEvent, err = s.voidStudentPackageDataForUpdateOrder(ctx, db, args, orderItemCourseOfCreateOrder, studentPackageAccessPath)
			if err != nil {
				return
			}
			if tmpEvent != nil {
				studentPackageEvents = append(studentPackageEvents, tmpEvent)
			}
		}
	}

	// To Create/Insert student package data, that removed in Update Order
	for courseID, orderItemCourseOfCreateOrder := range mapCourseIDWithOrderItemCourseOfCreateOrder {
		var tmpEvent *npb.EventStudentPackage
		key := fmt.Sprintf("%v_%v", studentID, courseID)
		studentPackageAccessPath := mapStudentCourseKeyWithStudentPackageAccessPath[key]
		tmpEvent, err = s.voidStudentPackageDataForCancelOrder(ctx, db, args, orderItemCourseOfCreateOrder, studentPackageAccessPath)
		if err != nil {
			return
		}
		if tmpEvent != nil {
			studentPackageEvents = append(studentPackageEvents, tmpEvent)
		}
	}

	return
}

func (s *StudentPackageService) voidStudentPackageDataForUpdateOrder(
	ctx context.Context,
	db database.QueryExecer,
	args utils.VoidStudentPackageArgs,
	_ entities.OrderItemCourse,
	studentPackageAccessPath entities.StudentPackageAccessPath,
) (
	eventMessage *npb.EventStudentPackage,
	err error,
) {
	var (
		studentPackageOrderOfCreateOrder,
		studentPackageOrderForUpdateOrder *entities.StudentPackageOrder
		studentPackageID = studentPackageAccessPath.StudentPackageID.String
		studentPackage   entities.StudentPackages
		studentCourse    = &entities.StudentCourse{}
		action           = pb.StudentPackageActions_STUDENT_PACKAGE_ACTION_UPSERT.String()
		flow             = "Void cancel order flow"
	)
	studentPackageOrderForUpdateOrder, err = s.StudentPackageOrderService.GetStudentPackageOrderByStudentPackageIDAndOrderID(ctx, db, studentPackageID, args.Order.OrderID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get student package order by student_package_id and order_id with student_package_id = %s, order_id = %s and err = %v", studentPackageID, args.Order.OrderID.String, err)
		return
	}

	studentPackage, err = s.StudentPackageRepo.GetByID(ctx, db, studentPackageID)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get student package by id with student_package_id = %s and err = %v", studentPackageID, err)
		return
	}

	productInfo, err := s.ProductRepo.GetByIDForUpdate(ctx, db, studentPackage.PackageID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get product by id with product_id = %s and err = %v", studentPackage.PackageID.String, err)
		return
	}

	err = s.StudentPackageOrderService.DeleteStudentPackageOrderByID(ctx, db, studentPackageOrderForUpdateOrder.ID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when soft delete student package order by id with student_package_order_id = %s and err = %v", studentPackageOrderForUpdateOrder.ID.String, err)
		return
	}

	err = s.StudentPackageOrderService.RevertStudentPackageOrderByStudentPackageOrderID(ctx, db,
		studentPackageOrderForUpdateOrder.FromStudentPackageOrderID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when revert student package order by id with student_package_order_id = %s and err = %v", studentPackageOrderForUpdateOrder.FromStudentPackageOrderID.String, err)
		return
	}

	studentPackageOrderOfCreateOrder, err = s.StudentPackageOrderService.GetStudentPackageOrderByStudentPackageOrderID(ctx, db,
		studentPackageOrderForUpdateOrder.FromStudentPackageOrderID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get student package order by id with student_package_order_id = %s and err = %v", studentPackageOrderForUpdateOrder.FromStudentPackageOrderID.String, err)
		return
	}

	_, err = s.StudentPackageOrderService.SetCurrentStudentPackageOrderByTimeAndStudentPackageID(ctx, db, studentPackageID)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when set current student package order by time and student package id with student_package_id = %s and err = %v", studentPackageID, err)
		return
	}
	position, err := s.getPositionOfStudentPackageOrder(ctx, db, *studentPackageOrderOfCreateOrder)
	if err != nil {
		return
	}

	isNonCompleteUpdateOrder := !studentPackageOrderOfCreateOrder.StartAt.Time.Equal(studentPackageOrderForUpdateOrder.StartAt.Time)
	isRecurringProduct := productInfo.BillingScheduleID.Status == pgtype.Present

	// In case: Void an order with recurring product or void a non-complete Update Order
	if isNonCompleteUpdateOrder || isRecurringProduct {
		var studentPackageObject entities.StudentPackages
		studentPackageObject, err = studentPackageOrderOfCreateOrder.GetStudentPackageObject()
		if err != nil {
			return
		}

		err = utils.GroupErrorFunc(
			studentPackageOrderOfCreateOrder.EndAt.Set(studentPackageOrderForUpdateOrder.EndAt),
			studentPackageObject.EndAt.Set(studentPackageOrderForUpdateOrder.EndAt),
			studentPackageOrderOfCreateOrder.StudentPackageObject.Set(studentPackageObject),
		)
		if err != nil {
			return
		}
	}

	studentPackage, *studentCourse, eventMessage, err = s.convertStudentPackageDataByStudentPackageOrder(ctx, db, *studentPackageOrderOfCreateOrder)
	if err != nil {
		return
	}
	switch position {
	case entities.PastStudentPackage:
		err = status.Errorf(codes.Internal, "error when void student package in past time with student_package_id = %v and student_package_order_id = %v", studentPackageOrderForUpdateOrder.StudentPackageID.String, studentPackageOrderForUpdateOrder.ID.String)
		return
	case entities.CurrentStudentPackage:
		err = utils.GroupErrorFunc(
			s.StudentPackageRepo.Upsert(ctx, db, &studentPackage),
			s.StudentCourseRepo.UpsertStudentCourse(ctx, db, *studentCourse),
			s.StudentPackageOrderService.UpdateStudentPackageOrder(ctx, db, *studentPackageOrderOfCreateOrder),
			s.writeStudentPackageLog(ctx, db, &studentPackage, studentCourse.CourseID.String, action, flow),
		)
		if err != nil {
			return
		}
	case entities.FutureStudentPackage:
	}
	return
}
