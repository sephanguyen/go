package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func CourseTypeRepoWithSqlMock() (*CourseTypeRepo, *testutil.MockDB) {
	r := &CourseTypeRepo{}
	return r, testutil.NewMockDB()
}

func TestCourseTypeRepo_GetByCourseIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseTypeRepoWithSqlMock()
	ids := []string{"id", "id-1"}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &ids)

		courseTypes, err := r.GetByIDs(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, courseTypes)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &ids)

		e := &CourseType{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.GetByIDs(ctx, mockDB.DB, ids)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"course_type_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestCourseTypeRepo_Import(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := CourseTypeRepoWithSqlMock()
	t.Run("successfully", func(t *testing.T) {
		ct := getRandomCourseTypes()
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
		ct := getRandomCourseTypes()

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

func getRandomCourseTypes() []*domain.CourseType {
	now := time.Now()
	c1 := &domain.CourseType{
		CourseTypeID: idutil.ULIDNow(),
		Name:         "type" + idutil.ULIDNow(),
		CreatedAt:    now,
		UpdatedAt:    now,
		IsArchived:   randBool(),
		Remarks:      "Some remarks 1",
		DeletedAt:    nil,
	}
	c2 := &domain.CourseType{
		CourseTypeID: idutil.ULIDNow(),
		Name:         "type" + idutil.ULIDNow(),
		CreatedAt:    now,
		UpdatedAt:    now,
		IsArchived:   randBool(),
		Remarks:      "Some remarks 2",
		DeletedAt:    nil,
	}
	ct := []*domain.CourseType{c1, c2}

	return ct
}
