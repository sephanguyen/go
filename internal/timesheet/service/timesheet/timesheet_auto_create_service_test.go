package timesheet

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/timesheet/repository"
	tpb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var errConnectionRefuse = errors.New("connection refused")

func TestTimesheetAutoCreateService_CreateTimesheetMultiple(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		timesheetRepo                  = new(mock_repositories.MockTimesheetRepoImpl)
		timesheetLessonHoursRepo       = new(mock_repositories.MockTimesheetLessonHoursRepoImpl)
		transportationExpenseRepo      = new(mock_repositories.MockTransportationExpenseRepoImpl)
		staffTransportationExpenseRepo = new(mock_repositories.MockStaffTransportationExpenseRepoImpl)

		db = new(mock_database.Ext)
		tx = new(mock_database.Tx)
	)

	s := AutoCreateTimesheetServiceImpl{
		DB:                             db,
		TimesheetRepo:                  timesheetRepo,
		TimesheetLessonHoursRepo:       timesheetLessonHoursRepo,
		TransportationExpenseRepo:      transportationExpenseRepo,
		StaffTransportationExpenseRepo: staffTransportationExpenseRepo,
	}

	timesheetDtos := []*dto.Timesheet{
		{
			StaffID:               GetTimesheetStaffID,
			LocationID:            GetTimesheetLocationID,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         GetTimesheetTimesheetDate,
			ListOtherWorkingHours: nil,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: TimesheetID1,
					LessonID:    TimesheetLessonHoursLessonID1,
				},
				{
					TimesheetID: TimesheetID1,
					LessonID:    TimesheetLessonHoursLessonID2,
				},
			},
		},
		{
			StaffID:               GetTimesheetStaffID,
			LocationID:            GetTimesheetLocationID,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         GetTimesheetTimesheetDate,
			ListOtherWorkingHours: nil,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: TimesheetID2,
					LessonID:    TimesheetLessonHoursLessonID3,
				},
			},
		},
	}
	timesheetDtos2 := []*dto.Timesheet{
		{
			StaffID:               GetTimesheetStaffID,
			LocationID:            GetTimesheetLocationID,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         GetTimesheetTimesheetDate,
			ListOtherWorkingHours: nil,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					LessonID: TimesheetLessonHoursLessonID1,
				},
				{
					LessonID: TimesheetLessonHoursLessonID2,
				},
			},
		},
		{
			StaffID:               GetTimesheetStaffID,
			LocationID:            GetTimesheetLocationID,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         GetTimesheetTimesheetDate,
			ListOtherWorkingHours: nil,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					LessonID: TimesheetLessonHoursLessonID3,
				},
			},
		},
	}
	timesheetDtosWithNothingNew := []*dto.Timesheet{
		{
			ID:                    TimesheetID1,
			StaffID:               GetTimesheetStaffID,
			LocationID:            GetTimesheetLocationID,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         GetTimesheetTimesheetDate,
			ListOtherWorkingHours: nil,
			IsCreated:             true,
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
		},
		{
			ID:                    TimesheetID2,
			StaffID:               GetTimesheetStaffID,
			LocationID:            GetTimesheetLocationID,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         GetTimesheetTimesheetDate,
			ListOtherWorkingHours: nil,
			IsCreated:             true,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: TimesheetID2,
					LessonID:    TimesheetLessonHoursLessonID3,
					IsCreated:   true,
				},
			},
		},
	}
	timesheetDtosWithoutTimesheetLessonHours := []*dto.Timesheet{
		{
			StaffID:                  GetTimesheetStaffID,
			LocationID:               GetTimesheetLocationID,
			TimesheetStatus:          tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:            GetTimesheetTimesheetDate,
			ListOtherWorkingHours:    nil,
			ListTimesheetLessonHours: nil,
		},
		{
			StaffID:                  GetTimesheetStaffID,
			LocationID:               GetTimesheetLocationID,
			TimesheetStatus:          tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:            GetTimesheetTimesheetDate,
			ListOtherWorkingHours:    nil,
			ListTimesheetLessonHours: nil,
		},
	}

	mapStaffTEs := map[string][]entity.StaffTransportationExpense{
		GetTimesheetStaffID: {
			{ID: database.Text(idutil.ULIDNow())},
		},
	}

	testCases := []struct {
		name             string
		ctx              context.Context
		request          interface{}
		timesheetOptions interface{}
		expectedErr      error
		setup            func(ctx context.Context)
	}{
		{
			name:        "create timesheet multiple success",
			ctx:         ctx,
			request:     timesheetDtos,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				staffTransportationExpenseRepo.On("FindListTransportExpensesByStaffIDsAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(mapStaffTEs, nil).Once()

				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil, nil).Once()
				timesheetLessonHoursRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil, nil).Once()
				transportationExpenseRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil).Once()
				tx.On("Commit", mock.Anything).Return(nil)
			},
		},
		{
			name:        "create timesheet multiple success with timesheet not include timesheet lesson hours",
			ctx:         ctx,
			request:     timesheetDtosWithoutTimesheetLessonHours,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				staffTransportationExpenseRepo.On("FindListTransportExpensesByStaffIDsAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(mapStaffTEs, nil).Once()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil, nil).Once()
				transportationExpenseRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil).Once()
				tx.On("Commit", mock.Anything).Return(nil)

			},
		},
		{
			name:        "create timesheet multiple success with timesheet not include timesheet lesson hours",
			ctx:         ctx,
			request:     timesheetDtosWithoutTimesheetLessonHours,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				staffTransportationExpenseRepo.On("FindListTransportExpensesByStaffIDsAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(mapStaffTEs, nil).Once()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil, nil).Once()
				transportationExpenseRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil).Once()
				tx.On("Commit", mock.Anything).Return(nil)

			},
		},
		{
			name:        "create timesheet multiple success  with nothing new",
			ctx:         ctx,
			request:     timesheetDtosWithNothingNew,
			expectedErr: nil,
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "create timesheet multiple failed TimesheetRepo UpsertMultiple error",
			ctx:         ctx,
			request:     timesheetDtos2,
			expectedErr: errConnectionRefuse,
			setup: func(ctx context.Context) {
				staffTransportationExpenseRepo.On("FindListTransportExpensesByStaffIDsAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(mapStaffTEs, nil).Once()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil, errConnectionRefuse).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "error case find staff Transport Expense",
			ctx:         ctx,
			request:     timesheetDtos,
			expectedErr: errConnectionRefuse,
			setup: func(ctx context.Context) {
				staffTransportationExpenseRepo.On("FindListTransportExpensesByStaffIDsAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(nil, errConnectionRefuse).Once()

			},
		},
		{
			name:        "error case upsert multiple lesson hours",
			ctx:         ctx,
			request:     timesheetDtos,
			expectedErr: errConnectionRefuse,
			setup: func(ctx context.Context) {
				staffTransportationExpenseRepo.On("FindListTransportExpensesByStaffIDsAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(mapStaffTEs, nil).Once()

				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil, nil).Once()
				timesheetLessonHoursRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil, errConnectionRefuse).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "error case upsert multiple transport expense",
			ctx:         ctx,
			request:     timesheetDtos,
			expectedErr: errConnectionRefuse,
			setup: func(ctx context.Context) {
				staffTransportationExpenseRepo.On("FindListTransportExpensesByStaffIDsAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(mapStaffTEs, nil).Once()

				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil, nil).Once()
				timesheetLessonHoursRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil, nil).Once()
				transportationExpenseRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(errConnectionRefuse).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.CreateTimesheetMultiple(ctx, testCase.request.([]*dto.Timesheet))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestTimesheetAutoCreateService_RemoveTimesheetLessonHoursMultiple(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		timesheetRepo            = new(mock_repositories.MockTimesheetRepoImpl)
		timesheetLessonHoursRepo = new(mock_repositories.MockTimesheetLessonHoursRepoImpl)
		db                       = new(mock_database.Ext)
		tx                       = new(mock_database.Tx)
	)

	s := AutoCreateTimesheetServiceImpl{
		DB:                       db,
		TimesheetRepo:            timesheetRepo,
		TimesheetLessonHoursRepo: timesheetLessonHoursRepo,
	}

	timesheet1 := []*dto.Timesheet{
		{
			StaffID:               GetTimesheetStaffID,
			LocationID:            GetTimesheetLocationID,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         GetTimesheetTimesheetDate,
			ListOtherWorkingHours: nil,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: TimesheetID1,
					LessonID:    TimesheetLessonHoursLessonID1,
					IsDeleted:   true,
				},
				{
					TimesheetID: TimesheetID1,
					LessonID:    TimesheetLessonHoursLessonID2,
				},
			},
		},
		{
			StaffID:               GetTimesheetStaffID,
			LocationID:            GetTimesheetLocationID,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         GetTimesheetTimesheetDate,
			ListOtherWorkingHours: nil,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: TimesheetID2,
					LessonID:    TimesheetLessonHoursLessonID3,
					IsDeleted:   true,
				},
			},
			IsDeleted: true,
		},
	}
	timesheet2 := []*dto.Timesheet{
		{
			StaffID:               GetTimesheetStaffID,
			LocationID:            GetTimesheetLocationID,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         GetTimesheetTimesheetDate,
			ListOtherWorkingHours: nil,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: TimesheetID1,
					LessonID:    TimesheetLessonHoursLessonID1,
				},
				{
					TimesheetID: TimesheetID1,
					LessonID:    TimesheetLessonHoursLessonID2,
				},
			},
		},
	}
	testCases := []struct {
		name             string
		ctx              context.Context
		request          interface{}
		timesheetOptions interface{}
		expectedErr      error
		setup            func(ctx context.Context)
	}{
		{
			name:        "remove timesheet lesson hours multiple success",
			ctx:         ctx,
			request:     timesheet1,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetLessonHoursRepo.On("SoftDelete", ctx, db, mock.Anything).Return(nil).Once()
				timesheetRepo.On("SoftDeleteByIDs", ctx, db, mock.Anything).Return(nil).Once()
				tx.On("Commit", mock.Anything).Return(nil)
			},
		},
		{
			name:        "remove timesheet lesson hours multiple success with remove nothing",
			ctx:         ctx,
			request:     timesheet2,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				//do nothing
			},
		},
		{
			name:        "remove timesheet lesson hours multiple failed TimesheetLessonHoursRepo SoftDelete error",
			ctx:         ctx,
			request:     timesheet1,
			expectedErr: errConnectionRefuse,
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetLessonHoursRepo.On("SoftDelete", ctx, db, mock.Anything).Return(errConnectionRefuse).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "remove timesheet lesson hours multiple failed TimesheetRepo SoftDeleteByIDs error",
			ctx:         ctx,
			request:     timesheet1,
			expectedErr: errConnectionRefuse,
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetLessonHoursRepo.On("SoftDelete", ctx, db, mock.Anything).Return(nil).Once()
				timesheetRepo.On("SoftDeleteByIDs", ctx, db, mock.Anything).Return(errConnectionRefuse).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)

			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.RemoveTimesheetLessonHoursMultiple(ctx, testCase.request.([]*dto.Timesheet))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestTimesheetAutoCreateService_CreateAndRemoveTimesheetMultiple(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		timesheetRepo                  = new(mock_repositories.MockTimesheetRepoImpl)
		timesheetLessonHoursRepo       = new(mock_repositories.MockTimesheetLessonHoursRepoImpl)
		transportationExpenseRepo      = new(mock_repositories.MockTransportationExpenseRepoImpl)
		staffTransportationExpenseRepo = new(mock_repositories.MockStaffTransportationExpenseRepoImpl)
		db                             = new(mock_database.Ext)
		tx                             = new(mock_database.Tx)
	)

	s := AutoCreateTimesheetServiceImpl{
		DB:                             db,
		TimesheetRepo:                  timesheetRepo,
		TimesheetLessonHoursRepo:       timesheetLessonHoursRepo,
		TransportationExpenseRepo:      transportationExpenseRepo,
		StaffTransportationExpenseRepo: staffTransportationExpenseRepo,
	}

	timesheet1 := []*dto.Timesheet{
		{
			StaffID:               GetTimesheetStaffID,
			LocationID:            GetTimesheetLocationID,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         GetTimesheetTimesheetDate,
			ListOtherWorkingHours: nil,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: TimesheetID1,
					LessonID:    TimesheetLessonHoursLessonID1,
					IsDeleted:   true,
				},
				{
					TimesheetID: TimesheetID1,
					LessonID:    TimesheetLessonHoursLessonID2,
				},
			},
		},
		{
			StaffID:               GetTimesheetStaffID,
			LocationID:            GetTimesheetLocationID,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         GetTimesheetTimesheetDate,
			ListOtherWorkingHours: nil,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: TimesheetID2,
					LessonID:    TimesheetLessonHoursLessonID3,
					IsDeleted:   true,
				},
			},
			IsDeleted: true,
		},
	}
	timesheet2 := []*dto.Timesheet{
		{
			StaffID:               GetTimesheetStaffID,
			LocationID:            GetTimesheetLocationID,
			TimesheetStatus:       tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:         GetTimesheetTimesheetDate,
			ListOtherWorkingHours: nil,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: TimesheetID1,
					LessonID:    TimesheetLessonHoursLessonID1,
				},
				{
					TimesheetID: TimesheetID1,
					LessonID:    TimesheetLessonHoursLessonID2,
				},
			},
		},
	}

	mapStaffTEs := map[string][]entity.StaffTransportationExpense{
		GetTimesheetStaffID: {
			{ID: database.Text(idutil.ULIDNow())},
		},
	}

	testCases := []struct {
		name                string
		ctx                 context.Context
		reqNewTimesheets    interface{}
		reqRemoveTimesheets interface{}
		expectedErr         error
		setup               func(ctx context.Context)
	}{
		{
			name:                "create and remove timesheet lesson hours multiple success",
			ctx:                 ctx,
			reqNewTimesheets:    timesheet2,
			reqRemoveTimesheets: timesheet1,
			expectedErr:         nil,
			setup: func(ctx context.Context) {
				staffTransportationExpenseRepo.On("FindListTransportExpensesByStaffIDsAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(mapStaffTEs, nil).Once()

				db.On("Begin", mock.Anything).Once().Return(tx, nil)

				timesheetRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil, nil).Once()
				timesheetLessonHoursRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil, nil).Once()
				transportationExpenseRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil).Once()
				timesheetLessonHoursRepo.On("SoftDelete", ctx, tx, mock.Anything).Return(nil).Once()
				timesheetRepo.On("SoftDeleteByIDs", ctx, tx, mock.Anything).Return(nil).Once()
				transportationExpenseRepo.On("SoftDeleteMultipleByTimesheetIDs", ctx, tx, mock.Anything).Return(nil).Once()

				tx.On("Commit", mock.Anything).Return(nil)
			},
		},
		{
			name:                "create and remove timesheet lesson hours multiple success with nothing create or update",
			ctx:                 ctx,
			reqNewTimesheets:    ([]*dto.Timesheet)(nil),
			reqRemoveTimesheets: ([]*dto.Timesheet)(nil),
			expectedErr:         nil,
			setup: func(ctx context.Context) {
				//do nothing
			},
		},
		{
			name:                "create timesheet multiple failed",
			ctx:                 ctx,
			reqNewTimesheets:    timesheet2,
			reqRemoveTimesheets: timesheet1,
			expectedErr:         errConnectionRefuse,
			setup: func(ctx context.Context) {
				staffTransportationExpenseRepo.On("FindListTransportExpensesByStaffIDsAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(mapStaffTEs, nil).Once()

				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil, errConnectionRefuse).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:                "create lesson hours multiple failed",
			ctx:                 ctx,
			reqNewTimesheets:    timesheet2,
			reqRemoveTimesheets: timesheet1,
			expectedErr:         errConnectionRefuse,
			setup: func(ctx context.Context) {
				staffTransportationExpenseRepo.On("FindListTransportExpensesByStaffIDsAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(mapStaffTEs, nil).Once()

				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil, nil).Once()
				timesheetLessonHoursRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil, errConnectionRefuse).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:                "remove timesheet failed",
			ctx:                 ctx,
			reqNewTimesheets:    timesheet2,
			reqRemoveTimesheets: timesheet1,
			expectedErr:         errConnectionRefuse,
			setup: func(ctx context.Context) {
				staffTransportationExpenseRepo.On("FindListTransportExpensesByStaffIDsAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(mapStaffTEs, nil).Once()

				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil, nil).Once()
				timesheetLessonHoursRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil, nil).Once()
				transportationExpenseRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil).Once()
				timesheetLessonHoursRepo.On("SoftDelete", ctx, tx, mock.Anything).Return(errConnectionRefuse).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)

			},
		},
		{
			name:                "remove timesheet lesson hours multiple failed",
			ctx:                 ctx,
			reqNewTimesheets:    timesheet2,
			reqRemoveTimesheets: timesheet1,
			expectedErr:         errConnectionRefuse,
			setup: func(ctx context.Context) {
				staffTransportationExpenseRepo.On("FindListTransportExpensesByStaffIDsAndLocation", ctx, db, mock.Anything, mock.Anything).
					Return(mapStaffTEs, nil).Once()

				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil, nil).Once()
				timesheetLessonHoursRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil, nil).Once()
				transportationExpenseRepo.On("UpsertMultiple", ctx, tx, mock.Anything).Return(nil).Once()
				timesheetLessonHoursRepo.On("SoftDelete", ctx, tx, mock.Anything).Return(nil).Once()
				timesheetRepo.On("SoftDeleteByIDs", ctx, tx, mock.Anything).Return(errConnectionRefuse).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)

			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.CreateAndRemoveTimesheetMultiple(
				ctx,
				testCase.reqNewTimesheets.([]*dto.Timesheet),
				testCase.reqRemoveTimesheets.([]*dto.Timesheet),
			)

			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestTimesheetAutoCreateService_UpdateLessonAutoCreateFlagState(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		timesheetRepo            = new(mock_repositories.MockTimesheetRepoImpl)
		timesheetLessonHoursRepo = new(mock_repositories.MockTimesheetLessonHoursRepoImpl)
		db                       = new(mock_database.Ext)
		tx                       = new(mock_database.Tx)
	)

	s := AutoCreateTimesheetServiceImpl{
		DB:                       db,
		TimesheetRepo:            timesheetRepo,
		TimesheetLessonHoursRepo: timesheetLessonHoursRepo,
	}

	testCases := []struct {
		name             string
		ctx              context.Context
		request          interface{}
		timesheetOptions interface{}
		expectedErr      error
		setup            func(ctx context.Context)
	}{
		{
			name:        "happy case update timesheet lesson auto create flag success",
			ctx:         ctx,
			request:     map[bool][]string{false: {"staff_id"}},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetLessonHoursRepo.On("UpdateTimesheetLessonAutoCreateFlagByTimesheetIDs", ctx, tx, mock.Anything, mock.Anything).Return(nil).Once()
				tx.On("Commit", mock.Anything).Return(nil)
			},
		},
		{
			name:        "failed case update timesheet lesson auto create flag",
			ctx:         ctx,
			request:     map[bool][]string{false: {"staff_id"}},
			expectedErr: errConnectionRefuse,
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetLessonHoursRepo.On("UpdateTimesheetLessonAutoCreateFlagByTimesheetIDs", ctx, tx, mock.Anything, mock.Anything).Return(errConnectionRefuse).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.UpdateLessonAutoCreateFlagState(ctx, testCase.request.(map[bool][]string))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}
