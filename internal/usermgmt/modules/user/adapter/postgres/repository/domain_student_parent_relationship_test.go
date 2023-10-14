package repository

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
	"github.com/pkg/errors"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func DomainDomainStudentParentRelationshipRepoWithSqlMock() (*DomainStudentParentRelationshipRepo, *testutil.MockDB) {
	r := &DomainStudentParentRelationshipRepo{}
	return r, testutil.NewMockDB()
}

func TestDomainStudentParentRelationshipRepo_SoftDeleteByStudentIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := DomainDomainStudentParentRelationshipRepoWithSqlMock()
	studentIDs := []string{"userID-1", "userID-2"}

	t.Run("err update", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, database.TextArray(studentIDs))
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := repo.SoftDeleteByStudentIDs(ctx, mockDB.DB, studentIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, database.TextArray(studentIDs))
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := repo.SoftDeleteByStudentIDs(ctx, mockDB.DB, studentIDs)
		assert.Nil(t, err)

		// move primaryField to the last
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"student_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestDomainStudentParentRelationshipRepo_SoftDeleteByParentIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := DomainDomainStudentParentRelationshipRepoWithSqlMock()
	parentIDs := []string{"userID-1", "userID-2"}

	t.Run("err update", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, database.TextArray(parentIDs))
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := repo.SoftDeleteByParentIDs(ctx, mockDB.DB, parentIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, database.TextArray(parentIDs))
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := repo.SoftDeleteByParentIDs(ctx, mockDB.DB, parentIDs)
		assert.Nil(t, err)

		// move primaryField to the last
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"parent_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestDomainStudentParentRelationshipRepo_GetByStudentIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentIDs := []string{uuid.NewString()}
	studentParent := &DomainStudentParentRelationship{}
	_, fieldMaps := studentParent.FieldMap()
	argsStudentParent := append([]interface{}{}, genSliceMock(len(fieldMaps))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := DomainDomainStudentParentRelationshipRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(studentIDs)).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsStudentParent...).Once().Return(nil)

		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		studentPackages, err := repo.GetByStudentIDs(ctx, mockDB.DB, studentIDs)
		assert.Nil(t, err)
		assert.NotNil(t, studentPackages)
	})

	t.Run("db Query returns error", func(t *testing.T) {
		repo, mockDB := DomainDomainStudentParentRelationshipRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(studentIDs)).Once().Return(mockDB.Rows, pgx.ErrTxClosed)

		studentPackages, err := repo.GetByStudentIDs(ctx, mockDB.DB, studentIDs)
		assert.NotNil(t, err)
		assert.Nil(t, studentPackages)
	})

	t.Run("rows Scan returns error", func(t *testing.T) {
		repo, mockDB := DomainDomainStudentParentRelationshipRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(studentIDs)).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsStudentParent...).Once().Return(pgx.ErrTxClosed)
		mockDB.Rows.On("Close").Once().Return(nil)

		studentPackages, err := repo.GetByStudentIDs(ctx, mockDB.DB, studentIDs)
		assert.NotNil(t, err)
		assert.Nil(t, studentPackages)
	})
}

func TestDomainStudentParentRelationshipRepo_GetByParentIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	parentIDs := []string{uuid.NewString()}
	studentParent := &DomainStudentParentRelationship{}
	_, fieldMaps := studentParent.FieldMap()
	argsStudentParent := append([]interface{}{}, genSliceMock(len(fieldMaps))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := DomainDomainStudentParentRelationshipRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(parentIDs)).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsStudentParent...).Once().Return(nil)

		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		studentPackages, err := repo.GetByParentIDs(ctx, mockDB.DB, parentIDs)
		assert.Nil(t, err)
		assert.NotNil(t, studentPackages)
	})

	t.Run("db Query returns error", func(t *testing.T) {
		repo, mockDB := DomainDomainStudentParentRelationshipRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(parentIDs)).Once().Return(mockDB.Rows, pgx.ErrTxClosed)

		studentPackages, err := repo.GetByParentIDs(ctx, mockDB.DB, parentIDs)
		assert.NotNil(t, err)
		assert.Nil(t, studentPackages)
	})

	t.Run("rows Scan returns error", func(t *testing.T) {
		repo, mockDB := DomainDomainStudentParentRelationshipRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(parentIDs)).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsStudentParent...).Once().Return(pgx.ErrTxClosed)
		mockDB.Rows.On("Close").Once().Return(nil)

		studentPackages, err := repo.GetByParentIDs(ctx, mockDB.DB, parentIDs)
		assert.NotNil(t, err)
		assert.Nil(t, studentPackages)
	})
}
