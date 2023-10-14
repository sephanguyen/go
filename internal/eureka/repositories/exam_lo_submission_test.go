package repositories

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExamLOSubmission_List(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	e := &entities.ExamLOSubmission{}
	fields := strings.Join(database.GetFieldNames(e), ",els.")

	t.Run("list without course", func(t *testing.T) {
		filter := &ExamLOSubmissionFilter{}
		filter.CourseID.Set(nil)
		filter.Limit = 10

		m := mockDB{
			QueryFn: func(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
				expectedQuery := fmt.Sprintf(listExamLOSubmissionsStmtTpl, fields, examJoin, filter.Limit)
				if query != expectedQuery {
					t.Errorf("unexpected query: got: %v, want: %v", query, expectedQuery)
				}
				return nil, pgx.ErrNoRows
			},
		}

		r := &ExamLOSubmissionRepo{}
		r.List(ctx, m, filter)
	})
}

func TestExamLOSubmission_ListExamLOSubmissionWithDates(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ExamLOSubmissionRepo{}
	submissionID := database.Text("submissionID-1")
	result := ExtendedExamLOSubmission{}
	fields, values := result.FieldMap()
	fields = append(fields, "start_date", "end_date", "available_from", "available_to")
	values = append(values, &result.StartDate, &result.EndDate, &result.AvailableFrom, &result.AvailableTo)
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				mockDB.MockScanFields(nil, fields, values)
			},
			req:          nil,
			expectedResp: []*ExtendedExamLOSubmission{&result},
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				mockDB.MockScanFields(pgx.ErrNoRows, fields, values)
			},
			req:          submissionID,
			expectedResp: nil,
			expectedErr:  fmt.Errorf("ExamLOSubmissionRepo.ListExamLOSubmissionWithDates.Scan: %w", pgx.ErrNoRows),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.ListExamLOSubmissionWithDates(ctx, mockDB.DB, []*StudyPlanItemIdentity{
				{
					StudentID:          database.Text("student-id"),
					StudyPlanID:        database.Text("study-plan-id"),
					LearningMaterialID: database.Text("learning-material-id"),
				},
			})
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestExamLOSubmission_GetTotalGradedPoint(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ExamLOSubmissionRepo{}
	submissionID := database.Text("submissionID-1")
	var result pgtype.Int4

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, submissionID)
				mockDB.MockScanFields(nil, []string{"total_graded_point"}, []interface{}{&result})
			},
			req:          submissionID,
			expectedResp: result,
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, submissionID)
				mockDB.MockScanFields(pgx.ErrNoRows, []string{"total_graded_point"}, []interface{}{&result})
			},
			req:          submissionID,
			expectedResp: database.Int4(0),
			expectedErr:  fmt.Errorf("database.Select: %w", pgx.ErrNoRows),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.GetTotalGradedPoint(ctx, mockDB.DB, testCase.req.(pgtype.Text))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestExamLOSubmission_GetLatestSubmissionID(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ExamLOSubmissionRepo{}
	submissionID := database.Text("submissionID-1")
	result := pgtype.Text{String: "", Status: pgtype.Present}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, submissionID)
				mockDB.MockScanFields(nil, []string{"submission_id"}, []interface{}{&result})
			},
			req:          submissionID,
			expectedResp: result,
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, submissionID)
				mockDB.MockScanFields(pgx.ErrNoRows, []string{"submission_id"}, []interface{}{&result})
			},
			req:          submissionID,
			expectedResp: pgtype.Text{String: "", Status: pgtype.Undefined},
			expectedErr:  fmt.Errorf("database.Select: %w", pgx.ErrNoRows),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.GetLatestSubmissionID(ctx, mockDB.DB, testCase.req.(pgtype.Text))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestExamLOSubmission_Update(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ExamLOSubmissionRepo{}
	submission := &entities.ExamLOSubmission{}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
			req:         submission,
			expectedErr: nil,
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), fmt.Errorf("error execute query"), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
			req:         submission,
			expectedErr: fmt.Errorf("database.UpdateFields: %w", fmt.Errorf("error execute query")),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.Update(ctx, mockDB.DB, testCase.req.(*entities.ExamLOSubmission))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestExamLOSubmission_Get(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ExamLOSubmissionRepo{}
	reqArgs := &GetExamLOSubmissionArgs{
		SubmissionID:      database.Text("submissionID-1"),
		ShuffledQuizSetID: database.Text("shuffledQuizSetID-1"),
	}
	result := &entities.ExamLOSubmission{}
	fields, values := result.FieldMap()

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.Text("submissionID-1"), database.Text("shuffledQuizSetID-1"))
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			req:          reqArgs,
			expectedResp: result,
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.Text("submissionID-1"), database.Text("shuffledQuizSetID-1"))
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			req:          reqArgs,
			expectedResp: nil,
			expectedErr:  fmt.Errorf("database.Select: %w", pgx.ErrNoRows),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.Get(ctx, mockDB.DB, testCase.req.(*GetExamLOSubmissionArgs))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestExamLOSubmission_Delete(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ExamLOSubmissionRepo{}

	examLOSubmission := entities.ExamLOSubmission{
		SubmissionID:      database.Text("examLOSubmissionID-1"),
		ShuffledQuizSetID: database.Text("shuffled_quiz_set_1"),
	}
	query := fmt.Sprintf(deleteSubmissionQuery, examLOSubmission.TableName())
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", mock.Anything, query, examLOSubmission.SubmissionID).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
			req:          examLOSubmission.SubmissionID,
			expectedErr:  nil,
			expectedResp: int64(1),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.Delete(ctx, mockDB.DB, examLOSubmission.SubmissionID)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}

func TestExamLOSubmission_Insert(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ExamLOSubmissionRepo{}
	validExamLOSubmission := &entities.ExamLOSubmission{
		SubmissionID:       database.Text("submission_id"),
		StudentID:          database.Text("student_id"),
		StudyPlanID:        database.Text("study_plan_id"),
		LearningMaterialID: database.Text("learning_material_id"),
		ShuffledQuizSetID:  database.Text("shuffled_quiz_set_id"),
		TotalPoint:         database.Int4(10),
		BaseEntity: entities.BaseEntity{
			CreatedAt: database.Timestamptz(time.Now()),
			UpdatedAt: database.Timestamptz(time.Now()),
		},
	}
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				_, fields := validExamLOSubmission.FieldMap()
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
			req: validExamLOSubmission,
		},
		{
			name: "error execute query",
			setup: func(ctx context.Context) {
				_, fields := validExamLOSubmission.FieldMap()
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), fmt.Errorf("error execute query"), args...)
			},
			req:         validExamLOSubmission,
			expectedErr: fmt.Errorf("database.Insert: %w", fmt.Errorf("error execute query")),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.Insert(ctx, mockDB.DB, testCase.req.(*entities.ExamLOSubmission))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestExamLOSubmission_GetLatestExamLOSubmission(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ExamLOSubmissionRepo{}
	submissionID := database.Text("submissionID-1")
	result := entities.ExamLOSubmission{
		SubmissionID: submissionID,
	}
	fields, values := result.FieldMap()
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, submissionID)
				mockDB.MockScanFields(nil, fields, values)
			},
			req:          submissionID,
			expectedResp: result,
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, submissionID)
				mockDB.MockScanFields(pgx.ErrNoRows, fields, values)
			},
			req:          submissionID,
			expectedResp: entities.ExamLOSubmission{},
			expectedErr:  fmt.Errorf("database.Select: %w", pgx.ErrNoRows),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.GetLatestExamLOSubmission(ctx, mockDB.DB, testCase.req.(pgtype.Text))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestExamLOSubmission_BulkUpdateApproveReject(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ExamLOSubmissionRepo{}
	agrs := &BulkUpdateApproveRejectArgs{
		SubmissionIDs: database.TextArray([]string{"submissionID_1", "submissionID_2"}),
		Status:        database.Text("SUBMISSION_STATUS_RETURNED"),
		LastAction:    database.Text("APPROVE_ACTION_APPROVED"),
		LastActionAt:  database.Timestamptz(time.Now()),
		LastActionBy:  database.Text("user-id"),
		StatusCond:    database.TextArray([]string{"SUBMISSION_STATUS_RETURNED", "SUBMISSION_STATUS_MARKED"}),
		UpdatedAt:     database.Timestamptz(time.Now()),
	}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", mock.Anything, mock.Anything, agrs.UpdatedAt, agrs.Status, agrs.LastAction, agrs.LastActionAt, agrs.LastActionBy,
					agrs.SubmissionIDs, agrs.StatusCond).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
			req:          agrs,
			expectedErr:  nil,
			expectedResp: 1,
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", mock.Anything, mock.Anything, agrs.UpdatedAt, agrs.Status, agrs.LastAction, agrs.LastActionAt, agrs.LastActionBy,
					agrs.SubmissionIDs, agrs.StatusCond).Once().Return(pgconn.CommandTag([]byte(`0`)), fmt.Errorf("error execute query"))
			},
			req:          agrs,
			expectedErr:  fmt.Errorf("db.Exec: %w", fmt.Errorf("error execute query")),
			expectedResp: 0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.BulkUpdateApproveReject(ctx, mockDB.DB, testCase.req.(*BulkUpdateApproveRejectArgs))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestExamLOSubmission_GetInvalidIDsByBulkApproveReject(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ExamLOSubmissionRepo{}
	submissionIDs := database.TextArray([]string{"submissionID_1", "submissionID_2"})
	statusCond := database.TextArray([]string{"SUBMISSION_STATUS_RETURNED", "SUBMISSION_STATUS_MARKED"})
	var results pgtype.TextArray

	type Req struct {
		SubmissionIDs pgtype.TextArray
		StatusCond    pgtype.TextArray
	}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, submissionIDs, statusCond)
				mockDB.MockScanFields(nil, []string{"submission_id"}, []interface{}{&results})
			},
			req: &Req{
				SubmissionIDs: submissionIDs,
				StatusCond:    statusCond,
			},
			expectedResp: results,
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, submissionIDs, statusCond)
				mockDB.MockScanFields(pgx.ErrNoRows, []string{"submission_id"}, []interface{}{&results})
			},
			req: &Req{
				SubmissionIDs: submissionIDs,
				StatusCond:    statusCond,
			},
			expectedResp: nil,
			expectedErr:  fmt.Errorf("database.Select: %w", pgx.ErrNoRows),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			req := testCase.req.(*Req)
			resp, err := repo.GetInvalidIDsByBulkApproveReject(ctx, mockDB.DB, req.SubmissionIDs, req.StatusCond)
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestExamLOSubmission_UpdateExamSubmissionTotalPointsWithResult(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ExamLOSubmissionRepo{}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
			expectedErr: nil,
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(pgconn.CommandTag([]byte(`0`)), fmt.Errorf("error execute query"))
			},
			expectedErr: fmt.Errorf("db.Exec: %w", fmt.Errorf("error execute query")),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.UpdateExamSubmissionTotalPointsWithResult(ctx, mockDB.DB, database.Text("1"), database.Int4(0), database.Text("passed"))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestExamLOSubmission_UpdateExamSubmissionTotalPoints(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ExamLOSubmissionRepo{}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
			expectedErr: nil,
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(pgconn.CommandTag([]byte(`0`)), fmt.Errorf("error execute query"))
			},
			expectedErr: fmt.Errorf("db.Exec: %w", fmt.Errorf("error execute query")),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.UpdateExamSubmissionTotalPoints(ctx, mockDB.DB, database.Text("1"), database.Int4(0))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}
