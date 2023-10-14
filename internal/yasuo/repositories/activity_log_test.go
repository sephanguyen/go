package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ActivityLogRepoWithSqlMock() (*ActivityLogRepo, *testutil.MockDB) {
	r := &ActivityLogRepo{}
	return r, testutil.NewMockDB()
}

func TestActivityLogRepo_CreateV2(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := ActivityLogRepoWithSqlMock()

	t.Run("insert error", func(t *testing.T) {
		e := &bob_entities.ActivityLog{}
		_, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		err := repo.CreateV2(ctx, mockDB.DB, e)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("no rows effected", func(t *testing.T) {
		e := &bob_entities.ActivityLog{}
		_, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		err := repo.CreateV2(ctx, mockDB.DB, e)
		assert.EqualError(t, err, "cannot insert new ActivityLog")
	})

	t.Run("insert success", func(t *testing.T) {
		e := &bob_entities.ActivityLog{}
		fields, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := repo.CreateV2(ctx, mockDB.DB, e)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})
}

func TestActivityLogRepo_Upsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	activityRepo := &ActivityLogRepo{}

	testCases := []TestCase{
		{
			name: "happy case",
			req: []*bob_entities.ActivityLog{
				{
					ID:         database.Text("activity-log-1"),
					UserID:     database.Text("user-id-1"),
					ActionType: database.Text("action-type-1"),
					CreatedAt:  database.Timestamptz(time.Now()),
					UpdatedAt:  database.Timestamptz(time.Now()),
					Payload:    database.JSONB("{}"),
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "batch error",
			req: []*bob_entities.ActivityLog{
				{
					ID:         database.Text("activity-log-1"),
					UserID:     database.Text("user-id-1"),
					ActionType: database.Text("action-type-1"),
					CreatedAt:  database.Timestamptz(time.Now()),
					UpdatedAt:  database.Timestamptz(time.Now()),
					Payload:    database.JSONB("{}"),
				},
			},
			expectedErr: fmt.Errorf("batchResults.Exec: %w", fmt.Errorf("batch error")),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, fmt.Errorf("batch error"))
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)

		err := activityRepo.Upsert(ctx, db, testCase.req.([]*bob_entities.ActivityLog))
		assert.Equal(t, testCase.expectedErr, err)
	}
}
