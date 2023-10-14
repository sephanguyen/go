package timesheet

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_repositories "github.com/manabie-com/backend/mock/timesheet/repository"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TestCaseCalculateCutOffDate struct {
	name             string
	cutOffDate       int
	dateToQuery      time.Time
	startDateExpect  int
	startMonthExpect int
	endDateExpect    int
	endMonthExpect   int
	startYearExpect  int
	endYearExpect    int
}

func TestTimesheetConfirmationService_GetPeriod(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var (
		tsPeriodRepo     = new(mock_repositories.MockTimesheetConfirmationPeriodRepoImpl)
		tsCutOffDateRepo = new(mock_repositories.MockTimesheetConfirmationCutOffDateRepoImpl)
		db               = new(mock_database.Ext)
		tx               = new(mock_database.Tx)
	)

	timeNow := time.Now()

	TsPeriodE := &entity.TimesheetConfirmationPeriod{
		ID:        database.Text("period-id"),
		StartDate: database.Timestamptz(timeNow),
		EndDate:   database.Timestamptz(timeNow),
	}

	TsPeriodDto := &dto.TimesheetConfirmationPeriod{
		ID:        "period-id",
		StartDate: timeNow,
		EndDate:   timeNow,
	}

	s := ConfirmationWindowServiceImpl{
		DB:                                  db,
		TimesheetConfirmationPeriodRepo:     tsPeriodRepo,
		TimesheetConfirmationCutOffDateRepo: tsCutOffDateRepo,
	}

	testCases := []TestCase{

		{
			name:         "happy case",
			ctx:          ctx,
			expectedErr:  nil,
			expectedResp: (*dto.TimesheetConfirmationPeriod)(TsPeriodDto),
			setup: func(ctx context.Context) {
				tsPeriodRepo.On("GetPeriodByDate", ctx, db, mock.Anything, mock.Anything).
					Return(TsPeriodE, nil).Once()
			},
		},
		{
			name:         "error case get cut off date failed",
			ctx:          ctx,
			expectedErr:  status.Error(codes.Internal, "err get cut off date"),
			expectedResp: (*dto.TimesheetConfirmationPeriod)(nil),
			setup: func(ctx context.Context) {
				tsPeriodRepo.On("GetPeriodByDate", ctx, db, mock.Anything, mock.Anything).
					Return(nil, status.Error(codes.Internal, fmt.Sprintf("err get cut off date"))).Once()
			},
		},
		{
			name:         "error case get cut off date by date failed",
			ctx:          ctx,
			expectedErr:  status.Error(codes.Internal, "err get cut off date by date"),
			expectedResp: (*dto.TimesheetConfirmationPeriod)(nil),
			setup: func(ctx context.Context) {
				tsPeriodRepo.On("GetPeriodByDate", ctx, db, mock.Anything, mock.Anything).
					Return(nil, nil).Once()
				tsCutOffDateRepo.On("GetCutOffDateByDate", ctx, db, mock.Anything).
					Return(nil, status.Error(codes.Internal, fmt.Sprintf("err get cut off date by date"))).Once()
			},
		},
	}

	// Do test
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.GetPeriod(testCase.ctx, timestamppb.Now())
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
			mock.AssertExpectationsForObjects(
				t,
				tsPeriodRepo,
				db,
				tx,
			)
		})
	}
}

func TestTimesheetConfirmationService_calculatePeriodByCutOffDate(t *testing.T) {
	t.Parallel()

	var (
		January     = 1
		February    = 2
		March       = 3
		April       = 4
		May         = 5
		December    = 12
		leapYear    = 2020
		notLeapYear = 2021
	)

	testCases := []TestCaseCalculateCutOffDate{
		{
			name:             "case cut off date is 0 and month request is not February and not leap year",
			cutOffDate:       0,
			dateToQuery:      time.Date(notLeapYear, time.Month(January), 5, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  1,
			endDateExpect:    31,
			startMonthExpect: January,
			endMonthExpect:   January,
			startYearExpect:  notLeapYear,
			endYearExpect:    notLeapYear,
		},
		{
			name:             "case cut off date is 0 and month request is February and not leap year",
			cutOffDate:       0,
			dateToQuery:      time.Date(notLeapYear, time.Month(February), 5, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  1,
			endDateExpect:    28,
			startMonthExpect: February,
			endMonthExpect:   February,
			startYearExpect:  notLeapYear,
			endYearExpect:    notLeapYear,
		},
		{
			name:             "case cut off date is 0 and month request is February and leap year",
			cutOffDate:       0,
			dateToQuery:      time.Date(leapYear, time.Month(February), 5, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  1,
			endDateExpect:    29,
			startMonthExpect: February,
			endMonthExpect:   February,
			startYearExpect:  leapYear,
			endYearExpect:    leapYear,
		},
		{
			name:             "case cut off date is <= 27 and month is not Dec and date request <= cut of date",
			cutOffDate:       15,
			dateToQuery:      time.Date(notLeapYear, time.Month(February), 15, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  16,
			endDateExpect:    15,
			startMonthExpect: January,
			endMonthExpect:   February,
			startYearExpect:  notLeapYear,
			endYearExpect:    notLeapYear,
		},
		{
			name:             "case cut off date is <= 27 and month request is not Dec and date request > cut of date",
			cutOffDate:       15,
			dateToQuery:      time.Date(notLeapYear, time.Month(January), 16, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  16,
			endDateExpect:    15,
			startMonthExpect: January,
			endMonthExpect:   February,
			startYearExpect:  notLeapYear,
			endYearExpect:    notLeapYear,
		},
		{
			name:             "case cut off date is <= 27 and month is not Dec and date request <= cut of date",
			cutOffDate:       15,
			dateToQuery:      time.Date(notLeapYear, time.Month(February), 15, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  16,
			endDateExpect:    15,
			startMonthExpect: January,
			endMonthExpect:   February,
			startYearExpect:  notLeapYear,
			endYearExpect:    notLeapYear,
		},
		{
			name:             "case cut off date is <= 27 and month request is Dec and date request > cut of date",
			cutOffDate:       15,
			dateToQuery:      time.Date(notLeapYear, time.Month(December), 16, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  16,
			endDateExpect:    15,
			startMonthExpect: December,
			endMonthExpect:   January,
			startYearExpect:  notLeapYear,
			endYearExpect:    notLeapYear + 1,
		},
		{
			name:             "case cut off date is <= 27 and month request is Jan and date request > cut of date",
			cutOffDate:       15,
			dateToQuery:      time.Date(notLeapYear, time.Month(January), 16, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  16,
			endDateExpect:    15,
			startMonthExpect: January,
			endMonthExpect:   February,
			startYearExpect:  notLeapYear,
			endYearExpect:    notLeapYear,
		},
		{
			name:             "case cut off date is <= 27 and month request is Jan and date request =< cut of date",
			cutOffDate:       15,
			dateToQuery:      time.Date(notLeapYear, time.Month(January), 15, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  16,
			endDateExpect:    15,
			startMonthExpect: December,
			endMonthExpect:   January,
			startYearExpect:  notLeapYear - 1,
			endYearExpect:    notLeapYear,
		},
		{
			name:             "case cut off date is = 28 and month request Feb and date request > cut of date and year request is leap year",
			cutOffDate:       28,
			dateToQuery:      time.Date(leapYear, time.Month(February), 29, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  29,
			endDateExpect:    28,
			startMonthExpect: February,
			endMonthExpect:   March,
			startYearExpect:  leapYear,
			endYearExpect:    leapYear,
		},
		{
			name:             "case cut off date is = 28 and month request Feb and date request > cut of date and year request is not leap year",
			cutOffDate:       28,
			dateToQuery:      time.Date(notLeapYear, time.Month(March), 15, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  1,
			endDateExpect:    28,
			startMonthExpect: March,
			endMonthExpect:   March,
			startYearExpect:  notLeapYear,
			endYearExpect:    notLeapYear,
		},
		{
			name:             "case cut off date is = 28 and month request Feb and date request <= cut of date",
			cutOffDate:       28,
			dateToQuery:      time.Date(leapYear, time.Month(February), 15, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  29,
			endDateExpect:    28,
			startMonthExpect: January,
			endMonthExpect:   February,
			startYearExpect:  leapYear,
			endYearExpect:    leapYear,
		},
		{
			name:             "case cut off date is = 29 and month request Feb and date request <= cut of date and year request is leap year",
			cutOffDate:       29,
			dateToQuery:      time.Date(leapYear, time.Month(February), 29, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  30,
			startMonthExpect: January,
			endDateExpect:    29,
			endMonthExpect:   February,
			startYearExpect:  leapYear,
			endYearExpect:    leapYear,
		},
		{
			name:             "case cut off date is = 29 and month request March and year request is leap year",
			cutOffDate:       29,
			dateToQuery:      time.Date(leapYear, time.Month(March), 15, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  1,
			startMonthExpect: March,
			endDateExpect:    29,
			endMonthExpect:   March,
			startYearExpect:  leapYear,
			endYearExpect:    leapYear,
		},
		{
			name:             "case cut off date is = 30 and month request May (30days) and date request <= cut of date ",
			cutOffDate:       30,
			dateToQuery:      time.Date(leapYear, time.Month(May), 15, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  1,
			startMonthExpect: May,
			endDateExpect:    30,
			endMonthExpect:   May,
			startYearExpect:  leapYear,
			endYearExpect:    leapYear,
		},
		{
			name:             "case cut off date is = 30 and month request March (30days) and date request <= cut of date ",
			cutOffDate:       30,
			dateToQuery:      time.Date(leapYear, time.Month(March), 15, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  1,
			startMonthExpect: March,
			endDateExpect:    30,
			endMonthExpect:   March,
			startYearExpect:  leapYear,
			endYearExpect:    leapYear,
		},
		{
			name:             "case cut off date is = 30 and month request February and date request <= cut of date and year is leap year",
			cutOffDate:       30,
			dateToQuery:      time.Date(leapYear, time.Month(February), 15, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  31,
			startMonthExpect: January,
			endDateExpect:    29,
			endMonthExpect:   February,
			startYearExpect:  leapYear,
			endYearExpect:    leapYear,
		},
		{
			name:             "case cut off date is = 30 and month request February and date request <= cut of date and year is not leap year",
			cutOffDate:       30,
			dateToQuery:      time.Date(notLeapYear, time.Month(February), 15, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  31,
			startMonthExpect: January,
			endDateExpect:    28,
			endMonthExpect:   February,
			startYearExpect:  notLeapYear,
			endYearExpect:    notLeapYear,
		},
		{
			name:             "case cut off date is = 31 and month request February and date request <= cut of date and year is not leap year",
			cutOffDate:       31,
			dateToQuery:      time.Date(notLeapYear, time.Month(February), 15, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  1,
			startMonthExpect: February,
			endDateExpect:    28,
			endMonthExpect:   February,
			startYearExpect:  notLeapYear,
			endYearExpect:    notLeapYear,
		},
		{
			name:             "case cut off date is = 31 and month request February and date request <= cut of date and year is leap year",
			cutOffDate:       31,
			dateToQuery:      time.Date(leapYear, time.Month(February), 15, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  1,
			startMonthExpect: February,
			endDateExpect:    29,
			endMonthExpect:   February,
			startYearExpect:  leapYear,
			endYearExpect:    leapYear,
		},
		{
			name:             "case cut off date is = 31 and month request April(30 days) and date request <= cut of date",
			cutOffDate:       31,
			dateToQuery:      time.Date(leapYear, time.Month(April), 15, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  1,
			startMonthExpect: April,
			endDateExpect:    30,
			endMonthExpect:   April,
			startYearExpect:  leapYear,
			endYearExpect:    leapYear,
		},
		{
			name:             "case cut off date is = 31 and month request May(31 days) and date request <= cut of date",
			cutOffDate:       31,
			dateToQuery:      time.Date(leapYear, time.Month(May), 15, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  1,
			startMonthExpect: May,
			endDateExpect:    31,
			endMonthExpect:   May,
			startYearExpect:  leapYear,
			endYearExpect:    leapYear,
		},
		{
			name:             "case cut off date is = 30 and month request Jan and date request <= cut of date",
			cutOffDate:       30,
			dateToQuery:      time.Date(leapYear, time.Month(January), 15, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  31,
			startMonthExpect: December,
			endDateExpect:    30,
			endMonthExpect:   January,
			startYearExpect:  leapYear - 1,
			endYearExpect:    leapYear,
		},
		{
			name:             "case cut off date is = 31 and month request Jan and date request <= cut of date",
			cutOffDate:       31,
			dateToQuery:      time.Date(leapYear, time.Month(January), 15, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  1,
			startMonthExpect: January,
			endDateExpect:    31,
			endMonthExpect:   January,
			startYearExpect:  leapYear,
			endYearExpect:    leapYear,
		},
		{
			name:             "case cut off date is = 1 and month request Dec and date request > cut of date",
			cutOffDate:       1,
			dateToQuery:      time.Date(leapYear, time.Month(December), 31, 0, 0, 0, 0, time.Now().Location()),
			startDateExpect:  2,
			startMonthExpect: December,
			endDateExpect:    1,
			endMonthExpect:   January,
			startYearExpect:  leapYear,
			endYearExpect:    leapYear + 1,
		},
	}

	// Do test
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			startDate, endDate, startMonth, endMonth, startYear, endYear := calculatePeriodByCutOffDate(testCase.cutOffDate, testCase.dateToQuery)
			assert.Equal(t, testCase.startDateExpect, startDate)
			assert.Equal(t, testCase.endDateExpect, endDate)
			assert.Equal(t, testCase.startMonthExpect, startMonth)
			assert.Equal(t, testCase.endMonthExpect, endMonth)
			assert.Equal(t, testCase.startYearExpect, startYear)
			assert.Equal(t, testCase.endYearExpect, endYear)
		})
	}
}

func TestTimesheetConfirmationService_ConfirmPeriod(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var (
		tsPeriodRepo      = new(mock_repositories.MockTimesheetConfirmationPeriodRepoImpl)
		tsCutOffDateRepo  = new(mock_repositories.MockTimesheetConfirmationCutOffDateRepoImpl)
		timesheetRepo     = new(mock_repositories.MockTimesheetRepoImpl)
		tsConfirmInfoRepo = new(mock_repositories.MockTimesheetConfirmationInfoRepoImpl)
		db                = new(mock_database.Ext)
		tx                = new(mock_database.Tx)
		mockJsm           = new(mock_nats.JetStreamManagement)
	)

	timeNow := time.Now()

	TsPeriodDto := &dto.TimesheetConfirmationPeriod{
		ID:        "period-id",
		StartDate: timeNow,
		EndDate:   timeNow,
	}

	TsPeriodE := &entity.TimesheetConfirmationPeriod{
		ID:        database.Text("period-id"),
		StartDate: database.Timestamptz(timeNow),
		EndDate:   database.Timestamptz(time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)),
	}

	TsPeriodEndDateAfterNowE := &entity.TimesheetConfirmationPeriod{
		ID:        database.Text("TsPeriodEndDateAfterNowE"),
		StartDate: database.Timestamptz(timeNow),
		EndDate:   database.Timestamptz(timeNow.Add(time.Hour * 48)),
		// EndDate:   database.Timestamptz(time.Now().Add(time.Hour)),
	}

	TsConfirmInfoE := &entity.TimesheetConfirmationInfo{}

	confirmTsRequest := &pb.ConfirmTimesheetWithLocationRequest{}
	confirmTsRequestWithPeriodAndLocation := &pb.ConfirmTimesheetWithLocationRequest{
		PeriodId:    "1",
		LocationIds: []string{"1"},
	}

	s := ConfirmationWindowServiceImpl{
		DB:                                  db,
		TimesheetConfirmationPeriodRepo:     tsPeriodRepo,
		TimesheetConfirmationCutOffDateRepo: tsCutOffDateRepo,
		TimesheetRepo:                       timesheetRepo,
		TimesheetConfirmationInfoRepo:       tsConfirmInfoRepo,
		JSM:                                 mockJsm,
	}

	var timesheetsE []*entity.Timesheet
	var timesheetE = entity.Timesheet{
		TimesheetID: database.Text("ts-1"),
	}
	timesheetsE = append(timesheetsE, &timesheetE)
	testCases := []TestCase{
		{
			name:         "error case publish event timesheet action log fail",
			ctx:          ctx,
			req:          confirmTsRequestWithPeriodAndLocation,
			expectedErr:  status.Error(codes.Internal, "PublishActionLogTimesheetEvent JSM.PublishAsyncContext failed, msgID: MsgID, Error"),
			expectedResp: (*dto.TimesheetConfirmationPeriod)(TsPeriodDto),
			setup: func(ctx context.Context) {
				tsPeriodRepo.On("GetPeriodByID", ctx, db, mock.Anything).
					Return(TsPeriodE, nil).Once()
				tsConfirmInfoRepo.On("GetConfirmationInfoByPeriodAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(nil, nil).Once()
				timesheetRepo.On("CountNotApprovedAndNotConfirmedTimesheet", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(0, nil).Once()

				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetRepo.On("UpdateTimesheetStatusToConfirmByDateAndLocation", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil).Once()
				tsConfirmInfoRepo.On("InsertConfirmationInfo", ctx, tx, mock.Anything).
					Return(nil, nil).Once()
				tx.On("Commit", mock.Anything).Return(nil)

				timesheetRepo.On("FindTimesheetInLocationByDateAndStatus", ctx, db, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(timesheetsE, nil).Once()
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("MsgID", fmt.Errorf("Error"))
			},
		},
		{
			name:         "happy case",
			ctx:          ctx,
			req:          confirmTsRequest,
			expectedErr:  nil,
			expectedResp: (*dto.TimesheetConfirmationPeriod)(TsPeriodDto),
			setup: func(ctx context.Context) {
				tsPeriodRepo.On("GetPeriodByID", ctx, db, mock.Anything).
					Return(TsPeriodE, nil).Once()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tsConfirmInfoRepo.On("GetConfirmInfoByPeriodAndLocation", ctx, tx, mock.Anything, mock.Anything).
					Return(nil, nil).Once()
				timesheetRepo.On("FindTimesheetInLocationByDateAndStatus", ctx, tx, mock.Anything, mock.Anything).
					Return(timesheetsE, nil).Once()
				timesheetRepo.On("FindTimesheetInLocationByDate", ctx, tx, mock.Anything, mock.Anything).
					Return(timesheetsE, nil).Once()
				timesheetRepo.On("UpdateTimesheetStatusMultiple", ctx, tx, mock.Anything, mock.Anything).
					Return(nil).Once()
				tsConfirmInfoRepo.On("InsertConfirmInfo", ctx, tx, mock.Anything, mock.Anything).
					Return(TsConfirmInfoE, nil).Once()
				tx.On("Commit", mock.Anything).Return(nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:         "error case find confirmation period error",
			ctx:          ctx,
			req:          confirmTsRequest,
			expectedErr:  status.Error(codes.Internal, "find confirmation period error: err get confirmation period"),
			expectedResp: (*dto.TimesheetConfirmationPeriod)(TsPeriodDto),
			setup: func(ctx context.Context) {
				tsPeriodRepo.On("GetPeriodByID", ctx, db, mock.Anything).
					Return(TsPeriodE, errors.New("err get confirmation period")).Once()
			},
		},
		{
			name:         "error case confirm before end of period date",
			ctx:          ctx,
			req:          confirmTsRequest,
			expectedErr:  status.Error(codes.InvalidArgument, "can not confirm period: TsPeriodEndDateAfterNowE before end date of this"),
			expectedResp: (*dto.TimesheetConfirmationPeriod)(nil),
			setup: func(ctx context.Context) {
				tsPeriodRepo.On("GetPeriodByID", ctx, db, mock.Anything, mock.Anything).
					Return(TsPeriodEndDateAfterNowE, nil).Once()
			},
		},
		{
			name:         "error case failed to retrieve confirmation info due to database error",
			ctx:          ctx,
			req:          confirmTsRequestWithPeriodAndLocation,
			expectedErr:  status.Error(codes.Internal, "find confirmation info error: internal error"),
			expectedResp: (*dto.TimesheetConfirmationPeriod)(nil),
			setup: func(ctx context.Context) {
				tsPeriodRepo.On("GetPeriodByID", ctx, db, mock.Anything).
					Return(TsPeriodE, nil).Once()
				tsConfirmInfoRepo.On("GetConfirmationInfoByPeriodAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(nil, errors.New("internal error")).Once()
			},
		},
		{
			name:         "error case timesheet in location and period already confirmed",
			ctx:          ctx,
			req:          confirmTsRequestWithPeriodAndLocation,
			expectedErr:  status.Error(codes.AlreadyExists, "timesheet in location: 1 and period: 1 is confirmed"),
			expectedResp: (*dto.TimesheetConfirmationPeriod)(nil),
			setup: func(ctx context.Context) {
				tsPeriodRepo.On("GetPeriodByID", ctx, db, mock.Anything).
					Return(TsPeriodE, nil).Once()
				tsConfirmInfoRepo.On("GetConfirmationInfoByPeriodAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(TsConfirmInfoE, nil).Once()
			},
		},
		{
			name:         "error case count not approved timesheet error",
			ctx:          ctx,
			req:          confirmTsRequestWithPeriodAndLocation,
			expectedErr:  status.Error(codes.Internal, "count not approved and not confirmed timesheet error: internal error"),
			expectedResp: (*dto.TimesheetConfirmationPeriod)(nil),
			setup: func(ctx context.Context) {
				tsPeriodRepo.On("GetPeriodByID", ctx, db, mock.Anything).
					Return(TsPeriodE, nil).Once()
				tsConfirmInfoRepo.On("GetConfirmationInfoByPeriodAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(nil, nil).Once()
				timesheetRepo.On("CountNotApprovedAndNotConfirmedTimesheet", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(0, errors.New("internal error")).Once()
			},
		},
		{
			name:         "error case timesheets in location not approved",
			ctx:          ctx,
			req:          confirmTsRequestWithPeriodAndLocation,
			expectedErr:  status.Error(codes.InvalidArgument, "some timesheet in location: 1 is not approved or confirmed, please approved all first"),
			expectedResp: (*dto.TimesheetConfirmationPeriod)(nil),
			setup: func(ctx context.Context) {
				tsPeriodRepo.On("GetPeriodByID", ctx, db, mock.Anything).
					Return(TsPeriodE, nil).Once()
				tsConfirmInfoRepo.On("GetConfirmationInfoByPeriodAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(nil, nil).Once()
				timesheetRepo.On("CountNotApprovedAndNotConfirmedTimesheet", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(1, nil).Once()
			},
		},
		{
			name:         "error case failed to update timesheet status to confirm by date and location",
			ctx:          ctx,
			req:          confirmTsRequestWithPeriodAndLocation,
			expectedErr:  status.Error(codes.Internal, "upsert multiple timesheet status to confirm error: internal error"),
			expectedResp: (*dto.TimesheetConfirmationPeriod)(nil),
			setup: func(ctx context.Context) {
				tsPeriodRepo.On("GetPeriodByID", ctx, db, mock.Anything).
					Return(TsPeriodE, nil).Once()
				tsConfirmInfoRepo.On("GetConfirmationInfoByPeriodAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(nil, nil).Once()
				timesheetRepo.On("CountNotApprovedAndNotConfirmedTimesheet", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(0, nil).Once()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetRepo.On("UpdateTimesheetStatusToConfirmByDateAndLocation", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("internal error")).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:         "error case failed to insert confirmation info",
			ctx:          ctx,
			req:          confirmTsRequestWithPeriodAndLocation,
			expectedErr:  status.Error(codes.Internal, "insert confirmation info when confirm timesheet error: internal error"),
			expectedResp: (*dto.TimesheetConfirmationPeriod)(nil),
			setup: func(ctx context.Context) {
				tsPeriodRepo.On("GetPeriodByID", ctx, db, mock.Anything).
					Return(TsPeriodE, nil).Once()
				tsConfirmInfoRepo.On("GetConfirmationInfoByPeriodAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(nil, nil).Once()
				timesheetRepo.On("CountNotApprovedAndNotConfirmedTimesheet", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(0, nil).Once()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetRepo.On("UpdateTimesheetStatusToConfirmByDateAndLocation", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil).Once()
				tsConfirmInfoRepo.On("InsertConfirmationInfo", ctx, tx, mock.Anything).
					Return(nil, errors.New("internal error")).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		testCase.setup(testCase.ctx)
		req := testCase.req.(*pb.ConfirmTimesheetWithLocationRequest)
		err := s.ConfirmPeriod(testCase.ctx, req)
		assert.Equal(t, testCase.expectedErr, err)
		mock.AssertExpectationsForObjects(
			t,
			tsPeriodRepo,
			db,
			tx,
		)
	}
}

func TestTimesheetConfirmationService_CheckConfirmInfoByDateAndLocation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var (
		tsPeriodRepo      = new(mock_repositories.MockTimesheetConfirmationPeriodRepoImpl)
		tsCutOffDateRepo  = new(mock_repositories.MockTimesheetConfirmationCutOffDateRepoImpl)
		tsConfirmInfoRepo = new(mock_repositories.MockTimesheetConfirmationInfoRepoImpl)
		db                = new(mock_database.Ext)
		tx                = new(mock_database.Tx)
	)

	timeNow := time.Now()

	TsPeriodE := &entity.TimesheetConfirmationPeriod{
		ID:        database.Text("period-id"),
		StartDate: database.Timestamptz(timeNow),
		EndDate:   database.Timestamptz(timeNow),
	}

	TsConfirmInfoE := &entity.TimesheetConfirmationInfo{}

	s := ConfirmationWindowServiceImpl{
		DB:                                  db,
		TimesheetConfirmationPeriodRepo:     tsPeriodRepo,
		TimesheetConfirmationCutOffDateRepo: tsCutOffDateRepo,
		TimesheetConfirmationInfoRepo:       tsConfirmInfoRepo,
	}

	testCases := []TestCase{
		{
			name:         "happy case not have any period belong to this timesheet date",
			ctx:          ctx,
			expectedErr:  nil,
			expectedResp: true,
			setup: func(ctx context.Context) {
				tsPeriodRepo.On("GetPeriodByDate", ctx, db, mock.Anything, mock.Anything).
					Return(TsPeriodE, pgx.ErrNoRows).Once()
			},
		},
		{
			name:         "happy case not have any confirmation info belong to this timesheet date",
			ctx:          ctx,
			expectedErr:  nil,
			expectedResp: true,
			setup: func(ctx context.Context) {
				tsPeriodRepo.On("GetPeriodByDate", ctx, db, mock.Anything, mock.Anything).
					Return(TsPeriodE, nil).Once()
				tsConfirmInfoRepo.On("GetConfirmationInfoByPeriodAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(TsConfirmInfoE, pgx.ErrNoRows).Once()
			},
		},
		{
			name:         "happy case have confirmation info belong to this timesheet date",
			ctx:          ctx,
			expectedErr:  nil,
			expectedResp: false,
			setup: func(ctx context.Context) {
				tsPeriodRepo.On("GetPeriodByDate", ctx, db, mock.Anything, mock.Anything).
					Return(TsPeriodE, nil).Once()
				tsConfirmInfoRepo.On("GetConfirmationInfoByPeriodAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(TsConfirmInfoE, nil).Once()
			},
		},
		{
			name:         "error get confirmation period",
			ctx:          ctx,
			expectedErr:  fmt.Errorf("find confirmation period error: "),
			expectedResp: false,
			setup: func(ctx context.Context) {
				tsPeriodRepo.On("GetPeriodByDate", ctx, db, mock.Anything, mock.Anything).
					Return(TsPeriodE, fmt.Errorf("")).Once()
				tsConfirmInfoRepo.On("GetConfirmationInfoByPeriodAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(TsConfirmInfoE, nil).Once()
			},
		},
	}

	// Do test
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.CheckModifyConditionByTimesheetDateAndLocation(testCase.ctx, timestamppb.Now(), "location_id")
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
			mock.AssertExpectationsForObjects(
				t,
				tsPeriodRepo,
				db,
				tx,
			)
		})
	}
}

func TestTimesheetConfirmationService_CheckModifyConditionByTimesheetID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var (
		tsPeriodRepo      = new(mock_repositories.MockTimesheetConfirmationPeriodRepoImpl)
		tsCutOffDateRepo  = new(mock_repositories.MockTimesheetConfirmationCutOffDateRepoImpl)
		tsConfirmInfoRepo = new(mock_repositories.MockTimesheetConfirmationInfoRepoImpl)
		db                = new(mock_database.Ext)
		tx                = new(mock_database.Tx)
		timesheetRepo     = new(mock_repositories.MockTimesheetRepoImpl)
	)

	timeNow := time.Now()

	TsPeriodE := &entity.TimesheetConfirmationPeriod{
		ID:        database.Text("period-id"),
		StartDate: database.Timestamptz(timeNow),
		EndDate:   database.Timestamptz(timeNow),
	}

	TsConfirmInfoE := &entity.TimesheetConfirmationInfo{}
	TimesheetE := &entity.Timesheet{}

	s := ConfirmationWindowServiceImpl{
		DB:                                  db,
		TimesheetConfirmationPeriodRepo:     tsPeriodRepo,
		TimesheetConfirmationCutOffDateRepo: tsCutOffDateRepo,
		TimesheetConfirmationInfoRepo:       tsConfirmInfoRepo,
		TimesheetRepo:                       timesheetRepo,
	}

	testCases := []TestCase{
		{
			name:         "happy case not have any period belong to this timesheet date",
			ctx:          ctx,
			expectedErr:  nil,
			expectedResp: true,
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetID", ctx, db, mock.Anything).
					Return(TimesheetE, nil).Once()
				tsPeriodRepo.On("GetPeriodByDate", ctx, db, mock.Anything, mock.Anything).
					Return(TsPeriodE, pgx.ErrNoRows).Once()
			},
		},
		{
			name:         "happy case not have any confirmation info belong to this timesheet date",
			ctx:          ctx,
			expectedErr:  nil,
			expectedResp: true,
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetID", ctx, db, mock.Anything).
					Return(TimesheetE, nil).Once()
				tsPeriodRepo.On("GetPeriodByDate", ctx, db, mock.Anything, mock.Anything).
					Return(TsPeriodE, nil).Once()
				tsConfirmInfoRepo.On("GetConfirmationInfoByPeriodAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(TsConfirmInfoE, pgx.ErrNoRows).Once()
			},
		},
		{
			name:         "happy case have confirmation info belong to this timesheet date",
			ctx:          ctx,
			expectedErr:  nil,
			expectedResp: false,
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetID", ctx, db, mock.Anything).
					Return(TimesheetE, nil).Once()
				tsPeriodRepo.On("GetPeriodByDate", ctx, db, mock.Anything, mock.Anything).
					Return(TsPeriodE, nil).Once()
				tsConfirmInfoRepo.On("GetConfirmationInfoByPeriodAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(TsConfirmInfoE, nil).Once()
			},
		},
		{
			name:         "error get confirmation period",
			ctx:          ctx,
			expectedErr:  fmt.Errorf("find confirmation period error: "),
			expectedResp: false,
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetID", ctx, db, mock.Anything).
					Return(TimesheetE, nil).Once()
				tsPeriodRepo.On("GetPeriodByDate", ctx, db, mock.Anything, mock.Anything).
					Return(TsPeriodE, fmt.Errorf("")).Once()
				tsConfirmInfoRepo.On("GetConfirmationInfoByPeriodAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(TsConfirmInfoE, nil).Once()
			},
		},
	}

	// Do test
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.CheckModifyConditionByTimesheetID(testCase.ctx, "timesheet_id")
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
			mock.AssertExpectationsForObjects(
				t,
				tsPeriodRepo,
				db,
				tx,
			)
		})
	}
}
