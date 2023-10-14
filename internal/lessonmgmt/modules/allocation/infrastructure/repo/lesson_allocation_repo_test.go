package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func LessonAllocationRepoWithSqlMock() (*LessonAllocationRepo, *testutil.MockDB) {
	r := &LessonAllocationRepo{}
	return r, testutil.NewMockDB()
}

func TestLessonAllocationRepo_CountPurchasedSlotPerStudentSubscription(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	l, mockDB := LessonAllocationRepoWithSqlMock()
	freq := uint8(2)
	startTime := time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)
	courseID := "course-id"
	locationID := "location-id"
	studentID := "student-id"
	t.Run("err", func(t *testing.T) {
		args := []interface{}{mock.Anything, mock.Anything, freq, startTime, endTime, courseID, locationID, studentID}

		fields := []string{"purchased_slot_total"}
		var count pgtype.Int2
		values := []interface{}{&count}
		mockDB.MockQueryRowArgs(t, args...)
		mockDB.MockRowScanFields(errors.New("error"), fields, values)

		_, err := l.CountPurchasedSlotPerStudentSubscription(ctx, mockDB.DB, freq, startTime, endTime, courseID, locationID, studentID)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
			mockDB.Row,
		)
	})
}

func TestLessonAllocationRepo_CountAssignedSlotPerStudentCourse(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	l, mockDB := LessonAllocationRepoWithSqlMock()
	courseID := "course-id"
	studentID := "student-id"
	t.Run("err", func(t *testing.T) {
		args := []interface{}{mock.Anything, mock.Anything, studentID, courseID}

		fields := []string{"assigned_slot"}
		var count pgtype.Int4
		values := []interface{}{&count}
		mockDB.MockQueryRowArgs(t, args...)
		mockDB.MockRowScanFields(errors.New("error"), fields, values)

		_, err := l.CountAssignedSlotPerStudentCourse(ctx, mockDB.DB, studentID, courseID)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
			mockDB.Row,
		)
	})

	t.Run("success", func(t *testing.T) {
		args := []interface{}{mock.Anything, mock.Anything, studentID, courseID}

		fields := []string{"assigned_slot"}
		var count pgtype.Int4
		values := []interface{}{&count}
		mockDB.MockQueryRowArgs(t, args...)
		mockDB.MockRowScanFields(nil, fields, values)

		_, err := l.CountAssignedSlotPerStudentCourse(ctx, mockDB.DB, studentID, courseID)
		require.Nil(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
			mockDB.Row,
		)
	})
}
