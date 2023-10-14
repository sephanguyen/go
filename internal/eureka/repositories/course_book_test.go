package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func CourseBookRepoWithSqlMock() (*CourseBookRepo, *testutil.MockDB) {
	r := &CourseBookRepo{}
	return r, testutil.NewMockDB()
}

func TestCourseBookRepo_FindByCourseIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseBookRepoWithSqlMock()
	ids := []string{"id", "id-1"}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &ids)

		books, err := r.FindByCourseIDs(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, books)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &ids)

		e := &entities.CoursesBooks{}
		fields, values := e.FieldMap()
		_ = e.CourseID.Set("id")
		_ = e.BookID.Set(ksuid.New().String())

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		books, err := r.FindByCourseIDs(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.Equal(t, map[string][]string{
			e.CourseID.String: {e.BookID.String},
		}, books)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}

func TestCourseBookRepo_FindByBookID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseBookRepoWithSqlMock()
	id := "id"
	pgID := database.Text(id)

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &pgID)

		courseIDs, err := r.FindByBookID(ctx, mockDB.DB, id)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, courseIDs)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &pgID)

		e := &entities.CoursesBooks{}
		fields, values := e.FieldMap()
		_ = e.CourseID.Set("id")
		_ = e.BookID.Set(ksuid.New().String())

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		courseIDs, err := r.FindByBookID(ctx, mockDB.DB, id)
		assert.Nil(t, err)
		assert.Equal(t, []string{e.CourseID.String}, courseIDs)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}

func TestCourseBookRepo_FindByBookIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseBookRepoWithSqlMock()
	ids := []string{"id-1", "id-2"}
	pgID := database.TextArray(ids)

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &pgID)

		courseIDs, err := r.FindByBookIDs(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, courseIDs)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &pgID)

		e := &entities.CoursesBooks{}
		fields, values := e.FieldMap()
		_ = e.CourseID.Set("id")
		_ = e.BookID.Set(ksuid.New().String())

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		courseIDs, err := r.FindByBookIDs(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.Equal(t, []*entities.CoursesBooks{e}, courseIDs)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}

func TestCourseBookRepo_FindByCourseIDAndBookID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseBookRepoWithSqlMock()
	bookID := "book-id"
	courseID := "course-id"
	pgBookID := database.Text(bookID)
	pgCourseID := database.Text(courseID)

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &pgBookID, &pgCourseID)

		courseIDs, err := r.FindByCourseIDAndBookID(ctx, mockDB.DB, pgBookID, pgCourseID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, courseIDs)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &pgBookID, &pgCourseID)

		e := &entities.CoursesBooks{}
		fields, values := e.FieldMap()
		_ = e.CourseID.Set("course-id")
		_ = e.BookID.Set("book-id")
		_ = e.BookID.Set(ksuid.New().String())

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.FindByCourseIDAndBookID(ctx, mockDB.DB, pgBookID, pgCourseID)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}

func TestCourseBookRepo_SoftDelete(t *testing.T) {
	t.Parallel()
	r, mockDB := CourseBookRepoWithSqlMock()
	courseIDs := database.TextArray([]string{"course-id-1", "course-id-2"})
	bookIDs := database.TextArray([]string{"book-id-1", "book-id-2"})

	testCases := []TestCase{
		{
			name:        "error cannot delete course book",
			expectedErr: errors.New("cannot delete course book"),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), &bookIDs, &courseIDs).Once().Return(nil, nil)
			},
		},
		{
			name:        "happy case",
			expectedErr: errors.New("cannot delete course book"),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), &bookIDs, &courseIDs).Once().Return(nil, nil)
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("Exec").Once().Return(cmdTag, nil)
				mockDB.DB.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := r.SoftDelete(ctx, mockDB.DB, courseIDs, bookIDs)
		if err != nil {
			assert.Equal(t, err.Error(), testCase.expectedErr.Error())
		} else {
			assert.Nil(t, err)
		}
	}
}
