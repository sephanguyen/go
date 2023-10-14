package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	repositories_bob "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TopicRepoWithSqlMock() (*TopicRepo, *testutil.MockDB) {
	r := &TopicRepo{}
	return r, testutil.NewMockDB()
}

func TestTopicRepo_FindSchoolIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := TopicRepoWithSqlMock()
	ids := []string{"id", "id-1"}

	pgIDs := database.TextArray([]string{"id", "id-1"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &pgIDs)

		schoolIDs, err := r.FindSchoolIDs(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, schoolIDs)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &pgIDs)

		e := &repositories_bob.EnSchoolID{}
		fields, values := e.FieldMap()
		e.SchoolID = 1

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		schoolIDs, err := r.FindSchoolIDs(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.Equal(t, []int32{e.SchoolID}, schoolIDs)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, "topics", "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"topic_id":   {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestTopicRepo_SoftDeleteV3(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := TopicRepoWithSqlMock()

	topicIDs := database.TextArray([]string{"mock-topic-id-1", "mock-topic-id-2"})

	t.Run("err update", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &topicIDs)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		_, err := r.SoftDeleteV3(ctx, mockDB.DB, topicIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &topicIDs)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		_, err := r.SoftDeleteV3(ctx, mockDB.DB, topicIDs)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertUpdatedTable(t, "topics")
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at", "updated_at")
	})
}
