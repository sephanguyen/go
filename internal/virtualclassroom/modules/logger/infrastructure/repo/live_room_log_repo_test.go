package repo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func LiveRoomLogRepoWithSqlMock() (*LiveRoomLogRepo, *testutil.MockDB) {
	l := &LiveRoomLogRepo{}
	return l, testutil.NewMockDB()
}

func TestLiveRoomLogRepo_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	dto := &LiveRoomLog{}
	_, values := dto.FieldMap()
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)

	t.Run("insert failed", func(t *testing.T) {
		mockRepo, mockDB := LiveRoomLogRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, puddle.ErrNotAvailable)

		err := mockRepo.Create(ctx, mockDB.DB, dto)
		assert.True(t, errors.Is(err, puddle.ErrNotAvailable))
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("no rows affected", func(t *testing.T) {
		mockRepo, mockDB := LiveRoomLogRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := mockRepo.Create(ctx, mockDB.DB, dto)
		assert.EqualError(t, err, "cannot insert new live_room_log")
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("insert success", func(t *testing.T) {
		mockRepo, mockDB := LiveRoomLogRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := mockRepo.Create(ctx, mockDB.DB, dto)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestLiveRoomLogRepo_AddAttendeeIDByChannelID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	channelID := "channel-id1"
	attendeeID := "user-id1"
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, channelID, attendeeID)

	t.Run("update failed", func(t *testing.T) {
		mockRepo, mockDB := LiveRoomLogRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, puddle.ErrNotAvailable)

		err := mockRepo.AddAttendeeIDByChannelID(ctx, mockDB.DB, channelID, attendeeID)
		assert.True(t, errors.Is(err, puddle.ErrNotAvailable))
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update success", func(t *testing.T) {
		mockRepo, mockDB := LiveRoomLogRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := mockRepo.AddAttendeeIDByChannelID(ctx, mockDB.DB, channelID, attendeeID)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestLiveRoomLogRepo_IncreaseTotalTimesByChannelID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	channelID := "channel-id1"
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, channelID)

	t.Run("update failed", func(t *testing.T) {
		logType := TotalTimesReconnection
		mockRepo, mockDB := LiveRoomLogRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, puddle.ErrNotAvailable)

		err := mockRepo.IncreaseTotalTimesByChannelID(ctx, mockDB.DB, channelID, logType)
		assert.True(t, errors.Is(err, puddle.ErrNotAvailable))
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update success", func(t *testing.T) {
		logType := TotalTimesUpdatingRoomState
		mockRepo, mockDB := LiveRoomLogRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := mockRepo.IncreaseTotalTimesByChannelID(ctx, mockDB.DB, channelID, logType)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("unsupported log type", func(t *testing.T) {
		logType := TotalTimes(10000)
		mockRepo, mockDB := LiveRoomLogRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := mockRepo.IncreaseTotalTimesByChannelID(ctx, mockDB.DB, channelID, logType)
		assert.EqualError(t, err, fmt.Sprintf("live room log type unsupported %v", logType))
	})
}

func TestLiveRoomLogRepo_CompleteLogByChannelID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	channelID := "channel-id1"
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, channelID)

	t.Run("update failed", func(t *testing.T) {
		mockRepo, mockDB := LiveRoomLogRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, puddle.ErrNotAvailable)

		err := mockRepo.CompleteLogByChannelID(ctx, mockDB.DB, channelID)
		assert.True(t, errors.Is(err, puddle.ErrNotAvailable))
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update success", func(t *testing.T) {
		mockRepo, mockDB := LiveRoomLogRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := mockRepo.CompleteLogByChannelID(ctx, mockDB.DB, channelID)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestLiveRoomLogRepo_GetLatestByChannelID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	channelID := "channel-id1"
	dto := &LiveRoomLog{}
	fields, values := dto.FieldMap()

	t.Run("failed", func(t *testing.T) {
		mockRepo, mockDB := LiveRoomLogRepoWithSqlMock()
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), &channelID)
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)

		results, err := mockRepo.GetLatestByChannelID(ctx, mockDB.DB, channelID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, results)
	})

	t.Run("successful", func(t *testing.T) {
		mockRepo, mockDB := LiveRoomLogRepoWithSqlMock()
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), &channelID)
		mockDB.MockRowScanFields(nil, fields, values)

		results, err := mockRepo.GetLatestByChannelID(ctx, mockDB.DB, channelID)
		assert.Nil(t, err)
		assert.NotNil(t, results)
	})
}
