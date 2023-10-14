package nats

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_repository "github.com/manabie-com/backend/mock/timesheet/repository"
	mock_timesheet "github.com/manabie-com/backend/mock/timesheet/service/autocreatetimesheet"
	mock_auto_flag_services "github.com/manabie-com/backend/mock/timesheet/service/autocreatetimesheetflag"
	mock_get_timesheet "github.com/manabie-com/backend/mock/timesheet/service/gettimesheet"
	mock_confirmation_window_services "github.com/manabie-com/backend/mock/timesheet/service/timesheet_confirmation"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	v11 "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	_TimesheetStaffID              = "get_timesheet_staff_id"
	_TimesheetLocationID1          = "get_timesheet_location_id_1"
	_TimesheetLocationID2          = "get_timesheet_location_id_2"
	_TimesheetLessonHoursLessonID1 = "timesheet_lesson_hours_lesson_id_1"
	_TimesheetLessonHoursLessonID2 = "timesheet_lesson_hours_lesson_id_2"
	_TimesheetLessonHoursLessonID3 = "timesheet_lesson_hours_lesson_id_3"
	_TimesheetID1                  = "get_timesheet_timesheet_id_1"
	_TimesheetID2                  = "get_timesheet_timesheet_id_2"
)

var errInternal = errors.New("internal errors")
var publishActionLogError = fmt.Errorf("PublishActionLogTimesheetEvent JSM.PublishAsyncContext failed, msgID: %s, %w", "", errInternal)
var expectedLessonHourError = fmt.Errorf(
	"timesheet not contains lesson hours have timesheet id: %v, lesson id: %v",
	_TimesheetID2, _TimesheetLessonHoursLessonID2)

func TestLessonNatsService_HandleEventLessonUpdate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var (
		mockAutoCreateTimesheetService       = new(mock_timesheet.MockAutoCreateTimesheetServiceImpl)
		mockGetTimesheetService              = new(mock_get_timesheet.MockGetTimesheetServiceImpl)
		mockAutoCreateFlagActivityLogService = new(mock_auto_flag_services.MockAutoCreateTimesheetFlagServiceImpl)
		mockConfirmationWindowService        = new(mock_confirmation_window_services.MockConfirmationWindowServiceImpl)
		mockLessonRepo                       = new(mock_repository.MockLessonRepoImpl)
		mockPartnerAutoCreateRepo            = new(mock_repository.MockPartnerAutoCreateTimesheetFlagRepoImpl)

		mockJsm = new(mock_nats.JetStreamManagement)
	)

	lessonNatsService := LessonNatsServiceImpl{
		JSM:                              mockJsm,
		GetTimesheetService:              mockGetTimesheetService,
		AutoCreateTimesheetService:       mockAutoCreateTimesheetService,
		AutoCreateFlagActivityLogService: mockAutoCreateFlagActivityLogService,
		ConfirmationWindowService:        mockConfirmationWindowService,
		LessonRepo:                       mockLessonRepo,
		PartnerAutoCreateTimesheetRepo:   mockPartnerAutoCreateRepo,
	}
	lessonIDs := []string{"lessonID_1", "lessonID_2"}
	teacherIDs := []string{"teacherID_1", "teacherID_2", _TimesheetStaffID}
	now := time.Now()
	event := &bpb.EvtLesson_UpdateLesson{
		LessonId:               lessonIDs[0],
		StartAtBefore:          timestamppb.New(now),
		StartAtAfter:           timestamppb.New(now),
		EndAtBefore:            timestamppb.New(now),
		EndAtAfter:             timestamppb.New(now),
		TeacherIdsBefore:       teacherIDs,
		TeacherIdsAfter:        teacherIDs,
		SchedulingStatusBefore: v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_DRAFT,
		SchedulingStatusAfter:  v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
	}
	timesheetDtos := []*dto.Timesheet{
		{
			ID:                    _TimesheetID1,
			StaffID:               _TimesheetStaffID,
			LocationID:            _TimesheetLocationID1,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         now,
			ListOtherWorkingHours: nil,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: _TimesheetID1,
					LessonID:    _TimesheetLessonHoursLessonID1,
					IsCreated:   true,
				},
				{
					TimesheetID: _TimesheetID1,
					LessonID:    _TimesheetLessonHoursLessonID2,
					IsCreated:   true,
				},
			},
			IsCreated: true,
		},
		{
			ID:                    _TimesheetID2,
			StaffID:               _TimesheetStaffID,
			LocationID:            _TimesheetLocationID1,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         now,
			ListOtherWorkingHours: nil,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: _TimesheetID2,
					LessonID:    _TimesheetLessonHoursLessonID3,
					IsCreated:   true,
				},
			},
			IsCreated: true,
		},
	}
	autoCreateFlagActivityLogs := []*dto.AutoCreateFlagActivityLog{
		{
			StaffID:    _TimesheetStaffID,
			ChangeTime: now,
			FlagOn:     false,
		},
		{
			StaffID:    _TimesheetStaffID,
			ChangeTime: now,
			FlagOn:     true,
		},
	}
	partnerAutoCreateValue := entity.PartnerAutoCreateTimesheetFlag{
		FlagOn: database.Bool(true),
	}
	testCases := []struct {
		name        string
		ctx         context.Context
		request     interface{}
		expectedErr error
		setup       func(ctx context.Context)
	}{
		{
			name:        "HandleEventLessonUpdate success",
			ctx:         ctx,
			request:     event,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(timesheetDtos, nil).Once()
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(&partnerAutoCreateValue, nil).Once()
				mockAutoCreateTimesheetService.On("CreateTimesheetMultiple", ctx, mock.Anything).Return(timesheetDtos, nil).Once()
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Times(len(timesheetDtos)).Return("", nil)
			},
		},
		{
			name: "HandleEventLessonUpdate success with do nothing default case",
			ctx:  ctx,
			request: &bpb.EvtLesson_UpdateLesson{
				LessonId:               lessonIDs[0],
				StartAtBefore:          timestamppb.New(now),
				StartAtAfter:           timestamppb.New(now),
				EndAtBefore:            timestamppb.New(now),
				EndAtAfter:             timestamppb.New(now),
				TeacherIdsBefore:       teacherIDs,
				TeacherIdsAfter:        teacherIDs,
				SchedulingStatusBefore: v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
				SchedulingStatusAfter:  v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "HandleEventLessonUpdate failed GetTimesheet error",
			ctx:         ctx,
			request:     event,
			expectedErr: errInternal,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, errInternal).Once()

			},
		},
		{
			name:        "HandleEventLessonUpdate failed CreateTimesheetMultiple error",
			ctx:         ctx,
			request:     event,
			expectedErr: errInternal,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(timesheetDtos, nil).Once()
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(&partnerAutoCreateValue, nil).Once()
				mockAutoCreateTimesheetService.On("CreateTimesheetMultiple", ctx, mock.Anything).Return(nil, errInternal).Once()
			},
		},
		{
			name:        "HandleEventLessonUpdate failed when get modify condition",
			ctx:         ctx,
			request:     event,
			expectedErr: errInternal,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(false, errInternal)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(timesheetDtos, nil).Once()
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(&partnerAutoCreateValue, nil).Once()
				mockAutoCreateTimesheetService.On("CreateTimesheetMultiple", ctx, mock.Anything).Return(nil, errInternal).Once()
			},
		},
		{
			name:        "HandleEventLessonUpdate can not modify timesheet because confirm condition",
			ctx:         ctx,
			request:     event,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(false, nil)
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(false, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(timesheetDtos, nil).Once()
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(&partnerAutoCreateValue, nil).Once()
				mockAutoCreateTimesheetService.On("CreateTimesheetMultiple", ctx, mock.Anything).Return(nil, nil).Once()
			},
		},
		{
			name:        "HandleEventLessonUpdate failed to publish action log event",
			ctx:         ctx,
			request:     event,
			expectedErr: publishActionLogError,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(timesheetDtos, nil).Once()
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(&partnerAutoCreateValue, nil).Once()
				mockAutoCreateTimesheetService.On("CreateTimesheetMultiple", ctx, mock.Anything).Return(timesheetDtos, nil).Once()
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", errInternal)
			},
		},
		{
			name:        "HandleEventLessonCreate event change date failed get Partner Auto Create Default error",
			ctx:         ctx,
			request:     event,
			expectedErr: errInternal,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, nil).Once()
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(timesheetDtos, nil).Once()
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(nil, errInternal).Once()
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := lessonNatsService.HandleEventLessonUpdate(ctx, testCase.request.(*bpb.EvtLesson_UpdateLesson))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestLessonNatsService_HandleLessonCreate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var (
		mockAutoCreateTimesheetService       = new(mock_timesheet.MockAutoCreateTimesheetServiceImpl)
		mockGetTimesheetService              = new(mock_get_timesheet.MockGetTimesheetServiceImpl)
		mockAutoCreateFlagActivityLogService = new(mock_auto_flag_services.MockAutoCreateTimesheetFlagServiceImpl)
		mockConfirmationWindowService        = new(mock_confirmation_window_services.MockConfirmationWindowServiceImpl)
		mockLessonRepo                       = new(mock_repository.MockLessonRepoImpl)
		mockPartnerAutoCreateRepo            = new(mock_repository.MockPartnerAutoCreateTimesheetFlagRepoImpl)

		mockJsm = new(mock_nats.JetStreamManagement)
	)

	lessonNatsService := LessonNatsServiceImpl{
		JSM:                              mockJsm,
		GetTimesheetService:              mockGetTimesheetService,
		AutoCreateTimesheetService:       mockAutoCreateTimesheetService,
		AutoCreateFlagActivityLogService: mockAutoCreateFlagActivityLogService,
		ConfirmationWindowService:        mockConfirmationWindowService,
		LessonRepo:                       mockLessonRepo,
		PartnerAutoCreateTimesheetRepo:   mockPartnerAutoCreateRepo,
	}

	teacherIDs := []string{"teacherID_1", "teacherID_2", _TimesheetStaffID}
	now := time.Now()
	event := &bpb.EvtLesson_CreateLessons{
		Lessons: []*bpb.EvtLesson_Lesson{
			{
				LessonId:         _TimesheetLessonHoursLessonID1,
				TeacherIds:       teacherIDs,
				LocationId:       _TimesheetLocationID1,
				StartAt:          timestamppb.New(now),
				EndAt:            timestamppb.New(now),
				SchedulingStatus: v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
			},
			{
				LessonId:         _TimesheetLessonHoursLessonID2,
				TeacherIds:       teacherIDs,
				LocationId:       _TimesheetLocationID1,
				StartAt:          timestamppb.New(now),
				EndAt:            timestamppb.New(now),
				SchedulingStatus: v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
			},
			{
				LessonId:         _TimesheetLessonHoursLessonID2,
				TeacherIds:       teacherIDs,
				LocationId:       _TimesheetLocationID1,
				StartAt:          timestamppb.New(now),
				EndAt:            timestamppb.New(now),
				SchedulingStatus: v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED,
			},
		},
	}
	timesheetDtos := []*dto.Timesheet{
		{
			ID:                    _TimesheetID1,
			StaffID:               teacherIDs[0],
			LocationID:            _TimesheetLocationID1,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         now,
			ListOtherWorkingHours: nil,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: _TimesheetID1,
					//LessonID:    _TimesheetLessonHoursLessonID1,
					LessonID:  idutil.ULIDNow(),
					IsCreated: true,
				},
				{
					TimesheetID: _TimesheetID1,
					//LessonID:    _TimesheetLessonHoursLessonID2,
					LessonID:  idutil.ULIDNow(),
					IsCreated: true,
				},
			},
			IsCreated: true,
		},
		{
			ID:                    _TimesheetID2,
			StaffID:               teacherIDs[1],
			LocationID:            _TimesheetLocationID1,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         now,
			ListOtherWorkingHours: nil,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: _TimesheetID2,
					LessonID:    _TimesheetLessonHoursLessonID3,
					IsCreated:   true,
				},
			},
			IsCreated: true,
		},
	}
	autoCreateFlagActivityLogs := []*dto.AutoCreateFlagActivityLog{
		{
			StaffID:    _TimesheetStaffID,
			ChangeTime: now,
			FlagOn:     false,
		},
		{
			StaffID:    _TimesheetStaffID,
			ChangeTime: now,
			FlagOn:     true,
		},
	}
	partnerAutoCreateValue := entity.PartnerAutoCreateTimesheetFlag{
		FlagOn: database.Bool(true),
	}
	testCases := []struct {
		name        string
		ctx         context.Context
		request     interface{}
		expectedErr error
		setup       func(ctx context.Context)
	}{
		{
			name:        "HandleEventLessonCreate success",
			ctx:         ctx,
			request:     event,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(timesheetDtos, nil)
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(&partnerAutoCreateValue, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(&partnerAutoCreateValue, nil).Once()
				mockAutoCreateTimesheetService.On("CreateTimesheetMultiple", ctx, mock.Anything).Return(timesheetDtos, nil).Once()
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Times(len(timesheetDtos)).Return("", nil)
			},
		},
		{
			name:        "HandleEventLessonCreate failed CreateTimesheetMultiple error",
			ctx:         ctx,
			request:     event,
			expectedErr: errInternal,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(timesheetDtos, nil)
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(&partnerAutoCreateValue, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(&partnerAutoCreateValue, nil).Once()
				mockAutoCreateTimesheetService.On("CreateTimesheetMultiple", ctx, mock.Anything).Return(nil, errInternal).Once()
			},
		},
		{
			name:        "HandleEventLessonCreate get confirmation condition error",
			ctx:         ctx,
			request:     event,
			expectedErr: errInternal,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, errInternal)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(timesheetDtos, nil)
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(&partnerAutoCreateValue, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(&partnerAutoCreateValue, nil).Once()
				mockAutoCreateTimesheetService.On("CreateTimesheetMultiple", ctx, mock.Anything).Return(nil, errInternal).Once()
			},
		},
		{
			name:        "HandleEventLessonCreate event change date failed get Partner Auto Create Default error",
			ctx:         ctx,
			request:     event,
			expectedErr: errInternal,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, nil).Once()
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(timesheetDtos, nil).Once()
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(nil, errInternal).Once()
			},
		},
		{
			name:        "HandleEventLessonCreate failed to publish timesheet action log event",
			ctx:         ctx,
			request:     event,
			expectedErr: publishActionLogError,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(timesheetDtos, nil)
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(&partnerAutoCreateValue, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(&partnerAutoCreateValue, nil).Once()
				mockAutoCreateTimesheetService.On("CreateTimesheetMultiple", ctx, mock.Anything).Return(timesheetDtos, nil).Once()
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", errInternal)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := lessonNatsService.HandleEventLessonCreate(ctx, testCase.request.(*bpb.EvtLesson_CreateLessons))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestLessonNatsService_HandleEventLessonUpdate_PublishedToDraft(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var (
		mockAutoCreateTimesheetService       = new(mock_timesheet.MockAutoCreateTimesheetServiceImpl)
		mockGetTimesheetService              = new(mock_get_timesheet.MockGetTimesheetServiceImpl)
		mockAutoCreateFlagActivityLogService = new(mock_auto_flag_services.MockAutoCreateTimesheetFlagServiceImpl)
		mockConfirmationWindowService        = new(mock_confirmation_window_services.MockConfirmationWindowServiceImpl)
		mockLessonRepo                       = new(mock_repository.MockLessonRepoImpl)

		mockJsm = new(mock_nats.JetStreamManagement)
	)

	lessonNatsService := LessonNatsServiceImpl{
		JSM:                              mockJsm,
		GetTimesheetService:              mockGetTimesheetService,
		AutoCreateTimesheetService:       mockAutoCreateTimesheetService,
		AutoCreateFlagActivityLogService: mockAutoCreateFlagActivityLogService,
		ConfirmationWindowService:        mockConfirmationWindowService,
		LessonRepo:                       mockLessonRepo,
	}
	lessonIDs := []string{_TimesheetLessonHoursLessonID1, _TimesheetLessonHoursLessonID2}
	teacherIDs := []string{"teacherID_1", "teacherID_2", _TimesheetStaffID}
	now := time.Now()

	eventPublishedToDraft1 := &bpb.EvtLesson_UpdateLesson{
		LessonId:               lessonIDs[0],
		StartAtBefore:          timestamppb.New(now),
		StartAtAfter:           timestamppb.New(now),
		EndAtBefore:            timestamppb.New(now),
		EndAtAfter:             timestamppb.New(now),
		TeacherIdsBefore:       teacherIDs,
		TeacherIdsAfter:        teacherIDs,
		SchedulingStatusBefore: v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
		SchedulingStatusAfter:  v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_DRAFT,
	}
	eventPublishedToDraft2 := &bpb.EvtLesson_UpdateLesson{
		LessonId:               lessonIDs[1],
		StartAtBefore:          timestamppb.New(now),
		StartAtAfter:           timestamppb.New(now),
		EndAtBefore:            timestamppb.New(now),
		EndAtAfter:             timestamppb.New(now),
		TeacherIdsBefore:       teacherIDs,
		TeacherIdsAfter:        teacherIDs,
		SchedulingStatusBefore: v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
		SchedulingStatusAfter:  v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_DRAFT,
	}
	timesheetDtos := []*dto.Timesheet{
		{
			ID:                    _TimesheetID1,
			StaffID:               _TimesheetStaffID,
			LocationID:            _TimesheetLocationID1,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         now,
			ListOtherWorkingHours: nil,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: _TimesheetID1,
					LessonID:    lessonIDs[0],
					IsCreated:   true,
				},
				{
					TimesheetID: _TimesheetID1,
					LessonID:    _TimesheetLessonHoursLessonID2,
					IsCreated:   true,
				},
			},
			IsCreated: true,
		},
		{
			ID:                    _TimesheetID2,
			StaffID:               _TimesheetStaffID,
			LocationID:            _TimesheetLocationID1,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         now,
			ListOtherWorkingHours: nil,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: _TimesheetID2,
					LessonID:    lessonIDs[0],
					IsCreated:   true,
				},
			},
			IsCreated: true,
		},
	}
	testCases := []struct {
		name        string
		ctx         context.Context
		request     interface{}
		expectedErr error
		setup       func(ctx context.Context)
	}{
		{
			name:        "HandleEventLessonUpdate event published to draft success",
			ctx:         ctx,
			request:     eventPublishedToDraft1,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(timesheetDtos, nil).Once()
				mockAutoCreateTimesheetService.On("RemoveTimesheetLessonHoursMultiple", ctx, mock.Anything).Return(timesheetDtos, nil).Once()
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Times(len(timesheetDtos)).Return("", nil)
			},
		},
		{
			name:        "HandleEventLessonUpdate event published to draft failed timesheet lesson hours not contains lessonID",
			ctx:         ctx,
			request:     eventPublishedToDraft2,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(timesheetDtos, nil).Once()
				mockAutoCreateTimesheetService.On("RemoveTimesheetLessonHoursMultiple", ctx, mock.Anything).Return(timesheetDtos, nil).Once()
			},
		},
		{
			name:        "HandleEventLessonUpdate event published to draft failed GetTimesheet error",
			ctx:         ctx,
			request:     eventPublishedToDraft1,
			expectedErr: errInternal,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, errInternal).Once()
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := lessonNatsService.HandleEventLessonUpdate(ctx, testCase.request.(*bpb.EvtLesson_UpdateLesson))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}

}

func TestLessonNatsService_HandleEventLessonUpdate_lessonUpdateChangeTeachers(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var (
		mockAutoCreateTimesheetService       = new(mock_timesheet.MockAutoCreateTimesheetServiceImpl)
		mockGetTimesheetService              = new(mock_get_timesheet.MockGetTimesheetServiceImpl)
		mockAutoCreateFlagActivityLogService = new(mock_auto_flag_services.MockAutoCreateTimesheetFlagServiceImpl)
		mockConfirmationWindowService        = new(mock_confirmation_window_services.MockConfirmationWindowServiceImpl)
		mockLessonRepo                       = new(mock_repository.MockLessonRepoImpl)
		mockPartnerAutoCreateRepo            = new(mock_repository.MockPartnerAutoCreateTimesheetFlagRepoImpl)
	)

	lessonNatsService := LessonNatsServiceImpl{
		GetTimesheetService:              mockGetTimesheetService,
		AutoCreateTimesheetService:       mockAutoCreateTimesheetService,
		AutoCreateFlagActivityLogService: mockAutoCreateFlagActivityLogService,
		ConfirmationWindowService:        mockConfirmationWindowService,
		LessonRepo:                       mockLessonRepo,
		PartnerAutoCreateTimesheetRepo:   mockPartnerAutoCreateRepo,
	}
	lessonIDs := []string{_TimesheetLessonHoursLessonID1, _TimesheetLessonHoursLessonID2}
	teacherIDsBef := []string{"teacherID_1", "teacherID_2", _TimesheetStaffID}
	teacherIDsAft := []string{"teacherID_2", "teacherID_3", _TimesheetStaffID}
	now := time.Now()

	eventChangeTeachers1 := &bpb.EvtLesson_UpdateLesson{
		LessonId:               lessonIDs[0],
		StartAtBefore:          timestamppb.New(now),
		StartAtAfter:           timestamppb.New(now),
		EndAtBefore:            timestamppb.New(now),
		EndAtAfter:             timestamppb.New(now),
		TeacherIdsBefore:       teacherIDsBef,
		TeacherIdsAfter:        teacherIDsAft,
		SchedulingStatusBefore: v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
		SchedulingStatusAfter:  v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
	}
	eventChangeTeachers2 := &bpb.EvtLesson_UpdateLesson{
		LessonId:               lessonIDs[1],
		StartAtBefore:          timestamppb.New(now),
		StartAtAfter:           timestamppb.New(now),
		EndAtBefore:            timestamppb.New(now),
		EndAtAfter:             timestamppb.New(now),
		TeacherIdsBefore:       teacherIDsBef,
		TeacherIdsAfter:        teacherIDsAft,
		SchedulingStatusBefore: v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
		SchedulingStatusAfter:  v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
	}
	timesheetDtos := []*dto.Timesheet{
		{
			ID:                    _TimesheetID1,
			StaffID:               _TimesheetStaffID,
			LocationID:            _TimesheetLocationID1,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         now,
			ListOtherWorkingHours: nil,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: _TimesheetID1,
					LessonID:    lessonIDs[0],
					IsCreated:   true,
				},
				{
					TimesheetID: _TimesheetID1,
					LessonID:    _TimesheetLessonHoursLessonID2,
					IsCreated:   true,
				},
			},
			IsCreated: true,
		},
		{
			ID:                    _TimesheetID2,
			StaffID:               _TimesheetStaffID,
			LocationID:            _TimesheetLocationID1,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         now,
			ListOtherWorkingHours: nil,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: _TimesheetID2,
					LessonID:    lessonIDs[0],
					IsCreated:   true,
				},
			},
			IsCreated: true,
		},
	}
	autoCreateFlagActivityLogs := []*dto.AutoCreateFlagActivityLog{
		{
			StaffID:    _TimesheetStaffID,
			ChangeTime: now,
			FlagOn:     false,
		},
		{
			StaffID:    _TimesheetStaffID,
			ChangeTime: now,
			FlagOn:     true,
		},
	}
	partnerAutoCreateValue := entity.PartnerAutoCreateTimesheetFlag{
		FlagOn: database.Bool(true),
	}
	testCases := []struct {
		name        string
		ctx         context.Context
		request     interface{}
		expectedErr error
		setup       func(ctx context.Context)
	}{
		{
			name:        "HandleEventLessonUpdate event change teachers success",
			ctx:         ctx,
			request:     eventChangeTeachers1,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, nil).Once()
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(timesheetDtos, nil).Once()
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(&partnerAutoCreateValue, nil).Once()
				mockAutoCreateTimesheetService.On("CreateAndRemoveTimesheetMultiple", ctx, mock.Anything, mock.Anything).Return(nil).Once()
			},
		},
		{
			name:        "HandleEventLessonUpdate event change teachers failed GetTimesheet new error",
			ctx:         ctx,
			request:     eventChangeTeachers1,
			expectedErr: errInternal,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, errInternal).Once()
			},
		},
		{
			name:        "HandleEventLessonUpdate event change teachers failed GetTimesheet removed error",
			ctx:         ctx,
			request:     eventChangeTeachers1,
			expectedErr: errInternal,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, nil).Once()
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, errInternal).Once()
			},
		},
		{
			name:        "HandleEventLessonUpdate event change teachers failed timesheet lesson hours not contains lessonID",
			ctx:         ctx,
			request:     eventChangeTeachers2,
			expectedErr: expectedLessonHourError,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, expectedLessonHourError).Once()
			},
		},
		{
			name:        "HandleEventLessonUpdate event change teachers failed create and update timesheet error",
			ctx:         ctx,
			request:     eventChangeTeachers1,
			expectedErr: errInternal,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, nil).Once()
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(timesheetDtos, nil).Once()
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(&partnerAutoCreateValue, nil).Once()
				mockAutoCreateTimesheetService.On("CreateAndRemoveTimesheetMultiple", ctx, mock.Anything, mock.Anything).Return(errInternal).Once()
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := lessonNatsService.HandleEventLessonUpdate(ctx, testCase.request.(*bpb.EvtLesson_UpdateLesson))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}

}

func TestLessonNatsService_HandleEventLessonUpdate_lessonUpdateChangeLessonDate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var (
		mockAutoCreateTimesheetService       = new(mock_timesheet.MockAutoCreateTimesheetServiceImpl)
		mockGetTimesheetService              = new(mock_get_timesheet.MockGetTimesheetServiceImpl)
		mockAutoCreateFlagActivityLogService = new(mock_auto_flag_services.MockAutoCreateTimesheetFlagServiceImpl)
		mockConfirmationWindowService        = new(mock_confirmation_window_services.MockConfirmationWindowServiceImpl)
		mockLessonRepo                       = new(mock_repository.MockLessonRepoImpl)
		mockPartnerAutoCreateRepo            = new(mock_repository.MockPartnerAutoCreateTimesheetFlagRepoImpl)
		mockJsm                              = new(mock_nats.JetStreamManagement)
	)

	lessonNatsService := LessonNatsServiceImpl{
		JSM:                              mockJsm,
		GetTimesheetService:              mockGetTimesheetService,
		AutoCreateTimesheetService:       mockAutoCreateTimesheetService,
		AutoCreateFlagActivityLogService: mockAutoCreateFlagActivityLogService,
		ConfirmationWindowService:        mockConfirmationWindowService,
		LessonRepo:                       mockLessonRepo,
		PartnerAutoCreateTimesheetRepo:   mockPartnerAutoCreateRepo,
	}
	lessonIDs := []string{_TimesheetLessonHoursLessonID1, _TimesheetLessonHoursLessonID2}
	teacherIDs := []string{"teacherID_1", "teacherID_2", _TimesheetStaffID}

	now := time.Now()
	dateBef := time.Now()
	dateAft := dateBef.AddDate(0, 0, 2)
	eventChangeDate1 := &bpb.EvtLesson_UpdateLesson{
		LessonId:               lessonIDs[0],
		StartAtBefore:          timestamppb.New(dateBef),
		StartAtAfter:           timestamppb.New(dateAft),
		EndAtBefore:            timestamppb.New(dateBef),
		EndAtAfter:             timestamppb.New(dateAft),
		TeacherIdsBefore:       teacherIDs,
		TeacherIdsAfter:        teacherIDs,
		SchedulingStatusBefore: v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
		SchedulingStatusAfter:  v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
	}
	eventChangeDate2 := &bpb.EvtLesson_UpdateLesson{
		LessonId:               lessonIDs[1],
		StartAtBefore:          timestamppb.New(dateBef),
		StartAtAfter:           timestamppb.New(dateAft),
		EndAtBefore:            timestamppb.New(dateBef),
		EndAtAfter:             timestamppb.New(dateAft),
		TeacherIdsBefore:       teacherIDs,
		TeacherIdsAfter:        teacherIDs,
		SchedulingStatusBefore: v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
		SchedulingStatusAfter:  v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
	}
	timesheetDtos := []*dto.Timesheet{
		{
			ID:                    _TimesheetID1,
			StaffID:               _TimesheetStaffID,
			LocationID:            _TimesheetLocationID1,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         now,
			ListOtherWorkingHours: nil,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: _TimesheetID1,
					LessonID:    lessonIDs[0],
					IsCreated:   true,
				},
				{
					TimesheetID: _TimesheetID1,
					LessonID:    _TimesheetLessonHoursLessonID1,
					IsCreated:   true,
				},
			},
			IsCreated: true,
		},
		{
			ID:                    _TimesheetID2,
			StaffID:               _TimesheetStaffID,
			LocationID:            _TimesheetLocationID1,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         now,
			ListOtherWorkingHours: nil,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: _TimesheetID2,
					LessonID:    lessonIDs[0],
					IsCreated:   true,
				},
			},
			IsCreated: true,
		},
	}
	autoCreateFlagActivityLogs := []*dto.AutoCreateFlagActivityLog{
		{
			StaffID:    _TimesheetStaffID,
			ChangeTime: now,
			FlagOn:     false,
		},
		{
			StaffID:    _TimesheetStaffID,
			ChangeTime: now,
			FlagOn:     true,
		},
	}
	partnerAutoCreateValue := entity.PartnerAutoCreateTimesheetFlag{
		FlagOn: database.Bool(true),
	}
	testCases := []struct {
		name        string
		ctx         context.Context
		request     interface{}
		expectedErr error
		setup       func(ctx context.Context)
	}{
		{
			name:        "HandleEventLessonUpdate event change date success",
			ctx:         ctx,
			request:     eventChangeDate1,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, nil).Once()
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(timesheetDtos, nil).Once()
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(&partnerAutoCreateValue, nil).Once()
				mockAutoCreateTimesheetService.On("CreateAndRemoveTimesheetMultiple", ctx, mock.Anything, mock.Anything).Return(nil).Once()
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Times(len(timesheetDtos)).Return("", nil)
			},
		},
		{
			name:        "HandleEventLessonUpdate event change date failed GetTimesheet new error",
			ctx:         ctx,
			request:     eventChangeDate1,
			expectedErr: errInternal,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, errInternal).Once()
			},
		},
		{
			name:        "HandleEventLessonUpdate event change date failed GetTimesheet removed error",
			ctx:         ctx,
			request:     eventChangeDate1,
			expectedErr: errInternal,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, nil).Once()
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, errInternal).Once()
			},
		},
		{
			name:        "HandleEventLessonUpdate event change date failed timesheet lesson hours not contains lessonID",
			ctx:         ctx,
			request:     eventChangeDate2,
			expectedErr: expectedLessonHourError,
			setup: func(ctx context.Context) {
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, expectedLessonHourError).Once()
			},
		},
		{
			name:        "HandleEventLessonUpdate event change date failed create and update timesheet error",
			ctx:         ctx,
			request:     eventChangeDate1,
			expectedErr: errInternal,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, nil).Once()
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(timesheetDtos, nil).Once()
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(&partnerAutoCreateValue, nil).Once()
				mockAutoCreateTimesheetService.On("CreateAndRemoveTimesheetMultiple", ctx, mock.Anything, mock.Anything).Return(errInternal).Once()
			},
		},
		{
			name:        "HandleEventLessonCreate event change date failed get Partner Auto Create Default error",
			ctx:         ctx,
			request:     eventChangeDate1,
			expectedErr: errInternal,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, nil).Once()
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(timesheetDtos, nil).Once()
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(nil, errInternal).Once()
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := lessonNatsService.HandleEventLessonUpdate(ctx, testCase.request.(*bpb.EvtLesson_UpdateLesson))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}

}

func TestLessonNatsService_HandleEventLessonUpdate_lessonUpdateChangeLocationID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var (
		mockAutoCreateTimesheetService       = new(mock_timesheet.MockAutoCreateTimesheetServiceImpl)
		mockGetTimesheetService              = new(mock_get_timesheet.MockGetTimesheetServiceImpl)
		mockAutoCreateFlagActivityLogService = new(mock_auto_flag_services.MockAutoCreateTimesheetFlagServiceImpl)
		mockConfirmationWindowService        = new(mock_confirmation_window_services.MockConfirmationWindowServiceImpl)
		mockLessonRepo                       = new(mock_repository.MockLessonRepoImpl)
		mockPartnerAutoCreateRepo            = new(mock_repository.MockPartnerAutoCreateTimesheetFlagRepoImpl)

		mockJsm = new(mock_nats.JetStreamManagement)
	)

	lessonNatsService := LessonNatsServiceImpl{
		JSM:                              mockJsm,
		GetTimesheetService:              mockGetTimesheetService,
		AutoCreateTimesheetService:       mockAutoCreateTimesheetService,
		AutoCreateFlagActivityLogService: mockAutoCreateFlagActivityLogService,
		ConfirmationWindowService:        mockConfirmationWindowService,
		LessonRepo:                       mockLessonRepo,
		PartnerAutoCreateTimesheetRepo:   mockPartnerAutoCreateRepo,
	}
	lessonIDs := []string{_TimesheetLessonHoursLessonID1, _TimesheetLessonHoursLessonID2}
	teacherIDs := []string{"teacherID_1", "teacherID_2", _TimesheetStaffID}

	now := time.Now()

	eventChangeLocation1 := &bpb.EvtLesson_UpdateLesson{
		LessonId:               lessonIDs[0],
		StartAtBefore:          timestamppb.New(now),
		StartAtAfter:           timestamppb.New(now),
		EndAtBefore:            timestamppb.New(now),
		EndAtAfter:             timestamppb.New(now),
		TeacherIdsBefore:       teacherIDs,
		TeacherIdsAfter:        teacherIDs,
		LocationIdBefore:       _TimesheetLocationID1,
		LocationIdAfter:        _TimesheetLocationID2,
		SchedulingStatusBefore: v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
		SchedulingStatusAfter:  v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
	}
	eventChangeLocation2 := &bpb.EvtLesson_UpdateLesson{
		LessonId:               lessonIDs[1],
		StartAtBefore:          timestamppb.New(now),
		StartAtAfter:           timestamppb.New(now),
		EndAtBefore:            timestamppb.New(now),
		EndAtAfter:             timestamppb.New(now),
		TeacherIdsBefore:       teacherIDs,
		TeacherIdsAfter:        teacherIDs,
		LocationIdBefore:       _TimesheetLocationID1,
		LocationIdAfter:        _TimesheetLocationID2,
		SchedulingStatusBefore: v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
		SchedulingStatusAfter:  v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
	}
	timesheetDtos := []*dto.Timesheet{
		{
			ID:                    _TimesheetID1,
			StaffID:               _TimesheetStaffID,
			LocationID:            _TimesheetLocationID1,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         now,
			ListOtherWorkingHours: nil,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: _TimesheetID1,
					LessonID:    lessonIDs[0],
					IsCreated:   true,
				},
				{
					TimesheetID: _TimesheetID1,
					LessonID:    _TimesheetLessonHoursLessonID2,
					IsCreated:   true,
				},
			},
			IsCreated: true,
		},
		{
			ID:                    _TimesheetID2,
			StaffID:               _TimesheetStaffID,
			LocationID:            _TimesheetLocationID1,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         now,
			ListOtherWorkingHours: nil,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: _TimesheetID2,
					LessonID:    lessonIDs[0],
					IsCreated:   true,
				},
			},
			IsCreated: true,
		},
	}
	autoCreateFlagActivityLogs := []*dto.AutoCreateFlagActivityLog{
		{
			StaffID:    _TimesheetStaffID,
			ChangeTime: now,
			FlagOn:     false,
		},
		{
			StaffID:    _TimesheetStaffID,
			ChangeTime: now,
			FlagOn:     true,
		},
	}
	partnerAutoCreateValue := entity.PartnerAutoCreateTimesheetFlag{
		FlagOn: database.Bool(true),
	}
	testCases := []struct {
		name        string
		ctx         context.Context
		request     interface{}
		expectedErr error
		setup       func(ctx context.Context)
	}{
		{
			name:        "HandleEventLessonUpdate event change location success",
			ctx:         ctx,
			request:     eventChangeLocation1,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, nil).Once()
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(timesheetDtos, nil).Once()
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(&partnerAutoCreateValue, nil).Once()
				mockAutoCreateTimesheetService.On("CreateAndRemoveTimesheetMultiple", ctx, mock.Anything, mock.Anything).Return(nil).Once()
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Times(len(timesheetDtos)).Return("", nil)
			},
		},
		{
			name:        "HandleEventLessonUpdate event change location failed GetTimesheet new error",
			ctx:         ctx,
			request:     eventChangeLocation1,
			expectedErr: errInternal,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, errInternal).Once()
			},
		},
		{
			name:        "HandleEventLessonUpdate event change location failed GetTimesheet removed error",
			ctx:         ctx,
			request:     eventChangeLocation1,
			expectedErr: errInternal,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, nil).Once()
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, errInternal).Once()
			},
		},
		{
			name:        "HandleEventLessonUpdate event change location failed timesheet lesson hours not contains lessonID",
			ctx:         ctx,
			request:     eventChangeLocation2,
			expectedErr: expectedLessonHourError,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, expectedLessonHourError).Once()
			},
		},
		{
			name:        "HandleEventLessonUpdate event change location failed create and update timesheet error",
			ctx:         ctx,
			request:     eventChangeLocation1,
			expectedErr: errInternal,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, nil).Once()
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(timesheetDtos, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(&partnerAutoCreateValue, nil).Once()
				mockAutoCreateTimesheetService.On("CreateAndRemoveTimesheetMultiple", ctx, mock.Anything, mock.Anything).Return(errInternal).Once()
			},
		},
		{
			name:        "HandleEventLessonCreate event change date failed get Partner Auto Create Default error",
			ctx:         ctx,
			request:     eventChangeLocation1,
			expectedErr: errInternal,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, nil).Once()
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(timesheetDtos, nil).Once()
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(nil, errInternal).Once()
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := lessonNatsService.HandleEventLessonUpdate(ctx, testCase.request.(*bpb.EvtLesson_UpdateLesson))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}

}

func TestLessonNatsService_HandleEventLessonDelete(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var (
		mockAutoCreateTimesheetService       = new(mock_timesheet.MockAutoCreateTimesheetServiceImpl)
		mockGetTimesheetService              = new(mock_get_timesheet.MockGetTimesheetServiceImpl)
		mockAutoCreateFlagActivityLogService = new(mock_auto_flag_services.MockAutoCreateTimesheetFlagServiceImpl)
		mockConfirmationWindowService        = new(mock_confirmation_window_services.MockConfirmationWindowServiceImpl)
		mockLessonRepo                       = new(mock_repository.MockLessonRepoImpl)
	)

	lessonNatsService := LessonNatsServiceImpl{
		GetTimesheetService:              mockGetTimesheetService,
		AutoCreateTimesheetService:       mockAutoCreateTimesheetService,
		AutoCreateFlagActivityLogService: mockAutoCreateFlagActivityLogService,
		ConfirmationWindowService:        mockConfirmationWindowService,
		LessonRepo:                       mockLessonRepo,
	}
	lessonIDs := []string{_TimesheetLessonHoursLessonID1, _TimesheetLessonHoursLessonID2, _TimesheetLessonHoursLessonID3}
	lessons := []*entity.Lesson{}
	now := time.Now()
	eventTimesheetContainDeletedLessonID := &bpb.EvtLesson_DeletedLessons{
		LessonIds: lessonIDs[:1],
	}
	eventTimesheetNotContainDeletedLessonID := &bpb.EvtLesson_DeletedLessons{
		LessonIds: lessonIDs[1:3],
	}
	timesheetDtos := []*dto.Timesheet{
		{
			ID:                    _TimesheetID1,
			StaffID:               _TimesheetStaffID,
			LocationID:            _TimesheetLocationID1,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         now,
			ListOtherWorkingHours: nil,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: _TimesheetID1,
					LessonID:    _TimesheetLessonHoursLessonID1,
					IsCreated:   true,
				},
			},
			IsCreated: true,
		},
	}
	testCases := []struct {
		name        string
		ctx         context.Context
		request     interface{}
		expectedErr error
		setup       func(ctx context.Context)
	}{
		{
			name:        "HandleEventLessonDelete success",
			ctx:         ctx,
			request:     eventTimesheetContainDeletedLessonID,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockLessonRepo.On("FindAllLessonsByIDsIgnoreDeletedAtCondition", ctx, mock.Anything, mock.Anything).Return(lessons, nil)
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheetByLessonIDs", ctx, mock.Anything).Return(timesheetDtos, nil).Once()
				mockAutoCreateTimesheetService.On("RemoveTimesheetLessonHoursMultiple", ctx, mock.Anything).Return(timesheetDtos, nil).Once()
			},
		},
		{
			name:        "HandleEventLessonDelete timesheet lesson hours not contains lessonID got failed",
			ctx:         ctx,
			request:     eventTimesheetNotContainDeletedLessonID,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockLessonRepo.On("FindAllLessonsByIDsIgnoreDeletedAtCondition", ctx, mock.Anything, mock.Anything).Return(lessons, nil)
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheetByLessonIDs", ctx, mock.Anything).Return(timesheetDtos, nil).Once()
				mockAutoCreateTimesheetService.On("RemoveTimesheetLessonHoursMultiple", ctx, mock.Anything).Return(timesheetDtos, nil).Twice()
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := lessonNatsService.HandleEventLessonDelete(ctx, testCase.request.(*bpb.EvtLesson_DeletedLessons))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestLessonNatsService_HandleEventLessonUpdate_lessonUpdateChangeLessonStartTime(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var (
		mockAutoCreateTimesheetService       = new(mock_timesheet.MockAutoCreateTimesheetServiceImpl)
		mockGetTimesheetService              = new(mock_get_timesheet.MockGetTimesheetServiceImpl)
		mockAutoCreateFlagActivityLogService = new(mock_auto_flag_services.MockAutoCreateTimesheetFlagServiceImpl)
		mockConfirmationWindowService        = new(mock_confirmation_window_services.MockConfirmationWindowServiceImpl)
		mockLessonRepo                       = new(mock_repository.MockLessonRepoImpl)
		mockPartnerAutoCreateRepo            = new(mock_repository.MockPartnerAutoCreateTimesheetFlagRepoImpl)
	)

	lessonNatsService := LessonNatsServiceImpl{
		GetTimesheetService:              mockGetTimesheetService,
		AutoCreateTimesheetService:       mockAutoCreateTimesheetService,
		AutoCreateFlagActivityLogService: mockAutoCreateFlagActivityLogService,
		ConfirmationWindowService:        mockConfirmationWindowService,
		LessonRepo:                       mockLessonRepo,
		PartnerAutoCreateTimesheetRepo:   mockPartnerAutoCreateRepo,
	}
	lessonIDs := []string{_TimesheetLessonHoursLessonID1, _TimesheetLessonHoursLessonID2}
	teacherIDs := []string{"teacherID_1", "teacherID_2", _TimesheetStaffID}

	timeToRunTest := time.Date(2022, 5, 21, 10, 00, 00, 0, time.UTC)
	dateBef := timeToRunTest.Add(-time.Hour)
	dateAft := timeToRunTest.Add(time.Hour)
	eventChangeDate1 := &bpb.EvtLesson_UpdateLesson{
		LessonId:               lessonIDs[0],
		StartAtBefore:          timestamppb.New(dateBef),
		EndAtBefore:            timestamppb.New(dateBef.Add(15 * time.Minute)),
		StartAtAfter:           timestamppb.New(dateAft),
		EndAtAfter:             timestamppb.New(dateAft.Add(15 * time.Minute)),
		TeacherIdsBefore:       teacherIDs,
		TeacherIdsAfter:        teacherIDs,
		SchedulingStatusBefore: v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
		SchedulingStatusAfter:  v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
	}

	autoCreateFlagActivityLogs := []*dto.AutoCreateFlagActivityLog{
		{
			StaffID:    _TimesheetStaffID,
			ChangeTime: timeToRunTest,
			FlagOn:     false,
		},
		{
			StaffID:    _TimesheetStaffID,
			ChangeTime: timeToRunTest,
			FlagOn:     true,
		},
	}
	partnerAutoCreateValue := entity.PartnerAutoCreateTimesheetFlag{
		FlagOn: database.Bool(true),
	}
	testCases := []struct {
		name        string
		ctx         context.Context
		request     interface{}
		expectedErr error
		setup       func(ctx context.Context)
	}{
		{
			name:        "happy case lessonUpdateChangeLessonStartTime event change start time success",
			ctx:         ctx,
			request:     eventChangeDate1,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, nil).Once()
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(&partnerAutoCreateValue, nil).Once()
				mockAutoCreateTimesheetService.On("UpdateLessonAutoCreateFlagState", ctx, mock.Anything).Return(nil).Once()
			},
		},
		{
			name:        "error case lessonUpdateChangeLessonStartTime change start time GetTimesheet new error",
			ctx:         ctx,
			request:     eventChangeDate1,
			expectedErr: errInternal,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, errInternal).Once()
			},
		},
		{
			name:        "error case lessonUpdateChangeLessonStartTime update flag fail",
			ctx:         ctx,
			request:     eventChangeDate1,
			expectedErr: errInternal,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, nil).Once()
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(&partnerAutoCreateValue, nil).Once()
				mockAutoCreateTimesheetService.On("UpdateLessonAutoCreateFlagState", ctx, mock.Anything).Return(errInternal).Once()
			},
		},
		{
			name:        "error mappingStaffAndAutoCreateFlagActivityLog get auto create flag log by teacher fail",
			ctx:         ctx,
			request:     eventChangeDate1,
			expectedErr: errInternal,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, nil).Once()
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(nil, errInternal).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(&partnerAutoCreateValue, nil).Once()
			},
		},
		{
			name:        "HandleEventLessonCreate event change date failed get Partner Auto Create Default error",
			ctx:         ctx,
			request:     eventChangeDate1,
			expectedErr: errInternal,
			setup: func(ctx context.Context) {
				mockConfirmationWindowService.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).Return(true, nil)
				mockGetTimesheetService.On("GetTimesheet", ctx, mock.Anything, mock.Anything).Return(nil, nil).Once()
				mockAutoCreateFlagActivityLogService.On("GetAutoCreateFlagLogByTeacherIDs", ctx, mock.Anything, mock.Anything).Return(autoCreateFlagActivityLogs, nil).Once()
				mockPartnerAutoCreateRepo.On("GetPartnerAutoCreateDefaultValue", ctx, mock.Anything, mock.Anything).Return(nil, errInternal).Once()
				mockAutoCreateTimesheetService.On("UpdateLessonAutoCreateFlagState", ctx, mock.Anything).Return(errInternal).Once()
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := lessonNatsService.HandleEventLessonUpdate(ctx, testCase.request.(*bpb.EvtLesson_UpdateLesson))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}

}
