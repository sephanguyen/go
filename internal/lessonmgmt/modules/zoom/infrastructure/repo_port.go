package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/infrastructure/repo"
)

type ZoomAccountRepo interface {
	GetZoomAccountByID(ctx context.Context, db database.QueryExecer, id string) (*domain.ZoomAccount, error)
	Upsert(ctx context.Context, db database.Ext, zoomAccounts domain.ZoomAccounts) error
	GetAllZoomAccount(ctx context.Context, db database.QueryExecer) ([]*repo.ZoomAccount, error)
}
