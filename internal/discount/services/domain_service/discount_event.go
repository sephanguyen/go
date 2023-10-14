package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	discountconstant "github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/discount/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/kafka"
	"github.com/manabie-com/backend/internal/golibs/kafka/payload"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/payment/utils"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DiscountEventService struct {
	JSM             nats.JetStreamManagement
	Kafka           kafka.KafkaManagement
	DiscountService interface {
		OrderValidationForDiscountEligibility(ctx context.Context, data []byte) (bool, error)
	}
	UserRepo interface {
		GetUserIDsByRoleNamesAndLocationID(ctx context.Context, db database.QueryExecer,
			roleNames []string, locationID string) (userIDs []string, err error)
		GetByID(ctx context.Context,
			db database.QueryExecer, studentID string) (entities.User, error)
	}
	OrderItemRepo interface {
		GetLatestByStudentProductID(ctx context.Context,
			db database.QueryExecer, studentProductID string) (entities.OrderItem, error)
	}
	ProductRepo interface {
		GetByID(ctx context.Context,
			db database.QueryExecer, productID string) (entities.Product, error)
	}
}

func (s *DiscountEventService) PublishEventForUpdateStudentProduct(ctx context.Context, updateOrderInfo entities.UpdateProductDiscount) (err error) {
	var (
		data []byte
	)

	data, err = json.Marshal(updateOrderInfo)
	if err != nil {
		return
	}

	_, err = s.JSM.PublishContext(ctx, constants.SubjectUpdateStudentProductCreated, data)
	if err != nil {
		return nats.HandlePushMsgFail(ctx, fmt.Errorf("PublishEventForUpdateStudentProduct JSM.PublishContext failed: %s", err))
	}
	return
}

func (s *DiscountEventService) PublishNotificationForStudentProductWithScheduleTag(ctx context.Context, tx database.QueryExecer,
	studentProduct entities.StudentProduct) (err error) {
	var (
		userIDs                 []string
		notificationMessage     payload.UpsertSystemNotification
		product                 entities.Product
		student                 entities.User
		orderItemForUpdateOrder entities.OrderItem
	)

	product, err = s.ProductRepo.GetByID(ctx, tx, studentProduct.ProductID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when get product by product_id with product_id = %s: %v", studentProduct.ProductID.String, err.Error())
		return
	}
	student, err = s.UserRepo.GetByID(ctx, tx, studentProduct.StudentID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when get student by student_id with student_id = %s: %v", studentProduct.StudentID.String, err.Error())
		return
	}
	// Latest order item by student_product_id is always for Update Order
	orderItemForUpdateOrder, err = s.OrderItemRepo.GetLatestByStudentProductID(ctx, tx, studentProduct.StudentProductID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when get order item by student product id with student_product_id = %s: %v", studentProduct.StudentProductID.String, err.Error())
		return
	}

	userIDs, err = s.UserRepo.GetUserIDsByRoleNamesAndLocationID(ctx, tx,
		[]string{
			constant.RoleCentreManager,
		}, studentProduct.LocationID.String)
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

	notificationMessage = payload.UpsertSystemNotification{
		ReferenceID: orderItemForUpdateOrder.OrderID.String,
		Content: []payload.SystemNotificationContent{
			{
				Language: string(utils.EnCode),
				Text:     fmt.Sprintf(discountconstant.EngNotificationContentTempForStudentProductWithScheduleTag, product.Name.String, student.Name.String),
			},
			{
				Language: string(utils.JpCode),
				Text:     fmt.Sprintf(discountconstant.JpNotificationContentTempForStudentProductWithScheduleTag, student.Name.String, product.Name.String),
			},
		},
		URL:        fmt.Sprintf(discountconstant.StudentBillingTabPathTemplate, studentProduct.StudentID.String),
		ValidFrom:  time.Now(),
		Recipients: recipients,
		IsDeleted:  false,
		Status:     payload.SystemNotificationStatusNew,
	}
	err = s.PublishNotification(ctx, notificationMessage)
	return
}

func (s *DiscountEventService) PublishNotification(ctx context.Context, publishNotificationMessage payload.UpsertSystemNotification) (err error) {
	if publishNotificationMessage.ReferenceID == "" {
		err = status.Errorf(codes.Internal, "error when missing reference_id when publish notification")
		return
	}
	var (
		data []byte
	)
	data, err = json.Marshal(&publishNotificationMessage)
	if err != nil {
		return fmt.Errorf("error when marshal message for PublishNotification: %v", err)
	}

	msgKey, err := json.Marshal(publishNotificationMessage.ReferenceID)
	if err != nil {
		return fmt.Errorf("error when marshal mgsKey for PublishNotification: %v", err)
	}
	spanName := "PUBLISHER.DISCOUNT" + constants.SystemNotificationUpsertingTopic
	err = s.Kafka.TracedPublishContext(ctx, spanName, constants.SystemNotificationUpsertingTopic, msgKey, data)
	if err != nil {
		return fmt.Errorf("error when publish failed kafka PublishNotification: %v", err)
	}
	return
}

func NewDiscountEventService(jsm nats.JetStreamManagement, kafka kafka.KafkaManagement) *DiscountEventService {
	return &DiscountEventService{
		JSM:           jsm,
		Kafka:         kafka,
		UserRepo:      &repositories.UserRepo{},
		OrderItemRepo: &repositories.OrderItemRepo{},
		ProductRepo:   &repositories.ProductRepo{},
	}
}
