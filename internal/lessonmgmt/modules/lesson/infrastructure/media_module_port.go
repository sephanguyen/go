package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
)

type MediaModulePort interface {
	RetrieveMediasByIDs(ctx context.Context, mediaIDs []string) (domain.Medias, error)
}
