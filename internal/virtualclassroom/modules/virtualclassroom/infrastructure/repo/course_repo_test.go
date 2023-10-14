package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func CourseRepoWithSqlMock() (*CourseRepo, *testutil.MockDB) {
	r := &CourseRepo{}
	return r, testutil.NewMockDB()
}

func TestCourseRepo_GetValidCoursesByCourseIDsAndStatus(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	courseRepo, mockDB := CourseRepoWithSqlMock()
	mockCourse := &Course{}
	fields, values := mockCourse.FieldMap()

	courseIDs := []string{"course-id1", "course-id2"}
	status := domain.StatusActive

	t.Run("successful", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &courseIDs)
		mockDB.MockScanFields(nil, fields, values)

		courses, err := courseRepo.GetValidCoursesByCourseIDsAndStatus(ctx, mockDB.DB, courseIDs, status)
		assert.NoError(t, err)
		assert.NotNil(t, courses)
	})

	t.Run("failed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &courseIDs)

		courses, err := courseRepo.GetValidCoursesByCourseIDsAndStatus(ctx, mockDB.DB, courseIDs, status)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, courses)
	})
}
