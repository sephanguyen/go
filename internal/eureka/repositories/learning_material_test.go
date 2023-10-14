package repositories

import (
	"context"
	"fmt"
	"strings"
	"testing"

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

func TestLearningMaterialRepo_Delete(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &LearningMaterialRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag(`1`), nil, mock.Anything, mock.Anything, database.Text("lm-id-1"))
			},
			req:          database.Text("lm-id-1"),
			expectedResp: nil,
		},
		{
			name: "missing learning material",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag(`0`), pgx.ErrNoRows, mock.Anything, mock.Anything, database.Text("lm-id-1"))
			},
			req:         database.Text("lm-id-1"),
			expectedErr: fmt.Errorf("db.Exec: %w", pgx.ErrNoRows),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.Delete(ctx, mockDB.DB, testCase.req.(pgtype.Text))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestLearningMaterialRepo_FindByIDs(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &LearningMaterialRepo{}

	query := fmt.Sprintf("SELECT %s FROM learning_material WHERE learning_material_id = ANY($1) AND deleted_at IS NULL", strings.Join(database.GetFieldNames(&entities.LearningMaterial{}), ","))
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, query, mock.Anything)
				e := &entities.LearningMaterial{
					ID: database.Text("lm-id-1"),
				}
				fields, values := e.FieldMap()
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			req: database.TextArray([]string{"lm-id-1"}),
			expectedResp: []*entities.LearningMaterial{
				{
					ID: database.Text("lm-id-1"),
				},
			},
		},
		{
			name: "missing book",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, query, mock.Anything)
				e := &entities.LearningMaterial{}
				fields, values := e.FieldMap()
				mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{values})
			},
			req:         database.TextArray([]string{"lm-id-1"}),
			expectedErr: fmt.Errorf("database.Select: %w", fmt.Errorf("rows.Scan: %w", pgx.ErrNoRows)),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.FindByIDs(ctx, mockDB.DB, testCase.req.(pgtype.TextArray))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp.([]*entities.LearningMaterial), resp)
			}
		})
	}
}

func TestLearningMaterialRepo_FindInfoByStudyPlanItemIdentity(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	rows := &mock_database.Rows{}
	repo := &LearningMaterialRepo{}

	type Req struct {
		studyPlanID pgtype.Text
		studentID   pgtype.Text
		lmID        pgtype.Text
	}

	testCases := []TestCase{
		{
			name: "happy case",
			req: &Req{
				studyPlanID: database.Text("study_plan_id"),
				studentID:   database.Text("student_id"),
				lmID:        database.Text(""),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name: "query error",
			req: &Req{
				studyPlanID: database.Text("study_plan_id"),
				studentID:   database.Text("student_id"),
				lmID:        database.Text(""),
			},
			expectedErr: fmt.Errorf("LearningMaterialRepo.FindByStudyPlanItemIdentity.Query: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			name: "scan error",
			req: &Req{
				studyPlanID: database.Text("study_plan_id"),
				studentID:   database.Text("student_id"),
				lmID:        database.Text(""),
			},
			expectedErr: fmt.Errorf("LearningMaterialRepo.FindByStudyPlanItemIdentity.Scan: %w", fmt.Errorf("error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)

				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
		{
			name: "rows error",
			req: &Req{
				studyPlanID: database.Text("study_plan_id"),
				studentID:   database.Text("student_id"),
				lmID:        database.Text(""),
			},
			expectedErr: fmt.Errorf("LearningMaterialRepo.FindByStudyPlanItemIdentity.Err: %w", fmt.Errorf("rows error")),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(fmt.Errorf("rows error"))
				rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.(*Req)
		_, err := repo.FindInfoByStudyPlanItemIdentity(ctx, db, req.studyPlanID, req.studentID, req.lmID)
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestLearningMaterialRepo_UpdateDisplayOrders(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &LearningMaterialRepo{}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
			req: []*entities.LearningMaterial{
				{
					ID:           database.Text("lm-id-1"),
					DisplayOrder: database.Int2(1),
				},
				{
					ID:           database.Text("lm-id-2"),
					DisplayOrder: database.Int2(1),
				},
			},
		},
		{
			name: "error send batch",
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, pgx.ErrTxClosed)
				batchResults.On("Close").Once().Return(nil)
			},
			req: []*entities.LearningMaterial{
				{
					ID:           database.Text("lm-id-1"),
					DisplayOrder: database.Int2(1),
				},
				{
					ID:           database.Text("lm-id-2"),
					DisplayOrder: database.Int2(1),
				},
			},
			expectedErr: fmt.Errorf("batchResults.Exec: %w", pgx.ErrTxClosed),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.UpdateDisplayOrders(ctx, mockDB.DB, testCase.req.([]*entities.LearningMaterial))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestLearningMaterialRepo_UpdateName(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &LearningMaterialRepo{}
	query := "UPDATE learning_material SET name = $1, updated_at = now() WHERE learning_material_id = $2::TEXT AND deleted_at IS NULL"
	lmName := database.Text("lm-name-1")
	lmID := database.Text("lm-id-1")

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, mock.Anything, query, lmName, lmID)
			},
			expectedResp: 1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			rowAff, err := repo.UpdateName(ctx, mockDB.DB, lmID, lmName)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, int(rowAff))

		})
	}
}
