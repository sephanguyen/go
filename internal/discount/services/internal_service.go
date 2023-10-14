package services

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/discount/repositories"
	domainService "github.com/manabie-com/backend/internal/discount/services/domain_service"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/kafka"
	"github.com/manabie-com/backend/internal/golibs/nats"
	paymentPb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/zap"
)

type InternalService struct {
	DB                    database.Ext
	Logger                *zap.Logger
	DiscountEventService  IDiscountEventServiceForInternalService
	DiscountTagService    IDiscountTagServiceForInternalService
	StudentProductService IStudentProductServiceForInternalService
	DiscountRepo          IDiscountRepoForInternalService
	UserService           IUserServiceForInternalService
}

type IDiscountEventServiceForInternalService interface {
	PublishEventForUpdateStudentProduct(ctx context.Context, updateOrderInfo entities.UpdateProductDiscount) error
	PublishNotificationForStudentProductWithScheduleTag(ctx context.Context, tx database.QueryExecer,
		studentProduct entities.StudentProduct) (err error)
}

type IDiscountTagServiceForInternalService interface {
	RetrieveDiscountEligibilityOfStudentProduct(ctx context.Context, db database.QueryExecer, userID string, locationID string, productID string) ([]*entities.UserDiscountTag, error)
	RetrieveDiscountTagsWithActivityOnDate(ctx context.Context, db database.QueryExecer, timestamp time.Time) ([]*entities.UserDiscountTag, error)
	RetrieveUserIDsWithActivityOnDate(ctx context.Context, db database.QueryExecer, timestamp time.Time) ([]string, error)
}

type IStudentProductServiceForInternalService interface {
	RetrieveStudentProductByID(ctx context.Context, db database.QueryExecer, studentProductID string) (entities.StudentProduct, error)
	RetrieveActiveStudentProductsOfStudentInLocation(ctx context.Context, db database.QueryExecer, studentID string, locationID string) ([]*entities.StudentProduct, error)
	RetrieveDiscountOfStudentProduct(ctx context.Context, studentProductID string) (entities.Discount, error)
}

type IDiscountRepoForInternalService interface {
	GetMaxDiscountByTypeAndDiscountTagIDs(ctx context.Context, db database.QueryExecer, discountType string, discountTagIDs []string) (entities.Discount, error)
	GetMaxProductDiscountByProductID(ctx context.Context, db database.QueryExecer, productID string) (entities.Discount, error)
}

type IUserServiceForInternalService interface {
	GetUserIDsByRoleNamesAndLocationID(ctx context.Context, db database.QueryExecer, roleNames []string, locationID string) (userIDs []string, err error)
}

type StudentWithLocation struct {
	StudentID  string
	LocationID string
}

func (s *InternalService) RetrieveStudentsCandidateForDiscountUpdateOnDate(
	ctx context.Context,
	timestamp time.Time,
) (
	studentWithLocationList []StudentWithLocation,
	err error,
) {
	studentIDs, err := s.DiscountTagService.RetrieveUserIDsWithActivityOnDate(ctx, s.DB, timestamp)
	if err != nil {
		return
	}

	studentWithLocationList = []StudentWithLocation{}
	for _, studentID := range studentIDs {
		studentDiscountEligibleLocation := StudentWithLocation{
			StudentID:  studentID,
			LocationID: "",
		}
		studentWithLocationList = append(studentWithLocationList, studentDiscountEligibleLocation)
	}

	return
}

func (s *InternalService) RetrieveActiveStudentProductsOfStudentInLocation(
	ctx context.Context,
	studentID string,
	locationID string,
) (
	studentProducts []*entities.StudentProduct,
	err error,
) {
	return s.StudentProductService.RetrieveActiveStudentProductsOfStudentInLocation(ctx, s.DB, studentID, locationID)
}

func (s *InternalService) RetrieveHighestDiscountOfStudentProduct(
	ctx context.Context,
	studentID string,
	locationID string,
	productID string,
) (
	discount entities.Discount,
	err error,
) {
	availableDiscountTags, err := s.DiscountTagService.RetrieveDiscountEligibilityOfStudentProduct(ctx, s.DB, studentID, locationID, productID)
	if err != nil {
		return entities.Discount{}, err
	}

	discountTagIDs := []string{}
	for _, discountTag := range availableDiscountTags {
		discountTagIDs = append(discountTagIDs, discountTag.DiscountTagID.String)
	}

	discount, err = s.DiscountRepo.GetMaxDiscountByTypeAndDiscountTagIDs(ctx, s.DB, paymentPb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE.String(), discountTagIDs)
	if err != nil || discount.DiscountID.Status != pgtype.Present {
		return s.DiscountRepo.GetMaxProductDiscountByProductID(ctx, s.DB, productID)
	}

	return
}

func (s *InternalService) RetrieveCurrentDiscountOfStudentProduct(
	ctx context.Context,
	studentProductID string,
) (
	discount entities.Discount,
	err error,
) {
	return s.StudentProductService.RetrieveDiscountOfStudentProduct(ctx, studentProductID)
}

func (s *InternalService) ValidateProductAndPublishUpdateOrderEvent(
	ctx context.Context,
	studentProductID string,
	discount entities.Discount,
) (
	err error,
) {
	var (
		discountAmountValue   float32
		effectiveDate         time.Time
		studentProductEndDate time.Time
		updateOrderInfo       entities.UpdateProductDiscount
	)

	studentProduct, err := s.StudentProductService.RetrieveStudentProductByID(ctx, s.DB, studentProductID)
	if err != nil {
		return
	}
	if studentProduct.StudentProductLabel.String == paymentPb.StudentProductLabel_UPDATE_SCHEDULED.String() ||
		studentProduct.StudentProductLabel.String == paymentPb.StudentProductLabel_WITHDRAWAL_SCHEDULED.String() ||
		studentProduct.StudentProductLabel.String == paymentPb.StudentProductLabel_GRADUATION_SCHEDULED.String() ||
		studentProduct.StudentProductLabel.String == paymentPb.StudentProductLabel_PAUSE_SCHEDULED.String() {
		s.Logger.Debug(fmt.Sprintf("failed to publish update order event for student product %s product has update scheduled tag", studentProductID))
		err = s.DiscountEventService.PublishNotificationForStudentProductWithScheduleTag(ctx, s.DB, studentProduct)
		if err != nil {
			s.Logger.Debug(fmt.Sprintf("failed to publish notification for student product with schedule tag with student_product_id = %s: %v", studentProductID, err))
			return
		}
		return nil
	}

	effectiveDate = time.Now()
	if studentProduct.StartDate.Time.After(effectiveDate) {
		err = studentProduct.StartDate.AssignTo(&effectiveDate)
		if err != nil {
			return
		}
	}

	err = studentProduct.EndDate.AssignTo(&studentProductEndDate)
	if err != nil {
		return
	}

	if discount.DiscountID.String != "" {
		err = discount.DiscountAmountValue.AssignTo(&discountAmountValue)
		if err != nil {
			return
		}

		updateOrderInfo = entities.UpdateProductDiscount{
			StudentID:             studentProduct.StudentID.String,
			LocationID:            studentProduct.LocationID.String,
			ProductID:             studentProduct.ProductID.String,
			StudentProductID:      studentProductID,
			EffectiveDate:         effectiveDate,
			StudentProductEndDate: studentProductEndDate,
			DiscountID:            discount.DiscountID.String,
			DiscountType:          paymentPb.DiscountType(paymentPb.DiscountType_value[discount.DiscountType.String]),
			DiscountAmountType:    paymentPb.DiscountAmountType(paymentPb.DiscountAmountType_value[discount.DiscountAmountType.String]),
			DiscountAmountValue:   discountAmountValue,
		}
	} else {
		// handle removal of discount
		updateOrderInfo = entities.UpdateProductDiscount{
			StudentID:             studentProduct.StudentID.String,
			LocationID:            studentProduct.LocationID.String,
			ProductID:             studentProduct.ProductID.String,
			StudentProductID:      studentProductID,
			EffectiveDate:         effectiveDate,
			StudentProductEndDate: studentProductEndDate,
			DiscountID:            discount.DiscountID.String,
			DiscountType:          paymentPb.DiscountType_DISCOUNT_TYPE_NONE,
			DiscountAmountType:    paymentPb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_NONE,
			DiscountAmountValue:   0,
		}
	}

	return s.DiscountEventService.PublishEventForUpdateStudentProduct(ctx, updateOrderInfo)
}

func NewInternalService(db database.Ext, jsm nats.JetStreamManagement, logger *zap.Logger, kafka kafka.KafkaManagement) *InternalService {
	return &InternalService{
		DB:                    db,
		Logger:                logger,
		DiscountEventService:  domainService.NewDiscountEventService(jsm, kafka),
		DiscountTagService:    domainService.NewDiscountTagService(db),
		StudentProductService: domainService.NewStudentProductService(db),
		DiscountRepo:          &repositories.DiscountRepo{},
		UserService:           domainService.NewUserService(),
	}
}
