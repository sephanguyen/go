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

func CourseClassRepoWithSqlMock() (*CourseClassRepo, *testutil.MockDB) {
	r := &CourseClassRepo{}
	return r, testutil.NewMockDB()
}

func TestCourseClassRepo_SoftDelete(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseClassRepoWithSqlMock()

	t.Run("should success", func(t *testing.T) {
		// Arrange
		e := &entities.CourseClass{}
		fields, _ := e.FieldMap()

		mockInput := []*entities.CourseClass{
			{CourseID: database.Text("course-1"), ClassID: database.Text("class-1")},
			{CourseID: database.Text("course-2"), ClassID: database.Text("class-2")},
			{CourseID: database.Text("course-3"), ClassID: database.Text("class-3")},
		}

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string"), "course-1", "class-1", "course-2", "class-2", "course-3", "class-3"})
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		// Action
		err := r.Delete(ctx, mockDB.DB, mockInput)

		// Assert
		assert.Nil(t, err)
		mockDB.RawStmt.AssertUpdatedTable(t, e.TableName())
		mockDB.RawStmt.AssertUpdatedFields(t, fields[5:]...)
	})
}

func TestCourseClassRepo_BulkUpsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	courseClassRepo := &CourseClassRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.CourseClass{
				{
					ID:       pgtype.Text{String: "1", Status: pgtype.Present},
					CourseID: pgtype.Text{String: "course-id", Status: pgtype.Present},
					ClassID:  pgtype.Text{String: "class-id", Status: pgtype.Present},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "error send batch",
			req: []*entities.CourseClass{
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
			expectedErr: fmt.Errorf("batchResults.Exec: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, pgx.ErrTxClosed)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := courseClassRepo.BulkUpsert(ctx, db, testCase.req.([]*entities.CourseClass))
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestCourseClassRepo_FindClassIDByCourseID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseClassRepoWithSqlMock()
	courseID := "course-id"
	pgCourseID := database.Text(courseID)

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &pgCourseID)

		courseIDs, err := r.FindClassIDByCourseID(ctx, mockDB.DB, pgCourseID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, courseIDs)
	})

	t.Run("success with select", func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &pgCourseID)
		mockDB.DB.On("Query").Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", mock.Anything).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)

		_, err := r.FindClassIDByCourseID(ctx, mockDB.DB, pgCourseID)
		assert.Nil(t, err)
	})
}

func TestCourseClassRepo_DeleteClass(t *testing.T) {
	t.Parallel()

	e := &entities.CourseClass{}
	r, mockDB := CourseClassRepoWithSqlMock()

	classID := "class__ID"

	testCases := []TestCase{
		{
			name: "Happy case delete single class",
			setup: func(ctx context.Context) {
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string"), classID})
				mockDB.MockExecArgs(t, pgconn.CommandTag([]byte(`1`)), nil, args...)
			},
			req: classID,
		},
		{
			name: "Exec query err",
			setup: func(ctx context.Context) {
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string"), classID})
				mockDB.MockExecArgs(t, nil, puddle.ErrClosedPool, args...)
			},
			req:         classID,
			expectedErr: fmt.Errorf("db.Exec: %w", puddle.ErrClosedPool),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			input := testCase.req.(string)

			err := r.DeleteClass(ctx, mockDB.DB, input)

			mockDB.RawStmt.AssertUpdatedTable(t, e.TableName())
			mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at")
			mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
				"class_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			})

			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, puddle.ErrClosedPool)
				return
			}

			assert.NoError(t, err)
		})

	}

}
