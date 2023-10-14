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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLOProgressionAnswerRepo_DeleteByStudyPlanIdentity(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &LOProgressionAnswerRepo{}

	loProgressionAnswers := []entities.LOProgressionAnswer{
		{
			ProgressionAnswerID: database.Text("progression-answer-id-1"),
			ProgressionID:       database.Text("progression-id"),
			ShuffledQuizSetID:   database.Text("shuffled_quiz_set_id"),
			StudentID:           database.Text("student-id"),
			LearningMaterialID:  database.Text("learningMaterialID"),
			StudyPlanID:         database.Text("study-plan-id"),
		},
		{
			ProgressionAnswerID: database.Text("progression-answer-id-2"),
			ProgressionID:       database.Text("progression-id"),
			ShuffledQuizSetID:   database.Text("shuffled_quiz_set_id"),
			StudentID:           database.Text("student-id"),
			LearningMaterialID:  database.Text("learningMaterialID"),
			StudyPlanID:         database.Text("study-plan-id"),
		},
		{
			ProgressionAnswerID: database.Text("progression-answer-id-3"),
			ProgressionID:       database.Text("progression-id"),
			ShuffledQuizSetID:   database.Text("shuffled_quiz_set_id"),
			StudentID:           database.Text("student-id"),
			LearningMaterialID:  database.Text("learningMaterialID"),
			StudyPlanID:         database.Text("study-plan-id"),
		},
	}
	req := StudyPlanItemIdentity{
		StudentID:          loProgressionAnswers[0].StudentID,
		StudyPlanID:        loProgressionAnswers[0].StudyPlanID,
		LearningMaterialID: loProgressionAnswers[0].LearningMaterialID,
	}
	query := fmt.Sprintf(`UPDATE %s SET deleted_at = now() WHERE learning_material_id = $1::TEXT AND student_id = $2::TEXT AND study_plan_id = $3::TEXT AND deleted_at IS NULL`, loProgressionAnswers[0].TableName())

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

func TestLOProgressionAnswerRepo_ListByProgressionAndExternalIDs(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &LOProgressionAnswerRepo{}

	progressionID := database.Text("progression-id")
	externalIDs := database.TextArray([]string{"external-1"})
	answers := entities.LOProgressionAnswers{
		{
			ProgressionAnswerID: database.Text("ProgressionAnswerID-0"),
			ShuffledQuizSetID:   database.Text("ShuffledQuizSetID"),
			ProgressionID:       database.Text("progression-id"),
			QuizExternalID:      database.Text("external-1"),
		},
	}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				fields, values := answers[0].FieldMap()
				mockDB.MockScanFields(nil, fields, values)
			},
			expectedErr:  nil,
			expectedResp: answers,
		},
		{
			name: "scan error case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				e := &entities.LOProgressionAnswer{}
				fields, values := e.FieldMap()
				mockDB.MockScanFields(fmt.Errorf("error scan"), fields, values)
			},
			expectedErr:  fmt.Errorf("rows.Scan: %w", fmt.Errorf("error scan")),
			expectedResp: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.ListByProgressionAndExternalIDs(ctx, mockDB.DB, progressionID, externalIDs)
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}
