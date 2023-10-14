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

func UserBasicInfoRepoWithSqlMock() (*UserBasicInfoRepo, *testutil.MockDB) {
	r := &UserBasicInfoRepo{}
	return r, testutil.NewMockDB()
}

func TestUserBasicInfoRepo_GetUserInfosByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	dto := &UserBasicInfo{}
	fields, values := dto.FieldMap()
	userIDs := []string{"user-id1", "user-id2", "user-id3"}

	t.Run("failed to get user infos", func(t *testing.T) {
		mockRepo, mockDB := UserBasicInfoRepoWithSqlMock()
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &userIDs)

		userInfos, err := mockRepo.GetUserInfosByIDs(ctx, mockDB.DB, userIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, userInfos)
	})

	t.Run("successfully get user infos", func(t *testing.T) {
		mockRepo, mockDB := UserBasicInfoRepoWithSqlMock()
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &userIDs)
		mockDB.MockScanFields(nil, fields, values)

		userInfos, err := mockRepo.GetUserInfosByIDs(ctx, mockDB.DB, userIDs)
		assert.Empty(t, err)
		assert.NotNil(t, userInfos)
		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})
}
