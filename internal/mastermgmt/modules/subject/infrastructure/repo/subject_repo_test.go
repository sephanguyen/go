package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/subject/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func SubjectRepoWithSqlMock() (*SubjectRepo, *testutil.MockDB) {
	r := &SubjectRepo{}
	return r, testutil.NewMockDB()
}

func getRandomSubjects() []*domain.Subject {
	now := time.Now()
	s1 := &domain.Subject{
		SubjectID: idutil.ULIDNow(),
		Name:      "sub" + idutil.ULIDNow(),
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil,
	}
	s2 := &domain.Subject{
		Name:      "sub" + idutil.ULIDNow(),
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil,
	}
	subjects := []*domain.Subject{s1, s2}

	return subjects
}

func TestSubjectRepo_Import(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := SubjectRepoWithSqlMock()
	t.Run("successfully", func(t *testing.T) {
		ct := getRandomSubjects()
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)

		err := r.Import(ctx, mockDB.DB, ct)
		require.Nil(t, err)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		ct := getRandomSubjects()

		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)

		err := r.Import(ctx, mockDB.DB, ct)
		require.Error(t, err)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}

func TestSubjectRepo_GetByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := SubjectRepoWithSqlMock()
	ids := []string{"2", "3"}

	t.Run("select errors", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, database.TextArray(ids))

		subjects, err := r.GetByIDs(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, subjects)
	})

	t.Run("select succeed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.TextArray(ids))

		e := &Subject{
			SubjectID: database.Text(idutil.ULIDNow()),
		}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.GetByIDs(ctx, mockDB.DB, ids)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"subject_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestSubjectRepo_GetByNames(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := SubjectRepoWithSqlMock()
	ids := []string{"2", "3"}

	t.Run("select errors", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, database.TextArray(ids))

		subjects, err := r.GetByNames(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, subjects)
	})

	t.Run("select succeed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.TextArray(ids))

		e := &Subject{
			SubjectID: database.Text(idutil.ULIDNow()),
		}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.GetByNames(ctx, mockDB.DB, ids)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"name": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestSubjectRepo_GetAll(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := SubjectRepoWithSqlMock()

	t.Run("select failed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything)

		g, err := r.GetAll(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, g)
	})

	t.Run("select succeeded", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)

		e := &Subject{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.GetAll(ctx, mockDB.DB)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
		})
	})
}
