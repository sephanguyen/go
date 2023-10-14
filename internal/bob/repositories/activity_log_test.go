package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ActivityLogRepoWithSqlMock() (*ActivityLogRepo, *testutil.MockDB) {
	r := &ActivityLogRepo{}
	return r, testutil.NewMockDB()
}

func TestActivityLogRepo_Insert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ActivityLogRepoWithSqlMock()

	t.Run("err insert", func(t *testing.T) {
		e := &entities_bob.ActivityLog{}
		_, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrNotAvailable, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.True(t, errors.Is(err, puddle.ErrNotAvailable))
	})

	t.Run("no rows affected", func(t *testing.T) {
		e := &entities_bob.ActivityLog{}
		_, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.EqualError(t, err, "cannot insert new ActivityLogRepo")
	})

	t.Run("success", func(t *testing.T) {
		e := &entities_bob.ActivityLog{}
		fields, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})
}

func TestActivityLogRepo_RetrieveLastCheckPromotionCode(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ActivityLogRepoWithSqlMock()

	ID := idutil.ULIDNow()
	studentID := database.Text(ID)

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			&studentID,
		)

		e := &entities_bob.ActivityLog{}
		fields, values := e.FieldMap()
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)

		results, err := r.RetrieveLastCheckPromotionCode(ctx, mockDB.DB, studentID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, results, "")
	})

	t.Run("scan field row success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			&studentID,
		)

		e := &entities_bob.ActivityLog{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		_, err := r.RetrieveLastCheckPromotionCode(ctx, mockDB.DB, studentID)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}

func TestActivityLogRepo_BulkImport(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	activityLogRepo := &ActivityLogRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities_bob.ActivityLog{
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
			req: []*entities_bob.ActivityLog{
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
		err := activityLogRepo.BulkImport(ctx, db, testCase.req.([]*entities_bob.ActivityLog))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}

	return
}
