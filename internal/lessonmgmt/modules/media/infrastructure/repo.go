package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
)

type MediaRepoInterface interface {
	RetrieveMediasByIDs(ctx context.Context, db database.QueryExecer, mediaIDs []string) (domain.Medias, error)
	CreateMedia(ctx context.Context, db database.QueryExecer, media *domain.Media) error
	DeleteMedias(ctx context.Context, db database.QueryExecer, mediaIDs []string) error
}
