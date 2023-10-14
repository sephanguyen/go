package mediaadapter

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/media"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

type MediaModuleAdapter struct {
	*media.Module
}

func (m *MediaModuleAdapter) RetrieveMediasByIDs(ctx context.Context, mediaIDs []string) (domain.Medias, error) {
	res, err := m.MediaGRPCService.RetrieveMediasByIDs(ctx, &lpb.RetrieveMediasByIDsRequest{
		MediaIds: mediaIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("MediaGRPCService.RetrieveMediasByIDs: %w", err)
	}

	return domain.FromRetrieveMediasByIDsResponse(res), nil
}

func (m *MediaModuleAdapter) CreateMedia(ctx context.Context, media *domain.Media) error {
	mediaPb := domain.FromMediaDomainToMediaProto(media)

	_, err := m.MediaGRPCService.CreateMedia(ctx, &lpb.CreateMediaRequest{
		Media: mediaPb,
	})
	if err != nil {
		return fmt.Errorf("MediaGRPCService.CreateMedia: %w", err)
	}

	return nil
}

func (m *MediaModuleAdapter) DeleteMedias(ctx context.Context, mediaIDs []string) error {
	_, err := m.MediaGRPCService.DeleteMedias(ctx, &lpb.DeleteMediasRequest{
		MediaIds: mediaIDs,
	})
	if err != nil {
		return fmt.Errorf("MediaGRPCService.DeleteMedias: %w", err)
	}

	return nil
}
