package controller

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/infrastructure/repo"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/service"
	infrastructure_lesson "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

type ClassDoAccountController struct {
	ClassDoAccountService *service.ClassDoAccountService
}

func InitClassDoAccountController(cfg *configs.ClassDoConfig, lessonmgmtDB database.Ext) *ClassDoAccountController {
	return &ClassDoAccountController{
		ClassDoAccountService: service.NewClassDoAccountService(cfg, lessonmgmtDB, &repo.ClassDoAccountRepo{}, &infrastructure_lesson.LessonRepo{}),
	}
}

func (c *ClassDoAccountController) ImportClassDoAccount(ctx context.Context, req *lpb.ImportClassDoAccountRequest) (res *lpb.ImportClassDoAccountResponse, err error) {
	return c.ClassDoAccountService.ImportClassDoAccount(ctx, req)
}

func (c *ClassDoAccountController) ExportClassDoAccount(ctx context.Context, req *lpb.ExportClassDoAccountRequest) (res *lpb.ExportClassDoAccountResponse, err error) {
	return c.ClassDoAccountService.ExportClassDoAccount(ctx, req)
}
