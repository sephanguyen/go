package importmasterdata

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/timesheet/repository"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestImportAccountingCategory(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	tx := new(mock_database.Tx)
	timesheetConfigRepo := new(mock_repositories.MockTimesheetConfigRepoImpl)

	s := &ImportTimesheetConfigService{
		DB:                  db,
		TimesheetConfigRepo: timesheetConfigRepo,
	}

	testcases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: []byte(
				`timesheet_config_id,config_type,config_value,is_archived
				,0,office,0`),
			expectedErr:  nil,
			expectedResp: []*pb.ImportTimesheetConfigError{},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Return(tx, nil)
				timesheetConfigRepo.On("Create", ctx, tx, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Return(nil)
			},
		},
		{
			name:        "no data in csv file",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "no data in csv file"),
			req:         []byte{},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - mismatched number of fields in header and content",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, ""),
			req: []byte(
				`timesheet_config_id,config_type,config_value
				,0,office,0`),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - number of column != 4",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - number of column should be 4"),
			req: []byte(
				`timesheet_config_id,config_type,config_value
				0,office,0`),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - first column name (toLowerCase) != timesheet_config_id",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - first column (toLowerCase) should be 'timesheet_config_id'"),
			req: []byte(
				`Number,config_type,config_value,is_archived
				1,0,office,0`),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - second column name (toLowerCase) != config_type",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - second column (toLowerCase) should be 'config_type'"),
			req: []byte(
				`timesheet_config_id,Naming,config_value,is_archived
				1,0,office,0`),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - third column name (toLowerCase) != config_value",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - third column (toLowerCase) should be 'config_value'"),
			req: []byte(
				`timesheet_config_id,config_type,Description,is_archived
				1,0,office,0`),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - fourth column name (toLowerCase) != is_archived",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - fourth column (toLowerCase) should be 'is_archived'"),
			req: []byte(
				`timesheet_config_id,config_type,config_value,IsArchived
				1,0,office,0`),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "parsing valid file (with error lines in response)",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: []byte(
				`timesheet_config_id,config_type,config_value,is_archived
				,1,office 2,1
				,0,office 3,2
				,0,office 4,
				3,0,office 5,0
				,0,office 6,1
				4,0,office 7,0`),
			expectedResp: []*pb.ImportTimesheetConfigError{
				{
					RowNumber: 2,
					Error:     fmt.Sprintf("unable to parse timesheet config item: %s", fmt.Errorf("invalid config_type")),
				},
				{
					RowNumber: 3,
					Error:     fmt.Sprintf("unable to parse timesheet config item: %s", fmt.Errorf("error parsing is_archived")),
				},
				{
					RowNumber: 4,
					Error:     fmt.Sprintf("unable to parse timesheet config item: %s", fmt.Errorf("missing mandatory data: is_archived")),
				},
				{
					RowNumber: 6,
					Error:     fmt.Sprintf("unable to create new timesheet config item: %s", pgx.ErrTxClosed),
				},
				{
					RowNumber: 7,
					Error:     fmt.Sprintf("unable to update timesheet config item: %s", pgx.ErrTxClosed),
				},
			},
			setup: func(ctx context.Context) {
				timesheetConfigRepo.On("Update", ctx, tx, mock.Anything).Once().Return(nil)
				timesheetConfigRepo.On("Create", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				timesheetConfigRepo.On("Update", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			resp, err := s.ImportTimesheetConfig(testCase.ctx, testCase.req.([]byte))

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.expectedResp.([]*pb.ImportTimesheetConfigError)
				for i, err := range resp {
					assert.Equal(t, err.RowNumber, expectedResp[i].RowNumber)
					assert.Contains(t, err.Error, expectedResp[i].Error)
				}
			}

			mock.AssertExpectationsForObjects(t, db, timesheetConfigRepo)
		})
	}
}
