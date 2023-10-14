package service

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/repositories"
	"github.com/manabie-com/backend/internal/payment/utils"

	"github.com/jackc/pgtype"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StudentPackageOrderRepo interface {
	Create(ctx context.Context, db database.QueryExecer, studentPackageOrder entities.StudentPackageOrder) (err error)
	Update(ctx context.Context, db database.QueryExecer, e entities.StudentPackageOrder) error
	ResetCurrentPosition(ctx context.Context, db database.QueryExecer, studentPackageID string) (err error)
	GetStudentPackageOrdersByStudentPackageID(
		ctx context.Context,
		db database.QueryExecer,
		studentPackageID string,
	) (studentPackageOrders []*entities.StudentPackageOrder, err error)
	SetCurrentStudentPackageByID(ctx context.Context, db database.QueryExecer, id string, isCurrent bool) error
	GetStudentPackageOrderByTimeAndStudentPackageID(ctx context.Context, db database.QueryExecer, studentPackageID string, startTime time.Time) (studentPackageOrder *entities.StudentPackageOrder, err error)
	GetByStudentPackageIDAndOrderID(ctx context.Context, db database.QueryExecer, studentPackageID, orderID string) (studentPackageOrder *entities.StudentPackageOrder, err error)
	SoftDeleteByID(ctx context.Context, db database.QueryExecer, id string) error
	RevertByID(ctx context.Context, db database.QueryExecer, id string) error
	Upsert(ctx context.Context, tx database.QueryExecer, studentPackageOrder entities.StudentPackageOrder,
	) (
		err error,
	)
	GetByStudentPackageOrderID(ctx context.Context, db database.QueryExecer, studentPackageOrderID string,
	) (studentPackageOrder *entities.StudentPackageOrder, err error)
	UpdateExecuteStatus(ctx context.Context, db database.QueryExecer, studentPackageOrder entities.StudentPackageOrder,
	) (err error)
	UpdateExecuteError(ctx context.Context, db database.QueryExecer, studentPackageOrder entities.StudentPackageOrder,
	) (err error)
}

type StudentPackageOrderService struct {
	studentPackageOrderRepo StudentPackageOrderRepo
}

func NewStudentPackageOrder() *StudentPackageOrderService {
	return &StudentPackageOrderService{
		studentPackageOrderRepo: &repositories.StudentPackageOrderRepo{},
	}
}

// GetPositionForStudentPackageByTime
// Used to get position by startTime, endTime of new student package order in available student package order list
// Used to check if time of student package order is overlap with other student package orders time
// Enums : [PastStudentPackage, CurrentStudentPackage, FutureStudentPackage]
func (s *StudentPackageOrderService) GetPositionForStudentPackageByTime(
	ctx context.Context,
	db database.QueryExecer,
	studentPackageID string,
	startTime time.Time,
	endTime time.Time,
) (
	studentPackagePosition entities.StudentPackagePosition,
	err error,
) {
	var (
		now                        = time.Now()
		studentPackageOrders       []*entities.StudentPackageOrder
		studentPackageOrderLen     int
		currentStudentPackageOrder *entities.StudentPackageOrder
	)

	studentPackageOrders, err = s.studentPackageOrderRepo.GetStudentPackageOrdersByStudentPackageID(ctx, db, studentPackageID)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get student package orders by student_package_id = %v: %v", studentPackageID, err)
		return
	}

	newSPOTimeRange := utils.TimeRange{
		FromTime: startTime,
		ToTime:   endTime,
	}
	currentStudentPackageOrder, isOverlap := s.checkTimeOverlapAndReturnCurrentStudentPackageOrder(studentPackageOrders, newSPOTimeRange)
	if isOverlap {
		err = utils.StatusErrWithDetail(
			codes.FailedPrecondition,
			constant.DuplicateCourses, &errdetails.DebugInfo{
				Detail: fmt.Sprintf("wrong start time in this student package id %v, student package end date %v, start time %v",
					studentPackageID, endTime, startTime),
			},
		)
		return
	}

	studentPackageOrderLen = len(studentPackageOrders)
	if studentPackageOrderLen == 0 {
		studentPackagePosition = entities.CurrentStudentPackage
		return
	}
	if studentPackageOrderLen > 0 {
		currentStudentPackageOrderTimeRange := utils.TimeRange{
			FromTime: currentStudentPackageOrder.StartAt.Time,
			ToTime:   currentStudentPackageOrder.EndAt.Time,
		}

		if utils.WithinRangeTime(now, newSPOTimeRange) {
			studentPackagePosition = entities.CurrentStudentPackage
			return
		}

		if utils.BeforeRangeTime(now, currentStudentPackageOrderTimeRange) {
			if utils.BeforeRangeTime(now, newSPOTimeRange) {
				// In case: now < newSPOTimeRange < currentSPOTTimeRange
				if utils.BeforeRangeTime(newSPOTimeRange.ToTime, currentStudentPackageOrderTimeRange) {
					studentPackagePosition = entities.CurrentStudentPackage
					return
				}
				// In case: now < currentSPOTTimeRange < newSPOTimeRange
				if utils.AfterRangeTime(newSPOTimeRange.FromTime, currentStudentPackageOrderTimeRange) {
					studentPackagePosition = entities.FutureStudentPackage
					return
				}
			}
			// In case: newSPOTimeRange < now < currentSPOTimeRange
			if utils.AfterRangeTime(now, newSPOTimeRange) {
				studentPackagePosition = entities.PastStudentPackage
				return
			}
		}

		if utils.WithinRangeTime(now, currentStudentPackageOrderTimeRange) {
			// In case: now ∈ currentSPOTimeRange < newSPOTimeRange
			if utils.BeforeRangeTime(now, newSPOTimeRange) {
				studentPackagePosition = entities.FutureStudentPackage
				return
			}
			// In case: newSPOTimeRange < now ∈ currentSPOTimeRange
			if utils.AfterRangeTime(now, newSPOTimeRange) {
				studentPackagePosition = entities.PastStudentPackage
				return
			}
		}

		if utils.AfterRangeTime(now, currentStudentPackageOrderTimeRange) {
			// In case: newSPOTimeRange < currentSPOTimeRange < now
			if utils.BeforeRangeTime(newSPOTimeRange.ToTime, currentStudentPackageOrderTimeRange) {
				studentPackagePosition = entities.PastStudentPackage
				return
			}
			// In case: currentSPOTimeRange < newSPOTimeRange < now
			if utils.AfterRangeTime(newSPOTimeRange.FromTime, currentStudentPackageOrderTimeRange) {
				studentPackagePosition = entities.CurrentStudentPackage
				return
			}
			return
		}
	}
	return
}

func (s *StudentPackageOrderService) SetCurrentStudentPackageOrderByTimeAndStudentPackageID(
	ctx context.Context,
	db database.QueryExecer,
	studentPackageID string,
) (
	studentPackageOrder *entities.StudentPackageOrder,
	err error,
) {
	studentPackageOrders, err := s.studentPackageOrderRepo.GetStudentPackageOrdersByStudentPackageID(ctx, db, studentPackageID)
	if err != nil {
		return
	}

	studentPackageOrderLen := len(studentPackageOrders)
	switch studentPackageOrderLen {
	case 0:
		return
	case 1:
		studentPackageOrder = studentPackageOrders[0]
	default:
		now := time.Now()
		for i, tmpStudentPackageOrder := range studentPackageOrders { // this array was order by start_at
			tmpSPOStartDate := tmpStudentPackageOrder.StartAt.Time
			tmpStudentPackageOrderTimeRange := utils.TimeRange{
				FromTime: tmpStudentPackageOrder.StartAt.Time,
				ToTime:   tmpStudentPackageOrder.EndAt.Time,
			}
			isNowWithinStudentPackageOrder := utils.WithinRangeTime(now, tmpStudentPackageOrderTimeRange)
			isNowBeforeFirstStudentPackageOrder := i == 0 && utils.BeforeRangeTime(now, tmpStudentPackageOrderTimeRange)
			isNowAfterLastStudentPackageOrder := i == studentPackageOrderLen-1 && utils.AfterRangeTime(now, tmpStudentPackageOrderTimeRange)
			isNowNearestFutureStudentPackageOrder := tmpSPOStartDate.Unix()-now.Unix() >= 0

			if isNowWithinStudentPackageOrder ||
				isNowBeforeFirstStudentPackageOrder ||
				isNowAfterLastStudentPackageOrder ||
				isNowNearestFutureStudentPackageOrder {
				studentPackageOrder = studentPackageOrders[i]
				break
			}
		}
	}
	err = utils.GroupErrorFunc(
		studentPackageOrder.IsCurrentStudentPackage.Set(true),
		s.studentPackageOrderRepo.ResetCurrentPosition(ctx, db, studentPackageOrder.StudentPackageID.String),
		s.studentPackageOrderRepo.SetCurrentStudentPackageByID(ctx, db, studentPackageOrder.ID.String, studentPackageOrder.IsCurrentStudentPackage.Bool),
	)
	if err != nil {
		return
	}
	return
}

// GetCurrentStudentPackageOrderByStudentPackageID
// Used to get current student package order in student package order list in db
// Return nil when there in no available student package orders fb
func (s *StudentPackageOrderService) GetCurrentStudentPackageOrderByStudentPackageID(
	ctx context.Context,
	db database.QueryExecer,
	studentPackageID string,
) (
	currentStudentPackageOrder *entities.StudentPackageOrder,
	err error,
) {
	studentPackageOrders, err := s.studentPackageOrderRepo.GetStudentPackageOrdersByStudentPackageID(ctx, db, studentPackageID)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when get student package orders by student_package_id = %v: %v", studentPackageID, err)
		return
	}

	for _, studentPackageOrder := range studentPackageOrders {
		if studentPackageOrder.IsCurrentStudentPackage.Status == pgtype.Present && studentPackageOrder.IsCurrentStudentPackage.Bool {
			currentStudentPackageOrder = studentPackageOrder
			return
		}
	}
	return
}

func (s *StudentPackageOrderService) InsertStudentPackageOrder(
	ctx context.Context,
	db database.QueryExecer,
	studentPackageOrder entities.StudentPackageOrder,
	positionStudentPackageOrder entities.StudentPackagePosition,
) (err error) {
	switch positionStudentPackageOrder {
	case entities.PastStudentPackage:
	case entities.CurrentStudentPackage:
		err = s.studentPackageOrderRepo.ResetCurrentPosition(ctx, db, studentPackageOrder.StudentPackageID.String)
		if err != nil {
			err = status.Errorf(codes.Internal, "reset current position with student_package_id = %s have error: %v", studentPackageOrder.StudentPackageID.String, err.Error())
			return
		}
		_ = studentPackageOrder.IsCurrentStudentPackage.Set(true)
	case entities.FutureStudentPackage:
	}
	err = s.studentPackageOrderRepo.Upsert(ctx, db, studentPackageOrder)
	if err != nil {
		err = status.Errorf(codes.Internal, "insert student package order have error : %v", err.Error())
		return
	}
	return
}

func (s *StudentPackageOrderService) checkTimeOverlapAndReturnCurrentStudentPackageOrder(
	studentPackageOrders []*entities.StudentPackageOrder,
	newStudentPackageOrderTimeRange utils.TimeRange,
) (currentStudentPackageOrder *entities.StudentPackageOrder, isOverlap bool) {
	for _, studentPackageOrder := range studentPackageOrders {
		if studentPackageOrder.IsCurrentStudentPackage.Status == pgtype.Present && studentPackageOrder.IsCurrentStudentPackage.Bool {
			currentStudentPackageOrder = studentPackageOrder
		}
		isOverlap = utils.CheckTimeRangeOverlap(newStudentPackageOrderTimeRange, utils.TimeRange{
			FromTime: studentPackageOrder.StartAt.Time,
			ToTime:   studentPackageOrder.EndAt.Time,
		})
		if isOverlap {
			return
		}
	}
	return
}

func (s *StudentPackageOrderService) GetStudentPackageOrderByStudentPackageIDAndTime(
	ctx context.Context,
	db database.QueryExecer,
	studentPackageID string,
	newStartTime time.Time,
) (
	studentPackageOrder *entities.StudentPackageOrder,
	err error,
) {
	var (
		studentPackageOrders []*entities.StudentPackageOrder
	)

	studentPackageOrders, err = s.studentPackageOrderRepo.GetStudentPackageOrdersByStudentPackageID(ctx, db, studentPackageID)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when getting student package order by student_package_id = %s and start time: %v", studentPackageID, err.Error())
		return
	}
	for _, spo := range studentPackageOrders {
		// In case: start_at <= new_start_at <= end_at
		if spo.StartAt.Status == pgtype.Present &&
			spo.EndAt.Status == pgtype.Present &&
			!spo.StartAt.Time.After(newStartTime) &&
			!spo.EndAt.Time.Before(newStartTime) {
			studentPackageOrder = spo
		}
	}
	return
}

func (s *StudentPackageOrderService) GetStudentPackageOrderByStudentPackageIDAndOrderID(
	ctx context.Context,
	db database.QueryExecer,
	studentPackageID, orderID string,
) (
	studentPackageOrder *entities.StudentPackageOrder,
	err error,
) {
	studentPackageOrder, err = s.studentPackageOrderRepo.GetByStudentPackageIDAndOrderID(ctx, db, studentPackageID, orderID)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when getting student package order by student package id and order id: %v", err.Error())
		return
	}
	return
}

func (s *StudentPackageOrderService) DeleteStudentPackageOrderByID(
	ctx context.Context,
	db database.QueryExecer,
	studentPackageOrderID string,
) (
	err error,
) {
	err = s.studentPackageOrderRepo.SoftDeleteByID(ctx, db, studentPackageOrderID)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when deleting student package order by id : %v", err.Error())
		return
	}
	return
}

func (s *StudentPackageOrderService) RevertStudentPackageOrderByStudentPackageOrderID(
	ctx context.Context,
	db database.QueryExecer,
	studentPackageOrderID string,
) (
	err error,
) {
	err = s.studentPackageOrderRepo.RevertByID(ctx, db, studentPackageOrderID)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when reverting student package order by id : %v", err.Error())
		return
	}
	return
}

func (s *StudentPackageOrderService) UpdateStudentPackageOrder(
	ctx context.Context,
	db database.QueryExecer,
	studentPackageOrder entities.StudentPackageOrder,
) (err error) {
	err = s.studentPackageOrderRepo.Update(ctx, db, studentPackageOrder)
	if err != nil {
		err = status.Errorf(codes.Internal, "insert student package order have error : %v", err.Error())
		return
	}
	return
}

func (s *StudentPackageOrderService) GetStudentPackageOrderByStudentPackageOrderID(
	ctx context.Context,
	db database.QueryExecer,
	studentPackageOrderID string,
) (
	studentPackageOrder *entities.StudentPackageOrder,
	err error,
) {
	studentPackageOrder, err = s.studentPackageOrderRepo.GetByStudentPackageOrderID(ctx, db, studentPackageOrderID)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when getting student package order by student package order id: %v", err.Error())
		return
	}
	return
}

func (s *StudentPackageOrderService) UpdateExecuteStatus(ctx context.Context, db database.QueryExecer, studentPackageOrder entities.StudentPackageOrder) (err error) {
	err = s.studentPackageOrderRepo.UpdateExecuteStatus(ctx, db, studentPackageOrder)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when update execute status of student package order with error: %v", err.Error())
	}
	return
}

func (s *StudentPackageOrderService) UpdateExecuteError(ctx context.Context, db database.QueryExecer, studentPackageOrder entities.StudentPackageOrder) (err error) {
	err = s.studentPackageOrderRepo.UpdateExecuteError(ctx, db, studentPackageOrder)
	if err != nil {
		err = status.Errorf(codes.Internal, "error when update execute error of student package order with error: %v", err.Error())
	}
	return
}
