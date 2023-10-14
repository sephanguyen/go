package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func MasterMgmtClassStudentRepoWithSqlMock() (*MasterMgmtClassStudentRepo, *testutil.MockDB) {
	r := &MasterMgmtClassStudentRepo{}
	return r, testutil.NewMockDB()
}

func TestMasterMgmtClassStudent_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := MasterMgmtClassStudentRepoWithSqlMock()

	t.Run("should upsert success", func(t *testing.T) {
		// Arrange
		e := &entities.ClassStudent{}
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
		e := &entities.ClassStudent{}
		fields, values := e.FieldMap()
		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		// Action
		err := r.Upsert(ctx, mockDB.DB, e)

		// Assert
		assert.Equal(t, fmt.Errorf("cannot upsert class student"), err)
		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})

	t.Run("should throw error pgsql", func(t *testing.T) {
		// Arrange
		e := &entities.ClassStudent{}
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

func TestMasterMgmtClassStudent_SoftDelete(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := MasterMgmtClassStudentRepoWithSqlMock()

	t.Run("should success", func(t *testing.T) {
		// Arrange
		e := &entities.ClassStudent{}
		fields, _ := e.FieldMap()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string")})
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		// Action
		err := r.SoftDelete(ctx, mockDB.DB, []string{e.StudentID.String}, []string{e.ClassID.String})

		// Assert
		assert.Nil(t, err)
		mockDB.RawStmt.AssertUpdatedTable(t, e.TableName())
		mockDB.RawStmt.AssertUpdatedFields(t, fields[4:]...)
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"student_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"class_id":   {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 2}},
		})
	})

	t.Run("should throw error pgsql", func(t *testing.T) {
		// Arrange
		e := &entities.ClassStudent{}
		fields, _ := e.FieldMap()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"))
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		// Action
		err := r.SoftDelete(ctx, mockDB.DB, []string{e.StudentID.String}, []string{e.ClassID.String})

		// Assert
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		mockDB.RawStmt.AssertUpdatedTable(t, e.TableName())
		mockDB.RawStmt.AssertUpdatedFields(t, fields[4:]...)
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"student_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"class_id":   {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 2}},
		})
	})
}
