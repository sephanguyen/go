package grpc

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/infrastructure"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/infrastructure/repositories"
	"github.com/manabie-com/backend/internal/golibs/chatvendor"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	cpb "github.com/manabie-com/backend/pkg/manabuf/conversationmgmt/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AgoraUserMgmtService struct {
	DB     database.Ext
	Logger *zap.Logger

	ChatVendorClient chatvendor.ChatVendorClient

	cpb.UnimplementedAgoraUserMgmtServiceServer

	UserBasicInfoRepo infrastructure.UserBasicInfoRepo
}

func NewAgoraUserMgmtService(db database.Ext, chatVendor chatvendor.ChatVendorClient, logger *zap.Logger) *AgoraUserMgmtService {
	return &AgoraUserMgmtService{
		DB:                db,
		Logger:            logger,
		ChatVendorClient:  chatVendor,
		UserBasicInfoRepo: &repositories.UserBasicInfoRepo{},
	}
}

// GetAppInfo returns user privilege token
// TODO: if biz logic scaled and get complex, move it to application handler folder
func (s *AgoraUserMgmtService) GetAppInfo(ctx context.Context, _ *cpb.GetAppInfoRequest) (*cpb.GetAppInfoResponse, error) {
	_, userID, _ := interceptors.GetUserInfoFromContext(ctx)

	if userID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user ID is empty")
	}

	userInfo, err := s.UserBasicInfoRepo.GetUsers(ctx, s.DB, []string{userID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("failed GetUsers: %+v", err))
	}

	if len(userInfo) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "user does not exist")
	}

	userToken, expireTime, err := s.ChatVendorClient.GetUserToken(userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("failed GetUserToken: %+v", err))
	}

	appKey := s.ChatVendorClient.GetAppKey()

	resp := &cpb.GetAppInfoResponse{
		AppKey:           appKey,
		TokenExpiredAt:   expireTime,
		CurrentUserToken: userToken,
	}
	return resp, nil
}
