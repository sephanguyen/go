package commands

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/notification/modules/media/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repo "github.com/manabie-com/backend/mock/notification/modules/media/repositories"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_UpsertMedia(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	mediaRepo := new(mock_repo.MockMediaRepo)

	t.Run("return empty err from repo", func(t *testing.T) {
		// arrange
		handler := UpsertMediaCommandHandler{
			DB:        db,
			MediaRepo: mediaRepo,
		}
		payload := UpsertMediaPayload{
			Medias: domain.Medias{
				{
					MediaID: pgtype.Text{},
				},
			},
		}
		mediaRepo.On("UpsertMediaBatch", ctx, db, mock.Anything).Once().Return(nil)

		//act
		err := handler.UpsertMedia(ctx, payload)

		//assert
		assert.Nil(t, err)
	})

}
