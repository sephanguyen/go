package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
	"github.com/segmentio/ksuid"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ClassStudentRepoWithSqlMock() (*ClassStudentRepo, *testutil.MockDB) {
	r := &ClassStudentRepo{}
	return r, testutil.NewMockDB()
}

func TestClassStudent_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ClassStudentRepoWithSqlMock()

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

func TestClassStudent_SoftDelete(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ClassStudentRepoWithSqlMock()

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

func TestClassStudent_SoftDeleteByCourseStudent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ClassStudentRepoWithSqlMock()

	t.Run("should success", func(t *testing.T) {
		// Arrange
		e := &entities.ClassStudent{}
		fields, _ := e.FieldMap()
		args := append([]interface{}{mock.Anything, mock.Anything, mock.Anything, mock.Anything})
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		// Action
		err := r.SoftDeleteByCourseStudent(ctx, mockDB.DB, &entities.CourseStudent{
			CourseID:  database.Text("courseID"),
			StudentID: database.Text("studentID"),
		})

		// Assert
		assert.Nil(t, err)
		mockDB.RawStmt.AssertUpdatedTable(t, e.TableName())
		mockDB.RawStmt.AssertUpdatedFields(t, fields[4:]...)
	})

	t.Run("should throw error pgsql", func(t *testing.T) {
		// Arrange
		e := &entities.ClassStudent{}
		fields, _ := e.FieldMap()
		args := append([]interface{}{mock.Anything, mock.Anything, mock.Anything, mock.Anything})
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		// Action
		err := r.SoftDeleteByCourseStudent(ctx, mockDB.DB, &entities.CourseStudent{
			CourseID:  database.Text("courseID"),
			StudentID: database.Text("studentID"),
		})

		// Assert
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		mockDB.RawStmt.AssertUpdatedTable(t, e.TableName())
		mockDB.RawStmt.AssertUpdatedFields(t, fields[4:]...)
	})
}

func TestClassStudent_BulkSoftDelete(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("should success", func(t *testing.T) {
		r, mockDB := ClassStudentRepoWithSqlMock()
		args := append([]interface{}{mock.Anything, mock.Anything, mock.Anything, mock.Anything})
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		batch := &mock_database.BatchResults{}
		batch.On("Len").Return(2)
		batch.On("Exec").Return(nil, nil)
		batch.On("Close").Return(nil)
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Return(batch)

		err := r.BulkSoftDelete(ctx, mockDB.DB, entities.ClassStudents{
			{
				ClassID:   database.Text("classID"),
				StudentID: database.Text("studentID"),
			},
		})

		// Assert
		assert.Nil(t, err)
	})

	t.Run("should throw error pgsql", func(t *testing.T) {

		r, mockDB := ClassStudentRepoWithSqlMock()
		batch := &mock_database.BatchResults{}
		batch.On("Queue", mock.Anything, mock.Anything)
		batch.On("Len").Return(1)
		batch.On("Exec").Return(nil, pgx.ErrNoRows)
		batch.On("Close").Return(nil)
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Return(batch)

		err := r.BulkSoftDelete(ctx, mockDB.DB, entities.ClassStudents{
			{
				ClassID:   database.Text("classID"),
				StudentID: database.Text("studentID"),
			},
		})

		// Assert
		assert.NotNil(t, err)
	})
}

func TestClassStudent_GetClassStudentByCourse(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ClassStudentRepoWithSqlMock()
	ids := []string{"id", "id-1"}
	courseIDs := database.TextArray(ids)
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, courseIDs)

		classStudents, err := r.GetClassStudentByCourse(ctx, mockDB.DB, courseIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, classStudents)
	})

	t.Run("success with get", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, courseIDs)

		e := &entities.ClassStudent{}
		fields, values := e.FieldMap()
		_ = e.StudentID.Set("id")
		_ = e.ClassID.Set(ksuid.New().String())

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		classStudents, err := r.GetClassStudentByCourse(ctx, mockDB.DB, courseIDs)
		assert.Nil(t, err)
		assert.Equal(t, []*entities.ClassStudent{e}, classStudents)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})
}

func TestClassStudent_GetClassStudentByCourseAndClassIds(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ClassStudentRepoWithSqlMock()
	ids := []string{"id", "id-1"}
	courseIDs := database.TextArray(ids)
	convertIDs := database.TextArray(ids)
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, courseIDs, convertIDs)

		classStudents, err := r.GetClassStudentByCourseAndClassIds(ctx, mockDB.DB, courseIDs, convertIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, classStudents)
	})

	t.Run("success with get", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, courseIDs, convertIDs)

		e := &entities.ClassStudent{}
		fields, values := e.FieldMap()
		_ = e.StudentID.Set("id")
		_ = e.ClassID.Set(ksuid.New().String())

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		classStudents, err := r.GetClassStudentByCourseAndClassIds(ctx, mockDB.DB, courseIDs, convertIDs)
		assert.Nil(t, err)
		assert.Equal(t, []*entities.ClassStudent{e}, classStudents)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})
}
