package timesheet

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/timesheet/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	AutoCreateFlagStaffID = idutil.ULIDNow()
	AutoCreateFlagFlagOn  = false
)

func TestAutoCreateTimesheetFlagService_UpsertFlag(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var (
		autoCreateFlagRepo       = new(mock_repositories.MockAutoCreateFlagRepoImpl)
		autoCreateFlagLogRepo    = new(mock_repositories.MockAutoCreateFlagActivityLogRepoImpl)
		timesheetRepo            = new(mock_repositories.MockTimesheetRepoImpl)
		timesheetLessonHoursRepo = new(mock_repositories.MockTimesheetLessonHoursRepoImpl)
		otherWorkingHoursRepo    = new(mock_repositories.MockOtherWorkingHoursRepoImpl)
		db                       = new(mock_database.Ext)
		tx                       = new(mock_database.Tx)
	)

	flagLogE := &entity.AutoCreateFlagActivityLog{
		StaffID:    database.Text(AutoCreateFlagStaffID),
		FlagOn:     database.Bool(AutoCreateFlagFlagOn),
		ChangeTime: database.Timestamptz(now),
		CreatedAt:  database.Timestamptz(now),
		UpdatedAt:  database.Timestamptz(now),
	}

	s := AutoCreateTimesheetFlagServiceImpl{
		DB:                       db,
		AutoCreateFlagRepo:       autoCreateFlagRepo,
		AutoCreateFlagLogRepo:    autoCreateFlagLogRepo,
		TimesheetRepo:            timesheetRepo,
		TimesheetLessonHoursRepo: timesheetLessonHoursRepo,
		OtherWorkingHoursRepo:    otherWorkingHoursRepo,
	}

	testCases := []TestCase{
		{
			name:        "happy case",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req:         &dto.AutoCreateTimesheetFlag{StaffID: AutoCreateFlagStaffID, FlagOn: false},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				autoCreateFlagRepo.On("Upsert", ctx, tx, mock.Anything).
					Return(nil).Once()
				autoCreateFlagLogRepo.On("SoftDeleteFlagLogsAfterTime", ctx, tx, mock.Anything, mock.Anything).
					Return(nil).Once()
				autoCreateFlagLogRepo.On("InsertFlagLog", ctx, tx, mock.Anything).
					Return(flagLogE, nil).Once()
				tx.On("Commit", mock.Anything).Return(nil)

			},
		},
		{
			name:        "error case failed create or update auto create flag",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req:         &dto.AutoCreateTimesheetFlag{StaffID: AutoCreateFlagStaffID, FlagOn: false},
			expectedErr: status.Error(codes.Internal, "transaction error: create or update auto create flag error: internal error"),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				autoCreateFlagRepo.On("Upsert", ctx, tx, mock.Anything).
					Return(errors.New("internal error")).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "error case failed to remove flag logs within time range",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req:         &dto.AutoCreateTimesheetFlag{StaffID: AutoCreateFlagStaffID, FlagOn: false},
			expectedErr: status.Error(codes.Internal, "transaction error: soft delete flag logs within time range error: internal error"),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				autoCreateFlagRepo.On("Upsert", ctx, tx, mock.Anything).
					Return(nil).Once()
				autoCreateFlagLogRepo.On("SoftDeleteFlagLogsAfterTime", ctx, tx, mock.Anything, mock.Anything).
					Return(errors.New("internal error")).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
	}

	// Do test
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*dto.AutoCreateTimesheetFlag)
			err := s.UpsertFlag(testCase.ctx, req)
			assert.Equal(t, testCase.expectedErr, err)
			mock.AssertExpectationsForObjects(
				t,
				autoCreateFlagRepo,
				db,
				tx,
			)
		})
	}
}

func TestAutoCreateTimesheetFlagService_UpdateLessonHoursFlag(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var (
		autoCreateFlagRepo       = new(mock_repositories.MockAutoCreateFlagRepoImpl)
		autoCreateFlagLogRepo    = new(mock_repositories.MockAutoCreateFlagActivityLogRepoImpl)
		timesheetRepo            = new(mock_repositories.MockTimesheetRepoImpl)
		timesheetLessonHoursRepo = new(mock_repositories.MockTimesheetLessonHoursRepoImpl)
		otherWorkingHoursRepo    = new(mock_repositories.MockOtherWorkingHoursRepoImpl)
		db                       = new(mock_database.Ext)
		tx                       = new(mock_database.Tx)
	)

	s := AutoCreateTimesheetFlagServiceImpl{
		DB:                       db,
		AutoCreateFlagRepo:       autoCreateFlagRepo,
		AutoCreateFlagLogRepo:    autoCreateFlagLogRepo,
		TimesheetRepo:            timesheetRepo,
		TimesheetLessonHoursRepo: timesheetLessonHoursRepo,
		OtherWorkingHoursRepo:    otherWorkingHoursRepo,
	}

	timesheetIDs := []string{"test-id"}
	testCases := []TestCase{
		{
			name:        "happy case",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req:         &dto.AutoCreateTimesheetFlag{StaffID: AutoCreateFlagStaffID, FlagOn: false},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				timesheetRepo.On("GetStaffTimesheetIDsAfterDateCanChange", ctx, db, mock.Anything, mock.Anything).
					Return(timesheetIDs, nil).Once()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetLessonHoursRepo.On("UpdateAutoCreateFlagStateAfterTime", ctx, tx, mock.Anything, mock.Anything, mock.Anything).
					Return(nil).Once()
				timesheetLessonHoursRepo.On("MapExistingLessonHoursByTimesheetIds", ctx, tx, mock.Anything).
					Return(map[string]struct{}{}, nil).Once()
				otherWorkingHoursRepo.On("MapExistingOWHsByTimesheetIds", ctx, tx, mock.Anything).
					Return(map[string]struct{}{}, nil).Once()
				timesheetRepo.On("RemoveTimesheetRemarkByTimesheetIDs", ctx, tx, mock.Anything).
					Return(nil).Once()
				tx.On("Commit", mock.Anything).Return(nil)

			},
		},
		{
			name:        "error case failed to update auto create flag after time",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req:         &dto.AutoCreateTimesheetFlag{StaffID: AutoCreateFlagStaffID, FlagOn: false},
			expectedErr: status.Error(codes.Internal, "transaction error: update future auto create flag error: internal error"),
			setup: func(ctx context.Context) {
				timesheetRepo.On("GetStaffTimesheetIDsAfterDateCanChange", ctx, db, mock.Anything, mock.Anything).
					Return(timesheetIDs, nil).Once()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetLessonHoursRepo.On("UpdateAutoCreateFlagStateAfterTime", ctx, tx, mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("internal error")).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "error case failed to retrieve list of lesson hours by timesheet ids",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req:         &dto.AutoCreateTimesheetFlag{StaffID: AutoCreateFlagStaffID, FlagOn: false},
			expectedErr: status.Error(codes.Internal, "transaction error: get list timesheet lesson hours by timesheet ids error: internal error"),
			setup: func(ctx context.Context) {
				timesheetRepo.On("GetStaffTimesheetIDsAfterDateCanChange", ctx, db, mock.Anything, mock.Anything).
					Return(timesheetIDs, nil).Once()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetLessonHoursRepo.On("UpdateAutoCreateFlagStateAfterTime", ctx, tx, mock.Anything, mock.Anything, mock.Anything).
					Return(nil).Once()
				timesheetLessonHoursRepo.On("MapExistingLessonHoursByTimesheetIds", ctx, tx, mock.Anything).
					Return(map[string]struct{}{}, errors.New("internal error")).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "error case failed to retrieve list of OWH hours by timesheet ids",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req:         &dto.AutoCreateTimesheetFlag{StaffID: AutoCreateFlagStaffID, FlagOn: false},
			expectedErr: status.Error(codes.Internal, "transaction error: get list timesheet OHWs by timesheet ids error: internal error"),
			setup: func(ctx context.Context) {
				timesheetRepo.On("GetStaffTimesheetIDsAfterDateCanChange", ctx, db, mock.Anything, mock.Anything).
					Return(timesheetIDs, nil).Once()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetLessonHoursRepo.On("UpdateAutoCreateFlagStateAfterTime", ctx, tx, mock.Anything, mock.Anything, mock.Anything).
					Return(nil).Once()
				timesheetLessonHoursRepo.On("MapExistingLessonHoursByTimesheetIds", ctx, tx, mock.Anything).
					Return(map[string]struct{}{}, nil).Once()
				otherWorkingHoursRepo.On("MapExistingOWHsByTimesheetIds", ctx, tx, mock.Anything).
					Return(map[string]struct{}{}, errors.New("internal error")).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "error case failed to remove empty remark for empty timesheet",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req:         &dto.AutoCreateTimesheetFlag{StaffID: AutoCreateFlagStaffID, FlagOn: false},
			expectedErr: status.Error(codes.Internal, "transaction error: remove remark for empty timesheet error: internal error"),
			setup: func(ctx context.Context) {
				timesheetRepo.On("GetStaffTimesheetIDsAfterDateCanChange", ctx, db, mock.Anything, mock.Anything).
					Return(timesheetIDs, nil).Once()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				timesheetLessonHoursRepo.On("UpdateAutoCreateFlagStateAfterTime", ctx, tx, mock.Anything, mock.Anything, mock.Anything).
					Return(nil).Once()
				timesheetLessonHoursRepo.On("MapExistingLessonHoursByTimesheetIds", ctx, tx, mock.Anything).
					Return(map[string]struct{}{}, nil).Once()
				otherWorkingHoursRepo.On("MapExistingOWHsByTimesheetIds", ctx, tx, mock.Anything).
					Return(map[string]struct{}{}, nil).Once()
				timesheetRepo.On("RemoveTimesheetRemarkByTimesheetIDs", ctx, tx, mock.Anything).
					Return(errors.New("internal error")).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
	}

	// Do test
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*dto.AutoCreateTimesheetFlag)
			err := s.UpdateLessonHoursFlag(testCase.ctx, req)
			assert.Equal(t, testCase.expectedErr, err)
			mock.AssertExpectationsForObjects(
				t,
				autoCreateFlagRepo,
				db,
				tx,
			)
		})
	}
}
