package timesheet

import (
	"context"
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

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStaffTransportationExpenseService_UpsertConfig(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var (
		staffTransportationExpenseRepo = new(mock_repositories.MockStaffTransportationExpenseRepoImpl)
		timesheetRepo                  = new(mock_repositories.MockTimesheetRepoImpl)
		transportationExpenseRepo      = new(mock_repositories.MockTransportationExpenseRepoImpl)
		db                             = new(mock_database.Ext)
		tx                             = new(mock_database.Tx)
		staffId                        = idutil.ULIDNow()
		mockJsm                        = new(mock_nats.JetStreamManagement)
	)

	s := StaffTransportationExpenseServiceImpl{
		DB:                             db,
		JSM:                            mockJsm,
		StaffTransportationExpenseRepo: staffTransportationExpenseRepo,
		TimesheetRepo:                  timesheetRepo,
		TransportationExpenseRepo:      transportationExpenseRepo,
	}

	staffTransportationExpenseDTO := &dto.StaffTransportationExpenses{
		ID: idutil.ULIDNow(),
	}

	ListStaffTransportationExpenseE := []*entity.StaffTransportationExpense{
		{
			ID: database.Text(idutil.ULIDNow()),
		},
	}

	ListTimesheetCanChange := []dto.TimesheetLocationDto{{TimesheetID: idutil.ULIDNow(), LocationID: "location_id_value"}}
	ListTimesheetCanChangeMultiple := []dto.TimesheetLocationDto{{TimesheetID: idutil.ULIDNow(), LocationID: "location_id_value"}, {TimesheetID: idutil.ULIDNow(), LocationID: "location_id_value-1"}}
	testCases := []TestCase{
		{
			name: "happy case upsert staff transportation expense config",
			ctx:  interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			req: &dto.ListStaffTransportationExpenses{
				staffTransportationExpenseDTO,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				staffTransportationExpenseRepo.On("FindListTransportExpensesByStaffIDs", ctx, db, mock.Anything).
					Return(ListStaffTransportationExpenseE, nil).Once()
				timesheetRepo.On("GetStaffFutureTimesheetIDsWithLocations", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(ListTimesheetCanChange, nil).Once()

				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				staffTransportationExpenseRepo.On("UpsertMultiple", ctx, tx, mock.Anything).
					Return(nil).Once()
				transportationExpenseRepo.On("SoftDeleteMultipleByTimesheetIDs", ctx, tx, mock.Anything).
					Return(nil).Once()
				transportationExpenseRepo.On("UpsertMultiple", ctx, tx, mock.Anything).
					Return(nil).Once()
				tx.On("Commit", mock.Anything).Return(nil)
			},
		},
		{
			name: "error case failed to retrieve staff transportation expenses",
			ctx:  interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			req: &dto.ListStaffTransportationExpenses{
				staffTransportationExpenseDTO,
			},
			expectedErr: fmt.Errorf("get list staff transport expenses error: %s", errors.New("internal error").Error()),
			setup: func(ctx context.Context) {
				staffTransportationExpenseRepo.On("FindListTransportExpensesByStaffIDs", ctx, db, mock.Anything).
					Return(nil, errors.New("internal error")).Once()
			},
		},
		{
			name: "error case failed to retrieve timesheet after date",
			ctx:  interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			req: &dto.ListStaffTransportationExpenses{
				staffTransportationExpenseDTO,
			},
			expectedErr: fmt.Errorf("get staff timesheet after date error: %s", errors.New("internal error").Error()),
			setup: func(ctx context.Context) {
				staffTransportationExpenseRepo.On("FindListTransportExpensesByStaffIDs", ctx, db, mock.Anything).
					Return(ListStaffTransportationExpenseE, nil).Once()
				timesheetRepo.On("GetStaffFutureTimesheetIDsWithLocations", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("internal error")).Once()
			},
		},
		{
			name: "error case upsert multiple staff transportation expense failed",
			ctx:  interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			req: &dto.ListStaffTransportationExpenses{
				staffTransportationExpenseDTO,
			},
			expectedErr: status.Error(codes.Internal, "internal error"),
			setup: func(ctx context.Context) {
				staffTransportationExpenseRepo.On("FindListTransportExpensesByStaffIDs", ctx, db, mock.Anything).
					Return(ListStaffTransportationExpenseE, nil).Once()
				timesheetRepo.On("GetStaffFutureTimesheetIDsWithLocations", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(ListTimesheetCanChange, nil).Once()

				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				staffTransportationExpenseRepo.On("UpsertMultiple", ctx, tx, mock.Anything).
					Return(errors.New("internal error")).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "error case soft delete multiple timesheet by ids failed",
			ctx:  interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			req: &dto.ListStaffTransportationExpenses{
				staffTransportationExpenseDTO,
			},
			expectedErr: status.Error(codes.Internal, "internal error"),
			setup: func(ctx context.Context) {
				staffTransportationExpenseRepo.On("FindListTransportExpensesByStaffIDs", ctx, db, mock.Anything).
					Return(ListStaffTransportationExpenseE, nil).Once()
				timesheetRepo.On("GetStaffFutureTimesheetIDsWithLocations", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(ListTimesheetCanChange, nil).Once()

				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				staffTransportationExpenseRepo.On("UpsertMultiple", ctx, tx, mock.Anything).
					Return(nil).Once()
				transportationExpenseRepo.On("SoftDeleteMultipleByTimesheetIDs", ctx, tx, mock.Anything).
					Return(errors.New("internal error")).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "error case upsert multiple transportation expenses failed",
			ctx:  interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			req: &dto.ListStaffTransportationExpenses{
				staffTransportationExpenseDTO,
			},
			expectedErr: status.Error(codes.Internal, "internal error"),
			setup: func(ctx context.Context) {
				staffTransportationExpenseRepo.On("FindListTransportExpensesByStaffIDs", ctx, db, mock.Anything).
					Return(ListStaffTransportationExpenseE, nil).Once()
				timesheetRepo.On("GetStaffFutureTimesheetIDsWithLocations", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(ListTimesheetCanChange, nil).Once()

				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				staffTransportationExpenseRepo.On("UpsertMultiple", ctx, tx, mock.Anything).
					Return(nil).Once()
				transportationExpenseRepo.On("SoftDeleteMultipleByTimesheetIDs", ctx, tx, mock.Anything).
					Return(nil).Once()
				transportationExpenseRepo.On("UpsertMultiple", ctx, tx, mock.Anything).
					Return(errors.New("internal error")).Once()
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "error case publish timesheet action log failed",
			ctx:  interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			req: &dto.ListStaffTransportationExpenses{
				&dto.StaffTransportationExpenses{
					ID:         idutil.ULIDNow(),
					LocationID: "location_id_value",
				},
			},
			expectedErr: status.Error(codes.Internal, "PublishActionLogTimesheetEvent JSM.PublishAsyncContext failed, msgID: MsgID, Error"),
			setup: func(ctx context.Context) {
				staffTransportationExpenseRepo.On("FindListTransportExpensesByStaffIDs", ctx, db, mock.Anything).
					Return(ListStaffTransportationExpenseE, nil).Once()
				timesheetRepo.On("GetStaffFutureTimesheetIDsWithLocations", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(ListTimesheetCanChange, nil).Once()

				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				staffTransportationExpenseRepo.On("UpsertMultiple", ctx, tx, mock.Anything).
					Return(nil).Once()
				transportationExpenseRepo.On("SoftDeleteMultipleByTimesheetIDs", ctx, tx, mock.Anything).
					Return(nil).Once()
				transportationExpenseRepo.On("UpsertMultiple", ctx, tx, mock.Anything).
					Return(nil).Once()
				tx.On("Commit", mock.Anything).Return(nil)
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Once().Return("MsgID", fmt.Errorf("Error"))
			},
		},
		{
			name: "case handle duplicate timesheet action log request",
			ctx:  interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			req: &dto.ListStaffTransportationExpenses{
				&dto.StaffTransportationExpenses{
					ID:         idutil.ULIDNow(),
					LocationID: "location_id_value",
				},
				&dto.StaffTransportationExpenses{
					ID:         idutil.ULIDNow(),
					LocationID: "location_id_value",
				},
				&dto.StaffTransportationExpenses{
					ID:         idutil.ULIDNow(),
					LocationID: "location_id_value",
				},
				&dto.StaffTransportationExpenses{
					ID:         idutil.ULIDNow(),
					LocationID: "location_id_value-1",
				},
				&dto.StaffTransportationExpenses{
					ID:         idutil.ULIDNow(),
					LocationID: "location_id_value-1",
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				staffTransportationExpenseRepo.On("FindListTransportExpensesByStaffIDs", ctx, db, mock.Anything).
					Return(ListStaffTransportationExpenseE, nil).Once()
				timesheetRepo.On("GetStaffFutureTimesheetIDsWithLocations", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(ListTimesheetCanChangeMultiple, nil).Once()

				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				staffTransportationExpenseRepo.On("UpsertMultiple", ctx, tx, mock.Anything).
					Return(nil).Once()
				transportationExpenseRepo.On("SoftDeleteMultipleByTimesheetIDs", ctx, tx, mock.Anything).
					Return(nil).Once()
				transportationExpenseRepo.On("UpsertMultiple", ctx, tx, mock.Anything).
					Return(nil).Once()
				tx.On("Commit", mock.Anything).Return(nil)
				// if this line throws an error, then duplicate timesheet action log request is not handled properly since we expect this to only be called based on the number of unique timesheet ids
				mockJsm.On("PublishAsyncContext", mock.Anything, "TimesheetActionLog.Created", mock.Anything, mock.Anything).Times(len(ListTimesheetCanChangeMultiple)).Return("", nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*dto.ListStaffTransportationExpenses)
			err := s.UpsertConfig(testCase.ctx, staffId, req)
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}
