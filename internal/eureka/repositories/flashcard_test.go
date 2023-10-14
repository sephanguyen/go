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

func TestFlashcardRepo_Insert(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &FlashcardRepo{}
	validFlashcard := &entities.Flashcard{
		LearningMaterial: entities.LearningMaterial{
			ID:           database.Text("lm-id-1"),
			Name:         database.Text("flashcard-1"),
			TopicID:      database.Text("topic-id-1"),
			Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String()),
			DisplayOrder: database.Int2(1),
		},
	}
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				_, fields := validFlashcard.FieldMap()
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
			req: validFlashcard,
		},
		{
			name: "error execute query",
			setup: func(ctx context.Context) {
				_, fields := validFlashcard.FieldMap()
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), fmt.Errorf("error execute query"), args...)
			},
			req:         validFlashcard,
			expectedErr: fmt.Errorf("database.Insert: %w", fmt.Errorf("error execute query")),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.Insert(ctx, mockDB.DB, testCase.req.(*entities.Flashcard))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestFlashcardRepo_Update(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &FlashcardRepo{}
	validFlashcard := &entities.Flashcard{
		LearningMaterial: entities.LearningMaterial{
			ID:   database.Text("lm-id-1"),
			Name: database.Text("flashcard-1"),
		},
	}
	query := "UPDATE flash_card SET name = $1, updated_at = $2 WHERE learning_material_id = $3;"
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, mock.Anything, query, &validFlashcard.Name, mock.Anything, &validFlashcard.ID)
			},
			req: validFlashcard,
		},
		{
			name: "error execute query",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), fmt.Errorf("error execute query"), mock.Anything, query, &validFlashcard.Name, mock.Anything, &validFlashcard.ID)
			},
			req:         validFlashcard,
			expectedErr: fmt.Errorf("database.UpdateFields: %w", fmt.Errorf("error execute query")),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.Update(ctx, mockDB.DB, testCase.req.(*entities.Flashcard))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestFlashcardRepo_ListFlashcard(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &FlashcardRepo{}
	listFlashcardArgs := ListFlashcardArgs{}
	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				e := &entities.Flashcard{}
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				fields, values := e.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
		{
			name:        "error no rows",
			expectedErr: fmt.Errorf("database.Select: err db.Query: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				e := &entities.Flashcard{}
				mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				fields, values := e.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			_, err := repo.ListFlashcard(ctx, mockDB.DB, &listFlashcardArgs)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestExamLORepo_ListFlashcardBase(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &FlashcardRepo{}
	listFlashcardArgs := ListFlashcardArgs{}

	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				e := &entities.FlashcardBase{}
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				fields, values := e.Flashcard.FieldMap()
				fields = append(fields, "total_question")
				values = append(values, &e.TotalQuestion)

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
		{
			name:        "error no rows",
			expectedErr: fmt.Errorf("db.Query: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				e := &entities.FlashcardBase{}
				mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				fields, values := e.Flashcard.FieldMap()
				fields = append(fields, "total_question")
				values = append(values, &e.TotalQuestion)

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			_, err := repo.ListFlashcardBase(ctx, mockDB.DB, &listFlashcardArgs)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestFlashcardRepo_ListByTopicIDs(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &FlashcardRepo{}
	var flashcard entities.Flashcard
	query := fmt.Sprintf(queryListByTopicIDs, strings.Join(database.GetFieldNames(&flashcard), ","), flashcard.TableName())
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, query, mock.Anything)
				e := &entities.Flashcard{
					LearningMaterial: entities.LearningMaterial{
						ID:      database.Text("flashcard-id"),
						TopicID: database.Text("topic-id-1"),
					},
				}
				fields, values := e.FieldMap()
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			req: database.TextArray([]string{"topic-id-1"}),
			expectedResp: []*entities.Flashcard{
				{
					LearningMaterial: entities.LearningMaterial{
						ID:      database.Text("flashcard-id"),
						TopicID: database.Text("topic-id-1"),
					}},
			},
		},
		{
			name: "error no rows",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, query, mock.Anything)
				e := &entities.Flashcard{}
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
				assert.Equal(t, testCase.expectedResp.([]*entities.Flashcard), resp)
			}
		})
	}
}

func TestFlashcardRepo_BulkInsert(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &FlashcardRepo{}
	validFlashcard1 := &entities.Flashcard{
		LearningMaterial: entities.LearningMaterial{
			ID:           database.Text("lm-id-1"),
			Name:         database.Text("flashcard-1"),
			TopicID:      database.Text("topic-id-1"),
			Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String()),
			DisplayOrder: database.Int2(1),
		},
	}
	validFlashcard2 := &entities.Flashcard{
		LearningMaterial: entities.LearningMaterial{
			ID:           database.Text("lm-id-2"),
			Name:         database.Text("flashcard-2"),
			TopicID:      database.Text("topic-id-2"),
			Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String()),
			DisplayOrder: database.Int2(2),
		},
	}
	validFlashcards := []*entities.Flashcard{
		validFlashcard1, validFlashcard2,
	}
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				_, fields1 := validFlashcard1.FieldMap()
				_, fields2 := validFlashcard2.FieldMap()
				var fields = make([]interface{}, 0)
				fields = append(fields, fields1...)
				fields = append(fields, fields2...)
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
			req: validFlashcards,
		},
		{
			name: "error execute query",
			setup: func(ctx context.Context) {
				_, fields1 := validFlashcard1.FieldMap()
				_, fields2 := validFlashcard2.FieldMap()
				var fields = make([]interface{}, 0)
				fields = append(fields, fields1...)
				fields = append(fields, fields2...)
				args := append([]interface{}{mock.Anything, mock.Anything}, fields...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), fmt.Errorf("error execute query"), args...)
			},
			req:         validFlashcards,
			expectedErr: fmt.Errorf("FlashcardRepo database.BulkInsert error: error execute query"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.BulkInsert(ctx, mockDB.DB, testCase.req.([]*entities.Flashcard))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}
