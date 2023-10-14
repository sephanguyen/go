package repositories

import (
	"context"
	"fmt"
	"testing"

	monitor_entities "github.com/manabie-com/backend/internal/eureka/entities/monitors"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func studyPlanMonitorRepoWithMock() (*StudyPlanMonitorRepo, *testutil.MockDB) {
	r := &StudyPlanMonitorRepo{}
	return r, testutil.NewMockDB()
}

func Test_RetrieveByFilter(t *testing.T) {
	t.Parallel()
	type TestCase struct {
		name         string
		intput1      *RetrieveFilter
		expectedResp interface{}
		expectedErr  error
		setup        func(ctx context.Context)
	}
	r, mockDB := studyPlanMonitorRepoWithMock()
	LCLTime := database.Text("10 mins")
	UCLTime := database.Text("2 mins")

	filter := &RetrieveFilter{
		StudyPlanMonitorType: database.Text(monitor_entities.StudyPlanMonitorType_STUDENT_STUDY_PLAN),
		IntervalTimeLCL:      &LCLTime,
		IntervalTimeULC:      &UCLTime,
	}
	testCases := []TestCase{
		{
			name:        "happy case",
			intput1:     filter,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				e := &monitor_entities.StudyPlanMonitor{}
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &filter.StudyPlanMonitorType, filter.IntervalTimeLCL, filter.IntervalTimeULC)
				fields, values := e.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
		{
			name:        "err query",
			intput1:     filter,
			expectedErr: fmt.Errorf("err db.Query: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				e := &monitor_entities.StudyPlanMonitor{}
				mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, &filter.StudyPlanMonitorType, filter.IntervalTimeLCL, filter.IntervalTimeULC)
				fields, values := e.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.setup(ctx)
			_, err := r.RetrieveByFilter(ctx, mockDB.DB, tc.intput1)
			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, tc.expectedErr, err)
			}
			if tc.expectedErr == nil {
				e := &monitor_entities.StudyPlanMonitor{}
				fields, _ := e.FieldMap()
				mockDB.RawStmt.AssertSelectedFields(t, fields...)
			}
		})
	}
}

func Test_SoftDelete(t *testing.T) {
	t.Parallel()

	type TestCase struct {
		name         string
		intput1      pgtype.TextArray
		expectedResp interface{}
		expectedErr  error
		setup        func(ctx context.Context)
	}

	r, mockDB := studyPlanMonitorRepoWithMock()
	monitorIDs := database.TextArray([]string{"m-1", "m-2"})
	testCases := []TestCase{
		{
			name:        "happy case",
			intput1:     monitorIDs,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &monitorIDs)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
		},
		{
			name:        "exec failed",
			intput1:     monitorIDs,
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &monitorIDs)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), pgx.ErrTxClosed, args...)
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.setup(ctx)
			err := r.SoftDelete(ctx, mockDB.DB, tc.intput1)
			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, tc.expectedErr, err)
			}
			if tc.expectedErr == nil {
				mockDB.RawStmt.AssertUpdatedTable(t, "study_plan_monitors")
				mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at")
			}
		})
	}
}

func Test_BulkUpsert(t *testing.T) {
	t.Parallel()
	type TestCase struct {
		name         string
		intput1      interface{}
		expectedResp interface{}
		expectedErr  error
		setup        func(ctx context.Context)
	}
	db := &mock_database.QueryExecer{}
	studyPlanMonitorRepo := &StudyPlanMonitorRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			intput1: []*monitor_entities.StudyPlanMonitor{
				{
					StudyPlanMonitorID: database.Text("mock-1"),
					StudentID:          database.Text("student-id"),
				},
				{
					StudyPlanMonitorID: database.Text("mock-2"),
					StudentID:          database.Text("student-id"),
				},
				{
					StudyPlanMonitorID: database.Text("mock-3"),
					StudentID:          database.Text("student-id"),
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "error send batch",
			intput1: []*monitor_entities.StudyPlanMonitor{
				{
					StudyPlanMonitorID: database.Text("mock-1"),
					StudentID:          database.Text("student-id"),
				},
				{
					StudyPlanMonitorID: database.Text("mock-2"),
					StudentID:          database.Text("student-id"),
				},
				{
					StudyPlanMonitorID: database.Text("mock-3"),
					StudentID:          database.Text("student-id"),
				},
			},
			expectedErr: fmt.Errorf("batchResults.Exec: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, pgx.ErrTxClosed)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			err := studyPlanMonitorRepo.BulkUpsert(ctx, db, testCase.intput1.([]*monitor_entities.StudyPlanMonitor))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func Test_SoftDeleteTypeStudyPlan(t *testing.T) {
	t.Parallel()

	type TestCase struct {
		name         string
		intput1      *RetrieveFilter
		expectedResp interface{}
		expectedErr  error
		setup        func(ctx context.Context)
	}

	r, mockDB := studyPlanMonitorRepoWithMock()
	LCLTime := database.Text("10 mins")
	UCLTime := database.Text("2 mins")

	filter := &RetrieveFilter{
		StudyPlanMonitorType: database.Text(monitor_entities.StudyPlanMonitorType_STUDENT_STUDY_PLAN),
		IntervalTimeLCL:      &LCLTime,
		IntervalTimeULC:      &UCLTime,
	}
	testCases := []TestCase{
		{
			name:        "happy case",
			intput1:     filter,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &filter.StudyPlanMonitorType, filter.IntervalTimeLCL, filter.IntervalTimeULC)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
		},
		{
			name:        "exec failed",
			intput1:     filter,
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &filter.StudyPlanMonitorType, filter.IntervalTimeLCL, filter.IntervalTimeULC)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), pgx.ErrTxClosed, args...)
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.setup(ctx)
			err := r.SoftDeleteTypeStudyPlan(ctx, mockDB.DB, tc.intput1)
			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, tc.expectedErr, err)
			}
			if tc.expectedErr == nil {
				mockDB.RawStmt.AssertUpdatedTable(t, "study_plan_monitors")
				mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at")
			}
		})
	}
}
