package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func LiveRoomStateRepoWithSqlMock() (*LiveRoomStateRepo, *testutil.MockDB) {
	r := &LiveRoomStateRepo{}
	return r, testutil.NewMockDB()
}

func TestLiveRoomStateRepo_GetLiveRoomStateByChannelID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	channelID := "channel-id1"
	mockDTO := &LiveRoomState{}
	fields, values := mockDTO.FieldMap()

	t.Run("successful", func(t *testing.T) {
		liveRoomStateRepo, mockDB := LiveRoomStateRepoWithSqlMock()
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), &channelID)
		mockDB.MockRowScanFields(nil, fields, values)

		liveRoomState, err := liveRoomStateRepo.GetLiveRoomStateByChannelID(ctx, mockDB.DB, channelID)
		assert.NoError(t, err)
		assert.NotNil(t, liveRoomState)
	})

	t.Run("failed with no rows found", func(t *testing.T) {
		liveRoomStateRepo, mockDB := LiveRoomStateRepoWithSqlMock()
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), &channelID)
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)

		liveRoomState, err := liveRoomStateRepo.GetLiveRoomStateByChannelID(ctx, mockDB.DB, channelID)
		assert.True(t, errors.Is(err, domain.ErrChannelNotFound))
		assert.NotNil(t, liveRoomState)
	})

	t.Run("failed", func(t *testing.T) {
		liveRoomStateRepo, mockDB := LiveRoomStateRepoWithSqlMock()
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), &channelID)
		mockDB.MockRowScanFields(pgx.ErrTxClosed, fields, values)

		liveRoomState, err := liveRoomStateRepo.GetLiveRoomStateByChannelID(ctx, mockDB.DB, channelID)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Nil(t, liveRoomState)
	})
}

func TestLiveRoomStateRepo_UpsertLiveRoomState(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	channelID := "channel-id1"
	mockDomainObject := &vc_domain.CurrentMaterial{
		MediaID:   "media-id1",
		UpdatedAt: time.Now(),
	}
	mockDTO := &LiveRoomState{
		CurrentMaterial: database.JSONB(mockDomainObject),
	}
	mockDomainObjectField := "current_material"

	fields := []string{"live_room_state_id", "channel_id", mockDomainObjectField}
	args := append([]interface{}{
		mock.Anything,
		mock.AnythingOfType("string")},
		genSliceMock(len(fields))...)

	t.Run("upsert failed", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomStateRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, pgx.ErrTxClosed)

		err := liveRoomRepo.UpsertLiveRoomState(ctx, mockDB.DB, channelID, mockDTO.CurrentMaterial, mockDomainObjectField)

		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("upsert successful", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomStateRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := liveRoomRepo.UpsertLiveRoomState(ctx, mockDB.DB, channelID, mockDTO.CurrentMaterial, mockDomainObjectField)

		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestLiveRoomStateRepo_UnSpotlight(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	channelID := "channel-id1"

	args := append([]interface{}{
		mock.Anything,
		mock.AnythingOfType("string")},
		mock.Anything)

	t.Run("upsert failed", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomStateRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, pgx.ErrTxClosed)

		err := liveRoomRepo.UnSpotlight(ctx, mockDB.DB, channelID)

		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("upsert successful", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomStateRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := liveRoomRepo.UnSpotlight(ctx, mockDB.DB, channelID)

		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestLiveRoomStateRepo_GetStreamingLearners(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var streamingLearners pgtype.TextArray
	fields, values := []string{"streaming_learners"}, []interface{}{&streamingLearners}
	channelID := "channel-id1"

	t.Run("successful", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomStateRepoWithSqlMock()
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &channelID)
		mockDB.MockRowScanFields(nil, fields, values)

		_, err := liveRoomRepo.GetStreamingLearners(ctx, mockDB.DB, channelID, true)
		assert.NoError(t, err)
	})

	t.Run("failed", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomStateRepoWithSqlMock()
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &channelID)
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)

		_, err := liveRoomRepo.GetStreamingLearners(ctx, mockDB.DB, channelID, true)
		assert.True(t, errors.Is(err, domain.ErrChannelNotFound))
	})
}

func TestLiveRoomStateRepo_IncreaseNumberOfStreaming(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	learnerID := "student-id1"
	learnerIDArray := []string{learnerID}
	channelID := "channel-id1"
	maximumLearnerStreamings := 20

	args := append([]interface{}{
		mock.Anything,
		mock.AnythingOfType("string")},
		mock.Anything,
		mock.Anything,
		&learnerIDArray,
		&learnerID,
		&maximumLearnerStreamings)

	t.Run("upsert failed", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomStateRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, pgx.ErrTxClosed)

		err := liveRoomRepo.IncreaseNumberOfStreaming(ctx, mockDB.DB, channelID, learnerID, maximumLearnerStreamings)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
	})

	t.Run("no rows affected after upsert", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomStateRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := liveRoomRepo.IncreaseNumberOfStreaming(ctx, mockDB.DB, channelID, learnerID, maximumLearnerStreamings)
		assert.EqualError(t, err, domain.ErrNoChannelUpdated.Error())
	})

	t.Run("upsert successful", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomStateRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := liveRoomRepo.IncreaseNumberOfStreaming(ctx, mockDB.DB, channelID, learnerID, maximumLearnerStreamings)
		assert.Nil(t, err)
	})
}

func TestLiveRoomStateRepo_DecreaseNumberOfStreaming(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	learnerID := "student-id1"
	channelID := "channel-id1"

	args := append([]interface{}{
		mock.Anything,
		mock.AnythingOfType("string")},
		&channelID,
		&learnerID)

	t.Run("update failed", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomStateRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, pgx.ErrTxClosed)

		err := liveRoomRepo.DecreaseNumberOfStreaming(ctx, mockDB.DB, channelID, learnerID)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
	})

	t.Run("no rows affected after update", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomStateRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := liveRoomRepo.DecreaseNumberOfStreaming(ctx, mockDB.DB, channelID, learnerID)
		assert.EqualError(t, err, domain.ErrNoChannelUpdated.Error())
	})

	t.Run("update successful", func(t *testing.T) {
		liveRoomRepo, mockDB := LiveRoomStateRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := liveRoomRepo.DecreaseNumberOfStreaming(ctx, mockDB.DB, channelID, learnerID)
		assert.Nil(t, err)
	})
}
