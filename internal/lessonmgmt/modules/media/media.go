package media

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/media/controller"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/media/infrastructure"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

type Module struct {
	MediaGRPCService lpb.MediaReaderServiceServer
}

func New(db database.Ext, repo infrastructure.MediaRepoInterface) *Module {
	return &Module{
		MediaGRPCService: controller.NewMediaGRPCService(db, repo),
	}
}
