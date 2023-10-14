package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/modules/media/domain"
)

type MediaRepo interface {
	UpsertMediaBatch(ctx context.Context, db database.QueryExecer, media domain.Medias) error
}
