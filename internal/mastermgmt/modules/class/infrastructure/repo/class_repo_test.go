package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func ClassRepoWithSqlMock() (*ClassRepo, *testutil.MockDB) {
	classRepo := &ClassRepo{}
	return classRepo, testutil.NewMockDB()
}

func TestClassRepo_GetByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	classRepo, mockDB := ClassRepoWithSqlMock()
	e := &Class{}
	fields, value := e.FieldMap()
	classID := idutil.ULIDNow()
	t.Run("error no row", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, value)
		gotClass, err := classRepo.GetByID(ctx, mockDB.DB, classID)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
		assert.Nil(t, gotClass)
	})
	t.Run("error tx closed", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockRowScanFields(pgx.ErrTxClosed, fields, value)
		gotClass, err := classRepo.GetByID(ctx, mockDB.DB, classID)
		assert.ErrorIs(t, err, pgx.ErrTxClosed)
		assert.Nil(t, gotClass)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockRowScanFields(nil, fields, value)
		gotClass, err := classRepo.GetByID(ctx, mockDB.DB, classID)
		assert.NoError(t, err)
		assert.NotNil(t, gotClass)
	})
}

func TestClassRepo_Insert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()
	classRepo, mockDB := ClassRepoWithSqlMock()
	classes := []*domain.Class{
		{
			ClassID:    "class-1",
			CourseID:   "course-1",
			LocationID: "location-1",
			Name:       "class-name",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ClassID:    "class-2",
			CourseID:   "course-2",
			LocationID: "location-2",
			Name:       "class-name",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}
	t.Run("success", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)
		err := classRepo.Insert(ctx, mockDB.DB, classes)
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)
		err := classRepo.Insert(ctx, mockDB.DB, classes)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}

func TestClassRepo_UpsertClassses(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()
	classRepo, mockDB := ClassRepoWithSqlMock()
	classes := []*domain.Class{
		{
			ClassID:    "class-1",
			CourseID:   "course-1",
			LocationID: "location-1",
			Name:       "class-name",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ClassID:    "class-2",
			CourseID:   "course-2",
			LocationID: "location-2",
			Name:       "class-name",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}
	t.Run("success", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)
		err := classRepo.UpsertClasses(ctx, mockDB.DB, classes)
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)
		err := classRepo.UpsertClasses(ctx, mockDB.DB, classes)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}

func TestClassRepo_UpdateClassNameByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	classRepo, mockDB := ClassRepoWithSqlMock()
	name := "name"
	id := "class-id"
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, name, id)
	t.Run("error with no rows affected", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)
		err := classRepo.UpdateClassNameByID(ctx, mockDB.DB, id, name)
		assert.Equal(t, err, domain.ErrNotFound)
	})
	t.Run("error", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), errors.New("something went wrong"), args...)
		err := classRepo.UpdateClassNameByID(ctx, mockDB.DB, id, name)
		assert.Equal(t, err, errors.New("something went wrong"))
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
		err := classRepo.UpdateClassNameByID(ctx, mockDB.DB, id, name)
		assert.NoError(t, err)
	})
}

func TestClassRepo_Delete(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	classRepo, mockDB := ClassRepoWithSqlMock()
	id := "class-id"
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, id)
	t.Run("error with no rows affected", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)
		err := classRepo.Delete(ctx, mockDB.DB, id)
		assert.Equal(t, err, domain.ErrNotFound)
	})
	t.Run("error", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), errors.New("something went wrong"), args...)
		err := classRepo.Delete(ctx, mockDB.DB, id)
		assert.Equal(t, err, errors.New("something went wrong"))
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
		err := classRepo.Delete(ctx, mockDB.DB, id)
		assert.NoError(t, err)
	})
}

func TestLocationRepo_RetrieveClassesByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	classRepo, mockDB := ClassRepoWithSqlMock()
	e := &Class{}
	classIds := []string{"class-id-1", "class-id-2"}
	fields, values := e.FieldMap()
	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, mock.Anything)
		classes, err := classRepo.RetrieveByIDs(ctx, mockDB.DB, classIds)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, classes)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		classes, err := classRepo.RetrieveByIDs(ctx, mockDB.DB, classIds)
		assert.NoError(t, err)
		assert.NotNil(t, classes)
	})
}

func TestClassRepo_FindByCourseIDsAndStudentIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ClassRepoWithSqlMock()
	css := []*domain.ClassWithCourseStudent{
		{
			CourseID:  "course-1",
			StudentID: "student-1",
			ClassID:   "class-1",
		},
		{
			CourseID:  "course-2",
			StudentID: "student-2",
			ClassID:   "class-2",
		},
	}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), &css[0].CourseID, &css[0].StudentID, &css[1].CourseID, &css[1].StudentID)

		result, err := r.FindByCourseIDsAndStudentIDs(ctx, mockDB.DB, css)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, result)
	})

	t.Run("success", func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &css[0].CourseID, &css[0].StudentID, &css[1].CourseID, &css[1].StudentID).Once().Return(rows, nil)
		rows.On("Next").Times(2).Return(true)
		y := 0
		for i := 0; i < 2; i++ {
			e := &ClassWithCourseStudent{}
			_, values := e.FieldMap()
			rows.On("Scan", values...).Once().Run(func(args mock.Arguments) {
				args[0].(*pgtype.Text).String = css[y].ClassID
				args[1].(*pgtype.Text).String = css[y].CourseID
				args[2].(*pgtype.Text).String = css[y].StudentID
				y++
			}).Return(nil)
		}
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		rows.On("Err").Once().Return(nil)

		result, err := r.FindByCourseIDsAndStudentIDs(ctx, mockDB.DB, css)
		assert.Nil(t, err)
		for i := 0; i < 2; i++ {
			assert.Equal(t, css[i].ClassID, result[i].ClassID)
			assert.Equal(t, css[i].CourseID, result[i].CourseID)
			assert.Equal(t, css[i].StudentID, result[i].StudentID)
		}

	})
}

func TestClassRepo_GetAll(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	classRepo, mockDB := ClassRepoWithSqlMock()
	e := &domain.ExportingClass{}
	fields, values := e.FieldMap()

	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything)
		gotClasses, err := classRepo.GetAll(ctx, mockDB.DB)

		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, gotClasses)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)
		mockDB.MockScanFields(nil, fields, values)
		gotClasses, err := classRepo.GetAll(ctx, mockDB.DB)

		assert.NoError(t, err)
		assert.NotNil(t, gotClasses)
	})
}
