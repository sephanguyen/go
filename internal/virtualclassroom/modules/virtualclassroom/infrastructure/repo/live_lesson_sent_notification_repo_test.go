package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func LiveLessonSentNotificationRepoWithSqlMock() (*LiveLessonSentNotificationRepo, *testutil.MockDB) {
	r := &LiveLessonSentNotificationRepo{}
	return r, testutil.NewMockDB()
}

func TestLiveLessonSentNotificationRepo_GetLiveLessonSentNotificationCount(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LiveLessonSentNotificationRepoWithSqlMock()
	mockTotal := database.Int8(1)
	totalFields := []string{"total"}
	totalValues := []interface{}{&mockTotal}

	t.Run("err row count", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		mockDB.MockRowScanFields(pgx.ErrNoRows, totalFields, totalValues)

		_, err := r.GetLiveLessonSentNotificationCount(ctx, mockDB.DB, mock.Anything, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
	})

	t.Run("success row count", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			mock.Anything,
		)

		mockDB.MockRowScanFields(nil, totalFields, totalValues)
		total, err := r.GetLiveLessonSentNotificationCount(ctx, mockDB.DB, mock.Anything, mock.Anything)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, int32(1))
	})
}

func TestLiveLessonSentNotificationRepo_CreateLiveLessonSentNotificationRecord(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LiveLessonSentNotificationRepoWithSqlMock()

	e := LiveLessonSentNotification{}
	_, values := e.FieldMap()
	values = append(values, "org-1")
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(values))...)

	t.Run("err create", func(t *testing.T) {
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed).Once()
		err := r.CreateLiveLessonSentNotificationRecord(ctx, mockDB.DB, mock.Anything, mock.Anything, time.Now())
		assert.NotNil(t, err)
	})

	t.Run("success create", func(t *testing.T) {
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := r.CreateLiveLessonSentNotificationRecord(ctx, mockDB.DB, mock.Anything, mock.Anything, time.Now())
		assert.Nil(t, err)
	})
}

func TestLiveLessonSentNotificationRepo_SoftDeleteLiveLessonSentNotificationRecord(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LiveLessonSentNotificationRepoWithSqlMock()

	e := LiveLessonSentNotification{}
	_, values := e.FieldMap()
	values = append(values, "org-1")
	lessonID := "lesson-1"
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, lessonID)

	t.Run("err update", func(t *testing.T) {
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed).Once()
		err := r.SoftDeleteLiveLessonSentNotificationRecord(ctx, mockDB.DB, lessonID)
		assert.NotNil(t, err)
	})

	t.Run("success update", func(t *testing.T) {
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := r.SoftDeleteLiveLessonSentNotificationRecord(ctx, mockDB.DB, lessonID)
		assert.Nil(t, err)
	})
}
