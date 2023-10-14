package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func LessonReportRepoWithSqlMock() (*LessonReportRepo, *testutil.MockDB) {
	r := &LessonReportRepo{}
	return r, testutil.NewMockDB()
}

func TestLessonReportRepo_FindLessonReportByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonReportRepoWithSqlMock()

	lessonReportID := database.Text("report_id_1")
	e := &entities.LessonReport{}
	selectFields, value := e.FieldMap()
	_ = e.LessonReportID.Set(lessonReportID)

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &lessonReportID)
		mockDB.MockRowScanFields(puddle.ErrClosedPool, selectFields, value)

		lesson_report, err := r.FindByID(ctx, mockDB.DB, lessonReportID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lesson_report)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &lessonReportID)
		mockDB.MockRowScanFields(nil, selectFields, value)

		lessonReport, err := r.FindByID(ctx, mockDB.DB, lessonReportID)

		assert.Nil(t, err)
		assert.Equal(t, &entities.LessonReport{LessonReportID: lessonReportID}, lessonReport)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}

func TestLessonReportRepo_DeleteReportsBelongToLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonReportRepoWithSqlMock()
	t.Run("delete successfully", func(t *testing.T) {
		const lessonID = "lesson-id-1"
		lessonIDPgt := database.Text(lessonID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, mock.Anything, mock.AnythingOfType("string"), &lessonIDPgt)

		err := r.DeleteReportsBelongToLesson(ctx, mockDB.DB, database.Text(lessonID))
		require.NoError(t, err)
	})

	t.Run("delete failed", func(t *testing.T) {
		const lessonID = "lesson-id-1"
		lessonIDPgt := database.Text(lessonID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), &lessonIDPgt)

		err := r.DeleteReportsBelongToLesson(ctx, mockDB.DB, database.Text(lessonID))
		require.Error(t, err)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
}
