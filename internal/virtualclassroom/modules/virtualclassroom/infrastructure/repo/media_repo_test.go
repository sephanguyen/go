package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	testing_util "github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func MediaRepoWithSqlMock() (*MediaRepo, *testing_util.MockDB) {
	r := &MediaRepo{}
	return r, testing_util.NewMockDB()
}
func TestMediaRepo_InsertMedia(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := MediaRepoWithSqlMock()
	t.Run("successfully", func(t *testing.T) {
		media := &domain.Media{
			ID:       "media-id-1",
			Name:     "media 1",
			Resource: "video-id-1",
			Type:     domain.MediaTypeVideo,
		}
		dto := Media{
			MediaID:  database.Text("media-id-1"),
			Name:     database.Text("media 1"),
			Resource: database.Text("video-id-1"),
			Type:     database.Text(string(domain.MediaTypeVideo)),
		}
		args := append([]interface{}{
			mock.Anything,
			mock.AnythingOfType("string")},
			&dto.MediaID, &dto.Name, &dto.Resource,
			mock.Anything, &dto.Type, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		_, err := r.InsertMedia(ctx, mockDB.DB, media)
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})

	t.Run("has error", func(t *testing.T) {
		media := &domain.Media{
			ID:       "media-id-1",
			Name:     "media 1",
			Resource: "video-id-1",
			Type:     domain.MediaTypeVideo,
		}
		dto := Media{
			MediaID:  database.Text("media-id-1"),
			Name:     database.Text("media 1"),
			Resource: database.Text("video-id-1"),
			Type:     database.Text(string(domain.MediaTypeVideo)),
		}
		args := append([]interface{}{
			mock.Anything,
			mock.AnythingOfType("string")},
			&dto.MediaID, &dto.Name, &dto.Resource,
			mock.Anything, &dto.Type, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), errors.New("error"), args...)

		_, err := r.InsertMedia(ctx, mockDB.DB, media)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}

func TestMediaRepo_ListByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	now := time.Now().UTC()
	r, mockDB := MediaRepoWithSqlMock()
	t.Run("successfully", func(t *testing.T) {
		ids := []string{"media-id-1", "media-id-2"}
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			ids,
		)
		e1 := &Media{
			MediaID:   database.Text("media-id-1"),
			Name:      database.Text("media 1"),
			Resource:  database.Text("video-id-1"),
			Type:      database.Text(string(domain.MediaTypeVideo)),
			Comments:  database.JSONB(nil),
			CreatedAt: database.Timestamptz(now),
			UpdatedAt: database.Timestamptz(now),
			DeletedAt: pgtype.Timestamptz{
				Status: pgtype.Null,
			},
			ConvertedImages: database.JSONB(nil),
		}
		fields, v1 := e1.FieldMap()
		e2 := &Media{
			MediaID:   database.Text("media-id-2"),
			Name:      database.Text("media 2"),
			Resource:  database.Text("video-id-2"),
			Type:      database.Text(string(domain.MediaTypeVideo)),
			Comments:  database.JSONB(nil),
			CreatedAt: database.Timestamptz(now),
			UpdatedAt: database.Timestamptz(now),
			DeletedAt: pgtype.Timestamptz{
				Status: pgtype.Null,
			},
			ConvertedImages: database.JSONB(nil),
		}
		_, v2 := e2.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			v1,
			v2,
		})
		actual, err := r.ListByIDs(ctx, mockDB.DB, ids)
		require.NoError(t, err)
		assert.EqualValues(t, domain.Medias{
			{
				ID:        "media-id-1",
				Name:      "media 1",
				Resource:  "video-id-1",
				Type:      domain.MediaTypeVideo,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				ID:        "media-id-2",
				Name:      "media 2",
				Resource:  "video-id-2",
				Type:      domain.MediaTypeVideo,
				CreatedAt: now,
				UpdatedAt: now,
			},
		}, actual)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})

	t.Run("has error", func(t *testing.T) {
		ids := []string{"media-id-1", "media-id-2"}
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything,
			mock.AnythingOfType("string"),
			ids,
		)
		e1 := &Media{
			MediaID:   database.Text("media-id-1"),
			Name:      database.Text("media 1"),
			Resource:  database.Text("video-id-1"),
			Type:      database.Text(string(domain.MediaTypeVideo)),
			CreatedAt: database.Timestamptz(now),
			UpdatedAt: database.Timestamptz(now),
			DeletedAt: pgtype.Timestamptz{
				Status: pgtype.Null,
			},
		}
		fields, v1 := e1.FieldMap()
		e2 := &Media{
			MediaID:   database.Text("media-id-2"),
			Name:      database.Text("media 2"),
			Resource:  database.Text("video-id-2"),
			Type:      database.Text(string(domain.MediaTypeVideo)),
			CreatedAt: database.Timestamptz(now),
			UpdatedAt: database.Timestamptz(now),
			DeletedAt: pgtype.Timestamptz{
				Status: pgtype.Null,
			},
		}
		_, v2 := e2.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			v1,
			v2,
		})
		_, err := r.ListByIDs(ctx, mockDB.DB, ids)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}

func TestMediaRepo_Delete(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := MediaRepoWithSqlMock()
	medias := []string{"media-id"}

	t.Run("err delete", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &medias)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := r.DeleteByIDs(ctx, mockDB.DB, medias)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &medias)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.DeleteByIDs(ctx, mockDB.DB, medias)
		assert.Nil(t, err)
	})
}
