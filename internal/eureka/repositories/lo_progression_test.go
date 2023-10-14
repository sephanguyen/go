package repositories

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLOProgressionRepo_Upsert(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &LOProgressionRepo{}

	loProgression := &entities.LOProgression{}
	fieldNames := database.GetFieldNames(loProgression)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))

	upsertLOProgression := `
    INSERT INTO %s (%s) VALUES (%s)
    ON CONFLICT (student_id, study_plan_id, learning_material_id) WHERE (deleted_at IS NULL) DO UPDATE SET
        shuffled_quiz_set_id = EXCLUDED.shuffled_quiz_set_id,
		quiz_external_ids = EXCLUDED.quiz_external_ids,
		last_index = EXCLUDED.last_index,
        updated_at = EXCLUDED.updated_at;
	`

	query := fmt.Sprintf(upsertLOProgression, loProgression.TableName(), strings.Join(fieldNames, ","), placeHolders)
	args := []interface{}{mock.Anything, query}

	scanFields := database.GetScanFields(loProgression, fieldNames)
	args = append(args, scanFields...)

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
			req: loProgression,
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("0"), pgx.ErrNoRows, args...)
			},
			req:         loProgression,
			expectedErr: fmt.Errorf("db.Exec: %w", pgx.ErrNoRows),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.Upsert(ctx, mockDB.DB, testCase.req.(*entities.LOProgression))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestLOProgressionRepo_DeleteByStudyPlanIdentity(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &LOProgressionRepo{}

	loProgression := entities.LOProgression{
		LastIndex:          database.Int4(5),
		ProgressionID:      database.Text("progression-id"),
		ShuffledQuizSetID:  database.Text("shuffled_quiz_set_id"),
		StudentID:          database.Text("student-id"),
		LearningMaterialID: database.Text("learningMaterialID"),
		StudyPlanID:        database.Text("study-plan-id"),
	}
	req := StudyPlanItemIdentity{
		StudentID:          loProgression.StudentID,
		StudyPlanID:        loProgression.StudyPlanID,
		LearningMaterialID: loProgression.LearningMaterialID,
	}
	query := fmt.Sprintf(`UPDATE %s SET deleted_at = now() WHERE learning_material_id = $1::TEXT AND student_id = $2::TEXT AND study_plan_id = $3::TEXT AND deleted_at IS NULL`, loProgression.TableName())

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
			resp, err := repo.DeleteByStudyPlanIdentity(ctx, mockDB.DB, req)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}

func TestLOProgressionRepo_GetByStudyPlanItemIdentity(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &LOProgressionRepo{}

	loProgression := &entities.LOProgression{
		LastIndex:          database.Int4(5),
		ProgressionID:      database.Text("progression-id"),
		ShuffledQuizSetID:  database.Text("shuffled_quiz_set_id"),
		StudentID:          database.Text("student-id"),
		LearningMaterialID: database.Text("learningMaterialID"),
		StudyPlanID:        database.Text("study-plan-id"),
		QuizExternalIDs:    database.TextArray([]string{"external-01", "external-02", "external-03", "external-05"}),
	}
	req := StudyPlanItemIdentity{
		StudentID:          loProgression.StudentID,
		StudyPlanID:        loProgression.StudyPlanID,
		LearningMaterialID: loProgression.LearningMaterialID,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, req.StudyPlanID, req.LearningMaterialID, req.StudentID, int64(0), int64(3))
				fields, values := loProgression.FieldMap()
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			req:          req,
			expectedErr:  nil,
			expectedResp: loProgression,
		},
		{
			name: "err no rows",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, int64(0), int64(3))

				e := &entities.LOProgression{}
				fields, values := e.FieldMap()
				mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{values})
			},
			req:          req,
			expectedErr:  pgx.ErrNoRows,
			expectedResp: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.GetByStudyPlanItemIdentity(ctx, mockDB.DB, req, database.Int8(0), database.Int8(3))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}
