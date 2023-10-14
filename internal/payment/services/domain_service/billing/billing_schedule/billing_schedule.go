package billing_schedule

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/repositories"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BillingScheduleService struct {
	BillingScheduleRepo interface {
		GetByIDForUpdate(ctx context.Context, db database.QueryExecer, billingScheduleID string) (entities.BillingSchedule, error)
	}
	BillingSchedulePeriodRepo interface {
		GetByIDForUpdate(ctx context.Context, db database.QueryExecer, billingSchedulePeriodID string) (entities.BillingSchedulePeriod, error)
		GetPeriodIDsByScheduleIDAndStartTimeForUpdate(ctx context.Context, db database.QueryExecer, billingScheduleID string, startTime time.Time) ([]pgtype.Text, error)
		GetPeriodIDsInRangeTimeByScheduleID(ctx context.Context, db database.QueryExecer, billingScheduleID string, startTime time.Time, endTime time.Time) ([]pgtype.Text, error)
		GetAllBillingPeriodsByBillingScheduleID(ctx context.Context, db database.QueryExecer, billingScheduleID string) ([]entities.BillingSchedulePeriod, error)
		GetNextBillingSchedulePeriod(ctx context.Context, db database.QueryExecer, billingScheduleID string, endTime time.Time) (entities.BillingSchedulePeriod, error)
		GetLatestBillingSchedulePeriod(
			ctx context.Context,
			db database.QueryExecer,
			BillingScheduleID string,
		) (
			latestBillingSchedulePeriod entities.BillingSchedulePeriod,
			err error,
		)
	}
	BillingRatioRepo interface {
		GetFirstRatioByBillingSchedulePeriodIDAndFromTime(
			ctx context.Context,
			db database.QueryExecer,
			billingSchedulePeriodID string,
			from time.Time,
		) (entities.BillingRatio, error)
		GetNextRatioByBillingSchedulePeriodIDAndPrevious(
			ctx context.Context,
			db database.QueryExecer,
			ratioOfProRatedBillingItem entities.BillingRatio,
		) (entities.BillingRatio, error)
	}
}

func (s *BillingScheduleService) CheckScheduleReturnProRatedItemAndMapPeriodInfo(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (
	proRatedBillItem utils.BillingItemData,
	ratioOfProRatedBillingItem entities.BillingRatio,
	normalBillItem []utils.BillingItemData,
	mapPeriodInfo map[string]entities.BillingSchedulePeriod,
	err error,
) {
	if !orderItemData.IsDisableProRatingFlag {
		return s.checkScheduleWithProductNoneDisableProRating(ctx, db, orderItemData)
	}
	normalBillItem, mapPeriodInfo, err = s.checkScheduleWithProductDisableProRating(ctx, db, orderItemData)

	return
}

func (s *BillingScheduleService) checkScheduleWithProductNoneDisableProRating(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (
	proRatedBillItem utils.BillingItemData,
	ratioOfProRatedBillingItem entities.BillingRatio,
	normalBillItem []utils.BillingItemData,
	mapPeriodInfo map[string]entities.BillingSchedulePeriod,
	err error,
) {
	var (
		ok                       bool
		minStartTime             *time.Time
		maxEndTime               *time.Time
		startTimeToCheckRange    time.Time
		proRatedBillItemPeriod   *entities.BillingSchedulePeriod
		countingUpcomingBillItem int
	)

	if len(orderItemData.BillItems) == 0 {
		err = status.Errorf(codes.FailedPrecondition, "Empty bill item in request")
		return
	}

	mapPeriodInfo = make(map[string]entities.BillingSchedulePeriod, len(orderItemData.BillItems))
	if orderItemData.OrderItem.StartDate != nil {
		startTimeToCheckRange = orderItemData.OrderItem.StartDate.AsTime()
	} else {
		startTimeToCheckRange = orderItemData.OrderItem.EffectiveDate.AsTime()
	}
	orderItemType := utils.ConvertOrderItemType(pb.OrderType(pb.OrderType_value[orderItemData.Order.OrderType.String]), orderItemData.BillItems[0].BillingItem)
	for i := range orderItemData.BillItems {
		item := orderItemData.BillItems[i]
		var billingSchedulePeriod entities.BillingSchedulePeriod
		if item.BillingItem.BillingSchedulePeriodId == nil {
			err = utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.BillItemHasNoSchedulePeriodID,
				&errdetails.DebugInfo{Detail: fmt.Sprintf(constant.BillItemHasNoSchedulePeriodIDDebugMsg, i)},
			)
			return
		}
		if _, ok = mapPeriodInfo[item.BillingItem.BillingSchedulePeriodId.Value]; ok {
			err = status.Error(codes.FailedPrecondition, "Bill item has duplicate billing schedule period ID")
			return
		}

		billingSchedulePeriod, err = s.isBillingSchedulePeriodValidAndReturnBillingSchedulePeriod(
			ctx,
			db,
			orderItemData.ProductInfo,
			item.BillingItem,
		)
		if err != nil {
			return
		}
		mapPeriodInfo[item.BillingItem.BillingSchedulePeriodId.Value] = billingSchedulePeriod

		if time.Now().Before(billingSchedulePeriod.BillingDate.Time) {
			if !item.IsUpcoming {
				err = status.Errorf(codes.FailedPrecondition, "This bill item should be in upcoming billing")
				return
			}
			countingUpcomingBillItem++
		} else {
			if item.IsUpcoming {
				err = status.Errorf(codes.FailedPrecondition, "This bill item should be in at order billing")
				return
			}
		}

		if minStartTime == nil || minStartTime.After(billingSchedulePeriod.StartDate.Time) {
			minStartTime = &billingSchedulePeriod.StartDate.Time
			proRatedBillItem = orderItemData.BillItems[i]
			proRatedBillItemPeriod = &billingSchedulePeriod
		}
		if maxEndTime == nil || maxEndTime.Before(billingSchedulePeriod.EndDate.Time) {
			maxEndTime = &billingSchedulePeriod.EndDate.Time
		}
	}

	if proRatedBillItemPeriod != nil {
		if proRatedBillItemPeriod.StartDate.Time.After(startTimeToCheckRange) ||
			proRatedBillItemPeriod.EndDate.Time.Before(startTimeToCheckRange) {
			err = status.Errorf(codes.FailedPrecondition, "Start date of product is outside of selected billing schedule period")
			return
		}
		ratioOfProRatedBillingItem, err = s.BillingRatioRepo.GetFirstRatioByBillingSchedulePeriodIDAndFromTime(ctx, db, proRatedBillItemPeriod.BillingSchedulePeriodID.String, startTimeToCheckRange)
		if err != nil {
			err = status.Errorf(codes.Internal, "Error when get product ratio of product %v with error %s", orderItemData.ProductInfo.ProductID.String, err.Error())
			return
		}
		if orderItemType == utils.OrderCancel || orderItemType == utils.OrderWithdraw || orderItemType == utils.OrderLOA || orderItemType == utils.OrderGraduate {
			ratioOfProRatedBillingItem, err = s.BillingRatioRepo.GetNextRatioByBillingSchedulePeriodIDAndPrevious(ctx, db, ratioOfProRatedBillingItem)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				err = status.Errorf(codes.Internal, "Error when get next product ratio of product %v with error %s", orderItemData.ProductInfo.ProductID.String, err.Error())
				return
			}
		}
	}

	if countingUpcomingBillItem == 0 {
		err = s.isReachLastPeriodOfSchedule(ctx, db, *maxEndTime, orderItemData.ProductInfo.BillingScheduleID.String)
		if err != nil {
			return
		}
	} else if countingUpcomingBillItem > 1 {
		err = status.Error(codes.FailedPrecondition, "Upcoming billing should only contain one item")
		return
	}
	err = s.isContinuePeriodOfScheduleValid(ctx, db, *minStartTime, *maxEndTime, orderItemData.ProductInfo.BillingScheduleID.String, len(mapPeriodInfo))

	for i, item := range orderItemData.BillItems {
		if item.BillingItem.BillingSchedulePeriodId.Value != proRatedBillItemPeriod.BillingSchedulePeriodID.String {
			normalBillItem = append(normalBillItem, orderItemData.BillItems[i])
		}
	}

	return
}

func (s *BillingScheduleService) checkScheduleWithProductDisableProRating(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData,
) (
	normalBillItem []utils.BillingItemData,
	mapPeriodInfoWithID map[string]entities.BillingSchedulePeriod,
	err error,
) {
	var (
		ok                       bool
		minStartTime             *time.Time
		maxEndTime               *time.Time
		startTimeToCheckRange    time.Time
		proRatedBillItemPeriod   *entities.BillingSchedulePeriod
		countingUpcomingBillItem int
	)

	if len(orderItemData.BillItems) == 0 {
		err = status.Errorf(codes.FailedPrecondition, "Empty bill item in request")
		return
	}

	mapPeriodInfoWithID = make(map[string]entities.BillingSchedulePeriod, len(orderItemData.BillItems))
	if orderItemData.OrderItem.StartDate != nil {
		startTimeToCheckRange = orderItemData.OrderItem.StartDate.AsTime()
	} else {
		startTimeToCheckRange = orderItemData.OrderItem.EffectiveDate.AsTime()
	}

	for i := range orderItemData.BillItems {
		item := orderItemData.BillItems[i]
		var billingSchedulePeriod entities.BillingSchedulePeriod
		if item.BillingItem.BillingSchedulePeriodId == nil {
			err = utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.BillItemHasNoSchedulePeriodID,
				&errdetails.DebugInfo{Detail: fmt.Sprintf(constant.BillItemHasNoSchedulePeriodIDDebugMsg, i)},
			)
			return
		}
		if _, ok = mapPeriodInfoWithID[item.BillingItem.BillingSchedulePeriodId.Value]; ok {
			err = status.Error(codes.FailedPrecondition, "Bill item has duplicate billing schedule period ID")
			return
		}

		billingSchedulePeriod, err = s.isBillingSchedulePeriodValidAndReturnBillingSchedulePeriod(
			ctx,
			db,
			orderItemData.ProductInfo,
			item.BillingItem,
		)
		if err != nil {
			return
		}
		mapPeriodInfoWithID[item.BillingItem.BillingSchedulePeriodId.Value] = billingSchedulePeriod

		if time.Now().Before(billingSchedulePeriod.BillingDate.Time) {
			if !item.IsUpcoming {
				err = status.Errorf(codes.FailedPrecondition, "This bill item should be in upcoming billing")
				return
			}
			countingUpcomingBillItem++
		} else {
			if item.IsUpcoming {
				err = status.Errorf(codes.FailedPrecondition, "This bill item should be in at order billing")
				return
			}
		}
		normalBillItem = append(normalBillItem, orderItemData.BillItems[i])
		if minStartTime == nil || minStartTime.After(billingSchedulePeriod.StartDate.Time) {
			minStartTime = &billingSchedulePeriod.StartDate.Time
			proRatedBillItemPeriod = &billingSchedulePeriod
		}
		if maxEndTime == nil || maxEndTime.Before(billingSchedulePeriod.EndDate.Time) {
			maxEndTime = &billingSchedulePeriod.EndDate.Time
		}
	}

	if proRatedBillItemPeriod != nil {
		if proRatedBillItemPeriod.StartDate.Time.After(startTimeToCheckRange) ||
			proRatedBillItemPeriod.EndDate.Time.Before(startTimeToCheckRange) {
			err = status.Errorf(codes.FailedPrecondition, "Start date of product is outside of selected billing schedule period")
			return
		}
	}

	if countingUpcomingBillItem == 0 {
		err = s.isReachLastPeriodOfSchedule(ctx, db, *maxEndTime, orderItemData.ProductInfo.BillingScheduleID.String)
		if err != nil {
			return
		}
	} else if countingUpcomingBillItem > 1 {
		err = status.Error(codes.FailedPrecondition, "Upcoming billing should only contain one item")
		return
	}

	err = s.isContinuePeriodOfScheduleValid(ctx, db, *minStartTime, *maxEndTime, orderItemData.ProductInfo.BillingScheduleID.String, len(mapPeriodInfoWithID))
	return
}

func (s *BillingScheduleService) isBillingScheduleValid(ctx context.Context, db database.QueryExecer, orderItemData utils.OrderItemData) (err error) {
	var billingSchedule entities.BillingSchedule
	billingSchedule, err = s.BillingScheduleRepo.GetByIDForUpdate(ctx, db, orderItemData.ProductInfo.BillingScheduleID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, "Fail to get billing schedule:: %v", err.Error())
		return
	}
	if billingSchedule.IsArchived.Bool {
		err = status.Error(codes.FailedPrecondition, "Selected billing schedule is removed or archived")
	}
	return
}

func (s *BillingScheduleService) isBillingSchedulePeriodValidAndReturnBillingSchedulePeriod(
	ctx context.Context,
	db database.QueryExecer,
	product entities.Product,
	billItem *pb.BillingItem,
) (billingSchedulePeriod entities.BillingSchedulePeriod, err error) {
	if billItem.BillingSchedulePeriodId == nil {
		err = utils.StatusErrWithDetail(
			codes.FailedPrecondition,
			constant.BillItemHasNoSchedulePeriodID,
			nil,
		)
		return
	}
	billingSchedulePeriod, err = s.BillingSchedulePeriodRepo.GetByIDForUpdate(ctx, db, billItem.BillingSchedulePeriodId.Value)
	if err != nil {
		err = status.Errorf(codes.Internal, "Fail to get billing schedule:: %v", err.Error())
		return
	}
	if billingSchedulePeriod.BillingScheduleID.String != product.BillingScheduleID.String {
		err = status.Errorf(codes.FailedPrecondition, "Billing schedule %v in system does not match billing schedule in product", billingSchedulePeriod.BillingSchedulePeriodID.String)
		return
	}
	if product.AvailableFrom.Time.After(billingSchedulePeriod.StartDate.Time) ||
		product.AvailableUntil.Time.Before(billingSchedulePeriod.EndDate.Time) {
		err = status.Errorf(codes.FailedPrecondition, "Billing schedule period %v has invalid time range", billingSchedulePeriod.BillingSchedulePeriodID.String)
		return
	}
	return
}

func (s *BillingScheduleService) isReachLastPeriodOfSchedule(
	ctx context.Context,
	db database.QueryExecer,
	startTime time.Time,
	scheduleID string,
) (err error) {
	periodIDs, err := s.BillingSchedulePeriodRepo.GetPeriodIDsByScheduleIDAndStartTimeForUpdate(ctx, db, scheduleID, startTime)
	if err != nil {
		err = status.Errorf(codes.Internal, "Fail to get billing schedule period by schedule ID: %s", err.Error())
		return
	}
	if len(periodIDs) > 0 {
		err = status.Errorf(codes.FailedPrecondition, "Missing upcoming billing last billing schedule period not reached yet")
	}
	return
}

func (s *BillingScheduleService) isContinuePeriodOfScheduleValid(
	ctx context.Context,
	db database.QueryExecer,
	startTime time.Time,
	endTime time.Time,
	scheduleID string,
	lenContinuePeriodOfSchedule int,
) (err error) {
	periodIDs, err := s.BillingSchedulePeriodRepo.GetPeriodIDsInRangeTimeByScheduleID(ctx, db, scheduleID, startTime, endTime)
	if err != nil {
		err = status.Errorf(codes.Internal, "Fail to get billing schedule periods: %s", err.Error())
		return
	}
	if len(periodIDs) != lenContinuePeriodOfSchedule {
		err = status.Errorf(codes.FailedPrecondition, "Number of periods retrieved and number of continue period of schedule does not match for time range %v to %v for billing schedule %v", startTime, endTime, scheduleID)
		return
	}
	return
}

func (s *BillingScheduleService) GetBillingSchedulePeriodByID(
	ctx context.Context,
	db database.QueryExecer,
	billingSchedulePeriodID string,
) (billingSchedulePeriod entities.BillingSchedulePeriod, err error) {
	billingSchedulePeriod, err = s.BillingSchedulePeriodRepo.GetByIDForUpdate(ctx, db, billingSchedulePeriodID)
	if err != nil {
		err = status.Errorf(codes.Internal, "Fail to get billing schedule period by ID: %v", err.Error())
	}

	return
}

func (s *BillingScheduleService) GetAllBillingPeriodsByBillingScheduleID(
	ctx context.Context,
	db database.QueryExecer,
	billingScheduleID string,
) (billingSchedulePeriods []entities.BillingSchedulePeriod, err error) {
	billingSchedulePeriods, err = s.BillingSchedulePeriodRepo.GetAllBillingPeriodsByBillingScheduleID(ctx, db, billingScheduleID)
	if err != nil {
		err = status.Errorf(codes.Internal, "Fail to get billing schedule period by schedule ID: %v", err.Error())
		return []entities.BillingSchedulePeriod{}, err
	}

	return billingSchedulePeriods, nil
}

func (s *BillingScheduleService) GetNextBillingSchedulePeriod(ctx context.Context, db database.QueryExecer, billingScheduleID string, endTime time.Time) (nextPeriod entities.BillingSchedulePeriod, err error) {
	nextPeriod, err = s.BillingSchedulePeriodRepo.GetNextBillingSchedulePeriod(ctx, db, billingScheduleID, endTime)
	if err != nil {
		err = fmt.Errorf("error while getting next billing period %v, with err: %v", billingScheduleID, err)
	}
	return
}

func (s *BillingScheduleService) GetLatestBillingSchedulePeriod(ctx context.Context, db database.QueryExecer, billingScheduleID string) (latestBillingSchedulePeriod entities.BillingSchedulePeriod, err error) {
	latestBillingSchedulePeriod, err = s.BillingSchedulePeriodRepo.GetLatestBillingSchedulePeriod(ctx, db, billingScheduleID)
	if err != nil {
		err = fmt.Errorf("err while getting latest bill schedule period: %v", err)
	}
	return
}

func NewBillingScheduleService() *BillingScheduleService {
	return &BillingScheduleService{
		BillingScheduleRepo:       &repositories.BillingScheduleRepo{},
		BillingSchedulePeriodRepo: &repositories.BillingSchedulePeriodRepo{},
		BillingRatioRepo:          &repositories.BillingRatioRepo{},
	}
}
