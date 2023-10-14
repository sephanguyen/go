package repo

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	virDomain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
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

func TestLessonRoomStateRepo_UpsertCurrentMaterial(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now().UTC()

	r, mockDB := LessonRoomStateRepoWithSqlMock()
	t.Run("successfully", func(t *testing.T) {
		mediaID := "media-id-1"
		expected := &domain.CurrentMaterial{
			LessonID:  "lesson-id-1",
			MediaID:   &mediaID,
			UpdatedAt: now,
			VideoState: &domain.VideoState{
				CurrentTime: domain.Duration(2 * time.Minute),
				PlayerState: domain.PlayerStatePlaying,
			},
		}

		lessonID := database.Text(expected.LessonID)
		mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, &lessonID, mock.Anything).
			Run(func(args mock.Arguments) {
				lessonRoomStateID := args[2].(*pgtype.Text)
				assert.NotEmpty(t, lessonRoomStateID.String)

				jsonb := args[4].(*pgtype.JSONB)
				state := &domain.CurrentMaterial{}
				err := jsonb.AssignTo(state)
				require.NoError(t, err)
				assert.Equal(t, *expected.MediaID, *state.MediaID)
				assert.Equal(t, expected.LessonID, state.LessonID)
				assert.Equal(t, expected.UpdatedAt, state.UpdatedAt)
				assert.Equal(t, expected.VideoState.CurrentTime, state.VideoState.CurrentTime)
				assert.Equal(t, expected.VideoState.PlayerState, state.VideoState.PlayerState)
			}).Return(nil, nil).Once()

		_, err := r.UpsertCurrentMaterial(ctx, mockDB.DB, expected)
		require.NoError(t, err)
	})

	t.Run("successfully without media id and video state", func(t *testing.T) {
		expected := &domain.CurrentMaterial{
			LessonID:  "lesson-id-1",
			UpdatedAt: now,
		}

		lessonID := database.Text(expected.LessonID)
		mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, &lessonID, mock.Anything).
			Run(func(args mock.Arguments) {
				lessonRoomStateID := args[2].(*pgtype.Text)
				assert.NotEmpty(t, lessonRoomStateID.String)

				jsonb := args[4].(*pgtype.JSONB)
				state := &domain.CurrentMaterial{}
				err := jsonb.AssignTo(state)
				require.NoError(t, err)
				assert.Nil(t, state.MediaID)
				assert.Equal(t, expected.LessonID, state.LessonID)
				assert.Equal(t, expected.UpdatedAt, state.UpdatedAt)
				assert.Nil(t, state.VideoState)
			}).Return(nil, nil).Once()

		_, err := r.UpsertCurrentMaterial(ctx, mockDB.DB, expected)
		require.NoError(t, err)
	})

	t.Run("has error", func(t *testing.T) {
		mediaID := "media-id-1"
		expected := &domain.CurrentMaterial{
			LessonID:  "lesson-id-1",
			MediaID:   &mediaID,
			UpdatedAt: now,
			VideoState: &domain.VideoState{
				CurrentTime: domain.Duration(2 * time.Minute),
				PlayerState: domain.PlayerStatePlaying,
			},
		}

		lessonID := database.Text(expected.LessonID)
		mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, &lessonID, mock.Anything).
			Run(func(args mock.Arguments) {
				lessonRoomStateID := args[2].(*pgtype.Text)
				assert.NotEmpty(t, lessonRoomStateID.String)

				jsonb := args[4].(*pgtype.JSONB)
				state := &domain.CurrentMaterial{}
				err := jsonb.AssignTo(state)
				require.NoError(t, err)
				assert.Equal(t, *expected.MediaID, *state.MediaID)
				assert.Equal(t, expected.LessonID, state.LessonID)
				assert.Equal(t, expected.UpdatedAt, state.UpdatedAt)
				assert.Equal(t, expected.VideoState.CurrentTime, state.VideoState.CurrentTime)
				assert.Equal(t, expected.VideoState.PlayerState, state.VideoState.PlayerState)
			}).Return(nil, errors.New("error")).Once()

		_, err := r.UpsertCurrentMaterial(ctx, mockDB.DB, expected)
		require.Error(t, err)
	})
}

func TestLessonRoomStateRepo_Spotlight(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	t.Run("success", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()
		lessonID := database.Text("lesson-id-1")
		spotlight := database.Text("user-id-1")
		mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, &lessonID, &spotlight).
			Return(pgconn.CommandTag([]byte(`0`)), nil)
		err := repo.Spotlight(ctx, mockDB.DB, lessonID, spotlight)
		require.NoError(t, err)
	})
	t.Run("failed", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()
		lessonID := database.Text("lesson-id-1")
		spotlight := database.Text("user-id-1")
		mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, &lessonID, &spotlight).
			Return(pgconn.CommandTag([]byte(`1`)), pgx.ErrTxClosed)
		err := repo.Spotlight(ctx, mockDB.DB, lessonID, spotlight)
		require.Error(t, err)
	})
}

func TestLessonRoomStateRepo_UnSpotlight(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	lessonID := database.Text("lesson-id-1")
	t.Run("success", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()
		mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), &lessonID).
			Return(pgconn.CommandTag([]byte(`0`)), nil)
		err := repo.UnSpotlight(ctx, mockDB.DB, lessonID)
		require.NoError(t, err)
	})
	t.Run("failed", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()
		mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), &lessonID).
			Return(pgconn.CommandTag([]byte(`1`)), pgx.ErrTxClosed)
		err := repo.UnSpotlight(ctx, mockDB.DB, lessonID)
		require.Error(t, err)
	})
}

func TestLessonRoomStateRepo_GetLessonRoomStateByLessonID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := LessonRoomStateRepoWithSqlMock()
	lessonID := database.Text("lesson-id-1")
	l := &LessonRoomState{}
	fields, value := l.FieldMap()
	t.Run("select not null value", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &lessonID)
		mockDB.MockRowScanFields(pgx.ErrTxClosed, fields, value)
		gotLessonRoomState, err := repo.GetLessonRoomStateByLessonID(ctx, mockDB.DB, lessonID)
		assert.ErrorIs(t, err, pgx.ErrTxClosed)
		assert.Nil(t, gotLessonRoomState)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &lessonID)
		mockDB.MockRowScanFields(nil, fields, value)
		gotLessonRoomState, err := repo.GetLessonRoomStateByLessonID(ctx, mockDB.DB, lessonID)
		assert.NoError(t, err)
		assert.NotNil(t, gotLessonRoomState)
	})

}

func TestLessonRoomStateRepo_UpsertWhiteboardZoomState(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	t.Run("success", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()
		lessonID := database.Text("lesson-id-1")
		w := &domain.WhiteboardZoomState{
			PdfScaleRatio: 23.32,
			CenterX:       243.5,
			CenterY:       -432.034,
			PdfWidth:      234.43,
			PdfHeight:     -0.33424,
		}

		mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, &lessonID, mock.MatchedBy(func(req *pgtype.JSONB) bool {
			var v domain.WhiteboardZoomState
			if err := json.Unmarshal(req.Bytes, &v); err != nil {
				return false
			}
			assert.EqualValues(t, *w, v)
			return true
		})).
			Return(pgconn.CommandTag([]byte(`0`)), nil)
		err := repo.UpsertWhiteboardZoomState(ctx, mockDB.DB, lessonID.String, w)
		require.NoError(t, err)
	})
	t.Run("failed", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()
		lessonID := database.Text("lesson-id-1")
		w := &domain.WhiteboardZoomState{
			PdfScaleRatio: 23.32,
			CenterX:       243.5,
			CenterY:       -432.034,
			PdfWidth:      234.43,
			PdfHeight:     -0.33424,
		}

		mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, &lessonID, mock.MatchedBy(func(req *pgtype.JSONB) bool {
			var v domain.WhiteboardZoomState
			if err := json.Unmarshal(req.Bytes, &v); err != nil {
				return false
			}
			assert.EqualValues(t, *w, v)
			return true
		})).Return(pgconn.CommandTag([]byte(`1`)), pgx.ErrTxClosed)
		err := repo.UpsertWhiteboardZoomState(ctx, mockDB.DB, lessonID.String, w)
		require.Error(t, err)
	})
}

func TestLessonRoomStateRepo_UpsertCurrentMaterialState(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	currentMaterialVideo := virDomain.CurrentMaterial{
		MediaID:   "media-id1",
		UpdatedAt: time.Now(),
		VideoState: &virDomain.VideoState{
			CurrentTime: virDomain.Duration(5),
			PlayerState: virDomain.PlayerStatePlaying,
		},
	}
	currentMaterialAudio := virDomain.CurrentMaterial{
		MediaID:   "media-id1",
		UpdatedAt: time.Now(),
		AudioState: &virDomain.AudioState{
			CurrentTime: virDomain.Duration(5),
			PlayerState: virDomain.PlayerStatePlaying,
		},
	}
	lessonID := database.Text("lesson-id-1")

	t.Run("success with video state", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()

		mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, &lessonID, mock.MatchedBy(func(req *pgtype.JSONB) bool {
			var v virDomain.CurrentMaterial
			if err := json.Unmarshal(req.Bytes, &v); err != nil {
				return false
			}
			assert.EqualValues(t, currentMaterialVideo.MediaID, v.MediaID)
			assert.EqualValues(t, currentMaterialVideo.VideoState.CurrentTime, v.VideoState.CurrentTime)
			assert.EqualValues(t, currentMaterialVideo.VideoState.PlayerState, v.VideoState.PlayerState)
			return true
		})).Return(pgconn.CommandTag([]byte(`0`)), nil)
		err := repo.UpsertCurrentMaterialState(ctx, mockDB.DB, lessonID, database.JSONB(currentMaterialVideo))
		require.NoError(t, err)
	})
	t.Run("success with audio state", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()

		mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, &lessonID, mock.MatchedBy(func(req *pgtype.JSONB) bool {
			var v virDomain.CurrentMaterial
			if err := json.Unmarshal(req.Bytes, &v); err != nil {
				return false
			}
			assert.EqualValues(t, currentMaterialAudio.MediaID, v.MediaID)
			assert.EqualValues(t, currentMaterialAudio.AudioState.CurrentTime, v.AudioState.CurrentTime)
			assert.EqualValues(t, currentMaterialAudio.AudioState.PlayerState, v.AudioState.PlayerState)
			return true
		})).Return(pgconn.CommandTag([]byte(`0`)), nil)
		err := repo.UpsertCurrentMaterialState(ctx, mockDB.DB, lessonID, database.JSONB(currentMaterialAudio))
		require.NoError(t, err)
	})
	t.Run("failed", func(t *testing.T) {
		repo, mockDB := LessonRoomStateRepoWithSqlMock()

		mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything, &lessonID, mock.MatchedBy(func(req *pgtype.JSONB) bool {
			var v virDomain.CurrentMaterial
			if err := json.Unmarshal(req.Bytes, &v); err != nil {
				return false
			}
			assert.EqualValues(t, currentMaterialVideo.MediaID, v.MediaID)
			assert.EqualValues(t, currentMaterialVideo.VideoState.CurrentTime, v.VideoState.CurrentTime)
			assert.EqualValues(t, currentMaterialVideo.VideoState.PlayerState, v.VideoState.PlayerState)
			return true
		})).Return(pgconn.CommandTag([]byte(`1`)), pgx.ErrTxClosed)
		err := repo.UpsertCurrentMaterialState(ctx, mockDB.DB, lessonID, database.JSONB(currentMaterialVideo))
		require.Error(t, err)
	})
}
