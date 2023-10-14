package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func CourseAccessPathRepoWithSqlMock() (*CourseAccessPathRepo, *testutil.MockDB) {
	r := &CourseAccessPathRepo{}
	return r, testutil.NewMockDB()
}

func TestCourseAccessPathRepo_FindByCourseIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseAccessPathRepoWithSqlMock()
	ids := []string{"id", "id-1"}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &ids)

		books, err := r.FindByCourseIDs(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, books)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &ids)

		e := &entities_bob.CourseAccessPath{}
		fields, values := e.FieldMap()
		_ = e.CourseID.Set("id")

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		courseLocations, err := r.FindByCourseIDs(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.Equal(t, map[string][]string{
			e.CourseID.String: {e.LocationID.String},
		}, courseLocations)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"course_id":  {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}
func TestCourseAccessPathRepo_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := CourseAccessPathRepoWithSqlMock()
	t.Run("successfully", func(t *testing.T) {
		course1 := &entities_bob.CourseAccessPath{
			LocationID: database.Text("location-1"),
			CourseID:   database.Text("course-1"),
			CreatedAt:  database.Timestamptz(time.Now()),
			UpdatedAt:  database.Timestamptz(time.Now()),
		}
		course2 := &entities_bob.CourseAccessPath{
			LocationID: database.Text("location-2"),
			CourseID:   database.Text("course-2"),
			CreatedAt:  database.Timestamptz(time.Now()),
			UpdatedAt:  database.Timestamptz(time.Now()),
		}
		caps := []*entities_bob.CourseAccessPath{course1, course2}
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)

		err := r.Upsert(ctx, mockDB.DB, caps)
		require.Equal(t, err, nil)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		course1 := &entities_bob.CourseAccessPath{
			LocationID: database.Text("location-1"),
			CourseID:   database.Text("course-1"),
			CreatedAt:  database.Timestamptz(time.Now()),
			UpdatedAt:  database.Timestamptz(time.Now()),
		}
		course2 := &entities_bob.CourseAccessPath{
			LocationID: database.Text("location-2"),
			CourseID:   database.Text("course-2"),
			CreatedAt:  database.Timestamptz(time.Now()),
			UpdatedAt:  database.Timestamptz(time.Now()),
		}
		caps := []*entities_bob.CourseAccessPath{course1, course2}
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)

		r.Upsert(ctx, mockDB.DB, caps)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}

func TestCourseAccessPathRepo_Delete(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseAccessPathRepoWithSqlMock()

	courseIDs := database.TextArray([]string{"courseID-1", "courseID-2"})

	t.Run("err update", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &courseIDs)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := r.Delete(ctx, mockDB.DB, courseIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &courseIDs)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.Delete(ctx, mockDB.DB, courseIDs)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertUpdatedTable(t, "course_access_paths")
		// move primaryField to the last
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at", "updated_at")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"course_id":  {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"deleted_at": {HasNullTest: true},
		})
	})
}
