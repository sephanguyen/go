package timesheet

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/timesheet/domain/common"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/timesheet/repository"
	tpb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	GetTimesheetStaffID               = "get_timesheet_staff_id"
	GetTimesheetLocationID            = "get_timesheet_location_id"
	TimesheetLessonHoursLessonID1     = "timesheet_lesson_hours_lesson_id_1"
	TimesheetLessonHoursLessonID2     = "timesheet_lesson_hours_lesson_id_2"
	TimesheetLessonHoursLessonID3     = "timesheet_lesson_hours_lesson_id_3"
	TimesheetID1                      = "get_timesheet_timesheet_id_1"
	TimesheetID2                      = "get_timesheet_timesheet_id_2"
	GetTimesheetTimesheetDate         = time.Now()
	TimesheetTransportationExpenseID1 = "timesheet_transportation_expense_id_1"
	TimesheetTransportationExpenseID2 = "timesheet_transportation_expense_id_2"
)

func TestGetTimesheetServiceImpl_GetTimesheet(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		timesheetRepo             = new(mock_repositories.MockTimesheetRepoImpl)
		timesheetLessonHoursRepo  = new(mock_repositories.MockTimesheetLessonHoursRepoImpl)
		otherWorkingHoursRepo     = new(mock_repositories.MockOtherWorkingHoursRepoImpl)
		transportationExpenseRepo = new(mock_repositories.MockTransportationExpenseRepoImpl)
		db                        = new(mock_database.Ext)
	)

	s := GetTimesheetServiceImpl{
		DB:                        db,
		TimesheetRepo:             timesheetRepo,
		TimesheetLessonHoursRepo:  timesheetLessonHoursRepo,
		OtherWorkingHoursRepo:     otherWorkingHoursRepo,
		TransportationExpenseRepo: transportationExpenseRepo,
	}

	timesheetEntities := []*entity.Timesheet{
		{
			TimesheetID:     database.Text(TimesheetID1),
			CreatedAt:       database.Timestamptz(now),
			UpdatedAt:       database.Timestamptz(now),
			StaffID:         database.Text(GetTimesheetStaffID),
			LocationID:      database.Text(GetTimesheetLocationID),
			TimesheetDate:   database.Timestamptz(GetTimesheetTimesheetDate),
			TimesheetStatus: database.Text(tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String()),
		},
		{
			TimesheetID:     database.Text(TimesheetID2),
			CreatedAt:       database.Timestamptz(now),
			UpdatedAt:       database.Timestamptz(now),
			StaffID:         database.Text(GetTimesheetStaffID),
			LocationID:      database.Text(GetTimesheetLocationID),
			TimesheetDate:   database.Timestamptz(GetTimesheetTimesheetDate),
			TimesheetStatus: database.Text(tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String()),
		},
	}
	timesheetLessonHoursEntities := []*entity.TimesheetLessonHours{
		{
			TimesheetID: database.Text(TimesheetID1),
			LessonID:    database.Text(TimesheetLessonHoursLessonID1),
			CreatedAt:   database.Timestamptz(now),
			UpdatedAt:   database.Timestamptz(now),
			DeletedAt:   pgtype.Timestamptz{Status: pgtype.Null},
		},
		{
			TimesheetID: database.Text(TimesheetID1),
			LessonID:    database.Text(TimesheetLessonHoursLessonID2),
			CreatedAt:   database.Timestamptz(now),
			UpdatedAt:   database.Timestamptz(now),
			DeletedAt:   pgtype.Timestamptz{Status: pgtype.Null},
		},
		{
			TimesheetID: database.Text(TimesheetID2),
			LessonID:    database.Text(TimesheetLessonHoursLessonID3),
			CreatedAt:   database.Timestamptz(now),
			UpdatedAt:   database.Timestamptz(now),
			DeletedAt:   pgtype.Timestamptz{Status: pgtype.Null},
		},
	}
	otherWorkingHoursEntities := []*entity.OtherWorkingHours{
		{
			TimesheetID: database.Text(TimesheetID1),
			StartTime:   database.Timestamptz(now),
			EndTime:     database.Timestamptz(now),
			CreatedAt:   database.Timestamptz(now),
			UpdatedAt:   database.Timestamptz(now),
			DeletedAt:   pgtype.Timestamptz{Status: pgtype.Null},
		},
		{
			TimesheetID: database.Text(TimesheetID1),
			StartTime:   database.Timestamptz(now),
			EndTime:     database.Timestamptz(now),
			CreatedAt:   database.Timestamptz(now),
			UpdatedAt:   database.Timestamptz(now),
			DeletedAt:   pgtype.Timestamptz{Status: pgtype.Null},
		},
		{
			TimesheetID: database.Text(TimesheetID2),
			StartTime:   database.Timestamptz(now),
			EndTime:     database.Timestamptz(now),
			CreatedAt:   database.Timestamptz(now),
			UpdatedAt:   database.Timestamptz(now),
			DeletedAt:   pgtype.Timestamptz{Status: pgtype.Null},
		},
	}
	transportExpenseEntities := []*entity.TransportationExpense{
		{
			TransportationExpenseID: database.Text(idutil.ULIDNow()),
			TimesheetID:             database.Text(TimesheetID1),
		},
	}
	timesheetDtos := []*dto.Timesheet{
		{
			ID:              TimesheetID1,
			StaffID:         GetTimesheetStaffID,
			LocationID:      GetTimesheetLocationID,
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   GetTimesheetTimesheetDate,
			ListOtherWorkingHours: []*dto.OtherWorkingHours{
				{
					TimesheetID: TimesheetID1,
					StartTime:   now,
					EndTime:     now,
				},
				{
					TimesheetID: TimesheetID1,
					StartTime:   now,
					EndTime:     now,
				},
			},
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: TimesheetID1,
					LessonID:    TimesheetLessonHoursLessonID1,
					IsCreated:   true,
				},
				{
					TimesheetID: TimesheetID1,
					LessonID:    TimesheetLessonHoursLessonID2,
					IsCreated:   true,
				},
			},
			ListTransportationExpenses: []*dto.TransportationExpenses{
				{
					TransportExpenseID: TimesheetTransportationExpenseID1,
					TimesheetID:        TimesheetID1,
				},
				{
					TransportExpenseID: TimesheetTransportationExpenseID2,
					TimesheetID:        TimesheetID1,
				},
			},
			IsCreated: true,
		},
		{
			ID:              TimesheetID2,
			StaffID:         GetTimesheetStaffID,
			LocationID:      GetTimesheetLocationID,
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   GetTimesheetTimesheetDate,
			ListOtherWorkingHours: []*dto.OtherWorkingHours{
				{
					TimesheetID: TimesheetID2,
					StartTime:   now,
					EndTime:     now,
				},
			},
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: TimesheetID2,
					LessonID:    TimesheetLessonHoursLessonID3,
					IsCreated:   true,
				},
			},
			IsCreated: true,
		},
	}
	timesheetDtosWithoutOtherWorkingHours := []*dto.Timesheet{
		{
			ID:              TimesheetID1,
			StaffID:         GetTimesheetStaffID,
			LocationID:      GetTimesheetLocationID,
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   GetTimesheetTimesheetDate,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: TimesheetID1,
					LessonID:    TimesheetLessonHoursLessonID1,
					IsCreated:   true,
				},
				{
					TimesheetID: TimesheetID1,
					LessonID:    TimesheetLessonHoursLessonID2,
					IsCreated:   true,
				},
			},
			IsCreated: true,
		},
		{
			ID:              TimesheetID2,
			StaffID:         GetTimesheetStaffID,
			LocationID:      GetTimesheetLocationID,
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   GetTimesheetTimesheetDate,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: TimesheetID2,
					LessonID:    TimesheetLessonHoursLessonID3,
					IsCreated:   true,
				},
			},
			IsCreated: true,
		},
	}
	timesheetDtosWithoutTimesheetLessonHours := []*dto.Timesheet{
		{
			ID:                       TimesheetID1,
			StaffID:                  GetTimesheetStaffID,
			LocationID:               GetTimesheetLocationID,
			TimesheetStatus:          tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:            GetTimesheetTimesheetDate,
			ListOtherWorkingHours:    nil,
			ListTimesheetLessonHours: nil,
			IsCreated:                true,
		},
		{
			ID:                       TimesheetID2,
			StaffID:                  GetTimesheetStaffID,
			LocationID:               GetTimesheetLocationID,
			TimesheetStatus:          tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:            GetTimesheetTimesheetDate,
			ListOtherWorkingHours:    nil,
			ListTimesheetLessonHours: nil,
			IsCreated:                true,
		},
	}
	testCases := []struct {
		name             string
		ctx              context.Context
		timesheetArgs    interface{}
		timesheetOptions interface{}
		expectedResp     []*dto.Timesheet
		expectedErr      error
		setup            func(ctx context.Context)
	}{
		{
			name: "get timesheet success",
			ctx:  ctx,
			timesheetArgs: &dto.TimesheetQueryArgs{
				StaffIDs:      []string{GetTimesheetStaffID},
				LocationID:    GetTimesheetLocationID,
				TimesheetDate: CreateTimesheetTimesheetDate,
			},
			timesheetOptions: &dto.TimesheetGetOptions{
				IsGetListOtherWorkingHours:     true,
				IsGetListTimesheetLessonHours:  true,
				IsGetListTransportationExpense: true,
			},
			expectedResp: timesheetDtos,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetArgs", ctx, db, mock.Anything).Return(timesheetEntities, nil).Once()
				timesheetLessonHoursRepo.On("FindByTimesheetIDs", ctx, db, mock.Anything).Return(timesheetLessonHoursEntities, nil).Once()
				otherWorkingHoursRepo.On("FindListOtherWorkingHoursByTimesheetIDs", ctx, db, mock.Anything).Return(otherWorkingHoursEntities, nil).Once()
				transportationExpenseRepo.On("FindListTransportExpensesByTimesheetIDs", ctx, db, mock.Anything).Return(transportExpenseEntities, nil).Once()
			},
		},
		{
			name: "get timesheet success with get options GetListTimesheetLessonHours true",
			ctx:  ctx,
			timesheetArgs: &dto.TimesheetQueryArgs{
				StaffIDs:      []string{GetTimesheetStaffID},
				LocationID:    GetTimesheetLocationID,
				TimesheetDate: CreateTimesheetTimesheetDate,
			},
			timesheetOptions: &dto.TimesheetGetOptions{
				IsGetListOtherWorkingHours:    false,
				IsGetListTimesheetLessonHours: true,
			},
			expectedResp: timesheetDtosWithoutOtherWorkingHours,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetArgs", ctx, db, mock.Anything).Return(timesheetEntities, nil).Once()
				timesheetLessonHoursRepo.On("FindByTimesheetIDs", ctx, db, mock.Anything).Return(timesheetLessonHoursEntities, nil).Once()
			},
		},
		{
			name: "get timesheet success with get options GetListTimesheetLessonHours false",
			ctx:  ctx,
			timesheetArgs: &dto.TimesheetQueryArgs{
				StaffIDs:      []string{GetTimesheetStaffID},
				LocationID:    GetTimesheetLocationID,
				TimesheetDate: CreateTimesheetTimesheetDate,
			},
			timesheetOptions: &dto.TimesheetGetOptions{
				IsGetListOtherWorkingHours:    false,
				IsGetListTimesheetLessonHours: false,
			},
			expectedResp: timesheetDtosWithoutTimesheetLessonHours,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetArgs", ctx, db, mock.Anything).Return(timesheetEntities, nil).Once()
			},
		},
		{
			name: "get timesheet success with nil timesheet lesson hours",
			ctx:  ctx,
			timesheetArgs: &dto.TimesheetQueryArgs{
				StaffIDs:      []string{GetTimesheetStaffID},
				LocationID:    GetTimesheetLocationID,
				TimesheetDate: CreateTimesheetTimesheetDate,
			},
			timesheetOptions: &dto.TimesheetGetOptions{
				IsGetListOtherWorkingHours:    false,
				IsGetListTimesheetLessonHours: true,
			},
			expectedResp: timesheetDtosWithoutTimesheetLessonHours,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetArgs", ctx, db, mock.Anything).Return(timesheetEntities, nil).Once()
				timesheetLessonHoursRepo.On("FindByTimesheetIDs", ctx, db, mock.Anything).Return(nil, nil).Once()
			},
		},
		{
			name: "get timesheet not found timesheet",
			ctx:  ctx,
			timesheetArgs: &dto.TimesheetQueryArgs{
				StaffIDs:      []string{GetTimesheetStaffID},
				LocationID:    GetTimesheetLocationID,
				TimesheetDate: CreateTimesheetTimesheetDate,
			},
			timesheetOptions: &dto.TimesheetGetOptions{
				IsGetListOtherWorkingHours:    false,
				IsGetListTimesheetLessonHours: true,
			},
			expectedResp: []*dto.Timesheet(nil),
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetArgs", ctx, db, mock.Anything).Return(nil, nil).Once()
			},
		},
		{
			name: "get timesheet failed TimesheetRepo FindTimesheetByTimesheetArgs error",
			ctx:  ctx,
			timesheetArgs: &dto.TimesheetQueryArgs{
				StaffIDs:      []string{GetTimesheetStaffID},
				LocationID:    GetTimesheetLocationID,
				TimesheetDate: CreateTimesheetTimesheetDate,
			},
			timesheetOptions: &dto.TimesheetGetOptions{
				IsGetListOtherWorkingHours:    false,
				IsGetListTimesheetLessonHours: true,
			},
			expectedResp: []*dto.Timesheet(nil),
			expectedErr:  errors.New("connection refused"),
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetArgs", ctx, db, mock.Anything).Return(nil, errors.New("connection refused")).Once()
			},
		},
		{
			name: "get timesheet failed TimesheetLessonHoursRepo FindByTimesheetIDs error",
			ctx:  ctx,
			timesheetArgs: &dto.TimesheetQueryArgs{
				StaffIDs:      []string{GetTimesheetStaffID},
				LocationID:    GetTimesheetLocationID,
				TimesheetDate: CreateTimesheetTimesheetDate,
			},
			timesheetOptions: &dto.TimesheetGetOptions{
				IsGetListOtherWorkingHours:    false,
				IsGetListTimesheetLessonHours: true,
			},
			expectedResp: []*dto.Timesheet(nil),
			expectedErr:  errors.New("connection refused"),
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByTimesheetArgs", ctx, db, mock.Anything).Return(timesheetEntities, nil).Once()
				timesheetLessonHoursRepo.On("FindByTimesheetIDs", ctx, db, mock.Anything).Return(nil, errors.New("connection refused")).Once()
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.GetTimesheet(testCase.ctx, testCase.timesheetArgs.(*dto.TimesheetQueryArgs), testCase.timesheetOptions.(*dto.TimesheetGetOptions))
			assert.Equal(t, testCase.expectedErr, err)
			assert.True(t, common.CompareListTimesheet(testCase.expectedResp, resp))
		})
	}
}

func TestGetTimesheetServiceImpl_GetTimesheetByLessonIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		timesheetRepo             = new(mock_repositories.MockTimesheetRepoImpl)
		timesheetLessonHoursRepo  = new(mock_repositories.MockTimesheetLessonHoursRepoImpl)
		otherWorkingHoursRepo     = new(mock_repositories.MockOtherWorkingHoursRepoImpl)
		transportationExpenseRepo = new(mock_repositories.MockTransportationExpenseRepoImpl)
		db                        = new(mock_database.Ext)
	)

	s := GetTimesheetServiceImpl{
		DB:                        db,
		TimesheetRepo:             timesheetRepo,
		TimesheetLessonHoursRepo:  timesheetLessonHoursRepo,
		OtherWorkingHoursRepo:     otherWorkingHoursRepo,
		TransportationExpenseRepo: transportationExpenseRepo,
	}

	timesheetEntities := []*entity.Timesheet{
		{
			TimesheetID:     database.Text(TimesheetID1),
			CreatedAt:       database.Timestamptz(now),
			UpdatedAt:       database.Timestamptz(now),
			StaffID:         database.Text(GetTimesheetStaffID),
			LocationID:      database.Text(GetTimesheetLocationID),
			TimesheetDate:   database.Timestamptz(GetTimesheetTimesheetDate),
			TimesheetStatus: database.Text(tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String()),
		},
		{
			TimesheetID:     database.Text(TimesheetID2),
			CreatedAt:       database.Timestamptz(now),
			UpdatedAt:       database.Timestamptz(now),
			StaffID:         database.Text(GetTimesheetStaffID),
			LocationID:      database.Text(GetTimesheetLocationID),
			TimesheetDate:   database.Timestamptz(GetTimesheetTimesheetDate),
			TimesheetStatus: database.Text(tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String()),
		},
	}
	timesheetLessonHoursEntities := []*entity.TimesheetLessonHours{
		{
			TimesheetID: database.Text(TimesheetID1),
			LessonID:    database.Text(TimesheetLessonHoursLessonID1),
			CreatedAt:   database.Timestamptz(now),
			UpdatedAt:   database.Timestamptz(now),
			DeletedAt:   pgtype.Timestamptz{Status: pgtype.Null},
		},
		{
			TimesheetID: database.Text(TimesheetID1),
			LessonID:    database.Text(TimesheetLessonHoursLessonID2),
			CreatedAt:   database.Timestamptz(now),
			UpdatedAt:   database.Timestamptz(now),
			DeletedAt:   pgtype.Timestamptz{Status: pgtype.Null},
		},
		{
			TimesheetID: database.Text(TimesheetID2),
			LessonID:    database.Text(TimesheetLessonHoursLessonID3),
			CreatedAt:   database.Timestamptz(now),
			UpdatedAt:   database.Timestamptz(now),
			DeletedAt:   pgtype.Timestamptz{Status: pgtype.Null},
		},
	}
	otherWorkingHoursEntities := []*entity.OtherWorkingHours{
		{
			TimesheetID: database.Text(TimesheetID1),
			StartTime:   database.Timestamptz(now),
			EndTime:     database.Timestamptz(now),
			CreatedAt:   database.Timestamptz(now),
			UpdatedAt:   database.Timestamptz(now),
			DeletedAt:   pgtype.Timestamptz{Status: pgtype.Null},
		},
		{
			TimesheetID: database.Text(TimesheetID1),
			StartTime:   database.Timestamptz(now),
			EndTime:     database.Timestamptz(now),
			CreatedAt:   database.Timestamptz(now),
			UpdatedAt:   database.Timestamptz(now),
			DeletedAt:   pgtype.Timestamptz{Status: pgtype.Null},
		},
		{
			TimesheetID: database.Text(TimesheetID2),
			StartTime:   database.Timestamptz(now),
			EndTime:     database.Timestamptz(now),
			CreatedAt:   database.Timestamptz(now),
			UpdatedAt:   database.Timestamptz(now),
			DeletedAt:   pgtype.Timestamptz{Status: pgtype.Null},
		},
	}
	transportExpenseEntities := []*entity.TransportationExpense{
		{
			TransportationExpenseID: database.Text(idutil.ULIDNow()),
			TimesheetID:             database.Text(TimesheetID1),
		},
	}
	timesheetDtos := []*dto.Timesheet{
		{
			ID:              TimesheetID1,
			StaffID:         GetTimesheetStaffID,
			LocationID:      GetTimesheetLocationID,
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   GetTimesheetTimesheetDate,
			ListOtherWorkingHours: []*dto.OtherWorkingHours{
				{
					TimesheetID: TimesheetID1,
					StartTime:   now,
					EndTime:     now,
				},
				{
					TimesheetID: TimesheetID1,
					StartTime:   now,
					EndTime:     now,
				},
			},
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: TimesheetID1,
					LessonID:    TimesheetLessonHoursLessonID1,
					IsCreated:   true,
				},
				{
					TimesheetID: TimesheetID1,
					LessonID:    TimesheetLessonHoursLessonID2,
					IsCreated:   true,
				},
			},
			IsCreated: true,
		},
		{
			ID:              TimesheetID2,
			StaffID:         GetTimesheetStaffID,
			LocationID:      GetTimesheetLocationID,
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   GetTimesheetTimesheetDate,
			ListOtherWorkingHours: []*dto.OtherWorkingHours{
				{
					TimesheetID: TimesheetID2,
					StartTime:   now,
					EndTime:     now,
				},
			},
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: TimesheetID2,
					LessonID:    TimesheetLessonHoursLessonID3,
					IsCreated:   true,
				},
			},
			IsCreated: true,
		},
	}
	testCases := []struct {
		name         string
		ctx          context.Context
		lessonIDs    []string
		expectedResp []*dto.Timesheet
		expectedErr  error
		setup        func(ctx context.Context)
	}{
		{
			name:         "get timesheet success",
			ctx:          ctx,
			lessonIDs:    []string{},
			expectedResp: timesheetDtos,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				timesheetRepo.On("FindTimesheetByLessonIDs", ctx, db, mock.Anything).Return(timesheetEntities, nil).Once()
				timesheetLessonHoursRepo.On("FindByTimesheetIDs", ctx, db, mock.Anything).Return(timesheetLessonHoursEntities, nil).Once()
				otherWorkingHoursRepo.On("FindListOtherWorkingHoursByTimesheetIDs", ctx, db, mock.Anything).Return(otherWorkingHoursEntities, nil).Once()
				transportationExpenseRepo.On("FindListTransportExpensesByTimesheetIDs", ctx, db, mock.Anything).Return(transportExpenseEntities, nil).Once()
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.GetTimesheetByLessonIDs(testCase.ctx, testCase.lessonIDs)
			assert.Equal(t, testCase.expectedErr, err)
			assert.True(t, common.CompareListTimesheet(testCase.expectedResp, resp))
		})
	}
}
