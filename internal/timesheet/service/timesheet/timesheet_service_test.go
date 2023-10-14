package timesheet

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/timesheet/domain/constant"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_repositories "github.com/manabie-com/backend/mock/timesheet/repository"
	mock_get_timesheet "github.com/manabie-com/backend/mock/timesheet/service/gettimesheet"
	tpb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	CreateTimesheetStaffID       = "user_id"
	CreateTimeSheetRemark        = "test remark"
	CreateTimesheetUserIDSuccess = CreateTimesheetStaffID
	CreateTimesheetUserIDFail    = "failed_user_id"
	CreateTimesheetLocationID    = "location_id"
)

var (
	CreateTimesheetID            = idutil.ULIDNow()
	UpdateTimesheetID            = idutil.ULIDNow()
	CreateTimesheetConfigID      = idutil.ULIDNow()
	UpdateTimesheetConfigID      = idutil.ULIDNow()
	LessonID                     = idutil.ULIDNow()
	CreateTimesheetDateNow       = time.Now()
	CreateTimesheetTimesheetDate = time.Now()
	UpdateOWHsID                 = idutil.ULIDNow()
	now                          = time.Now()
)

type TestCase struct {
	name            string
	ctx             context.Context
	req             interface{}
	expectedResp    interface{}
	expectedErr     error
	setup           func(ctx context.Context)
	reqString       string
	reqTimesheetIDs []string
	staffID         string
	locationName    string
	limit           int32
}

func TestTimesheetService_CreateTimesheet(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var (
		timesheetRepo            = new(mock_repositories.MockTimesheetRepoImpl)
		owhsRepo                 = new(mock_repositories.MockOtherWorkingHoursRepoImpl)
		transportExpenseRepo     = new(mock_repositories.MockTransportationExpenseRepoImpl)
		timesheetLessonHoursRepo = new(mock_repositories.MockTimesheetLessonHoursRepoImpl)
		mockGetTimesheetService  = new(mock_get_timesheet.MockGetTimesheetServiceImpl)
		db                       = new(mock_database.Ext)
		tx                       = new(mock_database.Tx)
	)

	s := ServiceImpl{
		DB:                        db,
		TimesheetRepo:             timesheetRepo,
		OtherWorkingHoursRepo:     owhsRepo,
		TransportationExpenseRepo: transportExpenseRepo,
		TimesheetLessonHoursRepo:  timesheetLessonHoursRepo,
		GetTimesheetService:       mockGetTimesheetService,
	}

	timesheetE := &entity.Timesheet{
		TimesheetID:     database.Text(CreateTimesheetID),
		CreatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		UpdatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		StaffID:         database.Text(CreateTimesheetStaffID),
		LocationID:      database.Text(CreateTimesheetLocationID),
		Remark:          database.Text(CreateTimeSheetRemark),
		TimesheetDate:   database.Timestamptz(CreateTimesheetTimesheetDate),
		TimesheetStatus: database.Text(tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String()),
	}

	timesheetDTO := &dto.Timesheet{
		ID:              CreateTimesheetID,
		StaffID:         CreateTimesheetStaffID,
		LocationID:      CreateTimesheetLocationID,
		Remark:          CreateTimeSheetRemark,
		TimesheetDate:   CreateTimesheetTimesheetDate,
		TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
		ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
			{
				TimesheetID: CreateTimesheetID,
				LessonID:    LessonID,
				FlagOn:      true,
			}},
	}

	timesheetDTOInvalidLHs := &dto.Timesheet{
		ID:              CreateTimesheetID,
		StaffID:         CreateTimesheetStaffID,
		LocationID:      CreateTimesheetLocationID,
		Remark:          CreateTimeSheetRemark,
		TimesheetDate:   CreateTimesheetTimesheetDate,
		TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
		ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
			{
				TimesheetID: CreateTimesheetID,
				LessonID:    LessonID,
				FlagOn:      false,
				IsCreated:   true,
			}},
	}

	testCases := []TestCase{
		{
			name: "happy case create new timesheet",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &dto.Timesheet{
				StaffID:       CreateTimesheetStaffID,
				LocationID:    CreateTimesheetLocationID,
				TimesheetDate: CreateTimesheetTimesheetDate,
				Remark:        CreateTimeSheetRemark,
				ListOtherWorkingHours: []*dto.OtherWorkingHours{
					{TimesheetConfigID: CreateTimesheetConfigID},
				}},
			expectedErr:  nil,
			expectedResp: CreateTimesheetID,
			setup: func(ctx context.Context) {
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).
					Return([]*dto.Timesheet{}, nil).Once()

				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetRepo.On("UpsertMultiple", ctx, tx, mock.Anything).
					Return([]*entity.Timesheet{timesheetE}, nil).Once()
				owhsRepo.On("UpsertMultiple", ctx, tx, mock.Anything).
					Return(nil).Once()
				tx.On("Commit", mock.Anything).Return(nil)
			},
		},
		{
			name: "error case unauthorized to modify timesheet",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDFail),
			req: &dto.Timesheet{
				StaffID:       CreateTimesheetStaffID,
				LocationID:    CreateTimesheetLocationID,
				TimesheetDate: CreateTimesheetTimesheetDate},
			expectedErr: status.Error(
				codes.PermissionDenied,
				fmt.Sprintf("unauthorized to modify timesheet, timesheetStaffID: %s", CreateTimesheetStaffID)),
			expectedResp: "",
			setup:        func(ctx context.Context) {},
		},
		{
			name: "error case failed to execute insert other working hours",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &dto.Timesheet{
				StaffID:       CreateTimesheetStaffID,
				LocationID:    CreateTimesheetLocationID,
				TimesheetDate: CreateTimesheetTimesheetDate,
				Remark:        CreateTimeSheetRemark,
				ListOtherWorkingHours: []*dto.OtherWorkingHours{
					{TimesheetConfigID: CreateTimesheetConfigID},
				}},
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("batchResults.Exec: %s", errors.New("internal error"))),
			expectedResp: "",
			setup: func(ctx context.Context) {
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).
					Return([]*dto.Timesheet{}, nil).Once()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetRepo.On("UpsertMultiple", ctx, tx, mock.Anything).
					Return([]*entity.Timesheet{timesheetE}, nil).Once()
				owhsRepo.On("UpsertMultiple", ctx, tx, mock.Anything).
					Return(fmt.Errorf("batchResults.Exec: %s", errors.New("internal error"))).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},

		{
			name: "error case failed to execute insert list transport expenses",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &dto.Timesheet{
				StaffID:       CreateTimesheetStaffID,
				LocationID:    CreateTimesheetLocationID,
				TimesheetDate: CreateTimesheetTimesheetDate,
				Remark:        CreateTimeSheetRemark,
				ListTransportationExpenses: []*dto.TransportationExpenses{
					{
						TransportExpenseID: CreateTimesheetStaffID,
						TransportationType: "BY_BUS",
						TransportationFrom: "HN",
						TransportationTo:   "HCM",
						CostAmount:         2,
						RoundTrip:          true,
					},
				}},
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("batchResults.Exec: %s", errors.New("internal error"))),
			expectedResp: "",
			setup: func(ctx context.Context) {
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).
					Return([]*dto.Timesheet{}, nil).Once()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetRepo.On("UpsertMultiple", ctx, tx, mock.Anything).
					Return([]*entity.Timesheet{timesheetE}, nil).Once()
				transportExpenseRepo.On("UpsertMultiple", ctx, tx, mock.Anything).
					Return(fmt.Errorf("batchResults.Exec: %s", errors.New("internal error"))).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},

		{
			name: "error case failed to find timesheet",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &dto.Timesheet{
				StaffID:       CreateTimesheetStaffID,
				LocationID:    CreateTimesheetLocationID,
				TimesheetDate: CreateTimesheetTimesheetDate},
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("err find Timesheet: %s", errors.New("internal error"))),
			expectedResp: "",
			setup: func(ctx context.Context) {
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("err find Timesheet: %w", errors.New("internal error"))).Once()
			},
		},
		{
			name: "error case failed to execute insert timesheet",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &dto.Timesheet{
				StaffID:       CreateTimesheetStaffID,
				LocationID:    CreateTimesheetLocationID,
				TimesheetDate: CreateTimesheetTimesheetDate},
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("err upsert Timesheet: %s", errors.New("internal error"))),
			expectedResp: "",
			setup: func(ctx context.Context) {
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).
					Return([]*dto.Timesheet{}, nil).Once()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetRepo.On("UpsertMultiple", ctx, tx, mock.Anything).
					Return(nil, fmt.Errorf("err upsert Timesheet: %w", errors.New("internal error"))).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "error case already exists timesheet",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &dto.Timesheet{
				StaffID:       CreateTimesheetStaffID,
				LocationID:    CreateTimesheetLocationID,
				TimesheetDate: CreateTimesheetTimesheetDate,
			},
			expectedErr:  status.Error(codes.AlreadyExists, constant.ErrorMessageDuplicateTimesheet),
			expectedResp: "",

			setup: func(ctx context.Context) {
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).
					Return([]*dto.Timesheet{timesheetDTO}, nil).Once()
			},
		},
		{
			name: "error case already exists timesheet and update timesheet failed",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &dto.Timesheet{
				StaffID:       CreateTimesheetStaffID,
				LocationID:    CreateTimesheetLocationID,
				TimesheetDate: CreateTimesheetTimesheetDate,
			},
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("err update Timesheet: %s", errors.New("internal error"))),
			expectedResp: "",
			setup: func(ctx context.Context) {
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).
					Return([]*dto.Timesheet{timesheetDTOInvalidLHs}, nil).Once()
				owhsRepo.On("FindListOtherWorkingHoursByTimesheetIDs", ctx, db, mock.Anything).
					Return(nil, nil).Once()
				transportExpenseRepo.On("FindListTransportExpensesByTimesheetIDs", ctx, db, mock.Anything).
					Return(nil, nil).Once()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetRepo.On("UpsertMultiple", ctx, tx, mock.Anything).
					Return(nil, fmt.Errorf("err update Timesheet: %w", errors.New("internal error"))).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*dto.Timesheet)
			resp, err := s.CreateTimesheet(testCase.ctx, req)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}

func TestTimesheetService_UpdateTimesheet(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var (
		timesheetRepo        = new(mock_repositories.MockTimesheetRepoImpl)
		owhsRepo             = new(mock_repositories.MockOtherWorkingHoursRepoImpl)
		transportExpenseRepo = new(mock_repositories.MockTransportationExpenseRepoImpl)
		lessonHoursRepo      = new(mock_repositories.MockTimesheetLessonHoursRepoImpl)
		db                   = new(mock_database.Ext)
		tx                   = new(mock_database.Tx)
		mockJsm              = new(mock_nats.JetStreamManagement)
	)

	s := ServiceImpl{
		DB:                        db,
		JSM:                       mockJsm,
		TimesheetRepo:             timesheetRepo,
		OtherWorkingHoursRepo:     owhsRepo,
		TransportationExpenseRepo: transportExpenseRepo,
		TimesheetLessonHoursRepo:  lessonHoursRepo,
	}

	timesheetE := &entity.Timesheet{
		TimesheetID:     database.Text(UpdateTimesheetID),
		StaffID:         database.Text(CreateTimesheetStaffID),
		LocationID:      database.Text(CreateTimesheetLocationID),
		Remark:          database.Text(CreateTimeSheetRemark),
		CreatedAt:       database.Timestamptz(now),
		UpdatedAt:       database.Timestamptz(now),
		TimesheetDate:   database.Timestamptz(now),
		TimesheetStatus: database.Text(tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String()),
	}

	listOWHsE := []*entity.OtherWorkingHours{
		{
			ID:                database.Text(UpdateOWHsID),
			TimesheetID:       database.Text(UpdateTimesheetID),
			TimesheetConfigID: database.Text(UpdateTimesheetConfigID),
			CreatedAt:         database.Timestamptz(now),
			UpdatedAt:         database.Timestamptz(now),
		},
	}

	listTransportExpenses := []*entity.TransportationExpense{
		{
			TransportationExpenseID: database.Text(UpdateOWHsID),
			TransportationType:      database.Text(tpb.TransportationType_TYPE_BUS.String()),
			TransportationFrom:      database.Text(CreateTimeSheetRemark),
			TransportationTo:        database.Text(CreateTimeSheetRemark),
			CostAmount:              database.Int4(10),
			RoundTrip:               database.Bool(true),
			CreatedAt:               database.Timestamptz(now),
			UpdatedAt:               database.Timestamptz(now),
		},
	}

	timesheetSubmitted := &entity.Timesheet{
		TimesheetID:     database.Text(UpdateTimesheetID),
		StaffID:         database.Text(CreateTimesheetStaffID),
		LocationID:      database.Text(CreateTimesheetLocationID),
		Remark:          database.Text(CreateTimeSheetRemark),
		CreatedAt:       database.Timestamptz(now),
		UpdatedAt:       database.Timestamptz(now),
		TimesheetDate:   database.Timestamptz(now),
		TimesheetStatus: database.Text(tpb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String()),
	}
	timesheetApproved := &entity.Timesheet{
		TimesheetID:     database.Text(UpdateTimesheetID),
		StaffID:         database.Text(CreateTimesheetStaffID),
		LocationID:      database.Text(CreateTimesheetLocationID),
		Remark:          database.Text(CreateTimeSheetRemark),
		CreatedAt:       database.Timestamptz(now),
		UpdatedAt:       database.Timestamptz(now),
		TimesheetDate:   database.Timestamptz(now),
		TimesheetStatus: database.Text(tpb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String()),
	}
	timesheetConfirmed := &entity.Timesheet{
		TimesheetID:     database.Text(UpdateTimesheetID),
		StaffID:         database.Text(CreateTimesheetStaffID),
		LocationID:      database.Text(CreateTimesheetLocationID),
		Remark:          database.Text(CreateTimeSheetRemark),
		CreatedAt:       database.Timestamptz(now),
		UpdatedAt:       database.Timestamptz(now),
		TimesheetDate:   database.Timestamptz(now),
		TimesheetStatus: database.Text(tpb.TimesheetStatus_TIMESHEET_STATUS_CONFIRMED.String()),
	}

	timesheetExistsLessonHours := map[string]struct{}{UpdateTimesheetID: {}}
	testCases := []TestCase{
		{
			name:        "happy case without other working hours",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req:         &dto.Timesheet{ID: UpdateTimesheetID},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetID", ctx, db, mock.Anything).
					Return(timesheetE, nil).Once()
				lessonHoursRepo.On("MapExistingLessonHoursByTimesheetIds", ctx, db, mock.Anything).
					Return(timesheetExistsLessonHours, nil).Once()
				owhsRepo.On("FindListOtherWorkingHoursByTimesheetIDs", ctx, db, mock.Anything).
					Return(nil, nil).Once()
				transportExpenseRepo.On("FindListTransportExpensesByTimesheetIDs", ctx, db, mock.Anything).
					Return(nil, nil).Once()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)

				timesheetRepo.On("UpdateTimeSheet", ctx, tx, mock.Anything).
					Return(timesheetE, nil).Once()

				tx.On("Commit", mock.Anything).Return(nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name: "happy case with other working hours update and insert",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &dto.Timesheet{
				ID:     UpdateTimesheetID,
				Remark: CreateTimeSheetRemark,
				ListOtherWorkingHours: []*dto.OtherWorkingHours{
					{ // update
						ID:                UpdateOWHsID,
						TimesheetConfigID: UpdateTimesheetConfigID,
					},
					{ // insert
						TimesheetConfigID: UpdateTimesheetConfigID,
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetID", ctx, db, mock.Anything).
					Return(timesheetE, nil).Once()
				lessonHoursRepo.On("MapExistingLessonHoursByTimesheetIds", ctx, db, mock.Anything).
					Return(timesheetExistsLessonHours, nil).Once()
				owhsRepo.On("FindListOtherWorkingHoursByTimesheetIDs", ctx, db, mock.Anything).
					Return(listOWHsE, nil).Once()
				transportExpenseRepo.On("FindListTransportExpensesByTimesheetIDs", ctx, db, mock.Anything).
					Return(nil, nil).Once()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)

				timesheetRepo.On("UpdateTimeSheet", ctx, tx, mock.Anything).
					Return(timesheetE, nil).Once()
				owhsRepo.On("UpsertMultiple", ctx, tx, mock.Anything).
					Return(nil).Once()
				tx.On("Commit", mock.Anything).Return(nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name: "happy case with transport expenses",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &dto.Timesheet{
				StaffID:       CreateTimesheetStaffID,
				LocationID:    CreateTimesheetLocationID,
				TimesheetDate: CreateTimesheetTimesheetDate,
				Remark:        CreateTimeSheetRemark,
				ListTransportationExpenses: []*dto.TransportationExpenses{
					{
						TransportationType: CreateTimesheetStaffID,
						TransportationFrom: "HN",
						TransportationTo:   "HAI PHONG",
						CostAmount:         18,
						RoundTrip:          true,
						Remarks:            "test update timesheet with transport expenses",
						IsDeleted:          false,
					},
					{
						TransportationType: CreateTimesheetStaffID,
						TransportationFrom: "DA NANG",
						TransportationTo:   "HCM",
						CostAmount:         10,
						RoundTrip:          false,
						Remarks:            "test update timesheet with transport expenses",
						IsDeleted:          true,
					},
				}},
			expectedErr:  nil,
			expectedResp: CreateTimesheetID,
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetID", ctx, db, mock.Anything).
					Return(timesheetE, nil).Once()
				lessonHoursRepo.On("MapExistingLessonHoursByTimesheetIds", ctx, db, mock.Anything).
					Return(timesheetExistsLessonHours, nil).Once()
				transportExpenseRepo.On("FindListTransportExpensesByTimesheetIDs", ctx, db, mock.Anything).
					Return(listTransportExpenses, nil).Once()
				owhsRepo.On("FindListOtherWorkingHoursByTimesheetIDs", ctx, db, mock.Anything).
					Return(nil, nil).Once()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetRepo.On("UpdateTimeSheet", ctx, tx, mock.Anything).
					Return(timesheetE, nil).Once()
				transportExpenseRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil).Once()
				tx.On("Commit", mock.Anything).Return(nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:        "error case publish timesheet action log event",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req:         &dto.Timesheet{ID: UpdateTimesheetID},
			expectedErr: status.Error(codes.Internal, "PublishActionLogTimesheetEvent JSM.PublishAsyncContext failed, msgID: MsgID, Error"),
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetID", ctx, db, mock.Anything).
					Return(timesheetE, nil).Once()
				lessonHoursRepo.On("MapExistingLessonHoursByTimesheetIds", ctx, db, mock.Anything).
					Return(timesheetExistsLessonHours, nil).Once()
				owhsRepo.On("FindListOtherWorkingHoursByTimesheetIDs", ctx, db, mock.Anything).
					Return(nil, nil).Once()
				transportExpenseRepo.On("FindListTransportExpensesByTimesheetIDs", ctx, db, mock.Anything).
					Return(nil, nil).Once()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)

				timesheetRepo.On("UpdateTimeSheet", ctx, tx, mock.Anything).
					Return(timesheetE, nil).Once()

				tx.On("Commit", mock.Anything).Return(nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("MsgID", fmt.Errorf("Error"))

			},
		},
		{
			name:        "error case can not check timesheet lesson hours exist by timesheetID",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req:         &dto.Timesheet{ID: UpdateTimesheetID},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("get list timesheet lesson hours by timesheet ids error: %s", pgx.ErrNoRows.Error())),
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetID", ctx, db, mock.Anything).
					Return(timesheetE, nil).Once()
				lessonHoursRepo.On("MapExistingLessonHoursByTimesheetIds", ctx, db, mock.Anything).
					Return(map[string]struct{}{}, pgx.ErrNoRows).Once()
			},
		},
		{
			name:        "error case can not check timesheet lesson hours exist by timesheetID",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req:         &dto.Timesheet{ID: UpdateTimesheetID},
			expectedErr: status.Error(codes.Internal, "timesheet empty, cannot update anything"),
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetID", ctx, db, mock.Anything).
					Return(timesheetE, nil).Once()
				lessonHoursRepo.On("MapExistingLessonHoursByTimesheetIds", ctx, db, mock.Anything).
					Return(map[string]struct{}{}, nil).Once()
			},
		},
		{
			name:        "error case can not find timesheet with timesheetID request",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDFail),
			req:         &dto.Timesheet{ID: UpdateTimesheetID},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("find timesheet error: %s", pgx.ErrNoRows.Error())),
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetID", ctx, db, mock.Anything).
					Return(nil, pgx.ErrNoRows).Once()
			},
		},
		{
			name:        "error case unauthorized to modify timesheet",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDFail),
			req:         &dto.Timesheet{ID: UpdateTimesheetID},
			expectedErr: status.Error(codes.PermissionDenied, fmt.Sprintf("unauthorized to modify timesheet, timesheetStaffID: %s", CreateTimesheetStaffID)),
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetID", ctx, db, mock.Anything).
					Return(timesheetE, nil).Once()
			},
		},
		{
			name:        "error case unauthorized to modify timesheet submitted by requester",
			ctx:         interceptors.ContextWithUserGroup(ctx, constant.RoleTeacher),
			req:         &dto.Timesheet{ID: UpdateTimesheetID},
			expectedErr: status.Error(codes.PermissionDenied, fmt.Sprintf("unauthorized to modify timesheet, timesheetStaffID: %s", CreateTimesheetStaffID)),
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetID", ctx, db, mock.Anything).
					Return(timesheetSubmitted, nil).Once()
			},
		},
		{
			name:        "error case unauthorized to modify timesheet approved by requester",
			ctx:         interceptors.ContextWithUserGroup(ctx, constant.RoleTeacher),
			req:         &dto.Timesheet{ID: UpdateTimesheetID},
			expectedErr: status.Error(codes.PermissionDenied, fmt.Sprintf("unauthorized to modify timesheet, timesheetStaffID: %s", CreateTimesheetStaffID)),
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetID", ctx, db, mock.Anything).
					Return(timesheetApproved, nil).Once()
			},
		},
		{
			name:        "error case unauthorized to modify timesheet confirmed by requester",
			ctx:         interceptors.ContextWithUserGroup(ctx, constant.RoleTeacher),
			req:         &dto.Timesheet{ID: UpdateTimesheetID},
			expectedErr: status.Error(codes.PermissionDenied, fmt.Sprintf("unauthorized to modify timesheet, timesheetStaffID: %s", CreateTimesheetStaffID)),
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetID", ctx, db, mock.Anything).
					Return(timesheetConfirmed, nil).Once()
			},
		},
		{
			name:        "error case unauthorized to modify timesheet approved by approver/confirmer",
			ctx:         interceptors.ContextWithUserGroup(ctx, constant.RoleSchoolAdmin),
			req:         &dto.Timesheet{ID: UpdateTimesheetID},
			expectedErr: status.Error(codes.PermissionDenied, fmt.Sprintf("unauthorized to modify timesheet, timesheetStaffID: %s", CreateTimesheetStaffID)),
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetID", ctx, db, mock.Anything).
					Return(timesheetApproved, nil).Once()
			},
		},
		{
			name:        "error case unauthorized to modify timesheet confirmed by approver/confirmer",
			ctx:         interceptors.ContextWithUserGroup(ctx, constant.RoleSchoolAdmin),
			req:         &dto.Timesheet{ID: UpdateTimesheetID},
			expectedErr: status.Error(codes.PermissionDenied, fmt.Sprintf("unauthorized to modify timesheet, timesheetStaffID: %s", CreateTimesheetStaffID)),
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetID", ctx, db, mock.Anything).
					Return(timesheetConfirmed, nil).Once()
			},
		},

		{
			name:        "error case failed to execute update timesheet",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req:         &dto.Timesheet{ID: UpdateTimesheetID},
			expectedErr: status.Error(codes.Internal, "Too many timesheet"),
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetID", ctx, db, mock.Anything).
					Return(timesheetE, nil).Once()
				lessonHoursRepo.On("MapExistingLessonHoursByTimesheetIds", ctx, db, mock.Anything).
					Return(timesheetExistsLessonHours, nil).Once()
				owhsRepo.On("FindListOtherWorkingHoursByTimesheetIDs", ctx, db, mock.Anything).
					Return(nil, nil).Once()
				transportExpenseRepo.On("FindListTransportExpensesByTimesheetIDs", ctx, db, mock.Anything).Return(nil, nil).Once()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)

				timesheetRepo.On("UpdateTimeSheet", ctx, tx, mock.Anything).
					Return(nil, errors.New("Too many timesheet")).Once()

				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "error case failed to Retrieve list other working hours by timesheet id",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &dto.Timesheet{
				ID: UpdateTimesheetID,
				ListOtherWorkingHours: []*dto.OtherWorkingHours{
					{ // update
						ID:                UpdateOWHsID,
						TimesheetConfigID: UpdateTimesheetConfigID,
					},
				},
			},
			expectedErr: status.Error(codes.Internal, "get list other working hours error: "+pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetID", ctx, db, mock.Anything).
					Return(timesheetE, nil).Once()
				lessonHoursRepo.On("MapExistingLessonHoursByTimesheetIds", ctx, db, mock.Anything).
					Return(timesheetExistsLessonHours, nil).Once()
				owhsRepo.On("FindListOtherWorkingHoursByTimesheetIDs", ctx, db, mock.Anything).
					Return(nil, pgx.ErrNoRows).Once()
			},
		},
		{
			name: "error case failed to execute update other working hours",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &dto.Timesheet{
				ID: UpdateTimesheetID,
				ListOtherWorkingHours: []*dto.OtherWorkingHours{
					{ // update
						ID:                UpdateOWHsID,
						TimesheetConfigID: UpdateTimesheetConfigID,
					},
				},
			},
			expectedErr: status.Error(codes.Internal, "failed"),
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetID", ctx, db, mock.Anything).
					Return(timesheetE, nil).Once()
				lessonHoursRepo.On("MapExistingLessonHoursByTimesheetIds", ctx, db, mock.Anything).
					Return(timesheetExistsLessonHours, nil).Once()
				owhsRepo.On("FindListOtherWorkingHoursByTimesheetIDs", ctx, db, mock.Anything).
					Return(listOWHsE, nil).Once()
				transportExpenseRepo.On("FindListTransportExpensesByTimesheetIDs", ctx, db, mock.Anything).
					Return(nil, nil).Once()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)

				timesheetRepo.On("UpdateTimeSheet", ctx, tx, mock.Anything).
					Return(timesheetE, nil).Once()
				owhsRepo.On("UpsertMultiple", ctx, tx, mock.Anything).
					Return(errors.New("failed")).Once()

				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},

		{
			name: "error case failed to execute insert other working hours",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &dto.Timesheet{
				ID: UpdateTimesheetID,
				ListOtherWorkingHours: []*dto.OtherWorkingHours{
					{ // insert
						TimesheetConfigID: UpdateTimesheetConfigID,
					},
				},
			},
			expectedErr: status.Error(codes.Internal, "failed"),
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetID", ctx, db, mock.Anything).
					Return(timesheetE, nil).Once()
				lessonHoursRepo.On("MapExistingLessonHoursByTimesheetIds", ctx, db, mock.Anything).
					Return(timesheetExistsLessonHours, nil).Once()
				owhsRepo.On("FindListOtherWorkingHoursByTimesheetIDs", ctx, db, mock.Anything).
					Return(listOWHsE, nil).Once()
				transportExpenseRepo.On("FindListTransportExpensesByTimesheetIDs", ctx, db, mock.Anything).Return(nil, nil).Once()

				db.On("Begin", mock.Anything).Once().Return(tx, nil)

				timesheetRepo.On("UpdateTimeSheet", ctx, tx, mock.Anything).
					Return(timesheetE, nil).Once()
				owhsRepo.On("UpsertMultiple", ctx, tx, mock.Anything).
					Return(errors.New("failed")).Once()

				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
	}

	// Do test
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*dto.Timesheet)
			err := s.UpdateTimesheet(testCase.ctx, req)
			assert.Equal(t, testCase.expectedErr, err)
			mock.AssertExpectationsForObjects(
				t,
				timesheetRepo,
				owhsRepo,
				db,
				tx,
			)
		})
	}
}
