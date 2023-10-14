package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/discount/repositories"
	domainService "github.com/manabie-com/backend/internal/discount/services/domain_service"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/payment/utils"
	paymentPb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/zap"
)

type DiscountService struct {
	DB                     database.Ext
	JSM                    nats.JetStreamManagement
	Logger                 *zap.Logger
	DiscountTagService     IDiscountTagServiceForDiscountService
	DiscountTrackerService IDiscountTrackerServiceForDiscountService
	ProductGroupService    IProductGroupServiceForDiscountService
	StudentProductService  IStudentProductServiceForDiscountService
	StudentSiblingService  IStudentSiblingServiceForDiscountService
	DiscountRepo           IDiscountRepoForDiscountService
}

type IDiscountTagServiceForDiscountService interface {
	RetrieveActiveDiscountTagIDsByDateAndUserID(ctx context.Context, db database.QueryExecer, timestamp time.Time, userID string) (userDiscountTagIDs []string, err error)
	RetrieveDiscountTagByDiscountTagID(ctx context.Context, db database.QueryExecer, discountTagID string) (discountTag *entities.DiscountTag, err error)
	SoftDeleteUserDiscountTagsByTypesAndUserID(ctx context.Context, db database.QueryExecer, userID string, discountTypes pgtype.TextArray) (err error)
	RetrieveEligibleDiscountTagsOfStudentInLocation(ctx context.Context, db database.QueryExecer, userID string, locationID string) (userDiscountTags []*entities.UserDiscountTag, err error)
	CreateUserDiscountTag(ctx context.Context, db database.QueryExecer, userDiscountTag *entities.UserDiscountTag) (err error)
	RetrieveEligibleDiscountTagsOfStudent(ctx context.Context, db database.QueryExecer, userID string) (userDiscountTags []*entities.UserDiscountTag, err error)
	UpdateDiscountTagOfStudentIDWithTimeSegment(ctx context.Context, db database.QueryExecer, studentID string, discountType string, discountTagID []string, timestampSegments []entities.TimestampSegment) (err error)
	RetrieveDiscountTagIDsByDiscountType(ctx context.Context, db database.QueryExecer, discountType string) (discountTagIDs []string, err error)
}

type IDiscountRepoForDiscountService interface {
	GetByDiscountTagIDs(ctx context.Context, db database.QueryExecer, discountTagIDs []string) ([]*entities.Discount, error)
}

type IStudentSiblingServiceForDiscountService interface {
	RetrieveStudentSiblingIDs(ctx context.Context, db database.QueryExecer, studentID string) ([]string, error)
}

type IProductGroupServiceForDiscountService interface {
	RetrieveEligibleProductGroupsOfStudentProductsByDiscountType(ctx context.Context, studentProducts []entities.StudentProduct, discountType string) (productDiscountGroups []entities.ProductDiscountGroup)
}

type IDiscountTrackerServiceForDiscountService interface {
	TrackDiscount(ctx context.Context, db database.QueryExecer, studentDiscontTracker *entities.StudentDiscountTracker) error
	UpdateTrackingDurationByStudentProduct(ctx context.Context, db database.QueryExecer, studentProduct entities.StudentProduct) error
	RetrieveSiblingDiscountTrackingHistoriesByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) (studentTrackingHistories map[string][]entities.StudentDiscountTracker, siblingTrackingHistories map[string][]entities.StudentDiscountTracker, err error)
}

type IStudentProductServiceForDiscountService interface {
	RetrieveStudentProductsByIDs(ctx context.Context, db database.QueryExecer, studentProductIDs []string) ([]entities.StudentProduct, error)
	RetrieveStudentProductByID(ctx context.Context, db database.QueryExecer, studentProductID string) (entities.StudentProduct, error)
	RetrieveStudentProductsByOrderID(ctx context.Context, db database.QueryExecer, orderID string) ([]entities.StudentProduct, error)
}

func (s *DiscountService) SubscribeToOrderWithProductInfoLog() error {
	option := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
			nats.DeliverNew(),
			nats.Bind(constants.StreamOrderWithProductInfoLog, constants.DurableOrderWithProductInfoLogCreated),
			nats.DeliverSubject(constants.DeliverOrderWithProductInfoLogCreated),
		},
	}

	_, err := s.JSM.QueueSubscribe(constants.SubjectOrderWithProductInfoLogCreated,
		constants.QueueOrderWithProductInfoLogCreated, option, s.ValidateOrderedProductsForDiscountEligibility)
	if err != nil {
		return fmt.Errorf("orderWithProductInfoLog.Subscribe: %w", err)
	}

	return nil
}

func (s *DiscountService) ValidateOrderedProductsForDiscountEligibility(ctx context.Context, data []byte) (res bool, err error) {
	_, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var orderEventLog *entities.OrderWithProductInfoLog
	if err := json.Unmarshal(data, &orderEventLog); err != nil {
		return false, err
	}

	// skip custom billing type
	if orderEventLog.OrderType == paymentPb.OrderType_ORDER_TYPE_CUSTOM_BILLING.Enum().String() {
		return true, nil
	}

	studentProducts, err := s.StudentProductService.RetrieveStudentProductsByIDs(ctx, s.DB, orderEventLog.StudentProductIDs)

	// check eligible sibling product groups of student products and eliminate products without group data
	productInSiblingDiscountGroups := s.ProductGroupService.RetrieveEligibleProductGroupsOfStudentProductsByDiscountType(ctx, studentProducts, paymentPb.DiscountType_DISCOUNT_TYPE_SIBLING.String())

	// validation for sibling discount
	siblingIDs, err := s.StudentSiblingService.RetrieveStudentSiblingIDs(ctx, s.DB, orderEventLog.StudentID)
	if err == nil && len(siblingIDs) > 0 {
		newTrackingFlag := false
		trackingUpdateFlag := false

		// void orders
		if orderEventLog.OrderStatus == paymentPb.OrderStatus_ORDER_STATUS_VOIDED.String() {
			trackingUpdateFlag, err = s.UpdateTrackingDataForVoidOrders(ctx, orderEventLog.OrderID)
		}

		// cases with updating of existing tracking records
		if orderEventLog.OrderStatus == paymentPb.OrderStatus_ORDER_STATUS_SUBMITTED.String() && (orderEventLog.OrderType == paymentPb.OrderType_ORDER_TYPE_WITHDRAWAL.String() ||
			orderEventLog.OrderType == paymentPb.OrderType_ORDER_TYPE_GRADUATE.String() ||
			orderEventLog.OrderType == paymentPb.OrderType_ORDER_TYPE_LOA.String() ||
			orderEventLog.OrderType == paymentPb.OrderType_ORDER_TYPE_UPDATE.String()) {
			trackingUpdateFlag, err = s.UpdateProductDiscountTrackingData(ctx, productInSiblingDiscountGroups)
			if err != nil {
				s.Logger.Error(fmt.Sprintf("updating tracker data failed for student %v with err: %v", orderEventLog.StudentID, err))
			}
		}

		// cases with creation of new student products
		if orderEventLog.OrderStatus == paymentPb.OrderStatus_ORDER_STATUS_SUBMITTED.String() && (orderEventLog.OrderType == paymentPb.OrderType_ORDER_TYPE_NEW.String() ||
			orderEventLog.OrderType == paymentPb.OrderType_ORDER_TYPE_ENROLLMENT.String() ||
			orderEventLog.OrderType == paymentPb.OrderType_ORDER_TYPE_RESUME.String() ||
			orderEventLog.OrderType == paymentPb.OrderType_ORDER_TYPE_UPDATE.String()) {
			newTrackingFlag, err = s.TrackValidSiblingDiscount(ctx, orderEventLog.StudentID, productInSiblingDiscountGroups)
			if err != nil {
				s.Logger.Error(fmt.Sprintf("sibling discount validation failed for student %v with err: %v", orderEventLog.StudentID, err))
			}
		}

		// validate tagging of students if there are updates in tracking data
		if newTrackingFlag || trackingUpdateFlag {
			err = s.TagValidSiblingDiscount(ctx, append(siblingIDs, orderEventLog.StudentID))
			if err != nil {
				err = fmt.Errorf("fail to tag student %v for sibling discount with err: %v", orderEventLog.StudentID, err)
				return
			}
		}
	}

	// validation for product combo discount
	err = s.TrackValidComboDiscount(ctx, orderEventLog.StudentID, studentProducts)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("combo discount validation failed for student %v with err: %v", orderEventLog.StudentID, err))
	}

	return true, nil
}

func (s *DiscountService) UpdateTrackingDataForVoidOrders(ctx context.Context, orderID string) (tracked bool, err error) {
	studentProducts, err := s.StudentProductService.RetrieveStudentProductsByOrderID(ctx, s.DB, orderID)
	if err != nil {
		return false, nil
	}

	for _, studentProduct := range studentProducts {
		err = s.DiscountTrackerService.UpdateTrackingDurationByStudentProduct(ctx, s.DB, studentProduct)
		if err != nil && !strings.Contains(err.Error(), "0 RowsAffected") {
			return false, nil
		}
	}

	return true, nil
}

// validate and track student product if it prequalifies or qualifies for sibling discount, does not guarantee tagging for sibling discount
func (s *DiscountService) TrackValidSiblingDiscount(ctx context.Context, studentID string, studentProductDiscountGroups []entities.ProductDiscountGroup) (tracked bool, err error) {
	if len(studentProductDiscountGroups) > 0 {
		tracked, err = s.TrackStudentDiscount(ctx, studentProductDiscountGroups)
		if err != nil {
			err = fmt.Errorf("fail to track student %v for sibling discount with err: %v", studentID, err)
			return
		}
	}

	return
}

// validate and track student product if it prequalifies or qualifies for combo discount, does not guarantee tagging for combo discount
func (s *DiscountService) TrackValidComboDiscount(ctx context.Context, studentID string, studentProducts []entities.StudentProduct) (err error) {
	comboProductDiscountGroup := s.ProductGroupService.RetrieveEligibleProductGroupsOfStudentProductsByDiscountType(ctx, studentProducts, paymentPb.DiscountType_DISCOUNT_TYPE_COMBO.String())
	if len(comboProductDiscountGroup) > 0 {
		_, err = s.TrackStudentDiscount(ctx, comboProductDiscountGroup)
		if err != nil {
			err = fmt.Errorf("fail to track student %v for combo discount with err: %v", studentID, err)
			return
		}
	}

	return
}

// if student is tracked for a new product that qualifies for sibling discount, re-tag all siblings for new duration of valid sibling discount
func (s *DiscountService) TagValidSiblingDiscount(ctx context.Context, studentIDs []string) (err error) {
	var (
		discountTagIDs           []string
		studentTrackingHistories map[string][]entities.StudentDiscountTracker
		siblingTrackingHistories map[string][]entities.StudentDiscountTracker
	)

	s.Logger.Info(fmt.Sprintf("TagValidSiblingDiscount: re-tagging student list with length %v, %v", len(studentIDs), studentIDs))
	studentTrackingHistories, siblingTrackingHistories, err = s.DiscountTrackerService.RetrieveSiblingDiscountTrackingHistoriesByStudentIDs(ctx, s.DB, studentIDs)
	if err != nil {
		return
	}

	discountTagIDs, err = s.DiscountTagService.RetrieveDiscountTagIDsByDiscountType(ctx, s.DB, paymentPb.DiscountType_DISCOUNT_TYPE_SIBLING.String())
	if err != nil {
		err = fmt.Errorf("fail to retrieve discount tag ids for sibling discount auto tagging with err: %v", err)
		return
	}

	for _, studentID := range studentIDs {
		var (
			studentDiscountTimeSegments []entities.TimestampSegment
			siblingDiscountTimeSegments []entities.TimestampSegment
			studentOverlappedDurations  []entities.TimestampSegment
		)

		if len(studentTrackingHistories[studentID]) == 0 || len(siblingTrackingHistories[studentID]) == 0 {
			err = s.DiscountTagService.SoftDeleteUserDiscountTagsByTypesAndUserID(ctx, s.DB, studentID, database.TextArray([]string{
				paymentPb.DiscountType_DISCOUNT_TYPE_SIBLING.String(),
			}))
			if err != nil && !strings.Contains(err.Error(), "0 RowsAffected") {
				return
			}

			continue
		}

		studentDiscountTimeSegments = s.GetExtendedDurationsOfTrackingData(studentTrackingHistories[studentID])
		siblingDiscountTimeSegments = s.GetExtendedDurationsOfTrackingData(siblingTrackingHistories[studentID])
		studentOverlappedDurations = s.GetOverlappedDurationsOfTrackingData(studentDiscountTimeSegments, siblingDiscountTimeSegments)

		for idx, overlappedDuration := range studentOverlappedDurations {
			s.Logger.Info(fmt.Sprintf("TagValidSiblingDiscount: overlapped duration for student %v item %v, start date: %v, end date: %v", studentID, (idx + 1), overlappedDuration.StartDate, overlappedDuration.EndDate))
		}

		err = s.DiscountTagService.UpdateDiscountTagOfStudentIDWithTimeSegment(ctx, s.DB, studentID, paymentPb.DiscountType_DISCOUNT_TYPE_SIBLING.String(), discountTagIDs, studentOverlappedDurations)
		if err != nil {
			err = fmt.Errorf("TagValidSiblingDiscount: fail to upsert user discount tag for sibling discount tagging for student %v with err: %v", studentID, err)
			s.Logger.Debug(fmt.Sprint(err))
			continue
		}
	}

	return
}

func (s *DiscountService) GetExtendedDurationsOfTrackingData(studentDiscountTrackingRecords []entities.StudentDiscountTracker) (timeSegments []entities.TimestampSegment) {
	timeSegments = []entities.TimestampSegment{}
	var (
		tmpTimeSegment entities.TimestampSegment
	)

	if len(studentDiscountTrackingRecords) == 0 {
		return
	}

	for idx, tracker := range studentDiscountTrackingRecords {
		if idx == 0 {
			tmpTimeSegment = entities.TimestampSegment{
				StartDate: tracker.StudentProductStartDate.Time,
				EndDate:   tracker.StudentProductEndDate.Time,
			}
			continue
		}
		if !tracker.StudentProductStartDate.Time.After(tmpTimeSegment.EndDate) && !tracker.StudentProductEndDate.Time.Before(tmpTimeSegment.EndDate) {
			tmpTimeSegment.EndDate = tracker.StudentProductEndDate.Time
			continue
		}

		timeSegments = append(timeSegments, tmpTimeSegment)
		tmpTimeSegment = entities.TimestampSegment{
			StartDate: tracker.StudentProductStartDate.Time,
			EndDate:   tracker.StudentProductEndDate.Time,
		}
	}

	timeSegments = append(timeSegments, tmpTimeSegment)
	return
}

func (s *DiscountService) GetOverlappedDurationsOfTrackingData(studentTimeSegments []entities.TimestampSegment, siblingTimeSegments []entities.TimestampSegment) (overlappedTimeSegments []entities.TimestampSegment) {
	if len(studentTimeSegments) == 0 || len(siblingTimeSegments) == 0 {
		return
	}

	var (
		tmpSiblingTimeSegment []entities.TimestampSegment
	)

	overlappedTimeSegments = []entities.TimestampSegment{}
	tmpSiblingTimeSegment = siblingTimeSegments

	for _, studentTimeSegment := range studentTimeSegments {
		tmpTimeSegment := entities.TimestampSegment{
			StartDate: studentTimeSegment.StartDate,
			EndDate:   studentTimeSegment.EndDate,
		}

		overlapCheck := false
		for _, siblingTimeSegment := range tmpSiblingTimeSegment {
			if siblingTimeSegment.EndDate.Before(tmpTimeSegment.StartDate) || siblingTimeSegment.StartDate.After(tmpTimeSegment.EndDate) {
				continue
			}

			if tmpTimeSegment.StartDate.Before(siblingTimeSegment.StartDate) {
				tmpTimeSegment.StartDate = siblingTimeSegment.StartDate
			}

			if tmpTimeSegment.EndDate.After(siblingTimeSegment.EndDate) {
				tmpTimeSegment.EndDate = siblingTimeSegment.EndDate

				overlappedTimeSegments = append(overlappedTimeSegments, tmpTimeSegment)
				tmpTimeSegment.EndDate = studentTimeSegment.EndDate

				overlapCheck = false
				continue
			}

			overlapCheck = true
		}

		if overlapCheck {
			overlappedTimeSegments = append(overlappedTimeSegments, tmpTimeSegment)
		}
	}

	return
}

func (s *DiscountService) TrackStudentDiscount(ctx context.Context, productDiscountGroup []entities.ProductDiscountGroup) (tracked bool, err error) {
	tracked = false
	for _, productDiscountGroup := range productDiscountGroup {
		for _, productGroup := range productDiscountGroup.ProductGroups {
			// skip tracking of outdated and cancelled products
			if productDiscountGroup.StudentProduct.ProductStatus.String != paymentPb.StudentProductStatus_ORDERED.String() &&
				productDiscountGroup.StudentProduct.StudentProductLabel.String != paymentPb.StudentProductLabel_CREATED.String() {
				continue
			}

			discountTracker := entities.StudentDiscountTracker{}
			err = utils.GroupErrorFunc(
				discountTracker.StudentID.Set(productDiscountGroup.StudentProduct.StudentID),
				discountTracker.LocationID.Set(productDiscountGroup.StudentProduct.LocationID),
				discountTracker.StudentProductID.Set(productDiscountGroup.StudentProduct.StudentProductID),
				discountTracker.ProductID.Set(productDiscountGroup.StudentProduct.ProductID),
				discountTracker.ProductGroupID.Set(productGroup.ProductGroupID.String),
				discountTracker.DiscountType.Set(productGroup.DiscountType.String),
				discountTracker.DiscountStatus.Set(nil),
				discountTracker.DiscountStartDate.Set(nil),
				discountTracker.DiscountEndDate.Set(nil),
				discountTracker.StudentProductStartDate.Set(productDiscountGroup.StudentProduct.StartDate),
				discountTracker.StudentProductEndDate.Set(productDiscountGroup.StudentProduct.EndDate),
				discountTracker.StudentProductStatus.Set(productDiscountGroup.StudentProduct.ProductStatus),
				discountTracker.UpdatedFromDiscountTrackerID.Set(nil),
				discountTracker.UpdatedToDiscountTrackerID.Set(nil),
			)
			if err != nil {
				return
			}

			err = s.DiscountTrackerService.TrackDiscount(ctx, s.DB, &discountTracker)
			if err != nil {
				return
			}

			tracked = true
			s.Logger.Info(fmt.Sprintf("Student %v tracked for for %v for student product %v",
				discountTracker.StudentID.String,
				discountTracker.DiscountType.String,
				productDiscountGroup.StudentProduct.StudentProductID))
		}
	}

	return
}

func (s *DiscountService) UpdateProductDiscountTrackingData(ctx context.Context, studentProductDiscountGroups []entities.ProductDiscountGroup) (tracked bool, err error) {
	for _, studentProductDiscountGroup := range studentProductDiscountGroups {
		var (
			newStudentProduct entities.StudentProduct
			oldStudentProduct entities.StudentProduct
		)

		newStudentProduct = studentProductDiscountGroup.StudentProduct
		if newStudentProduct.UpdatedFromStudentProductID.Status == pgtype.Present {
			oldStudentProduct, err = s.StudentProductService.RetrieveStudentProductByID(ctx, s.DB, newStudentProduct.UpdatedFromStudentProductID.String)
			if err != nil {
				return
			}
		} else {
			oldStudentProduct = newStudentProduct
		}

		tracked = true
		err = s.DiscountTrackerService.UpdateTrackingDurationByStudentProduct(ctx, s.DB, oldStudentProduct)
		if err != nil {
			return
		}
	}

	return
}

func NewDiscountService(db database.Ext, jsm nats.JetStreamManagement, logger *zap.Logger) *DiscountService {
	return &DiscountService{
		DB:                     db,
		JSM:                    jsm,
		Logger:                 logger,
		DiscountTagService:     domainService.NewDiscountTagService(db),
		DiscountTrackerService: domainService.NewDiscountTrackerService(db),
		ProductGroupService:    domainService.NewProductGroupService(db),
		StudentProductService:  domainService.NewStudentProductService(db),
		StudentSiblingService:  domainService.NewStudentSiblingService(db),
		DiscountRepo:           &repositories.DiscountRepo{},
	}
}
