package repositories

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExamLORepo_Insert(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ExamLORepo{}
	validExamLO := &entities.ExamLO{
		LearningMaterial: entities.LearningMaterial{
			ID:           database.Text("lm-id-1"),
			Name:         database.Text("exam-lo-1"),
			TopicID:      database.Text("topic-id-1"),
			Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_EXAM_LO.String()),
			DisplayOrder: database.Int2(1),
		},
		Instruction:   database.Text("exam-lo-instruction"),
		ManualGrading: database.Bool(true),
		GradeToPass:   database.Int4(1),
		TimeLimit:     database.Int4(1),
	}
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				_, fields := validExamLO.FieldMap()
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
			req: validExamLO,
		},
		{
			name: "error execute query",
			setup: func(ctx context.Context) {
				_, fields := validExamLO.FieldMap()
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), fmt.Errorf("error execute query"), args...)
			},
			req:         validExamLO,
			expectedErr: fmt.Errorf("database.Insert: %w", fmt.Errorf("error execute query")),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.Insert(ctx, mockDB.DB, testCase.req.(*entities.ExamLO))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestExamLORepo_Update(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ExamLORepo{}
	validExamLO := &entities.ExamLO{
		LearningMaterial: entities.LearningMaterial{
			ID:   database.Text("lm-id-1"),
			Name: database.Text("exam-lo-1"),
		},
		Instruction:    database.Text("Instruction"),
		GradeToPass:    database.Int4(10),
		ManualGrading:  database.Bool(true),
		TimeLimit:      database.Int4(1),
		MaximumAttempt: database.Int4(10),
		ApproveGrading: database.Bool(true),
		GradeCapping:   database.Bool(true),
		ReviewOption:   database.Text(sspb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_AFTER_DUE_DATE.String()),
	}
	query := "UPDATE exam_lo SET name = $1, instruction = $2, grade_to_pass = $3, manual_grading = $4, time_limit = $5, updated_at = $6, maximum_attempt = $7, approve_grading = $8, grade_capping = $9, review_option = $10 WHERE learning_material_id = $11;"
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(
					t,
					pgconn.CommandTag("1"),
					nil,
					mock.Anything,
					query,
					&validExamLO.Name,
					&validExamLO.Instruction,
					&validExamLO.GradeToPass,
					&validExamLO.ManualGrading,
					&validExamLO.TimeLimit,
					mock.Anything,
					&validExamLO.MaximumAttempt,
					&validExamLO.ApproveGrading,
					&validExamLO.GradeCapping,
					&validExamLO.ReviewOption,
					&validExamLO.ID)
			},
			req: validExamLO,
		},
		{
			name: "error execute query",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(
					t,
					pgconn.CommandTag("1"),
					fmt.Errorf("error execute query"),
					mock.Anything,
					query,
					&validExamLO.Name,
					&validExamLO.Instruction,
					&validExamLO.GradeToPass,
					&validExamLO.ManualGrading,
					&validExamLO.TimeLimit,
					mock.Anything,
					&validExamLO.MaximumAttempt,
					&validExamLO.ApproveGrading,
					&validExamLO.GradeCapping,
					&validExamLO.ReviewOption,
					&validExamLO.ID,
				)
			},
			req:         validExamLO,
			expectedErr: fmt.Errorf("database.UpdateFields: %w", fmt.Errorf("error execute query")),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.Update(ctx, mockDB.DB, testCase.req.(*entities.ExamLO))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestExamLORepo_ListByIDs(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ExamLORepo{}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
				e := &entities.ExamLO{
					LearningMaterial: entities.LearningMaterial{
						ID:           database.Text("lm-id-1"),
						TopicID:      database.Text("topic-id-1"),
						Name:         database.Text("exam-lo-1"),
						Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_EXAM_LO.String()),
						DisplayOrder: database.Int2(1),
					},
					Instruction:   database.Text("instruction"),
					GradeToPass:   database.Int4(1),
					ManualGrading: database.Bool(true),
					TimeLimit:     database.Int4(1),
				}
				fields, values := e.FieldMap()
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			req: database.TextArray([]string{"lm-id-1"}),
			expectedResp: []*entities.ExamLO{
				{
					LearningMaterial: entities.LearningMaterial{
						ID:           database.Text("lm-id-1"),
						TopicID:      database.Text("topic-id-1"),
						Name:         database.Text("exam-lo-1"),
						Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_EXAM_LO.String()),
						DisplayOrder: database.Int2(1),
					},
					Instruction:   database.Text("instruction"),
					GradeToPass:   database.Int4(1),
					ManualGrading: database.Bool(true),
					TimeLimit:     database.Int4(1),
				},
			},
		},
		{
			name: "error no rows",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
				e := &entities.ExamLO{}
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
			resp, err := repo.ListByIDs(ctx, mockDB.DB, testCase.req.(pgtype.TextArray))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp.([]*entities.ExamLO), resp)
			}
		})
	}
}

func TestExamLORepo_ListExamLOBaseByIDs(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ExamLORepo{}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
				e := &entities.ExamLOBase{
					ExamLO: entities.ExamLO{
						LearningMaterial: entities.LearningMaterial{
							ID:           database.Text("lm-id-1"),
							TopicID:      database.Text("topic-id-1"),
							Name:         database.Text("exam-lo-1"),
							Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_EXAM_LO.String()),
							DisplayOrder: database.Int2(1),
						},
						Instruction:   database.Text("instruction"),
						GradeToPass:   database.Int4(1),
						ManualGrading: database.Bool(true),
						TimeLimit:     database.Int4(1),
					},
					TotalQuestion: database.Int4(1),
				}
				fields, values := e.ExamLO.FieldMap()
				fields = append(fields, "total_question")
				values = append(values, &e.TotalQuestion)
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			req: database.TextArray([]string{"lm-id-1"}),
			expectedResp: []*entities.ExamLOBase{
				{
					ExamLO: entities.ExamLO{
						LearningMaterial: entities.LearningMaterial{
							ID:           database.Text("lm-id-1"),
							TopicID:      database.Text("topic-id-1"),
							Name:         database.Text("exam-lo-1"),
							Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_EXAM_LO.String()),
							DisplayOrder: database.Int2(1),
						},
						Instruction:   database.Text("instruction"),
						GradeToPass:   database.Int4(1),
						ManualGrading: database.Bool(true),
						TimeLimit:     database.Int4(1),
					},
					TotalQuestion: database.Int4(1),
				},
			},
		},
		{
			name: "error no rows",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
				e := &entities.ExamLOBase{}
				fields, values := e.ExamLO.FieldMap()
				fields = append(fields, "total_question")
				values = append(values, &e.TotalQuestion)
				mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{values})
			},
			req:         database.TextArray([]string{"lm-id-1"}),
			expectedErr: fmt.Errorf("rows.Scan: %w", pgx.ErrNoRows),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.ListExamLOBaseByIDs(ctx, mockDB.DB, testCase.req.(pgtype.TextArray))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp.([]*entities.ExamLOBase), resp)
			}
		})
	}
}

func TestExamLORepo_ListByTopicIDs(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ExamLORepo{}
	var examLO entities.ExamLO
	query := fmt.Sprintf(queryListByTopicIDs, strings.Join(database.GetFieldNames(&examLO), ","), examLO.TableName())
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, query, mock.Anything)
				e := &entities.ExamLO{
					LearningMaterial: entities.LearningMaterial{
						ID:      database.Text("exam-lo-id"),
						TopicID: database.Text("topic-id-1"),
					},
				}
				fields, values := e.FieldMap()
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			req: database.TextArray([]string{"topic-id-1"}),
			expectedResp: []*entities.ExamLO{
				{
					LearningMaterial: entities.LearningMaterial{
						ID:      database.Text("exam-lo-id"),
						TopicID: database.Text("topic-id-1"),
					}},
			},
		},
		{
			name: "error no rows",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, query, mock.Anything)
				e := &entities.ExamLO{}
				fields, values := e.FieldMap()
				mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{values})
			},
			req:         database.TextArray([]string{"topic-id-1"}),
			expectedErr: fmt.Errorf("database.Select: %w", fmt.Errorf("rows.Scan: %w", pgx.ErrNoRows)),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.ListByTopicIDs(ctx, mockDB.DB, testCase.req.(pgtype.TextArray))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp.([]*entities.ExamLO), resp)
			}
		})
	}
}

func TestExamLORepo_BulkInsert(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ExamLORepo{}
	validExamLO1 := &entities.ExamLO{
		LearningMaterial: entities.LearningMaterial{
			ID:           database.Text("lm-id-1"),
			Name:         database.Text("flashcard-1"),
			TopicID:      database.Text("topic-id-1"),
			Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_EXAM_LO.String()),
			DisplayOrder: database.Int2(1),
		},
	}
	validExamLO2 := &entities.ExamLO{
		LearningMaterial: entities.LearningMaterial{
			ID:           database.Text("lm-id-2"),
			Name:         database.Text("flashcard-2"),
			TopicID:      database.Text("topic-id-2"),
			Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_EXAM_LO.String()),
			DisplayOrder: database.Int2(2),
		},
	}
	validExamLOs := []*entities.ExamLO{
		validExamLO1, validExamLO2,
	}
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				_, fields1 := validExamLO1.FieldMap()
				_, fields2 := validExamLO2.FieldMap()
				var fields = make([]interface{}, 0)
				fields = append(fields, fields1...)
				fields = append(fields, fields2...)
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
			req: validExamLOs,
		},
		{
			name: "error execute query",
			setup: func(ctx context.Context) {
				_, fields1 := validExamLO1.FieldMap()
				_, fields2 := validExamLO2.FieldMap()
				var fields = make([]interface{}, 0)
				fields = append(fields, fields1...)
				fields = append(fields, fields2...)
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), fmt.Errorf("error execute query"), args...)
			},
			req:         validExamLOs,
			expectedErr: fmt.Errorf("ExamLORepo database.BulkInsert error: error execute query"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.BulkInsert(ctx, mockDB.DB, testCase.req.([]*entities.ExamLO))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestExamLORepo_Get(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ExamLORepo{}
	learningMaterialID := database.Text("learningMaterialID-1")
	result := &entities.ExamLO{}
	fields, values := result.FieldMap()

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, learningMaterialID)
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			req:          learningMaterialID,
			expectedResp: result,
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, learningMaterialID)
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			req:          learningMaterialID,
			expectedResp: nil,
			expectedErr:  fmt.Errorf("database.Select: %w", pgx.ErrNoRows),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.Get(ctx, mockDB.DB, testCase.req.(pgtype.Text))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}
