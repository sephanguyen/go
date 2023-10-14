package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/zeus/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestCase struct {
	name         string
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func ActivityLogRepoWithSqlMock() (*ActivityLogRepo, *testutil.MockDB) {
	r := &ActivityLogRepo{}
	return r, testutil.NewMockDB()
}

func TestActivityLogRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := ActivityLogRepoWithSqlMock()

	t.Run("insert error", func(t *testing.T) {
		e := &entities.ActivityLog{}
		_, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		err := repo.Create(ctx, mockDB.DB, e)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("no rows effected", func(t *testing.T) {
		e := &entities.ActivityLog{}
		_, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		err := repo.Create(ctx, mockDB.DB, e)
		assert.EqualError(t, err, "cannot insert new ActivityLog")
	})

	t.Run("insert success", func(t *testing.T) {
		e := &entities.ActivityLog{}
		fields, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := repo.Create(ctx, mockDB.DB, e)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})
}

func TestActivityLogRepo_Bulk(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	activityLogRepo := &ActivityLogRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.ActivityLog{
				{
					ID:         database.Text("id"),
					UserID:     database.Text("user-id"),
					ActionType: database.Text("action-type"),
					CreatedAt:  database.Timestamptz(time.Now()),
					UpdatedAt:  database.Timestamptz(time.Now()),
					Payload:    database.JSONB("{}"),
					DeletedAt:  database.Timestamptz(time.Now()),
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
			req: []*entities.ActivityLog{
				{
					ID:         database.Text("id-1"),
					UserID:     database.Text("user-id-1"),
					ActionType: database.Text("action-type-1"),
					CreatedAt:  database.Timestamptz(time.Now()),
					UpdatedAt:  database.Timestamptz(time.Now()),
					Payload:    database.JSONB("{}"),
					DeletedAt:  database.Timestamptz(time.Now()),
				},
				{
					ID:         database.Text("id-2"),
					UserID:     database.Text("user-id-2"),
					ActionType: database.Text("action-type-2"),
					CreatedAt:  database.Timestamptz(time.Now()),
					UpdatedAt:  database.Timestamptz(time.Now()),
					Payload:    database.JSONB("{}"),
					DeletedAt:  database.Timestamptz(time.Now()),
				},
				{
					ID:         database.Text("id-3"),
					UserID:     database.Text("user-id-3"),
					ActionType: database.Text("action-type-3"),
					CreatedAt:  database.Timestamptz(time.Now()),
					UpdatedAt:  database.Timestamptz(time.Now()),
					Payload:    database.JSONB("{}"),
					DeletedAt:  database.Timestamptz(time.Now()),
				},
				{
					ID:         database.Text("id-4"),
					UserID:     database.Text("user-id-4"),
					ActionType: database.Text("action-type-4"),
					CreatedAt:  database.Timestamptz(time.Now()),
					UpdatedAt:  database.Timestamptz(time.Now()),
					Payload:    database.JSONB("{}"),
					DeletedAt:  database.Timestamptz(time.Now()),
				},
				{
					ID:         database.Text("id-5"),
					UserID:     database.Text("user-id-5"),
					ActionType: database.Text("action-type-5"),
					CreatedAt:  database.Timestamptz(time.Now()),
					UpdatedAt:  database.Timestamptz(time.Now()),
					Payload:    database.JSONB("{}"),
					DeletedAt:  database.Timestamptz(time.Now()),
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
		err := activityLogRepo.CreateBulk(ctx, db, testCase.req.([]*entities.ActivityLog))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}

	return
}
