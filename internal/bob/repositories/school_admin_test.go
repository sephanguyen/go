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

func schoolAdminRepoWithMockSQL() (*SchoolAdminRepo, *testutil.MockDB) {
	r := &SchoolAdminRepo{}
	return r, testutil.NewMockDB()
}

func TestSchoolAdminRepo_Get(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := schoolAdminRepoWithMockSQL()
	t.Run("find error", func(t *testing.T) {
		id := database.Text("id")
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything,
			mock.AnythingOfType("string"),
			&id,
		)

		groups, err := r.Get(ctx, mockDB.DB, id)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Nil(t, groups)
	})

	t.Run("find success", func(tt *testing.T) {
		id := database.Text("id")
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			&id,
		)

		e := &entities_bob.SchoolAdmin{}
		fields, values := e.FieldMap()

		e.SchoolAdminID.Set(ksuid.New().String())
		e.SchoolID.Set(123)

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		admin, err := r.Get(ctx, mockDB.DB, id)
		assert.NoError(tt, err)
		assert.Equal(tt, e, admin)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"school_admin_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}
