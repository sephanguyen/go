package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func CourseClassRepoWithSqlMock() (*CourseClassRepo, *testutil.MockDB) {
	r := &CourseClassRepo{}
	return r, testutil.NewMockDB()
}

func TestCourseClassRepo_FindActiveCourseClassByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	courseClassRepo, mockDB := CourseClassRepoWithSqlMock()
	var (
		courseID pgtype.Text
		classID  pgtype.Int4
	)
	fields := []string{"course_id", "class_id"}
	values := []interface{}{&courseID, &classID}
	classIDs := []int32{123, 456, 789}

	t.Run("successful", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.Int4Array(classIDs), string(domain.CourseClassStatusActive))
		mockDB.MockScanFields(nil, fields, values)

		courseClasses, err := courseClassRepo.FindActiveCourseClassByID(ctx, mockDB.DB, classIDs)
		assert.NoError(t, err)
		assert.NotNil(t, courseClasses)
	})

	t.Run("failed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, database.Int4Array(classIDs), string(domain.CourseClassStatusActive))
		mockDB.MockScanFields(nil, fields, values)

		courseClasses, err := courseClassRepo.FindActiveCourseClassByID(ctx, mockDB.DB, classIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, courseClasses)
	})
}
