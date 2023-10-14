package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func CourseRepoWithSqlMock() (*CourseRepo, *testutil.MockDB) {
	r := &CourseRepo{}
	return r, testutil.NewMockDB()
}

func TestCourseRepo_UpdateEndDateByCourseIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseRepoWithSqlMock()
	t.Run("successfully", func(t *testing.T) {
		courseIDs := []string{"course-id-1", "course-id-2", "course-id-3"}
		endDate := time.Now()
		args := []interface{}{
			mock.Anything, mock.AnythingOfType("string"), &courseIDs, &endDate, mock.AnythingOfType("Time"),
		}
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.UpdateEndDateByCourseIDs(ctx, mockDB.DB, courseIDs, endDate)
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})

	t.Run("got error", func(t *testing.T) {
		courseIDs := []string{"course-id-1", "course-id-2", "course-id-3"}
		endDate := time.Now()
		args := []interface{}{
			mock.Anything, mock.AnythingOfType("string"), &courseIDs, &endDate, mock.AnythingOfType("Time"),
		}
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), errors.New("error"), args...)

		err := r.UpdateEndDateByCourseIDs(ctx, mockDB.DB, courseIDs, endDate)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}
