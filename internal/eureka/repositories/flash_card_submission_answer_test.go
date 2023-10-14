package repositories

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFlashCardSubmissionAnswer_Upsert(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &FlashCardSubmissionAnswerRepo{}

	flashCardSubmissionAnswer := &entities.FlashCardSubmissionAnswer{}
	fieldNames := database.GetFieldNames(flashCardSubmissionAnswer)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))

	upsertFlashCardSubmissionAnswer := `
    INSERT INTO %s (%s) VALUES (%s)
    ON CONFLICT ON CONSTRAINT flash_card_submission_answer_pk DO UPDATE SET
        student_text_answer = EXCLUDED.student_text_answer,
        correct_text_answer = EXCLUDED.correct_text_answer,
        student_index_answer = EXCLUDED.student_index_answer,
        correct_index_answer = EXCLUDED.correct_index_answer,
        is_correct = EXCLUDED.is_correct,
        is_accepted = EXCLUDED.is_accepted,
        point = EXCLUDED.point,
        updated_at = EXCLUDED.updated_at;
	`

	query := fmt.Sprintf(upsertFlashCardSubmissionAnswer, flashCardSubmissionAnswer.TableName(), strings.Join(fieldNames, ","), placeHolders)
	args := []interface{}{mock.Anything, query}

	scanFields := database.GetScanFields(flashCardSubmissionAnswer, fieldNames)
	args = append(args, scanFields...)

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
			req:         flashCardSubmissionAnswer,
			expectedErr: nil,
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("0"), pgx.ErrNoRows, args...)
			},
			req:         flashCardSubmissionAnswer,
			expectedErr: fmt.Errorf("db.Exec: %w", pgx.ErrNoRows),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.Upsert(ctx, mockDB.DB, testCase.req.(*entities.FlashCardSubmissionAnswer))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestFlashCardSubmissionAnswer_ListSubmissionAnswers(t *testing.T) {
	mockDB := testutil.NewMockDB()
	loSubmissionAnswerRepo := &LOSubmissionAnswerRepo{}
	type Req struct {
		SetID  pgtype.Text
		Limit  pgtype.Int8
		Offset pgtype.Int8
	}
	type Resp struct {
		Submissions []*entities.LOSubmissionAnswer
		QuizIDs     []pgtype.Text
	}
	testCases := []TestCase{{
		name: "happy case",
		req: Req{
			SetID:  database.Text("set-1"),
			Limit:  database.Int8(10),
			Offset: database.Int8(0),
		},
		setup: func(ctx context.Context) {
			setID := database.Text("set-1")
			mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, setID, database.Int8(10), database.Int8(0))

			quizIDs := []pgtype.Text{database.Text("quiz-1"), database.Text("quiz-2")}
			mockDB.MockScanArray(nil, []string{"quiz_id"}, [][]interface{}{
				{&quizIDs[0]},
				{&quizIDs[1]},
			})
			mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, setID, quizIDs)

			submissions := []*entities.LOSubmissionAnswer{
				{ShuffledQuizSetID: database.Text("set-1"), QuizID: database.Text("quiz-1")},
				{ShuffledQuizSetID: database.Text("set-1"), QuizID: database.Text("quiz-2")},
			}
			fields, _ := submissions[0].FieldMap()
			dst := sliceutils.Map(submissions, func(sub *entities.LOSubmissionAnswer) []interface{} {
				_, values := sub.FieldMap()
				return values
			})
			mockDB.MockScanArray(nil, fields, dst)
		},
		expectedResp: Resp{
			Submissions: []*entities.LOSubmissionAnswer{
				{ShuffledQuizSetID: database.Text("set-1"), QuizID: database.Text("quiz-1")},
				{ShuffledQuizSetID: database.Text("set-1"), QuizID: database.Text("quiz-2")},
			},
			QuizIDs: []pgtype.Text{database.Text("quiz-1"), database.Text("quiz-2")},
		},
	}}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.(Req)
		expectedResp := testCase.expectedResp.(Resp)

		submissions, quizIDs, err := loSubmissionAnswerRepo.ListSubmissionAnswers(ctx, mockDB.DB, req.SetID, req.Limit, req.Offset)

		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, len(expectedResp.Submissions), len(submissions))
			assert.Equal(t, expectedResp.QuizIDs, quizIDs)
			assert.Equal(t, nil, err)
		}
	}
}
