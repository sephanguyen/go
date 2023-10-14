package repo

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func CourseRepoWithSqlMock() (*CourseRepo, *testutil.MockDB) {
	courseRepo := &CourseRepo{}
	return courseRepo, testutil.NewMockDB()
}

func TestCourseRepo_GetByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	courseRepo, mockDB := CourseRepoWithSqlMock()
	e := &Course{}
	fields, value := e.FieldMap()
	t.Run("error no row", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, value)
		gotCourse, err := courseRepo.GetByID(ctx, mockDB.DB, "id")
		assert.ErrorIs(t, err, pgx.ErrNoRows)
		assert.Nil(t, gotCourse)
	})
	t.Run("error tx closed", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockRowScanFields(pgx.ErrTxClosed, fields, value)
		gotCourse, err := courseRepo.GetByID(ctx, mockDB.DB, "id")
		assert.ErrorIs(t, err, pgx.ErrTxClosed)
		assert.Nil(t, gotCourse)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockRowScanFields(nil, fields, value)
		gotCourse, err := courseRepo.GetByID(ctx, mockDB.DB, "id")
		assert.NoError(t, err)
		assert.NotNil(t, gotCourse)
	})
}

func TestCourseRepo_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := CourseRepoWithSqlMock()
	t.Run("successfully", func(t *testing.T) {
		course1 := &domain.Course{
			Name:         "name-1",
			CourseID:     "course-1",
			CourseTypeID: "course-type-id-1",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		course2 := &domain.Course{
			Name:         "name-2",
			CourseID:     "course-2",
			CourseTypeID: "course-type-id-2",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		caps := []*domain.Course{course1, course2}
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
		course1 := &domain.Course{
			Name:         "name-1",
			CourseID:     "course-1",
			CourseTypeID: "course-type-id-1",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		course2 := &domain.Course{
			Name:         "name-2",
			CourseID:     "course-2",
			CourseTypeID: "course-type-id-2",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		caps := []*domain.Course{course1, course2}
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

func TestCourseRepo_LinkSubjects(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := CourseRepoWithSqlMock()
	t.Run("successfully", func(t *testing.T) {
		course1 := &domain.Course{
			Name:         "name-1",
			CourseID:     "course-1",
			CourseTypeID: "course-type-id-1",
			SubjectIDs:   []string{"subject_1", "subject_2"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		course2 := &domain.Course{
			Name:         "name-2",
			CourseID:     "course-2",
			CourseTypeID: "course-type-id-2",
			SubjectIDs:   []string{"subject_3"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		courses := []*domain.Course{course1, course2}
		batchResults := &mock_database.BatchResults{}
		cmdTag1 := pgconn.CommandTag([]byte(`UPDATE 2`))
		cmdTag2 := pgconn.CommandTag([]byte(`UPDATE 1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag1, nil)
		batchResults.On("Exec").Once().Return(cmdTag2, nil)
		batchResults.On("Close").Once().Return(nil)

		err := r.LinkSubjects(ctx, mockDB.DB, courses)
		require.Equal(t, err, nil)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		course1 := &domain.Course{
			Name:         "name-1",
			CourseID:     "course-1",
			CourseTypeID: "course-type-id-1",
			SubjectIDs:   []string{"subject_1", "subject_2"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		course2 := &domain.Course{
			Name:         "name-2",
			CourseID:     "course-2",
			CourseTypeID: "course-type-id-2",
			SubjectIDs:   []string{"subject_3"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		courses := []*domain.Course{course1, course2}
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`UPDATE`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)

		r.LinkSubjects(ctx, mockDB.DB, courses)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}
func TestCourseRepo_GetByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseRepoWithSqlMock()
	keys := []string{"key", "key-1"}

	t.Run("select failed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, database.TextArray(keys))

		c, err := r.GetByIDs(ctx, mockDB.DB, keys)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, c)
	})

	t.Run("select succeeded", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.TextArray(keys))

		e := &Course{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.GetByIDs(ctx, mockDB.DB, keys)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"course_id":  {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestCourseRepo_GetAll(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseRepoWithSqlMock()

	t.Run("select failed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything)

		c, err := r.GetAll(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, c)
	})

	t.Run("select succeeded", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)

		e := &Course{}
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

func TestCourseRepo_Import(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := CourseRepoWithSqlMock()
	t.Run("successfully", func(t *testing.T) {
		courses := getRandomCourses()
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)

		err := r.Import(ctx, mockDB.DB, courses)
		require.Nil(t, err)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		courses := getRandomCourses()

		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)

		err := r.Import(ctx, mockDB.DB, courses)
		require.Error(t, err)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}
func TestCourseRepo_GetByPartnerIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseRepoWithSqlMock()
	keys := []string{"key", "key-1"}

	t.Run("select failed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, database.TextArray(keys))

		c, err := r.GetByPartnerIDs(ctx, mockDB.DB, keys)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, c)
	})

	t.Run("select succeeded", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.TextArray(keys))

		e := &Course{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.GetByPartnerIDs(ctx, mockDB.DB, keys)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at":        {HasNullTest: true},
			"course_partner_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func getRandomCourses() []*domain.Course {
	now := time.Now()
	c1 := &domain.Course{
		CourseID:   idutil.ULIDNow(),
		Name:       "some name" + idutil.ULIDNow(),
		CreatedAt:  now,
		UpdatedAt:  now,
		IsArchived: randBool(),
		Remarks:    "Some remarks",
		DeletedAt:  nil,
	}
	c2 := &domain.Course{
		CourseID:   idutil.ULIDNow(),
		Name:       "some name" + idutil.ULIDNow(),
		CreatedAt:  now,
		UpdatedAt:  now,
		IsArchived: randBool(),
		Remarks:    "Some remarks",
		DeletedAt:  nil,
	}
	courses := []*domain.Course{c1, c2}

	return courses
}

func randBool() bool {
	rand.Seed(time.Now().UnixNano())
	return (rand.Intn(2) == 1)
}
