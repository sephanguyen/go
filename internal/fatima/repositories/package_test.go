package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func PackageRepoWithSqlMock() (*PackageRepo, *testutil.MockDB) {
	r := &PackageRepo{}
	return r, testutil.NewMockDB()
}

func TestPackageRepo_Get(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := PackageRepoWithSqlMock()

	studentPackageID := ksuid.New().String()
	pgPackageID := database.Text(studentPackageID)

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
			mock.AnythingOfType("string"),
			&pgPackageID,
		)

		studentPackages, err := r.Get(ctx, mockDB.DB, pgPackageID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, studentPackages)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			&pgPackageID,
		)

		e := &entities.Package{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(idutil.ULIDNow())
		mockDB.MockScanFields(nil, fields, values)

		sPackage, err := r.Get(ctx, mockDB.DB, pgPackageID)
		assert.Nil(t, err)
		assert.Equal(t, e, sPackage)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"package_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestPackageRepo_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := PackageRepoWithSqlMock()

	t.Run("should upsert success", func(t *testing.T) {
		// Arrange
		e := &entities.Package{}
		fields, values := e.FieldMap()
		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		// Action
		err := r.Upsert(ctx, mockDB.DB, e)

		// Assert
		assert.Nil(t, err)
		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})

	t.Run("should throw error when no rows affected", func(t *testing.T) {
		// Arrange
		e := &entities.Package{}
		fields, values := e.FieldMap()
		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		// Action
		err := r.Upsert(ctx, mockDB.DB, e)

		// Assert
		assert.Equal(t, fmt.Errorf("cannot insert package"), err)
		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})

	t.Run("should throw error pgsql", func(t *testing.T) {
		// Arrange
		e := &entities.Package{}
		fields, values := e.FieldMap()
		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		// Action
		err := r.Upsert(ctx, mockDB.DB, e)

		// Assert
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})
}
