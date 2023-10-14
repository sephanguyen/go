package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTopicBulkImport_Batch(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	topicRepo := &TopicRepo{}
	// query := "INSERT INTO VALUES"
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.Topic{
				{
					ID: pgtype.Text{String: "1", Status: pgtype.Present},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "error send batch",
			req: []*entities.Topic{
				{
					ID: pgtype.Text{String: "1", Status: pgtype.Present},
				}, {
					ID: pgtype.Text{String: "2", Status: pgtype.Present},
				},
			},
			expectedErr: errors.New("batchResults.Exec: closed pool"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := topicRepo.BulkImport(ctx, db, testCase.req.([]*entities.Topic))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}

	return
}

func TestTopicRepo_UpdateTotalLOs(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	topicID := database.Text("topic-id")
	topicRepo := &TopicRepo{}
	testCases := []TestCase{
		{
			name:        "happy case",
			req:         topicID,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().
					Return(cmdTag, nil)
			},
		},
		{
			name:        "error send batch",
			req:         topicID,
			expectedErr: errors.Wrap(pgx.ErrNoRows, "db.Exec"),
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().
					Return(cmdTag, pgx.ErrNoRows)
			},
		},
		{
			name:        "rows affected != 1",
			req:         topicID,
			expectedErr: errors.Errorf("cannot update total_los for topic: %s", topicID.String),
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`2`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().
					Return(cmdTag, nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := topicRepo.UpdateTotalLOs(ctx, db, topicID)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
	return
}

func TestTopicRepo_UpdateStatus(t *testing.T) {
}

func TestTopicRepo_BulkUpsertWithoutDisplayOrder(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	topicRepo := &TopicRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.Topic{
				{
					ID:       database.Text("mock-topic-id-1"),
					Name:     database.Text("mock-topic-name-1"),
					SchoolID: database.Int4(1),
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "error send batch",
			req: []*entities.Topic{
				{
					ID: database.Text("mock-topic-id-1"),
				},
				{
					ID: database.Text("mock-topic-id-2"),
				},
				{
					ID: database.Text("mock-topic-id-3"),
				},
			},
			expectedErr: fmt.Errorf("batchResults.Exec: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, pgx.ErrTxClosed)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			err := topicRepo.BulkUpsertWithoutDisplayOrder(ctx, db, testCase.req.([]*entities.Topic))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			}
		})
	}
}

func TestTopicRepo_FindByBookID(t *testing.T) {
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

func Test_RetrieveBookTopic(t *testing.T) {
	t.Parallel()
	type TestCase struct {
		name         string
		input1       pgtype.TextArray
		expectedResp interface{}
		expectedErr  error
		setup        func(ctx context.Context)
	}
	mockDB := testutil.NewMockDB()
	rows := mockDB.Rows
	r := &TopicRepo{}
	e := &entities.Topic{}
	fields, _ := e.FieldMap()
	scanFields := database.GetScanFields(e, fields)
	var (
		bookID pgtype.Text
	)
	scanFields = append(scanFields, &bookID)

	topicIDs := database.TextArray([]string{"topic-1", "topic-2", "topic-3"})
	testCases := []TestCase{
		{
			name:        "happy case",
			input1:      topicIDs,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &topicIDs)
				mockDB.DB.On("Query").Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", scanFields...).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "err query",
			input1:      topicIDs,
			expectedErr: fmt.Errorf("row.Err: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &topicIDs)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(pgx.ErrNoRows)
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.setup(ctx)
			_, err := r.RetrieveBookTopic(ctx, mockDB.DB, tc.input1)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}
