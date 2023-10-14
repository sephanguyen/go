package repo

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func LessonRoomStateRepoWithSqlMock() (*LessonRoomStateRepo, *testutil.MockDB) {
	r := &LessonRoomStateRepo{}
	return r, testutil.NewMockDB()
}

func TestLessonRoomStateRepo_UpsertRecordingState(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	w := &domain.CompositeRecordingState{
		ResourceID:  "resource-id",
		SID:         "s-id",
		UID:         2342334,
		IsRecording: true,
		Creator:     "user-id",
	}
	lessonID := database.Text("lesson-id-1")
	t.Run("success", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()
		mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, &lessonID, mock.MatchedBy(func(req *interface{}) bool {
			reqJSON := (*req).(pgtype.JSONB)
			var v domain.CompositeRecordingState
			if err := json.Unmarshal(reqJSON.Bytes, &v); err != nil {
				return false
			}
			assert.EqualValues(t, *w, v)
			return true
		})).
			Return(pgconn.CommandTag([]byte(`0`)), nil)
		err := repo.UpsertRecordingState(ctx, mockDB.DB, lessonID.String, w)
		require.NoError(t, err)
	})
	t.Run("failed", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()
		mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, &lessonID, mock.MatchedBy(func(req *interface{}) bool {
			reqJSON := (*req).(pgtype.JSONB)
			var v domain.CompositeRecordingState
			if err := json.Unmarshal(reqJSON.Bytes, &v); err != nil {
				return false
			}
			assert.EqualValues(t, *w, v)
			return true
		})).Return(pgconn.CommandTag([]byte(`1`)), pgx.ErrTxClosed)
		err := repo.UpsertRecordingState(ctx, mockDB.DB, lessonID.String, w)
		require.Error(t, err)
	})
}

func TestLessonRoomStateRepo_UpdateRecordingState(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	w := &domain.CompositeRecordingState{
		ResourceID:  "resource-id",
		SID:         "s-id",
		UID:         2342334,
		IsRecording: true,
		Creator:     "user-id",
	}
	lessonID := database.Text("lesson-id-1")

	t.Run("success", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()
		mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), &lessonID, mock.MatchedBy(func(req *pgtype.JSONB) bool {
			var v domain.CompositeRecordingState
			if err := json.Unmarshal(req.Bytes, &v); err != nil {
				return false
			}
			assert.EqualValues(t, *w, v)
			return true
		})).
			Return(pgconn.CommandTag([]byte(`0`)), nil)
		err := repo.UpdateRecordingState(ctx, mockDB.DB, lessonID.String, w)
		require.NoError(t, err)
	})
	t.Run("failed", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()
		mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), &lessonID, mock.MatchedBy(func(req *pgtype.JSONB) bool {
			var v domain.CompositeRecordingState
			if err := json.Unmarshal(req.Bytes, &v); err != nil {
				return false
			}
			assert.EqualValues(t, *w, v)
			return true
		})).Return(pgconn.CommandTag([]byte(`1`)), pgx.ErrTxClosed)
		err := repo.UpdateRecordingState(ctx, mockDB.DB, lessonID.String, w)
		require.Error(t, err)
	})
}

func TestLessonRoomStateRepo_UpsertWhiteboardZoomState(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonID := "lesson-id1"

	args := append([]interface{}{
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.Anything,
		mock.Anything,
		mock.Anything})

	wbZoomState := new(domain.WhiteboardZoomState).SetDefault()

	t.Run("upsert failed", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, pgx.ErrTxClosed)

		err := repo.UpsertWhiteboardZoomState(ctx, mockDB.DB, lessonID, wbZoomState)

		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("upsert successful", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := repo.UpsertWhiteboardZoomState(ctx, mockDB.DB, lessonID, wbZoomState)

		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestLessonRoomStateRepo_UpsertSpotlightState(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonID := "lesson-id1"
	userID := "learner-id1"

	args := append([]interface{}{
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.Anything,
		mock.Anything,
		mock.Anything})

	t.Run("upsert failed", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, pgx.ErrTxClosed)

		err := repo.UpsertSpotlightState(ctx, mockDB.DB, lessonID, userID)

		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("upsert successful", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := repo.UpsertSpotlightState(ctx, mockDB.DB, lessonID, userID)

		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestLessonRoomStateRepo_UpsertCurrentMaterialState(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	currentMaterialVideo := &domain.CurrentMaterial{
		MediaID:   "media-id1",
		UpdatedAt: time.Now(),
		VideoState: &domain.VideoState{
			CurrentTime: domain.Duration(5),
			PlayerState: domain.PlayerStatePlaying,
		},
	}
	currentMaterialAudio := &domain.CurrentMaterial{
		MediaID:   "media-id1",
		UpdatedAt: time.Now(),
		AudioState: &domain.AudioState{
			CurrentTime: domain.Duration(5),
			PlayerState: domain.PlayerStatePlaying,
		},
	}
	lessonID := database.Text("lesson-id-1")

	t.Run("success with video state", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()
		mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, &lessonID, mock.MatchedBy(func(req *interface{}) bool {
			reqJSON := (*req).(pgtype.JSONB)
			var v domain.CurrentMaterial
			if err := json.Unmarshal(reqJSON.Bytes, &v); err != nil {
				return false
			}
			assert.EqualValues(t, currentMaterialVideo.MediaID, v.MediaID)
			assert.EqualValues(t, currentMaterialVideo.VideoState.CurrentTime, v.VideoState.CurrentTime)
			assert.EqualValues(t, currentMaterialVideo.VideoState.PlayerState, v.VideoState.PlayerState)
			return true
		})).Return(pgconn.CommandTag([]byte(`0`)), nil)

		err := repo.UpsertCurrentMaterialState(ctx, mockDB.DB, lessonID.String, currentMaterialVideo)
		require.NoError(t, err)
	})
	t.Run("success with audio state", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()
		mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, &lessonID, mock.MatchedBy(func(req *interface{}) bool {
			reqJSON := (*req).(pgtype.JSONB)
			var v domain.CurrentMaterial
			if err := json.Unmarshal(reqJSON.Bytes, &v); err != nil {
				return false
			}
			assert.EqualValues(t, currentMaterialAudio.MediaID, v.MediaID)
			assert.EqualValues(t, currentMaterialAudio.AudioState.CurrentTime, v.AudioState.CurrentTime)
			assert.EqualValues(t, currentMaterialAudio.AudioState.PlayerState, v.AudioState.PlayerState)
			return true
		})).Return(pgconn.CommandTag([]byte(`0`)), nil)

		err := repo.UpsertCurrentMaterialState(ctx, mockDB.DB, lessonID.String, currentMaterialAudio)
		require.NoError(t, err)
	})
	t.Run("failed", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()
		mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, &lessonID, mock.MatchedBy(func(req *interface{}) bool {
			reqJSON := (*req).(pgtype.JSONB)
			var v domain.CurrentMaterial
			if err := json.Unmarshal(reqJSON.Bytes, &v); err != nil {
				return false
			}
			assert.EqualValues(t, currentMaterialVideo.MediaID, v.MediaID)
			assert.EqualValues(t, currentMaterialVideo.VideoState.CurrentTime, v.VideoState.CurrentTime)
			assert.EqualValues(t, currentMaterialVideo.VideoState.PlayerState, v.VideoState.PlayerState)
			return true
		})).Return(pgconn.CommandTag([]byte(`1`)), pgx.ErrTxClosed)

		err := repo.UpsertCurrentMaterialState(ctx, mockDB.DB, lessonID.String, currentMaterialVideo)
		require.Error(t, err)
	})
}

func TestLessonRoomStateRepo_UpsertLiveRoomSessionTime(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonID := database.Text("lesson-id-1")
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()
		mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, &lessonID, mock.Anything).
			Return(pgconn.CommandTag([]byte(`0`)), nil)
		err := repo.UpsertLiveLessonSessionTime(ctx, mockDB.DB, lessonID.String, now)
		require.NoError(t, err)
	})
	t.Run("failed", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()
		mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, &lessonID, mock.Anything).
			Return(pgconn.CommandTag([]byte(`1`)), pgx.ErrTxClosed)
		err := repo.UpsertLiveLessonSessionTime(ctx, mockDB.DB, lessonID.String, now)
		require.Error(t, err)
	})
}

func TestLessonRoomStateRepo_GetLessonRoomStateByLessonID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonID := "lesson-id-1"
	l := &LessonRoomState{}
	fields, value := l.FieldMap()

	t.Run("select not null value", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &lessonID)
		mockDB.MockRowScanFields(pgx.ErrTxClosed, fields, value)

		lessonRoomState, err := repo.GetLessonRoomStateByLessonID(ctx, mockDB.DB, lessonID)
		assert.ErrorIs(t, err, pgx.ErrTxClosed)
		assert.Nil(t, lessonRoomState)
	})

	t.Run("failed with no rows found", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &lessonID)
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, value)

		lessonRoomState, err := repo.GetLessonRoomStateByLessonID(ctx, mockDB.DB, lessonID)
		assert.ErrorIs(t, err, domain.ErrLessonRoomStateNotFound)
		assert.NotNil(t, lessonRoomState)
	})

	t.Run("success", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &lessonID)
		mockDB.MockRowScanFields(nil, fields, value)

		lessonRoomState, err := repo.GetLessonRoomStateByLessonID(ctx, mockDB.DB, lessonID)
		assert.NoError(t, err)
		assert.NotNil(t, lessonRoomState)
	})

}
