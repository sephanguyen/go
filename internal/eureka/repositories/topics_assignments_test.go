package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func topicsAssignmentRepoWithMockSQL() (*TopicsAssignmentsRepo, *testutil.MockDB) {
	r := &TopicsAssignmentsRepo{}
	return r, testutil.NewMockDB()
}

func TestTopicsAssignmentsRepo_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := topicsAssignmentRepoWithMockSQL()
	ta := &entities.TopicsAssignments{
		TopicID:      database.Text("topic-1"),
		AssignmentID: database.Text("assignment-1"),
	}

	_, values := ta.FieldMap()

	t.Run("err update", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		err := r.Upsert(ctx, mockDB.DB, ta)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("no rows affected", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		err := r.Upsert(ctx, mockDB.DB, ta)
		assert.EqualError(t, err, fmt.Errorf("no rows were affected").Error())
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("3"), nil, args...)

		err := r.Upsert(ctx, mockDB.DB, ta)
		assert.Nil(t, err)
	})
}

func TestTopicsAssignmentsRepo_SoftDeleteByAssignmentIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := topicsAssignmentRepoWithMockSQL()

	taIDs := database.TextArray([]string{"ta-id-1", "ta-id-2"})

	t.Run("err update", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, taIDs)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)
		err := r.SoftDeleteByAssignmentIDs(ctx, mockDB.DB, taIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
	t.Run("no rows affected", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, taIDs)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)
		err := r.SoftDeleteByAssignmentIDs(ctx, mockDB.DB, taIDs)
		assert.EqualError(t, err, fmt.Errorf("no rows were affected").Error())
	})
	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, taIDs)
		mockDB.MockExecArgs(t, pgconn.CommandTag("3"), nil, args...)
		err := r.SoftDeleteByAssignmentIDs(ctx, mockDB.DB, taIDs)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertUpdatedTable(t, "topics_assignments")
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at")
	})
}

func TestTopicsAssignments_RetrieveByAssignmentIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := topicsAssignmentRepoWithMockSQL()
	assignmentIDs := []string{"id-1", "id-2"}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &assignmentIDs)

		courseIDs, err := r.RetrieveByAssignmentIDs(ctx, mockDB.DB, assignmentIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, courseIDs)
	})

	t.Run("success with select", func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &assignmentIDs)
		mockDB.DB.On("Query").Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)

		_, err := r.RetrieveByAssignmentIDs(ctx, mockDB.DB, assignmentIDs)
		assert.Nil(t, err)
	})
}

func TestTopicsAssignments_BulkUpdateDisplayOrder(t *testing.T) {
	t.Parallel()
	type BulkUpdateInput struct {
		TopicAssignments []*entities.TopicsAssignments
	}
	db := &mock_database.QueryExecer{}
	TopicAssignmentRepo := &TopicsAssignmentsRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: &BulkUpdateInput{
				TopicAssignments: []*entities.TopicsAssignments{
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
				TopicAssignments: []*entities.TopicsAssignments{
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
			err := TopicAssignmentRepo.BulkUpdateDisplayOrder(ctx, db, input.TopicAssignments)
			assert.Equal(t, testCase.expectedErr, err)
		})

	}
}
