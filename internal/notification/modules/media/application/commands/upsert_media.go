package commands

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/modules/media/infrastructure"
)

type UpsertMediaCommandHandler struct {
	DB        database.Ext
	MediaRepo infrastructure.MediaRepo
}

func (h *UpsertMediaCommandHandler) UpsertMedia(ctx context.Context, payload UpsertMediaPayload) (err error) {
	return h.MediaRepo.UpsertMediaBatch(ctx, h.DB, payload.Medias)
}
