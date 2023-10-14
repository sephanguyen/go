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

func TestLearningObjectiveRepoV2_Insert(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &LearningObjectiveRepoV2{}
	validLearningObjective := &entities.LearningObjectiveV2{
		LearningMaterial: entities.LearningMaterial{
			ID:           database.Text("lm-id-1"),
			Name:         database.Text("learning-objective-1"),
			TopicID:      database.Text("topic-id-1"),
			Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE.String()),
			DisplayOrder: database.Int2(1),
		},
	}
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				_, fields := validLearningObjective.FieldMap()
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
			req: validLearningObjective,
		},
		{
			name: "error execute query",
			setup: func(ctx context.Context) {
				_, fields := validLearningObjective.FieldMap()
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), fmt.Errorf("error execute query"), args...)
			},
			req:         validLearningObjective,
			expectedErr: fmt.Errorf("database.Insert: %w", fmt.Errorf("error execute query")),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.Insert(ctx, mockDB.DB, testCase.req.(*entities.LearningObjectiveV2))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestLearningObjectiveRepoV2_Update(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &LearningObjectiveRepoV2{}
	validLearningObjective := &entities.LearningObjectiveV2{
		LearningMaterial: entities.LearningMaterial{
			ID:   database.Text("lm-id-1"),
			Name: database.Text("learning_objective-1"),
		},
		Video:       database.Text("video-1"),
		VideoScript: database.Text("video_script-1"),
		StudyGuide:  database.Text("study-guide-1"),
	}
	query := "UPDATE learning_objective SET name = $1, updated_at = $2, video = $3, study_guide = $4, video_script = $5 WHERE learning_material_id = $6;"
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, mock.Anything, query, &validLearningObjective.Name, mock.Anything, mock.Anything, mock.Anything, mock.Anything, &validLearningObjective.ID)
			},
			req: validLearningObjective,
		},
		{
			name: "error execute query",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), fmt.Errorf("error execute query"), mock.Anything, query, &validLearningObjective.Name, mock.Anything, mock.Anything, mock.Anything, mock.Anything, &validLearningObjective.ID)
			},
			req:         validLearningObjective,
			expectedErr: fmt.Errorf("database.UpdateFields: %w", fmt.Errorf("error execute query")),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.Update(ctx, mockDB.DB, testCase.req.(*entities.LearningObjectiveV2))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestLearningObjectiveRepoV2_ListByIDs(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &LearningObjectiveRepoV2{}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
				e := &entities.LearningObjectiveV2{
					LearningMaterial: entities.LearningMaterial{
						ID:           database.Text("lm-id-1"),
						TopicID:      database.Text("topic-id-1"),
						Name:         database.Text("learning-objective-1"),
						Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE.String()),
						DisplayOrder: database.Int2(1),
					},
					Video:       database.Text("video-1"),
					StudyGuide:  database.Text("study-guide-1"),
					VideoScript: database.Text("video-script-1"),
				}
				fields, values := e.FieldMap()
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			req: database.TextArray([]string{"lm-id-1"}),
			expectedResp: []*entities.LearningObjectiveV2{
				{
					LearningMaterial: entities.LearningMaterial{
						ID:           database.Text("lm-id-1"),
						TopicID:      database.Text("topic-id-1"),
						Name:         database.Text("learning-objective-1"),
						Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE.String()),
						DisplayOrder: database.Int2(1),
					},
					Video:       database.Text("video-1"),
					StudyGuide:  database.Text("study-guide-1"),
					VideoScript: database.Text("video-script-1"),
				},
			},
		},
		{
			name: "missing learning objective",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
				e := &entities.LearningObjectiveV2{}
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
				assert.Equal(t, testCase.expectedResp.([]*entities.LearningObjectiveV2), resp)
			}
		})
	}
}

func TestLearningObjectiveRepoV2_ListLOBaseByIDs(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &LearningObjectiveRepoV2{}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
				e := &entities.LearningObjectiveBaseV2{
					LearningObjectiveV2: entities.LearningObjectiveV2{
						LearningMaterial: entities.LearningMaterial{
							ID:           database.Text("lm-id-1"),
							TopicID:      database.Text("topic-id-1"),
							Name:         database.Text("learning-objective-1"),
							Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE.String()),
							DisplayOrder: database.Int2(1),
						},
						Video:       database.Text("video-1"),
						StudyGuide:  database.Text("study-guide-1"),
						VideoScript: database.Text("video-script-1"),
					},
				}
				fields, values := e.FieldMap()
				fields = append(fields, "total_question")
				values = append(values, &e.TotalQuestion)
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			req: database.TextArray([]string{"lm-id-1"}),
			expectedResp: []*entities.LearningObjectiveBaseV2{
				{
					LearningObjectiveV2: entities.LearningObjectiveV2{
						LearningMaterial: entities.LearningMaterial{
							ID:           database.Text("lm-id-1"),
							TopicID:      database.Text("topic-id-1"),
							Name:         database.Text("learning-objective-1"),
							Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE.String()),
							DisplayOrder: database.Int2(1),
						},
						Video:       database.Text("video-1"),
						StudyGuide:  database.Text("study-guide-1"),
						VideoScript: database.Text("video-script-1"),
					},
				},
			},
		},
		{
			name: "missing learning objective",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
				e := &entities.LearningObjectiveBaseV2{}
				fields, values := e.FieldMap()
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
			resp, err := repo.ListLOBaseByIDs(ctx, mockDB.DB, testCase.req.(pgtype.TextArray))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp.([]*entities.LearningObjectiveBaseV2), resp)
			}
		})
	}
}

func TestLearningObjectiveRepoV2_ListByTopicIDs(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &LearningObjectiveRepoV2{}
	var learningObjectiveV2 entities.LearningObjectiveV2
	query := fmt.Sprintf(queryListByTopicIDs, strings.Join(database.GetFieldNames(&learningObjectiveV2), ","), learningObjectiveV2.TableName())
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, query, mock.Anything)
				e := &entities.LearningObjectiveV2{
					LearningMaterial: entities.LearningMaterial{
						ID:      database.Text("lo-assignment-id"),
						TopicID: database.Text("topic-id-1"),
					},
				}
				fields, values := e.FieldMap()
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			req: database.TextArray([]string{"topic-id-1"}),
			expectedResp: []*entities.LearningObjectiveV2{
				{
					LearningMaterial: entities.LearningMaterial{
						ID:      database.Text("lo-assignment-id"),
						TopicID: database.Text("topic-id-1"),
					}},
			},
		},
		{
			name: "error no rows",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, query, mock.Anything)
				e := &entities.LearningObjectiveV2{}
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
				assert.Equal(t, testCase.expectedResp.([]*entities.LearningObjectiveV2), resp)
			}
		})
	}
}

func TestLOV2Repo_BulkInsert(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &LearningObjectiveRepoV2{}
	validLo1 := &entities.LearningObjectiveV2{
		LearningMaterial: entities.LearningMaterial{
			ID:           database.Text("lm-id-1"),
			Name:         database.Text("flashcard-1"),
			TopicID:      database.Text("topic-id-1"),
			Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE.String()),
			DisplayOrder: database.Int2(1),
		},
	}
	validLo2 := &entities.LearningObjectiveV2{
		LearningMaterial: entities.LearningMaterial{
			ID:           database.Text("lm-id-2"),
			Name:         database.Text("flashcard-2"),
			TopicID:      database.Text("topic-id-2"),
			Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE.String()),
			DisplayOrder: database.Int2(2),
		},
	}
	validLOs := []*entities.LearningObjectiveV2{
		validLo1, validLo2,
	}
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				_, fields1 := validLo1.FieldMap()
				_, fields2 := validLo2.FieldMap()
				var fields = make([]interface{}, 0)
				fields = append(fields, fields1...)
				fields = append(fields, fields2...)
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
			req: validLOs,
		},
		{
			name: "error execute query",
			setup: func(ctx context.Context) {
				_, fields1 := validLo1.FieldMap()
				_, fields2 := validLo2.FieldMap()
				var fields = make([]interface{}, 0)
				fields = append(fields, fields1...)
				fields = append(fields, fields2...)
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), fmt.Errorf("error execute query"), args...)
			},
			req:         validLOs,
			expectedErr: fmt.Errorf("LearningObjectiveRepoV2 database.BulkInsert error: error execute query"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.BulkInsert(ctx, mockDB.DB, testCase.req.([]*entities.LearningObjectiveV2))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}
