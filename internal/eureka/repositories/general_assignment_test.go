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

func TestAssignmentRepo_List(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &GeneralAssignmentRepo{}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
				e := &entities.GeneralAssignment{
					LearningMaterial: entities.LearningMaterial{
						ID:           database.Text("lm-id-1"),
						TopicID:      database.Text("topic-id-1"),
						Name:         database.Text("lm-1"),
						Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_GENERAL_ASSIGNMENT.String()),
						DisplayOrder: database.Int2(1),
					},
					Attachments:            database.TextArray([]string{"attachment-1", "attachment-2"}),
					Instruction:            database.Text("instruction"),
					MaxGrade:               database.Int4(10),
					IsRequiredGrade:        database.Bool(true),
					AllowResubmission:      database.Bool(true),
					RequireAttachment:      database.Bool(false),
					AllowLateSubmission:    database.Bool(false),
					RequireAssignmentNote:  database.Bool(false),
					RequireVideoSubmission: database.Bool(true),
				}
				fields, values := e.FieldMap()
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			req: database.TextArray([]string{"lm-id-1"}),
			expectedResp: []*entities.GeneralAssignment{
				{
					LearningMaterial: entities.LearningMaterial{
						ID:           database.Text("lm-id-1"),
						TopicID:      database.Text("topic-id-1"),
						Name:         database.Text("lm-1"),
						Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_GENERAL_ASSIGNMENT.String()),
						DisplayOrder: database.Int2(1),
					},
					Attachments:            database.TextArray([]string{"attachment-1", "attachment-2"}),
					Instruction:            database.Text("instruction"),
					MaxGrade:               database.Int4(10),
					IsRequiredGrade:        database.Bool(true),
					AllowResubmission:      database.Bool(true),
					RequireAttachment:      database.Bool(false),
					AllowLateSubmission:    database.Bool(false),
					RequireAssignmentNote:  database.Bool(false),
					RequireVideoSubmission: database.Bool(true),
				},
			},
		},
		{
			name: "missing assignment",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
				e := &entities.GeneralAssignment{}
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
				assert.Equal(t, testCase.expectedResp.([]*entities.GeneralAssignment), resp)
			}
		})
	}
}

func TestGeneralAssignmentRepo_BulkInsert(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &GeneralAssignmentRepo{}
	validAssignment1 := &entities.GeneralAssignment{
		LearningMaterial: entities.LearningMaterial{
			ID:           database.Text("lm-id-1"),
			Name:         database.Text("flashcard-1"),
			TopicID:      database.Text("topic-id-1"),
			Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_GENERAL_ASSIGNMENT.String()),
			DisplayOrder: database.Int2(1),
		},
	}
	validAssignment2 := &entities.GeneralAssignment{
		LearningMaterial: entities.LearningMaterial{
			ID:           database.Text("lm-id-2"),
			Name:         database.Text("flashcard-2"),
			TopicID:      database.Text("topic-id-2"),
			Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_GENERAL_ASSIGNMENT.String()),
			DisplayOrder: database.Int2(2),
		},
	}
	validAssignments := []*entities.GeneralAssignment{
		validAssignment1, validAssignment2,
	}
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				_, fields1 := validAssignment1.FieldMap()
				_, fields2 := validAssignment2.FieldMap()
				var fields = make([]interface{}, 0)
				fields = append(fields, fields1...)
				fields = append(fields, fields2...)
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
			req: validAssignments,
		},
		{
			name: "error execute query",
			setup: func(ctx context.Context) {
				_, fields1 := validAssignment1.FieldMap()
				_, fields2 := validAssignment2.FieldMap()
				var fields = make([]interface{}, 0)
				fields = append(fields, fields1...)
				fields = append(fields, fields2...)
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), fmt.Errorf("error execute query"), args...)
			},
			req:         validAssignments,
			expectedErr: fmt.Errorf("GeneralAssignmentRepo database.BulkInsert error: error execute query"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.BulkInsert(ctx, mockDB.DB, testCase.req.([]*entities.GeneralAssignment))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestGeneralAssignmentRepo_ListByTopicIDs(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &GeneralAssignmentRepo{}
	var assignment entities.GeneralAssignment
	query := fmt.Sprintf(queryListByTopicIDs, strings.Join(database.GetFieldNames(&assignment), ","), assignment.TableName())
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, query, mock.Anything)
				e := &entities.GeneralAssignment{
					LearningMaterial: entities.LearningMaterial{
						ID:      database.Text("assignment-id"),
						TopicID: database.Text("topic-id-1"),
					},
				}
				fields, values := e.FieldMap()
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			req: database.TextArray([]string{"topic-id-1"}),
			expectedResp: []*entities.GeneralAssignment{
				{
					LearningMaterial: entities.LearningMaterial{
						ID:      database.Text("assignment-id"),
						TopicID: database.Text("topic-id-1"),
					}},
			},
		},
		{
			name: "error no rows",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, query, mock.Anything)
				e := &entities.GeneralAssignment{}
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
				assert.Equal(t, testCase.expectedResp.([]*entities.GeneralAssignment), resp)
			}
		})
	}
}

func TestGeneralAssignmentRepo_Insert(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &GeneralAssignmentRepo{}
	validAssignment := &entities.GeneralAssignment{
		LearningMaterial: entities.LearningMaterial{
			ID:           database.Text("lm-id-1"),
			Name:         database.Text("assignment-1"),
			TopicID:      database.Text("topic-id-1"),
			Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_GENERAL_ASSIGNMENT.String()),
			DisplayOrder: database.Int2(1),
		},
	}
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				_, fields := validAssignment.FieldMap()
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
			req: validAssignment,
		},
		{
			name: "error execute query",
			setup: func(ctx context.Context) {
				_, fields := validAssignment.FieldMap()
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), fmt.Errorf("error execute query"), args...)
			},
			req:         validAssignment,
			expectedErr: fmt.Errorf("db.Exec: %w", fmt.Errorf("error execute query")),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.Insert(ctx, mockDB.DB, testCase.req.(*entities.GeneralAssignment))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestGeneralAssignmentRepo_Update(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &GeneralAssignmentRepo{}
	validAssignment := &entities.GeneralAssignment{
		LearningMaterial: entities.LearningMaterial{
			ID:   database.Text("lm-id-1"),
			Name: database.Text("assignment-1"),
		},
	}
	query := "UPDATE assignment SET name = $1, attachments = $2, max_grade = $3, instruction = $4, is_required_grade = $5, allow_resubmission = $6, require_attachment = $7, allow_late_submission = $8, require_assignment_note = $9, require_video_submission = $10, updated_at = $11 WHERE learning_material_id = $12;"
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, mock.Anything, query, &validAssignment.Name, mock.Anything, &validAssignment.MaxGrade, &validAssignment.Instruction, &validAssignment.IsRequiredGrade, &validAssignment.AllowResubmission, &validAssignment.RequireAttachment, &validAssignment.AllowLateSubmission, &validAssignment.RequireAssignmentNote, &validAssignment.RequireVideoSubmission, &validAssignment.UpdatedAt, &validAssignment.LearningMaterial.ID)
			},
			req: validAssignment,
		},
		{
			name: "error execute query",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), fmt.Errorf("error execute query"), mock.Anything, query, &validAssignment.Name, mock.Anything, &validAssignment.MaxGrade, &validAssignment.Instruction, &validAssignment.IsRequiredGrade, &validAssignment.AllowResubmission, &validAssignment.RequireAttachment, &validAssignment.AllowLateSubmission, &validAssignment.RequireAssignmentNote, &validAssignment.RequireVideoSubmission, &validAssignment.UpdatedAt, &validAssignment.LearningMaterial.ID)
			},
			req:         validAssignment,
			expectedErr: fmt.Errorf("error execute query"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.Update(ctx, mockDB.DB, testCase.req.(*entities.GeneralAssignment))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}
