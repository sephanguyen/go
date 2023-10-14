package timesheet

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/timesheet/domain/constant"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_repositories "github.com/manabie-com/backend/mock/timesheet/repository"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	draftID                         = "draft-id"
	submittedID                     = "submitted-id"
	approvedID                      = "approved-id"
	confirmedID                     = "confirmed-id"
	timesheetShouldBeInDraftErr     = "timesheet status should be in Draft"
	invalidLessonStatusErr          = "invalid lesson status"
	submitStatusTxErr               = "update timesheet status submitted error: tx is closed"
	timesheetShouldBeInSubmittedErr = "timesheet status should be in Submitted"
)

func TestTimesheetService_DeleteTimesheet(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)
	mockTimesheetRepo := new(mock_repositories.MockTimesheetRepoImpl)
	mockTimesheetLessonHourRepo := new(mock_repositories.MockTimesheetLessonHoursRepoImpl)
	mockOtherWorkingHoursRepo := new(mock_repositories.MockOtherWorkingHoursRepoImpl)
	mockTransportExpenseRepo := new(mock_repositories.MockTransportationExpenseRepoImpl)

	s := TimesheetStateMachineService{
		DB:                        mockDB,
		TimesheetRepo:             mockTimesheetRepo,
		TimesheetLessonHoursRepo:  mockTimesheetLessonHourRepo,
		OtherWorkingHoursRepo:     mockOtherWorkingHoursRepo,
		TransportationExpenseRepo: mockTransportExpenseRepo,
	}
	timesheetInDraft := &entity.Timesheet{
		TimesheetID:     database.Text(draftID),
		CreatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		UpdatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		StaffID:         database.Text(CreateTimesheetStaffID),
		LocationID:      database.Text(CreateTimesheetLocationID),
		Remark:          database.Text(CreateTimeSheetRemark),
		TimesheetDate:   database.Timestamptz(CreateTimesheetTimesheetDate),
		TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String()),
	}

	timesheetInSubmitted := &entity.Timesheet{
		TimesheetID:     database.Text(submittedID),
		CreatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		UpdatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		StaffID:         database.Text(CreateTimesheetStaffID),
		LocationID:      database.Text(CreateTimesheetLocationID),
		Remark:          database.Text(CreateTimeSheetRemark),
		TimesheetDate:   database.Timestamptz(CreateTimesheetTimesheetDate),
		TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String()),
	}

	timesheetInApproved := &entity.Timesheet{
		TimesheetID:     database.Text(approvedID),
		CreatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		UpdatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		StaffID:         database.Text(CreateTimesheetStaffID),
		LocationID:      database.Text(CreateTimesheetLocationID),
		Remark:          database.Text(CreateTimeSheetRemark),
		TimesheetDate:   database.Timestamptz(CreateTimesheetTimesheetDate),
		TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String()),
	}

	timesheetInConfirmed := &entity.Timesheet{
		TimesheetID:     database.Text(confirmedID),
		CreatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		UpdatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		StaffID:         database.Text(CreateTimesheetStaffID),
		LocationID:      database.Text(CreateTimesheetLocationID),
		Remark:          database.Text(CreateTimeSheetRemark),
		TimesheetDate:   database.Timestamptz(CreateTimesheetTimesheetDate),
		TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_CONFIRMED.String()),
	}

	timesheetInDraftWithLessonHour := &entity.Timesheet{
		TimesheetID:     database.Text("draft-id-2"),
		CreatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		UpdatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		StaffID:         database.Text(CreateTimesheetStaffID),
		LocationID:      database.Text(CreateTimesheetLocationID),
		Remark:          database.Text(CreateTimeSheetRemark),
		TimesheetDate:   database.Timestamptz(CreateTimesheetTimesheetDate),
		TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String()),
	}

	timesheetBeforeDate := &entity.Timesheet{
		TimesheetID:     database.Text(submittedID),
		CreatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		UpdatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		StaffID:         database.Text(CreateTimesheetStaffID),
		LocationID:      database.Text(CreateTimesheetLocationID),
		Remark:          database.Text(CreateTimeSheetRemark),
		TimesheetDate:   database.Timestamptz(CreateTimesheetTimesheetDate.Add(-24 * time.Hour)),
		TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String()),
	}

	timesheetLessonHourSingleRecord := []*entity.TimesheetLessonHours{
		{
			TimesheetID: timesheetInDraftWithLessonHour.TimesheetID,
			LessonID:    database.Text("3"),
			FlagOn:      database.Bool(true),
		},
	}

	timesheetLessonHourMultiRecords := []*entity.TimesheetLessonHours{
		{
			TimesheetID: timesheetInDraftWithLessonHour.TimesheetID,
			LessonID:    database.Text("1"),
			FlagOn:      database.Bool(true),
		},
		{
			TimesheetID: timesheetInDraftWithLessonHour.TimesheetID,
			LessonID:    database.Text("2"),
			FlagOn:      database.Bool(true),
		},
	}

	testcases := []TestCase{
		{
			name:        "happy case delete draft timesheet with requester",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: nil,
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return([]*entity.TimesheetLessonHours{}, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockOtherWorkingHoursRepo.On("SoftDeleteByTimesheetID", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTransportExpenseRepo.On("SoftDeleteByTimesheetID", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTimesheetRepo.On("SoftDeleteByIDs", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
		},
		{
			name:        "happy case delete draft timesheet with approver and confirmer",
			ctx:         interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr: nil,
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return([]*entity.TimesheetLessonHours{}, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockOtherWorkingHoursRepo.On("SoftDeleteByTimesheetID", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTransportExpenseRepo.On("SoftDeleteByTimesheetID", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTimesheetRepo.On("SoftDeleteByIDs", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
		},
		{
			name:        "happy case delete submitted timesheet with approver and confirmer",
			ctx:         interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr: nil,
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInSubmitted, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return([]*entity.TimesheetLessonHours{}, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockOtherWorkingHoursRepo.On("SoftDeleteByTimesheetID", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTransportExpenseRepo.On("SoftDeleteByTimesheetID", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTimesheetRepo.On("SoftDeleteByIDs", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
		},
		{
			name:        "happy case delete old timesheet with approver and confirmer",
			ctx:         interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr: nil,
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetBeforeDate, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return([]*entity.TimesheetLessonHours{}, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockOtherWorkingHoursRepo.On("SoftDeleteByTimesheetID", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTransportExpenseRepo.On("SoftDeleteByTimesheetID", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTimesheetRepo.On("SoftDeleteByIDs", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
			},
		},
		{
			name:        "error case unauthorized to delete timesheet",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDFail),
			reqString:   draftID,
			expectedErr: status.Error(codes.PermissionDenied, fmt.Sprintf("unauthorized to modify timesheet, timesheetStaffID: %s", CreateTimesheetStaffID)),
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
			},
		},
		{
			name:        "find timesheet failed tx is closed",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDFail),
			expectedErr: status.Error(codes.Internal, "find timesheet error: tx is closed"),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "find timesheet failed no rows result",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDFail),
			expectedErr: status.Error(codes.Internal, "find timesheet error: no rows in result set"),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:        "failed delete timesheet status in submitted with requester",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.PermissionDenied, fmt.Sprintf("unauthorized to modify timesheet, timesheetStaffID: %s", CreateTimesheetStaffID)),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInSubmitted, nil)
			},
		},
		{
			name:        "failed timesheet status in approved",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.FailedPrecondition, fmt.Sprintf("can not delete timesheet %s in : %s status", approvedID, pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String())),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInApproved, nil)
			},
		},
		{
			name:        "failed timesheet status in confirmed",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.FailedPrecondition, fmt.Sprintf("can not delete timesheet %s in : %s status", confirmedID, pb.TimesheetStatus_TIMESHEET_STATUS_CONFIRMED.String())),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInConfirmed, nil)
			},
		},
		{
			name:        "find timesheet lesson hour records tx is closed",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.Internal, "find timesheet lesson hours error: tx is closed"),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "find timesheet lesson hour records no rows result",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.Internal, "find timesheet lesson hours error: no rows in result set"),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:        "failed deleting timesheet record contain single lesson hour record",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.FailedPrecondition, "timesheet record has lesson record"),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetLessonHourSingleRecord, nil)
			},
		},
		{
			name:        "failed deleting timesheet record contain multiple lesson hour record",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.FailedPrecondition, "timesheet record has lesson record"),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetLessonHourMultiRecords, nil)
			},
		},
		{
			name:        "failed deleting with other working hours record",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.Internal, "delete other working hours error: tx is closed"),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockOtherWorkingHoursRepo.On("SoftDeleteByTimesheetID", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", mock.Anything).Return(nil)

			},
		},
		{
			name:        "failed deleting with transportation expenses record",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.Internal, "delete transport expenses error: tx is closed"),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockOtherWorkingHoursRepo.On("SoftDeleteByTimesheetID", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTransportExpenseRepo.On("SoftDeleteByTimesheetID", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", mock.Anything).Return(nil)

			},
		},
		{
			name:        "failed deleting timesheet record with other working hours",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.Internal, "delete timesheet error: tx is closed"),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(nil, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockOtherWorkingHoursRepo.On("SoftDeleteByTimesheetID", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTransportExpenseRepo.On("SoftDeleteByTimesheetID", ctx, mockTx, mock.Anything).Once().Return(nil)
				mockTimesheetRepo.On("SoftDeleteByIDs", ctx, mockTx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				mockTx.On("Rollback", mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			err := s.DeleteTimesheet(testCase.ctx, testCase.reqString)
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockTimesheetRepo, mockTimesheetLessonHourRepo, mockOtherWorkingHoursRepo)
		})
	}
}

func TestTimesheetService_SubmitTimesheet(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockJsm := new(mock_nats.JetStreamManagement)
	mockTimesheetRepo := new(mock_repositories.MockTimesheetRepoImpl)
	mockTimesheetLessonHourRepo := new(mock_repositories.MockTimesheetLessonHoursRepoImpl)
	mockLessonRepo := new(mock_repositories.MockLessonRepoImpl)

	s := TimesheetStateMachineService{
		DB:                       mockDB,
		JSM:                      mockJsm,
		TimesheetRepo:            mockTimesheetRepo,
		TimesheetLessonHoursRepo: mockTimesheetLessonHourRepo,
		LessonRepo:               mockLessonRepo,
	}

	timesheetInDraft := &entity.Timesheet{
		TimesheetID:     database.Text(draftID),
		CreatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		UpdatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		StaffID:         database.Text(CreateTimesheetStaffID),
		LocationID:      database.Text(CreateTimesheetLocationID),
		Remark:          database.Text(CreateTimeSheetRemark),
		TimesheetDate:   database.Timestamptz(CreateTimesheetTimesheetDate),
		TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String()),
	}

	timesheetUpdatedEntity := &entity.Timesheet{
		TimesheetID:     database.Text(draftID),
		CreatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		UpdatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		StaffID:         database.Text(CreateTimesheetStaffID),
		LocationID:      database.Text(CreateTimesheetLocationID),
		Remark:          database.Text(CreateTimeSheetRemark),
		TimesheetDate:   database.Timestamptz(CreateTimesheetTimesheetDate),
		TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String()),
	}

	timesheetLessonHourSingleRecordCancelled := []*entity.TimesheetLessonHours{
		{
			TimesheetID: timesheetInDraft.TimesheetID,
			LessonID:    database.Text("3"),
		},
	}

	lessonSingleRecordCancelled := []*entity.Lesson{
		{
			LessonID:         timesheetLessonHourSingleRecordCancelled[0].LessonID,
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED.String()),
		},
	}

	timesheetLessonHourMultiRecordCancelled := []*entity.TimesheetLessonHours{
		{
			TimesheetID: timesheetInDraft.TimesheetID,
			LessonID:    database.Text("4"),
		},
		{
			TimesheetID: timesheetInDraft.TimesheetID,
			LessonID:    database.Text("5"),
		},
		{
			TimesheetID: timesheetInDraft.TimesheetID,
			LessonID:    database.Text("6"),
		},
	}

	lessonMultiRecordCancelled := []*entity.Lesson{
		{
			LessonID:         timesheetLessonHourMultiRecordCancelled[0].LessonID,
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED.String()),
		},
		{
			LessonID:         timesheetLessonHourMultiRecordCancelled[1].LessonID,
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED.String()),
		},
		{
			LessonID:         timesheetLessonHourMultiRecordCancelled[2].LessonID,
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED.String()),
		},
	}

	timesheetLessonHourSingleRecordCompleted := []*entity.TimesheetLessonHours{
		{
			TimesheetID: timesheetInDraft.TimesheetID,
			LessonID:    database.Text("7"),
		},
	}

	lessonSingleRecordCompleted := []*entity.Lesson{
		{
			LessonID:         timesheetLessonHourSingleRecordCompleted[0].LessonID,
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED.String()),
		},
	}

	timesheetLessonHourMultiRecordCompleted := []*entity.TimesheetLessonHours{
		{
			TimesheetID: timesheetInDraft.TimesheetID,
			LessonID:    database.Text("8"),
		},
		{
			TimesheetID: timesheetInDraft.TimesheetID,
			LessonID:    database.Text("9"),
		},
		{
			TimesheetID: timesheetInDraft.TimesheetID,
			LessonID:    database.Text("10"),
		},
	}

	lessonMultiRecordCompleted := []*entity.Lesson{
		{
			LessonID:         timesheetLessonHourMultiRecordCompleted[0].LessonID,
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED.String()),
		},
		{
			LessonID:         timesheetLessonHourMultiRecordCompleted[1].LessonID,
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED.String()),
		},
		{

			LessonID:         timesheetLessonHourMultiRecordCompleted[2].LessonID,
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED.String()),
		},
	}

	timesheetLessonHourMultiRecordCompletedCancelled := []*entity.TimesheetLessonHours{
		{
			TimesheetID: timesheetInDraft.TimesheetID,
			LessonID:    database.Text("11"),
		},
		{
			TimesheetID: timesheetInDraft.TimesheetID,
			LessonID:    database.Text("12"),
		},
		{
			TimesheetID: timesheetInDraft.TimesheetID,
			LessonID:    database.Text("13"),
		},
	}

	lessonMultiRecordCompletedCancelled := []*entity.Lesson{
		{
			LessonID:         timesheetLessonHourMultiRecordCompletedCancelled[0].LessonID,
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED.String()),
		},
		{
			LessonID:         timesheetLessonHourMultiRecordCompletedCancelled[1].LessonID,
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED.String()),
		},
		{

			LessonID:         timesheetLessonHourMultiRecordCompletedCancelled[2].LessonID,
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED.String()),
		},
	}

	timesheetLessonHourSingleRecordPublished := []*entity.TimesheetLessonHours{
		{
			TimesheetID: timesheetInDraft.TimesheetID,
			LessonID:    database.Text("14"),
		},
	}

	lessonSingleRecordPublished := []*entity.Lesson{
		{
			LessonID:         timesheetLessonHourSingleRecordPublished[0].LessonID,
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED.String()),
		},
	}

	timesheetLessonHourMultiRecordPublished := []*entity.TimesheetLessonHours{
		{
			TimesheetID: timesheetInDraft.TimesheetID,
			LessonID:    database.Text("15"),
		},
		{
			TimesheetID: timesheetInDraft.TimesheetID,
			LessonID:    database.Text("16"),
		},
		{
			TimesheetID: timesheetInDraft.TimesheetID,
			LessonID:    database.Text("17"),
		},
	}

	lessonMultiRecordPublished := []*entity.Lesson{
		{
			LessonID:         timesheetLessonHourMultiRecordPublished[0].LessonID,
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED.String()),
		},
		{
			LessonID:         timesheetLessonHourMultiRecordPublished[1].LessonID,
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED.String()),
		},
		{
			LessonID:         timesheetLessonHourMultiRecordPublished[2].LessonID,
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED.String()),
		},
	}

	timesheetLessonHourMultiRecordPublishedCancelled := []*entity.TimesheetLessonHours{
		{
			TimesheetID: timesheetInDraft.TimesheetID,
			LessonID:    database.Text("18"),
		},
		{
			TimesheetID: timesheetInDraft.TimesheetID,
			LessonID:    database.Text("19"),
		},
		{
			TimesheetID: timesheetInDraft.TimesheetID,
			LessonID:    database.Text("20"),
		},
	}

	lessonMultiRecordPublishedCancelled := []*entity.Lesson{
		{
			LessonID:         timesheetLessonHourMultiRecordPublishedCancelled[0].LessonID,
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED.String()),
		},
		{
			LessonID:         timesheetLessonHourMultiRecordPublished[1].LessonID,
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED.String()),
		},
		{
			LessonID:         timesheetLessonHourMultiRecordPublishedCancelled[2].LessonID,
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED.String()),
		},
	}

	timesheetLessonHourMultiRecordPublishedCompleted := []*entity.TimesheetLessonHours{
		{
			TimesheetID: timesheetInDraft.TimesheetID,
			LessonID:    database.Text("21"),
		},
		{
			TimesheetID: timesheetInDraft.TimesheetID,
			LessonID:    database.Text("22"),
		},
		{
			TimesheetID: timesheetInDraft.TimesheetID,
			LessonID:    database.Text("23"),
		},
	}

	lessonMultiRecordPublishedCompleted := []*entity.Lesson{
		{
			LessonID:         timesheetLessonHourMultiRecordPublishedCompleted[0].LessonID,
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED.String()),
		},
		{
			LessonID:         timesheetLessonHourMultiRecordPublishedCompleted[1].LessonID,
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED.String()),
		},
		{
			LessonID:         timesheetLessonHourMultiRecordPublishedCompleted[2].LessonID,
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED.String()),
		},
	}

	timesheetInSubmitted := &entity.Timesheet{
		TimesheetID:     database.Text(submittedID),
		CreatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		UpdatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		StaffID:         database.Text(CreateTimesheetStaffID),
		LocationID:      database.Text(CreateTimesheetLocationID),
		Remark:          database.Text(CreateTimeSheetRemark),
		TimesheetDate:   database.Timestamptz(CreateTimesheetTimesheetDate),
		TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String()),
	}

	timesheetInApproved := &entity.Timesheet{
		TimesheetID:     database.Text(approvedID),
		CreatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		UpdatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		StaffID:         database.Text(CreateTimesheetStaffID),
		LocationID:      database.Text(CreateTimesheetLocationID),
		Remark:          database.Text(CreateTimeSheetRemark),
		TimesheetDate:   database.Timestamptz(CreateTimesheetTimesheetDate),
		TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String()),
	}

	timesheetInConfirmed := &entity.Timesheet{
		TimesheetID:     database.Text(confirmedID),
		CreatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		UpdatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		StaffID:         database.Text(CreateTimesheetStaffID),
		LocationID:      database.Text(CreateTimesheetLocationID),
		Remark:          database.Text(CreateTimeSheetRemark),
		TimesheetDate:   database.Timestamptz(CreateTimesheetTimesheetDate),
		TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_CONFIRMED.String()),
	}

	timesheetDateTomorrow := &entity.Timesheet{
		TimesheetID:     database.Text(submittedID),
		CreatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		UpdatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		StaffID:         database.Text(CreateTimesheetStaffID),
		LocationID:      database.Text(CreateTimesheetLocationID),
		Remark:          database.Text(CreateTimeSheetRemark),
		TimesheetDate:   database.Timestamptz(time.Now().AddDate(0, 0, 1)),
		TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String()),
	}

	testcases := []TestCase{
		{
			name:        "happy case submit timesheet record no lesson hours",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: nil,
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return([]*entity.TimesheetLessonHours{}, nil)
				mockTimesheetRepo.On("UpdateTimeSheet", ctx, mockDB, mock.Anything).Once().Return(timesheetUpdatedEntity, nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:        "happy case submit timesheet record has lesson hour cancelled",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: nil,
			reqString:   draftID,
			setup: func(ctx context.Context) {
				timesheetInDraft.TimesheetStatus = database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetLessonHourSingleRecordCancelled, nil)
				mockLessonRepo.On("FindLessonsByIDs", ctx, mockDB, mock.Anything).Once().Return(lessonSingleRecordCancelled, nil)
				mockTimesheetRepo.On("UpdateTimeSheet", ctx, mockDB, mock.Anything).Once().Return(timesheetUpdatedEntity, nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:        "happy case submit timesheet record have multiple lesson hours cancelled",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: nil,
			reqString:   draftID,
			setup: func(ctx context.Context) {
				timesheetInDraft.TimesheetStatus = database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetLessonHourMultiRecordCancelled, nil)
				mockLessonRepo.On("FindLessonsByIDs", ctx, mockDB, mock.Anything).Once().Return(lessonMultiRecordCancelled, nil)
				mockTimesheetRepo.On("UpdateTimeSheet", ctx, mockDB, mock.Anything).Once().Return(timesheetUpdatedEntity, nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:        "happy case submit timesheet record has lesson hour completed",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: nil,
			reqString:   draftID,
			setup: func(ctx context.Context) {
				timesheetInDraft.TimesheetStatus = database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetLessonHourSingleRecordCompleted, nil)
				mockLessonRepo.On("FindLessonsByIDs", ctx, mockDB, mock.Anything).Once().Return(lessonSingleRecordCompleted, nil)
				mockTimesheetRepo.On("UpdateTimeSheet", ctx, mockDB, mock.Anything).Once().Return(timesheetUpdatedEntity, nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:        "happy case submit timesheet record have multiple lesson hours completed",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: nil,
			reqString:   draftID,
			setup: func(ctx context.Context) {
				timesheetInDraft.TimesheetStatus = database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetLessonHourMultiRecordCompleted, nil)
				mockLessonRepo.On("FindLessonsByIDs", ctx, mockDB, mock.Anything).Once().Return(lessonMultiRecordCompleted, nil)
				mockTimesheetRepo.On("UpdateTimeSheet", ctx, mockDB, mock.Anything).Once().Return(timesheetUpdatedEntity, nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:        "happy case submit timesheet record have multiple lesson hours completed and cancelled",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: nil,
			reqString:   draftID,
			setup: func(ctx context.Context) {
				timesheetInDraft.TimesheetStatus = database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetLessonHourMultiRecordCompletedCancelled, nil)
				mockLessonRepo.On("FindLessonsByIDs", ctx, mockDB, mock.Anything).Once().Return(lessonMultiRecordCompletedCancelled, nil)
				mockTimesheetRepo.On("UpdateTimeSheet", ctx, mockDB, mock.Anything).Once().Return(timesheetUpdatedEntity, nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:        "failed submit timesheet record has lesson status published",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.FailedPrecondition, invalidLessonStatusErr),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				timesheetInDraft.TimesheetStatus = database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetLessonHourSingleRecordPublished, nil)
				mockLessonRepo.On("FindLessonsByIDs", ctx, mockDB, mock.Anything).Once().Return(lessonSingleRecordPublished, nil)
			},
		},
		{
			name:        "failed submit timesheet record have multi lesson status published",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.FailedPrecondition, invalidLessonStatusErr),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetLessonHourMultiRecordPublished, nil)
				mockLessonRepo.On("FindLessonsByIDs", ctx, mockDB, mock.Anything).Once().Return(lessonMultiRecordPublished, nil)
			},
		},
		{
			name:        "failed submit timesheet record have multi lesson status published cancelled",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.FailedPrecondition, invalidLessonStatusErr),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetLessonHourMultiRecordPublishedCancelled, nil)
				mockLessonRepo.On("FindLessonsByIDs", ctx, mockDB, mock.Anything).Once().Return(lessonMultiRecordPublishedCancelled, nil)
			},
		},
		{
			name:        "failed submit timesheet record have multi lesson status published completed",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.FailedPrecondition, invalidLessonStatusErr),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetLessonHourMultiRecordPublishedCompleted, nil)
				mockLessonRepo.On("FindLessonsByIDs", ctx, mockDB, mock.Anything).Once().Return(lessonMultiRecordPublishedCompleted, nil)
			},
		},
		{
			name:        "error case unauthorized to submit timesheet",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDFail),
			reqString:   draftID,
			expectedErr: status.Error(codes.PermissionDenied, fmt.Sprintf("unauthorized to modify timesheet, timesheetStaffID: %s", CreateTimesheetStaffID)),
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
			},
		},
		{
			name:        "find timesheet failed tx is closed",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDFail),
			expectedErr: status.Error(codes.Internal, "find timesheet error: tx is closed"),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "find timesheet failed no rows result",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDFail),
			expectedErr: status.Error(codes.Internal, "find timesheet error: no rows in result set"),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:        "failed timesheet status in submitted",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.FailedPrecondition, timesheetShouldBeInDraftErr),
			reqString:   submittedID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInSubmitted, nil)
			},
		},
		{
			name:        "failed timesheet date not today",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.FailedPrecondition, "timesheet date should not be in future"),
			reqString:   submittedID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetDateTomorrow, nil)
			},
		},
		{
			name:        "failed timesheet status in approved",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.FailedPrecondition, timesheetShouldBeInDraftErr),
			reqString:   approvedID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInApproved, nil)
			},
		},
		{
			name:        "failed timesheet status in confirmed",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.FailedPrecondition, timesheetShouldBeInDraftErr),
			reqString:   confirmedID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInConfirmed, nil)
			},
		},
		{
			name:        "find timesheet lesson hour records tx is closed",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.Internal, "find timesheet lesson hours error: tx is closed"),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "find timesheet lesson hour records no rows result",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.Internal, "find timesheet lesson hours error: no rows in result set"),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:        "failed submit timesheet record no lesson hours tx is closed",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.Internal, submitStatusTxErr),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return([]*entity.TimesheetLessonHours{}, nil)
				mockTimesheetRepo.On("UpdateTimeSheet", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "failed find lessons tx is closed",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.Internal, "find lesson record error: tx is closed"),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				timesheetInDraft.TimesheetStatus = database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetLessonHourMultiRecordCompleted, nil)
				mockLessonRepo.On("FindLessonsByIDs", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "failed submit timesheet record has single lesson hour tx is closed",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.Internal, submitStatusTxErr),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetLessonHourSingleRecordCancelled, nil)
				mockLessonRepo.On("FindLessonsByIDs", ctx, mockDB, mock.Anything).Once().Return(lessonSingleRecordCancelled, nil)
				mockTimesheetRepo.On("UpdateTimeSheet", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "failed submit timesheet record have lesson hours tx is closed",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.Internal, submitStatusTxErr),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				timesheetInDraft.TimesheetStatus = database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetLessonHourMultiRecordCompleted, nil)
				mockLessonRepo.On("FindLessonsByIDs", ctx, mockDB, mock.Anything).Once().Return(lessonMultiRecordCompleted, nil)
				mockTimesheetRepo.On("UpdateTimeSheet", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "error case submit timesheet publish action log event failed",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.Internal, "PublishActionLogTimesheetEvent JSM.PublishAsyncContext failed, msgID: MsgID, Error"),
			reqString:   draftID,
			setup: func(ctx context.Context) {
				timesheetInDraft.TimesheetStatus = database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return([]*entity.TimesheetLessonHours{}, nil)
				mockTimesheetRepo.On("UpdateTimeSheet", ctx, mockDB, mock.Anything).Once().Return(timesheetUpdatedEntity, nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("MsgID", fmt.Errorf("Error"))
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			err := s.SubmitTimesheet(testCase.ctx, testCase.reqString)
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockTimesheetRepo, mockTimesheetLessonHourRepo, mockLessonRepo)
		})
	}
}

func TestTimesheetService_ApproveTimesheet(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	mockDB := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)
	mockTimesheetRepo := new(mock_repositories.MockTimesheetRepoImpl)
	mockTimesheetLessonHourRepo := new(mock_repositories.MockTimesheetLessonHoursRepoImpl)
	mockJsm := new(mock_nats.JetStreamManagement)
	mockLessonRepo := new(mock_repositories.MockLessonRepoImpl)

	s := TimesheetStateMachineService{
		DB:                       mockDB,
		JSM:                      mockJsm,
		TimesheetRepo:            mockTimesheetRepo,
		TimesheetLessonHoursRepo: mockTimesheetLessonHourRepo,
		LessonRepo:               mockLessonRepo,
	}

	singleSubmittedTimesheetRecord := &entity.Timesheet{
		TimesheetID:     database.Text("1"),
		TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String()),
	}

	timesheetSubmittedLessonHourSingle := []*entity.TimesheetLessonHours{
		{
			TimesheetID: singleSubmittedTimesheetRecord.TimesheetID,
			LessonID:    database.Text("3"),
		},
	}

	timesheetSubmittedLessonHourMulti := []*entity.TimesheetLessonHours{
		{
			TimesheetID: singleSubmittedTimesheetRecord.TimesheetID,
			LessonID:    database.Text("4"),
		},
		{
			TimesheetID: singleSubmittedTimesheetRecord.TimesheetID,
			LessonID:    database.Text("5"),
		},
		{
			TimesheetID: singleSubmittedTimesheetRecord.TimesheetID,
			LessonID:    database.Text("6"),
		},
		{
			TimesheetID: singleSubmittedTimesheetRecord.TimesheetID,
			LessonID:    database.Text("7"),
		},
	}

	multiSubmittedTimesheetRecord := []*entity.Timesheet{
		{
			TimesheetID:     database.Text("2"),
			TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String()),
		},
		{
			TimesheetID:     database.Text("3"),
			TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String()),
		},
		{
			TimesheetID:     database.Text("4"),
			TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String()),
		},
		{
			TimesheetID:     database.Text("5"),
			TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String()),
		},
	}

	multiTimesheetSubmittedSingleLessonHourEach := []*entity.TimesheetLessonHours{
		{
			TimesheetID: multiSubmittedTimesheetRecord[0].TimesheetID,
			LessonID:    database.Text("4"),
		},
		{
			TimesheetID: multiSubmittedTimesheetRecord[1].TimesheetID,
			LessonID:    database.Text("5"),
		},
		{
			TimesheetID: multiSubmittedTimesheetRecord[2].TimesheetID,
			LessonID:    database.Text("6"),
		},
		{
			TimesheetID: multiSubmittedTimesheetRecord[3].TimesheetID,
			LessonID:    database.Text("7"),
		},
	}

	multiTimesheetSubmittedMultiLessonHourEach := []*entity.TimesheetLessonHours{
		{
			TimesheetID: multiSubmittedTimesheetRecord[0].TimesheetID,
			LessonID:    database.Text("8"),
		},
		{
			TimesheetID: multiSubmittedTimesheetRecord[0].TimesheetID,
			LessonID:    database.Text("9"),
		},
		{
			TimesheetID: multiSubmittedTimesheetRecord[0].TimesheetID,
			LessonID:    database.Text("10"),
		},
		{
			TimesheetID: multiSubmittedTimesheetRecord[1].TimesheetID,
			LessonID:    database.Text("11"),
		},
		{
			TimesheetID: multiSubmittedTimesheetRecord[1].TimesheetID,
			LessonID:    database.Text("12"),
		},
		{
			TimesheetID: multiSubmittedTimesheetRecord[1].TimesheetID,
			LessonID:    database.Text("13"),
		},
		{
			TimesheetID: multiSubmittedTimesheetRecord[2].TimesheetID,
			LessonID:    database.Text("14"),
		},
		{
			TimesheetID: multiSubmittedTimesheetRecord[2].TimesheetID,
			LessonID:    database.Text("15"),
		},
		{
			TimesheetID: multiSubmittedTimesheetRecord[2].TimesheetID,
			LessonID:    database.Text("16"),
		},
		{
			TimesheetID: multiSubmittedTimesheetRecord[2].TimesheetID,
			LessonID:    database.Text("17"),
		},
		{
			TimesheetID: multiSubmittedTimesheetRecord[3].TimesheetID,
			LessonID:    database.Text("18"),
		},
		{
			TimesheetID: multiSubmittedTimesheetRecord[3].TimesheetID,
			LessonID:    database.Text("19"),
		},
		{
			TimesheetID: multiSubmittedTimesheetRecord[3].TimesheetID,
			LessonID:    database.Text("20"),
		},
	}

	listCompletedLesson := []*entity.Lesson{
		{
			LessonID:         database.Text("1"),
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED.String()),
		},
		{
			LessonID:         database.Text("2"),
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED.String()),
		},
	}

	listCanceledLesson := []*entity.Lesson{
		{
			LessonID:         database.Text("1"),
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED.String()),
		},
		{
			LessonID:         database.Text("2"),
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED.String()),
		},
	}

	listPublishedLesson := []*entity.Lesson{
		{
			LessonID:         database.Text("1"),
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED.String()),
		},
		{
			LessonID:         database.Text("2"),
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED.String()),
		},
	}

	listDraftLesson := []*entity.Lesson{
		{
			LessonID:         database.Text("1"),
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_DRAFT.String()),
		},
		{
			LessonID:         database.Text("2"),
			SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_DRAFT.String()),
		},
	}

	testcases := []TestCase{
		{
			name: "happy case approve single timesheet record with single lesson",

			ctx:             interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr:     nil,
			reqTimesheetIDs: []string{"1"},
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetIDsAndStatus", ctx, mockDB, mock.Anything, mock.Anything).Once().Return([]*entity.Timesheet{singleSubmittedTimesheetRecord}, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return([]*entity.TimesheetLessonHours{}, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTimesheetRepo.On("UpdateTimesheetStatusMultiple", ctx, mockTx, mock.Anything, mock.Anything).Return(nil).Once()
				mockTimesheetLessonHourRepo.On("FindByTimesheetIDs", ctx, mockTx, mock.Anything).Return(timesheetSubmittedLessonHourSingle, nil).Once()
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetLesson.Locked", mock.Anything, mock.Anything).Once().Return("", nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:            "happy case approve single timesheet record with multiple lesson",
			ctx:             interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr:     nil,
			reqTimesheetIDs: []string{"1"},
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetIDsAndStatus", ctx, mockDB, mock.Anything, mock.Anything).Once().Return([]*entity.Timesheet{singleSubmittedTimesheetRecord}, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return([]*entity.TimesheetLessonHours{}, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTimesheetRepo.On("UpdateTimesheetStatusMultiple", ctx, mockTx, mock.Anything, mock.Anything).Return(nil).Once()
				mockTimesheetLessonHourRepo.On("FindByTimesheetIDs", ctx, mockTx, mock.Anything).Return(timesheetSubmittedLessonHourMulti, nil).Once()
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetLesson.Locked", mock.Anything, mock.Anything).Once().Return("", nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:            "happy case approve multiple timesheet record with single lesson each",
			ctx:             interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr:     nil,
			reqTimesheetIDs: []string{"2", "3", "4", "5"},
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetIDsAndStatus", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(multiSubmittedTimesheetRecord, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Times(len(multiSubmittedTimesheetRecord)).Return([]*entity.TimesheetLessonHours{}, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTimesheetRepo.On("UpdateTimesheetStatusMultiple", ctx, mockTx, mock.Anything, mock.Anything).Return(nil).Once()
				mockTimesheetLessonHourRepo.On("FindByTimesheetIDs", ctx, mockTx, mock.Anything).Return(multiTimesheetSubmittedSingleLessonHourEach, nil).Once()
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetLesson.Locked", mock.Anything, mock.Anything).Once().Return("", nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:            "happy case approve multiple timesheet record with multi lesson each",
			ctx:             interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr:     nil,
			reqTimesheetIDs: []string{"2", "3", "4", "5"},
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetIDsAndStatus", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(multiSubmittedTimesheetRecord, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Times(len(multiSubmittedTimesheetRecord)).Return([]*entity.TimesheetLessonHours{}, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTimesheetRepo.On("UpdateTimesheetStatusMultiple", ctx, mockTx, mock.Anything, mock.Anything).Return(nil).Once()
				mockTimesheetLessonHourRepo.On("FindByTimesheetIDs", ctx, mockTx, mock.Anything).Return(multiTimesheetSubmittedMultiLessonHourEach, nil).Once()
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetLesson.Locked", mock.Anything, mock.Anything).Once().Return("", nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},

		{
			name:            "happy case approve timesheet with list completed lesson",
			ctx:             interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr:     nil,
			reqTimesheetIDs: []string{"1"},
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetIDsAndStatus", ctx, mockDB, mock.Anything, mock.Anything).Once().Return([]*entity.Timesheet{singleSubmittedTimesheetRecord}, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(multiTimesheetSubmittedSingleLessonHourEach, nil)
				mockLessonRepo.On("FindLessonsByIDs", ctx, mockDB, mock.Anything).Once().Return(listCompletedLesson, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTimesheetRepo.On("UpdateTimesheetStatusMultiple", ctx, mockTx, mock.Anything, mock.Anything).Return(nil).Once()
				mockTimesheetLessonHourRepo.On("FindByTimesheetIDs", ctx, mockTx, mock.Anything).Return(timesheetSubmittedLessonHourSingle, nil).Once()
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetLesson.Locked", mock.Anything, mock.Anything).Once().Return("", nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},

		{
			name:            "happy case approve timesheet with list canceled lesson",
			ctx:             interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr:     nil,
			reqTimesheetIDs: []string{"1"},
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetIDsAndStatus", ctx, mockDB, mock.Anything, mock.Anything).Once().Return([]*entity.Timesheet{singleSubmittedTimesheetRecord}, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(multiTimesheetSubmittedSingleLessonHourEach, nil)
				mockLessonRepo.On("FindLessonsByIDs", ctx, mockDB, mock.Anything).Once().Return(listCanceledLesson, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTimesheetRepo.On("UpdateTimesheetStatusMultiple", ctx, mockTx, mock.Anything, mock.Anything).Return(nil).Once()
				mockTimesheetLessonHourRepo.On("FindByTimesheetIDs", ctx, mockTx, mock.Anything).Return(timesheetSubmittedLessonHourSingle, nil).Once()
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetLesson.Locked", mock.Anything, mock.Anything).Once().Return("", nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},

		{
			name:            "failed approve timesheet with list published lesson",
			ctx:             interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr:     status.Error(codes.FailedPrecondition, "invalid lesson status"),
			reqTimesheetIDs: []string{"1"},
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetIDsAndStatus", ctx, mockDB, mock.Anything, mock.Anything).Once().Return([]*entity.Timesheet{singleSubmittedTimesheetRecord}, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(multiTimesheetSubmittedSingleLessonHourEach, nil)
				mockLessonRepo.On("FindLessonsByIDs", ctx, mockDB, mock.Anything).Once().Return(listPublishedLesson, nil)
			},
		},

		{
			name:            "failed approve timesheet with list draft lesson",
			ctx:             interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr:     status.Error(codes.FailedPrecondition, "invalid lesson status"),
			reqTimesheetIDs: []string{"1"},
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetIDsAndStatus", ctx, mockDB, mock.Anything, mock.Anything).Once().Return([]*entity.Timesheet{singleSubmittedTimesheetRecord}, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(multiTimesheetSubmittedSingleLessonHourEach, nil)
				mockLessonRepo.On("FindLessonsByIDs", ctx, mockDB, mock.Anything).Once().Return(listDraftLesson, nil)
			},
		},

		{
			name:            "failed find submitted timesheet tx closed",
			ctx:             interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr:     status.Error(codes.Internal, "find submitted timesheet error: tx is closed"),
			reqTimesheetIDs: []string{"1"},
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetIDsAndStatus", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:            "failed permission denied not user group school admin",
			ctx:             interceptors.ContextWithUserID(ctx, "teacher-id"),
			expectedErr:     status.Error(codes.PermissionDenied, "unauthorized to modify timesheet, timesheetStaffID: teacher-id"),
			reqTimesheetIDs: []string{"1"},
			setup: func(ctx context.Context) {
				// no setup, do nothing
			},
		},
		{
			name:            "failed upsert multiple tx closed",
			ctx:             interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr:     status.Error(codes.Internal, "upsert multiple timesheet error: tx is closed"),
			reqTimesheetIDs: []string{"2", "3", "4", "5"},
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetIDsAndStatus", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(multiSubmittedTimesheetRecord, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Times(len(multiSubmittedTimesheetRecord)).Return([]*entity.TimesheetLessonHours{}, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTimesheetRepo.On("UpdateTimesheetStatusMultiple", ctx, mockTx, mock.Anything, mock.Anything).Return(pgx.ErrTxClosed).Once()
				mockTx.On("Rollback", mock.Anything).Return(nil)
			},
		},
		{
			name:            "failed find submitted timesheet no rows",
			ctx:             interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr:     status.Error(codes.Internal, "find submitted timesheet error: no rows in result set"),
			reqTimesheetIDs: []string{"1"},
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetIDsAndStatus", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:            "failed find timesheet lessons no rows",
			ctx:             interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr:     status.Error(codes.Internal, "find by timesheet ids lesson hours error: no rows in result set"),
			reqTimesheetIDs: []string{"2", "3", "4", "5"},
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetIDsAndStatus", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(multiSubmittedTimesheetRecord, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Times(len(multiSubmittedTimesheetRecord)).Return([]*entity.TimesheetLessonHours{}, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTimesheetRepo.On("UpdateTimesheetStatusMultiple", ctx, mockTx, mock.Anything, mock.Anything).Return(nil).Once()
				mockTimesheetLessonHourRepo.On("FindByTimesheetIDs", ctx, mockTx, mock.Anything).Return(nil, pgx.ErrNoRows).Once()
				mockTx.On("Rollback", mock.Anything).Return(nil)
			},
		},
		{
			name:            "failed publish event lock lesson",
			ctx:             interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr:     status.Error(codes.Internal, "PublishLockLessonEvent JSM.PublishAsyncContext failed, msgID: test, publish error"),
			reqTimesheetIDs: []string{"2", "3", "4", "5"},
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetIDsAndStatus", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(multiSubmittedTimesheetRecord, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Times(len(multiSubmittedTimesheetRecord)).Return([]*entity.TimesheetLessonHours{}, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTimesheetRepo.On("UpdateTimesheetStatusMultiple", ctx, mockTx, mock.Anything, mock.Anything).Return(nil).Once()
				mockTimesheetLessonHourRepo.On("FindByTimesheetIDs", ctx, mockTx, mock.Anything).Return(multiTimesheetSubmittedMultiLessonHourEach, nil).Once()
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetLesson.Locked", mock.Anything, mock.Anything).Once().Return("test", errors.New("publish error"))
				mockTx.On("Rollback", mock.Anything).Return(nil)
			},
		},
		{
			name:            "failed to find submitted timesheet match with request count",
			ctx:             interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr:     status.Error(codes.Internal, "find submitted timesheet records not match with the request"),
			reqTimesheetIDs: []string{"2", "3", "4"},
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetIDsAndStatus", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(multiSubmittedTimesheetRecord, nil)
			},
		},
		{
			name:            "error case publish timesheet action log event",
			ctx:             interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr:     status.Error(codes.Internal, "PublishActionLogTimesheetEvent JSM.PublishAsyncContext failed, msgID: MsgID, Error"),
			reqTimesheetIDs: []string{"2", "3", "4", "5"},
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetIDsAndStatus", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(multiSubmittedTimesheetRecord, nil)
				mockTimesheetLessonHourRepo.On("FindTimesheetLessonHoursByTimesheetID", ctx, mockDB, mock.Anything).Times(len(multiSubmittedTimesheetRecord)).Return([]*entity.TimesheetLessonHours{}, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTimesheetRepo.On("UpdateTimesheetStatusMultiple", ctx, mockTx, mock.Anything, mock.Anything).Return(nil).Once()
				mockTimesheetLessonHourRepo.On("FindByTimesheetIDs", ctx, mockTx, mock.Anything).Return(multiTimesheetSubmittedMultiLessonHourEach, nil).Once()
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetLesson.Locked", mock.Anything, mock.Anything).Once().Return("", nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("MsgID", fmt.Errorf("Error"))
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			err := s.ApproveTimesheet(testCase.ctx, testCase.reqTimesheetIDs)
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockTimesheetRepo, mockTimesheetLessonHourRepo, mockJsm)
		})
	}
}

func TestTimesheetService_CancelApproveTimesheet(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	mockDB := new(mock_database.Ext)
	mockJsm := new(mock_nats.JetStreamManagement)
	mockTimesheetRepo := new(mock_repositories.MockTimesheetRepoImpl)

	s := TimesheetStateMachineService{
		DB:            mockDB,
		JSM:           mockJsm,
		TimesheetRepo: mockTimesheetRepo,
	}

	approvedTimesheetRecord := &entity.Timesheet{
		TimesheetID:     database.Text("1"),
		TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String()),
	}

	submittedTimesheetRecord := &entity.Timesheet{
		TimesheetID:     database.Text("1"),
		TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String()),
	}

	confirmedTimesheetRecord := &entity.Timesheet{
		TimesheetID:     database.Text("1"),
		TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_CONFIRMED.String()),
	}

	draftTimesheetRecord := &entity.Timesheet{
		TimesheetID:     database.Text("1"),
		TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String()),
	}

	testcases := []TestCase{
		{
			name:        "happy case cancel approve timesheet",
			ctx:         interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr: nil,
			reqString:   approvedTimesheetRecord.TimesheetID.String,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(approvedTimesheetRecord, nil)
				mockTimesheetRepo.On("UpdateTimeSheet", ctx, mockDB, mock.Anything).Once().Return(submittedTimesheetRecord, nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:        "invalid timesheet status submit",
			ctx:         interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr: status.Error(codes.FailedPrecondition, "timesheet status should be in Approved"),
			reqString:   submittedTimesheetRecord.TimesheetID.String,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(submittedTimesheetRecord, nil)
			},
		},
		{
			name:        "invalid timesheet status confirmed",
			ctx:         interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr: status.Error(codes.FailedPrecondition, "timesheet status should be in Approved"),
			reqString:   confirmedTimesheetRecord.TimesheetID.String,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(confirmedTimesheetRecord, nil)
			},
		},
		{
			name:        "invalid timesheet status draft",
			ctx:         interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr: status.Error(codes.FailedPrecondition, "timesheet status should be in Approved"),
			reqString:   draftTimesheetRecord.TimesheetID.String,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(draftTimesheetRecord, nil)
			},
		},
		{
			name:        "cannot find timesheet record no rows",
			ctx:         interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr: status.Error(codes.Internal, "find timesheet error: no rows in result set"),
			reqString:   "5222",
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:        "failed cancel approve timesheet",
			ctx:         interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr: status.Error(codes.Internal, "update timesheet status submitted error: tx is closed"),
			reqString:   approvedTimesheetRecord.TimesheetID.String,
			setup: func(ctx context.Context) {
				approvedTimesheetRecord.TimesheetStatus = database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String())
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(approvedTimesheetRecord, nil)
				mockTimesheetRepo.On("UpdateTimeSheet", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			err := s.CancelApproveTimesheet(testCase.ctx, testCase.reqString)
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockTimesheetRepo)
		})
	}
}

func TestTimesheetService_ConfirmTimesheet(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	mockDB := new(mock_database.Ext)
	mockTx := new(mock_database.Tx)
	mockJsm := new(mock_nats.JetStreamManagement)
	mockTimesheetRepo := new(mock_repositories.MockTimesheetRepoImpl)
	mockTimesheetLessonHourRepo := new(mock_repositories.MockTimesheetLessonHoursRepoImpl)

	s := TimesheetStateMachineService{
		DB:                       mockDB,
		JSM:                      mockJsm,
		TimesheetRepo:            mockTimesheetRepo,
		TimesheetLessonHoursRepo: mockTimesheetLessonHourRepo,
	}

	singleApprovedTimesheetRecord := &entity.Timesheet{
		TimesheetID:     database.Text("1"),
		TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String()),
	}

	multiApprovedTimesheetRecord := []*entity.Timesheet{
		{
			TimesheetID:     database.Text("2"),
			TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String()),
		},
		{
			TimesheetID:     database.Text("3"),
			TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String()),
		},
		{
			TimesheetID:     database.Text("4"),
			TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String()),
		},
	}

	testcases := []TestCase{
		{
			name: "happy case confirm single timesheet record",

			ctx:             interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr:     nil,
			reqTimesheetIDs: []string{singleApprovedTimesheetRecord.TimesheetID.String},
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetIDsAndStatus", ctx, mockDB, mock.Anything, mock.Anything).Once().Return([]*entity.Timesheet{singleApprovedTimesheetRecord}, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTimesheetRepo.On("UpdateTimesheetStatusMultiple", ctx, mockTx, mock.Anything, mock.Anything).Return(nil).Once()
				mockTx.On("Commit", mock.Anything).Return(nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:            "happy case confirm multiple timesheet record",
			ctx:             interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr:     nil,
			reqTimesheetIDs: []string{multiApprovedTimesheetRecord[0].TimesheetID.String, multiApprovedTimesheetRecord[1].TimesheetID.String, multiApprovedTimesheetRecord[2].TimesheetID.String},
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetIDsAndStatus", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(multiApprovedTimesheetRecord, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTimesheetRepo.On("UpdateTimesheetStatusMultiple", ctx, mockTx, mock.Anything, mock.Anything).Return(nil).Once()
				mockTx.On("Commit", mock.Anything).Return(nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Times(len(multiApprovedTimesheetRecord)).Return("", nil)
			},
		},
		{
			name:            "failed find approve timesheet tx closed",
			ctx:             interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr:     status.Error(codes.Internal, "find approved timesheet error: tx is closed"),
			reqTimesheetIDs: []string{"1"},
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetIDsAndStatus", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:            "failed permission denied not user group school admin",
			ctx:             interceptors.ContextWithUserID(ctx, "teacher-id"),
			expectedErr:     status.Error(codes.PermissionDenied, "unauthorized to modify timesheet, timesheetStaffID: teacher-id"),
			reqTimesheetIDs: []string{"1"},
			setup: func(ctx context.Context) {
				// no setup, do nothing
			},
		},
		{
			name:            "failed find approved timesheet no rows",
			ctx:             interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr:     status.Error(codes.Internal, "find approved timesheet error: no rows in result set"),
			reqTimesheetIDs: []string{"1"},
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetIDsAndStatus", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:            "failed update timesheet status multiple tx closed",
			ctx:             interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr:     status.Error(codes.Internal, "upsert multiple timesheet status to confirm error: tx is closed"),
			reqTimesheetIDs: []string{"2", "3", "4"},
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetIDsAndStatus", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(multiApprovedTimesheetRecord, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTimesheetRepo.On("UpdateTimesheetStatusMultiple", ctx, mockTx, mock.Anything, mock.Anything).Return(pgx.ErrTxClosed).Once()
				mockTx.On("Rollback", mock.Anything).Return(nil)
			},
		},
		{
			name:            "failed to find approved timesheet match with request count",
			ctx:             interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			expectedErr:     status.Error(codes.Internal, "find approved timesheet records not match with the request"),
			reqTimesheetIDs: []string{"2", "3", "4", "5"},
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetIDsAndStatus", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(multiApprovedTimesheetRecord, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			err := s.ConfirmTimesheet(testCase.ctx, testCase.reqTimesheetIDs)
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockTimesheetRepo, mockTimesheetLessonHourRepo)
		})
	}
}

func TestTimesheetService_CancelSubmissionTimesheet(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockJsm := new(mock_nats.JetStreamManagement)
	mockTimesheetRepo := new(mock_repositories.MockTimesheetRepoImpl)
	mockTimesheetLessonHourRepo := new(mock_repositories.MockTimesheetLessonHoursRepoImpl)

	s := TimesheetStateMachineService{
		DB:                       mockDB,
		JSM:                      mockJsm,
		TimesheetRepo:            mockTimesheetRepo,
		TimesheetLessonHoursRepo: mockTimesheetLessonHourRepo,
	}

	timesheetInDraft := &entity.Timesheet{
		TimesheetID:     database.Text(draftID),
		CreatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		UpdatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		StaffID:         database.Text(CreateTimesheetStaffID),
		LocationID:      database.Text(CreateTimesheetLocationID),
		Remark:          database.Text(CreateTimeSheetRemark),
		TimesheetDate:   database.Timestamptz(CreateTimesheetTimesheetDate),
		TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String()),
	}

	timesheetInSubmitted := &entity.Timesheet{
		TimesheetID:     database.Text(submittedID),
		CreatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		UpdatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		StaffID:         database.Text(CreateTimesheetStaffID),
		LocationID:      database.Text(CreateTimesheetLocationID),
		Remark:          database.Text(CreateTimeSheetRemark),
		TimesheetDate:   database.Timestamptz(CreateTimesheetTimesheetDate),
		TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String()),
	}

	timesheetInApproved := &entity.Timesheet{
		TimesheetID:     database.Text(approvedID),
		CreatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		UpdatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		StaffID:         database.Text(CreateTimesheetStaffID),
		LocationID:      database.Text(CreateTimesheetLocationID),
		Remark:          database.Text(CreateTimeSheetRemark),
		TimesheetDate:   database.Timestamptz(CreateTimesheetTimesheetDate),
		TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String()),
	}

	timesheetInConfirmed := &entity.Timesheet{
		TimesheetID:     database.Text(confirmedID),
		CreatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		UpdatedAt:       database.Timestamptz(CreateTimesheetDateNow),
		StaffID:         database.Text(CreateTimesheetStaffID),
		LocationID:      database.Text(CreateTimesheetLocationID),
		Remark:          database.Text(CreateTimeSheetRemark),
		TimesheetDate:   database.Timestamptz(CreateTimesheetTimesheetDate),
		TimesheetStatus: database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_CONFIRMED.String()),
	}

	testcases := []TestCase{
		{
			name:        "happy case cancel submit timesheet record",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: nil,
			reqString:   submittedID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInSubmitted, nil)
				mockTimesheetRepo.On("UpdateTimeSheet", ctx, mockDB, mock.Anything).Once().Return(timesheetInSubmitted, nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:        "failed to publish action log event",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.Internal, "PublishActionLogTimesheetEvent JSM.PublishAsyncContext failed, msgID: MsgID, Error"),
			reqString:   submittedID,
			setup: func(ctx context.Context) {
				timesheetInSubmitted.TimesheetStatus = database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String())
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInSubmitted, nil)
				mockTimesheetRepo.On("UpdateTimeSheet", ctx, mockDB, mock.Anything).Once().Return(timesheetInSubmitted, nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("MsgID", fmt.Errorf("Error"))
			},
		},
		{
			name:        "find timesheet failed no rows result",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDFail),
			expectedErr: status.Error(codes.Internal, "find timesheet error: no rows in result set"),
			reqString:   submittedID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:        "failed timesheet status in draff",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.FailedPrecondition, timesheetShouldBeInSubmittedErr),
			reqString:   submittedID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInDraft, nil)
			},
		},
		{
			name:        "failed timesheet status in approved",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.FailedPrecondition, timesheetShouldBeInSubmittedErr),
			reqString:   approvedID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInApproved, nil)
			},
		},
		{
			name:        "failed timesheet status in confirmed",
			ctx:         interceptors.ContextWithUserID(ctx, CreateTimesheetStaffID),
			expectedErr: status.Error(codes.FailedPrecondition, timesheetShouldBeInSubmittedErr),
			reqString:   confirmedID,
			setup: func(ctx context.Context) {
				mockTimesheetRepo.On("FindTimesheetByTimesheetID", ctx, mockDB, mock.Anything).Once().Return(timesheetInConfirmed, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			err := s.CancelSubmissionTimesheet(testCase.ctx, testCase.reqString)
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockTimesheetRepo, mockTimesheetLessonHourRepo)

		})
	}
}
