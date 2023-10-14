package controller

import (
	"context"

	lesson_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	zoom_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/infrastructure/repo"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/service"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

type ZoomAccountController struct {
	zoomAccountService *service.ZoomAccountService
}

func InitZoomAccountController(
	wrapperConnection *support.WrapperDBConnection,
	zoomService service.ZoomServiceInterface,
) *ZoomAccountController {
	return &ZoomAccountController{
		zoomAccountService: service.NewZoomAccountService(wrapperConnection, zoomService, &zoom_repo.ZoomAccountRepo{}, &lesson_repo.LessonRepo{}),
	}
}

func (l *ZoomAccountController) ImportZoomAccount(ctx context.Context, req *lpb.ImportZoomAccountRequest) (res *lpb.ImportZoomAccountResponse, err error) {
	return l.zoomAccountService.ImportZoomAccount(ctx, req)
}

func (l *ZoomAccountController) ExportZoomAccount(ctx context.Context, req *lpb.ExportZoomAccountRequest) (res *lpb.ExportZoomAccountResponse, err error) {
	return l.zoomAccountService.ExportZoomAccount(ctx, req)
}
