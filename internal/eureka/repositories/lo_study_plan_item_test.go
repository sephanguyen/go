package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBulkUpsertLoStudyPlanItem(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	loStudyPlanItemRepo := &LoStudyPlanItemRepo{}
	validLoStudyPlanItemReq := []*entities.LoStudyPlanItem{
		{
			LoID:            database.Text("1"),
			StudyPlanItemID: database.Text("study-plan-item-1"),
		},
		{
			LoID:            database.Text("2"),
			StudyPlanItemID: database.Text("study-plan-item-2"),
		},
	}
	testCases := []TestCase{
		{
			name:        "happy case",
			req:         validLoStudyPlanItemReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				var fields []interface{}
				for i := 0; i < len(validLoStudyPlanItemReq); i++ {
					_, field := validLoStudyPlanItemReq[i].FieldMap()
					fields = append(fields, field...)
				}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("Exec", args...).Once().Return(cmdTag, nil)
			},
		},
		{
			name:        "exec error",
			req:         validLoStudyPlanItemReq,
			expectedErr: fmt.Errorf("eureka_db.BulkUpsertLoStudyPlanItem error: exec error"),
			setup: func(ctx context.Context) {
				var fields []interface{}
				for i := 0; i < len(validLoStudyPlanItemReq); i++ {
					_, field := validLoStudyPlanItemReq[i].FieldMap()
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
		err := loStudyPlanItemRepo.BulkInsert(ctx, db, testCase.req.([]*entities.LoStudyPlanItem))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestLoStudyPlanRepo_CopyFromStudyPlan(t *testing.T) {
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

func TestLoStudyPlanRepo_UpdateCompleted(t *testing.T) {
	t.Parallel()
	r := &LoStudyPlanItemRepo{}
	mockDB := testutil.NewMockDB()
	loID := database.Text("lo_id")
	studyPlanItemID := database.Text("study_plan_item_id")
	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag([]byte(`1`)), nil, mock.Anything, mock.Anything, mock.Anything, studyPlanItemID, loID)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := r.UpdateCompleted(ctx, mockDB.DB, studyPlanItemID, loID)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestLoStudyPlanRepo_FindByStudyPlanItemIDs(t *testing.T) {
	t.Parallel()

	r := &LoStudyPlanItemRepo{}
	mockDB := testutil.NewMockDB()

	validReq := database.TextArray([]string{"valid-req-1", "valid-req-2"})
	testCases := []TestCase{
		{
			name:        "happy case",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				e := &entities.LoStudyPlanItem{}
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
				e := &entities.LoStudyPlanItem{}
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
		e := &entities.LoStudyPlanItem{}
		_, err := r.FindByStudyPlanItemIDs(ctx, mockDB.DB, validReq)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func Test_LoStudyPlanItem_BulkUpsertByStudyPlanItem(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	loStudyPlanItemRepo := &LoStudyPlanItemRepo{}
	validLoStudyPlanItemReq := []*entities.LoStudyPlanItem{
		{
			LoID:            database.Text("1"),
			StudyPlanItemID: database.Text("study-plan-item-1"),
		},
		{
			LoID:            database.Text("2"),
			StudyPlanItemID: database.Text("study-plan-item-2"),
		},
		{
			LoID:            database.Text("3"),
			StudyPlanItemID: database.Text("study-plan-item-3"),
		},
	}
	testCases := []TestCase{
		{
			name:        "happy case",
			req:         validLoStudyPlanItemReq,
			expectedErr: nil,
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validLoStudyPlanItemReq); i++ {
					_, field := validLoStudyPlanItemReq[i].FieldMap()
					fields = append(fields, field...)
				}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("Exec", args...).Once().Return(cmdTag, nil)
			},
		},
		{
			name:        "exec error",
			req:         validLoStudyPlanItemReq,
			expectedErr: fmt.Errorf("eureka_db.BulkUpsertLoByStudyPlanItem error: exec error"),
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validLoStudyPlanItemReq); i++ {
					_, field := validLoStudyPlanItemReq[i].FieldMap()
					fields = append(fields, field...)
				}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("Exec", args...).Once().Return(cmdTag, fmt.Errorf("exec error"))
			},
		},
		{
			name:        "no row affected",
			req:         validLoStudyPlanItemReq,
			expectedErr: fmt.Errorf("eureka_db.BulkUpsertLoByStudyPlanItem error: no row affected"),
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validLoStudyPlanItemReq); i++ {
					_, field := validLoStudyPlanItemReq[i].FieldMap()
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
		err := loStudyPlanItemRepo.BulkUpsertByStudyPlanItem(ctx, db, testCase.req.([]*entities.LoStudyPlanItem))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
