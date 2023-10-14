package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func LiveRoomRecordedVideosRepoWithSqlMock() (*LiveRoomRecordedVideosRepo, *testutil.MockDB) {
	r := &LiveRoomRecordedVideosRepo{}
	return r, testutil.NewMockDB()
}

func TestLiveRoomRecordedVideosRepo_InsertRecordedVideo(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	now := time.Now().UTC()
	recordedVideo := &vc_domain.RecordedVideo{
		ID:                 "recorded-video-id-1",
		RecordingChannelID: "channel-id-1",
		Description:        "description",
		DateTimeRecorded:   now,
		Creator:            "user-id-1",
	}
	err := recordedVideo.AssignByMediaFile(&media_domain.Media{
		ID:            "media-id-1",
		Resource:      "no need",
		Type:          media_domain.MediaTypeRecordingVideo,
		FileSizeBytes: 1000000,
		Duration:      200000,
	})
	require.NoError(t, err)

	recordedVideos := []*vc_domain.RecordedVideo{recordedVideo}

	t.Run("successfully", func(t *testing.T) {
		repo, _ := LiveRoomRecordedVideosRepoWithSqlMock()
		db := &mock_database.QueryExecer{}
		batchResults := &mock_database.BatchResults{}

		db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil)
		batchResults.On("Close").Once().Return(nil)

		err = repo.InsertRecordedVideos(ctx, db, recordedVideos)
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, db)
	})

	t.Run("failed", func(t *testing.T) {
		repo, _ := LiveRoomRecordedVideosRepoWithSqlMock()
		db := &mock_database.QueryExecer{}
		batchResults := &mock_database.BatchResults{}

		db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Return(pgconn.CommandTag("1"), puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)

		err = repo.InsertRecordedVideos(ctx, db, recordedVideos)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(t, db)
	})
}

func TestLiveRoomRecordedVideosRepo_GetLiveRoomRecordingsByChannelIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	channelIDs := []string{"channel-id1"}
	mockDTO := &LiveRoomRecordedVideo{
		RecordedVideoID: database.Text("record-id"),
		ChannelID:       database.Text("lesson-id1"),
	}
	fields, values := mockDTO.FieldMap()

	t.Run("success", func(t *testing.T) {
		repo, mockDB := LiveRoomRecordedVideosRepoWithSqlMock()
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &channelIDs)
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		result, err := repo.GetLiveRoomRecordingsByChannelIDs(ctx, mockDB.DB, channelIDs)
		assert.Nil(t, err)
		assert.Equal(t, LiveRoomRecordedVideos{mockDTO}.ToRecordedVideosEntity(), result)
	})

	t.Run("err select", func(t *testing.T) {
		repo, mockDB := LiveRoomRecordedVideosRepoWithSqlMock()
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &channelIDs)

		result, err := repo.GetLiveRoomRecordingsByChannelIDs(ctx, mockDB.DB, channelIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, result)
	})
}
