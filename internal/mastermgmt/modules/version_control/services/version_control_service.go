package services

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	pb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type VersionControlService struct {
	pb.UnimplementedVersionControlReaderServiceServer
	JoinClientVersions string
}

func (s *VersionControlService) VerifyAppVersion(ctx context.Context, req *pb.VerifyAppVersionRequest) (*pb.VerifyAppVersionResponse, error) {
	err := interceptors.CheckForceUpdateApp(ctx, s.JoinClientVersions)
	// force update return false instead off error
	if err != nil && status.Code(err) == codes.Aborted {
		ctxzap.Extract(ctx).Error("VerifyAppVersion", zap.Error(err))
		return &pb.VerifyAppVersionResponse{
			IsValid: false,
		}, nil
	} else if err != nil {
		return nil, err
	}

	return &pb.VerifyAppVersionResponse{
		IsValid: true,
	}, nil
}
