package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain/constant"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func LiveRoomActivityLogRepoWithSqlMock() (*LiveRoomActivityLogRepo, *testutil.MockDB) {
	r := &LiveRoomActivityLogRepo{}
	return r, testutil.NewMockDB()
}

func TestLiveRoomActivityLogRepo_CreateLog(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockRepo, mockDB := LiveRoomActivityLogRepoWithSqlMock()

	channelID := "channel-id1"
	userID := "user-id1"
	actionType := constant.LogActionTypePublish

	dto := &LiveRoomActivityLog{}
	fields, _ := dto.FieldMap()
	values := genSliceMock(len(fields))

	t.Run("error", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, nil, puddle.ErrClosedPool, args...)

		err := mockRepo.CreateLog(ctx, mockDB.DB, channelID, userID, actionType)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("no rows effected", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		err := mockRepo.CreateLog(ctx, mockDB.DB, channelID, userID, actionType)
		assert.EqualError(t, err, domain.ErrNoLiveRoomActivityLogCreated.Error())
	})

	t.Run("successful", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := mockRepo.CreateLog(ctx, mockDB.DB, channelID, userID, actionType)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertInsertedTable(t, dto.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})
}
