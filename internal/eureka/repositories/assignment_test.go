package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func assignmentRepoWithMock() (*AssignmentRepo, *testutil.MockDB) {
	r := &AssignmentRepo{}
	return r, testutil.NewMockDB()
}

func TestGetAssignmentSetting(t *testing.T) {
	t.Parallel()
	t.Run("error select", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		r, mockDB := assignmentRepoWithMock()
		id := database.Text("random-id")
		mockDB.MockQueryArgs(
			t, pgx.ErrNoRows, ctx, mock.AnythingOfType("string"), &id,
		)
		settings := &entities.AssignmentSetting{
			AllowLateSubmission:    true,
			AllowResubmission:      false,
			RequireAssignmentNote:  false,
			RequireAttachment:      true,
			RequireVideoSubmission: true,
		}
		var result pgtype.JSONB
		_ = result.Set(settings)
		mockDB.MockScanFields(nil, []string{"settings"}, []interface{}{&result})
		_, err := r.GetAssignmentSetting(ctx, mockDB.DB, id)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		mockDB.RawStmt.AssertSelectedFields(t, "settings")
	})
	t.Run("successful", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		r, mockDB := assignmentRepoWithMock()
		id := database.Text("random-id")
		mockDB.MockQueryArgs(
			t, nil, ctx, mock.AnythingOfType("string"), &id,
		)
		settings := &entities.AssignmentSetting{
			AllowLateSubmission:    true,
			AllowResubmission:      false,
			RequireAssignmentNote:  false,
			RequireAttachment:      true,
			RequireVideoSubmission: true,
		}
		var result pgtype.JSONB
		_ = result.Set(settings)
		mockDB.MockScanFields(nil, []string{"settings"}, []interface{}{&result})
		_, err := r.GetAssignmentSetting(ctx, mockDB.DB, id)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertSelectedFields(t, "settings")
	})
}

func TestAssignmentRepo_BulkUpsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	assignmentRepo := &AssignmentRepo{}
	validAssignmentReq := []*entities.Assignment{
		{
			ID: pgtype.Text{String: "1", Status: pgtype.Present},
		},
		{
			ID: pgtype.Text{String: "2", Status: pgtype.Present},
		},
		{
			ID: pgtype.Text{String: "3", Status: pgtype.Present},
		},
	}
	testCases := []TestCase{
		{
			name:        "happy case",
			req:         validAssignmentReq,
			expectedErr: nil,
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validAssignmentReq); i++ {
					_, field := validAssignmentReq[i].FieldMap()
					fields = append(fields, field...)
				}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("Exec", args...).Once().Return(cmdTag, nil)
			},
		},
		{
			name:        "exec error",
			req:         validAssignmentReq,
			expectedErr: fmt.Errorf("eureka_db.BulkUpsertAssignment error: exec error"),
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validAssignmentReq); i++ {
					_, field := validAssignmentReq[i].FieldMap()
					fields = append(fields, field...)
				}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("Exec", args...).Once().Return(cmdTag, fmt.Errorf("exec error"))
			},
		},
		{
			name:        "no row affected",
			req:         validAssignmentReq,
			expectedErr: fmt.Errorf("eureka_db.BulkUpsertAssignment error: no row affected"),
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validAssignmentReq); i++ {
					_, field := validAssignmentReq[i].FieldMap()
					fields = append(fields, field...)
				}
				cmdTag := pgconn.CommandTag([]byte(`0`))
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("Exec", args...).Once().Return(cmdTag, nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := assignmentRepo.BulkUpsert(ctx, db, testCase.req.([]*entities.Assignment))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestAssignmentRepo_IsStudentAssigned(t *testing.T) {
	t.Parallel()
	r, mockDB := assignmentRepoWithMock()
	validReq := []pgtype.Text{database.Text("study-plan-item-id"), database.Text("assignment-id"), database.Text("student-id")}
	testCases := []TestCase{
		{
			name:        "happy case",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &validReq[1], &validReq[2])

				var count pgtype.Int8
				mockDB.MockScanFields(nil, []string{"count"}, []interface{}{&count})
			},
		},
		{
			name:        "error no rows",
			req:         validReq,
			expectedErr: fmt.Errorf("err db.Query: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, &validReq[1], &validReq[2])

				var count pgtype.Int8
				mockDB.MockScanFields(nil, []string{"count"}, []interface{}{&count})
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := r.IsStudentAssigned(ctx, mockDB.DB, validReq[0], validReq[1], validReq[2])
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestAssignmentRepo_IsStudentAssignedV2(t *testing.T) {
	t.Parallel()
	r, mockDB := assignmentRepoWithMock()
	type Req struct {
		StudyPlanID pgtype.Text
		StudentID   pgtype.Text
	}
	validReq := &Req{
		StudyPlanID: database.Text("study_plan_id"),
		StudentID:   database.Text("student_id"),
	}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.Text("study_plan_id"), database.Text("student_id"))

				var isStudentAssigned pgtype.Bool
				mockDB.MockScanFields(nil, []string{"isStudentAssigned"}, []interface{}{&isStudentAssigned})
			},
		},
		{
			name:        "error no rows",
			req:         validReq,
			expectedErr: fmt.Errorf("err db.Query: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, database.Text("study_plan_id"), database.Text("student_id"))

				var isStudentAssigned pgtype.Bool
				mockDB.MockScanFields(nil, []string{"isStudentAssigned"}, []interface{}{&isStudentAssigned})
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.(*Req)
		_, err := r.IsStudentAssignedV2(ctx, mockDB.DB, req.StudyPlanID, req.StudentID)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestDeleteAssignment(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := assignmentRepoWithMock()

	ids := database.TextArray([]string{"id-1", "id-2", "id-3"})
	t.Run("err update", func(t *testing.T) {
		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, &ids)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		err := r.SoftDelete(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("no rows affected", func(t *testing.T) {
		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, &ids)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		err := r.SoftDelete(ctx, mockDB.DB, ids)
		assert.EqualError(t, err, fmt.Errorf("cannot delete assignments").Error())
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, &ids)
		mockDB.MockExecArgs(t, pgconn.CommandTag("3"), nil, args...)

		err := r.SoftDelete(ctx, mockDB.DB, ids)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertUpdatedTable(t, "assignments")
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"assignment_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestRetrieveAssignments(t *testing.T) {
	t.Parallel()
	r, mockDB := assignmentRepoWithMock()
	validReq := database.TextArray([]string{"valid-req-1", "valid-req-2"})
	testCases := []TestCase{
		{
			name:        "happy case",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				e := &entities.Assignment{}
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &validReq)
				fields, values := e.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
		{
			name:        "error no rows",
			req:         validReq,
			expectedErr: fmt.Errorf("err db.Query: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				e := &entities.Assignment{}
				mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, &validReq)
				fields, values := e.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		e := &entities.Assignment{}
		_, err := r.RetrieveAssignments(ctx, mockDB.DB, validReq)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"assignment_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"deleted_at":    {HasNullTest: true},
		})
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestAssignmentRepo_RetrieveAssignmentsByTopicIDs(t *testing.T) {
	t.Parallel()
	r, mockDB := assignmentRepoWithMock()
	validReq := database.TextArray([]string{"valid-req-1", "valid-req-2"})
	testCases := []TestCase{
		{
			name:        "happy case",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				e := &entities.Assignment{}
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &validReq)
				fields, values := e.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
		{
			name:        "error no rows",
			req:         validReq,
			expectedErr: fmt.Errorf("err db.Query: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				e := &entities.Assignment{}
				mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, &validReq)
				fields, values := e.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		e := &entities.Assignment{}
		_, err := r.RetrieveAssignmentsByTopicIDs(ctx, mockDB.DB, validReq)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestAssignmentRepo_UpdateDisplayOrders(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	assignmentRepo := &AssignmentRepo{}
	mDisplayOrder := map[pgtype.Text]pgtype.Int4{
		database.Text("lo-1"): database.Int4(1),
	}
	testCases := []TestCase{
		{
			name:        "happy case",
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
			name:        "error send batch",
			expectedErr: errors.New("batchResults.Exec: closed pool"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := assignmentRepo.UpdateDisplayOrders(ctx, db, mDisplayOrder)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}

	return
}

func TestCalculateHigestScore(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	rows := &mock_database.Rows{}
	repo := &AssignmentRepo{}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         database.TextArray([]string{"assignment_id"}),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything).Once().Return(nil)

				rows.On("Next").Once().Return(false)
			},
		},
		{
			name:        "query error",
			req:         database.TextArray([]string{"assignment_id"}),
			expectedErr: fmt.Errorf("AssignmentRepo.CalculateHigestScore.Query: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "scan error",
			req:         database.TextArray([]string{"assignment_id"}),
			expectedErr: fmt.Errorf("AssignmentRepo.CalculateHigestScore.Scan: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := repo.CalculateHigestScore(ctx, db, testCase.req.(pgtype.TextArray))
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestCalculateTaskAssignmentHighestScore(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	rows := &mock_database.Rows{}
	repo := &AssignmentRepo{}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         database.TextArray([]string{"assignment_id"}),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything).Once().Return(nil)

				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "query error",
			req:         database.TextArray([]string{"assignment_id"}),
			expectedErr: fmt.Errorf("AssignmentRepo.CalculateTaskAssignmentHighestScore.Query: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
				rows.On("Close").Once().Return(nil)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "scan error",
			req:         database.TextArray([]string{"assignment_id"}),
			expectedErr: fmt.Errorf("AssignmentRepo.CalculateTaskAssignmentHighestScore.Scan: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
				rows.On("Err").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := repo.CalculateTaskAssignmentHighestScore(ctx, db, testCase.req.(pgtype.TextArray))
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func Test_RetrieveBookAssignmentByIntervalTime(t *testing.T) {
	t.Parallel()
	type TestCase struct {
		name         string
		input1       pgtype.Text
		expectedResp interface{}
		expectedErr  error
		setup        func(ctx context.Context)
	}
	mockDB := testutil.NewMockDB()
	rows := mockDB.Rows
	r := &AssignmentRepo{}
	e := &entities.Assignment{}
	fields, _ := e.FieldMap()
	totalFields := append(fields, "book_id", "chapter_id", "topic_id")
	scanFields := database.GetScanFields(e, fields)
	var (
		bookID    pgtype.Text
		chapterID pgtype.Text
		topicID   pgtype.Text
	)
	scanFields = append(scanFields, &bookID, &chapterID, &topicID)

	intervalTime := database.Text("15 mins")
	testCases := []TestCase{
		{
			name:        "happy case",
			input1:      intervalTime,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, intervalTime)
				mockDB.DB.On("Query").Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", scanFields...).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "err query",
			input1:      intervalTime,
			expectedErr: fmt.Errorf("row.Err: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, intervalTime)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(pgx.ErrNoRows)
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.setup(ctx)
			_, err := r.RetrieveBookAssignmentByIntervalTime(ctx, mockDB.DB, tc.input1)
			assert.Equal(t, tc.expectedErr, err)
			mockDB.RawStmt.AssertSelectedFields(t, totalFields...)
		})
	}
}
