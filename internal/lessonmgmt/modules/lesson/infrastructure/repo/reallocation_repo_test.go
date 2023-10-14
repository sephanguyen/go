package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ReallocationRepoWithSqlMock() (*ReallocationRepo, *testutil.MockDB) {
	r := &ReallocationRepo{}
	return r, testutil.NewMockDB()
}

func TestReallocationRepo_GetFollowingReallocation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := ReallocationRepoWithSqlMock()
	e := &Reallocation{}
	fields, value := e.FieldMap()
	studentIds := []string{"student-id"}
	lessonID := "1"

	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &lessonID, &studentIds)
		gotResults, err := r.GetFollowingReallocation(ctx, mockDB.DB, lessonID, studentIds)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, gotResults)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &lessonID, &studentIds)
		mockDB.MockScanFields(nil, fields, value)
		gotResults, err := r.GetFollowingReallocation(ctx, mockDB.DB, lessonID, studentIds)
		assert.NoError(t, err)
		assert.NotNil(t, gotResults)
	})
}

func TestReallocationRepo_GetByNewLessonID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := ReallocationRepoWithSqlMock()
	e := &Reallocation{}
	fields, value := e.FieldMap()
	studentIds := []string{"student-id"}
	lessonID := "1"

	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, studentIds, lessonID)
		gotResults, err := r.GetByNewLessonID(ctx, mockDB.DB, studentIds, lessonID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, gotResults)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, studentIds, lessonID)
		mockDB.MockScanFields(nil, fields, value)
		gotResults, err := r.GetByNewLessonID(ctx, mockDB.DB, studentIds, lessonID)
		assert.NoError(t, err)
		assert.NotNil(t, gotResults)
	})

}
