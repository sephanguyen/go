package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func CourseLocationScheduleRepoWithSqlMock() (*CourseLocationScheduleRepo, *testutil.MockDB) {
	r := &CourseLocationScheduleRepo{}
	return r, testutil.NewMockDB()
}

func TestCourseLocationScheduleRepo_UpsertMultiCourseLocationSchedule(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	z, mockDB := CourseLocationScheduleRepoWithSqlMock()
	t.Run("error", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Return(cmdTag, errors.New("batchResults.Exec: closed pool"))
		batchResults.On("Close").Once().Return(nil)
		err := z.UpsertMultiCourseLocationSchedule(ctx, mockDB.DB, []*domain.CourseLocationSchedule{
			{
				ID: "123",
			},
		})

		assert.Error(t, err.Err)
	})
	t.Run("success", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)
		err := z.UpsertMultiCourseLocationSchedule(ctx, mockDB.DB, []*domain.CourseLocationSchedule{
			{
				ID: "123",
			},
		})
		assert.Nil(t, err)
	})
}

func TestCourseLocationScheduleRepo_ExportCourseLocationSchedule(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	z, mockDB := CourseLocationScheduleRepoWithSqlMock()
	courseLocationSchedule := &CourseLocationSchedule{}
	fields, values := courseLocationSchedule.FieldMap()
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockScanFields(nil, fields, values)
		data, err := z.ExportCourseLocationSchedule(ctx, mockDB.DB)
		assert.NoError(t, err)
		assert.NotNil(t, data)
	})
	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything)
		data, err := z.ExportCourseLocationSchedule(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, data)
	})
}
