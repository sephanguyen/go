package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"
	"github.com/manabie-com/backend/mock/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

func LiveRoomRepoWithSqlMock() (*LiveRoomRepo, *testutil.MockDB) {
	r := &LiveRoomRepo{}
	return r, testutil.NewMockDB()
}

func genSliceMock(n int) []interface{} {
	result := []interface{}{}
	for i := 0; i < n; i++ {
		result = append(result, mock.Anything)
	}
	return result
}

func TestLiveRoomRepo_GetLiveRoomByChannelName(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	channelName := "channel name"
	mockLiveRoom := &LiveRoom{}
	fields, values := mockLiveRoom.FieldMap()

	t.Run("successful", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomRepoWithSqlMock()
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), &channelName)
		mockDB.MockRowScanFields(nil, fields, values)

		liveRoom, err := liveRoomRepo.GetLiveRoomByChannelName(ctx, mockDB.DB, channelName)
		assert.NoError(t, err)
		assert.NotNil(t, liveRoom)
	})

	t.Run("failed", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomRepoWithSqlMock()
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), &channelName)
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)

		liveRoom, err := liveRoomRepo.GetLiveRoomByChannelName(ctx, mockDB.DB, channelName)
		assert.True(t, errors.Is(err, domain.ErrChannelNotFound))
		assert.Nil(t, liveRoom)
	})
}

func TestLiveRoomRepo_CreateLiveRoom(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	roomID := "room-id1"
	channelName := "channel name"
	channelID := "channel-id1"
	mockLiveRoom := &LiveRoom{}
	fields := database.GetFieldNamesExcepts(mockLiveRoom, []string{"ended_at", "deleted_at"})

	args := append([]interface{}{
		mock.Anything,
		mock.AnythingOfType("string")},
		genSliceMock(len(fields))...)

	t.Run("insert failed", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, pgx.ErrTxClosed)

		err := liveRoomRepo.CreateLiveRoom(ctx, mockDB.DB, channelID, channelName, roomID)

		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("no rows affected after insert", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := liveRoomRepo.CreateLiveRoom(ctx, mockDB.DB, channelID, channelName, roomID)

		assert.Equal(t, err, domain.ErrNoChannelCreated)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("insert successful", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := liveRoomRepo.CreateLiveRoom(ctx, mockDB.DB, channelID, channelName, roomID)

		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestLiveRoomRepo_GetLiveRoomByChannelID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	channelID := "channel-id1"
	mockLiveRoom := &LiveRoom{}
	fields, values := mockLiveRoom.FieldMap()

	t.Run("successful", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomRepoWithSqlMock()
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), &channelID)
		mockDB.MockRowScanFields(nil, fields, values)

		liveRoom, err := liveRoomRepo.GetLiveRoomByChannelName(ctx, mockDB.DB, channelID)
		assert.NoError(t, err)
		assert.NotNil(t, liveRoom)
	})

	t.Run("failed", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomRepoWithSqlMock()
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), &channelID)
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)

		liveRoom, err := liveRoomRepo.GetLiveRoomByChannelName(ctx, mockDB.DB, channelID)
		assert.True(t, errors.Is(err, domain.ErrChannelNotFound))
		assert.Nil(t, liveRoom)
	})
}

func TestLiveRoomRepo_EndLiveRoom(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	channelID := "channel-id1"
	endTime := time.Now()
	var endedAt pgtype.Timestamptz
	endedAt.Set(endTime)

	args := append([]interface{}{
		mock.Anything,
		mock.AnythingOfType("string")},
		&endedAt,
		&channelID)

	t.Run("update failed", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, pgx.ErrTxClosed)

		err := liveRoomRepo.EndLiveRoom(ctx, mockDB.DB, channelID, endTime)

		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("no rows affected after update", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := liveRoomRepo.EndLiveRoom(ctx, mockDB.DB, channelID, endTime)

		assert.NotNil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update successful", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := liveRoomRepo.EndLiveRoom(ctx, mockDB.DB, channelID, endTime)

		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestLiveRoomRepo_UpdateChannelRoomID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	channelID := "channel-id1"
	roomID := "room-id"

	args := append([]interface{}{
		mock.Anything,
		mock.AnythingOfType("string")},
		&roomID,
		&channelID)

	t.Run("update failed", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, pgx.ErrTxClosed)

		err := liveRoomRepo.UpdateChannelRoomID(ctx, mockDB.DB, channelID, roomID)

		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("no rows affected after update", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := liveRoomRepo.UpdateChannelRoomID(ctx, mockDB.DB, channelID, roomID)

		assert.NotNil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update successful", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := liveRoomRepo.UpdateChannelRoomID(ctx, mockDB.DB, channelID, roomID)

		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
