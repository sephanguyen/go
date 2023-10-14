package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/kafka"
	"github.com/manabie-com/backend/internal/golibs/kafka/payload"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/repositories"
	"github.com/manabie-com/backend/internal/payment/utils"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type SubscriptionService struct {
	JSM    nats.JetStreamManagement
	DB     database.Ext
	Kafka  kafka.KafkaManagement
	Config configs.CommonConfig

	StudentPackageClassRepo interface {
		GetByStudentPackageID(ctx context.Context, db database.QueryExecer, studentPackageID string) (studentPackageClass entities.StudentPackageClass, err error)
	}
	NotificationDateRepo interface {
		GetByOrderType(
			ctx context.Context,
			db database.QueryExecer,
			orderType string,
		) (notificationDate entities.NotificationDate, err error)
	}
	userRepo interface {
		GetUserIDsByRoleNamesAndLocationID(ctx context.Context, db database.QueryExecer, roleNames []string, locationID string,
		) (userIDs []string, err error)
	}

	orderRepo interface {
		GetLatestOrderByStudentIDAndLocationIDAndOrderType(
			ctx context.Context,
			db database.QueryExecer,
			studentID, locationID, orderType string,
		) (order entities.Order, err error)
		GetLatestOrderByStudentIDAndLocationID(
			ctx context.Context,
			db database.QueryExecer,
			studentID, locationID string,
		) (order entities.Order, err error)
	}
	gradeRepo interface {
		GetByID(ctx context.Context, db database.QueryExecer, gradeID string) (entities.Grade, error)
	}
}

func (s *SubscriptionService) Publish(ctx context.Context, tx database.QueryExecer, message utils.MessageSyncData) (err error) {
	//if len(message.StudentCourseMessage) > 0 {
	//	err = s.publishStudentCourseSyncEvent(ctx, message.StudentCourseMessage)
	//	if err != nil {
	//		return
	//	}
	//}

	// for order-base update of student status in usermgmt service
	err = s.publishOrderEventLog(ctx, message.Order, message.Student)
	if err != nil {
		grpclog.Warningf("Error when publishing order event log with error=%v and message=%v", err, map[string]interface{}{
			"order_id": message.Order.OrderID.String,
		})
	}
	// for discount automation in discount service
	err = s.publishOrderWithProductInfoLog(ctx, message.Order, message.StudentProducts)
	if err != nil {
		grpclog.Warningf("Error when publishing order with product info log with error=%v and message=%v", err, map[string]interface{}{
			"order_id": message.Order.OrderID.String,
		})
	}

	err = s.PublishNotification(ctx, message.SystemNotificationMessage)
	if err != nil {
		grpclog.Warningf("Error when publishing notification with error=%v and message=%v", err, map[string]interface{}{
			"order_id": message.Order.OrderID.String,
		})
	}

	if len(message.StudentPackages) != 0 {
		err = s.PublishStudentPackageForCreateOrder(ctx, message.StudentPackages)
		if err != nil {
			grpclog.Warningf("Error when publishing student package for create order with error=%v and message=%v", err, map[string]interface{}{
				"order_id": message.Order.OrderID.String,
			})
		}
	}
	return nil
}

func (s *SubscriptionService) PublishStudentPackageForCreateOrder(ctx context.Context, eventMessages []*npb.EventStudentPackage) (err error) {
	for _, message := range eventMessages {
		var (
			data       []byte
			dataV2     []byte
			msgID      string
			locationID string
		)
		data, err = proto.Marshal(message)
		if err != nil {
			return
		}
		msgID, err = s.JSM.PublishAsyncContext(ctx, constants.SubjectStudentPackageEventNats, data)
		if err != nil {
			return nats.HandlePushMsgFail(ctx, fmt.Errorf("SubjectStudentPackageEventNats JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, err))
		}

		if len(message.StudentPackage.Package.LocationIds) > 0 {
			locationID = message.StudentPackage.Package.LocationIds[0]
		}
		courseIds := message.StudentPackage.Package.CourseIds
		for i := 0; i < len(courseIds); i++ {
			var studentPackageClass entities.StudentPackageClass
			courseID := courseIds[i]
			studentPackageClass, err = s.StudentPackageClassRepo.GetByStudentPackageID(ctx, s.DB, message.StudentPackage.Package.StudentPackageId)
			if err != nil {
				return
			}
			eventV2 := &npb.EventStudentPackageV2{
				StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
					StudentId: message.StudentPackage.StudentId,
					Package: &npb.EventStudentPackageV2_PackageV2{
						CourseId:   courseID,
						ClassId:    studentPackageClass.ClassID.String,
						LocationId: locationID,
						StartDate:  message.StudentPackage.Package.StartDate,
						EndDate:    message.StudentPackage.Package.EndDate,
					},
					IsActive: true,
				},
			}
			dataV2, err = proto.Marshal(eventV2)
			if err != nil {
				return
			}
			msgID, err = s.JSM.PublishAsyncContext(ctx, constants.SubjectStudentPackageV2EventNats, dataV2)
			if err != nil {
				return nats.HandlePushMsgFail(ctx, fmt.Errorf("SubjectStudentPackageV2EventNats JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, err))
			}
		}
	}
	return
}

func (s *SubscriptionService) PublishStudentPackage(ctx context.Context, eventMessages []*npb.EventStudentPackage) (err error) {
	for _, message := range eventMessages {
		var (
			data                []byte
			dataV2              []byte
			msgID               string
			courseID            string
			locationID          string
			studentPackageClass entities.StudentPackageClass
		)
		data, err = proto.Marshal(message)
		if err != nil {
			return
		}
		msgID, err = s.JSM.PublishAsyncContext(ctx, constants.SubjectStudentPackageEventNats, data)
		if err != nil {
			return nats.HandlePushMsgFail(ctx, fmt.Errorf("SubjectStudentPackageEventNats JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, err))
		}

		if len(message.StudentPackage.Package.CourseIds) > 0 {
			courseID = message.StudentPackage.Package.CourseIds[0]
		}

		if len(message.StudentPackage.Package.LocationIds) > 0 {
			locationID = message.StudentPackage.Package.LocationIds[0]
		}
		studentPackageClass, err = s.StudentPackageClassRepo.GetByStudentPackageID(ctx, s.DB, message.StudentPackage.Package.StudentPackageId)
		if err != nil {
			return
		}

		eventV2 := &npb.EventStudentPackageV2{
			StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
				StudentId: message.StudentPackage.StudentId,
				Package: &npb.EventStudentPackageV2_PackageV2{
					CourseId:   courseID,
					ClassId:    studentPackageClass.ClassID.String,
					LocationId: locationID,
					StartDate:  message.StudentPackage.Package.StartDate,
					EndDate:    message.StudentPackage.Package.EndDate,
				},
				IsActive: true,
			},
		}
		dataV2, err = proto.Marshal(eventV2)
		if err != nil {
			return
		}
		msgID, err = s.JSM.PublishAsyncContext(ctx, constants.SubjectStudentPackageV2EventNats, dataV2)
		if err != nil {
			return nats.HandlePushMsgFail(ctx, fmt.Errorf("SubjectStudentPackageV2EventNats JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, err))
		}
	}
	return
}

func (s *SubscriptionService) PublishStudentClass(ctx context.Context, eventMessages []*npb.EventStudentPackageV2) (err error) {
	for _, message := range eventMessages {
		var (
			data  []byte
			msgID string
		)
		data, err = proto.Marshal(message)
		if err != nil {
			return
		}
		msgID, err = s.JSM.PublishAsyncContext(ctx, constants.SubjectStudentPackageV2EventNats, data)
		if err != nil {
			return nats.HandlePushMsgFail(ctx, fmt.Errorf("SubjectStudentPackageV2EventNats JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, err))
		}
	}
	return
}

func (s *SubscriptionService) publishStudentCourseSyncEvent(ctx context.Context, studentCourseSync []*pb.EventSyncStudentPackageCourse) (err error) {
	var (
		data  []byte
		msgID string
	)
	data, err = json.Marshal(studentCourseSync)
	if err != nil {
		return
	}
	msgID, err = s.JSM.PublishAsyncContext(ctx, constants.SubjectStudentCourseEventSync, data)
	if err != nil {
		return nats.HandlePushMsgFail(ctx, fmt.Errorf("SubjectStudentCourseEventSync JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, err))
	}
	return
}

func (s *SubscriptionService) publishOrderEventLog(
	ctx context.Context,
	order entities.Order,
	student entities.Student,
) (
	err error,
) {
	var (
		data  []byte
		msgID string
	)
	orderEventLog := entities.OrderEventLog{
		OrderStatus:         order.OrderStatus.String,
		OrderType:           order.OrderType.String,
		StudentID:           order.StudentID.String,
		LocationID:          order.LocationID.String,
		StartDate:           time.Now(),
		EndDate:             time.Time{},
		OrderID:             order.OrderID.String,
		OrderSequenceNumber: order.OrderSequenceNumber.Int,
	}

	if order.OrderType.String != pb.OrderType_ORDER_TYPE_CUSTOM_BILLING.String() {
		orderEventLog = entities.OrderEventLog{
			OrderStatus:         order.OrderStatus.String,
			OrderType:           order.OrderType.String,
			StudentID:           order.StudentID.String,
			LocationID:          order.LocationID.String,
			EnrollmentStatus:    student.EnrollmentStatus.String,
			StartDate:           time.Now(),
			EndDate:             time.Time{},
			OrderID:             order.OrderID.String,
			OrderSequenceNumber: order.OrderSequenceNumber.Int,
		}
	}

	if order.OrderType.String == pb.OrderType_ORDER_TYPE_WITHDRAWAL.String() ||
		order.OrderType.String == pb.OrderType_ORDER_TYPE_GRADUATE.String() {
		orderEventLog.StartDate = order.WithdrawalEffectiveDate.Time
	}
	if order.OrderType.String == pb.OrderType_ORDER_TYPE_LOA.String() ||
		order.OrderType.String == pb.OrderType_ORDER_TYPE_RESUME.String() {
		orderEventLog.StartDate = order.LOAStartDate.Time
	}

	data, err = json.Marshal(orderEventLog)
	if err != nil {
		return
	}

	msgID, err = s.JSM.PublishAsyncContext(ctx, constants.SubjectOrderEventLogCreated, data)
	if err != nil {
		return nats.HandlePushMsgFail(ctx, fmt.Errorf("PublishOrderEventLog JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, err))
	}
	return
}

func (s *SubscriptionService) publishOrderWithProductInfoLog(
	ctx context.Context,
	order entities.Order,
	studentProducts []entities.StudentProduct,
) (
	err error,
) {
	var (
		data  []byte
		msgID string
	)

	studentProductIds := []string{}
	for _, product := range studentProducts {
		studentProductIds = append(studentProductIds, product.StudentProductID.String)
	}

	orderWithProductInfoLog := entities.OrderWithProductInfoLog{
		OrderID:           order.OrderID.String,
		StudentID:         order.StudentID.String,
		LocationID:        order.LocationID.String,
		OrderStatus:       order.OrderStatus.String,
		OrderType:         order.OrderType.String,
		StudentProductIDs: studentProductIds,
	}

	data, err = json.Marshal(orderWithProductInfoLog)
	if err != nil {
		return
	}

	msgID, err = s.JSM.PublishAsyncContext(ctx, constants.SubjectOrderWithProductInfoLogCreated, data)
	if err != nil {
		return nats.HandlePushMsgFail(ctx, fmt.Errorf("PublishOrderWithProductInfoLog JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, err))
	}
	return
}

func (s *SubscriptionService) PublishNotification(ctx context.Context, publishNotificationMessage *payload.UpsertSystemNotification) (err error) {
	if publishNotificationMessage == nil || publishNotificationMessage.ReferenceID == "" {
		return
	}
	var (
		data []byte
	)
	data, err = json.Marshal(publishNotificationMessage)
	if err != nil {
		return fmt.Errorf("error when marshal message for PublishNotification: %v", err)
	}

	msgKey, err := json.Marshal(publishNotificationMessage.ReferenceID)
	if err != nil {
		return fmt.Errorf("error when marshal mgsKey for PublishNotification: %v", err)
	}
	spanName := "PUBLISHER.PAYMENT" + constants.SystemNotificationUpsertingTopic
	err = s.Kafka.TracedPublishContext(ctx, spanName, constants.SystemNotificationUpsertingTopic, msgKey, data)
	if err != nil {
		return fmt.Errorf("error when publish failed kafka PublishNotification: %v", err)
	}
	return
}

func (s *SubscriptionService) ToNotificationMessage(ctx context.Context, tx database.QueryExecer, order entities.Order, student entities.Student, upsertNotificationData utils.UpsertSystemNotificationData) (notificationMessage *payload.UpsertSystemNotification, err error) {
	switch order.OrderStatus.String {
	// In case: Void an order
	case pb.OrderStatus_ORDER_STATUS_VOIDED.String():
		notificationMessage = s.toNotificationMessageForVoid(ctx, order)
		return
		// In case: Create an order
	case pb.OrderStatus_ORDER_STATUS_SUBMITTED.String():
		notificationMessage, err = s.toNotificationMessageForCreate(ctx, tx, order, student, upsertNotificationData)
		return
	}
	return
}

func (s *SubscriptionService) toNotificationMessageForCreate(ctx context.Context, tx database.QueryExecer, order entities.Order, student entities.Student, upsertNotificationData utils.UpsertSystemNotificationData) (notificationMessage *payload.UpsertSystemNotification, err error) {
	switch order.OrderType.String {
	case pb.OrderType_ORDER_TYPE_LOA.String():
		var (
			userIDs          []string
			notificationDate entities.NotificationDate
			grade            entities.Grade
		)

		if upsertNotificationData.StudentDetailPath == "" {
			err = status.Errorf(codes.Internal, "Error when missing student detail path")
			return
		}

		notificationDate, err = s.NotificationDateRepo.GetByOrderType(ctx, tx, order.OrderType.String)
		if err != nil {
			err = status.Errorf(codes.Internal, "Error when get notification date by order type: %v", err.Error())
			return
		}
		endDate := utils.ConvertToLocalTime(upsertNotificationData.EndDate, upsertNotificationData.Timezone)
		daysInPrevMonthOfEndDate := utils.DaysIn(endDate.Month()-1, endDate.Year())
		if int(notificationDate.NotificationDate.Int) > daysInPrevMonthOfEndDate {
			return
		}

		grade, err = s.gradeRepo.GetByID(ctx, tx, student.GradeID.String)
		if err != nil {
			err = status.Errorf(codes.Internal, "Error when getting grade with grade_id=%s: %v", student.GradeID.String, err.Error())
			return
		}

		userIDs, err = s.userRepo.GetUserIDsByRoleNamesAndLocationID(ctx, tx,
			[]string{
				constant.RoleCentreManager,
				constant.RoleCentreStaff,
			}, order.LocationID.String)
		if err != nil {
			err = status.Errorf(codes.Internal, "Error when get user_ids by role names and location id: %v", err.Error())
			return
		}
		recipients := make([]payload.SystemNotificationRecipient, 0, len(userIDs))
		for _, userID := range userIDs {
			recipients = append(recipients, payload.SystemNotificationRecipient{
				UserID: userID,
			})
		}

		notificationMessage = &payload.UpsertSystemNotification{
			ReferenceID: order.OrderID.String,
			Content: []payload.SystemNotificationContent{
				{
					Language: string(utils.EnCode),
					Text:     fmt.Sprintf(utils.EngLOANotificationContentTemp, grade.Name.String, order.StudentFullName.String, upsertNotificationData.LocationName),
				},
				{
					Language: string(utils.JpCode),
					Text:     fmt.Sprintf(utils.JpLOANotificationContentTemp, upsertNotificationData.LocationName, grade.Name.String, order.StudentFullName.String),
				},
			},
			URL:        upsertNotificationData.StudentDetailPath,
			ValidFrom:  time.Date(endDate.Year(), endDate.Month()-1, int(notificationDate.NotificationDate.Int), 0, 0, 0, 0, endDate.Location()),
			Recipients: recipients,
			IsDeleted:  false,
		}
		return
	case pb.OrderType_ORDER_TYPE_RESUME.String():
		var (
			latestLOAOrder entities.Order
		)
		latestLOAOrder, err = s.orderRepo.GetLatestOrderByStudentIDAndLocationIDAndOrderType(ctx, tx, student.StudentID.String, order.LocationID.String, pb.OrderType_ORDER_TYPE_LOA.String())
		if err != nil {
			err = status.Errorf(codes.Internal, "error when get latest LOA order of student with student_id=%s and location_id=%s: %v", student.StudentID.String, order.LocationID.String, err)
			return
		}
		notificationMessage = &payload.UpsertSystemNotification{
			ReferenceID: latestLOAOrder.OrderID.String,
			IsDeleted:   true,
		}
		return
	case pb.OrderType_ORDER_TYPE_GRADUATE.String(),
		pb.OrderType_ORDER_TYPE_WITHDRAWAL.String():
		var (
			latestOrder entities.Order
		)
		latestOrder, err = s.orderRepo.GetLatestOrderByStudentIDAndLocationID(ctx, tx, student.StudentID.String, order.LocationID.String)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				err = nil
			} else {
				err = status.Errorf(codes.Internal, "error when get latest order of student with student_id=%s and location_id=%s: %v", student.StudentID.String, order.LocationID.String, err)
			}
			return
		}
		// In case: A student create a withdrawal/graduate order when is on LOA status
		if latestOrder.OrderType.String == pb.OrderType_ORDER_TYPE_LOA.String() &&
			latestOrder.OrderStatus.String == pb.OrderStatus_ORDER_STATUS_SUBMITTED.String() {
			notificationMessage = &payload.UpsertSystemNotification{
				ReferenceID: latestOrder.OrderID.String,
				IsDeleted:   true,
			}
		}
	default:
	}

	return
}

func (s *SubscriptionService) toNotificationMessageForVoid(_ context.Context, order entities.Order) (notificationMessage *payload.UpsertSystemNotification) {
	switch order.OrderType.String {
	case pb.OrderType_ORDER_TYPE_LOA.String():
		notificationMessage = &payload.UpsertSystemNotification{
			ReferenceID: order.OrderID.String,
			IsDeleted:   true,
		}
		return
	default:
	}
	return
}

func NewSubscriptionService(jsm nats.JetStreamManagement, db database.Ext, kafka kafka.KafkaManagement, config configs.CommonConfig) *SubscriptionService {
	return &SubscriptionService{
		JSM:                     jsm,
		DB:                      db,
		Kafka:                   kafka,
		Config:                  config,
		StudentPackageClassRepo: &repositories.StudentPackageClassRepo{},
		NotificationDateRepo:    &repositories.NotificationDateRepo{},
		userRepo:                &repositories.UserRepo{},
		orderRepo:               &repositories.OrderRepo{},
		gradeRepo:               &repositories.GradeRepo{},
	}
}
