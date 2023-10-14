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

func TestTaskAssignment_Insert(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &TaskAssignmentRepo{}
	validTaskAssignment := &entities.TaskAssignment{
		LearningMaterial: entities.LearningMaterial{
			ID:           database.Text("lm-id-1"),
			Name:         database.Text("task-assignment-1"),
			TopicID:      database.Text("topic-id-1"),
			Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_TASK_ASSIGNMENT.String()),
			DisplayOrder: database.Int2(1),
		},
		Attachments:               database.TextArray([]string{"attachment-1", "attachment-2"}),
		Instruction:               database.Text("instruction"),
		RequireDuration:           database.Bool(true),
		RequireCompleteDate:       database.Bool(true),
		RequireUnderstandingLevel: database.Bool(true),
		RequireCorrectness:        database.Bool(true),
		RequireAttachment:         database.Bool(false),
		RequireAssignmentNote:     database.Bool(false),
	}
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				_, fields := validTaskAssignment.FieldMap()
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
			req: validTaskAssignment,
		},
		{
			name: "error execute query",
			setup: func(ctx context.Context) {
				_, fields := validTaskAssignment.FieldMap()
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), fmt.Errorf("error execute query"), args...)
			},
			req:         validTaskAssignment,
			expectedErr: fmt.Errorf("database.Insert: %w", fmt.Errorf("error execute query")),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.Insert(ctx, mockDB.DB, testCase.req.(*entities.TaskAssignment))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestTaskAssignmentRepo_Update(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &TaskAssignmentRepo{}
	validTaskAssignment := &entities.TaskAssignment{
		LearningMaterial: entities.LearningMaterial{
			ID:   database.Text("lm-id-1"),
			Name: database.Text("task-assignment-1"),
		},
		Attachments:               database.TextArray([]string{"attachment-1", "attachment-2"}),
		Instruction:               database.Text("instruction"),
		RequireDuration:           database.Bool(true),
		RequireCompleteDate:       database.Bool(true),
		RequireUnderstandingLevel: database.Bool(true),
		RequireCorrectness:        database.Bool(true),
		RequireAttachment:         database.Bool(false),
		RequireAssignmentNote:     database.Bool(false),
	}
	fields := "name = $1, updated_at = $2, attachments = $3, instruction = $4, require_duration = $5, require_complete_date = $6, require_understanding_level = $7, require_correctness = $8, require_attachment = $9, require_assignment_note = $10"
	query := fmt.Sprintf("UPDATE task_assignment SET %s WHERE learning_material_id = $11;", fields)
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, mock.Anything, query, &validTaskAssignment.Name, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, &validTaskAssignment.ID)
			},
			req: validTaskAssignment,
		},
		{
			name: "error execute query",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), fmt.Errorf("error execute query"), mock.Anything, query, &validTaskAssignment.Name, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, &validTaskAssignment.ID)
			},
			req:         validTaskAssignment,
			expectedErr: fmt.Errorf("database.UpdateFields: %w", fmt.Errorf("error execute query")),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.Update(ctx, mockDB.DB, testCase.req.(*entities.TaskAssignment))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}
func TestTaskAssignment_ListByTopicIDs(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &TaskAssignmentRepo{}
	var taskAssignment entities.TaskAssignment
	query := fmt.Sprintf(queryListByTopicIDs, strings.Join(database.GetFieldNames(&taskAssignment), ","), taskAssignment.TableName())
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, query, mock.Anything)
				e := &entities.TaskAssignment{
					LearningMaterial: entities.LearningMaterial{
						ID:      database.Text("task-assignment-id"),
						TopicID: database.Text("topic-id-1"),
					},
				}
				fields, values := e.FieldMap()
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			req: database.TextArray([]string{"topic-id-1"}),
			expectedResp: []*entities.TaskAssignment{
				{
					LearningMaterial: entities.LearningMaterial{
						ID:      database.Text("task-assignment-id"),
						TopicID: database.Text("topic-id-1"),
					}},
			},
		},
		{
			name: "error no rows",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, query, mock.Anything)
				e := &entities.TaskAssignment{}
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
				assert.Equal(t, testCase.expectedResp.([]*entities.TaskAssignment), resp)
			}
		})
	}
}

func TestTaskAssignmentRepo_List(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &TaskAssignmentRepo{}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
				e := &entities.TaskAssignment{
					LearningMaterial: entities.LearningMaterial{
						ID:           database.Text("lm-id-1"),
						TopicID:      database.Text("topic-id-1"),
						Name:         database.Text("lm-1"),
						Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_TASK_ASSIGNMENT.String()),
						DisplayOrder: database.Int2(1),
					},
					Attachments:               database.TextArray([]string{"attachment-1", "attachment-2"}),
					Instruction:               database.Text("instruction"),
					RequireDuration:           database.Bool(true),
					RequireCompleteDate:       database.Bool(true),
					RequireUnderstandingLevel: database.Bool(false),
					RequireCorrectness:        database.Bool(false),
					RequireAttachment:         database.Bool(false),
					RequireAssignmentNote:     database.Bool(true),
				}
				fields, values := e.FieldMap()
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			req: database.TextArray([]string{"lm-id-1"}),
			expectedResp: []*entities.TaskAssignment{
				{
					LearningMaterial: entities.LearningMaterial{
						ID:           database.Text("lm-id-1"),
						TopicID:      database.Text("topic-id-1"),
						Name:         database.Text("lm-1"),
						Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_TASK_ASSIGNMENT.String()),
						DisplayOrder: database.Int2(1),
					},
					Attachments:               database.TextArray([]string{"attachment-1", "attachment-2"}),
					Instruction:               database.Text("instruction"),
					RequireDuration:           database.Bool(true),
					RequireCompleteDate:       database.Bool(true),
					RequireUnderstandingLevel: database.Bool(false),
					RequireCorrectness:        database.Bool(false),
					RequireAttachment:         database.Bool(false),
					RequireAssignmentNote:     database.Bool(true),
				},
			},
		},
		{
			name: "missing task assignment",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
				e := &entities.TaskAssignment{}
				fields, values := e.FieldMap()
				mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{values})
			},
			req:         database.TextArray([]string{"lm-id-1"}),
			expectedErr: fmt.Errorf("%w", fmt.Errorf("rows.Scan: %w", pgx.ErrNoRows)),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.List(ctx, mockDB.DB, testCase.req.(pgtype.TextArray))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp.([]*entities.TaskAssignment), resp)
			}
		})
	}
}
func TestTaskAssignmentRepo_BulkInsert(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &TaskAssignmentRepo{}
	validTaskAssignment1 := &entities.TaskAssignment{
		LearningMaterial: entities.LearningMaterial{
			ID:           database.Text("lm-id-1"),
			Name:         database.Text("flashcard-1"),
			TopicID:      database.Text("topic-id-1"),
			Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_TASK_ASSIGNMENT.String()),
			DisplayOrder: database.Int2(1),
		},
	}
	validTaskAssignment2 := &entities.TaskAssignment{
		LearningMaterial: entities.LearningMaterial{
			ID:      database.Text("lm-id-2"),
			Name:    database.Text("flashcard-2"),
			TopicID: database.Text("topic-id-2"),
			Type:    database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_TASK_ASSIGNMENT.String()), DisplayOrder: database.Int2(2),
		},
	}
	validOLs := []*entities.TaskAssignment{
		validTaskAssignment1, validTaskAssignment2,
	}
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				_, fields1 := validTaskAssignment1.FieldMap()
				_, fields2 := validTaskAssignment2.FieldMap()
				var fields = make([]interface{}, 0)
				fields = append(fields, fields1...)
				fields = append(fields, fields2...)
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
			req: validOLs,
		},
		{
			name: "error execute query",
			setup: func(ctx context.Context) {
				_, fields1 := validTaskAssignment1.FieldMap()
				_, fields2 := validTaskAssignment2.FieldMap()
				var fields = make([]interface{}, 0)
				fields = append(fields, fields1...)
				fields = append(fields, fields2...)
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), fmt.Errorf("error execute query"), args...)
			},
			req:         validOLs,
			expectedErr: fmt.Errorf("TaskAssignmentRepo database.BulkInsert error: error execute query"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.BulkInsert(ctx, mockDB.DB, testCase.req.([]*entities.TaskAssignment))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestTaskAssignmentRepo_Upsert(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &TaskAssignmentRepo{}
	validTA := &entities.TaskAssignment{}
	_, values := validTA.FieldMap()
	args := []interface{}{mock.Anything, mock.Anything}
	args = append(args, values...)
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
			req: validTA,
		},
		{
			name: "error execute query",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), pgx.ErrTxClosed, args...)
			},
			req:         validTA,
			expectedErr: fmt.Errorf("db.Exec: %w", pgx.ErrTxClosed),
		},
		{
			name: "error upsert failed",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)
			},
			req:         validTA,
			expectedErr: fmt.Errorf("upsert TaskAssignment failed"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.Upsert(ctx, mockDB.DB, testCase.req.(*entities.TaskAssignment))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}
