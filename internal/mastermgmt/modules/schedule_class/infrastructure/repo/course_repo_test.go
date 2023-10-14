package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func CourseRepoWithSqlMock() (*CourseRepo, *testutil.MockDB) {
	courseRepo := &CourseRepo{}
	return courseRepo, testutil.NewMockDB()
}

func TestCourseRepo_GetMapCourseByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	courseRepo, mockDB := CourseRepoWithSqlMock()

	ids := []string{"01", "02"}

	t.Run("success", func(t *testing.T) {
		rc := &Course{}
		fields, value := rc.FieldMap()

		rc.Name.Set("course_name")
		rc.CourseID.Set("course_id")

		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &ids)
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			value,
		})
		resp, err := courseRepo.GetMapCourseByIDs(ctx, mockDB.DB, ids)
		expectedMapCourse := make(map[string]*Course)
		expectedMapCourse["course_id"] = &Course{
			CourseID: database.Text("course_id"),
			Name:    database.Text("course_name"),
		}
		require.NoError(t, err)
		require.Equal(t, expectedMapCourse, resp)
	})

	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &ids)

		resp, err := courseRepo.GetMapCourseByIDs(ctx, mockDB.DB, ids)
		require.True(t, errors.Is(err, puddle.ErrClosedPool))
		require.Nil(t, resp)
	})
}
