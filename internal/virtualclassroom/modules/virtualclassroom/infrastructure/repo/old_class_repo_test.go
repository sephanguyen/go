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

func OldClassRepoWithSqlMock() (*OldClassRepo, *testutil.MockDB) {
	r := &OldClassRepo{}
	return r, testutil.NewMockDB()
}

func TestOldClassRepo_FindJoined(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	oldClassRepo, mockDB := OldClassRepoWithSqlMock()
	oldClass := &OldClass{}
	fields, values := oldClass.FieldMap()

	userID := "user_id1"

	t.Run("successful", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &userID, string(domain.ClassStatusActive), string(domain.ClassMemberStatusActive))
		mockDB.MockScanFields(nil, fields, values)

		oldClasses, err := oldClassRepo.FindJoined(ctx, mockDB.DB, userID)
		assert.NoError(t, err)
		assert.NotNil(t, oldClasses)
	})

	t.Run("failed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &userID, string(domain.ClassStatusActive), string(domain.ClassMemberStatusActive))

		oldClasses, err := oldClassRepo.FindJoined(ctx, mockDB.DB, userID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, oldClasses)
	})
}
