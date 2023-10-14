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

func StudentParentRepoWithSqlMock() (*StudentParentRepo, *testutil.MockDB) {
	r := &StudentParentRepo{}
	return r, testutil.NewMockDB()
}

func TestStudentParentRepo_GetStudentParents(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := StudentParentRepoWithSqlMock()
	dto := &StudentParent{}
	fields, values := dto.FieldMap()

	studentIDs := []string{"student-id1", "student-id2", "student-id3"}

	t.Run("successful", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, studentIDs)
		mockDB.MockScanFields(nil, fields, values)

		studentParents, err := repo.GetStudentParents(ctx, mockDB.DB, studentIDs)
		assert.NoError(t, err)
		assert.NotNil(t, studentParents)
	})

	t.Run("failed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, studentIDs)

		studentParents, err := repo.GetStudentParents(ctx, mockDB.DB, studentIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, studentParents)
	})
}
