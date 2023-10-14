package repo

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/grade/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func GradeRepoWithSqlMock() (*GradeRepo, *testutil.MockDB) {
	r := &GradeRepo{}
	return r, testutil.NewMockDB()
}

func TestGradeRepo_GetByPartnerInternalIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := GradeRepoWithSqlMock()
	ids := []string{"2", "3"}

	t.Run("select errors", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, database.TextArray(ids))

		grades, err := r.GetByPartnerInternalIDs(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, grades)
	})

	t.Run("select succeed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.TextArray(ids))

		e := &Grade{
			PartnerInternalID: database.Text(idutil.ULIDNow()),
		}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.GetByPartnerInternalIDs(ctx, mockDB.DB, ids)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"partner_internal_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestGradeRepo_Import(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := GradeRepoWithSqlMock()
	t.Run("successfully", func(t *testing.T) {
		g := getRandomGrades()
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)

		err := r.Import(ctx, mockDB.DB, g)
		require.Nil(t, err)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		g := getRandomGrades()

		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)

		err := r.Import(ctx, mockDB.DB, g)
		require.Error(t, err)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}

func TestGradeRepo_GetAll(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := GradeRepoWithSqlMock()

	t.Run("select failed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything)

		g, err := r.GetAll(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, g)
	})

	t.Run("select succeeded", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)

		e := &Grade{}
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

func getRandomGrades() []*domain.Grade {
	now := time.Now()
	g1 := &domain.Grade{
		ID:                idutil.ULIDNow(),
		PartnerInternalID: "partner_id" + idutil.ULIDNow(),
		Name:              "some name" + idutil.ULIDNow(),
		CreatedAt:         now,
		UpdatedAt:         now,
		IsArchived:        randBool(),
		DeletedAt:         nil,
	}
	g2 := &domain.Grade{
		ID:                idutil.ULIDNow(),
		PartnerInternalID: "partner_id" + idutil.ULIDNow(),
		Name:              "some name" + idutil.ULIDNow(),
		CreatedAt:         now,
		UpdatedAt:         now,
		IsArchived:        randBool(),
		DeletedAt:         nil,
	}
	grades := []*domain.Grade{
		g1, g2,
	}
	return grades
}

func randBool() bool {
	rand.Seed(time.Now().UnixNano())
	return (rand.Intn(2) == 1)
}
