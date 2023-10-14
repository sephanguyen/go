package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/mock/testutil"

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

func TestLessonTeacherRepo_GetTeachersByLessonID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := LessonTeacherRepoWithSqlMock()
	e := &LessonTeacher{}
	fields, value := e.FieldMap()
	lessonIDs := []string{"lesson-id", "lesson-id-2"}
	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, lessonIDs)
		gotLessonMembers, err := r.GetTeachersByLessonIDs(ctx, mockDB.DB, lessonIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, gotLessonMembers)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, lessonIDs)
		mockDB.MockScanFields(nil, fields, value)
		gotLessonMembers, err := r.GetTeachersByLessonIDs(ctx, mockDB.DB, lessonIDs)
		assert.NoError(t, err)
		assert.NotNil(t, gotLessonMembers)
	})

}

func TestLessonTeacherRepo_GetTeacherIDsOnlyByLessonIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	dto := &LessonTeacher{}
	fields, value := dto.FieldMap()
	lessonIDs := []string{"lesson-id1", "lesson-id2", "lesson-id3"}

	t.Run("error", func(t *testing.T) {
		mockRepo, mockDB := LessonTeacherRepoWithSqlMock()
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, lessonIDs)

		lessonTeachersMap, err := mockRepo.GetTeacherIDsOnlyByLessonIDs(ctx, mockDB.DB, lessonIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessonTeachersMap)
	})
	t.Run("success", func(t *testing.T) {
		mockRepo, mockDB := LessonTeacherRepoWithSqlMock()
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, lessonIDs)
		mockDB.MockScanFields(nil, fields, value)

		lessonTeachersMap, err := mockRepo.GetTeacherIDsOnlyByLessonIDs(ctx, mockDB.DB, lessonIDs)
		assert.NoError(t, err)
		assert.NotNil(t, lessonTeachersMap)
	})

}
