package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	"github.com/manabie-com/backend/mock/testutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func LessonRepoWithSqlMock() (LessonRepoImpl, *testutil.MockDB) {
	mockDB := testutil.NewMockDB()
	repo := LessonRepoImpl{}

	return repo, mockDB
}
func TestLessonRepoImpl_Retrieve(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonOne := &entity.Lesson{
		LessonID:         database.Text("12"),
		SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED.String()),
	}

	lessonTwo := &entity.Lesson{
		LessonID:         database.Text("13"),
		SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED.String()),
	}

	selectFields := []string{"lesson_id", "scheduling_status"}
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(selectFields))...)
	repo, mockDB := LessonRepoWithSqlMock()

	t.Run("failed to select lesson with scheduled status record", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		retrieveLessons, err := repo.FindLessonsByIDs(ctx, mockDB.DB, database.TextArray([]string{lessonOne.LessonID.String}))
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, retrieveLessons)
		mock.AssertExpectationsForObjects(t, mockDB.DB)

	})
	t.Run("No rows affected", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		retrieveLessons, err := repo.FindLessonsByIDs(ctx, mockDB.DB, database.TextArray([]string{"not-exist"}))
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, retrieveLessons)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("success retrieving single lesson with scheduled status record", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything)

		value := database.GetScanFields(lessonOne, selectFields)

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			value,
		})

		lessonRecords, err := repo.FindLessonsByIDs(ctx, mockDB.DB, database.TextArray([]string{lessonOne.LessonID.String}))
		assert.Nil(t, err)
		assert.Equal(t, []*entity.Lesson{
			{
				LessonID:         database.Text("12"),
				SchedulingStatus: database.Text(cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED.String()),
			},
		}, lessonRecords)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
	})

	t.Run("success retrieving multiple lesson with scheduled status records", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything)
		var lessonRecords []*entity.Lesson

		lessonRecords = append(lessonRecords, lessonOne)
		valueOne := database.GetScanFields(lessonOne, selectFields)

		lessonRecords = append(lessonRecords, lessonTwo)

		value := database.GetScanFields(lessonTwo, selectFields)

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			valueOne,
			value,
		})

		retrieveLessons, err := repo.FindLessonsByIDs(ctx, mockDB.DB, database.TextArray([]string{lessonOne.LessonID.String, lessonTwo.LessonID.String}))
		assert.Nil(t, err)
		assert.Equal(t, retrieveLessons, lessonRecords)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
	})
}
