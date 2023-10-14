package repositories

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

func TestExamLOSubmissionScore_List(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ExamLOSubmissionScoreRepo{}

	now := time.Now()
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				e := &entities.ExamLOSubmissionScore{
					BaseEntity: entities.BaseEntity{
						CreatedAt: database.Timestamptz(now),
						UpdatedAt: database.Timestamptz(now),
						DeletedAt: database.Timestamptz(now),
					},
					SubmissionID:      database.Text("submission-id"),
					QuizID:            database.Text("quiz-id"),
					TeacherID:         database.Text("teacher-id"),
					ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
					TeacherComment:    database.Text("teacher-comment"),
					IsCorrect: pgtype.BoolArray{
						Elements: []pgtype.Bool{
							database.Bool(true),
							database.Bool(false),
						},
					},
					IsAccepted: database.Bool(true),
					Point:      database.Int4(1),
				}
				fields, values := e.FieldMap()
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			req: &ExamLOSubmissionScoreFilter{
				SubmissionID:      database.Text("submission-id"),
				ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
			},
			expectedResp: []*entities.ExamLOSubmissionScore{
				{
					BaseEntity: entities.BaseEntity{
						CreatedAt: database.Timestamptz(now),
						UpdatedAt: database.Timestamptz(now),
						DeletedAt: database.Timestamptz(now),
					},
					SubmissionID:      database.Text("submission-id"),
					QuizID:            database.Text("quiz-id"),
					TeacherID:         database.Text("teacher-id"),
					ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
					TeacherComment:    database.Text("teacher-comment"),
					IsCorrect: pgtype.BoolArray{
						Elements: []pgtype.Bool{
							database.Bool(true),
							database.Bool(false),
						},
					},
					IsAccepted: database.Bool(true),
					Point:      database.Int4(1),
				},
			},
		},
		{
			name: "error no rows",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				e := &entities.ExamLOSubmissionScore{}
				fields, values := e.FieldMap()
				mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{values})
			},
			req: &ExamLOSubmissionScoreFilter{
				SubmissionID:      database.Text("submission-id"),
				ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
			},
			expectedErr: fmt.Errorf("database.Select: %w", fmt.Errorf("rows.Scan: %w", pgx.ErrNoRows)),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.List(ctx, mockDB.DB, testCase.req.(*ExamLOSubmissionScoreFilter))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp.([]*entities.ExamLOSubmissionScore), resp)
			}
		})
	}
}

func TestExamLOSubmissionScore_Upsert(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ExamLOSubmissionScoreRepo{}

	examLOSubmissionScore := &entities.ExamLOSubmissionScore{}
	fieldNames := database.GetFieldNames(examLOSubmissionScore)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))

	upsertExamLOSubmissionScore := `
    INSERT INTO %s (%s) VALUES (%s)
    ON CONFLICT ON CONSTRAINT exam_lo_submission_score_pk DO UPDATE SET
        teacher_id = EXCLUDED.teacher_id,
        teacher_comment = EXCLUDED.teacher_comment,
        is_correct = EXCLUDED.is_correct,
        is_accepted = EXCLUDED.is_accepted,
        updated_at = EXCLUDED.updated_at,
        point = EXCLUDED.point,
        shuffled_quiz_set_id = EXCLUDED.shuffled_quiz_set_id;
	`

	query := fmt.Sprintf(upsertExamLOSubmissionScore, examLOSubmissionScore.TableName(), strings.Join(fieldNames, ","), placeHolders)
	args := []interface{}{mock.Anything, query}

	scanFields := database.GetScanFields(examLOSubmissionScore, fieldNames)
	args = append(args, scanFields...)

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
			req:          examLOSubmissionScore,
			expectedResp: 1,
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("0"), pgx.ErrNoRows, args...)
			},
			req:          examLOSubmissionScore,
			expectedResp: 0,
			expectedErr:  fmt.Errorf("db.Exec: %w", pgx.ErrNoRows),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.Upsert(ctx, mockDB.DB, testCase.req.(*entities.ExamLOSubmissionScore))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestExamLOSubmissionScore_Delete(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &ExamLOSubmissionScoreRepo{}

	examLOSubmissionScore := entities.ExamLOSubmissionScore{
		SubmissionID:      database.Text("examLOSubmissionID-1"),
		ShuffledQuizSetID: database.Text("shuffled_quiz_set_1"),
	}
	query := fmt.Sprintf(deleteSubmissionQuery, examLOSubmissionScore.TableName())
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", mock.Anything, query, examLOSubmissionScore.SubmissionID).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
			req:          examLOSubmissionScore.SubmissionID,
			expectedErr:  nil,
			expectedResp: int64(1),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.Delete(ctx, mockDB.DB, examLOSubmissionScore.SubmissionID)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}
