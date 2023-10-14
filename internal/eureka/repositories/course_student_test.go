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

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func CourseStudentRepoWithSqlMock() (*CourseStudentRepo, *testutil.MockDB) {
	r := &CourseStudentRepo{}
	return r, testutil.NewMockDB()
}

func TestCourseStudentRepo_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseStudentRepoWithSqlMock()

	t.Run("should upsert success", func(t *testing.T) {
		// Arrange
		e := &entities.CourseStudent{}
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
		e := &entities.CourseStudent{}
		fields, values := e.FieldMap()
		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		// Action
		err := r.Upsert(ctx, mockDB.DB, e)

		// Assert
		assert.Equal(t, fmt.Errorf("cannot insert course student"), err)
		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})

	t.Run("should throw error pgsql", func(t *testing.T) {
		// Arrange
		e := &entities.CourseStudent{}
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

func TestCourseStudentRepo_SoftDelete(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseStudentRepoWithSqlMock()

	t.Run("should success", func(t *testing.T) {
		// Arrange
		e := &entities.CourseStudent{}
		_, values := e.FieldMap()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values[2].(*pgtype.Text).String, values[1].(*pgtype.Text).String)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		// Action
		err := r.SoftDelete(ctx, mockDB.DB, []string{e.StudentID.String}, []string{e.CourseID.String})

		// Assert
		assert.Nil(t, err)
		mockDB.RawStmt.AssertUpdatedTable(t, e.TableName())
		updatedFields := []string{"deleted_at"}
		mockDB.RawStmt.AssertUpdatedFields(t, updatedFields...)
	})

	t.Run("should throw error pgsql", func(t *testing.T) {
		// Arrange
		e := &entities.CourseStudent{}
		_, values := e.FieldMap()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values[2].(*pgtype.Text).String, values[1].(*pgtype.Text).String)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		// Action
		err := r.SoftDelete(ctx, mockDB.DB, []string{e.StudentID.String}, []string{e.CourseID.String})

		// Assert
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		mockDB.RawStmt.AssertUpdatedTable(t, e.TableName())
		updatedFields := []string{"deleted_at"}
		mockDB.RawStmt.AssertUpdatedFields(t, updatedFields...)
	})
}

func TestCourseStudentRepo_BulkUpsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	courseStudentRepo := &CourseStudentRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.CourseStudent{
				{
					ID:        pgtype.Text{String: "1", Status: pgtype.Present},
					CourseID:  pgtype.Text{String: "course-id", Status: pgtype.Present},
					StudentID: pgtype.Text{String: "student-id", Status: pgtype.Present},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				row := &mock_database.Row{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("QueryRow").Once().Return(row, nil)
				row.On("Scan", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "error send batch",
			req: []*entities.CourseStudent{
				{
					ID: pgtype.Text{String: "1", Status: pgtype.Present},
				},
				{
					ID: pgtype.Text{String: "2", Status: pgtype.Present},
				},
				{
					ID: pgtype.Text{String: "3", Status: pgtype.Present},
				},
			},
			expectedErr: fmt.Errorf("batchResults.QueryRow: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				row := &mock_database.Row{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("QueryRow").Once().Return(row)
				row.On("Scan", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				batchResults.On("QueryRow").Once().Return(row)
				row.On("Scan", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				batchResults.On("QueryRow").Once().Return(row)
				row.On("Scan", mock.Anything, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxClosed)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := courseStudentRepo.BulkUpsert(ctx, db, testCase.req.([]*entities.CourseStudent))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}

	return
}

func TestCourseStudentRepo_SearchStudents(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseStudentRepoWithSqlMock()
	rows := &mock_database.Rows{}
	courseIDs := database.TextArray([]string{"id-1"})
	// var studentIDs pgtype.TextArray

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			puddle.ErrClosedPool,
			mock.Anything,
			mock.AnythingOfType("string"),
			courseIDs,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)

		studentIds, _, err := r.SearchStudents(ctx, mockDB.DB, &SearchStudentsFilter{
			CourseIDs: courseIDs,
		})

		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, studentIds)
	})

	t.Run("success", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", mock.Anything, mock.Anything).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)
		rows.On("Close").Once().Return()

		studentIds, _, err := r.SearchStudents(ctx, mockDB.DB, &SearchStudentsFilter{
			CourseIDs: courseIDs,
		})

		assert.Nil(t, err)
		assert.NotNil(t, studentIds)
	})
}

func TestCourseStudentRepo_FindStudentByCourseID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseStudentRepoWithSqlMock()
	courseID := database.Text("id")

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			puddle.ErrClosedPool,
			mock.Anything,
			mock.AnythingOfType("string"),
			&courseID,
		)

		studentIds, err := r.FindStudentByCourseID(ctx, mockDB.DB, courseID)

		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, studentIds)
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			nil,
			mock.Anything,
			mock.AnythingOfType("string"),
			&courseID,
		)

		e := &entities.CourseStudent{}
		fields, _ := e.FieldMap()
		values := []interface{}{&e.StudentID.String}
		mockDB.MockScanArray(nil, []string{fields[3]}, [][]interface{}{values})

		studentIds, err := r.FindStudentByCourseID(ctx, mockDB.DB, courseID)

		assert.Nil(t, err)
		assert.NotNil(t, studentIds)
	})
}

func TestCourseStudentRepo_GetByCourseStudents(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	now := time.Now()
	r, mockDB := CourseStudentRepoWithSqlMock()
	expected := entities.CourseStudents{
		{
			BaseEntity: entities.BaseEntity{
				CreatedAt: database.Timestamptz(now),
				UpdatedAt: database.Timestamptz(now),
				DeletedAt: pgtype.Timestamptz{
					Status: pgtype.Null,
				},
			},
			ID:        database.Text("id-1"),
			CourseID:  database.Text("course-id-1"),
			StudentID: database.Text("student-id-1"),
			StartAt:   database.Timestamptz(now),
			EndAt:     database.Timestamptz(now),
		},
		{
			BaseEntity: entities.BaseEntity{
				CreatedAt: database.Timestamptz(now),
				UpdatedAt: database.Timestamptz(now),
				DeletedAt: pgtype.Timestamptz{
					Status: pgtype.Null,
				},
			},
			ID:        database.Text("id-2"),
			CourseID:  database.Text("course-id-2"),
			StudentID: database.Text("student-id-1"),
			StartAt:   database.Timestamptz(now),
			EndAt:     database.Timestamptz(now),
		},
		{
			BaseEntity: entities.BaseEntity{
				CreatedAt: database.Timestamptz(now),
				UpdatedAt: database.Timestamptz(now),
				DeletedAt: pgtype.Timestamptz{
					Status: pgtype.Null,
				},
			},
			ID:        database.Text("id-3"),
			CourseID:  database.Text("course-id-2"),
			StudentID: database.Text("student-id-2"),
			StartAt:   database.Timestamptz(now),
			EndAt:     database.Timestamptz(now),
		},
	}

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
			&expected[0].StudentID, &expected[0].CourseID,
			&expected[1].StudentID, &expected[1].CourseID,
			&expected[2].StudentID, &expected[2].CourseID,
		)

		e := entities.CourseStudent{}
		fields := database.GetFieldNames(&e)
		values := make([][]interface{}, 0, len(expected))
		for i := range expected {
			values = append(values, database.GetScanFields(expected[i], fields))
		}
		mockDB.MockScanArray(nil, fields, values)

		actual, err := r.GetByCourseStudents(ctx, mockDB.DB, entities.CourseStudents{
			{
				CourseID:  database.Text("course-id-1"),
				StudentID: database.Text("student-id-1"),
			},
			{
				CourseID:  database.Text("course-id-2"),
				StudentID: database.Text("student-id-1"),
			},
			{
				CourseID:  database.Text("course-id-2"),
				StudentID: database.Text("student-id-2"),
			},
		})
		assert.Nil(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything,
			&expected[0].StudentID, &expected[0].CourseID,
			&expected[1].StudentID, &expected[1].CourseID,
			&expected[2].StudentID, &expected[2].CourseID,
		)

		actual, err := r.GetByCourseStudents(ctx, mockDB.DB, entities.CourseStudents{
			{
				CourseID:  database.Text("course-id-1"),
				StudentID: database.Text("student-id-1"),
			},
			{
				CourseID:  database.Text("course-id-2"),
				StudentID: database.Text("student-id-1"),
			},
			{
				CourseID:  database.Text("course-id-2"),
				StudentID: database.Text("student-id-2"),
			},
		})
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, actual)
	})
}

func TestCourseStudentRepo_FindStudentTagByCourseID(t *testing.T) {
	mockDB := testutil.NewMockDB()
	ctx := context.Background()
	repo := &CourseStudentRepo{}

	e := entities.StudentTag{}
	fields, values := e.FieldMap()
	courseID := database.Text("course_id")

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, courseID)
				mockDB.MockScanFields(nil, fields, values)
			},
			req:         courseID,
			expectedErr: nil,
		},
		{
			name: "unexpected error",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, courseID)
				mockDB.MockScanFields(pgx.ErrNoRows, fields, values)
			},
			req:          courseID,
			expectedResp: nil,
			expectedErr:  fmt.Errorf("database.Select: %w", fmt.Errorf("rows.Scan: %w", pgx.ErrNoRows)),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := repo.FindStudentTagByCourseID(ctx, mockDB.DB, testCase.req.(pgtype.Text))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}
