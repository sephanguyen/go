package timesheet

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	nats_service_utils "github.com/manabie-com/backend/internal/timesheet/service/nats"
	pbc "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ConfirmationWindowServiceImpl struct {
	DB  database.Ext
	JSM nats.JetStreamManagement

	TimesheetConfirmationPeriodRepo interface {
		GetPeriodByDate(ctx context.Context, db database.QueryExecer, date time.Time) (*entity.TimesheetConfirmationPeriod, error)
		InsertPeriod(ctx context.Context, db database.QueryExecer, period *entity.TimesheetConfirmationPeriod) (*entity.TimesheetConfirmationPeriod, error)
		GetPeriodByID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entity.TimesheetConfirmationPeriod, error)
	}

	TimesheetConfirmationCutOffDateRepo interface {
		GetCutOffDateByDate(ctx context.Context, db database.QueryExecer, date time.Time) (*entity.TimesheetConfirmationCutOffDate, error)
	}

	TimesheetLocationListRepo interface {
		GetTimesheetLocationList(ctx context.Context, db database.QueryExecer, req *dto.GetTimesheetLocationListReq) ([]*dto.TimesheetLocation, error)
		GetTimesheetLocationCount(ctx context.Context, db database.QueryExecer, keyword string) (*dto.TimesheetLocationAggregate, error)
		GetNonConfirmedLocationCount(ctx context.Context, db database.QueryExecer, periodDate time.Time) (*dto.GetNonConfirmedLocationCountOut, error)
	}

	TimesheetRepo interface {
		FindTimesheetInLocationByDateAndStatus(ctx context.Context, db database.QueryExecer, locationID, timesheetStatus string, startDate, endDate time.Time) ([]*entity.Timesheet, error)
		CountNotApprovedAndNotConfirmedTimesheet(ctx context.Context, db database.QueryExecer, locationID string, startDate, endDate time.Time) (int, error)
		UpdateTimesheetStatusToConfirmByDateAndLocation(ctx context.Context, db database.QueryExecer, startDate, endDate time.Time, timesheetStatus, locationID string) error
		FindTimesheetByTimesheetID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entity.Timesheet, error)
	}

	TimesheetConfirmationInfoRepo interface {
		InsertConfirmationInfo(ctx context.Context, db database.QueryExecer, period *entity.TimesheetConfirmationInfo) (*entity.TimesheetConfirmationInfo, error)
		GetConfirmationInfoByPeriodAndLocation(ctx context.Context, db database.QueryExecer, periodID, locationID pgtype.Text) (*entity.TimesheetConfirmationInfo, error)
	}
}

func (s *ConfirmationWindowServiceImpl) GetPeriod(ctx context.Context, date *timestamppb.Timestamp) (*dto.TimesheetConfirmationPeriod, error) {
	dateToQuery := convertDate(date)
	periodE, err := s.TimesheetConfirmationPeriodRepo.GetPeriodByDate(ctx, s.DB, dateToQuery)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// don't have period => we must calculate and save period to db
	if periodE == nil {
		var (
			startDateConvert time.Time
			endDateConvert   time.Time
		)
		cutOffDateE, err := s.TimesheetConfirmationCutOffDateRepo.GetCutOffDateByDate(ctx, s.DB, dateToQuery)
		if err != nil {
			return nil, err
		}
		cutOffDate := cutOffDateE.CutOffDate.Int
		startDate, endDate, startMonth, endMonth, startYear, endYear := calculatePeriodByCutOffDate(int(cutOffDate), dateToQuery)

		// small first period from 1/1/2022 - (cutOfDate+1)/1/2022
		if cutOffDateE.CutOffDate.Int > 1 && startMonth == 12 && startYear == 2021 {
			startDateConvert = time.Date(2022, 1, 1, 0, 0, 0, 0, dateToQuery.Location())
			endDateConvert = time.Date(2022, 1, int(cutOffDate+1), 23, 59, 59, 0, dateToQuery.Location())
		} else { // normal period
			startDateConvert = time.Date(startYear, time.Month(startMonth), startDate, 0, 0, 0, 0, dateToQuery.Location())
			endDateConvert = time.Date(endYear, time.Month(endMonth), endDate, 23, 59, 59, 0, dateToQuery.Location())
		}

		periodE = &entity.TimesheetConfirmationPeriod{
			ID:        database.Text(idutil.ULIDNow()),
			StartDate: database.Timestamptz(startDateConvert),
			EndDate:   database.Timestamptz(endDateConvert),
		}

		_, err = s.TimesheetConfirmationPeriodRepo.InsertPeriod(ctx, s.DB, periodE)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}

	periodDto := &dto.TimesheetConfirmationPeriod{
		ID:        periodE.ID.String,
		StartDate: periodE.StartDate.Time,
		EndDate:   periodE.EndDate.Time,
	}

	return periodDto, nil
}

func convertDate(date *timestamppb.Timestamp) time.Time {
	convertDate := date.AsTime().In(timeutil.Timezone(pbc.COUNTRY_JP))
	return time.Date(convertDate.Year(), convertDate.Month(), convertDate.Day(), convertDate.Hour(), convertDate.Minute(), 0, 0, convertDate.Location())
}

func calculatePeriodByCutOffDate(cutOffDate int, dateToQuery time.Time) (startDate, endDate, startMonth, endMonth, startYear, endYear int) {
	monthRequest := int(dateToQuery.Month())
	yearRequest := dateToQuery.Year()
	dateRequest := dateToQuery.Day()
	endDateOfMonth, _ := getEndDateOfMonth(monthRequest, yearRequest)

	startDate = 1
	endDate = endDateOfMonth
	startMonth = monthRequest
	endMonth = monthRequest
	startYear = yearRequest
	endYear = yearRequest

	if cutOffDate != 0 {
		if dateRequest > cutOffDate {
			startDate = cutOffDate + 1
			endDate = cutOffDate
			endMonth = monthRequest + 1
		} else {
			startDate = cutOffDate + 1
			endDate = cutOffDate
			startMonth = monthRequest - 1
		}

		if startMonth > 12 {
			startMonth = 1
			startYear = yearRequest + 1
			endYear = yearRequest + 1
		}

		if endMonth > 12 {
			endMonth = 1
			endYear = yearRequest + 1
		}

		if startMonth < 1 {
			startMonth = 12
			startYear = yearRequest - 1
		}
		endDateOfStartMonth, _ := getEndDateOfMonth(startMonth, startYear)
		if startDate > endDateOfStartMonth {
			startDate = 1
			startMonth++

			if startMonth > 12 {
				startMonth = 1
				startYear++
			}
		}

		endDateOfEndMonth, _ := getEndDateOfMonth(endMonth, endYear)
		if endDate > endDateOfEndMonth {
			endDate = endDateOfEndMonth
		}
	}
	return
}

func getEndDateOfMonth(month, year int) (int, error) {
	switch month {
	case 1, 3, 5, 7, 8, 10, 12:
		return 31, nil
	case 4, 6, 9, 11:
		return 30, nil
	case 2:
		if year%4 == 0 {
			return 29, nil
		}
		return 28, nil
	default:
		return -1, fmt.Errorf("error when get end date of month: month %d, year %d", month, year)
	}
}

func (s *ConfirmationWindowServiceImpl) ConfirmPeriod(ctx context.Context, request *pb.ConfirmTimesheetWithLocationRequest) error {
	confirmPeriod, err := s.TimesheetConfirmationPeriodRepo.GetPeriodByID(ctx, s.DB, database.Text(request.PeriodId))
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("find confirmation period error: %s", err.Error()))
	}

	if confirmPeriod.EndDate.Time.Add(-(time.Hour * 24)).After(time.Now()) {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("can not confirm period: %s before end date of this", confirmPeriod.ID.String))
	}

	for _, locationID := range request.LocationIds {
		// check location of this period is confirmed or not
		confirmationInfo, err := s.TimesheetConfirmationInfoRepo.GetConfirmationInfoByPeriodAndLocation(ctx, s.DB, confirmPeriod.ID, database.Text(locationID))
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return status.Error(codes.Internal, fmt.Sprintf("find confirmation info error: %s", err.Error()))
		}
		if confirmationInfo != nil {
			return status.Error(codes.AlreadyExists, fmt.Sprintf("timesheet in location: %s and period: %s is confirmed", locationID, request.PeriodId))
		}

		// check all timesheet in this period and location is approved or not
		notApprovedTimesheetNum, err := s.TimesheetRepo.CountNotApprovedAndNotConfirmedTimesheet(ctx, s.DB, locationID, confirmPeriod.StartDate.Time, confirmPeriod.EndDate.Time)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("count not approved and not confirmed timesheet error: %s", err.Error()))
		}
		if notApprovedTimesheetNum > 0 {
			return status.Error(codes.InvalidArgument, fmt.Sprintf("some timesheet in location: %s is not approved or confirmed, please approved all first", locationID))
		}
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		for _, locationID := range request.LocationIds {
			// set timesheets confirm status
			err = s.TimesheetRepo.UpdateTimesheetStatusToConfirmByDateAndLocation(ctx, tx, confirmPeriod.StartDate.Time, confirmPeriod.EndDate.Time, pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String(), locationID)
			if err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("upsert multiple timesheet status to confirm error: %s", err.Error()))
			}

			// save confirmation info first to lock all action with timesheets in this period
			confirmInfoToInsert := entity.TimesheetConfirmationInfo{
				ID:         database.Text(idutil.ULIDNow()),
				LocationID: database.Text(locationID),
				PeriodID:   database.Text(request.PeriodId),
			}
			_, err = s.TimesheetConfirmationInfoRepo.InsertConfirmationInfo(ctx, tx, &confirmInfoToInsert)

			if err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("insert confirmation info when confirm timesheet error: %s", err.Error()))
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	timeExecuted := time.Now()
	// send approve timesheet event to NATS
	for _, locationID := range request.LocationIds {
		// get list timesheet in this location and date
		timesheets, err := s.TimesheetRepo.FindTimesheetInLocationByDateAndStatus(ctx, s.DB, locationID, pb.TimesheetStatus_TIMESHEET_STATUS_CONFIRMED.String(), confirmPeriod.StartDate.Time, confirmPeriod.EndDate.Time)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("find timesheet in location and date error: %s", err.Error()))
		}
		for _, timesheet := range timesheets {
			msg := &pb.TimesheetActionLogRequest{
				Action:      pb.TimesheetAction_CONFIRMED,
				ExecutedBy:  interceptors.UserIDFromContext(ctx),
				TimesheetId: timesheet.TimesheetID.String,
				IsSystem:    false,
				ExecutedAt:  timestamppb.New(timeExecuted),
			}
			err = nats_service_utils.PublishActionLogTimesheetEvent(ctx, msg, s.JSM)
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}
		}
	}

	return nil
}

func (s *ConfirmationWindowServiceImpl) GetTimesheetLocationList(ctx context.Context, request *dto.GetTimesheetLocationListReq) (*dto.GetTimesheetLocationListOut, error) {
	result := &dto.GetTimesheetLocationListOut{}
	locationList, err := s.TimesheetLocationListRepo.GetTimesheetLocationList(ctx, s.DB, request)
	if err != nil {
		return nil, err
	}
	locationAgg, err2 := s.TimesheetLocationListRepo.GetTimesheetLocationCount(ctx, s.DB, request.Keyword)

	if err2 != nil {
		return nil, err2
	}
	result.Locations = locationList
	result.LocationAggregate = locationAgg

	return result, nil
}

func (s *ConfirmationWindowServiceImpl) GetNonConfirmedLocationCount(ctx context.Context, request *dto.GetNonConfirmedLocationCountReq) (*dto.GetNonConfirmedLocationCountOut, error) {
	return s.TimesheetLocationListRepo.GetNonConfirmedLocationCount(ctx, s.DB, request.PeriodDate)
}

func (s *ConfirmationWindowServiceImpl) CheckModifyConditionByTimesheetDateAndLocation(ctx context.Context, timesheetDate *timestamppb.Timestamp, locationID string) (bool, error) {
	dateToQuery := convertDate(timesheetDate)
	periodE, err := s.TimesheetConfirmationPeriodRepo.GetPeriodByDate(ctx, s.DB, dateToQuery)
	if errors.Is(err, pgx.ErrNoRows) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("find confirmation period error: %s", err)
	}

	_, err = s.TimesheetConfirmationInfoRepo.GetConfirmationInfoByPeriodAndLocation(ctx, s.DB, periodE.ID, database.Text(locationID))
	if errors.Is(err, pgx.ErrNoRows) {
		return true, nil
	}

	if err != nil {
		return false, fmt.Errorf("find confirmation info error: %s", err)
	}

	return false, nil
}

func (s *ConfirmationWindowServiceImpl) CheckModifyConditionByTimesheetID(ctx context.Context, timesheetID string) (bool, error) {
	timesheet, err := s.TimesheetRepo.FindTimesheetByTimesheetID(ctx, s.DB, database.Text(timesheetID))
	if err != nil {
		return false, fmt.Errorf("find timesheet error: %s", err)
	}

	periodE, err := s.TimesheetConfirmationPeriodRepo.GetPeriodByDate(ctx, s.DB, timesheet.TimesheetDate.Time)
	if errors.Is(err, pgx.ErrNoRows) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("find confirmation period error: %s", err)
	}

	_, err = s.TimesheetConfirmationInfoRepo.GetConfirmationInfoByPeriodAndLocation(ctx, s.DB, periodE.ID, database.Text(timesheet.LocationID.String))
	if errors.Is(err, pgx.ErrNoRows) {
		return true, nil
	}

	if err != nil {
		return false, fmt.Errorf("find confirmation info error: %s", err)
	}

	return false, nil
}
