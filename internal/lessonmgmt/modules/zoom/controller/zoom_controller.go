package controller

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/clients"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	infrastructure_lesson "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/service"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

type ZoomController struct {
	zoomService       service.ZoomServiceInterface
	lessonRepo        *infrastructure_lesson.LessonRepo
	wrapperConnection *support.WrapperDBConnection
}

func InitZoomController(
	cfg *configs.ZoomConfig,
	wrapperConnection *support.WrapperDBConnection,
	externalConfigService service.ExternalConfigServiceInterface,
	httpClient clients.HTTPClientInterface) *ZoomController {
	return &ZoomController{
		zoomService:       service.InitZoomService(cfg, externalConfigService, httpClient),
		wrapperConnection: wrapperConnection,
		lessonRepo:        &infrastructure_lesson.LessonRepo{},
	}
}

func (c *ZoomController) GenerateZoomLink(ctx context.Context, req *lpb.GenerateZoomLinkRequest) (*lpb.GenerateZoomLinkResponse, error) {
	parmGenerateZoomLink, err := domain.ConverterZoomGenerateMeetingRequest(req)
	if err != nil {
		return nil, fmt.Errorf("controller GenerateZoomLink fail: %w", err)
	}
	data, err := c.zoomService.RetryGenerateZoomLink(ctx, req.AccountOwner, parmGenerateZoomLink)
	if err != nil {
		return nil, fmt.Errorf("controller GenerateZoomLink fail: %w", err)
	}
	return &lpb.GenerateZoomLinkResponse{
		Url: data.URL,
		Id:  fmt.Sprint(data.ZoomID),
		Occurrences: sliceutils.Map(data.Occurrences, func(val *domain.OccurrenceOfZoomResponse) *lpb.GenerateZoomLinkResponse_OccurrenceZoom {
			return &lpb.GenerateZoomLinkResponse_OccurrenceZoom{
				OccurrenceId: val.OccurrenceID,
				StartTime:    val.StartTime,
				Duration:     int32(val.Duration),
				Status:       val.Status,
			}
		}),
	}, nil
}

func (c *ZoomController) DeleteZoomLink(ctx context.Context, req *lpb.DeleteZoomLinkRequest) (*lpb.DeleteZoomLinkResponse, error) {
	conn, err := c.wrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	zoomID := req.ZoomId
	lessonID := req.LessonId
	if lessonID != "" {
		lesson, err := c.lessonRepo.GetLessonByID(ctx, conn, lessonID)
		if err != nil {
			return nil, fmt.Errorf("DeleteZoomLink GetLessonByID fail: %w", err)
		}
		if zoomID == "" {
			zoomID = lesson.ZoomID
		}
	}
	if zoomID != "" {
		_, err := c.zoomService.RetryDeleteZoomLink(ctx, zoomID)
		if err != nil {
			return nil, fmt.Errorf("DeleteZoomLink RetryDeleteZoomLink fail: %w", err)
		}
	}

	if lessonID != "" {
		err := c.lessonRepo.RemoveZoomLinkByLessonID(ctx, conn, lessonID)
		if err != nil {
			return nil, fmt.Errorf("DeleteZoomLink RemoveZoomLinkByLessonID fail: %w", err)
		}
	}
	return &lpb.DeleteZoomLinkResponse{}, nil
}
