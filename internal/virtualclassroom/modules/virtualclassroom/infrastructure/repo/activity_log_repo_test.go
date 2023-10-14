package repo

import (
	"context"
	"testing"
	"time"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
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
	userID := "user-id1"
	actionType := bob_entities.LogActionTypePublish
	payload := map[string]interface{}{
		"lesson_id": "lesson-id1",
	}

	e := &bob_entities.ActivityLog{}
	fields, _ := e.FieldMap()
	values := genSliceMock(len(fields))

	t.Run("error", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, nil, puddle.ErrClosedPool, args...)

		err := repo.Create(ctx, mockDB.DB, userID, actionType, payload)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("no rows effected", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		err := repo.Create(ctx, mockDB.DB, userID, actionType, payload)
		assert.EqualError(t, err, "cannot insert new ActivityLog")
	})

	t.Run("successful", func(t *testing.T) {

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := repo.Create(ctx, mockDB.DB, userID, actionType, payload)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})
}
