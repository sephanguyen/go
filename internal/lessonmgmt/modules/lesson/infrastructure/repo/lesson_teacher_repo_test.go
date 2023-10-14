package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func LessonTeacherRepoWithSqlMock() (*LessonTeacherRepo, *testutil.MockDB) {
	r := &LessonTeacherRepo{}
	return r, testutil.NewMockDB()
}
func TestLessonTeacherRepo_GetTeacherByLessonIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := LessonTeacherRepoWithSqlMock()
	e := &LessonTeacher{}
	fields, value := e.FieldMap()
	lessonID := []string{"lesson-id"}
	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, lessonID)
		gotLessonMembers, err := r.GetTeacherIDsByLessonIDs(ctx, mockDB.DB, lessonID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, gotLessonMembers)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, lessonID)
		mockDB.MockScanFields(nil, fields, value)
		gotLessonMembers, err := r.GetTeacherIDsByLessonIDs(ctx, mockDB.DB, lessonID)
		assert.NoError(t, err)
		assert.NotNil(t, gotLessonMembers)
	})

}

func TestLessonTeacherRepo_GetTeacherByLessonID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := LessonTeacherRepoWithSqlMock()
	e := &LessonTeacher{}
	fields, value := e.FieldMap()
	lessonID := "lesson-id"
	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, lessonID)
		gotLessonMembers, err := r.GetTeacherIDsByLessonID(ctx, mockDB.DB, lessonID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, gotLessonMembers)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, lessonID)
		mockDB.MockScanFields(nil, fields, value)
		gotLessonMembers, err := r.GetTeacherIDsByLessonID(ctx, mockDB.DB, lessonID)
		assert.NoError(t, err)
		assert.NotNil(t, gotLessonMembers)
	})

}

func TestLessonTeacherRepo_UpdateLessonTeacherName(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	l, mockDB := LessonTeacherRepoWithSqlMock()
	lessonTeachers := []*domain.UpdateLessonTeacherName{{TeacherID: "teacher-1", FullName: "Full name"}}

	t.Run("err update", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Return(cmdTag, errors.New("batchResults.Exec: closed pool"))
		batchResults.On("Close").Once().Return(nil)
		err := l.UpdateLessonTeacherNames(ctx, mockDB.DB, lessonTeachers)
		assert.Equal(t, "batchResults.Exec: batchResults.Exec: closed pool", err.Error())
	})

	t.Run("success", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)

		err := l.UpdateLessonTeacherNames(ctx, mockDB.DB, lessonTeachers)
		assert.Equal(t, nil, err)
	})
}

func TestLessonTeacherRepo_GetTeachersWithNamesByLessonIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockLessonTeacherRepo, mockDB := LessonTeacherRepoWithSqlMock()
	lessonIDs := []string{"test-lesson-id-1", "test-lesson-id-2", "test-lesson-id-3"}
	LessonTeacher := &LessonTeacher{}
	var name pgtype.Text
	fields := []string{
		"lesson_id",
		"teacher_id",
		"name",
	}
	values := append(database.GetScanFields(LessonTeacher, fields), &name)

	t.Run("failed to get lesson teachers", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), lessonIDs)

		lessonTeachers, err := mockLessonTeacherRepo.GetTeachersWithNamesByLessonIDs(ctx, mockDB.DB, lessonIDs, false)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessonTeachers)
	})

	t.Run("successfully fetched lesson teachers", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), lessonIDs)

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		lessonTeachers, err := mockLessonTeacherRepo.GetTeachersWithNamesByLessonIDs(ctx, mockDB.DB, lessonIDs, false)
		assert.Nil(t, err)
		assert.NotNil(t, lessonTeachers)
	})

	t.Run("successfully fetched lesson teachers using user public info", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), lessonIDs)

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		lessonTeachers, err := mockLessonTeacherRepo.GetTeachersWithNamesByLessonIDs(ctx, mockDB.DB, lessonIDs, true)
		assert.Nil(t, err)
		assert.NotNil(t, lessonTeachers)
	})
}
