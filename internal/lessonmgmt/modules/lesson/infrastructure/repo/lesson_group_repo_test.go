package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	media_infrastructure "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/infrastructure"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func LessonGroupRepoWithSqlMock() (*LessonGroupRepo, *testutil.MockDB) {
	r := &LessonGroupRepo{}
	return r, testutil.NewMockDB()
}

func TestLessonGroupRepo_Insert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonGroupRepoWithSqlMock()
	t.Run("success", func(t *testing.T) {
		e := &LessonGroup{}
		fields, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.Insert(ctx, mockDB.DB, e)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})

	t.Run("err insert", func(t *testing.T) {
		e := &LessonGroup{}
		_, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		err := r.Insert(ctx, mockDB.DB, e)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
}

func TestLessonGroupRepo_ListMediaByLessonArgs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonGroupRepoWithSqlMock()

	lessonID := "lesson-id-1"
	limit := uint32(2)
	offset := ""
	args := &domain.ListMediaByLessonArgs{
		LessonID: lessonID,
		Limit:    uint32(limit),
		Offset:   offset,
	}
	t.Run("err get", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			&args.LessonID,
		)
		scanGroupID := database.Text("lesson_group_id")
		scanCourseID := database.Text("course_id")
		mockDB.MockRowScanFields(nil, []string{"lesson_group_id", "course_id"}, []interface{}{&scanGroupID, &scanCourseID})
		mockDB.MockQueryArgs(t,
			pgx.ErrNoRows,
			mock.Anything,
			mock.AnythingOfType("string"),
			scanGroupID,
			scanCourseID,
			offset,
			limit,
		)

		_, err := r.ListMediaByLessonArgs(ctx, mockDB.DB, args)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			&args.LessonID,
		)
		scanneGroupID := database.Text("lesson_group_id")
		scanneCourseID := database.Text("course_id")
		mockDB.MockRowScanFields(nil, []string{"lesson_group_id", "course_id"}, []interface{}{&scanneGroupID, &scanneCourseID})
		mockDB.MockQueryArgs(t,
			nil,
			mock.Anything,
			mock.AnythingOfType("string"),
			scanneGroupID,
			scanneCourseID,
			offset,
			limit,
		)

		e := &media_infrastructure.Media{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		_, err := r.ListMediaByLessonArgs(ctx, mockDB.DB, args)
		assert.Nil(t, err)
	})
}
