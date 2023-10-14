package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	yasuo_entities "github.com/manabie-com/backend/internal/yasuo/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpsertSpeeches_Batch(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	row := &mock_database.Row{}
	speeches := &SpeechesRepository{}
	e := &yasuo_entities.Speeches{}
	_, values := e.FieldMap()

	testCases := []TestCase{
		{
			name: "happy case",
			req: []*yasuo_entities.Speeches{
				{
					Speeches: entities.Speeches{
						SpeechID: database.Text("speech_id_1"),
						Link:     database.Text("link_1"),
						Type:     database.Text("type_1"),
						QuizID:   database.Text("quiz_id_1"),
						Sentence: database.Text("sentence_1"),
						Settings: database.JSONB("{}"),
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("QueryRow").Once().Return(row, nil)
				row.On("Scan", values...).Once().Return(nil)
				batchResults.On("Close").Once().Return(nil)
			},
		}, {
			name: "error send batch",
			req: []*yasuo_entities.Speeches{
				{
					Speeches: entities.Speeches{
						SpeechID: database.Text("speech_id_1"),
						Link:     database.Text("link_1"),
						Type:     database.Text("type_1"),
						QuizID:   database.Text("quiz_id_1"),
						Sentence: database.Text("sentence_1"),
						Settings: database.JSONB("{}"),
					},
				},
				{
					Speeches: entities.Speeches{
						SpeechID: database.Text("speech_id_2"),
						Link:     database.Text("link_2"),
						Type:     database.Text("type_2"),
						QuizID:   database.Text("quiz_id_2"),
						Sentence: database.Text("sentence_2"),
						Settings: database.JSONB("{}"),
					},
				},
				{
					Speeches: entities.Speeches{
						SpeechID: database.Text("speech_id_3"),
						Link:     database.Text("link_3"),
						Type:     database.Text("type_3"),
						QuizID:   database.Text("quiz_id_3"),
						Sentence: database.Text("sentence_3"),
						Settings: database.JSONB("{}"),
					},
				},
				{
					Speeches: entities.Speeches{
						SpeechID: database.Text("speech_id_4"),
						Link:     database.Text("link_4"),
						Type:     database.Text("type_4"),
						QuizID:   database.Text("quiz_id_4"),
						Sentence: database.Text("sentence_4"),
						Settings: database.JSONB("{}"),
					},
				},
			},
			expectedErr: fmt.Errorf("batchResults.QueryRow: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				row := &mock_database.Row{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("QueryRow").Once().Return(row)
				row.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
				batchResults.On("QueryRow").Once().Return(row)
				row.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
				batchResults.On("QueryRow").Once().Return(row)
				row.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(pgx.ErrTxClosed)
				batchResults.On("QueryRow").Once().Return(row)
				row.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)

		_, err := speeches.UpsertSpeeches(
			ctx,
			db,
			testCase.req.([]*yasuo_entities.Speeches),
		)
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestSpeechRepository_CheckExistedSpeech(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	row := &mock_database.Row{}
	speeches := &SpeechesRepository{}

	testCases := []TestCase{
		{
			name: "happy case",
			req: &CheckExistedSpeechReq{
				Text:   database.Text("test text"),
				Config: database.JSONB(`{"language": "en-US"}`),
			},
			expectedResp: true,
			setup: func(ctx context.Context) {
				db.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(row)
				row.On("Scan", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*pgtype.Text)
					arg.String = "test"
					arg.Status = pgtype.Present
				})
			},
		},
		{
			name: "scan error",
			req: &CheckExistedSpeechReq{
				Text:   database.Text("test text"),
				Config: database.JSONB(`{"language": "en-US"}`),
			},
			expectedResp: false,
			setup: func(ctx context.Context) {
				db.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(row)
				row.On("Scan", mock.Anything).Once().Return(fmt.Errorf("scan error"))
			},
		},
		{
			name: "not found link error",
			req: &CheckExistedSpeechReq{
				Text:   database.Text("test text"),
				Config: database.JSONB(`{"language": "en-US"}`),
			},
			expectedResp: false,
			setup: func(ctx context.Context) {
				db.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(row)
				row.On("Scan", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*pgtype.Text)
					arg.String = ""
					arg.Status = pgtype.Present
				})
			},
		},
		{
			name: "return null error",
			req: &CheckExistedSpeechReq{
				Text:   database.Text("test text"),
				Config: database.JSONB(`{"language": "en-US"}`),
			},
			expectedResp: false,
			setup: func(ctx context.Context) {
				db.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(row)
				row.On("Scan", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*pgtype.Text)
					arg.String = ""
					arg.Status = pgtype.Null
				})
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)

		ok, _ := speeches.CheckExistedSpeech(
			ctx,
			db,
			testCase.req.(*CheckExistedSpeechReq),
		)
		assert.Equal(t, testCase.expectedResp, ok)
	}
}

func TestSpeechesRepository_RetrieveAllSpeaches(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	rows := &mock_database.Rows{}
	speeches := &SpeechesRepository{}

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Scan", mock.Anything).Return([]*yasuo_entities.Speeches{}, nil).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*pgtype.Text)
					arg.String = "test"
					arg.Status = pgtype.Present
				})

				rows.On("Next").Once().Return(false)

				rows.On("Close").Once().Return(nil)
				rows.On("Err").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := speeches.RetrieveAllSpeaches(
			ctx,
			db,
			database.Int8(10),
			database.Int8(0),
		)
		assert.NoError(t, err)
	}
}
