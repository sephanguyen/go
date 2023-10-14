package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// StudentEventLogsRepoWithSQLMock test repo with mock
func StudentEventLogsRepoWithSQLMock() (*StudentEventLogRepo, *testutil.MockDB) {
	r := &StudentEventLogRepo{}
	return r, testutil.NewMockDB()
}

func TestRepo_RetrieveStudentEventLogsByStudyPlanItemIDs(t *testing.T) {
	t.Parallel()
	r, mockDB := StudentEventLogsRepoWithSQLMock()
	ids := database.TextArray([]string{"ids"})
	e := &entities.StudentEventLog{}
	_ = e.ID.Set("id")

	testCases := []TestCase{
		{
			name: "err select",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
					mock.AnythingOfType("string"),
					&ids,
				)
			},
			expectedErr: fmt.Errorf("rows.Err: err db.Query: %w", puddle.ErrClosedPool),
		},
		{
			name: "success with select all fields",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything,
					mock.AnythingOfType("string"),
					&ids,
				)
				fields, values := e.FieldMap()
				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
			expectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := r.RetrieveStudentEventLogsByStudyPlanItemIDs(ctx, mockDB.DB, ids)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestRepo_LogsQuestionSubmitionByLO(t *testing.T) {
	t.Parallel()
	r, mockDB := StudentEventLogsRepoWithSQLMock()

	ids := database.TextArray([]string{"ids"})

	testCases := []TestCase{
		{
			name: "err select",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					&ids,
				)
			},
			expectedErr: fmt.Errorf("LogsQuestionSubmitionByLO.Query %w", puddle.ErrClosedPool),
		},
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					&ids,
				)
				mockDB.Rows.Mock.On("Next").Once().Return(true)
				mockDB.Rows.Mock.On("Next").Once().Return(false)
				mockDB.Rows.Mock.On("Scan", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockDB.Rows.Mock.On("Err").Once().Return(nil)
				mockDB.Rows.Mock.On("Close").Once().Return(nil)
			},
			expectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := r.LogsQuestionSubmitionByLO(ctx, mockDB.DB, "student-id", ids)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentEventLogs_DeleteByStudyPlanIdentities(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &StudentEventLogRepo{}

	studentEventLog := entities.StudentEventLog{
		ID:                 database.Int4(80),
		StudentID:          database.Text("student-id"),
		LearningMaterialID: database.Text("learningMaterialID"),
		StudyPlanID:        database.Text("study-plan-id"),
	}
	req := StudyPlanItemIdentity{
		StudentID:          studentEventLog.StudentID,
		StudyPlanID:        studentEventLog.StudyPlanID,
		LearningMaterialID: studentEventLog.LearningMaterialID,
	}
	query := fmt.Sprintf(`UPDATE %s SET deleted_at = now() WHERE learning_material_id = $1::TEXT AND student_id = $2::TEXT AND study_plan_id = $3::TEXT AND deleted_at IS NULL`, studentEventLog.TableName())

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", mock.Anything, query, req.LearningMaterialID, req.StudentID, req.StudyPlanID).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
			req:          req,
			expectedErr:  nil,
			expectedResp: int64(1),
		},
		{
			name: "no row",
			setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", mock.Anything, query, req.LearningMaterialID, req.StudentID, req.StudyPlanID).Once().Return(pgconn.CommandTag([]byte(`0`)), pgx.ErrNoRows)
			},
			req:          req,
			expectedErr:  fmt.Errorf("db.Exec: %w", pgx.ErrNoRows),
			expectedResp: int64(0),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.DeleteByStudyPlanIdentities(ctx, mockDB.DB, req)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}

// TODO:
func Test_Retrieve(t *testing.T) {

}
