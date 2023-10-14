package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func topicsLearningObjectivesRepoWithMockSQL() (*TopicsLearningObjectivesRepo, *testutil.MockDB) {
	r := &TopicsLearningObjectivesRepo{}
	return r, testutil.NewMockDB()
}

func TestTopicsLearningObjectivesRepo_RetrieveByLoIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := topicsLearningObjectivesRepoWithMockSQL()
	t.Run("retrieve error", func(t *testing.T) {
		loIDs := database.TextArray([]string{"1", "2", "3"})
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything,
			mock.AnythingOfType("string"),
			loIDs,
		)

		groups, err := r.RetrieveByLoIDs(ctx, mockDB.DB, loIDs)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Nil(t, groups)
	})

	t.Run("find success", func(tt *testing.T) {
		loIDs := database.TextArray([]string{"1", "2", "3"})
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			loIDs,
		)

		topic := &entities.Topic{}
		topicFields, topicValues := topic.FieldMap()
		lo := &entities.LearningObjective{}
		loFields, loValues := lo.FieldMap()
		tlo := &entities.TopicsLearningObjectives{}

		for i, topicField := range topicFields {
			topicFields[i] = "t." + topicField
		}

		for i, loField := range loFields {
			loFields[i] = "lo." + loField
		}

		var fields []string
		fields = append(fields, topicFields...)
		fields = append(fields, "tlo.created_at", "tlo.updated_at", "tlo.display_order")
		fields = append(fields, loFields...)

		var values []interface{}
		values = append(values, topicValues...)
		values = append(values, &tlo.CreatedAt, &tlo.UpdatedAt, &tlo.DisplayOrder)
		values = append(values, loValues...)

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.RetrieveByLoIDs(ctx, mockDB.DB, loIDs)
		assert.NoError(tt, err)
	})
}

func TestTopicsLearningObjectivesRepo_BulkImport(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	topicLearningObjectivesRepo := &TopicsLearningObjectivesRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.TopicsLearningObjectives{
				{
					TopicID:      database.Text("topic-id-1"),
					LoID:         database.Text("lo-id-1"),
					DisplayOrder: database.Int2(1),
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
			req: []*entities.TopicsLearningObjectives{
				{
					TopicID:      database.Text("topic-id-1"),
					LoID:         database.Text("lo-id-1"),
					DisplayOrder: database.Int2(1),
				},
				{
					TopicID:      database.Text("topic-id-2"),
					LoID:         database.Text("lo-id-2"),
					DisplayOrder: database.Int2(1),
				},
				{
					TopicID:      database.Text("topic-id-3"),
					LoID:         database.Text("lo-id-3"),
					DisplayOrder: database.Int2(1),
				},
				{
					TopicID:      database.Text("topic-id-4"),
					LoID:         database.Text("lo-id-4"),
					DisplayOrder: database.Int2(1),
				},
				{
					TopicID:      database.Text("topic-id-5"),
					LoID:         database.Text("lo-id-1"),
					DisplayOrder: database.Int2(5),
				},
			},
			expectedErr: fmt.Errorf("batchResults.Exec: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, pgx.ErrTxClosed)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := topicLearningObjectivesRepo.BulkImport(ctx, db, testCase.req.([]*entities.TopicsLearningObjectives))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}

	return
}

func TestTopicsLearningObjective_BulkUpdateDisplayOrder(t *testing.T) {
	t.Parallel()
	type BulkUpdateInput struct {
		TopicLearningObjectives []*entities.TopicsLearningObjectives
	}
	db := &mock_database.QueryExecer{}
	TopicLearningObjectiveRepo := &TopicsLearningObjectivesRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: &BulkUpdateInput{
				TopicLearningObjectives: []*entities.TopicsLearningObjectives{
					{
						TopicID: database.Text("mock-topic-1"),
					},
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
			req: &BulkUpdateInput{
				TopicLearningObjectives: []*entities.TopicsLearningObjectives{
					{
						TopicID: database.Text("mock-topic-1"),
					},
					{
						TopicID: database.Text("mock-topic-1"),
					},
				},
			},
			expectedErr: fmt.Errorf("batchResults.Exec: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, pgx.ErrTxClosed)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			input := testCase.req.(*BulkUpdateInput)
			err := TopicLearningObjectiveRepo.BulkUpdateDisplayOrder(ctx, db, input.TopicLearningObjectives)
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}
