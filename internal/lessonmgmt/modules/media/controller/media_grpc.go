package controller

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/media/infrastructure"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MediaGRPCService struct {
	db        database.Ext
	mediaRepo infrastructure.MediaRepoInterface
}

func NewMediaGRPCService(db database.Ext, repo infrastructure.MediaRepoInterface) *MediaGRPCService {
	return &MediaGRPCService{
		db:        db,
		mediaRepo: repo,
	}
}

func (s *MediaGRPCService) RetrieveMediasByIDs(ctx context.Context, req *lpb.RetrieveMediasByIDsRequest) (*lpb.RetrieveMediasByIDsResponse, error) {
	medias, err := s.mediaRepo.RetrieveMediasByIDs(ctx, s.db, req.GetMediaIds())
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("mediaRepo.ListByIDs: %v", err))
	}

	return medias.ToRetrieveMediasByIDsResponse(), nil
}

func (s *MediaGRPCService) CreateMedia(ctx context.Context, req *lpb.CreateMediaRequest) (*lpb.CreateMediaResponse, error) {
	media, err := domain.FromMediaProtoToMediaDomain(req.GetMedia())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid media: %s", err.Error()))
	}

	if err := s.mediaRepo.CreateMedia(ctx, s.db, media); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error in mediaRepo.CreateMedia, media %s: %s", media.ID, err.Error()))
	}

	return &lpb.CreateMediaResponse{}, nil
}

func (s *MediaGRPCService) DeleteMedias(ctx context.Context, req *lpb.DeleteMediasRequest) (*lpb.DeleteMediasResponse, error) {
	mediaIDs := req.GetMediaIds()
	if len(mediaIDs) == 0 {
		return nil, status.Error(codes.InvalidArgument, "media IDs cannot be empty")
	}

	if err := s.mediaRepo.DeleteMedias(ctx, s.db, mediaIDs); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error in mediaRepo.DeleteMedia, mediaIDs %v: %s", mediaIDs, err.Error()))
	}

	return &lpb.DeleteMediasResponse{}, nil
}
