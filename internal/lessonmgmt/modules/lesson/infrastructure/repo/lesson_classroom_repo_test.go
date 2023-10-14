package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func LessonClassroomRepoWithSqlMock() (*LessonClassroomRepo, *testutil.MockDB) {
	mockDB := testutil.NewMockDB()
	mockRepo := &LessonClassroomRepo{}
	return mockRepo, mockDB
}

func TestLessonClassroomRepo_GetClassroomIDsByLessonIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonClassroomRepo, mockDB := LessonClassroomRepoWithSqlMock()
	lessonIDs := []string{"lesson-id1", "lesson-id2"}
	lc := &LessonClassroom{}
	fields, values := lc.FieldMap()

	t.Run("successful get classroom ids", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, lessonIDs)
		mockDB.MockScanFields(nil, fields, values)
		lessonClassrooms, err := lessonClassroomRepo.GetClassroomIDsByLessonIDs(ctx, mockDB.DB, lessonIDs)
		assert.NoError(t, err)
		assert.NotNil(t, lessonClassrooms)
	})

	t.Run("failed get classroom ids", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, lessonIDs)
		mockDB.MockScanFields(pgx.ErrNoRows, fields, values)
		lessonClassrooms, err := lessonClassroomRepo.GetClassroomIDsByLessonIDs(ctx, mockDB.DB, lessonIDs)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, lessonClassrooms)
	})
}

func TestLessonClassroomRepo_GetClassroomIDsByLessonID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonClassroomRepo, mockDB := LessonClassroomRepoWithSqlMock()
	lessonID := "lesson-id1"
	lc := &LessonClassroom{}
	fields, values := lc.FieldMap()

	t.Run("successful get classroom ids", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, lessonID)
		mockDB.MockScanFields(nil, fields, values)
		lessonClassrooms, err := lessonClassroomRepo.GetClassroomIDsByLessonID(ctx, mockDB.DB, lessonID)
		assert.NoError(t, err)
		assert.NotNil(t, lessonClassrooms)
	})

	t.Run("failed get classroom ids", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, lessonID)
		mockDB.MockScanFields(pgx.ErrNoRows, fields, values)
		lessonClassrooms, err := lessonClassroomRepo.GetClassroomIDsByLessonID(ctx, mockDB.DB, lessonID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, lessonClassrooms)
	})
}

func TestLessonClassroomRepo_GetLessonClassroomsWithNamesByLessonIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonClassroomRepo, mockDB := LessonClassroomRepoWithSqlMock()
	lessonIDs := []string{"test-lesson-id-1", "test-lesson-id-2", "test-lesson-id-3"}
	lessonClassroom := &LessonClassroom{}
	var name pgtype.Text
	var roomArea pgtype.Text
	fields := []string{
		"lesson_id",
		"classroom_id",
		"name",
		"room_area",
	}
	values := append(database.GetScanFields(lessonClassroom, fields), &name, &roomArea)

	t.Run("failed to get lesson classrooms", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), lessonIDs)

		lessonClassrooms, err := lessonClassroomRepo.GetLessonClassroomsWithNamesByLessonIDs(ctx, mockDB.DB, lessonIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessonClassrooms)
	})

	t.Run("successfully fetched lesson classrooms", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), lessonIDs)

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		lessonClassrooms, err := lessonClassroomRepo.GetLessonClassroomsWithNamesByLessonIDs(ctx, mockDB.DB, lessonIDs)
		assert.Nil(t, err)
		assert.NotNil(t, lessonClassrooms)
	})
}
