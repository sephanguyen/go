package repositories

import (
	"context"
	"testing"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func userGroupRepoWithMockSQL() (*UserGroupRepo, *testutil.MockDB) {
	r := &UserGroupRepo{}
	return r, testutil.NewMockDB()
}

func TestUserGroupRepo_Find(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := userGroupRepoWithMockSQL()
	t.Run("find error", func(t *testing.T) {
		id := database.Text("id")
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything,
			mock.AnythingOfType("string"),
			&id,
		)

		groups, err := r.Find(ctx, mockDB.DB, id)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Nil(t, groups)
	})

	t.Run("find success", func(tt *testing.T) {
		id := database.Text("id")
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			&id,
		)

		e := &entities_bob.UserGroup{}
		fields, values := e.FieldMap()

		e.UserID.Set(ksuid.New().String())
		e.GroupID.Set(entities_bob.UserGroupAdmin)

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		groups, err := r.Find(ctx, mockDB.DB, id)
		assert.NoError(tt, err)
		assert.Equal(tt, []*entities_bob.UserGroup{e}, groups)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"user_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}
