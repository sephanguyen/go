package repositories

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLOSubmissionAnswer_Upsert(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &LOSubmissionAnswerRepo{}

	loSubmissionAnswer := &entities.LOSubmissionAnswer{}
	fieldNames := database.GetFieldNames(loSubmissionAnswer)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))

	upsertLOSubmissionAnswer := `
    INSERT INTO %s (%s) VALUES (%s)
    ON CONFLICT ON CONSTRAINT lo_submission_answer_pk DO UPDATE SET
        student_text_answer = EXCLUDED.student_text_answer,
        correct_text_answer = EXCLUDED.correct_text_answer,
        student_index_answer = EXCLUDED.student_index_answer,
        correct_index_answer = EXCLUDED.correct_index_answer,
        is_correct = EXCLUDED.is_correct,
        is_accepted = EXCLUDED.is_accepted,
        point = EXCLUDED.point,
        updated_at = EXCLUDED.updated_at,
        submitted_keys_answer = EXCLUDED.submitted_keys_answer,
        correct_keys_answer = EXCLUDED.correct_keys_answer;
	`

	query := fmt.Sprintf(upsertLOSubmissionAnswer, loSubmissionAnswer.TableName(), strings.Join(fieldNames, ","), placeHolders)
	args := []interface{}{mock.Anything, query}

	scanFields := database.GetScanFields(loSubmissionAnswer, fieldNames)
	args = append(args, scanFields...)

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
			req:         loSubmissionAnswer,
			expectedErr: nil,
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("0"), pgx.ErrNoRows, args...)
			},
			req:         loSubmissionAnswer,
			expectedErr: fmt.Errorf("db.Exec: %w", pgx.ErrNoRows),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			err := repo.Upsert(ctx, mockDB.DB, testCase.req.(*entities.LOSubmissionAnswer))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestLOSubmissionAnswer_List(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &LOSubmissionAnswerRepo{}

	e := &entities.LOSubmissionAnswer{}
	fields, values := e.FieldMap()

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			req: &LOSubmissionAnswerFilter{
				SubmissionID:      database.Text("submission-id"),
				ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
			},
		},
		{
			name: "error no rows",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{values})
			},
			req: &LOSubmissionAnswerFilter{
				SubmissionID:      database.Text("submission-id"),
				ShuffledQuizSetID: database.Text("shuffled-quiz-set-id"),
			},
			expectedErr: fmt.Errorf("database.Select: %w", fmt.Errorf("rows.Scan: %w", pgx.ErrNoRows)),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.List(ctx, mockDB.DB, testCase.req.(*LOSubmissionAnswerFilter))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp.([]*entities.LOSubmissionAnswer), resp)
			}
		})
	}
}

func TestLOSubmissionAnswer_ListSubmissionAnswers(t *testing.T) {
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

func TestLOSubmissionAnswer_BulkUpsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	loSubmissionAnswerRepo := &LOSubmissionAnswerRepo{}
	testCases := []TestCase{
		{
			name: "happy Case",
			req: []*entities.LOSubmissionAnswer{
				&entities.LOSubmissionAnswer{},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag(`1`)
				db.On("Exec", mock.Anything, mock.AnythingOfType("string"),
					mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything,
				).Once().Return(cmdTag, nil)
			},
		},
		{
			name: "error exec error",
			req: []*entities.LOSubmissionAnswer{
				&entities.LOSubmissionAnswer{},
			},
			expectedErr: fmt.Errorf("database.BulkUpsert error: error exec error"),
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag(`0`)
				db.On("Exec", mock.Anything, mock.AnythingOfType("string"),
					mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything,
				).Once().Return(cmdTag, fmt.Errorf("error exec error"))
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := loSubmissionAnswerRepo.BulkUpsert(ctx, db, testCase.req.([]*entities.LOSubmissionAnswer))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, nil, err)
		}
	}
}
