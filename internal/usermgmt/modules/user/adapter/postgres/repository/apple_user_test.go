package repository

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func AppleUserRepoWithSqlMock() (*AppleUserRepo, *testutil.MockDB) {
	r := &AppleUserRepo{}
	return r, testutil.NewMockDB()
}

func TestAppleUserRepo_Insert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userRepo, mockDB := AppleUserRepoWithSqlMock()

	t.Run("err insert", func(t *testing.T) {
		user := &entity.AppleUser{}
		_, values := user.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		err := userRepo.Create(ctx, mockDB.DB, user)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		user := &entity.AppleUser{}
		fields, values := user.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := userRepo.Create(ctx, mockDB.DB, user)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertInsertedTable(t, user.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})
}

func TestAppleUserRepo_Get(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := AppleUserRepoWithSqlMock()

	userID := idutil.ULIDNow()
	pgUserID := database.Text(userID)

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			&pgUserID,
		)

		e := &entity.AppleUser{}
		fields, values := e.FieldMap()
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)

		results, err := r.Get(ctx, mockDB.DB, pgUserID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, results)
	})

	t.Run("scan field row success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			&pgUserID,
		)

		e := &entity.AppleUser{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		results, err := r.Get(ctx, mockDB.DB, pgUserID)
		assert.Nil(t, err)
		assert.Equal(t, e, results)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"user_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}
