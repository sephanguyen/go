package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/queries/payloads"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func RecordedVideoRepoWithSqlMock() (*RecordedVideoRepo, *testutil.MockDB) {
	r := &RecordedVideoRepo{}
	return r, testutil.NewMockDB()
}

func TestRecordedVideoRepo_InsertRecordedVideo(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.QueryExecer{}
	now := time.Now().UTC()
	r, _ := RecordedVideoRepoWithSqlMock()
	t.Run("successfully", func(t *testing.T) {
		e := &domain.RecordedVideo{
			ID:                 "recorded-video-id-1",
			RecordingChannelID: "lesson-id-1",
			Description:        "description",
			DateTimeRecorded:   now,
			Creator:            "user-id-1",
		}
		err := e.AssignByMediaFile(&media_domain.Media{
			ID:            "media-id-1",
			Resource:      "no need",
			Type:          media_domain.MediaTypeRecordingVideo,
			FileSizeBytes: 1000000,
			Duration:      200000,
		})

		es := []*domain.RecordedVideo{e}
		require.NoError(t, err)

		batchResults := &mock_database.BatchResults{}
		db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Return(pgconn.CommandTag("1"), nil)
		batchResults.On("Close").Once().Return(nil)

		err = r.InsertRecordedVideos(ctx, db, es)
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(
			t,
			db,
		)
	})

	t.Run("has error", func(t *testing.T) {
		e := &domain.RecordedVideo{
			ID:                 "recorded-video-id-1",
			RecordingChannelID: "lesson-id-1",
			Description:        "description",
			DateTimeRecorded:   now,
			Creator:            "user-id-1",
		}
		err := e.AssignByMediaFile(&media_domain.Media{
			ID:            "media-id-1",
			Resource:      "no need",
			Type:          media_domain.MediaTypeRecordingVideo,
			FileSizeBytes: 1000000,
			Duration:      200000,
		})
		es := []*domain.RecordedVideo{e}
		require.NoError(t, err)

		batchResults := &mock_database.BatchResults{}
		db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Return(pgconn.CommandTag("1"), puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)

		err = r.InsertRecordedVideos(ctx, db, es)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			db,
		)
	})
}

func TestRecordedVideo_ToRecordedVideoEntity(t *testing.T) {
	t.Parallel()
	now := time.Time{}
	tcs := []struct {
		name     string
		dto      RecordedVideos
		expected domain.RecordedVideos
	}{
		{
			name: "full fields",
			dto: RecordedVideos{
				{
					RecordedVideoID:  database.Text("recorded-video-id-1"),
					LessonID:         database.Text("lesson-id-1"),
					Description:      database.Text("description 1"),
					DateTimeRecorded: database.Timestamptz(now),
					Creator:          database.Text("user-id-1"),
					CreatedAt:        database.Timestamptz(now),
					UpdatedAt:        database.Timestamptz(now),
					MediaID:          database.Text("media-id-1"),
				},
				{
					RecordedVideoID:  database.Text("recorded-video-id-2"),
					LessonID:         database.Text("lesson-id-2"),
					Description:      database.Text("description 2"),
					DateTimeRecorded: database.Timestamptz(now),
					Creator:          database.Text("user-id-2"),
					CreatedAt:        database.Timestamptz(now),
					UpdatedAt:        database.Timestamptz(now),
					MediaID:          database.Text("media-id-3"),
				},
			},
			expected: domain.RecordedVideos{
				{
					ID:                 "recorded-video-id-1",
					RecordingChannelID: "lesson-id-1",
					Description:        "description 1",
					DateTimeRecorded:   now,
					Creator:            "user-id-1",
					CreatedAt:          now,
					UpdatedAt:          now,
					Media: &media_domain.Media{
						ID: "media-id-1",
					},
				},
				{
					ID:                 "recorded-video-id-2",
					RecordingChannelID: "lesson-id-2",
					Description:        "description 2",
					DateTimeRecorded:   now,
					Creator:            "user-id-2",
					CreatedAt:          now,
					UpdatedAt:          now,
					Media: &media_domain.Media{
						ID: "media-id-3",
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.dto.ToRecordedVideosEntity()
			assert.EqualValues(t, tc.expected, actual)
		})
	}
}

func TestRecordedVideo_ListRecordingByLessonID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := RecordedVideoRepoWithSqlMock()

	t.Run("success with select", func(t *testing.T) {
		args := &payloads.GetRecordingByIDPayload{
			RecordedVideoID: "recorded-video-id",
		}

		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), mock.Anything)

		e := &RecordedVideo{}
		_, values := e.FieldMap()
		mockDB.Row.On("Scan", values...).Once().Return(nil)

		_, err := r.GetRecordingByID(ctx, mockDB.DB, args)
		assert.Nil(t, err)
	})

	t.Run("fail case", func(t *testing.T) {
		args := &payloads.GetRecordingByIDPayload{
			RecordedVideoID: "recorded-video-id",
		}

		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), mock.Anything)

		e := &RecordedVideo{}
		_, values := e.FieldMap()
		mockDB.Row.On("Scan", values...).Once().Return(pgx.ErrNoRows)

		_, err := r.GetRecordingByID(ctx, mockDB.DB, args)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
	})

}

func TestRecordedVideo_ListRecordingByLessonIDWithPaging(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := RecordedVideoRepoWithSqlMock()

	t.Run("success with select", func(t *testing.T) {
		args := &payloads.RetrieveRecordedVideosByLessonIDPayload{
			Limit:    2,
			LessonID: "lesson-id",
		}
		rows := mockDB.Rows
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

		mockDB.Rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)

		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			mock.Anything, mock.Anything, mock.Anything)
		mockDB.Row.On("Scan", mock.Anything).Once().Return(nil)

		_, _, _, _, err := r.ListRecordingByLessonIDWithPaging(ctx, mockDB.DB, args)
		assert.Nil(t, err)
	})

	t.Run("success with select by RecordedVideoID", func(t *testing.T) {
		args := &payloads.RetrieveRecordedVideosByLessonIDPayload{
			Limit:           2,
			LessonID:        "lesson-id",
			RecordedVideoID: "recording-1",
		}
		rows := mockDB.Rows
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

		mockDB.Rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)

		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		mockDB.Row.On("Scan", mock.Anything, mock.Anything).Once().Return(nil)

		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		mockDB.Row.On("Scan", mock.Anything).Once().Return(nil)

		_, _, _, _, err := r.ListRecordingByLessonIDWithPaging(ctx, mockDB.DB, args)
		assert.Nil(t, err)
	})

	t.Run("fail case", func(t *testing.T) {
		args := &payloads.RetrieveRecordedVideosByLessonIDPayload{
			Limit:    2,
			LessonID: "lesson-id",
		}
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		_, _, _, _, err := r.ListRecordingByLessonIDWithPaging(ctx, mockDB.DB, args)
		assert.NotNil(t, err)
	})

}

func TestRecordedVideoRepo_Delete(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := RecordedVideoRepoWithSqlMock()
	recordId := []string{"record-id"}

	t.Run("err delete", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &recordId)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := r.DeleteRecording(ctx, mockDB.DB, recordId)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &recordId)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.DeleteRecording(ctx, mockDB.DB, recordId)
		assert.Nil(t, err)
	})
}

func TestListRecordingByLessonIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := RecordedVideoRepoWithSqlMock()
	lessonIDs := []string{"lesson-id-1"}

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &lessonIDs)
		e := &RecordedVideo{
			RecordedVideoID: database.Text("record-id"),
			LessonID:        database.Text("lesson-id "),
		}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		result, err := r.ListRecordingByLessonIDs(ctx, mockDB.DB, lessonIDs)
		assert.Nil(t, err)
		assert.Equal(t, RecordedVideos{e}.ToRecordedVideosEntity(), result)
	})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &lessonIDs)
		result, err := r.ListRecordingByLessonIDs(ctx, mockDB.DB, lessonIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, result)
	})
}
