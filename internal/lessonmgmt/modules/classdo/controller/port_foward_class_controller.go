package controller

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/clients"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/infrastructure/repo"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/service"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

type PortForwardClassDoController struct {
	PortForwardClassDoService *service.PortForwardClassDoService
}

func InitPortForwardClassDoController(cfg *configs.ClassDoConfig, lessonmgmtDB database.Ext, httpClient clients.HTTPClientInterface) *PortForwardClassDoController {
	return &PortForwardClassDoController{
		PortForwardClassDoService: service.NewPortForwardClassDoService(cfg, lessonmgmtDB, httpClient, &repo.ClassDoAccountRepo{}),
	}
}

func (c *PortForwardClassDoController) PortForwardClassDo(ctx context.Context, req *lpb.PortForwardClassDoRequest) (*lpb.PortForwardClassDoResponse, error) {
	request := &domain.PortForwardClassDoRequest{}
	request.FromProto(req)
	response, err := c.PortForwardClassDoService.PortForwardClassDo(ctx, request)
	if err != nil {
		return nil, err
	}
	return response.ToProto(), nil
}
