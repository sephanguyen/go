package repositories

import (
	"context"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBulkUpsertAssignmentStudyPlanItem(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	assignmentStudyPlanItemRepo := &AssignmentStudyPlanItemRepo{}
	validAssignmentStudyPlanItemReq := []*entities.AssignmentStudyPlanItem{
		{
			AssignmentID:    database.Text("1"),
			StudyPlanItemID: database.Text("study-plan-item-1"),
		},
		{
			AssignmentID:    database.Text("2"),
			StudyPlanItemID: database.Text("study-plan-item-2"),
		},
		{
			AssignmentID:    database.Text("3"),
			StudyPlanItemID: database.Text("study-plan-item-3"),
		},
	}
	testCases := []TestCase{
		{
			name:        "happy case",
			req:         validAssignmentStudyPlanItemReq,
			expectedErr: nil,
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validAssignmentStudyPlanItemReq); i++ {
					_, field := validAssignmentStudyPlanItemReq[i].FieldMap()
					fields = append(fields, field...)
				}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("Exec", args...).Once().Return(cmdTag, nil)
			},
		},
		{
			name:        "exec error",
			req:         validAssignmentStudyPlanItemReq,
			expectedErr: fmt.Errorf("eureka_db.BulkUpsertAssignmentStudyPlanItem error: exec error"),
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validAssignmentStudyPlanItemReq); i++ {
					_, field := validAssignmentStudyPlanItemReq[i].FieldMap()
					fields = append(fields, field...)
				}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("Exec", args...).Once().Return(cmdTag, fmt.Errorf("exec error"))
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := assignmentStudyPlanItemRepo.BulkInsert(ctx, db, testCase.req.([]*entities.AssignmentStudyPlanItem))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestCopyFromStudyPlan(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	assignmentStudyPlanItemRepo := &AssignmentStudyPlanItemRepo{}
	validReq := database.TextArray([]string{"valid-req-1", "valid-req-2"})

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything).Once().Return(pgconn.CommandTag([]byte(`2`)), nil)
			},
		},
		{
			name:        "error send batch",
			req:         validReq,
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything).Once().Return(pgconn.CommandTag([]byte(`2`)), pgx.ErrTxClosed)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := assignmentStudyPlanItemRepo.CopyFromStudyPlan(ctx, db, testCase.req.(pgtype.TextArray))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func AssignmentStudyPlanItemRepoWithSqlMock() (*AssignmentStudyPlanItemRepo, *testutil.MockDB) {
	r := &AssignmentStudyPlanItemRepo{}
	return r, testutil.NewMockDB()
}

func TestFindByStudyPlanItemIDs(t *testing.T) {
	t.Parallel()
	r, mockDB := AssignmentStudyPlanItemRepoWithSqlMock()
	validReq := database.TextArray([]string{"valid-req-1", "valid-req-2"})

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				e := &entities.AssignmentStudyPlanItem{}
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
				e := &entities.AssignmentStudyPlanItem{}
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
		e := &entities.AssignmentStudyPlanItem{}
		_, err := r.FindByStudyPlanItemIDs(ctx, mockDB.DB, testCase.req.(pgtype.TextArray))
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestSoftDeleteByAssigmentIDs(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	rows := &mock_database.Rows{}

	assignmentStudyPlanItemRepo := &AssignmentStudyPlanItemRepo{}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         database.TextArray([]string{"assignment_id"}),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything).Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "query error",
			req:         database.TextArray([]string{"assignment_id"}),
			expectedErr: fmt.Errorf("db.Query: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
		},
		{
			name:        "error",
			req:         database.TextArray([]string{"assignment_id"}),
			expectedErr: fmt.Errorf("db.Error: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything).Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(fmt.Errorf("error"))
				rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := assignmentStudyPlanItemRepo.SoftDeleteByAssigmentIDs(ctx, db, testCase.req.(pgtype.TextArray))
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func Test_BulkEditAssignmentTime(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	assignmentStudyPlanItemRepo := &AssignmentStudyPlanItemRepo{}
	type Req struct {
		studentID pgtype.Text
		ens       []*entities.StudyPlanItem
	}

	testCases := []TestCase{
		{
			name: "happy case",
			req: &Req{
				studentID: database.Text("student-id"),
				ens: []*entities.StudyPlanItem{
					{
						ID: database.Text("study-plan-item-id"),
					},
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
			req: &Req{
				studentID: database.Text("student-id"),
				ens: []*entities.StudyPlanItem{
					{
						ID:        database.Text("study-plan-item-id"),
						StartDate: database.Timestamptz(time.Now()),
						EndDate:   database.Timestamptz(time.Now()),
					},
					{
						ID:        database.Text("study-plan-item-id"),
						StartDate: database.Timestamptz(time.Now()),
						EndDate:   database.Timestamptz(time.Now()),
					},
					{
						ID:        database.Text("study-plan-item-id"),
						StartDate: database.Timestamptz(time.Now()),
						EndDate:   database.Timestamptz(time.Now()),
					},
				},
			},
			expectedErr: fmt.Errorf("BulkEditAssignmentTime.Exec: %w", pgx.ErrTxClosed),
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
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.(*Req)
		err := assignmentStudyPlanItemRepo.BulkEditAssignmentTime(ctx, db, req.studentID, req.ens)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func Test_AssignmentStudyPlanItem_BulkUpsertByStudyPlanItem(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	assignmentStudyPlanItemRepo := &AssignmentStudyPlanItemRepo{}
	validAssignmentStudyPlanItemReq := []*entities.AssignmentStudyPlanItem{
		{
			AssignmentID:    database.Text("1"),
			StudyPlanItemID: database.Text("study-plan-item-1"),
		},
		{
			AssignmentID:    database.Text("1"),
			StudyPlanItemID: database.Text("study-plan-item-1"),
		},
	}
	testCases := []TestCase{
		{
			name:        "happy case",
			req:         validAssignmentStudyPlanItemReq,
			expectedErr: nil,
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validAssignmentStudyPlanItemReq); i++ {
					_, field := validAssignmentStudyPlanItemReq[i].FieldMap()
					fields = append(fields, field...)
				}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("Exec", args...).Once().Return(cmdTag, nil)
			},
		},
		{
			name:        "exec error",
			req:         validAssignmentStudyPlanItemReq,
			expectedErr: fmt.Errorf("eureka_db.BulkUpsertByStudyPlanItem error: exec error"),
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validAssignmentStudyPlanItemReq); i++ {
					_, field := validAssignmentStudyPlanItemReq[i].FieldMap()
					fields = append(fields, field...)
				}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("Exec", args...).Once().Return(cmdTag, fmt.Errorf("exec error"))
			},
		},
		{
			name:        "no row affected",
			req:         validAssignmentStudyPlanItemReq,
			expectedErr: fmt.Errorf("eureka_db.BulkUpsertByStudyPlanItem error: no row affected"),
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validAssignmentStudyPlanItemReq); i++ {
					_, field := validAssignmentStudyPlanItemReq[i].FieldMap()
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
		err := assignmentStudyPlanItemRepo.BulkUpsertByStudyPlanItem(ctx, db, testCase.req.([]*entities.AssignmentStudyPlanItem))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
