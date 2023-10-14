package services

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/yasuo/configurations"
	pb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UploadReaderService struct {
	DBTrace database.Ext
	Logger  *zap.Logger
	Config  *configurations.Config

	Uploader
}

func (s *UploadReaderService) RetrieveUploadInfo(ctx context.Context, _ *emptypb.Empty) (*pb.RetrieveUploadInfoResponse, error) {
	return &pb.RetrieveUploadInfoResponse{
		Endpoint: s.Config.Storage.Endpoint,
		Bucket:   s.Config.Storage.Bucket,
	}, nil
}
