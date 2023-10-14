package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	entities "github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TopicRepoWithSqlMock() (*TopicRepo, *testutil.MockDB) {
	r := &TopicRepo{}
	return r, testutil.NewMockDB()
}

func TestTopicRepo_RetrieveByIDs(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	topicRepo := &TopicRepo{}
	testCases := []TestCase{
		{
			name:        "retrieve error",
			req:         database.TextArray([]string{"topic-id"}),
			expectedErr: fmt.Errorf("database.Select: %w", fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed)),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "find success",
			req:         database.TextArray([]string{"topic-id"}),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				fields := database.GetFieldNames(&entities.Topic{})
				fieldDescriptions := make([]pgproto3.FieldDescription, 0, len(fields))
				for _, f := range fields {
					fieldDescriptions = append(fieldDescriptions, pgproto3.FieldDescription{Name: []byte(f)})
				}
				p := new(entities.Topic)

				mockDB.DB.On("Query", mock.Anything, mock.Anything).Once().Return()
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("FieldDescriptions").Return(fieldDescriptions)
				mockDB.Rows.On("Close").Once().Return(nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", database.GetScanFields(p, fields)...).Once().Return(nil)
				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.(pgtype.TextArray)
		_, err := topicRepo.RetrieveByIDs(ctx, mockDB.DB, req)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestTopicRepo_FindByBookIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	topicRepo := &TopicRepo{}

	t.Run("retrieve error", func(t *testing.T) {
		bookIDs := database.TextArray([]string{"book-id"})
		topicIDs := database.TextArray([]string{"topic-id"})
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything,
			mock.AnythingOfType("string"),
			&bookIDs,
			mock.Anything,
			mock.Anything,
		)

		groups, err := topicRepo.FindByBookIDs(ctx, mockDB.DB, bookIDs, topicIDs, pgtype.Int4{Status: pgtype.Null}, pgtype.Int4{Status: pgtype.Null})
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Nil(t, groups)
	})

	t.Run("find success", func(tt *testing.T) {
		bookIDs := database.TextArray([]string{"book-id"})
		topicIDs := database.TextArray([]string{"topic-id"})
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			&bookIDs,
			mock.Anything,
			mock.Anything,
		)
		topic := &entities.Topic{}
		topicFields, topicValues := topic.FieldMap()
		for i, topicField := range topicFields {
			topicFields[i] = "t." + topicField
		}
		var fields []string
		fields = append(fields, topicFields...)
		var values []interface{}
		values = append(values, topicValues...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := topicRepo.FindByBookIDs(ctx, mockDB.DB, bookIDs, topicIDs, pgtype.Int4{Status: pgtype.Null}, pgtype.Int4{Status: pgtype.Null})
		assert.NoError(tt, err)
	})
}

func TestTopicRepo_FindByIDsV2(t *testing.T) {
	r, db := TopicRepoWithSqlMock()
	t.Parallel()

	type Args struct {
		IDs   []string
		IsAll bool
	}

	ids := []string{"1", "2", "3"}
	topics := []*entities.Topic{
		{
			ID: database.Text("1"),
		},
		{
			ID: database.Text("2"),
		},
		{
			ID: database.Text("3"),
		},
	}
	m := map[string]*entities.Topic{
		"1": topics[0],
		"2": topics[1],
		"3": topics[2],
	}
	testCases := []TestCase{
		{
			name: "happy case",
			req: Args{
				IDs:   ids,
				IsAll: false,
			},
			setup: func(ctx context.Context) {
				db.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.TextArray(ids))
				fields, _ := topics[0].FieldMap()
				valuesArray := make([][]interface{}, 0)
				for _, e := range topics {
					_, values := e.FieldMap()
					valuesArray = append(valuesArray, values)
				}
				db.MockScanArray(nil, fields, valuesArray)
			},
			expectedErr:  nil,
			expectedResp: m,
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.(Args)

		resp, err := r.FindByIDsV2(ctx, db.DB, req.IDs, req.IsAll)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedResp, resp)
		}
	}
}

func TestTopicRepo_SoftDelete(t *testing.T) {
	r, db := TopicRepoWithSqlMock()
	t.Parallel()

	ids := []string{"1", "2", "3"}
	testCases := []TestCase{
		{
			name: "happy case",
			req:  ids,
			setup: func(ctx context.Context) {
				db.MockExecArgs(
					t,
					pgconn.CommandTag("3"),
					nil,
					mock.Anything,
					"UPDATE topics SET deleted_at = now(), updated_at = now() WHERE topic_id = ANY($1::_TEXT) AND deleted_at IS NULL",
					database.TextArray(ids),
				)
			},
			expectedErr:  nil,
			expectedResp: 3,
		},
		{
			name: "update error",
			req:  ids,
			setup: func(ctx context.Context) {
				db.MockExecArgs(
					t,
					pgconn.CommandTag("1"),
					puddle.ErrClosedPool,
					mock.Anything,
					mock.Anything,
					database.TextArray(ids),
				)
			},
			expectedErr: puddle.ErrClosedPool,
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.([]string)

		resp, err := r.SoftDelete(ctx, db.DB, req)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedResp, resp)
		}
	}
}

func TestTopicRepo_DuplicateTopics(t *testing.T) {
	t.Parallel()
	r, db := TopicRepoWithSqlMock()
	type duplicateTopicRequest struct {
		ChapterIDs    []string
		NewChapterIDs []string
	}
	rows := &mock_database.Rows{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: &duplicateTopicRequest{
				ChapterIDs:    []string{"chapter-id-1"},
				NewChapterIDs: []string{"new-chapter-id-1"},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Close").Once().Return(nil)
				batchResults.On("Query").Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything, mock.Anything).Return(nil)
				rows.On("Err").Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name: "error send batch",
			req: &duplicateTopicRequest{
				ChapterIDs:    []string{"chapter-id-1"},
				NewChapterIDs: []string{"new-chapter-id-1"},
			},
			expectedErr: fmt.Errorf("tx is closed"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Close").Once().Return(nil)
				batchResults.On("Query").Return(rows, pgx.ErrTxClosed)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.(*duplicateTopicRequest)
		_, err := r.DuplicateTopics(ctx, db.DB, database.TextArray(req.ChapterIDs), database.TextArray(req.NewChapterIDs))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
