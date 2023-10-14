package controller

import (
	"context"

	"github.com/manabie-com/backend/internal/notification/modules/media/application/commands"
	"github.com/manabie-com/backend/internal/notification/modules/media/controller/mappers"
	"github.com/manabie-com/backend/internal/notification/modules/media/domain"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *MediaModifierService) UpsertMedia(ctx context.Context, req *npb.UpsertMediaRequest) (*npb.UpsertMediaResponse, error) {
	listMediaDomain := make([]*domain.Media, 0, len(req.Media))
	for _, media := range req.Media {
		em, err := mappers.PbToMediaDomain(media)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		listMediaDomain = append(listMediaDomain, em)
	}

	err := svc.UpsertMediaCommandHandler.UpsertMedia(ctx, commands.UpsertMediaPayload{
		Medias: listMediaDomain,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "svc.UpsertMediaCommandHandler.UpsertMedia: %v", err)
	}

	mediaIDs := make([]string, 0, len(req.Media))
	for _, m := range listMediaDomain {
		mediaIDs = append(mediaIDs, m.MediaID.String)
	}
	return &npb.UpsertMediaResponse{
		MediaIds: mediaIDs,
	}, nil
}
